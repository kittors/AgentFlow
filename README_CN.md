<div align="center">

```
     █████╗  ██████╗ ███████╗███╗   ██╗████████╗███████╗██╗      ██████╗ ██╗    ██╗
    ██╔══██╗██╔════╝ ██╔════╝████╗  ██║╚══██╔══╝██╔════╝██║     ██╔═══██╗██║    ██║
    ███████║██║  ███╗█████╗  ██╔██╗ ██║   ██║   █████╗  ██║     ██║   ██║██║ █╗ ██║
    ██╔══██║██║   ██║██╔══╝  ██║╚██╗██║   ██║   ██╔══╝  ██║     ██║   ██║██║███╗██║
    ██║  ██║╚██████╔╝███████╗██║ ╚████║   ██║   ██║     ███████╗╚██████╔╝╚███╔███╔╝
    ╚═╝  ╚═╝ ╚═════╝ ╚══════╝╚═╝  ╚═══╝   ╚═╝   ╚═╝     ╚══════╝ ╚═════╝  ╚══╝╚══╝
```

**一个自主的高级智能伙伴，不仅分析问题，更持续工作直到完成实现和验证。**

[English](README.md) · [中文](README_CN.md)

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go 1.26](https://img.shields.io/badge/Go-1.26-00ADD8.svg?logo=go&logoColor=white)](https://go.dev)
[![CI](https://github.com/kittors/AgentFlow/actions/workflows/ci.yml/badge.svg)](https://github.com/kittors/AgentFlow/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/kittors/AgentFlow?include_prereleases&label=Release)](https://github.com/kittors/AgentFlow/releases)

**`五级路由`** · **`EHRB 安全检测`** · **`知识图谱记忆`** · **`子代理编排`** · **`跨平台二进制`**

</div>

---

## 项目概览

AgentFlow 现在以 **Go CLI** 的形式发布，规则、模板、阶段模块、Hooks、角色配置等静态资源都通过 `embed` 打进单个可执行文件中。安装和运行路径不再依赖 Python、`uv`、`pip` 或 PyInstaller。

当前 Go 版本已经覆盖：

- macOS / Linux / Windows 跨平台二进制
- `install` / `uninstall` / `update` / `status` / `clean` 的交互式 TUI
- 面向 Codex、Claude、Gemini、Qwen、Grok、OpenCode 的嵌入式部署资源
- `lite` / `standard` / `full` 三种 profile 规则组装
- Claude hooks 合并与清理
- Codex 多代理配置与角色文件注入
- Go 原生测试、CI、release、KB/session/template、扫描能力

## 安装方式

### 一键安装脚本

macOS / Linux：

```bash
curl -fsSL https://raw.githubusercontent.com/kittors/AgentFlow/main/install.sh | bash
```

Windows PowerShell：

```powershell
irm https://raw.githubusercontent.com/kittors/AgentFlow/main/install.ps1 | iex
```

Windows 常见问题（DNS/代理/raw 域名解析失败等）：见 [docs/troubleshooting/windows.md](docs/troubleshooting/windows.md)。

安装脚本下载的是最新已发布的 release 二进制。现在只要 push 到 `main`，GitHub 就会自动刷新一个 continuous release，所以 `curl | bash`、`npx agentflow` 和 `agentflow update` 都会跟随最新的 `main` 构建。如果你本机还有旧的 `uv` / Python 版 `agentflow` 排在 `PATH` 前面，请重开终端，或者执行 `export PATH="$HOME/.agentflow/bin:$PATH" && hash -r`，然后用 `which agentflow` 确认当前命中的路径。

### `npx` 启动

`npx agentflow` 会自动下载与当前平台匹配的 release 二进制并执行。

```bash
npx agentflow
```

### 手动下载二进制

从 [Releases](https://github.com/kittors/AgentFlow/releases) 下载对应平台文件：

- `agentflow-linux-amd64`
- `agentflow-linux-arm64`
- `agentflow-darwin-amd64`
- `agentflow-darwin-arm64`
- `agentflow-windows-amd64.exe`

然后放到 `PATH` 中，例如：

```bash
chmod +x agentflow-darwin-arm64
mv agentflow-darwin-arm64 ~/.local/bin/agentflow
agentflow version
```

### 本地构建

```bash
git clone https://github.com/kittors/AgentFlow.git
cd AgentFlow
go build -o ./bin/agentflow ./cmd/agentflow
./bin/agentflow version
```

## 快速开始

```bash
agentflow                       # 交互式 TUI
agentflow install codex         # 安装到指定 CLI
agentflow install --all         # 安装到所有检测到的目标
agentflow uninstall codex       # 从指定 CLI 卸载
agentflow uninstall codex --cli # 完整卸载：卸载 CLI 本体并默认删除配置目录（如需保留配置加 --keep-config）
agentflow uninstall --all       # 从所有已安装目标卸载
agentflow status                # 查看安装状态
agentflow clean                 # 清理 AgentFlow 缓存
agentflow update                # 自更新到最新 release 二进制
agentflow version               # 输出当前版本与更新提示
```

当未传子命令且 stdin 是 TTY 时，AgentFlow 会进入 Bubble Tea TUI。方向键、`Enter`、`Space`、`Esc` 在 macOS 和 Windows 终端里保持一致。

## 支持的目标 CLI

| 目标 | 配置目录 | 规则文件 | 额外集成 |
|------|----------|----------|----------|
| Codex CLI | `~/.codex/` | `AGENTS.md` | 注入 `agents/reviewer.toml`、`agents/architect.toml`，并 merge `config.toml` 多代理配置 |
| Claude Code | `~/.claude/` | `CLAUDE.md` | merge / 清理 `settings.json` hooks |
| Gemini CLI | `~/.gemini/` | `GEMINI.md` | 部署规则与嵌入模块 |
| Qwen CLI | `~/.qwen/` | `QWEN.md` | 部署规则与嵌入模块 |
| Grok CLI | `~/.grok/` | `GROK.md` | 部署规则与嵌入模块 |
| OpenCode | `~/.config/opencode/` | `AGENTS.md` | 部署规则与嵌入模块 |

## 核心能力

### 五级路由

每次输入都会按行动需求、目标明确度、决策范围、影响范围、EHRB 风险五个维度打分，路由到 `R0` 到 `R4`。

### EHRB 安全检测

在真正执行之前，拦截破坏性命令、敏感信息、权限越界、生产环境风险以及异常工具输出。

### 知识库与会话摘要

项目状态集中保存在 `.agentflow/`：

- `.agentflow/kb/`
- `.agentflow/kb/plan/`
- `.agentflow/kb/graph/`
- `.agentflow/kb/conventions/`
- `.agentflow/sessions/`

### 嵌入式部署资产

Go 二进制内置：

- `AGENTS.md`
- `SKILL.md`
- `agentflow/stages/`
- `agentflow/functions/`
- `agentflow/services/`
- `agentflow/templates/`
- `agentflow/hooks/`
- `agentflow/agents/`
- `agentflow/core/`

## 仓库结构

```text
AgentFlow/
├── cmd/agentflow/          # CLI 入口
├── internal/app/           # 命令分发与主流程
├── internal/ui/            # Bubble Tea TUI
├── internal/install/       # 部署 / 卸载逻辑
├── internal/update/        # GitHub release 检查与缓存
├── internal/kb/            # KB、session、template
├── internal/scan/          # graph、convention、dashboard、arch scan
├── internal/targets/       # CLI target 与 profile
├── agentflow/              # 随二进制分发的提示词资产
├── embed.go                # 静态资源嵌入
├── install.sh              # POSIX 二进制安装脚本
├── install.ps1             # Windows 二进制安装脚本
└── bin/agentflow.js        # npx 启动桥接
```

## 开发说明

### 环境要求

- Go `1.26.0`
- 如需校验 `npx` 桥接，再准备 Node.js `>=16`

### 常用命令

```bash
gofmt -w .
go test ./...
go build -o /tmp/agentflow ./cmd/agentflow
bash -n install.sh
node --check bin/agentflow.js
```

### Release 产物

`.github/workflows/release.yml` 会为以下平台构建二进制：

- Linux `amd64`
- Linux `arm64`
- macOS `amd64`
- macOS `arm64`
- Windows `amd64`

## FAQ

<details>
<summary><b>AgentFlow 还是 Python 工具吗？</b></summary>
不是。AgentFlow 现在已经改为 Go 可执行文件实现与分发，安装、hooks、测试和发布流程都走 Go CLI，不再依赖 Python 运行时脚本。
</details>

<details>
<summary><b>已有自定义规则文件会怎样？</b></summary>
安装前会自动备份成带时间戳的 `*_bak` 文件，然后再写入 AgentFlow 的规则文件。
</details>

<details>
<summary><b>profile 有什么区别？</b></summary>
`lite` 最小化，`standard` 加入共享运行模块，`full` 再追加子代理、注意力和 Hooks 相关说明。
</details>

<details>
<summary><b>Codex 安装时会改什么？</b></summary>
会把 `reviewer.toml` 和 `architect.toml` 部署到 `~/.codex/agents/`，并把 `[features] multi_agent = true` 以及 `[agents.*]` 段 merge 到 `~/.codex/config.toml`。
</details>

<details>
<summary><b>Claude 安装时会改什么？</b></summary>
会把 AgentFlow 的 hooks merge 进 `~/.claude/settings.json`，同时保留非 AgentFlow 的现有 hooks。
</details>

## 参与贡献

开发、测试和发布流程请看 [CONTRIBUTING.md](CONTRIBUTING.md)。

## 许可证

[MIT](LICENSE)
