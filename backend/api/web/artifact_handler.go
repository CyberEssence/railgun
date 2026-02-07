package api

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

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

// GetArtifactsByHost godoc
// @Summary      Получить артефакты хоста
// @Description  Возвращает список артефактов для конкретного хоста с поддержкой пагинации
// @Tags         Artifacts
// @Produce      json
// @Param        hostId   path      string  true   "ID хоста"
// @Param        page     query     int     false  "Номер страницы" default(1)
// @Param        per_page query     int     false  "Количество на странице" default(20)
// @Success      200      {object}  map[string]interface{} "Объект с массивом data и мета-данными"
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Security     BearerAuth
// @Router       /artifacts/host/{hostId} [get]
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

// GetArtifactByID godoc
// @Summary      Получить артефакт по ID
// @Description  Возвращает детальную информацию об одном артефакте
// @Tags         Artifacts
// @Produce      json
// @Param        id   path      int  true  "ID артефакта"
// @Success      200  {object}  dto.WindowsArtifactDTO
// @Failure      400  {object}  map[string]string "Неверный формат ID"
// @Failure      404  {object}  map[string]string "Артефакт не найден"
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /artifacts/id/{id} [get]
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

// SaveArtifact godoc
// @Summary      Сохранить артефакт
// @Description  Создает новую запись артефакта Windows в системе
// @Tags         Artifacts
// @Accept       json
// @Produce      json
// @Param        artifact  body      models.WindowsArtifact  true  "Данные артефакта"
// @Success      201       {object}  map[string]interface{} "Успешное сохранение и ID записи"
// @Failure      400       {object}  map[string]string
// @Failure      500       {object}  map[string]string
// @Security     BearerAuth
// @Router       /artifacts [post]
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

// SearchArtifacts godoc
// @Summary      Поиск артефактов
// @Description  Поиск по артефактам с фильтрацией по типу, критичности и текстовому запросу
// @Tags         Artifacts
// @Produce      json
// @Param        q         query     string  false  "Поисковый запрос"
// @Param        type      query     string  false  "Тип артефакта"
// @Param        severity  query     string  false  "Уровень серьезности"
// @Param        page      query     int     false  "Номер страницы" default(1)
// @Param        per_page  query     int     false  "Количество на странице" default(20)
// @Success      200       {object}  map[string]interface{} "Результаты поиска и мета-данные"
// @Failure      500       {object}  map[string]string
// @Security     BearerAuth
// @Router       /artifacts/search [get]
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
