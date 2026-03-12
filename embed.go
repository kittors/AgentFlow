package agentflowassets

import (
	"embed"
	"io/fs"
	"path"
)

// Files keeps the shipped AgentFlow static assets inside the Go binary.
//
//go:embed AGENTS.md SKILL.md all:agentflow
var Files embed.FS

func ReadFile(name string) ([]byte, error) {
	return Files.ReadFile(name)
}

func Sub(dir string) (fs.FS, error) {
	return fs.Sub(Files, path.Clean(dir))
}
