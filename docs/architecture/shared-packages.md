# Shared Packages Architecture

**Last Updated:** 2025-10-08\
*Common functionality packages for microservices*

Documentation for the shared packages (`pkg/`) that provide common functionality across all microservices.

## Overview

The `pkg/` module contains reusable packages that standardize common functionality across all services in the MTenant CRM platform. This promotes code reuse, consistency, and maintainability.

## Package Structure

```
pkg/
├── go.mod                    # Shared package module
├── database/                 # Database connection and management
│   ├── config.go            # Configuration parsing and validation
│   ├── pool.go              # Connection pool management
│   ├── health.go            # Health checks and monitoring
│   ├── metrics.go           # Database metrics collection
│   └── ex_test.go           # Integration tests
├── middleware/              # HTTP middleware (planned)
└── utils/                   # Common utilities (planned)
```

## Database Package (`pkg/database`)

### Purpose

The database package provides standardized database connection management, health monitoring, and metrics collection for all services.

### Key Features

- **Environment-based configuration** from `DATABASE_URL`
- **Connection pooling** with configurable limits and timeouts
- **Automatic retry logic** for connection failures
- **Health monitoring** with detailed statistics
- **Metrics collection** for observability
- **Multi-tenant aware** connection management

### Configuration (`config.go`)

**Configuration Structure:**
```go
type Config struct {
    Host            string        // Database host
    Port            int           // Database port
    Database        string        // Database name
    Username        string        // Database user
    Password        string        // Database password
    MaxConns        int32         // Maximum connections (default: 20)
    MinConns        int32         // Minimum connections (default: 5)
    MaxConnLifetime time.Duration // Connection lifetime (default: 60m)
    MaxConnIdleTime time.Duration // Idle timeout (default: 5m)
    ConnectTimeout  time.Duration // Connection timeout (default: 30s)
    QueryTimeout    time.Duration // Query timeout (default: 30s)
    MaxRetries      int           // Retry attempts (default: 5)
    RetryInterval   time.Duration // Retry delay (default: 10s)
    SSLMode         string        // SSL mode (auto-detected)
}
```

**Environment Loading:**
```go
// Load configuration from DATABASE_URL environment variable
config, err := database.LoadConfigFromEnv()
if err != nil {
    log.Fatal("Failed to load database config:", err)
}

// Example DATABASE_URL
// postgresql://admin:admin@localhost:5433/crm-platform?sslmode=disable
```

**Automatic SSL Mode Detection:**
- **Development**: `sslmode=disable` (when `ENVIRONMENT=dev`)
- **Production**: `sslmode=prefer` (default for other environments)
- **Override**: Specify `sslmode` in DATABASE_URL query parameters

### Connection Pool (`pool.go`)

**Pool Creation:**
```go
// Create new database pool with configuration
ctx := context.Background()
pool, err := database.NewPool(ctx, config)
if err != nil {
    log.Fatal("Failed to create database pool:", err)
}
defer pool.Close()
```

**Pool Features:**
- **Embedded pgxpool.Pool** - Full pgx functionality
- **Automatic retry logic** - Configurable retry attempts with exponential backoff
- **Connection validation** - Ping test on pool creation
- **Graceful shutdown** - Proper connection cleanup

**Pool Statistics:**
```go
// Get connection pool statistics
stats := pool.Stats()
fmt.Printf("Active connections: %d/%d\n", stats.AcquiredConns(), stats.MaxConns())
```

### Health Monitoring (`health.go`)

**Health Check Structure:**
```go
type HealthStatus struct {
    Healthy      bool          `json:"healthy"`
    ResponseTime time.Duration `json:"response_time"`
    Error        string        `json:"error,omitempty"`
    Stats        *PoolStats    `json:"stats"`
}
```

**Health Check Usage:**
```go
// Perform comprehensive health check
health := pool.HealthCheck(ctx)
if !health.Healthy {
    log.Printf("Database unhealthy: %s", health.Error)
}

// Simple health check
if pool.IsHealthy(ctx) {
    log.Println("Database is healthy")
}
```

