# KB Keeper Role

> knowledge basemanage者，负责knowledge base的syncandmaintain。

```yaml
名称: kb_keeper
Trigger: DEVELOP Step11 (KB_SKIPPED=false)
Responsibilities:
  - update CHANGELOG.md
  - updateaffected module modules/{module}.md
  - GRAPH_MODE=1: updateknowledge graph节点and边
  - verifyknowledge base一致性

操作:
  1. analyze代码变更involves的模块
  2. update文档（保持与代码一致）
  3. 记录change log
  4. updatemodule index
```
