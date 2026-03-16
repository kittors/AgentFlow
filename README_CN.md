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

## 什么是 AgentFlow？

AgentFlow 是一个**多 CLI 智能工作流系统**，它从根本上改变了 AI 编程助手的工作方式。AgentFlow 不会对每个请求一视同仁，而是按五个维度为每次输入打分，将其路由到对应的工作流深度——从即时回答到完整的架构评审。它为 AI 编程 CLI（Codex、Claude、Gemini、Qwen、Grok、OpenCode）和 IDE（Cursor、Windsurf、Trae、VS Code Copilot、Cline、Antigravity）提供了统一的规则层，带来**智能路由、安全检测、持久化记忆和子代理编排**。

AgentFlow 以**单个 Go 二进制**发布，所有工作流资源内嵌其中——不需要 Python、`pip`、`uv` 或 PyInstaller。

---

## ✨ 核心功能

### 🎯 五级智能路由（R0–R4）

每次输入按 5 个维度打分——**行动需求**、**目标明确度**、**决策范围**、**影响范围**和 **EHRB 风险**——然后路由到对应的工作流深度：

| 级别 | 分数 | 触发场景 | 工作流 |
|------|------|----------|--------|
| **R0** 💬 | ≤ 3 | 闲聊、问答、概念解释 | 直接回复，无评估 |
| **R1** ⚡ | 4–6 | 单文件修复、格式调整 | 快速流程：评估 → 修改 → 验收 |
| **R2** 📝 | 7–9 | 新功能、多文件修改 | 简化设计 → 开发 → KB 同步 |
| **R3** 📊 | 10–12 | 复杂功能、跨模块重构 | 完整设计（含多方案对比）→ 开发 → KB 同步 |
| **R4** 🏗️ | ≥ 13 | 系统级重构、技术栈迁移 | 深度评估 → 多方案架构评审 → 分阶段开发 |

自动升级机制：R1→R2（范围超出预期时）；R2→R3（产生架构级影响时）；R3→R4（系统级重构时）。

### 🛡️ EHRB 三层安全检测

**EHRB**（Extremely High Risk Behavior，极度高风险行为）在命令**执行前**设置三道安全门：

| 层级 | 检测内容 |
|------|----------|
| **第一层 — 关键词扫描** | `rm -rf`、`DROP TABLE`、`git push -f`、`chmod 777`、password/secret/token 泄漏、PII 数据、支付操作 |
| **第二层 — 语义分析** | 数据安全违规、权限绕过意图、环境误指、逻辑漏洞 |
| **第三层 — 工具输出审查** | 外部工具输出中的命令注入、格式劫持、敏感信息泄漏 |

四种处理模式：**交互式**（警告 → 用户确认）、**委托式**（警告 → 降级为交互）、**持续执行**（记录风险但不中断，完成后报告）和**外部工具输出**（安全 → 正常，可疑 → 提示，高风险 → 警告）。

### 🧠 三层记忆模型（L0 / L1 / L2）

通过结构化的记忆架构，上下文可以在会话和项目之间持续存在：

| 层级 | 作用域 | 位置 | 内容 |
|------|--------|------|------|
| **L0 — 用户记忆** | 全局（跨项目） | `~/.agentflow/user/profile.md` | 用户偏好、技术栈、沟通风格、全局规则 |
| **L1 — 项目知识库** | 项目级（跨会话） | `.agentflow/kb/` | 项目结构、模块文档、架构决策、编码规范 |
| **L2 — 会话记录** | 会话级 | `.agentflow/sessions/` | 任务进度、关键决策、遇到的问题、上下文快照 |

会话启动时自动加载记忆（L0 + 最新 L2），会话结束时自动保存。阶段切换时自动触发快照保存。

### 🕸️ 知识图谱记忆 *（AgentFlow 独有）*

超越扁平文件知识库，AgentFlow 构建**基于图的项目记忆**，使用节点和边（`nodes.json` + `edges.json`）。知识图谱支持：

- 通过 `~graph` 命令进行渐进式披露查询
- 代码变更时自动更新图谱
- 模块、决策和模式之间的节点关系
- 比传统文件知识库更丰富的上下文

### 🔍 编码规范自动提取 *（AgentFlow 独有）*

AgentFlow 自动从代码库中发现编码模式，并在开发过程中执行约束：

- 自动扫描项目，将规范提取到 `extracted.json`
- `CONVENTION_CHECK=1` 时验证新代码是否符合已有规范
- 通过 `~conventions` 命令管理规范
- 确保整个代码库的一致性

### 🏗️ 架构扫描 *（AgentFlow 独有）*

主动检测问题，在它们变成技术债务之前发现：

- 检测**大文件**（应拆分）
- 识别**循环依赖**
- 标记**缺失测试**的关键代码路径
- 通过 `~dashboard` 生成项目仪表盘
- 通过 `~scan` 命令使用

