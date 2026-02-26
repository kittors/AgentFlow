"""AgentFlow version check — check for updates via GitHub API."""

from __future__ import annotations

import json
import time
from pathlib import Path

from ._constants import REPO_API_LATEST, msg

_CACHE_DIR = Path.home() / ".cache" / "agentflow"
_CACHE_FILE = _CACHE_DIR / "version_cache.json"


def _read_cache() -> dict | None:
    """Read cached version info."""
    try:
        if _CACHE_FILE.exists():
            return json.loads(_CACHE_FILE.read_text(encoding="utf-8"))
    except Exception:
        pass
    return None


def _write_cache(data: dict) -> None:
    """Write version info to cache."""
    try:
        _CACHE_DIR.mkdir(parents=True, exist_ok=True)
        _CACHE_FILE.write_text(json.dumps(data, ensure_ascii=False), encoding="utf-8")
    except Exception:
        pass


def _fetch_latest_version() -> str | None:
    """Fetch latest version from GitHub API."""
    import urllib.error
    import urllib.request

    try:
        req = urllib.request.Request(
            REPO_API_LATEST,
            headers={"Accept": "application/vnd.github.v3+json", "User-Agent": "agentflow"},
        )
        with urllib.request.urlopen(req, timeout=5) as resp:
            data = json.loads(resp.read().decode("utf-8"))
            tag = data.get("tag_name", "")
            return tag.lstrip("v")
    except Exception:
        return None


def check_update(
    force: bool = False,
    cache_ttl_hours: int | None = None,
    show_version: bool = False,
) -> None:
    """Check for AgentFlow updates."""
    from importlib.metadata import version as get_version

    try:
        current = get_version("agentflow")
    except Exception:
        current = "unknown"

    if show_version:
        print(msg(f"  AgentFlow v{current}", f"  AgentFlow v{current}"))

    ttl = cache_ttl_hours if cache_ttl_hours is not None else 72
    if ttl == 0:
        return

    if not force:
        cache = _read_cache()
        if cache:
            cached_time = cache.get("timestamp", 0)
            if time.time() - cached_time < ttl * 3600:
                latest = cache.get("latest", "")
                if latest and latest != current:
                    print(msg(f"  ⬆️ 新版本可用: v{latest} (当前 v{current})",
                              f"  ⬆️ Update available: v{latest} (current v{current})"))
                    print(msg("     运行 agentflow update 更新",
                              "     Run: agentflow update"))
                return

    latest = _fetch_latest_version()
    if latest:
        _write_cache({"latest": latest, "timestamp": time.time()})
        if latest != current:
            print(msg(f"  ⬆️ 新版本可用: v{latest} (当前 v{current})",
                      f"  ⬆️ Update available: v{latest} (current v{current})"))
            print(msg("     运行 agentflow update 更新",
                      "     Run: agentflow update"))
