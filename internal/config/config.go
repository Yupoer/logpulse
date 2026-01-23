package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type RateLimitConfig struct {
	Enabled  bool
	Capacity int64   // Max burst requests
	Rate     float64 // Tokens per second refill rate
}

type Config struct {
	ServerPort   string
	DBUrl        string
	RedisAddr    string
	KafkaBrokers []string
	KafkaTopic   string
	ESAddress    string
	RateLimit    RateLimitConfig
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, relying on system environment variables")
	}

	// Helper to handle comma-separated brokers from env
	brokers := os.Getenv("KAFKA_BROKERS")
	brokerList := []string{}
	if brokers != "" {
		brokerList = strings.Split(brokers, ",")
	}

	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	dsn := dbUser + ":" + dbPass + "@tcp(" + dbHost + ":" + dbPort + ")/" + dbName + "?charset=utf8mb4&parseTime=True&loc=Local"

	// Rate Limit Config
	rateLimitEnabled := os.Getenv("RATE_LIMIT_ENABLED") == "true"
	rateLimitCapacity, _ := strconv.ParseInt(os.Getenv("RATE_LIMIT_CAPACITY"), 10, 64)
	if rateLimitCapacity == 0 {
		rateLimitCapacity = 100 // Default: 100 burst
	}
	rateLimitRate, _ := strconv.ParseFloat(os.Getenv("RATE_LIMIT_RATE"), 64)
	if rateLimitRate == 0 {
		rateLimitRate = 50 // Default: 50 tokens/sec
	}

	return &Config{
		ServerPort:   os.Getenv("SERVER_PORT"),
		DBUrl:        dsn,
		RedisAddr:    os.Getenv("REDIS_ADDR"),
		KafkaBrokers: brokerList,
		KafkaTopic:   os.Getenv("KAFKA_TOPIC"),
		ESAddress:    os.Getenv("ELASTICSEARCH_ADDRESS"),
		RateLimit: RateLimitConfig{
			Enabled:  rateLimitEnabled,
			Capacity: rateLimitCapacity,
			Rate:     rateLimitRate,
		},
	}
}
