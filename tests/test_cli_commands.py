"""Tests for cli.py — main() dispatcher, help, version, unknown commands."""

from unittest import mock

import pytest

from agentflow.cli import main, print_usage


class TestPrintUsage:
    def test_contains_all_targets(self, capsys):
        print_usage()
        out = capsys.readouterr().out
        for target in ("codex", "claude", "gemini", "qwen", "grok", "opencode"):
            assert target in out

    def test_contains_commands(self, capsys):
        print_usage()
        out = capsys.readouterr().out
        for cmd in ("install", "uninstall", "update", "clean", "status", "version"):
            assert cmd in out


class TestMainHelp:
    @pytest.mark.parametrize("flag", ["--help", "-h", "help"])
    def test_help_flags(self, flag, capsys):
        with mock.patch("sys.argv", ["agentflow", flag]):
            with pytest.raises(SystemExit) as exc_info:
                main()
            assert exc_info.value.code == 0
        out = capsys.readouterr().out
        assert "install" in out


class TestMainVersion:
    def test_version_command(self):
        with mock.patch("sys.argv", ["agentflow", "version"]):
            with mock.patch("agentflow.version_check.check_update") as mock_check:
                try:
                    main()
                except SystemExit:
                    pass
                mock_check.assert_called_once_with(show_version=True)


class TestMainInstall:
    def test_install_target(self):
        with mock.patch("sys.argv", ["agentflow", "install", "codex"]):
            with mock.patch("agentflow.installer.install", return_value=True) as m:
                try:
                    main()
                except SystemExit:
                    pass
                m.assert_called_once_with("codex")

    def test_install_all(self):
        with mock.patch("sys.argv", ["agentflow", "install", "--all"]):
            with mock.patch("agentflow.installer.install_all", return_value=True) as m:
                try:
                    main()
                except SystemExit:
                    pass
                m.assert_called_once()

    def test_install_failure_exits_1(self):
        with mock.patch("sys.argv", ["agentflow", "install", "codex"]):
            with mock.patch("agentflow.installer.install", return_value=False):
                with pytest.raises(SystemExit) as exc_info:
                    main()
                assert exc_info.value.code == 1


class TestMainUninstall:
    def test_uninstall_target(self):
        with mock.patch("sys.argv", ["agentflow", "uninstall", "codex"]):
            with mock.patch("agentflow.installer.uninstall", return_value=True) as m:
                try:
                    main()
                except SystemExit:
                    pass
                m.assert_called_once_with("codex")

    def test_uninstall_all(self):
        with mock.patch("sys.argv", ["agentflow", "uninstall", "--all"]):
            with mock.patch("agentflow.installer.uninstall_all") as m:
                try:
                    main()
                except SystemExit:
                    pass
                m.assert_called_once()


class TestMainOtherCommands:
    def test_update(self):
        with mock.patch("sys.argv", ["agentflow", "update"]):
            with mock.patch("agentflow.updater.update") as m:
                try:
                    main()
                except SystemExit:
                    pass
                m.assert_called_once_with(None)

    def test_update_with_branch(self):
        with mock.patch("sys.argv", ["agentflow", "update", "dev"]):
            with mock.patch("agentflow.updater.update") as m:
                try:
                    main()
                except SystemExit:
                    pass
                m.assert_called_once_with("dev")

    def test_clean(self):
        with mock.patch("sys.argv", ["agentflow", "clean"]):
            with mock.patch("agentflow.updater.clean") as m:
                try:
                    main()
                except SystemExit:
                    pass
                m.assert_called_once()

    def test_status(self):
        with mock.patch("sys.argv", ["agentflow", "status"]):
            with mock.patch("agentflow.updater.status") as m:
                try:
                    main()
                except SystemExit:
                    pass
                m.assert_called_once()


class TestMainUnknownCommand:
    def test_unknown_exits_1(self, capsys):
        with mock.patch("sys.argv", ["agentflow", "foobar"]):
            with pytest.raises(SystemExit) as exc_info:
                main()
            assert exc_info.value.code == 1


class TestMainNoArgs:
    def test_no_args_calls_interactive(self):
        with mock.patch("sys.argv", ["agentflow"]):
            with mock.patch("agentflow.cli._interactive_main") as m:
                with pytest.raises(SystemExit) as exc_info:
                    main()
                m.assert_called_once()
                assert exc_info.value.code == 0
