package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	YjsURL      string
	InternalKey string

	// Sandbox (code execution).
	SandboxImage    string
	SandboxPoolSize int
	SandboxTimeout  time.Duration
}

func LoadConfig() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://codeforge:codeforge_dev@localhost:5432/codeforge?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "dev-insecure-change-me"),
		YjsURL:      getEnv("YJS_URL", "http://127.0.0.1:1234"),
		InternalKey: getEnv("INTERNAL_KEY", "dev-internal-key"),

		SandboxImage:    getEnv("SANDBOX_IMAGE", "codeforge-runner-python"),
		SandboxPoolSize: getEnvInt("SANDBOX_POOL_SIZE", 3),
		SandboxTimeout:  time.Duration(getEnvInt("SANDBOX_TIMEOUT_SECONDS", 5)) * time.Second,
	}
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
