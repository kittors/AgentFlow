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

	if err := manager.Install(target, "context7", InstallOptions{Env: []string{"CONTEXT7_API_KEY=demo"}}); err != nil {
		t.Fatalf("install context7: %v", err)
	}

	settingsPath := filepath.Join(tmp, target.Dir, "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read settings: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("parse settings: %v", err)
	}

	mcpServers, ok := payload["mcpServers"].(map[string]any)
	if !ok {
		t.Fatalf("mcpServers missing or invalid")
	}
	context7, ok := mcpServers["context7"].(map[string]any)
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
	data, err = os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read settings: %v", err)
	}
	payload = nil
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("parse settings: %v", err)
	}
	mcpServers, _ = payload["mcpServers"].(map[string]any)
	if _, exists := mcpServers["context7"]; exists {
		t.Fatalf("expected context7 removed")
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
