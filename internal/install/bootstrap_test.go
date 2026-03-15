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

func TestManualInstallLinesWindowsIncludeWSLGuidance(t *testing.T) {
	restoreBootstrapTestEnv(t)
	runtimeGOOS = "windows"
	t.Setenv("WSL_DISTRO_NAME", "")
	t.Setenv("WSL_INTEROP", "")

	installer := New(i18n.NewCatalog(), &bytes.Buffer{})
	lines, err := installer.ManualInstallLines("gemini")
	if err != nil {
		t.Fatalf("ManualInstallLines returned error: %v", err)
	}

	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "WSL2") {
		t.Fatalf("expected WSL2 guidance, got %q", joined)
	}
	if !strings.Contains(joined, "npm install -g @google/gemini-cli") {
		t.Fatalf("expected npm install command, got %q", joined)
	}
}

func TestInstallCreatesMissingTargetDir(t *testing.T) {
	restoreBootstrapTestEnv(t)
	runtimeGOOS = "darwin"

	installer := New(i18n.NewCatalog(), &bytes.Buffer{})
	installer.HomeDir = t.TempDir()

	if err := installer.Install("gemini", "lite"); err != nil {
		t.Fatalf("Install returned error: %v", err)
	}

	targetDir := filepath.Join(installer.HomeDir, ".gemini")
	if info, err := os.Stat(targetDir); err != nil || !info.IsDir() {
		t.Fatalf("expected target dir to be created, err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(targetDir, "GEMINI.md")); err != nil {
		t.Fatalf("expected rules file to exist, err=%v", err)
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
