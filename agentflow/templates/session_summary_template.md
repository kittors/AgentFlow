# Session Summary Template (L2 Memory)

## Session Summary Format

```markdown
# Session: {session_id}

Date: {YYYY-MM-DD HH:MM}
Duration: {elapsed}
Route Level: {R0|R1|R2|R3|R4}
Mode: {INTERACTIVE|DELEGATED|TURBO}

## 任务摘要
{what_was_accomplished}

## 关键决策
- {decision_1}: {rationale}
- {decision_2}: {rationale}

## 修改文件
- `{file_path}` — {change_description}

## 发现的问题
- {issue_1}
- {issue_2}

## 未完成项
- [ ] {pending_task_1}
- [ ] {pending_task_2}

## 上下文传递
{context_for_next_session}
```

## Auto-Save Triggers

```yaml
triggers:
  - stage_transition    # DESIGN → DEVELOP 等阶段切换时
  - user_explicit       # 用户执行 ~memory save
  - context_overflow    # 上下文窗口超过 80% 时
  - session_end         # 会话结束时（Hooks 触发）
```

## L0 User Profile Template

```markdown
# User Profile

## Preferences
- Language: {zh-CN|en-US}
- Code Style: {concise|verbose}
- Review Level: {minimal|standard|thorough}

## Expertise
- Primary: {languages_and_frameworks}
- Secondary: {other_skills}

## Custom Rules
- {rule_1}
- {rule_2}
```
