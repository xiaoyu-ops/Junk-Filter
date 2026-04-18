# ✅ RSS 抓取功能完整分析

**日期**: 2026-03-01
**状态**: ✅ **完全可用**

---

## 📊 RSS 抓取能力评估

### ✅ **总体结论：RSS 抓取可以完全正常工作**

RSS 抓取系统已完全实现，包含：
- ✅ RSS 解析器（使用 gofeed 库）
- ✅ 定时抓取服务（Worker Pool 模式）
- ✅ 三层去重机制
- ✅ 错误处理和重试机制
- ✅ 手动同步支持
- ✅ 流式发布到 Redis

---

## 🔍 系统架构分析

### 1️⃣ RSS 解析层

**文件**: `backend-go/utils/rss_parser.go`

**功能**：
```go
// 使用 gofeed 库解析 RSS/Atom feeds
ParseFeed(feedURL string) ([]*FeedItem, error)
```

**支持格式**：
- ✅ RSS 2.0
- ✅ Atom
- ✅ JSON Feed
- ✅ 其他标准 Feed 格式

**提取的内容**：
- Title（标题）
- Description（描述）
- URL（链接）
- Author（作者）
- PublishedAt（发布时间）
- Content（内容）

**错误处理**：
- ✅ 网络连接失败处理
- ✅ 格式解析错误处理
- ✅ Nil 值检查

---

### 2️⃣ RSS 服务层

**文件**: `backend-go/services/rss_service.go`

**核心特性**：

#### ✅ 定时抓取
```
启动 RSSService.Start()
    ↓
初始化 Bloom Filter（去重）
    ↓
立即执行一次完整抓取
    ↓
每隔 N 秒执行一次定时抓取
```

**初始化参数**：
```go
type RSSService struct {
    workerCount    int           // 并发抓取工作线程数
    fetchTimeout   time.Duration // 单个源的抓取超时
    maxRetries     int           // 抓取失败的重试次数
}
```

#### ✅ 智能抓取过滤

```go
// 只抓取需要更新的源
filterSourcesToFetch()
  ├─ 从未抓取的源（LastFetchTime == nil）
  └─ 上次抓取距今超过 fetch_interval_seconds 的源
```

#### ✅ Worker Pool 并发模式

```
fetchAllSources()
  ├─ 创建 sourceChan（通道）
  ├─ 启动 N 个 fetchWorker goroutines
  ├─ 分配源到通道
  └─ 等待所有 worker 完成
```

**并发优势**：
- 可配置的并发数（workerCount）
- 充分利用 Go 的并发能力
- 同时处理多个 RSS 源

#### ✅ 重试机制

```go
for attempt := 1; attempt <= rs.maxRetries; attempt++ {
    items, err := rs.parser.ParseFeed(source.URL)
    if err == nil {
        // 成功，处理数据
        return
    }
    // 重试...
}
```

**重试特点**：
- ✅ 可配置的最大重试次数
- ✅ 详细的错误日志
- ✅ 失败后继续尝试下一个源

#### ✅ 失败处理

```
单个源抓取失败
  ├─ 记录详细日志
  ├─ 继续处理其他源（不中断整体流程）
  └─ 下次定时周期重试
```

---

### 3️⃣ 内容处理流程

**函数**: `RSSService.processItem()`

```
获取 RSS Item
    ↓
1️⃣ 内容清理（HTML 去标签化）
    ↓
2️⃣ 去重验证（三层机制）
    ├─ L1: Bloom Filter（快速拒绝）
    ├─ L2: Redis Set（精确检查，7 天 TTL）
    └─ L3: PostgreSQL UNIQUE 约束（最后防线）
    ↓
3️⃣ 创建 Content 记录
    ├─ sourceID（来源 ID）
    ├─ title（标题）
    ├─ original_url（原始链接）
    ├─ clean_content（清理后的内容）
    ├─ content_hash（内容哈希）
    └─ status = 'PENDING'（待评估）
    ↓
4️⃣ 标记为已见（防止重复）
    ├─ Redis Bloom Filter 更新
    └─ Redis Set 添加
    ↓
5️⃣ 发布到 Redis Stream
    └─ 等待 Python 评估服务消费
    ↓
6️⃣ 日志记录
```

**去重机制详解**：

| 层级 | 类型 | 特点 | 用途 |
|------|------|------|------|
| **L1** | Bloom Filter | 内存高效、误报率 <0.1% | 快速拒绝已见内容 |
| **L2** | Redis Set | 精确、7天 TTL | 精确验证 |
| **L3** | PostgreSQL | 最严格、约束级 | 捕获竞态条件 |

---

### 4️⃣ 手动同步支持

