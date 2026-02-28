# Go 后端模块化重构指南

## 📋 概述

本文档详细说明如何将 282 行 `main.go` 单体文件重构为模块化架构，包括配置管理、基础设施、业务接口和依赖注入的最佳实践。

---

## 🏗️ 新的模块化架构

```
backend-go/
├── internal/
│   ├── config/
│   │   └── config.go              # 配置管理（统一的配置加载和覆盖）
│   ├── infra/
│   │   └── infra.go               # 基础设施（数据库、Redis 封装）
│   ├── domain/
│   │   └── interfaces.go           # 业务接口定义（接口隔离）
│   └── service/
│       ├── rss_fetcher.go          # RSS 抓取器实现
│       ├── stream_publisher.go     # Stream 发布器实现
│       └── factory.go              # 服务工厂（依赖注入）
├── repositories/
│   ├── source_repo.go
│   ├── content_repo.go
│   └── ...
├── services/
│   ├── rss_service.go              # 旧的 RSSService（逐步迁移）
│   └── ...
├── handlers/
│   └── ...
├── models/
│   └── ...
├── main.go                          # 保留原有版本
└── main_refactored.go              # 新的简化版本（示范）
```

---

## 🔑 核心设计原则

### 1. **依赖注入（Dependency Injection）**

**问题（重构前）：**
```go
// ❌ 紧耦合：服务直接创建依赖
rssService := services.NewRSSService(
    sourceRepo,
    contentRepo,
    rdb,
    contentService,
    // ...
)
```

**解决方案（重构后）：**
```go
// ✅ 接口注入：依赖通过接口传入
fetcher := NewRSSFetcher(
    sourceRepo,           // domain.SourceRepository 接口
    contentRepo,          // domain.ContentRepository 接口
    publisher,            // domain.StreamPublisher 接口
    deduplicator,         // domain.ContentDeduplicator 接口
    workerCount,          // 配置参数
    fetchTimeout,
    maxRetries,
)
```

**优势：**
- ✅ 便于单元测试（可以 mock 接口）
- ✅ 降低耦合（依赖于抽象，不依赖具体）
- ✅ 灵活扩展（可以轻松替换实现）

---

### 2. **接口隔离（Interface Segregation）**

**业务接口定义（`domain/interfaces.go`）：**

```go
// 仓储接口
type SourceRepository interface {
    GetAll(ctx context.Context, enabled bool) ([]*models.Source, error)
    GetByID(ctx context.Context, id int64) (*models.Source, error)
    UpdateLastFetchTime(ctx context.Context, id int64, time time.Time) error
}

type ContentRepository interface {
    Create(ctx context.Context, req *models.CreateContentRequest) (*models.Content, error)
    UpdateStatus(ctx context.Context, id int64, status string) error
    GetByHash(ctx context.Context, hash string) (*models.Content, error)
}

// 服务接口
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

**优势：**
- 清晰的契约定义
- 便于测试和 mock
- 支持多个实现

---

### 3. **配置驱动（Configuration-Driven）**

**统一的配置加载（`internal/config/config.go`）：**

```go
cfg := config.Load()  // 一行代码加载所有配置
```

**加载顺序：**
1. 设置默认值（包含 P0 优化值）
2. 从 `config.yaml` 读取
3. 环境变量覆盖

**P0 优化值已应用：**
```go
// 数据库连接池
MaxOpenConns: 50   // P0: 优化值
MaxIdleConns: 10   // P0: 优化值

// RSS 抓取
WorkerCount: 20         // P0: 从 5 改为 20
Timeout: "30s"          // P0: 从 10s 改为 30s
FetchInterval: "30m"    // P0: 从 1h 改为 30m
```

---

## 🔧 核心组件详解

### 1. **Config 管理**

**文件：** `internal/config/config.go`

```go
// 一行加载所有配置
cfg := config.Load()

// 获取处理后的值
dsn := cfg.GetDSN()
timeout := cfg.GetFetchTimeout()  // 自动解析 Duration
interval := cfg.GetFetchInterval()
```

**优势：**
- 集中管理所有配置
- 自动处理类型转换
- 支持多源覆盖（默认值 → YAML → 环境变量）

---

### 2. **基础设施层**

**文件：** `internal/infra/infra.go`

```go
// 数据库初始化
db, err := infra.NewDatabase(cfg)
defer db.Close()

