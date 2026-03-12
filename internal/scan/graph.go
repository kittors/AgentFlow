package scan

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/kittors/AgentFlow/internal/projectroot"
)

type GraphNode struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Name     string                 `json:"name"`
	Path     string                 `json:"path"`
	Metadata map[string]interface{} `json:"metadata"`
}

type GraphEdge struct {
	Source   string                 `json:"source"`
	Target   string                 `json:"target"`
	Type     string                 `json:"type"`
	Weight   float64                `json:"weight"`
	Metadata map[string]interface{} `json:"metadata"`
}

type GraphPayload struct {
	Version int         `json:"version"`
	Nodes   []GraphNode `json:"nodes,omitempty"`
	Edges   []GraphEdge `json:"edges,omitempty"`
}

type GraphSummary struct {
	NodeCount int            `json:"node_count"`
	EdgeCount int            `json:"edge_count"`
	EdgeTypes map[string]int `json:"edge_types"`
}

func BuildGraph(root string, sourceDirs []string) (GraphSummary, error) {
	if len(sourceDirs) == 0 {
		sourceDirs = projectroot.DefaultScanDirs(root)
	}

	fileNodes, err := buildFileNodes(root, sourceDirs)
	if err != nil {
		return GraphSummary{}, err
	}
	moduleNodes, err := buildModuleNodes(root, sourceDirs)
	if err != nil {
		return GraphSummary{}, err
	}

	nodes := append(moduleNodes, fileNodes...)
	contains := buildContainsEdges(fileNodes, moduleNodes)
	imports, err := buildImportEdges(root, fileNodes, sourceDirs)
	if err != nil {
		return GraphSummary{}, err
	}
	calls, err := buildCallEdges(root, fileNodes, sourceDirs)
	if err != nil {
		return GraphSummary{}, err
	}
	edges := append(append(contains, imports...), calls...)

	paths := projectroot.NewPaths(root)
	if err := os.MkdirAll(paths.Graph, 0o755); err != nil {
		return GraphSummary{}, err
	}

	nodesData, err := json.MarshalIndent(GraphPayload{Version: 2, Nodes: nodes}, "", "  ")
	if err != nil {
		return GraphSummary{}, err
	}
	edgesData, err := json.MarshalIndent(GraphPayload{Version: 2, Edges: edges}, "", "  ")
	if err != nil {
		return GraphSummary{}, err
	}
	if err := os.WriteFile(filepath.Join(paths.Graph, "nodes.json"), nodesData, 0o644); err != nil {
		return GraphSummary{}, err
	}
	if err := os.WriteFile(filepath.Join(paths.Graph, "edges.json"), edgesData, 0o644); err != nil {
		return GraphSummary{}, err
	}
	if err := os.WriteFile(filepath.Join(paths.Graph, "graph.mmd"), []byte(exportMermaid(nodes, edges)), 0o644); err != nil {
		return GraphSummary{}, err
	}

	return GraphSummary{
		NodeCount: len(nodes),
		EdgeCount: len(edges),
		EdgeTypes: map[string]int{
			"contains": len(contains),
			"imports":  len(imports),
			"calls":    len(calls),
		},
	}, nil
}

func buildFileNodes(root string, sourceDirs []string) ([]GraphNode, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	nodes := make([]GraphNode, 0)
	files, err := collectFiles(root, sourceDirs, func(path string, entry os.DirEntry) bool {
		return !entry.IsDir() && !projectroot.IsHiddenName(entry.Name())
	})
	if err != nil {
		return nil, err
	}

	for _, path := range files {
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		nodes = append(nodes, GraphNode{
			ID:   nodeID(rel, "file"),
			Type: "file",
			Name: filepath.Base(rel),
			Path: rel,
			Metadata: map[string]interface{}{
				"description": "",
				"created_at":  now,
				"updated_at":  now,
				"tags":        []string{strings.TrimPrefix(filepath.Ext(rel), ".")},
			},
		})
	}
	slices.SortFunc(nodes, func(a, b GraphNode) int {
		return strings.Compare(a.Path, b.Path)
	})
	return nodes, nil
}

