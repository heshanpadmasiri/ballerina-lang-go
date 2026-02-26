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

package constfold

import (
	"fmt"
	"strconv"

	"ballerina-lang-go/ast"
	"ballerina-lang-go/model"
)

func foldExpression(cx *foldContext, expr ast.BLangExpression) ast.BLangExpression {
	switch e := expr.(type) {
	case *ast.BLangLiteral:
		return e
	case *ast.BLangNumericLiteral:
		return e
	case *ast.BLangGroupExpr:
		result := foldExpression(cx, e.Expression)
		if isLiteral(result) {
			return result
		}
		e.Expression = result
		return e
	case *ast.BLangBinaryExpr:
		return foldBinaryExpr(cx, e)
	case *ast.BLangUnaryExpr:
		return foldUnaryExpr(cx, e)
	case *ast.BLangSimpleVarRef:
		return foldVarRef(cx, e)
	case *ast.BLangConstRef:
		return foldConstRef(cx, e)
	case *ast.BLangInvocation:
		foldInvocation(cx, e)
		return e
	case *ast.BLangTypeConversionExpr:
		if e.Expression != nil {
			e.Expression = foldExpression(cx, e.Expression)
		}
		return e
	case *ast.BLangIndexBasedAccess:
		if e.Expr != nil {
			e.Expr = foldExpression(cx, e.Expr)
		}
		if e.IndexExpr != nil {
			e.IndexExpr = foldExpression(cx, e.IndexExpr)
		}
		return e
	case *ast.BLangListConstructorExpr:
		for i, child := range e.Exprs {
			e.Exprs[i] = foldExpression(cx, child)
		}
		return e
	default:
		return expr
	}
}

func foldBinaryExpr(cx *foldContext, expr *ast.BLangBinaryExpr) ast.BLangExpression {
	if expr.LhsExpr != nil {
		expr.LhsExpr = foldExpression(cx, expr.LhsExpr)
	}
	if expr.RhsExpr != nil {
		expr.RhsExpr = foldExpression(cx, expr.RhsExpr)
	}
	lhsVal, lhsOk := extractLiteralValue(expr.LhsExpr)
	rhsVal, rhsOk := extractLiteralValue(expr.RhsExpr)
	if !lhsOk || !rhsOk {
		return expr
	}
	result, ok := evaluateBinary(lhsVal, rhsVal, expr.OpKind)
	if !ok {
		return expr
	}
	resultTypeData := resultTypeDataForBinary(expr.LhsExpr, expr.RhsExpr, result)
	return createLiteral(result, resultTypeData)
}

func foldUnaryExpr(cx *foldContext, expr *ast.BLangUnaryExpr) ast.BLangExpression {
	if expr.Expr != nil {
		expr.Expr = foldExpression(cx, expr.Expr)
	}
	val, ok := extractLiteralValue(expr.Expr)
	if !ok {
		return expr
	}
	result, ok := evaluateUnary(val, expr.Operator)
	if !ok {
		return expr
	}
	return createLiteral(result, expr.Expr.GetTypeData())
}

func foldInvocation(cx *foldContext, expr *ast.BLangInvocation) {
	for i, arg := range expr.ArgExprs {
		expr.ArgExprs[i] = foldExpression(cx, arg)
	}
	for i, arg := range expr.RequiredArgs {
		expr.RequiredArgs[i] = foldExpression(cx, arg)
	}
	for i, arg := range expr.RestArgs {
		expr.RestArgs[i] = foldExpression(cx, arg)
	}
}

func foldVarRef(cx *foldContext, varRef *ast.BLangSimpleVarRef) ast.BLangExpression {
	ref := varRef.Symbol()
	if cx.compilerCtx.SymbolKind(ref) != model.SymbolKindConstant {
		return varRef
	}
	return resolveConstant(cx, ref, varRef)
}

func foldConstRef(cx *foldContext, constRef *ast.BLangConstRef) ast.BLangExpression {
	ref := constRef.Symbol()
	return resolveConstant(cx, ref, constRef)
}

