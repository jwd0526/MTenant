# Database Package

**Last Updated:** 2025-10-23\
*Added automatic metrics tracking integration for all database operations*

A high-performance PostgreSQL connection pooling and management package built on pgx v5. This package provides robust database connectivity, health monitoring, metrics collection, and production-ready features.

## 🏗️ Architecture Overview

The database package provides a comprehensive PostgreSQL integration layer featuring:

- **High-Performance Connection Pooling**: Built on pgx v5 with advanced pool management
- **Health Monitoring**: Real-time database health checks and connection statistics
- **Metrics Collection**: Detailed performance metrics and connection tracking
- **Production Ready**: Retry logic, timeouts, graceful shutdowns, and error handling
- **Configuration Management**: Environment-based configuration with sensible defaults

## 📁 Package Structure

```
database/
├── config.go      # Configuration management and environment parsing
├── health.go      # Health check implementation and monitoring
├── metrics.go     # Performance metrics collection and reporting
├── pool.go        # Connection pool implementation and management
└── README.md      # This documentation
```

## 🔧 Core Components

**Important: Metrics-Enabled Methods**

The `Pool` type embeds `*pgxpool.Pool` but provides wrapper methods for common operations that automatically track metrics:
- Use `pool.Query()`, `pool.QueryRow()`, `pool.Exec()` for automatic metrics tracking
- For advanced operations (transactions, batches, etc.), use `pool.Pool.Begin()`, `pool.Pool.SendBatch()` directly
- All wrapper methods are compatible with the underlying pgx interfaces

### 1. Configuration Management (`config.go`)

Handles database configuration with environment variable support and validation.

```go
// Load configuration from environment
config, err := database.LoadConfigFromEnv()
if err != nil {
    return fmt.Errorf("failed to load database config: %w", err)
}

// Configuration includes all connection parameters
type Config struct {
    Host            string
    Port            int
    Database        string
    Username        string
    Password        string
    MaxConns        int32
    MinConns        int32
    MaxConnLifetime time.Duration
    MaxConnIdleTime time.Duration
    ConnectTimeout  time.Duration
    QueryTimeout    time.Duration
    MaxRetries      int
    RetryInterval   time.Duration
    SSLMode         string
}
```

**Features:**
- **Environment Integration**: Parses `DATABASE_URL` with fallback to individual env vars
- **SSL Configuration**: Automatic SSL mode detection based on environment
- **Validation**: Comprehensive validation of all parameters
- **Defaults**: Production-ready default values

**Supported Environment Variables:**
- `DATABASE_URL`: Full PostgreSQL connection string (preferred)
- `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD`: Individual parameters
- `ENVIRONMENT`: Determines SSL mode defaults (dev=disable, prod=prefer)

### 2. Connection Pool (`pool.go`)

Advanced connection pooling with retry logic and graceful lifecycle management.

```go
// Create database pool with configuration
ctx := context.Background()
pool, err := database.NewPool(ctx, config)
if err != nil {
    return fmt.Errorf("failed to create pool: %w", err)
}
defer pool.Close()

// Use pool for database operations
rows, err := pool.Query(ctx, "SELECT * FROM deals WHERE created_at > $1", time.Now())
if err != nil {
    return err
}
defer rows.Close()
```

**Key Features:**
- **Connection Retry Logic**: Automatic retry with exponential backoff
- **Lifecycle Management**: Proper startup, health verification, and graceful shutdown
- **Embedded pgxpool**: Full compatibility with pgx v5 pool interface
- **Metrics Integration**: Automatic metrics collection for all operations
- **Context Support**: Full context propagation for timeouts and cancellation

**Pool Configuration:**
```go
// Production-optimized defaults
config := &database.Config{
    MaxConns:        20,              // Maximum concurrent connections
    MinConns:        5,               // Minimum idle connections
    MaxConnLifetime: time.Hour,       // Connection lifetime
    MaxConnIdleTime: time.Minute * 5, // Idle connection timeout
    ConnectTimeout:  time.Second * 30, // Connection establishment timeout
    QueryTimeout:    time.Second * 30, // Individual query timeout
}
```

### 3. Health Monitoring (`health.go`)

Comprehensive health checking with detailed connection pool statistics.

