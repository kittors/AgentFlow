# G7 | 模块加载

> 本文件从 AGENTS.md 拆分而来，包含完整的按需读取表和加载规则。
> 加载时机: 首次进入 R2/R3/R4 流程时自动加载。

## 按需读取表

| 触发条件 | 读取文件 |
|----------|----------|
| R2/R3/R4 进入方案设计 | stages/design.md |
| DESIGN 完成 → 进入 DEVELOP | stages/develop.md |
| 知识库操作 | services/knowledge.md |
| 记忆操作 | services/memory.md |
| 方案包操作 | services/package.md |
| 注意力管理 | services/attention.md |
| 子代理调用 | rlm/roles/{角色名}.md |
| ~init 命令 | functions/init.md, services/knowledge.md |
| ~review 命令 | functions/review.md |
| ~status 命令 | functions/status.md |
| ~scan 命令 | functions/scan.md |
| ~conventions 命令 | functions/conventions.md |
| ~graph 命令 | functions/graph.md |
| ~dashboard 命令 | functions/dashboard.md |
| ~memory 命令 | functions/memory.md |
| ~rlm 命令 | functions/rlm.md |
| ~validatekb 命令 | functions/validatekb.md |
| ~exec 命令 | functions/exec.md |
| ~exec <需求> (combo) | functions/exec.md, stages/design.md, stages/develop.md |
| ~auto 命令 | 同 R3 标准流程 |
| ~plan 命令 | functions/plan.md |
| ~plan list/show | functions/plan.md |
| ~plan <需求> | functions/plan.md, stages/design.md |
| ~help 命令 | functions/help.md |
| ~help <命令名> | functions/help.md, functions/{命令名}.md |

## 按需读取规则

```yaml
延迟加载: 仅在触发条件满足时读取对应模块文件
复用: 同一会话内已读取的模块不重复读取
失败降级: 模块文件缺失时，使用 AGENTS.md 中的基本规则
```
