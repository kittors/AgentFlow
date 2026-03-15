package install

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/kittors/AgentFlow/internal/debuglog"
	"github.com/kittors/AgentFlow/internal/targets"
)

const nvmInstallScriptURL = "https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.3/install.sh"
const kiroInstallScriptURL = "https://cli.kiro.dev/install"

var (
	runtimeGOOS = runtime.GOOS
	lookPath    = exec.LookPath
	runCombined = func(name string, args ...string) ([]byte, error) {
		cmd := exec.Command(name, args...)
		return cmd.CombinedOutput()
	}
)

type PlatformKind string

const (
	platformDarwin  PlatformKind = "darwin"
	platformLinux   PlatformKind = "linux"
	platformWindows PlatformKind = "windows"
	platformUnknown PlatformKind = "unknown"
)

type RuntimeStatus struct {
	Platform    PlatformKind
	InWSL       bool
	WSLReady    bool
	WSLDistro   string
	BashPath    string
	NodePath    string
	NodeVersion string
	NPMPath     string
	NPMVersion  string
	NVMReady    bool
}

type TargetStatus struct {
	Target               targets.Target
	CLIInstalled         bool
	CLIPath              string
	CLIPathScope         string
	ConfigDir            string
	ConfigDirExists      bool
	AgentFlowInstalled   bool
	BootstrapSupported   bool
	AutoInstallSupported bool
	Runtime              RuntimeStatus
	Notes                []string
}

func currentPlatform() PlatformKind {
	switch runtimeGOOS {
	case "darwin":
		return platformDarwin
	case "linux":
		return platformLinux
	case "windows":
		return platformWindows
	default:
		return platformUnknown
	}
}

func inWSL() bool {
	return strings.TrimSpace(os.Getenv("WSL_DISTRO_NAME")) != "" || strings.TrimSpace(os.Getenv("WSL_INTEROP")) != ""
}

func (i *Installer) RuntimeStatus() RuntimeStatus {
	done := debuglog.Timed("RuntimeStatus")
	defer done()
	status := RuntimeStatus{
		Platform: currentPlatform(),
		InWSL:    inWSL(),
	}

	if status.Platform == platformWindows {
		if bashPath, err := lookPath("bash"); err == nil {
			status.BashPath = bashPath
		}
		if distro, ready := detectWSL(); ready {
			status.WSLReady = true
			status.WSLDistro = distro
			status.NodePath = strings.TrimSpace(i.wslCommandPath("node"))
			status.NodeVersion = strings.TrimSpace(i.wslCommandValue("node --version"))
			status.NPMPath = strings.TrimSpace(i.wslCommandPath("npm"))
			status.NPMVersion = strings.TrimSpace(i.wslCommandValue("npm --version"))
			status.NVMReady = strings.TrimSpace(i.wslCommandValue(`if command -v nvm >/dev/null 2>&1; then printf yes; fi`)) == "yes"
		}
		return status
	}

	if bashPath, err := lookPath("bash"); err == nil {
		status.BashPath = bashPath
		status.NodePath = strings.TrimSpace(i.nativeCommandPath("node"))
		status.NodeVersion = strings.TrimSpace(i.nativeCommandValue("node --version"))
		status.NPMPath = strings.TrimSpace(i.nativeCommandPath("npm"))
		status.NPMVersion = strings.TrimSpace(i.nativeCommandValue("npm --version"))
		status.NVMReady = strings.TrimSpace(i.nativeCommandValue(`if command -v nvm >/dev/null 2>&1; then printf yes; fi`)) == "yes"
	}

	return status
}

func detectWSL() (string, bool) {
	if _, err := lookPath("wsl.exe"); err != nil {
		return "", false
	}
	output, err := runCombined("wsl.exe", "-l", "-q")
	if err != nil {
		return "", false
	}
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(strings.TrimPrefix(line, "\x00"))
		if line != "" {
			return line, true
		}
	}
	return "", false
}

