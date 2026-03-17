package ui

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

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
			case flowScreenCLI:
				return m.startBusy(flowActionCLIRefreshDetail, m.catalog.Msg("正在读取 CLI 详情…", "Loading CLI details..."))
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
			case flowScreenCLI:
				return m.startBusy(flowActionCLIRefreshDetail, m.catalog.Msg("正在读取 CLI 详情…", "Loading CLI details..."))
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
			case flowScreenCLI:
				return m.startBusy(flowActionCLIRefreshDetail, m.catalog.Msg("正在读取 CLI 详情…", "Loading CLI details..."))
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
			case flowScreenCLI:
				return m.startBusy(flowActionCLIRefreshDetail, m.catalog.Msg("正在读取 CLI 详情…", "Loading CLI details..."))
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
			case flowScreenCLI:
				return m.startBusy(flowActionCLIRefreshDetail, m.catalog.Msg("正在读取 CLI 详情…", "Loading CLI details..."))
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
			case flowScreenCLI:
				return m.startBusy(flowActionCLIRefreshDetail, m.catalog.Msg("正在读取 CLI 详情…", "Loading CLI details..."))
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
		ch := key.String()
		if ch == " " {
			if m.screen == flowScreenInstallTargets {
				m.toggleSelected(&m.installOptions, m.installCursor)
			}
			if m.screen == flowScreenUninstallTargets {
				m.toggleSelected(&m.uninstallOptions, m.uninstallCursor)
			}
		}
		if ch == "c" || ch == "C" {
			// Copy detail panel content to clipboard.
			screen := m.selectionForCurrentScreen()
			var textLines []string
			for _, panel := range screen.panels {
				if panel.Title != "" {
					textLines = append(textLines, panel.Title)
				}
				textLines = append(textLines, panel.Lines...)
				textLines = append(textLines, "")
			}
			if len(textLines) > 0 {
				text := strings.Join(textLines, "\n")
				if copyToClipboard(text) == nil {
					m.toast = m.catalog.Msg("✅ 已复制到剪贴板", "✅ Copied to clipboard")
					return m, tea.Tick(1500*time.Millisecond, func(time.Time) tea.Msg {
						return flowToastClearMsg{}
					})
				}
			}
		}
	}
	return m, nil
}

