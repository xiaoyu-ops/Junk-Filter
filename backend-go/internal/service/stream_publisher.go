package service

import (
	"context"
	"encoding/json"

	"github.com/go-redis/redis/v8"
	"github.com/junkfilter/backend-go/internal/domain"
	"github.com/junkfilter/backend-go/models"
)

// ============================================================================
// StreamPublisherImpl - Redis Stream 发布器实现
// ============================================================================

// StreamPublisherImpl 是 StreamPublisher 接口的实现
type StreamPublisherImpl struct {
	redis *redis.Client
}

// NewStreamPublisher 工厂函数：创建 Stream 发布器
func NewStreamPublisher(redis *redis.Client) domain.StreamPublisher {
	return &StreamPublisherImpl{
		redis: redis,
	}
}

// PublishToStream 发布内容到 Redis Stream
// 关键点：接口隔离，便于测试和替换
func (sp *StreamPublisherImpl) PublishToStream(ctx context.Context, content *models.Content) error {
	message := &models.StreamMessage{
		ContentID:   content.ID,
		TaskID:      content.TaskID.String(),
		Title:       content.Title,
		URL:         content.OriginalURL,
		Content:     content.CleanContent,
		PublishedAt: content.PublishedAt.Format("2006-01-02T15:04:05Z"),
		Platform:    content.Platform,
		AuthorName:  content.AuthorName,
		ContentHash: content.ContentHash,
	}

	// 序列化为 JSON
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// 添加到 Redis Stream（消息队列）
	err = sp.redis.XAdd(ctx, &redis.XAddArgs{
		Stream: "ingestion_queue",
		Values: map[string]interface{}{
			"data": string(data),
		},
	}).Err()

	return err
}

// GetStreamPending 获取待处理消息数
func (sp *StreamPublisherImpl) GetStreamPending(ctx context.Context) (int64, error) {
	info := sp.redis.XLen(ctx, "ingestion_queue")
	return info.Val(), info.Err()
}
