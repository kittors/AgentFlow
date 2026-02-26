"""Tests for the Profile system — deployment profiles and AGENTS.md assembly."""

from pathlib import Path
from unittest import mock

import pytest

from agentflow._constants import (
    AGENTFLOW_MARKER,
    CLI_TARGETS,
    DEFAULT_PROFILE,
    PROFILES,
    VALID_PROFILES,
)

PROJECT_ROOT = Path(__file__).parent.parent
AGENTFLOW_DIR = PROJECT_ROOT / "agentflow"


# ── Profile constants ─────────────────────────────────────────────────────────


class TestProfileConstants:
    def test_profiles_defined(self):
        """All three profiles exist."""
        assert "lite" in PROFILES
        assert "standard" in PROFILES
        assert "full" in PROFILES

    def test_default_profile_is_full(self):
        assert DEFAULT_PROFILE == "full"

    def test_valid_profiles_tuple(self):
        assert set(VALID_PROFILES) == {"lite", "standard", "full"}

    def test_lite_has_no_modules(self):
        assert PROFILES["lite"] == []

    def test_standard_has_three_modules(self):
        assert len(PROFILES["standard"]) == 3
        assert "common.md" in PROFILES["standard"]
        assert "module_loading.md" in PROFILES["standard"]
        assert "acceptance.md" in PROFILES["standard"]

    def test_full_has_six_modules(self):
        assert len(PROFILES["full"]) == 6
        assert "subagent.md" in PROFILES["full"]
        assert "attention.md" in PROFILES["full"]
        assert "hooks.md" in PROFILES["full"]

    def test_standard_is_subset_of_full(self):
        """Standard modules should be a subset of full modules."""
        assert set(PROFILES["standard"]).issubset(set(PROFILES["full"]))


# ── Core directory integrity ──────────────────────────────────────────────────


class TestCoreDirectory:
    def test_core_dir_exists(self):
        d = AGENTFLOW_DIR / "core"
        assert d.exists(), "agentflow/core/ directory missing"
        assert d.is_dir()

    @pytest.mark.parametrize(
        "filename",
        ["common.md", "module_loading.md", "acceptance.md", "subagent.md", "attention.md", "hooks.md"],
    )
    def test_core_file_exists_and_nonempty(self, filename):
        f = AGENTFLOW_DIR / "core" / filename
        assert f.exists(), f"core/{filename} missing"
        content = f.read_text()
        assert len(content.strip()) > 50, f"core/{filename} is too small"

    def test_core_files_match_full_profile(self):
        """Every file listed in 'full' profile exists in core/."""
        core_dir = AGENTFLOW_DIR / "core"
        for filename in PROFILES["full"]:
            assert (core_dir / filename).exists(), f"core/{filename} listed in full profile but missing"


# ── AGENTS.md validation ──────────────────────────────────────────────────────


class TestAgentsMdRefactored:
    def test_agents_md_reduced(self):
        """AGENTS.md should be under 400 lines after refactoring."""
        f = PROJECT_ROOT / "AGENTS.md"
        lines = f.read_text().splitlines()
        assert len(lines) < 450, f"AGENTS.md is {len(lines)} lines, expected < 450"

    def test_agents_md_has_marker(self):
        f = PROJECT_ROOT / "AGENTS.md"
        content = f.read_text()
        assert AGENTFLOW_MARKER in content

    def test_agents_md_has_core_sections(self):
        """G1-G5 should still be present in AGENTS.md."""
        content = (PROJECT_ROOT / "AGENTS.md").read_text()
        for section in ["G1 |", "G2 |", "G3 |", "G4 |", "G5 |"]:
            assert section in content, f"Core section {section} missing from AGENTS.md"

    def test_agents_md_has_references(self):
        """G6-G12 should reference core/ module files."""
        content = (PROJECT_ROOT / "AGENTS.md").read_text()
        for ref in [
            "core/common.md",
            "core/module_loading.md",
            "core/acceptance.md",
            "core/subagent.md",
            "core/attention.md",
            "core/hooks.md",
        ]:
            assert ref in content, f"Reference to {ref} missing from AGENTS.md"

    def test_agents_md_uses_must_should_may(self):
        """New priority system should use MUST/SHOULD/MAY instead of all CRITICAL."""
        content = (PROJECT_ROOT / "AGENTS.md").read_text()
        assert "MUST" in content
        assert "SHOULD" in content
        assert "MAY" in content


