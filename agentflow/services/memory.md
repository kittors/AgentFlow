# 三层记忆服务 (MemoryService)

> 管理 L0/L1/L2 三层记忆的读写和同步。

## 记忆架构

```yaml
L0 — 用户记忆（全局）:
  位置: ~/.agentflow/user/profile.md
  内容: 用户偏好、常用技术栈、沟通风格、全局规则
  生命周期: 跨项目持久化
  更新时机: 用户显式修改 或 ~memory set 命令

L1 — 项目知识库（项目级）:
  位置: .agentflow/kb/ (通过 KnowledgeService 管理)
  内容: 项目结构、模块文档、架构决策、编码规范
  生命周期: 项目级持久化（跨会话共享）
  更新时机: ~init, DEVELOP 完成后 KB 同步
  AgentFlow 增强: GRAPH_MODE=1 时使用知识图谱存储

L2 — 会话记录（会话级）:
  位置: .agentflow/sessions/
  内容: 本次会话的任务进度、关键决策、遇到的问题、上下文快照
  生命周期: 会话结束时自动保存
  更新时机: 阶段切换时自动保存快照
  注意: 会话数据不在 kb/ 下，因为是会话级非共享数据
```

## 操作协议

```yaml
会话启动:
  1. 加载 L0 profile.md（如存在）
  2. 加载最近 1 个 L2 会话摘要（如存在且相关）
     命令: 直接读取 .agentflow/sessions/ 下最新的 .md 文件
     或通过宿主 CLI 的文件读取工具加载最近会话摘要
  3. L1 按需通过 KnowledgeService 加载

阶段切换时自动保存快照（MUST）:
  触发条件: CURRENT_STAGE 发生变化时（如 DESIGN → DEVELOP）
  保存内容:
    - 当前阶段: {CURRENT_STAGE}
    - 任务进度: {completed}/{total}
    - 关键上下文: 方案选择、EHRB 标记等
  保存位置: .agentflow/sessions/{YYYYMMDD_HHMMSS}_snap.md
  命令: agentflow session save --quiet --stage={CURRENT_STAGE}
  失败降级: 手动创建快照文件

会话结束（MUST — 不可跳过）:
  1. 生成会话摘要:
     - 执行的任务列表
     - 关键决策
     - 遇到的问题和解决方案
     - 修改的文件清单
     - 未完成的事项
  2. 保存到 .agentflow/sessions/{YYYYMMDD_HHMMSS}.md
  3. 验证: 确认文件已创建且非空

L0 更新 (~memory set <key> <value>):
  - 更新 profile.md 中对应字段
  - 立即生效
```

## 会话摘要模板

```markdown
# Session: {session_id}
Date: {date}
Stage: {current_stage}

## Tasks
- {task_list}

## Decisions
- {decision_list}

## Issues
- {issue_list}

## Files Modified
- {file_list}

## Next Steps
- {next_steps}
```

## 路径规范（CRITICAL — 严格遵守）

```yaml
会话数据: .agentflow/sessions/           # 会话级，不在 kb/ 下
知识库:   .agentflow/kb/                 # 项目级，跨会话共享
全局记忆: ~/.agentflow/user/             # 用户级，跨项目共享

禁止: 在 .agentflow/kb/sessions/ 下存放会话数据
原因: sessions 是会话级临时数据，不属于项目知识库
```
