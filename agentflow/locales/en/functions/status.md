# ~status command

> 显示whenbefore工作流statusand计划进度。

```yaml
Trigger: 用户输入 ~status
Output:
  💡【AgentFlow】- status概览

  ── 工作流 ──
  路由级别: {ROUTE_LEVEL}
  执行Mode: {WORKFLOW_MODE}
  whenbefore阶段: {CURRENT_STAGE}
  tasks复杂度: {TASK_COMPLEXITY}
  EHRB 标记: {active_risks}

  ── knowledge base ──
  knowledge base: {KB_SKIPPED ? "已skip" : "已启用 (.agentflow/kb/)"}
  图Mode: {GRAPH_MODE ? "已启用" : "已关闭"}

  ── 活跃计划 ──
  （if有正在执行的plan package）
  📋 whenbefore方案: {feature_name}
  进度: ████████░░ {done}/{total} ({percentage}%)
  tasks摘要:
    [x] T1: {完成的taskstitle}
    [/] T2: {进行during的taskstitle}  ← whenbefore
    [ ] T3: {pending的taskstitle}

  （ifno活跃方案）
  📭 none活跃计划。use ~plan <需求> create。

  ── plan package ──
  最近方案:
  | 方案 | 进度 | status |
  |------|------|------|
  | {id} | {done}/{total} | ⬜/🔄/✅ |

  🔄 Next: {基于whenbeforestatus的建议操作}
```
