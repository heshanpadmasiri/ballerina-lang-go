// Copyright (c) 2025, WSO2 LLC. (http://www.wso2.com).
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
package parser

import (
	debugcommon "ballerina-lang-go/common"
	"ballerina-lang-go/parser/common"
	"ballerina-lang-go/parser/internal"
)

// This ports the current java parser as is, which IMO has an error recovery mechanism that is unnecessarily
// complicated. Sane approach is to insert error nodes (usually you have different onces for expr and stmt) and keep
// seeking until a synchronization point (typically end of block or end of stmt) is reached, and resume parsing

// Error recovery types
type Action int

const (
	ACTION_KEEP Action = iota
	ACTION_INSERT
	ACTION_REMOVE
)

type Solution struct {
	action        Action
	recoveredNode internal.STNode
	removedToken  internal.STToken
}

type InvalidNodeInfo struct {
	node           internal.STNode
	diagnosticCode *DiagnosticErrorCode
	args           []any
}

// ParserRuleContext constants (adding incrementally as needed)
type ParserRuleContext int

const (
	PARSER_RULE_CONTEXT_COMP_UNIT ParserRuleContext = iota
	PARSER_RULE_CONTEXT_TOP_LEVEL_NODE
	PARSER_RULE_CONTEXT_TOP_LEVEL_NODE_WITHOUT_METADATA
	PARSER_RULE_CONTEXT_TOP_LEVEL_NODE_WITHOUT_MODIFIER
	PARSER_RULE_CONTEXT_IMPORT_DECL
	PARSER_RULE_CONTEXT_IMPORT_KEYWORD
	PARSER_RULE_CONTEXT_IMPORT_ORG_OR_MODULE_NAME
	PARSER_RULE_CONTEXT_IMPORT_DECL_ORG_OR_MODULE_NAME_RHS
	PARSER_RULE_CONTEXT_SLASH
	PARSER_RULE_CONTEXT_DOT
	PARSER_RULE_CONTEXT_AT
)

type Parser struct {
	tokenReader          TokenReader
	dbgContext           *debugcommon.DebugContext
	insertedToken        internal.STToken
	invalidNodeInfoStack []InvalidNodeInfo
	contextStack         []ParserRuleContext
}

func CreateParser(tokenReader TokenReader, dbgContext *debugcommon.DebugContext) *Parser {
	return &Parser{
		tokenReader: tokenReader,
		dbgContext:  dbgContext,
	}
}

func (p *Parser) Parse() internal.STNode {
	return p.parseCompUnit()
}

func (p *Parser) parseCompUnit() internal.STNode {
	p.startContext(PARSER_RULE_CONTEXT_COMP_UNIT)
	var otherDecls []internal.STNode
	var importDecls []internal.STNode
	processImports := true
	token := p.peek()
	for token.Kind() != common.EOF_TOKEN {
		decl := p.parseTopLevelNode()
		if decl == nil {
			break
		}
		if decl.Kind() == common.IMPORT_DECLARATION {
			if processImports {
				importDecls = append(importDecls, decl)
			} else {
				// If an import occurs after any other module level declaration,
				// we add it to the other-decl list to preserve the order. But
				// log an error and mark it as invalid.
				p.updateLastNodeInListWithInvalidNode(&otherDecls, decl, &ERROR_IMPORT_DECLARATION_AFTER_OTHER_DECLARATIONS)
			}
		} else {
			if processImports {
				// While processing imports, if we reach any other declaration,
				// then mark this as the end of processing imports.
				processImports = false
			}
			otherDecls = append(otherDecls, decl)
		}
		token = p.peek()
	}

	eof := p.consume()
	p.endContext()

	imports := internal.CreateNodeList(importDecls)
	members := internal.CreateNodeList(otherDecls)
	return &internal.STModulePart{
		Imports:  &imports,
		Members:  &members,
		EofToken: eof,
	}
}

