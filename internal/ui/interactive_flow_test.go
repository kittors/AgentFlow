package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/kittors/AgentFlow/internal/i18n"
)

func TestFlowStatusActionStaysInMainScreen(t *testing.T) {
	model := newTestInteractiveFlowModel(InteractiveCallbacks{
		Status: func() Panel {
			return Panel{Title: "Environment", Lines: []string{"CLI status:", "  [OK] codex"}}
		},
	})
	model.mainCursor = 3

	next, cmd := model.handleEnter()
	model = next.(interactiveFlowModel)
	if !model.busy {
		t.Fatal("expected model to enter busy state for status refresh")
	}
	if cmd == nil {
		t.Fatal("expected status action to return a command")
	}

	next, _ = model.Update(flowResultMsg{
		action: flowActionRefreshStatus,
		status: model.callbacks.Status(),
	})
	model = next.(interactiveFlowModel)
	if model.busy {
		t.Fatal("expected busy state to clear after status refresh")
	}
	if model.screen != flowScreenMain {
		t.Fatalf("expected to stay on main screen, got %v", model.screen)
	}
	if model.status.Title != "Environment" {
		t.Fatalf("expected refreshed status panel, got %#v", model.status)
	}
}

func TestFlowEscReturnsSingleLevelFromInstallTargets(t *testing.T) {
	model := newTestInteractiveFlowModel(InteractiveCallbacks{
		Status: func() Panel { return Panel{Title: "Environment"} },
		InstallOptions: func() []Option {
			return []Option{{Value: "codex", Label: "codex", Badge: "CODEX"}}
		},
	})

	next, _ := model.handleEnter()
	model = next.(interactiveFlowModel)
	if model.screen != flowScreenInstallHub {
		t.Fatalf("expected install hub after selecting install, got %v", model.screen)
	}

	model.installHubCursor = 1
	next, _ = model.handleEnter()
	model = next.(interactiveFlowModel)
	if model.screen != flowScreenProfile {
		t.Fatalf("expected profile screen after selecting install-agentflow, got %v", model.screen)
	}

	next, _ = model.handleEnter()
	model = next.(interactiveFlowModel)
	if model.screen != flowScreenInstallTargets {
		t.Fatalf("expected install target screen, got %v", model.screen)
	}

	next, _ = model.handleBack()
	model = next.(interactiveFlowModel)
	if model.screen != flowScreenProfile {
		t.Fatalf("expected single Esc to return to profile, got %v", model.screen)
	}

	next, _ = model.handleBack()
	model = next.(interactiveFlowModel)
	if model.screen != flowScreenInstallHub {
		t.Fatalf("expected second Esc to return to install hub, got %v", model.screen)
	}

	next, _ = model.handleBack()
	model = next.(interactiveFlowModel)
	if model.screen != flowScreenMain {
		t.Fatalf("expected third Esc to return to main, got %v", model.screen)
	}
}

func TestFlowBootstrapBranchNavigatesAndReturnsSingleLevel(t *testing.T) {
	model := newTestInteractiveFlowModel(InteractiveCallbacks{
		Status: func() Panel { return Panel{Title: "Environment"} },
		BootstrapOptions: func() []Option {
			return []Option{{Value: "codex", Label: "Codex CLI", Badge: "CODEX"}}
		},
		BootstrapDetails: func(target string) Panel {
			return Panel{Title: "CLI details", Lines: []string{"target=" + target}}
		},
	})

	next, _ := model.handleEnter()
	model = next.(interactiveFlowModel)
	if model.screen != flowScreenInstallHub {
		t.Fatalf("expected install hub, got %v", model.screen)
	}

	next, _ = model.handleEnter()
	model = next.(interactiveFlowModel)
	if model.screen != flowScreenBootstrapTargets {
		t.Fatalf("expected bootstrap targets screen, got %v", model.screen)
	}
	if model.bootstrapDetail == nil || model.bootstrapDetail.Title != "CLI details" {
		t.Fatalf("expected bootstrap detail panel, got %#v", model.bootstrapDetail)
	}

	next, _ = model.handleEnter()
	model = next.(interactiveFlowModel)
	if model.screen != flowScreenBootstrapActions {
		t.Fatalf("expected bootstrap actions screen, got %v", model.screen)
	}

	next, _ = model.handleBack()
	model = next.(interactiveFlowModel)
	if model.screen != flowScreenBootstrapTargets {
		t.Fatalf("expected Esc to return to bootstrap targets, got %v", model.screen)
	}

	next, _ = model.handleBack()
	model = next.(interactiveFlowModel)
	if model.screen != flowScreenInstallHub {
		t.Fatalf("expected second Esc to return to install hub, got %v", model.screen)
	}
}

