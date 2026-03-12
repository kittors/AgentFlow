# Proposal: main 自动发布 Release

## 目标

- 将 GitHub Release 工作流从“仅 tag 触发”改为“push 到 `main` 自动发布”。
- 保持 `curl -fsSL .../install.sh | bash` 与 `agentflow update` 都能拿到最新 `main` 构建。
- 避免每次 push `main` 都生成新的 semver tag，降低版本管理混乱。

## 推荐方案

- 将 Release workflow 触发条件改为 `push` 到 `main`。
- 固定发布到一个持续更新的 Release/tag：
  - tag: `continuous`
  - release name: `continuous`
- 每次 main push 时：
  - 先删除/覆盖旧的 `continuous` tag
  - 重新用当前 `main` commit 创建 `continuous` tag
  - 重建 release 资产并覆盖到 `continuous` release
- Go 二进制版本号改为：
  - `package.json` 的基础版本 + `-main.<shortsha>`
  - 例如 `1.0.3-main.1ce4a1b`

## 原因

- 安装脚本和自更新目前都依赖 GitHub `releases/latest`。
- 如果改成“每次 main push 都创建新 semver release/tag”，会让 semver 失真。
- 使用固定的 `continuous` 发布位点可以让：
  - latest release 始终指向 `main`
  - 下载链接和资产名保持稳定
  - 版本号仍包含 commit 信息，便于回溯

## 影响范围

- `.github/workflows/release.yml`
- `internal/update/update.go`
- 可能涉及 `internal/update/update_test.go`
- 可能涉及安装文档与版本说明

## 风险

- `continuous` 不是 semver，需确认 update 检查逻辑不会误判。
- 覆盖式 release/tag 需要足够的 workflow 权限和稳定的清理顺序。
- `latest release` 将不再等于手工 semver release，而是等于 `main` 最新构建。

## 验收

- push 到 `main` 自动触发 `Release` workflow。
- 新 release 资产可被安装脚本解析并下载。
- `agentflow version` 显示 `x.y.z-main.<sha>` 风格版本。
- `agentflow update` 对 `continuous` 版本比较不出错。
