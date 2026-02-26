"""Tests for agentflow.scripts.config_helpers."""

import json
from pathlib import Path

from agentflow.scripts.config_helpers import (
    merge_json,
    read_json,
    read_toml,
    write_json,
    write_toml_section,
)


class TestReadToml:
    def test_reads_valid_toml(self, tmp_path: Path):
        f = tmp_path / "config.toml"
        f.write_text('[section]\nkey = "value"\n')
        data = read_toml(f)
        assert data["section"]["key"] == "value"

    def test_missing_file(self, tmp_path: Path):
        f = tmp_path / "missing.toml"
        data = read_toml(f)
        assert data == {}


class TestWriteTomlSection:
    def test_appends_section(self, tmp_path: Path):
        f = tmp_path / "config.toml"
        f.write_text("# existing\n")
        write_toml_section(f, "new_section", {"key": "val", "flag": True, "num": 42})
        content = f.read_text()
        assert "[new_section]" in content
        assert 'key = "val"' in content
        assert "flag = true" in content
        assert "num = 42" in content

    def test_skips_existing_section(self, tmp_path: Path):
        f = tmp_path / "config.toml"
        f.write_text("[existing]\nkey = 1\n")
        write_toml_section(f, "existing", {"key": 2})
        content = f.read_text()
        assert content.count("[existing]") == 1

    def test_creates_file_if_missing(self, tmp_path: Path):
        f = tmp_path / "new.toml"
        write_toml_section(f, "sec", {"x": "y"})
        assert f.exists()
        assert "[sec]" in f.read_text()


class TestReadJson:
    def test_reads_valid_json(self, tmp_path: Path):
        f = tmp_path / "data.json"
        f.write_text('{"key": "value"}\n')
        data = read_json(f)
        assert data["key"] == "value"

    def test_missing_file(self, tmp_path: Path):
        data = read_json(tmp_path / "missing.json")
        assert data == {}

    def test_invalid_json(self, tmp_path: Path):
        f = tmp_path / "bad.json"
        f.write_text("not json")
        data = read_json(f)
        assert data == {}


class TestWriteJson:
    def test_writes_pretty_json(self, tmp_path: Path):
        f = tmp_path / "out.json"
        write_json(f, {"key": "value", "list": [1, 2, 3]})
        assert f.exists()
        data = json.loads(f.read_text())
        assert data["key"] == "value"
        assert data["list"] == [1, 2, 3]

    def test_creates_parent_dirs(self, tmp_path: Path):
        f = tmp_path / "sub" / "dir" / "out.json"
        write_json(f, {"x": 1})
        assert f.exists()


class TestMergeJson:
    def test_merges_into_existing(self, tmp_path: Path):
        f = tmp_path / "data.json"
        f.write_text('{"a": 1, "b": 2}\n')
        result = merge_json(f, {"b": 3, "c": 4})
        assert result == {"a": 1, "b": 3, "c": 4}
        on_disk = json.loads(f.read_text())
        assert on_disk == {"a": 1, "b": 3, "c": 4}

    def test_merge_into_nonexistent(self, tmp_path: Path):
        f = tmp_path / "new.json"
        result = merge_json(f, {"key": "val"})
        assert result == {"key": "val"}
        assert f.exists()
