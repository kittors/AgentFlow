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
  .agentflow/kb/graph/         # GRAPH_MODE=1 时
  .agentflow/kb/conventions/   # CONVENTION_CHECK=1 时
  .agentflow/kb/archive/
  .agentflow/sessions/         # 会话记录（不在 kb/ 下）

complex 级别大型项目:
  - [RLM:原生子代理] 模块扫描并行 (按目录分配)
```

<init_execution_protocol>

## 初始化执行协议（CRITICAL — 必须完整执行所有步骤）

> **核心原则**: ~init 的目的是让 .agentflow/ 目录从空变为包含有实际内容的知识库。
> 空目录 = 初始化失败。每个子目录都必须包含实际文件。

### Phase 1: 目录结构创建

```bash
# 必须首先执行——创建所有目录
mkdir -p .agentflow/kb/modules .agentflow/kb/plan \
         .agentflow/kb/graph .agentflow/kb/conventions .agentflow/kb/archive \
         .agentflow/sessions
```

### Phase 2: Go 命令调用（MUST — 这些命令负责生成实际内容）

> 以下命令使用 AgentFlow 自带的 Go CLI 生成知识库内容。
> 如果 `agentflow ...` 不可用，则手动执行等效操作。

```yaml
步骤 A: 模板初始化（创建基础知识库结构）
  命令: agentflow init --quiet
  产出: .agentflow/kb/ 下的模板参考文件、所有子目录
  失败降级: 手动创建 INDEX.md 和 context.md

步骤 B: 模块扫描与 KB 同步（填充 modules/ 目录）
  命令: agentflow kb sync --quiet
  产出: .agentflow/kb/modules/_index.md + 各模块的 {module}.md
  失败降级: 手动扫描项目源码目录，生成模块清单

步骤 C: 编码规范提取（CONVENTION_CHECK=1 时，填充 conventions/ 目录）
  命令: agentflow conventions --quiet
  产出: .agentflow/kb/conventions/extracted.json
  失败降级: 手动分析代码风格，生成 JSON 格式的规范文档

步骤 D: 知识图谱初始化（GRAPH_MODE=1 时，填充 graph/ 目录）
  命令: agentflow graph --quiet
  产出: .agentflow/kb/graph/nodes.json, edges.json
  失败降级: 手动创建初始节点和边数据
```

### Phase 3: 手动补充内容（辅助脚本无法覆盖的部分）

```yaml
步骤 E: 生成 INDEX.md（项目概述）
  - 如果 Phase 2 步骤 A 未生成，手动创建
  - 内容: 项目名称、技术栈、架构概述、入口点

步骤 F: 生成 context.md（技术上下文）
  - 内容: 依赖清单、构建工具、运行环境、配置文件

步骤 G: 创建初始会话记录
  命令: agentflow session save --quiet --stage=INIT
  失败降级: 手动创建 .agentflow/sessions/{timestamp}.md
```

### Phase 4: 验证检查点（MUST — 初始化完成前必须通过）

```yaml
验证命令: ls -la .agentflow/kb/modules/ .agentflow/kb/conventions/ .agentflow/sessions/

验证标准:
  ✅ .agentflow/kb/INDEX.md 存在且非空
  ✅ .agentflow/kb/context.md 存在且非空
  ✅ .agentflow/kb/modules/_index.md 存在且非空
  ✅ .agentflow/kb/modules/ 下至少有一个模块文档
  ✅ CONVENTION_CHECK=1 时: .agentflow/kb/conventions/extracted.json 存在
  ✅ GRAPH_MODE=1 时: .agentflow/kb/graph/nodes.json 存在

未通过验证:
  - 检查哪个步骤失败
  - 使用降级方案手动创建缺失文件
  - 重新验证

禁止行为:
  - 不执行 Phase 2 的辅助脚本就直接报告初始化完成
  - 创建空目录后就报告成功
  - 跳过 Phase 4 的验证
```

</init_execution_protocol>
