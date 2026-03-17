# ~graph command

> AgentFlow 工作流command。

```yaml
Trigger: 用户输入 ~graph
flow: knowledge graph操作（AgentFlow 增强功能，需 GRAPH_MODE=1）
  子Command:
    - ~graph query <keyword>: 查询knowledge graph
    - ~graph visualize: generate HTML 可视化
    - ~graph stats: 显示图统计信息
  Output: 查询结果or可视化 HTML
```
