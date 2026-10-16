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

package cfg

import (
	"slices"
	"sync"

	"github.com/ballerina-nutcracker/ballerina/ast"
	"github.com/ballerina-nutcracker/ballerina/context"
	"github.com/ballerina-nutcracker/ballerina/model"
	"github.com/ballerina-nutcracker/ballerina/semantics/internal/common"
)

// varInitState tracks which variables are definitely initialized, using dense
// per-function variable indices. A variable is tracked if its bit is set in
// tracked, and definitely initialized if its bit is also set in initialized.
type varInitState struct {
	tracked     bitset
	initialized bitset
}

func newVarInitState() *varInitState {
	return &varInitState{}
}

func (s *varInitState) clone() *varInitState {
	return &varInitState{
		tracked:     slices.Clone(s.tracked),
		initialized: slices.Clone(s.initialized),
	}
}

func (s *varInitState) isEmpty() bool {
	return s.tracked.isEmpty()
}

// mergeFrom merges other into s in place. A variable is initialized only if it is
// initialized in both states; a variable tracked in only one state is treated as
// uninitialized. If either state is empty (unvisited block), the other state wins.
func (s *varInitState) mergeFrom(other *varInitState) {
	if other.isEmpty() {
		return
	}
	if s.isEmpty() {
		s.tracked = append(s.tracked[:0], other.tracked...)
		s.initialized = append(s.initialized[:0], other.initialized...)
		return
	}
	s.tracked.unionWith(other.tracked)
	s.initialized.intersectWith(other.initialized)
}

func (s *varInitState) markInitialized(idx int) {
	s.tracked.set(idx)
	s.initialized.set(idx)
}

func (s *varInitState) markUninitialized(idx int) {
	s.tracked.set(idx)
	s.initialized.clear(idx)
}

func (s *varInitState) isTracked(idx int) bool {
	return s.tracked.has(idx)
}

func (s *varInitState) isInitialized(idx int) bool {
	return s.initialized.has(idx)
}

type bitset []uint64

func (b bitset) has(i int) bool {
	w := i >> 6
	return w < len(b) && b[w]&(1<<(uint(i)&63)) != 0
}

func (b *bitset) set(i int) {
	w := i >> 6
	if w >= len(*b) {
		*b = append(*b, make([]uint64, w+1-len(*b))...)
	}
	(*b)[w] |= 1 << (uint(i) & 63)
}

func (b bitset) clear(i int) {
	w := i >> 6
	if w < len(b) {
		b[w] &^= 1 << (uint(i) & 63)
	}
}

func (b bitset) isEmpty() bool {
	for _, w := range b {
		if w != 0 {
			return false
		}
	}
	return true
}

func (b *bitset) unionWith(other bitset) {
	if len(other) > len(*b) {
		*b = append(*b, make([]uint64, len(other)-len(*b))...)
	}
	for i, w := range other {
		(*b)[i] |= w
	}
}

func (b bitset) intersectWith(other bitset) {
	n := min(len(b), len(other))
	for i := range n {
		b[i] &= other[i]
	}
	clear(b[n:])
}

// uninitVarAnalyzer performs data flow analysis for uninitialized variables
type uninitVarAnalyzer struct {
	ctx               *context.CompilerContext
	fn                ast.BLangNode
	fcfg              *functionCFG
	exitStates        []*varInitState
	varIndices        map[model.SymbolRef]int
	implicitInitState *varInitState // vars initialized by language constructs, used as entry state baseline
}

// newUninitVarAnalyzer creates a new analyzer for a function
func newUninitVarAnalyzer(ctx *context.CompilerContext, fn ast.BLangNode, fcfg *functionCFG) *uninitVarAnalyzer {
	analyzer := &uninitVarAnalyzer{
		ctx:        ctx,
		fn:         fn,
		fcfg:       fcfg,
		exitStates: make([]*varInitState, len(fcfg.bbs)),
		varIndices: make(map[model.SymbolRef]int),
	}
	analyzer.implicitInitState = analyzer.buildImplicitInitState(fn)
	for i := range analyzer.exitStates {
		analyzer.exitStates[i] = newVarInitState()
	}
	return analyzer
}

