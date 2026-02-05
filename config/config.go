package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort         string
	ElasticsearchURL   string
	ElasticsearchIndex string
}

func LoadConfig() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return &Config{
		ServerPort:         getEnv("SERVER_PORT", "8080"),
		ElasticsearchURL:   getEnv("ELASTICSEARCH_URL", "http://localhost:9200"),
		ElasticsearchIndex: getEnv("ELASTICSEARCH_INDEX", "products"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
