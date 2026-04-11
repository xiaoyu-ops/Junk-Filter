package handlers

import "github.com/gin-gonic/gin"

// RegisterSourceRoutes registers source-related routes
func RegisterSourceRoutes(router *gin.Engine, handler *SourceHandler) {
	sources := router.Group("/api/sources")
	{
		sources.GET("", handler.ListSources)
		sources.GET("/search", handler.SearchSources)  // ← 必须在 /:id 之前
		sources.POST("/:id/fetch", handler.FetchSourceNow)  // ← 手动同步
		sources.PUT("/:id/author-filter", handler.UpdateAuthorFilter)
		sources.GET("/:id/authors", handler.GetSourceAuthors)
		sources.GET("/:id", handler.GetSource)
		sources.POST("", handler.CreateSource)
		sources.PUT("/:id", handler.UpdateSource)
		sources.DELETE("/:id", handler.DeleteSource)
	}
}

// RegisterContentRoutes registers content-related routes
func RegisterContentRoutes(router *gin.Engine, handler *ContentHandler) {
	content := router.Group("/api/content")
	{
		content.GET("/stats", handler.GetContentStats)          // ← 必须在 /:id 之前
		content.GET("/stats/timeline", handler.GetContentTimeline) // 近N天评估趋势
		content.POST("/stop-evaluation", handler.StopEvaluation)
		content.POST("/restart-evaluation", handler.RestartEvaluation)
		content.GET("", handler.ListContent)
		content.GET("/:id", handler.GetContent)
	}
}

// RegisterEvaluationRoutes registers evaluation-related routes
func RegisterEvaluationRoutes(router *gin.Engine, handler *EvaluationHandler) {
	eval := router.Group("/api/evaluations")
	{
		eval.GET("", handler.ListEvaluationsByDecision)
		eval.GET("/high-scores", handler.ListHighScores)
	}
}

// RegisterMessageRoutes registers message-related routes
func RegisterMessageRoutes(router *gin.Engine, handler *MessageHandler) {
	// GET /api/tasks/{id}/messages
	router.GET("/api/tasks/:id/messages", handler.GetTaskMessages)

	// POST /api/tasks/{id}/messages
	router.POST("/api/tasks/:id/messages", handler.CreateMessageForTask)

	// Alternative routes
	messages := router.Group("/api/messages")
	{
		messages.GET("", handler.GetMessages)
		messages.POST("", handler.CreateMessage)
	}

	// DELETE /api/tasks/{id}/messages
	router.DELETE("/api/tasks/:id/messages", handler.DeleteTaskMessages)
}

// RegisterTaskChatRoutes registers task-specific chat routes (Agent tuning & consultation)
func RegisterTaskChatRoutes(router *gin.Engine, handler *TaskChatHandler) {
	// POST /api/tasks/{id}/chat - New endpoint for task-specific chat
	router.POST("/api/tasks/:id/chat", handler.HandleTaskChat)
}

// RegisterConfigRoutes registers LLM config routes
func RegisterConfigRoutes(router *gin.Engine, handler *ConfigHandler) {
	router.GET("/api/config/llm", handler.GetLLMConfig)
	router.POST("/api/config/llm", handler.SaveLLMConfig)
}

// RegisterAITaskRoutes registers AI-powered task creation routes
func RegisterAITaskRoutes(router *gin.Engine, handler *AITaskHandler) {
	router.POST("/api/tasks/ai-create", handler.HandleCreateTaskWithAI)
}

// RegisterNotificationRoutes registers notification-related routes
func RegisterNotificationRoutes(router *gin.Engine, handler *NotificationHandler) {
	notifications := router.Group("/api/notifications")
	{
		notifications.GET("", handler.ListNotifications)
		notifications.GET("/stream", handler.StreamNotifications)
		notifications.GET("/settings", handler.GetSettings)
		notifications.PUT("/settings", handler.UpdateSettings)
		notifications.POST("/test-push", handler.TestPush)
		notifications.PUT("/read-all", handler.MarkAllAsRead)
		notifications.PUT("/:id/read", handler.MarkAsRead)
	}
}

// RegisterThreadRoutes registers thread-related routes
func RegisterThreadRoutes(router *gin.Engine, handler *ThreadHandler) {
	router.GET("/api/tasks/:id/threads", handler.ListThreads)
	router.POST("/api/tasks/:id/threads", handler.CreateThread)
	router.PUT("/api/threads/:id", handler.UpdateThread)
	router.DELETE("/api/threads/:id", handler.DeleteThread)
	router.GET("/api/threads/:id/messages", handler.GetThreadMessages)
}

