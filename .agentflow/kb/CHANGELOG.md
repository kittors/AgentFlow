# Knowledge Changelog

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
