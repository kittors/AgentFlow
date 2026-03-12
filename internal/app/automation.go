package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kittors/AgentFlow/internal/kb"
	"github.com/kittors/AgentFlow/internal/projectroot"
	"github.com/kittors/AgentFlow/internal/scan"
)

func (a *App) runInit(args []string) int {
	root, quiet, err := parseRootAndQuiet(args, false)
	if err != nil {
		fmt.Fprintln(a.Stderr, err.Error())
		return 1
	}

	summary, err := kb.InitPaths(root)
	if err != nil {
		fmt.Fprintln(a.Stderr, err.Error())
		return 1
	}
	moduleSummary, err := kb.SyncModules(root, nil)
	if err != nil {
		fmt.Fprintln(a.Stderr, err.Error())
		return 1
	}
	report, err := scan.ScanConventions(root, nil)
	if err != nil {
		fmt.Fprintln(a.Stderr, err.Error())
		return 1
	}
	if _, err := scan.SaveConventions(root, report); err != nil {
		fmt.Fprintln(a.Stderr, err.Error())
		return 1
	}
	if _, err := scan.BuildGraph(root, nil); err != nil {
		fmt.Fprintln(a.Stderr, err.Error())
		return 1
	}
	if _, err := kb.CreateSession(root, kb.SessionInput{
		Stage:         "INIT",
		Tasks:         []string{"Initialize AgentFlow knowledge base"},
		FilesModified: summary.FilesCreated,
		NextSteps:     []string{"Review generated KB files and continue with development."},
	}); err != nil {
		fmt.Fprintln(a.Stderr, err.Error())
		return 1
	}

	if !quiet {
		fmt.Fprintf(
			a.Stdout,
			a.Catalog.Msg(
				"初始化完成。project_type=%s, files_created=%d, modules=%d\n",
				"Initialization complete. project_type=%s, files_created=%d, modules=%d\n",
			),
			summary.ProjectType,
			len(summary.FilesCreated),
			moduleSummary.ModulesFound,
		)
	}
	return 0
}

func (a *App) runKB(args []string) int {
	if len(args) == 0 || args[0] != "sync" {
		fmt.Fprintln(a.Stderr, a.Catalog.Msg("用法: agentflow kb sync [--root=<path>] [--quiet]", "Usage: agentflow kb sync [--root=<path>] [--quiet]"))
		return 1
	}

	root, quiet, err := parseRootAndQuiet(args[1:], true)
	if err != nil {
		fmt.Fprintln(a.Stderr, err.Error())
		return 1
	}

	summary, err := kb.SyncModules(root, nil)
	if err != nil {
		fmt.Fprintln(a.Stderr, err.Error())
		return 1
	}

	if !quiet {
		fmt.Fprintf(
			a.Stdout,
			a.Catalog.Msg("KB 同步完成。modules=%d files=%d\n", "KB sync complete. modules=%d files=%d\n"),
			summary.ModulesFound,
			summary.FilesWritten,
		)
	}
	return 0
}

func (a *App) runSession(args []string) int {
	if len(args) == 0 || args[0] != "save" {
		fmt.Fprintln(a.Stderr, a.Catalog.Msg("用法: agentflow session save [--root=<path>] [--quiet] [--stage=<name>]", "Usage: agentflow session save [--root=<path>] [--quiet] [--stage=<name>]"))
		return 1
	}

	root, quiet, stage, err := parseSessionArgs(args[1:], true)
	if err != nil {
		fmt.Fprintln(a.Stderr, err.Error())
		return 1
	}

	filename, err := kb.CreateSession(root, kb.SessionInput{
		Stage:     stage,
		Tasks:     []string{"Automatic session checkpoint"},
		Decisions: []string{"Saved by Go-native AgentFlow session command."},
		NextSteps: []string{"Continue the current task or review the latest session summary."},
	})
	if err != nil {
		fmt.Fprintln(a.Stderr, err.Error())
		return 1
	}

	if !quiet {
		fmt.Fprintf(a.Stdout, a.Catalog.Msg("会话已保存: %s\n", "Session saved: %s\n"), filepath.ToSlash(filename))
	}
	return 0
}

