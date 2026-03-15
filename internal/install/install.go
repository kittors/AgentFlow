package install

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	agentflowassets "github.com/kittors/AgentFlow"
	"github.com/kittors/AgentFlow/internal/config"
	"github.com/kittors/AgentFlow/internal/i18n"
	"github.com/kittors/AgentFlow/internal/targets"
)

var (
	targetSubagentFiles = map[string]string{
		"codex":    "subagent_codex.md",
		"claude":   "subagent_claude.md",
		"gemini":   "subagent_gemini.md",
		"qwen":     "subagent_other.md",
		"kiro":     "subagent_other.md",
		"opencode": "subagent_opencode.md",
		"grok":     "subagent_other.md",
	}
	targetHooksFiles = map[string]string{
		"codex":    "hooks_codex.md",
		"claude":   "hooks_claude.md",
		"gemini":   "hooks_other.md",
		"qwen":     "hooks_other.md",
		"kiro":     "hooks_other.md",
		"opencode": "hooks_other.md",
		"grok":     "hooks_other.md",
	}
)

type Installer struct {
	Catalog       i18n.Catalog
	Stdout        io.Writer
	HomeDir       string
	cachedRuntime *RuntimeStatus // cached to avoid repeated shell invocations
}

func New(catalog i18n.Catalog, stdout io.Writer) *Installer {
	homeDir, _ := os.UserHomeDir()
	return &Installer{
		Catalog: catalog,
		Stdout:  stdout,
		HomeDir: homeDir,
	}
}

// CachedRuntimeStatus returns the cached RuntimeStatus, computing it once
// on first call. This avoids spawning multiple expensive shell subprocesses
// (node --version, npm --version, nvm detection, etc.) on every call.
func (i *Installer) CachedRuntimeStatus() RuntimeStatus {
	if i.cachedRuntime != nil {
		return *i.cachedRuntime
	}
	status := i.RuntimeStatus()
	i.cachedRuntime = &status
	return status
}

// EnvVarConfig describes a single environment variable to configure for a CLI.
type EnvVarConfig struct {
	Label   string   // Human-readable label (e.g. "API Key")
	EnvVar  string   // Environment variable name (e.g. "OPENAI_API_KEY")
	Type    string   // "text" for free-form input, "select" for option list
	Options []string // For "select" type: available choices
	Default string   // Default value (pre-selected for "select", placeholder hint for "text")
}

// CLIConfigFields returns the environment variables and settings that should
// be configured after installing the given CLI target.
func (i *Installer) CLIConfigFields(targetName string) []EnvVarConfig {
	target, ok := targets.Lookup(targetName)
	if !ok {
		return nil
	}

	var fields []EnvVarConfig

	// API Key field.
	if target.APIKeyEnv != "" {
		fields = append(fields, EnvVarConfig{
			Label:  "API Key",
			EnvVar: target.APIKeyEnv,
			Type:   "text",
		})
	}

	// Base URL field.
	if target.BaseURLEnv != "" {
		fields = append(fields, EnvVarConfig{
			Label:  "Base URL",
			EnvVar: target.BaseURLEnv,
			Type:   "text",
		})
	}

	// Model selection field.
	if len(target.Models) > 0 {
		envVar := target.ModelEnv
		if envVar == "" {
			// Codex uses config.json instead of env var, but we still
			// present it as a selectable field for the UI.
			envVar = "__MODEL__"
		}
		options := make([]string, len(target.Models))
		defaultVal := ""
		for idx, m := range target.Models {
			options[idx] = m.Value
			if m.Default {
				defaultVal = m.Value
			}
		}
		fields = append(fields, EnvVarConfig{
			Label:   i.Catalog.Msg("默认模型", "Default Model"),
			EnvVar:  envVar,
			Type:    "select",
			Options: options,
			Default: defaultVal,
		})
	}

	// Codex-specific: thinking/reasoning level.
	if target.HasConfigFile && target.Name == "codex" {
		fields = append(fields, EnvVarConfig{
			Label:   i.Catalog.Msg("思考等级", "Thinking Level"),
			EnvVar:  "__CODEX_REASONING__",
			Type:    "select",
			Options: []string{"low", "medium", "high"},
			Default: "medium",
		})
	}

	return fields
}

