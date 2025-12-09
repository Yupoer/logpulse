package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Yupoer/logpulse/internal/domain"
	"github.com/Yupoer/logpulse/internal/service"
	"github.com/gin-gonic/gin"
)

type LogHandler struct {
	service *service.LogService
}

func NewLogHandler(service *service.LogService) *LogHandler {
	return &LogHandler{service: service}
}

func (h *LogHandler) CreateLog(c *gin.Context) {
	var entry domain.LogEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	totalCount, err := h.service.CreateLog(c.Request.Context(), &entry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process log"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Log saved",
		"id":           entry.ID, // Note: This will be 0 (async)
		"total_logged": totalCount,
	})
}

// GetLog
func (h *LogHandler) GetLog(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	entry, err := h.service.GetLog(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Log not found"})
		return
	}

	c.JSON(http.StatusOK, entry)
}

// SearchLogs handles GET /logs/search?q=keyword
func (h *LogHandler) SearchLogs(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	logs, err := h.service.SearchLogs(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count": len(logs),
		"data":  logs,
	})
}
