package app

import (
	"bytes"
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

func TestBuildClaudeEnvVarsUsesApiKeyForGatewayMode(t *testing.T) {
	application := New(&bytes.Buffer{}, &bytes.Buffer{})

	envVars := application.buildClaudeEnvVars("sk-gateway-token", "https://proxy.example.com/v1")
	if envVars["ANTHROPIC_API_KEY"] != "sk-gateway-token" {
		t.Fatalf("expected gateway API key, got %#v", envVars)
	}
	if envVars["ANTHROPIC_BASE_URL"] != "https://proxy.example.com/v1" {
		t.Fatalf("expected gateway base url, got %#v", envVars)
	}
	if value, ok := envVars["ANTHROPIC_AUTH_TOKEN"]; !ok || value != "" {
		t.Fatalf("expected auth token to be cleared in gateway mode, got %#v", envVars)
	}
}

func TestBuildClaudeEnvVarsUsesApiKeyForDirectMode(t *testing.T) {
	application := New(&bytes.Buffer{}, &bytes.Buffer{})

	envVars := application.buildClaudeEnvVars("sk-direct-key", "")
	if envVars["ANTHROPIC_API_KEY"] != "sk-direct-key" {
		t.Fatalf("expected direct API key, got %#v", envVars)
	}
	if value, ok := envVars["ANTHROPIC_AUTH_TOKEN"]; !ok || value != "" {
		t.Fatalf("expected auth token to be cleared in direct mode, got %#v", envVars)
	}
}

func TestWriteEnvConfigPanelWritesClaudeApiKeyForGateway(t *testing.T) {
	homeDir := t.TempDir()
	zshrcPath := filepath.Join(homeDir, ".zshrc")
	if err := os.WriteFile(zshrcPath, []byte("# AgentFlow CLI configuration\nexport ANTHROPIC_API_KEY=\"old-key\"\nexport ANTHROPIC_BASE_URL=\"https://old.example.com\"\n"), 0o644); err != nil {
		t.Fatalf("write .zshrc: %v", err)
	}

	application := New(&bytes.Buffer{}, &bytes.Buffer{})
	application.Installer.HomeDir = homeDir

	panel := application.writeEnvConfigPanel(map[string]string{
		"ANTHROPIC_API_KEY":  "sk-gateway-token",
		"ANTHROPIC_BASE_URL": "https://proxy.example.com/v1",
		"__CLAUDE_MODEL__":   "claude-sonnet-4-6",
	})
	if strings.Contains(strings.ToLower(panel.Title), "失败") || strings.Contains(strings.ToLower(panel.Title), "failed") {
		t.Fatalf("expected successful panel, got %#v", panel)
	}

	content, err := os.ReadFile(zshrcPath)
	if err != nil {
		t.Fatalf("read .zshrc: %v", err)
	}
	text := string(content)
	if !strings.Contains(text, "export ANTHROPIC_API_KEY=\"sk-gateway-token\"") {
		t.Fatalf("expected API key export in gateway mode, got %q", text)
	}
	if !strings.Contains(text, "export ANTHROPIC_BASE_URL=\"https://proxy.example.com/v1\"") {
		t.Fatalf("expected gateway base url export, got %q", text)
	}
	if strings.Contains(text, "ANTHROPIC_AUTH_TOKEN") {
		t.Fatalf("expected no ANTHROPIC_AUTH_TOKEN in gateway mode, got %q", text)
	}
}
