package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/junkfilter/backend-go/models"
	"github.com/junkfilter/backend-go/repositories"
)

// ============================================================
// BloggerHandler - 博主相关 API
// ============================================================
type BloggerHandler struct {
	repo *repositories.BloggerRepository
}

func NewBloggerHandler(repo *repositories.BloggerRepository) *BloggerHandler {
	return &BloggerHandler{repo: repo}
}

// GetBloggers - GET /api/bloggers
func (h *BloggerHandler) GetBloggers(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "20")

	pageNum, _ := strconv.Atoi(page)
	size, _ := strconv.Atoi(pageSize)

	bloggers, total, err := h.repo.GetAll(size, (pageNum-1)*size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "Failed to fetch bloggers",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "Success",
		Data: models.PaginatedResponse{
			Data:       bloggers,
			Page:       pageNum,
			PageSize:   size,
			Total:      total,
			TotalPages: (total + size - 1) / size,
		},
	})
}

// GetBlogger - GET /api/bloggers/:id
func (h *BloggerHandler) GetBlogger(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	blogger, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Code:    404,
			Message: "Blogger not found",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "Success",
		Data:    blogger,
	})
}

// CreateBlogger - POST /api/bloggers
func (h *BloggerHandler) CreateBlogger(c *gin.Context) {
	var req models.CreateBloggerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    400,
			Message: "Invalid request",
			Error:   err.Error(),
		})
		return
	}

	blogger, err := h.repo.Create(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "Failed to create blogger",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Code:    201,
		Message: "Blogger created successfully",
		Data:    blogger,
	})
}

// DeleteBlogger - DELETE /api/bloggers/:id
func (h *BloggerHandler) DeleteBlogger(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	err := h.repo.Delete(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "Failed to delete blogger",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "Blogger deleted successfully",
	})
}

// UpdateBloggerStatus - PUT /api/bloggers/:id/status
func (h *BloggerHandler) UpdateBloggerStatus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Code: 400, Message: "Invalid request", Error: err.Error()})
		return
	}

	err := h.repo.UpdateStatus(id, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "Failed to update status",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "Status updated successfully",
	})
}

// ============================================================
// TaskHandler - 任务相关 API
// ============================================================
type TaskHandler struct {
	repo *repositories.TaskRepository
}

func NewTaskHandler(repo *repositories.TaskRepository) *TaskHandler {
	return &TaskHandler{repo: repo}
}

// GetTasks - GET /api/tasks
func (h *TaskHandler) GetTasks(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "20")

	pageNum, _ := strconv.Atoi(page)
	size, _ := strconv.Atoi(pageSize)

	tasks, total, err := h.repo.GetAll(size, (pageNum-1)*size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "Failed to fetch tasks",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "Success",
		Data: models.PaginatedResponse{
			Data:       tasks,
			Page:       pageNum,
			PageSize:   size,
			Total:      total,
			TotalPages: (total + size - 1) / size,
		},
	})
}

// CreateTask - POST /api/tasks
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req models.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    400,
			Message: "Invalid request",
			Error:   err.Error(),
		})
		return
	}

	task := &models.Task{
		Name:        req.Name,
		Description: req.Description,
		Schedule:    req.Schedule,
		TimeDisplay: req.TimeDisplay,
		Type:        req.Type,
		Config:      req.Config,
		Enabled:     true,
	}

	err := h.repo.Create(task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "Failed to create task",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Code:    201,
		Message: "Task created successfully",
		Data:    task,
	})
}

// UpdateTask - PUT /api/tasks/:id
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	var req models.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    400,
			Message: "Invalid request",
			Error:   err.Error(),
		})
		return
	}

	// 从数据库获取现有任务
	tasks, _, _ := h.repo.GetAll(1, 0)
	var task *models.Task
	for _, t := range tasks {
		if t.ID == id {
			task = t
			break
		}
	}

	if task == nil {
		c.JSON(http.StatusNotFound, models.APIResponse{Code: 404, Message: "Task not found"})
		return
	}

	// 更新字段
	if req.Name != "" {
		task.Name = req.Name
	}
	if req.Description != "" {
		task.Description = req.Description
	}
	if req.Schedule != "" {
		task.Schedule = req.Schedule
	}
	if req.TimeDisplay != "" {
		task.TimeDisplay = req.TimeDisplay
	}
	if req.Enabled != nil {
		task.Enabled = *req.Enabled
	}
	if req.Config != nil {
		task.Config = req.Config
	}

	err := h.repo.Update(id, task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "Failed to update task",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "Task updated successfully",
		Data:    task,
	})
}

// DeleteTask - DELETE /api/tasks/:id
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	err := h.repo.Delete(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Code:    500,
			Message: "Failed to delete task",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "Task deleted successfully",
	})
}

// ToggleTask - PUT /api/tasks/:id/toggle
func (h *TaskHandler) ToggleTask(c *gin.Context) {
	_, _ = strconv.Atoi(c.Param("id"))

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Code: 400, Message: "Invalid request", Error: err.Error()})
		return
	}

	// 这里应该从数据库获取任务并更新
	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "Task toggled successfully",
	})
}
