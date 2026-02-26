"""AgentFlow shared constants and utility functions.

This module is the single source of truth for constants and low-level helpers
used across the package.  Every other module imports from here instead of
``cli.py``, which avoids circular-import issues and makes Pyre / Pyright happy.
"""

from __future__ import annotations

import locale
import os
import sys
from pathlib import Path

# ── Package metadata ──────────────────────────────────────────────────────────

__all__ = [
    "REPO_URL",
    "REPO_API_LATEST",
    "CLI_TARGETS",
    "PLUGIN_DIR_NAME",
    "AGENTFLOW_MARKER",
    "HOOKS_FINGERPRINT",
    "CODEX_NOTIFY_CMD",
    # locale helpers
    "detect_locale",
    "msg",
    # file helpers
    "is_agentflow_file",
    "backup_user_file",
    # path helpers
    "get_package_root",
    "get_agents_md_path",
    "get_skill_md_path",
    "get_agentflow_module_path",
    # CLI detection
    "detect_installed_clis",
    "detect_installed_targets",
    "detect_install_method",
]

REPO_URL: str = "https://github.com/kittors/AgentFlow"
REPO_API_LATEST: str = "https://api.github.com/repos/kittors/AgentFlow/releases/latest"

CLI_TARGETS: dict[str, dict[str, str]] = {
    "codex": {"dir": ".codex", "rules_file": "AGENTS.md"},
    "claude": {"dir": ".claude", "rules_file": "CLAUDE.md"},
    "gemini": {"dir": ".gemini", "rules_file": "GEMINI.md"},
    "qwen": {"dir": ".qwen", "rules_file": "QWEN.md"},
    "grok": {"dir": ".grok", "rules_file": "GROK.md"},
    "opencode": {"dir": ".config/opencode", "rules_file": "AGENTS.md"},
}

PLUGIN_DIR_NAME: str = "agentflow"

# Fingerprint marker to identify AgentFlow-created files
AGENTFLOW_MARKER: str = "AGENTFLOW_ROUTER:"

# Hooks identification
HOOKS_FINGERPRINT: str = "AgentFlow"
CODEX_NOTIFY_CMD: str = "agentflow --check-update --silent"


# ── Locale detection ──────────────────────────────────────────────────────────

def detect_locale() -> str:
    """Detect system locale.  Returns ``'zh'`` for Chinese locales, ``'en'``
    otherwise."""
    for var in ("LC_ALL", "LC_MESSAGES", "LANG", "LANGUAGE"):
        val = os.environ.get(var, "")
        if val.lower().startswith("zh"):
            return "zh"
    try:
        loc = locale.getlocale()[0] or ""
        if loc.lower().startswith("zh"):
            return "zh"
    except Exception:
        pass
    if sys.platform == "win32":
        try:
            import ctypes  # noqa: WPS433
            lcid = ctypes.windll.kernel32.GetUserDefaultUILanguage()  # type: ignore[attr-defined]
            if (lcid & 0xFF) == 0x04:
                return "zh"
        except Exception:
            pass
    return "en"


_LANG: str = detect_locale()


def msg(zh: str, en: str) -> str:
    """Return the *zh* or *en* string based on the detected locale."""
    return zh if _LANG == "zh" else en


# ── File identification & backup ──────────────────────────────────────────────

def is_agentflow_file(file_path: str | Path) -> bool:
    """Return ``True`` if *file_path* was created by AgentFlow."""
    try:
        full = Path(file_path).read_text(encoding="utf-8", errors="ignore")
        content = full[:1024]  # type: ignore[index]
        return AGENTFLOW_MARKER in content
    except Exception:
        return False


def backup_user_file(file_path: str | Path) -> Path:
    """Backup a non-AgentFlow file with a timestamp suffix.

    Returns the *Path* of the newly created backup.
    """
    import shutil
    from datetime import datetime

    file_path = Path(file_path)
    timestamp = datetime.now().strftime("%Y%m%d%H%M%S")
    backup_name = f"{file_path.stem}_{timestamp}_bak{file_path.suffix}"
    backup_path = file_path.parent / backup_name
    shutil.copy2(file_path, backup_path)
    return backup_path


# ── Package resource helpers ──────────────────────────────────────────────────

def get_package_root() -> Path:
    """Return the repository / package root directory."""
    return Path(__file__).parent.parent


def get_agents_md_path() -> Path:
    """Return the path to the ``AGENTS.md`` source file."""
    return get_package_root() / "AGENTS.md"


def get_skill_md_path() -> Path:
    """Return the path to the ``SKILL.md`` source file."""
    return get_package_root() / "SKILL.md"


def get_agentflow_module_path() -> Path:
    """Return the path to the ``agentflow/`` package directory."""
    return Path(__file__).parent


# ── CLI detection ─────────────────────────────────────────────────────────────

def detect_installed_clis() -> list[str]:
    """Detect which CLI config directories exist on disk."""
    installed: list[str] = []
    for name, config in CLI_TARGETS.items():
        cli_dir = Path.home() / config["dir"]
        if cli_dir.exists():
            installed.append(name)
    return installed


def detect_installed_targets() -> list[str]:
    """Detect which CLI targets already have AgentFlow installed."""
    installed: list[str] = []
    for name, config in CLI_TARGETS.items():
        cli_dir = Path.home() / config["dir"]
        plugin_dir = cli_dir / PLUGIN_DIR_NAME
        rules_file = cli_dir / config["rules_file"]
        if plugin_dir.exists() and rules_file.exists():
            if is_agentflow_file(rules_file):
                installed.append(name)
    return installed


def detect_install_method() -> str:
    """Detect whether ``agentflow`` was installed via *uv* or *pip*."""
    import subprocess

    try:
        result = subprocess.run(
            ["uv", "tool", "list"],
            capture_output=True,
            text=True,
            encoding="utf-8",
            errors="replace",
            timeout=5,
        )
        if result.returncode == 0 and "agentflow" in result.stdout:
            return "uv"
    except (FileNotFoundError, subprocess.TimeoutExpired):
        pass
    return "pip"
