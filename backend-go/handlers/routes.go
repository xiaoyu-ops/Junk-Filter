package handlers

import "github.com/gin-gonic/gin"

// RegisterSourceRoutes registers source-related routes
func RegisterSourceRoutes(router *gin.Engine, handler *SourceHandler) {
	sources := router.Group("/api/sources")
	{
		sources.GET("", handler.ListSources)
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

// RegisterChatRoutes registers chat-related routes (RSS content evaluation - DEPRECATED)
func RegisterChatRoutes(router *gin.Engine, handler *ChatHandler) {
	// GET /api/chat/stream?taskId=1&message=hello (OLD - for RSS content evaluation)
	router.GET("/api/chat/stream", handler.ChatStream)
}

// RegisterTaskChatRoutes registers task-specific chat routes (Agent tuning & consultation)
func RegisterTaskChatRoutes(router *gin.Engine, handler *TaskChatHandler) {
	// POST /api/tasks/{id}/chat - New endpoint for task-specific chat
	router.POST("/api/tasks/:id/chat", handler.HandleTaskChat)
}
