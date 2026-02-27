package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/junkfilter/backend-go/models"
	"github.com/junkfilter/backend-go/repositories"
)

// EvaluationHandler handles evaluation-related HTTP requests
type EvaluationHandler struct {
	evaluationRepo *repositories.EvaluationRepository
}

// NewEvaluationHandler creates a new evaluation handler
func NewEvaluationHandler(evaluationRepo *repositories.EvaluationRepository) *EvaluationHandler {
	return &EvaluationHandler{
		evaluationRepo: evaluationRepo,
	}
}

// ListEvaluationsByDecision lists evaluations by decision
func (eh *EvaluationHandler) ListEvaluationsByDecision(c *gin.Context) {
	decision := c.Query("decision")
	if decision == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "decision query parameter is required"})
		return
	}

	limit := 50
	offset := 0

	if limStr := c.Query("limit"); limStr != "" {
		if lim, err := strconv.Atoi(limStr); err == nil && lim > 0 {
			limit = lim
		}
	}

	if offStr := c.Query("offset"); offStr != "" {
		if off, err := strconv.Atoi(offStr); err == nil && off >= 0 {
			offset = off
		}
	}

	evaluations, err := eh.evaluationRepo.ListByDecision(c.Request.Context(), decision, limit, offset)
	if err != nil {
		log.Printf("Error listing evaluations: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list evaluations"})
		return
	}

	responses := make([]*models.EvaluationResponse, len(evaluations))
	for i, ev := range evaluations {
		responses[i] = ev.ToResponse()
	}

	c.JSON(http.StatusOK, responses)
}

// ListHighScores lists evaluations with high scores
func (eh *EvaluationHandler) ListHighScores(c *gin.Context) {
	minInnovation := 5
	minDepth := 4
	limit := 50
	offset := 0

	if innStr := c.Query("min_innovation"); innStr != "" {
		if inn, err := strconv.Atoi(innStr); err == nil {
			minInnovation = inn
		}
	}

	if depStr := c.Query("min_depth"); depStr != "" {
		if dep, err := strconv.Atoi(depStr); err == nil {
			minDepth = dep
		}
	}

	if limStr := c.Query("limit"); limStr != "" {
		if lim, err := strconv.Atoi(limStr); err == nil && lim > 0 {
			limit = lim
		}
	}

	if offStr := c.Query("offset"); offStr != "" {
		if off, err := strconv.Atoi(offStr); err == nil && off >= 0 {
			offset = off
		}
	}

	evaluations, err := eh.evaluationRepo.ListHighScores(c.Request.Context(), minInnovation, minDepth, limit, offset)
	if err != nil {
		log.Printf("Error listing high scores: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list evaluations"})
		return
	}

	responses := make([]*models.EvaluationResponse, len(evaluations))
	for i, ev := range evaluations {
		responses[i] = ev.ToResponse()
	}

	c.JSON(http.StatusOK, responses)
}
