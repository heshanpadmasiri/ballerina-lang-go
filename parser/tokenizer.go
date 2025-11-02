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
	debugcommon "ballerina-lang-go/common"
	"ballerina-lang-go/parser/internal"
)

type TokenReader struct {
	lexer        Lexer
	dbgContext   *debugcommon.DebugContext
	currentToken internal.STToken
	tokenBuffer  tokenBuffer
}

func CreateTokenReader(lexer Lexer, dbgContext *debugcommon.DebugContext) *TokenReader {
	return &TokenReader{
		lexer:        lexer,
		dbgContext:   dbgContext,
		currentToken: nil,
		tokenBuffer: tokenBuffer{
			tokens: make([]internal.STToken, BUFFER_SIZE),
		},
	}
}

func (t *TokenReader) Read() internal.STToken {
	if t.tokenBuffer.size > 0 {
		t.currentToken = t.tokenBuffer.consume()
	} else {
		t.currentToken = t.lexer.NextToken()
	}
	return t.currentToken
}

func (t *TokenReader) Peek() internal.STToken {
	if t.tokenBuffer.size > 0 {
		return t.tokenBuffer.peek()
	} else {
		token := t.lexer.NextToken()
		t.tokenBuffer.add(token)
		return token
	}
}

func (t *TokenReader) PeekN(n int) internal.STToken {
	if n >= BUFFER_SIZE {
		panic("n is too large")
	}
	remaining := n - t.tokenBuffer.size
	for remaining > 0 {
		token := t.lexer.NextToken()
		t.tokenBuffer.add(token)
		remaining--
	}
	return t.tokenBuffer.peekN(n)
}

func (t *TokenReader) Head() internal.STToken {
	return t.currentToken
}

func (t *TokenReader) StartMode(mode ParserMode) {
	t.lexer.StartMode(mode)
}

func (t *TokenReader) SwitchMode(mode ParserMode) {
	t.lexer.SwitchMode(mode)
}

func (t *TokenReader) EndMode() {
	t.lexer.EndMode()
}

const BUFFER_SIZE = 20

type tokenBuffer struct {
	size        int
	cursorIndex int
	insertIndex int
	tokens      []internal.STToken
}

func (t *tokenBuffer) add(token internal.STToken) {
	t.tokens[t.insertIndex] = token
	t.insertIndex = (t.insertIndex + 1) % BUFFER_SIZE
	t.size++
	if t.size == BUFFER_SIZE {
		panic("buffer overflow")
	}
}

func (t *tokenBuffer) peek() internal.STToken {
	return t.tokens[t.cursorIndex]
}

func (t *tokenBuffer) peekN(n int) internal.STToken {
	if n >= BUFFER_SIZE {
		panic("n is too large")
	}
	return t.tokens[t.cursorIndex+n]
}

func (t *tokenBuffer) consume() internal.STToken {
	if t.size == 0 {
		panic("no tokens to consume")
	}
	token := t.tokens[t.cursorIndex]
	t.cursorIndex = (t.cursorIndex + 1) % BUFFER_SIZE
	t.size--
	return token
}
