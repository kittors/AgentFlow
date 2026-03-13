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
	flowScreenBootstrapTargets
	flowScreenBootstrapActions
	flowScreenProfile
	flowScreenInstallTargets
	flowScreenUninstallTargets
)

type flowAction int

const (
	flowActionRefreshStatus flowAction = iota
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
	bootstrapOptions       []Option
	bootstrapActionOptions []Option
	profileOptions         []Option
	installOptions         []Option
	uninstallOptions       []Option
	uninstallCLIMode       bool

	mainCursor            int
	installHubCursor      int
	bootstrapCursor       int
	bootstrapActionCursor int
	profileCursor         int
	installCursor         int
	uninstallCursor       int

	selectedProfile         string
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

	return func() tea.Msg {
		switch action {
		case flowActionRefreshStatus:
			return flowResultMsg{
				action: action,
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
