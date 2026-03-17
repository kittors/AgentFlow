# ~help — 命令帮助

## 触发

```
~help
~help <command>
```

## 行为

```yaml
~help（无参数）:
  1. 列出所有可用的 ~ 命令及简要说明
  2. 格式:
     | 命令 | 说明 |
     |------|------|
     | ~init | 初始化知识库 |
     | ~plan | 方案设计（支持 list/show/combo） |
     | ~exec | 执行方案包或直接开发 |
     | ~auto | 全自动执行（R3标准流程） |
     | ~review | 代码审查 |
     | ~status | 查看当前状态 |
     | ~scan | 架构扫描 |
     | ~conventions | 编码规范提取与检查 |
     | ~graph | 知识图谱查询 |
     | ~dashboard | 生成项目仪表板 |
     | ~memory | 记忆管理 |
     | ~rlm | RLM 子代理操作 |
     | ~validatekb | 知识库验证 |
     | ~help | 显示此帮助信息 |
  3. 末尾提示: "输入 ~help <命令名> 查看详细用法"

~help <command>:
  1. 读取 functions/<command>.md 模块文件
  2. 输出该命令的详细用法说明
  3. 若命令不存在，提示可用命令列表
```

## 输出格式

```yaml
图标: 💬
格式: 💬【AgentFlow】- 帮助: {命令说明}
```

## 注意事项

```yaml
- 此命令不进入路由判定（R0-R4）
- 不触发 EHRB 检测
- 不修改任何文件
- 属于只读信息展示
```
