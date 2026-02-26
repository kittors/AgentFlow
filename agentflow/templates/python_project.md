# Python Project KB Template

## INDEX.md (Python)

```markdown
# {project_name}

## Overview
{description}

## Tech Stack
- Python: {3.10+|3.11+|3.12+}
- Build: {hatchling|setuptools|poetry|flit}
- Config: {pyproject.toml|setup.cfg}
- Linter: {ruff|flake8|pylint}
- Formatter: {ruff format|black}
- Type Checker: {pyright|mypy|pyre}
- Test: {pytest|unittest}

## Architecture
- Type: {CLI|Library|Web API|Data Pipeline|ML}
- Pattern: {Package|Script|Plugin|Framework}

## Entry Points
- CLI: `{project_name}.cli:main`
- Package: `{project_name}/__init__.py`
- Script: `{scripts/run.py}`

## Key Modules
- `{project_name}/core.py` — Core logic
- `{project_name}/utils.py` — Utility functions
- `{project_name}/models.py` — Data models
- `tests/` — Test suite
```

## context.md (Python)

```markdown
# Project Context

## Dependencies
### Runtime
{dependency_list}

### Development
- pytest >= {8.0}
- ruff >= {0.4}

## Build & Run
- Install: `pip install -e .`
- Install (dev): `pip install -e ".[dev]"`
- Run: `python -m {project_name}`
- Test: `pytest tests/ -v`
- Lint: `ruff check .`
- Format: `ruff format .`
- Type check: `pyright`

## Virtual Environment
- Tool: {uv|venv|conda|poetry}
- Create: `{uv venv .venv}`
- Activate: `{source .venv/bin/activate}`

## Configuration
- `pyproject.toml` — Package metadata + tool config
- `.python-version` — Python version pin
- `pyrightconfig.json` — Type checker config

## Publishing
- Registry: {PyPI|private}
- Build: `python -m build`
- Publish: `twine upload dist/*`
```
