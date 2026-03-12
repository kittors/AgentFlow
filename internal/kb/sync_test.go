package kb

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSyncModulesWritesIndexAndModuleFiles(t *testing.T) {
	root := t.TempDir()
	moduleDir := filepath.Join(root, filepath.Base(root), "demo")
	if err := os.MkdirAll(moduleDir, 0o755); err != nil {
		t.Fatalf("mkdir module dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(moduleDir, "main.py"), []byte("print('hi')\n"), 0o644); err != nil {
		t.Fatalf("write source file: %v", err)
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
