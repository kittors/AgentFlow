# ~exec 命令

> AgentFlow 工作流命令。

```yaml
触发: 用户输入 ~exec
流程: 直接执行已有方案包
  1. 列出可用方案包 ({KB_ROOT}/plan/)
  2. 用户选择
  3. 加载 tasks.md
  4. CURRENT_STAGE = DEVELOP
  5. 按 stages/develop.md 执行
```
