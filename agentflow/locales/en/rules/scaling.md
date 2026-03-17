# scaling rules

```yaml
sub-agent扩展:
  simple: 0 sub-agent（main agent直接执行）
  moderate: 1-2 sub-agent（关键tasks分配）
  complex: 2-4 sub-agent（按模块分配）
  architect: 3-6 sub-agent（按层级and模块分配）

parallel度:
  代码scan: 按模块数量分配，最多 4 个parallel
  approach ideation: R3 ≥ 2 parallel，R4 ≥ 3 parallel
  code changes: 按nonedependenciestasks数分配
  测试: 按测试套件分配

上下文manage:
  每个sub-agent: 仅传递必要上下文（tasks描述 + files involved + 规范）
  结果汇总: main agent统一汇总sub-agent结果
  conflict resolution: sub-agent结果冲突when，main agent裁决
```
