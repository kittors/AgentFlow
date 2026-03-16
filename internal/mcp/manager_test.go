package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kittors/AgentFlow/internal/targets"
)

func TestManagerInstallAndRemove(t *testing.T) {
	tmp := t.TempDir()
	target, ok := targets.Lookup("claude")
	if !ok {
		t.Fatalf("expected claude target")
	}

	manager := &Manager{HomeDir: tmp}

	claudeConfigPath := filepath.Join(tmp, ".claude.json")
	seed := []byte("{\"mcpServers\":{\"Playwright\":{\"command\":\"npx\",\"args\":[\"-y\",\"@playwright/mcp@latest\"],\"env\":{}}}}\n")
	if err := os.WriteFile(claudeConfigPath, seed, 0o600); err != nil {
		t.Fatalf("write claude config seed: %v", err)
	}
	initialPerm := os.FileMode(0o0)
	if info, err := os.Stat(claudeConfigPath); err != nil {
		t.Fatalf("stat claude config: %v", err)
	} else {
		initialPerm = info.Mode().Perm()
	}

	if err := manager.Install(target, "context7", InstallOptions{Env: []string{"CONTEXT7_API_KEY=demo"}}); err != nil {
		t.Fatalf("install context7: %v", err)
	}

	data, err := os.ReadFile(claudeConfigPath)
	if err != nil {
		t.Fatalf("read claude config: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("parse claude config: %v", err)
	}

	mcpServers, ok := payload["mcpServers"].(map[string]any)
	if !ok {
		t.Fatalf("mcpServers missing or invalid")
	}
	if _, ok := mcpServers["Playwright"]; !ok {
		t.Fatalf("expected Playwright preserved")
	}
	context7, ok := mcpServers["Context7"].(map[string]any)
	if !ok {
		t.Fatalf("context7 missing")
	}
	if context7["command"] != "npx" {
		t.Fatalf("unexpected command: %v", context7["command"])
	}
	env, ok := context7["env"].(map[string]any)
	if !ok {
		t.Fatalf("env missing")
	}
	if env["CONTEXT7_API_KEY"] != "demo" {
		t.Fatalf("unexpected api key: %v", env["CONTEXT7_API_KEY"])
	}

	if err := manager.Remove(target, "context7"); err != nil {
		t.Fatalf("remove context7: %v", err)
	}
	data, err = os.ReadFile(claudeConfigPath)
	if err != nil {
		t.Fatalf("read claude config: %v", err)
	}
	payload = nil
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("parse claude config: %v", err)
	}
	mcpServers, _ = payload["mcpServers"].(map[string]any)
	if _, exists := mcpServers["Context7"]; exists {
		t.Fatalf("expected context7 removed")
	}
	if info, err := os.Stat(claudeConfigPath); err != nil {
		t.Fatalf("stat claude config: %v", err)
	} else if info.Mode().Perm() != initialPerm {
		t.Fatalf("expected claude config perms to stay %v, got %v", initialPerm, info.Mode().Perm())
	}
}

func TestManagerListClaudeMergesClaudeJSONAndSettings(t *testing.T) {
	tmp := t.TempDir()
	target, ok := targets.Lookup("claude")
	if !ok {
		t.Fatalf("expected claude target")
	}

	manager := &Manager{HomeDir: tmp}

	claudeConfigPath := filepath.Join(tmp, ".claude.json")
	if err := os.WriteFile(claudeConfigPath, []byte("{\"mcpServers\":{\"Context7\":{},\"Playwright\":{}}}\n"), 0o600); err != nil {
		t.Fatalf("write .claude.json: %v", err)
	}
	settingsPath := filepath.Join(tmp, target.Dir, "settings.json")
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		t.Fatalf("mkdir settings dir: %v", err)
	}
	if err := os.WriteFile(settingsPath, []byte("{\"mcpServers\":{\"filesystem\":{}}}\n"), 0o644); err != nil {
		t.Fatalf("write settings.json: %v", err)
	}

	managedPath := filepath.Join(tmp, target.Dir, "mcp.json")
	if err := os.WriteFile(managedPath, []byte("{\"mcpServers\":{\"context7\":{}}}\n"), 0o644); err != nil {
		t.Fatalf("write managed mcp: %v", err)
	}

	names, err := manager.List(target)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	joined := strings.Join(names, ",")
	if !strings.Contains(joined, "Context7") {
		t.Fatalf("expected Context7 present, got %v", names)
	}
	if strings.Contains(joined, "context7") {
		t.Fatalf("expected no context7 duplicate, got %v", names)
	}
	if !strings.Contains(strings.ToLower(joined), "filesystem") {
		t.Fatalf("expected filesystem present, got %v", names)
	}
}

func TestResolveBuiltinFilesystemRequiresAllow(t *testing.T) {
	if _, err := ResolveBuiltin("filesystem", InstallOptions{}); err == nil {
		t.Fatalf("expected error")
	}
}

func TestManagerListReadsCodexConfigToml(t *testing.T) {
	tmp := t.TempDir()
	target, ok := targets.Lookup("codex")
	if !ok {
		t.Fatalf("expected codex target")
	}

	configPath := filepath.Join(tmp, target.Dir, "config.toml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(configPath, []byte(strings.Join([]string{
		`model = "gpt-5.2"`,
		``,
		`[mcp_servers.Playwright]`,
		`startup_timeout_sec = 60`,
		`command = "npx"`,
		`args = ["-y", "@playwright/mcp@latest"]`,
		``,
		`[features]`,
		`multi_agent = true`,
		``,
	}, "\n")), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	manager := &Manager{HomeDir: tmp}
	names, err := manager.List(target)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(names) != 1 || names[0] != "Playwright" {
		t.Fatalf("unexpected names: %#v", names)
	}
}

