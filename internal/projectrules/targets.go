package projectrules

import "sort"

const (
	KindCLI = "cli"
	KindIDE = "ide"
)

type Target struct {
	Name        string
	DisplayName string
	Kind        string
}

var all = map[string]Target{
	"codex":          {Name: "codex", DisplayName: "Codex CLI", Kind: KindCLI},
	"claude":         {Name: "claude", DisplayName: "Claude Code", Kind: KindCLI},
	"gemini":         {Name: "gemini", DisplayName: "Gemini CLI", Kind: KindCLI},
	"qwen":           {Name: "qwen", DisplayName: "Qwen CLI", Kind: KindCLI},
	"kiro":           {Name: "kiro", DisplayName: "Kiro (CLI/IDE)", Kind: KindCLI},
	"cursor":         {Name: "cursor", DisplayName: "Cursor", Kind: KindIDE},
	"windsurf":       {Name: "windsurf", DisplayName: "Windsurf", Kind: KindIDE},
	"trae":           {Name: "trae", DisplayName: "Trae", Kind: KindIDE},
	"vscode-copilot": {Name: "vscode-copilot", DisplayName: "VS Code Copilot", Kind: KindIDE},
	"cline":          {Name: "cline", DisplayName: "Cline (VS Code)", Kind: KindIDE},
	"antigravity":    {Name: "antigravity", DisplayName: "Antigravity", Kind: KindIDE},
}

func Lookup(name string) (Target, bool) {
	t, ok := all[name]
	return t, ok
}

func Names() []string {
	return []string{
		"codex",
		"claude",
		"gemini",
		"qwen",
		"kiro",
		"cursor",
		"windsurf",
		"trae",
		"vscode-copilot",
		"cline",
		"antigravity",
	}
}

func SortedNames() []string {
	names := append([]string(nil), Names()...)
	sort.Strings(names)
	return names
}
