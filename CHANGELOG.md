# Changelog

All notable changes to AgentFlow will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-02-26

### ✨ Features
- **5-Level Routing (R0–R4)**: Every input scored on 5 dimensions — R0 direct reply, R1 fast flow, R2 simplified, R3 standard, R4 architecture-level (unique)
- **RLM Sub-Agent Orchestration**: 6 specialized roles (reviewer, synthesizer, kb_keeper, pkg_keeper, writer, architect) plus native CLI sub-agents
- **EHRB Safety Detection**: Three-layer safety — keyword scan, semantic analysis, tool output inspection
- **Three-Layer Memory**: L0 user preferences (global), L1 project knowledge base, L2 session summaries
- **Knowledge Graph Memory** (`~graph`): Graph-based project memory with progressive disclosure queries
- **Convention Extraction** (`~conventions`): Auto-discover coding patterns from your codebase
- **Architecture Scanning** (`~scan`): Proactive detection of large files, circular deps, missing tests
- **Dashboard** (`~dashboard`): HTML project status dashboard generation
- **Context Window Management**: Proactive summarization when context exceeds 80%
- **14 Workflow Commands**: `~init`, `~auto`, `~plan`, `~exec`, `~status`, `~review`, `~scan`, `~conventions`, `~graph`, `~dashboard`, `~memory`, `~rlm`, `~validatekb`

### 🛠️ Infrastructure
- **6 CLI Targets**: Claude Code, Codex CLI, Gemini CLI, OpenCode, Qwen CLI, Grok CLI
- **Multiple Install Methods**: pip, uv, npx, one-line scripts (install.sh / install.ps1)
- **Interactive CLI**: Menu-driven installation with multi-target support
- **Auto Locale Detection**: CLI messages switch between zh-CN and en-US
- **Hooks Integration**: Claude Code lifecycle hooks + Codex CLI notify hook
- **SKILL.md**: Skill discovery metadata for all CLI targets
- **Windows Support**: File-locking safe operations with rename-aside fallback
- **CI/CD**: GitHub Actions for lint, test (Python 3.10–3.13, macOS/Linux/Windows), and auto-release
