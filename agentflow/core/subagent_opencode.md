### 原生子代理映射

```yaml
原生子代理映射:
  代码探索 → @explore
  代码实现 → @general
  测试运行 → @general
  方案评估 → @general
  方案设计 → @general
  架构评审 → 自定义子代理
  代码审查 → @general
```

### OpenCode 调用协议

```yaml
调用语法:
  探索: @explore {任务描述}
  通用: @general {任务描述}

并行: 由 OpenCode 原生调度
最大并行数: 由 CLI 能力决定
```
