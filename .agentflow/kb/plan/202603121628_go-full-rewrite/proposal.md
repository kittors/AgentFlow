# Go 全量重写方案

## 方案信息

- 方案 ID: `202603121628_go-full-rewrite`
- 创建时间: `2026-03-12 16:28:36`
- 当前基线分支: `main`
- 建议开发分支: `rewrite/go-full-rebuild`
- 执行模式: `DELEGATED_PLAN`

## 背景与目标

当前项目以 Python 为主，CLI/TUI、安装器、知识库同步、版本检查、架构扫描、会话管理等能力都运行在 Python 运行时之上。你希望将整个项目完整改写为 Go，以获得以下收益：

1. 统一 Windows 与 macOS 下的 TUI 行为，减少终端差异带来的体验波动。
2. 提供单一可执行文件，降低 Python 解释器、pip/uv、PyInstaller 相关分发负担。
3. 建立更直接的跨平台发布链路，让安装、更新和离线分发更可控。

本方案目标不是“局部替换”，而是完整迁移到 Go，并在验证通过后合并回 `main`。

## 当前项目事实基线

### 用户入口与分发

- Python 包入口为 `agentflow.cli:main`
- 交互式菜单位于 `agentflow/cli.py` 与 `agentflow/interactive.py`
- 当前依赖 `rich` 与 `InquirerPy` 提供 TUI
- 当前发布路径同时覆盖 `pip`、`uv`、`npx` 桥接、Shell/PowerShell 安装脚本
- 仓库已经包含二进制下载兜底逻辑，说明“独立可执行文件”本身就是当前产品方向的一部分

### 核心模块职责

- `agentflow/cli.py`: 命令分发、主菜单、命令行参数入口
- `agentflow/installer.py`: 安装、卸载、部署规则、Hook、技能与多代理配置
- `agentflow/updater.py` + `agentflow/version_check.py`: 更新、状态、缓存、版本检查
- `agentflow/_constants.py`: CLI 目标、路径、语言、资源定位、通用工具
- `agentflow/scripts/*`: 架构扫描、规范提取、图谱构建、KB 同步、会话管理、模板初始化
- `agentflow/functions/*`、`agentflow/stages/*`、`agentflow/templates/*`: AgentFlow 工作流资产，部署时复制到目标 CLI 配置目录

### 现有测试主题

仓库已有较完整的 pytest 覆盖面，重点覆盖：

- CLI 命令与交互行为
- 安装器与配置写入
- profile 组装
- KB / graph / convention / session / dashboard 脚本
- 更新、缓存、静态配置与模板文件

这批测试主题应该转译为 Go 的行为契约，而不是简单删除。

## 迁移范围

### 纳入重写范围

1. 所有 Python 可执行逻辑
2. 所有脚本型能力：KB、graph、conventions、session、scan、dashboard、version、update
3. 安装/卸载/更新流程
4. TUI 与命令解析
5. 发布、打包、跨平台构建流程
6. 测试体系与 CI

### 保留但改承载方式

1. `AGENTS.md`、`SKILL.md`、`agentflow/functions/*`、`agentflow/stages/*`、`agentflow/templates/*` 等规则与模板资产
2. 安装脚本 `install.sh` 与 `install.ps1`
3. `package.json` 的 npx 入口

这些内容建议转为 Go 二进制内嵌资源，通过 `embed` 提供部署与释放。

### 不在本次计划中执行

1. 当前回合不创建开发分支
2. 当前回合不修改源码
3. 当前回合不执行测试
4. 当前回合不合并到 `main`

## 约束与风险

### 关键约束

1. 命令面必须保持兼容，至少覆盖当前 `agentflow`, `install`, `uninstall`, `update`, `status`, `clean`, `version`
2. 规则文件、模板文件、skills、hooks 的输出内容与目录结构必须兼容现有目标 CLI
3. 安装器必须继续支持 macOS、Linux、Windows
4. 合并到 `main` 前必须通过自动化测试与跨平台 smoke test

### EHRB 风险记录

- 涉及 `main` 分支最终合并，属于低到中风险变更
- 属于技术栈整体迁移，存在行为回归风险
- 安装器、配置文件写入和跨平台路径差异属于高概率缺陷来源

## 方案对比

### 方案 A: 单仓 Go Monolith + 资源内嵌

核心思路：

- 使用 Go 重建所有运行时代码
- 将规则、模板、技能、hooks 作为静态资源使用 `embed` 打包进单一二进制
- 使用 Go 原生子包划分能力边界
- 通过 GoReleaser 或等效构建矩阵输出多平台二进制

优点：

- 分发模型最简单，目标最符合你的初衷
- Windows/macOS/Linux 体验更容易收敛
- 运行时依赖最少
- 安装脚本可以只负责拉取二进制

缺点：

- 首次重写成本最高
- 需要补齐 TOML/JSON/Markdown 资源装载与配置写回能力

### 方案 B: Go 外壳 + 保留部分 Python 脚本

