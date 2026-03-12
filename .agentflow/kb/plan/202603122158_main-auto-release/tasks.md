# Tasks: main 自动发布 Release

- [x] 调整 Release workflow 触发条件与持续发布策略
- [x] 实现 `continuous` tag/release 的覆盖式发布
- [x] 调整构建版本注入，支持 `x.y.z-main.<sha>`
- [x] 修正 update 检查逻辑，兼容非纯 semver 的 continuous 版本
- [x] 本地验证 workflow 相关脚本与 Go 测试
- [x] 推送到 `main`
- [x] 验证 GitHub CI 与 Release 结果
