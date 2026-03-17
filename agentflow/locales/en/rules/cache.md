# cache规then

```yaml
cacheStrategy:
  模块文件: 同一会话within已read的模块不重复read
  knowledge base: INDEX.md and context.md 在会话开始whenload，after续按需refresh
  图查询: GRAPH_MODE=1 when，图查询结果cache到stage transition

cache失效:
  模块文件: 会话结束
  knowledge base: 检测到文件变更when
  图查询: stage transitionwhen
  版本check: 按 UPDATE_CHECK 小whencache
```
