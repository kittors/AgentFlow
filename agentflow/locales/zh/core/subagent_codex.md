### 原生子代理映射

```yaml
原生子代理映射:
  代码探索 → explorer 角色子代理
  代码实现 → worker 角色子代理
  测试运行 → monitor 角色子代理
  方案评估 → worker 角色子代理
  方案设计 → default 角色子代理
  架构评审 → architect 角色子代理
  代码审查 → reviewer 角色子代理
```

### Codex CLI 调用协议

> **前置条件:** Codex 多代理为实验性功能，需要启用 `[features] multi_agent = true`。
> 如果未启用，所有子代理调用自动降级为主上下文执行。

```yaml
触发方式: 自然语言（Codex 自动编排子代理的创建、路由和结果汇总）

内置角色:
  default: 通用角色（默认）
  worker: 执行型角色（实现和修复）
  explorer: 只读探索角色（代码库分析，sandbox_mode=read-only）
  monitor: 长时间监控角色（等待和轮询，最长1小时）

AgentFlow 自定义角色（通过 config.toml [agents] 配置）:
  reviewer: 代码审查专家（安全、正确性、测试质量）
  architect: 架构评审师（方案对比、依赖分析、扩展性评估）

调用示例:
  单代理: "请用 explorer 子代理分析 src/ 目录的模块结构"
  多代理并行: "为以下 3 个模块各派一个子代理并行分析，完成后汇总: 1. auth 2. api 3. database"
  架构评审: "请用 architect 子代理评审当前的微服务架构方案"
  代码审查: "请用 reviewer 子代理审查 src/main.py 的安全性和代码质量"

管理:
  查看子代理: /agent 命令可切换和查看活跃子代理线程
  控制子代理: 直接对话即可引导、停止或关闭子代理

并行: Codex 自动处理并行调度，支持多个子代理同时运行
沙盒: 子代理继承父会话的沙盒策略，explorer 默认 read-only
最大并行数: 无限制
```
