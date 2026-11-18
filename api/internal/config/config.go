package config

import (
	"errors"
	"os"
	"strings"
)

type Config struct {
	Port          string
	DatabaseURL   string
	RedisAddr     string
	RedisPassword string
	ChromePath    string
	CORSOrigins   string
	Seed          bool
	AdminEmail    string
	AdminPassword string

	// Supabase Auth
	SupabaseURL            string
	SupabaseServiceRoleKey string

	// Supabase Storage (S3-compatible)
	S3Endpoint        string
	S3Region          string
	S3Bucket          string
	S3AccessKeyID     string
	S3SecretAccessKey string
	PublicFileBaseURL string
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
		DatabaseURL:   getenv("DATABASE_URL", ""),
		RedisAddr:     getenv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getenv("REDIS_PASSWORD", ""),
		ChromePath:    getenv("CHROME_PATH", ""),
		CORSOrigins:   getenv("CORS_ORIGINS", "http://localhost:3000,http://localhost:3001"),
		Seed:          getenv("SEED", "false") == "true",
		AdminEmail:    strings.TrimSpace(getenv("ADMIN_EMAIL", "")),
		AdminPassword: getenv("ADMIN_PASSWORD", ""),

		SupabaseURL:            getenv("SUPABASE_URL", ""),
		SupabaseServiceRoleKey: getenv("SUPABASE_SERVICE_ROLE_KEY", ""),

		S3Endpoint:        getenv("S3_ENDPOINT", ""),
		S3Region:          getenv("S3_REGION", ""),
		S3Bucket:          getenv("S3_BUCKET", "offer-media"),
		S3AccessKeyID:     getenv("S3_ACCESS_KEY_ID", ""),
		S3SecretAccessKey: getenv("S3_SECRET_ACCESS_KEY", ""),
		PublicFileBaseURL: getenv("PUBLIC_FILE_BASE_URL", ""),
	}
}

// Validate stops the app at startup when something it cannot run without is missing.
// S3 settings are checked later, by the features that use them.
func (c Config) Validate() error {
	if c.DatabaseURL == "" {
		return errors.New("DATABASE_URL is not set: use your Supabase session pooler connection string")
	}
	if c.SupabaseURL == "" {
		return errors.New("SUPABASE_URL is not set: the API needs it to verify logins")
	}
	if (c.AdminEmail == "") != (c.AdminPassword == "") {
		return errors.New("set both ADMIN_EMAIL and ADMIN_PASSWORD, or neither")
	}
	if c.AdminPassword != "" && (len(c.AdminPassword) < 8 || len(c.AdminPassword) > 72) {
		return errors.New("ADMIN_PASSWORD must be 8 to 72 characters")
	}
	return nil
}
