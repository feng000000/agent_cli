# My Agent
使用 golang 实现的 agent cli


## TODO

- [ ] 消息输入 在 server 层维护 队列, 工具执行时/等LLM响应时 可以暂存 message(支持取消), 可以在 拿到工具执行结果 后直接消费该条消息, 直接发送给大模型, 少一次 大模型的请求
- [ ] 协议支持
  - [ ] 支持 ACP 协议 (client server 分离)
  - [ ] 支持 mcp 客户端
  - [ ] 支持 mcp 服务端
- [ ] 输入类型
  - [x] 文本
  - [ ] 图片(文件)

- [ ] Reflection: 生成后自评阶段
- [ ] sub agent: 用 tool 实现

- [x] 工具调用
  - [ ] 文件系统基本工具
  - [ ] 写入长期记忆
  - [ ] 上下文压缩
- [ ] 自动上下文压缩
- [ ] 完善 simple TUI
- [ ] 工具沙箱执行 (landlock + seccomp)
- [ ] 长工具输出落盘, 清理 ctx.MessageParams


---
- [ ] 用 zap 输出日志