// Redis 初始化
redis, err := infra.NewRedis(cfg)
defer redis.Close()
```

**关键特点：**
- ✅ P0 连接池配置已应用
- ✅ 自动 Ping 测试连接
- ✅ 统一的生命周期管理

---

### 3. **业务接口**

**文件：** `internal/domain/interfaces.go`

这里定义了应用的**核心契约**：

- `SourceRepository` - 数据源仓储
- `ContentRepository` - 内容仓储
- `RSSFetcher` - RSS 抓取器
- `StreamPublisher` - Stream 发布器
- `ContentDeduplicator` - 内容去重器

**关键优势：**
- 清晰定义各个模块的职责
- 便于测试（创建 mock 实现）
- 支持多个实现共存

---

### 4. **RSS 抓取器实现**

**文件：** `internal/service/rss_fetcher.go`

**核心特点：**

```go
type RSSFetcherImpl struct {
    // 仓储层（接口）
    sourceRepo  domain.SourceRepository
    contentRepo domain.ContentRepository

    // 服务层（接口）
    publisher     domain.StreamPublisher
    deduplicator  domain.ContentDeduplicator

    // 配置参数
    workerCount  int              // P0: 配置驱动
    fetchTimeout time.Duration    // P0: 配置驱动
    maxRetries   int

    // 生命周期
    ticker   *time.Ticker
    stopChan chan struct{}
    wg       sync.WaitGroup
}
```

**关键改进：**

1. **Worker 池大小配置化**
   ```go
   // 从配置读取 Worker 数（P0 优化：20）
   rf.workerCount   // 由构造函数参数传入
   ```

2. **接口注入便于测试**
   ```go
   // 可以轻松注入 mock 实现
   fetcher := NewRSSFetcher(
       mockSourceRepo,
       mockContentRepo,
       mockPublisher,
       mockDeduplicator,
       // ...
   )
   ```

3. **清晰的处理流程**
   ```
   fetchAllSources()
   → filterSourcesToFetch()
   → processSourcesWithWorkerPool()    // P0: Worker 数配置化
       → fetchWorker()                   // Worker 线程
           → fetchSourceWithRetry()      // 重试逻辑
               → processItem()           // 处理单个项目
                   → deduplicator.ValidateContent()  // 去重
                   → contentRepo.Create()            // 保存
                   → publisher.PublishToStream()     // 发布
   ```

---

### 5. **服务工厂**

**文件：** `internal/service/factory.go`

这是**依赖注入的核心**：

```go
// 创建工厂
factory, err := service.NewFactory(cfg)

// 工厂负责创建所有服务
rssFetcher, err := factory.CreateRSSFetcher()
publisher, err := factory.CreateStreamPublisher()

// 工厂管理资源生命周期
defer factory.Close()
```

**关键职责：**
1. 初始化基础设施（数据库、Redis）
2. 初始化仓储（自动注入数据库连接）
3. 初始化服务（自动注入依赖）
4. 管理所有资源的关闭

**工厂模式的优势：**
- ✅ 单一职责：创建和管理对象
- ✅ 统一的初始化流程
- ✅ 便于后续扩展

---

## 📊 重构对比

### 重构前（282 行 main.go）

```go
// ❌ 问题：
// 1. main.go 混杂配置、初始化、业务逻辑
// 2. 紧耦合：Services 直接创建依赖
// 3. 难以扩展：修改配置需要改代码
// 4. 难以测试：无法 mock 依赖
// 5. 配置硬编码：P0 优化值分散在各处

func main() {
    cfg := loadConfig()
    db, _ := initDatabase(cfg)
    rdb := initRedis(cfg)

    sourceRepo := repositories.NewSourceRepository(db)
    contentService := services.NewContentService(rdb)
    rssService := services.NewRSSService(
        sourceRepo, contentRepo, rdb,
        contentService, workerCount,
        parseFetchTimeout, retryMax,
    )

    appCtx := &AppContext{
        DB: db,
        Redis: rdb,
        RSSService: rssService,
        // ... 其他字段
    }

    go rssService.Start(...)
    go startServer(cfg.Server.Port)
    select {}
}
```

### 重构后（模块化）

```go
// ✅ 改进：
// 1. main.go 只负责启动和协调
// 2. 松耦合：所有依赖都是接口
// 3. 易于扩展：新增服务无需修改 main
// 4. 易于测试：可以注入 mock 实现
// 5. 配置驱动：所有 P0 优化值在配置中

