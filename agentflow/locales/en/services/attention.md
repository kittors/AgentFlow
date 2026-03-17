# 注意力跟踪服务 (AttentionService)

> managewhenbefore会话的关注statusand进度快照。

## status跟踪

```yaml
whenbeforeStatus:
  ROUTE_LEVEL: {R0-R4}
  WORKFLOW_MODE: {INTERACTIVE|DELEGATED|DELEGATED_PLAN}
  CURRENT_STAGE: {EVALUATE|DESIGN|DEVELOP|COMPLETE}
  TASK_COMPLEXITY: {simple|moderate|complex|architect}
  TASK_PROGRESS: {completed}/{total}
  EHRB_FLAGS: {active_risks}

context window监控（AgentFlow 增强）:
  - whenbeforeuse: {estimated_tokens}
  - 阈值: 80% when触发压缩
  - 优先保留: tasks.md + EHRB + 活跃模块
```

## status栏generate

```yaml
based onwhenbeforestatus自动generate每个回复的首行:
  R0: 💬【AgentFlow】- 回复: {summary}
  R1: ⚡【AgentFlow】- 快速Fixed: {summary}
  R2: 📝【AgentFlow】- {stage} [{progress}]: {summary}
  R3: 📊【AgentFlow】- {stage} [{progress}]: {summary}
  R4: 🏗️【AgentFlow】- {stage} [{progress}]: {summary}

进度update:
  DESIGN: 📝【AgentFlow】- solution design [Phase {1-3}]: {current_step}
  DEVELOP: 📊【AgentFlow】- development implementation [{done}/{total}]: {current_task}
```
