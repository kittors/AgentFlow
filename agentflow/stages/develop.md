# DEVELOP 阶段 — 开发实施

> 本模块在 DESIGN 完成后加载，定义开发实施阶段的完整执行步骤。

## 入口判定

```yaml
NATURAL 入口（从 DESIGN 阶段流入）:
  - 沿用 DESIGN 阶段的 TASK_COMPLEXITY 和 tasks.md

DIRECT 入口（~exec 直接执行）:
  步骤1: 选择方案包
  步骤2: 首次评估 TASK_COMPLEXITY
  步骤3: 加载 tasks.md
```

## 执行流程

```yaml
步骤4: 环境准备
  - 确认工作分支（如适用）
  - 确认开发环境就绪

步骤5: 任务迭代（按 tasks.md 顺序）
  对每个待执行任务:
    a. 检查依赖是否已完成
    b. 标记任务为 [/] 执行中 → 立即写入 tasks.md
    c. 执行任务（见步骤6）
    d. 验证任务（见步骤7）
    e. 通过 → 标记 [x] | 失败 → 标记 [!] → 立即写入 tasks.md
    ❗ 每次状态变更必须立即更新 tasks.md 文件

步骤6: 代码改动
  simple:
    - 主代理直接执行
  moderate/complex/architect:
    - [RLM:原生子代理] 逐任务调用
    - 无依赖的任务可并行
    - CONVENTION_CHECK=1: 每次改动后检查编码规范

步骤7: 任务验证
  - 语法检查（lint/compile）
  - 运行相关测试
  - 手动验证（如必要）

步骤8: 测试补充
  - 新增功能: 补充对应测试
  - 修改功能: 更新现有测试
  - moderate/complex: [RLM:原生子代理] 生成测试

步骤9: 全量验证
  - 运行全部测试套件
  - lint 检查
  - CONVENTION_CHECK=1: 最终规范合规检查
  - 失败: 进入修复循环（最多3次）

步骤10: 代码审查
  触发条件:
    - TURBO 模式: 强制执行（不限复杂度）
    - 其他模式: complex/architect + 核心/安全模块时执行
  - [RLM:reviewer] 审查代码质量和安全性
  - 发现问题: 回到步骤6 修复

步骤11: 知识库同步（KB_SKIPPED=false 时）
  - [RLM:kb_keeper] 通过 KnowledgeService 更新:
    - CHANGELOG.md
    - 涉及模块的 modules/{module}.md
    - GRAPH_MODE=1: 更新知识图谱节点和边
  - 更新 tasks.md 状态

步骤12: 方案包归档
  - [RLM:pkg_keeper] 更新方案包状态
  - 移动到 archive/ (如适用)

步骤13: 完成
  - 输出完成摘要:
    - 完成的任务列表
    - 修改的文件列表
    - 测试结果
    - 知识库更新内容
  - CURRENT_STAGE = COMPLETE
```

## 失败处理

```yaml
任务失败:
  - 记录错误到 error_log
  - attempts += 1
  - attempts < max_attempts: 重试修复
    - max_attempts: TURBO=5 | 其他模式=3
  - attempts >= max_attempts: 标记 [!], 继续下一个不依赖此任务的任务
  - 所有后续任务都依赖失败任务: 报告阻塞 → ⛔ END_TURN

全量验证失败:
  - 分析失败原因
  - 修复（最多3次循环）
  - 仍失败: 报告问题 → ⛔ END_TURN
```

## TURBO 模式额外保障

```yaml
TURBO 完成后额外流程:
  步骤A: 完整 review
    - 强制执行 [RLM:reviewer] 全量审查
    - 覆盖: 安全性、正确性、代码质量、测试覆盖

  步骤B: 完整 test
    - 运行全部测试套件
    - lint / compile 检查

  步骤C: 修复循环（若步骤A/B发现问题）
    - 自动修复问题
    - 重新执行步骤A + B
    - 最多3轮修复循环

  步骤D: EHRB 风险报告
    - 输出所有被忽略的 EHRB 风险清单
    - 标记风险级别和建议措施

最终输出格式:
  💡【AgentFlow】- 持续执行 ✅: 全部任务完成

  ## 📊 完成报告

  ### ✅ 完成任务
  {按 tasks.md 列出所有完成的任务}

  ### 🔧 解决的问题
  {列出开发和 review 过程中发现并修复的问题}

  ### 🧪 测试结果
  {列出运行了哪些测试、通过/失败状态}

  ### 🔍 审查记录
  {review 轮次、发现的问题、修复情况}

  ### ⚠️ 风险规避（EHRB 记录）
  {列出被忽略但已记录的 EHRB 风险及建议措施，无则显示"无"}

  ### 📝 修改文件清单
  {列出所有修改/新增/删除的文件}

  🔄 下一步: {后续建议}
```
