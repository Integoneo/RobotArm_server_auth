package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddress           string
	StorageDriver         string
	MailerDriver          string
	JWTSecret             string
	TokenTTL              time.Duration
	UploadDir             string
	DatabaseURL           string
	PasswordResetCodeTTL  time.Duration
	SMTPHost              string
	SMTPPort              int
	SMTPUsername          string
	SMTPPassword          string
	SMTPFrom              string
	SupportRecipientEmail string
}

func Load() Config {
	return Config{
		HTTPAddress:           envOrDefault("HTTP_ADDRESS", ":8081"),
		StorageDriver:         envOrDefault("STORAGE_DRIVER", "auto"),
		MailerDriver:          envOrDefault("MAILER_DRIVER", "auto"),
		JWTSecret:             envOrDefault("JWT_SECRET", "change-me-in-production"),
		TokenTTL:              parseHoursOrDefault("TOKEN_TTL_HOURS", 24),
		UploadDir:             envOrDefault("UPLOAD_DIR", "uploads"),
		DatabaseURL:           envOrDefault("DATABASE_URL", ""),
		PasswordResetCodeTTL:  parseMinutesOrDefault("PASSWORD_RESET_CODE_TTL_MINUTES", 15),
		SMTPHost:              envOrDefault("SMTP_HOST", ""),
		SMTPPort:              parseIntOrDefault("SMTP_PORT", 587),
		SMTPUsername:          envOrDefault("SMTP_USERNAME", ""),
		SMTPPassword:          envOrDefault("SMTP_PASSWORD", ""),
		SMTPFrom:              envOrDefault("SMTP_FROM", ""),
		SupportRecipientEmail: envOrDefault("SUPPORT_RECIPIENT_EMAIL", ""),
	}
}

func (c Config) SMTPAddress() string {
	return fmt.Sprintf("%s:%d", c.SMTPHost, c.SMTPPort)
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func parseHoursOrDefault(key string, fallback int) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return time.Duration(fallback) * time.Hour
	}

	hours, err := strconv.Atoi(value)
	if err != nil || hours <= 0 {
		return time.Duration(fallback) * time.Hour
	}

	return time.Duration(hours) * time.Hour
}

func parseMinutesOrDefault(key string, fallback int) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return time.Duration(fallback) * time.Minute
	}

	minutes, err := strconv.Atoi(value)
	if err != nil || minutes <= 0 {
		return time.Duration(fallback) * time.Minute
	}

	return time.Duration(minutes) * time.Minute
}

func parseIntOrDefault(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}

	return parsed
}
