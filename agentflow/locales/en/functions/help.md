# ~help — command帮助

## 触发

```
~help
~help <command>
```

## 行为

```yaml
~help（no args）:
  1. 列出所有可用的 ~ command及简要说明
  2. Format:
     | command | 说明 |
     |------|------|
     | ~init | initializationknowledge base |
     | ~plan | solution design（支持 list/show/combo） |
     | ~exec | 执行plan packageor直接开发 |
     | ~auto | 全auto-execution（R3standard flow） |
     | ~review | code review |
     | ~status | 查看whenbeforestatus |
     | ~scan | architecture scan |
     | ~conventions | coding conventionsextract与check |
     | ~graph | knowledge graph查询 |
     | ~dashboard | generateitems目仪表板 |
     | ~memory | 记忆manage |
     | ~rlm | RLM sub-agent操作 |
     | ~validatekb | knowledge baseverify |
     | ~help | 显示此帮助信息 |
  3. 末尾提示: "输入 ~help <command名> 查看详细用法"

~help <command>:
  1. read functions/<command>.md 模块文件
  2. 输出该command的详细用法说明
  3. 若commanddoes not exist，提示可用command列表
```

## Output Format

```yaml
图标: 💬
Format: 💬【AgentFlow】- 帮助: {command说明}
```

## Notes

```yaml
- 此command不进入路由判定（R0-R4）
- 不触发 EHRB 检测
- 不Changed任何文件
- 属于只读信息展示
```
