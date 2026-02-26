# ~init 命令

> 初始化项目知识库。

```yaml
触发: 用户输入 ~init
流程:
  1. 扫描项目目录结构
  2. 识别技术栈 (package.json, pyproject.toml, Cargo.toml 等)
  3. 生成 {KB_ROOT}/INDEX.md (项目概述)
  4. 生成 {KB_ROOT}/context.md (技术上下文)
  5. 扫描模块并生成 modules/_index.md + 各模块文档
  6. GRAPH_MODE=1: 构建初始知识图谱 (nodes.json, edges.json)
  7. CONVENTION_CHECK=1: 提取编码规范到 conventions/extracted.json
  8. 输出初始化摘要

complex 级别大型项目:
  - [RLM:原生子代理] 模块扫描并行 (按目录分配)
```
