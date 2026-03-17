# ~status 命令

> 显示当前工作流状态和计划进度。

```yaml
触发: 用户输入 ~status
输出:
  💡【AgentFlow】- 状态概览

  ── 工作流 ──
  路由级别: {ROUTE_LEVEL}
  执行模式: {WORKFLOW_MODE}
  当前阶段: {CURRENT_STAGE}
  任务复杂度: {TASK_COMPLEXITY}
  EHRB 标记: {active_risks}

  ── 知识库 ──
  知识库: {KB_SKIPPED ? "已跳过" : "已启用 (.agentflow/kb/)"}
  图模式: {GRAPH_MODE ? "已启用" : "已关闭"}

  ── 活跃计划 ──
  （如果有正在执行的方案包）
  📋 当前方案: {feature_name}
  进度: ████████░░ {done}/{total} ({percentage}%)
  任务摘要:
    [x] T1: {完成的任务标题}
    [/] T2: {进行中的任务标题}  ← 当前
    [ ] T3: {待执行的任务标题}

  （如果没有活跃方案）
  📭 无活跃计划。使用 ~plan <需求> 创建。

  ── 方案包 ──
  最近方案:
  | 方案 | 进度 | 状态 |
  |------|------|------|
  | {id} | {done}/{total} | ⬜/🔄/✅ |

  🔄 下一步: {基于当前状态的建议操作}
```
