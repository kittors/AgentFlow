"""AgentFlow knowledge graph builder — construct and update graph from code.

Generates ``nodes.json`` and ``edges.json`` for the knowledge graph.
"""

from __future__ import annotations

import json
import hashlib
from pathlib import Path
from datetime import datetime, timezone


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
                nodes.append({
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
                })
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
                nodes.append({
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
                })
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
            edges.append({
                "source": module_map[parent_rel],
                "target": fnode["id"],
                "type": "contains",
                "weight": 1.0,
                "metadata": {"description": "directory contains file", "created_at": now},
            })
    return edges


def build_graph(project_root: Path, source_dirs: list[str] | None = None) -> dict:
    """Build a complete knowledge graph and write to disk.

    Returns summary with ``node_count`` and ``edge_count``.
    """
    graph_dir = project_root / ".agentflow" / "kb" / "graph"
    graph_dir.mkdir(parents=True, exist_ok=True)

    file_nodes = build_file_nodes(project_root, source_dirs)
    module_nodes = build_module_nodes(project_root, source_dirs)
    all_nodes = module_nodes + file_nodes

    edges = build_contains_edges(project_root, file_nodes, module_nodes)

    nodes_data = {"version": 1, "nodes": all_nodes}
    edges_data = {"version": 1, "edges": edges}

    (graph_dir / "nodes.json").write_text(
        json.dumps(nodes_data, indent=2, ensure_ascii=False), encoding="utf-8"
    )
    (graph_dir / "edges.json").write_text(
        json.dumps(edges_data, indent=2, ensure_ascii=False), encoding="utf-8"
    )

    return {"node_count": len(all_nodes), "edge_count": len(edges)}


if __name__ == "__main__":
    from .kb_sync import find_project_root

    root = find_project_root()
    if root:
        result = build_graph(root)
        print(f"Graph built: {result['node_count']} nodes, {result['edge_count']} edges")
    else:
        print("No .agentflow/ directory found. Run ~init first.")