// WriteCodexConfig writes model and reasoning settings to ~/.codex/config.json.
func (i *Installer) WriteCodexConfig(model, reasoning string) error {
	configPath := filepath.Join(i.HomeDir, ".codex", "config.json")

	var settings map[string]any
	if data, err := os.ReadFile(configPath); err == nil {
		_ = json.Unmarshal(data, &settings)
	}
	if settings == nil {
		settings = make(map[string]any)
	}

	if model != "" {
		settings["model"] = model
	}
	if reasoning != "" {
		settings["reasoning"] = reasoning
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0o644)
}

func (i *Installer) Install(targetName, profile string) error {
	target, ok := targets.Lookup(targetName)
	if !ok {
		return fmt.Errorf(i.Catalog.Msg("未知目标: %s", "unknown target: %s"), targetName)
	}
	if profile == "" {
		profile = targets.DefaultProfile
	}
	if !targets.ValidProfile(profile) {
		return fmt.Errorf(i.Catalog.Msg("未知 profile: %s", "unknown profile: %s"), profile)
	}

	cliDir := filepath.Join(i.HomeDir, target.Dir)
	if err := os.MkdirAll(cliDir, 0o755); err != nil {
		return fmt.Errorf(i.Catalog.Msg("无法创建目标目录: %s", "failed to create target directory: %s"), cliDir)
	}

	if err := i.deployRulesFile(target, profile); err != nil {
		return err
	}
	if err := i.deployModuleDir(target); err != nil {
		return err
	}
	if err := i.deploySkill(target); err != nil {
		return err
	}
	if err := i.deployHooks(target); err != nil {
		return err
	}
	if target.Name == "codex" {
		if err := i.deployCodexAgents(target); err != nil {
			return err
		}
	}
	if target.Name == "kiro" {
		if err := i.deployKiroAgent(target, profile); err != nil {
			return err
		}
	}
	return nil
}

func (i *Installer) InstallAll(profile string) (int, error) {
	success := 0
	for _, name := range i.AgentFlowInstallableTargets() {
		if err := i.Install(name, profile); err == nil {
			success++
		}
	}
	if success == 0 {
		return 0, fmt.Errorf("%s", i.Catalog.Msg("未检测到任何已安装的 CLI", "no installed CLIs detected"))
	}
	return success, nil
}

func (i *Installer) Uninstall(targetName string) error {
	target, ok := targets.Lookup(targetName)
	if !ok {
		return fmt.Errorf(i.Catalog.Msg("未知目标: %s", "unknown target: %s"), targetName)
	}
	cliDir := filepath.Join(i.HomeDir, target.Dir)
	rulesFile := filepath.Join(cliDir, target.RulesFile)
	if config.IsAgentFlowFile(rulesFile) {
		if err := config.SafeRemove(rulesFile); err != nil {
			return err
		}
	}
	if err := config.SafeRemove(filepath.Join(cliDir, config.PluginDirName)); err != nil {
		return err
	}
	if err := config.SafeRemove(filepath.Join(cliDir, "skills", "agentflow")); err != nil {
		return err
	}

	if target.Name == "codex" {
		if err := i.cleanCodexAgents(cliDir); err != nil {
			return err
		}
	}
	if target.Name == "claude" {
		if err := i.cleanClaudeHooks(filepath.Join(cliDir, "settings.json")); err != nil {
			return err
		}
	}
	if target.Name == "kiro" {
		if err := i.cleanKiroAgent(cliDir); err != nil {
			return err
		}
	}
	return nil
}

