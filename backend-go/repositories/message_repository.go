package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/junkfilter/backend-go/models"
)

// MessageRepository handles message database operations
type MessageRepository struct {
	db *sql.DB
}

// NewMessageRepository creates a new message repository
func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

// Create inserts a new message and returns its ID
func (mr *MessageRepository) Create(ctx context.Context, msg *models.Message) (int64, error) {
	query := `
		INSERT INTO messages (task_id, role, type, content, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id
	`

	var id int64
	err := mr.db.QueryRowContext(ctx, query, msg.TaskID, msg.Role, msg.Type, msg.Content, msg.Metadata).Scan(&id)
	if err != nil {
		log.Printf("Error creating message: %v", err)
		return 0, fmt.Errorf("failed to create message: %w", err)
	}

	return id, nil
}

// GetByTaskID retrieves all messages for a specific task
func (mr *MessageRepository) GetByTaskID(ctx context.Context, taskID int64) ([]models.Message, error) {
	query := `
		SELECT id, task_id, role, type, content, metadata, created_at, updated_at
		FROM messages
		WHERE task_id = $1
		ORDER BY created_at ASC
	`

	rows, err := mr.db.QueryContext(ctx, query, taskID)
	if err != nil {
		log.Printf("Error querying messages: %v", err)
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	messages := []models.Message{}

	for rows.Next() {
		var msg models.Message
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&msg.ID,
			&msg.TaskID,
			&msg.Role,
			&msg.Type,
			&msg.Content,
			&msg.Metadata,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			log.Printf("Error scanning message: %v", err)
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		msg.CreatedAt = createdAt
		msg.UpdatedAt = updatedAt
		messages = append(messages, msg)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating messages: %v", err)
		return nil, fmt.Errorf("failed to iterate messages: %w", err)
	}

	return messages, nil
}

// GetByID retrieves a single message by ID
func (mr *MessageRepository) GetByID(ctx context.Context, id int64) (*models.Message, error) {
	query := `
		SELECT id, task_id, role, type, content, metadata, created_at, updated_at
		FROM messages
		WHERE id = $1
	`

	msg := &models.Message{}
	var createdAt, updatedAt time.Time

	err := mr.db.QueryRowContext(ctx, query, id).Scan(
		&msg.ID,
		&msg.TaskID,
		&msg.Role,
		&msg.Type,
		&msg.Content,
		&msg.Metadata,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Printf("Error querying message: %v", err)
		return nil, fmt.Errorf("failed to query message: %w", err)
	}

	msg.CreatedAt = createdAt
	msg.UpdatedAt = updatedAt

	return msg, nil
}

// DeleteByTaskID deletes all messages for a task and returns count
func (mr *MessageRepository) DeleteByTaskID(ctx context.Context, taskID int64) (int64, error) {
	query := `DELETE FROM messages WHERE task_id = $1`

	result, err := mr.db.ExecContext(ctx, query, taskID)
	if err != nil {
		log.Printf("Error deleting messages: %v", err)
		return 0, fmt.Errorf("failed to delete messages: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return count, nil
}

// DeleteByID deletes a message by ID
func (mr *MessageRepository) DeleteByID(ctx context.Context, id int64) error {
	query := `DELETE FROM messages WHERE id = $1`

	result, err := mr.db.ExecContext(ctx, query, id)
	if err != nil {
		log.Printf("Error deleting message: %v", err)
		return fmt.Errorf("failed to delete message: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("message not found")
	}

	return nil
}

// Update updates a message
func (mr *MessageRepository) Update(ctx context.Context, msg *models.Message) error {
	query := `
		UPDATE messages
		SET role = $1, type = $2, content = $3, metadata = $4, updated_at = NOW()
		WHERE id = $5
	`

	result, err := mr.db.ExecContext(
		ctx,
		query,
		msg.Role,
		msg.Type,
		msg.Content,
		msg.Metadata,
		msg.ID,
	)
	if err != nil {
		log.Printf("Error updating message: %v", err)
		return fmt.Errorf("failed to update message: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("message not found")
	}

	return nil
}
