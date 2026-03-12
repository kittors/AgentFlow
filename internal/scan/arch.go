package scan

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/kittors/AgentFlow/internal/projectroot"
)

const (
	LargeFileLines = 500
	LargeFileBytes = 50_000
	MaxFuncLines   = 80
)

type LargeFile struct {
	File      string   `json:"file"`
	SizeBytes int64    `json:"size_bytes"`
	Lines     int      `json:"lines"`
	Issues    []string `json:"issues"`
}

type LongFunction struct {
	File      string `json:"file"`
	Function  string `json:"function"`
	Lines     int    `json:"lines"`
	StartLine int    `json:"start_line"`
}

type Report struct {
	LargeFiles      []LargeFile    `json:"large_files"`
	MissingTests    []string       `json:"missing_tests"`
	CircularImports [][]string     `json:"circular_imports"`
	LongFunctions   []LongFunction `json:"long_functions"`
}

func FullScan(root string, sourceDirs []string) (Report, error) {
	largeFiles, err := ScanLargeFiles(root, sourceDirs)
	if err != nil {
		return Report{}, err
	}
	missingTests, err := ScanMissingTests(root, sourceDirs)
	if err != nil {
		return Report{}, err
	}
	circularImports, err := ScanCircularImports(root, sourceDirs)
	if err != nil {
		return Report{}, err
	}
	longFunctions, err := ScanLongFunctions(root, sourceDirs)
	if err != nil {
		return Report{}, err
	}

	return Report{
		LargeFiles:      largeFiles,
		MissingTests:    missingTests,
		CircularImports: circularImports,
		LongFunctions:   longFunctions,
	}, nil
}

func ScanLargeFiles(root string, sourceDirs []string) ([]LargeFile, error) {
	files, err := collectFiles(root, sourceDirs, func(path string, entry os.DirEntry) bool {
		return !entry.IsDir() && !projectroot.IsHiddenName(entry.Name())
	})
	if err != nil {
		return nil, err
	}

	result := make([]LargeFile, 0)
	for _, path := range files {
		info, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		lines, err := countLines(path)
		if err != nil {
			lines = 0
		}

		issues := make([]string, 0, 2)
		if info.Size() > LargeFileBytes {
			issues = append(issues, "size")
		}
		if lines > LargeFileLines {
			issues = append(issues, "lines")
		}
		if len(issues) == 0 {
			continue
		}

		rel, _ := filepath.Rel(root, path)
		result = append(result, LargeFile{
			File:      filepath.ToSlash(rel),
			SizeBytes: info.Size(),
			Lines:     lines,
			Issues:    issues,
		})
	}

	slices.SortFunc(result, func(a, b LargeFile) int {
		return strings.Compare(a.File, b.File)
	})
	return result, nil
}

func ScanMissingTests(root string, sourceDirs []string) ([]string, error) {
	testNames := make(map[string]struct{})
	for _, testDir := range []string{"tests", "test"} {
		full := filepath.Join(root, testDir)
		entries, err := os.ReadDir(full)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasPrefix(entry.Name(), "test_") || filepath.Ext(entry.Name()) != ".py" {
				continue
			}
			testNames[strings.TrimSuffix(strings.TrimPrefix(entry.Name(), "test_"), ".py")] = struct{}{}
		}
	}

	files, err := collectFiles(root, sourceDirs, func(path string, entry os.DirEntry) bool {
		return !entry.IsDir() && filepath.Ext(entry.Name()) == ".py" &&
			!strings.HasPrefix(entry.Name(), "_") && !projectroot.IsHiddenName(entry.Name())
	})
	if err != nil {
		return nil, err
	}

	missing := make([]string, 0)
	for _, path := range files {
		stem := strings.TrimSuffix(filepath.Base(path), ".py")
		if _, ok := testNames[stem]; ok {
			continue
		}
		rel, _ := filepath.Rel(root, path)
		missing = append(missing, filepath.ToSlash(rel))
	}

	sortStrings(missing)
	return missing, nil
}

