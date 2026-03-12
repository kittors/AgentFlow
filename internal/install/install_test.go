package install

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kittors/AgentFlow/internal/i18n"
)

func TestBuildRulesContentUsesProfileAndTargetPlaceholders(t *testing.T) {
	installer := New(i18n.NewCatalog(), &bytes.Buffer{})

	content, err := installer.buildRulesContent("codex", "standard")
	if err != nil {
		t.Fatalf("buildRulesContent returned error: %v", err)
	}

	if strings.Contains(content, "{TARGET_CLI}") || strings.Contains(content, "{HOOKS_SUMMARY}") {
		t.Fatal("expected placeholders to be replaced")
	}
	if !strings.Contains(content, "Codex CLI") {
		t.Fatal("expected target display name in rendered rules")
	}
	if strings.Contains(content, "{CLI_SUBAGENT_PROTOCOL}") || strings.Contains(content, "{HOOKS_MATRIX}") {
		t.Fatal("expected core placeholders to be replaced")
	}
	if !strings.Contains(content, "PROFILE:standard") {
		t.Fatal("expected standard profile marker")
	}
	if !strings.Contains(content, "G6 | 通用规则") {
		t.Fatal("expected standard profile modules to be appended")
	}
}

func TestInstallAndUninstallCodex(t *testing.T) {
	homeDir := t.TempDir()
	codexDir := filepath.Join(homeDir, ".codex")
	if err := os.MkdirAll(codexDir, 0o755); err != nil {
		t.Fatalf("mkdir codex dir: %v", err)
	}
	rulesFile := filepath.Join(codexDir, "AGENTS.md")
	if err := os.WriteFile(rulesFile, []byte("# custom\n"), 0o644); err != nil {
		t.Fatalf("write seed rules: %v", err)
	}
	configFile := filepath.Join(codexDir, "config.toml")
	configSeed := "[features]\ntelemetry = false\n"
	if err := os.WriteFile(configFile, []byte(configSeed), 0o644); err != nil {
		t.Fatalf("write seed config: %v", err)
	}

	stdout := &bytes.Buffer{}
	installer := New(i18n.NewCatalog(), stdout)
	installer.HomeDir = homeDir

	if err := installer.Install("codex", "full"); err != nil {
		t.Fatalf("Install returned error: %v", err)
	}

	agentsContent, err := os.ReadFile(rulesFile)
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	if !strings.Contains(string(agentsContent), "PROFILE:full") {
		t.Fatal("expected full profile extension in AGENTS.md")
	}
	if strings.Contains(string(agentsContent), "{HOOKS_SUMMARY}") {
		t.Fatal("expected rules placeholders to be rendered")
	}
	if _, err := os.Stat(filepath.Join(codexDir, "agentflow")); err != nil {
		t.Fatalf("expected plugin dir: %v", err)
	}
	if _, err := os.Stat(filepath.Join(codexDir, "skills", "agentflow", "SKILL.md")); err != nil {
		t.Fatalf("expected skill file: %v", err)
	}
	for _, roleFile := range []string{"reviewer.toml", "architect.toml"} {
		rolePath := filepath.Join(codexDir, "agents", roleFile)
		content, err := os.ReadFile(rolePath)
		if err != nil {
			t.Fatalf("expected role file %s: %v", roleFile, err)
		}
		if !strings.Contains(string(content), "AGENTFLOW_ROUTER:") {
			t.Fatalf("expected marker in %s", roleFile)
		}
	}

	configContent, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("read config.toml: %v", err)
	}
	configText := string(configContent)
	if !strings.Contains(configText, "multi_agent = true") {
		t.Fatal("expected multi_agent to be enabled")
	}
	if !strings.Contains(configText, "[agents.reviewer]") || !strings.Contains(configText, "[agents.architect]") {
		t.Fatal("expected codex agent sections in config.toml")
	}

	entries, err := os.ReadDir(codexDir)
	if err != nil {
		t.Fatalf("ReadDir returned error: %v", err)
	}
	foundBackup := false
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "AGENTS_") && strings.HasSuffix(entry.Name(), "_bak.md") {
			foundBackup = true
			break
		}
	}
	if !foundBackup {
		t.Fatal("expected backup rules file")
	}

	if err := installer.Uninstall("codex"); err != nil {
		t.Fatalf("Uninstall returned error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(codexDir, "agentflow")); !os.IsNotExist(err) {
		t.Fatalf("expected plugin dir removed, got err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(codexDir, "agents", "reviewer.toml")); !os.IsNotExist(err) {
		t.Fatalf("expected reviewer role removed, got err=%v", err)
	}

	configContent, err = os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("read cleaned config.toml: %v", err)
	}
	configText = string(configContent)
	if strings.Contains(configText, "multi_agent = true") {
		t.Fatal("expected multi_agent flag to be removed on uninstall")
	}
	if strings.Contains(configText, "[agents.reviewer]") || strings.Contains(configText, "[agents.architect]") {
		t.Fatal("expected codex agent sections to be removed on uninstall")
	}
	if !strings.Contains(configText, "telemetry = false") {
		t.Fatal("expected user config to be preserved")
	}
}

func TestInstallAndUninstallClaudeHooks(t *testing.T) {
	homeDir := t.TempDir()
	claudeDir := filepath.Join(homeDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatalf("mkdir claude dir: %v", err)
	}

	settingsPath := filepath.Join(claudeDir, "settings.json")
	seed := map[string]any{
		"hooks": map[string]any{
			"PreToolUse": []any{
				map[string]any{
					"matcher": "",
					"hooks": []any{
						map[string]any{
							"type":        "command",
							"command":     "echo keep-me",
							"description": "custom-pretool",
						},
						map[string]any{
							"type":        "command",
							"command":     "echo remove-me",
							"description": "agentflow-old-hook",
						},
					},
				},
			},
		},
	}
	rawSeed, _ := json.MarshalIndent(seed, "", "  ")
	if err := os.WriteFile(settingsPath, rawSeed, 0o644); err != nil {
		t.Fatalf("write settings seed: %v", err)
	}

	installer := New(i18n.NewCatalog(), &bytes.Buffer{})
	installer.HomeDir = homeDir

	if err := installer.Install("claude", "full"); err != nil {
		t.Fatalf("Install returned error: %v", err)
	}

	content, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read settings.json: %v", err)
	}
	text := string(content)
	if !strings.Contains(text, "custom-pretool") {
		t.Fatal("expected existing custom hook to be preserved")
	}
	if !strings.Contains(text, "agentflow-kb-sync") || !strings.Contains(text, "agentflow-session-save") {
		t.Fatal("expected AgentFlow Claude hooks to be merged")
	}
	if strings.Contains(text, "agentflow-old-hook") {
		t.Fatal("expected previous AgentFlow hook entries to be replaced")
	}

	if err := installer.Uninstall("claude"); err != nil {
		t.Fatalf("Uninstall returned error: %v", err)
	}

	content, err = os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read cleaned settings.json: %v", err)
	}
	text = string(content)
	if !strings.Contains(text, "custom-pretool") {
		t.Fatal("expected custom hook to remain after uninstall")
	}
	if strings.Contains(text, "agentflow-kb-sync") || strings.Contains(text, "agentflow-session-save") {
		t.Fatal("expected AgentFlow Claude hooks to be removed on uninstall")
	}
}
