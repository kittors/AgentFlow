package ui

import (
	"strings"
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
	if model.screen != flowScreenInstallScope {
		t.Fatalf("expected install scope screen after selecting install-agentflow, got %v", model.screen)
	}

	model.installScopeCursor = 0 // Select Global install
	next, _ = model.handleEnter()
	model = next.(interactiveFlowModel)
	if model.screen != flowScreenProfile {
		t.Fatalf("expected profile screen after selecting global install, got %v", model.screen)
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
	if model.screen != flowScreenInstallScope {
		t.Fatalf("expected second Esc to return to install scope, got %v", model.screen)
	}

	next, _ = model.handleBack()
	model = next.(interactiveFlowModel)
	if model.screen != flowScreenInstallHub {
		t.Fatalf("expected third Esc to return to install hub, got %v", model.screen)
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

func TestFlowSelectingMCPTargetAutoLoadsList(t *testing.T) {
	model := newTestInteractiveFlowModel(InteractiveCallbacks{
		Status: func() Panel { return Panel{Title: "Environment"} },
		MCPList: func(target string) Panel {
			return Panel{Title: "MCP list", Lines: []string{"target=" + target}}
		},
		MCPRemoveOptions: func(target string) []Option { return nil },
	})
	model.screen = flowScreenMCPTargets
	model.mcpTargets = []Option{{Value: "codex", Label: "Codex CLI", Badge: "CODEX"}}
	model.mcpTargetCursor = 0

	next, cmd := model.handleEnter()
	model = next.(interactiveFlowModel)
	if !model.busy {
		t.Fatal("expected selecting mcp target to enter busy state")
	}
	if cmd == nil {
		t.Fatal("expected selecting mcp target to return a command")
	}

	next, _ = model.Update(flowResultMsg{
		action: flowActionMCPList,
		notice: panelRef(Panel{Title: "MCP list", Lines: []string{"target=codex"}}),
		status: model.callbacks.Status(),
	})
	model = next.(interactiveFlowModel)
	if model.screen != flowScreenMCPActions {
		t.Fatalf("expected to move to MCP actions screen, got %v", model.screen)
	}
	if model.notice == nil || model.notice.Title != "MCP list" {
		t.Fatalf("expected list notice panel, got %#v", model.notice)
	}
}

func TestFlowSelectingSkillTargetMovesToScope(t *testing.T) {
	model := newTestInteractiveFlowModel(InteractiveCallbacks{
		Status: func() Panel { return Panel{Title: "Environment"} },
	})
	model.screen = flowScreenSkillTargets
	model.skillTargets = []Option{{Value: "codex", Label: "Codex CLI", Badge: "CODEX"}}
	model.skillTargetCursor = 0

	next, cmd := model.handleEnter()
	model = next.(interactiveFlowModel)
	if cmd != nil {
		t.Fatalf("expected selecting skill target not to return a command, got %#v", cmd)
	}
	if model.screen != flowScreenSkillScope {
		t.Fatalf("expected to move to skill scope screen, got %v", model.screen)
	}
	if len(model.skillScopeOptions) != 2 {
		t.Fatalf("expected 2 scope options, got %#v", model.skillScopeOptions)
	}
}

func TestFlowMCPInstallOptionsAnnotateInstalledAndRecommended(t *testing.T) {
	model := newTestInteractiveFlowModel(InteractiveCallbacks{
		Status: func() Panel { return Panel{Title: "Environment"} },
		MCPInstallOptions: func() []Option {
			return []Option{{Value: "context7", Label: "context7", Badge: "PIN", Description: "Docs."}}
		},
		MCPRemoveOptions: func(target string) []Option {
			return []Option{{Value: "Context7", Label: "Context7", Badge: "DEL"}}
		},
	})
	model.screen = flowScreenMCPActions
	model.selectedMCPTarget = "codex"
	model.mcpActions = []Option{{Value: "install"}}
	model.mcpActionCursor = 0

	next, _ := model.handleEnter()
	model = next.(interactiveFlowModel)
	if model.screen != flowScreenMCPInstall {
		t.Fatalf("expected to enter MCP install screen, got %v", model.screen)
	}
	if len(model.mcpInstallOptions) != 1 {
		t.Fatalf("expected annotated install options, got %#v", model.mcpInstallOptions)
	}
	if model.mcpInstallOptions[0].Badge != "✓" {
		t.Fatalf("expected installed badge ✓, got %q", model.mcpInstallOptions[0].Badge)
	}
	if !strings.HasPrefix(model.mcpInstallOptions[0].Label, "★ ") {
		t.Fatalf("expected recommended prefix, got %q", model.mcpInstallOptions[0].Label)
	}
	if strings.Contains(model.mcpInstallOptions[0].Label, "PIN") || model.mcpInstallOptions[0].Badge == "PIN" {
		t.Fatalf("expected PIN not to be visible, got %#v", model.mcpInstallOptions[0])
	}
}

func TestFlowSkillInstallOptionsAnnotateInstalledAndRecommended(t *testing.T) {
	model := newTestInteractiveFlowModel(InteractiveCallbacks{
		Status: func() Panel { return Panel{Title: "Environment"} },
		SkillInstallOptions: func(target string) []Option {
			return []Option{{Value: "https://skills.sh/vercel/turborepo/turborepo", Label: "turborepo", Badge: "PIN", Description: "Turborepo."}}
		},
		SkillUninstallOptions: func(target string) []Option {
			return []Option{{Value: "turborepo", Label: "turborepo", Badge: "DEL"}}
		},
	})
	model.screen = flowScreenSkillActions
	model.selectedSkillTarget = "codex"
	model.skillActions = []Option{{Value: "install"}}
	model.skillActionCursor = 0

	next, cmd := model.handleEnter()
	model = next.(interactiveFlowModel)
	if !model.busy {
		t.Fatal("expected install action to enter busy state")
	}
	if cmd == nil {
		t.Fatal("expected install action to return a command")
	}

	raw := model.callbacks.SkillInstallOptions(model.selectedSkillTarget)
	annotated := model.annotateRecommendedSkillOptions(model.selectedSkillTarget, raw)
	next, _ = model.Update(flowResultMsg{
		action:       flowActionSkillLoadInstallOptions,
		status:       model.callbacks.Status(),
		skillOptions: annotated,
	})
	model = next.(interactiveFlowModel)

	if model.screen != flowScreenSkillInstall {
		t.Fatalf("expected to enter skill install screen, got %v", model.screen)
	}
	if len(model.skillInstallOptions) != 1 {
		t.Fatalf("expected annotated install options, got %#v", model.skillInstallOptions)
	}
	if model.skillInstallOptions[0].Badge != "✓" {
		t.Fatalf("expected installed badge ✓, got %q", model.skillInstallOptions[0].Badge)
	}
	if !strings.HasPrefix(model.skillInstallOptions[0].Label, "★ ") {
		t.Fatalf("expected recommended prefix, got %q", model.skillInstallOptions[0].Label)
	}
	if strings.Contains(model.skillInstallOptions[0].Label, "PIN") || model.skillInstallOptions[0].Badge == "PIN" {
		t.Fatalf("expected PIN not to be visible, got %#v", model.skillInstallOptions[0])
	}
}

func TestFlowUpdateActionRefreshesVersionAndNotice(t *testing.T) {
	model := newTestInteractiveFlowModel(InteractiveCallbacks{
		Status: func() Panel {
			return Panel{Title: "Environment", Lines: []string{"AgentFlow v1.0.4-main.deadbee"}}
		},
		Update: func(progress func(stage string, percent int, info string)) (Panel, string) {
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
	notice, version := model.callbacks.Update(func(string, int, string) {})
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
	if model.screen != flowScreenUpdateConfirm {
		t.Fatalf("expected update confirm screen after successful update, got %v", model.screen)
	}
	if len(model.updateConfirmOptions) != 2 {
		t.Fatalf("expected 2 confirm options (restart/cancel), got %d", len(model.updateConfirmOptions))
	}

	// Selecting "cancel" should return to main.
	model.updateConfirmCursor = 1
	next, _ = model.handleEnter()
	model = next.(interactiveFlowModel)
	if model.screen != flowScreenMain {
		t.Fatalf("expected cancel to return to main, got %v", model.screen)
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
			return []Option{{Value: "claude", Label: "Claude Code", Badge: "CLAUDE"}}
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
		callbacks.Update = func(progress func(stage string, percent int, info string)) (Panel, string) {
			return Panel{Title: "Update result"}, ""
		}
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
