package app

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/kittors/AgentFlow/internal/mcp"
	"github.com/kittors/AgentFlow/internal/skill"
	"github.com/kittors/AgentFlow/internal/targets"
	"github.com/kittors/AgentFlow/internal/ui"
)

func (a *App) mcpTargetOptions() []ui.Option {
	homeDir, _ := os.UserHomeDir()
	names := targets.SortedMCPTargetNames()
	options := make([]ui.Option, 0, len(names))
	for _, name := range names {
		target, _ := targets.LookupMCP(name)
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

func (a *App) mcpRemoveOptions(targetName string) []ui.Option {
	target, ok := targets.LookupMCP(targetName)
	if !ok {
		return nil
	}
	manager := mcp.NewManager()
	names, err := manager.List(target)
	if err != nil {
		return nil
	}
	options := make([]ui.Option, 0, len(names))
	for _, name := range names {
		options = append(options, ui.Option{
			Value:       name,
			Label:       name,
			Badge:       "DEL",
			Description: a.Catalog.Msg("移除该 MCP server 配置。", "Remove this MCP server configuration."),
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
	names := targets.SortedTargetNames()
	options := make([]ui.Option, 0, len(names))
	for _, name := range names {
		target, _ := targets.Lookup(name)
		options = append(options, ui.Option{
			Value:       target.Name,
			Label:       target.DisplayName,
			Badge:       strings.ToUpper(target.Name),
			Description: fmt.Sprintf(a.Catalog.Msg("skills 目录: %s", "skills dir: %s"), filepath.ToSlash(filepath.Join(homeDir, target.Dir, "skills"))),
		})
	}
	return options
}

func (a *App) skillInstallOptions() []ui.Option {
	recommended := []struct {
		Name  string
		URL   string
		Desc  string
		Badge string
	}{
		{
			Name:  "turborepo",
			URL:   "https://skills.sh/vercel/turborepo/turborepo",
			Desc:  a.Catalog.Msg("Turborepo 使用与最佳实践。", "Turborepo usage and best practices."),
			Badge: "PIN",
		},
		{
			Name:  "next-upgrade",
			URL:   "https://skills.sh/vercel-labs/next-skills/next-upgrade",
			Desc:  a.Catalog.Msg("升级 Next.js 的迁移指南与 codemods。", "Upgrade Next.js with migration guides and codemods."),
			Badge: "PIN",
		},
		{
			Name:  "cra-to-next-migration",
			URL:   "https://skills.sh/vercel-labs/migration-skills/cra-to-next-migration",
			Desc:  a.Catalog.Msg("Create React App 迁移到 Next.js 的完整指南。", "Comprehensive CRA → Next.js migration guide."),
			Badge: "PIN",
		},
	}

	options := make([]ui.Option, 0, len(recommended))
	for _, item := range recommended {
		options = append(options, ui.Option{
			Value:       item.URL,
			Label:       item.Name,
			Badge:       item.Badge,
			Description: item.Desc,
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
		options = append(options, ui.Option{
			Value:       name,
			Label:       name,
			Badge:       "DEL",
			Description: a.Catalog.Msg("卸载该 skill。", "Uninstall this skill."),
		})
	}
	return options
}

func (a *App) skillListPanel(targetName string) ui.Panel {
	target, ok := targets.Lookup(targetName)
	if !ok {
		return ui.Panel{
			Title: a.Catalog.Msg("Skills", "Skills"),
			Lines: []string{a.Catalog.Msg("未知目标。", "Unknown target.")},
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
			Lines: []string{a.Catalog.Msg("未知目标。", "Unknown target.")},
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
			Lines: []string{a.Catalog.Msg("未知目标。", "Unknown target.")},
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
