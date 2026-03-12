package ui

import (
	"fmt"
	"io"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

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

type Panel struct {
	Title string
	Lines []string
}

type selectionModel struct {
	catalog  i18n.Catalog
	title    string
	subtitle string
	hint     string
	options  []Option
	panels   []Panel

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
	headerStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(lipgloss.Color("39")).
			Padding(0, 0, 1, 0)
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("230"))
	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("109"))
	pillStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("24")).
			Padding(0, 1)
	focusPillStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("232")).
			Background(lipgloss.Color("81")).
			Bold(true).
			Padding(0, 1)
	listPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("238")).
			Padding(0, 1)
	detailPanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("63")).
				Padding(0, 1)
	rowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))
	selectedRowStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("230")).
				Background(lipgloss.Color("237")).
				Bold(true)
	badgeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("60")).
			Padding(0, 1)
	selectedBadgeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("232")).
				Background(lipgloss.Color("149")).
				Bold(true).
				Padding(0, 1)
	sectionTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("223"))
	primaryTextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))
	mutedTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244"))
	footerStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(lipgloss.Color("238")).
			Padding(1, 0, 0, 0)
	hintBadgeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("238")).
			Padding(0, 1)
)

func RunMainMenu(catalog i18n.Catalog, version string, panels []Panel, output io.Writer) (Action, bool, error) {
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
		subtitle: catalog.Msg("跨平台 Go CLI。操作列表保持紧凑，状态和结果统一收进右侧面板。", "Cross-platform Go CLI. Keep actions compact while status and results stay inside the right panel."),
		hint:     catalog.Msg("↑/↓ 或滚轮切换动作，Enter 执行，Esc 退出。", "Use ↑/↓ or the mouse wheel to change actions, Enter to run, Esc to exit."),
		options:  options,
		panels:   panels,
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
		tea.WithMouseCellMotion(),
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
	case tea.MouseMsg:
		switch value.Button {
		case tea.MouseButtonWheelUp:
			if m.cursor > 0 {
				m.cursor--
			}
		case tea.MouseButtonWheelDown:
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		}
	case tea.KeyMsg:
		switch value.String() {
		case "ctrl+c", "esc":
			m.canceled = true
			return m, tea.Quit
		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		case "pgup":
			m.cursor -= 5
			if m.cursor < 0 {
				m.cursor = 0
			}
		case "pgdown":
			m.cursor += 5
			if m.cursor > len(m.options)-1 {
				m.cursor = len(m.options) - 1
			}
		case "home":
			m.cursor = 0
		case "end":
			if len(m.options) > 0 {
				m.cursor = len(m.options) - 1
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
	header := m.renderHeader(contentWidth)
	footer := m.renderFooter(contentWidth)

	bodyHeight := m.contentHeight() - lipgloss.Height(header) - lipgloss.Height(footer) - 1
	if bodyHeight < 5 {
		bodyHeight = 5
	}

	body := m.renderBody(contentWidth, bodyHeight)
	view := lipgloss.JoinVertical(lipgloss.Left, header, body, footer)

	if m.width > 0 && m.height > 0 {
		return cropBlock(lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Top, view), m.height)
	}
	return view
}

func (m selectionModel) renderHeader(contentWidth int) string {
	badges := []string{
		pillStyle.Render("Go Binary"),
		pillStyle.Render(m.catalog.Msg("跨平台", "Cross-platform")),
		focusPillStyle.Render(fmt.Sprintf("%d/%d", m.cursor+1, max(1, len(m.options)))),
	}
	if m.multi {
		badges = append(badges, pillStyle.Render(fmt.Sprintf(m.catalog.Msg("已选 %d", "%d selected"), selectedCount(m.options))))
	}

	lines := []string{
		titleStyle.Render(m.title),
		lipgloss.JoinHorizontal(lipgloss.Left, badges...),
	}
	if strings.TrimSpace(m.subtitle) != "" {
		lines = append(lines, wrapStyledLine(subtitleStyle, contentWidth, m.subtitle)...)
	}

	return headerStyle.Width(contentWidth).Render(strings.Join(lines, "\n"))
}

func (m selectionModel) renderBody(contentWidth, bodyHeight int) string {
	if contentWidth >= 92 && bodyHeight >= 6 {
		listWidth := max(28, min(38, contentWidth/3))
		detailWidth := max(24, contentWidth-listWidth-1)
		list := clampBlockHeight(m.renderList(listWidth, bodyHeight), bodyHeight)
		details := clampBlockHeight(m.renderDetails(detailWidth, bodyHeight), bodyHeight)
		return lipgloss.JoinHorizontal(lipgloss.Top, list, " ", details)
	}

	listHeight := min(bodyHeight, max(5, len(m.options)+2))
	if bodyHeight >= 8 {
		listHeight = min(bodyHeight, max(4, min(len(m.options)+2, bodyHeight/2)))
	}
	if len(m.panels) > 0 && bodyHeight >= 9 {
		listHeight = min(listHeight, bodyHeight-5)
	}
	if listHeight > bodyHeight {
		listHeight = bodyHeight
	}
	detailHeight := bodyHeight - listHeight
	if detailHeight < 3 {
		detailHeight = 3
		listHeight = max(2, bodyHeight-detailHeight)
	}

	list := clampBlockHeight(m.renderList(contentWidth, listHeight), listHeight)
	details := clampBlockHeight(m.renderDetails(contentWidth, detailHeight), detailHeight)
	return lipgloss.JoinVertical(lipgloss.Left, list, details)
}

func (m selectionModel) renderList(width, height int) string {
	innerWidth := max(8, width-listPanelStyle.GetHorizontalFrameSize())
	visibleRows := max(1, height-listPanelStyle.GetVerticalFrameSize())
	start, end := m.visibleRange(visibleRows)

	rows := make([]string, 0, end-start)
	for index := start; index < end; index++ {
		rows = append(rows, m.renderRow(index, m.options[index], innerWidth))
	}
	if len(rows) == 0 {
		rows = append(rows, mutedTextStyle.Render(m.catalog.Msg("当前没有可显示的选项。", "There are no options to display.")))
	}

	content := strings.Join(rows, "\n")
	return lipgloss.NewStyle().Width(width).Height(height).Render(listPanelStyle.Width(innerWidth).Render(content))
}

func (m selectionModel) renderRow(index int, option Option, width int) string {
	prefix := "·"
	if index == m.cursor {
		prefix = "›"
	}
	if m.multi {
		prefix = "[ ]"
		if option.Selected {
			prefix = "[x]"
		}
	}

	badgeText := option.Badge
	if strings.TrimSpace(badgeText) == "" {
		badgeText = fmt.Sprintf("%02d", index+1)
	}

	badge := badgeStyle
	row := rowStyle
	if index == m.cursor {
		badge = selectedBadgeStyle
		row = selectedRowStyle
	}

	prefixWidth := lipgloss.Width(prefix) + 1
	badgeWidth := lipgloss.Width(badge.Render(badgeText)) + 1
	labelWidth := max(4, width-prefixWidth-badgeWidth)
	label := ansi.Truncate(option.Label, labelWidth, "…")

	line := lipgloss.JoinHorizontal(
		lipgloss.Left,
		prefix,
		" ",
		badge.Render(badgeText),
		" ",
		label,
	)
	return row.Width(width).Render(line)
}

func (m selectionModel) renderDetails(width, height int) string {
	innerWidth := max(8, width-detailPanelStyle.GetHorizontalFrameSize())
	visibleRows := max(1, height-detailPanelStyle.GetVerticalFrameSize())

	lines := m.detailLines(innerWidth)
	lines = clipLines(lines, visibleRows)
	if len(lines) == 0 {
		lines = []string{mutedTextStyle.Render(m.catalog.Msg("这里会显示当前动作的详情。", "Current action details appear here."))}
	}

	content := strings.Join(lines, "\n")
	return lipgloss.NewStyle().Width(width).Height(height).Render(detailPanelStyle.Width(innerWidth).Render(content))
}

func (m selectionModel) detailLines(width int) []string {
	if len(m.options) == 0 {
		return nil
	}

	current := m.options[m.cursor]
	lines := make([]string, 0, 12)

	for _, panel := range m.panels {
		if strings.TrimSpace(panel.Title) == "" && len(panel.Lines) == 0 {
			continue
		}
		lines = append(lines, "")
		if strings.TrimSpace(panel.Title) != "" {
			lines = append(lines, sectionTitleStyle.Render(panel.Title))
		}
		for _, line := range panel.Lines {
			if strings.TrimSpace(line) == "" {
				lines = append(lines, "")
				continue
			}
			lines = append(lines, primaryTextStyle.Render(ansi.Truncate(line, width, "…")))
		}
	}

	if len(lines) > 0 {
		lines = append(lines, "")
	}
	lines = append(lines,
		sectionTitleStyle.Render(m.catalog.Msg("当前动作", "Current action")),
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			selectedBadgeStyle.Render(defaultBadge(current, m.cursor)),
			" ",
			primaryTextStyle.Bold(true).Render(current.Label),
		),
	)
	if strings.TrimSpace(current.Description) != "" {
		lines = append(lines, wrapStyledLine(mutedTextStyle, width, current.Description)...)
	}

	return lines
}

