package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"railgun-core/internal/domain"
	"railgun-core/internal/domain/models"
)

type ArtifactHandler struct {
	artifactRepo domain.ArtifactRepository
}

func NewArtifactHandler(artifactRepo domain.ArtifactRepository) *ArtifactHandler {
	return &ArtifactHandler{
		artifactRepo: artifactRepo,
	}
}

func (h *ArtifactHandler) GetArtifactsByHost(c *gin.Context) {
	hostID := c.Param("hostId")
	if hostID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Host ID is required"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	artifacts, total, err := h.artifactRepo.GetArtifactsByHost(c, hostID, page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": artifacts,
		"meta": gin.H{
			"total":       total,
			"page":        page,
			"per_page":    perPage,
			"total_pages": (total + perPage - 1) / perPage,
		},
	})
}

func (h *ArtifactHandler) GetArtifactByID(c *gin.Context) {
	idStr := c.Param("id")

	// Пробуем как UUID
	artifact, err := h.artifactRepo.GetArtifactByUUID(c, idStr)
	if err == nil {
		c.JSON(http.StatusOK, artifact)
		return
	}

	// Если не нашли по UUID, пробуем как число (для обратной совместимости)
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Artifact not found"})
		return
	}

	artifact, err = h.artifactRepo.GetArtifactByID(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Artifact not found"})
		return
	}

	c.JSON(http.StatusOK, artifact)
}

func (h *ArtifactHandler) SaveArtifact(c *gin.Context) {
	var artifact models.WindowsArtifact

	if err := c.ShouldBindJSON(&artifact); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Генерируем UUID если не предоставлен
	if artifact.UUID == "" {
		artifact.UUID = uuid.New().String()
	}

	// Устанавливаем timestamp если не предоставлен
	if artifact.Timestamp.IsZero() {
		artifact.Timestamp = time.Now().UTC()
	}

	// Сохраняем артефакт
	err := h.artifactRepo.SaveArtifact(c, &artifact)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Artifact saved successfully",
		"id":      artifact.UUID, // Возвращаем UUID
	})
}

func (h *ArtifactHandler) SearchArtifacts(c *gin.Context) {
	query := c.Query("q")
	artifactType := c.Query("type")
	severity := c.Query("severity")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	artifacts, total, err := h.artifactRepo.SearchArtifacts(c, query, artifactType, severity, page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": artifacts,
		"meta": gin.H{
			"total":       total,
			"page":        page,
			"per_page":    perPage,
			"total_pages": (total + perPage - 1) / perPage,
		},
	})
}
