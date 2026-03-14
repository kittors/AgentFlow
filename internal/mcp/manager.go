package mcp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
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

var codexMCPSrvHeader = regexp.MustCompile(`^\s*\[\s*mcp_servers\.([^\]]+)\s*\]\s*$`)

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

	if target.Name == "codex" {
		return m.installIntoCodexConfig(spec)
	}

	managedPath := managedRecordPath(m.HomeDir, target)
	managed, err := readManagedMCP(managedPath)
	if err != nil {
		return err
	}
	managed[spec.Name] = spec.Config
	if err := writeManagedMCP(managedPath, managed); err != nil {
		return err
	}

	switch target.Name {
	case "claude":
		return m.installIntoClaudeSettings(spec)
	case "gemini":
		return m.installIntoJSONMCPServers(filepath.Join(m.HomeDir, target.Dir, "settings.json"), spec)
	case "qwen":
		return m.installIntoJSONMCPServers(filepath.Join(m.HomeDir, target.Dir, "settings.json"), spec)
	case "kiro":
		return m.installIntoJSONMCPServers(filepath.Join(m.HomeDir, target.Dir, "settings", "mcp.json"), spec)
	case "cursor":
		return m.installIntoJSONMCPServers(filepath.Join(m.HomeDir, target.Dir, "mcp.json"), spec)
	case "windsurf":
		return m.installIntoJSONMCPServers(filepath.Join(m.HomeDir, target.Dir, "mcp_config.json"), spec)
	default:
		// Keep managed record only for unknown targets.
		return nil
	}
}

func (m *Manager) Remove(target targets.Target, server string) error {
	server = strings.TrimSpace(server)
	if server == "" {
		return errors.New("missing server name")
	}

	if target.Name == "codex" {
		return m.removeFromCodexConfig(server)
	}

	managedPath := managedRecordPath(m.HomeDir, target)
	managed, err := readManagedMCP(managedPath)
	if err != nil {
		return err
	}
	delete(managed, server)
	if err := writeManagedMCP(managedPath, managed); err != nil {
		return err
	}

	switch target.Name {
	case "claude":
		return m.removeFromClaudeSettings(server)
	case "gemini":
		return m.removeFromJSONMCPServers(filepath.Join(m.HomeDir, target.Dir, "settings.json"), server)
	case "qwen":
		return m.removeFromJSONMCPServers(filepath.Join(m.HomeDir, target.Dir, "settings.json"), server)
	case "kiro":
		return m.removeFromJSONMCPServers(filepath.Join(m.HomeDir, target.Dir, "settings", "mcp.json"), server)
	case "cursor":
		return m.removeFromJSONMCPServers(filepath.Join(m.HomeDir, target.Dir, "mcp.json"), server)
	case "windsurf":
		return m.removeFromJSONMCPServers(filepath.Join(m.HomeDir, target.Dir, "mcp_config.json"), server)
	default:
		return nil
	}
}

