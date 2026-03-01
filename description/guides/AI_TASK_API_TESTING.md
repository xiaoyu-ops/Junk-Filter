# AI 任务创建 API 测试指南

## 前提条件

- Python 后端运行在 `http://localhost:8000`
- Go 后端运行在 `http://localhost:8080`
- 前端运行在 `http://localhost:5173`
- 环境变量配置完整（OPENAI_API_KEY, LLM_MODEL_ID, LLM_BASE_URL）

## API 端点

### Python API: POST /api/task/ai-create

#### 功能
使用 AI 分析用户自然语言需求，从可用 RSS 源中推荐最合适的源。

#### 请求格式

```http
POST http://localhost:8000/api/task/ai-create
Content-Type: application/json

{
  "message": "用户自然语言需求",
  "sources": [
    {
      "id": 1,
      "url": "https://example.com/rss",
      "author_name": "源名称",
      "platform": "github|blog|twitter|medium",
      "priority": 1-10,
      "enabled": true
    }
  ],
  "conversation_history": [
    {
      "role": "user|ai",
      "content": "消息内容"
    }
  ]
}
```

#### 响应格式

```json
{
  "reply": "AI 对用户的友好回复",
  "pending_task": {
    "id": "source-123",
    "title": "任务标题",
    "source_name": "推荐的源名称",
    "priority": 8,
    "description": "可选的任务描述"
  },
  "source_name": "推荐的源名称"
}
```

### Go API: POST /api/tasks/ai-create

#### 功能
前端调用的网关端点，自动获取源列表并调用 Python 后端。

#### 请求格式

```http
POST http://localhost:8080/api/tasks/ai-create
Content-Type: application/json

{
  "message": "用户自然语言需求",
  "conversation_history": [
    {
      "role": "user|ai",
      "content": "消息内容"
    }
  ]
}
```

#### 响应格式

与 Python API 相同。

## 测试场景

### 场景 1: 精确匹配源

**用户输入**:
```
"我想监控 GitHub Trends"
```

**预期响应（LLM 成功）**:
```json
{
  "reply": "我理解了！GitHub Trends 是一个很好的选择，可以帮助你及时了解开源项目的最新趋势。我建议创建一个任务来持续监控。",
  "pending_task": {
    "id": "source-1",
    "title": "监控 GitHub 趋势",
    "source_name": "GitHub Trends",
    "priority": 8
  },
  "source_name": "GitHub Trends"
}
```

**预期响应（LLM 失败，Fallback）**:
```json
{
  "reply": "我找到了 GitHub Trends 的订阅源。你想要创建一个监控任务吗？\n\n任务名称会是：\"监控 GitHub Trends\"，执行频率为每天。确认创建吗？",
  "pending_task": {
    "id": "source-1",
    "title": "监控 GitHub Trends",
    "source_name": "GitHub Trends",
    "priority": 8
  },
  "source_name": "GitHub Trends"
}
```

### 场景 2: 复杂自然语言理解

**用户输入**:
```
"我需要关注 AI 和机器学习领域最新的研究论文，特别是 LLM 相关的内容"
```

**预期响应（LLM 推荐最合适的源）**:
```json
{
  "reply": "根据你的需求，我推荐 ArXiv 和 Google Scholar 这两个源，它们都有很多最新的 AI 研究论文。特别推荐 ArXiv，因为它更新频率快，论文分类清晰。我建议创建一个任务监控 ArXiv 中的 LLM 相关论文。",
  "pending_task": {
    "id": "source-5",
    "title": "监控 ArXiv LLM 研究论文",
    "source_name": "ArXiv",
    "priority": 9
  },
  "source_name": "ArXiv"
}
```

### 场景 3: 无匹配源

**用户输入**:
```
"我想监控 NASA 官网的最新任务"
```

**预期响应（无匹配源）**:
```json
{
  "reply": "我们的默认源中没有找到匹配的订阅源。你可以：\n1. 提供 NASA RSS 源的 URL\n2. 告诉我你想要监控的具体内容类型\n3. 查看现有的源列表，选择一个相近的\n\n你想怎么做呢？",
  "pending_task": null,
  "source_name": null
}
```

### 场景 4: 多轮对话

**第一轮**:
```
用户: "我想监控技术新闻"
AI:   "我发现了几个科技新闻源，包括 HackerNews、TechCrunch 等。你更喜欢哪一个？"
```

**第二轮**:
```
用户: "HackerNews 好"
AI:   "很好！HackerNews 是关注技术趋势的绝佳选择。我为你创建一个监控任务。"
```

