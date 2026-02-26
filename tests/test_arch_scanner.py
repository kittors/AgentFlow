"""Tests for agentflow.scripts.arch_scanner."""

from pathlib import Path

from agentflow.scripts.arch_scanner import (
    LARGE_FILE_BYTES,
    LARGE_FILE_LINES,
    MAX_FUNCTION_LINES,
    full_scan,
    scan_circular_imports,
    scan_large_files,
    scan_long_functions,
    scan_missing_tests,
)

# ── scan_large_files ──────────────────────────────────────────────────────────


class TestScanLargeFiles:
    def test_detects_large_file_by_lines(self, tmp_path: Path):
        src = tmp_path / "src"
        src.mkdir()
        big = src / "big.py"
        big.write_text("\n" * (LARGE_FILE_LINES + 10))
        results = scan_large_files(tmp_path, source_dirs=["src"])
        assert len(results) == 1
        assert results[0]["lines"] > LARGE_FILE_LINES

    def test_detects_large_file_by_bytes(self, tmp_path: Path):
        src = tmp_path / "src"
        src.mkdir()
        big = src / "big.bin"
        big.write_bytes(b"x" * (LARGE_FILE_BYTES + 100))
        results = scan_large_files(tmp_path, source_dirs=["src"])
        assert len(results) == 1
        assert results[0]["size_bytes"] > LARGE_FILE_BYTES

    def test_ignores_small_files(self, tmp_path: Path):
        src = tmp_path / "src"
        src.mkdir()
        (src / "small.py").write_text("x = 1\n")
        results = scan_large_files(tmp_path, source_dirs=["src"])
        assert results == []

    def test_ignores_hidden_files(self, tmp_path: Path):
        src = tmp_path / "src"
        src.mkdir()
        (src / ".hidden.py").write_text("\n" * 1000)
        results = scan_large_files(tmp_path, source_dirs=["src"])
        assert results == []

    def test_skips_missing_dirs(self, tmp_path: Path):
        results = scan_large_files(tmp_path, source_dirs=["nonexistent"])
        assert results == []


# ── scan_missing_tests ────────────────────────────────────────────────────────


class TestScanMissingTests:
    def test_finds_untested_modules(self, tmp_path: Path):
        src = tmp_path / "src"
        src.mkdir()
        (src / "foo.py").write_text("def foo(): pass\n")
        (src / "bar.py").write_text("def bar(): pass\n")
        tests = tmp_path / "tests"
        tests.mkdir()
        (tests / "test_foo.py").write_text("def test_foo(): pass\n")
        missing = scan_missing_tests(tmp_path, source_dirs=["src"])
        assert any("bar.py" in m for m in missing)
        assert not any("foo.py" in m for m in missing)

    def test_ignores_private_files(self, tmp_path: Path):
        src = tmp_path / "src"
        src.mkdir()
        (src / "_internal.py").write_text("pass\n")
        (src / ".hidden.py").write_text("pass\n")
        missing = scan_missing_tests(tmp_path, source_dirs=["src"])
        assert missing == []

    def test_no_sources_returns_empty(self, tmp_path: Path):
        missing = scan_missing_tests(tmp_path, source_dirs=["nonexistent"])
        assert missing == []


# ── scan_circular_imports ─────────────────────────────────────────────────────


class TestScanCircularImports:
    def test_no_cycles_in_clean_project(self, tmp_path: Path):
        src = tmp_path / "src"
        src.mkdir()
        (src / "a.py").write_text("import os\n")
        (src / "b.py").write_text("import sys\n")
        cycles = scan_circular_imports(tmp_path, source_dirs=["src"])
        assert cycles == []

    def test_empty_project(self, tmp_path: Path):
        cycles = scan_circular_imports(tmp_path, source_dirs=["nonexistent"])
        assert cycles == []


# ── scan_long_functions ───────────────────────────────────────────────────────


class TestScanLongFunctions:
    def test_detects_long_function(self, tmp_path: Path):
        src = tmp_path / "src"
        src.mkdir()
        lines = ["def long_func():\n"] + [f"    x = {i}\n" for i in range(MAX_FUNCTION_LINES + 10)]
        (src / "mod.py").write_text("".join(lines))
        results = scan_long_functions(tmp_path, source_dirs=["src"])
        assert len(results) >= 1
        assert results[0]["function"] == "long_func"

    def test_ignores_short_functions(self, tmp_path: Path):
        src = tmp_path / "src"
        src.mkdir()
        (src / "mod.py").write_text("def short():\n    pass\n")
        results = scan_long_functions(tmp_path, source_dirs=["src"])
        assert results == []


# ── full_scan ─────────────────────────────────────────────────────────────────


class TestFullScan:
    def test_returns_all_sections(self, tmp_path: Path):
        (tmp_path / "src").mkdir()
        report = full_scan(tmp_path)
        assert "large_files" in report
        assert "missing_tests" in report
        assert "circular_imports" in report
        assert "long_functions" in report