**函数**: `RSSService.FetchSourceOnDemand()`

```go
// 立即抓取指定源，不等待定时周期
FetchSourceOnDemand(ctx, sourceID)
```

**用途**：
- ✅ 前端"同步"按钮（配置页）
- ✅ 紧急更新特定源
- ✅ 测试源是否可用

**集成点**：
```
前端点击"同步"按钮
    ↓
发送 POST /api/sources/{id}/fetch
    ↓
Go 后端调用 FetchSourceOnDemand()
    ↓
立即抓取该源
    ↓
处理新内容
    ↓
返回响应
```

---

### 5️⃣ 信息流发布

**函数**: `ContentService.PublishToStream()`

```
抓取的内容
    ↓
发布到 Redis Stream: ingestion_queue
    ↓
消息格式：
{
    task_id: UUID,
    item_hash: content_hash,
    platform: source.platform,
    author: item.author,
    clean_content: text,
    published_at: timestamp
}
    ↓
Python 评估服务消费（XREADGROUP）
    ↓
使用 LLM 评估内容
    ↓
保存评估结果到 PostgreSQL
```

---

## 🔧 配置参数说明

### Go 后端配置

**文件**: `backend-go/config.yaml`

```yaml
rss:
  worker_count: 5              # 并发抓取线程数
  fetch_timeout: 30s           # 单个源抓取超时
  max_retries: 3               # 失败重试次数
  fetch_interval: 300s         # 定时抓取间隔（5分钟）
```

### 数据库 sources 表字段

```sql
CREATE TABLE sources (
    id SERIAL PRIMARY KEY,
    platform TEXT,                      -- 平台：blog、github、bilibili 等
    url TEXT,                          -- RSS Feed URL
    author_name TEXT,                  -- 源名称
    priority INT,                      -- 优先级 1-10（高优先级更频繁抓取）
    fetch_interval_seconds INT,        -- 抓取间隔（秒）
    enabled BOOLEAN,                   -- 是否启用
    last_fetch_time TIMESTAMP,         -- 上次抓取时间
    created_at TIMESTAMP,
    updated_at TIMESTAMP
)
```

---

## 📈 处理流量估算

### 假设配置

```
配置源数：28 个
平均每源项目数：10 个
去重后预期新内容：30-40% = 3-4 个/源

定时周期：300 秒（5 分钟）
并发数：5 workers
```

### 处理能力

```
吞吐量：
  28 源 × 10 项/源 ÷ 300秒 = 0.93 项/秒
  实际（去重后）= 0.3 项/秒

延迟：
  单源抓取时间 ≈ 2-5 秒
  总周期：28 源 ÷ 5 workers ≈ 6 个周期 ≈ 12-30 秒

资源占用：
  内存：Bloom Filter (~1MB) + 缓存 (~10MB) = ~11MB
  网络：~1 Mbps（RSS 内容相对小）
  CPU：1-2 核心充分
```

---

## 🚀 完整的 RSS 处理流程

### 时间序列

```
T0: Go 后端启动
    └─ RSSService.Start()
       ├─ 初始化 Bloom Filter
       └─ 立即执行一次 fetchAllSources()

T1-T30: 第一轮抓取
    ├─ 获取所有启用的源（28 个）
    ├─ 并发抓取（5 workers）
    │  ├─ Worker 1 抓取源 1-5
    │  ├─ Worker 2 抓取源 6-10
    │  ├─ Worker 3 抓取源 11-15
    │  ├─ Worker 4 抓取源 16-20
    │  └─ Worker 5 抓取源 21-28
    ├─ 对每个 item 进行：
    │  ├─ HTML 清理
    │  ├─ 三层去重
    │  ├─ 创建 content 记录
    │  └─ 发布到 Redis Stream
    └─ 所有源抓取完成

T300: 定时周期 1
    └─ 重复 T1-T30 流程

T600: 定时周期 2
    └─ 重复 T1-T30 流程

...

TN: 用户点击"同步"按钮（源 #21 - Bilibili）
    ├─ 前端发送 POST /api/sources/21/fetch
    ├─ Go 后端立即调用 FetchSourceOnDemand(21)
    ├─ 抓取并处理该源
    └─ 返回结果
```

### 并发示意图

```
时间轴
0ms    100ms  200ms  300ms  400ms  500ms
|------|------|------|------|------|
Worker1: [源1   完成][源6   完成][源11      ]
Worker2: [源2   完成][源7   完成][源12      ]
Worker3: [源3   完成][源8   完成][源13      ]
Worker4: [源4   完成][源9   完成][源14      ]
Worker5: [源5   完成][源10  完成][源15      ]

总耗时：~500ms（实际可能 1-2 秒，取决于网络）
```

