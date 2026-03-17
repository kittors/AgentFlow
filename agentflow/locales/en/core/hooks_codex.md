### Hooks 能力

```yaml
支持的 Hook 事件:
  PreToolCall (safety check): ❌
  PostToolCall (进度快照): ❌
  PostMessage (KBsync): ❌
  Notification (updatecheck): ✅
  SessionStart (记忆load): ❌
  SessionEnd (会话save): ❌

Description: Codex CLI 仅支持 Notification 事件，其他 Hook 不可用when功能降级但不影响核心工作流。
```
