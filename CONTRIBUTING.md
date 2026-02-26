# Contributing to AgentFlow

Thank you for your interest in contributing! This guide will help you get started.

## Development Setup

```bash
# Clone the repository
git clone https://github.com/kittors/AgentFlow.git
cd AgentFlow

# Create a virtual environment
python -m venv .venv
source .venv/bin/activate  # macOS/Linux
# .venv\Scripts\Activate.ps1  # Windows

# Install in editable mode with dev dependencies
pip install -e ".[dev]"
```

## Code Standards

- **Python 3.10+** — use type hints everywhere
- **Linter**: [Ruff](https://github.com/astral-sh/ruff) (`ruff check agentflow/`)
- **Line length**: 120 characters
- **Encoding**: UTF-8 without BOM
- **Locale**: All user-facing strings must use `msg(zh, en)` for bilingual output
- **File operations**: Use `_safe_write()` / `_safe_remove()` helpers for Windows compatibility

## Running Tests

```bash
# Run all tests
pytest tests/ -v

# Run specific test file
pytest tests/test_installer.py -v
```

## Project Structure

```
AgentFlow/
├── AGENTS.md              ← Core prompt system (G1–G12)
├── SKILL.md               ← Skill discovery metadata
├── agentflow/
│   ├── cli.py             ← CLI entry point
│   ├── _constants.py      ← Shared constants and helpers
│   ├── installer.py       ← Deploy to CLI targets
│   ├── interactive.py     ← Interactive menus
│   ├── updater.py         ← Update/status/clean
│   ├── version_check.py   ← GitHub version check
│   ├── stages/            ← DESIGN + DEVELOP workflows
│   ├── services/          ← Knowledge, Memory, Package, Attention, Support
│   ├── rules/             ← State, Cache, Tools, Scaling
│   ├── rlm/roles/         ← 6 specialized agent roles
│   ├── functions/         ← 14 workflow commands
│   ├── templates/         ← KB/plan templates
│   └── hooks/             ← Claude Code + Codex CLI hooks
├── install.sh             ← One-line installer (macOS/Linux)
├── install.ps1            ← One-line installer (Windows)
├── bin/agentflow.js       ← npx bridge
└── tests/
```

## Making Changes

1. **Fork** the repository
2. **Create a branch**: `git checkout -b feature/my-feature`
3. **Make your changes** and add tests if applicable
4. **Run lint**: `ruff check agentflow/`
5. **Run tests**: `pytest tests/ -v`
6. **Commit**: use clear, descriptive commit messages
7. **Push** and open a **Pull Request**

## Adding a New CLI Target

1. Add the target config to `CLI_TARGETS` in `_constants.py`
2. Add hook config in `hooks/` if applicable
3. Add tests in `test_installer.py`
4. Update `README.md` and `README_CN.md`

## Adding a New Workflow Command

1. Create `agentflow/functions/your_command.md`
2. Add the command to the routing table in `AGENTS.md` (G7 section)
3. Add to `SKILL.md` commands list
4. Update `README.md` and `README_CN.md`

## Reporting Issues

Please include:
- OS and Python version
- CLI target(s) in use
- Steps to reproduce
- Expected vs actual behavior

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
