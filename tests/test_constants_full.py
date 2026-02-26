"""Comprehensive tests for _constants.py — path helpers, CLI detection, install method."""

from unittest import mock

from agentflow._constants import (
    AGENTFLOW_MARKER,
    CLI_TARGETS,
    PLUGIN_DIR_NAME,
    backup_user_file,
    detect_install_method,
    detect_installed_clis,
    detect_installed_targets,
    detect_locale,
    get_agentflow_module_path,
    get_agents_md_path,
    get_package_root,
    get_skill_md_path,
    is_agentflow_file,
    msg,
)

# ── Path helpers ──────────────────────────────────────────────────────────────


class TestPathHelpers:
    def test_get_package_root_is_directory(self):
        root = get_package_root()
        assert root.is_dir()
        assert (root / "agentflow").is_dir()

    def test_get_agents_md_path(self):
        p = get_agents_md_path()
        assert p.name == "AGENTS.md"
        assert p.exists()

    def test_get_skill_md_path(self):
        p = get_skill_md_path()
        assert p.name == "SKILL.md"

    def test_get_agentflow_module_path(self):
        p = get_agentflow_module_path()
        assert p.name == "agentflow"
        assert (p / "__init__.py").exists()


# ── Locale detection (extended) ───────────────────────────────────────────────


class TestLocaleExtended:
    def test_lc_all_zh(self):
        with mock.patch.dict("os.environ", {"LC_ALL": "zh_TW.UTF-8"}, clear=True):
            assert detect_locale() == "zh"

    def test_lc_messages_zh(self):
        with mock.patch.dict("os.environ", {"LC_MESSAGES": "zh_HK"}, clear=True):
            assert detect_locale() == "zh"

    def test_language_zh(self):
        with mock.patch.dict("os.environ", {"LANGUAGE": "zh_CN"}, clear=True):
            assert detect_locale() == "zh"

    def test_fallback_en(self):
        with mock.patch.dict("os.environ", {"LANG": "fr_FR.UTF-8"}, clear=True):
            assert detect_locale() == "en"

    def test_empty_env(self):
        with mock.patch.dict("os.environ", {}, clear=True):
            # Should fall through to locale.getlocale
            result = detect_locale()
            assert result in ("zh", "en")


# ── msg() ─────────────────────────────────────────────────────────────────────


class TestMsgExtended:
    def test_returns_string(self):
        result = msg("中文", "English")
        assert isinstance(result, str)
        assert result in ("中文", "English")


# ── File identification & backup ──────────────────────────────────────────────


class TestFileIdentification:
    def test_agentflow_file_at_boundary(self, tmp_path):
        """Marker near the 1024-byte boundary."""
        f = tmp_path / "edge.md"
        padding = "x" * 1000
        f.write_text(f"{padding}<!-- {AGENTFLOW_MARKER} v1.0.0 -->")
        assert is_agentflow_file(f)

    def test_agentflow_file_beyond_boundary(self, tmp_path):
        """Marker after the 1024-byte scan window."""
        f = tmp_path / "far.md"
        padding = "x" * 2000
        f.write_text(f"{padding}<!-- {AGENTFLOW_MARKER} v1.0.0 -->")
        assert not is_agentflow_file(f)

    def test_is_directory(self, tmp_path):
        """Passing a directory should return False."""
        assert not is_agentflow_file(tmp_path)


class TestBackupExtended:
    def test_backup_preserves_content(self, tmp_path):
        f = tmp_path / "rules.md"
        f.write_text("important content")
        backup = backup_user_file(f)
        assert backup.read_text() == "important content"
        assert "bak" in backup.name
        # Original still exists
        assert f.exists()

    def test_backup_suffix_matches(self, tmp_path):
        f = tmp_path / "test.toml"
        f.write_text("data")
        backup = backup_user_file(f)
        assert backup.suffix == ".toml"


# ── CLI detection ─────────────────────────────────────────────────────────────


class TestCLIDetection:
    def test_detect_no_clis(self, tmp_path):
        """No CLI dirs exist → empty list."""
        with mock.patch("pathlib.Path.home", return_value=tmp_path):
            assert detect_installed_clis() == []

    def test_detect_partial_clis(self, tmp_path):
        """Only some CLI dirs exist."""
        (tmp_path / ".codex").mkdir()
        (tmp_path / ".claude").mkdir()
        with mock.patch("pathlib.Path.home", return_value=tmp_path):
            detected = detect_installed_clis()
            assert "codex" in detected
            assert "claude" in detected
            assert "gemini" not in detected

    def test_detect_installed_targets_none(self, tmp_path):
        """CLI dirs exist but no AgentFlow installed."""
        for config in CLI_TARGETS.values():
            (tmp_path / config["dir"]).mkdir(parents=True, exist_ok=True)
        with mock.patch("pathlib.Path.home", return_value=tmp_path):
            assert detect_installed_targets() == []

    def test_detect_installed_targets_with_install(self, tmp_path):
        """Simulate an AgentFlow install to codex."""
        cli_dir = tmp_path / ".codex"
        cli_dir.mkdir()
        # Create plugin dir and rules file with marker
        (cli_dir / PLUGIN_DIR_NAME).mkdir()
        rules = cli_dir / "AGENTS.md"
        rules.write_text(f"<!-- {AGENTFLOW_MARKER} v1.0.0 -->\n# Test")
        with mock.patch("pathlib.Path.home", return_value=tmp_path):
            installed = detect_installed_targets()
            assert "codex" in installed


# ── Install method detection ──────────────────────────────────────────────────


class TestInstallMethod:
    def test_uv_detected(self):
        mock_result = mock.Mock()
        mock_result.returncode = 0
        mock_result.stdout = "agentflow 1.0.0\n"
        with mock.patch("subprocess.run", return_value=mock_result):
            assert detect_install_method() == "uv"

    def test_pip_fallback(self):
        with mock.patch("subprocess.run", side_effect=FileNotFoundError):
            assert detect_install_method() == "pip"

    def test_uv_without_agentflow(self):
        mock_result = mock.Mock()
        mock_result.returncode = 0
        mock_result.stdout = "ruff 0.4.0\n"
        with mock.patch("subprocess.run", return_value=mock_result):
            assert detect_install_method() == "pip"
