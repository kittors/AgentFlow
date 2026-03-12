# 任务清单

- [x] T1: 建立迁移基线与行为清单 | 文件: `README.md`, `README_CN.md`, `agentflow/`, `tests/` | deps: 无 | 验收: 输出一份 Python→Go 的功能映射表，覆盖现有命令、安装器、脚本能力和测试主题
- [x] T2: 创建重写分支 `rewrite/go-full-rebuild` | 文件: `.git` | deps: [T1] | 验收: 新分支从 `main` 切出且工作树干净，后续开发全部在该分支进行
- [x] T3: 建立 Go module、目录骨架与构建入口 | 文件: `go.mod`, `cmd/agentflow/main.go`, `internal/` | deps: [T1, T2] | 验收: `go test ./...` 可运行，`go build ./cmd/agentflow` 成功生成可执行文件
- [x] T4: 实现静态资源内嵌层 | 文件: `embed.go`, `AGENTS.md`, `SKILL.md`, `agentflow/functions/`, `agentflow/stages/`, `agentflow/templates/`, `agentflow/hooks/` | deps: [T3] | 验收: 二进制可读取并释放规则、模板、hooks、skills 资源，输出目录结构与当前项目兼容
- [x] T5: 迁移常量、locale、profile 与 CLI target 定义 | 文件: `internal/i18n/`, `internal/targets/`, `internal/config/` | deps: [T3] | 验收: Go 版本可以正确识别语言、profile、目标 CLI 路径与资源定位
- [x] T6: 重写命令分发与非交互入口 | 文件: `internal/app/`, `cmd/agentflow/main.go` | deps: [T4, T5] | 验收: `agentflow`, `agentflow install`, `agentflow uninstall`, `agentflow update`, `agentflow status`, `agentflow clean`, `agentflow version` 命令均可解析
- [x] T7: 重写跨平台 TUI 菜单 | 文件: `internal/ui/` | deps: [T6] | 验收: macOS 与 Windows 下主菜单交互一致，键盘导航、退出、回退行为稳定
- [x] T8: 重写安装/卸载/部署逻辑 | 文件: `internal/install/`, `internal/config/` | deps: [T4, T5, T6] | 验收: 可将 AgentFlow 部署到 Codex、Claude 等目标 CLI，保留备份、覆盖和多代理配置能力
- [x] T9: 重写版本检查、更新、状态与缓存管理 | 文件: `internal/update/` | deps: [T6] | 验收: 版本缓存、GitHub release 检查、状态输出与清理逻辑和现有行为一致
- [x] T10: 重写 KB / session / template 能力 | 文件: `internal/kb/` | deps: [T4, T5, T6] | 验收: `.agentflow/kb/`、`.agentflow/sessions/` 的生成、同步和会话摘要保存全部由 Go 完成
- [x] T11: 重写架构扫描、规范提取、图谱与 dashboard 能力 | 文件: `internal/scan/` | deps: [T10] | 验收: Go 版本可以产出与现有 Python 脚本同类结果，至少覆盖 large files、missing tests、graph、conventions、dashboard
- [x] T12: 调整安装脚本与 npx 桥接到 Go 二进制分发 | 文件: `install.sh`, `install.ps1`, `package.json`, `bin/` | deps: [T8, T9] | 验收: Shell、PowerShell、npx 路径都以下载/调用 Go 二进制为主，不再依赖 Python 运行时
- [x] T13: 将 pytest 行为契约迁移为 Go 测试 | 文件: `tests/`, `test/`, `internal/**/_test.go` | deps: [T6, T8, T9, T10, T11] | 验收: 单元测试、集成测试、golden test 覆盖 CLI、installer、profile、KB、graph、update 等关键模块
- [x] T14: 建立跨平台 CI 与发布矩阵 | 文件: `.github/workflows/`, `goreleaser.*` 或等效配置 | deps: [T12, T13] | 验收: 至少在 macOS、Linux、Windows 上构建并运行 smoke test，产出 release artifact
- [x] T15: 完成文档与迁移说明同步 | 文件: `README.md`, `README_CN.md`, `CHANGELOG.md`, `CONTRIBUTING.md` | deps: [T12, T13] | 验收: 文档改为 Go 实现说明，安装、开发、测试、发布流程全部更新
- [x] T16: 执行全量验证并准备合并 `main` | 文件: `.github/workflows/`, `docs/`, `test/` | deps: [T14, T15] | 验收: 分支通过自动化测试与人工 smoke test，形成 PR，明确合并到 `main` 的回滚与发布说明
