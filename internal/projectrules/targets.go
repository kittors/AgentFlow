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

var All = map[string]Target{
	"codex":  {Name: "codex", DisplayName: "Codex CLI", Kind: KindCLI},
	"claude": {Name: "claude", DisplayName: "Claude Code", Kind: KindCLI},
}

func Lookup(name string) (Target, bool) {
	t, ok := All[name]
	return t, ok
}

func Names() []string {
	return []string{
		"codex",
		"claude",
	}
}

func SortedNames() []string {
	names := append([]string(nil), Names()...)
	sort.Strings(names)
	return names
}
