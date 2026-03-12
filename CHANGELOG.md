# Changelog

All notable changes to AgentFlow are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- Rewrote the runtime CLI from Python to Go and moved distribution to cross-platform release binaries.
- Embedded `AGENTS.md`, `SKILL.md`, workflow modules, templates, hooks, and agent role files directly into the Go executable.
- Replaced the placeholder terminal menu with a Bubble Tea based cross-platform TUI for main actions plus interactive install and uninstall selection.
- Reworked installation flows so profile-aware rules generation, Claude hook merge/cleanup, and Codex multi-agent config injection are handled natively in Go.
- Switched repository development, CI, and release documentation to Go build and Go test workflows.

### Added

- `cmd/agentflow` as the Go entrypoint.
- `internal/ui` for the interactive TUI layer.
- `internal/install`, `internal/update`, `internal/kb`, `internal/scan`, and related Go-native test suites.
- GitHub Actions matrices that build and smoke-test Go release assets for Linux, macOS, and Windows.

### Removed

- Python-first install, test, and release flow from the primary documentation path.

## [1.0.0] - 2025-02-26

### Features

- 5-Level Routing (`R0`–`R4`) with proportional workflow depth
- EHRB three-layer safety detection
- Three-layer memory model (`L0`, `L1`, `L2`)
- Knowledge graph memory via `~graph`
- Convention extraction via `~conventions`
- Architecture scanning via `~scan`
- Project dashboard generation via `~dashboard`
- Multi-CLI deployment targets for Codex, Claude, Gemini, OpenCode, Qwen, and Grok
