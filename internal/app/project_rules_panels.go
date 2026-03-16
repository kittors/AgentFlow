package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kittors/AgentFlow/internal/projectrules"
	"github.com/kittors/AgentFlow/internal/ui"
)

func (a *App) projectRulesPanel(root, targetName string) ui.Panel {
	root = strings.TrimSpace(root)
	if root == "" {
		if wd, err := os.Getwd(); err == nil {
			root = wd
		}
	}
	absRoot, err := filepath.Abs(root)
	if err == nil {
		root = absRoot
	}

	target, ok := projectrules.Lookup(targetName)
	if !ok {
		return ui.Panel{
			Title: a.Catalog.Msg("项目规则文件", "Project rule files"),
			Lines: []string{a.Catalog.Msg("未知目标。", "Unknown target.")},
		}
	}

	lines := []string{
		fmt.Sprintf(a.Catalog.Msg("目录: %s", "Directory: %s"), filepath.ToSlash(root)),
		fmt.Sprintf(a.Catalog.Msg("目标: %s", "Target: %s"), target.DisplayName),
		"",
		a.Catalog.Msg("规则文件状态：", "Rule file status:"),
	}

	manager := projectrules.NewManager()
	statuses, detectErr := manager.Detect(root)
	if detectErr != nil {
		lines = append(lines, fmt.Sprintf(a.Catalog.Msg("  [错误] 检测失败: %v", "  [error] detect failed: %v"), detectErr))
	} else {
		found := false
		for _, status := range statuses {
			if status.Target != target.Name {
				continue
			}
			found = true
			state := a.Catalog.Msg("缺失", "missing")
			if status.Exists && status.Managed {
				state = a.Catalog.Msg("已存在（AgentFlow）", "present (AgentFlow)")
			} else if status.Exists {
				state = a.Catalog.Msg("已存在（用户）", "present (user)")
			}
			lines = append(lines, fmt.Sprintf("  - %s: %s", status.Detected, state))
		}
		if !found {
			lines = append(lines, a.Catalog.Msg("  - （无规则文件映射）", "  - (no mapped rule files)"))
		}
	}

	return ui.Panel{
		Title: a.Catalog.Msg("项目规则文件", "Project rule files"),
		Lines: lines,
	}
}

func (a *App) projectRulesInstallPanel(root, targetName, profile string) ui.Panel {
	root = strings.TrimSpace(root)
	if root == "" {
		if wd, err := os.Getwd(); err == nil {
			root = wd
		}
	}
	absRoot, err := filepath.Abs(root)
	if err == nil {
		root = absRoot
	}

	manager := projectrules.NewManager()
	written, err := manager.Install(root, []string{targetName}, projectrules.InstallOptions{Profile: profile})
	if err != nil {
		return errorPanel(a.Catalog.Msg("项目规则写入失败", "Project rules install failed"), err)
	}

	lines := []string{
		fmt.Sprintf(a.Catalog.Msg("目录: %s", "Directory: %s"), filepath.ToSlash(root)),
		fmt.Sprintf(a.Catalog.Msg("目标: %s", "Target: %s"), targetName),
		fmt.Sprintf(a.Catalog.Msg("Profile: %s", "Profile: %s"), profile),
		a.Catalog.Msg("已写入文件：", "Files written:"),
	}
	for _, path := range written {
		rel, _ := filepath.Rel(root, path)
		lines = append(lines, "  - "+filepath.ToSlash(rel))
	}
	return ui.Panel{
		Title: a.Catalog.Msg("项目规则写入结果", "Project rules install result"),
		Lines: lines,
	}
}

func (a *App) projectRulesUninstallPanel(root, targetName string) ui.Panel {
	root = strings.TrimSpace(root)
	if root == "" {
		if wd, err := os.Getwd(); err == nil {
			root = wd
		}
	}
	absRoot, err := filepath.Abs(root)
	if err == nil {
		root = absRoot
	}

	manager := projectrules.NewManager()
	removed, err := manager.Uninstall(root, []string{targetName})
	if err != nil {
		return errorPanel(a.Catalog.Msg("项目规则卸载失败", "Project rules uninstall failed"), err)
	}
	if len(removed) == 0 {
		return ui.Panel{
			Title: a.Catalog.Msg("项目规则卸载结果", "Project rules uninstall result"),
			Lines: []string{
				fmt.Sprintf(a.Catalog.Msg("目录: %s", "Directory: %s"), filepath.ToSlash(root)),
				a.Catalog.Msg("没有找到 AgentFlow 管理的项目规则文件。", "No AgentFlow-managed project rule files found."),
			},
		}
	}
	lines := []string{
		fmt.Sprintf(a.Catalog.Msg("目录: %s", "Directory: %s"), filepath.ToSlash(root)),
		a.Catalog.Msg("已删除文件：", "Files removed:"),
	}
	for _, path := range removed {
		rel, _ := filepath.Rel(root, path)
		lines = append(lines, "  - "+filepath.ToSlash(rel))
	}
	return ui.Panel{
		Title: a.Catalog.Msg("项目规则卸载结果", "Project rules uninstall result"),
		Lines: lines,
	}
}
