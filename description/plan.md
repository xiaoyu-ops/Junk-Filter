# Junk Filter - 功能演进计划

日期: 2026-03-09
状态: Phase 6-7 已完成，以下为新阶段规划

---

## 已完成里程碑

| Phase | 内容 | 状态 |
|-------|------|------|
| Phase 6 | 安全修复、代码清理、废弃端点移除 | 已完成 |
| Phase 7 | Docker 全栈容器化 + Tauri 桌面版 | 已完成 |

---

## Phase 8: RSS 图片抓取

### 目标
阅读页面展示文章原始图片，还原真实阅读体验。

### 当前问题
`rss_parser.go:112` 的 `CleanContent()` 用正则删除了所有 HTML 标签（包括 `<img>`），只存纯文本。图片 URL 在清洗阶段被丢弃。

### 实现方案

**8.1 Go 后端: 提取图片 URL**

文件: `backend-go/utils/rss_parser.go`

- 在 `CleanContent()` 之前，新增 `ExtractImageURLs(html string) []string`
- 用正则提取所有 `<img src="...">` 的 URL
- 返回去重后的图片 URL 列表

文件: `backend-go/utils/rss_parser.go` - `SanitizeFeedItem()`

- 调用 `ExtractImageURLs()` 提取图片
- 将图片列表存入 `FeedItem` 的新字段 `ImageURLs []string`

文件: `backend-go/models/content.go`

- 新增字段 `ImageURLs` (JSON 数组)

**8.2 数据库: 新增图片字段**

```sql
ALTER TABLE content ADD COLUMN image_urls JSONB DEFAULT '[]';
```

**8.3 Go 后端: 存储与 API**

文件: `backend-go/services/rss_service.go` - `processItem()`

- 将提取的图片 URL 列表写入 `image_urls` 字段

文件: `backend-go/handlers/` (相关查询接口)

- 返回 content 时包含 `image_urls` 字段

**8.4 前端: Reader 展示图片**

文件: `frontend-vue/src/components/Reader.vue`

- 渲染 `image_urls` 数组中的图片
- 图片懒加载 + 加载失败占位
- 点击图片可放大预览

### 涉及文件

| 文件 | 改动 |
|------|------|
| `backend-go/utils/rss_parser.go` | 新增 `ExtractImageURLs()`，修改 `SanitizeFeedItem()` |
| `backend-go/models/content.go` | 新增 `ImageURLs` 字段 |
| `backend-go/services/rss_service.go` | `processItem()` 存储图片 URL |
| `sql/` | 新迁移脚本，ALTER TABLE 加字段 |
| `frontend-vue/src/components/Reader.vue` | 图片渲染 |

---

## Phase 9: RSS 作者过滤

### 目标
支持只抓取 RSS 源中特定作者的文章，过滤掉不关注的作者。

### 当前状态
- Go RSS 解析器已经提取了 `Author` 字段（`rss_parser.go:73-80`）
- `content` 表已有 `author_name` 列
- `sources` 表有 `author_name` 但用途是源级别标识，没有过滤功能

### 实现方案

**9.1 数据库: sources 表新增过滤字段**

```sql
ALTER TABLE sources ADD COLUMN author_filter JSONB DEFAULT '{}';
-- 结构: {"mode": "whitelist|blacklist", "authors": ["Author A", "Author B"]}
-- mode=whitelist: 只抓取列表中的作者
-- mode=blacklist: 排除列表中的作者
-- 空对象 {} 表示不过滤（抓取所有）
```

**9.2 Go 后端: 抓取时过滤**

文件: `backend-go/services/rss_service.go` - `processItem()`

- 读取 source 的 `author_filter` 配置
- 在去重检查之前，先判断作者是否匹配过滤规则
- 不匹配的直接跳过，不进入去重和存储流程

文件: `backend-go/models/source.go`

- 新增 `AuthorFilter` 字段（JSON 结构）

**9.3 Go 后端: 管理 API**

文件: `backend-go/handlers/`

- `PUT /api/sources/:id/author-filter` - 设置过滤规则
- `GET /api/sources/:id/authors` - 列出该源已知的所有作者（从 content 表聚合）

**9.4 前端: Config 页面增强**

文件: `frontend-vue/src/components/Config.vue`

- 每个 RSS 源的展开行中新增"作者过滤"配置
- 支持白名单/黑名单模式切换
- 显示已抓取的作者列表供勾选

### 涉及文件

