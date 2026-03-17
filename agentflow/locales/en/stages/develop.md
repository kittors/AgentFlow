# DEVELOP Stage — Development Implementation

> This module is loaded after DESIGN completes, defining the complete execution steps for the development stage.

## Entry Determination

```yaml
NATURAL entry (flowing from DESIGN stage):
  - Carry over TASK_COMPLEXITY and tasks.md from DESIGN stage

DIRECT entry (~exec direct execution):
  Step 1: Select plan package
  Step 2: Initial TASK_COMPLEXITY assessment
  Step 3: Load tasks.md
```

## Execution Flow

```yaml
Step 4: Environment preparation
  - Confirm working branch (if applicable)
  - Confirm development environment is ready

Step 5: Task iteration (in tasks.md order)
  For each pending task:
    a. Check if dependencies are complete
    b. Mark task as [/] in progress → immediately write to tasks.md
    c. Execute task (see Step 6)
    d. Verify task (see Step 7)
    e. Pass → mark [x] | Fail → mark [!] → immediately write to tasks.md
    ❗ Every status change must immediately update the tasks.md file

Step 6: Code changes
  simple:
    - Main agent executes directly
  moderate/complex/architect:
    - [RLM: native sub-agent] Invoke per task
    - Tasks without dependencies can be parallelized
    - CONVENTION_CHECK=1: Check coding conventions after each change

Step 7: Task verification
  - Syntax check (lint/compile)
  - Run related tests
  - Manual verification (if necessary)

Step 8: Test supplementation
  - New features: Add corresponding tests
  - Modified features: Update existing tests
  - moderate/complex: [RLM: native sub-agent] Generate tests

Step 9: Full verification
  - Run complete test suite
  - Lint check
  - CONVENTION_CHECK=1: Final convention compliance check
  - Failure: Enter fix loop (max 3 times)

Step 10: Code review
  Trigger conditions:
    - TURBO mode: Mandatory (regardless of complexity)
    - Other modes: Execute when complex/architect + core/security modules
  - [RLM: reviewer] Review code quality and security
  - Issues found: Return to Step 6 to fix

Step 11: Knowledge base sync (when KB_SKIPPED=false, CRITICAL — must actually execute)
  - [RLM: kb_keeper] Update via KnowledgeService:
    - CHANGELOG.md
    - Affected module docs modules/{module}.md
    - GRAPH_MODE=1: Update knowledge graph nodes and edges
  - Update tasks.md status
  - Helper script invocation (recommended, auto-completes module scanning and sync):
    Command: agentflow kb sync --quiet
    Failure degradation: Manually update docs under modules/ directory

Step 12: Plan package archival
  - [RLM: pkg_keeper] Update plan package status
  - Move to archive/ (if applicable)

Step 13: Completion
  - Output completion summary:
    - List of completed tasks
    - List of modified files
    - Test results
    - Knowledge base update contents
  - CURRENT_STAGE = COMPLETE
  - Save session summary (MUST):
    Command: agentflow session save --quiet --stage=COMPLETE
    Failure degradation: Manually create .agentflow/sessions/{YYYYMMDD_HHMMSS}.md with task summary, decision log, modified file list
```

## Failure Handling

```yaml
Task failure:
  - Log error to error_log
  - attempts += 1
  - attempts < max_attempts: Retry fix
    - max_attempts: TURBO=5 | other modes=3
  - attempts >= max_attempts: Mark [!], continue to next task that doesn't depend on this one
  - All subsequent tasks depend on failed task: Report blockage → ⛔ END_TURN

Full verification failure:
  - Analyze failure cause
  - Fix (max 3 loop cycles)
  - Still failing: Report issues → ⛔ END_TURN
```

## TURBO Mode Additional Assurance

```yaml
TURBO additional flow after completion:
  Step A: Complete review
    - Mandatory [RLM: reviewer] full review
    - Coverage: security, correctness, code quality, test coverage

  Step B: Complete test
    - Run complete test suite
    - lint / compile check

  Step C: Fix loop (if Step A/B finds issues)
    - Auto-fix issues
    - Re-execute Step A + B
    - Max 3 fix loop rounds

  Step D: EHRB risk report
    - Output list of all ignored EHRB risks
    - Mark risk level and recommended measures

Final output format:
  💡【AgentFlow】- Continuous Execution ✅: All tasks complete

  ## 📊 Completion Report

  ### ✅ Completed Tasks
  {List all completed tasks per tasks.md}

  ### 🔧 Problems Solved
  {List problems found and fixed during development and review}

  ### 🧪 Test Results
  {List which tests were run, pass/fail status}

  ### 🔍 Review Log
  {Review rounds, issues found, fix status}

  ### ⚠️ Risk Mitigation (EHRB Records)
  {List ignored but logged EHRB risks and recommended measures, show "None" if empty}

  ### 📝 Modified File List
  {List all modified/added/deleted files}

  🔄 Next: {follow-up suggestions}
```
