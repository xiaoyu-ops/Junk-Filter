package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// NotificationHandler handles notification-related HTTP requests
type NotificationHandler struct {
	db              *sql.DB
	redis           *redis.Client
	pythonAPIBaseURL string
}

// Notification represents a notification record
type Notification struct {
	ID              int64     `json:"id"`
	ContentID       int64     `json:"content_id"`
	Title           string    `json:"title"`
	Summary         string    `json:"summary"`
	InnovationScore int       `json:"innovation_score"`
	DepthScore      int       `json:"depth_score"`
	Decision        string    `json:"decision"`
	IsRead          bool      `json:"is_read"`
	CreatedAt       time.Time `json:"created_at"`
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(db *sql.DB, redisClient *redis.Client, pythonAPIBaseURL string) *NotificationHandler {
	return &NotificationHandler{db: db, redis: redisClient, pythonAPIBaseURL: pythonAPIBaseURL}
}

// ListNotifications returns notifications with optional unread filter
// GET /api/notifications?unread=true&limit=50
func (nh *NotificationHandler) ListNotifications(c *gin.Context) {
	query := `SELECT id, content_id, title, summary, innovation_score, depth_score, decision, is_read, created_at
	          FROM notifications`

	args := []interface{}{}
	argIdx := 1

	if c.Query("unread") == "true" {
		query += " WHERE is_read = false"
	}

	query += " ORDER BY created_at DESC"

	limit := 50
	if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 {
		limit = l
	}
	query += fmt.Sprintf(" LIMIT $%d", argIdx)
	args = append(args, limit)

	rows, err := nh.db.QueryContext(c.Request.Context(), query, args...)
	if err != nil {
		log.Printf("Error listing notifications: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list notifications"})
		return
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		var n Notification
		var summary sql.NullString
		err := rows.Scan(&n.ID, &n.ContentID, &n.Title, &summary, &n.InnovationScore, &n.DepthScore, &n.Decision, &n.IsRead, &n.CreatedAt)
		if err != nil {
			log.Printf("Error scanning notification: %v", err)
			continue
		}
		if summary.Valid {
			n.Summary = summary.String
		}
		notifications = append(notifications, n)
	}

	if notifications == nil {
		notifications = []Notification{}
	}

	// Get unread count
	var unreadCount int
	nh.db.QueryRowContext(c.Request.Context(), "SELECT COUNT(*) FROM notifications WHERE is_read = false").Scan(&unreadCount)

	c.JSON(http.StatusOK, gin.H{
		"data":         notifications,
		"unread_count": unreadCount,
	})
}

// MarkAsRead marks a notification as read
// PUT /api/notifications/:id/read
func (nh *NotificationHandler) MarkAsRead(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	_, err = nh.db.ExecContext(c.Request.Context(), "UPDATE notifications SET is_read = true WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Marked as read"})
}

// MarkAllAsRead marks all notifications as read
// PUT /api/notifications/read-all
func (nh *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	_, err := nh.db.ExecContext(c.Request.Context(), "UPDATE notifications SET is_read = true WHERE is_read = false")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark all as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All marked as read"})
}

// NotificationSettings represents user-configurable notification rules
type NotificationSettings struct {
	MinInnovationScore  int                      `json:"min_innovation_score"`
	MinDepthScore       int                      `json:"min_depth_score"`
	NotifyOnInteresting bool                     `json:"notify_on_interesting"`
	WatchedSourceIDs    []int                    `json:"watched_source_ids"`
	Enabled             bool                     `json:"enabled"`
	PushChannels        []map[string]interface{} `json:"push_channels"`
}

