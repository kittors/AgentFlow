"""AgentFlow KB sync utility — synchronize code changes to knowledge base.

This script provides helpers for detecting code structure changes and
updating the project knowledge base accordingly.
"""

from __future__ import annotations

import json
from pathlib import Path
from datetime import datetime


def find_project_root(start: Path | None = None) -> Path | None:
    """Walk up from *start* to find a directory containing ``.agentflow/``."""
    current = start or Path.cwd()
    for parent in [current, *current.parents]:
        if (parent / ".agentflow").is_dir():
            return parent
    return None


def get_kb_root(project_root: Path) -> Path:
    """Return the knowledge base root directory."""
    return project_root / ".agentflow" / "kb"


def scan_modules(project_root: Path, source_dirs: list[str] | None = None) -> list[dict]:
    """Scan source directories and return a list of module descriptors.

    Each descriptor contains ``name``, ``path``, ``file_count``, and
    ``files`` (list of relative paths).
    """
    dirs = source_dirs or ["src", "lib", "app", project_root.name]
    modules: list[dict] = []

    for d in dirs:
        src = project_root / d
        if not src.is_dir():
            continue
        for child in sorted(src.iterdir()):
            if child.name.startswith((".", "_")):
                continue
            if child.is_dir():
                py_files = list(child.rglob("*.py"))
                ts_files = list(child.rglob("*.ts"))
                all_files = py_files + ts_files
                if all_files:
                    modules.append({
                        "name": child.name,
                        "path": str(child.relative_to(project_root)),
                        "file_count": len(all_files),
                        "files": [str(f.relative_to(project_root)) for f in all_files],
                    })
    return modules


def generate_module_index(modules: list[dict]) -> str:
    """Generate a Markdown index of modules."""
    lines = ["# Module Index\n"]
    for m in modules:
        lines.append(f"## {m['name']}")
        lines.append(f"- Path: `{m['path']}`")
        lines.append(f"- Files: {m['file_count']}")
        lines.append("")
    return "\n".join(lines)


def sync_kb(project_root: Path, source_dirs: list[str] | None = None) -> dict:
    """Synchronize knowledge base with current code structure.

    Returns a summary dict with ``modules_found`` and ``files_written``.
    """
    kb_root = get_kb_root(project_root)
    kb_root.mkdir(parents=True, exist_ok=True)
    modules_dir = kb_root / "modules"
    modules_dir.mkdir(exist_ok=True)

    modules = scan_modules(project_root, source_dirs)

    # Write module index
    index_content = generate_module_index(modules)
    (modules_dir / "_index.md").write_text(index_content, encoding="utf-8")

    files_written = 1
    for m in modules:
        content = f"# Module: {m['name']}\n\n"
        content += f"## Path\n`{m['path']}`\n\n"
        content += f"## Files ({m['file_count']})\n"
        for f in m["files"]:
            content += f"- `{f}`\n"
        (modules_dir / f"{m['name']}.md").write_text(content, encoding="utf-8")
        files_written += 1

    return {"modules_found": len(modules), "files_written": files_written}


if __name__ == "__main__":
    root = find_project_root()
    if root:
        result = sync_kb(root)
        print(f"KB synced: {result['modules_found']} modules, {result['files_written']} files written")
    else:
        print("No .agentflow/ directory found. Run ~init first.")
