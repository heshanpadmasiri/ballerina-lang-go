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
	"ballerina-lang-go/tools/diagnostics"
	"fmt"
	"strings"
)

// This represent green nodes in the syntax tree. Green nodes satisfy fallowing properties:
//   1. Immutable (except via `Replace`)
//   2. No parent reference
//   3. Know width but not position
//   4. We build them bottom up
// All green nodes satisfy STNode interface.
// All the actual nodes are generated in `st-node-gen.go` using `tree-gen` tool.
// We need these to be very memory efficient but not necessarily fast to build. We build these once but hold them
// for very long time.
//   We need to be careful about not holding lot of "medium" sized nodes for each token.
// As package name suggests, these nodes represent an internal implementation details of the parser and should not be
// exposed

type STNode interface {
	Kind() common.SyntaxKind
	Diagnostics() []STNodeDiagnostic
	Width() uint16
	WidthWithLeadingMinutiae() uint16
	WidthWithTrailingMinutiae() uint16
	WidthWithMinutiae() uint16
	Flags() uint8
	BucketCount() int
	HasDiagnostics() bool
	IsMissing() bool
	Tokens() []STToken
	FirstToken() STToken
	LastToken() STToken
	ChildBuckets() []STNode
	ChildInBucket(bucket int) STNode
	setDiagnostics(diagnostics []STNodeDiagnostic)
	ToSourceCode() string
	writeTo(builder *strings.Builder)
}

type STToken interface {
	STNode
	Text() string
	HasTrailingNewLine() bool
}

// Actual "base" types for AST nodes. We generate most of the actual nodes in st-node-gen.go.
//
//go:generate ../../compiler-tools/tree-gen/tree-gen -config ../nodes.json -type st-node -template ../../compiler-tools/tree-gen/templates/st-node.go.tmpl -output st-node-gen.go -util-template ../../compiler-tools/tree-gen/templates/st-node-util.go.tmpl -util-output st-node-util-gen.go
type (
	STTokenBase struct {
		kind                common.SyntaxKind
		width               uint16
		diagnostics         []STNodeDiagnostic
		flags               uint8
		leadingMinutiae     STNode
		trailingMinutiae    STNode
		lookbackTokenCount  int
		lookaheadTokenCount int
	}

	STMissingToken struct {
		STTokenBase
		diagnosticList []STNodeDiagnostic
	}

	STLiteralValueToken struct {
		STTokenBase
		// Ideally we don't need width here and instead calculate it on the go
		text string
	}

	// TODO: can this be based on STTokenBase as well?
	STInvalidTokenMinutiaeNode struct {
		STNodeBase
		token STToken
	}

	STInvalidToken struct {
		STTokenBase
		tokenText string
	}

	STIdentifierToken struct {
		STToken
		text string
	}

	STMinutiae struct {
		STNodeBase
		text string
	}

	STInvalidNodeMinutiae struct {
		STMinutiae
		invalidNode STNode
	}

	STNodeList struct {
		STNodeBase
	}

	STNodeBase struct {
		kind                      common.SyntaxKind
		diagnostics               []STNodeDiagnostic
		width                     uint16
		widthWithLeadingMinutiae  uint16
		widthWithTrailingMinutiae uint16
		widthWithMinutiae         uint16
		flags                     uint8
		bucketCount               int
		childBuckets              []STNode
	}

	STNodeDiagnostic struct {
		code diagnostics.DiagnosticCode
		args []any
	}
)

func (s STNodeList) IsEmpty() bool {
	return s.bucketCount == 0
}

// Shared common nodes
var (
	emptyNodeList = &STNodeList{
		STNodeBase: STNodeBase{
			kind: common.LIST,
		},
	}
)

// Methods for creating STNodes
func CreateInvalidTokenMinutiaeNode(token STToken) *STInvalidTokenMinutiaeNode {
	// TODO: update diagnostics
	return &STInvalidTokenMinutiaeNode{
		STNodeBase: STNodeBase{
			kind:         common.INVALID_TOKEN_MINUTIAE_NODE,
			childBuckets: []STNode{token},
			bucketCount:  1,
		},
		token: token,
	}
}

