package domain

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// LogEntry
type LogEntry struct {
	gorm.Model
	ServiceName string    `json:"service_name" gorm:"index"`
	Level       string    `json:"level"`
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
}

// LogRepository (MySQL)
type LogRepository interface {
	Create(ctx context.Context, entry *LogEntry) error
	GetByID(ctx context.Context, id uint) (*LogEntry, error)
}

// LogCacheRepository (Redis)
type LogCacheRepository interface {
	IncrementLogCount(ctx context.Context) error
	GetLogCount(ctx context.Context) (int64, error)
	// cache operations
	SetLog(ctx context.Context, entry *LogEntry) error
	GetLog(ctx context.Context, id uint) (*LogEntry, error)
}

// LogProducer
type LogProducer interface {
	SendLog(ctx context.Context, entry *LogEntry) error
	Close() error
}

// LogSearchRepository elasticsearch
type LogSearchRepository interface {
	BulkIndex(ctx context.Context, entries []*LogEntry) error
}
