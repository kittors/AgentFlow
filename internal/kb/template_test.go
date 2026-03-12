package kb

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitPathsCreatesCoreKBFiles(t *testing.T) {
	root := t.TempDir()
	summary, err := InitPaths(root)
	if err != nil {
		t.Fatalf("InitPaths returned error: %v", err)
	}
	if len(summary.FilesCreated) != 3 {
		t.Fatalf("expected 3 files created, got %#v", summary.FilesCreated)
	}

	for _, name := range []string{"INDEX.md", "context.md", "CHANGELOG.md"} {
		if _, err := os.Stat(filepath.Join(root, ".agentflow", "kb", name)); err != nil {
			t.Fatalf("expected %s to exist: %v", name, err)
		}
	}
}
