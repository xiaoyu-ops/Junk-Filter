package models

import (
	"time"

	"github.com/google/uuid"
)

// Content represents an article/feed item
type Content struct {
	ID           int64
	TaskID       uuid.UUID
	SourceID     int64
	Platform     string
	AuthorName   string
	Title        string
	OriginalURL  string
	ContentHash  string
	CleanContent string
	PublishedAt  *time.Time
	IngestedAt   time.Time
	Status       string // PENDING, PROCESSING, EVALUATED, DISCARDED
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// CreateContentRequest is the request body for creating content
type CreateContentRequest struct {
	SourceID     int64   `json:"source_id" binding:"required"`
	Platform     string  `json:"platform"`
	AuthorName   string  `json:"author_name"`
	Title        string  `json:"title" binding:"required"`
	OriginalURL  string  `json:"original_url" binding:"required"`
	ContentHash  string  `json:"content_hash"`
	CleanContent string  `json:"clean_content" binding:"required"`
	PublishedAt  *time.Time `json:"published_at"`
}

// ContentResponse is the response body for content
type ContentResponse struct {
	ID           int64      `json:"id"`
	TaskID       string     `json:"task_id"`
	SourceID     int64      `json:"source_id"`
	Platform     string     `json:"platform"`
	AuthorName   string     `json:"author_name"`
	Title        string     `json:"title"`
	OriginalURL  string     `json:"original_url"`
	ContentHash  string     `json:"content_hash"`
	CleanContent string     `json:"clean_content"`
	PublishedAt  *time.Time `json:"published_at"`
	IngestedAt   time.Time  `json:"ingested_at"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// ContentFilter for querying content
type ContentFilter struct {
	Status   string `form:"status"`
	SourceID int64  `form:"source_id"`
	Limit    int    `form:"limit,default=50"`
	Offset   int    `form:"offset,default=0"`
}

func (c *Content) ToResponse() *ContentResponse {
	return &ContentResponse{
		ID:           c.ID,
		TaskID:       c.TaskID.String(),
		SourceID:     c.SourceID,
		Platform:     c.Platform,
		AuthorName:   c.AuthorName,
		Title:        c.Title,
		OriginalURL:  c.OriginalURL,
		ContentHash:  c.ContentHash,
		CleanContent: c.CleanContent,
		PublishedAt:  c.PublishedAt,
		IngestedAt:   c.IngestedAt,
		Status:       c.Status,
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
	}
}
