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

package semantics

import (
	"ballerina-lang-go/ast"
	debugcommon "ballerina-lang-go/common"
	"ballerina-lang-go/context"
	"ballerina-lang-go/parser"
	"ballerina-lang-go/semtypes"
	"ballerina-lang-go/test_util"
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"
)

func TestConstFold(t *testing.T) {

	testPairs := test_util.GetValidTests(t, test_util.ConstFold)

	for _, testPair := range testPairs {
		t.Run(testPair.Name, func(t *testing.T) {
			t.Parallel()
			testConstFold(t, testPair)
		})
	}
}

func testConstFold(t *testing.T, testCase test_util.TestCase) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("ConstFold panicked for %s: %v", testCase.InputPath, r)
		}
	}()

	debugCtx := debugcommon.DebugContext{
		Channel: make(chan string),
	}
	cx := context.NewCompilerContext(semtypes.CreateTypeEnv())

	syntaxTree, err := parser.GetSyntaxTree(cx, &debugCtx, testCase.InputPath)
	if err != nil {
		t.Errorf("error getting syntax tree for %s: %v", testCase.InputPath, err)
		return
	}
	compilationUnit := ast.GetCompilationUnit(cx, syntaxTree)
	if compilationUnit == nil {
		t.Errorf("compilation unit is nil for %s", testCase.InputPath)
		return
	}
	pkg := ast.ToPackage(compilationUnit)

	importedSymbols := ResolveImports(cx, pkg, GetImplicitImports(cx))
	ResolveSymbols(cx, pkg, importedSymbols)

	FoldConstants(cx, pkg)

	prettyPrinter := ast.PrettyPrinter{}
	actualAST := prettyPrinter.Print(pkg)

	if *updateCFG {
		if test_util.UpdateIfNeeded(t, testCase.ExpectedPath, actualAST) {
			t.Errorf("updated expected constant-folded AST file: %s", testCase.ExpectedPath)
		}
		return
	}

	expectedAST := test_util.ReadExpectedFile(t, testCase.ExpectedPath)

	if actualAST != expectedAST {
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(expectedAST, actualAST, false)
		t.Errorf("Constant-folded AST mismatch for %s\nExpected file: %s\n%s",
			testCase.InputPath, testCase.ExpectedPath, dmp.DiffPrettyText(diffs))
		return
	}
}
