package app

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/kittors/AgentFlow/internal/mcp"
	"github.com/kittors/AgentFlow/internal/projectrules"
	"github.com/kittors/AgentFlow/internal/skill"
	"github.com/kittors/AgentFlow/internal/targets"
	"github.com/kittors/AgentFlow/internal/ui"
)

var skillSourceDescriptionCache sync.Map

func (a *App) mcpTargetOptions() []ui.Option {
	homeDir, _ := os.UserHomeDir()
	names := targets.SortedMCPTargetNames()

	// Build a set of installed targets (CLI or AgentFlow present).
	installed := make(map[string]bool)
	for _, status := range a.Installer.DetectTargetStatuses() {
		if status.CLIInstalled || status.AgentFlowInstalled {
			installed[status.Target.Name] = true
		}
	}

	options := make([]ui.Option, 0, len(names))
	for _, name := range names {
		target, _ := targets.LookupMCP(name)
		if !installed[target.Name] {
			continue // skip targets whose CLI is not installed
		}
		options = append(options, ui.Option{
			Value:       target.Name,
			Label:       target.DisplayName,
			Badge:       strings.ToUpper(target.Name),
			Description: fmt.Sprintf(a.Catalog.Msg("配置目录: %s", "Config dir: %s"), filepath.ToSlash(filepath.Join(homeDir, target.Dir))),
		})
	}
	return options
}

func (a *App) mcpInstallOptions() []ui.Option {
	specs := mcp.BuiltinServers()
	options := make([]ui.Option, 0, len(specs))
	for _, spec := range specs {
		badge := "MCP"
		if spec.Pinned {
			badge = "PIN"
		}
		options = append(options, ui.Option{
			Value:       spec.Name,
			Label:       spec.Name,
			Badge:       badge,
			Description: spec.Description,
		})
	}
	return options
}

func (a *App) mcpConfigFields(server string) []ui.ConfigField {
	switch strings.ToLower(strings.TrimSpace(server)) {
	case "tavily-custom":
		return []ui.ConfigField{
			{
				Label:  a.Catalog.Msg("Tavily 代理 URL", "Tavily Proxy URL"),
				EnvVar: "TAVILY_API_URL",
				Type:   "text",
			},
			{
				Label:  "Tavily API Key",
				EnvVar: "TAVILY_API_KEY",
				Type:   "text",
			},
		}
	default:
		return nil
	}
}

func (a *App) mcpInstallWithEnvPanel(targetName, server string, env map[string]string) ui.Panel {
	target, ok := targets.LookupMCP(targetName)
	if !ok {
		return ui.Panel{
			Title: a.Catalog.Msg("MCP 安装", "MCP install"),
			Lines: []string{a.Catalog.Msg("未知目标。", "Unknown target.")},
		}
	}

	options := mcp.InstallOptions{}
	for k, v := range env {
		options.Env = append(options.Env, k+"="+v)
	}

	manager := mcp.NewManager()
	if err := manager.Install(target, server, options); err != nil {
		return errorPanel(a.Catalog.Msg("MCP 安装失败", "MCP install failed"), err)
	}

	lines := []string{
		fmt.Sprintf(a.Catalog.Msg("目标: %s", "Target: %s"), target.DisplayName),
		fmt.Sprintf(a.Catalog.Msg("已安装: %s", "Installed: %s"), server),
	}
	return ui.Panel{
		Title: a.Catalog.Msg("MCP 安装结果", "MCP install result"),
		Lines: lines,
	}
}

func (a *App) mcpRemoveOptions(targetName string) []ui.Option {
	target, ok := targets.LookupMCP(targetName)
	if !ok {
		return nil
	}
	descByName := map[string]string{}
	for _, spec := range mcp.BuiltinServers() {
		descByName[strings.ToLower(spec.Name)] = strings.TrimSpace(spec.Description)
	}
	manager := mcp.NewManager()
	names, err := manager.List(target)
	if err != nil {
		return nil
	}
	options := make([]ui.Option, 0, len(names))
	for _, name := range names {
		desc := descByName[strings.ToLower(strings.TrimSpace(name))]
		if desc == "" {
			desc = a.Catalog.Msg("移除该 MCP server 配置。", "Remove this MCP server configuration.")
		} else {
			desc = desc + " " + a.Catalog.Msg("（Enter 移除）", "(Enter removes)")
		}
		options = append(options, ui.Option{
			Value:       name,
			Label:       name,
			Badge:       "DEL",
			Description: desc,
		})
	}
	return options
}

