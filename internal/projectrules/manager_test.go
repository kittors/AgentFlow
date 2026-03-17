package projectrules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestManagerInstallWritesExpectedFiles(t *testing.T) {
	root := t.TempDir()

	manager := NewManager()
	written, err := manager.Install(root, []string{"codex", "claude"}, InstallOptions{Profile: "lite"})
	if err != nil {
		t.Fatalf("install: %v", err)
	}
	if len(written) == 0 {
		t.Fatalf("expected files written")
	}

	for _, path := range []string{
		filepath.Join(root, "AGENTS.md"),
		filepath.Join(root, "CLAUDE.md"),
	} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if !strings.Contains(string(data), "AGENTFLOW_ROUTER:") {
			t.Fatalf("expected AgentFlow marker in %s", path)
		}
		content := string(data)
		if strings.Contains(content, ".agentflow/AGENTS.md") {
			t.Fatalf("expected full embedded rules instead of .agentflow/AGENTS.md reference in %s", path)
		}
		if !strings.Contains(content, "先路由再行动") {
			t.Fatalf("expected core AgentFlow rules in %s", path)
		}
	}

	// Project-local .agentflow/ should contain full rules and modules.
	localRulesFile := filepath.Join(root, ".agentflow", "AGENTS.md")
	if _, err := os.Stat(localRulesFile); err != nil {
		t.Fatalf("expected project-local .agentflow/AGENTS.md: %v", err)
	}
	localModuleDir := filepath.Join(root, ".agentflow", "agentflow")
	if _, err := os.Stat(localModuleDir); err != nil {
		t.Fatalf("expected project-local .agentflow/agentflow/ module dir: %v", err)
	}
}

func TestManagerInstallInjectsIntoExistingUserFile(t *testing.T) {
	root := t.TempDir()
	targetPath := filepath.Join(root, "AGENTS.md")
	if err := os.WriteFile(targetPath, []byte("user-rules"), 0o644); err != nil {
		t.Fatalf("write seed: %v", err)
	}

	manager := NewManager()
	if _, err := manager.Install(root, []string{"codex"}, InstallOptions{Profile: "lite"}); err != nil {
		t.Fatalf("install: %v", err)
	}

	data, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "user-rules") {
		t.Fatalf("expected user-rules to be preserved, got: %s", content)
	}
	if !strings.Contains(content, "AGENTFLOW_ROUTER:") {
		t.Fatalf("expected AGENTFLOW_ROUTER marker to be injected, got: %s", content)
	}
}

func TestManagerUninstallStripsMarker(t *testing.T) {
	root := t.TempDir()
	targetPath := filepath.Join(root, "AGENTS.md")
	mixedContent := "<!-- AGENTFLOW_ROUTER: v1.0.0 -->\n\n> **[AgentFlow 管理规则]**\n> 请务必严格按照全局规范（如全局规则或 `.agents/skills/agentflow/SKILL.md`）执行所有操作。\n\n<!-- /AGENTFLOW_ROUTER: -->\nuser-rules-custom"
	if err := os.WriteFile(targetPath, []byte(mixedContent), 0o644); err != nil {
		t.Fatalf("write seed: %v", err)
	}
	manager := NewManager()
	removed, err := manager.Uninstall(root, []string{"codex"})
	if err != nil {
		t.Fatalf("uninstall: %v", err)
	}
	if len(removed) == 0 {
		t.Fatalf("expected files processed, got 0")
	}

	// Should not be deleted
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		t.Fatalf("expected AGENTS.md to be preserved because it has custom rules")
	}

	data, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if strings.Contains(string(data), "AGENTFLOW_ROUTER:") {
		t.Fatalf("expected marker to be stripped, but it was found")
	}
	if string(data) != "user-rules-custom" {
		t.Fatalf("expected user-rules-custom, got %q", string(data))
	}
}

func TestManagerUninstallDeletesEmpty(t *testing.T) {
	root := t.TempDir()
	targetPath := filepath.Join(root, "AGENTS.md")
	markerContent := "<!-- AGENTFLOW_ROUTER: v1.0.0 -->\n\nsome text inside \n<!-- /AGENTFLOW_ROUTER: -->\n"
	if err := os.WriteFile(targetPath, []byte(markerContent), 0o644); err != nil {
		t.Fatalf("write seed: %v", err)
	}
	manager := NewManager()
	_, err := manager.Uninstall(root, []string{"codex"})
	if err != nil {
		t.Fatalf("uninstall: %v", err)
	}

	// Should be deleted
	if _, err := os.Stat(targetPath); err == nil {
		t.Fatalf("expected AGENTS.md to be deleted because it became empty")
	}
}

func TestManagerDetectDoesNotMisattributeSharedSkill(t *testing.T) {
	root := t.TempDir()

	manager := NewManager()
	if _, err := manager.Install(root, []string{"codex"}, InstallOptions{Profile: "lite"}); err != nil {
		t.Fatalf("install: %v", err)
	}

	statuses, err := manager.Detect(root)
	if err != nil {
		t.Fatalf("detect: %v", err)
	}

	states := make(map[string]Status, len(statuses))
	for _, status := range statuses {
		states[status.Target] = status
	}

	codexStatus, ok := states["codex"]
	if !ok || !codexStatus.Managed {
		t.Fatalf("expected codex to be detected as managed, got %#v", codexStatus)
	}

	claudeStatus, ok := states["claude"]
	if !ok {
		t.Fatalf("expected claude status entry, got %#v", statuses)
	}
	if claudeStatus.Exists || claudeStatus.Managed {
		t.Fatalf("expected claude to remain absent when only codex is installed, got %#v", claudeStatus)
	}
}

func TestManagerUninstallKeepsSharedSkillWhileOtherTargetRemains(t *testing.T) {
	root := t.TempDir()

	manager := NewManager()
	if _, err := manager.Install(root, []string{"codex", "claude"}, InstallOptions{Profile: "lite"}); err != nil {
		t.Fatalf("install: %v", err)
	}

	skillPath := filepath.Join(root, ".agents", "skills", "agentflow", "SKILL.md")
	if _, err := os.Stat(skillPath); err != nil {
		t.Fatalf("expected shared skill to exist: %v", err)
	}

	removed, err := manager.Uninstall(root, []string{"codex"})
	if err != nil {
		t.Fatalf("uninstall codex: %v", err)
	}
	for _, path := range removed {
		if path == skillPath {
			t.Fatalf("expected shared skill to be preserved while claude remains, removed=%#v", removed)
		}
	}

	if _, err := os.Stat(skillPath); err != nil {
		t.Fatalf("expected shared skill to remain after uninstalling codex: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "CLAUDE.md")); err != nil {
		t.Fatalf("expected claude rules file to remain: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".agentflow", "agentflow")); err != nil {
		t.Fatalf("expected project-local module dir to remain while claude is still installed: %v", err)
	}
}
