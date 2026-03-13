package targets

import "sort"

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