**Health Check Process:**
1. **Ping test** - Basic connectivity validation
2. **Query test** - Execute `SELECT 1` to verify functionality
3. **Pool statistics** - Connection pool health metrics
4. **Timeout handling** - 5-second timeout for health checks

### Metrics Collection (`metrics.go`)

**Metrics Structure:**
```go
type Metrics struct {
    // Connection metrics
    TotalConnections   int64
    FailedConnections  int64
    ActiveConnections  int64
    
    // Query metrics
    TotalQueries       int64
    FailedQueries      int64
    QueryDuration      int64 // nanoseconds
    
    // Health check metrics
    HealthChecks       int64
    FailedHealthChecks int64
    LastHealthCheck    int64 // unix timestamp
}
```

**Thread-Safe Operations:**
```go
// All metrics operations are atomic and thread-safe
metrics := pool.metrics

// Increment counters
metrics.IncrementConnections()
metrics.IncrementQueries()
metrics.IncrementHealthChecks()

// Update durations
metrics.AddQueryDuration(queryDuration)

// Get current snapshot
snapshot := metrics.GetMetrics()
```

## Service Integration

### Import and Usage

**Service Import:**
```go
import (
    "crm-platform/pkg/database"
)
```

**Standard Service Setup:**
```go
func main() {
    ctx := context.Background()
    
    // Load database configuration
    dbConfig, err := database.LoadConfigFromEnv()
    if err != nil {
        log.Fatal("Database config error:", err)
    }
    
    // Create connection pool
    dbPool, err := database.NewPool(ctx, dbConfig)
    if err != nil {
        log.Fatal("Database connection error:", err)
    }
    defer dbPool.Close()
    
    // Create SQLC queries instance
    queries := db.New(dbPool)
    
    // Set up HTTP handlers with database access
    handler := NewHandler(queries, dbPool)
    
    // Start server
    log.Println("Server starting...")
    log.Fatal(http.ListenAndServe(":8080", handler))
}
```

### Health Check Endpoint

**Standard Health Check Handler:**
```go
func (h *Handler) HealthCheck(c *gin.Context) {
    health := h.dbPool.HealthCheck(c.Request.Context())
    
    status := 200
    if !health.Healthy {
        status = 503
    }
    
    c.JSON(status, gin.H{
        "status":        health.Healthy,
        "response_time": health.ResponseTime.String(),
        "database":      health.Healthy,
        "stats":         health.Stats,
        "error":         health.Error,
    })
}
```

### Tenant Context Management

**Tenant Schema Setting:**
```go
func (h *Handler) setTenantContext(ctx context.Context, schemaName string) error {
    _, err := h.dbPool.Exec(ctx, "SET search_path = $1", schemaName)
    return err
}

// Usage in middleware
func TenantMiddleware(dbPool *database.Pool) gin.HandlerFunc {
    return func(c *gin.Context) {
        tenantID := c.GetString("tenant_id")
        schemaName := fmt.Sprintf("tenant_%s", tenantID)
        
        // Set tenant context for this request
        err := setTenantContext(c.Request.Context(), schemaName)
        if err != nil {
            c.JSON(500, gin.H{"error": "Failed to set tenant context"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

## Go Workspace Integration

### Module Dependencies

The shared package is included in the Go workspace:

```go
// go.work
go 1.24.3

use (
    ./pkg                    // Shared packages
    ./services/auth-service
    ./services/tenant-service
    ./services/contact-service
    ./services/deal-service
    ./services/communication-service
)
```

**Service Module Dependencies:**
```go
// services/auth-service/go.mod
require (
    github.com/jackc/pgx/v5 v5.7.5
    // pkg/database is automatically available via workspace
)
```

### Build System Integration

**Current Status**: The shared packages are automatically built and tested as part of the Go workspace. No special Makefile targets are needed.

**Integration via Go Workspace:**
```makefile
# Shared packages are automatically included via go.work
# No special build targets needed - Go workspace handles dependencies

