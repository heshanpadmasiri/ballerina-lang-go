/*
 * Copyright (c) 2025, WSO2 LLC. (http://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package internal

import "ballerina-lang-go/parser/common"

type STModulePart struct {
	STNode

	Imports STNode

	Members STNode

	EofToken STNode
}

func (n *STModulePart) Kind() common.SyntaxKind {
	return common.MODULE_PART
}

type STModuleMemberDeclarationNode = STNode

type STFunctionDefinition struct {
	STModuleMemberDeclarationNode

	Metadata STNode

	QualifierList STNode

	FunctionKeyword STNode

	FunctionName STNode

	RelativeResourcePath STNode

	FunctionSignature STNode

	FunctionBody STNode
}

func (n *STFunctionDefinition) Kind() common.SyntaxKind {
	return common.FUNCTION_DEFINITION
}

type STImportDeclarationNode struct {
	STNode

	ImportKeyword STNode

	OrgName STNode

	ModuleName STNode

	Prefix STNode

	Semicolon STNode
}

func (n *STImportDeclarationNode) Kind() common.SyntaxKind {
	return common.IMPORT_DECLARATION
}

type STListenerDeclarationNode struct {
	STModuleMemberDeclarationNode

	Metadata STNode

	VisibilityQualifier STNode

	ListenerKeyword STNode

	TypeDescriptor STNode

	VariableName STNode

	EqualsToken STNode

	Initializer STNode

	SemicolonToken STNode
}

func (n *STListenerDeclarationNode) Kind() common.SyntaxKind {
	return common.LISTENER_DECLARATION
}

type STTypeDefinitionNode struct {
	STModuleMemberDeclarationNode

	Metadata STNode

	VisibilityQualifier STNode

	TypeKeyword STNode

	TypeName STNode

	TypeDescriptor STNode

	SemicolonToken STNode
}

func (n *STTypeDefinitionNode) Kind() common.SyntaxKind {
	return common.TYPE_DEFINITION
}

type STServiceDeclarationNode struct {
	STModuleMemberDeclarationNode

	Metadata STNode

	Qualifiers STNode

	ServiceKeyword STNode

	TypeDescriptor STNode

	AbsoluteResourcePath STNode

	OnKeyword STNode

	Expressions STNode

	OpenBraceToken STNode

	Members STNode

	CloseBraceToken STNode

	SemicolonToken STNode
}

func (n *STServiceDeclarationNode) Kind() common.SyntaxKind {
	return common.SERVICE_DECLARATION
}

type STStatementNode = STNode

type STAssignmentStatementNode struct {
	STStatementNode

	VarRef STNode

	EqualsToken STNode

	Expression STNode

	SemicolonToken STNode
}

func (n *STAssignmentStatementNode) Kind() common.SyntaxKind {
	return common.ASSIGNMENT_STATEMENT
}

type STCompoundAssignmentStatementNode struct {
	STStatementNode

	LhsExpression STNode

	BinaryOperator STNode

	EqualsToken STNode

	RhsExpression STNode

	SemicolonToken STNode
}

func (n *STCompoundAssignmentStatementNode) Kind() common.SyntaxKind {
	return common.COMPOUND_ASSIGNMENT_STATEMENT
}

type STVariableDeclarationNode struct {
	STStatementNode

	Annotations STNode

	FinalKeyword STNode

	TypedBindingPattern STNode

	EqualsToken STNode

	Initializer STNode

	SemicolonToken STNode
}

func (n *STVariableDeclarationNode) Kind() common.SyntaxKind {
	return common.LOCAL_VAR_DECL
}

type STBlockStatementNode struct {
	STStatementNode

	OpenBraceToken STNode

	Statements STNode

	CloseBraceToken STNode
}

func (n *STBlockStatementNode) Kind() common.SyntaxKind {
	return common.BLOCK_STATEMENT
}

type STBreakStatementNode struct {
	STStatementNode

	BreakToken STNode

	SemicolonToken STNode
}

func (n *STBreakStatementNode) Kind() common.SyntaxKind {
	return common.BREAK_STATEMENT
}

type STFailStatementNode struct {
	STStatementNode

	FailKeyword STNode

	Expression STNode

	SemicolonToken STNode
}

func (n *STFailStatementNode) Kind() common.SyntaxKind {
	return common.FAIL_STATEMENT
}

type STExpressionStatementNode struct {
	STStatementNode

	Expression STNode

	SemicolonToken STNode
}

type STContinueStatementNode struct {
	STStatementNode

	ContinueToken STNode

	SemicolonToken STNode
}

func (n *STContinueStatementNode) Kind() common.SyntaxKind {
	return common.CONTINUE_STATEMENT
}

type STExternalFunctionBodyNode struct {
	STFunctionBodyNode

	EqualsToken STNode

	Annotations STNode

	ExternalKeyword STNode

	SemicolonToken STNode
}

func (n *STExternalFunctionBodyNode) Kind() common.SyntaxKind {
	return common.EXTERNAL_FUNCTION_BODY
}

type STIfElseStatementNode struct {
	STStatementNode

	IfKeyword STNode

	Condition STNode

	IfBody STNode

	ElseBody STNode
}

func (n *STIfElseStatementNode) Kind() common.SyntaxKind {
	return common.IF_ELSE_STATEMENT
}

type STElseBlockNode struct {
	STNode

	ElseKeyword STNode

	ElseBody STNode
}

func (n *STElseBlockNode) Kind() common.SyntaxKind {
	return common.ELSE_BLOCK
}

type STWhileStatementNode struct {
	STStatementNode

	WhileKeyword STNode

	Condition STNode

	WhileBody STNode

	OnFailClause STNode
}

func (n *STWhileStatementNode) Kind() common.SyntaxKind {
	return common.WHILE_STATEMENT
}

type STPanicStatementNode struct {
	STStatementNode

	PanicKeyword STNode

	Expression STNode

	SemicolonToken STNode
}

func (n *STPanicStatementNode) Kind() common.SyntaxKind {
	return common.PANIC_STATEMENT
}

type STReturnStatementNode struct {
	STStatementNode

	ReturnKeyword STNode

	Expression STNode

	SemicolonToken STNode
}

func (n *STReturnStatementNode) Kind() common.SyntaxKind {
	return common.RETURN_STATEMENT
}

type STLocalTypeDefinitionStatementNode struct {
	STStatementNode

	Annotations STNode

	TypeKeyword STNode

	TypeName STNode

	TypeDescriptor STNode

	SemicolonToken STNode
}

func (n *STLocalTypeDefinitionStatementNode) Kind() common.SyntaxKind {
	return common.LOCAL_TYPE_DEFINITION_STATEMENT
}

type STLockStatementNode struct {
	STStatementNode

	LockKeyword STNode

	BlockStatement STNode

	OnFailClause STNode
}

func (n *STLockStatementNode) Kind() common.SyntaxKind {
	return common.LOCK_STATEMENT
}

type STForkStatementNode struct {
	STStatementNode

	ForkKeyword STNode

	OpenBraceToken STNode

	NamedWorkerDeclarations STNode

	CloseBraceToken STNode
}

func (n *STForkStatementNode) Kind() common.SyntaxKind {
	return common.FORK_STATEMENT
}

type STForEachStatementNode struct {
	STStatementNode

	ForEachKeyword STNode

	TypedBindingPattern STNode

	InKeyword STNode

	ActionOrExpressionNode STNode

	BlockStatement STNode

	OnFailClause STNode
}

func (n *STForEachStatementNode) Kind() common.SyntaxKind {
	return common.FOREACH_STATEMENT
}

type STExpressionNode = STNode

type STBinaryExpressionNode struct {
	STExpressionNode

	LhsExpr STNode

	Operator STNode

	RhsExpr STNode
}

type STBracedExpressionNode struct {
	STExpressionNode

	OpenParen STNode

	Expression STNode

	CloseParen STNode
}

type STCheckExpressionNode struct {
	STExpressionNode

	CheckKeyword STNode

	Expression STNode
}

type STFieldAccessExpressionNode struct {
	STExpressionNode

	Expression STNode

	DotToken STNode

	FieldName STNode
}

func (n *STFieldAccessExpressionNode) Kind() common.SyntaxKind {
	return common.FIELD_ACCESS
}

type STFunctionCallExpressionNode struct {
	STExpressionNode

	FunctionName STNode

	OpenParenToken STNode

	Arguments STNode

	CloseParenToken STNode
}

func (n *STFunctionCallExpressionNode) Kind() common.SyntaxKind {
	return common.FUNCTION_CALL
}

type STMethodCallExpressionNode struct {
	STExpressionNode

	Expression STNode

	DotToken STNode

	MethodName STNode

	OpenParenToken STNode

	Arguments STNode

	CloseParenToken STNode
}

func (n *STMethodCallExpressionNode) Kind() common.SyntaxKind {
	return common.METHOD_CALL
}

type STMappingConstructorExpressionNode struct {
	STExpressionNode

	OpenBrace STNode

	Fields STNode

	CloseBrace STNode
}

func (n *STMappingConstructorExpressionNode) Kind() common.SyntaxKind {
	return common.MAPPING_CONSTRUCTOR
}

type STIndexedExpressionNode struct {
	STTypeDescriptorNode

	ContainerExpression STNode

	OpenBracket STNode

	KeyExpression STNode

	CloseBracket STNode
}

func (n *STIndexedExpressionNode) Kind() common.SyntaxKind {
	return common.INDEXED_EXPRESSION
}

type STTypeofExpressionNode struct {
	STExpressionNode

	TypeofKeyword STNode

	Expression STNode
}

func (n *STTypeofExpressionNode) Kind() common.SyntaxKind {
	return common.TYPEOF_EXPRESSION
}

type STUnaryExpressionNode struct {
	STExpressionNode

	UnaryOperator STNode

	Expression STNode
}

func (n *STUnaryExpressionNode) Kind() common.SyntaxKind {
	return common.UNARY_EXPRESSION
}

type STComputedNameFieldNode struct {
	STMappingFieldNode

	OpenBracket STNode

	FieldNameExpr STNode

	CloseBracket STNode

	ColonToken STNode

	ValueExpr STNode
}

func (n *STComputedNameFieldNode) Kind() common.SyntaxKind {
	return common.COMPUTED_NAME_FIELD
}

type STConstantDeclarationNode struct {
	STModuleMemberDeclarationNode

	Metadata STNode

	VisibilityQualifier STNode

	ConstKeyword STNode

	TypeDescriptor STNode

	VariableName STNode

	EqualsToken STNode

	Initializer STNode

	SemicolonToken STNode
}

func (n *STConstantDeclarationNode) Kind() common.SyntaxKind {
	return common.CONST_DECLARATION
}

type STParameterNode = STNode

type STDefaultableParameterNode struct {
	STParameterNode

	Annotations STNode

	TypeName STNode

	ParamName STNode

	EqualsToken STNode

	Expression STNode
}

func (n *STDefaultableParameterNode) Kind() common.SyntaxKind {
	return common.DEFAULTABLE_PARAM
}

type STRequiredParameterNode struct {
	STParameterNode

	Annotations STNode

	TypeName STNode

	ParamName STNode
}

func (n *STRequiredParameterNode) Kind() common.SyntaxKind {
	return common.REQUIRED_PARAM
}

type STIncludedRecordParameterNode struct {
	STParameterNode

	Annotations STNode

	AsteriskToken STNode

	TypeName STNode

	ParamName STNode
}

func (n *STIncludedRecordParameterNode) Kind() common.SyntaxKind {
	return common.INCLUDED_RECORD_PARAM
}

type STRestParameterNode struct {
	STParameterNode

	Annotations STNode

	TypeName STNode

	EllipsisToken STNode

	ParamName STNode
}

func (n *STRestParameterNode) Kind() common.SyntaxKind {
	return common.REST_PARAM
}

type STImportOrgNameNode struct {
	STNode

	OrgName STNode

	SlashToken STNode
}

func (n *STImportOrgNameNode) Kind() common.SyntaxKind {
	return common.IMPORT_ORG_NAME
}

type STImportPrefixNode struct {
	STNode

	AsKeyword STNode

	Prefix STNode
}

func (n *STImportPrefixNode) Kind() common.SyntaxKind {
	return common.IMPORT_PREFIX
}

type STMappingFieldNode = STNode

type STSpecificFieldNode struct {
	STMappingFieldNode

	ReadonlyKeyword STNode

	FieldName STNode

	Colon STNode

	ValueExpr STNode
}

func (n *STSpecificFieldNode) Kind() common.SyntaxKind {
	return common.SPECIFIC_FIELD
}

type STSpreadFieldNode struct {
	STMappingFieldNode

	Ellipsis STNode

	ValueExpr STNode
}

func (n *STSpreadFieldNode) Kind() common.SyntaxKind {
	return common.SPREAD_FIELD
}

type STFunctionArgumentNode = STNode

type STNamedArgumentNode struct {
	STFunctionArgumentNode

	ArgumentName STNode

	EqualsToken STNode

	Expression STNode
}

func (n *STNamedArgumentNode) Kind() common.SyntaxKind {
	return common.NAMED_ARG
}

type STPositionalArgumentNode struct {
	STFunctionArgumentNode

	Expression STNode
}

func (n *STPositionalArgumentNode) Kind() common.SyntaxKind {
	return common.POSITIONAL_ARG
}

type STRestArgumentNode struct {
	STFunctionArgumentNode

	Ellipsis STNode

	Expression STNode
}

func (n *STRestArgumentNode) Kind() common.SyntaxKind {
	return common.REST_ARG
}

type STInferredTypedescDefaultNode struct {
	STExpressionNode

	LtToken STNode

	GtToken STNode
}

func (n *STInferredTypedescDefaultNode) Kind() common.SyntaxKind {
	return common.INFERRED_TYPEDESC_DEFAULT
}

type STObjectTypeDescriptorNode struct {
	STTypeDescriptorNode

	ObjectTypeQualifiers STNode

	ObjectKeyword STNode

	OpenBrace STNode

	Members STNode

	CloseBrace STNode
}

func (n *STObjectTypeDescriptorNode) Kind() common.SyntaxKind {
	return common.OBJECT_TYPE_DESC
}

type STObjectConstructorExpressionNode struct {
	STExpressionNode

	Annotations STNode

	ObjectTypeQualifiers STNode

	ObjectKeyword STNode

	TypeReference STNode

	OpenBraceToken STNode

	Members STNode

	CloseBraceToken STNode
}

func (n *STObjectConstructorExpressionNode) Kind() common.SyntaxKind {
	return common.OBJECT_CONSTRUCTOR
}

type STRecordTypeDescriptorNode struct {
	STTypeDescriptorNode

	RecordKeyword STNode

	BodyStartDelimiter STNode

	Fields STNode

	RecordRestDescriptor STNode

	BodyEndDelimiter STNode
}

func (n *STRecordTypeDescriptorNode) Kind() common.SyntaxKind {
	return common.RECORD_TYPE_DESC
}

type STReturnTypeDescriptorNode struct {
	STNode

	ReturnsKeyword STNode

	Annotations STNode

	Type STNode
}

func (n *STReturnTypeDescriptorNode) Kind() common.SyntaxKind {
	return common.RETURN_TYPE_DESCRIPTOR
}

type STNilTypeDescriptorNode struct {
	STTypeDescriptorNode

	OpenParenToken STNode

	CloseParenToken STNode
}

func (n *STNilTypeDescriptorNode) Kind() common.SyntaxKind {
	return common.NIL_TYPE_DESC
}

type STOptionalTypeDescriptorNode struct {
	STTypeDescriptorNode

	TypeDescriptor STNode

	QuestionMarkToken STNode
}

func (n *STOptionalTypeDescriptorNode) Kind() common.SyntaxKind {
	return common.OPTIONAL_TYPE_DESC
}

type STObjectFieldNode struct {
	STNode

	Metadata STNode

	VisibilityQualifier STNode

	QualifierList STNode

	TypeName STNode

	FieldName STNode

	EqualsToken STNode

	Expression STNode

	SemicolonToken STNode
}

func (n *STObjectFieldNode) Kind() common.SyntaxKind {
	return common.OBJECT_FIELD
}

type STRecordFieldNode struct {
	STNode

	Metadata STNode

	ReadonlyKeyword STNode

	TypeName STNode

	FieldName STNode

	QuestionMarkToken STNode

	SemicolonToken STNode
}

func (n *STRecordFieldNode) Kind() common.SyntaxKind {
	return common.RECORD_FIELD
}

type STRecordFieldWithDefaultValueNode struct {
	STNode

	Metadata STNode

	ReadonlyKeyword STNode

	TypeName STNode

	FieldName STNode

	EqualsToken STNode

	Expression STNode

	SemicolonToken STNode
}

func (n *STRecordFieldWithDefaultValueNode) Kind() common.SyntaxKind {
	return common.RECORD_FIELD_WITH_DEFAULT_VALUE
}

type STRecordRestDescriptorNode struct {
	STNode

	TypeName STNode

	EllipsisToken STNode

	SemicolonToken STNode
}

func (n *STRecordRestDescriptorNode) Kind() common.SyntaxKind {
	return common.RECORD_REST_TYPE
}

type STTypeReferenceNode struct {
	STTypeDescriptorNode

	AsteriskToken STNode

	TypeName STNode

	SemicolonToken STNode
}

func (n *STTypeReferenceNode) Kind() common.SyntaxKind {
	return common.TYPE_REFERENCE
}

type STAnnotationNode struct {
	STNode

	AtToken STNode

	AnnotReference STNode

	AnnotValue STNode
}

func (n *STAnnotationNode) Kind() common.SyntaxKind {
	return common.ANNOTATION
}

type STMetadataNode struct {
	STNode

	DocumentationString STNode

	Annotations STNode
}

func (n STMetadataNode) Kind() common.SyntaxKind {
	return common.METADATA
}

type STModuleVariableDeclarationNode struct {
	STModuleMemberDeclarationNode

	Metadata STNode

	VisibilityQualifier STNode

	Qualifiers STNode

	TypedBindingPattern STNode

	EqualsToken STNode

	Initializer STNode

	SemicolonToken STNode
}

func (n *STModuleVariableDeclarationNode) Kind() common.SyntaxKind {
	return common.MODULE_VAR_DECL
}

type STTypeTestExpressionNode struct {
	STExpressionNode

	Expression STNode

	IsKeyword STNode

	TypeDescriptor STNode
}

func (n *STTypeTestExpressionNode) Kind() common.SyntaxKind {
	return common.TYPE_TEST_EXPRESSION
}

type STActionNode = STExpressionNode

type STRemoteMethodCallActionNode struct {
	STActionNode

	Expression STNode

	RightArrowToken STNode

	MethodName STNode

	OpenParenToken STNode

	Arguments STNode

	CloseParenToken STNode
}

func (n *STRemoteMethodCallActionNode) Kind() common.SyntaxKind {
	return common.REMOTE_METHOD_CALL_ACTION
}

type STMapTypeDescriptorNode struct {
	STTypeDescriptorNode

	MapKeywordToken STNode

	MapTypeParamsNode STNode
}

func (n *STMapTypeDescriptorNode) Kind() common.SyntaxKind {
	return common.MAP_TYPE_DESC
}

type STNilLiteralNode struct {
	STExpressionNode

	OpenParenToken STNode

	CloseParenToken STNode
}

func (n *STNilLiteralNode) Kind() common.SyntaxKind {
	return common.NIL_LITERAL
}

type STAnnotationDeclarationNode struct {
	STModuleMemberDeclarationNode

	Metadata STNode

	VisibilityQualifier STNode

	ConstKeyword STNode

	AnnotationKeyword STNode

	TypeDescriptor STNode

	AnnotationTag STNode

	OnKeyword STNode

	AttachPoints STNode

	SemicolonToken STNode
}

func (n *STAnnotationDeclarationNode) Kind() common.SyntaxKind {
	return common.ANNOTATION_DECLARATION
}

type STAnnotationAttachPointNode struct {
	STNode

	SourceKeyword STNode

	Identifiers STNode
}

func (n *STAnnotationAttachPointNode) Kind() common.SyntaxKind {
	return common.ANNOTATION_ATTACH_POINT
}

type STXMLNamespaceDeclarationNode struct {
	STStatementNode

	XmlnsKeyword STNode

	Namespaceuri STNode

	AsKeyword STNode

	NamespacePrefix STNode

	SemicolonToken STNode
}

func (n *STXMLNamespaceDeclarationNode) Kind() common.SyntaxKind {
	return common.XML_NAMESPACE_DECLARATION
}

type STModuleXMLNamespaceDeclarationNode struct {
	STModuleMemberDeclarationNode

	XmlnsKeyword STNode

	Namespaceuri STNode

	AsKeyword STNode

	NamespacePrefix STNode

	SemicolonToken STNode
}

func (n *STModuleXMLNamespaceDeclarationNode) Kind() common.SyntaxKind {
	return common.MODULE_XML_NAMESPACE_DECLARATION
}

type STFunctionBodyBlockNode struct {
	STFunctionBodyNode

	OpenBraceToken STNode

	NamedWorkerDeclarator STNode

	Statements STNode

	CloseBraceToken STNode

	SemicolonToken STNode
}

func (n *STFunctionBodyBlockNode) Kind() common.SyntaxKind {
	return common.FUNCTION_BODY_BLOCK
}

type STNamedWorkerDeclarationNode struct {
	STNode

	Annotations STNode

	TransactionalKeyword STNode

	WorkerKeyword STNode

	WorkerName STNode

	ReturnTypeDesc STNode

	WorkerBody STNode

	OnFailClause STNode
}

func (n *STNamedWorkerDeclarationNode) Kind() common.SyntaxKind {
	return common.NAMED_WORKER_DECLARATION
}

type STNamedWorkerDeclarator struct {
	STNode

	WorkerInitStatements STNode

	NamedWorkerDeclarations STNode
}

func (n *STNamedWorkerDeclarator) Kind() common.SyntaxKind {
	return common.NAMED_WORKER_DECLARATOR
}

type STBasicLiteralNode struct {
	STExpressionNode

	LiteralToken STNode
}

type STTypeDescriptorNode = STExpressionNode

type STNameReferenceNode = STTypeDescriptorNode

type STSimpleNameReferenceNode struct {
	STNameReferenceNode

	Name STNode
}

func (n *STSimpleNameReferenceNode) Kind() common.SyntaxKind {
	return common.SIMPLE_NAME_REFERENCE
}

type STQualifiedNameReferenceNode struct {
	STNameReferenceNode

	ModulePrefix STNode

	Colon STNode

	Identifier STNode
}

func (n *STQualifiedNameReferenceNode) Kind() common.SyntaxKind {
	return common.QUALIFIED_NAME_REFERENCE
}

type STBuiltinSimpleNameReferenceNode struct {
	STNameReferenceNode

	Name STNode
}

type STTrapExpressionNode struct {
	STExpressionNode

	TrapKeyword STNode

	Expression STNode
}

type STListConstructorExpressionNode struct {
	STExpressionNode

	OpenBracket STNode

	Expressions STNode

	CloseBracket STNode
}

func (n *STListConstructorExpressionNode) Kind() common.SyntaxKind {
	return common.LIST_CONSTRUCTOR
}

type STTypeCastExpressionNode struct {
	STExpressionNode

	LtToken STNode

	TypeCastParam STNode

	GtToken STNode

	Expression STNode
}

func (n *STTypeCastExpressionNode) Kind() common.SyntaxKind {
	return common.TYPE_CAST_EXPRESSION
}

type STTypeCastParamNode struct {
	STNode

	Annotations STNode

	Type STNode
}

func (n *STTypeCastParamNode) Kind() common.SyntaxKind {
	return common.TYPE_CAST_PARAM
}

type STUnionTypeDescriptorNode struct {
	STTypeDescriptorNode

	LeftTypeDesc STNode

	PipeToken STNode

	RightTypeDesc STNode
}

func (n *STUnionTypeDescriptorNode) Kind() common.SyntaxKind {
	return common.UNION_TYPE_DESC
}

type STTableConstructorExpressionNode struct {
	STExpressionNode

	TableKeyword STNode

	KeySpecifier STNode

	OpenBracket STNode

	Rows STNode

	CloseBracket STNode
}

func (n *STTableConstructorExpressionNode) Kind() common.SyntaxKind {
	return common.TABLE_CONSTRUCTOR
}

type STKeySpecifierNode struct {
	STNode

	KeyKeyword STNode

	OpenParenToken STNode

	FieldNames STNode

	CloseParenToken STNode
}

func (n *STKeySpecifierNode) Kind() common.SyntaxKind {
	return common.KEY_SPECIFIER
}

type STStreamTypeDescriptorNode struct {
	STTypeDescriptorNode

	StreamKeywordToken STNode

	StreamTypeParamsNode STNode
}

func (n *STStreamTypeDescriptorNode) Kind() common.SyntaxKind {
	return common.STREAM_TYPE_DESC
}

type STStreamTypeParamsNode struct {
	STNode

	LtToken STNode

	LeftTypeDescNode STNode

	CommaToken STNode

	RightTypeDescNode STNode

	GtToken STNode
}

func (n *STStreamTypeParamsNode) Kind() common.SyntaxKind {
	return common.STREAM_TYPE_PARAMS
}

type STLetExpressionNode struct {
	STExpressionNode

	LetKeyword STNode

	LetVarDeclarations STNode

	InKeyword STNode

	Expression STNode
}

func (n *STLetExpressionNode) Kind() common.SyntaxKind {
	return common.LET_EXPRESSION
}

type STLetVariableDeclarationNode struct {
	STNode

	Annotations STNode

	TypedBindingPattern STNode

	EqualsToken STNode

	Expression STNode
}

func (n *STLetVariableDeclarationNode) Kind() common.SyntaxKind {
	return common.LET_VAR_DECL
}

type STTemplateExpressionNode struct {
	STExpressionNode

	Type STNode

	StartBacktick STNode

	Content STNode

	EndBacktick STNode
}

type STXMLItemNode = STNode

type STXMLElementNode struct {
	STXMLItemNode

	StartTag STNode

	Content STNode

	EndTag STNode
}

func (n *STXMLElementNode) Kind() common.SyntaxKind {
	return common.XML_ELEMENT
}

type STXMLElementTagNode = STNode

type STXMLStartTagNode struct {
	STXMLElementTagNode

	LtToken STNode

	Name STNode

	Attributes STNode

	GetToken STNode
}

func (n *STXMLStartTagNode) Kind() common.SyntaxKind {
	return common.XML_ELEMENT_START_TAG
}

type STXMLEndTagNode struct {
	STXMLElementTagNode

	LtToken STNode

	SlashToken STNode

	Name STNode

	GetToken STNode
}

func (n *STXMLEndTagNode) Kind() common.SyntaxKind {
	return common.XML_ELEMENT_END_TAG
}

type STXMLNameNode = STNode

type STXMLSimpleNameNode struct {
	STXMLNameNode

	Name STNode
}

func (n *STXMLSimpleNameNode) Kind() common.SyntaxKind {
	return common.XML_SIMPLE_NAME
}

type STXMLQualifiedNameNode struct {
	STXMLNameNode

	Prefix STNode

	Colon STNode

	Name STNode
}

func (n *STXMLQualifiedNameNode) Kind() common.SyntaxKind {
	return common.XML_QUALIFIED_NAME
}

type STXMLEmptyElementNode struct {
	STXMLItemNode

	LtToken STNode

	Name STNode

	Attributes STNode

	SlashToken STNode

	GetToken STNode
}

func (n *STXMLEmptyElementNode) Kind() common.SyntaxKind {
	return common.XML_EMPTY_ELEMENT
}

type STInterpolationNode struct {
	STXMLItemNode

	InterpolationStartToken STNode

	Expression STNode

	InterpolationEndToken STNode
}

func (n *STInterpolationNode) Kind() common.SyntaxKind {
	return common.INTERPOLATION
}

type STXMLTextNode struct {
	STXMLItemNode

	Content STNode
}

func (n *STXMLTextNode) Kind() common.SyntaxKind {
	return common.XML_TEXT
}

type STXMLAttributeNode struct {
	STNode

	AttributeName STNode

	EqualToken STNode

	Value STNode
}

func (n *STXMLAttributeNode) Kind() common.SyntaxKind {
	return common.XML_ATTRIBUTE
}

type STXMLAttributeValue struct {
	STNode

	StartQuote STNode

	Value STNode

	EndQuote STNode
}

func (n *STXMLAttributeValue) Kind() common.SyntaxKind {
	return common.XML_ATTRIBUTE_VALUE
}

type STXMLComment struct {
	STXMLItemNode

	CommentStart STNode

	Content STNode

	CommentEnd STNode
}

func (n *STXMLComment) Kind() common.SyntaxKind {
	return common.XML_COMMENT
}

type STXMLCDATANode struct {
	STXMLItemNode

	CdataStart STNode

	Content STNode

	CdataEnd STNode
}

func (n *STXMLCDATANode) Kind() common.SyntaxKind {
	return common.XML_CDATA
}

type STXMLProcessingInstruction struct {
	STXMLItemNode

	PiStart STNode

	Target STNode

	Data STNode

	PiEnd STNode
}

func (n *STXMLProcessingInstruction) Kind() common.SyntaxKind {
	return common.XML_PI
}

type STTableTypeDescriptorNode struct {
	STTypeDescriptorNode

	TableKeywordToken STNode

	RowTypeParameterNode STNode

	KeyConstraintNode STNode
}

func (n *STTableTypeDescriptorNode) Kind() common.SyntaxKind {
	return common.TABLE_TYPE_DESC
}

type STTypeParameterNode struct {
	STNode

	LtToken STNode

	TypeNode STNode

	GtToken STNode
}

func (n *STTypeParameterNode) Kind() common.SyntaxKind {
	return common.TYPE_PARAMETER
}

type STKeyTypeConstraintNode struct {
	STNode

	KeyKeywordToken STNode

	TypeParameterNode STNode
}

func (n *STKeyTypeConstraintNode) Kind() common.SyntaxKind {
	return common.KEY_TYPE_CONSTRAINT
}

type STFunctionTypeDescriptorNode struct {
	STTypeDescriptorNode

	QualifierList STNode

	FunctionKeyword STNode

	FunctionSignature STNode
}

func (n *STFunctionTypeDescriptorNode) Kind() common.SyntaxKind {
	return common.FUNCTION_TYPE_DESC
}

type STFunctionSignatureNode struct {
	STNode

	OpenParenToken STNode

	Parameters STNode

	CloseParenToken STNode

	ReturnTypeDesc STNode
}

func (n *STFunctionSignatureNode) Kind() common.SyntaxKind {
	return common.FUNCTION_SIGNATURE
}

type STAnonymousFunctionExpressionNode = STExpressionNode

type STExplicitAnonymousFunctionExpressionNode struct {
	STAnonymousFunctionExpressionNode

	Annotations STNode

	QualifierList STNode

	FunctionKeyword STNode

	FunctionSignature STNode

	FunctionBody STNode
}

func (n *STExplicitAnonymousFunctionExpressionNode) Kind() common.SyntaxKind {
	return common.EXPLICIT_ANONYMOUS_FUNCTION_EXPRESSION
}

type STFunctionBodyNode = STNode

type STExpressionFunctionBodyNode struct {
	STFunctionBodyNode

	RightDoubleArrow STNode

	Expression STNode

	Semicolon STNode
}

func (n *STExpressionFunctionBodyNode) Kind() common.SyntaxKind {
	return common.EXPRESSION_FUNCTION_BODY
}

type STTupleTypeDescriptorNode struct {
	STTypeDescriptorNode

	OpenBracketToken STNode

	MemberTypeDesc STNode

	CloseBracketToken STNode
}

func (n *STTupleTypeDescriptorNode) Kind() common.SyntaxKind {
	return common.TUPLE_TYPE_DESC
}

type STParenthesisedTypeDescriptorNode struct {
	STTypeDescriptorNode

	OpenParenToken STNode

	Typedesc STNode

	CloseParenToken STNode
}

func (n *STParenthesisedTypeDescriptorNode) Kind() common.SyntaxKind {
	return common.PARENTHESISED_TYPE_DESC
}

type STNewExpressionNode = STExpressionNode

type STExplicitNewExpressionNode struct {
	STNewExpressionNode

	NewKeyword STNode

	TypeDescriptor STNode

	ParenthesizedArgList STNode
}

func (n *STExplicitNewExpressionNode) Kind() common.SyntaxKind {
	return common.EXPLICIT_NEW_EXPRESSION
}

type STImplicitNewExpressionNode struct {
	STNewExpressionNode

	NewKeyword STNode

	ParenthesizedArgList STNode
}

func (n *STImplicitNewExpressionNode) Kind() common.SyntaxKind {
	return common.IMPLICIT_NEW_EXPRESSION
}

type STParenthesizedArgList struct {
	STNode

	OpenParenToken STNode

	Arguments STNode

	CloseParenToken STNode
}

func (n *STParenthesizedArgList) Kind() common.SyntaxKind {
	return common.PARENTHESIZED_ARG_LIST
}

type STClauseNode = STNode

type STIntermediateClauseNode = STClauseNode

type STQueryConstructTypeNode struct {
	STNode

	Keyword STNode

	KeySpecifier STNode
}

func (n *STQueryConstructTypeNode) Kind() common.SyntaxKind {
	return common.QUERY_CONSTRUCT_TYPE
}

type STFromClauseNode struct {
	STIntermediateClauseNode

	FromKeyword STNode

	TypedBindingPattern STNode

	InKeyword STNode

	Expression STNode
}

func (n *STFromClauseNode) Kind() common.SyntaxKind {
	return common.FROM_CLAUSE
}

type STWhereClauseNode struct {
	STIntermediateClauseNode

	WhereKeyword STNode

	Expression STNode
}

func (n *STWhereClauseNode) Kind() common.SyntaxKind {
	return common.WHERE_CLAUSE
}

type STLetClauseNode struct {
	STIntermediateClauseNode

	LetKeyword STNode

	LetVarDeclarations STNode
}

func (n *STLetClauseNode) Kind() common.SyntaxKind {
	return common.LET_CLAUSE
}

type STJoinClauseNode struct {
	STIntermediateClauseNode

	OuterKeyword STNode

	JoinKeyword STNode

	TypedBindingPattern STNode

	InKeyword STNode

	Expression STNode

	JoinOnCondition STNode
}

func (n *STJoinClauseNode) Kind() common.SyntaxKind {
	return common.JOIN_CLAUSE
}

type STOnClauseNode struct {
	STClauseNode

	OnKeyword STNode

	LhsExpression STNode

	EqualsKeyword STNode

	RhsExpression STNode
}

func (n *STOnClauseNode) Kind() common.SyntaxKind {
	return common.ON_CLAUSE
}

type STLimitClauseNode struct {
	STIntermediateClauseNode

	LimitKeyword STNode

	Expression STNode
}

func (n *STLimitClauseNode) Kind() common.SyntaxKind {
	return common.LIMIT_CLAUSE
}

type STOnConflictClauseNode struct {
	STClauseNode

	OnKeyword STNode

	ConflictKeyword STNode

	Expression STNode
}

func (n *STOnConflictClauseNode) Kind() common.SyntaxKind {
	return common.ON_CONFLICT_CLAUSE
}

type STQueryPipelineNode struct {
	STNode

	FromClause STNode

	IntermediateClauses STNode
}

func (n *STQueryPipelineNode) Kind() common.SyntaxKind {
	return common.QUERY_PIPELINE
}

type STSelectClauseNode struct {
	STClauseNode

	SelectKeyword STNode

	Expression STNode
}

func (n *STSelectClauseNode) Kind() common.SyntaxKind {
	return common.SELECT_CLAUSE
}

type STCollectClauseNode struct {
	STClauseNode

	CollectKeyword STNode

	Expression STNode
}

func (n *STCollectClauseNode) Kind() common.SyntaxKind {
	return common.COLLECT_CLAUSE
}

type STQueryExpressionNode struct {
	STExpressionNode

	QueryConstructType STNode

	QueryPipeline STNode

	ResultClause STNode

	OnConflictClause STNode
}

func (n *STQueryExpressionNode) Kind() common.SyntaxKind {
	return common.QUERY_EXPRESSION
}

type STQueryActionNode struct {
	STActionNode

	QueryPipeline STNode

	DoKeyword STNode

	BlockStatement STNode
}

func (n *STQueryActionNode) Kind() common.SyntaxKind {
	return common.QUERY_ACTION
}

type STIntersectionTypeDescriptorNode struct {
	STTypeDescriptorNode

	LeftTypeDesc STNode

	BitwiseAndToken STNode

	RightTypeDesc STNode
}

func (n *STIntersectionTypeDescriptorNode) Kind() common.SyntaxKind {
	return common.INTERSECTION_TYPE_DESC
}

type STImplicitAnonymousFunctionParameters struct {
	STNode

	OpenParenToken STNode

	Parameters STNode

	CloseParenToken STNode
}

func (n *STImplicitAnonymousFunctionParameters) Kind() common.SyntaxKind {
	return common.INFER_PARAM_LIST
}

type STImplicitAnonymousFunctionExpressionNode struct {
	STAnonymousFunctionExpressionNode

	Params STNode

	RightDoubleArrow STNode

	Expression STNode
}

func (n *STImplicitAnonymousFunctionExpressionNode) Kind() common.SyntaxKind {
	return common.IMPLICIT_ANONYMOUS_FUNCTION_EXPRESSION
}

type STStartActionNode struct {
	STExpressionNode

	Annotations STNode

	StartKeyword STNode

	Expression STNode
}

func (n *STStartActionNode) Kind() common.SyntaxKind {
	return common.START_ACTION
}

type STFlushActionNode struct {
	STExpressionNode

	FlushKeyword STNode

	PeerWorker STNode
}

func (n *STFlushActionNode) Kind() common.SyntaxKind {
	return common.FLUSH_ACTION
}

type STSingletonTypeDescriptorNode struct {
	STTypeDescriptorNode

	SimpleContExprNode STNode
}

func (n *STSingletonTypeDescriptorNode) Kind() common.SyntaxKind {
	return common.SINGLETON_TYPE_DESC
}

type STMethodDeclarationNode struct {
	STNode

	Metadata STNode

	QualifierList STNode

	FunctionKeyword STNode

	MethodName STNode

	RelativeResourcePath STNode

	MethodSignature STNode

	Semicolon STNode
}

type STTypedBindingPatternNode struct {
	STNode

	TypeDescriptor STNode

	BindingPattern STNode
}

func (n *STTypedBindingPatternNode) Kind() common.SyntaxKind {
	return common.TYPED_BINDING_PATTERN
}

type STBindingPatternNode = STNode

type STCaptureBindingPatternNode struct {
	STBindingPatternNode

	VariableName STNode
}

func (n *STCaptureBindingPatternNode) Kind() common.SyntaxKind {
	return common.CAPTURE_BINDING_PATTERN
}

type STWildcardBindingPatternNode struct {
	STBindingPatternNode

	UnderscoreToken STNode
}

func (n *STWildcardBindingPatternNode) Kind() common.SyntaxKind {
	return common.WILDCARD_BINDING_PATTERN
}

type STListBindingPatternNode struct {
	STBindingPatternNode

	OpenBracket STNode

	BindingPatterns STNode

	CloseBracket STNode
}

func (n *STListBindingPatternNode) Kind() common.SyntaxKind {
	return common.LIST_BINDING_PATTERN
}

type STMappingBindingPatternNode struct {
	STBindingPatternNode

	OpenBrace STNode

	FieldBindingPatterns STNode

	CloseBrace STNode
}

func (n *STMappingBindingPatternNode) Kind() common.SyntaxKind {
	return common.MAPPING_BINDING_PATTERN
}

type STFieldBindingPatternNode = STBindingPatternNode

type STFieldBindingPatternFullNode struct {
	STFieldBindingPatternNode

	VariableName STNode

	Colon STNode

	BindingPattern STNode
}

func (n *STFieldBindingPatternFullNode) Kind() common.SyntaxKind {
	return common.FIELD_BINDING_PATTERN
}

type STFieldBindingPatternVarnameNode struct {
	STFieldBindingPatternNode

	VariableName STNode
}

func (n *STFieldBindingPatternVarnameNode) Kind() common.SyntaxKind {
	return common.FIELD_BINDING_PATTERN
}

type STRestBindingPatternNode struct {
	STBindingPatternNode

	EllipsisToken STNode

	VariableName STNode
}

func (n *STRestBindingPatternNode) Kind() common.SyntaxKind {
	return common.REST_BINDING_PATTERN
}

type STErrorBindingPatternNode struct {
	STBindingPatternNode

	ErrorKeyword STNode

	TypeReference STNode

	OpenParenthesis STNode

	ArgListBindingPatterns STNode

	CloseParenthesis STNode
}

func (n *STErrorBindingPatternNode) Kind() common.SyntaxKind {
	return common.ERROR_BINDING_PATTERN
}

type STNamedArgBindingPatternNode struct {
	STBindingPatternNode

	ArgName STNode

	EqualsToken STNode

	BindingPattern STNode
}

func (n *STNamedArgBindingPatternNode) Kind() common.SyntaxKind {
	return common.NAMED_ARG_BINDING_PATTERN
}

type STAsyncSendActionNode struct {
	STActionNode

	Expression STNode

	RightArrowToken STNode

	PeerWorker STNode
}

func (n *STAsyncSendActionNode) Kind() common.SyntaxKind {
	return common.ASYNC_SEND_ACTION
}

type STSyncSendActionNode struct {
	STActionNode

	Expression STNode

	SyncSendToken STNode

	PeerWorker STNode
}

func (n *STSyncSendActionNode) Kind() common.SyntaxKind {
	return common.SYNC_SEND_ACTION
}

type STReceiveActionNode struct {
	STActionNode

	LeftArrow STNode

	ReceiveWorkers STNode
}

func (n *STReceiveActionNode) Kind() common.SyntaxKind {
	return common.RECEIVE_ACTION
}

type STReceiveFieldsNode struct {
	STNode

	OpenBrace STNode

	ReceiveFields STNode

	CloseBrace STNode
}

func (n *STReceiveFieldsNode) Kind() common.SyntaxKind {
	return common.RECEIVE_FIELDS
}

type STAlternateReceiveNode struct {
	STNode

	Workers STNode
}

func (n *STAlternateReceiveNode) Kind() common.SyntaxKind {
	return common.ALTERNATE_RECEIVE
}

type STRestDescriptorNode struct {
	STNode

	TypeDescriptor STNode

	EllipsisToken STNode
}

func (n *STRestDescriptorNode) Kind() common.SyntaxKind {
	return common.REST_TYPE
}

type STDoubleGTTokenNode struct {
	STNode

	OpenGTToken STNode

	EndGTToken STNode
}

func (n *STDoubleGTTokenNode) Kind() common.SyntaxKind {
	return common.DOUBLE_GT_TOKEN
}

type STTrippleGTTokenNode struct {
	STNode

	OpenGTToken STNode

	MiddleGTToken STNode

	EndGTToken STNode
}

func (n *STTrippleGTTokenNode) Kind() common.SyntaxKind {
	return common.TRIPPLE_GT_TOKEN
}

type STWaitActionNode struct {
	STActionNode

	WaitKeyword STNode

	WaitFutureExpr STNode
}

func (n *STWaitActionNode) Kind() common.SyntaxKind {
	return common.WAIT_ACTION
}

type STWaitFieldsListNode struct {
	STNode

	OpenBrace STNode

	WaitFields STNode

	CloseBrace STNode
}

func (n *STWaitFieldsListNode) Kind() common.SyntaxKind {
	return common.WAIT_FIELDS_LIST
}

type STWaitFieldNode struct {
	STNode

	FieldName STNode

	Colon STNode

	WaitFutureExpr STNode
}

func (n *STWaitFieldNode) Kind() common.SyntaxKind {
	return common.WAIT_FIELD
}

type STAnnotAccessExpressionNode struct {
	STExpressionNode

	Expression STNode

	AnnotChainingToken STNode

	AnnotTagReference STNode
}

func (n *STAnnotAccessExpressionNode) Kind() common.SyntaxKind {
	return common.ANNOT_ACCESS
}

type STOptionalFieldAccessExpressionNode struct {
	STExpressionNode

	Expression STNode

	OptionalChainingToken STNode

	FieldName STNode
}

func (n *STOptionalFieldAccessExpressionNode) Kind() common.SyntaxKind {
	return common.OPTIONAL_FIELD_ACCESS
}

type STConditionalExpressionNode struct {
	STExpressionNode

	LhsExpression STNode

	QuestionMarkToken STNode

	MiddleExpression STNode

	ColonToken STNode

	EndExpression STNode
}

func (n *STConditionalExpressionNode) Kind() common.SyntaxKind {
	return common.CONDITIONAL_EXPRESSION
}

type STEnumDeclarationNode struct {
	STModuleMemberDeclarationNode

	Metadata STNode

	Qualifier STNode

	EnumKeywordToken STNode

	Identifier STNode

	OpenBraceToken STNode

	EnumMemberList STNode

	CloseBraceToken STNode

	SemicolonToken STNode
}

func (n *STEnumDeclarationNode) Kind() common.SyntaxKind {
	return common.ENUM_DECLARATION
}

type STEnumMemberNode struct {
	STNode

	Metadata STNode

	Identifier STNode

	EqualToken STNode

	ConstExprNode STNode
}

func (n *STEnumMemberNode) Kind() common.SyntaxKind {
	return common.ENUM_MEMBER
}

type STArrayTypeDescriptorNode struct {
	STTypeDescriptorNode

	MemberTypeDesc STNode

	Dimensions STNode
}

func (n *STArrayTypeDescriptorNode) Kind() common.SyntaxKind {
	return common.ARRAY_TYPE_DESC
}

type STArrayDimensionNode struct {
	STNode

	OpenBracket STNode

	ArrayLength STNode

	CloseBracket STNode
}

func (n *STArrayDimensionNode) Kind() common.SyntaxKind {
	return common.ARRAY_DIMENSION
}

type STTransactionStatementNode struct {
	STStatementNode

	TransactionKeyword STNode

	BlockStatement STNode

	OnFailClause STNode
}

func (n *STTransactionStatementNode) Kind() common.SyntaxKind {
	return common.TRANSACTION_STATEMENT
}

type STRollbackStatementNode struct {
	STStatementNode

	RollbackKeyword STNode

	Expression STNode

	Semicolon STNode
}

func (n *STRollbackStatementNode) Kind() common.SyntaxKind {
	return common.ROLLBACK_STATEMENT
}

type STRetryStatementNode struct {
	STStatementNode

	RetryKeyword STNode

	TypeParameter STNode

	Arguments STNode

	RetryBody STNode

	OnFailClause STNode
}

func (n *STRetryStatementNode) Kind() common.SyntaxKind {
	return common.RETRY_STATEMENT
}

type STCommitActionNode struct {
	STActionNode

	CommitKeyword STNode
}

func (n *STCommitActionNode) Kind() common.SyntaxKind {
	return common.COMMIT_ACTION
}

type STTransactionalExpressionNode struct {
	STExpressionNode

	TransactionalKeyword STNode
}

func (n *STTransactionalExpressionNode) Kind() common.SyntaxKind {
	return common.TRANSACTIONAL_EXPRESSION
}

type STByteArrayLiteralNode struct {
	STExpressionNode

	Type STNode

	StartBacktick STNode

	Content STNode

	EndBacktick STNode
}

func (n *STByteArrayLiteralNode) Kind() common.SyntaxKind {
	return common.BYTE_ARRAY_LITERAL
}

type STXMLNavigateExpressionNode = STExpressionNode

type STXMLFilterExpressionNode struct {
	STXMLNavigateExpressionNode

	Expression STNode

	XmlPatternChain STNode
}

func (n *STXMLFilterExpressionNode) Kind() common.SyntaxKind {
	return common.XML_FILTER_EXPRESSION
}

type STXMLStepExpressionNode struct {
	STXMLNavigateExpressionNode

	Expression STNode

	XmlStepStart STNode

	XmlStepExtend STNode
}

func (n *STXMLStepExpressionNode) Kind() common.SyntaxKind {
	return common.XML_STEP_EXPRESSION
}

type STXMLNamePatternChainingNode struct {
	STNode

	StartToken STNode

	XmlNamePattern STNode

	GtToken STNode
}

func (n *STXMLNamePatternChainingNode) Kind() common.SyntaxKind {
	return common.XML_NAME_PATTERN_CHAIN
}

type STXMLStepIndexedExtendNode struct {
	STNode

	OpenBracket STNode

	Expression STNode

	CloseBracket STNode
}

func (n *STXMLStepIndexedExtendNode) Kind() common.SyntaxKind {
	return common.XML_STEP_INDEXED_EXTEND
}

type STXMLStepMethodCallExtendNode struct {
	STNode

	DotToken STNode

	MethodName STNode

	ParenthesizedArgList STNode
}

func (n *STXMLStepMethodCallExtendNode) Kind() common.SyntaxKind {
	return common.XML_STEP_METHOD_CALL_EXTEND
}

type STXMLAtomicNamePatternNode struct {
	STNode

	Prefix STNode

	Colon STNode

	Name STNode
}

func (n *STXMLAtomicNamePatternNode) Kind() common.SyntaxKind {
	return common.XML_ATOMIC_NAME_PATTERN
}

type STTypeReferenceTypeDescNode struct {
	STTypeDescriptorNode

	TypeRef STNode
}

func (n *STTypeReferenceTypeDescNode) Kind() common.SyntaxKind {
	return common.TYPE_REFERENCE_TYPE_DESC
}

type STMatchStatementNode struct {
	STStatementNode

	MatchKeyword STNode

	Condition STNode

	OpenBrace STNode

	MatchClauses STNode

	CloseBrace STNode

	OnFailClause STNode
}

func (n *STMatchStatementNode) Kind() common.SyntaxKind {
	return common.MATCH_STATEMENT
}

type STMatchClauseNode struct {
	STNode

	MatchPatterns STNode

	MatchGuard STNode

	RightDoubleArrow STNode

	BlockStatement STNode
}

func (n *STMatchClauseNode) Kind() common.SyntaxKind {
	return common.MATCH_CLAUSE
}

type STMatchGuardNode struct {
	STNode

	IfKeyword STNode

	Expression STNode
}

func (n *STMatchGuardNode) Kind() common.SyntaxKind {
	return common.MATCH_GUARD
}

type STDistinctTypeDescriptorNode struct {
	STTypeDescriptorNode

	DistinctKeyword STNode

	TypeDescriptor STNode
}

func (n *STDistinctTypeDescriptorNode) Kind() common.SyntaxKind {
	return common.DISTINCT_TYPE_DESC
}

type STListMatchPatternNode struct {
	STNode

	OpenBracket STNode

	MatchPatterns STNode

	CloseBracket STNode
}

func (n *STListMatchPatternNode) Kind() common.SyntaxKind {
	return common.LIST_MATCH_PATTERN
}

type STRestMatchPatternNode struct {
	STNode

	EllipsisToken STNode

	VarKeywordToken STNode

	VariableName STNode
}

func (n *STRestMatchPatternNode) Kind() common.SyntaxKind {
	return common.REST_MATCH_PATTERN
}

type STMappingMatchPatternNode struct {
	STNode

	OpenBraceToken STNode

	FieldMatchPatterns STNode

	CloseBraceToken STNode
}

func (n *STMappingMatchPatternNode) Kind() common.SyntaxKind {
	return common.MAPPING_MATCH_PATTERN
}

type STFieldMatchPatternNode struct {
	STNode

	FieldNameNode STNode

	ColonToken STNode

	MatchPattern STNode
}

func (n *STFieldMatchPatternNode) Kind() common.SyntaxKind {
	return common.FIELD_MATCH_PATTERN
}

type STErrorMatchPatternNode struct {
	STNode

	ErrorKeyword STNode

	TypeReference STNode

	OpenParenthesisToken STNode

	ArgListMatchPatternNode STNode

	CloseParenthesisToken STNode
}

func (n *STErrorMatchPatternNode) Kind() common.SyntaxKind {
	return common.ERROR_MATCH_PATTERN
}

type STNamedArgMatchPatternNode struct {
	STNode

	Identifier STNode

	EqualToken STNode

	MatchPattern STNode
}

func (n *STNamedArgMatchPatternNode) Kind() common.SyntaxKind {
	return common.NAMED_ARG_MATCH_PATTERN
}

type STDocumentationNode = STNode

type STMarkdownDocumentationNode struct {
	STDocumentationNode

	DocumentationLines STNode
}

func (n *STMarkdownDocumentationNode) Kind() common.SyntaxKind {
	return common.MARKDOWN_DOCUMENTATION
}

type STMarkdownDocumentationLineNode struct {
	STDocumentationNode

	HashToken STNode

	DocumentElements STNode
}

type STMarkdownParameterDocumentationLineNode struct {
	STDocumentationNode

	HashToken STNode

	PlusToken STNode

	ParameterName STNode

	MinusToken STNode

	DocumentElements STNode
}

type STBallerinaNameReferenceNode struct {
	STDocumentationNode

	ReferenceType STNode

	StartBacktick STNode

	NameReference STNode

	EndBacktick STNode
}

func (n *STBallerinaNameReferenceNode) Kind() common.SyntaxKind {
	return common.BALLERINA_NAME_REFERENCE
}

type STInlineCodeReferenceNode struct {
	STDocumentationNode

	StartBacktick STNode

	CodeReference STNode

	EndBacktick STNode
}

func (n *STInlineCodeReferenceNode) Kind() common.SyntaxKind {
	return common.INLINE_CODE_REFERENCE
}

type STMarkdownCodeBlockNode struct {
	STDocumentationNode

	StartLineHashToken STNode

	StartBacktick STNode

	LangAttribute STNode

	CodeLines STNode

	EndLineHashToken STNode

	EndBacktick STNode
}

func (n *STMarkdownCodeBlockNode) Kind() common.SyntaxKind {
	return common.MARKDOWN_CODE_BLOCK
}

type STMarkdownCodeLineNode struct {
	STDocumentationNode

	HashToken STNode

	CodeDescription STNode
}

func (n *STMarkdownCodeLineNode) Kind() common.SyntaxKind {
	return common.MARKDOWN_CODE_LINE
}

type STOrderByClauseNode struct {
	STIntermediateClauseNode

	OrderKeyword STNode

	ByKeyword STNode

	OrderKey STNode
}

func (n *STOrderByClauseNode) Kind() common.SyntaxKind {
	return common.ORDER_BY_CLAUSE
}

type STOrderKeyNode struct {
	STNode

	Expression STNode

	OrderDirection STNode
}

func (n *STOrderKeyNode) Kind() common.SyntaxKind {
	return common.ORDER_KEY
}

type STGroupByClauseNode struct {
	STIntermediateClauseNode

	GroupKeyword STNode

	ByKeyword STNode

	GroupingKey STNode
}

func (n *STGroupByClauseNode) Kind() common.SyntaxKind {
	return common.GROUP_BY_CLAUSE
}

type STGroupingKeyVarDeclarationNode struct {
	STNode

	TypeDescriptor STNode

	SimpleBindingPattern STNode

	EqualsToken STNode

	Expression STNode
}

func (n *STGroupingKeyVarDeclarationNode) Kind() common.SyntaxKind {
	return common.GROUPING_KEY_VAR_DECLARATION
}

type STOnFailClauseNode struct {
	STClauseNode

	OnKeyword STNode

	FailKeyword STNode

	TypedBindingPattern STNode

	BlockStatement STNode
}

func (n *STOnFailClauseNode) Kind() common.SyntaxKind {
	return common.ON_FAIL_CLAUSE
}

type STDoStatementNode struct {
	STStatementNode

	DoKeyword STNode

	BlockStatement STNode

	OnFailClause STNode
}

func (n *STDoStatementNode) Kind() common.SyntaxKind {
	return common.DO_STATEMENT
}

type STClassDefinitionNode struct {
	STModuleMemberDeclarationNode

	Metadata STNode

	VisibilityQualifier STNode

	ClassTypeQualifiers STNode

	ClassKeyword STNode

	ClassName STNode

	OpenBrace STNode

	Members STNode

	CloseBrace STNode

	SemicolonToken STNode
}

func (n *STClassDefinitionNode) Kind() common.SyntaxKind {
	return common.CLASS_DEFINITION
}

type STResourcePathParameterNode struct {
	STNode

	OpenBracketToken STNode

	Annotations STNode

	TypeDescriptor STNode

	EllipsisToken STNode

	ParamName STNode

	CloseBracketToken STNode
}

type STRequiredExpressionNode struct {
	STExpressionNode

	QuestionMarkToken STNode
}

func (n *STRequiredExpressionNode) Kind() common.SyntaxKind {
	return common.REQUIRED_EXPRESSION
}

type STErrorConstructorExpressionNode struct {
	STExpressionNode

	ErrorKeyword STNode

	TypeReference STNode

	OpenParenToken STNode

	Arguments STNode

	CloseParenToken STNode
}

func (n *STErrorConstructorExpressionNode) Kind() common.SyntaxKind {
	return common.ERROR_CONSTRUCTOR
}

type STParameterizedTypeDescriptorNode struct {
	STTypeDescriptorNode

	KeywordToken STNode

	TypeParamNode STNode
}

type STSpreadMemberNode struct {
	STNode

	Ellipsis STNode

	Expression STNode
}

func (n *STSpreadMemberNode) Kind() common.SyntaxKind {
	return common.SPREAD_MEMBER
}

type STClientResourceAccessActionNode struct {
	STActionNode

	Expression STNode

	RightArrowToken STNode

	SlashToken STNode

	ResourceAccessPath STNode

	DotToken STNode

	MethodName STNode

	Arguments STNode
}

func (n *STClientResourceAccessActionNode) Kind() common.SyntaxKind {
	return common.CLIENT_RESOURCE_ACCESS_ACTION
}

type STComputedResourceAccessSegmentNode struct {
	STNode

	OpenBracketToken STNode

	Expression STNode

	CloseBracketToken STNode
}

func (n *STComputedResourceAccessSegmentNode) Kind() common.SyntaxKind {
	return common.COMPUTED_RESOURCE_ACCESS_SEGMENT
}

type STResourceAccessRestSegmentNode struct {
	STNode

	OpenBracketToken STNode

	EllipsisToken STNode

	Expression STNode

	CloseBracketToken STNode
}

func (n *STResourceAccessRestSegmentNode) Kind() common.SyntaxKind {
	return common.RESOURCE_ACCESS_REST_SEGMENT
}

type STReSequenceNode struct {
	STNode

	ReTerm STNode
}

func (n *STReSequenceNode) Kind() common.SyntaxKind {
	return common.RE_SEQUENCE
}

type STReTermNode = STNode

type STReAtomQuantifierNode struct {
	STReTermNode

	ReAtom STNode

	ReQuantifier STNode
}

func (n *STReAtomQuantifierNode) Kind() common.SyntaxKind {
	return common.RE_ATOM_QUANTIFIER
}

type STReAtomCharOrEscapeNode struct {
	STNode

	ReAtomCharOrEscape STNode
}

func (n *STReAtomCharOrEscapeNode) Kind() common.SyntaxKind {
	return common.RE_LITERAL_CHAR_DOT_OR_ESCAPE
}

type STReQuoteEscapeNode struct {
	STNode

	SlashToken STNode

	ReSyntaxChar STNode
}

func (n *STReQuoteEscapeNode) Kind() common.SyntaxKind {
	return common.RE_QUOTE_ESCAPE
}

type STReSimpleCharClassEscapeNode struct {
	STNode

	SlashToken STNode

	ReSimpleCharClassCode STNode
}

func (n *STReSimpleCharClassEscapeNode) Kind() common.SyntaxKind {
	return common.RE_SIMPLE_CHAR_CLASS_ESCAPE
}

type STReUnicodePropertyEscapeNode struct {
	STNode

	SlashToken STNode

	Property STNode

	OpenBraceToken STNode

	ReUnicodeProperty STNode

	CloseBraceToken STNode
}

func (n *STReUnicodePropertyEscapeNode) Kind() common.SyntaxKind {
	return common.RE_UNICODE_PROPERTY_ESCAPE
}

type STReUnicodePropertyNode = STNode

type STReUnicodeScriptNode struct {
	STReUnicodePropertyNode

	ScriptStart STNode

	ReUnicodePropertyValue STNode
}

func (n *STReUnicodeScriptNode) Kind() common.SyntaxKind {
	return common.RE_UNICODE_SCRIPT
}

type STReUnicodeGeneralCategoryNode struct {
	STReUnicodePropertyNode

	CategoryStart STNode

	ReUnicodeGeneralCategoryName STNode
}

func (n *STReUnicodeGeneralCategoryNode) Kind() common.SyntaxKind {
	return common.RE_UNICODE_GENERAL_CATEGORY
}

type STReCharacterClassNode struct {
	STNode

	OpenBracket STNode

	Negation STNode

	ReCharSet STNode

	CloseBracket STNode
}

func (n *STReCharacterClassNode) Kind() common.SyntaxKind {
	return common.RE_CHARACTER_CLASS
}

type STReCharSetRangeWithReCharSetNode struct {
	STNode

	ReCharSetRange STNode

	ReCharSet STNode
}

func (n *STReCharSetRangeWithReCharSetNode) Kind() common.SyntaxKind {
	return common.RE_CHAR_SET_RANGE_WITH_RE_CHAR_SET
}

type STReCharSetRangeNode struct {
	STNode

	LhsReCharSetAtom STNode

	MinusToken STNode

	RhsReCharSetAtom STNode
}

func (n *STReCharSetRangeNode) Kind() common.SyntaxKind {
	return common.RE_CHAR_SET_RANGE
}

type STReCharSetAtomWithReCharSetNoDashNode struct {
	STNode

	ReCharSetAtom STNode

	ReCharSetNoDash STNode
}

func (n *STReCharSetAtomWithReCharSetNoDashNode) Kind() common.SyntaxKind {
	return common.RE_CHAR_SET_ATOM_WITH_RE_CHAR_SET_NO_DASH
}

type STReCharSetRangeNoDashWithReCharSetNode struct {
	STNode

	ReCharSetRangeNoDash STNode

	ReCharSet STNode
}

func (n *STReCharSetRangeNoDashWithReCharSetNode) Kind() common.SyntaxKind {
	return common.RE_CHAR_SET_RANGE_NO_DASH_WITH_RE_CHAR_SET
}

type STReCharSetRangeNoDashNode struct {
	STNode

	ReCharSetAtomNoDash STNode

	MinusToken STNode

	ReCharSetAtom STNode
}

func (n *STReCharSetRangeNoDashNode) Kind() common.SyntaxKind {
	return common.RE_CHAR_SET_RANGE_NO_DASH
}

type STReCharSetAtomNoDashWithReCharSetNoDashNode struct {
	STNode

	ReCharSetAtomNoDash STNode

	ReCharSetNoDash STNode
}

func (n *STReCharSetAtomNoDashWithReCharSetNoDashNode) Kind() common.SyntaxKind {
	return common.RE_CHAR_SET_ATOM_NO_DASH_WITH_RE_CHAR_SET_NO_DASH
}

type STReCapturingGroupsNode struct {
	STNode

	OpenParenthesis STNode

	ReFlagExpression STNode

	ReSequences STNode

	CloseParenthesis STNode
}

func (n *STReCapturingGroupsNode) Kind() common.SyntaxKind {
	return common.RE_CAPTURING_GROUP
}

type STReFlagExpressionNode struct {
	STNode

	QuestionMark STNode

	ReFlagsOnOff STNode

	Colon STNode
}

func (n *STReFlagExpressionNode) Kind() common.SyntaxKind {
	return common.RE_FLAG_EXPR
}

type STReFlagsOnOffNode struct {
	STNode

	LhsReFlags STNode

	MinusToken STNode

	RhsReFlags STNode
}

func (n *STReFlagsOnOffNode) Kind() common.SyntaxKind {
	return common.RE_FLAGS_ON_OFF
}

type STReFlagsNode struct {
	STNode

	ReFlag STNode
}

func (n *STReFlagsNode) Kind() common.SyntaxKind {
	return common.RE_FLAGS
}

type STReAssertionNode struct {
	STReTermNode

	ReAssertion STNode
}

func (n *STReAssertionNode) Kind() common.SyntaxKind {
	return common.RE_ASSERTION
}

type STReQuantifierNode struct {
	STNode

	ReBaseQuantifier STNode

	NonGreedyChar STNode
}

func (n *STReQuantifierNode) Kind() common.SyntaxKind {
	return common.RE_QUANTIFIER
}

type STReBracedQuantifierNode struct {
	STNode

	OpenBraceToken STNode

	LeastTimesMatchedDigit STNode

	CommaToken STNode

	MostTimesMatchedDigit STNode

	CloseBraceToken STNode
}

func (n *STReBracedQuantifierNode) Kind() common.SyntaxKind {
	return common.RE_BRACED_QUANTIFIER
}

type STMemberTypeDescriptorNode struct {
	STNode

	Annotations STNode

	TypeDescriptor STNode
}

func (n *STMemberTypeDescriptorNode) Kind() common.SyntaxKind {
	return common.MEMBER_TYPE_DESC
}

type STReceiveFieldNode struct {
	STNode

	FieldName STNode

	Colon STNode

	PeerWorker STNode
}

func (n *STReceiveFieldNode) Kind() common.SyntaxKind {
	return common.RECEIVE_FIELD
}

type STNaturalExpressionNode struct {
	STExpressionNode

	ConstKeyword STNode

	NaturalKeyword STNode

	ParenthesizedArgList STNode

	OpenBraceToken STNode

	Prompt STNode

	CloseBraceToken STNode
}

func (n *STNaturalExpressionNode) Kind() common.SyntaxKind {
	return common.NATURAL_EXPRESSION
}

// FIXME:
type STAmbiguousCollectionNode struct {
	STNodeBase
	CollectionStartToken STNode
	CollectionEndToken   STNode
	Members              []STNode
}
