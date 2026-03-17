# G9+G10 | Sub-Agent Orchestration & Invocation Channels

> Split from AGENTS.md. Merges G9 (Sub-Agent Orchestration) and G10 (Invocation Channels).
> Load timing: Loaded when TASK_COMPLEXITY ≥ moderate and sub-agent invocation is needed.

---

## G9 | Sub-Agent Orchestration

### Complexity Determination Criteria

```yaml
Determination basis: Take the highest level across the following dimensions

| Dimension | simple | moderate | complex | architect |
|-----------|--------|----------|---------|-----------|
| Files involved | ≤3 | 4-10 | >10 | >20 |
| Modules involved | 1 | 2-3 | >3 | >5 |
| Task count | ≤3 | 4-8 | >8 | >15 |
| Cross-layer | single layer | two layers | three+ layers | full-stack+infra |
| New vs modify | pure modify | mixed | pure new/refactor | architecture-level rebuild |

Result: TASK_COMPLEXITY = simple | moderate | complex | architect
```

### Invocation Protocol (MUST)

```yaml
Role roster: reviewer, synthesizer, kb_keeper, pkg_keeper, writer, architect

Mandatory invocation rules:
  DESIGN:
    Native sub-agent — moderate/complex+ codebase scan mandatory | complex+ deep dependency analysis mandatory
    synthesizer — complex+ mandatory
    architect — R4/architect level mandatory
  DEVELOP:
    Native sub-agent — moderate/complex code changes mandatory
    reviewer — complex+ core/security modules mandatory
    kb_keeper — mandatory when KB_SKIPPED=false

Degradation: Sub-agent invocation failure → main context executes directly, mark [degraded execution]
```

{CLI_SUBAGENT_PROTOCOL}

### Sub-Agent Result Caching (AgentFlow Enhanced)

```yaml
Cache strategy:
  Purpose: Avoid sub-agents redundantly exploring the same content within a session
  Storage locations:
    explorer results → .agentflow/kb/cache/scan_result.json
    reviewer results → .agentflow/kb/cache/review_result.md
    architect results → .agentflow/kb/cache/arch_result.md
  Cache TTL: Valid within current session, auto-cleaned after session ends
  Reuse rules: Check cache before launching subsequent sub-agents, inject directly into sub-agent prompt on hit
  Example: reviewer launch finds explorer cache → inject directory structure summary into reviewer prompt
```

---

## G10 | Sub-Agent Invocation Channels

### Channel Definitions

```yaml
Channel types: native (CLI native sub-agent) | rlm (AgentFlow role)
Channel selection: Prefer native, degrade to main context simulation when unsupported
```

### Parallel Scheduling Rules

```yaml
Parallel conditions: Independent tasks (no data dependencies) + moderate/complex level
Parallel strategy:
  Code exploration: Assign by module, one sub-agent per module
  Approach ideation: R3 ≥ 2 sub-agents ideate different approaches in parallel
  Code changes: Assign by file/module, parallelize tasks without dependencies
  Testing: Assign by test suite
```

### Phased Parallel Strategy (AgentFlow Enhanced)

```yaml
Purpose: Use preceding sub-agent discoveries to improve accuracy of subsequent sub-agents, reducing redundant exploration

Two-phase pipeline:
  Phase 1 (Exploration):
    - Launch explorer sub-agent to scan project structure
    - Output: file tree, module index, entry points, dependency relationships
    - Results written to cache [→ G9 Sub-Agent Result Caching]
  Phase 2 (Analysis, parallel):
    - Based on Phase 1 results, simultaneously launch reviewer + architect analysis sub-agents
    - Advantage: Analysis sub-agents directly reference correct file paths without re-exploring directories
    - Each sub-agent prompt is injected with Phase 1 structure summary

Single-phase parallel (fallback):
  When task doesn't involve code exploration, or all sub-agents already have sufficient context, launch all in parallel directly

Decision rules:
  Complexity ≥ complex + multi-module → two-phase pipeline
  Complexity < complex or single module → single-phase parallel
```

### Sub-Agent Context Trimming (AgentFlow Enhanced)

```yaml
Purpose: Reduce redundant context inherited by sub-agents, lower token consumption

Trimming rules (by role):
  explorer: Only pass project path + scan target + KB INDEX.md summary
  reviewer: Only pass target file paths + conventions/ coding standards + explorer cache summary
  architect: Only pass KB INDEX.md + module index + dependency graph + explorer cache summary
  worker: Only pass task description + target files + related test files
  General rule: Do not pass full AGENTS.md to sub-agents, only pass that role's definition (rlm/roles/*.md or agents/*.toml)

Expected benefit: Reduce 60-80% of input token consumption (measured 503K → estimated 100-150K)
```

### Batch Spawn & Failure Handling (AgentFlow Enhanced)

```yaml
Batch Spawn protocol:
  Declarative: "Simultaneously create the following N sub-agents: [role+task list]"
  Atomicity: All spawn requests issued as a group, reducing main thread round-trips

Failure handling:
  Spawn failure: Skip failed sub-agent → continue launching remainder → mark [partial degradation]
  Sub-agent timeout: Single sub-agent exceeds 120s without output → auto-close → mark [timeout degradation]
  Sub-agent error: Sub-agent returns error → main context takes over that sub-task → mark [error degradation]
  Total failure: All sub-agents fail → degrade to main context serial execution → mark [full degradation]

Result collection:
  Strategy: Wait for all surviving sub-agents to complete then batch-collect (not close one by one)
  Timeout fallback: Total wait time cap = max(single estimated time) × 1.5
  Summary: Merge results grouped by role, annotate missing roles with [degraded/timeout]
```

### Degradation Handling

```yaml
Sub-agents unavailable: Main context executes directly
Parallelism unavailable: Serial execution
Marking: Mark [degraded execution] in tasks.md
Degradation hierarchy: Parallel sub-agents → Serial sub-agents → Main context direct execution
```