func resolveConstant(cx *foldContext, ref model.SymbolRef, fallback ast.BLangExpression) ast.BLangExpression {
	if lit, ok := cx.foldedConstants[ref]; ok {
		return copyLiteralNode(lit)
	}
	constant, ok := cx.constantsByRef[ref]
	if !ok {
		return fallback
	}
	if cx.folding[ref] {
		return fallback
	}
	cx.folding[ref] = true
	if constant.Expr == nil {
		delete(cx.folding, ref)
		return fallback
	}
	folded := foldExpression(cx, constant.Expr.(ast.BLangExpression))
	delete(cx.folding, ref)
	if lit, ok := asLiteral(folded); ok {
		cx.foldedConstants[ref] = lit
		constant.Expr = folded
		return copyLiteralNode(lit)
	}
	constant.Expr = folded
	return fallback
}

func isLiteral(expr ast.BLangExpression) bool {
	switch expr.(type) {
	case *ast.BLangLiteral, *ast.BLangNumericLiteral:
		return true
	}
	return false
}

func asLiteral(expr ast.BLangExpression) (model.LiteralNode, bool) {
	if lit, ok := expr.(model.LiteralNode); ok {
		return lit, true
	}
	return nil, false
}

func copyLiteralNode(lit model.LiteralNode) ast.BLangExpression {
	switch l := lit.(type) {
	case *ast.BLangNumericLiteral:
		cp := *l
		return &cp
	case *ast.BLangLiteral:
		cp := *l
		return &cp
	}
	return lit.(ast.BLangExpression)
}

func extractLiteralValue(expr ast.BLangExpression) (literalValue, bool) {
	lit, ok := asLiteral(expr)
	if !ok {
		return literalValue{}, false
	}
	return valueFromLiteral(lit)
}

func valueFromLiteral(lit model.LiteralNode) (literalValue, bool) {
	tag := typeTagFromLiteral(lit)
	switch tag {
	case model.TypeTags_INT, model.TypeTags_BYTE:
		if v, ok := lit.GetValue().(int64); ok {
			return newLiteralValue(v), true
		}
	case model.TypeTags_FLOAT:
		switch v := lit.GetValue().(type) {
		case string:
			f, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return literalValue{}, false
			}
			return newLiteralValue(f), true
		case float64:
			return newLiteralValue(v), true
		}
	case model.TypeTags_STRING:
		if v, ok := lit.GetValue().(string); ok {
			return newLiteralValue(v), true
		}
	case model.TypeTags_BOOLEAN:
		if v, ok := lit.GetValue().(bool); ok {
			return newLiteralValue(v), true
		}
	}
	return literalValue{}, false
}

func typeTagFromLiteral(lit model.LiteralNode) model.TypeTags {
	td := lit.(ast.BLangExpression).GetTypeData().TypeDescriptor
	if td == nil {
		return 0
	}
	if bt, ok := td.(ast.BType); ok {
		return bt.BTypeGetTag()
	}
	return 0
}

func resultTypeDataForBinary(lhs, rhs ast.BLangExpression, result literalValue) model.TypeData {
	if _, ok := result.val.(float64); ok {
		if typeTagFromExpr(lhs) == model.TypeTags_FLOAT {
			return lhs.GetTypeData()
		}
		return rhs.GetTypeData()
	}
	return lhs.GetTypeData()
}

func typeTagFromExpr(expr ast.BLangExpression) model.TypeTags {
	td := expr.GetTypeData().TypeDescriptor
	if td == nil {
		return 0
	}
	if bt, ok := td.(ast.BType); ok {
		return bt.BTypeGetTag()
	}
	return 0
}

func createLiteral(val literalValue, typeData model.TypeData) ast.BLangExpression {
	switch v := val.val.(type) {
	case int64:
		numLit := &ast.BLangNumericLiteral{
			BLangLiteral: ast.BLangLiteral{
				Value:         v,
				OriginalValue: fmt.Sprintf("%d", v),
			},
			Kind: model.NodeKind_INTEGER_LITERAL,
		}
		numLit.SetTypeData(typeData)
		return numLit
	case float64:
		s := formatFloat(v)
		numLit := &ast.BLangNumericLiteral{
			BLangLiteral: ast.BLangLiteral{
				Value:         s,
				OriginalValue: s,
			},
			Kind: model.NodeKind_DECIMAL_FLOATING_POINT_LITERAL,
		}
		numLit.SetTypeData(typeData)
		return numLit
	case bool:
		lit := &ast.BLangLiteral{
			Value:         v,
			OriginalValue: strconv.FormatBool(v),
		}
		lit.SetTypeData(typeData)
		return lit
	case string:
		lit := &ast.BLangLiteral{
			Value:         v,
			OriginalValue: v,
		}
		lit.SetTypeData(typeData)
		return lit
	}
	return nil
}
