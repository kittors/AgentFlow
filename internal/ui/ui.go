package ui

import (
	"fmt"
	"io"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/kittors/AgentFlow/internal/i18n"
)

type Action string

const (
	ActionInstall   Action = "install"
	ActionUninstall Action = "uninstall"
	ActionUpdate    Action = "update"
	ActionStatus    Action = "status"
	ActionClean     Action = "clean"
	ActionExit      Action = "exit"
)

type Option struct {
	Value       string
	Label       string
	Description string
	Selected    bool
}

type selectionModel struct {
	catalog  i18n.Catalog
	title    string
	subtitle string
	options  []Option
	cursor   int
	multi    bool
	done     bool
	canceled bool
	value    string
	values   []string
	width    int
	height   int
}

var (
	heroStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("81")).
			Padding(1, 2)
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("230"))
	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("151"))
	badgeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("24")).
			Padding(0, 1)
	highlightBadgeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("232")).
				Background(lipgloss.Color("149")).
				Bold(true).
				Padding(0, 1)
	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("221")).
			Bold(true)
	rowStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(lipgloss.Color("238")).
			Padding(0, 1)
	selectedRowStyle = lipgloss.NewStyle().
				Border(lipgloss.ThickBorder(), false, false, false, true).
				BorderForeground(lipgloss.Color("81")).
				Background(lipgloss.Color("236")).
				Foreground(lipgloss.Color("230")).
				Padding(0, 1)
	labelStyle = lipgloss.NewStyle().
			Bold(true)
	selectedLabelStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("230"))
	metaStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("109"))
	selectedMetaStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("153")).
				Bold(true)
	descStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244"))
	selectedDescStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("189"))
	footerStyle = lipgloss.NewStyle().
			BorderTop(true).
			BorderForeground(lipgloss.Color("238")).
			PaddingTop(1)
	hintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			PaddingTop(1)
)

func RunMainMenu(catalog i18n.Catalog, version string, output io.Writer) (Action, bool, error) {
	options := []Option{
		{Value: string(ActionInstall), Label: catalog.Msg("安装到 CLI", "Install to CLI targets"), Description: catalog.Msg("将 AgentFlow 写入 Codex、Claude、Gemini 等 CLI 配置。", "Write AgentFlow into Codex, Claude, Gemini, and other CLI configs.")},
		{Value: string(ActionUninstall), Label: catalog.Msg("卸载已安装目标", "Uninstall from installed targets"), Description: catalog.Msg("从已接入的 CLI 中移除 AgentFlow 规则与资源。", "Remove AgentFlow rules and assets from integrated CLI targets.")},
		{Value: string(ActionUpdate), Label: catalog.Msg("更新 AgentFlow", "Update AgentFlow"), Description: catalog.Msg("下载并替换当前 Go 二进制。", "Download and replace the current Go binary.")},
		{Value: string(ActionStatus), Label: catalog.Msg("查看状态", "Show status"), Description: catalog.Msg("查看已检测目标、版本和运行环境。", "Inspect detected targets, version, and runtime environment.")},
		{Value: string(ActionClean), Label: catalog.Msg("清理缓存", "Clean caches"), Description: catalog.Msg("清理缓存、临时下载和派生产物。", "Clean caches, temporary downloads, and derived artifacts.")},
		{Value: string(ActionExit), Label: catalog.Msg("退出", "Exit"), Description: catalog.Msg("关闭菜单并返回终端。", "Close the menu and return to the terminal.")},
	}

	value, _, canceled, err := runSelection(output, selectionModel{
		catalog:  catalog,
		title:    fmt.Sprintf("AgentFlow v%s", version),
		subtitle: catalog.Msg("跨平台 Go 可执行文件。使用 ↑/↓ 选择，Enter 执行，Esc 退出。", "Cross-platform Go executable. Use ↑/↓ to move, Enter to run, Esc to exit."),
		options:  options,
	})
	return Action(value), canceled, err
}

func SelectProfile(catalog i18n.Catalog, output io.Writer) (string, bool, error) {
	value, _, canceled, err := runSelection(output, selectionModel{
		catalog:  catalog,
		title:    catalog.Msg("选择部署 Profile", "Select deployment profile"),
		subtitle: catalog.Msg("lite 最省 token；full 包含子代理、注意力和 Hooks。", "lite is minimal; full includes sub-agents, attention, and hooks."),
		options: []Option{
			{Value: "lite", Label: "lite", Description: catalog.Msg("仅核心规则", "Core rules only")},
			{Value: "standard", Label: "standard", Description: catalog.Msg("核心规则 + 通用模块", "Core rules + common modules")},
			{Value: "full", Label: "full", Description: catalog.Msg("完整功能集", "Full feature set"), Selected: true},
		},
		cursor: 2,
	})
	return value, canceled, err
}

func SelectTargets(catalog i18n.Catalog, output io.Writer, title, subtitle string, options []Option) ([]string, bool, error) {
	_, values, canceled, err := runSelection(output, selectionModel{
		catalog:  catalog,
		title:    title,
		subtitle: subtitle,
		options:  options,
		multi:    true,
	})
	return values, canceled, err
}

