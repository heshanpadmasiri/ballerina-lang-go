/*
 * Copyright (c) 2025, WSO2 LLC. (http://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package main

import (
	debugcommon "ballerina-lang-go/common"
	"ballerina-lang-go/parser"
	"ballerina-lang-go/tools/text"
	"fmt"
	"os"
	"sync"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <file.bal> [-dump-tokens] [-dump-ast]\n", os.Args[0])
		os.Exit(1)
	}

	fileName := os.Args[1]
	dumpTokens := false
	dumpAST := false

	// Check for flags
	for _, arg := range os.Args[2:] {
		if arg == "-dump-tokens" {
			dumpTokens = true
		} else if arg == "-dump-ast" {
			dumpAST = true
		} else if len(arg) > 0 && arg[0] == '-' {
			panic(fmt.Sprintf("unsupported flag: %s", arg))
		}
	}

	// Initialize DebugContext if any dump flags are enabled
	var debugCtx *debugcommon.DebugContext
	var wg sync.WaitGroup
	flags := uint16(0)
	if dumpTokens {
		flags |= debugcommon.DUMP_TOKENS
	}
	if dumpAST {
		flags |= debugcommon.DUMP_AST
	}
	if flags != 0 {
		debugcommon.Init(flags)
		debugCtx = &debugcommon.DebugCtx

		// Start a goroutine to listen to the channel and print to stderr
		wg.Add(1)
		go func() {
			defer wg.Done()
			for msg := range debugCtx.Channel {
				fmt.Fprintf(os.Stderr, "%s\n", msg)
			}
		}()
	}

	// Read the file
	content, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading file %s: %v\n", fileName, err)
		os.Exit(1)
	}

	// Create CharReader from file content
	reader := text.CharReaderFromText(string(content))

	// Create Lexer with DebugContext
	lexer := parser.NewLexer(reader, debugCtx)

	// Create TokenReader from Lexer
	tokenReader := parser.CreateTokenReader(*lexer, debugCtx)

	// Create Parser from TokenReader
	ballerinaParser := parser.NewBallerinaParserFromTokenReader(tokenReader)

	// Parse the entire file (parser will internally call tokenizer)
	_ = ballerinaParser.Parse()

	if debugCtx != nil {
		close(debugCtx.Channel)
		wg.Wait()
	}
}
