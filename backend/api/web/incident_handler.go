package api

import (
	"net/http"
	"strconv"

	"railgun-core/internal/domain"

	"github.com/gin-gonic/gin"
)

type IncidentHandler struct {
	repo domain.IncidentRepository
}

func NewIncidentHandler(repo domain.IncidentRepository) *IncidentHandler {
	return &IncidentHandler{repo: repo}
}

// GetIncidents возвращает список последних инцидентов
func (h *IncidentHandler) GetIncidents(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, _ := strconv.Atoi(limitStr)

	incidents, err := h.repo.GetLatestIncidents(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch incidents"})
		return
	}

	c.JSON(http.StatusOK, incidents)
}
