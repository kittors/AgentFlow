package kb

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kittors/AgentFlow/internal/projectroot"
)

type Module struct {
	Name      string
	Path      string
	FileCount int
	Files     []string
}

type SyncSummary struct {
	ModulesFound int `json:"modules_found"`
	FilesWritten int `json:"files_written"`
}

func ScanModules(root string, sourceDirs []string) ([]Module, error) {
	if len(sourceDirs) == 0 {
		sourceDirs = projectroot.DefaultScanDirs(root)
	}

	modules := make([]Module, 0)
	for _, dir := range sourceDirs {
		src := filepath.Join(root, dir)
		info, err := os.Stat(src)
		if err != nil || !info.IsDir() {
			continue
		}

		children, err := os.ReadDir(src)
		if err != nil {
			return nil, err
		}
		for _, child := range children {
			if !child.IsDir() || strings.HasPrefix(child.Name(), ".") || strings.HasPrefix(child.Name(), "_") {
				continue
			}

			files := make([]string, 0)
			err = filepath.WalkDir(filepath.Join(src, child.Name()), func(path string, entry os.DirEntry, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}
				if entry.IsDir() {
					return nil
				}
				ext := filepath.Ext(entry.Name())
				if ext != ".py" && ext != ".ts" {
					return nil
				}
				rel, err := filepath.Rel(root, path)
				if err != nil {
					return err
				}
				files = append(files, filepath.ToSlash(rel))
				return nil
			})
			if err != nil {
				return nil, err
			}
			if len(files) == 0 {
				continue
			}

			sort.Strings(files)
			rel, err := filepath.Rel(root, filepath.Join(src, child.Name()))
			if err != nil {
				return nil, err
			}
			modules = append(modules, Module{
				Name:      child.Name(),
				Path:      filepath.ToSlash(rel),
				FileCount: len(files),
				Files:     files,
			})
		}
	}

	sort.Slice(modules, func(i, j int) bool {
		return modules[i].Name < modules[j].Name
	})
	return modules, nil
}

func GenerateModuleIndex(modules []Module) string {
	lines := []string{"# Module Index", ""}
	for _, module := range modules {
		lines = append(lines,
			fmt.Sprintf("## %s", module.Name),
			fmt.Sprintf("- Path: `%s`", module.Path),
			fmt.Sprintf("- Files: %d", module.FileCount),
			"",
		)
	}
	return strings.Join(lines, "\n")
}

func SyncModules(root string, sourceDirs []string) (SyncSummary, error) {
	paths := projectroot.NewPaths(root)
	if err := os.MkdirAll(paths.Modules, 0o755); err != nil {
		return SyncSummary{}, err
	}

	modules, err := ScanModules(root, sourceDirs)
	if err != nil {
		return SyncSummary{}, err
	}

	if err := os.WriteFile(filepath.Join(paths.Modules, "_index.md"), []byte(GenerateModuleIndex(modules)), 0o644); err != nil {
		return SyncSummary{}, err
	}

	filesWritten := 1
	for _, module := range modules {
		lines := []string{
			fmt.Sprintf("# Module: %s", module.Name),
			"",
			"## Path",
			fmt.Sprintf("`%s`", module.Path),
			"",
			fmt.Sprintf("## Files (%d)", module.FileCount),
		}
		for _, file := range module.Files {
			lines = append(lines, fmt.Sprintf("- `%s`", file))
		}
		lines = append(lines, "")

		filename := filepath.Join(paths.Modules, module.Name+".md")
		if err := os.WriteFile(filename, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
			return SyncSummary{}, err
		}
		filesWritten++
	}

	return SyncSummary{
		ModulesFound: len(modules),
		FilesWritten: filesWritten,
	}, nil
}
