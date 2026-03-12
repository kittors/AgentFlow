package kb

import (
	"strings"
	"testing"
)

func TestCreateAndExportSessions(t *testing.T) {
	root := t.TempDir()
	if _, err := CreateSession(root, SessionInput{
		Tasks:         []string{"task-1"},
		Decisions:     []string{"decision-1"},
		FilesModified: []string{"file.txt"},
	}); err != nil {
		t.Fatalf("CreateSession returned error: %v", err)
	}

	latest, err := LoadLatestSession(root)
	if err != nil {
		t.Fatalf("LoadLatestSession returned error: %v", err)
	}
	if latest["id"] == "" || !strings.Contains(latest["content"], "## Tasks") {
		t.Fatalf("unexpected latest session payload: %#v", latest)
	}

	jsonOutput, err := ExportSessions(root, "json")
	if err != nil {
		t.Fatalf("ExportSessions(json) returned error: %v", err)
	}
	if !strings.Contains(jsonOutput, "\"id\"") {
		t.Fatalf("expected json export, got %s", jsonOutput)
	}

	textOutput, err := ExportSessions(root, "text")
	if err != nil {
		t.Fatalf("ExportSessions(text) returned error: %v", err)
	}
	if !strings.Contains(textOutput, "[") {
		t.Fatalf("expected text export, got %s", textOutput)
	}
}
