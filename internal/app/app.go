package app

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/kittors/AgentFlow/internal/buildinfo"
	"github.com/kittors/AgentFlow/internal/config"
	"github.com/kittors/AgentFlow/internal/debuglog"
	"github.com/kittors/AgentFlow/internal/i18n"
	"github.com/kittors/AgentFlow/internal/install"
	"github.com/kittors/AgentFlow/internal/projectrules"
	"github.com/kittors/AgentFlow/internal/targets"
	"github.com/kittors/AgentFlow/internal/ui"
	"github.com/kittors/AgentFlow/internal/update"
)

type App struct {
	Stdout    io.Writer
	Stderr    io.Writer
	Catalog   i18n.Catalog
	Installer *install.Installer
	Checker   *update.Checker
	Version   string
}

func New(stdout, stderr io.Writer) *App {
	catalog := i18n.NewCatalog()
	return &App{
		Stdout:    stdout,
		Stderr:    stderr,
		Catalog:   catalog,
		Installer: install.New(catalog, stdout),
		Checker:   update.NewChecker(),
		Version:   buildinfo.CurrentVersion(),
	}
}

func (a *App) Run(args []string) int {
	debuglog.Init()
	debuglog.Log("App.Run args=%v version=%s", args, a.Version)
	if len(args) == 0 {
		return a.runInteractiveMainMenu()
	}

	switch args[0] {
	case "help", "-h", "--help":
		a.printUsage()
		return 0
	case "--check-update":
		silent := len(args) > 1 && args[1] == "--silent"
		a.printVersionCheck(!silent, 24)
		return 0
	case "version":
		a.printVersionCheck(true, 72)
		return 0
	case "status":
		a.printStatus()
		return 0
	case "clean":
		if err := a.runClean(); err != nil {
			fmt.Fprintln(a.Stderr, err.Error())
		}
		return 0
	case "install":
		return a.runInstall(args[1:])
	case "uninstall":
		return a.runUninstall(args[1:])
	case "update":
		return a.runUpdate(args[1:])
	case "init":
		return a.runInit(args[1:])
	case "kb":
		return a.runKB(args[1:])
	case "session":
		return a.runSession(args[1:])
	case "conventions":
		return a.runConventions(args[1:])
	case "graph":
		return a.runGraph(args[1:])
	case "dashboard":
		return a.runDashboard(args[1:])
	case "rules":
		return a.runRules(args[1:])
	case "skill":
		return a.runSkill(args[1:])
	case "mcp":
		return a.runMCP(args[1:])
	default:
		fmt.Fprintln(a.Stderr, a.Catalog.Msg("未知命令。", "Unknown command."))
		a.printUsage()
		return 1
	}
}

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
		MCPConfigFields:         a.mcpConfigFields,
		MCPRemove:               a.mcpRemovePanel,
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

func (a *App) runInstall(args []string) int {
	profile := targets.DefaultProfile
	targetName := ""
	installAll := false
	lang := ""

	for _, arg := range args {
		switch {
		case strings.HasPrefix(arg, "--profile="):
			profile = strings.TrimPrefix(arg, "--profile=")
		case strings.HasPrefix(arg, "--lang="):
			lang = strings.TrimPrefix(arg, "--lang=")
		case arg == "--all":
			installAll = true
		case strings.HasPrefix(arg, "--"):
			fmt.Fprintln(a.Stderr, a.Catalog.Msg("未知参数。", "Unknown flag."))
			return 1
		default:
			targetName = arg
		}
	}
	if lang == "" {
		lang = config.DefaultLang
	}

	if installAll {
		if _, err := a.Installer.InstallAll(profile, lang); err != nil {
			fmt.Fprintln(a.Stderr, err.Error())
			return 1
		}
		fmt.Fprintln(a.Stdout, a.Catalog.Msg("已完成全部目标安装。", "Installed all detected targets."))
		return 0
	}

	if targetName == "" {
		if stdinIsTTY() {
			return a.runInteractiveInstall()
		}
		fmt.Fprintln(a.Stderr, a.Catalog.Msg("缺少 target；请指定目标、使用 --all，或在交互式终端中直接运行。", "missing target; specify one, use --all, or run in an interactive terminal."))
		return 1
	}

	if err := a.Installer.Install(targetName, profile, lang); err != nil {
		fmt.Fprintln(a.Stderr, err.Error())
		return 1
	}
	fmt.Fprintln(a.Stdout, a.Catalog.Msg("安装完成。", "Install complete."))
	return 0
}

