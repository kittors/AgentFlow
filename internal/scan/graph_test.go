package scan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildGraphWritesVersionedArtifacts(t *testing.T) {
	root := t.TempDir()
	srcDir := filepath.Join(root, "src", "demo")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatalf("mkdir src dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "alpha.py"), []byte("def exported():\n    return 1\n"), 0o644); err != nil {
		t.Fatalf("write alpha.py: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "beta.py"), []byte("import alpha\n\nalpha.exported()\n"), 0o644); err != nil {
		t.Fatalf("write beta.py: %v", err)
	}

	summary, err := BuildGraph(root, nil)
	if err != nil {
		t.Fatalf("BuildGraph returned error: %v", err)
	}
	if summary.NodeCount == 0 || summary.EdgeCount == 0 {
		t.Fatalf("expected non-empty graph summary, got %#v", summary)
	}

	nodes, err := os.ReadFile(filepath.Join(root, ".agentflow", "kb", "graph", "nodes.json"))
	if err != nil {
		t.Fatalf("read nodes.json: %v", err)
	}
	if !strings.Contains(string(nodes), `"version": 2`) {
		t.Fatalf("expected version 2 in nodes.json, got %s", string(nodes))
	}

	mermaid, err := os.ReadFile(filepath.Join(root, ".agentflow", "kb", "graph", "graph.mmd"))
	if err != nil {
		t.Fatalf("read graph.mmd: %v", err)
	}
	if !strings.HasPrefix(string(mermaid), "graph LR") {
		t.Fatalf("unexpected mermaid output: %s", string(mermaid))
	}
}
