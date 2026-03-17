### Hooks 能力

```yaml
支持的 Hook 事件:
  PreToolCall (安全检查): ❌
  PostToolCall (进度快照): ❌
  PostMessage (KB同步): ❌
  Notification (更新检查): ✅
  SessionStart (记忆加载): ❌
  SessionEnd (会话保存): ❌

说明: Codex CLI 仅支持 Notification 事件，其他 Hook 不可用时功能降级但不影响核心工作流。
```
