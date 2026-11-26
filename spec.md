# Configurable module variables

## Goal

Support `configurable` module variable declarations in the compiler front end (ballerina-nutcracker/ballerina#1741).

- Every compile-time rule the specification places on configurable variables is enforced.
- Each configurable variable is desugared into a call to an internal function that receives:
  - the variable's organization, module name and variable name;
  - a closure that returns the default value, or `()` when the value is required (`?`);
  - the `typedesc` of the variable's type.
- The runtime implementation of that function only panics as unimplemented.

The change starts by fixing an existing hole in constant-context checking (a singleton-typed variable is accepted as an array length), so that a configurable variable is never accepted where a constant is required.

## Non-goals

- Reading `Config.toml`, environment variables or any other configuration source, and any PAL change.
- Runtime validation of a supplied configuration value against the typedesc; the future runtime implementation does that.
- Local (function-level) `configurable` declarations. The parser already rejects them.
- Exempting configurable variables from the unused-variable check. Unused non-public configurables are still reported.
- Table key values as a constant context. Tables are unimplemented: `nodebuilder/node_builder.go:6719` `n.unimplemented("transformTableTypeDescriptor unimplemented", tableTypeDescriptorNode)`.
- Combining `final` with `configurable`. The parser already reports it as implicitly final: `parser/parser.go:1342` `return &common.ERROR_CONFIGURABLE_VAR_IMPLICITLY_FINAL`.
- `?` outside a configurable declaration. The parser already reports it as an invalid token (P2).
- Two older bugs that configurable defaults expose because they are checked against `T & readonly` (decision, final review: separate issues, no change here):
  - A mapping constructor against `R & readonly` ignores `R`'s field defaults, so `configurable R r = {b: 1};` reports "missing non-defaultable required record field" when `R` defaults another field. `final R & readonly r = {b: 1};` fails the same way without this change.
  - A query expression is not built readonly against a readonly expected type, so `configurable int[] x = from int i in [1, 2, 3] select i;` reports "incompatible type: expected readonly&[int...], got [int...]". `final int[] & readonly` fails the same way without this change.

## Repository context

### Parser

- After `=`, a configurable module variable's initializer is parsed by `parseConfigurableVarDeclRhs`: `parser/parser.go:3775` `isConfigurable := isModuleVar && b.isSyntaxKindInList(varDeclQuals, st.CONFIGURABLE_KEYWORD)`.
- There, `?` becomes an `st.RequiredExpressionNode`: `parser/parser.go:3838` `expr = st.CreateRequiredExpressionNode(b.consume())`.
- `?` is never an expression anywhere else.
- When `configurable` is present, the qualifier list holds only `configurable`: `parser/parser.go:1335` `varDeclQualList = []st.STNode{configurableQual}`.

### Node builder

- `transformModuleVariableDeclaration` builds a module variable as `*ast.BLangVariable`: `nodebuilder/node_builder.go:2191` `simpleVar := ast.NewBLangVariable(pos, nameNode.Name, typeNode, expr, isDeclaredWithVar, flags)`.
- Its qualifiers are turned into flags by `moduleVariableFlags`, which today rejects `configurable`: `nodebuilder/node_builder.go:2219` `n.cx.Unimplemented("configurable module variables are not supported yet", pos)`.
- The required expression is dispatched to `transformRequiredExpression` (`nodebuilder/node_builder.go:439` `return n.transformRequiredExpression(node)`), which is unimplemented: `nodebuilder/node_builder.go:5843` `n.unimplemented("transformRequiredExpression unimplemented", requiredExpression)`.
- `getBLangVariableNode` builds only capture and wildcard patterns: `nodebuilder/node_builder.go:2743` `n.unimplemented("binding pattern is not implemented", bindingPattern)`.
- `var` without an initializer is reported with `SyntaxError`: `nodebuilder/node_builder.go:2195` `n.cx.SyntaxError("var-declared module variable must have an initializer expression for type inference", pos)`.

### AST

- The configurable flag and its accessor already exist:
  - `model/flag.go:51` `FlagConfigurable                 Flag = 1 << 29`
  - `ast/ast.go:395` `func (b *bLangVariableBase) IsConfigurable() bool     { return b.flags.Has(model.FlagConfigurable) }`
- Record field defaults keep the symbol and scope of the closure that desugar generates: `ast/types.go:118` `FnRef   model.SymbolRef`.
- A leaf expression node such as `BLangTypedescExpr` is handled in these places:
  - `ast/walk.go:485` `case *BLangTypedescExpr:`
  - `ast/set_type.go:83` `case *BLangTypedescExpr:`
  - `ast/pretty_printer.go:193` `case *BLangTypedescExpr:`
  - `semantics/internal/symbols/symbol_resolver.go:2781` `case *ast.BLangTypedescExpr:`
  - `semantics/internal/types/type_resolver.go:4213` `case *ast.BLangTypedescExpr:`
  - `semantics/internal/analysis/expressions.go:109` `case *ast.BLangTypedescExpr:`
  - `desugar/expression.go:90` `case *ast.BLangTypedescExpr:`
- Each expression switch ends in an internal error for unknown nodes: `semantics/internal/types/type_resolver.go:4280` `t.internalError(fmt.Sprintf("unsupported expression type: %T", expr), expr.GetPosition())`.

### Symbol resolution

- `allocateGlobalVarSymbol` copies the configurable and final flags to the symbol: `semantics/internal/symbols/symbol_resolver.go:999` `if globalVar.IsConfigurable() {`.
- Module variables are never const symbols: `semantics/internal/symbols/symbol_resolver.go:995` `symbol := model.NewVariableSymbol(name, isPublic, false, false, globalVar.Name.GetPosition())`.
- Non-public ones are tracked for the unused-variable check: `semantics/internal/symbols/symbol_resolver.go:982` `if !isPublic {`.
- Record default closures get a function symbol named by `nextDefaultSymbolName` in `recordDefaultScope()`, which at module level is the module scope, plus a fresh function scope. See `semantics/internal/symbols/symbol_resolver.go:1905` `field.Default.FnRef = symRef` and `semantics/internal/symbols/symbol_resolver.go:427` `return ms.scope`.

### Type resolution

- The declared type of a module variable becomes its symbol type: `semantics/internal/types/type_resolver.go:3851` `updateSymbolType(t, node, semType)`.
- The initializer is resolved against the symbol type: `semantics/internal/types/type_resolver.go:3878` `_, ok := resolveActionOrExpression(t, nil, node.Expr, expectedType)`.
- Semantic analysis validates the initializer against the symbol type: `semantics/internal/analysis/declarations.go:64` `validateVariable(cx, v, moduleVarExpectedType(v.IsListener(), cx.CompilerCtx().SymbolType(v.Symbol())))`.
- Existing patterns:
  - the readonly intersection for constants: `semantics/internal/types/type_resolver.go:9632` `t.setSymbolType(symbol, semtypes.Intersect(expectedType, semtypes.ValReadonly))`
  - the anydata check: `semantics/internal/types/type_resolver.go:9639` `return !semtypes.IsEmpty(cx, ty) && semtypes.IsSubtype(cx, ty, semtypes.CreateAnydata(cx))`
  - the emptiness check: `semantics/internal/types/type_resolver.go:8915` `if !ensureNotEmpty(t, result, func() {`
- Symbol resolution already rejects a type reference that names anything but a type, a constant or a parameter: `semantics/internal/symbols/symbol_resolver.go:1667` `resolver.GetCtx().SemanticError("Unknown type: "+name, n.GetPosition())`. This stops the pipeline before type resolution (stage 3), so it needs no change; S18 covers it for a configurable.
- Array lengths are checked by type only: `semantics/internal/types/type_resolver.go:8734` `t.semanticError("fixed-length array size must be a singleton int", lenExp.GetPosition())`. A singleton-typed variable is accepted (P10, P14).
- Match const patterns already require a constant symbol: `semantics/internal/types/type_resolver.go:9823` `return sym != nil && sym.Kind() == model.SymbolKindConstant`.
- Enum member values and xmlns URIs already use `isConstantExpression`, which rejects non-const variables: `semantics/internal/types/constant_evaluator.go:44` `if vs, ok := t.getSymbol(e.Symbol()).(model.ValueSymbol); ok && vs.IsConst() {`.

### Semantic analysis

- Assigning to a final variable is rejected, except for deferred initialization in `init` of a variable that has no initializer: `semantics/internal/analysis/statements.go:335` `deferredModuleInit := !compound && inFunction && fn.IsModuleInit() && moduleVar && md.HasNoInitializer()`.
- "No initializer" is `Expr == nil` here: `semantics/internal/analysis/scope/context.go:214` `v.Flags().Has(model.FlagIsolated), v.Expr == nil)`. It is the same in `semantics/internal/cfg/final_reassign_analyzer.go:141` `if v.Expr == nil && v.IsFinal() {` and `semantics/internal/cfg/uninit_field_analyzer.go:110` `if pkg.GlobalVars[i].Expr == nil {`.
- Isolated functions may read configurable variables of an isolated type: `semantics/internal/analysis/scope/isolation.go:48` `if (md.IsFinal() || md.isConfigurable()) && !md.IsIsolated() && semtypes.IsSubtype(tyCtx, md.Type(), isolatedTy) {`. Every readonly type is isolated: `semtypes/core.go:838` `cx._isolatedMemo = Union(ValReadonly, createIsolatedObject(cx))`.

### Desugar

- `desugarInitFn` collects each global's initializer and sorts the initializers by dependency. It emits an assignment for each one (`desugar/desugar.go:725` `initStmts = append(initStmts, buildInitAssignment(compilerCtx, node))`) and then clears them: `desugar/desugar.go:495` `pkg.GlobalVars[i].Expr = nil`.
- The resulting init function is desugared as a normal function: `desugar/desugar.go:740` `initFn, generatedFunctions := desugarFunction(pkgCtx, "", pkg.InitFunction, span)`.
- Calls to `lang.__internal` are built by `createLangInternalInvocation`, which sets the call's type to its `returnTy`: `desugar/query_expression.go:1953` `func createLangInternalInvocation(`.
- Record field defaults inside functions show how a closure is built:
  - a function from the pre-allocated symbol: `desugar/desugar.go:1237` `func createDefaultValueFunction(name string, defaultExpr ast.BLangExpression, requiredParams []ast.BLangVariable) *ast.BLangFunction {`
  - its signature: `desugar/desugar.go:1402` `ctx.compilerCtx.SetFunctionTypedSignature(fnRef, model.TypedFunctionSignature{ReturnType: fieldTy})`
  - its function type: `desugar/desugar.go:1304` `func recordFieldDefaultFnType(env semtypes.Env, fieldTy semtypes.SemType) semtypes.SemType {`
  - the lambda that wraps it: `desugar/desugar.go:1273` `lambda := &ast.BLangLambdaFunction{Function: fn}`
- Walking a lambda desugars its body: `desugar/expression.go:853` `expr.Function = desugarNestedFunction(cx, expr.Function)`.
- Literal and typedesc helpers:
  - `desugar/query_expression.go:1867` `func createStringLiteral(value string, pos diagnostics.Location) *ast.BLangLiteral {`
  - `desugar/expression.go:419` `func createNilLiteral(pos diagnostics.Location) *ast.BLangLiteral {`
  - a typedesc value is a `BLangTypedescExpr` with its `Constraint` set (`semantics/internal/types/type_resolver.go:8429` `expr := &ast.BLangTypedescExpr{Constraint: inferred}`), typed `typedesc<constraint>` (`semantics/internal/types/type_resolver.go:4314` `ty := semtypes.TypedescContaining(t.typeEnv(), constraint)`)

### BIR and runtime

- A call's result temporary takes the call node's type, not the callee's return type: `birgen/bir_gen.go:2111` `return finishCall(ctx, effect, callable.Type())`.
- `lang.__internal` functions are declared `external` (`lib/langlibs/ballerina/lang.__internal/0.0.1/any/lang.__internal.bal:17` `public isolated function querySort(any[] sortKeyRows, any[] sortDirections, any[] rowIndices, any[] payloadRows) = external;`). They are registered in Go: `lib/langlibs/go/lang.__internal/internal.go:39` `runtime.RegisterExternFunction(rt, orgName, moduleName, "querySort", func(ctx *extern.Context, args []values.BalValue) (values.BalValue, error) {`.
- A native function that returns an error panics. During module initialization the panic's first frame is `init` at the initializer's line (P12).

### Documentation

- The subset document lists module variable declarations with no restrictions: `doc/lang/subset10.md:17` `- [Module variable declarations](https://ballerina.io/spec/lang/master/#module-var-decl)`.

## Guardrails

- Do not change PAL or add any configuration reading.
- Keep `?` representable only inside a configurable module variable. The parser guarantees this, and no other construct may create a `BLangRequiredDefaultValueExpr`.
- A configurable variable always has a non-nil `Expr`, so the existing "no initializer" paths (deferred init in `init`, the uninitialized-variable and final-reassignment analyzers) never apply to it.
- Do not change the unused-variable rules.
- Slice 1 is its own commit and must not mention configurable variables.
- Desugar must not allocate symbol spaces or scopes; allocate them in symbol resolution, as record field defaults do.
- Store symbols in maps only by `model.SymbolRef`.

## API

### ast/expressions.go

Added:

```go
// BLangRequiredDefaultValueExpr is the `?` initializer of a configurable module
// variable: the value must be supplied by the configuration. Its type is the
// contextually expected type, the variable's declared type intersected with readonly.
BLangRequiredDefaultValueExpr struct {
	bLangExpressionBase
}
```

It implements `BLangExpression` and `BLangActionOrExpression` like other leaf expressions. It is handled in `ast/walk.go` as a leaf, in `ast/set_type.go`, and in `ast/pretty_printer.go`, where it prints as `(required-default-value)`.

Added (decision, slice 2: the spec first added a `ConfigurableDefault` field to `BLangVariable`; the user chose to keep everything about a configurable's initializer in `Expr`, as for any module variable and as `?` already is):

```go
// BLangConfigurableDefaultExpr is the initializer of a configurable module
// variable that has a default value. Expr is the default value; FnRef and FnScope
// are the symbol and scope of the closure desugar generates to evaluate it,
// allocated by symbol resolution. Its type is the type of Expr.
BLangConfigurableDefaultExpr struct {
	bLangExpressionBase
	Expr    BLangExpression
	FnRef   model.SymbolRef
	FnScope model.Scope
}
```

It is walked in `ast/walk.go` with `Expr` as its child, so the module-initializer dependency walk sees the default's references. It is handled in `ast/set_type.go` and in `ast/pretty_printer.go`, where it prints as `(configurable-default <expr>)`. `BLangVariable` is unchanged.

### nodebuilder/node_builder.go

Changed:

- `moduleVariableFlags`: `CONFIGURABLE_KEYWORD` sets `model.FlagConfigurable | model.FlagFinal` instead of reporting unimplemented.
- `transformRequiredExpression`: returns `&ast.BLangRequiredDefaultValueExpr{}` positioned at the `?` token.
- `transformModuleVariableDeclaration`: the initializer of a configurable declaration that is not `?` is wrapped in a `BLangConfigurableDefaultExpr` positioned at the initializer.
- `transformModuleVariableDeclaration`: for a configurable declaration, these checks run before the binding pattern is built:
  - With `var`, it calls `n.cx.SyntaxError("configurable variable must have an explicit type descriptor", <position of the var keyword>)` and returns.
  - With a binding pattern other than a capture binding pattern, it calls `n.cx.SyntaxError("configurable variable must be declared with a variable name", <position of the binding pattern>)` and returns.

### semantics/internal/symbols/symbol_resolver.go

Changed: the expression walker's case for `*ast.BLangConfigurableDefaultExpr` sets its `FnRef` to a function symbol named `resolver.nextDefaultSymbolName()`, added to `recordDefaultScope()`, then resolves `Expr`. The scope is `GetCtx().NewBlockScope(recordDefaultScope(), GetPkgID())`. This mirrors `allocateRecordDefaultSymbols`. (Decision, slice 2: the spec first named `NewFunctionScope`, which does not exist; the user chose to amend it to `NewBlockScope`.)

The expression walker gets a no-op case for `*ast.BLangRequiredDefaultValueExpr`.

### semantics/internal/types/type_resolver.go

Changed:

- `resolveGlobalVarType`: for a configurable variable, after resolving the declared type `T`:
  - If `T` is not a subtype of anydata, call `semanticError("configurable variable type must be a subtype of 'anydata'", <type node position>)`, set the symbol type to `semtypes.Never` and return false.
  - Otherwise compute `R = semtypes.Intersect(T, semtypes.ValReadonly)`. If `R` is empty (checked with `ensureNotEmpty(t, R, ...)`), report `"configurable variable type has no readonly values: intersection with readonly is empty"` at the type node position. Otherwise set the symbol type to `R`.
- `resolveExpressionInner`: for `*ast.BLangRequiredDefaultValueExpr`, sets its type to `expectedType` (via `setNodeType`) and returns that type. For `*ast.BLangConfigurableDefaultExpr`, resolves `Expr` against `expectedType` and takes its type.
- `resolveFixedArraySize`: if `lenExp` is a `*ast.BLangVarRef` or `*ast.BLangConstRef` whose symbol kind is not `SymbolKindConstant`, report `notConstantExpressionMessage` at `lenExp`'s position and return false. This check runs before the singleton check.

### semantics/internal/analysis/expressions.go

Changed: `validateActionOrExpr` accepts `*ast.BLangRequiredDefaultValueExpr` and returns empty facts.

### semantics/internal/analysis/expressions.go and scope/context.go

Changed (decision, slice 2): `validateActionOrExpr` validates the `Expr` of a `*ast.BLangConfigurableDefaultExpr` in `scope.NewExpressionClosure(cx, scope.ConfigurableDefault)`, a new non-isolated `ClosureKind`, because desugar evaluates it in a closure. The existing rule for expression closures then reports `check` in the default as unimplemented (`"'check' in a default expression is not supported"`), as it does for record field defaults and defaultable parameters (S19). Supporting `check` there would need the closure to return `R|error` and `configurableValue` to propagate the error; the user chose to defer it.

### lib/langlibs/ballerina/lang.__internal/0.0.1/any/lang.__internal.bal

Added:

```ballerina
# Returns the value of a configurable module variable. The runtime looks the variable up by
# organization, module and variable name in the configuration sources (configuration files or
# environment variables), validates the supplied value against `td`, and returns it as a
# read-only value. If no value is supplied, it returns the result of `defaultSupplier`, which is
# called only in that case; if `defaultSupplier` is `()`, the value is required and module
# initialization fails.
public function configurableValue(string org, string module, string name,
        (function () returns anydata)? defaultSupplier, typedesc<anydata> td) returns anydata = external;
```

It is not `isolated`, because the default value need not be an isolated expression.

### lib/langlibs/go/lang.__internal/internal.go

Added: a native `configurableValue` registered in `initInternalModule`. It returns `nil, errors.New("configurable variables are not supported yet")`, which the runtime turns into a panic.

### desugar/desugar.go

Changed: `desugarInitFn` / `buildInitAssignment`. For a configurable variable, the assignment's expression is:

```
lang.__internal:configurableValue("<org>", "<module>", "<name>", <default>, <typedesc of R>)
```

- `R` is the variable's symbol type.
- `<typedesc of R>` is `&ast.BLangTypedescExpr{Constraint: R}` with `Ty` set to `semtypes.TypedescContaining(env, R)`.
- The call's `Ty` is `R`.
- Whether the variable is configurable is read from its symbol, as for listeners; the initializer from `moduleInitNode.expr`.
- `<default>` is `()` for `?`. For a `BLangConfigurableDefaultExpr` it is a `BLangLambdaFunction` whose function comes from `createDefaultValueFunction(<FnRef name>, <Expr>, nil)`, with its symbol and scope set from `FnRef` and `FnScope`. Its typed signature is `{ReturnType: R}`, set as `setRecordDefaultFnSignature` does. Its lambda type is `recordFieldDefaultFnType(env, R)`.

## Design

### Front end

1. **Node builder.**
   - A configurable declaration gets `FlagConfigurable | FlagFinal`, so the existing final-assignment check rejects every assignment, including in `init` (S11).
   - The `var` check and the binding-pattern check run before the generic "binding pattern is not implemented" error (S6, S7).
   - `?` becomes `BLangRequiredDefaultValueExpr` and a default becomes `BLangConfigurableDefaultExpr`, so `Expr` is never nil for a configurable variable.
2. **Symbol resolution.**
   - The symbol already gets the configurable and final flags from the AST.
   - Allocate the closure symbol and scope of a `BLangConfigurableDefaultExpr`.
   - Unused tracking is unchanged (S16).
3. **Type resolution.**
   - `resolveGlobalVarType` runs the anydata check (S8) and the empty-intersection check (S9), then stores `T & readonly` as the symbol type.
   - `resolveGlobalVarInit` already uses the symbol type as the expected type. The default is therefore checked against `T & readonly`, and a list or mapping constructor builds a readonly value (S3, S10).
   - `?` takes the expected type.
4. **Analysis.**
   - `validateVariable` validates the default against `T & readonly`.
   - Isolation already lets isolated functions read the variable, because `T & readonly` is a readonly type (S3).
   - The default is not checked for isolation (S4).
5. **Constant contexts** (slice 1 for all variables; the configurable cases are S12, S13 and S18):
   - Type references to a configurable are rejected by the existing symbol-resolution check (S18).
   - Array lengths must be a literal or a constant reference (S2).
   - Match const patterns, enum member values and xmlns URIs already reject non-constants.

### Desugar

`collectModuleInitNodes` and the topological sort are unchanged: the dependency walk sees the original default expression, so dependencies are kept (S4). When the assignment is built, a configurable variable's expression is replaced by the `configurableValue` call described in the API. The lambda is desugared when `desugarFunction` walks the init function. `BLangRequiredDefaultValueExpr` never reaches the function walkers or BIR generation.

For a single-file program the organization and module are taken from `pkg.PackageID.OrgName.Value()` and `pkg.PackageID.PkgName.Value()`. `PkgName` is the full module name, for example `proj.util`.

### Runtime

`configurableValue` returns an error, so every program that initializes a configurable variable panics at that declaration's line, with first frame `init` (P12). Runtime tests are therefore `-p` tests marked on the declaration line (S3, S4, S5, S14).

### Completion

The last slice updates `doc/lang/subset10.md` and removes the pending block from `test_util/skip.go`. Each slice regenerates its tests' goldens (`corpus/ast`, `corpus/cfg`, `corpus/desugared`, `corpus/bir`, `corpus/integration`) with the drivers listed in `.agents/skills/manage-corpus-tests/SKILL.md`.

### Failure handling

All new diagnostics are `SyntaxError` (node builder) or `SemanticError` (type resolution). The type-resolution failure for a configurable stops the pipeline before stage 5, as other stage 3–4 errors do.

### Documentation

In `doc/lang/subset10.md`, under "Module variable declarations", add this restriction:

- The runtime does not support configurable variables yet: initializing a configurable variable panics.

(Decision, slice 2: a bullet "`configurable` module variables are supported in the front end" was dropped, because subset sub-bullets state only restrictions.)

Under "Checking expression", extend the existing restriction on `check` in default expressions to the default values of configurable variables (S19).

## Test scenarios

### S2: an array length must be a constant reference
- Rule: §5.4.1.1: `array-length := int-literal | constant-reference-expr | inferred-array-length`. A constant reference must refer to a `module-const-decl` (§6.4).
- Test: `corpus/bal/subset10/10-const/array-length-variable-e.bal` (new)
- Expected: an error at `int[n] a = [1, 2, 3];`, where `n` is `final 3 n = 3`.
- Fails today: diagnostic-mismatch
- Slice: 1

### S17: a local variable is not an array length
- Rule: as S2.
- Test: `corpus/bal/subset10/10-const/array-length-local-e.bal` (new)
- Expected: an error at `int[n] a = [1, 2, 3];`, where `n` is a local `3 n = 3`.
- Fails today: diagnostic-mismatch
- Slice: 1

### S3: a configurable with a default compiles and has type T & readonly
- Rule: §8.7: the static type is `T & readonly`, so an isolated function may reference the variable (§7.3.5) and its value can be assigned to `int[] & readonly`. The runtime panics during module initialization (agreed decision).
- Test: `corpus/bal/subset10/10-configurable/readonly-isolated-p.bal` (new)
- Expected: compiles; panics at the declaration line in `init`.
- Fails today: unimplemented
- Slice: 2

### S4: a default need not be isolated and may depend on module state
- Rule: §8.7 does not require the default to be an isolated expression. Initializer dependency order is preserved.
- Test: `corpus/bal/subset10/10-configurable/non-isolated-default-p.bal` (new)
- Expected: compiles; panics at the configurable's declaration line.
- Fails today: unimplemented
- Slice: 2

### S5: a configurable with a default value
- Rule: §8.7 `module-var-init := expression`. The runtime is unimplemented.
- Test: `corpus/bal/subset10/10-configurable/default-p.bal` (new)
- Expected: compiles; panics at `configurable int port = 9090;`.
- Fails today: unimplemented
- Slice: 2

### S14: a required public configurable
- Rule: §8.7 `module-var-init := ?`, allowed with `configurable`. `public` is allowed.
- Test: `corpus/bal/subset10/10-configurable/required-p.bal` (new)
- Expected: compiles; panics at `public configurable string host = ?;`.
- Fails today: unimplemented
- Slice: 2

### S6: a configurable needs an explicit type
- Rule: §8.7: the typed binding pattern of a configurable must use an explicit type descriptor, not `var`.
- Test: `corpus/bal/subset10/10-configurable/var-e.bal` (new)
- Expected: an error at `configurable var port = 9090;`.
- Fails today: unimplemented
- Slice: 2

### S7: a configurable's binding pattern must be a variable name
- Rule: §8.7: the binding pattern must be just a variable name.
- Test: `corpus/bal/subset10/10-configurable/binding-pattern-e.bal` (new)
- Expected: an error at the list binding pattern declaration.
- Fails today: unimplemented
- Slice: 2

### S8: a configurable's type must be a subtype of anydata
- Rule: §8.7: the type must be a subtype of `anydata`. Function and error types are not.
- Test: `corpus/bal/subset10/10-configurable/non-anydata-e.bal` (new)
- Expected: errors at both declarations.
- Fails today: unimplemented
- Slice: 2

### S9: the readonly intersection must not be empty
- Rule: §8.7 and §5.7.8: `never & readonly` denotes an empty set of shapes.
- Test: `corpus/bal/subset10/10-configurable/empty-readonly-e.bal` (new)
- Expected: an error at `public configurable never n = ?;`. The variable is public so that nothing has to reference a `never`-typed variable.
- Fails today: unimplemented
- Slice: 2

### S10: the default must belong to T & readonly
- Rule: §8.7: the default initializes a variable whose static type is `T & readonly`. A string is not an int, and a mutable `int[]` value is not `int[] & readonly`.
- Test: `corpus/bal/subset10/10-configurable/default-type-e.bal` (new)
- Expected: errors at both configurable declarations.
- Fails today: unimplemented
- Slice: 2

### S11: a configurable cannot be assigned
- Rule: §8.7: a configurable is implicitly final; nothing may assign to it outside its declaration, including `init`.
- Test: `corpus/bal/subset10/10-configurable/assign-e.bal` (new)
- Expected: errors at the two assignments in `init` and the two in `main`. The variables are public, so whether a write counts as a use does not matter.
- Fails today: unimplemented
- Slice: 2

### S12: a configurable is not a constant
- Rule: §6.4 and §5.4.1.1: array lengths, enum member values and xmlns URIs need a constant reference.
- Test: `corpus/bal/subset10/10-configurable/const-context-e.bal` (new)
- Expected: errors at `int[n] a`, `A = uri` and `xmlns uri as ns;`. The variables are public, so whether references in erroneous positions count as uses does not matter.
- Fails today: unimplemented
- Slice: 2

### S13: a configurable is not a match const pattern
- Rule: a const pattern's variable reference must refer to a constant (§6.4).
- Test: `corpus/bal/subset10/10-configurable/match-const-e.bal` (new)
- Expected: an error at `n => {`.
- Fails today: unimplemented
- Slice: 2

### S18: a configurable is not a type
- Rule: §5.7.1 and §5.7.2: a type reference must name a type definition, and a singleton type's constant reference must refer to a `module-const-decl` (§6.4).
- Test: `corpus/bal/subset10/10-configurable/type-ref-e.bal` (new)
- Expected: an error at `type T n;`. It is in its own test because this error comes from symbol resolution, which stops the pipeline before the type-resolution errors of S12.
- Fails today: unimplemented
- Slice: 2

### S16: an unused configurable is reported
- Rule: an unused non-public module variable is an error in this implementation (agreed decision; P8).
- Test: `corpus/bal/subset10/10-configurable/unused-e.bal` (new)
- Expected: an error at `configurable int port = 9090;`.
- Fails today: unimplemented
- Slice: 2

### S19: `check` in a configurable default is unimplemented
- Rule: §8.7 allows any expression as the default, including `check`. Its error would have to fail module initialization, which the default closure cannot propagate yet (decision, slice 2: reject as unimplemented, like `check` in record field defaults, #1219).
- Test: `corpus/bal/subset10/10-configurable/check-default-fp.bal` (new)
- Expected: an unimplemented error at `public configurable int port = check int:fromString("9090");`.
- Fails today: unimplemented
- Slice: 2

## Slices

### Slice 1: array lengths reject variables
- Scenarios: S2, S17
- Depends on: none

### Slice 2: configurable module variables
- Scenarios: S3, S4, S5, S6, S7, S8, S9, S10, S11, S12, S13, S14, S16, S18, S19
- Depends on: Slice 1

## Success criteria

- A variable cannot be used as an array length (S2, S17).
- Configurable variables compile and desugar to `lang.__internal:configurableValue`, which panics at runtime (S3, S4, S5, S14).
- Every structural and type rule of §8.7 is enforced (S6, S7, S8, S9, S10).
- Configurable variables are final, not constants and not types (S11, S12, S13, S18).
- Unused configurables are reported as before (S16).
- `check` in a configurable default is reported as unimplemented (S19).
- The desugared goldens of the slice 2 `-p` tests show the `configurableValue` call with organization, module, name, default closure or `()`, and typedesc (S3, S4, S5, S14).

## Probes

The `bal` binary in each command was built from this checkout with `go build -o /private/tmp/claude-501/probes/bal ./cli/cmd`.

### P2: `?` outside a configurable is an invalid token
Command: `/private/tmp/claude-501/probes/bal run /private/tmp/claude-501/probes/p2.bal`
Output:
```text
error[SYNTAX_ERROR]: invalid token
 --> p2.bal:1:10
```

### P8: an unused non-public module variable
Command: `/private/tmp/claude-501/probes/bal run /private/tmp/claude-501/probes/p8.bal`
Output:
```text
error[SEMANTIC_ERROR]: unused variable 'x'
 --> p8.bal:1:1
```

### P10: a singleton-typed final variable as an array length
Command: `/private/tmp/claude-501/probes/bal run /private/tmp/claude-501/probes/al.bal`
Output:
```text
[1,2,3]
```

### P12: a native panic during module initialization
Command: `/private/tmp/claude-501/probes/bal run /private/tmp/claude-501/probes/ip.bal`
Output:
```text
error: string index out of range: startIndex=5 endIndex=9 length=3
        at init(ip.bal:4)
```

### P14: a singleton-typed local variable as an array length
Command: `/private/tmp/claude-501/probes/bal run /private/tmp/claude-501/probes/q2.bal`
Output:
```text
[1,2,3]
```

## Approved tests

- `2ae0146b661f961e9ff77bd00756226f4804aa88dd496fdd3a668c2929448f02` `corpus/bal/subset10/10-const/array-length-variable-e.bal`
- `664f80f5f4e7804522fafc47d9c6ca5d254e11acd7f421d63f6780d68ab68dd1` `corpus/bal/subset10/10-configurable/readonly-isolated-p.bal`
- `48a06e30ac739e911d83223737f4a2cf58b6b174845d4bbe29ff57f5d6629fc0` `corpus/bal/subset10/10-configurable/non-isolated-default-p.bal`
- `967667cd50a2847b9f2ad39e64aed04bc2a7ed77c757319d3fd6d74143ced9db` `corpus/bal/subset10/10-configurable/default-p.bal`
- `1e8afab18c7b846469896c43346b86825e81db36d3f932578f5bf4325635c9ef` `corpus/bal/subset10/10-configurable/var-e.bal`
- `0dc8f4debdaac84fa39e0219db5673d5bb88ddece3c762f71a906eeee09b551d` `corpus/bal/subset10/10-configurable/binding-pattern-e.bal`
- `19cfca970fd302e8430208c7e4af97a78206d7148f7b02e4d8f110106d9131ee` `corpus/bal/subset10/10-configurable/non-anydata-e.bal`
- `930f0b0dcd2028eaad3504461ae9270bbc8f3efddb72a4d7d4daf11436abddc4` `corpus/bal/subset10/10-configurable/empty-readonly-e.bal`
- `e50020ff91523e7147f53966f0db355a6f8fd82fae377fed11576dccd0efaf5c` `corpus/bal/subset10/10-configurable/default-type-e.bal`
- `89fbc7df2b4502d208a39ea17e134efe17d7e03cf9fea0ad3ce60099c045c66b` `corpus/bal/subset10/10-configurable/assign-e.bal`
- `eda0faee20c7c341a7dad0b413915b3e14a60166129a4b597c8a372a56cee0e2` `corpus/bal/subset10/10-configurable/const-context-e.bal`
- `bfbc6296d5749e6f429e581e6e2d571a6d90c7b2171befe011a235c39301f15e` `corpus/bal/subset10/10-configurable/match-const-e.bal`
- `6ae30e6e8051a597952c3cda20ea731e0c9b522c0e9591ce9b1e74f91aeb6297` `corpus/bal/subset10/10-configurable/required-p.bal`
- `ab1dc498fb2d0ca2c6461a6e8b3d47b08b04be49fb12e80c8e2bd65113dfc55e` `corpus/bal/subset10/10-configurable/unused-e.bal`
- `901277e43516f2303f63298a12de8e4cafb7ded24e90d96cace2c3d3ff9bfade` `corpus/bal/subset10/10-const/array-length-local-e.bal`
- `da761bd51d99048c684983919417ff34e79c96904a3d51a784490dacf59cb00c` `corpus/bal/subset10/10-configurable/type-ref-e.bal`
- `0ce18eabf5731eb59072f2d919eac9fea908a09f49ec120803150cf80e00246e` `corpus/bal/subset10/10-configurable/check-default-fp.bal`
