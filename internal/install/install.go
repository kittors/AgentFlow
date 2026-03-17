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

	"github.com/charmbracelet/lipgloss"

	agentflowassets "github.com/kittors/AgentFlow"
	"github.com/kittors/AgentFlow/internal/config"
	"github.com/kittors/AgentFlow/internal/i18n"
	"github.com/kittors/AgentFlow/internal/targets"
)

var (
	targetSubagentFiles = map[string]string{
		"codex":  "subagent_codex.md",
		"claude": "subagent_claude.md",
	}
	targetHooksFiles = map[string]string{
		"codex":  "hooks_codex.md",
		"claude": "hooks_claude.md",
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

// InvalidateCache clears the cached RuntimeStatus so the next call
// to CachedRuntimeStatus() re-evaluates all shell commands.
func (i *Installer) InvalidateCache() {
	i.cachedRuntime = nil
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

// WriteCodexConfig writes Codex configuration files to ~/.codex/:
//   - auth.json — API key authentication token
//   - config.toml — model, reasoning, model_provider (base URL), and features
//
// Any empty parameter is skipped (existing values are preserved).
func (i *Installer) WriteCodexConfig(apiKey, baseURL, model, reasoning string) error {
	codexDir := filepath.Join(i.HomeDir, ".codex")
	if err := os.MkdirAll(codexDir, 0o755); err != nil {
		return err
	}

	// ── auth.json ──
	if apiKey != "" {
		authPath := filepath.Join(codexDir, "auth.json")
		var auth map[string]any
		if data, err := os.ReadFile(authPath); err == nil {
			_ = json.Unmarshal(data, &auth)
		}
		if auth == nil {
			auth = make(map[string]any)
		}
		auth["token"] = apiKey
		data, err := json.MarshalIndent(auth, "", "  ")
		if err != nil {
			return err
		}
		data = append(data, '\n')
		if err := config.SafeWrite(authPath, data, 0o600); err != nil {
			return err
		}
	}

	// ── config.toml ──
	configPath := filepath.Join(codexDir, "config.toml")
	var text string
	if data, err := os.ReadFile(configPath); err == nil {
		text = string(data)
	}

	// Model field.
	if model != "" {
		re := regexp.MustCompile(`(?m)^model\s*=.*$`)
		if re.MatchString(text) {
			text = re.ReplaceAllString(text, fmt.Sprintf(`model = "%s"`, model))
		} else {
			text = fmt.Sprintf("model = %q\n%s", model, text)
		}
	}
	// Reasoning effort.
	if reasoning != "" {
		re := regexp.MustCompile(`(?m)^model_reasoning_effort\s*=.*$`)
		if re.MatchString(text) {
			text = re.ReplaceAllString(text, fmt.Sprintf(`model_reasoning_effort = "%s"`, reasoning))
		} else {
			text = fmt.Sprintf("model_reasoning_effort = %q\n%s", reasoning, text)
		}
	}
	// Custom base_url for third-party API proxies.
	// Uses model_provider + [model_providers.agentflow] WITH env_key = "OPENAI_API_KEY".
	// Codex reads the API key from the OPENAI_API_KEY env var (written to shell RC
	// by writeEnvConfigPanel). cc-switch uses a local proxy to inject API keys,
	// but AgentFlow connects directly, so env_key is required.
	if baseURL != "" {
		// Remove any legacy openai_base_url (older AgentFlow versions used this).
		reOldBaseURL := regexp.MustCompile(`(?m)^openai_base_url\s*=.*\n`)
		text = reOldBaseURL.ReplaceAllString(text, "")

		providerName := "agentflow"
		// Set top-level model_provider.
		reProvider := regexp.MustCompile(`(?m)^model_provider\s*=.*$`)
		if reProvider.MatchString(text) {
			text = reProvider.ReplaceAllString(text, fmt.Sprintf(`model_provider = "%s"`, providerName))
		} else {
			text = insertTopLevel(text, fmt.Sprintf("model_provider = %q", providerName))
		}
		// Append or replace the [model_providers.agentflow] section.
		sectionRe := regexp.MustCompile(`(?ms)^\[model_providers\.` + regexp.QuoteMeta(providerName) + `\].*?(?:\n\[|\z)`)
		providerSection := fmt.Sprintf("[model_providers.%s]\nname = %q\nbase_url = %q\nenv_key = \"OPENAI_API_KEY\"\nwire_api = \"responses\"\n",
			providerName, providerName, baseURL)
		if loc := sectionRe.FindStringIndex(text); loc != nil {
			end := loc[1]
			if end > 0 && text[end-1] == '[' {
				end--
			}
			text = text[:loc[0]] + providerSection + text[end:]
		} else {
			text = strings.TrimRight(text, "\n") + "\n\n" + providerSection
		}
	}

	return config.SafeWrite(configPath, []byte(strings.TrimSpace(text)+"\n"), 0o644)
}

// WriteClaudeConfig writes the default model setting to ~/.claude.json.
func (i *Installer) WriteClaudeConfig(model string) error {
	if model == "" {
		return nil
	}
	configPath := filepath.Join(i.HomeDir, ".claude.json")

	var settings map[string]any
	if data, err := os.ReadFile(configPath); err == nil {
		_ = json.Unmarshal(data, &settings)
	}
	if settings == nil {
		settings = make(map[string]any)
	}

	settings["model"] = model

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	return config.SafeWrite(configPath, data, 0o644)
}

// cleanCodexBootstrapConfig removes AgentFlow-written config from Codex CLI files.
// This cleans config.toml fields (model, reasoning, base_url, model_provider)
// that were written by AgentFlow's bootstrap config step.
// NOTE: auth.json is intentionally NOT cleaned — the API key is the user's
// authentication credential and must be preserved across AgentFlow uninstalls.
// Best-effort: errors are silently ignored so uninstall always succeeds.
func (i *Installer) cleanCodexBootstrapConfig() {
	codexDir := filepath.Join(i.HomeDir, ".codex")

	// Clean config.toml — remove AgentFlow-written fields.
	configPath := filepath.Join(codexDir, "config.toml")
	if data, err := os.ReadFile(configPath); err == nil {
		text := string(data)

		// Remove top-level fields written by AgentFlow.
		for _, field := range []string{"model", "model_reasoning_effort", "openai_base_url", "model_provider"} {
			re := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(field) + `\s*=.*\n`)
			text = re.ReplaceAllString(text, "")
		}

		// Remove AgentFlow comment lines (e.g. inside [features]).
		reComment := regexp.MustCompile(`(?m)^#[^\n]*AGENTFLOW[^\n]*\n`)
		text = reComment.ReplaceAllString(text, "")

		// Remove [model_providers.agentflow] section entirely.
		sectionRe := regexp.MustCompile(`(?ms)^\[model_providers\.agentflow\].*?(?:\n\[|\z)`)
		if loc := sectionRe.FindStringIndex(text); loc != nil {
			end := loc[1]
			// If the match ends with a new section header "[", keep it.
			if end > 0 && text[end-1] == '[' {
				end--
			}
			text = text[:loc[0]] + text[end:]
		}

		text = collapseBlankLines(text)
		cleaned := strings.TrimSpace(text)
		if cleaned == "" {
			_ = config.SafeRemove(configPath)
		} else {
			_ = config.SafeWrite(configPath, []byte(cleaned+"\n"), 0o644)
		}
	}
}

// cleanClaudeBootstrapConfig removes AgentFlow-written config from Claude CLI files.
// This cleans the "model" field from ~/.claude.json.
// Best-effort: errors are silently ignored so uninstall always succeeds.
func (i *Installer) cleanClaudeBootstrapConfig() {
	configPath := filepath.Join(i.HomeDir, ".claude.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return
	}
	var settings map[string]any
	if json.Unmarshal(data, &settings) != nil || settings == nil {
		return
	}
	delete(settings, "model")
	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return
	}
	_ = config.SafeWrite(configPath, append(out, '\n'), 0o644)
}

// ReadCodexAuthKey reads the API key token from ~/.codex/auth.json.
func (i *Installer) ReadCodexAuthKey() string {
	authPath := filepath.Join(i.HomeDir, ".codex", "auth.json")
	data, err := os.ReadFile(authPath)
	if err != nil {
		return ""
	}
	var auth map[string]any
	if json.Unmarshal(data, &auth) != nil {
		return ""
	}
	if token, ok := auth["token"].(string); ok {
		return token
	}
	return ""
}

// ReadCodexConfigField reads a top-level TOML field from ~/.codex/config.toml.
// For "base_url", it also checks the active model_provider section.
func (i *Installer) ReadCodexConfigField(field string) string {
	configPath := filepath.Join(i.HomeDir, ".codex", "config.toml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return ""
	}
	text := string(data)

	// For base_url, check the model_providers section first (current approach),
	// then fall back to legacy openai_base_url top-level field.
	if field == "base_url" {
		// Check legacy openai_base_url (older AgentFlow versions).
		reTopLevel := regexp.MustCompile(`(?m)^openai_base_url\s*=\s*"([^"]*)"`)
		if m := reTopLevel.FindStringSubmatch(text); len(m) > 1 {
			return m[1]
		}
		// Find active model_provider name.
		reProvider := regexp.MustCompile(`(?m)^model_provider\s*=\s*"([^"]*)"`)
		if m := reProvider.FindStringSubmatch(text); len(m) > 1 {
			providerName := m[1]
			// Find base_url inside [model_providers.<name>] section.
			sectionRe := regexp.MustCompile(`(?ms)\[model_providers\.` + regexp.QuoteMeta(providerName) + `\].*?base_url\s*=\s*"([^"]*)"`)
			if sm := sectionRe.FindStringSubmatch(text); len(sm) > 1 {
				return sm[1]
			}
		}
	}

	// Read top-level field.
	re := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(field) + `\s*=\s*"([^"]*)"`)
	if m := re.FindStringSubmatch(text); len(m) > 1 {
		return m[1]
	}
	return ""
}