### 🤖 RLM 子代理编排

面对复杂任务，AgentFlow 根据任务复杂度调度专门的子代理：

| 角色 | 职责 |
|------|------|
| **Reviewer**（审查员） | 代码审查、质量检查、测试覆盖率分析 |
| **Synthesizer**（综合员） | 信息汇总、多源分析 |
| **KB Keeper**（知识守护者） | 知识库维护、文档同步 |
| **Pkg Keeper**（包管理员） | 依赖管理、包分析 |
| **Writer**（写作者） | 文档撰写、README 生成 |
| **Architect**（架构师） | 系统设计、架构评估 |

基于复杂度分派：`simple` → 单个代理，`moderate` → 2 个代理，`complex` → 3+ 个代理，`architect` → 全团队 + 架构评审。同时映射到 CLI 原生子代理（如 Codex 内置的多代理系统）。

### 📏 项目规则生成

AgentFlow 为 CLI 和 IDE 生成 AI 优化的项目规则文件，支持三种 profile 级别：

| Profile | 内容 |
|---------|------|
| **lite** | 最小化规则集——仅核心路由和安全检测 |
| **standard** | + 共享运行模块（通用规则、模块加载、验收标准） |
| **full** | + 子代理编排指导、注意力控制、Hook 集成 |

**支持的规则目标：**

| 类别 | 目标 | 规则格式 |
|------|------|----------|
| CLI | Codex、Claude、Gemini、Qwen、Grok、OpenCode | `AGENTS.md` / `CLAUDE.md` / `GEMINI.md` / `QWEN.md` / `GROK.md` |
| IDE | Cursor、Windsurf、Trae、VS Code Copilot、Cline、Antigravity、Kiro | `.cursorrules` / `.windsurfrules` / `.trae/rules.md` 等 |

### 🔌 Skills 与 MCP 生态

AgentFlow 集成了 **Skills** 和 **MCP（模型上下文协议）** 生态系统：

