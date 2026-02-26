# AgentFlow

> Multi-CLI agent workflow system — keeps going until tasks are implemented and verified.

<p align="center">
  <strong>🚀 5-Level Routing</strong> · <strong>🛡️ EHRB Safety</strong> · <strong>🧠 Knowledge Graph Memory</strong> · <strong>🤖 Sub-Agent Orchestration</strong>
</p>

---

## Quick Start

### Install via pip

```bash
pip install git+https://github.com/kittors/AgentFlow.git && agentflow
```

### Install via UV

```bash
uv tool install --from git+https://github.com/kittors/AgentFlow agentflow && agentflow
```

### Install to specific CLI

```bash
agentflow install codex       # Codex CLI
agentflow install claude      # Claude Code
agentflow install gemini      # Gemini CLI
agentflow install --all       # All detected CLIs
```

### Verify

```bash
agentflow status
agentflow version
```

### Uninstall

```bash
agentflow uninstall codex
agentflow uninstall --all
```

---

## Features

### 🎯 5-Level Routing (R0–R4)

Every input is scored on 5 dimensions and routed to the right process:

| Level | Trigger | Process |
|-------|---------|---------|
| R0 💬 | Score ≤ 3 (chat, Q&A) | Direct reply |
| R1 ⚡ | Score 4-6 (quick fix) | Fast: locate → fix → verify |
| R2 📝 | Score 7-9 (multi-file) | Simplified: confirm → design → develop |
| R3 📊 | Score 10-12 (complex) | Full: confirm → multi-proposal design → develop |
| R4 🏗️ | Score ≥ 13 (architecture) | Architecture: evaluate → design + review → phased develop |

### 🛡️ EHRB Safety Detection

Three-layer safety catches destructive operations before execution:
1. **Keyword scan** — `rm -rf`, `DROP TABLE`, `git push -f`, secrets, PII
2. **Semantic analysis** — permission bypass, environment mismatch, logic vulnerabilities
3. **Tool output inspection** — injection, format hijacking, data leakage

### 🤖 RLM Sub-Agent Orchestration

6 specialized roles + native CLI sub-agents, dispatched based on complexity:

| Role | Purpose | Trigger |
|------|---------|---------|
| reviewer | Code quality + security review | complex tasks with core modules |
| synthesizer | Multi-proposal analysis | complex + ≥3 evaluation dimensions |
| kb_keeper | Knowledge base sync | KB enabled |
| pkg_keeper | Proposal package management | Design/Develop stages |
| writer | Standalone document generation | Manual `~rlm spawn writer` |
| **architect** | System-level architecture review | R4 / architect complexity |

### 🧠 Three-Layer Memory

| Layer | Scope | Content |
|-------|-------|---------|
| L0 | Global (cross-project) | User preferences, tech stack, communication style |
| L1 | Project | Knowledge base, module docs, architecture decisions |
| L2 | Session | Task progress, decisions, context |

### ⚡ AgentFlow Unique Features

Features not found in similar tools:

| Feature | Command | Description |
|---------|---------|-------------|
| **Knowledge Graph** | `~graph` | Graph-based project memory with query and visualization |
| **Convention Extraction** | `~conventions` | Auto-discover coding patterns from your codebase |
| **Architecture Scan** | `~scan` | Proactive detection of large files, circular deps, missing tests |
| **Dashboard** | `~dashboard` | HTML project status dashboard |
| **R4 Architecture Routing** | (auto) | Dedicated workflow for system-level refactoring |
| **Architect Role** | (auto) | Specialized sub-agent for architecture review |

---

## Workflow Commands

| Command | Description |
|---------|-------------|
| `~init` | Initialize project knowledge base |
| `~auto` | Auto-execute with full workflow |
| `~plan` | Plan only, stop before development |
| `~exec` | Execute existing plan |
| `~status` | Show workflow status |
| `~review` | Code review |
| `~scan` | Architecture scan |
| `~conventions` | Extract/check coding conventions |
| `~graph` | Knowledge graph operations |
| `~dashboard` | Generate project dashboard |
| `~memory` | Manage memory layers |
| `~rlm` | Sub-agent management |
| `~validatekb` | Validate knowledge base consistency |

---

## Supported CLI Targets

| Target | Config Dir | Status |
|--------|-----------|--------|
| Claude Code | `~/.claude/` | ✅ |
| Codex CLI | `~/.codex/` | ✅ |
| Gemini CLI | `~/.gemini/` | ✅ |
| OpenCode | `~/.config/opencode/` | ✅ |
| Qwen CLI | `~/.qwen/` | ✅ |
| Grok CLI | `~/.grok/` | ✅ |

---

## Architecture

```
AgentFlow Package
├── AGENTS.md              ← Core prompt system (G1–G12)
├── SKILL.md               ← Skill discovery metadata
├── agentflow/
│   ├── cli.py             ← CLI entry point
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
└── tests/
```

---

## License

MIT
