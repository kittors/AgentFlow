package app

import (
	"fmt"
	"os"
	"strings"

	"github.com/kittors/AgentFlow/internal/config"
	"github.com/kittors/AgentFlow/internal/mcp"
	"github.com/kittors/AgentFlow/internal/projectrules"
	"github.com/kittors/AgentFlow/internal/targets"
	"github.com/kittors/AgentFlow/internal/ui"
)

func (a *App) uninstallTargetOptions() []ui.Option {
	installed := a.Installer.DetectInstalledTargets()
	options := make([]ui.Option, 0, len(installed))
	for _, name := range installed {
		options = append(options, ui.Option{
			Value:       name,
			Label:       name,
			Badge:       strings.ToUpper(name),
			Description: a.Catalog.Msg("移除该 CLI 中由 AgentFlow 写入的规则、技能和 hooks。", "Remove the AgentFlow rules, skills, and hooks written into this CLI."),
		})
	}
	return options
}

func (a *App) uninstallProjectTargetOptions() []ui.Option {
	wd, err := os.Getwd()
	if err != nil {
		return nil
	}
	manager := projectrules.NewManager()
	statuses, err := manager.Detect(wd)
	if err != nil {
		return nil
	}
	seen := make(map[string]bool)
	var options []ui.Option
	for _, status := range statuses {
		if !status.Exists || !status.Managed {
			continue
		}
		if seen[status.Target] {
			continue
		}
		seen[status.Target] = true
		options = append(options, ui.Option{
			Value:       "project:" + status.Target,
			Label:       fmt.Sprintf(a.Catalog.Msg("项目级 %s", "Project %s"), status.Target),
			Badge:       a.Catalog.Msg("项目", "PROJECT"),
			Description: fmt.Sprintf(a.Catalog.Msg("删除当前项目目录中 %s 的 AgentFlow 规则文件（%s）。", "Remove AgentFlow rule files for %s from project directory (%s)."), status.Target, status.Detected),
		})
	}
	return options
}

func (a *App) uninstallCLITargetOptions() []ui.Option {
	installed := a.Installer.DetectInstalledCLIs()
	options := make([]ui.Option, 0, len(installed))
	for _, name := range installed {
		options = append(options, ui.Option{
			Value:       name,
			Label:       name,
			Badge:       strings.ToUpper(name),
			Description: a.Catalog.Msg("卸载该 CLI 本体，并默认删除配置目录（完整卸载）。", "Uninstall the CLI tool and purge its config directory by default (full uninstall)."),
		})
	}
	return options
}

func (a *App) cliConfigFields(target string) []ui.ConfigField {
	fields := a.Installer.CLIConfigFields(target)
	if len(fields) == 0 {
		return nil
	}
	result := make([]ui.ConfigField, len(fields))
	for i, f := range fields {
		result[i] = ui.ConfigField{
			Label:   f.Label,
			EnvVar:  f.EnvVar,
			Type:    f.Type,
			Options: f.Options,
			Default: f.Default,
		}
	}
	return result
}