func (i *Installer) DetectTargetStatuses() []TargetStatus {
	done := debuglog.Timed("DetectTargetStatuses")
	defer done()
	runtimeStatus := i.CachedRuntimeStatus()
	agentflowInstalled := sliceToStatusSet(i.DetectInstalledTargets())

	names := targets.Names()
	statuses := make([]TargetStatus, len(names))
	var wg sync.WaitGroup
	for idx, name := range names {
		wg.Add(1)
		go func(idx int, name string) {
			defer wg.Done()
			target := targets.All[name]
			status := i.detectTargetStatusWith(target, runtimeStatus)
			status.AgentFlowInstalled = agentflowInstalled[name]
			statuses[idx] = status
		}(idx, name)
	}
	wg.Wait()
	return statuses
}

func (i *Installer) DetectBootstrapTargetStatuses() []TargetStatus {
	statuses := make([]TargetStatus, 0, len(targets.BootstrapNames()))
	allStatuses := make(map[string]TargetStatus, len(targets.Names()))
	for _, status := range i.DetectTargetStatuses() {
		allStatuses[status.Target.Name] = status
	}
	for _, name := range targets.BootstrapNames() {
		if status, ok := allStatuses[name]; ok {
			statuses = append(statuses, status)
		}
	}
	return statuses
}

func (i *Installer) DetectTargetStatus(name string) (TargetStatus, error) {
	target, ok := targets.Lookup(name)
	if !ok {
		return TargetStatus{}, fmt.Errorf(i.Catalog.Msg("未知目标: %s", "unknown target: %s"), name)
	}
	status := i.detectTargetStatus(target)
	for _, installedName := range i.DetectInstalledTargets() {
		if installedName == name {
			status.AgentFlowInstalled = true
			break
		}
	}
	return status, nil
}

func (i *Installer) detectTargetStatus(target targets.Target) TargetStatus {
	return i.detectTargetStatusWith(target, i.CachedRuntimeStatus())
}

func (i *Installer) detectTargetStatusWith(target targets.Target, runtimeStatus RuntimeStatus) TargetStatus {
	configDir := filepath.Join(i.HomeDir, target.Dir)
	status := TargetStatus{
		Target:             target,
		ConfigDir:          configDir,
		BootstrapSupported: target.BootstrapSupported,
		Runtime:            runtimeStatus,
	}
	if info, err := os.Stat(configDir); err == nil && info.IsDir() {
		status.ConfigDirExists = true
	}

	status.CLIPath, status.CLIPathScope = i.detectCLIPath(target, runtimeStatus)
	status.CLIInstalled = strings.TrimSpace(status.CLIPath) != ""
	status.AutoInstallSupported = status.BootstrapSupported && i.autoInstallSupported(target, runtimeStatus)
	status.Notes = i.targetNotes(target, status)
	return status
}

func (i *Installer) autoInstallSupported(target targets.Target, runtimeStatus RuntimeStatus) bool {
	if !target.BootstrapSupported {
		return false
	}

	if target.Name == "kiro" {
		switch runtimeStatus.Platform {
		case platformDarwin, platformLinux:
			return runtimeStatus.BashPath != "" && (i.hasNativeCommand("curl") || i.hasNativeCommand("wget"))
		case platformWindows:
			if runtimeStatus.InWSL {
				return runtimeStatus.BashPath != "" && (i.hasNativeCommand("curl") || i.hasNativeCommand("wget"))
			}
			if !runtimeStatus.WSLReady {
				return false
			}
			return i.hasWSLCommand("curl") || i.hasWSLCommand("wget")
		default:
			return false
		}
	}

	switch runtimeStatus.Platform {
	case platformDarwin, platformLinux:
		return runtimeStatus.BashPath != ""
	case platformWindows:
		if runtimeStatus.InWSL {
			return runtimeStatus.BashPath != ""
		}
		return runtimeStatus.WSLReady
	default:
		return false
	}
}

func (i *Installer) detectCLIPath(target targets.Target, runtimeStatus RuntimeStatus) (string, string) {
	done := debuglog.Timed("detectCLIPath(" + target.Name + ")")
	defer done()
	if target.Command == "" {
		if info, err := os.Stat(filepath.Join(i.HomeDir, target.Dir)); err == nil && info.IsDir() {
			return filepath.Join(i.HomeDir, target.Dir), "config"
		}
		return "", ""
	}

	switch runtimeStatus.Platform {
	case platformWindows:
		if path, err := lookPath(target.Command); err == nil {
			return path, "host"
		}
		if runtimeStatus.WSLReady {
			if path := strings.TrimSpace(i.wslCommandPath(target.Command)); path != "" {
				return path, "wsl"
			}
		}
	default:
		if path := strings.TrimSpace(i.nativeCommandPath(target.Command)); path != "" {
			return path, "shell"
		}
		if path, err := lookPath(target.Command); err == nil {
			return path, "host"
		}
	}
	return "", ""
}

