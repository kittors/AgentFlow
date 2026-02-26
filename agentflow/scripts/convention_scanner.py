"""AgentFlow convention scanner — extract coding conventions from a codebase.

Scans source files to detect naming patterns, import styles, and other
coding conventions, then writes the results to ``conventions/extracted.json``.
"""

from __future__ import annotations

import json
import re
from datetime import datetime, timezone
from pathlib import Path


def _detect_naming_style(names: list[str]) -> str:
    """Heuristic to determine the dominant naming style."""
    snake = sum(1 for n in names if re.match(r"^[a-z][a-z0-9_]*$", n))
    camel = sum(1 for n in names if re.match(r"^[a-z][a-zA-Z0-9]*$", n) and "_" not in n)
    pascal = sum(1 for n in names if re.match(r"^[A-Z][a-zA-Z0-9]*$", n))
    upper = sum(1 for n in names if re.match(r"^[A-Z][A-Z0-9_]*$", n))

    votes = {"snake_case": snake, "camelCase": camel, "PascalCase": pascal, "UPPER_SNAKE_CASE": upper}
    return max(votes, key=votes.get) if any(votes.values()) else "unknown"


def scan_python_conventions(project_root: Path, source_dirs: list[str] | None = None) -> dict:
    """Scan Python files and extract coding conventions."""
    dirs = source_dirs or ["src", "lib", "app", project_root.name]
    func_names: list[str] = []
    class_names: list[str] = []
    const_names: list[str] = []
    import_styles: dict[str, int] = {"absolute": 0, "relative": 0}
    docstring_styles: dict[str, int] = {"google": 0, "numpy": 0, "sphinx": 0, "none": 0}

    for d in dirs:
        src = project_root / d
        if not src.is_dir():
            continue
        for f in src.rglob("*.py"):
            try:
                content = f.read_text(encoding="utf-8", errors="ignore")
            except Exception:
                continue

            # Function names
            func_names.extend(re.findall(r"^def\s+(\w+)\s*\(", content, re.MULTILINE))
            # Class names
            class_names.extend(re.findall(r"^class\s+(\w+)", content, re.MULTILINE))
            # Constants (top-level UPPER_CASE assignments)
            const_names.extend(re.findall(r"^([A-Z][A-Z0-9_]{2,})\s*[:=]", content, re.MULTILINE))
            # Import styles
            import_styles["relative"] += len(re.findall(r"^from\s+\.", content, re.MULTILINE))
            import_styles["absolute"] += len(re.findall(r"^(?:from|import)\s+[a-zA-Z]", content, re.MULTILINE))
            # Docstring style detection
            if "Args:" in content or "Returns:" in content:
                docstring_styles["google"] += 1
            if "Parameters\n" in content or "----------" in content:
                docstring_styles["numpy"] += 1
            if ":param " in content or ":type " in content:
                docstring_styles["sphinx"] += 1

    result = {
        "project": project_root.name,
        "extracted_at": datetime.now(timezone.utc).isoformat(),
        "language": "python",
        "conventions": {
            "naming": {
                "functions": _detect_naming_style(func_names),
                "classes": _detect_naming_style(class_names) if class_names else "PascalCase",
                "constants": _detect_naming_style(const_names) if const_names else "UPPER_SNAKE_CASE",
            },
            "imports": {
                "style": max(import_styles, key=import_styles.get) if any(import_styles.values()) else "unknown",
                "absolute_count": import_styles["absolute"],
                "relative_count": import_styles["relative"],
            },
            "documentation": {
                "docstring_style": max(docstring_styles, key=docstring_styles.get),
            },
            "stats": {
                "functions_found": len(func_names),
                "classes_found": len(class_names),
                "constants_found": len(const_names),
            },
        },
    }
    return result


def save_conventions(project_root: Path, conventions: dict) -> Path:
    """Save extracted conventions to disk."""
    conv_dir = project_root / ".agentflow" / "kb" / "conventions"
    conv_dir.mkdir(parents=True, exist_ok=True)
    out = conv_dir / "extracted.json"
    out.write_text(json.dumps(conventions, indent=2, ensure_ascii=False), encoding="utf-8")
    return out


if __name__ == "__main__":
    from .kb_sync import find_project_root

    root = find_project_root()
    if root:
        conv = scan_python_conventions(root)
        path = save_conventions(root, conv)
        print(f"Conventions extracted: {conv['conventions']['stats']}")
        print(f"Saved to: {path}")
    else:
        print("No .agentflow/ directory found. Run ~init first.")
