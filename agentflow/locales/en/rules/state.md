# state management规then

```yaml
state variables作用域:
  全局: ROUTE_LEVEL, WORKFLOW_MODE, KB_SKIPPED, GRAPH_MODE
  阶段级: CURRENT_STAGE, TASK_COMPLEXITY
  tasks级: 每个 Task 的 status, attempts

state transitionsRules:
  CURRENT_STAGE:
    null → DESIGN: R2/R3/R4 确认after
    null → DEVELOP: ~exec 直接执行
    DESIGN → DEVELOP: 方案确认after
    DEVELOP → COMPLETE: 所有tasks完成
    任意 → null: 用户取消/重置

  ROUTE_LEVEL 升级:
    R1 → R2: 发现超出预期/EHRB
    R2 → R3: 架构级影响/跨模块
    R3 → R4: 系统级重构
    Degradation: 不允许

status重置:
  Trigger: 用户说"停止/取消/重置/during断"
  操作: 清空所有state variables，返回空闲status
  保留: L2 session summarysavewhenbefore进度
```