func (i *Installer) BootstrapCLI(targetName string) ([]string, error) {
	target, ok := targets.Lookup(targetName)
	if !ok {
		return nil, fmt.Errorf(i.Catalog.Msg("未知目标: %s", "unknown target: %s"), targetName)
	}
	if !target.BootstrapSupported {
		return nil, fmt.Errorf(i.Catalog.Msg("%s 当前还不支持自动安装。", "%s is not yet supported for automatic installation."), target.DisplayName)
	}

	runtimeStatus := i.RuntimeStatus()
	if !i.autoInstallSupported(target, runtimeStatus) {
		return nil, errors.New(strings.Join(i.manualInstallLines(target, runtimeStatus), "\n"))
	}

	var output []byte
	var err error
	if target.Name == "kiro" {
		output, err = i.runKiroBootstrapScript(runtimeStatus)
	} else {
		output, err = i.runBootstrapScript(target, runtimeStatus)
	}
	if err != nil {
		return failureLines(target, output), err
	}

	status := i.detectTargetStatus(target)
	lines := []string{
		fmt.Sprintf(i.Catalog.Msg("已准备 %s 的运行环境。", "Prepared the runtime environment for %s."), target.DisplayName),
	}
	if target.Name == "kiro" {
		lines = append(lines, fmt.Sprintf(i.Catalog.Msg("安装命令: curl -sSL %s | bash", "Install command: curl -sSL %s | bash"), kiroInstallScriptURL))
		lines = append(lines, i.Catalog.Msg("接下来可执行: kiro-cli login", "Next: run: kiro-cli login"))
	} else {
		lines = append(lines, fmt.Sprintf(i.Catalog.Msg("安装命令: npm install -g %s", "Install command: npm install -g %s"), target.NPMPackage))
	}
	if status.CLIInstalled {
		lines = append(lines, fmt.Sprintf(i.Catalog.Msg("CLI 可执行文件: %s", "CLI executable: %s"), status.CLIPath))
	}
	if status.Runtime.NodeVersion != "" {
		lines = append(lines, fmt.Sprintf(i.Catalog.Msg("Node: %s", "Node: %s"), status.Runtime.NodeVersion))
	}
	if status.Runtime.NPMVersion != "" {
		lines = append(lines, fmt.Sprintf(i.Catalog.Msg("npm: %s", "npm: %s"), status.Runtime.NPMVersion))
	}
	if runtimeStatus.Platform == platformWindows && !runtimeStatus.InWSL {
		lines = append(lines, i.Catalog.Msg("当前 Windows 通过 WSL2 执行安装；后续也建议在 WSL2 Bash 中使用这些 CLI。", "Windows installation was executed via WSL2; keep using these CLIs inside WSL2 Bash for the best experience."))
	}
	return lines, nil
}

func (i *Installer) ManualInstallLines(targetName string) ([]string, error) {
	target, ok := targets.Lookup(targetName)
	if !ok {
		return nil, fmt.Errorf(i.Catalog.Msg("未知目标: %s", "unknown target: %s"), targetName)
	}
	return i.manualInstallLines(target, i.RuntimeStatus()), nil
}

