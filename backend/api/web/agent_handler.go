package api

import (
	"net/http"
	"railgun-core/internal/usecase"

	"github.com/gin-gonic/gin"
)

type AgentHandler struct {
	service *usecase.AgentService
}

func NewAgentHandler(service *usecase.AgentService) *AgentHandler {
	return &AgentHandler{service: service}
}

func (h *AgentHandler) GetTask(c *gin.Context) {
	hostID := c.GetHeader("X-Agent-ID")
	if hostID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Agent ID missing"})
		return
	}

	task, err := h.service.GetAgentTask(c.Request.Context(), hostID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if task == nil {
		c.JSON(http.StatusOK, gin.H{"message": "no tasks"})
		return
	}

	c.JSON(http.StatusOK, task)
}

func (h *AgentHandler) ReportResult(c *gin.Context) {
	var input struct {
		TaskID int64  `json:"task_id"`
		Status string `json:"status"`
		Output string `json:"output"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hostID := c.GetHeader("X-Agent-ID")

	err := h.service.ReportTaskResult(c.Request.Context(), input.TaskID, input.Status, input.Output, hostID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "report accepted"})
}
