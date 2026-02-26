"""AgentFlow session manager — manage L2 session summaries.

Provides create, list, prune, and export operations for session memory.
"""

from __future__ import annotations

import json
from datetime import datetime, timezone
from pathlib import Path


def get_sessions_dir(project_root: Path) -> Path:
    """Return the sessions directory for the project."""
    return project_root / ".agentflow" / "kb" / "sessions"


def create_session(project_root: Path, session_data: dict | None = None) -> Path:
    """Create a new session summary file.

    Args:
        project_root: Project root containing ``.agentflow/``.
        session_data: Optional dict with ``tasks``, ``decisions``, ``files_modified``.

    Returns:
        Path to the created session file.
    """
    sessions_dir = get_sessions_dir(project_root)
    sessions_dir.mkdir(parents=True, exist_ok=True)

    now = datetime.now(timezone.utc)
    session_id = now.strftime("%Y%m%d_%H%M%S")
    data = session_data or {}

    content = f"# Session: {session_id}\n\n"
    content += f"Date: {now.strftime('%Y-%m-%d %H:%M UTC')}\n\n"
    content += "## Tasks\n"
    content += data.get("tasks", "- (no tasks recorded)\n") + "\n\n"
    content += "## Decisions\n"
    content += data.get("decisions", "- (no decisions recorded)\n") + "\n\n"
    content += "## Files Modified\n"
    for f in data.get("files_modified", []):
        content += f"- `{f}`\n"
    if not data.get("files_modified"):
        content += "- (none)\n"
    content += "\n## Next Steps\n"
    content += data.get("next_steps", "- (none)\n")

    session_file = sessions_dir / f"{session_id}.md"
    session_file.write_text(content, encoding="utf-8")
    return session_file


def list_sessions(project_root: Path) -> list[dict]:
    """List all session summaries with metadata."""
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
                "size_bytes": stat.st_size,
                "modified": datetime.fromtimestamp(stat.st_mtime, tz=timezone.utc).isoformat(),
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
        f.unlink()
    return len(to_remove)


def export_sessions(project_root: Path, output_format: str = "json") -> str:
    """Export all session metadata as JSON or Markdown."""
    sessions = list_sessions(project_root)

    if output_format == "json":
        return json.dumps(sessions, indent=2, ensure_ascii=False)

    # Markdown format
    lines = ["# Session History\n"]
    for s in sessions:
        lines.append(f"- **{s['id']}** — {s['modified']} ({s['size_bytes']}B)")
    return "\n".join(lines)


if __name__ == "__main__":
    from .kb_sync import find_project_root

    root = find_project_root()
    if root:
        sessions = list_sessions(root)
        print(f"Found {len(sessions)} sessions")
        for s in sessions[:5]:
            print(f"  {s['id']} — {s['modified']}")
    else:
        print("No .agentflow/ directory found.")
