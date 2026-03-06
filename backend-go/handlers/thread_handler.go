package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/junkfilter/backend-go/models"
	"github.com/junkfilter/backend-go/repositories"
)

type ThreadHandler struct {
	threadRepo  *repositories.ThreadRepository
	messageRepo *repositories.MessageRepository
}

func NewThreadHandler(threadRepo *repositories.ThreadRepository, messageRepo *repositories.MessageRepository) *ThreadHandler {
	return &ThreadHandler{
		threadRepo:  threadRepo,
		messageRepo: messageRepo,
	}
}

// ListThreads returns all threads for a task
// GET /api/tasks/:id/threads
func (h *ThreadHandler) ListThreads(c *gin.Context) {
	taskID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	threads, err := h.threadRepo.GetByTaskID(c.Request.Context(), taskID)
	if err != nil {
		log.Printf("Error listing threads: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list threads"})
		return
	}

	if threads == nil {
		threads = []models.Thread{}
	}

	c.JSON(http.StatusOK, gin.H{"data": threads})
}

// CreateThread creates a new thread under a task
// POST /api/tasks/:id/threads
func (h *ThreadHandler) CreateThread(c *gin.Context) {
	taskID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var req struct {
		Title string `json:"title" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title is required"})
		return
	}

	thread := &models.Thread{
		TaskID: taskID,
		Title:  req.Title,
	}

	id, err := h.threadRepo.Create(c.Request.Context(), thread)
	if err != nil {
		log.Printf("Error creating thread: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create thread"})
		return
	}

	thread.ID = id
	created, _ := h.threadRepo.GetByID(c.Request.Context(), id)
	if created != nil {
		thread = created
	}

	c.JSON(http.StatusCreated, thread)
}

// DeleteThread deletes a thread and its messages
// DELETE /api/threads/:id
func (h *ThreadHandler) DeleteThread(c *gin.Context) {
	threadID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid thread ID"})
		return
	}

	if err := h.threadRepo.Delete(c.Request.Context(), threadID); err != nil {
		log.Printf("Error deleting thread: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete thread"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Thread deleted"})
}

// GetThreadMessages returns messages for a specific thread
// GET /api/threads/:id/messages
func (h *ThreadHandler) GetThreadMessages(c *gin.Context) {
	threadID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid thread ID"})
		return
	}

	messages, err := h.messageRepo.GetByThreadID(c.Request.Context(), threadID)
	if err != nil {
		log.Printf("Error getting thread messages: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get messages"})
		return
	}

	if messages == nil {
		messages = []models.Message{}
	}

	c.JSON(http.StatusOK, messages)
}

// UpdateThread updates a thread title
// PUT /api/threads/:id
func (h *ThreadHandler) UpdateThread(c *gin.Context) {
	threadID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid thread ID"})
		return
	}

	var req struct {
		Title string `json:"title" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title is required"})
		return
	}

	if err := h.threadRepo.UpdateTitle(c.Request.Context(), threadID, req.Title); err != nil {
		log.Printf("Error updating thread: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update thread"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Thread updated"})
}
