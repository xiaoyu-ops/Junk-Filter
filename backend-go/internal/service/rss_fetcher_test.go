package service

import (
	"context"
	"testing"
	"time"

	"github.com/junkfilter/backend-go/models"
)

// ============================================================================
// RSSFetcher 单元测试
// ============================================================================

func TestNewRSSFetcher(t *testing.T) {
	mockSourceRepo := &MockSourceRepository{}
	mockContentRepo := &MockContentRepository{}
	mockPublisher := &MockStreamPublisher{}
	mockDedup := &MockContentDeduplicator{}

	fetcher := NewRSSFetcher(
		mockSourceRepo,
		mockContentRepo,
		mockPublisher,
		mockDedup,
		20,              // workerCount (P0 optimized)
		30*time.Second,  // fetchTimeout (P0 optimized)
		3,               // maxRetries
	)

	if fetcher == nil {
		t.Fatal("Expected fetcher to be non-nil")
	}

	t.Log("✓ TestNewRSSFetcher: 通过")
}

// ============================================================================

func TestRSSFetcherWithDependencyInjection(t *testing.T) {
	// 测试依赖注入的核心概念：
	// 所有依赖都通过接口传入，可以轻松替换实现

	mockSourceRepo := &MockSourceRepository{
		GetAllFunc: func(ctx context.Context, enabled bool) ([]*models.Source, error) {
			return []*models.Source{
				{
					ID:                   1,
					URL:                  "https://example.com/feed",
					FetchIntervalSeconds: 3600,
					Priority:             5,
					Enabled:              true,
				},
			}, nil
		},
	}

	mockContentRepo := &MockContentRepository{
		CreateFunc: func(ctx context.Context, req *models.CreateContentRequest) (*models.Content, error) {
			return &models.Content{
				ID:        1,
				Title:     req.Title,
				Status:    "PENDING",
				CreatedAt: time.Now(),
			}, nil
		},
	}

	mockPublisher := &MockStreamPublisher{
		PublishToStreamFunc: func(ctx context.Context, content *models.Content) error {
			return nil
		},
	}

	mockDedup := &MockContentDeduplicator{
		ValidateContentFunc: func(ctx context.Context, url, title, content string) (string, bool, error) {
			// 模拟：不重复
			return "hash_" + url, false, nil
		},
	}

	fetcher := NewRSSFetcher(
		mockSourceRepo,
		mockContentRepo,
		mockPublisher,
		mockDedup,
		5,
		10*time.Second,
		3,
	)

	if fetcher == nil {
		t.Fatal("Expected fetcher to be non-nil")
	}

	// 验证依赖注入正确
	// （这是一个简单验证，实际测试需要访问内部字段或通过行为验证）
	// TODO: 添加更多验证逻辑，比如调用 Start() 并验证行为

	t.Log("✓ TestRSSFetcherWithDependencyInjection: 通过")
}

// ============================================================================

func TestRSSFetcherStop(t *testing.T) {
	mockSourceRepo := &MockSourceRepository{}
	mockContentRepo := &MockContentRepository{}
	mockPublisher := &MockStreamPublisher{}
	mockDedup := &MockContentDeduplicator{}

	fetcher := NewRSSFetcher(
		mockSourceRepo,
		mockContentRepo,
		mockPublisher,
		mockDedup,
		5,
		10*time.Second,
		3,
	)

	// Stop 应该不会 panic
	fetcher.Stop()

	t.Log("✓ TestRSSFetcherStop: 通过")
}

// ============================================================================

func TestRSSFetcherWithVariousConfigs(t *testing.T) {
	// 表驱动测试：验证不同的配置参数组合
	tests := []struct {
		name        string
		workerCount int
		timeout     time.Duration
		maxRetries  int
		description string
	}{
		{
			name:        "P0 优化配置",
			workerCount: 20,      // P0 optimized
			timeout:     30*time.Second, // P0 optimized
			maxRetries:  3,
			description: "生产环境推荐配置",
		},
		{
			name:        "保守配置",
			workerCount: 5,
			timeout:     10*time.Second,
			maxRetries:  2,
			description: "开发/测试环境",
		},
		{
			name:        "激进配置",
			workerCount: 50,
			timeout:     60*time.Second,
			maxRetries:  5,
			description: "高并发场景",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSourceRepo := &MockSourceRepository{}
			mockContentRepo := &MockContentRepository{}
			mockPublisher := &MockStreamPublisher{}
			mockDedup := &MockContentDeduplicator{}

			fetcher := NewRSSFetcher(
				mockSourceRepo,
				mockContentRepo,
				mockPublisher,
				mockDedup,
				tt.workerCount,
				tt.timeout,
				tt.maxRetries,
			)

			if fetcher == nil {
				t.Errorf("%s: expected fetcher to be non-nil", tt.description)
			}

			t.Logf("✓ %s: workerCount=%d, timeout=%v", tt.description, tt.workerCount, tt.timeout)
		})
	}
}

