// Copyright (c) 2026, WSO2 LLC. (http://www.wso2.com).
//
// WSO2 LLC. licenses this file to you under the Apache License,
// Version 2.0 (the "License"); you may not use this file except
// in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package ast

import (
	"fmt"
)

// A Visitor's Visit method is invoked for each node encountered by [Walk].
// If the result visitor w is not nil, [Walk] visits each of the children
// of node with the visitor w, followed by a call of w.Visit(nil).
type Visitor interface {
	Visit(node BLangNode) (w Visitor)
	VisitTypeData(typeData *TypeData) (w Visitor)
}

// Walk traverses an AST in depth-first order: It starts by calling
// v.Visit(node); node must not be nil. If the visitor w returned by
// v.Visit(node) is not nil, Walk is invoked recursively with visitor
// w for each of the non-nil children of node, followed by a call of
// w.Visit(nil).
func Walk(v Visitor, node BLangNode) {
	if v = v.Visit(node); v == nil {
		return
	}
	switch node := node.(type) {
	// Section 1: Top-Level/Package Declarations
	case *BLangPackage:
		walkPackage(v, node)

	case *BLangCompilationUnit:
		walkCompilationUnit(v, node)

	case *BLangImportPackage:
		walkImportPackage(v, node)

	case *BLangService:
		walkService(v, node)

	case *BLangClassDefinition:
		walkClassDefinition(v, node)

	case *BLangAnnotation:
		walkAnnotation(v, node)

	case *BLangAnnotationAttachment:
		walkAnnotationAttachment(v, node)

	// Section 2: Type Definitions & Variables
	case *BLangTypeDefinition:
		walkTypeDefinition(v, node)

	case *BLangVariable:
		walkVariable(v, node)

	case *BLangXMLNS:
		walkXMLNS(v, node)

	// Section 3: Function & Body
	case *BLangFunction:
		walkFunction(v, node)

	case *BLangResourceMethod:
		walkResourceMethod(v, node)

	case *BLangReturnTypeDescriptor:
		walkReturnTypeDescriptor(v, node)

	case *BLangBlockFunctionBody:
		walkBlockFunctionBody(v, node)

	case *BLangExprFunctionBody:
		walkExprFunctionBody(v, node)

	case *BLangExternFunctionBody:
		// Leaf node - no children to walk

	// Section 4: Statements
	case *BLangBlockStmt:
		walkBlockStmt(v, node)

	case *BLangAssignment:
		walkAssignment(v, node)

	case *BLangCompoundAssignment:
		walkCompoundAssignment(v, node)

	case *BLangExpressionStmt:
		walkExpressionStmt(v, node)

	case *BLangIf:
		walkIf(v, node)

	case *BLangWhile:
		walkWhile(v, node)

	case *BLangForeach:
		walkForeach(v, node)

	case *BLangDo:
		walkDo(v, node)

	case *BLangLock:
		Walk(v, &node.Body)

	case *BLangMatchStatement:
		walkMatchStatement(v, node)

	case *BLangMatchClause:
		walkMatchClause(v, node)

	case *BLangVariableDef:
		walkVariableDef(v, node)

	case *BLangReturn:
		walkReturn(v, node)

	case *BLangPanic:
		walkPanic(v, node)

	case *BLangBreak:
		// Leaf node

	case *BLangContinue:
		// Leaf node

	// Section 5: Expressions - Basic
	case *BLangBinaryExpr:
		walkBinaryExpr(v, node)

	case *BLangTernaryExpr:
		walkTernaryExpr(v, node)

	case *BLangQueryExpr:
		walkQueryExpr(v, node)

	case *BLangUnaryExpr:
		walkUnaryExpr(v, node)

	case *BLangNilConditionalExpr:
		walkNilConditionalExpr(v, node)

	case *BLangCheckedExpr:
		walkCheckedExpr(v, node)

	case *BLangCheckPanickedExpr:
		walkCheckPanickedExpr(v, node)

	case *BLangTrapExpr:
		walkTrapExpr(v, node)

	case *BLangGroupExpr:
		walkGroupExpr(v, node)

	case *BLangIndexBasedAccess:
		walkIndexBasedAccess(v, node)

	case *BLangFieldBaseAccess:
		walkFieldBaseAccess(v, node)

	case *BLangListConstructorExpr:
		walkListConstructorExpr(v, node)
	case *BLangMappingConstructorExpr:
		walkMappingConstructorExpr(v, node)
	case *BLangErrorConstructorExpr:
		walkErrorConstructorExpr(v, node)

	case *BLangInvocation:
		walkInvocation(v, node)
	case *BLangLambdaFunction:
		walkLambdaFunction(v, node)

	case *BLangArrowFunction:
		walkArrowFunction(v, node)

	case *BLangAnnotAccessExpr:
		walkAnnotAccessExpr(v, node)

	case *BLangTypedescExpr:
		walkTypeDescriptor(v, node.typeDescriptor)

	case *BLangInferredTypedescDefault:
		// leaf — no children

	case *BLangDefaultArg:
		// leaf — no children

	case *BLangTypeConversionExpr:
		walkTypeConversionExpr(v, node)

	case *BLangTypeTestExpr:
		walkTypeTestExpr(v, node)

	case *BLangCommitExpr:
		panic("unimplemented")

	// Section 6: Expressions - Variable Refs
	case *BLangVarRef:
		walkVarRef(v, node)

	case *BLangConstRef:
		walkConstRef(v, node)

	case *BLangLiteral:
		// Leaf node

	case *BLangNumericLiteral:
		// Leaf node

	case *BLangXMLSequenceLiteral:
		walkXMLSequenceLiteral(v, node)

	case *BLangTemplateExpr:
		walkTemplateExpr(v, node)

	case *BLangXMLTemplateExpr:
		walkXMLTemplateExpr(v, node)

	case *BLangXMLElementLiteral:
		walkXMLElementLiteral(v, node)

	case *BLangXMLAttribute:
		walkXMLAttribute(v, node)

	case *BLangXMLPILiteral:
		// Leaf node

	case *BLangXMLCommentLiteral:
		// Leaf node

	case *BLangXMLTextLiteral:
		// Leaf node

	case *BLangXMLFilterExpression:
		walkXMLFilterExpression(v, node)

	case *BLangXMLStepExpression:
		walkXMLStepExpression(v, node)

	case *BLangXMLStepStart, *BLangXMLStepFilterExtend:
		// Leaf nodes

	case *BLangXMLStepIndexExtend:
		walkXMLStepIndexExtend(v, node)

	case *BLangXMLStepMethodCallExtend:
		walkXMLStepMethodCallExtend(v, node)

	// Section 7: Expressions - Worker
	case *BLangWorkerReceive:
		walkWorkerReceive(v, node)

	case *BLangAlternateWorkerReceive:
		walkAlternateWorkerReceive(v, node)

	// Section 8: Type Nodes
	case *BLangArrayType:
		walkArrayType(v, node)

	case *BLangUserDefinedType:
		walkUserDefinedType(v, node)

	case *BLangValueType:
		// No children to walk

	case *BLangBuiltInRefTypeNode:
		// No children to walk

	case *BLangFiniteTypeNode:
		walkFiniteTypeNode(v, node)

	case *BLangUnionTypeNode:
		walkUnionTypeNode(v, node)

	case *BLangIntersectionTypeNode:
		walkIntersectionTypeNode(v, node)

	case *BLangErrorTypeNode:
		WalkTypeData(v, &node.DetailType)

	case *BLangConstrainedType:
		walkConstrainedType(v, node)
	case *BLangStreamType:
		walkStreamType(v, node)
	case *BLangTupleTypeNode:
		walkTupleTypeNode(v, node)

	case *BLangRecordType:
		walkRecordType(v, node)

	case *BLangObjectType:
		walkObjectType(v, node)

	case *BObjectField:
		walkBObjectField(v, node)

	case *BMethodDecl:
		walkBMethodDecl(v, node)

	case *BLangFunctionType:
		walkFunctionType(v, node)

	case *BLangFunctionTypeParam:
		walkFunctionTypeParam(v, node)

	// Section 9: Binding Patterns
	case *BLangCaptureBindingPattern:
		Walk(v, &node.Identifier)

	case *BLangWildCardBindingPattern:
		// Leaf node

	case *BLangSimpleBindingPattern:
		walkSimpleBindingPattern(v, node)

	case *BLangErrorBindingPattern:
		walkErrorBindingPattern(v, node)

	case *BLangErrorMessageBindingPattern:
		walkErrorMessageBindingPattern(v, node)

	case *BLangErrorCauseBindingPattern:
		walkErrorCauseBindingPattern(v, node)

	case *BLangErrorFieldBindingPatterns:
		walkErrorFieldBindingPatterns(v, node)

	case *BLangNamedArgBindingPattern:
		walkNamedArgBindingPattern(v, node)

	case *BLangRestBindingPattern:
		walkRestBindingPattern(v, node)

	// Section 10: Clauses
	case *BLangFromClause:
		walkFromClause(v, node)

	case *BLangJoinClause:
		walkJoinClause(v, node)

	case *BLangSelectClause:
		walkSelectClause(v, node)

	case *BLangLetClause:
		walkLetClause(v, node)

	case *BLangWhereClause:
		walkWhereClause(v, node)

	case *BLangGroupByClause:
		walkGroupByClause(v, node)

	case *BLangGroupingKey:
		walkGroupingKey(v, node)

	case *BLangOnClause:
		walkOnClause(v, node)

	case *BLangLimitClause:
		walkLimitClause(v, node)

	case *BLangOrderByClause:
		walkOrderByClause(v, node)

	case *BLangOrderKey:
		walkOrderKey(v, node)

	case *BLangOnConflictClause:
		walkOnConflictClause(v, node)

	case *BLangOnFailClause:
		walkOnFailClause(v, node)

	case *BLangDoClause:
		walkDoClause(v, node)

	case *BLangCollectClause:
		walkCollectClause(v, node)

	// Section 11: Documentation
	case *BLangMarkdownDocumentation:
		walkMarkdownDocumentation(v, node)

	case *BLangMarkdownDocumentationLine:
		// Leaf node

	case *BLangMarkdownParameterDocumentation:
		walkMarkdownParameterDocumentation(v, node)

	case *BLangMarkdownReturnParameterDocumentation:
		// Leaf node

	case *BLangMarkDownDeprecationDocumentation:
		// Leaf node

	case *BLangMarkDownDeprecatedParametersDocumentation:
		walkMarkDownDeprecatedParametersDocumentation(v, node)

	// Section 12: Pattern Matching
	case *BLangConstPattern:
		walkConstPattern(v, node)

	case *BLangWildCardMatchPattern:
		// Leaf node

	// Section 13: Misc Leaf
	case *BLangIdentifier:
		// Leaf node

	case *BLangBadExprOrAction, *BLangBadStmt, *BLangBadTypeNode, *BLangBadTopLevelNode:
		// Leaf node

	case *BLangMarkdownReferenceDocumentation:
		// Leaf node

	case *BLangNamedArgsExpression:
		walkNamedArgsExpression(v, node)

	case *BLangNewExpression:
		walkNewExpression(v, node)

	case *BLangStartAction:
		walkStartAction(v, node)

	case *BLangSingleWaitAction:
		walkSingleWaitAction(v, node)

	case *BLangAlternateWaitAction:
		walkAlternateWaitAction(v, node)

	case *BLangMultipleWaitAction:
		walkMultipleWaitAction(v, node)

	case *BLangRemoteMethodCallAction:
		walkRemoteMethodCallAction(v, node)

	case *BLangClientResourceAccessAction:
		walkClientResourceAccessAction(v, node)

	default:
		panic(fmt.Sprintf("unexpected node type %T", node))
	}
	v.Visit(nil)
}

