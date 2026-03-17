package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/kittors/AgentFlow/internal/debuglog"
)

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
	case flowToastClearMsg:
		m.toast = ""
		return m, nil
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
		if value.cliDetail != nil {
			m.cliDetail = value.cliDetail
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
		case flowActionCLIRefreshDetail:
			// CLI detail was already stored above via cliDetail field.
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
		case flowActionProjectRulesUninstall:
			if value.notice != nil {
				m.notice = value.notice
			}
			m.screen = flowScreenSkillProjectActions
			m.resetDetailFocus()
		case flowActionInstallHubRefresh:
			if value.notice != nil {
				m.installHubStatusPanel = value.notice
			}
		case flowActionMCPList, flowActionMCPRemove, flowActionMCPBatchRemove:
			if value.notice != nil {
				m.notice = value.notice
			}
			m.screen = flowScreenMCPActions
			m.resetDetailFocus()
			m.mcpInstallOptions = nil
			m.mcpRemoveOptions = nil
			m.mcpConfigMode = false
		case flowActionMCPInstall, flowActionMCPInstallWithEnv:
			if value.notice != nil {
				m.notice = value.notice
			}
			if m.mcpReconfigMode {
				// After reconfiguring from MCP list, return to the list view.
				m.mcpReconfigMode = false
				m.mcpConfigMode = false
				// Refresh list options.
				if m.callbacks.MCPRemoveOptions != nil {
					installed := m.callbacks.MCPRemoveOptions(m.selectedMCPTarget)
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
				m.screen = flowScreenMCPList
			} else {
				// Stay on the install screen and refresh options to show ✓ marks.
				if m.callbacks.MCPInstallOptions != nil {
					m.mcpInstallOptions = m.annotateRecommendedMCPOptions(
						m.selectedMCPTarget, m.callbacks.MCPInstallOptions(),
					)
				}
				m.mcpConfigMode = false
				m.screen = flowScreenMCPInstall
			}
		case flowActionMCPBatchInstall:
			if value.notice != nil {
				m.notice = value.notice
			}
			m.mcpInstallOptions = nil
			m.mcpRemoveOptions = nil
			m.pendingMCPInstalls = nil
			// If tavily-custom was also selected, redirect to config screen.
			if strings.EqualFold(m.selectedMCPServer, "tavily-custom") && m.callbacks.MCPConfigFields != nil {
				fields := m.callbacks.MCPConfigFields("tavily-custom")
				if len(fields) > 0 {
					m.configTarget = "tavily-custom"
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
					break
				}
			}
			m.screen = flowScreenMCPActions
			m.resetDetailFocus()
			m.mcpConfigMode = false
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
			m.refreshBootstrapDetail()
		case flowActionInstall, flowActionUninstall, flowActionUninstallCLI, flowActionUninstallProject, flowActionClean:
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
