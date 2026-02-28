package service

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/junkfilter/backend-go/internal/domain"
	"github.com/junkfilter/backend-go/models"
	"github.com/junkfilter/backend-go/utils"
)

// ============================================================================
// RSSFetcherImpl - RSS 抓取器实现（优化版本）
// ============================================================================

// RSSFetcherImpl 是 RSSFetcher 接口的实现
// 关键改进：将连接池、发布器作为接口注入，降低耦合
type RSSFetcherImpl struct {
	// 仓储层（接口注入，便于测试）
	sourceRepo  domain.SourceRepository
	contentRepo domain.ContentRepository

	// 服务层（接口注入）
	publisher     domain.StreamPublisher
	deduplicator  domain.ContentDeduplicator

	// 配置参数
	workerCount  int
	fetchTimeout time.Duration
	maxRetries   int

	// RSS 解析器
	parser *utils.RSSParser

	// 生命周期管理
	ticker   *time.Ticker
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// NewRSSFetcher 工厂函数：创建 RSS 抓取器
// 这是依赖注入的关键点 - 所有依赖都通过参数传入
func NewRSSFetcher(
	sourceRepo domain.SourceRepository,
	contentRepo domain.ContentRepository,
	publisher domain.StreamPublisher,
	deduplicator domain.ContentDeduplicator,
	workerCount int,
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

// Start 启动 RSS 抓取服务
func (rf *RSSFetcherImpl) Start(ctx context.Context, interval time.Duration) error {
	if err := rf.deduplicator.InitializeBloomFilter(ctx); err != nil {
		log.Printf("Warning: Failed to initialize bloom filter: %v", err)
	}

	rf.ticker = time.NewTicker(interval)
	rf.wg.Add(1)
	go rf.run(ctx)

	return nil
}

// Stop 优雅关闭服务
func (rf *RSSFetcherImpl) Stop() {
	close(rf.stopChan)
	if rf.ticker != nil {
		rf.ticker.Stop()
	}
	rf.wg.Wait()
}

// run 主循环
func (rf *RSSFetcherImpl) run(ctx context.Context) {
	defer rf.wg.Done()

	// 启动时立即抓取一次
	rf.fetchAllSources(ctx)

	for {
		select {
		case <-rf.stopChan:
			return
		case <-rf.ticker.C:
			rf.fetchAllSources(ctx)
		}
	}
}

// fetchAllSources 抓取所有需要更新的数据源
func (rf *RSSFetcherImpl) fetchAllSources(ctx context.Context) {
	sources, err := rf.sourceRepo.GetAll(ctx, true)
	if err != nil {
		log.Printf("Error fetching sources: %v", err)
		return
	}

	toFetch := rf.filterSourcesToFetch(sources)
	if len(toFetch) == 0 {
		return
	}

	// 使用 Worker 池模式进行并发抓取
	rf.processSourcesWithWorkerPool(ctx, toFetch)
}

// filterSourcesToFetch 筛选需要抓取的数据源
func (rf *RSSFetcherImpl) filterSourcesToFetch(sources []*models.Source) []*models.Source {
	var toFetch []*models.Source
	now := time.Now()

	for _, source := range sources {
		shouldFetch := source.LastFetchTime == nil ||
			now.Sub(*source.LastFetchTime) > time.Duration(source.FetchIntervalSeconds)*time.Second

		if shouldFetch {
			toFetch = append(toFetch, source)
		}
	}

	return toFetch
}

// processSourcesWithWorkerPool 使用 Worker 池处理数据源
// 这是并发优化的核心：WorkerCount 由配置驱动（P0 优化）
func (rf *RSSFetcherImpl) processSourcesWithWorkerPool(ctx context.Context, sources []*models.Source) {
	sourceChan := make(chan *models.Source, len(sources))
	rf.wg.Add(rf.workerCount)

	// 启动 Worker（数量由 P0 配置决定）
	for i := 0; i < rf.workerCount; i++ {
		go rf.fetchWorker(ctx, sourceChan)
	}

	// 分发任务
	for _, source := range sources {
		sourceChan <- source
	}
	close(sourceChan)

	// 等待所有 Worker 完成
	rf.wg.Wait()
}

// fetchWorker Worker 线程：从通道读取任务
func (rf *RSSFetcherImpl) fetchWorker(ctx context.Context, sourceChan chan *models.Source) {
	defer rf.wg.Done()

	for source := range sourceChan {
		rf.fetchSourceWithRetry(ctx, source)
	}
}

// fetchSourceWithRetry 带重试的数据源抓取
func (rf *RSSFetcherImpl) fetchSourceWithRetry(ctx context.Context, source *models.Source) {
	fetchCtx, cancel := context.WithTimeout(ctx, rf.fetchTimeout)
	defer cancel()

	var lastErr error

	for attempt := 1; attempt <= rf.maxRetries; attempt++ {
		items, err := rf.parser.ParseFeed(source.URL)
		if err != nil {
			lastErr = err
			log.Printf("Attempt %d: Failed to fetch %s: %v", attempt, source.URL, err)
			continue
		}

		// 处理抓取到的项目
		for _, item := range items {
			rf.processItem(fetchCtx, source, item)
		}

		// 更新最后抓取时间
		if err := rf.sourceRepo.UpdateLastFetchTime(ctx, source.ID, time.Now()); err != nil {
			log.Printf("Failed to update last_fetch_time for source %d: %v", source.ID, err)
		}

		log.Printf("Successfully fetched %s (%d items)", source.URL, len(items))
		return
	}

	log.Printf("Failed to fetch %s after %d attempts: %v", source.URL, rf.maxRetries, lastErr)
}

// processItem 处理单个 RSS 项目
func (rf *RSSFetcherImpl) processItem(ctx context.Context, source *models.Source, item *utils.FeedItem) {
	// 数据清理
	item = utils.SanitizeFeedItem(item)

	// 三层去重检查
	contentHash, isDuplicate, err := rf.deduplicator.ValidateContent(
		ctx, item.URL, item.Title, item.Content,
	)
	if err != nil {
		log.Printf("Error validating content: %v", err)
		return
	}

	if isDuplicate {
		return
	}

	// 创建内容记录
	req := &models.CreateContentRequest{
		SourceID:     source.ID,
		Platform:     source.Platform,
		AuthorName:   item.Author,
		Title:        item.Title,
		OriginalURL:  item.URL,
		ContentHash:  contentHash,
		CleanContent: item.Content,
		PublishedAt:  item.PublishedAt,
	}

	content, err := rf.contentRepo.Create(ctx, req)
	if err != nil {
		log.Printf("Note: Could not create content (may be duplicate): %v", err)
		return
	}

	// 标记为已见
	if err := rf.deduplicator.MarkAsSeen(ctx, item.URL, contentHash); err != nil {
		log.Printf("Warning: Failed to mark URL as seen: %v", err)
	}

	// 发布到 Stream 进行评估（关键优化：接口调用）
	if err := rf.publisher.PublishToStream(ctx, content); err != nil {
		log.Printf("Error publishing to stream: %v", err)
		rf.contentRepo.UpdateStatus(ctx, content.ID, "PROCESSING")
	}

	log.Printf("Ingested: %s (ID: %d)", item.Title, content.ID)
}

// FetchSourceOnDemand 按需抓取单个数据源
func (rf *RSSFetcherImpl) FetchSourceOnDemand(ctx context.Context, sourceID int64) error {
	source, err := rf.sourceRepo.GetByID(ctx, sourceID)
	if err != nil {
		return err
	}

	if source == nil {
		return nil
	}

	rf.fetchSourceWithRetry(ctx, source)
	return nil
}
