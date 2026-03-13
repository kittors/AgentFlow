package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
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
