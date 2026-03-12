package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/kittors/AgentFlow/internal/i18n"
)

func TestSingleSelectEnterChoosesCurrentOption(t *testing.T) {
	model := selectionModel{
		catalog: i18n.NewCatalog(),
		options: []Option{
			{Value: "install", Label: "install"},
			{Value: "status", Label: "status"},
		},
		cursor: 1,
	}

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	result := next.(selectionModel)

	if !result.done {
		t.Fatal("expected selection to finish")
	}
	if result.value != "status" {
		t.Fatalf("expected selected value %q, got %q", "status", result.value)
	}
}

func TestMultiSelectSpaceAndEnterCollectSelections(t *testing.T) {
	model := selectionModel{
		catalog: i18n.NewCatalog(),
		multi:   true,
		options: []Option{
			{Value: "codex", Label: "codex"},
			{Value: "claude", Label: "claude"},
		},
	}

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeySpace})
	model = next.(selectionModel)
	next, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = next.(selectionModel)
	next, _ = model.Update(tea.KeyMsg{Type: tea.KeySpace})
	model = next.(selectionModel)
	next, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	result := next.(selectionModel)

	if !result.done {
		t.Fatal("expected multi-select to finish")
	}
	if len(result.values) != 2 {
		t.Fatalf("expected 2 selected values, got %d", len(result.values))
	}
	if result.values[0] != "codex" || result.values[1] != "claude" {
		t.Fatalf("unexpected selected values: %#v", result.values)
	}
}

func TestEscapeCancelsSelection(t *testing.T) {
	model := selectionModel{
		catalog: i18n.NewCatalog(),
		options: []Option{{Value: "exit", Label: "exit"}},
	}

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	result := next.(selectionModel)

	if !result.canceled {
		t.Fatal("expected selection to be canceled")
	}
}

func TestPrintableKeysDoNotMoveOrCancelSelection(t *testing.T) {
	model := selectionModel{
		catalog: i18n.NewCatalog(),
		options: []Option{
			{Value: "install", Label: "install"},
			{Value: "status", Label: "status"},
		},
		cursor: 1,
	}

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	model = next.(selectionModel)
	if model.cursor != 1 {
		t.Fatalf("expected printable key to leave cursor unchanged, got %d", model.cursor)
	}
	if model.canceled {
		t.Fatal("expected printable key not to cancel selection")
	}

	next, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	model = next.(selectionModel)
	if model.canceled {
		t.Fatal("expected q not to cancel selection")
	}
}

func TestViewShowsCurrentSummaryAndBadges(t *testing.T) {
	model := selectionModel{
		catalog:  i18n.NewCatalog(),
		title:    "AgentFlow v1.0.3",
		subtitle: "Cross-platform Go executable.",
		options: []Option{
			{Value: "install", Label: "Install", Badge: "SETUP", Description: "Write AgentFlow into CLI configs."},
			{Value: "update", Label: "Update AgentFlow", Badge: "UPDATE", Description: "Download and replace the current Go binary."},
		},
		cursor: 1,
		width:  100,
		height: 30,
	}

	view := model.View()

	if !strings.Contains(view, "AgentFlow v1.0.3") {
		t.Fatalf("expected title in view, got %q", view)
	}
	if !strings.Contains(view, "Go Binary") {
		t.Fatalf("expected Go Binary badge in view, got %q", view)
	}
	if !strings.Contains(view, "UPDATE") {
		t.Fatalf("expected card badge in view, got %q", view)
	}
	if !strings.Contains(view, "2/2") {
		t.Fatalf("expected cursor badge in view, got %q", view)
	}
	if !strings.Contains(view, model.currentSummary()) {
		t.Fatalf("expected current summary in view, got %q", view)
	}
}

func TestMultiSelectSummaryPromptsForSelection(t *testing.T) {
	model := selectionModel{
		catalog: i18n.NewCatalog(),
		multi:   true,
		options: []Option{
			{Value: "codex", Label: "codex", Badge: "CODEX"},
			{Value: "claude", Label: "claude", Badge: "CLAUDE"},
		},
	}

	summary := model.currentSummary()
	if !strings.Contains(summary, "Space") && !strings.Contains(summary, "选中") {
		t.Fatalf("expected summary to guide selection, got %q", summary)
	}
}

func TestMouseWheelMovesCursor(t *testing.T) {
	model := selectionModel{
		catalog: i18n.NewCatalog(),
		options: []Option{
			{Value: "install", Label: "install"},
			{Value: "status", Label: "status"},
			{Value: "clean", Label: "clean"},
		},
		cursor: 1,
	}

	next, _ := model.Update(tea.MouseMsg{Button: tea.MouseButtonWheelDown, Action: tea.MouseActionPress, Type: tea.MouseWheelDown})
	model = next.(selectionModel)
	if model.cursor != 2 {
		t.Fatalf("expected cursor to move down to 2, got %d", model.cursor)
	}

	next, _ = model.Update(tea.MouseMsg{Button: tea.MouseButtonWheelUp, Action: tea.MouseActionPress, Type: tea.MouseWheelUp})
	model = next.(selectionModel)
	if model.cursor != 1 {
		t.Fatalf("expected cursor to move up to 1, got %d", model.cursor)
	}
}

func TestViewKeepsContentInsideAvailableHeight(t *testing.T) {
	model := selectionModel{
		catalog:  i18n.NewCatalog(),
		title:    "AgentFlow",
		subtitle: "Compact layout should not overflow when the terminal is short.",
		options: []Option{
			{Value: "install", Label: "Install", Badge: "SETUP", Description: "Write AgentFlow into CLI configs."},
			{Value: "uninstall", Label: "Uninstall", Badge: "REMOVE", Description: "Remove AgentFlow from detected CLIs."},
			{Value: "update", Label: "Update", Badge: "UPDATE", Description: "Replace the current Go binary."},
			{Value: "status", Label: "Status", Badge: "STATUS", Description: "Inspect CLI status and executable path."},
			{Value: "clean", Label: "Clean", Badge: "CLEAN", Description: "Remove caches and temporary files."},
			{Value: "exit", Label: "Exit", Badge: "EXIT", Description: "Leave the menu."},
		},
		panels: []Panel{
			{Title: "Environment", Lines: []string{"Executable: /tmp/agentflow", "CLI status:", "  [OK] codex"}},
		},
		width:  90,
		height: 12,
	}

	view := model.View()
	if lines := strings.Count(view, "\n") + 1; lines > model.height {
		t.Fatalf("expected view height <= %d, got %d", model.height, lines)
	}
}

func TestViewRendersPanelsInsideDetailsPane(t *testing.T) {
	model := selectionModel{
		catalog: i18n.NewCatalog(),
		title:   "AgentFlow",
		options: []Option{
			{Value: "status", Label: "Status", Badge: "STATUS", Description: "Inspect status."},
		},
		panels: []Panel{
			{Title: "Environment", Lines: []string{"Executable: /tmp/agentflow", "CLI status:", "  [OK] codex"}},
		},
		width:  100,
		height: 18,
	}

	view := model.View()
	for _, needle := range []string{"Environment", "Executable: /tmp/agentflow", "CLI status:", "[OK] codex"} {
		if !strings.Contains(view, needle) {
			t.Fatalf("expected %q in view, got %q", needle, view)
		}
	}
}
