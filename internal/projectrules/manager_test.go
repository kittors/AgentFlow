package projectrules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestManagerInstallWritesExpectedFiles(t *testing.T) {
	root := t.TempDir()

	manager := NewManager()
	written, err := manager.Install(root, []string{"codex", "cursor", "antigravity"}, InstallOptions{Profile: "lite"})
	if err != nil {
		t.Fatalf("install: %v", err)
	}
	if len(written) == 0 {
		t.Fatalf("expected files written")
	}

	for _, path := range []string{
		filepath.Join(root, "AGENTS.md"),
		filepath.Join(root, ".cursor", "rules", "agentflow.mdc"),
		filepath.Join(root, ".agents", "skills", "agentflow", "SKILL.md"),
	} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if !strings.Contains(string(data), "AGENTFLOW_ROUTER:") {
			t.Fatalf("expected AgentFlow marker in %s", path)
		}
	}
}

func TestManagerInstallBacksUpExistingUserFile(t *testing.T) {
	root := t.TempDir()
	targetPath := filepath.Join(root, ".windsurfrules")
	if err := os.WriteFile(targetPath, []byte("user-rules"), 0o644); err != nil {
		t.Fatalf("write seed: %v", err)
	}

	manager := NewManager()
	if _, err := manager.Install(root, []string{"windsurf"}, InstallOptions{Profile: "lite"}); err != nil {
		t.Fatalf("install: %v", err)
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	foundBackup := false
	for _, entry := range entries {
		name := entry.Name()
		if strings.Contains(name, "_bak") && strings.HasSuffix(name, ".windsurfrules") {
			foundBackup = true
			break
		}
	}
	if !foundBackup {
		t.Fatalf("expected backup file for .windsurfrules")
	}
}
