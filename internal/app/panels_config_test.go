package app

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/kittors/AgentFlow/internal/projectrules"
)

func TestUninstallProjectTargetOptionsOnlyShowsInstalledTarget(t *testing.T) {
	root := t.TempDir()
	manager := projectrules.NewManager()
	if _, err := manager.Install(root, []string{"codex"}, projectrules.InstallOptions{Profile: "lite"}); err != nil {
		t.Fatalf("install project rules: %v", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() {
		_ = os.Chdir(wd)
	}()
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	application := New(&bytes.Buffer{}, &bytes.Buffer{})
	options := application.uninstallProjectTargetOptions()
	if len(options) != 1 {
		t.Fatalf("expected exactly one uninstall option, got %#v", options)
	}
	if options[0].Value != "project:codex" {
		t.Fatalf("expected only codex uninstall option, got %#v", options[0])
	}
	if filepath.ToSlash(options[0].Description) == "" {
		t.Fatal("expected uninstall option description to be populated")
	}
}