// The per-node walkXxx helpers keep Walk's stack frame small, since that frame is paid on every level of AST
// recursion. Helpers with loops that the compiler would otherwise inline into Walk are marked go:noinline.

func walkPackage(v Visitor, node *BLangPackage) {
	for i := range node.Imports {
		Walk(v, node.Imports[i])
	}
	for i := range node.XmlnsList {
		Walk(v, node.XmlnsList[i])
	}
	for i := range node.Constants {
		Walk(v, node.Constants[i])
	}
	for i := range node.GlobalVars {
		Walk(v, node.GlobalVars[i])
	}
	for i := range node.Services {
		Walk(v, node.Services[i])
	}
	for i := range node.Functions {
		Walk(v, node.Functions[i])
	}
	for i := range node.TypeDefinitions {
		Walk(v, node.TypeDefinitions[i])
	}
	for i := range node.Annotations {
		Walk(v, node.Annotations[i])
	}
	if node.InitFunction != nil {
		Walk(v, node.InitFunction)
	}
	for i := range node.ClassDefinitions {
		Walk(v, node.ClassDefinitions[i])
	}
	for _, badNode := range node.BadTopLevelNodes {
		Walk(v, badNode)
	}
}

//go:noinline
func walkCompilationUnit(v Visitor, node *BLangCompilationUnit) {
	for _, topLevelNode := range node.TopLevelNodes {
		Walk(v, topLevelNode.(BLangNode))
	}
}