func TestFlowPrintableKeysDoNotTriggerHiddenShortcuts(t *testing.T) {
	model := newTestInteractiveFlowModel(InteractiveCallbacks{
		Status: func() Panel { return Panel{Title: "Environment"} },
	})
	model.mainCursor = 1

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	model = next.(interactiveFlowModel)
	if model.mainCursor != 1 {
		t.Fatalf("expected j not to move cursor, got %d", model.mainCursor)
	}

	next, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	model = next.(interactiveFlowModel)
	if model.mainCursor != 1 {
		t.Fatalf("expected q not to change cursor, got %d", model.mainCursor)
	}
	if model.screen != flowScreenMain {
		t.Fatalf("expected to remain on main screen, got %v", model.screen)
	}
}

func TestFlowArrowKeysMoveMainCursor(t *testing.T) {
	model := newTestInteractiveFlowModel(InteractiveCallbacks{
		Status: func() Panel { return Panel{Title: "Environment"} },
	})

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = next.(interactiveFlowModel)
	if model.mainCursor != 1 {
		t.Fatalf("expected down arrow to move cursor to 1, got %d", model.mainCursor)
	}

	next, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	model = next.(interactiveFlowModel)
	if model.mainCursor != 0 {
		t.Fatalf("expected up arrow to move cursor back to 0, got %d", model.mainCursor)
	}
}

func TestFlowMouseWheelMovesMainCursor(t *testing.T) {
	model := newTestInteractiveFlowModel(InteractiveCallbacks{
		Status: func() Panel { return Panel{Title: "Environment"} },
	})

	next, _ := model.Update(tea.MouseMsg{Button: tea.MouseButtonWheelDown, Action: tea.MouseActionPress, Type: tea.MouseWheelDown})
	model = next.(interactiveFlowModel)
	if model.mainCursor != 1 {
		t.Fatalf("expected mouse wheel down to move cursor to 1, got %d", model.mainCursor)
	}

	next, _ = model.Update(tea.MouseMsg{Button: tea.MouseButtonWheelUp, Action: tea.MouseActionPress, Type: tea.MouseWheelUp})
	model = next.(interactiveFlowModel)
	if model.mainCursor != 0 {
		t.Fatalf("expected mouse wheel up to move cursor back to 0, got %d", model.mainCursor)
	}
}

func TestFlowSelectionForCurrentScreenPropagatesDetailFocusAndScroll(t *testing.T) {
	model := newTestInteractiveFlowModel(InteractiveCallbacks{
		Status: func() Panel { return Panel{Title: "Environment"} },
	})
	model.focusDetails = true
	model.detailScroll = 7

	screen := model.selectionForCurrentScreen()
	if !screen.focusDetails {
		t.Fatal("expected selection model to receive focusDetails=true")
	}
	if screen.detailScroll != 7 {
		t.Fatalf("expected selection model to receive detailScroll=7, got %d", screen.detailScroll)
	}
}

func TestFlowDetailScrollUsesArrowKeysWhenDetailHasFocus(t *testing.T) {
	model := newTestInteractiveFlowModel(InteractiveCallbacks{
		Status: func() Panel { return Panel{Title: "Environment"} },
	})
	model.mainCursor = 2
	model.focusDetails = true
	model.detailScroll = 0

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = next.(interactiveFlowModel)
	if model.mainCursor != 2 {
		t.Fatalf("expected down arrow to keep cursor when detail focused, got %d", model.mainCursor)
	}
	if model.detailScroll != 1 {
		t.Fatalf("expected down arrow to scroll detail when focused, got %d", model.detailScroll)
	}
}

