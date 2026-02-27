package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/junkfilter/backend-go/models"
)

// ============================================================
// BloggerRepository - 博主数据操作
// ============================================================
type BloggerRepository struct {
	db *sql.DB
}

func NewBloggerRepository(db *sql.DB) *BloggerRepository {
	return &BloggerRepository{db: db}
}

func (r *BloggerRepository) GetAll(limit, offset int) ([]*models.Blogger, int, error) {
	// 获取总数
	var total int
	err := r.db.QueryRow("SELECT COUNT(*) FROM bloggers").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	rows, err := r.db.Query(`
		SELECT id, name, bio, location, avatar, rss_feed, status, filter_rate, total_articles, created_at, updated_at
		FROM bloggers
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var bloggers []*models.Blogger
	for rows.Next() {
		b := &models.Blogger{}
		err := rows.Scan(&b.ID, &b.Name, &b.Bio, &b.Location, &b.Avatar, &b.RSSOFeed,
			&b.Status, &b.FilterRate, &b.TotalArticles, &b.CreatedAt, &b.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		bloggers = append(bloggers, b)
	}

	return bloggers, total, nil
}

func (r *BloggerRepository) GetByID(id int) (*models.Blogger, error) {
	b := &models.Blogger{}
	err := r.db.QueryRow(`
		SELECT id, name, bio, location, avatar, rss_feed, status, filter_rate, total_articles, created_at, updated_at
		FROM bloggers
		WHERE id = $1
	`, id).Scan(&b.ID, &b.Name, &b.Bio, &b.Location, &b.Avatar, &b.RSSOFeed,
		&b.Status, &b.FilterRate, &b.TotalArticles, &b.CreatedAt, &b.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("blogger not found")
	}
	return b, err
}

func (r *BloggerRepository) Create(req *models.CreateBloggerRequest) (*models.Blogger, error) {
	b := &models.Blogger{
		Name:      req.Name,
		Bio:       req.Bio,
		Location:  req.Location,
		RSSOFeed:  req.RSSFeed,
		Status:    "active",
		FilterRate: 0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := r.db.QueryRow(`
		INSERT INTO bloggers (name, bio, location, avatar, rss_feed, status, filter_rate, created_at, updated_at)
		VALUES ($1, $2, $3, '', $4, $5, $6, $7, $8)
		RETURNING id
	`, b.Name, b.Bio, b.Location, b.RSSOFeed, b.Status, b.FilterRate, b.CreatedAt, b.UpdatedAt).Scan(&b.ID)

	return b, err
}

func (r *BloggerRepository) Update(id int, b *models.Blogger) error {
	_, err := r.db.Exec(`
		UPDATE bloggers
		SET name = $1, bio = $2, location = $3, status = $4, filter_rate = $5, updated_at = NOW()
		WHERE id = $6
	`, b.Name, b.Bio, b.Location, b.Status, b.FilterRate, id)
	return err
}

func (r *BloggerRepository) Delete(id int) error {
	_, err := r.db.Exec("DELETE FROM bloggers WHERE id = $1", id)
	return err
}

func (r *BloggerRepository) UpdateStatus(id int, status string) error {
	_, err := r.db.Exec("UPDATE bloggers SET status = $1, updated_at = NOW() WHERE id = $2", status, id)
	return err
}


// ============================================================
// TaskRepository - 任务数据操作
// ============================================================
type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) GetAll(limit, offset int) ([]*models.Task, int, error) {
	var total int
	err := r.db.QueryRow("SELECT COUNT(*) FROM tasks").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(`
		SELECT id, name, description, schedule, time_display, enabled, type, config, last_executed_at, next_execute_at, execution_count, created_at, updated_at
		FROM tasks
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		t := &models.Task{}
		var config []byte
		err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.Schedule, &t.TimeDisplay, &t.Enabled,
			&t.Type, &config, &t.LastExecutedAt, &t.NextExecuteAt, &t.ExecutionCount, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		// 解析 config JSON
		if config != nil {
			t.Config.Scan(config)
		}
		tasks = append(tasks, t)
	}

	return tasks, total, nil
}

func (r *TaskRepository) Create(t *models.Task) error {
	err := r.db.QueryRow(`
		INSERT INTO tasks (name, description, schedule, time_display, enabled, type, config, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`, t.Name, t.Description, t.Schedule, t.TimeDisplay, t.Enabled, t.Type, t.Config, t.CreatedAt, t.UpdatedAt).Scan(&t.ID)
	return err
}

func (r *TaskRepository) Update(id int, t *models.Task) error {
	_, err := r.db.Exec(`
		UPDATE tasks
		SET name = $1, description = $2, schedule = $3, time_display = $4, enabled = $5, type = $6, config = $7, updated_at = NOW()
		WHERE id = $8
	`, t.Name, t.Description, t.Schedule, t.TimeDisplay, t.Enabled, t.Type, t.Config, id)
	return err
}

func (r *TaskRepository) Delete(id int) error {
	_, err := r.db.Exec("DELETE FROM tasks WHERE id = $1", id)
	return err
}

func (r *TaskRepository) UpdateExecutionTime(id int, nextExecuteAt time.Time) error {
	_, err := r.db.Exec(`
		UPDATE tasks
		SET last_executed_at = NOW(), next_execute_at = $1, execution_count = execution_count + 1
		WHERE id = $2
	`, nextExecuteAt, id)
	return err
}