func walkImportPackage(v Visitor, node *BLangImportPackage) {
	if node.OrgName != nil {
		Walk(v, node.OrgName)
	}
	for i := range node.PkgNameComps {
		Walk(v, &node.PkgNameComps[i])
	}
	if node.Alias != nil {
		Walk(v, node.Alias)
	}
	if node.Version != nil {
		Walk(v, node.Version)
	}
}

func walkService(v Visitor, node *BLangService) {
	for _, expr := range node.AttachedExprs {
		Walk(v, expr.(BLangNode))
	}
	if node.AttachPointLiteral != nil {
		Walk(v, node.AttachPointLiteral)
	}
	for i := range node.AbsoluteResourcePath {
		Walk(v, &node.AbsoluteResourcePath[i])
	}
	walkClassDefnBody(v, &node.classDefnBase)
}

func walkClassDefinition(v Visitor, node *BLangClassDefinition) {
	if node.Name != nil {
		Walk(v, node.Name)
	}
	walkClassDefnBody(v, &node.classDefnBase)
}

func walkAnnotation(v Visitor, node *BLangAnnotation) {
	if node.Name != nil {
		Walk(v, node.Name)
	}
	for i := range node.AnnAttachments {
		Walk(v, &node.AnnAttachments[i])
	}
	walkTypeDescriptor(v, node.typeDescriptor)
}

func walkAnnotationAttachment(v Visitor, node *BLangAnnotationAttachment) {
	Walk(v, node.Expr.(BLangNode))
	if node.AnnotationName != nil {
		Walk(v, node.AnnotationName)
	}
	if node.PkgAlias != nil {
		Walk(v, node.PkgAlias)
	}
}

func walkTypeDefinition(v Visitor, node *BLangTypeDefinition) {
	if node.Name != nil {
		Walk(v, node.Name)
	}
	WalkTypeData(v, &node.typeData)
	for i := range node.annAttachments {
		Walk(v, &node.annAttachments[i])
	}
}

func walkVariable(v Visitor, node *BLangVariable) {
	if node.Name != nil {
		Walk(v, node.Name)
	}
	for i := range node.AnnAttachments {
		Walk(v, &node.AnnAttachments[i])
	}
	if tn := node.TypeNode(); tn != nil {
		Walk(v, tn.(BLangNode))
	}
	if node.Expr != nil {
		Walk(v, node.Expr.(BLangNode))
	}
}

func walkXMLNS(v Visitor, node *BLangXMLNS) {
	Walk(v, node.namespaceURI.(BLangNode))
	if node.prefix != nil {
		Walk(v, node.prefix)
	}
}

