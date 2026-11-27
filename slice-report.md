## Slice 1/2: array lengths reject variables

### Commits
- 9059118a fix(types): reject non-constant references as array lengths

### Implemented
`resolveFixedArraySize` now rejects an array length that is a `*ast.BLangVarRef` or `*ast.BLangConstRef` whose symbol kind is not `SymbolKindConstant`. It reports `notConstantExpressionMessage` ("expression is not a constant expression") at the length expression, after resolving it and before the singleton check. A singleton-typed variable (module-level `final`, local, parameter, or a `public final` from another module) is no longer accepted as an array length (§5.4.1.1). Literals and constant references, local or from another module, are still accepted.

The symbol-kind test moved from `isValidConstPatternExpr` into a new helper, `isNonConstantReference`. The match const-pattern check uses the same helper, so its behaviour is unchanged.

### Tests now passing
- S2, an array length must be a constant reference: `corpus/bal/subset10/10-const/array-length-variable-e.bal`
- S17, a local variable is not an array length: `corpus/bal/subset10/10-const/array-length-local-e.bal`

### Deviations
none

### Added beyond the spec
- The message in two existing integration goldens changes from "fixed-length array size must be a singleton int" to "expression is not a constant expression", at the same location. They are `corpus/integration/subset8/08-list/fixedlength8-e.txtar` (`int[x]`, local `int x`) and `corpus/integration/subset10/10-worker/return-type-invalid1-e.txtar` (`int[n]`, local `int n`). Both lengths are non-constant references, so the new check reports them first.
- Process note: the commit-msg hook (`validate_commits.py`) rejects `fixup!` subjects, and hooks may not be bypassed. The round-1 review fix was therefore amended straight into the slice commit, which gives the same result as a fixup followed by the authorized squash.

### Open review findings
- P2 (B) `semantics/internal/types/type_resolver.go:9828`: `isNonConstantReference` calls `t.getSymbol(ref)` and then `Kind()`, keeping a `sym == nil` check. The reviewer cites the CLAUDE.md rule "call operations on symbols via compiler context" and suggests `t.compilerContext().SymbolKind(ref) != model.SymbolKindConstant` without the nil branch. This code came over unchanged from `isValidConstPatternExpr`. Only one reviewer raised it, so it was not fixed.
- P2 (A, outside the slice's scope; seen by both reviewers in round 2 as pre-existing) `semantics/internal/types/type_resolver.go` (module variable init): when the array length of a module-level variable is rejected, the initializer is still checked against the unresolved type. This adds a follow-on "list type not found in expected type" error, which is recorded in `corpus/integration/subset10/10-const/array-length-variable-e.txtar`. The base commit reports the same pair of errors. Suggested as a separate issue.

### Checks
- `go test -count=1 -timeout 50m -p 4 -skip '…' <frontend packages> ./corpus/... ./test_util/...` (the full-suite command from the project rules), on 9059118a: pass. The first run on the pre-review commit hit the default 10-minute timeout in three packages while reviewers were loading the 4-core machine; those packages passed when rerun with a longer timeout.
- Commit hooks (build, `golangci-lint` v2.15.0 installed with go1.27.1 per `.github/workflows/golangci-lint.yml`, lint-config checks): pass.
- `python3 .github/scripts/validate_commits.py "fix(types): reject non-constant references as array lengths"`: pass.
- `lint_spec.py spec.md --guard --slice-done 1`: 0 problems.
- `review-loop`: round 1 had one agreed P2 (inline the one-caller wrapper `isValidConstPatternExpr`), now fixed. Round 2 had no findings from either reviewer.

### Pending tests that already pass
- `corpus/bal/subset10/10-configurable/check-default-fp.bal`: the existing unimplemented error for configurable module variables lands on its marker line.
