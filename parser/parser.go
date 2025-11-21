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
	"ballerina-lang-go/parser/common"
	"ballerina-lang-go/parser/internal"
	"ballerina-lang-go/tools/diagnostics"
)

// FIXME: add these
type Solution struct {
	action        Action
	removedToken  internal.STToken
	recoveredNode internal.STNode
}

type Action uint8

const (
	ActionINSERT Action = iota
	ActionREMOVE
	ActionKEEP
)

type ParserErrorHandler interface {
	ReportError(errorCode common.DiagnosticErrorCode, args ...any)
	switchContext(context common.ParserRuleContext)
	getParentContext() common.ParserRuleContext
	endContext()
	startContext(context common.ParserRuleContext)
	recover(currentCtx common.ParserRuleContext, token internal.STToken, isCompletion bool) Solution
}

type invalidNodeInfo struct {
	node           internal.STNode
	diagnosticCode diagnostics.DiagnosticCode
	args           []interface{}
}

// FIXME: make this private
type AbstractParser struct {
	errorHandler         ParserErrorHandler
	tokenReader          TokenReader
	invalidNodeInfoStack []invalidNodeInfo
	insertedToken        internal.STToken
}

func NewInvalidNodeInfoFromInvalidNodeDiagnosticCodeArgs(invalidNode internal.STNode, diagnosticCode diagnostics.DiagnosticCode, args ...interface{}) invalidNodeInfo {
	this := invalidNodeInfo{}
	this.node = invalidNode
	this.diagnosticCode = diagnosticCode
	this.args = args
	return this
}

func NewAbstractParserFromTokenReaderErrorHandler(tokenReader TokenReader, errorHandler ParserErrorHandler) AbstractParser {
	this := AbstractParser{}
	this.invalidNodeInfoStack = make([]invalidNodeInfo, 0)
	this.insertedToken = nil
	// Default field initializations

	this.tokenReader = tokenReader
	this.errorHandler = errorHandler
	return this
}

func NewAbstractParserFromTokenReader(tokenReader TokenReader) AbstractParser {
	this := AbstractParser{}
	this.invalidNodeInfoStack = make([]invalidNodeInfo, 0)
	this.insertedToken = nil
	// Default field initializations

	this.tokenReader = tokenReader
	this.errorHandler = nil
	return this
}

func (this *AbstractParser) peek() internal.STToken {
	if this.insertedToken != nil {
		return this.insertedToken
	}
	return this.peek()
}

func (this *AbstractParser) peekN(n int) internal.STToken {
	if this.insertedToken == nil {
		return this.peekN(n)
	}
	if n == 1 {
		return this.insertedToken
	}
	if n > 0 {
		n = (n - 1)
	}
	return this.peekN(n)
}

func (this *AbstractParser) consume() internal.STToken {
	if this.insertedToken != nil {
		nextToken := this.insertedToken
		this.insertedToken = nil
		return nextToken
	}
	if len(this.invalidNodeInfoStack) == 0 {
		return this.tokenReader.Read()
	}
	return this.consumeWithInvalidNodes()
}

func (this *AbstractParser) consumeWithInvalidNodes() internal.STToken {
	token := this.tokenReader.Read()
	return this.consumeWithInvalidNodesWithToken(token)
}

func (this *AbstractParser) consumeWithInvalidNodesWithToken(token internal.STToken) internal.STToken {
	newToken := token
	for len(this.invalidNodeInfoStack) > 0 {
		invalidNodeInfo := this.invalidNodeInfoStack[len(this.invalidNodeInfoStack)-1]
		this.invalidNodeInfoStack = this.invalidNodeInfoStack[:len(this.invalidNodeInfoStack)-1]
		newToken = internal.ToToken(internal.CloneWithLeadingInvalidNodeMinutiae(newToken, invalidNodeInfo.node,
			invalidNodeInfo.diagnosticCode, invalidNodeInfo.args))
	}
	return newToken
}

func (this *AbstractParser) recover(token internal.STToken, currentCtx common.ParserRuleContext, isCompletion bool) Solution {
	isCompletion = isCompletion || token.Kind() == common.EOF_TOKEN
	sol := this.errorHandler.recover(currentCtx, token, isCompletion)
	if sol.action == ActionREMOVE {
		this.insertedToken = nil
		this.addInvalidTokenToNextToken(sol.removedToken)
	} else if sol.action == ActionINSERT {
		this.insertedToken = internal.ToToken(sol.recoveredNode)
	}
	return sol
}

func (this *AbstractParser) insertToken(kind common.SyntaxKind, context common.ParserRuleContext) {
	this.insertedToken = internal.CreateMissingTokenWithDiagnosticsFromParserRules(kind, context)
}

func (this *AbstractParser) removeInsertedToken() {
	this.insertedToken = nil
}

func (this *AbstractParser) isInvalidNodeStackEmpty() bool {
	return len(this.invalidNodeInfoStack) == 0
}

func (this *AbstractParser) startContext(context common.ParserRuleContext) {
	this.errorHandler.startContext(context)
}

func (this *AbstractParser) endContext() {
	this.errorHandler.endContext()
}

func (this *AbstractParser) getCurrentContext() common.ParserRuleContext {
	return this.errorHandler.getParentContext()
}

func (this *AbstractParser) switchContext(context common.ParserRuleContext) {
	this.errorHandler.switchContext(context)
}

func (this *AbstractParser) getNextNextToken() internal.STToken {
	return this.peekN(2)
}

func (this *AbstractParser) isNodeListEmpty(node internal.STNode) bool {
	nodeList, ok := node.(internal.STNodeList)
	if !ok {
		panic("node is not a STNodeList")
	}
	return nodeList.IsEmpty()
}

func (this *AbstractParser) cloneWithDiagnosticIfListEmpty(nodeList internal.STNode, target internal.STNode, diagnosticCode diagnostics.DiagnosticCode) internal.STNode {
	if this.isNodeListEmpty(nodeList) {
		return internal.AddDiagnostic(target, diagnosticCode)
	}
	return target
}

func (this *AbstractParser) updateLastNodeInListWithInvalidNode(nodeList []internal.STNode, invalidParam internal.STNode, diagnosticCode diagnostics.DiagnosticCode, args ...interface{}) []internal.STNode {
	prevNode := nodeList[len(nodeList)-1]
	nodeList = nodeList[:len(nodeList)-1]
	newNode := internal.CloneWithTrailingInvalidNodeMinutiae(prevNode, invalidParam, diagnosticCode, args)
	nodeList = append(nodeList, newNode)
	return nodeList
}

func (this *AbstractParser) updateFirstNodeInListWithLeadingInvalidNode(nodeList []internal.STNode, invalidParam internal.STNode, diagnosticCode diagnostics.DiagnosticCode, args ...interface{}) []internal.STNode {
	return this.updateANodeInListWithLeadingInvalidNode(nodeList, 0, invalidParam, diagnosticCode, args)
}

func (this *AbstractParser) updateANodeInListWithLeadingInvalidNode(nodeList []internal.STNode, indexOfTheNode int, invalidParam internal.STNode, diagnosticCode diagnostics.DiagnosticCode, args ...interface{}) []internal.STNode {
	node := nodeList[indexOfTheNode]
	newNode := internal.CloneWithLeadingInvalidNodeMinutiae(node, invalidParam, diagnosticCode, args)
	nodeList[indexOfTheNode] = newNode
	return nodeList
}

func (this *AbstractParser) invalidateRestAndAddToTrailingMinutiae(node internal.STNode) internal.STNode {
	node = this.addInvalidNodeStackToTrailingMinutiae(node)
	for this.peek().Kind() != common.EOF_TOKEN {
		invalidToken := this.consume()
		node = internal.CloneWithTrailingInvalidNodeMinutiae(node, invalidToken, &common.ERROR_INVALID_TOKEN, invalidToken.Text())
	}
	return node
}

func (this *AbstractParser) addInvalidNodeStackToTrailingMinutiae(node internal.STNode) internal.STNode {
	for len(this.invalidNodeInfoStack) != 0 {
		invalidNodeInfo := this.invalidNodeInfoStack[len(this.invalidNodeInfoStack)-1]
		this.invalidNodeInfoStack = this.invalidNodeInfoStack[:len(this.invalidNodeInfoStack)-1]
		node = internal.CloneWithTrailingInvalidNodeMinutiae(node, invalidNodeInfo.node, invalidNodeInfo.diagnosticCode, invalidNodeInfo.args)
	}
	return node
}

func (this *AbstractParser) addInvalidNodeToNextToken(invalidNode internal.STNode, diagnosticCode diagnostics.DiagnosticCode, args ...interface{}) {
	this.invalidNodeInfoStack = append(this.invalidNodeInfoStack, invalidNodeInfo{node: invalidNode, diagnosticCode: diagnosticCode, args: args})
}

func (this *AbstractParser) addInvalidTokenToNextToken(invalidNode internal.STToken) {
	this.invalidNodeInfoStack = append(this.invalidNodeInfoStack, invalidNodeInfo{node: invalidNode, diagnosticCode: &common.ERROR_INVALID_TOKEN, args: []interface{}{invalidNode.Text()}})
}

type BallerinaParser struct {
	AbstractParser
}

func newBallerinaParserFromTokenReader(tokenReader TokenReader) BallerinaParser {
	this := BallerinaParser{}
	// Default field initializations

	this.AbstractParser = AbstractParser{
		tokenReader: tokenReader,
		// FIXME:
		errorHandler: nil,
	}
	return this
}

func isParameterizedTypeToken(tokenKind common.SyntaxKind) bool {
	switch tokenKind {
	case common.TYPEDESC_KEYWORD, common.FUTURE_KEYWORD, common.XML_KEYWORD, common.ERROR_KEYWORD:
		return true
	default:
		return false
	}
}

func CreateBuiltinSimpleNameReference(token internal.STNode) internal.STNode {
	typeKind := this.getBuiltinTypeSyntaxKind(token.kind)
	return this.STNodeFactory.createBuiltinSimpleNameReferenceNode(typeKind, token)
}

func isCompoundBinaryOperator(tokenKind common.SyntaxKind) bool {
	switch tokenKind {
	case common.PLUS_TOKEN,
		common.MINUS_TOKEN,
		common.SLASH_TOKEN,
		common.ASTERISK_TOKEN,
		common.BITWISE_AND_TOKEN,
		common.BITWISE_XOR_TOKEN,
		common.PIPE_TOKEN,
		common.DOUBLE_LT_TOKEN,
		common.DOUBLE_GT_TOKEN,
		common.TRIPPLE_GT_TOKEN:
		return true
	default:
		return false
	}
}

func isTypeStartingToken(nextTokenKind common.SyntaxKind, nextNextToken internal.STToken) bool {
	switch nextTokenKind {
	case common.IDENTIFIER_TOKEN,
		common.SERVICE_KEYWORD,
		common.RECORD_KEYWORD,
		common.OBJECT_KEYWORD,
		common.ABSTRACT_KEYWORD,
		common.CLIENT_KEYWORD,
		common.OPEN_PAREN_TOKEN,
		common.MAP_KEYWORD,
		common.STREAM_KEYWORD,
		common.TABLE_KEYWORD,
		common.FUNCTION_KEYWORD,
		common.OPEN_BRACKET_TOKEN,
		common.DISTINCT_KEYWORD,
		common.ISOLATED_KEYWORD,
		common.TRANSACTIONAL_KEYWORD,
		common.TRANSACTION_KEYWORD,
		common.NATURAL_KEYWORD:
		return true
	default:
		if isParameterizedTypeToken(nextTokenKind) {
			return true
		}
		if isSingletonTypeDescStart(nextTokenKind, nextNextToken) {
			return true
		}
		return isSimpleType(nextTokenKind)
	}
}

func isSimpleType(nodeKind common.SyntaxKind) bool {
	switch nodeKind {
	case common.INT_KEYWORD,
		common.FLOAT_KEYWORD,
		common.DECIMAL_KEYWORD,
		common.BOOLEAN_KEYWORD,
		common.STRING_KEYWORD,
		common.BYTE_KEYWORD,
		common.JSON_KEYWORD,
		common.HANDLE_KEYWORD,
		common.ANY_KEYWORD,
		common.ANYDATA_KEYWORD,
		common.NEVER_KEYWORD,
		common.VAR_KEYWORD,
		common.READONLY_KEYWORD:
		return true
	default:
		return false
	}
}

func isPredeclaredPrefix(nodeKind common.SyntaxKind) bool {
	switch nodeKind {
	case common.BOOLEAN_KEYWORD,
		common.DECIMAL_KEYWORD,
		common.ERROR_KEYWORD,
		common.FLOAT_KEYWORD,
		common.FUNCTION_KEYWORD,
		common.FUTURE_KEYWORD,
		common.INT_KEYWORD,
		common.MAP_KEYWORD,
		common.NATURAL_KEYWORD,
		common.OBJECT_KEYWORD,
		common.STREAM_KEYWORD,
		common.STRING_KEYWORD,
		common.TABLE_KEYWORD,
		common.TRANSACTION_KEYWORD,
		common.TYPEDESC_KEYWORD,
		common.XML_KEYWORD:
		return true
	default:
		return false
	}
}

func getBuiltinTypeSyntaxKind(typeKeyword common.SyntaxKind) common.SyntaxKind {
	switch typeKeyword {
	case common.INT_KEYWORD:
		return common.INT_TYPE_DESC
	case common.FLOAT_KEYWORD:
		return common.FLOAT_TYPE_DESC
	case common.DECIMAL_KEYWORD:
		return common.DECIMAL_TYPE_DESC
	case common.BOOLEAN_KEYWORD:
		return common.BOOLEAN_TYPE_DESC
	case common.STRING_KEYWORD:
		return common.STRING_TYPE_DESC
	case common.BYTE_KEYWORD:
		return common.BYTE_TYPE_DESC
	case common.JSON_KEYWORD:
		return common.JSON_TYPE_DESC
	case common.HANDLE_KEYWORD:
		return common.HANDLE_TYPE_DESC
	case common.ANY_KEYWORD:
		return common.ANY_TYPE_DESC
	case common.ANYDATA_KEYWORD:
		return common.ANYDATA_TYPE_DESC
	case common.NEVER_KEYWORD:
		return common.NEVER_TYPE_DESC
	case common.VAR_KEYWORD:
		return common.VAR_TYPE_DESC
	case common.READONLY_KEYWORD:
		return common.READONLY_TYPE_DESC
	default:
		panic(typeKeyword.StrValue() + "is not a built-in type")
	}
}

func isKeyKeyword(token internal.STToken) bool {
	return ((token.Kind() == common.IDENTIFIER_TOKEN) && KEY == token.Text())
}

func isNaturalKeyword(token internal.STToken) bool {
	return ((token.Kind() == common.IDENTIFIER_TOKEN) && NATURAL == (token.Text()))
}

func isEndOfLetVarDeclarations(nextToken internal.STToken, nextNextToken internal.STToken) bool {
	tokenKind := nextToken.Kind()
	switch tokenKind {
	case common.COMMA_TOKEN, common.AT_TOKEN:
		return false
	case common.IN_KEYWORD:
		return true
	default:
		return (isGroupOrCollectKeyword(nextToken) || (!isTypeStartingToken(tokenKind, nextNextToken)))
	}
}

func isGroupOrCollectKeyword(nextToken internal.STToken) bool {
	return (isKeywordMatch(common.COLLECT_KEYWORD, nextToken) || isKeywordMatch(common.GROUP_KEYWORD, nextToken))
}

func isKeywordMatch(syntaxKind common.SyntaxKind, token internal.STToken) bool {
	return ((token.Kind() == common.IDENTIFIER_TOKEN) && syntaxKind.StrValue() == (token.Text()))
}

func isSingletonTypeDescStart(tokenKind common.SyntaxKind, nextNextToken internal.STToken) bool {
	switch tokenKind {
	case common.STRING_LITERAL_TOKEN,
		common.DECIMAL_INTEGER_LITERAL_TOKEN,
		common.HEX_INTEGER_LITERAL_TOKEN,
		common.DECIMAL_FLOATING_POINT_LITERAL_TOKEN,
		common.HEX_FLOATING_POINT_LITERAL_TOKEN,
		common.TRUE_KEYWORD,
		common.FALSE_KEYWORD,
		common.NULL_KEYWORD:
		return true
	case common.PLUS_TOKEN,
		common.MINUS_TOKEN:
		return isIntOrFloat(nextNextToken)
	default:
		return false
	}
}

func isIntOrFloat(token internal.STToken) bool {
	switch token.Kind() {
	case common.DECIMAL_INTEGER_LITERAL_TOKEN,
		common.HEX_INTEGER_LITERAL_TOKEN,
		common.DECIMAL_FLOATING_POINT_LITERAL_TOKEN,
		common.HEX_FLOATING_POINT_LITERAL_TOKEN:
		return true
	default:
		return false
	}
}

func isValidBase16LiteralContent(content string) bool {
	hexDigitCount := 0
	charArray := []byte(content)
	for _, c := range charArray {
		switch c {
		case TAB,
			NEWLINE,
			CARRIAGE_RETURN,
			SPACE:
			break
		default:
			if isHexDigit(c) {
				hexDigitCount++
			} else {
				return false
			}
			break
		}
	}
	return ((hexDigitCount % 2) == 0)
}

func isValidBase64LiteralContent(content string) bool {
	charArray := []byte(content)
	base64CharCount := 0
	paddingCharCount := 0
	for _, c := range charArray {
		switch c {
		case TAB,
			NEWLINE,
			CARRIAGE_RETURN,
			SPACE:
			break
		case EQUAL:
			paddingCharCount++
			break
		default:
			if isBase64Char(c) {
				if paddingCharCount == 0 {
					base64CharCount++
				} else {
					return false
				}
			} else {
				return false
			}
			break
		}
	}
	if paddingCharCount > 2 {
		return false
	} else if paddingCharCount == 0 {
		return ((base64CharCount % 4) == 0)
	} else {
		return base64CharCount%4 == 4-paddingCharCount
	}
}

func isBase64Char(c byte) bool {
	if ('a' <= c) && (c <= 'z') {
		return true
	}
	if ('A' <= c) && (c <= 'Z') {
		return true
	}
	if (c == '+') || (c == '/') {
		return true
	}
	return isDigit(c)
}

func isHexDigit(c byte) bool {
	if ('a' <= c) && (c <= 'f') {
		return true
	}
	if ('A' <= c) && (c <= 'F') {
		return true
	}
	return isDigit(c)
}

func isDigit(c byte) bool {
	return (('0' <= c) && (c <= '9'))
}

func (this *BallerinaParser) Parse() internal.STNode {
	return this.parseCompUnit()
}

func (this *BallerinaParser) ParseAsStatement() internal.STNode {
	this.startContext(common.PARSER_RULE_CONTEXT_COMP_UNIT)
	this.startContext(common.PARSER_RULE_CONTEXT_FUNC_DEF)
	this.startContext(common.PARSER_RULE_CONTEXT_FUNC_BODY_BLOCK)
	stmt := this.parseStatement()
	if (stmt == nil) || this.validateStatement(stmt) {
		stmt = this.createMissingSimpleVarDecl(false)
		stmt = this.invalidateRestAndAddToTrailingMinutiae(stmt)
		return stmt
	}
	if stmt.Kind() == common.NAMED_WORKER_DECLARATION {
		this.addInvalidNodeToNextToken(stmt, &common.ERROR_NAMED_WORKER_NOT_ALLOWED_HERE)
		stmt = this.createMissingSimpleVarDecl(false)
		stmt = this.invalidateRestAndAddToTrailingMinutiae(stmt)
		return stmt
	}
	stmt = this.invalidateRestAndAddToTrailingMinutiae(stmt)
	return stmt
}

func (this *BallerinaParser) ParseAsBlockStatement() internal.STNode {
	this.startContext(common.PARSER_RULE_CONTEXT_COMP_UNIT)
	this.startContext(common.PARSER_RULE_CONTEXT_FUNC_DEF)
	this.startContext(common.PARSER_RULE_CONTEXT_FUNC_BODY_BLOCK)
	this.startContext(common.PARSER_RULE_CONTEXT_WHILE_BLOCK)
	blockStmtNode := this.parseBlockNode()
	blockStmtNode = this.invalidateRestAndAddToTrailingMinutiae(blockStmtNode)
	return blockStmtNode
}

func (this *BallerinaParser) ParseAsStatements() internal.STNode {
	this.startContext(common.PARSER_RULE_CONTEXT_COMP_UNIT)
	this.startContext(common.PARSER_RULE_CONTEXT_FUNC_BODY_BLOCK)
	stmtsNode := this.parseStatements()
	stmtNodeList, ok := stmtsNode.(internal.STNodeList)
	if !ok {
		panic("stmtsNode is not a STNodeList")
	}
	var stmts []internal.STNode
	for i := 0; i < (stmtNodeList.Size() - 1); i++ {
		stmts = append(stmts, stmtNodeList.Get(i))
	}
	var lastStmt internal.STNode
	if stmtNodeList.Size() == 0 {
		lastStmt = this.createMissingSimpleVarDecl(false)
	} else {
		lastStmt = stmtNodeList.Get(stmtNodeList.Size() - 1)
	}
	lastStmt = this.invalidateRestAndAddToTrailingMinutiae(lastStmt)
	stmts = append(stmts, lastStmt)
	return internal.CreateNodeList(stmts)
}

func (this *BallerinaParser) ParseAsExpression() internal.STNode {
	this.startContext(common.PARSER_RULE_CONTEXT_COMP_UNIT)
	this.startContext(common.PARSER_RULE_CONTEXT_VAR_DECL_STMT)
	expr := this.parseExpression()
	expr = this.invalidateRestAndAddToTrailingMinutiae(expr)
	return expr
}

func (this *BallerinaParser) ParseAsActionOrExpression() internal.STNode {
	this.startContext(common.PARSER_RULE_CONTEXT_COMP_UNIT)
	this.startContext(common.PARSER_RULE_CONTEXT_FUNC_DEF)
	this.startContext(common.PARSER_RULE_CONTEXT_FUNC_BODY_BLOCK)
	this.startContext(common.PARSER_RULE_CONTEXT_VAR_DECL_STMT)
	actionOrExpr := this.parseActionOrExpression()
	actionOrExpr = this.invalidateRestAndAddToTrailingMinutiae(actionOrExpr)
	return actionOrExpr
}

func (this *BallerinaParser) ParseAsModuleMemberDeclaration() internal.STNode {
	this.startContext(common.PARSER_RULE_CONTEXT_COMP_UNIT)
	topLevelNode := this.parseTopLevelNode()
	if topLevelNode == nil {
		topLevelNode = this.createMissingSimpleVarDecl(true)
	}
	if topLevelNode.Kind() == common.IMPORT_DECLARATION {
		temp := topLevelNode
		topLevelNode = this.createMissingSimpleVarDecl(true)
		topLevelNode = internal.CloneWithTrailingInvalidNodeMinutiaeWithoutDiagnostics(topLevelNode, temp)
	}
	topLevelNode = this.invalidateRestAndAddToTrailingMinutiae(topLevelNode)
	return topLevelNode
}

func (this *BallerinaParser) ParseAsImportDeclaration() internal.STNode {
	this.startContext(common.PARSER_RULE_CONTEXT_COMP_UNIT)
	importDecl := this.parseImportDecl()
	importDecl = this.invalidateRestAndAddToTrailingMinutiae(importDecl)
	return importDecl
}

func (this *BallerinaParser) ParseAsTypeDescriptor() internal.STNode {
	this.startContext(common.PARSER_RULE_CONTEXT_COMP_UNIT)
	this.startContext(common.PARSER_RULE_CONTEXT_MODULE_TYPE_DEFINITION)
	typeDesc := this.parseTypeDescriptor(common.PARSER_RULE_CONTEXT_TYPE_DESC_IN_TYPE_DEF)
	typeDesc = this.invalidateRestAndAddToTrailingMinutiae(typeDesc)
	return typeDesc
}

func (this *BallerinaParser) ParseAsBindingPattern() internal.STNode {
	this.startContext(common.PARSER_RULE_CONTEXT_COMP_UNIT)
	this.startContext(common.PARSER_RULE_CONTEXT_VAR_DECL_STMT)
	bindingPattern := this.parseBindingPattern()
	bindingPattern = this.invalidateRestAndAddToTrailingMinutiae(bindingPattern)
	return bindingPattern
}

func (this *BallerinaParser) ParseAsFunctionBodyBlock() internal.STNode {
	this.startContext(common.PARSER_RULE_CONTEXT_COMP_UNIT)
	this.startContext(common.PARSER_RULE_CONTEXT_FUNC_DEF)
	funcBodyBlock := this.parseFunctionBodyBlock(false)
	funcBodyBlock = this.invalidateRestAndAddToTrailingMinutiae(funcBodyBlock)
	return funcBodyBlock
}

func (this *BallerinaParser) ParseAsObjectMember() internal.STNode {
	this.startContext(common.PARSER_RULE_CONTEXT_COMP_UNIT)
	this.startContext(common.PARSER_RULE_CONTEXT_SERVICE_DECL)
	this.startContext(common.PARSER_RULE_CONTEXT_OBJECT_CONSTRUCTOR_MEMBER)
	objectMember := this.parseObjectMember(common.PARSER_RULE_CONTEXT_OBJECT_CONSTRUCTOR_MEMBER)
	if objectMember == nil {
		objectMember = this.createMissingSimpleObjectFieldDefault()
	}
	objectMember = this.invalidateRestAndAddToTrailingMinutiae(objectMember)
	return objectMember
}

func (this *BallerinaParser) ParseAsIntermediateClause(allowActions bool) internal.STNode {
	this.startContext(common.PARSER_RULE_CONTEXT_COMP_UNIT)
	this.startContext(common.PARSER_RULE_CONTEXT_FUNC_DEF)
	this.startContext(common.PARSER_RULE_CONTEXT_FUNC_BODY_BLOCK)
	this.startContext(common.PARSER_RULE_CONTEXT_VAR_DECL_STMT)
	this.startContext(common.PARSER_RULE_CONTEXT_QUERY_EXPRESSION)
	var intermediateClause internal.STNode
	if !this.isEndOfIntermediateClause(peek().kind) {
		intermediateClause = this.parseIntermediateClause(true, allowActions)
	}
	if intermediateClause == nil {
		intermediateClause = this.createMissingWhereClause()
	}
	if intermediateClause.Kind() == common.SELECT_CLAUSE {
		temp := intermediateClause
		intermediateClause = this.createMissingWhereClause()
		intermediateClause = internal.CloneWithTrailingInvalidNodeMinutiae(intermediateClause, temp)
	}
	intermediateClause = this.invalidateRestAndAddToTrailingMinutiae(intermediateClause)
	return intermediateClause
}

func (this *BallerinaParser) ParseAsLetVarDeclaration(allowActions bool) internal.STNode {
	this.startContext(common.PARSER_RULE_CONTEXT_COMP_UNIT)
	this.startContext(common.PARSER_RULE_CONTEXT_VAR_DECL_STMT)
	this.switchContext(common.PARSER_RULE_CONTEXT_QUERY_EXPRESSION)
	this.switchContext(common.PARSER_RULE_CONTEXT_LET_CLAUSE_LET_VAR_DECL)
	letVarDeclaration := this.parseLetVarDecl(common.PARSER_RULE_CONTEXT_LET_CLAUSE_LET_VAR_DECL, true, allowActions)
	letVarDeclaration = this.invalidateRestAndAddToTrailingMinutiae(letVarDeclaration)
	return letVarDeclaration
}

func (this *BallerinaParser) ParseAsAnnotation() internal.STNode {
	this.startContext(common.PARSER_RULE_CONTEXT_COMP_UNIT)
	this.startContext(common.PARSER_RULE_CONTEXT_ANNOTATIONS)
	annotation := this.parseAnnotation()
	annotation = this.invalidateRestAndAddToTrailingMinutiae(annotation)
	return annotation
}

func (this *BallerinaParser) ParseAsMarkdownDocumentation() internal.STNode {
	this.startContext(common.PARSER_RULE_CONTEXT_COMP_UNIT)
	markdownDoc := this.parseMarkdownDocumentation()
	if this.markdownDoc.toSourceCode().isEmpty() {
		missingHash := internal.CreateMissingTokenWithDiagnostics(common.HASH_TOKEN,
			DiagnosticWarningCode.WARNING_MISSING_HASH_TOKEN)
		docLine := internal.CreateMarkdownDocumentationLineNode(common.MARKDOWN_DOCUMENTATION_LINE,
			missingHash, internal.CreateEmptyNodeList())
		markdownDoc = internal.CreateMarkdownDocumentationNode(internal.CreateNodeList(docLine))
	}
	markdownDoc = this.invalidateRestAndAddToTrailingMinutiae(markdownDoc)
	return markdownDoc
}

func (this *BallerinaParser) ParseWithContextx(context common.ParserRuleContext) internal.STNode {
	switch context {
	case common.PARSER_RULE_CONTEXT_COMP_UNIT:
		return this.parseCompUnit()
	case common.PARSER_RULE_CONTEXT_TOP_LEVEL_NODE:
		this.startContext(common.PARSER_RULE_CONTEXT_COMP_UNIT)
		return this.parseTopLevelNode()
	case common.PARSER_RULE_CONTEXT_STATEMENT:
		this.startContext(common.PARSER_RULE_CONTEXT_COMP_UNIT)
		this.startContext(common.PARSER_RULE_CONTEXT_FUNC_BODY_BLOCK)
		return this.parseStatement()
	case common.PARSER_RULE_CONTEXT_EXPRESSION:
		this.startContext(common.PARSER_RULE_CONTEXT_COMP_UNIT)
		this.startContext(common.PARSER_RULE_CONTEXT_VAR_DECL_STMT)
		return this.parseExpression()
	default:
		panic("Cannot start parsing from: " + context.String())
	}
}

func (this *BallerinaParser) parseCompUnit() internal.STNode {
	this.startContext(common.PARSER_RULE_CONTEXT_COMP_UNIT)
	var otherDecls []internal.STNode
	var importDecls []internal.STNode
	processImports := true
	token := this.peek()
	for token.Kind() != common.EOF_TOKEN {
		decl := this.parseTopLevelNode()
		if decl == nil {
			break
		}
		if decl.Kind() == common.IMPORT_DECLARATION {
			if processImports {
				importDecls = append(importDecls, decl)
			} else {
				this.updateLastNodeInListWithInvalidNode(otherDecls, decl,
					&common.ERROR_IMPORT_DECLARATION_AFTER_OTHER_DECLARATIONS)
			}
		} else {
			if processImports {
				processImports = false
			}
			otherDecls = append(otherDecls, decl)
		}
		token = this.peek()
	}
	eof := this.consume()
	this.endContext()
	return internal.CreateModulePartNode(internal.CreateNodeList(importDecls), internal.CreateNodeList(otherDecls), eof)
}

func (this *BallerinaParser) parseTopLevelNode() internal.STNode {
	nextToken := this.peek()
	var metadata internal.STNode
	switch nextToken.Kind() {
	case common.EOF_TOKEN:
		return nil
	case common.DOCUMENTATION_STRING, common.AT_TOKEN:
		metadata = this.parseMetaData()
		return this.parseTopLevelNodeWithMetadata(metadata)
	case common.IMPORT_KEYWORD,
		common.FINAL_KEYWORD,
		common.PUBLIC_KEYWORD,
		common.FUNCTION_KEYWORD,
		common.TYPE_KEYWORD,
		common.LISTENER_KEYWORD,
		common.CONST_KEYWORD,
		common.ANNOTATION_KEYWORD,
		common.XMLNS_KEYWORD,
		common.ENUM_KEYWORD,
		common.CLASS_KEYWORD,
		common.TRANSACTIONAL_KEYWORD,
		common.ISOLATED_KEYWORD,
		common.DISTINCT_KEYWORD,
		common.CLIENT_KEYWORD,
		common.READONLY_KEYWORD,
		common.CONFIGURABLE_KEYWORD,
		common.SERVICE_KEYWORD:
		metadata = internal.CreateEmptyNode()
		break
	case common.RESOURCE_KEYWORD:
	case common.REMOTE_KEYWORD:
		this.reportInvalidQualifier(this.consume())
		return this.parseTopLevelNode()
	case common.IDENTIFIER_TOKEN:
		if this.isModuleVarDeclStart(1) || this.nextToken.isMissing() {
			return this.parseModuleVarDecl(internal.CreateEmptyNode())
		}
	default:
		if isTypeStartingToken(nextToken.Kind()) && (nextToken.Kind() != common.IDENTIFIER_TOKEN) {
			metadata = internal.CreateEmptyNode()
			break
		}
		token := this.peek()
		solution := this.recover(token, common.PARSER_RULE_CONTEXT_TOP_LEVEL_NODE, this.isInsideABlock(token))
		if solution.action == ActionKEEP {
			metadata = internal.CreateEmptyNode()
			break
		}
		return this.parseTopLevelNodeWithMetadata(metadata)
	}
	return this.parseTopLevelNodeWithMetadata(metadata)
}

func (this *BallerinaParser) parseTopLevelNodeWithMetadata(metadata internal.STNode) internal.STNode {
	nextToken := this.peek()
	var publicQualifier internal.STNode
	switch nextToken.Kind() {
	case common.EOF_TOKEN:
		if metadata != nil {
			metadaNode, ok := metadata.(internal.STMetadataNode)
			if !ok {
				panic("metadata is not a STMetadataNode")
			}
			metadata = this.addMetadataNotAttachedDiagnostic(metadaNode)
			return this.createMissingSimpleVarDeclInner(metadata, true)
		}
		return nil
	case common.PUBLIC_KEYWORD:
		publicQualifier = this.consume()
	case common.FUNCTION_KEYWORD,
		common.TYPE_KEYWORD,
		common.LISTENER_KEYWORD,
		common.CONST_KEYWORD,
		common.FINAL_KEYWORD,
		common.IMPORT_KEYWORD,
		common.ANNOTATION_KEYWORD,
		common.XMLNS_KEYWORD,
		common.ENUM_KEYWORD,
		common.CLASS_KEYWORD,
		common.TRANSACTIONAL_KEYWORD,
		common.ISOLATED_KEYWORD,
		common.DISTINCT_KEYWORD,
		common.CLIENT_KEYWORD,
		common.READONLY_KEYWORD,
		common.SERVICE_KEYWORD,
		common.CONFIGURABLE_KEYWORD:
		break
	case common.RESOURCE_KEYWORD, common.REMOTE_KEYWORD:
		this.reportInvalidQualifier(this.consume())
		return this.parseTopLevelNodeWithMetadata(metadata)
	case common.IDENTIFIER_TOKEN:
		if this.isModuleVarDeclStart(1) {
			return this.parseModuleVarDecl(metadata)
		}
	default:
		if isTypeStartingToken(nextToken.Kind()) && (nextToken.Kind() != common.IDENTIFIER_TOKEN) {
			break
		}
		token := this.peek()
		solution := this.recover(token, common.PARSER_RULE_CONTEXT_TOP_LEVEL_NODE_WITHOUT_METADATA)
		if solution.action == ActionKEEP {
			publicQualifier = internal.CreateEmptyNode()
			break
		}
		return this.parseTopLevelNodeWithMetadata(metadata)
	}
	return this.parseTopLevelNodeWithMetadata(metadata, publicQualifier)
}

func (this *BallerinaParser) addMetadataNotAttachedDiagnostic(metadata internal.STMetadataNode) internal.STNode {
	docString := metadata.DocumentationString
	if docString != nil {
		docString = internal.AddDiagnostic(docString, common.ERROR_DOCUMENTATION_NOT_ATTACHED_TO_A_CONSTRUCT)
	}
	annotList := internal.STNodeList(metadata.annotations)
	annotations := this.addAnnotNotAttachedDiagnostic(annotList)
	return internal.CreateMetadataNode(docString, annotations)
}

func (this *BallerinaParser) addAnnotNotAttachedDiagnostic(annotList internal.STNodeList) internal.STNode {
	annotations := this.SyntaxErrors.updateAllNodesInNodeListWithDiagnostic(annotList,
		DiagnosticErrorCode.ERROR_ANNOTATION_NOT_ATTACHED_TO_A_CONSTRUCT)
	return annotations
}

func (this *BallerinaParser) isModuleVarDeclStart(lookahead int) bool {
	nextToken := this.peekN(lookahead + 1)
	switch nextToken.Kind() {
	case common.EQUAL_TOKEN, // Scenario: foo = . Even though this is not valid, consider this as a var-decl and
		// continue;
		common.OPEN_BRACKET_TOKEN,  // Scenario foo[] (Array type descriptor with custom type)
		common.QUESTION_MARK_TOKEN, // Scenario foo? (Optional type descriptor with custom type)
		common.PIPE_TOKEN,          // Scenario foo | (Union type descriptor with custom type)
		common.BITWISE_AND_TOKEN,   // Scenario foo & (Intersection type descriptor with custom type)
		common.OPEN_BRACE_TOKEN,    // Scenario foo{} (mapping-binding-pattern)
		common.ERROR_KEYWORD,       // Scenario foo error (error-binding-pattern)
		common.EOF_TOKEN:
		return true
	case common.IDENTIFIER_TOKEN:
		switch this.peekN(lookahead + 2).Kind() {
		case common.EQUAL_TOKEN,
			// Scenario: foo bar =
			common.SEMICOLON_TOKEN,
			// Scenario: foo bar;
			common.EOF_TOKEN:
			return true
		default:
			return false
		}
	case common.COLON_TOKEN:
		if lookahead > 1 {
			return false
		}
		switch this.peekN(lookahead + 2).Kind() {
		case common.IDENTIFIER_TOKEN:
			this.isModuleVarDeclStart(lookahead + 2)
		case common.EOF_TOKEN:
			return true
		default:
			return false
		}
	default:
		return false
	}
}

func (this *BallerinaParser) parseImportDecl() internal.STNode {
	this.startContext(common.PARSER_RULE_CONTEXT_IMPORT_DECL)
	this.tokenReader.StartMode(PARSER_MODE_IMPORT_MODE)
	importKeyword := this.parseImportKeyword()
	identifier := this.parseIdentifier(common.PARSER_RULE_CONTEXT_IMPORT_ORG_OR_MODULE_NAME)
	importDecl := this.parseImportDecl(importKeyword, identifier)
	this.tokenReader.EndMode()
	this.endContext()
	return importDecl
}

func (this *BallerinaParser) parseImportKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.IMPORT_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.IMPORT_KEYWORD)
		return this.parseImportKeyword()
	}
}

func (this *BallerinaParser) parseIdentifier(currentCtx ParserRuleContext) internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.IDENTIFIER_TOKEN {
		return this.consume()
	} else if token.kind == SyntaxKind.MAP_KEYWORD {
		mapKeyword := this.consume()
		return this.STNodeFactory.createIdentifierToken(mapKeyword.text(), mapKeyword.leadingMinutiae(),
			mapKeyword.trailingMinutiae(), mapKeyword.diagnostics())
	}
}

func (this *BallerinaParser) parseImportDecl(importKeyword internal.STNode, identifier internal.STNode) internal.STNode {
	nextToken := this.peek()
	var orgName internal.STNode
	var moduleName internal.STNode
	var alias internal.STNode
	switch nextToken.kind {
	case SLASH_TOKEN:
		slash := this.parseSlashToken()
		orgName = this.STNodeFactory.createImportOrgNameNode(identifier, slash)
		moduleName = this.parseModuleName()
		alias = this.parseImportPrefixDecl()
		break
	case DOT_TOKEN:
	case AS_KEYWORD:
		orgName = this.STNodeFactory.createEmptyNode()
		moduleName = this.parseModuleName(identifier)
		alias = this.parseImportPrefixDecl()
		break
	case SEMICOLON_TOKEN:
		orgName = this.STNodeFactory.createEmptyNode()
		moduleName = this.parseModuleName(identifier)
		alias = this.STNodeFactory.createEmptyNode()
		break
	default:
		this.recover(peek(), ParserRuleContext.IMPORT_DECL_ORG_OR_MODULE_NAME_RHS)
		return this.parseImportDecl(importKeyword, identifier)
	}
	semicolon := this.parseSemicolon()
	return this.STNodeFactory.createImportDeclarationNode(importKeyword, orgName, moduleName, alias, semicolon)
}

func (this *BallerinaParser) parseSlashToken() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.SLASH_TOKEN {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.SLASH)
		return this.parseSlashToken()
	}
}

func (this *BallerinaParser) parseDotToken() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.DOT_TOKEN {
		return this.consume()
	} else {
		this.recover(peek(), ParserRuleContext.DOT)
		return this.parseDotToken()
	}
}

func (this *BallerinaParser) parseModuleName() internal.STNode {
	moduleNameStart := this.parseIdentifier(ParserRuleContext.IMPORT_MODULE_NAME)
	return this.parseModuleName(moduleNameStart)
}

func (this *BallerinaParser) parseModuleName(moduleNameStart internal.STNode) internal.STNode {
	moduleNameParts := make([]interface{}, 0)
	this.moduleNameParts.add(moduleNameStart)
	nextToken := this.peek()
	for !this.isEndOfImportDecl(nextToken) {
		moduleNameSeparator := this.parseModuleNameRhs()
		if moduleNameSeparator == nil {
			break
		}
		this.moduleNameParts.add(moduleNameSeparator)
		this.moduleNameParts.add(parseIdentifier(ParserRuleContext.IMPORT_MODULE_NAME))
		nextToken = this.peek()
	}
	return this.STNodeFactory.createNodeList(moduleNameParts)
}

func (this *BallerinaParser) parseModuleNameRhs() internal.STNode {
	switch peek().kind {
	case DOT_TOKEN:
		return this.consume()
	case AS_KEYWORD, SEMICOLON_TOKEN:
		return nil
	default:
		this.recover(peek(), ParserRuleContext.AFTER_IMPORT_MODULE_NAME)
		return this.parseModuleNameRhs()
	}
}

func (this *BallerinaParser) isEndOfImportDecl(nextToken internal.STToken) bool {
	switch nextToken.kind {
	case SEMICOLON_TOKEN,
		PUBLIC_KEYWORD,
		FUNCTION_KEYWORD,
		TYPE_KEYWORD,
		ABSTRACT_KEYWORD,
		CONST_KEYWORD,
		EOF_TOKEN,
		SERVICE_KEYWORD,
		IMPORT_KEYWORD,
		FINAL_KEYWORD,
		TRANSACTIONAL_KEYWORD,
		ISOLATED_KEYWORD:
		return true
	default:
		return false
	}
}

func (this *BallerinaParser) parseDecimalIntLiteral(context ParserRuleContext) internal.STNode {
	nextToken := this.peek()
	if nextToken.kind == SyntaxKind.DECIMAL_INTEGER_LITERAL_TOKEN {
		return this.consume()
	} else {
		this.recover(peek(), context)
		return this.parseDecimalIntLiteral(context)
	}
}

func (this *BallerinaParser) parseImportPrefixDecl() internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case AS_KEYWORD:
		asKeyword := this.parseAsKeyword()
		prefix := this.parseImportPrefix()
		return this.STNodeFactory.createImportPrefixNode(asKeyword, prefix)
	case SEMICOLON_TOKEN:
		return this.STNodeFactory.createEmptyNode()
	default:
		if this.isEndOfImportDecl(nextToken) {
			return this.STNodeFactory.createEmptyNode()
		}
		this.recover(peek(), ParserRuleContext.IMPORT_PREFIX_DECL)
		return this.parseImportPrefixDecl()
	}
}

func (this *BallerinaParser) parseAsKeyword() internal.STNode {
	nextToken := this.peek()
	if nextToken.kind == SyntaxKind.AS_KEYWORD {
		return this.consume()
	} else {
		this.recover(peek(), ParserRuleContext.AS_KEYWORD)
		return this.parseAsKeyword()
	}
}

func (this *BallerinaParser) parseImportPrefix() internal.STNode {
	nextToken := this.peek()
	if nextToken.kind == SyntaxKind.IDENTIFIER_TOKEN {
		identifier := this.consume()
		if this.isUnderscoreToken(identifier) {
			return this.getUnderscoreKeyword(identifier)
		}
		return identifier
	} else if this.isPredeclaredPrefix(nextToken.kind) {
		preDeclaredPrefix := this.consume()
		return this.STNodeFactory.createIdentifierToken(preDeclaredPrefix.text(), preDeclaredPrefix.leadingMinutiae(),
			preDeclaredPrefix.trailingMinutiae())
	}
}

func (this *BallerinaParser) parseTopLevelNode(metadata internal.STNode, publicQualifier internal.STNode) internal.STNode {
	topLevelQualifiers := make([]interface{}, 0)
	return this.parseTopLevelNode(metadata, publicQualifier, topLevelQualifiers)
}

func (this *BallerinaParser) parseTopLevelNode(metadata internal.STNode, publicQualifier internal.STNode, qualifiers []STNode) internal.STNode {
	this.parseTopLevelQualifiers(qualifiers)
	nextToken := this.peek()
	switch nextToken.kind {
	case EOF_TOKEN:
		return this.createMissingSimpleVarDecl(metadata, publicQualifier, qualifiers, true)
	case FUNCTION_KEYWORD:
		return this.parseFuncDefOrFuncTypeDesc(metadata, publicQualifier, qualifiers, false, false)
	case TYPE_KEYWORD:
		this.reportInvalidQualifierList(qualifiers)
		return this.parseModuleTypeDefinition(metadata, publicQualifier)
	case CLASS_KEYWORD:
		return this.parseClassDefinition(metadata, publicQualifier, qualifiers)
	case LISTENER_KEYWORD:
		this.reportInvalidQualifierList(qualifiers)
		return this.parseListenerDeclaration(metadata, publicQualifier)
	case CONST_KEYWORD:
		this.reportInvalidQualifierList(qualifiers)
		return this.parseConstantDeclaration(metadata, publicQualifier)
	case ANNOTATION_KEYWORD:
		this.reportInvalidQualifierList(qualifiers)
		constKeyword := this.STNodeFactory.createEmptyNode()
		return this.parseAnnotationDeclaration(metadata, publicQualifier, constKeyword)
	case IMPORT_KEYWORD:
		this.reportInvalidMetaData(metadata, "import declaration")
		this.reportInvalidQualifier(publicQualifier)
		this.reportInvalidQualifierList(qualifiers)
		return this.parseImportDecl()
	case XMLNS_KEYWORD:
		this.reportInvalidMetaData(metadata, "XML namespace declaration")
		this.reportInvalidQualifier(publicQualifier)
		this.reportInvalidQualifierList(qualifiers)
		return this.parseXMLNamespaceDeclaration(true)
	case ENUM_KEYWORD:
		this.reportInvalidQualifierList(qualifiers)
		return this.parseEnumDeclaration(metadata, publicQualifier)
	case RESOURCE_KEYWORD:
	case REMOTE_KEYWORD:
		this.reportInvalidQualifier(consume())
		return this.parseTopLevelNode(metadata, publicQualifier, qualifiers)
	case IDENTIFIER_TOKEN:
		if this.isModuleVarDeclStart(1) {
			return this.parseModuleVarDecl(metadata, publicQualifier, qualifiers)
		}
	default:
		if this.isPossibleServiceDecl(qualifiers) {
			return this.parseServiceDeclOrVarDecl(metadata, publicQualifier, qualifiers)
		}
		if this.isTypeStartingToken(nextToken.kind) && (nextToken.kind != SyntaxKind.IDENTIFIER_TOKEN) {
			return this.parseModuleVarDecl(metadata, publicQualifier, qualifiers)
		}
		token := this.peek()
		solution := this.recover(token, ParserRuleContext.TOP_LEVEL_NODE_WITHOUT_MODIFIER)
		if solution.action == Action.KEEP {
			return this.parseModuleVarDecl(metadata, publicQualifier, qualifiers)
		}
		return this.parseTopLevelNode(metadata, publicQualifier, qualifiers)
	}
}

func (this *BallerinaParser) parseModuleVarDecl(metadata internal.STNode) internal.STNode {
	emptyList := make([]interface{}, 0)
	publicQualifier := this.STNodeFactory.createEmptyNode()
	return this.parseVariableDecl(metadata, publicQualifier, emptyList, emptyList, true)
}

func (this *BallerinaParser) parseModuleVarDecl(metadata internal.STNode, publicQualifier internal.STNode, topLevelQualifiers []STNode) internal.STNode {
	varDeclQuals := this.extractVarDeclQualifiers(topLevelQualifiers, true)
	return this.parseVariableDecl(metadata, publicQualifier, varDeclQuals, topLevelQualifiers, true)
}

func (this *BallerinaParser) extractVarDeclQualifiers(qualifiers []STNode, isModuleVar bool) []STNode {
	varDeclQualList := make([]interface{}, 0)
	initialListSize := len(qualifiers)
	configurableQualIndex := (-1)
	i := 0
	for ; (i < 2) && (i < initialListSize); i++ {
		qualifierKind := qualifiers.get(0).kind
		if (!this.isSyntaxKindInList(varDeclQualList, qualifierKind)) && this.isModuleVarDeclQualifier(qualifierKind) {
			this.varDeclQualList.add(qualifiers.remove(0))
			if qualifierKind == SyntaxKind.CONFIGURABLE_KEYWORD {
				configurableQualIndex = i
			}
			continue
		}
		break
	}
	if isModuleVar && (configurableQualIndex > (-1)) {
		configurableQual := this.varDeclQualList.get(configurableQualIndex)
		i := 0
		for ; i < len(varDeclQualList); i++ {
			if i < configurableQualIndex {
				invalidQual := this.varDeclQualList.get(i)
				configurableQual = this.SyntaxErrors.cloneWithLeadingInvalidNodeMinutiae(configurableQual, invalidQual,
					getInvalidQualifierError(invalidQual.kind), (invalidQual).text())
			} else if i > configurableQualIndex {
				invalidQual := this.varDeclQualList.get(i)
				configurableQual = this.SyntaxErrors.cloneWithTrailingInvalidNodeMinutiae(configurableQual, invalidQual,
					getInvalidQualifierError(invalidQual.kind), (invalidQual).text())
			}
		}
		varDeclQualList = make([]interface{}, 0)
	}
	return varDeclQualList
}

func (this *BallerinaParser) getInvalidQualifierError(qualifierKind SyntaxKind) DiagnosticErrorCode {
	if qualifierKind == SyntaxKind.FINAL_KEYWORD {
		return DiagnosticErrorCode.ERROR_CONFIGURABLE_VAR_IMPLICITLY_FINAL
	}
	return DiagnosticErrorCode.ERROR_QUALIFIER_NOT_ALLOWED
}

func (this *BallerinaParser) isModuleVarDeclQualifier(tokenKind SyntaxKind) bool {
	switch tokenKind {
	case FINAL_KEYWORD, ISOLATED_KEYWORD, CONFIGURABLE_KEYWORD:
		return true
	default:
		return false
	}
}

func (this *BallerinaParser) reportInvalidQualifier(qualifier internal.STNode) {
	if (qualifier != nil) && (qualifier.kind != SyntaxKind.NONE) {
		this.addInvalidNodeToNextToken(qualifier, DiagnosticErrorCode.ERROR_INVALID_QUALIFIER,
			(qualifier).text())
	}
}

func (this *BallerinaParser) reportInvalidMetaData(metadata internal.STNode, constructName String) {
	if (metadata != nil) && (metadata.kind != SyntaxKind.NONE) {
		this.addInvalidNodeToNextToken(metadata, DiagnosticErrorCode.ERROR_INVALID_METADATA, constructName)
	}
}

func (this *BallerinaParser) reportInvalidQualifierList(qualifiers []STNode) {
	for _, qual := range qualifiers {
		this.addInvalidNodeToNextToken(qual, DiagnosticErrorCode.ERROR_QUALIFIER_NOT_ALLOWED, (qual).text())
	}
}

func (this *BallerinaParser) reportInvalidStatementAnnots(annots internal.STNode, qualifiers []STNode) {
	diagnosticErrorCode := DiagnosticErrorCode.ERROR_ANNOTATIONS_ATTACHED_TO_STATEMENT
	this.reportInvalidAnnotations(annots, qualifiers, diagnosticErrorCode)
}

func (this *BallerinaParser) reportInvalidExpressionAnnots(annots internal.STNode, qualifiers []STNode) {
	diagnosticErrorCode := DiagnosticErrorCode.ERROR_ANNOTATIONS_ATTACHED_TO_EXPRESSION
	this.reportInvalidAnnotations(annots, qualifiers, diagnosticErrorCode)
}

func (this *BallerinaParser) reportInvalidAnnotations(annots internal.STNode, qualifiers []STNode, errorCode DiagnosticErrorCode) {
	if this.isNodeListEmpty(annots) {
		return
	}
	if this.qualifiers.isEmpty() {
		this.addInvalidNodeToNextToken(annots, errorCode)
	} else {
		this.updateFirstNodeInListWithLeadingInvalidNode(qualifiers, annots, errorCode)
	}
}

func (this *BallerinaParser) isTopLevelQualifier(tokenKind SyntaxKind) bool {
	var nextNextToken internal.STToken
	switch tokenKind {
	case FINAL_KEYWORD, // final-qualifier
		CONFIGURABLE_KEYWORD:
		return true
	case READONLY_KEYWORD:
		nextNextToken = this.getNextNextToken()
		switch nextNextToken.kind {
		case CLIENT_KEYWORD,
			SERVICE_KEYWORD,
			DISTINCT_KEYWORD,
			ISOLATED_KEYWORD,
			CLASS_KEYWORD:
			return true
		default:
			return false
		}
	case DISTINCT_KEYWORD:
		nextNextToken = this.getNextNextToken()
		switch nextNextToken.kind {
		case CLIENT_KEYWORD,
			SERVICE_KEYWORD,
			READONLY_KEYWORD,
			ISOLATED_KEYWORD,
			CLASS_KEYWORD:
			return true
		default:
			return false
		}
	default:
		return this.isTypeDescQualifier(tokenKind)
	}
}

func (this *BallerinaParser) isTypeDescQualifier(tokenKind SyntaxKind) bool {
	switch tokenKind {
	case TRANSACTIONAL_KEYWORD, // func-type-dec, func-def
		ISOLATED_KEYWORD, // func-type-dec, object-type-desc, func-def, class-def, isolated-final-qual
		CLIENT_KEYWORD,   // object-type-desc, class-def
		ABSTRACT_KEYWORD, // object-type-desc(outdated)
		SERVICE_KEYWORD:
		return true
	default:
		return false
	}
}

func (this *BallerinaParser) isObjectMemberQualifier(tokenKind SyntaxKind) bool {
	switch tokenKind {
	case REMOTE_KEYWORD, // method-def, method-decl
		RESOURCE_KEYWORD, // resource-method-def
		FINAL_KEYWORD:
		return true
	default:
		return this.isTypeDescQualifier(tokenKind)
	}
}

func (this *BallerinaParser) isExprQualifier(tokenKind SyntaxKind) bool {
	switch tokenKind {
	case TRANSACTIONAL_KEYWORD:
		nextNextToken := this.getNextNextToken()
		switch nextNextToken.kind {
		case CLIENT_KEYWORD,
			ABSTRACT_KEYWORD,
			ISOLATED_KEYWORD,
			OBJECT_KEYWORD,
			FUNCTION_KEYWORD:
			return true
		default:
			return false
		}
	default:
		return this.isTypeDescQualifier(tokenKind)
	}
}

func (this *BallerinaParser) parseTopLevelQualifiers(qualifiers []STNode) {
	for this.isTopLevelQualifier(peek().kind) {
		qualifier := this.consume()
		this.qualifiers.add(qualifier)
	}
}

func (this *BallerinaParser) parseTypeDescQualifiers(qualifiers []STNode) {
	for this.isTypeDescQualifier(peek().kind) {
		qualifier := this.consume()
		this.qualifiers.add(qualifier)
	}
}

func (this *BallerinaParser) parseObjectMemberQualifiers(qualifiers []STNode) {
	for this.isObjectMemberQualifier(peek().kind) {
		qualifier := this.consume()
		this.qualifiers.add(qualifier)
	}
}

func (this *BallerinaParser) parseExprQualifiers(qualifiers []STNode) {
	for this.isExprQualifier(peek().kind) {
		qualifier := this.consume()
		this.qualifiers.add(qualifier)
	}
}

func (this *BallerinaParser) parseOptionalRelativePath(isObjectMember bool) internal.STNode {
	var resourcePath internal.STNode
	nextToken := this.peek()
	switch nextToken.kind {
	case DOT_TOKEN:
	case IDENTIFIER_TOKEN:
	case OPEN_BRACKET_TOKEN:
		resourcePath = this.parseRelativeResourcePath()
		break
	case OPEN_PAREN_TOKEN:
		return this.STNodeFactory.createEmptyNodeList()
	default:
		this.recover(nextToken, ParserRuleContext.OPTIONAL_RELATIVE_PATH)
		return this.parseOptionalRelativePath(isObjectMember)
	}
	if !isObjectMember {
		this.addInvalidNodeToNextToken(resourcePath, DiagnosticErrorCode.ERROR_RESOURCE_PATH_IN_FUNCTION_DEFINITION)
		return this.STNodeFactory.createEmptyNodeList()
	}
	return resourcePath
}

func (this *BallerinaParser) parseFuncDefOrFuncTypeDesc(metadata internal.STNode, visibilityQualifier internal.STNode, qualifiers []STNode, isObjectMember bool, isObjectTypeDesc bool) internal.STNode {
	this.startContext(ParserRuleContext.FUNC_DEF_OR_FUNC_TYPE)
	functionKeyword := this.parseFunctionKeyword()
	funcDefOrType := this.parseFunctionKeywordRhs(metadata, visibilityQualifier, qualifiers, functionKeyword,
		isObjectMember, isObjectTypeDesc)
	return funcDefOrType
}

func (this *BallerinaParser) parseFunctionDefinition(metadata internal.STNode, visibilityQualifier internal.STNode, resourcePath internal.STNode, qualifiers []STNode, functionKeyword internal.STNode, name internal.STNode, isObjectMember bool, isObjectTypeDesc bool) internal.STNode {
	this.switchContext(ParserRuleContext.FUNC_DEF)
	funcSignature := this.parseFuncSignature(false)
	funcDef := this.parseFuncDefOrMethodDeclEnd(metadata, visibilityQualifier, qualifiers, functionKeyword, name,
		resourcePath, funcSignature, isObjectMember, isObjectTypeDesc)
	this.endContext()
	return funcDef
}

func (this *BallerinaParser) parseFuncDefOrFuncTypeDescRhs(metadata internal.STNode, visibilityQualifier internal.STNode, qualifiers []STNode, functionKeyword internal.STNode, name internal.STNode, isObjectMember bool, isObjectTypeDesc bool) internal.STNode {
	switch peek().kind {
	case OPEN_PAREN_TOKEN:
	case DOT_TOKEN:
	case IDENTIFIER_TOKEN:
	case OPEN_BRACKET_TOKEN:
		resourcePath := this.parseOptionalRelativePath(isObjectMember)
		return this.parseFunctionDefinition(metadata, visibilityQualifier, resourcePath, qualifiers, functionKeyword,
			name, isObjectMember, isObjectTypeDesc)
	case EQUAL_TOKEN:
	case SEMICOLON_TOKEN:
		this.endContext()
		extractQualifiersList := this.extractVarDeclOrObjectFieldQualifiers(qualifiers, isObjectMember,
			isObjectTypeDesc)
		typeDesc := this.createFunctionTypeDescriptor(qualifiers, functionKeyword,
			STNodeFactory.createEmptyNode(), false)
		if isObjectMember {
			objectFieldQualNodeList := this.STNodeFactory.createNodeList(extractQualifiersList)
			return this.parseObjectFieldRhs(metadata, visibilityQualifier, objectFieldQualNodeList, typeDesc, name,
				isObjectTypeDesc)
		}
		this.startContext(ParserRuleContext.VAR_DECL_STMT)
		funcTypeName := this.STNodeFactory.createSimpleNameReferenceNode(name)
		bindingPattern := this.createCaptureOrWildcardBP((funcTypeName).name)
		typedBindingPattern := this.STNodeFactory.createTypedBindingPatternNode(typeDesc, bindingPattern)
		return this.parseVarDeclRhs(metadata, visibilityQualifier, extractQualifiersList, typedBindingPattern, true)
	default:
		token := this.peek()
		this.recover(token, ParserRuleContext.FUNC_DEF_OR_TYPE_DESC_RHS)
		return this.parseFuncDefOrFuncTypeDescRhs(metadata, visibilityQualifier, qualifiers, functionKeyword, name,
			isObjectMember, isObjectTypeDesc)
	}
}

func (this *BallerinaParser) parseFunctionKeywordRhs(metadata internal.STNode, visibilityQualifier internal.STNode, qualifiers []STNode, functionKeyword internal.STNode, isObjectMember bool, isObjectTypeDesc bool) internal.STNode {
	switch peek().kind {
	case IDENTIFIER_TOKEN:
		name := this.consume()
		return this.parseFuncDefOrFuncTypeDescRhs(metadata, visibilityQualifier, qualifiers, functionKeyword, name,
			isObjectMember, isObjectTypeDesc)
	case OPEN_PAREN_TOKEN:
		this.switchContext(ParserRuleContext.VAR_DECL_STMT)
		this.startContext(ParserRuleContext.TYPE_DESC_IN_TYPE_BINDING_PATTERN)
		this.startContext(ParserRuleContext.FUNC_TYPE_DESC)
		funcSignature := this.parseFuncSignature(true)
		this.endContext()
		this.endContext()
		return this.parseFunctionTypeDescRhs(metadata, visibilityQualifier, qualifiers, functionKeyword,
			funcSignature, isObjectMember, isObjectTypeDesc)
	default:
		token := this.peek()
		if this.isValidTypeContinuationToken(token) || this.isBindingPatternsStartToken(token.kind) {
			return this.parseVarDeclWithFunctionType(metadata, visibilityQualifier, qualifiers, functionKeyword,
				STNodeFactory.createEmptyNode(), isObjectMember,
				isObjectTypeDesc, false)
		}
		this.recover(token, ParserRuleContext.FUNCTION_KEYWORD_RHS)
		return this.parseFunctionKeywordRhs(metadata, visibilityQualifier, qualifiers, functionKeyword,
			isObjectMember, isObjectTypeDesc)
	}
}

func (this *BallerinaParser) isBindingPatternsStartToken(tokenKind SyntaxKind) bool {
	switch tokenKind {
	case IDENTIFIER_TOKEN,
		OPEN_BRACKET_TOKEN,
		OPEN_BRACE_TOKEN,
		ERROR_KEYWORD:
		return true
	default:
		return false
	}
}

func (this *BallerinaParser) parseFuncDefOrMethodDeclEnd(metadata internal.STNode, visibilityQualifier internal.STNode, qualifierList []STNode, functionKeyword internal.STNode, name internal.STNode, resourcePath internal.STNode, funcSignature internal.STNode, isObjectMember bool, isObjectTypeDesc bool) internal.STNode {
	if !isObjectMember {
		return this.createFunctionDefinition(metadata, visibilityQualifier, qualifierList, functionKeyword, name,
			funcSignature)
	}
	hasResourcePath := (!this.isNodeListEmpty(resourcePath))
	hasResourceQual := this.isSyntaxKindInList(qualifierList, SyntaxKind.RESOURCE_KEYWORD)
	if hasResourceQual && (!hasResourcePath) {
		relativePath := make([]interface{}, 0)
		this.relativePath.add(STNodeFactory.createMissingToken(SyntaxKind.DOT_TOKEN))
		resourcePath = this.STNodeFactory.createNodeList(relativePath)
		var errorCode DiagnosticErrorCode
		if isObjectTypeDesc {
			errorCode = DiagnosticErrorCode.ERROR_MISSING_RESOURCE_PATH_IN_RESOURCE_ACCESSOR_DECLARATION
		} else {
			errorCode = DiagnosticErrorCode.ERROR_MISSING_RESOURCE_PATH_IN_RESOURCE_ACCESSOR_DEFINITION
		}
		name = this.SyntaxErrors.addDiagnostic(name, errorCode)
		hasResourcePath = true
	}
	if hasResourcePath {
		return this.createResourceAccessorDefnOrDecl(metadata, visibilityQualifier, qualifierList, functionKeyword, name,
			resourcePath, funcSignature, isObjectTypeDesc)
	}
	if isObjectTypeDesc {
		return this.createMethodDeclaration(metadata, visibilityQualifier, qualifierList, functionKeyword, name,
			funcSignature)
	} else {
		return this.createMethodDefinition(metadata, visibilityQualifier, qualifierList, functionKeyword, name,
			funcSignature)
	}
}

func (this *BallerinaParser) createFunctionDefinition(metadata internal.STNode, visibilityQualifier internal.STNode, qualifierList []STNode, functionKeyword internal.STNode, name internal.STNode, funcSignature internal.STNode) internal.STNode {
	validatedList := make([]interface{}, 0)
	i := 0
	for ; i < len(qualifierList); i++ {
		qualifier := this.qualifierList.get(i)
		nextIndex := (i + 1)
		if this.isSyntaxKindInList(validatedList, qualifier.kind) {
			this.updateLastNodeInListWithInvalidNode(validatedList, qualifier,
				DiagnosticErrorCode.ERROR_DUPLICATE_QUALIFIER, (qualifier).text())
			continue
		}
		if this.isRegularFuncQual(qualifier.kind) {
			this.validatedList.add(qualifier)
			continue
		}
		if len(qualifierList) == nextIndex {
			functionKeyword = this.SyntaxErrors.cloneWithLeadingInvalidNodeMinutiae(functionKeyword, qualifier,
				DiagnosticErrorCode.ERROR_QUALIFIER_NOT_ALLOWED, (qualifier).text())
		} else {
			this.updateANodeInListWithLeadingInvalidNode(qualifierList, nextIndex, qualifier,
				DiagnosticErrorCode.ERROR_QUALIFIER_NOT_ALLOWED, (qualifier).text())
		}
	}
	if visibilityQualifier != nil {
		this.validatedList.add(0, visibilityQualifier)
	}
	qualifiers := this.STNodeFactory.createNodeList(validatedList)
	resourcePath := this.STNodeFactory.createEmptyNodeList()
	body := this.parseFunctionBody()
	return this.STNodeFactory.createFunctionDefinitionNode(SyntaxKind.FUNCTION_DEFINITION, metadata, qualifiers,
		functionKeyword, name, resourcePath, funcSignature, body)
}

func (this *BallerinaParser) createMethodDefinition(metadata internal.STNode, visibilityQualifier internal.STNode, qualifierList []STNode, functionKeyword internal.STNode, name internal.STNode, funcSignature internal.STNode) internal.STNode {
	validatedList := make([]interface{}, 0)
	hasRemoteQual := false
	i := 0
	for ; i < len(qualifierList); i++ {
		qualifier := this.qualifierList.get(i)
		nextIndex := (i + 1)
		if this.isSyntaxKindInList(validatedList, qualifier.kind) {
			this.updateLastNodeInListWithInvalidNode(validatedList, qualifier,
				DiagnosticErrorCode.ERROR_DUPLICATE_QUALIFIER, (qualifier).text())
			continue
		}
		if qualifier.kind == SyntaxKind.REMOTE_KEYWORD {
			hasRemoteQual = true
			this.validatedList.add(qualifier)
			continue
		}
		if this.isRegularFuncQual(qualifier.kind) {
			this.validatedList.add(qualifier)
			continue
		}
		if len(qualifierList) == nextIndex {
			functionKeyword = this.SyntaxErrors.cloneWithLeadingInvalidNodeMinutiae(functionKeyword, qualifier,
				DiagnosticErrorCode.ERROR_QUALIFIER_NOT_ALLOWED, (qualifier).text())
		} else {
			this.updateANodeInListWithLeadingInvalidNode(qualifierList, nextIndex, qualifier,
				DiagnosticErrorCode.ERROR_QUALIFIER_NOT_ALLOWED, (qualifier).text())
		}
	}
	if visibilityQualifier != nil {
		if hasRemoteQual {
			this.updateFirstNodeInListWithLeadingInvalidNode(validatedList, visibilityQualifier,
				DiagnosticErrorCode.ERROR_REMOTE_METHOD_HAS_A_VISIBILITY_QUALIFIER)
		} else {
			this.validatedList.add(0, visibilityQualifier)
		}
	}
	qualifiers := this.STNodeFactory.createNodeList(validatedList)
	resourcePath := this.STNodeFactory.createEmptyNodeList()
	body := this.parseFunctionBody()
	return this.STNodeFactory.createFunctionDefinitionNode(SyntaxKind.OBJECT_METHOD_DEFINITION, metadata, qualifiers,
		functionKeyword, name, resourcePath, funcSignature, body)
}

func (this *BallerinaParser) createMethodDeclaration(metadata internal.STNode, visibilityQualifier internal.STNode, qualifierList []STNode, functionKeyword internal.STNode, name internal.STNode, funcSignature internal.STNode) internal.STNode {
	validatedList := make([]interface{}, 0)
	hasRemoteQual := false
	i := 0
	for ; i < len(qualifierList); i++ {
		qualifier := this.qualifierList.get(i)
		nextIndex := (i + 1)
		if this.isSyntaxKindInList(validatedList, qualifier.kind) {
			this.updateLastNodeInListWithInvalidNode(validatedList, qualifier,
				DiagnosticErrorCode.ERROR_DUPLICATE_QUALIFIER, (qualifier).text())
			continue
		}
		if qualifier.kind == SyntaxKind.REMOTE_KEYWORD {
			hasRemoteQual = true
			this.validatedList.add(qualifier)
			continue
		}
		if this.isRegularFuncQual(qualifier.kind) {
			this.validatedList.add(qualifier)
			continue
		}
		if len(qualifierList) == nextIndex {
			functionKeyword = this.SyntaxErrors.cloneWithLeadingInvalidNodeMinutiae(functionKeyword, qualifier,
				DiagnosticErrorCode.ERROR_QUALIFIER_NOT_ALLOWED, (qualifier).text())
		} else {
			this.updateANodeInListWithLeadingInvalidNode(qualifierList, nextIndex, qualifier,
				DiagnosticErrorCode.ERROR_QUALIFIER_NOT_ALLOWED, (qualifier).text())
		}
	}
	if visibilityQualifier != nil {
		if hasRemoteQual {
			this.updateFirstNodeInListWithLeadingInvalidNode(validatedList, visibilityQualifier,
				DiagnosticErrorCode.ERROR_REMOTE_METHOD_HAS_A_VISIBILITY_QUALIFIER)
		} else {
			this.validatedList.add(0, visibilityQualifier)
		}
	}
	qualifiers := this.STNodeFactory.createNodeList(validatedList)
	resourcePath := this.STNodeFactory.createEmptyNodeList()
	semicolon := this.parseSemicolon()
	return this.STNodeFactory.createMethodDeclarationNode(SyntaxKind.METHOD_DECLARATION, metadata, qualifiers,
		functionKeyword, name, resourcePath, funcSignature, semicolon)
}

func (this *BallerinaParser) createResourceAccessorDefnOrDecl(metadata internal.STNode, visibilityQualifier internal.STNode, qualifierList []STNode, functionKeyword internal.STNode, name internal.STNode, resourcePath internal.STNode, funcSignature internal.STNode, isObjectTypeDesc bool) internal.STNode {
	validatedList := make([]interface{}, 0)
	hasResourceQual := false
	i := 0
	for ; i < len(qualifierList); i++ {
		qualifier := this.qualifierList.get(i)
		nextIndex := (i + 1)
		if this.isSyntaxKindInList(validatedList, qualifier.kind) {
			this.updateLastNodeInListWithInvalidNode(validatedList, qualifier,
				DiagnosticErrorCode.ERROR_DUPLICATE_QUALIFIER, (qualifier).text())
			continue
		}
		if qualifier.kind == SyntaxKind.RESOURCE_KEYWORD {
			hasResourceQual = true
			this.validatedList.add(qualifier)
			continue
		}
		if this.isRegularFuncQual(qualifier.kind) {
			this.validatedList.add(qualifier)
			continue
		}
		if len(qualifierList) == nextIndex {
			functionKeyword = this.SyntaxErrors.cloneWithLeadingInvalidNodeMinutiae(functionKeyword, qualifier,
				DiagnosticErrorCode.ERROR_QUALIFIER_NOT_ALLOWED, (qualifier).text())
		} else {
			this.updateANodeInListWithLeadingInvalidNode(qualifierList, nextIndex, qualifier,
				DiagnosticErrorCode.ERROR_QUALIFIER_NOT_ALLOWED, (qualifier).text())
		}
	}
	if !hasResourceQual {
		this.validatedList.add(STNodeFactory.createMissingToken(SyntaxKind.RESOURCE_KEYWORD))
		functionKeyword = this.SyntaxErrors.addDiagnostic(functionKeyword, DiagnosticErrorCode.ERROR_MISSING_RESOURCE_KEYWORD)
	}
	if visibilityQualifier != nil {
		this.updateFirstNodeInListWithLeadingInvalidNode(validatedList, visibilityQualifier,
			DiagnosticErrorCode.ERROR_QUALIFIER_NOT_ALLOWED, (visibilityQualifier).text())
	}
	qualifiers := this.STNodeFactory.createNodeList(validatedList)
	if isObjectTypeDesc {
		semicolon := this.parseSemicolon()
		return this.STNodeFactory.createMethodDeclarationNode(SyntaxKind.RESOURCE_ACCESSOR_DECLARATION, metadata,
			qualifiers, functionKeyword, name, resourcePath, funcSignature, semicolon)
	} else {
		body := this.parseFunctionBody()
		return this.STNodeFactory.createFunctionDefinitionNode(SyntaxKind.RESOURCE_ACCESSOR_DEFINITION, metadata,
			qualifiers, functionKeyword, name, resourcePath, funcSignature, body)
	}
}

func (this *BallerinaParser) parseFuncSignature(isParamNameOptional bool) internal.STNode {
	openParenthesis := this.parseOpenParenthesis()
	parameters := this.parseParamList(isParamNameOptional)
	closeParenthesis := this.parseCloseParenthesis()
	this.endContext()
	returnTypeDesc := this.parseFuncReturnTypeDescriptor(isParamNameOptional)
	return this.STNodeFactory.createFunctionSignatureNode(openParenthesis, parameters, closeParenthesis, returnTypeDesc)
}

func (this *BallerinaParser) parseFunctionTypeDescRhs(metadata internal.STNode, visibilityQualifier internal.STNode, qualifiers []STNode, functionKeyword internal.STNode, funcSignature internal.STNode, isObjectMember bool, isObjectTypeDesc bool) internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case OPEN_BRACE_TOKEN:
	case EQUAL_TOKEN:
		break
	case SEMICOLON_TOKEN:
	case IDENTIFIER_TOKEN:
	case OPEN_BRACKET_TOKEN:
	default:
		return this.parseVarDeclWithFunctionType(metadata, visibilityQualifier, qualifiers, functionKeyword,
			funcSignature, isObjectMember, isObjectTypeDesc, true)
	}
	this.switchContext(ParserRuleContext.FUNC_DEF)
	name := this.SyntaxErrors.createMissingTokenWithDiagnostics(SyntaxKind.IDENTIFIER_TOKEN,
		DiagnosticErrorCode.ERROR_MISSING_FUNCTION_NAME)
	funcSignature = this.validateAndGetFuncParams(funcSignature)
	resourcePath := this.STNodeFactory.createEmptyNodeList()
	funcDef := this.parseFuncDefOrMethodDeclEnd(metadata, visibilityQualifier, qualifiers, functionKeyword,
		name, resourcePath, funcSignature, isObjectMember, isObjectTypeDesc)
	this.endContext()
	return funcDef
}

func (this *BallerinaParser) extractVarDeclOrObjectFieldQualifiers(qualifierList []STNode, isObjectMember bool, isObjectTypeDesc bool) []STNode {
	if isObjectMember {
		return this.extractObjectFieldQualifiers(qualifierList, isObjectTypeDesc)
	}
	return this.extractVarDeclQualifiers(qualifierList, false)
}

func (this *BallerinaParser) createFunctionTypeDescriptor(qualifierList []STNode, functionKeyword internal.STNode, funcSignature internal.STNode, hasFuncSignature bool) internal.STNode {
	nodes := this.createFuncTypeQualNodeList(qualifierList, functionKeyword, hasFuncSignature)
	qualifierNodeList := nodes[0]
	functionKeyword = nodes[1]
	return this.STNodeFactory.createFunctionTypeDescriptorNode(qualifierNodeList, functionKeyword, funcSignature)
}

func (this *BallerinaParser) parseVarDeclWithFunctionType(metadata internal.STNode, visibilityQualifier internal.STNode, qualifierList []STNode, functionKeyword internal.STNode, funcSignature internal.STNode, isObjectMember bool, isObjectTypeDesc bool, hasFuncSignature bool) internal.STNode {
	this.switchContext(ParserRuleContext.VAR_DECL_STMT)
	extractQualifiersList := this.extractVarDeclOrObjectFieldQualifiers(qualifierList, isObjectMember,
		isObjectTypeDesc)
	typeDesc := this.createFunctionTypeDescriptor(qualifierList, functionKeyword, funcSignature, hasFuncSignature)
	typeDesc = this.parseComplexTypeDescriptor(typeDesc,
		ParserRuleContext.TYPE_DESC_IN_TYPE_BINDING_PATTERN, true)
	if isObjectMember {
		this.endContext()
		objectFieldQualNodeList := this.STNodeFactory.createNodeList(extractQualifiersList)
		fieldName := this.parseVariableName()
		return this.parseObjectFieldRhs(metadata, visibilityQualifier, objectFieldQualNodeList, typeDesc, fieldName,
			isObjectTypeDesc)
	}
	typedBindingPattern := this.parseTypedBindingPatternTypeRhs(typeDesc, ParserRuleContext.VAR_DECL_STMT)
	return this.parseVarDeclRhs(metadata, visibilityQualifier, extractQualifiersList, typedBindingPattern, true)
}

func (this *BallerinaParser) validateAndGetFuncParams(signature internal.STFunctionSignatureNode) internal.STNode {
	parameters := signature.parameters
	paramCount := this.parameters.bucketCount()
	index := 0
	for ; index < paramCount; index++ {
		param := this.parameters.childInBucket(index)
		switch param.kind {
		case REQUIRED_PARAM:
			requiredParam := internal.STRequiredParameterNode(param)
			if this.isEmpty(requiredParam.paramName) {
				break
			}
			continue
		case DEFAULTABLE_PARAM:
			defaultableParam := internal.STDefaultableParameterNode(param)
			if this.isEmpty(defaultableParam.paramName) {
				break
			}
			continue
		case REST_PARAM:
			restParam := internal.STRestParameterNode(param)
			if this.isEmpty(restParam.paramName) {
				break
			}
			continue
		default:
			continue
		}
		break
	}
	if index == paramCount {
		return signature
	}
	updatedParams := this.getUpdatedParamList(parameters, index)
	return this.STNodeFactory.createFunctionSignatureNode(signature.openParenToken, updatedParams,
		signature.closeParenToken, signature.returnTypeDesc)
}

func (this *BallerinaParser) getUpdatedParamList(parameters internal.STNode, index int) internal.STNode {
	paramCount := this.parameters.bucketCount()
	newIndex := 0
	newParams := make([]interface{}, 0)
	for ; newIndex < index; newIndex++ {
		this.newParams.add(parameters.childInBucket(index))
	}
	for ; newIndex < paramCount; newIndex++ {
		param := this.parameters.childInBucket(newIndex)
		paramName := this.STNodeFactory.createMissingToken(SyntaxKind.IDENTIFIER_TOKEN)
		switch param.kind {
		case REQUIRED_PARAM:
			requiredParam := internal.STRequiredParameterNode(param)
			if this.isEmpty(requiredParam.paramName) {
				param = this.STNodeFactory.createRequiredParameterNode(requiredParam.annotations,
					requiredParam.typeName, paramName)
			}
			break
		case DEFAULTABLE_PARAM:
			defaultableParam := internal.STDefaultableParameterNode(param)
			if this.isEmpty(defaultableParam.paramName) {
				param = this.STNodeFactory.createDefaultableParameterNode(defaultableParam.annotations, defaultableParam.typeName,
					paramName, defaultableParam.equalsToken, defaultableParam.expression)
			}
		case REST_PARAM:
			restParam := internal.STRestParameterNode(param)
			if this.isEmpty(restParam.paramName) {
				param = this.STNodeFactory.createRestParameterNode(restParam.annotations, restParam.typeName,
					restParam.ellipsisToken, paramName)
			}
		default:
		}
		this.newParams.add(param)
	}
	return this.STNodeFactory.createNodeList(newParams)
}

func (this *BallerinaParser) isEmpty(node internal.STNode) bool {
	return (!this.SyntaxUtils.isSTNodePresent(node))
}

func (this *BallerinaParser) parseFunctionKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.FUNCTION_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.FUNCTION_KEYWORD)
		return this.parseFunctionKeyword()
	}
}

func (this *BallerinaParser) parseFunctionName() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.IDENTIFIER_TOKEN {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.FUNC_NAME)
		return this.parseFunctionName()
	}
}

func (this *BallerinaParser) parseArgListOpenParenthesis() internal.STNode {
	return this.parseOpenParenthesis(ParserRuleContext.ARG_LIST_OPEN_PAREN)
}

func (this *BallerinaParser) parseOpenParenthesis() internal.STNode {
	return this.parseOpenParenthesis(ParserRuleContext.OPEN_PARENTHESIS)
}

func (this *BallerinaParser) parseOpenParenthesis(ctx ParserRuleContext) internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.OPEN_PAREN_TOKEN {
		return this.consume()
	} else {
		this.recover(token, ctx)
		return this.parseOpenParenthesis(ctx)
	}
}

func (this *BallerinaParser) parseArgListCloseParenthesis() internal.STNode {
	return this.parseCloseParenthesis(ParserRuleContext.ARG_LIST_CLOSE_PAREN)
}

func (this *BallerinaParser) parseCloseParenthesis() internal.STNode {
	return this.parseCloseParenthesis(ParserRuleContext.CLOSE_PARENTHESIS)
}

func (this *BallerinaParser) parseCloseParenthesis(ctx ParserRuleContext) internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.CLOSE_PAREN_TOKEN {
		return this.consume()
	} else {
		this.recover(token, ctx)
		return this.parseCloseParenthesis(ctx)
	}
}

func (this *BallerinaParser) parseParamList(isParamNameOptional bool) internal.STNode {
	this.startContext(ParserRuleContext.PARAM_LIST)
	token := this.peek()
	if this.isEndOfParametersList(token.kind) {
		return this.STNodeFactory.createEmptyNodeList()
	}
	paramsList := make([]interface{}, 0)
	this.startContext(ParserRuleContext.REQUIRED_PARAM)
	firstParam := this.parseParameter(SyntaxKind.REQUIRED_PARAM, isParamNameOptional)
	prevParamKind := firstParam.kind
	this.paramsList.add(firstParam)
	paramOrderErrorPresent := false
	token = this.peek()
	for !this.isEndOfParametersList(token.kind) {
		paramEnd := this.parseParameterRhs()
		if paramEnd == nil {
			break
		}
		this.endContext()
		if prevParamKind == SyntaxKind.DEFAULTABLE_PARAM {
			this.startContext(ParserRuleContext.DEFAULTABLE_PARAM)
		} else {
			this.startContext(ParserRuleContext.REQUIRED_PARAM)
		}
		param := this.parseParameter(prevParamKind, isParamNameOptional)
		if paramOrderErrorPresent {
			this.updateLastNodeInListWithInvalidNode(paramsList, paramEnd, null)
			this.updateLastNodeInListWithInvalidNode(paramsList, param, null)
		} else {
			paramOrderError := this.validateParamOrder(param, prevParamKind)
			if paramOrderError == nil {
				this.paramsList.add(paramEnd)
				this.paramsList.add(param)
			} else {
				paramOrderErrorPresent = true
				this.updateLastNodeInListWithInvalidNode(paramsList, paramEnd, null)
				this.updateLastNodeInListWithInvalidNode(paramsList, param, paramOrderError)
			}
		}
		prevParamKind = param.kind
		token = this.peek()
	}
	this.endContext()
	return this.STNodeFactory.createNodeList(paramsList)
}

func (this *BallerinaParser) validateParamOrder(param internal.STNode, prevParamKind SyntaxKind) diagnostics.DiagnosticCode {
	if prevParamKind == SyntaxKind.REST_PARAM {
		return DiagnosticErrorCode.ERROR_PARAMETER_AFTER_THE_REST_PARAMETER
	} else if (prevParamKind == SyntaxKind.DEFAULTABLE_PARAM) && (param.kind == SyntaxKind.REQUIRED_PARAM) {
		return DiagnosticErrorCode.ERROR_REQUIRED_PARAMETER_AFTER_THE_DEFAULTABLE_PARAMETER
	}
}

func (this *BallerinaParser) isSyntaxKindInList(nodeList []STNode, kind SyntaxKind) bool {
	for _, node := range nodeList {
		if node.kind == kind {
			return true
		}
	}
	return false
}

func (this *BallerinaParser) isPossibleServiceDecl(nodeList []STNode) bool {
	if this.nodeList.isEmpty() {
		return false
	}
	firstElement := this.nodeList.get(0)
	switch firstElement.kind {
	case SERVICE_KEYWORD:
		return true
	case ISOLATED_KEYWORD:
		return ((len(nodeList) > 1) && (nodeList.get(1).kind == SyntaxKind.SERVICE_KEYWORD))
	default:
		return false
	}
}

func (this *BallerinaParser) parseParameterRhs() internal.STNode {
	return this.parseParameterRhs(peek().kind)
}

func (this *BallerinaParser) parseParameterRhs(tokenKind SyntaxKind) internal.STNode {
	switch tokenKind {
	case COMMA_TOKEN:
		return this.consume()
	case CLOSE_PAREN_TOKEN:
		return nil
	default:
		this.recover(peek(), ParserRuleContext.PARAM_END)
		return this.parseParameterRhs()
	}
}

func (this *BallerinaParser) parseParameter(annots internal.STNode, prevParamKind SyntaxKind, isParamNameOptional bool) internal.STNode {
	var inclusionSymbol internal.STNode
	nextToken := this.peek()
	switch nextToken.kind {
	case ASTERISK_TOKEN:
		inclusionSymbol = this.consume()
		break
	case IDENTIFIER_TOKEN:
		inclusionSymbol = this.STNodeFactory.createEmptyNode()
		break
	default:
		if this.isTypeStartingToken(nextToken.kind) {
			inclusionSymbol = this.STNodeFactory.createEmptyNode()
			break
		}
		token := this.peek()
		solution := this.recover(token, ParserRuleContext.PARAMETER_START_WITHOUT_ANNOTATION)
		if solution.action == Action.KEEP {
			inclusionSymbol = this.STNodeFactory.createEmptyNodeList()
			break
		}
		return this.parseParameter(annots, prevParamKind, isParamNameOptional)
	}
	ty := this.parseTypeDescriptor(ParserRuleContext.TYPE_DESC_BEFORE_IDENTIFIER)
	return this.parseAfterParamType(prevParamKind, annots, inclusionSymbol, ty, isParamNameOptional)
}

func (this *BallerinaParser) parseParameter(prevParamKind SyntaxKind, isParamNameOptional bool) internal.STNode {
	var annots internal.STNode
	nextToken := this.peek()
	switch nextToken.kind {
	case AT_TOKEN:
		annots = this.parseOptionalAnnotations()
		break
	case ASTERISK_TOKEN:
	case IDENTIFIER_TOKEN:
		annots = this.STNodeFactory.createEmptyNodeList()
		break
	default:
		if this.isTypeStartingToken(nextToken.kind) {
			annots = this.STNodeFactory.createEmptyNodeList()
			break
		}
		token := this.peek()
		solution := this.recover(token, ParserRuleContext.PARAMETER_START)
		if solution.action == Action.KEEP {
			annots = this.STNodeFactory.createEmptyNodeList()
			break
		}
		return this.parseParameter(prevParamKind, isParamNameOptional)
	}
	return this.parseParameter(annots, prevParamKind, isParamNameOptional)
}

func (this *BallerinaParser) parseAfterParamType(prevParamKind SyntaxKind, annots internal.STNode, inclusionSymbol internal.STNode, ty internal.STNode, isParamNameOptional bool) internal.STNode {
	var paramName internal.STNode
	token := this.peek()
	switch token.kind {
	case ELLIPSIS_TOKEN:
		if inclusionSymbol != nil {
			ty = this.SyntaxErrors.cloneWithLeadingInvalidNodeMinutiae(ty, inclusionSymbol,
				DiagnosticErrorCode.REST_PARAMETER_CANNOT_BE_INCLUDED_RECORD_PARAMETER)
		}
		this.switchContext(ParserRuleContext.REST_PARAM)
		ellipsis := this.parseEllipsis()
		if isParamNameOptional && (peek().kind != SyntaxKind.IDENTIFIER_TOKEN) {
			paramName = this.STNodeFactory.createEmptyNode()
		} else {
			paramName = this.parseVariableName()
		}
		return this.STNodeFactory.createRestParameterNode(annots, ty, ellipsis, paramName)
	case IDENTIFIER_TOKEN:
		paramName = this.parseVariableName()
		return this.parseParameterRhs(prevParamKind, annots, inclusionSymbol, ty, paramName)
	case EQUAL_TOKEN:
		if !isParamNameOptional {
			break
		}
		paramName = this.STNodeFactory.createEmptyNode()
		return this.parseParameterRhs(prevParamKind, annots, inclusionSymbol, ty, paramName)
	default:
		if !isParamNameOptional {
			break
		}
		paramName = this.STNodeFactory.createEmptyNode()
		return this.parseParameterRhs(prevParamKind, annots, inclusionSymbol, ty, paramName)
	}
	this.recover(token, ParserRuleContext.AFTER_PARAMETER_TYPE)
	return this.parseAfterParamType(prevParamKind, annots, inclusionSymbol, ty, false)
}

func (this *BallerinaParser) parseEllipsis() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.ELLIPSIS_TOKEN {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.ELLIPSIS)
		return this.parseEllipsis()
	}
}

func (this *BallerinaParser) parseParameterRhs(prevParamKind SyntaxKind, annots internal.STNode, inclusionSymbol internal.STNode, ty internal.STNode, paramName internal.STNode) internal.STNode {
	nextToken := this.peek()
	if this.isEndOfParameter(nextToken.kind) {
		if inclusionSymbol != nil {
			return this.STNodeFactory.createIncludedRecordParameterNode(annots, inclusionSymbol, ty, paramName)
		} else {
			return this.STNodeFactory.createRequiredParameterNode(annots, ty, paramName)
		}
	} else if nextToken.kind == SyntaxKind.EQUAL_TOKEN {
		if prevParamKind == SyntaxKind.REQUIRED_PARAM {
			this.switchContext(ParserRuleContext.DEFAULTABLE_PARAM)
		}
		equal := this.parseAssignOp()
		expr := this.parseInferredTypeDescDefaultOrExpression()
		if inclusionSymbol != nil {
			ty = this.SyntaxErrors.cloneWithLeadingInvalidNodeMinutiae(ty, inclusionSymbol,
				DiagnosticErrorCode.ERROR_DEFAULTABLE_PARAMETER_CANNOT_BE_INCLUDED_RECORD_PARAMETER)
		}
		return this.STNodeFactory.createDefaultableParameterNode(annots, ty, paramName, equal, expr)
	}
}

func (this *BallerinaParser) parseComma() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.COMMA_TOKEN {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.COMMA)
		return this.parseComma()
	}
}

func (this *BallerinaParser) parseFuncReturnTypeDescriptor(isFuncTypeDesc bool) internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case OPEN_BRACE_TOKEN:
	case EQUAL_TOKEN:
		return this.STNodeFactory.createEmptyNode()
	case RETURNS_KEYWORD:
		break
	case IDENTIFIER_TOKEN:
		if (!isFuncTypeDesc) || this.isSafeMissingReturnsParse() {
			break
		}
	default:
		nextNextToken := this.getNextNextToken()
		if nextNextToken.kind == SyntaxKind.RETURNS_KEYWORD {
			break
		}
		return this.STNodeFactory.createEmptyNode()
	}
	returnsKeyword := this.parseReturnsKeyword()
	annot := this.parseOptionalAnnotations()
	ty := this.parseTypeDescriptor(ParserRuleContext.TYPE_DESC_IN_RETURN_TYPE_DESC)
	return this.STNodeFactory.createReturnTypeDescriptorNode(returnsKeyword, annot, ty)
}

func (this *BallerinaParser) isSafeMissingReturnsParse() bool {
	for _, context := range this.this.errorHandler.getContextStack() {
		if !this.isSafeMissingReturnsParseCtx(context) {
			return false
		}
	}
	return true
}

func (this *BallerinaParser) isSafeMissingReturnsParseCtx(ctx ParserRuleContext) bool {
	switch ctx {
	case TYPE_DESC_IN_ANNOTATION_DECL,
		TYPE_DESC_BEFORE_IDENTIFIER,
		TYPE_DESC_BEFORE_IDENTIFIER_IN_GROUPING_KEY,
		TYPE_DESC_IN_RECORD_FIELD,
		TYPE_DESC_IN_PARAM,
		TYPE_DESC_IN_TYPE_BINDING_PATTERN,
		VAR_DECL_STARTED_WITH_DENTIFIER,
		TYPE_DESC_IN_PATH_PARAM,
		AMBIGUOUS_STMT:
		return false
	default:
		return true
	}
}

func (this *BallerinaParser) parseReturnsKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.RETURNS_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.RETURNS_KEYWORD)
		return this.parseReturnsKeyword()
	}
}

func (this *BallerinaParser) parseTypeDescriptor(context ParserRuleContext) internal.STNode {
	return this.parseTypeDescriptor(context, false, false, TypePrecedence.DEFAULT)
}

func (this *BallerinaParser) parseTypeDescriptor(context ParserRuleContext, precedence TypePrecedence) internal.STNode {
	return this.parseTypeDescriptor(context, false, false, precedence)
}

func (this *BallerinaParser) parseTypeDescriptor(qualifiers []STNode, context ParserRuleContext) internal.STNode {
	return this.parseTypeDescriptor(qualifiers, context, false, false, TypePrecedence.DEFAULT)
}

func (this *BallerinaParser) parseTypeDescriptorInExpression(isInConditionalExpr bool) internal.STNode {
	return this.parseTypeDescriptor(ParserRuleContext.TYPE_DESC_IN_EXPRESSION, false, isInConditionalExpr,
		TypePrecedence.DEFAULT)
}

func (this *BallerinaParser) parseTypeDescriptor(context ParserRuleContext, isTypedBindingPattern bool, isInConditionalExpr bool, precedence TypePrecedence) internal.STNode {
	typeDescQualifiers := make([]interface{}, 0)
	return this.parseTypeDescriptor(typeDescQualifiers, context, isTypedBindingPattern, isInConditionalExpr, precedence)
}

func (this *BallerinaParser) parseTypeDescriptor(qualifiers []STNode, context ParserRuleContext, isTypedBindingPattern bool, isInConditionalExpr bool, precedence TypePrecedence) internal.STNode {
	this.startContext(context)
	typeDesc := this.parseTypeDescriptorInternal(qualifiers, context, isTypedBindingPattern, isInConditionalExpr,
		precedence)
	this.endContext()
	return typeDesc
}

func (this *BallerinaParser) parseTypeDescriptorInternal(qualifiers []STNode, context ParserRuleContext, isTypedBindingPattern bool, isInConditionalExpr bool, precedence TypePrecedence) internal.STNode {
	typeDesc := this.parseTypeDescriptorInternal(qualifiers, context, isInConditionalExpr)
	if ((typeDesc.kind == SyntaxKind.VAR_TYPE_DESC) && (context != ParserRuleContext.TYPE_DESC_IN_TYPE_BINDING_PATTERN)) && (context != ParserRuleContext.TYPE_DESC_BEFORE_IDENTIFIER_IN_GROUPING_KEY) {
		missingToken := this.STNodeFactory.createMissingToken(SyntaxKind.IDENTIFIER_TOKEN)
		missingToken = this.SyntaxErrors.cloneWithLeadingInvalidNodeMinutiae(missingToken, typeDesc,
			DiagnosticErrorCode.ERROR_INVALID_USAGE_OF_VAR)
		typeDesc = this.STNodeFactory.createSimpleNameReferenceNode(missingToken)
	}
	return this.parseComplexTypeDescriptorInternal(typeDesc, context, isTypedBindingPattern, precedence)
}

func (this *BallerinaParser) parseComplexTypeDescriptor(typeDesc internal.STNode, context ParserRuleContext, isTypedBindingPattern bool) internal.STNode {
	this.startContext(context)
	complexTypeDesc := this.parseComplexTypeDescriptorInternal(typeDesc, context, isTypedBindingPattern,
		TypePrecedence.DEFAULT)
	this.endContext()
	return complexTypeDesc
}

func (this *BallerinaParser) parseComplexTypeDescriptorInternal(typeDesc internal.STNode, context ParserRuleContext, isTypedBindingPattern bool, precedence TypePrecedence) internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case QUESTION_MARK_TOKEN:
		if this.precedence.isHigherThanOrEqual(TypePrecedence.ARRAY_OR_OPTIONAL) {
			return typeDesc
		}
		isPossibleOptionalType := true
		nextNextToken := this.getNextNextToken()
		if ((context == ParserRuleContext.TYPE_DESC_IN_EXPRESSION) && (!this.isValidTypeContinuationToken(nextNextToken))) && this.isValidExprStart(nextNextToken.kind) {
			if nextNextToken.kind == OPEN_BRACE_TOKEN {
				grandParentCtx := this.this.errorHandler.getGrandParentContext()
				isPossibleOptionalType = ((grandParentCtx == ParserRuleContext.IF_BLOCK) || (grandParentCtx == ParserRuleContext.WHILE_BLOCK))
			} else {
				isPossibleOptionalType = false
			}
		}
		if !isPossibleOptionalType {
			return typeDesc
		}
		optionalTypeDes := this.parseOptionalTypeDescriptor(typeDesc)
		return this.parseComplexTypeDescriptorInternal(optionalTypeDes, context, isTypedBindingPattern, precedence)
	case OPEN_BRACKET_TOKEN:
		if isTypedBindingPattern {
			return typeDesc
		}
		if this.precedence.isHigherThanOrEqual(TypePrecedence.ARRAY_OR_OPTIONAL) {
			return typeDesc
		}
		arrayTypeDesc := this.parseArrayTypeDescriptor(typeDesc)
		return this.parseComplexTypeDescriptorInternal(arrayTypeDesc, context, false, precedence)
	case PIPE_TOKEN:
		if this.precedence.isHigherThanOrEqual(TypePrecedence.UNION) {
			return typeDesc
		}
		newTypeDesc := this.parseUnionTypeDescriptor(typeDesc, context, isTypedBindingPattern)
		return this.parseComplexTypeDescriptorInternal(newTypeDesc, context, isTypedBindingPattern, precedence)
	case BITWISE_AND_TOKEN:
		if this.precedence.isHigherThanOrEqual(TypePrecedence.INTERSECTION) {
			return typeDesc
		}
		newTypeDesc = this.parseIntersectionTypeDescriptor(typeDesc, context, isTypedBindingPattern)
		return this.parseComplexTypeDescriptorInternal(newTypeDesc, context, isTypedBindingPattern, precedence)
	default:
		return typeDesc
	}
}

func (this *BallerinaParser) isValidTypeContinuationToken(token internal.STToken) bool {
	switch token.kind {
	case QUESTION_MARK_TOKEN, OPEN_BRACKET_TOKEN, PIPE_TOKEN, BITWISE_AND_TOKEN:
		true
	default:
		false
	}
}

func (this *BallerinaParser) validateForUsageOfVar(typeDesc internal.STNode) internal.STNode {
	if typeDesc.kind != SyntaxKind.VAR_TYPE_DESC {
		return typeDesc
	}
	missingToken := this.STNodeFactory.createMissingToken(SyntaxKind.IDENTIFIER_TOKEN)
	missingToken = this.SyntaxErrors.cloneWithLeadingInvalidNodeMinutiae(missingToken, typeDesc,
		DiagnosticErrorCode.ERROR_INVALID_USAGE_OF_VAR)
	return this.STNodeFactory.createSimpleNameReferenceNode(missingToken)
}

func (this *BallerinaParser) parseTypeDescriptorInternal(qualifiers []STNode, context ParserRuleContext, isInConditionalExpr bool) internal.STNode {
	this.parseTypeDescQualifiers(qualifiers)
	nextToken := this.peek()
	if this.isQualifiedIdentifierPredeclaredPrefix(nextToken.kind) {
		return this.parseQualifiedTypeRefOrTypeDesc(qualifiers, isInConditionalExpr)
	}
	switch nextToken.kind {
	case IDENTIFIER_TOKEN:
		this.reportInvalidQualifierList(qualifiers)
		return this.parseTypeReference(isInConditionalExpr)
	case RECORD_KEYWORD:
		this.reportInvalidQualifierList(qualifiers)
		return this.parseRecordTypeDescriptor()
	case OBJECT_KEYWORD:
		objectTypeQualifiers := this.createObjectTypeQualNodeList(qualifiers)
		return this.parseObjectTypeDescriptor(consume(), objectTypeQualifiers)
	case OPEN_PAREN_TOKEN:
		this.reportInvalidQualifierList(qualifiers)
		return this.parseNilOrParenthesisedTypeDesc()
	case MAP_KEYWORD:
		this.reportInvalidQualifierList(qualifiers)
		return this.parseMapTypeDescriptor(consume())
	case STREAM_KEYWORD:
		this.reportInvalidQualifierList(qualifiers)
		return this.parseStreamTypeDescriptor(consume())
	case TABLE_KEYWORD:
		this.reportInvalidQualifierList(qualifiers)
		return this.parseTableTypeDescriptor(consume())
	case FUNCTION_KEYWORD:
		return this.parseFunctionTypeDesc(qualifiers)
	case OPEN_BRACKET_TOKEN:
		this.reportInvalidQualifierList(qualifiers)
		return this.parseTupleTypeDesc()
	case DISTINCT_KEYWORD:
		this.reportInvalidQualifierList(qualifiers)
		distinctKeyword := this.consume()
		return this.parseDistinctTypeDesc(distinctKeyword, context)
	case TRANSACTION_KEYWORD:
		this.reportInvalidQualifierList(qualifiers)
		return this.parseQualifiedIdentWithTransactionPrefix(context)
	default:
		if this.isParameterizedTypeToken(nextToken.kind) {
			this.reportInvalidQualifierList(qualifiers)
			return this.parseParameterizedTypeDescriptor(consume())
		}
		if this.isSingletonTypeDescStart(nextToken.kind, getNextNextToken()) {
			this.reportInvalidQualifierList(qualifiers)
			return this.parseSingletonTypeDesc()
		}
		if this.isSimpleType(nextToken.kind) {
			this.reportInvalidQualifierList(qualifiers)
			return this.parseSimpleTypeDescriptor()
		}
	}
	recoveryCtx := this.getTypeDescRecoveryCtx(qualifiers)
	solution := this.recover(peek(), recoveryCtx)
	if solution.action == Action.KEEP {
		this.reportInvalidQualifierList(qualifiers)
		return this.parseSingletonTypeDesc()
	}
	return this.parseTypeDescriptorInternal(qualifiers, context, isInConditionalExpr)
}

func (this *BallerinaParser) getTypeDescRecoveryCtx(qualifiers []STNode) ParserRuleContext {
	if this.qualifiers.isEmpty() {
		return ParserRuleContext.TYPE_DESCRIPTOR
	}
	lastQualifier := this.getLastNodeInList(qualifiers)
	switch lastQualifier.kind {
	case ISOLATED_KEYWORD:
		return ParserRuleContext.TYPE_DESC_WITHOUT_ISOLATED
	case TRANSACTIONAL_KEYWORD:
		return ParserRuleContext.FUNC_TYPE_DESC
	default:
		return ParserRuleContext.OBJECT_TYPE_DESCRIPTOR
	}
}

func (this *BallerinaParser) parseQualifiedIdentWithTransactionPrefix(context ParserRuleContext) internal.STNode {
	transactionKeyword := this.consume()
	identifier := this.STNodeFactory.createIdentifierToken(transactionKeyword.text(),
		transactionKeyword.leadingMinutiae(), transactionKeyword.trailingMinutiae())
	colon := this.SyntaxErrors.createMissingTokenWithDiagnostics(SyntaxKind.COLON_TOKEN,
		DiagnosticErrorCode.ERROR_MISSING_COLON_TOKEN)
	varOrFuncName := this.parseIdentifier(context)
	return this.createQualifiedNameReferenceNode(identifier, colon, varOrFuncName)
}

func (this *BallerinaParser) parseQualifiedTypeRefOrTypeDesc(qualifiers []STNode, isInConditionalExpr bool) internal.STNode {
	preDeclaredPrefix := this.consume()
	nextNextToken := this.getNextNextToken()
	if (preDeclaredPrefix.kind == SyntaxKind.TRANSACTION_KEYWORD) || (nextNextToken.kind == SyntaxKind.IDENTIFIER_TOKEN) {
		this.reportInvalidQualifierList(qualifiers)
		return this.parseQualifiedIdentifierWithPredeclPrefix(preDeclaredPrefix, isInConditionalExpr)
	}
	var context ParserRuleContext
	switch preDeclaredPrefix.kind {
	case MAP_KEYWORD:
		context = ParserRuleContext.MAP_TYPE_OR_TYPE_REF
		break
	case OBJECT_KEYWORD:
		context = ParserRuleContext.OBJECT_TYPE_OR_TYPE_REF
		break
	case STREAM_KEYWORD:
		context = ParserRuleContext.STREAM_TYPE_OR_TYPE_REF
		break
	case TABLE_KEYWORD:
		context = ParserRuleContext.TABLE_TYPE_OR_TYPE_REF
		break
	default:
		if this.isParameterizedTypeToken(preDeclaredPrefix.kind) {
			context = ParserRuleContext.PARAMETERIZED_TYPE_OR_TYPE_REF
		} else {
			context = ParserRuleContext.TYPE_DESC_RHS_OR_TYPE_REF
		}
	}
	solution := this.recover(peek(), context)
	if solution.action == Action.KEEP {
		this.reportInvalidQualifierList(qualifiers)
		return this.parseQualifiedIdentifierWithPredeclPrefix(preDeclaredPrefix, isInConditionalExpr)
	}
	return this.parseTypeDescStartWithPredeclPrefix(preDeclaredPrefix, qualifiers)
}

func (this *BallerinaParser) parseTypeDescStartWithPredeclPrefix(preDeclaredPrefix internal.STToken, qualifiers []STNode) internal.STNode {
	switch preDeclaredPrefix.kind {
	case MAP_KEYWORD:
		this.reportInvalidQualifierList(qualifiers)
		return this.parseMapTypeDescriptor(preDeclaredPrefix)
	case OBJECT_KEYWORD:
		objectTypeQualifiers := this.createObjectTypeQualNodeList(qualifiers)
		return this.parseObjectTypeDescriptor(preDeclaredPrefix, objectTypeQualifiers)
	case STREAM_KEYWORD:
		this.reportInvalidQualifierList(qualifiers)
		return this.parseStreamTypeDescriptor(preDeclaredPrefix)
	case TABLE_KEYWORD:
		this.reportInvalidQualifierList(qualifiers)
		return this.parseTableTypeDescriptor(preDeclaredPrefix)
	default:
		if this.isParameterizedTypeToken(preDeclaredPrefix.kind) {
			this.reportInvalidQualifierList(qualifiers)
			return this.parseParameterizedTypeDescriptor(preDeclaredPrefix)
		}
		return this.createBuiltinSimpleNameReference(preDeclaredPrefix)
	}
}

func (this *BallerinaParser) parseQualifiedIdentifierWithPredeclPrefix(preDeclaredPrefix internal.STToken, isInConditionalExpr bool) internal.STNode {
	identifier := this.STNodeFactory.createIdentifierToken(preDeclaredPrefix.text(),
		preDeclaredPrefix.leadingMinutiae(), preDeclaredPrefix.trailingMinutiae())
	return this.parseQualifiedIdentifier(identifier, isInConditionalExpr)
}

func (this *BallerinaParser) parseDistinctTypeDesc(distinctKeyword internal.STNode, context ParserRuleContext) internal.STNode {
	typeDesc := this.parseTypeDescriptor(context, TypePrecedence.DISTINCT)
	return this.STNodeFactory.createDistinctTypeDescriptorNode(distinctKeyword, typeDesc)
}

func (this *BallerinaParser) parseNilOrParenthesisedTypeDesc() internal.STNode {
	openParen := this.parseOpenParenthesis()
	return this.parseNilOrParenthesisedTypeDescRhs(openParen)
}

func (this *BallerinaParser) parseNilOrParenthesisedTypeDescRhs(openParen internal.STNode) internal.STNode {
	var closeParen internal.STNode
	nextToken := this.peek()
	switch nextToken.kind {
	case CLOSE_PAREN_TOKEN:
		closeParen = this.parseCloseParenthesis()
		return this.STNodeFactory.createNilTypeDescriptorNode(openParen, closeParen)
	default:
		if this.isTypeStartingToken(nextToken.kind) {
			typedesc := this.parseTypeDescriptor(ParserRuleContext.TYPE_DESC_IN_PARENTHESIS)
			closeParen = this.parseCloseParenthesis()
			return this.STNodeFactory.createParenthesisedTypeDescriptorNode(openParen, typedesc, closeParen)
		}
		this.recover(peek(), ParserRuleContext.NIL_OR_PARENTHESISED_TYPE_DESC_RHS)
		return this.parseNilOrParenthesisedTypeDescRhs(openParen)
	}
}

func (this *BallerinaParser) parseSimpleTypeInTerminalExpr() internal.STNode {
	this.startContext(ParserRuleContext.TYPE_DESC_IN_EXPRESSION)
	simpleTypeDescriptor := this.parseSimpleTypeDescriptor()
	this.endContext()
	return simpleTypeDescriptor
}

func (this *BallerinaParser) parseSimpleTypeDescriptor() internal.STNode {
	nextToken := this.peek()
	if this.isSimpleType(nextToken.kind) {
		token := this.consume()
		return this.createBuiltinSimpleNameReference(token)
	} else {
		this.recover(nextToken, ParserRuleContext.SIMPLE_TYPE_DESCRIPTOR)
		return this.parseSimpleTypeDescriptor()
	}
}

func (this *BallerinaParser) parseFunctionBody() internal.STNode {
	token := this.peek()
	switch token.kind {
	case EQUAL_TOKEN:
		return this.parseExternalFunctionBody()
	case OPEN_BRACE_TOKEN:
		return this.parseFunctionBodyBlock(false)
	case RIGHT_DOUBLE_ARROW_TOKEN:
		return this.parseExpressionFuncBody(false, false)
	default:
		this.recover(token, ParserRuleContext.FUNC_BODY)
		return this.parseFunctionBody()
	}
}

func (this *BallerinaParser) parseFunctionBodyBlock(isAnonFunc bool) internal.STNode {
	this.startContext(ParserRuleContext.FUNC_BODY_BLOCK)
	openBrace := this.parseOpenBrace()
	token := this.peek()
	firstStmtList := make([]interface{}, 0)
	workers := make([]interface{}, 0)
	secondStmtList := make([]interface{}, 0)
	currentCtx := ParserRuleContext.DEFAULT_WORKER_INIT
	hasNamedWorkers := false
	for !this.isEndOfFuncBodyBlock(token.kind, isAnonFunc) {
		stmt := this.parseStatement()
		if stmt == nil {
			break
		}
		if this.validateStatement(stmt) {
			continue
		}
		switch currentCtx {
		case DEFAULT_WORKER_INIT:
			if stmt.kind != SyntaxKind.NAMED_WORKER_DECLARATION {
				this.firstStmtList.add(stmt)
				break
			}
			currentCtx = ParserRuleContext.NAMED_WORKERS
			hasNamedWorkers = true
		case NAMED_WORKERS:
			if stmt.kind == SyntaxKind.NAMED_WORKER_DECLARATION {
				this.workers.add(stmt)
				break
			}
			currentCtx = ParserRuleContext.DEFAULT_WORKER
		case DEFAULT_WORKER:
		default:
			if stmt.kind == SyntaxKind.NAMED_WORKER_DECLARATION {
				this.updateLastNodeInListWithInvalidNode(secondStmtList, stmt,
					DiagnosticErrorCode.ERROR_NAMED_WORKER_NOT_ALLOWED_HERE)
				break
			}
			this.secondStmtList.add(stmt)
			break
		}
		token = this.peek()
	}
	var namedWorkersList internal.STNode
	var statements internal.STNode
	if hasNamedWorkers {
		workerInitStatements := this.STNodeFactory.createNodeList(firstStmtList)
		namedWorkers := this.STNodeFactory.createNodeList(workers)
		namedWorkersList = this.STNodeFactory.createNamedWorkerDeclarator(workerInitStatements, namedWorkers)
		statements = this.STNodeFactory.createNodeList(secondStmtList)
	} else {
		namedWorkersList = this.STNodeFactory.createEmptyNode()
		statements = this.STNodeFactory.createNodeList(firstStmtList)
	}
	closeBrace := this.parseCloseBrace()
	var semicolon internal.STNode
	if isAnonFunc {
		semicolon = this.STNodeFactory.createEmptyNode()
	} else {
		semicolon = this.parseOptionalSemicolon()
	}
	this.endContext()
	return this.STNodeFactory.createFunctionBodyBlockNode(openBrace, namedWorkersList, statements, closeBrace,
		semicolon)
}

func (this *BallerinaParser) isEndOfFuncBodyBlock(nextTokenKind SyntaxKind, isAnonFunc bool) bool {
	if isAnonFunc {
		switch nextTokenKind {
		case CLOSE_BRACE_TOKEN:
		case CLOSE_PAREN_TOKEN:
		case CLOSE_BRACKET_TOKEN:
		case OPEN_BRACE_TOKEN:
		case SEMICOLON_TOKEN:
		case COMMA_TOKEN:
		case PUBLIC_KEYWORD:
		case EOF_TOKEN:
		case EQUAL_TOKEN:
		case BACKTICK_TOKEN:
			return true
		default:
			break
		}
	}
	return this.isEndOfStatements()
}

func (this *BallerinaParser) isEndOfRecordTypeNode(nextTokenKind SyntaxKind) bool {
	return this.isEndOfModuleLevelNode(1)
}

func (this *BallerinaParser) isEndOfObjectTypeNode() bool {
	return this.isEndOfModuleLevelNode(1, true)
}

func (this *BallerinaParser) isEndOfStatements() bool {
	switch peek().kind {
	case RESOURCE_KEYWORD:
		true
	default:
		this.isEndOfModuleLevelNode(1)
	}
}

func (this *BallerinaParser) isEndOfModuleLevelNode(peekIndex int) bool {
	return this.isEndOfModuleLevelNode(peekIndex, false)
}

func (this *BallerinaParser) isEndOfModuleLevelNode(peekIndex int, isObject bool) bool {
	switch peek(peekIndex).kind {
	case EOF_TOKEN,
		CLOSE_BRACE_TOKEN,
		CLOSE_BRACE_PIPE_TOKEN,
		IMPORT_KEYWORD,
		ANNOTATION_KEYWORD,
		LISTENER_KEYWORD,
		CLASS_KEYWORD:
		return true
	case SERVICE_KEYWORD:
		return this.isServiceDeclStart(ParserRuleContext.OBJECT_CONSTRUCTOR_MEMBER, 1)
	case PUBLIC_KEYWORD:
		return ((!isObject) && this.isEndOfModuleLevelNode(peekIndex+1, false))
	case FUNCTION_KEYWORD:
		if isObject {
			return false
		}
		return ((peek(peekIndex+1).kind == SyntaxKind.IDENTIFIER_TOKEN) && (peek(peekIndex+2).kind == SyntaxKind.OPEN_PAREN_TOKEN))
	default:
		return false
	}
}

func (this *BallerinaParser) isEndOfParameter(tokenKind SyntaxKind) bool {
	switch tokenKind {
	case CLOSE_PAREN_TOKEN,
		CLOSE_BRACKET_TOKEN,
		SEMICOLON_TOKEN,
		COMMA_TOKEN,
		RETURNS_KEYWORD,
		TYPE_KEYWORD,
		IF_KEYWORD,
		WHILE_KEYWORD,
		DO_KEYWORD,
		AT_TOKEN:
		return true
	default:
		return this.isEndOfModuleLevelNode(1)
	}
}

func (this *BallerinaParser) isEndOfParametersList(tokenKind SyntaxKind) bool {
	switch tokenKind {
	case CLOSE_PAREN_TOKEN,
		SEMICOLON_TOKEN,
		RETURNS_KEYWORD,
		TYPE_KEYWORD,
		IF_KEYWORD,
		WHILE_KEYWORD,
		DO_KEYWORD,
		RIGHT_DOUBLE_ARROW_TOKEN:
		return true
	default:
		return this.isEndOfModuleLevelNode(1)
	}
}

func (this *BallerinaParser) parseStatementStartIdentifier() internal.STNode {
	return this.parseQualifiedIdentifier(ParserRuleContext.TYPE_NAME_OR_VAR_NAME)
}

func (this *BallerinaParser) parseVariableName() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.IDENTIFIER_TOKEN {
		return this.consume()
	} else {
		this.recover(peek(), ParserRuleContext.VARIABLE_NAME)
		return this.parseVariableName()
	}
}

func (this *BallerinaParser) parseOpenBrace() internal.STNode {
	token := this.peek()
	if token.kind == OPEN_BRACE_TOKEN {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.OPEN_BRACE)
		return this.parseOpenBrace()
	}
}

func (this *BallerinaParser) parseCloseBrace() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.CLOSE_BRACE_TOKEN {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.CLOSE_BRACE)
		return this.parseCloseBrace()
	}
}

func (this *BallerinaParser) parseExternalFunctionBody() internal.STNode {
	this.startContext(ParserRuleContext.EXTERNAL_FUNC_BODY)
	assign := this.parseAssignOp()
	return this.parseExternalFuncBodyRhs(assign)
}

func (this *BallerinaParser) parseExternalFuncBodyRhs(assign internal.STNode) internal.STNode {
	var annotation internal.STNode
	nextToken := this.peek()
	switch nextToken.kind {
	case AT_TOKEN:
		annotation = this.parseAnnotations()
		break
	case EXTERNAL_KEYWORD:
		annotation = this.STNodeFactory.createEmptyNodeList()
		break
	default:
		this.recover(nextToken, ParserRuleContext.EXTERNAL_FUNC_BODY_OPTIONAL_ANNOTS)
		return this.parseExternalFuncBodyRhs(assign)
	}
	externalKeyword := this.parseExternalKeyword()
	semicolon := this.parseSemicolon()
	this.endContext()
	return this.STNodeFactory.createExternalFunctionBodyNode(assign, annotation, externalKeyword, semicolon)
}

func (this *BallerinaParser) parseSemicolon() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.SEMICOLON_TOKEN {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.SEMICOLON)
		return this.parseSemicolon()
	}
}

func (this *BallerinaParser) parseOptionalSemicolon() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.SEMICOLON_TOKEN {
		return this.consume()
	}
	return this.STNodeFactory.createEmptyNode()
}

func (this *BallerinaParser) parseExternalKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.EXTERNAL_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.EXTERNAL_KEYWORD)
		return this.parseExternalKeyword()
	}
}

func (this *BallerinaParser) parseAssignOp() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.EQUAL_TOKEN {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.ASSIGN_OP)
		return this.parseAssignOp()
	}
}

func (this *BallerinaParser) parseBinaryOperator() internal.STNode {
	token := this.peek()
	if this.isBinaryOperator(token.kind) {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.BINARY_OPERATOR)
		return this.parseBinaryOperator()
	}
}

func (this *BallerinaParser) isBinaryOperator(kind SyntaxKind) bool {
	switch kind {
	case PLUS_TOKEN,
		MINUS_TOKEN,
		SLASH_TOKEN,
		ASTERISK_TOKEN,
		GT_TOKEN,
		LT_TOKEN,
		DOUBLE_EQUAL_TOKEN,
		TRIPPLE_EQUAL_TOKEN,
		LT_EQUAL_TOKEN,
		GT_EQUAL_TOKEN,
		NOT_EQUAL_TOKEN,
		NOT_DOUBLE_EQUAL_TOKEN,
		BITWISE_AND_TOKEN,
		BITWISE_XOR_TOKEN,
		PIPE_TOKEN,
		LOGICAL_AND_TOKEN,
		LOGICAL_OR_TOKEN,
		PERCENT_TOKEN,
		DOUBLE_LT_TOKEN,
		DOUBLE_GT_TOKEN,
		TRIPPLE_GT_TOKEN,
		ELLIPSIS_TOKEN,
		DOUBLE_DOT_LT_TOKEN,
		ELVIS_TOKEN:
		return true
	default:
		return false
	}
}

func (this *BallerinaParser) getOpPrecedence(binaryOpKind SyntaxKind) OperatorPrecedence {
	switch binaryOpKind {
	case ASTERISK_TOKEN, // multiplication
		SLASH_TOKEN, // division
		PERCENT_TOKEN:
		return OperatorPrecedence.MULTIPLICATIVE
	case PLUS_TOKEN,
		MINUS_TOKEN:
		return OperatorPrecedence.ADDITIVE
	case GT_TOKEN,
		LT_TOKEN,
		GT_EQUAL_TOKEN,
		LT_EQUAL_TOKEN,
		IS_KEYWORD,
		NOT_IS_KEYWORD:
		return OperatorPrecedence.BINARY_COMPARE
	case DOT_TOKEN,
		OPEN_BRACKET_TOKEN,
		OPEN_PAREN_TOKEN,
		ANNOT_CHAINING_TOKEN,
		OPTIONAL_CHAINING_TOKEN,
		DOT_LT_TOKEN,
		SLASH_LT_TOKEN,
		DOUBLE_SLASH_DOUBLE_ASTERISK_LT_TOKEN,
		SLASH_ASTERISK_TOKEN:
		return OperatorPrecedence.MEMBER_ACCESS
	case DOUBLE_EQUAL_TOKEN,
		TRIPPLE_EQUAL_TOKEN,
		NOT_EQUAL_TOKEN,
		NOT_DOUBLE_EQUAL_TOKEN:
		return OperatorPrecedence.EQUALITY
	case BITWISE_AND_TOKEN:
		return OperatorPrecedence.BITWISE_AND
	case BITWISE_XOR_TOKEN:
		return OperatorPrecedence.BITWISE_XOR
	case PIPE_TOKEN:
		return OperatorPrecedence.BITWISE_OR
	case LOGICAL_AND_TOKEN:
		return OperatorPrecedence.LOGICAL_AND
	case LOGICAL_OR_TOKEN:
		return OperatorPrecedence.LOGICAL_OR
	case RIGHT_ARROW_TOKEN:
		return OperatorPrecedence.REMOTE_CALL_ACTION
	case RIGHT_DOUBLE_ARROW_TOKEN:
		return OperatorPrecedence.ANON_FUNC_OR_LET
	case SYNC_SEND_TOKEN:
		return OperatorPrecedence.ACTION
	case DOUBLE_LT_TOKEN,
		DOUBLE_GT_TOKEN,
		TRIPPLE_GT_TOKEN:
		return OperatorPrecedence.SHIFT
	case ELLIPSIS_TOKEN,
		DOUBLE_DOT_LT_TOKEN:
		return OperatorPrecedence.RANGE
	case ELVIS_TOKEN:
		return OperatorPrecedence.ELVIS_CONDITIONAL
	case QUESTION_MARK_TOKEN,
		COLON_TOKEN:
		return OperatorPrecedence.CONDITIONAL
	default:
		panic("Unsupported binary operator '" + binaryOpKind + "'")
	}
}

func (this *BallerinaParser) getBinaryOperatorKindToInsert(opPrecedenceLevel OperatorPrecedence) SyntaxKind {
	switch opPrecedenceLevel {
	case MULTIPLICATIVE:
		return SyntaxKind.ASTERISK_TOKEN
	case DEFAULT,
		UNARY,
		ACTION,
		EXPRESSION_ACTION,
		REMOTE_CALL_ACTION,
		ANON_FUNC_OR_LET,
		QUERY,
		TRAP,
		ADDITIVE:
		return SyntaxKind.PLUS_TOKEN
	case SHIFT:
		return SyntaxKind.DOUBLE_LT_TOKEN
	case RANGE:
		return SyntaxKind.ELLIPSIS_TOKEN
	case BINARY_COMPARE:
		return SyntaxKind.LT_TOKEN
	case EQUALITY:
		return SyntaxKind.DOUBLE_EQUAL_TOKEN
	case BITWISE_AND:
		return SyntaxKind.BITWISE_AND_TOKEN
	case BITWISE_XOR:
		return SyntaxKind.BITWISE_XOR_TOKEN
	case BITWISE_OR:
		return SyntaxKind.PIPE_TOKEN
	case LOGICAL_AND:
		return SyntaxKind.LOGICAL_AND_TOKEN
	case LOGICAL_OR:
		return SyntaxKind.LOGICAL_OR_TOKEN
	case ELVIS_CONDITIONAL:
		return SyntaxKind.ELVIS_TOKEN
	default:
		panic(
			"Unsupported operator precedence level'" + opPrecedenceLevel + "'")
	}
}

func (this *BallerinaParser) getMissingBinaryOperatorContext(opPrecedenceLevel OperatorPrecedence) ParserRuleContext {
	switch opPrecedenceLevel {
	case MULTIPLICATIVE:
		return ParserRuleContext.ASTERISK
	case DEFAULT,
		UNARY,
		ACTION,
		EXPRESSION_ACTION,
		REMOTE_CALL_ACTION,
		ANON_FUNC_OR_LET,
		QUERY,
		TRAP,
		ADDITIVE:
		return ParserRuleContext.PLUS_TOKEN
	case SHIFT:
		return ParserRuleContext.DOUBLE_LT
	case RANGE:
		return ParserRuleContext.ELLIPSIS
	case BINARY_COMPARE:
		return ParserRuleContext.LT_TOKEN
	case EQUALITY:
		return ParserRuleContext.DOUBLE_EQUAL
	case BITWISE_AND:
		return ParserRuleContext.BITWISE_AND_OPERATOR
	case BITWISE_XOR:
		ParserRuleContext.BITWISE_XOR
	case BITWISE_OR:
		return ParserRuleContext.PIPE
	case LOGICAL_AND:
		return ParserRuleContext.LOGICAL_AND
	case LOGICAL_OR:
		return ParserRuleContext.LOGICAL_OR
	case ELVIS_CONDITIONAL:
		return ParserRuleContext.ELVIS
	default:
		panic(
			"Unsupported operator precedence level'" + opPrecedenceLevel + "'")
	}
}

func (this *BallerinaParser) parseModuleTypeDefinition(metadata internal.STNode, qualifier internal.STNode) internal.STNode {
	this.startContext(ParserRuleContext.MODULE_TYPE_DEFINITION)
	typeKeyword := this.parseTypeKeyword()
	typeName := this.parseTypeName()
	typeDescriptor := this.parseTypeDescriptor(ParserRuleContext.TYPE_DESC_IN_TYPE_DEF)
	semicolon := this.parseSemicolon()
	this.endContext()
	return this.STNodeFactory.createTypeDefinitionNode(metadata, qualifier, typeKeyword, typeName, typeDescriptor,
		semicolon)
}

func (this *BallerinaParser) parseClassDefinition(metadata internal.STNode, qualifier internal.STNode, qualifiers []STNode) internal.STNode {
	this.startContext(ParserRuleContext.MODULE_CLASS_DEFINITION)
	classTypeQualifiers := this.createClassTypeQualNodeList(qualifiers)
	classKeyword := this.parseClassKeyword()
	className := this.parseClassName()
	openBrace := this.parseOpenBrace()
	classMembers := this.parseObjectMembers(ParserRuleContext.CLASS_MEMBER)
	closeBrace := this.parseCloseBrace()
	semicolon := this.parseOptionalSemicolon()
	this.endContext()
	return this.STNodeFactory.createClassDefinitionNode(metadata, qualifier, classTypeQualifiers, classKeyword,
		className, openBrace, classMembers, closeBrace, semicolon)
}

func (this *BallerinaParser) isClassTypeQual(tokenKind SyntaxKind) bool {
	switch tokenKind {
	case READONLY_KEYWORD, DISTINCT_KEYWORD, ISOLATED_KEYWORD:
		return true
	default:
		return this.isObjectNetworkQual(tokenKind)
	}
}

func (this *BallerinaParser) isObjectTypeQual(tokenKind SyntaxKind) bool {
	switch tokenKind {
	case ISOLATED_KEYWORD:
		true
	default:
		this.isObjectNetworkQual(tokenKind)
	}
}

func (this *BallerinaParser) isObjectNetworkQual(tokenKind SyntaxKind) bool {
	switch tokenKind {
	case SERVICE_KEYWORD, CLIENT_KEYWORD:
		true
	default:
		false
	}
}

func (this *BallerinaParser) createClassTypeQualNodeList(qualifierList []STNode) internal.STNode {
	validatedList := make([]interface{}, 0)
	hasNetworkQual := false
	i := 0
	for ; i < len(qualifierList); i++ {
		qualifier := this.qualifierList.get(i)
		nextIndex := (i + 1)
		if this.isSyntaxKindInList(validatedList, qualifier.kind) {
			this.updateLastNodeInListWithInvalidNode(validatedList, qualifier,
				DiagnosticErrorCode.ERROR_DUPLICATE_QUALIFIER, (qualifier).text())
			continue
		}
		if this.isObjectNetworkQual(qualifier.kind) {
			if hasNetworkQual {
				this.updateLastNodeInListWithInvalidNode(validatedList, qualifier,
					DiagnosticErrorCode.ERROR_MORE_THAN_ONE_OBJECT_NETWORK_QUALIFIERS)
			} else {
				this.validatedList.add(qualifier)
				hasNetworkQual = true
			}
			continue
		}
		if this.isClassTypeQual(qualifier.kind) {
			this.validatedList.add(qualifier)
			continue
		}
		if len(qualifierList) == nextIndex {
			this.addInvalidNodeToNextToken(qualifier, DiagnosticErrorCode.ERROR_QUALIFIER_NOT_ALLOWED,
				(qualifier).text())
		} else {
			this.updateANodeInListWithLeadingInvalidNode(qualifierList, nextIndex, qualifier,
				DiagnosticErrorCode.ERROR_QUALIFIER_NOT_ALLOWED, (qualifier).text())
		}
	}
	return this.STNodeFactory.createNodeList(validatedList)
}

func (this *BallerinaParser) createObjectTypeQualNodeList(qualifierList []STNode) internal.STNode {
	validatedList := make([]interface{}, 0)
	hasNetworkQual := false
	i := 0
	for ; i < len(qualifierList); i++ {
		qualifier := this.qualifierList.get(i)
		nextIndex := (i + 1)
		if this.isSyntaxKindInList(validatedList, qualifier.kind) {
			this.updateLastNodeInListWithInvalidNode(validatedList, qualifier,
				DiagnosticErrorCode.ERROR_DUPLICATE_QUALIFIER, (qualifier).text())
			continue
		}
		if this.isObjectNetworkQual(qualifier.kind) {
			if hasNetworkQual {
				this.updateLastNodeInListWithInvalidNode(validatedList, qualifier,
					DiagnosticErrorCode.ERROR_MORE_THAN_ONE_OBJECT_NETWORK_QUALIFIERS)
			} else {
				this.validatedList.add(qualifier)
				hasNetworkQual = true
			}
			continue
		}
		if this.isObjectTypeQual(qualifier.kind) {
			this.validatedList.add(qualifier)
			continue
		}
		if len(qualifierList) == nextIndex {
			this.addInvalidNodeToNextToken(qualifier, DiagnosticErrorCode.ERROR_QUALIFIER_NOT_ALLOWED,
				(qualifier).text())
		} else {
			this.updateANodeInListWithLeadingInvalidNode(qualifierList, nextIndex, qualifier,
				DiagnosticErrorCode.ERROR_QUALIFIER_NOT_ALLOWED, (qualifier).text())
		}
	}
	return this.STNodeFactory.createNodeList(validatedList)
}

func (this *BallerinaParser) parseClassKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.CLASS_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.CLASS_KEYWORD)
		return this.parseClassKeyword()
	}
}

func (this *BallerinaParser) parseTypeKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.TYPE_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.TYPE_KEYWORD)
		return this.parseTypeKeyword()
	}
}

func (this *BallerinaParser) parseTypeName() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.IDENTIFIER_TOKEN {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.TYPE_NAME)
		return this.parseTypeName()
	}
}

func (this *BallerinaParser) parseClassName() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.IDENTIFIER_TOKEN {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.CLASS_NAME)
		return this.parseClassName()
	}
}

func (this *BallerinaParser) parseRecordTypeDescriptor() internal.STNode {
	this.startContext(ParserRuleContext.RECORD_TYPE_DESCRIPTOR)
	recordKeyword := this.parseRecordKeyword()
	bodyStartDelimiter := this.parseRecordBodyStartDelimiter()
	recordFields := make([]interface{}, 0)
	token := this.peek()
	recordRestDescriptor := this.STNodeFactory.createEmptyNode()
	for !this.isEndOfRecordTypeNode(token.kind) {
		field := this.parseFieldOrRestDescriptor()
		if field == nil {
			break
		}
		token = this.peek()
		if (field.kind == SyntaxKind.RECORD_REST_TYPE) && (bodyStartDelimiter.kind == OPEN_BRACE_TOKEN) {
			if this.recordFields.isEmpty() {
				bodyStartDelimiter = this.SyntaxErrors.cloneWithTrailingInvalidNodeMinutiae(bodyStartDelimiter, field,
					DiagnosticErrorCode.ERROR_INCLUSIVE_RECORD_TYPE_CANNOT_CONTAIN_REST_FIELD)
			} else {
				this.updateLastNodeInListWithInvalidNode(recordFields, field,
					DiagnosticErrorCode.ERROR_INCLUSIVE_RECORD_TYPE_CANNOT_CONTAIN_REST_FIELD)
			}
			continue
		} else if field.kind == SyntaxKind.RECORD_REST_TYPE {
			recordRestDescriptor = field
			for !this.isEndOfRecordTypeNode(token.kind) {
				invalidField := this.parseFieldOrRestDescriptor()
				if invalidField == nil {
					break
				}
				recordRestDescriptor = this.SyntaxErrors.cloneWithTrailingInvalidNodeMinutiae(recordRestDescriptor,
					invalidField, DiagnosticErrorCode.ERROR_MORE_RECORD_FIELDS_AFTER_REST_FIELD)
				token = this.peek()
			}
			break
		}
		this.recordFields.add(field)
	}
	fields := this.STNodeFactory.createNodeList(recordFields)
	bodyEndDelimiter := this.parseRecordBodyCloseDelimiter(bodyStartDelimiter.kind)
	this.endContext()
	return this.STNodeFactory.createRecordTypeDescriptorNode(recordKeyword, bodyStartDelimiter, fields,
		recordRestDescriptor, bodyEndDelimiter)
}

func (this *BallerinaParser) parseRecordBodyStartDelimiter() internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case OPEN_BRACE_PIPE_TOKEN:
		this.parseClosedRecordBodyStart()
	case OPEN_BRACE_TOKEN:
		this.parseOpenBrace()
	default:
		this.recover(nextToken, ParserRuleContext.RECORD_BODY_START)
		this.parseRecordBodyStartDelimiter()
	}
}

func (this *BallerinaParser) parseClosedRecordBodyStart() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.OPEN_BRACE_PIPE_TOKEN {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.CLOSED_RECORD_BODY_START)
		return this.parseClosedRecordBodyStart()
	}
}

func (this *BallerinaParser) parseRecordBodyCloseDelimiter(startingDelimeter SyntaxKind) internal.STNode {
	if startingDelimeter == SyntaxKind.OPEN_BRACE_PIPE_TOKEN {
		return this.parseClosedRecordBodyEnd()
	}
	return this.parseCloseBrace()
}

func (this *BallerinaParser) parseClosedRecordBodyEnd() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.CLOSE_BRACE_PIPE_TOKEN {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.CLOSED_RECORD_BODY_END)
		return this.parseClosedRecordBodyEnd()
	}
}

func (this *BallerinaParser) parseRecordKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.RECORD_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.RECORD_KEYWORD)
		return this.parseRecordKeyword()
	}
}

func (this *BallerinaParser) parseFieldOrRestDescriptor() internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case CLOSE_BRACE_TOKEN:
	case CLOSE_BRACE_PIPE_TOKEN:
		return nil
	case ASTERISK_TOKEN:
		this.startContext(ParserRuleContext.RECORD_FIELD)
		asterisk := this.consume()
		ty := this.parseTypeReferenceInTypeInclusion()
		semicolonToken := this.parseSemicolon()
		this.endContext()
		return this.STNodeFactory.createTypeReferenceNode(asterisk, ty, semicolonToken)
	case DOCUMENTATION_STRING:
	case AT_TOKEN:
		return this.parseRecordField()
	default:
		if this.isTypeStartingToken(nextToken.kind) {
			return this.parseRecordField()
		}
		this.recover(peek(), ParserRuleContext.RECORD_FIELD_OR_RECORD_END)
		return this.parseFieldOrRestDescriptor()
	}
}

func (this *BallerinaParser) parseRecordField() internal.STNode {
	this.startContext(ParserRuleContext.RECORD_FIELD)
	metadata := this.parseMetaData()
	fieldOrRestDesc := this.parseRecordField(peek(), metadata)
	this.endContext()
	return fieldOrRestDesc
}

func (this *BallerinaParser) parseRecordField(nextToken internal.STToken, metadata internal.STNode) internal.STNode {
	if nextToken.kind != SyntaxKind.READONLY_KEYWORD {
		ty := this.parseTypeDescriptor(ParserRuleContext.TYPE_DESC_IN_RECORD_FIELD)
		return this.parseFieldOrRestDescriptorRhs(metadata, ty)
	}
	var ty internal.STNode
	var readOnlyQualifier internal.STNode
	readOnlyQualifier = this.parseReadonlyKeyword()
	nextToken = this.peek()
	if nextToken.kind == SyntaxKind.IDENTIFIER_TOKEN {
		fieldNameOrTypeDesc := this.parseQualifiedIdentifier(ParserRuleContext.RECORD_FIELD_NAME_OR_TYPE_NAME)
		if fieldNameOrTypeDesc.kind == SyntaxKind.QUALIFIED_NAME_REFERENCE {
			ty = fieldNameOrTypeDesc
		} else {
			nextToken = this.peek()
			switch nextToken.kind {
			case SEMICOLON_TOKEN:
			case EQUAL_TOKEN:
				ty = this.createBuiltinSimpleNameReference(readOnlyQualifier)
				readOnlyQualifier = this.STNodeFactory.createEmptyNode()
				fieldName := (fieldNameOrTypeDesc).name
				return this.parseFieldDescriptorRhs(metadata, readOnlyQualifier, ty, fieldName)
			default:
				ty = this.parseComplexTypeDescriptor(fieldNameOrTypeDesc,
					ParserRuleContext.TYPE_DESC_IN_RECORD_FIELD, false)
			}
		}
	} else if nextToken.kind == SyntaxKind.ELLIPSIS_TOKEN {
		ty = this.createBuiltinSimpleNameReference(readOnlyQualifier)
		return this.parseFieldOrRestDescriptorRhs(metadata, ty)
	}
	return this.parseIndividualRecordField(metadata, readOnlyQualifier, ty)
}

func (this *BallerinaParser) parseIndividualRecordField(metadata internal.STNode, readOnlyQualifier internal.STNode, ty internal.STNode) internal.STNode {
	fieldName := this.parseVariableName()
	return this.parseFieldDescriptorRhs(metadata, readOnlyQualifier, ty, fieldName)
}

func (this *BallerinaParser) parseTypeReferenceInTypeInclusion() internal.STNode {
	typeReference := this.parseTypeDescriptor(ParserRuleContext.TYPE_REFERENCE_IN_TYPE_INCLUSION)
	if typeReference.kind == SyntaxKind.SIMPLE_NAME_REFERENCE {
		if this.typeReference.hasDiagnostics() {
			emptyNameReference := this.STNodeFactory.createSimpleNameReferenceNode
			(SyntaxErrors.createMissingTokenWithDiagnostics(SyntaxKind.IDENTIFIER_TOKEN,
				DiagnosticErrorCode.ERROR_MISSING_IDENTIFIER))
			return emptyNameReference
		}
		return typeReference
	}
	if typeReference.kind == SyntaxKind.QUALIFIED_NAME_REFERENCE {
		return typeReference
	}
	emptyNameReference := this.STNodeFactory.createSimpleNameReferenceNode(SyntaxErrors.createMissingToken(SyntaxKind.IDENTIFIER_TOKEN))
	emptyNameReference = this.SyntaxErrors.cloneWithTrailingInvalidNodeMinutiae(emptyNameReference, typeReference,
		DiagnosticErrorCode.ERROR_ONLY_TYPE_REFERENCE_ALLOWED_AS_TYPE_INCLUSIONS)
	return emptyNameReference
}

func (this *BallerinaParser) parseTypeReference() internal.STNode {
	return this.parseTypeReference(false)
}

func (this *BallerinaParser) parseTypeReference(isInConditionalExpr bool) internal.STNode {
	return this.parseQualifiedIdentifier(ParserRuleContext.TYPE_REFERENCE, isInConditionalExpr)
}

func (this *BallerinaParser) parseQualifiedIdentifier(currentCtx ParserRuleContext) internal.STNode {
	return this.parseQualifiedIdentifier(currentCtx, false)
}

func (this *BallerinaParser) parseQualifiedIdentifier(currentCtx ParserRuleContext, isInConditionalExpr bool) internal.STNode {
	token := this.peek()
	var typeRefOrPkgRef internal.STNode
	if token.kind == SyntaxKind.IDENTIFIER_TOKEN {
		typeRefOrPkgRef = this.consume()
	} else if this.isQualifiedIdentifierPredeclaredPrefix(token.kind) {
		preDeclaredPrefix := this.consume()
		typeRefOrPkgRef = this.STNodeFactory.createIdentifierToken(preDeclaredPrefix.text(),
			preDeclaredPrefix.leadingMinutiae(), preDeclaredPrefix.trailingMinutiae())
	}
	return this.parseQualifiedIdentifier(typeRefOrPkgRef, isInConditionalExpr)
}

func (this *BallerinaParser) parseQualifiedIdentifier(identifier internal.STNode, isInConditionalExpr bool) internal.STNode {
	nextToken := this.peek(1)
	if nextToken.kind != SyntaxKind.COLON_TOKEN {
		return this.STNodeFactory.createSimpleNameReferenceNode(identifier)
	}
	if isInConditionalExpr && (this.hasTrailingMinutiae(identifier) || this.hasTrailingMinutiae(nextToken)) {
		return this.ConditionalExprResolver.getSimpleNameRefNode(identifier)
	}
	nextNextToken := this.peek(2)
	switch nextNextToken.kind {
	case IDENTIFIER_TOKEN:
		colon := this.consume()
		varOrFuncName := this.consume()
		return this.createQualifiedNameReferenceNode(identifier, colon, varOrFuncName)
	case COLON_TOKEN:
		this.addInvalidTokenToNextToken(errorHandler.consumeInvalidToken())
		return this.parseQualifiedIdentifier(identifier, isInConditionalExpr)
	default:
		if (nextNextToken.kind == SyntaxKind.MAP_KEYWORD) && (peek(3).kind != SyntaxKind.LT_TOKEN) {
			colon = this.consume()
			mapKeyword := this.consume()
			refName := this.STNodeFactory.createIdentifierToken(mapKeyword.text(),
				mapKeyword.leadingMinutiae(), mapKeyword.trailingMinutiae(), mapKeyword.diagnostics())
			return this.createQualifiedNameReferenceNode(identifier, colon, refName)
		}
		if isInConditionalExpr {
			return this.ConditionalExprResolver.getSimpleNameRefNode(identifier)
		}
		colon = this.consume()
		varOrFuncName = this.SyntaxErrors.createMissingTokenWithDiagnostics(SyntaxKind.IDENTIFIER_TOKEN,
			DiagnosticErrorCode.ERROR_MISSING_IDENTIFIER)
		return this.createQualifiedNameReferenceNode(identifier, colon, varOrFuncName)
	}
}

func (this *BallerinaParser) createQualifiedNameReferenceNode(identifier internal.STNode, colon internal.STNode, varOrFuncName internal.STNode) internal.STNode {
	if this.hasTrailingMinutiae(identifier) || this.hasTrailingMinutiae(colon) {
		colon = this.SyntaxErrors.addDiagnostic(colon,
			DiagnosticErrorCode.ERROR_INTERVENING_WHITESPACES_ARE_NOT_ALLOWED)
	}
	return this.STNodeFactory.createQualifiedNameReferenceNode(identifier, colon, varOrFuncName)
}

func (this *BallerinaParser) parseFieldOrRestDescriptorRhs(metadata internal.STNode, ty internal.STNode) internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case ELLIPSIS_TOKEN:
		this.reportInvalidMetaData(metadata, "record rest descriptor")
		ellipsis := this.parseEllipsis()
		semicolonToken := this.parseSemicolon()
		return this.STNodeFactory.createRecordRestDescriptorNode(ty, ellipsis, semicolonToken)
	case IDENTIFIER_TOKEN:
		readonlyQualifier := this.STNodeFactory.createEmptyNode()
		return this.parseIndividualRecordField(metadata, readonlyQualifier, ty)
	default:
		this.recover(nextToken, ParserRuleContext.FIELD_OR_REST_DESCIPTOR_RHS)
		return this.parseFieldOrRestDescriptorRhs(metadata, ty)
	}
}

func (this *BallerinaParser) parseFieldDescriptorRhs(metadata internal.STNode, readonlyQualifier internal.STNode, ty internal.STNode, fieldName internal.STNode) internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case SEMICOLON_TOKEN:
		questionMarkToken := this.STNodeFactory.createEmptyNode()
		semicolonToken := this.parseSemicolon()
		return this.STNodeFactory.createRecordFieldNode(metadata, readonlyQualifier, ty, fieldName,
			questionMarkToken, semicolonToken)
	case QUESTION_MARK_TOKEN:
		questionMarkToken = this.parseQuestionMark()
		semicolonToken = this.parseSemicolon()
		return this.STNodeFactory.createRecordFieldNode(metadata, readonlyQualifier, ty, fieldName,
			questionMarkToken, semicolonToken)
	case EQUAL_TOKEN:
		equalsToken := this.parseAssignOp()
		expression := this.parseExpression()
		semicolonToken = this.parseSemicolon()
		return this.STNodeFactory.createRecordFieldWithDefaultValueNode(metadata, readonlyQualifier, ty, fieldName,
			equalsToken, expression, semicolonToken)
	default:
		this.recover(nextToken, ParserRuleContext.FIELD_DESCRIPTOR_RHS)
		return this.parseFieldDescriptorRhs(metadata, readonlyQualifier, ty, fieldName)
	}
}

func (this *BallerinaParser) parseQuestionMark() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.QUESTION_MARK_TOKEN {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.QUESTION_MARK)
		return this.parseQuestionMark()
	}
}

func (this *BallerinaParser) parseStatements() internal.STNode {
	stmts := make([]interface{}, 0)
	return this.parseStatements(stmts)
}

func (this *BallerinaParser) parseStatements(stmts []STNode) internal.STNode {
	for !this.isEndOfStatements() {
		stmt := this.parseStatement()
		if stmt == nil {
			break
		}
		if stmt.kind == SyntaxKind.NAMED_WORKER_DECLARATION {
			this.addInvalidNodeToNextToken(stmt, DiagnosticErrorCode.ERROR_NAMED_WORKER_NOT_ALLOWED_HERE)
			continue
		}
		if this.validateStatement(stmt) {
			continue
		}
		this.stmts.add(stmt)
	}
	return this.STNodeFactory.createNodeList(stmts)
}

func (this *BallerinaParser) parseStatement() internal.STNode {
	nextToken := this.peek()
	annots := this.STNodeFactory.createEmptyNodeList()
	switch nextToken.kind {
	case CLOSE_BRACE_TOKEN:
	case EOF_TOKEN:
		return nil
	case SEMICOLON_TOKEN:
		this.addInvalidTokenToNextToken(errorHandler.consumeInvalidToken())
		return this.parseStatement()
	case AT_TOKEN:
		annots = this.parseOptionalAnnotations()
		break
	default:
		if this.isStatementStartingToken(nextToken.kind) {
			break
		}
		token := this.peek()
		solution := this.recover(token, ParserRuleContext.STATEMENT)
		if solution.action == Action.KEEP {
			break
		}
		return this.parseStatement()
	}
	return this.parseStatement(annots)
}

func (this *BallerinaParser) validateStatement(statement internal.STNode) bool {
	switch statement.kind {
	case LOCAL_TYPE_DEFINITION_STATEMENT:
		this.addInvalidNodeToNextToken(statement, DiagnosticErrorCode.ERROR_LOCAL_TYPE_DEFINITION_NOT_ALLOWED)
		true
	case CONST_DECLARATION:
		this.addInvalidNodeToNextToken(statement, DiagnosticErrorCode.ERROR_LOCAL_CONST_DECL_NOT_ALLOWED)
		true
	default:
		false
	}
}

func (this *BallerinaParser) getAnnotations(nullbaleAnnot internal.STNode) internal.STNode {
	if nullbaleAnnot != nil {
		return nullbaleAnnot
	}
	return this.STNodeFactory.createEmptyNodeList()
}

func (this *BallerinaParser) parseStatement(annots internal.STNode) internal.STNode {
	typeDescQualifiers := make([]interface{}, 0)
	return this.parseStatement(annots, typeDescQualifiers)
}

func (this *BallerinaParser) parseStatement(annots internal.STNode, qualifiers []STNode) internal.STNode {
	this.parseTypeDescQualifiers(qualifiers)
	nextToken := this.peek()
	if this.isPredeclaredIdentifier(nextToken.kind) {
		return this.parseStmtStartsWithTypeOrExpr(getAnnotations(annots), qualifiers)
	}
	switch nextToken.kind {
	case CLOSE_BRACE_TOKEN:
	case EOF_TOKEN:
		publicQualifier := this.STNodeFactory.createEmptyNode()
		return this.createMissingSimpleVarDecl(getAnnotations(annots), publicQualifier, qualifiers, false)
	case SEMICOLON_TOKEN:
		this.addInvalidTokenToNextToken(errorHandler.consumeInvalidToken())
		return this.parseStatement(annots, qualifiers)
	case FINAL_KEYWORD:
		this.reportInvalidQualifierList(qualifiers)
		finalKeyword := this.consume()
		return this.parseVariableDecl(getAnnotations(annots), finalKeyword)
	case IF_KEYWORD:
		this.reportInvalidStatementAnnots(annots, qualifiers)
		this.reportInvalidQualifierList(qualifiers)
		return this.parseIfElseBlock()
	case WHILE_KEYWORD:
		this.reportInvalidStatementAnnots(annots, qualifiers)
		this.reportInvalidQualifierList(qualifiers)
		return this.parseWhileStatement()
	case DO_KEYWORD:
		this.reportInvalidStatementAnnots(annots, qualifiers)
		this.reportInvalidQualifierList(qualifiers)
		return this.parseDoStatement()
	case PANIC_KEYWORD:
		this.reportInvalidStatementAnnots(annots, qualifiers)
		this.reportInvalidQualifierList(qualifiers)
		return this.parsePanicStatement()
	case CONTINUE_KEYWORD:
		this.reportInvalidStatementAnnots(annots, qualifiers)
		this.reportInvalidQualifierList(qualifiers)
		return this.parseContinueStatement()
	case BREAK_KEYWORD:
		this.reportInvalidStatementAnnots(annots, qualifiers)
		this.reportInvalidQualifierList(qualifiers)
		return this.parseBreakStatement()
	case RETURN_KEYWORD:
		this.reportInvalidStatementAnnots(annots, qualifiers)
		this.reportInvalidQualifierList(qualifiers)
		return this.parseReturnStatement()
	case FAIL_KEYWORD:
		this.reportInvalidStatementAnnots(annots, qualifiers)
		this.reportInvalidQualifierList(qualifiers)
		return this.parseFailStatement()
	case TYPE_KEYWORD:
		this.reportInvalidQualifierList(qualifiers)
		return this.parseLocalTypeDefinitionStatement(getAnnotations(annots))
	case CONST_KEYWORD:
		this.reportInvalidQualifierList(qualifiers)
		return this.parseConstantDeclaration(annots, STNodeFactory.createEmptyNode())
	case LOCK_KEYWORD:
		this.reportInvalidStatementAnnots(annots, qualifiers)
		this.reportInvalidQualifierList(qualifiers)
		return this.parseLockStatement()
	case OPEN_BRACE_TOKEN:
		this.reportInvalidStatementAnnots(annots, qualifiers)
		this.reportInvalidQualifierList(qualifiers)
		return this.parseStatementStartsWithOpenBrace()
	case WORKER_KEYWORD:
		return this.parseNamedWorkerDeclaration(getAnnotations(annots), qualifiers)
	case FORK_KEYWORD:
		this.reportInvalidStatementAnnots(annots, qualifiers)
		this.reportInvalidQualifierList(qualifiers)
		return this.parseForkStatement()
	case FOREACH_KEYWORD:
		this.reportInvalidStatementAnnots(annots, qualifiers)
		this.reportInvalidQualifierList(qualifiers)
		return this.parseForEachStatement()
	case START_KEYWORD:
	case CHECK_KEYWORD:
	case CHECKPANIC_KEYWORD:
	case TRAP_KEYWORD:
	case FLUSH_KEYWORD:
	case LEFT_ARROW_TOKEN:
	case WAIT_KEYWORD:
	case FROM_KEYWORD:
	case COMMIT_KEYWORD:
		this.reportInvalidQualifierList(qualifiers)
		return this.parseExpressionStatement(getAnnotations(annots))
	case XMLNS_KEYWORD:
		this.reportInvalidStatementAnnots(annots, qualifiers)
		this.reportInvalidQualifierList(qualifiers)
		return this.parseXMLNamespaceDeclaration(false)
	case TRANSACTION_KEYWORD:
		return this.parseTransactionStmtOrVarDecl(annots, qualifiers, consume())
	case RETRY_KEYWORD:
		this.reportInvalidStatementAnnots(annots, qualifiers)
		this.reportInvalidQualifierList(qualifiers)
		return this.parseRetryStatement()
	case ROLLBACK_KEYWORD:
		this.reportInvalidStatementAnnots(annots, qualifiers)
		this.reportInvalidQualifierList(qualifiers)
		return this.parseRollbackStatement()
	case OPEN_BRACKET_TOKEN:
		this.reportInvalidQualifierList(qualifiers)
		return this.parseStatementStartsWithOpenBracket(getAnnotations(annots), false)
	case FUNCTION_KEYWORD:
	case OPEN_PAREN_TOKEN:
	case DECIMAL_INTEGER_LITERAL_TOKEN:
	case HEX_INTEGER_LITERAL_TOKEN:
	case STRING_LITERAL_TOKEN:
	case NULL_KEYWORD:
	case TRUE_KEYWORD:
	case FALSE_KEYWORD:
	case DECIMAL_FLOATING_POINT_LITERAL_TOKEN:
	case HEX_FLOATING_POINT_LITERAL_TOKEN:
	case STRING_KEYWORD:
	case XML_KEYWORD:
		return this.parseStmtStartsWithTypeOrExpr(getAnnotations(annots), qualifiers)
	case MATCH_KEYWORD:
		this.reportInvalidStatementAnnots(annots, qualifiers)
		this.reportInvalidQualifierList(qualifiers)
		return this.parseMatchStatement()
	case ERROR_KEYWORD:
		this.reportInvalidQualifierList(qualifiers)
		return this.parseErrorTypeDescOrErrorBP(getAnnotations(annots))
	default:
		if this.isValidExpressionStart(nextToken.kind, 1) {
			this.reportInvalidQualifierList(qualifiers)
			return this.parseStatementStartWithExpr(getAnnotations(annots))
		}
		if this.isTypeStartingToken(nextToken.kind) {
			publicQualifier = this.STNodeFactory.createEmptyNode()
			return this.parseVariableDecl(getAnnotations(annots), publicQualifier, nil, qualifiers,
				false)
		}
		token := this.peek()
		solution := this.recover(token, ParserRuleContext.STATEMENT_WITHOUT_ANNOTS)
		if solution.action == Action.KEEP {
			this.reportInvalidQualifierList(qualifiers)
			finalKeyword = this.STNodeFactory.createEmptyNode()
			return this.parseVariableDecl(getAnnotations(annots), finalKeyword)
		}
		return this.parseStatement(annots, qualifiers)
	}
}

func (this *BallerinaParser) parseVariableDecl(annots internal.STNode, finalKeyword internal.STNode) internal.STNode {
	typeDescQualifiers := make([]interface{}, 0)
	varDecQualifiers := make([]interface{}, 0)
	if finalKeyword != nil {
		this.varDecQualifiers.add(finalKeyword)
	}
	publicQualifier := this.STNodeFactory.createEmptyNode()
	return this.parseVariableDecl(annots, publicQualifier, varDecQualifiers, typeDescQualifiers, false)
}

func (this *BallerinaParser) parseVariableDecl(annots internal.STNode, publicQualifier internal.STNode, varDeclQuals []STNode, typeDescQualifiers []STNode, isModuleVar bool) internal.STNode {
	this.startContext(ParserRuleContext.VAR_DECL_STMT)
	typeBindingPattern := this.parseTypedBindingPattern(typeDescQualifiers,
		ParserRuleContext.VAR_DECL_STMT)
	return this.parseVarDeclRhs(annots, publicQualifier, varDeclQuals, typeBindingPattern, isModuleVar)
}

func (this *BallerinaParser) parseVarDeclTypeDescRhs(typeDesc internal.STNode, metadata internal.STNode, qualifiers []STNode, isTypedBindingPattern bool, isModuleVar bool) internal.STNode {
	publicQualifier := this.STNodeFactory.createEmptyNode()
	return this.parseVarDeclTypeDescRhs(typeDesc, metadata, publicQualifier, qualifiers, isTypedBindingPattern,
		isModuleVar)
}

func (this *BallerinaParser) parseVarDeclTypeDescRhs(typeDesc internal.STNode, metadata internal.STNode, publicQual internal.STNode, qualifiers []STNode, isTypedBindingPattern bool, isModuleVar bool) internal.STNode {
	this.startContext(ParserRuleContext.VAR_DECL_STMT)
	typeDesc = this.parseComplexTypeDescriptor(typeDesc,
		ParserRuleContext.TYPE_DESC_IN_TYPE_BINDING_PATTERN, isTypedBindingPattern)
	typedBindingPattern := this.parseTypedBindingPatternTypeRhs(typeDesc,
		ParserRuleContext.VAR_DECL_STMT)
	return this.parseVarDeclRhs(metadata, publicQual, qualifiers, typedBindingPattern, isModuleVar)
}

func (this *BallerinaParser) parseVarDeclRhs(metadata internal.STNode, varDeclQuals []STNode, typedBindingPattern internal.STNode, isModuleVar bool) internal.STNode {
	publicQualifier := this.STNodeFactory.createEmptyNode()
	return this.parseVarDeclRhs(metadata, publicQualifier, varDeclQuals, typedBindingPattern, isModuleVar)
}

func (this *BallerinaParser) parseVarDeclRhs(metadata internal.STNode, publicQualifier internal.STNode, varDeclQuals []STNode, typedBindingPattern internal.STNode, isModuleVar bool) internal.STNode {
	var assign internal.STNode
	var expr internal.STNode
	var semicolon internal.STNode
	hasVarInit := false
	isConfigurable := false
	if isModuleVar && this.isSyntaxKindInList(varDeclQuals, SyntaxKind.CONFIGURABLE_KEYWORD) {
		isConfigurable = true
	}
	nextToken := this.peek()
	switch nextToken.kind {
	case EQUAL_TOKEN:
		assign = this.parseAssignOp()
		if isModuleVar {
			if isConfigurable {
				expr = this.parseConfigurableVarDeclRhs()
			} else {
				expr = this.parseExpression()
			}
		} else {
			expr = this.parseActionOrExpression()
		}
		semicolon = this.parseSemicolon()
		hasVarInit = true
		break
	case SEMICOLON_TOKEN:
		assign = this.STNodeFactory.createEmptyNode()
		expr = this.STNodeFactory.createEmptyNode()
		semicolon = this.parseSemicolon()
		break
	default:
		this.recover(nextToken, ParserRuleContext.VAR_DECL_STMT_RHS)
		return this.parseVarDeclRhs(metadata, publicQualifier, varDeclQuals, typedBindingPattern, isModuleVar)
	}
	this.endContext()
	if !hasVarInit {
		bindingPatternKind := (typedBindingPattern).bindingPattern.kind
		if bindingPatternKind != SyntaxKind.CAPTURE_BINDING_PATTERN {
			assign = this.SyntaxErrors.createMissingTokenWithDiagnostics(SyntaxKind.EQUAL_TOKEN,
				DiagnosticErrorCode.ERROR_VARIABLE_DECL_HAVING_BP_MUST_BE_INITIALIZED)
			identifier := this.SyntaxErrors.createMissingToken(SyntaxKind.IDENTIFIER_TOKEN)
			expr = this.STNodeFactory.createSimpleNameReferenceNode(identifier)
		}
	}
	if isModuleVar {
		return this.createModuleVarDeclaration(metadata, publicQualifier, varDeclQuals, typedBindingPattern, assign,
			expr, semicolon, isConfigurable, hasVarInit)
	}
	var finalKeyword internal.STNode
	if this.varDeclQuals.isEmpty() {
		finalKeyword = this.STNodeFactory.createEmptyNode()
	} else {
		finalKeyword = this.varDeclQuals.get(0)
	}
	if metadata.kind == SyntaxKind.LIST {
		panic("assertion failed")
	}
	return this.STNodeFactory.createVariableDeclarationNode(metadata, finalKeyword, typedBindingPattern, assign,
		expr, semicolon)
}

func (this *BallerinaParser) parseConfigurableVarDeclRhs() internal.STNode {
	var expr internal.STNode
	nextToken := this.peek()
	switch nextToken.kind {
	case QUESTION_MARK_TOKEN:
		expr = this.STNodeFactory.createRequiredExpressionNode(consume())
		break
	default:
		if this.isValidExprStart(nextToken.kind) {
			expr = this.parseExpression()
			break
		}
		this.recover(nextToken, ParserRuleContext.CONFIG_VAR_DECL_RHS)
		return this.parseConfigurableVarDeclRhs()
	}
	return expr
}

func (this *BallerinaParser) createModuleVarDeclaration(metadata internal.STNode, publicQualifier internal.STNode, varDeclQuals []STNode, typedBindingPattern internal.STNode, assign internal.STNode, expr internal.STNode, semicolon internal.STNode, isConfigurable bool, hasVarInit bool) internal.STNode {
	if hasVarInit || this.varDeclQuals.isEmpty() {
		return this.createModuleVarDeclaration(metadata, publicQualifier, varDeclQuals, typedBindingPattern, assign,
			expr, semicolon)
	}
	if isConfigurable {
		return this.createConfigurableModuleVarDeclWithMissingInitializer(metadata, publicQualifier, varDeclQuals,
			typedBindingPattern, semicolon)
	}
	lastQualifier := this.getLastNodeInList(varDeclQuals)
	if lastQualifier.kind == SyntaxKind.ISOLATED_KEYWORD {
		lastQualifier = this.varDeclQuals.remove(varDeclQuals.size() - 1)
		typedBindingPattern = this.modifyTypedBindingPatternWithIsolatedQualifier(typedBindingPattern, lastQualifier)
	}
	return this.createModuleVarDeclaration(metadata, publicQualifier, varDeclQuals, typedBindingPattern, assign, expr,
		semicolon)
}

func (this *BallerinaParser) createConfigurableModuleVarDeclWithMissingInitializer(metadata internal.STNode, publicQualifier internal.STNode, varDeclQuals []STNode, typedBindingPattern internal.STNode, semicolon internal.STNode) internal.STNode {
	assign := this.SyntaxErrors.createMissingToken(SyntaxKind.EQUAL_TOKEN)
	assign = this.SyntaxErrors.addDiagnostic(assign,
		DiagnosticErrorCode.ERROR_CONFIGURABLE_VARIABLE_MUST_BE_INITIALIZED_OR_REQUIRED)
	questionMarkToken := this.SyntaxErrors.createMissingToken(SyntaxKind.QUESTION_MARK_TOKEN)
	expr := this.STNodeFactory.createRequiredExpressionNode(questionMarkToken)
	return this.createModuleVarDeclaration(metadata, publicQualifier, varDeclQuals, typedBindingPattern, assign, expr,
		semicolon)
}

func (this *BallerinaParser) createModuleVarDeclaration(metadata internal.STNode, publicQualifier internal.STNode, varDeclQuals []STNode, typedBindingPattern internal.STNode, assign internal.STNode, expr internal.STNode, semicolon internal.STNode) internal.STNode {
	if publicQualifier != nil {
		if (typedBindingPattern).typeDescriptor.kind == SyntaxKind.VAR_TYPE_DESC {
			if !this.varDeclQuals.isEmpty() {
				this.updateFirstNodeInListWithLeadingInvalidNode(varDeclQuals, publicQualifier,
					DiagnosticErrorCode.ERROR_VARIABLE_DECLARED_WITH_VAR_CANNOT_BE_PUBLIC)
			} else {
				typedBindingPattern = this.SyntaxErrors.cloneWithLeadingInvalidNodeMinutiae(typedBindingPattern,
					publicQualifier, DiagnosticErrorCode.ERROR_VARIABLE_DECLARED_WITH_VAR_CANNOT_BE_PUBLIC)
			}
			publicQualifier = this.STNodeFactory.createEmptyNode()
		} else if this.isSyntaxKindInList(varDeclQuals, SyntaxKind.ISOLATED_KEYWORD) {
			this.updateFirstNodeInListWithLeadingInvalidNode(varDeclQuals, publicQualifier,
				DiagnosticErrorCode.ERROR_ISOLATED_VAR_CANNOT_BE_DECLARED_AS_PUBLIC)
			publicQualifier = this.STNodeFactory.createEmptyNode()
		}
	}
	varDeclQualifiersNode := this.STNodeFactory.createNodeList(varDeclQuals)
	return this.STNodeFactory.createModuleVariableDeclarationNode(metadata, publicQualifier, varDeclQualifiersNode,
		typedBindingPattern, assign, expr, semicolon)
}

func (this *BallerinaParser) createMissingSimpleVarDecl(isModuleVar bool) internal.STNode {
	var metadata internal.STNode
	if isModuleVar {
		metadata = this.STNodeFactory.createEmptyNode()
	} else {
		metadata = this.STNodeFactory.createEmptyNodeList()
	}
	return this.createMissingSimpleVarDeclInner(metadata, isModuleVar)
}

func (this *BallerinaParser) createMissingSimpleVarDeclInner(metadata internal.STNode, isModuleVar bool) internal.STNode {
	publicQualifier := this.STNodeFactory.createEmptyNode()
	return this.createMissingSimpleVarDecl(metadata, publicQualifier, nil, isModuleVar)
}

func (this *BallerinaParser) createMissingSimpleVarDecl(metadata internal.STNode, publicQualifier internal.STNode, qualifiers []STNode, isModuleVar bool) internal.STNode {
	emptyNode := this.STNodeFactory.createEmptyNode()
	simpleTypeDescIdentifier := this.SyntaxErrors.createMissingTokenWithDiagnostics(SyntaxKind.IDENTIFIER_TOKEN,
		DiagnosticErrorCode.ERROR_MISSING_TYPE_DESC)
	identifier := this.SyntaxErrors.createMissingTokenWithDiagnostics(SyntaxKind.IDENTIFIER_TOKEN,
		DiagnosticErrorCode.ERROR_MISSING_VARIABLE_NAME)
	simpleNameRef := this.STNodeFactory.createSimpleNameReferenceNode(simpleTypeDescIdentifier)
	semicolon := this.SyntaxErrors.createMissingTokenWithDiagnostics(SyntaxKind.SEMICOLON_TOKEN,
		DiagnosticErrorCode.ERROR_MISSING_SEMICOLON_TOKEN)
	captureBP := this.STNodeFactory.createCaptureBindingPatternNode(identifier)
	typedBindingPattern := this.STNodeFactory.createTypedBindingPatternNode(simpleNameRef, captureBP)
	if isModuleVar {
		varDeclQuals := this.extractVarDeclQualifiers(qualifiers, true)
		typedBindingPattern = this.modifyNodeWithInvalidTokenList(qualifiers, typedBindingPattern)
		if this.isSyntaxKindInList(varDeclQuals, SyntaxKind.CONFIGURABLE_KEYWORD) {
			return this.createConfigurableModuleVarDeclWithMissingInitializer(metadata, publicQualifier, varDeclQuals,
				typedBindingPattern, semicolon)
		}
		varDeclQualNodeList := this.STNodeFactory.createNodeList(varDeclQuals)
		return this.STNodeFactory.createModuleVariableDeclarationNode(metadata, publicQualifier, varDeclQualNodeList,
			typedBindingPattern, emptyNode, emptyNode, semicolon)
	}
	typedBindingPattern = this.modifyNodeWithInvalidTokenList(qualifiers, typedBindingPattern)
	return this.STNodeFactory.createVariableDeclarationNode(metadata, emptyNode, typedBindingPattern, emptyNode,
		emptyNode, semicolon)
}

func (this *BallerinaParser) createMissingWhereClause() internal.STNode {
	whereKeyword := this.SyntaxErrors.createMissingTokenWithDiagnostics(SyntaxKind.WHERE_KEYWORD,
		DiagnosticErrorCode.ERROR_MISSING_WHERE_KEYWORD)
	missingIdentifier := this.SyntaxErrors.createMissingTokenWithDiagnostics(
		SyntaxKind.IDENTIFIER_TOKEN, DiagnosticErrorCode.ERROR_MISSING_EXPRESSION)
	missingExpr := this.STNodeFactory.createSimpleNameReferenceNode(missingIdentifier)
	return this.STNodeFactory.createWhereClauseNode(whereKeyword, missingExpr)
}

func (this *BallerinaParser) createMissingSimpleObjectFieldDefault() internal.STNode {
	metadata := internal.CreateEmptyNode()
	return this.createMissingSimpleObjectField(metadata, nil, false)
}

func (this *BallerinaParser) createMissingSimpleObjectField(metadata internal.STNode, qualifiers []STNode, isObjectTypeDesc bool) internal.STNode {
	emptyNode := this.STNodeFactory.createEmptyNode()
	simpleTypeDescIdentifier := this.SyntaxErrors.createMissingTokenWithDiagnostics(SyntaxKind.IDENTIFIER_TOKEN,
		DiagnosticErrorCode.ERROR_MISSING_TYPE_DESC)
	simpleNameRef := this.STNodeFactory.createSimpleNameReferenceNode(simpleTypeDescIdentifier)
	identifier := this.SyntaxErrors.createMissingTokenWithDiagnostics(SyntaxKind.IDENTIFIER_TOKEN,
		DiagnosticErrorCode.ERROR_MISSING_FIELD_NAME)
	semicolon := this.SyntaxErrors.createMissingTokenWithDiagnostics(SyntaxKind.SEMICOLON_TOKEN,
		DiagnosticErrorCode.ERROR_MISSING_SEMICOLON_TOKEN)
	objectFieldQualifiers := this.extractObjectFieldQualifiers(qualifiers, isObjectTypeDesc)
	objectFieldQualNodeList := this.STNodeFactory.createNodeList(objectFieldQualifiers)
	simpleNameRef = this.modifyNodeWithInvalidTokenList(qualifiers, simpleNameRef)
	if metadata != nil {
		metadata = this.addMetadataNotAttachedDiagnostic(metadata)
	}
	return this.STNodeFactory.createObjectFieldNode(metadata, emptyNode, objectFieldQualNodeList,
		simpleNameRef, identifier, emptyNode, emptyNode, semicolon)
}

func (this *BallerinaParser) createMissingSimpleObjectField() internal.STNode {
	metadata := this.STNodeFactory.createEmptyNode()
	qualifiers := make([]interface{}, 0)
	return this.createMissingSimpleObjectField(metadata, qualifiers, false)
}

func (this *BallerinaParser) modifyNodeWithInvalidTokenList(qualifiers []STNode, node internal.STNode) internal.STNode {
	i := (len(qualifiers) - 1)
	for ; i >= 0; i-- {
		qualifier := this.qualifiers.get(i)
		node = this.SyntaxErrors.cloneWithLeadingInvalidNodeMinutiae(node, qualifier)
	}
	return node
}

func (this *BallerinaParser) modifyTypedBindingPatternWithIsolatedQualifier(typedBindingPattern internal.STNode, isolatedQualifier internal.STNode) internal.STNode {
	typedBindingPatternNode := internal.STTypedBindingPatternNode(typedBindingPattern)
	typeDescriptor := typedBindingPatternNode.typeDescriptor
	bindingPattern := typedBindingPatternNode.bindingPattern
	switch typeDescriptor.kind {
	case OBJECT_TYPE_DESC:
		typeDescriptor = this.modifyObjectTypeDescWithALeadingQualifier(typeDescriptor, isolatedQualifier)
	case FUNCTION_TYPE_DESC:
		typeDescriptorthis.modifyFuncTypeDescWithALeadingQualifier(typeDescriptor, isolatedQualifier)
	default:
		typeDescriptor = this.SyntaxErrors.cloneWithLeadingInvalidNodeMinutiae(typeDescriptor, isolatedQualifier,
			DiagnosticErrorCode.ERROR_QUALIFIER_NOT_ALLOWED, (isolatedQualifier).text())
	}
	return this.STNodeFactory.createTypedBindingPatternNode(typeDescriptor, bindingPattern)
}

func (this *BallerinaParser) modifyObjectTypeDescWithALeadingQualifier(objectTypeDesc internal.STNode, newQualifier internal.STNode) internal.STNode {
	objectTypeDescriptorNode := internal.STObjectTypeDescriptorNode(objectTypeDesc)
	qualifierList := internal.STNodeList(objectTypeDescriptorNode.objectTypeQualifiers)
	newObjectTypeQualifiers := this.modifyNodeListWithALeadingQualifier(qualifierList, newQualifier)
	return this.objectTypeDescriptorNode.modify(newObjectTypeQualifiers, objectTypeDescriptorNode.objectKeyword,
		objectTypeDescriptorNode.openBrace, objectTypeDescriptorNode.members,
		objectTypeDescriptorNode.closeBrace)
}

func (this *BallerinaParser) modifyFuncTypeDescWithALeadingQualifier(funcTypeDesc internal.STNode, newQualifier internal.STNode) internal.STNode {
	funcTypeDescriptorNode := internal.STFunctionTypeDescriptorNode(funcTypeDesc)
	qualifierList := funcTypeDescriptorNode.qualifierList
	newfuncTypeQualifiers := this.modifyNodeListWithALeadingQualifier(qualifierList, newQualifier)
	return this.funcTypeDescriptorNode.modify(newfuncTypeQualifiers, funcTypeDescriptorNode.functionKeyword,
		funcTypeDescriptorNode.functionSignature)
}

func (this *BallerinaParser) modifyNodeListWithALeadingQualifier(qualifiers internal.STNode, newQualifier internal.STNode) internal.STNode {
	newQualifierList := make([]interface{}, 0)
	this.newQualifierList.add(newQualifier)
	qualifierNodeList := internal.STNodeList(qualifiers)
	i := 0
	for ; i < len(qualifierNodeList); i++ {
		qualifier := this.qualifierNodeList.get(i)
		if qualifier.kind == newQualifier.kind {
			this.updateLastNodeInListWithInvalidNode(newQualifierList, qualifier,
				DiagnosticErrorCode.ERROR_DUPLICATE_QUALIFIER, (qualifier).text())
		} else {
			this.newQualifierList.add(qualifier)
		}
	}
	return this.STNodeFactory.createNodeList(newQualifierList)
}

func (this *BallerinaParser) parseAssignmentStmtRhs(lvExpr internal.STNode) internal.STNode {
	assign := this.parseAssignOp()
	expr := this.parseActionOrExpression()
	semicolon := this.parseSemicolon()
	this.endContext()
	if (lvExpr.kind == SyntaxKind.ERROR_CONSTRUCTOR) && this.isPossibleErrorBindingPattern(lvExpr) {
		lvExpr = this.getBindingPattern(lvExpr, false)
	}
	if this.isWildcardBP(lvExpr) {
		lvExpr = this.getWildcardBindingPattern(lvExpr)
	}
	lvExprValid := this.isValidLVExpr(lvExpr)
	if !lvExprValid {
		identifier := this.SyntaxErrors.createMissingToken(SyntaxKind.IDENTIFIER_TOKEN)
		simpleNameRef := this.STNodeFactory.createSimpleNameReferenceNode(identifier)
		lvExpr = this.SyntaxErrors.cloneWithLeadingInvalidNodeMinutiae(simpleNameRef, lvExpr,
			DiagnosticErrorCode.ERROR_INVALID_EXPR_IN_ASSIGNMENT_LHS)
	}
	return this.STNodeFactory.createAssignmentStatementNode(lvExpr, assign, expr, semicolon)
}

func (this *BallerinaParser) parseExpression() internal.STNode {
	return this.parseExpression(DEFAULT_OP_PRECEDENCE, true, false)
}

func (this *BallerinaParser) parseActionOrExpression() internal.STNode {
	return this.parseExpression(DEFAULT_OP_PRECEDENCE, true, true)
}

func (this *BallerinaParser) parseActionOrExpressionInLhs(annots internal.STNode) internal.STNode {
	return this.parseExpression(DEFAULT_OP_PRECEDENCE, annots, false, true, false)
}

func (this *BallerinaParser) parseExpression(isRhsExpr bool) internal.STNode {
	return this.parseExpression(DEFAULT_OP_PRECEDENCE, isRhsExpr, false)
}

func (this *BallerinaParser) isValidLVExpr(expression internal.STNode) bool {
	switch expression.kind {
	case SIMPLE_NAME_REFERENCE,
		QUALIFIED_NAME_REFERENCE,
		LIST_BINDING_PATTERN,
		MAPPING_BINDING_PATTERN,
		ERROR_BINDING_PATTERN,
		WILDCARD_BINDING_PATTERN:
		true
	case FIELD_ACCESS:
		this.isValidLVMemberExpr(STFieldAccessExpressionNode(expression).expression)
	case INDEXED_EXPRESSION:
		this.isValidLVMemberExpr(STIndexedExpressionNode(expression).containerExpression)
	default:
		expression.(internal.STMissingToken)
	}
}

func (this *BallerinaParser) isValidLVMemberExpr(expression internal.STNode) bool {
	switch expression.kind {
	case SIMPLE_NAME_REFERENCE,
		QUALIFIED_NAME_REFERENCE:
		true
	case FIELD_ACCESS:
		this.isValidLVMemberExpr(STFieldAccessExpressionNode(expression).expression)
	case INDEXED_EXPRESSION:
		this.isValidLVMemberExpr(STIndexedExpressionNode(expression).containerExpression)
	case BRACED_EXPRESSION:
		this.isValidLVMemberExpr(STBracedExpressionNode(expression).expression)
	default:
		expression.(internal.STMissingToken)
	}
}

func (this *BallerinaParser) parseExpression(precedenceLevel OperatorPrecedence, isRhsExpr bool, allowActions bool) internal.STNode {
	return this.parseExpression(precedenceLevel, isRhsExpr, allowActions, false)
}

func (this *BallerinaParser) parseExpression(precedenceLevel OperatorPrecedence, isRhsExpr bool, allowActions bool, isInConditionalExpr bool) internal.STNode {
	return this.parseExpression(precedenceLevel, isRhsExpr, allowActions, false, isInConditionalExpr)
}

func (this *BallerinaParser) parseExpression(precedenceLevel OperatorPrecedence, isRhsExpr bool, allowActions bool, isInMatchGuard bool, isInConditionalExpr bool) internal.STNode {
	expr := this.parseTerminalExpression(isRhsExpr, allowActions, isInConditionalExpr)
	return this.parseExpressionRhs(precedenceLevel, expr, isRhsExpr, allowActions, isInMatchGuard, isInConditionalExpr)
}

func (this *BallerinaParser) invalidateActionAndGetMissingExpr(node internal.STNode) internal.STNode {
	identifier := this.SyntaxErrors.createMissingToken(SyntaxKind.IDENTIFIER_TOKEN)
	identifier = this.SyntaxErrors.cloneWithTrailingInvalidNodeMinutiae(identifier, node,
		DiagnosticErrorCode.ERROR_EXPRESSION_EXPECTED_ACTION_FOUND)
	return this.STNodeFactory.createSimpleNameReferenceNode(identifier)
}

func (this *BallerinaParser) parseExpression(precedenceLevel OperatorPrecedence, annots internal.STNode, isRhsExpr bool, allowActions bool, isInConditionalExpr bool) internal.STNode {
	expr := this.parseTerminalExpression(annots, isRhsExpr, allowActions, isInConditionalExpr)
	return this.parseExpressionRhs(precedenceLevel, expr, isRhsExpr, allowActions, false, isInConditionalExpr)
}

func (this *BallerinaParser) parseTerminalExpression(isRhsExpr bool, allowActions bool, isInConditionalExpr bool) internal.STNode {
	annots := this.STNodeFactory.createEmptyNodeList()
	if peek().kind == SyntaxKind.AT_TOKEN {
		annots = this.parseOptionalAnnotations()
	}
	return this.parseTerminalExpression(annots, isRhsExpr, allowActions, isInConditionalExpr)
}

func (this *BallerinaParser) parseTerminalExpression(annots internal.STNode, isRhsExpr bool, allowActions bool, isInConditionalExpr bool) internal.STNode {
	return this.parseTerminalExpression(annots, nil, isRhsExpr, allowActions, isInConditionalExpr)
}

func (this *BallerinaParser) parseTerminalExpression(annots internal.STNode, qualifiers []STNode, isRhsExpr bool, allowActions bool, isInConditionalExpr bool) internal.STNode {
	this.parseExprQualifiers(qualifiers)
	nextToken := this.peek()
	annotNodeList := internal.STNodeList(annots)
	if (!this.annotNodeList.isEmpty()) && (!this.isAnnotAllowedExprStart(nextToken)) {
		annots = this.addAnnotNotAttachedDiagnostic(annotNodeList)
		qualifierNodeList := this.createObjectTypeQualNodeList(qualifiers)
		return this.createMissingObjectConstructor(annots, qualifierNodeList)
	}
	this.validateExprAnnotsAndQualifiers(nextToken, annots, qualifiers)
	if this.isQualifiedIdentifierPredeclaredPrefix(nextToken.kind) {
		return this.parseQualifiedIdentifierOrExpression(isInConditionalExpr, isRhsExpr, allowActions)
	}
	switch nextToken.kind {
	case DECIMAL_INTEGER_LITERAL_TOKEN:
	case HEX_INTEGER_LITERAL_TOKEN:
	case STRING_LITERAL_TOKEN:
	case NULL_KEYWORD:
	case TRUE_KEYWORD:
	case FALSE_KEYWORD:
	case DECIMAL_FLOATING_POINT_LITERAL_TOKEN:
	case HEX_FLOATING_POINT_LITERAL_TOKEN:
		return this.parseBasicLiteral()
	case OPEN_PAREN_TOKEN:
		return this.parseBracedExpression(isRhsExpr, allowActions)
	case CHECK_KEYWORD:
	case CHECKPANIC_KEYWORD:
		return this.parseCheckExpression(isRhsExpr, allowActions, isInConditionalExpr)
	case OPEN_BRACE_TOKEN:
		return this.parseMappingConstructorExpr()
	case TYPEOF_KEYWORD:
		return this.parseTypeofExpression(isRhsExpr, isInConditionalExpr)
	case PLUS_TOKEN:
	case MINUS_TOKEN:
	case NEGATION_TOKEN:
	case EXCLAMATION_MARK_TOKEN:
		return this.parseUnaryExpression(isRhsExpr, isInConditionalExpr)
	case TRAP_KEYWORD:
		return this.parseTrapExpression(isRhsExpr, allowActions, isInConditionalExpr)
	case OPEN_BRACKET_TOKEN:
		return this.parseListConstructorExpr()
	case LT_TOKEN:
		return this.parseTypeCastExpr(isRhsExpr, allowActions, isInConditionalExpr)
	case TABLE_KEYWORD:
	case STREAM_KEYWORD:
	case FROM_KEYWORD:
	case MAP_KEYWORD:
		return this.parseTableConstructorOrQuery(isRhsExpr, allowActions)
	case ERROR_KEYWORD:
		return this.parseErrorConstructorExpr(consume())
	case LET_KEYWORD:
		return this.parseLetExpression(isRhsExpr, isInConditionalExpr)
	case BACKTICK_TOKEN:
		return this.parseTemplateExpression()
	case OBJECT_KEYWORD:
		return this.parseObjectConstructorExpression(annots, qualifiers)
	case XML_KEYWORD:
		return this.parseXMLTemplateExpression()
	case RE_KEYWORD:
		return this.parseRegExpTemplateExpression()
	case STRING_KEYWORD:
		nextNextToken := this.getNextNextToken()
		if nextNextToken.kind == SyntaxKind.BACKTICK_TOKEN {
			return this.parseStringTemplateExpression()
		}
		return this.parseSimpleTypeInTerminalExpr()
	case FUNCTION_KEYWORD:
		return this.parseExplicitFunctionExpression(annots, qualifiers, isRhsExpr)
	case NEW_KEYWORD:
		return this.parseNewExpression()
	case START_KEYWORD:
		return this.parseStartAction(annots)
	case FLUSH_KEYWORD:
		return this.parseFlushAction()
	case LEFT_ARROW_TOKEN:
		return this.parseReceiveAction()
	case WAIT_KEYWORD:
		return this.parseWaitAction()
	case COMMIT_KEYWORD:
		return this.parseCommitAction()
	case TRANSACTIONAL_KEYWORD:
		return this.parseTransactionalExpression()
	case BASE16_KEYWORD:
	case BASE64_KEYWORD:
		return this.parseByteArrayLiteral()
	case TRANSACTION_KEYWORD:
		return this.parseQualifiedIdentWithTransactionPrefix(ParserRuleContext.VARIABLE_REF)
	case IDENTIFIER_TOKEN:
		if this.isNaturalKeyword(nextToken) && (getNextNextToken().kind == OPEN_BRACE_TOKEN) {
			return this.parseNaturalExpression()
		}
		return this.parseQualifiedIdentifier(ParserRuleContext.VARIABLE_REF, isInConditionalExpr)
	case CONST_KEYWORD:
		if this.isNaturalKeyword(getNextNextToken()) {
			return this.parseNaturalExpression()
		}
	default:
		if this.isSimpleTypeInExpression(nextToken.kind) {
			return this.parseSimpleTypeInTerminalExpr()
		}
		this.recover(nextToken, ParserRuleContext.TERMINAL_EXPRESSION)
		return this.parseTerminalExpression(annots, qualifiers, isRhsExpr, allowActions, isInConditionalExpr)
	}
}

func (this *BallerinaParser) parseNaturalExpression() internal.STNode {
	this.startContext(ParserRuleContext.NATURAL_EXPRESSION)
	var optionalConstKeyword internal.STNode
	if peek().kind == SyntaxKind.CONST_KEYWORD {
		optionalConstKeyword = consume()
	} else {
		optionalConstKeyword = this.STNodeFactory.createEmptyNode()
	}
	naturalKeyword := this.parseNaturalKeyword()
	optionalParenthesizedArgList := this.parseOptionalParenthesizedArgList()
	return this.parseNaturalExprBody(optionalConstKeyword, naturalKeyword, optionalParenthesizedArgList)
}

func (this *BallerinaParser) parseNaturalExprBody(optionalConstKeyword internal.STNode, naturalKeyword internal.STNode, optionalParenthesizedArgList internal.STNode) internal.STNode {
	openBrace := this.parseOpenBrace()
	if this.openBrace.isMissing() {
		this.endContext()
		return this.createMissingNaturalExpressionNode(optionalConstKeyword, naturalKeyword,
			optionalParenthesizedArgList)
	}
	this.this.tokenReader.startMode(ParserMode.PROMPT)
	prompt := this.parsePromptContent()
	closeBrace := this.parseCloseBrace()
	if this.this.tokenReader.getCurrentMode() == ParserMode.PROMPT {
		this.this.tokenReader.endMode()
	}
	this.endContext()
	return this.STNodeFactory.createNaturalExpressionNode(optionalConstKeyword, naturalKeyword,
		optionalParenthesizedArgList, openBrace, prompt, closeBrace)
}

func (this *BallerinaParser) createMissingNaturalExpressionNode(optionalConstKeyword internal.STNode, naturalKeyword internal.STNode, optionalParenthesizedArgList internal.STNode) internal.STNode {
	openBrace := this.SyntaxErrors.createMissingToken(OPEN_BRACE_TOKEN)
	closeBrace := this.SyntaxErrors.createMissingToken(SyntaxKind.CLOSE_BRACE_TOKEN)
	prompt := this.STAbstractNodeFactory.createEmptyNodeList()
	naturalExpr := this.STNodeFactory.createNaturalExpressionNode(optionalConstKeyword, naturalKeyword,
		optionalParenthesizedArgList, openBrace, prompt, closeBrace)
	naturalExpr = this.SyntaxErrors.addDiagnostic(naturalExpr, DiagnosticErrorCode.ERROR_MISSING_NATURAL_PROMPT_BLOCK)
	return naturalExpr
}

func (this *BallerinaParser) parseOptionalParenthesizedArgList() internal.STNode {
	if peek().kind == SyntaxKind.OPEN_PAREN_TOKEN {
		return this.parseParenthesizedArgList()
	}
	return this.STNodeFactory.createEmptyNode()
}

func (this *BallerinaParser) parsePromptContent() internal.STNode {
	items := make([]interface{}, 0)
	nextToken := this.peek()
	for !this.isEndOfPromptContent(nextToken.kind) {
		contentItem := this.parsePromptItem()
		this.items.add(contentItem)
		nextToken = this.peek()
	}
	return this.STNodeFactory.createNodeList(items)
}

func (this *BallerinaParser) isEndOfPromptContent(kind SyntaxKind) bool {
	switch kind {
	case EOF_TOKEN, CLOSE_BRACE_TOKEN:
		return true
	default:
		return false
	}
}

func (this *BallerinaParser) parsePromptItem() internal.STNode {
	nextToken := this.peek()
	if nextToken.kind == SyntaxKind.INTERPOLATION_START_TOKEN {
		return this.parseInterpolation()
	}
	if nextToken.kind != SyntaxKind.PROMPT_CONTENT {
		nextToken = this.consume()
		return this.STNodeFactory.createLiteralValueToken(SyntaxKind.PROMPT_CONTENT,
			nextToken.text(), nextToken.leadingMinutiae(), nextToken.trailingMinutiae(),
			nextToken.diagnostics())
	}
	return this.consume()
}

func (this *BallerinaParser) createMissingObjectConstructor(annots internal.STNode, qualifierNodeList internal.STNode) internal.STNode {
	objectKeyword := this.SyntaxErrors.createMissingToken(SyntaxKind.OBJECT_KEYWORD)
	openBrace := this.SyntaxErrors.createMissingToken(OPEN_BRACE_TOKEN)
	closeBrace := this.SyntaxErrors.createMissingToken(SyntaxKind.CLOSE_BRACE_TOKEN)
	objConstructor := this.STNodeFactory.createObjectConstructorExpressionNode(annots, qualifierNodeList,
		objectKeyword, STNodeFactory.createEmptyNode(), openBrace, STNodeFactory.createEmptyNodeList(),
		closeBrace)
	objConstructor = this.SyntaxErrors.addDiagnostic(objConstructor,
		DiagnosticErrorCode.ERROR_MISSING_OBJECT_CONSTRUCTOR_EXPRESSION)
	return objConstructor
}

func (this *BallerinaParser) parseQualifiedIdentifierOrExpression(isInConditionalExpr bool, isRhsExpr bool, allowActions bool) internal.STNode {
	preDeclaredPrefix := this.consume()
	nextNextToken := this.getNextNextToken()
	if (nextNextToken.kind == SyntaxKind.IDENTIFIER_TOKEN) && (!this.isKeyKeyword(nextNextToken)) {
		return this.parseQualifiedIdentifierWithPredeclPrefix(preDeclaredPrefix, isInConditionalExpr)
	}
	var context ParserRuleContext
	switch preDeclaredPrefix.kind {
	case TABLE_KEYWORD:
		context = ParserRuleContext.TABLE_CONS_OR_QUERY_EXPR_OR_VAR_REF
		break
	case STREAM_KEYWORD:
		context = ParserRuleContext.QUERY_EXPR_OR_VAR_REF
		break
	case ERROR_KEYWORD:
		context = ParserRuleContext.ERROR_CONS_EXPR_OR_VAR_REF
		break
	default:
		return this.parseQualifiedIdentifierWithPredeclPrefix(preDeclaredPrefix, isInConditionalExpr)
	}
	solution := this.recover(peek(), context)
	if solution.action == Action.KEEP {
		return this.parseQualifiedIdentifierWithPredeclPrefix(preDeclaredPrefix, isInConditionalExpr)
	}
	if preDeclaredPrefix.kind == SyntaxKind.ERROR_KEYWORD {
		return this.parseErrorConstructorExpr(preDeclaredPrefix)
	}
	this.startContext(ParserRuleContext.TABLE_CONSTRUCTOR_OR_QUERY_EXPRESSION)
	var tableOrQuery internal.STNode
	if preDeclaredPrefix.kind == SyntaxKind.STREAM_KEYWORD {
		queryConstructType := this.parseQueryConstructType(preDeclaredPrefix, null)
		tableOrQuery = this.parseQueryExprRhs(queryConstructType, isRhsExpr, allowActions)
	} else {
		tableOrQuery = this.parseTableConstructorOrQuery(preDeclaredPrefix, isRhsExpr, allowActions)
	}
	this.endContext()
	return tableOrQuery
}

func (this *BallerinaParser) validateExprAnnotsAndQualifiers(nextToken internal.STToken, annots internal.STNode, qualifiers []STNode) {
	switch nextToken.kind {
	case START_KEYWORD:
		this.reportInvalidQualifierList(qualifiers)
		break
	case FUNCTION_KEYWORD:
	case OBJECT_KEYWORD:
	case AT_TOKEN:
		break
	default:
		if this.isValidExprStart(nextToken.kind) {
			this.reportInvalidExpressionAnnots(annots, qualifiers)
			this.reportInvalidQualifierList(qualifiers)
		}
	}
}

func (this *BallerinaParser) isAnnotAllowedExprStart(nextToken internal.STToken) bool {
	switch nextToken.kind {
	case START_KEYWORD, FUNCTION_KEYWORD, OBJECT_KEYWORD:
		return true
	default:
		return false
	}
}

func (this *BallerinaParser) isValidExprStart(tokenKind SyntaxKind) bool {
	switch tokenKind {
	case DECIMAL_INTEGER_LITERAL_TOKEN,
		HEX_INTEGER_LITERAL_TOKEN,
		STRING_LITERAL_TOKEN,
		NULL_KEYWORD,
		TRUE_KEYWORD,
		FALSE_KEYWORD,
		DECIMAL_FLOATING_POINT_LITERAL_TOKEN,
		HEX_FLOATING_POINT_LITERAL_TOKEN,
		IDENTIFIER_TOKEN,
		OPEN_PAREN_TOKEN,
		CHECK_KEYWORD,
		CHECKPANIC_KEYWORD,
		OPEN_BRACE_TOKEN,
		TYPEOF_KEYWORD,
		PLUS_TOKEN,
		MINUS_TOKEN,
		NEGATION_TOKEN,
		EXCLAMATION_MARK_TOKEN,
		TRAP_KEYWORD,
		OPEN_BRACKET_TOKEN,
		LT_TOKEN,
		TABLE_KEYWORD,
		STREAM_KEYWORD,
		FROM_KEYWORD,
		ERROR_KEYWORD,
		LET_KEYWORD,
		BACKTICK_TOKEN,
		XML_KEYWORD,
		RE_KEYWORD,
		STRING_KEYWORD,
		FUNCTION_KEYWORD,
		AT_TOKEN,
		NEW_KEYWORD,
		START_KEYWORD,
		FLUSH_KEYWORD,
		LEFT_ARROW_TOKEN,
		WAIT_KEYWORD,
		COMMIT_KEYWORD,
		SERVICE_KEYWORD,
		BASE16_KEYWORD,
		BASE64_KEYWORD,
		ISOLATED_KEYWORD,
		TRANSACTIONAL_KEYWORD,
		CLIENT_KEYWORD,
		NATURAL_KEYWORD,
		OBJECT_KEYWORD:
		return true
	default:
		if this.isPredeclaredPrefix(tokenKind) {
			return true
		}
		return this.isSimpleTypeInExpression(tokenKind)
	}
}

func (this *BallerinaParser) parseNewExpression() internal.STNode {
	newKeyword := this.parseNewKeyword()
	return this.parseNewKeywordRhs(newKeyword)
}

func (this *BallerinaParser) parseNewKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.NEW_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.NEW_KEYWORD)
		return this.parseNewKeyword()
	}
}

func (this *BallerinaParser) parseNewKeywordRhs(newKeyword internal.STNode) internal.STNode {
	nextToken := this.peek()
	if nextToken.kind == SyntaxKind.OPEN_PAREN_TOKEN {
		return this.parseImplicitNewExpr(newKeyword)
	}
	if this.isClassDescriptorStartToken(nextToken.kind) {
		return this.parseExplicitNewExpr(newKeyword)
	}
	return this.createImplicitNewExpr(newKeyword, STNodeFactory.createEmptyNode())
}

func (this *BallerinaParser) isClassDescriptorStartToken(tokenKind SyntaxKind) bool {
	return ((tokenKind == SyntaxKind.STREAM_KEYWORD) || this.isPredeclaredIdentifier(tokenKind))
}

func (this *BallerinaParser) parseExplicitNewExpr(newKeyword internal.STNode) internal.STNode {
	typeDescriptor := this.parseClassDescriptor()
	parenthesizedArgsList := this.parseParenthesizedArgList()
	return this.STNodeFactory.createExplicitNewExpressionNode(newKeyword, typeDescriptor, parenthesizedArgsList)
}

func (this *BallerinaParser) parseClassDescriptor() internal.STNode {
	this.startContext(ParserRuleContext.CLASS_DESCRIPTOR_IN_NEW_EXPR)
	var classDescriptor internal.STNode
	nextToken := this.peek()
	switch nextToken.kind {
	case STREAM_KEYWORD:
		classDescriptor = this.parseStreamTypeDescriptor(consume())
		break
	default:
		if this.isPredeclaredIdentifier(nextToken.kind) {
			classDescriptor = this.parseTypeReference()
			break
		}
		this.recover(nextToken, ParserRuleContext.CLASS_DESCRIPTOR)
		return this.parseClassDescriptor()
	}
	this.endContext()
	return classDescriptor
}

func (this *BallerinaParser) parseImplicitNewExpr(newKeyword internal.STNode) internal.STNode {
	parenthesizedArgList := this.parseParenthesizedArgList()
	return this.createImplicitNewExpr(newKeyword, parenthesizedArgList)
}

func (this *BallerinaParser) createImplicitNewExpr(newKeyword internal.STNode, parenthesizedArgList internal.STNode) internal.STNode {
	return this.STNodeFactory.createImplicitNewExpressionNode(newKeyword, parenthesizedArgList)
}

func (this *BallerinaParser) parseParenthesizedArgList() internal.STNode {
	openParan := this.parseArgListOpenParenthesis()
	arguments := this.parseArgsList()
	closeParan := this.parseArgListCloseParenthesis()
	return this.STNodeFactory.createParenthesizedArgList(openParan, arguments, closeParan)
}

func (this *BallerinaParser) parseExpressionRhs(precedenceLevel OperatorPrecedence, lhsExpr internal.STNode, isRhsExpr bool, allowActions bool) internal.STNode {
	return this.parseExpressionRhs(precedenceLevel, lhsExpr, isRhsExpr, allowActions, false, false)
}

func (this *BallerinaParser) parseExpressionRhs(currentPrecedenceLevel OperatorPrecedence, lhsExpr internal.STNode, isRhsExpr bool, allowActions bool, isInMatchGuard bool, isInConditionalExpr bool) internal.STNode {
	actionOrExpression := this.parseExpressionRhsInternal(currentPrecedenceLevel, lhsExpr, isRhsExpr, allowActions,
		isInMatchGuard, isInConditionalExpr)
	if ((!allowActions) && this.isAction(actionOrExpression)) && (actionOrExpression.kind != SyntaxKind.BRACED_ACTION) {
		actionOrExpression = this.invalidateActionAndGetMissingExpr(actionOrExpression)
	}
	return actionOrExpression
}

func (this *BallerinaParser) parseExpressionRhsInternal(currentPrecedenceLevel OperatorPrecedence, lhsExpr internal.STNode, isRhsExpr bool, allowActions bool, isInMatchGuard bool, isInConditionalExpr bool) internal.STNode {
	nextToken := this.peek()
	if this.isAction(lhsExpr) || this.isEndOfActionOrExpression(nextToken, isRhsExpr, isInMatchGuard) {
		return lhsExpr
	}
	nextTokenKind := nextToken.kind
	if !this.isValidExprRhsStart(nextTokenKind, lhsExpr.kind) {
		return this.recoverExpressionRhs(currentPrecedenceLevel, lhsExpr, isRhsExpr, allowActions, isInMatchGuard,
			isInConditionalExpr)
	}
	if (nextTokenKind == SyntaxKind.GT_TOKEN) && (peek(2).kind == SyntaxKind.GT_TOKEN) {
		if peek(3).kind == SyntaxKind.GT_TOKEN {
			nextTokenKind = SyntaxKind.TRIPPLE_GT_TOKEN
		} else {
			nextTokenKind = SyntaxKind.DOUBLE_GT_TOKEN
		}
	}
	nextOperatorPrecedence := this.getOpPrecedence(nextTokenKind)
	if this.currentPrecedenceLevel.isHigherThanOrEqual(nextOperatorPrecedence, allowActions) {
		return lhsExpr
	}
	var newLhsExpr internal.STNode
	var operator internal.STNode
	switch nextTokenKind {
	case OPEN_PAREN_TOKEN:
		newLhsExpr = this.parseFuncCallOrNaturalExpr(lhsExpr)
		break
	case OPEN_BRACKET_TOKEN:
		newLhsExpr = this.parseMemberAccessExpr(lhsExpr, isRhsExpr)
		break
	case DOT_TOKEN:
		newLhsExpr = this.parseFieldAccessOrMethodCall(lhsExpr, isInConditionalExpr)
		break
	case IS_KEYWORD:
	case NOT_IS_KEYWORD:
		newLhsExpr = this.parseTypeTestExpression(lhsExpr, isInConditionalExpr)
		break
	case RIGHT_ARROW_TOKEN:
		newLhsExpr = this.parseRemoteMethodCallOrClientResourceAccessOrAsyncSendAction(lhsExpr, isRhsExpr,
			isInMatchGuard)
		break
	case SYNC_SEND_TOKEN:
		newLhsExpr = this.parseSyncSendAction(lhsExpr)
		break
	case RIGHT_DOUBLE_ARROW_TOKEN:
		newLhsExpr = this.parseImplicitAnonFunc(lhsExpr, isRhsExpr)
		break
	case ANNOT_CHAINING_TOKEN:
		newLhsExpr = this.parseAnnotAccessExpression(lhsExpr, isInConditionalExpr)
		break
	case OPTIONAL_CHAINING_TOKEN:
		newLhsExpr = this.parseOptionalFieldAccessExpression(lhsExpr, isInConditionalExpr)
		break
	case QUESTION_MARK_TOKEN:
		newLhsExpr = this.parseConditionalExpression(lhsExpr, isInConditionalExpr)
		break
	case DOT_LT_TOKEN:
		newLhsExpr = this.parseXMLFilterExpression(lhsExpr)
		break
	case SLASH_LT_TOKEN:
	case DOUBLE_SLASH_DOUBLE_ASTERISK_LT_TOKEN:
	case SLASH_ASTERISK_TOKEN:
		newLhsExpr = this.parseXMLStepExpression(lhsExpr)
		break
	default:
		if (nextTokenKind == SyntaxKind.SLASH_TOKEN) && (peek(2).kind == SyntaxKind.LT_TOKEN) {
			expectedNodeType := this.getExpectedNodeKind(3)
			if expectedNodeType == SyntaxKind.XML_STEP_EXPRESSION {
				newLhsExpr = this.createXMLStepExpression(lhsExpr)
				break
			}
		}
		if nextTokenKind == SyntaxKind.DOUBLE_GT_TOKEN {
			operator = this.parseSignedRightShiftToken()
		} else if nextTokenKind == SyntaxKind.TRIPPLE_GT_TOKEN {
			operator = this.parseUnsignedRightShiftToken()
		}
		rhsExpr := this.parseExpression(nextOperatorPrecedence, isRhsExpr, false, isInConditionalExpr)
		newLhsExpr = this.STNodeFactory.createBinaryExpressionNode(SyntaxKind.BINARY_EXPRESSION, lhsExpr, operator,
			rhsExpr)
		break
	}
	return this.parseExpressionRhsInternal(currentPrecedenceLevel, newLhsExpr, isRhsExpr, allowActions, isInMatchGuard,
		isInConditionalExpr)
}

func (this *BallerinaParser) recoverExpressionRhs(currentPrecedenceLevel OperatorPrecedence, lhsExpr internal.STNode, isRhsExpr bool, allowActions bool, isInMatchGuard bool, isInConditionalExpr bool) internal.STNode {
	token := this.peek()
	lhsExprKind := lhsExpr.kind
	var solution Solution
	if (lhsExprKind == SyntaxKind.QUALIFIED_NAME_REFERENCE) || (lhsExprKind == SyntaxKind.SIMPLE_NAME_REFERENCE) {
		solution = this.recover(token, ParserRuleContext.VARIABLE_REF_RHS)
	} else {
		solution = this.recover(token, ParserRuleContext.EXPRESSION_RHS)
	}
	if solution.action == Action.REMOVE {
		return this.parseExpressionRhs(currentPrecedenceLevel, lhsExpr, isRhsExpr, allowActions, isInMatchGuard,
			isInConditionalExpr)
	}
	if solution.ctx == ParserRuleContext.BINARY_OPERATOR {
		binaryOpKind := this.getBinaryOperatorKindToInsert(currentPrecedenceLevel)
		binaryOpContext := this.getMissingBinaryOperatorContext(currentPrecedenceLevel)
		this.insertToken(binaryOpKind, binaryOpContext)
	}
	return this.parseExpressionRhsInternal(currentPrecedenceLevel, lhsExpr, isRhsExpr, allowActions, isInMatchGuard,
		isInConditionalExpr)
}

func (this *BallerinaParser) createXMLStepExpression(lhsExpr internal.STNode) internal.STNode {
	var newLhsExpr internal.STNode
	slashToken := this.parseSlashToken()
	ltToken := this.parseLTToken()
	var slashLT internal.STNode
	if this.hasTrailingMinutiae(slashToken) || this.hasLeadingMinutiae(ltToken) {
		diagnostics := make([]interface{}, 0)
		this.diagnostics.add(SyntaxErrors.createDiagnostic(DiagnosticErrorCode.ERROR_INVALID_WHITESPACE_IN_SLASH_LT_TOKEN))
		slashLT = this.STNodeFactory.createMissingToken(SyntaxKind.SLASH_LT_TOKEN, diagnostics)
		slashLT = this.SyntaxErrors.cloneWithLeadingInvalidNodeMinutiae(slashLT, slashToken)
		slashLT = this.SyntaxErrors.cloneWithLeadingInvalidNodeMinutiae(slashLT, ltToken)
	} else {
		slashLT = this.STNodeFactory.createToken(SyntaxKind.SLASH_LT_TOKEN, slashToken.leadingMinutiae(),
			ltToken.trailingMinutiae())
	}
	namePattern := this.parseXMLNamePatternChain(slashLT)
	xmlStepExtends := this.parseXMLStepExtends()
	newLhsExpr = this.STNodeFactory.createXMLStepExpressionNode(lhsExpr, namePattern, xmlStepExtends)
	return newLhsExpr
}

func (this *BallerinaParser) getExpectedNodeKind(lookahead int) SyntaxKind {
	nextToken := this.peek(lookahead)
	switch nextToken.kind {
	case ASTERISK_TOKEN:
		return SyntaxKind.XML_STEP_EXPRESSION
	case GT_TOKEN:
		break
	case PIPE_TOKEN:
		return this.getExpectedNodeKind(lookahead + 1)
	case IDENTIFIER_TOKEN:
		nextToken = this.peek(lookahead + 1)
		switch nextToken.kind {
		case GT_TOKEN:
			break
		case PIPE_TOKEN:
			return this.getExpectedNodeKind(lookahead + 1)
		case COLON_TOKEN:
			nextToken = this.peek(lookahead + 1)
			switch nextToken.kind {
			case ASTERISK_TOKEN:
			case GT_TOKEN:
				return SyntaxKind.XML_STEP_EXPRESSION
			case IDENTIFIER_TOKEN:
				nextToken = this.peek(lookahead + 1)
				if nextToken.kind == SyntaxKind.PIPE_TOKEN {
					return this.getExpectedNodeKind(lookahead + 1)
				}
				break
			default:
				return SyntaxKind.TYPE_CAST_EXPRESSION
			}
			break
		default:
			return SyntaxKind.TYPE_CAST_EXPRESSION
		}
		break
	default:
		return SyntaxKind.TYPE_CAST_EXPRESSION
	}
	nextToken = this.peek(lookahead + 1)
	switch nextToken.kind {
	case OPEN_BRACKET_TOKEN:
	case OPEN_BRACE_TOKEN:
	case PLUS_TOKEN:
	case MINUS_TOKEN:
	case FROM_KEYWORD:
	case LET_KEYWORD:
		return SyntaxKind.XML_STEP_EXPRESSION
	default:
		if this.isValidExpressionStart(nextToken.kind, lookahead) {
			break
		}
		return SyntaxKind.XML_STEP_EXPRESSION
	}
	return SyntaxKind.TYPE_CAST_EXPRESSION
}

func (this *BallerinaParser) hasTrailingMinutiae(node internal.STNode) bool {
	return (this.node.widthWithTrailingMinutiae() > this.node.width())
}

func (this *BallerinaParser) hasLeadingMinutiae(node internal.STNode) bool {
	return (this.node.widthWithLeadingMinutiae() > this.node.width())
}

func (this *BallerinaParser) isValidExprRhsStart(tokenKind SyntaxKind, precedingNodeKind SyntaxKind) bool {
	switch tokenKind {
	case OPEN_PAREN_TOKEN:
		return ((precedingNodeKind == SyntaxKind.QUALIFIED_NAME_REFERENCE) || (precedingNodeKind == SyntaxKind.SIMPLE_NAME_REFERENCE))
	case DOT_TOKEN,
		OPEN_BRACKET_TOKEN,
		IS_KEYWORD,
		RIGHT_ARROW_TOKEN,
		RIGHT_DOUBLE_ARROW_TOKEN,
		SYNC_SEND_TOKEN,
		ANNOT_CHAINING_TOKEN,
		OPTIONAL_CHAINING_TOKEN,
		COLON_TOKEN,
		DOT_LT_TOKEN,
		SLASH_LT_TOKEN,
		DOUBLE_SLASH_DOUBLE_ASTERISK_LT_TOKEN,
		SLASH_ASTERISK_TOKEN,
		NOT_IS_KEYWORD:
		return true
	case QUESTION_MARK_TOKEN:
		return ((getNextNextToken().kind != SyntaxKind.EQUAL_TOKEN) && (peek(3).kind != SyntaxKind.EQUAL_TOKEN))
	default:
		return this.isBinaryOperator(tokenKind)
	}
}

func (this *BallerinaParser) parseMemberAccessExpr(lhsExpr internal.STNode, isRhsExpr bool) internal.STNode {
	this.startContext(ParserRuleContext.MEMBER_ACCESS_KEY_EXPR)
	openBracket := this.parseOpenBracket()
	keyExpr := this.parseMemberAccessKeyExprs(isRhsExpr)
	closeBracket := this.parseCloseBracket()
	this.endContext()
	if isRhsExpr && this.(keyExpr).isEmpty() {
		missingVarRef := this.STNodeFactory.createSimpleNameReferenceNode(SyntaxErrors.createMissingToken(SyntaxKind.IDENTIFIER_TOKEN))
		keyExpr = this.STNodeFactory.createNodeList(missingVarRef)
		closeBracket = this.SyntaxErrors.addDiagnostic(closeBracket,
			DiagnosticErrorCode.ERROR_MISSING_KEY_EXPR_IN_MEMBER_ACCESS_EXPR)
	}
	return this.STNodeFactory.createIndexedExpressionNode(lhsExpr, openBracket, keyExpr, closeBracket)
}

func (this *BallerinaParser) parseMemberAccessKeyExprs(isRhsExpr bool) internal.STNode {
	exprList := make([]interface{}, 0)
	var keyExpr internal.STNode
	var keyExprEnd internal.STNode
	for !this.isEndOfTypeList(peek().kind) {
		keyExpr = this.parseKeyExpr(isRhsExpr)
		this.exprList.add(keyExpr)
		keyExprEnd = this.parseMemberAccessKeyExprEnd()
		if keyExprEnd == nil {
			break
		}
		this.exprList.add(keyExprEnd)
	}
	return this.STNodeFactory.createNodeList(exprList)
}

func (this *BallerinaParser) parseKeyExpr(isRhsExpr bool) internal.STNode {
	if (!isRhsExpr) && (peek().kind == SyntaxKind.ASTERISK_TOKEN) {
		return this.STNodeFactory.createBasicLiteralNode(SyntaxKind.ASTERISK_LITERAL, consume())
	}
	return this.parseExpression(isRhsExpr)
}

func (this *BallerinaParser) parseMemberAccessKeyExprEnd() internal.STNode {
	switch peek().kind {
	case COMMA_TOKEN:
		return this.parseComma()
	case CLOSE_BRACKET_TOKEN:
		return nil
	default:
		this.recover(peek(), ParserRuleContext.MEMBER_ACCESS_KEY_EXPR_END)
		return this.parseMemberAccessKeyExprEnd()
	}
}

func (this *BallerinaParser) parseCloseBracket() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.CLOSE_BRACKET_TOKEN {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.CLOSE_BRACKET)
		return this.parseCloseBracket()
	}
}

func (this *BallerinaParser) parseFieldAccessOrMethodCall(lhsExpr internal.STNode, isInConditionalExpr bool) internal.STNode {
	dotToken := this.parseDotToken()
	if this.isSpecialMethodName(peek()) {
		methodName := this.getKeywordAsSimpleNameRef()
		openParen := this.parseArgListOpenParenthesis()
		args := this.parseArgsList()
		closeParen := this.parseArgListCloseParenthesis()
		return this.STNodeFactory.createMethodCallExpressionNode(lhsExpr, dotToken, methodName, openParen, args,
			closeParen)
	}
	fieldOrMethodName := this.parseFieldAccessIdentifier(isInConditionalExpr)
	if fieldOrMethodName.kind == SyntaxKind.QUALIFIED_NAME_REFERENCE {
		return this.STNodeFactory.createFieldAccessExpressionNode(lhsExpr, dotToken, fieldOrMethodName)
	}
	nextToken := this.peek()
	if nextToken.kind == SyntaxKind.OPEN_PAREN_TOKEN {
		openParen := this.parseArgListOpenParenthesis()
		args := this.parseArgsList()
		closeParen := this.parseArgListCloseParenthesis()
		return this.STNodeFactory.createMethodCallExpressionNode(lhsExpr, dotToken, fieldOrMethodName, openParen, args,
			closeParen)
	}
	return this.STNodeFactory.createFieldAccessExpressionNode(lhsExpr, dotToken, fieldOrMethodName)
}

func (this *BallerinaParser) getKeywordAsSimpleNameRef() internal.STNode {
	mapKeyword := this.consume()
	methodName := this.STNodeFactory.createIdentifierToken(mapKeyword.text(), mapKeyword.leadingMinutiae(),
		mapKeyword.trailingMinutiae(), mapKeyword.diagnostics())
	methodName = this.STNodeFactory.createSimpleNameReferenceNode(methodName)
	return methodName
}

func (this *BallerinaParser) parseBracedExpression(isRhsExpr bool, allowActions bool) internal.STNode {
	openParen := this.parseOpenParenthesis()
	if peek().kind == SyntaxKind.CLOSE_PAREN_TOKEN {
		return this.STNodeFactory.createNilLiteralNode(openParen, consume())
	}
	this.startContext(ParserRuleContext.BRACED_EXPR_OR_ANON_FUNC_PARAMS)
	var expr internal.STNode
	if allowActions {
		expr = this.parseExpression(DEFAULT_OP_PRECEDENCE, isRhsExpr, true)
	} else {
		expr = this.parseExpression(isRhsExpr)
	}
	return this.parseBracedExprOrAnonFuncParamRhs(openParen, expr, isRhsExpr)
}

func (this *BallerinaParser) parseBracedExprOrAnonFuncParamRhs(openParen internal.STNode, expr internal.STNode, isRhsExpr bool) internal.STNode {
	nextToken := this.peek()
	if expr.kind == SyntaxKind.SIMPLE_NAME_REFERENCE {
		switch nextToken.kind {
		case CLOSE_PAREN_TOKEN:
			break
		case COMMA_TOKEN:
			return this.parseImplicitAnonFunc(openParen, expr, isRhsExpr)
		default:
			this.recover(nextToken, ParserRuleContext.BRACED_EXPR_OR_ANON_FUNC_PARAM_RHS)
			return this.parseBracedExprOrAnonFuncParamRhs(openParen, expr, isRhsExpr)
		}
	}
	closeParen := this.parseCloseParenthesis()
	this.endContext()
	if this.isAction(expr) {
		return this.STNodeFactory.createBracedExpressionNode(SyntaxKind.BRACED_ACTION, openParen, expr, closeParen)
	}
	return this.STNodeFactory.createBracedExpressionNode(SyntaxKind.BRACED_EXPRESSION, openParen, expr, closeParen)
}

func (this *BallerinaParser) isAction(node internal.STNode) bool {
	switch node.kind {
	case REMOTE_METHOD_CALL_ACTION,
		BRACED_ACTION,
		CHECK_ACTION,
		START_ACTION,
		TRAP_ACTION,
		FLUSH_ACTION,
		ASYNC_SEND_ACTION,
		SYNC_SEND_ACTION,
		RECEIVE_ACTION,
		WAIT_ACTION,
		QUERY_ACTION,
		COMMIT_ACTION,
		CLIENT_RESOURCE_ACCESS_ACTION:
		return true
	default:
		return false
	}
}

func (this *BallerinaParser) isEndOfActionOrExpression(nextToken internal.STToken, isRhsExpr bool, isInMatchGuard bool) bool {
	tokenKind := nextToken.kind
	if !isRhsExpr {
		if this.isCompoundAssignment(tokenKind) {
			return true
		}
		if isInMatchGuard && (tokenKind == SyntaxKind.RIGHT_DOUBLE_ARROW_TOKEN) {
			return true
		}
	}
	switch tokenKind {
	case EOF_TOKEN,
		CLOSE_BRACE_TOKEN,
		OPEN_BRACE_TOKEN,
		CLOSE_PAREN_TOKEN,
		CLOSE_BRACKET_TOKEN,
		SEMICOLON_TOKEN,
		COMMA_TOKEN,
		PUBLIC_KEYWORD,
		CONST_KEYWORD,
		LISTENER_KEYWORD,
		RESOURCE_KEYWORD,
		EQUAL_TOKEN,
		DOCUMENTATION_STRING,
		AT_TOKEN,
		AS_KEYWORD,
		IN_KEYWORD,
		FROM_KEYWORD,
		WHERE_KEYWORD,
		LET_KEYWORD,
		SELECT_KEYWORD,
		DO_KEYWORD,
		COLON_TOKEN,
		ON_KEYWORD,
		CONFLICT_KEYWORD,
		LIMIT_KEYWORD,
		JOIN_KEYWORD,
		OUTER_KEYWORD,
		ORDER_KEYWORD,
		BY_KEYWORD,
		ASCENDING_KEYWORD,
		DESCENDING_KEYWORD,
		EQUALS_KEYWORD,
		TYPE_KEYWORD:
		return true
	case RIGHT_DOUBLE_ARROW_TOKEN:
		returnisInMatchGuard
	case IDENTIFIER_TOKEN:
		return this.isGroupOrCollectKeyword(nextToken)
	default:
		return this.isSimpleType(tokenKind)
	}
}

func (this *BallerinaParser) parseBasicLiteral() internal.STNode {
	literalToken := this.consume()
	return this.parseBasicLiteral(literalToken)
}

func (this *BallerinaParser) parseBasicLiteral(literalToken internal.STNode) internal.STNode {
	var nodeKind SyntaxKind
	switch literalToken.kind {
	case NULL_KEYWORD:
		nodeKind = SyntaxKind.NULL_LITERAL
	case TRUE_KEYWORD,
		FALSE_KEYWORD:
		nodeKind = SyntaxKind.BOOLEAN_LITERAL
	case DECIMAL_INTEGER_LITERAL_TOKEN,
		DECIMAL_FLOATING_POINT_LITERAL_TOKEN,
		HEX_INTEGER_LITERAL_TOKEN,
		HEX_FLOATING_POINT_LITERAL_TOKEN:
		nodeKind = SyntaxKind.NUMERIC_LITERAL
	case STRING_LITERAL_TOKEN:
		nodeKind = SyntaxKind.STRING_LITERAL
	case ASTERISK_TOKEN:
		nodeKind = SyntaxKind.ASTERISK_LITERAL
	default:
		nodeKind = literalToken.kind
	}
	return this.STNodeFactory.createBasicLiteralNode(nodeKind, literalToken)
}

func (this *BallerinaParser) parseFuncCallOrNaturalExpr(identifier internal.STNode) internal.STNode {
	openParen := this.parseArgListOpenParenthesis()
	args := this.parseArgsList()
	closeParen := this.parseArgListCloseParenthesis()
	if (peek().kind == SyntaxKind.OPEN_BRACE_TOKEN) && this.isNaturalKeyword(identifier) {
		return this.parseNaturalExpression(identifier, openParen, args, closeParen)
	}
	return this.STNodeFactory.createFunctionCallExpressionNode(identifier, openParen, args, closeParen)
}

func (this *BallerinaParser) parseNaturalExpression(nameRef internal.STSimpleNameReferenceNode, openParen internal.STNode, args internal.STNode, closeParen internal.STNode) internal.STNode {
	this.startContext(ParserRuleContext.NATURAL_EXPRESSION)
	optionalConstKeyword := this.STNodeFactory.createEmptyNode()
	naturalKeyword := this.getNaturalKeyword(nameRef.name)
	parenthesizedArgList := this.STNodeFactory.createParenthesizedArgList(openParen, args, closeParen)
	return this.parseNaturalExprBody(optionalConstKeyword, naturalKeyword, parenthesizedArgList)
}

func (this *BallerinaParser) parseErrorBindingPatternOrErrorConstructor() internal.STNode {
	return this.parseErrorConstructorExpr(true)
}

func (this *BallerinaParser) parseErrorConstructorExpr(errorKeyword internal.STNode) internal.STNode {
	this.startContext(ParserRuleContext.ERROR_CONSTRUCTOR)
	return this.parseErrorConstructorExpr(errorKeyword, false)
}

func (this *BallerinaParser) parseErrorConstructorExpr(isAmbiguous bool) internal.STNode {
	this.startContext(ParserRuleContext.ERROR_CONSTRUCTOR)
	errorKeyword := this.parseErrorKeyword()
	return this.parseErrorConstructorExpr(errorKeyword, isAmbiguous)
}

func (this *BallerinaParser) parseErrorConstructorExpr(errorKeyword internal.STNode, isAmbiguous bool) internal.STNode {
	typeReference := this.parseErrorTypeReference()
	openParen := this.parseArgListOpenParenthesis()
	functionArgs := this.parseArgsList()
	var errorArgs internal.STNode
	if isAmbiguous {
		errorArgs = functionArgs
	} else {
		errorArgs = this.getErrorArgList(functionArgs)
	}
	closeParen := this.parseArgListCloseParenthesis()
	this.endContext()
	openParen = this.cloneWithDiagnosticIfListEmpty(errorArgs, openParen,
		DiagnosticErrorCode.ERROR_MISSING_ARG_WITHIN_PARENTHESIS)
	return this.STNodeFactory.createErrorConstructorExpressionNode(errorKeyword, typeReference, openParen, errorArgs,
		closeParen)
}

func (this *BallerinaParser) parseErrorTypeReference() internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case OPEN_PAREN_TOKEN:
		return this.STNodeFactory.createEmptyNode()
	default:
		if this.isPredeclaredIdentifier(nextToken.kind) {
			return this.parseTypeReference()
		}
		this.recover(nextToken, ParserRuleContext.ERROR_CONSTRUCTOR_RHS)
		return this.parseErrorTypeReference()
	}
}

func (this *BallerinaParser) getErrorArgList(functionArgs internal.STNode) internal.STNode {
	argList := internal.STNodeList(functionArgs)
	if this.argList.isEmpty() {
		return argList
	}
	errorArgList := make([]interface{}, 0)
	arg := this.argList.get(0)
	switch arg.kind {
	case POSITIONAL_ARG:
		this.errorArgList.add(arg)
		break
	case NAMED_ARG:
		arg = this.SyntaxErrors.addDiagnostic(arg,
			DiagnosticErrorCode.ERROR_MISSING_ERROR_MESSAGE_IN_ERROR_CONSTRUCTOR)
		this.errorArgList.add(arg)
		break
	default:
		arg = this.SyntaxErrors.addDiagnostic(arg,
			DiagnosticErrorCode.ERROR_MISSING_ERROR_MESSAGE_IN_ERROR_CONSTRUCTOR)
		arg = this.SyntaxErrors.addDiagnostic(arg, DiagnosticErrorCode.ERROR_REST_ARG_IN_ERROR_CONSTRUCTOR)
		this.errorArgList.add(arg)
		break
	}
	diagnosticErrorCode := DiagnosticErrorCode.ERROR_REST_ARG_IN_ERROR_CONSTRUCTOR
	hasPositionalArg := false
	var leadingComma internal.STNode
	i := 1
	for ; i < len(argList); i = i + 2 {
		leadingComma = this.argList.get(i)
		arg = this.argList.get(i + 1)
		if arg.kind == SyntaxKind.NAMED_ARG {
			this.errorArgList.add(leadingComma)
			this.errorArgList.add(arg)
			continue
		}
		if arg.kind == SyntaxKind.POSITIONAL_ARG {
			if !hasPositionalArg {
				this.errorArgList.add(leadingComma)
				this.errorArgList.add(arg)
				hasPositionalArg = true
				continue
			}
			diagnosticErrorCode = DiagnosticErrorCode.ERROR_ADDITIONAL_POSITIONAL_ARG_IN_ERROR_CONSTRUCTOR
		}
		this.updateLastNodeInListWithInvalidNode(errorArgList, leadingComma, null)
		this.updateLastNodeInListWithInvalidNode(errorArgList, arg, diagnosticErrorCode)
	}
	return this.STNodeFactory.createNodeList(errorArgList)
}

func (this *BallerinaParser) parseArgsList() internal.STNode {
	this.startContext(ParserRuleContext.ARG_LIST)
	token := this.peek()
	if this.isEndOfParametersList(token.kind) {
		args := this.STNodeFactory.createEmptyNodeList()
		this.endContext()
		return args
	}
	firstArg := this.parseArgument()
	argsList := this.parseArgList(firstArg)
	this.endContext()
	return argsList
}

func (this *BallerinaParser) parseArgList(firstArg internal.STNode) internal.STNode {
	argsList := nil
	this.argsList.add(firstArg)
	lastValidArgKind := firstArg.kind
	nextToken := this.peek()
	for !this.isEndOfParametersList(nextToken.kind) {
		argEnd := this.parseArgEnd()
		if argEnd == nil {
			break
		}
		curArg := this.parseArgument()
		errorCode := this.validateArgumentOrder(lastValidArgKind, curArg.kind)
		if errorCode == nil {
			this.argsList.add(argEnd)
			this.argsList.add(curArg)
			lastValidArgKind = curArg.kind
		} else if errorCode == DiagnosticErrorCode.ERROR_NAMED_ARG_FOLLOWED_BY_POSITIONAL_ARG {
			posArg, ok := curArg.(*STPositionalArgumentNode)
			if !ok {
				panic("parseArgList: expected STPositionalArgumentNode")
			}
			if posArg.expression.kind == SyntaxKind.SIMPLE_NAME_REFERENCE {
				missingEqual := this.SyntaxErrors.createMissingToken(SyntaxKind.EQUAL_TOKEN)
				missingIdentifier := this.SyntaxErrors.createMissingToken(SyntaxKind.IDENTIFIER_TOKEN)
				nameRef := this.STNodeFactory.createSimpleNameReferenceNode(missingIdentifier)
				expr := posArg.expression
				simpleNameExpr, ok := expr.(*STSimpleNameReferenceNode)
				if !ok {
					panic("parseArgList: expected STSimpleNameReferenceNode")
				}
				if simpleNameExpr.name.isMissing() {
					errorCode = DiagnosticErrorCode.ERROR_MISSING_NAMED_ARG
					expr = nameRef
				}
				curArg = this.STNodeFactory.createNamedArgumentNode(expr, missingEqual, nameRef)
				curArg = this.SyntaxErrors.addDiagnostic(curArg, errorCode)
				this.argsList.add(argEnd)
				this.argsList.add(curArg)
			}
		}
		nextToken = this.peek()
	}
	return this.STNodeFactory.createNodeList(argsList)
}

func (this *BallerinaParser) validateArgumentOrder(prevArgKind SyntaxKind, curArgKind SyntaxKind) DiagnosticErrorCode {
	errorCode := nil
	switch prevArgKind {
	case POSITIONAL_ARG:
		break
	case NAMED_ARG:
		if curArgKind == SyntaxKind.POSITIONAL_ARG {
			errorCode = DiagnosticErrorCode.ERROR_NAMED_ARG_FOLLOWED_BY_POSITIONAL_ARG
		}
		break
	case REST_ARG:
		errorCode = DiagnosticErrorCode.ERROR_REST_ARG_FOLLOWED_BY_ANOTHER_ARG
		break
	default:
		panic("Invalid SyntaxKind in an argument")
	}
	return errorCode
}

func (this *BallerinaParser) parseArgEnd() internal.STNode {
	switch peek().kind {
	case COMMA_TOKEN:
		return this.parseComma()
	case CLOSE_PAREN_TOKEN:
		return nil
	default:
		this.recover(peek(), ParserRuleContext.ARG_END)
		return this.parseArgEnd()
	}
}

func (this *BallerinaParser) parseArgument() internal.STNode {
	var arg internal.STNode
	nextToken := this.peek()
	switch nextToken.kind {
	case ELLIPSIS_TOKEN:
		ellipsis := this.consume()
		expr := this.parseExpression()
		arg = this.STNodeFactory.createRestArgumentNode(ellipsis, expr)
		break
	case IDENTIFIER_TOKEN:
		arg = this.parseNamedOrPositionalArg()
		break
	default:
		if this.isValidExprStart(nextToken.kind) {
			expr = this.parseExpression()
			arg = this.STNodeFactory.createPositionalArgumentNode(expr)
			break
		}
		this.recover(peek(), ParserRuleContext.ARG_START)
		return this.parseArgument()
	}
	return arg
}

func (this *BallerinaParser) parseNamedOrPositionalArg() internal.STNode {
	argNameOrExpr := this.parseTerminalExpression(true, false, false)
	secondToken := this.peek()
	switch secondToken.kind {
	case EQUAL_TOKEN:
		if argNameOrExpr.kind != SyntaxKind.SIMPLE_NAME_REFERENCE {
			break
		}
		equal := this.parseAssignOp()
		valExpr := this.parseExpression()
		return this.STNodeFactory.createNamedArgumentNode(argNameOrExpr, equal, valExpr)
	case COMMA_TOKEN:
	case CLOSE_PAREN_TOKEN:
		return this.STNodeFactory.createPositionalArgumentNode(argNameOrExpr)
	}
	argNameOrExpr = this.parseExpressionRhs(DEFAULT_OP_PRECEDENCE, argNameOrExpr, true, false)
	return this.STNodeFactory.createPositionalArgumentNode(argNameOrExpr)
}

func (this *BallerinaParser) parseObjectTypeDescriptor(objectKeyword internal.STNode, objectTypeQualifiers internal.STNode) internal.STNode {
	this.startContext(ParserRuleContext.OBJECT_TYPE_DESCRIPTOR)
	openBrace := this.parseOpenBrace()
	objectMemberDescriptors := this.parseObjectMembers(ParserRuleContext.OBJECT_TYPE_MEMBER)
	closeBrace := this.parseCloseBrace()
	this.endContext()
	return this.STNodeFactory.createObjectTypeDescriptorNode(objectTypeQualifiers, objectKeyword, openBrace,
		objectMemberDescriptors, closeBrace)
}

func (this *BallerinaParser) parseObjectConstructorExpression(annots internal.STNode, qualifiers []STNode) internal.STNode {
	this.startContext(ParserRuleContext.OBJECT_CONSTRUCTOR)
	objectTypeQualifier := this.createObjectTypeQualNodeList(qualifiers)
	objectKeyword := this.parseObjectKeyword()
	typeReference := this.parseObjectConstructorTypeReference()
	openBrace := this.parseOpenBrace()
	objectMembers := this.parseObjectMembers(ParserRuleContext.OBJECT_CONSTRUCTOR_MEMBER)
	closeBrace := this.parseCloseBrace()
	this.endContext()
	return this.STNodeFactory.createObjectConstructorExpressionNode(annots,
		objectTypeQualifier, objectKeyword, typeReference, openBrace, objectMembers, closeBrace)
}

func (this *BallerinaParser) parseObjectConstructorTypeReference() internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case OPEN_BRACE_TOKEN:
		return this.STNodeFactory.createEmptyNode()
	default:
		if this.isPredeclaredIdentifier(nextToken.kind) {
			return this.parseTypeReference()
		}
		this.recover(nextToken, ParserRuleContext.OBJECT_CONSTRUCTOR_TYPE_REF)
		return this.parseObjectConstructorTypeReference()
	}
}

func (this *BallerinaParser) isPredeclaredIdentifier(tokenKind SyntaxKind) bool {
	return ((tokenKind == SyntaxKind.IDENTIFIER_TOKEN) || this.isQualifiedIdentifierPredeclaredPrefix(tokenKind))
}

func (this *BallerinaParser) parseObjectKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.OBJECT_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.OBJECT_KEYWORD)
		return this.parseObjectKeyword()
	}
}

func (this *BallerinaParser) parseObjectMembers(context ParserRuleContext) internal.STNode {
	objectMembers := make([]interface{}, 0)
	for !this.isEndOfObjectTypeNode() {
		this.startContext(context)
		member := this.parseObjectMember(context)
		this.endContext()
		if member == nil {
			break
		}
		if (context == ParserRuleContext.OBJECT_CONSTRUCTOR_MEMBER) && (member.kind == SyntaxKind.TYPE_REFERENCE) {
			this.addInvalidNodeToNextToken(member, DiagnosticErrorCode.ERROR_TYPE_INCLUSION_IN_OBJECT_CONSTRUCTOR)
		} else {
			this.objectMembers.add(member)
		}
	}
	return this.STNodeFactory.createNodeList(objectMembers)
}

func (this *BallerinaParser) parseObjectMember(context ParserRuleContext) internal.STNode {
	var metadata internal.STNode
	nextToken := this.peek()
	switch nextToken.kind {
	case EOF_TOKEN:
	case CLOSE_BRACE_TOKEN:
		return nil
	case ASTERISK_TOKEN:
	case PUBLIC_KEYWORD:
	case PRIVATE_KEYWORD:
	case FINAL_KEYWORD:
	case REMOTE_KEYWORD:
	case FUNCTION_KEYWORD:
	case TRANSACTIONAL_KEYWORD:
	case ISOLATED_KEYWORD:
	case RESOURCE_KEYWORD:
		metadata = this.STNodeFactory.createEmptyNode()
		break
	case DOCUMENTATION_STRING:
	case AT_TOKEN:
		metadata = this.parseMetaData()
		break
	case RETURN_KEYWORD:
		this.addInvalidNodeToNextToken(consume(), DiagnosticErrorCode.ERROR_INVALID_TOKEN)
		return this.parseObjectMember(context)
	default:
		if this.isTypeStartingToken(nextToken.kind) {
			metadata = this.STNodeFactory.createEmptyNode()
			break
		}
		var recoveryCtx ParserRuleContext
		if context == ParserRuleContext.OBJECT_CONSTRUCTOR_MEMBER {
			recoveryCtx = ParserRuleContext.OBJECT_CONSTRUCTOR_MEMBER_START
		} else {
			recoveryCtx = ParserRuleContext.CLASS_MEMBER_OR_OBJECT_MEMBER_START
		}
		solution := this.recover(peek(), recoveryCtx)
		if solution.action == Action.KEEP {
			metadata = this.STNodeFactory.createEmptyNode()
			break
		}
		return this.parseObjectMember(context)
	}
	return this.parseObjectMemberWithoutMeta(metadata, context)
}

func (this *BallerinaParser) parseObjectMemberWithoutMeta(metadata internal.STNode, context ParserRuleContext) internal.STNode {
	isObjectTypeDesc := (context == ParserRuleContext.OBJECT_TYPE_MEMBER)
	var recoveryCtx ParserRuleContext
	if context == ParserRuleContext.OBJECT_CONSTRUCTOR_MEMBER {
		recoveryCtx = ParserRuleContext.OBJECT_CONS_MEMBER_WITHOUT_META
	} else {
		recoveryCtx = ParserRuleContext.CLASS_MEMBER_OR_OBJECT_MEMBER_WITHOUT_META
	}
	tyDescQualifiers := make([]interface{}, 0)
	return this.parseObjectMemberWithoutMeta(metadata, tyDescQualifiers, recoveryCtx, isObjectTypeDesc)
}

func (this *BallerinaParser) parseObjectMemberWithoutMeta(metadata internal.STNode, qualifiers []STNode, recoveryCtx ParserRuleContext, isObjectTypeDesc bool) internal.STNode {
	this.parseObjectMemberQualifiers(qualifiers)
	nextToken := this.peek()
	switch nextToken.kind {
	case EOF_TOKEN:
	case CLOSE_BRACE_TOKEN:
		if (metadata != nil) || (!this.qualifiers.isEmpty()) {
			return this.createMissingSimpleObjectField(metadata, qualifiers, isObjectTypeDesc)
		}
		return nil
	case PUBLIC_KEYWORD:
	case PRIVATE_KEYWORD:
		this.reportInvalidQualifierList(qualifiers)
		visibilityQualifier := this.consume()
		if isObjectTypeDesc && (visibilityQualifier.kind == SyntaxKind.PRIVATE_KEYWORD) {
			this.addInvalidNodeToNextToken(visibilityQualifier,
				DiagnosticErrorCode.ERROR_PRIVATE_QUALIFIER_IN_OBJECT_MEMBER_DESCRIPTOR)
			visibilityQualifier = this.STNodeFactory.createEmptyNode()
		}
		return this.parseObjectMethodOrField(metadata, visibilityQualifier, isObjectTypeDesc)
	case FUNCTION_KEYWORD:
		visibilityQualifier = this.STNodeFactory.createEmptyNode()
		return this.parseObjectMethodOrFuncTypeDesc(metadata, visibilityQualifier, qualifiers, isObjectTypeDesc)
	case ASTERISK_TOKEN:
		this.reportInvalidMetaData(metadata, "object ty inclusion")
		this.reportInvalidQualifierList(qualifiers)
		asterisk := this.consume()
		ty := this.parseTypeReferenceInTypeInclusion()
		semicolonToken := this.parseSemicolon()
		return this.STNodeFactory.createTypeReferenceNode(asterisk, ty, semicolonToken)
	case IDENTIFIER_TOKEN:
		if this.isObjectFieldStart() || this.nextToken.isMissing() {
			return this.parseObjectField(metadata, STNodeFactory.createEmptyNode(), qualifiers, isObjectTypeDesc)
		}
		if this.isObjectMethodStart(getNextNextToken()) {
			this.addInvalidTokenToNextToken(errorHandler.consumeInvalidToken())
			return this.parseObjectMemberWithoutMeta(metadata, qualifiers, recoveryCtx, isObjectTypeDesc)
		}
	default:
		if this.isTypeStartingToken(nextToken.kind) && (nextToken.kind != SyntaxKind.IDENTIFIER_TOKEN) {
			return this.parseObjectField(metadata, STNodeFactory.createEmptyNode(), qualifiers, isObjectTypeDesc)
		}
		solution := this.recover(peek(), recoveryCtx)
		if solution.action == Action.KEEP {
			return this.parseObjectField(metadata, STNodeFactory.createEmptyNode(), qualifiers, isObjectTypeDesc)
		}
		return this.parseObjectMemberWithoutMeta(metadata, qualifiers, recoveryCtx, isObjectTypeDesc)
	}
}

func (this *BallerinaParser) isObjectFieldStart() bool {
	nextNextToken := this.getNextNextToken()
	switch nextNextToken.kind {
	case ERROR_KEYWORD, // error-binding-pattern not allowed in fields
		OPEN_BRACE_TOKEN:
		return false
	case CLOSE_BRACE_TOKEN:
		return true
	default:
		return this.isModuleVarDeclStart(1)
	}
}

func (this *BallerinaParser) isObjectMethodStart(token internal.STToken) bool {
	switch token.kind {
	case FUNCTION_KEYWORD,
		REMOTE_KEYWORD,
		RESOURCE_KEYWORD,
		ISOLATED_KEYWORD,
		TRANSACTIONAL_KEYWORD:
		return true
	default:
		return false
	}
}

func (this *BallerinaParser) parseObjectMethodOrField(metadata internal.STNode, visibilityQualifier internal.STNode, isObjectTypeDesc bool) internal.STNode {
	objectMemberQualifiers := make([]interface{}, 0)
	return this.parseObjectMethodOrField(metadata, visibilityQualifier, objectMemberQualifiers, isObjectTypeDesc)
}

func (this *BallerinaParser) parseObjectMethodOrField(metadata internal.STNode, visibilityQualifier internal.STNode, qualifiers []STNode, isObjectTypeDesc bool) internal.STNode {
	this.parseObjectMemberQualifiers(qualifiers)
	nextToken := this.peek(1)
	nextNextToken := this.peek(2)
	switch nextToken.kind {
	case FUNCTION_KEYWORD:
		return this.parseObjectMethodOrFuncTypeDesc(metadata, visibilityQualifier, qualifiers, isObjectTypeDesc)
	case IDENTIFIER_TOKEN:
		if nextNextToken.kind != SyntaxKind.OPEN_PAREN_TOKEN {
			return this.parseObjectField(metadata, visibilityQualifier, qualifiers, isObjectTypeDesc)
		}
		break
	default:
		if this.isTypeStartingToken(nextToken.kind) {
			return this.parseObjectField(metadata, visibilityQualifier, qualifiers, isObjectTypeDesc)
		}
		break
	}
	this.recover(peek(), ParserRuleContext.OBJECT_FUNC_OR_FIELD_WITHOUT_VISIBILITY)
	return this.parseObjectMethodOrField(metadata, visibilityQualifier, qualifiers, isObjectTypeDesc)
}

func (this *BallerinaParser) parseObjectField(metadata internal.STNode, visibilityQualifier internal.STNode, qualifiers []STNode, isObjectTypeDesc bool) internal.STNode {
	objectFieldQualifiers := this.extractObjectFieldQualifiers(qualifiers, isObjectTypeDesc)
	objectFieldQualNodeList := this.STNodeFactory.createNodeList(objectFieldQualifiers)
	ty := this.parseTypeDescriptor(qualifiers, ParserRuleContext.TYPE_DESC_BEFORE_IDENTIFIER)
	fieldName := this.parseVariableName()
	return this.parseObjectFieldRhs(metadata, visibilityQualifier, objectFieldQualNodeList, ty, fieldName,
		isObjectTypeDesc)
}

func (this *BallerinaParser) extractObjectFieldQualifiers(qualifiers []STNode, isObjectTypeDesc bool) []STNode {
	objectFieldQualifiers := make([]interface{}, 0)
	if (!this.qualifiers.isEmpty()) && (!isObjectTypeDesc) {
		firstQualifier := this.qualifiers.get(0)
		if firstQualifier.kind == SyntaxKind.FINAL_KEYWORD {
			this.objectFieldQualifiers.add(qualifiers.remove(0))
		}
	}
	return objectFieldQualifiers
}

func (this *BallerinaParser) parseObjectFieldRhs(metadata internal.STNode, visibilityQualifier internal.STNode, qualifiers internal.STNode, ty internal.STNode, fieldName internal.STNode, isObjectTypeDesc bool) internal.STNode {
	nextToken := this.peek()
	var equalsToken internal.STNode
	var expression internal.STNode
	var semicolonToken internal.STNode
	switch nextToken.kind {
	case SEMICOLON_TOKEN:
		equalsToken = this.STNodeFactory.createEmptyNode()
		expression = this.STNodeFactory.createEmptyNode()
		semicolonToken = this.parseSemicolon()
		break
	case EQUAL_TOKEN:
		equalsToken = this.parseAssignOp()
		expression = this.parseExpression()
		semicolonToken = this.parseSemicolon()
		if isObjectTypeDesc {
			fieldName = this.SyntaxErrors.cloneWithTrailingInvalidNodeMinutiae(fieldName, equalsToken,
				DiagnosticErrorCode.ERROR_FIELD_INITIALIZATION_NOT_ALLOWED_IN_OBJECT_TYPE)
			fieldName = this.SyntaxErrors.cloneWithTrailingInvalidNodeMinutiae(fieldName, expression)
			equalsToken = this.STNodeFactory.createEmptyNode()
			expression = this.STNodeFactory.createEmptyNode()
		}
		break
	default:
		this.recover(peek(), ParserRuleContext.OBJECT_FIELD_RHS)
		return this.parseObjectFieldRhs(metadata, visibilityQualifier, qualifiers, ty, fieldName,
			isObjectTypeDesc)
	}
	return this.STNodeFactory.createObjectFieldNode(metadata, visibilityQualifier, qualifiers, ty, fieldName,
		equalsToken, expression, semicolonToken)
}

func (this *BallerinaParser) parseObjectMethodOrFuncTypeDesc(metadata internal.STNode, visibilityQualifier internal.STNode, qualifiers []STNode, isObjectTypeDesc bool) internal.STNode {
	return this.parseFuncDefOrFuncTypeDesc(metadata, visibilityQualifier, qualifiers, true, isObjectTypeDesc)
}

func (this *BallerinaParser) parseRelativeResourcePath() internal.STNode {
	this.startContext(ParserRuleContext.RELATIVE_RESOURCE_PATH)
	pathElementList := nil
	nextToken := this.peek()
	if nextToken.kind == SyntaxKind.DOT_TOKEN {
		this.pathElementList.add(consume())
		this.endContext()
		return this.STNodeFactory.createNodeList(pathElementList)
	}
	pathSegment := this.parseResourcePathSegment(true)
	this.pathElementList.add(pathSegment)
	var leadingSlash internal.STNode
	for !this.isEndRelativeResourcePath(nextToken.kind) {
		leadingSlash = this.parseRelativeResourcePathEnd()
		if leadingSlash == nil {
			break
		}
		this.pathElementList.add(leadingSlash)
		pathSegment = this.parseResourcePathSegment(false)
		this.pathElementList.add(pathSegment)
		nextToken = this.peek()
	}
	this.endContext()
	return this.createResourcePathNodeList(pathElementList)
}

func (this *BallerinaParser) isEndRelativeResourcePath(tokenKind SyntaxKind) bool {
	switch tokenKind {
	case EOF_TOKEN, OPEN_PAREN_TOKEN:
		return true
	default:
		return false
	}
}

func (this *BallerinaParser) createResourcePathNodeList(pathElementList []STNode) internal.STNode {
	if this.pathElementList.isEmpty() {
		return this.STNodeFactory.createEmptyNodeList()
	}
	validatedList := make([]interface{}, 0)
	firstElement := this.pathElementList.get(0)
	this.validatedList.add(firstElement)
	hasRestPram := (firstElement.kind == SyntaxKind.RESOURCE_PATH_REST_PARAM)
	i := 1
	for ; i < len(pathElementList); i = i + 2 {
		leadingSlash := this.pathElementList.get(i)
		pathSegment := this.pathElementList.get(i + 1)
		if hasRestPram {
			this.updateLastNodeInListWithInvalidNode(validatedList, leadingSlash, null)
			this.updateLastNodeInListWithInvalidNode(validatedList, pathSegment,
				DiagnosticErrorCode.ERROR_RESOURCE_PATH_SEGMENT_NOT_ALLOWED_AFTER_REST_PARAM)
			continue
		}
		hasRestPram = (pathSegment.kind == SyntaxKind.RESOURCE_PATH_REST_PARAM)
		this.validatedList.add(leadingSlash)
		this.validatedList.add(pathSegment)
	}
	return this.STNodeFactory.createNodeList(validatedList)
}

func (this *BallerinaParser) parseResourcePathSegment(isFirstSegment bool) internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case IDENTIFIER_TOKEN:
		if ((isFirstSegment && this.nextToken.isMissing()) && this.isInvalidNodeStackEmpty()) && (getNextNextToken().kind == SyntaxKind.SLASH_TOKEN) {
			this.removeInsertedToken()
			return this.SyntaxErrors.createMissingTokenWithDiagnostics(SyntaxKind.IDENTIFIER_TOKEN,
				DiagnosticErrorCode.ERROR_RESOURCE_PATH_CANNOT_BEGIN_WITH_SLASH)
		}
		return this.consume()
	case OPEN_BRACKET_TOKEN:
		return this.parseResourcePathParameter()
	default:
		this.recover(nextToken, ParserRuleContext.RESOURCE_PATH_SEGMENT)
		return this.parseResourcePathSegment(isFirstSegment)
	}
}

func (this *BallerinaParser) parseResourcePathParameter() internal.STNode {
	openBracket := this.parseOpenBracket()
	annots := this.parseOptionalAnnotations()
	ty := this.parseTypeDescriptor(ParserRuleContext.TYPE_DESC_IN_PATH_PARAM)
	ellipsis := this.parseOptionalEllipsis()
	paramName := this.parseOptionalPathParamName()
	closeBracket := this.parseCloseBracket()
	var pathPramKind SyntaxKind
	if ellipsis == nil {
		pathPramKind = SyntaxKind.RESOURCE_PATH_SEGMENT_PARAM
	} else {
		pathPramKind = SyntaxKind.RESOURCE_PATH_REST_PARAM
	}
	return this.STNodeFactory.createResourcePathParameterNode(pathPramKind, openBracket, annots, ty, ellipsis,
		paramName, closeBracket)
}

func (this *BallerinaParser) parseOptionalPathParamName() internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case IDENTIFIER_TOKEN:
		return this.consume()
	case CLOSE_BRACKET_TOKEN:
		return this.STNodeFactory.createEmptyNode()
	default:
		this.recover(nextToken, ParserRuleContext.OPTIONAL_PATH_PARAM_NAME)
		return this.parseOptionalPathParamName()
	}
}

func (this *BallerinaParser) parseOptionalEllipsis() internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case ELLIPSIS_TOKEN:
		return this.consume()
	case IDENTIFIER_TOKEN, CLOSE_BRACKET_TOKEN:
		return this.STNodeFactory.createEmptyNode()
	default:
		this.recover(nextToken, ParserRuleContext.PATH_PARAM_ELLIPSIS)
		return this.parseOptionalEllipsis()
	}
}

func (this *BallerinaParser) parseRelativeResourcePathEnd() internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case OPEN_PAREN_TOKEN, EOF_TOKEN:
		return nil
	case SLASH_TOKEN:
		return this.consume()
	default:
		this.recover(nextToken, ParserRuleContext.RELATIVE_RESOURCE_PATH_END)
		return this.parseRelativeResourcePathEnd()
	}
}

func (this *BallerinaParser) parseIfElseBlock() internal.STNode {
	this.startContext(ParserRuleContext.IF_BLOCK)
	ifKeyword := this.parseIfKeyword()
	condition := this.parseExpression()
	ifBody := this.parseBlockNode()
	this.endContext()
	elseBody := this.parseElseBlock()
	return this.STNodeFactory.createIfElseStatementNode(ifKeyword, condition, ifBody, elseBody)
}

func (this *BallerinaParser) parseIfKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.IF_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.IF_KEYWORD)
		return this.parseIfKeyword()
	}
}

func (this *BallerinaParser) parseElseKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.ELSE_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.ELSE_KEYWORD)
		return this.parseElseKeyword()
	}
}

func (this *BallerinaParser) parseBlockNode() internal.STNode {
	this.startContext(ParserRuleContext.BLOCK_STMT)
	openBrace := this.parseOpenBrace()
	stmts := this.parseStatements()
	closeBrace := this.parseCloseBrace()
	this.endContext()
	return this.STNodeFactory.createBlockStatementNode(openBrace, stmts, closeBrace)
}

func (this *BallerinaParser) parseElseBlock() internal.STNode {
	nextToken := this.peek()
	if nextToken.kind != SyntaxKind.ELSE_KEYWORD {
		return this.STNodeFactory.createEmptyNode()
	}
	elseKeyword := this.parseElseKeyword()
	elseBody := this.parseElseBody()
	return this.STNodeFactory.createElseBlockNode(elseKeyword, elseBody)
}

func (this *BallerinaParser) parseElseBody() internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind() {
	case IF_KEYWORD:
		this.parseIfElseBlock()
	case OPEN_BRACE_TOKEN:
		this.parseBlockNode()
	default:
		this.recover(peek(), ParserRuleContext.ELSE_BODY)
		this.parseElseBody()
	}
}

func (this *BallerinaParser) parseDoStatement() internal.STNode {
	this.startContext(ParserRuleContext.DO_BLOCK)
	doKeyword := this.parseDoKeyword()
	doBody := this.parseBlockNode()
	this.endContext()
	onFailClause := this.parseOptionalOnFailClause()
	return this.STNodeFactory.createDoStatementNode(doKeyword, doBody, onFailClause)
}

func (this *BallerinaParser) parseWhileStatement() internal.STNode {
	this.startContext(ParserRuleContext.WHILE_BLOCK)
	whileKeyword := this.parseWhileKeyword()
	condition := this.parseExpression()
	whileBody := this.parseBlockNode()
	this.endContext()
	onFailClause := this.parseOptionalOnFailClause()
	return this.STNodeFactory.createWhileStatementNode(whileKeyword, condition, whileBody, onFailClause)
}

func (this *BallerinaParser) parseWhileKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.WHILE_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.WHILE_KEYWORD)
		return this.parseWhileKeyword()
	}
}

func (this *BallerinaParser) parsePanicStatement() internal.STNode {
	this.startContext(ParserRuleContext.PANIC_STMT)
	panicKeyword := this.parsePanicKeyword()
	expression := this.parseExpression()
	semicolon := this.parseSemicolon()
	this.endContext()
	return this.STNodeFactory.createPanicStatementNode(panicKeyword, expression, semicolon)
}

func (this *BallerinaParser) parsePanicKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.PANIC_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.PANIC_KEYWORD)
		return this.parsePanicKeyword()
	}
}

func (this *BallerinaParser) parseCheckExpression(isRhsExpr bool, allowActions bool, isInConditionalExpr bool) internal.STNode {
	checkingKeyword := this.parseCheckingKeyword()
	expr := this.parseExpression(OperatorPrecedence.EXPRESSION_ACTION, isRhsExpr, allowActions, isInConditionalExpr)
	if this.isAction(expr) {
		return this.STNodeFactory.createCheckExpressionNode(SyntaxKind.CHECK_ACTION, checkingKeyword, expr)
	} else {
		return this.STNodeFactory.createCheckExpressionNode(SyntaxKind.CHECK_EXPRESSION, checkingKeyword, expr)
	}
}

func (this *BallerinaParser) parseCheckingKeyword() internal.STNode {
	token := this.peek()
	if (token.kind == SyntaxKind.CHECK_KEYWORD) || (token.kind == SyntaxKind.CHECKPANIC_KEYWORD) {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.CHECKING_KEYWORD)
		return this.parseCheckingKeyword()
	}
}

func (this *BallerinaParser) parseContinueStatement() internal.STNode {
	this.startContext(ParserRuleContext.CONTINUE_STATEMENT)
	continueKeyword := this.parseContinueKeyword()
	semicolon := this.parseSemicolon()
	this.endContext()
	return this.STNodeFactory.createContinueStatementNode(continueKeyword, semicolon)
}

func (this *BallerinaParser) parseContinueKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.CONTINUE_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.CONTINUE_KEYWORD)
		return this.parseContinueKeyword()
	}
}

func (this *BallerinaParser) parseFailStatement() internal.STNode {
	this.startContext(ParserRuleContext.FAIL_STATEMENT)
	failKeyword := this.parseFailKeyword()
	expr := this.parseExpression()
	semicolon := this.parseSemicolon()
	this.endContext()
	return this.STNodeFactory.createFailStatementNode(failKeyword, expr, semicolon)
}

func (this *BallerinaParser) parseFailKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.FAIL_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.FAIL_KEYWORD)
		return this.parseFailKeyword()
	}
}

func (this *BallerinaParser) parseReturnStatement() internal.STNode {
	this.startContext(ParserRuleContext.RETURN_STMT)
	returnKeyword := this.parseReturnKeyword()
	returnRhs := this.parseReturnStatementRhs(returnKeyword)
	this.endContext()
	return returnRhs
}

func (this *BallerinaParser) parseReturnKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.RETURN_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.RETURN_KEYWORD)
		return this.parseReturnKeyword()
	}
}

func (this *BallerinaParser) parseBreakStatement() internal.STNode {
	this.startContext(ParserRuleContext.BREAK_STATEMENT)
	breakKeyword := this.parseBreakKeyword()
	semicolon := this.parseSemicolon()
	this.endContext()
	return this.STNodeFactory.createBreakStatementNode(breakKeyword, semicolon)
}

func (this *BallerinaParser) parseBreakKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.BREAK_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.BREAK_KEYWORD)
		return this.parseBreakKeyword()
	}
}

func (this *BallerinaParser) parseReturnStatementRhs(returnKeyword internal.STNode) internal.STNode {
	var expr internal.STNode
	token := this.peek()
	switch token.kind {
	case SEMICOLON_TOKEN:
		expr = this.STNodeFactory.createEmptyNode()
	default:
		expr = this.parseActionOrExpression()
	}
	semicolon := this.parseSemicolon()
	return this.STNodeFactory.createReturnStatementNode(returnKeyword, expr, semicolon)
}

func (this *BallerinaParser) parseMappingConstructorExpr() internal.STNode {
	this.startContext(ParserRuleContext.MAPPING_CONSTRUCTOR)
	openBrace := this.parseOpenBrace()
	fields := this.parseMappingConstructorFields()
	closeBrace := this.parseCloseBrace()
	this.endContext()
	return this.STNodeFactory.createMappingConstructorExpressionNode(openBrace, fields, closeBrace)
}

func (this *BallerinaParser) parseMappingConstructorFields() internal.STNode {
	nextToken := this.peek()
	if this.isEndOfMappingConstructor(nextToken.kind) {
		return this.STNodeFactory.createEmptyNodeList()
	}
	fields := make([]interface{}, 0)
	field := this.parseMappingField(ParserRuleContext.FIRST_MAPPING_FIELD)
	if field != nil {
		this.fields.add(field)
	}
	return this.parseMappingConstructorFields(fields)
}

func (this *BallerinaParser) parseMappingConstructorFields(fields []STNode) internal.STNode {
	var nextToken internal.STToken
	var mappingFieldEnd internal.STNode
	nextToken = this.peek()
	for !this.isEndOfMappingConstructor(nextToken.kind) {
		mappingFieldEnd = this.parseMappingFieldEnd()
		if mappingFieldEnd == nil {
			break
		}
		this.fields.add(mappingFieldEnd)
		field := this.parseMappingField(ParserRuleContext.MAPPING_FIELD)
		this.fields.add(field)
		nextToken = this.peek()
	}
	return this.STNodeFactory.createNodeList(fields)
}

func (this *BallerinaParser) parseMappingFieldEnd() internal.STNode {
	switch this.peek().kind {
	case COMMA_TOKEN:
		return this.parseComma()
	case CLOSE_BRACE_TOKEN:
		return nil
	default:
		this.recover(this.peek(), common.PARSER_RULE_CONTEXT_MAPPING_FIELD_END)
		return this.parseMappingFieldEnd()
	}
}

func (this *BallerinaParser) isEndOfMappingConstructor(tokenKind SyntaxKind) bool {
	switch tokenKind {
	case IDENTIFIER_TOKEN,
		READONLY_KEYWORD:
		return false
	case EOF_TOKEN,
		DOCUMENTATION_STRING,
		AT_TOKEN,
		CLOSE_BRACE_TOKEN,
		SEMICOLON_TOKEN,
		PUBLIC_KEYWORD,
		PRIVATE_KEYWORD,
		FUNCTION_KEYWORD,
		RETURNS_KEYWORD,
		SERVICE_KEYWORD,
		TYPE_KEYWORD,
		LISTENER_KEYWORD,
		CONST_KEYWORD,
		FINAL_KEYWORD,
		RESOURCE_KEYWORD:
		return true
	default:
		return this.isSimpleType(tokenKind)
	}
}

func (this *BallerinaParser) parseMappingField(fieldContext ParserRuleContext) internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case IDENTIFIER_TOKEN:
		readonlyKeyword := this.STNodeFactory.createEmptyNode()
		return this.parseSpecificFieldWithOptionalValue(readonlyKeyword)
	case STRING_LITERAL_TOKEN:
		readonlyKeyword = this.STNodeFactory.createEmptyNode()
		return this.parseQualifiedSpecificField(readonlyKeyword)
	case READONLY_KEYWORD:
		readonlyKeyword = this.parseReadonlyKeyword()
		return this.parseSpecificField(readonlyKeyword)
	case OPEN_BRACKET_TOKEN:
		return this.parseComputedField()
	case ELLIPSIS_TOKEN:
		ellipsis := this.parseEllipsis()
		expr := this.parseExpression()
		return this.STNodeFactory.createSpreadFieldNode(ellipsis, expr)
	case CLOSE_BRACE_TOKEN:
		if fieldContext == ParserRuleContext.FIRST_MAPPING_FIELD {
			return nil
		}
	default:
		this.recover(nextToken, fieldContext)
		return this.parseMappingField(fieldContext)
	}
}

func (this *BallerinaParser) parseSpecificField(readonlyKeyword internal.STNode) internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case STRING_LITERAL_TOKEN:
		this.parseQualifiedSpecificField(readonlyKeyword)
	case IDENTIFIER_TOKEN:
		this.parseSpecificFieldWithOptionalValue(readonlyKeyword)
	default:
		this.recover(peek(), ParserRuleContext.SPECIFIC_FIELD)
		this.parseSpecificField(readonlyKeyword)
	}
}

func (this *BallerinaParser) parseQualifiedSpecificField(readonlyKeyword internal.STNode) internal.STNode {
	key := this.parseStringLiteral()
	colon := this.parseColon()
	valueExpr := this.parseExpression()
	return this.STNodeFactory.createSpecificFieldNode(readonlyKeyword, key, colon, valueExpr)
}

func (this *BallerinaParser) parseSpecificFieldWithOptionalValue(readonlyKeyword internal.STNode) internal.STNode {
	key := this.parseIdentifier(ParserRuleContext.MAPPING_FIELD_NAME)
	return this.parseSpecificFieldRhs(readonlyKeyword, key)
}

func (this *BallerinaParser) parseSpecificFieldRhs(readonlyKeyword internal.STNode, key internal.STNode) internal.STNode {
	var colon internal.STNode
	var valueExpr internal.STNode
	nextToken := this.peek()
	switch nextToken.kind {
	case COLON_TOKEN:
		colon = this.parseColon()
		valueExpr = this.parseExpression()
		break
	case COMMA_TOKEN:
		colon = this.STNodeFactory.createEmptyNode()
		valueExpr = this.STNodeFactory.createEmptyNode()
		break
	default:
		if this.isEndOfMappingConstructor(nextToken.kind) {
			colon = this.STNodeFactory.createEmptyNode()
			valueExpr = this.STNodeFactory.createEmptyNode()
			break
		}
		this.recover(nextToken, ParserRuleContext.SPECIFIC_FIELD_RHS)
		return this.parseSpecificFieldRhs(readonlyKeyword, key)
	}
	return this.STNodeFactory.createSpecificFieldNode(readonlyKeyword, key, colon, valueExpr)
}

func (this *BallerinaParser) parseStringLiteral() internal.STNode {
	token := this.peek()
	var stringLiteral internal.STNode
	if token.kind == SyntaxKind.STRING_LITERAL_TOKEN {
		stringLiteral = this.consume()
	} else {
		this.recover(token, ParserRuleContext.STRING_LITERAL_TOKEN)
		return this.parseStringLiteral()
	}
	return this.parseBasicLiteral(stringLiteral)
}

func (this *BallerinaParser) parseColon() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.COLON_TOKEN {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.COLON)
		return this.parseColon()
	}
}

func (this *BallerinaParser) parseReadonlyKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.READONLY_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.READONLY_KEYWORD)
		return this.parseReadonlyKeyword()
	}
}

func (this *BallerinaParser) parseComputedField() internal.STNode {
	this.startContext(ParserRuleContext.COMPUTED_FIELD_NAME)
	openBracket := this.parseOpenBracket()
	fieldNameExpr := this.parseExpression()
	closeBracket := this.parseCloseBracket()
	this.endContext()
	colon := this.parseColon()
	valueExpr := this.parseExpression()
	return this.STNodeFactory.createComputedNameFieldNode(openBracket, fieldNameExpr, closeBracket, colon, valueExpr)
}

func (this *BallerinaParser) parseOpenBracket() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.OPEN_BRACKET_TOKEN {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.OPEN_BRACKET)
		return this.parseOpenBracket()
	}
}

func (this *BallerinaParser) parseCompoundAssignmentStmtRhs(lvExpr internal.STNode) internal.STNode {
	binaryOperator := this.parseCompoundBinaryOperator()
	equalsToken := this.parseAssignOp()
	expr := this.parseActionOrExpression()
	semicolon := this.parseSemicolon()
	this.endContext()
	lvExprValid := this.isValidLVExpr(lvExpr)
	if !lvExprValid {
		identifier := this.SyntaxErrors.createMissingToken(SyntaxKind.IDENTIFIER_TOKEN)
		simpleNameRef := this.STNodeFactory.createSimpleNameReferenceNode(identifier)
		lvExpr = this.SyntaxErrors.cloneWithLeadingInvalidNodeMinutiae(simpleNameRef, lvExpr,
			DiagnosticErrorCode.ERROR_INVALID_EXPR_IN_COMPOUND_ASSIGNMENT_LHS)
	}
	return this.STNodeFactory.createCompoundAssignmentStatementNode(lvExpr, binaryOperator, equalsToken, expr,
		semicolon)
}

func (this *BallerinaParser) parseCompoundBinaryOperator() internal.STNode {
	token := this.peek()
	if this.isCompoundAssignment(token.kind) {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.COMPOUND_BINARY_OPERATOR)
		return this.parseCompoundBinaryOperator()
	}
}

func (this *BallerinaParser) parseServiceDeclOrVarDecl(metadata internal.STNode, publicQualifier internal.STNode, qualifiers []STNode) internal.STNode {
	this.startContext(ParserRuleContext.SERVICE_DECL)
	serviceDeclQualList := this.extractServiceDeclQualifiers(qualifiers)
	serviceKeyword := this.extractServiceKeyword(qualifiers)
	typeDesc := this.parseServiceDeclTypeDescriptor(qualifiers)
	if (typeDesc != nil) && (typeDesc.kind == SyntaxKind.OBJECT_TYPE_DESC) {
		return this.parseServiceDeclOrVarDecl(metadata, publicQualifier, serviceDeclQualList, serviceKeyword,
			typeDesc)
	} else {
		return this.parseServiceDecl(metadata, publicQualifier, serviceDeclQualList, serviceKeyword, typeDesc)
	}
}

func (this *BallerinaParser) parseServiceDeclOrVarDecl(metadata internal.STNode, publicQualifier internal.STNode, serviceDeclQualList []STNode, serviceKeyword internal.STNode, typeDesc internal.STNode) internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case SLASH_TOKEN:
	case ON_KEYWORD:
		return this.parseServiceDecl(metadata, publicQualifier, serviceDeclQualList, serviceKeyword, typeDesc)
	case OPEN_BRACKET_TOKEN:
	case IDENTIFIER_TOKEN:
	case OPEN_BRACE_TOKEN:
	case ERROR_KEYWORD:
		this.endContext()
		typeDesc = this.modifyObjectTypeDescWithALeadingQualifier(typeDesc, serviceKeyword)
		if !this.serviceDeclQualList.isEmpty() {
			isolatedQualifier := this.serviceDeclQualList.get(0)
			typeDesc = this.modifyObjectTypeDescWithALeadingQualifier(typeDesc, isolatedQualifier)
		}
		return this.parseVarDeclTypeDescRhs(typeDesc, metadata, publicQualifier, nil, true, true)
	default:
		this.recover(nextToken, ParserRuleContext.SERVICE_DECL_OR_VAR_DECL)
		return this.parseServiceDeclOrVarDecl(metadata, publicQualifier, serviceDeclQualList, serviceKeyword,
			typeDesc)
	}
}

func (this *BallerinaParser) extractServiceDeclQualifiers(qualifierList []STNode) []STNode {
	validatedList := make([]interface{}, 0)
	i := 0
	for ; i < len(qualifierList); i++ {
		qualifier := this.qualifierList.get(i)
		nextIndex := (i + 1)
		if qualifier.kind == SyntaxKind.SERVICE_KEYWORD {
			this.qualifierList.subList(0, i).clear()
			break
		}
		if this.isSyntaxKindInList(validatedList, qualifier.kind) {
			this.updateLastNodeInListWithInvalidNode(validatedList, qualifier,
				DiagnosticErrorCode.ERROR_DUPLICATE_QUALIFIER, (internal.ToToken(qualifier)).text())
			continue
		}
		if qualifier.kind == SyntaxKind.ISOLATED_KEYWORD {
			this.validatedList.add(qualifier)
			continue
		}
		if len(qualifierList) == nextIndex {
			this.addInvalidNodeToNextToken(qualifier, DiagnosticErrorCode.ERROR_QUALIFIER_NOT_ALLOWED,
				(internal.ToToken(qualifier)).text())
		} else {
			this.updateANodeInListWithLeadingInvalidNode(qualifierList, nextIndex, qualifier,
				DiagnosticErrorCode.ERROR_QUALIFIER_NOT_ALLOWED, (internal.ToToken(qualifier)).text())
		}
	}
	return validatedList
}

func (this *BallerinaParser) extractServiceKeyword(qualifierList []STNode) internal.STNode {
	if !this.qualifierList.isEmpty() {
		panic("assertion failed")
	}
	serviceKeyword := this.qualifierList.remove(0)
	if serviceKeyword.kind == SyntaxKind.SERVICE_KEYWORD {
		panic("assertion failed")
	}
	return serviceKeyword
}

func (this *BallerinaParser) parseServiceDecl(metadata internal.STNode, publicQualifier internal.STNode, qualList []STNode, serviceKeyword internal.STNode, serviceType internal.STNode) internal.STNode {
	if publicQualifier != nil {
		if !this.qualList.isEmpty() {
			this.updateFirstNodeInListWithLeadingInvalidNode(qualList, publicQualifier,
				DiagnosticErrorCode.ERROR_QUALIFIER_NOT_ALLOWED)
		} else {
			serviceKeyword = this.SyntaxErrors.cloneWithLeadingInvalidNodeMinutiae(serviceKeyword, publicQualifier,
				DiagnosticErrorCode.ERROR_QUALIFIER_NOT_ALLOWED)
		}
	}
	qualNodeList := this.STNodeFactory.createNodeList(qualList)
	resourcePath := this.parseOptionalAbsolutePathOrStringLiteral()
	onKeyword := this.parseOnKeyword()
	expressionList := this.parseListeners()
	openBrace := this.parseOpenBrace()
	objectMembers := this.parseObjectMembers(ParserRuleContext.OBJECT_CONSTRUCTOR_MEMBER)
	closeBrace := this.parseCloseBrace()
	semicolon := this.parseOptionalSemicolon()
	onKeyword = this.cloneWithDiagnosticIfListEmpty(expressionList, onKeyword, DiagnosticErrorCode.ERROR_MISSING_EXPRESSION)
	this.endContext()
	return this.STNodeFactory.createServiceDeclarationNode(metadata, qualNodeList, serviceKeyword, serviceType,
		resourcePath, onKeyword, expressionList, openBrace, objectMembers, closeBrace, semicolon)
}

func (this *BallerinaParser) parseServiceDeclTypeDescriptor(qualifiers []STNode) internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case SLASH_TOKEN:
	case ON_KEYWORD:
	case STRING_LITERAL_TOKEN:
		this.reportInvalidQualifierList(qualifiers)
		return this.STNodeFactory.createEmptyNode()
	default:
		if this.isTypeStartingToken(nextToken.kind) {
			return this.parseTypeDescriptor(qualifiers, ParserRuleContext.TYPE_DESC_IN_SERVICE)
		}
		this.recover(nextToken, ParserRuleContext.OPTIONAL_SERVICE_DECL_TYPE)
		return this.parseServiceDeclTypeDescriptor(qualifiers)
	}
}

func (this *BallerinaParser) parseOptionalAbsolutePathOrStringLiteral() internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case SLASH_TOKEN:
		return this.parseAbsoluteResourcePath()
	case STRING_LITERAL_TOKEN:
		stringLiteralToken := this.consume()
		stringLiteralNode := this.parseBasicLiteral(stringLiteralToken)
		return this.STNodeFactory.createNodeList(Collections.singletonList(stringLiteralNode))
	case ON_KEYWORD:
		return this.STNodeFactory.createEmptyNodeList()
	default:
		this.recover(nextToken, ParserRuleContext.OPTIONAL_ABSOLUTE_PATH)
		return this.parseOptionalAbsolutePathOrStringLiteral()
	}
}

func (this *BallerinaParser) parseAbsoluteResourcePath() internal.STNode {
	this.startContext(ParserRuleContext.ABSOLUTE_RESOURCE_PATH)
	identifierList := make([]interface{}, 0)
	nextToken := this.peek()
	var leadingSlash internal.STNode
	isInitialSlash := true
	for !this.isEndAbsoluteResourcePath(nextToken.kind) {
		leadingSlash = this.parseAbsoluteResourcePathEnd(isInitialSlash)
		if leadingSlash == nil {
			break
		}
		this.identifierList.add(leadingSlash)
		nextToken = this.peek()
		if isInitialSlash && (nextToken.kind == SyntaxKind.ON_KEYWORD) {
			break
		}
		isInitialSlash = false
		leadingSlash = this.parseIdentifier(ParserRuleContext.IDENTIFIER)
		this.identifierList.add(leadingSlash)
		nextToken = this.peek()
	}
	this.endContext()
	return this.STNodeFactory.createNodeList(identifierList)
}

func (this *BallerinaParser) isEndAbsoluteResourcePath(tokenKind SyntaxKind) bool {
	switch tokenKind {
	case EOF_TOKEN, ON_KEYWORD:
		return true
	default:
		returnfalse
	}
}

func (this *BallerinaParser) parseAbsoluteResourcePathEnd(isInitialSlash bool) internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case ON_KEYWORD:
	case EOF_TOKEN:
		return nil
	case SLASH_TOKEN:
		return this.consume()
	default:
		var context ParserRuleContext
		if isInitialSlash {
			context = ParserRuleContext.OPTIONAL_ABSOLUTE_PATH
		} else {
			context = ParserRuleContext.ABSOLUTE_RESOURCE_PATH_END
		}
		this.recover(nextToken, context)
		return this.parseAbsoluteResourcePathEnd(isInitialSlash)
	}
}

func (this *BallerinaParser) parseServiceKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.SERVICE_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.SERVICE_KEYWORD)
		return this.parseServiceKeyword()
	}
}

func (this *BallerinaParser) isCompoundAssignment(tokenKind SyntaxKind) bool {
	return (this.isCompoundBinaryOperator(tokenKind) && (getNextNextToken().kind == SyntaxKind.EQUAL_TOKEN))
}

func (this *BallerinaParser) parseOnKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.ON_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.ON_KEYWORD)
		return this.parseOnKeyword()
	}
}

func (this *BallerinaParser) parseListeners() internal.STNode {
	this.startContext(ParserRuleContext.LISTENERS_LIST)
	listeners := make([]interface{}, 0)
	nextToken := this.peek()
	if this.isEndOfListeners(nextToken.kind) {
		this.endContext()
		return this.STNodeFactory.createEmptyNodeList()
	}
	expr := this.parseExpression()
	this.listeners.add(expr)
	var listenersMemberEnd internal.STNode
	for !this.isEndOfListeners(peek().kind) {
		listenersMemberEnd = this.parseListenersMemberEnd()
		if listenersMemberEnd == nil {
			break
		}
		this.listeners.add(listenersMemberEnd)
		expr = this.parseExpression()
		this.listeners.add(expr)
	}
	this.endContext()
	return this.STNodeFactory.createNodeList(listeners)
}

func (this *BallerinaParser) isEndOfListeners(tokenKind SyntaxKind) bool {
	switch tokenKind {
	case OPEN_BRACE_TOKEN,
		EOF_TOKEN:
		return true
	default:
		return false
	}
}

func (this *BallerinaParser) parseListenersMemberEnd() internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case COMMA_TOKEN:
		return this.parseComma()
	case OPEN_BRACE_TOKEN:
		return nil
	default:
		this.recover(nextToken, ParserRuleContext.LISTENERS_LIST_END)
		return this.parseListenersMemberEnd()
	}
}

func (this *BallerinaParser) isServiceDeclStart(currentContext ParserRuleContext, lookahead int) bool {
	switch peek(lookahead + 1).kind {
	case IDENTIFIER_TOKEN:
		tokenAfterIdentifier := peek(lookahead + 2).kind
		switch tokenAfterIdentifier {
		case ON_KEYWORD,
			// service foo on ...
			OPEN_BRACE_TOKEN:
			true
		case EQUAL_TOKEN,
			// service foo = ...
			SEMICOLON_TOKEN,
			// service foo;
			QUESTION_MARK_TOKEN:
			false
		default:
			false
		}
	case ON_KEYWORD:
		true
	default:
		false
	}
}

func (this *BallerinaParser) parseListenerDeclaration(metadata internal.STNode, qualifier internal.STNode) internal.STNode {
	this.startContext(ParserRuleContext.LISTENER_DECL)
	listenerKeyword := this.parseListenerKeyword()
	if peek().kind == SyntaxKind.IDENTIFIER_TOKEN {
		listenerDecl := this.parseConstantOrListenerDeclWithOptionalType(metadata, qualifier, listenerKeyword, true)
		this.endContext()
		return listenerDecl
	}
	typeDesc := this.parseTypeDescriptor(ParserRuleContext.TYPE_DESC_BEFORE_IDENTIFIER)
	variableName := this.parseVariableName()
	equalsToken := this.parseAssignOp()
	initializer := this.parseExpression()
	semicolonToken := this.parseSemicolon()
	this.endContext()
	return this.STNodeFactory.createListenerDeclarationNode(metadata, qualifier, listenerKeyword, typeDesc, variableName,
		equalsToken, initializer, semicolonToken)
}

func (this *BallerinaParser) parseListenerKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.LISTENER_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.LISTENER_KEYWORD)
		return this.parseListenerKeyword()
	}
}

func (this *BallerinaParser) parseConstantDeclaration(metadata internal.STNode, qualifier internal.STNode) internal.STNode {
	this.startContext(ParserRuleContext.CONSTANT_DECL)
	constKeyword := this.parseConstantKeyword()
	return this.parseConstDecl(metadata, qualifier, constKeyword)
}

func (this *BallerinaParser) parseConstDecl(metadata internal.STNode, qualifier internal.STNode, constKeyword internal.STNode) internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case ANNOTATION_KEYWORD:
		this.endContext()
		return this.parseAnnotationDeclaration(metadata, qualifier, constKeyword)
	case IDENTIFIER_TOKEN:
		constantDecl := this.parseConstantOrListenerDeclWithOptionalType(metadata, qualifier, constKeyword, false)
		this.endContext()
		return constantDecl
	default:
		if this.isTypeStartingToken(nextToken.kind) {
			break
		}
		this.recover(peek(), ParserRuleContext.CONST_DECL_TYPE)
		return this.parseConstDecl(metadata, qualifier, constKeyword)
	}
	typeDesc := this.parseTypeDescriptor(ParserRuleContext.TYPE_DESC_BEFORE_IDENTIFIER)
	variableName := this.parseVariableName()
	equalsToken := this.parseAssignOp()
	initializer := this.parseExpression()
	semicolonToken := this.parseSemicolon()
	this.endContext()
	return this.STNodeFactory.createConstantDeclarationNode(metadata, qualifier, constKeyword, typeDesc, variableName,
		equalsToken, initializer, semicolonToken)
}

func (this *BallerinaParser) parseConstantOrListenerDeclWithOptionalType(metadata internal.STNode, qualifier internal.STNode, constKeyword internal.STNode, isListener bool) internal.STNode {
	varNameOrTypeName := this.parseStatementStartIdentifier()
	return this.parseConstantOrListenerDeclRhs(metadata, qualifier, constKeyword, varNameOrTypeName, isListener)
}

func (this *BallerinaParser) parseConstantOrListenerDeclRhs(metadata internal.STNode, qualifier internal.STNode, keyword internal.STNode, typeOrVarName internal.STNode, isListener bool) internal.STNode {
	if typeOrVarName.kind == SyntaxKind.QUALIFIED_NAME_REFERENCE {
		ty := typeOrVarName
		variableName := this.parseVariableName()
		return this.parseListenerOrConstRhs(metadata, qualifier, keyword, isListener, ty, variableName)
	}
	var ty internal.STNode
	var variableName internal.STNode
	switch peek().kind {
	case IDENTIFIER_TOKEN:
		ty = typeOrVarName
		variableName = this.parseVariableName()
		break
	case EQUAL_TOKEN:
		simpleNameNode, ok := typeOrVarName.(*STSimpleNameReferenceNode)
		if !ok {
			panic("parseConstantOrListenerDeclRhs: expected STSimpleNameReferenceNode")
		}
		variableName = simpleNameNode.name
		ty = this.STNodeFactory.createEmptyNode()
		break
	default:
		this.recover(peek(), ParserRuleContext.CONST_DECL_RHS)
		return this.parseConstantOrListenerDeclRhs(metadata, qualifier, keyword, typeOrVarName, isListener)
	}
	return this.parseListenerOrConstRhs(metadata, qualifier, keyword, isListener, ty, variableName)
}

func (this *BallerinaParser) parseListenerOrConstRhs(metadata internal.STNode, qualifier internal.STNode, keyword internal.STNode, isListener bool, ty internal.STNode, variableName internal.STNode) internal.STNode {
	equalsToken := this.parseAssignOp()
	initializer := this.parseExpression()
	semicolonToken := this.parseSemicolon()
	if isListener {
		return this.STNodeFactory.createListenerDeclarationNode(metadata, qualifier, keyword, ty, variableName,
			equalsToken, initializer, semicolonToken)
	}
	return this.STNodeFactory.createConstantDeclarationNode(metadata, qualifier, keyword, ty, variableName,
		equalsToken, initializer, semicolonToken)
}

func (this *BallerinaParser) parseConstantKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.CONST_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.CONST_KEYWORD)
		return this.parseConstantKeyword()
	}
}

func (this *BallerinaParser) parseTypeofExpression(isRhsExpr bool, isInConditionalExpr bool) internal.STNode {
	typeofKeyword := this.parseTypeofKeyword()
	expr := this.parseExpression(OperatorPrecedence.UNARY, isRhsExpr, false, isInConditionalExpr)
	return this.STNodeFactory.createTypeofExpressionNode(typeofKeyword, expr)
}

func (this *BallerinaParser) parseTypeofKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.TYPEOF_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.TYPEOF_KEYWORD)
		return this.parseTypeofKeyword()
	}
}

func (this *BallerinaParser) parseOptionalTypeDescriptor(typeDescriptorNode internal.STNode) internal.STNode {
	this.startContext(ParserRuleContext.OPTIONAL_TYPE_DESCRIPTOR)
	questionMarkToken := this.parseQuestionMark()
	this.endContext()
	return this.createOptionalTypeDesc(typeDescriptorNode, questionMarkToken)
}

func (this *BallerinaParser) createOptionalTypeDesc(typeDescNode internal.STNode, questionMarkToken internal.STNode) internal.STNode {
	if typeDescNode.kind == SyntaxKind.UNION_TYPE_DESC {
		unionTypeDesc := internal.STUnionTypeDescriptorNode(typeDescNode)
		middleTypeDesc := this.createOptionalTypeDesc(unionTypeDesc.rightTypeDesc, questionMarkToken)
		typeDescNode = this.mergeTypesWithUnion(unionTypeDesc.leftTypeDesc, unionTypeDesc.pipeToken, middleTypeDesc)
	} else if typeDescNode.kind == SyntaxKind.INTERSECTION_TYPE_DESC {
		intersectionTypeDesc := internal.STIntersectionTypeDescriptorNode(typeDescNode)
		middleTypeDesc := this.createOptionalTypeDesc(intersectionTypeDesc.rightTypeDesc, questionMarkToken)
		typeDescNode = this.mergeTypesWithIntersection(intersectionTypeDesc.leftTypeDesc,
			intersectionTypeDesc.bitwiseAndToken, middleTypeDesc)
	}
	return typeDescNode
}

func (this *BallerinaParser) parseUnaryExpression(isRhsExpr bool, isInConditionalExpr bool) internal.STNode {
	unaryOperator := this.parseUnaryOperator()
	expr := this.parseExpression(OperatorPrecedence.UNARY, isRhsExpr, false, isInConditionalExpr)
	return this.STNodeFactory.createUnaryExpressionNode(unaryOperator, expr)
}

func (this *BallerinaParser) parseUnaryOperator() internal.STNode {
	token := this.peek()
	if this.isUnaryOperator(token.kind) {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.UNARY_OPERATOR)
		return this.parseUnaryOperator()
	}
}

func (this *BallerinaParser) isUnaryOperator(kind SyntaxKind) bool {
	switch kind {
	case PLUS_TOKEN, MINUS_TOKEN, NEGATION_TOKEN, EXCLAMATION_MARK_TOKEN:
		true
	default:
		false
	}
}

func (this *BallerinaParser) parseArrayTypeDescriptor(memberTypeDesc internal.STNode) internal.STNode {
	this.startContext(ParserRuleContext.ARRAY_TYPE_DESCRIPTOR)
	openBracketToken := this.parseOpenBracket()
	arrayLengthNode := this.parseArrayLength()
	closeBracketToken := this.parseCloseBracket()
	this.endContext()
	return this.createArrayTypeDesc(memberTypeDesc, openBracketToken, arrayLengthNode, closeBracketToken)
}

func (this *BallerinaParser) createArrayTypeDesc(memberTypeDesc internal.STNode, openBracketToken internal.STNode, arrayLengthNode internal.STNode, closeBracketToken internal.STNode) internal.STNode {
	memberTypeDesc = this.validateForUsageOfVar(memberTypeDesc)
	if arrayLengthNode != nil {
		switch arrayLengthNode.kind {
		case ASTERISK_LITERAL:
		case SIMPLE_NAME_REFERENCE:
		case QUALIFIED_NAME_REFERENCE:
			break
		case NUMERIC_LITERAL:
			numericLiteralKind := arrayLengthNode.childInBucket(0).kind
			if (numericLiteralKind == SyntaxKind.DECIMAL_INTEGER_LITERAL_TOKEN) || (numericLiteralKind == SyntaxKind.HEX_INTEGER_LITERAL_TOKEN) {
				break
			}
		default:
			openBracketToken = this.SyntaxErrors.cloneWithTrailingInvalidNodeMinutiae(openBracketToken,
				arrayLengthNode, DiagnosticErrorCode.ERROR_INVALID_ARRAY_LENGTH)
			arrayLengthNode = this.STNodeFactory.createEmptyNode()
		}
	}
	arrayDimensions := make([]interface{}, 0)
	if memberTypeDesc.kind == SyntaxKind.ARRAY_TYPE_DESC {
		innerArrayType := internal.STArrayTypeDescriptorNode(memberTypeDesc)
		innerArrayDimensions := innerArrayType.dimensions
		dimensionCount := this.innerArrayDimensions.bucketCount()
		i := 0
		for ; i < dimensionCount; i++ {
			this.arrayDimensions.add(innerArrayDimensions.childInBucket(i))
		}
		memberTypeDesc = innerArrayType.memberTypeDesc
	}
	arrayDimension := this.STNodeFactory.createArrayDimensionNode(openBracketToken, arrayLengthNode,
		closeBracketToken)
	this.arrayDimensions.add(arrayDimension)
	arrayDimensionNodeList := this.STNodeFactory.createNodeList(arrayDimensions)
	return this.STNodeFactory.createArrayTypeDescriptorNode(memberTypeDesc, arrayDimensionNodeList)
}

func (this *BallerinaParser) parseArrayLength() internal.STNode {
	token := this.peek()
	switch token.kind {
	case DECIMAL_INTEGER_LITERAL_TOKEN,
		HEX_INTEGER_LITERAL_TOKEN,
		ASTERISK_TOKEN:
		this.parseBasicLiteral()
	case CLOSE_BRACKET_TOKEN:
		this.STNodeFactory.createEmptyNode()
	case IDENTIFIER_TOKEN:
		this.parseQualifiedIdentifier(ParserRuleContext.ARRAY_LENGTH)
	default:
		this.recover(token, ParserRuleContext.ARRAY_LENGTH)
		this.parseArrayLength()
	}
}

func (this *BallerinaParser) parseOptionalAnnotations() internal.STNode {
	this.startContext(ParserRuleContext.ANNOTATIONS)
	annotList := make([]interface{}, 0)
	nextToken := this.peek()
	for nextToken.kind == SyntaxKind.AT_TOKEN {
		this.annotList.add(parseAnnotation())
		nextToken = this.peek()
	}
	this.endContext()
	return this.STNodeFactory.createNodeList(annotList)
}

func (this *BallerinaParser) parseAnnotations() internal.STNode {
	this.startContext(ParserRuleContext.ANNOTATIONS)
	annotList := make([]interface{}, 0)
	this.annotList.add(parseAnnotation())
	for peek().kind == SyntaxKind.AT_TOKEN {
		this.annotList.add(parseAnnotation())
	}
	this.endContext()
	return this.STNodeFactory.createNodeList(annotList)
}

func (this *BallerinaParser) parseAnnotation() internal.STNode {
	atToken := this.parseAtToken()
	var annotReference internal.STNode
	if this.isPredeclaredIdentifier(peek().kind) {
		annotReference = this.parseQualifiedIdentifier(ParserRuleContext.ANNOT_REFERENCE)
	} else {
		annotReference = this.STNodeFactory.createMissingToken(SyntaxKind.IDENTIFIER_TOKEN)
		annotReference = this.STNodeFactory.createSimpleNameReferenceNode(annotReference)
	}
	var annotValue internal.STNode
	if peek().kind == OPEN_BRACE_TOKEN {
		annotValue = this.parseMappingConstructorExpr()
	} else {
		annotValue = this.STNodeFactory.createEmptyNode()
	}
	return this.STNodeFactory.createAnnotationNode(atToken, annotReference, annotValue)
}

func (this *BallerinaParser) parseAtToken() internal.STNode {
	nextToken := this.peek()
	if nextToken.kind == SyntaxKind.AT_TOKEN {
		return this.consume()
	} else {
		this.recover(nextToken, ParserRuleContext.AT)
		return this.parseAtToken()
	}
}

func (this *BallerinaParser) parseMetaData() internal.STNode {
	var docString internal.STNode
	var annotations internal.STNode
	switch peek().kind {
	case DOCUMENTATION_STRING:
		docString = this.parseMarkdownDocumentation()
		annotations = this.parseOptionalAnnotations()
		break
	case AT_TOKEN:
		docString = this.STNodeFactory.createEmptyNode()
		annotations = this.parseOptionalAnnotations()
		break
	default:
		return this.STNodeFactory.createEmptyNode()
	}
	return this.createMetadata(docString, annotations)
}

func (this *BallerinaParser) createMetadata(docString internal.STNode, annotations internal.STNode) internal.STNode {
	if (annotations == nil) && (docString == nil) {
		return this.STNodeFactory.createEmptyNode()
	} else {
		return this.STNodeFactory.createMetadataNode(docString, annotations)
	}
}

func (this *BallerinaParser) parseTypeTestExpression(lhsExpr internal.STNode, isInConditionalExpr bool) internal.STNode {
	isOrNotIsKeyword := this.parseIsOrNotIsKeyword()
	typeDescriptor := this.parseTypeDescriptorInExpression(isInConditionalExpr)
	return this.STNodeFactory.createTypeTestExpressionNode(lhsExpr, isOrNotIsKeyword, typeDescriptor)
}

func (this *BallerinaParser) parseIsOrNotIsKeyword() internal.STNode {
	token := this.peek()
	if (token.kind == SyntaxKind.IS_KEYWORD) || (token.kind == SyntaxKind.NOT_IS_KEYWORD) {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.IS_KEYWORD)
		return this.parseIsOrNotIsKeyword()
	}
}

func (this *BallerinaParser) parseLocalTypeDefinitionStatement(annots internal.STNode) internal.STNode {
	this.startContext(ParserRuleContext.LOCAL_TYPE_DEFINITION_STMT)
	typeKeyword := this.parseTypeKeyword()
	typeName := this.parseTypeName()
	typeDescriptor := this.parseTypeDescriptor(ParserRuleContext.TYPE_DESC_IN_TYPE_DEF)
	semicolon := this.parseSemicolon()
	this.endContext()
	return this.STNodeFactory.createLocalTypeDefinitionStatementNode(annots, typeKeyword, typeName, typeDescriptor,
		semicolon)
}

func (this *BallerinaParser) parseExpressionStatement(annots internal.STNode) internal.STNode {
	this.startContext(ParserRuleContext.EXPRESSION_STATEMENT)
	expression := this.parseActionOrExpressionInLhs(annots)
	return this.getExpressionAsStatement(expression)
}

func (this *BallerinaParser) parseStatementStartWithExpr(annots internal.STNode) internal.STNode {
	this.startContext(ParserRuleContext.AMBIGUOUS_STMT)
	expr := this.parseActionOrExpressionInLhs(annots)
	return this.parseStatementStartWithExprRhs(expr)
}

func (this *BallerinaParser) parseStatementStartWithExprRhs(expression internal.STNode) internal.STNode {
	nextTokenKind := peek().kind
	if this.isAction(expression) || (nextTokenKind == SyntaxKind.SEMICOLON_TOKEN) {
		return this.getExpressionAsStatement(expression)
	}
	switch nextTokenKind {
	case EQUAL_TOKEN:
		this.switchContext(ParserRuleContext.ASSIGNMENT_STMT)
		return this.parseAssignmentStmtRhs(expression)
	case IDENTIFIER_TOKEN:
	default:
		if this.isCompoundAssignment(nextTokenKind) {
			return this.parseCompoundAssignmentStmtRhs(expression)
		}
		var context ParserRuleContext
		if this.isPossibleExpressionStatement(expression) {
			context = ParserRuleContext.EXPR_STMT_RHS
		} else {
			context = ParserRuleContext.STMT_START_WITH_EXPR_RHS
		}
		this.recover(peek(), context)
		return this.parseStatementStartWithExprRhs(expression)
	}
}

func (this *BallerinaParser) isPossibleExpressionStatement(expression internal.STNode) bool {
	switch expression.kind {
	case METHOD_CALL,
		FUNCTION_CALL,
		CHECK_EXPRESSION,
		REMOTE_METHOD_CALL_ACTION,
		CHECK_ACTION,
		BRACED_ACTION,
		START_ACTION,
		TRAP_ACTION,
		FLUSH_ACTION,
		ASYNC_SEND_ACTION,
		SYNC_SEND_ACTION,
		RECEIVE_ACTION,
		WAIT_ACTION,
		QUERY_ACTION,
		COMMIT_ACTION:
		true
	default:
		false
	}
}

func (this *BallerinaParser) getExpressionAsStatement(expression internal.STNode) internal.STNode {
	switch expression.kind {
	case METHOD_CALL:
	case FUNCTION_CALL:
		return this.parseCallStatement(expression)
	case CHECK_EXPRESSION:
		return this.parseCheckStatement(expression)
	case REMOTE_METHOD_CALL_ACTION:
	case CHECK_ACTION:
	case BRACED_ACTION:
	case START_ACTION:
	case TRAP_ACTION:
	case FLUSH_ACTION:
	case ASYNC_SEND_ACTION:
	case SYNC_SEND_ACTION:
	case RECEIVE_ACTION:
	case WAIT_ACTION:
	case QUERY_ACTION:
	case COMMIT_ACTION:
	case CLIENT_RESOURCE_ACCESS_ACTION:
		return this.parseActionStatement(expression)
	default:
		semicolon := this.parseSemicolon()
		this.endContext()
		expression = this.getExpression(expression)
		exprStmt := this.STNodeFactory.createExpressionStatementNode(SyntaxKind.INVALID_EXPRESSION_STATEMENT,
			expression, semicolon)
		exprStmt = this.SyntaxErrors.addDiagnostic(exprStmt, DiagnosticErrorCode.ERROR_INVALID_EXPRESSION_STATEMENT)
		return exprStmt
	}
}

func (this *BallerinaParser) parseArrayTypeDescriptorNode(indexedExpr internal.STIndexedExpressionNode) internal.STNode {
	memberTypeDesc := this.getTypeDescFromExpr(indexedExpr.containerExpression)
	lengthExprs := internal.STNodeList(indexedExpr.keyExpression)
	if this.lengthExprs.isEmpty() {
		return this.createArrayTypeDesc(memberTypeDesc, indexedExpr.openBracket, STNodeFactory.createEmptyNode(),
			indexedExpr.closeBracket)
	}
	lengthExpr := this.lengthExprs.get(0)
	switch lengthExpr.kind {
	case SIMPLE_NAME_REFERENCE:
		nameRef := internal.STSimpleNameReferenceNode(lengthExpr)
		if this.nameRef.name.isMissing() {
			return this.createArrayTypeDesc(memberTypeDesc, indexedExpr.openBracket, STNodeFactory.createEmptyNode(),
				indexedExpr.closeBracket)
		}
		break
	case ASTERISK_LITERAL:
	case QUALIFIED_NAME_REFERENCE:
		break
	case NUMERIC_LITERAL:
		innerChildKind := lengthExpr.childInBucket(0).kind
		if (innerChildKind == SyntaxKind.DECIMAL_INTEGER_LITERAL_TOKEN) || (innerChildKind == SyntaxKind.HEX_INTEGER_LITERAL_TOKEN) {
			break
		}
	default:
		newOpenBracketWithDiagnostics := this.SyntaxErrors.cloneWithTrailingInvalidNodeMinutiae(
			indexedExpr.openBracket, lengthExpr, DiagnosticErrorCode.ERROR_INVALID_ARRAY_LENGTH)
		indexedExpr = this.indexedExpr.replace(indexedExpr.openBracket, newOpenBracketWithDiagnostics)
		lengthExpr = this.STNodeFactory.createEmptyNode()
	}
	return this.createArrayTypeDesc(memberTypeDesc, indexedExpr.openBracket, lengthExpr, indexedExpr.closeBracket)
}

func (this *BallerinaParser) parseCallStatement(expression internal.STNode) internal.STNode {
	return this.parseCallStatementOrCheckStatement(expression)
}

func (this *BallerinaParser) parseCheckStatement(expression internal.STNode) internal.STNode {
	return this.parseCallStatementOrCheckStatement(expression)
}

func (this *BallerinaParser) parseCallStatementOrCheckStatement(expression internal.STNode) internal.STNode {
	semicolon := this.parseSemicolon()
	this.endContext()
	return this.STNodeFactory.createExpressionStatementNode(SyntaxKind.CALL_STATEMENT, expression, semicolon)
}

func (this *BallerinaParser) parseActionStatement(action internal.STNode) internal.STNode {
	semicolon := this.parseSemicolon()
	this.endContext()
	return this.STNodeFactory.createExpressionStatementNode(SyntaxKind.ACTION_STATEMENT, action, semicolon)
}

func (this *BallerinaParser) parseClientResourceAccessAction(expression internal.STNode, rightArrow internal.STNode, slashToken internal.STNode, isRhsExpr bool, isInMatchGuard bool) internal.STNode {
	this.startContext(ParserRuleContext.CLIENT_RESOURCE_ACCESS_ACTION)
	resourceAccessPath := this.parseOptionalResourceAccessPath(isRhsExpr, isInMatchGuard)
	resourceAccessMethodDot := this.parseOptionalResourceAccessMethodDot(isRhsExpr, isInMatchGuard)
	resourceAccessMethodName := this.STNodeFactory.createEmptyNode()
	if resourceAccessMethodDot != nil {
		resourceAccessMethodName = this.STNodeFactory.createSimpleNameReferenceNode(parseFunctionName())
	}
	resourceMethodCallArgList := this.parseOptionalResourceAccessActionArgList(isRhsExpr, isInMatchGuard)
	this.endContext()
	return this.STNodeFactory.createClientResourceAccessActionNode(expression, rightArrow, slashToken,
		resourceAccessPath, resourceAccessMethodDot, resourceAccessMethodName, resourceMethodCallArgList)
}

func (this *BallerinaParser) parseOptionalResourceAccessPath(isRhsExpr bool, isInMatchGuard bool) internal.STNode {
	resourceAccessPath := this.STNodeFactory.createEmptyNodeList()
	nextToken := this.peek()
	switch nextToken.kind {
	case IDENTIFIER_TOKEN:
	case OPEN_BRACKET_TOKEN:
		resourceAccessPath = this.parseResourceAccessPath(isRhsExpr, isInMatchGuard)
		break
	case DOT_TOKEN:
	case OPEN_PAREN_TOKEN:
		break
	default:
		if this.isEndOfActionOrExpression(nextToken, isRhsExpr, isInMatchGuard) {
			break
		}
		this.recover(nextToken, ParserRuleContext.OPTIONAL_RESOURCE_ACCESS_PATH)
		return this.parseOptionalResourceAccessPath(isRhsExpr, isInMatchGuard)
	}
	return resourceAccessPath
}

func (this *BallerinaParser) parseOptionalResourceAccessMethodDot(isRhsExpr bool, isInMatchGuard bool) internal.STNode {
	dotToken := this.STNodeFactory.createEmptyNode()
	nextToken := this.peek()
	switch nextToken.kind {
	case DOT_TOKEN:
		dotToken = this.consume()
		break
	case OPEN_PAREN_TOKEN:
		break
	default:
		if this.isEndOfActionOrExpression(nextToken, isRhsExpr, isInMatchGuard) {
			break
		}
		this.recover(nextToken, ParserRuleContext.OPTIONAL_RESOURCE_ACCESS_METHOD)
		return this.parseOptionalResourceAccessMethodDot(isRhsExpr, isInMatchGuard)
	}
	return dotToken
}

func (this *BallerinaParser) parseOptionalResourceAccessActionArgList(isRhsExpr bool, isInMatchGuard bool) internal.STNode {
	argList := this.STNodeFactory.createEmptyNode()
	nextToken := this.peek()
	switch nextToken.kind {
	case OPEN_PAREN_TOKEN:
		argList = this.parseParenthesizedArgList()
		break
	default:
		if this.isEndOfActionOrExpression(nextToken, isRhsExpr, isInMatchGuard) {
			break
		}
		this.recover(nextToken, ParserRuleContext.OPTIONAL_RESOURCE_ACCESS_ACTION_ARG_LIST)
		return this.parseOptionalResourceAccessActionArgList(isRhsExpr, isInMatchGuard)
	}
	return argList
}

func (this *BallerinaParser) parseResourceAccessPath(isRhsExpr bool, isInMatchGuard bool) internal.STNode {
	pathSegmentList := make([]interface{}, 0)
	pathSegment := this.parseResourceAccessSegment()
	this.pathSegmentList.add(pathSegment)
	var leadingSlash internal.STNode
	previousPathSegmentNode := pathSegment
	for !this.isEndOfResourceAccessPathSegments(peek(), isRhsExpr, isInMatchGuard) {
		leadingSlash = this.parseResourceAccessSegmentRhs(isRhsExpr, isInMatchGuard)
		if leadingSlash == nil {
			break
		}
		pathSegment = this.parseResourceAccessSegment()
		if previousPathSegmentNode.kind == SyntaxKind.RESOURCE_ACCESS_REST_SEGMENT {
			this.updateLastNodeInListWithInvalidNode(pathSegmentList, leadingSlash, null)
			this.updateLastNodeInListWithInvalidNode(pathSegmentList, pathSegment,
				DiagnosticErrorCode.RESOURCE_ACCESS_SEGMENT_IS_NOT_ALLOWED_AFTER_REST_SEGMENT)
		} else {
			this.pathSegmentList.add(leadingSlash)
			this.pathSegmentList.add(pathSegment)
			previousPathSegmentNode = pathSegment
		}
	}
	return this.STNodeFactory.createNodeList(pathSegmentList)
}

func (this *BallerinaParser) parseResourceAccessSegment() internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case IDENTIFIER_TOKEN:
		this.consume()
	case OPEN_BRACKET_TOKEN:
		this.parseComputedOrResourceAccessRestSegment(consume())
	default:
		this.recover(nextToken, ParserRuleContext.RESOURCE_ACCESS_PATH_SEGMENT)
		this.parseResourceAccessSegment()
	}
}

func (this *BallerinaParser) parseComputedOrResourceAccessRestSegment(openBracket internal.STNode) internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case ELLIPSIS_TOKEN:
		ellipsisToken := this.consume()
		expression := this.parseExpression()
		closeBracketToken := this.parseCloseBracket()
		return this.STNodeFactory.createResourceAccessRestSegmentNode(openBracket, ellipsisToken,
			expression, closeBracketToken)
	default:
		if this.isValidExprStart(nextToken.kind) {
			expression = this.parseExpression()
			closeBracketToken = this.parseCloseBracket()
			return this.STNodeFactory.createComputedResourceAccessSegmentNode(openBracket, expression,
				closeBracketToken)
		}
		this.recover(nextToken, ParserRuleContext.COMPUTED_SEGMENT_OR_REST_SEGMENT)
		return this.parseComputedOrResourceAccessRestSegment(openBracket)
	}
}

func (this *BallerinaParser) parseResourceAccessSegmentRhs(isRhsExpr bool, isInMatchGuard bool) internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case SLASH_TOKEN:
		return this.consume()
	default:
		if this.isEndOfResourceAccessPathSegments(nextToken, isRhsExpr, isInMatchGuard) {
			return nil
		}
		this.recover(nextToken, ParserRuleContext.RESOURCE_ACCESS_SEGMENT_RHS)
		return this.parseResourceAccessSegmentRhs(isRhsExpr, isInMatchGuard)
	}
}

func (this *BallerinaParser) isEndOfResourceAccessPathSegments(nextToken internal.STToken, isRhsExpr bool, isInMatchGuard bool) bool {
	switch nextToken.kind {
	case DOT_TOKEN,
		OPEN_PAREN_TOKEN:
		true
	default:
		this.isEndOfActionOrExpression(nextToken, isRhsExpr, isInMatchGuard)
	}
}

func (this *BallerinaParser) parseRemoteMethodCallOrClientResourceAccessOrAsyncSendAction(expression internal.STNode, isRhsExpr bool, isInMatchGuard bool) internal.STNode {
	rightArrow := this.parseRightArrow()
	return this.parseClientResourceAccessOrAsyncSendActionRhs(expression, rightArrow, isRhsExpr, isInMatchGuard)
}

func (this *BallerinaParser) parseClientResourceAccessOrAsyncSendActionRhs(expression internal.STNode, rightArrow internal.STNode, isRhsExpr bool, isInMatchGuard bool) internal.STNode {
	var name internal.STNode
	nextToken := this.peek()
	switch nextToken.kind {
	case FUNCTION_KEYWORD:
		functionKeyword := this.consume()
		name = this.STNodeFactory.createSimpleNameReferenceNode(functionKeyword)
		return this.parseAsyncSendAction(expression, rightArrow, name)
	case CONTINUE_KEYWORD:
	case COMMIT_KEYWORD:
		name = this.getKeywordAsSimpleNameRef()
		break
	case SLASH_TOKEN:
		slashToken := this.consume()
		return this.parseClientResourceAccessAction(expression, rightArrow, slashToken, isRhsExpr, isInMatchGuard)
	default:
		if nextToken.kind == SyntaxKind.IDENTIFIER_TOKEN {
			nextNextToken := this.getNextNextToken()
			if ((nextNextToken.kind == SyntaxKind.OPEN_PAREN_TOKEN) || this.isEndOfActionOrExpression(nextNextToken, isRhsExpr, isInMatchGuard)) || this.nextToken.isMissing() {
				name = this.STNodeFactory.createSimpleNameReferenceNode(parseFunctionName())
				break
			}
		}
		token := this.peek()
		solution := this.recover(token, ParserRuleContext.REMOTE_OR_RESOURCE_CALL_OR_ASYNC_SEND_RHS)
		if solution.action == Action.KEEP {
			name = this.STNodeFactory.createSimpleNameReferenceNode(parseFunctionName())
			break
		}
		return this.parseClientResourceAccessOrAsyncSendActionRhs(expression, rightArrow, isRhsExpr, isInMatchGuard)
	}
	return this.parseRemoteCallOrAsyncSendEnd(expression, rightArrow, name)
}

func (this *BallerinaParser) parseRemoteCallOrAsyncSendEnd(expression internal.STNode, rightArrow internal.STNode, name internal.STNode) internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case OPEN_PAREN_TOKEN:
		return this.parseRemoteMethodCallAction(expression, rightArrow, name)
	case SEMICOLON_TOKEN:
	case CLOSE_PAREN_TOKEN:
	case OPEN_BRACE_TOKEN:
	case COMMA_TOKEN:
	case FROM_KEYWORD:
	case JOIN_KEYWORD:
	case ON_KEYWORD:
	case LET_KEYWORD:
	case WHERE_KEYWORD:
	case ORDER_KEYWORD:
	case LIMIT_KEYWORD:
	case SELECT_KEYWORD:
		return this.parseAsyncSendAction(expression, rightArrow, name)
	default:
		if this.isGroupOrCollectKeyword(nextToken) {
			return this.parseAsyncSendAction(expression, rightArrow, name)
		}
		this.recover(peek(), ParserRuleContext.REMOTE_CALL_OR_ASYNC_SEND_END)
		return this.parseRemoteCallOrAsyncSendEnd(expression, rightArrow, name)
	}
}

func (this *BallerinaParser) parseAsyncSendAction(expression internal.STNode, rightArrow internal.STNode, peerWorker internal.STNode) internal.STNode {
	return this.STNodeFactory.createAsyncSendActionNode(expression, rightArrow, peerWorker)
}

func (this *BallerinaParser) parseRemoteMethodCallAction(expression internal.STNode, rightArrow internal.STNode, name internal.STNode) internal.STNode {
	openParenToken := this.parseArgListOpenParenthesis()
	arguments := this.parseArgsList()
	closeParenToken := this.parseArgListCloseParenthesis()
	return this.STNodeFactory.createRemoteMethodCallActionNode(expression, rightArrow, name, openParenToken, arguments,
		closeParenToken)
}

func (this *BallerinaParser) parseRightArrow() internal.STNode {
	nextToken := this.peek()
	if nextToken.kind == SyntaxKind.RIGHT_ARROW_TOKEN {
		return this.consume()
	} else {
		this.recover(nextToken, ParserRuleContext.RIGHT_ARROW)
		return this.parseRightArrow()
	}
}

func (this *BallerinaParser) parseMapTypeDescriptor(mapKeyword internal.STNode) internal.STNode {
	typeParameter := this.parseTypeParameter()
	return this.STNodeFactory.createMapTypeDescriptorNode(mapKeyword, typeParameter)
}

func (this *BallerinaParser) parseParameterizedTypeDescriptor(keywordToken internal.STNode) internal.STNode {
	var typeParamNode internal.STNode
	nextToken := this.peek()
	if nextToken.kind == SyntaxKind.LT_TOKEN {
		typeParamNode = this.parseTypeParameter()
	} else {
		typeParamNode = this.STNodeFactory.createEmptyNode()
	}
	parameterizedTypeDescKind := this.getParameterizedTypeDescKind(keywordToken)
	return this.STNodeFactory.createParameterizedTypeDescriptorNode(parameterizedTypeDescKind, keywordToken,
		typeParamNode)
}

func (this *BallerinaParser) getParameterizedTypeDescKind(keywordToken internal.STNode) SyntaxKind {
	switch keywordToken.kind {
	case TYPEDESC_KEYWORD:
		SyntaxKind.TYPEDESC_TYPE_DESC
	case FUTURE_KEYWORD:
		SyntaxKind.FUTURE_TYPE_DESC
	case XML_KEYWORD:
		SyntaxKind.XML_TYPE_DESC
	default:
		SyntaxKind.ERROR_TYPE_DESC
	}
}

func (this *BallerinaParser) parseGTToken() internal.STNode {
	nextToken := this.peek()
	if nextToken.kind == SyntaxKind.GT_TOKEN {
		return this.consume()
	} else {
		this.recover(nextToken, ParserRuleContext.GT)
		return this.parseGTToken()
	}
}

func (this *BallerinaParser) parseLTToken() internal.STNode {
	nextToken := this.peek()
	if nextToken.kind == SyntaxKind.LT_TOKEN {
		return this.consume()
	} else {
		this.recover(nextToken, ParserRuleContext.LT)
		return this.parseLTToken()
	}
}

func (this *BallerinaParser) parseNilLiteral() internal.STNode {
	this.startContext(ParserRuleContext.NIL_LITERAL)
	openParenthesisToken := this.parseOpenParenthesis()
	closeParenthesisToken := this.parseCloseParenthesis()
	this.endContext()
	return this.STNodeFactory.createNilLiteralNode(openParenthesisToken, closeParenthesisToken)
}

func (this *BallerinaParser) parseAnnotationDeclaration(metadata internal.STNode, qualifier internal.STNode, constKeyword internal.STNode) internal.STNode {
	this.startContext(ParserRuleContext.ANNOTATION_DECL)
	annotationKeyword := this.parseAnnotationKeyword()
	annotDecl := this.parseAnnotationDeclFromType(metadata, qualifier, constKeyword, annotationKeyword)
	this.endContext()
	return annotDecl
}

func (this *BallerinaParser) parseAnnotationKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.ANNOTATION_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.ANNOTATION_KEYWORD)
		return this.parseAnnotationKeyword()
	}
}

func (this *BallerinaParser) parseAnnotationDeclFromType(metadata internal.STNode, qualifier internal.STNode, constKeyword internal.STNode, annotationKeyword internal.STNode) internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case IDENTIFIER_TOKEN:
		return this.parseAnnotationDeclWithOptionalType(metadata, qualifier, constKeyword, annotationKeyword)
	default:
		if this.isTypeStartingToken(nextToken.kind) {
			break
		}
		this.recover(peek(), ParserRuleContext.ANNOT_DECL_OPTIONAL_TYPE)
		return this.parseAnnotationDeclFromType(metadata, qualifier, constKeyword, annotationKeyword)
	}
	typeDesc := this.parseTypeDescriptor(ParserRuleContext.TYPE_DESC_IN_ANNOTATION_DECL)
	annotTag := this.parseAnnotationTag()
	return this.parseAnnotationDeclAttachPoints(metadata, qualifier, constKeyword, annotationKeyword, typeDesc,
		annotTag)
}

func (this *BallerinaParser) parseAnnotationTag() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.IDENTIFIER_TOKEN {
		return this.consume()
	} else {
		this.recover(peek(), ParserRuleContext.ANNOTATION_TAG)
		return this.parseAnnotationTag()
	}
}

func (this *BallerinaParser) parseAnnotationDeclWithOptionalType(metadata internal.STNode, qualifier internal.STNode, constKeyword internal.STNode, annotationKeyword internal.STNode) internal.STNode {
	typeDescOrAnnotTag := this.parseQualifiedIdentifier(ParserRuleContext.ANNOT_DECL_OPTIONAL_TYPE)
	if typeDescOrAnnotTag.kind == SyntaxKind.QUALIFIED_NAME_REFERENCE {
		annotTag := this.parseAnnotationTag()
		return this.parseAnnotationDeclAttachPoints(metadata, qualifier, constKeyword, annotationKeyword,
			typeDescOrAnnotTag, annotTag)
	}
	nextToken := this.peek()
	if (nextToken.kind == SyntaxKind.IDENTIFIER_TOKEN) || this.isValidTypeContinuationToken(nextToken) {
		typeDesc := this.parseComplexTypeDescriptor(typeDescOrAnnotTag,
			ParserRuleContext.TYPE_DESC_IN_ANNOTATION_DECL, false)
		annotTag := this.parseAnnotationTag()
		return this.parseAnnotationDeclAttachPoints(metadata, qualifier, constKeyword, annotationKeyword, typeDesc,
			annotTag)
	}
	simplenameNode, ok := typeDescOrAnnotTag.(*STSimpleNameReferenceNode)
	if !ok {
		panic("parseAnnotationDeclWithOptionalType: expected STSimpleNameReferenceNode")
	}
	annotTag := simplenameNode.name
	return this.parseAnnotationDeclRhs(metadata, qualifier, constKeyword, annotationKeyword, annotTag)
}

func (this *BallerinaParser) parseAnnotationDeclRhs(metadata internal.STNode, qualifier internal.STNode, constKeyword internal.STNode, annotationKeyword internal.STNode, typeDescOrAnnotTag internal.STNode) internal.STNode {
	nextToken := this.peek()
	var typeDesc internal.STNode
	var annotTag internal.STNode
	switch nextToken.kind {
	case IDENTIFIER_TOKEN:
		typeDesc = typeDescOrAnnotTag
		annotTag = this.parseAnnotationTag()
		break
	case SEMICOLON_TOKEN:
	case ON_KEYWORD:
		typeDesc = this.STNodeFactory.createEmptyNode()
		annotTag = typeDescOrAnnotTag
		break
	default:
		this.recover(peek(), ParserRuleContext.ANNOT_DECL_RHS)
		return this.parseAnnotationDeclRhs(metadata, qualifier, constKeyword, annotationKeyword, typeDescOrAnnotTag)
	}
	return this.parseAnnotationDeclAttachPoints(metadata, qualifier, constKeyword, annotationKeyword, typeDesc,
		annotTag)
}

func (this *BallerinaParser) parseAnnotationDeclAttachPoints(metadata internal.STNode, qualifier internal.STNode, constKeyword internal.STNode, annotationKeyword internal.STNode, typeDesc internal.STNode, annotTag internal.STNode) internal.STNode {
	var onKeyword internal.STNode
	var attachPoints internal.STNode
	nextToken := this.peek()
	switch nextToken.kind {
	case SEMICOLON_TOKEN:
		onKeyword = this.STNodeFactory.createEmptyNode()
		attachPoints = this.STNodeFactory.createEmptyNodeList()
		break
	case ON_KEYWORD:
		onKeyword = this.parseOnKeyword()
		attachPoints = this.parseAnnotationAttachPoints()
		onKeyword = this.cloneWithDiagnosticIfListEmpty(attachPoints, onKeyword,
			DiagnosticErrorCode.ERROR_MISSING_ANNOTATION_ATTACH_POINT)
		break
	default:
		this.recover(peek(), ParserRuleContext.ANNOT_OPTIONAL_ATTACH_POINTS)
		return this.parseAnnotationDeclAttachPoints(metadata, qualifier, constKeyword, annotationKeyword, typeDesc,
			annotTag)
	}
	semicolonToken := this.parseSemicolon()
	return this.STNodeFactory.createAnnotationDeclarationNode(metadata, qualifier, constKeyword, annotationKeyword,
		typeDesc, annotTag, onKeyword, attachPoints, semicolonToken)
}

func (this *BallerinaParser) parseAnnotationAttachPoints() internal.STNode {
	this.startContext(ParserRuleContext.ANNOT_ATTACH_POINTS_LIST)
	attachPoints := make([]interface{}, 0)
	nextToken := this.peek()
	if this.isEndAnnotAttachPointList(nextToken.kind) {
		this.endContext()
		return this.STNodeFactory.createEmptyNodeList()
	}
	attachPoint := this.parseAnnotationAttachPoint()
	this.attachPoints.add(attachPoint)
	nextToken = this.peek()
	var leadingComma internal.STNode
	for !this.isEndAnnotAttachPointList(nextToken.kind) {
		leadingComma = this.parseAttachPointEnd()
		if leadingComma == nil {
			break
		}
		this.attachPoints.add(leadingComma)
		attachPoint = this.parseAnnotationAttachPoint()
		if attachPoint == nil {
			missingAttachPointIdent := this.SyntaxErrors.createMissingToken(SyntaxKind.TYPE_KEYWORD)
			identList := this.STNodeFactory.createNodeList(missingAttachPointIdent)
			attachPoint = this.STNodeFactory.createAnnotationAttachPointNode(STNodeFactory.createEmptyNode(), identList)
			attachPoint = this.SyntaxErrors.addDiagnostic(attachPoint,
				DiagnosticErrorCode.ERROR_MISSING_ANNOTATION_ATTACH_POINT)
			this.attachPoints.add(attachPoint)
			break
		}
		this.attachPoints.add(attachPoint)
		nextToken = this.peek()
	}
	if (this.attachPoint.lastToken().isMissing() && (this.tokenReader.peek().kind == SyntaxKind.IDENTIFIER_TOKEN)) && (!this.this.tokenReader.head().hasTrailingNewline()) {
		nextNonVirtualToken := this.this.tokenReader.read()
		this.updateLastNodeInListWithInvalidNode(attachPoints, nextNonVirtualToken,
			DiagnosticErrorCode.ERROR_INVALID_TOKEN, nextNonVirtualToken.text())
	}
	this.endContext()
	return this.STNodeFactory.createNodeList(attachPoints)
}

func (this *BallerinaParser) parseAttachPointEnd() internal.STNode {
	switch peek().kind {
	case SEMICOLON_TOKEN:
		nil
	case COMMA_TOKEN:
		this.consume()
	default:
		this.recover(peek(), ParserRuleContext.ATTACH_POINT_END)
		this.parseAttachPointEnd()
	}
}

func (this *BallerinaParser) isEndAnnotAttachPointList(tokenKind SyntaxKind) bool {
	switch tokenKind {
	case EOF_TOKEN, SEMICOLON_TOKEN:
		true
	default:
		false
	}
}

func (this *BallerinaParser) parseAnnotationAttachPoint() internal.STNode {
	switch peek().kind {
	case EOF_TOKEN:
		return nil
	case ANNOTATION_KEYWORD:
	case EXTERNAL_KEYWORD:
	case VAR_KEYWORD:
	case CONST_KEYWORD:
	case LISTENER_KEYWORD:
	case WORKER_KEYWORD:
	case SOURCE_KEYWORD:
		sourceKeyword := this.parseSourceKeyword()
		return this.parseAttachPointIdent(sourceKeyword)
	case OBJECT_KEYWORD:
	case TYPE_KEYWORD:
	case FUNCTION_KEYWORD:
	case PARAMETER_KEYWORD:
	case RETURN_KEYWORD:
	case SERVICE_KEYWORD:
	case FIELD_KEYWORD:
	case RECORD_KEYWORD:
	case CLASS_KEYWORD:
		sourceKeyword = this.STNodeFactory.createEmptyNode()
		firstIdent := this.consume()
		return this.parseDualAttachPointIdent(sourceKeyword, firstIdent)
	default:
		this.recover(peek(), ParserRuleContext.ATTACH_POINT)
		return this.parseAnnotationAttachPoint()
	}
}

func (this *BallerinaParser) parseSourceKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.SOURCE_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.SOURCE_KEYWORD)
		return this.parseSourceKeyword()
	}
}

func (this *BallerinaParser) parseAttachPointIdent(sourceKeyword internal.STNode) internal.STNode {
	switch peek().kind {
	case ANNOTATION_KEYWORD:
	case EXTERNAL_KEYWORD:
	case VAR_KEYWORD:
	case CONST_KEYWORD:
	case LISTENER_KEYWORD:
	case WORKER_KEYWORD:
		firstIdent := this.consume()
		identList := this.STNodeFactory.createNodeList(firstIdent)
		return this.STNodeFactory.createAnnotationAttachPointNode(sourceKeyword, identList)
	case OBJECT_KEYWORD:
	case RESOURCE_KEYWORD:
	case RECORD_KEYWORD:
	case TYPE_KEYWORD:
	case FUNCTION_KEYWORD:
	case PARAMETER_KEYWORD:
	case RETURN_KEYWORD:
	case SERVICE_KEYWORD:
	case FIELD_KEYWORD:
	case CLASS_KEYWORD:
		firstIdent = this.consume()
		return this.parseDualAttachPointIdent(sourceKeyword, firstIdent)
	default:
		this.recover(peek(), ParserRuleContext.ATTACH_POINT_IDENT)
		return this.parseAttachPointIdent(sourceKeyword)
	}
}

func (this *BallerinaParser) parseDualAttachPointIdent(sourceKeyword internal.STNode, firstIdent internal.STNode) internal.STNode {
	var secondIdent internal.STNode
	switch firstIdent.kind {
	case OBJECT_KEYWORD:
		secondIdent = this.parseIdentAfterObjectIdent()
		break
	case RESOURCE_KEYWORD:
		secondIdent = this.parseFunctionIdent()
		break
	case RECORD_KEYWORD:
		secondIdent = this.parseFieldIdent()
		break
	case SERVICE_KEYWORD:
		return this.parseServiceAttachPoint(sourceKeyword, firstIdent)
	case TYPE_KEYWORD:
	case FUNCTION_KEYWORD:
	case PARAMETER_KEYWORD:
	case RETURN_KEYWORD:
	case FIELD_KEYWORD:
	case CLASS_KEYWORD:
	default:
		identList := this.STNodeFactory.createNodeList(firstIdent)
		return this.STNodeFactory.createAnnotationAttachPointNode(sourceKeyword, identList)
	}
	identList := this.STNodeFactory.createNodeList(firstIdent, secondIdent)
	return this.STNodeFactory.createAnnotationAttachPointNode(sourceKeyword, identList)
}

func (this *BallerinaParser) parseRemoteIdent() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.REMOTE_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.REMOTE_IDENT)
		return this.parseRemoteIdent()
	}
}

func (this *BallerinaParser) parseServiceAttachPoint(sourceKeyword internal.STNode, firstIdent internal.STNode) internal.STNode {
	var identList internal.STNode
	token := this.peek()
	switch token.kind {
	case REMOTE_KEYWORD:
		secondIdent := this.parseRemoteIdent()
		thirdIdent := this.parseFunctionIdent()
		identList = this.STNodeFactory.createNodeList(firstIdent, secondIdent, thirdIdent)
		return this.STNodeFactory.createAnnotationAttachPointNode(sourceKeyword, identList)
	case COMMA_TOKEN:
	case SEMICOLON_TOKEN:
		identList = this.STNodeFactory.createNodeList(firstIdent)
		return this.STNodeFactory.createAnnotationAttachPointNode(sourceKeyword, identList)
	default:
		this.recover(token, ParserRuleContext.SERVICE_IDENT_RHS)
		return this.parseServiceAttachPoint(sourceKeyword, firstIdent)
	}
}

func (this *BallerinaParser) parseIdentAfterObjectIdent() internal.STNode {
	token := this.peek()
	switch token.kind {
	case FUNCTION_KEYWORD, FIELD_KEYWORD:
		this.consume()
	default:
		this.recover(token, ParserRuleContext.IDENT_AFTER_OBJECT_IDENT)
		this.parseIdentAfterObjectIdent()
	}
}

func (this *BallerinaParser) parseFunctionIdent() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.FUNCTION_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.FUNCTION_IDENT)
		return this.parseFunctionIdent()
	}
}

func (this *BallerinaParser) parseFieldIdent() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.FIELD_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.FIELD_IDENT)
		return this.parseFieldIdent()
	}
}

func (this *BallerinaParser) parseXMLNamespaceDeclaration(isModuleVar bool) internal.STNode {
	this.startContext(ParserRuleContext.XML_NAMESPACE_DECLARATION)
	xmlnsKeyword := this.parseXMLNSKeyword()
	namespaceUri := this.parseSimpleConstExpr()
	for !this.isValidXMLNameSpaceURI(namespaceUri) {
		xmlnsKeyword = this.SyntaxErrors.cloneWithTrailingInvalidNodeMinutiae(xmlnsKeyword, namespaceUri,
			DiagnosticErrorCode.ERROR_INVALID_XML_NAMESPACE_URI)
		namespaceUri = this.parseSimpleConstExpr()
	}
	xmlnsDecl := this.parseXMLDeclRhs(xmlnsKeyword, namespaceUri, isModuleVar)
	this.endContext()
	return xmlnsDecl
}

func (this *BallerinaParser) parseXMLNSKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.XMLNS_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.XMLNS_KEYWORD)
		return this.parseXMLNSKeyword()
	}
}

func (this *BallerinaParser) isValidXMLNameSpaceURI(expr internal.STNode) bool {
	switch expr.kind {
	case STRING_LITERAL, QUALIFIED_NAME_REFERENCE, SIMPLE_NAME_REFERENCE:
		true
	default:
		false
	}
}

func (this *BallerinaParser) parseSimpleConstExpr() internal.STNode {
	this.startContext(ParserRuleContext.CONSTANT_EXPRESSION)
	expr := this.parseSimpleConstExprInternal()
	this.endContext()
	return expr
}

func (this *BallerinaParser) parseSimpleConstExprInternal() internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case STRING_LITERAL_TOKEN:
	case DECIMAL_INTEGER_LITERAL_TOKEN:
	case HEX_INTEGER_LITERAL_TOKEN:
	case DECIMAL_FLOATING_POINT_LITERAL_TOKEN:
	case HEX_FLOATING_POINT_LITERAL_TOKEN:
	case TRUE_KEYWORD:
	case FALSE_KEYWORD:
	case NULL_KEYWORD:
		return this.parseBasicLiteral()
	case PLUS_TOKEN:
	case MINUS_TOKEN:
		return this.parseSignedIntOrFloat()
	case OPEN_PAREN_TOKEN:
		return this.parseNilLiteral()
	default:
		if this.isPredeclaredIdentifier(nextToken.kind) {
			return this.parseQualifiedIdentifier(ParserRuleContext.VARIABLE_REF)
		}
		this.recover(nextToken, ParserRuleContext.CONSTANT_EXPRESSION_START)
		return this.parseSimpleConstExprInternal()
	}
}

func (this *BallerinaParser) parseXMLDeclRhs(xmlnsKeyword internal.STNode, namespaceUri internal.STNode, isModuleVar bool) internal.STNode {
	asKeyword := this.STNodeFactory.createEmptyNode()
	namespacePrefix := this.STNodeFactory.createEmptyNode()
	switch peek().kind {
	case AS_KEYWORD:
		asKeyword = this.parseAsKeyword()
		namespacePrefix = this.parseNamespacePrefix()
		break
	case SEMICOLON_TOKEN:
		break
	default:
		this.recover(peek(), ParserRuleContext.XML_NAMESPACE_PREFIX_DECL)
		return this.parseXMLDeclRhs(xmlnsKeyword, namespaceUri, isModuleVar)
	}
	semicolon := this.parseSemicolon()
	if isModuleVar {
		return this.STNodeFactory.createModuleXMLNamespaceDeclarationNode(xmlnsKeyword, namespaceUri, asKeyword,
			namespacePrefix, semicolon)
	}
	return this.STNodeFactory.createXMLNamespaceDeclarationNode(xmlnsKeyword, namespaceUri, asKeyword, namespacePrefix,
		semicolon)
}

func (this *BallerinaParser) parseNamespacePrefix() internal.STNode {
	nextToken := this.peek()
	if nextToken.kind == SyntaxKind.IDENTIFIER_TOKEN {
		return this.consume()
	} else {
		this.recover(peek(), ParserRuleContext.NAMESPACE_PREFIX)
		return this.parseNamespacePrefix()
	}
}

func (this *BallerinaParser) parseNamedWorkerDeclaration(annots internal.STNode, qualifiers []STNode) internal.STNode {
	this.startContext(ParserRuleContext.NAMED_WORKER_DECL)
	transactionalKeyword := this.getTransactionalKeyword(qualifiers)
	workerKeyword := this.parseWorkerKeyword()
	workerName := this.parseWorkerName()
	returnTypeDesc := this.parseReturnTypeDescriptor()
	workerBody := this.parseBlockNode()
	this.endContext()
	onFailClause := this.parseOptionalOnFailClause()
	return this.STNodeFactory.createNamedWorkerDeclarationNode(annots, transactionalKeyword, workerKeyword, workerName,
		returnTypeDesc, workerBody, onFailClause)
}

func (this *BallerinaParser) getTransactionalKeyword(qualifierList []STNode) internal.STNode {
	validatedList := make([]interface{}, 0)
	i := 0
	for ; i < len(qualifierList); i++ {
		qualifier := this.qualifierList.get(i)
		nextIndex := (i + 1)
		if this.isSyntaxKindInList(validatedList, qualifier.kind) {
			qualifierToken, ok := qualifier.(STToken)
			if !ok {
				panic("getTransactionalKeyword: expected STToken")
			}
			this.updateLastNodeInListWithInvalidNode(validatedList, qualifier,
				DiagnosticErrorCode.ERROR_DUPLICATE_QUALIFIER, qualifierToken.text())
		} else if qualifier.kind == SyntaxKind.TRANSACTIONAL_KEYWORD {
			this.validatedList.add(qualifier)
		}
	}
	var transactionalKeyword internal.STNode
	if this.validatedList.isEmpty() {
		transactionalKeyword = this.STNodeFactory.createEmptyNode()
	} else {
		transactionalKeyword = this.validatedList.get(0)
	}
	return transactionalKeyword
}

func (this *BallerinaParser) parseReturnTypeDescriptor() internal.STNode {
	token := this.peek()
	if token.kind != SyntaxKind.RETURNS_KEYWORD {
		return this.STNodeFactory.createEmptyNode()
	}
	returnsKeyword := this.consume()
	annot := this.parseOptionalAnnotations()
	ty := this.parseTypeDescriptor(ParserRuleContext.TYPE_DESC_IN_RETURN_TYPE_DESC)
	return this.STNodeFactory.createReturnTypeDescriptorNode(returnsKeyword, annot, ty)
}

func (this *BallerinaParser) parseWorkerKeyword() internal.STNode {
	nextToken := this.peek()
	if nextToken.kind == SyntaxKind.WORKER_KEYWORD {
		return this.consume()
	} else {
		this.recover(peek(), ParserRuleContext.WORKER_KEYWORD)
		return this.parseWorkerKeyword()
	}
}

func (this *BallerinaParser) parseWorkerName() internal.STNode {
	nextToken := this.peek()
	if nextToken.kind == SyntaxKind.IDENTIFIER_TOKEN {
		return this.consume()
	} else {
		this.recover(peek(), ParserRuleContext.WORKER_NAME)
		return this.parseWorkerName()
	}
}

func (this *BallerinaParser) parseLockStatement() internal.STNode {
	this.startContext(ParserRuleContext.LOCK_STMT)
	lockKeyword := this.parseLockKeyword()
	blockStatement := this.parseBlockNode()
	this.endContext()
	onFailClause := this.parseOptionalOnFailClause()
	return this.STNodeFactory.createLockStatementNode(lockKeyword, blockStatement, onFailClause)
}

func (this *BallerinaParser) parseLockKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.LOCK_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.LOCK_KEYWORD)
		return this.parseLockKeyword()
	}
}

func (this *BallerinaParser) parseUnionTypeDescriptor(leftTypeDesc internal.STNode, context ParserRuleContext, isTypedBindingPattern bool) internal.STNode {
	pipeToken := this.consume()
	rightTypeDesc := this.parseTypeDescriptorInternal(nil, context, isTypedBindingPattern, false,
		TypePrecedence.UNION)
	return this.mergeTypesWithUnion(leftTypeDesc, pipeToken, rightTypeDesc)
}

func (this *BallerinaParser) createUnionTypeDesc(leftTypeDesc internal.STNode, pipeToken internal.STNode, rightTypeDesc internal.STNode) internal.STNode {
	leftTypeDesc = this.validateForUsageOfVar(leftTypeDesc)
	rightTypeDesc = this.validateForUsageOfVar(rightTypeDesc)
	return this.STNodeFactory.createUnionTypeDescriptorNode(leftTypeDesc, pipeToken, rightTypeDesc)
}

func (this *BallerinaParser) parsePipeToken() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.PIPE_TOKEN {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.PIPE)
		return this.parsePipeToken()
	}
}

func (this *BallerinaParser) isTypeStartingToken(nodeKind SyntaxKind) bool {
	return this.isTypeStartingToken(nodeKind, getNextNextToken())
}

func (this *BallerinaParser) isSimpleTypeInExpression(nodeKind SyntaxKind) bool {
	switch nodeKind {
	case VAR_KEYWORD, READONLY_KEYWORD:
		false
	default:
		this.isSimpleType(nodeKind)
	}
}

func (this *BallerinaParser) isQualifiedIdentifierPredeclaredPrefix(nodeKind SyntaxKind) bool {
	return (this.isPredeclaredPrefix(nodeKind) && (getNextNextToken().kind == SyntaxKind.COLON_TOKEN))
}

func (this *BallerinaParser) parseForkKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.FORK_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.FORK_KEYWORD)
		return this.parseForkKeyword()
	}
}

func (this *BallerinaParser) parseForkStatement() internal.STNode {
	this.startContext(ParserRuleContext.FORK_STMT)
	forkKeyword := this.parseForkKeyword()
	openBrace := this.parseOpenBrace()
	workers := make([]interface{}, 0)
	for !this.isEndOfStatements() {
		stmt := this.parseStatement()
		if stmt == nil {
			break
		}
		if this.validateStatement(stmt) {
			continue
		}
		switch stmt.kind {
		case NAMED_WORKER_DECLARATION:
			this.workers.add(stmt)
			break
		default:
			if this.workers.isEmpty() {
				openBrace = this.SyntaxErrors.cloneWithTrailingInvalidNodeMinutiae(openBrace, stmt,
					DiagnosticErrorCode.ERROR_ONLY_NAMED_WORKERS_ALLOWED_HERE)
			} else {
				this.updateLastNodeInListWithInvalidNode(workers, stmt,
					DiagnosticErrorCode.ERROR_ONLY_NAMED_WORKERS_ALLOWED_HERE)
			}
		}
	}
	namedWorkerDeclarations := this.STNodeFactory.createNodeList(workers)
	closeBrace := this.parseCloseBrace()
	this.endContext()
	forkStmt := this.STNodeFactory.createForkStatementNode(forkKeyword, openBrace, namedWorkerDeclarations, closeBrace)
	if this.isNodeListEmpty(namedWorkerDeclarations) {
		return this.SyntaxErrors.addDiagnostic(forkStmt,
			DiagnosticErrorCode.ERROR_MISSING_NAMED_WORKER_DECLARATION_IN_FORK_STMT)
	}
	return forkStmt
}

func (this *BallerinaParser) parseTrapExpression(isRhsExpr bool, allowActions bool, isInConditionalExpr bool) internal.STNode {
	trapKeyword := this.parseTrapKeyword()
	expr := this.parseExpression(OperatorPrecedence.TRAP, isRhsExpr, allowActions, isInConditionalExpr)
	if this.isAction(expr) {
		return this.STNodeFactory.createTrapExpressionNode(SyntaxKind.TRAP_ACTION, trapKeyword, expr)
	}
	return this.STNodeFactory.createTrapExpressionNode(SyntaxKind.TRAP_EXPRESSION, trapKeyword, expr)
}

func (this *BallerinaParser) parseTrapKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.TRAP_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.TRAP_KEYWORD)
		return this.parseTrapKeyword()
	}
}

func (this *BallerinaParser) parseListConstructorExpr() internal.STNode {
	this.startContext(ParserRuleContext.LIST_CONSTRUCTOR)
	openBracket := this.parseOpenBracket()
	listMembers := this.parseListMembers()
	closeBracket := this.parseCloseBracket()
	this.endContext()
	return this.STNodeFactory.createListConstructorExpressionNode(openBracket, listMembers, closeBracket)
}

func (this *BallerinaParser) parseListMembers() internal.STNode {
	listMembers := make([]interface{}, 0)
	if this.isEndOfListConstructor(peek().kind) {
		return this.STNodeFactory.createEmptyNodeList()
	}
	listMember := this.parseListMember()
	this.listMembers.add(listMember)
	return this.parseListMembers(listMembers)
}

func (this *BallerinaParser) parseListMembers(listMembers []STNode) internal.STNode {
	var listConstructorMemberEnd internal.STNode
	for !this.isEndOfListConstructor(peek().kind) {
		listConstructorMemberEnd = this.parseListConstructorMemberEnd()
		if listConstructorMemberEnd == nil {
			break
		}
		this.listMembers.add(listConstructorMemberEnd)
		listMember := this.parseListMember()
		this.listMembers.add(listMember)
	}
	return this.STNodeFactory.createNodeList(listMembers)
}

func (this *BallerinaParser) parseListMember() internal.STNode {
	nextToken := this.peek()
	if nextToken.kind == SyntaxKind.ELLIPSIS_TOKEN {
		return this.parseSpreadMember()
	} else {
		return this.parseExpression()
	}
}

func (this *BallerinaParser) parseSpreadMember() internal.STNode {
	ellipsis := this.parseEllipsis()
	expr := this.parseExpression()
	return this.STNodeFactory.createSpreadMemberNode(ellipsis, expr)
}

func (this *BallerinaParser) isEndOfListConstructor(tokenKind SyntaxKind) bool {
	switch tokenKind {
	case EOF_TOKEN, CLOSE_BRACKET_TOKEN:
		true
	default:
		false
	}
}

func (this *BallerinaParser) parseListConstructorMemberEnd() internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case COMMA_TOKEN:
		this.consume()
	case CLOSE_BRACKET_TOKEN:
		nil
	default:
		this.recover(nextToken, ParserRuleContext.LIST_CONSTRUCTOR_MEMBER_END)
		this.parseListConstructorMemberEnd()
	}
}

func (this *BallerinaParser) parseForEachStatement() internal.STNode {
	this.startContext(ParserRuleContext.FOREACH_STMT)
	forEachKeyword := this.parseForEachKeyword()
	typedBindingPattern := this.parseTypedBindingPattern(ParserRuleContext.FOREACH_STMT)
	inKeyword := this.parseInKeyword()
	actionOrExpr := this.parseActionOrExpression()
	blockStatement := this.parseBlockNode()
	this.endContext()
	onFailClause := this.parseOptionalOnFailClause()
	return this.STNodeFactory.createForEachStatementNode(forEachKeyword, typedBindingPattern, inKeyword, actionOrExpr,
		blockStatement, onFailClause)
}

func (this *BallerinaParser) parseForEachKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.FOREACH_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.FOREACH_KEYWORD)
		return this.parseForEachKeyword()
	}
}

func (this *BallerinaParser) parseInKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.IN_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.IN_KEYWORD)
		return this.parseInKeyword()
	}
}

func (this *BallerinaParser) parseTypeCastExpr(isRhsExpr bool, allowActions bool, isInConditionalExpr bool) internal.STNode {
	this.startContext(ParserRuleContext.TYPE_CAST)
	ltToken := this.parseLTToken()
	return this.parseTypeCastExpr(ltToken, isRhsExpr, allowActions, isInConditionalExpr)
}

func (this *BallerinaParser) parseTypeCastExpr(ltToken internal.STNode, isRhsExpr bool, allowActions bool, isInConditionalExpr bool) internal.STNode {
	typeCastParam := this.parseTypeCastParam()
	gtToken := this.parseGTToken()
	this.endContext()
	expression := this.parseExpression(OperatorPrecedence.EXPRESSION_ACTION, isRhsExpr, allowActions, isInConditionalExpr)
	return this.STNodeFactory.createTypeCastExpressionNode(ltToken, typeCastParam, gtToken, expression)
}

func (this *BallerinaParser) parseTypeCastParam() internal.STNode {
	var annot internal.STNode
	var ty internal.STNode
	token := this.peek()
	switch token.kind {
	case AT_TOKEN:
		annot = this.parseOptionalAnnotations()
		token = this.peek()
		if this.isTypeStartingToken(token.kind) {
			ty = this.parseTypeDescriptor(ParserRuleContext.TYPE_DESC_IN_ANGLE_BRACKETS)
		} else {
			ty = this.STNodeFactory.createEmptyNode()
		}
		break
	default:
		annot = this.STNodeFactory.createEmptyNode()
		ty = this.parseTypeDescriptor(ParserRuleContext.TYPE_DESC_IN_ANGLE_BRACKETS)
		break
	}
	return this.STNodeFactory.createTypeCastParamNode(getAnnotations(annot), ty)
}

func (this *BallerinaParser) parseTableConstructorExprRhs(tableKeyword internal.STNode, keySpecifier internal.STNode) internal.STNode {
	this.switchContext(ParserRuleContext.TABLE_CONSTRUCTOR)
	openBracket := this.parseOpenBracket()
	rowList := this.parseRowList()
	closeBracket := this.parseCloseBracket()
	return this.STNodeFactory.createTableConstructorExpressionNode(tableKeyword, keySpecifier, openBracket, rowList,
		closeBracket)
}

func (this *BallerinaParser) parseTableKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.TABLE_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.TABLE_KEYWORD)
		return this.parseTableKeyword()
	}
}

func (this *BallerinaParser) parseRowList() internal.STNode {
	nextToken := this.peek()
	if this.isEndOfTableRowList(nextToken.kind) {
		return this.STNodeFactory.createEmptyNodeList()
	}
	mappings := make([]interface{}, 0)
	mapExpr := this.parseMappingConstructorExpr()
	this.mappings.add(mapExpr)
	nextToken = this.peek()
	var rowEnd internal.STNode
	for !this.isEndOfTableRowList(nextToken.kind) {
		rowEnd = this.parseTableRowEnd()
		if rowEnd == nil {
			break
		}
		this.mappings.add(rowEnd)
		mapExpr = this.parseMappingConstructorExpr()
		this.mappings.add(mapExpr)
		nextToken = this.peek()
	}
	return this.STNodeFactory.createNodeList(mappings)
}

func (this *BallerinaParser) isEndOfTableRowList(tokenKind SyntaxKind) bool {
	switch tokenKind {
	case EOF_TOKEN, CLOSE_BRACKET_TOKEN:
		true
	case COMMA_TOKEN, OPEN_BRACE_TOKEN:
		false
	default:
		this.isEndOfMappingConstructor(tokenKind)
	}
}

func (this *BallerinaParser) parseTableRowEnd() internal.STNode {
	switch peek().kind {
	case COMMA_TOKEN:
		this.parseComma()
	case CLOSE_BRACKET_TOKEN, EOF_TOKEN:
		nil
	default:
		this.recover(peek(), ParserRuleContext.TABLE_ROW_END)
		this.parseTableRowEnd()
	}
}

func (this *BallerinaParser) parseKeySpecifier() internal.STNode {
	this.startContext(ParserRuleContext.KEY_SPECIFIER)
	keyKeyword := this.parseKeyKeyword()
	openParen := this.parseOpenParenthesis()
	fieldNames := this.parseFieldNames()
	closeParen := this.parseCloseParenthesis()
	this.endContext()
	return this.STNodeFactory.createKeySpecifierNode(keyKeyword, openParen, fieldNames, closeParen)
}

func (this *BallerinaParser) parseKeyKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.KEY_KEYWORD {
		return this.consume()
	}
	if this.isKeyKeyword(token) {
		return this.getKeyKeyword(consume())
	}
	this.recover(token, ParserRuleContext.KEY_KEYWORD)
	return this.parseKeyKeyword()
}

func (this *BallerinaParser) getKeyKeyword(token internal.STToken) internal.STNode {
	return this.STNodeFactory.createToken(SyntaxKind.KEY_KEYWORD, token.leadingMinutiae(), token.trailingMinutiae(),
		token.diagnostics())
}

func (this *BallerinaParser) getUnderscoreKeyword(token internal.STToken) internal.STToken {
	return this.STNodeFactory.createToken(SyntaxKind.UNDERSCORE_KEYWORD, token.leadingMinutiae(),
		token.trailingMinutiae(), token.diagnostics())
}

func (this *BallerinaParser) parseNaturalKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.NATURAL_KEYWORD {
		return this.consume()
	}
	if this.isNaturalKeyword(token) {
		return this.getNaturalKeyword(consume())
	}
	this.recover(token, ParserRuleContext.NATURAL_KEYWORD)
	return this.parseNaturalKeyword()
}

func (this *BallerinaParser) isNaturalKeyword(node internal.STNode) bool {
	if node.kind != SyntaxKind.SIMPLE_NAME_REFERENCE {
		return false
	}
	simpleNameNode, ok := node.(*STSimpleNameReferenceNode)
	if !ok {
		panic("isNaturalKeyword: expected STSimpleNameReferenceNode")
	}
	nameToken, ok := simpleNameNode.name.(STToken)
	if !ok {
		panic("isNaturalKeyword: expected STToken")
	}
	return this.isNaturalKeyword(nameToken)
}

func (this *BallerinaParser) getNaturalKeyword(token internal.STToken) internal.STNode {
	return this.STNodeFactory.createToken(SyntaxKind.NATURAL_KEYWORD, token.leadingMinutiae(), token.trailingMinutiae(),
		token.diagnostics())
}

func (this *BallerinaParser) parseFieldNames() internal.STNode {
	nextToken := this.peek()
	if this.isEndOfFieldNamesList(nextToken.kind) {
		return this.STNodeFactory.createEmptyNodeList()
	}
	fieldNames := make([]interface{}, 0)
	fieldName := this.parseVariableName()
	this.fieldNames.add(fieldName)
	nextToken = this.peek()
	var leadingComma internal.STNode
	for !this.isEndOfFieldNamesList(nextToken.kind) {
		leadingComma = this.parseComma()
		this.fieldNames.add(leadingComma)
		fieldName = this.parseVariableName()
		this.fieldNames.add(fieldName)
		nextToken = this.peek()
	}
	return this.STNodeFactory.createNodeList(fieldNames)
}

func (this *BallerinaParser) isEndOfFieldNamesList(tokenKind SyntaxKind) bool {
	switch tokenKind {
	case COMMA_TOKEN, IDENTIFIER_TOKEN:
		false
	default:
		true
	}
}

func (this *BallerinaParser) parseErrorKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.ERROR_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.ERROR_KEYWORD)
		return this.parseErrorKeyword()
	}
}

func (this *BallerinaParser) parseStreamTypeDescriptor(streamKeywordToken internal.STNode) internal.STNode {
	var streamTypeParamsNode internal.STNode
	nextToken := this.peek()
	if nextToken.kind == SyntaxKind.LT_TOKEN {
		streamTypeParamsNode = this.parseStreamTypeParamsNode()
	} else {
		streamTypeParamsNode = this.STNodeFactory.createEmptyNode()
	}
	return this.STNodeFactory.createStreamTypeDescriptorNode(streamKeywordToken, streamTypeParamsNode)
}

func (this *BallerinaParser) parseStreamTypeParamsNode() internal.STNode {
	ltToken := this.parseLTToken()
	this.startContext(ParserRuleContext.TYPE_DESC_IN_STREAM_TYPE_DESC)
	leftTypeDescNode := this.parseTypeDescriptor(ParserRuleContext.TYPE_DESC_IN_STREAM_TYPE_DESC)
	streamTypedesc := this.parseStreamTypeParamsNode(ltToken, leftTypeDescNode)
	this.endContext()
	return streamTypedesc
}

func (this *BallerinaParser) parseStreamTypeParamsNode(ltToken internal.STNode, leftTypeDescNode internal.STNode) internal.STNode {
	var commaToken internal.STNode
	switch peek().kind {
	case COMMA_TOKEN:
		commaToken = this.parseComma()
		rightTypeDescNode = this.parseTypeDescriptor(ParserRuleContext.TYPE_DESC_IN_STREAM_TYPE_DESC)
		break
	case GT_TOKEN:
		commaToken = this.STNodeFactory.createEmptyNode()
		rightTypeDescNode = this.STNodeFactory.createEmptyNode()
		break
	default:
		this.recover(peek(), ParserRuleContext.STREAM_TYPE_FIRST_PARAM_RHS)
		return this.parseStreamTypeParamsNode(ltToken, leftTypeDescNode)
	}
	gtToken = this.parseGTToken()
	return this.STNodeFactory.createStreamTypeParamsNode(ltToken, leftTypeDescNode, commaToken, rightTypeDescNode,
		gtToken)
}

func (this *BallerinaParser) parseLetExpression(isRhsExpr bool, isInConditionalExpr bool) internal.STNode {
	letKeyword := this.parseLetKeyword()
	letVarDeclarations := this.parseLetVarDeclarations(ParserRuleContext.LET_EXPR_LET_VAR_DECL, isRhsExpr, false)
	inKeyword := this.parseInKeyword()
	letKeyword = this.cloneWithDiagnosticIfListEmpty(letVarDeclarations, letKeyword,
		DiagnosticErrorCode.ERROR_MISSING_LET_VARIABLE_DECLARATION)
	expression := this.parseExpression(OperatorPrecedence.REMOTE_CALL_ACTION, isRhsExpr, false,
		isInConditionalExpr)
	return this.STNodeFactory.createLetExpressionNode(letKeyword, letVarDeclarations, inKeyword, expression)
}

func (this *BallerinaParser) parseLetKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.LET_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.LET_KEYWORD)
		return this.parseLetKeyword()
	}
}

func (this *BallerinaParser) parseLetVarDeclarations(context ParserRuleContext, isRhsExpr bool, allowActions bool) internal.STNode {
	this.startContext(context)
	varDecls := make([]interface{}, 0)
	nextToken := this.peek()
	if this.isEndOfLetVarDeclarations(nextToken, getNextNextToken()) {
		this.endContext()
		return this.STNodeFactory.createEmptyNodeList()
	}
	varDec := this.parseLetVarDecl(context, isRhsExpr, allowActions)
	this.varDecls.add(varDec)
	nextToken = this.peek()
	var leadingComma internal.STNode
	for !this.isEndOfLetVarDeclarations(nextToken, getNextNextToken()) {
		leadingComma = this.parseComma()
		this.varDecls.add(leadingComma)
		varDec = this.parseLetVarDecl(context, isRhsExpr, allowActions)
		this.varDecls.add(varDec)
		nextToken = this.peek()
	}
	this.endContext()
	return this.STNodeFactory.createNodeList(varDecls)
}

func (this *BallerinaParser) parseLetVarDecl(context ParserRuleContext, isRhsExpr bool, allowActions bool) internal.STNode {
	annot := this.parseOptionalAnnotations()
	typedBindingPattern := this.parseTypedBindingPattern(ParserRuleContext.LET_EXPR_LET_VAR_DECL)
	assign := this.parseAssignOp()
	var expression internal.STNode
	if context == ParserRuleContext.LET_CLAUSE_LET_VAR_DECL {
		expression = this.parseExpression(OperatorPrecedence.QUERY, isRhsExpr, allowActions)
	} else {
		expression = this.parseExpression(OperatorPrecedence.ANON_FUNC_OR_LET, isRhsExpr, false)
	}
	return this.STNodeFactory.createLetVariableDeclarationNode(annot, typedBindingPattern, assign, expression)
}

func (this *BallerinaParser) parseTemplateExpression() internal.STNode {
	ty := this.STNodeFactory.createEmptyNode()
	startingBackTick := this.parseBacktickToken(ParserRuleContext.TEMPLATE_START)
	content := this.parseTemplateContent()
	endingBackTick := this.parseBacktickToken(ParserRuleContext.TEMPLATE_START)
	return this.STNodeFactory.createTemplateExpressionNode(SyntaxKind.RAW_TEMPLATE_EXPRESSION, ty, startingBackTick,
		content, endingBackTick)
}

func (this *BallerinaParser) parseTemplateContent() internal.STNode {
	items := make([]interface{}, 0)
	nextToken := this.peek()
	for !this.isEndOfBacktickContent(nextToken.kind) {
		contentItem := this.parseTemplateItem()
		this.items.add(contentItem)
		nextToken = this.peek()
	}
	return this.STNodeFactory.createNodeList(items)
}

func (this *BallerinaParser) isEndOfBacktickContent(kind SyntaxKind) bool {
	switch kind {
	case EOF_TOKEN, BACKTICK_TOKEN:
		true
	default:
		false
	}
}

func (this *BallerinaParser) parseTemplateItem() internal.STNode {
	nextToken := this.peek()
	if nextToken.kind == SyntaxKind.INTERPOLATION_START_TOKEN {
		return this.parseInterpolation()
	}
	if nextToken.kind != SyntaxKind.TEMPLATE_STRING {
		nextToken = this.consume()
		return this.STNodeFactory.createLiteralValueToken(SyntaxKind.TEMPLATE_STRING,
			nextToken.text(), nextToken.leadingMinutiae(), nextToken.trailingMinutiae(),
			nextToken.diagnostics())
	}
	return this.consume()
}

func (this *BallerinaParser) parseStringTemplateExpression() internal.STNode {
	ty := this.parseStringKeyword()
	startingBackTick := this.parseBacktickToken(ParserRuleContext.TEMPLATE_START)
	content := this.parseTemplateContent()
	endingBackTick := this.parseBacktickToken(ParserRuleContext.TEMPLATE_END)
	return this.STNodeFactory.createTemplateExpressionNode(SyntaxKind.STRING_TEMPLATE_EXPRESSION, ty, startingBackTick,
		content, endingBackTick)
}

func (this *BallerinaParser) parseStringKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.STRING_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.STRING_KEYWORD)
		return this.parseStringKeyword()
	}
}

func (this *BallerinaParser) parseXMLTemplateExpression() internal.STNode {
	xmlKeyword := this.parseXMLKeyword()
	startingBackTick := this.parseBacktickToken(ParserRuleContext.TEMPLATE_START)
	if this.startingBackTick.isMissing() {
		return this.createMissingTemplateExpressionNode(xmlKeyword, SyntaxKind.XML_TEMPLATE_EXPRESSION)
	}
	content := this.parseTemplateContentAsXML()
	endingBackTick := this.parseBacktickToken(ParserRuleContext.TEMPLATE_END)
	return this.STNodeFactory.createTemplateExpressionNode(SyntaxKind.XML_TEMPLATE_EXPRESSION, xmlKeyword,
		startingBackTick, content, endingBackTick)
}

func (this *BallerinaParser) parseXMLKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.XML_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.XML_KEYWORD)
		return this.parseXMLKeyword()
	}
}

func (this *BallerinaParser) parseTemplateContentAsXML() internal.STNode {
	expressions := make([]interface{}, 0)
	xmlStringBuilder := nil
	nextToken := this.peek()
	for !this.isEndOfBacktickContent(nextToken.kind) {
		contentItem := this.parseTemplateItem()
		if contentItem.kind == SyntaxKind.TEMPLATE_STRING {
			contentToken, ok := contentItem.(STToken)
			if !ok {
				panic("parseTemplateContentAsXML: expected STToken")
			}
			this.xmlStringBuilder.append(contentToken.text())
		} else {
			this.xmlStringBuilder.append("${}")
			this.expressions.add(contentItem)
		}
		nextToken = this.peek()
	}
	charReader := this.CharReader.from(xmlStringBuilder.toString())
	tokenReader := nil
	xmlParser := nil
	return this.xmlParser.parse()
}

func (this *BallerinaParser) parseRegExpTemplateExpression() internal.STNode {
	reKeyword := this.consume()
	startingBackTick := this.parseBacktickToken(ParserRuleContext.TEMPLATE_START)
	if this.startingBackTick.isMissing() {
		return this.createMissingTemplateExpressionNode(reKeyword, SyntaxKind.REGEX_TEMPLATE_EXPRESSION)
	}
	content := this.parseTemplateContentAsRegExp()
	endingBackTick := this.parseBacktickToken(ParserRuleContext.TEMPLATE_END)
	return this.STNodeFactory.createTemplateExpressionNode(SyntaxKind.REGEX_TEMPLATE_EXPRESSION, reKeyword,
		startingBackTick, content, endingBackTick)
}

func (this *BallerinaParser) createMissingTemplateExpressionNode(reKeyword internal.STNode, kind SyntaxKind) internal.STNode {
	startingBackTick := this.SyntaxErrors.createMissingToken(SyntaxKind.BACKTICK_TOKEN)
	endingBackTick := this.SyntaxErrors.createMissingToken(SyntaxKind.BACKTICK_TOKEN)
	content := this.STAbstractNodeFactory.createEmptyNodeList()
	templateExpr := this.STNodeFactory.createTemplateExpressionNode(kind, reKeyword, startingBackTick, content, endingBackTick)
	templateExpr = this.SyntaxErrors.addDiagnostic(templateExpr, DiagnosticErrorCode.ERROR_MISSING_BACKTICK_STRING)
	return templateExpr
}

func (this *BallerinaParser) parseTemplateContentAsRegExp() internal.STNode {
	this.this.tokenReader.startMode(ParserMode.REGEXP)
	expressions := make([]interface{}, 0)
	regExpStringBuilder := nil
	nextToken := this.peek()
	for !this.isEndOfBacktickContent(nextToken.kind) {
		contentItem := this.parseTemplateItem()
		if contentItem.kind == SyntaxKind.TEMPLATE_STRING {
			contentToken, ok := contentItem.(STToken)
			if !ok {
				panic("parseTemplateContentAsRegExp: expected STToken")
			}
			this.regExpStringBuilder.append(contentToken.text())
		} else {
			this.regExpStringBuilder.append("${}")
			this.expressions.add(contentItem)
		}
		nextToken = this.peek()
	}
	this.this.tokenReader.endMode()
	charReader := this.CharReader.from(regExpStringBuilder.toString())
	tokenReader := nil
	regExpParser := nil
	return this.regExpParser.parse()
}

func (this *BallerinaParser) parseInterpolation() internal.STNode {
	this.startContext(ParserRuleContext.INTERPOLATION)
	interpolStart := this.parseInterpolationStart()
	expr := this.parseExpression()
	for !this.isEndOfInterpolation() {
		nextToken := this.consume()
		expr = this.SyntaxErrors.cloneWithTrailingInvalidNodeMinutiae(expr, nextToken,
			DiagnosticErrorCode.ERROR_INVALID_TOKEN, nextToken.text())
	}
	closeBrace := this.parseCloseBrace()
	this.endContext()
	return this.STNodeFactory.createInterpolationNode(interpolStart, expr, closeBrace)
}

func (this *BallerinaParser) isEndOfInterpolation() bool {
	nextTokenKind := peek().kind
	switch nextTokenKind {
	case EOF_TOKEN, BACKTICK_TOKEN:
		true
	default:
		currentLexerMode := this.this.tokenReader.getCurrentMode()
		(((nextTokenKind == SyntaxKind.CLOSE_BRACE_TOKEN) && (currentLexerMode != ParserMode.INTERPOLATION)) && (currentLexerMode != ParserMode.INTERPOLATION_BRACED_CONTENT))
	}
}

func (this *BallerinaParser) parseInterpolationStart() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.INTERPOLATION_START_TOKEN {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.INTERPOLATION_START_TOKEN)
		return this.parseInterpolationStart()
	}
}

func (this *BallerinaParser) parseBacktickToken(ctx ParserRuleContext) internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.BACKTICK_TOKEN {
		return this.consume()
	} else {
		this.recover(token, ctx)
		return this.parseBacktickToken(ctx)
	}
}

func (this *BallerinaParser) parseTableTypeDescriptor(tableKeywordToken internal.STNode) internal.STNode {
	rowTypeParameterNode := this.parseRowTypeParameter()
	var keyConstraintNode internal.STNode
	nextToken := this.peek()
	if this.isKeyKeyword(nextToken) {
		keyKeywordToken := this.getKeyKeyword(consume())
		keyConstraintNode = this.parseKeyConstraint(keyKeywordToken)
	} else {
		keyConstraintNode = this.STNodeFactory.createEmptyNode()
	}
	return this.STNodeFactory.createTableTypeDescriptorNode(tableKeywordToken, rowTypeParameterNode, keyConstraintNode)
}

func (this *BallerinaParser) parseRowTypeParameter() internal.STNode {
	this.startContext(ParserRuleContext.ROW_TYPE_PARAM)
	rowTypeParameterNode := this.parseTypeParameter()
	this.endContext()
	return rowTypeParameterNode
}

func (this *BallerinaParser) parseTypeParameter() internal.STNode {
	ltToken := this.parseLTToken()
	typeNode := this.parseTypeDescriptor(ParserRuleContext.TYPE_DESC_IN_ANGLE_BRACKETS)
	gtToken := this.parseGTToken()
	return this.STNodeFactory.createTypeParameterNode(ltToken, typeNode, gtToken)
}

func (this *BallerinaParser) parseKeyConstraint(keyKeywordToken internal.STNode) internal.STNode {
	switch peek().kind {
	case OPEN_PAREN_TOKEN:
		this.parseKeySpecifier(keyKeywordToken)
	case LT_TOKEN:
		this.parseKeyTypeConstraint(keyKeywordToken)
	default:
		this.recover(peek(), ParserRuleContext.KEY_CONSTRAINTS_RHS)
		this.parseKeyConstraint(keyKeywordToken)
	}
}

func (this *BallerinaParser) parseKeySpecifier(keyKeywordToken internal.STNode) internal.STNode {
	this.startContext(ParserRuleContext.KEY_SPECIFIER)
	openParenToken := this.parseOpenParenthesis()
	fieldNamesNode := this.parseFieldNames()
	closeParenToken := this.parseCloseParenthesis()
	this.endContext()
	return this.STNodeFactory.createKeySpecifierNode(keyKeywordToken, openParenToken, fieldNamesNode, closeParenToken)
}

func (this *BallerinaParser) parseKeyTypeConstraint(keyKeywordToken internal.STNode) internal.STNode {
	typeParameterNode := this.parseTypeParameter()
	return this.STNodeFactory.createKeyTypeConstraintNode(keyKeywordToken, typeParameterNode)
}

func (this *BallerinaParser) parseFunctionTypeDesc(qualifiers []STNode) internal.STNode {
	this.startContext(ParserRuleContext.FUNC_TYPE_DESC)
	functionKeyword := this.parseFunctionKeyword()
	hasFuncSignature := false
	signature := this.STNodeFactory.createEmptyNode()
	if (peek().kind == SyntaxKind.OPEN_PAREN_TOKEN) || this.isSyntaxKindInList(qualifiers, SyntaxKind.TRANSACTIONAL_KEYWORD) {
		signature = this.parseFuncSignature(true)
		hasFuncSignature = true
	}
	nodes := this.createFuncTypeQualNodeList(qualifiers, functionKeyword, hasFuncSignature)
	qualifierList := nodes[0]
	functionKeyword = nodes[1]
	this.endContext()
	return this.STNodeFactory.createFunctionTypeDescriptorNode(qualifierList, functionKeyword, signature)
}

func (this *BallerinaParser) getLastNodeInList(nodeList []STNode) internal.STNode {
	return this.nodeList.get(nodeList.size() - 1)
}

func (this *BallerinaParser) createFuncTypeQualNodeList(qualifierList []STNode, functionKeyword internal.STNode, hasFuncSignature bool) []internal.STNode {
	validatedList := make([]interface{}, 0)
	i := 0
	for ; i < len(qualifierList); i++ {
		qualifier := this.qualifierList.get(i)
		nextIndex := (i + 1)
		if this.isSyntaxKindInList(validatedList, qualifier.kind) {
			qualifierToken, ok := qualifier.(STToken)
			if !ok {
				panic("createFuncTypeQualNodeList: expected STToken")
			}
			this.updateLastNodeInListWithInvalidNode(validatedList, qualifier,
				DiagnosticErrorCode.ERROR_DUPLICATE_QUALIFIER, qualifierToken.text())
		} else if hasFuncSignature && this.isRegularFuncQual(qualifier.kind) {
			this.validatedList.add(qualifier)
		}
	}
	nodeList := this.STNodeFactory.createNodeList(validatedList)
	return nil
}

func (this *BallerinaParser) isRegularFuncQual(tokenKind SyntaxKind) bool {
	switch tokenKind {
	case ISOLATED_KEYWORD, TRANSACTIONAL_KEYWORD:
		true
	default:
		false
	}
}

func (this *BallerinaParser) parseExplicitFunctionExpression(annots internal.STNode, qualifiers []STNode, isRhsExpr bool) internal.STNode {
	this.startContext(ParserRuleContext.ANON_FUNC_EXPRESSION)
	funcKeyword := this.parseFunctionKeyword()
	nodes := this.createFuncTypeQualNodeList(qualifiers, funcKeyword, true)
	qualifierList := nodes[0]
	funcKeyword = nodes[1]
	funcSignature := this.parseFuncSignature(false)
	funcBody := this.parseAnonFuncBody(isRhsExpr)
	return this.STNodeFactory.createExplicitAnonymousFunctionExpressionNode(annots, qualifierList, funcKeyword,
		funcSignature, funcBody)
}

func (this *BallerinaParser) parseAnonFuncBody(isRhsExpr bool) internal.STNode {
	switch peek().kind {
	case OPEN_BRACE_TOKEN:
	case EOF_TOKEN:
		body := this.parseFunctionBodyBlock(true)
		this.endContext()
		return body
	case RIGHT_DOUBLE_ARROW_TOKEN:
		this.endContext()
		return this.parseExpressionFuncBody(true, isRhsExpr)
	default:
		this.recover(peek(), ParserRuleContext.ANON_FUNC_BODY)
		return this.parseAnonFuncBody(isRhsExpr)
	}
}

func (this *BallerinaParser) parseExpressionFuncBody(isAnon bool, isRhsExpr bool) internal.STNode {
	rightDoubleArrow := this.parseDoubleRightArrow()
	expression := this.parseExpression(OperatorPrecedence.REMOTE_CALL_ACTION, isRhsExpr, false)
	var semiColon internal.STNode
	if isAnon {
		semiColon = this.STNodeFactory.createEmptyNode()
	} else {
		semiColon = this.parseSemicolon()
	}
	return this.STNodeFactory.createExpressionFunctionBodyNode(rightDoubleArrow, expression, semiColon)
}

func (this *BallerinaParser) parseDoubleRightArrow() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.RIGHT_DOUBLE_ARROW_TOKEN {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.EXPR_FUNC_BODY_START)
		return this.parseDoubleRightArrow()
	}
}

func (this *BallerinaParser) parseImplicitAnonFunc(params internal.STNode, isRhsExpr bool) internal.STNode {
	switch params.kind {
	case SIMPLE_NAME_REFERENCE:
	case INFER_PARAM_LIST:
		break
	case BRACED_EXPRESSION:
		bracedExpr, ok := params.(*STBracedExpressionNode)
		if !ok {
			panic("parseImplicitAnonFunc: expected STBracedExpressionNode")
		}
		params = this.getAnonFuncParam(bracedExpr)
		break
	case NIL_LITERAL:
		nilLiteralNode := internal.STNilLiteralNode(params)
		params = this.STNodeFactory.createImplicitAnonymousFunctionParameters(nilLiteralNode.openParenToken,
			STNodeFactory.createNodeList(nil), nilLiteralNode.closeParenToken)
		break
	default:
		syntheticParam := this.STNodeFactory.createMissingToken(SyntaxKind.IDENTIFIER_TOKEN)
		syntheticParam = this.SyntaxErrors.cloneWithLeadingInvalidNodeMinutiae(syntheticParam, params,
			DiagnosticErrorCode.ERROR_INVALID_PARAM_LIST_IN_INFER_ANONYMOUS_FUNCTION_EXPR)
		params = this.STNodeFactory.createSimpleNameReferenceNode(syntheticParam)
	}
	rightDoubleArrow := this.parseDoubleRightArrow()
	expression := this.parseExpression(OperatorPrecedence.REMOTE_CALL_ACTION, isRhsExpr, false)
	return this.STNodeFactory.createImplicitAnonymousFunctionExpressionNode(params, rightDoubleArrow, expression)
}

func (this *BallerinaParser) getAnonFuncParam(bracedExpression internal.STBracedExpressionNode) internal.STNode {
	paramList := make([]interface{}, 0)
	innerExpression := bracedExpression.expression
	openParen := bracedExpression.openParen
	if innerExpression.kind == SyntaxKind.SIMPLE_NAME_REFERENCE {
		this.paramList.add(innerExpression)
	} else {
		openParen = this.SyntaxErrors.cloneWithTrailingInvalidNodeMinutiae(openParen, innerExpression,
			DiagnosticErrorCode.ERROR_INVALID_PARAM_LIST_IN_INFER_ANONYMOUS_FUNCTION_EXPR)
	}
	return this.STNodeFactory.createImplicitAnonymousFunctionParameters(openParen,
		STNodeFactory.createNodeList(paramList), bracedExpression.closeParen)
}

func (this *BallerinaParser) parseImplicitAnonFunc(openParen internal.STNode, firstParam internal.STNode, isRhsExpr bool) internal.STNode {
	paramList := make([]interface{}, 0)
	this.paramList.add(firstParam)
	nextToken := this.peek()
	var paramEnd internal.STNode
	var param internal.STNode
	for !this.isEndOfAnonFuncParametersList(nextToken.kind) {
		paramEnd = this.parseImplicitAnonFuncParamEnd()
		if paramEnd == nil {
			break
		}
		this.paramList.add(paramEnd)
		param = this.parseIdentifier(ParserRuleContext.IMPLICIT_ANON_FUNC_PARAM)
		param = this.STNodeFactory.createSimpleNameReferenceNode(param)
		this.paramList.add(param)
		nextToken = this.peek()
	}
	params := this.STNodeFactory.createNodeList(paramList)
	closeParen := this.parseCloseParenthesis()
	this.endContext()
	inferedParams := this.STNodeFactory.createImplicitAnonymousFunctionParameters(openParen, params, closeParen)
	return this.parseImplicitAnonFunc(inferedParams, isRhsExpr)
}

func (this *BallerinaParser) parseImplicitAnonFuncParamEnd() internal.STNode {
	switch peek().kind {
	case COMMA_TOKEN:
		this.parseComma()
	case CLOSE_PAREN_TOKEN:
		nil
	default:
		this.recover(peek(), ParserRuleContext.ANON_FUNC_PARAM_RHS)
		this.parseImplicitAnonFuncParamEnd()
	}
}

func (this *BallerinaParser) isEndOfAnonFuncParametersList(tokenKind SyntaxKind) bool {
	switch tokenKind {
	case EOF_TOKEN,
		CLOSE_BRACE_TOKEN,
		CLOSE_PAREN_TOKEN,
		CLOSE_BRACKET_TOKEN,
		SEMICOLON_TOKEN,
		RETURNS_KEYWORD,
		TYPE_KEYWORD,
		LISTENER_KEYWORD,
		IF_KEYWORD,
		WHILE_KEYWORD,
		DO_KEYWORD,
		OPEN_BRACE_TOKEN,
		RIGHT_DOUBLE_ARROW_TOKEN:
		true
	default:
		false
	}
}

func (this *BallerinaParser) parseTupleTypeDesc() internal.STNode {
	openBracket := this.parseOpenBracket()
	this.startContext(ParserRuleContext.TUPLE_MEMBERS)
	memberTypeDesc := this.parseTupleMemberTypeDescList()
	closeBracket := this.parseCloseBracket()
	this.endContext()
	openBracket = this.cloneWithDiagnosticIfListEmpty(memberTypeDesc, openBracket,
		DiagnosticErrorCode.ERROR_MISSING_TYPE_DESC)
	return this.STNodeFactory.createTupleTypeDescriptorNode(openBracket, memberTypeDesc, closeBracket)
}

func (this *BallerinaParser) parseTupleMemberTypeDescList() internal.STNode {
	typeDescList := make([]interface{}, 0)
	nextToken := this.peek()
	if this.isEndOfTypeList(nextToken.kind) {
		return this.STNodeFactory.createEmptyNodeList()
	}
	typeDesc := this.parseTupleMember()
	return this.parseTupleTypeMembers(typeDesc, typeDescList)
}

func (this *BallerinaParser) parseTupleTypeMembers(firstMember internal.STNode, memberList []STNode) internal.STNode {
	var tupleMemberRhs internal.STNode
	for !this.isEndOfTypeList(peek().kind) {
		if firstMember.kind == SyntaxKind.REST_TYPE {
			firstMember = this.invalidateTypeDescAfterRestDesc(firstMember)
			break
		}
		tupleMemberRhs = this.parseTupleMemberRhs()
		if tupleMemberRhs == nil {
			break
		}
		this.memberList.add(firstMember)
		this.memberList.add(tupleMemberRhs)
		firstMember = this.parseTupleMember()
	}
	this.memberList.add(firstMember)
	return this.STNodeFactory.createNodeList(memberList)
}

func (this *BallerinaParser) parseTupleMember() internal.STNode {
	annot := this.parseOptionalAnnotations()
	typeDesc := this.parseTypeDescriptor(ParserRuleContext.TYPE_DESC_IN_TUPLE)
	return this.createMemberOrRestNode(annot, typeDesc)
}

func (this *BallerinaParser) createMemberOrRestNode(annot internal.STNode, typeDesc internal.STNode) internal.STNode {
	tupleMemberRhs := this.parseTypeDescInTupleRhs()
	if tupleMemberRhs != nil {
		annotList, ok := annot.(STNodeList)
		if !ok {
			panic("createMemberOrRestNode: expected STNodeList")
		}
		if !annotList.isEmpty() {
			typeDesc = this.SyntaxErrors.cloneWithLeadingInvalidNodeMinutiae(typeDesc, annot,
				DiagnosticErrorCode.ERROR_ANNOTATIONS_NOT_ALLOWED_FOR_TUPLE_REST_DESCRIPTOR)
		}
		return this.STNodeFactory.createRestDescriptorNode(typeDesc, tupleMemberRhs)
	}
	return this.STNodeFactory.createMemberTypeDescriptorNode(annot, typeDesc)
}

func (this *BallerinaParser) invalidateTypeDescAfterRestDesc(restDescriptor internal.STNode) internal.STNode {
	for !this.isEndOfTypeList(peek().kind) {
		tupleMemberRhs := this.parseTupleMemberRhs()
		if tupleMemberRhs == nil {
			break
		}
		restDescriptor = this.SyntaxErrors.cloneWithTrailingInvalidNodeMinutiae(restDescriptor, tupleMemberRhs, null)
		restDescriptor = this.SyntaxErrors.cloneWithTrailingInvalidNodeMinutiae(restDescriptor, parseTupleMember(),
			DiagnosticErrorCode.ERROR_TYPE_DESC_AFTER_REST_DESCRIPTOR)
	}
	return restDescriptor
}

func (this *BallerinaParser) parseTupleMemberRhs() internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case COMMA_TOKEN:
		this.parseComma()
	case CLOSE_BRACKET_TOKEN:
		nil
	default:
		this.recover(nextToken, ParserRuleContext.TUPLE_TYPE_MEMBER_RHS)
		this.parseTupleMemberRhs()
	}
}

func (this *BallerinaParser) parseTypeDescInTupleRhs() internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case COMMA_TOKEN, CLOSE_BRACKET_TOKEN:
		nil
	case ELLIPSIS_TOKEN:
		this.parseEllipsis()
	default:
		this.recover(nextToken, ParserRuleContext.TYPE_DESC_IN_TUPLE_RHS)
		this.parseTypeDescInTupleRhs()
	}
}

func (this *BallerinaParser) isEndOfTypeList(nextTokenKind SyntaxKind) bool {
	switch nextTokenKind {
	case CLOSE_BRACKET_TOKEN,
		CLOSE_BRACE_TOKEN,
		CLOSE_PAREN_TOKEN,
		EOF_TOKEN,
		EQUAL_TOKEN,
		SEMICOLON_TOKEN:
		true
	default:
		false
	}
}

func (this *BallerinaParser) parseTableConstructorOrQuery(isRhsExpr bool, allowActions bool) internal.STNode {
	this.startContext(ParserRuleContext.TABLE_CONSTRUCTOR_OR_QUERY_EXPRESSION)
	tableOrQueryExpr := this.parseTableConstructorOrQueryInternal(isRhsExpr, allowActions)
	this.endContext()
	return tableOrQueryExpr
}

func (this *BallerinaParser) parseTableConstructorOrQueryInternal(isRhsExpr bool, allowActions bool) internal.STNode {
	var queryConstructType internal.STNode
	switch peek().kind {
	case FROM_KEYWORD:
		queryConstructType = this.STNodeFactory.createEmptyNode()
		return this.parseQueryExprRhs(queryConstructType, isRhsExpr, allowActions)
	case TABLE_KEYWORD:
		tableKeyword := this.parseTableKeyword()
		return this.parseTableConstructorOrQuery(tableKeyword, isRhsExpr, allowActions)
	case STREAM_KEYWORD:
	case MAP_KEYWORD:
		streamOrMapKeyword := this.consume()
		keySpecifier := this.STNodeFactory.createEmptyNode()
		queryConstructType = this.parseQueryConstructType(streamOrMapKeyword, keySpecifier)
		return this.parseQueryExprRhs(queryConstructType, isRhsExpr, allowActions)
	default:
		this.recover(peek(), ParserRuleContext.TABLE_CONSTRUCTOR_OR_QUERY_START)
		return this.parseTableConstructorOrQueryInternal(isRhsExpr, allowActions)
	}
}

func (this *BallerinaParser) parseTableConstructorOrQuery(tableKeyword internal.STNode, isRhsExpr bool, allowActions bool) internal.STNode {
	var keySpecifier internal.STNode
	nextToken := this.peek()
	switch nextToken.kind {
	case OPEN_BRACKET_TOKEN:
		keySpecifier = this.STNodeFactory.createEmptyNode()
		return this.parseTableConstructorExprRhs(tableKeyword, keySpecifier)
	case KEY_KEYWORD:
		keySpecifier = this.parseKeySpecifier()
		return this.parseTableConstructorOrQueryRhs(tableKeyword, keySpecifier, isRhsExpr, allowActions)
	case IDENTIFIER_TOKEN:
		if this.isKeyKeyword(nextToken) {
			keySpecifier = this.parseKeySpecifier()
			return this.parseTableConstructorOrQueryRhs(tableKeyword, keySpecifier, isRhsExpr, allowActions)
		}
		break
	default:
		break
	}
	this.recover(peek(), ParserRuleContext.TABLE_KEYWORD_RHS)
	return this.parseTableConstructorOrQuery(tableKeyword, isRhsExpr, allowActions)
}

func (this *BallerinaParser) parseTableConstructorOrQueryRhs(tableKeyword internal.STNode, keySpecifier internal.STNode, isRhsExpr bool, allowActions bool) internal.STNode {
	switch peek().kind {
	case FROM_KEYWORD:
		this.parseQueryExprRhs(parseQueryConstructType(tableKeyword, keySpecifier), isRhsExpr, allowActions)
	case OPEN_BRACKET_TOKEN:
		this.parseTableConstructorExprRhs(tableKeyword, keySpecifier)
	default:
		this.recover(peek(), ParserRuleContext.TABLE_CONSTRUCTOR_OR_QUERY_RHS)
		this.parseTableConstructorOrQueryRhs(tableKeyword, keySpecifier, isRhsExpr, allowActions)
	}
}

func (this *BallerinaParser) parseQueryConstructType(keyword internal.STNode, keySpecifier internal.STNode) internal.STNode {
	return this.STNodeFactory.createQueryConstructTypeNode(keyword, keySpecifier)
}

func (this *BallerinaParser) parseQueryExprRhs(queryConstructType internal.STNode, isRhsExpr bool, allowActions bool) internal.STNode {
	this.switchContext(ParserRuleContext.QUERY_EXPRESSION)
	fromClause := this.parseFromClause(isRhsExpr, allowActions)
	clauses := make([]interface{}, 0)
	var intermediateClause internal.STNode
	selectClause := nil
	collectClause := nil
	for !this.isEndOfIntermediateClause(peek().kind) {
		intermediateClause = this.parseIntermediateClause(isRhsExpr, allowActions)
		if intermediateClause == nil {
			break
		}
		if selectClause != nil {
			selectClause = this.SyntaxErrors.cloneWithTrailingInvalidNodeMinutiae(selectClause, intermediateClause,
				DiagnosticErrorCode.ERROR_MORE_CLAUSES_AFTER_SELECT_CLAUSE)
			continue
		} else if collectClause != nil {
			collectClause = this.SyntaxErrors.cloneWithTrailingInvalidNodeMinutiae(collectClause, intermediateClause,
				DiagnosticErrorCode.ERROR_MORE_CLAUSES_AFTER_COLLECT_CLAUSE)
			continue
		}
		if intermediateClause.kind == SyntaxKind.SELECT_CLAUSE {
			selectClause = intermediateClause
		} else if intermediateClause.kind == SyntaxKind.COLLECT_CLAUSE {
			collectClause = intermediateClause
		}
		if this.isNestedQueryExpr() || (!this.isValidIntermediateQueryStart(peek())) {
			break
		}
	}
	if (peek().kind == SyntaxKind.DO_KEYWORD) && ((!this.isNestedQueryExpr()) || ((selectClause == nil) && (collectClause == nil))) {
		intermediateClauses := this.STNodeFactory.createNodeList(clauses)
		queryPipeline := this.STNodeFactory.createQueryPipelineNode(fromClause, intermediateClauses)
		return this.parseQueryAction(queryConstructType, queryPipeline, selectClause, collectClause)
	}
	if (selectClause == nil) && (collectClause == nil) {
		selectKeyword := this.SyntaxErrors.createMissingToken(SyntaxKind.SELECT_KEYWORD)
		expr := this.STNodeFactory.createSimpleNameReferenceNode(SyntaxErrors.createMissingToken(SyntaxKind.IDENTIFIER_TOKEN))
		selectClause = this.STNodeFactory.createSelectClauseNode(selectKeyword, expr)
		if this.clauses.isEmpty() {
			fromClause = this.SyntaxErrors.addDiagnostic(fromClause, DiagnosticErrorCode.ERROR_MISSING_SELECT_CLAUSE)
		} else {
			lastIndex := (len(clauses) - 1)
			intClauseWithDiagnostic := this.SyntaxErrors.addDiagnostic(clauses.get(lastIndex),
				DiagnosticErrorCode.ERROR_MISSING_SELECT_CLAUSE)
			this.clauses.set(lastIndex, intClauseWithDiagnostic)
		}
	}
	intermediateClauses := this.STNodeFactory.createNodeList(clauses)
	queryPipeline := this.STNodeFactory.createQueryPipelineNode(fromClause, intermediateClauses)
	onConflictClause := this.parseOnConflictClause(isRhsExpr)
	var clause internal.STNode
	if selectClause == nil {
		clause = collectClause
	} else {
		clause = selectClause
	}
	return this.STNodeFactory.createQueryExpressionNode(queryConstructType, queryPipeline,
		clause, onConflictClause)
}

func (this *BallerinaParser) isNestedQueryExpr() bool {
	return (this.Collections.frequency(this.errorHandler.getContextStack(), ParserRuleContext.QUERY_EXPRESSION) > 1)
}

func (this *BallerinaParser) isValidIntermediateQueryStart(token internal.STToken) bool {
	switch token.kind {
	case FROM_KEYWORD,
		WHERE_KEYWORD,
		LET_KEYWORD,
		SELECT_KEYWORD,
		JOIN_KEYWORD,
		OUTER_KEYWORD,
		ORDER_KEYWORD,
		BY_KEYWORD,
		ASCENDING_KEYWORD,
		DESCENDING_KEYWORD,
		LIMIT_KEYWORD:
		true
	case IDENTIFIER_TOKEN:
		this.isGroupOrCollectKeyword(token)
	default:
		false
	}
}

func (this *BallerinaParser) parseIntermediateClause(isRhsExpr bool, allowActions bool) internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case FROM_KEYWORD:
		return this.parseFromClause(isRhsExpr, allowActions)
	case WHERE_KEYWORD:
		return this.parseWhereClause(isRhsExpr)
	case LET_KEYWORD:
		return this.parseLetClause(isRhsExpr, allowActions)
	case SELECT_KEYWORD:
		return this.parseSelectClause(isRhsExpr, allowActions)
	case JOIN_KEYWORD:
	case OUTER_KEYWORD:
		return this.parseJoinClause(isRhsExpr)
	case ORDER_KEYWORD:
	case ASCENDING_KEYWORD:
	case DESCENDING_KEYWORD:
		return this.parseOrderByClause(isRhsExpr)
	case LIMIT_KEYWORD:
		return this.parseLimitClause(isRhsExpr)
	case DO_KEYWORD:
	case SEMICOLON_TOKEN:
	case ON_KEYWORD:
	case CONFLICT_KEYWORD:
		return nil
	default:
		if this.isKeywordMatch(SyntaxKind.COLLECT_KEYWORD, nextToken) {
			return this.parseCollectClause(isRhsExpr)
		}
		if this.isKeywordMatch(SyntaxKind.GROUP_KEYWORD, nextToken) {
			return this.parseGroupByClause(isRhsExpr)
		}
		this.recover(peek(), ParserRuleContext.QUERY_PIPELINE_RHS)
		return this.parseIntermediateClause(isRhsExpr, allowActions)
	}
}

func (this *BallerinaParser) parseCollectClause(isRhsExpr bool) internal.STNode {
	this.startContext(ParserRuleContext.COLLECT_CLAUSE)
	collectKeyword := this.parseCollectKeyword()
	expression := this.parseExpression(OperatorPrecedence.QUERY, isRhsExpr, false)
	this.endContext()
	return this.STNodeFactory.createCollectClauseNode(collectKeyword, expression)
}

func (this *BallerinaParser) parseCollectKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.COLLECT_KEYWORD {
		return this.consume()
	}
	if this.isKeywordMatch(SyntaxKind.COLLECT_KEYWORD, token) {
		return this.getCollectKeyword(consume())
	}
	this.recover(token, ParserRuleContext.COLLECT_KEYWORD)
	return this.parseCollectKeyword()
}

func (this *BallerinaParser) getCollectKeyword(token internal.STToken) internal.STNode {
	return this.STNodeFactory.createToken(SyntaxKind.COLLECT_KEYWORD, token.leadingMinutiae(), token.trailingMinutiae(),
		token.diagnostics())
}

func (this *BallerinaParser) parseJoinKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.JOIN_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.JOIN_KEYWORD)
		return this.parseJoinKeyword()
	}
}

func (this *BallerinaParser) parseEqualsKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.EQUALS_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.EQUALS_KEYWORD)
		return this.parseEqualsKeyword()
	}
}

func (this *BallerinaParser) isEndOfIntermediateClause(tokenKind SyntaxKind) bool {
	switch tokenKind {
	case CLOSE_BRACE_TOKEN,
		CLOSE_PAREN_TOKEN,
		CLOSE_BRACKET_TOKEN,
		OPEN_BRACE_TOKEN,
		SEMICOLON_TOKEN,
		PUBLIC_KEYWORD,
		FUNCTION_KEYWORD,
		EOF_TOKEN,
		RESOURCE_KEYWORD,
		LISTENER_KEYWORD,
		DOCUMENTATION_STRING,
		PRIVATE_KEYWORD,
		RETURNS_KEYWORD,
		SERVICE_KEYWORD,
		TYPE_KEYWORD,
		CONST_KEYWORD,
		FINAL_KEYWORD,
		DO_KEYWORD,
		ON_KEYWORD,
		CONFLICT_KEYWORD:
		true
	default:
		this.isValidExprRhsStart(tokenKind, SyntaxKind.NONE)
	}
}

func (this *BallerinaParser) parseFromClause(isRhsExpr bool, allowActions bool) internal.STNode {
	fromKeyword := this.parseFromKeyword()
	typedBindingPattern := this.parseTypedBindingPattern(ParserRuleContext.FROM_CLAUSE)
	inKeyword := this.parseInKeyword()
	expression := this.parseExpression(OperatorPrecedence.QUERY, isRhsExpr, allowActions)
	return this.STNodeFactory.createFromClauseNode(fromKeyword, typedBindingPattern, inKeyword, expression)
}

func (this *BallerinaParser) parseFromKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.FROM_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.FROM_KEYWORD)
		return this.parseFromKeyword()
	}
}

func (this *BallerinaParser) parseWhereClause(isRhsExpr bool) internal.STNode {
	whereKeyword := this.parseWhereKeyword()
	expression := this.parseExpression(OperatorPrecedence.QUERY, isRhsExpr, false)
	return this.STNodeFactory.createWhereClauseNode(whereKeyword, expression)
}

func (this *BallerinaParser) parseWhereKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.WHERE_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.WHERE_KEYWORD)
		return this.parseWhereKeyword()
	}
}

func (this *BallerinaParser) parseLimitKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.LIMIT_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.LIMIT_KEYWORD)
		return this.parseLimitKeyword()
	}
}

func (this *BallerinaParser) parseLetClause(isRhsExpr bool, allowActions bool) internal.STNode {
	letKeyword := this.parseLetKeyword()
	letVarDeclarations := this.parseLetVarDeclarations(ParserRuleContext.LET_CLAUSE_LET_VAR_DECL, isRhsExpr,
		allowActions)
	letKeyword = this.cloneWithDiagnosticIfListEmpty(letVarDeclarations, letKeyword,
		DiagnosticErrorCode.ERROR_MISSING_LET_VARIABLE_DECLARATION)
	return this.STNodeFactory.createLetClauseNode(letKeyword, letVarDeclarations)
}

func (this *BallerinaParser) parseGroupByClause(isRhsExpr bool) internal.STNode {
	this.startContext(ParserRuleContext.GROUP_BY_CLAUSE)
	groupKeyword := this.parseGroupKeyword()
	byKeyword := this.parseByKeyword()
	groupingKeys := this.parseGroupingKeyList(isRhsExpr)
	byKeyword = this.cloneWithDiagnosticIfListEmpty(groupingKeys, byKeyword,
		DiagnosticErrorCode.ERROR_MISSING_GROUPING_KEY)
	this.endContext()
	return this.STNodeFactory.createGroupByClauseNode(groupKeyword, byKeyword, groupingKeys)
}

func (this *BallerinaParser) parseGroupKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.GROUP_KEYWORD {
		return this.consume()
	}
	if this.isKeywordMatch(SyntaxKind.GROUP_KEYWORD, token) {
		return this.getGroupKeyword(consume())
	}
	this.recover(token, ParserRuleContext.GROUP_KEYWORD)
	return this.parseGroupKeyword()
}

func (this *BallerinaParser) getGroupKeyword(token internal.STToken) internal.STNode {
	return this.STNodeFactory.createToken(SyntaxKind.GROUP_KEYWORD, token.leadingMinutiae(), token.trailingMinutiae(),
		token.diagnostics())
}

func (this *BallerinaParser) parseOrderKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.ORDER_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.ORDER_KEYWORD)
		return this.parseOrderKeyword()
	}
}

func (this *BallerinaParser) parseByKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.BY_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.BY_KEYWORD)
		return this.parseByKeyword()
	}
}

func (this *BallerinaParser) parseOrderByClause(isRhsExpr bool) internal.STNode {
	orderKeyword := this.parseOrderKeyword()
	byKeyword := this.parseByKeyword()
	orderKeys := this.parseOrderKeyList(isRhsExpr)
	byKeyword = this.cloneWithDiagnosticIfListEmpty(orderKeys, byKeyword, DiagnosticErrorCode.ERROR_MISSING_ORDER_KEY)
	return this.STNodeFactory.createOrderByClauseNode(orderKeyword, byKeyword, orderKeys)
}

func (this *BallerinaParser) parseGroupingKeyList(isRhsExpr bool) internal.STNode {
	groupingKeys := make([]interface{}, 0)
	nextToken := this.peek()
	if this.isEndOfGroupByKeyListElement(nextToken) {
		return this.STNodeFactory.createEmptyNodeList()
	}
	groupingKey := this.parseGroupingKey(isRhsExpr)
	this.groupingKeys.add(groupingKey)
	nextToken = this.peek()
	var groupingKeyListMemberEnd internal.STNode
	for !this.isEndOfGroupByKeyListElement(nextToken) {
		groupingKeyListMemberEnd = this.parseGroupingKeyListMemberEnd()
		if groupingKeyListMemberEnd == nil {
			break
		}
		this.groupingKeys.add(groupingKeyListMemberEnd)
		groupingKey = this.parseGroupingKey(isRhsExpr)
		this.groupingKeys.add(groupingKey)
		nextToken = this.peek()
	}
	return this.STNodeFactory.createNodeList(groupingKeys)
}

func (this *BallerinaParser) parseOrderKeyList(isRhsExpr bool) internal.STNode {
	this.startContext(ParserRuleContext.ORDER_KEY_LIST)
	orderKeys := make([]interface{}, 0)
	nextToken := this.peek()
	if this.isEndOfOrderKeys(nextToken) {
		this.endContext()
		return this.STNodeFactory.createEmptyNodeList()
	}
	orderKey := this.parseOrderKey(isRhsExpr)
	this.orderKeys.add(orderKey)
	nextToken = this.peek()
	var orderKeyListMemberEnd internal.STNode
	for !this.isEndOfOrderKeys(nextToken) {
		orderKeyListMemberEnd = this.parseOrderKeyListMemberEnd()
		if orderKeyListMemberEnd == nil {
			break
		}
		this.orderKeys.add(orderKeyListMemberEnd)
		orderKey = this.parseOrderKey(isRhsExpr)
		this.orderKeys.add(orderKey)
		nextToken = this.peek()
	}
	this.endContext()
	return this.STNodeFactory.createNodeList(orderKeys)
}

func (this *BallerinaParser) isEndOfGroupByKeyListElement(nextToken internal.STToken) bool {
	switch nextToken.kind {
	case COMMA_TOKEN:
		false
	case EOF_TOKEN:
		true
	default:
		this.isQueryClauseStartToken(nextToken)
	}
}

func (this *BallerinaParser) isEndOfOrderKeys(nextToken internal.STToken) bool {
	switch nextToken.kind {
	case COMMA_TOKEN,
		ASCENDING_KEYWORD,
		DESCENDING_KEYWORD:
		false
	case SEMICOLON_TOKEN,
		EOF_TOKEN:
		true
	default:
		this.isQueryClauseStartToken(nextToken)
	}
}

func (this *BallerinaParser) isQueryClauseStartToken(nextToken internal.STToken) bool {
	switch nextToken.kind {
	case SELECT_KEYWORD,
		LET_KEYWORD,
		WHERE_KEYWORD,
		OUTER_KEYWORD,
		JOIN_KEYWORD,
		ORDER_KEYWORD,
		DO_KEYWORD,
		FROM_KEYWORD,
		LIMIT_KEYWORD:
		true
	case IDENTIFIER_TOKEN:
		this.isGroupOrCollectKeyword(nextToken)
	default:
		false
	}
}

func (this *BallerinaParser) parseGroupingKeyListMemberEnd() internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case COMMA_TOKEN:
		return this.consume()
	case EOF_TOKEN:
		return nil
	default:
		if this.isQueryClauseStartToken(nextToken) {
			return nil
		}
		this.recover(peek(), ParserRuleContext.GROUPING_KEY_LIST_ELEMENT_END)
		return this.parseGroupingKeyListMemberEnd()
	}
}

func (this *BallerinaParser) parseOrderKeyListMemberEnd() internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case COMMA_TOKEN:
		return this.parseComma()
	case EOF_TOKEN:
		return nil
	default:
		if this.isQueryClauseStartToken(nextToken) {
			return nil
		}
		this.recover(peek(), ParserRuleContext.ORDER_KEY_LIST_END)
		return this.parseOrderKeyListMemberEnd()
	}
}

func (this *BallerinaParser) parseGroupingKeyVariableDeclaration(isRhsExpr bool) internal.STNode {
	groupingKeyElementTypeDesc := this.parseTypeDescriptor(ParserRuleContext.TYPE_DESC_BEFORE_IDENTIFIER_IN_GROUPING_KEY)
	this.startContext(ParserRuleContext.BINDING_PATTERN_STARTING_IDENTIFIER)
	groupingKeySimpleBP := this.createCaptureOrWildcardBP(parseVariableName())
	this.endContext()
	equalsToken := this.parseAssignOp()
	groupingKeyExpression := this.parseExpression(OperatorPrecedence.QUERY, isRhsExpr, false)
	return this.STNodeFactory.createGroupingKeyVarDeclarationNode(groupingKeyElementTypeDesc, groupingKeySimpleBP,
		equalsToken, groupingKeyExpression)
}

func (this *BallerinaParser) parseGroupingKey(isRhsExpr bool) internal.STNode {
	nextToken := this.peek()
	nextTokenKind := nextToken.kind
	if (nextTokenKind == SyntaxKind.IDENTIFIER_TOKEN) && (!this.isPossibleGroupingKeyVarDeclaration()) {
		return this.STNodeFactory.createSimpleNameReferenceNode(parseVariableName())
	} else if this.isTypeStartingToken(nextTokenKind, nextToken) {
		return this.parseGroupingKeyVariableDeclaration(isRhsExpr)
	}
	this.recover(nextToken, ParserRuleContext.GROUPING_KEY_LIST_ELEMENT)
	return this.parseGroupingKey(isRhsExpr)
}

func (this *BallerinaParser) isPossibleGroupingKeyVarDeclaration() bool {
	nextNextTokenKind := getNextNextToken().kind
	return ((nextNextTokenKind == SyntaxKind.EQUAL_TOKEN) || ((nextNextTokenKind == SyntaxKind.IDENTIFIER_TOKEN) && (peek(3).kind == SyntaxKind.EQUAL_TOKEN)))
}

func (this *BallerinaParser) parseOrderKey(isRhsExpr bool) internal.STNode {
	expression := this.parseExpression(OperatorPrecedence.QUERY, isRhsExpr, false)
	var orderDirection internal.STNode
	nextToken := this.peek()
	switch nextToken.kind {
	case ASCENDING_KEYWORD, DESCENDING_KEYWORD:
		orderDirection = this.consume()
	default:
		orderDirection = this.STNodeFactory.createEmptyNode()
	}
	return this.STNodeFactory.createOrderKeyNode(expression, orderDirection)
}

func (this *BallerinaParser) parseSelectClause(isRhsExpr bool, allowActions bool) internal.STNode {
	this.startContext(ParserRuleContext.SELECT_CLAUSE)
	selectKeyword := this.parseSelectKeyword()
	expression := this.parseExpression(OperatorPrecedence.QUERY, isRhsExpr, allowActions)
	this.endContext()
	return this.STNodeFactory.createSelectClauseNode(selectKeyword, expression)
}

func (this *BallerinaParser) parseSelectKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.SELECT_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.SELECT_KEYWORD)
		return this.parseSelectKeyword()
	}
}

func (this *BallerinaParser) parseOnConflictClause(isRhsExpr bool) internal.STNode {
	nextToken := this.peek()
	if (nextToken.kind != SyntaxKind.ON_KEYWORD) && (nextToken.kind != SyntaxKind.CONFLICT_KEYWORD) {
		return this.STNodeFactory.createEmptyNode()
	}
	this.startContext(ParserRuleContext.ON_CONFLICT_CLAUSE)
	onKeyword := this.parseOnKeyword()
	conflictKeyword := this.parseConflictKeyword()
	this.endContext()
	expr := this.parseExpression(OperatorPrecedence.QUERY, isRhsExpr, false)
	return this.STNodeFactory.createOnConflictClauseNode(onKeyword, conflictKeyword, expr)
}

func (this *BallerinaParser) parseConflictKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.CONFLICT_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.CONFLICT_KEYWORD)
		return this.parseConflictKeyword()
	}
}

func (this *BallerinaParser) parseLimitClause(isRhsExpr bool) internal.STNode {
	limitKeyword := this.parseLimitKeyword()
	expr := this.parseExpression(OperatorPrecedence.QUERY, isRhsExpr, false)
	return this.STNodeFactory.createLimitClauseNode(limitKeyword, expr)
}

func (this *BallerinaParser) parseJoinClause(isRhsExpr bool) internal.STNode {
	this.startContext(ParserRuleContext.JOIN_CLAUSE)
	var outerKeyword internal.STNode
	nextToken := this.peek()
	if nextToken.kind == SyntaxKind.OUTER_KEYWORD {
		outerKeyword = this.consume()
	} else {
		outerKeyword = this.STNodeFactory.createEmptyNode()
	}
	joinKeyword := this.parseJoinKeyword()
	typedBindingPattern := this.parseTypedBindingPattern(ParserRuleContext.JOIN_CLAUSE)
	inKeyword := this.parseInKeyword()
	expression := this.parseExpression(OperatorPrecedence.QUERY, isRhsExpr, false)
	this.endContext()
	onCondition := this.parseOnClause(isRhsExpr)
	return this.STNodeFactory.createJoinClauseNode(outerKeyword, joinKeyword, typedBindingPattern, inKeyword, expression,
		onCondition)
}

func (this *BallerinaParser) parseOnClause(isRhsExpr bool) internal.STNode {
	nextToken := this.peek()
	if this.isQueryClauseStartToken(nextToken) {
		return this.createMissingOnClauseNode()
	}
	this.startContext(ParserRuleContext.ON_CLAUSE)
	onKeyword := this.parseOnKeyword()
	lhsExpression := this.parseExpression(OperatorPrecedence.QUERY, isRhsExpr, false)
	equalsKeyword := this.parseEqualsKeyword()
	this.endContext()
	rhsExpression := this.parseExpression(OperatorPrecedence.QUERY, isRhsExpr, false)
	return this.STNodeFactory.createOnClauseNode(onKeyword, lhsExpression, equalsKeyword, rhsExpression)
}

func (this *BallerinaParser) createMissingOnClauseNode() internal.STNode {
	onKeyword := this.SyntaxErrors.createMissingTokenWithDiagnostics(SyntaxKind.ON_KEYWORD,
		DiagnosticErrorCode.ERROR_MISSING_ON_KEYWORD)
	identifier := this.SyntaxErrors.createMissingTokenWithDiagnostics(SyntaxKind.IDENTIFIER_TOKEN,
		DiagnosticErrorCode.ERROR_MISSING_IDENTIFIER)
	equalsKeyword := this.SyntaxErrors.createMissingTokenWithDiagnostics(SyntaxKind.EQUALS_KEYWORD,
		DiagnosticErrorCode.ERROR_MISSING_EQUALS_KEYWORD)
	lhsExpression := this.STNodeFactory.createSimpleNameReferenceNode(identifier)
	rhsExpression := this.STNodeFactory.createSimpleNameReferenceNode(identifier)
	return this.STNodeFactory.createOnClauseNode(onKeyword, lhsExpression, equalsKeyword, rhsExpression)
}

func (this *BallerinaParser) parseStartAction(annots internal.STNode) internal.STNode {
	startKeyword := this.parseStartKeyword()
	expr := this.parseActionOrExpression()
	switch expr.kind {
	case FUNCTION_CALL:
	case METHOD_CALL:
	case REMOTE_METHOD_CALL_ACTION:
		break
	case SIMPLE_NAME_REFERENCE:
	case QUALIFIED_NAME_REFERENCE:
	case FIELD_ACCESS:
	case ASYNC_SEND_ACTION:
		expr = this.generateValidExprForStartAction(expr)
		break
	default:
		startKeyword = this.SyntaxErrors.cloneWithTrailingInvalidNodeMinutiae(startKeyword, expr,
			DiagnosticErrorCode.ERROR_INVALID_EXPRESSION_IN_START_ACTION)
		funcName := this.SyntaxErrors.createMissingToken(SyntaxKind.IDENTIFIER_TOKEN)
		funcName = this.STNodeFactory.createSimpleNameReferenceNode(funcName)
		openParenToken := this.SyntaxErrors.createMissingToken(SyntaxKind.OPEN_PAREN_TOKEN)
		closeParenToken := this.SyntaxErrors.createMissingToken(SyntaxKind.CLOSE_PAREN_TOKEN)
		expr = this.STNodeFactory.createFunctionCallExpressionNode(funcName, openParenToken,
			STNodeFactory.createEmptyNodeList(), closeParenToken)
		break
	}
	return this.STNodeFactory.createStartActionNode(getAnnotations(annots), startKeyword, expr)
}

func (this *BallerinaParser) generateValidExprForStartAction(expr internal.STNode) internal.STNode {
	openParenToken := this.SyntaxErrors.createMissingTokenWithDiagnostics(SyntaxKind.OPEN_PAREN_TOKEN,
		DiagnosticErrorCode.ERROR_MISSING_OPEN_PAREN_TOKEN)
	arguments := this.STNodeFactory.createEmptyNodeList()
	closeParenToken := this.SyntaxErrors.createMissingTokenWithDiagnostics(SyntaxKind.CLOSE_PAREN_TOKEN,
		DiagnosticErrorCode.ERROR_MISSING_CLOSE_PAREN_TOKEN)
	switch expr.kind {
	case FIELD_ACCESS:
		fieldAccessExpr := internal.STFieldAccessExpressionNode(expr)
		this.STNodeFactory.createMethodCallExpressionNode(fieldAccessExpr.expression,
			fieldAccessExpr.dotToken, fieldAccessExpr.fieldName, openParenToken, arguments,
			closeParenToken)
	case ASYNC_SEND_ACTION:
		asyncSendAction := internal.STAsyncSendActionNode(expr)
		this.STNodeFactory.createRemoteMethodCallActionNode(asyncSendAction.expression,
			asyncSendAction.rightArrowToken, asyncSendAction.peerWorker, openParenToken, arguments,
			closeParenToken)
	default:
		this.STNodeFactory.createFunctionCallExpressionNode(expr, openParenToken, arguments, closeParenToken)
	}
}

func (this *BallerinaParser) parseStartKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.START_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.START_KEYWORD)
		return this.parseStartKeyword()
	}
}

func (this *BallerinaParser) parseFlushAction() internal.STNode {
	flushKeyword := this.parseFlushKeyword()
	peerWorker := this.parseOptionalPeerWorkerName()
	return this.STNodeFactory.createFlushActionNode(flushKeyword, peerWorker)
}

func (this *BallerinaParser) parseFlushKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.FLUSH_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.FLUSH_KEYWORD)
		return this.parseFlushKeyword()
	}
}

func (this *BallerinaParser) parseOptionalPeerWorkerName() internal.STNode {
	token := this.peek()
	switch token.kind {
	case IDENTIFIER_TOKEN, FUNCTION_KEYWORD:
		this.STNodeFactory.createSimpleNameReferenceNode(consume())
	default:
		this.STNodeFactory.createEmptyNode()
	}
}

func (this *BallerinaParser) parseIntersectionTypeDescriptor(leftTypeDesc internal.STNode, context ParserRuleContext, isTypedBindingPattern bool) internal.STNode {
	bitwiseAndToken := this.consume()
	rightTypeDesc := this.parseTypeDescriptorInternal(nil, context, isTypedBindingPattern, false,
		TypePrecedence.INTERSECTION)
	return this.mergeTypesWithIntersection(leftTypeDesc, bitwiseAndToken, rightTypeDesc)
}

func (this *BallerinaParser) createIntersectionTypeDesc(leftTypeDesc internal.STNode, bitwiseAndToken internal.STNode, rightTypeDesc internal.STNode) internal.STNode {
	leftTypeDesc = this.validateForUsageOfVar(leftTypeDesc)
	rightTypeDesc = this.validateForUsageOfVar(rightTypeDesc)
	return this.STNodeFactory.createIntersectionTypeDescriptorNode(leftTypeDesc, bitwiseAndToken, rightTypeDesc)
}

func (this *BallerinaParser) parseSingletonTypeDesc() internal.STNode {
	simpleContExpr := this.parseSimpleConstExpr()
	return this.STNodeFactory.createSingletonTypeDescriptorNode(simpleContExpr)
}

func (this *BallerinaParser) parseSignedIntOrFloat() internal.STNode {
	operator := this.parseUnaryOperator()
	var literal internal.STNode
	nextToken := this.peek()

	switch nextToken.kind {

	case HEX_INTEGER_LITERAL_TOKEN,
		DECIMAL_FLOATING_POINT_LITERAL_TOKEN,
		HEX_FLOATING_POINT_LITERAL_TOKEN:
		literal = this.parseBasicLiteral()
	default:
		literal = this.STNodeFactory.createBasicLiteralNode(SyntaxKind.NUMERIC_LITERAL,
			parseDecimalIntLiteral(ParserRuleContext.DECIMAL_INTEGER_LITERAL_TOKEN))
	}
	return this.STNodeFactory.createUnaryExpressionNode(operator, literal)
}

func (this *BallerinaParser) isValidExpressionStart(nextTokenKind SyntaxKind, nextTokenIndex int) bool {
	nextTokenIndex++
	switch nextTokenKind {
	case DECIMAL_INTEGER_LITERAL_TOKEN:
	case HEX_INTEGER_LITERAL_TOKEN:
	case STRING_LITERAL_TOKEN:
	case NULL_KEYWORD:
	case TRUE_KEYWORD:
	case FALSE_KEYWORD:
	case DECIMAL_FLOATING_POINT_LITERAL_TOKEN:
	case HEX_FLOATING_POINT_LITERAL_TOKEN:
		nextNextTokenKind := peek(nextTokenIndex).kind
		if (nextNextTokenKind == SyntaxKind.PIPE_TOKEN) || (nextNextTokenKind == SyntaxKind.BITWISE_AND_TOKEN) {
			nextTokenIndex++
			return this.isValidExpressionStart(peek(nextTokenIndex).kind, nextTokenIndex)
		}
		return ((((nextNextTokenKind == SyntaxKind.SEMICOLON_TOKEN) || (nextNextTokenKind == SyntaxKind.COMMA_TOKEN)) || (nextNextTokenKind == SyntaxKind.CLOSE_BRACKET_TOKEN)) || this.isValidExprRhsStart(nextNextTokenKind, SyntaxKind.SIMPLE_NAME_REFERENCE))
	case IDENTIFIER_TOKEN:
		return this.isValidExprRhsStart(peek(nextTokenIndex).kind, SyntaxKind.SIMPLE_NAME_REFERENCE)
	case OPEN_PAREN_TOKEN:
	case CHECK_KEYWORD:
	case CHECKPANIC_KEYWORD:
	case OPEN_BRACE_TOKEN:
	case TYPEOF_KEYWORD:
	case NEGATION_TOKEN:
	case EXCLAMATION_MARK_TOKEN:
	case TRAP_KEYWORD:
	case OPEN_BRACKET_TOKEN:
	case LT_TOKEN:
	case FROM_KEYWORD:
	case LET_KEYWORD:
	case BACKTICK_TOKEN:
	case NEW_KEYWORD:
	case LEFT_ARROW_TOKEN:
	case FUNCTION_KEYWORD:
	case TRANSACTIONAL_KEYWORD:
	case ISOLATED_KEYWORD:
	case BASE16_KEYWORD:
	case BASE64_KEYWORD:
	case NATURAL_KEYWORD:
		return true
	case PLUS_TOKEN:
	case MINUS_TOKEN:
		return this.isValidExpressionStart(peek(nextTokenIndex).kind, nextTokenIndex)
	case TABLE_KEYWORD:
	case MAP_KEYWORD:
		return (peek(nextTokenIndex).kind == SyntaxKind.FROM_KEYWORD)
	case STREAM_KEYWORD:
		nextNextToken := this.peek(nextTokenIndex)
		return (((nextNextToken.kind == SyntaxKind.KEY_KEYWORD) || (nextNextToken.kind == SyntaxKind.OPEN_BRACKET_TOKEN)) || (nextNextToken.kind == SyntaxKind.FROM_KEYWORD))
	case ERROR_KEYWORD:
		return (peek(nextTokenIndex).kind == SyntaxKind.OPEN_PAREN_TOKEN)
	case XML_KEYWORD:
	case STRING_KEYWORD:
	case RE_KEYWORD:
		return (peek(nextTokenIndex).kind == SyntaxKind.BACKTICK_TOKEN)
	case START_KEYWORD:
	case FLUSH_KEYWORD:
	case WAIT_KEYWORD:
	default:
		return false
	}
}

func (this *BallerinaParser) parseSyncSendAction(expression internal.STNode) internal.STNode {
	syncSendToken := this.parseSyncSendToken()
	peerWorker := this.parsePeerWorkerName()
	return this.STNodeFactory.createSyncSendActionNode(expression, syncSendToken, peerWorker)
}

func (this *BallerinaParser) parsePeerWorkerName() internal.STNode {
	token := this.peek()
	switch token.kind {
	case IDENTIFIER_TOKEN,
		FUNCTION_KEYWORD:
		this.STNodeFactory.createSimpleNameReferenceNode(consume())
	default:
		this.recover(token, ParserRuleContext.PEER_WORKER_NAME)
		this.parsePeerWorkerName()
	}
}

func (this *BallerinaParser) parseSyncSendToken() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.SYNC_SEND_TOKEN {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.SYNC_SEND_TOKEN)
		return this.parseSyncSendToken()
	}
}

func (this *BallerinaParser) parseReceiveAction() internal.STNode {
	leftArrow := this.parseLeftArrowToken()
	receiveWorkers := this.parseReceiveWorkers()
	return this.STNodeFactory.createReceiveActionNode(leftArrow, receiveWorkers)
}

func (this *BallerinaParser) parseReceiveWorkers() internal.STNode {
	switch peek().kind {
	case FUNCTION_KEYWORD, IDENTIFIER_TOKEN:
		this.parseSingleOrAlternateReceiveWorkers()
	case OPEN_BRACE_TOKEN:
		this.parseMultipleReceiveWorkers()
	default:
		this.recover(peek(), ParserRuleContext.RECEIVE_WORKERS)
		this.parseReceiveWorkers()
	}
}

func (this *BallerinaParser) parseSingleOrAlternateReceiveWorkers() internal.STNode {
	this.startContext(ParserRuleContext.SINGLE_OR_ALTERNATE_WORKER)
	workers := make([]interface{}, 0)
	peerWorker := this.parsePeerWorkerName()
	this.workers.add(peerWorker)
	nextToken := this.peek()
	if nextToken.kind != SyntaxKind.PIPE_TOKEN {
		this.endContext()
		return peerWorker
	}
	for nextToken.kind == SyntaxKind.PIPE_TOKEN {
		pipeToken := this.consume()
		this.workers.add(pipeToken)
		peerWorker = this.parsePeerWorkerName()
		this.workers.add(peerWorker)
		nextToken = this.peek()
	}
	this.endContext()
	return this.STNodeFactory.createAlternateReceiveNode(STNodeFactory.createNodeList(workers))
}

func (this *BallerinaParser) parseMultipleReceiveWorkers() internal.STNode {
	this.startContext(ParserRuleContext.MULTI_RECEIVE_WORKERS)
	openBrace := this.parseOpenBrace()
	receiveFields := this.parseReceiveFields()
	closeBrace := this.parseCloseBrace()
	this.endContext()
	openBrace = this.cloneWithDiagnosticIfListEmpty(receiveFields, openBrace,
		DiagnosticErrorCode.ERROR_MISSING_RECEIVE_FIELD_IN_RECEIVE_ACTION)
	return this.STNodeFactory.createReceiveFieldsNode(openBrace, receiveFields, closeBrace)
}

func (this *BallerinaParser) parseReceiveFields() internal.STNode {
	receiveFields := make([]interface{}, 0)
	nextToken := this.peek()
	if this.isEndOfReceiveFields(nextToken.kind) {
		return this.STNodeFactory.createEmptyNodeList()
	}
	receiveField := this.parseReceiveField()
	this.receiveFields.add(receiveField)
	nextToken = this.peek()
	var recieveFieldEnd internal.STNode
	for !this.isEndOfReceiveFields(nextToken.kind) {
		recieveFieldEnd = this.parseReceiveFieldEnd()
		if recieveFieldEnd == nil {
			break
		}
		this.receiveFields.add(recieveFieldEnd)
		receiveField = this.parseReceiveField()
		this.receiveFields.add(receiveField)
		nextToken = this.peek()
	}
	return this.STNodeFactory.createNodeList(receiveFields)
}

func (this *BallerinaParser) isEndOfReceiveFields(nextTokenKind SyntaxKind) bool {
	switch nextTokenKind {
	case EOF_TOKEN, CLOSE_BRACE_TOKEN:
		true
	default:
		false
	}
}

func (this *BallerinaParser) parseReceiveFieldEnd() internal.STNode {
	switch peek().kind {
	case COMMA_TOKEN:
		this.parseComma()
	case CLOSE_BRACE_TOKEN:
		nil
	default:
		this.recover(peek(), ParserRuleContext.RECEIVE_FIELD_END)
		this.parseReceiveFieldEnd()
	}
}

func (this *BallerinaParser) parseReceiveField() internal.STNode {
	switch peek().kind {
	case FUNCTION_KEYWORD:
		functionKeyword := this.consume()
		this.STNodeFactory.createSimpleNameReferenceNode(functionKeyword)
	case IDENTIFIER_TOKEN:
		identifier := this.parseIdentifier(ParserRuleContext.RECEIVE_FIELD_NAME)
		this.createReceiveField(identifier)
	default:
		this.recover(peek(), ParserRuleContext.RECEIVE_FIELD)
		this.parseReceiveField()
	}
}

func (this *BallerinaParser) createReceiveField(identifier internal.STNode) internal.STNode {
	if peek().kind != SyntaxKind.COLON_TOKEN {
		return this.STNodeFactory.createSimpleNameReferenceNode(identifier)
	}
	identifier = this.STNodeFactory.createSimpleNameReferenceNode(identifier)
	colon := this.parseColon()
	peerWorker := this.parsePeerWorkerName()
	return this.STNodeFactory.createReceiveFieldNode(identifier, colon, peerWorker)
}

func (this *BallerinaParser) parseLeftArrowToken() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.LEFT_ARROW_TOKEN {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.LEFT_ARROW_TOKEN)
		return this.parseLeftArrowToken()
	}
}

func (this *BallerinaParser) parseSignedRightShiftToken() internal.STNode {
	firstToken := this.consume()
	if firstToken.kind == SyntaxKind.DOUBLE_GT_TOKEN {
		return firstToken
	}
	endLGToken := this.consume()
	doubleGTToken := this.STNodeFactory.createToken(SyntaxKind.DOUBLE_GT_TOKEN, firstToken.leadingMinutiae(),
		endLGToken.trailingMinutiae())
	if this.hasTrailingMinutiae(firstToken) {
		doubleGTToken = this.SyntaxErrors.addDiagnostic(doubleGTToken,
			DiagnosticErrorCode.ERROR_NO_WHITESPACES_ALLOWED_IN_RIGHT_SHIFT_OP)
	}
	return doubleGTToken
}

func (this *BallerinaParser) parseUnsignedRightShiftToken() internal.STNode {
	firstToken := this.consume()
	if firstToken.kind == SyntaxKind.TRIPPLE_GT_TOKEN {
		return firstToken
	}
	middleGTToken := this.consume()
	endLGToken := this.consume()
	unsignedRightShiftToken := this.STNodeFactory.createToken(SyntaxKind.TRIPPLE_GT_TOKEN,
		firstToken.leadingMinutiae(), endLGToken.trailingMinutiae())
	validOpenGTToken := (!this.hasTrailingMinutiae(firstToken))
	validMiddleGTToken := (!this.hasTrailingMinutiae(middleGTToken))
	if validOpenGTToken && validMiddleGTToken {
		return unsignedRightShiftToken
	}
	unsignedRightShiftToken = this.SyntaxErrors.addDiagnostic(unsignedRightShiftToken,
		DiagnosticErrorCode.ERROR_NO_WHITESPACES_ALLOWED_IN_UNSIGNED_RIGHT_SHIFT_OP)
	return unsignedRightShiftToken
}

func (this *BallerinaParser) parseWaitAction() internal.STNode {
	waitKeyword := this.parseWaitKeyword()
	if peek().kind == OPEN_BRACE_TOKEN {
		return this.parseMultiWaitAction(waitKeyword)
	}
	return this.parseSingleOrAlternateWaitAction(waitKeyword)
}

func (this *BallerinaParser) parseWaitKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.WAIT_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.WAIT_KEYWORD)
		return this.parseWaitKeyword()
	}
}

// func (this *BallerinaParser) parseSingleOrAlternateWaitAction(waitKeyword internal.STNode) internal.STNode {
// 	this.startContext(ParserRuleContext.ALTERNATE_WAIT_EXPRS)
// 	nextToken := this.peek()
// 	if this.isEndOfWaitFutureExprList(nextToken.kind) {
// 		this.endContext()
// 		waitFutureExprs := this.STNodeFactory
// 							.createSimpleNameReferenceNode(STNodeFactory.createMissingToken(SyntaxKind.IDENTIFIER_TOKEN))
// 		waitFutureExprs = this.SyntaxErrors.addDiagnostic(waitFutureExprs,
// 							DiagnosticErrorCode.ERROR_MISSING_WAIT_FUTURE_EXPRESSION)
// 		return this.STNodeFactory.createWaitActionNode(waitKeyword, waitFutureExprs)
// 	}
// 	var waitFutureExprs []internal.STNode
// 	waitField := this.parseWaitFutureExpr()
// 	this.waitFutureExprList.add(waitField)
// 	nextToken = this.peek()
// 	var waitFutureExprEnd internal.STNode
// 	for !this.isEndOfWaitFutureExprList(nextToken.kind) {
// 	waitFutureExprEnd = this.parseWaitFutureExprEnd()
// 	if (waitFutureExprEnd == nil) {
// 	break;
// 	}
// 	this.waitFutureExprList.add(waitFutureExprEnd)
// 	waitField = this.parseWaitFutureExpr()
// 	this.waitFutureExprList.add(waitField)
// 	nextToken = this.peek()
// 	}
// 	this.endContext()
// 	return this.STNodeFactory.createWaitActionNode(waitKeyword, waitFutureExprList.get(0))
// }

func (this *BallerinaParser) isEndOfWaitFutureExprList(nextTokenKind SyntaxKind) bool {
	switch nextTokenKind {
	case EOF_TOKEN, CLOSE_BRACE_TOKEN, SEMICOLON_TOKEN, OPEN_BRACE_TOKEN:
		true
	default:
		false
	}
}

func (this *BallerinaParser) parseWaitFutureExpr() internal.STNode {
	waitFutureExpr := this.parseActionOrExpression()
	if waitFutureExpr.kind == SyntaxKind.MAPPING_CONSTRUCTOR {
		waitFutureExpr = this.SyntaxErrors.addDiagnostic(waitFutureExpr,
			DiagnosticErrorCode.ERROR_MAPPING_CONSTRUCTOR_EXPR_AS_A_WAIT_EXPR)
	} else if this.isAction(waitFutureExpr) {
		waitFutureExpr = this.SyntaxErrors.addDiagnostic(waitFutureExpr, DiagnosticErrorCode.ERROR_ACTION_AS_A_WAIT_EXPR)
	}
	return waitFutureExpr
}

func (this *BallerinaParser) parseWaitFutureExprEnd() internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case PIPE_TOKEN:
		return this.parsePipeToken()
	default:
		if this.isEndOfWaitFutureExprList(nextToken.kind) || (!this.isValidExpressionStart(nextToken.kind, 1)) {
			return nil
		}
		this.recover(peek(), ParserRuleContext.WAIT_FUTURE_EXPR_END)
		return this.parseWaitFutureExprEnd()
	}
}

func (this *BallerinaParser) parseMultiWaitAction(waitKeyword internal.STNode) internal.STNode {
	this.startContext(ParserRuleContext.MULTI_WAIT_FIELDS)
	openBrace := this.parseOpenBrace()
	waitFields := this.parseWaitFields()
	closeBrace := this.parseCloseBrace()
	this.endContext()
	openBrace = this.cloneWithDiagnosticIfListEmpty(waitFields, openBrace,
		DiagnosticErrorCode.ERROR_MISSING_WAIT_FIELD_IN_WAIT_ACTION)
	waitFieldsNode := this.STNodeFactory.createWaitFieldsListNode(openBrace, waitFields, closeBrace)
	return this.STNodeFactory.createWaitActionNode(waitKeyword, waitFieldsNode)
}

func (this *BallerinaParser) parseWaitFields() internal.STNode {
	waitFields := make([]interface{}, 0)
	nextToken := this.peek()
	if this.isEndOfWaitFields(nextToken.kind) {
		return this.STNodeFactory.createEmptyNodeList()
	}
	waitField := this.parseWaitField()
	this.waitFields.add(waitField)
	nextToken = this.peek()
	var waitFieldEnd internal.STNode
	for !this.isEndOfWaitFields(nextToken.kind) {
		waitFieldEnd = this.parseWaitFieldEnd()
		if waitFieldEnd == nil {
			break
		}
		this.waitFields.add(waitFieldEnd)
		waitField = this.parseWaitField()
		this.waitFields.add(waitField)
		nextToken = this.peek()
	}
	return this.STNodeFactory.createNodeList(waitFields)
}

func (this *BallerinaParser) isEndOfWaitFields(nextTokenKind SyntaxKind) bool {
	switch nextTokenKind {
	case EOF_TOKEN, CLOSE_BRACE_TOKEN:
		true
	default:
		false
	}
}

func (this *BallerinaParser) parseWaitFieldEnd() internal.STNode {
	switch peek().kind {
	case COMMA_TOKEN:
		this.parseComma()
	case CLOSE_BRACE_TOKEN:
		nil
	default:
		this.recover(peek(), ParserRuleContext.WAIT_FIELD_END)
		this.parseWaitFieldEnd()
	}
}

func (this *BallerinaParser) parseWaitField() internal.STNode {
	switch peek().kind {
	case IDENTIFIER_TOKEN:
		identifier := this.parseIdentifier(ParserRuleContext.WAIT_FIELD_NAME)
		identifier = this.STNodeFactory.createSimpleNameReferenceNode(identifier)
		return this.createQualifiedWaitField(identifier)
	default:
		this.recover(peek(), ParserRuleContext.WAIT_FIELD_NAME)
		return this.parseWaitField()
	}
}

func (this *BallerinaParser) createQualifiedWaitField(identifier internal.STNode) internal.STNode {
	if peek().kind != SyntaxKind.COLON_TOKEN {
		return identifier
	}
	colon := this.parseColon()
	waitFutureExpr := this.parseWaitFutureExpr()
	return this.STNodeFactory.createWaitFieldNode(identifier, colon, waitFutureExpr)
}

func (this *BallerinaParser) parseAnnotAccessExpression(lhsExpr internal.STNode, isInConditionalExpr bool) internal.STNode {
	annotAccessToken := this.parseAnnotChainingToken()
	annotTagReference := this.parseFieldAccessIdentifier(isInConditionalExpr)
	return this.STNodeFactory.createAnnotAccessExpressionNode(lhsExpr, annotAccessToken, annotTagReference)
}

func (this *BallerinaParser) parseAnnotChainingToken() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.ANNOT_CHAINING_TOKEN {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.ANNOT_CHAINING_TOKEN)
		return this.parseAnnotChainingToken()
	}
}

func (this *BallerinaParser) parseFieldAccessIdentifier(isInConditionalExpr bool) internal.STNode {
	nextToken := this.peek()
	if !this.isPredeclaredIdentifier(nextToken.kind) {
		identifier := this.SyntaxErrors.createMissingTokenWithDiagnostics(SyntaxKind.IDENTIFIER_TOKEN,
			DiagnosticErrorCode.ERROR_MISSING_IDENTIFIER)
		return this.parseQualifiedIdentifier(identifier, isInConditionalExpr)
	}
	return this.parseQualifiedIdentifier(ParserRuleContext.FIELD_ACCESS_IDENTIFIER, isInConditionalExpr)
}

func (this *BallerinaParser) parseQueryAction(queryConstructType internal.STNode, queryPipeline internal.STNode, selectClause internal.STNode, collectClause internal.STNode) internal.STNode {
	if queryConstructType != nil {
		queryPipeline = this.SyntaxErrors.cloneWithLeadingInvalidNodeMinutiae(queryPipeline, queryConstructType,
			DiagnosticErrorCode.ERROR_QUERY_CONSTRUCT_TYPE_IN_QUERY_ACTION)
	}
	if selectClause != nil {
		queryPipeline = this.SyntaxErrors.cloneWithTrailingInvalidNodeMinutiae(queryPipeline, selectClause,
			DiagnosticErrorCode.ERROR_SELECT_CLAUSE_IN_QUERY_ACTION)
	}
	if collectClause != nil {
		queryPipeline = this.SyntaxErrors.cloneWithTrailingInvalidNodeMinutiae(queryPipeline, collectClause,
			DiagnosticErrorCode.ERROR_COLLECT_CLAUSE_IN_QUERY_ACTION)
	}
	this.startContext(ParserRuleContext.DO_CLAUSE)
	doKeyword := this.parseDoKeyword()
	blockStmt := this.parseBlockNode()
	this.endContext()
	return this.STNodeFactory.createQueryActionNode(queryPipeline, doKeyword, blockStmt)
}

func (this *BallerinaParser) parseDoKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.DO_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.DO_KEYWORD)
		return this.parseDoKeyword()
	}
}

func (this *BallerinaParser) parseOptionalFieldAccessExpression(lhsExpr internal.STNode, isInConditionalExpr bool) internal.STNode {
	optionalFieldAccessToken := this.parseOptionalChainingToken()
	fieldName := this.parseFieldAccessIdentifier(isInConditionalExpr)
	return this.STNodeFactory.createOptionalFieldAccessExpressionNode(lhsExpr, optionalFieldAccessToken, fieldName)
}

func (this *BallerinaParser) parseOptionalChainingToken() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.OPTIONAL_CHAINING_TOKEN {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.OPTIONAL_CHAINING_TOKEN)
		return this.parseOptionalChainingToken()
	}
}

func (this *BallerinaParser) parseConditionalExpression(lhsExpr internal.STNode, isInConditionalExpr bool) internal.STNode {
	this.startContext(ParserRuleContext.CONDITIONAL_EXPRESSION)
	questionMark := this.parseQuestionMark()
	middleExpr := this.parseExpression(OperatorPrecedence.ANON_FUNC_OR_LET, true, false, true)
	if peek().kind != SyntaxKind.COLON_TOKEN {
		if middleExpr.kind == SyntaxKind.CONDITIONAL_EXPRESSION {
			innerConditionalExpr := internal.STConditionalExpressionNode(middleExpr)
			innerMiddleExpr := innerConditionalExpr.middleExpression
			rightMostQNameRef := this.ConditionalExprResolver.getQualifiedNameRefNode(innerMiddleExpr, false)
			if rightMostQNameRef != nil {
				middleExpr = this.generateConditionalExprForRightMost(innerConditionalExpr.lhsExpression,
					innerConditionalExpr.questionMarkToken, innerMiddleExpr, rightMostQNameRef)
				this.endContext()
				return this.STNodeFactory.createConditionalExpressionNode(lhsExpr, questionMark, middleExpr,
					innerConditionalExpr.colonToken, innerConditionalExpr.endExpression)
			}
			leftMostQNameRef := this.ConditionalExprResolver.getQualifiedNameRefNode(innerMiddleExpr, true)
			if leftMostQNameRef != nil {
				middleExpr = this.generateConditionalExprForLeftMost(innerConditionalExpr.lhsExpression,
					innerConditionalExpr.questionMarkToken, innerMiddleExpr, leftMostQNameRef)
				this.endContext()
				return this.STNodeFactory.createConditionalExpressionNode(lhsExpr, questionMark, middleExpr,
					innerConditionalExpr.colonToken, innerConditionalExpr.endExpression)
			}
		}
		rightMostQNameRef := this.ConditionalExprResolver.getQualifiedNameRefNode(middleExpr, false)
		if rightMostQNameRef != nil {
			this.endContext()
			return this.generateConditionalExprForRightMost(lhsExpr, questionMark, middleExpr, rightMostQNameRef)
		}
		leftMostQNameRef := this.ConditionalExprResolver.getQualifiedNameRefNode(middleExpr, true)
		if leftMostQNameRef != nil {
			this.endContext()
			return this.generateConditionalExprForLeftMost(lhsExpr, questionMark, middleExpr, leftMostQNameRef)
		}
	}
	return this.parseConditionalExprRhs(lhsExpr, questionMark, middleExpr, isInConditionalExpr)
}

func (this *BallerinaParser) generateConditionalExprForRightMost(lhsExpr internal.STNode, questionMark internal.STNode, middleExpr internal.STNode, rightMostQualifiedNameRef internal.STNode) internal.STNode {
	qualifiedNameRef := internal.STQualifiedNameReferenceNode(rightMostQualifiedNameRef)
	endExpr := this.STNodeFactory.createSimpleNameReferenceNode(qualifiedNameRef.identifier)
	simpleNameRef := this.ConditionalExprResolver.getSimpleNameRefNode(qualifiedNameRef.modulePrefix)
	middleExpr = this.middleExpr.replace(rightMostQualifiedNameRef, simpleNameRef)
	return this.STNodeFactory.createConditionalExpressionNode(lhsExpr, questionMark, middleExpr, qualifiedNameRef.colon,
		endExpr)
}

func (this *BallerinaParser) generateConditionalExprForLeftMost(lhsExpr internal.STNode, questionMark internal.STNode, middleExpr internal.STNode, leftMostQualifiedNameRef internal.STNode) internal.STNode {
	qualifiedNameRef := internal.STQualifiedNameReferenceNode(leftMostQualifiedNameRef)
	simpleNameRef := this.STNodeFactory.createSimpleNameReferenceNode(qualifiedNameRef.identifier)
	endExpr := this.middleExpr.replace(leftMostQualifiedNameRef, simpleNameRef)
	middleExpr = this.ConditionalExprResolver.getSimpleNameRefNode(qualifiedNameRef.modulePrefix)
	return this.STNodeFactory.createConditionalExpressionNode(lhsExpr, questionMark, middleExpr, qualifiedNameRef.colon,
		endExpr)
}

func (this *BallerinaParser) parseConditionalExprRhs(lhsExpr internal.STNode, questionMark internal.STNode, middleExpr internal.STNode, isInConditionalExpr bool) internal.STNode {
	colon := this.parseColon()
	this.endContext()
	endExpr := this.parseExpression(OperatorPrecedence.ANON_FUNC_OR_LET, true, false,
		isInConditionalExpr)
	return this.STNodeFactory.createConditionalExpressionNode(lhsExpr, questionMark, middleExpr, colon, endExpr)
}

func (this *BallerinaParser) parseEnumDeclaration(metadata internal.STNode, qualifier internal.STNode) internal.STNode {
	this.startContext(ParserRuleContext.MODULE_ENUM_DECLARATION)
	enumKeywordToken := this.parseEnumKeyword()
	identifier := this.parseIdentifier(ParserRuleContext.MODULE_ENUM_NAME)
	openBraceToken := this.parseOpenBrace()
	enumMemberList := this.parseEnumMemberList()
	closeBraceToken := this.parseCloseBrace()
	semicolon := this.parseOptionalSemicolon()
	this.endContext()
	openBraceToken = this.cloneWithDiagnosticIfListEmpty(enumMemberList, openBraceToken,
		DiagnosticErrorCode.ERROR_MISSING_ENUM_MEMBER)
	return this.STNodeFactory.createEnumDeclarationNode(metadata, qualifier, enumKeywordToken, identifier,
		openBraceToken, enumMemberList, closeBraceToken, semicolon)
}

func (this *BallerinaParser) parseEnumKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.ENUM_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.ENUM_KEYWORD)
		return this.parseEnumKeyword()
	}
}

func (this *BallerinaParser) parseEnumMemberList() internal.STNode {
	this.startContext(ParserRuleContext.ENUM_MEMBER_LIST)
	if peek().kind == SyntaxKind.CLOSE_BRACE_TOKEN {
		return this.STNodeFactory.createEmptyNodeList()
	}
	enumMemberList := make([]interface{}, 0)
	enumMember := this.parseEnumMember()
	var enumMemberRhs internal.STNode
	for peek().kind != SyntaxKind.CLOSE_BRACE_TOKEN {
		enumMemberRhs = this.parseEnumMemberEnd()
		if enumMemberRhs == nil {
			break
		}
		this.enumMemberList.add(enumMember)
		this.enumMemberList.add(enumMemberRhs)
		enumMember = this.parseEnumMember()
	}
	this.enumMemberList.add(enumMember)
	this.endContext()
	return this.STNodeFactory.createNodeList(enumMemberList)
}

func (this *BallerinaParser) parseEnumMember() internal.STNode {
	var metadata internal.STNode
	switch peek().kind {
	case DOCUMENTATION_STRING, AT_TOKEN:
		metadata = this.parseMetaData()
	default:
		metadata = this.STNodeFactory.createEmptyNode()
	}
	identifierNode := this.parseIdentifier(ParserRuleContext.ENUM_MEMBER_NAME)
	return this.parseEnumMemberRhs(metadata, identifierNode)
}

func (this *BallerinaParser) parseEnumMemberRhs(metadata internal.STNode, identifierNode internal.STNode) internal.STNode {
	var equalToken internal.STNode
	switch peek().kind {
	case EQUAL_TOKEN:
		equalToken = this.parseAssignOp()
		constExprNode = this.parseExpression()
		break
	case COMMA_TOKEN:
	case CLOSE_BRACE_TOKEN:
		equalToken = this.STNodeFactory.createEmptyNode()
		constExprNode = this.STNodeFactory.createEmptyNode()
		break
	default:
		this.recover(peek(), ParserRuleContext.ENUM_MEMBER_RHS)
		return this.parseEnumMemberRhs(metadata, identifierNode)
	}
	return this.STNodeFactory.createEnumMemberNode(metadata, identifierNode, equalToken, constExprNode)
}

func (this *BallerinaParser) parseEnumMemberEnd() internal.STNode {
	switch peek().kind {
	case COMMA_TOKEN:
		this.parseComma()
	case CLOSE_BRACE_TOKEN:
		nil
	default:
		this.recover(peek(), ParserRuleContext.ENUM_MEMBER_END)
		this.parseEnumMemberEnd()
	}
}

func (this *BallerinaParser) parseTransactionStmtOrVarDecl(annots internal.STNode, qualifiers []STNode, transactionKeyword internal.STToken) internal.STNode {
	switch peek().kind {
	case OPEN_BRACE_TOKEN:
		this.reportInvalidStatementAnnots(annots, qualifiers)
		this.reportInvalidQualifierList(qualifiers)
		return this.parseTransactionStatement(transactionKeyword)
	case COLON_TOKEN:
		if getNextNextToken().kind == SyntaxKind.IDENTIFIER_TOKEN {
			typeDesc := this.parseQualifiedIdentifierWithPredeclPrefix(transactionKeyword, false)
			return this.parseVarDeclTypeDescRhs(typeDesc, annots, qualifiers, true, false)
		}
	default:
		solution := this.recover(peek(), ParserRuleContext.TRANSACTION_STMT_RHS_OR_TYPE_REF)
		if (solution.action == Action.KEEP) || ((solution.action == Action.INSERT) && (solution.tokenKind == SyntaxKind.COLON_TOKEN)) {
			typeDesc := this.parseQualifiedIdentifierWithPredeclPrefix(transactionKeyword, false)
			return this.parseVarDeclTypeDescRhs(typeDesc, annots, qualifiers, true, false)
		}
		return this.parseTransactionStmtOrVarDecl(annots, qualifiers, transactionKeyword)
	}
}

func (this *BallerinaParser) parseTransactionStatement(transactionKeyword internal.STNode) internal.STNode {
	this.startContext(ParserRuleContext.TRANSACTION_STMT)
	blockStmt := this.parseBlockNode()
	this.endContext()
	onFailClause := this.parseOptionalOnFailClause()
	return this.STNodeFactory.createTransactionStatementNode(transactionKeyword, blockStmt, onFailClause)
}

func (this *BallerinaParser) parseCommitAction() internal.STNode {
	commitKeyword := this.parseCommitKeyword()
	return this.STNodeFactory.createCommitActionNode(commitKeyword)
}

func (this *BallerinaParser) parseCommitKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.COMMIT_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.COMMIT_KEYWORD)
		return this.parseCommitKeyword()
	}
}

func (this *BallerinaParser) parseRetryStatement() internal.STNode {
	this.startContext(ParserRuleContext.RETRY_STMT)
	retryKeyword := this.parseRetryKeyword()
	retryStmt := this.parseRetryKeywordRhs(retryKeyword)
	return retryStmt
}

func (this *BallerinaParser) parseRetryKeywordRhs(retryKeyword internal.STNode) internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case LT_TOKEN:
		this.parseRetryTypeParamRhs(retryKeyword, parseTypeParameter())
	case OPEN_PAREN_TOKEN,
		OPEN_BRACE_TOKEN,
		TRANSACTION_KEYWORD:
		this.parseRetryTypeParamRhs(retryKeyword, STNodeFactory.createEmptyNode())
	default:
		this.recover(peek(), ParserRuleContext.RETRY_KEYWORD_RHS)
		this.parseRetryKeywordRhs(retryKeyword)
	}
}

func (this *BallerinaParser) parseRetryTypeParamRhs(retryKeyword internal.STNode, typeParam internal.STNode) internal.STNode {
	var args internal.STNode
	switch peek().kind {
	case OPEN_PAREN_TOKEN:
		args = this.parseParenthesizedArgList()
		break
	case OPEN_BRACE_TOKEN:
	case TRANSACTION_KEYWORD:
		args = this.STNodeFactory.createEmptyNode()
		break
	default:
		this.recover(peek(), ParserRuleContext.RETRY_TYPE_PARAM_RHS)
		return this.parseRetryTypeParamRhs(retryKeyword, typeParam)
	}
	blockStmt := this.parseRetryBody()
	this.endContext()
	onFailClause := this.parseOptionalOnFailClause()
	return this.STNodeFactory.createRetryStatementNode(retryKeyword, typeParam, args, blockStmt, onFailClause)
}

func (this *BallerinaParser) parseRetryBody() internal.STNode {
	switch peek().kind {
	case OPEN_BRACE_TOKEN:
		this.parseBlockNode()
	case TRANSACTION_KEYWORD:
		this.parseTransactionStatement(consume())
	default:
		this.recover(peek(), ParserRuleContext.RETRY_BODY)
		this.parseRetryBody()
	}
}

func (this *BallerinaParser) parseOptionalOnFailClause() internal.STNode {
	nextToken := this.peek()
	if nextToken.kind == SyntaxKind.ON_KEYWORD {
		return this.parseOnFailClause()
	}
	if this.isEndOfRegularCompoundStmt(nextToken.kind) {
		return this.STNodeFactory.createEmptyNode()
	}
	this.recover(nextToken, ParserRuleContext.REGULAR_COMPOUND_STMT_RHS)
	return this.parseOptionalOnFailClause()
}

func (this *BallerinaParser) isEndOfRegularCompoundStmt(nodeKind SyntaxKind) bool {
	switch nodeKind {
	case CLOSE_BRACE_TOKEN, SEMICOLON_TOKEN, AT_TOKEN, EOF_TOKEN:
		true
	default:
		this.isStatementStartingToken(nodeKind)
	}
}

func (this *BallerinaParser) isStatementStartingToken(nodeKind SyntaxKind) bool {
	switch nodeKind {
	case FINAL_KEYWORD:
	case IF_KEYWORD:
	case WHILE_KEYWORD:
	case DO_KEYWORD:
	case PANIC_KEYWORD:
	case CONTINUE_KEYWORD:
	case BREAK_KEYWORD:
	case RETURN_KEYWORD:
	case LOCK_KEYWORD:
	case OPEN_BRACE_TOKEN:
	case FORK_KEYWORD:
	case FOREACH_KEYWORD:
	case XMLNS_KEYWORD:
	case TRANSACTION_KEYWORD:
	case RETRY_KEYWORD:
	case ROLLBACK_KEYWORD:
	case MATCH_KEYWORD:
	case FAIL_KEYWORD:
	case CHECK_KEYWORD:
	case CHECKPANIC_KEYWORD:
	case TRAP_KEYWORD:
	case START_KEYWORD:
	case FLUSH_KEYWORD:
	case LEFT_ARROW_TOKEN:
	case WAIT_KEYWORD:
	case COMMIT_KEYWORD:
	case WORKER_KEYWORD:
	case TYPE_KEYWORD:
	case CONST_KEYWORD:
		return true
	default:
		if this.isTypeStartingToken(nodeKind) {
			return true
		}
		if this.isValidExpressionStart(nodeKind, 1) {
			return true
		}
		return false
	}
}

func (this *BallerinaParser) parseOnFailClause() internal.STNode {
	this.startContext(ParserRuleContext.ON_FAIL_CLAUSE)
	onKeyword := this.parseOnKeyword()
	failKeyword := this.parseFailKeyword()
	typedBindingPattern := this.parseOnfailOptionalBP()
	blockStatement := this.parseBlockNode()
	this.endContext()
	return this.STNodeFactory.createOnFailClauseNode(onKeyword, failKeyword, typedBindingPattern,
		blockStatement)
}

func (this *BallerinaParser) parseOnfailOptionalBP() internal.STNode {
	nextToken := this.peek()
	if nextToken.kind == OPEN_BRACE_TOKEN {
		return this.STAbstractNodeFactory.createEmptyNode()
	} else if this.isTypeStartingToken(nextToken.kind) {
		return this.parseTypedBindingPattern()
	}
}

func (this *BallerinaParser) parseTypedBindingPattern() internal.STNode {
	typeDescriptor := this.parseTypeDescriptor(ParserRuleContext.TYPE_DESC_IN_TYPE_BINDING_PATTERN,
		true, false, TypePrecedence.DEFAULT)
	bindingPattern := this.parseBindingPattern()
	return this.STNodeFactory.createTypedBindingPatternNode(typeDescriptor, bindingPattern)
}

func (this *BallerinaParser) parseRetryKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.RETRY_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.RETRY_KEYWORD)
		return this.parseRetryKeyword()
	}
}

func (this *BallerinaParser) parseRollbackStatement() internal.STNode {
	this.startContext(ParserRuleContext.ROLLBACK_STMT)
	rollbackKeyword := this.parseRollbackKeyword()
	var expression internal.STNode
	if peek().kind == SyntaxKind.SEMICOLON_TOKEN {
		expression = this.STNodeFactory.createEmptyNode()
	} else {
		expression = this.parseExpression()
	}
	semicolon := this.parseSemicolon()
	this.endContext()
	return this.STNodeFactory.createRollbackStatementNode(rollbackKeyword, expression, semicolon)
}

func (this *BallerinaParser) parseRollbackKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.ROLLBACK_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.ROLLBACK_KEYWORD)
		return this.parseRollbackKeyword()
	}
}

func (this *BallerinaParser) parseTransactionalExpression() internal.STNode {
	transactionalKeyword := this.parseTransactionalKeyword()
	return this.STNodeFactory.createTransactionalExpressionNode(transactionalKeyword)
}

func (this *BallerinaParser) parseTransactionalKeyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.TRANSACTIONAL_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.TRANSACTIONAL_KEYWORD)
		return this.parseTransactionalKeyword()
	}
}

func (this *BallerinaParser) parseByteArrayLiteral() internal.STNode {
	var ty internal.STNode
	if peek().kind == SyntaxKind.BASE16_KEYWORD {
		ty = this.parseBase16Keyword()
	} else {
		ty = this.parseBase64Keyword()
	}
	startingBackTick := this.parseBacktickToken(ParserRuleContext.TEMPLATE_START)
	if this.startingBackTick.isMissing() {
		startingBackTick = this.SyntaxErrors.createMissingToken(SyntaxKind.BACKTICK_TOKEN)
		endingBackTick := this.SyntaxErrors.createMissingToken(SyntaxKind.BACKTICK_TOKEN)
		content := this.STNodeFactory.createEmptyNode()
		byteArrayLiteral := this.STNodeFactory.createByteArrayLiteralNode(ty, startingBackTick, content, endingBackTick)
		byteArrayLiteral = this.SyntaxErrors.addDiagnostic(byteArrayLiteral, DiagnosticErrorCode.ERROR_MISSING_BYTE_ARRAY_CONTENT)
		return byteArrayLiteral
	}
	content := this.parseByteArrayContent()
	return this.parseByteArrayLiteral(ty, startingBackTick, content)
}

func (this *BallerinaParser) parseByteArrayLiteral(typeKeyword internal.STNode, startingBackTick internal.STNode, byteArrayContent internal.STNode) internal.STNode {
	content := this.STNodeFactory.createEmptyNode()
	newStartingBackTick := startingBackTick
	items := internal.STNodeList(byteArrayContent)
	if len(items) == 1 {
		item := this.items.get(0)
		if (typeKeyword.kind == SyntaxKind.BASE16_KEYWORD) && (!this.isValidBase16LiteralContent(item.toString())) {
			newStartingBackTick = this.SyntaxErrors.cloneWithTrailingInvalidNodeMinutiae(startingBackTick, item,
				DiagnosticErrorCode.ERROR_INVALID_BASE16_CONTENT_IN_BYTE_ARRAY_LITERAL)
		} else if (typeKeyword.kind == SyntaxKind.BASE64_KEYWORD) && (!this.isValidBase64LiteralContent(item.toString())) {
			newStartingBackTick = this.SyntaxErrors.cloneWithTrailingInvalidNodeMinutiae(startingBackTick, item,
				DiagnosticErrorCode.ERROR_INVALID_BASE64_CONTENT_IN_BYTE_ARRAY_LITERAL)
		}
	} else if len(items) > 1 {
		clonedStartingBackTick := startingBackTick
		index := 0
		for ; index < len(items); index++ {
			item := this.items.get(index)
			clonedStartingBackTick = this.SyntaxErrors.cloneWithTrailingInvalidNodeMinutiae(clonedStartingBackTick, item)
		}
		newStartingBackTick = this.SyntaxErrors.addDiagnostic(clonedStartingBackTick,
			DiagnosticErrorCode.ERROR_INVALID_CONTENT_IN_BYTE_ARRAY_LITERAL)
	}
	endingBackTick := this.parseBacktickToken(ParserRuleContext.TEMPLATE_END)
	return this.STNodeFactory.createByteArrayLiteralNode(typeKeyword, newStartingBackTick, content, endingBackTick)
}

func (this *BallerinaParser) parseBase16Keyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.BASE16_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.BASE16_KEYWORD)
		return this.parseBase16Keyword()
	}
}

func (this *BallerinaParser) parseBase64Keyword() internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.BASE64_KEYWORD {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.BASE64_KEYWORD)
		return this.parseBase64Keyword()
	}
}

func (this *BallerinaParser) parseByteArrayContent() internal.STNode {
	nextToken := this.peek()
	items := make([]interface{}, 0)
	for !this.isEndOfBacktickContent(nextToken.kind) {
		content := this.parseTemplateItem()
		this.items.add(content)
		nextToken = this.peek()
	}
	return this.STNodeFactory.createNodeList(items)
}

func (this *BallerinaParser) parseXMLFilterExpression(lhsExpr internal.STNode) internal.STNode {
	xmlNamePatternChain := this.parseXMLFilterExpressionRhs()
	return this.STNodeFactory.createXMLFilterExpressionNode(lhsExpr, xmlNamePatternChain)
}

func (this *BallerinaParser) parseXMLFilterExpressionRhs() internal.STNode {
	dotLTToken := this.parseDotLTToken()
	return this.parseXMLNamePatternChain(dotLTToken)
}

func (this *BallerinaParser) parseXMLNamePatternChain(startToken internal.STNode) internal.STNode {
	this.startContext(ParserRuleContext.XML_NAME_PATTERN)
	xmlNamePattern := this.parseXMLNamePattern()
	gtToken := this.parseGTToken()
	this.endContext()
	startToken = this.cloneWithDiagnosticIfListEmpty(xmlNamePattern, startToken,
		DiagnosticErrorCode.ERROR_MISSING_XML_ATOMIC_NAME_PATTERN)
	return this.STNodeFactory.createXMLNamePatternChainingNode(startToken, xmlNamePattern, gtToken)
}

func (this *BallerinaParser) parseXMLStepExtends() internal.STNode {
	nextToken := this.peek()
	if this.isEndOfXMLStepExtend(nextToken.kind) {
		return this.STNodeFactory.createEmptyNodeList()
	}
	xmlStepExtendList := make([]interface{}, 0)
	this.startContext(ParserRuleContext.XML_STEP_EXTENDS)
	var stepExtension internal.STNode
	for !this.isEndOfXMLStepExtend(nextToken.kind) {
		if nextToken.kind == SyntaxKind.DOT_TOKEN {
			stepExtension = this.parseXMLStepMethodCallExtend()
		} else if nextToken.kind == SyntaxKind.DOT_LT_TOKEN {
			stepExtension = this.parseXMLFilterExpressionRhs()
		}
		this.xmlStepExtendList.add(stepExtension)
		nextToken = this.peek()
	}
	this.endContext()
	return this.STNodeFactory.createNodeList(xmlStepExtendList)
}

func (this *BallerinaParser) parseXMLIndexedStepExtend() internal.STNode {
	this.startContext(ParserRuleContext.MEMBER_ACCESS_KEY_EXPR)
	openBracket := this.parseOpenBracket()
	keyExpr := this.parseKeyExpr(true)
	closeBracket := this.parseCloseBracket()
	this.endContext()
	return this.STNodeFactory.createXMLStepIndexedExtendNode(openBracket, keyExpr, closeBracket)
}

func (this *BallerinaParser) parseXMLStepMethodCallExtend() internal.STNode {
	dotToken := this.parseDotToken()
	methodName := this.parseMethodName()
	parenthesizedArgsList := this.parseParenthesizedArgList()
	return this.STNodeFactory.createXMLStepMethodCallExtendNode(dotToken, methodName, parenthesizedArgsList)
}

func (this *BallerinaParser) parseMethodName() internal.STNode {
	if this.isSpecialMethodName(peek()) {
		return this.getKeywordAsSimpleNameRef()
	}
	return this.STNodeFactory.createSimpleNameReferenceNode(parseIdentifier(ParserRuleContext.IDENTIFIER))
}

func (this *BallerinaParser) parseDotLTToken() internal.STNode {
	nextToken := this.peek()
	if nextToken.kind == SyntaxKind.DOT_LT_TOKEN {
		return this.consume()
	} else {
		this.recover(nextToken, ParserRuleContext.DOT_LT_TOKEN)
		return this.parseDotLTToken()
	}
}

func (this *BallerinaParser) parseXMLNamePattern() internal.STNode {
	xmlAtomicNamePatternList := make([]interface{}, 0)
	nextToken := this.peek()
	if this.isEndOfXMLNamePattern(nextToken.kind) {
		return this.STNodeFactory.createNodeList(xmlAtomicNamePatternList)
	}
	xmlAtomicNamePattern := this.parseXMLAtomicNamePattern()
	this.xmlAtomicNamePatternList.add(xmlAtomicNamePattern)
	var separator internal.STNode
	for !this.isEndOfXMLNamePattern(peek().kind) {
		separator = this.parseXMLNamePatternSeparator()
		if separator == nil {
			break
		}
		this.xmlAtomicNamePatternList.add(separator)
		xmlAtomicNamePattern = this.parseXMLAtomicNamePattern()
		this.xmlAtomicNamePatternList.add(xmlAtomicNamePattern)
	}
	return this.STNodeFactory.createNodeList(xmlAtomicNamePatternList)
}

func (this *BallerinaParser) isEndOfXMLNamePattern(tokenKind SyntaxKind) bool {
	switch tokenKind {
	case GT_TOKEN, EOF_TOKEN:
		true
	default:
		false
	}
}

func (this *BallerinaParser) isEndOfXMLStepExtend(tokenKind SyntaxKind) bool {
	switch tokenKind {
	case OPEN_BRACKET_TOKEN, DOT_LT_TOKEN:
		false
	case DOT_TOKEN:
		(peek(3).kind != SyntaxKind.OPEN_PAREN_TOKEN)
	default:
		true
	}
}

func (this *BallerinaParser) parseXMLNamePatternSeparator() internal.STNode {
	token := this.peek()
	switch token.kind {
	case PIPE_TOKEN:
		this.consume()
	case GT_TOKEN, EOF_TOKEN:
		nil
	default:
		this.recover(token, ParserRuleContext.XML_NAME_PATTERN_RHS)
		this.parseXMLNamePatternSeparator()
	}
}

func (this *BallerinaParser) parseXMLAtomicNamePattern() internal.STNode {
	this.startContext(ParserRuleContext.XML_ATOMIC_NAME_PATTERN)
	atomicNamePattern := this.parseXMLAtomicNamePatternBody()
	this.endContext()
	return atomicNamePattern
}

func (this *BallerinaParser) parseXMLAtomicNamePatternBody() internal.STNode {
	token := this.peek()
	var identifier internal.STNode
	switch token.kind {
	case ASTERISK_TOKEN:
		return this.consume()
	case IDENTIFIER_TOKEN:
		identifier = this.consume()
		break
	default:
		this.recover(token, ParserRuleContext.XML_ATOMIC_NAME_PATTERN_START)
		return this.parseXMLAtomicNamePatternBody()
	}
	return this.parseXMLAtomicNameIdentifier(identifier)
}

func (this *BallerinaParser) parseXMLAtomicNameIdentifier(identifier internal.STNode) internal.STNode {
	token := this.peek()
	if token.kind == SyntaxKind.COLON_TOKEN {
		colon := this.consume()
		nextToken := this.peek()
		if (nextToken.kind == SyntaxKind.IDENTIFIER_TOKEN) || (nextToken.kind == SyntaxKind.ASTERISK_TOKEN) {
			endToken := this.consume()
			return this.STNodeFactory.createXMLAtomicNamePatternNode(identifier, colon, endToken)
		}
	}
	return this.STNodeFactory.createSimpleNameReferenceNode(identifier)
}

func (this *BallerinaParser) parseXMLStepExpression(lhsExpr internal.STNode) internal.STNode {
	xmlStepStart := this.parseXMLStepStart()
	xmlStepExtends := this.parseXMLStepExtends()
	return this.STNodeFactory.createXMLStepExpressionNode(lhsExpr, xmlStepStart, xmlStepExtends)
}

func (this *BallerinaParser) parseXMLStepStart() internal.STNode {
	token := this.peek()
	var startToken internal.STNode
	switch token.kind {
	case SLASH_ASTERISK_TOKEN:
		return this.consume()
	case DOUBLE_SLASH_DOUBLE_ASTERISK_LT_TOKEN:
		startToken = this.parseDoubleSlashDoubleAsteriskLTToken()
		break
	case SLASH_LT_TOKEN:
	default:
		startToken = this.parseSlashLTToken()
		break
	}
	return this.parseXMLNamePatternChain(startToken)
}

func (this *BallerinaParser) parseSlashLTToken() internal.STNode {
	nextToken := this.peek()
	if nextToken.kind == SyntaxKind.SLASH_LT_TOKEN {
		return this.consume()
	} else {
		this.recover(nextToken, ParserRuleContext.SLASH_LT_TOKEN)
		return this.parseSlashLTToken()
	}
}

func (this *BallerinaParser) parseDoubleSlashDoubleAsteriskLTToken() internal.STNode {
	nextToken := this.peek()
	if nextToken.kind == SyntaxKind.DOUBLE_SLASH_DOUBLE_ASTERISK_LT_TOKEN {
		return this.consume()
	} else {
		this.recover(nextToken, ParserRuleContext.DOUBLE_SLASH_DOUBLE_ASTERISK_LT_TOKEN)
		return this.parseDoubleSlashDoubleAsteriskLTToken()
	}
}

func (this *BallerinaParser) parseMatchStatement() internal.STNode {
	this.startContext(ParserRuleContext.MATCH_STMT)
	matchKeyword := this.parseMatchKeyword()
	actionOrExpr := this.parseActionOrExpression()
	this.startContext(ParserRuleContext.MATCH_BODY)
	openBrace := this.parseOpenBrace()
	matchClausesList := make([]interface{}, 0)
	for !this.isEndOfMatchClauses(peek().kind) {
		clause := this.parseMatchClause()
		this.matchClausesList.add(clause)
	}
	matchClauses := this.STNodeFactory.createNodeList(matchClausesList)
	if this.isNodeListEmpty(matchClauses) {
		openBrace = this.SyntaxErrors.addDiagnostic(openBrace,
			DiagnosticErrorCode.ERROR_MATCH_STATEMENT_SHOULD_HAVE_ONE_OR_MORE_MATCH_CLAUSES)
	}
	closeBrace := this.parseCloseBrace()
	this.endContext()
	this.endContext()
	onFailClause := this.parseOptionalOnFailClause()
	return this.STNodeFactory.createMatchStatementNode(matchKeyword, actionOrExpr, openBrace, matchClauses, closeBrace,
		onFailClause)
}

func (this *BallerinaParser) parseMatchKeyword() internal.STNode {
	nextToken := this.peek()
	if nextToken.kind == SyntaxKind.MATCH_KEYWORD {
		return this.consume()
	} else {
		this.recover(nextToken, ParserRuleContext.MATCH_KEYWORD)
		return this.parseMatchKeyword()
	}
}

func (this *BallerinaParser) isEndOfMatchClauses(nextTokenKind SyntaxKind) bool {
	switch nextTokenKind {
	case EOF_TOKEN, CLOSE_BRACE_TOKEN, TYPE_KEYWORD:
		true
	default:
		this.isEndOfStatements()
	}
}

func (this *BallerinaParser) parseMatchClause() internal.STNode {
	matchPatterns := this.parseMatchPatternList()
	matchGuard := this.parseMatchGuard()
	rightDoubleArrow := this.parseDoubleRightArrow()
	blockStmt := this.parseBlockNode()
	if this.isNodeListEmpty(matchPatterns) {
		identifier := this.SyntaxErrors.createMissingToken(SyntaxKind.IDENTIFIER_TOKEN)
		constantPattern := this.STNodeFactory.createSimpleNameReferenceNode(identifier)
		matchPatterns = this.STNodeFactory.createNodeList(constantPattern)
		errorCode := DiagnosticErrorCode.ERROR_MISSING_MATCH_PATTERN
		if matchGuard != nil {
			matchGuard = this.SyntaxErrors.addDiagnostic(matchGuard, errorCode)
		} else {
			rightDoubleArrow = this.SyntaxErrors.addDiagnostic(rightDoubleArrow, errorCode)
		}
	}
	return this.STNodeFactory.createMatchClauseNode(matchPatterns, matchGuard, rightDoubleArrow, blockStmt)
}

func (this *BallerinaParser) parseMatchGuard() internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case IF_KEYWORD:
		ifKeyword := this.parseIfKeyword()
		expr := this.parseExpression(DEFAULT_OP_PRECEDENCE, true, false, true, false)
		return this.STNodeFactory.createMatchGuardNode(ifKeyword, expr)
	case RIGHT_DOUBLE_ARROW_TOKEN:
		return this.STNodeFactory.createEmptyNode()
	default:
		this.recover(nextToken, ParserRuleContext.OPTIONAL_MATCH_GUARD)
		return this.parseMatchGuard()
	}
}

func (this *BallerinaParser) parseMatchPatternList() internal.STNode {
	this.startContext(ParserRuleContext.MATCH_PATTERN)
	matchClauses := make([]interface{}, 0)
	for !this.isEndOfMatchPattern(peek().kind) {
		clause := this.parseMatchPattern()
		if clause == nil {
			break
		}
		this.matchClauses.add(clause)
		seperator := this.parseMatchPatternListMemberRhs()
		if seperator == nil {
			break
		}
		this.matchClauses.add(seperator)
	}
	this.endContext()
	return this.STNodeFactory.createNodeList(matchClauses)
}

func (this *BallerinaParser) isEndOfMatchPattern(nextTokenKind SyntaxKind) bool {
	switch nextTokenKind {
	case PIPE_TOKEN, IF_KEYWORD, RIGHT_DOUBLE_ARROW_TOKEN:
		true
	default:
		false
	}
}

func (this *BallerinaParser) parseMatchPattern() internal.STNode {
	nextToken := this.peek()
	if this.isPredeclaredIdentifier(nextToken.kind) {
		typeRefOrConstExpr := this.parseQualifiedIdentifier(ParserRuleContext.MATCH_PATTERN)
		return this.parseErrorMatchPatternOrConsPattern(typeRefOrConstExpr)
	}
	switch nextToken.kind {
	case OPEN_PAREN_TOKEN,
		NULL_KEYWORD,
		TRUE_KEYWORD,
		FALSE_KEYWORD,
		PLUS_TOKEN,
		MINUS_TOKEN,
		DECIMAL_INTEGER_LITERAL_TOKEN,
		HEX_INTEGER_LITERAL_TOKEN,
		DECIMAL_FLOATING_POINT_LITERAL_TOKEN,
		HEX_FLOATING_POINT_LITERAL_TOKEN,
		STRING_LITERAL_TOKEN:
		this.parseSimpleConstExpr()
	case VAR_KEYWORD:
		this.parseVarTypedBindingPattern()
	case OPEN_BRACKET_TOKEN:
		this.parseListMatchPattern()
	case OPEN_BRACE_TOKEN:
		this.parseMappingMatchPattern()
	case ERROR_KEYWORD:
		this.parseErrorMatchPattern()
	default:
		this.recover(nextToken, ParserRuleContext.MATCH_PATTERN_START)
		this.parseMatchPattern()
	}
}

func (this *BallerinaParser) parseMatchPatternListMemberRhs() internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case PIPE_TOKEN:
		this.parsePipeToken()
	case IF_KEYWORD,
		RIGHT_DOUBLE_ARROW_TOKEN:
		nil
	default:
		this.recover(nextToken, ParserRuleContext.MATCH_PATTERN_LIST_MEMBER_RHS)
		this.parseMatchPatternListMemberRhs()
	}
}

func (this *BallerinaParser) parseVarTypedBindingPattern() internal.STNode {
	varKeyword := this.parseVarKeyword()
	varTypeDesc := this.createBuiltinSimpleNameReference(varKeyword)
	bindingPattern := this.parseBindingPattern()
	return this.STNodeFactory.createTypedBindingPatternNode(varTypeDesc, bindingPattern)
}

func (this *BallerinaParser) parseVarKeyword() internal.STNode {
	nextToken := this.peek()
	if nextToken.kind == SyntaxKind.VAR_KEYWORD {
		return this.consume()
	} else {
		this.recover(nextToken, ParserRuleContext.VAR_KEYWORD)
		return this.parseVarKeyword()
	}
}

func (this *BallerinaParser) parseListMatchPattern() internal.STNode {
	this.startContext(ParserRuleContext.LIST_MATCH_PATTERN)
	openBracketToken := this.parseOpenBracket()
	matchPatternList := make([]interface{}, 0)
	listMatchPatternMemberRhs := nil
	isEndOfFields := false
	for !this.isEndOfListMatchPattern() {
		listMatchPatternMember := this.parseListMatchPatternMember()
		this.matchPatternList.add(listMatchPatternMember)
		listMatchPatternMemberRhs = this.parseListMatchPatternMemberRhs()
		if listMatchPatternMember.kind == SyntaxKind.REST_MATCH_PATTERN {
			isEndOfFields = true
			break
		}
		if listMatchPatternMemberRhs != nil {
			this.matchPatternList.add(listMatchPatternMemberRhs)
		} else {
			break
		}
	}
	for isEndOfFields && (listMatchPatternMemberRhs != nil) {
		this.updateLastNodeInListWithInvalidNode(matchPatternList, listMatchPatternMemberRhs, null)
		if peek().kind == SyntaxKind.CLOSE_BRACKET_TOKEN {
			break
		}
		invalidField := this.parseListMatchPatternMember()
		this.updateLastNodeInListWithInvalidNode(matchPatternList, invalidField,
			DiagnosticErrorCode.ERROR_MATCH_PATTERN_AFTER_REST_MATCH_PATTERN)
		listMatchPatternMemberRhs = this.parseListMatchPatternMemberRhs()
	}
	matchPatternListNode := this.STNodeFactory.createNodeList(matchPatternList)
	closeBracketToken := this.parseCloseBracket()
	this.endContext()
	return this.STNodeFactory.createListMatchPatternNode(openBracketToken, matchPatternListNode, closeBracketToken)
}

func (this *BallerinaParser) IsEndOfListMatchPattern() bool {
	switch peek().kind {
	case CLOSE_BRACKET_TOKEN, EOF_TOKEN:
		true
	default:
		false
	}
}

func (this *BallerinaParser) parseListMatchPatternMember() internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case ELLIPSIS_TOKEN:
		this.parseRestMatchPattern()
	default:
		this.parseMatchPattern()
	}
}

func (this *BallerinaParser) parseRestMatchPattern() internal.STNode {
	this.startContext(ParserRuleContext.REST_MATCH_PATTERN)
	ellipsisToken := this.parseEllipsis()
	varKeywordToken := this.parseVarKeyword()
	variableName := this.parseVariableName()
	this.endContext()
	simpleNameReferenceNode := internal.STSimpleNameReferenceNode(this.STNodeFactory.createSimpleNameReferenceNode(variableName))
	return this.STNodeFactory.createRestMatchPatternNode(ellipsisToken, varKeywordToken, simpleNameReferenceNode)
}

func (this *BallerinaParser) parseListMatchPatternMemberRhs() internal.STNode {
	switch peek().kind {
	case COMMA_TOKEN:
		this.parseComma()
	case CLOSE_BRACKET_TOKEN, EOF_TOKEN:
		nil
	default:
		this.recover(peek(), ParserRuleContext.LIST_MATCH_PATTERN_MEMBER_RHS)
		this.parseListMatchPatternMemberRhs()
	}
}

func (this *BallerinaParser) parseMappingMatchPattern() internal.STNode {
	this.startContext(ParserRuleContext.MAPPING_MATCH_PATTERN)
	openBraceToken := this.parseOpenBrace()
	fieldMatchPatterns := this.parseFieldMatchPatternList()
	closeBraceToken := this.parseCloseBrace()
	this.endContext()
	return this.STNodeFactory.createMappingMatchPatternNode(openBraceToken, fieldMatchPatterns, closeBraceToken)
}

func (this *BallerinaParser) parseFieldMatchPatternList() internal.STNode {
	fieldMatchPatterns := make([]interface{}, 0)
	fieldMatchPatternMember := this.parseFieldMatchPatternMember()
	if fieldMatchPatternMember == nil {
		return this.STNodeFactory.createEmptyNodeList()
	}
	this.fieldMatchPatterns.add(fieldMatchPatternMember)
	if fieldMatchPatternMember.kind == SyntaxKind.REST_MATCH_PATTERN {
		this.invalidateExtraFieldMatchPatterns(fieldMatchPatterns)
		return this.STNodeFactory.createNodeList(fieldMatchPatterns)
	}
	return this.parseFieldMatchPatternList(fieldMatchPatterns)
}

func (this *BallerinaParser) parseFieldMatchPatternList(fieldMatchPatterns []STNode) internal.STNode {
	for !this.isEndOfMappingMatchPattern() {
		fieldMatchPatternRhs := this.parseFieldMatchPatternRhs()
		if fieldMatchPatternRhs == nil {
			break
		}
		this.fieldMatchPatterns.add(fieldMatchPatternRhs)
		fieldMatchPatternMember := this.parseFieldMatchPatternMember()
		if fieldMatchPatternMember == nil {
			fieldMatchPatternMember = this.createMissingFieldMatchPattern()
		}
		this.fieldMatchPatterns.add(fieldMatchPatternMember)
		if fieldMatchPatternMember.kind == SyntaxKind.REST_MATCH_PATTERN {
			this.invalidateExtraFieldMatchPatterns(fieldMatchPatterns)
			break
		}
	}
	return this.STNodeFactory.createNodeList(fieldMatchPatterns)
}

func (this *BallerinaParser) createMissingFieldMatchPattern() internal.STNode {
	fieldName := this.SyntaxErrors.createMissingToken(SyntaxKind.IDENTIFIER_TOKEN)
	colon := this.SyntaxErrors.createMissingToken(SyntaxKind.COLON_TOKEN)
	identifier := this.SyntaxErrors.createMissingToken(SyntaxKind.IDENTIFIER_TOKEN)
	matchPattern := this.STNodeFactory.createSimpleNameReferenceNode(identifier)
	fieldMatchPatternMember := this.STNodeFactory.createFieldMatchPatternNode(fieldName, colon, matchPattern)
	fieldMatchPatternMember = this.SyntaxErrors.addDiagnostic(fieldMatchPatternMember,
		DiagnosticErrorCode.ERROR_MISSING_FIELD_MATCH_PATTERN_MEMBER)
	return fieldMatchPatternMember
}

func (this *BallerinaParser) invalidateExtraFieldMatchPatterns(fieldMatchPatterns []STNode) {
	for !this.isEndOfMappingMatchPattern() {
		fieldMatchPatternRhs := this.parseFieldMatchPatternRhs()
		if fieldMatchPatternRhs == nil {
			break
		}
		fieldMatchPatternMember := this.parseFieldMatchPatternMember()
		if fieldMatchPatternMember == nil {
			rhsToken, ok := fieldMatchPatternRhs.(STToken)
			if !ok {
				panic("invalidateExtraFieldMatchPatterns: expected STToken")
			}
			this.updateLastNodeInListWithInvalidNode(fieldMatchPatterns, fieldMatchPatternRhs,
				DiagnosticErrorCode.ERROR_INVALID_TOKEN, rhsToken.text())
		} else {
			this.updateLastNodeInListWithInvalidNode(fieldMatchPatterns, fieldMatchPatternRhs, null)
			this.updateLastNodeInListWithInvalidNode(fieldMatchPatterns, fieldMatchPatternMember,
				DiagnosticErrorCode.ERROR_MATCH_PATTERN_AFTER_REST_MATCH_PATTERN)
		}
	}
}

func (this *BallerinaParser) parseFieldMatchPatternMember() internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case IDENTIFIER_TOKEN:
		this.parseFieldMatchPattern()
	case ELLIPSIS_TOKEN:
		this.parseRestMatchPattern()
	case CLOSE_BRACE_TOKEN, EOF_TOKEN:
		nil
	default:
		this.recover(nextToken, ParserRuleContext.FIELD_MATCH_PATTERNS_START)
		this.parseFieldMatchPatternMember()
	}
}

func (this *BallerinaParser) ParseFieldMatchPattern() internal.STNode {
	fieldNameNode := this.parseVariableName()
	colonToken := this.parseColon()
	matchPattern := this.parseMatchPattern()
	return this.STNodeFactory.createFieldMatchPatternNode(fieldNameNode, colonToken, matchPattern)
}

func (this *BallerinaParser) IsEndOfMappingMatchPattern() bool {
	switch peek().kind {
	case CLOSE_BRACE_TOKEN, EOF_TOKEN:
		true
	default:
		false
	}
}

func (this *BallerinaParser) parseFieldMatchPatternRhs() internal.STNode {
	switch peek().kind {
	case COMMA_TOKEN:
		this.parseComma()
	case CLOSE_BRACE_TOKEN, EOF_TOKEN:
		nil
	default:
		this.recover(peek(), ParserRuleContext.FIELD_MATCH_PATTERN_MEMBER_RHS)
		this.parseFieldMatchPatternRhs()
	}
}

func (this *BallerinaParser) parseErrorMatchPatternOrConsPattern(typeRefOrConstExpr internal.STNode) internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case OPEN_PAREN_TOKEN:
		errorKeyword := this.SyntaxErrors.createMissingTokenWithDiagnostics(SyntaxKind.ERROR_KEYWORD,
			ParserRuleContext.ERROR_KEYWORD)
		this.startContext(ParserRuleContext.ERROR_MATCH_PATTERN)
		return this.parseErrorMatchPattern(errorKeyword, typeRefOrConstExpr)
	default:
		if this.isMatchPatternEnd(peek().kind) {
			return typeRefOrConstExpr
		}
		this.recover(peek(), ParserRuleContext.ERROR_MATCH_PATTERN_OR_CONST_PATTERN)
		return this.parseErrorMatchPatternOrConsPattern(typeRefOrConstExpr)
	}
}

func (this *BallerinaParser) isMatchPatternEnd(tokenKind SyntaxKind) bool {
	switch tokenKind {
	case RIGHT_DOUBLE_ARROW_TOKEN,
		COMMA_TOKEN,
		CLOSE_BRACE_TOKEN,
		CLOSE_BRACKET_TOKEN,
		CLOSE_PAREN_TOKEN,
		PIPE_TOKEN,
		IF_KEYWORD,
		EOF_TOKEN:
		true
	default:
		false
	}
}

func (this *BallerinaParser) parseErrorMatchPattern() internal.STNode {
	this.startContext(ParserRuleContext.ERROR_MATCH_PATTERN)
	errorKeyword := this.consume()
	return this.parseErrorMatchPattern(errorKeyword)
}

func (this *BallerinaParser) parseErrorMatchPattern(errorKeyword internal.STNode) internal.STNode {
	nextToken := this.peek()
	var typeRef internal.STNode
	switch nextToken.kind {
	case OPEN_PAREN_TOKEN:
		typeRef = this.STNodeFactory.createEmptyNode()
		break
	default:
		if this.isPredeclaredIdentifier(nextToken.kind) {
			typeRef = this.parseTypeReference()
			break
		}
		this.recover(peek(), ParserRuleContext.ERROR_MATCH_PATTERN_ERROR_KEYWORD_RHS)
		return this.parseErrorMatchPattern(errorKeyword)
	}
	return this.parseErrorMatchPattern(errorKeyword, typeRef)
}

func (this *BallerinaParser) parseErrorMatchPattern(errorKeyword internal.STNode, typeRef internal.STNode) internal.STNode {
	openParenthesisToken := this.parseOpenParenthesis()
	argListMatchPatternNode := this.parseErrorArgListMatchPatterns()
	closeParenthesisToken := this.parseCloseParenthesis()
	this.endContext()
	return this.STNodeFactory.createErrorMatchPatternNode(errorKeyword, typeRef, openParenthesisToken,
		argListMatchPatternNode, closeParenthesisToken)
}

func (this *BallerinaParser) parseErrorArgListMatchPatterns() internal.STNode {
	argListMatchPatterns := make([]interface{}, 0)
	if this.isEndOfErrorFieldMatchPatterns() {
		return this.STNodeFactory.createNodeList(argListMatchPatterns)
	}
	this.startContext(ParserRuleContext.ERROR_ARG_LIST_MATCH_PATTERN_FIRST_ARG)
	firstArg := this.parseErrorArgListMatchPattern(ParserRuleContext.ERROR_ARG_LIST_MATCH_PATTERN_START)
	this.endContext()
	if this.isSimpleMatchPattern(firstArg.kind) {
		this.argListMatchPatterns.add(firstArg)
		argEnd := this.parseErrorArgListMatchPatternEnd(ParserRuleContext.ERROR_MESSAGE_MATCH_PATTERN_END)
		if argEnd != nil {
			secondArg := this.parseErrorArgListMatchPattern(ParserRuleContext.ERROR_MESSAGE_MATCH_PATTERN_RHS)
			if this.isValidSecondArgMatchPattern(secondArg.kind) {
				this.argListMatchPatterns.add(argEnd)
				this.argListMatchPatterns.add(secondArg)
			} else {
				this.updateLastNodeInListWithInvalidNode(argListMatchPatterns, argEnd, null)
				this.updateLastNodeInListWithInvalidNode(argListMatchPatterns, secondArg,
					DiagnosticErrorCode.ERROR_MATCH_PATTERN_NOT_ALLOWED)
			}
		}
	} else {
		if (firstArg.kind != SyntaxKind.NAMED_ARG_MATCH_PATTERN) && (firstArg.kind != SyntaxKind.REST_MATCH_PATTERN) {
			this.addInvalidNodeToNextToken(firstArg, DiagnosticErrorCode.ERROR_MATCH_PATTERN_NOT_ALLOWED)
		} else {
			this.argListMatchPatterns.add(firstArg)
		}
	}
	this.parseErrorFieldMatchPatterns(argListMatchPatterns)
	return this.STNodeFactory.createNodeList(argListMatchPatterns)
}

func (this *BallerinaParser) isSimpleMatchPattern(matchPatternKind SyntaxKind) bool {
	switch matchPatternKind {
	case IDENTIFIER_TOKEN,
		SIMPLE_NAME_REFERENCE,
		QUALIFIED_NAME_REFERENCE,
		NUMERIC_LITERAL,
		STRING_LITERAL,
		NULL_LITERAL,
		NIL_LITERAL,
		BOOLEAN_LITERAL,
		TYPED_BINDING_PATTERN,
		UNARY_EXPRESSION:
		true
	default:
		false
	}
}

func (this *BallerinaParser) isValidSecondArgMatchPattern(syntaxKind SyntaxKind) bool {
	switch syntaxKind {
	case ERROR_MATCH_PATTERN,
		NAMED_ARG_MATCH_PATTERN,
		REST_MATCH_PATTERN:
		true
	default:
		this.isSimpleMatchPattern(syntaxKind)
	}
}

func (this *BallerinaParser) parseErrorFieldMatchPatterns(argListMatchPatterns []STNode) {
	lastValidArgKind := SyntaxKind.NAMED_ARG_MATCH_PATTERN
	for !this.isEndOfErrorFieldMatchPatterns() {
		argEnd := this.parseErrorArgListMatchPatternEnd(ParserRuleContext.ERROR_FIELD_MATCH_PATTERN_RHS)
		if argEnd == nil {
			break
		}
		currentArg := this.parseErrorArgListMatchPattern(ParserRuleContext.ERROR_FIELD_MATCH_PATTERN)
		errorCode := this.validateErrorFieldMatchPatternOrder(lastValidArgKind, currentArg.kind)
		if errorCode == nil {
			this.argListMatchPatterns.add(argEnd)
			this.argListMatchPatterns.add(currentArg)
			lastValidArgKind = currentArg.kind
		} else if this.argListMatchPatterns.isEmpty() {
			this.addInvalidNodeToNextToken(argEnd, null)
			this.addInvalidNodeToNextToken(currentArg, errorCode)
		}
	}
}

func (this *BallerinaParser) isEndOfErrorFieldMatchPatterns() bool {
	return this.isEndOfErrorFieldBindingPatterns()
}

func (this *BallerinaParser) parseErrorArgListMatchPatternEnd(currentCtx ParserRuleContext) internal.STNode {
	switch peek().kind {
	case COMMA_TOKEN:
		this.consume()
	case CLOSE_PAREN_TOKEN:
		nil
	default:
		this.recover(peek(), currentCtx)
		this.parseErrorArgListMatchPatternEnd(currentCtx)
	}
}

func (this *BallerinaParser) parseErrorArgListMatchPattern(context ParserRuleContext) internal.STNode {
	nextToken := this.peek()
	if this.isPredeclaredIdentifier(nextToken.kind) {
		return this.parseNamedArgOrSimpleMatchPattern()
	}
	switch nextToken.kind {
	case ELLIPSIS_TOKEN:
		return this.parseRestMatchPattern()
	case OPEN_PAREN_TOKEN:
	case NULL_KEYWORD:
	case TRUE_KEYWORD:
	case FALSE_KEYWORD:
	case PLUS_TOKEN:
	case MINUS_TOKEN:
	case DECIMAL_INTEGER_LITERAL_TOKEN:
	case HEX_INTEGER_LITERAL_TOKEN:
	case DECIMAL_FLOATING_POINT_LITERAL_TOKEN:
	case HEX_FLOATING_POINT_LITERAL_TOKEN:
	case STRING_LITERAL_TOKEN:
	case OPEN_BRACKET_TOKEN:
	case OPEN_BRACE_TOKEN:
	case ERROR_KEYWORD:
		return this.parseMatchPattern()
	case VAR_KEYWORD:
		varType := this.createBuiltinSimpleNameReference(consume())
		variableName := this.createCaptureOrWildcardBP(parseVariableName())
		return this.STNodeFactory.createTypedBindingPatternNode(varType, variableName)
	case CLOSE_PAREN_TOKEN:
		return this.SyntaxErrors.createMissingTokenWithDiagnostics(SyntaxKind.IDENTIFIER_TOKEN,
			DiagnosticErrorCode.ERROR_MISSING_MATCH_PATTERN)
	default:
		this.recover(nextToken, context)
		return this.parseErrorArgListMatchPattern(context)
	}
}

func (this *BallerinaParser) parseNamedArgOrSimpleMatchPattern() internal.STNode {
	constRefExpr := this.parseQualifiedIdentifier(ParserRuleContext.MATCH_PATTERN)
	if (constRefExpr.kind == SyntaxKind.QUALIFIED_NAME_REFERENCE) || (peek().kind != SyntaxKind.EQUAL_TOKEN) {
		return constRefExpr
	}
	simpleNameNode, ok := constRefExpr.(*STSimpleNameReferenceNode)
	if !ok {
		panic("parseNamedArgOrSimpleMatchPattern: expected STSimpleNameReferenceNode")
	}
	return this.parseNamedArgMatchPattern(simpleNameNode.name)
}

func (this *BallerinaParser) parseNamedArgMatchPattern(identifier internal.STNode) internal.STNode {
	this.startContext(ParserRuleContext.NAMED_ARG_MATCH_PATTERN)
	equalToken := this.parseAssignOp()
	matchPattern := this.parseMatchPattern()
	this.endContext()
	return this.STNodeFactory.createNamedArgMatchPatternNode(identifier, equalToken, matchPattern)
}

func (this *BallerinaParser) validateErrorFieldMatchPatternOrder(prevArgKind SyntaxKind, currentArgKind SyntaxKind) DiagnosticErrorCode {
	switch currentArgKind {
	case NAMED_ARG_MATCH_PATTERN,
		REST_MATCH_PATTERN:
		if prevArgKind == SyntaxKind.REST_MATCH_PATTERN {
			DiagnosticErrorCode.ERROR_REST_ARG_FOLLOWED_BY_ANOTHER_ARG
		}
		nil
	default:
		DiagnosticErrorCode.ERROR_MATCH_PATTERN_NOT_ALLOWED
	}
}

func (this *BallerinaParser) parseMarkdownDocumentation() internal.STNode {
	markdownDocLineList := make([]interface{}, 0)
	nextToken := this.peek()
	for nextToken.kind == SyntaxKind.DOCUMENTATION_STRING {
		documentationString := this.consume()
		parsedDocLines := this.parseDocumentationString(documentationString)
		this.appendParsedDocumentationLines(markdownDocLineList, parsedDocLines)
		nextToken = this.peek()
	}
	markdownDocLines := this.STNodeFactory.createNodeList(markdownDocLineList)
	return this.STNodeFactory.createMarkdownDocumentationNode(markdownDocLines)
}

func (this *BallerinaParser) parseDocumentationString(documentationStringToken internal.STToken) internal.STNode {
	leadingTriviaList := this.getLeadingTriviaList(documentationStringToken.leadingMinutiae())
	diagnostics := make([]interface{}, 0)
	charReader := this.CharReader.from(documentationStringToken.text())
	documentationLexer := nil
	tokenReader := nil
	documentationParser := nil
	return this.documentationParser.parse()
}

func (this *BallerinaParser) getLeadingTriviaList(leadingMinutiaeNode internal.STNode) []STNode {
	leadingTriviaList := make([]interface{}, 0)
	bucketCount := this.leadingMinutiaeNode.bucketCount()
	i := 0
	for ; i < bucketCount; i++ {
		this.leadingTriviaList.add(leadingMinutiaeNode.childInBucket(i))
	}
	return leadingTriviaList
}

func (this *BallerinaParser) appendParsedDocumentationLines(markdownDocLineList []STNode, parsedDocLines internal.STNode) {
	bucketCount := this.parsedDocLines.bucketCount()
	i := 0
	for ; i < bucketCount; i++ {
		markdownDocLine := this.parsedDocLines.childInBucket(i)
		this.markdownDocLineList.add(markdownDocLine)
	}
}

func (this *BallerinaParser) parseStmtStartsWithTypeOrExpr(annots internal.STNode, qualifiers []STNode) internal.STNode {
	this.startContext(ParserRuleContext.AMBIGUOUS_STMT)
	typeOrExpr := this.parseTypedBindingPatternOrExpr(qualifiers, true)
	return this.parseStmtStartsWithTypedBPOrExprRhs(annots, typeOrExpr)
}

func (this *BallerinaParser) parseStmtStartsWithTypedBPOrExprRhs(annots internal.STNode, typedBindingPatternOrExpr internal.STNode) internal.STNode {
	if typedBindingPatternOrExpr.kind == SyntaxKind.TYPED_BINDING_PATTERN {
		varDeclQualifiers := make([]interface{}, 0)
		this.switchContext(ParserRuleContext.VAR_DECL_STMT)
		return this.parseVarDeclRhs(annots, varDeclQualifiers, typedBindingPatternOrExpr, false)
	}
	expr := this.getExpression(typedBindingPatternOrExpr)
	expr = this.getExpression(parseExpressionRhs(DEFAULT_OP_PRECEDENCE, expr, false, true))
	return this.parseStatementStartWithExprRhs(expr)
}

func (this *BallerinaParser) parseTypedBindingPatternOrExpr(allowAssignment bool) internal.STNode {
	typeDescQualifiers := make([]interface{}, 0)
	return this.parseTypedBindingPatternOrExpr(typeDescQualifiers, allowAssignment)
}

func (this *BallerinaParser) parseTypedBindingPatternOrExpr(qualifiers []STNode, allowAssignment bool) internal.STNode {
	this.parseTypeDescQualifiers(qualifiers)
	nextToken := this.peek()
	var typeOrExpr internal.STNode
	if this.isPredeclaredIdentifier(nextToken.kind) {
		this.reportInvalidQualifierList(qualifiers)
		typeOrExpr = this.parseQualifiedIdentifier(ParserRuleContext.TYPE_NAME_OR_VAR_NAME)
		return this.parseTypedBindingPatternOrExprRhs(typeOrExpr, allowAssignment)
	}
	switch nextToken.kind {
	case OPEN_PAREN_TOKEN:
		this.reportInvalidQualifierList(qualifiers)
		return this.parseTypedBPOrExprStartsWithOpenParenthesis()
	case FUNCTION_KEYWORD:
		return this.parseAnonFuncExprOrTypedBPWithFuncType(qualifiers)
	case OPEN_BRACKET_TOKEN:
		this.reportInvalidQualifierList(qualifiers)
		typeOrExpr = this.parseTupleTypeDescOrListConstructor(STNodeFactory.createEmptyNodeList())
		return this.parseTypedBindingPatternOrExprRhs(typeOrExpr, allowAssignment)
	case DECIMAL_INTEGER_LITERAL_TOKEN:
	case HEX_INTEGER_LITERAL_TOKEN:
	case STRING_LITERAL_TOKEN:
	case NULL_KEYWORD:
	case TRUE_KEYWORD:
	case FALSE_KEYWORD:
	case DECIMAL_FLOATING_POINT_LITERAL_TOKEN:
	case HEX_FLOATING_POINT_LITERAL_TOKEN:
		this.reportInvalidQualifierList(qualifiers)
		basicLiteral := this.parseBasicLiteral()
		return this.parseTypedBindingPatternOrExprRhs(basicLiteral, allowAssignment)
	default:
		if this.isValidExpressionStart(nextToken.kind, 1) {
			this.reportInvalidQualifierList(qualifiers)
			return this.parseActionOrExpressionInLhs(STNodeFactory.createEmptyNodeList())
		}
		return this.parseTypedBindingPattern(qualifiers, ParserRuleContext.VAR_DECL_STMT)
	}
}

func (this *BallerinaParser) parseTypedBindingPatternOrExprRhs(typeOrExpr internal.STNode, allowAssignment bool) internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case PIPE_TOKEN:
	case BITWISE_AND_TOKEN:
		nextNextToken := this.peek(2)
		if nextNextToken.kind == SyntaxKind.EQUAL_TOKEN {
			return typeOrExpr
		}
		pipeOrAndToken := this.parseBinaryOperator()
		rhsTypedBPOrExpr := this.parseTypedBindingPatternOrExpr(allowAssignment)
		if rhsTypedBPOrExpr.kind == SyntaxKind.TYPED_BINDING_PATTERN {
			typedBP := internal.STTypedBindingPatternNode(rhsTypedBPOrExpr)
			typeOrExpr = this.getTypeDescFromExpr(typeOrExpr)
			newTypeDesc := this.mergeTypes(typeOrExpr, pipeOrAndToken, typedBP.typeDescriptor)
			return this.STNodeFactory.createTypedBindingPatternNode(newTypeDesc, typedBP.bindingPattern)
		}
		if peek().kind == SyntaxKind.EQUAL_TOKEN {
			return this.createCaptureBPWithMissingVarName(typeOrExpr, pipeOrAndToken, rhsTypedBPOrExpr)
		}
		return this.STNodeFactory.createBinaryExpressionNode(SyntaxKind.BINARY_EXPRESSION, typeOrExpr,
			pipeOrAndToken, rhsTypedBPOrExpr)
	case SEMICOLON_TOKEN:
		if this.isExpression(typeOrExpr.kind) {
			return typeOrExpr
		}
		if this.isDefiniteTypeDesc(typeOrExpr.kind) || (!this.isAllBasicLiterals(typeOrExpr)) {
			typeDesc := this.getTypeDescFromExpr(typeOrExpr)
			return this.parseTypeBindingPatternStartsWithAmbiguousNode(typeDesc)
		}
		return typeOrExpr
	case IDENTIFIER_TOKEN:
	case QUESTION_MARK_TOKEN:
		if this.isAmbiguous(typeOrExpr) || this.isDefiniteTypeDesc(typeOrExpr.kind) {
			typeDesc := this.getTypeDescFromExpr(typeOrExpr)
			return this.parseTypeBindingPatternStartsWithAmbiguousNode(typeDesc)
		}
		return typeOrExpr
	case EQUAL_TOKEN:
		return typeOrExpr
	case OPEN_BRACKET_TOKEN:
		return this.parseTypedBindingPatternOrMemberAccess(typeOrExpr, false, allowAssignment,
			ParserRuleContext.AMBIGUOUS_STMT)
	case OPEN_BRACE_TOKEN:
	case ERROR_KEYWORD:
		typeDesc := this.getTypeDescFromExpr(typeOrExpr)
		return this.parseTypeBindingPatternStartsWithAmbiguousNode(typeDesc)
	default:
		if this.isCompoundAssignment(nextToken.kind) {
			return typeOrExpr
		}
		if this.isValidExprRhsStart(nextToken.kind, typeOrExpr.kind) {
			return typeOrExpr
		}
		token := this.peek()
		typeOrExprKind := typeOrExpr.kind
		if (typeOrExprKind == SyntaxKind.QUALIFIED_NAME_REFERENCE) || (typeOrExprKind == SyntaxKind.SIMPLE_NAME_REFERENCE) {
			this.recover(token, ParserRuleContext.BINDING_PATTERN_OR_VAR_REF_RHS)
		} else {
			this.recover(token, ParserRuleContext.BINDING_PATTERN_OR_EXPR_RHS)
		}
		return this.parseTypedBindingPatternOrExprRhs(typeOrExpr, allowAssignment)
	}
}

func (this *BallerinaParser) createCaptureBPWithMissingVarName(lhsType internal.STNode, separatorToken internal.STNode, rhsType internal.STNode) internal.STNode {
	lhsType = this.getTypeDescFromExpr(lhsType)
	rhsType = this.getTypeDescFromExpr(rhsType)
	newTypeDesc := this.mergeTypes(lhsType, separatorToken, rhsType)
	identifier := this.SyntaxErrors.createMissingTokenWithDiagnostics(SyntaxKind.IDENTIFIER_TOKEN,
		ParserRuleContext.VARIABLE_NAME)
	captureBP := this.STNodeFactory.createCaptureBindingPatternNode(identifier)
	return this.STNodeFactory.createTypedBindingPatternNode(newTypeDesc, captureBP)
}

func (this *BallerinaParser) parseTypeBindingPatternStartsWithAmbiguousNode(typeDesc internal.STNode) internal.STNode {
	typeDesc = this.parseComplexTypeDescriptor(typeDesc, ParserRuleContext.TYPE_DESC_IN_TYPE_BINDING_PATTERN, true)
	return this.parseTypedBindingPatternTypeRhs(typeDesc, ParserRuleContext.VAR_DECL_STMT)
}

func (this *BallerinaParser) parseTypedBPOrExprStartsWithOpenParenthesis() internal.STNode {
	exprOrTypeDesc := this.parseTypedDescOrExprStartsWithOpenParenthesis()
	if this.isDefiniteTypeDesc(exprOrTypeDesc.kind) {
		return this.parseTypeBindingPatternStartsWithAmbiguousNode(exprOrTypeDesc)
	}
	return this.parseTypedBindingPatternOrExprRhs(exprOrTypeDesc, false)
}

func (this *BallerinaParser) isDefiniteTypeDesc(kind SyntaxKind) bool {
	return ((this.kind.compareTo(SyntaxKind.RECORD_TYPE_DESC) >= 0) && (this.kind.compareTo(SyntaxKind.FUTURE_TYPE_DESC) <= 0))
}

func (this *BallerinaParser) isDefiniteExpr(kind SyntaxKind) bool {
	if (kind == SyntaxKind.QUALIFIED_NAME_REFERENCE) || (kind == SyntaxKind.SIMPLE_NAME_REFERENCE) {
		return false
	}
	return ((this.kind.compareTo(SyntaxKind.BINARY_EXPRESSION) >= 0) && (this.kind.compareTo(SyntaxKind.ERROR_CONSTRUCTOR) <= 0))
}

func (this *BallerinaParser) isDefiniteAction(kind SyntaxKind) bool {
	return ((this.kind.compareTo(SyntaxKind.REMOTE_METHOD_CALL_ACTION) >= 0) && (this.kind.compareTo(SyntaxKind.CLIENT_RESOURCE_ACCESS_ACTION) <= 0))
}

func (this *BallerinaParser) parseTypedDescOrExprStartsWithOpenParenthesis() internal.STNode {
	openParen := this.parseOpenParenthesis()
	nextToken := this.peek()
	if nextToken.kind == SyntaxKind.CLOSE_PAREN_TOKEN {
		closeParen := this.parseCloseParenthesis()
		return this.parseTypeOrExprStartWithEmptyParenthesis(openParen, closeParen)
	}
	typeOrExpr := this.parseTypeDescOrExpr()
	if this.isAction(typeOrExpr) {
		closeParen := this.parseCloseParenthesis()
		return this.STNodeFactory.createBracedExpressionNode(SyntaxKind.BRACED_ACTION, openParen, typeOrExpr,
			closeParen)
	}
	if this.isExpression(typeOrExpr.kind) {
		this.startContext(ParserRuleContext.BRACED_EXPR_OR_ANON_FUNC_PARAMS)
		return this.parseBracedExprOrAnonFuncParamRhs(openParen, typeOrExpr, false)
	}
	typeDescNode := this.getTypeDescFromExpr(typeOrExpr)
	typeDescNode = this.parseComplexTypeDescriptor(typeDescNode, ParserRuleContext.TYPE_DESC_IN_PARENTHESIS, false)
	closeParen := this.parseCloseParenthesis()
	return this.STNodeFactory.createParenthesisedTypeDescriptorNode(openParen, typeDescNode, closeParen)
}

func (this *BallerinaParser) parseTypeDescOrExpr() internal.STNode {
	typeDescQualifiers := make([]interface{}, 0)
	return this.parseTypeDescOrExpr(typeDescQualifiers)
}

func (this *BallerinaParser) parseTypeDescOrExpr(qualifiers []STNode) internal.STNode {
	this.parseTypeDescQualifiers(qualifiers)
	nextToken := this.peek()
	var typeOrExpr internal.STNode
	switch nextToken.kind {
	case OPEN_PAREN_TOKEN:
		this.reportInvalidQualifierList(qualifiers)
		typeOrExpr = this.parseTypedDescOrExprStartsWithOpenParenthesis()
		break
	case FUNCTION_KEYWORD:
		typeOrExpr = this.parseAnonFuncExprOrFuncTypeDesc(qualifiers)
		break
	case IDENTIFIER_TOKEN:
		this.reportInvalidQualifierList(qualifiers)
		typeOrExpr = this.parseQualifiedIdentifier(ParserRuleContext.TYPE_NAME_OR_VAR_NAME)
		return this.parseTypeDescOrExprRhs(typeOrExpr)
	case OPEN_BRACKET_TOKEN:
		this.reportInvalidQualifierList(qualifiers)
		typeOrExpr = this.parseTupleTypeDescOrListConstructor(STNodeFactory.createEmptyNodeList())
		break
	case DECIMAL_INTEGER_LITERAL_TOKEN:
	case HEX_INTEGER_LITERAL_TOKEN:
	case STRING_LITERAL_TOKEN:
	case NULL_KEYWORD:
	case TRUE_KEYWORD:
	case FALSE_KEYWORD:
	case DECIMAL_FLOATING_POINT_LITERAL_TOKEN:
	case HEX_FLOATING_POINT_LITERAL_TOKEN:
		this.reportInvalidQualifierList(qualifiers)
		basicLiteral := this.parseBasicLiteral()
		return this.parseTypeDescOrExprRhs(basicLiteral)
	default:
		if this.isValidExpressionStart(nextToken.kind, 1) {
			this.reportInvalidQualifierList(qualifiers)
			return this.parseActionOrExpressionInLhs(STNodeFactory.createEmptyNodeList())
		}
		return this.parseTypeDescriptor(qualifiers, ParserRuleContext.TYPE_DESC_IN_TYPE_BINDING_PATTERN)
	}
	if this.isDefiniteTypeDesc(typeOrExpr.kind) {
		return this.parseComplexTypeDescriptor(typeOrExpr, ParserRuleContext.TYPE_DESC_IN_TYPE_BINDING_PATTERN, true)
	}
	return this.parseTypeDescOrExprRhs(typeOrExpr)
}

func (this *BallerinaParser) isExpression(kind SyntaxKind) bool {
	switch kind {
	case NUMERIC_LITERAL,
		STRING_LITERAL_TOKEN,
		NIL_LITERAL,
		NULL_LITERAL,
		BOOLEAN_LITERAL:
		true
	default:
		((this.kind.compareTo(SyntaxKind.BINARY_EXPRESSION) >= 0) && (this.kind.compareTo(SyntaxKind.ERROR_CONSTRUCTOR) <= 0))
	}
}

func (this *BallerinaParser) parseTypeOrExprStartWithEmptyParenthesis(openParen internal.STNode, closeParen internal.STNode) internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case RIGHT_DOUBLE_ARROW_TOKEN:
		params := this.STNodeFactory.createEmptyNodeList()
		anonFuncParam := this.STNodeFactory.createImplicitAnonymousFunctionParameters(openParen, params, closeParen)
		return this.parseImplicitAnonFunc(anonFuncParam, false)
	default:
		return this.STNodeFactory.createNilLiteralNode(openParen, closeParen)
	}
}

func (this *BallerinaParser) parseAnonFuncExprOrTypedBPWithFuncType(qualifiers []STNode) internal.STNode {
	exprOrTypeDesc := this.parseAnonFuncExprOrFuncTypeDesc(qualifiers)
	if this.isAction(exprOrTypeDesc) || this.isExpression(exprOrTypeDesc.kind) {
		return exprOrTypeDesc
	}
	return this.parseTypedBindingPatternTypeRhs(exprOrTypeDesc, ParserRuleContext.VAR_DECL_STMT)
}

func (this *BallerinaParser) parseAnonFuncExprOrFuncTypeDesc(qualifiers []STNode) internal.STNode {
	this.startContext(ParserRuleContext.FUNC_TYPE_DESC_OR_ANON_FUNC)
	var qualifierList internal.STNode
	functionKeyword := this.parseFunctionKeyword()
	var funcSignature internal.STNode
	if peek().kind == SyntaxKind.OPEN_PAREN_TOKEN {
		funcSignature = this.parseFuncSignature(true)
		nodes := this.createFuncTypeQualNodeList(qualifiers, functionKeyword, true)
		qualifierList = nodes[0]
		functionKeyword = nodes[1]
		this.endContext()
		return this.parseAnonFuncExprOrFuncTypeDesc(qualifierList, functionKeyword, funcSignature)
	}
	funcSignature = this.STNodeFactory.createEmptyNode()
	nodes := this.createFuncTypeQualNodeList(qualifiers, functionKeyword, false)
	qualifierList = nodes[0]
	functionKeyword = nodes[1]
	funcTypeDesc := this.STNodeFactory.createFunctionTypeDescriptorNode(qualifierList, functionKeyword,
		funcSignature)
	if this.getCurrentContext() != ParserRuleContext.STMT_START_BRACKETED_LIST {
		this.switchContext(ParserRuleContext.VAR_DECL_STMT)
		return this.parseComplexTypeDescriptor(funcTypeDesc, ParserRuleContext.TYPE_DESC_IN_TYPE_BINDING_PATTERN, true)
	}
	return this.parseComplexTypeDescriptor(funcTypeDesc, ParserRuleContext.TYPE_DESC_IN_TUPLE, false)
}

func (this *BallerinaParser) parseAnonFuncExprOrFuncTypeDesc(qualifierList internal.STNode, functionKeyword internal.STNode, funcSignature internal.STNode) internal.STNode {
	currentCtx := this.getCurrentContext()
	switch peek().kind {
	case OPEN_BRACE_TOKEN:
	case RIGHT_DOUBLE_ARROW_TOKEN:
		if currentCtx != ParserRuleContext.STMT_START_BRACKETED_LIST {
			this.switchContext(ParserRuleContext.EXPRESSION_STATEMENT)
		}
		this.startContext(ParserRuleContext.ANON_FUNC_EXPRESSION)
		funcSignatureNode, ok := funcSignature.(*STFunctionSignatureNode)
		if !ok {
			panic("parseAnonFuncExprOrFuncTypeDesc: expected STFunctionSignatureNode")
		}
		funcSignature = this.validateAndGetFuncParams(funcSignatureNode)
		funcBody := this.parseAnonFuncBody(false)
		annots := this.STNodeFactory.createEmptyNodeList()
		anonFunc := this.STNodeFactory.createExplicitAnonymousFunctionExpressionNode(annots, qualifierList,
			functionKeyword, funcSignature, funcBody)
		return this.parseExpressionRhs(DEFAULT_OP_PRECEDENCE, anonFunc, false, true)
	case IDENTIFIER_TOKEN:
	default:
		funcTypeDesc := this.STNodeFactory.createFunctionTypeDescriptorNode(qualifierList, functionKeyword,
			funcSignature)
		if currentCtx != ParserRuleContext.STMT_START_BRACKETED_LIST {
			this.switchContext(ParserRuleContext.VAR_DECL_STMT)
			return this.parseComplexTypeDescriptor(funcTypeDesc, ParserRuleContext.TYPE_DESC_IN_TYPE_BINDING_PATTERN,
				true)
		}
		return this.parseComplexTypeDescriptor(funcTypeDesc, ParserRuleContext.TYPE_DESC_IN_TUPLE, false)
	}
}

func (this *BallerinaParser) parseTypeDescOrExprRhs(typeOrExpr internal.STNode) internal.STNode {
	nextToken := this.peek()
	var typeDesc internal.STNode
	switch nextToken.kind {
	case PIPE_TOKEN:
	case BITWISE_AND_TOKEN:
		nextNextToken := this.peek(2)
		if nextNextToken.kind == SyntaxKind.EQUAL_TOKEN {
			return typeOrExpr
		}
		pipeOrAndToken := this.parseBinaryOperator()
		rhsTypeDescOrExpr := this.parseTypeDescOrExpr()
		if this.isExpression(rhsTypeDescOrExpr.kind) {
			return this.STNodeFactory.createBinaryExpressionNode(SyntaxKind.BINARY_EXPRESSION, typeOrExpr,
				pipeOrAndToken, rhsTypeDescOrExpr)
		}
		typeDesc = this.getTypeDescFromExpr(typeOrExpr)
		rhsTypeDescOrExpr = this.getTypeDescFromExpr(rhsTypeDescOrExpr)
		return this.mergeTypes(typeDesc, pipeOrAndToken, rhsTypeDescOrExpr)
	case IDENTIFIER_TOKEN:
	case QUESTION_MARK_TOKEN:
		typeDesc = this.parseComplexTypeDescriptor(getTypeDescFromExpr(typeOrExpr),
			ParserRuleContext.TYPE_DESC_IN_TYPE_BINDING_PATTERN, false)
		return typeDesc
	case SEMICOLON_TOKEN:
		return this.getTypeDescFromExpr(typeOrExpr)
	case EQUAL_TOKEN:
	case CLOSE_PAREN_TOKEN:
	case CLOSE_BRACE_TOKEN:
	case CLOSE_BRACKET_TOKEN:
	case EOF_TOKEN:
	case COMMA_TOKEN:
		return typeOrExpr
	case OPEN_BRACKET_TOKEN:
		return this.parseTypedBindingPatternOrMemberAccess(typeOrExpr, false, true,
			ParserRuleContext.AMBIGUOUS_STMT)
	case ELLIPSIS_TOKEN:
		ellipsis := this.parseEllipsis()
		typeOrExpr = this.getTypeDescFromExpr(typeOrExpr)
		return this.STNodeFactory.createRestDescriptorNode(typeOrExpr, ellipsis)
	default:
		if this.isCompoundAssignment(nextToken.kind) {
			return typeOrExpr
		}
		if this.isValidExprRhsStart(nextToken.kind, typeOrExpr.kind) {
			return this.parseExpressionRhs(DEFAULT_OP_PRECEDENCE, typeOrExpr, false, false, false, false)
		}
		this.recover(peek(), ParserRuleContext.TYPE_DESC_OR_EXPR_RHS)
		return this.parseTypeDescOrExprRhs(typeOrExpr)
	}
}

func (this *BallerinaParser) isAmbiguous(node internal.STNode) bool {
	switch node.kind {
	case SIMPLE_NAME_REFERENCE:
	case QUALIFIED_NAME_REFERENCE:
	case NIL_LITERAL:
	case NULL_LITERAL:
	case NUMERIC_LITERAL:
	case STRING_LITERAL:
	case BOOLEAN_LITERAL:
	case BRACKETED_LIST:
		return true
	case BINARY_EXPRESSION:
		binaryExpr := internal.STBinaryExpressionNode(node)
		if binaryExpr.operator.kind != SyntaxKind.PIPE_TOKEN {
			return false
		}
		return (this.isAmbiguous(binaryExpr.lhsExpr) && this.isAmbiguous(binaryExpr.rhsExpr))
	case BRACED_EXPRESSION:
		bracedExpr, ok := node.(*STBracedExpressionNode)
		if !ok {
			panic("isAmbiguous: expected STBracedExpressionNode")
		}
		return this.isAmbiguous(bracedExpr.expression)
	case INDEXED_EXPRESSION:
		indexExpr := internal.STIndexedExpressionNode(node)
		if !this.isAmbiguous(indexExpr.containerExpression) {
			return false
		}
		keys := indexExpr.keyExpression
		i := 0
		for ; i < this.keys.bucketCount(); i++ {
			item := this.keys.childInBucket(i)
			if item.kind == SyntaxKind.COMMA_TOKEN {
				continue
			}
			if !this.isAmbiguous(item) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func (this *BallerinaParser) isAllBasicLiterals(node internal.STNode) bool {
	switch node.kind {
	case NIL_LITERAL:
	case NULL_LITERAL:
	case NUMERIC_LITERAL:
	case STRING_LITERAL:
	case BOOLEAN_LITERAL:
		return true
	case BINARY_EXPRESSION:
		binaryExpr := internal.STBinaryExpressionNode(node)
		if binaryExpr.operator.kind != SyntaxKind.PIPE_TOKEN {
			return false
		}
		return (this.isAmbiguous(binaryExpr.lhsExpr) && this.isAmbiguous(binaryExpr.rhsExpr))
	case BRACED_EXPRESSION:
		bracedExpr, ok := node.(*STBracedExpressionNode)
		if !ok {
			panic("isAllBasicLiterals: expected STBracedExpressionNode")
		}
		return this.isAmbiguous(bracedExpr.expression)
	case BRACKETED_LIST:
		list := internal.STAmbiguousCollectionNode(node)
		for _, member := range list.members {
			if member.kind == SyntaxKind.COMMA_TOKEN {
				continue
			}
			if !this.isAllBasicLiterals(member) {
				return false
			}
		}
		return true
	case UNARY_EXPRESSION:
		unaryExpr := internal.STUnaryExpressionNode(node)
		if (unaryExpr.unaryOperator.kind != SyntaxKind.PLUS_TOKEN) && (unaryExpr.unaryOperator.kind != SyntaxKind.MINUS_TOKEN) {
			return false
		}
		return this.isNumericLiteral(unaryExpr.expression)
	default:
		return false
	}
}

func (this *BallerinaParser) isNumericLiteral(node internal.STNode) bool {
	switch node.kind {
	case NUMERIC_LITERAL:
		true
	default:
		false
	}
}

func (this *BallerinaParser) parseBindingPattern() internal.STNode {
	switch peek().kind {
	case OPEN_BRACKET_TOKEN:
		this.parseListBindingPattern()
	case IDENTIFIER_TOKEN:
		this.parseBindingPatternStartsWithIdentifier()
	case OPEN_BRACE_TOKEN:
		this.parseMappingBindingPattern()
	case ERROR_KEYWORD:
		this.parseErrorBindingPattern()
	default:
		this.recover(peek(), ParserRuleContext.BINDING_PATTERN)
		this.parseBindingPattern()
	}
}

func (this *BallerinaParser) parseBindingPatternStartsWithIdentifier() internal.STNode {
	argNameOrBindingPattern := this.parseQualifiedIdentifier(ParserRuleContext.BINDING_PATTERN_STARTING_IDENTIFIER)
	secondToken := this.peek()
	if secondToken.kind == SyntaxKind.OPEN_PAREN_TOKEN {
		this.startContext(ParserRuleContext.ERROR_BINDING_PATTERN)
		errorKeyword := this.SyntaxErrors.createMissingTokenWithDiagnostics(SyntaxKind.ERROR_KEYWORD,
			ParserRuleContext.ERROR_KEYWORD)
		return this.parseErrorBindingPattern(errorKeyword, argNameOrBindingPattern)
	}
	if argNameOrBindingPattern.kind != SyntaxKind.SIMPLE_NAME_REFERENCE {
		identifier := this.SyntaxErrors.createMissingToken(SyntaxKind.IDENTIFIER_TOKEN)
		identifier = this.SyntaxErrors.cloneWithLeadingInvalidNodeMinutiae(identifier, argNameOrBindingPattern,
			DiagnosticErrorCode.ERROR_FIELD_BP_INSIDE_LIST_BP)
		return this.STNodeFactory.createCaptureBindingPatternNode(identifier)
	}
	simpleNameNode, ok := argNameOrBindingPattern.(*STSimpleNameReferenceNode)
	if !ok {
		panic("parseBindingPatternStartsWithIdentifier: expected STSimpleNameReferenceNode")
	}
	return this.createCaptureOrWildcardBP(simpleNameNode.name)
}

func (this *BallerinaParser) createCaptureOrWildcardBP(varName internal.STNode) internal.STNode {
	var bindingPattern internal.STNode
	if this.isWildcardBP(varName) {
		bindingPattern = this.getWildcardBindingPattern(varName)
	} else {
		bindingPattern = this.STNodeFactory.createCaptureBindingPatternNode(varName)
	}
	return bindingPattern
}

func (this *BallerinaParser) parseListBindingPattern() internal.STNode {
	this.startContext(ParserRuleContext.LIST_BINDING_PATTERN)
	openBracket := this.parseOpenBracket()
	bindingPatternsList := make([]interface{}, 0)
	listBindingPattern := this.parseListBindingPattern(openBracket, bindingPatternsList)
	this.endContext()
	return listBindingPattern
}

func (this *BallerinaParser) parseListBindingPattern(openBracket internal.STNode, bindingPatternsList []STNode) internal.STNode {
	if this.isEndOfListBindingPattern(peek().kind) && this.bindingPatternsList.isEmpty() {
		closeBracket := this.parseCloseBracket()
		bindingPatternsNode := this.STNodeFactory.createNodeList(bindingPatternsList)
		return this.STNodeFactory.createListBindingPatternNode(openBracket, bindingPatternsNode, closeBracket)
	}
	listBindingPatternMember := this.parseListBindingPatternMember()
	this.bindingPatternsList.add(listBindingPatternMember)
	listBindingPattern := this.parseListBindingPattern(openBracket, listBindingPatternMember, bindingPatternsList)
	return listBindingPattern
}

func (this *BallerinaParser) parseListBindingPattern(openBracket internal.STNode, firstMember internal.STNode, bindingPatterns []STNode) internal.STNode {
	member := firstMember
	token := this.peek()
	listBindingPatternRhs := nil
	for (!this.isEndOfListBindingPattern(token.kind)) && (member.kind != SyntaxKind.REST_BINDING_PATTERN) {
		listBindingPatternRhs = this.parseListBindingPatternMemberRhs()
		if listBindingPatternRhs == nil {
			break
		}
		this.bindingPatterns.add(listBindingPatternRhs)
		member = this.parseListBindingPatternMember()
		this.bindingPatterns.add(member)
		token = this.peek()
	}
	closeBracket := this.parseCloseBracket()
	bindingPatternsNode := this.STNodeFactory.createNodeList(bindingPatterns)
	return this.STNodeFactory.createListBindingPatternNode(openBracket, bindingPatternsNode, closeBracket)
}

func (this *BallerinaParser) parseListBindingPatternMemberRhs() internal.STNode {
	switch peek().kind {
	case COMMA_TOKEN:
		this.parseComma()
	case CLOSE_BRACKET_TOKEN:
		nil
	default:
		this.recover(peek(), ParserRuleContext.LIST_BINDING_PATTERN_MEMBER_END)
		this.parseListBindingPatternMemberRhs()
	}
}

func (this *BallerinaParser) isEndOfListBindingPattern(nextTokenKind SyntaxKind) bool {
	switch nextTokenKind {
	case CLOSE_BRACKET_TOKEN, EOF_TOKEN:
		true
	default:
		false
	}
}

func (this *BallerinaParser) parseListBindingPatternMember() internal.STNode {
	switch peek().kind {
	case ELLIPSIS_TOKEN:
		this.parseRestBindingPattern()
	case OPEN_BRACKET_TOKEN,
		IDENTIFIER_TOKEN,
		OPEN_BRACE_TOKEN,
		ERROR_KEYWORD:
		this.parseBindingPattern()
	default:
		this.recover(peek(), ParserRuleContext.LIST_BINDING_PATTERN_MEMBER)
		this.parseListBindingPatternMember()
	}
}

func (this *BallerinaParser) parseRestBindingPattern() internal.STNode {
	this.startContext(ParserRuleContext.REST_BINDING_PATTERN)
	ellipsis := this.parseEllipsis()
	varName := this.parseVariableName()
	this.endContext()
	simpleNameReferenceNode := internal.STSimpleNameReferenceNode(this.STNodeFactory.createSimpleNameReferenceNode(varName))
	return this.STNodeFactory.createRestBindingPatternNode(ellipsis, simpleNameReferenceNode)
}

func (this *BallerinaParser) parseTypedBindingPattern(context ParserRuleContext) internal.STNode {
	typeDescQualifiers := make([]interface{}, 0)
	return this.parseTypedBindingPattern(typeDescQualifiers, context)
}

func (this *BallerinaParser) parseTypedBindingPattern(qualifiers []STNode, context ParserRuleContext) internal.STNode {
	typeDesc := this.parseTypeDescriptor(qualifiers,
		ParserRuleContext.TYPE_DESC_IN_TYPE_BINDING_PATTERN, true, false, TypePrecedence.DEFAULT)
	typeBindingPattern := this.parseTypedBindingPatternTypeRhs(typeDesc, context)
	return typeBindingPattern
}

func (this *BallerinaParser) parseMappingBindingPattern() internal.STNode {
	this.startContext(ParserRuleContext.MAPPING_BINDING_PATTERN)
	openBrace := this.parseOpenBrace()
	token := this.peek()
	if this.isEndOfMappingBindingPattern(token.kind) {
		closeBrace := this.parseCloseBrace()
		bindingPatternsNode := this.STNodeFactory.createEmptyNodeList()
		this.endContext()
		return this.STNodeFactory.createMappingBindingPatternNode(openBrace, bindingPatternsNode, closeBrace)
	}
	bindingPatterns := make([]interface{}, 0)
	prevMember := this.parseMappingBindingPatternMember()
	if prevMember.kind != SyntaxKind.REST_BINDING_PATTERN {
		this.bindingPatterns.add(prevMember)
	}
	return this.parseMappingBindingPattern(openBrace, bindingPatterns, prevMember)
}

func (this *BallerinaParser) parseMappingBindingPattern(openBrace internal.STNode, bindingPatterns []STNode, prevMember internal.STNode) internal.STNode {
	token := this.peek()
	mappingBindingPatternRhs := nil
	for (!this.isEndOfMappingBindingPattern(token.kind)) && (prevMember.kind != SyntaxKind.REST_BINDING_PATTERN) {
		mappingBindingPatternRhs = this.parseMappingBindingPatternEnd()
		if mappingBindingPatternRhs == nil {
			break
		}
		this.bindingPatterns.add(mappingBindingPatternRhs)
		prevMember = this.parseMappingBindingPatternMember()
		if prevMember.kind == SyntaxKind.REST_BINDING_PATTERN {
			break
		}
		this.bindingPatterns.add(prevMember)
		token = this.peek()
	}
	if prevMember.kind == SyntaxKind.REST_BINDING_PATTERN {
		this.bindingPatterns.add(prevMember)
	}
	closeBrace := this.parseCloseBrace()
	bindingPatternsNode := this.STNodeFactory.createNodeList(bindingPatterns)
	this.endContext()
	return this.STNodeFactory.createMappingBindingPatternNode(openBrace, bindingPatternsNode, closeBrace)
}

func (this *BallerinaParser) parseMappingBindingPatternMember() internal.STNode {
	token := this.peek()
	switch token.kind {
	case ELLIPSIS_TOKEN:
		this.parseRestBindingPattern()
	default:
		this.parseFieldBindingPattern()
	}
}

func (this *BallerinaParser) parseMappingBindingPatternEnd() internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case COMMA_TOKEN:
		this.parseComma()
	case CLOSE_BRACE_TOKEN:
		nil
	default:
		this.recover(nextToken, ParserRuleContext.MAPPING_BINDING_PATTERN_END)
		this.parseMappingBindingPatternEnd()
	}
}

func (this *BallerinaParser) parseFieldBindingPattern() internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case IDENTIFIER_TOKEN:
		identifier := this.parseIdentifier(ParserRuleContext.FIELD_BINDING_PATTERN_NAME)
		simpleNameReference := this.STNodeFactory.createSimpleNameReferenceNode(identifier)
		return this.parseFieldBindingPattern(simpleNameReference)
	default:
		this.recover(nextToken, ParserRuleContext.FIELD_BINDING_PATTERN_NAME)
		return this.parseFieldBindingPattern()
	}
}

func (this *BallerinaParser) parseFieldBindingPattern(simpleNameReference internal.STNode) internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case COMMA_TOKEN:
	case CLOSE_BRACE_TOKEN:
		return this.STNodeFactory.createFieldBindingPatternVarnameNode(simpleNameReference)
	case COLON_TOKEN:
		colon := this.parseColon()
		bindingPattern := this.parseBindingPattern()
		return this.STNodeFactory.createFieldBindingPatternFullNode(simpleNameReference, colon, bindingPattern)
	default:
		this.recover(nextToken, ParserRuleContext.FIELD_BINDING_PATTERN_END)
		return this.parseFieldBindingPattern(simpleNameReference)
	}
}

func (this *BallerinaParser) isEndOfMappingBindingPattern(nextTokenKind SyntaxKind) bool {
	return ((nextTokenKind == SyntaxKind.CLOSE_BRACE_TOKEN) || this.isEndOfModuleLevelNode(1))
}

func (this *BallerinaParser) parseErrorTypeDescOrErrorBP(annots internal.STNode) internal.STNode {
	nextNextToken := this.peek(2)
	switch nextNextToken.kind {
	case OPEN_PAREN_TOKEN:
		return this.parseAsErrorBindingPattern()
	case LT_TOKEN:
		return this.parseAsErrorTypeDesc(annots)
	case IDENTIFIER_TOKEN:
		nextNextNextTokenKind := peek(3).kind
		if (nextNextNextTokenKind == SyntaxKind.COLON_TOKEN) || (nextNextNextTokenKind == SyntaxKind.OPEN_PAREN_TOKEN) {
			return this.parseAsErrorBindingPattern()
		}
	default:
		return this.parseAsErrorTypeDesc(annots)
	}
}

func (this *BallerinaParser) parseAsErrorBindingPattern() internal.STNode {
	this.startContext(ParserRuleContext.ASSIGNMENT_STMT)
	return this.parseAssignmentStmtRhs(parseErrorBindingPattern())
}

func (this *BallerinaParser) parseAsErrorTypeDesc(annots internal.STNode) internal.STNode {
	finalKeyword := this.STNodeFactory.createEmptyNode()
	return this.parseVariableDecl(getAnnotations(annots), finalKeyword)
}

func (this *BallerinaParser) parseErrorBindingPattern() internal.STNode {
	this.startContext(ParserRuleContext.ERROR_BINDING_PATTERN)
	errorKeyword := this.parseErrorKeyword()
	return this.parseErrorBindingPattern(errorKeyword)
}

func (this *BallerinaParser) parseErrorBindingPattern(errorKeyword internal.STNode) internal.STNode {
	nextToken := this.peek()
	var typeRef internal.STNode
	switch nextToken.kind {
	case OPEN_PAREN_TOKEN:
		typeRef = this.STNodeFactory.createEmptyNode()
		break
	default:
		if this.isPredeclaredIdentifier(nextToken.kind) {
			typeRef = this.parseTypeReference()
			break
		}
		this.recover(peek(), ParserRuleContext.ERROR_BINDING_PATTERN_ERROR_KEYWORD_RHS)
		return this.parseErrorBindingPattern(errorKeyword)
	}
	return this.parseErrorBindingPattern(errorKeyword, typeRef)
}

func (this *BallerinaParser) parseErrorBindingPattern(errorKeyword internal.STNode, typeRef internal.STNode) internal.STNode {
	openParenthesis := this.parseOpenParenthesis()
	argListBindingPatterns := this.parseErrorArgListBindingPatterns()
	closeParenthesis := this.parseCloseParenthesis()
	this.endContext()
	return this.STNodeFactory.createErrorBindingPatternNode(errorKeyword, typeRef, openParenthesis,
		argListBindingPatterns, closeParenthesis)
}

func (this *BallerinaParser) parseErrorArgListBindingPatterns() internal.STNode {
	argListBindingPatterns := make([]interface{}, 0)
	if this.isEndOfErrorFieldBindingPatterns() {
		return this.STNodeFactory.createNodeList(argListBindingPatterns)
	}
	return this.parseErrorArgListBindingPatterns(argListBindingPatterns)
}

func (this *BallerinaParser) parseErrorArgListBindingPatterns(argListBindingPatterns []STNode) internal.STNode {
	firstArg := this.parseErrorArgListBindingPattern(ParserRuleContext.ERROR_ARG_LIST_BINDING_PATTERN_START, true)
	if firstArg == nil {
		return this.STNodeFactory.createNodeList(argListBindingPatterns)
	}
	switch firstArg.kind {
	case CAPTURE_BINDING_PATTERN:
	case WILDCARD_BINDING_PATTERN:
		this.argListBindingPatterns.add(firstArg)
		return this.parseErrorArgListBPWithoutErrorMsg(argListBindingPatterns)
	case ERROR_BINDING_PATTERN:
		missingIdentifier := this.SyntaxErrors.createMissingToken(SyntaxKind.IDENTIFIER_TOKEN)
		missingErrorMsgBP := this.STNodeFactory.createCaptureBindingPatternNode(missingIdentifier)
		missingErrorMsgBP = this.SyntaxErrors.addDiagnostic(missingErrorMsgBP,
			DiagnosticErrorCode.ERROR_MISSING_ERROR_MESSAGE_BINDING_PATTERN)
		missingComma := this.SyntaxErrors.createMissingTokenWithDiagnostics(SyntaxKind.COMMA_TOKEN,
			DiagnosticErrorCode.ERROR_MISSING_COMMA_TOKEN)
		this.argListBindingPatterns.add(missingErrorMsgBP)
		this.argListBindingPatterns.add(missingComma)
		this.argListBindingPatterns.add(firstArg)
		return this.parseErrorArgListBPWithoutErrorMsgAndCause(argListBindingPatterns, firstArg.kind)
	case REST_BINDING_PATTERN:
	case NAMED_ARG_BINDING_PATTERN:
		this.argListBindingPatterns.add(firstArg)
		return this.parseErrorArgListBPWithoutErrorMsgAndCause(argListBindingPatterns, firstArg.kind)
	default:
		this.addInvalidNodeToNextToken(firstArg, DiagnosticErrorCode.ERROR_BINDING_PATTERN_NOT_ALLOWED)
		return this.parseErrorArgListBindingPatterns(argListBindingPatterns)
	}
}

func (this *BallerinaParser) parseErrorArgListBPWithoutErrorMsg(argListBindingPatterns []STNode) internal.STNode {
	argEnd := this.parseErrorArgsBindingPatternEnd(ParserRuleContext.ERROR_MESSAGE_BINDING_PATTERN_END)
	if argEnd == nil {
		return this.STNodeFactory.createNodeList(argListBindingPatterns)
	}
	secondArg := this.parseErrorArgListBindingPattern(ParserRuleContext.ERROR_MESSAGE_BINDING_PATTERN_RHS, false)
	if secondArg != nil {
		panic("assertion failed")
	}
	switch secondArg.kind {
	case CAPTURE_BINDING_PATTERN:
	case WILDCARD_BINDING_PATTERN:
	case ERROR_BINDING_PATTERN:
	case REST_BINDING_PATTERN:
	case NAMED_ARG_BINDING_PATTERN:
		this.argListBindingPatterns.add(argEnd)
		this.argListBindingPatterns.add(secondArg)
		return this.parseErrorArgListBPWithoutErrorMsgAndCause(argListBindingPatterns, secondArg.kind)
	default:
		this.updateLastNodeInListWithInvalidNode(argListBindingPatterns, argEnd, null)
		this.updateLastNodeInListWithInvalidNode(argListBindingPatterns, secondArg,
			DiagnosticErrorCode.ERROR_BINDING_PATTERN_NOT_ALLOWED)
		return this.parseErrorArgListBPWithoutErrorMsg(argListBindingPatterns)
	}
}

func (this *BallerinaParser) parseErrorArgListBPWithoutErrorMsgAndCause(argListBindingPatterns []STNode, lastValidArgKind SyntaxKind) internal.STNode {
	for !this.isEndOfErrorFieldBindingPatterns() {
		argEnd := this.parseErrorArgsBindingPatternEnd(ParserRuleContext.ERROR_FIELD_BINDING_PATTERN_END)
		if argEnd == nil {
			break
		}
		currentArg := this.parseErrorArgListBindingPattern(ParserRuleContext.ERROR_FIELD_BINDING_PATTERN, false)
		if currentArg != nil {
			panic("assertion failed")
		}
		errorCode := this.validateErrorFieldBindingPatternOrder(lastValidArgKind, currentArg.kind)
		if errorCode == nil {
			this.argListBindingPatterns.add(argEnd)
			this.argListBindingPatterns.add(currentArg)
			lastValidArgKind = currentArg.kind
		} else if this.argListBindingPatterns.isEmpty() {
			this.addInvalidNodeToNextToken(argEnd, null)
			this.addInvalidNodeToNextToken(currentArg, errorCode)
		}
	}
	return this.STNodeFactory.createNodeList(argListBindingPatterns)
}

func (this *BallerinaParser) isEndOfErrorFieldBindingPatterns() bool {
	nextTokenKind := peek().kind
	switch nextTokenKind {
	case CLOSE_PAREN_TOKEN,
		EOF_TOKEN:
		true
	default:
		false
	}
}

func (this *BallerinaParser) parseErrorArgsBindingPatternEnd(currentCtx ParserRuleContext) internal.STNode {
	switch peek().kind {
	case COMMA_TOKEN:
		this.consume()
	case CLOSE_PAREN_TOKEN:
		nil
	default:
		this.recover(peek(), currentCtx)
		this.parseErrorArgsBindingPatternEnd(currentCtx)
	}
}

func (this *BallerinaParser) parseErrorArgListBindingPattern(context ParserRuleContext, isFirstArg bool) internal.STNode {
	switch peek().kind {
	case ELLIPSIS_TOKEN:
		return this.parseRestBindingPattern()
	case IDENTIFIER_TOKEN:
		argNameOrSimpleBindingPattern := this.consume()
		return this.parseNamedOrSimpleArgBindingPattern(argNameOrSimpleBindingPattern)
	case OPEN_BRACKET_TOKEN:
	case OPEN_BRACE_TOKEN:
	case ERROR_KEYWORD:
		return this.parseBindingPattern()
	case CLOSE_PAREN_TOKEN:
		if isFirstArg {
			return nil
		}
	default:
		this.recover(peek(), context)
		return this.parseErrorArgListBindingPattern(context, isFirstArg)
	}
}

func (this *BallerinaParser) parseNamedOrSimpleArgBindingPattern(argNameOrSimpleBindingPattern internal.STNode) internal.STNode {
	secondToken := this.peek()
	switch secondToken.kind {
	case EQUAL_TOKEN:
		equal := this.consume()
		bindingPattern := this.parseBindingPattern()
		return this.STNodeFactory.createNamedArgBindingPatternNode(argNameOrSimpleBindingPattern,
			equal, bindingPattern)
	case COMMA_TOKEN:
	case CLOSE_PAREN_TOKEN:
	default:
		return this.createCaptureOrWildcardBP(argNameOrSimpleBindingPattern)
	}
}

func (this *BallerinaParser) validateErrorFieldBindingPatternOrder(prevArgKind SyntaxKind, currentArgKind SyntaxKind) DiagnosticErrorCode {
	switch currentArgKind {
	case NAMED_ARG_BINDING_PATTERN,
		REST_BINDING_PATTERN:
		if prevArgKind == SyntaxKind.REST_BINDING_PATTERN {
			DiagnosticErrorCode.ERROR_REST_ARG_FOLLOWED_BY_ANOTHER_ARG
		}
		nil
	default:
		DiagnosticErrorCode.ERROR_BINDING_PATTERN_NOT_ALLOWED
	}
}

func (this *BallerinaParser) parseTypedBindingPatternTypeRhs(typeDesc internal.STNode, context ParserRuleContext) internal.STNode {
	return this.parseTypedBindingPatternTypeRhs(typeDesc, context, true)
}

func (this *BallerinaParser) parseTypedBindingPatternTypeRhs(typeDesc internal.STNode, context ParserRuleContext, isRoot bool) internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case IDENTIFIER_TOKEN:
	case OPEN_BRACE_TOKEN:
	case ERROR_KEYWORD:
		bindingPattern := this.parseBindingPattern()
		return this.STNodeFactory.createTypedBindingPatternNode(typeDesc, bindingPattern)
	case OPEN_BRACKET_TOKEN:
		typedBindingPattern := this.parseTypedBindingPatternOrMemberAccess(typeDesc, true, true, context)
		if typedBindingPattern.kind == SyntaxKind.TYPED_BINDING_PATTERN {
			panic("assertion failed")
		}
		return typedBindingPattern
	case CLOSE_PAREN_TOKEN:
	case COMMA_TOKEN:
	case CLOSE_BRACKET_TOKEN:
	case CLOSE_BRACE_TOKEN:
		if !isRoot {
			return typeDesc
		}
	default:
		this.recover(nextToken, ParserRuleContext.TYPED_BINDING_PATTERN_TYPE_RHS)
		return this.parseTypedBindingPatternTypeRhs(typeDesc, context, isRoot)
	}
}

func (this *BallerinaParser) parseTypedBindingPatternOrMemberAccess(typeDescOrExpr internal.STNode, isTypedBindingPattern bool, allowAssignment bool, context ParserRuleContext) internal.STNode {
	this.startContext(ParserRuleContext.BRACKETED_LIST)
	openBracket := this.parseOpenBracket()
	if this.isBracketedListEnd(peek().kind) {
		return this.parseAsArrayTypeDesc(typeDescOrExpr, openBracket, STNodeFactory.createEmptyNode(), context)
	}
	member := this.parseBracketedListMember(isTypedBindingPattern)
	currentNodeType := this.getBracketedListNodeType(member, isTypedBindingPattern)
	switch currentNodeType {
	case ARRAY_TYPE_DESC:
		typedBindingPattern := this.parseAsArrayTypeDesc(typeDescOrExpr, openBracket, member, context)
		return typedBindingPattern
	case LIST_BINDING_PATTERN:
		bindingPattern := this.parseAsListBindingPattern(openBracket, nil, member, false)
		typeDesc := this.getTypeDescFromExpr(typeDescOrExpr)
		return this.STNodeFactory.createTypedBindingPatternNode(typeDesc, bindingPattern)
	case INDEXED_EXPRESSION:
		return this.parseAsMemberAccessExpr(typeDescOrExpr, openBracket, member)
	case ARRAY_TYPE_DESC_OR_MEMBER_ACCESS:
		break
	case NONE:
	default:
		memberEnd := this.parseBracketedListMemberEnd()
		if memberEnd != nil {
			memberList := make([]interface{}, 0)
			this.memberList.add(getBindingPattern(member, true))
			this.memberList.add(memberEnd)
			bindingPattern = this.parseAsListBindingPattern(openBracket, memberList)
			typeDesc = this.getTypeDescFromExpr(typeDescOrExpr)
			return this.STNodeFactory.createTypedBindingPatternNode(typeDesc, bindingPattern)
		}
	}
	closeBracket := this.parseCloseBracket()
	this.endContext()
	return this.parseTypedBindingPatternOrMemberAccessRhs(typeDescOrExpr, openBracket, member, closeBracket,
		isTypedBindingPattern, allowAssignment, context)
}

func (this *BallerinaParser) parseAsMemberAccessExpr(typeNameOrExpr internal.STNode, openBracket internal.STNode, member internal.STNode) internal.STNode {
	member = this.parseExpressionRhs(DEFAULT_OP_PRECEDENCE, member, false, true)
	closeBracket := this.parseCloseBracket()
	this.endContext()
	keyExpr := this.STNodeFactory.createNodeList(member)
	memberAccessExpr := this.STNodeFactory.createIndexedExpressionNode(typeNameOrExpr, openBracket, keyExpr, closeBracket)
	return this.parseExpressionRhs(DEFAULT_OP_PRECEDENCE, memberAccessExpr, false, false)
}

func (this *BallerinaParser) isBracketedListEnd(nextTokenKind SyntaxKind) bool {
	switch nextTokenKind {
	case EOF_TOKEN, CLOSE_BRACKET_TOKEN:
		true
	default:
		false
	}
}

// func (this *BallerinaParser) parseBracketedListMember(isTypedBindingPattern bool) internal.STNode {
// nextToken := this.peek()
//
//	switch nextToken.kind {
//		case DECIMAL_INTEGER_LITERAL_TOKEN:
//		case HEX_INTEGER_LITERAL_TOKEN:
//		case ASTERISK_TOKEN:
//		case STRING_LITERAL_TOKEN:
//			return this.parseBasicLiteral()
//		case CLOSE_BRACKET_TOKEN:
//			return this.STNodeFactory.createEmptyNode()
//		case OPEN_BRACE_TOKEN:
//		case ERROR_KEYWORD:
//		case ELLIPSIS_TOKEN:
//		case OPEN_BRACKET_TOKEN:
//			return this.parseStatementStartBracketedListMember()
//		case IDENTIFIER_TOKEN:
//			if isTypedBindingPattern {
//				return this.parseQualifiedIdentifier(ParserRuleContext.VARIABLE_REF)
//			}
//		default:
//			if (((!isTypedBindingPattern) && this.isValidExpressionStart(nextToken.kind, 1)) || this.isQualifiedIdentifierPredeclaredPrefix(nextToken.kind)) {
//			// break;
//			}
//			var recoverContext ParserRuleContext
//			if isTypedBindingPattern {
//				recoverContext = ParserRuleContext.LIST_BINDING_MEMBER_OR_ARRAY_LENGTH
//			} else {
//				recoverContext = ParserRuleContext.BRACKETED_LIST_MEMBER
//			this.recover(peek(), recoverContext)
//			return this.parseBracketedListMember(isTypedBindingPattern)
//			}
//			expr := this.parseExpression()
//			if this.isWildcardBP(expr) {
//			return this.getWildcardBindingPattern(expr)
//			}
//			return expr
//	}
func (this *BallerinaParser) parseAsArrayTypeDesc(typeDesc internal.STNode, openBracket internal.STNode, member internal.STNode, context ParserRuleContext) internal.STNode {
	typeDesc = this.getTypeDescFromExpr(typeDesc)
	this.switchContext(ParserRuleContext.TYPE_DESC_IN_TYPE_BINDING_PATTERN)
	this.startContext(ParserRuleContext.ARRAY_TYPE_DESCRIPTOR)
	closeBracket := this.parseCloseBracket()
	this.endContext()
	this.endContext()
	return this.parseTypedBindingPatternOrMemberAccessRhs(typeDesc, openBracket, member, closeBracket, true, true,
		context)
}
func (this *BallerinaParser) parseBracketedListMemberEnd() internal.STNode {
	switch peek().kind {
	case COMMA_TOKEN:
		this.parseComma()
	case CLOSE_BRACKET_TOKEN:
		nil
	default:
		this.recover(peek(), ParserRuleContext.BRACKETED_LIST_MEMBER_END)
		this.parseBracketedListMemberEnd()
	}
}

func (this *BallerinaParser) parseTypedBindingPatternOrMemberAccessRhs(typeDescOrExpr internal.STNode, openBracket internal.STNode, member internal.STNode, closeBracket internal.STNode, isTypedBindingPattern bool, allowAssignment bool, context ParserRuleContext) internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case IDENTIFIER_TOKEN:
	case OPEN_BRACE_TOKEN:
	case ERROR_KEYWORD:
		typeDesc := this.getTypeDescFromExpr(typeDescOrExpr)
		arrayTypeDesc := this.getArrayTypeDesc(openBracket, member, closeBracket, typeDesc)
		return this.parseTypedBindingPatternTypeRhs(arrayTypeDesc, context)
	case OPEN_BRACKET_TOKEN:
		if isTypedBindingPattern {
			typeDesc = this.getTypeDescFromExpr(typeDescOrExpr)
			arrayTypeDesc = this.getArrayTypeDesc(openBracket, member, closeBracket, typeDesc)
			return this.parseTypedBindingPatternTypeRhs(arrayTypeDesc, context)
		}
		keyExpr := this.getKeyExpr(member)
		expr := this.STNodeFactory.createIndexedExpressionNode(typeDescOrExpr, openBracket, keyExpr, closeBracket)
		return this.parseTypedBindingPatternOrMemberAccess(expr, false, allowAssignment, context)
	case QUESTION_MARK_TOKEN:
		typeDesc = this.getTypeDescFromExpr(typeDescOrExpr)
		arrayTypeDesc = this.getArrayTypeDesc(openBracket, member, closeBracket, typeDesc)
		typeDesc = this.parseComplexTypeDescriptor(arrayTypeDesc,
			ParserRuleContext.TYPE_DESC_IN_TYPE_BINDING_PATTERN, true)
		return this.parseTypedBindingPatternTypeRhs(typeDesc, context)
	case PIPE_TOKEN:
	case BITWISE_AND_TOKEN:
		return this.parseComplexTypeDescInTypedBPOrExprRhs(typeDescOrExpr, openBracket, member, closeBracket,
			isTypedBindingPattern)
	case IN_KEYWORD:
		if ((context != ParserRuleContext.FOREACH_STMT) && (context != ParserRuleContext.FROM_CLAUSE)) && (context != ParserRuleContext.JOIN_CLAUSE) {
			break
		}
		return this.createTypedBindingPattern(typeDescOrExpr, openBracket, member, closeBracket)
	case EQUAL_TOKEN:
		if (context == ParserRuleContext.FOREACH_STMT) || (context == ParserRuleContext.FROM_CLAUSE) {
			break
		}
		if (isTypedBindingPattern || (!allowAssignment)) || (!this.isValidLVExpr(typeDescOrExpr)) {
			return this.createTypedBindingPattern(typeDescOrExpr, openBracket, member, closeBracket)
		}
		keyExpr = this.getKeyExpr(member)
		typeDescOrExpr = this.getExpression(typeDescOrExpr)
		return this.STNodeFactory.createIndexedExpressionNode(typeDescOrExpr, openBracket, keyExpr, closeBracket)
	case SEMICOLON_TOKEN:
		if (context == ParserRuleContext.FOREACH_STMT) || (context == ParserRuleContext.FROM_CLAUSE) {
			break
		}
		return this.createTypedBindingPattern(typeDescOrExpr, openBracket, member, closeBracket)
	case CLOSE_BRACE_TOKEN:
	case COMMA_TOKEN:
		if context == ParserRuleContext.AMBIGUOUS_STMT {
			keyExpr = this.getKeyExpr(member)
			return this.STNodeFactory.createIndexedExpressionNode(typeDescOrExpr, openBracket, keyExpr,
				closeBracket)
		}
	default:
		if (!isTypedBindingPattern) && this.isValidExprRhsStart(nextToken.kind, closeBracket.kind) {
			keyExpr = this.getKeyExpr(member)
			typeDescOrExpr = this.getExpression(typeDescOrExpr)
			return this.STNodeFactory.createIndexedExpressionNode(typeDescOrExpr, openBracket, keyExpr,
				closeBracket)
		}
		break
	}
	recoveryCtx := ParserRuleContext.BRACKETED_LIST_RHS
	if isTypedBindingPattern {
		recoveryCtx = ParserRuleContext.TYPE_DESC_RHS_OR_BP_RHS
	}
	this.recover(peek(), recoveryCtx)
	return this.parseTypedBindingPatternOrMemberAccessRhs(typeDescOrExpr, openBracket, member, closeBracket,
		isTypedBindingPattern, allowAssignment, context)
}

func (this *BallerinaParser) getKeyExpr(member internal.STNode) internal.STNode {
	if member == nil {
		keyIdentifier := this.SyntaxErrors.createMissingTokenWithDiagnostics(SyntaxKind.IDENTIFIER_TOKEN,
			DiagnosticErrorCode.ERROR_MISSING_KEY_EXPR_IN_MEMBER_ACCESS_EXPR)
		missingVarRef := this.STNodeFactory.createSimpleNameReferenceNode(keyIdentifier)
		return this.STNodeFactory.createNodeList(missingVarRef)
	}
	return this.STNodeFactory.createNodeList(member)
}

func (this *BallerinaParser) createTypedBindingPattern(typeDescOrExpr internal.STNode, openBracket internal.STNode, member internal.STNode, closeBracket internal.STNode) internal.STNode {
	bindingPatterns := this.STNodeFactory.createEmptyNodeList()
	if !this.isEmpty(member) {
		memberKind := member.kind
		if (memberKind == SyntaxKind.NUMERIC_LITERAL) || (memberKind == SyntaxKind.ASTERISK_LITERAL) {
			typeDesc := this.getTypeDescFromExpr(typeDescOrExpr)
			arrayTypeDesc := this.getArrayTypeDesc(openBracket, member, closeBracket, typeDesc)
			identifierToken := this.SyntaxErrors.createMissingTokenWithDiagnostics(SyntaxKind.IDENTIFIER_TOKEN,
				DiagnosticErrorCode.ERROR_MISSING_VARIABLE_NAME)
			variableName := this.STNodeFactory.createCaptureBindingPatternNode(identifierToken)
			return this.STNodeFactory.createTypedBindingPatternNode(arrayTypeDesc, variableName)
		}
		bindingPattern := this.getBindingPattern(member, true)
		bindingPatterns = this.STNodeFactory.createNodeList(bindingPattern)
	}
	bindingPattern := this.STNodeFactory.createListBindingPatternNode(openBracket, bindingPatterns, closeBracket)
	typeDesc := this.getTypeDescFromExpr(typeDescOrExpr)
	return this.STNodeFactory.createTypedBindingPatternNode(typeDesc, bindingPattern)
}

func (this *BallerinaParser) parseComplexTypeDescInTypedBPOrExprRhs(typeDescOrExpr internal.STNode, openBracket internal.STNode, member internal.STNode, closeBracket internal.STNode, isTypedBindingPattern bool) internal.STNode {
	pipeOrAndToken := this.parseUnionOrIntersectionToken()
	typedBindingPatternOrExpr := this.parseTypedBindingPatternOrExpr(false)
	if typedBindingPatternOrExpr.kind == SyntaxKind.TYPED_BINDING_PATTERN {
		lhsTypeDesc := this.getTypeDescFromExpr(typeDescOrExpr)
		lhsTypeDesc = this.getArrayTypeDesc(openBracket, member, closeBracket, lhsTypeDesc)
		rhsTypedBindingPattern := internal.STTypedBindingPatternNode(typedBindingPatternOrExpr)
		rhsTypeDesc := rhsTypedBindingPattern.typeDescriptor
		newTypeDesc := this.mergeTypes(lhsTypeDesc, pipeOrAndToken, rhsTypeDesc)
		return this.STNodeFactory.createTypedBindingPatternNode(newTypeDesc, rhsTypedBindingPattern.bindingPattern)
	}
	if isTypedBindingPattern {
		lhsTypeDesc := this.getTypeDescFromExpr(typeDescOrExpr)
		lhsTypeDesc = this.getArrayTypeDesc(openBracket, member, closeBracket, lhsTypeDesc)
		return this.createCaptureBPWithMissingVarName(lhsTypeDesc, pipeOrAndToken, typedBindingPatternOrExpr)
	}
	keyExpr := this.getExpression(member)
	containerExpr := this.getExpression(typeDescOrExpr)
	lhsExpr := this.STNodeFactory.createIndexedExpressionNode(containerExpr, openBracket, keyExpr, closeBracket)
	return this.STNodeFactory.createBinaryExpressionNode(SyntaxKind.BINARY_EXPRESSION, lhsExpr, pipeOrAndToken,
		typedBindingPatternOrExpr)
}

func (this *BallerinaParser) mergeTypes(lhsTypeDesc internal.STNode, pipeOrAndToken internal.STNode, rhsTypeDesc internal.STNode) internal.STNode {
	if pipeOrAndToken.kind == SyntaxKind.PIPE_TOKEN {
		return this.mergeTypesWithUnion(lhsTypeDesc, pipeOrAndToken, rhsTypeDesc)
	} else {
		return this.mergeTypesWithIntersection(lhsTypeDesc, pipeOrAndToken, rhsTypeDesc)
	}
}

func (this *BallerinaParser) mergeTypesWithUnion(lhsTypeDesc internal.STNode, pipeToken internal.STNode, rhsTypeDesc internal.STNode) internal.STNode {
	if rhsTypeDesc.kind == SyntaxKind.UNION_TYPE_DESC {
		rhsUnionTypeDesc := internal.STUnionTypeDescriptorNode(rhsTypeDesc)
		return this.replaceLeftMostUnionWithAUnion(lhsTypeDesc, pipeToken, rhsUnionTypeDesc)
	} else {
		return this.createUnionTypeDesc(lhsTypeDesc, pipeToken, rhsTypeDesc)
	}
}

// func (this *BallerinaParser) mergeTypesWithIntersection(lhsTypeDesc internal.STNode, bitwiseAndToken internal.STNode, rhsTypeDesc internal.STNode) internal.STNode {
// if (lhsTypeDesc.kind == SyntaxKind.UNION_TYPE_DESC) {
// lhsUnionTypeDesc := internal.STUnionTypeDescriptorNode(lhsTypeDesc)
// if (rhsTypeDesc.kind == SyntaxKind.INTERSECTION_TYPE_DESC) {
// rhsTypeDesc = this.replaceLeftMostIntersectionWithAIntersection(lhsUnionTypeDesc.rightTypeDesc,
//                         bitwiseAndToken, (STIntersectionTypeDescriptorNode) rhsTypeDesc)
// return this.createUnionTypeDesc(lhsUnionTypeDesc.leftTypeDesc, lhsUnionTypeDesc.pipeToken, rhsTypeDesc)
// }else if (rhsTypeDesc.kind == SyntaxKind.UNION_TYPE_DESC) {
// rhsTypeDesc = this.replaceLeftMostUnionWithAIntersection(lhsUnionTypeDesc.rightTypeDesc,
//                         bitwiseAndToken, (STUnionTypeDescriptorNode) rhsTypeDesc)
// return this.replaceLeftMostUnionWithAUnion(lhsUnionTypeDesc.leftTypeDesc,
//                         lhsUnionTypeDesc.pipeToken, (STUnionTypeDescriptorNode) rhsTypeDesc)
// }
// }
// if (rhsTypeDesc.kind == SyntaxKind.UNION_TYPE_DESC) {
// rhsUnionTypeDesc := internal.STUnionTypeDescriptorNode(rhsTypeDesc)
// return this.replaceLeftMostUnionWithAIntersection(lhsTypeDesc, bitwiseAndToken, rhsUnionTypeDesc)
// }else if (rhsTypeDesc.kind == SyntaxKind.INTERSECTION_TYPE_DESC) {
// rhsIntSecTypeDesc := internal.STIntersectionTypeDescriptorNode(rhsTypeDesc)
// return this.replaceLeftMostIntersectionWithAIntersection(lhsTypeDesc, bitwiseAndToken, rhsIntSecTypeDesc)
// }
// }

// func (this *BallerinaParser) replaceLeftMostUnionWithAUnion(typeDesc internal.STNode, pipeToken internal.STNode, unionTypeDesc internal.STUnionTypeDescriptorNode) internal.STNode {
// leftTypeDesc := unionTypeDesc.leftTypeDesc
// if (leftTypeDesc.kind == SyntaxKind.UNION_TYPE_DESC) {
// return this.unionTypeDesc.replace(unionTypeDesc.leftTypeDesc,
//                     replaceLeftMostUnionWithAUnion(typeDesc, pipeToken, (STUnionTypeDescriptorNode) leftTypeDesc))
// }
// leftTypeDesc = this.createUnionTypeDesc(typeDesc, pipeToken, leftTypeDesc)
// return this.unionTypeDesc.replace(unionTypeDesc.leftTypeDesc, leftTypeDesc)
// }

// func (this *BallerinaParser) replaceLeftMostUnionWithAIntersection(typeDesc internal.STNode, bitwiseAndToken internal.STNode, unionTypeDesc internal.STUnionTypeDescriptorNode) internal.STNode {
// leftTypeDesc := unionTypeDesc.leftTypeDesc
// if (leftTypeDesc.kind == SyntaxKind.UNION_TYPE_DESC) {
// return this.unionTypeDesc.replace(unionTypeDesc.leftTypeDesc,
//                     replaceLeftMostUnionWithAIntersection(typeDesc, bitwiseAndToken,
//                             (STUnionTypeDescriptorNode) leftTypeDesc))
// }
// if (leftTypeDesc.kind == SyntaxKind.INTERSECTION_TYPE_DESC) {
// return this.unionTypeDesc.replace(unionTypeDesc.leftTypeDesc,
//                     replaceLeftMostIntersectionWithAIntersection(typeDesc, bitwiseAndToken,
//                             (STIntersectionTypeDescriptorNode) leftTypeDesc))
// }
// leftTypeDesc = this.createIntersectionTypeDesc(typeDesc, bitwiseAndToken, leftTypeDesc)
// return this.unionTypeDesc.replace(unionTypeDesc.leftTypeDesc, leftTypeDesc)
// }

// func (this *BallerinaParser) replaceLeftMostIntersectionWithAIntersection(typeDesc internal.STNode, bitwiseAndToken internal.STNode, intersectionTypeDesc internal.STIntersectionTypeDescriptorNode) internal.STNode {
// leftTypeDesc := intersectionTypeDesc.leftTypeDesc
// if (leftTypeDesc.kind == SyntaxKind.INTERSECTION_TYPE_DESC) {
// return this.intersectionTypeDesc.replace(intersectionTypeDesc.leftTypeDesc,
//                     replaceLeftMostIntersectionWithAIntersection(typeDesc, bitwiseAndToken,
//                             (STIntersectionTypeDescriptorNode) leftTypeDesc))
// }
// leftTypeDesc = this.createIntersectionTypeDesc(typeDesc, bitwiseAndToken, leftTypeDesc)
// return this.intersectionTypeDesc.replace(intersectionTypeDesc.leftTypeDesc, leftTypeDesc)
// }

func (this *BallerinaParser) getArrayTypeDesc(openBracket internal.STNode, member internal.STNode, closeBracket internal.STNode, lhsTypeDesc internal.STNode) internal.STNode {
	if lhsTypeDesc.kind == SyntaxKind.UNION_TYPE_DESC {
		unionTypeDesc := internal.STUnionTypeDescriptorNode(lhsTypeDesc)
		middleTypeDesc := this.getArrayTypeDesc(openBracket, member, closeBracket, unionTypeDesc.rightTypeDesc)
		lhsTypeDesc = this.mergeTypesWithUnion(unionTypeDesc.leftTypeDesc, unionTypeDesc.pipeToken, middleTypeDesc)
	} else if lhsTypeDesc.kind == SyntaxKind.INTERSECTION_TYPE_DESC {
		intersectionTypeDesc := internal.STIntersectionTypeDescriptorNode(lhsTypeDesc)
		middleTypeDesc := this.getArrayTypeDesc(openBracket, member, closeBracket, intersectionTypeDesc.rightTypeDesc)
		lhsTypeDesc = this.mergeTypesWithIntersection(intersectionTypeDesc.leftTypeDesc,
			intersectionTypeDesc.bitwiseAndToken, middleTypeDesc)
	}
	return lhsTypeDesc
}

func (this *BallerinaParser) parseUnionOrIntersectionToken() internal.STNode {
	token := this.peek()
	if (token.kind == SyntaxKind.PIPE_TOKEN) || (token.kind == SyntaxKind.BITWISE_AND_TOKEN) {
		return this.consume()
	} else {
		this.recover(token, ParserRuleContext.UNION_OR_INTERSECTION_TOKEN)
		return this.parseUnionOrIntersectionToken()
	}
}

func (this *BallerinaParser) getBracketedListNodeType(memberNode internal.STNode, isTypedBindingPattern bool) SyntaxKind {
	if this.isEmpty(memberNode) {
		return SyntaxKind.NONE
	}
	if this.isDefiniteTypeDesc(memberNode.kind) {
		return SyntaxKind.TUPLE_TYPE_DESC
	}
	switch memberNode.kind {
	case ASTERISK_LITERAL:
		return SyntaxKind.ARRAY_TYPE_DESC
	case CAPTURE_BINDING_PATTERN:
	case LIST_BINDING_PATTERN:
	case REST_BINDING_PATTERN:
	case MAPPING_BINDING_PATTERN:
	case WILDCARD_BINDING_PATTERN:
		return SyntaxKind.LIST_BINDING_PATTERN
	case QUALIFIED_NAME_REFERENCE:
	case REST_TYPE:
		return SyntaxKind.TUPLE_TYPE_DESC
	case NUMERIC_LITERAL:
		if isTypedBindingPattern {
			return SyntaxKind.ARRAY_TYPE_DESC
		}
		return SyntaxKind.ARRAY_TYPE_DESC_OR_MEMBER_ACCESS
	case SIMPLE_NAME_REFERENCE:
	case BRACKETED_LIST:
	case MAPPING_BP_OR_MAPPING_CONSTRUCTOR:
		return SyntaxKind.NONE
	case ERROR_CONSTRUCTOR:
		if isTypedBindingPattern {
			return SyntaxKind.LIST_BINDING_PATTERN
		}
		errorCtorNode, ok := memberNode.(*STErrorConstructorExpressionNode)
		if !ok {
			panic("getBracketedListNodeType: expected STErrorConstructorExpressionNode")
		}
		if this.isPossibleErrorBindingPattern(errorCtorNode) {
			return SyntaxKind.NONE
		}
		return SyntaxKind.INDEXED_EXPRESSION
	default:
		if isTypedBindingPattern {
			return SyntaxKind.NONE
		}
		return SyntaxKind.INDEXED_EXPRESSION
	}
}

func (this *BallerinaParser) parseStatementStartsWithOpenBracket(annots internal.STNode, possibleMappingField bool) internal.STNode {
	this.startContext(ParserRuleContext.ASSIGNMENT_OR_VAR_DECL_STMT)
	return this.parseStatementStartsWithOpenBracket(annots, true, possibleMappingField)
}

func (this *BallerinaParser) parseMemberBracketedList() internal.STNode {
	annots := this.STNodeFactory.createEmptyNodeList()
	return this.parseStatementStartsWithOpenBracket(annots, false, false)
}

func (this *BallerinaParser) parseStatementStartsWithOpenBracket(annots internal.STNode, isRoot bool, possibleMappingField bool) internal.STNode {
	this.startContext(ParserRuleContext.STMT_START_BRACKETED_LIST)
	openBracket := this.parseOpenBracket()
	memberList := make([]interface{}, 0)
	for !this.isBracketedListEnd(peek().kind) {
		member := this.parseStatementStartBracketedListMember()
		currentNodeType := this.getStmtStartBracketedListType(member)
		switch currentNodeType {
		case TUPLE_TYPE_DESC:
			member = this.parseComplexTypeDescriptor(member, ParserRuleContext.TYPE_DESC_IN_TUPLE, false)
			member = this.createMemberOrRestNode(STNodeFactory.createEmptyNodeList(), member)
			return this.parseAsTupleTypeDesc(annots, openBracket, memberList, member, isRoot)
		case MEMBER_TYPE_DESC:
		case REST_TYPE:
			return this.parseAsTupleTypeDesc(annots, openBracket, memberList, member, isRoot)
		case LIST_BINDING_PATTERN:
			return this.parseAsListBindingPattern(openBracket, memberList, member, isRoot)
		case LIST_CONSTRUCTOR:
			return this.parseAsListConstructor(openBracket, memberList, member, isRoot)
		case LIST_BP_OR_LIST_CONSTRUCTOR:
			return this.parseAsListBindingPatternOrListConstructor(openBracket, memberList, member, isRoot)
		case TUPLE_TYPE_DESC_OR_LIST_CONST:
			return this.parseAsTupleTypeDescOrListConstructor(annots, openBracket, memberList, member, isRoot)
		case NONE:
		default:
			this.memberList.add(member)
			break
		}
		memberEnd := this.parseBracketedListMemberEnd()
		if memberEnd == nil {
			break
		}
		this.memberList.add(memberEnd)
	}
	closeBracket := this.parseCloseBracket()
	bracketedList := this.parseStatementStartBracketedListRhs(annots, openBracket, memberList, closeBracket,
		isRoot, possibleMappingField)
	return bracketedList
}

func (this *BallerinaParser) parseStatementStartBracketedListMember() internal.STNode {
	typeDescQualifiers := make([]interface{}, 0)
	return this.parseStatementStartBracketedListMember(typeDescQualifiers)
}

func (this *BallerinaParser) parseStatementStartBracketedListMember(qualifiers []STNode) internal.STNode {
	this.parseTypeDescQualifiers(qualifiers)
	nextToken := this.peek()
	switch nextToken.kind {
	case OPEN_BRACKET_TOKEN:
		this.reportInvalidQualifierList(qualifiers)
		return this.parseMemberBracketedList()
	case IDENTIFIER_TOKEN:
		this.reportInvalidQualifierList(qualifiers)
		identifier := this.parseQualifiedIdentifier(ParserRuleContext.VARIABLE_REF)
		if this.isWildcardBP(identifier) {
			simpleNameNode, ok := identifier.(*STSimpleNameReferenceNode)
			if !ok {
				panic("parseStatementStartBracketedListMember: expected STSimpleNameReferenceNode")
			}
			varName := simpleNameNode.name
			return this.getWildcardBindingPattern(varName)
		}
		nextToken = this.peek()
		if nextToken.kind == SyntaxKind.ELLIPSIS_TOKEN {
			ellipsis := this.parseEllipsis()
			return this.STNodeFactory.createRestDescriptorNode(identifier, ellipsis)
		}
		if (nextToken.kind != SyntaxKind.OPEN_BRACKET_TOKEN) && this.isValidTypeContinuationToken(nextToken) {
			return this.parseComplexTypeDescriptor(identifier, ParserRuleContext.TYPE_DESC_IN_TUPLE, false)
		}
		return this.parseExpressionRhs(DEFAULT_OP_PRECEDENCE, identifier, false, true)
	case OPEN_BRACE_TOKEN:
		this.reportInvalidQualifierList(qualifiers)
		return this.parseMappingBindingPatterOrMappingConstructor()
	case ERROR_KEYWORD:
		this.reportInvalidQualifierList(qualifiers)
		nextNextToken := this.getNextNextToken()
		if (nextNextToken.kind == SyntaxKind.OPEN_PAREN_TOKEN) || (nextNextToken.kind == SyntaxKind.IDENTIFIER_TOKEN) {
			return this.parseErrorBindingPatternOrErrorConstructor()
		}
		return this.parseTypeDescriptor(ParserRuleContext.TYPE_DESC_IN_TUPLE)
	case ELLIPSIS_TOKEN:
		this.reportInvalidQualifierList(qualifiers)
		return this.parseRestBindingOrSpreadMember()
	case XML_KEYWORD:
	case STRING_KEYWORD:
		this.reportInvalidQualifierList(qualifiers)
		if getNextNextToken().kind == SyntaxKind.BACKTICK_TOKEN {
			return this.parseExpression(false)
		}
		return this.parseTypeDescriptor(ParserRuleContext.TYPE_DESC_IN_TUPLE)
	case TABLE_KEYWORD:
	case STREAM_KEYWORD:
		this.reportInvalidQualifierList(qualifiers)
		if getNextNextToken().kind == SyntaxKind.LT_TOKEN {
			return this.parseTypeDescriptor(ParserRuleContext.TYPE_DESC_IN_TUPLE)
		}
		return this.parseExpression(false)
	case OPEN_PAREN_TOKEN:
		return this.parseTypeDescOrExpr(qualifiers)
	case FUNCTION_KEYWORD:
		return this.parseAnonFuncExprOrFuncTypeDesc(qualifiers)
	case AT_TOKEN:
		return this.parseTupleMember()
	default:
		if this.isValidExpressionStart(nextToken.kind, 1) {
			this.reportInvalidQualifierList(qualifiers)
			return this.parseExpression(false)
		}
		if this.isTypeStartingToken(nextToken.kind) {
			return this.parseTypeDescriptor(qualifiers, ParserRuleContext.TYPE_DESC_IN_TUPLE)
		}
		this.recover(peek(), ParserRuleContext.STMT_START_BRACKETED_LIST_MEMBER)
		return this.parseStatementStartBracketedListMember(qualifiers)
	}
}

func (this *BallerinaParser) parseRestBindingOrSpreadMember() internal.STNode {
	ellipsis := this.parseEllipsis()
	expr := this.parseExpression()
	if expr.kind == SyntaxKind.SIMPLE_NAME_REFERENCE {
		return this.STNodeFactory.createRestBindingPatternNode(ellipsis, expr)
	} else {
		return this.STNodeFactory.createSpreadMemberNode(ellipsis, expr)
	}
}

func (this *BallerinaParser) parseAsTupleTypeDescOrListConstructor(annots internal.STNode, openBracket internal.STNode, memberList []STNode, member internal.STNode, isRoot bool) internal.STNode {
	this.memberList.add(member)
	memberEnd := this.parseBracketedListMemberEnd()
	var tupleTypeDescOrListCons internal.STNode
	if memberEnd == nil {
		closeBracket := this.parseCloseBracket()
		tupleTypeDescOrListCons = this.parseTupleTypeDescOrListConstructorRhs(openBracket, memberList, closeBracket, isRoot)
	} else {
		this.memberList.add(memberEnd)
		tupleTypeDescOrListCons = this.parseTupleTypeDescOrListConstructor(annots, openBracket, memberList, isRoot)
	}
	return tupleTypeDescOrListCons
}

func (this *BallerinaParser) parseTupleTypeDescOrListConstructor(annots internal.STNode) internal.STNode {
	this.startContext(ParserRuleContext.BRACKETED_LIST)
	openBracket := this.parseOpenBracket()
	memberList := make([]interface{}, 0)
	return this.parseTupleTypeDescOrListConstructor(annots, openBracket, memberList, false)
}

func (this *BallerinaParser) parseTupleTypeDescOrListConstructor(annots internal.STNode, openBracket internal.STNode, memberList []STNode, isRoot bool) internal.STNode {
	nextToken := this.peek()
	for !this.isBracketedListEnd(nextToken.kind) {
		member := this.parseTupleTypeDescOrListConstructorMember(annots)
		currentNodeType := this.getParsingNodeTypeOfTupleTypeOrListCons(member)
		switch currentNodeType {
		case LIST_CONSTRUCTOR:
			return this.parseAsListConstructor(openBracket, memberList, member, isRoot)
		case REST_TYPE:
		case MEMBER_TYPE_DESC:
			return this.parseAsTupleTypeDesc(annots, openBracket, memberList, member, isRoot)
		case TUPLE_TYPE_DESC:
			member = this.parseComplexTypeDescriptor(member, ParserRuleContext.TYPE_DESC_IN_TUPLE, false)
			member = this.createMemberOrRestNode(STNodeFactory.createEmptyNodeList(), member)
			return this.parseAsTupleTypeDesc(annots, openBracket, memberList, member, isRoot)
		case TUPLE_TYPE_DESC_OR_LIST_CONST:
		default:
			this.memberList.add(member)
			break
		}
		memberEnd := this.parseBracketedListMemberEnd()
		if memberEnd == nil {
			break
		}
		this.memberList.add(memberEnd)
		nextToken = this.peek()
	}
	closeBracket := this.parseCloseBracket()
	return this.parseTupleTypeDescOrListConstructorRhs(openBracket, memberList, closeBracket, isRoot)
}

func (this *BallerinaParser) parseTupleTypeDescOrListConstructorMember(annots internal.STNode) internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case OPEN_BRACKET_TOKEN:
		return this.parseTupleTypeDescOrListConstructor(annots)
	case IDENTIFIER_TOKEN:
		identifier := this.parseQualifiedIdentifier(ParserRuleContext.VARIABLE_REF)
		if peek().kind == SyntaxKind.ELLIPSIS_TOKEN {
			ellipsis := this.parseEllipsis()
			return this.STNodeFactory.createRestDescriptorNode(identifier, ellipsis)
		}
		return this.parseExpressionRhs(DEFAULT_OP_PRECEDENCE, identifier, false, false)
	case OPEN_BRACE_TOKEN:
		return this.parseMappingConstructorExpr()
	case ERROR_KEYWORD:
		nextNextToken := this.getNextNextToken()
		if (nextNextToken.kind == SyntaxKind.OPEN_PAREN_TOKEN) || (nextNextToken.kind == SyntaxKind.IDENTIFIER_TOKEN) {
			return this.parseErrorConstructorExpr(false)
		}
		return this.parseTypeDescriptor(ParserRuleContext.TYPE_DESC_IN_TUPLE)
	case XML_KEYWORD:
	case STRING_KEYWORD:
		if getNextNextToken().kind == SyntaxKind.BACKTICK_TOKEN {
			return this.parseExpression(false)
		}
		return this.parseTypeDescriptor(ParserRuleContext.TYPE_DESC_IN_TUPLE)
	case TABLE_KEYWORD:
	case STREAM_KEYWORD:
		if getNextNextToken().kind == SyntaxKind.LT_TOKEN {
			return this.parseTypeDescriptor(ParserRuleContext.TYPE_DESC_IN_TUPLE)
		}
		return this.parseExpression(false)
	case OPEN_PAREN_TOKEN:
		return this.parseTypeDescOrExpr()
	case AT_TOKEN:
		return this.parseTupleMember()
	default:
		if this.isValidExpressionStart(nextToken.kind, 1) {
			return this.parseExpression(false)
		}
		if this.isTypeStartingToken(nextToken.kind) {
			return this.parseTypeDescriptor(ParserRuleContext.TYPE_DESC_IN_TUPLE)
		}
		this.recover(peek(), ParserRuleContext.TUPLE_TYPE_DESC_OR_LIST_CONST_MEMBER)
		return this.parseTupleTypeDescOrListConstructorMember(annots)
	}
}

func (this *BallerinaParser) getParsingNodeTypeOfTupleTypeOrListCons(memberNode internal.STNode) SyntaxKind {
	return this.getStmtStartBracketedListType(memberNode)
}

func (this *BallerinaParser) parseTupleTypeDescOrListConstructorRhs(openBracket internal.STNode, members []STNode, closeBracket internal.STNode, isRoot bool) internal.STNode {
	var tupleTypeOrListConst internal.STNode
	switch peek().kind {
	case COMMA_TOKEN:
	case CLOSE_BRACE_TOKEN:
	case CLOSE_BRACKET_TOKEN:
	case PIPE_TOKEN:
	case BITWISE_AND_TOKEN:
		if !isRoot {
			this.endContext()
			return nil
		}
	default:
		if this.isValidExprRhsStart(peek().kind, closeBracket.kind) || (isRoot && (peek().kind == SyntaxKind.EQUAL_TOKEN)) {
			members = this.getExpressionList(members, false)
			memberExpressions := this.STNodeFactory.createNodeList(members)
			tupleTypeOrListConst = this.STNodeFactory.createListConstructorExpressionNode(openBracket,
				memberExpressions, closeBracket)
			break
		}
		memberTypeDescs := this.STNodeFactory.createNodeList(getTupleMemberList(members))
		tupleTypeDesc := this.STNodeFactory.createTupleTypeDescriptorNode(openBracket, memberTypeDescs, closeBracket)
		tupleTypeOrListConst = this.parseComplexTypeDescriptor(tupleTypeDesc, ParserRuleContext.TYPE_DESC_IN_TUPLE, false)
	}
	this.endContext()
	if !isRoot {
		return tupleTypeOrListConst
	}
	annots := this.STNodeFactory.createEmptyNodeList()
	return this.parseStmtStartsWithTupleTypeOrExprRhs(annots, tupleTypeOrListConst, true)
}

func (this *BallerinaParser) parseStmtStartsWithTupleTypeOrExprRhs(annots internal.STNode, tupleTypeOrListConst internal.STNode, isRoot bool) internal.STNode {
	if (this.tupleTypeOrListConst.kind.compareTo(SyntaxKind.RECORD_TYPE_DESC) >= 0) && (this.tupleTypeOrListConst.kind.compareTo(SyntaxKind.TYPEDESC_TYPE_DESC) <= 0) {
		varDeclQualifiers := make([]interface{}, 0)
		typedBindingPattern := this.parseTypedBindingPatternTypeRhs(tupleTypeOrListConst, ParserRuleContext.VAR_DECL_STMT, isRoot)
		if !isRoot {
			return typedBindingPattern
		}
		this.switchContext(ParserRuleContext.VAR_DECL_STMT)
		return this.parseVarDeclRhs(annots, varDeclQualifiers, typedBindingPattern, false)
	}
	expr := this.getExpression(tupleTypeOrListConst)
	expr = this.parseExpressionRhs(DEFAULT_OP_PRECEDENCE, expr, false, true)
	return this.parseStatementStartWithExprRhs(expr)
}

func (this *BallerinaParser) parseAsTupleTypeDesc(annots internal.STNode, openBracket internal.STNode, memberList []STNode, member internal.STNode, isRoot bool) internal.STNode {
	memberList = this.getTupleMemberList(memberList)
	this.startContext(ParserRuleContext.TUPLE_MEMBERS)
	tupleTypeMembers := this.parseTupleTypeMembers(member, memberList)
	closeBracket := this.parseCloseBracket()
	this.endContext()
	tupleType := this.STNodeFactory.createTupleTypeDescriptorNode(openBracket, tupleTypeMembers, closeBracket)
	typeDesc := this.parseComplexTypeDescriptor(tupleType, ParserRuleContext.TYPE_DESC_IN_TYPE_BINDING_PATTERN, true)
	this.endContext()
	if !isRoot {
		return typeDesc
	}
	typedBindingPattern := this.parseTypedBindingPatternTypeRhs(typeDesc, ParserRuleContext.VAR_DECL_STMT, true)
	this.switchContext(ParserRuleContext.VAR_DECL_STMT)
	return this.parseVarDeclRhs(annots, nil, typedBindingPattern, false)
}

func (this *BallerinaParser) parseAsListBindingPattern(openBracket internal.STNode, memberList []STNode, member internal.STNode, isRoot bool) internal.STNode {
	memberList = this.getBindingPatternsList(memberList, true)
	this.memberList.add(getBindingPattern(member, true))
	this.switchContext(ParserRuleContext.LIST_BINDING_PATTERN)
	listBindingPattern := this.parseListBindingPattern(openBracket, member, memberList)
	this.endContext()
	if !isRoot {
		return listBindingPattern
	}
	return this.parseAssignmentStmtRhs(listBindingPattern)
}

func (this *BallerinaParser) parseAsListBindingPattern(openBracket internal.STNode, memberList []STNode) internal.STNode {
	memberList = this.getBindingPatternsList(memberList, true)
	this.switchContext(ParserRuleContext.LIST_BINDING_PATTERN)
	listBindingPattern := this.parseListBindingPattern(openBracket, memberList)
	this.endContext()
	return listBindingPattern
}

func (this *BallerinaParser) parseAsListBindingPatternOrListConstructor(openBracket internal.STNode, memberList []STNode, member internal.STNode, isRoot bool) internal.STNode {
	this.memberList.add(member)
	memberEnd := this.parseBracketedListMemberEnd()
	var listBindingPatternOrListCons internal.STNode
	if memberEnd == nil {
		closeBracket := this.parseCloseBracket()
		listBindingPatternOrListCons = this.parseListBindingPatternOrListConstructor(openBracket, memberList, closeBracket, isRoot)
	} else {
		this.memberList.add(memberEnd)
		listBindingPatternOrListCons = this.parseListBindingPatternOrListConstructor(openBracket, memberList, isRoot)
	}
	return listBindingPatternOrListCons
}

func (this *BallerinaParser) getStmtStartBracketedListType(memberNode internal.STNode) SyntaxKind {
	if (this.memberNode.kind.compareTo(SyntaxKind.RECORD_TYPE_DESC) >= 0) && (this.memberNode.kind.compareTo(SyntaxKind.FUTURE_TYPE_DESC) <= 0) {
		return SyntaxKind.TUPLE_TYPE_DESC
	}
	switch memberNode.kind {
	case WILDCARD_BINDING_PATTERN,
		CAPTURE_BINDING_PATTERN,
		LIST_BINDING_PATTERN,
		MAPPING_BINDING_PATTERN,
		ERROR_BINDING_PATTERN:
		SyntaxKind.LIST_BINDING_PATTERN
	case QUALIFIED_NAME_REFERENCE:
		SyntaxKind.TUPLE_TYPE_DESC
	case LIST_CONSTRUCTOR,
		MAPPING_CONSTRUCTOR,
		SPREAD_MEMBER:
		SyntaxKind.LIST_CONSTRUCTOR
	case MAPPING_BP_OR_MAPPING_CONSTRUCTOR,
		REST_BINDING_PATTERN:
		SyntaxKind.LIST_BP_OR_LIST_CONSTRUCTOR
	case SIMPLE_NAME_REFERENCE, // member is a simple type-ref/var-ref
		BRACKETED_LIST:
		SyntaxKind.NONE
	case ERROR_CONSTRUCTOR:
		errorCtorNode, ok := memberNode.(*STErrorConstructorExpressionNode)
		if !ok {
			panic("getStmtStartBracketedListType: expected STErrorConstructorExpressionNode")
		}
		if this.isPossibleErrorBindingPattern(errorCtorNode) {
			SyntaxKind.NONE
		}
		SyntaxKind.LIST_CONSTRUCTOR
	case INDEXED_EXPRESSION:
		SyntaxKind.TUPLE_TYPE_DESC_OR_LIST_CONST
	case MEMBER_TYPE_DESC:
		SyntaxKind.MEMBER_TYPE_DESC
	case REST_TYPE:
		SyntaxKind.REST_TYPE
	default:
		if (this.isExpression(memberNode.kind) && (!this.isAllBasicLiterals(memberNode))) && (!this.isAmbiguous(memberNode)) {
			SyntaxKind.LIST_CONSTRUCTOR
		}
		SyntaxKind.NONE
	}
}

func (this *BallerinaParser) isPossibleErrorBindingPattern(errorConstructor internal.STErrorConstructorExpressionNode) bool {
	args := errorConstructor.arguments
	size := this.args.bucketCount()
	i := 0
	for ; i < size; i++ {
		arg := this.args.childInBucket(i)
		if ((arg.kind != SyntaxKind.NAMED_ARG) && (arg.kind != SyntaxKind.POSITIONAL_ARG)) && (arg.kind != SyntaxKind.REST_ARG) {
			continue
		}
		functionArg, ok := arg.(STFunctionArgumentNode)
		if !ok {
			panic("isPossibleErrorBindingPattern: expected STFunctionArgumentNode")
		}
		if !this.isPosibleArgBindingPattern(functionArg) {
			return false
		}
	}
	return true
}

func (this *BallerinaParser) isPosibleArgBindingPattern(arg internal.STFunctionArgumentNode) bool {
	switch arg.kind {
	case POSITIONAL_ARG:
		positionalArg, ok := arg.(*STPositionalArgumentNode)
		if !ok {
			panic("isPosibleArgBindingPattern: expected STPositionalArgumentNode")
		}
		this.isPosibleBindingPattern(positionalArg.expression)
	case NAMED_ARG:
		namedArg, ok := arg.(*STNamedArgumentNode)
		if !ok {
			panic("isPosibleArgBindingPattern: expected STNamedArgumentNode")
		}
		this.isPosibleBindingPattern(namedArg.expression)
	case REST_ARG:
		restArg, ok := arg.(*STRestArgumentNode)
		if !ok {
			panic("isPosibleArgBindingPattern: expected STRestArgumentNode")
		}
		(restArg.expression.kind == SyntaxKind.SIMPLE_NAME_REFERENCE)
	default:
		false
	}
}

func (this *BallerinaParser) isPosibleBindingPattern(node internal.STNode) bool {
	switch node.kind {
	case SIMPLE_NAME_REFERENCE:
		return true
	case LIST_CONSTRUCTOR:
		listConstructor := internal.STListConstructorExpressionNode(node)
		i := 0
		for ; i < this.listConstructor.bucketCount(); i++ {
			expr := this.listConstructor.childInBucket(i)
			if !this.isPosibleBindingPattern(expr) {
				return false
			}
		}
		return true
	case MAPPING_CONSTRUCTOR:
		mappingConstructor := internal.STMappingConstructorExpressionNode(node)
		i := 0
		for ; i < this.mappingConstructor.bucketCount(); i++ {
			expr := this.mappingConstructor.childInBucket(i)
			if !this.isPosibleBindingPattern(expr) {
				return false
			}
		}
		return true
	case SPECIFIC_FIELD:
		specificField := internal.STSpecificFieldNode(node)
		if specificField.readonlyKeyword != nil {
			return false
		}
		if specificField.valueExpr == nil {
			return true
		}
		return this.isPosibleBindingPattern(specificField.valueExpr)
	case ERROR_CONSTRUCTOR:
		errorCtorNode, ok := node.(*STErrorConstructorExpressionNode)
		if !ok {
			panic("isPosibleBindingPattern: expected STErrorConstructorExpressionNode")
		}
		return this.isPossibleErrorBindingPattern(errorCtorNode)
	default:
		return false
	}
}

func (this *BallerinaParser) parseStatementStartBracketedListRhs(annots internal.STNode, openBracket internal.STNode, members []STNode, closeBracket internal.STNode, isRoot bool, possibleMappingField bool) internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case EQUAL_TOKEN:
		if !isRoot {
			this.endContext()
			return nil
		}
		memberBindingPatterns := this.STNodeFactory.createNodeList(getBindingPatternsList(members, true))
		listBindingPattern := this.STNodeFactory.createListBindingPatternNode(openBracket,
			memberBindingPatterns, closeBracket)
		this.endContext()
		this.switchContext(ParserRuleContext.ASSIGNMENT_STMT)
		return this.parseAssignmentStmtRhs(listBindingPattern)
	case IDENTIFIER_TOKEN:
	case OPEN_BRACE_TOKEN:
		if !isRoot {
			this.endContext()
			return nil
		}
		if this.members.isEmpty() {
			openBracket = this.SyntaxErrors.addDiagnostic(openBracket, DiagnosticErrorCode.ERROR_MISSING_TUPLE_MEMBER)
		}
		this.switchContext(ParserRuleContext.TYPE_DESC_IN_TYPE_BINDING_PATTERN)
		this.startContext(ParserRuleContext.TUPLE_MEMBERS)
		memberTypeDescs := this.STNodeFactory.createNodeList(getTupleMemberList(members))
		tupleTypeDesc := this.STNodeFactory.createTupleTypeDescriptorNode(openBracket, memberTypeDescs, closeBracket)
		this.endContext()
		typeDesc := this.parseComplexTypeDescriptor(tupleTypeDesc,
			ParserRuleContext.TYPE_DESC_IN_TYPE_BINDING_PATTERN, true)
		typedBindingPattern := this.parseTypedBindingPatternTypeRhs(typeDesc, ParserRuleContext.VAR_DECL_STMT)
		this.endContext()
		return this.parseStmtStartsWithTypedBPOrExprRhs(annots, typedBindingPattern)
	case OPEN_BRACKET_TOKEN:
		if !isRoot {
			memberTypeDescs = this.STNodeFactory.createNodeList(getTupleMemberList(members))
			tupleTypeDesc = this.STNodeFactory.createTupleTypeDescriptorNode(openBracket, memberTypeDescs, closeBracket)
			this.endContext()
			typeDesc = this.parseComplexTypeDescriptor(tupleTypeDesc, ParserRuleContext.TYPE_DESC_IN_TUPLE, false)
			return typeDesc
		}
		list := nil
		this.endContext()
		tpbOrExpr := this.parseTypedBindingPatternOrExprRhs(list, true)
		return this.parseStmtStartsWithTypedBPOrExprRhs(annots, tpbOrExpr)
	case COLON_TOKEN:
		if possibleMappingField && (len(members) == 1) {
			this.startContext(ParserRuleContext.MAPPING_CONSTRUCTOR)
			colon := this.parseColon()
			fieldNameExpr := this.getExpression(members.get(0))
			valueExpr := this.parseExpression()
			return this.STNodeFactory.createComputedNameFieldNode(openBracket, fieldNameExpr, closeBracket, colon,
				valueExpr)
		}
	default:
		this.endContext()
		if !isRoot {
			return nil
		}
		list = nil
		exprOrTPB := this.parseTypedBindingPatternOrExprRhs(list, false)
		return this.parseStmtStartsWithTypedBPOrExprRhs(annots, exprOrTPB)
	}
}

func (this *BallerinaParser) isWildcardBP(node internal.STNode) bool {
	switch node.kind {
	case SIMPLE_NAME_REFERENCE:
		simpleNameNode, ok := node.(*STSimpleNameReferenceNode)
		if !ok {
			panic("isWildcardBP: expected STSimpleNameReferenceNode")
		}
		nameToken, ok := simpleNameNode.name.(STToken)
		if !ok {
			panic("isWildcardBP: expected STToken")
		}
		this.isUnderscoreToken(nameToken)
	case IDENTIFIER_TOKEN:
		identifierToken, ok := node.(STToken)
		if !ok {
			panic("isWildcardBP: expected STToken")
		}
		this.isUnderscoreToken(identifierToken)
	default:
		false
	}
}

func (this *BallerinaParser) isUnderscoreToken(token internal.STToken) bool {
	return "_".equals(token.text())
}

func (this *BallerinaParser) getWildcardBindingPattern(identifier internal.STNode) internal.STNode {
	var underscore internal.STNode
	switch identifier.kind {
	case SIMPLE_NAME_REFERENCE:
		simpleNameNode, ok := identifier.(*STSimpleNameReferenceNode)
		if !ok {
			panic("getWildcardBindingPattern: expected STSimpleNameReferenceNode")
		}
		varName := simpleNameNode.name
		nameToken, ok := varName.(STToken)
		if !ok {
			panic("getWildcardBindingPattern: expected STToken")
		}
		underscore = this.getUnderscoreKeyword(nameToken)
		return this.STNodeFactory.createWildcardBindingPatternNode(underscore)
	case IDENTIFIER_TOKEN:
		identifierToken, ok := identifier.(STToken)
		if !ok {
			panic("getWildcardBindingPattern: expected STToken")
		}
		underscore = this.getUnderscoreKeyword(identifierToken)
		return this.STNodeFactory.createWildcardBindingPatternNode(underscore)
	default:
		panic("getWildcardBindingPattern: expected SIMPLE_NAME_REFERENCE or IDENTIFIER_TOKEN")
	}
}

func (this *BallerinaParser) parseStatementStartsWithOpenBrace() internal.STNode {
	this.startContext(ParserRuleContext.AMBIGUOUS_STMT)
	openBrace := this.parseOpenBrace()
	if peek().kind == SyntaxKind.CLOSE_BRACE_TOKEN {
		closeBrace := this.parseCloseBrace()
		switch peek().kind {
		case EQUAL_TOKEN:
			this.switchContext(ParserRuleContext.ASSIGNMENT_STMT)
			fields := this.STNodeFactory.createEmptyNodeList()
			bindingPattern := this.STNodeFactory.createMappingBindingPatternNode(openBrace, fields,
				closeBrace)
			return this.parseAssignmentStmtRhs(bindingPattern)
		case RIGHT_ARROW_TOKEN:
		case SYNC_SEND_TOKEN:
			this.switchContext(ParserRuleContext.EXPRESSION_STATEMENT)
			fields = this.STNodeFactory.createEmptyNodeList()
			expr := this.STNodeFactory.createMappingConstructorExpressionNode(openBrace, fields, closeBrace)
			expr = this.parseExpressionRhs(DEFAULT_OP_PRECEDENCE, expr, false, true)
			return this.parseStatementStartWithExprRhs(expr)
		default:
			statements := this.STNodeFactory.createEmptyNodeList()
			this.endContext()
			return this.STNodeFactory.createBlockStatementNode(openBrace, statements, closeBrace)
		}
	}
	member := this.parseStatementStartingBracedListFirstMember(openBrace.isMissing())
	nodeType := this.getBracedListType(member)
	var stmt internal.STNode
	switch nodeType {
	case MAPPING_BINDING_PATTERN:
		return this.parseStmtAsMappingBindingPatternStart(openBrace, member)
	case MAPPING_CONSTRUCTOR:
		return this.parseStmtAsMappingConstructorStart(openBrace, member)
	case MAPPING_BP_OR_MAPPING_CONSTRUCTOR:
		return this.parseStmtAsMappingBPOrMappingConsStart(openBrace, member)
	case BLOCK_STATEMENT:
		closeBrace := this.parseCloseBrace()
		stmt = this.STNodeFactory.createBlockStatementNode(openBrace, member, closeBrace)
		this.endContext()
		return stmt
	default:
		stmts := make([]interface{}, 0)
		this.stmts.add(member)
		statements := this.parseStatements(stmts)
		closeBrace = this.parseCloseBrace()
		this.endContext()
		return this.STNodeFactory.createBlockStatementNode(openBrace, statements, closeBrace)
	}
}

func (this *BallerinaParser) parseStmtAsMappingBindingPatternStart(openBrace internal.STNode, firstMappingField internal.STNode) internal.STNode {
	this.switchContext(ParserRuleContext.ASSIGNMENT_STMT)
	this.startContext(ParserRuleContext.MAPPING_BINDING_PATTERN)
	bindingPatterns := make([]interface{}, 0)
	if firstMappingField.kind != SyntaxKind.REST_BINDING_PATTERN {
		this.bindingPatterns.add(getBindingPattern(firstMappingField, false))
	}
	mappingBP := this.parseMappingBindingPattern(openBrace, bindingPatterns, firstMappingField)
	return this.parseAssignmentStmtRhs(mappingBP)
}

func (this *BallerinaParser) parseStmtAsMappingConstructorStart(openBrace internal.STNode, firstMember internal.STNode) internal.STNode {
	this.switchContext(ParserRuleContext.EXPRESSION_STATEMENT)
	this.startContext(ParserRuleContext.MAPPING_CONSTRUCTOR)
	members := make([]interface{}, 0)
	mappingCons := this.parseAsMappingConstructor(openBrace, members, firstMember)
	expr := this.parseExpressionRhs(DEFAULT_OP_PRECEDENCE, mappingCons, false, true)
	return this.parseStatementStartWithExprRhs(expr)
}

func (this *BallerinaParser) parseAsMappingConstructor(openBrace internal.STNode, members []STNode, member internal.STNode) internal.STNode {
	this.members.add(member)
	members = this.getExpressionList(members, true)
	this.switchContext(ParserRuleContext.MAPPING_CONSTRUCTOR)
	fields := this.parseMappingConstructorFields(members)
	closeBrace := this.parseCloseBrace()
	this.endContext()
	return this.STNodeFactory.createMappingConstructorExpressionNode(openBrace, fields, closeBrace)
}

func (this *BallerinaParser) parseStmtAsMappingBPOrMappingConsStart(openBrace internal.STNode, member internal.STNode) internal.STNode {
	this.startContext(ParserRuleContext.MAPPING_BP_OR_MAPPING_CONSTRUCTOR)
	members := make([]interface{}, 0)
	this.members.add(member)
	var bpOrConstructor internal.STNode
	memberEnd := this.parseMappingFieldEnd()
	if memberEnd == nil {
		closeBrace := this.parseCloseBrace()
		bpOrConstructor = this.parseMappingBindingPatternOrMappingConstructor(openBrace, members, closeBrace)
	} else {
		this.members.add(memberEnd)
		bpOrConstructor = this.parseMappingBindingPatternOrMappingConstructor(openBrace, members)
	}
	switch bpOrConstructor.kind {
	case MAPPING_CONSTRUCTOR:
		this.switchContext(ParserRuleContext.EXPRESSION_STATEMENT)
		expr := this.parseExpressionRhs(DEFAULT_OP_PRECEDENCE, bpOrConstructor, false, true)
		return this.parseStatementStartWithExprRhs(expr)
	case MAPPING_BINDING_PATTERN:
		this.switchContext(ParserRuleContext.ASSIGNMENT_STMT)
		bindingPattern := this.getBindingPattern(bpOrConstructor, false)
		return this.parseAssignmentStmtRhs(bindingPattern)
	case MAPPING_BP_OR_MAPPING_CONSTRUCTOR:
	default:
		if peek().kind == SyntaxKind.EQUAL_TOKEN {
			this.switchContext(ParserRuleContext.ASSIGNMENT_STMT)
			bindingPattern = this.getBindingPattern(bpOrConstructor, false)
			return this.parseAssignmentStmtRhs(bindingPattern)
		}
		this.switchContext(ParserRuleContext.EXPRESSION_STATEMENT)
		expr = this.getExpression(bpOrConstructor)
		expr = this.parseExpressionRhs(DEFAULT_OP_PRECEDENCE, expr, false, true)
		return this.parseStatementStartWithExprRhs(expr)
	}
}

func (this *BallerinaParser) parseStatementStartingBracedListFirstMember(isOpenBraceMissing bool) internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case READONLY_KEYWORD:
		readonlyKeyword := this.parseReadonlyKeyword()
		return this.bracedListMemberStartsWithReadonly(readonlyKeyword)
	case IDENTIFIER_TOKEN:
		readonlyKeyword = this.STNodeFactory.createEmptyNode()
		return this.parseIdentifierRhsInStmtStartingBrace(readonlyKeyword)
	case STRING_LITERAL_TOKEN:
		key := this.parseStringLiteral()
		if peek().kind == SyntaxKind.COLON_TOKEN {
			readonlyKeyword = this.STNodeFactory.createEmptyNode()
			colon := this.parseColon()
			valueExpr := this.parseExpression()
			return this.STNodeFactory.createSpecificFieldNode(readonlyKeyword, key, colon, valueExpr)
		}
		this.switchContext(ParserRuleContext.BLOCK_STMT)
		this.startContext(ParserRuleContext.AMBIGUOUS_STMT)
		expr := this.parseExpressionRhs(DEFAULT_OP_PRECEDENCE, key, false, true)
		return this.parseStatementStartWithExprRhs(expr)
	case OPEN_BRACKET_TOKEN:
		annots := this.STNodeFactory.createEmptyNodeList()
		return this.parseStatementStartsWithOpenBracket(annots, true)
	case OPEN_BRACE_TOKEN:
		this.switchContext(ParserRuleContext.BLOCK_STMT)
		return this.parseStatementStartsWithOpenBrace()
	case ELLIPSIS_TOKEN:
		return this.parseRestBindingPattern()
	default:
		if isOpenBraceMissing {
			readonlyKeyword = this.STNodeFactory.createEmptyNode()
			return this.parseIdentifierRhsInStmtStartingBrace(readonlyKeyword)
		}
		this.switchContext(ParserRuleContext.BLOCK_STMT)
		return this.parseStatements()
	}
}

func (this *BallerinaParser) bracedListMemberStartsWithReadonly(readonlyKeyword internal.STNode) internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case IDENTIFIER_TOKEN:
		return this.parseIdentifierRhsInStmtStartingBrace(readonlyKeyword)
	case STRING_LITERAL_TOKEN:
		if peek(2).kind == SyntaxKind.COLON_TOKEN {
			key := this.parseStringLiteral()
			colon := this.parseColon()
			valueExpr := this.parseExpression()
			return this.STNodeFactory.createSpecificFieldNode(readonlyKeyword, key, colon, valueExpr)
		}
	default:
		this.switchContext(ParserRuleContext.BLOCK_STMT)
		typeDesc := this.createBuiltinSimpleNameReference(readonlyKeyword)
		return this.parseVarDeclTypeDescRhs(typeDesc, STNodeFactory.createEmptyNodeList(), nil,
			true, false)
	}
}

func (this *BallerinaParser) parseIdentifierRhsInStmtStartingBrace(readonlyKeyword internal.STNode) internal.STNode {
	identifier := this.parseIdentifier(ParserRuleContext.VARIABLE_REF)
	switch peek().kind {
	case COMMA_TOKEN:
	case CLOSE_BRACE_TOKEN:
		colon := this.STNodeFactory.createEmptyNode()
		value := this.STNodeFactory.createEmptyNode()
		return this.STNodeFactory.createSpecificFieldNode(readonlyKeyword, identifier, colon, value)
	case COLON_TOKEN:
		colon = this.parseColon()
		if !this.isEmpty(readonlyKeyword) {
			value = this.parseExpression()
			return this.STNodeFactory.createSpecificFieldNode(readonlyKeyword, identifier, colon, value)
		}
		switch peek().kind {
		case OPEN_BRACKET_TOKEN:
			bindingPatternOrExpr := this.parseListBindingPatternOrListConstructor()
			this.getMappingField(identifier, colon, bindingPatternOrExpr)
		case OPEN_BRACE_TOKEN:
			bindingPatternOrExpr := this.parseMappingBindingPatterOrMappingConstructor()
			this.getMappingField(identifier, colon, bindingPatternOrExpr)
		case ERROR_KEYWORD:
			bindingPatternOrExpr := this.parseErrorBindingPatternOrErrorConstructor()
			this.getMappingField(identifier, colon, bindingPatternOrExpr)
		case IDENTIFIER_TOKEN:
			this.parseQualifiedIdentifierRhsInStmtStartBrace(identifier, colon)
		default:
			expr := this.parseExpression()
			this.getMappingField(identifier, colon, expr)
		}
	default:
		this.switchContext(ParserRuleContext.BLOCK_STMT)
		if !this.isEmpty(readonlyKeyword) {
			this.startContext(ParserRuleContext.VAR_DECL_STMT)
			bindingPattern := this.STNodeFactory.createCaptureBindingPatternNode(identifier)
			typedBindingPattern := this.STNodeFactory.createTypedBindingPatternNode(readonlyKeyword, bindingPattern)
			annots := this.STNodeFactory.createEmptyNodeList()
			varDeclQualifiers := make([]interface{}, 0)
			return this.parseVarDeclRhs(annots, varDeclQualifiers, typedBindingPattern, false)
		}
		this.startContext(ParserRuleContext.AMBIGUOUS_STMT)
		qualifiedIdentifier := this.parseQualifiedIdentifier(identifier, false)
		expr := this.parseTypedBindingPatternOrExprRhs(qualifiedIdentifier, true)
		annots := this.STNodeFactory.createEmptyNodeList()
		return this.parseStmtStartsWithTypedBPOrExprRhs(annots, expr)
	}
}

func (this *BallerinaParser) parseQualifiedIdentifierRhsInStmtStartBrace(identifier internal.STNode, colon internal.STNode) internal.STNode {
	secondIdentifier := this.parseIdentifier(ParserRuleContext.VARIABLE_REF)
	secondNameRef := this.STNodeFactory.createSimpleNameReferenceNode(secondIdentifier)
	if this.isWildcardBP(secondIdentifier) {
		wildcardBP := this.getWildcardBindingPattern(secondIdentifier)
		nameRef := this.STNodeFactory.createSimpleNameReferenceNode(identifier)
		return this.STNodeFactory.createFieldBindingPatternFullNode(nameRef, colon, wildcardBP)
	}
	qualifiedNameRef := this.createQualifiedNameReferenceNode(identifier, colon, secondIdentifier)
	switch peek().kind {
	case COMMA_TOKEN:
		return this.STNodeFactory.createSpecificFieldNode(STNodeFactory.createEmptyNode(), identifier, colon,
			secondNameRef)
	case OPEN_BRACE_TOKEN:
	case IDENTIFIER_TOKEN:
		this.switchContext(ParserRuleContext.BLOCK_STMT)
		this.startContext(ParserRuleContext.VAR_DECL_STMT)
		varDeclQualifiers := make([]interface{}, 0)
		typeBindingPattern := this.parseTypedBindingPatternTypeRhs(qualifiedNameRef, ParserRuleContext.VAR_DECL_STMT)
		annots := this.STNodeFactory.createEmptyNodeList()
		return this.parseVarDeclRhs(annots, varDeclQualifiers, typeBindingPattern, false)
	case OPEN_BRACKET_TOKEN:
		return this.parseMemberRhsInStmtStartWithBrace(identifier, colon, secondIdentifier, secondNameRef)
	case QUESTION_MARK_TOKEN:
		typeDesc := this.parseComplexTypeDescriptor(qualifiedNameRef,
			ParserRuleContext.TYPE_DESC_IN_TYPE_BINDING_PATTERN, true)
		varDeclQualifiers = make([]interface{}, 0)
		typeBindingPattern = this.parseTypedBindingPatternTypeRhs(typeDesc, ParserRuleContext.VAR_DECL_STMT)
		annots = this.STNodeFactory.createEmptyNodeList()
		return this.parseVarDeclRhs(annots, varDeclQualifiers, typeBindingPattern, false)
	case EQUAL_TOKEN:
	case SEMICOLON_TOKEN:
		return this.parseStatementStartWithExprRhs(qualifiedNameRef)
	case PIPE_TOKEN:
	case BITWISE_AND_TOKEN:
	default:
		return this.parseMemberWithExprInRhs(identifier, colon, secondIdentifier, secondNameRef)
	}
}

func (this *BallerinaParser) getBracedListType(member internal.STNode) SyntaxKind {
	switch member.kind {
	case FIELD_BINDING_PATTERN:
	case CAPTURE_BINDING_PATTERN:
	case LIST_BINDING_PATTERN:
	case MAPPING_BINDING_PATTERN:
	case WILDCARD_BINDING_PATTERN:
		return SyntaxKind.MAPPING_BINDING_PATTERN
	case SPECIFIC_FIELD:
		specificFieldNode, ok := member.(*STSpecificFieldNode)
		if !ok {
			panic("getBracedListType: expected STSpecificFieldNode")
		}
		expr := specificFieldNode.valueExpr
		if expr == nil {
			return SyntaxKind.MAPPING_BP_OR_MAPPING_CONSTRUCTOR
		}
		switch expr.kind {
		case SIMPLE_NAME_REFERENCE,
			LIST_BP_OR_LIST_CONSTRUCTOR,
			MAPPING_BP_OR_MAPPING_CONSTRUCTOR:
			SyntaxKind.MAPPING_BP_OR_MAPPING_CONSTRUCTOR
		case ERROR_BINDING_PATTERN:
			SyntaxKind.MAPPING_BINDING_PATTERN
		case ERROR_CONSTRUCTOR:
			errorCtorNode, ok := expr.(*STErrorConstructorExpressionNode)
			if !ok {
				panic("getBracedListType: expected STErrorConstructorExpressionNode")
			}
			if this.isPossibleErrorBindingPattern(errorCtorNode) {
				SyntaxKind.MAPPING_BP_OR_MAPPING_CONSTRUCTOR
			}
			SyntaxKind.MAPPING_CONSTRUCTOR
		default:
			SyntaxKind.MAPPING_CONSTRUCTOR
		}
	case SPREAD_FIELD:
	case COMPUTED_NAME_FIELD:
		return SyntaxKind.MAPPING_CONSTRUCTOR
	case SIMPLE_NAME_REFERENCE:
	case QUALIFIED_NAME_REFERENCE:
	case LIST_BP_OR_LIST_CONSTRUCTOR:
	case MAPPING_BP_OR_MAPPING_CONSTRUCTOR:
	case REST_BINDING_PATTERN:
		return SyntaxKind.MAPPING_BP_OR_MAPPING_CONSTRUCTOR
	case LIST:
		return SyntaxKind.BLOCK_STATEMENT
	default:
		return SyntaxKind.NONE
	}
}

func (this *BallerinaParser) parseMappingBindingPatterOrMappingConstructor() internal.STNode {
	this.startContext(ParserRuleContext.MAPPING_BP_OR_MAPPING_CONSTRUCTOR)
	openBrace := this.parseOpenBrace()
	memberList := make([]interface{}, 0)
	return this.parseMappingBindingPatternOrMappingConstructor(openBrace, memberList)
}

func (this *BallerinaParser) isBracedListEnd(nextTokenKind SyntaxKind) bool {
	switch nextTokenKind {
	case EOF_TOKEN, CLOSE_BRACE_TOKEN:
		true
	default:
		false
	}
}

func (this *BallerinaParser) parseMappingBindingPatternOrMappingConstructor(openBrace internal.STNode, memberList []STNode) internal.STNode {
	nextToken := this.peek()
	for !this.isBracedListEnd(nextToken.kind) {
		member := this.parseMappingBindingPatterOrMappingConstructorMember()
		currentNodeType := this.getTypeOfMappingBPOrMappingCons(member)
		switch currentNodeType {
		case MAPPING_CONSTRUCTOR:
			return this.parseAsMappingConstructor(openBrace, memberList, member)
		case MAPPING_BINDING_PATTERN:
			return this.parseAsMappingBindingPattern(openBrace, memberList, member)
		case MAPPING_BP_OR_MAPPING_CONSTRUCTOR:
		default:
			this.memberList.add(member)
			break
		}
		memberEnd := this.parseMappingFieldEnd()
		if memberEnd == nil {
			break
		}
		this.memberList.add(memberEnd)
		nextToken = this.peek()
	}
	closeBrace := this.parseCloseBrace()
	return this.parseMappingBindingPatternOrMappingConstructor(openBrace, memberList, closeBrace)
}

func (this *BallerinaParser) parseMappingBindingPatterOrMappingConstructorMember() internal.STNode {
	switch peek().kind {
	case IDENTIFIER_TOKEN:
		key := this.parseIdentifier(ParserRuleContext.MAPPING_FIELD_NAME)
		return this.parseMappingFieldRhs(key)
	case STRING_LITERAL_TOKEN:
		readonlyKeyword := this.STNodeFactory.createEmptyNode()
		key = this.parseStringLiteral()
		colon := this.parseColon()
		valueExpr := this.parseExpression()
		return this.STNodeFactory.createSpecificFieldNode(readonlyKeyword, key, colon, valueExpr)
	case OPEN_BRACKET_TOKEN:
		return this.parseComputedField()
	case ELLIPSIS_TOKEN:
		ellipsis := this.parseEllipsis()
		expr := this.parseExpression()
		if expr.kind == SyntaxKind.SIMPLE_NAME_REFERENCE {
			return this.STNodeFactory.createRestBindingPatternNode(ellipsis, expr)
		}
		return this.STNodeFactory.createSpreadFieldNode(ellipsis, expr)
	default:
		this.recover(peek(), ParserRuleContext.MAPPING_BP_OR_MAPPING_CONSTRUCTOR_MEMBER)
		return this.parseMappingBindingPatterOrMappingConstructorMember()
	}
}

func (this *BallerinaParser) parseMappingFieldRhs(key internal.STNode) internal.STNode {
	var colon internal.STNode
	var valueExpr internal.STNode
	switch peek().kind {
	case COLON_TOKEN:
		colon = this.parseColon()
		return this.parseMappingFieldValue(key, colon)
	case COMMA_TOKEN:
	case CLOSE_BRACE_TOKEN:
		readonlyKeyword := this.STNodeFactory.createEmptyNode()
		colon = this.STNodeFactory.createEmptyNode()
		valueExpr = this.STNodeFactory.createEmptyNode()
		return this.STNodeFactory.createSpecificFieldNode(readonlyKeyword, key, colon, valueExpr)
	default:
		token := this.peek()
		this.recover(token, ParserRuleContext.FIELD_BINDING_PATTERN_END)
		readonlyKeyword = this.STNodeFactory.createEmptyNode()
		return this.parseSpecificFieldRhs(readonlyKeyword, key)
	}
}

func (this *BallerinaParser) parseMappingFieldValue(key internal.STNode, colon internal.STNode) internal.STNode {
	var expr internal.STNode
	switch peek().kind {
	case IDENTIFIER_TOKEN:
		expr = this.parseExpression()
	case OPEN_BRACKET_TOKEN:
		expr = this.parseListBindingPatternOrListConstructor()
	case OPEN_BRACE_TOKEN:
		expr = this.parseMappingBindingPatterOrMappingConstructor()
	default:
		expr = this.parseExpression()
	}
	if this.isBindingPattern(expr.kind) {
		key = this.STNodeFactory.createSimpleNameReferenceNode(key)
		return this.STNodeFactory.createFieldBindingPatternFullNode(key, colon, expr)
	}
	readonlyKeyword := this.STNodeFactory.createEmptyNode()
	return this.STNodeFactory.createSpecificFieldNode(readonlyKeyword, key, colon, expr)
}

func (this *BallerinaParser) isBindingPattern(kind SyntaxKind) bool {
	switch kind {
	case FIELD_BINDING_PATTERN,
		MAPPING_BINDING_PATTERN,
		CAPTURE_BINDING_PATTERN,
		LIST_BINDING_PATTERN,
		WILDCARD_BINDING_PATTERN:
		true
	default:
		false
	}
}

func (this *BallerinaParser) getTypeOfMappingBPOrMappingCons(memberNode internal.STNode) SyntaxKind {
	switch memberNode.kind {
	case FIELD_BINDING_PATTERN:
	case MAPPING_BINDING_PATTERN:
	case CAPTURE_BINDING_PATTERN:
	case LIST_BINDING_PATTERN:
	case WILDCARD_BINDING_PATTERN:
		return SyntaxKind.MAPPING_BINDING_PATTERN
	case SPECIFIC_FIELD:
		specificFieldNode, ok := memberNode.(*STSpecificFieldNode)
		if !ok {
			panic("getTypeOfMappingBPOrMappingCons: expected STSpecificFieldNode")
		}
		expr := specificFieldNode.valueExpr
		if (((expr == nil) || (expr.kind == SyntaxKind.SIMPLE_NAME_REFERENCE)) || (expr.kind == SyntaxKind.LIST_BP_OR_LIST_CONSTRUCTOR)) || (expr.kind == SyntaxKind.MAPPING_BP_OR_MAPPING_CONSTRUCTOR) {
			return SyntaxKind.MAPPING_BP_OR_MAPPING_CONSTRUCTOR
		}
		return SyntaxKind.MAPPING_CONSTRUCTOR
	case SPREAD_FIELD:
	case COMPUTED_NAME_FIELD:
		return SyntaxKind.MAPPING_CONSTRUCTOR
	case MAPPING_BP_OR_MAPPING_CONSTRUCTOR:
	case SIMPLE_NAME_REFERENCE:
	case QUALIFIED_NAME_REFERENCE:
	case LIST_BP_OR_LIST_CONSTRUCTOR:
	case REST_BINDING_PATTERN:
	default:
		return SyntaxKind.MAPPING_BP_OR_MAPPING_CONSTRUCTOR
	}
}

func (this *BallerinaParser) parseMappingBindingPatternOrMappingConstructor(openBrace internal.STNode, members []STNode, closeBrace internal.STNode) internal.STNode {
	this.endContext()
	return nil
}

func (this *BallerinaParser) parseAsMappingBindingPattern(openBrace internal.STNode, members []STNode, member internal.STNode) internal.STNode {
	this.members.add(member)
	members = this.getBindingPatternsList(members, false)
	this.switchContext(ParserRuleContext.MAPPING_BINDING_PATTERN)
	return this.parseMappingBindingPattern(openBrace, members, member)
}

func (this *BallerinaParser) parseListBindingPatternOrListConstructor() internal.STNode {
	this.startContext(ParserRuleContext.BRACKETED_LIST)
	openBracket := this.parseOpenBracket()
	memberList := make([]interface{}, 0)
	return this.parseListBindingPatternOrListConstructor(openBracket, memberList, false)
}

func (this *BallerinaParser) parseListBindingPatternOrListConstructor(openBracket internal.STNode, memberList []STNode, isRoot bool) internal.STNode {
	nextToken := this.peek()
	for !this.isBracketedListEnd(nextToken.kind) {
		member := this.parseListBindingPatternOrListConstructorMember()
		currentNodeType := this.getParsingNodeTypeOfListBPOrListCons(member)
		switch currentNodeType {
		case LIST_CONSTRUCTOR:
			return this.parseAsListConstructor(openBracket, memberList, member, isRoot)
		case LIST_BINDING_PATTERN:
			return this.parseAsListBindingPattern(openBracket, memberList, member, isRoot)
		case LIST_BP_OR_LIST_CONSTRUCTOR:
		default:
			this.memberList.add(member)
			break
		}
		memberEnd := this.parseBracketedListMemberEnd()
		if memberEnd == nil {
			break
		}
		this.memberList.add(memberEnd)
		nextToken = this.peek()
	}
	closeBracket := this.parseCloseBracket()
	return this.parseListBindingPatternOrListConstructor(openBracket, memberList, closeBracket, isRoot)
}

func (this *BallerinaParser) parseListBindingPatternOrListConstructorMember() internal.STNode {
	nextToken := this.peek()
	switch nextToken.kind {
	case OPEN_BRACKET_TOKEN:
		return this.parseListBindingPatternOrListConstructor()
	case IDENTIFIER_TOKEN:
		identifier := this.parseQualifiedIdentifier(ParserRuleContext.VARIABLE_REF)
		if this.isWildcardBP(identifier) {
			return this.getWildcardBindingPattern(identifier)
		}
		return this.parseExpressionRhs(DEFAULT_OP_PRECEDENCE, identifier, false, false)
	case OPEN_BRACE_TOKEN:
		return this.parseMappingBindingPatterOrMappingConstructor()
	case ELLIPSIS_TOKEN:
		return this.parseRestBindingOrSpreadMember()
	default:
		if this.isValidExpressionStart(nextToken.kind, 1) {
			return this.parseExpression()
		}
		this.recover(peek(), ParserRuleContext.LIST_BP_OR_LIST_CONSTRUCTOR_MEMBER)
		return this.parseListBindingPatternOrListConstructorMember()
	}
}

func (this *BallerinaParser) getParsingNodeTypeOfListBPOrListCons(memberNode internal.STNode) SyntaxKind {
	switch memberNode.kind {
	case CAPTURE_BINDING_PATTERN,
		LIST_BINDING_PATTERN,
		MAPPING_BINDING_PATTERN,
		WILDCARD_BINDING_PATTERN:
		SyntaxKind.LIST_BINDING_PATTERN
	case SIMPLE_NAME_REFERENCE, // member is a simple type-ref/var-ref
		LIST_BP_OR_LIST_CONSTRUCTOR, // member is again ambiguous
		MAPPING_BP_OR_MAPPING_CONSTRUCTOR,
		REST_BINDING_PATTERN:
		SyntaxKind.LIST_BP_OR_LIST_CONSTRUCTOR
	default:
		SyntaxKind.LIST_CONSTRUCTOR
	}
}

func (this *BallerinaParser) parseAsListConstructor(openBracket internal.STNode, memberList []STNode, member internal.STNode, isRoot bool) internal.STNode {
	this.memberList.add(member)
	memberList = this.getExpressionList(memberList, false)
	this.switchContext(ParserRuleContext.LIST_CONSTRUCTOR)
	listMembers := this.parseListMembers(memberList)
	closeBracket := this.parseCloseBracket()
	listConstructor := this.STNodeFactory.createListConstructorExpressionNode(openBracket, listMembers, closeBracket)
	this.endContext()
	expr := this.parseExpressionRhs(DEFAULT_OP_PRECEDENCE, listConstructor, false, true)
	if !isRoot {
		return expr
	}
	return this.parseStatementStartWithExprRhs(expr)
}

func (this *BallerinaParser) parseListBindingPatternOrListConstructor(openBracket internal.STNode, members []STNode, closeBracket internal.STNode, isRoot bool) internal.STNode {
	var lbpOrListCons internal.STNode
	switch peek().kind {
	case COMMA_TOKEN:
	case CLOSE_BRACE_TOKEN:
	case CLOSE_BRACKET_TOKEN:
		if !isRoot {
			this.endContext()
			return nil
		}
	default:
		nextTokenKind := peek().kind
		if this.isValidExprRhsStart(nextTokenKind, closeBracket.kind) || ((nextTokenKind == SyntaxKind.SEMICOLON_TOKEN) && isRoot) {
			members = this.getExpressionList(members, false)
			memberExpressions := this.STNodeFactory.createNodeList(members)
			lbpOrListCons = this.STNodeFactory.createListConstructorExpressionNode(openBracket, memberExpressions,
				closeBracket)
			lbpOrListCons = this.parseExpressionRhs(DEFAULT_OP_PRECEDENCE, lbpOrListCons, false, true)
			break
		}
		members = this.getBindingPatternsList(members, true)
		bindingPatternsNode := this.STNodeFactory.createNodeList(members)
		lbpOrListCons = this.STNodeFactory.createListBindingPatternNode(openBracket, bindingPatternsNode,
			closeBracket)
		break
	}
	this.endContext()
	if !isRoot {
		return lbpOrListCons
	}
	if lbpOrListCons.kind == SyntaxKind.LIST_BINDING_PATTERN {
		return this.parseAssignmentStmtRhs(lbpOrListCons)
	} else {
		return this.parseStatementStartWithExprRhs(lbpOrListCons)
	}
}

func (this *BallerinaParser) parseMemberRhsInStmtStartWithBrace(identifier internal.STNode, colon internal.STNode, secondIdentifier internal.STNode, secondNameRef internal.STNode) internal.STNode {
	typedBPOrExpr := this.parseTypedBindingPatternOrMemberAccess(secondNameRef, false, true, ParserRuleContext.AMBIGUOUS_STMT)
	if this.isExpression(typedBPOrExpr.kind) {
		return this.parseMemberWithExprInRhs(identifier, colon, secondIdentifier, typedBPOrExpr)
	}
	this.switchContext(ParserRuleContext.BLOCK_STMT)
	this.startContext(ParserRuleContext.VAR_DECL_STMT)
	varDeclQualifiers := make([]interface{}, 0)
	annots := this.STNodeFactory.createEmptyNodeList()
	typedBP := internal.STTypedBindingPatternNode(typedBPOrExpr)
	qualifiedNameRef := this.createQualifiedNameReferenceNode(identifier, colon, secondIdentifier)
	newTypeDesc := this.mergeQualifiedNameWithTypeDesc(qualifiedNameRef, typedBP.typeDescriptor)
	newTypeBP := this.STNodeFactory.createTypedBindingPatternNode(newTypeDesc, typedBP.bindingPattern)
	publicQualifier := this.STNodeFactory.createEmptyNode()
	return this.parseVarDeclRhs(annots, publicQualifier, varDeclQualifiers, newTypeBP, false)
}

func (this *BallerinaParser) parseMemberWithExprInRhs(identifier internal.STNode, colon internal.STNode, secondIdentifier internal.STNode, memberAccessExpr internal.STNode) internal.STNode {
	expr := this.parseExpressionRhs(DEFAULT_OP_PRECEDENCE, memberAccessExpr, false, true)
	switch peek().kind {
	case COMMA_TOKEN:
	case CLOSE_BRACE_TOKEN:
		this.switchContext(ParserRuleContext.EXPRESSION_STATEMENT)
		readonlyKeyword := this.STNodeFactory.createEmptyNode()
		return this.STNodeFactory.createSpecificFieldNode(readonlyKeyword, identifier, colon, expr)
	case EQUAL_TOKEN:
	case SEMICOLON_TOKEN:
	default:
		this.switchContext(ParserRuleContext.BLOCK_STMT)
		this.startContext(ParserRuleContext.EXPRESSION_STATEMENT)
		qualifiedName := this.createQualifiedNameReferenceNode(identifier, colon, secondIdentifier)
		updatedExpr := this.mergeQualifiedNameWithExpr(qualifiedName, expr)
		return this.parseStatementStartWithExprRhs(updatedExpr)
	}
}

func (this *BallerinaParser) parseInferredTypeDescDefaultOrExpression() internal.STNode {
	nextToken := this.peek()
	nextTokenKind := nextToken.kind
	if nextTokenKind == SyntaxKind.LT_TOKEN {
		return this.parseInferredTypeDescDefaultOrExpression(consume())
	}
	if this.isValidExprStart(nextTokenKind) {
		return this.parseExpression()
	}
	this.recover(nextToken, ParserRuleContext.EXPR_START_OR_INFERRED_TYPEDESC_DEFAULT_START)
	return this.parseInferredTypeDescDefaultOrExpression()
}

func (this *BallerinaParser) parseInferredTypeDescDefaultOrExpression(ltToken internal.STToken) internal.STNode {
	nextToken := this.peek()
	if nextToken.kind == SyntaxKind.GT_TOKEN {
		return this.STNodeFactory.createInferredTypedescDefaultNode(ltToken, consume())
	}
	if this.isTypeStartingToken(nextToken.kind) || (nextToken.kind == SyntaxKind.AT_TOKEN) {
		this.startContext(ParserRuleContext.TYPE_CAST)
		expr := this.parseTypeCastExpr(ltToken, true, false, false)
		return this.parseExpressionRhs(DEFAULT_OP_PRECEDENCE, expr, true, false)
	}
	this.recover(nextToken, ParserRuleContext.TYPE_CAST_PARAM_START_OR_INFERRED_TYPEDESC_DEFAULT_END)
	return this.parseInferredTypeDescDefaultOrExpression(ltToken)
}

func (this *BallerinaParser) mergeQualifiedNameWithExpr(qualifiedName internal.STNode, exprOrAction internal.STNode) internal.STNode {
	switch exprOrAction.kind {
	case SIMPLE_NAME_REFERENCE:
		return qualifiedName
	case BINARY_EXPRESSION:
		binaryExpr := internal.STBinaryExpressionNode(exprOrAction)
		newLhsExpr := this.mergeQualifiedNameWithExpr(qualifiedName, binaryExpr.lhsExpr)
		return this.STNodeFactory.createBinaryExpressionNode(binaryExpr.kind, newLhsExpr, binaryExpr.operator,
			binaryExpr.rhsExpr)
	case FIELD_ACCESS:
		fieldAccess := internal.STFieldAccessExpressionNode(exprOrAction)
		newLhsExpr = this.mergeQualifiedNameWithExpr(qualifiedName, fieldAccess.expression)
		return this.STNodeFactory.createFieldAccessExpressionNode(newLhsExpr, fieldAccess.dotToken,
			fieldAccess.fieldName)
	case INDEXED_EXPRESSION:
		memberAccess := internal.STIndexedExpressionNode(exprOrAction)
		newLhsExpr = this.mergeQualifiedNameWithExpr(qualifiedName, memberAccess.containerExpression)
		return this.STNodeFactory.createIndexedExpressionNode(newLhsExpr, memberAccess.openBracket,
			memberAccess.keyExpression, memberAccess.closeBracket)
	case TYPE_TEST_EXPRESSION:
		typeTest := internal.STTypeTestExpressionNode(exprOrAction)
		newLhsExpr = this.mergeQualifiedNameWithExpr(qualifiedName, typeTest.expression)
		return this.STNodeFactory.createTypeTestExpressionNode(newLhsExpr, typeTest.isKeyword,
			typeTest.typeDescriptor)
	case ANNOT_ACCESS:
		annotAccess := internal.STAnnotAccessExpressionNode(exprOrAction)
		newLhsExpr = this.mergeQualifiedNameWithExpr(qualifiedName, annotAccess.expression)
		return this.STNodeFactory.createFieldAccessExpressionNode(newLhsExpr, annotAccess.annotChainingToken,
			annotAccess.annotTagReference)
	case OPTIONAL_FIELD_ACCESS:
		optionalFieldAccess := internal.STOptionalFieldAccessExpressionNode(exprOrAction)
		newLhsExpr = this.mergeQualifiedNameWithExpr(qualifiedName, optionalFieldAccess.expression)
		return this.STNodeFactory.createFieldAccessExpressionNode(newLhsExpr,
			optionalFieldAccess.optionalChainingToken, optionalFieldAccess.fieldName)
	case CONDITIONAL_EXPRESSION:
		conditionalExpr := internal.STConditionalExpressionNode(exprOrAction)
		newLhsExpr = this.mergeQualifiedNameWithExpr(qualifiedName, conditionalExpr.lhsExpression)
		return this.STNodeFactory.createConditionalExpressionNode(newLhsExpr, conditionalExpr.questionMarkToken,
			conditionalExpr.middleExpression, conditionalExpr.colonToken, conditionalExpr.endExpression)
	case REMOTE_METHOD_CALL_ACTION:
		remoteCall := internal.STRemoteMethodCallActionNode(exprOrAction)
		newLhsExpr = this.mergeQualifiedNameWithExpr(qualifiedName, remoteCall.expression)
		return this.STNodeFactory.createRemoteMethodCallActionNode(newLhsExpr, remoteCall.rightArrowToken,
			remoteCall.methodName, remoteCall.openParenToken, remoteCall.arguments,
			remoteCall.closeParenToken)
	case ASYNC_SEND_ACTION:
		asyncSend := internal.STAsyncSendActionNode(exprOrAction)
		newLhsExpr = this.mergeQualifiedNameWithExpr(qualifiedName, asyncSend.expression)
		return this.STNodeFactory.createAsyncSendActionNode(newLhsExpr, asyncSend.rightArrowToken,
			asyncSend.peerWorker)
	case SYNC_SEND_ACTION:
		syncSend := internal.STSyncSendActionNode(exprOrAction)
		newLhsExpr = this.mergeQualifiedNameWithExpr(qualifiedName, syncSend.expression)
		return this.STNodeFactory.createAsyncSendActionNode(newLhsExpr, syncSend.syncSendToken, syncSend.peerWorker)
	case FUNCTION_CALL:
		funcCall := internal.STFunctionCallExpressionNode(exprOrAction)
		return this.STNodeFactory.createFunctionCallExpressionNode(qualifiedName, funcCall.openParenToken,
			funcCall.arguments, funcCall.closeParenToken)
	default:
		return exprOrAction
	}
}

func (this *BallerinaParser) mergeQualifiedNameWithTypeDesc(qualifiedName internal.STNode, typeDesc internal.STNode) internal.STNode {
	switch typeDesc.kind {
	case SIMPLE_NAME_REFERENCE:
		return qualifiedName
	case ARRAY_TYPE_DESC:
		arrayTypeDesc := internal.STArrayTypeDescriptorNode(typeDesc)
		newMemberType := this.mergeQualifiedNameWithTypeDesc(qualifiedName, arrayTypeDesc.memberTypeDesc)
		return this.STNodeFactory.createArrayTypeDescriptorNode(newMemberType, arrayTypeDesc.dimensions)
	case UNION_TYPE_DESC:
		unionTypeDesc := internal.STUnionTypeDescriptorNode(typeDesc)
		newlhsType := this.mergeQualifiedNameWithTypeDesc(qualifiedName, unionTypeDesc.leftTypeDesc)
		return this.mergeTypesWithUnion(newlhsType, unionTypeDesc.pipeToken, unionTypeDesc.rightTypeDesc)
	case INTERSECTION_TYPE_DESC:
		intersectionTypeDesc := internal.STIntersectionTypeDescriptorNode(typeDesc)
		newlhsType = this.mergeQualifiedNameWithTypeDesc(qualifiedName, intersectionTypeDesc.leftTypeDesc)
		return this.mergeTypesWithIntersection(newlhsType, intersectionTypeDesc.bitwiseAndToken,
			intersectionTypeDesc.rightTypeDesc)
	case OPTIONAL_TYPE_DESC:
		optionalType := internal.STOptionalTypeDescriptorNode(typeDesc)
		newMemberType = this.mergeQualifiedNameWithTypeDesc(qualifiedName, optionalType.typeDescriptor)
		return this.STNodeFactory.createOptionalTypeDescriptorNode(newMemberType, optionalType.questionMarkToken)
	default:
		return typeDesc
	}
}

func (this *BallerinaParser) getTupleMemberList(ambiguousList []STNode) []STNode {
	tupleMemberList := make([]interface{}, 0)
	for _, item := range ambiguousList {
		if item.kind == SyntaxKind.COMMA_TOKEN {
			this.tupleMemberList.add(item)
		} else {
			this.tupleMemberList.add(STNodeFactory.createMemberTypeDescriptorNode(STNodeFactory.createEmptyNodeList(),
				getTypeDescFromExpr(item)))
		}
	}
	return tupleMemberList
}

func (this *BallerinaParser) getTypeDescFromExpr(expression internal.STNode) internal.STNode {
	if this.isDefiniteTypeDesc(expression.kind) || (expression.kind == SyntaxKind.COMMA_TOKEN) {
		return expression
	}
	switch expression.kind {
	case INDEXED_EXPRESSION:
		indexedExpr, ok := expression.(*STIndexedExpressionNode)
		if !ok {
			panic("getTypeDescFromExpr: expected STIndexedExpressionNode")
		}
		return this.parseArrayTypeDescriptorNode(indexedExpr)
	case NUMERIC_LITERAL:
	case BOOLEAN_LITERAL:
	case STRING_LITERAL:
	case NULL_LITERAL:
	case UNARY_EXPRESSION:
		return this.STNodeFactory.createSingletonTypeDescriptorNode(expression)
	case TYPE_REFERENCE_TYPE_DESC:
		typeRefNode, ok := expression.(*STTypeReferenceTypeDescNode)
		if !ok {
			panic("getTypeDescFromExpr: expected STTypeReferenceTypeDescNode")
		}
		return typeRefNode.typeRef
	case BRACED_EXPRESSION:
		bracedExpr := internal.STBracedExpressionNode(expression)
		typeDesc := this.getTypeDescFromExpr(bracedExpr.expression)
		return this.STNodeFactory.createParenthesisedTypeDescriptorNode(bracedExpr.openParen, typeDesc,
			bracedExpr.closeParen)
	case NIL_LITERAL:
		nilLiteral := internal.STNilLiteralNode(expression)
		return this.STNodeFactory.createNilTypeDescriptorNode(nilLiteral.openParenToken, nilLiteral.closeParenToken)
	case BRACKETED_LIST:
	case LIST_BP_OR_LIST_CONSTRUCTOR:
	case TUPLE_TYPE_DESC_OR_LIST_CONST:
		innerList := internal.STAmbiguousCollectionNode(expression)
		memberTypeDescs := this.STNodeFactory.createNodeList(getTupleMemberList(innerList.members))
		return this.STNodeFactory.createTupleTypeDescriptorNode(innerList.collectionStartToken, memberTypeDescs,
			innerList.collectionEndToken)
	case BINARY_EXPRESSION:
		binaryExpr := internal.STBinaryExpressionNode(expression)
		switch binaryExpr.operator.kind {
		case PIPE_TOKEN:
		case BITWISE_AND_TOKEN:
			lhsTypeDesc := this.getTypeDescFromExpr(binaryExpr.lhsExpr)
			rhsTypeDesc := this.getTypeDescFromExpr(binaryExpr.rhsExpr)
			return this.mergeTypes(lhsTypeDesc, binaryExpr.operator, rhsTypeDesc)
		default:
			break
		}
		return expression
	case SIMPLE_NAME_REFERENCE:
	case QUALIFIED_NAME_REFERENCE:
		return expression
	default:
		simpleTypeDescIdentifier := this.SyntaxErrors.createMissingTokenWithDiagnostics(
			SyntaxKind.IDENTIFIER_TOKEN, DiagnosticErrorCode.ERROR_MISSING_TYPE_DESC)
		simpleTypeDescIdentifier = this.SyntaxErrors.cloneWithTrailingInvalidNodeMinutiae(simpleTypeDescIdentifier,
			expression)
		return this.STNodeFactory.createSimpleNameReferenceNode(simpleTypeDescIdentifier)
	}
}

func (this *BallerinaParser) getBindingPatternsList(ambibuousList []STNode, isListBP bool) []STNode {
	bindingPatterns := make([]interface{}, 0)
	for _, item := range ambibuousList {
		this.bindingPatterns.add(getBindingPattern(item, isListBP))
	}
	return bindingPatterns
}

func (this *BallerinaParser) getBindingPattern(ambiguousNode internal.STNode, isListBP bool) internal.STNode {
	errorCode := DiagnosticErrorCode.ERROR_INVALID_BINDING_PATTERN
	if this.isEmpty(ambiguousNode) {
		return nil
	}
	switch ambiguousNode.kind {
	case WILDCARD_BINDING_PATTERN:
	case CAPTURE_BINDING_PATTERN:
	case LIST_BINDING_PATTERN:
	case MAPPING_BINDING_PATTERN:
	case ERROR_BINDING_PATTERN:
	case REST_BINDING_PATTERN:
	case FIELD_BINDING_PATTERN:
	case NAMED_ARG_BINDING_PATTERN:
	case COMMA_TOKEN:
		return ambiguousNode
	case SIMPLE_NAME_REFERENCE:
		simpleNameNode, ok := ambiguousNode.(*STSimpleNameReferenceNode)
		if !ok {
			panic("getBindingPattern: expected STSimpleNameReferenceNode")
		}
		varName := simpleNameNode.name
		return this.createCaptureOrWildcardBP(varName)
	case QUALIFIED_NAME_REFERENCE:
		if isListBP {
			errorCode = DiagnosticErrorCode.ERROR_FIELD_BP_INSIDE_LIST_BP
			break
		}
		qualifiedName := internal.STQualifiedNameReferenceNode(ambiguousNode)
		fieldName := this.STNodeFactory.createSimpleNameReferenceNode(qualifiedName.modulePrefix)
		return this.STNodeFactory.createFieldBindingPatternFullNode(fieldName, qualifiedName.colon,
			createCaptureOrWildcardBP(qualifiedName.identifier))
	case BRACKETED_LIST:
	case LIST_BP_OR_LIST_CONSTRUCTOR:
		innerList := internal.STAmbiguousCollectionNode(ambiguousNode)
		memberBindingPatterns := this.STNodeFactory.createNodeList(getBindingPatternsList(innerList.members, true))
		return this.STNodeFactory.createListBindingPatternNode(innerList.collectionStartToken, memberBindingPatterns,
			innerList.collectionEndToken)
	case MAPPING_BP_OR_MAPPING_CONSTRUCTOR:
		innerList = internal.STAmbiguousCollectionNode(ambiguousNode)
		bindingPatterns := make([]interface{}, 0)
		i := 0
		for ; i < len(innerList.members); i++ {
			bp := this.getBindingPattern(innerList.members.get(i), false)
			this.bindingPatterns.add(bp)
			if bp.kind == SyntaxKind.REST_BINDING_PATTERN {
				break
			}
		}
		memberBindingPatterns = this.STNodeFactory.createNodeList(bindingPatterns)
		return this.STNodeFactory.createMappingBindingPatternNode(innerList.collectionStartToken,
			memberBindingPatterns, innerList.collectionEndToken)
	case SPECIFIC_FIELD:
		field := internal.STSpecificFieldNode(ambiguousNode)
		fieldName = this.STNodeFactory.createSimpleNameReferenceNode(field.fieldName)
		if field.valueExpr == nil {
			return this.STNodeFactory.createFieldBindingPatternVarnameNode(fieldName)
		}
		return this.STNodeFactory.createFieldBindingPatternFullNode(fieldName, field.colon,
			getBindingPattern(field.valueExpr, false))
	case ERROR_CONSTRUCTOR:
		errorCons := internal.STErrorConstructorExpressionNode(ambiguousNode)
		args := errorCons.arguments
		size := this.args.bucketCount()
		bindingPatterns = make([]interface{}, 0)
		i := 0
		for ; i < size; i++ {
			arg := this.args.childInBucket(i)
			this.bindingPatterns.add(getBindingPattern(arg, false))
		}
		argListBindingPatterns := this.STNodeFactory.createNodeList(bindingPatterns)
		return this.STNodeFactory.createErrorBindingPatternNode(errorCons.errorKeyword, errorCons.typeReference,
			errorCons.openParenToken, argListBindingPatterns, errorCons.closeParenToken)
	case POSITIONAL_ARG:
		positionalArg := internal.STPositionalArgumentNode(ambiguousNode)
		return this.getBindingPattern(positionalArg.expression, false)
	case NAMED_ARG:
		namedArg := internal.STNamedArgumentNode(ambiguousNode)
		argNameNode, ok := namedArg.argumentName.(*STSimpleNameReferenceNode)
		if !ok {
			panic("getBindingPattern: expected STSimpleNameReferenceNode for named argument")
		}
		bindingPatternArgName := argNameNode.name
		return this.STNodeFactory.createNamedArgBindingPatternNode(bindingPatternArgName, namedArg.equalsToken,
			getBindingPattern(namedArg.expression, false))
	case REST_ARG:
		restArg := internal.STRestArgumentNode(ambiguousNode)
		return this.STNodeFactory.createRestBindingPatternNode(restArg.ellipsis, restArg.expression)
	}
	identifier := this.SyntaxErrors.createMissingToken(SyntaxKind.IDENTIFIER_TOKEN)
	identifier = this.SyntaxErrors.cloneWithLeadingInvalidNodeMinutiae(identifier, ambiguousNode, errorCode)
	return this.STNodeFactory.createCaptureBindingPatternNode(identifier)
}

func (this *BallerinaParser) getExpressionList(ambibuousList []STNode, isMappingConstructor bool) []STNode {
	exprList := make([]STNode, 0)
	for _, item := range ambibuousList {
		this.exprList.add(getExpression(item, isMappingConstructor))
	}
	return exprList
}

func (this *BallerinaParser) getExpression(ambiguousNode internal.STNode) internal.STNode {
	return this.getExpression(ambiguousNode, false)
}

// func (this *BallerinaParser) getExpression(ambiguousNode internal.STNode, isInMappingConstructor bool) internal.STNode {
// if (((this.isEmpty(ambiguousNode) || (this.isDefiniteExpr(ambiguousNode.kind) && (ambiguousNode.kind != SyntaxKind.INDEXED_EXPRESSION))) || this.isDefiniteAction(ambiguousNode.kind)) || (ambiguousNode.kind == SyntaxKind.COMMA_TOKEN)) {
// return ambiguousNode
// }
// switch ambiguousNode.kind {
// case BRACKETED_LIST:
// case LIST_BP_OR_LIST_CONSTRUCTOR:
// case TUPLE_TYPE_DESC_OR_LIST_CONST:
// innerList := internal.STAmbiguousCollectionNode(ambiguousNode)
// memberExprs := this.STNodeFactory.createNodeList(getExpressionList(innerList.members, false))
// return this.STNodeFactory.createListConstructorExpressionNode(innerList.collectionStartToken, memberExprs,
//                         innerList.collectionEndToken)
// case MAPPING_BP_OR_MAPPING_CONSTRUCTOR:
// innerList = internal.STAmbiguousCollectionNode(ambiguousNode)
// fieldList := make([]interface{}, 0)
// i := 0
// for ; (i < len(innerList.members)); i++ {
// field := this.innerList.members.get(i)
// var fieldNode internal.STNode
// if (field.kind == SyntaxKind.QUALIFIED_NAME_REFERENCE) {
// qualifiedNameRefNode := internal.STQualifiedNameReferenceNode(field)
// readOnlyKeyword := this.STNodeFactory.createEmptyNode()
// fieldName := qualifiedNameRefNode.modulePrefix
// colon := qualifiedNameRefNode.colon
// valueExpr := this.getExpression(qualifiedNameRefNode.identifier)
// fieldNode = this.STNodeFactory.createSpecificFieldNode(readOnlyKeyword, fieldName, colon, valueExpr)
// }else {
// fieldNode = this.getExpression(field, true)
// }
// this.fieldList.add(fieldNode)
// }
// fields := this.STNodeFactory.createNodeList(fieldList)
// return this.STNodeFactory.createMappingConstructorExpressionNode(innerList.collectionStartToken, fields,
//                         innerList.collectionEndToken)
// case REST_BINDING_PATTERN:
// restBindingPattern := internal.STRestBindingPatternNode(ambiguousNode)
// if isInMappingConstructor {
// return this.STNodeFactory.createSpreadFieldNode(restBindingPattern.ellipsisToken,
//                             restBindingPattern.variableName)
// }
// return this.STNodeFactory.createSpreadMemberNode(restBindingPattern.ellipsisToken,
//                         restBindingPattern.variableName)
// case SPECIFIC_FIELD:
// field := internal.STSpecificFieldNode(ambiguousNode)
// return this.STNodeFactory.createSpecificFieldNode(field.readonlyKeyword, field.fieldName, field.colon,
//                         getExpression(field.valueExpr))
// case ERROR_CONSTRUCTOR:
// errorCons := internal.STErrorConstructorExpressionNode(ambiguousNode)
// errorArgs := this.getErrorArgList(errorCons.arguments)
// return this.STNodeFactory.createErrorConstructorExpressionNode(errorCons.errorKeyword,
//                         errorCons.typeReference, errorCons.openParenToken, errorArgs, errorCons.closeParenToken)
// case IDENTIFIER_TOKEN:
// return this.STNodeFactory.createSimpleNameReferenceNode(ambiguousNode)
// case INDEXED_EXPRESSION:
// indexedExpressionNode := internal.STIndexedExpressionNode(ambiguousNode)
// keys := internal.STNodeList(indexedExpressionNode.keyExpression)
// if (!this.keys.isEmpty()) {
// return ambiguousNode
// }
// lhsExpr := indexedExpressionNode.containerExpression
// openBracket := indexedExpressionNode.openBracket
// closeBracket := indexedExpressionNode.closeBracket
// missingVarRef := this.STNodeFactory
//                         .createSimpleNameReferenceNode(SyntaxErrors.createMissingToken(SyntaxKind.IDENTIFIER_TOKEN))
// keyExpr := this.STNodeFactory.createNodeList(missingVarRef)
// closeBracket = this.SyntaxErrors.addDiagnostic(closeBracket,
//                         DiagnosticErrorCode.ERROR_MISSING_KEY_EXPR_IN_MEMBER_ACCESS_EXPR)
// return this.STNodeFactory.createIndexedExpressionNode(lhsExpr, openBracket, keyExpr, closeBracket)
// case SIMPLE_NAME_REFERENCE:
// case QUALIFIED_NAME_REFERENCE:
// case COMPUTED_NAME_FIELD:
// case SPREAD_FIELD:
// case SPREAD_MEMBER:
// return ambiguousNode
// default:
// simpleVarRef := this.SyntaxErrors.createMissingTokenWithDiagnostics(SyntaxKind.IDENTIFIER_TOKEN,
//                         DiagnosticErrorCode.ERROR_MISSING_EXPRESSION)
// simpleVarRef = this.SyntaxErrors.cloneWithTrailingInvalidNodeMinutiae(simpleVarRef, ambiguousNode)
// return this.STNodeFactory.createSimpleNameReferenceNode(simpleVarRef)
// }
// }

func (this *BallerinaParser) getMappingField(identifier internal.STNode, colon internal.STNode, bindingPatternOrExpr internal.STNode) internal.STNode {
	simpleNameRef := this.STNodeFactory.createSimpleNameReferenceNode(identifier)
	switch bindingPatternOrExpr.kind {
	case LIST_BINDING_PATTERN,
		MAPPING_BINDING_PATTERN:
		this.STNodeFactory.createFieldBindingPatternFullNode(simpleNameRef, colon, bindingPatternOrExpr)
	case LIST_CONSTRUCTOR,
		MAPPING_CONSTRUCTOR:
		readonlyKeyword := this.STNodeFactory.createEmptyNode()
		this.STNodeFactory.createSpecificFieldNode(readonlyKeyword, identifier, colon, bindingPatternOrExpr)
	default:
		readonlyKeyword := this.STNodeFactory.createEmptyNode()
		this.STNodeFactory.createSpecificFieldNode(readonlyKeyword, identifier, colon, bindingPatternOrExpr)
	}
}

func (this *BallerinaParser) recover(nextToken internal.STToken, currentCtx ParserRuleContext) Solution {
	if this.isInsideABlock(nextToken) {
		return this.this.recover(nextToken, currentCtx, true)
	} else {
		return this.this.recover(nextToken, currentCtx, false)
	}
}

func (this *BallerinaParser) isInsideABlock(nextToken internal.STToken) bool {
	if nextToken.kind != SyntaxKind.CLOSE_BRACE_TOKEN {
		return false
	}
	for _, ctx := range this.this.errorHandler.getContextStack() {
		if this.isBlockContext(ctx) {
			return true
		}
	}
	return false
}

func (this *BallerinaParser) isBlockContext(ctx ParserRuleContext) bool {
	switch ctx {
	case FUNC_BODY_BLOCK,
		CLASS_MEMBER,
		OBJECT_CONSTRUCTOR_MEMBER,
		OBJECT_TYPE_MEMBER,
		BLOCK_STMT,
		MATCH_BODY,
		MAPPING_MATCH_PATTERN,
		MAPPING_BINDING_PATTERN,
		MAPPING_CONSTRUCTOR,
		FORK_STMT,
		MULTI_RECEIVE_WORKERS,
		MULTI_WAIT_FIELDS,
		MODULE_ENUM_DECLARATION:
		true
	default:
		false
	}
}

func (this *BallerinaParser) isSpecialMethodName(token internal.STToken) bool {
	return (((token.kind == SyntaxKind.MAP_KEYWORD) || (token.kind == SyntaxKind.START_KEYWORD)) || (token.kind == SyntaxKind.JOIN_KEYWORD))
}
