package app

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/kittors/AgentFlow/internal/debuglog"
	"github.com/kittors/AgentFlow/internal/projectrules"
	"github.com/kittors/AgentFlow/internal/ui"
	"github.com/kittors/AgentFlow/internal/update"
)

func (a *App) cleanPanel() ui.Panel {
	cleaned, err := a.Installer.Clean()
	if err != nil {
		return errorPanel(a.Catalog.Msg("清理失败", "Clean failed"), err)
	}
	return ui.Panel{
		Title: a.Catalog.Msg("清理结果", "Clean result"),
		Lines: []string{
			fmt.Sprintf(a.Catalog.Msg("已清理 %d 个缓存目录。", "Cleaned %d cache directories."), cleaned),
		},
	}
}

func (a *App) statusPanel() ui.Panel {
	done := debuglog.Timed("statusPanel")
	defer done()
	lines := make([]string, 0, 16)
	if executable, err := os.Executable(); err == nil {
		lines = append(lines, fmt.Sprintf(a.Catalog.Msg("可执行文件: %s", "Executable: %s"), executable))
	}
	lines = append(lines, "")
	lines = append(lines, a.Installer.RuntimeSummaryLines()...)
	lines = append(lines, "")
	lines = append(lines, a.Installer.StatusLines()...)

	// Project-level rules status.
	if wd, wdErr := os.Getwd(); wdErr == nil {
		rulesManager := projectrules.NewManager()
		statuses, detectErr := rulesManager.Detect(wd)
		if detectErr == nil {
			hasAny := false
			for _, status := range statuses {
				if status.Exists {
					hasAny = true
					break
				}
			}
			if hasAny {
				lines = append(lines, "")
				lines = append(lines, a.Catalog.Msg("项目级规则:", "Project rules:"))
				for _, status := range statuses {
					if !status.Exists {
						continue
					}
					state := a.Catalog.Msg("已安装（AgentFlow）", "installed (AgentFlow)")
					if !status.Managed {
						state = a.Catalog.Msg("已存在（用户自定义）", "present (user)")
					}
					lines = append(lines, fmt.Sprintf("  %s: %s", status.Detected, state))
				}
			}
		}
	}

	if result, err := a.Checker.Check(a.Version, update.Options{CacheTTLHours: 72}); err == nil && result.UpdateAvailable {
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf(a.Catalog.Msg("可更新到 v%s", "Update available: v%s"), result.Latest))
	}
	return ui.Panel{
		Title: a.Catalog.Msg("环境状态", "Environment"),
		Lines: lines,
	}
}

func (a *App) mainMenuPanels(notice *ui.Panel) []ui.Panel {
	panels := make([]ui.Panel, 0, 2)
	if notice != nil && (strings.TrimSpace(notice.Title) != "" || len(notice.Lines) > 0) {
		panels = append(panels, *notice)
	}
	panels = append(panels, a.statusPanel())
	return panels
}

func (a *App) installTargetOptions() []ui.Option {
	statuses := a.Installer.DetectTargetStatuses()
	options := make([]ui.Option, 0, len(statuses))
	for _, status := range statuses {
		if !status.CLIInstalled && !status.AgentFlowInstalled && !status.ConfigDirExists {
			continue
		}

		description := a.Catalog.Msg("可继续部署 AgentFlow。", "Ready for AgentFlow deployment.")
		switch {
		case status.CLIInstalled && status.AgentFlowInstalled:
			description = a.Catalog.Msg("CLI 与 AgentFlow 都已就绪；再次执行会刷新到当前版本。", "Both the CLI and AgentFlow are ready; rerunning refreshes to the current version.")
		case status.CLIInstalled:
			description = a.Catalog.Msg("CLI 已安装，可直接部署 AgentFlow。", "The CLI is installed and ready for AgentFlow.")
		case status.AgentFlowInstalled:
			description = a.Catalog.Msg("已存在 AgentFlow 文件，但未检测到 CLI 可执行文件。", "AgentFlow files exist, but the CLI executable was not detected.")
		case status.ConfigDirExists:
			description = a.Catalog.Msg("已检测到配置目录，可提前写入 AgentFlow。", "A config directory was detected, so AgentFlow can be written in advance.")
		}
		options = append(options, ui.Option{
			Value:       status.Target.Name,
			Label:       status.Target.DisplayName,
			Badge:       strings.ToUpper(status.Target.Name),
			Description: description,
		})
	}
	return options
}

