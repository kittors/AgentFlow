package install

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kittors/AgentFlow/internal/config"
	"github.com/kittors/AgentFlow/internal/i18n"
)

func TestBuildRulesContentUsesProfileAndTargetPlaceholders(t *testing.T) {
	installer := New(i18n.NewCatalog(), &bytes.Buffer{})

	content, err := installer.buildRulesContent("codex", "standard", config.DefaultLang)
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

func TestBuildEntryRulesContentRewritesModuleLinksForCLIEntry(t *testing.T) {
	installer := New(i18n.NewCatalog(), &bytes.Buffer{})

	content, err := installer.buildEntryRulesContent("codex", "standard", config.DefaultLang, "../.agentflow/agentflow")
	if err != nil {
		t.Fatalf("buildEntryRulesContent returned error: %v", err)
	}

	if !strings.Contains(content, "](../.agentflow/agentflow/core/common.md)") {
		t.Fatalf("expected rewritten CLI module link, got %q", content)
	}
	if strings.Contains(content, "](agentflow/core/common.md)") {
		t.Fatal("expected original embedded module link to be rewritten for CLI entry")
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

	if err := installer.Install("codex", "full", config.DefaultLang); err != nil {
		t.Fatalf("Install returned error: %v", err)
	}

	// CLI AGENTS.md should now contain the full compiled AgentFlow rules.
	agentsContent, err := os.ReadFile(rulesFile)
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	agentsStr := string(agentsContent)
	if !strings.Contains(agentsStr, "AGENTFLOW_ROUTER:") {
		t.Fatal("expected AGENTFLOW_ROUTER marker in CLI AGENTS.md")
	}
	if strings.Contains(agentsStr, ".agentflow/AGENTS.md") {
		t.Fatal("expected CLI AGENTS.md to embed full rules instead of a .agentflow reference")
	}
	if !strings.Contains(agentsStr, "先路由再行动") {
		t.Fatal("expected CLI AGENTS.md to contain core AgentFlow rules")
	}
	if !strings.Contains(agentsStr, "](../.agentflow/agentflow/core/common.md)") {
		t.Fatal("expected CLI AGENTS.md links to point to ../.agentflow/agentflow/")
	}

	// Full rules should be in ~/.agentflow/AGENTS.md (global dir).
	globalRulesFile := filepath.Join(homeDir, ".agentflow", "AGENTS.md")
	globalContent, err := os.ReadFile(globalRulesFile)
	if err != nil {
		t.Fatalf("read global rules: %v", err)
	}
	if !strings.Contains(string(globalContent), "PROFILE:full") {
		t.Fatal("expected full profile extension in global AGENTS.md")
	}
	if strings.Contains(string(globalContent), "{HOOKS_SUMMARY}") {
		t.Fatal("expected rules placeholders to be rendered")
	}

	// Module dir should be in ~/.agentflow/agentflow/ (not CLI dir).
	globalModuleDir := filepath.Join(homeDir, ".agentflow", "agentflow")
	if _, err := os.Stat(globalModuleDir); err != nil {
		t.Fatalf("expected global module dir: %v", err)
	}
	// CLI-local agentflow dir should NOT exist (modules are centralized).
	if _, err := os.Stat(filepath.Join(codexDir, "agentflow")); !os.IsNotExist(err) {
		t.Fatal("expected CLI-local agentflow dir to NOT exist")
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
	if _, err := os.Stat(filepath.Join(codexDir, "agents", "reviewer.toml")); !os.IsNotExist(err) {
		t.Fatalf("expected reviewer role removed, got err=%v", err)
	}
	// Global module dir should still exist after single CLI uninstall.
	if _, err := os.Stat(globalModuleDir); err != nil {
		t.Fatalf("expected global module dir to persist after single uninstall: %v", err)
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

	if err := installer.Install("claude", "full", config.DefaultLang); err != nil {
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

func TestInstallClaudeWritesFullRulesEntryFile(t *testing.T) {
	homeDir := t.TempDir()
	claudeDir := filepath.Join(homeDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatalf("mkdir claude dir: %v", err)
	}

	installer := New(i18n.NewCatalog(), &bytes.Buffer{})
	installer.HomeDir = homeDir

	if err := installer.Install("claude", "standard", config.DefaultLang); err != nil {
		t.Fatalf("Install returned error: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(claudeDir, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	text := string(content)
	if strings.Contains(text, ".agentflow/AGENTS.md") {
		t.Fatal("expected CLAUDE.md to embed full rules instead of a .agentflow reference")
	}
	if !strings.Contains(text, "Claude Code") {
		t.Fatal("expected target-specific placeholders to be rendered in CLAUDE.md")
	}
	if !strings.Contains(text, "先路由再行动") {
		t.Fatal("expected CLAUDE.md to include core AgentFlow rules")
	}
}

func TestUninstallCodexPreservesUserManagedMultiAgentFlagAndReviewerSection(t *testing.T) {
	homeDir := t.TempDir()
	codexDir := filepath.Join(homeDir, ".codex")
	if err := os.MkdirAll(codexDir, 0o755); err != nil {
		t.Fatalf("mkdir codex dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(codexDir, "AGENTS.md"), []byte("<!-- AGENTFLOW_ROUTER: v1.0.0 -->\n"), 0o644); err != nil {
		t.Fatalf("write AGENTS.md: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(codexDir, "agentflow"), 0o755); err != nil {
		t.Fatalf("mkdir agentflow dir: %v", err)
	}

	configPath := filepath.Join(codexDir, "config.toml")
	seed := "[features]\nmulti_agent = true\ntelemetry = false\n\n[agents.reviewer]\ndescription = \"custom reviewer\"\nconfig_file = \"agents/custom-reviewer.toml\"\n"
	if err := os.WriteFile(configPath, []byte(seed), 0o644); err != nil {
		t.Fatalf("write config.toml: %v", err)
	}

	installer := New(i18n.NewCatalog(), &bytes.Buffer{})
	installer.HomeDir = homeDir

	if err := installer.Uninstall("codex"); err != nil {
		t.Fatalf("Uninstall returned error: %v", err)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config.toml: %v", err)
	}
	text := string(content)
	if !strings.Contains(text, "multi_agent = true") {
		t.Fatalf("expected existing multi_agent flag to be preserved, got %q", text)
	}
	if !strings.Contains(text, "custom reviewer") {
		t.Fatalf("expected existing reviewer section to be preserved, got %q", text)
	}
}

func TestInstallCodexBacksUpUserManagedReviewerRole(t *testing.T) {
	homeDir := t.TempDir()
	codexDir := filepath.Join(homeDir, ".codex")
	roleDir := filepath.Join(codexDir, "agents")
	if err := os.MkdirAll(roleDir, 0o755); err != nil {
		t.Fatalf("mkdir agents dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(codexDir, "AGENTS.md"), []byte("# custom\n"), 0o644); err != nil {
		t.Fatalf("write AGENTS.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(codexDir, "config.toml"), []byte("[features]\n"), 0o644); err != nil {
		t.Fatalf("write config.toml: %v", err)
	}

	reviewerPath := filepath.Join(roleDir, "reviewer.toml")
	if err := os.WriteFile(reviewerPath, []byte("description = \"my reviewer\"\n"), 0o644); err != nil {
		t.Fatalf("write reviewer.toml: %v", err)
	}

	installer := New(i18n.NewCatalog(), &bytes.Buffer{})
	installer.HomeDir = homeDir

	if err := installer.Install("codex", "full", config.DefaultLang); err != nil {
		t.Fatalf("Install returned error: %v", err)
	}

	content, err := os.ReadFile(reviewerPath)
	if err != nil {
		t.Fatalf("read reviewer.toml: %v", err)
	}
	if strings.Contains(string(content), "my reviewer") {
		t.Fatalf("expected reviewer.toml to be replaced during install, got %q", string(content))
	}

	entries, err := os.ReadDir(roleDir)
	if err != nil {
		t.Fatalf("ReadDir returned error: %v", err)
	}
	foundBackup := false
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "reviewer_") && strings.HasSuffix(entry.Name(), "_bak.toml") {
			foundBackup = true
			break
		}
	}
	if !foundBackup {
		t.Fatal("expected backup for overwritten reviewer.toml")
	}
}

func TestCLIConfigFieldsCodex(t *testing.T) {
	installer := New(i18n.NewCatalog(), &bytes.Buffer{})
	fields := installer.CLIConfigFields("codex")

	if len(fields) < 3 {
		t.Fatalf("expected at least 3 fields for codex, got %d", len(fields))
	}

	foundAPIKey, foundBaseURL, foundModel, foundReasoning := false, false, false, false
	for _, f := range fields {
		switch f.EnvVar {
		case "OPENAI_API_KEY":
			foundAPIKey = true
			if f.Type != "text" {
				t.Fatalf("expected text type for API Key, got %q", f.Type)
			}
		case "__CODEX_BASE_URL__":
			foundBaseURL = true
			if f.Type != "text" {
				t.Fatalf("expected text type for Base URL, got %q", f.Type)
			}
		case "__CODEX_MODEL__":
			foundModel = true
			if f.Type != "select" {
				t.Fatalf("expected select type for Model, got %q", f.Type)
			}
			if len(f.Options) == 0 {
				t.Fatal("expected model options for codex")
			}
			if f.Default != "gpt-5.2" {
				t.Fatalf("expected default model gpt-5.2, got %q", f.Default)
			}
		case "__CODEX_REASONING__":
			foundReasoning = true
			if f.Type != "select" {
				t.Fatalf("expected select type for Reasoning, got %q", f.Type)
			}
			if f.Default != "medium" {
				t.Fatalf("expected default reasoning medium, got %q", f.Default)
			}
		}
	}
	if !foundAPIKey {
		t.Fatal("expected OPENAI_API_KEY field")
	}
	if !foundBaseURL {
		t.Fatal("expected __CODEX_BASE_URL__ field")
	}
	if !foundModel {
		t.Fatal("expected __CODEX_MODEL__ field")
	}
	if !foundReasoning {
		t.Fatal("expected __CODEX_REASONING__ field")
	}
}

func TestCLIConfigFieldsClaude(t *testing.T) {
	installer := New(i18n.NewCatalog(), &bytes.Buffer{})
	fields := installer.CLIConfigFields("claude")

	if len(fields) < 3 {
		t.Fatalf("expected at least 3 fields for claude, got %d", len(fields))
	}

	foundAPIKey, foundBaseURL, foundModel := false, false, false
	for _, f := range fields {
		switch f.EnvVar {
		case "ANTHROPIC_API_KEY":
			foundAPIKey = true
		case "ANTHROPIC_BASE_URL":
			foundBaseURL = true
		case "__CLAUDE_MODEL__":
			foundModel = true
			if f.Type != "select" {
				t.Fatalf("expected select type for Model, got %q", f.Type)
			}
			if len(f.Options) == 0 {
				t.Fatal("expected model options for claude")
			}
		}
	}
	if !foundAPIKey {
		t.Fatal("expected ANTHROPIC_API_KEY field")
	}
	if !foundBaseURL {
		t.Fatal("expected ANTHROPIC_BASE_URL field")
	}
	if !foundModel {
		t.Fatal("expected __CLAUDE_MODEL__ field")
	}
}

func TestWriteCodexConfigMergesExisting(t *testing.T) {
	homeDir := t.TempDir()
	installer := New(i18n.NewCatalog(), &bytes.Buffer{})
	installer.HomeDir = homeDir

	// Write initial config with API key, base URL, model, and reasoning.
	if err := installer.WriteCodexConfig("sk-test-key", "https://api.example.com/v1", "o4-mini", "medium"); err != nil {
		t.Fatalf("WriteCodexConfig returned error: %v", err)
	}

	// Check auth.json.
	authPath := filepath.Join(homeDir, ".codex", "auth.json")
	authData, err := os.ReadFile(authPath)
	if err != nil {
		t.Fatalf("expected auth.json to exist: %v", err)
	}
	if !strings.Contains(string(authData), "sk-test-key") {
		t.Fatalf("expected API key in auth.json, got %s", string(authData))
	}

	// Check config.toml.
	configPath := filepath.Join(homeDir, ".codex", "config.toml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("expected config.toml to exist: %v", err)
	}

	text := string(data)
	if !strings.Contains(text, `model = "o4-mini"`) {
		t.Fatalf("expected model o4-mini, got %v", text)
	}
	if !strings.Contains(text, `model_reasoning_effort = "medium"`) {
		t.Fatalf("expected reasoning medium, got %v", text)
	}
	if !strings.Contains(text, `model_provider = "agentflow"`) {
		t.Fatalf("expected model_provider, got %v", text)
	}
	if !strings.Contains(text, `base_url = "https://api.example.com/v1"`) {
		t.Fatalf("expected base_url in model_providers section, got %v", text)
	}
	if strings.Contains(text, `openai_base_url`) {
		t.Fatalf("expected no legacy openai_base_url, got %v", text)
	}
	if !strings.Contains(text, `env_key = "OPENAI_API_KEY"`) {
		t.Fatalf("expected env_key in model_providers section, got %v", text)
	}

	// Overwrite model only, other settings should be preserved.
	if err := installer.WriteCodexConfig("", "", "gpt-4o", ""); err != nil {
		t.Fatalf("WriteCodexConfig returned error: %v", err)
	}

	data, err = os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("expected config.toml to exist: %v", err)
	}
	text = string(data)
	if !strings.Contains(text, `model = "gpt-4o"`) {
		t.Fatalf("expected model gpt-4o, got %v", text)
	}
	if !strings.Contains(text, `model_reasoning_effort = "medium"`) {
		t.Fatalf("expected reasoning to be preserved as medium, got %v", text)
	}
	if !strings.Contains(text, `model_provider = "agentflow"`) {
		t.Fatalf("expected model_provider to be preserved, got %v", text)
	}
	if !strings.Contains(text, `base_url = "https://api.example.com/v1"`) {
		t.Fatalf("expected base_url to be preserved, got %v", text)
	}
}

func TestWriteEnvConfigWritesToShellRC(t *testing.T) {
	// Override runtimeGOOS so that on Windows CI this test still exercises
	// the RC-file code path rather than the setx code path.
	old := runtimeGOOS
	runtimeGOOS = "linux"
	defer func() { runtimeGOOS = old }()

	homeDir := t.TempDir()
	installer := New(i18n.NewCatalog(), &bytes.Buffer{})
	installer.HomeDir = homeDir

	// Create a .zshrc so detectShellRC finds it.
	zshrcPath := filepath.Join(homeDir, ".zshrc")
	if err := os.WriteFile(zshrcPath, []byte("# existing config\n"), 0o644); err != nil {
		t.Fatalf("write .zshrc: %v", err)
	}

	envVars := map[string]string{
		"OPENAI_API_KEY":     "sk-test-key",
		"ANTHROPIC_BASE_URL": "https://api.example.com/v1",
	}

	lines, err := installer.WriteEnvConfig(envVars)
	if err != nil {
		t.Fatalf("WriteEnvConfig returned error: %v", err)
	}

	if len(lines) == 0 {
		t.Fatal("expected output lines")
	}

	content, err := os.ReadFile(zshrcPath)
	if err != nil {
		t.Fatalf("read .zshrc: %v", err)
	}
	text := string(content)
	if !strings.Contains(text, "OPENAI_API_KEY") {
		t.Fatal("expected OPENAI_API_KEY in .zshrc")
	}
	if !strings.Contains(text, "ANTHROPIC_BASE_URL") {
		t.Fatal("expected ANTHROPIC_BASE_URL in .zshrc")
	}
	if !strings.Contains(text, "sk-test-key") {
		t.Fatal("expected API key value in .zshrc")
	}
	if !strings.Contains(text, "# existing config") {
		t.Fatal("expected existing content preserved")
	}
}
