package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/Yupoer/logpulse/internal/config"
	"github.com/Yupoer/logpulse/internal/domain"
)

func main() {
	// 1. Load configuration
	cfg := config.LoadConfig()

	// 2. Database connection (MySQL)
	db, err := gorm.Open(mysql.Open(cfg.DBUrl), &gorm.Config{})
	if err != nil {
		log.Fatalf("MySQL connection failed: %v", err)
	}

	// AutoMigrate creates tables based on the struct.
	// Note: Avoid using AutoMigrate in production; use versioned migration tools instead.
	if err := db.AutoMigrate(&domain.LogEntry{}); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	// 3. Cache connection (Redis)
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Redis connection failed: %v", err)
	}

	// 4. Initialize Router
	r := gin.Default()

	r.POST("/logs", func(c *gin.Context) {
		var entry domain.LogEntry

		// Bind JSON payload to struct
		if err := c.ShouldBindJSON(&entry); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
			return
		}

		// Set timestamp if missing
		if entry.Timestamp.IsZero() {
			entry.Timestamp = time.Now()
		}

		// Write to MySQL
		if err := db.Create(&entry).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to persist log"})
			return
		}

		// Write to Redis (Increment counter)
		ctx := context.Background()
		rdb.Incr(ctx, "stats:log_count")

		// Retrieve current count for response
		count, _ := rdb.Get(ctx, "stats:log_count").Result()

		c.JSON(http.StatusCreated, gin.H{
			"message":      "Log saved",
			"id":           entry.ID,
			"total_logged": count,
		})
	})

	// Start server
	r.Run(":" + cfg.ServerPort)
}
