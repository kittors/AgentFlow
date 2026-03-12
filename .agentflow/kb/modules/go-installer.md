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
- Release metadata 阶段会校验当前 run 的 `GITHUB_SHA` 是否仍为 `origin/main` 头部，避免 rerun 旧任务把 continuous 回滚到旧提交
- `internal/update` 使用 release `name` 解析版本号，避免固定 tag 影响版本比较
- Release workflow 在创建 `continuous` 后会通过 `gh release edit --draft=false --latest` 显式发布，并追加公开校验步骤，避免 workflow 通过但 release 仍停留在 draft
- 安装脚本解析下载地址时会优先读取 `releases/tags/continuous`，仅在该接口不可用时回退到 `releases/latest`
- Go 自更新同样优先读取 `continuous` release，并在读取缓存时跳过 `continuous` 这类不可比较的畸形版本值
- Go 自更新默认 HTTP 超时已提高到 2 分钟，并有慢下载测试覆盖，避免较慢网络下更新 9-13MB 二进制时被 5 秒超时过早中断
- 当当前版本与缓存版本都是 `-main.` 构建但提交不同，版本检查会主动回源刷新，不再把旧缓存误报成“新版本可用”
- 主菜单交互结果不再直接写到终端普通缓冲区，而是通过 TUI 内部状态/结果面板展示
- `install.sh` 会自动备份并重定向用户目录中的旧 `agentflow` 命令入口，例如 `~/.local/bin/agentflow`，避免安装 Go 版后仍被历史 uv/Python 入口抢先命中

## 运维提示

- 如果安装完成后直接执行 `agentflow` 仍看到旧 UI 或旧版本，先检查 `type -a agentflow`
- 安装脚本会优先尝试自动接管用户目录中的旧入口；如果 shell 会话仍缓存了旧 PATH，执行 `export PATH="$HOME/.agentflow/bin:$PATH" && hash -r`，或重新加载 shell 配置后再试

## 未完成

- profile 组装与 CLI 特定内容替换
- hooks、agent roles、uninstall/install-all
- 更细致的错误报告与用户输出
