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
