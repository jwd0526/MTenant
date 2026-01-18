package config

import (
	"os"
	"strings"
)

// IsDevelopmentMode checks if running in development environment
func IsDevelopmentMode() bool {
	env := strings.ToLower(os.Getenv("ENVIRONMENT"))		
	return env == "dev" || env == "development" || env == ""
}

func GetDatabaseURL() string {
	return os.Getenv("DATABASE_URL")
}

func GetJWTSecret() string {
	return os.Getenv("SHARED_JWT_SECRET")
}