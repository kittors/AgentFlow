"""Tests for updater.py — update, status, clean."""

from pathlib import Path
from unittest import mock

import pytest

from agentflow._constants import CLI_TARGETS, PLUGIN_DIR_NAME


# ── update() ──────────────────────────────────────────────────────────────────


class TestUpdate:
    def test_update_uv_success(self):
        from agentflow.updater import update

        mock_result = mock.Mock()
        mock_result.returncode = 0
        mock_result.stdout = ""

        with (
            mock.patch("agentflow.updater.detect_install_method", return_value="uv"),
            mock.patch("subprocess.run", return_value=mock_result),
            mock.patch("agentflow.updater.detect_installed_targets", return_value=[]),
        ):
            update()  # Should not raise

    def test_update_pip_success(self):
        from agentflow.updater import update

        mock_result = mock.Mock()
        mock_result.returncode = 0
        mock_result.stdout = ""

        with (
            mock.patch("agentflow.updater.detect_install_method", return_value="pip"),
            mock.patch("subprocess.run", return_value=mock_result),
            mock.patch("agentflow.updater.detect_installed_targets", return_value=[]),
        ):
            update()

    def test_update_failure(self, capsys):
        from agentflow.updater import update

        mock_result = mock.Mock()
        mock_result.returncode = 1
        mock_result.stderr = "error message"

        with (
            mock.patch("agentflow.updater.detect_install_method", return_value="pip"),
            mock.patch("subprocess.run", return_value=mock_result),
        ):
            update()
        out = capsys.readouterr().out
        assert "❌" in out

    def test_update_with_branch(self):
        from agentflow.updater import update

        mock_result = mock.Mock()
        mock_result.returncode = 0
        mock_result.stdout = ""

        with (
            mock.patch("agentflow.updater.detect_install_method", return_value="uv"),
            mock.patch("subprocess.run", return_value=mock_result) as mock_run,
            mock.patch("agentflow.updater.detect_installed_targets", return_value=[]),
        ):
            update(branch="dev")
        # Should use --from git+URL@dev
        call_args = mock_run.call_args[0][0]
        assert any("dev" in str(arg) for arg in call_args)

    def test_update_redeploys(self):
        """After successful update, should re-deploy to installed targets."""
        from agentflow.updater import update

        mock_result = mock.Mock()
        mock_result.returncode = 0
        mock_result.stdout = ""

        with (
            mock.patch("agentflow.updater.detect_install_method", return_value="pip"),
            mock.patch("subprocess.run", return_value=mock_result),
            mock.patch("agentflow.updater.detect_installed_targets", return_value=["codex"]),
            mock.patch("agentflow.installer.install") as mock_install,
        ):
            update()
            mock_install.assert_called_once_with("codex")

    def test_update_exception(self, capsys):
        from agentflow.updater import update

        with (
            mock.patch("agentflow.updater.detect_install_method", return_value="pip"),
            mock.patch("subprocess.run", side_effect=Exception("network error")),
        ):
            update()
        out = capsys.readouterr().out
        assert "❌" in out


# ── status() ──────────────────────────────────────────────────────────────────


class TestStatus:
    def test_status_output(self, capsys, tmp_path):
        from agentflow.updater import status

        for config in CLI_TARGETS.values():
            (tmp_path / config["dir"]).mkdir(parents=True, exist_ok=True)

        with mock.patch("pathlib.Path.home", return_value=tmp_path):
            status()

        out = capsys.readouterr().out
        # Should contain version info and CLI names
        assert "codex" in out.lower() or "claude" in out.lower()


# ── clean() ───────────────────────────────────────────────────────────────────


class TestClean:
    def test_clean_no_installs(self, capsys, tmp_path):
        from agentflow.updater import clean

        with mock.patch("pathlib.Path.home", return_value=tmp_path):
            clean()
        out = capsys.readouterr().out
        # Should indicate nothing to clean
        assert any(word in out for word in ("No", "无", "Nothing", "未"))

    def test_clean_with_cache(self, capsys, tmp_path):
        from agentflow.updater import clean
        from agentflow._constants import AGENTFLOW_MARKER

        # Simulate codex install with cache
        cli_dir = tmp_path / ".codex"
        cli_dir.mkdir()
        plugin_dir = cli_dir / PLUGIN_DIR_NAME
        plugin_dir.mkdir()
        cache_dir = plugin_dir / "__pycache__"
        cache_dir.mkdir()
        (cache_dir / "test.pyc").write_text("fake")

        # Mark as installed
        rules = cli_dir / "AGENTS.md"
        rules.write_text(f"<!-- {AGENTFLOW_MARKER} v1.0.0 -->\n# Test")

        with mock.patch("pathlib.Path.home", return_value=tmp_path):
            clean()
        out = capsys.readouterr().out
        assert "✅" in out
        assert not cache_dir.exists()
