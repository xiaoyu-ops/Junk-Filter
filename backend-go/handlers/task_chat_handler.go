package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/junkfilter/backend-go/models"
	"github.com/junkfilter/backend-go/repositories"
)

// TaskChatHandler handles chat requests specific to a task (Agent tuning & consultation)
type TaskChatHandler struct {
	messageRepo      *repositories.MessageRepository
	sourceRepo       *repositories.SourceRepository
	evaluationRepo   *repositories.EvaluationRepository
	pythonAPIBaseURL string
}

// NewTaskChatHandler creates a new task chat handler
func NewTaskChatHandler(
	messageRepo *repositories.MessageRepository,
	sourceRepo *repositories.SourceRepository,
	evaluationRepo *repositories.EvaluationRepository,
	pythonAPIBaseURL string,
) *TaskChatHandler {
	return &TaskChatHandler{
		messageRepo:      messageRepo,
		sourceRepo:       sourceRepo,
		evaluationRepo:   evaluationRepo,
		pythonAPIBaseURL: pythonAPIBaseURL,
	}
}

// TaskChatRequest represents a user query for task-specific consultation
type TaskChatRequest struct {
	Message string `json:"message" binding:"required"`
}

// TaskChatResponse represents the AI's response to a user query
type TaskChatResponse struct {
	Reply              string          `json:"reply"`                   // 自然语言回复
	ReferencedCardIDs  []int64         `json:"referenced_card_ids"`    // 引用的评估卡片 ID
	ParameterUpdates   map[string]interface{} `json:"parameter_updates"` // 参数变更建议
	ContextUsed        map[string]interface{} `json:"context_used"`      // 用于调试：包含了什么上下文
}

// AgentContext contains all context for the Agent during a chat session
type AgentContext struct {
	TaskMetadata   map[string]interface{} `json:"task_metadata"`
	ChatHistory    []models.Message       `json:"chat_history"`
	RecentCards    []map[string]interface{} `json:"recent_cards"`
	CurrentConfig  map[string]interface{} `json:"current_config"`
	TaskDescription string                `json:"task_description"`
}

