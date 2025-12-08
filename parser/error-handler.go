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

import "ballerina-lang-go/parser/common"

type Solution struct {
	Ctx           ParserRuleContext
	Action        Action
	TokenText     string
	TokenKind     common.SyntaxKind
	RecoveredNode Node
	RemovedToken  Token
	Depth         int
}

type AbstractParserErrorHandler struct {
	tokenReader        TokenReader
	ctxStack           []ParserRuleContext
	previousTokenIndex int
	itterCount         int
}

var LOOKAHEAD_LIMIT = 4
var RESOLUTION_ITTER_LIMIT = 7
var COMPLETION_ITTER_LIMIT = 15

func NewSolutionFromActionCtxTokenKindTokenText(action Action, ctx ParserRuleContext, tokenKind common.SyntaxKind, tokenText string) Solution {
	this := Solution{}
	this(action, ctx, tokenKind, tokenText, (-1))
	return this
}

func NewSolutionFromActionCtxTokenKindTokenTextDepth(action Action, ctx ParserRuleContext, tokenKind common.SyntaxKind, tokenText string, depth int) Solution {
	this := Solution{}
	this.action = action
	this.ctx = ctx
	this.tokenText = tokenText
	this.tokenKind = tokenKind
	this.depth = depth
	return this
}

func NewAbstractParserErrorHandlerFromTokenReader(tokenReader TokenReader) AbstractParserErrorHandler {
	this := AbstractParserErrorHandler{}
	this.ctxStack = make([]interface{}, 0)
	// Default field initializations

	this.tokenReader = tokenReader
	this.previousTokenIndex = (-1)
	this.itterCount = 0
	return this
}

func (this *Solution) ToString() string {
	return (((this.action.toString() + "'") + tokenText) + "'")
}

func (this *AbstractParserErrorHandler) hasAlternativePaths(context ParserRuleContext) bool {
}

func (this *AbstractParserErrorHandler) seekMatch(context ParserRuleContext, lookahead int, currentDepth int, isEntryPoint bool) Result {
}

func (this *AbstractParserErrorHandler) getNextRule(context ParserRuleContext, nextLookahead int) ParserRuleContext {
}

func (this *AbstractParserErrorHandler) getExpectedTokenKind(context ParserRuleContext) common.SyntaxKind {
}

func (this *AbstractParserErrorHandler) getInsertSolution(context ParserRuleContext) Solution {
}

func (this *AbstractParserErrorHandler) Recover(currentCtx ParserRuleContext, nextToken Token, isCompletion bool) Solution {
	currentTokenIndex := this.this.tokenReader.getCurrentTokenIndex()
	if currentTokenIndex == this.previousTokenIndex {
		itterCount++
	} else {
		itterCount = 0
		previousTokenIndex = currentTokenIndex
	}
	fix := nil
	if isCompletion && (itterCount < COMPLETION_ITTER_LIMIT) {
		fix = this.getCompletion(currentCtx, nextToken)
	} else if itterCount < RESOLUTION_ITTER_LIMIT {
		fix = this.getResolution(currentCtx, nextToken)
	}
	if fix != nil {
		this.applyFix(currentCtx, fix)
		return fix
	}
	if isCompletion {
		if itterCount == COMPLETION_ITTER_LIMIT {
			panic("assertion failed")
		} else if itterCount == RESOLUTION_ITTER_LIMIT {
			panic("assertion failed")
		}
	}
	return this.getFailSafeSolution(currentCtx, nextToken)
}

func (this *AbstractParserErrorHandler) getResolution(currentCtx ParserRuleContext, nextToken Token) Solution {
	bestMatch := this.seekMatch(currentCtx)
	this.validateSolution(bestMatch, currentCtx, nextToken)
	sol := nil
	if bestMatch.matches > 0 {
		sol = bestMatch.solution
	}
	return sol
}

func (this *AbstractParserErrorHandler) getFailSafeSolution(currentCtx ParserRuleContext, nextToken Token) Solution {
	sol := nil
	sol.removedToken = this.consumeInvalidToken()
	return sol
}

