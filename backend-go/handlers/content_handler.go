package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/junkfilter/backend-go/models"
	"github.com/junkfilter/backend-go/repositories"
)

// ContentHandler handles content-related HTTP requests
type ContentHandler struct {
	contentRepo     *repositories.ContentRepository
	evaluationRepo  *repositories.EvaluationRepository
	sourceRepo      *repositories.SourceRepository
	db              *sql.DB
}

// NewContentHandler creates a new content handler
func NewContentHandler(
	contentRepo *repositories.ContentRepository,
	evaluationRepo *repositories.EvaluationRepository,
	sourceRepo *repositories.SourceRepository,
	db *sql.DB,
) *ContentHandler {
	return &ContentHandler{
		contentRepo:    contentRepo,
		evaluationRepo: evaluationRepo,
		sourceRepo:     sourceRepo,
		db:             db,
	}
}

// GetContent retrieves a content by ID
func (ch *ContentHandler) GetContent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid content ID"})
		return
	}

	content, err := ch.contentRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		log.Printf("Error getting content: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get content"})
		return
	}

	if content == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Content not found"})
		return
	}

	c.JSON(http.StatusOK, content.ToResponse())
}

// GetContentStats 获取内容统计信息（RSS 抓取进度）
func (ch *ContentHandler) GetContentStats(c *gin.Context) {
	// 查询各状态的数量
	type StatsResult struct {
		Status string `db:"status"`
		Count  int    `db:"count"`
	}

	query := `
		SELECT status, COUNT(*) as count
		FROM content
		GROUP BY status
	`

	rows, err := ch.db.QueryContext(c.Request.Context(), query)
	if err != nil {
		log.Printf("Error querying content stats: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get stats"})
		return
	}
	defer rows.Close()

	stats := map[string]int{
		"PENDING":    0,
		"PROCESSING": 0,
		"EVALUATED":  0,
		"DISCARDED":  0,
	}

	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			continue
		}
		stats[status] = count
	}

	c.JSON(http.StatusOK, gin.H{
		"pending":    stats["PENDING"],
		"processing": stats["PROCESSING"],
		"evaluated":  stats["EVALUATED"],
		"discarded":  stats["DISCARDED"],
		"total":      stats["PENDING"] + stats["PROCESSING"] + stats["EVALUATED"] + stats["DISCARDED"],
	})
}

// ListContent lists content with optional filtering
func (ch *ContentHandler) ListContent(c *gin.Context) {
	filter := &models.ContentFilter{
		Status:   c.Query("status"),
		Limit:    50,
		Offset:   0,
	}

	if limStr := c.Query("limit"); limStr != "" {
		if lim, err := strconv.Atoi(limStr); err == nil && lim > 0 {
			filter.Limit = lim
		}
	}

	if offStr := c.Query("offset"); offStr != "" {
		if off, err := strconv.Atoi(offStr); err == nil && off >= 0 {
			filter.Offset = off
		}
	}

	if sourceStr := c.Query("source_id"); sourceStr != "" {
		if sourceID, err := strconv.ParseInt(sourceStr, 10, 64); err == nil {
			filter.SourceID = sourceID
		}
	}

	contents, err := ch.contentRepo.List(c.Request.Context(), filter)
	if err != nil {
		log.Printf("Error listing content: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list content"})
		return
	}

	// Build source name cache to avoid N+1 queries
	sourceNames := make(map[int64]string)
	for _, cont := range contents {
		if _, exists := sourceNames[cont.SourceID]; !exists && cont.SourceID > 0 {
			source, err := ch.sourceRepo.GetByID(c.Request.Context(), cont.SourceID)
			if err == nil && source != nil {
				sourceNames[cont.SourceID] = source.AuthorName
			}
		}
	}

	// Build responses with evaluation data if available
	type ContentWithEvaluation struct {
		*models.ContentResponse
		Evaluation *models.EvaluationResponse `json:"evaluation,omitempty"`
		SourceName string                     `json:"source_name,omitempty"`
	}

	responses := make([]*ContentWithEvaluation, len(contents))
	for i, cont := range contents {
		// Get evaluation if exists
		evaluation, err := ch.evaluationRepo.GetByContentID(c.Request.Context(), cont.ID)
		response := &ContentWithEvaluation{
			ContentResponse: cont.ToResponse(),
			SourceName:      sourceNames[cont.SourceID],
		}

		if err == nil && evaluation != nil {
			response.Evaluation = evaluation.ToResponse()
		}

		responses[i] = response
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  responses,
		"count": len(responses),
	})
}

// StopEvaluation discards all PENDING and PROCESSING content to stop evaluation
func (ch *ContentHandler) StopEvaluation(c *gin.Context) {
	query := `
		UPDATE content SET status = 'DISCARDED', updated_at = NOW()
		WHERE status IN ('PENDING', 'PROCESSING')
	`
	result, err := ch.db.ExecContext(c.Request.Context(), query)
	if err != nil {
		log.Printf("Error stopping evaluation: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stop evaluation"})
		return
	}

	affected, _ := result.RowsAffected()
	log.Printf("[StopEvaluation] Discarded %d pending/processing items", affected)

	c.JSON(http.StatusOK, gin.H{
		"message":  "Evaluation stopped",
		"affected": affected,
	})
}

// RestartEvaluation resets DISCARDED content back to PENDING for re-evaluation
func (ch *ContentHandler) RestartEvaluation(c *gin.Context) {
	query := `
		UPDATE content SET status = 'PENDING', updated_at = NOW()
		WHERE status = 'DISCARDED'
	`
	result, err := ch.db.ExecContext(c.Request.Context(), query)
	if err != nil {
		log.Printf("Error restarting evaluation: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to restart evaluation"})
		return
	}

	affected, _ := result.RowsAffected()
	log.Printf("[RestartEvaluation] Reset %d discarded items to PENDING", affected)

	c.JSON(http.StatusOK, gin.H{
		"message":  "Evaluation restarted",
		"affected": affected,
	})
}

// GetContentWithEvaluation retrieves content along with its evaluation
func (ch *ContentHandler) GetContentWithEvaluation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid content ID"})
		return
	}

	content, err := ch.contentRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		log.Printf("Error getting content: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get content"})
		return
	}

	if content == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Content not found"})
		return
	}

	// Get evaluation if exists
	var evaluation *models.Evaluation
	evaluation, err = ch.evaluationRepo.GetByContentID(c.Request.Context(), id)
	if err != nil {
		log.Printf("Warning: Error getting evaluation: %v", err)
	}

	type ContentWithEvaluation struct {
		Content    *models.ContentResponse     `json:"content"`
		Evaluation *models.EvaluationResponse `json:"evaluation,omitempty"`
	}

	response := &ContentWithEvaluation{
		Content: content.ToResponse(),
	}

	if evaluation != nil {
		response.Evaluation = evaluation.ToResponse()
	}

	c.JSON(http.StatusOK, response)
}