func (a *App) writeEnvConfigPanel(envVars map[string]string) ui.Panel {
	// Separate normal env vars from special config-file fields.
	normalEnvVars := make(map[string]string)
	var codexAPIKey, codexBaseURL, codexModel, codexReasoning string
	var claudeModel string
	var modelEnvVar, modelValue string

	for key, value := range envVars {
		switch key {
		case "__CODEX_MODEL__":
			codexModel = value
		case "__CODEX_REASONING__":
			codexReasoning = value
		case "__CLAUDE_MODEL__":
			claudeModel = value
		case "__MODEL__": // fallback
			codexModel = value
		case "OPENAI_API_KEY":
			// API key goes to auth.json only. Codex reads it from there
			// via model_provider (no env_key, no env var needed).
			codexAPIKey = value
		case "__CODEX_BASE_URL__":
			// Base URL goes to [model_providers.agentflow].base_url in config.toml.
			codexBaseURL = value
		default:
			normalEnvVars[key] = value
			// Track model env var for other targets.
			if key == "GEMINI_MODEL" || key == "DASHSCOPE_MODEL" {
				modelEnvVar = key
				modelValue = value
			}
		}
	}

	var allLines []string

	// Write normal env vars to shell rc (excludes Codex API Key / Base URL).
	if len(normalEnvVars) > 0 {
		lines, err := a.Installer.WriteEnvConfig(normalEnvVars)
		if err != nil {
			return errorPanel(a.Catalog.Msg("配置写入失败", "Config write failed"), err)
		}
		allLines = append(allLines, lines...)
		// Also set in current process so changes take effect immediately.
		for key, value := range normalEnvVars {
			os.Setenv(key, value)
		}
	}

	// Write Codex config (auth.json + config.toml).
	if codexAPIKey != "" || codexBaseURL != "" || codexModel != "" || codexReasoning != "" {
		if err := a.Installer.WriteCodexConfig(codexAPIKey, codexBaseURL, codexModel, codexReasoning); err != nil {
			return errorPanel(a.Catalog.Msg("Codex 配置写入失败", "Codex config write failed"), err)
		}
		allLines = append(allLines, "")
		if codexAPIKey != "" {
			masked := codexAPIKey[:3] + strings.Repeat("*", len(codexAPIKey)-6) + codexAPIKey[len(codexAPIKey)-3:]
			allLines = append(allLines, fmt.Sprintf(a.Catalog.Msg("已写入 ~/.codex/auth.json (API Key: %s)", "Written to ~/.codex/auth.json (API Key: %s)"), masked))
		}
		allLines = append(allLines, a.Catalog.Msg("已写入 ~/.codex/config.toml:", "Written to ~/.codex/config.toml:"))
		if codexModel != "" {
			allLines = append(allLines, fmt.Sprintf("  model: %s", codexModel))
		}
		if codexReasoning != "" {
			allLines = append(allLines, fmt.Sprintf("  model_reasoning_effort: %s", codexReasoning))
		}
		if codexBaseURL != "" {
			allLines = append(allLines, fmt.Sprintf("  model_provider: agentflow (base_url: %s)", codexBaseURL))
		}
	}

	// Write Claude settings.json if applicable.
	if claudeModel != "" {
		if err := a.Installer.WriteClaudeConfig(claudeModel); err != nil {
			return errorPanel(a.Catalog.Msg("Claude 配置写入失败", "Claude config write failed"), err)
		}
		allLines = append(allLines, "")
		allLines = append(allLines, a.Catalog.Msg("已写入 ~/.claude.json:", "Written to ~/.claude.json:"))
		allLines = append(allLines, fmt.Sprintf("  model: %s", claudeModel))
	}

	// Report model env var if written.
	if modelEnvVar != "" && modelValue != "" {
		allLines = append(allLines, "")
		allLines = append(allLines, fmt.Sprintf(a.Catalog.Msg("默认模型已设置: %s=%s", "Default model set: %s=%s"), modelEnvVar, modelValue))
	}

	if len(allLines) == 0 {
		allLines = []string{a.Catalog.Msg("未写入任何配置（所有字段留空）。", "No configuration written (all fields left empty).")}
	}

	return ui.Panel{
		Title: a.Catalog.Msg("配置写入成功", "Configuration saved"),
		Lines: allLines,
	}
}

func (a *App) installTargetsPanel(profile string, targets []string) ui.Panel {
	success := 0
	lines := []string{
		fmt.Sprintf(a.Catalog.Msg("Profile: %s", "Profile: %s"), profile),
	}
	for _, name := range targets {
		if err := a.Installer.Install(name, profile, config.DefaultLang); err != nil {
			lines = append(lines, fmt.Sprintf(a.Catalog.Msg("[失败] %s: %v", "[failed] %s: %v"), name, err))
			continue
		}
		success++
		lines = append(lines, fmt.Sprintf(a.Catalog.Msg("[完成] %s", "[done] %s"), name))
	}
	lines = append([]string{
		fmt.Sprintf(a.Catalog.Msg("已完成 %d/%d 个目标安装。", "Completed installation for %d/%d targets."), success, len(targets)),
	}, lines...)
	if success == 0 {
		return ui.Panel{
			Title: a.Catalog.Msg("安装失败", "Install failed"),
			Lines: lines,
		}
	}
	return ui.Panel{
		Title: a.Catalog.Msg("安装结果", "Install result"),
		Lines: lines,
	}
}

