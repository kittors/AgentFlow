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

func TestViewShowsCurrentSummaryAndBadges(t *testing.T) {
	model := selectionModel{
		catalog:  i18n.NewCatalog(),
		title:    "AgentFlow v1.0.3",
		subtitle: "Cross-platform Go executable.",
		options: []Option{
			{Value: "install", Label: "Install", Description: "Write AgentFlow into CLI configs."},
			{Value: "update", Label: "Update AgentFlow", Description: "Download and replace the current Go binary."},
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
	if !strings.Contains(view, "2/2") {
		t.Fatalf("expected cursor badge in view, got %q", view)
	}
	if !strings.Contains(view, model.currentSummary()) {
		t.Fatalf("expected current summary in view, got %q", view)
	}
}
