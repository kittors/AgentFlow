# AgentFlow Development Guide

## Project Overview
AgentFlow is an AI-assisted development enhancement layer. It builds a knowledge graph of your project and progressively discloses relevant context to AI tools (Claude Code, Codex, Cursor).

## Architecture
- `agentflow/core/` — Config, context assembler, iteration engine
- `agentflow/memory/` — Knowledge graph, persistent store, project indexer, session updater
- `agentflow/planner/` — Requirement analysis and task decomposition
- `agentflow/conventions/` — Convention extraction and checking
- `agentflow/scanner/` — Proactive issue detection
- `agentflow/integrations/` — CLAUDE.md, AGENTS.md, .cursor/rules/ generators
- `agentflow/llm/` — Unified LLM client (OpenAI + Anthropic)
- `agentflow/cli.py` — Click-based CLI entry point

## Conventions
- Python 3.10+, type hints everywhere
- Pydantic v2 for data validation (use `model_dump()` not `dict()`)
- `logging` module for all output (no `print()`)
- Atomic file writes via temp file + `os.replace()`
- JSON for all persistent data
- Tests in `tests/`, fixtures in `conftest.py`

## Testing
- Offline tests: `pytest tests/test_memory.py tests/test_engine.py tests/test_scanner.py`
- Integration tests (need LLM_API_KEY): `pytest tests/test_planner.py tests/test_conventions.py`
- Use `tmp_project` fixture for tests needing a project directory
- Use `api_key` fixture for tests needing LLM access (auto-skips if not set)

## Key Design Decisions
- Knowledge graph uses networkx DiGraph with BFS traversal for progressive disclosure
- Memory nodes have types: module, pattern, decision, convention, issue, task
- Convention extraction samples max 5 files per language, max 3000 bytes each
- All LLM responses expected as JSON; fallback parsing strips markdown code blocks