func (i *Installer) deployKiroAgent(target targets.Target, profile string) error {
	cliDir := filepath.Join(i.HomeDir, target.Dir)

	// Kiro custom agents can reference a local prompt file.
	promptContent, err := i.buildRulesContent(target.Name, profile)
	if err != nil {
		return err
	}
	promptPath := filepath.Join(cliDir, "prompts", "agentflow.md")
	if err := config.SafeWrite(promptPath, []byte(promptContent), 0o644); err != nil {
		return err
	}

	agentPath := filepath.Join(cliDir, "agents", "agentflow.json")
	payload := map[string]any{
		"name":        "agentflow",
		"description": "AGENTFLOW_ROUTER: managed by AgentFlow (R0-R4 router, EHRB safety, KB-first workflow).",
		"prompt":      "file://../prompts/agentflow.md",
		"resources":   []string{},
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return config.SafeWrite(agentPath, data, 0o644)
}

func (i *Installer) cleanKiroAgent(cliDir string) error {
	for _, path := range []string{
		filepath.Join(cliDir, "agents", "agentflow.json"),
		filepath.Join(cliDir, "prompts", "agentflow.md"),
	} {
		if config.IsAgentFlowFile(path) {
			if err := config.SafeRemove(path); err != nil {
				return err
			}
		}
	}
	return nil
}

func (i *Installer) UninstallAll() (int, error) {
	installed := i.DetectInstalledTargets()
	for _, name := range installed {
		if err := i.Uninstall(name); err != nil {
			return 0, err
		}
	}
	return len(installed), nil
}

func (i *Installer) DetectInstalledCLIs() []string {
	result := make([]string, 0)
	for _, status := range i.DetectTargetStatuses() {
		if status.CLIInstalled {
			result = append(result, status.Target.Name)
			continue
		}
		if status.ConfigDirExists && status.Target.Command == "" {
			result = append(result, status.Target.Name)
		}
	}
	return result
}

func (i *Installer) DetectInstalledTargets() []string {
	result := make([]string, 0)
	for _, name := range targets.Names() {
		target := targets.All[name]
		cliDir := filepath.Join(i.HomeDir, target.Dir)
		pluginDir := filepath.Join(cliDir, config.PluginDirName)
		rulesFile := filepath.Join(cliDir, target.RulesFile)
		if _, err := os.Stat(pluginDir); err != nil {
			continue
		}
		if config.IsAgentFlowFile(rulesFile) {
			result = append(result, name)
		}
	}
	return result
}

func (i *Installer) Clean() (int, error) {
	cleaned := 0
	for _, name := range i.DetectInstalledTargets() {
		target := targets.All[name]
		cliDir := filepath.Join(i.HomeDir, target.Dir)
		for _, cacheDir := range []string{
			filepath.Join(cliDir, config.PluginDirName, "__pycache__"),
			filepath.Join(cliDir, config.PluginDirName, ".cache"),
		} {
			if _, err := os.Stat(cacheDir); err == nil {
				if err := config.SafeRemove(cacheDir); err != nil {
					return cleaned, err
				}
				cleaned++
			}
		}
	}
	return cleaned, nil
}

func (i *Installer) StatusLines() []string {
	lines := []string{
		i.Catalog.Msg("CLI 状态:", "CLI status:"),
	}
	for _, status := range i.DetectTargetStatuses() {
		switch {
		case status.CLIInstalled && status.AgentFlowInstalled:
			lines = append(lines, fmt.Sprintf("  [OK] %s", status.Target.Name))
		case status.CLIInstalled:
			lines = append(lines, fmt.Sprintf("  [CLI] %s", status.Target.Name))
		case status.AgentFlowInstalled:
			lines = append(lines, fmt.Sprintf("  [AF] %s", status.Target.Name))
		case status.ConfigDirExists:
			lines = append(lines, fmt.Sprintf("  [..] %s", status.Target.Name))
		default:
			lines = append(lines, fmt.Sprintf("  [--] %s", status.Target.Name))
		}
	}
	return lines
}

func (i *Installer) deployRulesFile(target targets.Target, profile string) error {
	rulesPath := filepath.Join(i.HomeDir, target.Dir, target.RulesFile)
	if _, err := os.Stat(rulesPath); err == nil && !config.IsAgentFlowFile(rulesPath) {
		if _, err := config.BackupUserFile(rulesPath); err != nil {
			return err
		}
	}

	rendered, err := i.buildRulesContent(target.Name, profile)
	if err != nil {
		return err
	}
	return config.SafeWrite(rulesPath, []byte(rendered), 0o644)
}

func (i *Installer) buildRulesContent(targetName, profile string) (string, error) {
	target, ok := targets.Lookup(targetName)
	if !ok {
		return "", fmt.Errorf("unknown target: %s", targetName)
	}
	if profile == "" {
		profile = targets.DefaultProfile
	}
	if !targets.ValidProfile(profile) {
		return "", fmt.Errorf("invalid profile: %s", profile)
	}

	content, err := readAssetWithFallback("agentflow/_AGENTS.md", "AGENTS.md")
	if err != nil {
		return "", err
	}

	rendered := string(content)
	if !strings.Contains(rendered, config.Marker) {
		rendered = "<!-- AGENTFLOW_ROUTER: v1.0.0 -->\n" + rendered
	}
	rendered = strings.ReplaceAll(rendered, "{TARGET_CLI}", target.DisplayName)
	rendered = strings.ReplaceAll(rendered, "{HOOKS_SUMMARY}", target.HooksSummary)

	modules := targets.Profiles[profile]
	if len(modules) == 0 {
		return rendered, nil
	}

	var builder strings.Builder
	builder.WriteString(rendered)
	builder.WriteString("\n\n---\n\n")
	builder.WriteString(fmt.Sprintf("<!-- PROFILE:%s — Extended modules appended below -->\n\n", profile))
	for _, modFile := range modules {
		moduleContent, err := i.buildCoreModuleForTarget(targetName, modFile)
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(moduleContent) == "" {
			continue
		}
		builder.WriteString(moduleContent)
		builder.WriteString("\n\n")
	}
	return strings.TrimRight(builder.String(), "\n") + "\n", nil
}

func (i *Installer) buildCoreModuleForTarget(targetName, modFile string) (string, error) {
	content, err := readAssetWithFallback(filepath.ToSlash(filepath.Join("agentflow", "core", modFile)))
	if err != nil {
		return "", err
	}

	rendered := string(content)
	if modFile == "subagent.md" {
		cliFile := targetSubagentFiles[targetName]
		if cliFile == "" {
			cliFile = "subagent_other.md"
		}
		cliContent, err := readAssetWithFallback(filepath.ToSlash(filepath.Join("agentflow", "core", cliFile)))
		if err != nil {
			return "", err
		}
		rendered = strings.ReplaceAll(rendered, "{CLI_SUBAGENT_PROTOCOL}", string(cliContent))
	}
	if modFile == "hooks.md" {
		cliFile := targetHooksFiles[targetName]
		if cliFile == "" {
			cliFile = "hooks_other.md"
		}
		cliContent, err := readAssetWithFallback(filepath.ToSlash(filepath.Join("agentflow", "core", cliFile)))
		if err != nil {
			return "", err
		}
		rendered = strings.ReplaceAll(rendered, "{HOOKS_MATRIX}", string(cliContent))
	}
	return rendered, nil
}

func (i *Installer) deployModuleDir(target targets.Target) error {
	baseDir := filepath.Join(i.HomeDir, target.Dir, config.PluginDirName)
	if err := config.SafeRemove(baseDir); err != nil {
		return err
	}
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return err
	}

	moduleFS, err := agentflowassets.Sub("agentflow")
	if err != nil {
		return err
	}

	return fs.WalkDir(moduleFS, ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == "." {
			return nil
		}
		targetPath := filepath.Join(baseDir, filepath.FromSlash(path))
		if entry.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}
		content, err := fs.ReadFile(moduleFS, path)
		if err != nil {
			return err
		}
		return config.SafeWrite(targetPath, content, 0o644)
	})
}