func (a *App) bootstrapTargetOptions() []ui.Option {
	statuses := a.Installer.DetectBootstrapTargetStatuses()
	options := make([]ui.Option, 0, len(statuses))
	for _, status := range statuses {
		description := a.Catalog.Msg("未检测到该 CLI，可执行快速安装。", "The CLI was not detected and can be installed quickly.")
		switch {
		case status.CLIInstalled && status.AgentFlowInstalled:
			description = a.Catalog.Msg("CLI 与 AgentFlow 都已就绪；可重装 CLI 或直接返回。", "Both the CLI and AgentFlow are ready; reinstall if needed or go back.")
		case status.CLIInstalled:
			description = a.Catalog.Msg("CLI 已安装，可直接切到 AgentFlow 安装分支。", "The CLI is already installed; switch to the AgentFlow install branch if needed.")
		case !status.AutoInstallSupported:
			description = a.Catalog.Msg("当前环境不满足自动安装条件；按 Enter 进入安装方式，再查看手动安装提示。", "Automatic installation is not available in this environment; press Enter to open install modes, then view the manual guidance.")
		}
		options = append(options, ui.Option{
			Value:       status.Target.Name,
			Label:       status.Target.DisplayName,
			Badge:       strings.ToUpper(status.Target.Name),
			Description: description,
		})
	}
	return options
}

func (a *App) bootstrapAutoSupported(targetName string) bool {
	status, err := a.Installer.DetectTargetStatus(targetName)
	if err != nil {
		return true
	}
	return status.AutoInstallSupported
}

