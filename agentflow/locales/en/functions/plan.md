# ~plan Command

> Planning mode — complete solution design only, do not enter development stage.

```yaml
Trigger: User inputs ~plan [subcommand|requirement description]
Subcommands:
  ~plan <requirement>: Run R3 DESIGN flow for requirement, stop after completion
  ~plan list: List all existing plan packages and their completion progress
  ~plan show [id]: View details and tasks.md status of specified plan package
  ~plan (no args): Equivalent to ~plan list
```

## ~plan <requirement>

```yaml
Flow:
  WORKFLOW_MODE = DELEGATED_PLAN
  1. Evaluate requirement per R3 requirement evaluation flow
  2. Enter DESIGN stage (load stages/design.md)
  3. Generate tasks.md (checklist format, see below)
  4. Save plan package to .agentflow/kb/plan/YYYYMMDDHHMM_<feature>/
  5. ⛔ END_TURN — do not enter DEVELOP

Output:
  💡【AgentFlow】- DESIGN ✅: Plan design complete
  📋 Plan package: .agentflow/kb/plan/{id}/
  📄 proposal.md — Plan details
  📄 tasks.md — Task checklist ({total} tasks)
  🔄 Next: ~exec to execute this plan | ~plan show {id} to view details
```

<file_persistence_protocol>

## File Persistence Protocol (CRITICAL — violating this protocol equals task failure)

> **Core principle**: The plan must be written to files first, then output a summary. Never only output the plan in conversation without persisting to disk.

### Mandatory Execution Steps

```yaml
Step 4a: Create plan package directory
  Command: mkdir -p .agentflow/kb/plan/YYYYMMDDHHMM_<feature>/

Step 4b: Write proposal.md
  - Use file write tools to write plan content to proposal.md
  - Content includes: plan summary, technology selection, module decomposition, interface definition, file change list

Step 4c: Write tasks.md
  - Use file write tools to write task checklist to tasks.md
  - Format must follow the Checklist format specification below

Step 4d: Verify files exist (MUST)
  - Use file read tools or ls command to confirm both files are created
  - If files don't exist, re-execute Steps 4b-4c
  - Include file paths as evidence in output
```

### Prohibited Behavior

```yaml
NEVER:
  - Output complete plan in conversation but not write to files
  - Use "the plan is currently in the conversation" as a deliverable
  - Skip Step 4d verification
  - Output "DESIGN ✅" before Step 4 is complete

Violation consequence: Does not meet ~plan completion criteria, task is considered failed
```

</file_persistence_protocol>

## ~plan list

```yaml
Output:
  💡【AgentFlow】- Plan List

  | # | Plan Package | Created | Progress | Status |
  |---|--------------|---------|----------|--------|
  | 1 | {id}_{feature} | {time} | {done}/{total} | ⬜/🔄/✅ |

  When list is empty:
    📭 No plan packages found.
    💡 Use ~plan <requirement> to create a new plan, or describe your requirements directly.

  🔄 Next: ~plan show {id} to view details | ~exec to execute plan
```

## ~plan show [id]

```yaml
Output:
  💡【AgentFlow】- Plan Details: {feature}

  📄 Plan summary: {proposal overview}
  📋 Task checklist:
    [x] T1: {completed task}
    [/] T2: {in-progress task}
    [ ] T3: {pending task}
    [!] T4: {failed task}

  Progress: {done}/{total} ({percentage}%)
  🔄 Next: ~exec to continue execution | ~exec {id} to execute this plan
```

## tasks.md Checklist Format (CRITICAL)

```yaml
Format specification:
  File location: .agentflow/kb/plan/{id}/tasks.md
  Symbols:
    "[ ]": Pending
    "[/]": In progress
    "[x]": Complete
    "[!]": Failed

  Each task must include:
    - id (T1, T2, ...)
    - Title
    - Files involved
    - Dependencies (deps: [T1] or none)
    - Acceptance criteria

  Example:
    ## Task Checklist

    - [x] T1: Create data models | Files: src/models.py | deps: none
    - [/] T2: Implement API endpoints | Files: src/api.py | deps: [T1]
    - [ ] T3: Write tests | Files: tests/test_api.py | deps: [T2]
    - [!] T4: Configure CI/CD | Files: .github/workflows/ | deps: none

Update rules (CRITICAL):
  - Task starts execution → mark [/]
  - Task passes acceptance → mark [x]
  - Task fails and exceeds retry limit → mark [!]
  - Immediately write to tasks.md after every mark change
  - DEVELOP stage Step 5e must update tasks.md
```
