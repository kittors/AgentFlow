# Session Summary Template (L2 Memory)

## Session Summary Format

```markdown
# Session: {session_id}

Date: {YYYY-MM-DD HH:MM}
Duration: {elapsed}
Route Level: {R0|R1|R2|R3|R4}
Mode: {INTERACTIVE|DELEGATED|TURBO}

## tasks摘要
{what_was_accomplished}

## key decisions
- {decision_1}: {rationale}
- {decision_2}: {rationale}

## modified files
- `{file_path}` — {change_description}

## 发现的问题
- {issue_1}
- {issue_2}

## not completeitems
- [ ] {pending_task_1}
- [ ] {pending_task_2}

## 上下文传递
{context_for_next_session}
```

## Auto-Save Triggers

```yaml
triggers:
  - stage_transition    # DESIGN → DEVELOP 等stage transitionwhen
  - user_explicit       # 用户执行 ~memory save
  - context_overflow    # context window超过 80% when
  - session_end         # 会话结束when（Hooks 触发）
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
