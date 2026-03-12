package ui

import (
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
