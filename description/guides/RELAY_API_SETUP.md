# 中转站 API 配置指南

如果你使用的是中转站 API（如 elysiver.h-e.top 等），需要额外配置 `LLM_BASE_URL`。

## 配置方法

在 `.env` 文件中添加以下配置：

```env
# OpenAI API 密钥（从中转站获得）
OPENAI_API_KEY=sk-nv1uaw0V3a7Gya7QOlgTxBdgChiSbJunzQMHvMjkXyMpOG1J

# 中转站地址（关键！）
LLM_BASE_URL=https://elysiver.h-e.top/v1

# 模型 ID（根据中转站支持的模型选择）
LLM_MODEL_ID=gpt-5.2

# 其他可选配置
LLM_TEMPERATURE=0.7
LLM_MAX_TOKENS=2000
LLM_TIMEOUT=30
```

## 关键参数说明

| 参数 | 说明 | 例子 |
|------|------|------|
| `OPENAI_API_KEY` | 从中转站获得的密钥 | `sk-...` |
| `LLM_BASE_URL` | 中转站的 API 端点 | `https://elysiver.h-e.top/v1` |
| `LLM_MODEL_ID` | 中转站支持的模型 | `gpt-5.2` 或 `gpt-4` 等 |

## 工作原理

系统会检查 `LLM_BASE_URL` 是否有值：

```
有 LLM_BASE_URL
    ↓
使用中转站地址: https://elysiver.h-e.top/v1
    ↓
调用中转站的模型 (gpt-5.2)

没有 LLM_BASE_URL
    ↓
使用官方 OpenAI 地址
    ↓
调用官方的模型
```

## 测试配置

### 方法 1: 使用测试脚本

```bash
cd D:\TrueSignal
python test_agent_llm.py
```

这个脚本会：
- 显示当前配置
- 发送 3 个测试消息到 LLM
- 显示 LLM 的回复
- 验证配置是否正确

### 方法 2: 在前端测试

1. 启动所有服务: `start-all.bat`
2. 打开 http://localhost:5173
3. 向 Agent 提问
4. 查看 Python Backend 窗口的日志

日志中应该显示：
```
Using custom LLM base_url: https://elysiver.h-e.top/v1
```

## 常见问题

### Q: 中转站 API 超时了
**A:**
- 增加 `LLM_TIMEOUT` 值: `LLM_TIMEOUT=60`
- 检查中转站是否正常运行
- 检查网络连接

### Q: 返回 "Invalid model" 错误
**A:**
- 确认 `LLM_MODEL_ID` 是中转站支持的模型
- 与中转站文档对比，使用正确的模型名称
- 常见的模型: `gpt-4`, `gpt-3.5-turbo`, `gpt-5.2` 等

### Q: 返回 "Invalid API key" 错误
**A:**
- 确认 API Key 是从中转站获得的
- Key 应该以 `sk-` 开头
- 确保没有多余的空格

### Q: 中转站服务中断，还能用吗？
**A:**
- 系统会自动降级到规则匹配
- Chat 功能继续可用，但用的是模板回复
- 日志会显示 LLM 错误信息

## 完整 .env 例子

```env
# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=truesignal
DB_PASSWORD=truesignal123
DB_NAME=truesignal

# Redis 配置
REDIS_URL=redis://localhost:6379/0

# 日志配置
LOG_LEVEL=INFO

# ========== LLM 配置（中转站） ==========
OPENAI_API_KEY=sk-nv1uaw0V3a7Gya7QOlgTxBdgChiSbJunzQMHvMjkXyMpOG1J
LLM_BASE_URL=https://elysiver.h-e.top/v1
LLM_MODEL_ID=gpt-5.2
LLM_TEMPERATURE=0.7
LLM_MAX_TOKENS=2000
LLM_TIMEOUT=30

# Go 后端配置
GO_SERVER_PORT=8080
```

## 优势

- ✅ 更便宜（中转站通常有折扣）
- ✅ 更快（本地中转，延迟低）
- ✅ 支持更多模型（如 gpt-5.2 等新模型）
- ✅ 可以切换不同中转站而无需改代码

## 支持的中转站

我们的代码支持任何兼容 OpenAI API 格式的中转站，包括：

- elysiver.h-e.top
- api.openai.com（官方）
- 其他兼容 OpenAI 的中转站

只需修改 `LLM_BASE_URL` 即可切换。

