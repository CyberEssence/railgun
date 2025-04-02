package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"railgun-core/models"
	"railgun-core/services"
)

func RegisterRoutes(r *gin.Engine, trafficSvc *services.TrafficService, artifactSvc *services.ArtifactService) {
	api := r.Group("/api")

	// Traffic endpoints
	traffic := api.Group("/traffic")
	{
		traffic.GET("/host/:hostID", func(c *gin.Context) {
			hostID := c.Param("hostID")

			from := time.Now().Add(-24 * time.Hour)
			to := time.Now()

			if fromStr := c.Query("from"); fromStr != "" {
				if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
					from = t
				}
			}

			if toStr := c.Query("to"); toStr != "" {
				if t, err := time.Parse(time.RFC3339, toStr); err == nil {
					to = t
				}
			}

			data, err := trafficSvc.GetTrafficByHost(c.Request.Context(), hostID, from, to)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, data)
		})

		traffic.GET("/stats/host/:hostID", func(c *gin.Context) {
			hostID := c.Param("hostID")

			from := time.Now().Add(-24 * time.Hour)
			to := time.Now()

			if fromStr := c.Query("from"); fromStr != "" {
				if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
					from = t
				}
			}

			if toStr := c.Query("to"); toStr != "" {
				if t, err := time.Parse(time.RFC3339, toStr); err == nil {
					to = t
				}
			}

			stats, err := trafficSvc.GetTrafficStats(c.Request.Context(), hostID, from, to)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, stats)
		})

		traffic.POST("/", func(c *gin.Context) {
			var traffic models.NetworkTraffic

			if err := c.ShouldBindJSON(&traffic); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			if traffic.Timestamp.IsZero() {
				traffic.Timestamp = time.Now()
			}

			if err := trafficSvc.SaveTraffic(c.Request.Context(), traffic); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusCreated, traffic)
		})
	}

	// Artifact endpoints
	artifacts := api.Group("/artifacts")
	{
		artifacts.GET("/host/:hostID", func(c *gin.Context) {
			hostID := c.Param("hostID")
			artifactType := c.Query("type")

			from := time.Now().Add(-24 * time.Hour)
			to := time.Now()

			if fromStr := c.Query("from"); fromStr != "" {
				if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
					from = t
				}
			}

			if toStr := c.Query("to"); toStr != "" {
				if t, err := time.Parse(time.RFC3339, toStr); err == nil {
					to = t
				}
			}

			artifacts, err := artifactSvc.GetArtifactsByHost(c.Request.Context(), hostID, artifactType, from, to)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, artifacts)
		})

		artifacts.GET("/:id", func(c *gin.Context) {
			idStr := c.Param("id")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
				return
			}

			artifact, err := artifactSvc.GetArtifactByID(c.Request.Context(), id)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			if artifact == nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Artifact not found"})
				return
			}

			c.JSON(http.StatusOK, artifact)
		})

		artifacts.POST("/", func(c *gin.Context) {
			var artifact models.WindowsArtifact

			if err := c.ShouldBindJSON(&artifact); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			if artifact.Timestamp.IsZero() {
				artifact.Timestamp = time.Now()
			}

			if err := artifactSvc.SaveArtifact(c.Request.Context(), artifact); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusCreated, artifact)
		})

		artifacts.GET("/search", func(c *gin.Context) {
			query := c.Query("q")
			if query == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Search query cannot be empty"})
				return
			}

			results, err := artifactSvc.SearchArtifacts(c.Request.Context(), query)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, results)
		})
	}
}