**conversation_history**:
```json
{
  "message": "HackerNews 好",
  "conversation_history": [
    {
      "role": "user",
      "content": "我想监控技术新闻"
    },
    {
      "role": "ai",
      "content": "我发现了几个科技新闻源，包括 HackerNews、TechCrunch 等。你更喜欢哪一个？"
    }
  ]
}
```

## cURL 测试命令

### 测试 1: 简单请求

```bash
curl -X POST http://localhost:8000/api/task/ai-create \
  -H "Content-Type: application/json" \
  -d '{
    "message": "我想监控 GitHub 项目",
    "sources": [
      {
        "id": 1,
        "url": "https://github.com/trending",
        "author_name": "GitHub Trends",
        "platform": "github",
        "priority": 8,
        "enabled": true
      },
      {
        "id": 2,
        "url": "https://news.ycombinator.com",
        "author_name": "Hacker News",
        "platform": "tech",
        "priority": 7,
        "enabled": true
      }
    ],
    "conversation_history": []
  }'
```

### 测试 2: 通过 Go 后端

```bash
curl -X POST http://localhost:8080/api/tasks/ai-create \
  -H "Content-Type: application/json" \
  -d '{
    "message": "我需要关注 AI 领域的最新发展",
    "conversation_history": [
      {
        "role": "user",
        "content": "我对 AI 感兴趣"
      },
      {
        "role": "ai",
        "content": "很好，我可以帮你找到相关的资源"
      }
    ]
  }'
```

### 测试 3: 空源列表

```bash
curl -X POST http://localhost:8000/api/task/ai-create \
  -H "Content-Type: application/json" \
  -d '{
    "message": "我想创建一个新源来监控什么内容？",
    "sources": [],
    "conversation_history": []
  }'
```

## 前端集成测试

### 通过浏览器测试

1. 打开 http://localhost:5173
2. 点击右上方的 AI 助手图标（🤖）或按钮
3. 在输入框中输入自然语言需求
4. 按 Enter 或点击发送按钮
5. 观察 AI 的回复和任务建议

### 测试 JavaScript API

在浏览器控制台运行:

```javascript
// 导入 useAPI
import { useAPI } from '@/composables/useAPI.js'

const { chat, tasks } = useAPI()

// 测试 createTaskWithAI
const response = await chat.createTaskWithAI(
  "我想监控 GitHub Python 项目",
  []
)

console.log(response)
// 输出:
// {
//   reply: "...",
//   pendingTask: {...},
//   sourceName: "..."
// }
```

## 错误场景

### 错误 1: LLM API 不可用

**症状**: AI 返回规则匹配结果而不是 LLM 分析

**日志**:
```
[TaskAnalyzer] LLM analysis failed: ..., falling back to rule-based
```

**检查项**:
- [ ] OPENAI_API_KEY 是否正确设置
- [ ] LLM_BASE_URL 是否可访问
- [ ] 网络连接是否正常

### 错误 2: Python API 超时

**症状**: Go 后端超时，自动降级到本地匹配

**日志**:
```
[AI Task] Python API call failed: context deadline exceeded, falling back to local analysis
```

**检查项**:
- [ ] Python 服务是否运行
- [ ] PYTHON_API_BASE_URL 是否正确
- [ ] 网络连接是否正常

### 错误 3: 源列表为空

**症状**: AI 无法进行推荐

**日志**:
```
[TaskAnalyzer] Analyzing message: ... (0 sources available)
```

**检查项**:
- [ ] /api/sources 端点是否返回源
- [ ] 源是否启用 (enabled=true)

## 性能基准

| 场景 | 平均响应时间 | 备注 |
|------|------------|------|
| LLM 成功 | 2-5 秒 | 取决于 LLM API 速度 |
| 本地匹配 | <100ms | 快速关键词匹配 |
| Go 后端调用 | +1-2 秒 | 网络开销 |

## 监控指标

### 关键日志

**Python 后端** (`/var/log/python-api.log`):
```
[TaskAnalyzer] Analyzing message: ...
[TaskAnalyzer] LLM response: ...
[TaskAnalyzer] Parsed response: ...
```

**Go 后端** (`/var/log/go-api.log`):
```
[AI Task] Python API call successful: ...
[AI Task] Python API call failed: ..., falling back to local analysis
```

### 关键指标

- **LLM 调用成功率**: TaskAnalyzer 中 LLM 成功调用数 / 总调用数
- **平均响应时间**: 从请求到响应的时间
- **Fallback 触发率**: 规则匹配使用次数 / 总请求数

---

建议每次修改后都运行这些测试，确保功能正常工作。