| 文件 | 改动 |
|------|------|
| `backend-go/models/source.go` | 新增 `AuthorFilter` 字段 |
| `backend-go/services/rss_service.go` | `processItem()` 加过滤逻辑 |
| `backend-go/handlers/` | 新增作者过滤 API |
| `sql/` | ALTER TABLE sources |
| `frontend-vue/src/components/Config.vue` | 作者过滤 UI |
| `frontend-vue/src/stores/useConfigStore.js` | 作者过滤状态管理 |

---

## Phase 10: 智能偏好学习 (Agent 评估调优)

### 目标
Agent 在与用户的自然对话中，自动捕捉用户偏好，逐步调整评估 prompt，使评估结果越来越符合用户口味。用户不需要知道 prompt 的存在。

### 核心机制

```
用户自然对话
    |
Agent 识别偏好信号 (隐式提取)
    |
结构化偏好画像 (存数据库)
    |
评估时注入偏好 (拼接到 prompt 末尾)
    |
评估结果逐渐贴合用户喜好
```

### 偏好信号类型

| 信号类型 | 对话示例 | 提取结果 |
|----------|----------|----------|
| 领域偏好 | "我主要关注 AI 和分布式系统" | `liked_topics += [AI, 分布式系统]` |
| 领域排斥 | "营销类的文章不用看" | `disliked_topics += [营销]` |
| 质量偏好 | "太水了，全是老生常谈" | `quality_bias = 偏好深度分析` |
| 风格偏好 | "有代码示例的文章更好" | `style_preference = 喜欢实战/代码` |
| 阈值偏好 | "8分以下别推了" | `score_threshold = 8` |
| 反馈信号 | "这篇不错" / "这篇没价值" | 强化/弱化对应特征 |

### 偏好画像结构

```json
{
  "source_id": 3,
  "liked_topics": ["AI", "分布式系统", "系统设计"],
  "disliked_topics": ["营销", "SEO"],
  "quality_bias": "偏好深度分析，不喜欢新闻速报",
  "style_preference": "喜欢有代码示例和实战经验的文章",
  "score_threshold": 8,
  "custom_notes": "关注学术前沿，但更看重可落地性",
  "updated_at": "2026-03-09"
}
```

### 实现方案

**10.1 数据库: 偏好存储**

```sql
CREATE TABLE user_preferences (
    id SERIAL PRIMARY KEY,
    source_id INT REFERENCES sources(id) ON DELETE CASCADE,
    preferences JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(source_id)  -- 每个源一个偏好画像
);
```

**10.2 Python 后端: 定义 LangChain Tools**

文件: `backend-python/agents/preference_tools.py` (新建)

```python
@tool
def get_user_preferences(source_id: int) -> str:
    """获取指定 RSS 源的用户偏好画像"""
    # 从 user_preferences 表读取

@tool
def update_user_preferences(source_id: int, updates: dict) -> str:
    """更新用户偏好，增量合并而非覆盖"""
    # 读取现有偏好 → 合并更新 → 写回数据库
```

**10.3 Python 后端: 升级聊天 Agent**

文件: `backend-python/api_server.py` - `task_chat()` 端点

- 将当前的"直接调 LLM"改为 `create_react_agent(llm, tools)`
- Agent 在对话中自动判断是否包含偏好信号
- 有偏好信号时静默调用 `update_user_preferences()`
- 无偏好信号时正常回复

聊天 Agent 的 system_prompt 中增加指引:
```
你在回复用户的同时，要注意识别对话中的偏好信号。
当用户表达了对某类内容的好恶、质量期望、领域兴趣时，
调用 update_user_preferences 工具静默记录，不需要告知用户。
```

**10.4 Python 后端: 评估时注入偏好**

文件: `backend-python/services/stream_consumer.py`

- 评估前查询 `user_preferences` 表
- 将偏好画像格式化为自然语言段落
- 拼接到评估 prompt 末尾

拼接示例:
```
[固定评估框架: 创新度/深度/决策标准]

[用户偏好提示]
- 此用户重视实用性和可操作性，涉及实战经验的内容应获得更高评分
- 用户关注 AI、分布式系统领域，相关主题可适当加分
- 用户不喜欢表面级新闻报道，纯资讯类内容应降低评分
```

**10.5 前端: 偏好可视化 (可选)**

文件: `frontend-vue/src/components/Config.vue`

- 每个 RSS 源的展开行中显示当前偏好画像（只读）
- 用户可手动重置偏好

### 涉及文件

| 文件 | 改动 |
|------|------|
| `sql/` | 新迁移: `user_preferences` 表 |
| `backend-python/agents/preference_tools.py` | 新建: 2 个 LangChain Tools |
| `backend-python/api_server.py` | 聊天端点改用 ReAct Agent |
| `backend-python/services/stream_consumer.py` | 评估前加载偏好并注入 prompt |
| `frontend-vue/src/components/Config.vue` | 偏好画像展示 (可选) |

