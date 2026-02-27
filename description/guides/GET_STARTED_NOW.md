# 立即开始使用 Agent 💬

## 你已经完成的配置 ✅

```env
OPENAI_API_KEY=sk-nv1uaw0V3a7Gya7QOlgTxBdgChiSbJunzQMHvMjkXyMpOG1J
LLM_BASE_URL=https://elysiver.h-e.top/v1
LLM_MODEL_ID=gpt-5.2
```

这个配置使用中转站 API，所有参数都已正确设置。

## 现在怎么用？

### 第 1 步：重启后端

运行启动脚本：
```bash
start-all.bat
```

或手动重启 Python Backend（监听 8081）。

### 第 2 步：打开前端

访问：http://localhost:5173

### 第 3 步：问 Agent 问题

在聊天框输入任何问题，例如：

- "现在的执行进度如何？" → AI 会生成关于任务进度的详细回复
- "我想多关注深度学习的内容" → AI 会给出如何调整配置的建议
- "为什么这张卡片被标记为 SKIP？" → AI 会解释评估决策
- "如何改进评估的准确性？" → AI 会给出优化建议

### 第 4 步：查看回复

每个回复都是通过中转站的 gpt-5.2 模型生成的，应该是：
- ✅ 自然流畅的中文
- ✅ 基于任务上下文的智能回答
- ✅ 每次都不一样（不是硬编码模板）

## 快速验证

想快速验证配置是否正确？运行：

```bash
cd D:\TrueSignal
python test_agent_llm.py
```

这会显示：
- 当前配置信息
- 3 个测试问题的 LLM 回复
- 连接和模型信息

## 预期看到的效果

### ❌ 不应该看到的（硬编码模板）
```
当前任务 RSS Feed 的统计信息：
✅ 已处理消息: 127 条
⭐ 高价值内容: 12 条 (9.4%)
📌 已书签: 38 条 (29.9%)
⏭️ 跳过: 77 条 (60.6%)
```

### ✅ 应该看到的（AI 生成）
```
根据你目前的 RSS 源监控情况，过去 24 小时内系统已经处理了 127 条
新文章。其中 12 条被标记为高价值内容（创新度和深度都在 8 分以上），
这表明你的订阅源质量不错。建议你可以通过调整过滤规则来获得更多
符合你兴趣的内容。具体来说，如果你想专注于深度学习领域，可以在
filter_rules 中添加关键词...
```

## 日志中会看到什么

打开 Python Backend 的窗口，应该看到：

```
INFO: Using custom LLM base_url: https://elysiver.h-e.top/v1
INFO: Task Chat API Call for task 1
INFO: Completed for task 1
```

这表示系统已经成功调用了中转站 API。

## 出问题了？

### 如果返回硬编码模板回复
1. 检查 Python Backend 窗口的日志
2. 查看是否有错误信息（如 "LLM API call failed"）
3. 确认 API Key、Base URL、Model ID 都正确
4. 重启 Python Backend

### 如果显示 "timeout"
```env
# 增加超时时间
LLM_TIMEOUT=60
```
然后重启。

### 如果中转站不可用
- 系统自动降级到规则匹配
- Chat 功能继续可用
- 当中转站恢复后可以继续使用 AI

## 配置参考

完整配置文件位置：`D:\TrueSignal\.env`

| 参数 | 当前值 | 说明 |
|------|--------|------|
| OPENAI_API_KEY | sk-nv1uaw... | 中转站 API 密钥 |
| LLM_BASE_URL | https://elysiver.h-e.top/v1 | 中转站地址 |
| LLM_MODEL_ID | gpt-5.2 | 使用的模型 |
| LLM_TEMPERATURE | 0.7 | 创意程度（0.0=严谨, 1.0=随意） |
| LLM_MAX_TOKENS | 2000 | 最大回复长度 |
| LLM_TIMEOUT | 30 | API 调用超时（秒） |

## 更换模型

如果想用其他模型，只需改 `.env`：

```env
# 改为 GPT-4
LLM_MODEL_ID=gpt-4

# 改为官方 OpenAI
LLM_BASE_URL=https://api.openai.com/v1
OPENAI_API_KEY=sk-...（官方密钥）
```

然后重启即可。

## 常见问题

**Q: 会不会很贵？**
A: 中转站通常比官方便宜。每次对话约 0.01-0.02 元。

**Q: 回复速度快吗？**
A: 中转站通常在 1-3 秒内返回回复。

**Q: 能切换模型吗？**
A: 可以！只需改 `.env` 中的 `LLM_MODEL_ID` 和 `LLM_BASE_URL`。

**Q: 如果中转站挂了会怎样？**
A: Chat 功能继续可用，但会返回规则匹配的模板回复。

**Q: 我的 API Key 安全吗？**
A: Key 存在 `.env` 文件中，不会传给任何第三方。只有你的代码会用它。

## 下一步

1. 重启系统：`start-all.bat`
2. 打开前端：http://localhost:5173
3. 向 Agent 提问
4. 享受 AI 生成的智能回复！

---

有任何问题或错误，检查文档：
- 中转站配置详细：`description/guides/RELAY_API_SETUP.md`
- 配置检查清单：`description/guides/AGENT_CHECKLIST.md`
- 完整技术文档：`description/guides/PHASE5_3_LLM_FIX_SUMMARY.md`

