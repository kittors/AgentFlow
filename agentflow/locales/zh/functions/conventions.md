# ~conventions 命令

> AgentFlow 工作流命令。

```yaml
触发: 用户输入 ~conventions
流程: 提取/检查编码规范（AgentFlow 增强功能）
  模式:
    - ~conventions: 显示当前编码规范
    - ~conventions extract: 从代码中自动提取规范
    - ~conventions check: 检查当前代码是否符合规范
  输出: 规范列表或合规报告
```
