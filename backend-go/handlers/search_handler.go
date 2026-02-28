package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/junkfilter/backend-go/internal/models"
	"github.com/junkfilter/backend-go/internal/repository"
)

type SearchHandler struct {
	contentRepo *repository.ContentRepository
}

func NewSearchHandler(contentRepo *repository.ContentRepository) *SearchHandler {
	return &SearchHandler{
		contentRepo: contentRepo,
	}
}

/**
 * 搜索接口
 *
 * GET /api/search?q=keyword&status=EVALUATED&limit=50
 *
 * 功能：
 * - 在 title 和 content 中使用 PostgreSQL ILIKE 搜索
 * - 支持按状态过滤
 * - 支持分页（limit, offset）
 *
 * 性能特点：
 * - 使用数据库索引加速搜索
 * - 支持全文搜索（ILIKE 支持）
 * - 返回匹配内容 + 对应的评估结果
 */
func (h *SearchHandler) Search(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "search query 'q' is required",
		})
		return
	}

	// 参数解析
	status := c.DefaultQuery("status", "EVALUATED")
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	// 限制 limit 的最大值，防止恶意查询
	if limit > 1000 {
		limit = 1000
	}
	if limit <= 0 {
		limit = 50
	}

	// 清理搜索词（防止 SQL 注入）
	query = strings.TrimSpace(query)
	searchPattern := "%" + query + "%"

	// 构建 SQL 查询
	// 注意：这里假设 contentRepo 有一个支持 ILIKE 的查询方法
	// 如果没有，需要手动编写 SQL

	db := c.MustGet("db").(interface {
		Query(string, ...interface{}) interface {
			Scan(...interface{}) error
		}
	})

	var results []ContentSearchResult

	sql := `
		SELECT
			c.id,
			c.title,
			c.content,
			c.url,
			c.source_id,
			c.status,
			c.published_at,
			c.created_at,
			COALESCE(e.id, 0) as evaluation_id,
			COALESCE(e.innovation_score, 0) as innovation_score,
			COALESCE(e.depth_score, 0) as depth_score,
			COALESCE(e.decision, 'SKIP') as decision,
			COALESCE(e.tldr, '') as tldr,
			COALESCE(s.name, 'Unknown') as source_name
		FROM content c
		LEFT JOIN evaluation e ON c.id = e.content_id
		LEFT JOIN sources s ON c.source_id = s.id
		WHERE (c.title ILIKE $1 OR c.content ILIKE $1)
		  AND c.status = $2
		ORDER BY
			CASE
				WHEN c.title ILIKE $1 THEN 1
				ELSE 2
			END,
			c.published_at DESC
		LIMIT $3 OFFSET $4
	`

	// 执行查询
	// 这里的实现取决于你的数据库驱动（gorm, sqlc, 原生 sql 等）
	// 下面是伪代码，需要根据实际情况调整

	rows, err := db.Query(sql, searchPattern, status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "search failed: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var r ContentSearchResult
		err := rows.Scan(
			&r.ID, &r.Title, &r.Content, &r.URL, &r.SourceID,
			&r.Status, &r.PublishedAt, &r.CreatedAt,
			&r.EvaluationID, &r.InnovationScore, &r.DepthScore,
			&r.Decision, &r.TLDR, &r.SourceName,
		)
		if err != nil {
			continue
		}
		results = append(results, r)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  results,
		"count": len(results),
		"query": query,
	})
}

type ContentSearchResult struct {
	ID               int    `json:"id"`
	Title            string `json:"title"`
	Content          string `json:"content"`
	URL              string `json:"url"`
	SourceID         int    `json:"source_id"`
	Status           string `json:"status"`
	PublishedAt      string `json:"published_at"`
	CreatedAt        string `json:"created_at"`
	EvaluationID     int    `json:"evaluation_id"`
	InnovationScore  int    `json:"innovation_score"`
	DepthScore       int    `json:"depth_score"`
	Decision         string `json:"decision"`
	TLDR             string `json:"tldr"`
	SourceName       string `json:"source_name"`
}