```go
// Check database health
ctx := context.Background()
health := pool.HealthCheck(ctx)

if !health.Healthy {
    log.Printf("Database unhealthy: %s", health.Error)
    return
}

log.Printf("Database healthy - Response time: %v", health.ResponseTime)
log.Printf("Active connections: %d/%d", health.Stats.AcquiredConns, health.Stats.MaxConns)
```

**Health Check Features:**
- **Connectivity Test**: Executes simple query to verify database connectivity
- **Response Time Tracking**: Measures query execution time
- **Connection Pool Statistics**: Detailed pool utilization metrics
- **Error Reporting**: Structured error information for troubleshooting

**Health Check Response:**
```go
type HealthStatus struct {
    Healthy      bool          `json:"healthy"`
    ResponseTime time.Duration `json:"response_time"`
    Error        string        `json:"error,omitempty"`
    Stats        *PoolStats    `json:"stats"`
}

type PoolStats struct {
    AcquireCount         int64 `json:"acquire_count"`
    AcquireDuration      int64 `json:"acquire_duration_ns"`
    AcquiredConns        int32 `json:"acquired_conns"`
    CanceledAcquireCount int64 `json:"canceled_acquire_count"`
    ConstructingConns    int32 `json:"constructing_conns"`
    EmptyAcquireCount    int64 `json:"empty_acquire_count"`
    IdleConns            int32 `json:"idle_conns"`
    MaxConns             int32 `json:"max_conns"`
    TotalConns           int32 `json:"total_conns"`
}
```

### 4. Metrics Collection (`metrics.go`)

Thread-safe metrics collection for monitoring database performance with automatic tracking.

```go
// Metrics are automatically collected during operations
// Access pool metrics at any time
metrics := pool.GetMetrics()

log.Printf("Total queries: %d", metrics.TotalQueries)
log.Printf("Failed queries: %d", metrics.FailedQueries)
if metrics.TotalQueries > 0 {
    avgDuration := time.Duration(metrics.QueryDuration / metrics.TotalQueries)
    log.Printf("Average query duration: %v", avgDuration)
}
log.Printf("Active connections: %d", metrics.ActiveConnections)
```

**Collected Metrics:**
- **Connection Metrics**: Total, failed, and active connection counts (tracked during pool creation)
- **Query Performance**: Query counts, failures, and execution times (tracked via wrapper methods)
- **Health Check Status**: Health check frequency and failure rates (tracked during health checks)
- **Timestamps**: Last health check and operation times

**Automatic Metrics Tracking:**
The Pool provides wrapper methods that automatically track metrics:
- `pool.Query()` - Tracks query count, duration, and failures
- `pool.QueryRow()` - Tracks query count and duration
- `pool.Exec()` - Tracks execution count, duration, and failures
- `pool.HealthCheck()` - Tracks health check count and failures
- Connection attempts are tracked during `NewPool()`

**Metrics Structure:**
```go
type Metrics struct {
    TotalConnections    int64  // Total connection attempts
    FailedConnections   int64  // Failed connection attempts
    ActiveConnections   int64  // Currently active connections
    TotalQueries        int64  // Total queries executed
    FailedQueries       int64  // Failed query attempts
    QueryDuration       int64  // Total query execution time (nanoseconds)
    HealthChecks        int64  // Total health checks performed
    FailedHealthChecks  int64  // Failed health check attempts
    LastHealthCheck     int64  // Last health check timestamp (Unix)
}
```

## 🚀 Usage Examples

### Basic Setup

```go
package main

import (
    "context"
    "log"
    "time"
    "crm-platform/deal-service/database"
)

func main() {
    // 1. Load configuration from environment
    config, err := database.LoadConfigFromEnv()
    if err != nil {
        log.Fatalf("Failed to load database config: %v", err)
    }

    // 2. Create database pool with retry logic
    ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
    defer cancel()

    pool, err := database.NewPool(ctx, config)
    if err != nil {
        log.Fatalf("Failed to create database pool: %v", err)
    }
    defer pool.Close()

    // 3. Verify database health
    health := pool.HealthCheck(context.Background())
    if !health.Healthy {
        log.Fatalf("Database health check failed: %s", health.Error)
    }

    log.Printf("Database connected successfully - Response time: %v", health.ResponseTime)

    // 4. Use pool for database operations (with automatic metrics tracking)
    rows, err := pool.Query(context.Background(), "SELECT version()")
    if err != nil {
        log.Fatalf("Query failed: %v", err)
    }
    defer rows.Close()

    for rows.Next() {
        var version string
        if err := rows.Scan(&version); err != nil {
            log.Fatalf("Scan failed: %v", err)
        }
        log.Printf("PostgreSQL version: %s", version)
    }

    // 5. Check metrics
    metrics := pool.GetMetrics()
    log.Printf("Total queries executed: %d", metrics.TotalQueries)
}
```

