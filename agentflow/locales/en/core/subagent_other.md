### 原生sub-agent映射

```yaml
原生sub-agent映射:
  代码探索 → 自定义sub-agent
  代码实现 → 自定义sub-agent
  测试运行 → 自定义sub-agent
  方案evaluate → 自定义sub-agent
  solution design → 自定义sub-agent
  architecture review → 自定义sub-agent
  code review → 自定义sub-agent
```

### 调用协议

```yaml
调用方式: 自定义sub-agent（based on可用能力适配）
parallel: 由 CLI 原生调度
最大parallel数: 由 CLI 能力决定
Degradation: sub-agent不可用when，主上下文直接执行
```
