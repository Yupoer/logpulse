package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
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
	"github.com/Yupoer/logpulse/internal/repository"
	"github.com/Yupoer/logpulse/internal/service"
)

func main() {
	// 1. Load Config
	cfg := config.LoadConfig()

	// 2. Infrastructure Setup
	// MySQL
	db, err := gorm.Open(mysql.Open(cfg.DBUrl), &gorm.Config{})
	if err != nil {
		log.Fatalf("MySQL Connection Failed: %v", err)
	}
	// Warning: AutoMigrate should be avoided in production
	db.AutoMigrate(&domain.LogEntry{})

	// Redis
	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Redis Connection Failed: %v", err)
	}

	// Kafka Producer
	producer, err := repository.NewKafkaProducer(cfg.KafkaBrokers, cfg.KafkaTopic)
	if err != nil {
		log.Fatalf("Failed to initialize Kafka Producer: %v", err)
	}
	defer producer.Close()

	// --- Elasticsearch Repository ---
	esRepo, err := repository.NewESLogRepository(cfg.ESAddress)
	if err != nil {
		// Log Warning instead of Fatal, allow app to start even if ES is temporarily down (optional strategy)
		// But for now, let's fail fast to ensure config is correct
		log.Fatalf("Failed to connect to Elasticsearch: %v", err)
	}

	// 3. Dependency Injection (Wiring)
	cacheRepo := repository.NewLogCacheRepository(rdb) // Renamed constructor
	logRepo := repository.NewLogRepository(db)         // We need this for Service now

	// Updated Service Injection
	logService := service.NewLogService(producer, logRepo, cacheRepo)
	logHandler := handler.NewLogHandler(logService)

	// --- Start Kafka Consumer Worker (Background Job) ---
	// We run this in a separate goroutine so it doesn't block the HTTP server.
	consumerWorker := repository.NewKafkaConsumer(logRepo, esRepo)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure cleanup on exit

	go func() {
		log.Println("Starting Kafka Consumer Worker...")
		// "logpulse-group" is the Consumer Group ID.
		// If you run multiple instances of this app, they will share the load.
		consumerWorker.StartConsumerGroup(ctx, cfg.KafkaBrokers, cfg.KafkaTopic, "logpulse-group")
	}()

	// 4. Router Setup
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) { c.JSON(200, gin.H{"message": "pong"}) })
	r.POST("/logs", logHandler.CreateLog)
	r.GET("/logs/:id", logHandler.GetLog) // [NEW] Register Route

	// 5. Server Setup with Graceful Shutdown
	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: r,
	}

	go func() {
		log.Printf("Starting server on port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server Listen Error: %v", err)
		}
	}()

	// 6. Graceful Shutdown Logic
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