func (a *App) mcpListPanel(targetName string) ui.Panel {
	target, ok := targets.LookupMCP(targetName)
	if !ok {
		return ui.Panel{
			Title: a.Catalog.Msg("MCP", "MCP"),
			Lines: []string{a.Catalog.Msg("未知目标。", "Unknown target.")},
		}
	}

	dotReady := lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("●")
	nameStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230"))
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	manager := mcp.NewManager()
	names, err := manager.List(target)
	if err != nil {
		return errorPanel(a.Catalog.Msg("MCP 列表", "MCP list"), err)
	}
	if len(names) == 0 {
		return ui.Panel{
			Title: a.Catalog.Msg("MCP 列表", "MCP list"),
			Lines: []string{
				fmt.Sprintf(a.Catalog.Msg("目标: %s", "Target: %s"), nameStyle.Render(target.DisplayName)),
				muted.Render(a.Catalog.Msg("未配置任何 MCP servers。", "No MCP servers configured.")),
				muted.Render(a.Catalog.Msg("提示：可进入“安装推荐 MCP”添加 Context7 / Playwright / Filesystem。", "Tip: use “Install recommended MCP” to add Context7 / Playwright / Filesystem.")),
			},
		}
	}
	lines := make([]string, 0, len(names)+6)
	lines = append(lines, fmt.Sprintf(a.Catalog.Msg("目标: %s", "Target: %s"), nameStyle.Render(target.DisplayName)))
	lines = append(lines, muted.Render(a.Catalog.Msg("图例：● 已配置（MCP 会按需启动）。", "Legend: ● configured (MCP starts on-demand).")))
	lines = append(lines, "")
	for _, name := range names {
		lines = append(lines, fmt.Sprintf("%s %s", dotReady, nameStyle.Render(name)))
	}
	return ui.Panel{
		Title: a.Catalog.Msg("MCP 列表", "MCP list"),
		Lines: lines,
	}
}

func (a *App) mcpInstallPanel(targetName, server string) ui.Panel {
	target, ok := targets.LookupMCP(targetName)
	if !ok {
		return ui.Panel{
			Title: a.Catalog.Msg("MCP 安装", "MCP install"),
			Lines: []string{a.Catalog.Msg("未知目标。", "Unknown target.")},
		}
	}

	options := mcp.InstallOptions{}
	if strings.EqualFold(server, "filesystem") {
		if wd, err := os.Getwd(); err == nil {
			options.Allow = []string{wd}
		}
	}

	manager := mcp.NewManager()
	if err := manager.Install(target, server, options); err != nil {
		return errorPanel(a.Catalog.Msg("MCP 安装失败", "MCP install failed"), err)
	}

	lines := []string{
		fmt.Sprintf(a.Catalog.Msg("目标: %s", "Target: %s"), target.DisplayName),
		fmt.Sprintf(a.Catalog.Msg("已安装: %s", "Installed: %s"), server),
	}
	if strings.EqualFold(server, "context7") {
		lines = append(lines, a.Catalog.Msg("提示：如需更高额度，可用 CLI 设置 CONTEXT7_API_KEY：agentflow mcp install <target> context7 --set-env=CONTEXT7_API_KEY=...", "Tip: for higher rate limits, set CONTEXT7_API_KEY via CLI: agentflow mcp install <target> context7 --set-env=CONTEXT7_API_KEY=..."))
	}
	if strings.EqualFold(server, "tavily-custom") {
		lines = append(lines, a.Catalog.Msg(
			"提示：tavily-custom 需要通过 CLI 配置 URL 和 Key：agentflow mcp install <target> tavily-custom --set-env=TAVILY_API_URL=<url> --set-env=TAVILY_API_KEY=<key>",
			"Tip: tavily-custom requires URL and Key via CLI: agentflow mcp install <target> tavily-custom --set-env=TAVILY_API_URL=<url> --set-env=TAVILY_API_KEY=<key>",
		))
	}
	return ui.Panel{
		Title: a.Catalog.Msg("MCP 安装结果", "MCP install result"),
		Lines: lines,
	}
}

func (a *App) mcpRemovePanel(targetName, server string) ui.Panel {
	target, ok := targets.LookupMCP(targetName)
	if !ok {
		return ui.Panel{
			Title: a.Catalog.Msg("MCP 移除", "MCP remove"),
			Lines: []string{a.Catalog.Msg("未知目标。", "Unknown target.")},
		}
	}

	manager := mcp.NewManager()
	if err := manager.Remove(target, server); err != nil {
		return errorPanel(a.Catalog.Msg("MCP 移除失败", "MCP remove failed"), err)
	}
	return ui.Panel{
		Title: a.Catalog.Msg("MCP 移除结果", "MCP remove result"),
		Lines: []string{
			fmt.Sprintf(a.Catalog.Msg("目标: %s", "Target: %s"), target.DisplayName),
			fmt.Sprintf(a.Catalog.Msg("已移除: %s", "Removed: %s"), server),
		},
	}
}