// ============================================================================

func TestMockSourceRepository(t *testing.T) {
	mockRepo := &MockSourceRepository{
		GetAllFunc: func(ctx context.Context, enabled bool) ([]*models.Source, error) {
			return []*models.Source{
				{ID: 1, URL: "https://example.com/feed1"},
				{ID: 2, URL: "https://example.com/feed2"},
			}, nil
		},
	}

	ctx := context.Background()
	sources, err := mockRepo.GetAll(ctx, true)

	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}

	if len(sources) != 2 {
		t.Errorf("Expected 2 sources, got %d", len(sources))
	}

	if mockRepo.GetAllCalled != 1 {
		t.Errorf("Expected GetAll to be called 1 time, got %d", mockRepo.GetAllCalled)
	}

	t.Log("✓ TestMockSourceRepository: 通过")
}

// ============================================================================

func TestMockStreamPublisher(t *testing.T) {
	mockPublisher := &MockStreamPublisher{}

	ctx := context.Background()
	content := &models.Content{
		ID:           1,
		Title:        "Test Article",
		CleanContent: "Test body",
		Status:       "PENDING",
		CreatedAt:    time.Now(),
	}

	err := mockPublisher.PublishToStream(ctx, content)
	if err != nil {
		t.Fatalf("PublishToStream failed: %v", err)
	}

	if mockPublisher.PublishToStreamCalled != 1 {
		t.Errorf("Expected PublishToStream to be called 1 time, got %d", mockPublisher.PublishToStreamCalled)
	}

	if len(mockPublisher.PublishedMessages) != 1 {
		t.Errorf("Expected 1 published message, got %d", len(mockPublisher.PublishedMessages))
	}

	if mockPublisher.PublishedMessages[0].ID != 1 {
		t.Errorf("Expected published message ID to be 1, got %d", mockPublisher.PublishedMessages[0].ID)
	}

	t.Log("✓ TestMockStreamPublisher: 通过")
}

// ============================================================================

func TestMockContentDeduplicator(t *testing.T) {
	mockDedup := &MockContentDeduplicator{
		ValidateContentFunc: func(ctx context.Context, url, title, content string) (string, bool, error) {
			// 模拟去重逻辑：如果 URL 包含 "duplicate" 则返回重复
			isDuplicate := false
			if url == "https://duplicate.com/feed" {
				isDuplicate = true
			}
			return "hash_" + url, isDuplicate, nil
		},
	}

	ctx := context.Background()

	// 测试非重复内容
	hash1, isDuplicate1, err1 := mockDedup.ValidateContent(ctx, "https://unique.com/feed", "Title", "Body")
	if err1 != nil {
		t.Fatalf("ValidateContent failed: %v", err1)
	}
	if isDuplicate1 {
		t.Error("Expected non-duplicate content")
	}
	if hash1 == "" {
		t.Error("Expected hash to be non-empty")
	}

	// 测试重复内容
	hash2, isDuplicate2, err2 := mockDedup.ValidateContent(ctx, "https://duplicate.com/feed", "Title", "Body")
	if err2 != nil {
		t.Fatalf("ValidateContent failed: %v", err2)
	}
	if !isDuplicate2 {
		t.Error("Expected duplicate content")
	}

	if mockDedup.ValidateContentCalled != 2 {
		t.Errorf("Expected ValidateContent to be called 2 times, got %d", mockDedup.ValidateContentCalled)
	}

	t.Logf("✓ TestMockContentDeduplicator: 通过 (hash1=%s, hash2=%s)", hash1, hash2)
}

// ============================================================================

func BenchmarkNewRSSFetcher(b *testing.B) {
	mockSourceRepo := &MockSourceRepository{}
	mockContentRepo := &MockContentRepository{}
	mockPublisher := &MockStreamPublisher{}
	mockDedup := &MockContentDeduplicator{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewRSSFetcher(
			mockSourceRepo,
			mockContentRepo,
			mockPublisher,
			mockDedup,
			20,
			30*time.Second,
			3,
		)
	}
}