# Build all services (includes shared packages automatically)
build: $(BUILD_TARGETS)

# Test all services (includes shared packages automatically)  
test:
	for service in $(SERVICES); do \
		cd $(PWD); \
		echo "Testing $$service..."; \
		cd services/$$service && go test -v ./...; \
	done

# Test shared packages independently (manual)
# cd pkg && go test -v ./...
```

## Testing

### Integration Tests

**Database Integration Test:**
```go
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
```

**Running Tests:**
```bash
# Test with database connection
export DATABASE_URL="postgresql://admin:admin@localhost:5433/crm-platform?sslmode=disable"
cd pkg && go test -v ./database

# Test without database (skips integration tests)
cd pkg && go test -v ./database
```

## Performance Considerations

### Connection Pool Tuning

**Production Settings:**
```go
config := &database.Config{
    MaxConns:        50,          // Scale with load
    MinConns:        10,          // Maintain base connections
    MaxConnLifetime: time.Hour,   // Rotate connections hourly
    MaxConnIdleTime: time.Minute * 15, // Close idle connections
    ConnectTimeout:  time.Second * 10, // Fast connection timeout
    QueryTimeout:    time.Second * 30, // Query timeout
}
```

**High-Traffic Services:**
- Increase `MaxConns` (50-100)
- Reduce `MaxConnIdleTime` (5-10 minutes)
- Enable connection monitoring

**Low-Traffic Services:**
- Use default settings (20 max connections)
- Longer idle timeouts (15-30 minutes)
- Focus on connection efficiency

### Monitoring and Observability

**Metrics Integration:**
```go
// Export metrics for Prometheus
func (h *Handler) MetricsHandler(c *gin.Context) {
    metrics := h.dbPool.metrics.GetMetrics()
    
    c.JSON(200, gin.H{
        "database_connections_total":    metrics.TotalConnections,
        "database_connections_failed":   metrics.FailedConnections,
        "database_connections_active":   metrics.ActiveConnections,
        "database_queries_total":        metrics.TotalQueries,
        "database_queries_failed":       metrics.FailedQueries,
        "database_query_duration_ns":    metrics.QueryDuration,
        "database_health_checks_total":  metrics.HealthChecks,
        "database_health_checks_failed": metrics.FailedHealthChecks,
        "database_last_health_check":    metrics.LastHealthCheck,
    })
}
```

## Current Limitations

### Missing Tenant-Aware Features (Ticket 1.2.15)

The current `pkg/database` package provides excellent basic connection pooling but **does not yet implement tenant-aware functionality**:

**Not Yet Implemented:**
- ❌ Tenant schema switching based on context
- ❌ Thread-safe tenant context management
- ❌ Per-tenant connection routing
- ❌ Automatic schema validation before queries
- ❌ Tenant-specific connection metrics

**Usage Impact:**
Services must manually manage tenant schema switching:
```go
// Current manual approach (will be automated in 1.2.15)
_, err := dbPool.Exec(ctx, "SET search_path = $1", tenantSchema)
```

### Planned Enhancements (Post-1.2.15)

**Advanced Connection Features:**
- Query logging and tracing
- Automatic retry for specific errors
- Circuit breaker pattern
- Query timeout enforcement per tenant

**Enhanced Metrics:**
- Query performance histograms per tenant
- Tenant-specific connection pool metrics
- Slow query detection and logging
- Resource usage monitoring per tenant

**Multi-Database Support:**
- Read/write split capability
- Multiple database configurations
- Database-specific connection pools
- Failover and load balancing

## Related Documentation

- [Database Design](./database.md) - Multi-tenant database architecture
- [Service Architecture](./services.md) - How services use shared packages
- [SQLC Implementation](./sqlc.md) - Database query generation
- [Development Setup](../development/setup.md) - Local development with shared packages