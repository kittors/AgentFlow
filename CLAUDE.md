# AgentFlow Development Guide

## Project Overview

AgentFlow is now implemented as a Go CLI that ships embedded workflow assets for Codex, Claude, and other supported assistants.

## Architecture

- `cmd/agentflow/` - CLI entrypoint
- `internal/app/` - command dispatch and automation commands
- `internal/install/` - installer, hooks, target config merge
- `internal/kb/` - knowledge-base bootstrap, sync, sessions
- `internal/scan/` - graph, conventions, dashboard, architecture scan
- `internal/ui/` - Bubble Tea based interactive UI
- `agentflow/` - embedded prompt assets, hooks, templates, role definitions

## Conventions

- Go `1.26.0`
- Format with `gofmt`
- Prefer safe file writes through `internal/config`
- Keep CLI output bilingual where the existing code uses `Catalog.Msg(zh, en)`
- Treat `agentflow/` as shipped product assets and keep docs aligned with the real CLI behavior

## Testing

- Main test command: `go test ./...`
- Fast targeted checks: `go test ./internal/app ./internal/install ./internal/kb`
- Build smoke: `go build -o /tmp/agentflow ./cmd/agentflow`
- Script checks: `bash -n install.sh` and `node --check bin/agentflow.js`

## Key Design Decisions

- Static workflow assets are embedded into the Go binary via `embed`
- Install flows write rules, hooks, and agent role files directly from embedded assets
- Hooks and helper automation should call Go CLI commands, not runtime Python scripts
- Cross-platform behavior is validated through Go tests and GitHub Actions release matrices
