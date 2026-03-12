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

## Overview

AgentFlow is now shipped as a **Go CLI** with embedded workflow assets. The executable bundles `AGENTS.md`, `SKILL.md`, stage modules, templates, hooks, and agent role files, so installation no longer depends on Python, `uv`, `pip`, or PyInstaller.

The current Go implementation covers:

- Cross-platform CLI binaries for macOS, Linux, and Windows
- Interactive TUI menus for `install`, `uninstall`, `update`, `status`, and `clean`
- Embedded deployment assets for Codex, Claude, Gemini, Qwen, Grok, and OpenCode
- Profile-aware rules generation (`lite`, `standard`, `full`)
- Claude hook merge / cleanup and Codex multi-agent config injection
- Go-native tests, CI, release artifacts, KB/session/template helpers, and scanning utilities

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

The installer downloads the latest published release binary. If you still have an older `uv`/Python install earlier on `PATH`, reopen the terminal or run `export PATH="$HOME/.agentflow/bin:$PATH" && hash -r`, then verify with `which agentflow`.

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

## Quick Start

```bash
agentflow                       # Interactive TUI
agentflow install codex         # Install to a specific CLI
agentflow install --all         # Install to all detected CLIs
agentflow uninstall codex       # Uninstall from a specific CLI
agentflow uninstall --all       # Uninstall from all installed targets
agentflow status                # Show installation status
agentflow clean                 # Remove AgentFlow caches
agentflow update                # Self-update to the latest release binary
agentflow version               # Print current version + update hint
```

When no command is provided and stdin is a TTY, AgentFlow opens the Bubble Tea based menu. Arrow keys, `Enter`, `Space`, and `Esc` behave consistently across macOS and Windows terminals.

## Supported Targets

| Target | Config Directory | Rules File | Extra Integration |
|--------|------------------|------------|-------------------|
| Codex CLI | `~/.codex/` | `AGENTS.md` | `agents/reviewer.toml`, `agents/architect.toml`, `config.toml` multi-agent merge |
| Claude Code | `~/.claude/` | `CLAUDE.md` | `settings.json` hook merge / cleanup |
| Gemini CLI | `~/.gemini/` | `GEMINI.md` | Rules + embedded module deployment |
| Qwen CLI | `~/.qwen/` | `QWEN.md` | Rules + embedded module deployment |
| Grok CLI | `~/.grok/` | `GROK.md` | Rules + embedded module deployment |
| OpenCode | `~/.config/opencode/` | `AGENTS.md` | Rules + embedded module deployment |

## Key Features

### 5-level routing

Every input is scored on action need, goal clarity, decision scope, impact range, and EHRB risk, then routed to `R0` through `R4`.

### EHRB safety

AgentFlow adds a three-layer safety gate for destructive commands, secrets, permissions, production risk, and suspicious tool output.

### Knowledge base and sessions

Project state is stored in `.agentflow/`:

- `.agentflow/kb/`
- `.agentflow/kb/plan/`
- `.agentflow/kb/graph/`
- `.agentflow/kb/conventions/`
- `.agentflow/sessions/`

### Embedded deployment assets

The Go binary embeds:

- `AGENTS.md`
- `SKILL.md`
- `agentflow/stages/`
- `agentflow/functions/`
- `agentflow/services/`
- `agentflow/templates/`
- `agentflow/hooks/`
- `agentflow/agents/`
- `agentflow/core/`

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
├── embed.go                # Static asset embedding
├── install.sh              # POSIX release-binary installer
├── install.ps1             # Windows release-binary installer
└── bin/agentflow.js        # npx bootstrap for the release binary
```

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

## FAQ

<details>
<summary><b>Is AgentFlow still Python-based?</b></summary>
No. AgentFlow is now implemented and distributed as a Go executable. Installation, hooks, testing, and release flows use the Go CLI rather than Python runtime scripts.
</details>

<details>
<summary><b>What happens if my existing rules file is custom?</b></summary>
AgentFlow backs it up with a timestamped `*_bak` file before replacing it.
</details>

<details>
<summary><b>What does the profile flag change?</b></summary>
`lite` deploys the smallest rules set, `standard` adds shared operational modules, and `full` appends sub-agent, attention, and hook guidance.
</details>

<details>
<summary><b>What does Codex installation change?</b></summary>
It deploys `reviewer.toml` and `architect.toml` into `~/.codex/agents/` and merges the required `[features] multi_agent = true` plus `[agents.*]` sections into `~/.codex/config.toml`.
</details>

<details>
<summary><b>What does Claude installation change?</b></summary>
It merges AgentFlow hook handlers into `~/.claude/settings.json` while preserving non-AgentFlow hooks.
</details>

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for Go development workflow, testing, and release expectations.

## License

[MIT](LICENSE)