func (m *Manager) List(target targets.Target) ([]string, error) {
	managedPath := managedRecordPath(m.HomeDir, target)
	managed, err := readManagedMCP(managedPath)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(managed))
	for key := range managed {
		names = append(names, key)
	}

	if target.Name == "codex" {
		codexServers, err := m.listCodexConfig()
		if err == nil {
			for key := range codexServers {
				if _, ok := managed[key]; ok {
					continue
				}
				names = append(names, key)
			}
		}
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

	if target.Name == "gemini" || target.Name == "qwen" {
		path := filepath.Join(m.HomeDir, target.Dir, "settings.json")
		servers, err := m.listJSONMCPServers(path)
		if err == nil {
			for key := range servers {
				if _, ok := managed[key]; ok {
					continue
				}
				names = append(names, key)
			}
		}
	}

	if target.Name == "kiro" {
		path := filepath.Join(m.HomeDir, target.Dir, "settings", "mcp.json")
		servers, err := m.listJSONMCPServers(path)
		if err == nil {
			for key := range servers {
				if _, ok := managed[key]; ok {
					continue
				}
				names = append(names, key)
			}
		}
	}

	if target.Name == "cursor" {
		path := filepath.Join(m.HomeDir, target.Dir, "mcp.json")
		servers, err := m.listJSONMCPServers(path)
		if err == nil {
			for key := range servers {
				if _, ok := managed[key]; ok {
					continue
				}
				names = append(names, key)
			}
		}
	}

	if target.Name == "windsurf" {
		path := filepath.Join(m.HomeDir, target.Dir, "mcp_config.json")
		servers, err := m.listJSONMCPServers(path)
		if err == nil {
			for key := range servers {
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

func (m *Manager) listCodexConfig() (map[string]any, error) {
	configPath := filepath.Join(m.HomeDir, targets.All["codex"].Dir, "config.toml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]any{}, nil
		}
		return nil, err
	}

	servers := map[string]any{}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		match := codexMCPSrvHeader.FindStringSubmatch(line)
		if len(match) != 2 {
			continue
		}
		name := strings.TrimSpace(match[1])
		if name == "" {
			continue
		}
		servers[name] = true
	}
	return servers, nil
}

func (m *Manager) installIntoCodexConfig(spec BuiltinSpec) error {
	configPath := filepath.Join(m.HomeDir, targets.All["codex"].Dir, "config.toml")

	data, err := os.ReadFile(configPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	key := codexServerKey(spec.Name)
	updated := upsertCodexMCPBlock(string(data), key, spec.Config)
	return config.SafeWrite(configPath, []byte(updated), 0o600)
}

func (m *Manager) removeFromCodexConfig(name string) error {
	configPath := filepath.Join(m.HomeDir, targets.All["codex"].Dir, "config.toml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	candidates := []string{strings.TrimSpace(name)}
	title := codexServerKey(name)
	if title != "" && title != candidates[0] {
		candidates = append(candidates, title)
	}

	updated := string(data)
	for _, candidate := range candidates {
		if strings.TrimSpace(candidate) == "" {
			continue
		}
		updated = removeCodexMCPBlock(updated, candidate)
	}

	if updated == string(data) {
		return nil
	}
	return config.SafeWrite(configPath, []byte(updated), 0o600)
}

func codexServerKey(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	// Codex CLI commonly uses TitleCase keys (e.g. [mcp_servers.Context7]).
	if len(name) == 1 {
		return strings.ToUpper(name)
	}
	return strings.ToUpper(name[:1]) + name[1:]
}

func upsertCodexMCPBlock(configText, key string, spec map[string]any) string {
	next := removeCodexMCPBlock(configText, key)
	block := renderCodexMCPBlock(key, spec)
	return insertCodexMCPBlock(next, block)
}

func removeCodexMCPBlock(configText, key string) string {
	lines := strings.Split(configText, "\n")
	header := "[mcp_servers." + key + "]"

	start := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == header {
			start = i
			break
		}
	}
	if start < 0 {
		return configText
	}

	end := len(lines)
	for i := start + 1; i < len(lines); i++ {
		if strings.HasPrefix(strings.TrimSpace(lines[i]), "[") && strings.HasSuffix(strings.TrimSpace(lines[i]), "]") {
			end = i
			break
		}
	}

	trimStart := start
	for trimStart > 0 && strings.TrimSpace(lines[trimStart-1]) == "" {
		trimStart--
	}
	trimEnd := end
	for trimEnd < len(lines) && strings.TrimSpace(lines[trimEnd]) == "" {
		trimEnd++
	}

	updated := append([]string{}, lines[:trimStart]...)
	updated = append(updated, lines[trimEnd:]...)
	return strings.TrimRight(strings.Join(updated, "\n"), "\n") + "\n"
}

func insertCodexMCPBlock(configText, block string) string {
	configText = strings.TrimRight(configText, "\n") + "\n"
	lines := strings.Split(configText, "\n")

	lastMCPHeader := -1
	for i, line := range lines {
		if codexMCPSrvHeader.MatchString(line) {
			lastMCPHeader = i
		}
	}
	if lastMCPHeader < 0 {
		return configText + "\n" + block
	}

	insertAt := len(lines)
	for i := lastMCPHeader + 1; i < len(lines); i++ {
		trim := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trim, "[") && strings.HasSuffix(trim, "]") && !strings.HasPrefix(trim, "[mcp_servers.") {
			insertAt = i
			break
		}
	}

	out := make([]string, 0, len(lines)+strings.Count(block, "\n")+2)
	out = append(out, lines[:insertAt]...)
	for len(out) > 0 && strings.TrimSpace(out[len(out)-1]) == "" {
		out = out[:len(out)-1]
	}
	out = append(out, "")
	out = append(out, strings.Split(strings.TrimRight(block, "\n"), "\n")...)
	out = append(out, "")
	out = append(out, lines[insertAt:]...)
	return strings.TrimRight(strings.Join(out, "\n"), "\n") + "\n"
}

func renderCodexMCPBlock(key string, spec map[string]any) string {
	command, _ := spec["command"].(string)
	rawArgs, _ := spec["args"].([]any)
	args := make([]string, 0, len(rawArgs))
	for _, arg := range rawArgs {
		args = append(args, fmt.Sprint(arg))
	}
	env := map[string]string{}
	if rawEnv, ok := spec["env"].(map[string]string); ok {
		env = rawEnv
	} else if rawEnvAny, ok := spec["env"].(map[string]any); ok {
		for k, v := range rawEnvAny {
			env[k] = fmt.Sprint(v)
		}
	}

	lines := []string{
		"[mcp_servers." + key + "]",
		"startup_timeout_sec = 60",
		fmt.Sprintf("command = %s", tomlString(command)),
		fmt.Sprintf("args = %s", tomlStringArray(args)),
	}
	if len(env) > 0 {
		lines = append(lines, fmt.Sprintf("env = %s", tomlInlineTable(env)))
	}
	return strings.Join(lines, "\n") + "\n"
}

func tomlString(value string) string {
	return strconv.Quote(value)
}

func tomlStringArray(values []string) string {
	quoted := make([]string, 0, len(values))
	for _, value := range values {
		quoted = append(quoted, tomlString(value))
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

func tomlInlineTable(values map[string]string) string {
	if len(values) == 0 {
		return "{}"
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s = %s", key, tomlString(values[key])))
	}
	return "{ " + strings.Join(parts, ", ") + " }"
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
	return m.installIntoJSONMCPServers(settingsPath, spec)
}

func (m *Manager) removeFromClaudeSettings(name string) error {
	settingsPath := filepath.Join(m.HomeDir, targets.All["claude"].Dir, "settings.json")
	return m.removeFromJSONMCPServers(settingsPath, name)
}

func (m *Manager) listClaudeSettings() (map[string]any, error) {
	settingsPath := filepath.Join(m.HomeDir, targets.All["claude"].Dir, "settings.json")
	return m.listJSONMCPServers(settingsPath)
}

func managedRecordPath(home string, target targets.Target) string {
	base := filepath.Join(home, target.Dir)
	switch target.Name {
	case "cursor", "windsurf":
		return filepath.Join(base, "agentflow.mcp.json")
	default:
		return filepath.Join(base, "mcp.json")
	}
}

func (m *Manager) installIntoJSONMCPServers(path string, spec BuiltinSpec) error {
	settings, err := readJSONMap(path)
	if err != nil {
		return err
	}

	mcpServers := map[string]any{}
	if existing, ok := settings["mcpServers"]; ok {
		mapped, ok := existing.(map[string]any)
		if !ok {
			return fmt.Errorf("%s: mcpServers must be an object", filepath.Base(path))
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
	return config.SafeWrite(path, payload, 0o644)
}

func (m *Manager) removeFromJSONMCPServers(path, name string) error {
	settings, err := readJSONMap(path)
	if err != nil {
		return err
	}

	existing, ok := settings["mcpServers"]
	if !ok {
		return nil
	}
	mcpServers, ok := existing.(map[string]any)
	if !ok {
		return fmt.Errorf("%s: mcpServers must be an object", filepath.Base(path))
	}
	delete(mcpServers, name)
	settings["mcpServers"] = mcpServers

	payload, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	payload = append(payload, '\n')
	return config.SafeWrite(path, payload, 0o644)
}

func (m *Manager) listJSONMCPServers(path string) (map[string]any, error) {
	settings, err := readJSONMap(path)
	if err != nil {
		return nil, err
	}
	existing, ok := settings["mcpServers"]
	if !ok {
		return map[string]any{}, nil
	}
	mcpServers, ok := existing.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s: mcpServers must be an object", filepath.Base(path))
	}
	return mcpServers, nil
}
