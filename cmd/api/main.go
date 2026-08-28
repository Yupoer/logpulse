package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/Yupoer/logpulse/internal/config"
	"github.com/Yupoer/logpulse/internal/domain"
	"github.com/Yupoer/logpulse/internal/handler"
	"github.com/Yupoer/logpulse/internal/metrics"
	"github.com/Yupoer/logpulse/internal/middleware"
	"github.com/Yupoer/logpulse/internal/repository"
	"github.com/Yupoer/logpulse/internal/service"
)

const (
	dependencyRetryAttempts = 30
	dependencyRetryDelay    = 2 * time.Second
)

func retryDependency(name string, action func() error) error {
	var err error
	for attempt := 1; attempt <= dependencyRetryAttempts; attempt++ {
		if err = action(); err == nil {
			return nil
		}
		if attempt < dependencyRetryAttempts {
			log.Printf("%s is not ready (attempt %d/%d): %v", name, attempt, dependencyRetryAttempts, err)
			time.Sleep(dependencyRetryDelay)
		}
	}
	return fmt.Errorf("%s unavailable after %d attempts: %w", name, dependencyRetryAttempts, err)
}

func main() {
	// 1. Load Config
	cfg := config.LoadConfig()

	// 2. Infrastructure Setup
	// MySQL
	var db *gorm.DB
	if err := retryDependency("MySQL", func() error {
		var err error
		db, err = gorm.Open(mysql.Open(cfg.DBUrl), &gorm.Config{})
		return err
	}); err != nil {
		log.Fatalf("MySQL Connection Failed: %v", err)
	}
	// Warning: AutoMigrate should be avoided in production
	if err := retryDependency("MySQL migration", func() error {
		return db.AutoMigrate(&domain.LogEntry{})
	}); err != nil {
		log.Fatalf("Database migration failed: %v", err)
	}

	// Redis
	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	if err := retryDependency("Redis", func() error {
		return rdb.Ping(context.Background()).Err()
	}); err != nil {
		log.Fatalf("Redis Connection Failed: %v", err)
	}

	// Kafka Producer
	var producer domain.LogProducer
	if err := retryDependency("Kafka producer", func() error {
		var err error
		producer, err = repository.NewKafkaProducer(cfg.KafkaBrokers, cfg.KafkaTopic)
		return err
	}); err != nil {
		log.Fatalf("Failed to initialize Kafka Producer: %v", err)
	}
	defer func() { _ = producer.Close() }()

	// ES Repo init
	var esRepo domain.LogSearchRepository
	if err := retryDependency("Elasticsearch", func() error {
		var err error
		esRepo, err = repository.NewESLogRepository(cfg.ESAddress)
		return err
	}); err != nil {
		log.Fatalf("Failed to connect to Elasticsearch: %v", err)
	}

	statsRepo := repository.NewLogCacheRepository(rdb)
	logRepo := repository.NewLogRepository(db)
	appMetrics := metrics.New()

	logService := service.NewLogService(producer, logRepo, statsRepo, esRepo)
	logHandler := handler.NewLogHandler(logService)

	// Start Kafka Consumer Worker (Background)
	consumerWorker := repository.NewKafkaConsumer(logRepo, esRepo, appMetrics)
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	consumerDone := make(chan error, 1)

	go func() {
		log.Println("Starting Kafka Consumer Worker...")
		// "logpulse-group" is the Consumer Group ID.
		// If run multiple instances of this app, they will share the load.
		consumerDone <- consumerWorker.StartConsumerGroup(ctx, cfg.KafkaBrokers, cfg.KafkaTopic, "logpulse-group")
	}()

	// 4. Router Setup
	r := gin.Default()
	r.Use(appMetrics.Middleware())

	// Rate Limiter Middleware (Token Bucket via Redis Lua Script)
	rateLimiter := middleware.NewRateLimiter(rdb, cfg.RateLimit)
	r.Use(rateLimiter.Middleware())

	r.GET("/ping", func(c *gin.Context) { c.JSON(200, gin.H{"message": "pong"}) })
	r.GET("/metrics", gin.WrapH(appMetrics.Handler()))
	r.POST("/logs", logHandler.CreateLog)
	r.GET("/logs/:id", logHandler.GetLog)
	r.GET("/logs/search", logHandler.SearchLogs)

	// 5. Server Setup with Graceful Shutdown
	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: r,
	}

	serverDone := make(chan error, 1)
	go func() {
		log.Printf("Starting server on port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverDone <- err
			return
		}
		serverDone <- nil
	}()

	// 6. Graceful Shutdown Logic. The signal context also stops the consumer.
	select {
	case <-ctx.Done():
	case err := <-serverDone:
		if err != nil {
			log.Printf("Server Listen Error: %v", err)
		}
		cancel()
	}
	log.Println("Shutting down server...")

	serverShutdownCtx, serverShutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer serverShutdownCancel()

	if err := srv.Shutdown(serverShutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}
	consumerWaitCtx, consumerWaitCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer consumerWaitCancel()
	select {
	case err := <-consumerDone:
		if err != nil {
			log.Printf("Kafka consumer stopped with error: %v", err)
		}
	case <-consumerWaitCtx.Done():
		log.Println("Kafka consumer shutdown timed out")
	}

	log.Println("Server exiting")
}
