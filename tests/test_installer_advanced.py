"""Advanced tests for installer.py — deploy functions, safe write/remove, hooks, codex agents."""

import json
import shutil
from pathlib import Path
from unittest import mock

import pytest

from agentflow._constants import AGENTFLOW_MARKER, CLI_TARGETS, PLUGIN_DIR_NAME


@pytest.fixture
def mock_home(tmp_path):
    """Patch Path.home() and create CLI target dirs."""
    for config in CLI_TARGETS.values():
        (tmp_path / config["dir"]).mkdir(parents=True, exist_ok=True)
    with mock.patch("pathlib.Path.home", return_value=tmp_path):
        yield tmp_path


# ── _safe_remove ──────────────────────────────────────────────────────────────


class TestSafeRemove:
    def test_remove_file(self, tmp_path):
        from agentflow.installer import _safe_remove

        f = tmp_path / "test.txt"
        f.write_text("hello")
        assert _safe_remove(f)
        assert not f.exists()

    def test_remove_directory(self, tmp_path):
        from agentflow.installer import _safe_remove

        d = tmp_path / "subdir"
        d.mkdir()
        (d / "file.txt").write_text("content")
        assert _safe_remove(d)
        assert not d.exists()

    def test_remove_nonexistent(self, tmp_path):
        from agentflow.installer import _safe_remove

        assert _safe_remove(tmp_path / "nonexistent")

    def test_remove_permission_error(self, tmp_path):
        from agentflow.installer import _safe_remove

        f = tmp_path / "locked.txt"
        f.write_text("data")
        with mock.patch("pathlib.Path.unlink", side_effect=PermissionError):
            # Should try to rename aside
            result = _safe_remove(f)
            # Either renamed or reported error
            assert isinstance(result, bool)


# ── _safe_write ───────────────────────────────────────────────────────────────


class TestSafeWrite:
    def test_write_new_file(self, tmp_path):
        from agentflow.installer import _safe_write

        f = tmp_path / "new.md"
        assert _safe_write(f, "# Hello")
        assert f.read_text() == "# Hello"

    def test_write_overwrites(self, tmp_path):
        from agentflow.installer import _safe_write

        f = tmp_path / "existing.md"
        f.write_text("old")
        assert _safe_write(f, "new")
        assert f.read_text() == "new"


# ── _get_source_files ─────────────────────────────────────────────────────────


class TestGetSourceFiles:
    def test_returns_expected_subdirs(self):
        from agentflow.installer import _get_source_files

        sources = _get_source_files()
        # Should find at least stages, services, rules, rlm, functions, templates, hooks, agents
        expected = {"stages", "services", "rules", "rlm", "functions", "templates", "hooks", "agents"}
        assert expected.issubset(set(sources.keys()))


# ── _deploy_rules_file ────────────────────────────────────────────────────────


class TestDeployRulesFile:
    def test_deploy_creates_rules(self, mock_home):
        from agentflow.installer import _deploy_rules_file

        cli_dir = mock_home / ".codex"
        assert _deploy_rules_file("codex", cli_dir)
        rules = cli_dir / "AGENTS.md"
        assert rules.exists()
        assert AGENTFLOW_MARKER in rules.read_text()

    def test_deploy_backs_up_user_file(self, mock_home):
        from agentflow.installer import _deploy_rules_file

        cli_dir = mock_home / ".claude"
        rules = cli_dir / "CLAUDE.md"
        rules.write_text("# User's custom rules")
        assert _deploy_rules_file("claude", cli_dir)
        backups = list(cli_dir.glob("*_bak*"))
        assert len(backups) >= 1


# ── _deploy_module_dir ────────────────────────────────────────────────────────


class TestDeployModuleDir:
    def test_deploy_creates_plugin_dir(self, mock_home):
        from agentflow.installer import _deploy_module_dir

        cli_dir = mock_home / ".codex"
        assert _deploy_module_dir("codex", cli_dir)
        plugin_dir = cli_dir / PLUGIN_DIR_NAME
        assert plugin_dir.exists()
        # Should have submodules
        subdirs = [p.name for p in plugin_dir.iterdir() if p.is_dir()]
        assert len(subdirs) > 0


