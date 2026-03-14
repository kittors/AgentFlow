package targets

import "sort"

var MCPAll = map[string]Target{
	"codex":  All["codex"],
	"claude": All["claude"],
	"gemini": All["gemini"],
	"qwen":   All["qwen"],
	"kiro":   All["kiro"],
	"cursor": {
		Name:        "cursor",
		DisplayName: "Cursor",
		Dir:         ".cursor",
	},
	"windsurf": {
		Name:        "windsurf",
		DisplayName: "Windsurf",
		Dir:         ".codeium/windsurf",
	},
	"trae": {
		Name:        "trae",
		DisplayName: "Trae",
		Dir:         ".trae",
	},
	"vscode-copilot": {
		Name:        "vscode-copilot",
		DisplayName: "VS Code Copilot",
		Dir:         ".vscode",
	},
	"cline": {
		Name:        "cline",
		DisplayName: "Cline (VS Code)",
		Dir:         ".cline",
	},
	"jetbrains": {
		Name:        "jetbrains",
		DisplayName: "JetBrains IDE",
		Dir:         ".jetbrains",
	},
	"antigravity": {
		Name:        "antigravity",
		DisplayName: "Antigravity",
		Dir:         ".gemini",
	},
}

func LookupMCP(name string) (Target, bool) {
	target, ok := MCPAll[name]
	return target, ok
}

func MCPNames() []string {
	return []string{
		"codex", "claude", "gemini", "qwen", "kiro",
		"cursor", "windsurf",
		"trae", "vscode-copilot", "cline", "jetbrains", "antigravity",
	}
}

func SortedMCPTargetNames() []string {
	names := append([]string(nil), MCPNames()...)
	sort.Strings(names)
	return names
}
