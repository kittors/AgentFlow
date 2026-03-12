"""AgentFlow session manager — create, list, prune and export session summaries.

Sessions are stored at ``.agentflow/sessions/`` (project root level, NOT under
``kb/``).  Only cross-session shared data (modules, plans, conventions) lives
under ``kb/``.
"""

from __future__ import annotations

import json
from datetime import datetime
from pathlib import Path


def get_sessions_dir(project_root: Path) -> Path:
    """Return the sessions directory for the project."""
    return project_root / ".agentflow" / "sessions"


def create_session(
    project_root: Path,
    *,
    tasks: list[str] | None = None,
    decisions: list[str] | None = None,
    issues: list[str] | None = None,
    next_steps: list[str] | None = None,
    files_modified: list[str] | None = None,
    stage: str | None = None,
    session_id: str | None = None,
) -> Path:
    """Create a new session summary file.

    Returns the *Path* of the written file.
    """
    sessions_dir = get_sessions_dir(project_root)
    sessions_dir.mkdir(parents=True, exist_ok=True)

    now = datetime.now()
    sid = session_id or now.strftime("%Y%m%d_%H%M%S")

    lines = [
        f"# Session: {sid}",
        f"Date: {now.strftime('%Y-%m-%d %H:%M:%S')}",
    ]
    if stage:
        lines.append(f"Stage: {stage}")
    lines.append("")

    def _section(title: str, items: list[str] | None) -> None:
        lines.append(f"## {title}")
        if items:
            for item in items:
                lines.append(f"- {item}")
        else:
            lines.append("- (none)")
        lines.append("")

    _section("Tasks", tasks)
    _section("Decisions", decisions)
    _section("Issues", issues)
    _section("Files Modified", files_modified)
    _section("Next Steps", next_steps)

    session_file = sessions_dir / f"{sid}.md"
    session_file.write_text("\n".join(lines), encoding="utf-8")
    return session_file


def load_latest_session(project_root: Path) -> dict | None:
    """Load the most recent session summary.

    Returns a dict with ``id``, ``path`` and ``content`` keys, or *None* if no
    sessions exist.
    """
    sessions_dir = get_sessions_dir(project_root)
    if not sessions_dir.is_dir():
        return None

    files = sorted(sessions_dir.glob("*.md"), reverse=True)
    if not files:
        return None

    latest = files[0]
    return {
        "id": latest.stem,
        "path": str(latest),
        "content": latest.read_text(encoding="utf-8"),
    }


def save_stage_snapshot(
    project_root: Path,
    *,
    current_stage: str,
    task_progress: str,
    context: str = "",
) -> Path:
    """Save a lightweight stage-transition snapshot.

    These snapshots live alongside full session summaries and are named with a
    ``_snap`` suffix so they can be distinguished easily.
    """
    sessions_dir = get_sessions_dir(project_root)
    sessions_dir.mkdir(parents=True, exist_ok=True)

    now = datetime.now()
    snap_id = now.strftime("%Y%m%d_%H%M%S") + "_snap"

    lines = [
        f"# Stage Snapshot: {current_stage}",
        f"Date: {now.strftime('%Y-%m-%d %H:%M:%S')}",
        f"Stage: {current_stage}",
        f"Progress: {task_progress}",
        "",
    ]
    if context:
        lines.extend(["## Context", context, ""])

    snap_file = sessions_dir / f"{snap_id}.md"
    snap_file.write_text("\n".join(lines), encoding="utf-8")
    return snap_file


def list_sessions(project_root: Path) -> list[dict]:
    """List all session summaries, newest first."""
    sessions_dir = get_sessions_dir(project_root)
    if not sessions_dir.is_dir():
        return []

    sessions: list[dict] = []
    for f in sorted(sessions_dir.glob("*.md"), reverse=True):
        stat = f.stat()
        sessions.append(
            {
                "id": f.stem,
                "path": str(f),
                "size": stat.st_size,
            }
        )
    return sessions


def prune_sessions(project_root: Path, keep: int = 20) -> int:
    """Remove old session files, keeping the most recent *keep* sessions.

    Returns the number of sessions removed.
    """
    sessions_dir = get_sessions_dir(project_root)
    if not sessions_dir.is_dir():
        return 0

    all_sessions = sorted(sessions_dir.glob("*.md"), reverse=True)
    to_remove = all_sessions[keep:]
    for f in to_remove:
        f.unlink(missing_ok=True)
    return len(to_remove)


def export_sessions(project_root: Path, output_format: str = "json") -> str:
    """Export session list as JSON or plain text."""
    sessions = list_sessions(project_root)

    if output_format == "json":
        return json.dumps(sessions, indent=2, ensure_ascii=False)

    # Plain text format
    lines = []
    for s in sessions:
        lines.append(f"[{s['id']}] {s['path']} ({s['size']}B)")
    return "\n".join(lines)


if __name__ == "__main__":
    from .kb_sync import find_project_root

    root = find_project_root()
    if root:
        sessions = list_sessions(root)
        print(f"Found {len(sessions)} sessions")
        for s in sessions[:5]:
            print(f"  {s['id']} ({s['size']}B)")
    else:
        print("No project root found")