// ReadCLIConfigModel reads the current model from the CLI's config file.
func (i *Installer) ReadCLIConfigModel(targetName string) string {
	switch targetName {
	case "codex":
		return i.ReadCodexConfigField("model")
	case "claude":
		configPath := filepath.Join(i.HomeDir, ".claude.json")
		data, err := os.ReadFile(configPath)
		if err != nil {
			return ""
		}
		var settings map[string]any
		if json.Unmarshal(data, &settings) != nil {
			return ""
		}
		if model, ok := settings["model"].(string); ok {
			return model
		}
		return ""
	default:
		return ""
	}
}

func (i *Installer) Install(targetName, profile, lang string) error {
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
	if lang == "" {
		lang = config.DefaultLang
	}

	// Deploy full rules + modules to ~/.agentflow/ first.
	if err := i.DeployGlobalRulesDir(targetName, profile, lang); err != nil {
		return err
	}

	cliDir := filepath.Join(i.HomeDir, target.Dir)
	if err := os.MkdirAll(cliDir, 0o755); err != nil {
		return fmt.Errorf(i.Catalog.Msg("无法创建目标目录: %s", "failed to create target directory: %s"), cliDir)
	}

	// Write full compiled rules to the CLI entry file.
	if err := i.deployRulesFile(target, profile, lang); err != nil {
		return err
	}
	// Module dir is now in ~/.agentflow/agentflow/, skip CLI-local deploy.
	if err := i.deploySkill(target, lang); err != nil {
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

func (i *Installer) InstallAll(profile, lang string) (int, error) {
	success := 0
	for _, name := range i.AgentFlowInstallableTargets() {
		if err := i.Install(name, profile, lang); err == nil {
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
	// Clean legacy CLI-local module dir if it still exists from older installs.
	_ = config.SafeRemove(filepath.Join(cliDir, config.PluginDirName))
	if err := config.SafeRemove(filepath.Join(cliDir, "skills", "agentflow")); err != nil {
		return err
	}

	if target.Name == "codex" {
		if err := i.cleanCodexAgents(cliDir); err != nil {
			return err
		}
		// Clean AgentFlow-written CLI config (auth.json, config.toml fields).
		i.cleanCodexBootstrapConfig()
	}
	if target.Name == "claude" {
		if err := i.cleanClaudeHooks(filepath.Join(cliDir, "settings.json")); err != nil {
			return err
		}
		// Clean AgentFlow-written Claude config (.claude.json model).
		i.cleanClaudeBootstrapConfig()
	}
	if target.Name == "kiro" {
		if err := i.cleanKiroAgent(cliDir); err != nil {
			return err
		}
	}

	// Clean AgentFlow env vars from shell RC (affects all targets).
	_ = i.CleanEnvConfig()

	return nil
}

func (i *Installer) deployKiroAgent(target targets.Target, profile string) error {
	cliDir := filepath.Join(i.HomeDir, target.Dir)

	// Kiro custom agents can reference a local prompt file.
	promptContent, err := i.buildRulesContent(target.Name, profile, config.DefaultLang)
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
	// Clean global rules when all targets are uninstalled.
	i.cleanGlobalRules()
	return len(installed), nil
}

func (i *Installer) DetectInstalledCLIs() []string {
	result := make([]string, 0)
	for _, status := range i.DetectTargetStatuses() {
		if status.CLIInstalled {
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
		rulesFile := filepath.Join(cliDir, target.RulesFile)
		// Detection is based on the AgentFlow-managed rules file in the CLI dir.
		// Module dir is now centralized in ~/.agentflow/.
		if config.IsAgentFlowFile(rulesFile) {
			result = append(result, name)
		}
	}
	return result
}

func (i *Installer) Clean() (int, error) {
	cleaned := 0
	globalModuleDir := filepath.Join(i.HomeDir, config.GlobalRulesDir, config.PluginDirName)
	for _, cacheDir := range []string{
		filepath.Join(globalModuleDir, "__pycache__"),
		filepath.Join(globalModuleDir, ".cache"),
	} {
		if _, err := os.Stat(cacheDir); err == nil {
			if err := config.SafeRemove(cacheDir); err != nil {
				return cleaned, err
			}
			cleaned++
		}
	}
	return cleaned, nil
}

// Status badge styles.
var (
	statusOK   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42"))  // green
	statusAF   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214")) // orange/yellow
	statusCLI  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("81"))  // cyan
	statusNone = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))            // dark gray
	legendDim  = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))            // muted
)

