package domain

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// LogEntry (保持不變)
type LogEntry struct {
	gorm.Model
	ServiceName string    `json:"service_name" gorm:"index"`
	Level       string    `json:"level"`
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
}

// LogRepository (MySQL) - 增加 FindByID
type LogRepository interface {
	Create(ctx context.Context, entry *LogEntry) error
	GetByID(ctx context.Context, id uint) (*LogEntry, error) // [新增]
}

// LogCacheRepository (Redis) - 結合了原本的 Stats 和新的 Cache 功能
// 這裡我們把 StatsRepository 改名為更通用的 LogCacheRepository
type LogCacheRepository interface {
	IncrementLogCount(ctx context.Context) error
	GetLogCount(ctx context.Context) (int64, error)
	// [新增] 快取操作
	SetLog(ctx context.Context, entry *LogEntry) error
	GetLog(ctx context.Context, id uint) (*LogEntry, error)
}

// LogProducer (保持不變)
type LogProducer interface {
	SendLog(ctx context.Context, entry *LogEntry) error
	Close() error
}
