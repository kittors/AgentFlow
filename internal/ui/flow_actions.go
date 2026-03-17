package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/kittors/AgentFlow/internal/debuglog"
)

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
		case flowActionCLIRefreshDetail:
			var detail Panel
			if m.callbacks.CLIDetailPanel != nil {
				target := ""
				if len(m.cliOptions) > 0 && m.cliCursor < len(m.cliOptions) {
					target = m.cliOptions[m.cliCursor].Value
				}
				if target != "" {
					detail = m.callbacks.CLIDetailPanel(target)
				}
			}
			return flowResultMsg{
				action:    action,
				status:    m.callbacks.Status(),
				cliDetail: panelRef(detail),
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
		case flowActionMCPInstallWithEnv:
			notice := Panel{}
			if m.callbacks.MCPInstallWithEnv != nil {
				notice = m.callbacks.MCPInstallWithEnv(selectedMCPTarget, selectedMCPServer, configEnvVars)
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
		case flowActionProjectRulesUninstall:
			notice := Panel{}
			if m.callbacks.ProjectRulesUninstall != nil {
				notice = m.callbacks.ProjectRulesUninstall(projectRoot, selectedSkillTarget)
			}
			projectRules := Panel{}
			if m.callbacks.ProjectRulesPanel != nil {
				projectRules = m.callbacks.ProjectRulesPanel(projectRoot, selectedSkillTarget)
			}
			return flowResultMsg{
				action:       action,
				notice:       panelRef(notice),
				status:       m.callbacks.Status(),
				projectRules: panelRef(projectRules),
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
		case flowActionUninstallProject:
			lines := []string{}
			for _, target := range selectedUninstallTargets {
				if !strings.HasPrefix(target, "project:") {
					continue
				}
				targetName := strings.TrimPrefix(target, "project:")
				if m.callbacks.ProjectRulesUninstall != nil {
					result := m.callbacks.ProjectRulesUninstall(projectRoot, targetName)
					lines = append(lines, result.Lines...)
				}
			}
			notice := Panel{
				Title: m.catalog.Msg("项目级卸载结果", "Project uninstall result"),
				Lines: lines,
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