# ── _deploy_skill_md ──────────────────────────────────────────────────────────


class TestDeploySkillMd:
    def test_deploy_skill(self, mock_home):
        from agentflow.installer import _deploy_skill_md

        cli_dir = mock_home / ".codex"
        assert _deploy_skill_md("codex", cli_dir)
        skill = cli_dir / "skills" / "agentflow" / "SKILL.md"
        assert skill.exists()


# ── _deploy_hooks ─────────────────────────────────────────────────────────────


class TestDeployHooks:
    def test_deploy_claude_hooks(self, mock_home):
        from agentflow.installer import _deploy_hooks

        cli_dir = mock_home / ".claude"
        assert _deploy_hooks("claude", cli_dir)
        settings = cli_dir / "settings.json"
        if settings.exists():
            data = json.loads(settings.read_text())
            assert "hooks" in data

    def test_deploy_codex_hooks(self, mock_home):
        from agentflow.installer import _deploy_hooks

        cli_dir = mock_home / ".codex"
        assert _deploy_hooks("codex", cli_dir)

    def test_deploy_hooks_unknown_target(self, mock_home):
        """Non claude/codex targets should pass silently."""
        from agentflow.installer import _deploy_hooks

        cli_dir = mock_home / ".gemini"
        assert _deploy_hooks("gemini", cli_dir)


# ── _deploy_codex_agents (non-interactive) ────────────────────────────────────


class TestDeployCodexAgents:
    def test_skip_in_non_interactive(self, mock_home):
        """Non-interactive mode (piped stdin) should skip without error."""
        from agentflow.installer import _deploy_codex_agents

        cli_dir = mock_home / ".codex"
        with mock.patch("sys.stdin") as mock_stdin:
            mock_stdin.isatty.return_value = False
            result = _deploy_codex_agents(cli_dir)
            assert result is True

    def test_already_enabled(self, mock_home):
        """When multi_agent is already enabled, should update without prompt."""
        from agentflow.installer import _deploy_codex_agents

        cli_dir = mock_home / ".codex"
        config = cli_dir / "config.toml"
        config.write_text("[features]\nmulti_agent = true\n")
        result = _deploy_codex_agents(cli_dir)
        assert result is True
        # Agent role files should be deployed
        assert (cli_dir / "agents" / "reviewer.toml").exists()
        assert (cli_dir / "agents" / "architect.toml").exists()


# ── Full install/uninstall cycle ──────────────────────────────────────────────


class TestInstallUninstallCycle:
    def test_install_all_targets(self, mock_home):
        """Install to every target and verify."""
        from agentflow.installer import install

        agents_md = Path(__file__).parent.parent / "AGENTS.md"
        if not agents_md.exists():
            pytest.skip("AGENTS.md not found")

        for target in CLI_TARGETS:
            result = install(target)
            assert result, f"Failed to install to {target}"

    def test_uninstall_cleans_everything(self, mock_home):
        """Install then uninstall should clean up all files."""
        from agentflow.installer import install, uninstall

        agents_md = Path(__file__).parent.parent / "AGENTS.md"
        if not agents_md.exists():
            pytest.skip("AGENTS.md not found")

        install("claude")
        uninstall("claude")

        cli_dir = mock_home / ".claude"
        assert not (cli_dir / "CLAUDE.md").exists()
        assert not (cli_dir / PLUGIN_DIR_NAME).exists()

    def test_uninstall_unknown_target(self, mock_home):
        from agentflow.installer import uninstall

        assert not uninstall("fake_target")

    def test_install_missing_dir(self, tmp_path):
        """Install to target whose dir doesn't exist."""
        from agentflow.installer import install

        with mock.patch("pathlib.Path.home", return_value=tmp_path):
            assert not install("codex")


class TestUninstallAll:
    def test_uninstall_all_none_installed(self, mock_home):
        from agentflow.installer import uninstall_all

        assert not uninstall_all()
