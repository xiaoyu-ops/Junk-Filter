package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/junkfilter/backend-go/models"
	"github.com/junkfilter/backend-go/repositories"
)

// AITaskHandler handles AI-powered task creation requests
type AITaskHandler struct {
	sourceRepo *repositories.SourceRepository
	taskRepo   *repositories.TaskRepository // 假设有 TaskRepository
	pythonAPIBaseURL string
}

// NewAITaskHandler creates a new AI task handler
func NewAITaskHandler(
	sourceRepo *repositories.SourceRepository,
	pythonAPIBaseURL string,
) *AITaskHandler {
	return &AITaskHandler{
		sourceRepo:       sourceRepo,
		pythonAPIBaseURL: pythonAPIBaseURL,
	}
}

// AICreateTaskRequest represents the request for AI-powered task creation
type AICreateTaskRequest struct {
	Message               string        `json:"message" binding:"required"`
	ConversationHistory   []ChatMessage `json:"conversation_history"`
}

// ChatMessage represents a message in the conversation
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AICreateTaskResponse represents the response from AI task creation
type AICreateTaskResponse struct {
	Reply       string           `json:"reply"`
	PendingTask *PendingTaskInfo `json:"pending_task,omitempty"`
	SourceName  string           `json:"source_name,omitempty"`
}

// PendingTaskInfo contains the information for a task pending user confirmation
type PendingTaskInfo struct {
	Name      string `json:"name"`
	SourceID  int64  `json:"source_id"`
	Frequency string `json:"frequency"`
	Command   string `json:"command"`
}

// HandleCreateTaskWithAI handles AI-powered task creation
// POST /api/tasks/ai-create
func (ah *AITaskHandler) HandleCreateTaskWithAI(c *gin.Context) {
	var req AICreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Message cannot be empty"})
		return
	}

	log.Printf("[AI Task] User input: %s", req.Message)

	// Step 1: Get all default sources for matching
	ctx := context.Background()
	sources, err := ah.sourceRepo.GetAll(ctx, true)
	if err != nil {
		log.Printf("[AI Task] Error fetching sources: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sources"})
		return
	}

	// Step 2: Call Python backend to analyze user input and suggest sources
	response, err := ah.callPythonAIAnalysis(req.Message, sources, req.ConversationHistory)
	if err != nil {
		log.Printf("[AI Task] Error calling Python AI: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI analysis failed"})
		return
	}

	log.Printf("[AI Task] AI Analysis Response: %+v", response)

	c.JSON(http.StatusOK, response)
}

// callPythonAIAnalysis calls the Python backend to analyze the user input
func (ah *AITaskHandler) callPythonAIAnalysis(
	userMessage string,
	sources []*models.Source,
	conversationHistory []ChatMessage,
) (*AICreateTaskResponse, error) {
	// Build source list for Python backend
	sourceList := make([]map[string]interface{}, len(sources))
	for i, source := range sources {
		sourceList[i] = map[string]interface{}{
			"id":                source.ID,
			"name":              source.AuthorName,
			"url":               source.URL,
			"platform":          source.Platform,
			"fetch_interval":    source.FetchIntervalSeconds,
		}
	}

	// Prepare conversation history
	history := make([]map[string]interface{}, len(conversationHistory))
	for i, msg := range conversationHistory {
		history[i] = map[string]interface{}{
			"role":    msg.Role,
			"content": msg.Content,
		}
	}

	// TODO: Call Python API in the future
	// pythonURL := fmt.Sprintf("%s/api/task/ai-create", ah.pythonAPIBaseURL)
	// payload := map[string]interface{}{
	// 	"message":                userMessage,
	// 	"sources":                sourceList,
	// 	"conversation_history":   history,
	// }
	// payloadBytes, _ := json.Marshal(payload)

	// For now, we'll do a simple implementation locally
	return ah.analyzeUserInputLocally(userMessage, sources, conversationHistory)
}

// analyzeUserInputLocally analyzes user input locally (placeholder implementation)
// In the future, this should call the Python backend
func (ah *AITaskHandler) analyzeUserInputLocally(
	userMessage string,
	sources []*models.Source,
	conversationHistory []ChatMessage,
) (*AICreateTaskResponse, error) {
	response := &AICreateTaskResponse{}

	// Simple keyword matching to find matching sources
	keywords := strings.ToLower(userMessage)
	var matchedSource *models.Source

	for _, source := range sources {
		sourceName := strings.ToLower(source.AuthorName)
		if strings.Contains(keywords, sourceName) || strings.Contains(keywords, source.Platform) {
			matchedSource = source
			break
		}
	}

	if matchedSource != nil {
		// Found a matching source
		response.Reply = fmt.Sprintf("我找到了 %s 的订阅源。你想要创建一个监控任务吗？\n\n任务名称会是：\"监控 %s\"，执行频率为每天。确认创建吗？",
			matchedSource.AuthorName, matchedSource.AuthorName)
		response.SourceName = matchedSource.AuthorName
		response.PendingTask = &PendingTaskInfo{
			Name:      fmt.Sprintf("监控 %s", matchedSource.AuthorName),
			SourceID:  matchedSource.ID,
			Frequency: "daily",
			Command:   fmt.Sprintf("监控 %s 源", matchedSource.AuthorName),
		}
	} else {
		// No matching source found
		response.Reply = "我们的默认源中没有找到匹配的订阅源。你可以提供 RSS 链接，或者告诉我你想要监控的具体内容？"
	}

	return response, nil
}
