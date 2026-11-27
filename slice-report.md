## Slice 1/2: array lengths reject variables
Review: `f65871f2..03eb2791`
- 03eb2791 fix(types): reject non-constant references as array lengths

Implemented: in `resolveFixedArraySize`, a length expression that is a `*ast.BLangVarRef` or `*ast.BLangConstRef` whose symbol kind is not `SymbolKindConstant` is reported with `notConstantExpressionMessage` ("expression is not a constant expression") and rejected. The check runs after the length expression is resolved and before the singleton check. So a module or local variable with a singleton int type is no longer accepted as an array length (P10, P14). The match const-pattern check now uses the same helper: `isValidConstPatternExpr` was renamed to `isNonConstantReference` with its result inverted, and its behaviour is unchanged.

Shape:
```
semantics/internal/types/type_resolver.go
  resolveFixedArraySize
+   isNonConstantReference(t, lenExp) -> semanticError(notConstantExpressionMessage)   (before the singleton check)
- isValidConstPatternExpr(t, expr) bool
+ isNonConstantReference(t, expr) bool        (inverted; match const pattern caller updated)
corpus/bal/subset10/10-const/array-length-{variable,local}-e.bal          (approved tests, added)
corpus/integration/subset10/10-const/array-length-{variable,local}-e.txtar (new goldens)
corpus/integration/subset8/08-list/fixedlength8-e.txtar                   (message change)
corpus/integration/subset10/10-worker/return-type-invalid1-e.txtar        (message change)
```

Tests now passing: S2 (`corpus/bal/subset10/10-const/array-length-variable-e.bal`), S17 (`corpus/bal/subset10/10-const/array-length-local-e.bal`). Both entries were removed from the pending block in `test_util/skip.go`, which stays uncommitted.

Things to know
- Spec deviations: none
- Added beyond the spec: none. As a side effect, two existing goldens changed. `int` variables used as lengths were already rejected (`fixedlength8-e`, `return-type-invalid1-e`), and they now report "expression is not a constant expression" instead of "fixed-length array size must be a singleton int".
- In S2 the module-level case also reports a follow-on error, "list type not found in expected type", at the initializer. The base commit already reports it for `final int n = 3; int[n] a = ...` (verified by reviewer B), so it is not new.
- Review findings left for you (unmatched between reviewers, or P3): see below.
- Pending tests that already pass: `corpus/bal/subset10/10-configurable/check-default-fp.bal` (S19). As an `-fp` test it expects an unimplemented error on its marked line, and the generic "configurable module variables are not supported yet" error lands there. It is also missing from the pending block in `test_util/skip.go`, so `TestIntegration` fails on it locally with "no golden" while the design state is uncommitted. Not unskipped or changed (design state).

### Deviations
none

### Open review findings
- [P2, reviewer B only; reviewer A reported none] Defensive nil check in `isNonConstantReference`: `semantics/internal/types/type_resolver.go` has `return sym == nil || sym.Kind() != model.SymbolKindConstant`. A reference is already resolved when this runs, so `sym` cannot be nil. nutcracker-review-guidelines, "Compiler invariants": a defensive nil check that hides a broken invariant is a defect. Suggested fix: `return t.getSymbol(ref).Kind() != model.SymbolKindConstant`. The check was carried over (inverted) from the old `isValidConstPatternExpr`. Not fixed, because review-loop does not fix unmatched P1–P3 findings.

Review: one round of review-loop (2 independent reviewers) with no accepted findings, so the review is complete. A commit-shape check found that the commit message overstated the change; the message was reworded (b73baa3d -> 03eb2791, code identical).

### Checks
On 03eb2791 (the only commit):
- `go build ./...`: ok.
- Commit hooks (build + `make lint`): pass. The preinstalled golangci-lint 2.5.0 (built with go1.25) cannot load the config for go1.27.1. I installed golangci-lint v2.15.0 (the CI version) built with `GOTOOLCHAIN=go1.27.1` into `/root/go/bin` and put it on `PATH`. No hook was skipped.
- `go vet` for `./semantics/...`: ok.
- Full suite from `.claude/rules/personal.md`: all packages pass except `./corpus`. There, `TestIntegration` failed on `fixedlength8-e` and `return-type-invalid1-e` (goldens then regenerated for the message change) and on `10-configurable/check-default-fp.bal` (uncommitted design state, see above).
- After regenerating: `go test ./corpus -run 'TestIntegration$'`: the only failure is `check-default-fp.bal` (no golden; the file is not part of the commit).
- `lint_spec.py spec.md --guard --slice-done 1`: 0 problems.
- `lint_spec.py --list-new-tests | nutcracker_check_failing.py` on the slice 2 tests: all 14 fail as predicted (unimplemented); `check-default-fp.bal` already passes.
