package main

import (
	"context"
	"log"
	"os"
	"time"

	"crm-platform/pkg/database"
	"crm-platform/tenant-service/internal/errors"
	"crm-platform/tenant-service/internal/handlers"
	"crm-platform/tenant-service/internal/services"

	"github.com/gin-gonic/gin"
)

// =============================================================================
// CONFIGURATION
// =============================================================================

// Load server port from environment with fallback
func getServerPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		return "8081"
	}
	return port
}

// =============================================================================
// SETUP FUNCTIONS
// =============================================================================

// Initialize database connection with retry logic
func setupDatabase() (*database.Pool, error) {
	// Load config from environment
	config, err := database.LoadConfigFromEnv()
	if err != nil {
		return nil, errors.ErrDatabase("failed to load database config: " + err.Error())
	}

	// Create connection pool with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := database.NewPool(ctx, config)
	if err != nil {
		return nil, errors.ErrDatabase("failed to create connection pool: " + err.Error())
	}

	// Test connection with ping
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, errors.ErrDatabase("failed to ping database: " + err.Error())
	}

	log.Println("Database connection established successfully")
	return pool, nil
}

// Initialize all handlers with database dependencies
func setupHandlers(pool *database.Pool) (*handlers.TenantHandler, *handlers.HealthHandler) {
	// Create service layer
	tenantService := services.NewTenantService(pool)

	// Create handler instances
	tenantHandler := handlers.NewTenantHandler(tenantService)
	healthHandler := handlers.NewHealthHandler(pool)

	log.Println("Handlers initialized successfully")
	return tenantHandler, healthHandler
}

// Register all API routes
func setupRoutes(router *gin.Engine, tenantHandler *handlers.TenantHandler, healthHandler *handlers.HealthHandler) {
	// Register system endpoints (no auth required for internal service)
	router.GET("/health", healthHandler.HealthCheck) // GET /health

	// Create internal API group
	internal := router.Group("/internal")

	// Register tenant endpoints
	tenants := internal.Group("/tenants")
	{
		tenants.POST("", tenantHandler.CreateTenant)                          // POST /internal/tenants
		tenants.GET("", tenantHandler.ListTenants)                            // GET /internal/tenants
		tenants.GET("/:id", tenantHandler.GetTenant)                          // GET /internal/tenants/:id
		tenants.GET("/subdomain/:subdomain", tenantHandler.GetTenantBySubdomain) // GET /internal/tenants/subdomain/:subdomain
		tenants.PUT("/:id", tenantHandler.UpdateTenant)                       // PUT /internal/tenants/:id
		tenants.GET("/:id/health", tenantHandler.GetTenantHealth)             // GET /internal/tenants/:id/health
	}

	// Register test endpoints (for development/testing)
	testTenants := internal.Group("/test-tenants")
	{
		testTenants.POST("", tenantHandler.CreateTestTenants)     // POST /internal/test-tenants
		testTenants.DELETE("", tenantHandler.DeleteTestTenants)   // DELETE /internal/test-tenants
	}

	log.Println("Routes registered successfully")
}

// =============================================================================
// MAIN APPLICATION
// =============================================================================

func main() {
	log.Println("Starting Tenant Service...")

	// Initialize Gin router
	router := gin.Default()

	// Setup database connection
	pool, err := setupDatabase()
	if err != nil {
		log.Fatal(err.Error())
	}
	defer pool.Close()

	// Setup handlers
	tenantHandler, healthHandler := setupHandlers(pool)

	// Setup routes
	setupRoutes(router, tenantHandler, healthHandler)

	// Get server port from environment
	port := getServerPort()

	log.Printf("Tenant Service running on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal(errors.ErrHandler("failed to start server: " + err.Error()).Error())
	}
}
