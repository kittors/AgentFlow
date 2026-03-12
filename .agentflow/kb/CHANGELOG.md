# Knowledge Changelog

## 2026-03-12

- 建立 Go 版资源内嵌入口 `embed.go`
- 在 `internal/config` 增加 marker、target、备份/rename-aside 命名辅助
- 在 `internal/install` 增加安装器骨架、安全写入、备份与目录复制逻辑
- 新增基础测试，`go test ./...` 通过