func (m selectionModel) renderFooter(contentWidth int) string {
	controls := lipgloss.JoinHorizontal(
		lipgloss.Left,
		hintBadgeStyle.Render("↑/↓"),
		" ",
		hintBadgeStyle.Render("Enter"),
		" ",
		hintBadgeStyle.Render("Esc"),
	)
	if m.multi {
		controls = lipgloss.JoinHorizontal(lipgloss.Left, controls, " ", hintBadgeStyle.Render("Space"))
	}

	lines := []string{}
	lines = append(lines, wrapStyledLine(primaryTextStyle, contentWidth, m.currentSummary())...)
	lines = append(lines, wrapStyledLine(mutedTextStyle, contentWidth, lipgloss.JoinHorizontal(lipgloss.Left, controls, "  ", m.hintText()))...)
	return footerStyle.Width(contentWidth).Render(strings.Join(lines, "\n"))
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
		return fmt.Sprintf(m.catalog.Msg("当前动作: %s", "Current action: %s"), current.Label)
	}
	return fmt.Sprintf(m.catalog.Msg("当前动作: %s | %s", "Current action: %s | %s"), current.Label, current.Description)
}

func (m selectionModel) visibleRange(visibleRows int) (int, int) {
	if visibleRows <= 0 || len(m.options) == 0 {
		return 0, 0
	}
	if len(m.options) <= visibleRows {
		return 0, len(m.options)
	}

	start := m.cursor - visibleRows/2
	if start < 0 {
		start = 0
	}
	end := start + visibleRows
	if end > len(m.options) {
		end = len(m.options)
		start = end - visibleRows
	}
	return start, end
}

