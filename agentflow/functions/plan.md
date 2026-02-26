# ~plan 命令

> 规划模式 — 仅完成方案设计，不进入开发阶段。

```yaml
触发: 用户输入 ~plan [子命令|需求描述]
子命令:
  ~plan <需求描述>: 对需求走 R3 DESIGN 流程，完成后停止
  ~plan list: 列出所有已有方案包及其完成进度
  ~plan show [id]: 查看指定方案包的详情和 tasks.md 状态
  ~plan (无参数): 等同于 ~plan list
```

## ~plan <需求描述>

```yaml
流程:
  WORKFLOW_MODE = DELEGATED_PLAN
  1. 按 R3 需求评估流程评估需求
  2. 进入 DESIGN 阶段（加载 stages/design.md）
  3. 生成 tasks.md（checklist 格式，见下方）
  4. 保存方案包到 .agentflow/kb/plan/YYYYMMDDHHMM_<feature>/
  5. ⛔ END_TURN — 不进入 DEVELOP

输出:
  💡【AgentFlow】- DESIGN ✅: 方案规划完成
  📋 方案包: .agentflow/kb/plan/{id}/
  📄 proposal.md — 方案详情
  📄 tasks.md — 任务清单（{total}项任务）
  🔄 下一步: ~exec 执行此方案 | ~plan show {id} 查看详情
```

## ~plan list

```yaml
输出:
  💡【AgentFlow】- 方案列表

  | # | 方案包 | 创建时间 | 进度 | 状态 |
  |---|--------|----------|------|------|
  | 1 | {id}_{feature} | {time} | {done}/{total} | ⬜/🔄/✅ |

  空列表时:
    📭 暂无方案包。
    💡 使用 ~plan <需求> 创建新方案，或直接描述需求。

  🔄 下一步: ~plan show {id} 查看详情 | ~exec 执行方案
```

## ~plan show [id]

```yaml
输出:
  💡【AgentFlow】- 方案详情: {feature}

  📄 方案摘要: {proposal 概述}
  📋 任务清单:
    [x] T1: {已完成任务}
    [/] T2: {进行中任务}
    [ ] T3: {待执行任务}
    [!] T4: {失败任务}

  进度: {done}/{total} ({percentage}%)
  🔄 下一步: ~exec 继续执行 | ~exec {id} 执行此方案
```

## tasks.md Checklist 格式（CRITICAL）

```yaml
格式规范:
  文件位置: .agentflow/kb/plan/{id}/tasks.md
  符号:
    "[ ]": 待执行
    "[/]": 执行中
    "[x]": 已完成
    "[!]": 失败

  每项任务必须包含:
    - id (T1, T2, ...)
    - 标题
    - 涉及文件
    - 依赖 (deps: [T1] 或 无)
    - 验收标准

  示例:
    ## 任务清单

    - [x] T1: 创建数据模型 | 文件: src/models.py | deps: 无
    - [/] T2: 实现 API 端点 | 文件: src/api.py | deps: [T1]
    - [ ] T3: 编写测试 | 文件: tests/test_api.py | deps: [T2]
    - [!] T4: 配置 CI/CD | 文件: .github/workflows/ | deps: 无

更新规则（CRITICAL）:
  - 任务开始执行 → 标记 [/]
  - 任务通过验收 → 标记 [x]
  - 任务失败且超过重试上限 → 标记 [!]
  - 每次标记变更后立即写入 tasks.md 文件
  - DEVELOP 阶段的步骤 5e 必须更新 tasks.md
```
