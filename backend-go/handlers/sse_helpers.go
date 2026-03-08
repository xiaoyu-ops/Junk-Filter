package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// SSEEvent represents a server-sent event
type SSEEvent struct {
	Status string      `json:"status"`
	Text   string      `json:"text,omitempty"`
	Data   interface{} `json:"data,omitempty"`
	Error  string      `json:"error,omitempty"`
}

// sendSSEEvent writes a single SSE event to the response
func sendSSEEvent(w http.ResponseWriter, flusher http.Flusher, event SSEEvent) {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling SSE event: %v", err)
		return
	}

	fmt.Fprintf(w, "data: %s\n\n", string(eventJSON))
	flusher.Flush()
}