func (i *Installer) manualInstallLines(target targets.Target, runtimeStatus RuntimeStatus) []string {
	if target.Name == "kiro" {
		lines := []string{
			fmt.Sprintf(i.Catalog.Msg("%s 的推荐安装命令:", "Recommended install command for %s:"), target.DisplayName),
			fmt.Sprintf("curl -sSL %s | bash", kiroInstallScriptURL),
		}
		if target.DocsURL != "" {
			lines = append(lines, fmt.Sprintf(i.Catalog.Msg("官方文档: %s", "Official docs: %s"), target.DocsURL))
		}
		if runtimeStatus.Platform == platformWindows && !runtimeStatus.InWSL {
			lines = append(lines, "")
			lines = append(lines, i.Catalog.Msg("Windows 建议先开启 WSL2，并在 WSL2 Bash 中安装与使用 Kiro CLI。", "On Windows, enable WSL2 and install/use Kiro CLI inside WSL2 Bash."))
		}
		return lines
	}

	lines := []string{
		fmt.Sprintf(i.Catalog.Msg("%s 的推荐安装命令:", "Recommended install command for %s:"), target.DisplayName),
		fmt.Sprintf("npm install -g %s", target.NPMPackage),
	}
	if target.MinNodeMajor > 0 {
		lines = append(lines, fmt.Sprintf(i.Catalog.Msg("需要 Node.js %d+。", "Requires Node.js %d+."), target.MinNodeMajor))
	} else {
		lines = append(lines, i.Catalog.Msg("建议使用最新的 Node.js LTS。", "Use the latest Node.js LTS."))
	}
	if target.DocsURL != "" {
		lines = append(lines, fmt.Sprintf(i.Catalog.Msg("官方文档: %s", "Official docs: %s"), target.DocsURL))
	}
	if runtimeStatus.Platform == platformWindows && !runtimeStatus.InWSL {
		lines = append(lines, "")
		lines = append(lines, i.Catalog.Msg("Windows 建议先开启 WSL2。很多 CLI 依赖 Bash 和沙箱能力，在 WSL2 中更稳定。", "On Windows, enable WSL2 first. Many CLIs rely on Bash behavior and sandbox support, and work more reliably inside WSL2."))
		lines = append(lines, i.Catalog.Msg("在 WSL2 Bash 中安装顺序:", "Suggested sequence inside WSL2 Bash:"))
		lines = append(lines, fmt.Sprintf("curl -fsSL %s | bash", nvmInstallScriptURL))
		if target.MinNodeMajor > 0 {
			lines = append(lines, fmt.Sprintf("export NVM_DIR=\"$HOME/.nvm\" && . \"$NVM_DIR/nvm.sh\" && nvm install %d && nvm use %d", target.MinNodeMajor, target.MinNodeMajor))
		} else {
			lines = append(lines, "export NVM_DIR=\"$HOME/.nvm\" && . \"$NVM_DIR/nvm.sh\" && nvm install --lts && nvm use --lts")
		}
		lines = append(lines, fmt.Sprintf("npm install -g %s", target.NPMPackage))
		return lines
	}

	lines = append(lines, "")
	lines = append(lines, i.Catalog.Msg("如果缺少 nvm，可先执行:", "If nvm is missing, run this first:"))
	lines = append(lines, fmt.Sprintf("curl -fsSL %s | bash", nvmInstallScriptURL))
	if target.MinNodeMajor > 0 {
		lines = append(lines, fmt.Sprintf("export NVM_DIR=\"$HOME/.nvm\" && . \"$NVM_DIR/nvm.sh\" && nvm install %d && nvm use %d", target.MinNodeMajor, target.MinNodeMajor))
	} else {
		lines = append(lines, "export NVM_DIR=\"$HOME/.nvm\" && . \"$NVM_DIR/nvm.sh\" && nvm install --lts && nvm use --lts")
	}
	return lines
}

