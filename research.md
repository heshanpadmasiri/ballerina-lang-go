# Research: configurable module variables

## How it works today

### Parsing

- After `=`, a module variable whose qualifiers include `configurable` is parsed by `parseConfigurableVarDeclRhs`; every other module variable is parsed with `parseExpression`: `parser/parser.go:3775` `isConfigurable := isModuleVar && b.isSyntaxKindInList(varDeclQuals, st.CONFIGURABLE_KEYWORD)`.
- In that path `?` becomes a `REQUIRED_EXPRESSION` node: `parser/parser.go:3838` `expr = st.CreateRequiredExpressionNode(b.consume())`.
- Outside a configurable declaration, `?` is not an expression. The parser reports it as an invalid token (P2).
- If `configurable` appears with other qualifiers, the declaration keeps only `configurable`; the others become invalid minutiae on it: `parser/parser.go:1335` `varDeclQualList = []st.STNode{configurableQual}`. A `final` there is reported as implicitly final: `parser/parser.go:1342` `return &common.ERROR_CONFIGURABLE_VAR_IMPLICITLY_FINAL`.
- The parser accepts `configurable var x = 1;` and destructuring patterns such as `configurable [int, string] [a, b] = ?;` without a diagnostic. The parser subagent established this with ST dumps; I did not re-run them.

### AST building

- The node builder rejects every `configurable` module variable. It fails before building the binding pattern or the initializer: `nodebuilder/node_builder.go:2219` `n.cx.Unimplemented("configurable module variables are not supported yet", pos)`. See P1.
- `final` and `isolated` set flags: `nodebuilder/node_builder.go:2215` `flags |= model.FlagFinal`.
- `?` has no AST form. It is unreachable today, because the qualifier check above returns first: `nodebuilder/node_builder.go:5843` `n.unimplemented("transformRequiredExpression unimplemented", requiredExpression)`.
- A module variable is a `*ast.BLangVariable`, which holds the type node, `Expr`, flags and `Name`: `ast/ast.go:286` `BLangVariable struct {`. The configurable flag already exists: `model/flag.go:51` `FlagConfigurable                 Flag = 1 << 29`. So does its accessor: `ast/ast.go:395` `func (b *bLangVariableBase) IsConfigurable() bool     { return b.flags.Has(model.FlagConfigurable) }`.
- Only capture and wildcard binding patterns are built. All others are unimplemented: `nodebuilder/node_builder.go:2743` `n.unimplemented("binding pattern is not implemented", bindingPattern)`.

### Symbol resolution

- The symbol resolver already copies the flag to the symbol: `semantics/internal/symbols/symbol_resolver.go:999` `if globalVar.IsConfigurable() {`.
- Non-public module variables are registered with the unused-variable tracker: `semantics/internal/symbols/symbol_resolver.go:982` `if !isPublic {`. A non-public module variable that is never referenced is a semantic error (P8).
- Each record field default gets a function symbol and a scope here, ready for the closure desugar generates: `semantics/internal/symbols/symbol_resolver.go:1905` `field.Default.FnRef = symRef`. They are stored on `BFieldDefault`: `ast/types.go:118` `FnRef   model.SymbolRef`. The comment explains why: desugar runs concurrently and must not allocate symbol spaces.

### Type resolution and semantic analysis