func walkFunction(v Visitor, node *BLangFunction) {
	if node.Name != nil {
		Walk(v, node.Name)
	}
	attachments := node.GetAnnotationAttachments()
	for i := range attachments {
		Walk(v, &attachments[i])
	}
	params := node.GetParameters()
	for i := range params {
		Walk(v, &params[i])
	}
	if node.RestParam != nil {
		Walk(v, node.RestParam)
	}
	if node.returnTypeDescriptor != nil {
		Walk(v, node.returnTypeDescriptor)
	}
	if node.Body != nil {
		Walk(v, node.Body.(BLangNode))
	}
}

func walkResourceMethod(v Visitor, node *BLangResourceMethod) {
	if node.Name != nil {
		Walk(v, node.Name)
	}
	attachments := node.GetAnnotationAttachments()
	for i := range attachments {
		Walk(v, &attachments[i])
	}
	for i := range node.ResourcePath {
		if tn := node.ResourcePath[i].ParamType; tn != nil {
			walkTypeDescriptor(v, tn)
		}
	}
	params := node.GetParameters()
	for i := range params {
		Walk(v, &params[i])
	}
	if node.RestParam != nil {
		Walk(v, node.RestParam)
	}
	if node.returnTypeDescriptor != nil {
		Walk(v, node.returnTypeDescriptor)
	}
	if node.Body != nil {
		Walk(v, node.Body.(BLangNode))
	}
}

func walkReturnTypeDescriptor(v Visitor, node *BLangReturnTypeDescriptor) {
	attachments := node.GetAnnotationAttachments()
	for i := range attachments {
		Walk(v, &attachments[i])
	}
	walkTypeDescriptor(v, node.TypeDescriptor)
}

//go:noinline
func walkBlockFunctionBody(v Visitor, node *BLangBlockFunctionBody) {
	for _, stmt := range node.Stmts {
		Walk(v, stmt.(BLangNode))
	}
}

func walkExprFunctionBody(v Visitor, node *BLangExprFunctionBody) {
	if node.Expr != nil {
		Walk(v, node.Expr.(BLangNode))
	}
}

//go:noinline
func walkBlockStmt(v Visitor, node *BLangBlockStmt) {
	for _, stmt := range node.Stmts {
		Walk(v, stmt.(BLangNode))
	}
}

func walkAssignment(v Visitor, node *BLangAssignment) {
	if node.VarRef != nil {
		Walk(v, node.VarRef.(BLangNode))
	}
	if node.Expr != nil {
		Walk(v, node.Expr.(BLangNode))
	}
}

func walkCompoundAssignment(v Visitor, node *BLangCompoundAssignment) {
	if node.VarRef != nil {
		Walk(v, node.VarRef.(BLangNode))
	}
	if node.Expr != nil {
		Walk(v, node.Expr.(BLangNode))
	}
}

func walkExpressionStmt(v Visitor, node *BLangExpressionStmt) {
	if node.Expr != nil {
		Walk(v, node.Expr.(BLangNode))
	}
}

func walkIf(v Visitor, node *BLangIf) {
	if node.Expr != nil {
		Walk(v, node.Expr.(BLangNode))
	}
	Walk(v, &node.Body)
	if node.ElseStmt != nil {
		Walk(v, node.ElseStmt.(BLangNode))
	}
}

func walkWhile(v Visitor, node *BLangWhile) {
	if node.Expr != nil {
		Walk(v, node.Expr.(BLangNode))
	}
	Walk(v, &node.Body)
	Walk(v, &node.OnFailClause)
}

func walkForeach(v Visitor, node *BLangForeach) {
	if node.Collection != nil {
		Walk(v, node.Collection.(BLangNode))
	}
	if node.VariableDef != nil {
		Walk(v, node.VariableDef)
	}
	Walk(v, &node.Body)
	if node.OnFailClause != nil {
		Walk(v, node.OnFailClause)
	}
}

func walkDo(v Visitor, node *BLangDo) {
	Walk(v, &node.Body)
	Walk(v, &node.OnFailClause)
}

func walkMatchStatement(v Visitor, node *BLangMatchStatement) {
	if node.Expr != nil {
		Walk(v, node.Expr.(BLangNode))
	}
	for i := range node.MatchClauses {
		Walk(v, &node.MatchClauses[i])
	}
}

func walkMatchClause(v Visitor, node *BLangMatchClause) {
	for _, pattern := range node.Patterns {
		Walk(v, pattern.(BLangNode))
	}
	if node.Guard != nil {
		Walk(v, node.Guard)
	}
	Walk(v, &node.Body)
}

func walkVariableDef(v Visitor, node *BLangVariableDef) {
	if node.Var != nil {
		Walk(v, node.Var)
	}
}

func walkReturn(v Visitor, node *BLangReturn) {
	if node.Expr != nil {
		Walk(v, node.Expr.(BLangNode))
	}
}

func walkPanic(v Visitor, node *BLangPanic) {
	if node.Expr != nil {
		Walk(v, node.Expr.(BLangNode))
	}
}

func walkBinaryExpr(v Visitor, node *BLangBinaryExpr) {
	if node.LhsExpr != nil {
		Walk(v, node.LhsExpr.(BLangNode))
	}
	if node.RhsExpr != nil {
		Walk(v, node.RhsExpr.(BLangNode))
	}
}

func walkTernaryExpr(v Visitor, node *BLangTernaryExpr) {
	if node.Condition != nil {
		Walk(v, node.Condition.(BLangNode))
	}
	if node.ThenExpr != nil {
		Walk(v, node.ThenExpr.(BLangNode))
	}
	if node.ElseExpr != nil {
		Walk(v, node.ElseExpr.(BLangNode))
	}
}

