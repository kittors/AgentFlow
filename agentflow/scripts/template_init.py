"""AgentFlow template initializer — set up KB structure from project templates.

Detects the project type and copies the appropriate templates into the
``.agentflow/kb/`` directory.
"""

from __future__ import annotations

import shutil
from pathlib import Path

# ── Project type detection ────────────────────────────────────────────────────

_INDICATORS: dict[str, list[str]] = {
    "frontend": [
        "package.json",
        "vite.config.ts",
        "next.config.js",
        "next.config.ts",
        "nuxt.config.ts",
        "angular.json",
        "svelte.config.js",
    ],
    "backend": [
        "requirements.txt",
        "Pipfile",
        "go.mod",
        "Cargo.toml",
        "pom.xml",
        "build.gradle",
    ],
    "python": [
        "pyproject.toml",
        "setup.py",
        "setup.cfg",
    ],
    "fullstack": [],  # detected by combination
}


def detect_project_type(project_root: Path) -> str:
    """Detect project type based on configuration files.

    Returns one of: ``'frontend'``, ``'backend'``, ``'fullstack'``, ``'python'``.
    Falls back to ``'python'`` if no indicators match.
    """
    has_frontend = any((project_root / f).exists() for f in _INDICATORS["frontend"])
    has_backend = any((project_root / f).exists() for f in _INDICATORS["backend"])
    has_python = any((project_root / f).exists() for f in _INDICATORS["python"])

    if has_frontend and has_backend:
        return "fullstack"
    if has_frontend:
        return "frontend"
    if has_python:
        return "python"
    if has_backend:
        return "backend"
    return "python"


def get_template_path(template_name: str) -> Path | None:
    """Get the path to a template file from the agentflow package."""
    templates_dir = Path(__file__).parent.parent / "templates"
    template_file = templates_dir / f"{template_name}.md"
    return template_file if template_file.exists() else None


def init_from_template(project_root: Path, project_type: str | None = None) -> dict:
    """Initialize KB structure from the appropriate template.

    Args:
        project_root: Project root directory.
        project_type: Override auto-detection (frontend/backend/fullstack/python).

    Returns:
        Summary dict with ``project_type`` and ``files_created``.
    """
    ptype = project_type or detect_project_type(project_root)

    kb_root = project_root / ".agentflow" / "kb"
    kb_root.mkdir(parents=True, exist_ok=True)

    files_created: list[str] = []

    # Copy type-specific template
    type_template = f"{ptype}_project"
    template_path = get_template_path(type_template)
    if template_path:
        dest = kb_root / f"{type_template}_reference.md"
        shutil.copy2(template_path, dest)
        files_created.append(str(dest.relative_to(project_root)))

    # Always copy base templates
    for base in ["kb_templates", "conventions_template", "graph_template", "session_summary_template"]:
        bp = get_template_path(base)
        if bp:
            dest = kb_root / f"{base}_reference.md"
            if not dest.exists():
                shutil.copy2(bp, dest)
                files_created.append(str(dest.relative_to(project_root)))

    # Create standard directories
    for subdir in ["modules", "plan", "sessions", "graph", "conventions", "archive"]:
        (kb_root / subdir).mkdir(exist_ok=True)

    return {"project_type": ptype, "files_created": files_created}


if __name__ == "__main__":
    from .kb_sync import find_project_root

    root = find_project_root()
    if root:
        result = init_from_template(root)
        print(f"Initialized as {result['project_type']} project")
        for f in result["files_created"]:
            print(f"  Created: {f}")
    else:
        # Initialize in current directory
        cwd = Path.cwd()
        (cwd / ".agentflow").mkdir(exist_ok=True)
        result = init_from_template(cwd)
        print(f"Initialized as {result['project_type']} project")
