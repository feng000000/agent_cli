# My Agent

使用 golang 实现的 agent cli


## TODO
- [x] loop 框架
- [x] LLM 调用 (deepseek)
- [x] 工具调用
  - [ ] 文件系统基本工具实现
  - [ ] 写入长期记忆
  - [ ] 长工具输出落盘, 清理 ctx.MessageParams
  - [ ] 工具沙箱执行
- [ ] 系统提示词模板
  - [ ] 读取长期记忆
- [ ] 命令调用
- [ ] 上下文压缩
---
- [ ] 用 zap 输出日志
- [ ] cli 显示优化
