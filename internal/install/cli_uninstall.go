package install

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/kittors/AgentFlow/internal/config"
	"github.com/kittors/AgentFlow/internal/targets"
)

func (i *Installer) UninstallCLI(targetName string) ([]string, error) {
	target, ok := targets.Lookup(targetName)
	if !ok {
		return nil, fmt.Errorf(i.Catalog.Msg("未知目标: %s", "unknown target: %s"), targetName)
	}

	runtimeStatus := i.RuntimeStatus()
	status := i.detectTargetStatus(target)
	scope := status.CLIPathScope

	// If the CLI wasn't detected, still allow best-effort uninstall for known package/script targets.
	if strings.TrimSpace(scope) == "" && runtimeStatus.Platform == platformWindows && !runtimeStatus.InWSL && runtimeStatus.WSLReady && target.RecommendWSLOnWindows {
		scope = "wsl"
	}

	output, err := i.runCLIUninstall(target, runtimeStatus, scope)
	if err != nil {
		trimmed := strings.TrimSpace(string(bytes.TrimSpace(output)))
		if trimmed != "" {
			return nil, fmt.Errorf("%s", trimmed)
		}
		return nil, err
	}

	lines := []string{
		fmt.Sprintf(i.Catalog.Msg("已卸载 %s。", "Uninstalled %s."), target.DisplayName),
	}
	switch {
	case target.Name == "kiro":
		lines = append(lines, "Command: kiro-cli uninstall")
	case target.NPMPackage != "":
		lines = append(lines, fmt.Sprintf("Command: npm uninstall -g %s", target.NPMPackage))
	}
	if runtimeStatus.Platform == platformWindows && scope == "wsl" && !runtimeStatus.InWSL {
		lines = append(lines, i.Catalog.Msg("该卸载在 WSL2 中执行。", "This uninstall was executed inside WSL2."))
	}
	return lines, nil
}

func (i *Installer) PurgeConfigDir(targetName string) error {
	target, ok := targets.Lookup(targetName)
	if !ok {
		return fmt.Errorf(i.Catalog.Msg("未知目标: %s", "unknown target: %s"), targetName)
	}

	// Always purge the host-side config dir (the same place AgentFlow writes into).
	if err := config.SafeRemove(filepath.Join(i.HomeDir, target.Dir)); err != nil {
		return err
	}

	// On Windows host, CLIs may have been installed inside WSL2. If we detect a WSL-scoped CLI,
	// also purge the WSL-side config dir best-effort.
	runtimeStatus := i.RuntimeStatus()
	if runtimeStatus.Platform == platformWindows && !runtimeStatus.InWSL && runtimeStatus.WSLReady {
		status := i.detectTargetStatus(target)
		if status.CLIPathScope == "wsl" {
			_, _ = i.purgeWSLConfigDir(target)
		}
	}
	return nil
}

func (i *Installer) ManualUninstallLines(targetName string) ([]string, error) {
	target, ok := targets.Lookup(targetName)
	if !ok {
		return nil, fmt.Errorf(i.Catalog.Msg("未知目标: %s", "unknown target: %s"), targetName)
	}
	status := i.detectTargetStatus(target)
	return i.manualUninstallLines(target, i.RuntimeStatus(), status), nil
}

func (i *Installer) purgeWSLConfigDir(target targets.Target) ([]byte, error) {
	// Note: this uses rm -rf inside WSL, only when full-uninstall was explicitly requested.
	script := fmt.Sprintf(`set -e
p="$HOME/%s"
if [ -e "$p" ]; then
  rm -rf -- "$p"
fi
`, target.Dir)
	return runCombined("wsl.exe", "bash", "-lc", script)
}

func (i *Installer) manualUninstallLines(target targets.Target, runtimeStatus RuntimeStatus, status TargetStatus) []string {
	lines := []string{
		fmt.Sprintf(i.Catalog.Msg("%s 的推荐卸载命令:", "Recommended uninstall command for %s:"), target.DisplayName),
	}

	switch {
	case target.Name == "kiro":
		lines = append(lines, "kiro-cli uninstall")
		lines = append(lines, "")
		lines = append(lines, i.Catalog.Msg("Ubuntu/Debian (如果是 apt 安装):", "Ubuntu/Debian (if installed via apt):"))
		lines = append(lines, "sudo apt remove kiro-cli")
		lines = append(lines, "sudo apt purge kiro-cli")
	case target.NPMPackage != "":
		lines = append(lines, fmt.Sprintf("npm uninstall -g %s", target.NPMPackage))
	default:
		lines = append(lines, i.Catalog.Msg("该目标暂无自动卸载路径；请参考官方文档。", "This target does not have an automatic uninstall path yet; follow the official docs."))
	}

	if target.DocsURL != "" {
		lines = append(lines, fmt.Sprintf(i.Catalog.Msg("官方文档: %s", "Official docs: %s"), target.DocsURL))
	}

	if runtimeStatus.Platform == platformWindows && !runtimeStatus.InWSL && target.RecommendWSLOnWindows {
		if status.Runtime.WSLReady {
			lines = append(lines, "")
			lines = append(lines, i.Catalog.Msg("Windows 建议在 WSL2 Bash 中安装与卸载这些 CLI。", "On Windows, install and uninstall these CLIs inside WSL2 Bash."))
		} else {
			lines = append(lines, "")
			lines = append(lines, i.Catalog.Msg("Windows 未检测到 WSL2；建议先启用 WSL2。", "WSL2 was not detected on Windows; enable WSL2 first."))
		}
	}
	return lines
}

