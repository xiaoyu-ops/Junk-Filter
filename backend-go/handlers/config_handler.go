package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ConfigHandler struct {
	db *sql.DB
}

func NewConfigHandler(db *sql.DB) *ConfigHandler {
	return &ConfigHandler{db: db}
}

type LLMConfigRequest struct {
	APIKey      string  `json:"api_key" binding:"required"`
	ModelName   string  `json:"model_name" binding:"required"`
	BaseURL     string  `json:"base_url"`
	Temperature float64 `json:"temperature"`
	MaxTokens   int     `json:"max_tokens"`
}

// SaveLLMConfig saves the LLM config from frontend into ai_config table
func (h *ConfigHandler) SaveLLMConfig(c *gin.Context) {
	var req LLMConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Temperature == 0 {
		req.Temperature = 0.7
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = 2048
	}

	// Upsert: update existing row (id=1) or insert if not exists
	_, err := h.db.ExecContext(c.Request.Context(), `
		INSERT INTO ai_config (id, api_key, default_model, base_url, temperature, max_tokens, updated_at)
		VALUES (1, $1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			api_key      = EXCLUDED.api_key,
			default_model = EXCLUDED.default_model,
			base_url     = EXCLUDED.base_url,
			temperature  = EXCLUDED.temperature,
			max_tokens   = EXCLUDED.max_tokens,
			updated_at   = EXCLUDED.updated_at
	`, req.APIKey, req.ModelName, req.BaseURL, req.Temperature, req.MaxTokens, time.Now())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "LLM config saved"})
}

// GetLLMConfig returns the current LLM config (api_key masked)
func (h *ConfigHandler) GetLLMConfig(c *gin.Context) {
	var model, baseURL, apiKey string
	var temperature float64
	var maxTokens int

	err := h.db.QueryRowContext(c.Request.Context(),
		`SELECT default_model, base_url, api_key, temperature, max_tokens FROM ai_config WHERE id = 1`,
	).Scan(&model, &baseURL, &apiKey, &temperature, &maxTokens)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"model_name":  "",
			"base_url":    "",
			"api_key":     "",
			"temperature": 0.7,
			"max_tokens":  2048,
		})
		return
	}

	// Mask api_key
	maskedKey := ""
	if len(apiKey) > 8 {
		maskedKey = apiKey[:8] + "****"
	}

	c.JSON(http.StatusOK, gin.H{
		"model_name":  model,
		"base_url":    baseURL,
		"api_key":     maskedKey,
		"temperature": temperature,
		"max_tokens":  maxTokens,
	})
}