func (i *Installer) StatusLines() []string {
	lines := []string{
		i.Catalog.Msg("CLI 状态:", "CLI status:"),
	}
	for _, status := range i.DetectTargetStatuses() {
		switch {
		case status.CLIInstalled && status.AgentFlowInstalled:
			lines = append(lines, fmt.Sprintf("  %s %s", statusOK.Render("✔ OK"), status.Target.Name))
		case status.CLIInstalled:
			lines = append(lines, fmt.Sprintf("  %s %s", statusCLI.Render("● CLI"), status.Target.Name))
		case status.AgentFlowInstalled:
			lines = append(lines, fmt.Sprintf("  %s %s",
				statusAF.Render("▲ "+i.Catalog.Msg("仅AF", "AF only")),
				status.Target.Name))
		default:
			lines = append(lines, fmt.Sprintf("  %s %s", statusNone.Render("○ --"), status.Target.Name))
		}
	}
	lines = append(lines, "")
	lines = append(lines, legendDim.Render(
		i.Catalog.Msg(
			"✔ CLI+AF 就绪 │ ● 仅CLI │ ▲ 仅AgentFlow规则 │ ○ 未安装",
			"✔ CLI+AF ready │ ● CLI only │ ▲ AF rules only │ ○ not installed",
		),
	))
	return lines
}

