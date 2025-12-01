package domain

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// LogEntry represents the core data model.
type LogEntry struct {
	gorm.Model
	ServiceName string    `json:"service_name" gorm:"index"`
	Level       string    `json:"level"`
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
}

// LogRepository defines the interface for log persistence.
// This allows us to swap MySQL with PostgreSQL or MockDB without changing the Service layer.
type LogRepository interface {
	Create(ctx context.Context, entry *LogEntry) error
}

// StatsRepository defines the interface for statistics (Redis).
type StatsRepository interface {
	IncrementLogCount(ctx context.Context) error
	GetLogCount(ctx context.Context) (int64, error)
}
