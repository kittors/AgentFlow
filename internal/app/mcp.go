package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kittors/AgentFlow/internal/mcp"
	"github.com/kittors/AgentFlow/internal/targets"
)

func (a *App) runMCP(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(a.Stderr, a.Catalog.Msg("用法: agentflow mcp <install|remove|list|search> ...", "Usage: agentflow mcp <install|remove|list|search> ..."))
		return 1
	}

	manager := mcp.NewManager()

	switch args[0] {
	case "list":
		if len(args) < 2 {
			fmt.Fprintln(a.Stderr, a.Catalog.Msg("用法: agentflow mcp list <target>", "Usage: agentflow mcp list <target>"))
			return 1
		}
		target, ok := targets.Lookup(args[1])
		if !ok {
			fmt.Fprintln(a.Stderr, a.Catalog.Msg("未知目标。", "Unknown target."))
			return 1
		}
		servers, err := manager.List(target)
		if err != nil {
			fmt.Fprintln(a.Stderr, err.Error())
			return 1
		}
		if len(servers) == 0 {
			fmt.Fprintln(a.Stdout, a.Catalog.Msg("未配置任何 MCP servers。", "No MCP servers configured."))
			return 0
		}
		for _, server := range servers {
			fmt.Fprintln(a.Stdout, server)
		}
		return 0

	case "remove":
		if len(args) < 3 {
			fmt.Fprintln(a.Stderr, a.Catalog.Msg("用法: agentflow mcp remove <target> <server>", "Usage: agentflow mcp remove <target> <server>"))
			return 1
		}
		target, ok := targets.Lookup(args[1])
		if !ok {
			fmt.Fprintln(a.Stderr, a.Catalog.Msg("未知目标。", "Unknown target."))
			return 1
		}
		if err := manager.Remove(target, args[2]); err != nil {
			fmt.Fprintln(a.Stderr, err.Error())
			return 1
		}
		fmt.Fprintln(a.Stdout, a.Catalog.Msg("已移除。", "Removed."))
		return 0

	case "install":
		if len(args) < 3 {
			fmt.Fprintln(a.Stderr, a.Catalog.Msg("用法: agentflow mcp install <target> <server> [--set-env=K=V] [--allow=<path>]", "Usage: agentflow mcp install <target> <server> [--set-env=K=V] [--allow=<path>]"))
			return 1
		}
		target, ok := targets.Lookup(args[1])
		if !ok {
			fmt.Fprintln(a.Stderr, a.Catalog.Msg("未知目标。", "Unknown target."))
			return 1
		}

		serverName := ""
		options := mcp.InstallOptions{}
		for _, arg := range args[2:] {
			switch {
			case strings.HasPrefix(arg, "--set-env="):
				options.Env = append(options.Env, strings.TrimPrefix(arg, "--set-env="))
			case strings.HasPrefix(arg, "--allow="):
				options.Allow = append(options.Allow, strings.TrimPrefix(arg, "--allow="))
			case strings.HasPrefix(arg, "--"):
				fmt.Fprintln(a.Stderr, a.Catalog.Msg("未知参数。", "Unknown flag."))
				return 1
			default:
				if serverName != "" {
					fmt.Fprintln(a.Stderr, a.Catalog.Msg("server 只能指定一个。", "server must be a single value."))
					return 1
				}
				serverName = arg
			}
		}
		if strings.TrimSpace(serverName) == "" {
			fmt.Fprintln(a.Stderr, a.Catalog.Msg("缺少 server。", "missing server."))
			return 1
		}

		if len(options.Allow) == 0 && strings.EqualFold(serverName, "filesystem") {
			if wd, err := os.Getwd(); err == nil {
				options.Allow = append(options.Allow, filepath.Clean(wd))
			}
		}

		if err := manager.Install(target, serverName, options); err != nil {
			fmt.Fprintln(a.Stderr, err.Error())
			return 1
		}
		fmt.Fprintln(a.Stdout, a.Catalog.Msg("已写入配置。", "Configuration written."))
		return 0

	case "search":
		if len(args) < 2 {
			fmt.Fprintln(a.Stderr, a.Catalog.Msg("用法: agentflow mcp search <keyword>", "Usage: agentflow mcp search <keyword>"))
			return 1
		}
		results, err := manager.Search(strings.Join(args[1:], " "))
		if err != nil {
			fmt.Fprintln(a.Stderr, err.Error())
			return 1
		}
		if len(results) == 0 {
			fmt.Fprintln(a.Stdout, a.Catalog.Msg("未找到匹配项。", "No matches found."))
			return 0
		}
		for _, line := range results {
			fmt.Fprintln(a.Stdout, line)
		}
		return 0

	default:
		fmt.Fprintln(a.Stderr, a.Catalog.Msg("未知子命令。", "Unknown subcommand."))
		return 1
	}
}
