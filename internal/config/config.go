package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv     string
	ServerPort string
	DBUrl      string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	cfg := &Config{
		AppEnv:     getEnv("APP_ENV", "local"),
		ServerPort: getEnv("SERVER_PORT", ":8080"),
		DBUrl:      getEnv("DB_URL", "postgresql://postgres:postgres@localhost:5432/appointment_db"),
	}

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
