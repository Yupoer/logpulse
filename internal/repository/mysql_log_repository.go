package repository

import (
	"context"

	"github.com/Yupoer/logpulse/internal/domain"
	"gorm.io/gorm"
)

type mysqlLogRepository struct {
	db *gorm.DB
}

// NewLogRepository is the factory function to inject DB dependency.
func NewLogRepository(db *gorm.DB) domain.LogRepository {
	return &mysqlLogRepository{db: db}
}

func (r *mysqlLogRepository) Create(ctx context.Context, entry *domain.LogEntry) error {
	// GORM supports Context to handle timeouts and cancellation
	return r.db.WithContext(ctx).Create(entry).Error
}