- The declared type becomes the symbol type: `semantics/internal/types/type_resolver.go:3851` `updateSymbolType(t, node, semType)`.
- The initializer is resolved with the symbol type as its expected type: `semantics/internal/types/type_resolver.go:3878` `_, ok := resolveActionOrExpression(t, nil, node.Expr, expectedType)`. It is validated against the same type in analysis: `semantics/internal/analysis/declarations.go:64` `validateVariable(cx, v, moduleVarExpectedType(v.IsListener(), cx.CompilerCtx().SymbolType(v.Symbol())))`.
- Constants intersect their type with readonly: `semantics/internal/types/type_resolver.go:9632` `t.setSymbolType(symbol, semtypes.Intersect(expectedType, semtypes.ValReadonly))`. No code does this for `final` variables.
- Constants check their type against anydata: `semantics/internal/types/type_resolver.go:9639` `return !semtypes.IsEmpty(cx, ty) && semtypes.IsSubtype(cx, ty, semtypes.CreateAnydata(cx))`.
- An empty intersection type descriptor is reported as `intersection type is empty (equivalent to never)`: `semantics/internal/types/type_resolver.go:8916` `t.semanticError("intersection type is empty (equivalent to never)", pos)`. See P9.
- Assigning to a final variable is an error everywhere: `semantics/internal/analysis/statements.go:339` `cx.SemanticErr(strings.Join([]string{"cannot assign a value to final variable '", ctx.SymbolName(sym), "'"}, ""), ref.GetPosition())`. The one exception is a final module variable that has no initializer, assigned in the module `init` function: `semantics/internal/analysis/statements.go:335` `deferredModuleInit := !compound && inFunction && fn.IsModuleInit() && moduleVar && md.HasNoInitializer()`. See P3.
- "No initializer" means `Expr == nil` in several places:
  - the scope metadata: `semantics/internal/analysis/scope/context.go:214` `v.Flags().Has(model.FlagIsolated), v.Expr == nil)`
  - the final reassignment analyzer: `semantics/internal/cfg/final_reassign_analyzer.go:141` `if v.Expr == nil && v.IsFinal() {`
  - the uninitialized-variable analyzer: `semantics/internal/cfg/uninit_field_analyzer.go:110` `if pkg.GlobalVars[i].Expr == nil {`
  - initializer resolution: `semantics/internal/types/type_resolver.go:3859` `if node.Expr == nil {`

### Isolation

- An isolated function may read a module variable that is `final` or configurable, not `isolated`, and of an isolated type: `semantics/internal/analysis/scope/isolation.go:48` `if (md.IsFinal() || md.isConfigurable()) && !md.IsIsolated() && semtypes.IsSubtype(tyCtx, md.Type(), isolatedTy) {`.
- The isolated type includes every readonly type: `semtypes/core.go:838` `cx._isolatedMemo = Union(ValReadonly, createIsolatedObject(cx))`.
- P4 shows the rule: an isolated function may read a `final int` but not a plain `int`.

### Constant-expression contexts

- A constant expression accepts a variable reference only when its symbol `IsConst()`: `semantics/internal/types/constant_evaluator.go:44` `if vs, ok := t.getSymbol(e.Symbol()).(model.ValueSymbol); ok && vs.IsConst() {`. Module variables are not const: `semantics/internal/symbols/symbol_resolver.go:995` `symbol := model.NewVariableSymbol(name, isPublic, false, false, globalVar.Name.GetPosition())`.
- Enum member values and xmlns URIs use this check (P6, P7): `semantics/internal/types/type_resolver.go:2177` `t.semanticError(notConstantExpressionMessage, offender.GetPosition())`.
- Match const patterns require a constant symbol (P5): `semantics/internal/types/type_resolver.go:9835` `t.semanticError("match pattern variable reference must refer to a constant", p.Expr.GetPosition())`.
- Array lengths are checked only by type, with no constant check: `semantics/internal/types/type_resolver.go:8734` `t.semanticError("fixed-length array size must be a singleton int", lenExp.GetPosition())`. A `final 3 n = 3;` variable is accepted as a length (P10).
- A variable name in a type position resolves to the variable's type with no error. Only parameters are checked: `semantics/internal/types/type_resolver.go:8788` `if !ok || !sym.IsParameter() {`. See P11.
- Tables are unimplemented in the node builder: `nodebuilder/node_builder.go:6719` `n.unimplemented("transformTableTypeDescriptor unimplemented", tableTypeDescriptorNode)`.

### Desugar and module init

