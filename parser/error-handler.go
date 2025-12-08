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
)

type Solution struct {
	Ctx           common.ParserRuleContext
	Action        Action
	TokenText     string
	TokenKind     common.SyntaxKind
	RecoveredNode internal.STNode
	RemovedToken  internal.STToken
	Depth         int
}

type Result struct {
	matches     int
	removeFixes int
	fixes       []*Solution
	solution    *Solution
}

func NewResult(fixes []*Solution, matches int) *Result {
	return &Result{
		fixes:   fixes,
		matches: matches,
	}
}

func (r *Result) peekFix() *Solution {
	if len(r.fixes) == 0 {
		return nil
	}
	return r.fixes[len(r.fixes)-1]
}

func (r *Result) popFix() *Solution {
	if len(r.fixes) == 0 {
		return nil
	}

	sol := r.fixes[len(r.fixes)-1]
	r.fixes = r.fixes[:len(r.fixes)-1]

	if sol.Action == ACTION_REMOVE {
		r.removeFixes--
	}
	return sol
}

func (r *Result) pushFix(sol *Solution) {
	if sol.Action == ACTION_REMOVE {
		r.removeFixes++
	}
	r.fixes = append(r.fixes, sol)
}

func (r *Result) fixesSize() int {
	return len(r.fixes)
}

type AbstractParserErrorHandler struct {
	tokenReader        *TokenReader
	ctxStack           []common.ParserRuleContext
	previousTokenIndex int
	itterCount         int
}

var LOOKAHEAD_LIMIT = 4
var RESOLUTION_ITTER_LIMIT = 7
var COMPLETION_ITTER_LIMIT = 15

func NewSolution(action Action, ctx common.ParserRuleContext, tokenKind common.SyntaxKind, tokenText string) *Solution {
	return NewSolutionWithDepth(action, ctx, tokenKind, tokenText, -1)
}

func NewSolutionWithDepth(action Action, ctx common.ParserRuleContext, tokenKind common.SyntaxKind, tokenText string, depth int) *Solution {
	return &Solution{
		Action:    action,
		Ctx:       ctx,
		TokenText: tokenText,
		TokenKind: tokenKind,
		Depth:     depth,
	}
}

func NewAbstractParserErrorHandlerFromTokenReader(tokenReader *TokenReader) *AbstractParserErrorHandler {
	return &AbstractParserErrorHandler{
		tokenReader:        tokenReader,
		ctxStack:           make([]common.ParserRuleContext, 0),
		previousTokenIndex: -1,
		itterCount:         0,
	}
}

func (this *Solution) ToString() string {
	actionStr := "UNKNOWN"
	switch this.Action {
	case ACTION_INSERT:
		actionStr = "INSERT"
	case ACTION_REMOVE:
		actionStr = "REMOVE"
	case ACTION_KEEP:
		actionStr = "KEEP"
	}
	return actionStr + "'" + this.TokenText + "'"
}

func (this *AbstractParserErrorHandler) Recover(currentCtx common.ParserRuleContext, nextToken internal.STToken, isCompletion bool) *Solution {
	currentTokenIndex := this.tokenReader.GetCurrentTokenIndex()
	if currentTokenIndex == this.previousTokenIndex {
		this.itterCount++
	} else {
		this.itterCount = 0
		this.previousTokenIndex = currentTokenIndex
	}
	var fix *Solution
	if isCompletion && (this.itterCount < COMPLETION_ITTER_LIMIT) {
		fix = this.getCompletion(currentCtx, nextToken)
	} else if this.itterCount < RESOLUTION_ITTER_LIMIT {
		fix = this.getResolution(currentCtx, nextToken)
	}
	if fix != nil {
		this.applyFix(currentCtx, fix)
		return fix
	}
	// Fail safe. This means we can't find a path to recover.
	if isCompletion {
		if this.itterCount == COMPLETION_ITTER_LIMIT {
			panic("fail safe reached")
		}
	} else {
		if this.itterCount == RESOLUTION_ITTER_LIMIT {
			panic("fail safe reached")
		}
	}
	return this.getFailSafeSolution(currentCtx, nextToken)
}

func (this *AbstractParserErrorHandler) getResolution(currentCtx common.ParserRuleContext, nextToken internal.STToken) *Solution {
	bestMatch := this.seekMatchStart(currentCtx)
	this.validateSolution(bestMatch, currentCtx, nextToken)
	var sol *Solution
	if bestMatch.matches > 0 {
		sol = bestMatch.solution
	}
	return sol
}

func (this *AbstractParserErrorHandler) getFailSafeSolution(currentCtx common.ParserRuleContext, nextToken internal.STToken) *Solution {
	sol := NewSolution(ACTION_REMOVE, currentCtx, nextToken.Kind(), nextToken.Text())
	sol.RemovedToken = this.ConsumeInvalidToken()
	return sol
}

