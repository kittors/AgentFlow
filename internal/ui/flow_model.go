package ui

import (
	"sync"

	"github.com/kittors/AgentFlow/internal/i18n"
)

type InteractiveCallbacks struct {
	Status                  func() Panel
	CLIDetailPanel          func(target string) Panel
	CLIInstalled            func(target string) bool
	MCPTargetOptions        func() []Option
	MCPInstallOptions       func() []Option
	MCPRemoveOptions        func(target string) []Option
	MCPList                 func(target string) Panel
	MCPInstall              func(target, server string) Panel
	MCPInstallWithEnv       func(target, server string, env map[string]string) Panel
	MCPBatchInstall         func(target string, servers []string) Panel
	MCPConfigFields         func(server string) []ConfigField
	MCPRemove               func(target, server string) Panel
	SkillTargetOptions      func() []Option
	SkillGlobalSupported    func(target string) bool
	SkillInstallOptions     func(target string) []Option
	SkillUninstallOptions   func(target string) []Option
	SkillList               func(target string) Panel
	SkillInstall            func(target, source string) Panel
	SkillUninstall          func(target, name string) Panel
	ProjectRulesPanel       func(root, target string) Panel
	ProjectRulesInstall     func(root, target, profile string) Panel
	ProjectRulesUninstall   func(root, target string) Panel
	UninstallProjectOptions func() []Option
	BootstrapOptions        func() []Option
	BootstrapAutoSupported  func(target string) bool
	BootstrapDetails        func(target string) Panel
	BootstrapAuto           func(target string) Panel
	BootstrapManual         func(target string) Panel
	InstallOptions          func() []Option
	UninstallOptions        func() []Option
	UninstallCLIOptions     func() []Option
	Install                 func(profile string, targets []string) Panel
	Uninstall               func(targets []string) Panel
	UninstallCLI            func(targets []string) Panel
	Update                  func(progress func(stage string, percent int, info string)) (Panel, string)
	Clean                   func() Panel
	CLIConfigFields         func(target string) []ConfigField
	WriteEnvConfig          func(envVars map[string]string) Panel
}

// ConfigField describes a single configurable field for a CLI.
type ConfigField struct {
	Label   string   // e.g. "API Key"
	EnvVar  string   // e.g. "OPENAI_API_KEY" or "__MODEL__" for non-env fields
	Type    string   // "text" for free-form input, "select" for option list
	Options []string // For "select" type: available choices
	Default string   // Default value (pre-selected for "select")
}

type flowScreen int

const (
	flowScreenMain flowScreen = iota
	flowScreenToolbox
	flowScreenCLI
	flowScreenAgentFlow
	flowScreenInstallHub
	flowScreenMCPTargets
	flowScreenMCPActions
	flowScreenMCPInstall
	flowScreenMCPList
	flowScreenMCPRemove
	flowScreenSkillTargets
	flowScreenSkillScope
	flowScreenSkillProjectActions
	flowScreenSkillProjectProfile
	flowScreenSkillActions
	flowScreenSkillInstall
	flowScreenSkillUninstall
	flowScreenBootstrapTargets
	flowScreenBootstrapActions
	flowScreenProfile
	flowScreenInstallScope
	flowScreenInstallTargets
	flowScreenUninstallTargets
	flowScreenUpdateConfirm
	flowScreenBootstrapConfig
)

type flowAction int

const (
	flowActionRefreshStatus flowAction = iota
	flowActionCLIRefreshDetail
	flowActionMCPRefreshSummary
	flowActionMCPList
	flowActionMCPInstall
	flowActionMCPRemove
	flowActionSkillRefreshSummary
	flowActionInstallHubRefresh
	flowActionSkillList
	flowActionSkillLoadInstallOptions
	flowActionSkillInstall
	flowActionSkillUninstall
	flowActionProjectRulesInstall
	flowActionProjectRulesUninstall
	flowActionUpdate
	flowActionClean
	flowActionBootstrapAuto
	flowActionInstall
	flowActionUninstall
	flowActionUninstallProject
	flowActionUninstallCLI
	flowActionWriteEnvConfig
	flowActionMCPInstallWithEnv
	flowActionMCPBatchInstall
)

