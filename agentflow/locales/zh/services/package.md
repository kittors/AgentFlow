# 方案包服务 (PackageService)

> 管理方案包的创建、存档和执行。

## 方案包结构

```
{KB_ROOT}/plan/YYYYMMDDHHMM_<feature>/
├── proposal.md    # 方案文档（技术选型、架构设计、接口定义）
└── tasks.md       # 任务清单（按依赖排序）
```

## tasks.md 格式

```markdown
# Task Plan: {feature_name}

## Approach
{approach_description}

## Tasks

### Task 1: {title}
- **Description**: {what to do}
- **Files**: {file_list}
- **Depends on**: {dependency_ids}
- **Verification**: {how to verify}
- **Status**: ⬜

### Task 2: {title}
...
```

## 操作协议

```yaml
创建 (DESIGN Phase 3):
  1. 生成 proposal.md（方案文档）
  2. 生成 tasks.md（任务清单）
  3. 保存到 plan/ 目录

执行 (~exec):
  1. 列出可用方案包
  2. 用户选择
  3. 加载 tasks.md → 进入 DEVELOP

归档 (DEVELOP 完成后):
  1. 更新 tasks.md 最终状态
  2. 记录完成时间
  3. 移动到 archive/（如用户指定）

状态更新:
  - 每个任务完成后立即更新 tasks.md
  - 符号: ⬜→🔄→✅ 或 ❌
```