func TestFlowDetailScrollUsesMouseWheelWhenDetailHasFocus(t *testing.T) {
	model := newTestInteractiveFlowModel(InteractiveCallbacks{
		Status: func() Panel { return Panel{Title: "Environment"} },
	})
	model.mainCursor = 1
	model.focusDetails = true
	model.detailScroll = 0

	next, _ := model.Update(tea.MouseMsg{Button: tea.MouseButtonWheelDown, Action: tea.MouseActionPress, Type: tea.MouseWheelDown})
	model = next.(interactiveFlowModel)
	if model.mainCursor != 1 {
		t.Fatalf("expected mouse wheel not to move cursor when detail focused, got %d", model.mainCursor)
	}
	if model.detailScroll != 1 {
		t.Fatalf("expected mouse wheel down to scroll detail when focused, got %d", model.detailScroll)
	}
}

func TestFlowCursorChangeResetsDetailScroll(t *testing.T) {
	model := newTestInteractiveFlowModel(InteractiveCallbacks{
		Status: func() Panel { return Panel{Title: "Environment"} },
	})
	model.mainCursor = 0
	model.focusDetails = false
	model.detailScroll = 5

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = next.(interactiveFlowModel)
	if model.mainCursor != 1 {
		t.Fatalf("expected down arrow to move cursor to 1, got %d", model.mainCursor)
	}
	if model.detailScroll != 0 {
		t.Fatalf("expected cursor change to reset detail scroll, got %d", model.detailScroll)
	}
}

func TestFlowUpdateActionRefreshesVersionAndNotice(t *testing.T) {
	model := newTestInteractiveFlowModel(InteractiveCallbacks{
		Status: func() Panel {
			return Panel{Title: "Environment", Lines: []string{"AgentFlow v1.0.4-main.deadbee"}}
		},
		Update: func() (Panel, string) {
			return Panel{
				Title: "Update result",
				Lines: []string{"Updated to v1.0.4-main.deadbee."},
			}, "1.0.4-main.deadbee"
		},
	})
	model.mainCursor = 2

	next, cmd := model.handleEnter()
	model = next.(interactiveFlowModel)
	if cmd == nil {
		t.Fatal("expected update action to return a command")
	}
	notice, version := model.callbacks.Update()
	next, _ = model.Update(flowResultMsg{
		action:  flowActionUpdate,
		notice:  panelRef(notice),
		status:  model.callbacks.Status(),
		version: version,
	})
	model = next.(interactiveFlowModel)

	if model.version != "1.0.4-main.deadbee" {
		t.Fatalf("expected version to refresh, got %q", model.version)
	}
	if model.notice == nil || model.notice.Title != "Update result" {
		t.Fatalf("expected update notice panel, got %#v", model.notice)
	}
	if model.screen != flowScreenMain {
		t.Fatalf("expected to return to main after update, got %v", model.screen)
	}
}

func TestFlowUpdateActionShowsBusyPanelInsideMainView(t *testing.T) {
	model := newTestInteractiveFlowModel(InteractiveCallbacks{
		Status: func() Panel { return Panel{Title: "Environment"} },
	})
	model.mainCursor = 2

	next, cmd := model.handleEnter()
	model = next.(interactiveFlowModel)
	if !model.busy {
		t.Fatal("expected update action to enter busy state")
	}
	if cmd == nil {
		t.Fatal("expected update action to return a command")
	}

	screen := model.selectionForCurrentScreen()
	if len(screen.panels) == 0 {
		t.Fatal("expected busy screen to render panels")
	}
	if screen.panels[0].Title != model.catalog.Msg("处理中", "Working") {
		t.Fatalf("expected busy panel to stay inside main view, got %#v", screen.panels[0])
	}
	if model.screen != flowScreenMain {
		t.Fatalf("expected update busy state to remain on main screen, got %v", model.screen)
	}
}