func (a *App) bootstrapTargetPanel(targetName string) ui.Panel {
	status, err := a.Installer.DetectTargetStatus(targetName)
	if err != nil {
		return errorPanel(a.Catalog.Msg("CLI 信息", "CLI details"), err)
	}

	// Styles for the detail panel.
	greenDot := lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("●")
	grayDot := lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render("○")
	labelStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	lines := make([]string, 0, 24)

	// ── Installation status section ──
	if status.CLIInstalled {
		location := status.CLIPath
		if strings.TrimSpace(status.CLIPathScope) != "" {
			location = fmt.Sprintf("%s (%s)", status.CLIPath, status.CLIPathScope)
		}
		lines = append(lines, fmt.Sprintf("%s %s: %s", greenDot,
			labelStyle.Render(a.Catalog.Msg("CLI 状态", "CLI status")),
			valueStyle.Render(a.Catalog.Msg("已安装", "installed"))))
		lines = append(lines, fmt.Sprintf("  %s", mutedStyle.Render(location)))
	} else {
		lines = append(lines, fmt.Sprintf("%s %s: %s", grayDot,
			labelStyle.Render(a.Catalog.Msg("CLI 状态", "CLI status")),
			mutedStyle.Render(a.Catalog.Msg("未安装", "not installed"))))
	}

	if status.AgentFlowInstalled {
		lines = append(lines, fmt.Sprintf("%s %s: %s", greenDot,
			labelStyle.Render("AgentFlow"),
			valueStyle.Render(a.Catalog.Msg("已安装", "installed"))))
	} else {
		lines = append(lines, fmt.Sprintf("%s %s: %s", grayDot,
			labelStyle.Render("AgentFlow"),
			mutedStyle.Render(a.Catalog.Msg("未安装", "not installed"))))
	}

	// ── Configuration section (API Key / Base URL / Model) ──
	// Only show when CLI is actually installed.
	if status.CLIInstalled {
		target := status.Target
		lines = append(lines, "")
		lines = append(lines, labelStyle.Render(a.Catalog.Msg("─── 配置状态 ───", "─── Configuration ───")))

		// API Key: read from env/rc for all targets, also check auth.json for Codex.
		if target.APIKeyEnv != "" {
			envVal := a.Installer.GetEnvOrRC(target.APIKeyEnv)
			if envVal == "" && target.Name == "codex" {
				envVal = a.Installer.ReadCodexAuthKey()
			}
			if envVal != "" {
				displayVal := envVal
				if len(envVal) > 6 {
					displayVal = envVal[:3] + strings.Repeat("*", len(envVal)-6) + envVal[len(envVal)-3:]
				}
				lines = append(lines, fmt.Sprintf("  %s API Key: %s",
					greenDot, valueStyle.Render(displayVal)))
			} else {
				lines = append(lines, fmt.Sprintf("  %s API Key: %s",
					grayDot, mutedStyle.Render(a.Catalog.Msg("未设置", "not set"))))
			}
		}

		// Base URL: read from env/rc, also check config.toml model_provider for Codex.
		if target.BaseURLEnv != "" {
			envVal := a.Installer.GetEnvOrRC(target.BaseURLEnv)
			if envVal == "" && target.Name == "codex" {
				envVal = a.Installer.ReadCodexConfigField("base_url")
			}
			if envVal != "" {
				lines = append(lines, fmt.Sprintf("  %s Base URL: %s",
					greenDot, valueStyle.Render(strings.TrimRight(envVal, "/"))))
			} else {
				lines = append(lines, fmt.Sprintf("  %s Base URL: %s",
					grayDot, mutedStyle.Render(a.Catalog.Msg("未设置", "not set"))))
			}
		}

		// Model: read from config file (config.toml for Codex, .claude.json for Claude).
		modelLabel := a.Catalog.Msg("模型", "Model")
		modelVal := a.Installer.ReadCLIConfigModel(target.Name)
		if modelVal != "" {
			lines = append(lines, fmt.Sprintf("  %s %s: %s",
				greenDot, modelLabel, valueStyle.Render(modelVal)))
		} else {
			lines = append(lines, fmt.Sprintf("  %s %s: %s",
				grayDot, modelLabel, mutedStyle.Render(a.Catalog.Msg("未设置", "not set"))))
		}

		// Reasoning level (Codex only).
		if target.Name == "codex" {
			reasoningVal := a.Installer.ReadCodexConfigField("model_reasoning_effort")
			reasoningLabel := a.Catalog.Msg("思考等级", "Thinking Level")
			if reasoningVal != "" {
				lines = append(lines, fmt.Sprintf("  %s %s: %s",
					greenDot, reasoningLabel, valueStyle.Render(reasoningVal)))
			} else {
				lines = append(lines, fmt.Sprintf("  %s %s: %s",
					grayDot, reasoningLabel, mutedStyle.Render(a.Catalog.Msg("未设置", "not set"))))
			}
		}
	}

	// ── Runtime environment ──
	lines = append(lines, "")
	lines = append(lines, a.Installer.RuntimeSummaryLines()...)

	if len(status.Notes) > 0 {
		lines = append(lines, "")
		lines = append(lines, status.Notes...)
	}

	return ui.Panel{
		Title: fmt.Sprintf(a.Catalog.Msg("%s 安装信息", "%s install details"), status.Target.DisplayName),
		Lines: lines,
	}
}

// cliInstalled returns true if the given CLI target is installed.
func (a *App) cliInstalled(targetName string) bool {
	status, err := a.Installer.DetectTargetStatus(targetName)
	if err != nil {
		return false
	}
	return status.CLIInstalled
}

