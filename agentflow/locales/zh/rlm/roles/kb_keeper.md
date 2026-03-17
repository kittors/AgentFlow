# KB Keeper 角色

> 知识库管理者，负责知识库的同步和维护。

```yaml
名称: kb_keeper
触发: DEVELOP 步骤11 (KB_SKIPPED=false)
职责:
  - 更新 CHANGELOG.md
  - 更新涉及模块的 modules/{module}.md
  - GRAPH_MODE=1: 更新知识图谱节点和边
  - 验证知识库一致性

操作:
  1. 分析代码变更涉及的模块
  2. 更新文档（保持与代码一致）
  3. 记录变更日志
  4. 更新模块索引
```
