# 注意力跟踪服务 (AttentionService)

> 管理当前会话的关注状态和进度快照。

## 状态跟踪

```yaml
当前状态:
  ROUTE_LEVEL: {R0-R4}
  WORKFLOW_MODE: {INTERACTIVE|DELEGATED|DELEGATED_PLAN}
  CURRENT_STAGE: {EVALUATE|DESIGN|DEVELOP|COMPLETE}
  TASK_COMPLEXITY: {simple|moderate|complex|architect}
  TASK_PROGRESS: {completed}/{total}
  EHRB_FLAGS: {active_risks}

上下文窗口监控（AgentFlow 增强）:
  - 当前使用: {estimated_tokens}
  - 阈值: 80% 时触发压缩
  - 优先保留: tasks.md + EHRB + 活跃模块
```

## 状态栏生成

```yaml
根据当前状态自动生成每个回复的首行:
  R0: 💬【AgentFlow】- 回复: {summary}
  R1: ⚡【AgentFlow】- 快速修复: {summary}
  R2: 📝【AgentFlow】- {stage} [{progress}]: {summary}
  R3: 📊【AgentFlow】- {stage} [{progress}]: {summary}
  R4: 🏗️【AgentFlow】- {stage} [{progress}]: {summary}

进度更新:
  DESIGN: 📝【AgentFlow】- 方案设计 [Phase {1-3}]: {current_step}
  DEVELOP: 📊【AgentFlow】- 开发实施 [{done}/{total}]: {current_task}
```
