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
[![Python 3.10+](https://img.shields.io/badge/Python-3.10+-3776AB.svg?logo=python&logoColor=white)](https://www.python.org)
[![CI](https://github.com/kittors/AgentFlow/actions/workflows/ci.yml/badge.svg)](https://github.com/kittors/AgentFlow/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/kittors/AgentFlow?include_prereleases&label=Release)](https://github.com/kittors/AgentFlow/releases)

**`五级路由`** · **`EHRB 安全检测`** · **`知识图谱记忆`** · **`子代理编排`** · **`编码规范提取`**

</div>

---

## 为什么选择 AgentFlow？

大多数 AI 助手可以分析任务，但在真正交付之前就停下了。AgentFlow 添加了**严格路由**、**分阶段执行**、**安全门控**和**验收检查** —— 确保每个任务都能高质量完成。

<table>
<tr><td>

**🎯 有何不同？**

- 📊 **五级路由** — 按比例投入精力，从快速修复到架构重设计
- 🏗️ **R4 架构模式** — 专为系统级重构设计的工作流
- 🧠 **知识图谱** — 基于图结构的项目记忆，跨会话持久化
- 🔍 **编码规范提取** — 自动发现并强制执行编码模式
- 📡 **架构扫描** — 主动检测代码异味和结构问题
- 📈 **项目仪表盘** — HTML 状态可视化

</td></tr>
</table>

## 快速开始

### 方式 A：一键安装脚本 _（推荐）_

**macOS / Linux：**

```bash
curl -fsSL https://raw.githubusercontent.com/kittors/AgentFlow/main/install.sh | bash
```

**Windows PowerShell：**

```powershell
irm https://raw.githubusercontent.com/kittors/AgentFlow/main/install.ps1 | iex
```

### 方式 B：npx (Node.js ≥ 16)

```bash
npx agentflow
```

### 方式 C：UV（隔离环境）

```bash
uv tool install --from git+https://github.com/kittors/AgentFlow agentflow && agentflow
```

### 方式 D：pip (Python ≥ 3.10)

```bash
pip install git+https://github.com/kittors/AgentFlow.git && agentflow
```

### 安装到指定 CLI

```bash
agentflow                     # 交互式菜单
agentflow install codex       # 直接指定目标
agentflow install --all       # 安装到所有检测到的 CLI
```

### 更新

```bash
agentflow update
```

### 验证与卸载

```bash
agentflow status              # 查看安装状态
agentflow version             # 查看版本
agentflow uninstall codex     # 从指定目标卸载
agentflow uninstall --all     # 从所有目标卸载
agentflow clean               # 清理缓存
```

---

## 功能特性

### 🎯 五级路由（R0–R4）

对每个输入进行 **5 个维度** 评分，路由到合适的处理流程：

| 级别 | 分数 | 场景 | 流程 |
|:----:|:----:|------|------|
| R0 💬 | ≤ 3 | 闲聊、问答 | 直接回复 |
| R1 ⚡ | 4-6 | 快速修复 | 定位 → 修改 → 验收 |
| R2 📝 | 7-9 | 多文件变更 | 确认 → 设计 → 开发 |
| R3 📊 | 10-12 | 复杂功能 | 确认 → 多方案设计 → 开发 |
| R4 🏗️ | ≥ 13 | 架构重构 | 评估 → 设计+评审 → 分阶段开发 |

**5 评分维度：** 行动需求(0-3) · 目标明确度(0-3) · 决策范围(0-3) · 影响范围(0-3) · EHRB 风险(0-3)

### 🛡️ EHRB 三层安全检测

在执行**之前**拦截破坏性操作：

| 层级 | 功能 |
|------|------|
| **关键词扫描** | 检测 `rm -rf`、`DROP TABLE`、`git push -f`、密钥、个人信息、支付操作 |
| **语义分析** | 识别权限绕过、环境误指、逻辑漏洞 |
| **工具输出检查** | 拦截指令注入、格式劫持、敏感信息泄露 |

### 🤖 RLM 子代理编排

**6 个专业角色** + 原生 CLI 子代理，根据任务复杂度自动调度：

| 角色 | 职责 | 触发条件 |
|------|------|----------|
| `reviewer` | 代码质量 + 安全审查 | 涉及核心模块的复杂任务 |
| `synthesizer` | 多方案综合分析 | 复杂 + 多评估维度 |
| `kb_keeper` | 知识库同步 | KB 开启时 |
| `pkg_keeper` | 方案包管理 | 设计/开发阶段 |
| `writer` | 文档生成 | 手动 `~rlm spawn writer` |
| `architect` | 系统级架构评审 | R4 / 架构级复杂度 |

### 🧠 三层记忆

| 层级 | 范围 | 内容 |
|:----:|------|------|
| **L0** | 全局（跨项目） | 用户偏好、技术栈、沟通风格 |
| **L1** | 项目级 | 知识库、模块文档、架构决策 |
| **L2** | 会话级 | 任务进度、决策记录、上下文 |

### ⚡ 独有功能

| 功能 | 命令 | 描述 |
|------|:----:|------|
| **知识图谱** | `~graph` | 基于图结构的项目记忆，支持渐进式查询 |
| **编码规范提取** | `~conventions` | 自动发现编码模式，开发阶段强制检查 |
| **架构扫描** | `~scan` | 检测大文件、循环依赖、缺失测试 |
| **仪表盘** | `~dashboard` | 生成 HTML 项目状态仪表盘 |
| **R4 架构路由** | _(自动)_ | 系统级变更的专属五阶段工作流 |
| **Architect 角色** | _(自动)_ | 架构评审专属子代理 |
| **上下文窗口管理** | _(自动)_ | 超过 80% 时主动总结释放 |
| **规范检查门控** | _(自动)_ | 开发阶段自动代码合规验证 |

---

## 工作流命令

所有命令在 **AI 聊天中** 使用，不是系统终端命令。

| 命令 | 描述 |
|:----:|------|
| `~init` | 初始化项目知识库 |
| `~auto` | 自动执行完整工作流 |
| `~plan` | 仅规划，开发前停止 |
| `~exec` | 执行已有方案 |
| `~status` | 显示工作流状态 |
| `~review` | 触发代码审查 |
| `~scan` | 架构扫描 |
| `~conventions` | 提取/检查编码规范 |
| `~graph` | 知识图谱操作 |
| `~dashboard` | 生成项目仪表盘 |
| `~memory` | 管理记忆层 |
| `~rlm` | 子代理管理 |
| `~validatekb` | 验证知识库一致性 |

---

## 支持的 CLI 目标

| 目标 | 配置目录 | Hooks |
|------|:--------:|:-----:|
| **Claude Code** | `~/.claude/` | ✅ 完整 (9 事件) |
| **Codex CLI** | `~/.codex/` | ✅ 通知 |
| **Gemini CLI** | `~/.gemini/` | — |
| **OpenCode** | `~/.config/opencode/` | — |
| **Qwen CLI** | `~/.qwen/` | — |
| **Grok CLI** | `~/.grok/` | — |

### Codex CLI 兼容性说明

> 以下 `config.toml` 设置可能影响 AgentFlow：

| 设置 | 影响 | 建议 |
|------|------|------|
| `steer = true` | 可能干扰工作流交互 | 出现问题时禁用 |
| `child_agents_md = true` | 可能与 AgentFlow 指令冲突 | 禁用 |
| `project_doc_max_bytes` | 值过低会截断 AGENTS.md | 安装时自动设为 98304 |
| `agent_max_depth = 1` | 限制子代理嵌套深度 | 保持 ≥ 2 |
| `agent_max_threads` | 限制并行子代理数 | 保持默认 (6) 或更高 |

---

## FAQ

<details>
<summary><b>这是 Python CLI 工具还是提示词包？</b></summary>
两者兼有。CLI 管理安装和更新；工作流行为来自部署到 AI 编码助手配置目录的 <code>AGENTS.md</code> 和模块文件。
</details>

<details>
<summary><b>应该安装到哪个目标？</b></summary>
安装到你使用的 CLI：<code>codex</code>、<code>claude</code>、<code>gemini</code>、<code>qwen</code>、<code>grok</code> 或 <code>opencode</code>。使用 <code>--all</code> 安装到所有检测到的 CLI。
</details>

<details>
<summary><b>如果规则文件已存在怎么办？</b></summary>
非 AgentFlow 文件会自动带时间戳备份。备份文件名会在控制台输出中显示。
</details>

<details>
<summary><b>什么是 RLM？</b></summary>
角色语言模型（Role Language Model）——一个子代理编排系统，包含 6 个专业角色和原生 CLI 子代理，根据任务复杂度自动调度。
</details>

<details>
<summary><b>记忆是否跨会话持久化？</b></summary>
是的。L0 用户记忆是全局的，L1 项目知识库是项目级的，L2 会话摘要在阶段转换时自动保存。
</details>

<details>
<summary><b>什么是 R4 架构模式？</b></summary>
专为系统级重构、技术栈迁移和全新架构设计的路由级别。包含额外的 EVALUATE 阶段、多方案设计配合架构评审、分阶段开发。这是同类工具中没有的功能。
</details>

---

## 贡献

请参阅 [CONTRIBUTING.md](CONTRIBUTING.md) 了解开发环境配置、代码规范和 PR 流程。

## 许可证

[MIT](LICENSE) — 自由使用。

---

<div align="center">

**AgentFlow** — 比分析更进一步，持续工作直到实现和验证完成。

</div>
