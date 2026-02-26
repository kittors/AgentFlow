"""Tests for agentflow.scripts.convention_scanner."""

from pathlib import Path

from agentflow.scripts.convention_scanner import (
    _detect_naming_style,
    save_conventions,
    scan_python_conventions,
)


class TestDetectNamingStyle:
    def test_snake_case(self):
        assert _detect_naming_style(["foo_bar", "hello_world", "get_data"]) == "snake_case"

    def test_camel_case(self):
        assert _detect_naming_style(["fooBar", "helloWorld", "getData"]) == "camelCase"

    def test_pascal_case(self):
        assert _detect_naming_style(["FooBar", "HelloWorld", "MyClass"]) == "PascalCase"

    def test_upper_snake_case(self):
        assert _detect_naming_style(["FOO_BAR", "MAX_SIZE", "API_KEY"]) == "UPPER_SNAKE_CASE"

    def test_empty_list(self):
        assert _detect_naming_style([]) == "unknown"

    def test_mixed_defaults_to_majority(self):
        names = ["foo_bar", "baz_qux", "hello_world", "MyClass"]
        assert _detect_naming_style(names) == "snake_case"


class TestScanPythonConventions:
    def test_detects_function_naming(self, tmp_path: Path):
        src = tmp_path / "src"
        src.mkdir()
        (src / "mod.py").write_text(
            "def get_data():\n    pass\n\ndef set_value():\n    pass\n\ndef process_items():\n    pass\n"
        )
        result = scan_python_conventions(tmp_path, source_dirs=["src"])
        assert result["conventions"]["naming"]["functions"] == "snake_case"
        assert result["conventions"]["stats"]["functions_found"] == 3

    def test_detects_class_naming(self, tmp_path: Path):
        src = tmp_path / "src"
        src.mkdir()
        (src / "mod.py").write_text("class MyClass:\n    pass\n\nclass BaseHandler:\n    pass\n")
        result = scan_python_conventions(tmp_path, source_dirs=["src"])
        assert result["conventions"]["naming"]["classes"] == "PascalCase"
        assert result["conventions"]["stats"]["classes_found"] == 2

    def test_detects_import_style(self, tmp_path: Path):
        src = tmp_path / "src"
        src.mkdir()
        (src / "mod.py").write_text("from .utils import helper\nfrom . import base\nimport os\n")
        result = scan_python_conventions(tmp_path, source_dirs=["src"])
        assert result["conventions"]["imports"]["relative_count"] == 2
        assert result["conventions"]["imports"]["absolute_count"] >= 1

    def test_detects_google_docstrings(self, tmp_path: Path):
        src = tmp_path / "src"
        src.mkdir()
        (src / "mod.py").write_text(
            "def foo():\n"
            '    """Do something.\n'
            "\n"
            "    Args:\n"
            "        x: int\n"
            "    Returns:\n"
            "        str\n"
            '    """\n'
            "    pass\n"
        )
        result = scan_python_conventions(tmp_path, source_dirs=["src"])
        assert result["conventions"]["documentation"]["docstring_style"] == "google"

    def test_empty_project(self, tmp_path: Path):
        result = scan_python_conventions(tmp_path, source_dirs=["nonexistent"])
        assert result["conventions"]["stats"]["functions_found"] == 0

    def test_result_structure(self, tmp_path: Path):
        src = tmp_path / "src"
        src.mkdir()
        (src / "mod.py").write_text("x = 1\n")
        result = scan_python_conventions(tmp_path, source_dirs=["src"])
        assert "project" in result
        assert "extracted_at" in result
        assert "language" in result
        assert result["language"] == "python"
        assert "conventions" in result


class TestSaveConventions:
    def test_saves_to_disk(self, tmp_path: Path):
        conventions = {"project": "test", "conventions": {"naming": {}}}
        (tmp_path / ".agentflow").mkdir()
        path = save_conventions(tmp_path, conventions)
        assert path.exists()
        assert "conventions" in path.parent.name
        content = path.read_text()
        assert "test" in content
