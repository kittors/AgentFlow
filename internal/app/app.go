package app

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kittors/AgentFlow/internal/buildinfo"
	"github.com/kittors/AgentFlow/internal/debuglog"
	"github.com/kittors/AgentFlow/internal/i18n"
	"github.com/kittors/AgentFlow/internal/install"
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
