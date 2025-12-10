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

package internal

import (
	"ballerina-lang-go/parser/common"
	"bytes"
	"encoding/json"
	"fmt"
)

// orderedJSONObject represents a JSON object with ordered fields
type orderedJSONObject struct {
	fields []fieldValue
}

type fieldValue struct {
	key   string
	value interface{}
}

func newOrderedJSONObject() *orderedJSONObject {
	return &orderedJSONObject{
		fields: make([]fieldValue, 0),
	}
}

func (oj *orderedJSONObject) set(key string, value interface{}) {
	// Check if key already exists and update it
	for i := range oj.fields {
		if oj.fields[i].key == key {
			oj.fields[i].value = value
			return
		}
	}
	// Add new field
	oj.fields = append(oj.fields, fieldValue{key: key, value: value})
}

func (oj *orderedJSONObject) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, fv := range oj.fields {
		if i > 0 {
			buf.WriteByte(',')
		}
		keyBytes, _ := json.Marshal(fv.key)
		buf.Write(keyBytes)
		buf.WriteByte(':')
		valBytes, err := json.Marshal(fv.value)
		if err != nil {
			return nil, err
		}
		buf.Write(valBytes)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// GenerateJSON converts an STNode to JSON format matching the Java SyntaxTreeJSONGenerator output
func GenerateJSON(node STNode) string {
	jsonObj := getJSON(node)
	jsonBytes, err := json.MarshalIndent(jsonObj, "", "  ")
	if err != nil {
		panic(fmt.Sprintf("Failed to marshal JSON: %v", err))
	}
	return string(jsonBytes)
}

// getJSON converts an STNode to a JSON-serializable map with ordered fields
// Uses orderedJSONObject to ensure field order matches Java output
func getJSON(node STNode) interface{} {
	if !IsSTNodePresent(node) {
		return nil
	}

	// Use ordered object to ensure "kind" comes first
	jsonNode := newOrderedJSONObject()
	kind := node.Kind()
	jsonNode.set("kind", kindName(kind))

	if node.IsMissing() {
		jsonNode.set("isMissing", true)
		addDiagnosticsOrdered(node, jsonNode)
		if isToken(node) {
			token := node.(STToken)
			addTriviaOrdered(token, jsonNode)
		}
		return jsonNode
	}

	addDiagnosticsOrdered(node, jsonNode)
	if isToken(node) {
		token := node.(STToken)
		// If the node is a terminal node with a dynamic value (i.e: non-syntax node)
		// then add the value to the json.
		// Note: "value" should come before "trailingMinutiae" in Java output
		if !isKeyword(kind) {
			jsonNode.set("value", getTokenText(token))
		}
		addTriviaOrdered(token, jsonNode)
	} else {
		addChildrenOrdered(node, jsonNode)
	}

	return jsonNode
}

// addChildrenOrdered adds the children array to an ordered JSON object
// Always include children array, even if empty (to match Java output)
func addChildrenOrdered(node STNode, jsonObj *orderedJSONObject) {
	children := make([]interface{}, 0)
	size := node.BucketCount()
	for i := 0; i < size; i++ {
		childNode := node.ChildInBucket(i)
		if childNode == nil || childNode.Kind() == common.NONE {
			continue
		}
		children = append(children, getJSON(childNode))
	}
	// Always include children array, even if empty (Java includes empty arrays)
	jsonObj.set("children", children)
}

// addTriviaOrdered adds leading and trailing minutiae to a token's JSON representation
func addTriviaOrdered(token STToken, jsonObj *orderedJSONObject) {
	leadingMinutiae := token.LeadingMinutiae()
	if leadingMinutiae != nil {
		var bucketCount int
		if IsSTNodeList(leadingMinutiae) {
			bucketCount = leadingMinutiae.(*STNodeList).Size()
		} else {
			bucketCount = leadingMinutiae.BucketCount()
		}
		if bucketCount != 0 {
			minutiaeList := addMinutiaeList(leadingMinutiae)
			if len(minutiaeList) > 0 {
				jsonObj.set("leadingMinutiae", minutiaeList)
			}
		}
	}

	trailingMinutiae := token.TrailingMinutiae()
	if trailingMinutiae != nil {
		var bucketCount int
		if IsSTNodeList(trailingMinutiae) {
			bucketCount = trailingMinutiae.(*STNodeList).Size()
		} else {
			bucketCount = trailingMinutiae.BucketCount()
		}
		if bucketCount != 0 {
			minutiaeList := addMinutiaeList(trailingMinutiae)
			if len(minutiaeList) > 0 {
				jsonObj.set("trailingMinutiae", minutiaeList)
			}
		}
	}
}

// addMinutiaeList converts a minutiae list node to a JSON array
func addMinutiaeList(minutiaeList STNode) []interface{} {
	minutiaeJsonArray := make([]interface{}, 0)
	if !IsSTNodeList(minutiaeList) {
		return minutiaeJsonArray
	}

	nodeList := minutiaeList.(*STNodeList)
	size := nodeList.Size()
	for i := 0; i < size; i++ {
		minutiae := nodeList.Get(i)
		if !IsSTNodePresent(minutiae) {
			continue
		}

		minutiaeJson := newOrderedJSONObject()
		minutiaeKind := minutiae.Kind()
		minutiaeJson.set("kind", kindName(minutiaeKind))

		switch minutiaeKind {
		case common.WHITESPACE_MINUTIAE, common.END_OF_LINE_MINUTIAE, common.COMMENT_MINUTIAE:
			if minutiaeNode, ok := minutiae.(*STMinutiae); ok {
				// "value" should come after "kind" in Java output
				minutiaeJson.set("value", minutiaeNode.text)
			}
		case common.INVALID_NODE_MINUTIAE:
			if invalidNodeMinutiae, ok := minutiae.(*STInvalidNodeMinutiae); ok {
				invalidNode := invalidNodeMinutiae.invalidNode
				minutiaeJson.set("invalidNode", getJSON(invalidNode))
			}
		default:
			panic(fmt.Sprintf("Unsupported minutiae kind: %v", minutiaeKind))
		}

		minutiaeJsonArray = append(minutiaeJsonArray, minutiaeJson)
	}
	return minutiaeJsonArray
}

// addDiagnosticsOrdered adds diagnostics to an ordered JSON object
func addDiagnosticsOrdered(node STNode, jsonObj *orderedJSONObject) {
	if !node.HasDiagnostics() {
		return
	}

	jsonObj.set("hasDiagnostics", true)
	diagnostics := node.Diagnostics()
	if len(diagnostics) == 0 {
		return
	}

	diagnosticsJsonArray := make([]interface{}, 0, len(diagnostics))
	for _, diag := range diagnostics {
		diagnosticsJsonArray = append(diagnosticsJsonArray, diag.code.DiagnosticId())
	}
	jsonObj.set("diagnostics", diagnosticsJsonArray)
}

// isKeyword checks if a SyntaxKind is a keyword
// Matches Java logic: any kind that comes before IDENTIFIER_TOKEN in enum order is a keyword
// Also includes EOF_TOKEN
func isKeyword(kind common.SyntaxKind) bool {
	return kind < common.IDENTIFIER_TOKEN || kind == common.EOF_TOKEN
}

// getTokenText extracts the text value from a token
// Matches Java ParserTestUtils.getTokenText logic
func getTokenText(token STToken) string {
	kind := token.Kind()
	switch kind {
	case common.IDENTIFIER_TOKEN:
		if identToken, ok := token.(*STIdentifierToken); ok {
			return identToken.text
		}
	case common.STRING_LITERAL_TOKEN:
		// Java removes surrounding quotes: substring(1, lastCharPosition)
		// where lastCharPosition is length-1 if ends with quote, else length
		val := token.Text()
		stringLen := len(val)
		lastCharPosition := stringLen
		if stringLen > 0 && val[stringLen-1] == '"' {
			lastCharPosition = stringLen - 1
		}
		if stringLen > 0 && lastCharPosition > 1 {
			return val[1:lastCharPosition]
		}
		return ""
	case common.DECIMAL_INTEGER_LITERAL_TOKEN,
		common.HEX_INTEGER_LITERAL_TOKEN,
		common.DECIMAL_FLOATING_POINT_LITERAL_TOKEN,
		common.HEX_FLOATING_POINT_LITERAL_TOKEN,
		common.PARAMETER_NAME,
		common.DEPRECATION_LITERAL,
		common.INVALID_TOKEN:
		// For these, try STLiteralValueToken first, then fall back to Text()
		if literalToken, ok := token.(*STLiteralValueToken); ok {
			return literalToken.text
		}
		return token.Text()
	default:
		// For other literal tokens (XML_TEXT, TEMPLATE_STRING, etc.), use Text() method
		if literalToken, ok := token.(*STLiteralValueToken); ok {
			return literalToken.text
		}
		return token.Text()
	}
	return token.Text()
}

// kindName is defined in ast_json_kindnames.go