func (i *Installer) targetNotes(target targets.Target, status TargetStatus) []string {
	notes := make([]string, 0, 4)
	switch {
	case status.AgentFlowInstalled && status.CLIInstalled:
		notes = append(notes, i.Catalog.Msg("CLI 与 AgentFlow 都已就绪。", "Both the CLI and AgentFlow are ready."))
	case status.CLIInstalled:
		notes = append(notes, i.Catalog.Msg("CLI 已安装，可以继续安装 AgentFlow。", "The CLI is installed and ready for AgentFlow."))
	case status.AgentFlowInstalled:
		notes = append(notes, i.Catalog.Msg("AgentFlow 文件已存在，但未检测到 CLI 可执行文件。", "AgentFlow files exist, but the CLI executable was not detected."))
	default:
		notes = append(notes, i.Catalog.Msg("尚未检测到该 CLI，可先安装 CLI 工具。", "The CLI was not detected yet; install the CLI tool first."))
	}

	if target.RecommendWSLOnWindows && status.Runtime.Platform == platformWindows && !status.Runtime.InWSL {
		if status.Runtime.WSLReady {
			notes = append(notes, i.Catalog.Msg("Windows 已检测到 WSL2，自动安装会优先在 WSL2 中执行。", "WSL2 is available on Windows; automatic installation will run inside WSL2."))
		} else {
			notes = append(notes, i.Catalog.Msg("Windows 建议先安装/启用 WSL2，再继续安装这些 CLI。", "On Windows, install or enable WSL2 before installing these CLIs."))
		}
	}
	if target.MinNodeMajor > 0 {
		notes = append(notes, fmt.Sprintf(i.Catalog.Msg("Node 要求: %d+", "Node requirement: %d+"), target.MinNodeMajor))
	}
	return notes
}

func (i *Installer) nativeCommandPath(command string) string {
	return i.nativeCommandValue(fmt.Sprintf("command -v %s 2>/dev/null || true", shellLiteral(command)))
}

