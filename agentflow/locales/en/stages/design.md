# DESIGN Stage — Solution Design

> This module is loaded after R2/R3/R4 routing confirmation, defining the complete execution steps for the design stage.

## Execution Flow

### Phase 1: Context Collection

```yaml
Step 1: Requirement parsing
  - Extract core requirements from confirmation info
  - Identify key constraints and boundary conditions

Step 2: Project context reading
  - Read {KB_ROOT}/INDEX.md (project overview)
  - Read {KB_ROOT}/context.md (tech stack info)
  - Read {KB_ROOT}/modules/_index.md (module list)
  - When GRAPH_MODE=1: Execute graph query to get related nodes

Step 3: Initial complexity assessment
  - Evaluate per G9 complexity determination criteria
  - Set TASK_COMPLEXITY

Step 4: Codebase scan (moderate/complex/architect with existing code)
  - [RLM: native sub-agent] Scan involved modules and files
  - Collect: directory structure, key interfaces, dependency relationships
  - Skip for simple or new project

Step 5: When CONVENTION_CHECK=1
  - Read {KB_ROOT}/conventions/extracted.json
  - Record coding conventions for involved modules

Step 6: Complexity confirmation
  - Confirm TASK_COMPLEXITY based on scan results
  - complex/architect + dependencies > 5 modules: [RLM: native sub-agent] deep dependency analysis
```

### Phase 2: Approach Ideation

```yaml
R2 simplified flow (skip multi-approach comparison):
  Step 7: Design single approach directly
    - Technology selection
    - Module decomposition
    - Interface definition
    - File change list
  Step 8: → Phase 3

R3/R4 standard flow (with multi-approach comparison):
  Step 7: Design constraint compilation
    - Performance requirements
    - Compatibility constraints
    - Security requirements

  Step 8: Approach draft generation
    - At least 2 approaches
    - R3: Each approach ideated independently by one sub-agent [RLM: native sub-agent, parallel]
    - R4 additionally: [RLM: architect] architecture review

  Step 9: Approach comparison
    - Dimensions: implementation complexity, maintainability, performance, extensibility
    - complex/architect + evaluation dimensions ≥ 3: [RLM: synthesizer] comprehensive analysis

  Step 10: Approach selection
    - INTERACTIVE: Output approach comparison → ⛔ END_TURN → wait for user selection
    - DELEGATED / TURBO: Auto-select recommended approach
```

### Phase 3: Detailed Planning

```yaml
Step 11: Generate task checklist (tasks.md)
  - Each task includes: id, title, description, files involved, dependencies, acceptance criteria
  - Ordered by dependency

Step 12: Plan package packaging (CRITICAL — must write to files)
  - [RLM: pkg_keeper] Create plan package via PackageService
  - Save to {KB_ROOT}/plan/YYYYMMDDHHMM_<feature>/
  - Specific operations:
    a. mkdir -p {KB_ROOT}/plan/YYYYMMDDHHMM_<feature>/
    b. Use file write tools to write proposal.md to plan package directory
    c. Use file write tools to write tasks.md to plan package directory
    d. Missing any file = Phase 3 not complete

Step 12a: File write verification (MUST — cannot skip)
  - Verify proposal.md and tasks.md files actually exist and are non-empty
  - Command: ls -la {KB_ROOT}/plan/YYYYMMDDHHMM_<feature>/
  - If verification fails, re-execute Step 12
  - Prohibited: Output "DESIGN ✅" before file verification passes

Step 13: Stage transition
  - Output approach summary and task checklist (with file paths as evidence)
  - DELEGATED / TURBO: Auto-enter DEVELOP
  - DELEGATED_PLAN: ⛔ END_TURN (planning complete, do not enter development)
  - INTERACTIVE: Ask whether to enter DEVELOP → ⛔ END_TURN
```

> **⚠️ Absolutely prohibited**: Output plan content in terminal/conversation but not write to files. Plan content must be persisted to files first, then output summary and file paths in conversation.