- `desugarInitFn` collects each global's initializer, sorts the initializers by dependency and turns each one into an assignment in the `init` function: `desugar/desugar.go:725` `initStmts = append(initStmts, buildInitAssignment(compilerCtx, node))`. It then clears the initializers: `desugar/desugar.go:495` `pkg.GlobalVars[i].Expr = nil`. The init function is desugared like any other: `desugar/desugar.go:740` `initFn, generatedFunctions := desugarFunction(pkgCtx, "", pkg.InitFunction, span)`.
- Desugar calls `lang.__internal` functions with a generic helper. It sets the call's type from its `returnTy` argument: `desugar/query_expression.go:1953` `func createLangInternalInvocation(`.
- Record field defaults inside a function become a `BLangLambdaFunction` built from the pre-allocated symbol: `desugar/desugar.go:1273` `lambda := &ast.BLangLambdaFunction{Function: fn}`. Walking a lambda desugars its body: `desugar/expression.go:853` `expr.Function = desugarNestedFunction(cx, expr.Function)`.
- A typedesc value is a `BLangTypedescExpr` with its `Constraint` set: `semantics/internal/types/type_resolver.go:8429` `expr := &ast.BLangTypedescExpr{Constraint: inferred}`.
- Desugar already builds a module-qualified key from the symbol's package: `desugar/desugar.go:555` `runtimeGlobals[pkg.Organization+"/"+pkg.Package+":"+n.name.GetValue()] = ref`.

### Runtime

- `lang.__internal` functions are declared `external` in Ballerina: `lib/langlibs/ballerina/lang.__internal/0.0.1/any/lang.__internal.bal:17` `public isolated function querySort(any[] sortKeyRows, any[] sortDirections, any[] rowIndices, any[] payloadRows) = external;`. They are registered in Go: `lib/langlibs/go/lang.__internal/internal.go:39` `runtime.RegisterExternFunction(rt, orgName, moduleName, "querySort", func(ctx *extern.Context, args []values.BalValue) (values.BalValue, error) {`.
- A native function that returns an error panics. A panic raised while initializing a module is reported against the `init` frame, at the line of the initializer (P12).
- There is no Config.toml reader. The only environment access goes through PAL: `platform/pal/platform.go:94` `GetEnv      func(name string) string`.

## Specification

The spec version is 2022R4 (ballerina.io/spec/lang/master).

- §8.7 grammar:
  - `module-init-var-decl := metadata [public] module-init-var-quals typed-binding-pattern = module-var-init ;`
  - `module-init-var-quals := (final | isolated-qual)* | configurable`
  - `module-var-init := expression | ?`
- §8.7 prose:
  - `?` is allowed only with `configurable`.
  - A configurable variable needs an explicit type descriptor, not `var`, and its binding pattern must be a plain variable name.
  - Its type must be a subtype of `anydata`.
  - It is implicitly final.
  - Its static type is the declared type intersected with `readonly`.
  - A value from the configuration overrides the initializer, and the initializer is then not evaluated.
- §5.7.8: an intersection type that denotes an empty set of shapes is an error.
- §7.3.5: an isolated function may reference a variable declared `final` or `configurable`, and not `isolated`, whose type is a subtype of `readonly` or `isolated object {}`. This always holds for `configurable`.
- §6.4: a `constant-reference-expr` must refer to a `module-const-decl`. That applies to singleton types, array lengths (§5.4.1.1), const patterns, enum member values and xmlns URIs.
- §8.2: before any module is initialized, the program must check that every required configurable has a value of the correct type; if one is missing, initialization fails. Configuration values belong to `anydata & readonly`.

## Files read

- parser/parser.go
- nodebuilder/node_builder.go
- ast/ast.go
- ast/types.go
- semantics/internal/symbols/symbol_resolver.go
- semantics/internal/types/type_resolver.go
- semantics/internal/types/constant_evaluator.go
- semantics/internal/analysis/declarations.go
- semantics/internal/analysis/statements.go
- semantics/internal/analysis/scope/isolation.go
- semantics/internal/analysis/scope/context.go
- semantics/internal/cfg/final_reassign_analyzer.go
- semantics/internal/cfg/uninit_field_analyzer.go
- desugar/desugar.go
- desugar/expression.go
- desugar/query_expression.go
- desugar/worker.go
- birgen/bir_gen.go
- runtime/internal/exec/terminators.go
- lib/langlibs/go/lang.__internal/internal.go
- platform/pal/platform.go