func (this *AbstractParserErrorHandler) validateSolution(bestMatch *Result, currentCtx common.ParserRuleContext, nextToken internal.STNode) {
	sol := bestMatch.solution
	if (sol == nil) || (sol.Action == ACTION_REMOVE) {
		return
	}
	if (sol.Action == ACTION_KEEP) && (nextToken.Kind() == common.DOCUMENTATION_STRING) {
		bestMatch.solution = NewSolution(ACTION_REMOVE, currentCtx, common.DOCUMENTATION_STRING, currentCtx.String())
	}
	if (sol.Action != ACTION_INSERT) || (bestMatch.fixesSize() < 2) {
		return
	}
	firstFix := bestMatch.popFix()
	secondFix := bestMatch.peekFix()
	bestMatch.pushFix(firstFix)
	if (secondFix.Action == ACTION_REMOVE) && (secondFix.Depth == 1) {
		bestMatch.solution = secondFix
	}
}

func (this *AbstractParserErrorHandler) getCompletion(context common.ParserRuleContext, nextToken internal.STToken) *Solution {
	tempCtxStack := this.ctxStack
	this.ctxStack = this.getCtxStackSnapshot()
	var sol *Solution
	func() {
		// TODO: check if we panic inside this method
		defer func() {
			if r := recover(); r != nil {
				if false {
					panic("assertion failed")
				}
				sol = this.getResolution(context, nextToken)
			}
		}()
		sol = this.getInsertSolution(context)
	}()

	this.ctxStack = tempCtxStack
	return sol
}

func (this *AbstractParserErrorHandler) ConsumeInvalidToken() internal.STToken {
	return this.tokenReader.Read()
}

func (this *AbstractParserErrorHandler) applyFix(currentCtx common.ParserRuleContext, fix *Solution) {
	if fix.Action == ACTION_REMOVE {
		fix.RemovedToken = this.ConsumeInvalidToken()
		fix.RecoveredNode = this.tokenReader.Peek()
		fix.TokenKind = this.tokenReader.Peek().Kind()
	} else if fix.Action == ACTION_INSERT {
		fix.RecoveredNode = this.handleMissingToken(currentCtx, fix)
	}
}

func (this *AbstractParserErrorHandler) handleMissingToken(currentCtx common.ParserRuleContext, fix *Solution) internal.STNode {
	return internal.CreateMissingTokenWithDiagnosticsFromParserRules(fix.TokenKind, fix.Ctx)
}

func (this *AbstractParserErrorHandler) getCtxStackSnapshot() []common.ParserRuleContext {
	snapshot := make([]common.ParserRuleContext, len(this.ctxStack))
	copy(snapshot, this.ctxStack)
	return snapshot
}

func (this *AbstractParserErrorHandler) seekMatchStart(currentCtx common.ParserRuleContext) *Result {
	tempCtxStack := this.ctxStack
	var bestMatch *Result
	func() {
		defer func() {
			if r := recover(); r != nil {
				if false {
					panic("assertion failed")
				}
				bestMatch = NewResult(make([]*Solution, 0), LOOKAHEAD_LIMIT-1)
				bestMatch.solution = NewSolution(ACTION_REMOVE, currentCtx, common.SyntaxKind(0), currentCtx.String())
			}
		}()
		bestMatch = this.seekMatchInSubTree(currentCtx, 1, 0, true)
	}()
	this.ctxStack = tempCtxStack

	return bestMatch
}

func (this *AbstractParserErrorHandler) seekMatchInSubTree(currentCtx common.ParserRuleContext, lookahead int, currentDepth int, isEntryPoint bool) *Result {
	tempCtxStack := this.ctxStack
	this.ctxStack = this.getCtxStackSnapshot()
	result := this.seekMatch(currentCtx, lookahead, currentDepth, isEntryPoint)
	this.ctxStack = tempCtxStack
	return result
}

func (this *AbstractParserErrorHandler) StartContext(context common.ParserRuleContext) {
	this.ctxStack = append(this.ctxStack, context)
}

func (this *AbstractParserErrorHandler) EndContext() {
	this.ctxStack = this.ctxStack[:len(this.ctxStack)-1]
}

func (this *AbstractParserErrorHandler) SwitchContext(context common.ParserRuleContext) {
	this.ctxStack = this.ctxStack[:len(this.ctxStack)-1]
	this.ctxStack = append(this.ctxStack, context)
}

func (this *AbstractParserErrorHandler) getParentContext() common.ParserRuleContext {
	return this.ctxStack[len(this.ctxStack)-1]
}

func (this *AbstractParserErrorHandler) getGrandParentContext() common.ParserRuleContext {
	parent := this.ctxStack[len(this.ctxStack)-1]
	this.ctxStack = this.ctxStack[:len(this.ctxStack)-1]

	grandParent := this.ctxStack[len(this.ctxStack)-1]

	this.ctxStack = append(this.ctxStack, parent)
	return grandParent
}

func (this *AbstractParserErrorHandler) hasAncestorContext(context common.ParserRuleContext) bool {
	for _, ctx := range this.ctxStack {
		if ctx == context {
			return true
		}
	}
	return false
}

func (this *AbstractParserErrorHandler) getContextStack() []common.ParserRuleContext {
	return this.ctxStack
}

