package install

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kittors/AgentFlow/internal/config"
	"github.com/kittors/AgentFlow/internal/i18n"
)

func TestDetectTargetStatusWindowsWithoutWSLDisablesAutoInstall(t *testing.T) {
	restoreBootstrapTestEnv(t)
	runtimeGOOS = "windows"
	t.Setenv("WSL_DISTRO_NAME", "")
	t.Setenv("WSL_INTEROP", "")

	lookPath = func(name string) (string, error) {
		switch name {
		case "bash":
			return `C:\Program Files\Git\bin\bash.exe`, nil
		default:
			return "", exec.ErrNotFound
		}
	}

	installer := New(i18n.NewCatalog(), &bytes.Buffer{})
	installer.HomeDir = t.TempDir()

	status, err := installer.DetectTargetStatus("codex")
	if err != nil {
		t.Fatalf("DetectTargetStatus returned error: %v", err)
	}
	if status.AutoInstallSupported {
		t.Fatal("expected auto install to be disabled when Windows has no WSL2")
	}
	if status.Runtime.WSLReady {
		t.Fatal("expected WSL2 to be reported as unavailable")
	}
	if joined := strings.Join(status.Notes, "\n"); !strings.Contains(joined, "WSL2") {
		t.Fatalf("expected WSL2 guidance in notes, got %q", joined)
	}
}

