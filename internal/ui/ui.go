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
	Badge       string
	Selected    bool
}

type selectionModel struct {
	catalog  i18n.Catalog
	title    string
	subtitle string
	hint     string
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
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("81")).
			Background(lipgloss.Color("235")).
			Padding(1, 2)
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("230"))
	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("151"))
	headerBadgeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("230")).
				Background(lipgloss.Color("24")).
				Padding(0, 1)
	headerFocusBadgeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("232")).
				Background(lipgloss.Color("149")).
				Bold(true).
				Padding(0, 1)
	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("239")).
			Padding(0, 1)
	selectedCardStyle = lipgloss.NewStyle().
				Border(lipgloss.DoubleBorder()).
				BorderForeground(lipgloss.Color("81")).
				Background(lipgloss.Color("236")).
				Padding(0, 1)
	badgeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("60")).
			Padding(0, 1)
	selectedBadgeStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("232")).
				Background(lipgloss.Color("149")).
				Padding(0, 1)
	cursorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("221"))
	mutedCursorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("242"))
	labelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("230"))
	selectedLabelStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("230"))
	descStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("246"))
	selectedDescStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("189"))
	footerStyle = lipgloss.NewStyle().
			BorderTop(true).
			BorderForeground(lipgloss.Color("238")).
			PaddingTop(1)
	footerSummaryStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))
	hintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))
	hintBadgeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("238")).
			Padding(0, 1)
)

func RunMainMenu(catalog i18n.Catalog, version string, output io.Writer) (Action, bool, error) {
	options := []Option{
		{
			Value:       string(ActionInstall),
			Label:       catalog.Msg("安装到 CLI", "Install to CLI targets"),
			Badge:       catalog.Msg("安装", "SETUP"),
			Description: catalog.Msg("把 AgentFlow 规则、模块、技能和 hooks 部署到 Codex、Claude、Gemini 等 CLI。", "Deploy AgentFlow rules, modules, skills, and hooks into Codex, Claude, Gemini, and other CLIs."),
		},
		{
			Value:       string(ActionUninstall),
			Label:       catalog.Msg("卸载已安装目标", "Uninstall from installed targets"),
			Badge:       catalog.Msg("移除", "REMOVE"),
			Description: catalog.Msg("从已接入 CLI 中清理 AgentFlow 产物，同时保留你的原有配置。", "Remove AgentFlow assets from integrated CLIs while preserving your own config where possible."),
		},
		{
			Value:       string(ActionUpdate),
			Label:       catalog.Msg("更新 AgentFlow", "Update AgentFlow"),
			Badge:       catalog.Msg("更新", "UPDATE"),
			Description: catalog.Msg("检查最新 release，并把当前 Go 二进制原地更新到新版本。", "Check the latest release and replace the current Go binary in place."),
		},
		{
			Value:       string(ActionStatus),
			Label:       catalog.Msg("查看状态", "Show status"),
			Badge:       catalog.Msg("状态", "STATUS"),
			Description: catalog.Msg("查看当前版本、可执行文件路径和所有支持 CLI 的接入状态。", "Inspect the current version, executable path, and integration status for every supported CLI."),
		},
		{
			Value:       string(ActionClean),
			Label:       catalog.Msg("清理缓存", "Clean caches"),
			Badge:       catalog.Msg("清理", "CLEAN"),
			Description: catalog.Msg("清除 AgentFlow 生成的缓存、临时目录和派生产物，保持环境整洁。", "Remove AgentFlow caches, temporary directories, and derived artifacts to keep the environment tidy."),
		},
		{
			Value:       string(ActionExit),
			Label:       catalog.Msg("退出", "Exit"),
			Badge:       catalog.Msg("退出", "EXIT"),
			Description: catalog.Msg("退出交互菜单并返回终端。", "Leave the interactive menu and return to the terminal."),
		},
	}

	value, _, canceled, err := runSelection(output, selectionModel{
		catalog:  catalog,
		title:    fmt.Sprintf("AgentFlow v%s", version),
		subtitle: catalog.Msg("跨平台 Go CLI。先选动作，再把 AgentFlow 布进你的代理工作流。", "Cross-platform Go CLI. Pick an action, then route AgentFlow into your agent workflow."),
		hint:     catalog.Msg("↑/↓ 切换动作，Enter 执行，Esc 退出。", "Use ↑/↓ to switch actions, Enter to run, Esc to exit."),
		options:  options,
	})
	return Action(value), canceled, err
}

