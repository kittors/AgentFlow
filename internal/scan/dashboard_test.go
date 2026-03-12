package scan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kittors/AgentFlow/internal/kb"
)

func TestGenerateDashboardContainsExpectedLabels(t *testing.T) {
	root := t.TempDir()
	moduleDir := filepath.Join(root, filepath.Base(root), "feature")
	if err := os.MkdirAll(moduleDir, 0o755); err != nil {
		t.Fatalf("mkdir module dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(moduleDir, "main.py"), []byte("print('hello')\n"), 0o644); err != nil {
		t.Fatalf("write source file: %v", err)
	}
	if _, err := kb.CreateSession(root, kb.SessionInput{Tasks: []string{"demo"}}); err != nil {
		t.Fatalf("CreateSession returned error: %v", err)
	}

	filename, err := GenerateDashboard(root, nil)
	if err != nil {
		t.Fatalf("GenerateDashboard returned error: %v", err)
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("read dashboard: %v", err)
	}
	output := string(content)
	for _, token := range []string{"<!DOCTYPE html>", "AgentFlow Dashboard", "Modules", "Source Files", "Sessions", "KB Status"} {
		if !strings.Contains(output, token) {
			t.Fatalf("expected token %q in dashboard", token)
		}
	}
}
