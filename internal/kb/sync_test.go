package kb

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSyncModulesWritesIndexAndModuleFiles(t *testing.T) {
	root := t.TempDir()
	moduleDir := filepath.Join(root, "internal", "demo")
	if err := os.MkdirAll(moduleDir, 0o755); err != nil {
		t.Fatalf("mkdir module dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(moduleDir, "main.go"), []byte("package demo\n"), 0o644); err != nil {
		t.Fatalf("write go source file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(moduleDir, "extra.ts"), []byte("export const x = 1;\n"), 0o644); err != nil {
		t.Fatalf("write ts file: %v", err)
	}

	summary, err := SyncModules(root, nil)
	if err != nil {
		t.Fatalf("SyncModules returned error: %v", err)
	}
	if summary.ModulesFound != 1 || summary.FilesWritten < 2 {
		t.Fatalf("unexpected summary: %#v", summary)
	}

	indexContent, err := os.ReadFile(filepath.Join(root, ".agentflow", "kb", "modules", "_index.md"))
	if err != nil {
		t.Fatalf("read index: %v", err)
	}
	if !strings.HasPrefix(string(indexContent), "# Module Index") {
		t.Fatalf("unexpected index content: %s", string(indexContent))
	}
}

func TestSyncModulesIncludesRootLevelGoFiles(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("write root go file: %v", err)
	}

	summary, err := SyncModules(root, nil)
	if err != nil {
		t.Fatalf("SyncModules returned error: %v", err)
	}
	if summary.ModulesFound != 1 {
		t.Fatalf("expected root module to be included, got %#v", summary)
	}

	content, err := os.ReadFile(filepath.Join(root, ".agentflow", "kb", "modules", "project-root.md"))
	if err != nil {
		t.Fatalf("read root module doc: %v", err)
	}
	if !strings.Contains(string(content), "main.go") || !strings.Contains(string(content), "`.`") {
		t.Fatalf("expected root-level Go file to be documented, got %s", string(content))
	}
}

func TestSyncModulesIncludesGoFiles(t *testing.T) {
	root := t.TempDir()
	moduleDir := filepath.Join(root, "internal", "demo")
	if err := os.MkdirAll(moduleDir, 0o755); err != nil {
		t.Fatalf("mkdir module dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(moduleDir, "main.go"), []byte("package demo\n"), 0o644); err != nil {
		t.Fatalf("write go source file: %v", err)
	}

	summary, err := SyncModules(root, nil)
	if err != nil {
		t.Fatalf("SyncModules returned error: %v", err)
	}
	if summary.ModulesFound != 1 {
		t.Fatalf("expected 1 module, got %#v", summary)
	}

	content, err := os.ReadFile(filepath.Join(root, ".agentflow", "kb", "modules", "demo.md"))
	if err != nil {
		t.Fatalf("read module doc: %v", err)
	}
	if !strings.Contains(string(content), "main.go") {
		t.Fatalf("expected Go source file to be documented, got %s", string(content))
	}
}
