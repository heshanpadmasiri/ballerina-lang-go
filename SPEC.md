# Worker message passing

Issue: https://github.com/ballerina-nutcracker/ballerina/issues/932. Refined intent: `INTENT.md`.

## Goals

- Support spec 7.8 worker message passing between the default worker and the named workers of a block function body (or lambda body):
  - async send `v -> w;`, sync send `v ->> w`
  - single receive `<- w`, multiple receive `<- {a: w1, w2, function}`
  - flush with a peer (`flush w`) and without (`flush`)
  - `function` as the peer naming the default worker
  - failure and panic propagation between peers (7.8.1–7.8.3)
  - a worker with undelivered async messages waits for them to be received before terminating, on both success and failure termination (7.8.1)
- Reject at symbol resolution every send/receive that doesn't line up: no matching send, no matching receive, interaction after wait, and every deadlock visible from message actions, unconditional direct-reference waits and worker termination.
- Enabling refactors, each its own commit passing `make lint test`, landed before message passing:
  1. Desugar the default worker of a function with named workers as a worker closure with its own future.
  2. Type resolve the default worker and the named workers concurrently (passes `make test-race`). Built on #1020 and `fix(worker): adapt named workers to capture groups`.
  3. Fix the parser turning `m:x ->> w` into an async send.

## Non-goals

- Message actions in a conditionally or repeatedly evaluated position report unimplemented "worker message send/receive is not supported here":
  - bodies and clauses of `if`, `while`, `foreach`, `match`, `do`, `lock`, nested `{}`
  - any query clause (`from int i in <- a select i`, `from int i in l select <- a`)
  - a lambda / anonymous function / object constructor between the action and the worker group that owns the peer. A lambda's own workers can exchange messages among themselves.
  - compound assignment RHS (`x += <- a;`, `x += v -> a;`): the parser accepts an action there (`parseCompoundAssignmentStmtRhs`), but it becomes a binary operand, so the node builder reports this unimplemented error
  - a send under `trap` (`trap (v ->> w)`): a trapped panic in the value would skip the send while the worker carries on (agreed with the user). `trap <- w` is allowed
  - any statement after an `if` without `else`: the node builder moves the statements that follow such an `if` into a block it adds, so they count as nested `{}` (agreed with the user; follow-up issue). `if c { return; } else { ... } 1 -> a;` keeps `1 -> a` a worker statement
