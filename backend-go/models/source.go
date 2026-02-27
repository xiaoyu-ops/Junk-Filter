package models

import "time"

// Source represents an RSS feed source
type Source struct {
	ID                   int64
	Platform             string
	URL                  string
	AuthorName           string
	AuthorID             *string // 改为指针，允许 NULL
	Priority             int
	LastFetchTime        *time.Time
	FetchIntervalSeconds int
	Enabled              bool
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// CreateSourceRequest is the request body for creating a source
type CreateSourceRequest struct {
	URL          string `json:"url" binding:"required"`
	AuthorName   string `json:"author_name"`
	Priority     int    `json:"priority" binding:"min=1,max=10"`
	FetchIntervalSeconds int `json:"fetch_interval_seconds"`
	Platform     string `json:"platform"`
}

// UpdateSourceRequest is the request body for updating a source
type UpdateSourceRequest struct {
	AuthorName   string `json:"author_name"`
	Priority     int    `json:"priority"`
	Enabled      bool   `json:"enabled"`
	FetchIntervalSeconds int `json:"fetch_interval_seconds"`
}

// SourceResponse is the response body for a source
type SourceResponse struct {
	ID                   int64      `json:"id"`
	Platform             string     `json:"platform"`
	URL                  string     `json:"url"`
	AuthorName           string     `json:"author_name"`
	Priority             int        `json:"priority"`
	LastFetchTime        *time.Time `json:"last_fetch_time"`
	FetchIntervalSeconds int        `json:"fetch_interval_seconds"`
	Enabled              bool       `json:"enabled"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

func (s *Source) ToResponse() *SourceResponse {
	return &SourceResponse{
		ID:                   s.ID,
		Platform:             s.Platform,
		URL:                  s.URL,
		AuthorName:           s.AuthorName,
		Priority:             s.Priority,
		LastFetchTime:        s.LastFetchTime,
		FetchIntervalSeconds: s.FetchIntervalSeconds,
		Enabled:              s.Enabled,
		CreatedAt:            s.CreatedAt,
		UpdatedAt:            s.UpdatedAt,
	}
}
