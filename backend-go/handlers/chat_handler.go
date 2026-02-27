package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/junkfilter/backend-go/repositories"
)

// ChatHandler handles chat stream requests
type ChatHandler struct {
	messageRepo      *repositories.MessageRepository
	pythonAPIBaseURL string
}

// NewChatHandler creates a new chat handler
func NewChatHandler(messageRepo *repositories.MessageRepository, pythonAPIBaseURL string) *ChatHandler {
	return &ChatHandler{
		messageRepo:      messageRepo,
		pythonAPIBaseURL: pythonAPIBaseURL,
	}
}

// ChatStreamRequest represents incoming chat request
type ChatStreamRequest struct {
	TaskID      int64   `form:"taskId" binding:"required"`
	Message     string  `form:"message" binding:"required"`
	Temperature float32 `form:"temperature"`
	TopP        float32 `form:"topP"`
	MaxTokens   int     `form:"maxTokens"`
}

// SSEEvent represents a server-sent event
type SSEEvent struct {
	Status string      `json:"status"`
	Text   string      `json:"text,omitempty"`
	Data   interface{} `json:"data,omitempty"`
	Error  string      `json:"error,omitempty"`
}

// ChatStream handles streaming chat requests
// GET /api/chat/stream?taskId=1&message=hello&temperature=0.7&topP=0.9&maxTokens=2048
func (ch *ChatHandler) ChatStream(c *gin.Context) {
	taskIDStr := c.Query("taskId")
	message := c.Query("message")
	temperatureStr := c.Query("temperature")
	topPStr := c.Query("topP")
	maxTokensStr := c.Query("maxTokens")

	if taskIDStr == "" || message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "taskId and message query parameters required"})
		return
	}

	// Parse optional configuration parameters
	temperature := float32(0.7) // default
	if temperatureStr != "" {
		if t, err := strconv.ParseFloat(temperatureStr, 32); err == nil {
			temperature = float32(t)
		}
	}

	topP := float32(0.9) // default
	if topPStr != "" {
		if p, err := strconv.ParseFloat(topPStr, 32); err == nil {
			topP = float32(p)
		}
	}

	maxTokens := 2000 // default
	if maxTokensStr != "" {
		if m, err := strconv.Atoi(maxTokensStr); err == nil {
			maxTokens = m
		}
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// Get the response writer for flushing
	w := c.Writer
	flusher, ok := w.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Streaming not supported"})
		return
	}

	// Parse taskID as int64
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		log.Printf("[Chat Stream] Invalid task ID format: %s", taskIDStr)
		sendSSEEvent(w, flusher, SSEEvent{
			Status: "error",
			Error:  "Invalid task ID format",
		})
		return
	}

	// Log the request
	log.Printf("[Chat Stream] TaskID: %d, Message: %s, Temperature: %.2f, TopP: %.2f, MaxTokens: %d",
		taskID, message, temperature, topP, maxTokens)

	// 1. Send processing status
	sendSSEEvent(w, flusher, SSEEvent{
		Status: "processing",
	})

	// 2. Call Python backend for evaluation/response with configuration
	pythonURL := fmt.Sprintf("%s/api/evaluate/stream", ch.pythonAPIBaseURL)

	payload := map[string]interface{}{
		"title":       "User Query",
		"content":     message,
		"temperature": temperature,
		"topP":        topP,
		"maxTokens":   maxTokens,
	}

	payloadBytes, _ := json.Marshal(payload)

	log.Printf("[Calling Python] URL: %s, Payload: %s", pythonURL, string(payloadBytes))

	req, _ := http.NewRequest("POST", pythonURL, strings.NewReader(string(payloadBytes)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error calling Python API: %v", err)
		sendSSEEvent(w, flusher, SSEEvent{
			Status: "error",
			Error:  fmt.Sprintf("Failed to call evaluation service: %v", err),
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Python API error: %s", string(body))
		sendSSEEvent(w, flusher, SSEEvent{
			Status: "error",
			Error:  fmt.Sprintf("Evaluation service returned status %d", resp.StatusCode),
		})
		return
	}

	// 4. Stream response from Python backend directly
	// Use io.Copy to directly relay the streaming response from Python
	// This preserves the exact SSE format and avoids connection issues
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("Error streaming from Python API: %v", err)
		// Don't try to send error SSE event here as the connection may be broken
		return
	}

	// 5. Send stream end marker to indicate successful completion
	sendSSEEvent(w, flusher, SSEEvent{
		Status: "stream_end",
	})

	// 6. Log completion
	log.Printf("[Chat Stream Completed] TaskID: %d", taskID)
}

// sendSSEEvent sends a single SSE event
func sendSSEEvent(w http.ResponseWriter, flusher http.Flusher, event SSEEvent) {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling SSE event: %v", err)
		return
	}

	fmt.Fprintf(w, "data: %s\n\n", string(eventJSON))
	flusher.Flush()
}
