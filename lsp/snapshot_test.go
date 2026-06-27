// Copyright (c) 2026, WSO2 LLC. (http://www.wso2.com).
//
// WSO2 LLC. licenses this file to you under the Apache License,
// Version 2.0 (the "License"); you may not use this file except
// in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lsp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"ballerina-lang-go/lsp/protocol"
)

func TestDidSaveDoesNotCreateSnapshotWhenContentUnchanged(t *testing.T) {
	root := t.TempDir()
	writeBuildProjectFiles(t, root, "public function main() {}")
	mainPath := filepath.Join(root, "main.bal")
	uri := uriFromPath(mainPath)

	server := NewServer(nil, nil)
	server.root = root
	server.snapshots[root] = NewBuildSnapshotManager(root)
	before := server.snapshots[root].Current().ID
	text := "public function main() {}"
	params, err := json.Marshal(protocol.DidSaveTextDocumentParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: uri},
		Text:         &text,
	})
	if err != nil {
		t.Fatal(err)
	}

	server.handleNotification("textDocument/didSave", params)

	if after := server.snapshots[root].Current().ID; after != before {
		t.Fatalf("snapshot ID = %d, want %d", after, before)
	}
}

func TestBuildSnapshotCanRefreshOpenFileContent(t *testing.T) {
	root := t.TempDir()
	writeBuildProjectFiles(t, root, "public function main() {}")
	mainPath := filepath.Join(root, "main.bal")

	manager := NewBuildSnapshotManager(root)
	old := manager.Current()
	uri := uriFromPath(mainPath)
	updated := SourceFile{
		URI:     uri,
		Path:    mainPath,
		File:    mainPath,
		Version: 1,
		Content: "public function main() { int i = 1; }",
		Open:    true,
	}

	newSnapshot := nextBuildSnapshot(old, func(files map[protocol.DocumentURI]SourceFile) {
		files[uri] = updated
	})
	if got := newSnapshot.Files[uri].Content; got != updated.Content {
		t.Fatalf("content = %q, want %q", got, updated.Content)
	}
}

func TestBuildSnapshotResetsGenerationAndCompilerEnvironment(t *testing.T) {
	root := t.TempDir()
	writeBuildProjectFiles(t, root, "public function main() {}")

	manager := NewBuildSnapshotManager(root)
	old := manager.Current()
	old.ID = maxIncrementalSnapshotID
	old.Modules[defaultModuleName].Stage = FrontendStageCFGAnalyzed

	newSnapshot := nextBuildSnapshot(old, nil)
	if newSnapshot.ID != initialSnapshotID {
		t.Fatalf("snapshot ID = %d, want %d", newSnapshot.ID, initialSnapshotID)
	}
	if newSnapshot.Env == old.Env {
		t.Fatal("compiler environment was reused after generation reset")
	}
	if got := newSnapshot.Modules[defaultModuleName].Stage; got != FrontendStageNone {
		t.Fatalf("module stage = %d, want %d", got, FrontendStageNone)
	}
}

func TestProjectModeUsesInitializedRootAsSnapshotKey(t *testing.T) {
	root := t.TempDir()
	writeBuildProjectFiles(t, root, "public function main() {}")
	nested := filepath.Join(root, "nested")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	writeBuildProjectFiles(t, nested, "public function nested() {}")

	server := NewServerWithMode(nil, nil, ServerModeProject)
	server.root = root
	path := filepath.Join(nested, "main.bal")
	if got := server.snapshotKey(SourceFile{URI: uriFromPath(path), Path: path, File: path}); got != root {
		t.Fatalf("snapshot key = %q, want %q", got, root)
	}
}

func TestSingleFileModeMaintainsOneSnapshot(t *testing.T) {
	root := t.TempDir()
	firstPath := filepath.Join(root, "first.bal")
	secondPath := filepath.Join(root, "second.bal")
	firstURI := uriFromPath(firstPath)
	secondURI := uriFromPath(secondPath)
	server := NewServerWithMode(nil, nil, ServerModeSingleFile)

	server.handleNotification("textDocument/didOpen", mustMarshal(protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{URI: firstURI, Version: 1, Text: "function first() {}"},
	}))
	server.handleNotification("textDocument/didOpen", mustMarshal(protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{URI: secondURI, Version: 1, Text: "function second() {}"},
	}))

	if len(server.snapshots) != 1 {
		t.Fatalf("snapshot count = %d, want 1", len(server.snapshots))
	}
	if server.singleFileURI != firstURI {
		t.Fatalf("active single file URI = %q, want %q", server.singleFileURI, firstURI)
	}
	if _, ok := server.snapshots[firstPath]; !ok {
		t.Fatalf("snapshot for first file was not created")
	}
	if _, ok := server.snapshots[secondPath]; ok {
		t.Fatalf("snapshot for second file was created")
	}
}

func writeBuildProjectFiles(t *testing.T, root string, mainContent string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, "Ballerina.toml"), []byte(`[package]
org = "testorg"
name = "sample"
version = "0.1.0"
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "main.bal"), []byte(mainContent), 0o644); err != nil {
		t.Fatal(err)
	}
}
