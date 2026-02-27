package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/junkfilter/backend-go/models"
	"github.com/junkfilter/backend-go/repositories"
	"github.com/junkfilter/backend-go/services"
)

// SourceHandler handles source-related HTTP requests
type SourceHandler struct {
	sourceRepo *repositories.SourceRepository
	rssService *services.RSSService
}

// NewSourceHandler creates a new source handler
func NewSourceHandler(sourceRepo *repositories.SourceRepository, rssService *services.RSSService) *SourceHandler {
	return &SourceHandler{
		sourceRepo: sourceRepo,
		rssService: rssService,
	}
}

// CreateSource creates a new RSS source
func (sh *SourceHandler) CreateSource(c *gin.Context) {
	var req models.CreateSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	source, err := sh.sourceRepo.Create(c.Request.Context(), &req)
	if err != nil {
		log.Printf("Error creating source: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create source"})
		return
	}

	c.JSON(http.StatusCreated, source.ToResponse())
}

// GetSource retrieves a source by ID
func (sh *SourceHandler) GetSource(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid source ID"})
		return
	}

	source, err := sh.sourceRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		log.Printf("Error getting source: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get source"})
		return
	}

	if source == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Source not found"})
		return
	}

	c.JSON(http.StatusOK, source.ToResponse())
}

// ListSources lists all sources
func (sh *SourceHandler) ListSources(c *gin.Context) {
	enabledOnly := c.Query("enabled") == "true"

	sources, err := sh.sourceRepo.GetAll(c.Request.Context(), enabledOnly)
	if err != nil {
		log.Printf("Error listing sources: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list sources"})
		return
	}

	responses := make([]*models.SourceResponse, len(sources))
	for i, s := range sources {
		responses[i] = s.ToResponse()
	}

	c.JSON(http.StatusOK, responses)
}

// UpdateSource updates a source
func (sh *SourceHandler) UpdateSource(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid source ID"})
		return
	}

	var req models.UpdateSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	source, err := sh.sourceRepo.Update(c.Request.Context(), id, &req)
	if err != nil {
		log.Printf("Error updating source: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update source"})
		return
	}

	c.JSON(http.StatusOK, source.ToResponse())
}

// DeleteSource deletes a source
func (sh *SourceHandler) DeleteSource(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid source ID"})
		return
	}

	if err := sh.sourceRepo.Delete(c.Request.Context(), id); err != nil {
		log.Printf("Error deleting source: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete source"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Source deleted"})
}

// FetchSourceNow triggers an immediate fetch for a source
func (sh *SourceHandler) FetchSourceNow(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid source ID"})
		return
	}

	if err := sh.rssService.FetchSourceOnDemand(c.Request.Context(), id); err != nil {
		log.Printf("Error fetching source: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch source"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Source fetch triggered"})
}

// RegisterSourceRoutes registers source routes
func RegisterSourceRoutes(router *gin.Engine, handler *SourceHandler) {
	sources := router.Group("/api/sources")
	{
		sources.POST("", handler.CreateSource)
		sources.GET("", handler.ListSources)
		sources.GET("/:id", handler.GetSource)
		sources.PUT("/:id", handler.UpdateSource)
		sources.DELETE("/:id", handler.DeleteSource)
		sources.POST("/:id/fetch", handler.FetchSourceNow)
	}
}
