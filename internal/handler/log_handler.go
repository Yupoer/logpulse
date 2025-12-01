package handler

import (
	"net/http"
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

// CreateLog handles POST /logs requests.
func (h *LogHandler) CreateLog(c *gin.Context) {
	var entry domain.LogEntry

	// 1. Parse JSON
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	// 2. Set Default Timestamp
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	// 3. Call Service Layer
	// We pass c.Request.Context() to propagate cancellation/timeout signals.
	totalCount, err := h.service.CreateLog(c.Request.Context(), &entry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process log"})
		return
	}

	// 4. Return Response
	c.JSON(http.StatusCreated, gin.H{
		"message":      "Log saved",
		"id":           entry.ID,
		"total_logged": totalCount,
	})
}
