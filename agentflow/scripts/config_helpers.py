"""AgentFlow config helpers — read and manipulate TOML/JSON configuration files.

Provides safe read/write operations for CLI config files with merge support.
"""

from __future__ import annotations

import json
from pathlib import Path


def read_toml(path: Path) -> dict:
    """Read a TOML file and return its contents as a dict."""
    try:
        import tomllib
    except ModuleNotFoundError:
        import tomli as tomllib  # type: ignore[no-redef]

    if not path.exists():
        return {}
    return tomllib.loads(path.read_text(encoding="utf-8"))


def write_toml_section(path: Path, section: str, data: dict) -> None:
    """Append or update a TOML section in a config file.

    This is a simple append-only strategy.  For full TOML manipulation,
    consider using ``tomli-w`` or ``tomlkit``.
    """
    content = path.read_text(encoding="utf-8") if path.exists() else ""

    section_header = f"[{section}]"
    if section_header in content:
        return  # section already exists, skip

    lines = [f"\n{section_header}"]
    for key, value in data.items():
        if isinstance(value, bool):
            lines.append(f"{key} = {'true' if value else 'false'}")
        elif isinstance(value, str):
            lines.append(f'{key} = "{value}"')
        elif isinstance(value, (int, float)):
            lines.append(f"{key} = {value}")

    content += "\n".join(lines) + "\n"
    path.write_text(content, encoding="utf-8")


def read_json(path: Path) -> dict:
    """Read a JSON file and return its contents as a dict."""
    if not path.exists():
        return {}
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except (json.JSONDecodeError, OSError):
        return {}


def write_json(path: Path, data: dict) -> None:
    """Write a dict to a JSON file with pretty formatting."""
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(data, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")


def merge_json(path: Path, updates: dict) -> dict:
    """Read a JSON file, merge updates (shallow), and write back.

    Returns the merged result.
    """
    existing = read_json(path)
    existing.update(updates)
    write_json(path, existing)
    return existing


def get_cli_config_path(target: str) -> Path | None:
    """Get the config file path for a CLI target."""
    from .._constants import CLI_TARGETS

    if target not in CLI_TARGETS:
        return None
    config = CLI_TARGETS[target]
    return Path.home() / config["dir"]


if __name__ == "__main__":
    print("AgentFlow config helpers loaded.")
    print("Available functions: read_toml, write_toml_section, read_json, write_json, merge_json")