func (p *Parser) parseTopLevelNode() internal.STNode {
	nextToken := p.peek()
	var metadata internal.STNode
	switch nextToken.Kind() {
	case common.EOF_TOKEN:
		return nil
	case common.DOCUMENTATION_STRING, common.AT_TOKEN:
		metadata = p.parseMetadata()
		return p.parseTopLevelNodeWithMetadata(metadata)
	case common.IMPORT_KEYWORD, common.FINAL_KEYWORD, common.PUBLIC_KEYWORD, common.FUNCTION_KEYWORD, common.TYPE_KEYWORD, common.LISTENER_KEYWORD, common.CONST_KEYWORD, common.ANNOTATION_KEYWORD, common.XMLNS_KEYWORD, common.ENUM_KEYWORD, common.CLASS_KEYWORD, common.TRANSACTIONAL_KEYWORD, common.ISOLATED_KEYWORD, common.DISTINCT_KEYWORD, common.CLIENT_KEYWORD, common.READONLY_KEYWORD, common.CONFIGURABLE_KEYWORD, common.SERVICE_KEYWORD:
		metadata = p.createEmptyNode()
	case common.RESOURCE_KEYWORD, common.REMOTE_KEYWORD:
		// Special case to invalidate
		invalidToken := p.consume()
		p.reportInvalidQualifier(invalidToken)
		return p.parseTopLevelNode()
	case common.IDENTIFIER_TOKEN:
		// Here we assume that after recovering, we'll never reach here.
		// Otherwise the tokenOffset will not be 1.
		if p.isModuleVarDeclStart(1) || nextToken.IsMissing() {
			// This is an early exit, so that we don't have to do the same check again.
			return p.parseModuleVarDecl(p.createEmptyNode())
		}
		// Else fall through
	default:
		if p.isTypeStartingToken(nextToken.Kind()) && nextToken.Kind() != common.IDENTIFIER_TOKEN {
			metadata = p.createEmptyNode()
			break
		}

		token := p.peek()
		solution := p.recover(token, PARSER_RULE_CONTEXT_TOP_LEVEL_NODE)

		if solution.action == ACTION_KEEP {
			// If the solution is KEEP, that means next immediate token is
			// at the correct place, but some token after that is not. There only one such
			// cases here, which is the `case IDENTIFIER_TOKEN`. So accept it, and continue.
			metadata = p.createEmptyNode()
			break
		}

		return p.parseTopLevelNode()
	}

	return p.parseTopLevelNodeWithMetadata(metadata)
}

// AbstractParser helper methods

func (p *Parser) peek() internal.STToken {
	if p.insertedToken != nil {
		return p.insertedToken
	}
	return p.tokenReader.Peek()
}

func (p *Parser) peekN(k int) internal.STToken {
	if p.insertedToken == nil {
		return p.tokenReader.PeekN(k)
	}

	if k == 1 {
		return p.insertedToken
	}

	if k > 0 {
		k = k - 1
	}

	return p.tokenReader.PeekN(k)
}

func (p *Parser) consume() internal.STToken {
	if p.insertedToken != nil {
		nextToken := p.insertedToken
		p.insertedToken = nil
		return p.consumeWithInvalidNodesFromToken(nextToken)
	}

	if len(p.invalidNodeInfoStack) == 0 {
		return p.tokenReader.Read()
	}

	return p.consumeWithInvalidNodes()
}

func (p *Parser) consumeWithInvalidNodes() internal.STToken {
	token := p.tokenReader.Read()
	return p.consumeWithInvalidNodesFromToken(token)
}

func (p *Parser) consumeWithInvalidNodesFromToken(token internal.STToken) internal.STToken {
	for len(p.invalidNodeInfoStack) > 0 {
		invalidNodeInfo := p.invalidNodeInfoStack[len(p.invalidNodeInfoStack)-1]
		p.invalidNodeInfoStack = p.invalidNodeInfoStack[:len(p.invalidNodeInfoStack)-1]
		// TODO: cloneWithLeadingInvalidNodeMinutiae - need to implement this
		token = p.cloneWithLeadingInvalidNodeMinutiae(token, invalidNodeInfo.node, invalidNodeInfo.diagnosticCode, invalidNodeInfo.args...)
	}
	return token
}

