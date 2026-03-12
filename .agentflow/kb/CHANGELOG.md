# Knowledge Changelog

## 2026-03-12

- 建立 Go 版资源内嵌入口 `embed.go`
- 在 `internal/config` 增加 marker、target、备份/rename-aside 命名辅助
- 在 `internal/install` 增加安装器骨架、安全写入、备份与目录复制逻辑
- 新增基础测试，`go test ./...` 通过
- Release 工作流改为 `main` push 自动发布 continuous release
- `internal/update` 现在优先读取 release `name` 作为版本号，兼容固定 `continuous` tag
- README / README_CN 已补充“`main` 自动 continuous release”说明
