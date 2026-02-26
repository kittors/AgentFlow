# 工具使用规则

```yaml
文件操作:
  优先级: CLI内置工具 > CLI内置Shell工具 > 原生Shell命令
  安全: 删除操作前检查 EHRB
  编码: 始终使用 UTF-8 无 BOM

Shell 命令:
  平台检测: 检查 Bash/PowerShell 可用性
  超时: 单个命令不超过 60 秒
  错误处理: 捕获退出码，非零即报告

Git 操作:
  提交消息: BILINGUAL_COMMIT=1 → 双语格式
  分支: 不在 main/master 上直接操作（EHRB 检测）
  推送: --force 必须 EHRB 确认

外部工具:
  MCP/插件: 按外部工具路径执行
  SKILL: 匹配后完全委托给技能处理
  降级: 工具不可用时，主代理直接执行
```
