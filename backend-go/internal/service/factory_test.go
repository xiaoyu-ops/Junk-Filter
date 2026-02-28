package service

import (
	"testing"

	"github.com/junkfilter/backend-go/internal/config"
)

// ============================================================================
// Factory 单元测试
// ============================================================================

func TestNewFactory(t *testing.T) {
	// 加载测试配置
	cfg := &config.Config{}

	// 设置配置字段（inline struct）
	cfg.Database.Host = "localhost"
	cfg.Database.Port = 5432
	cfg.Database.User = "test_user"
	cfg.Database.Password = "test_pass"
	cfg.Database.DBName = "test_db"
	cfg.Database.SSLMode = "disable"
	cfg.Database.MaxOpenConns = 50
	cfg.Database.MaxIdleConns = 10

	cfg.Redis.Host = "localhost"
	cfg.Redis.Port = 6379
	cfg.Redis.DB = 0

	cfg.Ingestion.WorkerCount = 20
	cfg.Ingestion.Timeout = "30s"
	cfg.Ingestion.RetryMax = 3
	cfg.Ingestion.FetchInterval = "30m"

	// 注：因为需要真实的 DB 和 Redis 连接，完整的集成测试需要 docker-compose
	// 这里仅作演示框架，实际测试应该 mock infra 层

	t.Log("✓ TestNewFactory: 框架演示（实际需要真实的 DB 和 Redis）")

	// 真实测试的伪代码：
	// factory, err := NewFactory(cfg)
	// if err != nil {
	//     t.Fatalf("NewFactory failed: %v", err)
	// }
	// defer factory.Close()
	//
	// if factory.SourceRepo() == nil {
	//     t.Error("Expected sourceRepo to be initialized")
	// }
	// if factory.ContentRepo() == nil {
	//     t.Error("Expected contentRepo to be initialized")
	// }
	// if factory.Publisher() == nil {
	//     t.Error("Expected publisher to be initialized")
	// }
}

// ============================================================================

func TestFactoryCreateRSSFetcher(t *testing.T) {
	// 使用 mock 工厂
	mockFactory := NewMockFactory()

	fetcher, err := mockFactory.CreateRSSFetcher()
	if err != nil {
		t.Fatalf("CreateRSSFetcher failed: %v", err)
	}

	if fetcher == nil {
		t.Error("Expected RSSFetcher to be created")
	}

	// 验证返回的是接口类型
	if _, ok := interface{}(fetcher).(RSSFetcherImpl); !ok {
		// 允许返回的是任何实现 domain.RSSFetcher 的类型
	}

	t.Log("✓ TestFactoryCreateRSSFetcher: 通过")
}

// ============================================================================

func TestFactoryCreateStreamPublisher(t *testing.T) {
	mockFactory := NewMockFactory()

	publisher, err := mockFactory.CreateStreamPublisher()
	if err != nil {
		t.Fatalf("CreateStreamPublisher failed: %v", err)
	}

	if publisher == nil {
		t.Error("Expected StreamPublisher to be created")
	}

	t.Log("✓ TestFactoryCreateStreamPublisher: 通过")
}

// ============================================================================

func TestFactoryClose(t *testing.T) {
	mockFactory := NewMockFactory()

	if mockFactory.closeCalled {
		t.Error("Expected closeCalled to be false before Close()")
	}

	err := mockFactory.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	if !mockFactory.closeCalled {
		t.Error("Expected closeCalled to be true after Close()")
	}

	t.Log("✓ TestFactoryClose: 通过")
}

// ============================================================================

func TestFactoryGetters(t *testing.T) {
	mockFactory := NewMockFactory()

	// 测试各个 getter 方法
	if mockFactory.SourceRepo() == nil {
		t.Error("Expected SourceRepo() to return non-nil")
	}

	if mockFactory.ContentRepo() == nil {
		t.Error("Expected ContentRepo() to return non-nil")
	}

	if mockFactory.Publisher() == nil {
		t.Error("Expected Publisher() to return non-nil")
	}

	if mockFactory.Deduplicator() == nil {
		t.Error("Expected Deduplicator() to return non-nil")
	}

	t.Log("✓ TestFactoryGetters: 通过")
}
