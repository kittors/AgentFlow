# G9+G10 | 子代理编排与调用通道

> 本文件从 AGENTS.md 拆分而来，合并了 G9（子代理编排）和 G10（调用通道）。
> 加载时机: TASK_COMPLEXITY ≥ moderate 且需要子代理调用时加载。

---

## G9 | 子代理编排

### 复杂度判定标准

```yaml
判定依据: 取以下维度最高级别

| 维度 | simple | moderate | complex | architect |
|------|--------|----------|---------|-----------|
| 涉及文件数 | ≤3 | 4-10 | >10 | >20 |
| 涉及模块数 | 1 | 2-3 | >3 | >5 |
| 任务数 | ≤3 | 4-8 | >8 | >15 |
| 跨层级 | 单层 | 双层 | 三层+ | 全栈+基础设施 |
| 新建vs修改 | 纯修改 | 混合 | 纯新建/重构 | 架构级重建 |

结果: TASK_COMPLEXITY = simple | moderate | complex | architect
```

### 调用协议（MUST）

```yaml
角色清单: reviewer, synthesizer, kb_keeper, pkg_keeper, writer, architect
原生子代理映射:
  代码探索 → Codex: explorer 角色子代理 | Claude: Task(subagent_type="Explore") | OpenCode: @explore | Gemini: codebase_investigator | Qwen: 自定义子代理
  代码实现 → Codex: worker 角色子代理 | Claude: Task(subagent_type="general-purpose") | OpenCode: @general | Gemini: generalist_agent | Qwen: 自定义子代理
  测试运行 → Codex: monitor 角色子代理 | Claude: Task(subagent_type="general-purpose") | OpenCode: @general
  方案评估 → Codex: worker 角色子代理 | Claude: Task(subagent_type="general-purpose")
  方案设计 → Codex: default 角色子代理 | Claude: Task(subagent_type="Plan")
  架构评审 → Codex: architect 角色子代理 | Claude: Task(subagent_type="general-purpose") | 其他: 自定义子代理
  代码审查 → Codex: reviewer 角色子代理 | Claude: Task(subagent_type="general-purpose")

强制调用规则:
  DESIGN:
    原生子代理 — moderate/complex+ 代码库扫描强制 | complex+ 深度依赖分析强制
    synthesizer — complex+ 强制
    architect — R4/architect级别 强制
  DEVELOP:
    原生子代理 — moderate/complex 代码改动强制
    reviewer — complex+ 核心/安全模块强制
    kb_keeper — KB_SKIPPED=false 时强制

降级: 子代理调用失败 → 主上下文直接执行，标记 [降级执行]
```

### 子代理结果缓存（AgentFlow 增强）

```yaml
缓存策略:
  目的: 避免同一会话内子代理重复探索相同内容
  存储位置:
    explorer 结果 → .agentflow/kb/cache/scan_result.json
    reviewer 结果 → .agentflow/kb/cache/review_result.md
    architect 结果 → .agentflow/kb/cache/arch_result.md
  缓存 TTL: 当前会话内有效，会话结束后自动清理
  复用规则: 后续子代理启动前检查缓存，命中时直接注入子代理 prompt 中
  示例: reviewer 启动前已有 explorer 缓存 → 将目录结构摘要注入 reviewer prompt
```

---

## G10 | 子代理调用通道

### 调用通道定义

```yaml
通道类型: native（CLI原生子代理）| rlm（AgentFlow角色）
通道选择: 优先 native，不支持时降级到主上下文模拟
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
```

### 其他 CLI 调用协议

```yaml
OpenCode: @explore | @general 语法
Gemini: codebase_investigator | generalist_agent
Qwen: 自定义子代理
Grok: 自定义子代理（根据可用能力适配）
```

### 并行调度规则（适用所有 CLI）

```yaml
并行条件: 独立任务（无数据依赖）+ moderate/complex 级别
并行策略:
  代码探索: 按模块分配，每模块一个子代理
  方案构思: R3 ≥ 2个子代理并行构思不同方案
  代码改动: 按文件/模块分配，无依赖的任务并行
  测试: 按测试套件分配
最大并行数: 由 CLI 能力决定（Claude 4, Codex 无限制）
```

### 分阶段并行策略（AgentFlow 增强）

```yaml
目的: 利用先行子代理的发现提升后续子代理的精准度，减少重复探索

两阶段 pipeline:
  第一阶段（探索）:
    - 启动 explorer 子代理完成项目结构扫描
    - 产出: 文件树、模块索引、入口点、依赖关系
    - 结果写入缓存 [→ G9 子代理结果缓存]
  第二阶段（分析，并行）:
    - 基于第一阶段结果，同时启动 reviewer + architect 等分析子代理
    - 优势: 分析子代理直接引用正确的文件路径，无需重复探索目录
    - 每个子代理 prompt 中注入第一阶段的结构摘要

单阶段并行（回退）:
  当任务不涉及代码探索、或所有子代理已有足够上下文时，直接全部并行启动

决策规则:
  复杂度 ≥ complex + 涉及多模块 → 两阶段 pipeline
  复杂度 < complex 或 单模块 → 单阶段并行
```

### 子代理上下文裁剪（AgentFlow 增强）

```yaml
目的: 减少子代理继承的冗余上下文，降低 token 消耗

裁剪规则（按角色）:
  explorer: 仅传递项目路径 + 扫描目标 + KB INDEX.md 摘要
  reviewer: 仅传递目标文件路径 + conventions/ 编码规范 + explorer 缓存摘要
  architect: 仅传递 KB INDEX.md + 模块索引 + 依赖图 + explorer 缓存摘要
  worker: 仅传递任务描述 + 目标文件 + 相关测试文件
  通用规则: 不向子代理传递完整 AGENTS.md，只传递该角色的定义（rlm/roles/*.md 或 agents/*.toml）

预期收益: 减少 60-80% 的 input token 消耗（实测 503K → 预估 100-150K）
```

### 批量 Spawn 与故障处理（AgentFlow 增强）

```yaml
批量 Spawn 协议:
  声明式: "同时创建以下 N 个子代理: [角色+任务列表]"
  原子性: 所有 spawn 请求作为一组发出，减少主线程往返

故障处理:
  spawn 失败: 跳过失败的子代理 → 继续启动其余 → 标记 [部分降级]
  子代理超时: 单个子代理超过 120s 无输出 → 自动关闭 → 标记 [超时降级]
  子代理异常: 子代理返回错误 → 主上下文接管该子任务 → 标记 [异常降级]
  全部失败: 所有子代理均失败 → 降级为主上下文串行执行 → 标记 [全量降级]

结果收集:
  策略: 等待所有存活子代理完成后批量收集（非逐个 close）
  超时兜底: 总等待时间上限 = max(单个预估时间) × 1.5
  汇总: 按角色分组合并结果，缺失角色标注 [降级/超时]
```

### 降级处理

```yaml
子代理不可用: 主上下文直接执行
并行不可用: 串行执行
Agent Teams 不可用: 降级为 Task 子代理
标记: 在 tasks.md 标记 [降级执行]
降级层级: 并行子代理 → 串行子代理 → 主上下文直接执行
```