func SelectLanguage(defaultLanguage string, output io.Writer) (string, bool, error) {
	catalog := i18n.NewCatalogWithLanguage(defaultLanguage)
	cursor := 1
	if strings.EqualFold(defaultLanguage, string(i18n.LocaleZH)) {
		cursor = 0
	}

	value, _, canceled, err := runSelection(output, selectionModel{
		catalog:  catalog,
		title:    "Select language / 选择语言",
		subtitle: "Choose how AgentFlow should speak on this machine. / 选择 AgentFlow 在这台设备上的界面语言。",
		hint:     "Enter 确认 / Enter to confirm · Esc 返回 / Esc to go back",
		options: []Option{
			{
				Value:       string(i18n.LocaleZH),
				Label:       "中文",
				Badge:       "ZH",
				Description: "界面、提示和安装反馈优先显示中文。",
			},
			{
				Value:       string(i18n.LocaleEN),
				Label:       "English",
				Badge:       "EN",
				Description: "Use English for menus, prompts, and installation feedback.",
			},
		},
		cursor: cursor,
	})
	return value, canceled, err
}

func SelectProfile(catalog i18n.Catalog, output io.Writer) (string, bool, error) {
	value, _, canceled, err := runSelection(output, selectionModel{
		catalog:  catalog,
		title:    catalog.Msg("选择部署 Profile", "Select deployment profile"),
		subtitle: catalog.Msg("Profile 会影响注入到目标 CLI 的规则深度和功能范围。", "Profiles control how much AgentFlow logic is injected into the target CLI."),
		hint:     catalog.Msg("↑/↓ 切换 Profile，Enter 确认。", "Use ↑/↓ to switch profiles, then press Enter."),
		options: []Option{
			{Value: "lite", Label: "lite", Badge: catalog.Msg("轻量", "LITE"), Description: catalog.Msg("只部署核心规则，最省 token。", "Deploy only the core rules for the smallest token footprint.")},
			{Value: "standard", Label: "standard", Badge: catalog.Msg("标准", "STANDARD"), Description: catalog.Msg("核心规则 + 常用模块，适合大多数项目。", "Core rules plus the common modules for most projects.")},
			{Value: "full", Label: "full", Badge: catalog.Msg("完整", "FULL"), Description: catalog.Msg("完整功能集，包含子代理、注意力和 Hooks。", "Full feature set including sub-agents, attention, and hooks."), Selected: true},
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
		hint:     catalog.Msg("Space 选择多个目标，Enter 执行，Esc 取消。", "Use Space to select multiple targets, Enter to run, Esc to cancel."),
		options:  options,
		multi:    true,
	})
	return values, canceled, err
}

func runSelection(output io.Writer, model selectionModel) (string, []string, bool, error) {
	if output == nil {
		output = io.Discard
	}

	program := tea.NewProgram(
		model,
		tea.WithOutput(output),
		tea.WithAltScreen(),
	)
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
	builder.WriteString(heroStyle.Width(contentWidth).Render(m.renderHeader(contentWidth)))
	builder.WriteString("\n\n")

	for index, option := range m.options {
		builder.WriteString(m.renderOption(index, option, contentWidth))
		builder.WriteString("\n")
	}

	builder.WriteString("\n")
	builder.WriteString(footerStyle.Width(contentWidth).Render(m.renderFooter()))
	builder.WriteString("\n")

	view := builder.String()
	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, view)
	}
	return view
}

