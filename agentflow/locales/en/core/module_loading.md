# G7 | Module Loading

> Split from AGENTS.md. Contains the complete on-demand loading table and loading rules.
> Load timing: Auto-loaded on first entry into R2/R3/R4 flow.

## On-Demand Loading Table

| Trigger Condition | Read File |
|-------------------|-----------|
| R2/R3/R4 enters solution design | stages/design.md |
| DESIGN complete → enter DEVELOP | stages/develop.md |
| Knowledge base operations | services/knowledge.md |
| Memory operations | services/memory.md |
| Plan package operations | services/package.md |
| Attention management | services/attention.md |
| Sub-agent invocation | rlm/roles/{role_name}.md |
| ~init command | functions/init.md, services/knowledge.md |
| ~review command | functions/review.md |
| ~status command | functions/status.md |
| ~scan command | functions/scan.md |
| ~conventions command | functions/conventions.md |
| ~graph command | functions/graph.md |
| ~dashboard command | functions/dashboard.md |
| ~memory command | functions/memory.md |
| ~rlm command | functions/rlm.md |
| ~validatekb command | functions/validatekb.md |
| ~exec command | functions/exec.md |
| ~exec <requirement> (combo) | functions/exec.md, stages/design.md, stages/develop.md |
| ~auto command | Same as R3 standard flow |
| ~plan command | functions/plan.md |
| ~plan list/show | functions/plan.md |
| ~plan <requirement> | functions/plan.md, stages/design.md |
| ~help command | functions/help.md |
| ~help <command_name> | functions/help.md, functions/{command_name}.md |

## On-Demand Loading Rules

```yaml
Lazy loading: Only read corresponding module files when trigger conditions are met
Reuse: Already-read modules within the same session are not re-read
Failure degradation: When module files are missing, use base rules from AGENTS.md
```
