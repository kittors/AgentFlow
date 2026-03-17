# AgentFlow Plan Templates

## proposal.md Template

```markdown
# 方案: {feature_name}

## 概述
{proposal_summary}

## 目标
- {goal_1}
- {goal_2}

## 方案描述
{detailed_description}

## 影响范围
- 涉及模块: {modules}
- 涉及文件: {file_count} 个
- 预估工作量: {effort}

## 技术方案
{technical_approach}

## 风险与缓解
| 风险 | 等级 | 缓解措施 |
|------|------|----------|
| {risk_1} | {level} | {mitigation} |

## 验证计划
- [ ] {test_1}
- [ ] {test_2}
```

## tasks.md Template

```markdown
# 任务清单: {feature_name}

方案: {proposal_ref}
创建时间: {created_at}

## 阶段一: {phase_1_name}
- [ ] {task_1}
- [ ] {task_2}

## 阶段二: {phase_2_name}
- [ ] {task_3}
- [ ] {task_4}

## 验收条件
- [ ] 代码编译/lint 通过
- [ ] 测试全部通过
- [ ] 知识库已同步
```

## Multi-Proposal Comparison Template (R3+)

```markdown
# 方案对比: {feature_name}

## 方案 A: {name_a}
- 思路: {approach_a}
- 优势: {pros_a}
- 劣势: {cons_a}
- 复杂度: {complexity_a}

## 方案 B: {name_b}
- 思路: {approach_b}
- 优势: {pros_b}
- 劣势: {cons_b}
- 复杂度: {complexity_b}

## 推荐
推荐方案 {recommended}，理由: {reason}
```