func (this *AbstractParserErrorHandler) validateSolution(bestMatch Result, currentCtx ParserRuleContext, nextToken Node) {
	sol := bestMatch.solution
	if (sol == nil) || (sol.action == Action_REMOVE) {
		return
	}
	if (sol.action == Action_KEEP) && (nextToken.kind == SyntaxKind_DOCUMENTATION_STRING) {
		bestMatch.solution = nil
	}
	if (sol.action != Action_INSERT) || (this.bestMatch.fixesSize() < 2) {
		return
	}
	firstFix := this.bestMatch.popFix()
	secondFix := this.bestMatch.peekFix()
	this.bestMatch.pushFix(firstFix)
	if (secondFix.action == Action_REMOVE) && (secondFix.depth == 1) {
		bestMatch.solution = secondFix
	}
}

func (this *AbstractParserErrorHandler) getCompletion(context ParserRuleContext, nextToken Token) Solution {
	tempCtxStack := this.ctxStack
	this.ctxStack = this.getCtxStackSnapshot()
	var sol Solution
	func() {
		defer func() {
			if r := recover(); r != nil {
				if _, ok := r.(IllegalStateException); ok {
					if false {
						panic("assertion failed")
					}
					sol = this.getResolution(context, nextToken)
				} else {
					panic(r) // re-panic if it's not a handled exception
				}
			}
		}()
		sol = this.getInsertSolution(context)
	}()

	this.ctxStack = tempCtxStack
	return sol
}

func (this *AbstractParserErrorHandler) ConsumeInvalidToken() Token {
	return this.this.tokenReader.read()
}

func (this *AbstractParserErrorHandler) applyFix(currentCtx ParserRuleContext, fix Solution) {
	if fix.action == Action_REMOVE {
		fix.removedToken = this.consumeInvalidToken()
		fix.recoveredNode = this.this.tokenReader.peek()
		fix.tokenKind = this.tokenReader.peek().kind
	} else if fix.action == Action_INSERT {
		fix.recoveredNode = this.handleMissingToken(currentCtx, fix)
	}
}

func (this *AbstractParserErrorHandler) handleMissingToken(currentCtx ParserRuleContext, fix Solution) Node {
	return this.SyntaxErrors.createMissingTokenWithDiagnostics(fix.tokenKind, fix.ctx)
}

func (this *AbstractParserErrorHandler) getCtxStackSnapshot() []ParserRuleContext {
	return this.this.ctxStack.clone()
}

func (this *AbstractParserErrorHandler) seekMatch(currentCtx ParserRuleContext) Result {
	tempCtxStack := this.ctxStack
	var bestMatch Result
	func() {
		defer func() {
			if r := recover(); r != nil {
				if _, ok := r.(IllegalStateException); ok {
					if false {
						panic("assertion failed")
					}
					bestMatch = nil
					bestMatch.solution = nil
				} else {
					panic(r) // re-panic if it's not a handled exception
				}
			}
		}()
		bestMatch = this.seekMatchInSubTree(currentCtx, 1, 0, true)
	}()
	this.ctxStack = tempCtxStack

	return bestMatch
}

func (this *AbstractParserErrorHandler) seekMatchInSubTree(currentCtx ParserRuleContext, lookahead int, currentDepth int, isEntryPoint bool) Result {
	tempCtxStack := this.ctxStack
	this.ctxStack = this.getCtxStackSnapshot()
	result := this.seekMatch(currentCtx, lookahead, currentDepth, isEntryPoint)
	this.ctxStack = tempCtxStack
	return result
}

func (this *AbstractParserErrorHandler) StartContext(context ParserRuleContext) {
	this.this.ctxStack.push(context)
}

func (this *AbstractParserErrorHandler) EndContext() {
	this.this.ctxStack.pop()
}

func (this *AbstractParserErrorHandler) SwitchContext(context ParserRuleContext) {
	this.this.ctxStack.pop()
	this.this.ctxStack.push(context)
}

func (this *AbstractParserErrorHandler) getParentContext() ParserRuleContext {
	return this.this.ctxStack.peek()
}

func (this *AbstractParserErrorHandler) getGrandParentContext() ParserRuleContext {
	parent := this.this.ctxStack.pop()
	grandParent := this.this.ctxStack.peek()
	this.this.ctxStack.push(parent)
	return grandParent
}

func (this *AbstractParserErrorHandler) hasAncestorContext(context ParserRuleContext) bool {
	return this.this.ctxStack.contains(context)
}

func (this *AbstractParserErrorHandler) getContextStack() []ParserRuleContext {
	return this.ctxStack
}

