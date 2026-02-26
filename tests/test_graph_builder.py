"""Tests for agentflow.scripts.graph_builder."""

import json
from pathlib import Path

from agentflow.scripts.graph_builder import (
    _node_id,
    build_call_edges,
    build_contains_edges,
    build_file_nodes,
    build_graph,
    build_import_edges,
    build_module_nodes,
    export_mermaid,
)


class TestNodeId:
    def test_deterministic(self):
        assert _node_id("foo", "file") == _node_id("foo", "file")

    def test_different_types_different_ids(self):
        assert _node_id("foo", "file") != _node_id("foo", "module")

    def test_returns_12_chars(self):
        assert len(_node_id("test", "file")) == 12


class TestBuildFileNodes:
    def test_creates_nodes_for_files(self, tmp_path: Path):
        src = tmp_path / "src"
        src.mkdir()
        (src / "app.py").write_text("pass\n")
        (src / "utils.py").write_text("pass\n")
        nodes = build_file_nodes(tmp_path, source_dirs=["src"])
        assert len(nodes) == 2
        names = {n["name"] for n in nodes}
        assert names == {"app.py", "utils.py"}

    def test_node_structure(self, tmp_path: Path):
        src = tmp_path / "src"
        src.mkdir()
        (src / "main.py").write_text("pass\n")
        nodes = build_file_nodes(tmp_path, source_dirs=["src"])
        node = nodes[0]
        assert "id" in node
        assert node["type"] == "file"
        assert node["name"] == "main.py"
        assert "metadata" in node
        assert "created_at" in node["metadata"]

    def test_ignores_hidden_files(self, tmp_path: Path):
        src = tmp_path / "src"
        src.mkdir()
        (src / ".hidden").write_text("secret\n")
        (src / "visible.py").write_text("pass\n")
        nodes = build_file_nodes(tmp_path, source_dirs=["src"])
        assert len(nodes) == 1

    def test_empty_dir(self, tmp_path: Path):
        nodes = build_file_nodes(tmp_path, source_dirs=["nonexistent"])
        assert nodes == []


class TestBuildModuleNodes:
    def test_creates_nodes_for_directories(self, tmp_path: Path):
        src = tmp_path / "src"
        pkg = src / "mypackage"
        pkg.mkdir(parents=True)
        (pkg / "__init__.py").write_text("")
        nodes = build_module_nodes(tmp_path, source_dirs=["src"])
        assert len(nodes) == 1
        assert nodes[0]["type"] == "module"
        assert nodes[0]["name"] == "mypackage"

    def test_ignores_hidden_and_underscore_dirs(self, tmp_path: Path):
        src = tmp_path / "src"
        (src / ".git" / "objects").mkdir(parents=True)
        (src / "__pycache__").mkdir(parents=True)
        (src / "real_module").mkdir(parents=True)
        nodes = build_module_nodes(tmp_path, source_dirs=["src"])
        names = {n["name"] for n in nodes}
        assert ".git" not in names
        assert "__pycache__" not in names


class TestBuildContainsEdges:
    def test_creates_edges(self, tmp_path: Path):
        src = tmp_path / "src"
        pkg = src / "pkg"
        pkg.mkdir(parents=True)
        (pkg / "a.py").write_text("pass\n")

        file_nodes = build_file_nodes(tmp_path, source_dirs=["src"])
        module_nodes = build_module_nodes(tmp_path, source_dirs=["src"])
        edges = build_contains_edges(tmp_path, file_nodes, module_nodes)
        assert len(edges) >= 1
        assert edges[0]["type"] == "contains"
        assert edges[0]["weight"] == 1.0


class TestBuildImportEdges:
    def test_detects_import_statement(self, tmp_path: Path):
        src = tmp_path / "src"
        src.mkdir()
        (src / "utils.py").write_text("def helper(): pass\n")
        (src / "main.py").write_text("import utils\n\nutils.helper()\n")
        file_nodes = build_file_nodes(tmp_path, source_dirs=["src"])
        edges = build_import_edges(tmp_path, file_nodes, source_dirs=["src"])
        assert len(edges) == 1
        assert edges[0]["type"] == "imports"
        assert edges[0]["weight"] == 0.8

    def test_detects_from_import(self, tmp_path: Path):
        src = tmp_path / "src"
        src.mkdir()
        (src / "utils.py").write_text("def helper(): pass\n")
        (src / "app.py").write_text("from utils import helper\n")
        file_nodes = build_file_nodes(tmp_path, source_dirs=["src"])
        edges = build_import_edges(tmp_path, file_nodes, source_dirs=["src"])
        assert len(edges) == 1

    def test_ignores_stdlib_imports(self, tmp_path: Path):
        src = tmp_path / "src"
        src.mkdir()
        (src / "app.py").write_text("import os\nimport sys\nimport json\n")
        file_nodes = build_file_nodes(tmp_path, source_dirs=["src"])
        edges = build_import_edges(tmp_path, file_nodes, source_dirs=["src"])
        assert edges == []

    def test_no_self_import(self, tmp_path: Path):
        src = tmp_path / "src"
        src.mkdir()
        (src / "app.py").write_text("# no imports\nx = 1\n")
        file_nodes = build_file_nodes(tmp_path, source_dirs=["src"])
        edges = build_import_edges(tmp_path, file_nodes, source_dirs=["src"])
        assert edges == []