func TestManagerInstallAndRemoveUpdatesCodexConfigToml(t *testing.T) {
	tmp := t.TempDir()
	target, ok := targets.Lookup("codex")
	if !ok {
		t.Fatalf("expected codex target")
	}

	configPath := filepath.Join(tmp, target.Dir, "config.toml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(configPath, []byte(strings.Join([]string{
		`model = "gpt-5.2"`,
		``,
		`[features]`,
		`multi_agent = true`,
		``,
	}, "\n")), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	manager := &Manager{HomeDir: tmp}
	if err := manager.Install(target, "context7", InstallOptions{Env: []string{"CONTEXT7_API_KEY=demo"}}); err != nil {
		t.Fatalf("install context7: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	text := string(data)
	if !strings.Contains(text, "[mcp_servers.Context7]") {
		t.Fatalf("expected Context7 block in config.toml, got:\n%s", text)
	}
	if !strings.Contains(text, `env = { CONTEXT7_API_KEY = "demo" }`) {
		t.Fatalf("expected env inline table, got:\n%s", text)
	}

	if err := manager.Remove(target, "Context7"); err != nil {
		t.Fatalf("remove Context7: %v", err)
	}
	data, err = os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config after remove: %v", err)
	}
	if strings.Contains(string(data), "[mcp_servers.Context7]") {
		t.Fatalf("expected Context7 block removed, got:\n%s", string(data))
	}
}

func TestResolveBuiltinTavily(t *testing.T) {
	spec, err := ResolveBuiltin("tavily", InstallOptions{Env: []string{"TAVILY_API_KEY=tvly-test123"}})
	if err != nil {
		t.Fatalf("resolve tavily: %v", err)
	}
	if spec.Name != "tavily" {
		t.Fatalf("unexpected name: %s", spec.Name)
	}
	env, ok := spec.Config["env"].(map[string]string)
	if !ok {
		t.Fatalf("env missing")
	}
	if env["TAVILY_API_KEY"] != "tvly-test123" {
		t.Fatalf("unexpected api key: %v", env["TAVILY_API_KEY"])
	}
}

func TestResolveBuiltinTavilyDefaultEnv(t *testing.T) {
	spec, err := ResolveBuiltin("tavily", InstallOptions{})
	if err != nil {
		t.Fatalf("resolve tavily: %v", err)
	}
	env, ok := spec.Config["env"].(map[string]string)
	if !ok {
		t.Fatalf("env missing")
	}
	if env["TAVILY_API_KEY"] != "${TAVILY_API_KEY}" {
		t.Fatalf("expected placeholder, got: %v", env["TAVILY_API_KEY"])
	}
}

func TestResolveBuiltinTavilyCustom(t *testing.T) {
	spec, err := ResolveBuiltin("tavily-custom", InstallOptions{
		Env: []string{
			"TAVILY_API_URL=http://104.194.69.137:9874",
			"TAVILY_API_KEY=tvly-test123",
		},
	})
	if err != nil {
		t.Fatalf("resolve tavily-custom: %v", err)
	}
	if spec.Name != "tavily-custom" {
		t.Fatalf("unexpected name: %s", spec.Name)
	}
	env, ok := spec.Config["env"].(map[string]string)
	if !ok {
		t.Fatalf("env missing")
	}
	if env["TAVILY_API_URL"] != "http://104.194.69.137:9874" {
		t.Fatalf("unexpected api url: %v", env["TAVILY_API_URL"])
	}
	if env["TAVILY_API_KEY"] != "tvly-test123" {
		t.Fatalf("unexpected api key: %v", env["TAVILY_API_KEY"])
	}
	args, ok := spec.Config["args"].([]any)
	if !ok || len(args) != 1 {
		t.Fatalf("expected args with script path, got: %v", spec.Config["args"])
	}
	if !strings.HasSuffix(args[0].(string), filepath.Join("tavily-custom-mcp", "index.js")) {
		t.Fatalf("expected script path ending in tavily-custom-mcp/index.js, got: %v", args[0])
	}
}

func TestResolveBuiltinTavilyCustomMissingURL(t *testing.T) {
	_, err := ResolveBuiltin("tavily-custom", InstallOptions{
		Env: []string{"TAVILY_API_KEY=tvly-test123"},
	})
	if err == nil {
		t.Fatal("expected error for missing TAVILY_API_URL")
	}
	if !strings.Contains(err.Error(), "TAVILY_API_URL") {
		t.Fatalf("error should mention TAVILY_API_URL, got: %v", err)
	}
}

func TestResolveBuiltinTavilyCustomMissingKey(t *testing.T) {
	_, err := ResolveBuiltin("tavily-custom", InstallOptions{
		Env: []string{"TAVILY_API_URL=http://example.com"},
	})
	if err == nil {
		t.Fatal("expected error for missing TAVILY_API_KEY")
	}
	if !strings.Contains(err.Error(), "TAVILY_API_KEY") {
		t.Fatalf("error should mention TAVILY_API_KEY, got: %v", err)
	}
}

func TestResolveBuiltinTavilyCustomMissingBoth(t *testing.T) {
	_, err := ResolveBuiltin("tavily-custom", InstallOptions{})
	if err == nil {
		t.Fatal("expected error for missing env vars")
	}
}