func (this *AbstractParserErrorHandler) seekInAlternativesPaths(lookahead int, currentDepth int, currentMatches int, alternativeRules []ParserRuleContext, isEntryPoint bool) Result {
	results := nil
	bestMatchIndex := 0
	for _, rule := range alternativeRules {
		tempCtxStack := this.ctxStack
		var result Result
		func() {
			defer func() {
				if r := recover(); r != nil {
					if _, ok := r.(IllegalStateException); ok {
						if false {
							panic("assertion failed")
						}
						continue
					} else {
						panic(r) // re-panic if it's not a handled exception
					}
				}
			}()
			result = this.seekMatchInSubTree(rule, lookahead, currentDepth, isEntryPoint)
		}()
		this.ctxStack = tempCtxStack

		if this.hasFoundBestAlternative(result) {
			return this.getFinalResult(currentMatches, result)
		}
		similarResutls := results[result.matches]
		if similarResutls == nil {
			similarResutls = make([]interface{}, 0)
			results[result.matches] = similarResutls
			if bestMatchIndex < result.matches {
				bestMatchIndex = result.matches
			}
		}
		this.similarResutls.add(result)
	}
	bestMatches := results[bestMatchIndex]
	bestMatch := this.bestMatches.get(0)
	var currentMatch Result
	i := 1
	for ; i < len(bestMatches); i++ {
		currentMatch = this.bestMatches.get(i)
		currentMatchRemoveFixes := currentMatch.removeFixes
		bestMatchRemoveFixes := bestMatch.removeFixes
		if bestMatchRemoveFixes == 0 {
			break
		}
		if currentMatchRemoveFixes == bestMatchRemoveFixes {
			currentSol := this.bestMatch.peekFix()
			foundSol := this.currentMatch.peekFix()
			if (currentSol.action == Action_REMOVE) && (foundSol.action == Action_INSERT) {
				bestMatch = currentMatch
			}
		} else if currentMatchRemoveFixes < bestMatchRemoveFixes {
			bestMatch = currentMatch
		}
	}
	return this.getFinalResult(currentMatches, bestMatch)
}

func (this *AbstractParserErrorHandler) hasFoundBestAlternative(result Result) bool {
	if result.matches < (LOOKAHEAD_LIMIT - 1) {
		return false
	}
	if result.solution == nil {
		return true
	}
	return (result.solution.action != Action_REMOVE)
}

func (this *AbstractParserErrorHandler) getFinalResult(currentMatches int, bestMatch Result) Result {
	bestMatch.matches = currentMatches
	return bestMatch
}

func (this *AbstractParserErrorHandler) fixAndContinue(currentCtx ParserRuleContext, lookahead int, currentDepth int, matchingRulesCount int, isEntryPoint bool) Result {
	fixedPathResult := this.fixAndContinue(currentCtx, lookahead, currentDepth)
	if isEntryPoint {
		fixedPathResult.solution = this.fixedPathResult.peekFix()
	} else {
		fixedPathResult.solution = nil
	}
	return this.getFinalResult(matchingRulesCount, fixedPathResult)
}

func (this *AbstractParserErrorHandler) fixAndContinue(currentCtx ParserRuleContext, lookahead int, currentDepth int) Result {
	deletionResult := this.seekMatchInSubTree(currentCtx, lookahead+1, currentDepth+1, false)
	nextCtx := this.getNextRule(currentCtx, lookahead)
	insertionResult := this.seekMatchInSubTree(nextCtx, lookahead, currentDepth+1, false)
	var fixedPathResult Result
	var action Solution
	if (insertionResult.matches == 0) && (deletionResult.matches == 0) {
		action = nil
		this.insertionResult.pushFix(action)
		fixedPathResult = insertionResult
	} else if insertionResult.matches == deletionResult.matches {
		if insertionResult.removeFixes <= (deletionResult.removeFixes + 1) {
			action = nil
			this.insertionResult.pushFix(action)
			fixedPathResult = insertionResult
		} else {
			token := this.this.tokenReader.peek(lookahead)
			action = nil
			this.deletionResult.pushFix(action)
			fixedPathResult = deletionResult
		}
	} else if insertionResult.matches > deletionResult.matches {
		action = nil
		this.insertionResult.pushFix(action)
		fixedPathResult = insertionResult
	} else {
		token := this.this.tokenReader.peek(lookahead)
		action = nil
		this.deletionResult.pushFix(action)
		fixedPathResult = deletionResult
	}
	return fixedPathResult
}