func (i *Installer) deployRulesFile(target targets.Target, profile, lang string) error {
	rulesPath := filepath.Join(i.HomeDir, target.Dir, target.RulesFile)
	if _, err := os.Stat(rulesPath); err == nil && !config.IsAgentFlowFile(rulesPath) {
		if _, err := config.BackupUserFile(rulesPath); err != nil {
			return err
		}
	}

	moduleDir := filepath.Join(i.HomeDir, config.GlobalRulesDir, config.PluginDirName)
	moduleLinkBase, err := filepath.Rel(filepath.Dir(rulesPath), moduleDir)
	if err != nil {
		return err
	}

	rendered, err := i.buildEntryRulesContent(target.Name, profile, lang, moduleLinkBase)
	if err != nil {
		return err
	}
	return config.SafeWrite(rulesPath, []byte(rendered), 0o644)
}

func (i *Installer) buildEntryRulesContent(targetName, profile, lang, moduleLinkBase string) (string, error) {
	rendered, err := i.buildRulesContent(targetName, profile, lang)
	if err != nil {
		return "", err
	}
	return config.RewriteEmbeddedModuleLinks(rendered, moduleLinkBase), nil
}

func (i *Installer) buildRulesContent(targetName, profile, lang string) (string, error) {
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
	if lang == "" {
		lang = config.DefaultLang
	}

	content, err := readLangAsset(lang, "agentflow/_AGENTS.md", "AGENTS.md")
	if err != nil {
		return "", err
	}

	rendered := string(content)
	if !strings.Contains(rendered, config.Marker) {
		rendered = "<!-- AGENTFLOW_ROUTER: v1.0.0 -->\n" + rendered
	}
	rendered = strings.ReplaceAll(rendered, "{TARGET_CLI}", target.DisplayName)
	rendered = strings.ReplaceAll(rendered, "{HOOKS_SUMMARY}", target.HooksSummary)

	// Strip sections beyond the selected profile instead of appending
	// full module content. The _AGENTS.md already contains summaries
	// with lazy-loading file references for each profile section.
	rendered = stripBeyondProfile(rendered, profile)

	return strings.TrimRight(rendered, "\n") + "\n", nil
}

