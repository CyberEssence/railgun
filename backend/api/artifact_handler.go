package api

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"railgun-core/internal/domain"
	"railgun-core/internal/models"
)

type ArtifactHandler struct {
	artifactRepo domain.ArtifactRepository
}

func NewArtifactHandler(artifactRepo domain.ArtifactRepository) *ArtifactHandler {
	return &ArtifactHandler{
		artifactRepo: artifactRepo,
	}
}

// GetArtifactsByHost возвращает артефакты для указанного хоста
func (h *ArtifactHandler) GetArtifactsByHost(c *gin.Context) {
	hostID := c.Param("hostId")
	if hostID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Host ID is required"})
		return
	}

	// Получаем параметры пагинации
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	// Получаем артефакты
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

// GetArtifactByID возвращает артефакт по ID
func (h *ArtifactHandler) GetArtifactByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid artifact ID"})
		return
	}

	artifact, err := h.artifactRepo.GetArtifactByID(c, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Artifact not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, artifact)
}

// SaveArtifact сохраняет артефакт
func (h *ArtifactHandler) SaveArtifact(c *gin.Context) {
	var artifact models.WindowsArtifact
	if err := c.ShouldBindJSON(&artifact); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Сохраняем артефакт
	err := h.artifactRepo.SaveArtifact(c, &artifact)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Artifact saved successfully",
		"id":      artifact.ID,
	})
}

// SearchArtifacts выполняет поиск артефактов
func (h *ArtifactHandler) SearchArtifacts(c *gin.Context) {
	// Получаем параметры поиска
	query := c.Query("q")
	artifactType := c.Query("type")
	severity := c.Query("severity")

	// Получаем параметры пагинации
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	// Выполняем поиск
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
