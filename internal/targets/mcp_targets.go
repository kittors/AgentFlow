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
}

func LookupMCP(name string) (Target, bool) {
	target, ok := MCPAll[name]
	return target, ok
}

func MCPNames() []string {
	return []string{"codex", "claude", "gemini", "qwen", "kiro", "cursor", "windsurf"}
}

func SortedMCPTargetNames() []string {
	names := append([]string(nil), MCPNames()...)
	sort.Strings(names)
	return names
}
