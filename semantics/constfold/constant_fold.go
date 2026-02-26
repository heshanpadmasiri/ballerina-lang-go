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
	"sync"

	"ballerina-lang-go/ast"
	"ballerina-lang-go/context"
	"ballerina-lang-go/model"
)

type foldContext struct {
	compilerCtx     *context.CompilerContext
	constantsByRef  map[model.SymbolRef]*ast.BLangConstant
	foldedConstants map[model.SymbolRef]model.LiteralNode
	folding         map[model.SymbolRef]bool
}

func FoldConstants(cx *context.CompilerContext, pkg *ast.BLangPackage) {
	foldCtx := &foldContext{
		compilerCtx:     cx,
		constantsByRef:  buildConstantsByRef(pkg),
		foldedConstants: make(map[model.SymbolRef]model.LiteralNode),
		folding:         make(map[model.SymbolRef]bool),
	}

	foldPackageConstants(foldCtx, pkg)
	foldFunctionBodies(foldCtx, pkg)
}

func buildConstantsByRef(pkg *ast.BLangPackage) map[model.SymbolRef]*ast.BLangConstant {
	m := make(map[model.SymbolRef]*ast.BLangConstant, len(pkg.Constants))
	for i := range pkg.Constants {
		c := &pkg.Constants[i]
		m[c.Symbol()] = c
	}
	return m
}

func foldPackageConstants(cx *foldContext, pkg *ast.BLangPackage) {
	for i := range pkg.Constants {
		c := &pkg.Constants[i]
		if c.Expr == nil {
			continue
		}
		folded := foldExpression(cx, c.Expr.(ast.BLangExpression))
		c.Expr = folded
		if lit, ok := asLiteral(folded); ok {
			cx.foldedConstants[c.Symbol()] = lit
		}
	}
}

func foldFunctionBodies(cx *foldContext, pkg *ast.BLangPackage) {
	var wg sync.WaitGroup
	var panicErr any

	processFn := func(fn *ast.BLangFunction) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					panicErr = r
				}
			}()
			foldFunctionBody(cx, fn)
		}()
	}

	for i := range pkg.Functions {
		processFn(&pkg.Functions[i])
	}

	for i := range pkg.ClassDefinitions {
		class := &pkg.ClassDefinitions[i]
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					panicErr = r
				}
			}()
			for j := range class.Functions {
				foldFunctionBody(cx, &class.Functions[j])
			}
			if class.InitFunction != nil {
				foldFunctionBody(cx, class.InitFunction)
			}
		}()
	}

	if pkg.InitFunction != nil {
		processFn(pkg.InitFunction)
	}
	if pkg.StartFunction != nil {
		processFn(pkg.StartFunction)
	}
	if pkg.StopFunction != nil {
		processFn(pkg.StopFunction)
	}

	wg.Wait()
	if panicErr != nil {
		panic(panicErr)
	}
}
