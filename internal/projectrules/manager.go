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

type writeFile struct {
	RelPath string
	Content []byte
	Mode    os.FileMode
}

func expectedPaths(root, targetName string) []string {
	switch targetName {
	case "codex":
		return []string{filepath.Join(root, "AGENTS.md")}
	case "claude":
		return []string{filepath.Join(root, "CLAUDE.md")}
	case "gemini":
		return []string{filepath.Join(root, "GEMINI.md")}
	case "qwen":
		return []string{filepath.Join(root, "QWEN.md")}
	case "kiro":
		return []string{filepath.Join(root, "KIRO.md")}
	case "cursor":
		return []string{filepath.Join(root, ".cursor", "rules", "agentflow.mdc")}
	case "windsurf":
		return []string{filepath.Join(root, ".windsurfrules")}
	case "trae":
		return []string{filepath.Join(root, ".trae", "rules", "agentflow.md")}
	case "vscode-copilot":
		return []string{filepath.Join(root, ".github", "copilot-instructions.md")}
	case "cline":
		return []string{filepath.Join(root, ".clinerules")}
	case "antigravity":
		return []string{filepath.Join(root, ".agents", "skills", "agentflow", "SKILL.md")}
	default:
		return nil
	}
}

func buildWrites(target Target, profile string) ([]writeFile, error) {
	switch target.Name {
	case "codex", "claude", "gemini", "qwen", "kiro":
		content, err := buildAgentFlowRules(target.Name, profile)
		if err != nil {
			return nil, err
		}
		filename := rulesFilenameForCLITarget(target.Name)
		return []writeFile{{
			RelPath: filename,
			Content: []byte(content),
			Mode:    0o644,
		}}, nil

	case "cursor":
		content, err := readAssetWithFallback("agentflow/project_rules/ide_cursor_agentflow.mdc")
		if err != nil {
			return nil, err
		}
		return []writeFile{{RelPath: filepath.Join(".cursor", "rules", "agentflow.mdc"), Content: content, Mode: 0o644}}, nil

	case "windsurf":
		content, err := readAssetWithFallback("agentflow/project_rules/ide_windsurf_agentflow.md")
		if err != nil {
			return nil, err
		}
		return []writeFile{{RelPath: ".windsurfrules", Content: content, Mode: 0o644}}, nil

	case "trae":
		content, err := readAssetWithFallback("agentflow/project_rules/ide_trae_agentflow.md")
		if err != nil {
			return nil, err
		}
		return []writeFile{{RelPath: filepath.Join(".trae", "rules", "agentflow.md"), Content: content, Mode: 0o644}}, nil

	case "vscode-copilot":
		content, err := readAssetWithFallback("agentflow/project_rules/ide_vscode_copilot_agentflow.md")
		if err != nil {
			return nil, err
		}
		return []writeFile{{RelPath: filepath.Join(".github", "copilot-instructions.md"), Content: content, Mode: 0o644}}, nil

	case "cline":
		content, err := readAssetWithFallback("agentflow/project_rules/ide_cline_agentflow.md")
		if err != nil {
			return nil, err
		}
		return []writeFile{{RelPath: ".clinerules", Content: content, Mode: 0o644}}, nil

	case "antigravity":
		content, err := readAssetWithFallback("agentflow/_SKILL.md", "SKILL.md")
		if err != nil {
			return nil, err
		}
		return []writeFile{{RelPath: filepath.Join(".agents", "skills", "agentflow", "SKILL.md"), Content: content, Mode: 0o644}}, nil
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
	case "gemini":
		return "GEMINI.md"
	case "qwen":
		return "QWEN.md"
	case "kiro":
		return "KIRO.md"
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
	case "gemini":
		return "subagent_gemini.md"
	case "opencode":
		return "subagent_opencode.md"
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
