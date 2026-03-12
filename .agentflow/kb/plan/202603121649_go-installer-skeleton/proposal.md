# Go 安装骨架方案

## 目标

在以下受限路径内建立可编译的 Go 安装基础设施骨架：

- `embed.go`
- `internal/config/**`
- `internal/install/**`

本次实现优先保留 Python installer 的关键语义：

- AgentFlow marker 判定
- 备份文件命名
- safe write（临时文件 + 原子替换）
- rename-aside（删除/覆盖失败时改名旁路）
- target 安装入口的基本结构

## 设计

### 资源嵌入

- 在根包新增 `embed.go`
- 通过 `embed.FS` 打包 `AGENTS.md`、`SKILL.md` 以及 `agentflow/` 目录资源
- 暴露只读访问器，供安装器读取源资源

### 配置层

- 在 `internal/config` 定义：
  - marker、插件目录名、默认 profile
  - CLI target 元数据
  - 备份文件命名辅助函数
  - 是否为 AgentFlow 文件的判定辅助函数

### 安装层

- 在 `internal/install` 定义：
  - `Installer` 与 `TargetInstaller` 基础结构
  - `safeWrite`、`removeOrRenameAside`、`backupIfNeeded`
  - rules file / module dir / skill file 的最小安装入口
- 先不补齐所有 Python 目标特性，只保证结构可扩展、代码可编译

## 验收

- `go test ./...` 可通过
- 关键语义具备独立工具函数
- 安装入口可对单 target 执行最小部署

## 非目标

- 不在本轮补齐所有 CLI 特定 hooks / agents / uninstall 细节
- 不改动限定范围以外的项目源码
