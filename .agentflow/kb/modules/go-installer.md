# go-installer

## 范围

- `embed.go`
- `internal/config/**`
- `internal/install/**`

## 当前实现

- 根包通过 `embed.FS` 暴露 `AGENTS.md`、`SKILL.md` 与 `agentflow/` 目录资源
- `internal/config` 提供 marker、target 定义、备份与 rename-aside 命名规则
- `internal/install` 提供 `Installer` / `TargetInstaller`、安全写入、备份、目录复制与最小 target 安装流程
- GitHub Release 现已切换为 `main` 自动刷新 `continuous` release，安装脚本默认跟随最新 `main` 构建
- `internal/update` 使用 release `name` 解析版本号，避免固定 tag 影响版本比较

## 未完成

- profile 组装与 CLI 特定内容替换
- hooks、agent roles、uninstall/install-all
- 更细致的错误报告与用户输出
