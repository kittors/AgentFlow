"""AgentFlow knowledge graph builder — construct and update graph from code.

Generates ``nodes.json`` and ``edges.json`` for the knowledge graph.
Supports ``contains``, ``imports``, and ``calls`` edge types, plus Mermaid export.
"""

from __future__ import annotations

import ast
import hashlib
import json
from datetime import datetime, timezone
from pathlib import Path


def _node_id(name: str, node_type: str) -> str:
    """Generate a deterministic node ID from name and type."""
    return hashlib.sha256(f"{node_type}:{name}".encode()).hexdigest()[:12]


def build_file_nodes(project_root: Path, source_dirs: list[str] | None = None) -> list[dict]:
    """Scan source files and create graph nodes for each."""
    dirs = source_dirs or ["src", "lib", "app"]
    nodes: list[dict] = []
    now = datetime.now(timezone.utc).isoformat()

    for d in dirs:
        src = project_root / d
        if not src.is_dir():
            continue
        for f in sorted(src.rglob("*")):
            if f.is_file() and not f.name.startswith("."):
                rel = str(f.relative_to(project_root))
                nodes.append(
                    {
                        "id": _node_id(rel, "file"),
                        "type": "file",
                        "name": f.name,
                        "path": rel,
                        "metadata": {
                            "description": "",
                            "created_at": now,
                            "updated_at": now,
                            "tags": [f.suffix.lstrip(".")],
                        },
                    }
                )
    return nodes


def build_module_nodes(project_root: Path, source_dirs: list[str] | None = None) -> list[dict]:
    """Create graph nodes for directories (modules)."""
    dirs = source_dirs or ["src", "lib", "app"]
    nodes: list[dict] = []
    now = datetime.now(timezone.utc).isoformat()

    for d in dirs:
        src = project_root / d
        if not src.is_dir():
            continue
        for child in sorted(src.rglob("*")):
            if child.is_dir() and not child.name.startswith((".", "_")):
                rel = str(child.relative_to(project_root))
                nodes.append(
                    {
                        "id": _node_id(rel, "module"),
                        "type": "module",
                        "name": child.name,
                        "path": rel,
                        "metadata": {
                            "description": "",
                            "created_at": now,
                            "updated_at": now,
                            "tags": [],
                        },
                    }
                )
    return nodes


def build_contains_edges(project_root: Path, file_nodes: list[dict], module_nodes: list[dict]) -> list[dict]:
    """Create 'contains' edges from modules to files."""
    edges: list[dict] = []
    now = datetime.now(timezone.utc).isoformat()

    module_map = {n["path"]: n["id"] for n in module_nodes}

    for fnode in file_nodes:
        fpath = Path(fnode["path"])
        parent_rel = str(fpath.parent)
        if parent_rel in module_map:
            edges.append(
                {
                    "source": module_map[parent_rel],
                    "target": fnode["id"],
                    "type": "contains",
                    "weight": 1.0,
                    "metadata": {"description": "directory contains file", "created_at": now},
                }
            )
    return edges


def build_import_edges(
    project_root: Path,
    file_nodes: list[dict],
    source_dirs: list[str] | None = None,
) -> list[dict]:
    """Create 'imports' edges by parsing Python import statements.

    Uses AST to accurately extract ``import X`` and ``from X import Y`` and
    resolves them to file nodes when possible.
    """
    dirs = source_dirs or ["src", "lib", "app"]
    edges: list[dict] = []
    now = datetime.now(timezone.utc).isoformat()

    # Build lookup: module stem -> node id  (e.g. "utils" -> id)
    stem_to_node: dict[str, str] = {}
    for node in file_nodes:
        if node["name"].endswith(".py"):
            stem_to_node[Path(node["path"]).stem] = node["id"]

    for d in dirs:
        src = project_root / d
        if not src.is_dir():
            continue
        for f in src.rglob("*.py"):
            rel = str(f.relative_to(project_root))
            source_id = _node_id(rel, "file")
            try:
                tree = ast.parse(f.read_text(encoding="utf-8", errors="ignore"))
            except SyntaxError:
                continue
            imported_modules: set[str] = set()
            for ast_node in ast.walk(tree):
                if isinstance(ast_node, ast.Import):
                    for alias in ast_node.names:
                        imported_modules.add(alias.name.split(".")[0])
                elif isinstance(ast_node, ast.ImportFrom):
                    if ast_node.module:
                        imported_modules.add(ast_node.module.split(".")[0])
            for mod_name in imported_modules:
                if mod_name in stem_to_node and stem_to_node[mod_name] != source_id:
                    edges.append(
                        {
                            "source": source_id,
                            "target": stem_to_node[mod_name],
                            "type": "imports",
                            "weight": 0.8,
                            "metadata": {"description": f"imports {mod_name}", "created_at": now},
                        }
                    )
    return edges