type flowResultMsg struct {
	action          flowAction
	notice          *Panel
	status          Panel
	version         string
	bootstrapDetail *Panel
	projectRules    *Panel
	mcpSummary      *Panel
	skillSummary    *Panel
	skillOptions    []Option
	cliDetail       *Panel
}

type flowTickMsg struct{}
type flowToastClearMsg struct{}

// updateProgressState holds thread-safe progress info that is written by the
// update goroutine and polled by the busy tick.
type updateProgressState struct {
	mu      sync.Mutex
	stage   string // e.g. "checking", "found:1.2.3", "downloading:1.2.3"
	percent int    // 0–100 for download, -1 for indeterminate
}

type interactiveFlowModel struct {
	catalog   i18n.Catalog
	version   string
	callbacks InteractiveCallbacks

	width  int
	height int

	screen         flowScreen
	busy           bool
	spin           int
	activeAction   flowAction // which action owns the current busy state
	initLoading    bool       // true while Init's refreshStatusCmd is in flight
	focusDetails   bool
	detailScroll   int
	updateProgress *updateProgressState // shared progress polled by tick

	mainOptions            []Option
	toolboxOptions         []Option
	toolboxCursor          int
	cliOptions             []Option
	cliCursor              int
	agentflowOptions       []Option
	agentflowCursor        int
	cliDetail              *Panel
	installHubOptions      []Option
	projectRoot            string
	mcpTargets             []Option
	mcpActions             []Option
	mcpInstallOptions      []Option
	mcpRemoveOptions       []Option
	skillTargets           []Option
	skillScopeOptions      []Option
	skillProjectActions    []Option
	skillActions           []Option
	skillInstallOptions    []Option
	skillUninstallOptions  []Option
	bootstrapOptions       []Option
	bootstrapActionOptions []Option
	profileOptions         []Option
	installOptions         []Option
	uninstallOptions       []Option
	uninstallCLIMode       bool
	enteredFromCLI         bool // true when BootstrapActions was entered from CLI screen (not BootstrapTargets)
	mcpConfigMode          bool // true when config fields are for MCP install (not bootstrap)
	projectInstallMode     bool
	updateConfirmOptions   []Option
	updateConfirmCursor    int
	installHubStatusPanel  *Panel

	// Bootstrap config screen state.
	configFields        []configFieldState
	pendingConfigFields []ConfigField // cached after bootstrap; shown when user chooses "configure"
	configFieldCursor   int
	configEditing       bool
	configTarget        string

	mainCursor               int
	installHubCursor         int
	mcpTargetCursor          int
	mcpActionCursor          int
	mcpInstallCursor         int
	mcpRemoveCursor          int
	mcpListCursor            int
	skillTargetCursor        int
	skillScopeCursor         int
	skillProjectActionCursor int
	skillActionCursor        int
	skillInstallCursor       int
	skillUninstallCursor     int
	bootstrapCursor          int
	bootstrapActionCursor    int
	profileCursor            int
	installScopeCursor       int
	installCursor            int
	uninstallCursor          int

	selectedProfile         string
	selectedMCPTarget       string
	selectedMCPServer       string
	pendingMCPInstalls      []string // servers to batch-install (excluding tavily-custom)
	mcpListOptions          []Option // installed MCPs shown as options in list view
	selectedSkillTarget     string
	selectedSkillScope      string
	selectedProjectProfile  string
	selectedSkillValue      string
	selectedBootstrapTarget string
	notice                  *Panel
	toast                   string // transient toast text, auto-clears after 1.5s
	status                  Panel
	bootstrapDetail         *Panel
	projectRulesDetail      *Panel
	mcpSummary              *Panel
	skillSummary            *Panel
}

func (m *interactiveFlowModel) resetDetailFocus() {
	m.focusDetails = false
	m.detailScroll = 0
}

// configFieldState holds the editing state for a single config field.
type configFieldState struct {
	Label        string   // e.g. "API Key"
	EnvVar       string   // e.g. "OPENAI_API_KEY"
	Value        string   // current typed value (for text type)
	FieldType    string   // "text" or "select"
	Options      []string // available choices (for select type)
	OptionCursor int      // currently selected option index (for select type)
	Dirty        bool     // true if user explicitly modified this field
	CursorPos    int      // cursor position within Value (for text type)
}
