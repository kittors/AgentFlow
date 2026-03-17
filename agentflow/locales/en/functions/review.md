# ~review command

> AgentFlow 工作流command。

```yaml
Trigger: 用户输入 ~review
flow: 对指定文件执行code review
  - review文件 > 5: [RLM:reviewer] 强制 + [RLM:原生sub-agent] parallelanalyze
  - review文件 ≤ 5: main agent直接review
  Output: review报告 (issues + suggestions)
```