func (p *Parser) recover(token internal.STToken, currentCtx ParserRuleContext) Solution {
	// Stub implementation - returns KEEP action
	// TODO: implement full error recovery logic
	return Solution{
		action: ACTION_KEEP,
	}
}

func (p *Parser) insertToken(kind common.SyntaxKind, context ParserRuleContext) {
	// TODO: implement createMissingTokenWithDiagnostics
	p.insertedToken = p.createMissingTokenWithDiagnostics(kind, context)
}

func (p *Parser) removeInsertedToken() {
	p.insertedToken = nil
}

func (p *Parser) startContext(context ParserRuleContext) {
	p.contextStack = append(p.contextStack, context)
}

func (p *Parser) endContext() {
	if len(p.contextStack) > 0 {
		p.contextStack = p.contextStack[:len(p.contextStack)-1]
	}
}

func (p *Parser) getCurrentContext() ParserRuleContext {
	if len(p.contextStack) == 0 {
		return PARSER_RULE_CONTEXT_COMP_UNIT
	}
	return p.contextStack[len(p.contextStack)-1]
}

func (p *Parser) switchContext(context ParserRuleContext) {
	if len(p.contextStack) > 0 {
		p.contextStack[len(p.contextStack)-1] = context
	} else {
		p.contextStack = append(p.contextStack, context)
	}
}

func (p *Parser) addInvalidNodeToNextToken(invalidNode internal.STNode, diagnosticCode *DiagnosticErrorCode, args ...any) {
	p.invalidNodeInfoStack = append(p.invalidNodeInfoStack, InvalidNodeInfo{
		node:           invalidNode,
		diagnosticCode: diagnosticCode,
		args:           args,
	})
}

func (p *Parser) addInvalidTokenToNextToken(invalidToken internal.STToken) {
	p.addInvalidNodeToNextToken(invalidToken, &ERROR_INVALID_TOKEN, invalidToken.Text())
}

func (p *Parser) updateLastNodeInListWithInvalidNode(nodeList *[]internal.STNode, invalidParam internal.STNode, diagnosticCode *DiagnosticErrorCode) {
	if len(*nodeList) == 0 {
		return
	}
	lastIndex := len(*nodeList) - 1
	prevNode := (*nodeList)[lastIndex]
	*nodeList = (*nodeList)[:lastIndex]
	newNode := p.cloneWithTrailingInvalidNodeMinutiae(prevNode, invalidParam, diagnosticCode)
	*nodeList = append(*nodeList, newNode)
}

func (p *Parser) invalidateRestAndAddToTrailingMinutiae(node internal.STNode) internal.STNode {
	node = p.addInvalidNodeStackToTrailingMinutiae(node)

	for p.peek().Kind() != common.EOF_TOKEN {
		invalidToken := p.consume()
		node = p.cloneWithTrailingInvalidNodeMinutiae(node, invalidToken, &ERROR_INVALID_TOKEN, invalidToken.Text())
	}

	return node
}

func (p *Parser) addInvalidNodeStackToTrailingMinutiae(node internal.STNode) internal.STNode {
	for len(p.invalidNodeInfoStack) > 0 {
		invalidNodeInfo := p.invalidNodeInfoStack[len(p.invalidNodeInfoStack)-1]
		p.invalidNodeInfoStack = p.invalidNodeInfoStack[:len(p.invalidNodeInfoStack)-1]
		node = p.cloneWithTrailingInvalidNodeMinutiae(node, invalidNodeInfo.node, invalidNodeInfo.diagnosticCode, invalidNodeInfo.args...)
	}
	return node
}

