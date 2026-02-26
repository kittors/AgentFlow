# DESIGN 阶段 — 方案设计

> 本模块在 R2/R3/R4 路由确认后加载，定义方案设计阶段的完整执行步骤。

## 执行流程

### Phase 1: 上下文收集

```yaml
步骤1: 需求解析
  - 从确认信息中提取核心需求
  - 识别关键约束和边界条件

步骤2: 项目上下文读取
  - 读取 {KB_ROOT}/INDEX.md（项目概述）
  - 读取 {KB_ROOT}/context.md（技术栈信息）
  - 读取 {KB_ROOT}/modules/_index.md（模块清单）
  - GRAPH_MODE=1 时: 执行图查询，获取相关节点

步骤3: 复杂度初评
  - 按 G9 复杂度判定标准评估
  - 设置 TASK_COMPLEXITY

步骤4: 代码库扫描（moderate/complex/architect 且有现有代码时）
  - [RLM:原生子代理] 扫描涉及的模块和文件
  - 收集: 目录结构、关键接口、依赖关系
  - simple 或新建项目时跳过

步骤5: CONVENTION_CHECK=1 时
  - 读取 {KB_ROOT}/conventions/extracted.json
  - 记录涉及模块的编码规范

步骤6: 复杂度确认
  - 结合扫描结果确认 TASK_COMPLEXITY
  - complex/architect + 依赖 > 5模块: [RLM:原生子代理] 深度依赖分析
```

### Phase 2: 方案构思

```yaml
R2 简化流程（跳过多方案对比）:
  步骤7: 直接设计单一方案
    - 技术选型
    - 模块划分
    - 接口定义
    - 文件变更清单
  步骤8: → Phase 3

R3/R4 标准流程（含多方案对比）:
  步骤7: 设计约束整理
    - 性能要求
    - 兼容性约束
    - 安全要求

  步骤8: 方案草案生成
    - 至少2个方案
    - R3: 每个方案由一个子代理独立构思 [RLM:原生子代理，并行]
    - R4 额外: [RLM:architect] 架构评审

  步骤9: 方案对比
    - 维度: 实现复杂度、可维护性、性能、可扩展性
    - complex/architect + 评估维度 ≥ 3: [RLM:synthesizer] 综合分析

  步骤10: 方案选择
    - INTERACTIVE: 输出方案对比 → ⛔ END_TURN → 等待用户选择
    - DELEGATED / TURBO: 自动选择推荐方案
```

### Phase 3: 详细规划

```yaml
步骤11: 生成任务清单 (tasks.md)
  - 每个任务包含: id, 标题, 描述, 涉及文件, 依赖关系, 验收标准
  - 按依赖顺序排列

步骤12: 方案包打包
  - [RLM:pkg_keeper] 通过 PackageService 创建方案包
  - 保存到 {KB_ROOT}/plan/YYYYMMDDHHMM_<feature>/

步骤13: 阶段切换
  - 输出方案摘要和任务清单
  - DELEGATED / TURBO: 自动进入 DEVELOP
  - DELEGATED_PLAN: ⛔ END_TURN（规划完成，不进入开发）
  - INTERACTIVE: 询问是否进入 DEVELOP → ⛔ END_TURN
```
