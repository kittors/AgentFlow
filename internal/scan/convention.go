package scan

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/kittors/AgentFlow/internal/projectroot"
)

type ConventionReport struct {
	Project     string                 `json:"project"`
	ExtractedAt string                 `json:"extracted_at"`
	Language    string                 `json:"language"`
	Conventions map[string]interface{} `json:"conventions"`
}

func ScanConventions(root string, sourceDirs []string) (ConventionReport, error) {
	language, err := detectConventionLanguage(root, sourceDirs)
	if err != nil {
		return ConventionReport{}, err
	}

	switch language {
	case "go":
		return ScanGoConventions(root, sourceDirs)
	case "python":
		return ScanPythonConventions(root, sourceDirs)
	default:
		return ConventionReport{
			Project:     projectroot.ProjectName(root),
			ExtractedAt: time.Now().UTC().Format(time.RFC3339),
			Language:    "unknown",
			Conventions: map[string]interface{}{
				"stats": map[string]int{"files_found": 0},
			},
		}, nil
	}
}

func ScanPythonConventions(root string, sourceDirs []string) (ConventionReport, error) {
	if len(sourceDirs) == 0 {
		sourceDirs = projectroot.DefaultScanDirs(root)
	}

	functions := make([]string, 0)
	classes := make([]string, 0)
	constants := make([]string, 0)
	importStyles := map[string]int{"absolute": 0, "relative": 0}
	docStyles := map[string]int{"google": 0, "numpy": 0, "sphinx": 0, "none": 0}

	funcPattern := regexp.MustCompile(`^def\s+(\w+)\s*\(`)
	classPattern := regexp.MustCompile(`^class\s+(\w+)`)
	constPattern := regexp.MustCompile(`^([A-Z][A-Z0-9_]{2,})\s*[:=]`)
	relativeImportPattern := regexp.MustCompile(`^from\s+\.`)
	absoluteImportPattern := regexp.MustCompile(`^(?:from|import)\s+[a-zA-Z]`)

	files, err := collectFiles(root, sourceDirs, func(path string, entry os.DirEntry) bool {
		return !entry.IsDir() && filepath.Ext(entry.Name()) == ".py"
	})
	if err != nil {
		return ConventionReport{}, err
	}

	for _, path := range files {
		content, err := os.ReadFile(path)
		if err != nil {
			return ConventionReport{}, err
		}
		text := string(content)
		for _, line := range strings.Split(text, "\n") {
			if match := funcPattern.FindStringSubmatch(line); len(match) > 1 {
				functions = append(functions, match[1])
			}
			if match := classPattern.FindStringSubmatch(line); len(match) > 1 {
				classes = append(classes, match[1])
			}
			if match := constPattern.FindStringSubmatch(line); len(match) > 1 {
				constants = append(constants, match[1])
			}
			if relativeImportPattern.MatchString(line) {
				importStyles["relative"]++
			}
			if absoluteImportPattern.MatchString(line) {
				importStyles["absolute"]++
			}
		}

		switch {
		case strings.Contains(text, "Args:") || strings.Contains(text, "Returns:"):
			docStyles["google"]++
		case strings.Contains(text, "Parameters\n") || strings.Contains(text, "----------"):
			docStyles["numpy"]++
		case strings.Contains(text, ":param ") || strings.Contains(text, ":type "):
			docStyles["sphinx"]++
		default:
			docStyles["none"]++
		}
	}

	report := ConventionReport{
		Project:     projectroot.ProjectName(root),
		ExtractedAt: time.Now().UTC().Format(time.RFC3339),
		Language:    "python",
		Conventions: map[string]interface{}{
			"naming": map[string]string{
				"functions": detectNamingStyle(functions),
				"classes":   defaultStyle(detectNamingStyle(classes), "PascalCase"),
				"constants": defaultStyle(detectNamingStyle(constants), "UPPER_SNAKE_CASE"),
			},
			"imports": map[string]interface{}{
				"style":          dominantKey(importStyles),
				"absolute_count": importStyles["absolute"],
				"relative_count": importStyles["relative"],
			},
			"documentation": map[string]string{
				"docstring_style": dominantKey(docStyles),
			},
			"stats": map[string]int{
				"functions_found": len(functions),
				"classes_found":   len(classes),
				"constants_found": len(constants),
			},
		},
	}
	return report, nil
}