// stripBeyondProfile removes sections of _AGENTS.md that are beyond
// the selected profile level:
//   - lite: strips everything from <!-- PROFILE:standard --> onward
//   - standard: strips everything from <!-- PROFILE:full --> onward
//   - full: keeps all content (summaries + references, no inlined modules)
func stripBeyondProfile(text, profile string) string {
	switch profile {
	case "lite":
		if idx := strings.Index(text, "<!-- PROFILE:standard"); idx > 0 {
			// Keep the trailing footer if present
			footer := "\n---\n\n> **AgentFlow** — Go beyond analysis; keep working until implementation and verification are complete.\n"
			return strings.TrimRight(text[:idx], "\n") + "\n" + footer
		}
	case "standard":
		if idx := strings.Index(text, "<!-- PROFILE:full"); idx > 0 {
			footer := "\n---\n\n> **AgentFlow** — Go beyond analysis; keep working until implementation and verification are complete.\n"
			return strings.TrimRight(text[:idx], "\n") + "\n" + footer
		}
	}
	// "full" or no markers found: keep everything as-is.
	return text
}

// readLangAsset reads an embedded asset from the language-specific directory.
// For example, readLangAsset("en", "agentflow/_AGENTS.md") tries "agentflow/en/_AGENTS.md"
// first, falling back to "agentflow/_AGENTS.md" (root fallback).
func readLangAsset(lang string, paths ...string) ([]byte, error) {
	var langPaths []string
	for _, p := range paths {
		// Insert lang dir: "agentflow/_AGENTS.md" → "agentflow/locales/{lang}/_AGENTS.md"
		if strings.HasPrefix(p, "agentflow/") {
			rest := strings.TrimPrefix(p, "agentflow/")
			langPaths = append(langPaths, "agentflow/locales/"+lang+"/"+rest)
		} else {
			langPaths = append(langPaths, p)
		}
	}
	langPaths = append(langPaths, paths...) // fallback to root
	return readAssetWithFallback(langPaths...)
}

