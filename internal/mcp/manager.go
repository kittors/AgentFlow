package mcp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/kittors/AgentFlow/internal/config"
	"github.com/kittors/AgentFlow/internal/targets"
)

const defaultHTTPTimeout = 2 * time.Minute

type InstallOptions struct {
	Env   []string
	Allow []string
}

type Manager struct {
	Client    *http.Client
	HomeDir   string
	UserAgent string
}

func NewManager() *Manager {
	homeDir, _ := os.UserHomeDir()
	return &Manager{
		Client:    &http.Client{Timeout: defaultHTTPTimeout},
		HomeDir:   homeDir,
		UserAgent: "agentflow-go",
	}
}

func (m *Manager) Install(target targets.Target, server string, options InstallOptions) error {
	spec, err := ResolveBuiltin(server, options)
	if err != nil {
		return err
	}

	managedPath := filepath.Join(m.HomeDir, target.Dir, "mcp.json")
	managed, err := readManagedMCP(managedPath)
	if err != nil {
		return err
	}
	managed[spec.Name] = spec.Config
	if err := writeManagedMCP(managedPath, managed); err != nil {
		return err
	}

	if target.Name == "claude" {
		if err := m.installIntoClaudeSettings(spec); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) Remove(target targets.Target, server string) error {
	server = strings.TrimSpace(server)
	if server == "" {
		return errors.New("missing server name")
	}

	managedPath := filepath.Join(m.HomeDir, target.Dir, "mcp.json")
	managed, err := readManagedMCP(managedPath)
	if err != nil {
		return err
	}
	delete(managed, server)
	if err := writeManagedMCP(managedPath, managed); err != nil {
		return err
	}

	if target.Name == "claude" {
		if err := m.removeFromClaudeSettings(server); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) List(target targets.Target) ([]string, error) {
	managedPath := filepath.Join(m.HomeDir, target.Dir, "mcp.json")
	managed, err := readManagedMCP(managedPath)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(managed))
	for key := range managed {
		names = append(names, key)
	}

	if target.Name == "claude" {
		claudeServers, err := m.listClaudeSettings()
		if err == nil {
			for key := range claudeServers {
				if _, ok := managed[key]; ok {
					continue
				}
				names = append(names, key)
			}
		}
	}

	sort.Strings(names)
	return names, nil
}

func (m *Manager) Search(keyword string) ([]string, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return nil, errors.New("missing keyword")
	}

	results := make([]string, 0, 16)
	for _, spec := range BuiltinServers() {
		if strings.Contains(strings.ToLower(spec.Name), strings.ToLower(keyword)) ||
			strings.Contains(strings.ToLower(spec.Description), strings.ToLower(keyword)) {
			prefix := ""
			if spec.Pinned {
				prefix = "[pinned] "
			}
			results = append(results, fmt.Sprintf("%s%s - %s", prefix, spec.Name, spec.Description))
		}
	}

	market, err := SearchMarket(m.Client, keyword, m.UserAgent)
	if err == nil {
		results = append(results, market...)
	}
	if len(results) == 0 {
		return []string{}, nil
	}
	return results, nil
}

func readJSONMap(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]any{}, nil
		}
		return nil, err
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return map[string]any{}, nil
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	if payload == nil {
		payload = map[string]any{}
	}
	return payload, nil
}

func readManagedMCP(path string) (map[string]any, error) {
	payload, err := readJSONMap(path)
	if err != nil {
		return nil, err
	}
	mcpServers := map[string]any{}
	if raw, ok := payload["mcpServers"]; ok {
		mapped, ok := raw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%s: mcpServers must be an object", filepath.Base(path))
		}
		mcpServers = mapped
	}
	return mcpServers, nil
}

func writeManagedMCP(path string, servers map[string]any) error {
	payload := map[string]any{
		"mcpServers": servers,
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return config.SafeWrite(path, data, 0o644)
}

func (m *Manager) installIntoClaudeSettings(spec BuiltinSpec) error {
	settingsPath := filepath.Join(m.HomeDir, targets.All["claude"].Dir, "settings.json")
	settings, err := readJSONMap(settingsPath)
	if err != nil {
		return err
	}

	mcpServers := map[string]any{}
	if existing, ok := settings["mcpServers"]; ok {
		mapped, ok := existing.(map[string]any)
		if !ok {
			return errors.New("settings.json: mcpServers must be an object")
		}
		mcpServers = mapped
	}

	mcpServers[spec.Name] = spec.Config
	settings["mcpServers"] = mcpServers

	payload, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	payload = append(payload, '\n')
	return config.SafeWrite(settingsPath, payload, 0o644)
}

func (m *Manager) removeFromClaudeSettings(name string) error {
	settingsPath := filepath.Join(m.HomeDir, targets.All["claude"].Dir, "settings.json")
	settings, err := readJSONMap(settingsPath)
	if err != nil {
		return err
	}

	existing, ok := settings["mcpServers"]
	if !ok {
		return nil
	}
	mcpServers, ok := existing.(map[string]any)
	if !ok {
		return errors.New("settings.json: mcpServers must be an object")
	}
	delete(mcpServers, name)
	settings["mcpServers"] = mcpServers

	payload, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	payload = append(payload, '\n')
	return config.SafeWrite(settingsPath, payload, 0o644)
}

func (m *Manager) listClaudeSettings() (map[string]any, error) {
	settingsPath := filepath.Join(m.HomeDir, targets.All["claude"].Dir, "settings.json")
	settings, err := readJSONMap(settingsPath)
	if err != nil {
		return nil, err
	}
	existing, ok := settings["mcpServers"]
	if !ok {
		return map[string]any{}, nil
	}
	mcpServers, ok := existing.(map[string]any)
	if !ok {
		return nil, errors.New("settings.json: mcpServers must be an object")
	}
	return mcpServers, nil
}