func ScanGoConventions(root string, sourceDirs []string) (ConventionReport, error) {
	if len(sourceDirs) == 0 {
		sourceDirs = projectroot.DefaultScanDirs(root)
	}

	functions := make([]string, 0)
	types := make([]string, 0)
	importGroups := 0
	docComments := 0

	funcPattern := regexp.MustCompile(`^func\s+(?:\([^)]+\)\s*)?([A-Za-z_]\w*)\s*\(`)
	typePattern := regexp.MustCompile(`^type\s+([A-Za-z_]\w*)\s+`)
	docPattern := regexp.MustCompile(`^//\s+[A-Z]`)

	files, err := collectFiles(root, sourceDirs, func(path string, entry os.DirEntry) bool {
		return !entry.IsDir() && filepath.Ext(entry.Name()) == ".go"
	})
	if err != nil {
		return ConventionReport{}, err
	}

	for _, path := range files {
		content, err := os.ReadFile(path)
		if err != nil {
			return ConventionReport{}, err
		}
		text := string(content)
		if strings.Contains(text, "import (\n") {
			importGroups++
		}
		for _, line := range strings.Split(text, "\n") {
			trimmed := strings.TrimSpace(line)
			if match := funcPattern.FindStringSubmatch(trimmed); len(match) > 1 {
				functions = append(functions, match[1])
			}
			if match := typePattern.FindStringSubmatch(trimmed); len(match) > 1 {
				types = append(types, match[1])
			}
			if docPattern.MatchString(trimmed) {
				docComments++
			}
		}
	}

	return ConventionReport{
		Project:     projectroot.ProjectName(root),
		ExtractedAt: time.Now().UTC().Format(time.RFC3339),
		Language:    "go",
		Conventions: map[string]interface{}{
			"naming": map[string]string{
				"functions": detectNamingStyle(functions),
				"types":     defaultStyle(detectNamingStyle(types), "PascalCase"),
			},
			"imports": map[string]interface{}{
				"style":          "go_import_blocks",
				"grouped_blocks": importGroups,
			},
			"documentation": map[string]interface{}{
				"line_comments": docComments,
			},
			"stats": map[string]int{
				"functions_found": len(functions),
				"types_found":     len(types),
				"files_found":     len(files),
			},
		},
	}, nil
}

func SaveConventions(root string, report ConventionReport) (string, error) {
	paths := projectroot.NewPaths(root)
	if err := os.MkdirAll(paths.Conventions, 0o755); err != nil {
		return "", err
	}
	payload, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", err
	}
	filename := filepath.Join(paths.Conventions, "extracted.json")
	return filename, os.WriteFile(filename, payload, 0o644)
}

func detectNamingStyle(names []string) string {
	snake := regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	camel := regexp.MustCompile(`^[a-z][a-zA-Z0-9]*$`)
	pascal := regexp.MustCompile(`^[A-Z][a-zA-Z0-9]*$`)
	upper := regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)

	votes := map[string]int{
		"snake_case":       0,
		"camelCase":        0,
		"PascalCase":       0,
		"UPPER_SNAKE_CASE": 0,
	}
	for _, name := range names {
		switch {
		case upper.MatchString(name):
			votes["UPPER_SNAKE_CASE"]++
		case pascal.MatchString(name):
			votes["PascalCase"]++
		case camel.MatchString(name) && !strings.Contains(name, "_"):
			votes["camelCase"]++
		case snake.MatchString(name):
			votes["snake_case"]++
		}
	}
	return dominantKey(votes)
}

func dominantKey(values map[string]int) string {
	bestKey := "unknown"
	bestValue := -1
	for key, value := range values {
		if value > bestValue {
			bestKey = key
			bestValue = value
		}
	}
	if bestValue <= 0 {
		return "unknown"
	}
	return bestKey
}

func defaultStyle(style, fallback string) string {
	if style == "unknown" {
		return fallback
	}
	return style
}

func detectConventionLanguage(root string, sourceDirs []string) (string, error) {
	if len(sourceDirs) == 0 {
		sourceDirs = projectroot.DefaultScanDirs(root)
	}

	files, err := collectFiles(root, sourceDirs, func(path string, entry os.DirEntry) bool {
		if entry.IsDir() {
			return false
		}
		ext := filepath.Ext(entry.Name())
		return ext == ".go" || ext == ".py"
	})
	if err != nil {
		return "", err
	}

	goCount := 0
	pythonCount := 0
	for _, path := range files {
		switch filepath.Ext(path) {
		case ".go":
			goCount++
		case ".py":
			pythonCount++
		}
	}

	switch {
	case goCount > 0 && goCount >= pythonCount:
		return "go", nil
	case pythonCount > 0:
		return "python", nil
	default:
		return "unknown", nil
	}
}
