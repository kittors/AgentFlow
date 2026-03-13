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
	if target.Name != "claude" {
		return fmt.Errorf("target does not support MCP config yet: %s", target.Name)
	}

	spec, err := ResolveBuiltin(server, options)
	if err != nil {
		return err
	}

	settingsPath := filepath.Join(m.HomeDir, target.Dir, "settings.json")
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
	return config.SafeWrite(settingsPath, payload, 0o644)
}

func (m *Manager) Remove(target targets.Target, server string) error {
	if target.Name != "claude" {
		return fmt.Errorf("target does not support MCP config yet: %s", target.Name)
	}
	settingsPath := filepath.Join(m.HomeDir, target.Dir, "settings.json")
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
	delete(mcpServers, server)
	settings["mcpServers"] = mcpServers

	payload, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return config.SafeWrite(settingsPath, payload, 0o644)
}

func (m *Manager) List(target targets.Target) ([]string, error) {
	if target.Name != "claude" {
		return nil, fmt.Errorf("target does not support MCP config yet: %s", target.Name)
	}
	settingsPath := filepath.Join(m.HomeDir, target.Dir, "settings.json")
	settings, err := readJSONMap(settingsPath)
	if err != nil {
		return nil, err
	}

	existing, ok := settings["mcpServers"]
	if !ok {
		return []string{}, nil
	}
	mcpServers, ok := existing.(map[string]any)
	if !ok {
		return nil, errors.New("settings.json: mcpServers must be an object")
	}

	names := make([]string, 0, len(mcpServers))
	for key := range mcpServers {
		names = append(names, key)
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