func (m selectionModel) renderHeader(contentWidth int) string {
	badges := []string{
		headerBadgeStyle.Render("Go Binary"),
		headerBadgeStyle.Render(m.catalog.Msg("跨平台", "Cross-platform")),
		headerFocusBadgeStyle.Render(fmt.Sprintf("%d/%d", m.cursor+1, max(1, len(m.options)))),
	}
	if m.multi {
		badges = append(badges, headerBadgeStyle.Render(fmt.Sprintf(m.catalog.Msg("已选 %d", "%d selected"), selectedCount(m.options))))
	}

	lines := []string{
		titleStyle.Render(m.title),
		lipgloss.JoinHorizontal(lipgloss.Left, badges...),
	}
	if m.subtitle != "" {
		lines = append(lines, subtitleStyle.Width(contentWidth-4).Render(m.subtitle))
	}
	return strings.Join(lines, "\n")
}

func (m selectionModel) renderOption(index int, option Option, contentWidth int) string {
	card := cardStyle
	badge := badgeStyle
	label := labelStyle
	description := descStyle
	cursor := mutedCursorStyle.Render("·")

	if index == m.cursor {
		card = selectedCardStyle
		badge = selectedBadgeStyle
		label = selectedLabelStyle
		description = selectedDescStyle
		cursor = cursorStyle.Render("▶")
	}

	prefix := cursor
	if m.multi {
		prefix = "[ ]"
		if option.Selected {
			prefix = "[x]"
		}
		if index == m.cursor {
			prefix = cursorStyle.Render(prefix)
		} else {
			prefix = mutedCursorStyle.Render(prefix)
		}
	}

	badgeText := option.Badge
	if strings.TrimSpace(badgeText) == "" {
		badgeText = fmt.Sprintf("%02d", index+1)
	}

	lines := []string{
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			prefix,
			" ",
			badge.Render(badgeText),
			" ",
			label.Render(option.Label),
		),
	}
	if strings.TrimSpace(option.Description) != "" {
		lines = append(lines, "  "+description.Width(max(20, contentWidth-8)).Render(option.Description))
	}

	return card.Width(contentWidth).Render(strings.Join(lines, "\n"))
}

func (m selectionModel) renderFooter() string {
	controls := lipgloss.JoinHorizontal(
		lipgloss.Left,
		hintBadgeStyle.Render("↑/↓"),
		" ",
		hintBadgeStyle.Render("Enter"),
		" ",
		hintBadgeStyle.Render("Esc"),
	)

	lines := []string{
		footerSummaryStyle.Render(m.currentSummary()),
		lipgloss.JoinHorizontal(lipgloss.Left, controls, "  ", hintStyle.Render(m.hintText())),
	}
	return strings.Join(lines, "\n")
}

func (m selectionModel) hintText() string {
	if strings.TrimSpace(m.hint) != "" {
		return m.hint
	}
	if m.multi {
		return m.catalog.Msg("Space 切换选择，Enter 执行，Esc 取消。", "Space toggles selection, Enter runs, Esc cancels.")
	}
	return m.catalog.Msg("Enter 执行，Esc 返回。", "Enter runs, Esc goes back.")
}

func (m selectionModel) currentSummary() string {
	if len(m.options) == 0 {
		return m.catalog.Msg("当前没有可显示的选项。", "There are no options to display.")
	}

	current := m.options[m.cursor]
	if m.multi {
		if selectedCount(m.options) == 0 {
			return fmt.Sprintf(
				m.catalog.Msg("当前目标: %s | 先用 Space 选中要执行的项目。", "Current target: %s | Press Space before running."),
				current.Label,
			)
		}
		return fmt.Sprintf(
			m.catalog.Msg("当前目标: %s | 已选择 %d 项，按 Enter 执行。", "Current target: %s | %d selected, press Enter to continue."),
			current.Label,
			selectedCount(m.options),
		)
	}

	if strings.TrimSpace(current.Description) == "" {
		return fmt.Sprintf(
			m.catalog.Msg("当前动作: %s", "Current action: %s"),
			current.Label,
		)
	}
	return fmt.Sprintf(
		m.catalog.Msg("当前动作: %s | %s", "Current action: %s | %s"),
		current.Label,
		current.Description,
	)
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

func (m selectionModel) contentWidth() int {
	if m.width <= 0 {
		return 84
	}

	width := m.width - 10
	switch {
	case width < 58:
		return 58
	case width > 94:
		return 94
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
