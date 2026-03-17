# ~validatekb command

> AgentFlow 工作流command。

```yaml
Trigger: 用户输入 ~validatekb
flow: verifyknowledge base一致性
  check:
    - module documentation与实际代码一致
    - 索引文件完整
    - 图节点与文件对应
  knowledge base文件 > 10: [RLM:原生sub-agent] parallelverify
```