func (i *Installer) deploySkill(target targets.Target) error {
	content, err := readAssetWithFallback("agentflow/_SKILL.md", "SKILL.md")
	if err != nil {
		return err
	}
	skillPath := filepath.Join(i.HomeDir, target.Dir, "skills", "agentflow", "SKILL.md")
	return config.SafeWrite(skillPath, content, 0o644)
}

func (i *Installer) deployHooks(target targets.Target) error {
	if target.Name != "claude" {
		return nil
	}

	content, err := readAssetWithFallback("agentflow/hooks/claude_hooks.json")
	if err != nil {
		return err
	}

	settingsPath := filepath.Join(i.HomeDir, target.Dir, "settings.json")
	var settings map[string]any
	if data, err := os.ReadFile(settingsPath); err == nil {
		_ = json.Unmarshal(data, &settings)
	}
	if settings == nil {
		settings = make(map[string]any)
	}

	var payload map[string]any
	if err := json.Unmarshal(content, &payload); err != nil {
		return err
	}

	existingHooks := map[string]any{}
	if rawHooks, ok := settings["hooks"]; ok {
		if mapped, ok := rawHooks.(map[string]any); ok {
			existingHooks = mapped
		}
	}

	newHooks, _ := payload["hooks"].(map[string]any)
	for eventName, rawGroups := range newHooks {
		existingGroups := normalizeHookGroups(existingHooks[eventName])
		cleanedGroups := make([]map[string]any, 0, len(existingGroups))
		for _, group := range existingGroups {
			filtered := filterAgentFlowHooks(group)
			if len(filtered["hooks"].([]map[string]any)) == 0 {
				continue
			}
			cleanedGroups = append(cleanedGroups, filtered)
		}
		cleanedGroups = append(cleanedGroups, normalizeHookGroups(rawGroups)...)
		existingHooks[eventName] = cleanedGroups
	}

	settings["hooks"] = existingHooks
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return config.SafeWrite(settingsPath, data, 0o644)
}

