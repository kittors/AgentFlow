package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunKBSyncCreatesModuleIndex(t *testing.T) {
	root := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd returned error: %v", err)
	}
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(root); err != nil {
		t.Fatalf("Chdir returned error: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".agentflow"), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}

	moduleDir := filepath.Join(root, "internal", "demo")
	if err := os.MkdirAll(moduleDir, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(moduleDir, "main.go"), []byte("package demo\n"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	application := New(stdout, stderr)

	if code := application.Run([]string{"kb", "sync", "--quiet"}); code != 0 {
		t.Fatalf("expected exit code 0, got %d stderr=%q", code, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(root, ".agentflow", "kb", "modules", "_index.md")); err != nil {
		t.Fatalf("expected module index to be created: %v", err)
	}
}

func TestRunSessionSaveCreatesSummary(t *testing.T) {
	root := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd returned error: %v", err)
	}
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(root); err != nil {
		t.Fatalf("Chdir returned error: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".agentflow"), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	application := New(stdout, stderr)

	if code := application.Run([]string{"session", "save", "--quiet", "--stage=HOOK"}); code != 0 {
		t.Fatalf("expected exit code 0, got %d stderr=%q", code, stderr.String())
	}

	entries, err := os.ReadDir(filepath.Join(root, ".agentflow", "sessions"))
	if err != nil {
		t.Fatalf("ReadDir returned error: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected session file to be created")
	}
	content, err := os.ReadFile(filepath.Join(root, ".agentflow", "sessions", entries[0].Name()))
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if !strings.Contains(string(content), "Stage: HOOK") {
		t.Fatalf("expected stage to be persisted, got %q", string(content))
	}
}

func TestRunKBSyncFailsOutsideAgentFlowProject(t *testing.T) {
	root := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd returned error: %v", err)
	}
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(root); err != nil {
		t.Fatalf("Chdir returned error: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	application := New(stdout, stderr)

	if code := application.Run([]string{"kb", "sync", "--quiet"}); code == 0 {
		t.Fatalf("expected non-zero exit code outside AgentFlow project, stderr=%q", stderr.String())
	}
	if _, err := os.Stat(filepath.Join(root, ".agentflow")); !os.IsNotExist(err) {
		t.Fatalf("expected no .agentflow directory to be created, got %v", err)
	}
}

func TestRunSessionSaveFailsOutsideAgentFlowProject(t *testing.T) {
	root := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd returned error: %v", err)
	}
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(root); err != nil {
		t.Fatalf("Chdir returned error: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	application := New(stdout, stderr)

	if code := application.Run([]string{"session", "save", "--quiet"}); code == 0 {
		t.Fatalf("expected non-zero exit code outside AgentFlow project, stderr=%q", stderr.String())
	}
	if _, err := os.Stat(filepath.Join(root, ".agentflow")); !os.IsNotExist(err) {
		t.Fatalf("expected no .agentflow directory to be created, got %v", err)
	}
}

func TestRunInitCreatesCoreKBFiles(t *testing.T) {
	root := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd returned error: %v", err)
	}
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(root); err != nil {
		t.Fatalf("Chdir returned error: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	application := New(stdout, stderr)

	if code := application.Run([]string{"init", "--quiet"}); code != 0 {
		t.Fatalf("expected exit code 0, got %d stderr=%q", code, stderr.String())
	}

	for _, name := range []string{"INDEX.md", "context.md", "CHANGELOG.md"} {
		if _, err := os.Stat(filepath.Join(root, ".agentflow", "kb", name)); err != nil {
			t.Fatalf("expected %s to be created: %v", name, err)
		}
	}
}

func TestRunInitReturnsErrorOnPartialFailure(t *testing.T) {
	root := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd returned error: %v", err)
	}
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(root); err != nil {
		t.Fatalf("Chdir returned error: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".agentflow", "kb"), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".agentflow", "kb", "conventions"), []byte("blocked"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	application := New(stdout, stderr)

	if code := application.Run([]string{"init", "--quiet"}); code == 0 {
		t.Fatalf("expected init to fail on partial initialization error, stderr=%q", stderr.String())
	}
}