核心思路：

- 只把 CLI/TUI 与安装器迁到 Go
- graph、KB、session、convention 等脚本暂时保留 Python

优点：

- 初期工作量更低
- 可以更快得到 Go 的 TUI 和分发收益

缺点：

- 无法实现“完整改成 Go”
- 最终仍需处理 Python 依赖与运行时
- 发布模型会变成混合系统，复杂度更高

## 推荐方案

推荐采用方案 A。

原因：

1. 你的要求是“整个项目完整改成 Go，全量重写”，方案 B 不满足目标。
2. 当前仓库体量适中，核心逻辑集中在 CLI、安装器与一组辅助脚本，属于适合一次性迁移的规模。
3. 既然已经明确把“跨平台一致性”和“单可执行文件”作为主目标，就不应保留 Python 运行时尾巴。

## 推荐目标架构

```text
cmd/agentflow/
  main.go

internal/app/
  command.go
  dispatch.go

internal/ui/
  menu.go
  prompt.go
  render.go

internal/i18n/
  locale.go
  messages.go

internal/assets/
  embed.go

internal/targets/
  targets.go
  profiles.go

internal/install/
  install.go
  uninstall.go
  deploy.go
  codex.go
  claude.go

internal/config/
  toml.go
  json.go
  files.go

internal/update/
  version.go
  update.go
  cache.go

internal/kb/
  init.go
  sync.go
  session.go
  template.go

internal/scan/
  arch.go
  convention.go
  graph.go
  dashboard.go

internal/gitflow/
  branch.go
  merge.go

test/
  e2e/
  golden/
```

## 技术选型建议

### 命令与 TUI

- 命令解析: `cobra` 或标准库 `flag` + 自定义 dispatcher
- TUI: `bubbletea` + `bubbles` + `lipgloss`

建议：

- 若追求一致交互与后续扩展，优先 `bubbletea` 生态
- 若希望依赖更少，可保留命令解析层简单化，但 TUI 仍建议使用 `bubbletea`

### 资源装载

- 使用 `embed` 打包 `AGENTS.md`、`SKILL.md`、`agentflow/functions/*`、`agentflow/stages/*`、`agentflow/templates/*`、`agentflow/hooks/*`

### 配置读写

- TOML: `BurntSushi/toml` 或 `pelletier/go-toml/v2`
- JSON: Go 标准库
- 文件备份、路径探测、权限处理全部收敛到 `internal/config`

### 发布

- 使用 GitHub Actions 构建矩阵
- 输出目标至少包括：
  - `darwin-amd64`
  - `darwin-arm64`
  - `linux-amd64`
  - `linux-arm64`
  - `windows-amd64`

## 实施阶段

### Phase 0: 基线冻结

- 记录 Python 当前命令面、安装路径、产物结构、测试主题
- 明确哪些行为必须保持兼容

### Phase 1: Go 骨架与资源层

- 建立 Go module
- 建立 assets/embed 机制
- 建立 CLI target/profile 常量层

### Phase 2: CLI/TUI 与安装器

- 实现主命令
- 实现安装/卸载/状态/清理/版本/更新
- 实现 Codex/Claude 等目标的规则部署

### Phase 3: AgentFlow 辅助能力迁移

- KB sync
- session 管理
- convention 扫描
- graph 构建
- arch scan
- dashboard 生成

### Phase 4: 测试与发布链路

- 将 pytest 覆盖主题转写为 `go test`
- 建立 Golden Test 与跨平台 smoke test
- 调整 shell / powershell 安装器与 npx 桥接

### Phase 5: 切换与合并

- 在 `rewrite/go-full-rebuild` 分支完成联调
- 通过 CI 与人工 smoke
- 提 PR 合并回 `main`

## 完成定义

满足以下条件才允许合并到 `main`：

1. 无 Python 运行时代码作为正式发布依赖
2. Go 二进制可以在 macOS、Windows、Linux 启动
3. 现有核心命令行为完成迁移
4. 工作流静态资产部署正确
5. 自动化测试通过
6. 安装、更新、卸载、状态检查完成 smoke test
7. README / README_CN / CHANGELOG / 贡献文档同步完成

## 里程碑验收建议

- 里程碑 M1: Go CLI 可以替代当前 Python 主入口
- 里程碑 M2: 安装器与部署逻辑完成，能正确写入目标 CLI 配置
- 里程碑 M3: KB / graph / conventions / sessions 全迁移
- 里程碑 M4: 跨平台构建与测试矩阵通过
- 里程碑 M5: 分支验证通过并准备合并 `main`

## 执行建议

执行此方案时，建议走 `~exec 202603121628_go-full-rewrite`，并按下列原则推进：

1. 先建 Go 骨架和资源内嵌，不要一开始就平移所有细节实现
2. 先保命令兼容和安装器，再迁脚本性能力
3. 每完成一个大模块，就补对应的 Go 测试
4. Python 代码在最终切换前仅作为比对基线，不要长期维持双实现
