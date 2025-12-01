package service

import (
	"context"
	"log"

	"github.com/Yupoer/logpulse/internal/domain"
)

type LogService struct {
	producer  domain.LogProducer
	logRepo   domain.LogRepository      // [Re-added] Service needs to read from DB now
	cacheRepo domain.LogCacheRepository // Renamed from statsRepo
}

// Update Constructor
func NewLogService(producer domain.LogProducer, logRepo domain.LogRepository, cacheRepo domain.LogCacheRepository) *LogService {
	return &LogService{
		producer:  producer,
		logRepo:   logRepo,
		cacheRepo: cacheRepo,
	}
}

// CreateLog
func (s *LogService) CreateLog(ctx context.Context, entry *domain.LogEntry) (int64, error) {
	if err := s.producer.SendLog(ctx, entry); err != nil {
		return 0, err
	}
	_ = s.cacheRepo.IncrementLogCount(ctx)
	return s.cacheRepo.GetLogCount(ctx)
}

// [NEW] GetLog implements Cache-Aside Pattern
func (s *LogService) GetLog(ctx context.Context, id uint) (*domain.LogEntry, error) {
	// 1. Check Redis Cache (Fast Path)
	cachedEntry, err := s.cacheRepo.GetLog(ctx, id)
	if err != nil {
		log.Printf("[Warn] Cache error: %v", err)
		// Don't fail the request if cache is down, just proceed to DB
	}
	if cachedEntry != nil {
		log.Println("[Cache] Hit")
		return cachedEntry, nil
	}

	// 2. Cache Miss -> Check MySQL (Slow Path)
	log.Println("[Cache] Miss, querying DB...")
	dbEntry, err := s.logRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 3. Write back to Cache (Async or Sync)
	// We do it synchronously here for simplicity
	if err := s.cacheRepo.SetLog(ctx, dbEntry); err != nil {
		log.Printf("[Warn] Failed to set cache: %v", err)
	}

	return dbEntry, nil
}
