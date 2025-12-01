package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort   string
	DBUrl        string
	RedisAddr    string
	KafkaBrokers []string // Changed to slice for multiple brokers
	KafkaTopic   string
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

	return &Config{
		ServerPort:   os.Getenv("SERVER_PORT"),
		DBUrl:        dsn,
		RedisAddr:    os.Getenv("REDIS_ADDR"),
		KafkaBrokers: brokerList,
		KafkaTopic:   os.Getenv("KAFKA_TOPIC"),
	}
}
