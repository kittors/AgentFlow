"""Tests for agentflow.scripts.session_manager."""

import json
from pathlib import Path

from agentflow.scripts.session_manager import (
    create_session,
    export_sessions,
    get_sessions_dir,
    list_sessions,
    load_latest_session,
    prune_sessions,
    save_stage_snapshot,
)


class TestGetSessionsDir:
    def test_returns_correct_path(self, tmp_path: Path):
        d = get_sessions_dir(tmp_path)
        assert d == tmp_path / ".agentflow" / "sessions"


class TestCreateSession:
    def test_creates_session_file(self, tmp_path: Path):
        path = create_session(tmp_path)
        assert path.exists()
        assert path.suffix == ".md"
        content = path.read_text()
        assert "# Session:" in content
        assert "## Tasks" in content
        assert "## Decisions" in content

    def test_with_session_data(self, tmp_path: Path):
        path = create_session(
            tmp_path,
            tasks=["Implement feature X", "Fix bug Y"],
            decisions=["Use PostgreSQL"],
            files_modified=["src/main.py", "tests/test_main.py"],
            next_steps=["Deploy to staging"],
        )
        content = path.read_text()
        assert "feature X" in content
        assert "PostgreSQL" in content
        assert "src/main.py" in content
        assert "Deploy to staging" in content

    def test_with_stage(self, tmp_path: Path):
        path = create_session(tmp_path, stage="DESIGN")
        content = path.read_text()
        assert "Stage: DESIGN" in content

    def test_custom_session_id(self, tmp_path: Path):
        path = create_session(tmp_path, session_id="custom_123")
        assert path.stem == "custom_123"
        content = path.read_text()
        assert "custom_123" in content

    def test_session_id_in_filename(self, tmp_path: Path):
        path = create_session(tmp_path)
        # Filename should be timestamp-based like 20260227_120000.md
        assert len(path.stem) == 15  # YYYYMMDD_HHMMSS


class TestLoadLatestSession:
    def test_loads_latest(self, tmp_path: Path):
        sessions_dir = get_sessions_dir(tmp_path)
        sessions_dir.mkdir(parents=True)
        (sessions_dir / "20260101_000000.md").write_text("old session\n")
        (sessions_dir / "20260201_000000.md").write_text("new session\n")

        result = load_latest_session(tmp_path)
        assert result is not None
        assert result["id"] == "20260201_000000"
        assert "new session" in result["content"]

    def test_returns_none_when_empty(self, tmp_path: Path):
        result = load_latest_session(tmp_path)
        assert result is None

    def test_returns_none_no_dir(self, tmp_path: Path):
        result = load_latest_session(tmp_path)
        assert result is None


class TestSaveStageSnapshot:
    def test_creates_snapshot_file(self, tmp_path: Path):
        path = save_stage_snapshot(
            tmp_path,
            current_stage="DESIGN",
            task_progress="3/10",
            context="Chose React for frontend",
        )
        assert path.exists()
        assert path.stem.endswith("_snap")
        content = path.read_text()
        assert "Stage: DESIGN" in content
        assert "Progress: 3/10" in content
        assert "React for frontend" in content

    def test_snapshot_without_context(self, tmp_path: Path):
        path = save_stage_snapshot(
            tmp_path,
            current_stage="DEVELOP",
            task_progress="5/5",
        )
        assert path.exists()
        content = path.read_text()
        assert "DEVELOP" in content
        assert "## Context" not in content


class TestListSessions:
    def test_lists_sessions(self, tmp_path: Path):
        create_session(tmp_path, session_id="sess_001")
        create_session(tmp_path, session_id="sess_002")
        sessions = list_sessions(tmp_path)
        assert len(sessions) >= 2

    def test_session_structure(self, tmp_path: Path):
        create_session(tmp_path)
        sessions = list_sessions(tmp_path)
        s = sessions[0]
        assert "id" in s
        assert "path" in s
        assert "size" in s

    def test_empty_dir(self, tmp_path: Path):
        sessions = list_sessions(tmp_path)
        assert sessions == []

    def test_sorted_newest_first(self, tmp_path: Path):
        sessions_dir = get_sessions_dir(tmp_path)
        sessions_dir.mkdir(parents=True)
        (sessions_dir / "20260101_000000.md").write_text("old\n")
        (sessions_dir / "20260201_000000.md").write_text("new\n")
        sessions = list_sessions(tmp_path)
        assert sessions[0]["id"] == "20260201_000000"


class TestPruneSessions:
    def test_keeps_recent_sessions(self, tmp_path: Path):
        sessions_dir = get_sessions_dir(tmp_path)
        sessions_dir.mkdir(parents=True)
        for i in range(5):
            (sessions_dir / f"2026010{i}_000000.md").write_text(f"session {i}\n")

        removed = prune_sessions(tmp_path, keep=3)
        assert removed == 2
        remaining = list(sessions_dir.glob("*.md"))
        assert len(remaining) == 3

    def test_no_pruning_needed(self, tmp_path: Path):
        sessions_dir = get_sessions_dir(tmp_path)
        sessions_dir.mkdir(parents=True)
        (sessions_dir / "20260101_000000.md").write_text("session\n")
        removed = prune_sessions(tmp_path, keep=10)
        assert removed == 0

    def test_empty_dir(self, tmp_path: Path):
        removed = prune_sessions(tmp_path, keep=5)
        assert removed == 0


class TestExportSessions:
    def test_json_export(self, tmp_path: Path):
        create_session(tmp_path)
        output = export_sessions(tmp_path, output_format="json")
        data = json.loads(output)
        assert isinstance(data, list)
        assert len(data) >= 1

    def test_text_export(self, tmp_path: Path):
        create_session(tmp_path, session_id="test_sess")
        output = export_sessions(tmp_path, output_format="text")
        assert "test_sess" in output