### Web Service Integration

```go
package main

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "time"
    
    "crm-platform/deal-service/database"
)

func main() {
    // Initialize database
    config, _ := database.LoadConfigFromEnv()
    pool, err := database.NewPool(context.Background(), config)
    if err != nil {
        log.Fatal(err)
    }
    defer pool.Close()
    
    // Health check endpoint
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        health := pool.HealthCheck(r.Context())
        
        w.Header().Set("Content-Type", "application/json")
        
        if health.Healthy {
            w.WriteHeader(http.StatusOK)
        } else {
            w.WriteHeader(http.StatusServiceUnavailable)
        }
        
        json.NewEncoder(w).Encode(health)
    })
    
    // Metrics endpoint
    http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
        metrics := pool.GetMetrics()
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(metrics)
    })
    
    // Start server
    server := &http.Server{
        Addr:    ":8080",
        Handler: http.DefaultServeMux,
    }
    
    log.Printf("Server starting on :8080")
    log.Fatal(server.ListenAndServe())
}
```

### Transaction Handling

```go
func TransferFunds(ctx context.Context, pool *database.Pool, fromAccount, toAccount int32, amount float64) error {
    // Begin transaction
    tx, err := pool.Begin(ctx)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback(ctx) // Always rollback on error
    
    // Debit source account
    _, err = tx.Exec(ctx, 
        "UPDATE accounts SET balance = balance - $1 WHERE id = $2", 
        amount, fromAccount)
    if err != nil {
        return fmt.Errorf("failed to debit account: %w", err)
    }
    
    // Credit destination account
    _, err = tx.Exec(ctx, 
        "UPDATE accounts SET balance = balance + $1 WHERE id = $2", 
        amount, toAccount)
    if err != nil {
        return fmt.Errorf("failed to credit account: %w", err)
    }
    
    // Record transaction log
    _, err = tx.Exec(ctx,
        "INSERT INTO transaction_log (from_account, to_account, amount, created_at) VALUES ($1, $2, $3, $4)",
        fromAccount, toAccount, amount, time.Now())
    if err != nil {
        return fmt.Errorf("failed to log transaction: %w", err)
    }
    
    // Commit transaction
    if err := tx.Commit(ctx); err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }
    
    return nil
}
```

### Advanced Configuration

```go
func CreateOptimizedPool() (*database.Pool, error) {
    // Custom configuration for high-load scenarios
    config := &database.Config{
        Host:            "db.example.com",
        Port:            5432,
        Database:        "production_db",
        Username:        "app_user",
        Password:        "secure_password",
        MaxConns:        50,              // Higher for high-load
        MinConns:        10,              // Maintain warm connections
        MaxConnLifetime: time.Hour * 2,   // Longer lifetime
        MaxConnIdleTime: time.Minute * 10, // Longer idle time
        ConnectTimeout:  time.Second * 15, // Shorter connection timeout
        QueryTimeout:    time.Second * 60, // Longer query timeout for analytics
        MaxRetries:      3,               // Fewer retries for faster failure
        RetryInterval:   time.Second * 5,  // Shorter retry interval
        SSLMode:         "require",       // Required SSL for production
    }
    
    ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
    defer cancel()
    
    pool, err := database.NewPool(ctx, config)
    if err != nil {
        return nil, fmt.Errorf("failed to create optimized pool: %w", err)
    }
    
    // Verify initial health
    health := pool.HealthCheck(context.Background())
    if !health.Healthy {
        pool.Close()
        return nil, fmt.Errorf("initial health check failed: %s", health.Error)
    }
    
    return pool, nil
}
```

## 🔒 Security Features

### 1. Connection Security

