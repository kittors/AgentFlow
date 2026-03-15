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
		APIKeyEnv:             "OPENAI_API_KEY",
		BaseURLEnv:            "OPENAI_BASE_URL",
		HasConfigFile:         true,
		Models: []ModelOption{
			{Value: "o4-mini", Label: "o4-mini (推荐/Recommended)", Default: true},
			{Value: "gpt-4.1", Label: "GPT-4.1"},
			{Value: "gpt-4.1-mini", Label: "GPT-4.1 Mini"},
			{Value: "gpt-4o", Label: "GPT-4o"},
			{Value: "o3", Label: "o3"},
			{Value: "o3-mini", Label: "o3-mini"},
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
		RecommendWSLOnWindows: true,
		APIKeyEnv:             "ANTHROPIC_API_KEY",
		BaseURLEnv:            "ANTHROPIC_BASE_URL",
		ModelEnv:              "ANTHROPIC_MODEL",
		Models: []ModelOption{
			{Value: "claude-sonnet-4-20250514", Label: "Claude Sonnet 4 (推荐/Recommended)", Default: true},
			{Value: "claude-opus-4-20250514", Label: "Claude Opus 4"},
			{Value: "claude-haiku-3.5-20241022", Label: "Claude Haiku 3.5"},
		},
	},
	"gemini": {
		Name:                  "gemini",
		DisplayName:           "Gemini CLI",
		Dir:                   ".gemini",
		RulesFile:             "GEMINI.md",
		HooksSummary:          "当前不支持 Hooks; 所有 Hook 功能自动降级，不影响核心工作流。",
		Command:               "gemini",
		NPMPackage:            "@google/gemini-cli",
		MinNodeMajor:          20,
		DocsURL:               "https://google-gemini.github.io/gemini-cli/docs/get-started/",
		BootstrapSupported:    true,
		RecommendWSLOnWindows: true,
		APIKeyEnv:             "GEMINI_API_KEY",
		BaseURLEnv:            "GEMINI_API_BASE_URL",
		ModelEnv:              "GEMINI_MODEL",
		Models: []ModelOption{
			{Value: "gemini-2.5-pro", Label: "Gemini 2.5 Pro (推荐/Recommended)", Default: true},
			{Value: "gemini-2.5-flash", Label: "Gemini 2.5 Flash"},
		},
	},
	"qwen": {
		Name:                  "qwen",
		DisplayName:           "Qwen CLI",
		Dir:                   ".qwen",
		RulesFile:             "QWEN.md",
		HooksSummary:          "当前不支持 Hooks; 所有 Hook 功能自动降级，不影响核心工作流。",
		Command:               "qwen",
		NPMPackage:            "@qwen-code/qwen-code",
		MinNodeMajor:          20,
		DocsURL:               "https://qwen-code.github.io/qwen-code-docs/",
		BootstrapSupported:    true,
		RecommendWSLOnWindows: true,
		APIKeyEnv:             "DASHSCOPE_API_KEY",
		BaseURLEnv:            "DASHSCOPE_BASE_URL",
		ModelEnv:              "DASHSCOPE_MODEL",
		Models: []ModelOption{
			{Value: "qwen3-coder", Label: "Qwen3 Coder (推荐/Recommended)", Default: true},
			{Value: "qwen-max", Label: "Qwen Max"},
			{Value: "qwen-plus", Label: "Qwen Plus"},
		},
	},
	"kiro": {
		Name:                  "kiro",
		DisplayName:           "Kiro CLI",
		Dir:                   ".kiro",
		RulesFile:             "KIRO.md",
		HooksSummary:          "当前不支持 Hooks; 所有 Hook 功能自动降级，不影响核心工作流。",
		Command:               "kiro-cli",
		DocsURL:               "https://kiro.dev/cli/",
		BootstrapSupported:    true,
		RecommendWSLOnWindows: true,
	},
	"grok": {
		Name:         "grok",
		DisplayName:  "Grok CLI",
		Dir:          ".grok",
		RulesFile:    "GROK.md",
		HooksSummary: "当前不支持 Hooks; 所有 Hook 功能自动降级，不影响核心工作流。",
	},
	"opencode": {
		Name:         "opencode",
		DisplayName:  "OpenCode",
		Dir:          ".config/opencode",
		RulesFile:    "AGENTS.md",
		HooksSummary: "当前不支持 Hooks; 所有 Hook 功能自动降级，不影响核心工作流。",
	},
}

var SupportedTargets = All

func Names() []string {
	return []string{"codex", "claude", "gemini", "qwen", "kiro", "grok", "opencode"}
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
	return []string{"codex", "claude", "gemini", "qwen", "kiro"}
}