func runSelection(output io.Writer, model selectionModel) (string, []string, bool, error) {
	if output == nil {
		output = io.Discard
	}

	program := tea.NewProgram(model, tea.WithOutput(output))
	finalModel, err := program.Run()
	if err != nil {
		return "", nil, false, err
	}

	result, ok := finalModel.(selectionModel)
	if !ok {
		return "", nil, false, fmt.Errorf("unexpected model type %T", finalModel)
	}
	return result.value, result.values, result.canceled, nil
}

func (m selectionModel) Init() tea.Cmd {
	return nil
}

func (m selectionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch value := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = value.Width
		m.height = value.Height
	case tea.KeyMsg:
		switch value.String() {
		case "ctrl+c", "esc", "q":
			m.canceled = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		case " ":
			if m.multi && len(m.options) > 0 {
				m.options[m.cursor].Selected = !m.options[m.cursor].Selected
			}
		case "enter":
			if len(m.options) == 0 {
				m.canceled = true
				return m, tea.Quit
			}
			if m.multi {
				selected := selectedValues(m.options)
				if len(selected) == 0 {
					return m, nil
				}
				m.values = selected
			} else {
				m.value = m.options[m.cursor].Value
			}
			m.done = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m selectionModel) View() string {
	contentWidth := m.contentWidth()
	var builder strings.Builder

	builder.WriteString("\n")
	headerBadges := []string{
		badgeStyle.Render(m.catalog.Msg("Go Binary", "Go Binary")),
		badgeStyle.Render(m.catalog.Msg("Cross-platform", "Cross-platform")),
		highlightBadgeStyle.Render(fmt.Sprintf("%d/%d", m.cursor+1, max(1, len(m.options)))),
	}
	if m.multi {
		headerBadges = append(headerBadges, badgeStyle.Render(fmt.Sprintf(m.catalog.Msg("已选 %d", "%d selected"), selectedCount(m.options))))
	}

	headerLines := []string{
		titleStyle.Render(m.title),
		lipgloss.JoinHorizontal(lipgloss.Left, headerBadges...),
	}
	if m.subtitle != "" {
		headerLines = append(headerLines, subtitleStyle.Width(contentWidth-4).Render(m.subtitle))
	}
	builder.WriteString(heroStyle.Width(contentWidth).Render(strings.Join(headerLines, "\n")))
	builder.WriteString("\n\n")

	for index, option := range m.options {
		cursor := "  "
		if index == m.cursor {
			cursor = cursorStyle.Render("→ ")
		}

		marker := "•"
		if m.multi {
			marker = "[ ]"
			if option.Selected {
				marker = "[x]"
			}
		}

		rowLabel := labelStyle.Render(option.Label)
		rowDesc := descStyle.Width(max(12, contentWidth-10)).Render(option.Description)
		rowMeta := metaStyle.Render(option.Value)
		if index == m.cursor {
			rowLabel = selectedLabelStyle.Render(option.Label)
			rowDesc = selectedDescStyle.Width(max(12, contentWidth-10)).Render(option.Description)
			rowMeta = selectedMetaStyle.Render(option.Value)
		}

		row := cursor + marker + " " + rowLabel
		if option.Value != "" {
			row += "  " + rowMeta
		}
		if option.Description != "" {
			row += "\n    " + rowDesc
		}
		if index == m.cursor {
			builder.WriteString(selectedRowStyle.Width(contentWidth).Render(row))
		} else {
			builder.WriteString(rowStyle.Width(contentWidth).Render(row))
		}
		builder.WriteString("\n")
	}

	builder.WriteString("\n")
	builder.WriteString(footerStyle.Width(contentWidth).Render(m.currentSummary()))
	builder.WriteString("\n")
	if m.multi {
		builder.WriteString(hintStyle.Render(m.catalog.Msg("Space 切换选择，Enter 执行，Esc 取消。", "Space toggles selection, Enter runs, Esc cancels.")))
	} else {
		builder.WriteString(hintStyle.Render(m.catalog.Msg("Enter 执行，Esc 返回。", "Enter runs, Esc goes back.")))
	}
	builder.WriteString("\n")

	view := builder.String()
	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, view)
	}
	return view
}

func selectedValues(options []Option) []string {
	values := make([]string, 0, len(options))
	for _, option := range options {
		if option.Selected {
			values = append(values, option.Value)
		}
	}
	return values
}

func selectedCount(options []Option) int {
	count := 0
	for _, option := range options {
		if option.Selected {
			count++
		}
	}
	return count
}

func (m selectionModel) currentSummary() string {
	if len(m.options) == 0 {
		return m.catalog.Msg("当前没有可显示的选项。", "There are no options to display.")
	}

	current := m.options[m.cursor]
	if m.multi {
		return fmt.Sprintf(
			m.catalog.Msg("当前目标: %s | 已选择 %d 项。", "Current target: %s | %d selected."),
			current.Label,
			selectedCount(m.options),
		)
	}

	return fmt.Sprintf(
		m.catalog.Msg("当前动作: %s | %s", "Current action: %s | %s"),
		current.Label,
		current.Description,
	)
}

func (m selectionModel) contentWidth() int {
	if m.width <= 0 {
		return 72
	}

	width := m.width - 8
	switch {
	case width < 48:
		return 48
	case width > 88:
		return 88
	default:
		return width
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
