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

// ============================================================================
// Solution struct - represents a fix for a parser error
// ============================================================================

type Solution struct {
	Ctx           common.ParserRuleContext
	Action        Action
	TokenText     string
	TokenKind     common.SyntaxKind
	RecoveredNode internal.STNode
	RemovedToken  internal.STToken
	Depth         int
}

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

// ============================================================================
// Result struct - holds results of error recovery attempts
// ============================================================================

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

// ============================================================================
// Constants
// ============================================================================

var LOOKAHEAD_LIMIT = 4
var RESOLUTION_ITTER_LIMIT = 7
var COMPLETION_ITTER_LIMIT = 15

// ============================================================================
// AbstractParserErrorHandlerData - Field access interface
// ============================================================================

type AbstractParserErrorHandlerData interface {
	GetTokenReader() *TokenReader
	SetTokenReader(*TokenReader)
	GetCtxStack() []common.ParserRuleContext
	SetCtxStack([]common.ParserRuleContext)
	GetPreviousTokenIndex() int
	SetPreviousTokenIndex(int)
	GetItterCount() int
	SetItterCount(int)
}

// ============================================================================
// AbstractParserErrorHandlerBase - Base struct with fields
// ============================================================================

type AbstractParserErrorHandlerBase struct {
	tokenReader        *TokenReader
	ctxStack           []common.ParserRuleContext
	previousTokenIndex int
	itterCount         int
}

func NewAbstractParserErrorHandlerBase(tokenReader *TokenReader) *AbstractParserErrorHandlerBase {
	return &AbstractParserErrorHandlerBase{
		tokenReader:        tokenReader,
		ctxStack:           make([]common.ParserRuleContext, 0),
		previousTokenIndex: -1,
		itterCount:         0,
	}
}

// Getter/setter implementations for AbstractParserErrorHandlerBase

func (b *AbstractParserErrorHandlerBase) GetTokenReader() *TokenReader {
	return b.tokenReader
}

func (b *AbstractParserErrorHandlerBase) SetTokenReader(tokenReader *TokenReader) {
	b.tokenReader = tokenReader
}

func (b *AbstractParserErrorHandlerBase) GetCtxStack() []common.ParserRuleContext {
	return b.ctxStack
}

func (b *AbstractParserErrorHandlerBase) SetCtxStack(ctxStack []common.ParserRuleContext) {
	b.ctxStack = ctxStack
}

func (b *AbstractParserErrorHandlerBase) GetPreviousTokenIndex() int {
	return b.previousTokenIndex
}

func (b *AbstractParserErrorHandlerBase) SetPreviousTokenIndex(previousTokenIndex int) {
	b.previousTokenIndex = previousTokenIndex
}

func (b *AbstractParserErrorHandlerBase) GetItterCount() int {
	return b.itterCount
}

func (b *AbstractParserErrorHandlerBase) SetItterCount(itterCount int) {
	b.itterCount = itterCount
}

// ============================================================================
// AbstractParserErrorHandler - Main interface
// ============================================================================

type AbstractParserErrorHandler interface {
	AbstractParserErrorHandlerData

	// Abstract methods (to be implemented by concrete classes like BallerinaParserErrorHandler)
	HasAlternativePaths(context common.ParserRuleContext) bool
	SeekMatch(context common.ParserRuleContext, lookahead int, currentDepth int, isEntryPoint bool) *Result
	GetNextRule(context common.ParserRuleContext, nextLookahead int) common.ParserRuleContext
	GetExpectedTokenKind(context common.ParserRuleContext) common.SyntaxKind
	GetInsertSolution(context common.ParserRuleContext) *Solution

	// Default/concrete methods (implemented in AbstractParserErrorHandlerMethods)
	Recover(currentCtx common.ParserRuleContext, nextToken internal.STToken, isCompletion bool) *Solution
	ConsumeInvalidToken() internal.STToken
	StartContext(context common.ParserRuleContext)
	EndContext()
	SwitchContext(context common.ParserRuleContext)
	GetParentContext() common.ParserRuleContext
	GetGrandParentContext() common.ParserRuleContext
	HasAncestorContext(context common.ParserRuleContext) bool
	GetContextStack() []common.ParserRuleContext
}

// ============================================================================
// AbstractParserErrorHandlerMethods - Default method implementations
// ============================================================================

type AbstractParserErrorHandlerMethods struct {
	Self AbstractParserErrorHandler
}

func (m *AbstractParserErrorHandlerMethods) Recover(currentCtx common.ParserRuleContext, nextToken internal.STToken, isCompletion bool) *Solution {
	currentTokenIndex := m.Self.GetTokenReader().GetCurrentTokenIndex()
	if currentTokenIndex == m.Self.GetPreviousTokenIndex() {
		m.Self.SetItterCount(m.Self.GetItterCount() + 1)
	} else {
		m.Self.SetItterCount(0)
		m.Self.SetPreviousTokenIndex(currentTokenIndex)
	}
	var fix *Solution
	if isCompletion && (m.Self.GetItterCount() < COMPLETION_ITTER_LIMIT) {
		fix = m.getCompletion(currentCtx, nextToken)
	} else if m.Self.GetItterCount() < RESOLUTION_ITTER_LIMIT {
		fix = m.getResolution(currentCtx, nextToken)
	}
	if fix != nil {
		m.applyFix(currentCtx, fix)
		return fix
	}
	// Fail safe. This means we can't find a path to recover.
	if isCompletion {
		if m.Self.GetItterCount() == COMPLETION_ITTER_LIMIT {
			panic("fail safe reached")
		}
	} else {
		if m.Self.GetItterCount() == RESOLUTION_ITTER_LIMIT {
			panic("fail safe reached")
		}
	}
	return m.getFailSafeSolution(currentCtx, nextToken)
}

func (m *AbstractParserErrorHandlerMethods) getResolution(currentCtx common.ParserRuleContext, nextToken internal.STToken) *Solution {
	bestMatch := m.seekMatchStart(currentCtx)
	m.validateSolution(bestMatch, currentCtx, nextToken)
	var sol *Solution
	if bestMatch.matches > 0 {
		sol = bestMatch.solution
	}
	return sol
}

func (m *AbstractParserErrorHandlerMethods) getFailSafeSolution(currentCtx common.ParserRuleContext, nextToken internal.STToken) *Solution {
	sol := NewSolution(ACTION_REMOVE, currentCtx, nextToken.Kind(), nextToken.Text())
	sol.RemovedToken = m.Self.ConsumeInvalidToken()
	return sol
}

func (m *AbstractParserErrorHandlerMethods) validateSolution(bestMatch *Result, currentCtx common.ParserRuleContext, nextToken internal.STNode) {
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

func (m *AbstractParserErrorHandlerMethods) getCompletion(context common.ParserRuleContext, nextToken internal.STToken) *Solution {
	tempCtxStack := m.Self.GetCtxStack()
	m.Self.SetCtxStack(m.getCtxStackSnapshot())
	var sol *Solution
	func() {
		// TODO: check if we panic inside this method
		defer func() {
			if r := recover(); r != nil {
				if false {
					panic("assertion failed")
				}
				sol = m.getResolution(context, nextToken)
			}
		}()
		sol = m.Self.GetInsertSolution(context)
	}()

	m.Self.SetCtxStack(tempCtxStack)
	return sol
}

func (m *AbstractParserErrorHandlerMethods) ConsumeInvalidToken() internal.STToken {
	return m.Self.GetTokenReader().Read()
}

func (m *AbstractParserErrorHandlerMethods) applyFix(currentCtx common.ParserRuleContext, fix *Solution) {
	if fix.Action == ACTION_REMOVE {
		fix.RemovedToken = m.Self.ConsumeInvalidToken()
		fix.RecoveredNode = m.Self.GetTokenReader().Peek()
		fix.TokenKind = m.Self.GetTokenReader().Peek().Kind()
	} else if fix.Action == ACTION_INSERT {
		fix.RecoveredNode = m.handleMissingToken(currentCtx, fix)
	}
}

func (m *AbstractParserErrorHandlerMethods) handleMissingToken(currentCtx common.ParserRuleContext, fix *Solution) internal.STNode {
	return internal.CreateMissingTokenWithDiagnosticsFromParserRules(fix.TokenKind, fix.Ctx)
}

func (m *AbstractParserErrorHandlerMethods) getCtxStackSnapshot() []common.ParserRuleContext {
	ctxStack := m.Self.GetCtxStack()
	snapshot := make([]common.ParserRuleContext, len(ctxStack))
	copy(snapshot, ctxStack)
	return snapshot
}

func (m *AbstractParserErrorHandlerMethods) seekMatchStart(currentCtx common.ParserRuleContext) *Result {
	tempCtxStack := m.Self.GetCtxStack()
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
		bestMatch = m.seekMatchInSubTree(currentCtx, 1, 0, true)
	}()
	m.Self.SetCtxStack(tempCtxStack)

	return bestMatch
}

func (m *AbstractParserErrorHandlerMethods) seekMatchInSubTree(currentCtx common.ParserRuleContext, lookahead int, currentDepth int, isEntryPoint bool) *Result {
	tempCtxStack := m.Self.GetCtxStack()
	m.Self.SetCtxStack(m.getCtxStackSnapshot())
	result := m.Self.SeekMatch(currentCtx, lookahead, currentDepth, isEntryPoint)
	m.Self.SetCtxStack(tempCtxStack)
	return result
}

func (m *AbstractParserErrorHandlerMethods) StartContext(context common.ParserRuleContext) {
	ctxStack := m.Self.GetCtxStack()
	m.Self.SetCtxStack(append(ctxStack, context))
}

func (m *AbstractParserErrorHandlerMethods) EndContext() {
	ctxStack := m.Self.GetCtxStack()
	m.Self.SetCtxStack(ctxStack[:len(ctxStack)-1])
}

func (m *AbstractParserErrorHandlerMethods) SwitchContext(context common.ParserRuleContext) {
	ctxStack := m.Self.GetCtxStack()
	ctxStack = ctxStack[:len(ctxStack)-1]
	m.Self.SetCtxStack(append(ctxStack, context))
}

func (m *AbstractParserErrorHandlerMethods) GetParentContext() common.ParserRuleContext {
	ctxStack := m.Self.GetCtxStack()
	return ctxStack[len(ctxStack)-1]
}

func (m *AbstractParserErrorHandlerMethods) GetGrandParentContext() common.ParserRuleContext {
	ctxStack := m.Self.GetCtxStack()
	parent := ctxStack[len(ctxStack)-1]
	ctxStack = ctxStack[:len(ctxStack)-1]

	grandParent := ctxStack[len(ctxStack)-1]

	m.Self.SetCtxStack(append(ctxStack, parent))
	return grandParent
}

func (m *AbstractParserErrorHandlerMethods) HasAncestorContext(context common.ParserRuleContext) bool {
	ctxStack := m.Self.GetCtxStack()
	for _, ctx := range ctxStack {
		if ctx == context {
			return true
		}
	}
	return false
}

func (m *AbstractParserErrorHandlerMethods) GetContextStack() []common.ParserRuleContext {
	return m.Self.GetCtxStack()
}

func (m *AbstractParserErrorHandlerMethods) seekInAlternativesPaths(lookahead int, currentDepth int, currentMatches int, alternativeRules []common.ParserRuleContext, isEntryPoint bool) *Result {
	results := make([][]*Result, LOOKAHEAD_LIMIT)
	bestMatchIndex := 0

	for _, rule := range alternativeRules {
		tempCtxStack := m.Self.GetCtxStack()
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
			result = m.seekMatchInSubTree(rule, lookahead, currentDepth, isEntryPoint)
		}()
		m.Self.SetCtxStack(tempCtxStack)

		if shouldContinue {
			continue
		}

		if m.hasFoundBestAlternative(result) {
			return m.getFinalResult(currentMatches, result)
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
	return m.getFinalResult(currentMatches, bestMatch)
}

func (m *AbstractParserErrorHandlerMethods) hasFoundBestAlternative(result *Result) bool {
	if result.matches < (LOOKAHEAD_LIMIT - 1) {
		return false
	}
	if result.solution == nil {
		return true
	}
	return (result.solution.Action != ACTION_REMOVE)
}

func (m *AbstractParserErrorHandlerMethods) getFinalResult(currentMatches int, bestMatch *Result) *Result {
	bestMatch.matches += currentMatches
	return bestMatch
}

func (m *AbstractParserErrorHandlerMethods) fixAndContinue(currentCtx common.ParserRuleContext, lookahead int, currentDepth int, matchingRulesCount int, isEntryPoint bool) *Result {
	fixedPathResult := m.fixAndContinueCore(currentCtx, lookahead, currentDepth)
	if isEntryPoint {
		fixedPathResult.solution = fixedPathResult.peekFix()
	} else {
		fixedPathResult.solution = NewSolution(ACTION_KEEP, currentCtx, m.Self.GetExpectedTokenKind(currentCtx), currentCtx.String())
	}
	return m.getFinalResult(matchingRulesCount, fixedPathResult)
}

func (m *AbstractParserErrorHandlerMethods) fixAndContinueCore(currentCtx common.ParserRuleContext, lookahead int, currentDepth int) *Result {
	deletionResult := m.seekMatchInSubTree(currentCtx, lookahead+1, currentDepth+1, false)
	nextCtx := m.Self.GetNextRule(currentCtx, lookahead)
	insertionResult := m.seekMatchInSubTree(nextCtx, lookahead, currentDepth+1, false)
	var fixedPathResult *Result
	var action *Solution

	if (insertionResult.matches == 0) && (deletionResult.matches == 0) {
		action = NewSolutionWithDepth(ACTION_INSERT, currentCtx, m.Self.GetExpectedTokenKind(currentCtx), currentCtx.String(), currentDepth)
		insertionResult.pushFix(action)
		fixedPathResult = insertionResult
	} else if insertionResult.matches == deletionResult.matches {
		if insertionResult.removeFixes <= (deletionResult.removeFixes + 1) {
			action = NewSolutionWithDepth(ACTION_INSERT, currentCtx, m.Self.GetExpectedTokenKind(currentCtx), currentCtx.String(), currentDepth)
			insertionResult.pushFix(action)
			fixedPathResult = insertionResult
		} else {
			token := m.Self.GetTokenReader().PeekN(lookahead)
			action = NewSolutionWithDepth(ACTION_REMOVE, currentCtx, token.Kind(), token.Text(), currentDepth)
			deletionResult.pushFix(action)
			fixedPathResult = deletionResult
		}
	} else if insertionResult.matches > deletionResult.matches {
		action = NewSolutionWithDepth(ACTION_INSERT, currentCtx, m.Self.GetExpectedTokenKind(currentCtx), currentCtx.String(), currentDepth)
		insertionResult.pushFix(action)
		fixedPathResult = insertionResult
	} else {
		token := m.Self.GetTokenReader().PeekN(lookahead)
		action = NewSolutionWithDepth(ACTION_REMOVE, currentCtx, token.Kind(), token.Text(), currentDepth)
		deletionResult.pushFix(action)
		fixedPathResult = deletionResult
	}
	return fixedPathResult
}

type BallerinaParserErrorHandler struct {
	ParserErrorHandler
}

var FUNC_TYPE_OR_DEF_OPTIONAL_RETURNS = []ParserRuleContext{ParserRuleContext_RETURNS_KEYWORD, ParserRuleContext_FUNC_BODY_OR_TYPE_DESC_RHS}
var FUNC_BODY_OR_TYPE_DESC_RHS = []ParserRuleContext{ParserRuleContext_FUNC_BODY, ParserRuleContext_MODULE_LEVEL_AMBIGUOUS_FUNC_TYPE_DESC_RHS}
var FUNC_DEF_OPTIONAL_RETURNS = []ParserRuleContext{ParserRuleContext_RETURNS_KEYWORD, ParserRuleContext_FUNC_BODY}
var METHOD_DECL_OPTIONAL_RETURNS = []ParserRuleContext{ParserRuleContext_RETURNS_KEYWORD, ParserRuleContext_SEMICOLON}
var FUNC_BODY = []ParserRuleContext{ParserRuleContext_FUNC_BODY_BLOCK, ParserRuleContext_EXTERNAL_FUNC_BODY}
var EXTERNAL_FUNC_BODY_OPTIONAL_ANNOTS = []ParserRuleContext{ParserRuleContext_ANNOTATIONS, ParserRuleContext_EXTERNAL_KEYWORD}
var ANNON_FUNC_OPTIONAL_RETURNS = []ParserRuleContext{ParserRuleContext_RETURNS_KEYWORD, ParserRuleContext_ANON_FUNC_BODY}
var ANON_FUNC_BODY = []ParserRuleContext{ParserRuleContext_FUNC_BODY_BLOCK, ParserRuleContext_EXPLICIT_ANON_FUNC_EXPR_BODY_START}
var FUNC_TYPE_OPTIONAL_RETURNS = []ParserRuleContext{ParserRuleContext_RETURNS_KEYWORD, ParserRuleContext_FUNC_TYPE_DESC_END}
var FUNC_TYPE_OR_ANON_FUNC_OPTIONAL_RETURNS = []ParserRuleContext{ParserRuleContext_RETURNS_KEYWORD, ParserRuleContext_FUNC_TYPE_DESC_RHS_OR_ANON_FUNC_BODY}
var FUNC_TYPE_DESC_RHS_OR_ANON_FUNC_BODY = []ParserRuleContext{ParserRuleContext_ANON_FUNC_BODY, ParserRuleContext_STMT_LEVEL_AMBIGUOUS_FUNC_TYPE_DESC_RHS}
var WORKER_NAME_RHS = []ParserRuleContext{ParserRuleContext_RETURNS_KEYWORD, ParserRuleContext_BLOCK_STMT}
var STATEMENTS = []ParserRuleContext{ParserRuleContext_CLOSE_BRACE, ParserRuleContext_ASSIGNMENT_STMT, ParserRuleContext_VAR_DECL_STMT, ParserRuleContext_IF_BLOCK, ParserRuleContext_WHILE_BLOCK, ParserRuleContext_CALL_STMT, ParserRuleContext_PANIC_STMT, ParserRuleContext_CONTINUE_STATEMENT, ParserRuleContext_BREAK_STATEMENT, ParserRuleContext_RETURN_STMT, ParserRuleContext_MATCH_STMT, ParserRuleContext_EXPRESSION_STATEMENT, ParserRuleContext_LOCK_STMT, ParserRuleContext_NAMED_WORKER_DECL, ParserRuleContext_FORK_STMT, ParserRuleContext_FOREACH_STMT, ParserRuleContext_XML_NAMESPACE_DECLARATION, ParserRuleContext_TRANSACTION_STMT, ParserRuleContext_RETRY_STMT, ParserRuleContext_ROLLBACK_STMT, ParserRuleContext_DO_BLOCK, ParserRuleContext_FAIL_STATEMENT, ParserRuleContext_BLOCK_STMT}
var ASSIGNMENT_STMT_RHS = []ParserRuleContext{ParserRuleContext_ASSIGN_OP, ParserRuleContext_COMPOUND_BINARY_OPERATOR}
var VAR_DECL_RHS = []ParserRuleContext{ParserRuleContext_ASSIGN_OP, ParserRuleContext_SEMICOLON}
var TOP_LEVEL_NODE = []ParserRuleContext{ParserRuleContext_EOF, ParserRuleContext_TOP_LEVEL_NODE_WITHOUT_METADATA, ParserRuleContext_DOC_STRING, ParserRuleContext_ANNOTATIONS}
var TOP_LEVEL_NODE_WITHOUT_METADATA = []ParserRuleContext{ParserRuleContext_EOF, ParserRuleContext_TOP_LEVEL_NODE_WITHOUT_MODIFIER, ParserRuleContext_PUBLIC_KEYWORD}
var TOP_LEVEL_NODE_WITHOUT_MODIFIER = []ParserRuleContext{ParserRuleContext_EOF, ParserRuleContext_FUNC_DEF, ParserRuleContext_MODULE_VAR_DECL, ParserRuleContext_MODULE_CLASS_DEFINITION, ParserRuleContext_SERVICE_DECL, ParserRuleContext_LISTENER_DECL, ParserRuleContext_MODULE_TYPE_DEFINITION, ParserRuleContext_CONSTANT_DECL, ParserRuleContext_ANNOTATION_DECL, ParserRuleContext_XML_NAMESPACE_DECLARATION, ParserRuleContext_MODULE_ENUM_DECLARATION, ParserRuleContext_IMPORT_DECL}
var FUNC_DEF_START = []ParserRuleContext{ParserRuleContext_FUNCTION_KEYWORD, ParserRuleContext_FUNC_DEF_FIRST_QUALIFIER}
var FUNC_DEF_WITHOUT_FIRST_QUALIFIER = []ParserRuleContext{ParserRuleContext_FUNCTION_KEYWORD, ParserRuleContext_FUNC_DEF_SECOND_QUALIFIER}
var TYPE_OR_VAR_NAME = []ParserRuleContext{ParserRuleContext_VARIABLE_NAME, ParserRuleContext_TYPE_DESC_IN_TYPE_BINDING_PATTERN}
var FIELD_DESCRIPTOR_RHS = []ParserRuleContext{ParserRuleContext_SEMICOLON, ParserRuleContext_QUESTION_MARK, ParserRuleContext_ASSIGN_OP}
var FIELD_OR_REST_DESCIPTOR_RHS = []ParserRuleContext{ParserRuleContext_VARIABLE_NAME, ParserRuleContext_ELLIPSIS}
var RECORD_BODY_START = []ParserRuleContext{ParserRuleContext_CLOSED_RECORD_BODY_START, ParserRuleContext_OPEN_BRACE}
var RECORD_BODY_END = []ParserRuleContext{ParserRuleContext_CLOSED_RECORD_BODY_END, ParserRuleContext_CLOSE_BRACE}
var TYPE_DESCRIPTORS = []ParserRuleContext{ParserRuleContext_SIMPLE_TYPE_DESC_IDENTIFIER, ParserRuleContext_TYPE_REFERENCE, ParserRuleContext_SIMPLE_TYPE_DESCRIPTOR, ParserRuleContext_OBJECT_TYPE_DESCRIPTOR, ParserRuleContext_RECORD_TYPE_DESCRIPTOR, ParserRuleContext_MAP_TYPE_DESCRIPTOR, ParserRuleContext_PARAMETERIZED_TYPE, ParserRuleContext_TUPLE_TYPE_DESC_START, ParserRuleContext_STREAM_KEYWORD, ParserRuleContext_TABLE_KEYWORD, ParserRuleContext_FUNC_TYPE_DESC, ParserRuleContext_CONSTANT_EXPRESSION, ParserRuleContext_PARENTHESISED_TYPE_DESC_START}
var TYPE_DESCRIPTOR_WITHOUT_ISOLATED = []ParserRuleContext{ParserRuleContext_FUNC_TYPE_DESC, ParserRuleContext_OBJECT_TYPE_DESCRIPTOR}
var CLASS_DESCRIPTOR = []ParserRuleContext{ParserRuleContext_TYPE_REFERENCE, ParserRuleContext_STREAM_KEYWORD}
var RECORD_FIELD_OR_RECORD_END = []ParserRuleContext{ParserRuleContext_RECORD_BODY_END, ParserRuleContext_RECORD_FIELD}
var RECORD_FIELD_START = []ParserRuleContext{ParserRuleContext_TYPE_DESC_IN_RECORD_FIELD, ParserRuleContext_ASTERISK, ParserRuleContext_ANNOTATIONS}
var RECORD_FIELD_WITHOUT_METADATA = []ParserRuleContext{ParserRuleContext_ASTERISK, ParserRuleContext_TYPE_DESC_IN_RECORD_FIELD}
var ARG_START_OR_ARG_LIST_END = []ParserRuleContext{ParserRuleContext_ARG_LIST_END, ParserRuleContext_ARG_START}
var ARG_START = []ParserRuleContext{ParserRuleContext_VARIABLE_NAME, ParserRuleContext_ELLIPSIS, ParserRuleContext_EXPRESSION}
var ARG_END = []ParserRuleContext{ParserRuleContext_ARG_LIST_END, ParserRuleContext_COMMA}
var NAMED_OR_POSITIONAL_ARG_RHS = []ParserRuleContext{ParserRuleContext_ARG_END, ParserRuleContext_ASSIGN_OP}
var OPTIONAL_FIELD_INITIALIZER = []ParserRuleContext{ParserRuleContext_ASSIGN_OP, ParserRuleContext_SEMICOLON}
var ON_FAIL_OPTIONAL_BINDING_PATTERN = []ParserRuleContext{ParserRuleContext_BLOCK_STMT, ParserRuleContext_TYPE_DESC_IN_TYPE_BINDING_PATTERN}
var GROUPING_KEY_LIST_ELEMENT = []ParserRuleContext{ParserRuleContext_VARIABLE_NAME, ParserRuleContext_TYPE_DESC_BEFORE_IDENTIFIER_IN_GROUPING_KEY}
var GROUPING_KEY_LIST_ELEMENT_END = []ParserRuleContext{ParserRuleContext_GROUP_BY_CLAUSE_END, ParserRuleContext_COMMA}
var CLASS_MEMBER_OR_OBJECT_MEMBER_START = []ParserRuleContext{ParserRuleContext_ASTERISK, ParserRuleContext_OBJECT_FUNC_OR_FIELD, ParserRuleContext_CLOSE_BRACE, ParserRuleContext_DOC_STRING, ParserRuleContext_ANNOTATIONS}
var OBJECT_CONSTRUCTOR_MEMBER_START = []ParserRuleContext{ParserRuleContext_OBJECT_FUNC_OR_FIELD, ParserRuleContext_CLOSE_BRACE, ParserRuleContext_DOC_STRING, ParserRuleContext_ANNOTATIONS}
var CLASS_MEMBER_OR_OBJECT_MEMBER_WITHOUT_META = []ParserRuleContext{ParserRuleContext_OBJECT_FUNC_OR_FIELD, ParserRuleContext_ASTERISK, ParserRuleContext_CLOSE_BRACE}
var OBJECT_CONS_MEMBER_WITHOUT_META = []ParserRuleContext{ParserRuleContext_OBJECT_FUNC_OR_FIELD, ParserRuleContext_CLOSE_BRACE}
var OBJECT_FUNC_OR_FIELD = []ParserRuleContext{ParserRuleContext_OBJECT_FUNC_OR_FIELD_WITHOUT_VISIBILITY, ParserRuleContext_OBJECT_MEMBER_VISIBILITY_QUAL}
var OBJECT_FUNC_OR_FIELD_WITHOUT_VISIBILITY = []ParserRuleContext{ParserRuleContext_OBJECT_FIELD_START, ParserRuleContext_OBJECT_METHOD_START}
var OBJECT_FIELD_QUALIFIER = []ParserRuleContext{ParserRuleContext_TYPE_DESC_BEFORE_IDENTIFIER, ParserRuleContext_FINAL_KEYWORD}
var OBJECT_METHOD_START = []ParserRuleContext{ParserRuleContext_FUNC_DEF, ParserRuleContext_OBJECT_METHOD_FIRST_QUALIFIER}
var OBJECT_METHOD_WITHOUT_FIRST_QUALIFIER = []ParserRuleContext{ParserRuleContext_FUNC_DEF, ParserRuleContext_OBJECT_METHOD_SECOND_QUALIFIER}
var OBJECT_METHOD_WITHOUT_SECOND_QUALIFIER = []ParserRuleContext{ParserRuleContext_FUNC_DEF, ParserRuleContext_OBJECT_METHOD_THIRD_QUALIFIER}
var OBJECT_METHOD_WITHOUT_THIRD_QUALIFIER = []ParserRuleContext{ParserRuleContext_FUNC_DEF, ParserRuleContext_OBJECT_METHOD_FOURTH_QUALIFIER}
var OBJECT_TYPE_START = []ParserRuleContext{ParserRuleContext_OBJECT_KEYWORD, ParserRuleContext_FIRST_OBJECT_TYPE_QUALIFIER}
var OBJECT_TYPE_WITHOUT_FIRST_QUALIFIER = []ParserRuleContext{ParserRuleContext_OBJECT_KEYWORD, ParserRuleContext_SECOND_OBJECT_TYPE_QUALIFIER}
var OBJECT_CONSTRUCTOR_START = []ParserRuleContext{ParserRuleContext_OBJECT_KEYWORD, ParserRuleContext_FIRST_OBJECT_CONS_QUALIFIER}
var OBJECT_CONS_WITHOUT_FIRST_QUALIFIER = []ParserRuleContext{ParserRuleContext_OBJECT_KEYWORD, ParserRuleContext_SECOND_OBJECT_CONS_QUALIFIER}
var OBJECT_CONSTRUCTOR_RHS = []ParserRuleContext{ParserRuleContext_OPEN_BRACE, ParserRuleContext_TYPE_REFERENCE}
var ELSE_BODY = []ParserRuleContext{ParserRuleContext_IF_BLOCK, ParserRuleContext_OPEN_BRACE}
var ELSE_BLOCK = []ParserRuleContext{ParserRuleContext_ELSE_KEYWORD, ParserRuleContext_STATEMENT}
var CALL_STATEMENT = []ParserRuleContext{ParserRuleContext_CHECKING_KEYWORD, ParserRuleContext_VARIABLE_REF, ParserRuleContext_EXPRESSION}
var IMPORT_PREFIX_DECL = []ParserRuleContext{ParserRuleContext_AS_KEYWORD, ParserRuleContext_SEMICOLON}
var IMPORT_DECL_ORG_OR_MODULE_NAME_RHS = []ParserRuleContext{ParserRuleContext_SLASH, ParserRuleContext_AFTER_IMPORT_MODULE_NAME}
var AFTER_IMPORT_MODULE_NAME = []ParserRuleContext{ParserRuleContext_AS_KEYWORD, ParserRuleContext_DOT, ParserRuleContext_SEMICOLON}
var MAJOR_MINOR_VERSION_END = []ParserRuleContext{ParserRuleContext_DOT, ParserRuleContext_AS_KEYWORD, ParserRuleContext_SEMICOLON}
var RETURN_RHS = []ParserRuleContext{ParserRuleContext_SEMICOLON, ParserRuleContext_EXPRESSION}
var EXPRESSION_START = []ParserRuleContext{ParserRuleContext_BASIC_LITERAL, ParserRuleContext_NIL_LITERAL, ParserRuleContext_VARIABLE_REF, ParserRuleContext_ACCESS_EXPRESSION, ParserRuleContext_TYPE_CAST, ParserRuleContext_BRACED_EXPRESSION, ParserRuleContext_TABLE_CONSTRUCTOR_OR_QUERY_EXPRESSION, ParserRuleContext_LIST_CONSTRUCTOR, ParserRuleContext_LET_EXPRESSION, ParserRuleContext_TEMPLATE_START, ParserRuleContext_XML_KEYWORD, ParserRuleContext_STRING_KEYWORD, ParserRuleContext_BASE16_KEYWORD, ParserRuleContext_BASE64_KEYWORD, ParserRuleContext_ANON_FUNC_EXPRESSION, ParserRuleContext_ERROR_KEYWORD, ParserRuleContext_NEW_KEYWORD, ParserRuleContext_START_KEYWORD, ParserRuleContext_FLUSH_KEYWORD, ParserRuleContext_LEFT_ARROW_TOKEN, ParserRuleContext_WAIT_KEYWORD, ParserRuleContext_COMMIT_KEYWORD, ParserRuleContext_OBJECT_CONSTRUCTOR, ParserRuleContext_ERROR_CONSTRUCTOR, ParserRuleContext_TRANSACTIONAL_KEYWORD, ParserRuleContext_TYPEOF_EXPRESSION, ParserRuleContext_TRAP_KEYWORD, ParserRuleContext_UNARY_EXPRESSION, ParserRuleContext_CHECKING_KEYWORD, ParserRuleContext_MAPPING_CONSTRUCTOR, ParserRuleContext_RE_KEYWORD, ParserRuleContext_NATURAL_EXPRESSION}
var FIRST_MAPPING_FIELD_START = []ParserRuleContext{ParserRuleContext_MAPPING_FIELD, ParserRuleContext_CLOSE_BRACE}
var MAPPING_FIELD_START = []ParserRuleContext{ParserRuleContext_SPECIFIC_FIELD, ParserRuleContext_ELLIPSIS, ParserRuleContext_COMPUTED_FIELD_NAME, ParserRuleContext_READONLY_KEYWORD}
var SPECIFIC_FIELD = []ParserRuleContext{ParserRuleContext_MAPPING_FIELD_NAME, ParserRuleContext_STRING_LITERAL_TOKEN}
var SPECIFIC_FIELD_RHS = []ParserRuleContext{ParserRuleContext_COLON, ParserRuleContext_MAPPING_FIELD_END}
var MAPPING_FIELD_END = []ParserRuleContext{ParserRuleContext_CLOSE_BRACE, ParserRuleContext_COMMA}
var CONST_DECL_RHS = []ParserRuleContext{ParserRuleContext_TYPE_NAME_OR_VAR_NAME, ParserRuleContext_ASSIGN_OP}
var ARRAY_LENGTH = []ParserRuleContext{ParserRuleContext_CLOSE_BRACKET, ParserRuleContext_DECIMAL_INTEGER_LITERAL_TOKEN, ParserRuleContext_HEX_INTEGER_LITERAL_TOKEN, ParserRuleContext_ASTERISK, ParserRuleContext_VARIABLE_REF}
var PARAM_LIST = []ParserRuleContext{ParserRuleContext_CLOSE_PARENTHESIS, ParserRuleContext_REQUIRED_PARAM}
var PARAMETER_START = []ParserRuleContext{ParserRuleContext_PARAMETER_START_WITHOUT_ANNOTATION, ParserRuleContext_ANNOTATIONS}
var PARAMETER_START_WITHOUT_ANNOTATION = []ParserRuleContext{ParserRuleContext_TYPE_DESC_IN_PARAM, ParserRuleContext_ASTERISK}
var REQUIRED_PARAM_NAME_RHS = []ParserRuleContext{ParserRuleContext_PARAM_END, ParserRuleContext_ASSIGN_OP}
var PARAM_END = []ParserRuleContext{ParserRuleContext_COMMA, ParserRuleContext_CLOSE_PARENTHESIS}
var STMT_START_WITH_EXPR_RHS = []ParserRuleContext{ParserRuleContext_SEMICOLON, ParserRuleContext_ASSIGN_OP, ParserRuleContext_RIGHT_ARROW, ParserRuleContext_COMPOUND_BINARY_OPERATOR}
var EXPR_STMT_RHS = []ParserRuleContext{ParserRuleContext_SEMICOLON, ParserRuleContext_ASSIGN_OP, ParserRuleContext_RIGHT_ARROW, ParserRuleContext_COMPOUND_BINARY_OPERATOR}
var EXPRESSION_STATEMENT_START = []ParserRuleContext{ParserRuleContext_CHECKING_KEYWORD, ParserRuleContext_OPEN_PARENTHESIS, ParserRuleContext_START_KEYWORD, ParserRuleContext_FLUSH_KEYWORD}
var ANNOT_DECL_OPTIONAL_TYPE = []ParserRuleContext{ParserRuleContext_ANNOTATION_TAG, ParserRuleContext_TYPE_DESC_BEFORE_IDENTIFIER}
var CONST_DECL_TYPE = []ParserRuleContext{ParserRuleContext_TYPE_DESC_BEFORE_IDENTIFIER, ParserRuleContext_VARIABLE_NAME}
var ANNOT_DECL_RHS = []ParserRuleContext{ParserRuleContext_ANNOTATION_TAG, ParserRuleContext_ON_KEYWORD, ParserRuleContext_SEMICOLON}
var ANNOT_OPTIONAL_ATTACH_POINTS = []ParserRuleContext{ParserRuleContext_ON_KEYWORD, ParserRuleContext_SEMICOLON}
var ATTACH_POINT = []ParserRuleContext{ParserRuleContext_SOURCE_KEYWORD, ParserRuleContext_ATTACH_POINT_IDENT}
var ATTACH_POINT_IDENT = []ParserRuleContext{ParserRuleContext_SINGLE_KEYWORD_ATTACH_POINT_IDENT, ParserRuleContext_OBJECT_IDENT, ParserRuleContext_SERVICE_IDENT, ParserRuleContext_RECORD_IDENT}
var SERVICE_IDENT_RHS = []ParserRuleContext{ParserRuleContext_REMOTE_IDENT, ParserRuleContext_ATTACH_POINT_END}
var ATTACH_POINT_END = []ParserRuleContext{ParserRuleContext_COMMA, ParserRuleContext_SEMICOLON}
var XML_NAMESPACE_PREFIX_DECL = []ParserRuleContext{ParserRuleContext_AS_KEYWORD, ParserRuleContext_SEMICOLON}
var CONSTANT_EXPRESSION = []ParserRuleContext{ParserRuleContext_BASIC_LITERAL, ParserRuleContext_VARIABLE_REF, ParserRuleContext_PLUS_TOKEN, ParserRuleContext_MINUS_TOKEN, ParserRuleContext_NIL_LITERAL}
var LIST_CONSTRUCTOR_FIRST_MEMBER = []ParserRuleContext{ParserRuleContext_CLOSE_BRACKET, ParserRuleContext_LIST_CONSTRUCTOR_MEMBER}
var LIST_CONSTRUCTOR_MEMBER = []ParserRuleContext{ParserRuleContext_EXPRESSION, ParserRuleContext_ELLIPSIS}
var TYPE_CAST_PARAM = []ParserRuleContext{ParserRuleContext_TYPE_DESC_IN_ANGLE_BRACKETS, ParserRuleContext_ANNOTATIONS}
var TYPE_CAST_PARAM_RHS = []ParserRuleContext{ParserRuleContext_TYPE_DESC_IN_ANGLE_BRACKETS, ParserRuleContext_GT}
var TABLE_KEYWORD_RHS = []ParserRuleContext{ParserRuleContext_KEY_SPECIFIER, ParserRuleContext_TABLE_CONSTRUCTOR}
var ROW_LIST_RHS = []ParserRuleContext{ParserRuleContext_CLOSE_BRACKET, ParserRuleContext_MAPPING_CONSTRUCTOR}
var TABLE_ROW_END = []ParserRuleContext{ParserRuleContext_COMMA, ParserRuleContext_CLOSE_BRACKET}
var KEY_SPECIFIER_RHS = []ParserRuleContext{ParserRuleContext_CLOSE_PARENTHESIS, ParserRuleContext_VARIABLE_NAME}
var TABLE_KEY_RHS = []ParserRuleContext{ParserRuleContext_COMMA, ParserRuleContext_CLOSE_PARENTHESIS}
var LET_VAR_DECL_START = []ParserRuleContext{ParserRuleContext_TYPE_DESC_IN_TYPE_BINDING_PATTERN, ParserRuleContext_ANNOTATIONS}
var STREAM_TYPE_FIRST_PARAM_RHS = []ParserRuleContext{ParserRuleContext_COMMA, ParserRuleContext_GT}
var TEMPLATE_MEMBER = []ParserRuleContext{ParserRuleContext_TEMPLATE_STRING, ParserRuleContext_INTERPOLATION_START_TOKEN, ParserRuleContext_TEMPLATE_END}
var TEMPLATE_STRING_RHS = []ParserRuleContext{ParserRuleContext_INTERPOLATION_START_TOKEN, ParserRuleContext_TEMPLATE_END}
var KEY_CONSTRAINTS_RHS = []ParserRuleContext{ParserRuleContext_OPEN_PARENTHESIS, ParserRuleContext_LT}
var FUNCTION_KEYWORD_RHS = []ParserRuleContext{ParserRuleContext_FUNC_NAME, ParserRuleContext_FUNC_TYPE_FUNC_KEYWORD_RHS}
var FUNC_TYPE_FUNC_KEYWORD_RHS_START = []ParserRuleContext{ParserRuleContext_FUNC_TYPE_DESC_END, ParserRuleContext_OPEN_PARENTHESIS}
var TYPE_DESC_RHS = []ParserRuleContext{ParserRuleContext_END_OF_TYPE_DESC, ParserRuleContext_ARRAY_TYPE_DESCRIPTOR, ParserRuleContext_OPTIONAL_TYPE_DESCRIPTOR, ParserRuleContext_PIPE, ParserRuleContext_BITWISE_AND_OPERATOR}
var TABLE_TYPE_DESC_RHS = []ParserRuleContext{ParserRuleContext_KEY_KEYWORD, ParserRuleContext_TYPE_DESC_RHS}
var NEW_KEYWORD_RHS = []ParserRuleContext{ParserRuleContext_ARG_LIST_OPEN_PAREN, ParserRuleContext_CLASS_DESCRIPTOR_IN_NEW_EXPR, ParserRuleContext_EXPRESSION_RHS}
var TABLE_CONSTRUCTOR_OR_QUERY_START = []ParserRuleContext{ParserRuleContext_TABLE_KEYWORD, ParserRuleContext_STREAM_KEYWORD, ParserRuleContext_QUERY_EXPRESSION, ParserRuleContext_MAP_KEYWORD}
var TABLE_CONSTRUCTOR_OR_QUERY_RHS = []ParserRuleContext{ParserRuleContext_TABLE_CONSTRUCTOR, ParserRuleContext_QUERY_EXPRESSION}
var QUERY_PIPELINE_RHS = []ParserRuleContext{ParserRuleContext_QUERY_EXPRESSION_RHS, ParserRuleContext_INTERMEDIATE_CLAUSE, ParserRuleContext_QUERY_ACTION_RHS}
var INTERMEDIATE_CLAUSE_START = []ParserRuleContext{ParserRuleContext_WHERE_CLAUSE, ParserRuleContext_FROM_CLAUSE, ParserRuleContext_LET_CLAUSE, ParserRuleContext_JOIN_CLAUSE, ParserRuleContext_ORDER_BY_CLAUSE, ParserRuleContext_LIMIT_CLAUSE, ParserRuleContext_GROUP_BY_CLAUSE}
var RESULT_CLAUSE = []ParserRuleContext{ParserRuleContext_SELECT_CLAUSE, ParserRuleContext_COLLECT_CLAUSE}
var BRACED_EXPR_OR_ANON_FUNC_PARAM_RHS = []ParserRuleContext{ParserRuleContext_CLOSE_PARENTHESIS, ParserRuleContext_COMMA}
var ANNOTATION_REF_RHS = []ParserRuleContext{ParserRuleContext_ANNOTATION_END, ParserRuleContext_MAPPING_CONSTRUCTOR}
var INFER_PARAM_END_OR_PARENTHESIS_END = []ParserRuleContext{ParserRuleContext_CLOSE_PARENTHESIS, ParserRuleContext_EXPR_FUNC_BODY_START}
var OPTIONAL_PEER_WORKER = []ParserRuleContext{ParserRuleContext_PEER_WORKER_NAME, ParserRuleContext_EXPRESSION_RHS}
var TYPE_DESC_IN_TUPLE_RHS = []ParserRuleContext{ParserRuleContext_CLOSE_BRACKET, ParserRuleContext_COMMA, ParserRuleContext_ELLIPSIS}
var TUPLE_TYPE_MEMBER_RHS = []ParserRuleContext{ParserRuleContext_CLOSE_BRACKET, ParserRuleContext_COMMA}
var LIST_CONSTRUCTOR_MEMBER_END = []ParserRuleContext{ParserRuleContext_CLOSE_BRACKET, ParserRuleContext_COMMA}
var NIL_OR_PARENTHESISED_TYPE_DESC_RHS = []ParserRuleContext{ParserRuleContext_CLOSE_PARENTHESIS, ParserRuleContext_TYPE_DESCRIPTOR}
var BINDING_PATTERN = []ParserRuleContext{ParserRuleContext_BINDING_PATTERN_STARTING_IDENTIFIER, ParserRuleContext_MAPPING_BINDING_PATTERN, ParserRuleContext_LIST_BINDING_PATTERN, ParserRuleContext_ERROR_BINDING_PATTERN}
var LIST_BINDING_PATTERNS_START = []ParserRuleContext{ParserRuleContext_LIST_BINDING_PATTERN_MEMBER, ParserRuleContext_CLOSE_BRACKET}
var LIST_BINDING_PATTERN_CONTENTS = []ParserRuleContext{ParserRuleContext_BINDING_PATTERN, ParserRuleContext_REST_BINDING_PATTERN}
var LIST_BINDING_PATTERN_MEMBER_END = []ParserRuleContext{ParserRuleContext_CLOSE_BRACKET, ParserRuleContext_COMMA}
var MAPPING_BINDING_PATTERN_MEMBER = []ParserRuleContext{ParserRuleContext_REST_BINDING_PATTERN, ParserRuleContext_FIELD_BINDING_PATTERN}
var MAPPING_BINDING_PATTERN_END = []ParserRuleContext{ParserRuleContext_COMMA, ParserRuleContext_CLOSE_BRACE}
var FIELD_BINDING_PATTERN_END = []ParserRuleContext{ParserRuleContext_COMMA, ParserRuleContext_COLON, ParserRuleContext_CLOSE_BRACE}
var ERROR_BINDING_PATTERN_ERROR_KEYWORD_RHS = []ParserRuleContext{ParserRuleContext_OPEN_PARENTHESIS, ParserRuleContext_TYPE_REFERENCE}
var ERROR_ARG_LIST_BINDING_PATTERN_START = []ParserRuleContext{ParserRuleContext_SIMPLE_BINDING_PATTERN, ParserRuleContext_ERROR_FIELD_BINDING_PATTERN, ParserRuleContext_CLOSE_PARENTHESIS}
var ERROR_MESSAGE_BINDING_PATTERN_END = []ParserRuleContext{ParserRuleContext_ERROR_MESSAGE_BINDING_PATTERN_END_COMMA, ParserRuleContext_CLOSE_PARENTHESIS}
var ERROR_MESSAGE_BINDING_PATTERN_RHS = []ParserRuleContext{ParserRuleContext_ERROR_CAUSE_SIMPLE_BINDING_PATTERN, ParserRuleContext_ERROR_BINDING_PATTERN, ParserRuleContext_ERROR_FIELD_BINDING_PATTERN}
var ERROR_FIELD_BINDING_PATTERN = []ParserRuleContext{ParserRuleContext_NAMED_ARG_BINDING_PATTERN, ParserRuleContext_REST_BINDING_PATTERN}
var ERROR_FIELD_BINDING_PATTERN_END = []ParserRuleContext{ParserRuleContext_CLOSE_PARENTHESIS, ParserRuleContext_COMMA}
var REMOTE_OR_RESOURCE_CALL_OR_ASYNC_SEND_RHS = []ParserRuleContext{ParserRuleContext_DEFAULT_WORKER_NAME_IN_ASYNC_SEND, ParserRuleContext_RESOURCE_METHOD_CALL_SLASH_TOKEN, ParserRuleContext_PEER_WORKER_NAME, ParserRuleContext_METHOD_NAME}
var REMOTE_CALL_OR_ASYNC_SEND_END = []ParserRuleContext{ParserRuleContext_ARG_LIST_OPEN_PAREN, ParserRuleContext_SEMICOLON}
var RECEIVE_WORKERS = []ParserRuleContext{ParserRuleContext_SINGLE_OR_ALTERNATE_WORKER, ParserRuleContext_MULTI_RECEIVE_WORKERS}
var SINGLE_OR_ALTERNATE_WORKER_SEPARATOR = []ParserRuleContext{ParserRuleContext_SINGLE_OR_ALTERNATE_WORKER_END, ParserRuleContext_PIPE}
var RECEIVE_FIELD = []ParserRuleContext{ParserRuleContext_PEER_WORKER_NAME, ParserRuleContext_RECEIVE_FIELD_NAME}
var RECEIVE_FIELD_END = []ParserRuleContext{ParserRuleContext_CLOSE_BRACE, ParserRuleContext_COMMA}
var WAIT_KEYWORD_RHS = []ParserRuleContext{ParserRuleContext_MULTI_WAIT_FIELDS, ParserRuleContext_ALTERNATE_WAIT_EXPRS}
var WAIT_FIELD_NAME_RHS = []ParserRuleContext{ParserRuleContext_COLON, ParserRuleContext_WAIT_FIELD_END}
var WAIT_FIELD_END = []ParserRuleContext{ParserRuleContext_CLOSE_BRACE, ParserRuleContext_COMMA}
var WAIT_FUTURE_EXPR_END = []ParserRuleContext{ParserRuleContext_ALTERNATE_WAIT_EXPR_LIST_END, ParserRuleContext_PIPE}
var ENUM_MEMBER_START = []ParserRuleContext{ParserRuleContext_ENUM_MEMBER_NAME, ParserRuleContext_DOC_STRING, ParserRuleContext_ANNOTATIONS}
var ENUM_MEMBER_RHS = []ParserRuleContext{ParserRuleContext_ASSIGN_OP, ParserRuleContext_ENUM_MEMBER_END}
var ENUM_MEMBER_END = []ParserRuleContext{ParserRuleContext_COMMA, ParserRuleContext_CLOSE_BRACE}
var MEMBER_ACCESS_KEY_EXPR_END = []ParserRuleContext{ParserRuleContext_COMMA, ParserRuleContext_CLOSE_BRACKET}
var ROLLBACK_RHS = []ParserRuleContext{ParserRuleContext_SEMICOLON, ParserRuleContext_EXPRESSION}
var RETRY_KEYWORD_RHS = []ParserRuleContext{ParserRuleContext_LT, ParserRuleContext_RETRY_TYPE_PARAM_RHS}
var RETRY_TYPE_PARAM_RHS = []ParserRuleContext{ParserRuleContext_ARG_LIST_OPEN_PAREN, ParserRuleContext_RETRY_BODY}
var RETRY_BODY = []ParserRuleContext{ParserRuleContext_BLOCK_STMT, ParserRuleContext_TRANSACTION_STMT}
var LIST_BP_OR_TUPLE_TYPE_MEMBER = []ParserRuleContext{ParserRuleContext_TYPE_DESCRIPTOR, ParserRuleContext_LIST_BINDING_PATTERN_MEMBER}
var LIST_BP_OR_TUPLE_TYPE_DESC_RHS = []ParserRuleContext{ParserRuleContext_ASSIGN_OP, ParserRuleContext_VARIABLE_NAME}
var BRACKETED_LIST_MEMBER_END = []ParserRuleContext{ParserRuleContext_COMMA, ParserRuleContext_CLOSE_BRACKET}
var BRACKETED_LIST_MEMBER = []ParserRuleContext{ParserRuleContext_EXPRESSION, ParserRuleContext_BINDING_PATTERN}
var LIST_BINDING_MEMBER_OR_ARRAY_LENGTH = []ParserRuleContext{ParserRuleContext_CLOSE_BRACKET, ParserRuleContext_BINDING_PATTERN, ParserRuleContext_ARRAY_LENGTH_START}
var BRACKETED_LIST_RHS = []ParserRuleContext{ParserRuleContext_ASSIGN_OP, ParserRuleContext_TYPE_DESC_RHS_OR_BP_RHS, ParserRuleContext_EXPRESSION_RHS}
var BINDING_PATTERN_OR_VAR_REF_RHS = []ParserRuleContext{ParserRuleContext_VARIABLE_REF_RHS, ParserRuleContext_ASSIGN_OP, ParserRuleContext_TYPE_DESC_RHS_OR_BP_RHS}
var TYPE_DESC_RHS_OR_BP_RHS = []ParserRuleContext{ParserRuleContext_TYPE_DESC_RHS_IN_TYPED_BP, ParserRuleContext_LIST_BINDING_PATTERN_RHS}
var XML_NAVIGATE_EXPR = []ParserRuleContext{ParserRuleContext_XML_FILTER_EXPR, ParserRuleContext_XML_STEP_EXPR}
var XML_NAME_PATTERN_RHS = []ParserRuleContext{ParserRuleContext_GT, ParserRuleContext_PIPE}
var XML_ATOMIC_NAME_PATTERN_START = []ParserRuleContext{ParserRuleContext_ASTERISK, ParserRuleContext_XML_ATOMIC_NAME_IDENTIFIER}
var XML_ATOMIC_NAME_IDENTIFIER_RHS = []ParserRuleContext{ParserRuleContext_ASTERISK, ParserRuleContext_IDENTIFIER}
var XML_STEP_START = []ParserRuleContext{ParserRuleContext_SLASH_ASTERISK_TOKEN, ParserRuleContext_DOUBLE_SLASH_DOUBLE_ASTERISK_LT_TOKEN, ParserRuleContext_SLASH_LT_TOKEN}
var XML_STEP_EXTEND = []ParserRuleContext{ParserRuleContext_XML_STEP_EXTEND_END, ParserRuleContext_DOT, ParserRuleContext_DOT_LT_TOKEN, ParserRuleContext_MEMBER_ACCESS_KEY_EXPR}
var XML_STEP_START_END = []ParserRuleContext{ParserRuleContext_EXPRESSION_RHS, ParserRuleContext_XML_STEP_EXTENDS}
var MATCH_PATTERN_LIST_MEMBER_RHS = []ParserRuleContext{ParserRuleContext_MATCH_PATTERN_END, ParserRuleContext_PIPE}
var OPTIONAL_MATCH_GUARD = []ParserRuleContext{ParserRuleContext_RIGHT_DOUBLE_ARROW, ParserRuleContext_IF_KEYWORD}
var MATCH_PATTERN_START = []ParserRuleContext{ParserRuleContext_CONSTANT_EXPRESSION, ParserRuleContext_VAR_KEYWORD, ParserRuleContext_MAPPING_MATCH_PATTERN, ParserRuleContext_LIST_MATCH_PATTERN, ParserRuleContext_ERROR_MATCH_PATTERN}
var LIST_MATCH_PATTERNS_START = []ParserRuleContext{ParserRuleContext_LIST_MATCH_PATTERN_MEMBER, ParserRuleContext_CLOSE_BRACKET}
var LIST_MATCH_PATTERN_MEMBER = []ParserRuleContext{ParserRuleContext_MATCH_PATTERN_START, ParserRuleContext_REST_MATCH_PATTERN}
var LIST_MATCH_PATTERN_MEMBER_RHS = []ParserRuleContext{ParserRuleContext_COMMA, ParserRuleContext_CLOSE_BRACKET}
var FIELD_MATCH_PATTERNS_START = []ParserRuleContext{ParserRuleContext_FIELD_MATCH_PATTERN_MEMBER, ParserRuleContext_CLOSE_BRACE}
var FIELD_MATCH_PATTERN_MEMBER = []ParserRuleContext{ParserRuleContext_VARIABLE_NAME, ParserRuleContext_REST_MATCH_PATTERN}
var FIELD_MATCH_PATTERN_MEMBER_RHS = []ParserRuleContext{ParserRuleContext_COMMA, ParserRuleContext_CLOSE_BRACE}
var ERROR_MATCH_PATTERN_OR_CONST_PATTERN = []ParserRuleContext{ParserRuleContext_OPEN_PARENTHESIS, ParserRuleContext_MATCH_PATTERN_RHS}
var ERROR_MATCH_PATTERN_ERROR_KEYWORD_RHS = []ParserRuleContext{ParserRuleContext_OPEN_PARENTHESIS, ParserRuleContext_TYPE_REFERENCE}
var ERROR_ARG_LIST_MATCH_PATTERN_START = []ParserRuleContext{ParserRuleContext_CONSTANT_EXPRESSION, ParserRuleContext_VAR_KEYWORD, ParserRuleContext_ERROR_FIELD_MATCH_PATTERN, ParserRuleContext_CLOSE_PARENTHESIS}
var ERROR_MESSAGE_MATCH_PATTERN_END = []ParserRuleContext{ParserRuleContext_ERROR_MESSAGE_MATCH_PATTERN_END_COMMA, ParserRuleContext_CLOSE_PARENTHESIS}
var ERROR_MESSAGE_MATCH_PATTERN_RHS = []ParserRuleContext{ParserRuleContext_ERROR_CAUSE_MATCH_PATTERN, ParserRuleContext_ERROR_MATCH_PATTERN, ParserRuleContext_ERROR_FIELD_MATCH_PATTERN}
var ERROR_FIELD_MATCH_PATTERN = []ParserRuleContext{ParserRuleContext_NAMED_ARG_MATCH_PATTERN, ParserRuleContext_REST_MATCH_PATTERN}
var ERROR_FIELD_MATCH_PATTERN_RHS = []ParserRuleContext{ParserRuleContext_COMMA, ParserRuleContext_CLOSE_PARENTHESIS}
var NAMED_ARG_MATCH_PATTERN_RHS = []ParserRuleContext{ParserRuleContext_NAMED_ARG_MATCH_PATTERN, ParserRuleContext_REST_MATCH_PATTERN}
var ORDER_KEY_LIST_END = []ParserRuleContext{ParserRuleContext_ORDER_CLAUSE_END, ParserRuleContext_COMMA}
var LIST_BP_OR_LIST_CONSTRUCTOR_MEMBER = []ParserRuleContext{ParserRuleContext_LIST_BINDING_PATTERN_MEMBER, ParserRuleContext_LIST_CONSTRUCTOR_FIRST_MEMBER}
var TUPLE_TYPE_DESC_OR_LIST_CONST_MEMBER = []ParserRuleContext{ParserRuleContext_TYPE_DESCRIPTOR, ParserRuleContext_LIST_CONSTRUCTOR_FIRST_MEMBER}
var JOIN_CLAUSE_START = []ParserRuleContext{ParserRuleContext_JOIN_KEYWORD, ParserRuleContext_OUTER_KEYWORD}
var MAPPING_BP_OR_MAPPING_CONSTRUCTOR_MEMBER = []ParserRuleContext{ParserRuleContext_MAPPING_BINDING_PATTERN_MEMBER, ParserRuleContext_MAPPING_FIELD}
var LISTENERS_LIST_END = []ParserRuleContext{ParserRuleContext_OBJECT_CONSTRUCTOR_BLOCK, ParserRuleContext_COMMA}
var FUNC_TYPE_DESC_START = []ParserRuleContext{ParserRuleContext_FUNCTION_KEYWORD, ParserRuleContext_FUNC_TYPE_FIRST_QUALIFIER}
var FUNC_TYPE_DESC_START_WITHOUT_FIRST_QUAL = []ParserRuleContext{ParserRuleContext_FUNCTION_KEYWORD, ParserRuleContext_FUNC_TYPE_SECOND_QUALIFIER}
var MODULE_CLASS_DEFINITION_START = []ParserRuleContext{ParserRuleContext_CLASS_KEYWORD, ParserRuleContext_FIRST_CLASS_TYPE_QUALIFIER}
var CLASS_DEF_WITHOUT_FIRST_QUALIFIER = []ParserRuleContext{ParserRuleContext_CLASS_KEYWORD, ParserRuleContext_SECOND_CLASS_TYPE_QUALIFIER}
var CLASS_DEF_WITHOUT_SECOND_QUALIFIER = []ParserRuleContext{ParserRuleContext_CLASS_KEYWORD, ParserRuleContext_THIRD_CLASS_TYPE_QUALIFIER}
var CLASS_DEF_WITHOUT_THIRD_QUALIFIER = []ParserRuleContext{ParserRuleContext_CLASS_KEYWORD, ParserRuleContext_FOURTH_CLASS_TYPE_QUALIFIER}
var REGULAR_COMPOUND_STMT_RHS = []ParserRuleContext{ParserRuleContext_STATEMENT, ParserRuleContext_ON_FAIL_CLAUSE}
var NAMED_WORKER_DECL_START = []ParserRuleContext{ParserRuleContext_WORKER_KEYWORD, ParserRuleContext_TRANSACTIONAL_KEYWORD}
var SERVICE_DECL_START = []ParserRuleContext{ParserRuleContext_SERVICE_KEYWORD, ParserRuleContext_SERVICE_DECL_QUALIFIER}
var OPTIONAL_SERVICE_DECL_TYPE = []ParserRuleContext{ParserRuleContext_TYPE_DESC_IN_SERVICE, ParserRuleContext_OPTIONAL_ABSOLUTE_PATH}
var OPTIONAL_ABSOLUTE_PATH = []ParserRuleContext{ParserRuleContext_ABSOLUTE_RESOURCE_PATH, ParserRuleContext_STRING_LITERAL_TOKEN, ParserRuleContext_ON_KEYWORD}
var ABSOLUTE_RESOURCE_PATH_START = []ParserRuleContext{ParserRuleContext_SLASH, ParserRuleContext_ABSOLUTE_PATH_SINGLE_SLASH}
var ABSOLUTE_RESOURCE_PATH_END = []ParserRuleContext{ParserRuleContext_SLASH, ParserRuleContext_SERVICE_DECL_RHS}
var SERVICE_DECL_OR_VAR_DECL = []ParserRuleContext{ParserRuleContext_OPTIONAL_ABSOLUTE_PATH, ParserRuleContext_SERVICE_VAR_DECL_RHS}
var OPTIONAL_RELATIVE_PATH = []ParserRuleContext{ParserRuleContext_OPEN_PARENTHESIS, ParserRuleContext_RELATIVE_RESOURCE_PATH}
var FUNC_DEF_OR_TYPE_DESC_RHS = []ParserRuleContext{ParserRuleContext_OPEN_PARENTHESIS, ParserRuleContext_RELATIVE_RESOURCE_PATH, ParserRuleContext_SEMICOLON, ParserRuleContext_ASSIGN_OP}
var RELATIVE_RESOURCE_PATH_START = []ParserRuleContext{ParserRuleContext_DOT, ParserRuleContext_RESOURCE_PATH_SEGMENT}
var RESOURCE_PATH_SEGMENT = []ParserRuleContext{ParserRuleContext_PATH_SEGMENT_IDENT, ParserRuleContext_RESOURCE_PATH_PARAM}
var PATH_PARAM_OPTIONAL_ANNOTS = []ParserRuleContext{ParserRuleContext_TYPE_DESC_IN_PATH_PARAM, ParserRuleContext_ANNOTATIONS}
var PATH_PARAM_ELLIPSIS = []ParserRuleContext{ParserRuleContext_OPTIONAL_PATH_PARAM_NAME, ParserRuleContext_ELLIPSIS}
var OPTIONAL_PATH_PARAM_NAME = []ParserRuleContext{ParserRuleContext_VARIABLE_NAME, ParserRuleContext_CLOSE_BRACKET}
var RELATIVE_RESOURCE_PATH_END = []ParserRuleContext{ParserRuleContext_RESOURCE_ACCESSOR_DEF_OR_DECL_RHS, ParserRuleContext_SLASH}
var CONFIG_VAR_DECL_RHS = []ParserRuleContext{ParserRuleContext_EXPRESSION, ParserRuleContext_QUESTION_MARK}
var ERROR_CONSTRUCTOR_RHS = []ParserRuleContext{ParserRuleContext_ARG_LIST_OPEN_PAREN, ParserRuleContext_TYPE_REFERENCE}
var OPTIONAL_TYPE_PARAMETER = []ParserRuleContext{ParserRuleContext_LT, ParserRuleContext_TYPE_DESC_RHS}
var MAP_TYPE_OR_TYPE_REF = []ParserRuleContext{ParserRuleContext_COLON, ParserRuleContext_LT}
var OBJECT_TYPE_OR_TYPE_REF = []ParserRuleContext{ParserRuleContext_COLON, ParserRuleContext_OBJECT_TYPE_OBJECT_KEYWORD_RHS}
var STREAM_TYPE_OR_TYPE_REF = []ParserRuleContext{ParserRuleContext_COLON, ParserRuleContext_LT}
var TABLE_TYPE_OR_TYPE_REF = []ParserRuleContext{ParserRuleContext_COLON, ParserRuleContext_ROW_TYPE_PARAM}
var PARAMETERIZED_TYPE_OR_TYPE_REF = []ParserRuleContext{ParserRuleContext_COLON, ParserRuleContext_OPTIONAL_TYPE_PARAMETER}
var TYPE_DESC_RHS_OR_TYPE_REF = []ParserRuleContext{ParserRuleContext_COLON, ParserRuleContext_TYPE_DESC_RHS}
var TRANSACTION_STMT_RHS_OR_TYPE_REF = []ParserRuleContext{ParserRuleContext_TYPE_REF_COLON, ParserRuleContext_TRANSACTION_STMT_TRANSACTION_KEYWORD_RHS}
var TABLE_CONS_OR_QUERY_EXPR_OR_VAR_REF = []ParserRuleContext{ParserRuleContext_VAR_REF_COLON, ParserRuleContext_EXPRESSION_START_TABLE_KEYWORD_RHS}
var QUERY_EXPR_OR_VAR_REF = []ParserRuleContext{ParserRuleContext_VAR_REF_COLON, ParserRuleContext_QUERY_CONSTRUCT_TYPE_RHS}
var ERROR_CONS_EXPR_OR_VAR_REF = []ParserRuleContext{ParserRuleContext_VAR_REF_COLON, ParserRuleContext_ERROR_CONS_ERROR_KEYWORD_RHS}
var QUALIFIED_IDENTIFIER = []ParserRuleContext{ParserRuleContext_QUALIFIED_IDENTIFIER_START_IDENTIFIER, ParserRuleContext_QUALIFIED_IDENTIFIER_PREDECLARED_PREFIX}
var MODULE_VAR_DECL_START = []ParserRuleContext{ParserRuleContext_VAR_DECL_STMT, ParserRuleContext_MODULE_VAR_FIRST_QUAL}
var MODULE_VAR_WITHOUT_FIRST_QUAL = []ParserRuleContext{ParserRuleContext_VAR_DECL_STMT, ParserRuleContext_MODULE_VAR_SECOND_QUAL}
var MODULE_VAR_WITHOUT_SECOND_QUAL = []ParserRuleContext{ParserRuleContext_VAR_DECL_STMT, ParserRuleContext_MODULE_VAR_THIRD_QUAL}
var EXPR_START_OR_INFERRED_TYPEDESC_DEFAULT_START = []ParserRuleContext{ParserRuleContext_EXPRESSION, ParserRuleContext_INFERRED_TYPEDESC_DEFAULT_START_LT}
var TYPE_CAST_PARAM_START_OR_INFERRED_TYPEDESC_DEFAULT_END = []ParserRuleContext{ParserRuleContext_TYPE_CAST_PARAM_START, ParserRuleContext_INFERRED_TYPEDESC_DEFAULT_END_GT}
var END_OF_PARAMS_OR_NEXT_PARAM_START = []ParserRuleContext{ParserRuleContext_CLOSE_PARENTHESIS, ParserRuleContext_COMMA}
var PARAM_START = []ParserRuleContext{ParserRuleContext_TYPE_DESC_IN_PARAM, ParserRuleContext_ANNOTATIONS}
var PARAM_RHS = []ParserRuleContext{ParserRuleContext_VARIABLE_NAME, ParserRuleContext_REST_PARAM_RHS}
var FUNC_TYPE_PARAM_RHS = []ParserRuleContext{ParserRuleContext_PARAM_END, ParserRuleContext_PARAM_RHS}
var ANNOTATION_DECL_START = []ParserRuleContext{ParserRuleContext_ANNOTATION_KEYWORD, ParserRuleContext_CONST_KEYWORD}
var OPTIONAL_RESOURCE_ACCESS_PATH = []ParserRuleContext{ParserRuleContext_RESOURCE_ACCESS_PATH_SEGMENT, ParserRuleContext_OPTIONAL_RESOURCE_ACCESS_METHOD}
var RESOURCE_ACCESS_PATH_SEGMENT = []ParserRuleContext{ParserRuleContext_IDENTIFIER, ParserRuleContext_OPEN_BRACKET}
var COMPUTED_SEGMENT_OR_REST_SEGMENT = []ParserRuleContext{ParserRuleContext_EXPRESSION, ParserRuleContext_ELLIPSIS}
var RESOURCE_ACCESS_SEGMENT_RHS = []ParserRuleContext{ParserRuleContext_SLASH, ParserRuleContext_OPTIONAL_RESOURCE_ACCESS_METHOD}
var OPTIONAL_RESOURCE_ACCESS_METHOD = []ParserRuleContext{ParserRuleContext_DOT, ParserRuleContext_OPTIONAL_RESOURCE_ACCESS_ACTION_ARG_LIST}
var OPTIONAL_RESOURCE_ACCESS_ACTION_ARG_LIST = []ParserRuleContext{ParserRuleContext_ARG_LIST_OPEN_PAREN, ParserRuleContext_ACTION_END}
var OPTIONAL_TOP_LEVEL_SEMICOLON = []ParserRuleContext{ParserRuleContext_TOP_LEVEL_NODE, ParserRuleContext_SEMICOLON}
var TUPLE_MEMBER = []ParserRuleContext{ParserRuleContext_ANNOTATIONS, ParserRuleContext_TYPE_DESC_IN_TUPLE}
var NATURAL_EXPRESSION_START = []ParserRuleContext{ParserRuleContext_NATURAL_KEYWORD, ParserRuleContext_CONST_KEYWORD}
var OPTIONAL_PARENTHESIZED_ARG_LIST = []ParserRuleContext{ParserRuleContext_ARG_LIST_OPEN_PAREN, ParserRuleContext_OPEN_BRACE}

func NewBallerinaParserErrorHandlerFromTokenReader(tokenReader TokenReader) BallerinaParserErrorHandler {
	this := BallerinaParserErrorHandler{}
	super(tokenReader)
	return this
}

func (this *BallerinaParserErrorHandler) isEndOfObjectTypeNode(nextLookahead int) bool {
	nextToken := this.this.tokenReader.peek(nextLookahead)
	switch nextToken.kind {
	case CLOSE_BRACE_TOKEN,
		EOF_TOKEN,
		CLOSE_BRACE_PIPE_TOKEN,
		TYPE_KEYWORD,
		SERVICE_KEYWORD:
		true
	default:
		nextNextToken := this.this.tokenReader.peek(nextLookahead + 1)
		switch nextNextToken.kind {
		case CLOSE_BRACE_TOKEN,
			EOF_TOKEN,
			CLOSE_BRACE_PIPE_TOKEN,
			TYPE_KEYWORD,
			SERVICE_KEYWORD:
			true
		default:
			false
		}
	}
}

func (this *BallerinaParserErrorHandler) seekMatch(currentCtx ParserRuleContext, lookahead int, currentDepth int, isEntryPoint bool) Result {
	var hasMatch bool
	var skipRule bool
	matchingRulesCount := 0
	for currentDepth < LOOKAHEAD_LIMIT {
		skipRule = false
		lookahead = this.getNextLookahead(lookahead)
		nextToken := this.this.tokenReader.peek(lookahead)
		switch currentCtx {
		case EOF:
			hasMatch = (nextToken.kind == SyntaxKind_EOF_TOKEN)
			break
		case FUNC_NAME,
			CLASS_NAME,
			VARIABLE_NAME,
			TYPE_NAME,
			IMPORT_ORG_OR_MODULE_NAME,
			IMPORT_MODULE_NAME,
			MAPPING_FIELD_NAME,
			QUALIFIED_IDENTIFIER_START_IDENTIFIER,
			SIMPLE_TYPE_DESC_IDENTIFIER,
			IDENTIFIER,
			ANNOTATION_TAG,
			NAMESPACE_PREFIX,
			WORKER_NAME,
			IMPLICIT_ANON_FUNC_PARAM,
			METHOD_NAME,
			RECEIVE_FIELD_NAME,
			WAIT_FIELD_NAME,
			FIELD_BINDING_PATTERN_NAME,
			XML_ATOMIC_NAME_IDENTIFIER,
			SIMPLE_BINDING_PATTERN,
			ERROR_CAUSE_SIMPLE_BINDING_PATTERN,
			PATH_SEGMENT_IDENT,
			MODULE_ENUM_NAME,
			ENUM_MEMBER_NAME,
			NAMED_ARG_BINDING_PATTERN:
			hasMatch = (nextToken.kind == SyntaxKind_IDENTIFIER_TOKEN)
			break
		case IMPORT_PREFIX:
			hasMatch = ((nextToken.kind == SyntaxKind_IDENTIFIER_TOKEN) || this.BallerinaParser.isPredeclaredPrefix(nextToken.kind))
			break
		case QUALIFIED_IDENTIFIER_PREDECLARED_PREFIX:
			hasMatch = this.BallerinaParser.isPredeclaredPrefix(nextToken.kind)
			break
		case OPEN_PARENTHESIS,
			PARENTHESISED_TYPE_DESC_START,
			ARG_LIST_OPEN_PAREN:
			hasMatch = (nextToken.kind == SyntaxKind_OPEN_PAREN_TOKEN)
			break
		case CLOSE_PARENTHESIS,
			ARG_LIST_CLOSE_PAREN:
			hasMatch = (nextToken.kind == SyntaxKind_CLOSE_PAREN_TOKEN)
			break
		case SIMPLE_TYPE_DESCRIPTOR:
			hasMatch = (((this.BallerinaParser.isSimpleType(nextToken.kind) || (nextToken.kind == SyntaxKind_ERROR_KEYWORD)) || (nextToken.kind == SyntaxKind_STREAM_KEYWORD)) || (nextToken.kind == SyntaxKind_TYPEDESC_KEYWORD))
			break
		case OPEN_BRACE:
			hasMatch = (nextToken.kind == SyntaxKind_OPEN_BRACE_TOKEN)
			break
		case CLOSE_BRACE:
			hasMatch = (nextToken.kind == SyntaxKind_CLOSE_BRACE_TOKEN)
			break
		case ASSIGN_OP:
			hasMatch = (nextToken.kind == SyntaxKind_EQUAL_TOKEN)
			break
		case SEMICOLON:
			hasMatch = (nextToken.kind == SyntaxKind_SEMICOLON_TOKEN)
			break
		case BINARY_OPERATOR:
			hasMatch = this.isBinaryOperator(nextToken)
			break
		case COMMA,
			ERROR_MESSAGE_BINDING_PATTERN_END_COMMA,
			ERROR_MESSAGE_MATCH_PATTERN_END_COMMA:
			hasMatch = (nextToken.kind == SyntaxKind_COMMA_TOKEN)
			break
		case CLOSED_RECORD_BODY_END:
			hasMatch = (nextToken.kind == SyntaxKind_CLOSE_BRACE_PIPE_TOKEN)
			break
		case CLOSED_RECORD_BODY_START:
			hasMatch = (nextToken.kind == SyntaxKind_OPEN_BRACE_PIPE_TOKEN)
			break
		case ELLIPSIS:
			hasMatch = (nextToken.kind == SyntaxKind_ELLIPSIS_TOKEN)
			break
		case QUESTION_MARK:
			hasMatch = (nextToken.kind == SyntaxKind_QUESTION_MARK_TOKEN)
			break
		case FIRST_OBJECT_CONS_QUALIFIER,
			SECOND_OBJECT_CONS_QUALIFIER,
			FIRST_OBJECT_TYPE_QUALIFIER,
			SECOND_OBJECT_TYPE_QUALIFIER:
			hasMatch = (((nextToken.kind == SyntaxKind_CLIENT_KEYWORD) || (nextToken.kind == SyntaxKind_ISOLATED_KEYWORD)) || (nextToken.kind == SyntaxKind_SERVICE_KEYWORD))
			break
		case FIRST_CLASS_TYPE_QUALIFIER,
			SECOND_CLASS_TYPE_QUALIFIER,
			THIRD_CLASS_TYPE_QUALIFIER,
			FOURTH_CLASS_TYPE_QUALIFIER:
			hasMatch = (((((nextToken.kind == SyntaxKind_DISTINCT_KEYWORD) || (nextToken.kind == SyntaxKind_CLIENT_KEYWORD)) || (nextToken.kind == SyntaxKind_READONLY_KEYWORD)) || (nextToken.kind == SyntaxKind_ISOLATED_KEYWORD)) || (nextToken.kind == SyntaxKind_SERVICE_KEYWORD))
			break
		case OPEN_BRACKET,
			TUPLE_TYPE_DESC_START:
			hasMatch = (nextToken.kind == SyntaxKind_OPEN_BRACKET_TOKEN)
			break
		case CLOSE_BRACKET:
			hasMatch = (nextToken.kind == SyntaxKind_CLOSE_BRACKET_TOKEN)
			break
		case DOT,
			METHOD_CALL_DOT:
			hasMatch = (nextToken.kind == SyntaxKind_DOT_TOKEN)
			break
		case BOOLEAN_LITERAL:
			hasMatch = ((nextToken.kind == SyntaxKind_TRUE_KEYWORD) || (nextToken.kind == SyntaxKind_FALSE_KEYWORD))
			break
		case DECIMAL_INTEGER_LITERAL_TOKEN:
			hasMatch = (nextToken.kind == SyntaxKind_DECIMAL_INTEGER_LITERAL_TOKEN)
			break
		case SLASH,
			ABSOLUTE_PATH_SINGLE_SLASH,
			RESOURCE_METHOD_CALL_SLASH_TOKEN:
			hasMatch = (nextToken.kind == SyntaxKind_SLASH_TOKEN)
			break
		case BASIC_LITERAL:
			hasMatch = this.isBasicLiteral(nextToken.kind)
			break
		case COLON,
			VAR_REF_COLON,
			TYPE_REF_COLON:
			hasMatch = (nextToken.kind == SyntaxKind_COLON_TOKEN)
			break
		case STRING_LITERAL_TOKEN:
			hasMatch = (nextToken.kind == SyntaxKind_STRING_LITERAL_TOKEN)
			break
		case UNARY_OPERATOR:
			hasMatch = this.isUnaryOperator(nextToken)
			break
		case HEX_INTEGER_LITERAL_TOKEN:
			hasMatch = (nextToken.kind == SyntaxKind_HEX_INTEGER_LITERAL_TOKEN)
			break
		case AT:
			hasMatch = (nextToken.kind == SyntaxKind_AT_TOKEN)
			break
		case RIGHT_ARROW:
			hasMatch = (nextToken.kind == SyntaxKind_RIGHT_ARROW_TOKEN)
			break
		case PARAMETERIZED_TYPE:
			hasMatch = this.BallerinaParser.isParameterizedTypeToken(nextToken.kind)
			break
		case LT,
			STREAM_TYPE_PARAM_START_TOKEN,
			INFERRED_TYPEDESC_DEFAULT_START_LT:
			hasMatch = (nextToken.kind == SyntaxKind_LT_TOKEN)
			break
		case GT,
			INFERRED_TYPEDESC_DEFAULT_END_GT:
			hasMatch = (nextToken.kind == SyntaxKind_GT_TOKEN)
			break
		case FIELD_IDENT:
			hasMatch = (nextToken.kind == SyntaxKind_FIELD_KEYWORD)
			break
		case FUNCTION_IDENT:
			hasMatch = (nextToken.kind == SyntaxKind_FUNCTION_KEYWORD)
			break
		case IDENT_AFTER_OBJECT_IDENT:
			hasMatch = ((nextToken.kind == SyntaxKind_FUNCTION_KEYWORD) || (nextToken.kind == SyntaxKind_FIELD_KEYWORD))
			break
		case SINGLE_KEYWORD_ATTACH_POINT_IDENT:
			hasMatch = this.isSingleKeywordAttachPointIdent(nextToken.kind)
			break
		case OBJECT_IDENT:
			hasMatch = (nextToken.kind == SyntaxKind_OBJECT_KEYWORD)
			break
		case RECORD_IDENT:
			hasMatch = (nextToken.kind == SyntaxKind_RECORD_KEYWORD)
			break
		case SERVICE_IDENT:
			hasMatch = (nextToken.kind == SyntaxKind_SERVICE_KEYWORD)
			break
		case REMOTE_IDENT:
			hasMatch = (nextToken.kind == SyntaxKind_REMOTE_KEYWORD)
			break
		case DECIMAL_FLOATING_POINT_LITERAL_TOKEN:
			hasMatch = (nextToken.kind == SyntaxKind_DECIMAL_FLOATING_POINT_LITERAL_TOKEN)
			break
		case HEX_FLOATING_POINT_LITERAL_TOKEN:
			hasMatch = (nextToken.kind == SyntaxKind_HEX_FLOATING_POINT_LITERAL_TOKEN)
			break
		case PIPE:
			hasMatch = (nextToken.kind == SyntaxKind_PIPE_TOKEN)
			break
		case TEMPLATE_START,
			TEMPLATE_END:
			hasMatch = (nextToken.kind == SyntaxKind_BACKTICK_TOKEN)
			break
		case ASTERISK:
			hasMatch = (nextToken.kind == SyntaxKind_ASTERISK_TOKEN)
			break
		case BITWISE_AND_OPERATOR:
			hasMatch = (nextToken.kind == SyntaxKind_BITWISE_AND_TOKEN)
			break
		case EXPR_FUNC_BODY_START,
			RIGHT_DOUBLE_ARROW:
			hasMatch = (nextToken.kind == SyntaxKind_RIGHT_DOUBLE_ARROW_TOKEN)
			break
		case PLUS_TOKEN:
			hasMatch = (nextToken.kind == SyntaxKind_PLUS_TOKEN)
			break
		case MINUS_TOKEN:
			hasMatch = (nextToken.kind == SyntaxKind_MINUS_TOKEN)
			break
		case SIGNED_INT_OR_FLOAT_RHS:
			hasMatch = this.BallerinaParser.isIntOrFloat(nextToken)
			break
		case SYNC_SEND_TOKEN:
			hasMatch = (nextToken.kind == SyntaxKind_SYNC_SEND_TOKEN)
			break
		case PEER_WORKER_NAME:
			hasMatch = ((nextToken.kind == SyntaxKind_FUNCTION_KEYWORD) || (nextToken.kind == SyntaxKind_IDENTIFIER_TOKEN))
			break
		case LEFT_ARROW_TOKEN:
			hasMatch = (nextToken.kind == SyntaxKind_LEFT_ARROW_TOKEN)
			break
		case ANNOT_CHAINING_TOKEN:
			hasMatch = (nextToken.kind == SyntaxKind_ANNOT_CHAINING_TOKEN)
			break
		case OPTIONAL_CHAINING_TOKEN:
			hasMatch = (nextToken.kind == SyntaxKind_OPTIONAL_CHAINING_TOKEN)
			break
		case TRANSACTIONAL_KEYWORD:
			hasMatch = (nextToken.kind == SyntaxKind_TRANSACTIONAL_KEYWORD)
			break
		case SERVICE_DECL_QUALIFIER:
			hasMatch = (nextToken.kind == SyntaxKind_ISOLATED_KEYWORD)
			break
		case UNION_OR_INTERSECTION_TOKEN:
			hasMatch = ((nextToken.kind == SyntaxKind_PIPE_TOKEN) || (nextToken.kind == SyntaxKind_BITWISE_AND_TOKEN))
			break
		case DOT_LT_TOKEN:
			hasMatch = (nextToken.kind == SyntaxKind_DOT_LT_TOKEN)
			break
		case SLASH_LT_TOKEN:
			hasMatch = (nextToken.kind == SyntaxKind_SLASH_LT_TOKEN)
			break
		case DOUBLE_SLASH_DOUBLE_ASTERISK_LT_TOKEN:
			hasMatch = (nextToken.kind == SyntaxKind_DOUBLE_SLASH_DOUBLE_ASTERISK_LT_TOKEN)
			break
		case SLASH_ASTERISK_TOKEN:
			hasMatch = (nextToken.kind == SyntaxKind_SLASH_ASTERISK_TOKEN)
			break
		case KEY_KEYWORD:
			hasMatch = ((nextToken.kind == SyntaxKind_KEY_KEYWORD) || this.BallerinaParser.isKeyKeyword(nextToken))
			break
		case NATURAL_KEYWORD:
			hasMatch = ((nextToken.kind == SyntaxKind_NATURAL_KEYWORD) || this.BallerinaParser.isNaturalKeyword(nextToken))
			break
		case VAR_KEYWORD:
			hasMatch = (nextToken.kind == SyntaxKind_VAR_KEYWORD)
			break
		case ORDER_DIRECTION:
			hasMatch = ((nextToken.kind == SyntaxKind_ASCENDING_KEYWORD) || (nextToken.kind == SyntaxKind_DESCENDING_KEYWORD))
			break
		case OBJECT_MEMBER_VISIBILITY_QUAL:
			hasMatch = ((nextToken.kind == SyntaxKind_PRIVATE_KEYWORD) || (nextToken.kind == SyntaxKind_PUBLIC_KEYWORD))
			break
		case OBJECT_METHOD_FIRST_QUALIFIER,
			OBJECT_METHOD_SECOND_QUALIFIER,
			OBJECT_METHOD_THIRD_QUALIFIER,
			OBJECT_METHOD_FOURTH_QUALIFIER:
			hasMatch = ((((nextToken.kind == SyntaxKind_ISOLATED_KEYWORD) || (nextToken.kind == SyntaxKind_TRANSACTIONAL_KEYWORD)) || (nextToken.kind == SyntaxKind_REMOTE_KEYWORD)) || (nextToken.kind == SyntaxKind_RESOURCE_KEYWORD))
			break
		case FUNC_DEF_FIRST_QUALIFIER,
			FUNC_DEF_SECOND_QUALIFIER,
			FUNC_TYPE_FIRST_QUALIFIER,
			FUNC_TYPE_SECOND_QUALIFIER:
			hasMatch = ((nextToken.kind == SyntaxKind_ISOLATED_KEYWORD) || (nextToken.kind == SyntaxKind_TRANSACTIONAL_KEYWORD))
			break
		case MODULE_VAR_FIRST_QUAL,
			MODULE_VAR_THIRD_QUAL:
			hasMatch = (((nextToken.kind == SyntaxKind_FINAL_KEYWORD) || (nextToken.kind == SyntaxKind_ISOLATED_KEYWORD)) || (nextToken.kind == SyntaxKind_CONFIGURABLE_KEYWORD))
			break
		case COMPOUND_BINARY_OPERATOR:
			hasMatch = this.BallerinaParser.isCompoundBinaryOperator(nextToken.kind)
			break
		case IS_KEYWORD:
			hasMatch = ((nextToken.kind == SyntaxKind_IS_KEYWORD) || (nextToken.kind == SyntaxKind_NOT_IS_KEYWORD))
			break
		case VARIABLE_REF,
			TYPE_REFERENCE_IN_TYPE_INCLUSION,
			TYPE_REFERENCE,
			ANNOT_REFERENCE,
			FIELD_ACCESS_IDENTIFIER,
			TYPE_DESC_IN_ANNOTATION_DECL,
			TYPE_DESC_BEFORE_IDENTIFIER,
			TYPE_DESC_IN_RECORD_FIELD,
			TYPE_DESC_IN_PARAM,
			TYPE_DESC_IN_TYPE_BINDING_PATTERN,
			TYPE_DESC_IN_TYPE_DEF,
			TYPE_DESC_IN_ANGLE_BRACKETS,
			TYPE_DESC_IN_RETURN_TYPE_DESC,
			TYPE_DESC_IN_EXPRESSION,
			TYPE_DESC_IN_STREAM_TYPE_DESC,
			TYPE_DESC_IN_PARENTHESIS,
			TYPE_DESC_IN_SERVICE,
			TYPE_DESC_IN_PATH_PARAM,
			TYPE_DESC_BEFORE_IDENTIFIER_IN_GROUPING_KEY:
			fallthrough
		default:
			if this.isKeyword(currentCtx) {
				expectedTokenKind := this.getExpectedKeywordKind(currentCtx)
				hasMatch = ((nextToken.kind == expectedTokenKind) || this.BallerinaParser.isKeywordMatch(expectedTokenKind, nextToken))
				break
			}
			if this.hasAlternativePaths(currentCtx) {
				return this.seekMatchInAlternativePaths(currentCtx, lookahead, currentDepth, matchingRulesCount,
					isEntryPoint)
			}
			skipRule = true
			hasMatch = true
			break
		}
		if !hasMatch {
			return this.fixAndContinue(currentCtx, lookahead, currentDepth, matchingRulesCount, isEntryPoint)
		}
		if !skipRule {
			currentDepth++
			matchingRulesCount++
			lookahead++
			isEntryPoint = false
		}
		currentCtx = this.getNextRule(currentCtx, lookahead)
	}
	result := nil
	result.solution = nil
	return result
}

func (this *BallerinaParserErrorHandler) getNextLookahead(lookahead int) int {
	for this.tokenReader.peek(lookahead).kind == SyntaxKind_DOCUMENTATION_STRING {
		lookahead++
	}
	return lookahead
}

func (this *BallerinaParserErrorHandler) isKeyword(currentCtx ParserRuleContext) bool {
	switch currentCtx {
	case EOF,
		PUBLIC_KEYWORD,
		PRIVATE_KEYWORD,
		FUNCTION_KEYWORD,
		NEW_KEYWORD,
		SELECT_KEYWORD,
		WHERE_KEYWORD,
		FROM_KEYWORD,
		ORDER_KEYWORD,
		GROUP_KEYWORD,
		BY_KEYWORD,
		START_KEYWORD,
		FLUSH_KEYWORD,
		DEFAULT_WORKER_NAME_IN_ASYNC_SEND,
		WAIT_KEYWORD,
		CHECKING_KEYWORD,
		FAIL_KEYWORD,
		DO_KEYWORD,
		TRANSACTION_KEYWORD,
		TRANSACTIONAL_KEYWORD,
		COMMIT_KEYWORD,
		RETRY_KEYWORD,
		ROLLBACK_KEYWORD,
		ENUM_KEYWORD,
		MATCH_KEYWORD,
		RETURNS_KEYWORD,
		EXTERNAL_KEYWORD,
		RECORD_KEYWORD,
		TYPE_KEYWORD,
		OBJECT_KEYWORD,
		ABSTRACT_KEYWORD,
		CLIENT_KEYWORD,
		IF_KEYWORD,
		ELSE_KEYWORD,
		WHILE_KEYWORD,
		PANIC_KEYWORD,
		AS_KEYWORD,
		LOCK_KEYWORD,
		IMPORT_KEYWORD,
		CONTINUE_KEYWORD,
		BREAK_KEYWORD,
		RETURN_KEYWORD,
		SERVICE_KEYWORD,
		ON_KEYWORD,
		LISTENER_KEYWORD,
		CONST_KEYWORD,
		FINAL_KEYWORD,
		TYPEOF_KEYWORD,
		IS_KEYWORD,
		NOT_IS_KEYWORD,
		NULL_KEYWORD,
		ANNOTATION_KEYWORD,
		SOURCE_KEYWORD,
		XMLNS_KEYWORD,
		WORKER_KEYWORD,
		FORK_KEYWORD,
		TRAP_KEYWORD,
		FOREACH_KEYWORD,
		IN_KEYWORD,
		TABLE_KEYWORD,
		KEY_KEYWORD,
		ERROR_KEYWORD,
		LET_KEYWORD,
		STREAM_KEYWORD,
		XML_KEYWORD,
		RE_KEYWORD,
		STRING_KEYWORD,
		BASE16_KEYWORD,
		BASE64_KEYWORD,
		DISTINCT_KEYWORD,
		CONFLICT_KEYWORD,
		LIMIT_KEYWORD,
		EQUALS_KEYWORD,
		JOIN_KEYWORD,
		OUTER_KEYWORD,
		CLASS_KEYWORD,
		MAP_KEYWORD,
		COLLECT_KEYWORD,
		NATURAL_KEYWORD:
		true
	default:
		false
	}
}

func (this *BallerinaParserErrorHandler) hasAlternativePaths(currentCtx ParserRuleContext) bool {
	switch currentCtx {
	case TOP_LEVEL_NODE,
		TOP_LEVEL_NODE_WITHOUT_MODIFIER,
		TOP_LEVEL_NODE_WITHOUT_METADATA,
		FUNC_OPTIONAL_RETURNS,
		FUNC_BODY_OR_TYPE_DESC_RHS,
		ANON_FUNC_BODY,
		FUNC_BODY,
		EXPRESSION,
		TERMINAL_EXPRESSION,
		VAR_DECL_STMT_RHS,
		EXPRESSION_RHS,
		VARIABLE_REF_RHS,
		STATEMENT,
		STATEMENT_WITHOUT_ANNOTS,
		PARAM_LIST,
		REQUIRED_PARAM_NAME_RHS,
		TYPE_NAME_OR_VAR_NAME,
		FIELD_DESCRIPTOR_RHS,
		FIELD_OR_REST_DESCIPTOR_RHS,
		RECORD_BODY_END,
		RECORD_BODY_START,
		TYPE_DESCRIPTOR,
		TYPE_DESC_WITHOUT_ISOLATED,
		RECORD_FIELD_OR_RECORD_END,
		RECORD_FIELD_START,
		RECORD_FIELD_WITHOUT_METADATA,
		ARG_START,
		ARG_START_OR_ARG_LIST_END,
		NAMED_OR_POSITIONAL_ARG_RHS,
		ARG_END,
		CLASS_MEMBER_OR_OBJECT_MEMBER_START,
		OBJECT_CONSTRUCTOR_MEMBER_START,
		CLASS_MEMBER_OR_OBJECT_MEMBER_WITHOUT_META,
		OBJECT_CONS_MEMBER_WITHOUT_META,
		OPTIONAL_FIELD_INITIALIZER,
		OBJECT_METHOD_START,
		OBJECT_FUNC_OR_FIELD,
		OBJECT_FUNC_OR_FIELD_WITHOUT_VISIBILITY,
		OBJECT_TYPE_START,
		OBJECT_CONSTRUCTOR_START,
		ELSE_BLOCK,
		ELSE_BODY,
		CALL_STMT_START,
		IMPORT_PREFIX_DECL,
		IMPORT_DECL_ORG_OR_MODULE_NAME_RHS,
		AFTER_IMPORT_MODULE_NAME,
		RETURN_STMT_RHS,
		ACCESS_EXPRESSION,
		FIRST_MAPPING_FIELD,
		MAPPING_FIELD,
		SPECIFIC_FIELD,
		SPECIFIC_FIELD_RHS,
		MAPPING_FIELD_END,
		OPTIONAL_ABSOLUTE_PATH,
		CONST_DECL_TYPE,
		CONST_DECL_RHS,
		ARRAY_LENGTH,
		PARAMETER_START,
		PARAMETER_START_WITHOUT_ANNOTATION,
		STMT_START_WITH_EXPR_RHS,
		EXPR_STMT_RHS,
		EXPRESSION_STATEMENT_START,
		ANNOT_DECL_OPTIONAL_TYPE,
		ANNOT_DECL_RHS,
		ANNOT_OPTIONAL_ATTACH_POINTS,
		ATTACH_POINT,
		ATTACH_POINT_IDENT,
		ATTACH_POINT_END,
		XML_NAMESPACE_PREFIX_DECL,
		CONSTANT_EXPRESSION_START,
		TYPE_DESC_RHS,
		LIST_CONSTRUCTOR_FIRST_MEMBER,
		LIST_CONSTRUCTOR_MEMBER,
		TYPE_CAST_PARAM,
		TYPE_CAST_PARAM_RHS,
		TABLE_KEYWORD_RHS,
		ROW_LIST_RHS,
		TABLE_ROW_END,
		KEY_SPECIFIER_RHS,
		TABLE_KEY_RHS,
		LET_VAR_DECL_START,
		ORDER_KEY_LIST_END,
		STREAM_TYPE_FIRST_PARAM_RHS,
		TEMPLATE_MEMBER,
		TEMPLATE_STRING_RHS,
		FUNCTION_KEYWORD_RHS,
		FUNC_TYPE_FUNC_KEYWORD_RHS_START,
		WORKER_NAME_RHS,
		BINDING_PATTERN,
		LIST_BINDING_PATTERNS_START,
		LIST_BINDING_PATTERN_MEMBER_END,
		FIELD_BINDING_PATTERN_END,
		LIST_BINDING_PATTERN_MEMBER,
		MAPPING_BINDING_PATTERN_END,
		MAPPING_BINDING_PATTERN_MEMBER,
		KEY_CONSTRAINTS_RHS,
		TABLE_TYPE_DESC_RHS,
		NEW_KEYWORD_RHS,
		TABLE_CONSTRUCTOR_OR_QUERY_START,
		TABLE_CONSTRUCTOR_OR_QUERY_RHS,
		QUERY_PIPELINE_RHS,
		BRACED_EXPR_OR_ANON_FUNC_PARAM_RHS,
		ANON_FUNC_PARAM_RHS,
		PARAM_END,
		ANNOTATION_REF_RHS,
		INFER_PARAM_END_OR_PARENTHESIS_END,
		TYPE_DESC_IN_TUPLE_RHS,
		TUPLE_TYPE_MEMBER_RHS,
		LIST_CONSTRUCTOR_MEMBER_END,
		NIL_OR_PARENTHESISED_TYPE_DESC_RHS,
		REMOTE_OR_RESOURCE_CALL_OR_ASYNC_SEND_RHS,
		REMOTE_CALL_OR_ASYNC_SEND_END,
		RECEIVE_WORKERS,
		RECEIVE_FIELD,
		RECEIVE_FIELD_END,
		WAIT_KEYWORD_RHS,
		WAIT_FIELD_NAME_RHS,
		WAIT_FIELD_END,
		WAIT_FUTURE_EXPR_END,
		OPTIONAL_PEER_WORKER,
		ENUM_MEMBER_START,
		ENUM_MEMBER_RHS,
		ENUM_MEMBER_END,
		MEMBER_ACCESS_KEY_EXPR_END,
		ROLLBACK_RHS,
		RETRY_KEYWORD_RHS,
		RETRY_TYPE_PARAM_RHS,
		RETRY_BODY,
		STMT_START_BRACKETED_LIST_MEMBER,
		STMT_START_BRACKETED_LIST_RHS,
		BINDING_PATTERN_OR_EXPR_RHS,
		BINDING_PATTERN_OR_VAR_REF_RHS,
		BRACKETED_LIST_RHS,
		BRACKETED_LIST_MEMBER,
		BRACKETED_LIST_MEMBER_END,
		TYPE_DESC_RHS_OR_BP_RHS,
		LIST_BINDING_MEMBER_OR_ARRAY_LENGTH,
		XML_NAVIGATE_EXPR,
		XML_NAME_PATTERN_RHS,
		XML_ATOMIC_NAME_PATTERN_START,
		XML_ATOMIC_NAME_IDENTIFIER_RHS,
		XML_STEP_START,
		XML_STEP_EXTEND,
		FUNC_TYPE_DESC_RHS_OR_ANON_FUNC_BODY,
		OPTIONAL_MATCH_GUARD,
		MATCH_PATTERN_LIST_MEMBER_RHS,
		MATCH_PATTERN_START,
		LIST_MATCH_PATTERNS_START,
		LIST_MATCH_PATTERN_MEMBER,
		LIST_MATCH_PATTERN_MEMBER_RHS,
		ERROR_BINDING_PATTERN_ERROR_KEYWORD_RHS,
		ERROR_ARG_LIST_BINDING_PATTERN_START,
		ERROR_MESSAGE_BINDING_PATTERN_END,
		ERROR_MESSAGE_BINDING_PATTERN_RHS,
		ERROR_FIELD_BINDING_PATTERN,
		ERROR_FIELD_BINDING_PATTERN_END,
		FIELD_MATCH_PATTERNS_START,
		FIELD_MATCH_PATTERN_MEMBER,
		FIELD_MATCH_PATTERN_MEMBER_RHS,
		ERROR_MATCH_PATTERN_OR_CONST_PATTERN,
		ERROR_MATCH_PATTERN_ERROR_KEYWORD_RHS,
		ERROR_ARG_LIST_MATCH_PATTERN_START,
		ERROR_MESSAGE_MATCH_PATTERN_END,
		ERROR_MESSAGE_MATCH_PATTERN_RHS,
		ERROR_FIELD_MATCH_PATTERN,
		ERROR_FIELD_MATCH_PATTERN_RHS,
		NAMED_ARG_MATCH_PATTERN_RHS,
		EXTERNAL_FUNC_BODY_OPTIONAL_ANNOTS,
		LIST_BP_OR_LIST_CONSTRUCTOR_MEMBER,
		TUPLE_TYPE_DESC_OR_LIST_CONST_MEMBER,
		OBJECT_METHOD_WITHOUT_FIRST_QUALIFIER,
		OBJECT_METHOD_WITHOUT_SECOND_QUALIFIER,
		OBJECT_METHOD_WITHOUT_THIRD_QUALIFIER,
		JOIN_CLAUSE_START,
		INTERMEDIATE_CLAUSE_START,
		MAPPING_BP_OR_MAPPING_CONSTRUCTOR_MEMBER,
		TYPE_DESC_OR_EXPR_RHS,
		LISTENERS_LIST_END,
		REGULAR_COMPOUND_STMT_RHS,
		NAMED_WORKER_DECL_START,
		FUNC_TYPE_DESC_START,
		ANON_FUNC_EXPRESSION_START,
		MODULE_CLASS_DEFINITION_START,
		OBJECT_CONSTRUCTOR_TYPE_REF,
		OBJECT_FIELD_QUALIFIER,
		OPTIONAL_SERVICE_DECL_TYPE,
		SERVICE_IDENT_RHS,
		ABSOLUTE_RESOURCE_PATH_START,
		ABSOLUTE_RESOURCE_PATH_END,
		SERVICE_DECL_OR_VAR_DECL,
		OPTIONAL_RELATIVE_PATH,
		RELATIVE_RESOURCE_PATH_START,
		RELATIVE_RESOURCE_PATH_END,
		RESOURCE_PATH_SEGMENT,
		PATH_PARAM_OPTIONAL_ANNOTS,
		PATH_PARAM_ELLIPSIS,
		OPTIONAL_PATH_PARAM_NAME,
		OBJECT_CONS_WITHOUT_FIRST_QUALIFIER,
		OBJECT_TYPE_WITHOUT_FIRST_QUALIFIER,
		CONFIG_VAR_DECL_RHS,
		SERVICE_DECL_START,
		ERROR_CONSTRUCTOR_RHS,
		OPTIONAL_TYPE_PARAMETER,
		MAP_TYPE_OR_TYPE_REF,
		OBJECT_TYPE_OR_TYPE_REF,
		STREAM_TYPE_OR_TYPE_REF,
		TABLE_TYPE_OR_TYPE_REF,
		PARAMETERIZED_TYPE_OR_TYPE_REF,
		TYPE_DESC_RHS_OR_TYPE_REF,
		TRANSACTION_STMT_RHS_OR_TYPE_REF,
		TABLE_CONS_OR_QUERY_EXPR_OR_VAR_REF,
		QUERY_EXPR_OR_VAR_REF,
		ERROR_CONS_EXPR_OR_VAR_REF,
		QUALIFIED_IDENTIFIER,
		CLASS_DEF_WITHOUT_FIRST_QUALIFIER,
		CLASS_DEF_WITHOUT_SECOND_QUALIFIER,
		CLASS_DEF_WITHOUT_THIRD_QUALIFIER,
		FUNC_DEF_START,
		FUNC_DEF_WITHOUT_FIRST_QUALIFIER,
		FUNC_TYPE_DESC_START_WITHOUT_FIRST_QUAL,
		MODULE_VAR_DECL_START,
		MODULE_VAR_WITHOUT_FIRST_QUAL,
		MODULE_VAR_WITHOUT_SECOND_QUAL,
		FUNC_DEF_OR_TYPE_DESC_RHS,
		CLASS_DESCRIPTOR,
		EXPR_START_OR_INFERRED_TYPEDESC_DEFAULT_START,
		TYPE_CAST_PARAM_START_OR_INFERRED_TYPEDESC_DEFAULT_END,
		END_OF_PARAMS_OR_NEXT_PARAM_START,
		ASSIGNMENT_STMT_RHS,
		PARAM_START,
		PARAM_RHS,
		FUNC_TYPE_PARAM_RHS,
		ANNOTATION_DECL_START,
		ON_FAIL_OPTIONAL_BINDING_PATTERN,
		OPTIONAL_RESOURCE_ACCESS_PATH,
		RESOURCE_ACCESS_PATH_SEGMENT,
		COMPUTED_SEGMENT_OR_REST_SEGMENT,
		RESOURCE_ACCESS_SEGMENT_RHS,
		OPTIONAL_RESOURCE_ACCESS_METHOD,
		OPTIONAL_RESOURCE_ACCESS_ACTION_ARG_LIST,
		OPTIONAL_TOP_LEVEL_SEMICOLON,
		TUPLE_MEMBER,
		GROUPING_KEY_LIST_ELEMENT,
		GROUPING_KEY_LIST_ELEMENT_END,
		RESULT_CLAUSE,
		SINGLE_OR_ALTERNATE_WORKER_SEPARATOR,
		XML_STEP_START_END,
		NATURAL_EXPRESSION_START,
		OPTIONAL_PARENTHESIZED_ARG_LIST:
		true
	default:
		false
	}
}

func (this *BallerinaParserErrorHandler) getShortestAlternative(currentCtx ParserRuleContext) ParserRuleContext {
	switch currentCtx {
	case TOP_LEVEL_NODE,
		TOP_LEVEL_NODE_WITHOUT_MODIFIER,
		TOP_LEVEL_NODE_WITHOUT_METADATA:
		ParserRuleContext_EOF
	case FUNC_OPTIONAL_RETURNS:
		ParserRuleContext_RETURNS_KEYWORD
	case FUNC_BODY_OR_TYPE_DESC_RHS:
		ParserRuleContext_FUNC_BODY
	case ANON_FUNC_BODY:
		ParserRuleContext_EXPLICIT_ANON_FUNC_EXPR_BODY_START
	case FUNC_BODY:
		ParserRuleContext_FUNC_BODY_BLOCK
	case EXPRESSION,
		TERMINAL_EXPRESSION:
		ParserRuleContext_VARIABLE_REF
	case VAR_DECL_STMT_RHS:
		ParserRuleContext_SEMICOLON
	case EXPRESSION_RHS,
		VARIABLE_REF_RHS:
		ParserRuleContext_BINARY_OPERATOR
	case STATEMENT,
		STATEMENT_WITHOUT_ANNOTS:
		ParserRuleContext_VAR_DECL_STMT
	case PARAM_LIST:
		ParserRuleContext_CLOSE_PARENTHESIS
	case REQUIRED_PARAM_NAME_RHS:
		ParserRuleContext_PARAM_END
	case TYPE_NAME_OR_VAR_NAME:
		ParserRuleContext_VARIABLE_NAME
	case FIELD_DESCRIPTOR_RHS:
		ParserRuleContext_SEMICOLON
	case FIELD_OR_REST_DESCIPTOR_RHS:
		ParserRuleContext_VARIABLE_NAME
	case RECORD_BODY_END:
		ParserRuleContext_CLOSE_BRACE
	case RECORD_BODY_START, OPTIONAL_PARENTHESIZED_ARG_LIST:
		ParserRuleContext_OPEN_BRACE
	case TYPE_DESCRIPTOR:
		ParserRuleContext_SIMPLE_TYPE_DESC_IDENTIFIER
	case TYPE_DESC_WITHOUT_ISOLATED:
		ParserRuleContext_FUNC_TYPE_DESC
	case RECORD_FIELD_OR_RECORD_END:
		ParserRuleContext_RECORD_BODY_END
	case RECORD_FIELD_START,
		RECORD_FIELD_WITHOUT_METADATA:
		ParserRuleContext_TYPE_DESC_IN_RECORD_FIELD
	case ARG_START:
		ParserRuleContext_EXPRESSION
	case ARG_START_OR_ARG_LIST_END:
		ParserRuleContext_ARG_LIST_END
	case NAMED_OR_POSITIONAL_ARG_RHS:
		ParserRuleContext_ARG_END
	case ARG_END:
		ParserRuleContext_ARG_LIST_END
	case CLASS_MEMBER_OR_OBJECT_MEMBER_START,
		OBJECT_CONSTRUCTOR_MEMBER_START,
		CLASS_MEMBER_OR_OBJECT_MEMBER_WITHOUT_META,
		OBJECT_CONS_MEMBER_WITHOUT_META:
		ParserRuleContext_CLOSE_BRACE
	case OPTIONAL_FIELD_INITIALIZER:
		ParserRuleContext_SEMICOLON
	case ON_FAIL_OPTIONAL_BINDING_PATTERN:
		ParserRuleContext_BLOCK_STMT
	case OBJECT_METHOD_START:
		ParserRuleContext_FUNC_DEF_OR_FUNC_TYPE
	case OBJECT_FUNC_OR_FIELD:
		ParserRuleContext_OBJECT_FUNC_OR_FIELD_WITHOUT_VISIBILITY
	case OBJECT_FUNC_OR_FIELD_WITHOUT_VISIBILITY:
		ParserRuleContext_OBJECT_FIELD_START
	case OBJECT_TYPE_START,
		OBJECT_CONSTRUCTOR_START:
		ParserRuleContext_OBJECT_KEYWORD
	case ELSE_BLOCK:
		ParserRuleContext_STATEMENT
	case ELSE_BODY:
		ParserRuleContext_OPEN_BRACE
	case CALL_STMT_START:
		ParserRuleContext_VARIABLE_REF
	case IMPORT_PREFIX_DECL:
		ParserRuleContext_SEMICOLON
	case IMPORT_DECL_ORG_OR_MODULE_NAME_RHS:
		ParserRuleContext_AFTER_IMPORT_MODULE_NAME
	case AFTER_IMPORT_MODULE_NAME,
		RETURN_STMT_RHS:
		ParserRuleContext_SEMICOLON
	case ACCESS_EXPRESSION:
		ParserRuleContext_VARIABLE_REF
	case FIRST_MAPPING_FIELD:
		ParserRuleContext_CLOSE_BRACE
	case MAPPING_FIELD:
		ParserRuleContext_SPECIFIC_FIELD
	case SPECIFIC_FIELD:
		ParserRuleContext_MAPPING_FIELD_NAME
	case SPECIFIC_FIELD_RHS:
		ParserRuleContext_MAPPING_FIELD_END
	case MAPPING_FIELD_END:
		ParserRuleContext_CLOSE_BRACE
	case OPTIONAL_ABSOLUTE_PATH:
		ParserRuleContext_ON_KEYWORD
	case CONST_DECL_TYPE:
		ParserRuleContext_VARIABLE_NAME
	case CONST_DECL_RHS:
		ParserRuleContext_ASSIGN_OP
	case ARRAY_LENGTH:
		ParserRuleContext_CLOSE_BRACKET
	case PARAMETER_START:
		ParserRuleContext_PARAMETER_START_WITHOUT_ANNOTATION
	case PARAMETER_START_WITHOUT_ANNOTATION:
		ParserRuleContext_TYPE_DESC_IN_PARAM
	case STMT_START_WITH_EXPR_RHS,
		EXPR_STMT_RHS:
		ParserRuleContext_SEMICOLON
	case EXPRESSION_STATEMENT_START:
		ParserRuleContext_VARIABLE_REF
	case ANNOT_DECL_OPTIONAL_TYPE:
		ParserRuleContext_ANNOTATION_TAG
	case ANNOT_DECL_RHS,
		ANNOT_OPTIONAL_ATTACH_POINTS:
		ParserRuleContext_SEMICOLON
	case ATTACH_POINT:
		ParserRuleContext_ATTACH_POINT_IDENT
	case ATTACH_POINT_IDENT:
		ParserRuleContext_SINGLE_KEYWORD_ATTACH_POINT_IDENT
	case ATTACH_POINT_END,
		BINDING_PATTERN_OR_VAR_REF_RHS:
		ParserRuleContext_SEMICOLON
	case XML_NAMESPACE_PREFIX_DECL:
		ParserRuleContext_SEMICOLON
	case CONSTANT_EXPRESSION_START:
		ParserRuleContext_VARIABLE_REF
	case TYPE_DESC_RHS:
		ParserRuleContext_END_OF_TYPE_DESC
	case LIST_CONSTRUCTOR_FIRST_MEMBER:
		ParserRuleContext_CLOSE_BRACKET
	case LIST_CONSTRUCTOR_MEMBER:
		ParserRuleContext_EXPRESSION
	case TYPE_CAST_PARAM,
		TYPE_CAST_PARAM_RHS:
		ParserRuleContext_TYPE_DESC_IN_ANGLE_BRACKETS
	case TABLE_KEYWORD_RHS:
		ParserRuleContext_TABLE_CONSTRUCTOR
	case ROW_LIST_RHS,
		TABLE_ROW_END:
		ParserRuleContext_CLOSE_BRACKET
	case KEY_SPECIFIER_RHS,
		TABLE_KEY_RHS:
		ParserRuleContext_CLOSE_PARENTHESIS
	case LET_VAR_DECL_START:
		ParserRuleContext_TYPE_DESC_IN_TYPE_BINDING_PATTERN
	case ORDER_KEY_LIST_END:
		ParserRuleContext_ORDER_CLAUSE_END
	case GROUPING_KEY_LIST_ELEMENT_END:
		ParserRuleContext_GROUP_BY_CLAUSE_END
	case GROUPING_KEY_LIST_ELEMENT:
		ParserRuleContext_VARIABLE_NAME
	case STREAM_TYPE_FIRST_PARAM_RHS:
		ParserRuleContext_GT
	case TEMPLATE_MEMBER,
		TEMPLATE_STRING_RHS:
		ParserRuleContext_TEMPLATE_END
	case FUNCTION_KEYWORD_RHS:
		ParserRuleContext_FUNC_TYPE_FUNC_KEYWORD_RHS
	case FUNC_TYPE_FUNC_KEYWORD_RHS_START:
		ParserRuleContext_FUNC_TYPE_DESC_END
	case WORKER_NAME_RHS:
		ParserRuleContext_BLOCK_STMT
	case BINDING_PATTERN:
		ParserRuleContext_BINDING_PATTERN_STARTING_IDENTIFIER
	case LIST_BINDING_PATTERNS_START,
		LIST_BINDING_PATTERN_MEMBER_END:
		ParserRuleContext_CLOSE_BRACKET
	case FIELD_BINDING_PATTERN_END:
		ParserRuleContext_CLOSE_BRACE
	case LIST_BINDING_PATTERN_MEMBER:
		ParserRuleContext_BINDING_PATTERN
	case MAPPING_BINDING_PATTERN_END:
		ParserRuleContext_CLOSE_BRACE
	case MAPPING_BINDING_PATTERN_MEMBER:
		ParserRuleContext_FIELD_BINDING_PATTERN
	case KEY_CONSTRAINTS_RHS:
		ParserRuleContext_OPEN_PARENTHESIS
	case TABLE_TYPE_DESC_RHS:
		ParserRuleContext_TYPE_DESC_RHS
	case NEW_KEYWORD_RHS:
		ParserRuleContext_EXPRESSION_RHS
	case TABLE_CONSTRUCTOR_OR_QUERY_START:
		ParserRuleContext_TABLE_KEYWORD
	case TABLE_CONSTRUCTOR_OR_QUERY_RHS:
		ParserRuleContext_TABLE_CONSTRUCTOR
	case QUERY_PIPELINE_RHS:
		ParserRuleContext_QUERY_EXPRESSION_RHS
	case BRACED_EXPR_OR_ANON_FUNC_PARAM_RHS,
		ANON_FUNC_PARAM_RHS:
		ParserRuleContext_CLOSE_PARENTHESIS
	case PARAM_END:
		ParserRuleContext_CLOSE_PARENTHESIS
	case ANNOTATION_REF_RHS:
		ParserRuleContext_ANNOTATION_END
	case INFER_PARAM_END_OR_PARENTHESIS_END:
		ParserRuleContext_CLOSE_PARENTHESIS
	case TYPE_DESC_IN_TUPLE_RHS:
		ParserRuleContext_CLOSE_BRACKET
	case TUPLE_TYPE_MEMBER_RHS:
		ParserRuleContext_CLOSE_BRACKET
	case LIST_CONSTRUCTOR_MEMBER_END:
		ParserRuleContext_CLOSE_BRACKET
	case NIL_OR_PARENTHESISED_TYPE_DESC_RHS:
		ParserRuleContext_CLOSE_PARENTHESIS
	case REMOTE_OR_RESOURCE_CALL_OR_ASYNC_SEND_RHS:
		ParserRuleContext_PEER_WORKER_NAME
	case REMOTE_CALL_OR_ASYNC_SEND_END:
		ParserRuleContext_SEMICOLON
	case RECEIVE_WORKERS,
		RECEIVE_FIELD:
		ParserRuleContext_PEER_WORKER_NAME
	case RECEIVE_FIELD_END:
		ParserRuleContext_CLOSE_BRACE
	case WAIT_KEYWORD_RHS:
		ParserRuleContext_MULTI_WAIT_FIELDS
	case WAIT_FIELD_NAME_RHS:
		ParserRuleContext_WAIT_FIELD_END
	case WAIT_FIELD_END:
		ParserRuleContext_CLOSE_BRACE
	case WAIT_FUTURE_EXPR_END:
		ParserRuleContext_ALTERNATE_WAIT_EXPR_LIST_END
	case OPTIONAL_PEER_WORKER:
		ParserRuleContext_EXPRESSION_RHS
	case ENUM_MEMBER_START:
		ParserRuleContext_ENUM_MEMBER_NAME
	case ENUM_MEMBER_RHS:
		ParserRuleContext_ENUM_MEMBER_END
	case ENUM_MEMBER_END:
		ParserRuleContext_CLOSE_BRACE
	case MEMBER_ACCESS_KEY_EXPR_END:
		ParserRuleContext_CLOSE_BRACKET
	case ROLLBACK_RHS:
		ParserRuleContext_SEMICOLON
	case RETRY_KEYWORD_RHS:
		ParserRuleContext_RETRY_TYPE_PARAM_RHS
	case RETRY_TYPE_PARAM_RHS:
		ParserRuleContext_RETRY_BODY
	case RETRY_BODY:
		ParserRuleContext_BLOCK_STMT
	case STMT_START_BRACKETED_LIST_MEMBER:
		ParserRuleContext_TYPE_DESCRIPTOR
	case STMT_START_BRACKETED_LIST_RHS:
		ParserRuleContext_VARIABLE_NAME
	case BINDING_PATTERN_OR_EXPR_RHS,
		BRACKETED_LIST_RHS:
		ParserRuleContext_TYPE_DESC_RHS_OR_BP_RHS
	case BRACKETED_LIST_MEMBER:
		ParserRuleContext_EXPRESSION
	case BRACKETED_LIST_MEMBER_END:
		ParserRuleContext_CLOSE_BRACKET
	case TYPE_DESC_RHS_OR_BP_RHS:
		ParserRuleContext_TYPE_DESC_RHS_IN_TYPED_BP
	case LIST_BINDING_MEMBER_OR_ARRAY_LENGTH:
		ParserRuleContext_BINDING_PATTERN
	case XML_NAVIGATE_EXPR:
		ParserRuleContext_XML_FILTER_EXPR
	case XML_NAME_PATTERN_RHS:
		ParserRuleContext_GT
	case XML_ATOMIC_NAME_PATTERN_START:
		ParserRuleContext_XML_ATOMIC_NAME_IDENTIFIER
	case XML_ATOMIC_NAME_IDENTIFIER_RHS:
		ParserRuleContext_IDENTIFIER
	case XML_STEP_START:
		ParserRuleContext_SLASH_ASTERISK_TOKEN
	case FUNC_TYPE_DESC_RHS_OR_ANON_FUNC_BODY:
		ParserRuleContext_ANON_FUNC_BODY
	case OPTIONAL_MATCH_GUARD:
		ParserRuleContext_RIGHT_DOUBLE_ARROW
	case MATCH_PATTERN_LIST_MEMBER_RHS:
		ParserRuleContext_MATCH_PATTERN_END
	case MATCH_PATTERN_START:
		ParserRuleContext_CONSTANT_EXPRESSION
	case LIST_MATCH_PATTERNS_START:
		ParserRuleContext_CLOSE_BRACKET
	case LIST_MATCH_PATTERN_MEMBER:
		ParserRuleContext_MATCH_PATTERN_START
	case LIST_MATCH_PATTERN_MEMBER_RHS:
		ParserRuleContext_CLOSE_BRACKET
	case ERROR_BINDING_PATTERN_ERROR_KEYWORD_RHS:
		ParserRuleContext_OPEN_PARENTHESIS
	case ERROR_ARG_LIST_BINDING_PATTERN_START,
		ERROR_MESSAGE_BINDING_PATTERN_END:
		ParserRuleContext_CLOSE_PARENTHESIS
	case ERROR_MESSAGE_BINDING_PATTERN_RHS:
		ParserRuleContext_ERROR_CAUSE_SIMPLE_BINDING_PATTERN
	case ERROR_FIELD_BINDING_PATTERN:
		ParserRuleContext_NAMED_ARG_BINDING_PATTERN
	case ERROR_FIELD_BINDING_PATTERN_END:
		ParserRuleContext_CLOSE_PARENTHESIS
	case FIELD_MATCH_PATTERNS_START:
		ParserRuleContext_CLOSE_BRACE
	case FIELD_MATCH_PATTERN_MEMBER:
		ParserRuleContext_VARIABLE_NAME
	case FIELD_MATCH_PATTERN_MEMBER_RHS:
		ParserRuleContext_CLOSE_BRACE
	case ERROR_MATCH_PATTERN_OR_CONST_PATTERN:
		ParserRuleContext_MATCH_PATTERN_RHS
	case ERROR_MATCH_PATTERN_ERROR_KEYWORD_RHS:
		ParserRuleContext_OPEN_PARENTHESIS
	case ERROR_ARG_LIST_MATCH_PATTERN_START,
		ERROR_MESSAGE_MATCH_PATTERN_END:
		ParserRuleContext_CLOSE_PARENTHESIS
	case ERROR_MESSAGE_MATCH_PATTERN_RHS:
		ParserRuleContext_ERROR_CAUSE_MATCH_PATTERN
	case ERROR_FIELD_MATCH_PATTERN:
		ParserRuleContext_NAMED_ARG_MATCH_PATTERN
	case ERROR_FIELD_MATCH_PATTERN_RHS:
		ParserRuleContext_CLOSE_PARENTHESIS
	case NAMED_ARG_MATCH_PATTERN_RHS:
		ParserRuleContext_NAMED_ARG_MATCH_PATTERN
	case EXTERNAL_FUNC_BODY_OPTIONAL_ANNOTS:
		ParserRuleContext_EXTERNAL_KEYWORD
	case LIST_BP_OR_LIST_CONSTRUCTOR_MEMBER:
		ParserRuleContext_LIST_BINDING_PATTERN_MEMBER
	case TUPLE_TYPE_DESC_OR_LIST_CONST_MEMBER:
		ParserRuleContext_TYPE_DESCRIPTOR
	case OBJECT_METHOD_WITHOUT_FIRST_QUALIFIER,
		OBJECT_METHOD_WITHOUT_SECOND_QUALIFIER,
		OBJECT_METHOD_WITHOUT_THIRD_QUALIFIER,
		FUNC_DEF:
		ParserRuleContext_FUNC_DEF_OR_FUNC_TYPE
	case JOIN_CLAUSE_START:
		ParserRuleContext_JOIN_KEYWORD
	case INTERMEDIATE_CLAUSE_START:
		ParserRuleContext_WHERE_CLAUSE
	case MAPPING_BP_OR_MAPPING_CONSTRUCTOR_MEMBER:
		ParserRuleContext_MAPPING_BINDING_PATTERN_MEMBER
	case TYPE_DESC_OR_EXPR_RHS:
		ParserRuleContext_TYPE_DESC_RHS_OR_BP_RHS
	case LISTENERS_LIST_END:
		ParserRuleContext_OBJECT_CONSTRUCTOR_BLOCK
	case REGULAR_COMPOUND_STMT_RHS:
		ParserRuleContext_STATEMENT
	case NAMED_WORKER_DECL_START:
		ParserRuleContext_WORKER_KEYWORD
	case FUNC_TYPE_DESC_START,
		FUNC_DEF_START,
		ANON_FUNC_EXPRESSION_START:
		ParserRuleContext_FUNCTION_KEYWORD
	case MODULE_CLASS_DEFINITION_START:
		ParserRuleContext_CLASS_KEYWORD
	case OBJECT_CONSTRUCTOR_TYPE_REF:
		ParserRuleContext_OPEN_BRACE
	case OBJECT_FIELD_QUALIFIER:
		ParserRuleContext_TYPE_DESC_BEFORE_IDENTIFIER
	case OPTIONAL_SERVICE_DECL_TYPE:
		ParserRuleContext_OPTIONAL_ABSOLUTE_PATH
	case SERVICE_IDENT_RHS:
		ParserRuleContext_ATTACH_POINT_END
	case ABSOLUTE_RESOURCE_PATH_START:
		ParserRuleContext_ABSOLUTE_PATH_SINGLE_SLASH
	case ABSOLUTE_RESOURCE_PATH_END:
		ParserRuleContext_SERVICE_DECL_RHS
	case SERVICE_DECL_OR_VAR_DECL:
		ParserRuleContext_SERVICE_VAR_DECL_RHS
	case OPTIONAL_RELATIVE_PATH:
		ParserRuleContext_OPEN_PARENTHESIS
	case RELATIVE_RESOURCE_PATH_START:
		ParserRuleContext_DOT
	case RELATIVE_RESOURCE_PATH_END:
		ParserRuleContext_RESOURCE_ACCESSOR_DEF_OR_DECL_RHS
	case RESOURCE_PATH_SEGMENT:
		ParserRuleContext_PATH_SEGMENT_IDENT
	case PATH_PARAM_OPTIONAL_ANNOTS:
		ParserRuleContext_TYPE_DESC_IN_PATH_PARAM
	case PATH_PARAM_ELLIPSIS,
		OPTIONAL_PATH_PARAM_NAME:
		ParserRuleContext_CLOSE_BRACKET
	case OBJECT_CONS_WITHOUT_FIRST_QUALIFIER,
		OBJECT_TYPE_WITHOUT_FIRST_QUALIFIER:
		ParserRuleContext_OBJECT_KEYWORD
	case CONFIG_VAR_DECL_RHS:
		ParserRuleContext_EXPRESSION
	case SERVICE_DECL_START:
		ParserRuleContext_SERVICE_KEYWORD
	case ERROR_CONSTRUCTOR_RHS:
		ParserRuleContext_ARG_LIST_OPEN_PAREN
	case OPTIONAL_TYPE_PARAMETER:
		ParserRuleContext_TYPE_DESC_RHS
	case MAP_TYPE_OR_TYPE_REF:
		ParserRuleContext_LT
	case OBJECT_TYPE_OR_TYPE_REF:
		ParserRuleContext_OBJECT_TYPE_OBJECT_KEYWORD_RHS
	case STREAM_TYPE_OR_TYPE_REF:
		ParserRuleContext_LT
	case TABLE_TYPE_OR_TYPE_REF:
		ParserRuleContext_ROW_TYPE_PARAM
	case PARAMETERIZED_TYPE_OR_TYPE_REF:
		ParserRuleContext_OPTIONAL_TYPE_PARAMETER
	case TYPE_DESC_RHS_OR_TYPE_REF:
		ParserRuleContext_TYPE_DESC_RHS
	case TRANSACTION_STMT_RHS_OR_TYPE_REF:
		ParserRuleContext_TRANSACTION_STMT_TRANSACTION_KEYWORD_RHS
	case TABLE_CONS_OR_QUERY_EXPR_OR_VAR_REF:
		ParserRuleContext_EXPRESSION_START_TABLE_KEYWORD_RHS
	case QUERY_EXPR_OR_VAR_REF:
		ParserRuleContext_QUERY_CONSTRUCT_TYPE_RHS
	case ERROR_CONS_EXPR_OR_VAR_REF:
		ParserRuleContext_ERROR_CONS_ERROR_KEYWORD_RHS
	case QUALIFIED_IDENTIFIER:
		ParserRuleContext_QUALIFIED_IDENTIFIER_START_IDENTIFIER
	case CLASS_DEF_WITHOUT_FIRST_QUALIFIER,
		CLASS_DEF_WITHOUT_SECOND_QUALIFIER,
		CLASS_DEF_WITHOUT_THIRD_QUALIFIER:
		ParserRuleContext_CLASS_KEYWORD
	case FUNC_DEF_WITHOUT_FIRST_QUALIFIER:
		ParserRuleContext_FUNC_DEF_OR_FUNC_TYPE
	case FUNC_TYPE_DESC_START_WITHOUT_FIRST_QUAL:
		ParserRuleContext_FUNCTION_KEYWORD
	case MODULE_VAR_DECL_START,
		MODULE_VAR_WITHOUT_FIRST_QUAL,
		MODULE_VAR_WITHOUT_SECOND_QUAL:
		ParserRuleContext_VAR_DECL_STMT
	case FUNC_DEF_OR_TYPE_DESC_RHS:
		ParserRuleContext_SEMICOLON
	case CLASS_DESCRIPTOR:
		ParserRuleContext_TYPE_REFERENCE
	case EXPR_START_OR_INFERRED_TYPEDESC_DEFAULT_START:
		ParserRuleContext_EXPRESSION
	case TYPE_CAST_PARAM_START_OR_INFERRED_TYPEDESC_DEFAULT_END:
		ParserRuleContext_INFERRED_TYPEDESC_DEFAULT_END_GT
	case END_OF_PARAMS_OR_NEXT_PARAM_START:
		ParserRuleContext_CLOSE_PARENTHESIS
	case ASSIGNMENT_STMT_RHS:
		ParserRuleContext_ASSIGN_OP
	case PARAM_START:
		ParserRuleContext_TYPE_DESC_IN_PARAM
	case PARAM_RHS:
		ParserRuleContext_VARIABLE_NAME
	case FUNC_TYPE_PARAM_RHS:
		ParserRuleContext_PARAM_END
	case ANNOTATION_DECL_START:
		ParserRuleContext_ANNOTATION_KEYWORD
	case OPTIONAL_RESOURCE_ACCESS_PATH,
		RESOURCE_ACCESS_SEGMENT_RHS:
		ParserRuleContext_OPTIONAL_RESOURCE_ACCESS_METHOD
	case RESOURCE_ACCESS_PATH_SEGMENT:
		ParserRuleContext_IDENTIFIER
	case COMPUTED_SEGMENT_OR_REST_SEGMENT:
		ParserRuleContext_EXPRESSION
	case OPTIONAL_RESOURCE_ACCESS_METHOD:
		ParserRuleContext_OPTIONAL_RESOURCE_ACCESS_ACTION_ARG_LIST
	case OPTIONAL_RESOURCE_ACCESS_ACTION_ARG_LIST:
		ParserRuleContext_ACTION_END
	case OPTIONAL_TOP_LEVEL_SEMICOLON:
		ParserRuleContext_TOP_LEVEL_NODE
	case TUPLE_MEMBER:
		ParserRuleContext_TYPE_DESC_IN_TUPLE
	case RESULT_CLAUSE:
		ParserRuleContext_SELECT_CLAUSE
	case SINGLE_OR_ALTERNATE_WORKER_SEPARATOR:
		ParserRuleContext_SINGLE_OR_ALTERNATE_WORKER_END
	case XML_STEP_EXTEND:
		ParserRuleContext_XML_STEP_EXTEND_END
	case XML_STEP_START_END:
		ParserRuleContext_EXPRESSION_RHS
	case NATURAL_EXPRESSION_START:
		ParserRuleContext_NATURAL_KEYWORD
	default:
		panic("Alternative path entry not found")
	}
}

func (this *BallerinaParserErrorHandler) seekMatchInAlternativePaths(currentCtx ParserRuleContext, lookahead int, currentDepth int, matchingRulesCount int, isEntryPoint bool) Result {
	var alternativeRules []ParserRuleContext
	switch currentCtx {
	case TOP_LEVEL_NODE:
		alternativeRules = TOP_LEVEL_NODE
		break
	case TOP_LEVEL_NODE_WITHOUT_MODIFIER:
		alternativeRules = TOP_LEVEL_NODE_WITHOUT_MODIFIER
		break
	case TOP_LEVEL_NODE_WITHOUT_METADATA:
		alternativeRules = TOP_LEVEL_NODE_WITHOUT_METADATA
		break
	case FUNC_DEF_START:
		alternativeRules = FUNC_DEF_START
		break
	case FUNC_DEF_WITHOUT_FIRST_QUALIFIER:
		alternativeRules = FUNC_DEF_WITHOUT_FIRST_QUALIFIER
		break
	case FUNC_TYPE_DESC_START_WITHOUT_FIRST_QUAL:
		alternativeRules = FUNC_TYPE_DESC_START_WITHOUT_FIRST_QUAL
		break
	case FUNC_OPTIONAL_RETURNS:
		parentCtx := this.getParentContext()
		var alternatives []ParserRuleContext
		if parentCtx == ParserRuleContext_FUNC_DEF {
			grandParentCtx := this.getGrandParentContext()
			if grandParentCtx == ParserRuleContext_OBJECT_TYPE_MEMBER {
				alternatives = METHOD_DECL_OPTIONAL_RETURNS
			} else {
				alternatives = FUNC_DEF_OPTIONAL_RETURNS
			}
		} else if parentCtx == ParserRuleContext_ANON_FUNC_EXPRESSION {
			alternatives = ANNON_FUNC_OPTIONAL_RETURNS
		} else if parentCtx == ParserRuleContext_FUNC_TYPE_DESC {
			alternatives = FUNC_TYPE_OPTIONAL_RETURNS
		} else if parentCtx == ParserRuleContext_FUNC_TYPE_DESC_OR_ANON_FUNC {
			alternatives = FUNC_TYPE_OR_ANON_FUNC_OPTIONAL_RETURNS
		} else {
			alternatives = FUNC_TYPE_OR_DEF_OPTIONAL_RETURNS
		}
		alternativeRules = alternatives
		break
	case FUNC_BODY_OR_TYPE_DESC_RHS:
		alternativeRules = FUNC_BODY_OR_TYPE_DESC_RHS
		break
	case FUNC_TYPE_DESC_RHS_OR_ANON_FUNC_BODY:
		alternativeRules = FUNC_TYPE_DESC_RHS_OR_ANON_FUNC_BODY
		break
	case ANON_FUNC_BODY:
		alternativeRules = ANON_FUNC_BODY
		break
	case FUNC_BODY:
		alternativeRules = FUNC_BODY
		break
	case PARAM_LIST:
		alternativeRules = PARAM_LIST
		break
	case REQUIRED_PARAM_NAME_RHS:
		alternativeRules = REQUIRED_PARAM_NAME_RHS
		break
	case FIELD_DESCRIPTOR_RHS:
		alternativeRules = FIELD_DESCRIPTOR_RHS
		break
	case FIELD_OR_REST_DESCIPTOR_RHS:
		alternativeRules = FIELD_OR_REST_DESCIPTOR_RHS
		break
	case RECORD_BODY_END:
		alternativeRules = RECORD_BODY_END
		break
	case RECORD_BODY_START:
		alternativeRules = RECORD_BODY_START
		break
	case TYPE_DESCRIPTOR:
		if this.isInTypeDescContext() {
			panic("assertion failed")
		}
		alternativeRules = TYPE_DESCRIPTORS
		break
	case TYPE_DESC_WITHOUT_ISOLATED:
		alternativeRules = TYPE_DESCRIPTOR_WITHOUT_ISOLATED
		break
	case CLASS_DESCRIPTOR:
		alternativeRules = CLASS_DESCRIPTOR
		break
	case RECORD_FIELD_OR_RECORD_END:
		alternativeRules = RECORD_FIELD_OR_RECORD_END
		break
	case RECORD_FIELD_START:
		alternativeRules = RECORD_FIELD_START
		break
	case RECORD_FIELD_WITHOUT_METADATA:
		alternativeRules = RECORD_FIELD_WITHOUT_METADATA
		break
	case CLASS_MEMBER_OR_OBJECT_MEMBER_START:
		alternativeRules = CLASS_MEMBER_OR_OBJECT_MEMBER_START
		break
	case OBJECT_CONSTRUCTOR_MEMBER_START:
		alternativeRules = OBJECT_CONSTRUCTOR_MEMBER_START
		break
	case CLASS_MEMBER_OR_OBJECT_MEMBER_WITHOUT_META:
		alternativeRules = CLASS_MEMBER_OR_OBJECT_MEMBER_WITHOUT_META
		break
	case OBJECT_CONS_MEMBER_WITHOUT_META:
		alternativeRules = OBJECT_CONS_MEMBER_WITHOUT_META
		break
	case OPTIONAL_FIELD_INITIALIZER:
		alternativeRules = OPTIONAL_FIELD_INITIALIZER
		break
	case ON_FAIL_OPTIONAL_BINDING_PATTERN:
		alternativeRules = ON_FAIL_OPTIONAL_BINDING_PATTERN
		break
	case OBJECT_METHOD_START:
		alternativeRules = OBJECT_METHOD_START
		break
	case OBJECT_METHOD_WITHOUT_FIRST_QUALIFIER:
		alternativeRules = OBJECT_METHOD_WITHOUT_FIRST_QUALIFIER
		break
	case OBJECT_METHOD_WITHOUT_SECOND_QUALIFIER:
		alternativeRules = OBJECT_METHOD_WITHOUT_SECOND_QUALIFIER
		break
	case OBJECT_METHOD_WITHOUT_THIRD_QUALIFIER:
		alternativeRules = OBJECT_METHOD_WITHOUT_THIRD_QUALIFIER
		break
	case OBJECT_FUNC_OR_FIELD:
		alternativeRules = OBJECT_FUNC_OR_FIELD
		break
	case OBJECT_FUNC_OR_FIELD_WITHOUT_VISIBILITY:
		alternativeRules = OBJECT_FUNC_OR_FIELD_WITHOUT_VISIBILITY
		break
	case OBJECT_TYPE_START:
		alternativeRules = OBJECT_TYPE_START
		break
	case OBJECT_CONSTRUCTOR_START:
		alternativeRules = OBJECT_CONSTRUCTOR_START
		break
	case IMPORT_PREFIX_DECL:
		alternativeRules = IMPORT_PREFIX_DECL
		break
	case IMPORT_DECL_ORG_OR_MODULE_NAME_RHS:
		alternativeRules = IMPORT_DECL_ORG_OR_MODULE_NAME_RHS
		break
	case AFTER_IMPORT_MODULE_NAME:
		alternativeRules = AFTER_IMPORT_MODULE_NAME
		break
	case OPTIONAL_ABSOLUTE_PATH:
		alternativeRules = OPTIONAL_ABSOLUTE_PATH
		break
	case CONST_DECL_TYPE:
		alternativeRules = CONST_DECL_TYPE
		break
	case CONST_DECL_RHS:
		alternativeRules = CONST_DECL_RHS
		break
	case PARAMETER_START:
		alternativeRules = PARAMETER_START
		break
	case PARAMETER_START_WITHOUT_ANNOTATION:
		alternativeRules = PARAMETER_START_WITHOUT_ANNOTATION
		break
	case ANNOT_DECL_OPTIONAL_TYPE:
		alternativeRules = ANNOT_DECL_OPTIONAL_TYPE
		break
	case ANNOT_DECL_RHS:
		alternativeRules = ANNOT_DECL_RHS
		break
	case ANNOT_OPTIONAL_ATTACH_POINTS:
		alternativeRules = ANNOT_OPTIONAL_ATTACH_POINTS
		break
	case ATTACH_POINT:
		alternativeRules = ATTACH_POINT
		break
	case ATTACH_POINT_IDENT:
		alternativeRules = ATTACH_POINT_IDENT
		break
	case ATTACH_POINT_END:
		alternativeRules = ATTACH_POINT_END
		break
	case XML_NAMESPACE_PREFIX_DECL:
		alternativeRules = XML_NAMESPACE_PREFIX_DECL
		break
	case ENUM_MEMBER_START:
		alternativeRules = ENUM_MEMBER_START
		break
	case ENUM_MEMBER_RHS:
		alternativeRules = ENUM_MEMBER_RHS
		break
	case ENUM_MEMBER_END:
		alternativeRules = ENUM_MEMBER_END
		break
	case EXTERNAL_FUNC_BODY_OPTIONAL_ANNOTS:
		alternativeRules = EXTERNAL_FUNC_BODY_OPTIONAL_ANNOTS
		break
	case LIST_BP_OR_LIST_CONSTRUCTOR_MEMBER:
		alternativeRules = LIST_BP_OR_LIST_CONSTRUCTOR_MEMBER
		break
	case TUPLE_TYPE_DESC_OR_LIST_CONST_MEMBER:
		alternativeRules = TUPLE_TYPE_DESC_OR_LIST_CONST_MEMBER
		break
	case MAPPING_BP_OR_MAPPING_CONSTRUCTOR_MEMBER:
		alternativeRules = MAPPING_BP_OR_MAPPING_CONSTRUCTOR_MEMBER
		break
	case FUNC_TYPE_DESC_START,
		ANON_FUNC_EXPRESSION_START:
		alternativeRules = FUNC_TYPE_DESC_START
		break
	case MODULE_CLASS_DEFINITION_START:
		alternativeRules = MODULE_CLASS_DEFINITION_START
		break
	case CLASS_DEF_WITHOUT_FIRST_QUALIFIER:
		alternativeRules = CLASS_DEF_WITHOUT_FIRST_QUALIFIER
		break
	case CLASS_DEF_WITHOUT_SECOND_QUALIFIER:
		alternativeRules = CLASS_DEF_WITHOUT_SECOND_QUALIFIER
		break
	case CLASS_DEF_WITHOUT_THIRD_QUALIFIER:
		alternativeRules = CLASS_DEF_WITHOUT_THIRD_QUALIFIER
		break
	case OBJECT_CONSTRUCTOR_TYPE_REF:
		alternativeRules = OBJECT_CONSTRUCTOR_RHS
		break
	case OBJECT_FIELD_QUALIFIER:
		alternativeRules = OBJECT_FIELD_QUALIFIER
		break
	case CONFIG_VAR_DECL_RHS:
		alternativeRules = CONFIG_VAR_DECL_RHS
		break
	case OPTIONAL_SERVICE_DECL_TYPE:
		alternativeRules = OPTIONAL_SERVICE_DECL_TYPE
		break
	case SERVICE_IDENT_RHS:
		alternativeRules = SERVICE_IDENT_RHS
		break
	case ABSOLUTE_RESOURCE_PATH_START:
		alternativeRules = ABSOLUTE_RESOURCE_PATH_START
		break
	case ABSOLUTE_RESOURCE_PATH_END:
		alternativeRules = ABSOLUTE_RESOURCE_PATH_END
		break
	case SERVICE_DECL_OR_VAR_DECL:
		alternativeRules = SERVICE_DECL_OR_VAR_DECL
		break
	case OPTIONAL_RELATIVE_PATH:
		alternativeRules = OPTIONAL_RELATIVE_PATH
		break
	case RELATIVE_RESOURCE_PATH_START:
		alternativeRules = RELATIVE_RESOURCE_PATH_START
		break
	case RESOURCE_PATH_SEGMENT:
		alternativeRules = RESOURCE_PATH_SEGMENT
		break
	case PATH_PARAM_OPTIONAL_ANNOTS:
		alternativeRules = PATH_PARAM_OPTIONAL_ANNOTS
		break
	case PATH_PARAM_ELLIPSIS:
		alternativeRules = PATH_PARAM_ELLIPSIS
		break
	case OPTIONAL_PATH_PARAM_NAME:
		alternativeRules = OPTIONAL_PATH_PARAM_NAME
		break
	case RELATIVE_RESOURCE_PATH_END:
		alternativeRules = RELATIVE_RESOURCE_PATH_END
		break
	case SERVICE_DECL_START:
		alternativeRules = SERVICE_DECL_START
		break
	case OPTIONAL_TYPE_PARAMETER:
		alternativeRules = OPTIONAL_TYPE_PARAMETER
		break
	case MAP_TYPE_OR_TYPE_REF:
		alternativeRules = MAP_TYPE_OR_TYPE_REF
		break
	case OBJECT_TYPE_OR_TYPE_REF:
		alternativeRules = OBJECT_TYPE_OR_TYPE_REF
		break
	case STREAM_TYPE_OR_TYPE_REF:
		alternativeRules = STREAM_TYPE_OR_TYPE_REF
		break
	case TABLE_TYPE_OR_TYPE_REF:
		alternativeRules = TABLE_TYPE_OR_TYPE_REF
		break
	case PARAMETERIZED_TYPE_OR_TYPE_REF:
		alternativeRules = PARAMETERIZED_TYPE_OR_TYPE_REF
		break
	case TYPE_DESC_RHS_OR_TYPE_REF:
		alternativeRules = TYPE_DESC_RHS_OR_TYPE_REF
		break
	case TRANSACTION_STMT_RHS_OR_TYPE_REF:
		alternativeRules = TRANSACTION_STMT_RHS_OR_TYPE_REF
		break
	case TABLE_CONS_OR_QUERY_EXPR_OR_VAR_REF:
		alternativeRules = TABLE_CONS_OR_QUERY_EXPR_OR_VAR_REF
		break
	case QUERY_EXPR_OR_VAR_REF:
		alternativeRules = QUERY_EXPR_OR_VAR_REF
		break
	case ERROR_CONS_EXPR_OR_VAR_REF:
		alternativeRules = ERROR_CONS_EXPR_OR_VAR_REF
		break
	case QUALIFIED_IDENTIFIER:
		alternativeRules = QUALIFIED_IDENTIFIER
		break
	case MODULE_VAR_DECL_START:
		alternativeRules = MODULE_VAR_DECL_START
		break
	case MODULE_VAR_WITHOUT_FIRST_QUAL:
		alternativeRules = MODULE_VAR_WITHOUT_FIRST_QUAL
		break
	case MODULE_VAR_WITHOUT_SECOND_QUAL:
		alternativeRules = MODULE_VAR_WITHOUT_SECOND_QUAL
		break
	case OBJECT_TYPE_WITHOUT_FIRST_QUALIFIER:
		alternativeRules = OBJECT_TYPE_WITHOUT_FIRST_QUALIFIER
		break
	case FUNC_DEF_OR_TYPE_DESC_RHS:
		alternativeRules = FUNC_DEF_OR_TYPE_DESC_RHS
		break
	case EXPR_START_OR_INFERRED_TYPEDESC_DEFAULT_START:
		alternativeRules = EXPR_START_OR_INFERRED_TYPEDESC_DEFAULT_START
		break
	case TYPE_CAST_PARAM_START_OR_INFERRED_TYPEDESC_DEFAULT_END:
		alternativeRules = TYPE_CAST_PARAM_START_OR_INFERRED_TYPEDESC_DEFAULT_END
		break
	case END_OF_PARAMS_OR_NEXT_PARAM_START:
		alternativeRules = END_OF_PARAMS_OR_NEXT_PARAM_START
		break
	case PARAM_START:
		alternativeRules = PARAM_START
		break
	case PARAM_RHS:
		alternativeRules = PARAM_RHS
		break
	case FUNC_TYPE_PARAM_RHS:
		alternativeRules = FUNC_TYPE_PARAM_RHS
		break
	case ANNOTATION_DECL_START:
		alternativeRules = ANNOTATION_DECL_START
		break
	case OPTIONAL_TOP_LEVEL_SEMICOLON:
		alternativeRules = OPTIONAL_TOP_LEVEL_SEMICOLON
		break
	case TUPLE_MEMBER:
		alternativeRules = TUPLE_MEMBER
		break
	default:
		return this.seekMatchInStmtRelatedAlternativePaths(currentCtx, lookahead, currentDepth, matchingRulesCount,
			isEntryPoint)
	}
	return this.seekInAlternativesPaths(lookahead, currentDepth, matchingRulesCount, alternativeRules, isEntryPoint)
}

func (this *BallerinaParserErrorHandler) seekMatchInStmtRelatedAlternativePaths(currentCtx ParserRuleContext, lookahead int, currentDepth int, matchingRulesCount int, isEntryPoint bool) Result {
	var alternativeRules []ParserRuleContext
	switch currentCtx {
	case VAR_DECL_STMT_RHS:
		alternativeRules = VAR_DECL_RHS
		break
	case STATEMENT,
		STATEMENT_WITHOUT_ANNOTS:
		return this.seekInStatements(currentCtx, lookahead, currentDepth, matchingRulesCount, isEntryPoint)
	case TYPE_NAME_OR_VAR_NAME:
		alternativeRules = TYPE_OR_VAR_NAME
		break
	case ELSE_BLOCK:
		alternativeRules = ELSE_BLOCK
		break
	case ELSE_BODY:
		alternativeRules = ELSE_BODY
		break
	case CALL_STMT_START:
		alternativeRules = CALL_STATEMENT
		break
	case RETURN_STMT_RHS:
		alternativeRules = RETURN_RHS
		break
	case ARRAY_LENGTH:
		alternativeRules = ARRAY_LENGTH
		break
	case STMT_START_WITH_EXPR_RHS:
		alternativeRules = STMT_START_WITH_EXPR_RHS
		break
	case EXPR_STMT_RHS:
		alternativeRules = EXPR_STMT_RHS
		break
	case EXPRESSION_STATEMENT_START:
		alternativeRules = EXPRESSION_STATEMENT_START
		break
	case TYPE_DESC_RHS:
		if this.isInTypeDescContext() {
			panic("assertion failed")
		}
		alternativeRules = TYPE_DESC_RHS
		break
	case STREAM_TYPE_FIRST_PARAM_RHS:
		alternativeRules = STREAM_TYPE_FIRST_PARAM_RHS
		break
	case FUNCTION_KEYWORD_RHS:
		alternativeRules = FUNCTION_KEYWORD_RHS
		break
	case FUNC_TYPE_FUNC_KEYWORD_RHS_START:
		alternativeRules = FUNC_TYPE_FUNC_KEYWORD_RHS_START
		break
	case WORKER_NAME_RHS:
		alternativeRules = WORKER_NAME_RHS
		break
	case BINDING_PATTERN:
		alternativeRules = BINDING_PATTERN
		break
	case LIST_BINDING_PATTERNS_START:
		alternativeRules = LIST_BINDING_PATTERNS_START
		break
	case LIST_BINDING_PATTERN_MEMBER_END:
		alternativeRules = LIST_BINDING_PATTERN_MEMBER_END
		break
	case LIST_BINDING_PATTERN_MEMBER:
		alternativeRules = LIST_BINDING_PATTERN_CONTENTS
		break
	case MAPPING_BINDING_PATTERN_END:
		alternativeRules = MAPPING_BINDING_PATTERN_END
		break
	case FIELD_BINDING_PATTERN_END:
		alternativeRules = FIELD_BINDING_PATTERN_END
		break
	case MAPPING_BINDING_PATTERN_MEMBER:
		alternativeRules = MAPPING_BINDING_PATTERN_MEMBER
		break
	case ERROR_BINDING_PATTERN_ERROR_KEYWORD_RHS:
		alternativeRules = ERROR_BINDING_PATTERN_ERROR_KEYWORD_RHS
		break
	case ERROR_ARG_LIST_BINDING_PATTERN_START:
		alternativeRules = ERROR_ARG_LIST_BINDING_PATTERN_START
		break
	case ERROR_MESSAGE_BINDING_PATTERN_END:
		alternativeRules = ERROR_MESSAGE_BINDING_PATTERN_END
		break
	case ERROR_MESSAGE_BINDING_PATTERN_RHS:
		alternativeRules = ERROR_MESSAGE_BINDING_PATTERN_RHS
		break
	case ERROR_FIELD_BINDING_PATTERN:
		alternativeRules = ERROR_FIELD_BINDING_PATTERN
		break
	case ERROR_FIELD_BINDING_PATTERN_END:
		alternativeRules = ERROR_FIELD_BINDING_PATTERN_END
		break
	case KEY_CONSTRAINTS_RHS:
		alternativeRules = KEY_CONSTRAINTS_RHS
		break
	case TABLE_TYPE_DESC_RHS:
		alternativeRules = TABLE_TYPE_DESC_RHS
		break
	case TYPE_DESC_IN_TUPLE_RHS:
		alternativeRules = TYPE_DESC_IN_TUPLE_RHS
		break
	case TUPLE_TYPE_MEMBER_RHS:
		alternativeRules = TUPLE_TYPE_MEMBER_RHS
		break
	case LIST_CONSTRUCTOR_MEMBER_END:
		alternativeRules = LIST_CONSTRUCTOR_MEMBER_END
		break
	case NIL_OR_PARENTHESISED_TYPE_DESC_RHS:
		alternativeRules = NIL_OR_PARENTHESISED_TYPE_DESC_RHS
		break
	case REMOTE_OR_RESOURCE_CALL_OR_ASYNC_SEND_RHS:
		alternativeRules = REMOTE_OR_RESOURCE_CALL_OR_ASYNC_SEND_RHS
		break
	case REMOTE_CALL_OR_ASYNC_SEND_END:
		alternativeRules = REMOTE_CALL_OR_ASYNC_SEND_END
		break
	case OPTIONAL_RESOURCE_ACCESS_PATH:
		alternativeRules = OPTIONAL_RESOURCE_ACCESS_PATH
		break
	case RESOURCE_ACCESS_PATH_SEGMENT:
		alternativeRules = RESOURCE_ACCESS_PATH_SEGMENT
		break
	case COMPUTED_SEGMENT_OR_REST_SEGMENT:
		alternativeRules = COMPUTED_SEGMENT_OR_REST_SEGMENT
		break
	case RESOURCE_ACCESS_SEGMENT_RHS:
		alternativeRules = RESOURCE_ACCESS_SEGMENT_RHS
		break
	case OPTIONAL_RESOURCE_ACCESS_METHOD:
		alternativeRules = OPTIONAL_RESOURCE_ACCESS_METHOD
		break
	case OPTIONAL_RESOURCE_ACCESS_ACTION_ARG_LIST:
		alternativeRules = OPTIONAL_RESOURCE_ACCESS_ACTION_ARG_LIST
		break
	case RECEIVE_WORKERS:
		alternativeRules = RECEIVE_WORKERS
		break
	case RECEIVE_FIELD:
		alternativeRules = RECEIVE_FIELD
		break
	case RECEIVE_FIELD_END:
		alternativeRules = RECEIVE_FIELD_END
		break
	case WAIT_KEYWORD_RHS:
		alternativeRules = WAIT_KEYWORD_RHS
		break
	case WAIT_FIELD_NAME_RHS:
		alternativeRules = WAIT_FIELD_NAME_RHS
		break
	case WAIT_FIELD_END:
		alternativeRules = WAIT_FIELD_END
		break
	case WAIT_FUTURE_EXPR_END:
		alternativeRules = WAIT_FUTURE_EXPR_END
		break
	case OPTIONAL_PEER_WORKER:
		alternativeRules = OPTIONAL_PEER_WORKER
		break
	case ROLLBACK_RHS:
		alternativeRules = ROLLBACK_RHS
		break
	case RETRY_KEYWORD_RHS:
		alternativeRules = RETRY_KEYWORD_RHS
		break
	case RETRY_TYPE_PARAM_RHS:
		alternativeRules = RETRY_TYPE_PARAM_RHS
		break
	case RETRY_BODY:
		alternativeRules = RETRY_BODY
		break
	case STMT_START_BRACKETED_LIST_MEMBER:
		alternativeRules = LIST_BP_OR_TUPLE_TYPE_MEMBER
		break
	case STMT_START_BRACKETED_LIST_RHS:
		alternativeRules = LIST_BP_OR_TUPLE_TYPE_DESC_RHS
		break
	case BRACKETED_LIST_MEMBER_END:
		alternativeRules = BRACKETED_LIST_MEMBER_END
		break
	case BRACKETED_LIST_MEMBER:
		alternativeRules = BRACKETED_LIST_MEMBER
		break
	case BRACKETED_LIST_RHS,
		BINDING_PATTERN_OR_EXPR_RHS,
		TYPE_DESC_OR_EXPR_RHS:
		alternativeRules = BRACKETED_LIST_RHS
		break
	case BINDING_PATTERN_OR_VAR_REF_RHS:
		alternativeRules = BINDING_PATTERN_OR_VAR_REF_RHS
		break
	case TYPE_DESC_RHS_OR_BP_RHS:
		alternativeRules = TYPE_DESC_RHS_OR_BP_RHS
		break
	case LIST_BINDING_MEMBER_OR_ARRAY_LENGTH:
		alternativeRules = LIST_BINDING_MEMBER_OR_ARRAY_LENGTH
		break
	case MATCH_PATTERN_LIST_MEMBER_RHS:
		alternativeRules = MATCH_PATTERN_LIST_MEMBER_RHS
		break
	case MATCH_PATTERN_START:
		alternativeRules = MATCH_PATTERN_START
		break
	case LIST_MATCH_PATTERNS_START:
		alternativeRules = LIST_MATCH_PATTERNS_START
		break
	case LIST_MATCH_PATTERN_MEMBER:
		alternativeRules = LIST_MATCH_PATTERN_MEMBER
		break
	case LIST_MATCH_PATTERN_MEMBER_RHS:
		alternativeRules = LIST_MATCH_PATTERN_MEMBER_RHS
		break
	case FIELD_MATCH_PATTERNS_START:
		alternativeRules = FIELD_MATCH_PATTERNS_START
		break
	case FIELD_MATCH_PATTERN_MEMBER:
		alternativeRules = FIELD_MATCH_PATTERN_MEMBER
		break
	case FIELD_MATCH_PATTERN_MEMBER_RHS:
		alternativeRules = FIELD_MATCH_PATTERN_MEMBER_RHS
		break
	case ERROR_MATCH_PATTERN_OR_CONST_PATTERN:
		alternativeRules = ERROR_MATCH_PATTERN_OR_CONST_PATTERN
		break
	case ERROR_MATCH_PATTERN_ERROR_KEYWORD_RHS:
		alternativeRules = ERROR_MATCH_PATTERN_ERROR_KEYWORD_RHS
		break
	case ERROR_ARG_LIST_MATCH_PATTERN_START:
		alternativeRules = ERROR_ARG_LIST_MATCH_PATTERN_START
		break
	case ERROR_MESSAGE_MATCH_PATTERN_END:
		alternativeRules = ERROR_MESSAGE_MATCH_PATTERN_END
		break
	case ERROR_MESSAGE_MATCH_PATTERN_RHS:
		alternativeRules = ERROR_MESSAGE_MATCH_PATTERN_RHS
		break
	case ERROR_FIELD_MATCH_PATTERN:
		alternativeRules = ERROR_FIELD_MATCH_PATTERN
		break
	case ERROR_FIELD_MATCH_PATTERN_RHS:
		alternativeRules = ERROR_FIELD_MATCH_PATTERN_RHS
		break
	case NAMED_ARG_MATCH_PATTERN_RHS:
		alternativeRules = NAMED_ARG_MATCH_PATTERN_RHS
		break
	case JOIN_CLAUSE_START:
		alternativeRules = JOIN_CLAUSE_START
		break
	case INTERMEDIATE_CLAUSE_START:
		alternativeRules = INTERMEDIATE_CLAUSE_START
		break
	case REGULAR_COMPOUND_STMT_RHS:
		alternativeRules = REGULAR_COMPOUND_STMT_RHS
		break
	case NAMED_WORKER_DECL_START:
		alternativeRules = NAMED_WORKER_DECL_START
		break
	case ASSIGNMENT_STMT_RHS:
		alternativeRules = ASSIGNMENT_STMT_RHS
		break
	default:
		return this.seekMatchInExprRelatedAlternativePaths(currentCtx, lookahead, currentDepth, matchingRulesCount,
			isEntryPoint)
	}
	return this.seekInAlternativesPaths(lookahead, currentDepth, matchingRulesCount, alternativeRules, isEntryPoint)
}

func (this *BallerinaParserErrorHandler) seekMatchInExprRelatedAlternativePaths(currentCtx ParserRuleContext, lookahead int, currentDepth int, matchingRulesCount int, isEntryPoint bool) Result {
	var alternativeRules []ParserRuleContext
	switch currentCtx {
	case EXPRESSION,
		TERMINAL_EXPRESSION:
		alternativeRules = EXPRESSION_START
		break
	case ARG_START:
		alternativeRules = ARG_START
		break
	case ARG_START_OR_ARG_LIST_END:
		alternativeRules = ARG_START_OR_ARG_LIST_END
		break
	case NAMED_OR_POSITIONAL_ARG_RHS:
		alternativeRules = NAMED_OR_POSITIONAL_ARG_RHS
		break
	case ARG_END:
		alternativeRules = ARG_END
		break
	case ACCESS_EXPRESSION:
		return this.seekInAccessExpression(currentCtx, lookahead, currentDepth, matchingRulesCount, isEntryPoint)
	case FIRST_MAPPING_FIELD:
		alternativeRules = FIRST_MAPPING_FIELD_START
		break
	case MAPPING_FIELD:
		alternativeRules = MAPPING_FIELD_START
		break
	case SPECIFIC_FIELD:
		alternativeRules = SPECIFIC_FIELD
		break
	case SPECIFIC_FIELD_RHS:
		alternativeRules = SPECIFIC_FIELD_RHS
		break
	case MAPPING_FIELD_END:
		alternativeRules = MAPPING_FIELD_END
		break
	case LET_VAR_DECL_START:
		alternativeRules = LET_VAR_DECL_START
		break
	case ORDER_KEY_LIST_END:
		alternativeRules = ORDER_KEY_LIST_END
		break
	case GROUPING_KEY_LIST_ELEMENT:
		alternativeRules = GROUPING_KEY_LIST_ELEMENT
		break
	case GROUPING_KEY_LIST_ELEMENT_END:
		alternativeRules = GROUPING_KEY_LIST_ELEMENT_END
		break
	case TEMPLATE_MEMBER:
		alternativeRules = TEMPLATE_MEMBER
		break
	case TEMPLATE_STRING_RHS:
		alternativeRules = TEMPLATE_STRING_RHS
		break
	case CONSTANT_EXPRESSION_START:
		alternativeRules = CONSTANT_EXPRESSION
		break
	case LIST_CONSTRUCTOR_FIRST_MEMBER:
		alternativeRules = LIST_CONSTRUCTOR_FIRST_MEMBER
		break
	case LIST_CONSTRUCTOR_MEMBER:
		alternativeRules = LIST_CONSTRUCTOR_MEMBER
		break
	case TYPE_CAST_PARAM:
		alternativeRules = TYPE_CAST_PARAM
		break
	case TYPE_CAST_PARAM_RHS:
		alternativeRules = TYPE_CAST_PARAM_RHS
		break
	case TABLE_KEYWORD_RHS:
		alternativeRules = TABLE_KEYWORD_RHS
		break
	case ROW_LIST_RHS:
		alternativeRules = ROW_LIST_RHS
		break
	case TABLE_ROW_END:
		alternativeRules = TABLE_ROW_END
		break
	case KEY_SPECIFIER_RHS:
		alternativeRules = KEY_SPECIFIER_RHS
		break
	case TABLE_KEY_RHS:
		alternativeRules = TABLE_KEY_RHS
		break
	case NEW_KEYWORD_RHS:
		alternativeRules = NEW_KEYWORD_RHS
		break
	case TABLE_CONSTRUCTOR_OR_QUERY_START:
		alternativeRules = TABLE_CONSTRUCTOR_OR_QUERY_START
		break
	case TABLE_CONSTRUCTOR_OR_QUERY_RHS:
		alternativeRules = TABLE_CONSTRUCTOR_OR_QUERY_RHS
		break
	case QUERY_PIPELINE_RHS:
		alternativeRules = QUERY_PIPELINE_RHS
		break
	case BRACED_EXPR_OR_ANON_FUNC_PARAM_RHS,
		ANON_FUNC_PARAM_RHS:
		alternativeRules = BRACED_EXPR_OR_ANON_FUNC_PARAM_RHS
		break
	case PARAM_END:
		alternativeRules = PARAM_END
		break
	case ANNOTATION_REF_RHS:
		alternativeRules = ANNOTATION_REF_RHS
		break
	case INFER_PARAM_END_OR_PARENTHESIS_END:
		alternativeRules = INFER_PARAM_END_OR_PARENTHESIS_END
		break
	case XML_NAVIGATE_EXPR:
		alternativeRules = XML_NAVIGATE_EXPR
		break
	case XML_NAME_PATTERN_RHS:
		alternativeRules = XML_NAME_PATTERN_RHS
		break
	case XML_ATOMIC_NAME_PATTERN_START:
		alternativeRules = XML_ATOMIC_NAME_PATTERN_START
		break
	case XML_ATOMIC_NAME_IDENTIFIER_RHS:
		alternativeRules = XML_ATOMIC_NAME_IDENTIFIER_RHS
		break
	case XML_STEP_START:
		alternativeRules = XML_STEP_START
		break
	case XML_STEP_EXTEND:
		alternativeRules = XML_STEP_EXTEND
		break
	case XML_STEP_START_END:
		alternativeRules = XML_STEP_START_END
		break
	case OPTIONAL_MATCH_GUARD:
		alternativeRules = OPTIONAL_MATCH_GUARD
		break
	case MEMBER_ACCESS_KEY_EXPR_END:
		alternativeRules = MEMBER_ACCESS_KEY_EXPR_END
		break
	case LISTENERS_LIST_END:
		alternativeRules = LISTENERS_LIST_END
		break
	case OBJECT_CONS_WITHOUT_FIRST_QUALIFIER:
		alternativeRules = OBJECT_CONS_WITHOUT_FIRST_QUALIFIER
		break
	case RESULT_CLAUSE:
		alternativeRules = RESULT_CLAUSE
		break
	case EXPRESSION_RHS:
		return this.seekMatchInExpressionRhs(lookahead, currentDepth, matchingRulesCount, isEntryPoint, false)
	case VARIABLE_REF_RHS:
		return this.seekMatchInExpressionRhs(lookahead, currentDepth, matchingRulesCount, isEntryPoint, true)
	case ERROR_CONSTRUCTOR_RHS:
		alternativeRules = ERROR_CONSTRUCTOR_RHS
		break
	case SINGLE_OR_ALTERNATE_WORKER_SEPARATOR:
		alternativeRules = SINGLE_OR_ALTERNATE_WORKER_SEPARATOR
		break
	case NATURAL_EXPRESSION_START:
		alternativeRules = NATURAL_EXPRESSION_START
		break
	case OPTIONAL_PARENTHESIZED_ARG_LIST:
		alternativeRules = OPTIONAL_PARENTHESIZED_ARG_LIST
		break
	default:
		panic("seekMatchInExprRelatedAlternativePaths found: " + currentCtx)
	}
	return this.seekInAlternativesPaths(lookahead, currentDepth, matchingRulesCount, alternativeRules, isEntryPoint)
}

func (this *BallerinaParserErrorHandler) seekInStatements(currentCtx ParserRuleContext, lookahead int, currentDepth int, currentMatches int, isEntryPoint bool) Result {
	nextToken := this.this.tokenReader.peek(lookahead)
	if nextToken.kind == SyntaxKind_SEMICOLON_TOKEN {
		result := this.seekMatchInSubTree(ParserRuleContext.STATEMENT, lookahead+1, currentDepth+1,
			isEntryPoint)
		this.result.pushFix(NewSolution(ACTION_REMOVE, currentCtx, nextToken.kind, nextToken.toString(), currentDepth))
		return this.getFinalResult(currentMatches, result)
	}
	return this.seekInAlternativesPaths(lookahead, currentDepth, currentMatches, STATEMENTS, isEntryPoint)
}

func (this *BallerinaParserErrorHandler) seekInAccessExpression(currentCtx ParserRuleContext, lookahead int, currentDepth int, currentMatches int, isEntryPoint bool) Result {
	nextToken := this.this.tokenReader.peek(lookahead)
	currentDepth++
	if nextToken.kind != SyntaxKind_IDENTIFIER_TOKEN {
		return this.fixAndContinue(currentCtx, lookahead, currentDepth, currentMatches, isEntryPoint)
	}
	var nextContext ParserRuleContext
	nextNextToken := this.this.tokenReader.peek(lookahead + 1)
	switch nextNextToken.kind {
	case OPEN_PAREN_TOKEN:
		nextContext = ParserRuleContext_OPEN_PARENTHESIS
	case DOT_TOKEN:
		nextContext = ParserRuleContext_DOT
	case OPEN_BRACKET_TOKEN:
		nextContext = ParserRuleContext_MEMBER_ACCESS_KEY_EXPR
	default:
		nextContext = this.getNextRuleForExpr()
	}
	currentMatches++
	lookahead++
	result := this.seekMatch(nextContext, lookahead, currentDepth, isEntryPoint)
	return this.getFinalResult(currentMatches, result)
}

func (this *BallerinaParserErrorHandler) seekMatchInExpressionRhs(lookahead int, currentDepth int, currentMatches int, isEntryPoint bool, allowFuncCall bool) Result {
	parentCtx := this.getParentContext()
	alternatives := nil
	switch parentCtx {
	case ARG_LIST:
		alternatives = []ParserRuleContext{ParserRuleContext_COMMA, ParserRuleContext_BINARY_OPERATOR, ParserRuleContext_DOT, ParserRuleContext_ANNOT_CHAINING_TOKEN, ParserRuleContext_OPTIONAL_CHAINING_TOKEN, ParserRuleContext_CONDITIONAL_EXPRESSION, ParserRuleContext_XML_NAVIGATE_EXPR, ParserRuleContext_MEMBER_ACCESS_KEY_EXPR, ParserRuleContext_ARG_LIST_END}
		break
	case MAPPING_CONSTRUCTOR,
		MULTI_WAIT_FIELDS,
		MAPPING_BP_OR_MAPPING_CONSTRUCTOR:
		alternatives = []ParserRuleContext{ParserRuleContext_CLOSE_BRACE, ParserRuleContext_COMMA, ParserRuleContext_BINARY_OPERATOR, ParserRuleContext_DOT, ParserRuleContext_ANNOT_CHAINING_TOKEN, ParserRuleContext_OPTIONAL_CHAINING_TOKEN, ParserRuleContext_CONDITIONAL_EXPRESSION, ParserRuleContext_XML_NAVIGATE_EXPR, ParserRuleContext_MEMBER_ACCESS_KEY_EXPR}
		break
	case COMPUTED_FIELD_NAME:
		alternatives = []ParserRuleContext{ParserRuleContext_CLOSE_BRACKET, ParserRuleContext_BINARY_OPERATOR, ParserRuleContext_DOT, ParserRuleContext_ANNOT_CHAINING_TOKEN, ParserRuleContext_OPTIONAL_CHAINING_TOKEN, ParserRuleContext_CONDITIONAL_EXPRESSION, ParserRuleContext_XML_NAVIGATE_EXPR, ParserRuleContext_MEMBER_ACCESS_KEY_EXPR, ParserRuleContext_OPEN_BRACKET}
		break
	case LISTENERS_LIST:
		alternatives = []ParserRuleContext{ParserRuleContext_LISTENERS_LIST_END, ParserRuleContext_BINARY_OPERATOR, ParserRuleContext_DOT, ParserRuleContext_ANNOT_CHAINING_TOKEN, ParserRuleContext_OPTIONAL_CHAINING_TOKEN, ParserRuleContext_CONDITIONAL_EXPRESSION, ParserRuleContext_XML_NAVIGATE_EXPR, ParserRuleContext_MEMBER_ACCESS_KEY_EXPR}
		break
	case LIST_CONSTRUCTOR,
		MEMBER_ACCESS_KEY_EXPR,
		BRACKETED_LIST,
		STMT_START_BRACKETED_LIST:
		alternatives = []ParserRuleContext{ParserRuleContext_COMMA, ParserRuleContext_BINARY_OPERATOR, ParserRuleContext_DOT, ParserRuleContext_ANNOT_CHAINING_TOKEN, ParserRuleContext_OPTIONAL_CHAINING_TOKEN, ParserRuleContext_CONDITIONAL_EXPRESSION, ParserRuleContext_XML_NAVIGATE_EXPR, ParserRuleContext_MEMBER_ACCESS_KEY_EXPR, ParserRuleContext_CLOSE_BRACKET}
		break
	case LET_EXPR_LET_VAR_DECL:
		alternatives = []ParserRuleContext{ParserRuleContext_IN_KEYWORD, ParserRuleContext_COMMA, ParserRuleContext_BINARY_OPERATOR, ParserRuleContext_DOT, ParserRuleContext_ANNOT_CHAINING_TOKEN, ParserRuleContext_OPTIONAL_CHAINING_TOKEN, ParserRuleContext_CONDITIONAL_EXPRESSION, ParserRuleContext_XML_NAVIGATE_EXPR, ParserRuleContext_MEMBER_ACCESS_KEY_EXPR}
		break
	case LET_CLAUSE_LET_VAR_DECL:
		alternatives = []ParserRuleContext{ParserRuleContext_COMMA, ParserRuleContext_BINARY_OPERATOR, ParserRuleContext_DOT, ParserRuleContext_ANNOT_CHAINING_TOKEN, ParserRuleContext_OPTIONAL_CHAINING_TOKEN, ParserRuleContext_CONDITIONAL_EXPRESSION, ParserRuleContext_XML_NAVIGATE_EXPR, ParserRuleContext_MEMBER_ACCESS_KEY_EXPR, ParserRuleContext_LET_CLAUSE_END}
		break
	case ORDER_KEY_LIST:
		alternatives = []ParserRuleContext{ParserRuleContext_ORDER_DIRECTION, ParserRuleContext_ORDER_KEY_LIST_END, ParserRuleContext_BINARY_OPERATOR, ParserRuleContext_DOT, ParserRuleContext_ANNOT_CHAINING_TOKEN, ParserRuleContext_OPTIONAL_CHAINING_TOKEN, ParserRuleContext_CONDITIONAL_EXPRESSION, ParserRuleContext_XML_NAVIGATE_EXPR, ParserRuleContext_MEMBER_ACCESS_KEY_EXPR}
		break
	case GROUP_BY_CLAUSE:
		alternatives = []ParserRuleContext{ParserRuleContext_GROUPING_KEY_LIST_ELEMENT_END, ParserRuleContext_BINARY_OPERATOR, ParserRuleContext_DOT, ParserRuleContext_ANNOT_CHAINING_TOKEN, ParserRuleContext_OPTIONAL_CHAINING_TOKEN, ParserRuleContext_CONDITIONAL_EXPRESSION, ParserRuleContext_XML_NAVIGATE_EXPR, ParserRuleContext_MEMBER_ACCESS_KEY_EXPR}
		break
	case QUERY_EXPRESSION:
		alternatives = []ParserRuleContext{ParserRuleContext_QUERY_PIPELINE_RHS, ParserRuleContext_BINARY_OPERATOR, ParserRuleContext_DOT, ParserRuleContext_ANNOT_CHAINING_TOKEN, ParserRuleContext_OPTIONAL_CHAINING_TOKEN, ParserRuleContext_CONDITIONAL_EXPRESSION, ParserRuleContext_XML_NAVIGATE_EXPR, ParserRuleContext_MEMBER_ACCESS_KEY_EXPR}
		break
	default:
		if this.isParameter(parentCtx) {
			alternatives = []ParserRuleContext{ParserRuleContext_CLOSE_PARENTHESIS, ParserRuleContext_BINARY_OPERATOR, ParserRuleContext_DOT, ParserRuleContext_ANNOT_CHAINING_TOKEN, ParserRuleContext_OPTIONAL_CHAINING_TOKEN, ParserRuleContext_CONDITIONAL_EXPRESSION, ParserRuleContext_XML_NAVIGATE_EXPR, ParserRuleContext_MEMBER_ACCESS_KEY_EXPR, ParserRuleContext_COMMA}
		}
		break
	}
	if alternatives != nil {
		if allowFuncCall {
			alternatives = this.modifyAlternativesWithArgListStart(alternatives)
		}
		return this.seekInAlternativesPaths(lookahead, currentDepth, currentMatches, alternatives, isEntryPoint)
	}
	var nextContext ParserRuleContext
	if ((parentCtx == ParserRuleContext_IF_BLOCK) || (parentCtx == ParserRuleContext_WHILE_BLOCK)) || (parentCtx == ParserRuleContext_FOREACH_STMT) {
		nextContext = ParserRuleContext_BLOCK_STMT
	} else if parentCtx == ParserRuleContext_MATCH_STMT {
		nextContext = ParserRuleContext_MATCH_BODY
	} else if parentCtx == ParserRuleContext_CALL_STMT {
		nextContext = ParserRuleContext_METHOD_CALL_DOT
	} else if (((((this.isStatement(parentCtx) || (parentCtx == ParserRuleContext_RECORD_FIELD)) || (parentCtx == ParserRuleContext_OBJECT_CONSTRUCTOR_MEMBER)) || (parentCtx == ParserRuleContext_CLASS_MEMBER)) || (parentCtx == ParserRuleContext_OBJECT_TYPE_MEMBER)) || (parentCtx == ParserRuleContext_LISTENER_DECL)) || (parentCtx == ParserRuleContext_CONSTANT_DECL) {
		nextContext = ParserRuleContext_SEMICOLON
	} else if parentCtx == ParserRuleContext_ANNOTATIONS {
		nextContext = ParserRuleContext_ANNOTATION_END
	} else if parentCtx == ParserRuleContext_INTERPOLATION {
		nextContext = ParserRuleContext_CLOSE_BRACE
	} else if (parentCtx == ParserRuleContext_BRACED_EXPRESSION) || (parentCtx == ParserRuleContext_BRACED_EXPR_OR_ANON_FUNC_PARAMS) {
		nextContext = ParserRuleContext_CLOSE_PARENTHESIS
	} else if parentCtx == ParserRuleContext_FUNC_DEF {
		nextContext = ParserRuleContext_SEMICOLON
	} else if parentCtx == ParserRuleContext_ALTERNATE_WAIT_EXPRS {
		nextContext = ParserRuleContext_ALTERNATE_WAIT_EXPR_LIST_END
	} else if parentCtx == ParserRuleContext_CONDITIONAL_EXPRESSION {
		nextContext = ParserRuleContext_COLON
	} else if parentCtx == ParserRuleContext_ENUM_MEMBER_LIST {
		nextContext = ParserRuleContext_ENUM_MEMBER_END
	} else if parentCtx == ParserRuleContext_MATCH_BODY {
		nextContext = ParserRuleContext_RIGHT_DOUBLE_ARROW
	} else if (parentCtx == ParserRuleContext_SELECT_CLAUSE) || (parentCtx == ParserRuleContext_COLLECT_CLAUSE) {
		nextToken := this.this.tokenReader.peek(lookahead)
		switch nextToken.kind {
		case ON_KEYWORD, CONFLICT_KEYWORD:
			nextContext = ParserRuleContext_ON_CONFLICT_CLAUSE
		default:
			nextContext = ParserRuleContext_QUERY_EXPRESSION_END
		}
	} else if parentCtx == ParserRuleContext_JOIN_CLAUSE {
		nextContext = ParserRuleContext_ON_CLAUSE
	} else if parentCtx == ParserRuleContext_ON_CLAUSE {
		nextContext = ParserRuleContext_EQUALS_KEYWORD
	} else if parentCtx == ParserRuleContext_CLIENT_RESOURCE_ACCESS_ACTION {
		nextContext = ParserRuleContext_CLOSE_BRACKET
	} else {
		panic("seekMatchInExpressionRhs found: " + parentCtx)
	}
	alternatives = this.getExpressionRhsAlternatives(nextContext)
	if allowFuncCall {
		alternatives = this.modifyAlternativesWithArgListStart(alternatives)
	}
	return this.seekInAlternativesPaths(lookahead, currentDepth, currentMatches, alternatives, isEntryPoint)
}

func (this *BallerinaParserErrorHandler) getExpressionRhsAlternatives(nextContext ParserRuleContext) []ParserRuleContext {
	if ((nextContext == ParserRuleContext_SEMICOLON) || (nextContext == ParserRuleContext_QUERY_EXPRESSION_END)) || (nextContext == ParserRuleContext_MATCH_BODY) {
		return []ParserRuleContext{nextContext, ParserRuleContext_BINARY_OPERATOR, ParserRuleContext_IS_KEYWORD, ParserRuleContext_DOT, ParserRuleContext_ANNOT_CHAINING_TOKEN, ParserRuleContext_OPTIONAL_CHAINING_TOKEN, ParserRuleContext_CONDITIONAL_EXPRESSION, ParserRuleContext_XML_NAVIGATE_EXPR, ParserRuleContext_MEMBER_ACCESS_KEY_EXPR, ParserRuleContext_RIGHT_ARROW, ParserRuleContext_SYNC_SEND_TOKEN}
	}
	return []ParserRuleContext{ParserRuleContext_BINARY_OPERATOR, ParserRuleContext_IS_KEYWORD, ParserRuleContext_DOT, ParserRuleContext_ANNOT_CHAINING_TOKEN, ParserRuleContext_OPTIONAL_CHAINING_TOKEN, ParserRuleContext_CONDITIONAL_EXPRESSION, ParserRuleContext_XML_NAVIGATE_EXPR, ParserRuleContext_MEMBER_ACCESS_KEY_EXPR, ParserRuleContext_RIGHT_ARROW, ParserRuleContext_SYNC_SEND_TOKEN, nextContext}
}

func (this *BallerinaParserErrorHandler) modifyAlternativesWithArgListStart(alternatives []ParserRuleContext) []ParserRuleContext {
	newAlternatives := nil
	this.System.arraycopy(alternatives, 0, newAlternatives, 0, alternatives.length)
	newAlternatives[alternatives.length] = ParserRuleContext_ARG_LIST_OPEN_PAREN
	return newAlternatives
}

func (this *BallerinaParserErrorHandler) getNextRule(currentCtx ParserRuleContext, nextLookahead int) ParserRuleContext {
	this.startContextIfRequired(currentCtx)
	var parentCtx ParserRuleContext
	var nextToken Token
	switch currentCtx {
	case EOF:
		return ParserRuleContext_EOF
	case COMP_UNIT:
		return ParserRuleContext_TOP_LEVEL_NODE
	case FUNC_DEF:
		return ParserRuleContext_FUNC_DEF_START
	case FUNC_DEF_FIRST_QUALIFIER:
		return ParserRuleContext_FUNC_DEF_WITHOUT_FIRST_QUALIFIER
	case FUNC_TYPE_FIRST_QUALIFIER:
		return ParserRuleContext_FUNC_TYPE_DESC_START_WITHOUT_FIRST_QUAL
	case FUNC_TYPE_SECOND_QUALIFIER,
		FUNC_DEF_SECOND_QUALIFIER,
		FUNC_DEF_OR_FUNC_TYPE:
		return ParserRuleContext_FUNCTION_KEYWORD
	case ANON_FUNC_EXPRESSION:
		return ParserRuleContext_ANON_FUNC_EXPRESSION_START
	case FUNC_TYPE_DESC:
		return ParserRuleContext_FUNC_TYPE_DESC_START
	case EXTERNAL_FUNC_BODY:
		return ParserRuleContext_ASSIGN_OP
	case FUNC_BODY_BLOCK:
		return ParserRuleContext_OPEN_BRACE
	case STATEMENT,
		STATEMENT_WITHOUT_ANNOTS:
		this.endContext()
		return ParserRuleContext_CLOSE_BRACE
	case ASSIGN_OP:
		return this.getNextRuleForEqualOp()
	case COMPOUND_BINARY_OPERATOR:
		return ParserRuleContext_ASSIGN_OP
	case CLOSE_BRACE:
		return this.getNextRuleForCloseBrace(nextLookahead)
	case CLOSE_PARENTHESIS:
		return this.getNextRuleForCloseParenthesis()
	case EXPRESSION,
		BASIC_LITERAL:
		return this.getNextRuleForExpr()
	case FUNC_NAME:
		grandParentCtx := this.getGrandParentContext()
		if (grandParentCtx == ParserRuleContext_OBJECT_CONSTRUCTOR_MEMBER) || (grandParentCtx == ParserRuleContext_CLASS_MEMBER) {
			return ParserRuleContext_OPTIONAL_RELATIVE_PATH
		}
		return ParserRuleContext_OPEN_PARENTHESIS
	case OPEN_BRACE:
		return this.getNextRuleForOpenBrace()
	case OPEN_PARENTHESIS:
		return this.getNextRuleForOpenParenthesis()
	case SEMICOLON:
		return this.getNextRuleForSemicolon(nextLookahead)
	case SIMPLE_TYPE_DESCRIPTOR:
		return ParserRuleContext_TYPE_DESC_RHS
	case VARIABLE_NAME,
		PARAMETER_NAME_RHS:
		return this.getNextRuleForVarName()
	case REQUIRED_PARAM,
		DEFAULTABLE_PARAM,
		REST_PARAM:
		return ParserRuleContext_PARAM_START
	case REST_PARAM_RHS:
		this.switchContext(ParserRuleContext.REST_PARAM)
		return ParserRuleContext_ELLIPSIS
	case ASSIGNMENT_STMT:
		return ParserRuleContext_VARIABLE_NAME
	case VAR_DECL_STMT:
		return ParserRuleContext_TYPE_DESC_IN_TYPE_BINDING_PATTERN
	case EXPRESSION_RHS:
		return ParserRuleContext_BINARY_OPERATOR
	case BINARY_OPERATOR:
		return ParserRuleContext_EXPRESSION
	case COMMA:
		return this.getNextRuleForComma()
	case AFTER_PARAMETER_TYPE:
		return this.getNextRuleForParamType()
	case MODULE_TYPE_DEFINITION:
		return ParserRuleContext_TYPE_KEYWORD
	case CLOSED_RECORD_BODY_END:
		this.endContext()
		nextToken = this.this.tokenReader.peek(nextLookahead)
		if nextToken.kind == SyntaxKind_EOF_TOKEN {
			return ParserRuleContext_EOF
		}
		return ParserRuleContext_TYPE_DESC_RHS
	case CLOSED_RECORD_BODY_START:
		return ParserRuleContext_RECORD_FIELD_OR_RECORD_END
	case ELLIPSIS:
		parentCtx = this.getParentContext()
		switch parentCtx {
		case MAPPING_CONSTRUCTOR,
			LIST_CONSTRUCTOR,
			ARG_LIST:
			ParserRuleContext_EXPRESSION
		case STMT_START_BRACKETED_LIST,
			BRACKETED_LIST,
			TUPLE_MEMBERS:
			ParserRuleContext_CLOSE_BRACKET
		case REST_MATCH_PATTERN:
			ParserRuleContext_VAR_KEYWORD
		case RELATIVE_RESOURCE_PATH:
			ParserRuleContext_OPTIONAL_PATH_PARAM_NAME
		case CLIENT_RESOURCE_ACCESS_ACTION:
			ParserRuleContext_EXPRESSION
		default:
			ParserRuleContext_VARIABLE_NAME
		}
	case QUESTION_MARK:
		return this.getNextRuleForQuestionMark()
	case RECORD_TYPE_DESCRIPTOR:
		return ParserRuleContext_RECORD_KEYWORD
	case ASTERISK:
		parentCtx = this.getParentContext()
		switch parentCtx {
		case ARRAY_TYPE_DESCRIPTOR:
			ParserRuleContext_CLOSE_BRACKET
		case XML_ATOMIC_NAME_PATTERN:
			this.endContext()
			ParserRuleContext_XML_NAME_PATTERN_RHS
		case REQUIRED_PARAM,
			DEFAULTABLE_PARAM:
			ParserRuleContext_TYPE_DESC_IN_PARAM
		default:
			ParserRuleContext_TYPE_REFERENCE_IN_TYPE_INCLUSION
		}
	case TYPE_NAME:
		return ParserRuleContext_TYPE_DESC_IN_TYPE_DEF
	case OBJECT_TYPE_DESCRIPTOR:
		return ParserRuleContext_OBJECT_TYPE_START
	case SECOND_OBJECT_CONS_QUALIFIER,
		SECOND_OBJECT_TYPE_QUALIFIER:
		return ParserRuleContext_OBJECT_KEYWORD
	case FIRST_OBJECT_CONS_QUALIFIER:
		return ParserRuleContext_OBJECT_CONS_WITHOUT_FIRST_QUALIFIER
	case FIRST_OBJECT_TYPE_QUALIFIER:
		return ParserRuleContext_OBJECT_TYPE_WITHOUT_FIRST_QUALIFIER
	case OPEN_BRACKET:
		return this.getNextRuleForOpenBracket()
	case CLOSE_BRACKET:
		return this.getNextRuleForCloseBracket()
	case DOT:
		return this.getNextRuleForDot()
	case METHOD_CALL_DOT:
		return ParserRuleContext_VARIABLE_NAME
	case BLOCK_STMT:
		return ParserRuleContext_OPEN_BRACE
	case IF_BLOCK:
		return ParserRuleContext_IF_KEYWORD
	case WHILE_BLOCK:
		return ParserRuleContext_WHILE_KEYWORD
	case DO_BLOCK:
		return ParserRuleContext_DO_KEYWORD
	case CALL_STMT:
		return ParserRuleContext_CALL_STMT_START
	case PANIC_STMT:
		return ParserRuleContext_PANIC_KEYWORD
	case FUNC_CALL:
		return ParserRuleContext_IMPORT_PREFIX
	case IMPORT_PREFIX,
		NAMESPACE_PREFIX:
		return ParserRuleContext_SEMICOLON
	case SLASH:
		parentCtx = this.getParentContext()
		if parentCtx == ParserRuleContext_ABSOLUTE_RESOURCE_PATH {
			return ParserRuleContext_IDENTIFIER
		} else if parentCtx == ParserRuleContext_RELATIVE_RESOURCE_PATH {
			return ParserRuleContext_RESOURCE_PATH_SEGMENT
		} else if parentCtx == ParserRuleContext_CLIENT_RESOURCE_ACCESS_ACTION {
			return ParserRuleContext_RESOURCE_ACCESS_PATH_SEGMENT
		}
		return ParserRuleContext_IMPORT_MODULE_NAME
	case IMPORT_ORG_OR_MODULE_NAME:
		return ParserRuleContext_IMPORT_DECL_ORG_OR_MODULE_NAME_RHS
	case IMPORT_MODULE_NAME:
		return ParserRuleContext_AFTER_IMPORT_MODULE_NAME
	case IMPORT_DECL:
		return ParserRuleContext_IMPORT_KEYWORD
	case CONTINUE_STATEMENT:
		return ParserRuleContext_CONTINUE_KEYWORD
	case BREAK_STATEMENT:
		return ParserRuleContext_BREAK_KEYWORD
	case RETURN_STMT:
		return ParserRuleContext_RETURN_KEYWORD
	case FAIL_STATEMENT:
		return ParserRuleContext_FAIL_KEYWORD
	case ACCESS_EXPRESSION:
		return ParserRuleContext_VARIABLE_REF
	case MAPPING_FIELD_NAME:
		return ParserRuleContext_SPECIFIC_FIELD_RHS
	case COLON:
		return this.getNextRuleForColon()
	case VAR_REF_COLON:
		this.startContext(ParserRuleContext.VARIABLE_REF)
		return ParserRuleContext_IDENTIFIER
	case TYPE_REF_COLON:
		this.startContext(ParserRuleContext.VAR_DECL_STMT)
		this.startContext(ParserRuleContext.TYPE_DESC_IN_TYPE_BINDING_PATTERN)
		return ParserRuleContext_IDENTIFIER
	case STRING_LITERAL_TOKEN:
		parentCtx = this.getParentContext()
		if parentCtx == ParserRuleContext_SERVICE_DECL {
			return ParserRuleContext_ON_KEYWORD
		}
		return ParserRuleContext_COLON
	case COMPUTED_FIELD_NAME:
		return ParserRuleContext_OPEN_BRACKET
	case LISTENERS_LIST:
		return ParserRuleContext_EXPRESSION
	case SERVICE_DECL:
		return ParserRuleContext_SERVICE_DECL_START
	case SERVICE_DECL_QUALIFIER:
		return ParserRuleContext_SERVICE_KEYWORD
	case LISTENER_DECL:
		return ParserRuleContext_LISTENER_KEYWORD
	case CONSTANT_DECL:
		return ParserRuleContext_CONST_KEYWORD
	case TYPEOF_EXPRESSION:
		return ParserRuleContext_TYPEOF_KEYWORD
	case OPTIONAL_TYPE_DESCRIPTOR:
		return ParserRuleContext_QUESTION_MARK
	case UNARY_EXPRESSION:
		return ParserRuleContext_UNARY_OPERATOR
	case UNARY_OPERATOR:
		return ParserRuleContext_EXPRESSION
	case ARRAY_TYPE_DESCRIPTOR:
		return ParserRuleContext_OPEN_BRACKET
	case AT:
		return ParserRuleContext_ANNOT_REFERENCE
	case DOC_STRING:
		return ParserRuleContext_ANNOTATIONS
	case ANNOTATIONS:
		return ParserRuleContext_AT
	case MAPPING_CONSTRUCTOR:
		return ParserRuleContext_OPEN_BRACE
	case VARIABLE_REF,
		TYPE_REFERENCE,
		TYPE_REFERENCE_IN_TYPE_INCLUSION,
		ANNOT_REFERENCE,
		FIELD_ACCESS_IDENTIFIER:
		return ParserRuleContext_QUALIFIED_IDENTIFIER_START_IDENTIFIER
	case QUALIFIED_IDENTIFIER_START_IDENTIFIER,
		XML_ATOMIC_NAME_IDENTIFIER:
		nextToken = this.this.tokenReader.peek(nextLookahead)
		if nextToken.kind == SyntaxKind_COLON_TOKEN {
			return ParserRuleContext_COLON
		}
	case IDENTIFIER,
		SIMPLE_TYPE_DESC_IDENTIFIER:
		return this.getNextRuleForIdentifier()
	case QUALIFIED_IDENTIFIER_PREDECLARED_PREFIX:
		return ParserRuleContext_COLON
	case PATH_SEGMENT_IDENT:
		return ParserRuleContext_RELATIVE_RESOURCE_PATH_END
	case NIL_LITERAL:
		return ParserRuleContext_OPEN_PARENTHESIS
	case LOCAL_TYPE_DEFINITION_STMT:
		return ParserRuleContext_TYPE_KEYWORD
	case RIGHT_ARROW:
		return ParserRuleContext_EXPRESSION
	case DECIMAL_INTEGER_LITERAL_TOKEN,
		HEX_INTEGER_LITERAL_TOKEN:
		return this.getNextRuleForDecimalIntegerLiteral()
	case EXPRESSION_STATEMENT:
		return ParserRuleContext_EXPRESSION_STATEMENT_START
	case LOCK_STMT:
		return ParserRuleContext_LOCK_KEYWORD
	case LOCK_KEYWORD:
		return ParserRuleContext_BLOCK_STMT
	case RECORD_FIELD:
		return ParserRuleContext_RECORD_FIELD_START
	case ANNOTATION_TAG:
		return ParserRuleContext_ANNOT_OPTIONAL_ATTACH_POINTS
	case ANNOT_ATTACH_POINTS_LIST:
		return ParserRuleContext_ATTACH_POINT
	case FIELD_IDENT,
		FUNCTION_IDENT,
		IDENT_AFTER_OBJECT_IDENT,
		SINGLE_KEYWORD_ATTACH_POINT_IDENT:
		return ParserRuleContext_ATTACH_POINT_END
	case OBJECT_IDENT:
		return ParserRuleContext_IDENT_AFTER_OBJECT_IDENT
	case RECORD_IDENT:
		return ParserRuleContext_FIELD_IDENT
	case SERVICE_IDENT:
		return ParserRuleContext_SERVICE_IDENT_RHS
	case REMOTE_IDENT:
		return ParserRuleContext_FUNCTION_IDENT
	case ANNOTATION_DECL:
		return ParserRuleContext_ANNOTATION_DECL_START
	case XML_NAMESPACE_DECLARATION:
		return ParserRuleContext_XMLNS_KEYWORD
	case CONSTANT_EXPRESSION:
		return ParserRuleContext_CONSTANT_EXPRESSION_START
	case NAMED_WORKER_DECL:
		return ParserRuleContext_NAMED_WORKER_DECL_START
	case WORKER_NAME:
		return ParserRuleContext_WORKER_NAME_RHS
	case FORK_STMT:
		return ParserRuleContext_FORK_KEYWORD
	case XML_FILTER_EXPR:
		return ParserRuleContext_DOT_LT_TOKEN
	case DOT_LT_TOKEN:
		return ParserRuleContext_XML_NAME_PATTERN
	case XML_NAME_PATTERN:
		return ParserRuleContext_XML_ATOMIC_NAME_PATTERN
	case XML_ATOMIC_NAME_PATTERN:
		return ParserRuleContext_XML_ATOMIC_NAME_PATTERN_START
	case XML_STEP_EXPR:
		return ParserRuleContext_XML_STEP_START
	case SLASH_ASTERISK_TOKEN:
		return ParserRuleContext_EXPRESSION_RHS
	case DOUBLE_SLASH_DOUBLE_ASTERISK_LT_TOKEN,
		SLASH_LT_TOKEN:
		return ParserRuleContext_XML_NAME_PATTERN
	case OBJECT_CONSTRUCTOR:
		return ParserRuleContext_OBJECT_CONSTRUCTOR_START
	case ABSOLUTE_RESOURCE_PATH:
		return ParserRuleContext_ABSOLUTE_RESOURCE_PATH_START
	case ABSOLUTE_PATH_SINGLE_SLASH:
		return ParserRuleContext_SERVICE_DECL_RHS
	case SERVICE_DECL_RHS:
		this.endContext()
		return ParserRuleContext_ON_KEYWORD
	case OBJECT_CONSTRUCTOR_BLOCK:
		this.endContext()
		return ParserRuleContext_OPEN_BRACE
	case SERVICE_VAR_DECL_RHS:
		this.switchContext(ParserRuleContext.VAR_DECL_STMT)
		return ParserRuleContext_TYPED_BINDING_PATTERN_TYPE_RHS
	case RELATIVE_RESOURCE_PATH:
		return ParserRuleContext_RELATIVE_RESOURCE_PATH_START
	case RESOURCE_PATH_PARAM:
		return ParserRuleContext_OPEN_BRACKET
	case RESOURCE_ACCESSOR_DEF_OR_DECL_RHS:
		this.endContext()
		return ParserRuleContext_OPEN_PARENTHESIS
	case ERROR_CONSTRUCTOR:
		return ParserRuleContext_ERROR_KEYWORD
	case ERROR_CONS_ERROR_KEYWORD_RHS:
		this.startContext(ParserRuleContext.ERROR_CONSTRUCTOR)
		return ParserRuleContext_ERROR_CONSTRUCTOR_RHS
	case LIST_BINDING_PATTERN_RHS:
		return this.getNextRuleForBindingPattern()
	case TUPLE_MEMBERS:
		return ParserRuleContext_TUPLE_MEMBER
	case SINGLE_OR_ALTERNATE_WORKER:
		return ParserRuleContext_PEER_WORKER_NAME
	case XML_STEP_EXTENDS:
		return ParserRuleContext_XML_STEP_EXTEND
	case XML_STEP_EXTEND_END,
		SINGLE_OR_ALTERNATE_WORKER_END:
		this.endContext()
		return ParserRuleContext_EXPRESSION_RHS
	default:
		return this.getNextRuleInternal(currentCtx, nextLookahead)
	}
}

func (this *BallerinaParserErrorHandler) getNextRuleInternal(currentCtx ParserRuleContext, nextLookahead int) ParserRuleContext {
	var parentCtx ParserRuleContext
	var grandParentCtx ParserRuleContext
	switch currentCtx {
	case LIST_CONSTRUCTOR:
		return ParserRuleContext_OPEN_BRACKET
	case FOREACH_STMT:
		return ParserRuleContext_FOREACH_KEYWORD
	case TYPE_CAST:
		return ParserRuleContext_LT
	case PIPE:
		parentCtx = this.getParentContext()
		if parentCtx == ParserRuleContext_ALTERNATE_WAIT_EXPRS {
			return ParserRuleContext_EXPRESSION
		} else if parentCtx == ParserRuleContext_XML_NAME_PATTERN {
			return ParserRuleContext_XML_ATOMIC_NAME_PATTERN
		} else if parentCtx == ParserRuleContext_MATCH_PATTERN {
			return ParserRuleContext_MATCH_PATTERN_START
		} else if parentCtx == ParserRuleContext_SINGLE_OR_ALTERNATE_WORKER {
			return ParserRuleContext_PEER_WORKER_NAME
		}
		return ParserRuleContext_TYPE_DESCRIPTOR
	case TABLE_CONSTRUCTOR:
		return ParserRuleContext_OPEN_BRACKET
	case KEY_SPECIFIER:
		return ParserRuleContext_KEY_KEYWORD
	case LET_EXPRESSION:
		return ParserRuleContext_LET_KEYWORD
	case LET_EXPR_LET_VAR_DECL,
		LET_CLAUSE_LET_VAR_DECL:
		return ParserRuleContext_LET_VAR_DECL_START
	case ORDER_KEY_LIST:
		return ParserRuleContext_EXPRESSION
	case END_OF_TYPE_DESC:
		return this.getNextRuleForTypeDescriptor()
	case TYPED_BINDING_PATTERN:
		return ParserRuleContext_TYPE_DESCRIPTOR
	case BINDING_PATTERN_STARTING_IDENTIFIER:
		return ParserRuleContext_VARIABLE_NAME
	case REST_BINDING_PATTERN:
		return ParserRuleContext_ELLIPSIS
	case LIST_BINDING_PATTERN:
		return ParserRuleContext_OPEN_BRACKET
	case MAPPING_BINDING_PATTERN:
		return ParserRuleContext_OPEN_BRACE
	case FIELD_BINDING_PATTERN:
		return ParserRuleContext_FIELD_BINDING_PATTERN_NAME
	case FIELD_BINDING_PATTERN_NAME:
		return ParserRuleContext_FIELD_BINDING_PATTERN_END
	case LT:
		return this.getNextRuleForLt()
	case INFERRED_TYPEDESC_DEFAULT_START_LT:
		return ParserRuleContext_INFERRED_TYPEDESC_DEFAULT_END_GT
	case GT:
		return this.getNextRuleForGt()
	case STREAM_TYPE_PARAM_START_TOKEN:
		return ParserRuleContext_TYPE_DESC_IN_STREAM_TYPE_DESC
	case INFERRED_TYPEDESC_DEFAULT_END_GT:
		return ParserRuleContext_END_OF_PARAMS_OR_NEXT_PARAM_START
	case TYPE_CAST_PARAM_START:
		this.startContext(ParserRuleContext.TYPE_CAST)
		return ParserRuleContext_TYPE_CAST_PARAM
	case TEMPLATE_END:
		return ParserRuleContext_EXPRESSION_RHS
	case TEMPLATE_START:
		return ParserRuleContext_TEMPLATE_BODY
	case TEMPLATE_BODY:
		return ParserRuleContext_TEMPLATE_MEMBER
	case TEMPLATE_STRING:
		return ParserRuleContext_TEMPLATE_STRING_RHS
	case INTERPOLATION_START_TOKEN:
		return ParserRuleContext_EXPRESSION
	case ARG_LIST_OPEN_PAREN:
		return ParserRuleContext_ARG_LIST
	case ARG_LIST_END:
		this.endContext()
		return ParserRuleContext_ARG_LIST_CLOSE_PAREN
	case ARG_LIST_CLOSE_PAREN:
		parentCtx = this.getParentContext()
		if parentCtx == ParserRuleContext_ERROR_CONSTRUCTOR {
			this.endContext()
		} else if parentCtx == ParserRuleContext_XML_STEP_EXTENDS {
			return ParserRuleContext_XML_STEP_EXTEND
		} else if parentCtx == ParserRuleContext_CLIENT_RESOURCE_ACCESS_ACTION {
			return ParserRuleContext_ACTION_END
		} else if parentCtx == ParserRuleContext_NATURAL_EXPRESSION {
			return ParserRuleContext_OPEN_BRACE
		}
		return ParserRuleContext_EXPRESSION_RHS
	case ARG_LIST:
		return ParserRuleContext_ARG_START_OR_ARG_LIST_END
	case QUERY_EXPRESSION_END:
		this.endContext()
		this.endContext()
		return ParserRuleContext_EXPRESSION_RHS
	case TYPE_DESC_IN_ANNOTATION_DECL,
		TYPE_DESC_BEFORE_IDENTIFIER,
		TYPE_DESC_IN_RECORD_FIELD,
		TYPE_DESC_IN_PARAM,
		TYPE_DESC_IN_TYPE_BINDING_PATTERN,
		TYPE_DESC_IN_TYPE_DEF,
		TYPE_DESC_IN_ANGLE_BRACKETS,
		TYPE_DESC_IN_RETURN_TYPE_DESC,
		TYPE_DESC_IN_EXPRESSION,
		TYPE_DESC_IN_STREAM_TYPE_DESC,
		TYPE_DESC_IN_PARENTHESIS,
		TYPE_DESC_IN_TUPLE,
		TYPE_DESC_IN_SERVICE,
		TYPE_DESC_IN_PATH_PARAM,
		TYPE_DESC_BEFORE_IDENTIFIER_IN_GROUPING_KEY:
		return ParserRuleContext_TYPE_DESCRIPTOR
	case CLASS_DESCRIPTOR_IN_NEW_EXPR:
		return ParserRuleContext_CLASS_DESCRIPTOR
	case VAR_DECL_STARTED_WITH_DENTIFIER,
		TYPE_DESC_RHS_IN_TYPED_BP:
		this.startContext(ParserRuleContext.TYPE_DESC_IN_TYPE_BINDING_PATTERN)
		return ParserRuleContext_TYPE_DESC_RHS
	case ROW_TYPE_PARAM:
		return ParserRuleContext_LT
	case PARENTHESISED_TYPE_DESC_START:
		return ParserRuleContext_TYPE_DESC_IN_PARENTHESIS
	case SELECT_CLAUSE:
		return ParserRuleContext_SELECT_KEYWORD
	case COLLECT_CLAUSE:
		return ParserRuleContext_COLLECT_KEYWORD
	case WHERE_CLAUSE:
		return ParserRuleContext_WHERE_KEYWORD
	case FROM_CLAUSE:
		return ParserRuleContext_FROM_KEYWORD
	case LET_CLAUSE:
		return ParserRuleContext_LET_KEYWORD
	case ORDER_BY_CLAUSE:
		return ParserRuleContext_ORDER_KEYWORD
	case GROUP_BY_CLAUSE:
		return ParserRuleContext_GROUP_KEYWORD
	case ON_CONFLICT_CLAUSE:
		return ParserRuleContext_ON_KEYWORD
	case LIMIT_CLAUSE:
		return ParserRuleContext_LIMIT_KEYWORD
	case JOIN_CLAUSE:
		return ParserRuleContext_JOIN_CLAUSE_START
	case ON_CLAUSE:
		return ParserRuleContext_ON_KEYWORD
	case QUERY_EXPRESSION:
		return ParserRuleContext_FROM_CLAUSE
	case QUERY_CONSTRUCT_TYPE_RHS:
		this.startContext(ParserRuleContext.QUERY_EXPRESSION)
		return ParserRuleContext_FROM_CLAUSE
	case EXPRESSION_START_TABLE_KEYWORD_RHS:
		this.startContext(ParserRuleContext.TABLE_CONSTRUCTOR_OR_QUERY_EXPRESSION)
		return ParserRuleContext_TABLE_KEYWORD_RHS
	case QUERY_EXPRESSION_RHS:
		parentCtx = this.getParentContext()
		if parentCtx == ParserRuleContext_LET_CLAUSE_LET_VAR_DECL {
			this.endContext()
		}
		return ParserRuleContext_RESULT_CLAUSE
	case INTERMEDIATE_CLAUSE:
		parentCtx = this.getParentContext()
		if parentCtx == ParserRuleContext_LET_CLAUSE_LET_VAR_DECL {
			this.endContext()
		}
		return ParserRuleContext_INTERMEDIATE_CLAUSE_START
	case QUERY_ACTION_RHS:
		return ParserRuleContext_DO_CLAUSE
	case TABLE_CONSTRUCTOR_OR_QUERY_EXPRESSION:
		return ParserRuleContext_TABLE_CONSTRUCTOR_OR_QUERY_START
	case BITWISE_AND_OPERATOR:
		return ParserRuleContext_TYPE_DESCRIPTOR
	case EXPR_FUNC_BODY_START:
		return ParserRuleContext_EXPRESSION
	case MODULE_LEVEL_AMBIGUOUS_FUNC_TYPE_DESC_RHS:
		this.endContext()
		this.startContext(ParserRuleContext.VAR_DECL_STMT)
		this.startContext(ParserRuleContext.TYPE_DESC_IN_TYPE_BINDING_PATTERN)
		return ParserRuleContext_TYPE_DESC_RHS
	case STMT_LEVEL_AMBIGUOUS_FUNC_TYPE_DESC_RHS:
		this.endContext()
		if !this.isInTypeDescContext() {
			this.switchContext(ParserRuleContext.VAR_DECL_STMT)
			this.startContext(ParserRuleContext.TYPE_DESC_IN_TYPE_BINDING_PATTERN)
		}
		return ParserRuleContext_TYPE_DESC_RHS
	case FUNC_TYPE_DESC_END:
		this.endContext()
		return ParserRuleContext_TYPE_DESC_RHS
	case BRACED_EXPR_OR_ANON_FUNC_PARAMS:
		return ParserRuleContext_IMPLICIT_ANON_FUNC_PARAM
	case IMPLICIT_ANON_FUNC_PARAM:
		return ParserRuleContext_BRACED_EXPR_OR_ANON_FUNC_PARAM_RHS
	case EXPLICIT_ANON_FUNC_EXPR_BODY_START:
		this.endContext()
		return ParserRuleContext_EXPR_FUNC_BODY_START
	case OBJECT_CONSTRUCTOR_MEMBER:
		return ParserRuleContext_OBJECT_CONSTRUCTOR_MEMBER_START
	case CLASS_MEMBER,
		OBJECT_TYPE_MEMBER:
		return ParserRuleContext_CLASS_MEMBER_OR_OBJECT_MEMBER_START
	case ANNOTATION_END:
		return this.getNextRuleForAnnotationEnd(nextLookahead)
	case PLUS_TOKEN,
		MINUS_TOKEN:
		return ParserRuleContext_SIGNED_INT_OR_FLOAT_RHS
	case SIGNED_INT_OR_FLOAT_RHS:
		return this.getNextRuleForExpr()
	case TUPLE_TYPE_DESC_START:
		return ParserRuleContext_TUPLE_MEMBERS
	case METHOD_NAME:
		parentCtx = this.getParentContext()
		if parentCtx == ParserRuleContext_XML_STEP_EXTENDS {
			return ParserRuleContext_ARG_LIST_OPEN_PAREN
		}
		return ParserRuleContext_OPTIONAL_RESOURCE_ACCESS_ACTION_ARG_LIST
	case DEFAULT_WORKER_NAME_IN_ASYNC_SEND:
		return ParserRuleContext_SEMICOLON
	case SYNC_SEND_TOKEN:
		return ParserRuleContext_PEER_WORKER_NAME
	case LEFT_ARROW_TOKEN:
		return ParserRuleContext_RECEIVE_WORKERS
	case MULTI_RECEIVE_WORKERS:
		return ParserRuleContext_OPEN_BRACE
	case RECEIVE_FIELD_NAME:
		return ParserRuleContext_COLON
	case WAIT_FIELD_NAME:
		return ParserRuleContext_WAIT_FIELD_NAME_RHS
	case ALTERNATE_WAIT_EXPR_LIST_END:
		return this.getNextRuleForWaitExprListEnd()
	case MULTI_WAIT_FIELDS:
		return ParserRuleContext_OPEN_BRACE
	case ALTERNATE_WAIT_EXPRS:
		return ParserRuleContext_EXPRESSION
	case ANNOT_CHAINING_TOKEN:
		return ParserRuleContext_FIELD_ACCESS_IDENTIFIER
	case DO_CLAUSE:
		return ParserRuleContext_DO_KEYWORD
	case LET_CLAUSE_END:
		parentCtx = this.getParentContext()
		if parentCtx == ParserRuleContext_LET_CLAUSE_LET_VAR_DECL {
			this.endContext()
			return ParserRuleContext_QUERY_PIPELINE_RHS
		}
		return ParserRuleContext_QUERY_PIPELINE_RHS
	case GROUP_BY_CLAUSE_END,
		ORDER_CLAUSE_END,
		JOIN_CLAUSE_END:
		this.endContext()
		return ParserRuleContext_QUERY_PIPELINE_RHS
	case MEMBER_ACCESS_KEY_EXPR:
		return ParserRuleContext_OPEN_BRACKET
	case OPTIONAL_CHAINING_TOKEN:
		return ParserRuleContext_FIELD_ACCESS_IDENTIFIER
	case CONDITIONAL_EXPRESSION:
		return ParserRuleContext_QUESTION_MARK
	case TRANSACTION_STMT:
		return ParserRuleContext_TRANSACTION_KEYWORD
	case RETRY_STMT:
		return ParserRuleContext_RETRY_KEYWORD
	case ROLLBACK_STMT:
		return ParserRuleContext_ROLLBACK_KEYWORD
	case MODULE_ENUM_DECLARATION:
		return ParserRuleContext_ENUM_KEYWORD
	case MODULE_ENUM_NAME:
		return ParserRuleContext_OPEN_BRACE
	case ENUM_MEMBER_LIST:
		return ParserRuleContext_ENUM_MEMBER_START
	case ENUM_MEMBER_NAME:
		return ParserRuleContext_ENUM_MEMBER_RHS
	case TYPED_BINDING_PATTERN_TYPE_RHS:
		return ParserRuleContext_BINDING_PATTERN
	case UNION_OR_INTERSECTION_TOKEN:
		return ParserRuleContext_TYPE_DESCRIPTOR
	case MATCH_STMT:
		return ParserRuleContext_MATCH_KEYWORD
	case MATCH_BODY:
		return ParserRuleContext_OPEN_BRACE
	case MATCH_PATTERN:
		return ParserRuleContext_MATCH_PATTERN_START
	case MATCH_PATTERN_END:
		this.endContext()
		return this.getNextRuleForMatchPattern()
	case MATCH_PATTERN_RHS:
		return this.getNextRuleForMatchPattern()
	case RIGHT_DOUBLE_ARROW:
		return ParserRuleContext_BLOCK_STMT
	case LIST_MATCH_PATTERN:
		return ParserRuleContext_OPEN_BRACKET
	case REST_MATCH_PATTERN:
		return ParserRuleContext_ELLIPSIS
	case ERROR_BINDING_PATTERN:
		return ParserRuleContext_ERROR_KEYWORD
	case SIMPLE_BINDING_PATTERN:
		return ParserRuleContext_ERROR_MESSAGE_BINDING_PATTERN_END
	case ERROR_MESSAGE_BINDING_PATTERN_END_COMMA:
		return ParserRuleContext_ERROR_MESSAGE_BINDING_PATTERN_RHS
	case ERROR_CAUSE_SIMPLE_BINDING_PATTERN:
		return ParserRuleContext_ERROR_FIELD_BINDING_PATTERN_END
	case NAMED_ARG_BINDING_PATTERN:
		return ParserRuleContext_ASSIGN_OP
	case MAPPING_MATCH_PATTERN:
		return ParserRuleContext_OPEN_BRACE
	case ERROR_MATCH_PATTERN:
		return ParserRuleContext_ERROR_KEYWORD
	case ERROR_ARG_LIST_MATCH_PATTERN_FIRST_ARG:
		return ParserRuleContext_ERROR_ARG_LIST_MATCH_PATTERN_START
	case ERROR_MESSAGE_MATCH_PATTERN_END_COMMA:
		return ParserRuleContext_ERROR_MESSAGE_MATCH_PATTERN_RHS
	case ERROR_CAUSE_MATCH_PATTERN:
		return ParserRuleContext_ERROR_FIELD_MATCH_PATTERN_RHS
	case NAMED_ARG_MATCH_PATTERN:
		return ParserRuleContext_IDENTIFIER
	case MODULE_CLASS_DEFINITION:
		return ParserRuleContext_MODULE_CLASS_DEFINITION_START
	case FIRST_CLASS_TYPE_QUALIFIER:
		return ParserRuleContext_CLASS_DEF_WITHOUT_FIRST_QUALIFIER
	case SECOND_CLASS_TYPE_QUALIFIER:
		return ParserRuleContext_CLASS_DEF_WITHOUT_SECOND_QUALIFIER
	case THIRD_CLASS_TYPE_QUALIFIER:
		return ParserRuleContext_CLASS_DEF_WITHOUT_THIRD_QUALIFIER
	case FOURTH_CLASS_TYPE_QUALIFIER:
		return ParserRuleContext_CLASS_KEYWORD
	case CLASS_KEYWORD:
		return ParserRuleContext_CLASS_NAME
	case CLASS_NAME:
		return ParserRuleContext_OPEN_BRACE
	case OBJECT_MEMBER_VISIBILITY_QUAL:
		return ParserRuleContext_OBJECT_FUNC_OR_FIELD_WITHOUT_VISIBILITY
	case OBJECT_FIELD_START:
		parentCtx = this.getParentContext()
		if parentCtx == ParserRuleContext_OBJECT_TYPE_MEMBER {
			return ParserRuleContext_TYPE_DESC_BEFORE_IDENTIFIER
		}
		return ParserRuleContext_OBJECT_FIELD_QUALIFIER
	case ON_FAIL_CLAUSE:
		return ParserRuleContext_ON_KEYWORD
	case OBJECT_FIELD_RHS:
		grandParentCtx = this.getGrandParentContext()
		if grandParentCtx == ParserRuleContext_OBJECT_TYPE_DESCRIPTOR {
			return ParserRuleContext_SEMICOLON
		} else {
			return ParserRuleContext_OPTIONAL_FIELD_INITIALIZER
		}
	case OBJECT_METHOD_FIRST_QUALIFIER:
		return ParserRuleContext_OBJECT_METHOD_WITHOUT_FIRST_QUALIFIER
	case OBJECT_METHOD_SECOND_QUALIFIER:
		return ParserRuleContext_OBJECT_METHOD_WITHOUT_SECOND_QUALIFIER
	case OBJECT_METHOD_THIRD_QUALIFIER:
		return ParserRuleContext_OBJECT_METHOD_WITHOUT_THIRD_QUALIFIER
	case OBJECT_METHOD_FOURTH_QUALIFIER:
		return ParserRuleContext_FUNC_DEF
	case MODULE_VAR_DECL:
		return ParserRuleContext_MODULE_VAR_DECL_START
	case MODULE_VAR_FIRST_QUAL:
		return ParserRuleContext_MODULE_VAR_WITHOUT_FIRST_QUAL
	case MODULE_VAR_SECOND_QUAL:
		return ParserRuleContext_MODULE_VAR_WITHOUT_SECOND_QUAL
	case MODULE_VAR_THIRD_QUAL:
		return ParserRuleContext_VAR_DECL_STMT
	case PARAMETERIZED_TYPE:
		return ParserRuleContext_OPTIONAL_TYPE_PARAMETER
	case MAP_TYPE_DESCRIPTOR:
		return ParserRuleContext_MAP_KEYWORD
	case FUNC_TYPE_FUNC_KEYWORD_RHS:
		return this.getNextRuleForFuncTypeFuncKeywordRhs()
	case TRANSACTION_STMT_TRANSACTION_KEYWORD_RHS:
		this.startContext(ParserRuleContext.TRANSACTION_STMT)
		return ParserRuleContext_BLOCK_STMT
	case BRACED_EXPRESSION:
		return ParserRuleContext_OPEN_PARENTHESIS
	case ARRAY_LENGTH_START:
		this.switchContext(ParserRuleContext.TYPE_DESC_IN_TYPE_BINDING_PATTERN)
		this.startContext(ParserRuleContext.ARRAY_TYPE_DESCRIPTOR)
		return ParserRuleContext_ARRAY_LENGTH
	case RESOURCE_METHOD_CALL_SLASH_TOKEN:
		return ParserRuleContext_CLIENT_RESOURCE_ACCESS_ACTION
	case CLIENT_RESOURCE_ACCESS_ACTION:
		return ParserRuleContext_OPTIONAL_RESOURCE_ACCESS_PATH
	case ACTION_END:
		parentCtx = this.getParentContext()
		if parentCtx == ParserRuleContext_CLIENT_RESOURCE_ACCESS_ACTION {
			this.endContext()
		}
		return this.getNextRuleForAction()
	case NATURAL_EXPRESSION:
		return ParserRuleContext_NATURAL_EXPRESSION_START
	default:
		return this.getNextRuleForKeywords(currentCtx, nextLookahead)
	}
}

func (this *BallerinaParserErrorHandler) getNextRuleForKeywords(currentCtx ParserRuleContext, nextLookahead int) ParserRuleContext {
	var parentCtx ParserRuleContext
	switch currentCtx {
	case PUBLIC_KEYWORD:
		parentCtx = this.getParentContext()
		if (((parentCtx == ParserRuleContext_OBJECT_TYPE_DESCRIPTOR) || (parentCtx == ParserRuleContext_OBJECT_CONSTRUCTOR_MEMBER)) || (parentCtx == ParserRuleContext_CLASS_MEMBER)) || (parentCtx == ParserRuleContext_OBJECT_TYPE_MEMBER) {
			return ParserRuleContext_OBJECT_FUNC_OR_FIELD_WITHOUT_VISIBILITY
		} else if this.isParameter(parentCtx) {
			return ParserRuleContext_TYPE_DESC_IN_PARAM
		}
		return ParserRuleContext_TOP_LEVEL_NODE_WITHOUT_MODIFIER
	case PRIVATE_KEYWORD:
		return ParserRuleContext_OBJECT_FUNC_OR_FIELD_WITHOUT_VISIBILITY
	case ON_KEYWORD:
		parentCtx = this.getParentContext()
		if parentCtx == ParserRuleContext_ANNOTATION_DECL {
			return ParserRuleContext_ANNOT_ATTACH_POINTS_LIST
		} else if parentCtx == ParserRuleContext_ON_CONFLICT_CLAUSE {
			return ParserRuleContext_CONFLICT_KEYWORD
		} else if parentCtx == ParserRuleContext_ON_CLAUSE {
			return ParserRuleContext_EXPRESSION
		} else if parentCtx == ParserRuleContext_ON_FAIL_CLAUSE {
			return ParserRuleContext_FAIL_KEYWORD
		}
		return ParserRuleContext_LISTENERS_LIST
	case SERVICE_KEYWORD:
		return ParserRuleContext_OPTIONAL_SERVICE_DECL_TYPE
	case LISTENER_KEYWORD:
		return ParserRuleContext_TYPE_DESC_BEFORE_IDENTIFIER
	case FINAL_KEYWORD:
		parentCtx = this.getParentContext()
		if (parentCtx == ParserRuleContext_OBJECT_CONSTRUCTOR_MEMBER) || (parentCtx == ParserRuleContext_CLASS_MEMBER) {
			return ParserRuleContext_TYPE_DESC_BEFORE_IDENTIFIER
		}
		return ParserRuleContext_TYPE_DESC_IN_TYPE_BINDING_PATTERN
	case CONST_KEYWORD:
		parentCtx = this.getParentContext()
		if parentCtx == ParserRuleContext_ANNOTATION_DECL {
			return ParserRuleContext_ANNOTATION_KEYWORD
		}
		if parentCtx == ParserRuleContext_NATURAL_EXPRESSION {
			return ParserRuleContext_NATURAL_KEYWORD
		}
		return ParserRuleContext_CONST_DECL_TYPE
	case NATURAL_KEYWORD:
		return ParserRuleContext_OPTIONAL_PARENTHESIZED_ARG_LIST
	case TYPEOF_KEYWORD:
		return ParserRuleContext_EXPRESSION
	case IS_KEYWORD:
		return ParserRuleContext_TYPE_DESC_IN_EXPRESSION
	case NULL_KEYWORD:
		return ParserRuleContext_EXPRESSION_RHS
	case ANNOTATION_KEYWORD:
		return ParserRuleContext_ANNOT_DECL_OPTIONAL_TYPE
	case SOURCE_KEYWORD:
		return ParserRuleContext_ATTACH_POINT_IDENT
	case XMLNS_KEYWORD:
		return ParserRuleContext_CONSTANT_EXPRESSION
	case WORKER_KEYWORD:
		return ParserRuleContext_WORKER_NAME
	case IF_KEYWORD:
		return ParserRuleContext_EXPRESSION
	case ELSE_KEYWORD:
		return ParserRuleContext_ELSE_BODY
	case WHILE_KEYWORD:
		return ParserRuleContext_EXPRESSION
	case CHECKING_KEYWORD:
		return ParserRuleContext_EXPRESSION
	case FAIL_KEYWORD:
		if this.getParentContext() == ParserRuleContext_ON_FAIL_CLAUSE {
			return ParserRuleContext_ON_FAIL_OPTIONAL_BINDING_PATTERN
		}
		return ParserRuleContext_EXPRESSION
	case PANIC_KEYWORD:
		return ParserRuleContext_EXPRESSION
	case IMPORT_KEYWORD:
		return ParserRuleContext_IMPORT_ORG_OR_MODULE_NAME
	case AS_KEYWORD:
		parentCtx = this.getParentContext()
		if parentCtx == ParserRuleContext_IMPORT_DECL {
			return ParserRuleContext_IMPORT_PREFIX
		} else if parentCtx == ParserRuleContext_XML_NAMESPACE_DECLARATION {
			return ParserRuleContext_NAMESPACE_PREFIX
		}
		panic("next rule of as keyword found: " + parentCtx)
	case CONTINUE_KEYWORD,
		BREAK_KEYWORD:
		return ParserRuleContext_SEMICOLON
	case RETURN_KEYWORD:
		return ParserRuleContext_RETURN_STMT_RHS
	case EXTERNAL_KEYWORD:
		return ParserRuleContext_SEMICOLON
	case FUNCTION_KEYWORD:
		parentCtx = this.getParentContext()
		if parentCtx == ParserRuleContext_ANON_FUNC_EXPRESSION {
			return ParserRuleContext_OPEN_PARENTHESIS
		} else if parentCtx == ParserRuleContext_FUNC_TYPE_DESC {
			return ParserRuleContext_FUNC_TYPE_FUNC_KEYWORD_RHS_START
		} else if parentCtx == ParserRuleContext_FUNC_DEF {
			return ParserRuleContext_FUNC_NAME
		}
		return ParserRuleContext_FUNCTION_KEYWORD_RHS
	case RETURNS_KEYWORD:
		return ParserRuleContext_TYPE_DESC_IN_RETURN_TYPE_DESC
	case RECORD_KEYWORD:
		return ParserRuleContext_RECORD_BODY_START
	case TYPE_KEYWORD:
		return ParserRuleContext_TYPE_NAME
	case OBJECT_KEYWORD:
		parentCtx = this.getParentContext()
		if parentCtx == ParserRuleContext_OBJECT_CONSTRUCTOR {
			return ParserRuleContext_OBJECT_CONSTRUCTOR_TYPE_REF
		}
		return ParserRuleContext_OPEN_BRACE
	case OBJECT_TYPE_OBJECT_KEYWORD_RHS:
		this.startContext(ParserRuleContext.OBJECT_TYPE_DESCRIPTOR)
		return ParserRuleContext_OPEN_BRACE
	case ABSTRACT_KEYWORD,
		CLIENT_KEYWORD:
		return ParserRuleContext_OBJECT_KEYWORD
	case FORK_KEYWORD:
		return ParserRuleContext_OPEN_BRACE
	case TRAP_KEYWORD:
		return ParserRuleContext_EXPRESSION
	case FOREACH_KEYWORD:
		return ParserRuleContext_TYPE_DESC_IN_TYPE_BINDING_PATTERN
	case IN_KEYWORD:
		parentCtx = this.getParentContext()
		if parentCtx == ParserRuleContext_LET_EXPR_LET_VAR_DECL {
			this.endContext()
		}
		return ParserRuleContext_EXPRESSION
	case KEY_KEYWORD:
		if this.isInTypeDescContext() {
			return ParserRuleContext_KEY_CONSTRAINTS_RHS
		}
		return ParserRuleContext_OPEN_PARENTHESIS
	case ERROR_KEYWORD:
		return this.getNextRuleForErrorKeyword()
	case LET_KEYWORD:
		parentCtx = this.getParentContext()
		if parentCtx == ParserRuleContext_QUERY_EXPRESSION {
			nextToken := this.this.tokenReader.peek(nextLookahead)
			nextNextToken := this.this.tokenReader.peek(nextLookahead + 1)
			if this.BallerinaParser.isEndOfLetVarDeclarations(nextToken, nextNextToken) {
				return ParserRuleContext_LET_CLAUSE_END
			}
			return ParserRuleContext_LET_CLAUSE_LET_VAR_DECL
		} else if parentCtx == ParserRuleContext_LET_CLAUSE_LET_VAR_DECL {
			this.endContext()
			return ParserRuleContext_LET_CLAUSE_LET_VAR_DECL
		}
		return ParserRuleContext_LET_EXPR_LET_VAR_DECL
	case TABLE_KEYWORD:
		if this.isInTypeDescContext() {
			return ParserRuleContext_ROW_TYPE_PARAM
		}
		return ParserRuleContext_TABLE_KEYWORD_RHS
	case STREAM_KEYWORD:
		parentCtx = this.getParentContext()
		if parentCtx == ParserRuleContext_TABLE_CONSTRUCTOR_OR_QUERY_EXPRESSION {
			return ParserRuleContext_QUERY_EXPRESSION
		}
		return ParserRuleContext_STREAM_TYPE_PARAM_START_TOKEN
	case NEW_KEYWORD:
		return ParserRuleContext_NEW_KEYWORD_RHS
	case XML_KEYWORD,
		RE_KEYWORD,
		STRING_KEYWORD,
		BASE16_KEYWORD,
		BASE64_KEYWORD:
		return ParserRuleContext_TEMPLATE_START
	case SELECT_KEYWORD,
		COLLECT_KEYWORD:
		return ParserRuleContext_EXPRESSION
	case WHERE_KEYWORD:
		parentCtx = this.getParentContext()
		if parentCtx == ParserRuleContext_LET_CLAUSE_LET_VAR_DECL {
			this.endContext()
		}
		return ParserRuleContext_EXPRESSION
	case ORDER_KEYWORD,
		GROUP_KEYWORD:
		return ParserRuleContext_BY_KEYWORD
	case BY_KEYWORD:
		parentCtx = this.getParentContext()
		if parentCtx == ParserRuleContext_GROUP_BY_CLAUSE {
			return ParserRuleContext_GROUPING_KEY_LIST_ELEMENT
		}
		return ParserRuleContext_ORDER_KEY_LIST
	case ORDER_DIRECTION:
		return ParserRuleContext_ORDER_KEY_LIST_END
	case FROM_KEYWORD:
		parentCtx = this.getParentContext()
		if parentCtx == ParserRuleContext_LET_CLAUSE_LET_VAR_DECL {
			this.endContext()
		}
		return ParserRuleContext_TYPE_DESC_IN_TYPE_BINDING_PATTERN
	case JOIN_KEYWORD:
		return ParserRuleContext_TYPE_DESC_IN_TYPE_BINDING_PATTERN
	case START_KEYWORD:
		return ParserRuleContext_EXPRESSION
	case FLUSH_KEYWORD:
		return ParserRuleContext_OPTIONAL_PEER_WORKER
	case PEER_WORKER_NAME:
		parentCtx = this.getParentContext()
		if parentCtx == ParserRuleContext_MULTI_RECEIVE_WORKERS {
			return ParserRuleContext_RECEIVE_FIELD_END
		} else if parentCtx == ParserRuleContext_SINGLE_OR_ALTERNATE_WORKER {
			return ParserRuleContext_SINGLE_OR_ALTERNATE_WORKER_SEPARATOR
		}
		return ParserRuleContext_EXPRESSION_RHS
	case WAIT_KEYWORD:
		return ParserRuleContext_WAIT_KEYWORD_RHS
	case DO_KEYWORD,
		TRANSACTION_KEYWORD:
		return ParserRuleContext_BLOCK_STMT
	case COMMIT_KEYWORD:
		return ParserRuleContext_EXPRESSION_RHS
	case ROLLBACK_KEYWORD:
		return ParserRuleContext_ROLLBACK_RHS
	case RETRY_KEYWORD:
		return ParserRuleContext_RETRY_KEYWORD_RHS
	case TRANSACTIONAL_KEYWORD:
		parentCtx = this.getParentContext()
		if parentCtx == ParserRuleContext_NAMED_WORKER_DECL {
			return ParserRuleContext_WORKER_KEYWORD
		}
		return ParserRuleContext_EXPRESSION_RHS
	case ENUM_KEYWORD:
		return ParserRuleContext_MODULE_ENUM_NAME
	case MATCH_KEYWORD:
		return ParserRuleContext_EXPRESSION
	case READONLY_KEYWORD:
		parentCtx = this.getParentContext()
		if ((parentCtx == ParserRuleContext_MAPPING_CONSTRUCTOR) || (parentCtx == ParserRuleContext_MAPPING_BP_OR_MAPPING_CONSTRUCTOR)) || (parentCtx == ParserRuleContext_MAPPING_FIELD) {
			return ParserRuleContext_SPECIFIC_FIELD
		}
		panic("next rule of readonly keyword found: " + currentCtx)
	case DISTINCT_KEYWORD:
		return ParserRuleContext_TYPE_DESCRIPTOR
	case VAR_KEYWORD:
		parentCtx = this.getParentContext()
		if ((parentCtx == ParserRuleContext_REST_MATCH_PATTERN) || (parentCtx == ParserRuleContext_ERROR_MATCH_PATTERN)) || (parentCtx == ParserRuleContext_ERROR_ARG_LIST_MATCH_PATTERN_FIRST_ARG) {
			return ParserRuleContext_VARIABLE_NAME
		}
		return ParserRuleContext_BINDING_PATTERN
	case EQUALS_KEYWORD:
		if this.getParentContext() == ParserRuleContext_ON_CLAUSE {
			panic("assertion failed")
		}
		this.endContext()
		return ParserRuleContext_EXPRESSION
	case CONFLICT_KEYWORD:
		this.endContext()
		return ParserRuleContext_EXPRESSION
	case LIMIT_KEYWORD:
		return ParserRuleContext_EXPRESSION
	case OUTER_KEYWORD:
		return ParserRuleContext_JOIN_KEYWORD
	case MAP_KEYWORD:
		parentCtx = this.getParentContext()
		if parentCtx == ParserRuleContext_TABLE_CONSTRUCTOR_OR_QUERY_EXPRESSION {
			return ParserRuleContext_QUERY_EXPRESSION
		}
		return ParserRuleContext_LT
	default:
		panic("getNextRuleForKeywords found: " + currentCtx)
	}
}

func (this *BallerinaParserErrorHandler) startContextIfRequired(currentCtx ParserRuleContext) {
	switch currentCtx {
	case COMP_UNIT,
		FUNC_DEF_OR_FUNC_TYPE,
		ANON_FUNC_EXPRESSION,
		FUNC_DEF,
		FUNC_TYPE_DESC,
		EXTERNAL_FUNC_BODY,
		FUNC_BODY_BLOCK,
		STATEMENT,
		STATEMENT_WITHOUT_ANNOTS,
		VAR_DECL_STMT,
		ASSIGNMENT_STMT,
		REQUIRED_PARAM,
		DEFAULTABLE_PARAM,
		REST_PARAM,
		MODULE_TYPE_DEFINITION,
		RECORD_FIELD,
		RECORD_TYPE_DESCRIPTOR,
		OBJECT_TYPE_DESCRIPTOR,
		ARG_LIST,
		OBJECT_FUNC_OR_FIELD,
		IF_BLOCK,
		BLOCK_STMT,
		WHILE_BLOCK,
		PANIC_STMT,
		CALL_STMT,
		IMPORT_DECL,
		CONTINUE_STATEMENT,
		BREAK_STATEMENT,
		RETURN_STMT,
		FAIL_STATEMENT,
		COMPUTED_FIELD_NAME,
		LISTENERS_LIST,
		SERVICE_DECL,
		LISTENER_DECL,
		CONSTANT_DECL,
		OPTIONAL_TYPE_DESCRIPTOR,
		ARRAY_TYPE_DESCRIPTOR,
		ANNOTATIONS,
		VARIABLE_REF,
		TYPE_REFERENCE_IN_TYPE_INCLUSION,
		TYPE_REFERENCE,
		ANNOT_REFERENCE,
		FIELD_ACCESS_IDENTIFIER,
		MAPPING_CONSTRUCTOR,
		LOCAL_TYPE_DEFINITION_STMT,
		EXPRESSION_STATEMENT,
		NIL_LITERAL,
		LOCK_STMT,
		ANNOTATION_DECL,
		ANNOT_ATTACH_POINTS_LIST,
		XML_NAMESPACE_DECLARATION,
		CONSTANT_EXPRESSION,
		NAMED_WORKER_DECL,
		FORK_STMT,
		FOREACH_STMT,
		LIST_CONSTRUCTOR,
		TYPE_CAST,
		KEY_SPECIFIER,
		LET_EXPR_LET_VAR_DECL,
		LET_CLAUSE_LET_VAR_DECL,
		ORDER_KEY_LIST,
		ROW_TYPE_PARAM,
		TABLE_CONSTRUCTOR_OR_QUERY_EXPRESSION,
		OBJECT_CONSTRUCTOR_MEMBER,
		CLASS_MEMBER,
		OBJECT_TYPE_MEMBER,
		LIST_BINDING_PATTERN,
		MAPPING_BINDING_PATTERN,
		REST_BINDING_PATTERN,
		TYPED_BINDING_PATTERN,
		BINDING_PATTERN_STARTING_IDENTIFIER,
		MULTI_RECEIVE_WORKERS,
		MULTI_WAIT_FIELDS,
		ALTERNATE_WAIT_EXPRS,
		DO_CLAUSE,
		MEMBER_ACCESS_KEY_EXPR,
		CONDITIONAL_EXPRESSION,
		DO_BLOCK,
		TRANSACTION_STMT,
		RETRY_STMT,
		ROLLBACK_STMT,
		MODULE_ENUM_DECLARATION,
		ENUM_MEMBER_LIST,
		XML_NAME_PATTERN,
		XML_ATOMIC_NAME_PATTERN,
		MATCH_STMT,
		MATCH_BODY,
		MATCH_PATTERN,
		LIST_MATCH_PATTERN,
		REST_MATCH_PATTERN,
		ERROR_BINDING_PATTERN,
		MAPPING_MATCH_PATTERN,
		ERROR_MATCH_PATTERN,
		ERROR_ARG_LIST_MATCH_PATTERN_FIRST_ARG,
		NAMED_ARG_MATCH_PATTERN,
		SELECT_CLAUSE,
		COLLECT_CLAUSE,
		JOIN_CLAUSE,
		GROUP_BY_CLAUSE,
		ON_FAIL_CLAUSE,
		BRACED_EXPR_OR_ANON_FUNC_PARAMS,
		MODULE_CLASS_DEFINITION,
		OBJECT_CONSTRUCTOR,
		ABSOLUTE_RESOURCE_PATH,
		RELATIVE_RESOURCE_PATH,
		ERROR_CONSTRUCTOR,
		CLASS_DESCRIPTOR_IN_NEW_EXPR,
		BRACED_EXPRESSION,
		CLIENT_RESOURCE_ACCESS_ACTION,
		TUPLE_MEMBERS,
		SINGLE_OR_ALTERNATE_WORKER,
		XML_STEP_EXTENDS,
		NATURAL_EXPRESSION,
		TYPE_DESC_IN_ANNOTATION_DECL,
		TYPE_DESC_BEFORE_IDENTIFIER,
		TYPE_DESC_IN_RECORD_FIELD,
		TYPE_DESC_IN_PARAM,
		TYPE_DESC_IN_TYPE_BINDING_PATTERN,
		TYPE_DESC_IN_TYPE_DEF,
		TYPE_DESC_IN_ANGLE_BRACKETS,
		TYPE_DESC_IN_RETURN_TYPE_DESC,
		TYPE_DESC_IN_EXPRESSION,
		TYPE_DESC_IN_STREAM_TYPE_DESC,
		TYPE_DESC_IN_PARENTHESIS,
		TYPE_DESC_IN_TUPLE,
		TYPE_DESC_IN_SERVICE,
		TYPE_DESC_IN_PATH_PARAM,
		TYPE_DESC_BEFORE_IDENTIFIER_IN_GROUPING_KEY:
		this.startContext(currentCtx)
		break
	default:
		break
	}
	switch currentCtx {
	case TABLE_CONSTRUCTOR,
		QUERY_EXPRESSION,
		ON_CONFLICT_CLAUSE,
		ON_CLAUSE:
		this.switchContext(currentCtx)
		break
	default:
		break
	}
}

func (this *BallerinaParserErrorHandler) getNextRuleForCloseParenthesis() ParserRuleContext {
	var parentCtx ParserRuleContext
	parentCtx = this.getParentContext()
	if parentCtx == ParserRuleContext_PARAM_LIST {
		this.endContext()
		return ParserRuleContext_FUNC_OPTIONAL_RETURNS
	} else if this.isParameter(parentCtx) {
		this.endContext()
		this.endContext()
		return ParserRuleContext_FUNC_OPTIONAL_RETURNS
	} else if parentCtx == ParserRuleContext_NIL_LITERAL {
		this.endContext()
		return this.getNextRuleForExpr()
	} else if parentCtx == ParserRuleContext_KEY_SPECIFIER {
		this.endContext()
		if this.isInTypeDescContext() {
			return ParserRuleContext_TYPE_DESC_RHS
		}
		return ParserRuleContext_TABLE_CONSTRUCTOR_OR_QUERY_RHS
	} else if parentCtx == ParserRuleContext_TYPE_DESC_IN_PARENTHESIS {
		this.endContext()
		return ParserRuleContext_TYPE_DESC_RHS
	} else if this.isInTypeDescContext() {
		return ParserRuleContext_TYPE_DESC_RHS
	} else if parentCtx == ParserRuleContext_BRACED_EXPR_OR_ANON_FUNC_PARAMS {
		this.endContext()
		return ParserRuleContext_INFER_PARAM_END_OR_PARENTHESIS_END
	} else if parentCtx == ParserRuleContext_BRACED_EXPRESSION {
		this.endContext()
		return ParserRuleContext_EXPRESSION_RHS
	} else if parentCtx == ParserRuleContext_ERROR_MATCH_PATTERN {
		this.endContext()
		return this.getNextRuleForMatchPattern()
	} else if parentCtx == ParserRuleContext_NAMED_ARG_MATCH_PATTERN {
		this.endContext()
		this.endContext()
		return this.getNextRuleForMatchPattern()
	} else if parentCtx == ParserRuleContext_ERROR_BINDING_PATTERN {
		this.endContext()
		return this.getNextRuleForBindingPattern()
	} else if parentCtx == ParserRuleContext_ERROR_ARG_LIST_MATCH_PATTERN_FIRST_ARG {
		this.endContext()
		this.endContext()
		return this.getNextRuleForBindingPattern()
	}
	return ParserRuleContext_EXPRESSION_RHS
}

func (this *BallerinaParserErrorHandler) getNextRuleForOpenParenthesis() ParserRuleContext {
	parentCtx := this.getParentContext()
	if parentCtx == ParserRuleContext_EXPRESSION_STATEMENT {
		return ParserRuleContext_EXPRESSION_STATEMENT_START
	} else if ((this.isStatement(parentCtx) || this.isExpressionContext(parentCtx)) || (parentCtx == ParserRuleContext_ARRAY_TYPE_DESCRIPTOR)) || (parentCtx == ParserRuleContext_BRACED_EXPRESSION) {
		return ParserRuleContext_EXPRESSION
	} else if ((((parentCtx == ParserRuleContext_FUNC_DEF_OR_FUNC_TYPE) || (parentCtx == ParserRuleContext_FUNC_TYPE_DESC)) || (parentCtx == ParserRuleContext_FUNC_DEF)) || (parentCtx == ParserRuleContext_ANON_FUNC_EXPRESSION)) || (parentCtx == ParserRuleContext_FUNC_TYPE_DESC_OR_ANON_FUNC) {
		this.startContext(ParserRuleContext.PARAM_LIST)
		return ParserRuleContext_PARAM_LIST
	} else if parentCtx == ParserRuleContext_NIL_LITERAL {
		return ParserRuleContext_CLOSE_PARENTHESIS
	} else if parentCtx == ParserRuleContext_KEY_SPECIFIER {
		return ParserRuleContext_KEY_SPECIFIER_RHS
	} else if this.isInTypeDescContext() {
		this.startContext(ParserRuleContext.KEY_SPECIFIER)
		return ParserRuleContext_KEY_SPECIFIER_RHS
	} else if this.isParameter(parentCtx) {
		return ParserRuleContext_EXPRESSION
	} else if parentCtx == ParserRuleContext_ERROR_MATCH_PATTERN {
		return ParserRuleContext_ERROR_ARG_LIST_MATCH_PATTERN_FIRST_ARG
	} else if this.isInMatchPatternCtx(parentCtx) {
		this.startContext(ParserRuleContext.ERROR_MATCH_PATTERN)
		return ParserRuleContext_ERROR_ARG_LIST_MATCH_PATTERN_FIRST_ARG
	} else if parentCtx == ParserRuleContext_ERROR_BINDING_PATTERN {
		return ParserRuleContext_ERROR_ARG_LIST_BINDING_PATTERN_START
	}
	return ParserRuleContext_EXPRESSION
}

func (this *BallerinaParserErrorHandler) isInMatchPatternCtx(context ParserRuleContext) bool {
	switch context {
	case MATCH_PATTERN,
		LIST_MATCH_PATTERN,
		MAPPING_MATCH_PATTERN,
		ERROR_MATCH_PATTERN,
		NAMED_ARG_MATCH_PATTERN:
		true
	default:
		false
	}
}

func (this *BallerinaParserErrorHandler) getNextRuleForOpenBrace() ParserRuleContext {
	parentCtx := this.getParentContext()
	switch parentCtx {
	case OBJECT_TYPE_DESCRIPTOR:
		ParserRuleContext_OBJECT_TYPE_MEMBER
	case MODULE_CLASS_DEFINITION:
		ParserRuleContext_CLASS_MEMBER
	case OBJECT_CONSTRUCTOR,
		SERVICE_DECL:
		ParserRuleContext_OBJECT_CONSTRUCTOR_MEMBER
	case RECORD_TYPE_DESCRIPTOR:
		ParserRuleContext_RECORD_FIELD
	case MAPPING_CONSTRUCTOR:
		ParserRuleContext_FIRST_MAPPING_FIELD
	case FORK_STMT:
		ParserRuleContext_NAMED_WORKER_DECL
	case MULTI_RECEIVE_WORKERS:
		ParserRuleContext_RECEIVE_FIELD
	case MULTI_WAIT_FIELDS:
		ParserRuleContext_WAIT_FIELD_NAME
	case MODULE_ENUM_DECLARATION:
		ParserRuleContext_ENUM_MEMBER_LIST
	case MAPPING_BINDING_PATTERN:
		ParserRuleContext_MAPPING_BINDING_PATTERN_MEMBER
	case MAPPING_MATCH_PATTERN:
		ParserRuleContext_FIELD_MATCH_PATTERNS_START
	case MATCH_BODY:
		ParserRuleContext_MATCH_PATTERN
	case NATURAL_EXPRESSION:
		ParserRuleContext_CLOSE_BRACE
	default:
		ParserRuleContext_STATEMENT
	}
}

func (this *BallerinaParserErrorHandler) isExpressionContext(ctx ParserRuleContext) bool {
	switch ctx {
	case LISTENERS_LIST,
		MAPPING_CONSTRUCTOR,
		COMPUTED_FIELD_NAME,
		LIST_CONSTRUCTOR,
		INTERPOLATION,
		ARG_LIST,
		LET_EXPR_LET_VAR_DECL,
		LET_CLAUSE_LET_VAR_DECL,
		TABLE_CONSTRUCTOR,
		QUERY_EXPRESSION,
		TABLE_CONSTRUCTOR_OR_QUERY_EXPRESSION,
		ORDER_KEY_LIST,
		GROUP_BY_CLAUSE,
		SELECT_CLAUSE,
		COLLECT_CLAUSE,
		JOIN_CLAUSE,
		ON_CONFLICT_CLAUSE:
		true
	default:
		false
	}
}

func (this *BallerinaParserErrorHandler) getNextRuleForParamType() ParserRuleContext {
	var parentCtx ParserRuleContext
	parentCtx = this.getParentContext()
	if (parentCtx == ParserRuleContext_REQUIRED_PARAM) || (parentCtx == ParserRuleContext_DEFAULTABLE_PARAM) {
		if this.hasAncestorContext(ParserRuleContext.FUNC_TYPE_DESC) {
			return ParserRuleContext_FUNC_TYPE_PARAM_RHS
		}
		return ParserRuleContext_PARAM_RHS
	} else if parentCtx == ParserRuleContext_REST_PARAM {
		return ParserRuleContext_ELLIPSIS
	} else {
		panic("getNextRuleForParamType found: " + parentCtx)
	}
}

func (this *BallerinaParserErrorHandler) getNextRuleForComma() ParserRuleContext {
	parentCtx := this.getParentContext()
	switch parentCtx {
	case PARAM_LIST,
		REQUIRED_PARAM,
		DEFAULTABLE_PARAM,
		REST_PARAM:
		this.endContext()
		parentCtx
	case ARG_LIST:
		ParserRuleContext_ARG_START
	case MAPPING_CONSTRUCTOR:
		ParserRuleContext_MAPPING_FIELD
	case LIST_CONSTRUCTOR:
		ParserRuleContext_LIST_CONSTRUCTOR_MEMBER
	case LISTENERS_LIST,
		ORDER_KEY_LIST:
		ParserRuleContext_EXPRESSION
	case GROUP_BY_CLAUSE:
		ParserRuleContext_GROUPING_KEY_LIST_ELEMENT
	case ANNOT_ATTACH_POINTS_LIST:
		ParserRuleContext_ATTACH_POINT
	case TABLE_CONSTRUCTOR:
		ParserRuleContext_MAPPING_CONSTRUCTOR
	case KEY_SPECIFIER:
		ParserRuleContext_VARIABLE_NAME
	case LET_EXPR_LET_VAR_DECL,
		LET_CLAUSE_LET_VAR_DECL:
		ParserRuleContext_LET_VAR_DECL_START
	case TYPE_DESC_IN_STREAM_TYPE_DESC:
		ParserRuleContext_TYPE_DESCRIPTOR
	case BRACED_EXPR_OR_ANON_FUNC_PARAMS:
		ParserRuleContext_IMPLICIT_ANON_FUNC_PARAM
	case TUPLE_MEMBERS:
		ParserRuleContext_TUPLE_MEMBER
	case LIST_BINDING_PATTERN:
		ParserRuleContext_LIST_BINDING_PATTERN_MEMBER
	case MAPPING_BINDING_PATTERN,
		MAPPING_BP_OR_MAPPING_CONSTRUCTOR:
		ParserRuleContext_MAPPING_BINDING_PATTERN_MEMBER
	case MULTI_RECEIVE_WORKERS:
		ParserRuleContext_RECEIVE_FIELD
	case MULTI_WAIT_FIELDS:
		ParserRuleContext_WAIT_FIELD_NAME
	case ENUM_MEMBER_LIST:
		ParserRuleContext_ENUM_MEMBER_START
	case MEMBER_ACCESS_KEY_EXPR:
		ParserRuleContext_MEMBER_ACCESS_KEY_EXPR_END
	case STMT_START_BRACKETED_LIST:
		ParserRuleContext_STMT_START_BRACKETED_LIST_MEMBER
	case BRACKETED_LIST:
		ParserRuleContext_BRACKETED_LIST_MEMBER
	case LIST_MATCH_PATTERN:
		ParserRuleContext_LIST_MATCH_PATTERN_MEMBER
	case ERROR_BINDING_PATTERN:
		ParserRuleContext_ERROR_FIELD_BINDING_PATTERN
	case MAPPING_MATCH_PATTERN:
		ParserRuleContext_FIELD_MATCH_PATTERN_MEMBER
	case ERROR_MATCH_PATTERN:
		ParserRuleContext_ERROR_FIELD_MATCH_PATTERN
	case NAMED_ARG_MATCH_PATTERN:
		this.endContext()
		ParserRuleContext_NAMED_ARG_MATCH_PATTERN_RHS
	default:
		panic("getNextRuleForComma found: " + parentCtx)
	}
}

func (this *BallerinaParserErrorHandler) getNextRuleForTypeDescriptor() ParserRuleContext {
	parentCtx := this.getParentContext()
	switch parentCtx {
	case TYPE_DESC_IN_ANNOTATION_DECL:
		this.endContext()
		if this.isInTypeDescContext() {
			return ParserRuleContext_TYPE_DESC_RHS
		}
		return ParserRuleContext_ANNOTATION_TAG
	case TYPE_DESC_BEFORE_IDENTIFIER,
		TYPE_DESC_IN_RECORD_FIELD:
		this.endContext()
		if this.isInTypeDescContext() {
			return ParserRuleContext_TYPE_DESC_RHS
		}
		return ParserRuleContext_VARIABLE_NAME
	case TYPE_DESC_IN_TYPE_BINDING_PATTERN:
		this.endContext()
		if this.isInTypeDescContext() {
			return ParserRuleContext_TYPE_DESC_RHS
		}
		return ParserRuleContext_BINDING_PATTERN
	case TYPE_DESC_IN_PARAM:
		this.endContext()
		if this.isInTypeDescContext() {
			return ParserRuleContext_TYPE_DESC_RHS
		}
		return ParserRuleContext_AFTER_PARAMETER_TYPE
	case TYPE_DESC_IN_TYPE_DEF:
		this.endContext()
		if this.isInTypeDescContext() {
			return ParserRuleContext_TYPE_DESC_RHS
		}
		return ParserRuleContext_SEMICOLON
	case TYPE_DESC_IN_ANGLE_BRACKETS:
		this.endContext()
		return ParserRuleContext_GT
	case TYPE_DESC_IN_RETURN_TYPE_DESC:
		this.endContext()
		if this.isInTypeDescContext() {
			return ParserRuleContext_TYPE_DESC_RHS
		}
		parentCtx = this.getParentContext()
		switch parentCtx {
		case FUNC_TYPE_DESC:
			this.endContext()
			return ParserRuleContext_TYPE_DESC_RHS
		case FUNC_DEF_OR_FUNC_TYPE:
			return ParserRuleContext_FUNC_BODY_OR_TYPE_DESC_RHS
		case FUNC_TYPE_DESC_OR_ANON_FUNC:
			return ParserRuleContext_FUNC_TYPE_DESC_RHS_OR_ANON_FUNC_BODY
		case FUNC_DEF:
			grandParentCtx := this.getGrandParentContext()
			if grandParentCtx == ParserRuleContext_OBJECT_TYPE_MEMBER {
				return ParserRuleContext_SEMICOLON
			} else {
				return ParserRuleContext_FUNC_BODY
			}
		case ANON_FUNC_EXPRESSION:
			return ParserRuleContext_ANON_FUNC_BODY
		case NAMED_WORKER_DECL:
			return ParserRuleContext_BLOCK_STMT
		default:
			panic("next rule of type-desc-in-return-type found: " + parentCtx)
		}
	case TYPE_DESC_IN_EXPRESSION:
		this.endContext()
		if this.isInTypeDescContext() {
			return ParserRuleContext_TYPE_DESC_RHS
		}
		return ParserRuleContext_EXPRESSION_RHS
	case COMP_UNIT:
		this.startContext(ParserRuleContext.VAR_DECL_STMT)
		return ParserRuleContext_VARIABLE_NAME
	case OBJECT_CONSTRUCTOR_MEMBER,
		CLASS_MEMBER,
		OBJECT_TYPE_MEMBER:
		return ParserRuleContext_VARIABLE_NAME
	case ANNOTATION_DECL:
		return ParserRuleContext_IDENTIFIER
	case TYPE_DESC_IN_STREAM_TYPE_DESC:
		return ParserRuleContext_STREAM_TYPE_FIRST_PARAM_RHS
	case TYPE_DESC_IN_PARENTHESIS:
		return ParserRuleContext_CLOSE_PARENTHESIS
	case TYPE_DESC_IN_TUPLE:
		this.endContext()
		return ParserRuleContext_TYPE_DESC_IN_TUPLE_RHS
	case STMT_START_BRACKETED_LIST:
		return ParserRuleContext_TYPE_DESC_IN_TUPLE_RHS
	case TYPE_REFERENCE_IN_TYPE_INCLUSION:
		this.endContext()
		return ParserRuleContext_SEMICOLON
	case TYPE_DESC_IN_SERVICE:
		this.endContext()
		return ParserRuleContext_OPTIONAL_ABSOLUTE_PATH
	case TYPE_DESC_IN_PATH_PARAM:
		this.endContext()
		return ParserRuleContext_PATH_PARAM_ELLIPSIS
	case TYPE_DESC_BEFORE_IDENTIFIER_IN_GROUPING_KEY:
		this.endContext()
		return ParserRuleContext_BINDING_PATTERN_STARTING_IDENTIFIER
	default:
		return ParserRuleContext_EXPRESSION_RHS
	}
}

func (this *BallerinaParserErrorHandler) isInTypeDescContext() bool {
	switch this.getParentContext() {
	case TYPE_DESC_IN_ANNOTATION_DECL,
		TYPE_DESC_BEFORE_IDENTIFIER,
		TYPE_DESC_BEFORE_IDENTIFIER_IN_GROUPING_KEY,
		TYPE_DESC_IN_RECORD_FIELD,
		TYPE_DESC_IN_PARAM,
		TYPE_DESC_IN_TYPE_BINDING_PATTERN,
		TYPE_DESC_IN_TYPE_DEF,
		TYPE_DESC_IN_ANGLE_BRACKETS,
		TYPE_DESC_IN_RETURN_TYPE_DESC,
		TYPE_DESC_IN_EXPRESSION,
		TYPE_DESC_IN_STREAM_TYPE_DESC,
		TYPE_DESC_IN_PARENTHESIS,
		TYPE_DESC_IN_TUPLE,
		TYPE_DESC_IN_SERVICE,
		TYPE_DESC_IN_PATH_PARAM,
		STMT_START_BRACKETED_LIST,
		BRACKETED_LIST,
		TYPE_REFERENCE_IN_TYPE_INCLUSION:
		true
	default:
		false
	}
}

func (this *BallerinaParserErrorHandler) getNextRuleForEqualOp() ParserRuleContext {
	parentCtx := this.getParentContext()
	switch parentCtx {
	case EXTERNAL_FUNC_BODY:
		ParserRuleContext_EXTERNAL_FUNC_BODY_OPTIONAL_ANNOTS
	case REQUIRED_PARAM,
		DEFAULTABLE_PARAM:
		ParserRuleContext_EXPR_START_OR_INFERRED_TYPEDESC_DEFAULT_START
	case RECORD_FIELD,
		ARG_LIST,
		OBJECT_CONSTRUCTOR_MEMBER,
		CLASS_MEMBER,
		OBJECT_TYPE_MEMBER,
		LISTENER_DECL,
		CONSTANT_DECL,
		LET_EXPR_LET_VAR_DECL,
		LET_CLAUSE_LET_VAR_DECL,
		ENUM_MEMBER_LIST,
		GROUP_BY_CLAUSE:
		ParserRuleContext_EXPRESSION
	case FUNC_DEF_OR_FUNC_TYPE:
		this.switchContext(ParserRuleContext.VAR_DECL_STMT)
		ParserRuleContext_EXPRESSION
	case NAMED_ARG_MATCH_PATTERN:
		ParserRuleContext_MATCH_PATTERN
	case ERROR_BINDING_PATTERN:
		ParserRuleContext_BINDING_PATTERN
	default:
		if this.isStatement(parentCtx) {
			ParserRuleContext_EXPRESSION
		}
		panic("getNextRuleForEqualOp found: " + parentCtx)
	}
}

func (this *BallerinaParserErrorHandler) getNextRuleForCloseBrace(nextLookahead int) ParserRuleContext {
	parentCtx := this.getParentContext()
	var nextToken Token
	switch parentCtx {
	case FUNC_BODY_BLOCK:
		this.endContext()
		return this.getNextRuleForCloseBraceInFuncBody()
	case CLASS_MEMBER:
		this.endContext()
	case SERVICE_DECL,
		MODULE_CLASS_DEFINITION:
		this.endContext()
		return ParserRuleContext_OPTIONAL_TOP_LEVEL_SEMICOLON
	case OBJECT_CONSTRUCTOR_MEMBER:
		this.endContext()
		parentCtx = this.getParentContext()
		if parentCtx == ParserRuleContext_SERVICE_DECL {
			this.endContext()
			return ParserRuleContext_TOP_LEVEL_NODE
		}
		this.endContext()
		return ParserRuleContext_EXPRESSION_RHS
	case OBJECT_TYPE_MEMBER:
		this.endContext()
	case RECORD_TYPE_DESCRIPTOR,
		OBJECT_TYPE_DESCRIPTOR:
		this.endContext()
		return ParserRuleContext_TYPE_DESC_RHS
	case BLOCK_STMT,
		AMBIGUOUS_STMT:
		this.endContext()
		parentCtx = this.getParentContext()
		switch parentCtx {
		case LOCK_STMT,
			FOREACH_STMT,
			WHILE_BLOCK,
			DO_BLOCK,
			RETRY_STMT:
			this.endContext()
			return ParserRuleContext_REGULAR_COMPOUND_STMT_RHS
		case ON_FAIL_CLAUSE:
			this.endContext()
			return ParserRuleContext_STATEMENT
		case IF_BLOCK:
			this.endContext()
			return ParserRuleContext_ELSE_BLOCK
		case TRANSACTION_STMT:
			this.endContext()
			parentCtx = this.getParentContext()
			if parentCtx == ParserRuleContext_RETRY_STMT {
				this.endContext()
			}
			return ParserRuleContext_REGULAR_COMPOUND_STMT_RHS
		case NAMED_WORKER_DECL:
			this.endContext()
			parentCtx = this.getParentContext()
			if parentCtx == ParserRuleContext_FORK_STMT {
				nextToken = this.this.tokenReader.peek(nextLookahead)
				switch nextToken.kind {
				case CLOSE_BRACE_TOKEN:
					ParserRuleContext_CLOSE_BRACE
				default:
					ParserRuleContext_REGULAR_COMPOUND_STMT_RHS
				}
			} else {
				return ParserRuleContext_REGULAR_COMPOUND_STMT_RHS
			}
		case MATCH_BODY:
			return ParserRuleContext_MATCH_PATTERN
		case DO_CLAUSE:
			return ParserRuleContext_QUERY_EXPRESSION_END
		default:
			return ParserRuleContext_STATEMENT
		}
	case MAPPING_CONSTRUCTOR:
		this.endContext()
		parentCtx = this.getParentContext()
		if parentCtx == ParserRuleContext_TABLE_CONSTRUCTOR {
			return ParserRuleContext_TABLE_ROW_END
		}
		if parentCtx == ParserRuleContext_ANNOTATIONS {
			return ParserRuleContext_ANNOTATION_END
		}
		return this.getNextRuleForExpr()
	case STMT_START_BRACKETED_LIST:
		return ParserRuleContext_BRACKETED_LIST_MEMBER_END
	case MAPPING_BINDING_PATTERN,
		MAPPING_BP_OR_MAPPING_CONSTRUCTOR:
		this.endContext()
		return this.getNextRuleForBindingPattern()
	case FORK_STMT:
		this.endContext()
		return ParserRuleContext_STATEMENT
	case INTERPOLATION:
		this.endContext()
		return ParserRuleContext_TEMPLATE_MEMBER
	case OBJECT_CONSTRUCTOR,
		MULTI_RECEIVE_WORKERS,
		MULTI_WAIT_FIELDS,
		NATURAL_EXPRESSION:
		this.endContext()
		return ParserRuleContext_EXPRESSION_RHS
	case ENUM_MEMBER_LIST:
		this.endContext()
		this.endContext()
		return ParserRuleContext_OPTIONAL_TOP_LEVEL_SEMICOLON
	case MATCH_BODY:
		this.endContext()
		this.endContext()
		return ParserRuleContext_REGULAR_COMPOUND_STMT_RHS
	case MAPPING_MATCH_PATTERN:
		this.endContext()
		return this.getNextRuleForMatchPattern()
	case MATCH_STMT:
		this.endContext()
		return ParserRuleContext_REGULAR_COMPOUND_STMT_RHS
	default:
		panic("getNextRuleForCloseBrace found: " + parentCtx)
	}
}

func (this *BallerinaParserErrorHandler) getNextRuleForCloseBraceInFuncBody() ParserRuleContext {
	var parentCtx ParserRuleContext
	parentCtx = this.getParentContext()
	switch parentCtx {
	case OBJECT_CONSTRUCTOR_MEMBER:
		ParserRuleContext_OBJECT_CONSTRUCTOR_MEMBER_START
	case CLASS_MEMBER,
		OBJECT_TYPE_MEMBER:
		ParserRuleContext_CLASS_MEMBER_OR_OBJECT_MEMBER_START
	case COMP_UNIT:
		ParserRuleContext_OPTIONAL_TOP_LEVEL_SEMICOLON
	case FUNC_DEF,
		FUNC_DEF_OR_FUNC_TYPE:
		this.endContext()
		this.getNextRuleForCloseBraceInFuncBody()
	default:
		this.endContext()
		ParserRuleContext_EXPRESSION_RHS
	}
}

func (this *BallerinaParserErrorHandler) getNextRuleForAnnotationEnd(nextLookahead int) ParserRuleContext {
	var parentCtx ParserRuleContext
	var nextToken Token
	nextToken = this.this.tokenReader.peek(nextLookahead)
	if nextToken.kind == SyntaxKind_AT_TOKEN {
		return ParserRuleContext_AT
	}
	this.endContext()
	parentCtx = this.getParentContext()
	switch parentCtx {
	case COMP_UNIT:
		ParserRuleContext_TOP_LEVEL_NODE_WITHOUT_METADATA
	case FUNC_DEF,
		FUNC_TYPE_DESC,
		FUNC_DEF_OR_FUNC_TYPE,
		ANON_FUNC_EXPRESSION,
		FUNC_TYPE_DESC_OR_ANON_FUNC:
		ParserRuleContext_TYPE_DESC_IN_RETURN_TYPE_DESC
	case LET_EXPR_LET_VAR_DECL,
		LET_CLAUSE_LET_VAR_DECL:
		ParserRuleContext_TYPE_DESC_IN_TYPE_BINDING_PATTERN
	case RECORD_FIELD:
		ParserRuleContext_RECORD_FIELD_WITHOUT_METADATA
	case OBJECT_CONSTRUCTOR_MEMBER:
		ParserRuleContext_OBJECT_CONS_MEMBER_WITHOUT_META
	case CLASS_MEMBER,
		OBJECT_TYPE_MEMBER:
		ParserRuleContext_CLASS_MEMBER_OR_OBJECT_MEMBER_WITHOUT_META
	case FUNC_BODY_BLOCK:
		ParserRuleContext_STATEMENT_WITHOUT_ANNOTS
	case EXTERNAL_FUNC_BODY:
		ParserRuleContext_EXTERNAL_KEYWORD
	case TYPE_CAST:
		ParserRuleContext_TYPE_CAST_PARAM_RHS
	case ENUM_MEMBER_LIST:
		ParserRuleContext_ENUM_MEMBER_NAME
	case RELATIVE_RESOURCE_PATH:
		ParserRuleContext_TYPE_DESC_IN_PATH_PARAM
	case TUPLE_MEMBERS:
		ParserRuleContext_TUPLE_MEMBER
	default:
		if this.isParameter(parentCtx) {
			ParserRuleContext_TYPE_DESC_IN_PARAM
		}
		ParserRuleContext_EXPRESSION
	}
}

func (this *BallerinaParserErrorHandler) getNextRuleForVarName() ParserRuleContext {
	parentCtx := this.getParentContext()
	switch parentCtx {
	case ASSIGNMENT_STMT:
		return ParserRuleContext_ASSIGNMENT_STMT_RHS
	case CALL_STMT:
		return ParserRuleContext_ARG_LIST
	case REQUIRED_PARAM,
		PARAM_LIST:
		return ParserRuleContext_REQUIRED_PARAM_NAME_RHS
	case DEFAULTABLE_PARAM:
		return ParserRuleContext_ASSIGN_OP
	case REST_PARAM:
		return ParserRuleContext_PARAM_END
	case FOREACH_STMT:
		return ParserRuleContext_IN_KEYWORD
	case BINDING_PATTERN_STARTING_IDENTIFIER:
		return this.getNextRuleForBindingPattern(true)
	case TYPED_BINDING_PATTERN,
		LIST_BINDING_PATTERN,
		STMT_START_BRACKETED_LIST_MEMBER,
		REST_BINDING_PATTERN,
		FIELD_BINDING_PATTERN,
		MAPPING_BINDING_PATTERN,
		MAPPING_BP_OR_MAPPING_CONSTRUCTOR,
		ERROR_BINDING_PATTERN:
		return this.getNextRuleForBindingPattern()
	case LISTENER_DECL,
		CONSTANT_DECL:
		return ParserRuleContext_VAR_DECL_STMT_RHS
	case RECORD_FIELD:
		return ParserRuleContext_FIELD_DESCRIPTOR_RHS
	case ARG_LIST:
		return ParserRuleContext_NAMED_OR_POSITIONAL_ARG_RHS
	case OBJECT_CONSTRUCTOR_MEMBER,
		CLASS_MEMBER,
		OBJECT_TYPE_MEMBER:
		return ParserRuleContext_OBJECT_FIELD_RHS
	case ARRAY_TYPE_DESCRIPTOR:
		return ParserRuleContext_CLOSE_BRACKET
	case KEY_SPECIFIER:
		return ParserRuleContext_TABLE_KEY_RHS
	case LET_EXPR_LET_VAR_DECL,
		LET_CLAUSE_LET_VAR_DECL:
		return ParserRuleContext_ASSIGN_OP
	case ANNOTATION_DECL:
		return ParserRuleContext_ANNOT_OPTIONAL_ATTACH_POINTS
	case QUERY_EXPRESSION,
		JOIN_CLAUSE:
		return ParserRuleContext_IN_KEYWORD
	case REST_MATCH_PATTERN:
		this.endContext()
		parentCtx = this.getParentContext()
		if parentCtx == ParserRuleContext_MAPPING_MATCH_PATTERN {
			return ParserRuleContext_CLOSE_BRACE
		}
		if (parentCtx == ParserRuleContext_ERROR_MATCH_PATTERN) || (parentCtx == ParserRuleContext_ERROR_ARG_LIST_MATCH_PATTERN_FIRST_ARG) {
			return ParserRuleContext_CLOSE_PARENTHESIS
		}
		return ParserRuleContext_CLOSE_BRACKET
	case MAPPING_MATCH_PATTERN:
		return ParserRuleContext_COLON
	case ON_FAIL_CLAUSE:
		return ParserRuleContext_BLOCK_STMT
	case ERROR_ARG_LIST_MATCH_PATTERN_FIRST_ARG:
		this.endContext()
		return ParserRuleContext_ERROR_MESSAGE_MATCH_PATTERN_END
	case ERROR_MATCH_PATTERN:
		return ParserRuleContext_ERROR_FIELD_MATCH_PATTERN_RHS
	case RELATIVE_RESOURCE_PATH:
		return ParserRuleContext_CLOSE_BRACKET
	case GROUP_BY_CLAUSE:
		return ParserRuleContext_GROUPING_KEY_LIST_ELEMENT_END
	default:
		if this.isStatement(parentCtx) {
			return ParserRuleContext_VAR_DECL_STMT_RHS
		}
		panic("getNextRuleForVarName found: " + parentCtx)
	}
}

func (this *BallerinaParserErrorHandler) getNextRuleForSemicolon(nextLookahead int) ParserRuleContext {
	var nextToken Token
	parentCtx := this.getParentContext()
	if parentCtx == ParserRuleContext_EXTERNAL_FUNC_BODY {
		this.endContext()
		return this.getNextRuleForSemicolon(nextLookahead)
	} else if parentCtx == ParserRuleContext_QUERY_EXPRESSION {
		this.endContext()
		return this.getNextRuleForSemicolon(nextLookahead)
	} else if this.isExpressionContext(parentCtx) {
		this.endContext()
		return ParserRuleContext_STATEMENT
	} else if parentCtx == ParserRuleContext_VAR_DECL_STMT {
		this.endContext()
		parentCtx = this.getParentContext()
		if parentCtx == ParserRuleContext_COMP_UNIT {
			return ParserRuleContext_TOP_LEVEL_NODE
		}
		return ParserRuleContext_STATEMENT
	} else if this.isStatement(parentCtx) {
		this.endContext()
		return ParserRuleContext_STATEMENT
	} else if parentCtx == ParserRuleContext_RECORD_FIELD {
		this.endContext()
		return ParserRuleContext_RECORD_FIELD_OR_RECORD_END
	} else if parentCtx == ParserRuleContext_XML_NAMESPACE_DECLARATION {
		this.endContext()
		parentCtx = this.getParentContext()
		if parentCtx == ParserRuleContext_COMP_UNIT {
			return ParserRuleContext_TOP_LEVEL_NODE
		}
		return ParserRuleContext_STATEMENT
	} else if ((parentCtx == ParserRuleContext_MODULE_TYPE_DEFINITION) || (parentCtx == ParserRuleContext_LISTENER_DECL)) || (parentCtx == ParserRuleContext_ANNOTATION_DECL) {
		this.endContext()
		return ParserRuleContext_TOP_LEVEL_NODE
	} else if parentCtx == ParserRuleContext_CONSTANT_DECL {
		this.endContext()
		parentCtx = this.getParentContext()
		if parentCtx == ParserRuleContext_FUNC_BODY_BLOCK {
			return ParserRuleContext_STATEMENT
		}
		return ParserRuleContext_TOP_LEVEL_NODE
	} else if ((parentCtx == ParserRuleContext_OBJECT_CONSTRUCTOR_MEMBER) || (parentCtx == ParserRuleContext_CLASS_MEMBER)) || (parentCtx == ParserRuleContext_OBJECT_TYPE_MEMBER) {
		if this.isEndOfObjectTypeNode(nextLookahead) {
			this.endContext()
			return ParserRuleContext_CLOSE_BRACE
		}
		if parentCtx == ParserRuleContext_OBJECT_CONSTRUCTOR_MEMBER {
			return ParserRuleContext_OBJECT_CONSTRUCTOR_MEMBER_START
		} else {
			return ParserRuleContext_CLASS_MEMBER_OR_OBJECT_MEMBER_START
		}
	} else if parentCtx == ParserRuleContext_IMPORT_DECL {
		this.endContext()
		nextToken = this.this.tokenReader.peek(nextLookahead)
		if nextToken.kind == SyntaxKind_EOF_TOKEN {
			return ParserRuleContext_EOF
		}
		return ParserRuleContext_TOP_LEVEL_NODE
	} else if parentCtx == ParserRuleContext_ANNOT_ATTACH_POINTS_LIST {
		this.endContext()
		this.endContext()
		nextToken = this.this.tokenReader.peek(nextLookahead)
		if nextToken.kind == SyntaxKind_EOF_TOKEN {
			return ParserRuleContext_EOF
		}
		return ParserRuleContext_TOP_LEVEL_NODE
	} else if (parentCtx == ParserRuleContext_FUNC_DEF) || (parentCtx == ParserRuleContext_FUNC_DEF_OR_FUNC_TYPE) {
		this.endContext()
		nextToken = this.this.tokenReader.peek(nextLookahead)
		if nextToken.kind == SyntaxKind_EOF_TOKEN {
			return ParserRuleContext_EOF
		}
		return this.getNextRuleForSemicolon(nextLookahead)
	} else if parentCtx == ParserRuleContext_MODULE_CLASS_DEFINITION {
		return ParserRuleContext_CLASS_MEMBER
	} else if parentCtx == ParserRuleContext_OBJECT_CONSTRUCTOR {
		return ParserRuleContext_OBJECT_CONSTRUCTOR_MEMBER
	} else if parentCtx == ParserRuleContext_COMP_UNIT {
		return ParserRuleContext_TOP_LEVEL_NODE
	} else {
		panic("getNextRuleForSemicolon found: " + parentCtx)
	}
}

func (this *BallerinaParserErrorHandler) getNextRuleForDot() ParserRuleContext {
	parentCtx := this.getParentContext()
	switch parentCtx {
	case IMPORT_DECL:
		ParserRuleContext_IMPORT_MODULE_NAME
	case RELATIVE_RESOURCE_PATH:
		ParserRuleContext_RESOURCE_ACCESSOR_DEF_OR_DECL_RHS
	case CLIENT_RESOURCE_ACCESS_ACTION:
		ParserRuleContext_METHOD_NAME
	case XML_STEP_EXTENDS:
		ParserRuleContext_METHOD_NAME
	default:
		ParserRuleContext_FIELD_ACCESS_IDENTIFIER
	}
}

func (this *BallerinaParserErrorHandler) getNextRuleForQuestionMark() ParserRuleContext {
	parentCtx := this.getParentContext()
	switch parentCtx {
	case OPTIONAL_TYPE_DESCRIPTOR:
		this.endContext()
		ParserRuleContext_TYPE_DESC_RHS
	case CONDITIONAL_EXPRESSION:
		ParserRuleContext_EXPRESSION
	default:
		ParserRuleContext_SEMICOLON
	}
}

func (this *BallerinaParserErrorHandler) getNextRuleForOpenBracket() ParserRuleContext {
	parentCtx := this.getParentContext()
	switch parentCtx {
	case ARRAY_TYPE_DESCRIPTOR:
		ParserRuleContext_ARRAY_LENGTH
	case LIST_CONSTRUCTOR:
		ParserRuleContext_LIST_CONSTRUCTOR_FIRST_MEMBER
	case TABLE_CONSTRUCTOR:
		ParserRuleContext_ROW_LIST_RHS
	case LIST_BINDING_PATTERN:
		ParserRuleContext_LIST_BINDING_PATTERNS_START
	case LIST_MATCH_PATTERN:
		ParserRuleContext_LIST_MATCH_PATTERNS_START
	case RELATIVE_RESOURCE_PATH:
		ParserRuleContext_PATH_PARAM_OPTIONAL_ANNOTS
	case CLIENT_RESOURCE_ACCESS_ACTION:
		ParserRuleContext_COMPUTED_SEGMENT_OR_REST_SEGMENT
	default:
		if this.isInTypeDescContext() {
			ParserRuleContext_TUPLE_MEMBERS
		}
		ParserRuleContext_EXPRESSION
	}
}

func (this *BallerinaParserErrorHandler) getNextRuleForCloseBracket() ParserRuleContext {
	parentCtx := this.getParentContext()
	switch parentCtx {
	case ARRAY_TYPE_DESCRIPTOR,
		TUPLE_MEMBERS:
		this.endContext()
		parentCtx = this.getParentContext()
		if parentCtx == ParserRuleContext_STMT_START_BRACKETED_LIST {
			return this.getNextRuleForCloseBracket()
		}
		return ParserRuleContext_TYPE_DESC_RHS
	case COMPUTED_FIELD_NAME:
		this.endContext()
		return ParserRuleContext_COLON
	case LIST_BINDING_PATTERN:
		this.endContext()
		return this.getNextRuleForBindingPattern()
	case LIST_CONSTRUCTOR,
		TABLE_CONSTRUCTOR,
		MEMBER_ACCESS_KEY_EXPR:
		this.endContext()
		parentCtx = this.getParentContext()
		if parentCtx == ParserRuleContext_XML_STEP_EXTENDS {
			return ParserRuleContext_XML_STEP_EXTEND
		}
		return this.getNextRuleForExpr()
	case STMT_START_BRACKETED_LIST:
		this.endContext()
		parentCtx = this.getParentContext()
		if parentCtx == ParserRuleContext_STMT_START_BRACKETED_LIST {
			return ParserRuleContext_BRACKETED_LIST_MEMBER_END
		}
		return ParserRuleContext_STMT_START_BRACKETED_LIST_RHS
	case BRACKETED_LIST:
		this.endContext()
		return ParserRuleContext_BRACKETED_LIST_RHS
	case LIST_MATCH_PATTERN:
		this.endContext()
		return this.getNextRuleForMatchPattern()
	case RELATIVE_RESOURCE_PATH:
		return ParserRuleContext_RELATIVE_RESOURCE_PATH_END
	case CLIENT_RESOURCE_ACCESS_ACTION:
		return ParserRuleContext_RESOURCE_ACCESS_SEGMENT_RHS
	default:
		return this.getNextRuleForExpr()
	}
}

func (this *BallerinaParserErrorHandler) getNextRuleForDecimalIntegerLiteral() ParserRuleContext {
	parentCtx := this.getParentContext()
	switch parentCtx {
	case CONSTANT_EXPRESSION:
		this.endContext()
		this.getNextRuleForConstExpr()
	default:
		ParserRuleContext_CLOSE_BRACKET
	}
}

func (this *BallerinaParserErrorHandler) getNextRuleForExpr() ParserRuleContext {
	var parentCtx ParserRuleContext
	parentCtx = this.getParentContext()
	if parentCtx == ParserRuleContext_CONSTANT_EXPRESSION {
		this.endContext()
		return this.getNextRuleForConstExpr()
	}
	return ParserRuleContext_EXPRESSION_RHS
}

func (this *BallerinaParserErrorHandler) getNextRuleForExprStartsWithVarRef() ParserRuleContext {
	var parentCtx ParserRuleContext
	parentCtx = this.getParentContext()
	if parentCtx == ParserRuleContext_CONSTANT_EXPRESSION {
		this.endContext()
		return this.getNextRuleForConstExpr()
	} else if parentCtx == ParserRuleContext_ARRAY_TYPE_DESCRIPTOR {
		return ParserRuleContext_CLOSE_BRACKET
	} else if parentCtx == ParserRuleContext_CALL_STMT {
		return ParserRuleContext_ARG_LIST_OPEN_PAREN
	}
	return ParserRuleContext_VARIABLE_REF_RHS
}

func (this *BallerinaParserErrorHandler) getNextRuleForConstExpr() ParserRuleContext {
	parentCtx := this.getParentContext()
	switch parentCtx {
	case XML_NAMESPACE_DECLARATION:
		ParserRuleContext_XML_NAMESPACE_PREFIX_DECL
	default:
		if this.isInTypeDescContext() {
			ParserRuleContext_TYPE_DESC_RHS
		}
		this.getNextRuleForMatchPattern()
	}
}

func (this *BallerinaParserErrorHandler) getNextRuleForLt() ParserRuleContext {
	parentCtx := this.getParentContext()
	switch parentCtx {
	case TYPE_CAST:
		ParserRuleContext_TYPE_CAST_PARAM
	default:
		ParserRuleContext_TYPE_DESC_IN_ANGLE_BRACKETS
	}
}

func (this *BallerinaParserErrorHandler) getNextRuleForGt() ParserRuleContext {
	parentCtx := this.getParentContext()
	if parentCtx == ParserRuleContext_TYPE_DESC_IN_STREAM_TYPE_DESC {
		this.endContext()
		parentCtx = this.getParentContext()
		if parentCtx == ParserRuleContext_CLASS_DESCRIPTOR_IN_NEW_EXPR {
			this.endContext()
			return ParserRuleContext_ARG_LIST_OPEN_PAREN
		}
		return ParserRuleContext_TYPE_DESC_RHS
	}
	if this.isInTypeDescContext() {
		return ParserRuleContext_TYPE_DESC_RHS
	}
	if parentCtx == ParserRuleContext_ROW_TYPE_PARAM {
		this.endContext()
		return ParserRuleContext_TABLE_TYPE_DESC_RHS
	} else if parentCtx == ParserRuleContext_RETRY_STMT {
		return ParserRuleContext_RETRY_TYPE_PARAM_RHS
	}
	if parentCtx == ParserRuleContext_XML_NAME_PATTERN {
		this.endContext()
		parentCtx = this.getParentContext()
		if parentCtx == ParserRuleContext_XML_STEP_EXTENDS {
			return ParserRuleContext_XML_STEP_EXTEND
		}
		return ParserRuleContext_XML_STEP_START_END
	}
	this.endContext()
	return ParserRuleContext_EXPRESSION
}

func (this *BallerinaParserErrorHandler) getNextRuleForBindingPattern() ParserRuleContext {
	return this.getNextRuleForBindingPattern(false)
}

func (this *BallerinaParserErrorHandler) getNextRuleForBindingPattern(isCaptureBP bool) ParserRuleContext {
	parentCtx := this.getParentContext()
	switch parentCtx {
	case BINDING_PATTERN_STARTING_IDENTIFIER,
		TYPED_BINDING_PATTERN:
		this.endContext()
		return this.getNextRuleForBindingPattern(isCaptureBP)
	case FOREACH_STMT,
		QUERY_EXPRESSION,
		JOIN_CLAUSE:
		return ParserRuleContext_IN_KEYWORD
	case LIST_BINDING_PATTERN,
		STMT_START_BRACKETED_LIST,
		BRACKETED_LIST:
		return ParserRuleContext_LIST_BINDING_PATTERN_MEMBER_END
	case MAPPING_BINDING_PATTERN,
		MAPPING_BP_OR_MAPPING_CONSTRUCTOR:
		return ParserRuleContext_MAPPING_BINDING_PATTERN_END
	case REST_BINDING_PATTERN:
		this.endContext()
		parentCtx = this.getParentContext()
		if parentCtx == ParserRuleContext_LIST_BINDING_PATTERN {
			return ParserRuleContext_CLOSE_BRACKET
		} else if parentCtx == ParserRuleContext_ERROR_BINDING_PATTERN {
			return ParserRuleContext_CLOSE_PARENTHESIS
		}
		return ParserRuleContext_CLOSE_BRACE
	case AMBIGUOUS_STMT:
		this.switchContext(ParserRuleContext.VAR_DECL_STMT)
		if isCaptureBP {
			return ParserRuleContext.VAR_DECL_STMT_RHS
		} else {
			return ParserRuleContext.ASSIGN_OP
		}
	case ASSIGNMENT_OR_VAR_DECL_STMT,
		VAR_DECL_STMT:
		if isCaptureBP {
			return ParserRuleContext.VAR_DECL_STMT_RHS
		} else {
			return ParserRuleContext.ASSIGN_OP
		}
	case LET_CLAUSE_LET_VAR_DECL,
		LET_EXPR_LET_VAR_DECL,
		ASSIGNMENT_STMT,
		GROUP_BY_CLAUSE:
		return ParserRuleContext_ASSIGN_OP
	case MATCH_PATTERN:
		return ParserRuleContext_MATCH_PATTERN_LIST_MEMBER_RHS
	case LIST_MATCH_PATTERN:
		return ParserRuleContext_LIST_MATCH_PATTERN_MEMBER_RHS
	case ERROR_BINDING_PATTERN:
		return ParserRuleContext_ERROR_FIELD_BINDING_PATTERN_END
	case ERROR_ARG_LIST_MATCH_PATTERN_FIRST_ARG:
		this.endContext()
		return ParserRuleContext_ERROR_FIELD_MATCH_PATTERN_RHS
	case ON_FAIL_CLAUSE:
		return ParserRuleContext_BLOCK_STMT
	default:
		return this.getNextRuleForMatchPattern()
	}
}

func (this *BallerinaParserErrorHandler) getNextRuleForWaitExprListEnd() ParserRuleContext {
	this.endContext()
	return ParserRuleContext_EXPRESSION_RHS
}

func (this *BallerinaParserErrorHandler) getNextRuleForIdentifier() ParserRuleContext {
	var parentCtx ParserRuleContext
	parentCtx = this.getParentContext()
	switch parentCtx {
	case VARIABLE_REF:
		this.endContext()
		return this.getNextRuleForExprStartsWithVarRef()
	case TYPE_REFERENCE:
		this.endContext()
		return this.getNextRuleForTypeReference()
	case TYPE_REFERENCE_IN_TYPE_INCLUSION:
		this.endContext()
		if this.isInTypeDescContext() {
			return ParserRuleContext_TYPE_DESC_RHS
		}
		return ParserRuleContext_SEMICOLON
	case ANNOT_REFERENCE:
		this.endContext()
		return ParserRuleContext_ANNOTATION_REF_RHS
	case ANNOTATION_DECL:
		return ParserRuleContext_ANNOT_OPTIONAL_ATTACH_POINTS
	case FIELD_ACCESS_IDENTIFIER:
		this.endContext()
		return ParserRuleContext_VARIABLE_REF_RHS
	case XML_ATOMIC_NAME_PATTERN:
		this.endContext()
		return ParserRuleContext_XML_NAME_PATTERN_RHS
	case NAMED_ARG_MATCH_PATTERN:
		return ParserRuleContext_ASSIGN_OP
	case MODULE_CLASS_DEFINITION:
		return ParserRuleContext_OPEN_BRACE
	case COMP_UNIT:
		return ParserRuleContext_TOP_LEVEL_NODE
	case OBJECT_CONSTRUCTOR_MEMBER,
		CLASS_MEMBER,
		OBJECT_TYPE_MEMBER:
		return ParserRuleContext_SEMICOLON
	case ABSOLUTE_RESOURCE_PATH:
		return ParserRuleContext_ABSOLUTE_RESOURCE_PATH_END
	case CLIENT_RESOURCE_ACCESS_ACTION:
		return ParserRuleContext_RESOURCE_ACCESS_SEGMENT_RHS
	case XML_STEP_EXTENDS:
		return ParserRuleContext_ARG_LIST_OPEN_PAREN
	default:
		if this.isInTypeDescContext() {
			return ParserRuleContext_TYPE_DESC_RHS
		}
		panic("getNextRuleForIdentifier found: " + parentCtx)
	}
}

func (this *BallerinaParserErrorHandler) getNextRuleForColon() ParserRuleContext {
	var parentCtx ParserRuleContext
	parentCtx = this.getParentContext()
	switch parentCtx {
	case MAPPING_CONSTRUCTOR:
		ParserRuleContext_EXPRESSION
	case MULTI_RECEIVE_WORKERS:
		ParserRuleContext_PEER_WORKER_NAME
	case MULTI_WAIT_FIELDS:
		ParserRuleContext_EXPRESSION
	case CONDITIONAL_EXPRESSION:
		this.endContext()
		ParserRuleContext_EXPRESSION
	case MAPPING_BINDING_PATTERN,
		MAPPING_BP_OR_MAPPING_CONSTRUCTOR:
		ParserRuleContext_VARIABLE_NAME
	case FIELD_BINDING_PATTERN:
		this.endContext()
		ParserRuleContext_VARIABLE_NAME
	case XML_ATOMIC_NAME_PATTERN:
		ParserRuleContext_XML_ATOMIC_NAME_IDENTIFIER_RHS
	case MAPPING_MATCH_PATTERN:
		ParserRuleContext_MATCH_PATTERN
	default:
		ParserRuleContext_IDENTIFIER
	}
}

func (this *BallerinaParserErrorHandler) getNextRuleForMatchPattern() ParserRuleContext {
	parentCtx := this.getParentContext()
	switch parentCtx {
	case LIST_MATCH_PATTERN:
		ParserRuleContext_LIST_MATCH_PATTERN_MEMBER_RHS
	case MAPPING_MATCH_PATTERN:
		ParserRuleContext_FIELD_MATCH_PATTERN_MEMBER_RHS
	case MATCH_PATTERN:
		ParserRuleContext_MATCH_PATTERN_LIST_MEMBER_RHS
	case ERROR_MATCH_PATTERN,
		NAMED_ARG_MATCH_PATTERN:
		ParserRuleContext_ERROR_FIELD_MATCH_PATTERN_RHS
	case ERROR_ARG_LIST_MATCH_PATTERN_FIRST_ARG:
		this.endContext()
		ParserRuleContext_ERROR_MESSAGE_MATCH_PATTERN_END
	default:
		ParserRuleContext_OPTIONAL_MATCH_GUARD
	}
}

func (this *BallerinaParserErrorHandler) getNextRuleForTypeReference() ParserRuleContext {
	parentCtx := this.getParentContext()
	switch parentCtx {
	case ERROR_CONSTRUCTOR:
		ParserRuleContext_ARG_LIST_OPEN_PAREN
	case OBJECT_CONSTRUCTOR:
		ParserRuleContext_OPEN_BRACE
	case ERROR_MATCH_PATTERN,
		ERROR_BINDING_PATTERN:
		ParserRuleContext_OPEN_PARENTHESIS
	case CLASS_DESCRIPTOR_IN_NEW_EXPR:
		this.endContext()
		ParserRuleContext_ARG_LIST_OPEN_PAREN
	default:
		if this.isInTypeDescContext() {
			ParserRuleContext_TYPE_DESC_RHS
		}
		panic("getNextRuleForTypeReference found: " + parentCtx)
	}
}

func (this *BallerinaParserErrorHandler) getNextRuleForErrorKeyword() ParserRuleContext {
	if this.isInTypeDescContext() {
		return ParserRuleContext_LT
	}
	parentCtx := this.getParentContext()
	switch parentCtx {
	case ERROR_MATCH_PATTERN:
		ParserRuleContext_ERROR_MATCH_PATTERN_ERROR_KEYWORD_RHS
	case ERROR_BINDING_PATTERN:
		ParserRuleContext_ERROR_BINDING_PATTERN_ERROR_KEYWORD_RHS
	case ERROR_CONSTRUCTOR:
		ParserRuleContext_ERROR_CONSTRUCTOR_RHS
	default:
		ParserRuleContext_ARG_LIST_OPEN_PAREN
	}
}

func (this *BallerinaParserErrorHandler) getNextRuleForFuncTypeFuncKeywordRhs() ParserRuleContext {
	parentCtx := this.getParentContext()
	if parentCtx == ParserRuleContext_FUNC_DEF_OR_FUNC_TYPE {
		this.endContext()
		parentCtx = this.getParentContext()
		switch parentCtx {
		case OBJECT_TYPE_MEMBER,
			CLASS_MEMBER,
			OBJECT_CONSTRUCTOR_MEMBER:
			this.startContext(ParserRuleContext.TYPE_DESC_BEFORE_IDENTIFIER)
			break
		case COMP_UNIT:
			fallthrough
		default:
			this.startContext(ParserRuleContext.VAR_DECL_STMT)
			this.startContext(ParserRuleContext.TYPE_DESC_IN_TYPE_BINDING_PATTERN)
			break
		}
	} else if this.getGrandParentContext() == ParserRuleContext_OBJECT_TYPE_MEMBER {
		this.switchContext(ParserRuleContext.TYPE_DESC_BEFORE_IDENTIFIER)
	}
	if this.isInTypeDescContext() {
		panic("assertion failed")
	}
	this.startContext(ParserRuleContext.FUNC_TYPE_DESC)
	return ParserRuleContext_FUNC_TYPE_FUNC_KEYWORD_RHS_START
}

func (this *BallerinaParserErrorHandler) getNextRuleForAction() ParserRuleContext {
	parentCtx := this.getParentContext()
	switch parentCtx {
	case MATCH_STMT:
		ParserRuleContext_MATCH_BODY
	case FOREACH_STMT:
		ParserRuleContext_BLOCK_STMT
	default:
		ParserRuleContext_SEMICOLON
	}
}

func (this *BallerinaParserErrorHandler) isStatement(parentCtx ParserRuleContext) bool {
	switch parentCtx {
	case STATEMENT,
		STATEMENT_WITHOUT_ANNOTS,
		VAR_DECL_STMT,
		ASSIGNMENT_STMT,
		ASSIGNMENT_OR_VAR_DECL_STMT,
		IF_BLOCK,
		BLOCK_STMT,
		WHILE_BLOCK,
		DO_BLOCK,
		CALL_STMT,
		PANIC_STMT,
		CONTINUE_STATEMENT,
		BREAK_STATEMENT,
		RETURN_STMT,
		FAIL_STATEMENT,
		LOCAL_TYPE_DEFINITION_STMT,
		EXPRESSION_STATEMENT,
		LOCK_STMT,
		FORK_STMT,
		FOREACH_STMT,
		TRANSACTION_STMT,
		RETRY_STMT,
		ROLLBACK_STMT,
		AMBIGUOUS_STMT,
		MATCH_STMT:
		true
	default:
		false
	}
}

func (this *BallerinaParserErrorHandler) isBinaryOperator(token Token) bool {
	switch token.kind {
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
		DOUBLE_LT_TOKEN,
		DOUBLE_GT_TOKEN,
		TRIPPLE_GT_TOKEN,
		ELLIPSIS_TOKEN,
		DOUBLE_DOT_LT_TOKEN,
		ELVIS_TOKEN:
		true
	case RIGHT_ARROW_TOKEN,
		RIGHT_DOUBLE_ARROW_TOKEN:
		true
	default:
		false
	}
}

func (this *BallerinaParserErrorHandler) isParameter(ctx ParserRuleContext) bool {
	switch ctx {
	case REQUIRED_PARAM, DEFAULTABLE_PARAM, REST_PARAM, PARAM_LIST:
		true
	default:
		false
	}
}

func (this *BallerinaParserErrorHandler) getInsertSolution(ctx ParserRuleContext) Solution {
	kind := this.getExpectedTokenKind(ctx)
	if kind != SyntaxKind_NONE {
		return nil
	}
	if this.hasAlternativePaths(ctx) {
		ctx = this.getShortestAlternative(ctx)
		return this.getInsertSolution(ctx)
	}
	ctx = this.getNextRule(ctx, 1)
	return this.getInsertSolution(ctx)
}

func (this *BallerinaParserErrorHandler) getExpectedTokenKind(ctx ParserRuleContext) common.SyntaxKind {
	switch ctx {
	case EXTERNAL_FUNC_BODY:
		SyntaxKind_EQUAL_TOKEN
	case FUNC_BODY_BLOCK:
		SyntaxKind_OPEN_BRACE_TOKEN
	case FUNC_DEF,
		FUNC_DEF_OR_FUNC_TYPE,
		FUNC_TYPE_DESC,
		FUNC_TYPE_DESC_OR_ANON_FUNC:
		SyntaxKind_FUNCTION_KEYWORD
	case SIMPLE_TYPE_DESCRIPTOR:
		SyntaxKind_ANY_KEYWORD
	case REQUIRED_PARAM,
		VAR_DECL_STMT,
		ASSIGNMENT_OR_VAR_DECL_STMT,
		DEFAULTABLE_PARAM,
		REST_PARAM,
		TYPE_NAME,
		TYPE_REFERENCE_IN_TYPE_INCLUSION,
		TYPE_REFERENCE,
		SIMPLE_TYPE_DESC_IDENTIFIER,
		FIELD_ACCESS_IDENTIFIER,
		FUNC_NAME,
		CLASS_NAME,
		VARIABLE_NAME,
		IMPORT_MODULE_NAME,
		IMPORT_ORG_OR_MODULE_NAME,
		IMPORT_PREFIX,
		VARIABLE_REF,
		BASIC_LITERAL, // return var-ref for any kind of terminal expression
		IDENTIFIER,
		QUALIFIED_IDENTIFIER_START_IDENTIFIER,
		NAMESPACE_PREFIX,
		IMPLICIT_ANON_FUNC_PARAM,
		METHOD_NAME,
		PEER_WORKER_NAME,
		RECEIVE_FIELD_NAME,
		WAIT_FIELD_NAME,
		FIELD_BINDING_PATTERN_NAME,
		XML_ATOMIC_NAME_IDENTIFIER,
		MAPPING_FIELD_NAME,
		WORKER_NAME,
		NAMED_WORKERS,
		ANNOTATION_TAG,
		AFTER_PARAMETER_TYPE,
		MODULE_ENUM_NAME,
		ENUM_MEMBER_NAME,
		TYPED_BINDING_PATTERN_TYPE_RHS,
		ASSIGNMENT_STMT,
		EXPRESSION,
		TERMINAL_EXPRESSION,
		XML_NAME,
		ACCESS_EXPRESSION,
		BINDING_PATTERN_STARTING_IDENTIFIER,
		COMPUTED_FIELD_NAME,
		SIMPLE_BINDING_PATTERN,
		ERROR_FIELD_BINDING_PATTERN,
		ERROR_CAUSE_SIMPLE_BINDING_PATTERN,
		PATH_SEGMENT_IDENT,
		TYPE_DESCRIPTOR,
		NAMED_ARG_BINDING_PATTERN:
		SyntaxKind_IDENTIFIER_TOKEN
	case DECIMAL_INTEGER_LITERAL_TOKEN,
		SIGNED_INT_OR_FLOAT_RHS:
		SyntaxKind_DECIMAL_INTEGER_LITERAL_TOKEN
	case STRING_LITERAL_TOKEN:
		SyntaxKind_STRING_LITERAL_TOKEN
	case OPTIONAL_TYPE_DESCRIPTOR:
		SyntaxKind_OPTIONAL_TYPE_DESC
	case ARRAY_TYPE_DESCRIPTOR:
		SyntaxKind_ARRAY_TYPE_DESC
	case HEX_INTEGER_LITERAL_TOKEN:
		SyntaxKind_HEX_INTEGER_LITERAL_TOKEN
	case OBJECT_FIELD_RHS:
		SyntaxKind_SEMICOLON_TOKEN
	case DECIMAL_FLOATING_POINT_LITERAL_TOKEN:
		SyntaxKind_DECIMAL_FLOATING_POINT_LITERAL_TOKEN
	case HEX_FLOATING_POINT_LITERAL_TOKEN:
		SyntaxKind_HEX_FLOATING_POINT_LITERAL_TOKEN
	case STATEMENT,
		STATEMENT_WITHOUT_ANNOTS:
		SyntaxKind_CLOSE_BRACE_TOKEN
	case ERROR_MATCH_PATTERN,
		NIL_LITERAL:
		SyntaxKind_OPEN_PAREN_TOKEN
	default:
		this.getExpectedSeperatorTokenKind(ctx)
	}
}

func (this *BallerinaParserErrorHandler) getExpectedSeperatorTokenKind(ctx ParserRuleContext) common.SyntaxKind {
	switch ctx {
	case BITWISE_AND_OPERATOR:
		SyntaxKind_BITWISE_AND_TOKEN
	case EQUAL_OR_RIGHT_ARROW,
		ASSIGN_OP:
		SyntaxKind_EQUAL_TOKEN
	case EOF:
		SyntaxKind_EOF_TOKEN
	case BINARY_OPERATOR:
		SyntaxKind_PLUS_TOKEN
	case CLOSE_BRACE:
		SyntaxKind_CLOSE_BRACE_TOKEN
	case CLOSE_PARENTHESIS,
		ARG_LIST_CLOSE_PAREN:
		SyntaxKind_CLOSE_PAREN_TOKEN
	case COMMA,
		ERROR_MESSAGE_BINDING_PATTERN_END_COMMA,
		ERROR_MESSAGE_MATCH_PATTERN_END_COMMA:
		SyntaxKind_COMMA_TOKEN
	case OPEN_BRACE:
		SyntaxKind_OPEN_BRACE_TOKEN
	case OPEN_PARENTHESIS,
		ARG_LIST_OPEN_PAREN,
		PARENTHESISED_TYPE_DESC_START:
		SyntaxKind_OPEN_PAREN_TOKEN
	case SEMICOLON:
		SyntaxKind_SEMICOLON_TOKEN
	case ASTERISK:
		SyntaxKind_ASTERISK_TOKEN
	case CLOSED_RECORD_BODY_END:
		SyntaxKind_CLOSE_BRACE_PIPE_TOKEN
	case CLOSED_RECORD_BODY_START:
		SyntaxKind_OPEN_BRACE_PIPE_TOKEN
	case ELLIPSIS:
		SyntaxKind_ELLIPSIS_TOKEN
	case QUESTION_MARK:
		SyntaxKind_QUESTION_MARK_TOKEN
	case CLOSE_BRACKET:
		SyntaxKind_CLOSE_BRACKET_TOKEN
	case DOT,
		METHOD_CALL_DOT:
		SyntaxKind_DOT_TOKEN
	case OPEN_BRACKET,
		TUPLE_TYPE_DESC_START:
		SyntaxKind_OPEN_BRACKET_TOKEN
	case SLASH,
		ABSOLUTE_PATH_SINGLE_SLASH,
		RESOURCE_METHOD_CALL_SLASH_TOKEN:
		SyntaxKind_SLASH_TOKEN
	case COLON,
		TYPE_REF_COLON,
		VAR_REF_COLON:
		SyntaxKind_COLON_TOKEN
	case UNARY_OPERATOR,
		COMPOUND_BINARY_OPERATOR,
		UNARY_EXPRESSION,
		EXPRESSION_RHS:
		SyntaxKind_PLUS_TOKEN
	case AT:
		SyntaxKind_AT_TOKEN
	case RIGHT_ARROW:
		SyntaxKind_RIGHT_ARROW_TOKEN
	case GT,
		INFERRED_TYPEDESC_DEFAULT_END_GT:
		SyntaxKind_GT_TOKEN
	case LT,
		STREAM_TYPE_PARAM_START_TOKEN,
		INFERRED_TYPEDESC_DEFAULT_START_LT:
		SyntaxKind_LT_TOKEN
	case SYNC_SEND_TOKEN:
		SyntaxKind_SYNC_SEND_TOKEN
	case ANNOT_CHAINING_TOKEN:
		SyntaxKind_ANNOT_CHAINING_TOKEN
	case OPTIONAL_CHAINING_TOKEN:
		SyntaxKind_OPTIONAL_CHAINING_TOKEN
	case DOT_LT_TOKEN:
		SyntaxKind_DOT_LT_TOKEN
	case SLASH_LT_TOKEN:
		SyntaxKind_SLASH_LT_TOKEN
	case DOUBLE_SLASH_DOUBLE_ASTERISK_LT_TOKEN:
		SyntaxKind_DOUBLE_SLASH_DOUBLE_ASTERISK_LT_TOKEN
	case SLASH_ASTERISK_TOKEN:
		SyntaxKind_SLASH_ASTERISK_TOKEN
	case PLUS_TOKEN:
		SyntaxKind_PLUS_TOKEN
	case MINUS_TOKEN:
		SyntaxKind_MINUS_TOKEN
	case LEFT_ARROW_TOKEN:
		SyntaxKind_LEFT_ARROW_TOKEN
	case TEMPLATE_END,
		TEMPLATE_START:
		SyntaxKind_BACKTICK_TOKEN
	case LT_TOKEN:
		SyntaxKind_LT_TOKEN
	case GT_TOKEN:
		SyntaxKind_GT_TOKEN
	case INTERPOLATION_START_TOKEN:
		SyntaxKind_INTERPOLATION_START_TOKEN
	case EXPR_FUNC_BODY_START,
		RIGHT_DOUBLE_ARROW:
		SyntaxKind_RIGHT_DOUBLE_ARROW_TOKEN
	default:
		this.getExpectedKeywordKind(ctx)
	}
}

func (this *BallerinaParserErrorHandler) getExpectedKeywordKind(ctx ParserRuleContext) common.SyntaxKind {
	switch ctx {
	case EXTERNAL_KEYWORD:
		SyntaxKind_EXTERNAL_KEYWORD
	case FUNCTION_KEYWORD,
		IDENT_AFTER_OBJECT_IDENT,
		FUNCTION_IDENT,
		OPTIONAL_PEER_WORKER,
		DEFAULT_WORKER_NAME_IN_ASYNC_SEND:
		SyntaxKind_FUNCTION_KEYWORD
	case RETURNS_KEYWORD:
		SyntaxKind_RETURNS_KEYWORD
	case PUBLIC_KEYWORD:
		SyntaxKind_PUBLIC_KEYWORD
	case RECORD_FIELD,
		RECORD_KEYWORD,
		RECORD_IDENT:
		SyntaxKind_RECORD_KEYWORD
	case TYPE_KEYWORD,
		SINGLE_KEYWORD_ATTACH_POINT_IDENT:
		SyntaxKind_TYPE_KEYWORD
	case OBJECT_KEYWORD,
		OBJECT_IDENT,
		OBJECT_TYPE_DESCRIPTOR:
		SyntaxKind_OBJECT_KEYWORD
	case PRIVATE_KEYWORD:
		SyntaxKind_PRIVATE_KEYWORD
	case REMOTE_IDENT:
		SyntaxKind_REMOTE_KEYWORD
	case ABSTRACT_KEYWORD:
		SyntaxKind_ABSTRACT_KEYWORD
	case CLIENT_KEYWORD:
		SyntaxKind_CLIENT_KEYWORD
	case IF_KEYWORD:
		SyntaxKind_IF_KEYWORD
	case ELSE_KEYWORD:
		SyntaxKind_ELSE_KEYWORD
	case WHILE_KEYWORD:
		SyntaxKind_WHILE_KEYWORD
	case CHECKING_KEYWORD:
		SyntaxKind_CHECK_KEYWORD
	case FAIL_KEYWORD:
		SyntaxKind_FAIL_KEYWORD
	case AS_KEYWORD:
		SyntaxKind_AS_KEYWORD
	case BOOLEAN_LITERAL:
		SyntaxKind_TRUE_KEYWORD
	case IMPORT_KEYWORD:
		SyntaxKind_IMPORT_KEYWORD
	case ON_KEYWORD:
		SyntaxKind_ON_KEYWORD
	case PANIC_KEYWORD:
		SyntaxKind_PANIC_KEYWORD
	case RETURN_KEYWORD:
		SyntaxKind_RETURN_KEYWORD
	case SERVICE_KEYWORD,
		SERVICE_IDENT:
		SyntaxKind_SERVICE_KEYWORD
	case BREAK_KEYWORD:
		SyntaxKind_BREAK_KEYWORD
	case LISTENER_KEYWORD:
		SyntaxKind_LISTENER_KEYWORD
	case CONTINUE_KEYWORD:
		SyntaxKind_CONTINUE_KEYWORD
	case CONST_KEYWORD:
		SyntaxKind_CONST_KEYWORD
	case FINAL_KEYWORD:
		SyntaxKind_FINAL_KEYWORD
	case IS_KEYWORD:
		SyntaxKind_IS_KEYWORD
	case TYPEOF_KEYWORD:
		SyntaxKind_TYPEOF_KEYWORD
	case MAP_KEYWORD,
		MAP_TYPE_DESCRIPTOR:
		SyntaxKind_MAP_KEYWORD
	case PARAMETERIZED_TYPE,
		ERROR_KEYWORD,
		ERROR_BINDING_PATTERN:
		SyntaxKind_ERROR_KEYWORD
	case NULL_KEYWORD:
		SyntaxKind_NULL_KEYWORD
	case LOCK_KEYWORD:
		SyntaxKind_LOCK_KEYWORD
	case ANNOTATION_KEYWORD:
		SyntaxKind_ANNOTATION_KEYWORD
	case FIELD_IDENT:
		SyntaxKind_FIELD_KEYWORD
	case XMLNS_KEYWORD,
		XML_NAMESPACE_DECLARATION:
		SyntaxKind_XMLNS_KEYWORD
	case SOURCE_KEYWORD:
		SyntaxKind_SOURCE_KEYWORD
	case START_KEYWORD:
		SyntaxKind_START_KEYWORD
	case FLUSH_KEYWORD:
		SyntaxKind_FLUSH_KEYWORD
	case WAIT_KEYWORD:
		SyntaxKind_WAIT_KEYWORD
	case TRANSACTION_KEYWORD:
		SyntaxKind_TRANSACTION_KEYWORD
	case TRANSACTIONAL_KEYWORD:
		SyntaxKind_TRANSACTIONAL_KEYWORD
	case COMMIT_KEYWORD:
		SyntaxKind_COMMIT_KEYWORD
	case RETRY_KEYWORD:
		SyntaxKind_RETRY_KEYWORD
	case ROLLBACK_KEYWORD:
		SyntaxKind_ROLLBACK_KEYWORD
	case ENUM_KEYWORD:
		SyntaxKind_ENUM_KEYWORD
	case MATCH_KEYWORD:
		SyntaxKind_MATCH_KEYWORD
	case NEW_KEYWORD:
		SyntaxKind_NEW_KEYWORD
	case FORK_KEYWORD:
		SyntaxKind_FORK_KEYWORD
	case NAMED_WORKER_DECL,
		WORKER_KEYWORD:
		SyntaxKind_WORKER_KEYWORD
	case TRAP_KEYWORD:
		SyntaxKind_TRAP_KEYWORD
	case FOREACH_KEYWORD:
		SyntaxKind_FOREACH_KEYWORD
	case IN_KEYWORD:
		SyntaxKind_IN_KEYWORD
	case PIPE,
		UNION_OR_INTERSECTION_TOKEN:
		SyntaxKind_PIPE_TOKEN
	case TABLE_KEYWORD:
		SyntaxKind_TABLE_KEYWORD
	case KEY_KEYWORD:
		SyntaxKind_KEY_KEYWORD
	case STREAM_KEYWORD:
		SyntaxKind_STREAM_KEYWORD
	case LET_KEYWORD:
		SyntaxKind_LET_KEYWORD
	case XML_KEYWORD:
		SyntaxKind_XML_KEYWORD
	case RE_KEYWORD:
		SyntaxKind_RE_KEYWORD
	case STRING_KEYWORD:
		SyntaxKind_STRING_KEYWORD
	case BASE16_KEYWORD:
		SyntaxKind_BASE16_KEYWORD
	case BASE64_KEYWORD:
		SyntaxKind_BASE64_KEYWORD
	case SELECT_KEYWORD:
		SyntaxKind_SELECT_KEYWORD
	case WHERE_KEYWORD:
		SyntaxKind_WHERE_KEYWORD
	case FROM_KEYWORD:
		SyntaxKind_FROM_KEYWORD
	case ORDER_KEYWORD:
		SyntaxKind_ORDER_KEYWORD
	case GROUP_KEYWORD:
		SyntaxKind_GROUP_KEYWORD
	case BY_KEYWORD:
		SyntaxKind_BY_KEYWORD
	case ORDER_DIRECTION:
		SyntaxKind_ASCENDING_KEYWORD
	case DO_KEYWORD:
		SyntaxKind_DO_KEYWORD
	case DISTINCT_KEYWORD:
		SyntaxKind_DISTINCT_KEYWORD
	case VAR_KEYWORD:
		SyntaxKind_VAR_KEYWORD
	case CONFLICT_KEYWORD:
		SyntaxKind_CONFLICT_KEYWORD
	case LIMIT_KEYWORD:
		SyntaxKind_LIMIT_KEYWORD
	case EQUALS_KEYWORD:
		SyntaxKind_EQUALS_KEYWORD
	case JOIN_KEYWORD:
		SyntaxKind_JOIN_KEYWORD
	case OUTER_KEYWORD:
		SyntaxKind_OUTER_KEYWORD
	case CLASS_KEYWORD:
		SyntaxKind_CLASS_KEYWORD
	case COLLECT_KEYWORD:
		SyntaxKind_COLLECT_KEYWORD
	case NATURAL_KEYWORD:
		SyntaxKind_NATURAL_KEYWORD
	default:
		this.getExpectedQualifierKind(ctx)
	}
}

func (this *BallerinaParserErrorHandler) getExpectedQualifierKind(ctx ParserRuleContext) common.SyntaxKind {
	switch ctx {
	case FIRST_OBJECT_CONS_QUALIFIER,
		SECOND_OBJECT_CONS_QUALIFIER,
		FIRST_OBJECT_TYPE_QUALIFIER,
		SECOND_OBJECT_TYPE_QUALIFIER:
		SyntaxKind_OBJECT_KEYWORD
	case FIRST_CLASS_TYPE_QUALIFIER,
		SECOND_CLASS_TYPE_QUALIFIER,
		THIRD_CLASS_TYPE_QUALIFIER,
		FOURTH_CLASS_TYPE_QUALIFIER:
		SyntaxKind_CLASS_KEYWORD
	case FUNC_DEF_FIRST_QUALIFIER,
		FUNC_DEF_SECOND_QUALIFIER,
		FUNC_TYPE_FIRST_QUALIFIER,
		FUNC_TYPE_SECOND_QUALIFIER,
		OBJECT_METHOD_FIRST_QUALIFIER,
		OBJECT_METHOD_SECOND_QUALIFIER,
		OBJECT_METHOD_THIRD_QUALIFIER,
		OBJECT_METHOD_FOURTH_QUALIFIER:
		SyntaxKind_FUNCTION_KEYWORD
	case MODULE_VAR_FIRST_QUAL,
		MODULE_VAR_SECOND_QUAL,
		MODULE_VAR_THIRD_QUAL,
		OBJECT_MEMBER_VISIBILITY_QUAL:
		SyntaxKind_IDENTIFIER_TOKEN
	case SERVICE_DECL_QUALIFIER:
		SyntaxKind_SERVICE_KEYWORD
	default:
		SyntaxKind_NONE
	}
}

func (this *BallerinaParserErrorHandler) isBasicLiteral(kind common.SyntaxKind) bool {
	switch kind {
	case DECIMAL_INTEGER_LITERAL_TOKEN,
		HEX_INTEGER_LITERAL_TOKEN,
		STRING_LITERAL_TOKEN,
		TRUE_KEYWORD,
		FALSE_KEYWORD,
		NULL_KEYWORD,
		DECIMAL_FLOATING_POINT_LITERAL_TOKEN,
		HEX_FLOATING_POINT_LITERAL_TOKEN:
		true
	default:
		false
	}
}

func (this *BallerinaParserErrorHandler) isUnaryOperator(token Token) bool {
	switch token.kind {
	case PLUS_TOKEN,
		MINUS_TOKEN,
		NEGATION_TOKEN,
		EXCLAMATION_MARK_TOKEN:
		true
	default:
		false
	}
}

func (this *BallerinaParserErrorHandler) isSingleKeywordAttachPointIdent(tokenKind common.SyntaxKind) bool {
	switch tokenKind {
	case ANNOTATION_KEYWORD,
		EXTERNAL_KEYWORD,
		VAR_KEYWORD,
		CONST_KEYWORD,
		LISTENER_KEYWORD,
		WORKER_KEYWORD,
		TYPE_KEYWORD,
		FUNCTION_KEYWORD,
		PARAMETER_KEYWORD,
		RETURN_KEYWORD,
		FIELD_KEYWORD,
		CLASS_KEYWORD:
		true
	default:
		false
	}
}