func (a *App) uninstallTargetsPanel(targets []string) ui.Panel {
	success := 0
	lines := make([]string, 0, len(targets)+1)
	for _, name := range targets {
		if err := a.Installer.Uninstall(name); err != nil {
			lines = append(lines, fmt.Sprintf(a.Catalog.Msg("[失败] %s: %v", "[failed] %s: %v"), name, err))
			continue
		}
		success++
		lines = append(lines, fmt.Sprintf(a.Catalog.Msg("[完成] %s", "[done] %s"), name))
	}
	lines = append([]string{
		fmt.Sprintf(a.Catalog.Msg("已完成 %d/%d 个目标卸载。", "Completed uninstall for %d/%d targets."), success, len(targets)),
	}, lines...)
	if success == 0 {
		return ui.Panel{
			Title: a.Catalog.Msg("卸载失败", "Uninstall failed"),
			Lines: lines,
		}
	}
	return ui.Panel{
		Title: a.Catalog.Msg("卸载结果", "Uninstall result"),
		Lines: lines,
	}
}

func (a *App) uninstallCLITargetsPanel(targetNames []string) ui.Panel {
	success := 0
	lines := make([]string, 0, len(targetNames)+1)
	for _, name := range targetNames {
		// Clean up MCP servers BEFORE removing the config dir,
		// because MCP configs may live outside the config dir
		// (e.g. ~/.claude.json for Claude Code).
		if mcpTarget, ok := targets.LookupMCP(name); ok {
			mcpMgr := mcp.NewManager()
			if mcpServers, err := mcpMgr.List(mcpTarget); err == nil {
				for _, srv := range mcpServers {
					_ = mcpMgr.Remove(mcpTarget, srv)
				}
				if len(mcpServers) > 0 {
					lines = append(lines, fmt.Sprintf(a.Catalog.Msg("已清理 %d 个 MCP 服务器配置。", "Cleaned up %d MCP server configs."), len(mcpServers)))
				}
			}
		}

		if err := a.Installer.Uninstall(name); err != nil {
			lines = append(lines, fmt.Sprintf(a.Catalog.Msg("[失败] %s: %v", "[failed] %s: %v"), name, err))
			continue
		}
		if _, err := a.Installer.UninstallCLI(name); err != nil {
			lines = append(lines, fmt.Sprintf(a.Catalog.Msg("[失败] %s: %v", "[failed] %s: %v"), name, err))
			continue
		}
		if err := a.Installer.PurgeConfigDir(name); err != nil {
			lines = append(lines, fmt.Sprintf(a.Catalog.Msg("[失败] %s: %v", "[failed] %s: %v"), name, err))
			continue
		}

		// Clear process-level env vars so that re-installing in the same
		// session doesn't show stale API keys / Base URLs.
		if target, ok := targets.Lookup(name); ok {
			if target.APIKeyEnv != "" && !strings.HasPrefix(target.APIKeyEnv, "__") {
				os.Unsetenv(target.APIKeyEnv)
			}
			if target.BaseURLEnv != "" && !strings.HasPrefix(target.BaseURLEnv, "__") {
				os.Unsetenv(target.BaseURLEnv)
			}
		}

		success++
		lines = append(lines, fmt.Sprintf(a.Catalog.Msg("[完成] %s", "[done] %s"), name))
	}
	lines = append([]string{
		fmt.Sprintf(a.Catalog.Msg("已完成 %d/%d 个 CLI 卸载。", "Completed CLI uninstall for %d/%d targets."), success, len(targetNames)),
	}, lines...)
	if success == 0 {
		return ui.Panel{
			Title: a.Catalog.Msg("卸载失败", "Uninstall failed"),
			Lines: lines,
		}
	}
	return ui.Panel{
		Title: a.Catalog.Msg("卸载结果", "Uninstall result"),
		Lines: lines,
	}
}
