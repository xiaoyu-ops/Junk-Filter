package service

import (
	"log"

	"github.com/junkfilter/backend-go/internal/config"
	"github.com/junkfilter/backend-go/internal/domain"
	"github.com/junkfilter/backend-go/internal/infra"
	"github.com/junkfilter/backend-go/repositories"
	"github.com/junkfilter/backend-go/services"
)

// ============================================================================
// ServiceFactory - 依赖注入工厂
// ============================================================================

// Factory 实现 domain.ServiceFactory 接口
// 职责：创建和管理所有服务及其依赖的生命周期
type Factory struct {
	// 基础设施
	db    *infra.Database
	redis *infra.Redis

	// 配置
	cfg *config.Config

	// 仓储（缓存）
	sourceRepo      domain.SourceRepository
	contentRepo     domain.ContentRepository

	// 服务（缓存）
	publisher      domain.StreamPublisher
	deduplicator   domain.ContentDeduplicator
}

// NewFactory 创建服务工厂
// 这是初始化的核心：所有依赖关系在这里管理
func NewFactory(cfg *config.Config) (*Factory, error) {
	// Step 1: 初始化基础设施
	log.Println("\n========== Initializing Infrastructure ==========")

	db, err := infra.NewDatabase(cfg)
	if err != nil {
		return nil, err
	}

	redis, err := infra.NewRedis(cfg)
	if err != nil {
		db.Close()
		return nil, err
	}

	// Step 2: 初始化仓储
	log.Println("\n========== Initializing Repositories ==========")

	sourceRepo := repositories.NewSourceRepository(db.Conn())
	contentRepo := repositories.NewContentRepository(db.Conn())

	log.Println("✓ Repositories initialized")

	// Step 3: 初始化服务
	log.Println("\n========== Initializing Services ==========")

	publisher := NewStreamPublisher(redis.Client())
	log.Println("✓ StreamPublisher created")

	// Deduplicator 是现有的 services.DedupService，需要兼容
	// 因为 DedupService 期望具体的 *repositories.ContentRepository，而不是接口
	deduplicator := services.NewDedupService(redis.Client(), contentRepo)

	return &Factory{
		db:             db,
		redis:          redis,
		cfg:            cfg,
		sourceRepo:     sourceRepo,
		contentRepo:    contentRepo,
		publisher:      publisher,
		deduplicator:   deduplicator,
	}, nil
}

// CreateRSSFetcher 创建 RSS 抓取器
// 这是优雅的依赖注入示例：所有依赖通过参数传入
func (f *Factory) CreateRSSFetcher() (domain.RSSFetcher, error) {
	fetcher := NewRSSFetcher(
		f.sourceRepo,
		f.contentRepo,
		f.publisher,
		f.deduplicator,
		f.cfg.Ingestion.WorkerCount,      // P0: 从配置读取（20）
		f.cfg.GetFetchTimeout(),            // P0: 从配置读取（30s）
		f.cfg.Ingestion.RetryMax,
	)

	log.Println("✓ RSSFetcher created with:")
	log.Printf("  - WorkerCount: %d (P0 optimized)", f.cfg.Ingestion.WorkerCount)
	log.Printf("  - FetchTimeout: %s (P0 optimized)", f.cfg.Ingestion.Timeout)
	log.Printf("  - RetryMax: %d", f.cfg.Ingestion.RetryMax)

	return fetcher, nil
}

// CreateStreamPublisher 创建 Stream 发布器
func (f *Factory) CreateStreamPublisher() (domain.StreamPublisher, error) {
	return f.publisher, nil
}

// Close 关闭工厂及所有资源
func (f *Factory) Close() error {
	if f.redis != nil {
		if err := f.redis.Close(); err != nil {
			log.Printf("Warning: Failed to close Redis: %v", err)
		}
	}

	if f.db != nil {
		if err := f.db.Close(); err != nil {
			log.Printf("Warning: Failed to close Database: %v", err)
		}
	}

	return nil
}

// ============================================================================
// 便利方法：直接获取各个组件
// ============================================================================

// DB 返回数据库连接
func (f *Factory) DB() *infra.Database {
	return f.db
}

// Redis 返回 Redis 客户端
func (f *Factory) Redis() *infra.Redis {
	return f.redis
}

// Config 返回配置
func (f *Factory) Config() *config.Config {
	return f.cfg
}

// SourceRepo 返回数据源仓储
func (f *Factory) SourceRepo() domain.SourceRepository {
	return f.sourceRepo
}

// ContentRepo 返回内容仓储
func (f *Factory) ContentRepo() domain.ContentRepository {
	return f.contentRepo
}

// Publisher 返回 Stream 发布器
func (f *Factory) Publisher() domain.StreamPublisher {
	return f.publisher
}

// Deduplicator 返回去重器
func (f *Factory) Deduplicator() domain.ContentDeduplicator {
	return f.deduplicator
}
