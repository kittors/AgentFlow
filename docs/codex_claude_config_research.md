# Codex 与 Claude Code 配置机制调研报告

*日期: 2026-03-16*

## 1. Codex CLI (`@openai/codex`)

通过查阅官方文档并提取 Codex 的配置结构，我重新梳理了它的配置行为。您说得对，我之前的实现（手动写入 `~/.codex/config.json` 并试图将 `OPENAI_API_KEY` 改为 `CODEX_API_KEY`）是完全错误的。

### 官方配置机制：
- **配置文件：** Codex 的本地配置存储为 TOML 格式，路径位于 `~/.codex/config.toml`，而不是 JSON。
- **模型与推理思考等级（Reasoning Setting）：**
  在 `config.toml` 中，模型和思考等级是根节点或配置文件的独立属性：
  ```toml
  model = "gpt-5.2"
  model_reasoning_effort = "medium"  # 可选: none, minimal, low, medium, high, xhigh
  ```
- **环境变量：** 默认情况下，Codex 连接 OpenAI，并会自动读取 `OPENAI_API_KEY` 和 `OPENAI_BASE_URL`。
- **自定义模型提供商：** 如果需要使用自定义的网关和模型（比如 `gpt-5.1-codex-max`），必须在 `config.toml` 里面配置一个提供商（Provider）块，才能映射到您自定义的环境变量：
  ```toml
  [model_providers."custom-provider"]
  name = "Custom Provider"
  base_url = "https://api.custom-provider.com/v1"
  env_key = "CUSTOM_API_KEY"
  ```
- **命令行热覆盖：** Codex 可以通过 `-c` 或 `--config` 参数在启动时动态覆盖配置，例如：`codex -c model="gpt-5.2" -c model_reasoning_effort="high"`。

## 2. Claude Code (`@anthropic-ai/claude-code`)

Claude Code 的配置方式稍有不同。

### 官方配置机制：
- **环境变量：** 默认依赖 `ANTHROPIC_API_KEY`，可附加 `ANTHROPIC_BASE_URL`。但对于**模型**，它**不**直接读取类似 `ANTHROPIC_MODEL` 这样的环境变量来更改默认全局模型。
- **持久化配置：**
  Claude Code 采用 `settings.json` 模式持久化状态。全局配置位于 `~/.claude.json`，项目级配置位于 `.claude/settings.json`。若要强制修改默认使用模型（比如 `claude-opus-4-6`），必须通过写入这些 JSON 配置文件生效。
- **命令行参数：** 或者，可以在每次运行命令时显式传递模型参数：`claude --model claude-opus-4-6`。

## AgentFlow 原有配置错误总结及修复方案
1. AgentFlow 曾试图通过 `__CODEX_REASONING__` 拼装并输出为 `~/.codex/config.json`。**修复：** 必须使用 TOML 序列化写入 `~/.codex/config.toml`，键名为 `model` 和 `model_reasoning_effort`。
2. AgentFlow 试图将环境变量强行输出为 `CODEX_API_KEY`，破坏了 Codex 的原生识别。**修复：** 默认提供 `OPENAI_API_KEY`，或者在生成的 `config.toml` 里的 `[model_providers]` 声明中映射您的自定义 API Key 名称。
3. 对于 Claude Code，只写入 `ANTHROPIC_MODEL` 环境变量是无效的。**修复：** 我们应该专门针对 Claude Code，读取选定的模型后，写入到其标准的配置文件 `~/.claude.json` 或者在调用该命令行工具组时下发。
