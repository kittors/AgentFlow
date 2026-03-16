package ui

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/kittors/AgentFlow/internal/debuglog"
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
	SkillGlobalSupported   func(target string) bool
	SkillInstallOptions    func(target string) []Option
	SkillUninstallOptions  func(target string) []Option
	SkillList              func(target string) Panel
	SkillInstall           func(target, source string) Panel
	SkillUninstall         func(target, name string) Panel
	ProjectRulesPanel      func(root, target string) Panel
	ProjectRulesInstall    func(root, target, profile string) Panel
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
	Update                 func(progress func(stage string, percent int, info string)) (Panel, string)
	Clean                  func() Panel
	CLIConfigFields        func(target string) []ConfigField
	WriteEnvConfig         func(envVars map[string]string) Panel
}

// ConfigField describes a single configurable field for a CLI.
type ConfigField struct {
	Label   string   // e.g. "API Key"
	EnvVar  string   // e.g. "OPENAI_API_KEY" or "__MODEL__" for non-env fields
	Type    string   // "text" for free-form input, "select" for option list
	Options []string // For "select" type: available choices
	Default string   // Default value (pre-selected for "select")
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
	flowScreenSkillScope
	flowScreenSkillProjectActions
	flowScreenSkillProjectProfile
	flowScreenSkillActions
	flowScreenSkillInstall
	flowScreenSkillUninstall
	flowScreenBootstrapTargets
	flowScreenBootstrapActions
	flowScreenProfile
	flowScreenInstallTargets
	flowScreenUninstallTargets
	flowScreenUpdateConfirm
	flowScreenBootstrapConfig
)

type flowAction int

const (
	flowActionRefreshStatus flowAction = iota
	flowActionMCPRefreshSummary
	flowActionMCPList
	flowActionMCPInstall
	flowActionMCPRemove
	flowActionSkillRefreshSummary
	flowActionInstallHubRefresh
	flowActionSkillList
	flowActionSkillLoadInstallOptions
	flowActionSkillInstall
	flowActionSkillUninstall
	flowActionProjectRulesInstall
	flowActionUpdate
	flowActionClean
	flowActionBootstrapAuto
	flowActionInstall
	flowActionUninstall
	flowActionUninstallCLI
	flowActionWriteEnvConfig
)

type flowResultMsg struct {
	action          flowAction
	notice          *Panel
	status          Panel
	version         string
	bootstrapDetail *Panel
	projectRules    *Panel
	mcpSummary      *Panel
	skillSummary    *Panel
	skillOptions    []Option
}

type flowTickMsg struct{}

// updateProgressState holds thread-safe progress info that is written by the
// update goroutine and polled by the busy tick.
type updateProgressState struct {
	mu      sync.Mutex
	stage   string // e.g. "checking", "found:1.2.3", "downloading:1.2.3"
	percent int    // 0–100 for download, -1 for indeterminate
}

type interactiveFlowModel struct {
	catalog   i18n.Catalog
	version   string
	callbacks InteractiveCallbacks

	width  int
	height int

	screen         flowScreen
	busy           bool
	spin           int
	activeAction   flowAction // which action owns the current busy state
	initLoading    bool       // true while Init's refreshStatusCmd is in flight
	focusDetails   bool
	detailScroll   int
	updateProgress *updateProgressState // shared progress polled by tick

	mainOptions            []Option
	installHubOptions      []Option
	projectRoot            string
	mcpTargets             []Option
	mcpActions             []Option
	mcpInstallOptions      []Option
	mcpRemoveOptions       []Option
	skillTargets           []Option
	skillScopeOptions      []Option
	skillProjectActions    []Option
	skillActions           []Option
	skillInstallOptions    []Option
	skillUninstallOptions  []Option
	bootstrapOptions       []Option
	bootstrapActionOptions []Option
	profileOptions         []Option
	installOptions         []Option
	uninstallOptions       []Option
	uninstallCLIMode       bool
	projectInstallMode     bool
	updateConfirmOptions   []Option
	updateConfirmCursor    int
	installHubStatusPanel  *Panel

	// Bootstrap config screen state.
	configFields        []configFieldState
	pendingConfigFields []ConfigField // cached after bootstrap; shown when user chooses "configure"
	configFieldCursor   int
	configEditing       bool
	configTarget        string

	mainCursor               int
	installHubCursor         int
	mcpTargetCursor          int
	mcpActionCursor          int
	mcpInstallCursor         int
	mcpRemoveCursor          int
	skillTargetCursor        int
	skillScopeCursor         int
	skillProjectActionCursor int
	skillActionCursor        int
	skillInstallCursor       int
	skillUninstallCursor     int
	bootstrapCursor          int
	bootstrapActionCursor    int
	profileCursor            int
	installCursor            int
	uninstallCursor          int

	selectedProfile         string
	selectedMCPTarget       string
	selectedMCPServer       string
	selectedSkillTarget     string
	selectedSkillScope      string
	selectedProjectProfile  string
	selectedSkillValue      string
	selectedBootstrapTarget string
	notice                  *Panel
	status                  Panel
	bootstrapDetail         *Panel
	projectRulesDetail      *Panel
	mcpSummary              *Panel
	skillSummary            *Panel
}

func (m *interactiveFlowModel) resetDetailFocus() {
	m.focusDetails = false
	m.detailScroll = 0
}

// configFieldState holds the editing state for a single config field.
type configFieldState struct {
	Label        string   // e.g. "API Key"
	EnvVar       string   // e.g. "OPENAI_API_KEY"
	Value        string   // current typed value (for text type)
	FieldType    string   // "text" or "select"
	Options      []string // available choices (for select type)
	OptionCursor int      // currently selected option index (for select type)
	Dirty        bool     // true if user explicitly modified this field
}