//go:noinline
func walkQueryExpr(v Visitor, node *BLangQueryExpr) {
	for i := range node.QueryClauseList {
		Walk(v, node.QueryClauseList[i])
	}
}

func walkUnaryExpr(v Visitor, node *BLangUnaryExpr) {
	if node.Expr != nil {
		Walk(v, node.Expr.(BLangNode))
	}
}

func walkNilConditionalExpr(v Visitor, node *BLangNilConditionalExpr) {
	if node.LhsExpr != nil {
		Walk(v, node.LhsExpr.(BLangNode))
	}
	if node.RhsExpr != nil {
		Walk(v, node.RhsExpr.(BLangNode))
	}
}

func walkCheckedExpr(v Visitor, node *BLangCheckedExpr) {
	if node.Expr != nil {
		Walk(v, node.Expr.(BLangNode))
	}
}

func walkCheckPanickedExpr(v Visitor, node *BLangCheckPanickedExpr) {
	if node.Expr != nil {
		Walk(v, node.Expr.(BLangNode))
	}
}

func walkTrapExpr(v Visitor, node *BLangTrapExpr) {
	if node.Expr != nil {
		Walk(v, node.Expr.(BLangNode))
	}
}

func walkGroupExpr(v Visitor, node *BLangGroupExpr) {
	if node.Expression != nil {
		Walk(v, node.Expression.(BLangNode))
	}
}

func walkIndexBasedAccess(v Visitor, node *BLangIndexBasedAccess) {
	if node.Expr != nil {
		Walk(v, node.Expr.(BLangNode))
	}
	if node.IndexExpr != nil {
		Walk(v, node.IndexExpr.(BLangNode))
	}
}

func walkFieldBaseAccess(v Visitor, node *BLangFieldBaseAccess) {
	if node.Expr != nil {
		Walk(v, node.Expr.(BLangNode))
	}
}

//go:noinline
func walkListConstructorExpr(v Visitor, node *BLangListConstructorExpr) {
	for _, expr := range node.Exprs {
		Walk(v, expr.(BLangNode))
	}
}

func walkMappingConstructorExpr(v Visitor, node *BLangMappingConstructorExpr) {
	for _, f := range node.Fields {
		if kv, ok := f.(*BLangMappingKeyValueField); ok {
			if kv.Key != nil && kv.Key.Expr != nil {
				Walk(v, kv.Key.Expr.(BLangNode))
			}
			if kv.ValueExpr != nil {
				Walk(v, kv.ValueExpr.(BLangNode))
			}
		}
	}
}

func walkErrorConstructorExpr(v Visitor, node *BLangErrorConstructorExpr) {
	if node.ErrorTypeRef != nil {
		Walk(v, node.ErrorTypeRef)
	}
	for _, arg := range node.PositionalArgs {
		Walk(v, arg.(BLangNode))
	}
	for i := range node.NamedArgs {
		Walk(v, &node.NamedArgs[i])
	}
}

func walkInvocation(v Visitor, node *BLangInvocation) {
	if node.PkgAlias != nil {
		Walk(v, node.PkgAlias)
	}
	if node.Name != nil {
		Walk(v, node.Name)
	}
	if node.Expr != nil {
		Walk(v, node.Expr.(BLangNode))
	}
	for _, arg := range node.ArgExprs {
		Walk(v, arg.(BLangNode))
	}
}

func walkLambdaFunction(v Visitor, node *BLangLambdaFunction) {
	if node.Function != nil {
		Walk(v, node.Function)
	}
}

func walkArrowFunction(v Visitor, node *BLangArrowFunction) {
	for i := range node.Params {
		Walk(v, &node.Params[i])
	}
	if node.Body != nil {
		Walk(v, node.Body)
	}
}

func walkAnnotAccessExpr(v Visitor, node *BLangAnnotAccessExpr) {
	if node.Expr != nil {
		Walk(v, node.Expr.(BLangNode))
	}
	if node.PkgAlias != nil {
		Walk(v, node.PkgAlias)
	}
	if node.AnnotationName != nil {
		Walk(v, node.AnnotationName)
	}
}

func walkTypeConversionExpr(v Visitor, node *BLangTypeConversionExpr) {
	if node.Expression != nil {
		Walk(v, node.Expression.(BLangNode))
	}
	if node.TypeDescriptor != nil {
		Walk(v, node.TypeDescriptor.(BLangNode))
	}
}

func walkTypeTestExpr(v Visitor, node *BLangTypeTestExpr) {
	if node.Expr != nil {
		Walk(v, node.Expr.(BLangNode))
	}
	WalkTypeData(v, &node.Type)
}

func walkVarRef(v Visitor, node *BLangVarRef) {
	if node.PkgAlias != nil {
		Walk(v, node.PkgAlias)
	}
	if node.VariableName != nil {
		Walk(v, node.VariableName)
	}
}

func walkConstRef(v Visitor, node *BLangConstRef) {
	if node.PkgAlias != nil {
		Walk(v, node.PkgAlias)
	}
	if node.VariableName != nil {
		Walk(v, node.VariableName)
	}
}

//go:noinline
func walkXMLSequenceLiteral(v Visitor, node *BLangXMLSequenceLiteral) {
	for _, child := range node.Children {
		Walk(v, child)
	}
}

