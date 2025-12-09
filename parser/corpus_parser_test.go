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
	"ballerina-lang-go/tools/text"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseCorpusFiles(t *testing.T) {
	// Try both relative paths - from package directory and from project root
	corpusBalDir := "../corpus/bal"
	if _, err := os.Stat(corpusBalDir); os.IsNotExist(err) {
		// Try alternative path (when running from project root)
		corpusBalDir = "./corpus/bal"
		if _, err := os.Stat(corpusBalDir); os.IsNotExist(err) {
			t.Skipf("Corpus directory not found (tried ../corpus/bal and ./corpus/bal), skipping test")
		}
	}

	// Find all .bal files
	var balFiles []string
	err := filepath.Walk(corpusBalDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".bal") {
			balFiles = append(balFiles, path)
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Error walking corpus/bal directory: %v", err)
	}

	if len(balFiles) == 0 {
		t.Fatalf("No .bal files found in %s", corpusBalDir)
	}

	// Create subtests for each file
	// Note: Running sequentially (not in parallel) to identify which files cause crashes
	for _, balFile := range balFiles {
		balFile := balFile // capture loop variable
		t.Run(balFile, func(t *testing.T) {
			// Run sequentially to identify failing files
			// t.Parallel() // Commented out to run sequentially
			parseFile(t, balFile)
		})
	}
}

func parseFile(t *testing.T, filePath string) {
	// Print which file we're processing at the beginning (always visible)
	fmt.Fprintf(os.Stderr, "Parsing: %s\n", filePath)

	// Use a helper function to catch panics completely
	err := parseFileWithRecovery(filePath)
	if err != nil {
		t.Errorf("FAILED: %s - %v", filePath, err)
	}
}

// parseFileWithRecovery parses a file and returns any error or panic as an error
func parseFileWithRecovery(filePath string) (err error) {
	// Catch any panics and convert them to errors
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	// Read file content
	content, readErr := os.ReadFile(filePath)
	if readErr != nil {
		return fmt.Errorf("error reading file: %w", readErr)
	}

	// Create CharReader from file content
	reader := text.CharReaderFromText(string(content))

	// Create Lexer (no debug context needed for tests)
	lexer := NewLexer(reader, nil)

	// Create TokenReader from Lexer
	tokenReader := CreateTokenReader(*lexer, nil)

	// Create Parser from TokenReader
	ballerinaParser := NewBallerinaParserFromTokenReader(*tokenReader)

	// Parse the entire file - this may panic
	ast := ballerinaParser.Parse()

	// Verify that Parse() returns a non-nil STNode
	if ast == nil {
		return fmt.Errorf("Parse() returned nil AST")
	}

	// Verify it's a valid STNode by checking its Kind
	if ast.Kind() == 0 {
		return fmt.Errorf("Parse() returned AST with invalid Kind (0)")
	}

	return nil
}
