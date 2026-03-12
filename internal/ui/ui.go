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
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true)
	labelStyle  = lipgloss.NewStyle().Bold(true)
	descStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
)

func RunMainMenu(catalog i18n.Catalog, version string, output io.Writer) (Action, bool, error) {
	options := []Option{
		{Value: string(ActionInstall), Label: catalog.Msg("安装到 CLI", "Install to CLI targets"), Description: "install"},
		{Value: string(ActionUninstall), Label: catalog.Msg("卸载已安装目标", "Uninstall from installed targets"), Description: "uninstall"},
		{Value: string(ActionUpdate), Label: catalog.Msg("检查更新", "Check for updates"), Description: "update"},
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
	builder.WriteString(titleStyle.Render(m.title))
	builder.WriteString("\n")
	if m.subtitle != "" {
		builder.WriteString(descStyle.Render(m.subtitle))
		builder.WriteString("\n")
	}
	builder.WriteString("\n")

	for index, option := range m.options {
		cursor := "  "
		if index == m.cursor {
			cursor = cursorStyle.Render("› ")
		}

		marker := "•"
		if m.multi {
			marker = "[ ]"
			if option.Selected {
				marker = "[x]"
			}
		}

		builder.WriteString(cursor)
		builder.WriteString(marker)
		builder.WriteString(" ")
		builder.WriteString(labelStyle.Render(option.Label))
		if option.Description != "" {
			builder.WriteString(" ")
			builder.WriteString(descStyle.Render(option.Description))
		}
		builder.WriteString("\n")
	}

	builder.WriteString("\n")
	if m.multi {
		builder.WriteString(descStyle.Render(m.catalog.Msg("Space 切换，Enter 确认，Esc 取消。", "Space toggles, Enter confirms, Esc cancels.")))
	} else {
		builder.WriteString(descStyle.Render(m.catalog.Msg("Enter 确认，Esc 取消。", "Enter confirms, Esc cancels.")))
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
