# Codex 配置全链路修复计划

*日期: 2026-03-17*

## 背景

cc-switch 的 Codex 配置方案依赖**本地代理服务器**来注入 API Key（`forwarder.rs` + `handlers.rs`）。AgentFlow 不运行本地代理，因此 Codex 直接连接第三方 API 时必须通过 `model_provider` + `env_key` + **环境变量**来提供 API Key。

## 当前问题

| 问题 | 原因 |
|------|------|
| Codex 报 "Token data is not available" | `[model_providers.agentflow]` 没有 `env_key`，且 `OPENAI_API_KEY` 环境变量未设置 |
| auth.json 中的 token 没被使用 | `auth.json` 的 token 在有 `model_provider` 时不被 Codex 使用；只有无 `model_provider` 时才使用 |

## 修复方案

### 1. `WriteCodexConfig`（install.go）

**改动**：在 `[model_providers.agentflow]` section 中**加上** `env_key = "OPENAI_API_KEY"`

```toml
model_provider = "agentflow"

[model_providers.agentflow]
name = "agentflow"
base_url = "https://cliproxy.07230805.xyz/v1"
env_key = "OPENAI_API_KEY"      # ← 新增
wire_api = "responses"
```

**保留**：auth.json 仍然写入 API Key（备份 + 兼容）

---

### 2. `writeEnvConfigPanel`（panels_config.go）

**改动**：
- `OPENAI_API_KEY` 从 `codexAPIKey`（仅 auth.json）路由**回** `normalEnvVars`，这样它会同时：
  - 写入 `.zshrc`（通过 `WriteEnvConfig`）
  - 调用 `os.Setenv` 使当前进程立即生效
- 仍然传给 `WriteCodexConfig` 写入 auth.json（双保险）

**UI 显示**：配置完成后显示 "已写入 ~/.zshrc: OPENAI_API_KEY=sk-***xxx"

---

### 3. `cleanCodexBootstrapConfig`（install.go）

**已正确**：清理 `model_provider` + `[model_providers.agentflow]` section，不删 auth.json。

---

### 4. `CleanEnvConfig`（bootstrap.go）

**已正确**：卸载时自动清理 `.zshrc` 中所有 `# AgentFlow CLI configuration` 块下的 `export` 行。

---

### 5. `UninstallCLI`（cli_uninstall.go）

**已正确**：卸载后调用 `CleanEnvConfig()` 清理 shell RC。

---

### 6. `targets.go` 注释

**改动**：更新 Codex target 注释，说明 `env_key + OPENAI_API_KEY` 环境变量方案。

---

### 7. `codex_claude_config_research.md`

**改动**：更新文档反映最终方案（model_provider + env_key + shell RC）。

---

## 测试清单

- [ ] `WriteCodexConfig` 生成包含 `env_key` 的 config.toml
- [ ] `writeEnvConfigPanel` 将 `OPENAI_API_KEY` 写入 `.zshrc`
- [ ] 新终端中 `codex` 可正常工作
- [ ] 全局安装规则不影响 Codex 配置
- [ ] 项目级安装规则不影响 Codex 配置
- [ ] MCP 安装/卸载不影响 Codex 配置
- [ ] Skill 安装不影响 Codex 配置
- [ ] AgentFlow 规则卸载清理 config.toml 中的 AgentFlow 字段
- [ ] CLI 卸载清理 `.zshrc` 中的环境变量
- [ ] CLI 完整卸载（PurgeConfigDir）删除 `~/.codex/` 目录

## 不受影响的链路

以下路径**不涉及 Codex 配置**，无需修改：

- MCP 安装/卸载：只操作 `config.toml` 的 `[mcp_servers]` section
- Skill 安装：只操作 `.agents/skills/` 目录
- 全局/项目级规则安装：只操作 `AGENTS.md` / `.agents/` 和 Codex `config.toml` 的 `[features]` section
- 全局/项目级规则卸载：`cleanCodexBootstrapConfig` 只清理 AgentFlow 写入的字段
