package service

import (
	"context"
	"time"

	"github.com/junkfilter/backend-go/internal/domain"
	"github.com/junkfilter/backend-go/models"
)

// ============================================================================
// Mock 实现 - 用于单元测试
// ============================================================================

// MockSourceRepository 是 SourceRepository 接口的 mock 实现
type MockSourceRepository struct {
	GetAllFunc              func(ctx context.Context, enabled bool) ([]*models.Source, error)
	GetByIDFunc             func(ctx context.Context, id int64) (*models.Source, error)
	UpdateLastFetchTimeFunc func(ctx context.Context, id int64, t time.Time) error

	// 调用追踪
	GetAllCalled              int
	GetByIDCalled             int
	UpdateLastFetchTimeCalled int
}

func (m *MockSourceRepository) GetAll(ctx context.Context, enabled bool) ([]*models.Source, error) {
	m.GetAllCalled++
	if m.GetAllFunc != nil {
		return m.GetAllFunc(ctx, enabled)
	}
	return []*models.Source{}, nil
}

func (m *MockSourceRepository) GetByID(ctx context.Context, id int64) (*models.Source, error) {
	m.GetByIDCalled++
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockSourceRepository) UpdateLastFetchTime(ctx context.Context, id int64, t time.Time) error {
	m.UpdateLastFetchTimeCalled++
	if m.UpdateLastFetchTimeFunc != nil {
		return m.UpdateLastFetchTimeFunc(ctx, id, t)
	}
	return nil
}

// ============================================================================

// MockContentRepository 是 ContentRepository 接口的 mock 实现
type MockContentRepository struct {
	CreateFunc        func(ctx context.Context, req *models.CreateContentRequest) (*models.Content, error)
	UpdateStatusFunc  func(ctx context.Context, id int64, status string) error
	GetByHashFunc     func(ctx context.Context, hash string) (*models.Content, error)

	// 调用追踪
	CreateCalled       int
	UpdateStatusCalled int
	GetByHashCalled    int
}

func (m *MockContentRepository) Create(ctx context.Context, req *models.CreateContentRequest) (*models.Content, error) {
	m.CreateCalled++
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, req)
	}
	return &models.Content{
		ID:     int64(m.CreateCalled),
		Status: "PENDING",
	}, nil
}

func (m *MockContentRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	m.UpdateStatusCalled++
	if m.UpdateStatusFunc != nil {
		return m.UpdateStatusFunc(ctx, id, status)
	}
	return nil
}

func (m *MockContentRepository) GetByHash(ctx context.Context, hash string) (*models.Content, error) {
	m.GetByHashCalled++
	if m.GetByHashFunc != nil {
		return m.GetByHashFunc(ctx, hash)
	}
	return nil, nil
}

// ============================================================================

// MockStreamPublisher 是 StreamPublisher 接口的 mock 实现
type MockStreamPublisher struct {
	PublishToStreamFunc  func(ctx context.Context, content *models.Content) error
	GetStreamPendingFunc func(ctx context.Context) (int64, error)

	// 调用追踪
	PublishToStreamCalled  int
	GetStreamPendingCalled int
	PublishedMessages      []*models.Content
}

func (m *MockStreamPublisher) PublishToStream(ctx context.Context, content *models.Content) error {
	m.PublishToStreamCalled++
	m.PublishedMessages = append(m.PublishedMessages, content)
	if m.PublishToStreamFunc != nil {
		return m.PublishToStreamFunc(ctx, content)
	}
	return nil
}

func (m *MockStreamPublisher) GetStreamPending(ctx context.Context) (int64, error) {
	m.GetStreamPendingCalled++
	if m.GetStreamPendingFunc != nil {
		return m.GetStreamPendingFunc(ctx)
	}
	return 0, nil
}

// ============================================================================

// MockContentDeduplicator 是 ContentDeduplicator 接口的 mock 实现
type MockContentDeduplicator struct {
	InitializeBloomFilterFunc func(ctx context.Context) error
	ValidateContentFunc       func(ctx context.Context, url, title, content string) (string, bool, error)
	MarkAsSeenFunc            func(ctx context.Context, url, hash string) error

	// 调用追踪
	InitializeBloomFilterCalled int
	ValidateContentCalled       int
	MarkAsSeenCalled            int
}

func (m *MockContentDeduplicator) InitializeBloomFilter(ctx context.Context) error {
	m.InitializeBloomFilterCalled++
	if m.InitializeBloomFilterFunc != nil {
		return m.InitializeBloomFilterFunc(ctx)
	}
	return nil
}

func (m *MockContentDeduplicator) ValidateContent(ctx context.Context, url, title, content string) (string, bool, error) {
	m.ValidateContentCalled++
	if m.ValidateContentFunc != nil {
		return m.ValidateContentFunc(ctx, url, title, content)
	}
	// 默认返回：hash、is_duplicate(false)、error(nil)
	return "mock_hash_" + url, false, nil
}

func (m *MockContentDeduplicator) MarkAsSeen(ctx context.Context, url, hash string) error {
	m.MarkAsSeenCalled++
	if m.MarkAsSeenFunc != nil {
		return m.MarkAsSeenFunc(ctx, url, hash)
	}
	return nil
}

// ============================================================================

// NewMockFactory 创建一个带有 mock 实现的工厂（用于测试）
func NewMockFactory() *MockFactory {
	return &MockFactory{
		sourceRepo:    &MockSourceRepository{},
		contentRepo:   &MockContentRepository{},
		publisher:     &MockStreamPublisher{},
		deduplicator:  &MockContentDeduplicator{},
	}
}

// MockFactory 是 Factory 的 mock 版本（用于测试 main 或 app 层）
type MockFactory struct {
	sourceRepo    domain.SourceRepository
	contentRepo   domain.ContentRepository
	publisher     domain.StreamPublisher
	deduplicator  domain.ContentDeduplicator
	closeCalled   bool
}

func (mf *MockFactory) CreateRSSFetcher() (domain.RSSFetcher, error) {
	return NewRSSFetcher(
		mf.sourceRepo,
		mf.contentRepo,
		mf.publisher,
		mf.deduplicator,
		5,
		10*time.Second,
		3,
	), nil
}

func (mf *MockFactory) CreateStreamPublisher() (domain.StreamPublisher, error) {
	return mf.publisher, nil
}

func (mf *MockFactory) SourceRepo() domain.SourceRepository {
	return mf.sourceRepo
}

func (mf *MockFactory) ContentRepo() domain.ContentRepository {
	return mf.contentRepo
}

func (mf *MockFactory) Publisher() domain.StreamPublisher {
	return mf.publisher
}

func (mf *MockFactory) Deduplicator() domain.ContentDeduplicator {
	return mf.deduplicator
}

func (mf *MockFactory) Close() error {
	mf.closeCalled = true
	return nil
}
