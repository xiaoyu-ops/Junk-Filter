package services

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/junkfilter/backend-go/models"
	"github.com/junkfilter/backend-go/repositories"
	"github.com/junkfilter/backend-go/utils"
)

// RSSService handles RSS fetching and processing
type RSSService struct {
	parser          *utils.RSSParser
	sourceRepo      *repositories.SourceRepository
	contentRepo     *repositories.ContentRepository
	dedupService    *DedupService
	contentService  *ContentService
	redis           *redis.Client
	workerCount     int
	fetchTimeout    time.Duration
	maxRetries      int
	ticker          *time.Ticker
	stopChan        chan struct{}
	wg              sync.WaitGroup
}

// NewRSSService creates a new RSS service
func NewRSSService(
	sourceRepo *repositories.SourceRepository,
	contentRepo *repositories.ContentRepository,
	redis *redis.Client,
	contentService *ContentService,
	workerCount int,
	fetchTimeout time.Duration,
	maxRetries int,
	proxyURL string,
) *RSSService {
	return &RSSService{
		parser:         utils.NewRSSParser(proxyURL),
		sourceRepo:     sourceRepo,
		contentRepo:    contentRepo,
		dedupService:   NewDedupService(redis, contentRepo),
		contentService: contentService,
		redis:          redis,
		workerCount:    workerCount,
		fetchTimeout:   fetchTimeout,
		maxRetries:     maxRetries,
		stopChan:       make(chan struct{}),
	}
}

// Start begins the RSS fetching service.
// The ticker polls every 30s, but each source is only fetched when its own
// fetch_interval_seconds has elapsed — one ticker serves all sources with different intervals.
func (rs *RSSService) Start(ctx context.Context, interval time.Duration) error {
	// Initialize bloom filter
	if err := rs.dedupService.InitializeBloomFilter(ctx); err != nil {
		log.Printf("Warning: Failed to initialize bloom filter: %v", err)
	}

	// 使用更小的轮询间隔（30秒），以支持更频繁的更新需求
	// 这样即使用户设置 30 分钟更新，也能及时响应
	minInterval := 30 * time.Second
	if interval < minInterval {
		minInterval = interval
	}

	rs.ticker = time.NewTicker(minInterval)
	rs.wg.Add(1)
	go rs.run(ctx)
	log.Printf("✓ RSS Service started with poll interval: %v (configured: %v)", minInterval, interval)
	return nil
}

// Stop stops the RSS fetching service
func (rs *RSSService) Stop() {
	close(rs.stopChan)
	if rs.ticker != nil {
		rs.ticker.Stop()
	}
	rs.wg.Wait()
}

func (rs *RSSService) run(ctx context.Context) {
	defer rs.wg.Done()

	// Fetch immediately on start
	rs.fetchAllSources(ctx)

	for {
		select {
		case <-rs.stopChan:
			return
		case <-rs.ticker.C:
			rs.fetchAllSources(ctx)
		}
	}
}

func (rs *RSSService) fetchAllSources(ctx context.Context) {
	sources, err := rs.sourceRepo.GetAll(ctx, true)
	if err != nil {
		log.Printf("Error fetching sources: %v", err)
		return
	}

	// Check which sources need fetching based on last_fetch_time and interval
	toFetch := rs.filterSourcesToFetch(sources)

	if len(toFetch) == 0 {
		return
	}

	// Worker pool pattern - use a local WaitGroup to avoid race with rs.wg
	sourceChan := make(chan *models.Source, len(toFetch))
	var localWg sync.WaitGroup
	localWg.Add(rs.workerCount)

	for i := 0; i < rs.workerCount; i++ {
		go func() {
			defer localWg.Done()
			for source := range sourceChan {
				rs.fetchSource(ctx, source)
			}
		}()
	}

	for _, source := range toFetch {
		sourceChan <- source
	}
	close(sourceChan)

	localWg.Wait()
}

func (rs *RSSService) filterSourcesToFetch(sources []*models.Source) []*models.Source {
	var toFetch []*models.Source
	now := time.Now()

	for _, source := range sources {
		shouldFetch := source.LastFetchTime == nil || // Never fetched
			now.Sub(*source.LastFetchTime) > time.Duration(source.FetchIntervalSeconds)*time.Second

		if shouldFetch {
			toFetch = append(toFetch, source)
		}
	}

	return toFetch
}