// HandleTaskChat handles streaming chat requests specific to a task
// POST /api/tasks/{id}/chat
func (ch *TaskChatHandler) HandleTaskChat(c *gin.Context) {
	taskIDStr := c.Param("id")

	// Parse task ID from "source-123" format or plain number
	var taskID int64
	if strings.HasPrefix(taskIDStr, "source-") {
		var err error
		taskID, err = strconv.ParseInt(strings.TrimPrefix(taskIDStr, "source-"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID format"})
			return
		}
	} else {
		var err error
		taskID, err = strconv.ParseInt(taskIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID format"})
			return
		}
	}

	// Parse request body
	var req TaskChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Message cannot be empty"})
		return
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	w := c.Writer
	flusher, ok := w.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Streaming not supported"})
		return
	}

	log.Printf("[Task Chat] TaskID: %d, Message: %s", taskID, req.Message)

	// Send initial processing status
	sendSSEEvent(w, flusher, SSEEvent{
		Status: "processing",
		Data:   map[string]string{"phase": "gathering_context"},
	})

	// Step 1: Gather context
	ctx := context.Background()
	agentCtx, err := ch.gatherAgentContext(ctx, taskID)
	if err != nil {
		log.Printf("[Task Chat] Error gathering context: %v", err)
		sendSSEEvent(w, flusher, SSEEvent{
			Status: "error",
			Error:  "Failed to gather context: " + err.Error(),
		})
		return
	}

	log.Printf("[Task Chat] Context gathered. Task: %v", agentCtx.TaskMetadata["name"])

	// Step 2: Save user message to database
	metadataJSON, _ := json.Marshal(map[string]interface{}{
		"message_type": "user_query",
		"context_snapshot": agentCtx,
	})
	metadataStr := string(metadataJSON)

	userMsg := &models.Message{
		TaskID:    taskID,
		Role:      "user",
		Type:      "text",
		Content:   req.Message,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata:  &metadataStr,
	}

	userMsgID, err := ch.messageRepo.Create(ctx, userMsg)
	if err != nil {
		log.Printf("[Task Chat] Error saving user message: %v", err)
		// Don't fail, continue
	} else {
		userMsg.ID = userMsgID
		log.Printf("[Task Chat] User message saved: ID=%d", userMsgID)
	}

	// Step 3: Call Python backend with full context
	sendSSEEvent(w, flusher, SSEEvent{
		Status: "processing",
		Data:   map[string]string{"phase": "calling_agent"},
	})

	pythonURL := fmt.Sprintf("%s/api/task/%d/chat", ch.pythonAPIBaseURL, taskID)

	payload := map[string]interface{}{
		"message":        req.Message,
		"task_id":        taskID,
		"agent_context":  agentCtx,
	}

	payloadBytes, _ := json.Marshal(payload)

	log.Printf("[Task Chat] Calling Python at %s", pythonURL)

	pythonReq, _ := http.NewRequest("POST", pythonURL, strings.NewReader(string(payloadBytes)))
	pythonReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(pythonReq)
	if err != nil {
		log.Printf("[Task Chat] Error calling Python API: %v", err)
		sendSSEEvent(w, flusher, SSEEvent{
			Status: "error",
			Error:  "Failed to reach Agent: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[Task Chat] Python API error: %s", string(body))
		sendSSEEvent(w, flusher, SSEEvent{
			Status: "error",
			Error:  fmt.Sprintf("Agent returned status %d", resp.StatusCode),
		})
		return
	}

	// Step 4: Stream response from Python directly
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("[Task Chat] Error streaming from Python: %v", err)
		return
	}

	// Step 5: Send stream end marker
	sendSSEEvent(w, flusher, SSEEvent{
		Status: "stream_end",
	})

	log.Printf("[Task Chat] Completed for TaskID: %d", taskID)
}

// gatherAgentContext collects all contextual information for the Agent
func (ch *TaskChatHandler) gatherAgentContext(ctx context.Context, taskID int64) (*AgentContext, error) {
	agentCtx := &AgentContext{
		TaskMetadata:  make(map[string]interface{}),
		ChatHistory:   []models.Message{},
		RecentCards:   []map[string]interface{}{},
		CurrentConfig: make(map[string]interface{}),
	}

	// 1. Get task metadata
	source, err := ch.sourceRepo.GetByID(ctx, taskID)
	if err == nil && source != nil {
		agentCtx.TaskMetadata = map[string]interface{}{
			"id":               source.ID,
			"name":             source.AuthorName,
			"url":              source.URL,
			"priority":         source.Priority,
			"enabled":          source.Enabled,
			"last_fetch_time":  source.LastFetchTime,
		}
		agentCtx.TaskDescription = fmt.Sprintf("正在监控 RSS 源：%s (优先级 %d)", source.AuthorName, source.Priority)
	}

	// 2. Get recent chat history (last 10 messages)
	messages, err := ch.messageRepo.GetByTaskID(ctx, taskID)
	if err == nil && len(messages) > 0 {
		// Keep only last 10
		start := 0
		if len(messages) > 10 {
			start = len(messages) - 10
		}
		agentCtx.ChatHistory = messages[start:]
	}

	// 3. Get recent evaluation cards (for context/reference)
	// Note: Assuming evaluationRepo.GetRecentByTaskID exists or can be adapted
	// For now, we'll add a placeholder
	agentCtx.RecentCards = []map[string]interface{}{
		// This would be populated by queries to the evaluation table
		// Format: {id, decision, innovation_score, depth_score, tldr, key_concepts, timestamp}
	}

	// 4. Current Agent configuration (from settings or task-specific overrides)
	agentCtx.CurrentConfig = map[string]interface{}{
		"temperature": 0.7,
		"topP":        0.9,
		"maxTokens":   2000,
		"filter_rules": "default",  // Would be fetched from task config
	}

	return agentCtx, nil
}
