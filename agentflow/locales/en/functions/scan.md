# ~scan command

> AgentFlow 工作流command。

```yaml
Trigger: 用户输入 ~scan
flow: scanitems目发现潜在问题（AgentFlow 增强功能）
  检测items:
    - 大文件 (> 500行)
    - 循环dependencies
    - 缺失测试
    - 未use的导入
    - 安全漏洞模式
  Output: scan报告，writeknowledge base
```
