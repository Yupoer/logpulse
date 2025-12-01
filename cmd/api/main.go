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
	// Tech Debt: AutoMigrate should be avoided in production
	db.AutoMigrate(&domain.LogEntry{})

	// Redis
	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Redis Connection Failed: %v", err)
	}

	// 3. Dependency Injection (Wiring)
	// Repo -> Service -> Handler
	logRepo := repository.NewLogRepository(db)
	statsRepo := repository.NewStatsRepository(rdb)

	logService := service.NewLogService(logRepo, statsRepo)
	logHandler := handler.NewLogHandler(logService)

	// 4. Router Setup
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) { c.JSON(200, gin.H{"message": "pong"}) })
	r.POST("/logs", logHandler.CreateLog) // Register the new handler

	// 5. Server Setup with Graceful Shutdown
	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: r,
	}

	// Run server in a separate goroutine so it doesn't block the main thread
	go func() {
		log.Printf("Starting server on port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server Listen Error: %v", err)
		}
	}()

	// 6. Graceful Shutdown Logic
	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // Block here until signal is received
	log.Println("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
