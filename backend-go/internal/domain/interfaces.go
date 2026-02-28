package domain

import (
	"context"
	"time"

	"github.com/junkfilter/backend-go/models"
)

// ============================================================================
// 仓储接口（Repository Interfaces）
// ============================================================================

// SourceRepository 定义数据源仓储接口
type SourceRepository interface {
	GetAll(ctx context.Context, enabled bool) ([]*models.Source, error)
	GetByID(ctx context.Context, id int64) (*models.Source, error)
	UpdateLastFetchTime(ctx context.Context, id int64, time time.Time) error
}

// ContentRepository 定义内容仓储接口
type ContentRepository interface {
	Create(ctx context.Context, req *models.CreateContentRequest) (*models.Content, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
	GetByHash(ctx context.Context, hash string) (*models.Content, error)
}

// ============================================================================
// 服务接口（Service Interfaces）
// ============================================================================

// RSSFetcher 定义 RSS 抓取器接口
type RSSFetcher interface {
	// Start 启动 RSS 抓取服务，按指定间隔周期运行
	Start(ctx context.Context, interval time.Duration) error

	// Stop 优雅关闭服务
	Stop()

	// FetchSourceOnDemand 按需抓取单个数据源
	FetchSourceOnDemand(ctx context.Context, sourceID int64) error
}

// StreamPublisher 定义 Stream 发布器接口
type StreamPublisher interface {
	// PublishToStream 发布内容到 Redis Stream 用于评估
	PublishToStream(ctx context.Context, content *models.Content) error

	// GetStreamPending 获取待处理消息数
	GetStreamPending(ctx context.Context) (int64, error)
}

// ContentDeduplcator 定义内容去重器接口
type ContentDeduplicator interface {
	// ValidateContent 验证内容是否重复
	// 返回: (contentHash, isDuplicate, error)
	ValidateContent(ctx context.Context, url, title, content string) (string, bool, error)

	// MarkAsSeen 标记内容为已见
	MarkAsSeen(ctx context.Context, url, hash string) error

	// InitializeBloomFilter 初始化 Bloom 过滤器
	InitializeBloomFilter(ctx context.Context) error
}

// ============================================================================
// 注入器工厂接口
// ============================================================================

// ServiceFactory 定义服务工厂接口，用于依赖注入
type ServiceFactory interface {
	// CreateRSSFetcher 创建 RSS 抓取器
	CreateRSSFetcher() (RSSFetcher, error)

	// CreateStreamPublisher 创建 Stream 发布器
	CreateStreamPublisher() (StreamPublisher, error)

	// Close 关闭工厂及所有关联资源
	Close() error
}
