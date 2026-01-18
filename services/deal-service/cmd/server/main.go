package main

import (
	"context"
	"log"
	"os"
	"time"

	"crm-platform/pkg/database"
	"crm-platform/deal-service/internal/errors"
	"crm-platform/deal-service/internal/handlers"
	"crm-platform/pkg/middleware"

	"github.com/gin-gonic/gin"
)

// =============================================================================
// CONFIGURATION
// =============================================================================

// Load server port from environment with fallback
func getServerPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		return "8080"
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
func setupHandlers(pool *database.Pool) (*handlers.DealHandler, *handlers.SystemHandler) {
	// Create handler instances
	dealHandler := handlers.NewDealHandler(pool)
	systemHandler := handlers.NewSystemHandler(pool)
	
	log.Println("Handlers initialized successfully")
	return dealHandler, systemHandler
}

// Setup middleware stack in correct order
func setupMiddleware(router *gin.Engine) {
	// Add middleware in critical order
	// Auth middleware first - validates JWT and sets user context
	router.Use(middleware.AuthMiddleware())
	
	// Tenant middleware second - converts tenant ID to request context
	router.Use(middleware.TenantMiddleware())
	
	log.Println("Middleware configured successfully")
}

// Register all API routes
func setupRoutes(router *gin.Engine, dealHandler *handlers.DealHandler, systemHandler *handlers.SystemHandler) {
	// Register system endpoints (no auth required)
	router.GET("/health", systemHandler.HealthCheck)  // GET /health
	
	// Create API version groups
	v1 := router.Group("/api/v1")
	
	// Register deal endpoints
	deals := v1.Group("/deals")
	{
		deals.POST("", dealHandler.CreateDeal)           		// POST /api/v1/deals
		deals.GET("", dealHandler.ListDeals)             		// GET /api/v1/deals
		deals.GET("/pipeline", dealHandler.GetPipelineView) 	// GET /api/v1/deals/pipeline
		deals.GET("/owner/:id", dealHandler.GetDealsByOwner) 	// GET /api/v1/deals/owner/:id
		deals.GET("/:id", dealHandler.GetDeal)           		// GET /api/v1/deals/:id
		deals.PUT("/:id", dealHandler.UpdateDeal)        		// PUT /api/v1/deals/:id
		deals.PUT("/:id/close", dealHandler.CloseDeal)   		// PUT /api/v1/deals/:id/close
		deals.DELETE("/:id", dealHandler.DeleteDeal)     		// DELETE /api/v1/deals/:id
	}
	
	log.Println("Routes registered successfully")
}

// =============================================================================
// MAIN APPLICATION
// =============================================================================

func main() {
	log.Println("Starting Deal Service...")
	
	// Initialize Gin router
	router := gin.Default()
	
	// Setup database connection
	pool, err := setupDatabase()
	if err != nil {
		log.Fatal(err.Error())
	}
	defer pool.Close()
	
	// Setup middleware stack
	setupMiddleware(router)
	
	// Setup handlers
	dealHandler, systemHandler := setupHandlers(pool)
	
	// Setup routes
	setupRoutes(router, dealHandler, systemHandler)
	
	// Get server port from environment
	port := getServerPort()
	
	log.Printf("Deal Service running on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal(errors.ErrHandler("failed to start server: " + err.Error()).Error())
	}
}