// GetSettings returns the current notification settings
// GET /api/notifications/settings
func (nh *NotificationHandler) GetSettings(c *gin.Context) {
	var settings NotificationSettings
	var watchedJSON, pushJSON []byte

	err := nh.db.QueryRowContext(c.Request.Context(),
		`SELECT min_innovation_score, min_depth_score, notify_on_interesting, watched_source_ids, enabled, push_channels
		 FROM notification_settings WHERE id = 1`,
	).Scan(&settings.MinInnovationScore, &settings.MinDepthScore, &settings.NotifyOnInteresting, &watchedJSON, &settings.Enabled, &pushJSON)

	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("[Notification] Error reading settings: %v", err)
		}
		c.JSON(http.StatusOK, NotificationSettings{
			MinInnovationScore:  8,
			MinDepthScore:       7,
			NotifyOnInteresting: true,
			WatchedSourceIDs:    []int{},
			Enabled:             true,
			PushChannels:        []map[string]interface{}{},
		})
		return
	}

	if watchedJSON != nil {
		if err := json.Unmarshal(watchedJSON, &settings.WatchedSourceIDs); err != nil {
			log.Printf("[Notification] Error parsing watched_source_ids: %v", err)
		}
	}
	if settings.WatchedSourceIDs == nil {
		settings.WatchedSourceIDs = []int{}
	}
	if pushJSON != nil {
		if err := json.Unmarshal(pushJSON, &settings.PushChannels); err != nil {
			log.Printf("[Notification] Error parsing push_channels: %v", err)
		}
	}
	if settings.PushChannels == nil {
		settings.PushChannels = []map[string]interface{}{}
	}

	c.JSON(http.StatusOK, settings)
}

// UpdateSettings updates the notification settings
// PUT /api/notifications/settings
func (nh *NotificationHandler) UpdateSettings(c *gin.Context) {
	var req NotificationSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.WatchedSourceIDs == nil {
		req.WatchedSourceIDs = []int{}
	}
	if req.PushChannels == nil {
		req.PushChannels = []map[string]interface{}{}
	}
	watchedJSON, _ := json.Marshal(req.WatchedSourceIDs)
	pushJSON, _ := json.Marshal(req.PushChannels)

	_, err := nh.db.ExecContext(c.Request.Context(),
		`INSERT INTO notification_settings (id, min_innovation_score, min_depth_score, notify_on_interesting, watched_source_ids, enabled, push_channels, updated_at)
		 VALUES (1, $1, $2, $3, $4, $5, $6, NOW())
		 ON CONFLICT (id) DO UPDATE SET
		   min_innovation_score = $1, min_depth_score = $2, notify_on_interesting = $3,
		   watched_source_ids = $4, enabled = $5, push_channels = $6, updated_at = NOW()`,
		req.MinInnovationScore, req.MinDepthScore, req.NotifyOnInteresting, watchedJSON, req.Enabled, pushJSON,
	)
	if err != nil {
		log.Printf("Error updating notification settings: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Settings updated", "settings": req})
}

// StreamNotifications provides SSE endpoint for real-time notification push
// GET /api/notifications/stream
func (nh *NotificationHandler) StreamNotifications(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Streaming not supported"})
		return
	}

	// Subscribe to Redis Pub/Sub notifications channel
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	pubsub := nh.redis.Subscribe(ctx, "notifications")
	defer pubsub.Close()

	ch := pubsub.Channel()

	// Send initial heartbeat
	fmt.Fprintf(c.Writer, "data: {\"type\":\"connected\"}\n\n")
	flusher.Flush()

	// Listen for notifications
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return
			}
			// Parse and forward the notification
			var notification map[string]interface{}
			if err := json.Unmarshal([]byte(msg.Payload), &notification); err != nil {
				continue
			}
			data, _ := json.Marshal(gin.H{
				"type": "notification",
				"data": notification,
			})
			fmt.Fprintf(c.Writer, "data: %s\n\n", string(data))
			flusher.Flush()

		case <-ctx.Done():
			return

		case <-time.After(30 * time.Second):
			// Heartbeat
			fmt.Fprintf(c.Writer, "data: {\"type\":\"heartbeat\"}\n\n")
			flusher.Flush()
		}
	}
}

// TestPush proxies the test-push request to the Python backend
// POST /api/notifications/test-push
func (nh *NotificationHandler) TestPush(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	pythonURL := fmt.Sprintf("%s/api/notifications/test-push", nh.pythonAPIBaseURL)
	resp, err := http.Post(pythonURL, "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("[TestPush] Failed to reach Python API: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to reach push service"})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, "application/json", respBody)
}
