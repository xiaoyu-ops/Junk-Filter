package handlers

import (
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
}

// NewContentHandler creates a new content handler
func NewContentHandler(
	contentRepo *repositories.ContentRepository,
	evaluationRepo *repositories.EvaluationRepository,
) *ContentHandler {
	return &ContentHandler{
		contentRepo:    contentRepo,
		evaluationRepo: evaluationRepo,
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

// ListContent lists content with filtering
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

	responses := make([]*models.ContentResponse, len(contents))
	for i, cont := range contents {
		responses[i] = cont.ToResponse()
	}

	c.JSON(http.StatusOK, responses)
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
