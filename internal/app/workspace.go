package app

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kittors/AgentFlow/internal/mcp"
	"github.com/kittors/AgentFlow/internal/projectrules"
	"github.com/kittors/AgentFlow/internal/skill"
	"github.com/kittors/AgentFlow/internal/targets"
	"github.com/kittors/AgentFlow/internal/ui"
)

func (a *App) workspaceTargetOptions() []ui.Option {
	options := make([]ui.Option, 0, len(projectrules.Names()))
	for _, name := range projectrules.SortedNames() {
		target, _ := projectrules.Lookup(name)
		badge := strings.ToUpper(target.Kind)
		options = append(options, ui.Option{
			Value:       target.Name,
			Label:       target.DisplayName,
			Badge:       badge,
			Description: a.Catalog.Msg("查看全局 MCP/Skills，并写入项目级规则文件。", "Inspect global MCP/Skills and install project rule files."),
		})
	}
	return options
}

func (a *App) workspacePanel(root, targetName string) ui.Panel {
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
			Title: a.Catalog.Msg("Workspace 概览", "Workspace summary"),
			Lines: []string{a.Catalog.Msg("未知目标。", "Unknown target.")},
		}
	}

	lines := []string{
		fmt.Sprintf(a.Catalog.Msg("目录: %s", "Workspace: %s"), filepath.ToSlash(root)),
		fmt.Sprintf(a.Catalog.Msg("目标: %s", "Target: %s"), target.DisplayName),
		"",
		a.Catalog.Msg("项目规则文件（项目级 Skill/规则）：", "Project rule files (project-level skills/rules):"),
	}

	rulesManager := projectrules.NewManager()
	statuses, detectErr := rulesManager.Detect(root)
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

	lines = append(lines, "")
	lines = append(lines, a.Catalog.Msg("全局 MCP（仅全局）：", "Global MCP (global-only):"))
	if mcpTarget, ok := targets.LookupMCP(target.Name); ok {
		mcpManager := mcp.NewManager()
		servers, err := mcpManager.List(mcpTarget)
		if err != nil {
			lines = append(lines, fmt.Sprintf(a.Catalog.Msg("  [错误] 读取失败: %v", "  [error] read failed: %v"), err))
		} else if len(servers) == 0 {
			lines = append(lines, a.Catalog.Msg("  - （未配置）", "  - (none configured)"))
		} else {
			for _, server := range servers {
				lines = append(lines, "  - "+server)
			}
		}
	} else {
		lines = append(lines, a.Catalog.Msg("  - （该目标不支持 MCP 管理）", "  - (MCP management not supported for this target)"))
	}

	lines = append(lines, "")
	lines = append(lines, a.Catalog.Msg("全局 Skills：", "Global skills:"))
	if skillTarget, ok := targets.Lookup(target.Name); ok {
		skillManager := skill.NewManager()
		skills, err := skillManager.List(skillTarget)
		if err != nil {
			lines = append(lines, fmt.Sprintf(a.Catalog.Msg("  [错误] 读取失败: %v", "  [error] read failed: %v"), err))
		} else if len(skills) == 0 {
			lines = append(lines, a.Catalog.Msg("  - （未安装）", "  - (none installed)"))
		} else {
			sort.Strings(skills)
			for _, name := range skills {
				lines = append(lines, "  - "+name)
			}
		}
	} else {
		lines = append(lines, a.Catalog.Msg("  - （该目标不支持 Skill 管理）", "  - (skill management not supported for this target)"))
	}

	return ui.Panel{
		Title: a.Catalog.Msg("Workspace 概览", "Workspace summary"),
		Lines: lines,
	}
}

func (a *App) workspaceInstallRulesPanel(root, targetName, profile string) ui.Panel {
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
		fmt.Sprintf(a.Catalog.Msg("目录: %s", "Workspace: %s"), filepath.ToSlash(root)),
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
