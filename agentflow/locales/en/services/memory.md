# 三层记忆服务 (MemoryService)

> manage L0/L1/L2 三层记忆的读写andsync。

## 记忆架构

```yaml
L0 — user memory（全局）:
  位置: ~/.agentflow/user/profile.md
  within容: 用户偏好、常用tech stack、沟通风格、全局规then
  生命周期: 跨items目persistence
  updatewhen机: 用户显式Changed or ~memory set command

L1 — items目knowledge base（items目级）:
  位置: .agentflow/kb/ (via KnowledgeService manage)
  within容: items目结构、module documentation、架构决策、coding conventions
  生命周期: items目级persistence（跨会话共享）
  updatewhen机: ~init, DEVELOP 完成after KB sync
  AgentFlow 增强: GRAPH_MODE=1 whenuseknowledge graph存储

L2 — 会话记录（会话级）:
  位置: .agentflow/sessions/
  within容: 本次会话的tasks进度、key decisions、遇到的问题、上下文快照
  生命周期: 会话结束when自动save
  updatewhen机: stage transitionwhen自动save快照
  注意: 会话数据不在 kb/ 下，因为是会话级非共享数据
```

## 操作协议

```yaml
会话启动:
  1. load L0 profile.md（如存在）
  2. load最近 1 个 L2 session summary（如存在且相关）
     Command: 直接read .agentflow/sessions/ 下最新的 .md 文件
     orvia宿主 CLI 的文件read工具load最近session summary
  3. L1 按需via KnowledgeService load

stage transitionwhen自动save快照（MUST）:
  Trigger conditions: CURRENT_STAGE 发生变化when（如 DESIGN → DEVELOP）
  Save content:
    - whenbefore阶段: {CURRENT_STAGE}
    - tasks进度: {completed}/{total}
    - 关键上下文: approach selection、EHRB 标记等
  Save location: .agentflow/sessions/{YYYYMMDD_HHMMSS}_snap.md
  Command: agentflow session save --quiet --stage={CURRENT_STAGE}
  Failure degradation: 手动create快照文件

会话结束（MUST — 不可skip）:
  1. generatesession summary:
     - 执行的tasks列表
     - key decisions
     - 遇到的问题and解决方案
     - Changed的文件清单
     - not complete的事items
  2. save到 .agentflow/sessions/{YYYYMMDD_HHMMSS}.md
  3. Verification: 确认文件已create且非空

L0 update (~memory set <key> <value>):
  - update profile.md during对应字段
  - 立即生效
```

## session summary模板

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
knowledge base:   .agentflow/kb/                 # items目级，跨会话共享
全局记忆: ~/.agentflow/user/             # 用户级，shared across projects

Prohibited: 在 .agentflow/kb/sessions/ 下存放会话数据
原因: sessions 是会话级临when数据，不属于items目knowledge base
```
