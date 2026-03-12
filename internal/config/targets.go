package config

import "sort"

// Target describes the install layout for a supported CLI.
type Target struct {
	Name      string
	Dir       string
	RulesFile string
}

var targets = map[string]Target{
	"codex":    {Name: "codex", Dir: ".codex", RulesFile: "AGENTS.md"},
	"claude":   {Name: "claude", Dir: ".claude", RulesFile: "CLAUDE.md"},
	"gemini":   {Name: "gemini", Dir: ".gemini", RulesFile: "GEMINI.md"},
	"grok":     {Name: "grok", Dir: ".grok", RulesFile: "GROK.md"},
	"opencode": {Name: "opencode", Dir: ".config/opencode", RulesFile: "AGENTS.md"},
	"qwen":     {Name: "qwen", Dir: ".qwen", RulesFile: "QWEN.md"},
}

// TargetByName returns a configured CLI target.
func TargetByName(name string) (Target, bool) {
	target, ok := targets[name]
	return target, ok
}

// Targets returns all known targets in stable name order.
func Targets() []Target {
	names := make([]string, 0, len(targets))
	for name := range targets {
		names = append(names, name)
	}
	sort.Strings(names)

	out := make([]Target, 0, len(names))
	for _, name := range names {
		out = append(out, targets[name])
	}
	return out
}
