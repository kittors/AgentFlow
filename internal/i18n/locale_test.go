package i18n

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSavePreferredLocalePersistsLanguage(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("USERPROFILE", homeDir)
	t.Setenv("AGENTFLOW_LANG", "")
	t.Setenv("LC_ALL", "")
	t.Setenv("LC_MESSAGES", "")
	t.Setenv("LANG", "")
	t.Setenv("LANGUAGE", "")

	if err := SavePreferredLocale("zh"); err != nil {
		t.Fatalf("SavePreferredLocale returned error: %v", err)
	}

	path := filepath.Join(homeDir, ".agentflow", "preferences", "locale")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if string(data) != "zh\n" {
		t.Fatalf("expected locale file to contain zh, got %q", string(data))
	}

	locale, ok := LoadPreferredLocale()
	if !ok {
		t.Fatal("expected preferred locale to be available")
	}
	if locale != "zh" {
		t.Fatalf("expected zh locale, got %q", locale)
	}
	if DetectLocale() != "zh" {
		t.Fatalf("expected DetectLocale to use saved preference, got %q", DetectLocale())
	}
}

func TestDetectLocaleFallsBackToEnvironment(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("USERPROFILE", homeDir)
	t.Setenv("AGENTFLOW_LANG", "en")
	t.Setenv("LC_ALL", "")
	t.Setenv("LC_MESSAGES", "")
	t.Setenv("LANG", "")
	t.Setenv("LANGUAGE", "")

	if locale := DetectLocale(); locale != "en" {
		t.Fatalf("expected DetectLocale to use AGENTFLOW_LANG, got %q", locale)
	}
}