//go:noinline
func walkTemplateExpr(v Visitor, node *BLangTemplateExpr) {
	for _, ins := range node.Insertions {
		Walk(v, ins)
	}
}

//go:noinline
func walkXMLTemplateExpr(v Visitor, node *BLangXMLTemplateExpr) {
	for _, ins := range node.Insertions {
		Walk(v, ins)
	}
}

func walkXMLElementLiteral(v Visitor, node *BLangXMLElementLiteral) {
	for i := range node.Attrs {
		Walk(v, &node.Attrs[i])
	}
	if node.Content != nil {
		Walk(v, node.Content)
	}
}

func walkXMLAttribute(v Visitor, node *BLangXMLAttribute) {
	if node.Value != nil {
		Walk(v, node.Value)
	}
}

func walkXMLFilterExpression(v Visitor, node *BLangXMLFilterExpression) {
	if node.Expression != nil {
		Walk(v, node.Expression)
	}
}

func walkXMLStepExpression(v Visitor, node *BLangXMLStepExpression) {
	if node.Expression != nil {
		Walk(v, node.Expression)
	}
	if node.Start != nil {
		Walk(v, node.Start)
	}
	for _, extension := range node.Extensions {
		Walk(v, extension)
	}
}

func walkXMLStepIndexExtend(v Visitor, node *BLangXMLStepIndexExtend) {
	if node.Expression != nil {
		Walk(v, node.Expression)
	}
}

//go:noinline
func walkXMLStepMethodCallExtend(v Visitor, node *BLangXMLStepMethodCallExtend) {
	if node.Invocation != nil {
		for _, arg := range node.Invocation.ArgExprs {
			Walk(v, arg)
		}
	}
}

func walkWorkerReceive(v Visitor, node *BLangWorkerReceive) {
	if node.WorkerIdentifier != nil {
		Walk(v, node.WorkerIdentifier)
	}
}

//go:noinline
func walkAlternateWorkerReceive(v Visitor, node *BLangAlternateWorkerReceive) {
	for i := range node.workerReceives {
		Walk(v, &node.workerReceives[i])
	}
}

func walkArrayType(v Visitor, node *BLangArrayType) {
	WalkTypeData(v, &node.Elemtype)
	for i := range node.Sizes {
		if node.Sizes[i] != nil {
			Walk(v, node.Sizes[i].(BLangNode))
		}
	}
}

func walkUserDefinedType(v Visitor, node *BLangUserDefinedType) {
	Walk(v, &node.PkgAlias)
	Walk(v, &node.TypeName)
}

//go:noinline
func walkFiniteTypeNode(v Visitor, node *BLangFiniteTypeNode) {
	for i := range node.ValueSpace {
		Walk(v, node.ValueSpace[i].(BLangNode))
	}
}

func walkUnionTypeNode(v Visitor, node *BLangUnionTypeNode) {
	WalkTypeData(v, &node.lhs)
	WalkTypeData(v, &node.rhs)
}

func walkIntersectionTypeNode(v Visitor, node *BLangIntersectionTypeNode) {
	WalkTypeData(v, &node.lhs)
	WalkTypeData(v, &node.rhs)
}

func walkConstrainedType(v Visitor, node *BLangConstrainedType) {
	WalkTypeData(v, &node.Type)
	WalkTypeData(v, &node.Constraint)
}

func walkStreamType(v Visitor, node *BLangStreamType) {
	WalkTypeData(v, &node.ValueType)
	WalkTypeData(v, &node.CompletionType)
}

func walkTupleTypeNode(v Visitor, node *BLangTupleTypeNode) {
	for i := range node.Members {
		for j := range node.Members[i].AnnAttachments {
			Walk(v, &node.Members[i].AnnAttachments[j])
		}
		Walk(v, node.Members[i].TypeDesc.(BLangNode))
	}
	if node.Rest != nil {
		Walk(v, node.Rest.(BLangNode))
	}
}

func walkRecordType(v Visitor, node *BLangRecordType) {
	for _, inclusion := range node.TypeInclusions {
		Walk(v, inclusion.(BLangNode))
	}
	for i := range node.fields {
		field := &node.fields[i]
		attachments := field.GetAnnotationAttachments()
		for j := range attachments {
			Walk(v, &attachments[j])
		}
		Walk(v, field.Type.(BLangNode))
		if field.Default != nil {
			Walk(v, field.Default.Expr.(BLangNode))
		}
	}
	if node.RestType != nil {
		Walk(v, node.RestType.(BLangNode))
	}
}

//go:noinline
func walkObjectType(v Visitor, node *BLangObjectType) {
	for member := range node.Members() {
		Walk(v, member.(BLangNode))
	}
}

func walkBObjectField(v Visitor, node *BObjectField) {
	attachments := node.GetAnnotationAttachments()
	for i := range attachments {
		Walk(v, &attachments[i])
	}
	Walk(v, node.Ty.(BLangNode))
}

func walkBMethodDecl(v Visitor, node *BMethodDecl) {
	for _, param := range node.RequiredParams {
		Walk(v, param.TypeDesc.(BLangNode))
	}
	if node.ReturnTypeDescriptor != nil {
		Walk(v, node.ReturnTypeDescriptor.(BLangNode))
	}
}

