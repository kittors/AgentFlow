package projectrules

import (
	"fmt"
	"io/fs"
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
	Lang    string
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

		path := targetRulesPath(root, name)
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

	lang := strings.TrimSpace(options.Lang)
	if lang == "" {
		lang = config.DefaultLang
	}

	// Deploy full rules + modules to {root}/.agentflow/ so IDE agents
	// can access them within the workspace (they cannot read ~/.agentflow/).
	targetName := "codex"
	if len(targetNames) > 0 {
		targetName = targetNames[0]
	}
	if err := m.deployProjectRules(root, targetName, profile, lang); err != nil {
		return nil, err
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

		files, err := buildWrites(target, profile, lang)
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
			if !containsPath(written, dst) {
				written = append(written, dst)
			}
		}
	}
	return written, nil
}

// deployProjectRules deploys the full compiled rules and module files
// into {root}/.agentflow/ so IDE agents can read them within the workspace.
// Unlike the global install (which writes to ~/.agentflow/), project-level
// installs must keep everything inside the project directory because IDE
// agents are restricted to the workspace.
func (m *Manager) deployProjectRules(root, targetName, profile, lang string) error {
	localDir := filepath.Join(root, config.GlobalRulesDir)
	if err := os.MkdirAll(localDir, 0o755); err != nil {
		return err
	}

	rendered, err := buildGlobalRulesContent(targetName, profile, lang)
	if err != nil {
		return err
	}
	rulesPath := filepath.Join(localDir, config.GlobalRulesFile)
	if err := config.SafeWrite(rulesPath, []byte(rendered), 0o644); err != nil {
		return err
	}

	return deployModuleDirTo(filepath.Join(localDir, config.PluginDirName), lang)
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

		path := targetRulesPath(root, name)
		existing, err := os.ReadFile(path)
		if err != nil {
			continue // file doesn't exist
		}

		re := regexp.MustCompile(`(?s)<!-- ` + regexp.QuoteMeta(config.AgentFlowMarker) + `.*?<!-- /` + regexp.QuoteMeta(config.AgentFlowMarker) + `.*?-->\n?`)
		if re.Match(existing) {
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

	if !hasManagedProjectTargets(root) {
		skillPath := sharedSkillPath(root)
		if config.IsAgentFlowFile(skillPath) {
			if err := config.SafeRemove(skillPath); err != nil {
				return removed, err
			}
			removed = append(removed, skillPath)
		}

		localRulesPath := filepath.Join(root, config.GlobalRulesDir, config.GlobalRulesFile)
		if config.IsAgentFlowFile(localRulesPath) {
			_ = config.SafeRemove(localRulesPath)
		}
		_ = config.SafeRemove(filepath.Join(root, config.GlobalRulesDir, config.PluginDirName))
		cleanEmptyParents(root, sharedSkillPath(root))
	}

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

func targetRulesPath(root, targetName string) string {
	switch targetName {
	case "codex":
		return filepath.Join(root, "AGENTS.md")
	case "claude":
		return filepath.Join(root, "CLAUDE.md")
	default:
		return filepath.Join(root, "AGENTS.md")
	}
}

func sharedSkillPath(root string) string {
	return filepath.Join(root, ".agents", "skills", "agentflow", "SKILL.md")
}

func hasManagedProjectTargets(root string) bool {
	for _, name := range Names() {
		if config.IsAgentFlowFile(targetRulesPath(root, name)) {
			return true
		}
	}
	return false
}

func containsPath(paths []string, candidate string) bool {
	for _, path := range paths {
		if path == candidate {
			return true
		}
	}
	return false
}

func buildWrites(target Target, profile, lang string) ([]writeFile, error) {
	if profile == "" {
		profile = targets.DefaultProfile
	}
	switch target.Name {
	case "codex", "claude":
		filename := rulesFilenameForCLITarget(target.Name)
		rulesContent, err := buildProjectEntryRulesContent(target.Name, profile, lang)
		if err != nil {
			return nil, err
		}

		skillContent, err := readLangAsset(lang, "agentflow/_SKILL.md", "SKILL.md")
		if err != nil {
			return nil, err
		}
		skillPath := filepath.Join(".agents", "skills", "agentflow", "SKILL.md")

		return []writeFile{
			{
				RelPath: filename,
				Content: []byte(rulesContent),
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

func buildProjectEntryRulesContent(targetName, profile, lang string) (string, error) {
	rendered, err := buildGlobalRulesContent(targetName, profile, lang)
	if err != nil {
		return "", err
	}
	return config.RewriteEmbeddedModuleLinks(rendered, filepath.Join(config.GlobalRulesDir, config.PluginDirName)), nil
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

// buildGlobalRulesContent compiles the full rules content, mirroring the
// logic in install.buildRulesContent(). This avoids importing the install
// package which would create a circular dependency.
func buildGlobalRulesContent(targetName, profile, lang string) (string, error) {
	target, ok := targets.Lookup(targetName)
	if !ok {
		return "", fmt.Errorf("unknown target: %s", targetName)
	}
	if profile == "" {
		profile = targets.DefaultProfile
	}
	if !targets.ValidProfile(profile) {
		return "", fmt.Errorf("invalid profile: %s", profile)
	}

	content, err := readLangAsset(lang, "agentflow/_AGENTS.md", "AGENTS.md")
	if err != nil {
		return "", err
	}

	rendered := string(content)
	if !strings.Contains(rendered, config.Marker) {
		rendered = "<!-- AGENTFLOW_ROUTER: v1.0.0 -->\n" + rendered
	}
	rendered = strings.ReplaceAll(rendered, "{TARGET_CLI}", target.DisplayName)
	rendered = strings.ReplaceAll(rendered, "{HOOKS_SUMMARY}", target.HooksSummary)

	// Strip sections beyond the selected profile (no module inlining).
	rendered = stripBeyondProfile(rendered, profile)

	return strings.TrimRight(rendered, "\n") + "\n", nil
}

// stripBeyondProfile removes sections beyond the selected profile level.
func stripBeyondProfile(text, profile string) string {
	switch profile {
	case "lite":
		if idx := strings.Index(text, "<!-- PROFILE:standard"); idx > 0 {
			footer := "\n---\n\n> **AgentFlow** — 比分析更进一步，持续工作直到实现和验证完成。\n"
			return strings.TrimRight(text[:idx], "\n") + "\n" + footer
		}
	case "standard":
		if idx := strings.Index(text, "<!-- PROFILE:full"); idx > 0 {
			footer := "\n---\n\n> **AgentFlow** — 比分析更进一步，持续工作直到实现和验证完成。\n"
			return strings.TrimRight(text[:idx], "\n") + "\n" + footer
		}
	}
	return text
}

var (
	targetSubagentFiles = map[string]string{
		"codex":  "subagent_codex.md",
		"claude": "subagent_claude.md",
	}
	targetHooksFiles = map[string]string{
		"codex":  "hooks_codex.md",
		"claude": "hooks_claude.md",
	}
)

func buildCoreModuleForTarget(targetName, modFile string) (string, error) {
	content, err := readAssetWithFallback(filepath.ToSlash(filepath.Join("agentflow", "core", modFile)))
	if err != nil {
		return "", err
	}

	rendered := string(content)
	if modFile == "subagent.md" {
		cliFile := targetSubagentFiles[targetName]
		if cliFile == "" {
			cliFile = "subagent_other.md"
		}
		cliContent, err := readAssetWithFallback(filepath.ToSlash(filepath.Join("agentflow", "core", cliFile)))
		if err != nil {
			return "", err
		}
		rendered = strings.ReplaceAll(rendered, "{CLI_SUBAGENT_PROTOCOL}", string(cliContent))
	}
	if modFile == "hooks.md" {
		cliFile := targetHooksFiles[targetName]
		if cliFile == "" {
			cliFile = "hooks_other.md"
		}
		cliContent, err := readAssetWithFallback(filepath.ToSlash(filepath.Join("agentflow", "core", cliFile)))
		if err != nil {
			return "", err
		}
		rendered = strings.ReplaceAll(rendered, "{HOOKS_MATRIX}", string(cliContent))
	}
	return rendered, nil
}

// deployModuleDirTo deploys embedded agentflow module files to the target directory.
// Pass 1: shared files (skip zh/, en/ dirs). Pass 2: overlay from agentflow/{lang}/.
func deployModuleDirTo(baseDir, lang string) error {
	if err := config.SafeRemove(baseDir); err != nil {
		return err
	}
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return err
	}

	moduleFS, err := agentflowassets.Sub("agentflow")
	if err != nil {
		return err
	}

	// Pass 1: Deploy shared (non-language) files, skipping zh/ and en/ dirs.
	if err := fs.WalkDir(moduleFS, ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == "." {
			return nil
		}
		if entry.IsDir() && path == "locales" {
			return fs.SkipDir
		}
		if !entry.IsDir() && (path == "_AGENTS.md" || path == "_SKILL.md") {
			return nil
		}
		if entry.IsDir() {
			return os.MkdirAll(filepath.Join(baseDir, filepath.FromSlash(path)), 0o755)
		}
		content, err := fs.ReadFile(moduleFS, path)
		if err != nil {
			return err
		}
		return config.SafeWrite(filepath.Join(baseDir, filepath.FromSlash(path)), content, 0o644)
	}); err != nil {
		return err
	}

	// Pass 2: Overlay language-specific files from agentflow/locales/{lang}/.
	langFS, err := agentflowassets.Sub("agentflow/locales/" + lang)
	if err != nil {
		return nil // no language dir — not an error
	}
	return fs.WalkDir(langFS, ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == "." {
			return nil
		}
		if !entry.IsDir() && (path == "_AGENTS.md" || path == "_SKILL.md") {
			return nil
		}
		if entry.IsDir() {
			return os.MkdirAll(filepath.Join(baseDir, filepath.FromSlash(path)), 0o755)
		}
		content, err := fs.ReadFile(langFS, path)
		if err != nil {
			return err
		}
		return config.SafeWrite(filepath.Join(baseDir, filepath.FromSlash(path)), content, 0o644)
	})
}

// readLangAsset reads an embedded asset from the language-specific directory.
// e.g. readLangAsset("en", "agentflow/_AGENTS.md") tries "agentflow/en/_AGENTS.md"
// first, then falls back to "agentflow/_AGENTS.md".
func readLangAsset(lang string, paths ...string) ([]byte, error) {
	var langPaths []string
	for _, p := range paths {
		if strings.HasPrefix(p, "agentflow/") {
			rest := strings.TrimPrefix(p, "agentflow/")
			langPaths = append(langPaths, "agentflow/locales/"+lang+"/"+rest)
		} else {
			langPaths = append(langPaths, p)
		}
	}
	langPaths = append(langPaths, paths...) // fallback to root
	return readAssetWithFallback(langPaths...)
}