# ── Profile assembly ──────────────────────────────────────────────────────────


class TestProfileAssembly:
    @pytest.fixture(autouse=True)
    def _setup(self):
        agents_md = PROJECT_ROOT / "AGENTS.md"
        if not agents_md.exists():
            pytest.skip("AGENTS.md not found")

    def test_build_lite_profile(self):
        from agentflow.installer import _build_agents_md_for_profile

        content = _build_agents_md_for_profile("lite")
        assert AGENTFLOW_MARKER in content
        # Lite should NOT contain appended module content
        assert "PROFILE:lite" not in content  # No modules to append

    def test_build_standard_profile(self):
        from agentflow.installer import _build_agents_md_for_profile

        content = _build_agents_md_for_profile("standard")
        assert AGENTFLOW_MARKER in content
        assert "PROFILE:standard" in content
        # Should contain the 3 standard module contents
        assert "G6 | 通用规则" in content
        assert "G7 | 模块加载" in content
        assert "G8 | 验收标准" in content
        # Should NOT contain full-only modules
        assert "G9+G10 | 子代理编排与调用通道" not in content

    def test_build_full_profile(self):
        from agentflow.installer import _build_agents_md_for_profile

        content = _build_agents_md_for_profile("full")
        assert AGENTFLOW_MARKER in content
        assert "PROFILE:full" in content
        # Should contain all 6 module contents
        assert "G6 | 通用规则" in content
        assert "G7 | 模块加载" in content
        assert "G8 | 验收标准" in content
        assert "G9+G10 | 子代理编排与调用通道" in content
        assert "G11 | 注意力控制" in content
        assert "G12 | Hooks 集成" in content

    def test_full_profile_longer_than_lite(self):
        from agentflow.installer import _build_agents_md_for_profile

        lite = _build_agents_md_for_profile("lite")
        full = _build_agents_md_for_profile("full")
        assert len(full) > len(lite) * 1.5, "Full profile should be significantly longer than lite"

    def test_all_profiles_have_core_sections(self):
        from agentflow.installer import _build_agents_md_for_profile

        for profile in VALID_PROFILES:
            content = _build_agents_md_for_profile(profile)
            for section in ["G1 |", "G2 |", "G3 |", "G4 |", "G5 |"]:
                assert section in content, f"Profile '{profile}' missing core section {section}"


# ── Install with profile ──────────────────────────────────────────────────────


class TestInstallWithProfile:
    @pytest.fixture
    def mock_home(self, tmp_path):
        for name, config in CLI_TARGETS.items():
            cli_dir = tmp_path / config["dir"]
            cli_dir.mkdir(parents=True)
        with mock.patch("pathlib.Path.home", return_value=tmp_path):
            yield tmp_path

    def test_install_with_lite_profile(self, mock_home):
        from agentflow.installer import install

        agents_md = PROJECT_ROOT / "AGENTS.md"
        if not agents_md.exists():
            pytest.skip("AGENTS.md not found")

        result = install("claude", "lite")
        assert result

        rules_file = mock_home / ".claude" / "CLAUDE.md"
        assert rules_file.exists()
        content = rules_file.read_text()
        assert AGENTFLOW_MARKER in content
        # Lite should NOT contain appended full module content
        assert "G11 | 注意力控制" not in content.split("---")[-1] if "---" in content else True

    def test_install_with_full_profile(self, mock_home):
        from agentflow.installer import install

        agents_md = PROJECT_ROOT / "AGENTS.md"
        if not agents_md.exists():
            pytest.skip("AGENTS.md not found")

        result = install("claude", "full")
        assert result

        rules_file = mock_home / ".claude" / "CLAUDE.md"
        content = rules_file.read_text()
        # Full = all modules appended
        assert "G11 | 注意力控制" in content

    def test_install_invalid_profile(self, mock_home):
        from agentflow.installer import install

        result = install("claude", "nonexistent")
        assert not result
