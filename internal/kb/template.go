package kb

import (
	"os"
	"path/filepath"

	"github.com/kittors/AgentFlow/internal/projectroot"
)

type TemplateInitSummary struct {
	ProjectType  string   `json:"project_type"`
	FilesCreated []string `json:"files_created"`
}

var indicators = map[string][]string{
	"frontend": {
		"package.json",
		"vite.config.ts",
		"next.config.js",
		"next.config.ts",
		"nuxt.config.ts",
		"angular.json",
		"svelte.config.js",
	},
	"backend": {
		"requirements.txt",
		"Pipfile",
		"go.mod",
		"Cargo.toml",
		"pom.xml",
		"build.gradle",
	},
	"python": {
		"pyproject.toml",
		"setup.py",
		"setup.cfg",
	},
}

func DetectProjectType(root string) string {
	hasFrontend := containsAny(root, indicators["frontend"])
	hasBackend := containsAny(root, indicators["backend"])
	hasPython := containsAny(root, indicators["python"])

	switch {
	case hasFrontend && hasBackend:
		return "fullstack"
	case hasFrontend:
		return "frontend"
	case hasPython:
		return "python"
	case hasBackend:
		return "backend"
	default:
		return "python"
	}
}

func InitPaths(root string) (TemplateInitSummary, error) {
	paths := projectroot.NewPaths(root)
	dirs := []string{
		paths.Modules,
		paths.Plan,
		paths.Graph,
		paths.Conventions,
		paths.Archive,
		paths.Sessions,
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return TemplateInitSummary{}, err
		}
	}

	return TemplateInitSummary{
		ProjectType:  DetectProjectType(root),
		FilesCreated: nil,
	}, nil
}

func containsAny(root string, names []string) bool {
	for _, name := range names {
		if _, err := os.Stat(filepath.Join(root, name)); err == nil {
			return true
		}
	}
	return false
}
