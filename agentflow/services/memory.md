# 三层记忆服务 (MemoryService)

> 管理 L0/L1/L2 三层记忆的读写和同步。

## 记忆架构

```yaml
L0 — 用户记忆（全局）:
  位置: {AGENTFLOW_ROOT}/user/profile.md
  内容: 用户偏好、常用技术栈、沟通风格、全局规则
  生命周期: 跨项目持久化
  更新时机: 用户显式修改 或 ~memory set 命令

L1 — 项目知识库（项目级）:
  位置: {KB_ROOT}/ (通过 KnowledgeService 管理)
  内容: 项目结构、模块文档、架构决策、编码规范
  生命周期: 项目级持久化
  更新时机: ~init, DEVELOP 完成后 KB 同步
  AgentFlow 增强: GRAPH_MODE=1 时使用知识图谱存储

L2 — 会话摘要（会话级）:
  位置: {KB_ROOT}/sessions/{session_id}.md 或 {AGENTFLOW_ROOT}/user/sessions/
  内容: 本次会话的关键决策、进度、上下文
  生命周期: 会话结束时自动保存
  更新时机: 阶段切换时自动保存快照
```

## 操作协议

```yaml
会话启动:
  1. 加载 L0 profile.md（如存在）
  2. 加载最近1个 L2 会话摘要（如存在且相关）
  3. L1 按需通过 KnowledgeService 加载

会话结束:
  1. 生成会话摘要:
     - 执行的任务
     - 关键决策
     - 遇到的问题和解决方案
     - 未完成的事项
  2. 保存到 L2 sessions/

L0 更新 (~memory set <key> <value>):
  - 更新 profile.md 中对应字段
  - 立即生效

L2 快照（阶段切换时）:
  - 当前阶段: {CURRENT_STAGE}
  - 任务进度: {completed}/{total}
  - 关键上下文: 方案选择、EHRB 标记等
```

## 会话摘要模板

```markdown
# Session: {session_id}
Date: {date}
Duration: {duration}

## Tasks
- {task_list}

## Decisions
- {decision_list}

## Issues
- {issue_list}

## Next Steps
- {next_steps}
```
