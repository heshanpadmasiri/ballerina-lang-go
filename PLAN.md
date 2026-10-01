# Plan: worker message passing (SPEC.md)

## Context
Issue #932: support spec 7.8 worker message passing (send/receive/flush, `function` peer, failure/panic propagation, termination delivery), with compile-time pairing/deadlock detection in symbol resolution. SPEC.md is authoritative; this plan only maps it onto the code and orders the commits. Branch `feat/worker-msg-passing` (on top of #1020 + `fix(worker): adapt named workers to capture groups`). Every commit must pass `make lint test`; from refactor 2 on also `make test-race`. Intermediate commits may be transient (e.g. unimplemented errors in later stages); the end state matches the spec.

Facts from exploration that shape the plan:
- Node builder stubs: `nodebuilder/node_builder.go:4462-4550, 5891`. Old AST nodes (`ast/expressions.go:74-88, 286, 388-401`) are unused outside `ast/`.
- Symbols: `resolveBlockFunctionBody` / `declareNamedWorkers` / `resolveNamedWorker` at `semantics/internal/symbols/symbol_resolver.go:1247-1280`; `workerSymbolResolver` 158, `nearestBlockResolver` 446, `isShadowed` 1835.
- Types: `resolveBlockFunctionBody` / `declareNamedWorkerType` / `resolveNamedWorker` at `semantics/internal/types/type_resolver.go:2359-2433`. Workers today share the parent's `ephemeralState`, have no `xmlStepOwner` (names collide as `$xmlStep$$<n>`), and lambda/worker `implicitImports` are dropped (2466, 2555, 2417). `enterEphemeral` / `snapshotArgumentState` in `type_resolver_candidate_state.go:75, 147`. `packageTypeResolver` has no locking.
- Desugar: `desugarWorkerRegion` `desugar/worker.go:41`, `walkBlockFunctionBody` `desugar/statement.go:80`, `functionContext` `desugar/desugar.go:193`, `createLangInternalInvocation` `desugar/query_expression.go:2021` (no lookup check).
- Runtime: latch pattern in `lib/langlibs/go/lang.__internal/internal.go:113-179` (handle = raw Go pointer, `ctx.Yield()` loop); `values/future.go` (`IsComplete`, `GetClaimed`); `waitAnyFuture` `runtime/internal/exec/terminators.go:149`; hidden frames `runtime/internal/exec/errors.go:153-176`.
- CFG: worker graphs in `semantics/internal/cfg/control_flow_analyzer.go:121-300`, `analyzeWorkerExplicitReturn` `cfg_analyzer.go:100`.
- Stores pattern: `context/internal/capturegroups/store.go`.
- `gochecksumtype` is **not** enabled in `.golangci.yml`; it must be enabled as part of commit 4b (the spec relies on it).