func (p *Parser) isNodeListEmpty(node internal.STNode) bool {
	// TODO: implement proper check for STNodeList
	if node == nil {
		return true
	}
	if node.Kind() == common.LIST {
		if nodeList, ok := node.(*internal.STNodeList); ok {
			return nodeList.BucketCount() == 0
		}
	}
	return false
}

func (p *Parser) cloneWithDiagnosticIfListEmpty(nodeList internal.STNode, target internal.STNode, diagnosticCode *DiagnosticErrorCode) internal.STNode {
	if p.isNodeListEmpty(nodeList) {
		return internal.AddSyntaxDiagnostic(target, internal.CreateDiagnostic(diagnosticCode))
	}
	return target
}

// Helper methods for creating nodes

func (p *Parser) createEmptyNode() internal.STNode {
	return nil
}

func (p *Parser) createMissingToken(kind common.SyntaxKind) internal.STToken {
	// TODO: implement proper missing token creation
	return internal.CreateTokenFrom(kind, internal.CreateEmptyNodeList(), internal.CreateEmptyNodeList())
}

func (p *Parser) createMissingTokenWithDiagnostics(kind common.SyntaxKind, context ParserRuleContext) internal.STToken {
	// TODO: implement proper missing token with diagnostics
	diagnostic := internal.CreateDiagnostic(p.getErrorCode(context))
	return p.createMissingTokenWithDiagnosticCode(kind, diagnostic)
}

func (p *Parser) createMissingTokenWithDiagnosticCode(kind common.SyntaxKind, diagnostic internal.STNodeDiagnostic) internal.STToken {
	// TODO: implement proper missing token with diagnostics
	token := p.createMissingToken(kind)
	return internal.AddSyntaxDiagnostic(token, diagnostic)
}

func (p *Parser) getErrorCode(context ParserRuleContext) *DiagnosticErrorCode {
	// TODO: implement error code mapping based on context
	return &ERROR_MISSING_TOKEN
}

// Clone methods for invalid node minutiae (stubs - need proper implementation)

func (p *Parser) cloneWithLeadingInvalidNodeMinutiae(token internal.STToken, invalidNode internal.STNode, diagnosticCode *DiagnosticErrorCode, args ...any) internal.STToken {
	// TODO: implement proper clone with leading invalid node minutiae
	diagnostic := internal.CreateDiagnostic(diagnosticCode, args...)
	return internal.AddSyntaxDiagnostic(token, diagnostic)
}

func (p *Parser) cloneWithTrailingInvalidNodeMinutiae(node internal.STNode, invalidNode internal.STNode, diagnosticCode *DiagnosticErrorCode, args ...any) internal.STNode {
	// TODO: implement proper clone with trailing invalid node minutiae
	diagnostic := internal.CreateDiagnostic(diagnosticCode, args...)
	return internal.AddSyntaxDiagnostic(node, diagnostic)
}

// Parsing methods