func main() {
    log.Println("Loading configuration...")
    cfg := config.Load()  // ← 统一加载

    log.Println("Initializing factory...")
    factory, _ := service.NewFactory(cfg)  // ← 工厂管理一切
    defer factory.Close()

    log.Println("Creating RSS fetcher...")
    rssFetcher, _ := factory.CreateRSSFetcher()  // ← 接口注入

    log.Println("Starting application...")
    app := &App{
        factory:    factory,
        rssFetcher: rssFetcher,
    }
    app.start(cfg)

    select {}
}
```

---

## 🧪 单元测试示例

**重构后的模块化设计，测试变得简单：**

```go
// 创建 mock 实现
type mockSourceRepo struct {
    sources []*models.Source
}

func (m *mockSourceRepo) GetAll(ctx context.Context, enabled bool) ([]*models.Source, error) {
    return m.sources, nil
}

// ... 其他方法

// 测试 RSSFetcher
func TestRSSFetcher(t *testing.T) {
    fetcher := NewRSSFetcher(
        &mockSourceRepo{},           // ✅ 注入 mock
        &mockContentRepo{},
        &mockPublisher{},
        &mockDeduplicator{},
        20,   // workerCount
        30*time.Second,
        3,
    )

    // 测试逻辑
    err := fetcher.FetchSourceOnDemand(ctx, 1)
    assert.NoError(t, err)
}
```

---

## 📈 P0 优化在新架构中的应用

### 1. **数据库连接池**

```go
// infra/infra.go
maxOpenConns := cfg.Database.MaxOpenConns  // 50（P0 优化）
maxIdleConns := cfg.Database.MaxIdleConns  // 10（P0 优化）
db.SetMaxOpenConns(maxOpenConns)
db.SetMaxIdleConns(maxIdleConns)
```

### 2. **Worker 数配置化**

```go
// service/factory.go → service/rss_fetcher.go
fetcher := NewRSSFetcher(
    // ...
    cfg.Ingestion.WorkerCount,    // 20（P0 优化，从 5 改为）
    cfg.GetFetchTimeout(),          // 30s（P0 优化，从 10s 改为）
    cfg.Ingestion.RetryMax,
)
```

### 3. **抓取间隔优化**

```go
// main_refactored.go
fetchInterval := cfg.GetFetchInterval()  // 30m（P0 优化，从 1h 改为）
rssFetcher.Start(ctx, fetchInterval)
```

---

## 🚀 迁移计划

### Phase 1: 并行运行（推荐）

1. 保留原有 `main.go` 不动
2. 创建新的 `main_refactored.go`
3. 逐步迁移功能到新架构
4. 两个版本并行测试

### Phase 2: 完全迁移

1. 新版本通过所有测试
2. 将 `main_refactored.go` 重命名为 `main.go`
3. 删除旧的 `main.go`
4. 更新 CI/CD 配置

### Phase 3: 清理

1. 删除重复的 services（旧的 RSSService）
2. 统一使用新的接口
3. 更新文档和示例

---

## 📚 最佳实践总结

| 原则 | 实现方式 | 文件 |
|------|---------|------|
| **配置管理** | 统一的 Config 结构 + 多源加载 | `internal/config/config.go` |
| **基础设施** | 数据库、Redis 封装 | `internal/infra/infra.go` |
| **业务接口** | 清晰的契约定义 | `internal/domain/interfaces.go` |
| **依赖注入** | 工厂模式 | `internal/service/factory.go` |
| **功能隔离** | 各服务独立实现接口 | `internal/service/*.go` |
| **生命周期** | 工厂管理资源 | `internal/service/factory.go` |

---

## ✅ 检查清单

- [ ] 配置加载工作正常
- [ ] 基础设施初始化成功
- [ ] RSS 抓取器接收所有依赖
- [ ] Stream 发布器独立运行
- [ ] 工厂管理所有资源
- [ ] 单元测试可以注入 mock
- [ ] P0 优化值在配置中生效
- [ ] 新旧版本兼容运行

---

## 🔗 相关文件

- 配置示例：`config.yaml`（含 P0 优化值）
- 现有仓储：`repositories/*.go`（保持不变）
- 现有 handlers：`handlers/*.go`（需要适配）
- 测试示例：待编写

---

**完成时间：** 2026-02-28
**修改人：** Claude Code
**版本：** 1.0（初始模块化架构）
