package install

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"
	"testing"

	"github.com/kittors/AgentFlow/internal/i18n"
)

func TestUninstallCLINpmUsesBashOnUnix(t *testing.T) {
	restoreCLIUninstallTestEnv(t)
	runtimeGOOS = "darwin"

	lookPath = func(name string) (string, error) {
		if name == "bash" {
			return "/bin/bash", nil
		}
		return "", exec.ErrNotFound
	}

	called := 0
	runCombined = func(name string, args ...string) ([]byte, error) {
		if name != "bash" {
			return nil, errors.New("unexpected command")
		}
		script := args[len(args)-1]
		// The new uninstall script contains the package name for both fnm and nvm attempts.
		if strings.Contains(script, "uninstall -g '@openai/codex'") {
			called++
			return []byte("ok\n"), nil
		}
		// Allow runtime probing + CLI path probing.
		if strings.Contains(script, "command -v 'codex'") {
			return []byte("/usr/local/bin/codex\n"), nil
		}
		return []byte(""), nil
	}

	installer := New(i18n.NewCatalog(), &bytes.Buffer{})
	installer.HomeDir = t.TempDir()

	lines, err := installer.UninstallCLI("codex")
	if err != nil {
		t.Fatalf("UninstallCLI returned error: %v", err)
	}
	if called != 1 {
		t.Fatalf("expected npm uninstall to be executed once, got %d", called)
	}
	if joined := strings.Join(lines, "\n"); !strings.Contains(joined, "npm uninstall -g @openai/codex") {
		t.Fatalf("expected npm uninstall command in output, got %q", joined)
	}
}

func TestUninstallCLINpmUsesWSLWhenDetected(t *testing.T) {
	restoreCLIUninstallTestEnv(t)
	runtimeGOOS = "windows"
	t.Setenv("WSL_DISTRO_NAME", "")
	t.Setenv("WSL_INTEROP", "")

	lookPath = func(name string) (string, error) {
		switch name {
		case "wsl.exe":
			return `C:\Windows\System32\wsl.exe`, nil
		default:
			return "", exec.ErrNotFound
		}
	}

	var uninstallSeen bool
	runCombined = func(name string, args ...string) ([]byte, error) {
		if name != "wsl.exe" {
			return nil, errors.New("unexpected command")
		}
		if len(args) >= 2 && args[0] == "-l" && args[1] == "-q" {
			return []byte("Ubuntu\n"), nil
		}
		if len(args) >= 3 && args[0] == "bash" && args[1] == "-lc" {
			script := args[len(args)-1]
			switch {
			case strings.Contains(script, "command -v 'codex'"):
				return []byte("/usr/bin/codex\n"), nil
			case strings.Contains(script, "uninstall -g '@openai/codex'"):
				uninstallSeen = true
				return []byte("ok\n"), nil
			default:
				return []byte(""), nil
			}
		}
		return []byte(""), nil
	}

	installer := New(i18n.NewCatalog(), &bytes.Buffer{})
	installer.HomeDir = t.TempDir()

	if _, err := installer.UninstallCLI("codex"); err != nil {
		t.Fatalf("UninstallCLI returned error: %v", err)
	}
	if !uninstallSeen {
		t.Fatal("expected WSL npm uninstall to be executed")
	}
}

func restoreCLIUninstallTestEnv(t *testing.T) {
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