- **Skills**：安装、列出和卸载可复用的 AI 技能包（如来自 [Vercel Skills 索引](https://vercel.com/docs/agent-resources/skills)）
- **MCP 服务器**：跨目标安装、搜索和管理 MCP 服务器；配置自动写入各工具的原生配置文件（`config.toml`、`settings.json`、`mcp.json` 等）
- **推荐 MCP**：`context7`（文档查询）、`playwright`（浏览器自动化）、`filesystem`（文件操作）

### 🔄 执行模式

四种执行模式让你完全掌控自动化程度：

| 模式 | 行为 |
|------|------|
| **交互式**（默认） | 逐步执行，关键决策点等待用户确认 |
| **委托式**（`~auto`） | 阶段间自动推进，仅遇 EHRB 风险时暂停 |
| **委托规划**（`~plan`） | 自动设计后停止——生成方案和任务清单，不触碰代码 |
| **持续执行** | 全自主：自动执行所有阶段、自动审查、自动测试，最多 5 次重试；完成后输出完整报告 |

### 📊 内置命令

| 命令 | 说明 |
|------|------|
| `~init` | 初始化项目知识库（扫描、索引、图谱、规范） |
| `~auto` | 全自动执行工作流 |
| `~plan` | 仅规划——设计 + 任务清单，不修改代码 |
| `~exec` | 执行已有方案，或 `~exec <需求>` 进入组合流程 |
| `~status` | 查看工作流状态和任务进度 |
| `~review` | 触发代码审查 |
| `~scan` | 架构扫描（大文件、循环依赖、缺失测试） |
| `~conventions` | 提取和检查编码规范 |
| `~graph` | 知识图谱操作 |
| `~dashboard` | 生成项目仪表盘 |
| `~memory` | 管理记忆层（L0/L1/L2） |
| `~rlm` | 子代理管理 |
| `~validatekb` | 验证知识库一致性 |

---

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

> Windows 建议：优先直接从 GitHub Releases 下载 `agentflow-windows-amd64.exe`，放到 `PATH` 后即可使用。

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

### 离线安装（无网络 / 内网环境）

如果目标机器无法访问互联网，可以在另一台联网机器上下载二进制，然后传输过去。

#### 第一步 — 在联网机器上下载

前往 [Releases](https://github.com/kittors/AgentFlow/releases)，下载与**目标机器**操作系统和 CPU 架构相匹配的文件：

| 操作系统 | CPU 架构 | 文件名 |
|----------|---------|--------|
| macOS | Apple Silicon (M1/M2/M3/M4) | `agentflow-darwin-arm64` |
| macOS | Intel | `agentflow-darwin-amd64` |
| Linux | x86_64 | `agentflow-linux-amd64` |
| Linux | ARM64 / aarch64 | `agentflow-linux-arm64` |
| Windows | x86_64 | `agentflow-windows-amd64.exe` |

#### 第二步 — 传输

通过 U 盘、`scp`、共享文件夹或其他可用方式将下载的文件传输到目标机器。

#### 第三步 — 安装

**macOS**

```bash
# 移除隔离属性（从其他 Mac 传输或通过浏览器下载时可能被标记）
xattr -d com.apple.quarantine agentflow-darwin-arm64 2>/dev/null

chmod +x agentflow-darwin-arm64
sudo mv agentflow-darwin-arm64 /usr/local/bin/agentflow
agentflow version
```

**Linux**

```bash
chmod +x agentflow-linux-amd64
sudo mv agentflow-linux-amd64 /usr/local/bin/agentflow
agentflow version
```

> 如果没有 `sudo` 权限，移动到用户可写且已在 `PATH` 中的目录，例如 `~/.local/bin/`：
>
> ```bash
> mkdir -p ~/.local/bin
> mv agentflow-linux-amd64 ~/.local/bin/agentflow
> # 确保 ~/.local/bin 在 PATH 中
> export PATH="$HOME/.local/bin:$PATH"
> ```

**Windows**

```powershell
# 1. 创建 AgentFlow 目录（如不存在）
New-Item -ItemType Directory -Force -Path "$env:LOCALAPPDATA\AgentFlow"

# 2. 移动二进制文件
Move-Item agentflow-windows-amd64.exe "$env:LOCALAPPDATA\AgentFlow\agentflow.exe"

# 3. 添加到 PATH（当前用户，永久生效）
$currentPath = [Environment]::GetEnvironmentVariable('Path', 'User')
if ($currentPath -notlike "*AgentFlow*") {
    [Environment]::SetEnvironmentVariable('Path', "$currentPath;$env:LOCALAPPDATA\AgentFlow", 'User')
}

# 4. 在当前会话刷新 PATH
$env:Path = [Environment]::GetEnvironmentVariable('Path', 'User') + ";" + [Environment]::GetEnvironmentVariable('Path', 'Machine')

# 5. 验证
agentflow version
```

> 添加 `PATH` 后可能需要**重新打开终端窗口**才能生效。

---

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

# 项目级规则（CLI + IDE）
agentflow rules detect
agentflow rules install codex claude gemini qwen kiro cursor windsurf trae vscode-copilot cline antigravity

# Skills（Codex）
agentflow skill list codex
agentflow skill install codex https://skills.sh/vercel/turborepo/turborepo
agentflow skill uninstall codex turborepo

# MCP 服务器（全局）
agentflow mcp install claude context7 --set-env=CONTEXT7_API_KEY=YOUR_API_KEY
agentflow mcp install claude playwright
agentflow mcp list claude
agentflow mcp search playwright
```

当未传子命令且 stdin 是 TTY 时，AgentFlow 会进入 Bubble Tea TUI。方向键、`Enter`、`Space`、`Esc` 在 macOS 和 Windows 终端里保持一致。

---

## 支持的目标

### CLI 目标

| 目标 | 配置目录 | 规则文件 | 额外集成 |
|------|----------|----------|----------|
| Codex CLI | `~/.codex/` | `AGENTS.md` | 注入 `agents/reviewer.toml`、`agents/architect.toml`，并 merge `config.toml` 多代理配置 |
| Claude Code | `~/.claude/` | `CLAUDE.md` | merge / 清理 `settings.json` hooks |
| Gemini CLI | `~/.gemini/` | `GEMINI.md` | 部署规则与嵌入模块 |
| Qwen CLI | `~/.qwen/` | `QWEN.md` | 部署规则与嵌入模块 |
| Grok CLI | `~/.grok/` | `GROK.md` | 部署规则与嵌入模块 |
| OpenCode | `~/.config/opencode/` | `AGENTS.md` | 部署规则与嵌入模块 |

### IDE 目标

| 目标 | 规则格式 |
|------|----------|
| Cursor | `.cursorrules` |
| Windsurf | `.windsurfrules` |
| Trae | `.trae/rules.md` |
| VS Code Copilot | `.github/copilot-instructions.md` |
| Cline | `.clinerules` |
| Kiro | Kiro 规则格式 |
| Antigravity | `.agents/` 目录 |

---

## 工作原理

```
┌──────────────────────────────────────────────────────────────┐
│                       用户输入                                │
└──────────────────┬───────────────────────────────────────────┘
                   │
                   ▼
┌──────────────────────────────────────────────────────────────┐
│  ① 路由器：按 5 维度打分 → 路由到 R0/R1/R2/R3/R4             │
│     • ~命令 → 命令处理器                                      │
│     • Skill/MCP 匹配 → 工具协议                               │
│     • 其他 → 级别路由                                          │
└──────────────────┬───────────────────────────────────────────┘
                   │
         ┌─────────┼─────────┬──────────┬──────────┐
         ▼         ▼         ▼          ▼          ▼
      R0 💬     R1 ⚡     R2 📝     R3 📊     R4 🏗️
      直接       快速      简化       标准       架构级
      回复       修复      流程       流程       流程
                   │         │          │          │
                   ▼         ▼          ▼          ▼
         ┌─────────────────────────────────────────────────────┐
         │  ② EHRB 安全门（三层检测）                            │
         │     关键词 → 语义分析 → 工具输出审查                   │
         └──────────────────┬──────────────────────────────────┘
                            │
                            ▼
         ┌─────────────────────────────────────────────────────┐
         │  ③ 阶段链：EVALUATE → DESIGN → DEVELOP              │
         │     按阶段加载模块 · 子代理分派                        │
         └──────────────────┬──────────────────────────────────┘
                            │
                            ▼
         ┌─────────────────────────────────────────────────────┐
         │  ④ KB 同步 + 会话保存                                │
         │     L1 知识库 · L2 会话 · 图谱更新                    │
         └─────────────────────────────────────────────────────┘
```

---

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
│   ├── core/               # 路由、安全、子代理、注意力、Hooks
│   ├── stages/             # DESIGN 和 DEVELOP 阶段模块
│   ├── functions/          # ~init、~plan、~exec、~scan、~graph 等
│   ├── services/           # 记忆、知识库、注意力、包管理服务
│   ├── agents/             # reviewer.toml、architect.toml 角色配置
│   ├── rlm/roles/          # 6 个专业子代理角色定义
│   ├── hooks/              # Claude hooks JSON、Codex hooks TOML
│   ├── templates/          # KB、方案、会话、变更日志、规范模板
│   ├── project_rules/      # IDE 适配规则文件（Cursor、Windsurf 等）
│   └── rules/              # 缓存、扩展、状态、工具规则
├── embed.go                # 静态资源嵌入
├── install.sh              # POSIX 二进制安装脚本
├── install.ps1             # Windows 二进制安装脚本
└── bin/agentflow.js        # npx 启动桥接
```

---

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

---

## FAQ

<details>
<summary><b>AgentFlow 还是 Python 工具吗？</b></summary>
不是。AgentFlow 现在已经改为 Go 可执行文件实现与分发，安装、hooks、测试和发布流程都走 Go CLI，不再依赖 Python 运行时脚本。
</details>

<details>
<summary><b>已有自定义规则文件会怎样？</b></summary>
安装前会自动备份成带时间戳的 <code>*_bak</code> 文件，然后再写入 AgentFlow 的规则文件。
</details>

<details>
<summary><b>profile 有什么区别？</b></summary>
<code>lite</code> 最小化（仅核心路由和安全检测），<code>standard</code> 加入共享运行模块（通用规则、模块加载、验收标准），<code>full</code> 再追加子代理编排指导、注意力控制和 Hook 集成。
</details>

<details>
<summary><b>Codex 安装时会改什么？</b></summary>
会把 <code>reviewer.toml</code> 和 <code>architect.toml</code> 部署到 <code>~/.codex/agents/</code>，并把 <code>[features] multi_agent = true</code> 以及 <code>[agents.*]</code> 段 merge 到 <code>~/.codex/config.toml</code>。
</details>

<details>
<summary><b>Claude 安装时会改什么？</b></summary>
会把 AgentFlow 的 hooks merge 进 <code>~/.claude/settings.json</code>，同时保留非 AgentFlow 的现有 hooks。
</details>

<details>
<summary><b>AgentFlow 和直接写 AGENTS.md / CLAUDE.md 有什么区别？</b></summary>
AgentFlow 是一个完整的工作流系统，而不仅仅是一个规则文件。它增加了五级智能路由、EHRB 三层安全检测、三层持久化记忆（用户/项目/会话）、知识图谱记忆、自动编码规范提取、架构扫描、6 种专业角色的子代理编排，以及结构化阶段链（EVALUATE → DESIGN → DEVELOP）。单纯的规则文件只给 AI 提供指令；AgentFlow 给 AI 提供了一套完整的操作框架。
</details>

<details>
<summary><b>能同时在多个 CLI 中使用 AgentFlow 吗？</b></summary>
可以。<code>agentflow install --all</code> 会部署到所有检测到的 CLI 目标。每个目标都有自己的规则文件格式和额外集成（hooks、多代理配置等）。<code>.agentflow/</code> 中的项目级知识库在所有 CLI 之间共享。
</details>

---

## 参与贡献

开发、测试和发布流程请看 [CONTRIBUTING.md](CONTRIBUTING.md)。

## 许可证

[MIT](LICENSE)