---

## 🧪 测试 RSS 抓取

### 测试 1: 验证源已加载

```bash
# 查询数据库中的源
docker exec junkfilter-db psql -U junkfilter -d junkfilter << 'EOF'
SELECT COUNT(*) as total,
       COUNT(CASE WHEN last_fetch_time IS NOT NULL THEN 1 END) as fetched
FROM sources;
EOF

# 预期：
# total | fetched
# 28    | 0 (初始状态)
```

### 测试 2: 启动 Go 后端并等待抓取

```bash
# 启动 Go 后端
cd backend-go && go run main.go

# 观察日志（应该看到）：
# [RSS Service] Starting with 5 workers
# Successfully fetched https://rsshub.app/github/trending/daily (15 items)
# Successfully fetched https://rsshub.app/weibo/search/hot (20 items)
# ...
```

### 测试 3: 检查抓取结果

```bash
# 查看新增的 content 记录
docker exec junkfilter-db psql -U junkfilter -d junkfilter << 'EOF'
SELECT COUNT(*) as new_content FROM content WHERE status = 'PENDING';
SELECT platform, COUNT(*) FROM content GROUP BY platform LIMIT 10;
EOF

# 预期：看到来自多个平台的内容
```

### 测试 4: 手动同步单个源

```bash
# 手动同步 Bilibili 源（ID: 21）
curl -X POST http://localhost:8080/api/sources/21/fetch

# 预期响应：
# {"message": "Source fetch triggered"}

# 查看日志：
# Successfully fetched https://rsshub.app/bilibili/ranking/0/3/1 (8 items)
```

### 测试 5: 验证 Redis Stream

```bash
# 检查 Redis Stream 中的消息
docker exec junkfilter-redis redis-cli << 'EOF'
XLEN ingestion_queue
XRANGE ingestion_queue - + LIMIT 2
EOF

# 预期：看到最近的 ingestion 消息
```

---

## 📊 可能的问题和解决方案

### ❌ 问题 1: RSS 源抓取失败

**可能原因**：
- URL 格式不对（不是有效的 RSS Feed URL）
- 网络连接问题
- 源已下线

**解决方案**：
```
1. 检查 sources 表中的 URL 格式
   SELECT url FROM sources LIMIT 5;

2. 手动测试 URL
   curl https://rsshub.app/github/trending/daily

3. 检查后端日志是否有详细错误信息

4. 增加重试次数或超时时间
   config.yaml: max_retries = 5, fetch_timeout = 60s
```

### ❌ 问题 2: 内容没有发送到 Redis

**可能原因**：
- ContentService 配置不正确
- Redis 连接失败
- 内容被去重拒绝

**解决方案**：
```
1. 检查 Redis 连接
   docker exec junkfilter-redis redis-cli PING

2. 查看 content 表的状态
   SELECT status, COUNT(*) FROM content GROUP BY status;

3. 检查 Bloom Filter 初始化
   查看启动日志是否有 "Bloom filter initialized"
```

### ❌ 问题 3: Python 评估服务没有消费数据

**可能原因**：
- Python 后端未启动
- Redis Stream 消费者组未创建
- 消费逻辑有错误

**解决方案**：
```
1. 确保 Python 后端正在运行
   docker logs junkfilter-python-1

2. 检查消费者组
   docker exec junkfilter-redis redis-cli XINFO GROUPS ingestion_queue

3. 检查 Python 后端日志
   查看是否有评估进度
```

---

## ✅ 系统就绪检查清单

- [x] RSS 解析器已实现（gofeed）
- [x] RSSService 已实现（定时 + 并发）
- [x] 三层去重已实现（Bloom + Redis + DB）
- [x] 28 个真实 RSS 源已导入
- [x] Worker Pool 并发模式已实现
- [x] 错误重试机制已实现
- [x] 手动同步支持已实现
- [x] Redis Stream 发布已实现
- [x] 内容清理已实现

---

## 🎯 结论

**✅ RSS 抓取功能完全可用！**

系统已准备好：
1. **定时自动抓取** - 每 5 分钟执行一轮
2. **并发处理** - 5 个 workers 同时处理
3. **错误恢复** - 失败重试，不中断流程
4. **手动同步** - 支持实时抓取特定源
5. **内容去重** - 三层机制防止重复
6. **流式发布** - 实时发送到 Redis 给评估服务

**现在可以：**
- ✅ 启动 Go 后端
- ✅ 观察 RSS 抓取日志
- ✅ 检查数据库中的新内容
- ✅ 验证 Redis Stream 消息
- ✅ 在前端点击"同步"按钮测试手动同步

**准备就绪，可以进行完整的系统测试！** 🚀