func (i *Installer) nativeCommandValue(script string) string {
	if _, err := lookPath("bash"); err != nil {
		return ""
	}
	fullScript := `export NVM_DIR="$HOME/.nvm"
if [ -s "$NVM_DIR/nvm.sh" ]; then . "$NVM_DIR/nvm.sh" >/dev/null 2>&1; fi
` + script
	output, err := runCombined("bash", "-lc", fullScript)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func (i *Installer) wslCommandPath(command string) string {
	return i.wslCommandValue(fmt.Sprintf("command -v %s 2>/dev/null || true", shellLiteral(command)))
}

func (i *Installer) wslCommandValue(script string) string {
	if _, err := lookPath("wsl.exe"); err != nil {
		return ""
	}
	fullScript := `export NVM_DIR="$HOME/.nvm"
if [ -s "$NVM_DIR/nvm.sh" ]; then . "$NVM_DIR/nvm.sh" >/dev/null 2>&1; fi
` + script
	output, err := runCombined("wsl.exe", "bash", "-lc", fullScript)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func (i *Installer) runBootstrapScript(target targets.Target, runtimeStatus RuntimeStatus) ([]byte, error) {
	nodeInstall := "nvm install --lts && nvm use --lts"
	if target.MinNodeMajor > 0 {
		nodeInstall = fmt.Sprintf("nvm install %d && nvm use %d", target.MinNodeMajor, target.MinNodeMajor)
	}
	script := fmt.Sprintf(`set -e
export NVM_DIR="$HOME/.nvm"
if [ ! -s "$NVM_DIR/nvm.sh" ]; then
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL %s | bash
  elif command -v wget >/dev/null 2>&1; then
    wget -qO- %s | bash
  else
    echo "AgentFlow: curl or wget is required to install nvm." >&2
    exit 11
  fi
fi
. "$NVM_DIR/nvm.sh"
%s
npm install -g %s
command -v %s >/dev/null 2>&1
node --version
npm --version
`, shellLiteral(nvmInstallScriptURL), shellLiteral(nvmInstallScriptURL), nodeInstall, shellLiteral(target.NPMPackage), shellLiteral(target.Command))

	switch runtimeStatus.Platform {
	case platformWindows:
		return runCombined("wsl.exe", "bash", "-lc", script)
	default:
		return runCombined("bash", "-lc", script)
	}
}

func (i *Installer) runKiroBootstrapScript(runtimeStatus RuntimeStatus) ([]byte, error) {
	script := fmt.Sprintf(`set -e
if command -v curl >/dev/null 2>&1; then
  curl -sSL %s | bash
elif command -v wget >/dev/null 2>&1; then
  wget -qO- %s | bash
else
  echo "AgentFlow: curl or wget is required to install Kiro CLI." >&2
  exit 11
fi
command -v kiro-cli >/dev/null 2>&1
kiro-cli --version || true
`, shellLiteral(kiroInstallScriptURL), shellLiteral(kiroInstallScriptURL))

	switch runtimeStatus.Platform {
	case platformWindows:
		return runCombined("wsl.exe", "bash", "-lc", script)
	default:
		return runCombined("bash", "-lc", script)
	}
}

func (i *Installer) hasNativeCommand(command string) bool {
	return strings.TrimSpace(i.nativeCommandValue(fmt.Sprintf("command -v %s 2>/dev/null || true", shellLiteral(command)))) != ""
}

func (i *Installer) hasWSLCommand(command string) bool {
	return strings.TrimSpace(i.wslCommandValue(fmt.Sprintf("command -v %s 2>/dev/null || true", shellLiteral(command)))) != ""
}

func (i *Installer) RuntimeSummaryLines() []string {
	status := i.CachedRuntimeStatus()
	lines := []string{
		fmt.Sprintf(i.Catalog.Msg("运行环境: %s", "Runtime: %s"), status.runtimeLabel(i.Catalog)),
	}
	if status.Platform == platformWindows && !status.InWSL {
		switch {
		case status.WSLReady:
			lines = append(lines, fmt.Sprintf(i.Catalog.Msg("WSL2: 已检测到 %s", "WSL2: detected %s"), status.WSLDistro))
		default:
			lines = append(lines, i.Catalog.Msg("WSL2: 未检测到（Windows 上建议开启 WSL2）", "WSL2: not detected (recommended on Windows)"))
		}
	}
	if status.NodeVersion != "" {
		lines = append(lines, fmt.Sprintf(i.Catalog.Msg("Node: %s", "Node: %s"), status.NodeVersion))
	} else {
		lines = append(lines, i.Catalog.Msg("Node: 未检测到", "Node: not detected"))
	}
	if status.NPMVersion != "" {
		lines = append(lines, fmt.Sprintf(i.Catalog.Msg("npm: %s", "npm: %s"), status.NPMVersion))
	} else {
		lines = append(lines, i.Catalog.Msg("npm: 未检测到", "npm: not detected"))
	}
	if status.NVMReady {
		lines = append(lines, i.Catalog.Msg("nvm: 可用", "nvm: available"))
	} else {
		lines = append(lines, i.Catalog.Msg("nvm: 未检测到", "nvm: not detected"))
	}
	return lines
}

func (i *Installer) AgentFlowInstallableTargets() []string {
	statuses := i.DetectTargetStatuses()
	result := make([]string, 0, len(statuses))
	for _, status := range statuses {
		if status.CLIInstalled || status.AgentFlowInstalled || status.ConfigDirExists {
			result = append(result, status.Target.Name)
		}
	}
	return result
}

func (s RuntimeStatus) runtimeLabel(catalog interface{ Msg(string, string) string }) string {
	switch {
	case s.InWSL:
		return catalog.Msg("Linux (WSL)", "Linux (WSL)")
	case s.Platform == platformWindows:
		return catalog.Msg("Windows", "Windows")
	case s.Platform == platformDarwin:
		return catalog.Msg("macOS", "macOS")
	case s.Platform == platformLinux:
		return catalog.Msg("Linux", "Linux")
	default:
		return string(s.Platform)
	}
}

func shellLiteral(value string) string {
	if value == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}

func failureLines(target targets.Target, output []byte) []string {
	lines := []string{
		fmt.Sprintf("Auto install for %s failed.", target.DisplayName),
		fmt.Sprintf("Package: %s", target.NPMPackage),
	}
	trimmed := strings.TrimSpace(string(bytes.TrimSpace(output)))
	if trimmed == "" {
		return lines
	}
	for _, line := range tailLines(trimmed, 8) {
		lines = append(lines, line)
	}
	return lines
}

func tailLines(text string, limit int) []string {
	if limit <= 0 {
		return nil
	}
	lines := strings.Split(text, "\n")
	if len(lines) <= limit {
		return lines
	}
	return lines[len(lines)-limit:]
}

func sliceToStatusSet(values []string) map[string]bool {
	result := make(map[string]bool, len(values))
	for _, value := range values {
		result[value] = true
	}
	return result
}

func nodeMajor(version string) int {
	version = strings.TrimSpace(strings.TrimPrefix(version, "v"))
	part := version
	if idx := strings.IndexByte(part, '.'); idx >= 0 {
		part = part[:idx]
	}
	major, _ := strconv.Atoi(part)
	return major
}

// EnvVarConfig describes a single environment variable to configure for a CLI.
type EnvVarConfig struct {
	Label  string // Human-readable label (e.g. "API Key")
	EnvVar string // Environment variable name (e.g. "OPENAI_API_KEY")
}

// CLIConfigFields returns the environment variables that should be configured
// after installing the given CLI target. Returns nil if no configuration is needed.
func (i *Installer) CLIConfigFields(targetName string) []EnvVarConfig {
	target, ok := targets.Lookup(targetName)
	if !ok {
		return nil
	}
	if target.APIKeyEnv == "" {
		return nil
	}
	fields := []EnvVarConfig{
		{Label: "API Key", EnvVar: target.APIKeyEnv},
	}
	if target.BaseURLEnv != "" {
		fields = append(fields, EnvVarConfig{Label: "Base URL", EnvVar: target.BaseURLEnv})
	}
	return fields
}

// WriteEnvConfig writes environment variable exports to the user's shell config
// file (~/.zshrc, ~/.bashrc, or ~/.profile). It returns descriptive lines about
// what was written and any errors encountered.
func (i *Installer) WriteEnvConfig(envVars map[string]string) ([]string, error) {
	if len(envVars) == 0 {
		return []string{i.Catalog.Msg("没有需要写入的配置。", "No configuration to write.")}, nil
	}

	rcFile := i.detectShellRC()
	if rcFile == "" {
		return nil, errors.New(i.Catalog.Msg("未找到 shell 配置文件（~/.zshrc 或 ~/.bashrc）。", "Could not find a shell config file (~/.zshrc or ~/.bashrc)."))
	}

	// Read existing content to check for duplicates.
	existing, _ := os.ReadFile(rcFile)
	content := string(existing)

	var newLines []string
	for envVar, value := range envVars {
		if strings.TrimSpace(value) == "" {
			continue
		}
		exportLine := fmt.Sprintf("export %s=%s", envVar, shellQuote(value))

		// Comment out any existing export for this variable.
		marker := fmt.Sprintf("export %s=", envVar)
		if strings.Contains(content, marker) {
			updatedContent := strings.Builder{}
			for _, line := range strings.Split(content, "\n") {
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(trimmed, marker) && !strings.HasPrefix(trimmed, "#") {
					updatedContent.WriteString("# " + line + "  # replaced by AgentFlow\n")
				} else {
					updatedContent.WriteString(line + "\n")
				}
			}
			content = updatedContent.String()
		}
		newLines = append(newLines, exportLine)
	}

	if len(newLines) == 0 {
		return []string{i.Catalog.Msg("所有配置项均为空，未写入。", "All config values were empty; nothing written.")}, nil
	}

	// Append the new exports.
	block := "\n# AgentFlow CLI configuration\n" + strings.Join(newLines, "\n") + "\n"
	content = strings.TrimRight(content, "\n") + "\n" + block

	if err := os.WriteFile(rcFile, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf(i.Catalog.Msg("写入 %s 失败: %v", "failed to write %s: %v"), rcFile, err)
	}

	lines := []string{
		fmt.Sprintf(i.Catalog.Msg("已写入 %s:", "Written to %s:"), rcFile),
	}
	for _, line := range newLines {
		lines = append(lines, "  "+line)
	}
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf(i.Catalog.Msg("运行 source %s 或重启终端以生效。", "Run source %s or restart your terminal to apply."), rcFile))
	return lines, nil
}

// detectShellRC finds the best shell config file to write to.
func (i *Installer) detectShellRC() string {
	candidates := []string{".zshrc", ".bashrc", ".profile"}
	for _, name := range candidates {
		path := filepath.Join(i.HomeDir, name)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	// Default to .zshrc on macOS, .bashrc elsewhere.
	if currentPlatform() == platformDarwin {
		return filepath.Join(i.HomeDir, ".zshrc")
	}
	return filepath.Join(i.HomeDir, ".bashrc")
}

// shellQuote wraps a value in double quotes, escaping existing double quotes.
func shellQuote(value string) string {
	escaped := strings.ReplaceAll(value, `"`, `\"`)
	return `"` + escaped + `"`
}
