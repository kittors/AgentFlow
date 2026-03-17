# 工具use规then

```yaml
file operations:
  Priority: CLIwithin置工具 > CLIwithin置Shell工具 > 原生Shellcommand
  安全: delete操作beforecheck EHRB
  编码: 始终use UTF-8 none BOM

Shell Command:
  平台检测: check Bash/PowerShell 可用性
  超when: 单个command不超过 60 秒
  error handling: 捕获退出码，非零即报告

Git 操作:
  提交消息: BILINGUAL_COMMIT=1 → 双语格式
  分支: 不在 main/master 上直接操作（EHRB 检测）
  推送: --force must EHRB 确认

outside部工具:
  MCP/插件: 按outside部工具路径执行
  SKILL: 匹配after完全委托给技能处理
  Degradation: 工具不可用when，main agent直接执行

文档/dependencies查询（推荐）:
  - 第三方库 API、配置、版本差异：优先use Context7（若可用）
  - 若none Context7：再use Web 搜索 / 官方文档
```
