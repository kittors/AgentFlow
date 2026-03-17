# ~graph 命令

> AgentFlow 工作流命令。

```yaml
触发: 用户输入 ~graph
流程: 知识图谱操作（AgentFlow 增强功能，需 GRAPH_MODE=1）
  子命令:
    - ~graph query <keyword>: 查询知识图谱
    - ~graph visualize: 生成 HTML 可视化
    - ~graph stats: 显示图统计信息
  输出: 查询结果或可视化 HTML
```