func (m interactiveFlowModel) handleBack() (tea.Model, tea.Cmd) {
	switch m.screen {
	case flowScreenMain:
		return m, tea.Quit
	case flowScreenToolbox:
		m.screen = flowScreenMain
	case flowScreenCLI:
		m.screen = flowScreenToolbox
	case flowScreenAgentFlow:
		m.screen = flowScreenMain
	case flowScreenInstallHub:
		m.screen = flowScreenMain
	case flowScreenMCPTargets:
		m.screen = flowScreenToolbox
	case flowScreenMCPActions:
		m.screen = flowScreenMCPTargets
	case flowScreenMCPInstall, flowScreenMCPRemove, flowScreenMCPList:
		m.screen = flowScreenMCPActions
	case flowScreenSkillTargets:
		if m.projectInstallMode {
			m.projectInstallMode = false
			m.screen = flowScreenInstallHub
		} else {
			m.screen = flowScreenToolbox
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
		m.screen = flowScreenCLI
	case flowScreenBootstrapActions:
		if m.enteredFromCLI {
			// Return to CLI list when we entered from CLI screen.
			m.screen = flowScreenCLI
			m.enteredFromCLI = false
		} else {
			m.screen = flowScreenBootstrapTargets
		}
		// Reset post-install state so selecting a different CLI target starts fresh.
		m.pendingConfigFields = nil
		m.bootstrapActionOptions = m.defaultBootstrapActionOptions()
	case flowScreenProfile:
		m.screen = flowScreenInstallScope
	case flowScreenInstallScope:
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
		if m.mcpConfigMode {
			m.screen = flowScreenMCPInstall
			m.mcpConfigMode = false
		} else {
			m.screen = flowScreenBootstrapActions
		}
		m.configEditing = false
	}
	m.notice = nil
	return m, nil
}

// handleConfigKey processes key events during the config step-by-step wizard.
func (m interactiveFlowModel) handleConfigKey(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.Type {
	case tea.KeyEsc:
		// Skip remaining config and save what we have.
		m.focusDetails = false
		m.detailScroll = 0
		return m.handleEnter()
	case tea.KeyEnter:
		// Accept current field and advance to next.
		if m.configFieldCursor >= 0 && m.configFieldCursor < len(m.configFields) {
			f := &m.configFields[m.configFieldCursor]
			// For select fields, mark dirty if user pressed Enter (accepting default counts).
			if f.FieldType == "select" && len(f.Options) > 0 {
				f.Value = f.Options[f.OptionCursor]
				f.Dirty = true
			}
			// Mark text fields as dirty if they have content.
			if f.FieldType != "select" && strings.TrimSpace(f.Value) != "" {
				f.Dirty = true
			}
		}
		// Advance to next field or save if done.
		if m.configFieldCursor < len(m.configFields)-1 {
			m.configFieldCursor++
			return m, nil
		}
		// All fields completed — trigger save.
		m.focusDetails = false
		m.detailScroll = 0
		return m.handleEnter()
	case tea.KeyLeft:
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
			} else if f.FieldType != "select" {
				// Move text cursor left.
				if f.CursorPos > 0 {
					f.CursorPos--
				}
			}
		}
		return m, nil
	case tea.KeyRight:
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
			} else if f.FieldType != "select" {
				// Move text cursor right.
				if f.CursorPos < len(f.Value) {
					f.CursorPos++
				}
			}
		}
		return m, nil
	case tea.KeyBackspace:
		if m.configFieldCursor >= 0 && m.configFieldCursor < len(m.configFields) {
			f := &m.configFields[m.configFieldCursor]
			if f.FieldType != "select" && f.CursorPos > 0 {
				// Delete character before cursor.
				f.Value = f.Value[:f.CursorPos-1] + f.Value[f.CursorPos:]
				f.CursorPos--
				f.Dirty = true
			}
		}
		return m, nil
	case tea.KeyRunes:
		if m.configFieldCursor >= 0 && m.configFieldCursor < len(m.configFields) {
			f := &m.configFields[m.configFieldCursor]
			if f.FieldType != "select" {
				// Insert characters at cursor position.
				// Strip bracket chars from bracketed paste sequences.
				ins := strings.NewReplacer("[", "", "]", "").Replace(key.String())
				if ins != "" {
					f.Value = f.Value[:f.CursorPos] + ins + f.Value[f.CursorPos:]
					f.CursorPos += len(ins)
					f.Dirty = true
				}
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
		case ActionToolbox:
			m.screen = flowScreenToolbox
			m.toolboxCursor = 0
			m.notice = nil
			return m, nil
		case ActionAgentFlow:
			m.screen = flowScreenAgentFlow
			m.agentflowCursor = 0
			m.notice = nil
			return m, nil
		case ActionUpdate:
			return m.startBusy(flowActionUpdate, m.catalog.Msg("正在检查最新版本并更新…", "Checking the latest release and updating..."))
		case ActionClean:
			return m.startBusy(flowActionClean, m.catalog.Msg("正在清理缓存…", "Cleaning caches..."))
		case ActionExit:
			return m, tea.Quit
		}
	case flowScreenToolbox:
		if len(m.toolboxOptions) == 0 {
			return m, nil
		}
		switch Action(m.toolboxOptions[m.toolboxCursor].Value) {
		case ActionCLI:
			// Build CLI options from bootstrap options.
			m.bootstrapOptions = m.bootstrapOptionsList()
			if len(m.bootstrapOptions) == 0 {
				m.notice = panelRef(Panel{
					Title: m.catalog.Msg("CLI", "CLI"),
					Lines: []string{m.catalog.Msg("未检测到可管理的 CLI。", "No manageable CLIs detected.")},
				})
				return m, nil
			}
			m.cliOptions = cloneOptions(m.bootstrapOptions)
			m.screen = flowScreenCLI
			m.cliCursor = 0
			m.notice = nil
			m.focusDetails = false
			m.detailScroll = 0
			// Refresh CLI detail panel for the first CLI.
			if len(m.cliOptions) > 0 {
				return m.startBusy(flowActionCLIRefreshDetail, m.catalog.Msg("正在读取 CLI 详情…", "Loading CLI details..."))
			}
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
		}
	case flowScreenCLI:
		if len(m.cliOptions) == 0 {
			return m, nil
		}
		m.selectedBootstrapTarget = m.cliOptions[m.cliCursor].Value
		m.screen = flowScreenBootstrapActions
		m.enteredFromCLI = true
		m.notice = nil
		m.focusDetails = false
		m.detailScroll = 0

		// Check if CLI is already installed → show configure/uninstall instead of install.
		cliInstalled := false
		if m.callbacks.CLIInstalled != nil {
			cliInstalled = m.callbacks.CLIInstalled(m.selectedBootstrapTarget)
		}

		if cliInstalled {
			// CLI is installed: show "Configure" and "Uninstall CLI".
			m.bootstrapActionOptions = m.installedCLIActionOptions()
			m.bootstrapActionCursor = 0
		} else {
			// CLI is not installed: show "Auto install" / "Manual install".
			m.bootstrapActionOptions = m.defaultBootstrapActionOptions()
			m.bootstrapActionCursor = 0
			if m.callbacks.BootstrapAutoSupported != nil && !m.callbacks.BootstrapAutoSupported(m.selectedBootstrapTarget) {
				m.bootstrapActionCursor = 1
			}
		}
		m.refreshBootstrapDetail()
		return m, nil
	case flowScreenAgentFlow:
		if len(m.agentflowOptions) == 0 {
			return m, nil
		}
		switch m.agentflowOptions[m.agentflowCursor].Value {
		case "install-global":
			// Global install: select profile → select targets → install
			m.projectInstallMode = false
			m.screen = flowScreenProfile
			m.profileCursor = 2 // default to "full"
			m.notice = nil
			return m, nil
		case "install-project":
			// Project install: select targets → select profile → install
			m.projectInstallMode = true
			if m.callbacks.UninstallProjectOptions != nil {
				m.skillTargets = cloneOptions(m.callbacks.UninstallProjectOptions())
			}
			if m.callbacks.SkillTargetOptions != nil {
				m.skillTargets = cloneOptions(m.callbacks.SkillTargetOptions())
			}
			m.installOptions = m.installOptionsList()
			m.screen = flowScreenInstallTargets
			m.installCursor = 0
			m.notice = nil
			return m, nil
		case "uninstall":
			m.uninstallCLIMode = false
			if m.callbacks.UninstallOptions != nil {
				m.uninstallOptions = cloneOptions(m.callbacks.UninstallOptions())
			}
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
		actions := m.dynamicMCPActions()
		if len(actions) == 0 || m.mcpActionCursor >= len(actions) {
			return m, nil
		}
		switch actions[m.mcpActionCursor].Value {
		case "list":
			// Build list options from installed MCPs with their original descriptions.
			if m.callbacks.MCPRemoveOptions != nil {
				installed := m.callbacks.MCPRemoveOptions(m.selectedMCPTarget)
				// Build description map from install options for clean descriptions.
				descMap := map[string]string{}
				if m.callbacks.MCPInstallOptions != nil {
					for _, opt := range m.callbacks.MCPInstallOptions() {
						descMap[strings.ToLower(opt.Value)] = opt.Description
					}
				}
				m.mcpListOptions = make([]Option, 0, len(installed))
				for _, opt := range installed {
					desc := descMap[strings.ToLower(opt.Value)]
					if desc == "" {
						desc = m.catalog.Msg("已配置。", "Configured.")
					}
					m.mcpListOptions = append(m.mcpListOptions, Option{
						Value:       opt.Value,
						Label:       opt.Value,
						Badge:       "✓",
						Description: desc,
					})
				}
			}
			if len(m.mcpListOptions) == 0 {
				m.notice = panelRef(Panel{
					Title: m.catalog.Msg("MCP", "MCP"),
					Lines: []string{m.catalog.Msg("未配置任何 MCP servers。", "No MCP servers configured.")},
				})
				return m, nil
			}
			m.screen = flowScreenMCPList
			m.mcpListCursor = 0
			m.notice = nil
			m.focusDetails = false
			m.detailScroll = 0
			return m, nil
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
		// Check if this server needs config fields (e.g. tavily-custom)
		if m.callbacks.MCPConfigFields != nil {
			fields := m.callbacks.MCPConfigFields(m.selectedMCPServer)
			if len(fields) > 0 {
				m.configTarget = m.selectedMCPServer
				m.configFields = make([]configFieldState, len(fields))
				for idx, f := range fields {
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
				m.mcpConfigMode = true
				m.screen = flowScreenBootstrapConfig
				return m, nil
			}
		}
		return m.startBusy(flowActionMCPInstall, m.catalog.Msg("正在写入 MCP 配置…", "Writing MCP configuration…"))
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
		case "uninstall-rules":
			return m.startBusy(flowActionProjectRulesUninstall, m.catalog.Msg("正在卸载项目规则文件…", "Removing project rule files..."))
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
			m.screen = flowScreenInstallScope
			m.installScopeCursor = 0
			m.notice = nil
			return m, nil
		case "uninstall-agentflow":
			m.uninstallCLIMode = false
			// Build combined uninstall targets: global + project.
			var options []Option
			if m.callbacks.UninstallOptions != nil {
				for _, opt := range m.callbacks.UninstallOptions() {
					opt.Badge = m.catalog.Msg("全局", "GLOBAL")
					options = append(options, opt)
				}
			}
			if m.callbacks.UninstallProjectOptions != nil {
				for _, opt := range m.callbacks.UninstallProjectOptions() {
					options = append(options, opt)
				}
			}
			if len(options) == 0 {
				m.notice = panelRef(Panel{
					Title: m.catalog.Msg("卸载结果", "Uninstall result"),
					Lines: []string{m.catalog.Msg("未检测到已安装的 AgentFlow（全局或项目级）。", "No AgentFlow installations found (global or project-level).")},
				})
				return m, nil
			}
			m.uninstallOptions = options
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
	case flowScreenInstallScope:
		scopeOptions := m.installScopeOptionsList()
		if m.installScopeCursor >= len(scopeOptions) {
			return m, nil
		}
		switch scopeOptions[m.installScopeCursor].Value {
		case "scope-global":
			m.installOptions = m.installOptionsList()
			if len(m.installOptions) == 0 {
				m.notice = panelRef(Panel{
					Title: m.catalog.Msg("安装提示", "Install hint"),
					Lines: []string{
						m.catalog.Msg("还没有可安装 AgentFlow 的 CLI。先进入「安装 CLI 工具」分支完成 Codex 或 Claude 的安装。", "There are no CLI targets ready for AgentFlow yet. Use the CLI install branch first to install Codex or Claude."),
					},
				})
				return m, nil
			}
			m.screen = flowScreenProfile
			m.notice = nil
			return m, nil
		case "scope-project":
			if m.callbacks.SkillTargetOptions == nil {
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
		}
	case flowScreenBootstrapTargets:
		if len(m.bootstrapOptions) == 0 {
			return m, nil
		}
		m.selectedBootstrapTarget = m.bootstrapOptions[m.bootstrapCursor].Value
		// Reset bootstrap options to default auto/manual so stale post-install
		// options (configure/done) from a previous target don't persist.
		m.pendingConfigFields = nil
		m.bootstrapActionOptions = m.defaultBootstrapActionOptions()
		m.screen = flowScreenBootstrapActions
		m.enteredFromCLI = false
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
		case "reconfigure":
			// Load config fields for this CLI and enter config wizard.
			if m.callbacks.CLIConfigFields != nil {
				fields := m.callbacks.CLIConfigFields(m.selectedBootstrapTarget)
				if len(fields) > 0 {
					m.configTarget = m.selectedBootstrapTarget
					m.configFields = make([]configFieldState, len(fields))
					for idx, f := range fields {
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
					m.mcpConfigMode = false
					m.screen = flowScreenBootstrapConfig
					return m, nil
				}
			}
			m.notice = panelRef(Panel{
				Title: m.catalog.Msg("配置提示", "Configuration note"),
				Lines: []string{m.catalog.Msg("该 CLI 没有可配置的选项。", "This CLI has no configurable options.")},
			})
			return m, nil
		case "uninstall-single-cli":
			// Uninstall single CLI.
			m.uninstallCLIMode = true
			m.uninstallOptions = []Option{
				{Value: m.selectedBootstrapTarget, Selected: true},
			}
			return m.startBusy(flowActionUninstallCLI, m.catalog.Msg("正在卸载所选 CLI…", "Uninstalling selected CLI..."))
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
		// Check if any selected targets are project-level (prefixed with "project:").
		hasProject := false
		hasGlobal := false
		for _, v := range selected {
			if strings.HasPrefix(v, "project:") {
				hasProject = true
			} else {
				hasGlobal = true
			}
		}
		if hasProject && !hasGlobal {
			return m.startBusy(flowActionUninstallProject, m.catalog.Msg("正在卸载项目级规则…", "Uninstalling project-level rules..."))
		}
		if hasGlobal && !hasProject {
			return m.startBusy(flowActionUninstall, m.catalog.Msg("正在卸载所选目标…", "Uninstalling selected targets..."))
		}
		// Both: run global first, then project.
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
		if m.mcpConfigMode {
			// MCP config mode: install the MCP server with collected env vars.
			if len(envVars) == 0 {
				m.screen = flowScreenMCPInstall
				m.mcpConfigMode = false
				m.configEditing = false
				m.notice = panelRef(Panel{
					Title: m.catalog.Msg("配置跳过", "Configuration skipped"),
					Lines: []string{m.catalog.Msg("未输入任何配置，已取消安装。", "No configuration entered; install cancelled.")},
				})
				return m, nil
			}
			m.configEditing = false
			return m.startBusy(flowActionMCPInstallWithEnv, m.catalog.Msg("正在写入 MCP 配置…", "Writing MCP configuration..."))
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
