package service

import (
	"context"
	"log"

	"github.com/Yupoer/logpulse/internal/domain"
)

type LogService struct {
	producer  domain.LogProducer
	logRepo   domain.LogRepository
	cacheRepo domain.LogCacheRepository
	esRepo    domain.LogSearchRepository
}

func NewLogService(producer domain.LogProducer, logRepo domain.LogRepository, cacheRepo domain.LogCacheRepository, esRepo domain.LogSearchRepository) *LogService {
	return &LogService{
		producer:  producer,
		logRepo:   logRepo,
		cacheRepo: cacheRepo,
		esRepo:    esRepo,
	}
}

func (s *LogService) CreateLog(ctx context.Context, entry *domain.LogEntry) (int64, error) {
	if err := s.producer.SendLog(ctx, entry); err != nil {
		return 0, err
	}
	_ = s.cacheRepo.IncrementLogCount(ctx)
	return s.cacheRepo.GetLogCount(ctx)
}

func (s *LogService) GetLog(ctx context.Context, id uint) (*domain.LogEntry, error) {
	// 1. Check Redis Cache
	cachedEntry, err := s.cacheRepo.GetLog(ctx, id)
	if err != nil {
		log.Printf("[Warn] Cache error: %v", err)
	}
	if cachedEntry != nil {
		log.Println("[Cache] Hit")
		return cachedEntry, nil
	}

	// 2. Cache Miss -> Check MySQL
	log.Println("[Cache] Miss, querying DB...")
	dbEntry, err := s.logRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 3. Write back to Cache
	if err := s.cacheRepo.SetLog(ctx, dbEntry); err != nil {
		log.Printf("[Warn] Failed to set cache: %v", err)
	}

	return dbEntry, nil
}

func (s *LogService) SearchLogs(ctx context.Context, query string) ([]*domain.LogEntry, error) {
	return s.esRepo.Search(ctx, query)
}
