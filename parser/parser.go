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
