# 知识库服务 (KnowledgeService)

> 管理项目知识库的创建、更新和查询。

## 知识库结构

```
{KB_ROOT}/                          # 默认 .agentflow/
├── INDEX.md                        # 项目概述（技术栈、架构、入口点）
├── context.md                      # 项目上下文（依赖、配置、环境）
├── CHANGELOG.md                    # 变更日志
├── modules/
│   ├── _index.md                   # 模块清单
│   └── {module}.md                 # 各模块文档
├── plan/
│   └── YYYYMMDDHHMM_<feature>/
│       ├── proposal.md             # 方案文档
│       └── tasks.md                # 任务清单
├── sessions/
│   └── {session_id}.md             # 会话摘要
├── graph/                          # AgentFlow 增强
│   ├── nodes.json                  # 知识图谱节点
│   └── edges.json                  # 知识图谱边
├── conventions/                    # AgentFlow 增强
│   └── extracted.json              # 自动提取的编码规范
└── archive/
    ├── _index.md
    └── YYYY-MM/
```

## 操作协议

```yaml
初始化 (~init):
  1. 扫描项目目录结构
  2. 识别技术栈（package.json, pyproject.toml, pom.xml 等）
  3. 生成 INDEX.md 和 context.md
  4. 扫描模块并生成 modules/_index.md
  5. GRAPH_MODE=1: 构建初始知识图谱
  6. CONVENTION_CHECK=1: 提取编码规范

更新:
  - 代码变更后: 更新涉及模块的 modules/{module}.md
  - 新增模块: 创建新的 modules/{module}.md, 更新 _index.md
  - 删除模块: 归档到 archive/, 更新 _index.md
  - GRAPH_MODE=1: 更新图节点和边

CHANGELOG 格式:
  ## [YYYY-MM-DD HH:MM] {feature_name}
  - {change_type}: {description}
  - 涉及文件: {file_list}

写入规则:
  - 目录/文件不存在时自动创建
  - 禁止在 {KB_ROOT}/ 外创建知识库文件
  - 动态目录在首次写入时创建
```
