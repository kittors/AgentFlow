# ~status 命令

> 显示当前工作流状态。

```yaml
触发: 用户输入 ~status
输出:
  💡【AgentFlow】- 状态概览
  
  路由级别: {ROUTE_LEVEL}
  执行模式: {WORKFLOW_MODE}
  当前阶段: {CURRENT_STAGE}
  任务复杂度: {TASK_COMPLEXITY}
  任务进度: {completed}/{total}
  EHRB 标记: {active_risks}
  知识库: {KB_SKIPPED ? "已跳过" : "已启用"}
  图模式: {GRAPH_MODE ? "已启用" : "已关闭"}
```
