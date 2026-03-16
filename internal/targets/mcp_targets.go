package targets

import "sort"

var MCPAll = map[string]Target{
	"codex":  All["codex"],
	"claude": All["claude"],
}

func LookupMCP(name string) (Target, bool) {
	target, ok := MCPAll[name]
	return target, ok
}

func MCPNames() []string {
	return []string{"codex", "claude"}
}

func SortedMCPTargetNames() []string {
	names := append([]string(nil), MCPNames()...)
	sort.Strings(names)
	return names
}
