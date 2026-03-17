# G12 | Hooks Integration

> Split from AGENTS.md. Contains Hooks capabilities and degradation principles.
> Load timing: Reference only, does not affect core workflow.

{HOOKS_MATRIX}

## Degradation Principles

```yaml
Hooks available: Auto-execute safety checks, progress snapshots, KB sync
Hooks unavailable: Functionality degrades but does not affect core workflow
  - Safety checks: Rely on EHRB rules (G2)
  - Progress snapshots: Manually output at stage transitions
  - KB sync: Manually trigger at DEVELOP completion
```
