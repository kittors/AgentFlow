package projectrules

import (
	"fmt"
	"os"
	"path/filepath"
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
			if _, statErr := os.Stat(dst); statErr == nil && !config.IsAgentFlowFile(dst) {
				if _, err := config.BackupUserFile(dst); err != nil {
					return nil, err
				}
			}
			if err := config.SafeWrite(dst, file.Content, file.Mode); err != nil {
				return nil, err
			}
			written = append(written, dst)
		}
	}
	return written, nil
}

// Uninstall removes AgentFlow-managed project rule files from the given root.
// Only files that contain the AgentFlow marker are removed (user files are left alone).
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
			if _, statErr := os.Stat(path); statErr != nil {
				continue // file doesn't exist
			}
			if !config.IsAgentFlowFile(path) {
				continue // not managed by AgentFlow, skip
			}
			if err := config.SafeRemove(path); err != nil {
				return removed, err
			}
			removed = append(removed, path)
		}
	}
	return removed, nil
}

type writeFile struct {
	RelPath string
	Content []byte
	Mode    os.FileMode
}

func expectedPaths(root, targetName string) []string {
	switch targetName {
	case "codex":
		return []string{
			filepath.Join(root, "AGENTS.md"),
			filepath.Join(root, ".agentflow", "rules_codex.md"),
		}
	case "claude":
		return []string{
			filepath.Join(root, "CLAUDE.md"),
			filepath.Join(root, ".agentflow", "rules_claude.md"),
		}
	default:
		return nil
	}

}

func buildWrites(target Target, profile string) ([]writeFile, error) {
	switch target.Name {
	case "codex", "claude":
		content, err := buildAgentFlowRules(target.Name, profile)
		if err != nil {
			return nil, err
		}
		filename := rulesFilenameForCLITarget(target.Name)
		rulesPath := filepath.Join(".agentflow", "rules_"+target.Name+".md")

		refContent := fmt.Sprintf("<!-- %s v1.0.0 -->\n\n# AgentFlow 管理规则\n请务必严格按照 `%s` 中定义的规则和规范执行所有操作。\n", config.AgentFlowMarker, filepath.ToSlash(rulesPath))

		return []writeFile{
			{
				RelPath: filename,
				Content: []byte(refContent),
				Mode:    0o644,
			},
			{
				RelPath: rulesPath,
				Content: []byte(content),
				Mode:    0o644,
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

func buildAgentFlowRules(targetName, profile string) (string, error) {
	t, ok := targets.Lookup(targetName)
	if !ok {
		return "", fmt.Errorf("unknown cli target: %s", targetName)
	}
	if profile == "" {
		profile = targets.DefaultProfile
	}
	if !targets.ValidProfile(profile) {
		return "", fmt.Errorf("invalid profile: %s", profile)
	}

	content, err := readAssetWithFallback("agentflow/_AGENTS.md", "AGENTS.md")
	if err != nil {
		return "", err
	}

	rendered := string(content)
	if !strings.Contains(rendered, config.AgentFlowMarker) {
		rendered = "<!-- AGENTFLOW_ROUTER: v1.0.0 -->\n" + rendered
	}
	rendered = strings.ReplaceAll(rendered, "{TARGET_CLI}", t.DisplayName)
	rendered = strings.ReplaceAll(rendered, "{HOOKS_SUMMARY}", t.HooksSummary)

	modules := targets.Profiles[profile]
	if len(modules) == 0 {
		return strings.TrimRight(rendered, "\n") + "\n", nil
	}

	var builder strings.Builder
	builder.WriteString(rendered)
	builder.WriteString("\n\n---\n\n")
	builder.WriteString(fmt.Sprintf("<!-- PROFILE:%s — Extended modules appended below -->\n\n", profile))
	for _, modFile := range modules {
		moduleContent, err := buildCoreModuleForTarget(targetName, modFile)
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(moduleContent) == "" {
			continue
		}
		builder.WriteString(moduleContent)
		builder.WriteString("\n\n")
	}
	return strings.TrimRight(builder.String(), "\n") + "\n", nil
}

func buildCoreModuleForTarget(targetName, modFile string) (string, error) {
	content, err := readAssetWithFallback(filepath.ToSlash(filepath.Join("agentflow", "core", modFile)))
	if err != nil {
		return "", err
	}

	rendered := string(content)
	if modFile == "subagent.md" {
		cliFile := installSubagentFile(targetName)
		cliContent, err := readAssetWithFallback(filepath.ToSlash(filepath.Join("agentflow", "core", cliFile)))
		if err != nil {
			return "", err
		}
		rendered = strings.ReplaceAll(rendered, "{CLI_SUBAGENT_PROTOCOL}", string(cliContent))
	}
	if modFile == "hooks.md" {
		cliFile := installHooksFile(targetName)
		cliContent, err := readAssetWithFallback(filepath.ToSlash(filepath.Join("agentflow", "core", cliFile)))
		if err != nil {
			return "", err
		}
		rendered = strings.ReplaceAll(rendered, "{HOOKS_MATRIX}", string(cliContent))
	}
	return rendered, nil
}

func installSubagentFile(targetName string) string {
	switch targetName {
	case "codex":
		return "subagent_codex.md"
	case "claude":
		return "subagent_claude.md"
	default:
		return "subagent_other.md"
	}
}

func installHooksFile(targetName string) string {
	switch targetName {
	case "codex":
		return "hooks_codex.md"
	case "claude":
		return "hooks_claude.md"
	default:
		return "hooks_other.md"
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