## Gaps

- I have not checked whether BIR generation accepts a call whose `Ty` (`T & readonly`) differs from the callee's declared return type (`anydata`) (assumption: `createLangMapGetInvocation` does the same, but I have not seen it through BIR generation). A spike will settle this.

## Questions asked

How module variables, qualifiers and `?` flow through the parser, AST, symbol resolution, type resolution, analysis, desugar, BIR and runtime; how final, isolation and constant-expression contexts treat module variables; how desugar calls internal functions and builds closures; and what configuration sources exist at runtime.

## Probes

The `bal` binary in each command was built from this checkout with `go build -o /private/tmp/claude-501/probes/bal ./cli/cmd`. The probe sources are under `/private/tmp/claude-501/probes/`.

### P1: configurable is rejected by the node builder
Source: `configurable int x = ?;`
Command: `/private/tmp/claude-501/probes/bal run /private/tmp/claude-501/probes/p1.bal`
Output:
```text
fatal[UNIMPLEMENTED_ERROR]: configurable module variables are not supported yet
 --> p1.bal:1:1
```

### P2: `?` outside a configurable is an invalid token
Command: `/private/tmp/claude-501/probes/bal run /private/tmp/claude-501/probes/p2.bal`
Output:
```text
error[SYNTAX_ERROR]: invalid token
 --> p2.bal:1:10
```

### P3: a final module variable with an initializer cannot be assigned in init
Command: `/private/tmp/claude-501/probes/bal run /private/tmp/claude-501/probes/p3.bal`
Output:
```text
error[SEMANTIC_ERROR]: cannot assign a value to final variable 'x'
 --> p3.bal:3:5
```

### P4: an isolated function may read a final int but not a plain int
Command: `/private/tmp/claude-501/probes/bal run /private/tmp/claude-501/probes/p4.bal`
Output:
```text
error[SEMANTIC_ERROR]: access of mutable variable
 --> p4.bal:5:16
```

### P5: a final variable in a match const pattern
Command: `/private/tmp/claude-501/probes/bal run /private/tmp/claude-501/probes/p5.bal`
Output:
```text
error[SEMANTIC_ERROR]: match pattern variable reference must refer to a constant
 --> p5.bal:5:9
```

### P6: a final variable as an enum member value
Command: `/private/tmp/claude-501/probes/bal run /private/tmp/claude-501/probes/p6.bal`
Output:
```text
error[SEMANTIC_ERROR]: expression is not a constant expression
 --> p6.bal:3:9
```

### P7: a final variable as an xmlns URI
Command: `/private/tmp/claude-501/probes/bal run /private/tmp/claude-501/probes/p7.bal`
Output:
```text
error[SEMANTIC_ERROR]: expression is not a constant expression
 --> p7.bal:2:7
```

### P8: an unused non-public module variable
Command: `/private/tmp/claude-501/probes/bal run /private/tmp/claude-501/probes/p8.bal`
Output:
```text
error[SEMANTIC_ERROR]: unused variable 'x'
 --> p8.bal:1:1
```

### P9: an empty intersection type descriptor
Command: `/private/tmp/claude-501/probes/bal run /private/tmp/claude-501/probes/p9.bal`
Output:
```text
error[SEMANTIC_ERROR]: intersection type is empty (equivalent to never)
 --> p9.bal:3:8
```

### P10: a singleton-typed final variable is accepted as an array length
Command: `/private/tmp/claude-501/probes/bal run /private/tmp/claude-501/probes/al.bal`
Output:
```text
[1,2,3]
```

### P11: a variable name used as a type gives that variable's type
Command: `/private/tmp/claude-501/probes/bal run /private/tmp/claude-501/probes/ty.bal`
Output:
```text
5
```

### P12: a native panic during module initialization
Command: `/private/tmp/claude-501/probes/bal run /private/tmp/claude-501/probes/ip.bal`
Output:
```text
error: string index out of range: startIndex=5 endIndex=9 length=3
        at init(ip.bal:4)
```
