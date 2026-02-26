---
name: agentflow
description: "AgentFlow — Multi-CLI agent workflow system with routing, safety detection, sub-agent orchestration, and knowledge graph memory."
version: "1.0.0"
tags:
  - workflow
  - routing
  - safety
  - memory
  - knowledge-graph
  - convention-check
  - architecture-scan
commands:
  - "~init"
  - "~auto"
  - "~plan"
  - "~exec"
  - "~status"
  - "~review"
  - "~scan"
  - "~conventions"
  - "~graph"
  - "~dashboard"
  - "~memory"
  - "~rlm"
  - "~validatekb"
---

# AgentFlow Skill

AgentFlow is a multi-CLI agent workflow system that keeps going until tasks are implemented and verified.

## Features

- **5-Level Routing (R0–R4)**: Every input is scored on 5 dimensions and routed proportionally — simple queries stay fast, complex tasks get full process, architecture-level tasks get dedicated review.
- **RLM Sub-Agent Orchestration**: 6 specialized roles (reviewer, synthesizer, kb_keeper, pkg_keeper, writer, architect) plus native CLI sub-agents, dispatched based on task complexity.
- **EHRB Safety Detection**: Three-layer safety (keyword scan, semantic analysis, tool output inspection) catches destructive operations before execution.
- **Three-Layer Memory**: L0 user memory, L1 project knowledge base, L2 session summaries — context survives across sessions.
- **Knowledge Graph Memory**: Graph-based project memory with progressive disclosure queries (AgentFlow unique).
- **Convention Extraction**: Automatic coding pattern discovery from your codebase (AgentFlow unique).
- **Architecture Scanning**: Proactive issue detection — large files, circular deps, missing tests (AgentFlow unique).

## Quick Commands

| Command | Description |
|---------|-------------|
| `~init` | Initialize project knowledge base |
| `~auto` | Auto-execute with full workflow |
| `~plan` | Plan only, stop before development |
| `~exec` | Execute existing plan |
| `~status` | Show workflow status |
| `~review` | Code review |
| `~scan` | Architecture scan (AgentFlow unique) |
| `~conventions` | Extract/check coding conventions (AgentFlow unique) |
| `~graph` | Knowledge graph operations (AgentFlow unique) |
| `~dashboard` | Generate project dashboard (AgentFlow unique) |
| `~memory` | Manage memory layers |
| `~rlm` | Sub-agent management |
| `~validatekb` | Validate knowledge base consistency |