func (p *Parser) parseTopLevelNodeWithMetadata(metadata internal.STNode) internal.STNode {
	nextToken := p.peek()
	var publicQualifier internal.STNode
	switch nextToken.Kind() {
	case common.EOF_TOKEN:
		if metadata != nil {
			metadata = p.addMetadataNotAttachedDiagnostic(metadata)
			return p.createMissingSimpleVarDecl(metadata, true)
		}
		return nil
	case common.PUBLIC_KEYWORD:
		publicQualifier = p.consume()
	case common.FUNCTION_KEYWORD, common.TYPE_KEYWORD, common.LISTENER_KEYWORD, common.CONST_KEYWORD, common.FINAL_KEYWORD, common.IMPORT_KEYWORD, common.ANNOTATION_KEYWORD, common.XMLNS_KEYWORD, common.ENUM_KEYWORD, common.CLASS_KEYWORD, common.TRANSACTIONAL_KEYWORD, common.ISOLATED_KEYWORD, common.DISTINCT_KEYWORD, common.CLIENT_KEYWORD, common.READONLY_KEYWORD, common.SERVICE_KEYWORD, common.CONFIGURABLE_KEYWORD:
		// Top level qualifiers
		break
	case common.RESOURCE_KEYWORD, common.REMOTE_KEYWORD:
		// Special case to invalidate
		invalidToken := p.consume()
		p.reportInvalidQualifier(invalidToken)
		return p.parseTopLevelNode()
	case common.IDENTIFIER_TOKEN:
		// Here we assume that after recovering, we'll never reach here.
		// Otherwise the tokenOffset will not be 1.
		if p.isModuleVarDeclStart(1) || nextToken.IsMissing() {
			// This is an early exit, so that we don't have to do the same check again.
			return p.parseModuleVarDecl(nil)
		}
		fallthrough
		// Else fall through
	default:
		if nextToken.Kind() != common.IDENTIFIER_TOKEN && p.isTypeStartingToken(nextToken.Kind()) {
			metadata = nil
			break
		}

		token := p.peek()
		solution := p.recover(token, PARSER_RULE_CONTEXT_TOP_LEVEL_NODE)

		if solution.action == ACTION_KEEP {
			// If the solution is KEEP, that means next immediate token is
			// at the correct place, but some token after that is not. There only one such
			// cases here, which is the `case IDENTIFIER_TOKEN`. So accept it, and continue.
			metadata = nil
			break
		}

		return p.parseTopLevelNode()
	}

	return p.parseTopLevelNodeWithMetadata(metadata)
}

func (p *Parser) parseTopLevelNodeWithMetadataQualifierAndQualifiers(metadata internal.STNode, publicQualifier internal.STNode, qualifiers []internal.STNode) internal.STNode {
	p.parseTopLevelQualifiers(&qualifiers)
	nextToken := p.peek()
	switch nextToken.Kind() {
	case common.EOF_TOKEN:
		return p.createMissingSimpleVarDeclWithQualifiersImpl(metadata, publicQualifier, qualifiers, true)
	case common.FUNCTION_KEYWORD:
		// Anything starts with a function keyword could be a function definition
		// or a module-var-decl with function type desc.
		return p.parseFuncDefOrFuncTypeDesc(metadata, publicQualifier, qualifiers, false, false)
	case common.TYPE_KEYWORD:
		p.reportInvalidQualifierList(qualifiers)
		return p.parseModuleTypeDefinition(metadata, publicQualifier)
	case common.CLASS_KEYWORD:
		return p.parseClassDefinition(metadata, publicQualifier, qualifiers)
	case common.LISTENER_KEYWORD:
		p.reportInvalidQualifierList(qualifiers)
		return p.parseListenerDeclaration(metadata, publicQualifier)
	case common.CONST_KEYWORD:
		p.reportInvalidQualifierList(qualifiers)
		return p.parseConstantDeclaration(metadata, publicQualifier)
	case common.ANNOTATION_KEYWORD:
		p.reportInvalidQualifierList(qualifiers)
		constKeyword := p.createEmptyNode()
		return p.parseAnnotationDeclaration(metadata, publicQualifier, constKeyword)
	case common.IMPORT_KEYWORD:
		p.reportInvalidMetaData(metadata, "import declaration")
		p.reportInvalidQualifier(publicQualifier)
		p.reportInvalidQualifierList(qualifiers)
		return p.parseImportDecl()
	case common.XMLNS_KEYWORD:
		p.reportInvalidMetaData(metadata, "XML namespace declaration")
		p.reportInvalidQualifier(publicQualifier)
		p.reportInvalidQualifierList(qualifiers)
		return p.parseXMLNamespaceDeclaration(true)
	case common.ENUM_KEYWORD:
		p.reportInvalidQualifierList(qualifiers)
		return p.parseEnumDeclaration(metadata, publicQualifier)
	case common.RESOURCE_KEYWORD, common.REMOTE_KEYWORD:
		// Special case to invalidate
		invalidToken := p.consume()
		p.reportInvalidQualifier(invalidToken)
		return p.parseTopLevelNodeWithMetadataQualifierAndQualifiers(metadata, publicQualifier, qualifiers)
	case common.IDENTIFIER_TOKEN:
		// Here we assume that after recovering, we'll never reach here.
		// Otherwise the tokenOffset will not be 1.
		if p.isModuleVarDeclStart(1) {
			return p.parseModuleVarDeclWithQualifiersImpl(metadata, publicQualifier, qualifiers)
		}
		// fall through
	default:
		if p.isPossibleServiceDecl(qualifiers) {
			return p.parseServiceDeclOrVarDecl(metadata, publicQualifier, qualifiers)
		}

		if p.isTypeStartingToken(nextToken.Kind()) && nextToken.Kind() != common.IDENTIFIER_TOKEN {
			return p.parseModuleVarDeclWithQualifiersImpl(metadata, publicQualifier, qualifiers)
		}

		token := p.peek()
		solution := p.recover(token, PARSER_RULE_CONTEXT_TOP_LEVEL_NODE_WITHOUT_MODIFIER)

		if solution.action == ACTION_KEEP {
			// If the solution is KEEP, that means next immediate token is
			// at the correct place, but some token after that is not. There only one such
			// cases here, which is the `case IDENTIFIER_TOKEN`. So accept it, and continue.
			return p.parseModuleVarDeclWithQualifiersImpl(metadata, publicQualifier, qualifiers)
		}

		return p.parseTopLevelNodeWithMetadataQualifierAndQualifiers(metadata, publicQualifier, qualifiers)
	}
	return nil // should never reach here
}