func CreateLiteralValueToken(kind common.SyntaxKind,
	text string,
	leadingTrivia STNode,
	trailingTrivia STNode) STToken {
	return &STLiteralValueToken{
		STTokenBase: STTokenBase{
			kind:             kind,
			width:            uint16(len(text)),
			leadingMinutiae:  leadingTrivia,
			trailingMinutiae: trailingTrivia,
		},
		text: text,
	}
}

func CreateDiagnostic(code diagnostics.DiagnosticCode, args ...any) STNodeDiagnostic {
	return STNodeDiagnostic{
		code: code,
		args: args,
	}
}

func CreateTokenFrom(kind common.SyntaxKind, leadingMinutiae STNode, trailingMinutiae STNode) STToken {
	width := uint16(len(kind.StrValue()))
	return &STTokenBase{
		kind:             kind,
		width:            width,
		leadingMinutiae:  leadingMinutiae,
		trailingMinutiae: trailingMinutiae,
	}
}

func CreateNodeList(nodes []STNode) STNodeList {
	// Return shared empty instance for empty lists
	if len(nodes) == 0 {
		return *emptyNodeList
	}
	return STNodeList{
		STNodeBase: STNodeBase{
			kind:         common.LIST,
			childBuckets: nodes,
		},
	}
}

func CreateEmptyNodeList() STNode {
	return emptyNodeList
}

func CreateMinutiae(kind common.SyntaxKind, text string) *STMinutiae {
	return &STMinutiae{
		STNodeBase: STNodeBase{
			kind: kind,
		},
		text: text,
	}
}

func CreateInvalidToken(tokenText string) *STInvalidToken {
	return &STInvalidToken{
		STTokenBase: STTokenBase{
			kind:  common.INVALID_TOKEN,
			width: uint16(len(tokenText)),
		},
		tokenText: tokenText,
	}
}

func CreateInvalidNodeMinutiae(invalidToken STInvalidToken) STNode {
	return CreateInvalidTokenMinutiaeNode(&invalidToken)
}

func CreateIdentifierToken(text string, leadingTrivia STNode, trailingTrivia STNode) STToken {
	return &STIdentifierToken{
		STToken: &STTokenBase{
			kind:             common.IDENTIFIER_TOKEN,
			width:            uint16(len(text)),
			leadingMinutiae:  leadingTrivia,
			trailingMinutiae: trailingTrivia,
		},
		text: text,
	}
}

// findToken searches for a token in the child buckets.
// Direction determines iteration order: forward searches from first to last child,
// backward searches from last to first child.
func (n *STNodeBase) findToken(dir direction) STToken {
	start, end, step := 0, len(n.childBuckets), 1
	if dir == backward {
		start, end, step = len(n.childBuckets)-1, -1, -1
	}

	for i := start; i != end; i += step {
		child := n.childBuckets[i]
		if isToken(child) {
			token, ok := child.(*STTokenBase)
			if ok {
				return token
			}
			panic("expected STToken")
		}
		if (!IsSTNodePresent(child) || IsSTNodeList(child)) && child.BucketCount() == 0 {
			continue
		}
		var token STToken
		if dir == forward {
			token = child.FirstToken()
		} else {
			token = child.LastToken()
		}
		if IsSTNodePresent(token) {
			return token
		}
	}
	if dir == forward {
		panic("failed to find first token")
	}
	panic("failed to find last token")
}
func (n STNodeBase) FirstToken() STToken {
	return n.findToken(forward)
}

func (n STNodeBase) LastToken() STToken {
	return n.findToken(backward)
}

func (t STNodeBase) ToSourceCode() string {
	builder := strings.Builder{}
	t.writeTo(&builder)
	return builder.String()
}

func (t STNodeBase) writeTo(builder *strings.Builder) {
	for _, child := range t.childBuckets {
		if IsSTNodePresent(child) {
			child.writeTo(builder)
		}
	}
}

func (n STNodeBase) setDiagnostics(diagnostics []STNodeDiagnostic) {
	n.diagnostics = diagnostics
}

func (n STNodeBase) ChildInBucket(bucket int) STNode {
	return n.childBuckets[bucket]
}