// cliDetailPanel returns a rich detail panel for a CLI, showing installation status,
// configuration, installed MCPs, installed skills, and version information.
func (a *App) cliDetailPanel(targetName string) ui.Panel {
	status, err := a.Installer.DetectTargetStatus(targetName)
	if err != nil {
		return errorPanel(a.Catalog.Msg("CLI 详情", "CLI details"), err)
	}

	greenDot := lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("●")
	grayDot := lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render("○")
	labelStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	blueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39"))

	lines := make([]string, 0, 40)

	// ── CLI Status ──
	if status.CLIInstalled {
		location := status.CLIPath
		if strings.TrimSpace(status.CLIPathScope) != "" {
			location = fmt.Sprintf("%s (%s)", status.CLIPath, status.CLIPathScope)
		}
		lines = append(lines, fmt.Sprintf("%s %s: %s", greenDot,
			labelStyle.Render(a.Catalog.Msg("CLI 状态", "CLI status")),
			valueStyle.Render(a.Catalog.Msg("已安装", "installed"))))
		lines = append(lines, fmt.Sprintf("  %s", mutedStyle.Render(location)))
	} else {
		lines = append(lines, fmt.Sprintf("%s %s: %s", grayDot,
			labelStyle.Render(a.Catalog.Msg("CLI 状态", "CLI status")),
			mutedStyle.Render(a.Catalog.Msg("未安装", "not installed"))))
	}

	// ── AgentFlow Status ──
	if status.AgentFlowInstalled {
		lines = append(lines, fmt.Sprintf("%s %s: %s", greenDot,
			labelStyle.Render("AgentFlow"),
			valueStyle.Render(a.Catalog.Msg("已安装", "installed"))))
	} else {
		lines = append(lines, fmt.Sprintf("%s %s: %s", grayDot,
			labelStyle.Render("AgentFlow"),
			mutedStyle.Render(a.Catalog.Msg("未安装", "not installed"))))
	}

	// ── Configuration (only when CLI installed) ──
	if status.CLIInstalled {
		target := status.Target
		lines = append(lines, "")
		lines = append(lines, labelStyle.Render(a.Catalog.Msg("─── 配置状态 ───", "─── Configuration ───")))

		if target.APIKeyEnv != "" {
			envVal := a.Installer.GetEnvOrRC(target.APIKeyEnv)
			if envVal == "" && target.Name == "codex" {
				envVal = a.Installer.ReadCodexAuthKey()
			}
			if envVal != "" {
				displayVal := envVal
				if len(envVal) > 6 {
					displayVal = envVal[:3] + strings.Repeat("*", len(envVal)-6) + envVal[len(envVal)-3:]
				}
				lines = append(lines, fmt.Sprintf("  %s API Key: %s", greenDot, valueStyle.Render(displayVal)))
			} else {
				lines = append(lines, fmt.Sprintf("  %s API Key: %s", grayDot, mutedStyle.Render(a.Catalog.Msg("未设置", "not set"))))
			}
		}

		if target.BaseURLEnv != "" {
			envVal := a.Installer.GetEnvOrRC(target.BaseURLEnv)
			if envVal == "" && target.Name == "codex" {
				envVal = a.Installer.ReadCodexConfigField("base_url")
			}
			if envVal != "" {
				lines = append(lines, fmt.Sprintf("  %s Base URL: %s", greenDot, valueStyle.Render(strings.TrimRight(envVal, "/"))))
			} else {
				lines = append(lines, fmt.Sprintf("  %s Base URL: %s", grayDot, mutedStyle.Render(a.Catalog.Msg("未设置", "not set"))))
			}
		}

		modelVal := a.Installer.ReadCLIConfigModel(target.Name)
		if modelVal != "" {
			lines = append(lines, fmt.Sprintf("  %s %s: %s", greenDot,
				a.Catalog.Msg("模型", "Model"), valueStyle.Render(modelVal)))
		} else {
			lines = append(lines, fmt.Sprintf("  %s %s: %s", grayDot,
				a.Catalog.Msg("模型", "Model"), mutedStyle.Render(a.Catalog.Msg("未设置", "not set"))))
		}

		// Reasoning level (Codex only).
		if target.Name == "codex" {
			reasoningVal := a.Installer.ReadCodexConfigField("model_reasoning_effort")
			reasoningLabel := a.Catalog.Msg("思考等级", "Thinking Level")
			if reasoningVal != "" {
				lines = append(lines, fmt.Sprintf("  %s %s: %s",
					greenDot, reasoningLabel, valueStyle.Render(reasoningVal)))
			} else {
				lines = append(lines, fmt.Sprintf("  %s %s: %s",
					grayDot, reasoningLabel, mutedStyle.Render(a.Catalog.Msg("未设置", "not set"))))
			}
		}
	}

	// ── Installed MCP & Skills (only when CLI is installed) ──
	if status.CLIInstalled {
		mcpList := a.mcpListPanel(targetName)
		if len(mcpList.Lines) > 0 {
			lines = append(lines, "")
			lines = append(lines, labelStyle.Render(a.Catalog.Msg("─── 已安装 MCP ───", "─── Installed MCP ───")))
			for _, line := range mcpList.Lines {
				trimmed := strings.TrimSpace(line)
				if trimmed == "" || strings.HasPrefix(trimmed, "─") || strings.HasPrefix(trimmed, "=") {
					continue
				}
				lines = append(lines, fmt.Sprintf("  %s %s", valueStyle.Render("✔"), trimmed))
			}
		} else {
			lines = append(lines, "")
			lines = append(lines, labelStyle.Render(a.Catalog.Msg("─── 已安装 MCP ───", "─── Installed MCP ───")))
			lines = append(lines, fmt.Sprintf("  %s %s", grayDot, mutedStyle.Render(a.Catalog.Msg("暂无", "none"))))
		}

		skillList := a.skillListPanel(targetName)
		if len(skillList.Lines) > 0 {
			lines = append(lines, "")
			lines = append(lines, labelStyle.Render(a.Catalog.Msg("─── 已安装 Skill ───", "─── Installed Skills ───")))
			for _, line := range skillList.Lines {
				trimmed := strings.TrimSpace(line)
				if trimmed == "" || strings.HasPrefix(trimmed, "─") || strings.HasPrefix(trimmed, "=") {
					continue
				}
				lines = append(lines, fmt.Sprintf("  %s %s", blueStyle.Render("✔"), trimmed))
			}
		} else {
			lines = append(lines, "")
			lines = append(lines, labelStyle.Render(a.Catalog.Msg("─── 已安装 Skill ───", "─── Installed Skills ───")))
			lines = append(lines, fmt.Sprintf("  %s %s", grayDot, mutedStyle.Render(a.Catalog.Msg("暂无", "none"))))
		}
	} else {
		// CLI not installed: show a brief tool introduction.
		lines = append(lines, "")
		target := status.Target
		switch target.Name {
		case "codex":
			lines = append(lines, mutedStyle.Render(a.Catalog.Msg(
				"Codex CLI 是 OpenAI 推出的终端 AI 编码助手，支持代码生成、调试和重构。",
				"Codex CLI is OpenAI's terminal AI coding assistant for code generation, debugging, and refactoring.")))
		case "claude":
			lines = append(lines, mutedStyle.Render(a.Catalog.Msg(
				"Claude Code 是 Anthropic 推出的终端 AI 编码助手，支持代码理解、生成和项目导航。",
				"Claude Code is Anthropic's terminal AI coding assistant for code understanding, generation, and project navigation.")))
		default:
			lines = append(lines, mutedStyle.Render(a.Catalog.Msg(
				"该 CLI 尚未安装，可通过 Enter 进入安装。",
				"This CLI is not installed. Press Enter to install.")))
		}
		if target.NPMPackage != "" {
			lines = append(lines, mutedStyle.Render(fmt.Sprintf("  npm: %s", target.NPMPackage)))
		}
		if target.DocsURL != "" {
			lines = append(lines, mutedStyle.Render(fmt.Sprintf("  %s: %s",
				a.Catalog.Msg("文档", "Docs"), target.DocsURL)))
		}
	}

	// ── Version ──
	lines = append(lines, "")
	lines = append(lines, a.Installer.RuntimeSummaryLines()...)

	return ui.Panel{
		Title: fmt.Sprintf(a.Catalog.Msg("%s 详情", "%s details"), status.Target.DisplayName),
		Lines: lines,
	}
}

func (a *App) bootstrapAutoPanel(targetName string) ui.Panel {
	lines, err := a.Installer.BootstrapCLI(targetName)
	if err != nil {
		return ui.Panel{
			Title: "❌ " + a.Catalog.Msg("CLI 安装失败", "CLI install failed"),
			Lines: []string{err.Error()},
		}
	}
	// Highlight all output lines in green for success.
	greenStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	highlighted := make([]string, len(lines))
	for i, line := range lines {
		highlighted[i] = greenStyle.Render(line)
	}
	return ui.Panel{
		Title: "✅ " + a.Catalog.Msg("CLI 安装成功", "CLI installed successfully"),
		Lines: highlighted,
	}
}

func (a *App) bootstrapManualPanel(targetName string) ui.Panel {
	lines, err := a.Installer.ManualInstallLines(targetName)
	if err != nil {
		return errorPanel(a.Catalog.Msg("手动安装提示", "Manual install guidance"), err)
	}
	return ui.Panel{
		Title: a.Catalog.Msg("手动安装提示", "Manual install guidance"),
		Lines: lines,
	}
}