func (i *Installer) runCLIUninstall(target targets.Target, runtimeStatus RuntimeStatus, scope string) ([]byte, error) {
	if target.Name == "kiro" {
		return i.runKiroUninstall(runtimeStatus, scope)
	}
	if target.NPMPackage != "" {
		return i.runNpmUninstall(target, runtimeStatus, scope)
	}
	return nil, errors.New(strings.Join(i.manualUninstallLines(target, runtimeStatus, i.detectTargetStatus(target)), "\n"))
}

func (i *Installer) runKiroUninstall(runtimeStatus RuntimeStatus, scope string) ([]byte, error) {
	script := "set -e\nkiro-cli uninstall\n"
	if runtimeStatus.Platform == platformWindows && scope == "wsl" && !runtimeStatus.InWSL {
		return runCombined("wsl.exe", "bash", "-lc", script)
	}
	if runtimeStatus.Platform == platformWindows && !runtimeStatus.InWSL {
		// Kiro CLI is not guaranteed to be installed on Windows host.
		return nil, errors.New(strings.Join(i.manualUninstallLines(targets.All["kiro"], runtimeStatus, TargetStatus{Runtime: runtimeStatus}), "\n"))
	}
	return runCombined("bash", "-lc", script)
}

func (i *Installer) runNpmUninstall(target targets.Target, runtimeStatus RuntimeStatus, scope string) ([]byte, error) {
	if target.NPMPackage == "" {
		return nil, fmt.Errorf("missing npm package for %s", target.Name)
	}

	// Build a script that tries npm uninstall from BOTH fnm and nvm environments.
	// On systems with both node managers, packages may be installed under either
	// or both. A single `npm uninstall -g` only affects the active npm's prefix.
	uninstallScript := nodeManagerUninstallScript(target.NPMPackage)

	if runtimeStatus.Platform == platformWindows && scope == "wsl" && !runtimeStatus.InWSL {
		return runCombined("wsl.exe", "bash", "-lc", uninstallScript)
	}

	if runtimeStatus.Platform == platformWindows && !runtimeStatus.InWSL {
		// Windows host: fall back to running npm directly (if present).
		output, err := runCombined("npm", "uninstall", "-g", target.NPMPackage)
		if err != nil {
			return output, errors.New(strings.Join(i.manualUninstallLines(target, runtimeStatus, TargetStatus{Runtime: runtimeStatus}), "\n"))
		}
		return output, nil
	}

	return runCombined("bash", "-lc", uninstallScript)
}

// nodeManagerSetup returns a shell snippet that ensures the correct Node.js
// environment is active, supporting both fnm and nvm. fnm is checked first
// because on systems where both are installed, fnm is typically the primary.
func nodeManagerSetup() string {
	return `# Try fnm first (Fast Node Manager)
if command -v fnm >/dev/null 2>&1; then
  eval "$(fnm env)" >/dev/null 2>&1
elif [ -s "$HOME/.local/share/fnm/fnm" ]; then
  eval "$("$HOME/.local/share/fnm/fnm" env)" >/dev/null 2>&1
fi
# Fall back to nvm
export NVM_DIR="$HOME/.nvm"
if [ -s "$NVM_DIR/nvm.sh" ]; then . "$NVM_DIR/nvm.sh" >/dev/null 2>&1; fi`
}

// nodeManagerUninstallScript returns a shell script that tries npm uninstall
// from BOTH fnm and nvm environments. This handles the common case where a
// user has both node managers and packages are installed under one or both.
// The script does NOT use `set -e` because individual npm uninstall calls may
// fail (package not found in that environment) which is expected and harmless.
func nodeManagerUninstallScript(npmPackage string) string {
	pkg := shellLiteral(npmPackage)
	return fmt.Sprintf(`_AF_UNINSTALLED=0

# --- Try fnm first ---
_AF_FNM_NPM=""
if command -v fnm >/dev/null 2>&1; then
  eval "$(fnm env)" >/dev/null 2>&1
  _AF_FNM_NPM="$(command -v npm 2>/dev/null || true)"
elif [ -s "$HOME/.local/share/fnm/fnm" ]; then
  eval "$("$HOME/.local/share/fnm/fnm" env)" >/dev/null 2>&1
  _AF_FNM_NPM="$(command -v npm 2>/dev/null || true)"
fi

if [ -n "$_AF_FNM_NPM" ]; then
  "$_AF_FNM_NPM" uninstall -g %s >/dev/null 2>&1 && _AF_UNINSTALLED=1 || true
fi

# --- Try nvm ---
_AF_NVM_NPM=""
export NVM_DIR="$HOME/.nvm"
if [ -s "$NVM_DIR/nvm.sh" ]; then
  . "$NVM_DIR/nvm.sh" >/dev/null 2>&1
  _AF_NVM_NPM="$(command -v npm 2>/dev/null || true)"
fi

if [ -n "$_AF_NVM_NPM" ] && [ "$_AF_NVM_NPM" != "$_AF_FNM_NPM" ]; then
  "$_AF_NVM_NPM" uninstall -g %s >/dev/null 2>&1 && _AF_UNINSTALLED=1 || true
fi

# --- Fallback: plain npm on PATH ---
if [ "$_AF_UNINSTALLED" = "0" ]; then
  _AF_PATH_NPM="$(command -v npm 2>/dev/null || true)"
  if [ -n "$_AF_PATH_NPM" ]; then
    "$_AF_PATH_NPM" uninstall -g %s || exit 1
  else
    echo "npm not found" >&2
    exit 1
  fi
fi
`, pkg, pkg, pkg)
}