func ScanCircularImports(root string, sourceDirs []string) ([][]string, error) {
	files, err := collectFiles(root, sourceDirs, func(path string, entry os.DirEntry) bool {
		return !entry.IsDir() && filepath.Ext(entry.Name()) == ".py"
	})
	if err != nil {
		return nil, err
	}

	graph := make(map[string][]string, len(files))
	importPattern := regexp.MustCompile(`^(?:from|import)\s+([\w.]+)`)

	for _, path := range files {
		rel, _ := filepath.Rel(root, path)
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		matches := importPattern.FindAllStringSubmatch(string(content), -1)
		imports := make([]string, 0, len(matches))
		for _, match := range matches {
			imports = append(imports, strings.Split(match[1], ".")[0])
		}
		graph[filepath.ToSlash(rel)] = imports
	}

	visited := make(map[string]bool, len(graph))
	stack := make(map[string]bool, len(graph))
	pathStack := make([]string, 0, len(graph))
	cycles := make([][]string, 0)

	var dfs func(node string)
	dfs = func(node string) {
		visited[node] = true
		stack[node] = true
		pathStack = append(pathStack, node)

		for neighbor := range graph {
			if neighbor == node {
				continue
			}
			moduleName := strings.TrimSuffix(filepath.Base(neighbor), ".py")
			if !slices.Contains(graph[node], moduleName) {
				continue
			}
			if stack[neighbor] {
				idx := slices.Index(pathStack, neighbor)
				if idx >= 0 {
					cycle := append([]string(nil), pathStack[idx:]...)
					cycle = append(cycle, neighbor)
					cycles = append(cycles, cycle)
				}
				continue
			}
			if !visited[neighbor] {
				dfs(neighbor)
			}
		}

		pathStack = pathStack[:len(pathStack)-1]
		delete(stack, node)
	}

	for node := range graph {
		if !visited[node] {
			dfs(node)
		}
	}

	return cycles, nil
}

func ScanLongFunctions(root string, sourceDirs []string) ([]LongFunction, error) {
	files, err := collectFiles(root, sourceDirs, func(path string, entry os.DirEntry) bool {
		return !entry.IsDir() && filepath.Ext(entry.Name()) == ".py"
	})
	if err != nil {
		return nil, err
	}

	funcPattern := regexp.MustCompile(`^(\s*)def\s+(\w+)\s*\(`)
	result := make([]LongFunction, 0)

	for _, path := range files {
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		lines := strings.Split(string(content), "\n")
		currentFunc := ""
		funcStart := 0
		for i, line := range lines {
			match := funcPattern.FindStringSubmatch(line)
			if len(match) == 0 {
				continue
			}
			if currentFunc != "" && i-funcStart > MaxFuncLines {
				rel, _ := filepath.Rel(root, path)
				result = append(result, LongFunction{
					File:      filepath.ToSlash(rel),
					Function:  currentFunc,
					Lines:     i - funcStart,
					StartLine: funcStart + 1,
				})
			}
			currentFunc = match[2]
			funcStart = i
		}
		if currentFunc != "" && len(lines)-funcStart > MaxFuncLines {
			rel, _ := filepath.Rel(root, path)
			result = append(result, LongFunction{
				File:      filepath.ToSlash(rel),
				Function:  currentFunc,
				Lines:     len(lines) - funcStart,
				StartLine: funcStart + 1,
			})
		}
	}

	slices.SortFunc(result, func(a, b LongFunction) int {
		if cmp := strings.Compare(a.File, b.File); cmp != 0 {
			return cmp
		}
		return strings.Compare(a.Function, b.Function)
	})
	return result, nil
}

func collectFiles(root string, sourceDirs []string, include func(path string, entry os.DirEntry) bool) ([]string, error) {
	if len(sourceDirs) == 0 {
		sourceDirs = projectroot.DefaultScanDirs(root)
	}

	files := make([]string, 0)
	for _, dir := range sourceDirs {
		src := filepath.Join(root, dir)
		info, err := os.Stat(src)
		if err != nil || !info.IsDir() {
			continue
		}
		rootOnly := filepath.Clean(dir) == "."
		err = filepath.WalkDir(src, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				if projectroot.IsHiddenName(entry.Name()) {
					return filepath.SkipDir
				}
				if rootOnly && path != src {
					return filepath.SkipDir
				}
			}
			if include(path, entry) {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return files, nil
}

func countLines(path string) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	count := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		count++
	}
	return count, scanner.Err()
}

func sortStrings(items []string) {
	slices.SortFunc(items, func(a, b string) int {
		return strings.Compare(a, b)
	})
}
