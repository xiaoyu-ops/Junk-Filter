package models

import (
	"database/sql/driver"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// StringArray is a custom type that supports both PostgreSQL text[] arrays and JSON arrays
type StringArray []string

func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return pq.StringArray(nil).Value()
	}
	return pq.StringArray(a).Value()
}

func (a *StringArray) Scan(src interface{}) error {
	if src == nil {
		*a = nil
		return nil
	}

	// Try PostgreSQL native array format first (e.g. {url1,url2})
	var pgArr pq.StringArray
	if err := pgArr.Scan(src); err == nil {
		*a = StringArray(pgArr)
		return nil
	}

	// Fallback: try JSON format (e.g. ["url1","url2"])
	var data []byte
	switch v := src.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		*a = nil
		return nil
	}

	s := strings.TrimSpace(string(data))
	if s == "" || s == "{}" || s == "[]" {
		*a = StringArray{}
		return nil
	}

	return json.Unmarshal(data, a)
}

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
	ImageURLs    StringArray
	PublishedAt  *time.Time
	IngestedAt   time.Time
	Status       string // PENDING, PROCESSING, EVALUATED, DISCARDED
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// CreateContentRequest is the request body for creating content
type CreateContentRequest struct {
	SourceID     int64       `json:"source_id" binding:"required"`
	Platform     string      `json:"platform"`
	AuthorName   string      `json:"author_name"`
	Title        string      `json:"title" binding:"required"`
	OriginalURL  string      `json:"original_url" binding:"required"`
	ContentHash  string      `json:"content_hash"`
	CleanContent string      `json:"clean_content" binding:"required"`
	ImageURLs    StringArray `json:"image_urls"`
	PublishedAt  *time.Time  `json:"published_at"`
}

// ContentResponse is the response body for content
type ContentResponse struct {
	ID           int64       `json:"id"`
	TaskID       string      `json:"task_id"`
	SourceID     int64       `json:"source_id"`
	Platform     string      `json:"platform"`
	AuthorName   string      `json:"author_name"`
	Title        string      `json:"title"`
	OriginalURL  string      `json:"original_url"`
	ContentHash  string      `json:"content_hash"`
	CleanContent string      `json:"clean_content"`
	ImageURLs    StringArray `json:"image_urls"`
	PublishedAt  *time.Time  `json:"published_at"`
	IngestedAt   time.Time   `json:"ingested_at"`
	Status       string      `json:"status"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
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
		ImageURLs:    c.ImageURLs,
		PublishedAt:  c.PublishedAt,
		IngestedAt:   c.IngestedAt,
		Status:       c.Status,
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
	}
}