- Allowed positions: action statement, variable initializer, assignment RHS, `return`, `check`, `trap`, parentheses, `match` subject, `foreach` collection. Apart from compound assignment, the parser rejects actions as operands (verified), so a statement outside a query holds at most one message action.
- Alternate receive `<- a | b` reports unimplemented "alternate receive is not supported".
- Fork statement workers (already unimplemented).
- Message actions (send, receive, flush) in a body with workers that is outside a module-level function or a class or service method report unimplemented "worker message send/receive is not supported outside a module-level function or method": a lambda that is (or is nested in) a module variable initializer, a parameter default, or a class or record field default. These bodies are type resolved with module-level nodes by `ResolvePublicNodes`, whose lazily resolved state isn't safe for concurrent use, so their workers are resolved one after another (agreed with the user). Named workers without message actions there keep working.
- Narrowing soundness for captured variables beyond #1020 plus the worker adaptation commit (#956, #967). Worker bodies follow the lambda capture policy. This is user visible (`worker w { if v is int { int y = v; } }` is now an error) and must be called out in the PR description.
- The existing `desugarNestedFunction` bug (it doesn't copy `defaultClosureVars`, so a function-typed local with default args called from a worker body or lambda crashes). File an issue; not fixed here.

### Limitations

- Pairing models only unconditional waits whose operand is a direct worker reference. A wait in control flow, inside a lambda, or on an alias (`future<int> f = w; wait f`) is not modelled.
- Deadlocks that only happen when a worker fails (`check`, error `return`) are not detected.
- `F(p)` is the error part of `p`'s declared return type, not the precise set of errors reachable before the action (#1056).
- The send expression isn't typed with the receive's contextually expected type; types flow send → receive only. The multiple receive record type is built without the expected type (#1055).
- Flush only reports on async sends since the last sync send / flush to that peer. A second `flush b` after the first returned `b`'s error returns `()`.
- Following 7.8.1, a worker that terminates normally while any async message it sent is still unreceived panics with the receiver's termination value. This happens whether or not an earlier flush / sync send error for that receiver was handled, e.g. `worker a returns error? { 1 -> b; check flush b; }` panics when `b` fails before receiving.
- Diagnostics are appended in arrival order, so the CLI order of type resolution errors from different workers of one function is nondeterministic (as it already is across functions; corpus tests sort them).
- In an isolated function each worker is alone on its thread, so a blocked message action busy-yields (same as `waitOnLatch`).
- Stack traces keep only the innermost hidden frame's location. A panic in a hidden generated function (lambda, default-parameter function) called from the default worker's trailing statements therefore loses the call-site frame, or reports the first trailing statement's line for it, as named workers already do. A panic in a visible function keeps its trace unchanged.

## Success criteria

- Every scenario under Tests passes as a corpus test. `make lint test` and `make test-race` pass after each commit of the work order.
- `corpus/bal/subset10/10-worker/message-passing1-fe.bal` and `flush1-fe.bal` are replaced by `-v` / `-e` tests.
- Existing worker tests are unchanged, except that `lock1-p`'s `@panic` marker moves to `return wait w;`.

# Design

## Work order

1. Refactor desugar: the default worker becomes a worker (includes the default worker symbol).
2. Refactor the type resolver, one commit each:
   1. merge `implicitImports` of worker/lambda resolvers into the parent
   2. unique XML-step owner per worker
   3. `snapshotArgumentState` restores only the symbols the trial wrote
   4. per-goroutine `ephemeralState`
   5. the default worker's trailing statements get their own child resolver; the parent atom side table is read-only while children run
   6. resolve the default worker and the named workers concurrently
3. Parser `->>` merge fix.
4. Message passing.

## Pipeline flow

```
node builder   -> builds send/receive/flush nodes; rejects async send outside statement position, alternate receive
symbol resolver-> declares the default worker symbol, validates peers and placement, allocates a message handle per send,
                  builds each worker's queue, computes flush coverage, runs pairing at the end of the body
                  (on success every receive has its handle; any error stops the pipeline before type resolution)
type resolver  -> resolves worker bodies concurrently; a send publishes its expression type under its handle,
                  a receive blocks until that type is published (or poisoned)
semantic       -> sent type <: value:Cloneable; async send to a receiver with non-empty F must be Covered
CFG            -> no reachable success termination before all message actions of the worker
desugar        -> default worker closure, one WorkerMessage per send, lang.__internal calls, termination delivery
runtime        -> WorkerMessage latches + peer futures
```

## Default worker

- Only for block function bodies (functions, methods, lambdas) with named workers.
- Symbol resolution declares a worker symbol whose name source code can't produce (`$function`; `worker 'function {}` is a valid user worker named `function`). It is added to the body's function scope without the `isShadowed` check, so a lambda with its own workers nested in a body with workers doesn't report "Variable already defined". It is stored in `BLangBlockFunctionBody.DefaultWorker`. The peer keyword `function` resolves through that field, never through name lookup. The symbol is never referenced as a value.
- Type resolution sets its type to `future<R>`, where `R` is the enclosing function's (or lambda's) declared return type.
- Desugar:

```ballerina
function f() returns T {
    <init stmts>
    handle $latch = createLatch();
    handle $m1 = createWorkerMessage(); ...                       // one per send (step 4 only)
    future<T> $default; future<T1> $w1;
    var $cd = function () returns T { waitOnLatch($latch); <trailing stmts> };  // $worker:<owner>:$function
    var $c1 = function () returns T1 { waitOnLatch($latch); <body 1> };         // $worker:<owner>:w1
    $default = start $cd(); $w1 = start $c1();
    openLatch($latch);
    return wait $default;
}
```

- The trailing statements are lowered in the outer function context and then wrapped in the lambda. `desugarNestedFunction` isn't used. BIR gen resolves captures lexically.
- `$default`'s start has a position spanning the trailing statements (the body's closing brace when there are none), so `-p` traces don't change (the new strand's stack is seeded with the parent's frames and the spawn frame collapses when it encloses the panic line).
- The `$default` slot is in `workerFutureSlots` under `DefaultWorker`. `workerFutureSlots` and `workerMessages` are extended, not replaced, by a nested body with its own workers.

## Placement and peers (symbol resolution)

- The owning worker of an action is found from the nearest function resolver:
  - a named worker body resolver → that worker, group = the declaring body's group
  - a function or lambda resolver → the default worker of that body, group = that body's group. `declareNamedWorkers` sets the group on the body's resolver, after `InitStmts` are walked and before the worker bodies and trailing statements, so there is no group in `InitStmts` or in a body without workers
- Peer rules:
  - `function` from the default worker → "worker can't send/receive to itself" error. This includes `InitStmts` and bodies without workers, where the default worker is still the owner. A name there → "undefined worker" (workers aren't in scope in `InitStmts`, 7.3.2)
  - `function` elsewhere → the group's default worker
  - a name → must be a named worker of the group. A worker of an enclosing group reached through a lambda → unimplemented. Otherwise "undefined worker".
  - a peer appears at most once in a multiple receive
- Placement: walk from the action's resolver to the nearest block resolver, seeing through `workerSymbolResolver`. If it isn't a function resolver → unimplemented. Verified: `if` / `while` / `foreach` / `do` / `lock` / `{}` / match clause bodies and query clauses all have block resolvers. `match` subjects and `foreach` collections are resolved on the enclosing resolver.
- The peer is stored as a `model.SymbolRef` on the node, not a `BLangSimpleVarRef`, so `checkWorkerReference` never counts it.

## Message handle

- A send allocates a `model.WorkerMessageRef` (zero = unset) in symbol resolution and stores it on its node, in its queue element and in its worker's send list. Pairing sets the handle on the matching receive (per field for a multiple receive).
- Type resolution uses the message type store:
  - `Publish(ref, ty)` once per send (publishing twice is an internal error)
  - `PublishPoisonIfUnset(ref)` for every send of a worker, deferred by that worker's goroutine
  - `Type(ref)` blocks until published or poisoned. A poisoned receive gets no type and reports no diagnostic, but asserts that the context already has errors (otherwise internal error), so an unpublished send can't reach semantic analysis.
- Candidate trials (`enterEphemeral`) install a fresh store for the trial. Children built during the trial inherit it, and it is discarded with the trial. Trials also run on `packageTypeResolver` (module-level `new`), so both resolver kinds hold a store, read through one accessor `resolverMessageTypes(t)` (like `resolverEphemeralState`).
- Handles are allocated by an atomic counter on `CompilerEnvironment`, separate from the stores, so trial stores never mint refs.
- Desugar keeps the `WorkerMessage` variable per handle in `functionContext.workerMessages`.

## Queues and flush coverage

- While resolving each worker body (and the default worker's trailing statements), append in evaluation order: message actions, plus unconditional waits (`wait w`, `wait a | b`, `wait {x: a, y: b}`) whose operands are direct references to group workers.
- Flush coverage is computed in source order per worker:

```go
pending := map[model.SymbolRef][]*ast.BLangWorkerAsyncSendAction{}
for each message action a of the worker, in source order:
    switch a := a.(type) {
    case *ast.BLangWorkerAsyncSendAction:
        pending[a.Peer] = append(pending[a.Peer], a)
    case *ast.BLangWorkerSyncSendAction:
        markCovered(pending[a.Peer])
        pending[a.Peer] = nil
    case *ast.BLangWorkerFlushAction:
        peers := a.Peer, or every other worker of the group in declaration order (default worker first)
        for p in peers:
            if len(pending[p]) > 0 {
                a.Covered = append(a.Covered, ast.BLangWorkerFlushCoverage{Peer: p, Messages: refs(pending[p])})
            }
            markCovered(pending[p])
            pending[p] = nil
    }
```

## Pairing

- Queue elements are a sum type, one struct per kind with a private marker method, each holding its AST node. Pairing writes the handle straight into the receive node (`node.Message = ref`, or `field.Message` for a multiple receive field). The pseudo code below uses `kind` for the type switch and `peers`, `message` and `setRecvMessage` as shorthand for accessors on the element.
- Runs once per body with workers, at the end of `resolveBlockFunctionBody`, over every worker including the default one. It is skipped when any message action of the group failed peer or placement validation (`workerMessageGroup.invalid`), so its counterpart doesn't get a second, unrelated error. Order: default worker first, then declaration order. Reports the first error deterministically.
- What each action blocks on:

| Action | Blocks until |
|---|---|
| `v -> p` | never |
| `v ->> p` | `p` receives this message (FIFO: earlier async sends to `p` first) |
| `<- p` | `p` has sent the matching message |
| `<- {a, b}` | every peer has sent its message |
| `flush p` | every message sent to `p` so far is received |
| `flush` | every queue to every peer is empty |
| `wait p` | `p` terminated |
| `wait a \| b` | any of them terminated with success, or all terminated (`waitAnyFuture` skips errors); modelled optimistically as "any terminated" |
| `wait {a, b}` | all of them terminated |
| end of worker | every async message it sent is received |

0. Count check (schedule independent): for every ordered pair `p → q`, the number of sends from `p` to `q` must equal the number of receives (single, or multiple receive fields) at `q` from `p`. On a mismatch report "no matching receive" at the first excess send, or "no matching send" at the first excess receive, in source order. This keeps missing sends/receives from surfacing as deadlocks (e.g. `worker a { int x = <- b; 2 -> b; } worker b { int y = <- a; }` is "no matching send" at `<- b`).
1. Interaction after wait (static; flush and `wait a | b` excluded). `wait a | b` is left to the simulation because which operand finished is unknown, e.g. `worker b { int x = <- function; }`, `_ = wait a | b; 1 -> b;` is valid:

```go
for _, worker := range order {
    waited := set[model.SymbolRef]{}
    for _, e := range states[worker].queue {
        switch e.kind {
        case waitAll:
            waited.addAll(e.peers)
        case asyncSend, syncSend, singleRecv, multipleRecv:
            if containsAny(e.peers, waited) {
                return error(e.pos, "worker interaction after wait action")
            }
        }
    }
}
```

2. Simulation (monotone conditions, so any firing order reaches the same state):

```go
pending := order
for len(pending) > 0 {
    advanced := false
    var nextPending []model.SymbolRef
    for _, worker := range pending {
        state := states[worker]
        if step(states, worker, state) { advanced = true }
        if !state.terminated { nextPending = append(nextPending, worker) }
    }
    pending = nextPending
    if !advanced && len(pending) > 0 { return reportStuck(states, pending) }
}

func step(states workerStateMap, worker model.SymbolRef, state *workerMessageState) bool {
    next, ok := state.next()
    if !ok {
        if !allAsyncQueuesEmpty(state) { return false } // 7.8.1
        state.terminated = true
        return true
    }
    switch next.kind {
    case asyncSend:
        state.asyncSentQueue[next.peers[0]] = append(state.asyncSentQueue[next.peers[0]], next)
        state.proceed(); return true
    case syncSend:
        dest := next.peers[0]
        if len(state.asyncSentQueue[dest]) > 0 { return false }
        destState := states[dest]
        destNext, ok := destState.next()
        if !ok || destNext.kind != singleRecv || destNext.peers[0] != worker { return false }
        destNext.setRecvMessage[0](next.message)
        state.proceed(); destState.proceed(); return true
    case singleRecv:
        source := states[next.peers[0]]
        if len(source.asyncSentQueue[worker]) == 0 { return false }
        next.setRecvMessage[0](source.asyncSentQueue[worker][0].message)
        source.asyncSentQueue[worker] = source.asyncSentQueue[worker][1:]
        state.proceed(); return true
    case multipleRecv:
        for _, peer := range next.peers {
            if !fieldAvailable(states[peer], worker) { return false }
        }
        for i, peer := range next.peers {
            source := states[peer]
            if len(source.asyncSentQueue[worker]) > 0 {
                next.setRecvMessage[i](source.asyncSentQueue[worker][0].message)
                source.asyncSentQueue[worker] = source.asyncSentQueue[worker][1:]
            } else {
                sourceNext, _ := source.next()
                next.setRecvMessage[i](sourceNext.message)
                source.proceed()
            }
        }
        state.proceed(); return true
    case flush:
        for _, peer := range next.peers {
            if len(state.asyncSentQueue[peer]) > 0 { return false }
        }
        state.proceed(); return true
    case waitAll:
        for _, peer := range next.peers {
            if !states[peer].terminated { return false }
        }
        state.proceed(); return true
    case waitAny:
        for _, peer := range next.peers {
            if states[peer].terminated { state.proceed(); return true }
        }
        return false
    }
    internalError(next.pos, "invalid message kind")
    return false
}

func fieldAvailable(source *workerMessageState, worker model.SymbolRef) bool {
    if len(source.asyncSentQueue[worker]) > 0 { return true }
    sourceNext, ok := source.next()
    return ok && sourceNext.kind == syncSend && sourceNext.peers[0] == worker
}
```

3. Reporting when stuck:

```go
func reportStuck(states workerStateMap, pending []model.SymbolRef) {
    for _, worker := range pending {
        if pos, msg, ok := unmatched(states, worker); ok { return error(pos, msg) }
    }
    return error(blockedPos(states[pending[0]]), "worker message deadlock")
}

func unmatched(states workerStateMap, worker model.SymbolRef) (diagnostics.Location, string, bool) {
    state := states[worker]
    next, ok := state.next()
    if !ok {
        for _, dest := range order {
            if q := state.asyncSentQueue[dest]; len(q) > 0 && states[dest].done() {
                return q[0].pos, "no matching receive", true
            }
        }
        return diagnostics.Location{}, "", false
    }
    switch next.kind {
    case syncSend:
        if states[next.peers[0]].done() { return next.pos, "no matching receive", true }
    case singleRecv, multipleRecv:
        for _, peer := range next.peers {
            if states[peer].done() && !fieldAvailable(states[peer], worker) {
                return next.pos, "no matching send", true
            }
        }
    case flush:
        for _, peer := range next.peers {
            if q := state.asyncSentQueue[peer]; len(q) > 0 && states[peer].done() {
                return q[0].pos, "no matching receive", true
            }
        }
    }
    return diagnostics.Location{}, "", false
}
```

- Cost O(actions × workers).

## Type resolution

- `resolveBlockFunctionBody` with workers:
  1. resolve `InitStmts` on the parent resolver
  2. declare every worker's type, including the default worker symbol
  3. before starting any goroutine, build one child resolver per named worker plus one for the trailing statements. Each child has its own:
     - atom side table
     - `semtypes.Context`
     - `implicitImports`
     - `monoCounters` / `xmlStepCounter`
     - `xmlStepOwner` (`<owner>$<worker>`, `<owner>$function`)
     - `ephemeralState` seeded with the parent's depth
     - the parent's current message type store

     Named worker children use the worker's return type and the startup chain with a function boundary. The default child uses the function's return type and the startup chain with every worker's capture group, with no boundary.
  4. resolve each child in its own goroutine. Each goroutine defers `PublishPoisonIfUnset` for its worker's sends (so siblings blocked on a receive finish) and recovers a Go panic (internal error), which the parent re-raises after the join (precedent: `projects/package_compilation.go:159`). The poison defer runs on panic too, so blocked siblings are released and the join completes; siblings are not otherwise stopped. Join.
  5. merge each child's `implicitImports` into the parent and OR its `refusedDependent` into the parent's
- While children run, the parent's atom side table is only read.
- Module-level nodes (resolved by `ResolvePublicNodes`) are resolved lazily, and `packageTypeResolver` state (`lazyResolutionStatus`, deferred emptiness checks, ...) isn't safe for concurrent use. `packageTypeResolver.resolvingPublicNodes` is set while `ResolvePublicNodes` runs, and then the children of a body with workers (e.g. `var f = function () { worker a { ... } ... };`) are resolved one after another on the resolving goroutine. Symbol resolution rejects message actions in such bodies (see Non-goals), so no child waits on a receive. Function bodies resolved by `ResolvePrivateNodes` run concurrently.
- `snapshotArgumentState` restore writes back only the symbols declared inside the argument subtree.
- Types:
  - send expression: resolved with no expected type from the receive
  - single receive: `T|F(w)`
  - multiple receive: the closed record of the per-field `T` (built like a mapping constructor with no expected type), unioned with each field's `F`
  - sync send: `F(w)|()`
  - flush: the union over `Covered` of `F(p)|()`, or `()` if `Covered` is empty
- `F(p)` is the error part of `p`'s declared return type (for `function`, from the default worker symbol's `future<R>`).
- The context checks the result as usual. `v ->> w;` / `flush w;` with non-empty `F` hit the existing "expression value must be assigned".
- No deadlock: type resolution blocks only on receive→send, which is a subset of the runtime orderings proved feasible by pairing.

## Semantic analysis and CFG

- Sent expression type must be a subtype of `value:Cloneable` (inline `IsSubtype(cx, ty, CreateCloneable(cx))`).
- An async send to a receiver with non-empty `F` is an error unless `Covered`.
- CFG: it is an error if a worker (named or default) can terminate with success before executing all its message actions (sends, receives and flushes; stricter than the spec for flush). With the non-goals this means a reachable `return` inside control flow followed by a message action of the same worker, including a `return` in `InitStmts` when the default worker's trailing statements have message actions. Allowed if the returned expression's static type is a subtype of `error`.

## Desugar lowering

- `$m` is the send's `WorkerMessage`, `$w` the peer's future slot:
  - `v -> w;` → `setWorkerMessageValue($m, v)`
  - `v ->> w` → `setWorkerMessageValue($m, v); waitWorkerMessageReceived($m, $w)`, type `F(w)|()`
  - `<- w` → `getWorkerMessageValue($m, $w)`
  - `<- {a: w1, b: w2}` → a record built from `getWorkerMessageValues([$m1, $m2], [$w1, $w2])`, or the error it returns
  - `flush w` / `flush` → `flushWorkerMessages([[<Covered messages of p1>], ...], [$p1, ...])`, or `()` if `Covered` is empty
- `WorkerMessage[]`, `WorkerMessage[][]` and future lists are built with list constructors (a handle list can't be mutated at runtime).
- The determined type of every generated call is set directly.
- Termination delivery, only for workers (including the default) with at least one async send:

```ballerina
var $c1 = function () returns T1 {
    waitOnLatch($latch);
    T1 $r = (function () returns T1 { <body 1> })();     // $worker: prefixed, hidden
    awaitWorkerMessageDelivery([<async send messages>], [<their receivers' futures>]);
    return $r;
};
```

## Runtime (`lang.__internal`)

- `WorkerMessage` is a `handle` over a goroutine-safe struct holding a value and two one-shot latches, `set` and `received`.
- Every blocking function loops: check latch / `IsComplete()`, else `<-ctx.Yield()`. None of them block on a Go channel.
- `GetClaimed()` is called only after `IsComplete()`. It doesn't claim, so a later `wait` still works, and it re-panics with the stored `panicWithStack`, keeping the peer's stack. This is documented in code.
- Outcomes are decided per message, never from the peer's termination alone: check the message latch (`set` for a receive, `received` for a send) first, and consult `IsComplete()` / `GetClaimed()` only while it is still unreleased, re-checking the latch after seeing the peer complete. A receiver that takes the message and then fails therefore doesn't fail the sync send / flush / delivery.
- A flush or delivery reports a receiver's outcome only if at least one of its messages was never received.
- A peer that completed with success without doing its matching action is unreachable after the compile time checks: internal-error panic.

# API changes

## parser/parser.go

- `func (b *ballerinaParser) mergeQualifiedNameWithExpr(qualifiedName st.STNode, exprOrAction st.STNode) st.STNode` (private, unchanged signature)
  - Current: the `SYNC_SEND_ACTION` case calls `st.CreateAsyncSendActionNode`.
  - New: calls `st.CreateSyncSendActionNode(newLhsExpr, syncSend.SyncSendToken, syncSend.PeerWorker)`. Affects only a `{` statement whose first member is `m:x ->> w`.

## ast/expressions.go

- Removed: `Channel`, `(*Channel).WorkerPairId`, `(*Channel).ChannelId`, `WorkerPairId`, `BLangWorkerSendReceiveExprBase`, `BLangWorkerReceive` (+ `GetWorkerName`, `ToActionString`, `isAction`), `BLangWorkerSendExprBase` (+ `GetExpr`, `GetWorkerName`), `BLangAlternateWorkerReceive` (+ `ToActionString`). `ast.TypeKindChannel` and the node builder's `TypeKindChannel` mapping stay.
- Added (public, all embed `bLangActionBase`; each implements `isAction`, `ToActionString` and the node interface assertions):

```go
// BLangWorkerPeer is the peer of a message action as written in source.
type BLangWorkerPeer struct {
    Name   string          // constants.DefaultWorkerName ("$function") for the `function` keyword; a quoted 'function is "function"
    Symbol model.SymbolRef // set by symbol resolution
    Pos       diagnostics.Location
}
type BLangWorkerAsyncSendAction struct {
    bLangActionBase
    Expr    BLangExpression
    Peer    BLangWorkerPeer
    Message model.WorkerMessageRef // allocated by symbol resolution
    Covered bool                   // a later sync send or flush to Peer exists
}
type BLangWorkerSyncSendAction struct {
    bLangActionBase
    Expr    BLangExpression
    Peer    BLangWorkerPeer
    Message model.WorkerMessageRef
}
type BLangWorkerReceiveAction struct {
    bLangActionBase
    Peer    BLangWorkerPeer
    Message model.WorkerMessageRef // set by pairing
}
type BLangWorkerReceiveField struct {
    FieldName string // defaults to the peer's source name (`function` for the keyword)
    Peer      BLangWorkerPeer
    Message   model.WorkerMessageRef // set by pairing
}
type BLangWorkerMultipleReceiveAction struct {
    bLangActionBase
    Fields []BLangWorkerReceiveField
}
type BLangWorkerFlushCoverage struct {
    Peer     model.SymbolRef
    Messages []model.WorkerMessageRef
}
type BLangWorkerFlushAction struct {
    bLangActionBase
    Peer    *BLangWorkerPeer // nil when no peer
    Covered []BLangWorkerFlushCoverage
}
```

## ast/walk.go, ast/pretty_printer.go

- `Walk`: remove the `BLangWorkerReceive` / `BLangAlternateWorkerReceive` cases. The send cases walk `Expr`; receive and flush have no children.
- The pretty printer prints all five nodes (peer names; send `Expr`), feeding AST goldens.

## ast/ast.go

- `BLangBlockFunctionBody`: add `DefaultWorker model.SymbolRef` (zero if no workers) and `DefaultWorkerSendMessages []model.WorkerMessageRef`.
- `BLangNamedWorkerDeclaration`: add `SendMessages []model.WorkerMessageRef`.

## model

- Add `type WorkerMessageRef int` (public; zero = unset).
- Add `constants.DefaultWorkerName = "$function"` in `common/constants` (shared by the node builder, symbols and desugar).

## context/env.go, context/context.go, context/internal/workermessages (new)

- New package `workermessages` with `type Store struct` (RWMutex + map + per-ref ready channel).
  - `func NewStore() *Store`
  - `func (s *Store) Publish(ref model.WorkerMessageRef, ty semtypes.SemType) error`: error if already published
  - `func (s *Store) PublishPoisonIfUnset(ref model.WorkerMessageRef)`
  - `func (s *Store) Type(ref model.WorkerMessageRef) (semtypes.SemType, bool)`: blocks; `false` when poisoned
- `CompilerEnvironment` holds the root store and an atomic handle counter. `CompilerContext` exposes `NewWorkerMessage() model.WorkerMessageRef` (the counter), `NewWorkerMessageTypeStore() *workermessages.Store` (trial stores), and a public type alias for the store so the type resolver can hold it.

## nodebuilder/node_builder.go

- `transformAsyncSendAction`, `transformSyncSendAction`, `transformReceiveAction`, `transformReceiveFields`, `transformFlushAction` (private)
  - Current: report unimplemented and return a bad node.
  - New: build the new nodes. The peer keyword `function` (`SimpleNameReferenceNode` with a `FUNCTION_KEYWORD` token) records `BLangWorkerPeer{Name: constants.DefaultWorkerName}`; the pretty printer prints it as `function`. A quoted `'function` records `Name: "function"`, a normal worker name.
  - A receive whose peer is an alternate (`<- a | b`) reports unimplemented "alternate receive is not supported".
- Where the node builder builds an assignment RHS, variable initializer, `check`, `trap`, `return`, `match` subject, `foreach` collection or brace (parenthesized) expression: an async send operand reports "async send action can only be used as a statement".

## semantics/internal/symbols

- `declareNamedWorkers(resolver *blockSymbolResolver, body *ast.BLangBlockFunctionBody)` (private)
  - Current: takes `workers []*ast.BLangNamedWorkerDeclaration` and declares their symbols.
  - New: also declares the default worker symbol, sets `body.DefaultWorker`, and creates the body's `workerMessageGroup`, marked `withModuleLevelNodes` when `inModuleLevelFunction(resolver)` is false. `messageOwner` reports unimplemented for message actions of such a group.
- New private `inModuleLevelFunction(resolver *blockSymbolResolver) bool`: walks the resolver chain; false through a default-expression resolver, otherwise true when the outermost function resolver (whose parent is the compilation unit, or a class or service) is a named function or method.
- `resolveBlockFunctionBody(resolver *blockSymbolResolver, body *ast.BLangBlockFunctionBody)`: after resolving everything, runs `pairWorkerMessages(resolver, group)` when the body has workers.
- `blockSymbolResolver` gains `messageGroup *workerMessageGroup` (set on the resolver of a body with workers) and `messageWorker model.SymbolRef` (the worker whose statements it resolves).
- New private file `worker_messages.go`:
  - `messageQueueElement` interface with private marker `isMessageQueueElement()` and `pos() diagnostics.Location`, implemented by `asyncSendElement{node *ast.BLangWorkerAsyncSendAction}`, `syncSendElement{node *ast.BLangWorkerSyncSendAction}`, `singleRecvElement{node *ast.BLangWorkerReceiveAction}`, `multipleRecvElement{node *ast.BLangWorkerMultipleReceiveAction}`, `flushElement{node *ast.BLangWorkerFlushAction, peers []model.SymbolRef}` `waitAllElement{peers []model.SymbolRef, pos}` and `waitAnyElement{peers []model.SymbolRef, pos}` (switches checked by `gochecksumtype`)
  - `workerMessageState{index, queue, asyncSentQueue map[model.SymbolRef][]messageQueueElement, terminated}` with `next`, `proceed`, `done`
  - `workerStateMap map[model.SymbolRef]*workerMessageState`
  - `workerMessageGroup{order []model.SymbolRef, states workerStateMap, invalid bool}`; flush coverage walks the same queues
  - `resolveMessagePeer(resolver *blockSymbolResolver, peer *ast.BLangWorkerPeer, pos diagnostics.Location) (group *workerMessageGroup, worker model.SymbolRef, ok bool)`: peer and placement validation; sets `peer.Symbol`
  - `computeFlushCoverage`, `checkMessageCounts`, `checkInteractionAfterWait`, `simulateWorkerMessages`, `step`, `fieldAvailable`, `reportStuck`, `unmatched`, `blockedPos`, `pairWorkerMessages`
- Visitor cases for the five nodes: validate, allocate handles (sends), append to the queue and the worker's `SendMessages` / `DefaultWorkerSendMessages`. Wait nodes with direct worker operands in an unconditional position also append to the queue.

## semantics/internal/types

- `resolveBlockFunctionBody(t typeResolver, chain *binding, body *ast.BLangBlockFunctionBody)` (private): concurrent flow as above.
- `declareNamedWorkerType`: also sets the default worker symbol's type `future<R>`.
- `resolveNamedWorker`: split into building the child resolver (before goroutines) and resolving it.
- New `resolveDefaultWorkerStatements` child.
- `functionTypeResolver` and `packageTypeResolver` gain `messageTypes *workermessages.Store`, read through the new private `resolverMessageTypes(t typeResolver) *workermessages.Store`. `functionTypeResolver.ephemeralState` is now owned per goroutine.
- `enterEphemeral(t typeResolver) func()`: also installs a fresh message type store and restores the previous one on exit.
- `snapshotArgumentState(t typeResolver, args []ast.BLangExpression) func()`: restore writes back only symbols declared inside the argument subtree.
- New private resolvers for the five actions, plus `workerFailureType(t, peer model.SymbolRef) semtypes.SemType`.

## semantics/internal/analysis/semantic_analyzer.go

- New checks on send nodes: Cloneable; uncovered async send to a receiver with non-empty `F`.

## semantics/internal/cfg

- New check in the CFG analyzer for each worker graph (named and default): success-termination before all message actions.

## desugar

- `desugarWorkerRegion(cx *functionContext, body *ast.BLangBlockFunctionBody) []ast.StatementNode`
  - Current: latch, slots, closures, starts, open latch; `walkBlockFunctionBody` then appends the lowered trailing statements.
  - New: also creates the `$m` messages, builds the default worker closure from the lowered trailing statements (started first), wraps workers with async sends for termination delivery, and makes the body end with `return wait $default`. `walkBlockFunctionBody` changes accordingly.
- `functionContext` gains `workerMessages map[model.WorkerMessageRef]*ast.BLangVarRef`. It is propagated like `workerFutureSlots`; both are extended by nested bodies with workers.
- New lowering functions for the five nodes and invocation helpers for the new `lang.__internal` functions.
- `createLangInternalInvocation`: reports `cx.internalError` for a missing module or symbol instead of producing a zero `SymbolRef`.
- `workerClosureName`: the default worker uses `$worker:<owner>:$function`.

## lib/langlibs/ballerina/lang.__internal/.../lang.__internal.bal, lib/langlibs/go/lang.__internal/internal.go

- New `public isolated ... = external` functions, registered in `internal.go`:
  - `createWorkerMessage() returns handle`
  - `setWorkerMessageValue(handle m, any|error value)`: `values.Clone` then set
  - `getWorkerMessageValue(handle m, future<any|error> sender) returns any|error`
  - `getWorkerMessageValues(handle[] ms, future<any|error>[] senders) returns (any|error)[]|error`: all or nothing (a message can itself be an error value; a list is never an error, so the failure case stays distinct). If a sender terminates first, releases `received` on every message and returns the error / re-panics.
  - `waitWorkerMessageReceived(handle m, future<any|error> receiver) returns error?`
  - `flushWorkerMessages(handle[][] ms, future<any|error>[] receivers) returns error?`: over the receivers that left a message unreceived, panic first (declaration order), else the first error, else `()`
  - `awaitWorkerMessageDelivery(handle[] ms, future<any|error>[] receivers)`: like flush but panics with the receiver's termination value. Skips messages whose `set` was never released (the sender failed before that send).
- New private Go type `workerMessage` (value + `set` / `received` latches, mutex).

# Tests

- Refactor 1:
  - every existing `10-worker` test passes
  - `lock1-p` `@panic` moves to `return wait w;` (golden updated)
  - `panic1-p` trace unchanged
  - a default-worker panic trace in trailing statements unchanged
  - output order of existing `-v` tests unchanged
- Refactor 2:
  - `-v`: two workers with different XML steps print `<b>1</b><c>2</c>`
  - `-v`: a lambda with workers passed to a multi-candidate `new`, at function and module level
  - after step 4: `A|B x = new (function () returns int { worker w { 1 -> function; } return <- w; });` at function level (the send is resolved three times without a double-publish error)
  - `make test-race` passes
- Parser: `{ io:x ->> w; }` is a sync send. A message action in `{}` is a non-goal, so this is a `-fe` test; the parser corpus output shows the sync send (AST goldens only exist for `-v`/`-p` tests). A committed regression check is a follow-up issue (agreed with the user).
- Node builder `-e`:
  - `x += <- a;` and `x += v -> a;` unimplemented
  - async send as var initializer, assignment RHS, `check`, `trap`, `return`, `match` subject, `foreach` collection, parenthesized
  - `<- a | b` unimplemented
- Symbol resolution `-e`:
  - undefined peer
  - `function` self send and receive from the default worker
  - `<- w` / `-> function` in `InitStmts`
  - duplicate multiple-receive peer
  - unimplemented in `if`, `while`, `foreach` body, `match` clause, `lock`, `{}`, query `from` and `select`
  - `-fv`: send, receive and flush in a module variable initializer lambda (passed to a multi-candidate `new`), a parameter default, a class field default and a record field default (`message-module-level1-fv`)
  - peer through a lambda boundary
  - extra send (no matching receive, async and sync)
  - extra receive (no matching send, single and multiple)
  - interaction after `wait` (send and receive side, `wait a | b`, `wait {a, b}`)
  - cyclic receive deadlock
  - deadlock through a wait
  - deadlock through termination delivery
  - deterministic first error with several faults
  - count mismatch reported as "no matching send" where the simulation alone would report a deadlock
  - a placement-rejected send whose counterpart receive has no `@error` marker (pairing skipped)
- Symbol resolution `-v`: `_ = wait a | b; 1 -> b;` where `b` receives from `function` first.
- Symbol resolution `-v`: `worker 'function` exchanging messages with the default worker, `function` in a multiple receive (`r.'function`).
- Type `-e`:
  - receive assigned to an incompatible type
  - `v ->> w;` and `flush w;` unassigned with non-empty `F`
  - multiple receive into an incompatible record
- Semantic `-e`: non-Cloneable send (object value), uncovered async send to an error-returning receiver.
- Semantic `-v`: the same send covered by a later sync send, by `flush w`, and by `flush`.
- CFG `-e`:
  - success `return` in `if` before a send, a receive and a flush (named and default workers)
  - `return` in `InitStmts` with trailing message actions
- CFG `-v`: error-typed `return` before a message action.
- CFG tests put the `return` in an `if` with an `else` (see Non-goals).
- `-v`: the #1055 tuple example `[1, "x"] -> function; [int, string] t = <- a;` type checks, since a list constructor without an expected type already gets the tuple type (agreed with the user).
- `-v` runtime:
  - async send + receive
  - an error value sent as a message, received by single and multiple receive
  - message passing in a method and in an object `init` using `self`
  - a lambda with its own message-passing workers called inside a `while` body (fresh messages per iteration)
  - a lambda with its own message-passing workers (using its `function`) inside a function that has workers
  - a function with workers and no trailing statements
  - sync send + receive
  - FIFO ordering of several sends
  - multiple receive (named fields, shorthand, `function` field)
  - flush with and without a peer, and a flush covering nothing
  - send and receive with `function` both ways
  - message passing inside a lambda's own workers
  - allowed positions: `check`, `trap`, `return`, `match` subject, `foreach` collection, parentheses
  - values are cloned (mutating after send isn't seen)
  - isolated and non-isolated functions
  - receive gets the sender's error
  - sync send gets the receiver's error
  - flush returns the first error in declaration order
  - multiple receive returns the error of the sender that failed
  - a worker failing before a later async send doesn't deadlock
  - a receive from `function` after the function's `wait $default` path
  - `wait w` after message passing with `w` still works
- `-v` runtime: a receiver that fails / panics right after receiving; the sync send and the flush still return `()`.
- `-p` runtime: `check flush b` (and `check (v ->> b)` followed by an unreceived async send) against a receiver that fails before receiving: `flush` returns the error, then termination delivery panics.
- `-p` runtime:
  - sender panic re-raised at a receive (first line is the sender's panic site)
  - receiver panic seen by a sync send and by a flush
  - termination delivery panic with the receiver's value
  - an ignored flush error panics at termination
  - flush prefers a panic over an error
