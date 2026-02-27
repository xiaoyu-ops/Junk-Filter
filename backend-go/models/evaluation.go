package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Evaluation represents an LLM evaluation result
type Evaluation struct {
	ID                int64
	ContentID         int64
	TaskID            uuid.UUID
	InnovationScore   int
	DepthScore        int
	Decision          string
	Reasoning         string
	TLDR              string
	KeyConcepts       pq.StringArray
	EvaluatedAt       time.Time
	EvaluatorVersion  string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// EvaluationResponse is the response body for evaluation
type EvaluationResponse struct {
	ID               int64    `json:"id"`
	ContentID        int64    `json:"content_id"`
	TaskID           string   `json:"task_id"`
	InnovationScore  int      `json:"innovation_score"`
	DepthScore       int      `json:"depth_score"`
	Decision         string   `json:"decision"`
	Reasoning        string   `json:"reasoning"`
	TLDR             string   `json:"tldr"`
	KeyConcepts      []string `json:"key_concepts"`
	EvaluatedAt      time.Time `json:"evaluated_at"`
	EvaluatorVersion string   `json:"evaluator_version"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// EvaluationRequest for storing evaluation
type EvaluationRequest struct {
	ContentID        int64    `json:"content_id" binding:"required"`
	InnovationScore  int      `json:"innovation_score" binding:"min=0,max=10"`
	DepthScore       int      `json:"depth_score" binding:"min=0,max=10"`
	Decision         string   `json:"decision" binding:"required"`
	Reasoning        string   `json:"reasoning"`
	TLDR             string   `json:"tldr"`
	KeyConcepts      []string `json:"key_concepts"`
	EvaluatorVersion string   `json:"evaluator_version"`
}

func (e *Evaluation) ToResponse() *EvaluationResponse {
	concepts := make([]string, len(e.KeyConcepts))
	copy(concepts, e.KeyConcepts)
	return &EvaluationResponse{
		ID:               e.ID,
		ContentID:        e.ContentID,
		TaskID:           e.TaskID.String(),
		InnovationScore:  e.InnovationScore,
		DepthScore:       e.DepthScore,
		Decision:         e.Decision,
		Reasoning:        e.Reasoning,
		TLDR:             e.TLDR,
		KeyConcepts:      concepts,
		EvaluatedAt:      e.EvaluatedAt,
		EvaluatorVersion: e.EvaluatorVersion,
		CreatedAt:        e.CreatedAt,
		UpdatedAt:        e.UpdatedAt,
	}
}

// StreamMessage represents a message in Redis Stream
type StreamMessage struct {
	ContentID    int64  `json:"content_id"`
	TaskID       string `json:"task_id"`
	Title        string `json:"title"`
	URL          string `json:"url"`
	Content      string `json:"content"`
	PublishedAt  string `json:"published_at"`
	Platform     string `json:"platform"`
	AuthorName   string `json:"author_name"`
	ContentHash  string `json:"content_hash"`
}

func (s *StreamMessage) MarshalBinary() ([]byte, error) {
	return json.Marshal(s)
}

func (s *StreamMessage) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *StreamMessage) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte(value.(string)), &s)
	}
	return json.Unmarshal(bytes, &s)
}