def build_call_edges(
    project_root: Path,
    file_nodes: list[dict],
    source_dirs: list[str] | None = None,
) -> list[dict]:
    """Create 'calls' edges by detecting function calls across files.

    Uses a two-pass approach:
    1. Collect all top-level function definitions per file.
    2. Scan for calls to those functions in other files.
    """
    dirs = source_dirs or ["src", "lib", "app"]
    edges: list[dict] = []
    now = datetime.now(timezone.utc).isoformat()

    # Pass 1: collect function definitions per file
    func_to_file_id: dict[str, str] = {}  # func_name -> defining file node id
    py_files: list[tuple[Path, str]] = []  # (path, rel)

    for d in dirs:
        src = project_root / d
        if not src.is_dir():
            continue
        for f in src.rglob("*.py"):
            rel = str(f.relative_to(project_root))
            py_files.append((f, rel))
            try:
                tree = ast.parse(f.read_text(encoding="utf-8", errors="ignore"))
            except SyntaxError:
                continue
            file_id = _node_id(rel, "file")
            for ast_node in ast.iter_child_nodes(tree):
                if isinstance(ast_node, ast.FunctionDef | ast.AsyncFunctionDef):
                    if not ast_node.name.startswith("_"):
                        func_to_file_id[ast_node.name] = file_id

    # Pass 2: find calls to those functions in other files
    seen_edges: set[tuple[str, str]] = set()
    for f, rel in py_files:
        caller_id = _node_id(rel, "file")
        try:
            tree = ast.parse(f.read_text(encoding="utf-8", errors="ignore"))
        except SyntaxError:
            continue
        for ast_node in ast.walk(tree):
            if isinstance(ast_node, ast.Call):
                func_name = None
                if isinstance(ast_node.func, ast.Name):
                    func_name = ast_node.func.id
                elif isinstance(ast_node.func, ast.Attribute):
                    func_name = ast_node.func.attr
                if func_name and func_name in func_to_file_id and func_to_file_id[func_name] != caller_id:
                    edge_key = (caller_id, func_to_file_id[func_name])
                    if edge_key not in seen_edges:
                        seen_edges.add(edge_key)
                        edges.append(
                            {
                                "source": caller_id,
                                "target": func_to_file_id[func_name],
                                "type": "calls",
                                "weight": 0.6,
                                "metadata": {
                                    "description": f"calls {func_name}",
                                    "created_at": now,
                                },
                            }
                        )
    return edges


def export_mermaid(nodes: list[dict], edges: list[dict]) -> str:
    """Export a knowledge graph as a Mermaid flowchart string.

    Returns a ready-to-render Mermaid diagram definition.
    """
    lines = ["graph LR"]

    # Sanitize label for Mermaid (quote special chars)
    def _label(text: str) -> str:
        return text.replace('"', "'").replace("[", "(").replace("]", ")")

    # Emit nodes with shapes based on type
    for n in nodes:
        nid = n["id"]
        label = _label(n["name"])
        if n["type"] == "module":
            lines.append(f'    {nid}[["{label}"]]")  # stadium shape for modules')
        else:
            lines.append(f'    {nid}["{label}"]')

    # Emit edges with labels based on type
    style_map = {
        "contains": "-->",
        "imports": "-.->",
        "calls": "==>",
    }
    for e in edges:
        arrow = style_map.get(e["type"], "-->")
        label = e["type"]
        lines.append(f"    {e['source']} {arrow}|{label}| {e['target']}")

    return "\n".join(lines) + "\n"


def build_graph(project_root: Path, source_dirs: list[str] | None = None) -> dict:
    """Build a complete knowledge graph and write to disk.

    Returns summary with ``node_count``, ``edge_count``, and edge type breakdown.
    """
    graph_dir = project_root / ".agentflow" / "kb" / "graph"
    graph_dir.mkdir(parents=True, exist_ok=True)

    file_nodes = build_file_nodes(project_root, source_dirs)
    module_nodes = build_module_nodes(project_root, source_dirs)
    all_nodes = module_nodes + file_nodes

    # Build all edge types
    contains_edges = build_contains_edges(project_root, file_nodes, module_nodes)
    import_edges = build_import_edges(project_root, file_nodes, source_dirs)
    call_edges = build_call_edges(project_root, file_nodes, source_dirs)
    all_edges = contains_edges + import_edges + call_edges

    nodes_data = {"version": 2, "nodes": all_nodes}
    edges_data = {"version": 2, "edges": all_edges}

    (graph_dir / "nodes.json").write_text(json.dumps(nodes_data, indent=2, ensure_ascii=False), encoding="utf-8")
    (graph_dir / "edges.json").write_text(json.dumps(edges_data, indent=2, ensure_ascii=False), encoding="utf-8")

    # Export Mermaid visualization
    mermaid = export_mermaid(all_nodes, all_edges)
    (graph_dir / "graph.mmd").write_text(mermaid, encoding="utf-8")

    return {
        "node_count": len(all_nodes),
        "edge_count": len(all_edges),
        "edge_types": {
            "contains": len(contains_edges),
            "imports": len(import_edges),
            "calls": len(call_edges),
        },
    }


if __name__ == "__main__":
    from .kb_sync import find_project_root

    root = find_project_root()
    if root:
        result = build_graph(root)
        print(f"Graph built: {result['node_count']} nodes, {result['edge_count']} edges")
        print(f"  contains: {result['edge_types']['contains']}")
        print(f"  imports:  {result['edge_types']['imports']}")
        print(f"  calls:    {result['edge_types']['calls']}")
    else:
        print("No .agentflow/ directory found. Run ~init first.")