// varIndex returns the dense index of symRef, assigning a new one if needed.
func (a *uninitVarAnalyzer) varIndex(symRef model.SymbolRef) int {
	if idx, ok := a.varIndices[symRef]; ok {
		return idx
	}
	idx := len(a.varIndices)
	a.varIndices[symRef] = idx
	return idx
}

func (a *uninitVarAnalyzer) isTracked(state *varInitState, symRef model.SymbolRef) bool {
	idx, ok := a.varIndices[symRef]
	return ok && state.isTracked(idx)
}

func (a *uninitVarAnalyzer) isInitialized(state *varInitState, symRef model.SymbolRef) bool {
	idx, ok := a.varIndices[symRef]
	return ok && state.isInitialized(idx)
}

func (a *uninitVarAnalyzer) buildImplicitInitState(fn ast.BLangNode) *varInitState {
	marker := &implicitVarMarker{analyzer: a, state: newVarInitState()}
	ast.Walk(marker, fn)
	return marker.state
}

type implicitVarMarker struct {
	analyzer *uninitVarAnalyzer
	state    *varInitState
}

func (m *implicitVarMarker) Visit(node ast.BLangNode) ast.Visitor {
	if node == nil {
		return nil
	}
	if foreach, ok := node.(*ast.BLangForeach); ok && foreach.VariableDef != nil {
		m.state.markInitialized(m.analyzer.varIndex(foreach.VariableDef.Var.Symbol()))
	}
	return m
}

func (m *implicitVarMarker) VisitTypeData(*ast.TypeData) ast.Visitor { return m }

func (a *uninitVarAnalyzer) analyze() {
	if len(a.fcfg.bbs) == 0 {
		return
	}
	for _, i := range a.fcfg.topoOrder {
		bb := &a.fcfg.bbs[i]
		a.exitStates[i] = a.analyzeBlock(bb, a.mergePredecessors(bb))
	}
}

// mergePredecessors merges the exit states of all non-backedge predecessors.
func (a *uninitVarAnalyzer) mergePredecessors(bb *basicBlock) *varInitState {
	result := (*varInitState)(nil)
	for _, parentID := range bb.parents {
		if slices.Contains(bb.backedgeParents, parentID) {
			continue
		}
		if result == nil {
			result = a.exitStates[parentID].clone()
		} else {
			result.mergeFrom(a.exitStates[parentID])
		}
	}
	if result == nil {
		result = newVarInitState()
	}
	result.tracked.unionWith(a.implicitInitState.tracked)
	result.initialized.unionWith(a.implicitInitState.initialized)
	return result
}

// analyzeBlock performs intra-block analysis
func (a *uninitVarAnalyzer) analyzeBlock(bb *basicBlock, state *varInitState) *varInitState {
	for _, node := range bb.nodes {
		a.analyzeNode(node, state)
	}
	return state
}

// analyzeNode processes a single node in the CFG
func (a *uninitVarAnalyzer) analyzeNode(node ast.Node, state *varInitState) {
	switch n := node.(type) {
	case *ast.BLangVariableDef:
		idx := a.varIndex(n.Var.Symbol())
		if n.Var.Expr != nil {
			a.checkExpression(n.Var.Expr, state)
			state.markInitialized(idx)
		} else if !state.isInitialized(idx) {
			state.markUninitialized(idx)
		}
	case *ast.BLangAssignment:
		a.checkExpression(n.Expr, state)
		a.markAssignmentTarget(n.VarRef, state)
	case *ast.BLangCompoundAssignment:
		a.checkExpression(n.VarRef, state)
		a.checkExpression(n.Expr, state)
		a.markAssignmentTarget(n.VarRef, state)
	case ast.BLangExpression:
		// Expression nodes (like conditions in while loops) need to be checked
		a.checkExpression(n, state)
	case ast.BLangNode:
		checker := &varRefChecker{analyzer: a, state: state}
		ast.Walk(checker, n)
	default:
		a.ctx.InternalError("unexpected node", node.GetPosition())
	}
}

