package app

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kittors/AgentFlow/internal/buildinfo"
	"github.com/kittors/AgentFlow/internal/i18n"
	"github.com/kittors/AgentFlow/internal/install"
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

	var notice *ui.Panel
	for {
		action, canceled, err := ui.RunMainMenu(a.Catalog, a.Version, a.mainMenuPanels(notice), a.Stdout)
		if err != nil {
			fmt.Fprintln(a.Stderr, err.Error())
			return 1
		}
		if canceled || action == ui.ActionExit {
			return 0
		}

		notice = nil
		switch action {
		case ui.ActionInstall:
			panel, _ := a.runInteractiveInstallPanel()
			notice = nonEmptyPanel(panel)
		case ui.ActionUninstall:
			panel, _ := a.runInteractiveUninstallPanel()
			notice = nonEmptyPanel(panel)
		case ui.ActionUpdate:
			panel := a.updatePanel(nil)
			notice = nonEmptyPanel(panel)
		case ui.ActionStatus:
			panel := a.statusPanel()
			panel.Title = a.Catalog.Msg("当前状态", "Current status")
			notice = nonEmptyPanel(panel)
		case ui.ActionClean:
			panel := a.cleanPanel()
			notice = nonEmptyPanel(panel)
		default:
			return 0
		}
	}
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

	for _, arg := range args {
		switch {
		case strings.HasPrefix(arg, "--profile="):
			profile = strings.TrimPrefix(arg, "--profile=")
		case arg == "--all":
			installAll = true
		case strings.HasPrefix(arg, "--"):
			fmt.Fprintln(a.Stderr, a.Catalog.Msg("未知参数。", "Unknown flag."))
			return 1
		default:
			targetName = arg
		}
	}

	if installAll {
		if _, err := a.Installer.InstallAll(profile); err != nil {
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

	if err := a.Installer.Install(targetName, profile); err != nil {
		fmt.Fprintln(a.Stderr, err.Error())
		return 1
	}
	fmt.Fprintln(a.Stdout, a.Catalog.Msg("安装完成。", "Install complete."))
	return 0
}

func (a *App) runUninstall(args []string) int {
	targetName := ""
	uninstallAll := false

	for _, arg := range args {
		switch {
		case arg == "--all":
			uninstallAll = true
		case strings.HasPrefix(arg, "--"):
			fmt.Fprintln(a.Stderr, a.Catalog.Msg("未知参数。", "Unknown flag."))
			return 1
		default:
			targetName = arg
		}
	}

	if uninstallAll {
		if _, err := a.Installer.UninstallAll(); err != nil {
			fmt.Fprintln(a.Stderr, err.Error())
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
	fmt.Fprintln(a.Stdout, a.Catalog.Msg("卸载完成。", "Uninstall complete."))
	return 0
}

func (a *App) runInteractiveInstall() int {
	panel, code := a.runInteractiveInstallPanel()
	a.printPanel(panel)
	return code
}

func (a *App) runInteractiveInstallPanel() (ui.Panel, int) {
	detected := a.Installer.DetectInstalledCLIs()
	if len(detected) == 0 {
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

	installed := sliceToSet(a.Installer.DetectInstalledTargets())
	options := make([]ui.Option, 0, len(detected))
	for _, name := range detected {
		description := a.Catalog.Msg("已检测到该 CLI，可直接部署 AgentFlow。", "Detected on this machine and ready for AgentFlow deployment.")
		if _, ok := installed[name]; ok {
			description = a.Catalog.Msg("该 CLI 已接入 AgentFlow，再次执行会覆盖为当前版本。", "AgentFlow is already installed for this CLI; rerunning refreshes it to the current version.")
		}
		options = append(options, ui.Option{
			Value:       name,
			Label:       name,
			Badge:       strings.ToUpper(name),
			Description: description,
		})
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

	success := 0
	lines := []string{
		fmt.Sprintf(a.Catalog.Msg("Profile: %s", "Profile: %s"), profile),
	}
	for _, name := range selected {
		if err := a.Installer.Install(name, profile); err != nil {
			lines = append(lines, fmt.Sprintf(a.Catalog.Msg("[失败] %s: %v", "[failed] %s: %v"), name, err))
			continue
		}
		success++
		lines = append(lines, fmt.Sprintf(a.Catalog.Msg("[完成] %s", "[done] %s"), name))
	}

	lines = append([]string{
		fmt.Sprintf(a.Catalog.Msg("已完成 %d/%d 个目标安装。", "Completed installation for %d/%d targets."), success, len(selected)),
	}, lines...)
	if success == 0 {
		return ui.Panel{
			Title: a.Catalog.Msg("安装失败", "Install failed"),
			Lines: lines,
		}, 1
	}
	return ui.Panel{
		Title: a.Catalog.Msg("安装结果", "Install result"),
		Lines: lines,
	}, 0
}

func (a *App) runInteractiveUninstall() int {
	panel, code := a.runInteractiveUninstallPanel()
	a.printPanel(panel)
	return code
}

func (a *App) runInteractiveUninstallPanel() (ui.Panel, int) {
	installed := a.Installer.DetectInstalledTargets()
	if len(installed) == 0 {
		return ui.Panel{
			Title: a.Catalog.Msg("卸载结果", "Uninstall result"),
			Lines: []string{a.Catalog.Msg("未检测到已安装的 AgentFlow。", "No AgentFlow installations found.")},
		}, 0
	}

	options := make([]ui.Option, 0, len(installed))
	for _, name := range installed {
		options = append(options, ui.Option{
			Value:       name,
			Label:       name,
			Badge:       strings.ToUpper(name),
			Description: a.Catalog.Msg("移除该 CLI 中由 AgentFlow 写入的规则、技能和 hooks。", "Remove the AgentFlow rules, skills, and hooks written into this CLI."),
		})
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

	success := 0
	lines := make([]string, 0, len(selected)+1)
	for _, name := range selected {
		if err := a.Installer.Uninstall(name); err != nil {
			lines = append(lines, fmt.Sprintf(a.Catalog.Msg("[失败] %s: %v", "[failed] %s: %v"), name, err))
			continue
		}
		success++
		lines = append(lines, fmt.Sprintf(a.Catalog.Msg("[完成] %s", "[done] %s"), name))
	}

	lines = append([]string{
		fmt.Sprintf(a.Catalog.Msg("已完成 %d/%d 个目标卸载。", "Completed uninstall for %d/%d targets."), success, len(selected)),
	}, lines...)
	if success == 0 {
		return ui.Panel{
			Title: a.Catalog.Msg("卸载失败", "Uninstall failed"),
			Lines: lines,
		}, 1
	}
	return ui.Panel{
		Title: a.Catalog.Msg("卸载结果", "Uninstall result"),
		Lines: lines,
	}, 0
}

func (a *App) runUpdate(args []string) int {
	panel := a.updatePanel(args)
	a.printPanel(panel)
	if strings.Contains(strings.ToLower(panel.Title), "失败") || strings.Contains(strings.ToLower(panel.Title), "failed") {
		return 1
	}
	return 0
}

func (a *App) updatePanel(args []string) ui.Panel {
	branch := ""
	if len(args) > 0 {
		branch = args[0]
	}
	if branch != "" {
		return ui.Panel{
			Title: a.Catalog.Msg("更新失败", "Update failed"),
			Lines: []string{fmt.Sprintf(a.Catalog.Msg("当前 Go update 不支持分支参数: %s", "The Go update command does not support branch arguments: %s"), branch)},
		}
	}
	result, err := a.Checker.SelfUpdate(a.Version)
	if err != nil {
		return errorPanel(a.Catalog.Msg("更新失败", "Update failed"), err)
	}
	if !result.UpdateAvailable {
		return ui.Panel{
			Title: a.Catalog.Msg("更新结果", "Update result"),
			Lines: []string{a.Catalog.Msg("当前已是最新版本。", "Already on the latest version.")},
		}
	}
	return ui.Panel{
		Title: a.Catalog.Msg("更新结果", "Update result"),
		Lines: []string{
			fmt.Sprintf(a.Catalog.Msg("已更新到 v%s。", "Updated to v%s."), result.Latest),
			a.Catalog.Msg("请重新运行 agentflow，进入新版本。", "Restart agentflow to enter the new version."),
		},
	}
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
	fmt.Fprintln(a.Stdout, "  uninstall [target|--all]")
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
	lines := make([]string, 0, 10)
	if executable, err := os.Executable(); err == nil {
		lines = append(lines, fmt.Sprintf(a.Catalog.Msg("可执行文件: %s", "Executable: %s"), executable))
	}
	lines = append(lines, "")
	lines = append(lines, a.Installer.StatusLines()...)
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

func (a *App) printPanel(panel ui.Panel) {
	if strings.TrimSpace(panel.Title) != "" {
		fmt.Fprintln(a.Stdout, panel.Title)
	}
	for _, line := range panel.Lines {
		fmt.Fprintln(a.Stdout, line)
	}
}

func errorPanel(title string, err error) ui.Panel {
	return ui.Panel{
		Title: title,
		Lines: []string{err.Error()},
	}
}

func nonEmptyPanel(panel ui.Panel) *ui.Panel {
	if strings.TrimSpace(panel.Title) == "" && len(panel.Lines) == 0 {
		return nil
	}
	return &panel
}
