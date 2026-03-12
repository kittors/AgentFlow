package projectroot

import (
	"path/filepath"
	"strings"
)

const (
	AgentFlowDir = ".agentflow"
	KBDir        = "kb"
	SessionsDir  = "sessions"
)

type Paths struct {
	ProjectRoot string
	AgentFlow   string
	KB          string
	Modules     string
	Plan        string
	Graph       string
	Conventions string
	Archive     string
	Sessions    string
}

func NewPaths(root string) Paths {
	agentflow := filepath.Join(root, AgentFlowDir)
	kb := filepath.Join(agentflow, KBDir)
	return Paths{
		ProjectRoot: root,
		AgentFlow:   agentflow,
		KB:          kb,
		Modules:     filepath.Join(kb, "modules"),
		Plan:        filepath.Join(kb, "plan"),
		Graph:       filepath.Join(kb, "graph"),
		Conventions: filepath.Join(kb, "conventions"),
		Archive:     filepath.Join(kb, "archive"),
		Sessions:    filepath.Join(agentflow, SessionsDir),
	}
}

func ProjectName(root string) string {
	cleaned := filepath.Clean(root)
	return filepath.Base(cleaned)
}

func DefaultScanDirs(root string) []string {
	return []string{"src", "lib", "app", ProjectName(root)}
}

func IsHiddenName(name string) bool {
	return strings.HasPrefix(name, ".")
}