func (p *Parser) parseMetadata() internal.STNode {
	var docString internal.STNode
	var annotations internal.STNode
	switch p.peek().Kind() {
	case common.DOCUMENTATION_STRING:
		docString = p.parseMarkdownDocumentation()
		annotations = p.parseOptionalAnnotations()
	case common.AT_TOKEN:
		docString = p.createEmptyNode()
		annotations = p.parseOptionalAnnotations()
	default:
		return p.createEmptyNode()
	}

	return p.createMetadata(docString, annotations)
}

func (p *Parser) isTypeStartingToken(kind common.SyntaxKind) bool {
	// TODO: implement proper check based on Java implementation
	// This should check if the token kind can start a type descriptor
	switch kind {
	case common.INT_KEYWORD, common.FLOAT_KEYWORD, common.STRING_KEYWORD, common.BOOLEAN_KEYWORD,
		common.DECIMAL_KEYWORD, common.ANY_KEYWORD, common.JSON_KEYWORD, common.XML_KEYWORD,
		common.ERROR_KEYWORD, common.MAP_KEYWORD, common.TABLE_KEYWORD, common.STREAM_KEYWORD,
		common.FUTURE_KEYWORD, common.TYPEDESC_KEYWORD, common.HANDLE_KEYWORD, common.READONLY_KEYWORD,
		common.NIL_TYPE_DESC, common.OPTIONAL_TYPE_DESC, common.ARRAY_TYPE_DESC,
		common.RECORD_TYPE_DESC, common.OBJECT_TYPE_DESC, common.UNION_TYPE_DESC,
		common.INTERSECTION_TYPE_DESC, common.FUNCTION_TYPE_DESC:
		return true
	default:
		return false
	}
}

// Stub methods for methods that will be implemented in later phases

func (p *Parser) isModuleVarDeclStart(lookahead int) bool {
	panic("unimplemented")
}

func (p *Parser) addMetadataNotAttachedDiagnostic(metadata internal.STNode) internal.STNode {
	panic("unimplemented")
}

func (p *Parser) createMissingSimpleVarDecl(metadata internal.STNode, isModuleVar bool) internal.STNode {
	panic("unimplemented")
}

