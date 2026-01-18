package config

import (
	"flag"
	"os"
)

// Config holds all application configuration
type Config struct {
	Port      string
	DBPath    string
	JWTSecret string
	AIAPIKey  string
}

// Load loads configuration from environment variables and command-line flags
func Load() *Config {
	// Use PORT environment variable if available (for Render, Railway, etc.)
	portEnv := getEnv("PORT", "8080")

	// Command-line flags
	dbPath := flag.String("db", "./algovault.db", "Path to SQLite database file (used only if DATABASE_URL is not set)")
	port := flag.String("port", portEnv, "Server port")
	jwtSecret := flag.String("jwt-secret", getEnv("JWT_SECRET", "your-secret-key-change-in-production"), "JWT secret key")
	aiAPIKey := flag.String("ai-api-key", getEnv("AI_API_KEY", "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"), "OpenRouter API key")
	flag.Parse()

	return &Config{
		Port:      *port,
		DBPath:    *dbPath,
		JWTSecret: *jwtSecret,
		AIAPIKey:  *aiAPIKey,
	}
}

// getEnv gets environment variable or returns default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
