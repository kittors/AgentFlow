<div align="center">

```
     █████╗  ██████╗ ███████╗███╗   ██╗████████╗███████╗██╗      ██████╗ ██╗    ██╗
    ██╔══██╗██╔════╝ ██╔════╝████╗  ██║╚══██╔══╝██╔════╝██║     ██╔═══██╗██║    ██║
    ███████║██║  ███╗█████╗  ██╔██╗ ██║   ██║   █████╗  ██║     ██║   ██║██║ █╗ ██║
    ██╔══██║██║   ██║██╔══╝  ██║╚██╗██║   ██║   ██╔══╝  ██║     ██║   ██║██║███╗██║
    ██║  ██║╚██████╔╝███████╗██║ ╚████║   ██║   ██║     ███████╗╚██████╔╝╚███╔███╔╝
    ╚═╝  ╚═╝ ╚═════╝ ╚══════╝╚═╝  ╚═══╝   ╚═╝   ╚═╝     ╚══════╝ ╚═════╝  ╚══╝╚══╝
```

**An autonomous advanced AI partner that keeps working until implementation and verification are complete.**

[English](README.md) · [中文](README_CN.md)

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go 1.26](https://img.shields.io/badge/Go-1.26-00ADD8.svg?logo=go&logoColor=white)](https://go.dev)
[![CI](https://github.com/kittors/AgentFlow/actions/workflows/ci.yml/badge.svg)](https://github.com/kittors/AgentFlow/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/kittors/AgentFlow?include_prereleases&label=Release)](https://github.com/kittors/AgentFlow/releases)

**`5-Level Routing`** · **`EHRB Safety`** · **`Knowledge Graph Memory`** · **`Sub-Agent Orchestration`** · **`Cross-Platform Binary`**

</div>

---

## What is AgentFlow?

AgentFlow is a **multi-CLI agent workflow system** that transforms how AI coding assistants work. Instead of treating every request the same way, AgentFlow scores each input across five dimensions and routes it to the appropriate workflow depth — from instant answers to full-scale architecture reviews. It wraps AI coding CLIs (Codex, Claude Code) with a unified rule layer that brings **intelligent routing, safety detection, persistent memory, and sub-agent orchestration**.

AgentFlow is shipped as a **single Go binary** with all workflow assets embedded — no Python, `pip`, `uv`, or PyInstaller required.

---

## ✨ Key Features

### 🎯 5-Level Intelligent Routing (R0–R4)

Every input is scored on 5 dimensions — **action need**, **goal clarity**, **decision scope**, **impact range**, and **EHRB risk** — and routed to the appropriate workflow depth:

| Level | Score | Trigger | Workflow |
|-------|-------|---------|----------|
| **R0** 💬 | ≤ 3 | Chat, Q&A, concept explanations | Direct response, no evaluation |
| **R1** ⚡ | 4–6 | Single-file fixes, formatting | Quick flow: evaluate → fix → verify |
| **R2** 📝 | 7–9 | New features, multi-file changes | Simplified design → develop → KB sync |
| **R3** 📊 | 10–12 | Complex features, cross-module refactors | Full design (with multi-scheme comparison) → develop → KB sync |
| **R4** 🏗️ | ≥ 13 | System-level refactors, stack migrations | Deep evaluation → multi-scheme architecture review → phased development |

Automatic level escalation: R1→R2 when scope exceeds expectations; R2→R3 for architectural impact; R3→R4 for system-level restructuring.

### 🛡️ EHRB Three-Layer Safety Detection

**EHRB** (Extremely High Risk Behavior) adds a triple safety gate that intercepts destructive operations **before execution**:

| Layer | What It Catches |
|-------|----------------|
| **Layer 1 — Keyword Scan** | `rm -rf`, `DROP TABLE`, `git push -f`, `chmod 777`, password/secret/token leakage, PII data, payment operations |
| **Layer 2 — Semantic Analysis** | Data safety violations, permission bypass attempts, environment targeting mistakes, logic vulnerabilities |
| **Layer 3 — Tool Output Inspection** | Command injection in external tool output, format hijacking, sensitive information leakage |

Four processing modes: **Interactive** (warn → user confirm), **Delegated** (warn → downgrade to interactive), **Turbo** (log risk, continue, report at end), and **External Tool Output** (safe → normal, suspicious → prompt, high-risk → warning).

### 🧠 Three-Layer Memory Model (L0 / L1 / L2)

Context survives across sessions and projects through a structured memory architecture:

| Layer | Scope | Location | Content |
|-------|-------|----------|---------|
| **L0 — User Memory** | Global (cross-project) | `~/.agentflow/user/profile.md` | User preferences, tech stack, communication style, global rules |
| **L1 — Project Knowledge Base** | Project-level (cross-session) | `.agentflow/kb/` | Project structure, module docs, architecture decisions, coding conventions |
| **L2 — Session Records** | Session-level | `.agentflow/sessions/` | Task progress, key decisions, issues encountered, context snapshots |

Memory is automatically loaded on session start (L0 + latest L2) and saved on session end. Stage transitions trigger automatic snapshot saves.

### 🕸️ Knowledge Graph Memory *(AgentFlow Unique)*

Beyond flat-file knowledge bases, AgentFlow builds a **graph-based project memory** with nodes and edges (`nodes.json` + `edges.json`). The knowledge graph supports:

- Progressive disclosure queries via `~graph`
- Automatic graph updates on code changes
- Node-level relationships between modules, decisions, and patterns
- Richer context than traditional file-based memory

### 🔍 Convention Extraction *(AgentFlow Unique)*

AgentFlow automatically discovers coding patterns from your codebase and enforces them during development:

- Auto-scans your project to extract conventions into `extracted.json`
- Validates new code against established patterns when `CONVENTION_CHECK=1`
- Manages conventions via `~conventions` command
- Ensures consistency across the entire codebase

### 🏗️ Architecture Scanning *(AgentFlow Unique)*

Proactive issue detection that finds problems before they become tech debt:

- Detects **large files** that should be split
- Identifies **circular dependencies** between modules
- Flags **missing tests** for critical code paths
- Generates project dashboards via `~dashboard`
- Accessible via `~scan` command

### 🤖 RLM Sub-Agent Orchestration

For complex tasks, AgentFlow dispatches specialized sub-agents based on task complexity:

| Role | Responsibility |
|------|---------------|
| **Reviewer** | Code review, quality checks, test coverage analysis |
| **Synthesizer** | Information aggregation, multi-source analysis |
| **KB Keeper** | Knowledge base maintenance, documentation sync |
| **Pkg Keeper** | Dependency management, package analysis |
| **Writer** | Documentation writing, README generation |
| **Architect** | System design, architecture evaluation |

Complexity-based dispatching: `simple` → single agent, `moderate` → 2 agents, `complex` → 3+ agents, `architect` → full team with architecture review. Also maps to native CLI sub-agents (e.g., Codex's built-in multi-agent system).

### 📏 Project Rules Generation

AgentFlow generates AI-optimized project rule files for both CLIs and IDEs, with three profile levels:

| Profile | Content |
|---------|---------|
| **lite** | Minimal rule set — core routing and safety only |
| **standard** | + Shared operational modules (common rules, module loading, acceptance criteria) |
| **full** | + Sub-agent orchestration guidance, attention control, hook integration |

**Supported rule targets:**

| Category | Targets | Rule Format |
|----------|---------|------------|
| CLI | Codex, Claude | `AGENTS.md` / `CLAUDE.md` |
| IDE | Cursor, Windsurf, Trae, VS Code Copilot, Cline, Antigravity, Kiro | `.cursorrules` / `.windsurfrules` / `.trae/rules.md` / etc. |

### 🔌 Skills & MCP Ecosystem

AgentFlow integrates with the **Skills** and **MCP (Model Context Protocol)** ecosystems:

- **Skills**: Install, list, and uninstall reusable AI skill packages (e.g., from [Vercel Skills Directory](https://vercel.com/docs/agent-resources/skills))
- **MCP Servers**: Install, search, and manage MCP servers across targets; config is automatically written to each tool's native config file (`config.toml`, `settings.json`, `mcp.json`, etc.)
- **Recommended MCPs**: `context7` (documentation lookup), `playwright` (browser automation), `filesystem` (file operations)

### 🔄 Execution Modes

Four execution modes give you full control over automation level:

| Mode | Behavior |
|------|----------|
| **Interactive** (default) | Step-by-step with user confirmation at key decision points |
| **Delegated** (`~auto`) | Auto-advance between stages; pause only on EHRB risks |
| **Delegated Plan** (`~plan`) | Auto-design, then stop — generates proposal + task list without touching code |
| **Turbo** | Fully autonomous: auto-execute all stages, auto-review, auto-test, up to 5 retries; outputs a comprehensive completion report |

### 📊 Built-in Commands

| Command | Description |
|---------|-------------|
| `~init` | Initialize project knowledge base (scan, index, graph, conventions) |
| `~auto` | Full auto-execute workflow |
| `~plan` | Plan only — design + task list, no code changes |
| `~exec` | Execute an existing plan, or `~exec <requirement>` for combo flow |
| `~status` | Show workflow status and task progress |
| `~review` | Trigger code review |
| `~scan` | Architecture scan (large files, circular deps, missing tests) |
| `~conventions` | Extract and check coding conventions |
| `~graph` | Knowledge graph operations |
| `~dashboard` | Generate project dashboard |
| `~memory` | Manage memory layers (L0/L1/L2) |
| `~rlm` | Sub-agent management |
| `~validatekb` | Validate knowledge base consistency |

---

## Install

### One-line installer

macOS / Linux:

```bash
curl -fsSL https://raw.githubusercontent.com/kittors/AgentFlow/main/install.sh | bash
```

Windows PowerShell:

```powershell
irm https://raw.githubusercontent.com/kittors/AgentFlow/main/install.ps1 | iex
```

> Windows recommendation: download `agentflow-windows-amd64.exe` from GitHub Releases and put it on your `PATH`.

Windows troubleshooting: see [docs/troubleshooting/windows.md](docs/troubleshooting/windows.md).

The installer downloads the latest published release binary. Pushes to `main` now refresh a continuous GitHub Release automatically, so `curl | bash`, `npx agentflow`, and `agentflow update` all follow the latest `main` build. If you still have an older `uv`/Python install earlier on `PATH`, reopen the terminal or run `export PATH="$HOME/.agentflow/bin:$PATH" && hash -r`, then verify with `which agentflow`.

### `npx` bootstrap

`npx agentflow` downloads the matching release binary into the local cache and runs it.

```bash
npx agentflow
```

### Manual binary install

Download the matching asset from [Releases](https://github.com/kittors/AgentFlow/releases):

- `agentflow-linux-amd64`
- `agentflow-linux-arm64`
- `agentflow-darwin-amd64`
- `agentflow-darwin-arm64`
- `agentflow-windows-amd64.exe`

Then place it somewhere on `PATH`, for example:

```bash
chmod +x agentflow-darwin-arm64
mv agentflow-darwin-arm64 ~/.local/bin/agentflow
agentflow version
```

### Local build

```bash
git clone https://github.com/kittors/AgentFlow.git
cd AgentFlow
go build -o ./bin/agentflow ./cmd/agentflow
./bin/agentflow version
```

### Offline install (air-gapped / restricted network)

If you are on a machine without Internet access, you can transfer the binary from another machine that does.

#### Step 1 — Download on a connected machine

Go to [Releases](https://github.com/kittors/AgentFlow/releases) and download the binary that matches the **target** machine's OS and architecture:

| OS | Architecture | File |
|----|-------------|------|
| macOS | Apple Silicon (M1/M2/M3/M4) | `agentflow-darwin-arm64` |
| macOS | Intel | `agentflow-darwin-amd64` |
| Linux | x86_64 | `agentflow-linux-amd64` |
| Linux | ARM64 / aarch64 | `agentflow-linux-arm64` |
| Windows | x86_64 | `agentflow-windows-amd64.exe` |

#### Step 2 — Transfer

Copy the downloaded file to the target machine via USB drive, `scp`, shared drive, or any other method available in your environment.

#### Step 3 — Install

**macOS**

```bash
# Remove quarantine attribute (if transferred from another Mac or downloaded via browser)
xattr -d com.apple.quarantine agentflow-darwin-arm64 2>/dev/null

chmod +x agentflow-darwin-arm64
sudo mv agentflow-darwin-arm64 /usr/local/bin/agentflow
agentflow version
```

**Linux**

```bash
chmod +x agentflow-linux-amd64
sudo mv agentflow-linux-amd64 /usr/local/bin/agentflow
agentflow version
```

> If you don't have `sudo`, move it to a user-writable directory already on `PATH`, for example `~/.local/bin/`:
>
> ```bash
> mkdir -p ~/.local/bin
> mv agentflow-linux-amd64 ~/.local/bin/agentflow
> # Make sure ~/.local/bin is in your PATH
> export PATH="$HOME/.local/bin:$PATH"
> ```

**Windows**

```powershell
# 1. Create a directory for AgentFlow (if it does not exist)
New-Item -ItemType Directory -Force -Path "$env:LOCALAPPDATA\AgentFlow"

# 2. Move the binary
Move-Item agentflow-windows-amd64.exe "$env:LOCALAPPDATA\AgentFlow\agentflow.exe"

# 3. Add to PATH (current user, persistent)
$currentPath = [Environment]::GetEnvironmentVariable('Path', 'User')
if ($currentPath -notlike "*AgentFlow*") {
    [Environment]::SetEnvironmentVariable('Path', "$currentPath;$env:LOCALAPPDATA\AgentFlow", 'User')
}

# 4. Reload PATH in current session
$env:Path = [Environment]::GetEnvironmentVariable('Path', 'User') + ";" + [Environment]::GetEnvironmentVariable('Path', 'Machine')

# 5. Verify
agentflow version
```

> After adding to `PATH` you may need to open a **new terminal window** for the change to take effect.

---

## Quick Start

```bash
agentflow                       # Interactive TUI
agentflow install codex         # Install to a specific CLI
agentflow install --all         # Install to all detected CLIs
agentflow uninstall codex       # Uninstall from a specific CLI
agentflow uninstall codex --cli # Full uninstall: remove the CLI tool + purge its config directory (use --keep-config to preserve config)
agentflow uninstall --all       # Uninstall from all installed targets
agentflow status                # Show installation status
agentflow clean                 # Remove AgentFlow caches
agentflow update                # Self-update to the latest release binary
agentflow version               # Print current version + update hint

# Project rules (CLI + IDE)
agentflow rules detect
agentflow rules install codex claude

# Skills (Codex)
agentflow skill list codex
agentflow skill install codex https://skills.sh/vercel/turborepo/turborepo
agentflow skill uninstall codex turborepo

# MCP servers (global)
agentflow mcp install claude context7 --set-env=CONTEXT7_API_KEY=YOUR_API_KEY
agentflow mcp install claude playwright
agentflow mcp list claude
agentflow mcp search playwright
```

When no command is provided and stdin is a TTY, AgentFlow opens the Bubble Tea based menu. Arrow keys, `Enter`, `Space`, and `Esc` behave consistently across macOS and Windows terminals.

---

## Supported Targets

### CLI Targets

| Target | Config Directory | Rules File | Extra Integration |
|--------|------------------|------------|-------------------|
| Codex CLI | `~/.codex/` | `AGENTS.md` | `agents/reviewer.toml`, `agents/architect.toml`, `config.toml` multi-agent merge |
| Claude Code | `~/.claude/` | `CLAUDE.md` | `settings.json` hook merge / cleanup |

### IDE Targets

| Target | Rules Format |
|--------|-------------|
| Antigravity | `.agents/` directory |

---

## How It Works

```
┌──────────────────────────────────────────────────────────────┐
│                      User Input                              │
└──────────────────┬───────────────────────────────────────────┘
                   │
                   ▼
┌──────────────────────────────────────────────────────────────┐
│  ① Router: Score 5 dimensions → Route to R0/R1/R2/R3/R4     │
│     • ~command → Command handler                             │
│     • Skill/MCP match → Tool protocol                        │
│     • Otherwise → Level routing                              │
└──────────────────┬───────────────────────────────────────────┘
                   │
         ┌─────────┼─────────┬──────────┬──────────┐
         ▼         ▼         ▼          ▼          ▼
      R0 💬     R1 ⚡     R2 📝     R3 📊     R4 🏗️
      Reply     Quick     Simplified  Standard   Architecture
      only      fix       flow        flow       flow
                   │         │          │          │
                   ▼         ▼          ▼          ▼
         ┌─────────────────────────────────────────────────────┐
         │  ② EHRB Safety Gate (3 layers)                      │
         │     Keywords → Semantics → Tool Output              │
         └──────────────────┬──────────────────────────────────┘
                            │
                            ▼
         ┌─────────────────────────────────────────────────────┐
         │  ③ Stage Chain: EVALUATE → DESIGN → DEVELOP         │
         │     Module loading per stage · Sub-agent dispatch    │
         └──────────────────┬──────────────────────────────────┘
                            │
                            ▼
         ┌─────────────────────────────────────────────────────┐
         │  ④ KB Sync + Session Save                           │
         │     L1 knowledge base · L2 session · Graph updates  │
         └─────────────────────────────────────────────────────┘
```

---

## Repository Layout

```text
AgentFlow/
├── cmd/agentflow/          # CLI entrypoint
├── internal/app/           # Command dispatch + main flows
├── internal/ui/            # Bubble Tea TUI
├── internal/install/       # Target deployment / uninstall logic
├── internal/update/        # GitHub release checks + cache
├── internal/kb/            # KB, sessions, templates
├── internal/scan/          # Graph, convention, dashboard, arch scan
├── internal/targets/       # CLI targets and profiles
├── agentflow/              # Shipped prompt assets and templates
│   ├── core/               # Routing, safety, subagent, attention, hooks
│   ├── stages/             # DESIGN and DEVELOP stage modules
│   ├── functions/          # ~init, ~plan, ~exec, ~scan, ~graph, etc.
│   ├── services/           # Memory, knowledge, attention, package services
│   ├── agents/             # reviewer.toml, architect.toml role configs
│   ├── rlm/roles/          # 6 specialized sub-agent role definitions
│   ├── hooks/              # Claude hooks JSON, Codex hooks TOML
│   ├── templates/          # KB, plan, session, changelog, convention templates
│   ├── project_rules/      # IDE-specific rule files (Cursor, Windsurf, etc.)
│   └── rules/              # Caching, scaling, state, tools rules
├── embed.go                # Static asset embedding
├── install.sh              # POSIX release-binary installer
├── install.ps1             # Windows release-binary installer
└── bin/agentflow.js        # npx bootstrap for the release binary
```

---

## Development

### Requirements

- Go `1.26.0`
- Node.js `>=16` only if you want to validate the `npx` bridge

### Common commands

```bash
gofmt -w .
go test ./...
go build -o /tmp/agentflow ./cmd/agentflow
bash -n install.sh
node --check bin/agentflow.js
```

### Cross-platform release assets

Release builds are produced by `.github/workflows/release.yml` for:

- Linux `amd64`
- Linux `arm64`
- macOS `amd64`
- macOS `arm64`
- Windows `amd64`

---

## FAQ

<details>
<summary><b>Is AgentFlow still Python-based?</b></summary>
No. AgentFlow is now implemented and distributed as a Go executable. Installation, hooks, testing, and release flows use the Go CLI rather than Python runtime scripts.
</details>

<details>
<summary><b>What happens if my existing rules file is custom?</b></summary>
AgentFlow backs it up with a timestamped <code>*_bak</code> file before replacing it.
</details>

<details>
<summary><b>What does the profile flag change?</b></summary>
<code>lite</code> deploys the smallest rules set (core routing + safety), <code>standard</code> adds shared operational modules (common rules, module loading, acceptance criteria), and <code>full</code> appends sub-agent orchestration, attention control, and hook guidance.
</details>

<details>
<summary><b>What does Codex installation change?</b></summary>
It deploys <code>reviewer.toml</code> and <code>architect.toml</code> into <code>~/.codex/agents/</code> and merges the required <code>[features] multi_agent = true</code> plus <code>[agents.*]</code> sections into <code>~/.codex/config.toml</code>.
</details>

<details>
<summary><b>What does Claude installation change?</b></summary>
It merges AgentFlow hook handlers into <code>~/.claude/settings.json</code> while preserving non-AgentFlow hooks.
</details>

<details>
<summary><b>How does AgentFlow differ from plain AGENTS.md / CLAUDE.md files?</b></summary>
AgentFlow is a complete workflow system, not just a rules file. It adds 5-level intelligent routing, EHRB three-layer safety detection, three-layer persistent memory (user/project/session), knowledge graph memory, automatic convention extraction, architecture scanning, sub-agent orchestration with 6 specialized roles, and structured stage chains (EVALUATE → DESIGN → DEVELOP). A plain rules file gives instructions; AgentFlow gives the AI a full operating framework.
</details>

<details>
<summary><b>Can I use AgentFlow with multiple CLIs simultaneously?</b></summary>
Yes. <code>agentflow install --all</code> deploys to every detected CLI target. Each target gets its own rule file format and extra integrations (hooks, multi-agent config, etc.). The project-level knowledge base in <code>.agentflow/</code> is shared across all CLIs.
</details>

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for Go development workflow, testing, and release expectations.

## License

[MIT](LICENSE)
