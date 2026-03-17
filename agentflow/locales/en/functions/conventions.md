# ~conventions command

> AgentFlow 工作流command。

```yaml
Trigger: 用户输入 ~conventions
flow: extract/checkcoding conventions（AgentFlow 增强功能）
  Mode:
    - ~conventions: 显示whenbeforecoding conventions
    - ~conventions extract: 从代码during自动extract规范
    - ~conventions check: checkwhenbefore代码是否符合规范
  Output: 规范列表or合规报告
```
