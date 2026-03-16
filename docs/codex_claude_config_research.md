# Codex 与 Claude Code 配置机制调研报告

*日期: 2026-03-16 (v2 — 参考 [cc-switch](https://github.com/farion1231/cc-switch) 源码更新)*

## 1. Codex CLI (`@openai/codex`)

### 配置文件结构

Codex 使用**两个**独立文件存储配置：

| 文件 | 路径 | 格式 | 说明 |
|------|------|------|------|
| **认证** | `~/.codex/auth.json` | JSON | 存储 API Key 令牌 |
| **设置** | `~/.codex/config.toml` | TOML | 存储模型、提供商、推理等级等 |

### config.toml 完整字段

```toml
# ── 顶层字段 ──
model = "gpt-5.2"                       # 当前使用模型
model_reasoning_effort = "medium"       # 推理思考等级: none / minimal / low / medium / high / xhigh
model_provider = "custom-provider"      # 可选: 指定使用哪个自定义提供商（对应下方 model_providers 表名）
base_url = "https://api.openai.com/v1"  # 可选: 顶层 base_url（如果未使用 model_provider）

# ── 自定义提供商 ──
[model_providers."custom-provider"]
name = "Custom Provider"
base_url = "https://api.custom-provider.com/v1"   # 该提供商的 API 地址
env_key = "CUSTOM_API_KEY"                         # 可选: 映射到哪个环境变量读取 Key
wire_api = "responses"                             # 可选: API 协议类型

# ── 多代理（AgentFlow 使用） ──
[features]
multi_agent = true

[agents.reviewer]
description = "Code reviewer agent"
config_file = "agents/reviewer.toml"
```

### 配置优先级

1. **命令行参数**：`codex -c model="gpt-5.2" -c model_reasoning_effort="high"` 最高优先
2. **config.toml**：持久化配置
3. **环境变量**：`OPENAI_API_KEY` / `OPENAI_BASE_URL` 作为默认连接方式
4. **model_provider 机制**：当设置了 `model_provider = "xxx"` 时，Codex 从 `[model_providers.xxx]` 读取 `base_url`，并使用 `env_key` 指定的环境变量作为 API Key

### cc-switch 的多提供商管理方式

cc-switch 为每个提供商生成独立的配置文件对：
- `~/.codex/auth-<provider-name>.json`（认证）
- `~/.codex/config-<provider-name>.toml`（配置）

切换提供商时，将对应文件内容原子性地写回 `auth.json` + `config.toml`。

### AgentFlow 的正确配置方式

- **API Key**：写入 `OPENAI_API_KEY` 环境变量（shell rc），或支持自定义提供商时写入 `[model_providers]` 的 `env_key`
- **Base URL**：写入 `OPENAI_BASE_URL` 环境变量，或写入 `[model_providers]` 的 `base_url`
- **模型**：写入 `config.toml` 顶层 `model` 字段（通过 `WriteCodexConfig`）
- **推理等级**：写入 `config.toml` 顶层 `model_reasoning_effort` 字段

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
| `ANTHROPIC_API_KEY` | API 认证密钥 |
| `ANTHROPIC_BASE_URL` | 自定义 API 端点（网关/代理） |

---

## 3. AgentFlow 配置写入策略总结

| CLI | API Key | Base URL | 模型 | 其他 |
|-----|---------|----------|------|------|
| **Codex** | `OPENAI_API_KEY` (env) | `OPENAI_BASE_URL` (env) | `config.toml` → `model` | `config.toml` → `model_reasoning_effort` |
| **Claude** | `ANTHROPIC_API_KEY` (env) | `ANTHROPIC_BASE_URL` (env) | `~/.claude.json` → `model` | — |

> **设计原则**：环境变量写入 shell rc 文件（`.zshrc` / `.bashrc`）；文件配置使用各 CLI 的原生格式。
