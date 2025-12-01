package service

import (
	"context"

	"github.com/Yupoer/logpulse/internal/domain"
)

// LogService handles the business logic for logs.
type LogService struct {
	logRepo   domain.LogRepository
	statsRepo domain.StatsRepository
}

// NewLogService injects dependencies.
func NewLogService(logRepo domain.LogRepository, statsRepo domain.StatsRepository) *LogService {
	return &LogService{
		logRepo:   logRepo,
		statsRepo: statsRepo,
	}
}

// CreateLog executes the business flow: Save Log -> Increment Counter -> Get Total Count.
func (s *LogService) CreateLog(ctx context.Context, entry *domain.LogEntry) (int64, error) {
	// 1. Persist log to MySQL
	if err := s.logRepo.Create(ctx, entry); err != nil {
		return 0, err
	}

	// 2. Increment Redis counter
	// Note: We ignore the error here for now to avoid failing the request if only cache is down.
	// In production, we should log this error.
	_ = s.statsRepo.IncrementLogCount(ctx)

	// 3. Retrieve current total count
	return s.statsRepo.GetLogCount(ctx)
}
