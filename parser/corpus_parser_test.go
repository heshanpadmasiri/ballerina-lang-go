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
	"ballerina-lang-go/parser/internal"
	"ballerina-lang-go/tools/text"
	"encoding/json"
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
	total := len(balFiles)
	for i, balFile := range balFiles {
		balFile := balFile // capture loop variable
		index := i + 1
		t.Run(balFile, func(t *testing.T) {
			// Run sequentially to identify failing files
			// t.Parallel() // Commented out to run sequentially
			parseFile(t, balFile, index, total)
		})
	}
}

func parseFile(t *testing.T, filePath string, index int, total int) {
	// Print file name at the beginning (before parsing) so we can see it even if stack overflow occurs
	fmt.Fprintf(os.Stderr, "[%d/%d] %s ... ", index, total, filePath)

	// Use a helper function to catch panics completely
	err := parseFileWithRecovery(filePath)

	// Print success/failure at the end
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAILED: %v\n", err)
		t.Errorf("FAILED: %s - %v", filePath, err)
	} else {
		fmt.Fprintf(os.Stderr, "SUCCESS\n")
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
	ballerinaParser := NewBallerinaParserFromTokenReader(tokenReader)

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

	// Generate JSON from the parsed AST
	actualJSON := internal.GenerateJSON(ast)

	// Determine expected JSON file path
	// Replace .bal with .json and change directory from corpus/bal to corpus/parser
	expectedJSONPath := strings.TrimSuffix(filePath, ".bal") + ".json"
	expectedJSONPath = strings.Replace(expectedJSONPath, string(filepath.Separator)+"corpus"+string(filepath.Separator)+"bal"+string(filepath.Separator), string(filepath.Separator)+"corpus"+string(filepath.Separator)+"parser"+string(filepath.Separator), 1)

	// Read expected JSON file
	expectedJSONBytes, readErr := os.ReadFile(expectedJSONPath)
	if readErr != nil {
		// If expected JSON file doesn't exist, skip this file
		if os.IsNotExist(readErr) {
			return fmt.Errorf("expected JSON file not found: %s (skipping)", expectedJSONPath)
		}
		return fmt.Errorf("error reading expected JSON file: %w", readErr)
	}

	expectedJSON := string(expectedJSONBytes)

	// Normalize both JSON strings by parsing and re-marshaling to handle whitespace differences
	var expectedObj, actualObj interface{}
	if err := json.Unmarshal([]byte(expectedJSON), &expectedObj); err == nil {
		if normalized, err := json.MarshalIndent(expectedObj, "", "  "); err == nil {
			expectedJSON = string(normalized)
		}
	}
	if err := json.Unmarshal([]byte(actualJSON), &actualObj); err == nil {
		if normalized, err := json.MarshalIndent(actualObj, "", "  "); err == nil {
			actualJSON = string(normalized)
		}
	}

	// Compare JSON strings exactly (no tolerance for formatting differences)
	if actualJSON != expectedJSON {
		// Split into lines for line-by-line comparison
		expectedLines := strings.Split(expectedJSON, "\n")
		actualLines := strings.Split(actualJSON, "\n")

		// Build detailed diff showing line numbers and differences
		var diffBuilder strings.Builder
		diffBuilder.WriteString("\nJSON mismatch - showing differences:\n\n")

		maxLines := len(expectedLines)
		if len(actualLines) > maxLines {
			maxLines = len(actualLines)
		}

		diffCount := 0
		const maxDiffsToShow = 20

		// Show line-by-line differences
		for i := 0; i < maxLines && diffCount < maxDiffsToShow; i++ {
			lineNum := i + 1
			expectedLine := ""
			actualLine := ""

			if i < len(expectedLines) {
				expectedLine = expectedLines[i]
			}
			if i < len(actualLines) {
				actualLine = actualLines[i]
			}

			if expectedLine != actualLine {
				diffCount++
				diffBuilder.WriteString(fmt.Sprintf("Line %d:\n", lineNum))
				if expectedLine == "" {
					diffBuilder.WriteString("  Expected: (empty)\n")
				} else {
					diffBuilder.WriteString(fmt.Sprintf("  Expected: %s\n", expectedLine))
				}
				if actualLine == "" {
					diffBuilder.WriteString("  Actual:   (empty)\n\n")
				} else {
					diffBuilder.WriteString(fmt.Sprintf("  Actual:   %s\n\n", actualLine))
				}
			}
		}

		if diffCount >= maxDiffsToShow {
			diffBuilder.WriteString(fmt.Sprintf("... (showing first %d differences, more exist)\n", maxDiffsToShow))
		}

		diffBuilder.WriteString(fmt.Sprintf("Total lines different: %d+\n", diffCount))
		diffBuilder.WriteString("Use diff tool for full comparison\n")

		return fmt.Errorf("JSON mismatch for %s\nExpected file: %s\n%s", filePath, expectedJSONPath, diffBuilder.String())
	}

	return nil
}