```go
// SSL configuration based on environment
func (c *Config) ConnectionString() string {
    // Automatic SSL mode selection
    if config.IsDevelopmentMode() {
        sslMode = "disable"  // Development convenience
    } else {
        sslMode = "require"  // Production security
    }
    
    return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
        c.Username, c.Password, c.Host, c.Port, c.Database, c.SSLMode)
}
```

### 2. Credential Management

```go
// Environment-based credential loading
config, err := database.LoadConfigFromEnv()
if err != nil {
    return fmt.Errorf("failed to load config: %w", err)
}

// Validates required credentials
if config.Username == "" || config.Password == "" {
    return database.ErrPasswordRequired
}
```

### 3. Connection Validation

```go
// Parse and validate database URL
u, err := url.Parse(dbConnString)
if err != nil {
    return nil, fmt.Errorf("%w: %v", ErrInvalidDatabaseURL, err)
}

// Validate database name exists
if db == "" {
    return nil, ErrDatabaseNameRequired
}

// Ensure password is provided
if !hasPassword {
    return nil, ErrPasswordRequired
}
```

## 📊 Performance Optimization

### 1. Connection Pool Tuning

```go
// Production recommendations
config := &database.Config{
    // Set MaxConns based on your workload:
    // - Web apps: 10-20 connections
    // - High-throughput APIs: 20-50 connections
    // - Background workers: 5-15 connections
    MaxConns: 20,
    
    // Keep minimum connections warm
    // - Should be ~25% of MaxConns
    MinConns: 5,
    
    // Connection lifetime management
    // - Rotate connections regularly
    // - Prevents connection leaks
    MaxConnLifetime: time.Hour,
    
    // Idle connection cleanup
    // - Free unused connections
    // - Reduces resource usage
    MaxConnIdleTime: time.Minute * 5,
}
```

### 2. Query Optimization

```go
// Use contexts for query timeouts
ctx, cancel := context.WithTimeout(context.Background(), time.Second * 30)
defer cancel()

// Long-running queries get appropriate timeouts
rows, err := pool.Query(ctx, "SELECT * FROM large_table")
if err != nil {
    // Handle timeout vs other errors
    if ctx.Err() == context.DeadlineExceeded {
        return fmt.Errorf("query timed out")
    }
    return fmt.Errorf("query failed: %w", err)
}
```

### 3. Monitoring and Metrics

```go
// Regular metrics collection
func monitorDatabase(pool *database.Pool) {
    ticker := time.NewTicker(time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        metrics := pool.GetMetrics()
        health := pool.HealthCheck(context.Background())
        
        // Log key metrics
        log.Printf("Database metrics - Queries: %d, Errors: %d, Response: %v", 
            metrics.TotalQueries, 
            metrics.FailedQueries, 
            health.ResponseTime)
        
        // Alert on high error rate
        if metrics.TotalQueries > 0 {
            errorRate := float64(metrics.FailedQueries) / float64(metrics.TotalQueries)
            if errorRate > 0.05 { // 5% error threshold
                log.Printf("HIGH ERROR RATE: %.2f%%", errorRate*100)
            }
        }
        
        // Alert on connection pool exhaustion
        if health.Stats.AcquiredConns >= health.Stats.MaxConns-2 {
            log.Printf("CONNECTION POOL NEAR EXHAUSTION: %d/%d", 
                health.Stats.AcquiredConns, 
                health.Stats.MaxConns)
        }
    }
}
```

## 🧪 Testing Support

### Test Database Setup

```go
func setupTestDatabase(t *testing.T) *database.Pool {
    // Load test configuration
    config := &database.Config{
        Host:            "localhost",
        Port:            5432,
        Database:        "test_db",
        Username:        "test_user",
        Password:        "test_pass",
        MaxConns:        5,              // Smaller pool for tests
        MinConns:        1,
        MaxConnLifetime: time.Minute * 5, // Shorter lifetime
        MaxConnIdleTime: time.Minute,    // Quick cleanup
        ConnectTimeout:  time.Second * 10,
        QueryTimeout:    time.Second * 10,
        SSLMode:         "disable",      // Test convenience
    }
    
    ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
    defer cancel()
    
    pool, err := database.NewPool(ctx, config)
    require.NoError(t, err)
    
    // Verify test database health
    health := pool.HealthCheck(context.Background())
    require.True(t, health.Healthy, "Test database must be healthy")
    
    return pool
}
```