// markAssignmentTarget marks the target of an assignment as initialized
func (a *uninitVarAnalyzer) markAssignmentTarget(expr ast.BLangExpression, state *varInitState) {
	// Check for wildcard pattern - no tracking needed
	if _, ok := expr.(*ast.BLangWildCardBindingPattern); ok {
		return
	}

	// For simple variable references, mark as initialized
	if nodeWithSymbol, ok := expr.(ast.NodeWithSymbol); ok {
		if idx, ok := a.varIndices[nodeWithSymbol.Symbol()]; ok && state.isTracked(idx) {
			state.markInitialized(idx)
		}
		return
	}

	// For complex expressions (index access, field access), check the expression
	// but don't mark anything as initialized - these modify existing data
	a.checkExpression(expr, state)
}

// checkExpression walks an expression and checks all variable references
func (a *uninitVarAnalyzer) checkExpression(expr ast.BLangActionOrExpression, state *varInitState) {
	checker := &varRefChecker{analyzer: a, state: state}
	ast.Walk(checker, expr)
}

// checkVariableReference checks if a variable is initialized before use
func (a *uninitVarAnalyzer) checkVariableReference(symRef model.SymbolRef, node ast.Node, state *varInitState) {
	if !a.isTracked(state, symRef) {
		return
	}

	if !a.isInitialized(state, symRef) {
		// Variable is tracked but not initialized - report error
		a.ctx.SemanticError("variable may not be initialized", node.GetPosition())
	}
}

// varRefChecker is a visitor that checks variable references in expressions
type varRefChecker struct {
	analyzer *uninitVarAnalyzer
	state    *varInitState
}

var _ ast.Visitor = &varRefChecker{}

func (v *varRefChecker) Visit(node ast.BLangNode) ast.Visitor {
	if node == nil {
		return nil
	}

	// Check if this node is a variable reference
	if inv, ok := node.(*ast.BLangInvocation); ok && ast.IsStreamOperation(inv) {
		return nil
	}
	if nodeWithSymbol, ok := node.(ast.NodeWithSymbol); ok {
		v.analyzer.checkVariableReference(nodeWithSymbol.Symbol(), node, v.state)
	}

	return v
}

func (v *varRefChecker) VisitTypeData(typeData *ast.TypeData) ast.Visitor {
	// TypeDesc could have default values
	return v
}

// analyzeUninitializedVars is the public entry point for uninitialized variable analysis
func analyzeUninitializedVars(ctx *context.CompilerContext, pkg *ast.BLangPackage, cfg *PackageCFG) {
	var wg sync.WaitGroup
	for _, fn := range common.PackageFunctionDecls(pkg) {
		wg.Add(1)
		go func(fn common.FunctionDecl) {
			defer wg.Done()
			analyzeFunctionUninitializedVars(ctx, fn, cfg)
		}(fn)
	}
	for i := range cfg.lambdaCfgs {
		lambda := &cfg.lambdaCfgs[i]
		wg.Go(func() { newUninitVarAnalyzer(ctx, lambda.fn, &lambda.cfg).analyze() })
	}
	wg.Wait()
}

// analyzeFunctionUninitializedVars analyzes a single function for uninitialized variables
func analyzeFunctionUninitializedVars(ctx *context.CompilerContext, fn common.FunctionDecl, cfg *PackageCFG) {
	fnCfg, ok := cfg.lookupFunctionCfg(fn.Symbol())
	if !ok {
		return
	}

	analyzer := newUninitVarAnalyzer(ctx, fn, &fnCfg)
	analyzer.analyze()
}
