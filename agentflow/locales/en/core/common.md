# G6 | Common Rules

> Split from AGENTS.md. Contains term mapping, state variables, turn control, task symbols, and persistence rules.
> Load timing: Auto-loaded on first entry into R2/R3/R4 flow.

## Term Mapping (Stage Names)

```yaml
EVALUATE: Requirement evaluation  # R4-only deep evaluation
DESIGN: Solution design
DEVELOP: Development implementation
```

## State Variable Definitions

```yaml
ROUTE_LEVEL: R0|R1|R2|R3|R4
WORKFLOW_MODE: INTERACTIVE|DELEGATED|DELEGATED_PLAN|TURBO
CURRENT_STAGE: EVALUATE|DESIGN|DEVELOP|COMPLETE
TASK_COMPLEXITY: simple|moderate|complex|architect  # architect is AgentFlow-added level
KB_SKIPPED: true|false
GRAPH_MODE: true|false  # AgentFlow enhanced
```

## Turn Control Rules (MUST)

```yaml
⛔ END_TURN rules:
  - After R2/R3/R4 confirmation info output
  - After follow-up questions output
  - When approach selection requires user decision
  - When EHRB detects risk

Prohibit END_TURN:
  - During R0 replies
  - During R1 execution (except EHRB)
  - Between stages in DELEGATED mode
  - During TURBO mode (never interrupt, including EHRB)
  - During tool path execution
```

## Task Status Symbols

```yaml
Display symbols: ⬜ Pending | 🔄 In progress | ✅ Complete | ❌ Failed | ⏭️ Skipped | ⚠️ Degraded execution
tasks.md checklist: [ ] Pending | [/] In progress | [x] Complete | [!] Failed
```

## Persistence Commitment Rules (CRITICAL — globally effective)

> This rule applies to all stages and all commands.

```yaml
Core principle:
  Any operation that the flow requires to "save", "write", "generate", or "create"
  MUST use file operation tools (create_file / write_to_file / shell redirect) to actually execute.
  Never substitute file writes with terminal conversation output.

Applicable scenarios:
  - ~plan generates plan package → must write proposal.md + tasks.md to file system
  - ~init initializes knowledge base → must create actual files (INDEX.md, context.md, modules/*)
  - KB sync after DEVELOP → must update CHANGELOG.md and module docs
  - Session end → must write session summary to .agentflow/sessions/

Verification method:
  After every "save/write" operation, use ls or file read tools to confirm file exists and is non-empty.
  If file doesn't exist, immediately re-execute the write operation.

Prohibited behavior:
  - "The plan has been output above" → ❌ cannot substitute for file write
  - "See the content above" → ❌ cannot substitute for file write
  - Creating empty directories without populating content → ❌ equals initialization failure
  - Skipping verification steps → ❌ not allowed
```