func (this *AbstractParserErrorHandler) seekInAlternativesPaths(lookahead int, currentDepth int, currentMatches int, alternativeRules []common.ParserRuleContext, isEntryPoint bool) *Result {
	results := make([][]*Result, LOOKAHEAD_LIMIT)
	bestMatchIndex := 0

	for _, rule := range alternativeRules {
		tempCtxStack := this.ctxStack
		var result *Result
		shouldContinue := false
		func() {
			defer func() {
				if r := recover(); r != nil {
					if false {
						panic("assertion failed")
					}
					shouldContinue = true
				}
			}()
			result = this.seekMatchInSubTree(rule, lookahead, currentDepth, isEntryPoint)
		}()
		this.ctxStack = tempCtxStack

		if shouldContinue {
			continue
		}

		if this.hasFoundBestAlternative(result) {
			return this.getFinalResult(currentMatches, result)
		}
		similarResults := results[result.matches]
		if similarResults == nil {
			similarResults = make([]*Result, 0)
			results[result.matches] = similarResults
			if bestMatchIndex < result.matches {
				bestMatchIndex = result.matches
			}
		}
		results[result.matches] = append(results[result.matches], result)
	}

	bestMatches := results[bestMatchIndex]
	bestMatch := bestMatches[0]
	for i := 1; i < len(bestMatches); i++ {
		currentMatch := bestMatches[i]
		currentMatchRemoveFixes := currentMatch.removeFixes
		bestMatchRemoveFixes := bestMatch.removeFixes
		if bestMatchRemoveFixes == 0 {
			break
		}
		if currentMatchRemoveFixes == bestMatchRemoveFixes {
			currentSol := bestMatch.peekFix()
			foundSol := currentMatch.peekFix()
			if (currentSol.Action == ACTION_REMOVE) && (foundSol.Action == ACTION_INSERT) {
				bestMatch = currentMatch
			}
		} else if currentMatchRemoveFixes < bestMatchRemoveFixes {
			bestMatch = currentMatch
		}
	}
	return this.getFinalResult(currentMatches, bestMatch)
}

func (this *AbstractParserErrorHandler) hasFoundBestAlternative(result *Result) bool {
	if result.matches < (LOOKAHEAD_LIMIT - 1) {
		return false
	}
	if result.solution == nil {
		return true
	}
	return (result.solution.Action != ACTION_REMOVE)
}

func (this *AbstractParserErrorHandler) getFinalResult(currentMatches int, bestMatch *Result) *Result {
	bestMatch.matches += currentMatches
	return bestMatch
}

func (this *AbstractParserErrorHandler) fixAndContinue(currentCtx common.ParserRuleContext, lookahead int, currentDepth int, matchingRulesCount int, isEntryPoint bool) *Result {
	fixedPathResult := this.fixAndContinueCore(currentCtx, lookahead, currentDepth)
	if isEntryPoint {
		fixedPathResult.solution = fixedPathResult.peekFix()
	} else {
		fixedPathResult.solution = NewSolution(ACTION_KEEP, currentCtx, this.getExpectedTokenKind(currentCtx), currentCtx.String())
	}
	return this.getFinalResult(matchingRulesCount, fixedPathResult)
}

func (this *AbstractParserErrorHandler) fixAndContinueCore(currentCtx common.ParserRuleContext, lookahead int, currentDepth int) *Result {
	deletionResult := this.seekMatchInSubTree(currentCtx, lookahead+1, currentDepth+1, false)
	nextCtx := this.getNextRule(currentCtx, lookahead)
	insertionResult := this.seekMatchInSubTree(nextCtx, lookahead, currentDepth+1, false)
	var fixedPathResult *Result
	var action *Solution

	if (insertionResult.matches == 0) && (deletionResult.matches == 0) {
		action = NewSolutionWithDepth(ACTION_INSERT, currentCtx, this.getExpectedTokenKind(currentCtx), currentCtx.String(), currentDepth)
		insertionResult.pushFix(action)
		fixedPathResult = insertionResult
	} else if insertionResult.matches == deletionResult.matches {
		if insertionResult.removeFixes <= (deletionResult.removeFixes + 1) {
			action = NewSolutionWithDepth(ACTION_INSERT, currentCtx, this.getExpectedTokenKind(currentCtx), currentCtx.String(), currentDepth)
			insertionResult.pushFix(action)
			fixedPathResult = insertionResult
		} else {
			token := this.tokenReader.PeekN(lookahead)
			action = NewSolutionWithDepth(ACTION_REMOVE, currentCtx, token.Kind(), token.Text(), currentDepth)
			deletionResult.pushFix(action)
			fixedPathResult = deletionResult
		}
	} else if insertionResult.matches > deletionResult.matches {
		action = NewSolutionWithDepth(ACTION_INSERT, currentCtx, this.getExpectedTokenKind(currentCtx), currentCtx.String(), currentDepth)
		insertionResult.pushFix(action)
		fixedPathResult = insertionResult
	} else {
		token := this.tokenReader.PeekN(lookahead)
		action = NewSolutionWithDepth(ACTION_REMOVE, currentCtx, token.Kind(), token.Text(), currentDepth)
		deletionResult.pushFix(action)
		fixedPathResult = deletionResult
	}
	return fixedPathResult
}