func (a *App) runUninstall(args []string) int {
	targetName := ""
	uninstallAll := false
	removeCLI := false
	purgeConfig := false
	keepConfig := false

	for _, arg := range args {
		switch {
		case arg == "--all":
			uninstallAll = true
		case arg == "--cli":
			removeCLI = true
		case arg == "--purge-config":
			purgeConfig = true
		case arg == "--keep-config":
			keepConfig = true
		case strings.HasPrefix(arg, "--"):
			fmt.Fprintln(a.Stderr, a.Catalog.Msg("未知参数。", "Unknown flag."))
			return 1
		default:
			targetName = arg
		}
	}

	if purgeConfig && keepConfig {
		fmt.Fprintln(a.Stderr, a.Catalog.Msg("--purge-config 与 --keep-config 不能同时使用。", "--purge-config and --keep-config are mutually exclusive."))
		return 1
	}
	if purgeConfig && !removeCLI {
		fmt.Fprintln(a.Stderr, a.Catalog.Msg("--purge-config 需要与 --cli 一起使用。", "--purge-config requires --cli."))
		return 1
	}
	// Full uninstall defaults to purging the config directory unless explicitly kept.
	if removeCLI && !keepConfig {
		purgeConfig = true
	}

	if uninstallAll {
		targetsToProcess := a.Installer.DetectInstalledTargets()
		if removeCLI {
			targetSet := make(map[string]bool, len(targetsToProcess))
			for _, name := range targetsToProcess {
				targetSet[name] = true
			}
			for _, name := range a.Installer.DetectInstalledCLIs() {
				targetSet[name] = true
			}
			targetsToProcess = targetsToProcess[:0]
			for name := range targetSet {
				targetsToProcess = append(targetsToProcess, name)
			}
		}

		success := 0
		for _, name := range targetsToProcess {
			if err := a.Installer.Uninstall(name); err != nil {
				fmt.Fprintln(a.Stderr, err.Error())
				continue
			}
			if removeCLI {
				if _, err := a.Installer.UninstallCLI(name); err != nil {
					fmt.Fprintln(a.Stderr, err.Error())
					continue
				}
				if purgeConfig {
					if err := a.Installer.PurgeConfigDir(name); err != nil {
						fmt.Fprintln(a.Stderr, err.Error())
						continue
					}
				}
			}
			success++
		}
		if success == 0 {
			fmt.Fprintln(a.Stderr, a.Catalog.Msg("未卸载任何目标。", "No targets were uninstalled."))
			return 1
		}
		return 0
	}

	if targetName == "" {
		if stdinIsTTY() {
			return a.runInteractiveUninstall()
		}
		fmt.Fprintln(a.Stderr, a.Catalog.Msg("缺少 target；请指定目标、使用 --all，或在交互式终端中直接运行。", "missing target; specify one, use --all, or run in an interactive terminal."))
		return 1
	}

	if err := a.Installer.Uninstall(targetName); err != nil {
		fmt.Fprintln(a.Stderr, err.Error())
		return 1
	}
	if removeCLI {
		if _, err := a.Installer.UninstallCLI(targetName); err != nil {
			fmt.Fprintln(a.Stderr, err.Error())
			return 1
		}
		if purgeConfig {
			if err := a.Installer.PurgeConfigDir(targetName); err != nil {
				fmt.Fprintln(a.Stderr, err.Error())
				return 1
			}
		}
	}
	fmt.Fprintln(a.Stdout, a.Catalog.Msg("卸载完成。", "Uninstall complete."))
	return 0
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

func (a *App) runUpdate(args []string) int {
	if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
		panel := ui.Panel{
			Title: a.Catalog.Msg("更新失败", "Update failed"),
			Lines: []string{
				fmt.Sprintf(a.Catalog.Msg("当前 Go update 不支持分支参数: %s", "The Go update command does not support branch arguments: %s"), args[0]),
			},
		}
		a.printPanel(panel)
		return 1
	}
	panel, _ := a.updatePanel(func(string, int, string) {})
	a.printPanel(panel)
	if strings.Contains(strings.ToLower(panel.Title), "失败") || strings.Contains(strings.ToLower(panel.Title), "failed") {
		return 1
	}
	return 0
}

func (a *App) updatePanel(progress func(stage string, percent int, info string)) (ui.Panel, string) {
	done := debuglog.Timed("updatePanel")
	defer done()
	result, err := a.Checker.SelfUpdateWithProgress(a.Version, func(stage string, percent int, info string) {
		if progress != nil {
			progress(stage, percent, info)
		}
	})
	if err != nil {
		return errorPanel(a.Catalog.Msg("更新失败", "Update failed"), err), ""
	}
	if !result.UpdateAvailable {
		return ui.Panel{
			Title: a.Catalog.Msg("更新结果", "Update result"),
			Lines: []string{a.Catalog.Msg("当前已是最新版本。", "Already on the latest version.")},
		}, ""
	}
	a.Version = result.Latest
	return ui.Panel{
		Title: a.Catalog.Msg("更新结果", "Update result"),
		Lines: []string{
			fmt.Sprintf(a.Catalog.Msg("已更新到 v%s。", "Updated to v%s."), result.Latest),
			a.Catalog.Msg("请重新运行 agentflow，进入新版本。", "Restart agentflow to enter the new version."),
		},
	}, result.Latest
}