func (p *Parser) createMissingSimpleVarDeclWithQualifiers(metadata internal.STNode, publicQualifier internal.STNode, qualifiers []internal.STNode, isModuleVar bool) internal.STNode {
	return p.createMissingSimpleVarDeclWithQualifiersImpl(metadata, publicQualifier, qualifiers, isModuleVar)
}

func (p *Parser) createMissingSimpleVarDeclWithQualifiersImpl(metadata internal.STNode, publicQualifier internal.STNode, qualifiers []internal.STNode, isModuleVar bool) internal.STNode {
	panic("unimplemented")
}

func (p *Parser) parseModuleVarDecl(metadata internal.STNode) internal.STNode {
	panic("unimplemented")
}

func (p *Parser) parseModuleVarDeclWithQualifiers(metadata internal.STNode, publicQualifier internal.STNode, qualifiers []internal.STNode) internal.STNode {
	return p.parseModuleVarDeclWithQualifiersImpl(metadata, publicQualifier, qualifiers)
}

func (p *Parser) parseModuleVarDeclWithQualifiersImpl(metadata internal.STNode, publicQualifier internal.STNode, qualifiers []internal.STNode) internal.STNode {
	panic("unimplemented")
}

func (p *Parser) reportInvalidQualifier(qualifier internal.STNode) {
	panic("unimplemented")
}

func (p *Parser) reportInvalidMetaData(metadata internal.STNode, constructName string) {
	panic("unimplemented")
}

func (p *Parser) reportInvalidQualifierList(qualifiers []internal.STNode) {
	panic("unimplemented")
}

func (p *Parser) parseTopLevelQualifiers(qualifiers *[]internal.STNode) {
	panic("unimplemented")
}

func (p *Parser) parseFuncDefOrFuncTypeDesc(metadata internal.STNode, visibilityQualifier internal.STNode, qualifiers []internal.STNode, isObjectMember bool, isClassMember bool) internal.STNode {
	panic("unimplemented")
}

func (p *Parser) parseModuleTypeDefinition(metadata internal.STNode, qualifier internal.STNode) internal.STNode {
	panic("unimplemented")
}

func (p *Parser) parseClassDefinition(metadata internal.STNode, qualifier internal.STNode, qualifiers []internal.STNode) internal.STNode {
	panic("unimplemented")
}

func (p *Parser) parseListenerDeclaration(metadata internal.STNode, publicQualifier internal.STNode) internal.STNode {
	panic("unimplemented")
}

func (p *Parser) parseConstantDeclaration(metadata internal.STNode, publicQualifier internal.STNode) internal.STNode {
	panic("unimplemented")
}

func (p *Parser) parseAnnotationDeclaration(metadata internal.STNode, publicQualifier internal.STNode, constKeyword internal.STNode) internal.STNode {
	panic("unimplemented")
}

func (p *Parser) parseImportDecl() internal.STNode {
	panic("unimplemented")
}

func (p *Parser) parseXMLNamespaceDeclaration(isTopLevel bool) internal.STNode {
	panic("unimplemented")
}

func (p *Parser) parseEnumDeclaration(metadata internal.STNode, publicQualifier internal.STNode) internal.STNode {
	panic("unimplemented")
}

func (p *Parser) isPossibleServiceDecl(qualifiers []internal.STNode) bool {
	panic("unimplemented")
}

func (p *Parser) parseServiceDeclOrVarDecl(metadata internal.STNode, publicQualifier internal.STNode, qualifiers []internal.STNode) internal.STNode {
	panic("unimplemented")
}

func (p *Parser) createMetadata(docString internal.STNode, annotations internal.STNode) internal.STNode {
	if annotations == nil && docString == nil {
		return p.createEmptyNode()
	}
	// TODO: implement STMetadataNode creation
	panic("unimplemented: createMetadataNode")
}

func (p *Parser) parseMarkdownDocumentation() internal.STNode {
	panic("unimplemented")
}

func (p *Parser) parseOptionalAnnotations() internal.STNode {
	panic("unimplemented")
}
