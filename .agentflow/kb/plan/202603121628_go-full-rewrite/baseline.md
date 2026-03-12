# Python 到 Go 迁移基线

## 范围说明

本文件冻结 2026-03-12 当前 Python 实现的公开行为，用作 Go 全量重写的兼容对照基线。

## 入口与分发

| 维度 | Python 当前实现 | Go 迁移目标 |
|---|---|---|
| 主入口 | `pyproject.toml` 中 `agentflow = "agentflow.cli:main"` | `cmd/agentflow/main.go` |
| 打包入口 | `pyinstaller_entry.py` 转发到 `agentflow.cli:main` | 统一为 Go 二进制，不再需要 Python 打包入口 |
| 交互模式 | 有 TTY 且依赖可用时进入 Rich + InquirerPy TUI | 统一迁到 Go TUI |
| 降级模式 | 无 TTY 或依赖缺失时回退纯文本菜单 | 保留降级语义 |
| 分发路径 | `pip` / `uv` / `npx` / shell / powershell / release binary | 收敛到 Go release binary，保留 shell、powershell、npx 桥接 |

## 命令面冻结

| 命令 | Python 当前行为 | Go 必须兼容 |
|---|---|---|
| `agentflow` | 无参数进入交互菜单 | 保持 |
| `agentflow install` | 支持目标 CLI、`--all`、profile | 保持 |
| `agentflow uninstall` | 支持指定目标或全部卸载 | 保持 |
| `agentflow update` | 根据 uv/pip 更新并重部署 | 保持 |
| `agentflow status` | 输出版本、安装方式、各 CLI 状态 | 保持 |
| `agentflow clean` | 清理缓存目录 | 保持 |
| `agentflow version` | 输出版本 | 保持 |
| `agentflow --check-update --silent` | 走缓存检查更新 | 保持 |

## CLI/TUI 行为契约

| 主题 | Python 当前行为 | Go 必须兼容 |
|---|---|---|
| 帮助与未知命令 | 帮助返回成功，未知命令返回失败 | 保持退出码语义 |
| 菜单返回 | 执行操作后可返回主菜单或退出 | 保持 |
| 语言 | `AGENTFLOW_LANG` 优先，其次系统 locale | 保持 |
| TTY 检测 | 决定是否启用 TUI、多代理提示等交互 | 保持 |
| 非交互 | 非 TTY 环境下不应阻塞等待用户输入 | 保持 |

## 安装器行为映射

| 主题 | Python 当前实现 | Go 迁移目标 |
|---|---|---|
| 目标矩阵 | `codex` / `claude` / `gemini` / `qwen` / `grok` / `opencode` | 保持 |
| 规则部署 | 写入各 CLI 的规则文件 | 保持 |
| 模块部署 | 复制 `agentflow/` 模块目录 | 改为从嵌入资源释放 |
| Skill 部署 | 复制 `SKILL.md` 到目标 CLI skill 目录 | 保持 |
| Hooks 部署 | 按目标写入 hooks | 保持 |
| 备份策略 | 非 AgentFlow 文件先备份再覆盖 | 保持 |
| 标记识别 | 通过 `AGENTFLOW_ROUTER:` 判断文件归属 | 保持 |
| 原子写入 | 写临时文件再替换目标文件 | 保持 |
| Windows 锁文件 | 删除失败时 rename-aside | 保持 |
| Codex 多代理 | 配置 `config.toml` 与 `agents/*.toml` | 保持 |

## 更新与状态行为映射

| 主题 | Python 当前实现 | Go 迁移目标 |
|---|---|---|
| 安装方式探测 | 优先识别 `uv tool list`，否则视为 `pip` | Go 二进制分发后重定义安装来源，但保留用户可见语义 |
| 版本检查 | GitHub release API + 72 小时缓存 | 保持 |
| 版本格式 | 去掉 tag 前缀 `v` | 保持 |
| 更新后动作 | 成功更新后自动重部署已安装目标 | 保持 |
| 状态 | 版本、安装方式、每个 CLI 的安装状态 | 保持 |
| 清理 | 删除已安装目标下 `agentflow/__pycache__` 等缓存 | 改为 Go 版本缓存目录，但保留 `clean` 命令含义 |

## 规则/模板资产迁移

| 资产 | 当前承载方式 | Go 迁移方式 |
|---|---|---|
| `AGENTS.md` | 仓库根文件 + 安装时复制 | `embed` + 释放 |
| `SKILL.md` | 仓库根文件 + 安装时复制 | `embed` + 释放 |
| `agentflow/functions/*` | Markdown 资产 | `embed` + 释放 |
| `agentflow/stages/*` | Markdown 资产 | `embed` + 释放 |
| `agentflow/templates/*` | Markdown 模板 | `embed` + 释放 |
| `agentflow/hooks/*` | hook 配置模板 | `embed` + 释放 |

## 辅助脚本能力映射

| 模块 | Python 当前职责 | Go 迁移包建议 | 兼容要求 |
|---|---|---|---|
| `template_init.py` | 初始化 `.agentflow/kb/` 骨架与模板引用 | `internal/kb/template.go` | 目录结构与文件命名兼容 |
| `kb_sync.py` | 扫描模块并生成 `modules/*.md` | `internal/kb/sync.go` | 保持 `_index.md` 与模块文档结构 |
| `session_manager.py` | 生成、列出、裁剪、导出会话摘要 | `internal/kb/session.go` | 保持 `.agentflow/sessions/*.md` 命名与区段 |
| `convention_scanner.py` | 提取命名、导入、文档风格到 JSON | `internal/scan/convention.go` | 保持 `conventions/extracted.json` 主结构 |
| `graph_builder.py` | 生成 `nodes.json`、`edges.json`、`graph.mmd` | `internal/scan/graph.go` | 保持稳定 node id 与边类型 |
| `arch_scanner.py` | large files、missing tests、circular imports、long functions | `internal/scan/arch.go` | 保持主要扫描结果字段 |
| `dashboard_generator.py` | 生成 HTML 仪表盘 | `internal/scan/dashboard.go` | 保持 dashboard 文件输出 |
| `cache_manager.py` | KB cache 管理 | `internal/update/cache.go` 或 `internal/kb/cache.go` | 保持 cache 基本接口语义 |
| `config_helpers.py` | 读写 TOML/JSON | `internal/config/` | 保持辅助能力 |

## 测试主题冻结

| 领域 | 当前 pytest 覆盖主题 | Go 测试迁移方向 |
|---|---|---|
| CLI | 帮助、命令分发、无参数交互、未知命令 | `internal/app` + `cmd` 单元/集成测试 |
| Locale | `detect_locale()`、`msg()` | `internal/i18n` 单元测试 |
| Installer | 目标探测、文件备份、部署、卸载、Codex/Claude 特例 | `internal/install` 单元/集成测试 |
| Profile | `lite/standard/full` 组装 | `internal/targets` 单元测试 |
| Update | cache TTL、版本检查、状态、清理 | `internal/update` 单元测试 |
| KB | init、sync、session | `internal/kb` 单元测试 |
| Scan | graph、convention、arch、dashboard | `internal/scan` 单元测试 |
| 静态资产 | Markdown、JSON、TOML 配置完整性 | golden test + embed 资源测试 |

## 首批开发优先级

1. `go.mod` 与 `cmd/agentflow/main.go`
2. `internal/assets`、`internal/i18n`、`internal/targets`
3. `internal/app` 命令分发
4. `internal/install` 文件部署主链路
5. `internal/update` 与 `internal/kb`
6. `internal/scan`
7. 测试、CI、文档与发布
