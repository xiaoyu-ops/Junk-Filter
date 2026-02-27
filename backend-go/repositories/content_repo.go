package repositories

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/junkfilter/backend-go/models"
)

type ContentRepository struct {
	db *sql.DB
}

func NewContentRepository(db *sql.DB) *ContentRepository {
	return &ContentRepository{db: db}
}

// Create inserts a new content
func (cr *ContentRepository) Create(ctx context.Context, req *models.CreateContentRequest) (*models.Content, error) {
	content := &models.Content{
		TaskID:       uuid.New(),
		SourceID:     req.SourceID,
		Platform:     req.Platform,
		AuthorName:   req.AuthorName,
		Title:        req.Title,
		OriginalURL:  req.OriginalURL,
		ContentHash:  req.ContentHash,
		CleanContent: req.CleanContent,
		PublishedAt:  req.PublishedAt,
		IngestedAt:   time.Now(),
		Status:       "PENDING",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := cr.db.QueryRowContext(ctx,
		`INSERT INTO content (task_id, source_id, platform, author_name, title, original_url,
		                      content_hash, clean_content, published_at, ingested_at, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		 RETURNING id, task_id, created_at, updated_at`,
		content.TaskID, content.SourceID, content.Platform, content.AuthorName, content.Title,
		content.OriginalURL, content.ContentHash, content.CleanContent, content.PublishedAt,
		content.IngestedAt, content.Status, content.CreatedAt, content.UpdatedAt,
	).Scan(&content.ID, &content.TaskID, &content.CreatedAt, &content.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return content, nil
}

// GetByID retrieves content by ID
func (cr *ContentRepository) GetByID(ctx context.Context, id int64) (*models.Content, error) {
	content := &models.Content{}
	var publishedAt sql.NullTime

	err := cr.db.QueryRowContext(ctx,
		`SELECT id, task_id, source_id, platform, author_name, title, original_url,
		        content_hash, clean_content, published_at, ingested_at, status, created_at, updated_at
		 FROM content WHERE id = $1`,
		id,
	).Scan(&content.ID, &content.TaskID, &content.SourceID, &content.Platform, &content.AuthorName,
		&content.Title, &content.OriginalURL, &content.ContentHash, &content.CleanContent,
		&publishedAt, &content.IngestedAt, &content.Status, &content.CreatedAt, &content.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if publishedAt.Valid {
		content.PublishedAt = &publishedAt.Time
	}

	return content, nil
}

// GetByTaskID retrieves content by task ID
func (cr *ContentRepository) GetByTaskID(ctx context.Context, taskID uuid.UUID) (*models.Content, error) {
	content := &models.Content{}
	var publishedAt sql.NullTime

	err := cr.db.QueryRowContext(ctx,
		`SELECT id, task_id, source_id, platform, author_name, title, original_url,
		        content_hash, clean_content, published_at, ingested_at, status, created_at, updated_at
		 FROM content WHERE task_id = $1`,
		taskID,
	).Scan(&content.ID, &content.TaskID, &content.SourceID, &content.Platform, &content.AuthorName,
		&content.Title, &content.OriginalURL, &content.ContentHash, &content.CleanContent,
		&publishedAt, &content.IngestedAt, &content.Status, &content.CreatedAt, &content.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if publishedAt.Valid {
		content.PublishedAt = &publishedAt.Time
	}

	return content, nil
}

// GetByURL retrieves content by original URL
func (cr *ContentRepository) GetByURL(ctx context.Context, url string) (*models.Content, error) {
	content := &models.Content{}
	var publishedAt sql.NullTime

	err := cr.db.QueryRowContext(ctx,
		`SELECT id, task_id, source_id, platform, author_name, title, original_url,
		        content_hash, clean_content, published_at, ingested_at, status, created_at, updated_at
		 FROM content WHERE original_url = $1`,
		url,
	).Scan(&content.ID, &content.TaskID, &content.SourceID, &content.Platform, &content.AuthorName,
		&content.Title, &content.OriginalURL, &content.ContentHash, &content.CleanContent,
		&publishedAt, &content.IngestedAt, &content.Status, &content.CreatedAt, &content.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if publishedAt.Valid {
		content.PublishedAt = &publishedAt.Time
	}

	return content, nil
}

// GetByHash retrieves content by content hash
func (cr *ContentRepository) GetByHash(ctx context.Context, hash string) (*models.Content, error) {
	content := &models.Content{}
	var publishedAt sql.NullTime

	err := cr.db.QueryRowContext(ctx,
		`SELECT id, task_id, source_id, platform, author_name, title, original_url,
		        content_hash, clean_content, published_at, ingested_at, status, created_at, updated_at
		 FROM content WHERE content_hash = $1`,
		hash,
	).Scan(&content.ID, &content.TaskID, &content.SourceID, &content.Platform, &content.AuthorName,
		&content.Title, &content.OriginalURL, &content.ContentHash, &content.CleanContent,
		&publishedAt, &content.IngestedAt, &content.Status, &content.CreatedAt, &content.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if publishedAt.Valid {
		content.PublishedAt = &publishedAt.Time
	}

	return content, nil
}

// List retrieves multiple contents with filtering
func (cr *ContentRepository) List(ctx context.Context, filter *models.ContentFilter) ([]*models.Content, error) {
	query := `SELECT id, task_id, source_id, platform, author_name, title, original_url,
	                 content_hash, clean_content, published_at, ingested_at, status, created_at, updated_at
	          FROM content WHERE 1=1`

	args := []interface{}{}
	argIndex := 1

	if filter.Status != "" {
		query += " AND status = $" + strconv.Itoa(argIndex)
		args = append(args, filter.Status)
		argIndex++
	}

	if filter.SourceID > 0 {
		query += " AND source_id = $" + strconv.Itoa(argIndex)
		args = append(args, filter.SourceID)
		argIndex++
	}

	query += " ORDER BY created_at DESC LIMIT $" + strconv.Itoa(argIndex) + " OFFSET $" + strconv.Itoa(argIndex+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := cr.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contents []*models.Content
	for rows.Next() {
		content := &models.Content{}
		var publishedAt sql.NullTime

		err := rows.Scan(&content.ID, &content.TaskID, &content.SourceID, &content.Platform, &content.AuthorName,
			&content.Title, &content.OriginalURL, &content.ContentHash, &content.CleanContent,
			&publishedAt, &content.IngestedAt, &content.Status, &content.CreatedAt, &content.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if publishedAt.Valid {
			content.PublishedAt = &publishedAt.Time
		}

		contents = append(contents, content)
	}

	return contents, rows.Err()
}

// UpdateStatus updates the status of a content
func (cr *ContentRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	_, err := cr.db.ExecContext(ctx,
		"UPDATE content SET status = $1, updated_at = $2 WHERE id = $3",
		status, time.Now(), id,
	)
	return err
}

// UpdateStatusByTaskID updates the status using task ID
func (cr *ContentRepository) UpdateStatusByTaskID(ctx context.Context, taskID uuid.UUID, status string) error {
	_, err := cr.db.ExecContext(ctx,
		"UPDATE content SET status = $1, updated_at = $2 WHERE task_id = $3",
		status, time.Now(), taskID,
	)
	return err
}
