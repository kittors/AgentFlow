<div align="center">

<!-- ASCII Banner -->
```
     █████╗  ██████╗ ███████╗███╗   ██╗████████╗███████╗██╗      ██████╗ ██╗    ██╗
    ██╔══██╗██╔════╝ ██╔════╝████╗  ██║╚══██╔══╝██╔════╝██║     ██╔═══██╗██║    ██║
    ███████║██║  ███╗█████╗  ██╔██╗ ██║   ██║   █████╗  ██║     ██║   ██║██║ █╗ ██║
    ██╔══██║██║   ██║██╔══╝  ██║╚██╗██║   ██║   ██╔══╝  ██║     ██║   ██║██║███╗██║
    ██║  ██║╚██████╔╝███████╗██║ ╚████║   ██║   ██║     ███████╗╚██████╔╝╚███╔███╔╝
    ╚═╝  ╚═╝ ╚═════╝ ╚══════╝╚═╝  ╚═══╝   ╚═╝   ╚═╝     ╚══════╝ ╚═════╝  ╚══╝╚══╝
```

**An autonomous advanced AI partner that doesn't just analyze — it keeps working until implementation and verification are complete.**

[English](README.md) · [中文](README_CN.md)

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Python 3.10+](https://img.shields.io/badge/Python-3.10+-3776AB.svg?logo=python&logoColor=white)](https://www.python.org)
[![CI](https://github.com/kittors/AgentFlow/actions/workflows/ci.yml/badge.svg)](https://github.com/kittors/AgentFlow/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/kittors/AgentFlow?include_prereleases&label=Release)](https://github.com/kittors/AgentFlow/releases)

**`5-Level Routing`** · **`EHRB Safety`** · **`Knowledge Graph Memory`** · **`Sub-Agent Orchestration`** · **`Convention Extraction`**

</div>

---

## Why AgentFlow?

Most AI assistants can analyze tasks but stop before real delivery. AgentFlow adds **strict routing**, **staged execution**, **safety gates**, and **verification** — ensuring every task reaches completion with quality.

<table>
<tr><td>

**🎯 What makes it different?**

- 📊 **5-level routing** — proportional effort, from quick fixes to architecture redesign
- 🏗️ **R4 architecture mode** — dedicated workflow for system-level refactoring
- 🧠 **Knowledge graph** — graph-based project memory that persists across sessions
- 🔍 **Convention extraction** — auto-discover and enforce coding patterns
- 📡 **Architecture scanning** — proactive detection of code smells and structural issues
- 📈 **Project dashboard** — HTML status visualization

</td></tr>
</table>

## Quick Start

### Method A: One-line install script _(recommended)_

**macOS / Linux:**

```bash
curl -fsSL https://raw.githubusercontent.com/kittors/AgentFlow/main/install.sh | bash
```

**Windows PowerShell:**

```powershell
irm https://raw.githubusercontent.com/kittors/AgentFlow/main/install.ps1 | iex
```

### Method B: npx (Node.js ≥ 16)

```bash
npx agentflow
```

### Method C: UV _(isolated environment)_

```bash
uv tool install --from git+https://github.com/kittors/AgentFlow agentflow && agentflow
```

### Method D: pip (Python ≥ 3.10)

```bash
pip install git+https://github.com/kittors/AgentFlow.git && agentflow
```

### Install to specific CLI targets

```bash
agentflow                     # Interactive menu
agentflow install codex       # Specify target directly
agentflow install --all       # Install to all detected CLIs
```

### Update

```bash
agentflow update
```

### Verify & Uninstall

```bash
agentflow status              # Check installation status
agentflow version             # Show version
agentflow uninstall codex     # Uninstall from a target
agentflow uninstall --all     # Uninstall from all targets
agentflow clean               # Clean caches
```

---

## Features

### 🎯 5-Level Routing (R0–R4)

Every input is scored on **5 dimensions** and routed to the right process level:

| Level | Score | Use Case | Process |
|:-----:|:-----:|----------|---------|
| R0 💬 | ≤ 3 | Chat, Q&A, explanations | Direct reply |
| R1 ⚡ | 4-6 | Quick fix, single-file edit | Locate → Fix → Verify |
| R2 📝 | 7-9 | Multi-file changes, new features | Confirm → Design → Develop |
| R3 📊 | 10-12 | Complex features, refactoring | Confirm → Multi-proposal Design → Develop |
| R4 🏗️ | ≥ 13 | Architecture redesign, migrations | Evaluate → Design + Review → Phased Develop |

**5 scoring dimensions:** Action Need (0-3) · Target Clarity (0-3) · Decision Scope (0-3) · Impact Range (0-3) · EHRB Risk (0-3)

### 🛡️ EHRB Three-Layer Safety

Catches destructive operations **before** execution:

| Layer | What it does |
|-------|-------------|
| **Keyword scan** | Detects `rm -rf`, `DROP TABLE`, `git push -f`, secrets, PII, payment ops |
| **Semantic analysis** | Identifies permission bypass, environment mismatch, logic vulnerabilities |
| **Tool output inspection** | Catches injection, format hijacking, sensitive data leakage |

### 🤖 RLM Sub-Agent Orchestration

**6 specialized roles** + native CLI sub-agents, dispatched based on task complexity:

| Role | Purpose | When |
|------|---------|------|
| `reviewer` | Code quality + security review | Complex tasks with core modules |
| `synthesizer` | Multi-proposal analysis | Complex + multiple dimensions |
| `kb_keeper` | Knowledge base sync | KB enabled |
| `pkg_keeper` | Proposal package management | Design/Develop stages |
| `writer` | Document generation | Manual `~rlm spawn writer` |
| `architect` | System-level architecture review | R4 / architect complexity |

**Native sub-agent mapping per CLI:**

| Action | Codex CLI | Claude Code | OpenCode | Gemini |
|--------|-----------|-------------|----------|--------|
| Explore | `spawn_agent(explorer)` | `Task(Explore)` | `@explore` | `codebase_investigator` |
| Implement | `spawn_agent(worker)` | `Task(general-purpose)` | `@general` | `generalist_agent` |
| Test | `spawn_agent(awaiter)` | `Task(general-purpose)` | `@general` | — |
| Plan | Plan mode | `Task(Plan)` | — | — |

### 🧠 Three-Layer Memory

| Layer | Scope | Content |
|:-----:|-------|---------|
| **L0** | Global (cross-project) | User preferences, tech stack, communication style |
| **L1** | Project | Knowledge base, module docs, architecture decisions |
| **L2** | Session | Task progress, decisions, context summaries |

### ⚡ Unique Features

Features that set AgentFlow apart:

| Feature | Command | Description |
|---------|:-------:|-------------|
| **Knowledge Graph** | `~graph` | Graph-based project memory with nodes.json/edges.json, progressive disclosure queries |
| **Convention Extraction** | `~conventions` | Auto-discover coding patterns, enforce consistency in DEVELOP stage |
| **Architecture Scan** | `~scan` | Detect large files, circular deps, missing tests, code smells |
| **Dashboard** | `~dashboard` | Generate HTML project status dashboard |
| **R4 Architecture Routing** | _(auto)_ | Dedicated 5-stage workflow for system-level changes |
| **Architect Role** | _(auto)_ | Specialized sub-agent for architecture review |
| **Context Window Management** | _(auto)_ | Proactive summarization when context exceeds 80% |
| **Convention Check Gate** | _(auto)_ | Automated code compliance verification in DEVELOP stage |

---

## Workflow Commands

All commands run **inside AI chat**, not your system shell.

| Command | Description |
|:-------:|-------------|
| `~init` | Initialize project knowledge base |
| `~auto` | Auto-execute with full workflow |
| `~plan` | Plan only — stop before development |
| `~exec` | Execute an existing plan |
| `~status` | Show workflow status and progress |
| `~review` | Trigger code review |
| `~scan` | Architecture scan — detect structural issues |
| `~conventions` | Extract and check coding conventions |
| `~graph` | Knowledge graph operations |
| `~dashboard` | Generate project dashboard |
| `~memory` | Manage memory layers (L0/L1/L2) |
| `~rlm` | Sub-agent management and dispatch |
| `~validatekb` | Validate knowledge base consistency |

---

## Supported CLI Targets

| Target | Config Directory | Hooks |
|--------|:----------------:|:-----:|
| **Claude Code** | `~/.claude/` | ✅ Full (9 events) |
| **Codex CLI** | `~/.codex/` | ✅ Notify |
| **Gemini CLI** | `~/.gemini/` | — |
| **OpenCode** | `~/.config/opencode/` | — |
| **Qwen CLI** | `~/.qwen/` | — |
| **Grok CLI** | `~/.grok/` | — |

### Codex CLI Compatibility Notes

> The following `config.toml` settings may affect AgentFlow behavior:

| Setting | Impact | Recommendation |
|---------|--------|----------------|
| `steer = true` | May interfere with workflow interaction | Disable if issues occur |
| `child_agents_md = true` | May conflict with AgentFlow instructions | Disable |
| `project_doc_max_bytes` | AGENTS.md truncated if too low | Auto-set to 98304 during install |
| `agent_max_depth = 1` | Limits sub-agent nesting | Keep ≥ 2 |
| `agent_max_threads` | Limits parallel sub-agents | Keep default (6) or higher |

---

## How It Works

```
┌─────────────────────────────────────────────────────────────────┐
│  User Input                                                     │
└─────┬───────────────────────────────────────────────────────────┘
      │
      ▼
┌─────────────────┐     ┌──────────────────────────────────────┐
│  G4 Router      │────▶│  R0: Direct Reply                    │
│  5-dimension    │     │  R1: Fast Flow (locate→fix→verify)   │
│  scoring        │     │  R2: Simplified (confirm→design→dev) │
│  (0-15)         │     │  R3: Standard (multi-proposal)       │
│                 │     │  R4: Architecture (evaluate→staged)  │
└─────────────────┘     └──────────┬───────────────────────────┘
                                   │
                        ┌──────────▼──────────┐
                        │  EHRB Safety Gate    │
                        │  (3-layer scan)      │
                        └──────────┬──────────┘
                                   │
              ┌────────────────────┼────────────────────┐
              ▼                    ▼                    ▼
     ┌────────────┐      ┌────────────┐       ┌────────────┐
     │  DESIGN    │      │  DEVELOP   │       │  KB Sync   │
     │  stage     │─────▶│  stage     │──────▶│  + Verify  │
     │            │      │            │       │            │
     └────────────┘      └────────────┘       └────────────┘
              │                    │
              ▼                    ▼
     ┌────────────────────────────────────┐
     │  RLM Sub-Agents                    │
     │  reviewer│synthesizer│architect│...│
     │  + Native CLI sub-agents           │
     └────────────────────────────────────┘
              │
              ▼
     ┌────────────────────────────────────┐
     │  Three-Layer Memory                │
     │  L0 (user) │ L1 (project) │ L2 (session) │
     └────────────────────────────────────┘
```

---

## Architecture

```
AgentFlow/
├── AGENTS.md              ← Core prompt system (G1–G12)
├── SKILL.md               ← Skill discovery metadata
├── install.sh             ← One-line installer (macOS/Linux)
├── install.ps1            ← One-line installer (Windows)
├── package.json           ← npx bridge
├── agentflow/
│   ├── cli.py             ← CLI entry point (6 commands)
│   ├── _constants.py      ← Shared constants & helpers
│   ├── installer.py       ← Deploy to CLI targets
│   ├── interactive.py     ← Interactive menus
│   ├── updater.py         ← Update / status / clean
│   ├── version_check.py   ← GitHub version check + cache
│   ├── stages/            ← DESIGN + DEVELOP stage modules
│   ├── services/          ← Knowledge, Memory, Package, Attention, Support
│   ├── rules/             ← State, Cache, Tools, Scaling
│   ├── rlm/roles/         ← 6 specialized agent roles
│   ├── functions/         ← 14 workflow commands (~init, ~graph, etc.)
│   ├── templates/         ← KB/plan templates
│   └── hooks/             ← Claude Code (JSON) + Codex CLI (TOML)
└── tests/                 ← pytest test suite
```

---

## Data Points

| Metric | Count |
|--------|:-----:|
| CLI targets | 6 |
| Routing levels | 5 (R0–R4) |
| RLM roles | 6 |
| Workflow commands | 14 |
| Stage modules | 2 |
| Service modules | 5 |
| Rule modules | 4 |
| Hook configs | 2 |
| KB/plan templates | 1 |

---

## FAQ

<details>
<summary><b>Is this a Python CLI tool or a prompt package?</b></summary>
Both. The CLI manages installation and updates; the workflow behavior comes from <code>AGENTS.md</code> and the module files deployed to your AI coding assistant's config directory.
</details>

<details>
<summary><b>Which target should I install?</b></summary>
Install for the CLI you use: <code>codex</code>, <code>claude</code>, <code>gemini</code>, <code>qwen</code>, <code>grok</code>, or <code>opencode</code>. Use <code>--all</code> for all detected CLIs.
</details>

<details>
<summary><b>What if a rules file already exists?</b></summary>
Non-AgentFlow files are automatically backed up with a timestamp before replacement. You'll see the backup name in the console output.
</details>

<details>
<summary><b>What is RLM?</b></summary>
Role Language Model — a sub-agent orchestration system with 6 specialized roles plus native CLI sub-agents, dispatched based on task complexity.
</details>

<details>
<summary><b>Where does project knowledge go?</b></summary>
In the project-local knowledge base directory, auto-synced when code changes. Structure includes <code>modules/</code>, <code>graph/</code>, <code>conventions/</code>, <code>plan/</code>, and <code>sessions/</code>.
</details>

<details>
<summary><b>Does memory persist across sessions?</b></summary>
Yes. L0 user memory is global, L1 project KB is per-project, L2 session summaries are auto-saved at stage transitions.
</details>

<details>
<summary><b>What are Hooks?</b></summary>
Lifecycle hooks auto-deployed during installation. Claude Code gets full event hooks (safety checks, progress snapshots, KB sync, etc.); Codex CLI gets a notify hook for update checks. All optional — features degrade gracefully without hooks.
</details>

<details>
<summary><b>What is R4 Architecture Mode?</b></summary>
A dedicated routing level for system-level refactoring, tech stack migrations, and full architecture redesign. Includes an extra EVALUATE stage, multi-proposal design with architecture review, and phased development. Not available in similar tools.
</details>

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, code standards, and PR workflow.

## License

[MIT](LICENSE) — use it freely.

---

<div align="center">

**AgentFlow** — More than analysis. Keeps working until implementation and verification are complete.

</div>
