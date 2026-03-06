package models

import "time"

// Thread represents a sub-conversation within a task
type Thread struct {
	ID        int64     `json:"id"`
	TaskID    int64     `json:"task_id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Message represents a chat message
type Message struct {
	ID          int64     `json:"id"`
	TaskID      int64     `json:"task_id"`
	ThreadID    *int64    `json:"thread_id,omitempty"`
	Role        string    `json:"role"`                // 'user', 'ai'
	Type        string    `json:"type"`                // 'text', 'execution'
	MessageType string    `json:"message_type"`        // 'user_query', 'ai_reply', 'system_notification'
	Content     string    `json:"content"`
	Metadata    *string   `json:"metadata,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
