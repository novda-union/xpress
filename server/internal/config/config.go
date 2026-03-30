package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL      string
	TelegramBotToken string
	TelegramGateway  string
	JWTSecret        string
	AppURL           string
	Port             string
}

func Load() *Config {
	// Load .env from project root (one level up from server/)
	_ = godotenv.Load("../.env")
	_ = godotenv.Load(".env")
	return &Config{
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://xpressgo:xpressgo@localhost:5433/xpressgo?sslmode=disable"),
		TelegramBotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
		TelegramGateway:  getEnv("TELEGRAM_GATEWAY_TOKEN", ""),
		JWTSecret:        getEnv("JWT_SECRET", "dev-secret-change-me"),
		AppURL:           getEnv("APP_URL", "https://xpressgo.home.arpa"),
		Port:             getEnv("PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