func normalizeIdentifier(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func withRecommendedPrefix(label string) string {
	if strings.TrimSpace(label) == "" {
		return label
	}
	return "★ " + label
}

func (m interactiveFlowModel) installedMCPSet(target string) map[string]bool {
	installed := map[string]bool{}
	if m.callbacks.MCPRemoveOptions == nil {
		return installed
	}
	for _, option := range m.callbacks.MCPRemoveOptions(target) {
		key := normalizeIdentifier(option.Value)
		if key == "" {
			continue
		}
		installed[key] = true
	}
	return installed
}

func (m interactiveFlowModel) installedSkillSet(target string) map[string]bool {
	installed := map[string]bool{}
	if m.callbacks.SkillUninstallOptions == nil {
		return installed
	}
	for _, option := range m.callbacks.SkillUninstallOptions(target) {
		key := normalizeIdentifier(option.Value)
		if key == "" {
			continue
		}
		installed[key] = true
	}
	return installed
}

func indexOfOptionValue(options []Option, value string) int {
	value = strings.TrimSpace(value)
	if value == "" {
		return -1
	}
	for index, option := range options {
		if strings.EqualFold(strings.TrimSpace(option.Value), value) {
			return index
		}
	}
	return -1
}

func (m interactiveFlowModel) annotateRecommendedMCPOptions(target string, options []Option) []Option {
	installed := m.installedMCPSet(target)
	updated := cloneOptions(options)
	for index := range updated {
		key := normalizeIdentifier(updated[index].Value)
		if key == "" {
			key = normalizeIdentifier(updated[index].Label)
		}
		isInstalled := installed[key]
		isRecommended := strings.EqualFold(strings.TrimSpace(updated[index].Badge), "PIN")

		if isRecommended {
			updated[index].Label = withRecommendedPrefix(updated[index].Label)
		}

		if isInstalled {
			updated[index].Badge = "✓"
			if strings.TrimSpace(updated[index].Description) != "" {
				updated[index].Description = updated[index].Description + " " + m.catalog.Msg("（已安装：Enter 将更新配置）", "(installed: Enter updates config)")
			} else {
				updated[index].Description = m.catalog.Msg("已安装：Enter 将更新配置。", "Installed: Enter updates config.")
			}
		} else {
			updated[index].Badge = "+"
			if strings.TrimSpace(updated[index].Description) != "" {
				updated[index].Description = updated[index].Description + " " + m.catalog.Msg("（未安装：Enter 安装）", "(not installed: Enter installs)")
			} else {
				updated[index].Description = m.catalog.Msg("未安装：Enter 安装。", "Not installed: Enter installs.")
			}
		}
	}
	return updated
}

func (m interactiveFlowModel) annotateRecommendedSkillOptions(target string, options []Option) []Option {
	installed := m.installedSkillSet(target)
	updated := cloneOptions(options)
	for index := range updated {
		key := normalizeIdentifier(updated[index].Label)
		isInstalled := installed[key]
		isRecommended := strings.EqualFold(strings.TrimSpace(updated[index].Badge), "PIN")

		if isRecommended {
			updated[index].Label = withRecommendedPrefix(updated[index].Label)
		}

		if isInstalled {
			updated[index].Badge = "✓"
			if strings.TrimSpace(updated[index].Description) != "" {
				updated[index].Description = updated[index].Description + " " + m.catalog.Msg("（已安装：Enter 将更新/重装）", "(installed: Enter reinstalls/updates)")
			} else {
				updated[index].Description = m.catalog.Msg("已安装：Enter 将更新/重装。", "Installed: Enter reinstalls/updates.")
			}
		} else {
			updated[index].Badge = "+"
			if strings.TrimSpace(updated[index].Description) != "" {
				updated[index].Description = updated[index].Description + " " + m.catalog.Msg("（未安装：Enter 安装）", "(not installed: Enter installs)")
			} else {
				updated[index].Description = m.catalog.Msg("未安装：Enter 安装。", "Not installed: Enter installs.")
			}
		}
	}
	return updated
}

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
				Value:       string(ActionInstall),
				Label:       catalog.Msg("安装 / 卸载", "Install / Uninstall"),
				Badge:       catalog.Msg("安装", "SETUP"),
				Description: catalog.Msg("安装或卸载 CLI 工具、AgentFlow 全局/项目级配置。", "Install or uninstall CLI tools and AgentFlow global/project configurations."),
			},
			{
				Value:       string(ActionMCP),
				Label:       catalog.Msg("管理 MCP Servers", "Manage MCP servers"),
				Badge:       catalog.Msg("MCP", "MCP"),
				Description: catalog.Msg("为任意 CLI/IDE 写入、查看与移除 MCP servers 配置，并置顶推荐 Context7 / Playwright / Filesystem / Tavily。", "Write, inspect, and remove MCP server configs for any CLI/IDE, with pinned recommendations like Context7 / Playwright / Filesystem / Tavily."),
			},
			{
				Value:       string(ActionSkill),
				Label:       catalog.Msg("管理 Skills", "Manage skills"),
				Badge:       catalog.Msg("SKILL", "SKILL"),
				Description: catalog.Msg("为任意 CLI 安装、查看与卸载 skills（支持从 skills.sh/GitHub 安装）。", "Install, inspect, and uninstall skills for any CLI (supports skills.sh/GitHub sources)."),
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
				Label:       catalog.Msg("安装 AgentFlow 到已安装 CLI（全局）", "Install AgentFlow into existing CLIs (global)"),
				Badge:       catalog.Msg("全局", "GLOBAL"),
				Description: catalog.Msg("对已经存在的 CLI 写入 AgentFlow 规则、模块、技能和 hooks（写入用户级配置目录）。", "Write AgentFlow rules, modules, skills, and hooks into CLIs that already exist (user-level config directory)."),
			},
			{
				Value:       "install-project",
				Label:       catalog.Msg("安装 AgentFlow 到当前项目（项目级）", "Install AgentFlow into current project (project-level)"),
				Badge:       catalog.Msg("项目", "PROJECT"),
				Description: catalog.Msg("将 AgentFlow 规则文件写入当前工作目录，适合团队协作和项目级配置。", "Write AgentFlow rule files into the current working directory, ideal for team collaboration and project-level configuration."),
			},
			{
				Value:       "uninstall-agentflow",
				Label:       catalog.Msg("卸载 AgentFlow（保留 CLI）", "Uninstall AgentFlow (keep CLIs)"),
				Badge:       catalog.Msg("卸载", "REMOVE"),
				Description: catalog.Msg("从已接入 CLI 中清理 AgentFlow 产物，同时保留你的原有配置。", "Remove AgentFlow assets from integrated CLIs while preserving your own config where possible."),
			},
			{
				Value:       "uninstall-cli",
				Label:       catalog.Msg("卸载 CLI 工具（完整卸载）", "Uninstall CLI tools (full removal)"),
				Badge:       catalog.Msg("CLI", "CLI"),
				Description: catalog.Msg("卸载 Codex / Claude / Gemini 等 CLI 本体，并默认删除配置目录。", "Uninstall CLI tools like Codex / Claude / Gemini and purge their config directories by default."),
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

func (m interactiveFlowModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch value := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = value.Width
		m.height = value.Height
		return m, nil
	case flowTickMsg:
		// Keep the spinner alive while busy OR while init is loading.
		if !m.busy && !m.initLoading {
			return m, nil
		}
		m.spin = (m.spin + 1) % len(spinnerFrames)
		return m, busyTickCmd()
	case flowResultMsg:
		debuglog.Log("[MSG] flowResultMsg action=%d busy=%v activeAction=%d initLoading=%v", value.action, m.busy, m.activeAction, m.initLoading)
		// Only clear busy state if this result matches the active action,
		// preventing a stale init-status result from killing an update spinner.
		if value.action == flowActionRefreshStatus && m.initLoading {
			m.initLoading = false
			// If no other action is running, also clear busy.
			if m.activeAction == flowActionRefreshStatus || m.activeAction == 0 {
				m.busy = false
				m.spin = 0
			}
		} else if value.action == m.activeAction {
			m.busy = false
			m.spin = 0
			m.activeAction = 0
		}
		// Always apply data from the result regardless of busy state.
		if strings.TrimSpace(value.version) != "" {
			m.version = value.version
		}
		if strings.TrimSpace(value.status.Title) != "" || len(value.status.Lines) > 0 {
			m.status = value.status
		}
		if value.bootstrapDetail != nil {
			m.bootstrapDetail = value.bootstrapDetail
		}
		if value.projectRules != nil {
			m.projectRulesDetail = value.projectRules
		}
		if value.mcpSummary != nil {
			m.mcpSummary = value.mcpSummary
		}
		if value.skillSummary != nil {
			m.skillSummary = value.skillSummary
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
		case flowActionProjectRulesInstall:
			if value.notice != nil {
				m.notice = value.notice
			}
			if m.projectInstallMode {
				m.projectInstallMode = false
				m.screen = flowScreenInstallHub
			} else {
				m.screen = flowScreenSkillProjectActions
			}
			m.resetDetailFocus()
		case flowActionInstallHubRefresh:
			if value.notice != nil {
				m.installHubStatusPanel = value.notice
			}
		case flowActionMCPList, flowActionMCPInstall, flowActionMCPRemove:
			if value.notice != nil {
				m.notice = value.notice
			}
			m.screen = flowScreenMCPActions
			m.resetDetailFocus()
			m.mcpInstallOptions = nil
			m.mcpRemoveOptions = nil
		case flowActionSkillRefreshSummary:
			if value.notice != nil {
				m.notice = value.notice
			}
		case flowActionSkillLoadInstallOptions:
			m.skillInstallOptions = value.skillOptions
			if len(m.skillInstallOptions) == 0 {
				m.notice = panelRef(Panel{
					Title: m.catalog.Msg("Skills", "Skills"),
					Lines: []string{m.catalog.Msg("没有可用的推荐 skills。", "No recommended skills are available.")},
				})
				return m, nil
			}
			m.screen = flowScreenSkillInstall
			m.skillInstallCursor = 0
			m.resetDetailFocus()
		case flowActionSkillList, flowActionSkillInstall, flowActionSkillUninstall:
			if value.notice != nil {
				m.notice = value.notice
			}
			m.screen = flowScreenSkillActions
			m.resetDetailFocus()
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
			// After successful bootstrap, offer post-install choices instead
			// of auto-entering the config screen. The user can choose to
			// configure custom API key / Base URL / Model, or exit and use
			// the official CLI setup flow.
			var hasConfig bool
			if m.callbacks.CLIConfigFields != nil {
				fields := m.callbacks.CLIConfigFields(m.selectedBootstrapTarget)
				if len(fields) > 0 {
					m.pendingConfigFields = fields
					hasConfig = true
				}
			}
			// Replace bootstrap action options with post-install choices.
			postOpts := make([]Option, 0, 2)
			if hasConfig {
				postOpts = append(postOpts, Option{
					Value:       "configure",
					Label:       m.catalog.Msg("自定义配置（API Key / Base URL / 模型）", "Custom configuration (API Key / Base URL / Model)"),
					Badge:       m.catalog.Msg("配置", "CONFIG"),
					Description: m.catalog.Msg("手动输入 API Key、Base URL、选择模型等。适用于自建网关或第三方 API 代理。", "Manually enter API Key, Base URL, and select a model. Useful for self-hosted gateways or third-party API proxies."),
				})
			}
			postOpts = append(postOpts, Option{
				Value:       "done",
				Label:       m.catalog.Msg("完成（使用官方默认配置）", "Done (use official default configuration)"),
				Badge:       m.catalog.Msg("完成", "DONE"),
				Description: m.catalog.Msg("退出安装向导。您可以稍后启动 CLI 并通过官方流程完成配置。", "Exit the installer. You can launch the CLI later and complete setup via the official flow."),
			})
			m.bootstrapActionOptions = postOpts
			m.bootstrapActionCursor = 0
			m.screen = flowScreenBootstrapActions
		case flowActionWriteEnvConfig:
			if value.notice != nil {
				m.notice = value.notice
			}
			m.screen = flowScreenBootstrapActions
			m.configEditing = false
		case flowActionInstall, flowActionUninstall, flowActionUninstallCLI, flowActionClean:
			m.screen = flowScreenMain
			m.resetDetailFocus()
			m.installOptions = nil
			m.uninstallOptions = nil
			m.uninstallCLIMode = false
			m.clearSelections()
			if value.notice != nil {
				m.notice = value.notice
			}
		case flowActionUpdate:
			m.updateProgress = nil
			if value.notice != nil {
				m.notice = value.notice
			}
			if strings.TrimSpace(value.version) != "" {
				// Update succeeded with new version: show restart confirmation.
				m.updateConfirmOptions = []Option{
					{
						Value:       "restart",
						Label:       m.catalog.Msg("立即重启", "Restart now"),
						Badge:       m.catalog.Msg("推荐", "REC"),
						Description: m.catalog.Msg("立即重启 AgentFlow 以使用新版本。", "Restart AgentFlow immediately to use the new version."),
					},
					{
						Value:       "cancel",
						Label:       m.catalog.Msg("稍后手动重启", "Restart later"),
						Badge:       "",
						Description: m.catalog.Msg("返回主菜单，稍后手动重启 agentflow。", "Go back to the main menu; restart agentflow manually later."),
					},
				}
				m.updateConfirmCursor = 0
				m.screen = flowScreenUpdateConfirm
				m.resetDetailFocus()
			} else {
				// Already on latest or update failed: just go back to main.
				m.screen = flowScreenMain
				m.resetDetailFocus()
			}
		}
		return m, nil
	case tea.MouseMsg:
		if m.busy {
			return m, nil
		}
		switch {
		case value.Button == tea.MouseButtonWheelUp || value.Type == tea.MouseWheelUp:
			if m.focusDetails {
				m.detailScroll--
			} else {
				m.moveCursor(-1)
			}
		case value.Button == tea.MouseButtonWheelDown || value.Type == tea.MouseWheelDown:
			if m.focusDetails {
				m.detailScroll++
			} else {
				m.moveCursor(1)
			}
		}
		return m, nil
	case tea.KeyMsg:
		if value.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		if m.busy {
			// Allow navigation AND action keys during busy state.
			// Starting a new action (Enter/Space) will override the active
			// action via startBusy; the old goroutine's result is still
			// applied for data but won't clear the busy state because
			// activeAction no longer matches.
			switch value.Type {
			case tea.KeyUp, tea.KeyDown, tea.KeyLeft, tea.KeyRight,
				tea.KeyTab, tea.KeyPgUp, tea.KeyPgDown, tea.KeyHome, tea.KeyEnd,
				tea.KeyEsc:
				return m.handleBusyNavKey(value)
			case tea.KeyEnter, tea.KeySpace:
				// Fall through to handleKey so a new action can be started.
			default:
				return m, nil
			}
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
	// Config screen intercepts all keys for text input.
	if m.screen == flowScreenBootstrapConfig && m.configEditing {
		return m.handleConfigKey(key)
	}
	switch key.Type {
	case tea.KeyLeft:
		m.focusDetails = false
		return m, nil
	case tea.KeyRight:
		m.focusDetails = true
		return m, nil
	case tea.KeyTab:
		m.focusDetails = !m.focusDetails
		return m, nil
	case tea.KeyUp:
		prev := m.currentCursor()
		if m.focusDetails {
			m.detailScroll--
		} else {
			m.moveCursor(-1)
		}
		if !m.focusDetails && m.currentCursor() != prev {
			switch m.screen {
			case flowScreenSkillTargets:
				return m.startBusy(flowActionSkillRefreshSummary, m.catalog.Msg("正在读取项目/全局 Skill 信息…", "Loading project/global skill status..."))
			case flowScreenMCPTargets:
				return m.startBusy(flowActionMCPRefreshSummary, m.catalog.Msg("正在读取 MCP 配置…", "Reading MCP configuration..."))
			}
		}
		return m, nil
	case tea.KeyDown:
		prev := m.currentCursor()
		if m.focusDetails {
			m.detailScroll++
		} else {
			m.moveCursor(1)
		}
		if !m.focusDetails && m.currentCursor() != prev {
			switch m.screen {
			case flowScreenSkillTargets:
				return m.startBusy(flowActionSkillRefreshSummary, m.catalog.Msg("正在读取项目/全局 Skill 信息…", "Loading project/global skill status..."))
			case flowScreenMCPTargets:
				return m.startBusy(flowActionMCPRefreshSummary, m.catalog.Msg("正在读取 MCP 配置…", "Reading MCP configuration..."))
			}
		}
		return m, nil
	case tea.KeyPgUp:
		prev := m.currentCursor()
		if m.focusDetails {
			m.detailScroll -= 5
		} else {
			m.moveCursor(-5)
		}
		if !m.focusDetails && m.currentCursor() != prev {
			switch m.screen {
			case flowScreenSkillTargets:
				return m.startBusy(flowActionSkillRefreshSummary, m.catalog.Msg("正在读取项目/全局 Skill 信息…", "Loading project/global skill status..."))
			case flowScreenMCPTargets:
				return m.startBusy(flowActionMCPRefreshSummary, m.catalog.Msg("正在读取 MCP 配置…", "Reading MCP configuration..."))
			}
		}
		return m, nil
	case tea.KeyPgDown:
		prev := m.currentCursor()
		if m.focusDetails {
			m.detailScroll += 5
		} else {
			m.moveCursor(5)
		}
		if !m.focusDetails && m.currentCursor() != prev {
			switch m.screen {
			case flowScreenSkillTargets:
				return m.startBusy(flowActionSkillRefreshSummary, m.catalog.Msg("正在读取项目/全局 Skill 信息…", "Loading project/global skill status..."))
			case flowScreenMCPTargets:
				return m.startBusy(flowActionMCPRefreshSummary, m.catalog.Msg("正在读取 MCP 配置…", "Reading MCP configuration..."))
			}
		}
		return m, nil
	case tea.KeyHome:
		prev := m.currentCursor()
		if m.focusDetails {
			m.detailScroll = 0
		} else {
			m.setCursor(0)
		}
		if !m.focusDetails && m.currentCursor() != prev {
			switch m.screen {
			case flowScreenSkillTargets:
				return m.startBusy(flowActionSkillRefreshSummary, m.catalog.Msg("正在读取项目/全局 Skill 信息…", "Loading project/global skill status..."))
			case flowScreenMCPTargets:
				return m.startBusy(flowActionMCPRefreshSummary, m.catalog.Msg("正在读取 MCP 配置…", "Reading MCP configuration..."))
			}
		}
		return m, nil
	case tea.KeyEnd:
		prev := m.currentCursor()
		if m.focusDetails {
			m.detailScroll = 1 << 30
		} else {
			m.setCursor(m.currentOptionsLen() - 1)
		}
		if !m.focusDetails && m.currentCursor() != prev {
			switch m.screen {
			case flowScreenSkillTargets:
				return m.startBusy(flowActionSkillRefreshSummary, m.catalog.Msg("正在读取项目/全局 Skill 信息…", "Loading project/global skill status..."))
			case flowScreenMCPTargets:
				return m.startBusy(flowActionMCPRefreshSummary, m.catalog.Msg("正在读取 MCP 配置…", "Reading MCP configuration..."))
			}
		}
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
		m.focusDetails = false
		m.detailScroll = 0
		return m.handleBack()
	case tea.KeyEnter:
		m.focusDetails = false
		m.detailScroll = 0
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
		if m.projectInstallMode {
			m.projectInstallMode = false
			m.screen = flowScreenInstallHub
		} else {
			m.screen = flowScreenMain
		}
	case flowScreenSkillScope:
		m.screen = flowScreenSkillTargets
	case flowScreenSkillProjectActions:
		m.screen = flowScreenSkillScope
	case flowScreenSkillProjectProfile:
		if m.projectInstallMode {
			m.screen = flowScreenSkillTargets
		} else {
			m.screen = flowScreenSkillProjectActions
		}
	case flowScreenSkillActions:
		m.screen = flowScreenSkillScope
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
	case flowScreenUpdateConfirm:
		m.screen = flowScreenMain
	case flowScreenBootstrapConfig:
		// Skip config by pressing Esc.
		m.screen = flowScreenBootstrapActions
		m.configEditing = false
	}
	m.notice = nil
	return m, nil
}

// handleConfigKey processes key events during the config text-input screen.
func (m interactiveFlowModel) handleConfigKey(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.Type {
	case tea.KeyEsc:
		// Skip config.
		m.screen = flowScreenBootstrapActions
		m.configEditing = false
		return m, nil
	case tea.KeyEnter:
		// Save config.
		m.focusDetails = false
		m.detailScroll = 0
		return m.handleEnter()
	case tea.KeyUp:
		if m.configFieldCursor > 0 {
			m.configFieldCursor--
		}
		return m, nil
	case tea.KeyDown, tea.KeyTab:
		if m.configFieldCursor < len(m.configFields)-1 {
			m.configFieldCursor++
		}
		return m, nil
	case tea.KeyLeft:
		// For select fields: move to previous option.
		if m.configFieldCursor >= 0 && m.configFieldCursor < len(m.configFields) {
			f := &m.configFields[m.configFieldCursor]
			if f.FieldType == "select" && len(f.Options) > 0 {
				if f.OptionCursor > 0 {
					f.OptionCursor--
				} else {
					f.OptionCursor = len(f.Options) - 1
				}
				f.Value = f.Options[f.OptionCursor]
				f.Dirty = true
			}
		}
		return m, nil
	case tea.KeyRight:
		// For select fields: move to next option.
		if m.configFieldCursor >= 0 && m.configFieldCursor < len(m.configFields) {
			f := &m.configFields[m.configFieldCursor]
			if f.FieldType == "select" && len(f.Options) > 0 {
				if f.OptionCursor < len(f.Options)-1 {
					f.OptionCursor++
				} else {
					f.OptionCursor = 0
				}
				f.Value = f.Options[f.OptionCursor]
				f.Dirty = true
			}
		}
		return m, nil
	case tea.KeyBackspace:
		if m.configFieldCursor >= 0 && m.configFieldCursor < len(m.configFields) {
			f := &m.configFields[m.configFieldCursor]
			if f.FieldType != "select" {
				v := f.Value
				if len(v) > 0 {
					f.Value = v[:len(v)-1]
					f.Dirty = true
				}
			}
		}
		return m, nil
	case tea.KeyRunes:
		if m.configFieldCursor >= 0 && m.configFieldCursor < len(m.configFields) {
			f := &m.configFields[m.configFieldCursor]
			if f.FieldType != "select" {
				f.Value += key.String()
				f.Dirty = true
			}
		}
		return m, nil
	}
	return m, nil
}

func (m interactiveFlowModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.screen {
	case flowScreenMain:
		switch Action(m.mainOptions[m.mainCursor].Value) {
		case ActionInstall:
			m.screen = flowScreenInstallHub
			m.installHubCursor = 0
			m.notice = nil
			// Do NOT start any async operation here.
			// statusPanel() runs 13+ shell commands + network requests which blocks
			// the TUI for many seconds. Use the already-cached m.status instead.
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
			m.focusDetails = false
			m.detailScroll = 0
			return m.startBusy(flowActionMCPRefreshSummary, m.catalog.Msg("正在读取 MCP 配置…", "Reading MCP configuration..."))
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
			m.focusDetails = false
			m.detailScroll = 0
			return m.startBusy(flowActionSkillRefreshSummary, m.catalog.Msg("正在读取项目/全局 Skill 信息…", "Loading project/global skill status..."))
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
		m.focusDetails = false
		m.detailScroll = 0
		return m.startBusy(flowActionMCPList, m.catalog.Msg("正在读取 MCP 配置…", "Reading MCP configuration..."))
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
				m.mcpInstallOptions = m.annotateRecommendedMCPOptions(m.selectedMCPTarget, m.callbacks.MCPInstallOptions())
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
		if m.projectInstallMode {
			// Project install mode: skip scope selection, go directly to profile.
			m.screen = flowScreenSkillProjectProfile
			m.notice = nil
			m.focusDetails = false
			m.detailScroll = 0
			return m, nil
		}
		m.skillScopeOptions = m.skillScopeOptionsList()
		m.skillScopeCursor = 0
		if len(m.skillScopeOptions) > 0 {
			m.selectedSkillScope = m.skillScopeOptions[0].Value
		}
		m.screen = flowScreenSkillScope
		m.notice = nil
		m.focusDetails = false
		m.detailScroll = 0
		return m, nil
	case flowScreenSkillScope:
		if len(m.skillScopeOptions) == 0 {
			return m, nil
		}
		m.selectedSkillScope = m.skillScopeOptions[m.skillScopeCursor].Value
		switch m.selectedSkillScope {
		case "project":
			m.skillProjectActions = m.skillProjectActionsList()
			m.skillProjectActionCursor = 0
			m.screen = flowScreenSkillProjectActions
			m.notice = nil
			m.focusDetails = false
			m.detailScroll = 0
			return m, nil
		case "global":
			if m.callbacks.SkillGlobalSupported != nil && !m.callbacks.SkillGlobalSupported(m.selectedSkillTarget) {
				m.notice = panelRef(Panel{
					Title: m.catalog.Msg("Skills", "Skills"),
					Lines: []string{m.catalog.Msg("该目标不支持全局 Skill（仅支持项目级规则文件）。", "This target does not support global skills (project rules only).")},
				})
				return m, nil
			}
			m.screen = flowScreenSkillActions
			m.skillActionCursor = 0
			m.notice = nil
			m.focusDetails = false
			m.detailScroll = 0
			return m.startBusy(flowActionSkillList, m.catalog.Msg("正在读取已安装 skills…", "Reading installed skills..."))
		default:
			return m, nil
		}
	case flowScreenSkillProjectActions:
		if len(m.skillProjectActions) == 0 {
			return m, nil
		}
		switch m.skillProjectActions[m.skillProjectActionCursor].Value {
		case "refresh":
			return m.startBusy(flowActionSkillRefreshSummary, m.catalog.Msg("正在刷新项目/全局 Skill 信息…", "Refreshing project/global skill status..."))
		case "install-rules":
			if m.callbacks.ProjectRulesInstall == nil {
				m.notice = panelRef(Panel{
					Title: m.catalog.Msg("项目规则", "Project rules"),
					Lines: []string{m.catalog.Msg("当前构建未启用项目规则写入回调。", "Project-rules install callback is not enabled in this build.")},
				})
				return m, nil
			}
			m.screen = flowScreenSkillProjectProfile
			m.notice = nil
			m.focusDetails = false
			m.detailScroll = 0
			return m, nil
		default:
			return m, nil
		}
	case flowScreenSkillProjectProfile:
		m.selectedProjectProfile = m.profileOptions[m.profileCursor].Value
		return m.startBusy(flowActionProjectRulesInstall, m.catalog.Msg("正在写入项目规则文件…", "Writing project rule files..."))
	case flowScreenSkillActions:
		if len(m.skillActions) == 0 {
			return m, nil
		}
		switch m.skillActions[m.skillActionCursor].Value {
		case "list":
			m.selectedSkillValue = ""
			return m.startBusy(flowActionSkillList, m.catalog.Msg("正在读取已安装 skills…", "Reading installed skills..."))
		case "install":
			return m.startBusy(flowActionSkillLoadInstallOptions, m.catalog.Msg("正在加载推荐 skills…", "Loading recommended skills..."))
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
		case "install-project":
			if m.callbacks.SkillTargetOptions == nil {
				m.notice = panelRef(Panel{
					Title: m.catalog.Msg("项目级安装", "Project install"),
					Lines: []string{m.catalog.Msg("当前构建未启用项目级安装回调。", "Project install callbacks are not enabled in this build.")},
				})
				return m, nil
			}
			m.skillTargets = cloneOptions(m.callbacks.SkillTargetOptions())
			if len(m.skillTargets) == 0 {
				m.notice = panelRef(Panel{
					Title: m.catalog.Msg("项目级安装", "Project install"),
					Lines: []string{m.catalog.Msg("未检测到可管理的目标。", "No manageable targets detected.")},
				})
				return m, nil
			}
			m.projectInstallMode = true
			m.screen = flowScreenSkillTargets
			m.skillTargetCursor = 0
			m.selectedSkillTarget = m.skillTargets[0].Value
			m.notice = nil
			m.focusDetails = false
			m.detailScroll = 0
			return m.startBusy(flowActionSkillRefreshSummary, m.catalog.Msg("正在读取项目/全局 Skill 信息…", "Loading project/global skill status..."))
		case "uninstall-agentflow":
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
		case "uninstall-cli":
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
		case "configure":
			// User chose to configure custom settings.
			if len(m.pendingConfigFields) > 0 {
				m.configTarget = m.selectedBootstrapTarget
				m.configFields = make([]configFieldState, len(m.pendingConfigFields))
				for idx, f := range m.pendingConfigFields {
					fieldType := f.Type
					if fieldType == "" {
						fieldType = "text"
					}
					state := configFieldState{
						Label:     f.Label,
						EnvVar:    f.EnvVar,
						FieldType: fieldType,
						Options:   f.Options,
					}
					if fieldType == "select" && len(f.Options) > 0 {
						state.Value = f.Default
						for i, opt := range f.Options {
							if opt == f.Default {
								state.OptionCursor = i
								break
							}
						}
					}
					m.configFields[idx] = state
				}
				m.configFieldCursor = 0
				m.configEditing = true
				m.screen = flowScreenBootstrapConfig
			}
			return m, nil
		case "done":
			// Return to main menu.
			m.pendingConfigFields = nil
			m.screen = flowScreenMain
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
	case flowScreenUpdateConfirm:
		if len(m.updateConfirmOptions) == 0 {
			return m, nil
		}
		switch m.updateConfirmOptions[m.updateConfirmCursor].Value {
		case "restart":
			exePath, err := os.Executable()
			if err != nil {
				m.notice = panelRef(Panel{
					Title: m.catalog.Msg("重启失败", "Restart failed"),
					Lines: []string{fmt.Sprintf(m.catalog.Msg("无法获取当前可执行文件路径: %v", "Cannot determine executable path: %v"), err)},
				})
				m.screen = flowScreenMain
				return m, nil
			}
			// Use syscall.Exec to replace the current process with the updated binary.
			err = syscall.Exec(exePath, os.Args, os.Environ())
			if err != nil {
				m.notice = panelRef(Panel{
					Title: m.catalog.Msg("重启失败", "Restart failed"),
					Lines: []string{fmt.Sprintf(m.catalog.Msg("重启失败: %v。请手动运行 agentflow。", "Restart failed: %v. Please run agentflow manually."), err)},
				})
				m.screen = flowScreenMain
				return m, nil
			}
			// Should not reach here; syscall.Exec replaces the process.
			return m, tea.Quit
		case "cancel":
			m.screen = flowScreenMain
			return m, nil
		}
	case flowScreenBootstrapConfig:
		// Collect only user-modified values (Dirty flag tracks explicit edits).
		envVars := make(map[string]string)
		for _, f := range m.configFields {
			if f.Dirty && strings.TrimSpace(f.Value) != "" {
				envVars[f.EnvVar] = f.Value
			}
		}
		if len(envVars) == 0 {
			// All empty — skip config.
			m.screen = flowScreenBootstrapActions
			m.configEditing = false
			m.notice = panelRef(Panel{
				Title: m.catalog.Msg("配置跳过", "Configuration skipped"),
				Lines: []string{m.catalog.Msg("未输入任何配置，将使用官方默认值。", "No configuration entered; using official defaults.")},
			})
			return m, nil
		}
		m.configEditing = false
		return m.startBusy(flowActionWriteEnvConfig, m.catalog.Msg("正在写入配置…", "Writing configuration..."))
	}
	return m, nil
}

func (m interactiveFlowModel) startBusy(action flowAction, message string) (tea.Model, tea.Cmd) {
	m.busy = true
	m.spin = 0
	m.activeAction = action
	if action == flowActionUpdate {
		m.updateProgress = &updateProgressState{percent: -1}
	} else {
		m.updateProgress = nil
	}
	return m, tea.Batch(m.runActionCmd(action), busyTickCmd())
}

// handleBusyNavKey handles navigation keys while an async operation is running.
// This ensures menu navigation is never blocked by background tasks.
// Only cursor movement and focus switching are allowed; no new async actions are started.
func (m interactiveFlowModel) handleBusyNavKey(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.Type {
	case tea.KeyUp:
		if m.focusDetails {
			m.detailScroll--
		} else {
			m.moveCursor(-1)
		}
	case tea.KeyDown:
		if m.focusDetails {
			m.detailScroll++
		} else {
			m.moveCursor(1)
		}
	case tea.KeyLeft:
		m.focusDetails = false
	case tea.KeyRight:
		m.focusDetails = true
	case tea.KeyTab:
		m.focusDetails = !m.focusDetails
	case tea.KeyPgUp:
		if m.focusDetails {
			m.detailScroll -= 5
		} else {
			m.moveCursor(-5)
		}
	case tea.KeyPgDown:
		if m.focusDetails {
			m.detailScroll += 5
		} else {
			m.moveCursor(5)
		}
	case tea.KeyHome:
		if m.focusDetails {
			m.detailScroll = 0
		} else {
			m.setCursor(0)
		}
	case tea.KeyEnd:
		if m.focusDetails {
			m.detailScroll = 1 << 30
		} else {
			m.setCursor(m.currentOptionsLen() - 1)
		}
	case tea.KeyEsc:
		// During busy, Esc only switches focus back to left panel.
		m.focusDetails = false
		m.detailScroll = 0
	}
	return m, nil
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
	previousCursor := m.currentCursor()
	switch m.screen {
	case flowScreenMain:
		if len(m.mainOptions) == 0 {
			m.mainCursor = 0
			break
		}
		if cursor > len(m.mainOptions)-1 {
			cursor = len(m.mainOptions) - 1
		}
		m.mainCursor = cursor
	case flowScreenInstallHub:
		if len(m.installHubOptions) == 0 {
			m.installHubCursor = 0
			break
		}
		if cursor > len(m.installHubOptions)-1 {
			cursor = len(m.installHubOptions) - 1
		}
		m.installHubCursor = cursor
	case flowScreenMCPTargets:
		if len(m.mcpTargets) == 0 {
			m.mcpTargetCursor = 0
			break
		}
		if cursor > len(m.mcpTargets)-1 {
			cursor = len(m.mcpTargets) - 1
		}
		m.mcpTargetCursor = cursor
		m.selectedMCPTarget = m.mcpTargets[cursor].Value
	case flowScreenMCPActions:
		if len(m.mcpActions) == 0 {
			m.mcpActionCursor = 0
			break
		}
		if cursor > len(m.mcpActions)-1 {
			cursor = len(m.mcpActions) - 1
		}
		m.mcpActionCursor = cursor
	case flowScreenMCPInstall:
		if len(m.mcpInstallOptions) == 0 {
			m.mcpInstallCursor = 0
			break
		}
		if cursor > len(m.mcpInstallOptions)-1 {
			cursor = len(m.mcpInstallOptions) - 1
		}
		m.mcpInstallCursor = cursor
	case flowScreenMCPRemove:
		if len(m.mcpRemoveOptions) == 0 {
			m.mcpRemoveCursor = 0
			break
		}
		if cursor > len(m.mcpRemoveOptions)-1 {
			cursor = len(m.mcpRemoveOptions) - 1
		}
		m.mcpRemoveCursor = cursor
	case flowScreenSkillTargets:
		if len(m.skillTargets) == 0 {
			m.skillTargetCursor = 0
			break
		}
		if cursor > len(m.skillTargets)-1 {
			cursor = len(m.skillTargets) - 1
		}
		m.skillTargetCursor = cursor
		m.selectedSkillTarget = m.skillTargets[cursor].Value
	case flowScreenSkillScope:
		if len(m.skillScopeOptions) == 0 {
			m.skillScopeCursor = 0
			break
		}
		if cursor > len(m.skillScopeOptions)-1 {
			cursor = len(m.skillScopeOptions) - 1
		}
		m.skillScopeCursor = cursor
		m.selectedSkillScope = m.skillScopeOptions[cursor].Value
	case flowScreenSkillProjectActions:
		if len(m.skillProjectActions) == 0 {
			m.skillProjectActionCursor = 0
			break
		}
		if cursor > len(m.skillProjectActions)-1 {
			cursor = len(m.skillProjectActions) - 1
		}
		m.skillProjectActionCursor = cursor
	case flowScreenSkillProjectProfile:
		if len(m.profileOptions) == 0 {
			m.profileCursor = 0
			break
		}
		if cursor > len(m.profileOptions)-1 {
			cursor = len(m.profileOptions) - 1
		}
		m.profileCursor = cursor
	case flowScreenSkillActions:
		if len(m.skillActions) == 0 {
			m.skillActionCursor = 0
			break
		}
		if cursor > len(m.skillActions)-1 {
			cursor = len(m.skillActions) - 1
		}
		m.skillActionCursor = cursor
	case flowScreenSkillInstall:
		if len(m.skillInstallOptions) == 0 {
			m.skillInstallCursor = 0
			break
		}
		if cursor > len(m.skillInstallOptions)-1 {
			cursor = len(m.skillInstallOptions) - 1
		}
		m.skillInstallCursor = cursor
	case flowScreenSkillUninstall:
		if len(m.skillUninstallOptions) == 0 {
			m.skillUninstallCursor = 0
			break
		}
		if cursor > len(m.skillUninstallOptions)-1 {
			cursor = len(m.skillUninstallOptions) - 1
		}
		m.skillUninstallCursor = cursor
	case flowScreenBootstrapTargets:
		if len(m.bootstrapOptions) == 0 {
			m.bootstrapCursor = 0
			break
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
			break
		}
		if cursor > len(m.bootstrapActionOptions)-1 {
			cursor = len(m.bootstrapActionOptions) - 1
		}
		m.bootstrapActionCursor = cursor
	case flowScreenProfile:
		if len(m.profileOptions) == 0 {
			m.profileCursor = 0
			break
		}
		if cursor > len(m.profileOptions)-1 {
			cursor = len(m.profileOptions) - 1
		}
		m.profileCursor = cursor
	case flowScreenInstallTargets:
		if len(m.installOptions) == 0 {
			m.installCursor = 0
			break
		}
		if cursor > len(m.installOptions)-1 {
			cursor = len(m.installOptions) - 1
		}
		m.installCursor = cursor
	case flowScreenUninstallTargets:
		if len(m.uninstallOptions) == 0 {
			m.uninstallCursor = 0
			break
		}
		if cursor > len(m.uninstallOptions)-1 {
			cursor = len(m.uninstallOptions) - 1
		}
		m.uninstallCursor = cursor
	case flowScreenUpdateConfirm:
		if len(m.updateConfirmOptions) == 0 {
			m.updateConfirmCursor = 0
			break
		}
		if cursor > len(m.updateConfirmOptions)-1 {
			cursor = len(m.updateConfirmOptions) - 1
		}
		m.updateConfirmCursor = cursor
	case flowScreenBootstrapConfig:
		if len(m.configFields) == 0 {
			m.configFieldCursor = 0
			break
		}
		if cursor > len(m.configFields)-1 {
			cursor = len(m.configFields) - 1
		}
		m.configFieldCursor = cursor
	}

	if m.currentCursor() != previousCursor {
		m.detailScroll = 0
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
	case flowScreenSkillScope:
		return m.skillScopeCursor
	case flowScreenSkillProjectActions:
		return m.skillProjectActionCursor
	case flowScreenSkillProjectProfile:
		return m.profileCursor
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
	case flowScreenUpdateConfirm:
		return m.updateConfirmCursor
	case flowScreenBootstrapConfig:
		return m.configFieldCursor
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
	case flowScreenSkillScope:
		return len(m.skillScopeOptions)
	case flowScreenSkillProjectActions:
		return len(m.skillProjectActions)
	case flowScreenSkillProjectProfile:
		return len(m.profileOptions)
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
	case flowScreenUpdateConfirm:
		return len(m.updateConfirmOptions)
	case flowScreenBootstrapConfig:
		return len(m.configFields)
	default:
		return len(m.mainOptions)
	}
}

func (m interactiveFlowModel) refreshStatusCmd(withNotice bool) tea.Cmd {
	return func() tea.Msg {
		done := debuglog.Timed("refreshStatusCmd")
		defer done()
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
	projectRoot := m.projectRoot
	selectedMCPTarget := m.selectedMCPTarget
	selectedMCPServer := m.selectedMCPServer
	selectedSkillTarget := m.selectedSkillTarget
	selectedSkillValue := m.selectedSkillValue
	selectedProjectProfile := m.selectedProjectProfile
	// Capture config fields for writeEnvConfig.
	configEnvVars := make(map[string]string)
	for _, f := range m.configFields {
		if strings.TrimSpace(f.Value) != "" {
			configEnvVars[f.EnvVar] = f.Value
		}
	}

	return func() tea.Msg {
		done := debuglog.Timed(fmt.Sprintf("runActionCmd(%d)", action))
		defer done()
		switch action {
		case flowActionRefreshStatus:
			return flowResultMsg{
				action: action,
				status: m.callbacks.Status(),
			}
		case flowActionInstallHubRefresh:
			// Only call Status() — it's lightweight.
			// Do NOT call InstallOptions() here: it triggers DetectTargetStatuses()
			// which runs shell commands for every CLI target and causes multi-second
			// blocking that makes the TUI appear frozen.
			return flowResultMsg{
				action: action,
				status: m.callbacks.Status(),
			}
		case flowActionMCPRefreshSummary, flowActionMCPList:
			summary := Panel{}
			if m.callbacks.MCPList != nil {
				summary = m.callbacks.MCPList(selectedMCPTarget)
			}
			return flowResultMsg{
				action:     action,
				status:     m.callbacks.Status(),
				mcpSummary: panelRef(summary),
			}
		case flowActionMCPInstall:
			notice := Panel{}
			if m.callbacks.MCPInstall != nil {
				notice = m.callbacks.MCPInstall(selectedMCPTarget, selectedMCPServer)
			}
			summary := Panel{}
			if m.callbacks.MCPList != nil {
				summary = m.callbacks.MCPList(selectedMCPTarget)
			}
			return flowResultMsg{
				action:     action,
				notice:     panelRef(notice),
				status:     m.callbacks.Status(),
				mcpSummary: panelRef(summary),
			}
		case flowActionMCPRemove:
			notice := Panel{}
			if m.callbacks.MCPRemove != nil {
				notice = m.callbacks.MCPRemove(selectedMCPTarget, selectedMCPServer)
			}
			summary := Panel{}
			if m.callbacks.MCPList != nil {
				summary = m.callbacks.MCPList(selectedMCPTarget)
			}
			return flowResultMsg{
				action:     action,
				notice:     panelRef(notice),
				status:     m.callbacks.Status(),
				mcpSummary: panelRef(summary),
			}
		case flowActionSkillRefreshSummary:
			projectRules := Panel{}
			if m.callbacks.ProjectRulesPanel != nil {
				projectRules = m.callbacks.ProjectRulesPanel(projectRoot, selectedSkillTarget)
			}
			summary := Panel{}
			if m.callbacks.SkillList != nil {
				summary = m.callbacks.SkillList(selectedSkillTarget)
			}
			return flowResultMsg{
				action:       action,
				status:       m.callbacks.Status(),
				projectRules: panelRef(projectRules),
				skillSummary: panelRef(summary),
			}
		case flowActionProjectRulesInstall:
			notice := Panel{}
			if m.callbacks.ProjectRulesInstall != nil {
				notice = m.callbacks.ProjectRulesInstall(projectRoot, selectedSkillTarget, selectedProjectProfile)
			}
			projectRules := Panel{}
			if m.callbacks.ProjectRulesPanel != nil {
				projectRules = m.callbacks.ProjectRulesPanel(projectRoot, selectedSkillTarget)
			}
			summary := Panel{}
			if m.callbacks.SkillList != nil {
				summary = m.callbacks.SkillList(selectedSkillTarget)
			}
			return flowResultMsg{
				action:       action,
				notice:       panelRef(notice),
				status:       m.callbacks.Status(),
				projectRules: panelRef(projectRules),
				skillSummary: panelRef(summary),
			}
		case flowActionSkillList:
			projectRules := Panel{}
			if m.callbacks.ProjectRulesPanel != nil {
				projectRules = m.callbacks.ProjectRulesPanel(projectRoot, selectedSkillTarget)
			}
			summary := Panel{}
			if m.callbacks.SkillList != nil {
				summary = m.callbacks.SkillList(selectedSkillTarget)
			}
			return flowResultMsg{
				action:       action,
				status:       m.callbacks.Status(),
				projectRules: panelRef(projectRules),
				skillSummary: panelRef(summary),
			}
		case flowActionSkillLoadInstallOptions:
			options := []Option{}
			if m.callbacks.SkillInstallOptions != nil {
				options = m.callbacks.SkillInstallOptions(selectedSkillTarget)
				options = m.annotateRecommendedSkillOptions(selectedSkillTarget, options)
			}
			return flowResultMsg{
				action:       action,
				status:       m.callbacks.Status(),
				skillOptions: options,
			}
		case flowActionSkillInstall:
			notice := Panel{}
			if m.callbacks.SkillInstall != nil {
				notice = m.callbacks.SkillInstall(selectedSkillTarget, selectedSkillValue)
			}
			summary := Panel{}
			if m.callbacks.SkillList != nil {
				summary = m.callbacks.SkillList(selectedSkillTarget)
			}
			return flowResultMsg{
				action:       action,
				notice:       panelRef(notice),
				status:       m.callbacks.Status(),
				skillSummary: panelRef(summary),
			}
		case flowActionSkillUninstall:
			notice := Panel{}
			if m.callbacks.SkillUninstall != nil {
				notice = m.callbacks.SkillUninstall(selectedSkillTarget, selectedSkillValue)
			}
			summary := Panel{}
			if m.callbacks.SkillList != nil {
				summary = m.callbacks.SkillList(selectedSkillTarget)
			}
			return flowResultMsg{
				action:       action,
				notice:       panelRef(notice),
				status:       m.callbacks.Status(),
				skillSummary: panelRef(summary),
			}
		case flowActionUpdate:
			progress := m.updateProgress
			notice, version := m.callbacks.Update(func(stage string, percent int, info string) {
				if progress != nil {
					progress.mu.Lock()
					if info != "" {
						progress.stage = stage + ":" + info
					} else {
						progress.stage = stage
					}
					progress.percent = percent
					progress.mu.Unlock()
				}
			})
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
		case flowActionWriteEnvConfig:
			notice := Panel{}
			if m.callbacks.WriteEnvConfig != nil {
				notice = m.callbacks.WriteEnvConfig(configEnvVars)
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
	model.focusDetails = m.focusDetails
	model.detailScroll = m.detailScroll

	switch m.screen {
	case flowScreenInstallHub:
		model.subtitle = m.catalog.Msg("先决定要安装 CLI 工具，还是把 AgentFlow 写入已经存在的 CLI。Esc 返回主菜单。", "Choose whether to install CLI tools first, or write AgentFlow into CLIs that already exist. Press Esc to return.")
		model.hint = m.catalog.Msg("↑/↓ 切换安装路径，Enter 继续，Esc 返回。", "Use ↑/↓ to choose the install path, Enter to continue, Esc to go back.")
		model.options = cloneOptions(m.installHubOptions)
		model.cursor = m.installHubCursor
		model.panels = m.installHubPanels()
	case flowScreenMCPTargets:
		model.subtitle = m.catalog.Msg("选择要管理 MCP 的目标（CLI/IDE）。Esc 返回主菜单。", "Choose which target (CLI/IDE) to manage MCP for. Press Esc to return.")
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
		model.subtitle = fmt.Sprintf(m.catalog.Msg("当前目录: %s。选择要管理 Skill 的目标（CLI/IDE）。Esc 返回主菜单。", "Current directory: %s. Choose which target (CLI/IDE) to manage skills for. Press Esc to return."), m.projectRoot)
		model.hint = m.catalog.Msg("↑/↓ 切换目标，Enter 继续，Esc 返回。", "Use ↑/↓ to switch targets, Enter to continue, Esc to go back.")
		model.options = cloneOptions(m.skillTargets)
		model.cursor = m.skillTargetCursor
		model.panels = m.skillTargetPanels()
	case flowScreenSkillScope:
		model.subtitle = fmt.Sprintf(m.catalog.Msg("目标: %s。选择安装范围（项目/全局）。Esc 返回目标列表。", "Target: %s. Choose install scope (project/global). Press Esc to go back."), m.selectedSkillTarget)
		model.hint = m.catalog.Msg("↑/↓ 选择范围，Enter 继续，Esc 返回。", "Use ↑/↓ to choose scope, Enter to continue, Esc to go back.")
		model.options = cloneOptions(m.skillScopeOptions)
		model.cursor = m.skillScopeCursor
		model.panels = m.skillScopePanels()
	case flowScreenSkillProjectActions:
		model.subtitle = fmt.Sprintf(m.catalog.Msg("目标: %s（项目级规则）。Esc 返回范围选择。", "Target: %s (project rules). Press Esc to go back."), m.selectedSkillTarget)
		model.hint = m.catalog.Msg("↑/↓ 选择操作，Enter 执行，Esc 返回。", "Use ↑/↓ to choose an action, Enter to run, Esc to go back.")
		model.options = cloneOptions(m.skillProjectActions)
		model.cursor = m.skillProjectActionCursor
		model.panels = m.skillProjectPanels()
	case flowScreenSkillProjectProfile:
		model.subtitle = m.catalog.Msg("选择 Profile（用于写入 CLI 项目规则）。Esc 返回操作列表。", "Select a profile for writing CLI project rules. Press Esc to go back.")
		model.hint = m.catalog.Msg("↑/↓ 切换 Profile，Enter 确认，Esc 返回。", "Use ↑/↓ to switch profile, Enter to confirm, Esc to go back.")
		model.options = cloneOptions(m.profileOptions)
		model.cursor = m.profileCursor
		model.panels = m.skillProjectProfilePanels()
	case flowScreenSkillActions:
		model.subtitle = fmt.Sprintf(m.catalog.Msg("目标: %s（全局 Skills）。Esc 返回范围选择。", "Target: %s (global skills). Press Esc to go back."), m.selectedSkillTarget)
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
	case flowScreenUpdateConfirm:
		model.subtitle = fmt.Sprintf(m.catalog.Msg("AgentFlow 已更新到 v%s。是否立即重启？", "AgentFlow has been updated to v%s. Restart now?"), m.version)
		model.hint = m.catalog.Msg("↑/↓ 选择，Enter 确认，Esc 返回主菜单。", "Use ↑/↓ to choose, Enter to confirm, Esc to go back.")
		model.options = cloneOptions(m.updateConfirmOptions)
		model.cursor = m.updateConfirmCursor
		panels := make([]Panel, 0, 2)
		if m.notice != nil {
			panels = append(panels, *m.notice)
		}
		panels = append(panels, m.status)
		model.panels = panels
	case flowScreenBootstrapConfig:
		model.subtitle = fmt.Sprintf(m.catalog.Msg("配置 %s（可选，留空跳过）", "Configure %s (optional, leave empty to skip)"), m.configTarget)
		model.hint = m.catalog.Msg("↑/↓ 切换字段，文本框直接输入，选择框 ←/→ 切换，Enter 保存，Esc 跳过。", "↑/↓ fields, type for text, ←/→ for select, Enter save, Esc skip.")
		// Build virtual options from config fields to display as a form.
		options := make([]Option, len(m.configFields))
		for idx, f := range m.configFields {
			var displayValue string
			if f.FieldType == "select" && len(f.Options) > 0 {
				// Show selector with ◀ current ▶ indicator.
				current := f.Options[f.OptionCursor]
				if idx == m.configFieldCursor && m.configEditing {
					displayValue = fmt.Sprintf("◀ %s ▶  (%d/%d)", current, f.OptionCursor+1, len(f.Options))
				} else {
					displayValue = current
				}
			} else {
				displayValue = f.Value
				if displayValue == "" {
					displayValue = m.catalog.Msg("(留空则使用官方默认)", "(leave empty to use official default)")
				}
				// Show cursor indicator for editing field.
				if idx == m.configFieldCursor && m.configEditing {
					displayValue = f.Value + "█"
				}
			}
			labelText := f.Label
			if f.EnvVar != "" && !strings.HasPrefix(f.EnvVar, "__") {
				labelText = fmt.Sprintf("%s (%s)", f.Label, f.EnvVar)
			}
			options[idx] = Option{
				Value:       f.EnvVar,
				Label:       labelText,
				Badge:       f.Label,
				Description: displayValue,
			}
		}
		model.options = options
		model.cursor = m.configFieldCursor
		panels := make([]Panel, 0, 2)
		if m.notice != nil {
			panels = append(panels, *m.notice)
		}
		panels = append(panels, m.status)
		model.panels = panels
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
	if m.installHubStatusPanel != nil {
		panels = append(panels, *m.installHubStatusPanel)
	}
	panels = append(panels, m.status)
	return panels
}

func (m interactiveFlowModel) skillScopeOptionsList() []Option {
	globalSupported := true
	if m.callbacks.SkillGlobalSupported != nil {
		globalSupported = m.callbacks.SkillGlobalSupported(m.selectedSkillTarget)
	}

	globalBadge := "GLOBAL"
	globalDesc := m.catalog.Msg("把 Skill 安装到目标 CLI 的全局 skills 目录（IDE 通常不支持）。", "Install skills into the target CLI's global skills directory (IDEs usually do not support this).")
	if !globalSupported {
		globalBadge = "N/A"
		globalDesc = m.catalog.Msg("该目标不支持全局 Skill（仍可写入项目级规则文件）。", "This target does not support global skills (project rules are still supported).")
	}

	return []Option{
		{
			Value:       "project",
			Label:       m.catalog.Msg("项目安装（规则文件）", "Project install (rule files)"),
			Badge:       "PROJECT",
			Description: m.catalog.Msg("把 AgentFlow 规则写入当前项目目录的规则文件（项目级 Skill/规则）。", "Write AgentFlow rules into project rule files for the current directory (project-level skills/rules)."),
		},
		{
			Value:       "global",
			Label:       m.catalog.Msg("全局安装（Skills）", "Global install (skills)"),
			Badge:       globalBadge,
			Description: globalDesc,
		},
	}
}

func (m interactiveFlowModel) skillProjectActionsList() []Option {
	return []Option{
		{
			Value:       "refresh",
			Label:       m.catalog.Msg("刷新概览", "Refresh summary"),
			Badge:       "↻",
			Description: m.catalog.Msg("重新读取项目规则文件与全局 Skills 状态。", "Reload project rule file status and global skills summary."),
		},
		{
			Value:       "install-rules",
			Label:       m.catalog.Msg("写入项目规则文件", "Write project rule files"),
			Badge:       m.catalog.Msg("写入", "WRITE"),
			Description: m.catalog.Msg("把 AgentFlow 规则写入当前项目目录（存在用户文件时自动备份）。", "Write AgentFlow rules into this directory (backs up existing user files)."),
		},
	}
}

func (m interactiveFlowModel) skillScopePanels() []Panel {
	panels := make([]Panel, 0, 5)
	if m.notice != nil {
		panels = append(panels, *m.notice)
	}
	panels = append(panels, Panel{
		Title: m.catalog.Msg("目录信息", "Directory"),
		Lines: []string{
			fmt.Sprintf(m.catalog.Msg("当前目录: %s", "Current directory: %s"), m.projectRoot),
			m.catalog.Msg("项目级 Skill/规则通过规则文件表达（如 AGENTS.md、.windsurfrules、.cursor/rules/*.mdc）。", "Project-level skills/rules are represented as rule files (e.g. AGENTS.md, .windsurfrules, .cursor/rules/*.mdc)."),
			m.catalog.Msg("MCP 仅支持全局；请在 MCP 菜单里管理。", "MCP is global-only; manage it under the MCP menu."),
		},
	})
	if m.projectRulesDetail != nil {
		panels = append(panels, *m.projectRulesDetail)
	}
	if m.skillSummary != nil {
		panels = append(panels, *m.skillSummary)
	}
	panels = append(panels, m.status)
	return panels
}

func (m interactiveFlowModel) skillProjectPanels() []Panel {
	panels := make([]Panel, 0, 5)
	if m.notice != nil {
		panels = append(panels, *m.notice)
	}
	if m.projectRulesDetail != nil {
		panels = append(panels, *m.projectRulesDetail)
	}
	if m.skillSummary != nil {
		panels = append(panels, *m.skillSummary)
	}
	panels = append(panels, Panel{
		Title: m.catalog.Msg("操作提示", "Action notes"),
		Lines: []string{
			m.catalog.Msg("写入项目规则会把 AgentFlow 规则写入当前项目目录。若已存在用户文件会先备份再覆盖。", "Writing project rules installs AgentFlow rules into this directory. Existing user files are backed up before overwrite."),
			m.catalog.Msg("全局 Skills 仅对支持 skills 目录的 CLI 生效。", "Global skills apply only to CLIs that support a skills directory."),
		},
	})
	panels = append(panels, m.status)
	return panels
}

func (m interactiveFlowModel) skillProjectProfilePanels() []Panel {
	panels := make([]Panel, 0, 4)
	if m.notice != nil {
		panels = append(panels, *m.notice)
	}
	if m.projectRulesDetail != nil {
		panels = append(panels, *m.projectRulesDetail)
	}
	panels = append(panels, Panel{
		Title: m.catalog.Msg("Profile 说明", "Profile guide"),
		Lines: []string{
			m.catalog.Msg("Profile 只影响写入到 CLI 项目规则文件里的模块深度（lite/standard/full）。", "Profile controls how much AgentFlow logic is embedded into CLI project rules (lite/standard/full)."),
			m.catalog.Msg("IDE 规则文件不受 Profile 影响（仍写入同一套核心规则）。", "IDE rule files are not affected by profile (core rules only)."),
		},
	})
	panels = append(panels, m.status)
	return panels
}

func (m interactiveFlowModel) mcpTargetPanels() []Panel {
	panels := make([]Panel, 0, 4)
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
	if m.mcpSummary != nil {
		panels = append(panels, *m.mcpSummary)
	}
	panels = append(panels, m.status)
	return panels
}

func (m interactiveFlowModel) mcpActionPanels() []Panel {
	panels := make([]Panel, 0, 4)
	if m.notice != nil {
		panels = append(panels, *m.notice)
	}
	if m.mcpSummary != nil {
		panels = append(panels, *m.mcpSummary)
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
	panels := make([]Panel, 0, 4)
	if m.notice != nil {
		panels = append(panels, *m.notice)
	}
	if m.mcpSummary != nil {
		panels = append(panels, *m.mcpSummary)
	}
	panels = append(panels, Panel{
		Title: m.catalog.Msg("安装说明", "Install guide"),
		Lines: []string{
			m.catalog.Msg("选择一个推荐 MCP server 并按 Enter 写入配置。", "Select a recommended MCP server and press Enter to write the config."),
			m.catalog.Msg("图例：★ 推荐 · ✓ 已安装（Enter 更新配置） · + 未安装（Enter 安装）。", "Legend: ★ recommended · ✓ installed (Enter updates config) · + not installed (Enter installs)."),
			m.catalog.Msg("Filesystem 默认会把当前工作目录加入 allowlist。", "Filesystem will add the current working directory to its allowlist by default."),
		},
	})
	return panels
}

func (m interactiveFlowModel) mcpRemovePanels() []Panel {
	panels := make([]Panel, 0, 4)
	if m.notice != nil {
		panels = append(panels, *m.notice)
	}
	if m.mcpSummary != nil {
		panels = append(panels, *m.mcpSummary)
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
	panels := make([]Panel, 0, 6)
	if m.notice != nil {
		panels = append(panels, *m.notice)
	}
	panels = append(panels, Panel{
		Title: m.catalog.Msg("目录信息", "Directory"),
		Lines: []string{
			fmt.Sprintf(m.catalog.Msg("当前目录: %s", "Current directory: %s"), m.projectRoot),
			m.catalog.Msg("项目级 Skill/规则通过规则文件表达（如 AGENTS.md、.windsurfrules、.cursor/rules/*.mdc）。", "Project-level skills/rules are represented as rule files (e.g. AGENTS.md, .windsurfrules, .cursor/rules/*.mdc)."),
		},
	})
	panels = append(panels, Panel{
		Title: m.catalog.Msg("Skill 提示", "Skill note"),
		Lines: []string{
			m.catalog.Msg("全局 Skills 适用于支持 skills 目录的 CLI；IDE 通常只支持项目级规则文件。", "Global skills apply to CLIs with a skills directory; IDEs typically support project rule files only."),
			m.catalog.Msg("选择目标后可进一步选择“项目安装 / 全局安装”。", "After selecting a target, choose “Project install / Global install”."),
		},
	})
	if m.projectRulesDetail != nil {
		panels = append(panels, *m.projectRulesDetail)
	}
	if m.skillSummary != nil {
		panels = append(panels, *m.skillSummary)
	}
	panels = append(panels, m.status)
	return panels
}

func (m interactiveFlowModel) skillActionPanels() []Panel {
	panels := make([]Panel, 0, 6)
	if m.notice != nil {
		panels = append(panels, *m.notice)
	}
	if m.projectRulesDetail != nil {
		panels = append(panels, *m.projectRulesDetail)
	}
	if m.skillSummary != nil {
		panels = append(panels, *m.skillSummary)
	}
	panels = append(panels, Panel{
		Title: m.catalog.Msg("Skill 管理", "Skill management"),
		Lines: []string{
			fmt.Sprintf(m.catalog.Msg("目标: %s", "Target: %s"), m.selectedSkillTarget),
			m.catalog.Msg("提示：要安装任意 skill，可用 `agentflow skill install <target> <skills.sh URL>`。", "Tip: To install any skill, use `agentflow skill install <target> <skills.sh URL>`."),
			m.catalog.Msg("推荐 Skill 的简介会从其 SKILL.md 的 description 读取。", "Recommended skill descriptions are read from SKILL.md front matter `description`."),
		},
	})
	panels = append(panels, m.status)
	return panels
}

func (m interactiveFlowModel) skillInstallPanels() []Panel {
	panels := make([]Panel, 0, 6)
	if m.notice != nil {
		panels = append(panels, *m.notice)
	}
	if m.projectRulesDetail != nil {
		panels = append(panels, *m.projectRulesDetail)
	}
	if m.skillSummary != nil {
		panels = append(panels, *m.skillSummary)
	}
	panels = append(panels, Panel{
		Title: m.catalog.Msg("安装说明", "Install guide"),
		Lines: []string{
			m.catalog.Msg("选择一个推荐 skill 并按 Enter 安装。", "Select a recommended skill and press Enter to install."),
			m.catalog.Msg("图例：★ 推荐 · ✓ 已安装（Enter 更新/重装） · + 未安装（Enter 安装）。", "Legend: ★ recommended · ✓ installed (Enter reinstalls/updates) · + not installed (Enter installs)."),
			m.catalog.Msg("如需安装其他 skill，请在终端使用 `agentflow skill install`。", "To install other skills, use `agentflow skill install` in the terminal."),
		},
	})
	panels = append(panels, m.status)
	return panels
}

func (m interactiveFlowModel) skillUninstallPanels() []Panel {
	panels := make([]Panel, 0, 6)
	if m.notice != nil {
		panels = append(panels, *m.notice)
	}
	if m.projectRulesDetail != nil {
		panels = append(panels, *m.projectRulesDetail)
	}
	if m.skillSummary != nil {
		panels = append(panels, *m.skillSummary)
	}
	panels = append(panels, Panel{
		Title: m.catalog.Msg("卸载说明", "Uninstall guide"),
		Lines: []string{
			m.catalog.Msg("选择要卸载的 skill 并按 Enter。", "Select the skill to uninstall and press Enter."),
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
	// For sub-screen operations, the screen itself indicates the action.
	switch {
	case m.screen == flowScreenMCPTargets || m.screen == flowScreenMCPActions || m.screen == flowScreenMCPInstall || m.screen == flowScreenMCPRemove:
		return m.catalog.Msg("正在更新 MCP 配置…", "Updating MCP configuration...")
	case m.screen == flowScreenSkillTargets || m.screen == flowScreenSkillScope || m.screen == flowScreenSkillProjectActions || m.screen == flowScreenSkillProjectProfile || m.screen == flowScreenSkillActions || m.screen == flowScreenSkillInstall || m.screen == flowScreenSkillUninstall:
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
	}

	// For main-menu-level operations, use the tracked active action
	// (NOT the cursor position) so moving the cursor doesn't change
	// the displayed message mid-operation.
	switch m.activeAction {
	case flowActionUpdate:
		if m.updateProgress != nil {
			return m.updateProgressMessage()
		}
		return m.catalog.Msg("正在检查最新版本并更新…", "Checking the latest release and updating...")
	case flowActionClean:
		return m.catalog.Msg("正在清理缓存…", "Cleaning caches...")
	case flowActionRefreshStatus:
		return m.catalog.Msg("正在刷新状态…", "Refreshing status...")
	default:
		return m.catalog.Msg("正在执行，请稍候…", "Working, please wait...")
	}
}

// updateProgressMessage renders a human-readable message from the current
// update progress state by reading the shared thread-safe state.
func (m interactiveFlowModel) updateProgressMessage() string {
	if m.updateProgress == nil {
		return m.catalog.Msg("正在检查最新版本并更新…", "Checking the latest release and updating...")
	}

	m.updateProgress.mu.Lock()
	stageRaw := m.updateProgress.stage
	percent := m.updateProgress.percent
	m.updateProgress.mu.Unlock()

	if stageRaw == "" {
		return m.catalog.Msg("正在检查最新版本并更新…", "Checking the latest release and updating...")
	}

	parts := strings.SplitN(stageRaw, ":", 2)
	stage := parts[0]
	info := ""
	if len(parts) > 1 {
		info = parts[1]
	}

	switch stage {
	case "checking":
		return m.catalog.Msg("正在检查最新版本…", "Checking for the latest version...")
	case "found":
		if info != "" {
			return fmt.Sprintf(m.catalog.Msg("发现新版本 v%s，准备下载…", "Found new version v%s, preparing download..."), info)
		}
		return m.catalog.Msg("发现新版本，准备下载…", "Found new version, preparing download...")
	case "downloading":
		if percent >= 0 && percent <= 100 {
			if info != "" {
				return fmt.Sprintf(m.catalog.Msg("正在下载 v%s… (%d%%)", "Downloading v%s... (%d%%)"), info, percent)
			}
			return fmt.Sprintf(m.catalog.Msg("正在下载… (%d%%)", "Downloading... (%d%%)"), percent)
		}
		if info != "" {
			return fmt.Sprintf(m.catalog.Msg("正在下载 v%s…", "Downloading v%s..."), info)
		}
		return m.catalog.Msg("正在下载…", "Downloading...")
	case "replacing":
		return m.catalog.Msg("正在替换二进制文件…", "Replacing binary...")
	default:
		return m.catalog.Msg("正在检查最新版本并更新…", "Checking the latest release and updating...")
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

var spinnerFrames = []string{"◐", "◓", "◑", "◒"}
