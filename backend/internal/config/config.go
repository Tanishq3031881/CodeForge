package config

import "os"

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	YjsURL      string
	InternalKey string
}

func LoadConfig() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://codeforge:codeforge_dev@localhost:5432/codeforge?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "dev-insecure-change-me"),
		YjsURL:      getEnv("YJS_URL", "http://127.0.0.1:1234"),
		InternalKey: getEnv("INTERNAL_KEY", "dev-internal-key"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

