"""Tests for static config file integrity — JSON, TOML, Markdown validation."""

import json
from pathlib import Path

import pytest

PROJECT_ROOT = Path(__file__).parent.parent
AGENTFLOW_DIR = PROJECT_ROOT / "agentflow"


# ── JSON configs ──────────────────────────────────────────────────────────────


class TestJSONConfigs:
    def test_claude_hooks_valid_json(self):
        f = AGENTFLOW_DIR / "hooks" / "claude_hooks.json"
        assert f.exists(), "claude_hooks.json missing"
        data = json.loads(f.read_text())
        assert "hooks" in data
        assert isinstance(data["hooks"], list)
        assert len(data["hooks"]) > 0

    def test_claude_hooks_types(self):
        f = AGENTFLOW_DIR / "hooks" / "claude_hooks.json"
        data = json.loads(f.read_text())
        for hook in data["hooks"]:
            assert "type" in hook
            assert "command" in hook
            assert "description" in hook


# ── TOML configs ──────────────────────────────────────────────────────────────


class TestTOMLConfigs:
    @pytest.fixture(autouse=True)
    def _import_tomllib(self):
        try:
            import tomllib

            self.tomllib = tomllib
        except ModuleNotFoundError:
            import tomli as tomllib  # type: ignore

            self.tomllib = tomllib

    def test_codex_hooks_valid_toml(self):
        f = AGENTFLOW_DIR / "hooks" / "codex_hooks.toml"
        assert f.exists(), "codex_hooks.toml missing"
        data = self.tomllib.loads(f.read_text())
        assert "notify" in data

    def test_reviewer_toml(self):
        f = AGENTFLOW_DIR / "agents" / "reviewer.toml"
        assert f.exists()
        data = self.tomllib.loads(f.read_text())
        assert len(data) > 0

    def test_architect_toml(self):
        f = AGENTFLOW_DIR / "agents" / "architect.toml"
        assert f.exists()
        data = self.tomllib.loads(f.read_text())
        assert len(data) > 0

    def test_codex_agents_toml(self):
        f = AGENTFLOW_DIR / "agents" / "codex_agents.toml"
        assert f.exists()
        data = self.tomllib.loads(f.read_text())
        assert "agents" in data


# ── Markdown source files ─────────────────────────────────────────────────────


class TestMarkdownFiles:
    def test_agents_md_exists(self):
        f = PROJECT_ROOT / "AGENTS.md"
        assert f.exists()
        content = f.read_text()
        assert len(content) > 100

    def test_skill_md_exists(self):
        f = PROJECT_ROOT / "SKILL.md"
        assert f.exists()
        content = f.read_text()
        assert len(content) > 50

    @pytest.mark.parametrize(
        "subdir,min_files",
        [
            ("functions", 5),
            ("services", 3),
            ("stages", 2),
            ("rules", 3),
            ("core", 6),
        ],
    )
    def test_subdir_has_md_files(self, subdir, min_files):
        d = AGENTFLOW_DIR / subdir
        assert d.exists(), f"{subdir}/ missing"
        md_files = list(d.glob("*.md"))
        assert len(md_files) >= min_files, f"{subdir}/ has only {len(md_files)} .md files, expected >= {min_files}"

    def test_rlm_roles_exist(self):
        d = AGENTFLOW_DIR / "rlm" / "roles"
        assert d.exists()
        md_files = list(d.glob("*.md"))
        assert len(md_files) >= 4, f"rlm/roles/ has only {len(md_files)} roles"

    @pytest.mark.parametrize("role", ["architect", "reviewer", "synthesizer", "kb_keeper"])
    def test_rlm_role_nonempty(self, role):
        f = AGENTFLOW_DIR / "rlm" / "roles" / f"{role}.md"
        assert f.exists(), f"rlm/roles/{role}.md missing"
        assert len(f.read_text().strip()) > 20

    def test_templates_exist(self):
        d = AGENTFLOW_DIR / "templates"
        assert d.exists()
        files = list(d.glob("*.md"))
        assert len(files) >= 1

    @pytest.mark.parametrize(
        "function_name",
        [
            "init",
            "scan",
            "review",
            "plan",
            "exec",
            "auto",
            "status",
            "graph",
            "memory",
            "dashboard",
        ],
    )
    def test_function_md_nonempty(self, function_name):
        f = AGENTFLOW_DIR / "functions" / f"{function_name}.md"
        assert f.exists(), f"functions/{function_name}.md missing"
        assert len(f.read_text().strip()) > 20
