package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/junkfilter/backend-go/models"
	"github.com/junkfilter/backend-go/repositories"
)

// MessageHandler handles message-related HTTP requests
type MessageHandler struct {
	messageRepo *repositories.MessageRepository
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(messageRepo *repositories.MessageRepository) *MessageHandler {
	return &MessageHandler{
		messageRepo: messageRepo,
	}
}

// CreateMessageRequest request body (用于 /api/messages 路由)
type CreateMessageRequest struct {
	TaskID  int64   `json:"task_id" binding:"required"`
	Role    string  `json:"role" binding:"required,oneof=user ai"`
	Type    string  `json:"type" binding:"required,oneof=text execution"`
	Content string  `json:"content" binding:"required"`
	Metadata *string `json:"metadata,omitempty"`
}

// CreateMessageForTaskRequest request body (用于 /api/tasks/{id}/messages 路由)
type CreateMessageForTaskRequest struct {
	Role    string  `json:"role" binding:"required,oneof=user ai"`
	Type    string  `json:"type" binding:"required,oneof=text execution"`
	Content string  `json:"content" binding:"required"`
	Metadata *string `json:"metadata,omitempty"`
}

// GetTaskMessages retrieves all messages for a task
func (mh *MessageHandler) GetTaskMessages(c *gin.Context) {
	taskIDStr := c.Param("id")
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	messages, err := mh.messageRepo.GetByTaskID(c.Request.Context(), taskID)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error fetching messages: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
		return
	}

	if messages == nil {
		messages = []models.Message{}
	}

	c.JSON(http.StatusOK, messages)
}

// CreateMessage saves a new message
func (mh *MessageHandler) CreateMessage(c *gin.Context) {
	var req CreateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	msg := &models.Message{
		TaskID:    req.TaskID,
		Role:      req.Role,
		Type:      req.Type,
		Content:   req.Content,
		Metadata:  req.Metadata,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	id, err := mh.messageRepo.Create(c.Request.Context(), msg)
	if err != nil {
		log.Printf("Error creating message: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create message"})
		return
	}

	msg.ID = id
	c.JSON(http.StatusCreated, msg)
}

// CreateMessageForTask creates a message for a specific task (task_id from URL param)
func (mh *MessageHandler) CreateMessageForTask(c *gin.Context) {
	taskIDStr := c.Param("id")
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var req CreateMessageForTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	msg := &models.Message{
		TaskID:    taskID,
		Role:      req.Role,
		Type:      req.Type,
		Content:   req.Content,
		Metadata:  req.Metadata,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	id, err := mh.messageRepo.Create(c.Request.Context(), msg)
	if err != nil {
		log.Printf("Error creating message: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create message"})
		return
	}

	msg.ID = id
	c.JSON(http.StatusCreated, msg)
}
func (mh *MessageHandler) GetMessages(c *gin.Context) {
	taskIDStr := c.Query("task_id")
	if taskIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task_id query parameter required"})
		return
	}

	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	messages, err := mh.messageRepo.GetByTaskID(c.Request.Context(), taskID)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error fetching messages: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
		return
	}

	if messages == nil {
		messages = []models.Message{}
	}

	c.JSON(http.StatusOK, messages)
}

// DeleteTaskMessages deletes all messages for a task
func (mh *MessageHandler) DeleteTaskMessages(c *gin.Context) {
	taskIDStr := c.Param("id")
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	count, err := mh.messageRepo.DeleteByTaskID(c.Request.Context(), taskID)
	if err != nil {
		log.Printf("Error deleting messages: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": count})
}
