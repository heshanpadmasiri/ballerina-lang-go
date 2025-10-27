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

import "reflect"

// replaceInner recursively replaces a target node with a replacement node.
// Returns true if the replacement was made, false otherwise.
func replaceInner(current STNode, target STNode, replacement STNode) (bool, STNode) {
	if current == nil {
		return false, nil
	}
	if current == target {
		return true, replacement
	}
	switch n := current.(type) {

	case *STModulePart:

		modifiedImports, importsNode := replaceInner(n.Imports, target, replacement)

		modifiedMembers, membersNode := replaceInner(n.Members, target, replacement)

		modifiedEofToken, eofTokenNode := replaceInner(n.EofToken, target, replacement)

		modified := modifiedImports || modifiedMembers || modifiedEofToken
		if modified {
			return true, &STModulePart{

				STNode: n.STNode,

				Imports: importsNode,

				Members: membersNode,

				EofToken: eofTokenNode,
			}
		}
		return false, current

	case *STFunctionDefinition:

		modifiedMetadata, metadataNode := replaceInner(n.Metadata, target, replacement)

		modifiedQualifierList, qualifierListNode := replaceInner(n.QualifierList, target, replacement)

		modifiedFunctionKeyword, functionKeywordNode := replaceInner(n.FunctionKeyword, target, replacement)

		modifiedFunctionName, functionNameNode := replaceInner(n.FunctionName, target, replacement)

		modifiedRelativeResourcePath, relativeResourcePathNode := replaceInner(n.RelativeResourcePath, target, replacement)

		modifiedFunctionSignature, functionSignatureNode := replaceInner(n.FunctionSignature, target, replacement)

		modifiedFunctionBody, functionBodyNode := replaceInner(n.FunctionBody, target, replacement)

		modified := modifiedMetadata || modifiedQualifierList || modifiedFunctionKeyword || modifiedFunctionName || modifiedRelativeResourcePath || modifiedFunctionSignature || modifiedFunctionBody
		if modified {
			return true, &STFunctionDefinition{

				STModuleMemberDeclarationNode: n.STModuleMemberDeclarationNode,

				Metadata: metadataNode,

				QualifierList: qualifierListNode,

				FunctionKeyword: functionKeywordNode,

				FunctionName: functionNameNode,

				RelativeResourcePath: relativeResourcePathNode,

				FunctionSignature: functionSignatureNode,

				FunctionBody: functionBodyNode,
			}
		}
		return false, current

	case *STImportDeclarationNode:

		modifiedImportKeyword, importKeywordNode := replaceInner(n.ImportKeyword, target, replacement)

		modifiedOrgName, orgNameNode := replaceInner(n.OrgName, target, replacement)

		modifiedModuleName, moduleNameNode := replaceInner(n.ModuleName, target, replacement)

		modifiedPrefix, prefixNode := replaceInner(n.Prefix, target, replacement)

		modifiedSemicolon, semicolonNode := replaceInner(n.Semicolon, target, replacement)

		modified := modifiedImportKeyword || modifiedOrgName || modifiedModuleName || modifiedPrefix || modifiedSemicolon
		if modified {
			return true, &STImportDeclarationNode{

				STNode: n.STNode,

				ImportKeyword: importKeywordNode,

				OrgName: orgNameNode,

				ModuleName: moduleNameNode,

				Prefix: prefixNode,

				Semicolon: semicolonNode,
			}
		}
		return false, current

	case *STListenerDeclarationNode:

		modifiedMetadata, metadataNode := replaceInner(n.Metadata, target, replacement)

		modifiedVisibilityQualifier, visibilityQualifierNode := replaceInner(n.VisibilityQualifier, target, replacement)

		modifiedListenerKeyword, listenerKeywordNode := replaceInner(n.ListenerKeyword, target, replacement)

		modifiedTypeDescriptor, typeDescriptorNode := replaceInner(n.TypeDescriptor, target, replacement)

		modifiedVariableName, variableNameNode := replaceInner(n.VariableName, target, replacement)

		modifiedEqualsToken, equalsTokenNode := replaceInner(n.EqualsToken, target, replacement)

		modifiedInitializer, initializerNode := replaceInner(n.Initializer, target, replacement)

		modifiedSemicolonToken, semicolonTokenNode := replaceInner(n.SemicolonToken, target, replacement)

		modified := modifiedMetadata || modifiedVisibilityQualifier || modifiedListenerKeyword || modifiedTypeDescriptor || modifiedVariableName || modifiedEqualsToken || modifiedInitializer || modifiedSemicolonToken
		if modified {
			return true, &STListenerDeclarationNode{

				STModuleMemberDeclarationNode: n.STModuleMemberDeclarationNode,

				Metadata: metadataNode,

				VisibilityQualifier: visibilityQualifierNode,

				ListenerKeyword: listenerKeywordNode,

				TypeDescriptor: typeDescriptorNode,

				VariableName: variableNameNode,

				EqualsToken: equalsTokenNode,

				Initializer: initializerNode,

				SemicolonToken: semicolonTokenNode,
			}
		}
		return false, current

	case *STTypeDefinitionNode:

		modifiedMetadata, metadataNode := replaceInner(n.Metadata, target, replacement)

		modifiedVisibilityQualifier, visibilityQualifierNode := replaceInner(n.VisibilityQualifier, target, replacement)

		modifiedTypeKeyword, typeKeywordNode := replaceInner(n.TypeKeyword, target, replacement)

		modifiedTypeName, typeNameNode := replaceInner(n.TypeName, target, replacement)

		modifiedTypeDescriptor, typeDescriptorNode := replaceInner(n.TypeDescriptor, target, replacement)

		modifiedSemicolonToken, semicolonTokenNode := replaceInner(n.SemicolonToken, target, replacement)

		modified := modifiedMetadata || modifiedVisibilityQualifier || modifiedTypeKeyword || modifiedTypeName || modifiedTypeDescriptor || modifiedSemicolonToken
		if modified {
			return true, &STTypeDefinitionNode{

				STModuleMemberDeclarationNode: n.STModuleMemberDeclarationNode,

				Metadata: metadataNode,

				VisibilityQualifier: visibilityQualifierNode,

				TypeKeyword: typeKeywordNode,

				TypeName: typeNameNode,

				TypeDescriptor: typeDescriptorNode,

				SemicolonToken: semicolonTokenNode,
			}
		}
		return false, current

	case *STServiceDeclarationNode:

		modifiedMetadata, metadataNode := replaceInner(n.Metadata, target, replacement)

		modifiedQualifiers, qualifiersNode := replaceInner(n.Qualifiers, target, replacement)

		modifiedServiceKeyword, serviceKeywordNode := replaceInner(n.ServiceKeyword, target, replacement)

		modifiedTypeDescriptor, typeDescriptorNode := replaceInner(n.TypeDescriptor, target, replacement)

		modifiedAbsoluteResourcePath, absoluteResourcePathNode := replaceInner(n.AbsoluteResourcePath, target, replacement)

		modifiedOnKeyword, onKeywordNode := replaceInner(n.OnKeyword, target, replacement)

		modifiedExpressions, expressionsNode := replaceInner(n.Expressions, target, replacement)

		modifiedOpenBraceToken, openBraceTokenNode := replaceInner(n.OpenBraceToken, target, replacement)

		modifiedMembers, membersNode := replaceInner(n.Members, target, replacement)

		modifiedCloseBraceToken, closeBraceTokenNode := replaceInner(n.CloseBraceToken, target, replacement)

		modifiedSemicolonToken, semicolonTokenNode := replaceInner(n.SemicolonToken, target, replacement)

		modified := modifiedMetadata || modifiedQualifiers || modifiedServiceKeyword || modifiedTypeDescriptor || modifiedAbsoluteResourcePath || modifiedOnKeyword || modifiedExpressions || modifiedOpenBraceToken || modifiedMembers || modifiedCloseBraceToken || modifiedSemicolonToken
		if modified {
			return true, &STServiceDeclarationNode{

				STModuleMemberDeclarationNode: n.STModuleMemberDeclarationNode,

				Metadata: metadataNode,

				Qualifiers: qualifiersNode,

				ServiceKeyword: serviceKeywordNode,

				TypeDescriptor: typeDescriptorNode,

				AbsoluteResourcePath: absoluteResourcePathNode,

				OnKeyword: onKeywordNode,

				Expressions: expressionsNode,

				OpenBraceToken: openBraceTokenNode,

				Members: membersNode,

				CloseBraceToken: closeBraceTokenNode,

				SemicolonToken: semicolonTokenNode,
			}
		}
		return false, current

	case *STAssignmentStatementNode:

		modifiedVarRef, varRefNode := replaceInner(n.VarRef, target, replacement)

		modifiedEqualsToken, equalsTokenNode := replaceInner(n.EqualsToken, target, replacement)

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modifiedSemicolonToken, semicolonTokenNode := replaceInner(n.SemicolonToken, target, replacement)

		modified := modifiedVarRef || modifiedEqualsToken || modifiedExpression || modifiedSemicolonToken
		if modified {
			return true, &STAssignmentStatementNode{

				STStatementNode: n.STStatementNode,

				VarRef: varRefNode,

				EqualsToken: equalsTokenNode,

				Expression: expressionNode,

				SemicolonToken: semicolonTokenNode,
			}
		}
		return false, current

	case *STCompoundAssignmentStatementNode:

		modifiedLhsExpression, lhsExpressionNode := replaceInner(n.LhsExpression, target, replacement)

		modifiedBinaryOperator, binaryOperatorNode := replaceInner(n.BinaryOperator, target, replacement)

		modifiedEqualsToken, equalsTokenNode := replaceInner(n.EqualsToken, target, replacement)

		modifiedRhsExpression, rhsExpressionNode := replaceInner(n.RhsExpression, target, replacement)

		modifiedSemicolonToken, semicolonTokenNode := replaceInner(n.SemicolonToken, target, replacement)

		modified := modifiedLhsExpression || modifiedBinaryOperator || modifiedEqualsToken || modifiedRhsExpression || modifiedSemicolonToken
		if modified {
			return true, &STCompoundAssignmentStatementNode{

				STStatementNode: n.STStatementNode,

				LhsExpression: lhsExpressionNode,

				BinaryOperator: binaryOperatorNode,

				EqualsToken: equalsTokenNode,

				RhsExpression: rhsExpressionNode,

				SemicolonToken: semicolonTokenNode,
			}
		}
		return false, current

	case *STVariableDeclarationNode:

		modifiedAnnotations, annotationsNode := replaceInner(n.Annotations, target, replacement)

		modifiedFinalKeyword, finalKeywordNode := replaceInner(n.FinalKeyword, target, replacement)

		modifiedTypedBindingPattern, typedBindingPatternNode := replaceInner(n.TypedBindingPattern, target, replacement)

		modifiedEqualsToken, equalsTokenNode := replaceInner(n.EqualsToken, target, replacement)

		modifiedInitializer, initializerNode := replaceInner(n.Initializer, target, replacement)

		modifiedSemicolonToken, semicolonTokenNode := replaceInner(n.SemicolonToken, target, replacement)

		modified := modifiedAnnotations || modifiedFinalKeyword || modifiedTypedBindingPattern || modifiedEqualsToken || modifiedInitializer || modifiedSemicolonToken
		if modified {
			return true, &STVariableDeclarationNode{

				STStatementNode: n.STStatementNode,

				Annotations: annotationsNode,

				FinalKeyword: finalKeywordNode,

				TypedBindingPattern: typedBindingPatternNode,

				EqualsToken: equalsTokenNode,

				Initializer: initializerNode,

				SemicolonToken: semicolonTokenNode,
			}
		}
		return false, current

	case *STBlockStatementNode:

		modifiedOpenBraceToken, openBraceTokenNode := replaceInner(n.OpenBraceToken, target, replacement)

		modifiedStatements, statementsNode := replaceInner(n.Statements, target, replacement)

		modifiedCloseBraceToken, closeBraceTokenNode := replaceInner(n.CloseBraceToken, target, replacement)

		modified := modifiedOpenBraceToken || modifiedStatements || modifiedCloseBraceToken
		if modified {
			return true, &STBlockStatementNode{

				STStatementNode: n.STStatementNode,

				OpenBraceToken: openBraceTokenNode,

				Statements: statementsNode,

				CloseBraceToken: closeBraceTokenNode,
			}
		}
		return false, current

	case *STBreakStatementNode:

		modifiedBreakToken, breakTokenNode := replaceInner(n.BreakToken, target, replacement)

		modifiedSemicolonToken, semicolonTokenNode := replaceInner(n.SemicolonToken, target, replacement)

		modified := modifiedBreakToken || modifiedSemicolonToken
		if modified {
			return true, &STBreakStatementNode{

				STStatementNode: n.STStatementNode,

				BreakToken: breakTokenNode,

				SemicolonToken: semicolonTokenNode,
			}
		}
		return false, current

	case *STFailStatementNode:

		modifiedFailKeyword, failKeywordNode := replaceInner(n.FailKeyword, target, replacement)

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modifiedSemicolonToken, semicolonTokenNode := replaceInner(n.SemicolonToken, target, replacement)

		modified := modifiedFailKeyword || modifiedExpression || modifiedSemicolonToken
		if modified {
			return true, &STFailStatementNode{

				STStatementNode: n.STStatementNode,

				FailKeyword: failKeywordNode,

				Expression: expressionNode,

				SemicolonToken: semicolonTokenNode,
			}
		}
		return false, current

	case *STExpressionStatementNode:

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modifiedSemicolonToken, semicolonTokenNode := replaceInner(n.SemicolonToken, target, replacement)

		modified := modifiedExpression || modifiedSemicolonToken
		if modified {
			return true, &STExpressionStatementNode{

				STStatementNode: n.STStatementNode,

				Expression: expressionNode,

				SemicolonToken: semicolonTokenNode,
			}
		}
		return false, current

	case *STContinueStatementNode:

		modifiedContinueToken, continueTokenNode := replaceInner(n.ContinueToken, target, replacement)

		modifiedSemicolonToken, semicolonTokenNode := replaceInner(n.SemicolonToken, target, replacement)

		modified := modifiedContinueToken || modifiedSemicolonToken
		if modified {
			return true, &STContinueStatementNode{

				STStatementNode: n.STStatementNode,

				ContinueToken: continueTokenNode,

				SemicolonToken: semicolonTokenNode,
			}
		}
		return false, current

	case *STExternalFunctionBodyNode:

		modifiedEqualsToken, equalsTokenNode := replaceInner(n.EqualsToken, target, replacement)

		modifiedAnnotations, annotationsNode := replaceInner(n.Annotations, target, replacement)

		modifiedExternalKeyword, externalKeywordNode := replaceInner(n.ExternalKeyword, target, replacement)

		modifiedSemicolonToken, semicolonTokenNode := replaceInner(n.SemicolonToken, target, replacement)

		modified := modifiedEqualsToken || modifiedAnnotations || modifiedExternalKeyword || modifiedSemicolonToken
		if modified {
			return true, &STExternalFunctionBodyNode{

				STFunctionBodyNode: n.STFunctionBodyNode,

				EqualsToken: equalsTokenNode,

				Annotations: annotationsNode,

				ExternalKeyword: externalKeywordNode,

				SemicolonToken: semicolonTokenNode,
			}
		}
		return false, current

	case *STIfElseStatementNode:

		modifiedIfKeyword, ifKeywordNode := replaceInner(n.IfKeyword, target, replacement)

		modifiedCondition, conditionNode := replaceInner(n.Condition, target, replacement)

		modifiedIfBody, ifBodyNode := replaceInner(n.IfBody, target, replacement)

		modifiedElseBody, elseBodyNode := replaceInner(n.ElseBody, target, replacement)

		modified := modifiedIfKeyword || modifiedCondition || modifiedIfBody || modifiedElseBody
		if modified {
			return true, &STIfElseStatementNode{

				STStatementNode: n.STStatementNode,

				IfKeyword: ifKeywordNode,

				Condition: conditionNode,

				IfBody: ifBodyNode,

				ElseBody: elseBodyNode,
			}
		}
		return false, current

	case *STElseBlockNode:

		modifiedElseKeyword, elseKeywordNode := replaceInner(n.ElseKeyword, target, replacement)

		modifiedElseBody, elseBodyNode := replaceInner(n.ElseBody, target, replacement)

		modified := modifiedElseKeyword || modifiedElseBody
		if modified {
			return true, &STElseBlockNode{

				STNode: n.STNode,

				ElseKeyword: elseKeywordNode,

				ElseBody: elseBodyNode,
			}
		}
		return false, current

	case *STWhileStatementNode:

		modifiedWhileKeyword, whileKeywordNode := replaceInner(n.WhileKeyword, target, replacement)

		modifiedCondition, conditionNode := replaceInner(n.Condition, target, replacement)

		modifiedWhileBody, whileBodyNode := replaceInner(n.WhileBody, target, replacement)

		modifiedOnFailClause, onFailClauseNode := replaceInner(n.OnFailClause, target, replacement)

		modified := modifiedWhileKeyword || modifiedCondition || modifiedWhileBody || modifiedOnFailClause
		if modified {
			return true, &STWhileStatementNode{

				STStatementNode: n.STStatementNode,

				WhileKeyword: whileKeywordNode,

				Condition: conditionNode,

				WhileBody: whileBodyNode,

				OnFailClause: onFailClauseNode,
			}
		}
		return false, current

	case *STPanicStatementNode:

		modifiedPanicKeyword, panicKeywordNode := replaceInner(n.PanicKeyword, target, replacement)

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modifiedSemicolonToken, semicolonTokenNode := replaceInner(n.SemicolonToken, target, replacement)

		modified := modifiedPanicKeyword || modifiedExpression || modifiedSemicolonToken
		if modified {
			return true, &STPanicStatementNode{

				STStatementNode: n.STStatementNode,

				PanicKeyword: panicKeywordNode,

				Expression: expressionNode,

				SemicolonToken: semicolonTokenNode,
			}
		}
		return false, current

	case *STReturnStatementNode:

		modifiedReturnKeyword, returnKeywordNode := replaceInner(n.ReturnKeyword, target, replacement)

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modifiedSemicolonToken, semicolonTokenNode := replaceInner(n.SemicolonToken, target, replacement)

		modified := modifiedReturnKeyword || modifiedExpression || modifiedSemicolonToken
		if modified {
			return true, &STReturnStatementNode{

				STStatementNode: n.STStatementNode,

				ReturnKeyword: returnKeywordNode,

				Expression: expressionNode,

				SemicolonToken: semicolonTokenNode,
			}
		}
		return false, current

	case *STLocalTypeDefinitionStatementNode:

		modifiedAnnotations, annotationsNode := replaceInner(n.Annotations, target, replacement)

		modifiedTypeKeyword, typeKeywordNode := replaceInner(n.TypeKeyword, target, replacement)

		modifiedTypeName, typeNameNode := replaceInner(n.TypeName, target, replacement)

		modifiedTypeDescriptor, typeDescriptorNode := replaceInner(n.TypeDescriptor, target, replacement)

		modifiedSemicolonToken, semicolonTokenNode := replaceInner(n.SemicolonToken, target, replacement)

		modified := modifiedAnnotations || modifiedTypeKeyword || modifiedTypeName || modifiedTypeDescriptor || modifiedSemicolonToken
		if modified {
			return true, &STLocalTypeDefinitionStatementNode{

				STStatementNode: n.STStatementNode,

				Annotations: annotationsNode,

				TypeKeyword: typeKeywordNode,

				TypeName: typeNameNode,

				TypeDescriptor: typeDescriptorNode,

				SemicolonToken: semicolonTokenNode,
			}
		}
		return false, current

	case *STLockStatementNode:

		modifiedLockKeyword, lockKeywordNode := replaceInner(n.LockKeyword, target, replacement)

		modifiedBlockStatement, blockStatementNode := replaceInner(n.BlockStatement, target, replacement)

		modifiedOnFailClause, onFailClauseNode := replaceInner(n.OnFailClause, target, replacement)

		modified := modifiedLockKeyword || modifiedBlockStatement || modifiedOnFailClause
		if modified {
			return true, &STLockStatementNode{

				STStatementNode: n.STStatementNode,

				LockKeyword: lockKeywordNode,

				BlockStatement: blockStatementNode,

				OnFailClause: onFailClauseNode,
			}
		}
		return false, current

	case *STForkStatementNode:

		modifiedForkKeyword, forkKeywordNode := replaceInner(n.ForkKeyword, target, replacement)

		modifiedOpenBraceToken, openBraceTokenNode := replaceInner(n.OpenBraceToken, target, replacement)

		modifiedNamedWorkerDeclarations, namedWorkerDeclarationsNode := replaceInner(n.NamedWorkerDeclarations, target, replacement)

		modifiedCloseBraceToken, closeBraceTokenNode := replaceInner(n.CloseBraceToken, target, replacement)

		modified := modifiedForkKeyword || modifiedOpenBraceToken || modifiedNamedWorkerDeclarations || modifiedCloseBraceToken
		if modified {
			return true, &STForkStatementNode{

				STStatementNode: n.STStatementNode,

				ForkKeyword: forkKeywordNode,

				OpenBraceToken: openBraceTokenNode,

				NamedWorkerDeclarations: namedWorkerDeclarationsNode,

				CloseBraceToken: closeBraceTokenNode,
			}
		}
		return false, current

	case *STForEachStatementNode:

		modifiedForEachKeyword, forEachKeywordNode := replaceInner(n.ForEachKeyword, target, replacement)

		modifiedTypedBindingPattern, typedBindingPatternNode := replaceInner(n.TypedBindingPattern, target, replacement)

		modifiedInKeyword, inKeywordNode := replaceInner(n.InKeyword, target, replacement)

		modifiedActionOrExpressionNode, actionOrExpressionNodeNode := replaceInner(n.ActionOrExpressionNode, target, replacement)

		modifiedBlockStatement, blockStatementNode := replaceInner(n.BlockStatement, target, replacement)

		modifiedOnFailClause, onFailClauseNode := replaceInner(n.OnFailClause, target, replacement)

		modified := modifiedForEachKeyword || modifiedTypedBindingPattern || modifiedInKeyword || modifiedActionOrExpressionNode || modifiedBlockStatement || modifiedOnFailClause
		if modified {
			return true, &STForEachStatementNode{

				STStatementNode: n.STStatementNode,

				ForEachKeyword: forEachKeywordNode,

				TypedBindingPattern: typedBindingPatternNode,

				InKeyword: inKeywordNode,

				ActionOrExpressionNode: actionOrExpressionNodeNode,

				BlockStatement: blockStatementNode,

				OnFailClause: onFailClauseNode,
			}
		}
		return false, current

	case *STBinaryExpressionNode:

		modifiedLhsExpr, lhsExprNode := replaceInner(n.LhsExpr, target, replacement)

		modifiedOperator, operatorNode := replaceInner(n.Operator, target, replacement)

		modifiedRhsExpr, rhsExprNode := replaceInner(n.RhsExpr, target, replacement)

		modified := modifiedLhsExpr || modifiedOperator || modifiedRhsExpr
		if modified {
			return true, &STBinaryExpressionNode{

				STExpressionNode: n.STExpressionNode,

				LhsExpr: lhsExprNode,

				Operator: operatorNode,

				RhsExpr: rhsExprNode,
			}
		}
		return false, current

	case *STBracedExpressionNode:

		modifiedOpenParen, openParenNode := replaceInner(n.OpenParen, target, replacement)

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modifiedCloseParen, closeParenNode := replaceInner(n.CloseParen, target, replacement)

		modified := modifiedOpenParen || modifiedExpression || modifiedCloseParen
		if modified {
			return true, &STBracedExpressionNode{

				STExpressionNode: n.STExpressionNode,

				OpenParen: openParenNode,

				Expression: expressionNode,

				CloseParen: closeParenNode,
			}
		}
		return false, current

	case *STCheckExpressionNode:

		modifiedCheckKeyword, checkKeywordNode := replaceInner(n.CheckKeyword, target, replacement)

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modified := modifiedCheckKeyword || modifiedExpression
		if modified {
			return true, &STCheckExpressionNode{

				STExpressionNode: n.STExpressionNode,

				CheckKeyword: checkKeywordNode,

				Expression: expressionNode,
			}
		}
		return false, current

	case *STFieldAccessExpressionNode:

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modifiedDotToken, dotTokenNode := replaceInner(n.DotToken, target, replacement)

		modifiedFieldName, fieldNameNode := replaceInner(n.FieldName, target, replacement)

		modified := modifiedExpression || modifiedDotToken || modifiedFieldName
		if modified {
			return true, &STFieldAccessExpressionNode{

				STExpressionNode: n.STExpressionNode,

				Expression: expressionNode,

				DotToken: dotTokenNode,

				FieldName: fieldNameNode,
			}
		}
		return false, current

	case *STFunctionCallExpressionNode:

		modifiedFunctionName, functionNameNode := replaceInner(n.FunctionName, target, replacement)

		modifiedOpenParenToken, openParenTokenNode := replaceInner(n.OpenParenToken, target, replacement)

		modifiedArguments, argumentsNode := replaceInner(n.Arguments, target, replacement)

		modifiedCloseParenToken, closeParenTokenNode := replaceInner(n.CloseParenToken, target, replacement)

		modified := modifiedFunctionName || modifiedOpenParenToken || modifiedArguments || modifiedCloseParenToken
		if modified {
			return true, &STFunctionCallExpressionNode{

				STExpressionNode: n.STExpressionNode,

				FunctionName: functionNameNode,

				OpenParenToken: openParenTokenNode,

				Arguments: argumentsNode,

				CloseParenToken: closeParenTokenNode,
			}
		}
		return false, current

	case *STMethodCallExpressionNode:

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modifiedDotToken, dotTokenNode := replaceInner(n.DotToken, target, replacement)

		modifiedMethodName, methodNameNode := replaceInner(n.MethodName, target, replacement)

		modifiedOpenParenToken, openParenTokenNode := replaceInner(n.OpenParenToken, target, replacement)

		modifiedArguments, argumentsNode := replaceInner(n.Arguments, target, replacement)

		modifiedCloseParenToken, closeParenTokenNode := replaceInner(n.CloseParenToken, target, replacement)

		modified := modifiedExpression || modifiedDotToken || modifiedMethodName || modifiedOpenParenToken || modifiedArguments || modifiedCloseParenToken
		if modified {
			return true, &STMethodCallExpressionNode{

				STExpressionNode: n.STExpressionNode,

				Expression: expressionNode,

				DotToken: dotTokenNode,

				MethodName: methodNameNode,

				OpenParenToken: openParenTokenNode,

				Arguments: argumentsNode,

				CloseParenToken: closeParenTokenNode,
			}
		}
		return false, current

	case *STMappingConstructorExpressionNode:

		modifiedOpenBrace, openBraceNode := replaceInner(n.OpenBrace, target, replacement)

		modifiedFields, fieldsNode := replaceInner(n.Fields, target, replacement)

		modifiedCloseBrace, closeBraceNode := replaceInner(n.CloseBrace, target, replacement)

		modified := modifiedOpenBrace || modifiedFields || modifiedCloseBrace
		if modified {
			return true, &STMappingConstructorExpressionNode{

				STExpressionNode: n.STExpressionNode,

				OpenBrace: openBraceNode,

				Fields: fieldsNode,

				CloseBrace: closeBraceNode,
			}
		}
		return false, current

	case *STIndexedExpressionNode:

		modifiedContainerExpression, containerExpressionNode := replaceInner(n.ContainerExpression, target, replacement)

		modifiedOpenBracket, openBracketNode := replaceInner(n.OpenBracket, target, replacement)

		modifiedKeyExpression, keyExpressionNode := replaceInner(n.KeyExpression, target, replacement)

		modifiedCloseBracket, closeBracketNode := replaceInner(n.CloseBracket, target, replacement)

		modified := modifiedContainerExpression || modifiedOpenBracket || modifiedKeyExpression || modifiedCloseBracket
		if modified {
			return true, &STIndexedExpressionNode{

				STTypeDescriptorNode: n.STTypeDescriptorNode,

				ContainerExpression: containerExpressionNode,

				OpenBracket: openBracketNode,

				KeyExpression: keyExpressionNode,

				CloseBracket: closeBracketNode,
			}
		}
		return false, current

	case *STTypeofExpressionNode:

		modifiedTypeofKeyword, typeofKeywordNode := replaceInner(n.TypeofKeyword, target, replacement)

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modified := modifiedTypeofKeyword || modifiedExpression
		if modified {
			return true, &STTypeofExpressionNode{

				STExpressionNode: n.STExpressionNode,

				TypeofKeyword: typeofKeywordNode,

				Expression: expressionNode,
			}
		}
		return false, current

	case *STUnaryExpressionNode:

		modifiedUnaryOperator, unaryOperatorNode := replaceInner(n.UnaryOperator, target, replacement)

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modified := modifiedUnaryOperator || modifiedExpression
		if modified {
			return true, &STUnaryExpressionNode{

				STExpressionNode: n.STExpressionNode,

				UnaryOperator: unaryOperatorNode,

				Expression: expressionNode,
			}
		}
		return false, current

	case *STComputedNameFieldNode:

		modifiedOpenBracket, openBracketNode := replaceInner(n.OpenBracket, target, replacement)

		modifiedFieldNameExpr, fieldNameExprNode := replaceInner(n.FieldNameExpr, target, replacement)

		modifiedCloseBracket, closeBracketNode := replaceInner(n.CloseBracket, target, replacement)

		modifiedColonToken, colonTokenNode := replaceInner(n.ColonToken, target, replacement)

		modifiedValueExpr, valueExprNode := replaceInner(n.ValueExpr, target, replacement)

		modified := modifiedOpenBracket || modifiedFieldNameExpr || modifiedCloseBracket || modifiedColonToken || modifiedValueExpr
		if modified {
			return true, &STComputedNameFieldNode{

				STMappingFieldNode: n.STMappingFieldNode,

				OpenBracket: openBracketNode,

				FieldNameExpr: fieldNameExprNode,

				CloseBracket: closeBracketNode,

				ColonToken: colonTokenNode,

				ValueExpr: valueExprNode,
			}
		}
		return false, current

	case *STConstantDeclarationNode:

		modifiedMetadata, metadataNode := replaceInner(n.Metadata, target, replacement)

		modifiedVisibilityQualifier, visibilityQualifierNode := replaceInner(n.VisibilityQualifier, target, replacement)

		modifiedConstKeyword, constKeywordNode := replaceInner(n.ConstKeyword, target, replacement)

		modifiedTypeDescriptor, typeDescriptorNode := replaceInner(n.TypeDescriptor, target, replacement)

		modifiedVariableName, variableNameNode := replaceInner(n.VariableName, target, replacement)

		modifiedEqualsToken, equalsTokenNode := replaceInner(n.EqualsToken, target, replacement)

		modifiedInitializer, initializerNode := replaceInner(n.Initializer, target, replacement)

		modifiedSemicolonToken, semicolonTokenNode := replaceInner(n.SemicolonToken, target, replacement)

		modified := modifiedMetadata || modifiedVisibilityQualifier || modifiedConstKeyword || modifiedTypeDescriptor || modifiedVariableName || modifiedEqualsToken || modifiedInitializer || modifiedSemicolonToken
		if modified {
			return true, &STConstantDeclarationNode{

				STModuleMemberDeclarationNode: n.STModuleMemberDeclarationNode,

				Metadata: metadataNode,

				VisibilityQualifier: visibilityQualifierNode,

				ConstKeyword: constKeywordNode,

				TypeDescriptor: typeDescriptorNode,

				VariableName: variableNameNode,

				EqualsToken: equalsTokenNode,

				Initializer: initializerNode,

				SemicolonToken: semicolonTokenNode,
			}
		}
		return false, current

	case *STDefaultableParameterNode:

		modifiedAnnotations, annotationsNode := replaceInner(n.Annotations, target, replacement)

		modifiedTypeName, typeNameNode := replaceInner(n.TypeName, target, replacement)

		modifiedParamName, paramNameNode := replaceInner(n.ParamName, target, replacement)

		modifiedEqualsToken, equalsTokenNode := replaceInner(n.EqualsToken, target, replacement)

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modified := modifiedAnnotations || modifiedTypeName || modifiedParamName || modifiedEqualsToken || modifiedExpression
		if modified {
			return true, &STDefaultableParameterNode{

				STParameterNode: n.STParameterNode,

				Annotations: annotationsNode,

				TypeName: typeNameNode,

				ParamName: paramNameNode,

				EqualsToken: equalsTokenNode,

				Expression: expressionNode,
			}
		}
		return false, current

	case *STRequiredParameterNode:

		modifiedAnnotations, annotationsNode := replaceInner(n.Annotations, target, replacement)

		modifiedTypeName, typeNameNode := replaceInner(n.TypeName, target, replacement)

		modifiedParamName, paramNameNode := replaceInner(n.ParamName, target, replacement)

		modified := modifiedAnnotations || modifiedTypeName || modifiedParamName
		if modified {
			return true, &STRequiredParameterNode{

				STParameterNode: n.STParameterNode,

				Annotations: annotationsNode,

				TypeName: typeNameNode,

				ParamName: paramNameNode,
			}
		}
		return false, current

	case *STIncludedRecordParameterNode:

		modifiedAnnotations, annotationsNode := replaceInner(n.Annotations, target, replacement)

		modifiedAsteriskToken, asteriskTokenNode := replaceInner(n.AsteriskToken, target, replacement)

		modifiedTypeName, typeNameNode := replaceInner(n.TypeName, target, replacement)

		modifiedParamName, paramNameNode := replaceInner(n.ParamName, target, replacement)

		modified := modifiedAnnotations || modifiedAsteriskToken || modifiedTypeName || modifiedParamName
		if modified {
			return true, &STIncludedRecordParameterNode{

				STParameterNode: n.STParameterNode,

				Annotations: annotationsNode,

				AsteriskToken: asteriskTokenNode,

				TypeName: typeNameNode,

				ParamName: paramNameNode,
			}
		}
		return false, current

	case *STRestParameterNode:

		modifiedAnnotations, annotationsNode := replaceInner(n.Annotations, target, replacement)

		modifiedTypeName, typeNameNode := replaceInner(n.TypeName, target, replacement)

		modifiedEllipsisToken, ellipsisTokenNode := replaceInner(n.EllipsisToken, target, replacement)

		modifiedParamName, paramNameNode := replaceInner(n.ParamName, target, replacement)

		modified := modifiedAnnotations || modifiedTypeName || modifiedEllipsisToken || modifiedParamName
		if modified {
			return true, &STRestParameterNode{

				STParameterNode: n.STParameterNode,

				Annotations: annotationsNode,

				TypeName: typeNameNode,

				EllipsisToken: ellipsisTokenNode,

				ParamName: paramNameNode,
			}
		}
		return false, current

	case *STImportOrgNameNode:

		modifiedOrgName, orgNameNode := replaceInner(n.OrgName, target, replacement)

		modifiedSlashToken, slashTokenNode := replaceInner(n.SlashToken, target, replacement)

		modified := modifiedOrgName || modifiedSlashToken
		if modified {
			return true, &STImportOrgNameNode{

				STNode: n.STNode,

				OrgName: orgNameNode,

				SlashToken: slashTokenNode,
			}
		}
		return false, current

	case *STImportPrefixNode:

		modifiedAsKeyword, asKeywordNode := replaceInner(n.AsKeyword, target, replacement)

		modifiedPrefix, prefixNode := replaceInner(n.Prefix, target, replacement)

		modified := modifiedAsKeyword || modifiedPrefix
		if modified {
			return true, &STImportPrefixNode{

				STNode: n.STNode,

				AsKeyword: asKeywordNode,

				Prefix: prefixNode,
			}
		}
		return false, current

	case *STSpecificFieldNode:

		modifiedReadonlyKeyword, readonlyKeywordNode := replaceInner(n.ReadonlyKeyword, target, replacement)

		modifiedFieldName, fieldNameNode := replaceInner(n.FieldName, target, replacement)

		modifiedColon, colonNode := replaceInner(n.Colon, target, replacement)

		modifiedValueExpr, valueExprNode := replaceInner(n.ValueExpr, target, replacement)

		modified := modifiedReadonlyKeyword || modifiedFieldName || modifiedColon || modifiedValueExpr
		if modified {
			return true, &STSpecificFieldNode{

				STMappingFieldNode: n.STMappingFieldNode,

				ReadonlyKeyword: readonlyKeywordNode,

				FieldName: fieldNameNode,

				Colon: colonNode,

				ValueExpr: valueExprNode,
			}
		}
		return false, current

	case *STSpreadFieldNode:

		modifiedEllipsis, ellipsisNode := replaceInner(n.Ellipsis, target, replacement)

		modifiedValueExpr, valueExprNode := replaceInner(n.ValueExpr, target, replacement)

		modified := modifiedEllipsis || modifiedValueExpr
		if modified {
			return true, &STSpreadFieldNode{

				STMappingFieldNode: n.STMappingFieldNode,

				Ellipsis: ellipsisNode,

				ValueExpr: valueExprNode,
			}
		}
		return false, current

	case *STNamedArgumentNode:

		modifiedArgumentName, argumentNameNode := replaceInner(n.ArgumentName, target, replacement)

		modifiedEqualsToken, equalsTokenNode := replaceInner(n.EqualsToken, target, replacement)

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modified := modifiedArgumentName || modifiedEqualsToken || modifiedExpression
		if modified {
			return true, &STNamedArgumentNode{

				STFunctionArgumentNode: n.STFunctionArgumentNode,

				ArgumentName: argumentNameNode,

				EqualsToken: equalsTokenNode,

				Expression: expressionNode,
			}
		}
		return false, current

	case *STPositionalArgumentNode:

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modified := modifiedExpression
		if modified {
			return true, &STPositionalArgumentNode{

				STFunctionArgumentNode: n.STFunctionArgumentNode,

				Expression: expressionNode,
			}
		}
		return false, current

	case *STRestArgumentNode:

		modifiedEllipsis, ellipsisNode := replaceInner(n.Ellipsis, target, replacement)

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modified := modifiedEllipsis || modifiedExpression
		if modified {
			return true, &STRestArgumentNode{

				STFunctionArgumentNode: n.STFunctionArgumentNode,

				Ellipsis: ellipsisNode,

				Expression: expressionNode,
			}
		}
		return false, current

	case *STInferredTypedescDefaultNode:

		modifiedLtToken, ltTokenNode := replaceInner(n.LtToken, target, replacement)

		modifiedGtToken, gtTokenNode := replaceInner(n.GtToken, target, replacement)

		modified := modifiedLtToken || modifiedGtToken
		if modified {
			return true, &STInferredTypedescDefaultNode{

				STExpressionNode: n.STExpressionNode,

				LtToken: ltTokenNode,

				GtToken: gtTokenNode,
			}
		}
		return false, current

	case *STObjectTypeDescriptorNode:

		modifiedObjectTypeQualifiers, objectTypeQualifiersNode := replaceInner(n.ObjectTypeQualifiers, target, replacement)

		modifiedObjectKeyword, objectKeywordNode := replaceInner(n.ObjectKeyword, target, replacement)

		modifiedOpenBrace, openBraceNode := replaceInner(n.OpenBrace, target, replacement)

		modifiedMembers, membersNode := replaceInner(n.Members, target, replacement)

		modifiedCloseBrace, closeBraceNode := replaceInner(n.CloseBrace, target, replacement)

		modified := modifiedObjectTypeQualifiers || modifiedObjectKeyword || modifiedOpenBrace || modifiedMembers || modifiedCloseBrace
		if modified {
			return true, &STObjectTypeDescriptorNode{

				STTypeDescriptorNode: n.STTypeDescriptorNode,

				ObjectTypeQualifiers: objectTypeQualifiersNode,

				ObjectKeyword: objectKeywordNode,

				OpenBrace: openBraceNode,

				Members: membersNode,

				CloseBrace: closeBraceNode,
			}
		}
		return false, current

	case *STObjectConstructorExpressionNode:

		modifiedAnnotations, annotationsNode := replaceInner(n.Annotations, target, replacement)

		modifiedObjectTypeQualifiers, objectTypeQualifiersNode := replaceInner(n.ObjectTypeQualifiers, target, replacement)

		modifiedObjectKeyword, objectKeywordNode := replaceInner(n.ObjectKeyword, target, replacement)

		modifiedTypeReference, typeReferenceNode := replaceInner(n.TypeReference, target, replacement)

		modifiedOpenBraceToken, openBraceTokenNode := replaceInner(n.OpenBraceToken, target, replacement)

		modifiedMembers, membersNode := replaceInner(n.Members, target, replacement)

		modifiedCloseBraceToken, closeBraceTokenNode := replaceInner(n.CloseBraceToken, target, replacement)

		modified := modifiedAnnotations || modifiedObjectTypeQualifiers || modifiedObjectKeyword || modifiedTypeReference || modifiedOpenBraceToken || modifiedMembers || modifiedCloseBraceToken
		if modified {
			return true, &STObjectConstructorExpressionNode{

				STExpressionNode: n.STExpressionNode,

				Annotations: annotationsNode,

				ObjectTypeQualifiers: objectTypeQualifiersNode,

				ObjectKeyword: objectKeywordNode,

				TypeReference: typeReferenceNode,

				OpenBraceToken: openBraceTokenNode,

				Members: membersNode,

				CloseBraceToken: closeBraceTokenNode,
			}
		}
		return false, current

	case *STRecordTypeDescriptorNode:

		modifiedRecordKeyword, recordKeywordNode := replaceInner(n.RecordKeyword, target, replacement)

		modifiedBodyStartDelimiter, bodyStartDelimiterNode := replaceInner(n.BodyStartDelimiter, target, replacement)

		modifiedFields, fieldsNode := replaceInner(n.Fields, target, replacement)

		modifiedRecordRestDescriptor, recordRestDescriptorNode := replaceInner(n.RecordRestDescriptor, target, replacement)

		modifiedBodyEndDelimiter, bodyEndDelimiterNode := replaceInner(n.BodyEndDelimiter, target, replacement)

		modified := modifiedRecordKeyword || modifiedBodyStartDelimiter || modifiedFields || modifiedRecordRestDescriptor || modifiedBodyEndDelimiter
		if modified {
			return true, &STRecordTypeDescriptorNode{

				STTypeDescriptorNode: n.STTypeDescriptorNode,

				RecordKeyword: recordKeywordNode,

				BodyStartDelimiter: bodyStartDelimiterNode,

				Fields: fieldsNode,

				RecordRestDescriptor: recordRestDescriptorNode,

				BodyEndDelimiter: bodyEndDelimiterNode,
			}
		}
		return false, current

	case *STReturnTypeDescriptorNode:

		modifiedReturnsKeyword, returnsKeywordNode := replaceInner(n.ReturnsKeyword, target, replacement)

		modifiedAnnotations, annotationsNode := replaceInner(n.Annotations, target, replacement)

		modifiedType, typeNode := replaceInner(n.Type, target, replacement)

		modified := modifiedReturnsKeyword || modifiedAnnotations || modifiedType
		if modified {
			return true, &STReturnTypeDescriptorNode{

				STNode: n.STNode,

				ReturnsKeyword: returnsKeywordNode,

				Annotations: annotationsNode,

				Type: typeNode,
			}
		}
		return false, current

	case *STNilTypeDescriptorNode:

		modifiedOpenParenToken, openParenTokenNode := replaceInner(n.OpenParenToken, target, replacement)

		modifiedCloseParenToken, closeParenTokenNode := replaceInner(n.CloseParenToken, target, replacement)

		modified := modifiedOpenParenToken || modifiedCloseParenToken
		if modified {
			return true, &STNilTypeDescriptorNode{

				STTypeDescriptorNode: n.STTypeDescriptorNode,

				OpenParenToken: openParenTokenNode,

				CloseParenToken: closeParenTokenNode,
			}
		}
		return false, current

	case *STOptionalTypeDescriptorNode:

		modifiedTypeDescriptor, typeDescriptorNode := replaceInner(n.TypeDescriptor, target, replacement)

		modifiedQuestionMarkToken, questionMarkTokenNode := replaceInner(n.QuestionMarkToken, target, replacement)

		modified := modifiedTypeDescriptor || modifiedQuestionMarkToken
		if modified {
			return true, &STOptionalTypeDescriptorNode{

				STTypeDescriptorNode: n.STTypeDescriptorNode,

				TypeDescriptor: typeDescriptorNode,

				QuestionMarkToken: questionMarkTokenNode,
			}
		}
		return false, current

	case *STObjectFieldNode:

		modifiedMetadata, metadataNode := replaceInner(n.Metadata, target, replacement)

		modifiedVisibilityQualifier, visibilityQualifierNode := replaceInner(n.VisibilityQualifier, target, replacement)

		modifiedQualifierList, qualifierListNode := replaceInner(n.QualifierList, target, replacement)

		modifiedTypeName, typeNameNode := replaceInner(n.TypeName, target, replacement)

		modifiedFieldName, fieldNameNode := replaceInner(n.FieldName, target, replacement)

		modifiedEqualsToken, equalsTokenNode := replaceInner(n.EqualsToken, target, replacement)

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modifiedSemicolonToken, semicolonTokenNode := replaceInner(n.SemicolonToken, target, replacement)

		modified := modifiedMetadata || modifiedVisibilityQualifier || modifiedQualifierList || modifiedTypeName || modifiedFieldName || modifiedEqualsToken || modifiedExpression || modifiedSemicolonToken
		if modified {
			return true, &STObjectFieldNode{

				STNode: n.STNode,

				Metadata: metadataNode,

				VisibilityQualifier: visibilityQualifierNode,

				QualifierList: qualifierListNode,

				TypeName: typeNameNode,

				FieldName: fieldNameNode,

				EqualsToken: equalsTokenNode,

				Expression: expressionNode,

				SemicolonToken: semicolonTokenNode,
			}
		}
		return false, current

	case *STRecordFieldNode:

		modifiedMetadata, metadataNode := replaceInner(n.Metadata, target, replacement)

		modifiedReadonlyKeyword, readonlyKeywordNode := replaceInner(n.ReadonlyKeyword, target, replacement)

		modifiedTypeName, typeNameNode := replaceInner(n.TypeName, target, replacement)

		modifiedFieldName, fieldNameNode := replaceInner(n.FieldName, target, replacement)

		modifiedQuestionMarkToken, questionMarkTokenNode := replaceInner(n.QuestionMarkToken, target, replacement)

		modifiedSemicolonToken, semicolonTokenNode := replaceInner(n.SemicolonToken, target, replacement)

		modified := modifiedMetadata || modifiedReadonlyKeyword || modifiedTypeName || modifiedFieldName || modifiedQuestionMarkToken || modifiedSemicolonToken
		if modified {
			return true, &STRecordFieldNode{

				STNode: n.STNode,

				Metadata: metadataNode,

				ReadonlyKeyword: readonlyKeywordNode,

				TypeName: typeNameNode,

				FieldName: fieldNameNode,

				QuestionMarkToken: questionMarkTokenNode,

				SemicolonToken: semicolonTokenNode,
			}
		}
		return false, current

	case *STRecordFieldWithDefaultValueNode:

		modifiedMetadata, metadataNode := replaceInner(n.Metadata, target, replacement)

		modifiedReadonlyKeyword, readonlyKeywordNode := replaceInner(n.ReadonlyKeyword, target, replacement)

		modifiedTypeName, typeNameNode := replaceInner(n.TypeName, target, replacement)

		modifiedFieldName, fieldNameNode := replaceInner(n.FieldName, target, replacement)

		modifiedEqualsToken, equalsTokenNode := replaceInner(n.EqualsToken, target, replacement)

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modifiedSemicolonToken, semicolonTokenNode := replaceInner(n.SemicolonToken, target, replacement)

		modified := modifiedMetadata || modifiedReadonlyKeyword || modifiedTypeName || modifiedFieldName || modifiedEqualsToken || modifiedExpression || modifiedSemicolonToken
		if modified {
			return true, &STRecordFieldWithDefaultValueNode{

				STNode: n.STNode,

				Metadata: metadataNode,

				ReadonlyKeyword: readonlyKeywordNode,

				TypeName: typeNameNode,

				FieldName: fieldNameNode,

				EqualsToken: equalsTokenNode,

				Expression: expressionNode,

				SemicolonToken: semicolonTokenNode,
			}
		}
		return false, current

	case *STRecordRestDescriptorNode:

		modifiedTypeName, typeNameNode := replaceInner(n.TypeName, target, replacement)

		modifiedEllipsisToken, ellipsisTokenNode := replaceInner(n.EllipsisToken, target, replacement)

		modifiedSemicolonToken, semicolonTokenNode := replaceInner(n.SemicolonToken, target, replacement)

		modified := modifiedTypeName || modifiedEllipsisToken || modifiedSemicolonToken
		if modified {
			return true, &STRecordRestDescriptorNode{

				STNode: n.STNode,

				TypeName: typeNameNode,

				EllipsisToken: ellipsisTokenNode,

				SemicolonToken: semicolonTokenNode,
			}
		}
		return false, current

	case *STTypeReferenceNode:

		modifiedAsteriskToken, asteriskTokenNode := replaceInner(n.AsteriskToken, target, replacement)

		modifiedTypeName, typeNameNode := replaceInner(n.TypeName, target, replacement)

		modifiedSemicolonToken, semicolonTokenNode := replaceInner(n.SemicolonToken, target, replacement)

		modified := modifiedAsteriskToken || modifiedTypeName || modifiedSemicolonToken
		if modified {
			return true, &STTypeReferenceNode{

				STTypeDescriptorNode: n.STTypeDescriptorNode,

				AsteriskToken: asteriskTokenNode,

				TypeName: typeNameNode,

				SemicolonToken: semicolonTokenNode,
			}
		}
		return false, current

	case *STAnnotationNode:

		modifiedAtToken, atTokenNode := replaceInner(n.AtToken, target, replacement)

		modifiedAnnotReference, annotReferenceNode := replaceInner(n.AnnotReference, target, replacement)

		modifiedAnnotValue, annotValueNode := replaceInner(n.AnnotValue, target, replacement)

		modified := modifiedAtToken || modifiedAnnotReference || modifiedAnnotValue
		if modified {
			return true, &STAnnotationNode{

				STNode: n.STNode,

				AtToken: atTokenNode,

				AnnotReference: annotReferenceNode,

				AnnotValue: annotValueNode,
			}
		}
		return false, current

	case *STMetadataNode:

		modifiedDocumentationString, documentationStringNode := replaceInner(n.DocumentationString, target, replacement)

		modifiedAnnotations, annotationsNode := replaceInner(n.Annotations, target, replacement)

		modified := modifiedDocumentationString || modifiedAnnotations
		if modified {
			return true, &STMetadataNode{

				STNode: n.STNode,

				DocumentationString: documentationStringNode,

				Annotations: annotationsNode,
			}
		}
		return false, current

	case *STModuleVariableDeclarationNode:

		modifiedMetadata, metadataNode := replaceInner(n.Metadata, target, replacement)

		modifiedVisibilityQualifier, visibilityQualifierNode := replaceInner(n.VisibilityQualifier, target, replacement)

		modifiedQualifiers, qualifiersNode := replaceInner(n.Qualifiers, target, replacement)

		modifiedTypedBindingPattern, typedBindingPatternNode := replaceInner(n.TypedBindingPattern, target, replacement)

		modifiedEqualsToken, equalsTokenNode := replaceInner(n.EqualsToken, target, replacement)

		modifiedInitializer, initializerNode := replaceInner(n.Initializer, target, replacement)

		modifiedSemicolonToken, semicolonTokenNode := replaceInner(n.SemicolonToken, target, replacement)

		modified := modifiedMetadata || modifiedVisibilityQualifier || modifiedQualifiers || modifiedTypedBindingPattern || modifiedEqualsToken || modifiedInitializer || modifiedSemicolonToken
		if modified {
			return true, &STModuleVariableDeclarationNode{

				STModuleMemberDeclarationNode: n.STModuleMemberDeclarationNode,

				Metadata: metadataNode,

				VisibilityQualifier: visibilityQualifierNode,

				Qualifiers: qualifiersNode,

				TypedBindingPattern: typedBindingPatternNode,

				EqualsToken: equalsTokenNode,

				Initializer: initializerNode,

				SemicolonToken: semicolonTokenNode,
			}
		}
		return false, current

	case *STTypeTestExpressionNode:

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modifiedIsKeyword, isKeywordNode := replaceInner(n.IsKeyword, target, replacement)

		modifiedTypeDescriptor, typeDescriptorNode := replaceInner(n.TypeDescriptor, target, replacement)

		modified := modifiedExpression || modifiedIsKeyword || modifiedTypeDescriptor
		if modified {
			return true, &STTypeTestExpressionNode{

				STExpressionNode: n.STExpressionNode,

				Expression: expressionNode,

				IsKeyword: isKeywordNode,

				TypeDescriptor: typeDescriptorNode,
			}
		}
		return false, current

	case *STRemoteMethodCallActionNode:

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modifiedRightArrowToken, rightArrowTokenNode := replaceInner(n.RightArrowToken, target, replacement)

		modifiedMethodName, methodNameNode := replaceInner(n.MethodName, target, replacement)

		modifiedOpenParenToken, openParenTokenNode := replaceInner(n.OpenParenToken, target, replacement)

		modifiedArguments, argumentsNode := replaceInner(n.Arguments, target, replacement)

		modifiedCloseParenToken, closeParenTokenNode := replaceInner(n.CloseParenToken, target, replacement)

		modified := modifiedExpression || modifiedRightArrowToken || modifiedMethodName || modifiedOpenParenToken || modifiedArguments || modifiedCloseParenToken
		if modified {
			return true, &STRemoteMethodCallActionNode{

				STActionNode: n.STActionNode,

				Expression: expressionNode,

				RightArrowToken: rightArrowTokenNode,

				MethodName: methodNameNode,

				OpenParenToken: openParenTokenNode,

				Arguments: argumentsNode,

				CloseParenToken: closeParenTokenNode,
			}
		}
		return false, current

	case *STMapTypeDescriptorNode:

		modifiedMapKeywordToken, mapKeywordTokenNode := replaceInner(n.MapKeywordToken, target, replacement)

		modifiedMapTypeParamsNode, mapTypeParamsNodeNode := replaceInner(n.MapTypeParamsNode, target, replacement)

		modified := modifiedMapKeywordToken || modifiedMapTypeParamsNode
		if modified {
			return true, &STMapTypeDescriptorNode{

				STTypeDescriptorNode: n.STTypeDescriptorNode,

				MapKeywordToken: mapKeywordTokenNode,

				MapTypeParamsNode: mapTypeParamsNodeNode,
			}
		}
		return false, current

	case *STNilLiteralNode:

		modifiedOpenParenToken, openParenTokenNode := replaceInner(n.OpenParenToken, target, replacement)

		modifiedCloseParenToken, closeParenTokenNode := replaceInner(n.CloseParenToken, target, replacement)

		modified := modifiedOpenParenToken || modifiedCloseParenToken
		if modified {
			return true, &STNilLiteralNode{

				STExpressionNode: n.STExpressionNode,

				OpenParenToken: openParenTokenNode,

				CloseParenToken: closeParenTokenNode,
			}
		}
		return false, current

	case *STAnnotationDeclarationNode:

		modifiedMetadata, metadataNode := replaceInner(n.Metadata, target, replacement)

		modifiedVisibilityQualifier, visibilityQualifierNode := replaceInner(n.VisibilityQualifier, target, replacement)

		modifiedConstKeyword, constKeywordNode := replaceInner(n.ConstKeyword, target, replacement)

		modifiedAnnotationKeyword, annotationKeywordNode := replaceInner(n.AnnotationKeyword, target, replacement)

		modifiedTypeDescriptor, typeDescriptorNode := replaceInner(n.TypeDescriptor, target, replacement)

		modifiedAnnotationTag, annotationTagNode := replaceInner(n.AnnotationTag, target, replacement)

		modifiedOnKeyword, onKeywordNode := replaceInner(n.OnKeyword, target, replacement)

		modifiedAttachPoints, attachPointsNode := replaceInner(n.AttachPoints, target, replacement)

		modifiedSemicolonToken, semicolonTokenNode := replaceInner(n.SemicolonToken, target, replacement)

		modified := modifiedMetadata || modifiedVisibilityQualifier || modifiedConstKeyword || modifiedAnnotationKeyword || modifiedTypeDescriptor || modifiedAnnotationTag || modifiedOnKeyword || modifiedAttachPoints || modifiedSemicolonToken
		if modified {
			return true, &STAnnotationDeclarationNode{

				STModuleMemberDeclarationNode: n.STModuleMemberDeclarationNode,

				Metadata: metadataNode,

				VisibilityQualifier: visibilityQualifierNode,

				ConstKeyword: constKeywordNode,

				AnnotationKeyword: annotationKeywordNode,

				TypeDescriptor: typeDescriptorNode,

				AnnotationTag: annotationTagNode,

				OnKeyword: onKeywordNode,

				AttachPoints: attachPointsNode,

				SemicolonToken: semicolonTokenNode,
			}
		}
		return false, current

	case *STAnnotationAttachPointNode:

		modifiedSourceKeyword, sourceKeywordNode := replaceInner(n.SourceKeyword, target, replacement)

		modifiedIdentifiers, identifiersNode := replaceInner(n.Identifiers, target, replacement)

		modified := modifiedSourceKeyword || modifiedIdentifiers
		if modified {
			return true, &STAnnotationAttachPointNode{

				STNode: n.STNode,

				SourceKeyword: sourceKeywordNode,

				Identifiers: identifiersNode,
			}
		}
		return false, current

	case *STXMLNamespaceDeclarationNode:

		modifiedXmlnsKeyword, xmlnsKeywordNode := replaceInner(n.XmlnsKeyword, target, replacement)

		modifiedNamespaceuri, namespaceuriNode := replaceInner(n.Namespaceuri, target, replacement)

		modifiedAsKeyword, asKeywordNode := replaceInner(n.AsKeyword, target, replacement)

		modifiedNamespacePrefix, namespacePrefixNode := replaceInner(n.NamespacePrefix, target, replacement)

		modifiedSemicolonToken, semicolonTokenNode := replaceInner(n.SemicolonToken, target, replacement)

		modified := modifiedXmlnsKeyword || modifiedNamespaceuri || modifiedAsKeyword || modifiedNamespacePrefix || modifiedSemicolonToken
		if modified {
			return true, &STXMLNamespaceDeclarationNode{

				STStatementNode: n.STStatementNode,

				XmlnsKeyword: xmlnsKeywordNode,

				Namespaceuri: namespaceuriNode,

				AsKeyword: asKeywordNode,

				NamespacePrefix: namespacePrefixNode,

				SemicolonToken: semicolonTokenNode,
			}
		}
		return false, current

	case *STModuleXMLNamespaceDeclarationNode:

		modifiedXmlnsKeyword, xmlnsKeywordNode := replaceInner(n.XmlnsKeyword, target, replacement)

		modifiedNamespaceuri, namespaceuriNode := replaceInner(n.Namespaceuri, target, replacement)

		modifiedAsKeyword, asKeywordNode := replaceInner(n.AsKeyword, target, replacement)

		modifiedNamespacePrefix, namespacePrefixNode := replaceInner(n.NamespacePrefix, target, replacement)

		modifiedSemicolonToken, semicolonTokenNode := replaceInner(n.SemicolonToken, target, replacement)

		modified := modifiedXmlnsKeyword || modifiedNamespaceuri || modifiedAsKeyword || modifiedNamespacePrefix || modifiedSemicolonToken
		if modified {
			return true, &STModuleXMLNamespaceDeclarationNode{

				STModuleMemberDeclarationNode: n.STModuleMemberDeclarationNode,

				XmlnsKeyword: xmlnsKeywordNode,

				Namespaceuri: namespaceuriNode,

				AsKeyword: asKeywordNode,

				NamespacePrefix: namespacePrefixNode,

				SemicolonToken: semicolonTokenNode,
			}
		}
		return false, current

	case *STFunctionBodyBlockNode:

		modifiedOpenBraceToken, openBraceTokenNode := replaceInner(n.OpenBraceToken, target, replacement)

		modifiedNamedWorkerDeclarator, namedWorkerDeclaratorNode := replaceInner(n.NamedWorkerDeclarator, target, replacement)

		modifiedStatements, statementsNode := replaceInner(n.Statements, target, replacement)

		modifiedCloseBraceToken, closeBraceTokenNode := replaceInner(n.CloseBraceToken, target, replacement)

		modifiedSemicolonToken, semicolonTokenNode := replaceInner(n.SemicolonToken, target, replacement)

		modified := modifiedOpenBraceToken || modifiedNamedWorkerDeclarator || modifiedStatements || modifiedCloseBraceToken || modifiedSemicolonToken
		if modified {
			return true, &STFunctionBodyBlockNode{

				STFunctionBodyNode: n.STFunctionBodyNode,

				OpenBraceToken: openBraceTokenNode,

				NamedWorkerDeclarator: namedWorkerDeclaratorNode,

				Statements: statementsNode,

				CloseBraceToken: closeBraceTokenNode,

				SemicolonToken: semicolonTokenNode,
			}
		}
		return false, current

	case *STNamedWorkerDeclarationNode:

		modifiedAnnotations, annotationsNode := replaceInner(n.Annotations, target, replacement)

		modifiedTransactionalKeyword, transactionalKeywordNode := replaceInner(n.TransactionalKeyword, target, replacement)

		modifiedWorkerKeyword, workerKeywordNode := replaceInner(n.WorkerKeyword, target, replacement)

		modifiedWorkerName, workerNameNode := replaceInner(n.WorkerName, target, replacement)

		modifiedReturnTypeDesc, returnTypeDescNode := replaceInner(n.ReturnTypeDesc, target, replacement)

		modifiedWorkerBody, workerBodyNode := replaceInner(n.WorkerBody, target, replacement)

		modifiedOnFailClause, onFailClauseNode := replaceInner(n.OnFailClause, target, replacement)

		modified := modifiedAnnotations || modifiedTransactionalKeyword || modifiedWorkerKeyword || modifiedWorkerName || modifiedReturnTypeDesc || modifiedWorkerBody || modifiedOnFailClause
		if modified {
			return true, &STNamedWorkerDeclarationNode{

				STNode: n.STNode,

				Annotations: annotationsNode,

				TransactionalKeyword: transactionalKeywordNode,

				WorkerKeyword: workerKeywordNode,

				WorkerName: workerNameNode,

				ReturnTypeDesc: returnTypeDescNode,

				WorkerBody: workerBodyNode,

				OnFailClause: onFailClauseNode,
			}
		}
		return false, current

	case *STNamedWorkerDeclarator:

		modifiedWorkerInitStatements, workerInitStatementsNode := replaceInner(n.WorkerInitStatements, target, replacement)

		modifiedNamedWorkerDeclarations, namedWorkerDeclarationsNode := replaceInner(n.NamedWorkerDeclarations, target, replacement)

		modified := modifiedWorkerInitStatements || modifiedNamedWorkerDeclarations
		if modified {
			return true, &STNamedWorkerDeclarator{

				STNode: n.STNode,

				WorkerInitStatements: workerInitStatementsNode,

				NamedWorkerDeclarations: namedWorkerDeclarationsNode,
			}
		}
		return false, current

	case *STBasicLiteralNode:

		modifiedLiteralToken, literalTokenNode := replaceInner(n.LiteralToken, target, replacement)

		modified := modifiedLiteralToken
		if modified {
			return true, &STBasicLiteralNode{

				STExpressionNode: n.STExpressionNode,

				LiteralToken: literalTokenNode,
			}
		}
		return false, current

	case *STSimpleNameReferenceNode:

		modifiedName, nameNode := replaceInner(n.Name, target, replacement)

		modified := modifiedName
		if modified {
			return true, &STSimpleNameReferenceNode{

				STNameReferenceNode: n.STNameReferenceNode,

				Name: nameNode,
			}
		}
		return false, current

	case *STQualifiedNameReferenceNode:

		modifiedModulePrefix, modulePrefixNode := replaceInner(n.ModulePrefix, target, replacement)

		modifiedColon, colonNode := replaceInner(n.Colon, target, replacement)

		modifiedIdentifier, identifierNode := replaceInner(n.Identifier, target, replacement)

		modified := modifiedModulePrefix || modifiedColon || modifiedIdentifier
		if modified {
			return true, &STQualifiedNameReferenceNode{

				STNameReferenceNode: n.STNameReferenceNode,

				ModulePrefix: modulePrefixNode,

				Colon: colonNode,

				Identifier: identifierNode,
			}
		}
		return false, current

	case *STBuiltinSimpleNameReferenceNode:

		modifiedName, nameNode := replaceInner(n.Name, target, replacement)

		modified := modifiedName
		if modified {
			return true, &STBuiltinSimpleNameReferenceNode{

				STNameReferenceNode: n.STNameReferenceNode,

				Name: nameNode,
			}
		}
		return false, current

	case *STTrapExpressionNode:

		modifiedTrapKeyword, trapKeywordNode := replaceInner(n.TrapKeyword, target, replacement)

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modified := modifiedTrapKeyword || modifiedExpression
		if modified {
			return true, &STTrapExpressionNode{

				STExpressionNode: n.STExpressionNode,

				TrapKeyword: trapKeywordNode,

				Expression: expressionNode,
			}
		}
		return false, current

	case *STListConstructorExpressionNode:

		modifiedOpenBracket, openBracketNode := replaceInner(n.OpenBracket, target, replacement)

		modifiedExpressions, expressionsNode := replaceInner(n.Expressions, target, replacement)

		modifiedCloseBracket, closeBracketNode := replaceInner(n.CloseBracket, target, replacement)

		modified := modifiedOpenBracket || modifiedExpressions || modifiedCloseBracket
		if modified {
			return true, &STListConstructorExpressionNode{

				STExpressionNode: n.STExpressionNode,

				OpenBracket: openBracketNode,

				Expressions: expressionsNode,

				CloseBracket: closeBracketNode,
			}
		}
		return false, current

	case *STTypeCastExpressionNode:

		modifiedLtToken, ltTokenNode := replaceInner(n.LtToken, target, replacement)

		modifiedTypeCastParam, typeCastParamNode := replaceInner(n.TypeCastParam, target, replacement)

		modifiedGtToken, gtTokenNode := replaceInner(n.GtToken, target, replacement)

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modified := modifiedLtToken || modifiedTypeCastParam || modifiedGtToken || modifiedExpression
		if modified {
			return true, &STTypeCastExpressionNode{

				STExpressionNode: n.STExpressionNode,

				LtToken: ltTokenNode,

				TypeCastParam: typeCastParamNode,

				GtToken: gtTokenNode,

				Expression: expressionNode,
			}
		}
		return false, current

	case *STTypeCastParamNode:

		modifiedAnnotations, annotationsNode := replaceInner(n.Annotations, target, replacement)

		modifiedType, typeNode := replaceInner(n.Type, target, replacement)

		modified := modifiedAnnotations || modifiedType
		if modified {
			return true, &STTypeCastParamNode{

				STNode: n.STNode,

				Annotations: annotationsNode,

				Type: typeNode,
			}
		}
		return false, current

	case *STUnionTypeDescriptorNode:

		modifiedLeftTypeDesc, leftTypeDescNode := replaceInner(n.LeftTypeDesc, target, replacement)

		modifiedPipeToken, pipeTokenNode := replaceInner(n.PipeToken, target, replacement)

		modifiedRightTypeDesc, rightTypeDescNode := replaceInner(n.RightTypeDesc, target, replacement)

		modified := modifiedLeftTypeDesc || modifiedPipeToken || modifiedRightTypeDesc
		if modified {
			return true, &STUnionTypeDescriptorNode{

				STTypeDescriptorNode: n.STTypeDescriptorNode,

				LeftTypeDesc: leftTypeDescNode,

				PipeToken: pipeTokenNode,

				RightTypeDesc: rightTypeDescNode,
			}
		}
		return false, current

	case *STTableConstructorExpressionNode:

		modifiedTableKeyword, tableKeywordNode := replaceInner(n.TableKeyword, target, replacement)

		modifiedKeySpecifier, keySpecifierNode := replaceInner(n.KeySpecifier, target, replacement)

		modifiedOpenBracket, openBracketNode := replaceInner(n.OpenBracket, target, replacement)

		modifiedRows, rowsNode := replaceInner(n.Rows, target, replacement)

		modifiedCloseBracket, closeBracketNode := replaceInner(n.CloseBracket, target, replacement)

		modified := modifiedTableKeyword || modifiedKeySpecifier || modifiedOpenBracket || modifiedRows || modifiedCloseBracket
		if modified {
			return true, &STTableConstructorExpressionNode{

				STExpressionNode: n.STExpressionNode,

				TableKeyword: tableKeywordNode,

				KeySpecifier: keySpecifierNode,

				OpenBracket: openBracketNode,

				Rows: rowsNode,

				CloseBracket: closeBracketNode,
			}
		}
		return false, current

	case *STKeySpecifierNode:

		modifiedKeyKeyword, keyKeywordNode := replaceInner(n.KeyKeyword, target, replacement)

		modifiedOpenParenToken, openParenTokenNode := replaceInner(n.OpenParenToken, target, replacement)

		modifiedFieldNames, fieldNamesNode := replaceInner(n.FieldNames, target, replacement)

		modifiedCloseParenToken, closeParenTokenNode := replaceInner(n.CloseParenToken, target, replacement)

		modified := modifiedKeyKeyword || modifiedOpenParenToken || modifiedFieldNames || modifiedCloseParenToken
		if modified {
			return true, &STKeySpecifierNode{

				STNode: n.STNode,

				KeyKeyword: keyKeywordNode,

				OpenParenToken: openParenTokenNode,

				FieldNames: fieldNamesNode,

				CloseParenToken: closeParenTokenNode,
			}
		}
		return false, current

	case *STStreamTypeDescriptorNode:

		modifiedStreamKeywordToken, streamKeywordTokenNode := replaceInner(n.StreamKeywordToken, target, replacement)

		modifiedStreamTypeParamsNode, streamTypeParamsNodeNode := replaceInner(n.StreamTypeParamsNode, target, replacement)

		modified := modifiedStreamKeywordToken || modifiedStreamTypeParamsNode
		if modified {
			return true, &STStreamTypeDescriptorNode{

				STTypeDescriptorNode: n.STTypeDescriptorNode,

				StreamKeywordToken: streamKeywordTokenNode,

				StreamTypeParamsNode: streamTypeParamsNodeNode,
			}
		}
		return false, current

	case *STStreamTypeParamsNode:

		modifiedLtToken, ltTokenNode := replaceInner(n.LtToken, target, replacement)

		modifiedLeftTypeDescNode, leftTypeDescNodeNode := replaceInner(n.LeftTypeDescNode, target, replacement)

		modifiedCommaToken, commaTokenNode := replaceInner(n.CommaToken, target, replacement)

		modifiedRightTypeDescNode, rightTypeDescNodeNode := replaceInner(n.RightTypeDescNode, target, replacement)

		modifiedGtToken, gtTokenNode := replaceInner(n.GtToken, target, replacement)

		modified := modifiedLtToken || modifiedLeftTypeDescNode || modifiedCommaToken || modifiedRightTypeDescNode || modifiedGtToken
		if modified {
			return true, &STStreamTypeParamsNode{

				STNode: n.STNode,

				LtToken: ltTokenNode,

				LeftTypeDescNode: leftTypeDescNodeNode,

				CommaToken: commaTokenNode,

				RightTypeDescNode: rightTypeDescNodeNode,

				GtToken: gtTokenNode,
			}
		}
		return false, current

	case *STLetExpressionNode:

		modifiedLetKeyword, letKeywordNode := replaceInner(n.LetKeyword, target, replacement)

		modifiedLetVarDeclarations, letVarDeclarationsNode := replaceInner(n.LetVarDeclarations, target, replacement)

		modifiedInKeyword, inKeywordNode := replaceInner(n.InKeyword, target, replacement)

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modified := modifiedLetKeyword || modifiedLetVarDeclarations || modifiedInKeyword || modifiedExpression
		if modified {
			return true, &STLetExpressionNode{

				STExpressionNode: n.STExpressionNode,

				LetKeyword: letKeywordNode,

				LetVarDeclarations: letVarDeclarationsNode,

				InKeyword: inKeywordNode,

				Expression: expressionNode,
			}
		}
		return false, current

	case *STLetVariableDeclarationNode:

		modifiedAnnotations, annotationsNode := replaceInner(n.Annotations, target, replacement)

		modifiedTypedBindingPattern, typedBindingPatternNode := replaceInner(n.TypedBindingPattern, target, replacement)

		modifiedEqualsToken, equalsTokenNode := replaceInner(n.EqualsToken, target, replacement)

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modified := modifiedAnnotations || modifiedTypedBindingPattern || modifiedEqualsToken || modifiedExpression
		if modified {
			return true, &STLetVariableDeclarationNode{

				STNode: n.STNode,

				Annotations: annotationsNode,

				TypedBindingPattern: typedBindingPatternNode,

				EqualsToken: equalsTokenNode,

				Expression: expressionNode,
			}
		}
		return false, current

	case *STTemplateExpressionNode:

		modifiedType, typeNode := replaceInner(n.Type, target, replacement)

		modifiedStartBacktick, startBacktickNode := replaceInner(n.StartBacktick, target, replacement)

		modifiedContent, contentNode := replaceInner(n.Content, target, replacement)

		modifiedEndBacktick, endBacktickNode := replaceInner(n.EndBacktick, target, replacement)

		modified := modifiedType || modifiedStartBacktick || modifiedContent || modifiedEndBacktick
		if modified {
			return true, &STTemplateExpressionNode{

				STExpressionNode: n.STExpressionNode,

				Type: typeNode,

				StartBacktick: startBacktickNode,

				Content: contentNode,

				EndBacktick: endBacktickNode,
			}
		}
		return false, current

	case *STXMLElementNode:

		modifiedStartTag, startTagNode := replaceInner(n.StartTag, target, replacement)

		modifiedContent, contentNode := replaceInner(n.Content, target, replacement)

		modifiedEndTag, endTagNode := replaceInner(n.EndTag, target, replacement)

		modified := modifiedStartTag || modifiedContent || modifiedEndTag
		if modified {
			return true, &STXMLElementNode{

				STXMLItemNode: n.STXMLItemNode,

				StartTag: startTagNode,

				Content: contentNode,

				EndTag: endTagNode,
			}
		}
		return false, current

	case *STXMLStartTagNode:

		modifiedLtToken, ltTokenNode := replaceInner(n.LtToken, target, replacement)

		modifiedName, nameNode := replaceInner(n.Name, target, replacement)

		modifiedAttributes, attributesNode := replaceInner(n.Attributes, target, replacement)

		modifiedGetToken, getTokenNode := replaceInner(n.GetToken, target, replacement)

		modified := modifiedLtToken || modifiedName || modifiedAttributes || modifiedGetToken
		if modified {
			return true, &STXMLStartTagNode{

				STXMLElementTagNode: n.STXMLElementTagNode,

				LtToken: ltTokenNode,

				Name: nameNode,

				Attributes: attributesNode,

				GetToken: getTokenNode,
			}
		}
		return false, current

	case *STXMLEndTagNode:

		modifiedLtToken, ltTokenNode := replaceInner(n.LtToken, target, replacement)

		modifiedSlashToken, slashTokenNode := replaceInner(n.SlashToken, target, replacement)

		modifiedName, nameNode := replaceInner(n.Name, target, replacement)

		modifiedGetToken, getTokenNode := replaceInner(n.GetToken, target, replacement)

		modified := modifiedLtToken || modifiedSlashToken || modifiedName || modifiedGetToken
		if modified {
			return true, &STXMLEndTagNode{

				STXMLElementTagNode: n.STXMLElementTagNode,

				LtToken: ltTokenNode,

				SlashToken: slashTokenNode,

				Name: nameNode,

				GetToken: getTokenNode,
			}
		}
		return false, current

	case *STXMLSimpleNameNode:

		modifiedName, nameNode := replaceInner(n.Name, target, replacement)

		modified := modifiedName
		if modified {
			return true, &STXMLSimpleNameNode{

				STXMLNameNode: n.STXMLNameNode,

				Name: nameNode,
			}
		}
		return false, current

	case *STXMLQualifiedNameNode:

		modifiedPrefix, prefixNode := replaceInner(n.Prefix, target, replacement)

		modifiedColon, colonNode := replaceInner(n.Colon, target, replacement)

		modifiedName, nameNode := replaceInner(n.Name, target, replacement)

		modified := modifiedPrefix || modifiedColon || modifiedName
		if modified {
			return true, &STXMLQualifiedNameNode{

				STXMLNameNode: n.STXMLNameNode,

				Prefix: prefixNode,

				Colon: colonNode,

				Name: nameNode,
			}
		}
		return false, current

	case *STXMLEmptyElementNode:

		modifiedLtToken, ltTokenNode := replaceInner(n.LtToken, target, replacement)

		modifiedName, nameNode := replaceInner(n.Name, target, replacement)

		modifiedAttributes, attributesNode := replaceInner(n.Attributes, target, replacement)

		modifiedSlashToken, slashTokenNode := replaceInner(n.SlashToken, target, replacement)

		modifiedGetToken, getTokenNode := replaceInner(n.GetToken, target, replacement)

		modified := modifiedLtToken || modifiedName || modifiedAttributes || modifiedSlashToken || modifiedGetToken
		if modified {
			return true, &STXMLEmptyElementNode{

				STXMLItemNode: n.STXMLItemNode,

				LtToken: ltTokenNode,

				Name: nameNode,

				Attributes: attributesNode,

				SlashToken: slashTokenNode,

				GetToken: getTokenNode,
			}
		}
		return false, current

	case *STInterpolationNode:

		modifiedInterpolationStartToken, interpolationStartTokenNode := replaceInner(n.InterpolationStartToken, target, replacement)

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modifiedInterpolationEndToken, interpolationEndTokenNode := replaceInner(n.InterpolationEndToken, target, replacement)

		modified := modifiedInterpolationStartToken || modifiedExpression || modifiedInterpolationEndToken
		if modified {
			return true, &STInterpolationNode{

				STXMLItemNode: n.STXMLItemNode,

				InterpolationStartToken: interpolationStartTokenNode,

				Expression: expressionNode,

				InterpolationEndToken: interpolationEndTokenNode,
			}
		}
		return false, current

	case *STXMLTextNode:

		modifiedContent, contentNode := replaceInner(n.Content, target, replacement)

		modified := modifiedContent
		if modified {
			return true, &STXMLTextNode{

				STXMLItemNode: n.STXMLItemNode,

				Content: contentNode,
			}
		}
		return false, current

	case *STXMLAttributeNode:

		modifiedAttributeName, attributeNameNode := replaceInner(n.AttributeName, target, replacement)

		modifiedEqualToken, equalTokenNode := replaceInner(n.EqualToken, target, replacement)

		modifiedValue, valueNode := replaceInner(n.Value, target, replacement)

		modified := modifiedAttributeName || modifiedEqualToken || modifiedValue
		if modified {
			return true, &STXMLAttributeNode{

				STNode: n.STNode,

				AttributeName: attributeNameNode,

				EqualToken: equalTokenNode,

				Value: valueNode,
			}
		}
		return false, current

	case *STXMLAttributeValue:

		modifiedStartQuote, startQuoteNode := replaceInner(n.StartQuote, target, replacement)

		modifiedValue, valueNode := replaceInner(n.Value, target, replacement)

		modifiedEndQuote, endQuoteNode := replaceInner(n.EndQuote, target, replacement)

		modified := modifiedStartQuote || modifiedValue || modifiedEndQuote
		if modified {
			return true, &STXMLAttributeValue{

				STNode: n.STNode,

				StartQuote: startQuoteNode,

				Value: valueNode,

				EndQuote: endQuoteNode,
			}
		}
		return false, current

	case *STXMLComment:

		modifiedCommentStart, commentStartNode := replaceInner(n.CommentStart, target, replacement)

		modifiedContent, contentNode := replaceInner(n.Content, target, replacement)

		modifiedCommentEnd, commentEndNode := replaceInner(n.CommentEnd, target, replacement)

		modified := modifiedCommentStart || modifiedContent || modifiedCommentEnd
		if modified {
			return true, &STXMLComment{

				STXMLItemNode: n.STXMLItemNode,

				CommentStart: commentStartNode,

				Content: contentNode,

				CommentEnd: commentEndNode,
			}
		}
		return false, current

	case *STXMLCDATANode:

		modifiedCdataStart, cdataStartNode := replaceInner(n.CdataStart, target, replacement)

		modifiedContent, contentNode := replaceInner(n.Content, target, replacement)

		modifiedCdataEnd, cdataEndNode := replaceInner(n.CdataEnd, target, replacement)

		modified := modifiedCdataStart || modifiedContent || modifiedCdataEnd
		if modified {
			return true, &STXMLCDATANode{

				STXMLItemNode: n.STXMLItemNode,

				CdataStart: cdataStartNode,

				Content: contentNode,

				CdataEnd: cdataEndNode,
			}
		}
		return false, current

	case *STXMLProcessingInstruction:

		modifiedPiStart, piStartNode := replaceInner(n.PiStart, target, replacement)

		modifiedTarget, targetNode := replaceInner(n.Target, target, replacement)

		modifiedData, dataNode := replaceInner(n.Data, target, replacement)

		modifiedPiEnd, piEndNode := replaceInner(n.PiEnd, target, replacement)

		modified := modifiedPiStart || modifiedTarget || modifiedData || modifiedPiEnd
		if modified {
			return true, &STXMLProcessingInstruction{

				STXMLItemNode: n.STXMLItemNode,

				PiStart: piStartNode,

				Target: targetNode,

				Data: dataNode,

				PiEnd: piEndNode,
			}
		}
		return false, current

	case *STTableTypeDescriptorNode:

		modifiedTableKeywordToken, tableKeywordTokenNode := replaceInner(n.TableKeywordToken, target, replacement)

		modifiedRowTypeParameterNode, rowTypeParameterNodeNode := replaceInner(n.RowTypeParameterNode, target, replacement)

		modifiedKeyConstraintNode, keyConstraintNodeNode := replaceInner(n.KeyConstraintNode, target, replacement)

		modified := modifiedTableKeywordToken || modifiedRowTypeParameterNode || modifiedKeyConstraintNode
		if modified {
			return true, &STTableTypeDescriptorNode{

				STTypeDescriptorNode: n.STTypeDescriptorNode,

				TableKeywordToken: tableKeywordTokenNode,

				RowTypeParameterNode: rowTypeParameterNodeNode,

				KeyConstraintNode: keyConstraintNodeNode,
			}
		}
		return false, current

	case *STTypeParameterNode:

		modifiedLtToken, ltTokenNode := replaceInner(n.LtToken, target, replacement)

		modifiedTypeNode, typeNodeNode := replaceInner(n.TypeNode, target, replacement)

		modifiedGtToken, gtTokenNode := replaceInner(n.GtToken, target, replacement)

		modified := modifiedLtToken || modifiedTypeNode || modifiedGtToken
		if modified {
			return true, &STTypeParameterNode{

				STNode: n.STNode,

				LtToken: ltTokenNode,

				TypeNode: typeNodeNode,

				GtToken: gtTokenNode,
			}
		}
		return false, current

	case *STKeyTypeConstraintNode:

		modifiedKeyKeywordToken, keyKeywordTokenNode := replaceInner(n.KeyKeywordToken, target, replacement)

		modifiedTypeParameterNode, typeParameterNodeNode := replaceInner(n.TypeParameterNode, target, replacement)

		modified := modifiedKeyKeywordToken || modifiedTypeParameterNode
		if modified {
			return true, &STKeyTypeConstraintNode{

				STNode: n.STNode,

				KeyKeywordToken: keyKeywordTokenNode,

				TypeParameterNode: typeParameterNodeNode,
			}
		}
		return false, current

	case *STFunctionTypeDescriptorNode:

		modifiedQualifierList, qualifierListNode := replaceInner(n.QualifierList, target, replacement)

		modifiedFunctionKeyword, functionKeywordNode := replaceInner(n.FunctionKeyword, target, replacement)

		modifiedFunctionSignature, functionSignatureNode := replaceInner(n.FunctionSignature, target, replacement)

		modified := modifiedQualifierList || modifiedFunctionKeyword || modifiedFunctionSignature
		if modified {
			return true, &STFunctionTypeDescriptorNode{

				STTypeDescriptorNode: n.STTypeDescriptorNode,

				QualifierList: qualifierListNode,

				FunctionKeyword: functionKeywordNode,

				FunctionSignature: functionSignatureNode,
			}
		}
		return false, current

	case *STFunctionSignatureNode:

		modifiedOpenParenToken, openParenTokenNode := replaceInner(n.OpenParenToken, target, replacement)

		modifiedParameters, parametersNode := replaceInner(n.Parameters, target, replacement)

		modifiedCloseParenToken, closeParenTokenNode := replaceInner(n.CloseParenToken, target, replacement)

		modifiedReturnTypeDesc, returnTypeDescNode := replaceInner(n.ReturnTypeDesc, target, replacement)

		modified := modifiedOpenParenToken || modifiedParameters || modifiedCloseParenToken || modifiedReturnTypeDesc
		if modified {
			return true, &STFunctionSignatureNode{

				STNode: n.STNode,

				OpenParenToken: openParenTokenNode,

				Parameters: parametersNode,

				CloseParenToken: closeParenTokenNode,

				ReturnTypeDesc: returnTypeDescNode,
			}
		}
		return false, current

	case *STExplicitAnonymousFunctionExpressionNode:

		modifiedAnnotations, annotationsNode := replaceInner(n.Annotations, target, replacement)

		modifiedQualifierList, qualifierListNode := replaceInner(n.QualifierList, target, replacement)

		modifiedFunctionKeyword, functionKeywordNode := replaceInner(n.FunctionKeyword, target, replacement)

		modifiedFunctionSignature, functionSignatureNode := replaceInner(n.FunctionSignature, target, replacement)

		modifiedFunctionBody, functionBodyNode := replaceInner(n.FunctionBody, target, replacement)

		modified := modifiedAnnotations || modifiedQualifierList || modifiedFunctionKeyword || modifiedFunctionSignature || modifiedFunctionBody
		if modified {
			return true, &STExplicitAnonymousFunctionExpressionNode{

				STAnonymousFunctionExpressionNode: n.STAnonymousFunctionExpressionNode,

				Annotations: annotationsNode,

				QualifierList: qualifierListNode,

				FunctionKeyword: functionKeywordNode,

				FunctionSignature: functionSignatureNode,

				FunctionBody: functionBodyNode,
			}
		}
		return false, current

	case *STExpressionFunctionBodyNode:

		modifiedRightDoubleArrow, rightDoubleArrowNode := replaceInner(n.RightDoubleArrow, target, replacement)

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modifiedSemicolon, semicolonNode := replaceInner(n.Semicolon, target, replacement)

		modified := modifiedRightDoubleArrow || modifiedExpression || modifiedSemicolon
		if modified {
			return true, &STExpressionFunctionBodyNode{

				STFunctionBodyNode: n.STFunctionBodyNode,

				RightDoubleArrow: rightDoubleArrowNode,

				Expression: expressionNode,

				Semicolon: semicolonNode,
			}
		}
		return false, current

	case *STTupleTypeDescriptorNode:

		modifiedOpenBracketToken, openBracketTokenNode := replaceInner(n.OpenBracketToken, target, replacement)

		modifiedMemberTypeDesc, memberTypeDescNode := replaceInner(n.MemberTypeDesc, target, replacement)

		modifiedCloseBracketToken, closeBracketTokenNode := replaceInner(n.CloseBracketToken, target, replacement)

		modified := modifiedOpenBracketToken || modifiedMemberTypeDesc || modifiedCloseBracketToken
		if modified {
			return true, &STTupleTypeDescriptorNode{

				STTypeDescriptorNode: n.STTypeDescriptorNode,

				OpenBracketToken: openBracketTokenNode,

				MemberTypeDesc: memberTypeDescNode,

				CloseBracketToken: closeBracketTokenNode,
			}
		}
		return false, current

	case *STParenthesisedTypeDescriptorNode:

		modifiedOpenParenToken, openParenTokenNode := replaceInner(n.OpenParenToken, target, replacement)

		modifiedTypedesc, typedescNode := replaceInner(n.Typedesc, target, replacement)

		modifiedCloseParenToken, closeParenTokenNode := replaceInner(n.CloseParenToken, target, replacement)

		modified := modifiedOpenParenToken || modifiedTypedesc || modifiedCloseParenToken
		if modified {
			return true, &STParenthesisedTypeDescriptorNode{

				STTypeDescriptorNode: n.STTypeDescriptorNode,

				OpenParenToken: openParenTokenNode,

				Typedesc: typedescNode,

				CloseParenToken: closeParenTokenNode,
			}
		}
		return false, current

	case *STExplicitNewExpressionNode:

		modifiedNewKeyword, newKeywordNode := replaceInner(n.NewKeyword, target, replacement)

		modifiedTypeDescriptor, typeDescriptorNode := replaceInner(n.TypeDescriptor, target, replacement)

		modifiedParenthesizedArgList, parenthesizedArgListNode := replaceInner(n.ParenthesizedArgList, target, replacement)

		modified := modifiedNewKeyword || modifiedTypeDescriptor || modifiedParenthesizedArgList
		if modified {
			return true, &STExplicitNewExpressionNode{

				STNewExpressionNode: n.STNewExpressionNode,

				NewKeyword: newKeywordNode,

				TypeDescriptor: typeDescriptorNode,

				ParenthesizedArgList: parenthesizedArgListNode,
			}
		}
		return false, current

	case *STImplicitNewExpressionNode:

		modifiedNewKeyword, newKeywordNode := replaceInner(n.NewKeyword, target, replacement)

		modifiedParenthesizedArgList, parenthesizedArgListNode := replaceInner(n.ParenthesizedArgList, target, replacement)

		modified := modifiedNewKeyword || modifiedParenthesizedArgList
		if modified {
			return true, &STImplicitNewExpressionNode{

				STNewExpressionNode: n.STNewExpressionNode,

				NewKeyword: newKeywordNode,

				ParenthesizedArgList: parenthesizedArgListNode,
			}
		}
		return false, current

	case *STParenthesizedArgList:

		modifiedOpenParenToken, openParenTokenNode := replaceInner(n.OpenParenToken, target, replacement)

		modifiedArguments, argumentsNode := replaceInner(n.Arguments, target, replacement)

		modifiedCloseParenToken, closeParenTokenNode := replaceInner(n.CloseParenToken, target, replacement)

		modified := modifiedOpenParenToken || modifiedArguments || modifiedCloseParenToken
		if modified {
			return true, &STParenthesizedArgList{

				STNode: n.STNode,

				OpenParenToken: openParenTokenNode,

				Arguments: argumentsNode,

				CloseParenToken: closeParenTokenNode,
			}
		}
		return false, current

	case *STQueryConstructTypeNode:

		modifiedKeyword, keywordNode := replaceInner(n.Keyword, target, replacement)

		modifiedKeySpecifier, keySpecifierNode := replaceInner(n.KeySpecifier, target, replacement)

		modified := modifiedKeyword || modifiedKeySpecifier
		if modified {
			return true, &STQueryConstructTypeNode{

				STNode: n.STNode,

				Keyword: keywordNode,

				KeySpecifier: keySpecifierNode,
			}
		}
		return false, current

	case *STFromClauseNode:

		modifiedFromKeyword, fromKeywordNode := replaceInner(n.FromKeyword, target, replacement)

		modifiedTypedBindingPattern, typedBindingPatternNode := replaceInner(n.TypedBindingPattern, target, replacement)

		modifiedInKeyword, inKeywordNode := replaceInner(n.InKeyword, target, replacement)

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modified := modifiedFromKeyword || modifiedTypedBindingPattern || modifiedInKeyword || modifiedExpression
		if modified {
			return true, &STFromClauseNode{

				STIntermediateClauseNode: n.STIntermediateClauseNode,

				FromKeyword: fromKeywordNode,

				TypedBindingPattern: typedBindingPatternNode,

				InKeyword: inKeywordNode,

				Expression: expressionNode,
			}
		}
		return false, current

	case *STWhereClauseNode:

		modifiedWhereKeyword, whereKeywordNode := replaceInner(n.WhereKeyword, target, replacement)

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modified := modifiedWhereKeyword || modifiedExpression
		if modified {
			return true, &STWhereClauseNode{

				STIntermediateClauseNode: n.STIntermediateClauseNode,

				WhereKeyword: whereKeywordNode,

				Expression: expressionNode,
			}
		}
		return false, current

	case *STLetClauseNode:

		modifiedLetKeyword, letKeywordNode := replaceInner(n.LetKeyword, target, replacement)

		modifiedLetVarDeclarations, letVarDeclarationsNode := replaceInner(n.LetVarDeclarations, target, replacement)

		modified := modifiedLetKeyword || modifiedLetVarDeclarations
		if modified {
			return true, &STLetClauseNode{

				STIntermediateClauseNode: n.STIntermediateClauseNode,

				LetKeyword: letKeywordNode,

				LetVarDeclarations: letVarDeclarationsNode,
			}
		}
		return false, current

	case *STJoinClauseNode:

		modifiedOuterKeyword, outerKeywordNode := replaceInner(n.OuterKeyword, target, replacement)

		modifiedJoinKeyword, joinKeywordNode := replaceInner(n.JoinKeyword, target, replacement)

		modifiedTypedBindingPattern, typedBindingPatternNode := replaceInner(n.TypedBindingPattern, target, replacement)

		modifiedInKeyword, inKeywordNode := replaceInner(n.InKeyword, target, replacement)

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modifiedJoinOnCondition, joinOnConditionNode := replaceInner(n.JoinOnCondition, target, replacement)

		modified := modifiedOuterKeyword || modifiedJoinKeyword || modifiedTypedBindingPattern || modifiedInKeyword || modifiedExpression || modifiedJoinOnCondition
		if modified {
			return true, &STJoinClauseNode{

				STIntermediateClauseNode: n.STIntermediateClauseNode,

				OuterKeyword: outerKeywordNode,

				JoinKeyword: joinKeywordNode,

				TypedBindingPattern: typedBindingPatternNode,

				InKeyword: inKeywordNode,

				Expression: expressionNode,

				JoinOnCondition: joinOnConditionNode,
			}
		}
		return false, current

	case *STOnClauseNode:

		modifiedOnKeyword, onKeywordNode := replaceInner(n.OnKeyword, target, replacement)

		modifiedLhsExpression, lhsExpressionNode := replaceInner(n.LhsExpression, target, replacement)

		modifiedEqualsKeyword, equalsKeywordNode := replaceInner(n.EqualsKeyword, target, replacement)

		modifiedRhsExpression, rhsExpressionNode := replaceInner(n.RhsExpression, target, replacement)

		modified := modifiedOnKeyword || modifiedLhsExpression || modifiedEqualsKeyword || modifiedRhsExpression
		if modified {
			return true, &STOnClauseNode{

				STClauseNode: n.STClauseNode,

				OnKeyword: onKeywordNode,

				LhsExpression: lhsExpressionNode,

				EqualsKeyword: equalsKeywordNode,

				RhsExpression: rhsExpressionNode,
			}
		}
		return false, current

	case *STLimitClauseNode:

		modifiedLimitKeyword, limitKeywordNode := replaceInner(n.LimitKeyword, target, replacement)

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modified := modifiedLimitKeyword || modifiedExpression
		if modified {
			return true, &STLimitClauseNode{

				STIntermediateClauseNode: n.STIntermediateClauseNode,

				LimitKeyword: limitKeywordNode,

				Expression: expressionNode,
			}
		}
		return false, current

	case *STOnConflictClauseNode:

		modifiedOnKeyword, onKeywordNode := replaceInner(n.OnKeyword, target, replacement)

		modifiedConflictKeyword, conflictKeywordNode := replaceInner(n.ConflictKeyword, target, replacement)

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modified := modifiedOnKeyword || modifiedConflictKeyword || modifiedExpression
		if modified {
			return true, &STOnConflictClauseNode{

				STClauseNode: n.STClauseNode,

				OnKeyword: onKeywordNode,

				ConflictKeyword: conflictKeywordNode,

				Expression: expressionNode,
			}
		}
		return false, current

	case *STQueryPipelineNode:

		modifiedFromClause, fromClauseNode := replaceInner(n.FromClause, target, replacement)

		modifiedIntermediateClauses, intermediateClausesNode := replaceInner(n.IntermediateClauses, target, replacement)

		modified := modifiedFromClause || modifiedIntermediateClauses
		if modified {
			return true, &STQueryPipelineNode{

				STNode: n.STNode,

				FromClause: fromClauseNode,

				IntermediateClauses: intermediateClausesNode,
			}
		}
		return false, current

	case *STSelectClauseNode:

		modifiedSelectKeyword, selectKeywordNode := replaceInner(n.SelectKeyword, target, replacement)

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modified := modifiedSelectKeyword || modifiedExpression
		if modified {
			return true, &STSelectClauseNode{

				STClauseNode: n.STClauseNode,

				SelectKeyword: selectKeywordNode,

				Expression: expressionNode,
			}
		}
		return false, current

	case *STCollectClauseNode:

		modifiedCollectKeyword, collectKeywordNode := replaceInner(n.CollectKeyword, target, replacement)

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modified := modifiedCollectKeyword || modifiedExpression
		if modified {
			return true, &STCollectClauseNode{

				STClauseNode: n.STClauseNode,

				CollectKeyword: collectKeywordNode,

				Expression: expressionNode,
			}
		}
		return false, current

	case *STQueryExpressionNode:

		modifiedQueryConstructType, queryConstructTypeNode := replaceInner(n.QueryConstructType, target, replacement)

		modifiedQueryPipeline, queryPipelineNode := replaceInner(n.QueryPipeline, target, replacement)

		modifiedResultClause, resultClauseNode := replaceInner(n.ResultClause, target, replacement)

		modifiedOnConflictClause, onConflictClauseNode := replaceInner(n.OnConflictClause, target, replacement)

		modified := modifiedQueryConstructType || modifiedQueryPipeline || modifiedResultClause || modifiedOnConflictClause
		if modified {
			return true, &STQueryExpressionNode{

				STExpressionNode: n.STExpressionNode,

				QueryConstructType: queryConstructTypeNode,

				QueryPipeline: queryPipelineNode,

				ResultClause: resultClauseNode,

				OnConflictClause: onConflictClauseNode,
			}
		}
		return false, current

	case *STQueryActionNode:

		modifiedQueryPipeline, queryPipelineNode := replaceInner(n.QueryPipeline, target, replacement)

		modifiedDoKeyword, doKeywordNode := replaceInner(n.DoKeyword, target, replacement)

		modifiedBlockStatement, blockStatementNode := replaceInner(n.BlockStatement, target, replacement)

		modified := modifiedQueryPipeline || modifiedDoKeyword || modifiedBlockStatement
		if modified {
			return true, &STQueryActionNode{

				STActionNode: n.STActionNode,

				QueryPipeline: queryPipelineNode,

				DoKeyword: doKeywordNode,

				BlockStatement: blockStatementNode,
			}
		}
		return false, current

	case *STIntersectionTypeDescriptorNode:

		modifiedLeftTypeDesc, leftTypeDescNode := replaceInner(n.LeftTypeDesc, target, replacement)

		modifiedBitwiseAndToken, bitwiseAndTokenNode := replaceInner(n.BitwiseAndToken, target, replacement)

		modifiedRightTypeDesc, rightTypeDescNode := replaceInner(n.RightTypeDesc, target, replacement)

		modified := modifiedLeftTypeDesc || modifiedBitwiseAndToken || modifiedRightTypeDesc
		if modified {
			return true, &STIntersectionTypeDescriptorNode{

				STTypeDescriptorNode: n.STTypeDescriptorNode,

				LeftTypeDesc: leftTypeDescNode,

				BitwiseAndToken: bitwiseAndTokenNode,

				RightTypeDesc: rightTypeDescNode,
			}
		}
		return false, current

	case *STImplicitAnonymousFunctionParameters:

		modifiedOpenParenToken, openParenTokenNode := replaceInner(n.OpenParenToken, target, replacement)

		modifiedParameters, parametersNode := replaceInner(n.Parameters, target, replacement)

		modifiedCloseParenToken, closeParenTokenNode := replaceInner(n.CloseParenToken, target, replacement)

		modified := modifiedOpenParenToken || modifiedParameters || modifiedCloseParenToken
		if modified {
			return true, &STImplicitAnonymousFunctionParameters{

				STNode: n.STNode,

				OpenParenToken: openParenTokenNode,

				Parameters: parametersNode,

				CloseParenToken: closeParenTokenNode,
			}
		}
		return false, current

	case *STImplicitAnonymousFunctionExpressionNode:

		modifiedParams, paramsNode := replaceInner(n.Params, target, replacement)

		modifiedRightDoubleArrow, rightDoubleArrowNode := replaceInner(n.RightDoubleArrow, target, replacement)

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modified := modifiedParams || modifiedRightDoubleArrow || modifiedExpression
		if modified {
			return true, &STImplicitAnonymousFunctionExpressionNode{

				STAnonymousFunctionExpressionNode: n.STAnonymousFunctionExpressionNode,

				Params: paramsNode,

				RightDoubleArrow: rightDoubleArrowNode,

				Expression: expressionNode,
			}
		}
		return false, current

	case *STStartActionNode:

		modifiedAnnotations, annotationsNode := replaceInner(n.Annotations, target, replacement)

		modifiedStartKeyword, startKeywordNode := replaceInner(n.StartKeyword, target, replacement)

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modified := modifiedAnnotations || modifiedStartKeyword || modifiedExpression
		if modified {
			return true, &STStartActionNode{

				STExpressionNode: n.STExpressionNode,

				Annotations: annotationsNode,

				StartKeyword: startKeywordNode,

				Expression: expressionNode,
			}
		}
		return false, current

	case *STFlushActionNode:

		modifiedFlushKeyword, flushKeywordNode := replaceInner(n.FlushKeyword, target, replacement)

		modifiedPeerWorker, peerWorkerNode := replaceInner(n.PeerWorker, target, replacement)

		modified := modifiedFlushKeyword || modifiedPeerWorker
		if modified {
			return true, &STFlushActionNode{

				STExpressionNode: n.STExpressionNode,

				FlushKeyword: flushKeywordNode,

				PeerWorker: peerWorkerNode,
			}
		}
		return false, current

	case *STSingletonTypeDescriptorNode:

		modifiedSimpleContExprNode, simpleContExprNodeNode := replaceInner(n.SimpleContExprNode, target, replacement)

		modified := modifiedSimpleContExprNode
		if modified {
			return true, &STSingletonTypeDescriptorNode{

				STTypeDescriptorNode: n.STTypeDescriptorNode,

				SimpleContExprNode: simpleContExprNodeNode,
			}
		}
		return false, current

	case *STMethodDeclarationNode:

		modifiedMetadata, metadataNode := replaceInner(n.Metadata, target, replacement)

		modifiedQualifierList, qualifierListNode := replaceInner(n.QualifierList, target, replacement)

		modifiedFunctionKeyword, functionKeywordNode := replaceInner(n.FunctionKeyword, target, replacement)

		modifiedMethodName, methodNameNode := replaceInner(n.MethodName, target, replacement)

		modifiedRelativeResourcePath, relativeResourcePathNode := replaceInner(n.RelativeResourcePath, target, replacement)

		modifiedMethodSignature, methodSignatureNode := replaceInner(n.MethodSignature, target, replacement)

		modifiedSemicolon, semicolonNode := replaceInner(n.Semicolon, target, replacement)

		modified := modifiedMetadata || modifiedQualifierList || modifiedFunctionKeyword || modifiedMethodName || modifiedRelativeResourcePath || modifiedMethodSignature || modifiedSemicolon
		if modified {
			return true, &STMethodDeclarationNode{

				STNode: n.STNode,

				Metadata: metadataNode,

				QualifierList: qualifierListNode,

				FunctionKeyword: functionKeywordNode,

				MethodName: methodNameNode,

				RelativeResourcePath: relativeResourcePathNode,

				MethodSignature: methodSignatureNode,

				Semicolon: semicolonNode,
			}
		}
		return false, current

	case *STTypedBindingPatternNode:

		modifiedTypeDescriptor, typeDescriptorNode := replaceInner(n.TypeDescriptor, target, replacement)

		modifiedBindingPattern, bindingPatternNode := replaceInner(n.BindingPattern, target, replacement)

		modified := modifiedTypeDescriptor || modifiedBindingPattern
		if modified {
			return true, &STTypedBindingPatternNode{

				STNode: n.STNode,

				TypeDescriptor: typeDescriptorNode,

				BindingPattern: bindingPatternNode,
			}
		}
		return false, current

	case *STCaptureBindingPatternNode:

		modifiedVariableName, variableNameNode := replaceInner(n.VariableName, target, replacement)

		modified := modifiedVariableName
		if modified {
			return true, &STCaptureBindingPatternNode{

				STBindingPatternNode: n.STBindingPatternNode,

				VariableName: variableNameNode,
			}
		}
		return false, current

	case *STWildcardBindingPatternNode:

		modifiedUnderscoreToken, underscoreTokenNode := replaceInner(n.UnderscoreToken, target, replacement)

		modified := modifiedUnderscoreToken
		if modified {
			return true, &STWildcardBindingPatternNode{

				STBindingPatternNode: n.STBindingPatternNode,

				UnderscoreToken: underscoreTokenNode,
			}
		}
		return false, current

	case *STListBindingPatternNode:

		modifiedOpenBracket, openBracketNode := replaceInner(n.OpenBracket, target, replacement)

		modifiedBindingPatterns, bindingPatternsNode := replaceInner(n.BindingPatterns, target, replacement)

		modifiedCloseBracket, closeBracketNode := replaceInner(n.CloseBracket, target, replacement)

		modified := modifiedOpenBracket || modifiedBindingPatterns || modifiedCloseBracket
		if modified {
			return true, &STListBindingPatternNode{

				STBindingPatternNode: n.STBindingPatternNode,

				OpenBracket: openBracketNode,

				BindingPatterns: bindingPatternsNode,

				CloseBracket: closeBracketNode,
			}
		}
		return false, current

	case *STMappingBindingPatternNode:

		modifiedOpenBrace, openBraceNode := replaceInner(n.OpenBrace, target, replacement)

		modifiedFieldBindingPatterns, fieldBindingPatternsNode := replaceInner(n.FieldBindingPatterns, target, replacement)

		modifiedCloseBrace, closeBraceNode := replaceInner(n.CloseBrace, target, replacement)

		modified := modifiedOpenBrace || modifiedFieldBindingPatterns || modifiedCloseBrace
		if modified {
			return true, &STMappingBindingPatternNode{

				STBindingPatternNode: n.STBindingPatternNode,

				OpenBrace: openBraceNode,

				FieldBindingPatterns: fieldBindingPatternsNode,

				CloseBrace: closeBraceNode,
			}
		}
		return false, current

	case *STFieldBindingPatternFullNode:

		modifiedVariableName, variableNameNode := replaceInner(n.VariableName, target, replacement)

		modifiedColon, colonNode := replaceInner(n.Colon, target, replacement)

		modifiedBindingPattern, bindingPatternNode := replaceInner(n.BindingPattern, target, replacement)

		modified := modifiedVariableName || modifiedColon || modifiedBindingPattern
		if modified {
			return true, &STFieldBindingPatternFullNode{

				STFieldBindingPatternNode: n.STFieldBindingPatternNode,

				VariableName: variableNameNode,

				Colon: colonNode,

				BindingPattern: bindingPatternNode,
			}
		}
		return false, current

	case *STFieldBindingPatternVarnameNode:

		modifiedVariableName, variableNameNode := replaceInner(n.VariableName, target, replacement)

		modified := modifiedVariableName
		if modified {
			return true, &STFieldBindingPatternVarnameNode{

				STFieldBindingPatternNode: n.STFieldBindingPatternNode,

				VariableName: variableNameNode,
			}
		}
		return false, current

	case *STRestBindingPatternNode:

		modifiedEllipsisToken, ellipsisTokenNode := replaceInner(n.EllipsisToken, target, replacement)

		modifiedVariableName, variableNameNode := replaceInner(n.VariableName, target, replacement)

		modified := modifiedEllipsisToken || modifiedVariableName
		if modified {
			return true, &STRestBindingPatternNode{

				STBindingPatternNode: n.STBindingPatternNode,

				EllipsisToken: ellipsisTokenNode,

				VariableName: variableNameNode,
			}
		}
		return false, current

	case *STErrorBindingPatternNode:

		modifiedErrorKeyword, errorKeywordNode := replaceInner(n.ErrorKeyword, target, replacement)

		modifiedTypeReference, typeReferenceNode := replaceInner(n.TypeReference, target, replacement)

		modifiedOpenParenthesis, openParenthesisNode := replaceInner(n.OpenParenthesis, target, replacement)

		modifiedArgListBindingPatterns, argListBindingPatternsNode := replaceInner(n.ArgListBindingPatterns, target, replacement)

		modifiedCloseParenthesis, closeParenthesisNode := replaceInner(n.CloseParenthesis, target, replacement)

		modified := modifiedErrorKeyword || modifiedTypeReference || modifiedOpenParenthesis || modifiedArgListBindingPatterns || modifiedCloseParenthesis
		if modified {
			return true, &STErrorBindingPatternNode{

				STBindingPatternNode: n.STBindingPatternNode,

				ErrorKeyword: errorKeywordNode,

				TypeReference: typeReferenceNode,

				OpenParenthesis: openParenthesisNode,

				ArgListBindingPatterns: argListBindingPatternsNode,

				CloseParenthesis: closeParenthesisNode,
			}
		}
		return false, current

	case *STNamedArgBindingPatternNode:

		modifiedArgName, argNameNode := replaceInner(n.ArgName, target, replacement)

		modifiedEqualsToken, equalsTokenNode := replaceInner(n.EqualsToken, target, replacement)

		modifiedBindingPattern, bindingPatternNode := replaceInner(n.BindingPattern, target, replacement)

		modified := modifiedArgName || modifiedEqualsToken || modifiedBindingPattern
		if modified {
			return true, &STNamedArgBindingPatternNode{

				STBindingPatternNode: n.STBindingPatternNode,

				ArgName: argNameNode,

				EqualsToken: equalsTokenNode,

				BindingPattern: bindingPatternNode,
			}
		}
		return false, current

	case *STAsyncSendActionNode:

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modifiedRightArrowToken, rightArrowTokenNode := replaceInner(n.RightArrowToken, target, replacement)

		modifiedPeerWorker, peerWorkerNode := replaceInner(n.PeerWorker, target, replacement)

		modified := modifiedExpression || modifiedRightArrowToken || modifiedPeerWorker
		if modified {
			return true, &STAsyncSendActionNode{

				STActionNode: n.STActionNode,

				Expression: expressionNode,

				RightArrowToken: rightArrowTokenNode,

				PeerWorker: peerWorkerNode,
			}
		}
		return false, current

	case *STSyncSendActionNode:

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modifiedSyncSendToken, syncSendTokenNode := replaceInner(n.SyncSendToken, target, replacement)

		modifiedPeerWorker, peerWorkerNode := replaceInner(n.PeerWorker, target, replacement)

		modified := modifiedExpression || modifiedSyncSendToken || modifiedPeerWorker
		if modified {
			return true, &STSyncSendActionNode{

				STActionNode: n.STActionNode,

				Expression: expressionNode,

				SyncSendToken: syncSendTokenNode,

				PeerWorker: peerWorkerNode,
			}
		}
		return false, current

	case *STReceiveActionNode:

		modifiedLeftArrow, leftArrowNode := replaceInner(n.LeftArrow, target, replacement)

		modifiedReceiveWorkers, receiveWorkersNode := replaceInner(n.ReceiveWorkers, target, replacement)

		modified := modifiedLeftArrow || modifiedReceiveWorkers
		if modified {
			return true, &STReceiveActionNode{

				STActionNode: n.STActionNode,

				LeftArrow: leftArrowNode,

				ReceiveWorkers: receiveWorkersNode,
			}
		}
		return false, current

	case *STReceiveFieldsNode:

		modifiedOpenBrace, openBraceNode := replaceInner(n.OpenBrace, target, replacement)

		modifiedReceiveFields, receiveFieldsNode := replaceInner(n.ReceiveFields, target, replacement)

		modifiedCloseBrace, closeBraceNode := replaceInner(n.CloseBrace, target, replacement)

		modified := modifiedOpenBrace || modifiedReceiveFields || modifiedCloseBrace
		if modified {
			return true, &STReceiveFieldsNode{

				STNode: n.STNode,

				OpenBrace: openBraceNode,

				ReceiveFields: receiveFieldsNode,

				CloseBrace: closeBraceNode,
			}
		}
		return false, current

	case *STAlternateReceiveNode:

		modifiedWorkers, workersNode := replaceInner(n.Workers, target, replacement)

		modified := modifiedWorkers
		if modified {
			return true, &STAlternateReceiveNode{

				STNode: n.STNode,

				Workers: workersNode,
			}
		}
		return false, current

	case *STRestDescriptorNode:

		modifiedTypeDescriptor, typeDescriptorNode := replaceInner(n.TypeDescriptor, target, replacement)

		modifiedEllipsisToken, ellipsisTokenNode := replaceInner(n.EllipsisToken, target, replacement)

		modified := modifiedTypeDescriptor || modifiedEllipsisToken
		if modified {
			return true, &STRestDescriptorNode{

				STNode: n.STNode,

				TypeDescriptor: typeDescriptorNode,

				EllipsisToken: ellipsisTokenNode,
			}
		}
		return false, current

	case *STDoubleGTTokenNode:

		modifiedOpenGTToken, openGTTokenNode := replaceInner(n.OpenGTToken, target, replacement)

		modifiedEndGTToken, endGTTokenNode := replaceInner(n.EndGTToken, target, replacement)

		modified := modifiedOpenGTToken || modifiedEndGTToken
		if modified {
			return true, &STDoubleGTTokenNode{

				STNode: n.STNode,

				OpenGTToken: openGTTokenNode,

				EndGTToken: endGTTokenNode,
			}
		}
		return false, current

	case *STTrippleGTTokenNode:

		modifiedOpenGTToken, openGTTokenNode := replaceInner(n.OpenGTToken, target, replacement)

		modifiedMiddleGTToken, middleGTTokenNode := replaceInner(n.MiddleGTToken, target, replacement)

		modifiedEndGTToken, endGTTokenNode := replaceInner(n.EndGTToken, target, replacement)

		modified := modifiedOpenGTToken || modifiedMiddleGTToken || modifiedEndGTToken
		if modified {
			return true, &STTrippleGTTokenNode{

				STNode: n.STNode,

				OpenGTToken: openGTTokenNode,

				MiddleGTToken: middleGTTokenNode,

				EndGTToken: endGTTokenNode,
			}
		}
		return false, current

	case *STWaitActionNode:

		modifiedWaitKeyword, waitKeywordNode := replaceInner(n.WaitKeyword, target, replacement)

		modifiedWaitFutureExpr, waitFutureExprNode := replaceInner(n.WaitFutureExpr, target, replacement)

		modified := modifiedWaitKeyword || modifiedWaitFutureExpr
		if modified {
			return true, &STWaitActionNode{

				STActionNode: n.STActionNode,

				WaitKeyword: waitKeywordNode,

				WaitFutureExpr: waitFutureExprNode,
			}
		}
		return false, current

	case *STWaitFieldsListNode:

		modifiedOpenBrace, openBraceNode := replaceInner(n.OpenBrace, target, replacement)

		modifiedWaitFields, waitFieldsNode := replaceInner(n.WaitFields, target, replacement)

		modifiedCloseBrace, closeBraceNode := replaceInner(n.CloseBrace, target, replacement)

		modified := modifiedOpenBrace || modifiedWaitFields || modifiedCloseBrace
		if modified {
			return true, &STWaitFieldsListNode{

				STNode: n.STNode,

				OpenBrace: openBraceNode,

				WaitFields: waitFieldsNode,

				CloseBrace: closeBraceNode,
			}
		}
		return false, current

	case *STWaitFieldNode:

		modifiedFieldName, fieldNameNode := replaceInner(n.FieldName, target, replacement)

		modifiedColon, colonNode := replaceInner(n.Colon, target, replacement)

		modifiedWaitFutureExpr, waitFutureExprNode := replaceInner(n.WaitFutureExpr, target, replacement)

		modified := modifiedFieldName || modifiedColon || modifiedWaitFutureExpr
		if modified {
			return true, &STWaitFieldNode{

				STNode: n.STNode,

				FieldName: fieldNameNode,

				Colon: colonNode,

				WaitFutureExpr: waitFutureExprNode,
			}
		}
		return false, current

	case *STAnnotAccessExpressionNode:

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modifiedAnnotChainingToken, annotChainingTokenNode := replaceInner(n.AnnotChainingToken, target, replacement)

		modifiedAnnotTagReference, annotTagReferenceNode := replaceInner(n.AnnotTagReference, target, replacement)

		modified := modifiedExpression || modifiedAnnotChainingToken || modifiedAnnotTagReference
		if modified {
			return true, &STAnnotAccessExpressionNode{

				STExpressionNode: n.STExpressionNode,

				Expression: expressionNode,

				AnnotChainingToken: annotChainingTokenNode,

				AnnotTagReference: annotTagReferenceNode,
			}
		}
		return false, current

	case *STOptionalFieldAccessExpressionNode:

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modifiedOptionalChainingToken, optionalChainingTokenNode := replaceInner(n.OptionalChainingToken, target, replacement)

		modifiedFieldName, fieldNameNode := replaceInner(n.FieldName, target, replacement)

		modified := modifiedExpression || modifiedOptionalChainingToken || modifiedFieldName
		if modified {
			return true, &STOptionalFieldAccessExpressionNode{

				STExpressionNode: n.STExpressionNode,

				Expression: expressionNode,

				OptionalChainingToken: optionalChainingTokenNode,

				FieldName: fieldNameNode,
			}
		}
		return false, current

	case *STConditionalExpressionNode:

		modifiedLhsExpression, lhsExpressionNode := replaceInner(n.LhsExpression, target, replacement)

		modifiedQuestionMarkToken, questionMarkTokenNode := replaceInner(n.QuestionMarkToken, target, replacement)

		modifiedMiddleExpression, middleExpressionNode := replaceInner(n.MiddleExpression, target, replacement)

		modifiedColonToken, colonTokenNode := replaceInner(n.ColonToken, target, replacement)

		modifiedEndExpression, endExpressionNode := replaceInner(n.EndExpression, target, replacement)

		modified := modifiedLhsExpression || modifiedQuestionMarkToken || modifiedMiddleExpression || modifiedColonToken || modifiedEndExpression
		if modified {
			return true, &STConditionalExpressionNode{

				STExpressionNode: n.STExpressionNode,

				LhsExpression: lhsExpressionNode,

				QuestionMarkToken: questionMarkTokenNode,

				MiddleExpression: middleExpressionNode,

				ColonToken: colonTokenNode,

				EndExpression: endExpressionNode,
			}
		}
		return false, current

	case *STEnumDeclarationNode:

		modifiedMetadata, metadataNode := replaceInner(n.Metadata, target, replacement)

		modifiedQualifier, qualifierNode := replaceInner(n.Qualifier, target, replacement)

		modifiedEnumKeywordToken, enumKeywordTokenNode := replaceInner(n.EnumKeywordToken, target, replacement)

		modifiedIdentifier, identifierNode := replaceInner(n.Identifier, target, replacement)

		modifiedOpenBraceToken, openBraceTokenNode := replaceInner(n.OpenBraceToken, target, replacement)

		modifiedEnumMemberList, enumMemberListNode := replaceInner(n.EnumMemberList, target, replacement)

		modifiedCloseBraceToken, closeBraceTokenNode := replaceInner(n.CloseBraceToken, target, replacement)

		modifiedSemicolonToken, semicolonTokenNode := replaceInner(n.SemicolonToken, target, replacement)

		modified := modifiedMetadata || modifiedQualifier || modifiedEnumKeywordToken || modifiedIdentifier || modifiedOpenBraceToken || modifiedEnumMemberList || modifiedCloseBraceToken || modifiedSemicolonToken
		if modified {
			return true, &STEnumDeclarationNode{

				STModuleMemberDeclarationNode: n.STModuleMemberDeclarationNode,

				Metadata: metadataNode,

				Qualifier: qualifierNode,

				EnumKeywordToken: enumKeywordTokenNode,

				Identifier: identifierNode,

				OpenBraceToken: openBraceTokenNode,

				EnumMemberList: enumMemberListNode,

				CloseBraceToken: closeBraceTokenNode,

				SemicolonToken: semicolonTokenNode,
			}
		}
		return false, current

	case *STEnumMemberNode:

		modifiedMetadata, metadataNode := replaceInner(n.Metadata, target, replacement)

		modifiedIdentifier, identifierNode := replaceInner(n.Identifier, target, replacement)

		modifiedEqualToken, equalTokenNode := replaceInner(n.EqualToken, target, replacement)

		modifiedConstExprNode, constExprNodeNode := replaceInner(n.ConstExprNode, target, replacement)

		modified := modifiedMetadata || modifiedIdentifier || modifiedEqualToken || modifiedConstExprNode
		if modified {
			return true, &STEnumMemberNode{

				STNode: n.STNode,

				Metadata: metadataNode,

				Identifier: identifierNode,

				EqualToken: equalTokenNode,

				ConstExprNode: constExprNodeNode,
			}
		}
		return false, current

	case *STArrayTypeDescriptorNode:

		modifiedMemberTypeDesc, memberTypeDescNode := replaceInner(n.MemberTypeDesc, target, replacement)

		modifiedDimensions, dimensionsNode := replaceInner(n.Dimensions, target, replacement)

		modified := modifiedMemberTypeDesc || modifiedDimensions
		if modified {
			return true, &STArrayTypeDescriptorNode{

				STTypeDescriptorNode: n.STTypeDescriptorNode,

				MemberTypeDesc: memberTypeDescNode,

				Dimensions: dimensionsNode,
			}
		}
		return false, current

	case *STArrayDimensionNode:

		modifiedOpenBracket, openBracketNode := replaceInner(n.OpenBracket, target, replacement)

		modifiedArrayLength, arrayLengthNode := replaceInner(n.ArrayLength, target, replacement)

		modifiedCloseBracket, closeBracketNode := replaceInner(n.CloseBracket, target, replacement)

		modified := modifiedOpenBracket || modifiedArrayLength || modifiedCloseBracket
		if modified {
			return true, &STArrayDimensionNode{

				STNode: n.STNode,

				OpenBracket: openBracketNode,

				ArrayLength: arrayLengthNode,

				CloseBracket: closeBracketNode,
			}
		}
		return false, current

	case *STTransactionStatementNode:

		modifiedTransactionKeyword, transactionKeywordNode := replaceInner(n.TransactionKeyword, target, replacement)

		modifiedBlockStatement, blockStatementNode := replaceInner(n.BlockStatement, target, replacement)

		modifiedOnFailClause, onFailClauseNode := replaceInner(n.OnFailClause, target, replacement)

		modified := modifiedTransactionKeyword || modifiedBlockStatement || modifiedOnFailClause
		if modified {
			return true, &STTransactionStatementNode{

				STStatementNode: n.STStatementNode,

				TransactionKeyword: transactionKeywordNode,

				BlockStatement: blockStatementNode,

				OnFailClause: onFailClauseNode,
			}
		}
		return false, current

	case *STRollbackStatementNode:

		modifiedRollbackKeyword, rollbackKeywordNode := replaceInner(n.RollbackKeyword, target, replacement)

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modifiedSemicolon, semicolonNode := replaceInner(n.Semicolon, target, replacement)

		modified := modifiedRollbackKeyword || modifiedExpression || modifiedSemicolon
		if modified {
			return true, &STRollbackStatementNode{

				STStatementNode: n.STStatementNode,

				RollbackKeyword: rollbackKeywordNode,

				Expression: expressionNode,

				Semicolon: semicolonNode,
			}
		}
		return false, current

	case *STRetryStatementNode:

		modifiedRetryKeyword, retryKeywordNode := replaceInner(n.RetryKeyword, target, replacement)

		modifiedTypeParameter, typeParameterNode := replaceInner(n.TypeParameter, target, replacement)

		modifiedArguments, argumentsNode := replaceInner(n.Arguments, target, replacement)

		modifiedRetryBody, retryBodyNode := replaceInner(n.RetryBody, target, replacement)

		modifiedOnFailClause, onFailClauseNode := replaceInner(n.OnFailClause, target, replacement)

		modified := modifiedRetryKeyword || modifiedTypeParameter || modifiedArguments || modifiedRetryBody || modifiedOnFailClause
		if modified {
			return true, &STRetryStatementNode{

				STStatementNode: n.STStatementNode,

				RetryKeyword: retryKeywordNode,

				TypeParameter: typeParameterNode,

				Arguments: argumentsNode,

				RetryBody: retryBodyNode,

				OnFailClause: onFailClauseNode,
			}
		}
		return false, current

	case *STCommitActionNode:

		modifiedCommitKeyword, commitKeywordNode := replaceInner(n.CommitKeyword, target, replacement)

		modified := modifiedCommitKeyword
		if modified {
			return true, &STCommitActionNode{

				STActionNode: n.STActionNode,

				CommitKeyword: commitKeywordNode,
			}
		}
		return false, current

	case *STTransactionalExpressionNode:

		modifiedTransactionalKeyword, transactionalKeywordNode := replaceInner(n.TransactionalKeyword, target, replacement)

		modified := modifiedTransactionalKeyword
		if modified {
			return true, &STTransactionalExpressionNode{

				STExpressionNode: n.STExpressionNode,

				TransactionalKeyword: transactionalKeywordNode,
			}
		}
		return false, current

	case *STByteArrayLiteralNode:

		modifiedType, typeNode := replaceInner(n.Type, target, replacement)

		modifiedStartBacktick, startBacktickNode := replaceInner(n.StartBacktick, target, replacement)

		modifiedContent, contentNode := replaceInner(n.Content, target, replacement)

		modifiedEndBacktick, endBacktickNode := replaceInner(n.EndBacktick, target, replacement)

		modified := modifiedType || modifiedStartBacktick || modifiedContent || modifiedEndBacktick
		if modified {
			return true, &STByteArrayLiteralNode{

				STExpressionNode: n.STExpressionNode,

				Type: typeNode,

				StartBacktick: startBacktickNode,

				Content: contentNode,

				EndBacktick: endBacktickNode,
			}
		}
		return false, current

	case *STXMLFilterExpressionNode:

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modifiedXmlPatternChain, xmlPatternChainNode := replaceInner(n.XmlPatternChain, target, replacement)

		modified := modifiedExpression || modifiedXmlPatternChain
		if modified {
			return true, &STXMLFilterExpressionNode{

				STXMLNavigateExpressionNode: n.STXMLNavigateExpressionNode,

				Expression: expressionNode,

				XmlPatternChain: xmlPatternChainNode,
			}
		}
		return false, current

	case *STXMLStepExpressionNode:

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modifiedXmlStepStart, xmlStepStartNode := replaceInner(n.XmlStepStart, target, replacement)

		modifiedXmlStepExtend, xmlStepExtendNode := replaceInner(n.XmlStepExtend, target, replacement)

		modified := modifiedExpression || modifiedXmlStepStart || modifiedXmlStepExtend
		if modified {
			return true, &STXMLStepExpressionNode{

				STXMLNavigateExpressionNode: n.STXMLNavigateExpressionNode,

				Expression: expressionNode,

				XmlStepStart: xmlStepStartNode,

				XmlStepExtend: xmlStepExtendNode,
			}
		}
		return false, current

	case *STXMLNamePatternChainingNode:

		modifiedStartToken, startTokenNode := replaceInner(n.StartToken, target, replacement)

		modifiedXmlNamePattern, xmlNamePatternNode := replaceInner(n.XmlNamePattern, target, replacement)

		modifiedGtToken, gtTokenNode := replaceInner(n.GtToken, target, replacement)

		modified := modifiedStartToken || modifiedXmlNamePattern || modifiedGtToken
		if modified {
			return true, &STXMLNamePatternChainingNode{

				STNode: n.STNode,

				StartToken: startTokenNode,

				XmlNamePattern: xmlNamePatternNode,

				GtToken: gtTokenNode,
			}
		}
		return false, current

	case *STXMLStepIndexedExtendNode:

		modifiedOpenBracket, openBracketNode := replaceInner(n.OpenBracket, target, replacement)

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modifiedCloseBracket, closeBracketNode := replaceInner(n.CloseBracket, target, replacement)

		modified := modifiedOpenBracket || modifiedExpression || modifiedCloseBracket
		if modified {
			return true, &STXMLStepIndexedExtendNode{

				STNode: n.STNode,

				OpenBracket: openBracketNode,

				Expression: expressionNode,

				CloseBracket: closeBracketNode,
			}
		}
		return false, current

	case *STXMLStepMethodCallExtendNode:

		modifiedDotToken, dotTokenNode := replaceInner(n.DotToken, target, replacement)

		modifiedMethodName, methodNameNode := replaceInner(n.MethodName, target, replacement)

		modifiedParenthesizedArgList, parenthesizedArgListNode := replaceInner(n.ParenthesizedArgList, target, replacement)

		modified := modifiedDotToken || modifiedMethodName || modifiedParenthesizedArgList
		if modified {
			return true, &STXMLStepMethodCallExtendNode{

				STNode: n.STNode,

				DotToken: dotTokenNode,

				MethodName: methodNameNode,

				ParenthesizedArgList: parenthesizedArgListNode,
			}
		}
		return false, current

	case *STXMLAtomicNamePatternNode:

		modifiedPrefix, prefixNode := replaceInner(n.Prefix, target, replacement)

		modifiedColon, colonNode := replaceInner(n.Colon, target, replacement)

		modifiedName, nameNode := replaceInner(n.Name, target, replacement)

		modified := modifiedPrefix || modifiedColon || modifiedName
		if modified {
			return true, &STXMLAtomicNamePatternNode{

				STNode: n.STNode,

				Prefix: prefixNode,

				Colon: colonNode,

				Name: nameNode,
			}
		}
		return false, current

	case *STTypeReferenceTypeDescNode:

		modifiedTypeRef, typeRefNode := replaceInner(n.TypeRef, target, replacement)

		modified := modifiedTypeRef
		if modified {
			return true, &STTypeReferenceTypeDescNode{

				STTypeDescriptorNode: n.STTypeDescriptorNode,

				TypeRef: typeRefNode,
			}
		}
		return false, current

	case *STMatchStatementNode:

		modifiedMatchKeyword, matchKeywordNode := replaceInner(n.MatchKeyword, target, replacement)

		modifiedCondition, conditionNode := replaceInner(n.Condition, target, replacement)

		modifiedOpenBrace, openBraceNode := replaceInner(n.OpenBrace, target, replacement)

		modifiedMatchClauses, matchClausesNode := replaceInner(n.MatchClauses, target, replacement)

		modifiedCloseBrace, closeBraceNode := replaceInner(n.CloseBrace, target, replacement)

		modifiedOnFailClause, onFailClauseNode := replaceInner(n.OnFailClause, target, replacement)

		modified := modifiedMatchKeyword || modifiedCondition || modifiedOpenBrace || modifiedMatchClauses || modifiedCloseBrace || modifiedOnFailClause
		if modified {
			return true, &STMatchStatementNode{

				STStatementNode: n.STStatementNode,

				MatchKeyword: matchKeywordNode,

				Condition: conditionNode,

				OpenBrace: openBraceNode,

				MatchClauses: matchClausesNode,

				CloseBrace: closeBraceNode,

				OnFailClause: onFailClauseNode,
			}
		}
		return false, current

	case *STMatchClauseNode:

		modifiedMatchPatterns, matchPatternsNode := replaceInner(n.MatchPatterns, target, replacement)

		modifiedMatchGuard, matchGuardNode := replaceInner(n.MatchGuard, target, replacement)

		modifiedRightDoubleArrow, rightDoubleArrowNode := replaceInner(n.RightDoubleArrow, target, replacement)

		modifiedBlockStatement, blockStatementNode := replaceInner(n.BlockStatement, target, replacement)

		modified := modifiedMatchPatterns || modifiedMatchGuard || modifiedRightDoubleArrow || modifiedBlockStatement
		if modified {
			return true, &STMatchClauseNode{

				STNode: n.STNode,

				MatchPatterns: matchPatternsNode,

				MatchGuard: matchGuardNode,

				RightDoubleArrow: rightDoubleArrowNode,

				BlockStatement: blockStatementNode,
			}
		}
		return false, current

	case *STMatchGuardNode:

		modifiedIfKeyword, ifKeywordNode := replaceInner(n.IfKeyword, target, replacement)

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modified := modifiedIfKeyword || modifiedExpression
		if modified {
			return true, &STMatchGuardNode{

				STNode: n.STNode,

				IfKeyword: ifKeywordNode,

				Expression: expressionNode,
			}
		}
		return false, current

	case *STDistinctTypeDescriptorNode:

		modifiedDistinctKeyword, distinctKeywordNode := replaceInner(n.DistinctKeyword, target, replacement)

		modifiedTypeDescriptor, typeDescriptorNode := replaceInner(n.TypeDescriptor, target, replacement)

		modified := modifiedDistinctKeyword || modifiedTypeDescriptor
		if modified {
			return true, &STDistinctTypeDescriptorNode{

				STTypeDescriptorNode: n.STTypeDescriptorNode,

				DistinctKeyword: distinctKeywordNode,

				TypeDescriptor: typeDescriptorNode,
			}
		}
		return false, current

	case *STListMatchPatternNode:

		modifiedOpenBracket, openBracketNode := replaceInner(n.OpenBracket, target, replacement)

		modifiedMatchPatterns, matchPatternsNode := replaceInner(n.MatchPatterns, target, replacement)

		modifiedCloseBracket, closeBracketNode := replaceInner(n.CloseBracket, target, replacement)

		modified := modifiedOpenBracket || modifiedMatchPatterns || modifiedCloseBracket
		if modified {
			return true, &STListMatchPatternNode{

				STNode: n.STNode,

				OpenBracket: openBracketNode,

				MatchPatterns: matchPatternsNode,

				CloseBracket: closeBracketNode,
			}
		}
		return false, current

	case *STRestMatchPatternNode:

		modifiedEllipsisToken, ellipsisTokenNode := replaceInner(n.EllipsisToken, target, replacement)

		modifiedVarKeywordToken, varKeywordTokenNode := replaceInner(n.VarKeywordToken, target, replacement)

		modifiedVariableName, variableNameNode := replaceInner(n.VariableName, target, replacement)

		modified := modifiedEllipsisToken || modifiedVarKeywordToken || modifiedVariableName
		if modified {
			return true, &STRestMatchPatternNode{

				STNode: n.STNode,

				EllipsisToken: ellipsisTokenNode,

				VarKeywordToken: varKeywordTokenNode,

				VariableName: variableNameNode,
			}
		}
		return false, current

	case *STMappingMatchPatternNode:

		modifiedOpenBraceToken, openBraceTokenNode := replaceInner(n.OpenBraceToken, target, replacement)

		modifiedFieldMatchPatterns, fieldMatchPatternsNode := replaceInner(n.FieldMatchPatterns, target, replacement)

		modifiedCloseBraceToken, closeBraceTokenNode := replaceInner(n.CloseBraceToken, target, replacement)

		modified := modifiedOpenBraceToken || modifiedFieldMatchPatterns || modifiedCloseBraceToken
		if modified {
			return true, &STMappingMatchPatternNode{

				STNode: n.STNode,

				OpenBraceToken: openBraceTokenNode,

				FieldMatchPatterns: fieldMatchPatternsNode,

				CloseBraceToken: closeBraceTokenNode,
			}
		}
		return false, current

	case *STFieldMatchPatternNode:

		modifiedFieldNameNode, fieldNameNodeNode := replaceInner(n.FieldNameNode, target, replacement)

		modifiedColonToken, colonTokenNode := replaceInner(n.ColonToken, target, replacement)

		modifiedMatchPattern, matchPatternNode := replaceInner(n.MatchPattern, target, replacement)

		modified := modifiedFieldNameNode || modifiedColonToken || modifiedMatchPattern
		if modified {
			return true, &STFieldMatchPatternNode{

				STNode: n.STNode,

				FieldNameNode: fieldNameNodeNode,

				ColonToken: colonTokenNode,

				MatchPattern: matchPatternNode,
			}
		}
		return false, current

	case *STErrorMatchPatternNode:

		modifiedErrorKeyword, errorKeywordNode := replaceInner(n.ErrorKeyword, target, replacement)

		modifiedTypeReference, typeReferenceNode := replaceInner(n.TypeReference, target, replacement)

		modifiedOpenParenthesisToken, openParenthesisTokenNode := replaceInner(n.OpenParenthesisToken, target, replacement)

		modifiedArgListMatchPatternNode, argListMatchPatternNodeNode := replaceInner(n.ArgListMatchPatternNode, target, replacement)

		modifiedCloseParenthesisToken, closeParenthesisTokenNode := replaceInner(n.CloseParenthesisToken, target, replacement)

		modified := modifiedErrorKeyword || modifiedTypeReference || modifiedOpenParenthesisToken || modifiedArgListMatchPatternNode || modifiedCloseParenthesisToken
		if modified {
			return true, &STErrorMatchPatternNode{

				STNode: n.STNode,

				ErrorKeyword: errorKeywordNode,

				TypeReference: typeReferenceNode,

				OpenParenthesisToken: openParenthesisTokenNode,

				ArgListMatchPatternNode: argListMatchPatternNodeNode,

				CloseParenthesisToken: closeParenthesisTokenNode,
			}
		}
		return false, current

	case *STNamedArgMatchPatternNode:

		modifiedIdentifier, identifierNode := replaceInner(n.Identifier, target, replacement)

		modifiedEqualToken, equalTokenNode := replaceInner(n.EqualToken, target, replacement)

		modifiedMatchPattern, matchPatternNode := replaceInner(n.MatchPattern, target, replacement)

		modified := modifiedIdentifier || modifiedEqualToken || modifiedMatchPattern
		if modified {
			return true, &STNamedArgMatchPatternNode{

				STNode: n.STNode,

				Identifier: identifierNode,

				EqualToken: equalTokenNode,

				MatchPattern: matchPatternNode,
			}
		}
		return false, current

	case *STMarkdownDocumentationNode:

		modifiedDocumentationLines, documentationLinesNode := replaceInner(n.DocumentationLines, target, replacement)

		modified := modifiedDocumentationLines
		if modified {
			return true, &STMarkdownDocumentationNode{

				STDocumentationNode: n.STDocumentationNode,

				DocumentationLines: documentationLinesNode,
			}
		}
		return false, current

	case *STMarkdownDocumentationLineNode:

		modifiedHashToken, hashTokenNode := replaceInner(n.HashToken, target, replacement)

		modifiedDocumentElements, documentElementsNode := replaceInner(n.DocumentElements, target, replacement)

		modified := modifiedHashToken || modifiedDocumentElements
		if modified {
			return true, &STMarkdownDocumentationLineNode{

				STDocumentationNode: n.STDocumentationNode,

				HashToken: hashTokenNode,

				DocumentElements: documentElementsNode,
			}
		}
		return false, current

	case *STMarkdownParameterDocumentationLineNode:

		modifiedHashToken, hashTokenNode := replaceInner(n.HashToken, target, replacement)

		modifiedPlusToken, plusTokenNode := replaceInner(n.PlusToken, target, replacement)

		modifiedParameterName, parameterNameNode := replaceInner(n.ParameterName, target, replacement)

		modifiedMinusToken, minusTokenNode := replaceInner(n.MinusToken, target, replacement)

		modifiedDocumentElements, documentElementsNode := replaceInner(n.DocumentElements, target, replacement)

		modified := modifiedHashToken || modifiedPlusToken || modifiedParameterName || modifiedMinusToken || modifiedDocumentElements
		if modified {
			return true, &STMarkdownParameterDocumentationLineNode{

				STDocumentationNode: n.STDocumentationNode,

				HashToken: hashTokenNode,

				PlusToken: plusTokenNode,

				ParameterName: parameterNameNode,

				MinusToken: minusTokenNode,

				DocumentElements: documentElementsNode,
			}
		}
		return false, current

	case *STBallerinaNameReferenceNode:

		modifiedReferenceType, referenceTypeNode := replaceInner(n.ReferenceType, target, replacement)

		modifiedStartBacktick, startBacktickNode := replaceInner(n.StartBacktick, target, replacement)

		modifiedNameReference, nameReferenceNode := replaceInner(n.NameReference, target, replacement)

		modifiedEndBacktick, endBacktickNode := replaceInner(n.EndBacktick, target, replacement)

		modified := modifiedReferenceType || modifiedStartBacktick || modifiedNameReference || modifiedEndBacktick
		if modified {
			return true, &STBallerinaNameReferenceNode{

				STDocumentationNode: n.STDocumentationNode,

				ReferenceType: referenceTypeNode,

				StartBacktick: startBacktickNode,

				NameReference: nameReferenceNode,

				EndBacktick: endBacktickNode,
			}
		}
		return false, current

	case *STInlineCodeReferenceNode:

		modifiedStartBacktick, startBacktickNode := replaceInner(n.StartBacktick, target, replacement)

		modifiedCodeReference, codeReferenceNode := replaceInner(n.CodeReference, target, replacement)

		modifiedEndBacktick, endBacktickNode := replaceInner(n.EndBacktick, target, replacement)

		modified := modifiedStartBacktick || modifiedCodeReference || modifiedEndBacktick
		if modified {
			return true, &STInlineCodeReferenceNode{

				STDocumentationNode: n.STDocumentationNode,

				StartBacktick: startBacktickNode,

				CodeReference: codeReferenceNode,

				EndBacktick: endBacktickNode,
			}
		}
		return false, current

	case *STMarkdownCodeBlockNode:

		modifiedStartLineHashToken, startLineHashTokenNode := replaceInner(n.StartLineHashToken, target, replacement)

		modifiedStartBacktick, startBacktickNode := replaceInner(n.StartBacktick, target, replacement)

		modifiedLangAttribute, langAttributeNode := replaceInner(n.LangAttribute, target, replacement)

		modifiedCodeLines, codeLinesNode := replaceInner(n.CodeLines, target, replacement)

		modifiedEndLineHashToken, endLineHashTokenNode := replaceInner(n.EndLineHashToken, target, replacement)

		modifiedEndBacktick, endBacktickNode := replaceInner(n.EndBacktick, target, replacement)

		modified := modifiedStartLineHashToken || modifiedStartBacktick || modifiedLangAttribute || modifiedCodeLines || modifiedEndLineHashToken || modifiedEndBacktick
		if modified {
			return true, &STMarkdownCodeBlockNode{

				STDocumentationNode: n.STDocumentationNode,

				StartLineHashToken: startLineHashTokenNode,

				StartBacktick: startBacktickNode,

				LangAttribute: langAttributeNode,

				CodeLines: codeLinesNode,

				EndLineHashToken: endLineHashTokenNode,

				EndBacktick: endBacktickNode,
			}
		}
		return false, current

	case *STMarkdownCodeLineNode:

		modifiedHashToken, hashTokenNode := replaceInner(n.HashToken, target, replacement)

		modifiedCodeDescription, codeDescriptionNode := replaceInner(n.CodeDescription, target, replacement)

		modified := modifiedHashToken || modifiedCodeDescription
		if modified {
			return true, &STMarkdownCodeLineNode{

				STDocumentationNode: n.STDocumentationNode,

				HashToken: hashTokenNode,

				CodeDescription: codeDescriptionNode,
			}
		}
		return false, current

	case *STOrderByClauseNode:

		modifiedOrderKeyword, orderKeywordNode := replaceInner(n.OrderKeyword, target, replacement)

		modifiedByKeyword, byKeywordNode := replaceInner(n.ByKeyword, target, replacement)

		modifiedOrderKey, orderKeyNode := replaceInner(n.OrderKey, target, replacement)

		modified := modifiedOrderKeyword || modifiedByKeyword || modifiedOrderKey
		if modified {
			return true, &STOrderByClauseNode{

				STIntermediateClauseNode: n.STIntermediateClauseNode,

				OrderKeyword: orderKeywordNode,

				ByKeyword: byKeywordNode,

				OrderKey: orderKeyNode,
			}
		}
		return false, current

	case *STOrderKeyNode:

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modifiedOrderDirection, orderDirectionNode := replaceInner(n.OrderDirection, target, replacement)

		modified := modifiedExpression || modifiedOrderDirection
		if modified {
			return true, &STOrderKeyNode{

				STNode: n.STNode,

				Expression: expressionNode,

				OrderDirection: orderDirectionNode,
			}
		}
		return false, current

	case *STGroupByClauseNode:

		modifiedGroupKeyword, groupKeywordNode := replaceInner(n.GroupKeyword, target, replacement)

		modifiedByKeyword, byKeywordNode := replaceInner(n.ByKeyword, target, replacement)

		modifiedGroupingKey, groupingKeyNode := replaceInner(n.GroupingKey, target, replacement)

		modified := modifiedGroupKeyword || modifiedByKeyword || modifiedGroupingKey
		if modified {
			return true, &STGroupByClauseNode{

				STIntermediateClauseNode: n.STIntermediateClauseNode,

				GroupKeyword: groupKeywordNode,

				ByKeyword: byKeywordNode,

				GroupingKey: groupingKeyNode,
			}
		}
		return false, current

	case *STGroupingKeyVarDeclarationNode:

		modifiedTypeDescriptor, typeDescriptorNode := replaceInner(n.TypeDescriptor, target, replacement)

		modifiedSimpleBindingPattern, simpleBindingPatternNode := replaceInner(n.SimpleBindingPattern, target, replacement)

		modifiedEqualsToken, equalsTokenNode := replaceInner(n.EqualsToken, target, replacement)

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modified := modifiedTypeDescriptor || modifiedSimpleBindingPattern || modifiedEqualsToken || modifiedExpression
		if modified {
			return true, &STGroupingKeyVarDeclarationNode{

				STNode: n.STNode,

				TypeDescriptor: typeDescriptorNode,

				SimpleBindingPattern: simpleBindingPatternNode,

				EqualsToken: equalsTokenNode,

				Expression: expressionNode,
			}
		}
		return false, current

	case *STOnFailClauseNode:

		modifiedOnKeyword, onKeywordNode := replaceInner(n.OnKeyword, target, replacement)

		modifiedFailKeyword, failKeywordNode := replaceInner(n.FailKeyword, target, replacement)

		modifiedTypedBindingPattern, typedBindingPatternNode := replaceInner(n.TypedBindingPattern, target, replacement)

		modifiedBlockStatement, blockStatementNode := replaceInner(n.BlockStatement, target, replacement)

		modified := modifiedOnKeyword || modifiedFailKeyword || modifiedTypedBindingPattern || modifiedBlockStatement
		if modified {
			return true, &STOnFailClauseNode{

				STClauseNode: n.STClauseNode,

				OnKeyword: onKeywordNode,

				FailKeyword: failKeywordNode,

				TypedBindingPattern: typedBindingPatternNode,

				BlockStatement: blockStatementNode,
			}
		}
		return false, current

	case *STDoStatementNode:

		modifiedDoKeyword, doKeywordNode := replaceInner(n.DoKeyword, target, replacement)

		modifiedBlockStatement, blockStatementNode := replaceInner(n.BlockStatement, target, replacement)

		modifiedOnFailClause, onFailClauseNode := replaceInner(n.OnFailClause, target, replacement)

		modified := modifiedDoKeyword || modifiedBlockStatement || modifiedOnFailClause
		if modified {
			return true, &STDoStatementNode{

				STStatementNode: n.STStatementNode,

				DoKeyword: doKeywordNode,

				BlockStatement: blockStatementNode,

				OnFailClause: onFailClauseNode,
			}
		}
		return false, current

	case *STClassDefinitionNode:

		modifiedMetadata, metadataNode := replaceInner(n.Metadata, target, replacement)

		modifiedVisibilityQualifier, visibilityQualifierNode := replaceInner(n.VisibilityQualifier, target, replacement)

		modifiedClassTypeQualifiers, classTypeQualifiersNode := replaceInner(n.ClassTypeQualifiers, target, replacement)

		modifiedClassKeyword, classKeywordNode := replaceInner(n.ClassKeyword, target, replacement)

		modifiedClassName, classNameNode := replaceInner(n.ClassName, target, replacement)

		modifiedOpenBrace, openBraceNode := replaceInner(n.OpenBrace, target, replacement)

		modifiedMembers, membersNode := replaceInner(n.Members, target, replacement)

		modifiedCloseBrace, closeBraceNode := replaceInner(n.CloseBrace, target, replacement)

		modifiedSemicolonToken, semicolonTokenNode := replaceInner(n.SemicolonToken, target, replacement)

		modified := modifiedMetadata || modifiedVisibilityQualifier || modifiedClassTypeQualifiers || modifiedClassKeyword || modifiedClassName || modifiedOpenBrace || modifiedMembers || modifiedCloseBrace || modifiedSemicolonToken
		if modified {
			return true, &STClassDefinitionNode{

				STModuleMemberDeclarationNode: n.STModuleMemberDeclarationNode,

				Metadata: metadataNode,

				VisibilityQualifier: visibilityQualifierNode,

				ClassTypeQualifiers: classTypeQualifiersNode,

				ClassKeyword: classKeywordNode,

				ClassName: classNameNode,

				OpenBrace: openBraceNode,

				Members: membersNode,

				CloseBrace: closeBraceNode,

				SemicolonToken: semicolonTokenNode,
			}
		}
		return false, current

	case *STResourcePathParameterNode:

		modifiedOpenBracketToken, openBracketTokenNode := replaceInner(n.OpenBracketToken, target, replacement)

		modifiedAnnotations, annotationsNode := replaceInner(n.Annotations, target, replacement)

		modifiedTypeDescriptor, typeDescriptorNode := replaceInner(n.TypeDescriptor, target, replacement)

		modifiedEllipsisToken, ellipsisTokenNode := replaceInner(n.EllipsisToken, target, replacement)

		modifiedParamName, paramNameNode := replaceInner(n.ParamName, target, replacement)

		modifiedCloseBracketToken, closeBracketTokenNode := replaceInner(n.CloseBracketToken, target, replacement)

		modified := modifiedOpenBracketToken || modifiedAnnotations || modifiedTypeDescriptor || modifiedEllipsisToken || modifiedParamName || modifiedCloseBracketToken
		if modified {
			return true, &STResourcePathParameterNode{

				STNode: n.STNode,

				OpenBracketToken: openBracketTokenNode,

				Annotations: annotationsNode,

				TypeDescriptor: typeDescriptorNode,

				EllipsisToken: ellipsisTokenNode,

				ParamName: paramNameNode,

				CloseBracketToken: closeBracketTokenNode,
			}
		}
		return false, current

	case *STRequiredExpressionNode:

		modifiedQuestionMarkToken, questionMarkTokenNode := replaceInner(n.QuestionMarkToken, target, replacement)

		modified := modifiedQuestionMarkToken
		if modified {
			return true, &STRequiredExpressionNode{

				STExpressionNode: n.STExpressionNode,

				QuestionMarkToken: questionMarkTokenNode,
			}
		}
		return false, current

	case *STErrorConstructorExpressionNode:

		modifiedErrorKeyword, errorKeywordNode := replaceInner(n.ErrorKeyword, target, replacement)

		modifiedTypeReference, typeReferenceNode := replaceInner(n.TypeReference, target, replacement)

		modifiedOpenParenToken, openParenTokenNode := replaceInner(n.OpenParenToken, target, replacement)

		modifiedArguments, argumentsNode := replaceInner(n.Arguments, target, replacement)

		modifiedCloseParenToken, closeParenTokenNode := replaceInner(n.CloseParenToken, target, replacement)

		modified := modifiedErrorKeyword || modifiedTypeReference || modifiedOpenParenToken || modifiedArguments || modifiedCloseParenToken
		if modified {
			return true, &STErrorConstructorExpressionNode{

				STExpressionNode: n.STExpressionNode,

				ErrorKeyword: errorKeywordNode,

				TypeReference: typeReferenceNode,

				OpenParenToken: openParenTokenNode,

				Arguments: argumentsNode,

				CloseParenToken: closeParenTokenNode,
			}
		}
		return false, current

	case *STParameterizedTypeDescriptorNode:

		modifiedKeywordToken, keywordTokenNode := replaceInner(n.KeywordToken, target, replacement)

		modifiedTypeParamNode, typeParamNodeNode := replaceInner(n.TypeParamNode, target, replacement)

		modified := modifiedKeywordToken || modifiedTypeParamNode
		if modified {
			return true, &STParameterizedTypeDescriptorNode{

				STTypeDescriptorNode: n.STTypeDescriptorNode,

				KeywordToken: keywordTokenNode,

				TypeParamNode: typeParamNodeNode,
			}
		}
		return false, current

	case *STSpreadMemberNode:

		modifiedEllipsis, ellipsisNode := replaceInner(n.Ellipsis, target, replacement)

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modified := modifiedEllipsis || modifiedExpression
		if modified {
			return true, &STSpreadMemberNode{

				STNode: n.STNode,

				Ellipsis: ellipsisNode,

				Expression: expressionNode,
			}
		}
		return false, current

	case *STClientResourceAccessActionNode:

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modifiedRightArrowToken, rightArrowTokenNode := replaceInner(n.RightArrowToken, target, replacement)

		modifiedSlashToken, slashTokenNode := replaceInner(n.SlashToken, target, replacement)

		modifiedResourceAccessPath, resourceAccessPathNode := replaceInner(n.ResourceAccessPath, target, replacement)

		modifiedDotToken, dotTokenNode := replaceInner(n.DotToken, target, replacement)

		modifiedMethodName, methodNameNode := replaceInner(n.MethodName, target, replacement)

		modifiedArguments, argumentsNode := replaceInner(n.Arguments, target, replacement)

		modified := modifiedExpression || modifiedRightArrowToken || modifiedSlashToken || modifiedResourceAccessPath || modifiedDotToken || modifiedMethodName || modifiedArguments
		if modified {
			return true, &STClientResourceAccessActionNode{

				STActionNode: n.STActionNode,

				Expression: expressionNode,

				RightArrowToken: rightArrowTokenNode,

				SlashToken: slashTokenNode,

				ResourceAccessPath: resourceAccessPathNode,

				DotToken: dotTokenNode,

				MethodName: methodNameNode,

				Arguments: argumentsNode,
			}
		}
		return false, current

	case *STComputedResourceAccessSegmentNode:

		modifiedOpenBracketToken, openBracketTokenNode := replaceInner(n.OpenBracketToken, target, replacement)

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modifiedCloseBracketToken, closeBracketTokenNode := replaceInner(n.CloseBracketToken, target, replacement)

		modified := modifiedOpenBracketToken || modifiedExpression || modifiedCloseBracketToken
		if modified {
			return true, &STComputedResourceAccessSegmentNode{

				STNode: n.STNode,

				OpenBracketToken: openBracketTokenNode,

				Expression: expressionNode,

				CloseBracketToken: closeBracketTokenNode,
			}
		}
		return false, current

	case *STResourceAccessRestSegmentNode:

		modifiedOpenBracketToken, openBracketTokenNode := replaceInner(n.OpenBracketToken, target, replacement)

		modifiedEllipsisToken, ellipsisTokenNode := replaceInner(n.EllipsisToken, target, replacement)

		modifiedExpression, expressionNode := replaceInner(n.Expression, target, replacement)

		modifiedCloseBracketToken, closeBracketTokenNode := replaceInner(n.CloseBracketToken, target, replacement)

		modified := modifiedOpenBracketToken || modifiedEllipsisToken || modifiedExpression || modifiedCloseBracketToken
		if modified {
			return true, &STResourceAccessRestSegmentNode{

				STNode: n.STNode,

				OpenBracketToken: openBracketTokenNode,

				EllipsisToken: ellipsisTokenNode,

				Expression: expressionNode,

				CloseBracketToken: closeBracketTokenNode,
			}
		}
		return false, current

	case *STReSequenceNode:

		modifiedReTerm, reTermNode := replaceInner(n.ReTerm, target, replacement)

		modified := modifiedReTerm
		if modified {
			return true, &STReSequenceNode{

				STNode: n.STNode,

				ReTerm: reTermNode,
			}
		}
		return false, current

	case *STReAtomQuantifierNode:

		modifiedReAtom, reAtomNode := replaceInner(n.ReAtom, target, replacement)

		modifiedReQuantifier, reQuantifierNode := replaceInner(n.ReQuantifier, target, replacement)

		modified := modifiedReAtom || modifiedReQuantifier
		if modified {
			return true, &STReAtomQuantifierNode{

				STReTermNode: n.STReTermNode,

				ReAtom: reAtomNode,

				ReQuantifier: reQuantifierNode,
			}
		}
		return false, current

	case *STReAtomCharOrEscapeNode:

		modifiedReAtomCharOrEscape, reAtomCharOrEscapeNode := replaceInner(n.ReAtomCharOrEscape, target, replacement)

		modified := modifiedReAtomCharOrEscape
		if modified {
			return true, &STReAtomCharOrEscapeNode{

				STNode: n.STNode,

				ReAtomCharOrEscape: reAtomCharOrEscapeNode,
			}
		}
		return false, current

	case *STReQuoteEscapeNode:

		modifiedSlashToken, slashTokenNode := replaceInner(n.SlashToken, target, replacement)

		modifiedReSyntaxChar, reSyntaxCharNode := replaceInner(n.ReSyntaxChar, target, replacement)

		modified := modifiedSlashToken || modifiedReSyntaxChar
		if modified {
			return true, &STReQuoteEscapeNode{

				STNode: n.STNode,

				SlashToken: slashTokenNode,

				ReSyntaxChar: reSyntaxCharNode,
			}
		}
		return false, current

	case *STReSimpleCharClassEscapeNode:

		modifiedSlashToken, slashTokenNode := replaceInner(n.SlashToken, target, replacement)

		modifiedReSimpleCharClassCode, reSimpleCharClassCodeNode := replaceInner(n.ReSimpleCharClassCode, target, replacement)

		modified := modifiedSlashToken || modifiedReSimpleCharClassCode
		if modified {
			return true, &STReSimpleCharClassEscapeNode{

				STNode: n.STNode,

				SlashToken: slashTokenNode,

				ReSimpleCharClassCode: reSimpleCharClassCodeNode,
			}
		}
		return false, current

	case *STReUnicodePropertyEscapeNode:

		modifiedSlashToken, slashTokenNode := replaceInner(n.SlashToken, target, replacement)

		modifiedProperty, propertyNode := replaceInner(n.Property, target, replacement)

		modifiedOpenBraceToken, openBraceTokenNode := replaceInner(n.OpenBraceToken, target, replacement)

		modifiedReUnicodeProperty, reUnicodePropertyNode := replaceInner(n.ReUnicodeProperty, target, replacement)

		modifiedCloseBraceToken, closeBraceTokenNode := replaceInner(n.CloseBraceToken, target, replacement)

		modified := modifiedSlashToken || modifiedProperty || modifiedOpenBraceToken || modifiedReUnicodeProperty || modifiedCloseBraceToken
		if modified {
			return true, &STReUnicodePropertyEscapeNode{

				STNode: n.STNode,

				SlashToken: slashTokenNode,

				Property: propertyNode,

				OpenBraceToken: openBraceTokenNode,

				ReUnicodeProperty: reUnicodePropertyNode,

				CloseBraceToken: closeBraceTokenNode,
			}
		}
		return false, current

	case *STReUnicodeScriptNode:

		modifiedScriptStart, scriptStartNode := replaceInner(n.ScriptStart, target, replacement)

		modifiedReUnicodePropertyValue, reUnicodePropertyValueNode := replaceInner(n.ReUnicodePropertyValue, target, replacement)

		modified := modifiedScriptStart || modifiedReUnicodePropertyValue
		if modified {
			return true, &STReUnicodeScriptNode{

				STReUnicodePropertyNode: n.STReUnicodePropertyNode,

				ScriptStart: scriptStartNode,

				ReUnicodePropertyValue: reUnicodePropertyValueNode,
			}
		}
		return false, current

	case *STReUnicodeGeneralCategoryNode:

		modifiedCategoryStart, categoryStartNode := replaceInner(n.CategoryStart, target, replacement)

		modifiedReUnicodeGeneralCategoryName, reUnicodeGeneralCategoryNameNode := replaceInner(n.ReUnicodeGeneralCategoryName, target, replacement)

		modified := modifiedCategoryStart || modifiedReUnicodeGeneralCategoryName
		if modified {
			return true, &STReUnicodeGeneralCategoryNode{

				STReUnicodePropertyNode: n.STReUnicodePropertyNode,

				CategoryStart: categoryStartNode,

				ReUnicodeGeneralCategoryName: reUnicodeGeneralCategoryNameNode,
			}
		}
		return false, current

	case *STReCharacterClassNode:

		modifiedOpenBracket, openBracketNode := replaceInner(n.OpenBracket, target, replacement)

		modifiedNegation, negationNode := replaceInner(n.Negation, target, replacement)

		modifiedReCharSet, reCharSetNode := replaceInner(n.ReCharSet, target, replacement)

		modifiedCloseBracket, closeBracketNode := replaceInner(n.CloseBracket, target, replacement)

		modified := modifiedOpenBracket || modifiedNegation || modifiedReCharSet || modifiedCloseBracket
		if modified {
			return true, &STReCharacterClassNode{

				STNode: n.STNode,

				OpenBracket: openBracketNode,

				Negation: negationNode,

				ReCharSet: reCharSetNode,

				CloseBracket: closeBracketNode,
			}
		}
		return false, current

	case *STReCharSetRangeWithReCharSetNode:

		modifiedReCharSetRange, reCharSetRangeNode := replaceInner(n.ReCharSetRange, target, replacement)

		modifiedReCharSet, reCharSetNode := replaceInner(n.ReCharSet, target, replacement)

		modified := modifiedReCharSetRange || modifiedReCharSet
		if modified {
			return true, &STReCharSetRangeWithReCharSetNode{

				STNode: n.STNode,

				ReCharSetRange: reCharSetRangeNode,

				ReCharSet: reCharSetNode,
			}
		}
		return false, current

	case *STReCharSetRangeNode:

		modifiedLhsReCharSetAtom, lhsReCharSetAtomNode := replaceInner(n.LhsReCharSetAtom, target, replacement)

		modifiedMinusToken, minusTokenNode := replaceInner(n.MinusToken, target, replacement)

		modifiedRhsReCharSetAtom, rhsReCharSetAtomNode := replaceInner(n.RhsReCharSetAtom, target, replacement)

		modified := modifiedLhsReCharSetAtom || modifiedMinusToken || modifiedRhsReCharSetAtom
		if modified {
			return true, &STReCharSetRangeNode{

				STNode: n.STNode,

				LhsReCharSetAtom: lhsReCharSetAtomNode,

				MinusToken: minusTokenNode,

				RhsReCharSetAtom: rhsReCharSetAtomNode,
			}
		}
		return false, current

	case *STReCharSetAtomWithReCharSetNoDashNode:

		modifiedReCharSetAtom, reCharSetAtomNode := replaceInner(n.ReCharSetAtom, target, replacement)

		modifiedReCharSetNoDash, reCharSetNoDashNode := replaceInner(n.ReCharSetNoDash, target, replacement)

		modified := modifiedReCharSetAtom || modifiedReCharSetNoDash
		if modified {
			return true, &STReCharSetAtomWithReCharSetNoDashNode{

				STNode: n.STNode,

				ReCharSetAtom: reCharSetAtomNode,

				ReCharSetNoDash: reCharSetNoDashNode,
			}
		}
		return false, current

	case *STReCharSetRangeNoDashWithReCharSetNode:

		modifiedReCharSetRangeNoDash, reCharSetRangeNoDashNode := replaceInner(n.ReCharSetRangeNoDash, target, replacement)

		modifiedReCharSet, reCharSetNode := replaceInner(n.ReCharSet, target, replacement)

		modified := modifiedReCharSetRangeNoDash || modifiedReCharSet
		if modified {
			return true, &STReCharSetRangeNoDashWithReCharSetNode{

				STNode: n.STNode,

				ReCharSetRangeNoDash: reCharSetRangeNoDashNode,

				ReCharSet: reCharSetNode,
			}
		}
		return false, current

	case *STReCharSetRangeNoDashNode:

		modifiedReCharSetAtomNoDash, reCharSetAtomNoDashNode := replaceInner(n.ReCharSetAtomNoDash, target, replacement)

		modifiedMinusToken, minusTokenNode := replaceInner(n.MinusToken, target, replacement)

		modifiedReCharSetAtom, reCharSetAtomNode := replaceInner(n.ReCharSetAtom, target, replacement)

		modified := modifiedReCharSetAtomNoDash || modifiedMinusToken || modifiedReCharSetAtom
		if modified {
			return true, &STReCharSetRangeNoDashNode{

				STNode: n.STNode,

				ReCharSetAtomNoDash: reCharSetAtomNoDashNode,

				MinusToken: minusTokenNode,

				ReCharSetAtom: reCharSetAtomNode,
			}
		}
		return false, current

	case *STReCharSetAtomNoDashWithReCharSetNoDashNode:

		modifiedReCharSetAtomNoDash, reCharSetAtomNoDashNode := replaceInner(n.ReCharSetAtomNoDash, target, replacement)

		modifiedReCharSetNoDash, reCharSetNoDashNode := replaceInner(n.ReCharSetNoDash, target, replacement)

		modified := modifiedReCharSetAtomNoDash || modifiedReCharSetNoDash
		if modified {
			return true, &STReCharSetAtomNoDashWithReCharSetNoDashNode{

				STNode: n.STNode,

				ReCharSetAtomNoDash: reCharSetAtomNoDashNode,

				ReCharSetNoDash: reCharSetNoDashNode,
			}
		}
		return false, current

	case *STReCapturingGroupsNode:

		modifiedOpenParenthesis, openParenthesisNode := replaceInner(n.OpenParenthesis, target, replacement)

		modifiedReFlagExpression, reFlagExpressionNode := replaceInner(n.ReFlagExpression, target, replacement)

		modifiedReSequences, reSequencesNode := replaceInner(n.ReSequences, target, replacement)

		modifiedCloseParenthesis, closeParenthesisNode := replaceInner(n.CloseParenthesis, target, replacement)

		modified := modifiedOpenParenthesis || modifiedReFlagExpression || modifiedReSequences || modifiedCloseParenthesis
		if modified {
			return true, &STReCapturingGroupsNode{

				STNode: n.STNode,

				OpenParenthesis: openParenthesisNode,

				ReFlagExpression: reFlagExpressionNode,

				ReSequences: reSequencesNode,

				CloseParenthesis: closeParenthesisNode,
			}
		}
		return false, current

	case *STReFlagExpressionNode:

		modifiedQuestionMark, questionMarkNode := replaceInner(n.QuestionMark, target, replacement)

		modifiedReFlagsOnOff, reFlagsOnOffNode := replaceInner(n.ReFlagsOnOff, target, replacement)

		modifiedColon, colonNode := replaceInner(n.Colon, target, replacement)

		modified := modifiedQuestionMark || modifiedReFlagsOnOff || modifiedColon
		if modified {
			return true, &STReFlagExpressionNode{

				STNode: n.STNode,

				QuestionMark: questionMarkNode,

				ReFlagsOnOff: reFlagsOnOffNode,

				Colon: colonNode,
			}
		}
		return false, current

	case *STReFlagsOnOffNode:

		modifiedLhsReFlags, lhsReFlagsNode := replaceInner(n.LhsReFlags, target, replacement)

		modifiedMinusToken, minusTokenNode := replaceInner(n.MinusToken, target, replacement)

		modifiedRhsReFlags, rhsReFlagsNode := replaceInner(n.RhsReFlags, target, replacement)

		modified := modifiedLhsReFlags || modifiedMinusToken || modifiedRhsReFlags
		if modified {
			return true, &STReFlagsOnOffNode{

				STNode: n.STNode,

				LhsReFlags: lhsReFlagsNode,

				MinusToken: minusTokenNode,

				RhsReFlags: rhsReFlagsNode,
			}
		}
		return false, current

	case *STReFlagsNode:

		modifiedReFlag, reFlagNode := replaceInner(n.ReFlag, target, replacement)

		modified := modifiedReFlag
		if modified {
			return true, &STReFlagsNode{

				STNode: n.STNode,

				ReFlag: reFlagNode,
			}
		}
		return false, current

	case *STReAssertionNode:

		modifiedReAssertion, reAssertionNode := replaceInner(n.ReAssertion, target, replacement)

		modified := modifiedReAssertion
		if modified {
			return true, &STReAssertionNode{

				STReTermNode: n.STReTermNode,

				ReAssertion: reAssertionNode,
			}
		}
		return false, current

	case *STReQuantifierNode:

		modifiedReBaseQuantifier, reBaseQuantifierNode := replaceInner(n.ReBaseQuantifier, target, replacement)

		modifiedNonGreedyChar, nonGreedyCharNode := replaceInner(n.NonGreedyChar, target, replacement)

		modified := modifiedReBaseQuantifier || modifiedNonGreedyChar
		if modified {
			return true, &STReQuantifierNode{

				STNode: n.STNode,

				ReBaseQuantifier: reBaseQuantifierNode,

				NonGreedyChar: nonGreedyCharNode,
			}
		}
		return false, current

	case *STReBracedQuantifierNode:

		modifiedOpenBraceToken, openBraceTokenNode := replaceInner(n.OpenBraceToken, target, replacement)

		modifiedLeastTimesMatchedDigit, leastTimesMatchedDigitNode := replaceInner(n.LeastTimesMatchedDigit, target, replacement)

		modifiedCommaToken, commaTokenNode := replaceInner(n.CommaToken, target, replacement)

		modifiedMostTimesMatchedDigit, mostTimesMatchedDigitNode := replaceInner(n.MostTimesMatchedDigit, target, replacement)

		modifiedCloseBraceToken, closeBraceTokenNode := replaceInner(n.CloseBraceToken, target, replacement)

		modified := modifiedOpenBraceToken || modifiedLeastTimesMatchedDigit || modifiedCommaToken || modifiedMostTimesMatchedDigit || modifiedCloseBraceToken
		if modified {
			return true, &STReBracedQuantifierNode{

				STNode: n.STNode,

				OpenBraceToken: openBraceTokenNode,

				LeastTimesMatchedDigit: leastTimesMatchedDigitNode,

				CommaToken: commaTokenNode,

				MostTimesMatchedDigit: mostTimesMatchedDigitNode,

				CloseBraceToken: closeBraceTokenNode,
			}
		}
		return false, current

	case *STMemberTypeDescriptorNode:

		modifiedAnnotations, annotationsNode := replaceInner(n.Annotations, target, replacement)

		modifiedTypeDescriptor, typeDescriptorNode := replaceInner(n.TypeDescriptor, target, replacement)

		modified := modifiedAnnotations || modifiedTypeDescriptor
		if modified {
			return true, &STMemberTypeDescriptorNode{

				STNode: n.STNode,

				Annotations: annotationsNode,

				TypeDescriptor: typeDescriptorNode,
			}
		}
		return false, current

	case *STReceiveFieldNode:

		modifiedFieldName, fieldNameNode := replaceInner(n.FieldName, target, replacement)

		modifiedColon, colonNode := replaceInner(n.Colon, target, replacement)

		modifiedPeerWorker, peerWorkerNode := replaceInner(n.PeerWorker, target, replacement)

		modified := modifiedFieldName || modifiedColon || modifiedPeerWorker
		if modified {
			return true, &STReceiveFieldNode{

				STNode: n.STNode,

				FieldName: fieldNameNode,

				Colon: colonNode,

				PeerWorker: peerWorkerNode,
			}
		}
		return false, current

	case *STNaturalExpressionNode:

		modifiedConstKeyword, constKeywordNode := replaceInner(n.ConstKeyword, target, replacement)

		modifiedNaturalKeyword, naturalKeywordNode := replaceInner(n.NaturalKeyword, target, replacement)

		modifiedParenthesizedArgList, parenthesizedArgListNode := replaceInner(n.ParenthesizedArgList, target, replacement)

		modifiedOpenBraceToken, openBraceTokenNode := replaceInner(n.OpenBraceToken, target, replacement)

		modifiedPrompt, promptNode := replaceInner(n.Prompt, target, replacement)

		modifiedCloseBraceToken, closeBraceTokenNode := replaceInner(n.CloseBraceToken, target, replacement)

		modified := modifiedConstKeyword || modifiedNaturalKeyword || modifiedParenthesizedArgList || modifiedOpenBraceToken || modifiedPrompt || modifiedCloseBraceToken
		if modified {
			return true, &STNaturalExpressionNode{

				STExpressionNode: n.STExpressionNode,

				ConstKeyword: constKeywordNode,

				NaturalKeyword: naturalKeywordNode,

				ParenthesizedArgList: parenthesizedArgListNode,

				OpenBraceToken: openBraceTokenNode,

				Prompt: promptNode,

				CloseBraceToken: closeBraceTokenNode,
			}
		}
		return false, current

	case STToken:
		return false, current
	default:
		panic("unsupported node type: " + reflect.TypeOf(current).Name())
	}
}
