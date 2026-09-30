# Add support for worker message passing
+ Issue: https://github.com/ballerina-nutcracker/ballerina/issues/932

## Goal
+ Support message passing as specified in [spec](https://ballerina.io/spec/lang/master/#section_7.8) with the following restrictions (see Non-goals)
  + async send (`v -> w;`), sync send (`v ->> w`)
  + single receive (`<- w`) and multiple receive (`<- {a: w1, b: w2}`, `<- {w1, function}`)
  + flush with and without a peer (`flush w`, `flush`)
  + `function` as the peer referring to the default worker
  + failure and panic propagation between peers (7.8.1, 7.8.2, 7.8.3)
  + a worker with undelivered async messages waits for them to be received before terminating, on both success and failure termination (7.8.1)
+ Detect at compile time (symbol resolution) every send/receive that doesn't line up: no matching send, no matching receive, interaction after wait, and every deadlock visible from message actions, unconditional waits and worker termination (see pairing algorithm)
+ Enabling refactors (done and committed first, each passing `make lint test`)
  1. Desugar the default worker of a function with named workers as a worker (its own closure and future)
  2. Type resolve the default worker and the named workers of a function concurrently, passing `make test-race`. Built on the capture narrowing fix ([#1020](https://github.com/ballerina-nutcracker/ballerina/pull/1020)), which is cherry-picked into this stack together with `fix(worker): adapt named workers to capture groups`

## Non-goals
+ Message passing in a conditionally or repeatedly evaluated position. These get an unimplemented error "worker message send/receive is not supported here"
  + bodies and clauses of control flow statements: `if`, `while`, `foreach`, `match`, `do`, `lock`, nested blocks `{}`
  + query expressions (any clause, e.g. `from int i in l select <- a`, `from int i in <- a select i`)
  + a lambda / anonymous function / object constructor between the action and the worker group that owns the peer (e.g. a lambda inside worker `a` sending to sibling `b`). A lambda's own named workers can exchange messages among themselves
  + compound assignment RHS (`x += <- a;`, `x += v -> a;`): the parser accepts an action there but it becomes a binary operand; reported by the node builder
+ Allowed positions: action statement, variable initializer, assignment right hand side, `return`, `check`, `trap`, parentheses, `match` subject, `foreach` collection (the last two are evaluated once, unconditionally), e.g. `int x = check <- a;`, `return <- a;`, `match <- a { ... }`, `foreach int i in <- a { ... }`
  + Apart from compound assignment, the parser never accepts an action as an operand (`foo(<- a)`, `(<- a) -> b`, `c ? <- a : 1`, `let ... in <- a`, `[<- a]` are syntax errors today), so a statement outside a query contains at most one message action
+ Alternate receive (`<- a | b`): unimplemented error
+ Fork statement workers (already unimplemented)
+ Narrowing soundness for variables captured by workers beyond what #1020 plus the worker adaptation commit give ([#956](https://github.com/ballerina-nutcracker/ballerina/issues/956), [#967](https://github.com/ballerina-nutcracker/ballerina/issues/967)). With them, worker bodies follow the lambda capture policy: a mutable variable captured by any worker can't be narrowed in any worker body or in the default worker statements after the workers
  + This is a user visible change (e.g. `worker w { if v is int { int y = v; } }` is now an error). The PR description must call it out

## Limitation
+ Pairing models only unconditional waits whose operand is a direct worker reference. Deadlocks through a wait in control flow, a wait inside a lambda (called or not) or a wait on an alias of a worker (`future<int> f = w; wait f`) are not detected
+ Existing bug, out of scope (file an issue): `desugarNestedFunction` doesn't copy `defaultClosureVars`, so calling a local function-typed variable that relies on its default args (declared in init stmts) from a worker body or a lambda crashes with a nil deref. The default worker refactor doesn't use `desugarNestedFunction`, so it doesn't add to this
+ Deadlocks that only happen when a worker fails (a `check` or an error `return` fires) are not detected. The failing worker's termination waits for its undelivered messages while the receiver waits on something the failing worker would have done later
```ballerina
worker w { 1 -> r; check f(); 2 -> s; }
worker r { int a = <- s; int b = <- w; }
worker s { int c = <- w; 3 -> r; }   // deadlocks only if f() fails
```
+ Failure type of a peer is the error part of its declared return type, not the precise set of errors it can terminate with before the send/receive ([#1056](https://github.com/ballerina-nutcracker/ballerina/issues/1056))
+ Send expression is not interpreted with the contextually expected type of the receive (7.8.1). Types flow only from send to receive, so context dependent expressions are typed on their own
  + jBallerina does the same. Tracked in [#1055](https://github.com/ballerina-nutcracker/ballerina/issues/1055) (can be addressed later using the message handles)
```ballerina
worker a { [1, "x"] -> b; }            // (int|string)[]
worker b { [int, string] t = <- a; }   // error, needs `<[int, string]>[1, "x"] -> b;`
```
+ Multiple receive result type is built without the contextually expected type (closed record of the send types) ([#1055](https://github.com/ballerina-nutcracker/ballerina/issues/1055))
+ Flush deviates from the spec to keep its static type sound. The spec types a flush as `()` after an intervening sync send or flush, but at runtime a flush returns the receiver's error whenever the queue is not empty, e.g. a second `flush b` after the first returned `b`'s error. Here a flush only checks the async sends since the last sync send / flush to that peer, so the second flush returns `()`
+ Following 7.8.1, a worker terminating normally with an unreceived async message panics with the receiver's termination value, whether or not an earlier flush error for that receiver was handled (e.g. `1 -> b; check flush b;` panics when `b` fails before receiving)
+ Diagnostic order across workers of one function is nondeterministic after concurrent type resolution (as across functions today)

## Design
+ Since symbol resolution proves that every send/receive lines up, no runtime queue is needed to dispatch messages dynamically. Each send/receive pair gets its own `WorkerMessage` at runtime

### Runtime (`lang.__internal`)
+ `WorkerMessage` holds a clonable value and two latches: `set` (value is available) and `received` (value was taken by the receiver)
  + Must be goroutine safe, workers of an isolated function run in parallel
  + All blocking functions poll the latch and yield (`<-ctx.Yield()`) like `waitOnLatch`. In an isolated function each worker is alone on its thread, so this busy-yields, as `waitOnLatch` does today (accepted)
    + never block on a Go channel; that would stall siblings sharing a non-isolated thread
+ A `WorkerMessage` alone can't tell whether the peer terminated without sending/receiving, so every blocking function also takes the peer's future
  + `IsComplete()` to check (non blocking) and `GetClaimed()` to get the outcome, called only after `IsComplete()` is true since it blocks the goroutine. `GetClaimed` doesn't claim the future so a later `wait` on the worker still works
  + This should be documented in the code
  + `GetClaimed()` re-panics with the stored `panicWithStack`, so re-panic through it to keep the peer's stack. Panicking with the bare error value loses it
+ The runtime already captures a worker's return value, error and panic in its future (`startFuture`), so panics anywhere in the peer are observed without codegen support
+ `WorkerMessage` is a `handle`. `WorkerMessage[]` args must be built by a list constructor in desugar (`SemTypeForValue` has no handle case, so a handle list can't be mutated)
+ Declare every new function in `lang.__internal.bal` and register it in `internal.go`. `createLangInternalInvocation` silently ignores missing symbols today; make it an internal error
+ Functions
  + `createWorkerMessage() returns WorkerMessage`
  + `setWorkerMessageValue(WorkerMessage m, any|error value)`
    + clones the value (`values.Clone`), sets it and releases `set`
  + `getWorkerMessageValue(WorkerMessage m, future<any|error> sender) returns any|error`
    + loop: if `set` is released take the value, release `received` and return it; else if sender is complete (recheck `set` first) return the sender's error or re-panic with its panic value
  + `getWorkerMessageValues(WorkerMessage[] ms, future<any|error>[] senders) returns (any|error)[]|error` (multiple receive; a message can itself be an error)
    + waits until every message is `set` and only then takes all of them and releases all `received`. If a sender terminates before its value is set, releases `received` on every message (the receive action was executed, so their sync senders and flushes complete with `()`) and returns its error / re-panics
  + `waitWorkerMessageReceived(WorkerMessage m, future<any|error> receiver) returns error?` (sync send)
    + waits until `received` is released or the receiver terminates; returns the receiver's error or re-panics
  + `flushWorkerMessages(WorkerMessage[][] ms, future<any|error>[] receivers) returns error?` (`ms[i]` are the covered messages to `receivers[i]`)
    + waits for every receiver until all its messages are `received` or it terminates. Then: if any receiver panicked, re-panic with the first in declaration order; else return the first error in declaration order; else `()`
  + `awaitWorkerMessageDelivery(WorkerMessage[] ms, future<any|error>[] receivers)` (worker termination, 7.8.1)
    + same as flush but panics with the receiver's termination value instead of returning an error
    + skips messages whose `set` was never released: a worker that fails (`check`) before a later async send never sets that message, and waiting on it would deadlock (the receiver waits for the sender to terminate)
      + this also applies when an earlier flush already returned that receiver's error and the worker ignored it (follows the spec, intentional)
  + Outcomes are decided per message: check the message latch first and consult the peer's future only while it is unreleased (recheck after seeing it complete), so a peer that received and then failed doesn't fail the sender. Flush / delivery report a receiver only if one of its messages was never received
  + A peer that completed with success without doing its matching action is unreachable given the compile time checks: internal-error panic
+ Desugared calls are not type checked, so desugar sets the determined type of each call directly (as `createLangMapGetInvocation` does)

### AST / node builder
+ New nodes for async send, sync send, single receive, multiple receive and flush, all embedding `bLangActionBase`
  + Remove the unused `Channel`, `WorkerPairId`, `BLangWorkerSendReceiveExprBase`, `BLangWorkerReceive`, `BLangWorkerSendExprBase`, `BLangAlternateWorkerReceive`
  + Each send / receive (field) holds
    1. `Message model.WorkerMessageRef`: handle to the compile time information about the message (see MessageHandle)
    2. the peer as a source name (`function` for the default worker) and a worker symbol set by the symbol resolver (you need it to get the worker future in desugar)
      - the default worker has a real worker symbol (see Symbol resolution)
      - store the peer as a symbol, not a `BLangSimpleVarRef`, so the worker reference rule (`checkWorkerReference`) never sees it. e.g. `1 -> w; int x = <- w; int y = wait w;` has one worker reference (the `wait`)
  + Async send also holds `Covered bool` (a later sync send or flush to the peer exists). Flush holds an optional peer and `Covered []{Peer, Messages}` (see Flush coverage)
+ `function` as a peer is a `SimpleNameReferenceNode` whose token is `FUNCTION_KEYWORD`; the node builder records the name `function`
+ Async send and sync send are both actions, but an async send can only be used as a statement (`v -> w;`); unlike sync send (`F(w)|()`) its result can't be used as a value. The parser accepts it in other positions; reject each in the node builder with "async send action can only be used as a statement": `x = v -> w;`, `int y = v -> w;`, `check v -> w;`, `trap v -> w`, `return v -> w;`, `match v -> w {}`, `foreach int i in v -> w {}`, and inside parentheses `x = (v -> w);` (the parser accepts it; `(v -> w);` as a statement is already a syntax error)
+ `<- a | b` reports unimplemented "alternate receive is not supported"
+ Fix parser `mergeQualifiedNameWithExpr` turning `m:x ->> w;` into an async send (`SYNC_SEND_ACTION` case calls `CreateAsyncSendActionNode`)
  + only fires for a statement starting with `{` whose first member is `m:x ->> w` (e.g. `{ io:x ->> w; }`), not for a plain `m:x ->> w;`

## MessageHandle
+ AST nodes of a send/receive pair share a message handle (`model.WorkerMessageRef`, zero means unset). The send allocates it; pairing sets it on the receive
+ The compiler env holds only the message type, in a store safe for concurrent use
  + `Publish(ref, type)`: send side, exactly once (publishing twice is an internal error)
  + `PublishPoisonIfUnset(ref)`: releases receivers when the send is never resolved
  + `Type(ref)`: receive side, blocks until published or poisoned
+ The `WorkerMessage` variable of each handle is desugar-local state (`functionContext.workerMessages`), not in the env

### Symbol resolution
+ Default worker symbol: `declareNamedWorkers` also declares a worker symbol with an internal name source code can't produce (`$function`) and records it on the body. It is added without the `isShadowed` check (a lambda with workers nested in a body with workers declares its own). `function` can't be used as the symbol name: `worker 'function {}` compiles today and stores the name `function`. The peer keyword `function` resolves through `DefaultWorker`; a peer `'function` refers to the user's named worker (`BLangBlockFunctionBody.DefaultWorker`). Its type is `future<R>` where `R` is the enclosing function's (or lambda's) declared return type. It is never referenced as a value
  + refactor 1 (desugar) already introduces it, without any message passing
+ Validate peers
  + peer must be a worker of the same group (the named workers of the body and its default worker)
    + a worker of an enclosing body reached through a lambda is unimplemented (see non-goals)
    + no such worker is "undefined worker"
  + `function` used from the default worker is a self send/receive error (named worker self reference is already unknown symbol)
  + a peer can appear only once in a multiple receive
  + a lambda's top level has its own function resolver, so placement alone accepts it; peer validation rejects a peer outside the current group
+ placement: walk from the action's resolver to the nearest block resolver (seeing through `workerSymbolResolver`); if it is not a function resolver the position is conditional/repeated → unimplemented. No new resolver or marker is needed: `if`, `while`, `foreach`, `do`, `lock`, `{}`, match clause bodies (block statements) and query clauses already have their own block resolvers, while a `match` subject and a `foreach` collection are walked on the enclosing resolver
+ Every send allocates its message handle, stores it on its node and in its queue element, and appends it to its worker's send list (`BLangNamedWorkerDeclaration.SendMessages`, `BLangBlockFunctionBody.DefaultWorkerSendMessages`; the type resolver uses these to publish poison)
+ Every receive registers a callback that sets the handle on its node (one per field for a multiple receive)
+ While resolving each worker body (and the default worker statements after the worker declarations) build that worker's queue in evaluation order: message actions and unconditional waits (`wait w`, `wait a | b`, `wait {x: a, y: b}`) whose operands are direct references to workers of the group. Waits in control flow are left out
+ At the end of each function body with workers (`resolveBlockFunctionBody`) run the pairing algorithm once with every worker including `function`. Lambdas with workers run it for their own body
+ Before simulating, check per ordered pair that send and receive counts match ("no matching send/receive" at the first excess action), so missing actions aren't reported as deadlocks
+ Skip pairing for a group where any message action failed peer/placement validation (no second unrelated error)
+ Pairing reports the first error it detects, with a position, deterministically (default worker first, then declaration order)
+ Any error here stops the pipeline before stage 5, so type resolution only sees lined up programs

### Flush coverage
+ Computed once while building a worker's queue, in source order, stored on the nodes
```go
pending := map[model.SymbolRef][]*ast.BLangWorkerAsyncSendAction{}
for each message action a of the worker, in source order:
    switch a := a.(type) {
    case *ast.BLangWorkerAsyncSendAction:
        pending[a.Peer] = append(pending[a.Peer], a)
    case *ast.BLangWorkerSyncSendAction:
        markCovered(pending[a.Peer]) // sets Covered = true on each
        pending[a.Peer] = nil
    case *ast.BLangWorkerFlushAction:
        peers := a.Peer, or every other worker of the group in declaration order
        for p in peers:
            if len(pending[p]) > 0 {
                a.Covered = append(a.Covered, BLangWorkerFlushCoverage{Peer: p, Messages: refs(pending[p])})
            }
            markCovered(pending[p])
            pending[p] = nil
    }
```
+ A flush with nothing covered is valid, has type `()` and returns immediately (spec 7.8.3)

### Failure type
+ `F(p)` for peer `p` is the error part of `p`'s declared return type. For `function` it is the error part of the enclosing function's (or lambda's) declared return type (read from the default worker symbol's `future<R>`)
+ The spec's no message type `N` is always empty since sends in conditional contexts are a non-goal

### Type resolution
+ Types flow only in the send direction
  + The send expression is resolved without an expected type from the receive (only constrained to `value:Cloneable`)
  + Single receive type is `T|F(w)` where `T` is the send expression type. Multiple receive type is the closed record built from each field's `T` (like a mapping constructor with no expected type) with the union of each field's `F`. The receive's context then checks this type as usual
  + Sync send type is `F(w)|()`. Flush type is the union over its `Covered` peers `p` of `F(p)|()`, `()` if `Covered` is empty. Neither waits on the peer since `F` comes from declared return types
  + An action statement whose type isn't nil is already an error ("expression value must be assigned"), so `v ->> w;` and `flush w;` with non-empty `F` must be assigned
+ You get the message type `T` from the message handle, which the send action publishes. The receive blocks on it
+ Resolve the default worker and the workers of the function concurrently (follows on from the desugar refactor, before starting on message passing; must pass with race detection). `resolveBlockFunctionBody` with workers:
  1. resolve `InitStmts` on the parent resolver
  2. declare every worker's type, including the default worker symbol
  3. build a child resolver for every named worker and one for the default worker's statements before starting any goroutine. Each child has its own atom side table, `semtypes.Context`, `implicitImports`, `monoCounters`, a unique `xmlStepOwner` (`<owner>$<worker>`, `<owner>$function`), its own `ephemeralState` seeded with the parent's depth, and the parent's current message type store
     + named worker: worker return type, chain with a function boundary (as today)
     + default worker: function return type, the startup chain with every worker's capture group, no function boundary
  4. resolve each child in its own goroutine and join
  5. merge each child's `implicitImports` into the parent; OR each child's `refusedDependent` into the parent's
  + While the children run, the parent's atom side tables are only read (only `InitStmts` wrote them). Capture groups, binding chains, the type env, symbol spaces and diagnostics are already safe for concurrent use
  + Closures don't affect each other's narrowings (validated: workers never see outer or sibling narrowings, and with capture groups every worker's group is on the startup chain before any body is resolved, so the default worker doesn't depend on worker bodies)
  + `snapshotArgumentState` restore writes back only the symbols the trial wrote (today it writes every symbol the arguments reference, racing with siblings reading outer variables). Invariant: a trial only writes symbols declared inside the argument subtree
+ Receives block until the send publishes
  + The symbol resolver proved that message passing can always complete without a deadlock, so this can't deadlock either
  + Every send must publish exactly once on every path out of its worker's resolution (failure, skipped/unreachable statements, early return on a failed worker return type, panics). Each goroutine defers `PublishPoisonIfUnset` for every send of its worker, so a receiver never waits forever. A poisoned receive fails without a new diagnostic. On a Go panic (internal error) the goroutine recovers it, the poison defer still runs so siblings finish, and the parent re-raises it after the join
  + Candidate trials (the multi-alternative `new` path, the only `enterEphemeral` site) resolve their arguments once per candidate and again for the winner, e.g. `A|B x = new (function () returns int { worker w { 1 -> function; } return <- w; });` resolves `1 -> function` three times. `enterEphemeral` installs a fresh message type store for the trial, inherited by child resolvers built during it and discarded with it. Every send/receive resolved during a trial pairs within the trial

### Semantic analysis
+ Static type of the sent expression must be a subtype of `value:Cloneable`
+ An async send to a receiver with non-empty `F` is an error unless a later sync send or flush to that receiver in the same worker covers it (`Covered`, 7.8.1)

### CFG analysis
+ It is an error if a worker can terminate with success before executing all its message actions (7.8.4). Message actions here are sends, receives and flushes (stricter than the spec, which only requires sends and receives). With the non-goals the only way is a reachable `return` inside control flow followed by a message action of the same worker; it is an error unless the static type of the returned value is a subtype of error (failure termination is allowed)
  + applies to named workers and to the default worker, including a `return` in `InitStmts` when the default worker has message actions after the worker declarations

### Desugar
+ `WorkerMessage`s are created in the worker region before the closures are built (after the startup latch) and captured by the worker closures. `functionContext.workerMessages` maps each handle to its variable; like `workerFutureSlots` it is propagated to nested function contexts, and both are extended (not replaced) by a lambda with its own workers
+ Default worker: refactor it to be just like another worker, and the function blocks on it
  + only for function bodies with named workers
  + lower the trailing statements in the outer function context, then wrap them in a lambda. Don't use `desugarNestedFunction` (doesn't copy `defaultClosureVars`, restarts `desugarSymbolCounter`). BIR gen resolves captures lexically, so the wrapped statements work unchanged
  + closure name `$worker:<owner>:$function` (hidden from stack traces; `$worker:<owner>:function` is taken by a user worker `'function`). Return type is the function's final return type; the function body becomes `return wait $default;`
  + start the default closure first, waiting on the startup latch like the others, so output order doesn't change
  + give its start a position spanning the trailing statements so `-p` stack traces don't change (the default strand's stack is seeded with the parent's frames and the spawn frame collapses when it encloses the panic line)
+ `lock1-p` changes: the default worker now starts first, so the "attempted strand start while holding a lock" panic is reported at the default worker's start. Move its `@panic` marker to `return wait w;` and update the golden
  + its slot lives in `workerFutureSlots` under the `DefaultWorker` symbol like any named worker
  + receives from `function` never claim its future (`GetClaimed` only)
```ballerina
function f() returns T {
    <init stmts>
    handle $latch = createLatch();
    handle $m1 = createWorkerMessage(); ...          // one per send
    future<T> $default; future<T1> $w1;
    var $cd = function () returns T { waitOnLatch($latch); <trailing stmts> };   // $worker:<owner>:function
    var $c1 = function () returns T1 { waitOnLatch($latch); <body 1> };          // $worker:<owner>:w1
    $default = start $cd(); $w1 = start $c1();
    openLatch($latch);
    return wait $default;
}
```
+ A receive that re-panics with the sender's panic value carries the sender's captured stack, so the first stack trace line is the sender's panic site
+ Lowering (`$m` worker message, `$w` peer's future slot)
  + `v -> w;` -> `setWorkerMessageValue($m, v)`
  + `v ->> w` -> `setWorkerMessageValue($m, v); waitWorkerMessageReceived($m, $w)`, type `F(w)|()`
  + `<- w` -> `getWorkerMessageValue($m, $w)`
  + `<- {a: w1, b: w2}` -> record built from `getWorkerMessageValues([$m1, $m2], [$w1, $w2])` (or the error it returns)
  + `flush w` / `flush` -> `flushWorkerMessages([[<Covered messages of p1>], ...], [$p1, ...])` over the flush's `Covered` entries; `()` if `Covered` is empty
+ Worker termination (7.8.1): the worker closure runs its body as an inner closure then calls `awaitWorkerMessageDelivery` on the `WorkerMessage`s of its async sends before returning the result. This covers normal return and `check` failure; a panic skips it which is fine since the worker already terminated abnormally
  + applies to the default worker too; only emitted for workers with at least one async send
  + the inner closure uses the `$worker:` prefix so its frame is hidden from stack traces
```ballerina
var $c1 = function () returns T1 {
    waitOnLatch($latch);
    T1 $r = (function () returns T1 { <body 1> })();
    awaitWorkerMessageDelivery([<async send messages>], [<their receivers' futures>]);
    return $r;
};
```

## Symbol resolution worker message pairing algorithm
+ Reasoned from spec 7.7/7.8. Since message actions are unconditional, on the success path no peer terminates without doing its matching action, so the "or the peer terminates" escapes only fire on failure/panic paths
+ What each action blocks on

| Action | Blocks until |
|---|---|
| `v -> p` | never |
| `v ->> p` | `p` receives this message (FIFO: every earlier async send to `p` was received) |
| `<- p` | `p` has sent the matching message |
| `<- {a, b}` | every peer has sent its message |
| `flush p` | every message sent to `p` so far is received |
| `flush` | every queue to every peer is empty |
| `wait p` | `p` terminated |
| `wait a \| b` | any terminated with success, or all terminated (modelled optimistically as any terminated, so never a false deadlock) |
| `wait {a, b}` | all of them terminated |
| end of worker (7.8.1) | every async message it sent is received |

+ Every blocking condition, once true, stays true, so a simulation firing enabled actions in any order reaches the same final state. Getting stuck is a deadlock on every success path execution
+ Consequences
  + `flush p` after `wait p` never blocks (`p` received everything before terminating normally), so flush is not part of the wait check
  + send or receive with `p` after `wait p` always deadlocks (the receive side too: `p`'s sync send, or its termination after an async send, waits for us)
  + waits through other workers and termination delivery are caught by the simulation, e.g.
```ballerina
worker a { int _ = wait b; 1 -> c; }
worker b { int x = <- c; }
worker c { int y = <- a; 2 -> b; }        // deadlock through wait

worker a { 1 -> c; }
worker b { () _ = wait a; 2 -> c; }
worker c { int y = <- b; int x = <- a; }  // a can't terminate until c takes 1
```

1. While resolving each worker build its queue
```go
type messageKind int

const (
    asyncSend messageKind = iota + 1 // zero is invalid
    syncSend
    singleRecv
    multipleRecv
    flush
    waitAll // wait w, wait {x: a, y: b}
    waitAny // wait a | b
)

type messageQueueElement struct {
    kind messageKind
    pos  diagnostics.Location
    // asyncSend, syncSend, singleRecv, flush with a peer: peers[0]
    // multipleRecv: one per field
    // flush without a peer: every other worker of the group
    // waitAll, waitAny: the waited workers
    peers []model.SymbolRef
    // asyncSend, syncSend
    message model.WorkerMessageRef
    // singleRecv: setRecvMessage[0]; multipleRecv: setRecvMessage[i] is for peers[i]
    setRecvMessage []func(model.WorkerMessageRef)
}

type workerMessageState struct {
    index          int
    queue          []messageQueueElement
    asyncSentQueue map[model.SymbolRef][]messageQueueElement // undelivered async sends per receiver
    terminated     bool
}

func (w *workerMessageState) next() (messageQueueElement, bool) // false when every action ran
func (w *workerMessageState) proceed()
func (w *workerMessageState) done() bool                        // every action ran (may still be terminating)

type workerStateMap map[model.SymbolRef]*workerMessageState
```
2. Interaction after wait (static, before the simulation; only for a clearer message, the simulation would report these as deadlocks). Flush is excluded
```go
for _, worker := range order { // default worker first, then declaration order
    waited := set[model.SymbolRef]{}
    for _, e := range states[worker].queue {
        switch e.kind {
        case waitAll: // waitAny is left to the simulation (`wait a | b; 1 -> b;` can be valid)
            waited.addAll(e.peers)
        case asyncSend, syncSend, singleRecv, multipleRecv:
            if containsAny(e.peers, waited) {
                return error(e.pos, "worker interaction after wait action")
            }
        }
    }
}
```
3. Simulate
```go
pending := order
for len(pending) > 0 {
    advanced := false
    var nextPending []model.SymbolRef
    for _, worker := range pending {
        state := states[worker]
        if step(states, worker, state) {
            advanced = true
        }
        if !state.terminated {
            nextPending = append(nextPending, worker)
        }
    }
    pending = nextPending
    if !advanced && len(pending) > 0 {
        return reportStuck(states, pending)
    }
}
// every worker terminated, so every async queue was drained

func step(states workerStateMap, worker model.SymbolRef, state *workerMessageState) bool {
    next, ok := state.next()
    if !ok {
        // 7.8.1: termination waits until every async message is received
        if !allAsyncQueuesEmpty(state) {
            return false
        }
        state.terminated = true
        return true
    }
    switch next.kind {
    case asyncSend:
        state.asyncSentQueue[next.peers[0]] = append(state.asyncSentQueue[next.peers[0]], next)
        state.proceed()
        return true
    case syncSend:
        dest := next.peers[0]
        if len(state.asyncSentQueue[dest]) > 0 {
            return false // FIFO: earlier async sends to dest first
        }
        destState := states[dest]
        destNext, ok := destState.next()
        if !ok || destNext.kind != singleRecv || destNext.peers[0] != worker {
            return false // a multipleRecv releases the sync sender itself
        }
        destNext.setRecvMessage[0](next.message)
        state.proceed()
        destState.proceed()
        return true
    case singleRecv:
        source := states[next.peers[0]]
        if len(source.asyncSentQueue[worker]) == 0 {
            return false // a sync send to us is fired by the sender
        }
        next.setRecvMessage[0](source.asyncSentQueue[worker][0].message)
        source.asyncSentQueue[worker] = source.asyncSentQueue[worker][1:]
        state.proceed()
        return true
    case multipleRecv:
        for _, peer := range next.peers {
            if !fieldAvailable(states[peer], worker) {
                return false
            }
        }
        for i, peer := range next.peers {
            source := states[peer]
            if len(source.asyncSentQueue[worker]) > 0 {
                next.setRecvMessage[i](source.asyncSentQueue[worker][0].message)
                source.asyncSentQueue[worker] = source.asyncSentQueue[worker][1:]
            } else { // its next action is a sync send to us
                sourceNext, _ := source.next()
                next.setRecvMessage[i](sourceNext.message)
                source.proceed()
            }
        }
        state.proceed()
        return true
    case flush:
        for _, peer := range next.peers {
            if len(state.asyncSentQueue[peer]) > 0 {
                return false
            }
        }
        state.proceed()
        return true
    case waitAll:
        for _, peer := range next.peers {
            if !states[peer].terminated {
                return false
            }
        }
        state.proceed()
        return true
    case waitAny:
        for _, peer := range next.peers {
            if states[peer].terminated {
                state.proceed()
                return true
            }
        }
        return false
    }
    internalError(next.pos, "invalid message kind")
    return false
}

// a field from source is available if it has a queued async message for us or
// its next action is a sync send to us
func fieldAvailable(source *workerMessageState, worker model.SymbolRef) bool {
    if len(source.asyncSentQueue[worker]) > 0 {
        return true
    }
    sourceNext, ok := source.next()
    return ok && sourceNext.kind == syncSend && sourceNext.peers[0] == worker
}
```
4. Report when stuck: first unmatched action in fixed order, else deadlock
```go
func reportStuck(states workerStateMap, pending []model.SymbolRef) {
    for _, worker := range pending {
        if pos, msg, ok := unmatched(states, worker); ok {
            return error(pos, msg)
        }
    }
    return error(blockedPos(states[pending[0]]), "worker message deadlock")
}

// the worker is blocked on a peer that ran all its actions (done), so what it
// needs can never happen
func unmatched(states workerStateMap, worker model.SymbolRef) (diagnostics.Location, string, bool) {
    state := states[worker]
    next, ok := state.next()
    if !ok { // terminating with an undelivered message to a done peer
        for _, dest := range order {
            if q := state.asyncSentQueue[dest]; len(q) > 0 && states[dest].done() {
                return q[0].pos, "no matching receive", true
            }
        }
        return diagnostics.Location{}, "", false
    }
    switch next.kind {
    case syncSend:
        if states[next.peers[0]].done() {
            return next.pos, "no matching receive", true
        }
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

// position of the next action, or of the first undelivered async send when terminating
func blockedPos(state *workerMessageState) diagnostics.Location
```
+ Cost: each round fires at least one action, so O(actions × workers)

## Work order
1. Refactor desugar to treat the default worker section as a worker, including the default worker symbol (commit, all tests pass)
2. Refactor the type resolver to resolve worker bodies, including the default worker, concurrently. Each item below is a separate commit that passes `make lint test`
   1. Merge `implicitImports` of worker/lambda resolvers into the function's
   2. Make XML-step function names unique across workers (worker resolvers have `xmlStepOwner == ""` today, so `worker A {return x/<b>;} worker B {return x/<c>;}` prints `<c>2</c><c>2</c>`)
   3. Fix the `snapshotArgumentState` restore race (restore only the symbols the trial wrote)
   4. Per-goroutine `ephemeralState`
   5. Resolve the default worker's trailing statements with their own child resolver; parent atom side tables are read-only while workers resolve
   6. Resolve the default worker and the named workers concurrently (all tests pass with `make test-race`)
3. Fix the parser `->>` merge bug
4. Implement worker message passing
