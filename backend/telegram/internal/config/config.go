package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	BotToken       string
	BackendBaseURL string
	PublicSiteURL  string
	BotLink        string
	SupportContact string
}

func Load() Config {
	return Config{
		BotToken:       GetEnv("BOT_TOKEN", ""),
		BackendBaseURL: GetEnv("BACKEND_BASE_URL", "http://localhost:8090"),
		PublicSiteURL:  GetEnv("PUBLIC_SITE_URL", "https://your.domain"),
		BotLink:        GetEnv("TELEGRAM_BOT_LINK", "https://t.me/your_telegram_bot"),
		SupportContact: GetEnv("SUPPORT_CONTACT", "@your_support_contact"),
	}
}

func GetEnv(key, defaultValue string) string {
	if err := godotenv.Load("/telegram/.env"); err != nil {
		log.Printf("Notice: Could not load .env file from /telegram/.env: %v", err)
	}
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Notice: Could not load .env file from /.env: %v", err)
	}
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