## Commit 1 — `refactor(desugar): run the default worker as a worker closure`
- `common/constants`: add `DefaultWorkerName = "$function"`.
- `ast/ast.go`: `BLangBlockFunctionBody.DefaultWorker model.SymbolRef`.
- Symbols: `declareNamedWorkers(resolver, body)` also declares the `$function` worker symbol (no `isShadowed` check) when `len(body.Workers) > 0` and sets `body.DefaultWorker`.
- Types: `declareNamedWorkerType` path also sets the default worker symbol to `future<R>` (R = enclosing function/lambda return type; take it from the resolver's `retTy`).
- Desugar (`desugar/worker.go`, `statement.go`):
  - lower trailing `body.Stmts` in the outer context, wrap in a lambda `$worker:<owner>:$function` (no `desugarNestedFunction`), prepend `waitOnLatch`.
  - declare `$default` slot in `workerFutureSlots[DefaultWorker]`; start it first; its start position spans the trailing statements (closing brace when empty).
  - body ends with `return wait $default`.
  - `workerClosureName` handles the default worker; `workerFutureSlots` is extended (not replaced) by nested bodies.
  - `createLangInternalInvocation` reports `cx.internalError` on missing module/symbol.
- Tests: all `10-worker` pass; `lock1-p` `@panic` moves to `return wait w;` (+ txtar/goldens); `panic1-p` unchanged; add a `-p` with a panic in trailing statements whose trace is unchanged; existing `-v` output order unchanged. Update `corpus/desugared`/`corpus/bir` goldens with `-update`.

## Commit 2.x — type resolver refactors (one commit each)
1. `refactor(types): merge child implicit imports into the parent` — after lambda (2466, 2555) and worker (2417) resolution, merge `ft.implicitImports` into the parent (`addImplicitImport` per entry, respecting ephemeral drop).
2. `refactor(types): give each worker its own xml step owner` — `xmlStepOwner = <owner>$<worker>`; default worker child (step 5) uses `<owner>$function`. Test `-v` printing `<b>1</b><c>2</c>` from two workers.
3. `refactor(types): restore only symbols written by an argument trial` — `argumentStateSnapshotter` records symbols declared inside the argument subtree and restore writes back only those.
4. `refactor(types): own ephemeral state per goroutine` — child resolvers get a copy seeded with the parent's depth; OR `refusedDependent` back into the parent after the child finishes.
5. `refactor(types): resolve default worker statements in a child resolver` — new `resolveDefaultWorkerStatements`; split `resolveNamedWorker` into build + resolve. Every child built before resolution (own atom side table, `semtypes.Context`, `implicitImports`, `monoCounters`/`xmlStepCounter`, owner, ephemeral state). Named children: worker return type, chain as today + function boundary. Default child: function return type, chain with all capture groups, no boundary. Parent atom side table is read-only while children run. Still sequential.
6. `refactor(types): resolve workers of a body concurrently` — one goroutine per child with recover → parent re-raises after join (pattern: `projects/package_compilation.go:146-176`); merge imports/`refusedDependent` after join. While `ResolvePublicNodes` runs (`packageTypeResolver.resolvingPublicNodes`) the children are resolved one after another instead. Tests: `-v` lambda with workers passed to multi-candidate `new` at function and module level; `make test-race`.

## Commit 3 — `fix(parser): keep qualified-name sync send as sync send`
- `parser/parser.go:14290-14297`: `st.CreateSyncSendActionNode(...)`. AST golden for `{ io:x ->> w; }` is added with commit 4a once the node builder builds sends. Until then only the parser changes, verified by a parser-level test if one exists; otherwise the golden lands in 4a. (The spec's test list only asks for the AST golden.)

## Commit 4 — message passing (split into sub-commits; each passes lint/test)
**4a `feat(ast): add worker message action nodes`**
- Remove old `Channel`/`WorkerPairId`/`BLangWorker*` nodes (keep `TypeKindChannel`); add the spec's 7 types in `ast/expressions.go` embedding `bLangActionBase`; `Walk` + pretty printer cases.
- `model.WorkerMessageRef`; `ast.go`: `DefaultWorkerSendMessages`, `BLangNamedWorkerDeclaration.SendMessages`.
- Node builder builds the nodes (`function` keyword → `DefaultWorkerName`, quoted `'function` → `"function"`), alternate receive → unimplemented "alternate receive is not supported", async send in var init / assignment RHS / check / trap / return / match subject / foreach collection / braced → "async send action can only be used as a statement"; compound assignment → "worker message send/receive is not supported here". Allow actions in match subject and foreach collection (currently `createExpression`) and in braced exprs.
- Transient: symbol resolver reports unimplemented "worker message passing is not supported" for the new nodes, so `message-passing1-fe`/`flush1-fe` keep passing (golden diagnostics updated).
- Tests: node-builder `-e` list; parser AST golden `{ io:x ->> w; }`.

**4b `feat(symbols): pair worker messages`**
- `context/internal/workermessages` (Store) + `CompilerEnvironment` atomic counter; `CompilerContext.NewWorkerMessage`, `NewWorkerMessageTypeStore`, public alias.
- `blockSymbolResolver.messageGroup`/`messageWorker`; `declareNamedWorkers` creates the group after `InitStmts`.
- New `symbols/worker_messages.go` exactly per spec (queue sum type with `//sumtype:decl`, `resolveMessagePeer`, flush coverage, count check, interaction-after-wait, simulation, `reportStuck`/`unmatched`/`blockedPos`, `pairWorkerMessages`); visitor cases for the five nodes plus unconditional direct-ref waits.
- Enable `gochecksumtype` in `.golangci.yml` (fix any existing violations).
- Transient: type resolver reports unimplemented on the new nodes.
- Tests: all symbol-resolution `-e`/`-v` cases from spec.

**4c `feat(types): type worker message actions`**
- Resolver fields `messageTypes` + `resolverMessageTypes`; `enterEphemeral` installs a fresh store; children inherit the current store; per-goroutine `PublishPoisonIfUnset` defer; symbol resolution rejects message actions in bodies resolved by `ResolvePublicNodes` (`message-module-level1-fv`).
- Resolvers for 5 actions + `workerFailureType`; types per spec.
- Semantic analyzer: cases in `analyzeActionOrExpression` (Cloneable via `semtypes.CreateCloneable`, uncovered async send).
- CFG: success-termination-before-message-action check per worker graph (named + default, incl. `InitStmts` return).
- Transient: desugar reports unimplemented for message nodes.
- Tests: type, semantic, CFG `-e`/`-v`; refactor-2 post-step-4 test `A|B x = new (function () ... )` at function level.

**4d `feat(worker): run worker message passing`**
- `lang.__internal.bal` + `internal.go`: `workerMessage` type and the 7 externs per spec (yield loops, per-message outcome logic, `GetClaimed` without claim, doc comment).
- Desugar: `functionContext.workerMessages`, `$m` creation in `desugarWorkerRegion`, lowering of 5 nodes, list constructors for handle/future lists, determined types set, termination delivery wrapper for workers with async sends.
- Replace `message-passing1-fe` / `flush1-fe` with `-v`/`-e` tests; add all runtime `-v`/`-p` tests from spec.

## Follow-ups (need user OK before doing)
- File issue for `desugarNestedFunction` not copying `defaultClosureVars`.
- PR description must call out the narrowing behavior change.

## Verification
- Per commit: `make lint test`; from 2.6 on also `make test-race`.
- Corpus goldens via `go test ./... -update` in affected stage packages and `go test ./corpus -update`; review diffs for unexpected changes (only `lock1-p` among existing worker tests).
- Spot-run: `go run ./cli/cmd run corpus/bal/subset10/10-worker/<new>-v.bal`.
