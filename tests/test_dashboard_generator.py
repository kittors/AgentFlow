"""Tests for agentflow.scripts.dashboard_generator."""

from pathlib import Path

from agentflow.scripts.dashboard_generator import (
    _count_source_files,
    generate_dashboard,
)


class TestCountSourceFiles:
    def test_counts_python_files(self, tmp_path: Path):
        (tmp_path / "app.py").write_text("pass\n")
        (tmp_path / "utils.py").write_text("pass\n")
        assert _count_source_files(tmp_path) == 2

    def test_counts_multiple_extensions(self, tmp_path: Path):
        (tmp_path / "app.py").write_text("pass\n")
        (tmp_path / "index.ts").write_text("export {}\n")
        (tmp_path / "main.go").write_text("package main\n")
        assert _count_source_files(tmp_path) == 3

    def test_ignores_non_source_files(self, tmp_path: Path):
        (tmp_path / "readme.md").write_text("# Hi\n")
        (tmp_path / "data.csv").write_text("a,b\n")
        assert _count_source_files(tmp_path) == 0

    def test_empty_dir(self, tmp_path: Path):
        assert _count_source_files(tmp_path) == 0


class TestGenerateDashboard:
    def test_creates_html_file(self, tmp_path: Path):
        (tmp_path / ".agentflow" / "kb").mkdir(parents=True)
        path = generate_dashboard(tmp_path)
        assert path.exists()
        assert path.suffix == ".html"

    def test_html_contains_project_name(self, tmp_path: Path):
        (tmp_path / ".agentflow" / "kb").mkdir(parents=True)
        path = generate_dashboard(tmp_path)
        content = path.read_text()
        assert tmp_path.name in content

    def test_html_has_structure(self, tmp_path: Path):
        (tmp_path / ".agentflow" / "kb").mkdir(parents=True)
        path = generate_dashboard(tmp_path)
        content = path.read_text()
        assert "<!DOCTYPE html>" in content
        assert "AgentFlow Dashboard" in content
        assert "Modules" in content
        assert "Source Files" in content

    def test_with_modules(self, tmp_path: Path):
        kb = tmp_path / ".agentflow" / "kb"
        modules = kb / "modules"
        modules.mkdir(parents=True)
        (modules / "_index.md").write_text("# Index\n")
        (modules / "auth.md").write_text("# Auth\n")
        (modules / "api.md").write_text("# API\n")

        path = generate_dashboard(tmp_path)
        content = path.read_text()
        assert "auth" in content
        assert "api" in content

    def test_with_sessions(self, tmp_path: Path):
        kb = tmp_path / ".agentflow" / "kb"
        sessions = kb / "sessions"
        sessions.mkdir(parents=True)
        (sessions / "20260227_120000.md").write_text("# Session\n")

        path = generate_dashboard(tmp_path)
        content = path.read_text()
        assert "20260227_120000" in content

    def test_kb_status_active(self, tmp_path: Path):
        kb = tmp_path / ".agentflow" / "kb"
        kb.mkdir(parents=True)
        (kb / "INDEX.md").write_text("# Project\n")
        path = generate_dashboard(tmp_path)
        content = path.read_text()
        assert "Active" in content

    def test_kb_status_not_initialized(self, tmp_path: Path):
        (tmp_path / ".agentflow").mkdir()
        path = generate_dashboard(tmp_path)
        content = path.read_text()
        assert "Not initialized" in content
