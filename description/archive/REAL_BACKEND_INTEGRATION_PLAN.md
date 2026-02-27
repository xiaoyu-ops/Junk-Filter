# 真实后端接入方案 - 完整评估和实施计划

**制定日期**: 2026-02-27
**目标**: 替换 Mock 后端，接入真实 Go + Python + 真实 RSS + LLM API
**难度**: ⭐⭐⭐⭐ (高) | **工作量**: 3-4 周 | **优先级**: ⭐⭐⭐⭐⭐

---

## 📋 当前状态评估

### 前端现状
- ✅ Go 后端已连接 (http://localhost:8080)
  - RSS 源管理 API
  - 内容查询 API
  - 评估结果 API

- ⚠️ Mock 后端仍在使用 (http://localhost:3000)
  - 消息 API
  - SSE 聊天流
  - 执行历史

### 后端现状

**Go 后端**
- ✅ 已实现
  - 主框架 (Gin)
  - 数据库连接 (PostgreSQL)
  - Redis 连接
  - 4 个 Handler（source, content, evaluation, junk_filter）
  - 6+ API 端点

- ⚠️ 缺失
  - 消息管理 API (POST/GET /api/tasks/{id}/messages)
  - 执行历史 API
  - SSE 聊天流端点 (/api/chat/stream)
  - Config 保存和加载 API

**Python 后端**
- ✅ 已实现
  - LangGraph 评估引擎
  - Redis Stream 消费
  - 异步数据库操作
  - 配置管理

- ⚠️ 缺失
  - HTTP API 服务器（当前只有异步消费者）
  - 没有暴露 REST 接口
  - 需要与前端集成

---

## 🎯 核心问题分析

### 问题 1: Python 后端没有 HTTP 服务
当前 Python 后端只是消费者，没有 HTTP API 层。

**解决方案**:
```
选项 A: 添加 FastAPI 或 Flask HTTP 层
  ├─ 添加 /api/evaluate (文本评估接口)
  └─ 时间: 3-5 天

选项 B: 让 Go 后端调用 Python 评估服务
  ├─ Go 通过 HTTP/gRPC 调用 Python
  └─ 时间: 2-3 天

推荐: 选项 B (更简洁)
```

### 问题 2: 消息存储在 Mock 后端
消息 API 暂未实现在真实后端。

**解决方案**:
```
在 Go 后端添加消息表和 API
├─ Schema: messages 表 (task_id, role, content, etc)
├─ Handler: POST /api/tasks/{id}/messages
├─ Handler: GET /api/tasks/{id}/messages
└─ 时间: 2-3 天
```

### 问题 3: SSE 聊天流未实现
当前 Mock 后端提供 SSE，真实后端未实现。

**解决方案**:
```
在 Go 后端实现 SSE 流端点
├─ GET /api/chat/stream?taskId={id}&message={msg}
├─ 调用 Python 评估服务
├─ 流式返回 AI 回复
└─ 时间: 2-3 天
```

### 问题 4: 真实 LLM API 未配置
需要接入真实的 OpenAI / Claude / 其他 LLM API。

**解决方案**:
```
配置 LLM API 密钥和端点
├─ Python 后端: 更新 config.py
├─ 支持多个提供商选择
├─ 环境变量配置
└─ 时间: 1 天
```

### 问题 5: 真实 RSS 源
当前只有演示数据。

**解决方案**:
```
导入真实 RSS 源
├─ 允许用户添加真实 RSS URL
├─ 触发真实抓取
├─ 存储到数据库
└─ 时间: 1 天 (前端) + 触发成本
```

---

## 📊 分阶段实施方案

### Phase 5.1: 基础设施准备（3-5 天）

#### 第 1 天: Python HTTP 层实现

**在 Python 后端添加 FastAPI**

```bash
# 1. 更新 requirements.txt
fastapi==0.104.1
uvicorn==0.24.0
pydantic==2.5.0

# 2. 创建 api_server.py
```

```python
# backend-python/api_server.py
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from agents.content_evaluator import ContentEvaluationAgent

app = FastAPI()
evaluator = ContentEvaluationAgent()

class EvaluationRequest(BaseModel):
    title: str
    content: str

@app.post("/api/evaluate")
async def evaluate(request: EvaluationRequest):
    """评估内容"""
    result = await evaluator.evaluate(
        title=request.title,
        content=request.content
    )
    return result

@app.post("/api/evaluate/stream")
async def evaluate_stream(request: EvaluationRequest):
    """流式评估（用于前端 SSE）"""
    async def stream_generator():
        # 流式返回评估过程
        yield f"data: {{'status': 'processing'}}\n\n"
        result = await evaluator.evaluate(...)
        yield f"data: {json.dumps(result)}\n\n"

    return StreamingResponse(stream_generator())

@app.get("/api/health")
async def health():
    return {"status": "healthy"}
```

**运行 Python API 服务**
```bash
cd backend-python
uvicorn api_server:app --host 0.0.0.0 --port 8081 --reload
```

**成本**: 3-5 天

#### 第 2 天: Go 后端补充缺失 API

**添加消息表和 API**

```sql
-- sql/messages_table.sql
CREATE TABLE IF NOT EXISTS messages (
  id SERIAL PRIMARY KEY,
  task_id INT REFERENCES sources(id),
  role VARCHAR (10),           -- 'user', 'ai'
  type VARCHAR (20),           -- 'text', 'execution'
  content TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_messages_task_id ON messages(task_id);
```

**Go Handler: 消息 API**

```go
// backend-go/handlers/message_handler.go

func GetTaskMessages(c *gin.Context) {
    taskID := c.Param("id")
    messages, err := repositories.GetMessages(taskID)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    c.JSON(200, messages)
}

func SaveMessage(c *gin.Context) {
    var msg Message
    c.BindJSON(&msg)
    err := repositories.SaveMessage(msg)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    c.JSON(201, msg)
}
```

**Go Router 配置**

```go
// 在 main.go 中添加
router.GET("/api/tasks/:id/messages", handlers.GetTaskMessages)
router.POST("/api/tasks/:id/messages", handlers.SaveMessage)
```

**成本**: 2-3 天

#### 第 3 天: SSE 聊天流实现

**Go Handler: SSE 端点**

```go
// backend-go/handlers/chat_handler.go

func ChatStream(c *gin.Context) {
    taskID := c.Query("taskId")
    message := c.Query("message")

    // 保存用户消息
    repositories.SaveMessage(Message{
        TaskID: taskID,
        Role: "user",
        Content: message,
    })

    c.Header("Content-Type", "text/event-stream")
    c.Header("Cache-Control", "no-cache")
    c.Header("Connection", "keep-alive")

    // 调用 Python 评估服务
    result := callPythonEvaluator(message)

    // 流式返回
    flusher := c.Writer.(http.Flusher)

    // 发送开始事件
    fmt.Fprintf(c.Writer, "data: %s\n\n", `{"status":"processing"}`)
    flusher.Flush()

    // 流式返回评估文本
    for chunk := range result.TextChunks {
        fmt.Fprintf(c.Writer, "data: %s\n\n", chunk)
        flusher.Flush()
    }

    // 发送完成事件
    fmt.Fprintf(c.Writer, "data: %s\n\n", `{"status":"complete"}`)
    flusher.Flush()
}
```

**成本**: 2-3 天

---

### Phase 5.2: LLM API 配置（1 天）

#### 配置真实 LLM API

**步骤 1: 选择 LLM 提供商**

```
选项:
□ OpenAI (GPT-4o, GPT-4, GPT-3.5)
□ Anthropic (Claude 3)
□ Deepseek (深度求索)
□ Qwen (阿里通义千问)
□ Other (其他)

推荐: OpenAI GPT-4o (最稳定，评估效果最好)
```

**步骤 2: 获取 API 密钥**

```
OpenAI:
1. 访问 https://platform.openai.com
2. 创建 API Key (限额 $5-20/月 可免费试用)
3. 记录密钥

或使用模拟密钥进行测试:
export OPENAI_API_KEY="sk-test-abc123..."
```

**步骤 3: 更新 Python 配置**

```python
# backend-python/config.py
from pydantic_settings import BaseSettings

class Settings(BaseSettings):
    # LLM 配置
    llm_provider: str = "openai"  # openai, anthropic, qwen
    llm_model: str = "gpt-4o"
    llm_api_key: str = "sk-..."
    llm_api_base: str = "https://api.openai.com/v1"
    llm_temperature: float = 0.7
    llm_max_tokens: int = 2000

    class Config:
        env_file = ".env"

settings = Settings()
```

**步骤 4: 更新 LangGraph 配置**

```python
# backend-python/agents/content_evaluator.py
from langchain_openai import ChatOpenAI

llm = ChatOpenAI(
    model=settings.llm_model,
    api_key=settings.llm_api_key,
    temperature=settings.llm_temperature,
    max_tokens=settings.llm_max_tokens,
)
```

**成本**: 1 天

---

### Phase 5.3: 真实 RSS 源导入（1-2 天）

#### 添加真实 RSS 源

**推荐的 RSS 源**

```
技术类:
• https://news.ycombinator.com/rss
• https://techcrunch.com/feed/
• https://www.theverge.com/rss/index.xml
• https://www.arstechnica.com/feed/
• https://feeds.arstechnica.com/arstechnica/index

AI/ML:
• https://openai.com/feed
• https://www.anthropic.com/feed
• https://feed.arxiv.org/rss/cs.AI (每日论文)
• https://huggingface.co/blog/feed.xml

创业/投资:
• https://www.sequoiacap.com/feed/
• https://blog.ycombinator.com/feed/
• https://www.notion.so/RSS-Feed

中文:
• https://www.infoq.cn/feed
• https://www.cnblogs.com/rss.aspx
• https://www.geekbang.org/feed (付费内容不可用)
```

**前端配置**

在 Config.vue 中添加"推荐源"按钮:

```vue
<template>
  <div class="config">
    <button @click="addRecommendedSources">
      导入推荐 RSS 源
    </button>
  </div>
</template>

<script setup>
const recommendedSources = [
  {
    url: 'https://news.ycombinator.com/rss',
    author_name: 'Hacker News',
    priority: 9,
    platform: 'news'
  },
  // ... 更多源
]

const addRecommendedSources = async () => {
  for (const source of recommendedSources) {
    await api.sources.create(source)
  }
}
</script>
```

**手动添加**

在 Config.vue 中，用户可以直接输入 RSS URL 添加。

**成本**: 1-2 天

---

### Phase 5.4: 前端适配（2-3 天）

#### 适配列表

**1. 更新 API 基础 URL**

```javascript
// frontend-vue/.env.production
VITE_API_URL=http://localhost:8080
VITE_PYTHON_API_URL=http://localhost:8081
# 去除 VITE_MOCK_URL
```

**2. 更新 useAPI.js**

```javascript
// 删除 Mock URL，使用真实 API
const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080'
const pythonUrl = import.meta.env.VITE_PYTHON_API_URL || 'http://localhost:8081'

export const useAPI = () => {
  return {
    messages: {
      get: (taskId) => request(`/api/tasks/${taskId}/messages`, {
        baseUrl: apiUrl  // ✅ 改为真实 API
      }),
      save: (data) => request('/api/tasks/messages', {
        method: 'POST',
        data,
        baseUrl: apiUrl  // ✅ 改为真实 API
      }),
    },

    chat: {
      stream: (taskId, message) => request(
        `/api/chat/stream?taskId=${taskId}&message=${message}`,
        { baseUrl: apiUrl }  // ✅ 改为真实 API
      )
    },

    // Python 评估接口
    evaluate: {
      direct: (title, content) => request('/api/evaluate', {
        method: 'POST',
        data: { title, content },
        baseUrl: pythonUrl  // 调用 Python API
      }),

      stream: (title, content) => request('/api/evaluate/stream', {
        method: 'POST',
        data: { title, content },
        baseUrl: pythonUrl
      })
    }
  }
}
```

**3. 更新 TaskChat.vue**

```javascript
// 使用真实 SSE 端点
const response = await fetch(`${apiUrl}/api/chat/stream?taskId=${taskStore.selectedTaskId}&message=${messageText}`)
const reader = response.body.getReader()
const decoder = new TextDecoder()

while (true) {
  const { done, value } = await reader.read()
  if (done) break

  const text = decoder.decode(value)
  const lines = text.split('\n')

  for (const line of lines) {
    if (line.startsWith('data: ')) {
      const data = JSON.parse(line.slice(6))
      if (data.text) {
        aiText += data.text
      }
    }
  }
}
```

**4. 删除 Mock 后端配置**

```javascript
// 在 useAPI.js 中移除所有 mockUrl 引用
```

**成本**: 2-3 天

---

### Phase 5.5: 集成测试和修复（3-4 天）

#### 完整流程测试

**测试检查清单**

```
[ ] 1. Go 后端启动
    └─ docker-compose up -d
    └─ curl http://localhost:8080/api/health

[ ] 2. Python HTTP 服务启动
    └─ cd backend-python && uvicorn api_server:app --port 8081
    └─ curl http://localhost:8081/api/health

[ ] 3. 前端连接 Go 后端
    └─ 访问 Config 页面，添加 RSS 源
    └─ 检查 Network 中的 /api/sources 请求
    └─ 验证源已保存到数据库

[ ] 4. 消息保存和加载
    └─ 选择任务，发送消息
    └─ 检查 POST /api/tasks/{id}/messages
    └─ 刷新页面，验证消息持久化

[ ] 5. SSE 聊天流
    └─ 发送消息，触发 /api/chat/stream
    └─ 监听 SSE 事件流
    └─ 验证 AI 回复流式显示

[ ] 6. LLM 评估
    └─ 确认 API Key 配置正确
    └─ 触发评估，验证 AI 回复内容

[ ] 7. 执行历史
    └─ 手动执行任务
    └─ 检查历史记录保存
    └─ 验证统计数据准确
```

**常见问题排查**

```
问题 1: CORS 错误
解决: 在 Go 后端添加 CORS 中间件
router.Use(cors.Default())

问题 2: Python API 连接失败
解决: 检查 uvicorn 是否运行在 8081 端口

问题 3: LLM API 超时
解决: 增加超时时间，或使用异步请求

问题 4: SSE 连接断开
解决: 检查浏览器网络，可能需要重新连接

问题 5: 消息不保存
解决: 检查数据库连接和表权限
```

**成本**: 3-4 天

---

## 📊 实施时间表

```
Week 1 (Phase 5.1 - 基础设施)
├─ Day 1-2: Python HTTP 层 (FastAPI)
├─ Day 3-4: Go 消息 API + SSE
└─ Day 5: 缓冲日期

Week 2 (Phase 5.2 - LLM & RSS)
├─ Day 1: LLM API 配置
├─ Day 2-3: 真实 RSS 源导入
├─ Day 4: 前端小调整
└─ Day 5: 缓冲日期

Week 3 (Phase 5.3 - 前端适配)
├─ Day 1-2: 更新 useAPI.js
├─ Day 3: 更新组件
├─ Day 4-5: 测试和修复

总计: 3-4 周
```

---

## 💰 成本评估

### 开发成本
- **工作量**: 3-4 周
- **团队**: 2-3 人（后端 1-2 人，前端 1 人）
- **难度**: ⭐⭐⭐⭐ (高，但可管理)

### API 成本

**OpenAI API** (推荐)
```
成本: ~$0.15 per 1K input tokens + $0.60 per 1K output tokens (GPT-4o)

估算:
• 每条评估: ~500 tokens input + 300 tokens output
• 成本: $0.075 + $0.18 = ~$0.25 每条
• 100 条/天 = $25/天 ≈ $750/月

建议:
• 设置账户日限额: $10/日 (API 配额面板)
• 使用免费试用: $5 初始额度
• 或使用便宜的模型: GPT-3.5 (~$0.001 per 1K tokens)
```

**其他选项**
```
Deepseek (便宜) - 按量付费，成本 ~1/10 OpenAI
Qwen (中国) - 按量付费，国内访问快
Claude (贵) - 按量付费，评估效果好但贵
Local LLaMA - 免费，但需要本地 GPU
```

---

## 🎯 风险评估

### 高风险项

| 风险 | 概率 | 影响 | 缓解方案 |
|------|------|------|---------|
| LLM API 超时 | 高 | 中 | 增加超时，使用异步队列 |
| RSS 源失效 | 中 | 低 | 定期检查，添加错误处理 |
| CORS 问题 | 中 | 中 | 提前配置 CORS 中间件 |
| 消息数据丢失 | 低 | 高 | 完整的事务处理，备份 |

### 降低风险

✅ **充分的备份**
- 保留 Mock 后端数据，可快速回滚
- 数据库定期备份
- 版本控制所有改动

✅ **逐步迁移**
- 先完成基础设施 (Phase 5.1)
- 再集成 LLM API (Phase 5.2)
- 最后全量切换 (Phase 5.3)

✅ **完整测试**
- 所有新端点都有测试
- 集成测试覆盖主要流程
- 性能测试验证性能指标

---

## ✅ 完整检查清单

### 准备阶段
- [ ] 确认 OpenAI API Key (或选择其他 LLM)
- [ ] 准备 RSS 源列表
- [ ] 团队讨论技术方案
- [ ] 创建 feature branch

### 开发阶段
- [ ] Phase 5.1: Python + Go API
- [ ] Phase 5.2: LLM 配置
- [ ] Phase 5.3: RSS 导入
- [ ] Phase 5.4: 前端适配
- [ ] Phase 5.5: 集成测试

### 验证阶段
- [ ] 所有 API 端点测试通过
- [ ] SSE 流式正常工作
- [ ] 消息完整保存和加载
- [ ] LLM 评估有效
- [ ] 性能满足要求 (< 100ms)

### 上线阶段
- [ ] 数据库迁移脚本
- [ ] 备份现有数据
- [ ] 文档更新
- [ ] 团队培训
- [ ] 灰度发布

---

## 📌 关键决策需要你确认

### 1️⃣ LLM 选择
```
推荐: OpenAI GPT-4o
  • 评估效果最好
  • API 最稳定
  • 文档完善

成本: ~$0.25 每条评估

你的选择: [ ] OpenAI [ ] Deepseek [ ] 其他
```

### 2️⃣ 迁移策略
```
选项 A: 快速迁移 (1 周)
  • 一次性替换所有 Mock
  • 风险高，快速反馈

选项 B: 分步迁移 (2 周)
  • 先迁移消息，后迁移 SSE
  • 风险低，保险但慢

推荐: 选项 B

你的选择: [ ] A [ ] B
```

### 3️⃣ 工作量承受
```
预计 3-4 周，需要：
  • 1-2 名后端工程师
  • 1 名前端工程师

你能提供: [ ] 2-3 人 [ ] 1-2 人 [ ] 其他
```

---

## 🚀 立即行动方案

### 今天 (Day 1)
```
[ ] 1. 确认上述三个关键决策
[ ] 2. 获取 OpenAI API Key (如选择)
[ ] 3. 准备 RSS 源列表
[ ] 4. 创建 feature branch: real-backend-integration
```

### 本周 (Week 1)
```
[ ] 1. Phase 5.1: Python FastAPI 实现
[ ] 2. Phase 5.1: Go 消息 API
[ ] 3. Phase 5.1: SSE 端点实现
```

### 下周 (Week 2)
```
[ ] 1. Phase 5.2: LLM API 配置
[ ] 2. Phase 5.3: RSS 源导入
[ ] 3. Phase 5.4: 前端适配
```

### 第三周 (Week 3)
```
[ ] 1. Phase 5.5: 集成测试
[ ] 2. Bug 修复
[ ] 3. 性能优化
```

---

**方案制定**: Claude Haiku
**日期**: 2026-02-27
**状态**: 待确认和启动
