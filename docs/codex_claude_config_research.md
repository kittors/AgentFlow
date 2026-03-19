# Codex 与 Claude Code 配置机制调研报告

*日期: 2026-03-17 (v3 — 深入分析 [cc-switch](https://github.com/farion1231/cc-switch) 源码 `codex_config.rs`)*

## 1. Codex CLI (`@openai/codex`)

### 配置文件结构

Codex 使用**两个**独立文件存储配置：

| 文件 | 路径 | 格式 | 说明 |
|------|------|------|------|
| **认证** | `~/.codex/auth.json` | JSON | 存储 API Key 令牌（`{"token": "sk-xxx"}`） |
| **设置** | `~/.codex/config.toml` | TOML | 存储模型、提供商、推理等级等 |

> **重要**：`auth.json` 中的 `token` 字段就是 API Key，适用于所有提供商（官方和第三方），不是 OAuth token。cc-switch 在切换提供商时，将 `auth-<provider>.json` 的内容**原子性写回** `auth.json`，证明每个提供商都有自己的 API Key 存储在 auth.json 格式中。

### config.toml 完整字段

```toml
# ── 顶层字段 ──
model = "gpt-5.2"                       # 当前使用模型
model_reasoning_effort = "medium"       # 推理思考等级: none / minimal / low / medium / high / xhigh
model_provider = "custom-provider"      # 可选: 指定使用哪个自定义提供商（对应下方 model_providers 表名）

# ── 自定义提供商 ──
[model_providers."custom-provider"]
name = "Custom Provider"
base_url = "https://api.custom-provider.com/v1"   # 该提供商的 API 地址
env_key = "CUSTOM_API_KEY"                         # 可选: 映射到哪个环境变量读取 Key（不设则用 auth.json）
wire_api = "responses"                             # 可选: API 协议类型

# ── 多代理（AgentFlow 使用） ──
[features]
multi_agent = true

[agents.reviewer]
description = "Code reviewer agent"
config_file = "agents/reviewer.toml"
```

### 认证机制详解（基于 cc-switch 源码分析）

cc-switch 的 `codex_config.rs` 揭示了 Codex 的完整认证逻辑：

1. **API Key 存储在 `auth.json`**：每个提供商对应 `auth-<provider>.json`，切换时通过 `write_codex_live_atomic()` 原子性写回 `auth.json` + `config.toml`
2. **`env_key` 是可选的**：cc-switch 测试用例中，部分 `[model_providers]` section 有 `env_key`，部分没有。当没有 `env_key` 时，Codex 应从 `auth.json` 读取 token
3. **`base_url` 写入位置**：
   - 如果存在 `model_provider`，写入 `[model_providers.xxx].base_url`
   - 如果不存在 `model_provider`，写入顶层 `base_url`（注意：不是 `openai_base_url`）

### 配置优先级

1. **命令行参数**：`codex -c model="gpt-5.2" -c model_reasoning_effort="high"` 最高优先
2. **config.toml**：持久化配置
3. **环境变量**：`OPENAI_API_KEY` / `OPENAI_BASE_URL` 作为默认连接方式
4. **model_provider 机制**：当设置了 `model_provider = "xxx"` 时，Codex 从 `[model_providers.xxx]` 读取 `base_url`；如果同时设置了 `env_key`，从对应环境变量读取 API Key；如果没有 `env_key`，从 `auth.json` 读取

### cc-switch 的多提供商管理方式

cc-switch 为每个提供商生成独立的配置文件对：
- `~/.codex/auth-<provider-name>.json`（认证）
- `~/.codex/config-<provider-name>.toml`（配置）

切换提供商时，将对应文件内容原子性地写回 `auth.json` + `config.toml`。

### AgentFlow 的正确配置方式

- **API Key**：写入 `~/.codex/auth.json`（`{"token": "sk-xxx"}`）
- **Base URL**：写入 `config.toml` 的 `[model_providers.agentflow].base_url`（通过 `model_provider` 机制）
- **模型**：写入 `config.toml` 顶层 `model` 字段
- **推理等级**：写入 `config.toml` 顶层 `model_reasoning_effort` 字段
- **不设置 `env_key`**：让 Codex 从 `auth.json` 读取 API Key，无需环境变量

> **跨平台优势**：不依赖环境变量（shell rc），只写文件，Windows/macOS/Linux 行为一致。

---

## 2. Claude Code (`@anthropic-ai/claude-code`)

### 安装方式

```bash
npm install -g @anthropic-ai/claude-code
```

安装后首次运行 `claude` 命令会启动**交互式向导**，完成认证和初始配置。Claude Code **不需要** AgentFlow 替它完成配置——用户可选择跳过自定义配置，使用 `claude` 自带的官方流程。

### 配置文件结构

| 文件 | 路径 | 格式 | 说明 |
|------|------|------|------|
| **全局设置** | `~/.claude.json` | JSON | 存储 `model` 等全局属性 |
| **CLI 设置** | `~/.claude/settings.json` | JSON | 存储 hooks、权限等 |
| **项目设置** | `.claude/settings.json` | JSON | 项目级配置覆盖 |

### 模型配置

Claude Code **不支持**通过 `ANTHROPIC_MODEL` 环境变量设置默认模型。修改默认模型需要：

1. **写入 `~/.claude.json`**：`{ "model": "claude-opus-4-6" }`（AgentFlow 采用此方式）
2. **命令行参数**：`claude --model claude-opus-4-6`

### 环境变量

| 变量 | 用途 |
|------|------|
| `ANTHROPIC_API_KEY` | API 认证密钥（直连与网关模式均使用） |
| `ANTHROPIC_BASE_URL` | 自定义 API 端点（网关/代理） |

> **注意**：`ANTHROPIC_AUTH_TOKEN` 是 Claude Code 内部 OAuth 流程使用的变量，不应用于第三方 API Key。无论是直连还是网关模式，都应使用 `ANTHROPIC_API_KEY`。

---

## 3. AgentFlow 配置写入策略总结

| CLI | API Key | Base URL | 模型 | 其他 |
|-----|---------|----------|------|------|
| **Codex** | `auth.json` → `token` | `config.toml` → `[model_providers.agentflow].base_url` | `config.toml` → `model` | `config.toml` → `model_reasoning_effort` |
| **Claude（直连）** | `ANTHROPIC_API_KEY` (env) | — | `~/.claude.json` → `model` | — |
| **Claude（网关）** | `ANTHROPIC_API_KEY` (env) | `ANTHROPIC_BASE_URL` (env) | `~/.claude.json` → `model` | — |

> **设计原则**：Codex 使用 `auth.json` + `model_provider` 机制（无 `env_key`），完全通过文件配置，跨平台兼容。Claude 使用环境变量写入 shell rc 文件（`.zshrc` / `.bashrc`）；无论是否配置 `ANTHROPIC_BASE_URL`，都统一使用 `ANTHROPIC_API_KEY`（不使用 `ANTHROPIC_AUTH_TOKEN`，因为后者是 Claude Code 内部 OAuth 机制专用）。
