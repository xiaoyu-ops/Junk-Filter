package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/junkfilter/backend-go/models"
)

type SourceRepository struct {
	db *sql.DB
}

func NewSourceRepository(db *sql.DB) *SourceRepository {
	return &SourceRepository{db: db}
}

// Create inserts a new source
func (sr *SourceRepository) Create(ctx context.Context, req *models.CreateSourceRequest) (*models.Source, error) {
	source := &models.Source{
		Platform:             req.Platform,
		URL:                  req.URL,
		AuthorName:           req.AuthorName,
		Priority:             req.Priority,
		FetchIntervalSeconds: req.FetchIntervalSeconds,
		Enabled:              true,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	if source.Priority == 0 {
		source.Priority = 5
	}
	if source.FetchIntervalSeconds == 0 {
		source.FetchIntervalSeconds = 3600
	}

	err := sr.db.QueryRowContext(ctx,
		`INSERT INTO sources (platform, url, author_name, priority, fetch_interval_seconds, enabled, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id, created_at, updated_at`,
		source.Platform, source.URL, source.AuthorName, source.Priority,
		source.FetchIntervalSeconds, source.Enabled, source.CreatedAt, source.UpdatedAt,
	).Scan(&source.ID, &source.CreatedAt, &source.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return source, nil
}

// GetByID retrieves a source by ID
func (sr *SourceRepository) GetByID(ctx context.Context, id int64) (*models.Source, error) {
	source := &models.Source{}
	var lastFetchTime sql.NullTime
	var authorID sql.NullString

	err := sr.db.QueryRowContext(ctx,
		`SELECT id, platform, url, author_name, author_id, priority, last_fetch_time,
		        fetch_interval_seconds, enabled, created_at, updated_at
		 FROM sources WHERE id = $1`,
		id,
	).Scan(&source.ID, &source.Platform, &source.URL, &source.AuthorName, &authorID,
		&source.Priority, &lastFetchTime, &source.FetchIntervalSeconds, &source.Enabled,
		&source.CreatedAt, &source.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if lastFetchTime.Valid {
		source.LastFetchTime = &lastFetchTime.Time
	}

	if authorID.Valid {
		source.AuthorID = &authorID.String
	}

	return source, nil
}

// GetAll retrieves all enabled sources
func (sr *SourceRepository) GetAll(ctx context.Context, enabledOnly bool) ([]*models.Source, error) {
	query := `SELECT id, platform, url, author_name, author_id, priority, last_fetch_time,
	                fetch_interval_seconds, enabled, created_at, updated_at
	          FROM sources`

	if enabledOnly {
		query += " WHERE enabled = TRUE"
	}

	query += " ORDER BY priority DESC, created_at DESC"

	rows, err := sr.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sources []*models.Source
	for rows.Next() {
		source := &models.Source{}
		var lastFetchTime sql.NullTime
		var authorID sql.NullString

		err := rows.Scan(&source.ID, &source.Platform, &source.URL, &source.AuthorName, &authorID,
			&source.Priority, &lastFetchTime, &source.FetchIntervalSeconds, &source.Enabled,
			&source.CreatedAt, &source.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if lastFetchTime.Valid {
			source.LastFetchTime = &lastFetchTime.Time
		}

		if authorID.Valid {
			source.AuthorID = &authorID.String
		}

		sources = append(sources, source)
	}

	return sources, rows.Err()
}

// Update updates a source
func (sr *SourceRepository) Update(ctx context.Context, id int64, req *models.UpdateSourceRequest) (*models.Source, error) {
	source, err := sr.GetByID(ctx, id)
	if err != nil || source == nil {
		return nil, err
	}

	if req.AuthorName != "" {
		source.AuthorName = req.AuthorName
	}
	if req.Priority > 0 {
		source.Priority = req.Priority
	}
	if req.FetchIntervalSeconds > 0 {
		source.FetchIntervalSeconds = req.FetchIntervalSeconds
	}
	source.Enabled = req.Enabled
	source.UpdatedAt = time.Now()

	_, err = sr.db.ExecContext(ctx,
		`UPDATE sources SET author_name = $1, priority = $2, fetch_interval_seconds = $3, enabled = $4, updated_at = $5
		 WHERE id = $6`,
		source.AuthorName, source.Priority, source.FetchIntervalSeconds, source.Enabled, source.UpdatedAt, id,
	)

	if err != nil {
		return nil, err
	}

	return source, nil
}

// Delete deletes a source
func (sr *SourceRepository) Delete(ctx context.Context, id int64) error {
	result, err := sr.db.ExecContext(ctx, "DELETE FROM sources WHERE id = $1", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// UpdateLastFetchTime updates the last_fetch_time for a source
func (sr *SourceRepository) UpdateLastFetchTime(ctx context.Context, id int64, lastFetchTime time.Time) error {
	_, err := sr.db.ExecContext(ctx,
		"UPDATE sources SET last_fetch_time = $1, updated_at = $2 WHERE id = $3",
		lastFetchTime, time.Now(), id,
	)
	return err
}
