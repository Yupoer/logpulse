package domain

import (
	"time"

	"gorm.io/gorm"
)

// LogEntry represents the log record in the database.
type LogEntry struct {
	gorm.Model
	ServiceName string    `json:"service_name" gorm:"index"` // Index added for search performance
	Level       string    `json:"level"`
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
}
