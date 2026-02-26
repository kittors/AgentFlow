"""Tests for the CLI module."""

from unittest import mock

from agentflow._constants import (
    CLI_TARGETS,
    detect_locale,
    msg,
)


class TestLocaleDetection:
    def test_detect_chinese(self):
        """Test Chinese locale detection."""
        with mock.patch.dict("os.environ", {"LANG": "zh_CN.UTF-8"}):
            assert detect_locale() == "zh"

    def test_detect_english(self):
        """Test English locale detection."""
        with mock.patch.dict("os.environ", {"LANG": "en_US.UTF-8"}, clear=True):
            assert detect_locale() == "en"


class TestMsg:
    def test_msg_zh(self):
        """Test message in Chinese."""
        with mock.patch("agentflow._constants._LANG", "zh"):
            assert msg("中文", "English") == "中文"

    def test_msg_en(self):
        """Test message in English."""
        with mock.patch("agentflow._constants._LANG", "en"):
            assert msg("中文", "English") == "English"


class TestCLITargets:
    def test_six_targets(self):
        """Test that we support 6 CLI targets."""
        assert len(CLI_TARGETS) == 6
        assert "claude" in CLI_TARGETS
        assert "codex" in CLI_TARGETS
        assert "gemini" in CLI_TARGETS
        assert "qwen" in CLI_TARGETS
        assert "grok" in CLI_TARGETS
        assert "opencode" in CLI_TARGETS

    def test_targets_have_dir_and_rules(self):
        """Test that each target has dir and rules_file."""
        for name, config in CLI_TARGETS.items():
            assert "dir" in config, f"{name} missing dir"
            assert "rules_file" in config, f"{name} missing rules_file"
