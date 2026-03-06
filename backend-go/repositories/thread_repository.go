package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/junkfilter/backend-go/models"
)

type ThreadRepository struct {
	db *sql.DB
}

func NewThreadRepository(db *sql.DB) *ThreadRepository {
	return &ThreadRepository{db: db}
}

// Create inserts a new thread and returns its ID
func (tr *ThreadRepository) Create(ctx context.Context, thread *models.Thread) (int64, error) {
	query := `
		INSERT INTO threads (task_id, title, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING id
	`
	var id int64
	err := tr.db.QueryRowContext(ctx, query, thread.TaskID, thread.Title).Scan(&id)
	if err != nil {
		log.Printf("Error creating thread: %v", err)
		return 0, fmt.Errorf("failed to create thread: %w", err)
	}
	return id, nil
}

// GetByTaskID retrieves all threads for a task, ordered by most recent activity
func (tr *ThreadRepository) GetByTaskID(ctx context.Context, taskID int64) ([]models.Thread, error) {
	query := `
		SELECT t.id, t.task_id, t.title, t.created_at, t.updated_at
		FROM threads t
		WHERE t.task_id = $1
		ORDER BY t.updated_at DESC
	`
	rows, err := tr.db.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to query threads: %w", err)
	}
	defer rows.Close()

	var threads []models.Thread
	for rows.Next() {
		var t models.Thread
		if err := rows.Scan(&t.ID, &t.TaskID, &t.Title, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan thread: %w", err)
		}
		threads = append(threads, t)
	}
	return threads, rows.Err()
}

// GetByID retrieves a single thread
func (tr *ThreadRepository) GetByID(ctx context.Context, id int64) (*models.Thread, error) {
	query := `SELECT id, task_id, title, created_at, updated_at FROM threads WHERE id = $1`
	t := &models.Thread{}
	err := tr.db.QueryRowContext(ctx, query, id).Scan(&t.ID, &t.TaskID, &t.Title, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get thread: %w", err)
	}
	return t, nil
}

// Delete removes a thread (messages cascade-deleted)
func (tr *ThreadRepository) Delete(ctx context.Context, id int64) error {
	_, err := tr.db.ExecContext(ctx, "DELETE FROM threads WHERE id = $1", id)
	return err
}

// UpdateTitle updates thread title and timestamp
func (tr *ThreadRepository) UpdateTitle(ctx context.Context, id int64, title string) error {
	_, err := tr.db.ExecContext(ctx, "UPDATE threads SET title = $1, updated_at = NOW() WHERE id = $2", title, id)
	return err
}

// TouchUpdatedAt bumps the updated_at timestamp (called when a new message is added)
func (tr *ThreadRepository) TouchUpdatedAt(ctx context.Context, id int64) error {
	_, err := tr.db.ExecContext(ctx, "UPDATE threads SET updated_at = NOW() WHERE id = $1", id)
	return err
}
