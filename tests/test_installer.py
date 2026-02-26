"""Tests for the installer module."""

from pathlib import Path
from unittest import mock

import pytest

from agentflow._constants import AGENTFLOW_MARKER, CLI_TARGETS, PLUGIN_DIR_NAME


@pytest.fixture
def temp_home(tmp_path):
    """Create a temporary home directory with CLI target dirs."""
    for name, config in CLI_TARGETS.items():
        cli_dir = tmp_path / config["dir"]
        cli_dir.mkdir(parents=True)
    return tmp_path


@pytest.fixture
def mock_home(temp_home):
    """Patch Path.home() to return temp directory."""
    with mock.patch("pathlib.Path.home", return_value=temp_home):
        yield temp_home


class TestFileIdentification:
    def test_is_agentflow_file_true(self, tmp_path):
        """Test identifying AgentFlow-created files."""
        from agentflow._constants import is_agentflow_file

        f = tmp_path / "test.md"
        f.write_text(f"<!-- {AGENTFLOW_MARKER} v1.0.0 -->\n# Content")
        assert is_agentflow_file(f)

    def test_is_agentflow_file_false(self, tmp_path):
        """Test non-AgentFlow files."""
        from agentflow._constants import is_agentflow_file

        f = tmp_path / "test.md"
        f.write_text("# Just a regular file")
        assert not is_agentflow_file(f)

    def test_is_agentflow_file_missing(self, tmp_path):
        """Test missing file."""
        from agentflow._constants import is_agentflow_file

        f = tmp_path / "nonexistent.md"
        assert not is_agentflow_file(f)


class TestBackup:
    def test_backup_user_file(self, tmp_path):
        """Test backing up user files."""
        from agentflow._constants import backup_user_file

        f = tmp_path / "CLAUDE.md"
        f.write_text("user content")
        backup = backup_user_file(f)
        assert backup.exists()
        assert "bak" in backup.name
        assert backup.read_text() == "user content"


class TestCLIDetection:
    def test_detect_installed_clis(self, mock_home):
        """Test detecting installed CLIs."""
        from agentflow._constants import detect_installed_clis

        detected = detect_installed_clis()
        assert len(detected) == len(CLI_TARGETS)

    def test_detect_installed_targets_empty(self, mock_home):
        """Test no AgentFlow installations detected."""
        from agentflow._constants import detect_installed_targets

        installed = detect_installed_targets()
        assert len(installed) == 0


class TestInstaller:
    def test_install_unknown_target(self, mock_home):
        """Test installing to unknown target."""
        from agentflow.installer import install

        assert not install("unknown_cli")

    def test_install_claude(self, mock_home):
        """Test installing to Claude Code directory."""
        from agentflow.installer import install

        agents_md = Path(__file__).parent.parent / "AGENTS.md"
        if not agents_md.exists():
            pytest.skip("AGENTS.md not found")

        result = install("claude")
        assert result

        rules_file = mock_home / ".claude" / "CLAUDE.md"
        assert rules_file.exists()
        content = rules_file.read_text()
        assert AGENTFLOW_MARKER in content

        plugin_dir = mock_home / ".claude" / PLUGIN_DIR_NAME
        assert plugin_dir.exists()

    def test_uninstall_claude(self, mock_home):
        """Test uninstalling from Claude Code."""
        from agentflow.installer import install, uninstall

        agents_md = Path(__file__).parent.parent / "AGENTS.md"
        if not agents_md.exists():
            pytest.skip("AGENTS.md not found")

        install("claude")
        result = uninstall("claude")
        assert result

        rules_file = mock_home / ".claude" / "CLAUDE.md"
        assert not rules_file.exists()

        plugin_dir = mock_home / ".claude" / PLUGIN_DIR_NAME
        assert not plugin_dir.exists()

    def test_install_preserves_user_files(self, mock_home):
        """Test that install backs up user's existing files."""
        from agentflow.installer import install

        agents_md = Path(__file__).parent.parent / "AGENTS.md"
        if not agents_md.exists():
            pytest.skip("AGENTS.md not found")

        rules_file = mock_home / ".claude" / "CLAUDE.md"
        rules_file.write_text("# User's custom CLAUDE.md")

        install("claude")

        backups = list((mock_home / ".claude").glob("*_bak*"))
        assert len(backups) >= 1


class TestInstallAll:
    def test_install_all(self, mock_home):
        """Test installing to all targets."""
        from agentflow.installer import install_all

        agents_md = Path(__file__).parent.parent / "AGENTS.md"
        if not agents_md.exists():
            pytest.skip("AGENTS.md not found")

        result = install_all()
        assert result