func (a *App) runClean() error {
	cleaned, err := a.Installer.Clean()
	if err != nil {
		return err
	}
	fmt.Fprintf(a.Stdout, a.Catalog.Msg("已清理 %d 个缓存目录。\n", "Cleaned %d cache directories.\n"), cleaned)
	return nil
}

func (a *App) printStatus() {
	fmt.Fprintf(a.Stdout, "AgentFlow v%s\n", a.Version)
	for _, line := range a.statusPanel().Lines {
		fmt.Fprintln(a.Stdout, line)
	}
}

func (a *App) printUsage() {
	fmt.Fprintln(a.Stdout, "Usage: agentflow [command]")
	fmt.Fprintln(a.Stdout, "")
	fmt.Fprintln(a.Stdout, "Commands:")
	fmt.Fprintln(a.Stdout, "  install [target|--all] [--profile=<lite|standard|full>]")
	fmt.Fprintln(a.Stdout, "  uninstall [target|--all] [--cli] [--keep-config] [--purge-config]")
	fmt.Fprintln(a.Stdout, "  update [branch]")
	fmt.Fprintln(a.Stdout, "  status")
	fmt.Fprintln(a.Stdout, "  clean")
	fmt.Fprintln(a.Stdout, "  version")
	fmt.Fprintln(a.Stdout, "  help")
	fmt.Fprintln(a.Stdout, "  init [--root=<path>] [--quiet]")
	fmt.Fprintln(a.Stdout, "  kb sync [--root=<path>] [--quiet]")
	fmt.Fprintln(a.Stdout, "  session save [--root=<path>] [--stage=<name>] [--quiet]")
	fmt.Fprintln(a.Stdout, "  conventions [--root=<path>] [--quiet]")
	fmt.Fprintln(a.Stdout, "  graph [--root=<path>] [--quiet]")
	fmt.Fprintln(a.Stdout, "  dashboard [--root=<path>] [--quiet]")
	fmt.Fprintln(a.Stdout, "  rules <detect|install> ...")
	fmt.Fprintln(a.Stdout, "  skill <install|uninstall|list> ...")
	fmt.Fprintln(a.Stdout, "  mcp <install|remove|list|search> ...")
	fmt.Fprintln(a.Stdout, "  --check-update [--silent]")
}

func sliceToSet(values []string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		result[value] = struct{}{}
	}
	return result
}

func stdinIsTTY() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

func (a *App) printVersionCheck(showVersion bool, ttl int) {
	if showVersion {
		fmt.Fprintf(a.Stdout, "AgentFlow v%s\n", a.Version)
	}
	result, err := a.Checker.Check(a.Version, update.Options{CacheTTLHours: ttl})
	if err != nil {
		return
	}
	if result.UpdateAvailable {
		fmt.Fprintf(a.Stdout, a.Catalog.Msg("  ⬆️ 新版本可用: v%s (当前 v%s)\n", "  ⬆️ Update available: v%s (current v%s)\n"), result.Latest, result.Current)
		fmt.Fprintln(a.Stdout, a.Catalog.Msg("     运行 agentflow update 更新", "     Run: agentflow update"))
	}
}

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

	// ── Installed MCP ──
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

	// ── Installed Skills ──
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
			// For Codex, API key goes to auth.json (not shell rc).
			codexAPIKey = value
		case "__CODEX_BASE_URL__":
			// Codex CLI v0.115+ uses openai_base_url in config.toml; no longer an env var.
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
			os.Setenv("OPENAI_API_KEY", codexAPIKey)
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

func (a *App) uninstallCLITargetsPanel(targets []string) ui.Panel {
	success := 0
	lines := make([]string, 0, len(targets)+1)
	for _, name := range targets {
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
		success++
		lines = append(lines, fmt.Sprintf(a.Catalog.Msg("[完成] %s", "[done] %s"), name))
	}
	lines = append([]string{
		fmt.Sprintf(a.Catalog.Msg("已完成 %d/%d 个 CLI 卸载。", "Completed CLI uninstall for %d/%d targets."), success, len(targets)),
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

func (a *App) printPanel(panel ui.Panel) {
	if strings.TrimSpace(panel.Title) != "" {
		fmt.Fprintln(a.Stdout, panel.Title)
	}
	for _, line := range panel.Lines {
		fmt.Fprintln(a.Stdout, line)
	}
}

func errorPanel(title string, err error) ui.Panel {
	lines := strings.Split(strings.ReplaceAll(err.Error(), "\r\n", "\n"), "\n")
	if len(lines) == 0 {
		lines = []string{err.Error()}
	}
	return ui.Panel{
		Title: title,
		Lines: lines,
	}
}

func nonEmptyPanel(panel ui.Panel) *ui.Panel {
	if strings.TrimSpace(panel.Title) == "" && len(panel.Lines) == 0 {
		return nil
	}
	return &panel
}