func TestStatusLinesReflectCLIAndAgentFlowStateSeparately(t *testing.T) {
	restoreBootstrapTestEnv(t)
	runtimeGOOS = "darwin"

	lookPath = func(name string) (string, error) {
		if name == "bash" {
			return "/bin/bash", nil
		}
		return "", exec.ErrNotFound
	}
	runCombined = func(name string, args ...string) ([]byte, error) {
		if name != "bash" {
			return nil, errors.New("unexpected command")
		}
		script := args[len(args)-1]
		switch {
		case strings.Contains(script, "command -v 'codex'"):
			return []byte("/usr/local/bin/codex\n"), nil
		case strings.Contains(script, "command -v 'claude'"):
			return []byte(""), nil
		default:
			return []byte(""), nil
		}
	}

	installer := New(i18n.NewCatalog(), &bytes.Buffer{})
	installer.HomeDir = t.TempDir()

	claudeDir := filepath.Join(installer.HomeDir, ".claude")
	if err := os.MkdirAll(filepath.Join(claudeDir, config.PluginDirName), 0o755); err != nil {
		t.Fatalf("mkdir plugin dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(claudeDir, "CLAUDE.md"), []byte("<!-- AGENTFLOW_ROUTER: v1.0.0 -->\n"), 0o644); err != nil {
		t.Fatalf("write CLAUDE.md: %v", err)
	}

	lines := strings.Join(installer.StatusLines(), "\n")
	if !strings.Contains(lines, "CLI") || !strings.Contains(lines, "codex") {
		t.Fatalf("expected codex CLI-only state, got %q", lines)
	}
	if !strings.Contains(lines, "AF") || !strings.Contains(lines, "claude") {
		t.Fatalf("expected claude AgentFlow-only state, got %q", lines)
	}
}

func restoreBootstrapTestEnv(t *testing.T) {
	t.Helper()
	previousLookPath := lookPath
	previousRunCombined := runCombined
	previousRuntimeGOOS := runtimeGOOS
	t.Cleanup(func() {
		lookPath = previousLookPath
		runCombined = previousRunCombined
		runtimeGOOS = previousRuntimeGOOS
	})
}

func TestDetectNVMWindowsWithNVMHome(t *testing.T) {
	t.Setenv("NVM_HOME", `C:\Users\Test\AppData\Roaming\nvm`)
	if !detectNVMWindows() {
		t.Fatal("expected detectNVMWindows to return true when NVM_HOME is set")
	}
}

func TestDetectNVMWindowsWithLookPath(t *testing.T) {
	restoreBootstrapTestEnv(t)
	t.Setenv("NVM_HOME", "")
	lookPath = func(name string) (string, error) {
		if name == "nvm" {
			return `C:\Users\Test\AppData\Roaming\nvm\nvm.exe`, nil
		}
		return "", exec.ErrNotFound
	}
	if !detectNVMWindows() {
		t.Fatal("expected detectNVMWindows to return true when nvm is in PATH")
	}
}

func TestDetectNVMWindowsNotInstalled(t *testing.T) {
	restoreBootstrapTestEnv(t)
	t.Setenv("NVM_HOME", "")
	lookPath = func(name string) (string, error) {
		return "", exec.ErrNotFound
	}
	if detectNVMWindows() {
		t.Fatal("expected detectNVMWindows to return false when nvm-windows is not installed")
	}
}

func TestRuntimeStatusWindowsNativeNVMDetected(t *testing.T) {
	restoreBootstrapTestEnv(t)
	runtimeGOOS = "windows"
	t.Setenv("WSL_DISTRO_NAME", "")
	t.Setenv("WSL_INTEROP", "")
	t.Setenv("NVM_HOME", `C:\Users\Test\AppData\Roaming\nvm`)

	lookPath = func(name string) (string, error) {
		switch name {
		case "node":
			return `C:\Program Files\nodejs\node.exe`, nil
		case "npm":
			return `C:\Program Files\nodejs\npm.cmd`, nil
		case "wsl.exe":
			return "", exec.ErrNotFound // No WSL
		default:
			return "", exec.ErrNotFound
		}
	}
	runCombined = func(name string, args ...string) ([]byte, error) {
		switch name {
		case "node":
			return []byte("v24.14.0\n"), nil
		case "npm":
			return []byte("11.9.0\n"), nil
		default:
			return nil, errors.New("unexpected command")
		}
	}

	installer := New(i18n.NewCatalog(), &bytes.Buffer{})
	installer.HomeDir = t.TempDir()
	status := installer.RuntimeStatus()

	if !status.NVMReady {
		t.Fatal("expected NVMReady to be true when NVM_HOME is set on Windows")
	}
	if status.NodeVersion != "v24.14.0" {
		t.Fatalf("expected NodeVersion v24.14.0, got %q", status.NodeVersion)
	}
}

func TestDetectNPMMirrorWithChineseLocale(t *testing.T) {
	restoreBootstrapTestEnv(t)
	t.Setenv("LANG", "zh_CN.UTF-8")
	t.Setenv("LC_ALL", "")
	t.Setenv("LANGUAGE", "")
	t.Setenv("LC_MESSAGES", "")

	lookPath = func(name string) (string, error) {
		return "", exec.ErrNotFound // No npm available
	}

	mirror := detectNPMMirror()
	if mirror != "https://registry.npmmirror.com" {
		t.Fatalf("expected npmmirror URL, got %q", mirror)
	}
}

func TestDetectNPMMirrorWithExistingConfig(t *testing.T) {
	restoreBootstrapTestEnv(t)
	t.Setenv("LANG", "en_US.UTF-8")
	t.Setenv("LC_ALL", "")
	t.Setenv("LANGUAGE", "")
	t.Setenv("LC_MESSAGES", "")
	runtimeGOOS = "linux" // Avoid Windows PowerShell check

	lookPath = func(name string) (string, error) {
		if name == "npm" {
			return "/usr/bin/npm", nil
		}
		return "", exec.ErrNotFound
	}
	runCombined = func(name string, args ...string) ([]byte, error) {
		if name == "/usr/bin/npm" && len(args) == 3 && args[0] == "config" && args[1] == "get" && args[2] == "registry" {
			return []byte("https://registry.npmmirror.com\n"), nil
		}
		return nil, errors.New("unexpected")
	}

	mirror := detectNPMMirror()
	if mirror != "https://registry.npmmirror.com" {
		t.Fatalf("expected user-configured mirror, got %q", mirror)
	}
}

func TestDetectNPMMirrorNoMirrorNeeded(t *testing.T) {
	restoreBootstrapTestEnv(t)
	t.Setenv("LANG", "en_US.UTF-8")
	t.Setenv("LC_ALL", "")
	t.Setenv("LANGUAGE", "")
	t.Setenv("LC_MESSAGES", "")
	runtimeGOOS = "linux"

	lookPath = func(name string) (string, error) {
		if name == "npm" {
			return "/usr/bin/npm", nil
		}
		return "", exec.ErrNotFound
	}
	runCombined = func(name string, args ...string) ([]byte, error) {
		if name == "/usr/bin/npm" && len(args) == 3 && args[0] == "config" {
			return []byte("https://registry.npmjs.org/\n"), nil
		}
		return nil, errors.New("unexpected")
	}

	mirror := detectNPMMirror()
	if mirror != "" {
		t.Fatalf("expected no mirror for en_US locale, got %q", mirror)
	}
}