func TestFlowCleanActionReturnsToMainWithNotice(t *testing.T) {
	model := newTestInteractiveFlowModel(InteractiveCallbacks{
		Status: func() Panel {
			return Panel{Title: "Environment", Lines: []string{"CLI status:", "  [OK] codex"}}
		},
		Clean: func() Panel {
			return Panel{Title: "Clean result", Lines: []string{"Cleaned 2 cache directories."}}
		},
	})
	model.mainCursor = 4

	next, cmd := model.handleEnter()
	model = next.(interactiveFlowModel)
	if !model.busy {
		t.Fatal("expected clean action to enter busy state")
	}
	if cmd == nil {
		t.Fatal("expected clean action to return a command")
	}

	next, _ = model.Update(flowResultMsg{
		action: flowActionClean,
		notice: panelRef(Panel{
			Title: "Clean result",
			Lines: []string{"Cleaned 2 cache directories."},
		}),
		status: model.callbacks.Status(),
	})
	model = next.(interactiveFlowModel)

	if model.busy {
		t.Fatal("expected busy state to clear after clean action")
	}
	if model.screen != flowScreenMain {
		t.Fatalf("expected clean action to return to main screen, got %v", model.screen)
	}
	if model.notice == nil || model.notice.Title != "Clean result" {
		t.Fatalf("expected clean result notice, got %#v", model.notice)
	}
}

func TestFlowEscReturnsSingleLevelFromUninstallTargets(t *testing.T) {
	model := newTestInteractiveFlowModel(InteractiveCallbacks{
		Status: func() Panel { return Panel{Title: "Environment"} },
		UninstallOptions: func() []Option {
			return []Option{{Value: "codex", Label: "codex", Badge: "CODEX"}}
		},
	})
	model.mainCursor = 1

	next, _ := model.handleEnter()
	model = next.(interactiveFlowModel)
	if model.screen != flowScreenUninstallTargets {
		t.Fatalf("expected uninstall target screen after selecting uninstall, got %v", model.screen)
	}

	next, _ = model.handleBack()
	model = next.(interactiveFlowModel)
	if model.screen != flowScreenMain {
		t.Fatalf("expected single Esc to return to main, got %v", model.screen)
	}
}

func TestFlowUninstallActionReturnsToMainWithNotice(t *testing.T) {
	model := newTestInteractiveFlowModel(InteractiveCallbacks{
		Status: func() Panel {
			return Panel{Title: "Environment", Lines: []string{"CLI status:", "  [OK] codex"}}
		},
		UninstallOptions: func() []Option {
			return []Option{{Value: "codex", Label: "codex", Badge: "CODEX", Selected: true}}
		},
		Uninstall: func(targets []string) Panel {
			return Panel{Title: "Uninstall result", Lines: []string{"[done] codex"}}
		},
	})
	model.mainCursor = 1

	next, _ := model.handleEnter()
	model = next.(interactiveFlowModel)
	if model.screen != flowScreenUninstallTargets {
		t.Fatalf("expected uninstall target screen, got %v", model.screen)
	}

	next, cmd := model.handleEnter()
	model = next.(interactiveFlowModel)
	if !model.busy {
		t.Fatal("expected uninstall action to enter busy state")
	}
	if cmd == nil {
		t.Fatal("expected uninstall action to return a command")
	}

	screen := model.selectionForCurrentScreen()
	if len(screen.panels) == 0 || screen.panels[0].Title != model.catalog.Msg("处理中", "Working") {
		t.Fatalf("expected working panel during uninstall, got %#v", screen.panels)
	}

	next, _ = model.Update(flowResultMsg{
		action: flowActionUninstall,
		notice: panelRef(Panel{
			Title: "Uninstall result",
			Lines: []string{"[done] codex"},
		}),
		status: model.callbacks.Status(),
	})
	model = next.(interactiveFlowModel)

	if model.busy {
		t.Fatal("expected busy state to clear after uninstall action")
	}
	if model.screen != flowScreenMain {
		t.Fatalf("expected uninstall action to return to main screen, got %v", model.screen)
	}
	if model.notice == nil || model.notice.Title != "Uninstall result" {
		t.Fatalf("expected uninstall result notice, got %#v", model.notice)
	}
}