func (a *App) runConventions(args []string) int {
	root, quiet, err := parseRootAndQuiet(args, false)
	if err != nil {
		fmt.Fprintln(a.Stderr, err.Error())
		return 1
	}

	report, err := scan.ScanConventions(root, nil)
	if err != nil {
		fmt.Fprintln(a.Stderr, err.Error())
		return 1
	}
	filename, err := scan.SaveConventions(root, report)
	if err != nil {
		fmt.Fprintln(a.Stderr, err.Error())
		return 1
	}

	if !quiet {
		fmt.Fprintf(a.Stdout, a.Catalog.Msg("规范提取完成: %s\n", "Convention scan complete: %s\n"), filepath.ToSlash(filename))
	}
	return 0
}

func (a *App) runGraph(args []string) int {
	root, quiet, err := parseRootAndQuiet(args, false)
	if err != nil {
		fmt.Fprintln(a.Stderr, err.Error())
		return 1
	}

	summary, err := scan.BuildGraph(root, nil)
	if err != nil {
		fmt.Fprintln(a.Stderr, err.Error())
		return 1
	}

	if !quiet {
		payload, _ := json.Marshal(summary)
		fmt.Fprintf(a.Stdout, a.Catalog.Msg("图谱构建完成: %s\n", "Graph build complete: %s\n"), string(payload))
	}
	return 0
}

func (a *App) runDashboard(args []string) int {
	root, quiet, err := parseRootAndQuiet(args, false)
	if err != nil {
		fmt.Fprintln(a.Stderr, err.Error())
		return 1
	}

	filename, err := scan.GenerateDashboard(root, nil)
	if err != nil {
		fmt.Fprintln(a.Stderr, err.Error())
		return 1
	}

	if !quiet {
		fmt.Fprintf(a.Stdout, a.Catalog.Msg("Dashboard 已生成: %s\n", "Dashboard generated: %s\n"), filepath.ToSlash(filename))
	}
	return 0
}

func parseRootAndQuiet(args []string, requireAgentFlow bool) (string, bool, error) {
	root := ""
	quiet := false
	for _, arg := range args {
		switch {
		case strings.HasPrefix(arg, "--root="):
			root = strings.TrimPrefix(arg, "--root=")
		case arg == "--quiet":
			quiet = true
		case arg == "":
		default:
			return "", false, fmt.Errorf("unknown flag: %s", arg)
		}
	}

	resolvedRoot, err := resolveProjectRoot(root, requireAgentFlow)
	if err != nil {
		return "", false, err
	}
	return resolvedRoot, quiet, nil
}

func parseSessionArgs(args []string, requireAgentFlow bool) (string, bool, string, error) {
	root := ""
	quiet := false
	stage := "SESSION"
	for _, arg := range args {
		switch {
		case strings.HasPrefix(arg, "--root="):
			root = strings.TrimPrefix(arg, "--root=")
		case strings.HasPrefix(arg, "--stage="):
			stage = strings.TrimPrefix(arg, "--stage=")
		case arg == "--quiet":
			quiet = true
		case arg == "":
		default:
			return "", false, "", fmt.Errorf("unknown flag: %s", arg)
		}
	}

	resolvedRoot, err := resolveProjectRoot(root, requireAgentFlow)
	if err != nil {
		return "", false, "", err
	}
	return resolvedRoot, quiet, stage, nil
}

func resolveProjectRoot(explicitRoot string, requireAgentFlow bool) (string, error) {
	if explicitRoot != "" {
		resolvedRoot, err := filepath.Abs(explicitRoot)
		if err != nil {
			return "", err
		}
		if requireAgentFlow {
			root, err := projectroot.Find(resolvedRoot)
			if err != nil {
				if errors.Is(err, projectroot.ErrRootNotFound) {
					return "", fmt.Errorf("agentflow project root not found from %s", resolvedRoot)
				}
				return "", err
			}
			return root, nil
		}
		return resolvedRoot, nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if !requireAgentFlow {
		if root, findErr := projectroot.Find(cwd); findErr == nil {
			return root, nil
		}
		return cwd, nil
	}

	root, err := projectroot.Find(cwd)
	if err != nil {
		if errors.Is(err, projectroot.ErrRootNotFound) {
			return "", fmt.Errorf("agentflow project root not found from %s", cwd)
		}
		return "", err
	}
	return root, nil
}
