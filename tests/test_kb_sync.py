"""Tests for agentflow.scripts.kb_sync."""

from pathlib import Path

from agentflow.scripts.kb_sync import (
    find_project_root,
    generate_module_index,
    get_kb_root,
    scan_modules,
    sync_kb,
)


class TestFindProjectRoot:
    def test_finds_root_in_current(self, tmp_path: Path):
        (tmp_path / ".agentflow").mkdir()
        root = find_project_root(tmp_path)
        assert root == tmp_path

    def test_finds_root_in_parent(self, tmp_path: Path):
        (tmp_path / ".agentflow").mkdir()
        child = tmp_path / "src" / "pkg"
        child.mkdir(parents=True)
        root = find_project_root(child)
        assert root == tmp_path

    def test_returns_none_when_not_found(self, tmp_path: Path):
        root = find_project_root(tmp_path)
        assert root is None


class TestGetKbRoot:
    def test_returns_correct_path(self, tmp_path: Path):
        kb = get_kb_root(tmp_path)
        assert kb == tmp_path / ".agentflow" / "kb"


class TestScanModules:
    def test_finds_modules(self, tmp_path: Path):
        pkg = tmp_path / "src" / "auth"
        pkg.mkdir(parents=True)
        (pkg / "handler.py").write_text("pass\n")
        (pkg / "models.py").write_text("pass\n")

        modules = scan_modules(tmp_path, source_dirs=["src"])
        assert len(modules) == 1
        assert modules[0]["name"] == "auth"
        assert modules[0]["file_count"] == 2

    def test_ignores_hidden_dirs(self, tmp_path: Path):
        src = tmp_path / "src"
        hidden = src / ".cache"
        hidden.mkdir(parents=True)
        (hidden / "data.py").write_text("pass\n")
        modules = scan_modules(tmp_path, source_dirs=["src"])
        assert modules == []

    def test_ignores_underscore_dirs(self, tmp_path: Path):
        src = tmp_path / "src"
        pycache = src / "__pycache__"
        pycache.mkdir(parents=True)
        (pycache / "mod.cpython-312.pyc").write_bytes(b"data")
        modules = scan_modules(tmp_path, source_dirs=["src"])
        assert modules == []

    def test_empty_dir(self, tmp_path: Path):
        modules = scan_modules(tmp_path, source_dirs=["nonexistent"])
        assert modules == []

    def test_module_descriptor_structure(self, tmp_path: Path):
        pkg = tmp_path / "src" / "api"
        pkg.mkdir(parents=True)
        (pkg / "routes.py").write_text("pass\n")
        modules = scan_modules(tmp_path, source_dirs=["src"])
        m = modules[0]
        assert "name" in m
        assert "path" in m
        assert "file_count" in m
        assert "files" in m


class TestGenerateModuleIndex:
    def test_generates_markdown(self):
        modules = [
            {"name": "auth", "path": "src/auth", "file_count": 3, "files": []},
            {"name": "api", "path": "src/api", "file_count": 2, "files": []},
        ]
        md = generate_module_index(modules)
        assert "# Module Index" in md
        assert "## auth" in md
        assert "## api" in md
        assert "Files: 3" in md

    def test_empty_modules(self):
        md = generate_module_index([])
        assert "# Module Index" in md


class TestSyncKb:
    def test_creates_kb_structure(self, tmp_path: Path):
        pkg = tmp_path / "src" / "core"
        pkg.mkdir(parents=True)
        (pkg / "main.py").write_text("pass\n")

        result = sync_kb(tmp_path, source_dirs=["src"])
        assert result["modules_found"] == 1
        assert result["files_written"] >= 2  # _index.md + core.md

        kb = tmp_path / ".agentflow" / "kb"
        assert (kb / "modules" / "_index.md").exists()
        assert (kb / "modules" / "core.md").exists()

    def test_no_sources(self, tmp_path: Path):
        result = sync_kb(tmp_path, source_dirs=["nonexistent"])
        assert result["modules_found"] == 0
        assert result["files_written"] == 1  # just _index.md
