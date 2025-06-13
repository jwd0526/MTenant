package database

import (
	"context"
	"log"
	"os"
	"testing"
)

func TestConnectionPool(t *testing.T) {
	// Skip if no DATABASE_URL
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	ctx := context.Background()

	// Load configuration
	config, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Create pool
	pool, err := NewPool(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}
	defer pool.Close()

	// Test health check
	health := pool.HealthCheck(ctx)
	if !health.Healthy {
		t.Fatalf("Health check failed: %s", health.Error)
	}

	log.Printf("Health check passed in %v", health.ResponseTime)
	log.Printf("Pool stats: %+v", health.Stats)
}