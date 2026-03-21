package targets

import "sort"

// ModelOption describes a selectable model for a CLI target.
type ModelOption struct {
	Value   string // model identifier, e.g. "claude-sonnet-4-20250514"
	Label   string // human-readable label, e.g. "Claude Sonnet 4"
	Default bool   // whether this is the recommended default
}

type Target struct {
	Name                  string
	DisplayName           string
	Dir                   string
	RulesFile             string
	HooksSummary          string
	Command               string
	NPMPackage            string
	MinNodeMajor          int
	DocsURL               string
	BootstrapSupported    bool
	RecommendWSLOnWindows bool
	APIKeyEnv             string        // Environment variable for API key (empty = no config needed)
	BaseURLEnv            string        // Environment variable for custom base URL (empty = not configurable)
	ModelEnv              string        // Environment variable for model selection (empty = not configurable via env)
	Models                []ModelOption // Available models for selection
	// HasConfigFile indicates this CLI stores settings in a local config file
	// (e.g. ~/.codex/config.json) and needs special handling beyond env vars.
	HasConfigFile bool
	// NativeInstallPS1 is the PowerShell install command for Windows (no npm needed).
	// Example: "irm https://claude.ai/install.ps1 | iex"
	NativeInstallPS1 string
	// NativeInstallShell is the shell install command for macOS/Linux (no npm needed).
	// Example: "curl -fsSL https://claude.ai/install.sh | bash"
	NativeInstallShell string
}

var All = map[string]Target{
	"codex": {
		Name:                  "codex",
		DisplayName:           "Codex CLI",
		Dir:                   ".codex",
		RulesFile:             "AGENTS.md",
		HooksSummary:          "仅支持 Notification 事件; 其他 Hook 不可用时功能自动降级，不影响核心工作流。",
		Command:               "codex",
		NPMPackage:            "@openai/codex",
		DocsURL:               "https://help.openai.com/en/articles/11096431-openai-codex-ci-getting-started",
		BootstrapSupported:    true,
		RecommendWSLOnWindows: true,
		// Codex uses model_provider + [model_providers.agentflow] with
		// env_key = "OPENAI_API_KEY" for custom base URLs. The API key
		// is written to shell RC as OPENAI_API_KEY env var. We use an
		// internal marker for BaseURLEnv so the UI shows a Base URL
		// input, but the value goes to WriteCodexConfig (config.toml).
		APIKeyEnv:     "OPENAI_API_KEY",
		BaseURLEnv:    "__CODEX_BASE_URL__",
		ModelEnv:      "__CODEX_MODEL__",
		HasConfigFile: true,
		Models: []ModelOption{
			{Value: "gpt-5.2", Label: "GPT-5.2 (推荐/Recommended)", Default: true},
			{Value: "gpt-5.1-codex-max", Label: "GPT-5.1 Codex Max"},
			{Value: "gpt-5.1-codex-mini", Label: "GPT-5.1 Codex Mini"},
			{Value: "gpt-5.3-codex", Label: "GPT-5.3 Codex"},
			{Value: "gpt-5.4", Label: "GPT-5.4"},
		},
	},
	"claude": {
		Name:                  "claude",
		DisplayName:           "Claude Code",
		Dir:                   ".claude",
		RulesFile:             "CLAUDE.md",
		HooksSummary:          "支持全部 6 种 Hook 事件（PreToolCall, PostToolCall, PostMessage, Notification, SessionStart, SessionEnd）; Hooks 不可用时功能降级但不影响核心工作流。",
		Command:               "claude",
		NPMPackage:            "@anthropic-ai/claude-code",
		MinNodeMajor:          18,
		DocsURL:               "https://docs.anthropic.com/en/docs/claude-code/getting-started",
		BootstrapSupported:    true,
		RecommendWSLOnWindows: false, // Native install is now preferred on Windows.
		// Claude Code reads config from ~/.claude/settings.json (env section).
		// We use internal markers so the UI shows input fields, but values
		// are written to settings.json via WriteClaudeSettings (not shell RC).
		APIKeyEnv:          "__CLAUDE_API_KEY__",
		BaseURLEnv:         "__CLAUDE_BASE_URL__",
		ModelEnv:           "__CLAUDE_MODEL__",
		HasConfigFile:      true,
		NativeInstallPS1:   "irm https://claude.ai/install.ps1 | iex",
		NativeInstallShell: "curl -fsSL https://claude.ai/install.sh | bash",
		Models: []ModelOption{
			{Value: "claude-opus-4-6", Label: "Claude Opus 4.6 (推荐/Recommended)", Default: true},
			{Value: "claude-haiku-4-5-20251001", Label: "Claude Haiku 4.5"},
			{Value: "claude-opus-4-5-20251101", Label: "Claude Opus 4.5"},
			{Value: "claude-sonnet-4-5-20250929", Label: "Claude Sonnet 4.5"},
			{Value: "claude-sonnet-4-6", Label: "Claude Sonnet 4.6"},
		},
	},
}

var SupportedTargets = All

func Names() []string {
	return []string{"codex", "claude"}
}

func Lookup(name string) (Target, bool) {
	target, ok := All[name]
	return target, ok
}

func SortedTargetNames() []string {
	names := append([]string(nil), Names()...)
	sort.Strings(names)
	return names
}

func BootstrapNames() []string {
	return []string{"codex", "claude"}
}
