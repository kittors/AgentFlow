# G12 | Hooks 集成

> 本文件从 AGENTS.md 拆分而来，包含 Hooks 能力矩阵和降级原则。
> 加载时机: 仅供参考，不影响核心工作流。

## Hooks 能力矩阵

| Hook 事件 | Claude Code | Codex CLI | 其他 CLI |
|-----------|------------|-----------|----------|
| PreToolCall (安全检查) | ✅ | ❌ | ❌ |
| PostToolCall (进度快照) | ✅ | ❌ | ❌ |
| PostMessage (KB同步) | ✅ | ❌ | ❌ |
| Notification (更新检查) | ✅ | ✅ | ❌ |
| SessionStart (记忆加载) | ✅ | ❌ | ❌ |
| SessionEnd (会话保存) | ✅ | ❌ | ❌ |

## 降级原则

```yaml
Hooks 可用: 自动执行安全检查、进度快照、KB同步
Hooks 不可用: 功能降级但不影响核心工作流
  - 安全检查: 依赖 EHRB 规则（G2）
  - 进度快照: 手动在阶段切换时输出
  - KB同步: 在 DEVELOP 完成时手动触发
```