---

## Phase 11: 后台通知系统

### 目标
系统在后台持续抓取和评估，当发现高价值文章时主动通知用户。定位: 默默运行，有好内容才打扰。

### 当前基础
- 评估分数已存在 `evaluation` 表 (`innovation_score`, `depth_score`)
- `user_subscription` 表已有阈值字段 (`min_innovation_score`, `min_depth_score`)
- Tauri 桌面版可调用系统原生通知

### 实现方案

**11.1 Python 后端: 评估后触发通知**

文件: `backend-python/services/stream_consumer.py`

- 评估完成后，检查分数是否超过阈值
- 超过阈值的文章写入 `notifications` 表
- 同时通过 Redis Pub/Sub 发布通知事件

**11.2 数据库: 通知表**

```sql
CREATE TABLE notifications (
    id SERIAL PRIMARY KEY,
    content_id INT REFERENCES content(id),
    title VARCHAR(500),
    summary TEXT,
    innovation_score INT,
    depth_score INT,
    decision VARCHAR(50),
    is_read BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW()
);
```

**11.3 Go 后端: 通知 API + SSE 推送**

文件: `backend-go/handlers/`

- `GET /api/notifications` - 获取通知列表 (支持 ?unread=true)
- `PUT /api/notifications/:id/read` - 标记已读
- `GET /api/notifications/stream` - SSE 实时推送端点

Go 后端订阅 Redis Pub/Sub 的通知频道，有新通知时通过 SSE 推送给前端。

**11.4 前端: 通知中心**

文件: `frontend-vue/src/components/NotificationBell.vue` (新建)

- 导航栏铃铛图标 + 未读数角标
- 点击展开通知下拉列表
- 点击通知项跳转到对应文章

文件: `frontend-vue/src/composables/useNotification.js` (新建)

- SSE 连接监听实时通知
- 管理未读数状态

**11.5 Tauri 桌面版: 系统原生通知**

文件: `frontend-vue/src/composables/useNotification.js`

- 检测 Tauri 环境时，调用 `@tauri-apps/plugin-notification` 发送系统通知
- 即使窗口最小化也能弹出系统级提醒
- Web 模式下 fallback 到浏览器 Notification API

### 涉及文件

| 文件 | 改动 |
|------|------|
| `sql/` | 新迁移: `notifications` 表 |
| `backend-python/services/stream_consumer.py` | 评估后检查阈值 + 写通知 |
| `backend-go/handlers/` | 通知 API + SSE 推送 |
| `frontend-vue/src/components/NotificationBell.vue` | 新建: 通知铃铛组件 |
| `frontend-vue/src/composables/useNotification.js` | 新建: 通知逻辑 + Tauri 系统通知 |

---

## 执行顺序

```
Phase 8   RSS 图片抓取          3-5 小时    <- 从这里开始 (改动小，立即可用)
Phase 9   RSS 作者过滤          5-8 小时
Phase 10  智能偏好学习         15-20 小时   (核心差异化功能)
Phase 11  后台通知系统         12-15 小时
```

### 依赖关系

```
Phase 8 (图片) ──── 独立，无依赖
Phase 9 (作者) ──── 独立，无依赖
Phase 10 (偏好) ─── 独立，但建议在 Phase 9 之后
                    (作者过滤 + 偏好学习配合效果更好)
Phase 11 (通知) ─── 依赖 Phase 10
                    (通知阈值可结合偏好动态调整)
```

Phase 8 和 Phase 9 可并行开发，互不影响。

---

## 关键文件索引

| 文件 | 用途 |
|------|------|
| `backend-go/utils/rss_parser.go` | RSS 解析、HTML 清洗、图片提取 |
| `backend-go/services/rss_service.go` | RSS 抓取流程、作者过滤 |
| `backend-go/models/source.go` | RSS 源数据模型 |
| `backend-go/models/content.go` | 文章内容数据模型 |
| `backend-python/api_server.py` | 聊天端点 (将升级为 ReAct Agent) |
| `backend-python/agents/content_evaluator.py` | 评估 Agent (接收偏好注入) |
| `backend-python/services/stream_consumer.py` | 评估消费者 (偏好加载 + 通知触发) |
| `frontend-vue/src/components/Config.vue` | 配置页 (作者过滤 UI) |
| `frontend-vue/src/components/Reader.vue` | 阅读页 (图片展示) |
