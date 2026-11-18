package config

import "os"

type Config struct {
	Port          string
	DatabaseURL   string
	RedisAddr     string
	RedisPassword string
	ChromePath    string
	CORSOrigins   string
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func Load() Config {
	return Config{
		Port:          getenv("PORT", "8080"),
		DatabaseURL:   getenv("DATABASE_URL", "postgres://supermarket:supermarket@localhost:5432/supermarket?sslmode=disable"),
		RedisAddr:     getenv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getenv("REDIS_PASSWORD", ""),
		ChromePath:    getenv("CHROME_PATH", ""),
		CORSOrigins:   getenv("CORS_ORIGINS", "http://localhost:3000,http://localhost:3001"),
	}
}
