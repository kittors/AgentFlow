package app

import (
	"fmt"
	"strings"

	"github.com/kittors/AgentFlow/internal/config"
	"github.com/kittors/AgentFlow/internal/debuglog"
	"github.com/kittors/AgentFlow/internal/targets"
	"github.com/kittors/AgentFlow/internal/ui"
)

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