func (a *App) skillTargetOptions() []ui.Option {
	homeDir, _ := os.UserHomeDir()
	names := projectrules.SortedNames()
	options := make([]ui.Option, 0, len(names))
	for _, name := range names {
		target, _ := projectrules.Lookup(name)
		badge := strings.ToUpper(target.Kind)
		desc := []string{a.Catalog.Msg("项目级：可写入规则文件。", "Project: rule files supported.")}
		if target.Kind == projectrules.KindCLI {
			if cliTarget, ok := targets.Lookup(target.Name); ok {
				desc = append(desc, fmt.Sprintf(a.Catalog.Msg("全局 skills 目录: %s", "Global skills dir: %s"), filepath.ToSlash(filepath.Join(homeDir, cliTarget.Dir, "skills"))))
			} else {
				desc = append(desc, a.Catalog.Msg("全局：该目标不支持 Skills。", "Global: skills not supported for this target."))
			}
		} else {
			desc = append(desc, a.Catalog.Msg("全局：通常不支持 Skills。", "Global: skills are usually not supported."))
		}
		options = append(options, ui.Option{
			Value:       target.Name,
			Label:       target.DisplayName,
			Badge:       badge,
			Description: strings.Join(desc, " · "),
		})
	}
	return options
}

func (a *App) skillGlobalSupported(targetName string) bool {
	_, ok := targets.Lookup(targetName)
	return ok
}

func (a *App) skillInstallOptions(targetName string) []ui.Option {
	if !a.skillGlobalSupported(targetName) {
		return nil
	}

	recommended := []struct {
		Name  string
		URL   string
		Badge string
	}{
		{
			Name:  "turborepo",
			URL:   "https://skills.sh/vercel/turborepo/turborepo",
			Badge: "PIN",
		},
		{
			Name:  "next-upgrade",
			URL:   "https://skills.sh/vercel-labs/next-skills/next-upgrade",
			Badge: "PIN",
		},
		{
			Name:  "cra-to-next-migration",
			URL:   "https://skills.sh/vercel-labs/migration-skills/cra-to-next-migration",
			Badge: "PIN",
		},
	}

	options := make([]ui.Option, 0, len(recommended))
	client := &http.Client{Timeout: 25 * time.Second}
	for _, item := range recommended {
		desc := a.skillDescriptionFromSource(client, item.URL)
		if strings.TrimSpace(desc) == "" {
			desc = a.Catalog.Msg("（未找到 SKILL.md description）", "(SKILL.md description not found)")
		}
		options = append(options, ui.Option{
			Value:       item.URL,
			Label:       item.Name,
			Badge:       item.Badge,
			Description: desc,
		})
	}
	return options
}

func (a *App) skillUninstallOptions(targetName string) []ui.Option {
	target, ok := targets.Lookup(targetName)
	if !ok {
		return nil
	}
	manager := skill.NewManager()
	names, err := manager.List(target)
	if err != nil {
		return nil
	}
	sort.Strings(names)
	options := make([]ui.Option, 0, len(names))
	for _, name := range names {
		desc := a.skillDescriptionFromInstalled(target, name)
		if strings.TrimSpace(desc) == "" {
			desc = a.Catalog.Msg("卸载该 skill。", "Uninstall this skill.")
		}
		options = append(options, ui.Option{
			Value:       name,
			Label:       name,
			Badge:       "DEL",
			Description: desc,
		})
	}
	return options
}

func (a *App) skillListPanel(targetName string) ui.Panel {
	target, ok := targets.Lookup(targetName)
	if !ok {
		return ui.Panel{
			Title: a.Catalog.Msg("Skills", "Skills"),
			Lines: []string{a.Catalog.Msg("该目标不支持全局 Skill。", "This target does not support global skills.")},
		}
	}
	dotReady := lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("●")
	nameStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230"))
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	manager := skill.NewManager()
	names, err := manager.List(target)
	if err != nil {
		return errorPanel(a.Catalog.Msg("Skill 列表", "Skill list"), err)
	}
	if len(names) == 0 {
		return ui.Panel{
			Title: a.Catalog.Msg("Skill 列表", "Skill list"),
			Lines: []string{
				fmt.Sprintf(a.Catalog.Msg("目标: %s", "Target: %s"), nameStyle.Render(target.DisplayName)),
				muted.Render(a.Catalog.Msg("未检测到已安装的 skill。", "No installed skills detected.")),
				muted.Render(a.Catalog.Msg("提示：可进入“安装推荐 Skill”或使用 `agentflow skill install <target> <skills.sh URL>`。", "Tip: use “Install recommended skills” or `agentflow skill install <target> <skills.sh URL>`.")),
			},
		}
	}
	lines := make([]string, 0, len(names)+6)
	lines = append(lines, fmt.Sprintf(a.Catalog.Msg("目标: %s", "Target: %s"), nameStyle.Render(target.DisplayName)))
	lines = append(lines, muted.Render(a.Catalog.Msg("图例：● 已安装。", "Legend: ● installed.")))
	lines = append(lines, "")
	for _, name := range names {
		lines = append(lines, fmt.Sprintf("%s %s", dotReady, nameStyle.Render(name)))
	}
	return ui.Panel{
		Title: a.Catalog.Msg("Skill 列表", "Skill list"),
		Lines: lines,
	}
}

