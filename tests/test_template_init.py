"""Tests for agentflow.scripts.template_init."""

from pathlib import Path

from agentflow.scripts.template_init import (
    detect_project_type,
    get_template_path,
    init_from_template,
)


class TestDetectProjectType:
    def test_detects_frontend(self, tmp_path: Path):
        (tmp_path / "package.json").write_text("{}\n")
        assert detect_project_type(tmp_path) == "frontend"

    def test_detects_frontend_vite(self, tmp_path: Path):
        (tmp_path / "vite.config.ts").write_text("export default {}\n")
        assert detect_project_type(tmp_path) == "frontend"

    def test_detects_backend(self, tmp_path: Path):
        (tmp_path / "requirements.txt").write_text("flask\n")
        assert detect_project_type(tmp_path) == "backend"

    def test_detects_backend_go(self, tmp_path: Path):
        (tmp_path / "go.mod").write_text("module example.com/app\n")
        assert detect_project_type(tmp_path) == "backend"

    def test_detects_python(self, tmp_path: Path):
        (tmp_path / "pyproject.toml").write_text("[project]\nname = 'test'\n")
        assert detect_project_type(tmp_path) == "python"

    def test_detects_fullstack(self, tmp_path: Path):
        (tmp_path / "package.json").write_text("{}\n")
        (tmp_path / "requirements.txt").write_text("django\n")
        assert detect_project_type(tmp_path) == "fullstack"

    def test_defaults_to_python(self, tmp_path: Path):
        assert detect_project_type(tmp_path) == "python"


class TestGetTemplatePath:
    def test_existing_template(self):
        path = get_template_path("python_project")
        assert path is not None
        assert path.exists()

    def test_nonexistent_template(self):
        path = get_template_path("nonexistent_template_xyz")
        assert path is None

    def test_all_project_templates_exist(self):
        for name in ["python_project", "frontend_project", "backend_project", "fullstack_project"]:
            path = get_template_path(name)
            assert path is not None, f"Template {name} is missing"


class TestInitFromTemplate:
    def test_creates_kb_structure(self, tmp_path: Path):
        (tmp_path / ".agentflow").mkdir()
        (tmp_path / "pyproject.toml").write_text("[project]\n")

        result = init_from_template(tmp_path)
        assert result["project_type"] == "python"
        assert len(result["files_created"]) > 0

        kb = tmp_path / ".agentflow" / "kb"
        assert kb.is_dir()
        assert (kb / "modules").is_dir()
        assert (kb / "plan").is_dir()
        assert (kb / "sessions").is_dir()
        assert (kb / "graph").is_dir()
        assert (kb / "conventions").is_dir()
        assert (kb / "archive").is_dir()

    def test_override_project_type(self, tmp_path: Path):
        (tmp_path / ".agentflow").mkdir()
        result = init_from_template(tmp_path, project_type="frontend")
        assert result["project_type"] == "frontend"

    def test_idempotent_base_templates(self, tmp_path: Path):
        (tmp_path / ".agentflow").mkdir()
        result1 = init_from_template(tmp_path)
        result2 = init_from_template(tmp_path)
        # Base templates should not be re-created if already present
        assert len(result2["files_created"]) <= len(result1["files_created"])

    def test_auto_detects_project_type(self, tmp_path: Path):
        (tmp_path / ".agentflow").mkdir()
        (tmp_path / "package.json").write_text("{}\n")
        result = init_from_template(tmp_path)
        assert result["project_type"] == "frontend"
