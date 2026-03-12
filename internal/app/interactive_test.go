package app

import (
	"bytes"
	"testing"

	"github.com/kittors/AgentFlow/internal/i18n"
)

func TestEnsureInteractiveLanguageUsesSavedPreference(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("USERPROFILE", homeDir)

	if err := i18n.SavePreferredLocale("zh"); err != nil {
		t.Fatalf("SavePreferredLocale returned error: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	application := New(stdout, stderr)

	code, ok := application.ensureInteractiveLanguage()
	if !ok || code != 0 {
		t.Fatalf("expected onboarding to succeed, got code=%d ok=%v stderr=%q", code, ok, stderr.String())
	}
	if application.Catalog.Language() != "zh" {
		t.Fatalf("expected catalog language zh, got %q", application.Catalog.Language())
	}
	if application.Installer.Catalog.Language() != "zh" {
		t.Fatalf("expected installer catalog language zh, got %q", application.Installer.Catalog.Language())
	}
}