func TestFlowBootstrapBranchDefaultsToManualWhenAutoUnsupported(t *testing.T) {
	model := newTestInteractiveFlowModel(InteractiveCallbacks{
		Status: func() Panel { return Panel{Title: "Environment"} },
		BootstrapOptions: func() []Option {
			return []Option{{Value: "gemini", Label: "Gemini CLI", Badge: "GEMINI"}}
		},
		BootstrapAutoSupported: func(target string) bool {
			return false
		},
	})

	next, _ := model.handleEnter()
	model = next.(interactiveFlowModel)
	next, _ = model.handleEnter()
	model = next.(interactiveFlowModel)
	next, _ = model.handleEnter()
	model = next.(interactiveFlowModel)

	if model.screen != flowScreenBootstrapActions {
		t.Fatalf("expected bootstrap actions screen, got %v", model.screen)
	}
	if model.bootstrapActionCursor != 1 {
		t.Fatalf("expected manual action to be preselected when auto install is unavailable, got %d", model.bootstrapActionCursor)
	}
}

func newTestInteractiveFlowModel(callbacks InteractiveCallbacks) interactiveFlowModel {
	catalog := i18n.NewCatalog()
	if callbacks.Status == nil {
		callbacks.Status = func() Panel { return Panel{Title: "Environment"} }
	}
	if callbacks.BootstrapOptions == nil {
		callbacks.BootstrapOptions = func() []Option { return nil }
	}
	if callbacks.BootstrapAutoSupported == nil {
		callbacks.BootstrapAutoSupported = func(target string) bool { return true }
	}
	if callbacks.BootstrapDetails == nil {
		callbacks.BootstrapDetails = func(target string) Panel { return Panel{Title: "CLI details"} }
	}
	if callbacks.BootstrapAuto == nil {
		callbacks.BootstrapAuto = func(target string) Panel { return Panel{Title: "CLI install result"} }
	}
	if callbacks.BootstrapManual == nil {
		callbacks.BootstrapManual = func(target string) Panel { return Panel{Title: "Manual install guidance"} }
	}
	if callbacks.InstallOptions == nil {
		callbacks.InstallOptions = func() []Option { return nil }
	}
	if callbacks.UninstallOptions == nil {
		callbacks.UninstallOptions = func() []Option { return nil }
	}
	if callbacks.Install == nil {
		callbacks.Install = func(profile string, targets []string) Panel { return Panel{Title: "Install result"} }
	}
	if callbacks.Uninstall == nil {
		callbacks.Uninstall = func(targets []string) Panel { return Panel{Title: "Uninstall result"} }
	}
	if callbacks.Update == nil {
		callbacks.Update = func() (Panel, string) { return Panel{Title: "Update result"}, "" }
	}
	if callbacks.Clean == nil {
		callbacks.Clean = func() Panel { return Panel{Title: "Clean result"} }
	}

	return interactiveFlowModel{
		catalog:   catalog,
		version:   "1.0.3-main.test",
		callbacks: callbacks,
		screen:    flowScreenMain,
		status:    Panel{Title: "Environment", Lines: []string{"Loading status..."}},
		mainOptions: []Option{
			{Value: string(ActionInstall), Label: "Install", Badge: "SETUP"},
			{Value: string(ActionUninstall), Label: "Uninstall", Badge: "REMOVE"},
			{Value: string(ActionUpdate), Label: "Update", Badge: "UPDATE"},
			{Value: string(ActionStatus), Label: "Status", Badge: "STATUS"},
			{Value: string(ActionClean), Label: "Clean", Badge: "CLEAN"},
			{Value: string(ActionExit), Label: "Exit", Badge: "EXIT"},
		},
		installHubOptions: []Option{
			{Value: "bootstrap-cli", Label: "Install CLI tools", Badge: "CLI"},
			{Value: "install-agentflow", Label: "Install AgentFlow into existing CLIs", Badge: "APPLY"},
		},
		bootstrapActionOptions: []Option{
			{Value: "auto", Label: "Automatic install", Badge: "AUTO"},
			{Value: "manual", Label: "Show manual install guidance", Badge: "MANUAL"},
		},
		profileOptions: []Option{
			{Value: "lite", Label: "lite", Badge: "LITE"},
			{Value: "standard", Label: "standard", Badge: "STANDARD"},
			{Value: "full", Label: "full", Badge: "FULL"},
		},
		profileCursor: 2,
	}
}
