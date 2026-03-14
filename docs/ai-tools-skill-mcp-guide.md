# AI 编码工具 — Skill 与 MCP 安装配置指南

> **最后更新**: 2026-03-14
>
> 本文档汇总了当前主流 AI 编码 CLI 工具和 IDE 中 **Skill（技能/自定义指令）** 和 **MCP（Model Context Protocol，模型上下文协议）** 的安装与配置方法。

---

## 目录

- [概念说明](#概念说明)
- [CLI 工具](#cli-工具)
  - [Claude Code（Anthropic）](#1-claude-codeanthropic)
  - [Codex CLI（OpenAI）](#2-codex-cliopenai)
  - [Gemini CLI（Google）](#3-gemini-cligoogle)
  - [Qwen Code（阿里/通义千问）](#4-qwen-code阿里通义千问)
  - [Aider](#5-aider)
  - [Augment Code CLI（Auggie）](#6-augment-code-cliauggie)
- [IDE 工具](#ide-工具)
  - [Cursor](#7-cursor)
  - [Windsurf（Codeium）](#8-windsurfcodeium)
  - [Trae（字节跳动）](#9-trae字节跳动)
  - [Kiro（AWS）](#10-kiroaws)
  - [VS Code + GitHub Copilot](#11-vs-code--github-copilot)
  - [VS Code + Cline 扩展](#12-vs-code--cline-扩展)
  - [JetBrains IDE](#13-jetbrains-ide)
  - [Antigravity（Google DeepMind）](#14-antigravitygoogle-deepmind)
- [快速对比表](#快速对比表)
- [通用 MCP Server 安装参考](#通用-mcp-server-安装参考)

---

## 概念说明

### 什么是 Skill？

**Skill（技能）** 是 AI 编码工具的自定义指令机制，允许开发者通过特定的文件或配置来引导 AI 的行为。不同工具有不同的命名：

| 工具 | Skill 文件 | 说明 |
|------|-----------|------|
| Claude Code | `CLAUDE.md` | 项目级自定义指令 |
| Codex CLI | `AGENTS.md`、`codex.md` | 代理指令文件 |
| Gemini CLI | `GEMINI.md` | 项目级自定义指令 |
| Qwen Code | `QWEN.md` | 项目级自定义指令 |
| Cursor | `.cursor/rules/`、`.cursorrules` | 项目规则文件 |
| Windsurf | `.windsurfrules` | 项目规则 |
| Trae | `.trae/rules/` | 项目规则 |
| Kiro | `.kiro/` + Powers | 技能包系统 |
| GitHub Copilot | `.github/copilot-instructions.md` | 项目自定义指令 |
| Cline | `.clinerules` | 自定义规则 |
| JetBrains | AI Assistant 设置面板 | IDE 内设置 |
| Augment Code | `~/.augment/skills/` | 技能目录 |

### 什么是 MCP？

**MCP（Model Context Protocol）** 是一个开放标准协议，用于连接 AI 模型与外部工具和数据源。通过 MCP Server，AI 工具可以：

- 查询数据库、调用 API
- 操作 GitHub、Jira、Slack 等服务
- 执行浏览器自动化、文件操作等

**MCP Server 传输类型：**

| 类型 | 说明 |
|------|------|
| **STDIO** | 本地进程，通过标准输入/输出通信（最常用） |
| **SSE** | 通过 Server-Sent Events 远程通信 |
| **Streamable HTTP** | 通过流式 HTTP 通信 |

---

## AgentFlow 一键配置（推荐）

如果你在项目里使用了 **AgentFlow**（本仓库），可以用 CLI 一键完成：

### 1) 写入「项目级 Skill/规则文件」（CLI + IDE）

在项目根目录执行：

```bash
# 检测当前项目已有的规则文件
agentflow rules detect

# 写入指定目标的项目级规则文件（可多选）
agentflow rules install codex claude gemini qwen kiro cursor windsurf trae vscode-copilot cline antigravity

# 可选：控制注入深度（lite/standard/full）
agentflow rules install codex claude --profile=full
```

> 说明：如果目标文件已存在且不是 AgentFlow 管理文件，AgentFlow 会先在同目录生成备份文件再覆盖写入。

### 2) 管理「全局 MCP」配置（按工具真实配置文件写入）

```bash
# 查看某个工具已配置的 MCP server
agentflow mcp list gemini

# 安装内置 MCP（示例：Context7）
agentflow mcp install gemini context7 --set-env=CONTEXT7_API_KEY=...

# Cursor / Windsurf 也支持（写入各自的用户级 JSON 配置）
agentflow mcp install cursor filesystem --allow=/path/to/dir
agentflow mcp list windsurf
```

---

## CLI 工具

### 1. Claude Code（Anthropic）

**安装 CLI：**

```bash
npm install -g @anthropic-ai/claude-code
```

**配置 Skill — `CLAUDE.md`：**

在项目根目录创建 `CLAUDE.md` 文件，写入项目特定的指令，如编码规范、架构决策等：

```markdown
# 项目规范
- 使用 TypeScript 严格模式
- 所有 API 返回统一 Response 格式
- 测试覆盖率不低于 80%
```

Claude Code 启动时会自动读取此文件作为上下文。

**配置 MCP：**

Claude Code 支持通过 `claude mcp` 命令管理 MCP Server：

```bash
# 添加 MCP Server
claude mcp add <server-name> -- <command> [args...]

# 示例：添加 GitHub MCP Server
claude mcp add github -- npx -y @modelcontextprotocol/server-github

# 示例：添加文件系统 MCP Server
claude mcp add filesystem -- npx -y @modelcontextprotocol/server-filesystem /path/to/dir

# 查看已配置的 MCP Server
claude mcp list

# 删除 MCP Server
claude mcp remove <server-name>
```

**配置文件位置：** `~/.claude/` 目录下

也可以编辑 JSON 配置文件手动配置：

```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_TOKEN": "${GITHUB_TOKEN}"
      }
    }
  }
}
```

**项目级 MCP：** 在项目根目录创建 `.mcp.json` 文件即可配置项目级 MCP Server。

---

### 2. Codex CLI（OpenAI）

**安装 CLI：**

```bash
# npm 安装
npm install -g @openai/codex

# macOS Homebrew
brew install --cask codex
```

**配置 Skill — `AGENTS.md`：**

在项目根目录创建 `AGENTS.md` 或 `codex.md` 文件。Codex CLI 会自动加载该文件作为上下文指令。

**配置 MCP：**

方式一：CLI 命令

```bash
# 添加 MCP Server
codex mcp add <server-name> -- <command> [args...]

# 示例：添加 Context7 文档 MCP
codex mcp add context7 -- npx -y @upstash/context7-mcp

# 查看 MCP 帮助
codex mcp --help
```

方式二：编辑配置文件 `~/.codex/config.toml`

```toml
[mcp_servers.github]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-github"]

[mcp_servers.github.env]
GITHUB_TOKEN = "your-token-here"
```

**项目级 MCP：** 在项目根目录创建 `.codex/config.toml` 文件。

---

### 3. Gemini CLI（Google）

**安装 CLI：**

```bash
npm install -g @anthropic-ai/gemini-cli
# 或者
npx -y @anthropic-ai/gemini-cli
```

> 需要 Node.js 18+

**配置 Skill — `GEMINI.md`：**

在项目根目录创建 `GEMINI.md` 文件，格式与 `CLAUDE.md` 类似。

**配置 MCP：**

编辑配置文件 `~/.gemini/settings.json`：

```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_TOKEN": "${GITHUB_TOKEN}"
      }
    },
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/path/to/dir"]
    }
  }
}
```

Gemini CLI 也支持通过命令行管理 MCP：

```bash
gemini mcp add <server-name> -- <command> [args...]
```

---

### 4. Qwen Code（阿里/通义千问）

**安装 CLI：**

```bash
# npm 安装
npm install -g @anthropic-ai/qwen-code

# 或使用 curl
curl -fsSL https://chat.qwen.ai/install.sh | sh
```

> 需要 Node.js v18+

**配置 Skill — `QWEN.md`：**

在项目根目录创建 `QWEN.md` 文件。

**配置 MCP：**

方式一：CLI 命令

```bash
# 添加 MCP Server
qwen mcp add <server-name> -- <command> [args...]

# 删除 MCP Server
qwen mcp remove <server-name>
```

方式二：编辑 `~/.qwen/settings.json`

```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_TOKEN": "${GITHUB_TOKEN}"
      }
    }
  },
  "modelProviders": [
    {
      "name": "bailian",
      "baseURL": "https://dashscope.aliyuncs.com/compatible-mode/v1",
      "envKey": "BAILIAN_API_KEY"
    }
  ]
}
```

**API Key 获取：** 登录 [阿里云百炼平台](https://bailian.console.aliyun.com/) 获取 `BAILIAN_API_KEY`。

---

### 5. Aider

**安装 CLI：**

```bash
pip install aider-chat
```

**配置 Skill：**

Aider 使用 `.aider.conf.yml` 配置文件（支持用户级 `~/.aider.conf.yml` 和项目级）：

```yaml
model: claude-3-5-sonnet
auto-commits: true
dark-mode: true
```

也支持 `.env` 文件配置环境变量。

**配置 MCP：**

Aider 本身不直接作为 MCP Client，但提供了 **Aider MCP Server**，允许其他 MCP Client 使用 Aider 的编辑能力：

```bash
# 安装 Aider MCP Server
pip install aider-mcp

# 运行
aider-mcp --repo-path=/path/to/your/repo
```

也可以通过 **mcpm-aider bridge** 方式集成：

```bash
# 安装桥接工具
pip install mcpm

# 启动桥接
mcpm-aider bridge
```

---

### 6. Augment Code CLI（Auggie）

**安装 CLI：**

从 [augmentcode.com](https://augmentcode.com) 下载安装。

**配置 Skill：**

Augment Code 使用 `~/.augment/skills/` 目录存放技能文件：

```
~/.augment/
├── skills/          # 技能目录
├── settings.json    # 配置文件
└── rules/           # 规则文件
```

**配置 MCP：**

方式一：设置面板（Easy MCP，一键配置）

打开 Augment Settings → MCP Servers → 选择预配置的 MCP Server → 填入 API Token

方式二：编辑 `~/.augment/settings.json`

```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_TOKEN": "${GITHUB_TOKEN}"
      }
    }
  }
}
```

---

## IDE 工具

### 7. Cursor

**安装：** 从 [cursor.com](https://cursor.com) 下载安装

**配置 Skill：**

- **项目级规则：** 在项目根目录创建 `.cursor/rules/` 目录，放入 `.mdc` 规则文件
- **旧版方式：** 项目根目录创建 `.cursorrules` 文件
- **全局规则：** `Cursor Settings → General → Rules for AI`

```markdown
# .cursor/rules/coding-standards.mdc
---
description: 编码规范
globs: ["**/*.ts", "**/*.tsx"]
---
- 使用 TypeScript 严格模式
- 组件使用函数式写法
```

**配置 MCP：**

方式一：GUI 设置

1. 打开 `Cursor Settings → Features → MCP Servers`（或 `Tools and Integrations`）
2. 点击 `Add new MCP server`
3. 填入名称、类型（command/SSE）、命令

方式二：编辑配置文件 `~/.cursor/mcp.json`

```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_TOKEN": "${GITHUB_TOKEN}"
      }
    },
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/path/to/dir"]
    }
  }
}
```

方式三：项目级 `.cursor/mcp.json`

---

### 8. Windsurf（Codeium）

**安装：** 从 [windsurf.com](https://windsurf.com) 下载安装

**配置 Skill：**

在项目根目录创建 `.windsurfrules` 文件：

```markdown
- 使用 React 函数式组件
- 状态管理使用 Zustand
- 样式使用 TailwindCSS
```

也可在 `Windsurf Settings → Cascade → Custom Instructions` 设置全局规则。

**配置 MCP：**

1. 打开 `Windsurf Settings` (`Cmd/Ctrl + ,`)
2. 导航到 `Plugins (MCP servers)` → `Manage Plugins` → `View raw config`
3. 编辑 `mcp_config.json`

**配置文件位置：**

- **macOS：** `~/.codeium/windsurf/mcp_config.json`
- **Windows：** `%USERPROFILE%\.codeium\windsurf\mcp_config.json`

```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_TOKEN": "${GITHUB_TOKEN}"
      }
    }
  }
}
```

**项目级配置：** 在项目根目录创建 `.windsurf/config.json`

---

### 9. Trae（字节跳动）

**安装：** 从 [trae.ai](https://trae.ai) 下载安装

**配置 Skill：**

- **项目级：** 在项目中创建 `.trae/rules/` 目录或项目级规则文件
- **全局级：** Trae Settings → AI Rules

**配置 MCP：**

方式一：Trae 内置 Marketplace

在 Trae 中搜索并一键安装 MCP Server。

方式二：手动配置

通过 Trae 设置面板添加自定义 MCP Server，支持三种传输类型：

| 类型 | 配置方式 |
|------|---------|
| stdio | 本地命令执行 |
| SSE | 远程 URL |
| Streamable HTTP | 远程 URL |

方式三：Agent 绑定

在 Trae 中创建 Agent 时，可以将特定 MCP Server 绑定到该 Agent。

**配置文件格式示例：**

```json
{
  "mcpServers": {
    "playwright": {
      "command": "npx",
      "args": ["-y", "@anthropic/mcp-playwright"]
    }
  }
}
```

---

### 10. Kiro（AWS）

**安装：** 从 [kiro.dev](https://kiro.dev) 下载安装

**配置 Skill — Powers 系统：**

Kiro 独有的 **Powers** 系统是一种自包含的技能包，捆绑了 MCP Server + 文档 + 最佳实践：

```
.kiro/
└── powers/
    └── my-power/
        ├── power.md      # 技能说明
        └── mcp.json      # 关联的 MCP 配置
```

Powers 会根据对话上下文中的关键词按需加载。

**配置 MCP：**

方式一：命令面板

1. 按 `Cmd + Shift + P`（Mac）或 `Ctrl + Shift + P`（Win/Linux）
2. 搜索 `MCP` → 选择 `Kiro: Open workspace MCP config` 或 `Kiro: Open user MCP config`

方式二：手动编辑配置文件

- **项目级：** `.kiro/settings/mcp.json`
- **用户级：** `~/.kiro/settings/mcp.json`

```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_TOKEN": "${GITHUB_TOKEN}"
      }
    }
  }
}
```

> ⚠️ 安全建议：使用环境变量引用 `${API_TOKEN}` 而非硬编码敏感信息。

---

### 11. VS Code + GitHub Copilot

**安装：** 在 VS Code 中安装 `GitHub Copilot` 和 `GitHub Copilot Chat` 扩展

**配置 Skill：**

在项目中创建 `.github/copilot-instructions.md`：

```markdown
# 项目指令
- 代码风格遵循 Airbnb 规范
- 使用 ESM 模块
- 注释使用中文
```

也可以在 `.vscode/settings.json` 中配置自定义指令。

**配置 MCP：**

方式一：Agent Mode GUI

1. 在 Copilot Chat 中进入 `Agent Mode`
2. 点击工具菜单 → `Add More Tools` → `Add MCP Server`
3. 选择类型（NPX/PIP/Docker 等）并输入包名

方式二：配置文件 `.vscode/mcp.json`

```json
{
  "servers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_TOKEN": "${input:github_token}"
      }
    }
  },
  "inputs": [
    {
      "id": "github_token",
      "type": "promptString",
      "description": "GitHub Personal Access Token",
      "password": true
    }
  ]
}
```

方式三：自动发现

开启 `chat.mcp.discovery.enabled` 设置后，VS Code 会自动发现其他应用（如 Claude Desktop）已配置的 MCP Server。

---

### 12. VS Code + Cline 扩展

**安装：** 在 VS Code 扩展市场搜索 `Cline` 并安装

**配置 Skill：**

- 在项目根目录创建 `.clinerules` 文件
- 或在 Cline 设置中配置 Custom Instructions

**配置 MCP：**

方式一：Cline 内置 Marketplace

Cline 有内建的 MCP Marketplace，支持一键安装：

1. 打开 Cline 面板
2. 点击 `MCP Servers` 标签
3. 浏览分类并一键安装

方式二：设置文件

在 VS Code `settings.json` 中配置 `cline.mcpServers`：

```json
{
  "cline.mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_TOKEN": "your-token"
      }
    }
  }
}
```

方式三：自动安装

给 Cline 一个 MCP Server 的 GitHub 仓库 URL，它会自动克隆、构建并配置。

---

### 13. JetBrains IDE

> 适用于 IntelliJ IDEA、WebStorm、PyCharm、GoLand 等全系列 IDE

**安装 AI Assistant：**

1. `File → Settings → Plugins` → 搜索 `AI Assistant` → 安装
2. 需要 JetBrains AI Service 许可证

**配置 Skill：**

通过 `Preferences → AI Assistant Settings` 配置自定义指令和偏好。

**配置 MCP（作为 Client，连接外部 MCP Server）：**

> 需要 IntelliJ IDEA 2025.1+

1. 导航到 `Settings → Tools → AI Assistant → Model Context Protocol (MCP)`
2. 点击 `Add` 添加新的 MCP Server
3. 粘贴 MCP Server 的 JSON 配置

```json
{
  "command": "npx",
  "args": ["-y", "@modelcontextprotocol/server-filesystem", "/path"],
  "env": {}
}
```

**配置 MCP（作为 Server，供外部工具连接）：**

> 需要 IntelliJ IDEA 2025.2+

1. 导航到 `Settings → Tools → MCP Server`
2. 勾选 `Enable MCP Server`
3. 在 `Clients Auto-Configuration` 中选择目标 Client（Claude Desktop、Cursor 等）并点击 `Auto-Configure`

---

### 14. Antigravity（Google DeepMind）

Antigravity 是 Google DeepMind 开发的 AI 编码助手，通常内置于 Gemini 生态系统中。

**配置 Skill：**

Antigravity 支持通过项目级文件提供上下文指令（类似于 `GEMINI.md`）。同时还通过 `skills` 目录提供可扩展的技能：

```
.agents/
└── skills/
    └── my-skill/
        ├── SKILL.md      # 技能指令文件（必需）
        ├── scripts/      # 辅助脚本
        └── examples/     # 参考示例
```

**配置 MCP：**

Antigravity 支持在配置中声明 MCP Server，配置方式与 Gemini 生态一致。通常通过 IDE 设置面板或 JSON 配置文件管理。

---

## 快速对比表

| 工具 | 类型 | Skill 文件 | MCP 配置位置 | MCP 管理方式 |
|------|------|-----------|-------------|-------------|
| **Claude Code** | CLI | `CLAUDE.md` | `~/.claude/`、`.mcp.json` | `claude mcp add/remove/list` |
| **Codex CLI** | CLI | `AGENTS.md` | `~/.codex/config.toml` | `codex mcp add` |
| **Gemini CLI** | CLI | `GEMINI.md` | `~/.gemini/settings.json` | `gemini mcp add` |
| **Qwen Code** | CLI | `QWEN.md` | `~/.qwen/settings.json` | `qwen mcp add/remove` |
| **Aider** | CLI | `.aider.conf.yml` | 作为 MCP Server 供他用 | `pip install aider-mcp` |
| **Augment Code** | CLI | `~/.augment/skills/` | `~/.augment/settings.json` | GUI / JSON |
| **Cursor** | IDE | `.cursor/rules/*.mdc` | `~/.cursor/mcp.json` | GUI / JSON |
| **Windsurf** | IDE | `.windsurfrules` | `~/.codeium/windsurf/mcp_config.json` | GUI / JSON |
| **Trae** | IDE | `.trae/rules/` | Trae 设置面板 | Marketplace / JSON |
| **Kiro** | IDE | `.kiro/powers/` | `.kiro/settings/mcp.json` | 命令面板 / JSON |
| **VS Code Copilot** | IDE 扩展 | `.github/copilot-instructions.md` | `.vscode/mcp.json` | Agent Mode GUI |
| **Cline** | IDE 扩展 | `.clinerules` | VS Code `settings.json` | 内置 Marketplace |
| **JetBrains** | IDE | AI Assistant 设置 | `Settings → MCP` | GUI / JSON |
| **Antigravity** | IDE | `.agents/skills/` | Gemini 生态配置 | IDE 设置 |

---

## 通用 MCP Server 安装参考

以下是一些常用 MCP Server 的安装命令，适用于所有支持 MCP 的工具：

### 文件系统

```bash
npx -y @modelcontextprotocol/server-filesystem /path/to/allowed/dir
```

### GitHub

```bash
npx -y @modelcontextprotocol/server-github
# 需要设置 GITHUB_TOKEN 环境变量
```

### PostgreSQL

```bash
npx -y @modelcontextprotocol/server-postgres postgresql://user:pass@host:5432/db
```

### SQLite

```bash
npx -y @modelcontextprotocol/server-sqlite /path/to/database.db
```

### Brave Search

```bash
npx -y @modelcontextprotocol/server-brave-search
# 需要设置 BRAVE_API_KEY 环境变量
```

### Context7（文档查询）

```bash
npx -y @upstash/context7-mcp
```

### Playwright（浏览器自动化）

```bash
npx -y @anthropic/mcp-playwright
```

### Docker

```bash
npx -y @modelcontextprotocol/server-docker
```

### 更多 MCP Server

- **MCP 官方仓库：** [github.com/modelcontextprotocol/servers](https://github.com/modelcontextprotocol/servers)
- **MCP 市场：** [mcp.so](https://mcp.so)、[mcpmarket.com](https://mcpmarket.com)
- **Glama MCP 列表：** [glama.ai/mcp/servers](https://glama.ai/mcp/servers)

---

## 通用 JSON 配置模板

绝大多数工具的 MCP 配置格式高度统一，以下模板可作为参考：

```json
{
  "mcpServers": {
    "<server-name>": {
      "command": "<executable>",
      "args": ["<arg1>", "<arg2>"],
      "env": {
        "<ENV_VAR>": "<value>"
      }
    }
  }
}
```

**常见 `command` 值：**

| 命令 | 说明 |
|------|------|
| `npx` | Node.js 包执行器（最常用） |
| `uvx` | Python uv 包执行器 |
| `python` | Python 脚本 |
| `docker` | Docker 容器 |
| `node` | Node.js 脚本 |

---

> 💡 **提示：** MCP 生态发展迅速，建议定期关注 [modelcontextprotocol.io](https://modelcontextprotocol.io) 获取最新信息。
