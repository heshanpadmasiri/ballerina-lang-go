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
//

package internal

import (
	"ballerina-lang-go/parser/common"
	"testing"
)

// Test helpers to create nodes without directly exposing STNodeBase construction

func createTestModulePart(imports, members, eofToken STNode) *STModulePart {
	childBuckets := []STNode{}
	if imports != nil {
		childBuckets = append(childBuckets, imports)
	}
	if members != nil {
		childBuckets = append(childBuckets, members)
	}
	if eofToken != nil {
		childBuckets = append(childBuckets, eofToken)
	}

	return &STModulePart{
		STNode: &STNodeBase{
			kind: common.MODULE_PART,
		},
		Imports:  imports,
		Members:  members,
		EofToken: eofToken,
	}
}

func createTestImportDeclaration(importKeyword, orgName, moduleName, prefix, semicolon STNode) *STImportDeclarationNode {
	childBuckets := []STNode{}
	if importKeyword != nil {
		childBuckets = append(childBuckets, importKeyword)
	}
	if orgName != nil {
		childBuckets = append(childBuckets, orgName)
	}
	if moduleName != nil {
		childBuckets = append(childBuckets, moduleName)
	}
	if prefix != nil {
		childBuckets = append(childBuckets, prefix)
	}
	if semicolon != nil {
		childBuckets = append(childBuckets, semicolon)
	}

	return &STImportDeclarationNode{
		STNode: &STNodeBase{
			kind: common.IMPORT_DECLARATION,
		},
		ImportKeyword: importKeyword,
		OrgName:       orgName,
		ModuleName:    moduleName,
		Prefix:        prefix,
		Semicolon:     semicolon,
	}
}

func createTestFunctionDefinition() *STFunctionDefinition {
	return &STFunctionDefinition{
		STModuleMemberDeclarationNode: &STNodeBase{
			kind: common.FUNCTION_DEFINITION,
		},
		Metadata:             nil,
		QualifierList:        nil,
		FunctionKeyword:      nil,
		FunctionName:         nil,
		RelativeResourcePath: nil,
		FunctionSignature:    nil,
		FunctionBody:         nil,
	}
}

func TestReplace_TopNode(t *testing.T) {
	child1 := CreateInvalidToken("child1")
	child2 := CreateInvalidToken("child2")

	root := createTestModulePart(child1, child2, nil)

	replacement := CreateInvalidToken("replacement")

	result := Replace(root, root, replacement)

	if result != replacement {
		t.Errorf("Top node was not replaced correctly")
	}
}

func TestReplace_MiddleNode(t *testing.T) {
	leaf1 := CreateTokenFrom(common.FUNCTION_KEYWORD, nil, nil)
	leaf2 := CreateTokenFrom(common.OPEN_BRACE_TOKEN, nil, nil)

	middle1 := createTestImportDeclaration(leaf1, leaf2, nil, nil, nil)
	middle2 := createTestFunctionDefinition()

	root := createTestModulePart(middle1, middle2, nil)

	replacement := CreateInvalidToken("replacement")

	result := Replace(root, middle1, replacement)

	newModule, ok := result.(*STModulePart)
	if !ok {
		t.Errorf("Result is not a STModulePart")
	}

	if newModule.Imports != replacement {
		t.Errorf("Middle node was not replaced correctly")
	}

}

func TestReplace_LeafNode(t *testing.T) {
	leaf1 := CreateTokenFrom(common.FUNCTION_KEYWORD, nil, nil)
	leaf2 := CreateTokenFrom(common.OPEN_BRACE_TOKEN, nil, nil)

	root := createTestImportDeclaration(leaf1, leaf2, nil, nil, nil)

	replacement := CreateTokenFrom(common.IMPORT_KEYWORD, nil, nil)

	result := Replace(root, leaf1, replacement)

	newImport, ok := result.(*STImportDeclarationNode)
	if !ok {
		t.Errorf("Result is not a STImportDeclarationNode")
	}

	// leaf1 (ImportKeyword) should now be replacement
	if newImport.ImportKeyword != replacement {
		t.Errorf("Leaf node ImportKeyword was not replaced correctly")
	}

	// leaf2 (OrgName) should still be unchanged
	if newImport.OrgName != leaf2 {
		t.Errorf("Leaf node OrgName should be unchanged")
	}
}

func TestReplace_NodeNotFound(t *testing.T) {
	child1 := CreateInvalidToken("child1")
	child2 := CreateInvalidToken("child2")

	root := createTestModulePart(child1, child2, nil)

	notInTree := CreateTokenFrom(common.EOF_TOKEN, nil, nil)
	replacement := CreateInvalidToken("replacement")

	result := Replace(root, notInTree, replacement)

	newModule, ok := result.(*STModulePart)
	if !ok {
		t.Errorf("Result is not a STModulePart")
	}

	// Nothing should be replaced - children should be the same
	if newModule.Imports != child1 {
		t.Errorf("Child Imports should be unchanged when target not found")
	}

	if newModule.Members != child2 {
		t.Errorf("Child Members should be unchanged when target not found")
	}
}

func TestToSexpr_SimpleNode(t *testing.T) {
	// Create a simple node with children
	child1 := CreateTokenFrom(common.FUNCTION_KEYWORD, nil, nil)
	child2 := CreateTokenFrom(common.OPEN_BRACE_TOKEN, nil, nil)

	node := createTestImportDeclaration(child1, child2, nil, nil, nil)

	// Convert to S-expression
	result := ToSexpr(node)

	// Hardcoded expected output
	expected := `(2000 0 0x00 ()
  (function 8 0x00 ())
  ({ 1 0x00 ()))`

	if result != expected {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result)
	}
}

func TestToSexpr_NodeList(t *testing.T) {
	// Create a node list with nested structure
	child1 := CreateTokenFrom(common.FUNCTION_KEYWORD, nil, nil)
	child2 := CreateTokenFrom(common.IDENTIFIER_TOKEN, nil, nil)

	nodeList := CreateNodeList(child1, child2)

	// Convert to S-expression
	result := ToSexpr(nodeList)

	// Hardcoded expected output
	expected := `(1 0 0x00 ()
  (function 8 0x00 ())
  (ident, "UNKNOWN" 0 0x00 ()))`

	if result != expected {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result)
	}
}