### Test Cleanup

```go
func TestWithDatabase(t *testing.T) {
    pool := setupTestDatabase(t)
    defer pool.Close()
    
    // Run test operations
    _, err := pool.Exec(context.Background(), "CREATE TEMPORARY TABLE test_data (id SERIAL, name TEXT)")
    require.NoError(t, err)
    
    // Test automatically cleans up when pool closes
}
```

### Integration Testing

```go
func TestDatabaseIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping database integration test")
    }
    
    pool := setupTestDatabase(t)
    defer pool.Close()
    
    // Test basic operations
    t.Run("InsertAndQuery", func(t *testing.T) {
        tx, err := pool.Begin(context.Background())
        require.NoError(t, err)
        defer tx.Rollback(context.Background())
        
        _, err = tx.Exec(context.Background(), 
            "CREATE TEMPORARY TABLE users (id SERIAL PRIMARY KEY, name TEXT)")
        require.NoError(t, err)
        
        _, err = tx.Exec(context.Background(), 
            "INSERT INTO users (name) VALUES ($1)", "Test User")
        require.NoError(t, err)
        
        var name string
        err = tx.QueryRow(context.Background(), 
            "SELECT name FROM users WHERE id = 1").Scan(&name)
        require.NoError(t, err)
        assert.Equal(t, "Test User", name)
    })
    
    // Test health checks
    t.Run("HealthCheck", func(t *testing.T) {
        health := pool.HealthCheck(context.Background())
        assert.True(t, health.Healthy)
        assert.Greater(t, health.ResponseTime, time.Duration(0))
        assert.NotNil(t, health.Stats)
    })
    
    // Test metrics
    t.Run("Metrics", func(t *testing.T) {
        initialMetrics := pool.GetMetrics()
        
        // Perform operation
        _, err := pool.Exec(context.Background(), "SELECT 1")
        require.NoError(t, err)
        
        finalMetrics := pool.GetMetrics()
        assert.Greater(t, finalMetrics.TotalQueries, initialMetrics.TotalQueries)
    })
}
```

## 🚨 Error Handling

### Configuration Errors

```go
var (
    ErrDatabaseURLRequired  = fmt.Errorf("DATABASE_URL environment variable is required")
    ErrInvalidDatabaseURL   = fmt.Errorf("failed to parse DATABASE_URL")
    ErrInvalidPort         = fmt.Errorf("invalid port number")
    ErrDatabaseNameRequired = fmt.Errorf("database name is required in URL")
    ErrPasswordRequired    = fmt.Errorf("password is required in DATABASE_URL")
)
```

### Pool Errors

```go
var (
    ErrNilConfig           = fmt.Errorf("config cannot be nil")
    ErrConnectionTimeout   = fmt.Errorf("connection timeout exceeded")
    ErrMaxRetriesExceeded  = fmt.Errorf("maximum retry attempts exceeded")
)
```

### Error Handling Patterns

```go
// Configuration error handling
config, err := database.LoadConfigFromEnv()
if err != nil {
    switch {
    case errors.Is(err, database.ErrDatabaseURLRequired):
        log.Fatal("DATABASE_URL environment variable must be set")
    case errors.Is(err, database.ErrInvalidDatabaseURL):
        log.Fatalf("Invalid DATABASE_URL format: %v", err)
    default:
        log.Fatalf("Database configuration error: %v", err)
    }
}

// Pool creation error handling
pool, err := database.NewPool(ctx, config)
if err != nil {
    if errors.Is(err, database.ErrConnectionTimeout) {
        log.Fatal("Database connection timed out - check network and credentials")
    }
    log.Fatalf("Failed to create database pool: %v", err)
}

// Query error handling with context
ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
defer cancel()

rows, err := pool.Query(ctx, "SELECT * FROM large_table")
if err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        return fmt.Errorf("query timed out after 30 seconds")
    }
    return fmt.Errorf("database query failed: %w", err)
}
```

## 🔧 Configuration Reference

### Environment Variables

