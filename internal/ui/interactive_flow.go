package ui

import (
	"fmt"
	"io"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/kittors/AgentFlow/internal/i18n"
)

type InteractiveCallbacks struct {
	Status                 func() Panel
	MCPTargetOptions       func() []Option
	MCPInstallOptions      func() []Option
	MCPRemoveOptions       func(target string) []Option
	MCPList                func(target string) Panel
	MCPInstall             func(target, server string) Panel
	MCPRemove              func(target, server string) Panel
	SkillTargetOptions     func() []Option
	SkillInstallOptions    func() []Option
	SkillUninstallOptions  func(target string) []Option
	SkillList              func(target string) Panel
	SkillInstall           func(target, source string) Panel
	SkillUninstall         func(target, name string) Panel
	BootstrapOptions       func() []Option
	BootstrapAutoSupported func(target string) bool
	BootstrapDetails       func(target string) Panel
	BootstrapAuto          func(target string) Panel
	BootstrapManual        func(target string) Panel
	InstallOptions         func() []Option
	UninstallOptions       func() []Option
	UninstallCLIOptions    func() []Option
	Install                func(profile string, targets []string) Panel
	Uninstall              func(targets []string) Panel
	UninstallCLI           func(targets []string) Panel
	Update                 func() (Panel, string)
	Clean                  func() Panel
}

type flowScreen int

const (
	flowScreenMain flowScreen = iota
	flowScreenInstallHub
	flowScreenMCPTargets
	flowScreenMCPActions
	flowScreenMCPInstall
	flowScreenMCPRemove
	flowScreenSkillTargets
	flowScreenSkillActions
	flowScreenSkillInstall
	flowScreenSkillUninstall
	flowScreenBootstrapTargets
	flowScreenBootstrapActions
	flowScreenProfile
	flowScreenInstallTargets
	flowScreenUninstallTargets
)

type flowAction int

const (
	flowActionRefreshStatus flowAction = iota
	flowActionMCPList
	flowActionMCPInstall
	flowActionMCPRemove
	flowActionSkillList
	flowActionSkillInstall
	flowActionSkillUninstall
	flowActionUpdate
	flowActionClean
	flowActionBootstrapAuto
	flowActionInstall
	flowActionUninstall
	flowActionUninstallCLI
)

type flowResultMsg struct {
	action          flowAction
	notice          *Panel
	status          Panel
	version         string
	bootstrapDetail *Panel
}

type flowTickMsg struct{}

type interactiveFlowModel struct {
	catalog   i18n.Catalog
	version   string
	callbacks InteractiveCallbacks

	width  int
	height int

	screen flowScreen
	busy   bool
	spin   int

	mainOptions            []Option
	installHubOptions      []Option
	mcpTargets             []Option
	mcpActions             []Option
	mcpInstallOptions      []Option
	mcpRemoveOptions       []Option
	skillTargets           []Option
	skillActions           []Option
	skillInstallOptions    []Option
	skillUninstallOptions  []Option
	bootstrapOptions       []Option
	bootstrapActionOptions []Option
	profileOptions         []Option
	installOptions         []Option
	uninstallOptions       []Option
	uninstallCLIMode       bool

	mainCursor            int
	installHubCursor      int
	mcpTargetCursor       int
	mcpActionCursor       int
	mcpInstallCursor      int
	mcpRemoveCursor       int
	skillTargetCursor     int
	skillActionCursor     int
	skillInstallCursor    int
	skillUninstallCursor  int
	bootstrapCursor       int
	bootstrapActionCursor int
	profileCursor         int
	installCursor         int
	uninstallCursor       int

	selectedProfile         string
	selectedMCPTarget       string
	selectedMCPServer       string
	selectedSkillTarget     string
	selectedSkillValue      string
	selectedBootstrapTarget string
	notice                  *Panel
	status                  Panel
	bootstrapDetail         *Panel
}

