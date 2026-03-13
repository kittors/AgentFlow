package app

import (
	"fmt"
	"strings"

	"github.com/kittors/AgentFlow/internal/skill"
	"github.com/kittors/AgentFlow/internal/targets"
)

func (a *App) runSkill(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(a.Stderr, a.Catalog.Msg("用法: agentflow skill <install|uninstall|list> ...", "Usage: agentflow skill <install|uninstall|list> ..."))
		return 1
	}

	manager := skill.NewManager()

	switch args[0] {
	case "list":
		if len(args) < 2 {
			fmt.Fprintln(a.Stderr, a.Catalog.Msg("用法: agentflow skill list <target>", "Usage: agentflow skill list <target>"))
			return 1
		}
		target, ok := targets.Lookup(args[1])
		if !ok {
			fmt.Fprintln(a.Stderr, a.Catalog.Msg("未知目标。", "Unknown target."))
			return 1
		}
		names, err := manager.List(target)
		if err != nil {
			fmt.Fprintln(a.Stderr, err.Error())
			return 1
		}
		if len(names) == 0 {
			fmt.Fprintln(a.Stdout, a.Catalog.Msg("未检测到已安装的 skill。", "No installed skills detected."))
			return 0
		}
		for _, name := range names {
			fmt.Fprintln(a.Stdout, name)
		}
		return 0

	case "uninstall":
		if len(args) < 3 {
			fmt.Fprintln(a.Stderr, a.Catalog.Msg("用法: agentflow skill uninstall <target> <skillName>", "Usage: agentflow skill uninstall <target> <skillName>"))
			return 1
		}
		target, ok := targets.Lookup(args[1])
		if !ok {
			fmt.Fprintln(a.Stderr, a.Catalog.Msg("未知目标。", "Unknown target."))
			return 1
		}
		if err := manager.Uninstall(target, args[2]); err != nil {
			fmt.Fprintln(a.Stderr, err.Error())
			return 1
		}
		fmt.Fprintln(a.Stdout, a.Catalog.Msg("已卸载。", "Uninstalled."))
		return 0

	case "install":
		if len(args) < 3 {
			fmt.Fprintln(a.Stderr, a.Catalog.Msg("用法: agentflow skill install <target> <source> [--skill=<name>] [--ref=<gitref>] [--force]", "Usage: agentflow skill install <target> <source> [--skill=<name>] [--ref=<gitref>] [--force]"))
			return 1
		}
		target, ok := targets.Lookup(args[1])
		if !ok {
			fmt.Fprintln(a.Stderr, a.Catalog.Msg("未知目标。", "Unknown target."))
			return 1
		}

		source := ""
		options := skill.InstallOptions{}
		for _, arg := range args[2:] {
			switch {
			case strings.HasPrefix(arg, "--skill="):
				options.Skill = strings.TrimPrefix(arg, "--skill=")
			case strings.HasPrefix(arg, "--ref="):
				options.Ref = strings.TrimPrefix(arg, "--ref=")
			case arg == "--force":
				options.Force = true
			case strings.HasPrefix(arg, "--"):
				fmt.Fprintln(a.Stderr, a.Catalog.Msg("未知参数。", "Unknown flag."))
				return 1
			default:
				if source != "" {
					fmt.Fprintln(a.Stderr, a.Catalog.Msg("source 只能指定一个。", "source must be a single value."))
					return 1
				}
				source = arg
			}
		}
		if strings.TrimSpace(source) == "" {
			fmt.Fprintln(a.Stderr, a.Catalog.Msg("缺少 source。", "missing source."))
			return 1
		}
		name, err := manager.Install(target, source, options)
		if err != nil {
			fmt.Fprintln(a.Stderr, err.Error())
			return 1
		}
		fmt.Fprintf(a.Stdout, a.Catalog.Msg("安装完成: %s\n", "Installed: %s\n"), name)
		return 0

	default:
		fmt.Fprintln(a.Stderr, a.Catalog.Msg("未知子命令。", "Unknown subcommand."))
		return 1
	}
}