class TestBuildCallEdges:
    def test_detects_cross_file_call(self, tmp_path: Path):
        src = tmp_path / "src"
        src.mkdir()
        (src / "utils.py").write_text("def helper():\n    return 42\n")
        (src / "main.py").write_text("from utils import helper\n\nresult = helper()\n")
        file_nodes = build_file_nodes(tmp_path, source_dirs=["src"])
        edges = build_call_edges(tmp_path, file_nodes, source_dirs=["src"])
        assert len(edges) >= 1
        assert edges[0]["type"] == "calls"
        assert edges[0]["weight"] == 0.6

    def test_ignores_private_functions(self, tmp_path: Path):
        src = tmp_path / "src"
        src.mkdir()
        (src / "utils.py").write_text("def _internal():\n    pass\n")
        (src / "main.py").write_text("_internal()\n")
        file_nodes = build_file_nodes(tmp_path, source_dirs=["src"])
        edges = build_call_edges(tmp_path, file_nodes, source_dirs=["src"])
        assert edges == []

    def test_deduplicates_edges(self, tmp_path: Path):
        src = tmp_path / "src"
        src.mkdir()
        (src / "utils.py").write_text("def helper():\n    pass\n")
        (src / "main.py").write_text("helper()\nhelper()\nhelper()\n")
        file_nodes = build_file_nodes(tmp_path, source_dirs=["src"])
        edges = build_call_edges(tmp_path, file_nodes, source_dirs=["src"])
        # Even with 3 calls, only 1 edge
        assert len(edges) <= 1


class TestExportMermaid:
    def test_produces_valid_output(self):
        nodes = [
            {"id": "n1", "type": "module", "name": "auth"},
            {"id": "n2", "type": "file", "name": "login.py"},
        ]
        edges = [
            {"source": "n1", "target": "n2", "type": "contains"},
        ]
        result = export_mermaid(nodes, edges)
        assert "graph LR" in result
        assert "n1" in result
        assert "n2" in result
        assert "contains" in result

    def test_different_arrow_styles(self):
        nodes = [
            {"id": "a", "type": "file", "name": "a.py"},
            {"id": "b", "type": "file", "name": "b.py"},
            {"id": "c", "type": "file", "name": "c.py"},
        ]
        edges = [
            {"source": "a", "target": "b", "type": "imports"},
            {"source": "b", "target": "c", "type": "calls"},
        ]
        result = export_mermaid(nodes, edges)
        assert "-.->|imports|" in result
        assert "==>|calls|" in result

    def test_empty_graph(self):
        result = export_mermaid([], [])
        assert "graph LR" in result


class TestBuildGraph:
    def test_writes_json_files(self, tmp_path: Path):
        src = tmp_path / "src"
        src.mkdir()
        (src / "main.py").write_text("pass\n")
        (tmp_path / ".agentflow").mkdir()

        result = build_graph(tmp_path, source_dirs=["src"])
        assert result["node_count"] >= 1
        assert result["edge_count"] >= 0
        assert "edge_types" in result

        graph_dir = tmp_path / ".agentflow" / "kb" / "graph"
        assert (graph_dir / "nodes.json").exists()
        assert (graph_dir / "edges.json").exists()
        assert (graph_dir / "graph.mmd").exists()

        nodes_data = json.loads((graph_dir / "nodes.json").read_text())
        assert nodes_data["version"] == 2
        assert len(nodes_data["nodes"]) >= 1

    def test_edge_types_breakdown(self, tmp_path: Path):
        src = tmp_path / "src"
        pkg = src / "pkg"
        pkg.mkdir(parents=True)
        (pkg / "utils.py").write_text("def helper():\n    pass\n")
        (pkg / "main.py").write_text("import utils\nhelper()\n")
        (tmp_path / ".agentflow").mkdir()

        result = build_graph(tmp_path, source_dirs=["src"])
        assert "contains" in result["edge_types"]
        assert "imports" in result["edge_types"]
        assert "calls" in result["edge_types"]

    def test_mermaid_file_generated(self, tmp_path: Path):
        src = tmp_path / "src"
        src.mkdir()
        (src / "app.py").write_text("x = 1\n")
        (tmp_path / ".agentflow").mkdir()
        build_graph(tmp_path, source_dirs=["src"])
        mmd = (tmp_path / ".agentflow" / "kb" / "graph" / "graph.mmd").read_text()
        assert "graph LR" in mmd
