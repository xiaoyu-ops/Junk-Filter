package models

import "time"

// Message represents a chat message
type Message struct {
	ID        int64     `json:"id"`
	TaskID    int64     `json:"task_id"`
	Role      string    `json:"role"`      // 'user', 'ai'
	Type      string    `json:"type"`      // 'text', 'execution'
	Content   string    `json:"content"`
	Metadata  *string   `json:"metadata,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
