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

	if err := installer.Install("codex", "full"); err != nil {
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

func TestInstallAndUninstallGemini(t *testing.T) {
	homeDir := t.TempDir()
	installer := New(i18n.NewCatalog(), &bytes.Buffer{})
	installer.HomeDir = homeDir

	if err := installer.Install("gemini", "lite"); err != nil {
		t.Fatalf("Install returned error: %v", err)
	}

	geminiDir := filepath.Join(homeDir, ".gemini")
	rulesFile := filepath.Join(geminiDir, "GEMINI.md")
	content, err := os.ReadFile(rulesFile)
	if err != nil {
		t.Fatalf("expected rules file to exist: %v", err)
	}
	if !strings.Contains(string(content), "AGENTFLOW_ROUTER:") {
		t.Fatal("expected marker in rules file")
	}
	if !strings.Contains(string(content), "Gemini CLI") {
		t.Fatal("expected target display name in rules file")
	}
	if _, err := os.Stat(filepath.Join(geminiDir, "agentflow")); err != nil {
		t.Fatalf("expected plugin dir: %v", err)
	}
	if _, err := os.Stat(filepath.Join(geminiDir, "skills", "agentflow", "SKILL.md")); err != nil {
		t.Fatalf("expected skill file: %v", err)
	}

	if err := installer.Uninstall("gemini"); err != nil {
		t.Fatalf("Uninstall returned error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(geminiDir, "agentflow")); !os.IsNotExist(err) {
		t.Fatalf("expected plugin dir removed, got err=%v", err)
	}
	if _, err := os.Stat(rulesFile); !os.IsNotExist(err) {
		t.Fatalf("expected rules file removed, got err=%v", err)
	}
}

func TestInstallAndUninstallQwen(t *testing.T) {
	homeDir := t.TempDir()
	installer := New(i18n.NewCatalog(), &bytes.Buffer{})
	installer.HomeDir = homeDir

	if err := installer.Install("qwen", "standard"); err != nil {
		t.Fatalf("Install returned error: %v", err)
	}

	qwenDir := filepath.Join(homeDir, ".qwen")
	rulesFile := filepath.Join(qwenDir, "QWEN.md")
	content, err := os.ReadFile(rulesFile)
	if err != nil {
		t.Fatalf("expected rules file to exist: %v", err)
	}
	if !strings.Contains(string(content), "AGENTFLOW_ROUTER:") {
		t.Fatal("expected marker in rules file")
	}
	if !strings.Contains(string(content), "Qwen CLI") {
		t.Fatal("expected target display name in rules file")
	}
	if !strings.Contains(string(content), "PROFILE:standard") {
		t.Fatal("expected standard profile marker")
	}
	if _, err := os.Stat(filepath.Join(qwenDir, "agentflow")); err != nil {
		t.Fatalf("expected plugin dir: %v", err)
	}
	if _, err := os.Stat(filepath.Join(qwenDir, "skills", "agentflow", "SKILL.md")); err != nil {
		t.Fatalf("expected skill file: %v", err)
	}

	if err := installer.Uninstall("qwen"); err != nil {
		t.Fatalf("Uninstall returned error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(qwenDir, "agentflow")); !os.IsNotExist(err) {
		t.Fatalf("expected plugin dir removed, got err=%v", err)
	}
	if _, err := os.Stat(rulesFile); !os.IsNotExist(err) {
		t.Fatalf("expected rules file removed, got err=%v", err)
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
		case "OPENAI_BASE_URL":
			foundBaseURL = true
			if f.Type != "text" {
				t.Fatalf("expected text type for Base URL, got %q", f.Type)
			}
		case "__MODEL__":
			foundModel = true
			if f.Type != "select" {
				t.Fatalf("expected select type for Model, got %q", f.Type)
			}
			if len(f.Options) == 0 {
				t.Fatal("expected model options for codex")
			}
			if f.Default != "o4-mini" {
				t.Fatalf("expected default model o4-mini, got %q", f.Default)
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
		t.Fatal("expected OPENAI_BASE_URL field")
	}
	if !foundModel {
		t.Fatal("expected __MODEL__ field")
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
		case "ANTHROPIC_MODEL":
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
		t.Fatal("expected ANTHROPIC_MODEL field")
	}
}

func TestCLIConfigFieldsGemini(t *testing.T) {
	installer := New(i18n.NewCatalog(), &bytes.Buffer{})
	fields := installer.CLIConfigFields("gemini")

	if len(fields) < 3 {
		t.Fatalf("expected at least 3 fields for gemini, got %d", len(fields))
	}

	foundAPIKey, foundBaseURL, foundModel := false, false, false
	for _, f := range fields {
		switch f.EnvVar {
		case "GEMINI_API_KEY":
			foundAPIKey = true
		case "GEMINI_API_BASE_URL":
			foundBaseURL = true
		case "GEMINI_MODEL":
			foundModel = true
			if f.Type != "select" {
				t.Fatalf("expected select type for Model, got %q", f.Type)
			}
			if f.Default != "gemini-2.5-pro" {
				t.Fatalf("expected default model gemini-2.5-pro, got %q", f.Default)
			}
		}
	}
	if !foundAPIKey {
		t.Fatal("expected GEMINI_API_KEY field")
	}
	if !foundBaseURL {
		t.Fatal("expected GEMINI_API_BASE_URL field")
	}
	if !foundModel {
		t.Fatal("expected GEMINI_MODEL field")
	}
}

func TestCLIConfigFieldsQwen(t *testing.T) {
	installer := New(i18n.NewCatalog(), &bytes.Buffer{})
	fields := installer.CLIConfigFields("qwen")

	if len(fields) < 3 {
		t.Fatalf("expected at least 3 fields for qwen, got %d", len(fields))
	}

	foundAPIKey, foundBaseURL, foundModel := false, false, false
	for _, f := range fields {
		switch f.EnvVar {
		case "DASHSCOPE_API_KEY":
			foundAPIKey = true
		case "DASHSCOPE_BASE_URL":
			foundBaseURL = true
		case "DASHSCOPE_MODEL":
			foundModel = true
			if f.Type != "select" {
				t.Fatalf("expected select type for Model, got %q", f.Type)
			}
			if f.Default != "qwen3-coder" {
				t.Fatalf("expected default model qwen3-coder, got %q", f.Default)
			}
			if len(f.Options) != 3 {
				t.Fatalf("expected 3 model options for qwen, got %d", len(f.Options))
			}
		}
	}
	if !foundAPIKey {
		t.Fatal("expected DASHSCOPE_API_KEY field")
	}
	if !foundBaseURL {
		t.Fatal("expected DASHSCOPE_BASE_URL field")
	}
	if !foundModel {
		t.Fatal("expected DASHSCOPE_MODEL field")
	}
}

func TestCLIConfigFieldsKiroReturnsEmpty(t *testing.T) {
	installer := New(i18n.NewCatalog(), &bytes.Buffer{})
	fields := installer.CLIConfigFields("kiro")

	if len(fields) != 0 {
		t.Fatalf("expected no config fields for kiro, got %d", len(fields))
	}
}

func TestWriteCodexConfigMergesExisting(t *testing.T) {
	homeDir := t.TempDir()
	installer := New(i18n.NewCatalog(), &bytes.Buffer{})
	installer.HomeDir = homeDir

	// Write initial config.
	if err := installer.WriteCodexConfig("o4-mini", "medium"); err != nil {
		t.Fatalf("WriteCodexConfig returned error: %v", err)
	}

	configPath := filepath.Join(homeDir, ".codex", "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("expected config.json to exist: %v", err)
	}

	var config map[string]any
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("expected valid JSON: %v", err)
	}
	if config["model"] != "o4-mini" {
		t.Fatalf("expected model o4-mini, got %v", config["model"])
	}
	if config["reasoning"] != "medium" {
		t.Fatalf("expected reasoning medium, got %v", config["reasoning"])
	}

	// Overwrite model only, reasoning should be preserved.
	if err := installer.WriteCodexConfig("gpt-4o", ""); err != nil {
		t.Fatalf("WriteCodexConfig returned error: %v", err)
	}

	data, err = os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("expected config.json to exist: %v", err)
	}
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("expected valid JSON: %v", err)
	}
	if config["model"] != "gpt-4o" {
		t.Fatalf("expected model gpt-4o, got %v", config["model"])
	}
	if config["reasoning"] != "medium" {
		t.Fatalf("expected reasoning to be preserved as medium, got %v", config["reasoning"])
	}
}

func TestWriteEnvConfigWritesToShellRC(t *testing.T) {
	homeDir := t.TempDir()
	installer := New(i18n.NewCatalog(), &bytes.Buffer{})
	installer.HomeDir = homeDir

	// Create a .zshrc so detectShellRC finds it.
	zshrcPath := filepath.Join(homeDir, ".zshrc")
	if err := os.WriteFile(zshrcPath, []byte("# existing config\n"), 0o644); err != nil {
		t.Fatalf("write .zshrc: %v", err)
	}

	envVars := map[string]string{
		"OPENAI_API_KEY":  "sk-test-key",
		"OPENAI_BASE_URL": "https://api.example.com/v1",
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
	if !strings.Contains(text, "OPENAI_BASE_URL") {
		t.Fatal("expected OPENAI_BASE_URL in .zshrc")
	}
	if !strings.Contains(text, "sk-test-key") {
		t.Fatal("expected API key value in .zshrc")
	}
	if !strings.Contains(text, "# existing config") {
		t.Fatal("expected existing content preserved")
	}
}