func walkFunctionType(v Visitor, node *BLangFunctionType) {
	for i := range node.RequiredParams {
		Walk(v, node.RequiredParams[i])
	}
	if node.RestParam != nil {
		Walk(v, node.RestParam)
	}
	if node.ReturnTypeDescriptor != nil {
		Walk(v, node.ReturnTypeDescriptor.(BLangNode))
	}
}

func walkFunctionTypeParam(v Visitor, node *BLangFunctionTypeParam) {
	if node.Name != nil {
		Walk(v, node.Name)
	}
	if node.TypeDesc != nil {
		Walk(v, node.TypeDesc.(BLangNode))
	}
	if node.InitExpr != nil {
		Walk(v, node.InitExpr.(BLangNode))
	}
}

func walkSimpleBindingPattern(v Visitor, node *BLangSimpleBindingPattern) {
	if node.CaptureBindingPattern != nil {
		Walk(v, node.CaptureBindingPattern)
	}
	if node.WildCardBindingPattern != nil {
		Walk(v, node.WildCardBindingPattern)
	}
}

func walkErrorBindingPattern(v Visitor, node *BLangErrorBindingPattern) {
	if node.ErrorTypeReference != nil {
		Walk(v, node.ErrorTypeReference)
	}
	if node.ErrorMessageBindingPattern != nil {
		Walk(v, node.ErrorMessageBindingPattern)
	}
	if node.ErrorCauseBindingPattern != nil {
		Walk(v, node.ErrorCauseBindingPattern)
	}
	if node.ErrorFieldBindingPatterns != nil {
		Walk(v, node.ErrorFieldBindingPatterns)
	}
}

func walkErrorMessageBindingPattern(v Visitor, node *BLangErrorMessageBindingPattern) {
	if node.SimpleBindingPattern != nil {
		Walk(v, node.SimpleBindingPattern)
	}
}

func walkErrorCauseBindingPattern(v Visitor, node *BLangErrorCauseBindingPattern) {
	if node.SimpleBindingPattern != nil {
		Walk(v, node.SimpleBindingPattern)
	}
	if node.ErrorBindingPattern != nil {
		Walk(v, node.ErrorBindingPattern)
	}
}

func walkErrorFieldBindingPatterns(v Visitor, node *BLangErrorFieldBindingPatterns) {
	for i := range node.NamedArgBindingPatterns {
		Walk(v, &node.NamedArgBindingPatterns[i])
	}
	if node.RestBindingPattern != nil {
		Walk(v, node.RestBindingPattern)
	}
}

func walkNamedArgBindingPattern(v Visitor, node *BLangNamedArgBindingPattern) {
	if node.ArgName != nil {
		Walk(v, node.ArgName)
	}
	if node.BindingPattern != nil {
		Walk(v, node.BindingPattern.(BLangNode))
	}
}

func walkRestBindingPattern(v Visitor, node *BLangRestBindingPattern) {
	if node.VariableName != nil {
		Walk(v, node.VariableName)
	}
}

func walkFromClause(v Visitor, node *BLangFromClause) {
	if node.Collection != nil {
		Walk(v, node.Collection.(BLangNode))
	}
	if node.VariableDefinitionNode != nil {
		Walk(v, node.VariableDefinitionNode)
	}
}

func walkJoinClause(v Visitor, node *BLangJoinClause) {
	if node.Collection != nil {
		Walk(v, node.Collection.(BLangNode))
	}
	if node.OnClause.OnExpr != nil {
		Walk(v, node.OnClause.OnExpr.(BLangNode))
	}
	if node.VariableDefinitionNode != nil {
		Walk(v, node.VariableDefinitionNode)
	}
	if node.OnClause.EqualsExpr != nil {
		Walk(v, node.OnClause.EqualsExpr.(BLangNode))
	}
}

func walkSelectClause(v Visitor, node *BLangSelectClause) {
	if node.Expression != nil {
		Walk(v, node.Expression.(BLangNode))
	}
}

//go:noinline
func walkLetClause(v Visitor, node *BLangLetClause) {
	for i := range node.LetVarDeclarations {
		Walk(v, &node.LetVarDeclarations[i])
	}
}

func walkWhereClause(v Visitor, node *BLangWhereClause) {
	if node.Expression != nil {
		Walk(v, node.Expression.(BLangNode))
	}
}

//go:noinline
func walkGroupByClause(v Visitor, node *BLangGroupByClause) {
	for _, groupingKey := range node.GetGroupingKeyList() {
		Walk(v, groupingKey)
	}
}

func walkGroupingKey(v Visitor, node *BLangGroupingKey) {
	if groupingKey := node.GetGroupingKey(); groupingKey != nil {
		Walk(v, groupingKey.(BLangNode))
	}
}

func walkOnClause(v Visitor, node *BLangOnClause) {
	if node.OnExpr != nil {
		Walk(v, node.OnExpr.(BLangNode))
	}
	if node.EqualsExpr != nil {
		Walk(v, node.EqualsExpr.(BLangNode))
	}
}

func walkLimitClause(v Visitor, node *BLangLimitClause) {
	if node.Expression != nil {
		Walk(v, node.Expression.(BLangNode))
	}
}

//go:noinline
func walkOrderByClause(v Visitor, node *BLangOrderByClause) {
	for i := range node.OrderByKeyList {
		Walk(v, &node.OrderByKeyList[i])
	}
}

func walkOrderKey(v Visitor, node *BLangOrderKey) {
	if node.Expression != nil {
		Walk(v, node.Expression.(BLangNode))
	}
}

