package projectrules

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	agentflowassets "github.com/kittors/AgentFlow"
	"github.com/kittors/AgentFlow/internal/config"
	"github.com/kittors/AgentFlow/internal/targets"
)

type InstallOptions struct {
	Profile string
}

type Status struct {
	Target   string
	Path     string
	Exists   bool
	Managed  bool
	Kind     string
	Detected string
}

type Manager struct{}

func NewManager() *Manager { return &Manager{} }

func (m *Manager) Detect(root string) ([]Status, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	results := make([]Status, 0, len(Names()))
	for _, name := range Names() {
		target, ok := Lookup(name)
		if !ok {
			continue
		}

		paths := expectedPaths(root, name)
		for _, path := range paths {
			exists := false
			if info, statErr := os.Stat(path); statErr == nil && !info.IsDir() {
				exists = true
			}
			managed := exists && config.IsAgentFlowFile(path)
			results = append(results, Status{
				Target:   target.Name,
				Path:     path,
				Exists:   exists,
				Managed:  managed,
				Kind:     target.Kind,
				Detected: filepath.ToSlash(strings.TrimPrefix(path, root+string(filepath.Separator))),
			})
		}
	}
	return results, nil
}

func (m *Manager) Install(root string, targetNames []string, options InstallOptions) ([]string, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	profile := strings.TrimSpace(options.Profile)
	if profile == "" {
		profile = targets.DefaultProfile
	}
	if !targets.ValidProfile(profile) {
		return nil, fmt.Errorf("invalid profile: %s", profile)
	}

	written := make([]string, 0, len(targetNames))
	for _, name := range targetNames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		target, ok := Lookup(name)
		if !ok {
			return nil, fmt.Errorf("unknown target: %s", name)
		}

		files, err := buildWrites(target, profile)
		if err != nil {
			return nil, err
		}

		for _, file := range files {
			dst := filepath.Join(root, file.RelPath)
			existing, err := os.ReadFile(dst)
			exists := err == nil

			if exists && !config.IsAgentFlowFile(dst) && !file.Inject {
				if _, err := config.BackupUserFile(dst); err != nil {
					return nil, err
				}
			}

			finalContent := file.Content
			if file.Inject && exists {
				re := regexp.MustCompile(`(?s)<!-- ` + regexp.QuoteMeta(config.AgentFlowMarker) + `.*?<!-- /` + regexp.QuoteMeta(config.AgentFlowMarker) + `.*?-->\n?`)
				if re.Match(existing) {
					finalContent = []byte(re.ReplaceAllString(string(existing), string(file.Content)))
				} else {
					finalContent = append(file.Content, existing...)
				}
			}

			if err := config.SafeWrite(dst, finalContent, file.Mode); err != nil {
				return nil, err
			}
			written = append(written, dst)
		}
	}
	return written, nil
}

// Uninstall removes AgentFlow-managed project rule files from the given root.
// It cleanly strips injected marker blocks from user files without deleting user content.
func (m *Manager) Uninstall(root string, targetNames []string) ([]string, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	removed := make([]string, 0, len(targetNames))
	for _, name := range targetNames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if _, ok := Lookup(name); !ok {
			return nil, fmt.Errorf("unknown target: %s", name)
		}

		paths := expectedPaths(root, name)
		for _, path := range paths {
			existing, err := os.ReadFile(path)
			if err != nil {
				continue // file doesn't exist
			}

			re := regexp.MustCompile(`(?s)<!-- ` + regexp.QuoteMeta(config.AgentFlowMarker) + `.*?<!-- /` + regexp.QuoteMeta(config.AgentFlowMarker) + `.*?-->\n?`)
			if fileIsSkill := strings.HasSuffix(path, filepath.FromSlash(".agents/skills/agentflow/SKILL.md")); fileIsSkill {
				if err := config.SafeRemove(path); err != nil {
					return removed, err
				}
				removed = append(removed, path)
			} else if re.Match(existing) {
				newContent := []byte(re.ReplaceAllString(string(existing), ""))
				if strings.TrimSpace(string(newContent)) == "" {
					if err := config.SafeRemove(path); err != nil {
						return removed, err
					}
				} else {
					if err := config.SafeWrite(path, newContent, 0o644); err != nil {
						return removed, err
					}
				}
				removed = append(removed, path)
			} else if config.IsAgentFlowFile(path) {
				if err := config.SafeRemove(path); err != nil {
					return removed, err
				}
				removed = append(removed, path)
			}
		}
	}
	cleanEmptyParents(root, filepath.Join(root, ".agents", "skills", "agentflow", "SKILL.md"))
	return removed, nil
}

func cleanEmptyParents(root, startPath string) {
	dir := filepath.Dir(startPath)
	for dir != root && strings.HasPrefix(dir, root) && len(dir) > len(root) {
		entries, err := os.ReadDir(dir)
		if err != nil || len(entries) > 0 {
			break
		}
		os.Remove(dir)
		dir = filepath.Dir(dir)
	}
}

type writeFile struct {
	RelPath string
	Content []byte
	Mode    os.FileMode
	Inject  bool
}

func expectedPaths(root, targetName string) []string {
	paths := []string{filepath.Join(root, ".agents", "skills", "agentflow", "SKILL.md")}
	switch targetName {
	case "codex":
		paths = append(paths, filepath.Join(root, "AGENTS.md"))
	case "claude":
		paths = append(paths, filepath.Join(root, "CLAUDE.md"))
	}
	return paths
}

func buildWrites(target Target, profile string) ([]writeFile, error) {
	if profile == "" {
		profile = targets.DefaultProfile
	}
	switch target.Name {
	case "codex", "claude":
		filename := rulesFilenameForCLITarget(target.Name)
		refContent := fmt.Sprintf("<!-- %s v1.0.0 -->\n\n> **[AgentFlow 管理规则]**\n> 请务必严格按照全局规范（如全局规则或 `.agents/skills/agentflow/SKILL.md`）执行所有操作。\n\n<!-- /%s -->\n", config.AgentFlowMarker, config.AgentFlowMarker)

		skillContent, err := readAssetWithFallback("agentflow/_SKILL.md", "SKILL.md")
		if err != nil {
			return nil, err
		}
		skillPath := filepath.Join(".agents", "skills", "agentflow", "SKILL.md")

		return []writeFile{
			{
				RelPath: filename,
				Content: []byte(refContent),
				Mode:    0o644,
				Inject:  true,
			},
			{
				RelPath: skillPath,
				Content: skillContent,
				Mode:    0o644,
				Inject:  false,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported target: %s", target.Name)
	}
}

func rulesFilenameForCLITarget(name string) string {
	if t, ok := targets.Lookup(name); ok && strings.TrimSpace(t.RulesFile) != "" {
		return t.RulesFile
	}
	// Fallbacks for project rules files.
	switch name {
	case "codex":
		return "AGENTS.md"
	case "claude":
		return "CLAUDE.md"
	default:
		return "AGENTS.md"
	}
}

func readAssetWithFallback(paths ...string) ([]byte, error) {
	var lastErr error
	for _, candidate := range paths {
		content, err := agentflowassets.ReadFile(filepath.ToSlash(candidate))
		if err == nil {
			return content, nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("asset not found")
	}
	return nil, lastErr
}
