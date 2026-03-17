package ui

import (
	"io"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/kittors/AgentFlow/internal/i18n"
)

func RunInteractiveFlow(catalog i18n.Catalog, version string, callbacks InteractiveCallbacks, output io.Writer) error {
	if output == nil {
		output = io.Discard
	}

	wd, _ := os.Getwd()

	model := interactiveFlowModel{
		catalog:      catalog,
		version:      version,
		callbacks:    callbacks,
		screen:       flowScreenMain,
		projectRoot:  wd,
		busy:         true,                    // spinner shows immediately while Init loads status
		initLoading:  true,                    // tracks that Init's refreshStatusCmd is in flight
		activeAction: flowActionRefreshStatus, // initial action for busyMessage
		status: Panel{
			Title: catalog.Msg("环境状态", "Environment"),
			Lines: []string{catalog.Msg("正在加载状态…", "Loading status...")},
		},
		mainOptions: []Option{
			{
				Value:       string(ActionToolbox),
				Label:       catalog.Msg("🧰 工具箱", "🧰 Toolbox"),
				Badge:       catalog.Msg("工具", "TOOLS"),
				Description: catalog.Msg("管理 CLI 工具、MCP Servers 和 Skills。", "Manage CLI tools, MCP servers, and skills."),
			},
			{
				Value:       string(ActionAgentFlow),
				Label:       catalog.Msg("⚡ AgentFlow", "⚡ AgentFlow"),
				Badge:       catalog.Msg("规则", "RULES"),
				Description: catalog.Msg("安装或卸载 AgentFlow 规则（全局/项目级）。", "Install or uninstall AgentFlow rules (global/project-level)."),
			},
			{
				Value:       string(ActionClean),
				Label:       catalog.Msg("🧹 清理缓存", "🧹 Clean caches"),
				Badge:       catalog.Msg("清理", "CLEAN"),
				Description: catalog.Msg("清除 AgentFlow 生成的缓存、临时目录和派生产物。", "Remove AgentFlow caches, temporary directories, and derived artifacts."),
			},
			{
				Value:       string(ActionUpdate),
				Label:       catalog.Msg("⬆️  更新", "⬆️  Update"),
				Badge:       catalog.Msg("更新", "UPDATE"),
				Description: catalog.Msg("检测并安装 AgentFlow 的最新版本。", "Check for and install the latest AgentFlow version."),
			},
			{
				Value:       string(ActionExit),
				Label:       catalog.Msg("🚪 退出", "🚪 Exit"),
				Badge:       catalog.Msg("退出", "EXIT"),
				Description: catalog.Msg("退出交互菜单并返回终端。", "Leave the interactive menu and return to the terminal."),
			},
		},
		toolboxOptions: []Option{
			{
				Value:       string(ActionCLI),
				Label:       catalog.Msg("CLI 工具", "CLI Tools"),
				Badge:       "CLI",
				Description: catalog.Msg("安装、配置和管理 Codex / Claude Code 等 CLI 工具。", "Install, configure, and manage CLI tools like Codex / Claude Code."),
			},
			{
				Value:       string(ActionMCP),
				Label:       catalog.Msg("MCP Servers", "MCP Servers"),
				Badge:       "MCP",
				Description: catalog.Msg("为 CLI 写入、查看与移除 MCP servers 配置。", "Write, inspect, and remove MCP server configs for CLIs."),
			},
			{
				Value:       string(ActionSkill),
				Label:       catalog.Msg("Skills", "Skills"),
				Badge:       "SKILL",
				Description: catalog.Msg("为 CLI 安装、查看与卸载 skills。", "Install, inspect, and uninstall skills for CLIs."),
			},
		},
		agentflowOptions: []Option{
			{
				Value:       "install-global",
				Label:       catalog.Msg("全局安装", "Global install"),
				Badge:       catalog.Msg("全局", "GLOBAL"),
				Description: catalog.Msg("安装 AgentFlow 规则到全局用户配置目录（~/.codex, ~/.claude），所有项目共享。", "Install AgentFlow rules to global user config dirs, shared across all projects."),
			},
			{
				Value:       "install-project",
				Label:       catalog.Msg("项目级安装", "Project install"),
				Badge:       catalog.Msg("项目", "PROJECT"),
				Description: catalog.Msg("安装 AgentFlow 规则到当前项目目录，仅对当前项目生效。", "Install AgentFlow rules to the current project directory."),
			},
			{
				Value:       "uninstall",
				Label:       catalog.Msg("卸载 AgentFlow", "Uninstall AgentFlow"),
				Badge:       catalog.Msg("卸载", "REMOVE"),
				Description: catalog.Msg("卸载已安装的 AgentFlow 规则（全局/项目级）。", "Remove installed AgentFlow rules (global/project-level)."),
			},
		},
		installHubOptions: []Option{
			{
				Value:       "bootstrap-cli",
				Label:       catalog.Msg("安装 CLI 工具", "Install CLI tools"),
				Badge:       catalog.Msg("CLI", "CLI"),
				Description: catalog.Msg("快速安装 Codex 和 Claude Code，并补齐 Node 依赖。", "Quickly install Codex and Claude Code, including Node prerequisites."),
			},
			{
				Value:       "install-agentflow",
				Label:       catalog.Msg("安装 AgentFlow", "Install AgentFlow"),
				Badge:       catalog.Msg("安装", "INSTALL"),
				Description: catalog.Msg("安装 AgentFlow 规则到全局 CLI 或当前项目目录。全局安装写入用户配置目录（~/.codex, ~/.claude），项目安装写入当前工作目录。", "Install AgentFlow rules globally (user config dirs) or into the current project directory."),
			},
			{
				Value:       "uninstall-agentflow",
				Label:       catalog.Msg("卸载 AgentFlow", "Uninstall AgentFlow"),
				Badge:       catalog.Msg("卸载", "REMOVE"),
				Description: catalog.Msg("卸载已安装的 AgentFlow 规则（全局/项目级），同时保留 CLI 工具和你的原有配置。", "Remove installed AgentFlow rules (global/project), while preserving CLI tools and your own config."),
			},
			{
				Value:       "uninstall-cli",
				Label:       catalog.Msg("卸载 CLI 工具（完整卸载）", "Uninstall CLI tools (full removal)"),
				Badge:       catalog.Msg("CLI", "CLI"),
				Description: catalog.Msg("卸载 Codex / Claude，并默认删除配置目录。", "Uninstall CLI tools like Codex / Claude and purge their config directories by default."),
			},
		},
		mcpActions: []Option{
			{Value: "list", Label: catalog.Msg("查看已配置 MCP", "List configured MCP"), Badge: catalog.Msg("列表", "LIST"), Description: catalog.Msg("列出该 CLI 已配置的 MCP servers。", "List MCP servers configured for this CLI.")},
			{Value: "install", Label: catalog.Msg("安装推荐 MCP", "Install recommended MCP"), Badge: catalog.Msg("安装", "ADD"), Description: catalog.Msg("安装置顶推荐：Context7 / Playwright / Filesystem / Tavily。", "Install pinned recommendations: Context7 / Playwright / Filesystem / Tavily.")},
			{Value: "remove", Label: catalog.Msg("移除 MCP", "Remove MCP"), Badge: catalog.Msg("移除", "DEL"), Description: catalog.Msg("从该 CLI 中移除已配置的 MCP server。", "Remove an MCP server from this CLI.")},
		},
		skillActions: []Option{
			{Value: "list", Label: catalog.Msg("查看已安装 Skill", "List installed skills"), Badge: catalog.Msg("列表", "LIST"), Description: catalog.Msg("列出该 CLI 已安装的 skills。", "List skills installed for this CLI.")},
			{Value: "install", Label: catalog.Msg("安装推荐 Skill", "Install recommended skills"), Badge: catalog.Msg("安装", "ADD"), Description: catalog.Msg("安装一些常用示例 skill（也可用 CLI 安装任意 skills.sh/GitHub skill）。", "Install a few common example skills (use the CLI to install any skills.sh/GitHub skill).")},
			{Value: "uninstall", Label: catalog.Msg("卸载 Skill", "Uninstall a skill"), Badge: catalog.Msg("移除", "DEL"), Description: catalog.Msg("从该 CLI 中卸载已安装的 skill。", "Uninstall a skill from this CLI.")},
		},
		bootstrapActionOptions: []Option{
			{
				Value:       "auto",
				Label:       catalog.Msg("自动安装", "Automatic install"),
				Badge:       catalog.Msg("自动", "AUTO"),
				Description: catalog.Msg("自动检查 nvm / Node，并安装所选 CLI。", "Automatically verify nvm / Node and install the selected CLI."),
			},
			{
				Value:       "manual",
				Label:       catalog.Msg("查看手动安装提示", "Show manual install guidance"),
				Badge:       catalog.Msg("手动", "MANUAL"),
				Description: catalog.Msg("显示适合当前平台的手动安装步骤和命令。", "Show manual installation steps and commands for the current platform."),
			},
		},
		profileOptions: []Option{
			{Value: "lite", Label: "lite", Badge: catalog.Msg("轻量", "LITE"), Description: catalog.Msg("只部署核心规则，最省 token。", "Deploy only the core rules for the smallest token footprint.")},
			{Value: "standard", Label: "standard", Badge: catalog.Msg("标准", "STANDARD"), Description: catalog.Msg("核心规则 + 常用模块，适合大多数项目。", "Core rules plus the common modules for most projects.")},
			{Value: "full", Label: "full", Badge: catalog.Msg("完整", "FULL"), Description: catalog.Msg("完整功能集，包含子代理、注意力和 Hooks。", "Full feature set including sub-agents, attention, and hooks."), Selected: true},
		},
		profileCursor: 2,
	}

	program := newInteractiveProgram(model, output)
	_, err := program.Run()
	return err
}

func (m interactiveFlowModel) Init() tea.Cmd {
	// NOTE: Init() has a value receiver so setting m.busy here would be lost.
	// Instead we set initLoading=true in the struct literal (NewInteractiveFlow)
	// and handle it via flowResultMsg / flowTickMsg.
	return tea.Batch(m.refreshStatusCmd(false), busyTickCmd())
}

func (m interactiveFlowModel) View() string {
	screen := m.selectionForCurrentScreen()
	return screen.View()
}
