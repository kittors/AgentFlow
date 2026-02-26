"""AgentFlow architecture scanner — detect structural issues in a codebase.

Detects large files, circular dependencies (Python imports), missing test
coverage, and common code smells.
"""

from __future__ import annotations

import re
from pathlib import Path


# ── Thresholds ────────────────────────────────────────────────────────────────

LARGE_FILE_LINES = 500
LARGE_FILE_BYTES = 50_000
MAX_FUNCTION_LINES = 80


def scan_large_files(project_root: Path, source_dirs: list[str] | None = None) -> list[dict]:
    """Find files exceeding size thresholds."""
    dirs = source_dirs or ["src", "lib", "app", project_root.name]
    results: list[dict] = []

    for d in dirs:
        src = project_root / d
        if not src.is_dir():
            continue
        for f in src.rglob("*"):
            if not f.is_file() or f.name.startswith("."):
                continue
            size = f.stat().st_size
            try:
                lines = len(f.read_text(encoding="utf-8", errors="ignore").splitlines())
            except Exception:
                lines = 0

            issues: list[str] = []
            if size > LARGE_FILE_BYTES:
                issues.append(f"size={size}B (>{LARGE_FILE_BYTES}B)")
            if lines > LARGE_FILE_LINES:
                issues.append(f"lines={lines} (>{LARGE_FILE_LINES})")
            if issues:
                results.append({
                    "file": str(f.relative_to(project_root)),
                    "size_bytes": size,
                    "lines": lines,
                    "issues": issues,
                })
    return results


def scan_missing_tests(project_root: Path, source_dirs: list[str] | None = None) -> list[str]:
    """Find source modules that lack corresponding test files."""
    dirs = source_dirs or ["src", "lib", "app", project_root.name]
    test_dirs = ["tests", "test"]
    missing: list[str] = []

    # Collect existing test filenames
    test_files: set[str] = set()
    for td in test_dirs:
        test_path = project_root / td
        if test_path.is_dir():
            for f in test_path.rglob("test_*.py"):
                # test_foo.py → foo
                test_files.add(f.stem.removeprefix("test_"))

    # Check source modules
    for d in dirs:
        src = project_root / d
        if not src.is_dir():
            continue
        for f in src.rglob("*.py"):
            if f.name.startswith(("_", ".")):
                continue
            module_name = f.stem
            if module_name not in test_files:
                missing.append(str(f.relative_to(project_root)))

    return missing


def scan_circular_imports(project_root: Path, source_dirs: list[str] | None = None) -> list[list[str]]:
    """Detect circular import chains in Python files (simple heuristic)."""
    dirs = source_dirs or ["src", "lib", "app", project_root.name]
    # Build import graph
    graph: dict[str, set[str]] = {}

    for d in dirs:
        src = project_root / d
        if not src.is_dir():
            continue
        for f in src.rglob("*.py"):
            rel = str(f.relative_to(project_root))
            imports: set[str] = set()
            try:
                content = f.read_text(encoding="utf-8", errors="ignore")
                for match in re.findall(r"^(?:from|import)\s+([\w.]+)", content, re.MULTILINE):
                    imports.add(match.split(".")[0])
            except Exception:
                continue
            graph[rel] = imports

    # Simple cycle detection (DFS)
    cycles: list[list[str]] = []
    visited: set[str] = set()
    rec_stack: set[str] = set()
    path: list[str] = []

    def _dfs(node: str) -> None:
        visited.add(node)
        rec_stack.add(node)
        path.append(node)

        for neighbor_key, neighbor_imports in graph.items():
            if neighbor_key == node:
                continue
            neighbor_module = Path(neighbor_key).stem
            if neighbor_module in graph.get(node, set()):
                if neighbor_key in rec_stack:
                    idx = path.index(neighbor_key)
                    cycles.append(path[idx:] + [neighbor_key])
                elif neighbor_key not in visited:
                    _dfs(neighbor_key)

        path.pop()
        rec_stack.discard(node)

    for node in graph:
        if node not in visited:
            _dfs(node)

    return cycles


def scan_long_functions(project_root: Path, source_dirs: list[str] | None = None) -> list[dict]:
    """Find functions exceeding the line count threshold."""
    dirs = source_dirs or ["src", "lib", "app", project_root.name]
    results: list[dict] = []

    for d in dirs:
        src = project_root / d
        if not src.is_dir():
            continue
        for f in src.rglob("*.py"):
            try:
                lines = f.read_text(encoding="utf-8", errors="ignore").splitlines()
            except Exception:
                continue

            current_func: str | None = None
            func_start = 0
            for i, line in enumerate(lines):
                match = re.match(r"^(\s*)def\s+(\w+)\s*\(", line)
                if match:
                    # Check previous function
                    if current_func and (i - func_start) > MAX_FUNCTION_LINES:
                        results.append({
                            "file": str(f.relative_to(project_root)),
                            "function": current_func,
                            "lines": i - func_start,
                            "start_line": func_start + 1,
                        })
                    current_func = match.group(2)
                    func_start = i

            # Check last function
            if current_func and (len(lines) - func_start) > MAX_FUNCTION_LINES:
                results.append({
                    "file": str(f.relative_to(project_root)),
                    "function": current_func,
                    "lines": len(lines) - func_start,
                    "start_line": func_start + 1,
                })

    return results


def full_scan(project_root: Path) -> dict:
    """Run all architecture scans and return a combined report."""
    return {
        "large_files": scan_large_files(project_root),
        "missing_tests": scan_missing_tests(project_root),
        "circular_imports": scan_circular_imports(project_root),
        "long_functions": scan_long_functions(project_root),
    }


if __name__ == "__main__":
    from .kb_sync import find_project_root

    root = find_project_root()
    if root:
        report = full_scan(root)
        print(f"Large files: {len(report['large_files'])}")
        print(f"Missing tests: {len(report['missing_tests'])}")
        print(f"Circular imports: {len(report['circular_imports'])}")
        print(f"Long functions: {len(report['long_functions'])}")
    else:
        print("No .agentflow/ directory found.")
