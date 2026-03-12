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
	Status           func() Panel
	InstallOptions   func() []Option
	UninstallOptions func() []Option
	Install          func(profile string, targets []string) Panel
	Uninstall        func(targets []string) Panel
	Update           func() (Panel, string)
	Clean            func() Panel
}

type flowScreen int

const (
	flowScreenMain flowScreen = iota
	flowScreenProfile
	flowScreenInstallTargets
	flowScreenUninstallTargets
)

type flowAction int

const (
	flowActionRefreshStatus flowAction = iota
	flowActionUpdate
	flowActionClean
	flowActionInstall
	flowActionUninstall
)

type flowResultMsg struct {
	action  flowAction
	notice  *Panel
	status  Panel
	version string
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

	mainOptions      []Option
	profileOptions   []Option
	installOptions   []Option
	uninstallOptions []Option

	mainCursor      int
	profileCursor   int
	installCursor   int
	uninstallCursor int

	selectedProfile string
	notice          *Panel
	status          Panel
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
				Label:       catalog.Msg("安装到 CLI", "Install to CLI targets"),
				Badge:       catalog.Msg("安装", "SETUP"),
				Description: catalog.Msg("把 AgentFlow 规则、模块、技能和 hooks 部署到 Codex、Claude、Gemini 等 CLI。", "Deploy AgentFlow rules, modules, skills, and hooks into Codex, Claude, Gemini, and other CLIs."),
			},
			{
				Value:       string(ActionUninstall),
				Label:       catalog.Msg("卸载已安装目标", "Uninstall from installed targets"),
				Badge:       catalog.Msg("移除", "REMOVE"),
				Description: catalog.Msg("从已接入 CLI 中清理 AgentFlow 产物，同时保留你的原有配置。", "Remove AgentFlow assets from integrated CLIs while preserving your own config where possible."),
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
		switch value.action {
		case flowActionRefreshStatus:
			if value.notice != nil {
				m.notice = value.notice
			}
		case flowActionInstall, flowActionUninstall, flowActionUpdate, flowActionClean:
			m.screen = flowScreenMain
			m.installOptions = nil
			m.uninstallOptions = nil
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
		switch value.Button {
		case tea.MouseButtonWheelUp:
			m.moveCursor(-1)
		case tea.MouseButtonWheelDown:
			m.moveCursor(1)
		}
		return m, nil
	case tea.KeyMsg:
		switch value.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
		if m.busy {
			return m, nil
		}
		return m.handleKey(value.String())
	}

	return m, nil
}

func (m interactiveFlowModel) View() string {
	screen := m.selectionForCurrentScreen()
	return screen.View()
}

func (m interactiveFlowModel) handleKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up":
		m.moveCursor(-1)
		return m, nil
	case "down":
		m.moveCursor(1)
		return m, nil
	case "pgup":
		m.moveCursor(-5)
		return m, nil
	case "pgdown":
		m.moveCursor(5)
		return m, nil
	case "home":
		m.setCursor(0)
		return m, nil
	case "end":
		m.setCursor(m.currentOptionsLen() - 1)
		return m, nil
	case " ":
		if m.screen == flowScreenInstallTargets {
			m.toggleSelected(&m.installOptions, m.installCursor)
		}
		if m.screen == flowScreenUninstallTargets {
			m.toggleSelected(&m.uninstallOptions, m.uninstallCursor)
		}
		return m, nil
	case "esc":
		return m.handleBack()
	case "enter":
		return m.handleEnter()
	}
	return m, nil
}

func (m interactiveFlowModel) handleBack() (tea.Model, tea.Cmd) {
	switch m.screen {
	case flowScreenMain:
		return m, tea.Quit
	case flowScreenProfile:
		m.screen = flowScreenMain
	case flowScreenInstallTargets:
		m.screen = flowScreenProfile
	case flowScreenUninstallTargets:
		m.screen = flowScreenMain
	}
	m.notice = nil
	return m, nil
}

func (m interactiveFlowModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.screen {
	case flowScreenMain:
		switch Action(m.mainOptions[m.mainCursor].Value) {
		case ActionInstall:
			m.installOptions = cloneOptions(m.callbacks.InstallOptions())
			if len(m.installOptions) == 0 {
				m.notice = panelRef(Panel{
					Title: m.catalog.Msg("安装结果", "Install result"),
					Lines: []string{m.catalog.Msg("未检测到任何已安装的 CLI。", "No installed CLIs detected.")},
				})
				return m, nil
			}
			m.screen = flowScreenProfile
			m.notice = nil
			return m, nil
		case ActionUninstall:
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
		case ActionUpdate:
			return m.startBusy(flowActionUpdate, m.catalog.Msg("正在检查最新版本并更新…", "Checking the latest release and updating..."))
		case ActionStatus:
			return m.startBusy(flowActionRefreshStatus, m.catalog.Msg("正在刷新状态…", "Refreshing status..."))
		case ActionClean:
			return m.startBusy(flowActionClean, m.catalog.Msg("正在清理缓存…", "Cleaning caches..."))
		case ActionExit:
			return m, tea.Quit
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
		return m.startBusy(flowActionUninstall, m.catalog.Msg("正在卸载所选目标…", "Uninstalling selected targets..."))
	}
	return m, nil
}

func (m interactiveFlowModel) startBusy(action flowAction, message string) (tea.Model, tea.Cmd) {
	m.busy = true
	m.spin = 0
	return m, tea.Batch(m.runActionCmd(action), busyTickCmd())
}

func (m interactiveFlowModel) moveCursor(delta int) {
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

func (m interactiveFlowModel) setCursor(cursor int) {
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
	case flowScreenProfile:
		model.subtitle = m.catalog.Msg("选择部署 Profile。Esc 返回主菜单。", "Select a deployment profile. Press Esc to return.")
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
		model.subtitle = m.catalog.Msg("选择要卸载的目标。Esc 返回主菜单。", "Choose uninstall targets. Press Esc to return to the main menu.")
		model.hint = m.catalog.Msg("Space 选择多个目标，Enter 卸载，Esc 返回。", "Use Space to select multiple targets, Enter to uninstall, Esc to go back.")
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

func (m interactiveFlowModel) profilePanels() []Panel {
	panels := make([]Panel, 0, 2)
	if m.notice != nil {
		panels = append(panels, *m.notice)
	}
	panels = append(panels, Panel{
		Title: m.catalog.Msg("部署说明", "Profile guide"),
		Lines: []string{
			m.catalog.Msg("先选 Profile，再选择要写入 AgentFlow 的 CLI。", "Pick a profile first, then choose which CLIs should receive AgentFlow."),
			m.catalog.Msg("按一次 Esc 就能回到主菜单。", "A single Esc returns to the main menu."),
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
	case m.screen == flowScreenInstallTargets:
		return m.catalog.Msg("正在安装所选目标…", "Installing selected targets...")
	case m.screen == flowScreenUninstallTargets:
		return m.catalog.Msg("正在卸载所选目标…", "Uninstalling selected targets...")
	case m.mainOptions[m.mainCursor].Value == string(ActionUpdate):
		return m.catalog.Msg("正在检查最新版本并更新…", "Checking the latest release and updating...")
	case m.mainOptions[m.mainCursor].Value == string(ActionClean):
		return m.catalog.Msg("正在清理缓存…", "Cleaning caches...")
	default:
		return m.catalog.Msg("正在刷新状态…", "Refreshing status...")
	}
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
