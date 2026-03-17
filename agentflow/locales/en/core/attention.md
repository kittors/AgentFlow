# G11 | Attention Control

> Split from AGENTS.md. Contains attention rules and context window management.
> Load timing: Loaded when context window usage > 50% or when sub-agent orchestration is needed.

```yaml
Attention rules:
  Priority: Latest user message > Current stage module > AGENTS.md core rules > Historical context
  Context window optimization (AgentFlow enhanced):
    - When exceeding 80% context window, proactively summarize history and release early conversation
    - Prioritize retaining: Current task's tasks.md + latest EHRB detection results + active modules
    - Can release: Detailed output from completed stages, old tool outputs
  Focus maintenance:
    - Each reply's status bar must reflect the current actual state
    - Do not forget the incomplete task list
    - EHRB risk markers persist once set until explicitly cleared

Main thread wait period behavior (AgentFlow enhanced):
  Purpose: When sub-agents are running, main thread should not idle-poll; use wait time for valuable preprocessing
  Executable during wait:
    - Preload next stage module files (stages/develop.md etc.)
    - Prepare result summary templates (pre-build structure by sub-agent role)
    - Read KB historical data (cache hit checks, session summaries etc.)
    - Check conventions/ coding standards (prepare for subsequent DEVELOP stage)
  Prohibited:
    - Do not modify files during wait (avoid write conflicts with sub-agents)
    - Do not start new sub-agents (wait for current batch to complete before starting next batch)
```