func walkOnConflictClause(v Visitor, node *BLangOnConflictClause) {
	if node.Expression != nil {
		Walk(v, node.Expression.(BLangNode))
	}
}

func walkOnFailClause(v Visitor, node *BLangOnFailClause) {
	if node.Body != nil {
		Walk(v, node.Body)
	}
	if node.VariableDefinitionNode != nil {
		Walk(v, node.VariableDefinitionNode)
	}
}

func walkDoClause(v Visitor, node *BLangDoClause) {
	if node.Body != nil {
		Walk(v, node.Body)
	}
}

func walkCollectClause(v Visitor, node *BLangCollectClause) {
	if node.Expression != nil {
		Walk(v, node.Expression.(BLangNode))
	}
}

func walkMarkdownDocumentation(v Visitor, node *BLangMarkdownDocumentation) {
	for i := range node.DocumentationLines {
		Walk(v, &node.DocumentationLines[i])
	}
	for i := range node.Parameters {
		Walk(v, &node.Parameters[i])
	}
	for i := range node.References {
		Walk(v, &node.References[i])
	}
	if node.ReturnParameter != nil {
		Walk(v, node.ReturnParameter)
	}
	if node.DeprecationDocumentation != nil {
		Walk(v, node.DeprecationDocumentation)
	}
	if node.DeprecatedParametersDocumentation != nil {
		Walk(v, node.DeprecatedParametersDocumentation)
	}
}

func walkMarkdownParameterDocumentation(v Visitor, node *BLangMarkdownParameterDocumentation) {
	if node.ParameterName != nil {
		Walk(v, node.ParameterName)
	}
}

//go:noinline
func walkMarkDownDeprecatedParametersDocumentation(v Visitor, node *BLangMarkDownDeprecatedParametersDocumentation) {
	for i := range node.Parameters {
		Walk(v, &node.Parameters[i])
	}
}

func walkConstPattern(v Visitor, node *BLangConstPattern) {
	if node.Expr != nil {
		Walk(v, node.Expr.(BLangNode))
	}
}

func walkNamedArgsExpression(v Visitor, node *BLangNamedArgsExpression) {
	if node.Name != nil {
		Walk(v, node.Name)
	}
	Walk(v, node.Expr.(BLangNode))
}

func walkNewExpression(v Visitor, node *BLangNewExpression) {
	if node.TypeDescriptor != nil {
		Walk(v, node.TypeDescriptor)
	}
	for _, arg := range node.ArgsExprs {
		Walk(v, arg.(BLangNode))
	}
}

func walkStartAction(v Visitor, node *BLangStartAction) {
	if node.Call != nil {
		Walk(v, node.Call)
	}
}

func walkSingleWaitAction(v Visitor, node *BLangSingleWaitAction) {
	if node.FutureExpr != nil {
		Walk(v, node.FutureExpr)
	}
}

//go:noinline
func walkAlternateWaitAction(v Visitor, node *BLangAlternateWaitAction) {
	for _, futureExpr := range node.FutureExprs {
		Walk(v, futureExpr)
	}
}

//go:noinline
func walkMultipleWaitAction(v Visitor, node *BLangMultipleWaitAction) {
	for _, futureExpr := range node.FutureExprs {
		Walk(v, futureExpr)
	}
}

func walkRemoteMethodCallAction(v Visitor, node *BLangRemoteMethodCallAction) {
	if node.Name != nil {
		Walk(v, node.Name)
	}
	if node.Expr != nil {
		Walk(v, node.Expr.(BLangNode))
	}
	for _, arg := range node.ArgExprs {
		Walk(v, arg.(BLangNode))
	}
}

func walkClientResourceAccessAction(v Visitor, node *BLangClientResourceAccessAction) {
	if node.Expr != nil {
		Walk(v, node.Expr)
	}
	for i := range node.Path {
		if e := node.Path[i].Expr; e != nil {
			Walk(v, e)
		}
	}
	for _, arg := range node.ArgExprs {
		Walk(v, arg)
	}
}

// We need to do this because TypeData is not a ast node but within it there can be ast nodes. Need to think if this is
// the correct appraoch.
func WalkTypeData(v Visitor, typeData *TypeData) {
	if w := v.VisitTypeData(typeData); w != nil {
		v = w
	}
	if typeData.TypeDescriptor == nil {
		v.Visit(nil)
		return
	}
	td := typeData.TypeDescriptor
	if tdNode, ok := td.(BLangNode); ok {
		Walk(v, tdNode)
	}
	v.Visit(nil)
}

func walkClassDefnBody(v Visitor, b *classDefnBase) {
	for i := range b.AnnAttachments {
		Walk(v, &b.AnnAttachments[i])
	}
	if b.InitFunction != nil {
		Walk(v, b.InitFunction)
	}
	for _, method := range b.Methods {
		Walk(v, method)
	}
	for _, method := range b.ResourceMethods {
		Walk(v, method)
	}
	for _, field := range b.Fields {
		Walk(v, field)
	}
	for _, badNode := range b.BadTopLevelNodes {
		Walk(v, badNode)
	}
	WalkTypeData(v, &b.typeData)
}

func walkTypeDescriptor(v Visitor, td TypeDescriptor) {
	if td == nil {
		return
	}
	if tdNode, ok := td.(BLangNode); ok {
		Walk(v, tdNode)
	}
}
