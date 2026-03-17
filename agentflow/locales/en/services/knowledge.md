# knowledge base服务 (KnowledgeService)

> manageitems目knowledge base的create、updateand查询。

## knowledge base结构

```
{KB_ROOT}/                          # 默认 .agentflow/
├── INDEX.md                        # project overview（tech stack、架构、入口点）
├── context.md                      # items目上下文（dependencies、配置、环境）
├── CHANGELOG.md                    # change log
├── modules/
│   ├── _index.md                   # module list
│   └── {module}.md                 # 各module documentation
├── plan/
│   └── YYYYMMDDHHMM_<feature>/
│       ├── proposal.md             # 方案文档
│       └── tasks.md                # task checklist
├── graph/                          # AgentFlow 增强
│   ├── nodes.json                  # knowledge graph节点
│   └── edges.json                  # knowledge graph边
├── conventions/                    # AgentFlow 增强
│   └── extracted.json              # 自动extract的coding conventions
└── archive/
    ├── _index.md
    └── YYYY-MM/
```

## 操作协议

```yaml
initialization (~init):
  1. scanitems目目录结构
  2. 识别tech stack（package.json, pyproject.toml, pom.xml 等）
  3. generate INDEX.md and context.md
  4. scan模块并generate modules/_index.md
  5. GRAPH_MODE=1: 构建初始knowledge graph
  6. CONVENTION_CHECK=1: extractcoding conventions

update:
  - 代码变更after: updateaffected module modules/{module}.md
  - Added模块: create新的 modules/{module}.md, update _index.md
  - delete模块: archive到 archive/, update _index.md
  - GRAPH_MODE=1: update图节点and边

CHANGELOG Format:
  ## [YYYY-MM-DD HH:MM] {feature_name}
  - {change_type}: {description}
  - files involved: {file_list}

writeRules:
  - 目录/文件does not existwhen自动create
  - 禁止在 {KB_ROOT}/ outsidecreateknowledge base文件
  - 动态目录在首次writewhencreate
```
