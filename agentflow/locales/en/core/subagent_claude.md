### 原生sub-agent映射

```yaml
原生sub-agent映射:
  代码探索 → Task(subagent_type="Explore")
  代码实现 → Task(subagent_type="general-purpose")
  测试运行 → Task(subagent_type="general-purpose")
  方案evaluate → Task(subagent_type="general-purpose")
  solution design → Task(subagent_type="Plan")
  architecture review → Task(subagent_type="general-purpose")
  code review → Task(subagent_type="general-purpose")
```

### Claude Code 调用协议

```yaml
Task sub-agent:
  语法: "[createsub-agent] {提示词}"  # use Task tool
  Type: Explore | Plan | general-purpose
  parallel: 最多4个同when
  Result: 等待所有sub-agent完成after汇总

Agent Teams（实验性）:
  语法: use Claude Code Agent Teams API
  场景: 多Role协作（reviewer + architect + writer）
  Degradation: Agent Teams 不可用when，降级为串行 Task sub-agent

最大parallel数: 4
```
