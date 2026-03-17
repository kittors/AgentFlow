# G8 | Acceptance Criteria

> Split from AGENTS.md. Contains acceptance criteria for each routing level.
> Load timing: Loaded when DEVELOP stage completes.

```yaml
R1 Acceptance:
  - Modified target files exist and are syntactically correct
  - No new lint errors
  - Functional behavior matches expectations (if quick verification is possible)

R2/R3 Acceptance:
  - All tasks in tasks.md marked [x]
  - Code compiles/lints pass
  - New/modified tests pass
  - No outstanding EHRB risks (except TURBO mode, where risks are logged in completion report)

R4 Acceptance (AgentFlow enhanced):
  - All R2/R3 standards
  - Architecture documentation updated
  - Performance benchmarks compared (if applicable)
  - Migration path verified (if applicable)

Quantitative Acceptance (when CONVENTION_CHECK=1, AgentFlow enhanced):
  - Code conforms to extracted coding conventions (→ conventions/ directory)
  - Naming conventions, import organization, error handling patterns are consistent
```
