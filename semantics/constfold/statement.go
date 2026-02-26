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
	"ballerina-lang-go/ast"
)

func foldFunctionBody(cx *foldContext, fn *ast.BLangFunction) {
	if fn.Body == nil {
		return
	}
	switch body := fn.Body.(type) {
	case *ast.BLangBlockFunctionBody:
		foldBlockFunctionBody(cx, body)
	case *ast.BLangExprFunctionBody:
		if body.Expr != nil {
			body.Expr = foldExpression(cx, body.Expr.(ast.BLangExpression))
		}
	}
}

func foldBlockFunctionBody(cx *foldContext, body *ast.BLangBlockFunctionBody) {
	for i, stmt := range body.Stmts {
		body.Stmts[i] = foldStatement(cx, stmt)
	}
}

func foldStatement(cx *foldContext, node ast.BLangStatement) ast.BLangStatement {
	switch stmt := node.(type) {
	case *ast.BLangBlockStmt:
		foldBlockStmt(cx, stmt)
		return stmt
	case *ast.BLangAssignment:
		foldAssignment(cx, stmt)
		return stmt
	case *ast.BLangCompoundAssignment:
		foldCompoundAssignment(cx, stmt)
		return stmt
	case *ast.BLangExpressionStmt:
		foldExpressionStmt(cx, stmt)
		return stmt
	case *ast.BLangIf:
		foldIf(cx, stmt)
		return stmt
	case *ast.BLangWhile:
		foldWhile(cx, stmt)
		return stmt
	case *ast.BLangDo:
		foldDo(cx, stmt)
		return stmt
	case *ast.BLangForeach:
		foldForeach(cx, stmt)
		return stmt
	case *ast.BLangSimpleVariableDef:
		foldSimpleVariableDef(cx, stmt)
		return stmt
	case *ast.BLangReturn:
		foldReturn(cx, stmt)
		return stmt
	default:
		return stmt
	}
}

func foldBlockStmt(cx *foldContext, stmt *ast.BLangBlockStmt) {
	for i, child := range stmt.Stmts {
		stmt.Stmts[i] = foldStatement(cx, child)
	}
}

func foldAssignment(cx *foldContext, stmt *ast.BLangAssignment) {
	if stmt.Expr != nil {
		stmt.Expr = foldExpression(cx, stmt.Expr)
	}
}

func foldCompoundAssignment(cx *foldContext, stmt *ast.BLangCompoundAssignment) {
	if stmt.Expr != nil {
		stmt.Expr = foldExpression(cx, stmt.Expr)
	}
}

func foldExpressionStmt(cx *foldContext, stmt *ast.BLangExpressionStmt) {
	if stmt.Expr != nil {
		stmt.Expr = foldExpression(cx, stmt.Expr)
	}
}

func foldIf(cx *foldContext, stmt *ast.BLangIf) {
	if stmt.Expr != nil {
		stmt.Expr = foldExpression(cx, stmt.Expr)
	}
	foldBlockStmt(cx, &stmt.Body)
	if stmt.ElseStmt != nil {
		stmt.ElseStmt = foldStatement(cx, stmt.ElseStmt)
	}
}

func foldWhile(cx *foldContext, stmt *ast.BLangWhile) {
	if stmt.Expr != nil {
		stmt.Expr = foldExpression(cx, stmt.Expr)
	}
	foldBlockStmt(cx, &stmt.Body)
}

func foldDo(cx *foldContext, stmt *ast.BLangDo) {
	foldBlockStmt(cx, &stmt.Body)
}

func foldForeach(cx *foldContext, stmt *ast.BLangForeach) {
	if stmt.Collection != nil {
		stmt.Collection = foldExpression(cx, stmt.Collection)
	}
	foldBlockStmt(cx, &stmt.Body)
}

func foldSimpleVariableDef(cx *foldContext, stmt *ast.BLangSimpleVariableDef) {
	if stmt.Var != nil && stmt.Var.Expr != nil {
		stmt.Var.Expr = foldExpression(cx, stmt.Var.Expr.(ast.BLangExpression))
	}
}

func foldReturn(cx *foldContext, stmt *ast.BLangReturn) {
	if stmt.Expr != nil {
		stmt.Expr = foldExpression(cx, stmt.Expr)
	}
}