func (n *STNodeBase) copy() *STNodeBase {
	diagnosticsCopy := make([]STNodeDiagnostic, len(n.diagnostics))
	copy(diagnosticsCopy, n.diagnostics)

	childBucketsCopy := make([]STNode, len(n.childBuckets))
	copy(childBucketsCopy, n.childBuckets)

	return &STNodeBase{
		kind:                      n.kind,
		diagnostics:               diagnosticsCopy,
		width:                     n.width,
		widthWithLeadingMinutiae:  n.widthWithLeadingMinutiae,
		widthWithTrailingMinutiae: n.widthWithTrailingMinutiae,
		widthWithMinutiae:         n.widthWithMinutiae,
		flags:                     n.flags,
		bucketCount:               n.bucketCount,
		childBuckets:              childBucketsCopy,
	}
}

func (n STNodeBase) Kind() common.SyntaxKind {
	return n.kind
}

func (n STNodeBase) Diagnostics() []STNodeDiagnostic {
	return n.diagnostics
}

func (n STNodeBase) Width() uint16 {
	return n.width
}

func (n STNodeBase) WidthWithLeadingMinutiae() uint16 {
	return n.widthWithLeadingMinutiae
}

func (n STNodeBase) WidthWithTrailingMinutiae() uint16 {
	return n.widthWithTrailingMinutiae
}

func (n STNodeBase) WidthWithMinutiae() uint16 {
	return n.widthWithMinutiae
}

func (n STNodeBase) Flags() uint8 {
	return n.flags
}

func (n STNodeBase) BucketCount() int {
	return n.bucketCount
}

func (n STNodeBase) ChildBuckets() []STNode {
	return n.childBuckets
}

func (n STNodeBase) HasDiagnostics() bool {
	return isFlagSet(n.flags, HAS_DIAGNOSTIC)
}

func (n STNodeBase) IsMissing() bool {
	return isFlagSet(n.flags, IS_MISSING)
}

func (n STNodeBase) Tokens() []STToken {
	tokens := make([]STToken, 0, len(n.childBuckets))
	for _, child := range n.childBuckets {
		if IsSTNodePresent(child) {
			tokens = append(tokens, child.Tokens()...)
		}
	}
	return tokens
}

func (n STTokenBase) Kind() common.SyntaxKind {
	return n.kind
}

func (n STTokenBase) Diagnostics() []STNodeDiagnostic {
	return n.diagnostics
}

func (n STTokenBase) Width() uint16 {
	return n.width
}

func (n STTokenBase) WidthWithLeadingMinutiae() uint16 {
	return n.width + n.leadingMinutiae.Width()
}

func (n STTokenBase) WidthWithTrailingMinutiae() uint16 {
	return n.width + n.trailingMinutiae.Width()
}

func (n STTokenBase) WidthWithMinutiae() uint16 {
	return n.width + n.leadingMinutiae.Width() + n.trailingMinutiae.Width()
}

func (n STTokenBase) Flags() uint8 {
	return n.flags
}

func (n STTokenBase) BucketCount() int {
	return 0
}

func (n STTokenBase) ChildBuckets() []STNode {
	return nil
}

func (n STTokenBase) HasDiagnostics() bool {
	return isFlagSet(n.flags, HAS_DIAGNOSTIC)
}

func (n STTokenBase) ChildInBucket(bucket int) STNode {
	panic("ChildInBucket is not supported for STToken")
}

func (n STTokenBase) IsMissing() bool {
	return isFlagSet(n.flags, IS_MISSING)
}

func (n STTokenBase) Tokens() []STToken {
	return nil
}

func (t STTokenBase) Text() string {
	return t.kind.StrValue()
}

func (t STTokenBase) FirstToken() STToken {
	return &t
}

func (t STTokenBase) LastToken() STToken {
	return &t
}

func (t STTokenBase) HasTrailingNewLine() bool {
	stNodeList := t.trailingMinutiae.(*STNodeList)
	for i := 0; i < stNodeList.size(); i++ {
		if stNodeList.get(i).Kind() == common.END_OF_LINE_MINUTIAE {
			return true
		}
	}
	return false
}

func (t STTokenBase) ToSourceCode() string {
	builder := strings.Builder{}
	t.writeTo(&builder)
	return builder.String()
}

func (t STTokenBase) writeTo(builder *strings.Builder) {
	t.leadingMinutiae.writeTo(builder)
	builder.WriteString(t.kind.StrValue())
	t.trailingMinutiae.writeTo(builder)
}

func (t STTokenBase) setDiagnostics(diagnostics []STNodeDiagnostic) {
	t.diagnostics = diagnostics
}

