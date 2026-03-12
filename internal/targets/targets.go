package targets

import "sort"

type Target struct {
	Name         string
	DisplayName  string
	Dir          string
	RulesFile    string
	HooksSummary string
}

var All = map[string]Target{
	"codex": {
		Name:         "codex",
		DisplayName:  "Codex CLI",
		Dir:          ".codex",
		RulesFile:    "AGENTS.md",
		HooksSummary: "仅支持 Notification 事件; 其他 Hook 不可用时功能自动降级，不影响核心工作流。",
	},
	"claude": {
		Name:         "claude",
		DisplayName:  "Claude Code",
		Dir:          ".claude",
		RulesFile:    "CLAUDE.md",
		HooksSummary: "支持全部 6 种 Hook 事件（PreToolCall, PostToolCall, PostMessage, Notification, SessionStart, SessionEnd）; Hooks 不可用时功能降级但不影响核心工作流。",
	},
	"gemini": {
		Name:         "gemini",
		DisplayName:  "Gemini CLI",
		Dir:          ".gemini",
		RulesFile:    "GEMINI.md",
		HooksSummary: "当前不支持 Hooks; 所有 Hook 功能自动降级，不影响核心工作流。",
	},
	"qwen": {
		Name:         "qwen",
		DisplayName:  "Qwen CLI",
		Dir:          ".qwen",
		RulesFile:    "QWEN.md",
		HooksSummary: "当前不支持 Hooks; 所有 Hook 功能自动降级，不影响核心工作流。",
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
	return []string{"codex", "claude", "gemini", "qwen", "grok", "opencode"}
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
