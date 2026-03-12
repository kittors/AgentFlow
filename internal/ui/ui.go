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
}

var (
	heroStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("45")).
			Padding(1, 2)
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("31")).
			Padding(0, 1)
	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("109"))
	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("221")).
			Bold(true)
	selectedRowStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("24")).
				Foreground(lipgloss.Color("230")).
				Padding(0, 1)
	labelStyle = lipgloss.NewStyle().
			Bold(true)
	selectedLabelStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("230"))
	descStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244"))
	selectedDescStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("189"))
	hintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			BorderTop(true).
			BorderForeground(lipgloss.Color("238")).
			PaddingTop(1)
)

func RunMainMenu(catalog i18n.Catalog, version string, output io.Writer) (Action, bool, error) {
	options := []Option{
		{Value: string(ActionInstall), Label: catalog.Msg("安装到 CLI", "Install to CLI targets"), Description: "install"},
		{Value: string(ActionUninstall), Label: catalog.Msg("卸载已安装目标", "Uninstall from installed targets"), Description: "uninstall"},
		{Value: string(ActionUpdate), Label: catalog.Msg("更新 AgentFlow", "Update AgentFlow"), Description: "self-update"},
		{Value: string(ActionStatus), Label: catalog.Msg("查看状态", "Show status"), Description: "status"},
		{Value: string(ActionClean), Label: catalog.Msg("清理缓存", "Clean caches"), Description: "clean"},
		{Value: string(ActionExit), Label: catalog.Msg("退出", "Exit"), Description: "exit"},
	}

	value, _, canceled, err := runSelection(output, selectionModel{
		catalog:  catalog,
		title:    fmt.Sprintf("AgentFlow v%s", version),
		subtitle: catalog.Msg("使用 ↑/↓ 选择，Enter 确认，Esc 退出。", "Use ↑/↓ to move, Enter to confirm, Esc to exit."),
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
	var builder strings.Builder

	builder.WriteString("\n")
	headerLines := []string{titleStyle.Render(m.title)}
	if m.subtitle != "" {
		headerLines = append(headerLines, subtitleStyle.Render(m.subtitle))
	}
	builder.WriteString(heroStyle.Render(strings.Join(headerLines, "\n")))
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
		rowDesc := descStyle.Render(option.Description)
		if index == m.cursor {
			rowLabel = selectedLabelStyle.Render(option.Label)
			rowDesc = selectedDescStyle.Render(option.Description)
		}

		row := cursor + marker + " " + rowLabel
		if option.Description != "" {
			row += "  " + rowDesc
		}
		if index == m.cursor {
			builder.WriteString(selectedRowStyle.Render(row))
		} else {
			builder.WriteString(row)
		}
		builder.WriteString("\n")
	}

	builder.WriteString("\n")
	if m.multi {
		builder.WriteString(hintStyle.Render(m.catalog.Msg("Space 切换选择，Enter 执行，Esc 取消。", "Space toggles selection, Enter runs, Esc cancels.")))
	} else {
		builder.WriteString(hintStyle.Render(m.catalog.Msg("Enter 执行，Esc 返回。", "Enter runs, Esc goes back.")))
	}
	builder.WriteString("\n")

	return builder.String()
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
