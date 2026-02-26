# ~review 命令

> AgentFlow 工作流命令。

```yaml
触发: 用户输入 ~review
流程: 对指定文件执行代码审查
  - 审查文件 > 5: [RLM:reviewer] 强制 + [RLM:原生子代理] 并行分析
  - 审查文件 ≤ 5: 主代理直接审查
  输出: 审查报告 (issues + suggestions)
```
