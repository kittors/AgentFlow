# interactive-ui

## 范围

- `internal/ui/**`
- `internal/app/app.go` 中的交互式主菜单流程

## 当前实现

- 选择器由高卡片堆叠布局改为紧凑列表 + 详情面板，减轻小终端高度压力
- 主菜单、Profile、安装目标、卸载目标、busy/loading 与结果提示统一收敛到 `internal/ui/interactive_flow.go` 的单会话状态机
- 主菜单详情面板可直接显示：
  - 当前动作说明
  - 环境状态（可执行文件路径、CLI 状态、更新提示）
  - 最近一次交互式动作结果
- 启动 Bubble Tea 时启用 `WithMouseCellMotion()`，滚轮上/下被映射为菜单游标切换
- `internal/app` 为交互式 `install` / `uninstall` / `update` / `clean` 提供 panel 化结果，避免信息泄漏到 TUI 外部终端缓冲区
- 交互式动作执行时会在当前 TUI 内显示 busy panel，而不是退出后打印普通终端输出
- `Esc` 回退规则固定为：
  - 主菜单 `Esc` 退出
  - `install targets` `Esc` 返回 `profile`
  - `profile` `Esc` 返回主菜单
  - `uninstall targets` `Esc` 返回主菜单
- UI 测试覆盖滚轮切换、受限高度裁剪、面板内容内嵌渲染，以及 `update` / `clean` / `uninstall` 的原位执行与回退层级

## 已知边界

- 极小终端下详情面板仍可能显示省略号，但不会再把内容顶出 TUI 外部
- 交互式更新只反馈“已替换二进制并提示重启”，当前进程不会热切换版本
- 首次语言选择仍是独立的 Bubble Tea 会话；进入主菜单后，其余动作已保持在单一会话内
