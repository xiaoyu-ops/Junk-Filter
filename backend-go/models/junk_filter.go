package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// ============================================================
// Blogger - 博主模型
// ============================================================
type Blogger struct {
	ID            int       `json:"id"`
	Name          string    `json:"name"`
	Bio           string    `json:"bio"`
	Location      string    `json:"location"`
	Avatar        string    `json:"avatar"`
	RSSOFeed      string    `json:"rss_feed"`
	Status        string    `json:"status"` // active, paused, blocked
	FilterRate    float64   `json:"filter_rate"`
	TotalArticles int       `json:"total_articles"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ============================================================
// Feed - RSS 源模型
// ============================================================
type Feed struct {
	ID                int       `json:"id"`
	URL               string    `json:"url"`
	Name              string    `json:"name"`
	UpdateFrequency   string    `json:"update_frequency"` // hourly, 30min, 2hours
	Status            string    `json:"status"`            // active, paused, failed
	LastFetchTime     *time.Time `json:"last_fetch_time"`
	ErrorCount        int       `json:"error_count"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// ============================================================
// JFContent - JunkFilter 内容模型（避免与 Content 冲突）
// ============================================================
type JFContent struct {
	ID              int       `json:"id"`
	BloggerID       int       `json:"blogger_id"`
	Title           string    `json:"title"`
	Summary         string    `json:"summary"`
	OriginalURL     string    `json:"original_url"`
	PublishedAt     time.Time `json:"published_at"`

	// AI 评分
	QualityScore    *float64  `json:"quality_score"`
	RelevanceScore  *float64  `json:"relevance_score"`
	AIDecision      string    `json:"ai_decision"` // pending, approved, rejected, review

	// JunkFilter 兼容字段
	InnovationScore *int      `json:"innovation_score"`
	DepthScore      *int      `json:"depth_score"`
	Decision        string    `json:"decision"`
	TLDR            string    `json:"tldr"`

	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	// 关联数据
	Blogger         *Blogger  `json:"blogger,omitempty"`
}

// ============================================================
// Task - 任务模型
// ============================================================
type Task struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`

	// 时间调度
	Schedule        string    `json:"schedule"` // cron: "0 9 * * *"
	TimeDisplay     string    `json:"time_display"` // "09:00 AM"
	Enabled         bool      `json:"enabled"`

	// 任务配置
	Type            string    `json:"type"` // summary, filter, monitor
	Config          JSONMap   `json:"config"`

	// 执行统计
	LastExecutedAt  *time.Time `json:"last_executed_at"`
	NextExecuteAt   *time.Time `json:"next_execute_at"`
	ExecutionCount  int       `json:"execution_count"`

	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// JSONMap 是自定义 JSON 类型
type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	bytes := value.([]byte)
	return json.Unmarshal(bytes, &j)
}

// ============================================================
// AIConfig - AI 配置模型
// ============================================================
type AIConfig struct {
	ID            int       `json:"id"`
	DefaultModel  string    `json:"default_model"`
	APIKey        string    `json:"api_key"`
	Temperature   float32   `json:"temperature"`
	MaxTokens     int       `json:"max_tokens"`
	BatchSize     int       `json:"batch_size"`
	RetryCount    int       `json:"retry_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ============================================================
// ContentStats - 内容统计模型
// ============================================================
type ContentStats struct {
	ID                 int       `json:"id"`
	BloggerID          int       `json:"blogger_id"`
	StatDate           time.Time `json:"stat_date"`

	ApprovedCount      int       `json:"approved_count"`
	RejectedCount      int       `json:"rejected_count"`
	ReviewCount        int       `json:"review_count"`
	TotalCount         int       `json:"total_count"`

	AvgQualityScore    *float64  `json:"avg_quality_score"`
	AvgRelevanceScore  *float64  `json:"avg_relevance_score"`

	CreatedAt          time.Time `json:"created_at"`
}

// ============================================================
// TaskLog - 任务执行日志
// ============================================================
type TaskLog struct {
	ID             int       `json:"id"`
	TaskID         int       `json:"task_id"`

	Status         string    `json:"status"` // success, failed, running
	ExecutedAt     time.Time `json:"executed_at"`
	CompletedAt    *time.Time `json:"completed_at"`

	ItemsProcessed int       `json:"items_processed"`
	ItemsCreated   int       `json:"items_created"`
	ErrorMessage   string    `json:"error_message"`
}

// ============================================================
// BloggerStats - 博主统计视图
// ============================================================
type BloggerStats struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	RSSOFeed        string    `json:"rss_feed"`
	TotalArticles   int       `json:"total_articles"`
	ApprovedCount   int       `json:"approved_count"`
	RejectedCount   int       `json:"rejected_count"`
	AvgQuality      *float64  `json:"avg_quality"`
	AvgRelevance    *float64  `json:"avg_relevance"`
	LatestArticle   *time.Time `json:"latest_article"`
}

// ============================================================
// 请求/响应 DTO
// ============================================================

// CreateBloggerRequest - 创建博主请求
type CreateBloggerRequest struct {
	Name     string `json:"name" binding:"required"`
	Bio      string `json:"bio"`
	Location string `json:"location"`
	RSSFeed  string `json:"rss_feed" binding:"required"`
}

// CreateTaskRequest - 创建任务请求
type CreateTaskRequest struct {
	Name        string      `json:"name" binding:"required"`
	Description string      `json:"description"`
	Schedule    string      `json:"schedule" binding:"required"`
	TimeDisplay string      `json:"time_display"`
	Type        string      `json:"type" binding:"required"`
	Config      JSONMap     `json:"config"`
}

// UpdateTaskRequest - 更新任务请求
type UpdateTaskRequest struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Schedule    string      `json:"schedule"`
	TimeDisplay string      `json:"time_display"`
	Type        string      `json:"type"`
	Config      JSONMap     `json:"config"`
	Enabled     *bool       `json:"enabled"`
}

// PaginationQuery - 分页查询
type PaginationQuery struct {
	Page     int `json:"page" form:"page" binding:"min=1" default:"1"`
	PageSize int `json:"page_size" form:"page_size" binding:"min=1,max=100" default:"20"`
}

// PaginatedResponse - 分页响应
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	Total      int         `json:"total"`
	TotalPages int         `json:"total_pages"`
}

// APIResponse - 通用 API 响应
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}
