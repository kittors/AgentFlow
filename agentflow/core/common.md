# G6 | 通用规则

> 本文件从 AGENTS.md 拆分而来，包含术语映射、状态变量、回合控制、任务符号和持久化规则。
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

## 持久化承诺规则（CRITICAL — 全局生效）

> 此规则适用于所有阶段和所有命令。

```yaml
核心原则:
  凡是流程要求"保存"、"写入"、"生成"、"创建"的操作，
  必须使用文件操作工具（create_file / write_to_file / shell 重定向）实际执行。
  绝不能用终端对话输出替代文件写入。

适用场景:
  - ~plan 生成方案包 → 必须写入 proposal.md + tasks.md 到文件系统
  - ~init 初始化知识库 → 必须创建实际文件（INDEX.md, context.md, modules/*）
  - DEVELOP 完成后 KB 同步 → 必须更新 CHANGELOG.md 和模块文档
  - 会话结束 → 必须写入会话摘要到 .agentflow/sessions/

验证方法:
  每次"保存/写入"操作后，用 ls 或文件读取工具确认文件存在且非空。
  如果文件不存在，立即重新执行写入操作。

禁止行为:
  - "方案已在上面输出" → ❌ 不能替代文件写入
  - "请看上面的内容" → ❌ 不能替代文件写入
  - 只创建空目录不填充内容 → ❌ 等于初始化失败
  - 跳过验证步骤 → ❌ 不允许
```
