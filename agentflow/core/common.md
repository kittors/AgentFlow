# G6 | 通用规则

> 本文件从 AGENTS.md 拆分而来，包含术语映射、状态变量、回合控制和任务符号定义。
> 加载时机: 首次进入 R2/R3/R4 流程时自动加载。

## 术语映射（阶段名称）

```yaml
EVALUATE: 需求评估  # R4 专用深度评估
DESIGN: 方案设计
DEVELOP: 开发实施
```

## 状态变量定义

```yaml
ROUTE_LEVEL: R0|R1|R2|R3|R4
WORKFLOW_MODE: INTERACTIVE|DELEGATED|DELEGATED_PLAN|TURBO
CURRENT_STAGE: EVALUATE|DESIGN|DEVELOP|COMPLETE
TASK_COMPLEXITY: simple|moderate|complex|architect  # architect 为 AgentFlow 新增级别
KB_SKIPPED: true|false
GRAPH_MODE: true|false  # AgentFlow 增强
```

## 回合控制规则（MUST）

```yaml
⛔ END_TURN 规则:
  - R2/R3/R4 确认信息输出后
  - 追问问题输出后
  - 方案选择需用户决策时
  - EHRB 检测到风险时

禁止 END_TURN:
  - R0 回复中
  - R1 执行中（除 EHRB）
  - DELEGATED 模式阶段间
  - TURBO 模式中（完全不中断，包括 EHRB）
  - 工具路径执行中
```

## 任务状态符号

```yaml
显示符号: ⬜ 待执行 | 🔄 执行中 | ✅ 完成 | ❌ 失败 | ⏭️ 跳过 | ⚠️ 降级执行
tasks.md checklist: [ ] 待执行 | [/] 执行中 | [x] 完成 | [!] 失败
```
