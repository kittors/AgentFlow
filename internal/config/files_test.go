package config

import (
	"path/filepath"
	"testing"
	"time"
)

func TestBackupPath(t *testing.T) {
	now := time.Date(2026, 3, 12, 16, 55, 1, 0, time.UTC)
	got := BackupPath(filepath.Join("tmp", "CLAUDE.md"), now)
	want := filepath.Join("tmp", "CLAUDE_20260312165501_bak.md")
	if got != want {
		t.Fatalf("unexpected backup path: got %q want %q", got, want)
	}
}

func TestRenameAsidePath(t *testing.T) {
	now := time.Date(2026, 3, 12, 16, 55, 1, 0, time.UTC)
	got := RenameAsidePath(filepath.Join("tmp", "agentflow"), now)
	want := filepath.Join("tmp", "agentflow._agentflow_old_20260312165501")
	if got != want {
		t.Fatalf("unexpected rename-aside path: got %q want %q", got, want)
	}
}

func TestHasMarker(t *testing.T) {
	if !HasMarker([]byte("<!-- AGENTFLOW_ROUTER: v1.0.0 -->")) {
		t.Fatal("expected marker to be detected")
	}
}
