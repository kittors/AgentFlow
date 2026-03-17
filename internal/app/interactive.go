package app

import (
	"fmt"
	"strings"

	"github.com/kittors/AgentFlow/internal/i18n"
	"github.com/kittors/AgentFlow/internal/ui"
)

func (a *App) runInteractiveMainMenu() int {
	if !stdinIsTTY() {
		a.printUsage()
		return 0
	}
	if code, ok := a.ensureInteractiveLanguage(); !ok {
		return code
	}

	if err := ui.RunInteractiveFlow(a.Catalog, a.Version, ui.InteractiveCallbacks{
		Status:                  a.statusPanel,
		CLIDetailPanel:          a.cliDetailPanel,
		CLIInstalled:            a.cliInstalled,
		MCPTargetOptions:        a.mcpTargetOptions,
		MCPInstallOptions:       a.mcpInstallOptions,
		MCPRemoveOptions:        a.mcpRemoveOptions,
		MCPList:                 a.mcpListPanel,
		MCPInstall:              a.mcpInstallPanel,
		MCPInstallWithEnv:       a.mcpInstallWithEnvPanel,
		MCPBatchInstall:         a.mcpBatchInstallPanel,
		MCPConfigFields:         a.mcpConfigFields,
		MCPRemove:               a.mcpRemovePanel,
		MCPBatchRemove:          a.mcpBatchRemovePanel,
		SkillTargetOptions:      a.skillTargetOptions,
		SkillGlobalSupported:    a.skillGlobalSupported,
		SkillInstallOptions:     a.skillInstallOptions,
		SkillUninstallOptions:   a.skillUninstallOptions,
		SkillList:               a.skillListPanel,
		SkillInstall:            a.skillInstallPanel,
		SkillUninstall:          a.skillUninstallPanel,
		ProjectRulesPanel:       a.projectRulesPanel,
		ProjectRulesInstall:     a.projectRulesInstallPanel,
		ProjectRulesUninstall:   a.projectRulesUninstallPanel,
		BootstrapOptions:        a.bootstrapTargetOptions,
		BootstrapAutoSupported:  a.bootstrapAutoSupported,
		BootstrapDetails:        a.bootstrapTargetPanel,
		BootstrapAuto:           a.bootstrapAutoPanel,
		BootstrapManual:         a.bootstrapManualPanel,
		InstallOptions:          a.installTargetOptions,
		UninstallOptions:        a.uninstallTargetOptions,
		UninstallProjectOptions: a.uninstallProjectTargetOptions,
		UninstallCLIOptions:     a.uninstallCLITargetOptions,
		Install:                 a.installTargetsPanel,
		Uninstall:               a.uninstallTargetsPanel,
		UninstallCLI:            a.uninstallCLITargetsPanel,
		Update:                  a.updatePanel,
		Clean:                   a.cleanPanel,
		CLIConfigFields:         a.cliConfigFields,
		WriteEnvConfig:          a.writeEnvConfigPanel,
	}, a.Stdout); err != nil {
		fmt.Fprintln(a.Stderr, err.Error())
		return 1
	}
	return 0
}

func (a *App) ensureInteractiveLanguage() (int, bool) {
	if language, ok := i18n.LoadPreferredLocale(); ok {
		a.setCatalog(i18n.NewCatalogWithLanguage(language))
		return 0, true
	}

	language, canceled, err := ui.SelectLanguage(i18n.DetectLocaleFromEnvironment(), a.Stdout)
	if err != nil {
		fmt.Fprintln(a.Stderr, err.Error())
		return 1, false
	}
	if canceled {
		return 0, false
	}
	if err := i18n.SavePreferredLocale(language); err != nil {
		fmt.Fprintln(a.Stderr, err.Error())
		return 1, false
	}

	a.setCatalog(i18n.NewCatalogWithLanguage(language))
	return 0, true
}

func (a *App) setCatalog(catalog i18n.Catalog) {
	a.Catalog = catalog
	if a.Installer != nil {
		a.Installer.Catalog = catalog
	}
}

func (a *App) runInteractiveInstall() int {
	panel, code := a.runInteractiveInstallPanel()
	a.printPanel(panel)
	return code
}

func (a *App) runInteractiveInstallPanel() (ui.Panel, int) {
	options := a.installTargetOptions()
	if len(options) == 0 {
		return ui.Panel{
			Title: a.Catalog.Msg("安装结果", "Install result"),
			Lines: []string{a.Catalog.Msg("未检测到任何已安装的 CLI。", "No installed CLIs detected.")},
		}, 0
	}

	profile, canceled, err := ui.SelectProfile(a.Catalog, a.Stdout)
	if err != nil {
		return errorPanel(a.Catalog.Msg("安装失败", "Install failed"), err), 1
	}
	if canceled {
		return ui.Panel{}, 0
	}

	selected, canceled, err := ui.SelectTargets(
		a.Catalog,
		a.Stdout,
		a.Catalog.Msg("选择要安装的目标", "Select targets to install"),
		a.Catalog.Msg("Space 选择多个目标，Enter 开始安装。", "Use Space to select multiple targets, then Enter to install."),
		options,
	)
	if err != nil {
		return errorPanel(a.Catalog.Msg("安装失败", "Install failed"), err), 1
	}
	if canceled {
		return ui.Panel{}, 0
	}

	panel := a.installTargetsPanel(profile, selected)
	if strings.Contains(panel.Title, a.Catalog.Msg("失败", "failed")) || strings.Contains(strings.ToLower(panel.Title), "failed") {
		return panel, 1
	}
	return panel, 0
}

func (a *App) runInteractiveUninstall() int {
	panel, code := a.runInteractiveUninstallPanel()
	a.printPanel(panel)
	return code
}

func (a *App) runInteractiveUninstallPanel() (ui.Panel, int) {
	options := a.uninstallTargetOptions()
	if len(options) == 0 {
		return ui.Panel{
			Title: a.Catalog.Msg("卸载结果", "Uninstall result"),
			Lines: []string{a.Catalog.Msg("未检测到已安装的 AgentFlow。", "No AgentFlow installations found.")},
		}, 0
	}

	selected, canceled, err := ui.SelectTargets(
		a.Catalog,
		a.Stdout,
		a.Catalog.Msg("选择要卸载的目标", "Select targets to uninstall"),
		a.Catalog.Msg("Space 选择多个目标，Enter 开始卸载。", "Use Space to select multiple targets, then Enter to uninstall."),
		options,
	)
	if err != nil {
		return errorPanel(a.Catalog.Msg("卸载失败", "Uninstall failed"), err), 1
	}
	if canceled {
		return ui.Panel{}, 0
	}

	panel := a.uninstallTargetsPanel(selected)
	if strings.Contains(panel.Title, a.Catalog.Msg("失败", "failed")) || strings.Contains(strings.ToLower(panel.Title), "failed") {
		return panel, 1
	}
	return panel, 0
}
