# Knowledge Changelog

## 2026-03-13

- 修复 continuous release 曾被错误保留为 draft 的发布链路，workflow 现在在创建 release 后通过 GitHub REST API 显式 patch 为非 draft 并标记 latest
- `install.sh` 与 `internal/update` 现在优先读取 `releases/tags/continuous`，失败后才回退 `releases/latest`，避免公开 latest API 暂时落到旧稳定版
- 更新缓存现在会忽略畸形版本值 `continuous` / `unknown`，避免升级后仍误报“有新版本”
- 排查确认用户终端可能仍先命中 `~/.local/bin/agentflow` 的旧 uv 版本；Go 二进制实际安装路径仍是 `~/.agentflow/bin/agentflow`

## 2026-03-12

- 建立 Go 版资源内嵌入口 `embed.go`
- 在 `internal/config` 增加 marker、target、备份/rename-aside 命名辅助
- 在 `internal/install` 增加安装器骨架、安全写入、备份与目录复制逻辑
- 新增基础测试，`go test ./...` 通过
- Release 工作流改为 `main` push 自动发布 continuous release，并阻止旧 run rerun 后覆盖最新 `continuous`
- `internal/update` 现在优先读取 release `name` 作为版本号，兼容固定 `continuous` tag
- README / README_CN 已补充“`main` 自动 continuous release”说明
- 交互式 TUI 主菜单改为“动作列表 + 内嵌状态/结果面板”，`status` 等信息不再打印到 TUI 外部
- 选择器现在会按终端高度裁剪显示，并启用鼠标滚轮事件接管，避免界面溢出和滚轮触发终端外层滚动
- 交互主流程改为单一 Bubble Tea 状态机，`更新` / `状态` / `清理` / `卸载` 不再退出后重进 TUI
- `install -> profile -> targets` 与 `uninstall -> targets` 的 `Esc` 回退层级被固定为单次返回一级
- 为交互状态机补充 `update` / `clean` / `uninstall` / busy panel 行为测试，`go test ./...` 持续通过
- 移除交互式菜单中未公开的 `j/k/g/G/q` 字母快捷键，普通字母输入不再意外触发导航或退出
