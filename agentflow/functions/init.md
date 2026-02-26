# ~init 命令

> 初始化项目知识库。

```yaml
触发: 用户输入 ~init
流程:
  1. 创建 .agentflow/ 目录（如不存在）
  2. 创建 .agentflow/kb/ 目录
  3. 扫描项目目录结构
  4. 识别技术栈 (package.json, pyproject.toml, Cargo.toml 等)
  5. 生成 .agentflow/kb/INDEX.md (项目概述)
  6. 生成 .agentflow/kb/context.md (技术上下文)
  7. 扫描模块并生成 .agentflow/kb/modules/_index.md + 各模块文档
  8. GRAPH_MODE=1: 构建初始知识图谱 (.agentflow/kb/graph/nodes.json, edges.json)
  9. CONVENTION_CHECK=1: 提取编码规范到 .agentflow/kb/conventions/extracted.json
  10. 输出初始化摘要

目录创建顺序:
  .agentflow/
  .agentflow/kb/
  .agentflow/kb/modules/
  .agentflow/kb/plan/
  .agentflow/kb/sessions/
  .agentflow/kb/graph/         # GRAPH_MODE=1 时
  .agentflow/kb/conventions/   # CONVENTION_CHECK=1 时
  .agentflow/kb/archive/
  .agentflow/sessions/

complex 级别大型项目:
  - [RLM:原生子代理] 模块扫描并行 (按目录分配)
```
