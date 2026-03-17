### 原生子代理映射

```yaml
原生子代理映射:
  代码探索 → Task(subagent_type="Explore")
  代码实现 → Task(subagent_type="general-purpose")
  测试运行 → Task(subagent_type="general-purpose")
  方案评估 → Task(subagent_type="general-purpose")
  方案设计 → Task(subagent_type="Plan")
  架构评审 → Task(subagent_type="general-purpose")
  代码审查 → Task(subagent_type="general-purpose")
```

### Claude Code 调用协议

```yaml
Task 子代理:
  语法: "[创建子代理] {提示词}"  # 使用 Task tool
  类型: Explore | Plan | general-purpose
  并行: 最多4个同时
  结果: 等待所有子代理完成后汇总

Agent Teams（实验性）:
  语法: 使用 Claude Code Agent Teams API
  场景: 多角色协作（reviewer + architect + writer）
  降级: Agent Teams 不可用时，降级为串行 Task 子代理

最大并行数: 4
```
