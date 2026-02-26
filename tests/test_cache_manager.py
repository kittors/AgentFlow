"""Tests for agentflow.scripts.cache_manager."""

from pathlib import Path

from agentflow.scripts.cache_manager import (
    clear_cache,
    get_cache_dir,
    invalidate_cache,
    list_cache,
    read_cache,
    write_cache,
)


class TestGetCacheDir:
    def test_returns_correct_path(self, tmp_path: Path):
        d = get_cache_dir(tmp_path)
        assert d == tmp_path / ".agentflow" / "kb" / "cache"


class TestWriteAndReadCache:
    def test_write_dict_creates_json(self, tmp_path: Path):
        data = {"modules": ["auth", "api"], "count": 2}
        path = write_cache(tmp_path, "scan_result", data)
        assert path.suffix == ".json"
        assert path.exists()

    def test_write_string_creates_md(self, tmp_path: Path):
        path = write_cache(tmp_path, "review_result", "# Review\nAll good.")
        assert path.suffix == ".md"
        assert path.exists()

    def test_read_dict_cache(self, tmp_path: Path):
        data = {"modules": ["auth"], "count": 1}
        write_cache(tmp_path, "test_key", data)
        result = read_cache(tmp_path, "test_key")
        assert isinstance(result, dict)
        assert result["modules"] == ["auth"]

    def test_read_string_cache(self, tmp_path: Path):
        write_cache(tmp_path, "review", "# Review content")
        result = read_cache(tmp_path, "review")
        assert isinstance(result, str)
        assert "Review content" in result

    def test_read_missing_key(self, tmp_path: Path):
        result = read_cache(tmp_path, "nonexistent")
        assert result is None

    def test_dict_cache_has_timestamp(self, tmp_path: Path):
        write_cache(tmp_path, "ts_test", {"value": 1})
        result = read_cache(tmp_path, "ts_test")
        assert "_cached_at" in result


class TestListCache:
    def test_lists_entries(self, tmp_path: Path):
        write_cache(tmp_path, "scan", {"x": 1})
        write_cache(tmp_path, "review", "text")
        entries = list_cache(tmp_path)
        assert len(entries) == 2
        keys = {e["key"] for e in entries}
        assert "scan" in keys
        assert "review" in keys

    def test_entry_structure(self, tmp_path: Path):
        write_cache(tmp_path, "test", {"x": 1})
        entries = list_cache(tmp_path)
        e = entries[0]
        assert "key" in e
        assert "type" in e
        assert "size_bytes" in e
        assert "path" in e

    def test_empty_cache(self, tmp_path: Path):
        entries = list_cache(tmp_path)
        assert entries == []


class TestClearCache:
    def test_clears_all(self, tmp_path: Path):
        write_cache(tmp_path, "a", {"x": 1})
        write_cache(tmp_path, "b", "text")
        count = clear_cache(tmp_path)
        assert count == 2
        assert list_cache(tmp_path) == []

    def test_empty_cache(self, tmp_path: Path):
        count = clear_cache(tmp_path)
        assert count == 0


class TestInvalidateCache:
    def test_removes_specific_key(self, tmp_path: Path):
        write_cache(tmp_path, "keep", {"x": 1})
        write_cache(tmp_path, "remove", {"y": 2})
        result = invalidate_cache(tmp_path, "remove")
        assert result is True
        assert read_cache(tmp_path, "remove") is None
        assert read_cache(tmp_path, "keep") is not None

    def test_nonexistent_key(self, tmp_path: Path):
        result = invalidate_cache(tmp_path, "nope")
        assert result is False
