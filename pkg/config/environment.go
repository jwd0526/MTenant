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

// IsProductionMode checks if running in production environment
func IsProductionMode() bool {
	env := strings.ToLower(os.Getenv("ENVIRONMENT"))
	return env == "prod" || env == "production"
}

// IsStagingMode checks if running in staging environment
func IsStagingMode() bool {
	env := strings.ToLower(os.Getenv("ENVIRONMENT"))
	return env == "stage" || env == "staging"
}

// GetEnvironment returns the current environment name
func GetEnvironment() string {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		return "development"
	}
	return strings.ToLower(env)
}

// GetDatabaseURL returns the database connection URL
func GetDatabaseURL() string {
	return os.Getenv("DATABASE_URL")
}

// GetJWTSecret returns the shared JWT secret for authentication
func GetJWTSecret() string {
	return os.Getenv("SHARED_JWT_SECRET")
}

// GetServicePort returns the port for the service with a fallback
func GetServicePort(defaultPort string) string {
	port := os.Getenv("PORT")
	if port == "" {
		return defaultPort
	}
	return port
}