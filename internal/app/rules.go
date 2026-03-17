package app

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kittors/AgentFlow/internal/projectrules"
	"github.com/kittors/AgentFlow/internal/ui"
)

func (a *App) runRules(args []string) int {
	if len(args) == 0 {
		if !stdinIsTTY() {
			fmt.Fprintln(a.Stderr, a.Catalog.Msg("用法: agentflow rules <detect|install> ...", "Usage: agentflow rules <detect|install> ..."))
			return 1
		}
		return a.runRulesInteractive()
	}

	switch args[0] {
	case "detect":
		root, quiet, err := parseRootAndQuiet(args[1:], false)
		if err != nil {
			fmt.Fprintln(a.Stderr, err.Error())
			return 1
		}
		manager := projectrules.NewManager()
		statuses, err := manager.Detect(root)
		if err != nil {
			fmt.Fprintln(a.Stderr, err.Error())
			return 1
		}
		if !quiet {
			a.printRuleStatuses(root, statuses)
		}
		return 0

	case "install":
		targets, root, quiet, profile, lang, err := parseRulesInstallArgs(args[1:])
		if err != nil {
			fmt.Fprintln(a.Stderr, err.Error())
			return 1
		}
		manager := projectrules.NewManager()
		written, err := manager.Install(root, targets, projectrules.InstallOptions{Profile: profile, Lang: lang})
		if err != nil {
			fmt.Fprintln(a.Stderr, err.Error())
			return 1
		}
		if !quiet {
			for _, path := range written {
				rel, _ := filepath.Rel(root, path)
				fmt.Fprintln(a.Stdout, filepath.ToSlash(rel))
			}
			fmt.Fprintln(a.Stdout, a.Catalog.Msg("已写入项目规则文件。", "Project rule files written."))
		}
		return 0

	default:
		fmt.Fprintln(a.Stderr, a.Catalog.Msg("未知子命令。", "Unknown subcommand."))
		fmt.Fprintln(a.Stderr, a.Catalog.Msg("用法: agentflow rules <detect|install> ...", "Usage: agentflow rules <detect|install> ..."))
		return 1
	}
}

func (a *App) runRulesInteractive() int {
	root, err := resolveProjectRoot("", false)
	if err != nil {
		fmt.Fprintln(a.Stderr, err.Error())
		return 1
	}

	options := make([]ui.Option, 0, len(projectrules.Names()))
	for _, name := range projectrules.SortedNames() {
		target, _ := projectrules.Lookup(name)
		badge := strings.ToUpper(target.Kind)
		options = append(options, ui.Option{
			Value:       target.Name,
			Label:       target.DisplayName,
			Badge:       badge,
			Description: a.Catalog.Msg("写入该工具的项目级 Skill/规则文件。", "Write project-level skill/rules for this tool."),
		})
	}

	selected, canceled, err := ui.SelectTargets(
		a.Catalog,
		a.Stdout,
		a.Catalog.Msg("选择要写入的项目规则目标", "Select project rule targets"),
		fmt.Sprintf(a.Catalog.Msg("项目目录: %s", "Project root: %s"), filepath.ToSlash(root)),
		options,
	)
	if err != nil {
		fmt.Fprintln(a.Stderr, err.Error())
		return 1
	}
	if canceled || len(selected) == 0 {
		return 0
	}

	profile, canceled, err := ui.SelectProfile(a.Catalog, a.Stdout)
	if err != nil {
		fmt.Fprintln(a.Stderr, err.Error())
		return 1
	}
	if canceled {
		return 0
	}

	manager := projectrules.NewManager()
	if _, err := manager.Install(root, selected, projectrules.InstallOptions{Profile: profile}); err != nil {
		fmt.Fprintln(a.Stderr, err.Error())
		return 1
	}
	fmt.Fprintln(a.Stdout, a.Catalog.Msg("已写入项目规则文件。", "Project rule files written."))
	return 0
}

func parseRulesInstallArgs(args []string) ([]string, string, bool, string, string, error) {
	root := ""
	quiet := false
	profile := ""
	lang := ""
	targets := make([]string, 0, 8)
	for _, arg := range args {
		switch {
		case strings.HasPrefix(arg, "--root="):
			root = strings.TrimPrefix(arg, "--root=")
		case strings.HasPrefix(arg, "--profile="):
			profile = strings.TrimPrefix(arg, "--profile=")
		case strings.HasPrefix(arg, "--lang="):
			lang = strings.TrimPrefix(arg, "--lang=")
		case arg == "--quiet":
			quiet = true
		case strings.HasPrefix(arg, "--"):
			return nil, "", false, "", "", fmt.Errorf("unknown flag: %s", arg)
		default:
			targets = append(targets, arg)
		}
	}

	resolvedRoot, err := resolveProjectRoot(root, false)
	if err != nil {
		return nil, "", false, "", "", err
	}

	dedup := sliceToSet(targets)
	targets = targets[:0]
	for value := range dedup {
		targets = append(targets, value)
	}
	sort.Strings(targets)

	if len(targets) == 0 {
		return nil, "", false, "", "", fmt.Errorf("missing targets")
	}

	return targets, resolvedRoot, quiet, profile, lang, nil
}

func (a *App) printRuleStatuses(root string, statuses []projectrules.Status) {
	sort.Slice(statuses, func(i, j int) bool {
		if statuses[i].Target == statuses[j].Target {
			return statuses[i].Path < statuses[j].Path
		}
		return statuses[i].Target < statuses[j].Target
	})

	fmt.Fprintf(a.Stdout, a.Catalog.Msg("项目目录: %s\n", "Project root: %s\n"), filepath.ToSlash(root))
	for _, status := range statuses {
		rel, _ := filepath.Rel(root, status.Path)
		line := fmt.Sprintf("%s\t%s", status.Target, filepath.ToSlash(rel))
		switch {
		case !status.Exists:
			line += "\t" + a.Catalog.Msg("缺失", "missing")
		case status.Managed:
			line += "\t" + a.Catalog.Msg("已安装(AgentFlow)", "installed (AgentFlow)")
		default:
			line += "\t" + a.Catalog.Msg("已存在(用户)", "present (user)")
		}
		fmt.Fprintln(a.Stdout, line)
	}
}
