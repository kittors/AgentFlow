package install

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
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
		if strings.Contains(script, "npm uninstall -g '@openai/codex'") {
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
			case strings.Contains(script, "npm uninstall -g '@openai/codex'"):
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

func TestInstallAndUninstallKiroCreatesAndRemovesAgentFiles(t *testing.T) {
	installer := New(i18n.NewCatalog(), &bytes.Buffer{})
	installer.HomeDir = t.TempDir()

	if err := installer.Install("kiro", "lite"); err != nil {
		t.Fatalf("Install returned error: %v", err)
	}

	kiroDir := filepath.Join(installer.HomeDir, ".kiro")
	agentPath := filepath.Join(kiroDir, "agents", "agentflow.json")
	promptPath := filepath.Join(kiroDir, "prompts", "agentflow.md")

	agentBytes, err := os.ReadFile(agentPath)
	if err != nil {
		t.Fatalf("expected agent config to exist: %v", err)
	}
	if !strings.Contains(string(agentBytes), "AGENTFLOW_ROUTER:") {
		t.Fatalf("expected marker in agent config, got %q", string(agentBytes))
	}

	promptBytes, err := os.ReadFile(promptPath)
	if err != nil {
		t.Fatalf("expected prompt file to exist: %v", err)
	}
	if !strings.Contains(string(promptBytes), "AGENTFLOW_ROUTER:") {
		t.Fatal("expected marker in prompt file")
	}

	if err := installer.Uninstall("kiro"); err != nil {
		t.Fatalf("Uninstall returned error: %v", err)
	}
	if _, err := os.Stat(agentPath); !os.IsNotExist(err) {
		t.Fatalf("expected agent config removed, got err=%v", err)
	}
	if _, err := os.Stat(promptPath); !os.IsNotExist(err) {
		t.Fatalf("expected prompt removed, got err=%v", err)
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