func (a *App) skillInstallPanel(targetName, source string) ui.Panel {
	target, ok := targets.Lookup(targetName)
	if !ok {
		return ui.Panel{
			Title: a.Catalog.Msg("Skill 安装", "Skill install"),
			Lines: []string{a.Catalog.Msg("该目标不支持全局 Skill。", "This target does not support global skills.")},
		}
	}
	manager := skill.NewManager()
	name, err := manager.Install(target, source, skill.InstallOptions{})
	if err != nil {
		return errorPanel(a.Catalog.Msg("Skill 安装失败", "Skill install failed"), err)
	}
	return ui.Panel{
		Title: a.Catalog.Msg("Skill 安装结果", "Skill install result"),
		Lines: []string{
			fmt.Sprintf(a.Catalog.Msg("目标: %s", "Target: %s"), target.DisplayName),
			fmt.Sprintf(a.Catalog.Msg("已安装: %s", "Installed: %s"), name),
		},
	}
}

func (a *App) skillUninstallPanel(targetName, name string) ui.Panel {
	target, ok := targets.Lookup(targetName)
	if !ok {
		return ui.Panel{
			Title: a.Catalog.Msg("Skill 卸载", "Skill uninstall"),
			Lines: []string{a.Catalog.Msg("该目标不支持全局 Skill。", "This target does not support global skills.")},
		}
	}
	manager := skill.NewManager()
	if err := manager.Uninstall(target, name); err != nil {
		return errorPanel(a.Catalog.Msg("Skill 卸载失败", "Skill uninstall failed"), err)
	}
	return ui.Panel{
		Title: a.Catalog.Msg("Skill 卸载结果", "Skill uninstall result"),
		Lines: []string{
			fmt.Sprintf(a.Catalog.Msg("目标: %s", "Target: %s"), target.DisplayName),
			fmt.Sprintf(a.Catalog.Msg("已卸载: %s", "Uninstalled: %s"), name),
		},
	}
}

func (a *App) skillDescriptionFromInstalled(target targets.Target, name string) string {
	homeDir, _ := os.UserHomeDir()
	path := filepath.Join(homeDir, target.Dir, "skills", name, "SKILL.md")
	desc, _ := skill.ParseSkillDescription(path)
	return strings.TrimSpace(desc)
}

func (a *App) skillDescriptionFromSource(client *http.Client, source string) string {
	source = strings.TrimSpace(source)
	if source == "" {
		return ""
	}
	if cached, ok := skillSourceDescriptionCache.Load(source); ok {
		if value, ok := cached.(string); ok {
			return value
		}
	}

	repo := source
	skillHint := ""
	if strings.HasPrefix(repo, "https://skills.sh/") || repo == "https://skills.sh" {
		resolvedRepo, resolvedSkill, err := skill.ResolveSkillsDotSh(client, repo, "agentflow-go")
		if err != nil {
			return ""
		}
		repo = resolvedRepo
		skillHint = resolvedSkill
	}

	owner, name, err := skill.ParseGitHubRepo(repo)
	if err != nil {
		return ""
	}
	ref, err := skill.ResolveDefaultBranch(client, owner, name, "agentflow-go")
	if err != nil {
		return ""
	}

	cacheDir, err := os.MkdirTemp("", "agentflow-skill-meta-cache-*")
	if err != nil {
		return ""
	}
	defer os.RemoveAll(cacheDir)

	zipPath, err := skill.DownloadGitHubZip(client, cacheDir, owner, name, ref, "agentflow-go")
	if err != nil {
		return ""
	}
	defer os.Remove(zipPath)

	extractDir, err := os.MkdirTemp("", "agentflow-skill-meta-unzip-*")
	if err != nil {
		return ""
	}
	defer os.RemoveAll(extractDir)

	rootDir, err := skill.UnzipRoot(zipPath, extractDir)
	if err != nil {
		return ""
	}
	sourceDir, _, err := skill.ResolveSkillDir(rootDir, skillHint)
	if err != nil {
		return ""
	}

	desc, _ := skill.ParseSkillDescription(filepath.Join(sourceDir, "SKILL.md"))
	desc = strings.TrimSpace(desc)
	if desc != "" {
		skillSourceDescriptionCache.Store(source, desc)
	}
	return desc
}
