package ui

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

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
	case flowScreenToolbox:
		if len(m.toolboxOptions) == 0 {
			m.toolboxCursor = 0
			break
		}
		if cursor > len(m.toolboxOptions)-1 {
			cursor = len(m.toolboxOptions) - 1
		}
		m.toolboxCursor = cursor
	case flowScreenCLI:
		if len(m.cliOptions) == 0 {
			m.cliCursor = 0
			break
		}
		if cursor > len(m.cliOptions)-1 {
			cursor = len(m.cliOptions) - 1
		}
		m.cliCursor = cursor
	case flowScreenAgentFlow:
		if len(m.agentflowOptions) == 0 {
			m.agentflowCursor = 0
			break
		}
		if cursor > len(m.agentflowOptions)-1 {
			cursor = len(m.agentflowOptions) - 1
		}
		m.agentflowCursor = cursor
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
	case flowScreenMCPList:
		if len(m.mcpListOptions) == 0 {
			m.mcpListCursor = 0
			break
		}
		if cursor > len(m.mcpListOptions)-1 {
			cursor = len(m.mcpListOptions) - 1
		}
		m.mcpListCursor = cursor
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
	case flowScreenInstallScope:
		scopeOpts := m.installScopeOptionsList()
		if len(scopeOpts) == 0 {
			m.installScopeCursor = 0
			break
		}
		if cursor > len(scopeOpts)-1 {
			cursor = len(scopeOpts) - 1
		}
		m.installScopeCursor = cursor
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
	case flowScreenToolbox:
		return m.toolboxCursor
	case flowScreenCLI:
		return m.cliCursor
	case flowScreenAgentFlow:
		return m.agentflowCursor
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
	case flowScreenMCPList:
		return m.mcpListCursor
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
	case flowScreenInstallScope:
		return m.installScopeCursor
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
	case flowScreenToolbox:
		return len(m.toolboxOptions)
	case flowScreenCLI:
		return len(m.cliOptions)
	case flowScreenAgentFlow:
		return len(m.agentflowOptions)
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
	case flowScreenMCPList:
		return len(m.mcpListOptions)
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
	case flowScreenInstallScope:
		return len(m.installScopeOptionsList())
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
		{
			Value:       "uninstall-rules",
			Label:       m.catalog.Msg("卸载项目规则文件", "Remove project rule files"),
			Badge:       m.catalog.Msg("删除", "DELETE"),
			Description: m.catalog.Msg("删除当前项目目录中 AgentFlow 管理的规则文件（不影响用户自定义文件）。", "Remove AgentFlow-managed rule files from this directory (user files are preserved)."),
		},
	}
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

// defaultBootstrapActionOptions returns the standard auto/manual action
// options used before any CLI has been installed. This is used to reset
// the bootstrap actions after navigating back from a post-install state.
func (m interactiveFlowModel) defaultBootstrapActionOptions() []Option {
	return []Option{
		{
			Value:       "auto",
			Label:       m.catalog.Msg("自动安装", "Automatic install"),
			Badge:       m.catalog.Msg("自动", "AUTO"),
			Description: m.catalog.Msg("自动检查 nvm / Node，并安装所选 CLI。", "Automatically verify nvm / Node and install the selected CLI."),
		},
		{
			Value:       "manual",
			Label:       m.catalog.Msg("查看手动安装提示", "Show manual install guidance"),
			Badge:       m.catalog.Msg("手动", "MANUAL"),
			Description: m.catalog.Msg("显示适合当前平台的手动安装步骤和命令。", "Show manual installation steps and commands for the current platform."),
		},
	}
}

// installedCLIActionOptions returns action options for a CLI that is already
// installed: reconfigure and uninstall.
func (m interactiveFlowModel) installedCLIActionOptions() []Option {
	return []Option{
		{
			Value:       "reconfigure",
			Label:       m.catalog.Msg("重新配置", "Reconfigure"),
			Badge:       m.catalog.Msg("配置", "CONFIG"),
			Description: m.catalog.Msg("重新设置 API Key、Base URL、默认模型、思考等级等配置项。", "Re-configure API Key, Base URL, default model, thinking level, and other settings."),
		},
		{
			Value:       "uninstall-single-cli",
			Label:       m.catalog.Msg("卸载此 CLI", "Uninstall this CLI"),
			Badge:       m.catalog.Msg("卸载", "REMOVE"),
			Description: m.catalog.Msg("卸载当前 CLI 工具及其配置目录。", "Uninstall this CLI tool and its configuration directory."),
		},
	}
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

func (m interactiveFlowModel) installScopeOptionsList() []Option {
	return []Option{
		{
			Value:       "scope-global",
			Label:       m.catalog.Msg("全局安装", "Global install"),
			Badge:       m.catalog.Msg("全局", "GLOBAL"),
			Description: m.catalog.Msg("将 AgentFlow 规则写入用户级配置目录（~/.codex, ~/.claude），对所有项目生效。", "Write AgentFlow rules into user-level config dirs (~/.codex, ~/.claude), effective for all projects."),
		},
		{
			Value:       "scope-project",
			Label:       m.catalog.Msg("项目级安装", "Project install"),
			Badge:       m.catalog.Msg("项目", "PROJECT"),
			Description: m.catalog.Msg("将 AgentFlow 规则文件写入当前工作目录，仅对本项目生效，适合团队协作。", "Write AgentFlow rule files into the current working directory, effective for this project only, ideal for team collaboration."),
		},
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

var spinnerFrames = []string{"◐", "◓", "◑", "◒"}

// copyToClipboard copies the given text to the system clipboard.
// Uses pbcopy on macOS, xclip or xsel on Linux.
func copyToClipboard(text string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "linux":
		if _, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		} else if _, err := exec.LookPath("xsel"); err == nil {
			cmd = exec.Command("xsel", "--clipboard", "--input")
		} else {
			return fmt.Errorf("no clipboard tool found")
		}
	default:
		return fmt.Errorf("clipboard not supported on %s", runtime.GOOS)
	}
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

// initConfigFieldState converts a ConfigField into an editable configFieldState,
// pre-filling CurrentValue (or Default) so the wizard shows existing config.
func initConfigFieldState(f ConfigField) configFieldState {
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
		// Prefer CurrentValue, fall back to Default.
		chosen := f.CurrentValue
		if chosen == "" {
			chosen = f.Default
		}
		state.Value = chosen
		for i, opt := range f.Options {
			if opt == chosen {
				state.OptionCursor = i
				break
			}
		}
	} else {
		// Text field: pre-fill with CurrentValue and place cursor at end.
		state.Value = f.CurrentValue
		state.CursorPos = len(f.CurrentValue)
	}
	return state
}
