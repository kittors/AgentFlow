# Contributing to AgentFlow

AgentFlow now uses a **Go-first** development workflow. This guide covers the current toolchain, repository layout, and validation steps expected before opening a pull request.

## Requirements

- Go `1.26.0`
- Git
- Node.js `>=16` only if you touch `bin/agentflow.js`
- PowerShell if you need to validate `install.ps1` locally on Windows

## Setup

```bash
git clone https://github.com/kittors/AgentFlow.git
cd AgentFlow
go test ./...
go build -o ./bin/agentflow ./cmd/agentflow
```

## Repository Structure

```text
AgentFlow/
├── cmd/agentflow/          # CLI entrypoint
├── internal/app/           # Command dispatch and runtime flows
├── internal/ui/            # Bubble Tea TUI
├── internal/install/       # Install / uninstall / config merge logic
├── internal/update/        # Release checks and cache
├── internal/kb/            # KB, plan, session helpers
├── internal/scan/          # Graph, convention, dashboard, arch scan
├── internal/targets/       # Supported CLIs and deployment profiles
├── agentflow/              # Embedded prompt assets, hooks, templates, roles
├── install.sh              # POSIX installer
├── install.ps1             # Windows installer
└── bin/agentflow.js        # npx bridge
```

## Coding Standards

- Use `gofmt` for all Go code.
- Keep user-facing output bilingual where the code already uses `Catalog.Msg(zh, en)`.
- Use `internal/config` helpers for safe writes, backups, and Windows-friendly cleanup.
- Prefer adding or updating tests with each behavior change.
- Keep embedded asset behavior aligned with the shipped files under `agentflow/`.

## Validation

Run the full local validation set before opening a PR:

```bash
gofmt -w .
go test ./...
go build -o /tmp/agentflow ./cmd/agentflow
bash -n install.sh
node --check bin/agentflow.js
```

On Windows, also validate:

```powershell
$null = [System.Management.Automation.PSParser]::Tokenize((Get-Content -Raw ./install.ps1), [ref]$null)
```

## Pull Requests

1. Create a branch from `main`.
2. Keep changes scoped and explain user-visible behavior clearly.
3. Include tests for install logic, target compatibility, parsing, or TUI behavior when applicable.
4. Update `README.md`, `README_CN.md`, `CHANGELOG.md`, or embedded assets if the behavior changed.
5. Summarize platform impact in the PR description if macOS, Linux, and Windows behavior differ.

## Release Notes

GitHub Actions builds release binaries from `.github/workflows/release.yml` for:

- Linux `amd64`
- Linux `arm64`
- macOS `amd64`
- macOS `arm64`
- Windows `amd64`

If you change asset names, install scripts, or the `npx` bootstrap path, update:

- `install.sh`
- `install.ps1`
- `bin/agentflow.js`
- `.github/workflows/release.yml`

## Reporting Issues

Please include:

- OS and architecture
- CLI target in use (`codex`, `claude`)
- Exact command you ran
- Expected behavior vs actual behavior
- Relevant config snippets if install/uninstall integration is involved

## License

By contributing, you agree that your contributions are licensed under the MIT License.
