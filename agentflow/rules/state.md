# 状态管理规则

```yaml
状态变量作用域:
  全局: ROUTE_LEVEL, WORKFLOW_MODE, KB_SKIPPED, GRAPH_MODE
  阶段级: CURRENT_STAGE, TASK_COMPLEXITY
  任务级: 每个 Task 的 status, attempts

状态转换规则:
  CURRENT_STAGE:
    null → DESIGN: R2/R3/R4 确认后
    null → DEVELOP: ~exec 直接执行
    DESIGN → DEVELOP: 方案确认后
    DEVELOP → COMPLETE: 所有任务完成
    任意 → null: 用户取消/重置

  ROUTE_LEVEL 升级:
    R1 → R2: 发现超出预期/EHRB
    R2 → R3: 架构级影响/跨模块
    R3 → R4: 系统级重构
    降级: 不允许

状态重置:
  触发: 用户说"停止/取消/重置/中断"
  操作: 清空所有状态变量，返回空闲状态
  保留: L2 会话摘要保存当前进度
```