func RunInteractiveFlow(catalog i18n.Catalog, version string, callbacks InteractiveCallbacks, output io.Writer) error {
	if output == nil {
		output = io.Discard
	}

	model := interactiveFlowModel{
		catalog:   catalog,
		version:   version,
		callbacks: callbacks,
		screen:    flowScreenMain,
		status: Panel{
			Title: catalog.Msg("环境状态", "Environment"),
			Lines: []string{catalog.Msg("正在加载状态…", "Loading status...")},
		},
		mainOptions: []Option{
			{
				Value:       string(ActionInstall),
				Label:       catalog.Msg("安装", "Install"),
				Badge:       catalog.Msg("安装", "SETUP"),
				Description: catalog.Msg("先安装 Codex / Claude / Gemini 等 CLI，或继续把 AgentFlow 部署到已存在的 CLI。", "Install Codex / Claude / Gemini first, or deploy AgentFlow into CLIs that already exist."),
			},
			{
				Value:       string(ActionMCP),
				Label:       catalog.Msg("管理 MCP Servers", "Manage MCP servers"),
				Badge:       catalog.Msg("MCP", "MCP"),
				Description: catalog.Msg("为任意 CLI 写入、查看与移除 MCP servers 配置，并置顶推荐 Context7 / Playwright / Filesystem。", "Write, inspect, and remove MCP server configs for any CLI, with pinned recommendations like Context7 / Playwright / Filesystem."),
			},
			{
				Value:       string(ActionSkill),
				Label:       catalog.Msg("管理 Skills", "Manage skills"),
				Badge:       catalog.Msg("SKILL", "SKILL"),
				Description: catalog.Msg("为任意 CLI 安装、查看与卸载 skills（支持从 skills.sh/GitHub 安装）。", "Install, inspect, and uninstall skills for any CLI (supports skills.sh/GitHub sources)."),
			},
			{
				Value:       string(ActionUninstall),
				Label:       catalog.Msg("卸载已安装目标", "Uninstall from installed targets"),
				Badge:       catalog.Msg("移除", "REMOVE"),
				Description: catalog.Msg("从已接入 CLI 中清理 AgentFlow 产物，同时保留你的原有配置。", "Remove AgentFlow assets from integrated CLIs while preserving your own config where possible."),
			},
			{
				Value:       string(ActionUninstallCLI),
				Label:       catalog.Msg("卸载 CLI 工具", "Uninstall CLI tools"),
				Badge:       catalog.Msg("CLI", "CLI"),
				Description: catalog.Msg("卸载 Codex / Claude / Gemini / Qwen / Kiro 等 CLI 本体，并默认删除配置目录（完整卸载）。", "Uninstall CLI tools like Codex / Claude / Gemini / Qwen / Kiro and purge their config directories by default (full uninstall)."),
			},
			{
				Value:       string(ActionUpdate),
				Label:       catalog.Msg("更新 AgentFlow", "Update AgentFlow"),
				Badge:       catalog.Msg("更新", "UPDATE"),
				Description: catalog.Msg("检查最新 release，并把当前 Go 二进制原地更新到新版本。", "Check the latest release and replace the current Go binary in place."),
			},
			{
				Value:       string(ActionStatus),
				Label:       catalog.Msg("查看状态", "Show status"),
				Badge:       catalog.Msg("状态", "STATUS"),
				Description: catalog.Msg("刷新版本、可执行文件路径和所有支持 CLI 的接入状态。", "Refresh the current version, executable path, and integration status for every supported CLI."),
			},
			{
				Value:       string(ActionClean),
				Label:       catalog.Msg("清理缓存", "Clean caches"),
				Badge:       catalog.Msg("清理", "CLEAN"),
				Description: catalog.Msg("清除 AgentFlow 生成的缓存、临时目录和派生产物，保持环境整洁。", "Remove AgentFlow caches, temporary directories, and derived artifacts to keep the environment tidy."),
			},
			{
				Value:       string(ActionExit),
				Label:       catalog.Msg("退出", "Exit"),
				Badge:       catalog.Msg("退出", "EXIT"),
				Description: catalog.Msg("退出交互菜单并返回终端。", "Leave the interactive menu and return to the terminal."),
			},
		},
		installHubOptions: []Option{
			{
				Value:       "bootstrap-cli",
				Label:       catalog.Msg("安装 CLI 工具", "Install CLI tools"),
				Badge:       catalog.Msg("CLI", "CLI"),
				Description: catalog.Msg("快速安装 Codex、Claude Code、Gemini、Qwen、Kiro CLI，并补齐 Node / nvm / WSL2 依赖（Kiro 使用官方脚本）。", "Quickly install Codex, Claude Code, Gemini, Qwen, and Kiro CLI, including Node / nvm / WSL2 prerequisites (Kiro uses the official install script)."),
			},
			{
				Value:       "install-agentflow",
				Label:       catalog.Msg("安装 AgentFlow 到已安装 CLI", "Install AgentFlow into existing CLIs"),
				Badge:       catalog.Msg("接入", "APPLY"),
				Description: catalog.Msg("对已经存在的 CLI 写入 AgentFlow 规则、模块、技能和 hooks。", "Write AgentFlow rules, modules, skills, and hooks into CLIs that already exist."),
			},
		},
		mcpActions: []Option{
			{Value: "list", Label: catalog.Msg("查看已配置 MCP", "List configured MCP"), Badge: catalog.Msg("列表", "LIST"), Description: catalog.Msg("列出该 CLI 已配置的 MCP servers。", "List MCP servers configured for this CLI.")},
			{Value: "install", Label: catalog.Msg("安装推荐 MCP", "Install recommended MCP"), Badge: catalog.Msg("安装", "ADD"), Description: catalog.Msg("安装置顶推荐：Context7 / Playwright / Filesystem。", "Install pinned recommendations: Context7 / Playwright / Filesystem.")},
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
	return m.refreshStatusCmd(false)
}

func (m interactiveFlowModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch value := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = value.Width
		m.height = value.Height
		return m, nil
	case flowTickMsg:
		if !m.busy {
			return m, nil
		}
		m.spin = (m.spin + 1) % len(spinnerFrames)
		return m, busyTickCmd()
	case flowResultMsg:
		m.busy = false
		m.spin = 0
		if strings.TrimSpace(value.version) != "" {
			m.version = value.version
		}
		if strings.TrimSpace(value.status.Title) != "" || len(value.status.Lines) > 0 {
			m.status = value.status
		}
		if value.bootstrapDetail != nil {
			m.bootstrapDetail = value.bootstrapDetail
		}
		switch value.action {
		case flowActionRefreshStatus:
			m.bootstrapOptions = m.bootstrapOptionsList()
			m.installOptions = m.installOptionsList()
			m.uninstallOptions = m.uninstallOptionsList()
			if value.notice != nil {
				m.notice = value.notice
			}
			m.refreshBootstrapDetail()
		case flowActionMCPList, flowActionMCPInstall, flowActionMCPRemove:
			if value.notice != nil {
				m.notice = value.notice
			}
			m.screen = flowScreenMCPActions
			m.mcpInstallOptions = nil
			m.mcpRemoveOptions = nil
		case flowActionSkillList, flowActionSkillInstall, flowActionSkillUninstall:
			if value.notice != nil {
				m.notice = value.notice
			}
			m.screen = flowScreenSkillActions
			m.skillInstallOptions = nil
			m.skillUninstallOptions = nil
		case flowActionBootstrapAuto:
			m.bootstrapOptions = m.bootstrapOptionsList()
			m.installOptions = m.installOptionsList()
			m.uninstallOptions = m.uninstallOptionsList()
			if value.notice != nil {
				m.notice = value.notice
			}
			m.refreshBootstrapDetail()
		case flowActionInstall, flowActionUninstall, flowActionUninstallCLI, flowActionUpdate, flowActionClean:
			m.screen = flowScreenMain
			m.installOptions = nil
			m.uninstallOptions = nil
			m.uninstallCLIMode = false
			m.clearSelections()
			if value.notice != nil {
				m.notice = value.notice
			}
		}
		return m, nil
	case tea.MouseMsg:
		if m.busy {
			return m, nil
		}
		switch {
		case value.Button == tea.MouseButtonWheelUp || value.Type == tea.MouseWheelUp:
			m.moveCursor(-1)
		case value.Button == tea.MouseButtonWheelDown || value.Type == tea.MouseWheelDown:
			m.moveCursor(1)
		}
		return m, nil
	case tea.KeyMsg:
		if value.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		if m.busy {
			return m, nil
		}
		return m.handleKey(value)
	}

	return m, nil
}

func (m interactiveFlowModel) View() string {
	screen := m.selectionForCurrentScreen()
	return screen.View()
}

func (m interactiveFlowModel) handleKey(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.Type {
	case tea.KeyUp:
		m.moveCursor(-1)
		return m, nil
	case tea.KeyDown:
		m.moveCursor(1)
		return m, nil
	case tea.KeyPgUp:
		m.moveCursor(-5)
		return m, nil
	case tea.KeyPgDown:
		m.moveCursor(5)
		return m, nil
	case tea.KeyHome:
		m.setCursor(0)
		return m, nil
	case tea.KeyEnd:
		m.setCursor(m.currentOptionsLen() - 1)
		return m, nil
	case tea.KeySpace:
		if m.screen == flowScreenInstallTargets {
			m.toggleSelected(&m.installOptions, m.installCursor)
		}
		if m.screen == flowScreenUninstallTargets {
			m.toggleSelected(&m.uninstallOptions, m.uninstallCursor)
		}
		return m, nil
	case tea.KeyEsc:
		return m.handleBack()
	case tea.KeyEnter:
		return m.handleEnter()
	case tea.KeyRunes:
		if key.String() == " " {
			if m.screen == flowScreenInstallTargets {
				m.toggleSelected(&m.installOptions, m.installCursor)
			}
			if m.screen == flowScreenUninstallTargets {
				m.toggleSelected(&m.uninstallOptions, m.uninstallCursor)
			}
		}
	}
	return m, nil
}

func (m interactiveFlowModel) handleBack() (tea.Model, tea.Cmd) {
	switch m.screen {
	case flowScreenMain:
		return m, tea.Quit
	case flowScreenInstallHub:
		m.screen = flowScreenMain
	case flowScreenMCPTargets:
		m.screen = flowScreenMain
	case flowScreenMCPActions:
		m.screen = flowScreenMCPTargets
	case flowScreenMCPInstall, flowScreenMCPRemove:
		m.screen = flowScreenMCPActions
	case flowScreenSkillTargets:
		m.screen = flowScreenMain
	case flowScreenSkillActions:
		m.screen = flowScreenSkillTargets
	case flowScreenSkillInstall, flowScreenSkillUninstall:
		m.screen = flowScreenSkillActions
	case flowScreenBootstrapTargets:
		m.screen = flowScreenInstallHub
	case flowScreenBootstrapActions:
		m.screen = flowScreenBootstrapTargets
	case flowScreenProfile:
		m.screen = flowScreenInstallHub
	case flowScreenInstallTargets:
		m.screen = flowScreenProfile
	case flowScreenUninstallTargets:
		m.screen = flowScreenMain
		m.uninstallCLIMode = false
	}
	m.notice = nil
	return m, nil
}

func (m interactiveFlowModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.screen {
	case flowScreenMain:
		switch Action(m.mainOptions[m.mainCursor].Value) {
		case ActionInstall:
			m.screen = flowScreenInstallHub
			m.notice = nil
			return m, nil
		case ActionMCP:
			if m.callbacks.MCPTargetOptions == nil {
				m.notice = panelRef(Panel{
					Title: m.catalog.Msg("MCP", "MCP"),
					Lines: []string{m.catalog.Msg("当前构建未启用 MCP 管理回调。", "MCP management callbacks are not enabled in this build.")},
				})
				return m, nil
			}
			m.mcpTargets = cloneOptions(m.callbacks.MCPTargetOptions())
			if len(m.mcpTargets) == 0 {
				m.notice = panelRef(Panel{
					Title: m.catalog.Msg("MCP", "MCP"),
					Lines: []string{m.catalog.Msg("未检测到可管理的目标。", "No manageable targets detected.")},
				})
				return m, nil
			}
			m.screen = flowScreenMCPTargets
			m.mcpTargetCursor = 0
			m.selectedMCPTarget = m.mcpTargets[0].Value
			m.notice = nil
			return m, nil
		case ActionSkill:
			if m.callbacks.SkillTargetOptions == nil {
				m.notice = panelRef(Panel{
					Title: m.catalog.Msg("Skills", "Skills"),
					Lines: []string{m.catalog.Msg("当前构建未启用 Skill 管理回调。", "Skill management callbacks are not enabled in this build.")},
				})
				return m, nil
			}
			m.skillTargets = cloneOptions(m.callbacks.SkillTargetOptions())
			if len(m.skillTargets) == 0 {
				m.notice = panelRef(Panel{
					Title: m.catalog.Msg("Skills", "Skills"),
					Lines: []string{m.catalog.Msg("未检测到可管理的目标。", "No manageable targets detected.")},
				})
				return m, nil
			}
			m.screen = flowScreenSkillTargets
			m.skillTargetCursor = 0
			m.selectedSkillTarget = m.skillTargets[0].Value
			m.notice = nil
			return m, nil
		case ActionUninstall:
			m.uninstallCLIMode = false
			m.uninstallOptions = cloneOptions(m.callbacks.UninstallOptions())
			if len(m.uninstallOptions) == 0 {
				m.notice = panelRef(Panel{
					Title: m.catalog.Msg("卸载结果", "Uninstall result"),
					Lines: []string{m.catalog.Msg("未检测到已安装的 AgentFlow。", "No AgentFlow installations found.")},
				})
				return m, nil
			}
			m.screen = flowScreenUninstallTargets
			m.notice = nil
			return m, nil
		case ActionUninstallCLI:
			m.uninstallCLIMode = true
			if m.callbacks.UninstallCLIOptions != nil {
				m.uninstallOptions = cloneOptions(m.callbacks.UninstallCLIOptions())
			} else {
				m.uninstallOptions = nil
			}
			if len(m.uninstallOptions) == 0 {
				m.notice = panelRef(Panel{
					Title: m.catalog.Msg("卸载结果", "Uninstall result"),
					Lines: []string{m.catalog.Msg("未检测到可卸载的 CLI。", "No CLI installations found.")},
				})
				return m, nil
			}
			m.screen = flowScreenUninstallTargets
			m.notice = nil
			return m, nil
		case ActionUpdate:
			return m.startBusy(flowActionUpdate, m.catalog.Msg("正在检查最新版本并更新…", "Checking the latest release and updating..."))
		case ActionStatus:
			return m.startBusy(flowActionRefreshStatus, m.catalog.Msg("正在刷新状态…", "Refreshing status..."))
		case ActionClean:
			return m.startBusy(flowActionClean, m.catalog.Msg("正在清理缓存…", "Cleaning caches..."))
		case ActionExit:
			return m, tea.Quit
		}
	case flowScreenMCPTargets:
		if len(m.mcpTargets) == 0 {
			return m, nil
		}
		m.selectedMCPTarget = m.mcpTargets[m.mcpTargetCursor].Value
		m.screen = flowScreenMCPActions
		m.mcpActionCursor = 0
		m.notice = nil
		return m, nil
	case flowScreenMCPActions:
		if len(m.mcpActions) == 0 {
			return m, nil
		}
		switch m.mcpActions[m.mcpActionCursor].Value {
		case "list":
			m.selectedMCPServer = ""
			return m.startBusy(flowActionMCPList, m.catalog.Msg("正在读取 MCP 配置…", "Reading MCP configuration..."))
		case "install":
			if m.callbacks.MCPInstallOptions != nil {
				m.mcpInstallOptions = cloneOptions(m.callbacks.MCPInstallOptions())
			} else {
				m.mcpInstallOptions = nil
			}
			if len(m.mcpInstallOptions) == 0 {
				m.notice = panelRef(Panel{
					Title: m.catalog.Msg("MCP", "MCP"),
					Lines: []string{m.catalog.Msg("没有可用的推荐 MCP。", "No recommended MCP servers are available.")},
				})
				return m, nil
			}
			m.screen = flowScreenMCPInstall
			m.mcpInstallCursor = 0
			m.notice = nil
			return m, nil
		case "remove":
			if m.callbacks.MCPRemoveOptions != nil {
				m.mcpRemoveOptions = cloneOptions(m.callbacks.MCPRemoveOptions(m.selectedMCPTarget))
			} else {
				m.mcpRemoveOptions = nil
			}
			if len(m.mcpRemoveOptions) == 0 {
				m.notice = panelRef(Panel{
					Title: m.catalog.Msg("MCP", "MCP"),
					Lines: []string{m.catalog.Msg("未配置任何 MCP servers。", "No MCP servers configured.")},
				})
				return m, nil
			}
			m.screen = flowScreenMCPRemove
			m.mcpRemoveCursor = 0
			m.notice = nil
			return m, nil
		}
	case flowScreenMCPInstall:
		if len(m.mcpInstallOptions) == 0 {
			return m, nil
		}
		m.selectedMCPServer = m.mcpInstallOptions[m.mcpInstallCursor].Value
		return m.startBusy(flowActionMCPInstall, m.catalog.Msg("正在写入 MCP 配置…", "Writing MCP configuration..."))
	case flowScreenMCPRemove:
		if len(m.mcpRemoveOptions) == 0 {
			return m, nil
		}
		m.selectedMCPServer = m.mcpRemoveOptions[m.mcpRemoveCursor].Value
		return m.startBusy(flowActionMCPRemove, m.catalog.Msg("正在移除 MCP 配置…", "Removing MCP configuration..."))
	case flowScreenSkillTargets:
		if len(m.skillTargets) == 0 {
			return m, nil
		}
		m.selectedSkillTarget = m.skillTargets[m.skillTargetCursor].Value
		m.screen = flowScreenSkillActions
		m.skillActionCursor = 0
		m.notice = nil
		return m, nil
	case flowScreenSkillActions:
		if len(m.skillActions) == 0 {
			return m, nil
		}
		switch m.skillActions[m.skillActionCursor].Value {
		case "list":
			m.selectedSkillValue = ""
			return m.startBusy(flowActionSkillList, m.catalog.Msg("正在读取已安装 skills…", "Reading installed skills..."))
		case "install":
			if m.callbacks.SkillInstallOptions != nil {
				m.skillInstallOptions = cloneOptions(m.callbacks.SkillInstallOptions())
			} else {
				m.skillInstallOptions = nil
			}
			if len(m.skillInstallOptions) == 0 {
				m.notice = panelRef(Panel{
					Title: m.catalog.Msg("Skills", "Skills"),
					Lines: []string{m.catalog.Msg("没有可用的推荐 skills。", "No recommended skills are available.")},
				})
				return m, nil
			}
			m.screen = flowScreenSkillInstall
			m.skillInstallCursor = 0
			m.notice = nil
			return m, nil
		case "uninstall":
			if m.callbacks.SkillUninstallOptions != nil {
				m.skillUninstallOptions = cloneOptions(m.callbacks.SkillUninstallOptions(m.selectedSkillTarget))
			} else {
				m.skillUninstallOptions = nil
			}
			if len(m.skillUninstallOptions) == 0 {
				m.notice = panelRef(Panel{
					Title: m.catalog.Msg("Skills", "Skills"),
					Lines: []string{m.catalog.Msg("未检测到已安装的 skill。", "No installed skills detected.")},
				})
				return m, nil
			}
			m.screen = flowScreenSkillUninstall
			m.skillUninstallCursor = 0
			m.notice = nil
			return m, nil
		}
	case flowScreenSkillInstall:
		if len(m.skillInstallOptions) == 0 {
			return m, nil
		}
		m.selectedSkillValue = m.skillInstallOptions[m.skillInstallCursor].Value
		return m.startBusy(flowActionSkillInstall, m.catalog.Msg("正在安装 skill…", "Installing skill..."))
	case flowScreenSkillUninstall:
		if len(m.skillUninstallOptions) == 0 {
			return m, nil
		}
		m.selectedSkillValue = m.skillUninstallOptions[m.skillUninstallCursor].Value
		return m.startBusy(flowActionSkillUninstall, m.catalog.Msg("正在卸载 skill…", "Uninstalling skill..."))
	case flowScreenInstallHub:
		switch m.installHubOptions[m.installHubCursor].Value {
		case "bootstrap-cli":
			m.bootstrapOptions = m.bootstrapOptionsList()
			if len(m.bootstrapOptions) == 0 {
				m.notice = panelRef(Panel{
					Title: m.catalog.Msg("CLI 安装", "CLI install"),
					Lines: []string{m.catalog.Msg("当前没有可用的 CLI 安装目标。", "There are no CLI install targets available right now.")},
				})
				return m, nil
			}
			m.screen = flowScreenBootstrapTargets
			m.bootstrapCursor = 0
			m.selectedBootstrapTarget = m.bootstrapOptions[0].Value
			m.notice = nil
			m.refreshBootstrapDetail()
			return m, nil
		case "install-agentflow":
			m.installOptions = m.installOptionsList()
			if len(m.installOptions) == 0 {
				m.notice = panelRef(Panel{
					Title: m.catalog.Msg("安装提示", "Install hint"),
					Lines: []string{
						m.catalog.Msg("还没有可安装 AgentFlow 的 CLI。先进入“安装 CLI 工具”分支完成 Codex、Claude 或 Gemini 的安装。", "There are no CLI targets ready for AgentFlow yet. Use the CLI install branch first to install Codex, Claude, or Gemini."),
					},
				})
				return m, nil
			}
			m.screen = flowScreenProfile
			m.notice = nil
			return m, nil
		}
	case flowScreenBootstrapTargets:
		if len(m.bootstrapOptions) == 0 {
			return m, nil
		}
		m.selectedBootstrapTarget = m.bootstrapOptions[m.bootstrapCursor].Value
		m.screen = flowScreenBootstrapActions
		m.bootstrapActionCursor = 0
		if !m.bootstrapAutoSupported(m.selectedBootstrapTarget) && len(m.bootstrapActionOptions) > 1 {
			m.bootstrapActionCursor = 1
		}
		m.notice = nil
		m.refreshBootstrapDetail()
		return m, nil
	case flowScreenBootstrapActions:
		if len(m.bootstrapActionOptions) == 0 {
			return m, nil
		}
		switch m.bootstrapActionOptions[m.bootstrapActionCursor].Value {
		case "auto":
			return m.startBusy(flowActionBootstrapAuto, m.catalog.Msg("正在安装所选 CLI…", "Installing the selected CLI..."))
		case "manual":
			if m.callbacks.BootstrapManual != nil {
				panel := m.callbacks.BootstrapManual(m.selectedBootstrapTarget)
				m.notice = panelRef(panel)
			}
			m.refreshBootstrapDetail()
			return m, nil
		}
	case flowScreenProfile:
		m.selectedProfile = m.profileOptions[m.profileCursor].Value
		m.clearSelections()
		m.screen = flowScreenInstallTargets
		m.notice = nil
		return m, nil
	case flowScreenInstallTargets:
		selected := selectedValues(m.installOptions)
		if len(selected) == 0 {
			m.notice = panelRef(Panel{
				Title: m.catalog.Msg("安装提示", "Install hint"),
				Lines: []string{m.catalog.Msg("请至少选择一个目标。", "Choose at least one target.")},
			})
			return m, nil
		}
		return m.startBusy(flowActionInstall, m.catalog.Msg("正在安装所选目标…", "Installing selected targets..."))
	case flowScreenUninstallTargets:
		selected := selectedValues(m.uninstallOptions)
		if len(selected) == 0 {
			m.notice = panelRef(Panel{
				Title: m.catalog.Msg("卸载提示", "Uninstall hint"),
				Lines: []string{m.catalog.Msg("请至少选择一个目标。", "Choose at least one target.")},
			})
			return m, nil
		}
		if m.uninstallCLIMode {
			return m.startBusy(flowActionUninstallCLI, m.catalog.Msg("正在卸载所选 CLI…", "Uninstalling selected CLIs..."))
		}
		return m.startBusy(flowActionUninstall, m.catalog.Msg("正在卸载所选目标…", "Uninstalling selected targets..."))
	}
	return m, nil
}

func (m interactiveFlowModel) startBusy(action flowAction, message string) (tea.Model, tea.Cmd) {
	m.busy = true
	m.spin = 0
	return m, tea.Batch(m.runActionCmd(action), busyTickCmd())
}

func (m *interactiveFlowModel) moveCursor(delta int) {
	length := m.currentOptionsLen()
	if length == 0 {
		return
	}
	current := m.currentCursor()
	current += delta
	if current < 0 {
		current = 0
	}
	if current > length-1 {
		current = length - 1
	}
	m.setCursor(current)
}

func (m *interactiveFlowModel) setCursor(cursor int) {
	if cursor < 0 {
		cursor = 0
	}
	switch m.screen {
	case flowScreenMain:
		if len(m.mainOptions) == 0 {
			m.mainCursor = 0
			return
		}
		if cursor > len(m.mainOptions)-1 {
			cursor = len(m.mainOptions) - 1
		}
		m.mainCursor = cursor
	case flowScreenInstallHub:
		if len(m.installHubOptions) == 0 {
			m.installHubCursor = 0
			return
		}
		if cursor > len(m.installHubOptions)-1 {
			cursor = len(m.installHubOptions) - 1
		}
		m.installHubCursor = cursor
	case flowScreenMCPTargets:
		if len(m.mcpTargets) == 0 {
			m.mcpTargetCursor = 0
			return
		}
		if cursor > len(m.mcpTargets)-1 {
			cursor = len(m.mcpTargets) - 1
		}
		m.mcpTargetCursor = cursor
		m.selectedMCPTarget = m.mcpTargets[cursor].Value
	case flowScreenMCPActions:
		if len(m.mcpActions) == 0 {
			m.mcpActionCursor = 0
			return
		}
		if cursor > len(m.mcpActions)-1 {
			cursor = len(m.mcpActions) - 1
		}
		m.mcpActionCursor = cursor
	case flowScreenMCPInstall:
		if len(m.mcpInstallOptions) == 0 {
			m.mcpInstallCursor = 0
			return
		}
		if cursor > len(m.mcpInstallOptions)-1 {
			cursor = len(m.mcpInstallOptions) - 1
		}
		m.mcpInstallCursor = cursor
	case flowScreenMCPRemove:
		if len(m.mcpRemoveOptions) == 0 {
			m.mcpRemoveCursor = 0
			return
		}
		if cursor > len(m.mcpRemoveOptions)-1 {
			cursor = len(m.mcpRemoveOptions) - 1
		}
		m.mcpRemoveCursor = cursor
	case flowScreenSkillTargets:
		if len(m.skillTargets) == 0 {
			m.skillTargetCursor = 0
			return
		}
		if cursor > len(m.skillTargets)-1 {
			cursor = len(m.skillTargets) - 1
		}
		m.skillTargetCursor = cursor
		m.selectedSkillTarget = m.skillTargets[cursor].Value
	case flowScreenSkillActions:
		if len(m.skillActions) == 0 {
			m.skillActionCursor = 0
			return
		}
		if cursor > len(m.skillActions)-1 {
			cursor = len(m.skillActions) - 1
		}
		m.skillActionCursor = cursor
	case flowScreenSkillInstall:
		if len(m.skillInstallOptions) == 0 {
			m.skillInstallCursor = 0
			return
		}
		if cursor > len(m.skillInstallOptions)-1 {
			cursor = len(m.skillInstallOptions) - 1
		}
		m.skillInstallCursor = cursor
	case flowScreenSkillUninstall:
		if len(m.skillUninstallOptions) == 0 {
			m.skillUninstallCursor = 0
			return
		}
		if cursor > len(m.skillUninstallOptions)-1 {
			cursor = len(m.skillUninstallOptions) - 1
		}
		m.skillUninstallCursor = cursor
	case flowScreenBootstrapTargets:
		if len(m.bootstrapOptions) == 0 {
			m.bootstrapCursor = 0
			return
		}
		if cursor > len(m.bootstrapOptions)-1 {
			cursor = len(m.bootstrapOptions) - 1
		}
		m.bootstrapCursor = cursor
		m.selectedBootstrapTarget = m.bootstrapOptions[cursor].Value
		m.refreshBootstrapDetail()
	case flowScreenBootstrapActions:
		if len(m.bootstrapActionOptions) == 0 {
			m.bootstrapActionCursor = 0
			return
		}
		if cursor > len(m.bootstrapActionOptions)-1 {
			cursor = len(m.bootstrapActionOptions) - 1
		}
		m.bootstrapActionCursor = cursor
	case flowScreenProfile:
		if len(m.profileOptions) == 0 {
			m.profileCursor = 0
			return
		}
		if cursor > len(m.profileOptions)-1 {
			cursor = len(m.profileOptions) - 1
		}
		m.profileCursor = cursor
	case flowScreenInstallTargets:
		if len(m.installOptions) == 0 {
			m.installCursor = 0
			return
		}
		if cursor > len(m.installOptions)-1 {
			cursor = len(m.installOptions) - 1
		}
		m.installCursor = cursor
	case flowScreenUninstallTargets:
		if len(m.uninstallOptions) == 0 {
			m.uninstallCursor = 0
			return
		}
		if cursor > len(m.uninstallOptions)-1 {
			cursor = len(m.uninstallOptions) - 1
		}
		m.uninstallCursor = cursor
	}
}

func (m interactiveFlowModel) currentCursor() int {
	switch m.screen {
	case flowScreenInstallHub:
		return m.installHubCursor
	case flowScreenMCPTargets:
		return m.mcpTargetCursor
	case flowScreenMCPActions:
		return m.mcpActionCursor
	case flowScreenMCPInstall:
		return m.mcpInstallCursor
	case flowScreenMCPRemove:
		return m.mcpRemoveCursor
	case flowScreenSkillTargets:
		return m.skillTargetCursor
	case flowScreenSkillActions:
		return m.skillActionCursor
	case flowScreenSkillInstall:
		return m.skillInstallCursor
	case flowScreenSkillUninstall:
		return m.skillUninstallCursor
	case flowScreenBootstrapTargets:
		return m.bootstrapCursor
	case flowScreenBootstrapActions:
		return m.bootstrapActionCursor
	case flowScreenProfile:
		return m.profileCursor
	case flowScreenInstallTargets:
		return m.installCursor
	case flowScreenUninstallTargets:
		return m.uninstallCursor
	default:
		return m.mainCursor
	}
}

func (m interactiveFlowModel) currentOptionsLen() int {
	switch m.screen {
	case flowScreenInstallHub:
		return len(m.installHubOptions)
	case flowScreenMCPTargets:
		return len(m.mcpTargets)
	case flowScreenMCPActions:
		return len(m.mcpActions)
	case flowScreenMCPInstall:
		return len(m.mcpInstallOptions)
	case flowScreenMCPRemove:
		return len(m.mcpRemoveOptions)
	case flowScreenSkillTargets:
		return len(m.skillTargets)
	case flowScreenSkillActions:
		return len(m.skillActions)
	case flowScreenSkillInstall:
		return len(m.skillInstallOptions)
	case flowScreenSkillUninstall:
		return len(m.skillUninstallOptions)
	case flowScreenBootstrapTargets:
		return len(m.bootstrapOptions)
	case flowScreenBootstrapActions:
		return len(m.bootstrapActionOptions)
	case flowScreenProfile:
		return len(m.profileOptions)
	case flowScreenInstallTargets:
		return len(m.installOptions)
	case flowScreenUninstallTargets:
		return len(m.uninstallOptions)
	default:
		return len(m.mainOptions)
	}
}

func (m interactiveFlowModel) refreshStatusCmd(withNotice bool) tea.Cmd {
	return func() tea.Msg {
		status := m.callbacks.Status()
		result := flowResultMsg{
			action: flowActionRefreshStatus,
			status: status,
		}
		if withNotice {
			result.notice = panelRef(Panel{
				Title: m.catalog.Msg("状态已刷新", "Status refreshed"),
				Lines: []string{m.catalog.Msg("状态信息已更新。", "Status information has been refreshed.")},
			})
		}
		return result
	}
}

func (m interactiveFlowModel) runActionCmd(action flowAction) tea.Cmd {
	selectedProfile := m.selectedProfile
	selectedInstallTargets := selectedValues(m.installOptions)
	selectedUninstallTargets := selectedValues(m.uninstallOptions)
	selectedBootstrapTarget := m.selectedBootstrapTarget
	selectedMCPTarget := m.selectedMCPTarget
	selectedMCPServer := m.selectedMCPServer
	selectedSkillTarget := m.selectedSkillTarget
	selectedSkillValue := m.selectedSkillValue

	return func() tea.Msg {
		switch action {
		case flowActionRefreshStatus:
			return flowResultMsg{
				action: action,
				status: m.callbacks.Status(),
			}
		case flowActionMCPList:
			notice := Panel{}
			if m.callbacks.MCPList != nil {
				notice = m.callbacks.MCPList(selectedMCPTarget)
			}
			return flowResultMsg{
				action: action,
				notice: panelRef(notice),
				status: m.callbacks.Status(),
			}
		case flowActionMCPInstall:
			notice := Panel{}
			if m.callbacks.MCPInstall != nil {
				notice = m.callbacks.MCPInstall(selectedMCPTarget, selectedMCPServer)
			}
			return flowResultMsg{
				action: action,
				notice: panelRef(notice),
				status: m.callbacks.Status(),
			}
		case flowActionMCPRemove:
			notice := Panel{}
			if m.callbacks.MCPRemove != nil {
				notice = m.callbacks.MCPRemove(selectedMCPTarget, selectedMCPServer)
			}
			return flowResultMsg{
				action: action,
				notice: panelRef(notice),
				status: m.callbacks.Status(),
			}
		case flowActionSkillList:
			notice := Panel{}
			if m.callbacks.SkillList != nil {
				notice = m.callbacks.SkillList(selectedSkillTarget)
			}
			return flowResultMsg{
				action: action,
				notice: panelRef(notice),
				status: m.callbacks.Status(),
			}
		case flowActionSkillInstall:
			notice := Panel{}
			if m.callbacks.SkillInstall != nil {
				notice = m.callbacks.SkillInstall(selectedSkillTarget, selectedSkillValue)
			}
			return flowResultMsg{
				action: action,
				notice: panelRef(notice),
				status: m.callbacks.Status(),
			}
		case flowActionSkillUninstall:
			notice := Panel{}
			if m.callbacks.SkillUninstall != nil {
				notice = m.callbacks.SkillUninstall(selectedSkillTarget, selectedSkillValue)
			}
			return flowResultMsg{
				action: action,
				notice: panelRef(notice),
				status: m.callbacks.Status(),
			}
		case flowActionUpdate:
			notice, version := m.callbacks.Update()
			return flowResultMsg{
				action:  action,
				notice:  panelRef(notice),
				status:  m.callbacks.Status(),
				version: version,
			}
		case flowActionClean:
			notice := m.callbacks.Clean()
			return flowResultMsg{
				action: action,
				notice: panelRef(notice),
				status: m.callbacks.Status(),
			}
		case flowActionBootstrapAuto:
			notice := Panel{}
			if m.callbacks.BootstrapAuto != nil {
				notice = m.callbacks.BootstrapAuto(selectedBootstrapTarget)
			}
			detail := Panel{}
			if m.callbacks.BootstrapDetails != nil {
				detail = m.callbacks.BootstrapDetails(selectedBootstrapTarget)
			}
			return flowResultMsg{
				action:          action,
				notice:          panelRef(notice),
				status:          m.callbacks.Status(),
				bootstrapDetail: panelRef(detail),
			}
		case flowActionInstall:
			notice := m.callbacks.Install(selectedProfile, selectedInstallTargets)
			return flowResultMsg{
				action: action,
				notice: panelRef(notice),
				status: m.callbacks.Status(),
			}
		case flowActionUninstall:
			notice := m.callbacks.Uninstall(selectedUninstallTargets)
			return flowResultMsg{
				action: action,
				notice: panelRef(notice),
				status: m.callbacks.Status(),
			}
		case flowActionUninstallCLI:
			notice := Panel{}
			if m.callbacks.UninstallCLI != nil {
				notice = m.callbacks.UninstallCLI(selectedUninstallTargets)
			}
			return flowResultMsg{
				action: action,
				notice: panelRef(notice),
				status: m.callbacks.Status(),
			}
		default:
			return flowResultMsg{
				action: action,
				status: m.callbacks.Status(),
			}
		}
	}
}

func (m interactiveFlowModel) selectionForCurrentScreen() selectionModel {
	model := selectionModel{
		catalog: m.catalog,
		title:   fmt.Sprintf("AgentFlow v%s", m.version),
		width:   m.width,
		height:  m.height,
	}

	switch m.screen {
	case flowScreenInstallHub:
		model.subtitle = m.catalog.Msg("先决定要安装 CLI 工具，还是把 AgentFlow 写入已经存在的 CLI。Esc 返回主菜单。", "Choose whether to install CLI tools first, or write AgentFlow into CLIs that already exist. Press Esc to return.")
		model.hint = m.catalog.Msg("↑/↓ 切换安装路径，Enter 继续，Esc 返回。", "Use ↑/↓ to choose the install path, Enter to continue, Esc to go back.")
		model.options = cloneOptions(m.installHubOptions)
		model.cursor = m.installHubCursor
		model.panels = m.installHubPanels()
	case flowScreenMCPTargets:
		model.subtitle = m.catalog.Msg("选择要管理 MCP 的目标 CLI。Esc 返回主菜单。", "Choose which CLI target to manage MCP for. Press Esc to return.")
		model.hint = m.catalog.Msg("↑/↓ 切换目标，Enter 继续，Esc 返回。", "Use ↑/↓ to switch targets, Enter to continue, Esc to go back.")
		model.options = cloneOptions(m.mcpTargets)
		model.cursor = m.mcpTargetCursor
		model.panels = m.mcpTargetPanels()
	case flowScreenMCPActions:
		model.subtitle = fmt.Sprintf(m.catalog.Msg("MCP 管理目标: %s。Esc 返回目标列表。", "MCP target: %s. Press Esc to go back."), m.selectedMCPTarget)
		model.hint = m.catalog.Msg("↑/↓ 选择操作，Enter 执行，Esc 返回。", "Use ↑/↓ to choose an action, Enter to run, Esc to go back.")
		model.options = cloneOptions(m.mcpActions)
		model.cursor = m.mcpActionCursor
		model.panels = m.mcpActionPanels()
	case flowScreenMCPInstall:
		model.subtitle = fmt.Sprintf(m.catalog.Msg("为 %s 安装推荐 MCP。Esc 返回操作列表。", "Install recommended MCP for %s. Press Esc to go back."), m.selectedMCPTarget)
		model.hint = m.catalog.Msg("↑/↓ 选择 MCP，Enter 安装，Esc 返回。", "Use ↑/↓ to choose an MCP server, Enter to install, Esc to go back.")
		model.options = cloneOptions(m.mcpInstallOptions)
		model.cursor = m.mcpInstallCursor
		model.panels = m.mcpInstallPanels()
	case flowScreenMCPRemove:
		model.subtitle = fmt.Sprintf(m.catalog.Msg("从 %s 移除 MCP。Esc 返回操作列表。", "Remove MCP from %s. Press Esc to go back."), m.selectedMCPTarget)
		model.hint = m.catalog.Msg("↑/↓ 选择 MCP，Enter 移除，Esc 返回。", "Use ↑/↓ to choose an MCP server, Enter to remove, Esc to go back.")
		model.options = cloneOptions(m.mcpRemoveOptions)
		model.cursor = m.mcpRemoveCursor
		model.panels = m.mcpRemovePanels()
	case flowScreenSkillTargets:
		model.subtitle = m.catalog.Msg("选择要管理 Skill 的目标 CLI。Esc 返回主菜单。", "Choose which CLI target to manage skills for. Press Esc to return.")
		model.hint = m.catalog.Msg("↑/↓ 切换目标，Enter 继续，Esc 返回。", "Use ↑/↓ to switch targets, Enter to continue, Esc to go back.")
		model.options = cloneOptions(m.skillTargets)
		model.cursor = m.skillTargetCursor
		model.panels = m.skillTargetPanels()
	case flowScreenSkillActions:
		model.subtitle = fmt.Sprintf(m.catalog.Msg("Skill 管理目标: %s。Esc 返回目标列表。", "Skill target: %s. Press Esc to go back."), m.selectedSkillTarget)
		model.hint = m.catalog.Msg("↑/↓ 选择操作，Enter 执行，Esc 返回。", "Use ↑/↓ to choose an action, Enter to run, Esc to go back.")
		model.options = cloneOptions(m.skillActions)
		model.cursor = m.skillActionCursor
		model.panels = m.skillActionPanels()
	case flowScreenSkillInstall:
		model.subtitle = fmt.Sprintf(m.catalog.Msg("为 %s 安装推荐 Skill。Esc 返回操作列表。", "Install recommended skills for %s. Press Esc to go back."), m.selectedSkillTarget)
		model.hint = m.catalog.Msg("↑/↓ 选择 skill，Enter 安装，Esc 返回。", "Use ↑/↓ to choose a skill, Enter to install, Esc to go back.")
		model.options = cloneOptions(m.skillInstallOptions)
		model.cursor = m.skillInstallCursor
		model.panels = m.skillInstallPanels()
	case flowScreenSkillUninstall:
		model.subtitle = fmt.Sprintf(m.catalog.Msg("从 %s 卸载 Skill。Esc 返回操作列表。", "Uninstall a skill from %s. Press Esc to go back."), m.selectedSkillTarget)
		model.hint = m.catalog.Msg("↑/↓ 选择 skill，Enter 卸载，Esc 返回。", "Use ↑/↓ to choose a skill, Enter to uninstall, Esc to go back.")
		model.options = cloneOptions(m.skillUninstallOptions)
		model.cursor = m.skillUninstallCursor
		model.panels = m.skillUninstallPanels()
	case flowScreenBootstrapTargets:
		model.subtitle = m.catalog.Msg("选择要安装的 CLI 工具。Esc 返回安装中心。", "Choose which CLI tool to install. Press Esc to return to the install hub.")
		model.hint = m.catalog.Msg("↑/↓ 切换 CLI，Enter 选择安装方式，Esc 返回。", "Use ↑/↓ to switch CLIs, Enter to choose the install mode, Esc to go back.")
		model.options = cloneOptions(m.bootstrapOptions)
		model.cursor = m.bootstrapCursor
		model.panels = m.bootstrapTargetPanels()
	case flowScreenBootstrapActions:
		model.subtitle = m.catalog.Msg("选择自动安装，或先查看当前平台的手动安装提示。Esc 返回 CLI 列表。", "Choose automatic installation, or inspect the manual guidance for this platform first. Press Esc to go back.")
		model.hint = m.catalog.Msg("↑/↓ 切换安装方式，Enter 执行，Esc 返回。", "Use ↑/↓ to choose the install mode, Enter to run, Esc to go back.")
		model.options = cloneOptions(m.bootstrapActionOptions)
		model.cursor = m.bootstrapActionCursor
		model.panels = m.bootstrapActionPanels()
	case flowScreenProfile:
		model.subtitle = m.catalog.Msg("选择部署 Profile。Esc 返回安装中心。", "Select a deployment profile. Press Esc to return to the install hub.")
		model.hint = m.catalog.Msg("↑/↓ 切换 Profile，Enter 下一步，Esc 返回。", "Use ↑/↓ to switch profiles, Enter to continue, Esc to go back.")
		model.options = cloneOptions(m.profileOptions)
		model.cursor = m.profileCursor
		model.panels = m.profilePanels()
	case flowScreenInstallTargets:
		model.subtitle = m.catalog.Msg("选择要安装的目标。Esc 返回 Profile。", "Choose install targets. Press Esc to return to profile.")
		model.hint = m.catalog.Msg("Space 选择多个目标，Enter 安装，Esc 返回。", "Use Space to select multiple targets, Enter to install, Esc to go back.")
		model.options = cloneOptions(m.installOptions)
		model.cursor = m.installCursor
		model.multi = true
		model.panels = m.installPanels()
	case flowScreenUninstallTargets:
		if m.uninstallCLIMode {
			model.subtitle = m.catalog.Msg("选择要卸载的 CLI 工具。Esc 返回主菜单。", "Choose which CLI tools to uninstall. Press Esc to return to the main menu.")
			model.hint = m.catalog.Msg("Space 选择多个目标，Enter 卸载 CLI，Esc 返回。", "Use Space to select multiple targets, Enter to uninstall CLIs, Esc to go back.")
		} else {
			model.subtitle = m.catalog.Msg("选择要卸载的目标。Esc 返回主菜单。", "Choose uninstall targets. Press Esc to return to the main menu.")
			model.hint = m.catalog.Msg("Space 选择多个目标，Enter 卸载，Esc 返回。", "Use Space to select multiple targets, Enter to uninstall, Esc to go back.")
		}
		model.options = cloneOptions(m.uninstallOptions)
		model.cursor = m.uninstallCursor
		model.multi = true
		model.panels = m.uninstallPanels()
	default:
		model.subtitle = m.catalog.Msg("跨平台 Go CLI。所有动作都在同一个 TUI 会话内完成。", "Cross-platform Go CLI. All actions now stay inside one TUI session.")
		model.hint = m.catalog.Msg("↑/↓ 或滚轮切换动作，Enter 执行，Esc 退出。", "Use ↑/↓ or the mouse wheel to change actions, Enter to run, Esc to exit.")
		model.options = cloneOptions(m.mainOptions)
		model.cursor = m.mainCursor
		model.panels = m.mainPanels()
	}

	if m.busy {
		model.hint = m.catalog.Msg("正在执行，请稍候。", "Working, please wait.")
		model.panels = append([]Panel{m.busyPanel()}, model.panels...)
	}

	return model
}

func (m interactiveFlowModel) mainPanels() []Panel {
	panels := make([]Panel, 0, 2)
	if m.notice != nil {
		panels = append(panels, *m.notice)
	}
	panels = append(panels, m.status)
	return panels
}

func (m interactiveFlowModel) installHubPanels() []Panel {
	panels := make([]Panel, 0, 3)
	if m.notice != nil {
		panels = append(panels, *m.notice)
	}
	panels = append(panels, Panel{
		Title: m.catalog.Msg("安装中心", "Install hub"),
		Lines: []string{
			m.catalog.Msg("先安装 CLI 工具时，AgentFlow 会帮你检查 Node / npm / nvm，并在 Windows 上提示 WSL2。", "When installing CLI tools first, AgentFlow checks Node / npm / nvm and warns about WSL2 on Windows."),
			m.catalog.Msg("如果 CLI 已经存在，再进入 AgentFlow 安装分支写入规则、技能和 hooks。", "If the CLI already exists, continue into the AgentFlow install branch to write rules, skills, and hooks."),
		},
	})
	panels = append(panels, m.status)
	return panels
}

func (m interactiveFlowModel) mcpTargetPanels() []Panel {
	panels := make([]Panel, 0, 3)
	if m.notice != nil {
		panels = append(panels, *m.notice)
	}
	panels = append(panels, Panel{
		Title: m.catalog.Msg("MCP 提示", "MCP note"),
		Lines: []string{
			m.catalog.Msg("AgentFlow 会把 MCP servers 配置写入目标 CLI 的配置目录。", "AgentFlow writes MCP server configs into the target CLI config directory."),
			m.catalog.Msg("置顶推荐：Context7（依赖/API 文档）、Playwright（浏览器自动化）、Filesystem（文件操作）。", "Pinned: Context7 (docs), Playwright (browser), Filesystem (files)."),
		},
	})
	panels = append(panels, m.status)
	return panels
}

func (m interactiveFlowModel) mcpActionPanels() []Panel {
	panels := make([]Panel, 0, 2)
	if m.notice != nil {
		panels = append(panels, *m.notice)
	}
	panels = append(panels, Panel{
		Title: m.catalog.Msg("MCP 管理", "MCP management"),
		Lines: []string{
			fmt.Sprintf(m.catalog.Msg("目标: %s", "Target: %s"), m.selectedMCPTarget),
			m.catalog.Msg("提示：Context7 如需更高额度可设置 CONTEXT7_API_KEY（可用 CLI 命令添加 env）。", "Tip: For higher Context7 rate limits, set CONTEXT7_API_KEY (use the CLI command to add env)."),
		},
	})
	return panels
}

func (m interactiveFlowModel) mcpInstallPanels() []Panel {
	panels := make([]Panel, 0, 2)
	if m.notice != nil {
		panels = append(panels, *m.notice)
	}
	panels = append(panels, Panel{
		Title: m.catalog.Msg("安装说明", "Install guide"),
		Lines: []string{
			m.catalog.Msg("选择一个推荐 MCP server 并按 Enter 写入配置。", "Select a recommended MCP server and press Enter to write the config."),
			m.catalog.Msg("Filesystem 默认会把当前工作目录加入 allowlist。", "Filesystem will add the current working directory to its allowlist by default."),
		},
	})
	return panels
}

func (m interactiveFlowModel) mcpRemovePanels() []Panel {
	panels := make([]Panel, 0, 2)
	if m.notice != nil {
		panels = append(panels, *m.notice)
	}
	panels = append(panels, Panel{
		Title: m.catalog.Msg("移除说明", "Remove guide"),
		Lines: []string{
			m.catalog.Msg("选择要移除的 MCP server 并按 Enter。", "Select the MCP server to remove and press Enter."),
		},
	})
	return panels
}

func (m interactiveFlowModel) skillTargetPanels() []Panel {
	panels := make([]Panel, 0, 3)
	if m.notice != nil {
		panels = append(panels, *m.notice)
	}
	panels = append(panels, Panel{
		Title: m.catalog.Msg("Skill 提示", "Skill note"),
		Lines: []string{
			m.catalog.Msg("AgentFlow 会把 skill 安装到目标 CLI 的 skills 目录下。", "AgentFlow installs skills into the target CLI skills directory."),
			m.catalog.Msg("推荐从 skills.sh/GitHub 安装；复杂场景可用 CLI 命令指定 --skill / --ref。", "Use skills.sh/GitHub sources; advanced installs can specify --skill / --ref via CLI."),
		},
	})
	panels = append(panels, m.status)
	return panels
}

func (m interactiveFlowModel) skillActionPanels() []Panel {
	panels := make([]Panel, 0, 2)
	if m.notice != nil {
		panels = append(panels, *m.notice)
	}
	panels = append(panels, Panel{
		Title: m.catalog.Msg("Skill 管理", "Skill management"),
		Lines: []string{
			fmt.Sprintf(m.catalog.Msg("目标: %s", "Target: %s"), m.selectedSkillTarget),
			m.catalog.Msg("提示：要安装任意 skill，可用 `agentflow skill install <target> <skills.sh URL>`。", "Tip: To install any skill, use `agentflow skill install <target> <skills.sh URL>`."),
		},
	})
	return panels
}

func (m interactiveFlowModel) skillInstallPanels() []Panel {
	panels := make([]Panel, 0, 2)
	if m.notice != nil {
		panels = append(panels, *m.notice)
	}
	panels = append(panels, Panel{
		Title: m.catalog.Msg("安装说明", "Install guide"),
		Lines: []string{
			m.catalog.Msg("选择一个推荐 skill 并按 Enter 安装。", "Select a recommended skill and press Enter to install."),
			m.catalog.Msg("如需安装其他 skill，请在终端使用 `agentflow skill install`。", "To install other skills, use `agentflow skill install` in the terminal."),
		},
	})
	return panels
}

func (m interactiveFlowModel) skillUninstallPanels() []Panel {
	panels := make([]Panel, 0, 2)
	if m.notice != nil {
		panels = append(panels, *m.notice)
	}
	panels = append(panels, Panel{
		Title: m.catalog.Msg("卸载说明", "Uninstall guide"),
		Lines: []string{
			m.catalog.Msg("选择要卸载的 skill 并按 Enter。", "Select the skill to uninstall and press Enter."),
		},
	})
	return panels
}

func (m interactiveFlowModel) bootstrapTargetPanels() []Panel {
	panels := make([]Panel, 0, 2)
	if m.notice != nil {
		panels = append(panels, *m.notice)
	}
	if m.bootstrapDetail != nil {
		panels = append(panels, *m.bootstrapDetail)
	}
	return panels
}

func (m interactiveFlowModel) bootstrapActionPanels() []Panel {
	panels := make([]Panel, 0, 3)
	if m.notice != nil {
		panels = append(panels, *m.notice)
	}
	if m.bootstrapDetail != nil {
		panels = append(panels, *m.bootstrapDetail)
	}
	panels = append(panels, Panel{
		Title: m.catalog.Msg("安装方式", "Install modes"),
		Lines: []string{
			m.catalog.Msg("自动安装会优先补齐 nvm / Node，再安装所选 CLI。", "Automatic installation ensures nvm / Node first, then installs the selected CLI."),
			m.catalog.Msg("手动安装提示会给出当前平台对应的命令和建议。", "Manual guidance shows platform-specific commands and recommendations."),
		},
	})
	return panels
}

func (m interactiveFlowModel) profilePanels() []Panel {
	panels := make([]Panel, 0, 2)
	if m.notice != nil {
		panels = append(panels, *m.notice)
	}
	panels = append(panels, Panel{
		Title: m.catalog.Msg("部署说明", "Profile guide"),
		Lines: []string{
			m.catalog.Msg("先选 Profile，再选择要写入 AgentFlow 的 CLI。", "Pick a profile first, then choose which CLIs should receive AgentFlow."),
			m.catalog.Msg("按一次 Esc 就能回到安装中心。", "A single Esc returns to the install hub."),
		},
	})
	return panels
}

func (m interactiveFlowModel) installPanels() []Panel {
	panels := make([]Panel, 0, 2)
	if m.notice != nil {
		panels = append(panels, *m.notice)
	}
	panels = append(panels, Panel{
		Title: m.catalog.Msg("准备安装", "Install plan"),
		Lines: []string{
			fmt.Sprintf(m.catalog.Msg("Profile: %s", "Profile: %s"), m.selectedProfile),
			fmt.Sprintf(m.catalog.Msg("已选择 %d 个目标。", "%d targets selected."), selectedCount(m.installOptions)),
		},
	})
	return panels
}

func (m interactiveFlowModel) uninstallPanels() []Panel {
	panels := make([]Panel, 0, 2)
	if m.notice != nil {
		panels = append(panels, *m.notice)
	}
	if m.uninstallCLIMode {
		panels = append(panels, Panel{
			Title: m.catalog.Msg("准备卸载 CLI", "CLI uninstall plan"),
			Lines: []string{
				fmt.Sprintf(m.catalog.Msg("已选择 %d 个目标。", "%d targets selected."), selectedCount(m.uninstallOptions)),
				m.catalog.Msg("该操作会卸载 CLI 本体，并默认删除配置目录（完整卸载）。", "This will uninstall the CLI tool and purge its config directory by default (full uninstall)."),
				m.catalog.Msg("按 Enter 执行卸载，按 Esc 返回主菜单。", "Press Enter to uninstall, Esc to go back."),
			},
		})
		return panels
	}
	panels = append(panels, Panel{
		Title: m.catalog.Msg("准备卸载", "Uninstall plan"),
		Lines: []string{
			fmt.Sprintf(m.catalog.Msg("已选择 %d 个目标。", "%d targets selected."), selectedCount(m.uninstallOptions)),
			m.catalog.Msg("按 Enter 执行卸载，按 Esc 返回主菜单。", "Press Enter to uninstall, Esc to go back."),
		},
	})
	return panels
}

func (m interactiveFlowModel) busyPanel() Panel {
	return Panel{
		Title: m.catalog.Msg("处理中", "Working"),
		Lines: []string{
			fmt.Sprintf("%s %s", spinnerFrames[m.spin], m.busyMessage()),
		},
	}
}

func (m interactiveFlowModel) busyMessage() string {
	switch {
	case m.screen == flowScreenMCPActions || m.screen == flowScreenMCPInstall || m.screen == flowScreenMCPRemove:
		return m.catalog.Msg("正在更新 MCP 配置…", "Updating MCP configuration...")
	case m.screen == flowScreenSkillActions || m.screen == flowScreenSkillInstall || m.screen == flowScreenSkillUninstall:
		return m.catalog.Msg("正在更新 skills…", "Updating skills...")
	case m.screen == flowScreenBootstrapActions:
		return m.catalog.Msg("正在安装所选 CLI…", "Installing the selected CLI...")
	case m.screen == flowScreenInstallTargets:
		return m.catalog.Msg("正在安装所选目标…", "Installing selected targets...")
	case m.screen == flowScreenUninstallTargets:
		if m.uninstallCLIMode {
			return m.catalog.Msg("正在卸载所选 CLI…", "Uninstalling selected CLIs...")
		}
		return m.catalog.Msg("正在卸载所选目标…", "Uninstalling selected targets...")
	case m.mainOptions[m.mainCursor].Value == string(ActionUpdate):
		return m.catalog.Msg("正在检查最新版本并更新…", "Checking the latest release and updating...")
	case m.mainOptions[m.mainCursor].Value == string(ActionClean):
		return m.catalog.Msg("正在清理缓存…", "Cleaning caches...")
	default:
		return m.catalog.Msg("正在刷新状态…", "Refreshing status...")
	}
}

func (m *interactiveFlowModel) refreshBootstrapDetail() {
	if m.callbacks.BootstrapDetails == nil {
		m.bootstrapDetail = nil
		return
	}
	target := strings.TrimSpace(m.selectedBootstrapTarget)
	if target == "" && m.screen == flowScreenBootstrapTargets && len(m.bootstrapOptions) > 0 {
		target = m.bootstrapOptions[m.bootstrapCursor].Value
	}
	if target == "" {
		m.bootstrapDetail = nil
		return
	}
	panel := m.callbacks.BootstrapDetails(target)
	m.bootstrapDetail = panelRef(panel)
}

func (m interactiveFlowModel) bootstrapAutoSupported(target string) bool {
	if m.callbacks.BootstrapAutoSupported == nil {
		return true
	}
	return m.callbacks.BootstrapAutoSupported(target)
}

func (m interactiveFlowModel) bootstrapOptionsList() []Option {
	if m.callbacks.BootstrapOptions == nil {
		return nil
	}
	options := cloneOptions(m.callbacks.BootstrapOptions())
	if len(options) == 0 {
		return nil
	}
	return options
}

func (m interactiveFlowModel) installOptionsList() []Option {
	if m.callbacks.InstallOptions == nil {
		return nil
	}
	return cloneOptions(m.callbacks.InstallOptions())
}

func (m interactiveFlowModel) uninstallOptionsList() []Option {
	if m.callbacks.UninstallOptions == nil {
		return nil
	}
	return cloneOptions(m.callbacks.UninstallOptions())
}

func (m *interactiveFlowModel) toggleSelected(options *[]Option, cursor int) {
	if options == nil || cursor < 0 || cursor >= len(*options) {
		return
	}
	(*options)[cursor].Selected = !(*options)[cursor].Selected
}

func (m *interactiveFlowModel) clearSelections() {
	for index := range m.installOptions {
		m.installOptions[index].Selected = false
	}
	for index := range m.uninstallOptions {
		m.uninstallOptions[index].Selected = false
	}
}

func cloneOptions(options []Option) []Option {
	if len(options) == 0 {
		return nil
	}
	cloned := make([]Option, len(options))
	copy(cloned, options)
	return cloned
}

func panelRef(panel Panel) *Panel {
	return &panel
}

func busyTickCmd() tea.Cmd {
	return tea.Tick(120*time.Millisecond, func(time.Time) tea.Msg {
		return flowTickMsg{}
	})
}

var spinnerFrames = []string{"·", "••", "•••"}