| Variable | Description | Example | Default |
|----------|-------------|---------|---------|
| `DATABASE_URL` | Complete PostgreSQL connection string | `postgres://user:pass@localhost:5432/db?sslmode=disable` | Required |
| `DB_HOST` | Database host (if not using DATABASE_URL) | `localhost` | `localhost` |
| `DB_PORT` | Database port | `5432` | `5432` |
| `DB_NAME` | Database name | `myapp` | Required |
| `DB_USER` | Database username | `appuser` | Required |
| `DB_PASSWORD` | Database password | `secretpass` | Required |
| `DB_SSLMODE` | SSL mode | `require`, `disable` | `prefer` |

### Configuration Defaults

```go
// Production-ready defaults
config := Config{
    MaxConns:        20,                // Concurrent connections
    MinConns:        5,                 // Minimum idle connections  
    MaxConnLifetime: time.Minute * 60,  // 1 hour connection lifetime
    MaxConnIdleTime: time.Minute * 5,   // 5 minute idle timeout
    ConnectTimeout:  time.Second * 30,  // 30 second connection timeout
    QueryTimeout:    time.Second * 30,  // 30 second query timeout
    MaxRetries:      5,                 // Connection retry attempts
    RetryInterval:   time.Second * 10,  // 10 second retry delay
    SSLMode:         "prefer",          // SSL preferred
}
```

## 🚀 Production Deployment

### Environment Setup

```bash
# Production environment variables
export ENVIRONMENT=production
export DATABASE_URL="postgres://appuser:securepassword@db.example.com:5432/production_db?sslmode=require"

# Optional overrides
export DB_MAX_CONNS=30
export DB_MIN_CONNS=10
export DB_CONNECT_TIMEOUT=15s
export DB_QUERY_TIMEOUT=60s
```

### Docker Configuration

```dockerfile
# Multi-stage build for production
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o main cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .

# Health check using database health endpoint
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./main"]
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deal-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: deal-service
  template:
    metadata:
      labels:
        app: deal-service
    spec:
      containers:
      - name: deal-service
        image: deal-service:latest
        ports:
        - containerPort: 8080
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: database-secret
              key: url
        - name: ENVIRONMENT
          value: "production"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          limits:
            memory: "512Mi"
            cpu: "500m"
          requests:
            memory: "256Mi"
            cpu: "250m"
```

## 📚 Best Practices

### 1. **Connection Pool Sizing**
```go
// Size pool based on workload
// - CPU-bound: MaxConns = CPU cores * 2
// - I/O-bound: MaxConns = higher (20-50)
// - Mixed workload: Start with 20, monitor and adjust

config.MaxConns = 20
config.MinConns = config.MaxConns / 4 // Keep 25% warm
```

### 2. **Error Handling**
```go
// Always handle context timeouts
ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
defer cancel()

rows, err := pool.Query(ctx, sql, args...)
if err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        // Handle timeout specifically
        return ErrQueryTimeout
    }
    return fmt.Errorf("query failed: %w", err)
}
```

### 3. **Resource Management**
```go
// Always close resources
rows, err := pool.Query(ctx, sql, args...)
if err != nil {
    return err
}
defer rows.Close() // Essential - prevents connection leaks

// Use defer for cleanup in all code paths
tx, err := pool.Begin(ctx)
if err != nil {
    return err
}
defer tx.Rollback(ctx) // Safe to call even after Commit
```

### 4. **Health Monitoring**
```go
// Regular health checks
go func() {
    ticker := time.NewTicker(time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        health := pool.HealthCheck(context.Background())
        if !health.Healthy {
            log.Printf("Database unhealthy: %s", health.Error)
            // Implement alerting here
        }
    }
}()
```

### 5. **Graceful Shutdown**
```go
// Proper shutdown sequence
func gracefulShutdown(pool *database.Pool, server *http.Server) {
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    
    <-sigChan
    log.Println("Shutting down gracefully...")
    
    // 1. Stop accepting new requests
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := server.Shutdown(ctx); err != nil {
        log.Printf("HTTP server shutdown error: %v", err)
    }
    
    // 2. Close database connections
    pool.Close()
    log.Println("Database pool closed")
    
    log.Println("Shutdown complete")
}
```

---

This database package provides enterprise-grade PostgreSQL connectivity with production-ready features, comprehensive monitoring, and excellent performance characteristics.