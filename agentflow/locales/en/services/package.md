# plan package服务 (PackageService)

> manageplan package的create、存档and执行。

## plan package结构

```
{KB_ROOT}/plan/YYYYMMDDHHMM_<feature>/
├── proposal.md    # 方案文档（technology selection、架构design、interface definition）
└── tasks.md       # task checklist（按dependencies排序）
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
create (DESIGN Phase 3):
  1. generate proposal.md（方案文档）
  2. generate tasks.md（task checklist）
  3. save到 plan/ 目录

执行 (~exec):
  1. 列出可用plan package
  2. user selects择
  3. load tasks.md → 进入 DEVELOP

archive (DEVELOP 完成after):
  1. update tasks.md 最终status
  2. 记录完成when间
  3. move to archive/（如用户指定）

statusupdate:
  - 每个tasks完成after立即update tasks.md
  - symbols: ⬜→🔄→✅ or ❌
```