func (m selectionModel) contentWidth() int {
	if m.width <= 0 {
		return 96
	}

	width := m.width - 4
	switch {
	case width > 112:
		return 112
	case width >= 24:
		return width
	case m.width > 4:
		return m.width - 2
	default:
		return m.width
	}
}

func (m selectionModel) contentHeight() int {
	if m.height <= 0 {
		return 24
	}
	if m.height > 3 {
		return m.height - 1
	}
	return m.height
}

func wrapStyledLine(style lipgloss.Style, width int, text string) []string {
	if width <= 0 {
		return nil
	}
	rendered := style.Width(width).MaxWidth(width).Render(text)
	return strings.Split(rendered, "\n")
}

func clipLines(lines []string, limit int) []string {
	if limit <= 0 {
		return nil
	}
	if len(lines) <= limit {
		return lines
	}
	clipped := append([]string{}, lines[:limit]...)
	clipped[limit-1] = mutedTextStyle.Render("…")
	return clipped
}

func clampBlockHeight(content string, height int) string {
	if height <= 0 {
		return ""
	}
	return lipgloss.NewStyle().Height(height).MaxHeight(height).Render(content)
}

func cropBlock(content string, limit int) string {
	if limit <= 0 {
		return ""
	}
	lines := strings.Split(content, "\n")
	if len(lines) <= limit {
		return content
	}
	return strings.Join(lines[:limit], "\n")
}

func defaultBadge(option Option, index int) string {
	if strings.TrimSpace(option.Badge) != "" {
		return option.Badge
	}
	return fmt.Sprintf("%02d", index+1)
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