func (s *STNodeList) get(i int) STNode {
	rangeCheck(i, s.bucketCount)
	return s.childBuckets[i]
}

func (s *STNodeList) add(node STNode) *STNodeList {
	STNodeBase := s.copy()
	STNodeBase.childBuckets = append(STNodeBase.childBuckets, node)
	STNodeBase.bucketCount++
	return &STNodeList{STNodeBase: *STNodeBase}
}

func (s STNodeList) BucketCount() int {
	return s.bucketCount
}

func (s *STNodeList) size() int {
	return s.bucketCount
}

func (s STNodeDiagnostic) DiagnosticCode() diagnostics.DiagnosticCode {
	return s.code
}

// Modification methods
func Replace(current STNode, target STNode, replacement STNode) STNode {
	// TODO: this is doing value comparison which is super expensive, need to think of a better way to do this
	_, result := replaceInner(current, target, replacement)
	return result
}

func AddSyntaxDiagnostic[T STNode](node T, diagnostic STNodeDiagnostic) T {
	return AddSyntaxDiagnostics(node, []STNodeDiagnostic{diagnostic})
}

func AddSyntaxDiagnostics[T STNode](node T, diagnostics []STNodeDiagnostic) T {
	if len(diagnostics) == 0 {
		return node
	}

	oldDiagnostics := node.Diagnostics()
	if len(oldDiagnostics) == 0 {
		return modifyWithDiagnostics(node, diagnostics)
	}

	// Merge all diagnostics
	newDiagnostics := make([]STNodeDiagnostic, len(oldDiagnostics))
	copy(newDiagnostics, oldDiagnostics)
	newDiagnostics = append(newDiagnostics, diagnostics...)
	return modifyWithDiagnostics(node, newDiagnostics)
}

func modifyWithDiagnostics[T STNode](base T, diagnostics []STNodeDiagnostic) T {
	copy := base
	copy.setDiagnostics(diagnostics)
	return copy
}

// Utility functions
func rangeCheck(index, size int) {
	if index >= size || index < 0 {
		panic(fmt.Sprintf("index out of bounds: %d, size: %d", index, size))
	}
}

// getLiteralTokenName returns the name for literal tokens used in S-expression format
func getLiteralTokenName(kind common.SyntaxKind) string {
	switch kind {
	case common.IDENTIFIER_TOKEN:
		return "ident"
	case common.STRING_LITERAL_TOKEN:
		return "string"
	case common.DECIMAL_INTEGER_LITERAL_TOKEN:
		return "int"
	case common.HEX_INTEGER_LITERAL_TOKEN:
		return "hexInt"
	case common.DECIMAL_FLOATING_POINT_LITERAL_TOKEN:
		return "float"
	case common.HEX_FLOATING_POINT_LITERAL_TOKEN:
		return "hexFloat"
	case common.XML_TEXT_CONTENT:
		return "xmlText"
	case common.TEMPLATE_STRING:
		return "templateString"
	case common.PROMPT_CONTENT:
		return "promptContent"
	default:
		return ""
	}
}

// toSexpr converts an STNode to S-expression format: (Kind width flags (diagnostics) *children)
func ToSexpr(node STNode) string {
	return toSexprIndented(node, 0)
}

