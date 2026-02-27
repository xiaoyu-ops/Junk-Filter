package services

import (
	"context"
	"encoding/json"

	"github.com/go-redis/redis/v8"
	"github.com/junkfilter/backend-go/models"
)

// ContentService handles content processing and Stream publishing
type ContentService struct {
	redis *redis.Client
}

// NewContentService creates a new content service
func NewContentService(redis *redis.Client) *ContentService {
	return &ContentService{
		redis: redis,
	}
}

// PublishToStream publishes a content to Redis Stream for evaluation
func (cs *ContentService) PublishToStream(ctx context.Context, content *models.Content) error {
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

	// Marshal to JSON
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// Add to Stream
	err = cs.redis.XAdd(ctx, &redis.XAddArgs{
		Stream: "ingestion_queue",
		Values: map[string]interface{}{
			"data": string(data),
		},
	}).Err()

	return err
}

// GetStreamPending returns pending messages count
func (cs *ContentService) GetStreamPending(ctx context.Context) (int64, error) {
	info := cs.redis.XLen(ctx, "ingestion_queue")
	return info.Val(), info.Err()
}
