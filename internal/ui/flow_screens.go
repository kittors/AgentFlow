package ui

import (
	"fmt"
	"strings"
)

func (m interactiveFlowModel) selectionForCurrentScreen() selectionModel {
	model := selectionModel{
		catalog: m.catalog,
		title:   fmt.Sprintf("AgentFlow v%s", m.version),
		width:   m.width,
		height:  m.height,
	}
	model.focusDetails = m.focusDetails
	model.detailScroll = m.detailScroll
	model.toast = m.toast

	switch m.screen {
	case flowScreenToolbox:
		model.subtitle = m.catalog.Msg("选择要管理的工具类型。Esc 返回主菜单。", "Choose the tool category to manage. Press Esc to return.")
		model.hint = m.catalog.Msg("↑/↓ 切换类型，Enter 进入，Esc 返回。", "Use ↑/↓ to switch category, Enter to continue, Esc to go back.")
		model.options = cloneOptions(m.toolboxOptions)
		model.cursor = m.toolboxCursor
		panels := make([]Panel, 0, 2)
		if m.notice != nil {
			panels = append(panels, *m.notice)
		}
		panels = append(panels, m.status)
		model.panels = panels
	case flowScreenCLI:
		model.subtitle = m.catalog.Msg("选择要管理的 CLI 工具。Esc 返回工具箱。", "Choose the CLI tool to manage. Press Esc to return to toolbox.")
		model.hint = m.catalog.Msg("↑/↓ 切换 CLI，Enter 进入安装/配置，Esc 返回。", "Use ↑/↓ to switch CLI, Enter to install/configure, Esc to go back.")
		model.options = cloneOptions(m.cliOptions)
		model.cursor = m.cliCursor
		panels := make([]Panel, 0, 2)
		if m.notice != nil {
			panels = append(panels, *m.notice)
		}
		if m.cliDetail != nil {
			panels = append(panels, *m.cliDetail)
		} else {
			panels = append(panels, m.status)
		}
		model.panels = panels
	case flowScreenAgentFlow:
		model.subtitle = fmt.Sprintf(m.catalog.Msg("管理 AgentFlow 规则。当前目录: %s。Esc 返回主菜单。", "Manage AgentFlow rules. Current directory: %s. Press Esc to return."), m.projectRoot)
		model.hint = m.catalog.Msg("↑/↓ 选择操作，Enter 执行，Esc 返回。", "Use ↑/↓ to choose an action, Enter to run, Esc to go back.")
		model.options = cloneOptions(m.agentflowOptions)
		model.cursor = m.agentflowCursor
		panels := make([]Panel, 0, 2)
		if m.notice != nil {
			panels = append(panels, *m.notice)
		}
		panels = append(panels, m.status)
		model.panels = panels
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
		model.options = cloneOptions(m.dynamicMCPActions())
		model.cursor = m.mcpActionCursor
		model.panels = m.mcpActionPanels()
	case flowScreenMCPInstall:
		model.subtitle = fmt.Sprintf(m.catalog.Msg("为 %s 安装推荐 MCP。Esc 返回。", "Install recommended MCP for %s. Esc to go back."), m.selectedMCPTarget)
		model.hint = m.catalog.Msg("↑/↓ 选择 MCP，Enter 安装，Esc 返回。", "Use ↑/↓ to choose an MCP server, Enter to install, Esc to go back.")
		model.options = cloneOptions(m.mcpInstallOptions)
		model.cursor = m.mcpInstallCursor
		model.panels = m.mcpInstallPanels()
	case flowScreenMCPList:
		model.subtitle = fmt.Sprintf(m.catalog.Msg("已安装的 MCP 列表 (%s)。Esc 返回。", "Installed MCP list (%s). Esc to go back."), m.selectedMCPTarget)
		model.hint = m.catalog.Msg("↑/↓ 浏览已安装 MCP，Esc 返回。", "Use ↑/↓ to browse installed MCPs, Esc to go back.")
		model.options = cloneOptions(m.mcpListOptions)
		model.cursor = m.mcpListCursor
		// Show the description of the currently selected MCP as the panel.
		panels := make([]Panel, 0, 2)
		if m.notice != nil {
			panels = append(panels, *m.notice)
		}
		if len(m.mcpListOptions) > 0 {
			idx := m.mcpListCursor
			if idx >= len(m.mcpListOptions) {
				idx = len(m.mcpListOptions) - 1
			}
			opt := m.mcpListOptions[idx]
			desc := opt.Description
			if desc == "" {
				desc = m.catalog.Msg("该 MCP 已配置。", "This MCP is configured.")
			}
			panels = append(panels, Panel{
				Title: opt.Value,
				Lines: []string{desc},
			})
		}
		model.panels = panels
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
		model.subtitle = m.catalog.Msg("选择部署 Profile。Esc 返回范围选择。", "Select a deployment profile. Press Esc to return to scope selection.")
		model.hint = m.catalog.Msg("↑/↓ 切换 Profile，Enter 下一步，Esc 返回。", "Use ↑/↓ to switch profiles, Enter to continue, Esc to go back.")
		model.options = cloneOptions(m.profileOptions)
		model.cursor = m.profileCursor
		model.panels = m.profilePanels()
	case flowScreenInstallScope:
		model.subtitle = m.catalog.Msg("选择安装范围：全局或项目级。Esc 返回安装中心。", "Choose install scope: global or project-level. Press Esc to return to install hub.")
		model.hint = m.catalog.Msg("↑/↓ 切换范围，Enter 确认，Esc 返回。", "Use ↑/↓ to switch scope, Enter to confirm, Esc to go back.")
		scopeOptions := m.installScopeOptionsList()
		model.options = cloneOptions(scopeOptions)
		model.cursor = m.installScopeCursor
		panels := make([]Panel, 0, 2)
		if m.notice != nil {
			panels = append(panels, *m.notice)
		}
		panels = append(panels, Panel{
			Title: m.catalog.Msg("安装范围说明", "Install scope"),
			Lines: []string{
				m.catalog.Msg("全局安装：写入用户配置目录（如 ~/.codex, ~/.claude），所有项目共享。", "Global: writes to user config dirs (e.g. ~/.codex, ~/.claude), shared across all projects."),
				m.catalog.Msg("项目安装：写入当前工作目录，仅对当前项目生效，适合团队共享。", "Project: writes to current working directory, effective for this project only, ideal for team sharing."),
				"",
				fmt.Sprintf(m.catalog.Msg("当前目录: %s", "Current directory: %s"), m.projectRoot),
			},
		})
		model.panels = panels
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
		// Step-by-step wizard: show one field at a time.
		stepNum := m.configFieldCursor + 1
		totalSteps := len(m.configFields)
		if m.configFieldCursor >= 0 && m.configFieldCursor < len(m.configFields) {
			f := m.configFields[m.configFieldCursor]
			fieldLabel := f.Label
			if f.EnvVar != "" && !strings.HasPrefix(f.EnvVar, "__") {
				fieldLabel = fmt.Sprintf("%s (%s)", f.Label, f.EnvVar)
			}
			model.subtitle = fmt.Sprintf(
				m.catalog.Msg("配置 %s — 步骤 %d/%d：%s", "Configure %s — Step %d/%d: %s"),
				m.configTarget, stepNum, totalSteps, fieldLabel,
			)
			if f.FieldType == "select" {
				model.hint = m.catalog.Msg("←/→ 切换选项，Enter 确认并继续，Esc 跳过剩余步骤。", "←/→ to switch options, Enter to confirm and continue, Esc to skip remaining.")
			} else {
				model.hint = m.catalog.Msg("直接输入文本，Enter 确认并继续（留空跳过），Esc 跳过剩余步骤。", "Type to input, Enter to confirm and continue (empty = skip), Esc to skip remaining.")
			}

			// Left option: just show step name, no value echo.
			typeIcon := "📝"
			if f.FieldType == "select" {
				typeIcon = "🔽"
			}
			model.options = []Option{{
				Value:       f.EnvVar,
				Label:       fmt.Sprintf("%s %s", typeIcon, fieldLabel),
				Badge:       fmt.Sprintf("%d/%d", stepNum, totalSteps),
				Description: m.catalog.Msg("在右侧面板输入 →", "Input in the right panel →"),
			}}
			model.cursor = 0
		}

		// Build detail panel with prominent input area.
		var summaryLines []string
		summaryLines = append(summaryLines,
			fmt.Sprintf(m.catalog.Msg("目标: %s", "Target: %s"), m.configTarget),
			"",
		)

		for idx, f := range m.configFields {
			label := f.Label
			if idx < m.configFieldCursor {
				// Completed fields.
				val := f.Value
				if val == "" {
					val = m.catalog.Msg("(已跳过)", "(skipped)")
				} else if f.FieldType == "text" && len(val) > 6 {
					if strings.Contains(strings.ToLower(f.Label), "key") {
						val = val[:3] + "***" + val[len(val)-3:]
					}
				}
				summaryLines = append(summaryLines,
					fmt.Sprintf("  ✅ %s: %s", label, val))
			} else if idx == m.configFieldCursor {
				// Current field — show prominent input.
				summaryLines = append(summaryLines, "")
				if f.FieldType == "select" && len(f.Options) > 0 {
					current := f.Options[f.OptionCursor]
					summaryLines = append(summaryLines,
						fmt.Sprintf("  🔽 %s:", label),
						fmt.Sprintf("     ◀ %s ▶  (%d/%d)", current, f.OptionCursor+1, len(f.Options)),
					)
				} else {
					summaryLines = append(summaryLines,
						fmt.Sprintf("  ✏️  %s:", label))

					// Prominent input box.
					var inputDisplay string
					if f.Value == "" {
						inputDisplay = m.catalog.Msg("请在此输入…", "type here…")
					} else {
						// Show value with cursor.
						pos := f.CursorPos
						if pos > len(f.Value) {
							pos = len(f.Value)
						}
						inputDisplay = f.Value[:pos] + "█" + f.Value[pos:]
					}
					// Styled input line: ┃ value █ ┃
					summaryLines = append(summaryLines,
						"  ┌──────────────────────────────────────┐",
						fmt.Sprintf("  │ %s", inputDisplay),
						"  └──────────────────────────────────────┘",
					)
				}
				summaryLines = append(summaryLines, "")
			} else {
				summaryLines = append(summaryLines,
					fmt.Sprintf("  ○  %s: %s",
						label,
						m.catalog.Msg("待输入", "pending")))
			}
		}
		model.panels = []Panel{{
			Title: m.catalog.Msg("配置进度", "Configuration progress"),
			Lines: summaryLines,
		}}
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

// dynamicMCPActions returns the mcpActions list, but hides "list" when no MCPs
// are installed for the current target.
func (m interactiveFlowModel) dynamicMCPActions() []Option {
	hasMCPs := false
	if m.callbacks.MCPRemoveOptions != nil {
		hasMCPs = len(m.callbacks.MCPRemoveOptions(m.selectedMCPTarget)) > 0
	}
	if hasMCPs {
		return m.mcpActions
	}
	// Filter out "list" action.
	filtered := make([]Option, 0, len(m.mcpActions))
	for _, opt := range m.mcpActions {
		if opt.Value != "list" {
			filtered = append(filtered, opt)
		}
	}
	return filtered
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
