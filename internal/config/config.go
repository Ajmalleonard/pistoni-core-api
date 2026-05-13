package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	DatabaseURL        string
	SeedFromProduction bool
	SeedSourceURL      string
	JWTSecret          string
	GoogleClientID     string
	GoogleClientSecret string
	SMTPHost           string
	SMTPPort           string
	SMTPUser           string
	SMTPPass           string
	SMTPFromName       string
	SMTPFromEmail      string
	AppName            string
	SocketPath         string
	SocketPort         string
	RedisURL           string
	RedisHost          string
	RedisPort          int
	RedisPassword      string
	RedisDB            int
	RedisQueueStream   string
	RedisProfileTTL    int
}

func Load() Config {
	_ = godotenv.Load(".env")

	cfg := Config{
		Port:               getEnv("PORT", "4000"),
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		SeedFromProduction: getEnvBool("SEED_FROM_PRODUCTION", false),
		SeedSourceURL:      getEnv("SEED_SOURCE_URL", "https://core.api.pistoni.io"),
		JWTSecret:          os.Getenv("JWT_SECRET"),
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		SMTPHost:           getEnvAny([]string{"SMTP_HOST", "MAIL_HOST"}, ""),
		SMTPPort:           getEnvAny([]string{"SMTP_PORT", "MAIL_PORT"}, ""),
		SMTPUser:           getEnvAny([]string{"SMTP_USER", "MAIL_USER"}, ""),
		SMTPPass:           getEnvAny([]string{"SMTP_PASS", "MAIL_KEYS"}, ""),
		SMTPFromName:       getEnvAny([]string{"SMTP_FROM_NAME", "MAIL_NAME"}, "Pistoni"),
		SMTPFromEmail:      getEnvAny([]string{"SMTP_FROM_EMAIL", "MAIL_USER"}, ""),
		AppName:            getEnv("APP_NAME", "Pistoni"),
		SocketPath:         getEnv("SOCKET_PATH", "/socket.io"),
		SocketPort:         getEnv("SOCKET_PORT", ""),
		RedisURL:           getEnv("REDIS_URL", ""),
		RedisHost:          getEnv("REDIS_HOST", ""),
		RedisPort:          getEnvInt("REDIS_PORT", 6379),
		RedisPassword:      getEnv("REDIS_PASSWORD", ""),
		RedisDB:            getEnvInt("REDIS_DB", 0),
		RedisQueueStream:   getEnv("REDIS_QUEUE_STREAM", "pistoni:chat:signals"),
		RedisProfileTTL:    getEnvInt("REDIS_PROFILE_CACHE_TTL_SECONDS", 120),
	}

	if cfg.DatabaseURL == "" {
		log.Println("warning: DATABASE_URL missing")
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvAny(keys []string, fallback string) string {
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	switch strings.ToLower(v) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}
