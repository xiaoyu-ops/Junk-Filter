# Go 后端模块化架构 - 代码范式示例

## 📖 目录

1. [基础使用](#基础使用)
2. [依赖注入模式](#依赖注入模式)
3. [单元测试](#单元测试)
4. [扩展新功能](#扩展新功能)
5. [常见问题](#常见问题)

---

## 基础使用

### 最小化初始化

```go
package main

import (
    "context"
    "log"
    "github.com/junkfilter/backend-go/internal/config"
    "github.com/junkfilter/backend-go/internal/service"
)

func main() {
    // 1. 加载配置（自动从 config.yaml + 环境变量）
    cfg := config.Load()

    // 2. 创建服务工厂（所有依赖都在这里初始化）
    factory, err := service.NewFactory(cfg)
    if err != nil {
        log.Fatalf("Failed to initialize: %v", err)
    }
    defer factory.Close()

    // 3. 创建 RSS 抓取器（接收所有依赖）
    fetcher, err := factory.CreateRSSFetcher()
    if err != nil {
        log.Fatalf("Failed to create fetcher: %v", err)
    }

    // 4. 启动服务
    ctx := context.Background()
    interval := cfg.GetFetchInterval()  // 30m（P0 优化）

    fetcher.Start(ctx, interval)
    defer fetcher.Stop()

    // 保持运行
    select {}
}
```

---

## 依赖注入模式

### 1. 接口优先

**定义接口（business contract）：**

```go
// domain/interfaces.go
type RSSFetcher interface {
    Start(ctx context.Context, interval time.Duration) error
    Stop()
    FetchSourceOnDemand(ctx context.Context, sourceID int64) error
}

type StreamPublisher interface {
    PublishToStream(ctx context.Context, content *models.Content) error
    GetStreamPending(ctx context.Context) (int64, error)
}
```

### 2. 工厂创建

**通过工厂创建实现（不直接 new）：**

```go
// service/factory.go
func (f *Factory) CreateRSSFetcher() (domain.RSSFetcher, error) {
    return NewRSSFetcher(
        f.sourceRepo,      // 注入仓储
        f.contentRepo,
        f.publisher,       // 注入服务
        f.deduplicator,
        f.cfg.Ingestion.WorkerCount,    // 注入配置
        f.cfg.GetFetchTimeout(),
        f.cfg.Ingestion.RetryMax,
    ), nil
}
```

### 3. 构造函数注入

**所有依赖通过参数传入：**

```go
// service/rss_fetcher.go
func NewRSSFetcher(
    sourceRepo domain.SourceRepository,      // ✅ 接口
    contentRepo domain.ContentRepository,    // ✅ 接口
    publisher domain.StreamPublisher,        // ✅ 接口
    deduplicator domain.ContentDeduplicator, // ✅ 接口
    workerCount int,                         // ✅ 配置
    fetchTimeout time.Duration,
    maxRetries int,
) domain.RSSFetcher {
    return &RSSFetcherImpl{
        sourceRepo:    sourceRepo,
        contentRepo:   contentRepo,
        publisher:     publisher,
        deduplicator:  deduplicator,
        workerCount:   workerCount,
        fetchTimeout:  fetchTimeout,
        maxRetries:    maxRetries,
        parser:        utils.NewRSSParser(),
        stopChan:      make(chan struct{}),
    }
}
```

---

## 单元测试

### 1. 创建 Mock 实现

```go
// test/mock_repository.go
package test

import (
    "context"
    "github.com/junkfilter/backend-go/models"
)

// MockSourceRepository 是 SourceRepository 的 mock 实现
type MockSourceRepository struct {
    GetAllFunc              func(ctx context.Context, enabled bool) ([]*models.Source, error)
    GetByIDFunc             func(ctx context.Context, id int64) (*models.Source, error)
    UpdateLastFetchTimeFunc func(ctx context.Context, id int64, t time.Time) error
}

// 实现接口
func (m *MockSourceRepository) GetAll(ctx context.Context, enabled bool) ([]*models.Source, error) {
    return m.GetAllFunc(ctx, enabled)
}

func (m *MockSourceRepository) GetByID(ctx context.Context, id int64) (*models.Source, error) {
    return m.GetByIDFunc(ctx, id)
}

func (m *MockSourceRepository) UpdateLastFetchTime(ctx context.Context, id int64, t time.Time) error {
    return m.UpdateLastFetchTimeFunc(ctx, id, t)
}
```

### 2. 编写测试

```go
// test/rss_fetcher_test.go
package test

import (
    "context"
    "testing"
    "time"
    "github.com/junkfilter/backend-go/models"
    "github.com/junkfilter/backend-go/internal/service"
)

func TestRSSFetcherStart(t *testing.T) {
    // 1. 创建 mock 依赖
    mockSourceRepo := &MockSourceRepository{
        GetAllFunc: func(ctx context.Context, enabled bool) ([]*models.Source, error) {
            return []*models.Source{
                {
                    ID:                    1,
                    URL:                   "https://example.com/feed",
                    FetchIntervalSeconds:  3600,
                    Priority:              5,
                },
            }, nil
        },
        UpdateLastFetchTimeFunc: func(ctx context.Context, id int64, t time.Time) error {
            return nil
        },
    }

    mockContentRepo := &MockContentRepository{
        CreateFunc: func(ctx context.Context, req *models.CreateContentRequest) (*models.Content, error) {
            return &models.Content{
                ID:    1,
                Title: req.Title,
                Status: "PENDING",
            }, nil
        },
    }

    mockPublisher := &MockPublisher{
        PublishToStreamFunc: func(ctx context.Context, c *models.Content) error {
            return nil
        },
    }

    mockDedup := &MockDeduplicator{
        InitializeBloomFilterFunc: func(ctx context.Context) error {
            return nil
        },
        ValidateContentFunc: func(ctx context.Context, url, title, content string) (string, bool, error) {
            return "hash123", false, nil  // 不重复
        },
        MarkAsSeenFunc: func(ctx context.Context, url, hash string) error {
            return nil
        },
    }

    // 2. 创建被测对象，注入 mock
    fetcher := service.NewRSSFetcher(
        mockSourceRepo,
        mockContentRepo,
        mockPublisher,
        mockDedup,
        5,                    // workerCount
        10*time.Second,       // fetchTimeout
        3,                    // maxRetries
    )

    // 3. 执行测试
    ctx := context.Background()
    err := fetcher.Start(ctx, 1*time.Hour)
    if err != nil {
        t.Errorf("Start failed: %v", err)
    }

    // 4. 验证行为
    defer fetcher.Stop()

    // 验证 mock 方法被调用
    if !mockSourceRepo.GetAllCalled {
        t.Error("Expected GetAll to be called")
    }
}

// 测试按需抓取
func TestFetchSourceOnDemand(t *testing.T) {
    mockSourceRepo := &MockSourceRepository{
        GetByIDFunc: func(ctx context.Context, id int64) (*models.Source, error) {
            return &models.Source{
                ID:  id,
                URL: "https://example.com/feed",
            }, nil
        },
        // ... 其他方法
    }

    fetcher := service.NewRSSFetcher(
        mockSourceRepo,
        &MockContentRepository{},
        &MockPublisher{},
        &MockDeduplicator{},
        5,
        10*time.Second,
        3,
    )

    // 执行
    ctx := context.Background()
    err := fetcher.FetchSourceOnDemand(ctx, 1)

    // 验证
    if err != nil {
        t.Errorf("FetchSourceOnDemand failed: %v", err)
    }
}
```

### 3. 表驱动测试

```go
func TestRSSFetcherWithVariousConfigs(t *testing.T) {
    tests := []struct {
        name        string
        workerCount int
        timeout     time.Duration
        maxRetries  int
        shouldPass  bool
    }{
        {
            name:        "P0 optimized config",
            workerCount: 20,      // P0 优化值
            timeout:     30*time.Second,  // P0 优化值
            maxRetries:  3,
            shouldPass:  true,
        },
        {
            name:        "Custom config",
            workerCount: 10,
            timeout:     60*time.Second,
            maxRetries:  5,
            shouldPass:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            fetcher := service.NewRSSFetcher(
                &MockSourceRepository{},
                &MockContentRepository{},
                &MockPublisher{},
                &MockDeduplicator{},
                tt.workerCount,
                tt.timeout,
                tt.maxRetries,
            )

            err := fetcher.Start(context.Background(), 1*time.Hour)
            defer fetcher.Stop()

            if (err == nil) != tt.shouldPass {
                t.Errorf("Start: got error=%v, want=%v", err, !tt.shouldPass)
            }
        })
    }
}
```

---

## 扩展新功能

### 场景：添加新的数据源爬虫

**Step 1: 定义接口**

```go
// domain/interfaces.go
type SourceCrawler interface {
    Crawl(ctx context.Context, source *models.Source) (items []*utils.FeedItem, err error)
}
```

**Step 2: 实现爬虫**

```go
// service/web_crawler.go
type WebCrawlerImpl struct {
    client *http.Client
}

func NewWebCrawler(client *http.Client) domain.SourceCrawler {
    return &WebCrawlerImpl{
        client: client,
    }
}

func (wc *WebCrawlerImpl) Crawl(ctx context.Context, source *models.Source) ([]*utils.FeedItem, error) {
    // 爬虫逻辑
    return items, nil
}
```

**Step 3: 在工厂中注册**

```go
// service/factory.go
func (f *Factory) CreateSourceCrawler() (domain.SourceCrawler, error) {
    return NewWebCrawler(&http.Client{}), nil
}
```

**Step 4: 在 RSS 抓取器中使用**

```go
// service/rss_fetcher.go
type RSSFetcherImpl struct {
    // ... 现有字段
    crawler domain.SourceCrawler  // ← 新增
}

func NewRSSFetcher(
    // ... 现有参数
    crawler domain.SourceCrawler,  // ← 注入爬虫
) domain.RSSFetcher {
    return &RSSFetcherImpl{
        // ...
        crawler: crawler,
    }
}

func (rf *RSSFetcherImpl) fetchSourceWithRetry(ctx context.Context, source *models.Source) {
    // 优先尝试爬虫
    if rf.crawler != nil {
        items, err := rf.crawler.Crawl(ctx, source)
        if err == nil {
            for _, item := range items {
                rf.processItem(ctx, source, item)
            }
            return
        }
    }

    // 降级到 RSS 解析
    items, _ := rf.parser.ParseFeed(source.URL)
    for _, item := range items {
        rf.processItem(ctx, source, item)
    }
}
```

---

## 配置最佳实践

### 1. P0 优化值的配置

**config.yaml:**

```yaml
database:
  host: localhost
  port: 5432
  user: truesignal
  password: truesignal123
  dbname: truesignal
  sslmode: disable
  max_open_conns: 50    # ← P0: 优化值（从 20 改为 50）
  max_idle_conns: 10    # ← P0: 新增

redis:
  host: redis
  port: 6379
  db: 0

server:
  port: 8080

ingestion:
  worker_count: 20           # ← P0: 优化值（从 5 改为 20）
  timeout: 30s               # ← P0: 优化值（从 10s 改为 30s）
  retry_max: 3
  fetch_interval: 30m        # ← P0: 优化值（从 1h 改为 30m）
```

### 2. 环境变量覆盖

```bash
# 运行时覆盖配置
export DB_HOST=prod-db.example.com
export INGESTION_WORKERS=50  # 在 prod 中使用更多 workers
export INGESTION_FETCH_INTERVAL=15m  # 更频繁的抓取

./backend-go
```

### 3. 多环境配置

```bash
# 开发环境
export ENV=dev
./backend-go  # 使用默认值和 config.yaml

# 生产环境
export ENV=prod
export DB_HOST=prod-db
export INGESTION_WORKERS=50
export INGESTION_TIMEOUT=60s
./backend-go
```

---

## 常见问题

### Q1: 如何添加新的依赖？

**A:**
1. 在 `domain/interfaces.go` 定义接口
2. 在 `internal/service/` 中实现接口
3. 在 `service/factory.go` 中的工厂方法中添加
4. 修改受影响的构造函数，增加参数

```go
// Example: 添加 Logger
type Logger interface {
    Debug(msg string)
    Error(msg string, err error)
}

// 在工厂中创建
func (f *Factory) CreateLogger() Logger {
    return logging.NewLogger(f.cfg.Log.Level)
}

// 在 RSSFetcher 中使用
func NewRSSFetcher(
    // ... 现有参数
    logger domain.Logger,  // ← 新增
) domain.RSSFetcher {
    return &RSSFetcherImpl{
        // ...
        logger: logger,
    }
}
```

### Q2: 如何测试工厂本身？

**A:**

```go
func TestFactoryInitialization(t *testing.T) {
    cfg := &config.Config{}
    cfg.Database.Host = "localhost"
    cfg.Database.Port = 5432
    // ... 设置其他配置

    factory, err := service.NewFactory(cfg)
    if err != nil {
        t.Errorf("NewFactory failed: %v", err)
    }

    // 验证工厂能创建各个组件
    fetcher, _ := factory.CreateRSSFetcher()
    if fetcher == nil {
        t.Error("CreateRSSFetcher returned nil")
    }

    publisher, _ := factory.CreateStreamPublisher()
    if publisher == nil {
        t.Error("CreateStreamPublisher returned nil")
    }

    factory.Close()
}
```

### Q3: P0 优化值在新架构中如何生效？

**A:**

1. **配置加载**：`config.Load()` 读取 `config.yaml` 中的优化值
2. **工厂创建**：`NewFactory()` 将配置传给所有初始化方法
3. **服务使用**：`NewRSSFetcher()` 接收配置参数（WorkerCount、FetchTimeout）
4. **运行时应用**：`RSSFetcherImpl` 使用这些参数进行并发抓取

```go
// 配置中的 P0 优化值
WorkerCount: 20         // ← 20 个 workers（从 5 改为）
FetchTimeout: 30s       // ← 30 秒超时（从 10s 改为）

// 在工厂中使用
fetcher := NewRSSFetcher(
    // ...
    cfg.Ingestion.WorkerCount,    // 接收 20
    cfg.GetFetchTimeout(),         // 接收 30s
    // ...
)

// 在 RSSFetcher 中应用
rf.processSourcesWithWorkerPool(ctx, sources)  // 使用 20 个 workers
```

---

## 总结

✅ **模块化架构的核心优势：**

| 方面 | 优势 |
|------|------|
| **可测试性** | 通过接口注入，轻松创建 mock |
| **可扩展性** | 添加新功能不需要修改现有代码 |
| **可维护性** | 清晰的职责划分和依赖关系 |
| **配置管理** | P0 优化值在配置中集中管理 |
| **并发优化** | Worker 数、超时等参数配置化 |

✅ **P0 优化在新架构中的应用：**

- 数据库连接池：50 个连接（P0 优化）
- Worker 数：20 个（P0 优化，从 5 改为）
- 抓取超时：30 秒（P0 优化，从 10s 改为）
- 抓取间隔：30 分钟（P0 优化，从 1h 改为）

---

**完成时间：** 2026-02-28
**版本：** 1.0
