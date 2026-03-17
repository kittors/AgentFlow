# ~exec Command

> Execution mode — execute existing plan packages or start directly from requirements.

```yaml
Trigger: User inputs ~exec [plan_ID|requirement description]
Subcommands:
  ~exec (no args): List available plan packages and select to execute
  ~exec <plan_ID>: Directly execute specified plan package
  ~exec <requirement>: Combo mode — plan first, then confirm, then execute
```

## ~exec (no args)

```yaml
Flow:
  1. Scan .agentflow/kb/plan/ directory
  2. Filter incomplete plan packages (status not ✅)

When plan packages exist:
  💡【AgentFlow】- Executable Plans

  | # | Plan Package | Progress | Status |
  |---|--------------|----------|--------|
  | 1 | {id}_{feature} | {done}/{total} | ⬜/🔄 |

  Enter a number to select and execute, or enter 0 to cancel.
  🔄 Next: Enter number to execute | ~exec <requirement> to create and execute new plan

When no plan packages:
  📭 No executable plan packages found.
  💡 You can:
    1. ~plan <requirement> — Create a plan first then execute
    2. ~exec <requirement> — Describe requirements directly, I'll plan then execute
    3. Describe your requirements directly, I'll handle per routing level

  🔄 Next: Describe your requirements or use ~plan to create a plan

When all complete:
  ✅ All tasks in plan packages are complete!
  💡 You can:
    1. ~plan <new requirement> — Create new plan
    2. Describe new requirements directly
  🔄 Next: Start new task
```

## ~exec <plan_ID>

```yaml
Flow:
  1. Load .agentflow/kb/plan/{id}/tasks.md
  2. Find first incomplete task ([ ] or [!])
  3. Select execution mode:
     Please select execution mode:
     1. Interactive execution: Wait for your confirmation at key decision points. (Recommended)
     2. Fully automatic execution: Auto-complete all stages, pause only on risk detection.
     3. Continuous execution: Fully auto-complete all tasks, multiple review and test cycles, output complete report when done.
  4. CURRENT_STAGE = DEVELOP
  5. Continue execution from that task per stages/develop.md

When plan ID not found:
  ❌ Plan package not found: {id}
  💡 Use ~plan list to view available plans
```

## ~exec <requirement> (Combo Mode)

```yaml
Flow:
  1. 📝 Enter DESIGN stage first (same as ~plan <requirement>)
  2. Output plan summary after generating tasks.md
  3. Ask for confirmation:
     💡【AgentFlow】- DESIGN ✅: Plan design complete

     📋 Plan summary: {summary}
     📄 Task checklist: {total} tasks

     Please select execution mode:
     1. Interactive execution: Wait for your confirmation at key decision points. (Recommended)
     2. Fully automatic execution: Auto-complete all stages, pause only on risk detection.
     3. Continuous execution: Fully auto-complete all tasks, multiple review and test cycles, output complete report when done.
     4. Save plan only, execute later.

  4. User selects 1/2/3 → Set corresponding WORKFLOW_MODE → DEVELOP
  5. User selects 4 → Save plan then ⛔ END_TURN
```