// DeployGlobalRulesDir deploys the full compiled rules and module files
// to ~/.agentflow/. Both global and project-level installs call this
// so the centralized rules directory always exists.
func (i *Installer) DeployGlobalRulesDir(targetName, profile, lang string) error {
	globalDir := filepath.Join(i.HomeDir, config.GlobalRulesDir)
	if err := os.MkdirAll(globalDir, 0o755); err != nil {
		return err
	}

	// Write full compiled rules to ~/.agentflow/AGENTS.md.
	rendered, err := i.buildRulesContent(targetName, profile, lang)
	if err != nil {
		return err
	}
	rulesPath := filepath.Join(globalDir, config.GlobalRulesFile)
	if err := config.SafeWrite(rulesPath, []byte(rendered), 0o644); err != nil {
		return err
	}

	// Deploy module files to ~/.agentflow/agentflow/.
	return i.deployModuleDirTo(filepath.Join(globalDir, config.PluginDirName), lang)
}

// cleanGlobalRules removes the centralized rules and module dir from
// ~/.agentflow/. Called only by UninstallAll().
func (i *Installer) cleanGlobalRules() {
	globalDir := filepath.Join(i.HomeDir, config.GlobalRulesDir)
	rulesPath := filepath.Join(globalDir, config.GlobalRulesFile)
	if config.IsAgentFlowFile(rulesPath) {
		_ = config.SafeRemove(rulesPath)
	}
	_ = config.SafeRemove(filepath.Join(globalDir, config.PluginDirName))
}

// deployModuleDirTo deploys the embedded agentflow module directory
// to the specified base directory, filtering by language.
// It first deploys shared files from agentflow/ root (skipping zh/, en/ dirs),
// then overlays language-specific files from agentflow/{lang}/.
func (i *Installer) deployModuleDirTo(baseDir, lang string) error {
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

	// Pass 1: Deploy shared (non-language) files, skipping zh/ and en/ dirs.
	if err := fs.WalkDir(moduleFS, ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == "." {
			return nil
		}
		// Skip language directories entirely.
		if entry.IsDir() && path == "locales" {
			return fs.SkipDir
		}
		// Skip root-level _AGENTS.md / _SKILL.md (deployed separately via buildRulesContent).
		if !entry.IsDir() && (path == "_AGENTS.md" || path == "_SKILL.md") {
			return nil
		}
		if entry.IsDir() {
			return os.MkdirAll(filepath.Join(baseDir, filepath.FromSlash(path)), 0o755)
		}
		content, err := fs.ReadFile(moduleFS, path)
		if err != nil {
			return err
		}
		return config.SafeWrite(filepath.Join(baseDir, filepath.FromSlash(path)), content, 0o644)
	}); err != nil {
		return err
	}

	// Pass 2: Overlay language-specific files from agentflow/locales/{lang}/.
	langFS, err := agentflowassets.Sub("agentflow/locales/" + lang)
	if err != nil {
		// No language dir found — not an error, just skip.
		return nil
	}
	return fs.WalkDir(langFS, ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == "." {
			return nil
		}
		// Skip root _AGENTS.md / _SKILL.md (these are handled by buildRulesContent/deploySkill).
		if !entry.IsDir() && (path == "_AGENTS.md" || path == "_SKILL.md") {
			return nil
		}
		if entry.IsDir() {
			return os.MkdirAll(filepath.Join(baseDir, filepath.FromSlash(path)), 0o755)
		}
		content, err := fs.ReadFile(langFS, path)
		if err != nil {
			return err
		}
		return config.SafeWrite(filepath.Join(baseDir, filepath.FromSlash(path)), content, 0o644)
	})
}

func (i *Installer) deploySkill(target targets.Target, lang string) error {
	content, err := readLangAsset(lang, "agentflow/_SKILL.md", "SKILL.md")
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

// insertTopLevel inserts a top-level TOML key=value line before the first
// section header "[". This prevents the field from landing inside an
// existing section like [features] when appended to the end of the file.
func insertTopLevel(text, line string) string {
	// Find the first section header.
	idx := strings.Index(text, "\n[")
	if idx >= 0 {
		// Insert before the newline that starts the section.
		return text[:idx] + "\n" + line + text[idx:]
	}
	// No sections: just append.
	return strings.TrimRight(text, "\n") + "\n" + line + "\n"
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