func (rs *RSSService) fetchSource(ctx context.Context, source *models.Source) {
	// Create a timeout context for this fetch
	fetchCtx, cancel := context.WithTimeout(ctx, rs.fetchTimeout)
	defer cancel()

	var lastErr error
	for attempt := 1; attempt <= rs.maxRetries; attempt++ {
		if attempt > 1 {
			backoff := time.Duration(attempt-1) * 2 * time.Second
			time.Sleep(backoff)
		}
		items, err := rs.parser.ParseFeed(source.URL)
		if err != nil {
			lastErr = err
			log.Printf("Attempt %d: Failed to fetch %s: %v", attempt, source.URL, err)
			continue
		}

		// Process items
		for _, item := range items {
			rs.processItem(fetchCtx, source, item)
		}

		// Update last fetch time
		if err := rs.sourceRepo.UpdateLastFetchTime(ctx, source.ID, time.Now()); err != nil {
			log.Printf("Failed to update last_fetch_time for source %d: %v", source.ID, err)
		}

		log.Printf("Successfully fetched %s (%d items)", source.URL, len(items))
		return
	}

	log.Printf("Failed to fetch %s after %d attempts: %v", source.URL, rs.maxRetries, lastErr)
}

func (rs *RSSService) processItem(ctx context.Context, source *models.Source, item *utils.FeedItem) {
	item = utils.SanitizeFeedItem(item)

	// Short-content filter before dedup — skip RSS excerpts with no real body,
	// saving Redis and DB round-trips for content we'd discard anyway
	if len([]rune(item.Content)) < 200 {
		log.Printf("[Skip] Content too short (%d runes), skipping: %s", len([]rune(item.Content)), item.Title)
		return
	}

	// Author filter before dedup — same reason: skip early before any I/O
	if source.ShouldFilterAuthor(item.Author) {
		return
	}

	// Check for duplicates
	contentHash, isDuplicate, err := rs.dedupService.ValidateContent(
		ctx, item.URL, item.Title, item.Content,
	)
	if err != nil {
		log.Printf("Error validating content: %v", err)
		return
	}

	if isDuplicate {
		return // Skip duplicate
	}

	// Create content record
	// Use item author, fallback to source author name
	authorName := item.Author
	if authorName == "" {
		authorName = source.AuthorName
	}

	req := &models.CreateContentRequest{
		SourceID:     source.ID,
		Platform:     source.Platform,
		AuthorName:   authorName,
		Title:        item.Title,
		OriginalURL:  item.URL,
		ContentHash:  contentHash,
		CleanContent: item.Content,
		ImageURLs:    item.ImageURLs,
		PublishedAt:  item.PublishedAt,
	}

	content, err := rs.contentRepo.Create(ctx, req)
	if err != nil {
		// Might be a duplicate from concurrent insert (L3 constraint)
		log.Printf("Note: Could not create content (may be duplicate): %v", err)
		return
	}

	// Mark as seen
	if err := rs.dedupService.MarkAsSeen(ctx, item.URL, contentHash); err != nil {
		log.Printf("Warning: Failed to mark URL as seen: %v", err)
	}

	// Publish to Stream
	if err := rs.contentService.PublishToStream(ctx, content); err != nil {
		log.Printf("Error publishing to stream: %v", err)
		// Keep status as PENDING so it can be retried
		rs.contentRepo.UpdateStatus(ctx, content.ID, "PENDING")
		return
	}

	log.Printf("Ingested: %s (ID: %d)", item.Title, content.ID)
}

// SetProxyURL updates the RSS proxy at runtime
func (rs *RSSService) SetProxyURL(proxyURL string) {
	rs.parser.SetProxyURL(proxyURL)
}

// GetProxyURL returns the current proxy URL
func (rs *RSSService) GetProxyURL() string {
	return rs.parser.GetProxyURL()
}

// FetchSourceOnDemand fetches a specific source immediately
func (rs *RSSService) FetchSourceOnDemand(ctx context.Context, sourceID int64) error {
	source, err := rs.sourceRepo.GetByID(ctx, sourceID)
	if err != nil {
		return err
	}

	if source == nil {
		return nil
	}

	rs.fetchSource(ctx, source)
	return nil
}
