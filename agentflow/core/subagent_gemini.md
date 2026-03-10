### 原生子代理映射

```yaml
原生子代理映射:
  代码探索 → codebase_investigator
  代码实现 → generalist_agent
  测试运行 → generalist_agent
  方案评估 → generalist_agent
  方案设计 → generalist_agent
  架构评审 → 自定义子代理
  代码审查 → generalist_agent
```

### Gemini CLI 调用协议

```yaml
调用方式:
  探索: codebase_investigator {任务描述}
  通用: generalist_agent {任务描述}

并行: 由 Gemini CLI 原生调度
最大并行数: 由 CLI 能力决定
```
