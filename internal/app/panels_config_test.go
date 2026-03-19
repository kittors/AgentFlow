package app

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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

func TestWriteEnvConfigPanelWritesClaudeSettingsJSON(t *testing.T) {
	homeDir := t.TempDir()
	claudeDir := filepath.Join(homeDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatalf("mkdir .claude: %v", err)
	}

	application := New(&bytes.Buffer{}, &bytes.Buffer{})
	application.Installer.HomeDir = homeDir

	panel := application.writeEnvConfigPanel(map[string]string{
		"__CLAUDE_API_KEY__":  "sk-gateway-token",
		"__CLAUDE_BASE_URL__": "https://proxy.example.com/v1",
		"__CLAUDE_MODEL__":    "claude-sonnet-4-6",
	})
	if strings.Contains(strings.ToLower(panel.Title), "失败") || strings.Contains(strings.ToLower(panel.Title), "failed") {
		t.Fatalf("expected successful panel, got %#v", panel)
	}

	// Verify settings.json was written.
	settingsPath := filepath.Join(claudeDir, "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read settings.json: %v", err)
	}
	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("parse settings.json: %v", err)
	}

	// Check env section.
	envMap, ok := settings["env"].(map[string]any)
	if !ok {
		t.Fatalf("expected env section in settings.json, got %v", settings)
	}
	if envMap["ANTHROPIC_API_KEY"] != "sk-gateway-token" {
		t.Fatalf("expected API key in env, got %v", envMap["ANTHROPIC_API_KEY"])
	}
	if envMap["ANTHROPIC_BASE_URL"] != "https://proxy.example.com/v1" {
		t.Fatalf("expected base URL in env, got %v", envMap["ANTHROPIC_BASE_URL"])
	}
	if envMap["CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC"] != "1" {
		t.Fatal("expected CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1 when base URL is set")
	}

	// Check model.
	if settings["model"] != "claude-sonnet-4-6" {
		t.Fatalf("expected model in settings.json, got %v", settings["model"])
	}
}