func buildModuleNodes(root string, sourceDirs []string) ([]GraphNode, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	nodes := make([]GraphNode, 0)
	seen := make(map[string]struct{})
	for _, dir := range sourceDirs {
		if filepath.Clean(dir) == "." {
			continue
		}
		src := filepath.Join(root, dir)
		if info, err := os.Stat(src); err != nil || !info.IsDir() {
			continue
		}
		err := filepath.WalkDir(src, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if !entry.IsDir() || path == src || strings.HasPrefix(entry.Name(), ".") || strings.HasPrefix(entry.Name(), "_") {
				return nil
			}
			rel, _ := filepath.Rel(root, path)
			rel = filepath.ToSlash(rel)
			if _, ok := seen[rel]; ok {
				return nil
			}
			seen[rel] = struct{}{}
			nodes = append(nodes, GraphNode{
				ID:   nodeID(rel, "module"),
				Type: "module",
				Name: entry.Name(),
				Path: rel,
				Metadata: map[string]interface{}{
					"description": "",
					"created_at":  now,
					"updated_at":  now,
					"tags":        []string{},
				},
			})
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	slices.SortFunc(nodes, func(a, b GraphNode) int {
		return strings.Compare(a.Path, b.Path)
	})
	return nodes, nil
}

func buildContainsEdges(fileNodes, moduleNodes []GraphNode) []GraphEdge {
	now := time.Now().UTC().Format(time.RFC3339)
	modules := make(map[string]string, len(moduleNodes))
	for _, node := range moduleNodes {
		modules[node.Path] = node.ID
	}

	edges := make([]GraphEdge, 0)
	for _, node := range fileNodes {
		parent := filepath.ToSlash(filepath.Dir(node.Path))
		if parent == "." {
			continue
		}
		moduleID, ok := modules[parent]
		if !ok {
			continue
		}
		edges = append(edges, GraphEdge{
			Source: moduleID,
			Target: node.ID,
			Type:   "contains",
			Weight: 1.0,
			Metadata: map[string]interface{}{
				"description": "directory contains file",
				"created_at":  now,
			},
		})
	}
	return edges
}

func buildImportEdges(root string, fileNodes []GraphNode, sourceDirs []string) ([]GraphEdge, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	stemToID := make(map[string]string)
	for _, node := range fileNodes {
		if strings.HasSuffix(node.Name, ".py") {
			stemToID[strings.TrimSuffix(node.Name, ".py")] = node.ID
			continue
		}
		if strings.HasSuffix(node.Name, ".go") {
			stemToID[strings.TrimSuffix(node.Name, ".go")] = node.ID
		}
	}

	pythonImportPattern := regexp.MustCompile(`^(?:from|import)\s+([\w.]+)`)
	goImportPattern := regexp.MustCompile(`"([^"]+)"`)
	edges := make([]GraphEdge, 0)
	files, err := collectFiles(root, sourceDirs, func(path string, entry os.DirEntry) bool {
		if entry.IsDir() {
			return false
		}
		ext := filepath.Ext(entry.Name())
		return ext == ".py" || ext == ".go"
	})
	if err != nil {
		return nil, err
	}
	for _, path := range files {
		rel, _ := filepath.Rel(root, path)
		sourceID := nodeID(filepath.ToSlash(rel), "file")
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		text := string(content)
		switch filepath.Ext(path) {
		case ".py":
			for _, match := range pythonImportPattern.FindAllStringSubmatch(text, -1) {
				moduleName := strings.Split(match[1], ".")[0]
				targetID, ok := stemToID[moduleName]
				if !ok || targetID == sourceID {
					continue
				}
				edges = append(edges, GraphEdge{
					Source: sourceID,
					Target: targetID,
					Type:   "imports",
					Weight: 0.8,
					Metadata: map[string]interface{}{
						"description": fmt.Sprintf("imports %s", moduleName),
						"created_at":  now,
					},
				})
			}
		case ".go":
			for _, match := range goImportPattern.FindAllStringSubmatch(text, -1) {
				moduleName := pathBaseWithoutVersion(match[1])
				targetID, ok := stemToID[moduleName]
				if !ok || targetID == sourceID {
					continue
				}
				edges = append(edges, GraphEdge{
					Source: sourceID,
					Target: targetID,
					Type:   "imports",
					Weight: 0.8,
					Metadata: map[string]interface{}{
						"description": fmt.Sprintf("imports %s", moduleName),
						"created_at":  now,
					},
				})
			}
		}
	}
	return dedupeEdges(edges), nil
}

func buildCallEdges(root string, fileNodes []GraphNode, sourceDirs []string) ([]GraphEdge, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	functionToFile := make(map[string]string)
	files, err := collectFiles(root, sourceDirs, func(path string, entry os.DirEntry) bool {
		if entry.IsDir() {
			return false
		}
		ext := filepath.Ext(entry.Name())
		return ext == ".py" || ext == ".go"
	})
	if err != nil {
		return nil, err
	}

	pythonDefPattern := regexp.MustCompile(`^def\s+(\w+)\s*\(`)
	goDefPattern := regexp.MustCompile(`^func\s+(?:\([^)]+\)\s*)?([A-Za-z_]\w*)\s*\(`)
	callPattern := regexp.MustCompile(`\b([A-Za-z_]\w*)\s*\(`)
	fileContents := make(map[string]string, len(files))

	for _, path := range files {
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		text := string(content)
		fileContents[path] = text
		rel, _ := filepath.Rel(root, path)
		fileID := nodeID(filepath.ToSlash(rel), "file")
		for _, line := range strings.Split(text, "\n") {
			trimmed := strings.TrimSpace(line)
			switch filepath.Ext(path) {
			case ".py":
				if match := pythonDefPattern.FindStringSubmatch(trimmed); len(match) > 1 && !strings.HasPrefix(match[1], "_") {
					functionToFile[match[1]] = fileID
				}
			case ".go":
				if match := goDefPattern.FindStringSubmatch(trimmed); len(match) > 1 {
					functionToFile[match[1]] = fileID
				}
			}
		}
	}

	edges := make([]GraphEdge, 0)
	for _, path := range files {
		rel, _ := filepath.Rel(root, path)
		sourceID := nodeID(filepath.ToSlash(rel), "file")
		for _, match := range callPattern.FindAllStringSubmatch(fileContents[path], -1) {
			targetID, ok := functionToFile[match[1]]
			if !ok || targetID == sourceID {
				continue
			}
			edges = append(edges, GraphEdge{
				Source: sourceID,
				Target: targetID,
				Type:   "calls",
				Weight: 0.6,
				Metadata: map[string]interface{}{
					"description": fmt.Sprintf("calls %s", match[1]),
					"created_at":  now,
				},
			})
		}
	}
	return dedupeEdges(edges), nil
}

func exportMermaid(nodes []GraphNode, edges []GraphEdge) string {
	lines := []string{"graph LR"}
	for _, node := range nodes {
		label := strings.NewReplacer(`"`, `'`, "[", "(", "]", ")").Replace(node.Name)
		if node.Type == "module" {
			lines = append(lines, fmt.Sprintf(`    %s[["%s"]]`, node.ID, label))
			continue
		}
		lines = append(lines, fmt.Sprintf(`    %s["%s"]`, node.ID, label))
	}

	styles := map[string]string{
		"contains": "-->",
		"imports":  "-.->",
		"calls":    "==>",
	}
	for _, edge := range edges {
		lines = append(lines, fmt.Sprintf("    %s %s|%s| %s", edge.Source, styles[edge.Type], edge.Type, edge.Target))
	}
	return strings.Join(lines, "\n") + "\n"
}

func nodeID(name, nodeType string) string {
	hash := sha256.Sum256([]byte(nodeType + ":" + name))
	return hex.EncodeToString(hash[:])[:12]
}

func dedupeEdges(edges []GraphEdge) []GraphEdge {
	seen := make(map[string]struct{}, len(edges))
	out := make([]GraphEdge, 0, len(edges))
	for _, edge := range edges {
		key := edge.Source + "|" + edge.Target + "|" + edge.Type
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, edge)
	}
	return out
}

func pathBaseWithoutVersion(importPath string) string {
	parts := strings.Split(strings.TrimSpace(importPath), "/")
	for index := len(parts) - 1; index >= 0; index-- {
		part := parts[index]
		if part == "" || regexp.MustCompile(`^v\d+$`).MatchString(part) {
			continue
		}
		return part
	}
	return ""
}
