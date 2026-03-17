### 原生sub-agent映射

```yaml
原生sub-agent映射:
  代码探索 → explorer Rolesub-agent
  代码实现 → worker Rolesub-agent
  测试运行 → monitor Rolesub-agent
  方案evaluate → worker Rolesub-agent
  solution design → default Rolesub-agent
  architecture review → architect Rolesub-agent
  code review → reviewer Rolesub-agent
```

### Codex CLI 调用协议

> **Prerequisites:** Codex 多代理为实验性功能，need to启用 `[features] multi_agent = true`。
> if未启用，所有sub-agent调用自动降级为主上下文执行。

```yaml
触发方式: 自然语言（Codex 自动编排sub-agent的create、路由and结果汇总）

within置Role:
  default: 通用Role（默认）
  worker: 执行型Role（实现andFixed）
  explorer: 只读探索Role（代码库analyze，sandbox_mode=read-only）
  monitor: 长when间监控Role（等待and轮询，最长1小when）

AgentFlow 自定义Role（via config.toml [agents] 配置）:
  reviewer: code review专家（安全、正确性、测试质量）
  architect: architecture review师（approach comparison、dependenciesanalyze、扩展性evaluate）

调用Example:
  单代理: "please用 explorer sub-agentanalyze src/ 目录的模块结构"
  多代理parallel: "为以下 3 个模块各派一个sub-agentparallelanalyze，完成after汇总: 1. auth 2. api 3. database"
  architecture review: "please用 architect sub-agent评审whenbefore的微服务架构方案"
  code review: "please用 reviewer sub-agentreview src/main.py 的安全性and代码质量"

manage:
  查看sub-agent: /agent command可切换and查看活跃sub-agent线程
  控制sub-agent: 直接对话即可引导、停止or关闭sub-agent

parallel: Codex 自动处理parallel调度，支持多个sub-agent同when运行
沙盒: sub-agent继承父会话的沙盒策略，explorer 默认 read-only
最大parallel数: none限制
```