// toSexprIndented converts an STNode to S-expression format with indentation
func toSexprIndented(node STNode, indentLevel int) string {
	if !IsSTNodePresent(node) {
		return "nil"
	}

	var builder strings.Builder
	nextIndent := strings.Repeat("  ", indentLevel+1)

	// Start the S-expression: (
	builder.WriteString("(")

	// Kind
	kind := node.Kind()
	kindStr := kind.StrValue()
	if kindStr == "" {
		kindStr = fmt.Sprintf("%d", kind.Tag())
	}
	// Special case for literal tokens
	literalName := getLiteralTokenName(kind)
	if literalName != "" {
		var text string
		// Try STIdentifierToken first
		if identToken, ok := node.(*STIdentifierToken); ok {
			text = identToken.text
		} else if literalToken, ok := node.(*STLiteralValueToken); ok {
			// Try STLiteralValueToken for other literal tokens
			text = literalToken.text
		} else {
			// Can this happen?
			text = "UNKNOWN"
		}
		kindStr = fmt.Sprintf("%s, \"%s\"", literalName, text)
	}
	builder.WriteString(kindStr)

	// Width
	builder.WriteString(" ")
	builder.WriteString(fmt.Sprintf("%d", node.Width()))

	// Flags (hex format)
	builder.WriteString(" ")
	builder.WriteString(fmt.Sprintf("0x%02x", node.Flags()))

	// Diagnostics
	builder.WriteString(" (")
	diagnostics := node.Diagnostics()
	for i, diag := range diagnostics {
		if i > 0 {
			builder.WriteString(" ")
		}
		// Format diagnostic as: DiagnosticId arg1 arg2 ...
		builder.WriteString(diag.code.DiagnosticId())
		for _, arg := range diag.args {
			builder.WriteString(" ")
			builder.WriteString(fmt.Sprintf("%v", arg))
		}
	}
	builder.WriteString(")")

	// Children
	children := node.ChildBuckets()
	// Add each child on a new line with indentation
	for _, child := range children {
		builder.WriteString("\n")
		builder.WriteString(nextIndent)
		builder.WriteString(toSexprIndented(child, indentLevel+1))
	}

	builder.WriteString(")")
	return builder.String()
}

type direction uint8

const (
	forward direction = iota
	backward
)

// Flag constants for STNode
const (
	HAS_DIAGNOSTIC uint8 = 1 << 1 // 0x02
	IS_MISSING     uint8 = 1 << 2 // 0x04
)

// isFlagSet checks whether the given flag is set in the given flags.
func isFlagSet(flags uint8, flag uint8) bool {
	return (flags & flag) != 0
}

func IsSTNodeList(child STNode) bool {
	return child.Kind() == common.LIST
}

func IsSTNodePresent(child STNode) bool {
	return child != nil
}

func isToken(node STNode) bool {
	_, ok := node.(STToken)
	return ok
}

func CloneWithLeadingInvalidNodeMinutiae(toClone STNode, invalidNode STNode, diagnosticCode diagnostics.DiagnosticCode, args ...any) STNode {
	firstToken := toClone.FirstToken()
	firstTokenWithInvalidNodeMinutiae := CloneWithLeadingInvalidNodeMinutiae(firstToken,
		invalidNode, diagnosticCode, args)
	return Replace(toClone, firstToken, firstTokenWithInvalidNodeMinutiae)
}

func CreateMissingTokenWithDiagnosticsFromParserRules(expectedKind common.SyntaxKind, currentCtx common.ParserRuleContext) STToken {
	return CreateMissingTokenWithDiagnostics(expectedKind, currentCtx.GetErrorCode())
}

func CreateMissingTokenWithDiagnostics(expectedKind common.SyntaxKind, diagnosticCode diagnostics.DiagnosticCode) STToken {
	diagnosticList := []STNodeDiagnostic{CreateDiagnosticWithArgs(diagnosticCode)}
	return CreateMissingToken(expectedKind, diagnosticList)
}

func CreateDiagnosticWithArgs(diagnosticCode diagnostics.DiagnosticCode, args ...any) STNodeDiagnostic {
	return STNodeDiagnostic{
		code: diagnosticCode,
		args: args,
	}
}

func CreateMissingToken(expectedKind common.SyntaxKind, diagnosticList []STNodeDiagnostic) STToken {
	return &STMissingToken{
		STTokenBase: STTokenBase{
			kind: expectedKind,
		},
		diagnosticList: diagnosticList,
	}
}

func AddDiagnostic(node STNode, diagnosticCode diagnostics.DiagnosticCode, args ...any) STNode {
	return AddSyntaxDiagnostic(node, CreateDiagnostic(diagnosticCode, args...))
}

func CloneWithTrailingInvalidNodeMinutiae(toClone STNode, invalidNode STNode, diagnosticCode diagnostics.DiagnosticCode, args ...any) STNode {
	lastToken := toClone.LastToken()
	lastTokenWithInvalidNodeMinutiae := CloneWithTrailingInvalidNodeMinutiae(lastToken,
		invalidNode, diagnosticCode, args)
	return Replace(toClone, lastToken, lastTokenWithInvalidNodeMinutiae)
}

func ToToken(node STNode) STToken {
	tok, ok := node.(STToken)
	if !ok {
		panic("node is not a STToken")
	}
	return tok
}
