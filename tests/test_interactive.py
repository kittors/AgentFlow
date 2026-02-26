"""Tests for interactive.py — interactive install/uninstall menus."""

from unittest import mock

import pytest

from agentflow._constants import CLI_TARGETS


# ── interactive_install ───────────────────────────────────────────────────────


class TestInteractiveInstall:
    def test_no_clis_detected(self, capsys):
        from agentflow.interactive import interactive_install

        with mock.patch("agentflow.interactive.detect_installed_clis", return_value=[]):
            result = interactive_install()
        assert result is False

    def test_cancel_with_0(self):
        from agentflow.interactive import interactive_install

        with (
            mock.patch("agentflow.interactive.detect_installed_clis", return_value=["codex"]),
            mock.patch("agentflow.interactive.detect_installed_targets", return_value=[]),
            mock.patch("builtins.input", return_value="0"),
        ):
            assert interactive_install() is False

    def test_cancel_empty_input(self):
        from agentflow.interactive import interactive_install

        with (
            mock.patch("agentflow.interactive.detect_installed_clis", return_value=["codex"]),
            mock.patch("agentflow.interactive.detect_installed_targets", return_value=[]),
            mock.patch("builtins.input", return_value=""),
        ):
            assert interactive_install() is False

    def test_eof_handling(self):
        from agentflow.interactive import interactive_install

        with (
            mock.patch("agentflow.interactive.detect_installed_clis", return_value=["codex"]),
            mock.patch("agentflow.interactive.detect_installed_targets", return_value=[]),
            mock.patch("builtins.input", side_effect=EOFError),
        ):
            assert interactive_install() is False

    def test_select_single(self):
        from agentflow.interactive import interactive_install

        with (
            mock.patch("agentflow.interactive.detect_installed_clis", return_value=["codex", "claude"]),
            mock.patch("agentflow.interactive.detect_installed_targets", return_value=[]),
            mock.patch("builtins.input", return_value="1"),
            mock.patch("agentflow.installer.install", return_value=True) as mock_install,
        ):
            result = interactive_install()
            assert result is True
            mock_install.assert_called_once_with("codex")

    def test_select_all(self):
        from agentflow.interactive import interactive_install

        with (
            mock.patch("agentflow.interactive.detect_installed_clis", return_value=["codex"]),
            mock.patch("agentflow.interactive.detect_installed_targets", return_value=[]),
            mock.patch("builtins.input", return_value="A"),
            mock.patch("agentflow.installer.install_all", return_value=True) as mock_all,
        ):
            result = interactive_install()
            assert result is True
            mock_all.assert_called_once()

    def test_select_by_name(self):
        from agentflow.interactive import interactive_install

        with (
            mock.patch("agentflow.interactive.detect_installed_clis", return_value=["codex", "claude"]),
            mock.patch("agentflow.interactive.detect_installed_targets", return_value=[]),
            mock.patch("builtins.input", return_value="claude"),
            mock.patch("agentflow.installer.install", return_value=True) as mock_install,
        ):
            result = interactive_install()
            assert result is True
            mock_install.assert_called_once_with("claude")

    def test_invalid_selection(self, capsys):
        from agentflow.interactive import interactive_install

        with (
            mock.patch("agentflow.interactive.detect_installed_clis", return_value=["codex"]),
            mock.patch("agentflow.interactive.detect_installed_targets", return_value=[]),
            mock.patch("builtins.input", return_value="99"),
        ):
            result = interactive_install()
            assert result is False

    def test_multi_select(self):
        from agentflow.interactive import interactive_install

        with (
            mock.patch("agentflow.interactive.detect_installed_clis", return_value=["codex", "claude", "gemini"]),
            mock.patch("agentflow.interactive.detect_installed_targets", return_value=[]),
            mock.patch("builtins.input", return_value="1,2"),
            mock.patch("agentflow.installer.install", return_value=True) as mock_install,
        ):
            result = interactive_install()
            assert result is True
            assert mock_install.call_count == 2


# ── interactive_uninstall ─────────────────────────────────────────────────────


class TestInteractiveUninstall:
    def test_no_installs(self, capsys):
        from agentflow.interactive import interactive_uninstall

        with mock.patch("agentflow.interactive.detect_installed_targets", return_value=[]):
            result = interactive_uninstall()
        assert result is False

    def test_cancel_with_0(self):
        from agentflow.interactive import interactive_uninstall

        with (
            mock.patch("agentflow.interactive.detect_installed_targets", return_value=["codex"]),
            mock.patch("builtins.input", return_value="0"),
        ):
            assert interactive_uninstall() is False

    def test_eof_handling(self):
        from agentflow.interactive import interactive_uninstall

        with (
            mock.patch("agentflow.interactive.detect_installed_targets", return_value=["codex"]),
            mock.patch("builtins.input", side_effect=KeyboardInterrupt),
        ):
            assert interactive_uninstall() is False

    def test_select_single(self):
        from agentflow.interactive import interactive_uninstall

        with (
            mock.patch("agentflow.interactive.detect_installed_targets", return_value=["codex"]),
            mock.patch("builtins.input", return_value="1"),
            mock.patch("agentflow.installer.uninstall", return_value=True) as mock_uninst,
        ):
            result = interactive_uninstall()
            assert result is True
            mock_uninst.assert_called_once_with("codex")

    def test_select_all(self):
        from agentflow.interactive import interactive_uninstall

        with (
            mock.patch("agentflow.interactive.detect_installed_targets", return_value=["codex", "claude"]),
            mock.patch("builtins.input", return_value="A"),
            mock.patch("agentflow.installer.uninstall_all", return_value=True) as mock_all,
        ):
            result = interactive_uninstall()
            assert result is True
            mock_all.assert_called_once()