func (i *Installer) deployCodexAgents(target targets.Target) error {
	cliDir := filepath.Join(i.HomeDir, target.Dir)
	if err := i.deployAgentTomlFiles(cliDir); err != nil {
		return err
	}
	return i.mergeCodexConfig(filepath.Join(cliDir, "config.toml"))
}

func (i *Installer) deployAgentTomlFiles(cliDir string) error {
	agentsDir := filepath.Join(cliDir, "agents")
	if err := os.MkdirAll(agentsDir, 0o755); err != nil {
		return err
	}

	for _, roleFile := range []string{"reviewer.toml", "architect.toml"} {
		targetPath := filepath.Join(agentsDir, roleFile)
		if _, err := os.Stat(targetPath); err == nil && !config.IsAgentFlowFile(targetPath) {
			if _, err := config.BackupUserFile(targetPath); err != nil {
				return err
			}
		}

		content, err := readAssetWithFallback(filepath.ToSlash(filepath.Join("agentflow", "agents", roleFile)))
		if err != nil {
			return err
		}
		rendered := string(content)
		if !strings.Contains(rendered, config.Marker) {
			rendered = "# AGENTFLOW_ROUTER: managed by AgentFlow\n" + rendered
		}
		if err := config.SafeWrite(targetPath, []byte(rendered), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func (i *Installer) mergeCodexConfig(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	text := string(data)
	text = ensureFeatureMultiAgent(text)
	text = ensureSection(
		text,
		"agents.reviewer",
		"[agents.reviewer]\n"+
			"# AGENTFLOW_ROUTER: managed by AgentFlow\n"+
			"description = \"AgentFlow code reviewer: security, correctness, test quality analysis.\"\n"+
			"config_file = \"agents/reviewer.toml\"\n",
	)
	text = ensureSection(
		text,
		"agents.architect",
		"[agents.architect]\n"+
			"# AGENTFLOW_ROUTER: managed by AgentFlow\n"+
			"description = \"AgentFlow architect: architectural evaluation, dependency analysis, multi-approach comparison.\"\n"+
			"config_file = \"agents/architect.toml\"\n",
	)

	return config.SafeWrite(configPath, []byte(strings.TrimSpace(text)+"\n"), 0o644)
}

func (i *Installer) cleanCodexAgents(cliDir string) error {
	for _, roleFile := range []string{"reviewer.toml", "architect.toml"} {
		fullPath := filepath.Join(cliDir, "agents", roleFile)
		if config.IsAgentFlowFile(fullPath) {
			if err := config.SafeRemove(fullPath); err != nil {
				return err
			}
		}
	}
	return i.cleanCodexConfig(filepath.Join(cliDir, "config.toml"))
}

func (i *Installer) cleanCodexConfig(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	text := string(data)
	text = removeSection(text, "agents.reviewer", "AgentFlow", `config_file = "agents/reviewer.toml"`)
	text = removeSection(text, "agents.architect", "AgentFlow", `config_file = "agents/architect.toml"`)
	text = removeFeatureMultiAgent(text)
	text = collapseBlankLines(text)

	return config.SafeWrite(configPath, []byte(strings.TrimSpace(text)+"\n"), 0o644)
}

func (i *Installer) cleanClaudeHooks(settingsPath string) error {
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil
	}
	rawHooks, ok := settings["hooks"]
	if !ok {
		return nil
	}

	hooksMap, ok := rawHooks.(map[string]any)
	if !ok {
		settings["hooks"] = map[string]any{}
		cleaned, err := json.MarshalIndent(settings, "", "  ")
		if err != nil {
			return err
		}
		return config.SafeWrite(settingsPath, cleaned, 0o644)
	}

	cleanedHooks := make(map[string]any, len(hooksMap))
	for eventName, rawGroups := range hooksMap {
		groups := normalizeHookGroups(rawGroups)
		cleanedGroups := make([]map[string]any, 0, len(groups))
		for _, group := range groups {
			filtered := filterAgentFlowHooks(group)
			if len(filtered["hooks"].([]map[string]any)) == 0 {
				continue
			}
			cleanedGroups = append(cleanedGroups, filtered)
		}
		if len(cleanedGroups) > 0 {
			cleanedHooks[eventName] = cleanedGroups
		}
	}

	settings["hooks"] = cleanedHooks
	cleaned, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return config.SafeWrite(settingsPath, cleaned, 0o644)
}

func readAssetWithFallback(paths ...string) ([]byte, error) {
	var lastErr error
	for _, candidate := range paths {
		content, err := agentflowassets.ReadFile(filepath.ToSlash(candidate))
		if err == nil {
			return content, nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("asset not found")
	}
	return nil, lastErr
}

func normalizeHookGroups(raw any) []map[string]any {
	groups, ok := raw.([]any)
	if !ok {
		if typedGroups, ok := raw.([]map[string]any); ok {
			return typedGroups
		}
		return nil
	}

	result := make([]map[string]any, 0, len(groups))
	for _, item := range groups {
		group, ok := item.(map[string]any)
		if !ok {
			continue
		}
		result = append(result, normalizeHookGroup(group))
	}
	return result
}

func normalizeHookGroup(group map[string]any) map[string]any {
	normalized := make(map[string]any, len(group))
	for key, value := range group {
		normalized[key] = value
	}

	hooks := make([]map[string]any, 0)
	switch rawHooks := group["hooks"].(type) {
	case []any:
		hooks = make([]map[string]any, 0, len(rawHooks))
		for _, item := range rawHooks {
			hook, ok := item.(map[string]any)
			if !ok {
				continue
			}
			clone := make(map[string]any, len(hook))
			for key, value := range hook {
				clone[key] = value
			}
			hooks = append(hooks, clone)
		}
	case []map[string]any:
		hooks = make([]map[string]any, 0, len(rawHooks))
		for _, hook := range rawHooks {
			clone := make(map[string]any, len(hook))
			for key, value := range hook {
				clone[key] = value
			}
			hooks = append(hooks, clone)
		}
	}
	normalized["hooks"] = hooks
	return normalized
}

func filterAgentFlowHooks(group map[string]any) map[string]any {
	normalized := normalizeHookGroup(group)
	hooks, _ := normalized["hooks"].([]map[string]any)
	filtered := make([]map[string]any, 0, len(hooks))
	for _, hook := range hooks {
		description, _ := hook["description"].(string)
		if isAgentFlowHook(description) {
			continue
		}
		filtered = append(filtered, hook)
	}
	normalized["hooks"] = filtered
	return normalized
}

func isAgentFlowHook(description string) bool {
	return strings.HasPrefix(description, config.PluginDirName) || strings.HasPrefix(strings.ToLower(description), "agentflow")
}

func ensureFeatureMultiAgent(text string) string {
	start, end, section := sectionRange(text, "features")
	if start == -1 {
		if strings.TrimSpace(text) == "" {
			return "[features]\n# AGENTFLOW_ROUTER: managed by AgentFlow\nmulti_agent = true\n"
		}
		return strings.TrimRight(text, "\n") + "\n\n[features]\n# AGENTFLOW_ROUTER: managed by AgentFlow\nmulti_agent = true\n"
	}

	linePattern := regexp.MustCompile(`(?m)^multi_agent\s*=.*$`)
	if linePattern.MatchString(section) {
		current := linePattern.FindString(section)
		if strings.TrimSpace(current) == "multi_agent = true" && !strings.Contains(section, "# AGENTFLOW_ROUTER: managed by AgentFlow") {
			return text
		}
		replacement := "# AGENTFLOW_ROUTER: managed by AgentFlow\nmulti_agent = true"
		section = linePattern.ReplaceAllString(section, replacement)
	} else {
		section = strings.TrimRight(section, "\n") + "\n# AGENTFLOW_ROUTER: managed by AgentFlow\nmulti_agent = true\n"
	}
	return text[:start] + section + text[end:]
}

func ensureSection(text, header, section string) string {
	if strings.Contains(text, "["+header+"]") {
		return text
	}
	if strings.TrimSpace(text) == "" {
		return section
	}
	return strings.TrimRight(text, "\n") + "\n\n" + section
}

func removeSection(text, header string, signatures ...string) string {
	start, end, _ := sectionRange(text, header)
	if start == -1 {
		return text
	}
	section := text[start:end]
	if !isManagedSection(section, signatures...) {
		return text
	}
	return text[:start] + text[end:]
}

func removeFeatureMultiAgent(text string) string {
	start, end, section := sectionRange(text, "features")
	if start == -1 {
		return text
	}

	linePattern := regexp.MustCompile(`(?m)^# AGENTFLOW_ROUTER: managed by AgentFlow\nmulti_agent\s*=\s*true\s*\n?`)
	cleaned := linePattern.ReplaceAllString(section, "")
	cleaned = collapseBlankLines(cleaned)
	if strings.TrimSpace(cleaned) == "[features]" {
		return text[:start] + text[end:]
	}
	return text[:start] + cleaned + text[end:]
}

func collapseBlankLines(text string) string {
	text = regexp.MustCompile(`\n{3,}`).ReplaceAllString(text, "\n\n")
	return strings.TrimLeft(text, "\n")
}

func sectionRange(text, header string) (int, int, string) {
	marker := "[" + header + "]"
	start := strings.Index(text, marker)
	if start == -1 {
		return -1, -1, ""
	}

	if lineStart := strings.LastIndex(text[:start], "\n"); lineStart >= 0 {
		start = lineStart + 1
	} else {
		start = 0
	}

	searchStart := start + len(marker)
	rest := text[searchStart:]
	nextRel := regexp.MustCompile(`(?m)^\[`).FindStringIndex(rest)
	end := len(text)
	if nextRel != nil {
		end = searchStart + nextRel[0]
	}
	return start, end, text[start:end]
}

func isManagedSection(section string, signatures ...string) bool {
	if strings.Contains(section, config.Marker) {
		return true
	}
	for _, signature := range signatures {
		if !strings.Contains(section, signature) {
			return false
		}
	}
	return len(signatures) > 0
}
