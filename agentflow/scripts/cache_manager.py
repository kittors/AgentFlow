"""AgentFlow cache manager — manage sub-agent result caches.

Caches live under ``.agentflow/kb/cache/`` and expire at the end of each
session.  This module provides helpers for reading, writing, and cleaning
cached sub-agent results.
"""

from __future__ import annotations

import json
from datetime import datetime, timezone
from pathlib import Path


def get_cache_dir(project_root: Path) -> Path:
    """Return the cache directory for sub-agent results."""
    return project_root / ".agentflow" / "kb" / "cache"


def write_cache(project_root: Path, key: str, data: dict | str) -> Path:
    """Write a cache entry.

    Args:
        project_root: Project root with ``.agentflow/``.
        key: Cache key (e.g. ``"scan_result"``, ``"review_result"``).
        data: JSON-serializable dict or raw string content.

    Returns:
        Path to the cache file.
    """
    cache_dir = get_cache_dir(project_root)
    cache_dir.mkdir(parents=True, exist_ok=True)

    ext = ".json" if isinstance(data, dict) else ".md"
    cache_file = cache_dir / f"{key}{ext}"

    if isinstance(data, dict):
        data["_cached_at"] = datetime.now(timezone.utc).isoformat()
        cache_file.write_text(json.dumps(data, indent=2, ensure_ascii=False), encoding="utf-8")
    else:
        cache_file.write_text(data, encoding="utf-8")

    return cache_file


def read_cache(project_root: Path, key: str) -> dict | str | None:
    """Read a cache entry by key. Returns ``None`` if not found."""
    cache_dir = get_cache_dir(project_root)

    for ext in (".json", ".md"):
        cache_file = cache_dir / f"{key}{ext}"
        if cache_file.exists():
            content = cache_file.read_text(encoding="utf-8")
            if ext == ".json":
                return json.loads(content)
            return content
    return None


def list_cache(project_root: Path) -> list[dict]:
    """List all cache entries with metadata."""
    cache_dir = get_cache_dir(project_root)
    if not cache_dir.is_dir():
        return []

    entries: list[dict] = []
    for f in sorted(cache_dir.iterdir()):
        if f.is_file() and not f.name.startswith("."):
            entries.append(
                {
                    "key": f.stem,
                    "type": f.suffix.lstrip("."),
                    "size_bytes": f.stat().st_size,
                    "path": str(f),
                }
            )
    return entries


def clear_cache(project_root: Path) -> int:
    """Remove all cache files. Returns count of files removed."""
    cache_dir = get_cache_dir(project_root)
    if not cache_dir.is_dir():
        return 0

    count = 0
    for f in cache_dir.iterdir():
        if f.is_file():
            f.unlink()
            count += 1
    return count


def invalidate_cache(project_root: Path, key: str) -> bool:
    """Remove a specific cache entry. Returns ``True`` if removed."""
    cache_dir = get_cache_dir(project_root)
    for ext in (".json", ".md"):
        cache_file = cache_dir / f"{key}{ext}"
        if cache_file.exists():
            cache_file.unlink()
            return True
    return False


if __name__ == "__main__":
    from .kb_sync import find_project_root

    root = find_project_root()
    if root:
        entries = list_cache(root)
        print(f"Cache entries: {len(entries)}")
        for e in entries:
            print(f"  {e['key']}.{e['type']} — {e['size_bytes']}B")
    else:
        print("No .agentflow/ directory found.")
