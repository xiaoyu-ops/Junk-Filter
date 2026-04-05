package models

import (
	"encoding/json"
	"strings"
	"time"
)

// AuthorFilter defines author-level filtering rules for a source
type AuthorFilter struct {
	Mode    string   `json:"mode"`    // "whitelist", "blacklist", or "" (no filter)
	Authors []string `json:"authors"` // List of author names
}

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
	FaviconURL           *string
	AuthorFilterJSON     *string // raw JSONB from DB
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// GetAuthorFilter parses the author_filter JSONB field
func (s *Source) GetAuthorFilter() AuthorFilter {
	if s.AuthorFilterJSON == nil || *s.AuthorFilterJSON == "" || *s.AuthorFilterJSON == "{}" {
		return AuthorFilter{}
	}
	var af AuthorFilter
	if err := json.Unmarshal([]byte(*s.AuthorFilterJSON), &af); err != nil {
		return AuthorFilter{}
	}
	return af
}

// ShouldFilterAuthor checks if an article author should be filtered out
func (s *Source) ShouldFilterAuthor(author string) bool {
	af := s.GetAuthorFilter()
	if af.Mode == "" || len(af.Authors) == 0 {
		return false // no filter
	}

	found := false
	for _, a := range af.Authors {
		if strings.EqualFold(a, author) {
			found = true
			break
		}
	}

	if af.Mode == "whitelist" {
		return !found // filter out if NOT in whitelist
	}
	// blacklist
	return found // filter out if IN blacklist
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
	ID                   int64         `json:"id"`
	Platform             string        `json:"platform"`
	URL                  string        `json:"url"`
	AuthorName           string        `json:"author_name"`
	Priority             int           `json:"priority"`
	LastFetchTime        *time.Time    `json:"last_fetch_time"`
	FetchIntervalSeconds int           `json:"fetch_interval_seconds"`
	Enabled              bool          `json:"enabled"`
	FaviconURL           *string       `json:"favicon_url"`
	AuthorFilter         *AuthorFilter `json:"author_filter,omitempty"`
	CreatedAt            time.Time     `json:"created_at"`
	UpdatedAt            time.Time     `json:"updated_at"`
}

func (s *Source) ToResponse() *SourceResponse {
	resp := &SourceResponse{
		ID:                   s.ID,
		Platform:             s.Platform,
		URL:                  s.URL,
		AuthorName:           s.AuthorName,
		Priority:             s.Priority,
		LastFetchTime:        s.LastFetchTime,
		FetchIntervalSeconds: s.FetchIntervalSeconds,
		Enabled:              s.Enabled,
		FaviconURL:           s.FaviconURL,
		CreatedAt:            s.CreatedAt,
		UpdatedAt:            s.UpdatedAt,
	}
	af := s.GetAuthorFilter()
	if af.Mode != "" {
		resp.AuthorFilter = &af
	}
	return resp
}
