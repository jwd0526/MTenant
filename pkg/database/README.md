# Database Package Documentation

[![Go Reference](https://pkg.go.dev/badge/crm-platform/pkg/database.svg)](https://pkg.go.dev/crm-platform/pkg/database)
[![Go Report Card](https://goreportcard.com/badge/crm-platform/pkg/database)](https://goreportcard.com/report/crm-platform/pkg/database)

The `database` package provides robust database connection management, health monitoring, and metrics collection for PostgreSQL-based applications. It offers connection pooling with automatic retry logic, comprehensive health checks, and performance monitoring capabilities.

## Table of Contents

- [Features](#features)
- [Quick Start](#quick-start)
- [Architecture](#architecture)
- [API Reference](#api-reference)
- [Examples](#examples)
- [Best Practices](#best-practices)
- [Testing](#testing)
- [Contributing](#contributing)

## Features

- **Environment-Based Configuration** - Automatic configuration from `DATABASE_URL` environment variable
- **Connection Pooling** - Efficient connection pooling with configurable limits and timeouts
- **Automatic Retry Logic** - Built-in retry mechanism for connection failures with exponential backoff
- **Health Monitoring** - Comprehensive health checks with detailed statistics and timing
- **Metrics Collection** - Thread-safe metrics tracking for observability and monitoring
- **SSL Support** - Automatic SSL mode detection based on environment
- **Production Ready** - Battle-tested error handling and logging
- **High Performance** - Minimal overhead with optimized connection management

## Quick Start

### Installation

```bash
go get crm-platform/pkg/database
```

### Basic Usage

```go
package main

import (
    "context"
    "log"
    
    "crm-platform/pkg/database"
)

func main() {
    // Load configuration from DATABASE_URL environment variable
    config, err := database.LoadConfigFromEnv()
    if err != nil {
        log.Fatal("Failed to load config:", err)
    }
    
    // Create connection pool with retry logic
    ctx := context.Background()
    pool, err := database.NewPool(ctx, config)
    if err != nil {
        log.Fatal("Failed to create pool:", err)
    }
    defer pool.Close()
    
    // Verify database health
    health := pool.HealthCheck(ctx)
    if !health.Healthy {
        log.Fatal("Database unhealthy:", health.Error)
    }
    
    log.Printf("Database connected in %v", health.ResponseTime)
    log.Printf("Pool stats: %+v", health.Stats)
    
    // Execute queries
    rows, err := pool.Query(ctx, "SELECT id, name FROM users WHERE active = $1", true)
    if err != nil {
        log.Fatal("Query failed:", err)
    }
    defer rows.Close()
    
    // Process results...
}
```

## Architecture

### Connection Pool Architecture

The database package wraps `pgxpool.Pool` with additional functionality for health monitoring, metrics collection, and enhanced error handling:

```
Application Layer
       ↓
Database Package (database.Pool)
       ↓
pgxpool.Pool (PostgreSQL Driver)
       ↓
PostgreSQL Database
```

### Configuration Flow

```mermaid
graph LR
    A[DATABASE_URL] --> B[LoadConfigFromEnv]
    B --> C[Parse URL]
    C --> D[Extract Components]
    D --> E[Set Defaults]
    E --> F[Create Config]
    F --> G[NewPool]
    G --> H[Connection Pool]
```

### Components

| Component | Purpose | Responsibility |
|-----------|---------|----------------|
| **Config** | Configuration management | Parse DATABASE_URL, set defaults, validate settings |
| **Pool** | Connection pooling | Manage connections, execute queries, health monitoring |
| **HealthStatus** | Health monitoring | Track database health, response times, error states |
| **Metrics** | Performance tracking | Collect connection/query metrics, thread-safe counters |

## API Reference

### Configuration Management

#### `LoadConfigFromEnv() (*Config, error)`

Loads database configuration from the `DATABASE_URL` environment variable.

```go
config, err := database.LoadConfigFromEnv()
if err != nil {
    log.Fatal("Config error:", err)
}
```

**Environment Variable Format:**
```bash
DATABASE_URL=postgresql://username:password@host:port/database?sslmode=disable
```

**Configuration Options:**
- `MaxConns`: Maximum connections (default: 20)
- `MinConns`: Minimum connections (default: 5)
- `MaxConnLifetime`: Connection lifetime (default: 60 minutes)
- `MaxConnIdleTime`: Idle timeout (default: 5 minutes)
- `ConnectTimeout`: Connection timeout (default: 30 seconds)
- `QueryTimeout`: Query timeout (default: 30 seconds)
- `MaxRetries`: Retry attempts (default: 5)
- `RetryInterval`: Retry delay (default: 10 seconds)

#### `(*Config) ConnectionString() string`

Returns the PostgreSQL connection string for pgx.

```go
connStr := config.ConnectionString()
// Returns: "postgres://user:pass@host:port/db?sslmode=disable"
```

### Connection Pool Management

#### `NewPool(ctx context.Context, config *Config) (*Pool, error)`

Creates a new database connection pool with retry logic and health validation.

```go
ctx := context.Background()
pool, err := database.NewPool(ctx, config)
if err != nil {
    return fmt.Errorf("pool creation failed: %w", err)
}
defer pool.Close()
```

**Features:**
- Automatic retry with exponential backoff
- Connection validation on startup
- Configurable pool settings
- Embedded metrics collection

#### `(*Pool) Close()`

Gracefully closes the connection pool and all connections.

```go
defer pool.Close() // Always defer pool closure
```

#### `(*Pool) Stats() *pgxpool.Stat`

Returns detailed connection pool statistics.

```go
stats := pool.Stats()
fmt.Printf("Active connections: %d/%d", stats.AcquiredConns(), stats.MaxConns())
```

### Health Monitoring

#### `(*Pool) HealthCheck(ctx context.Context) *HealthStatus`

Performs comprehensive health check with timing and detailed diagnostics.

```go
health := pool.HealthCheck(ctx)
if !health.Healthy {
    log.Printf("Database unhealthy: %s", health.Error)
    log.Printf("Response time: %v", health.ResponseTime)
    log.Printf("Pool stats: %+v", health.Stats)
}
```

**Health Check Process:**
1. **Ping Test** - Basic connectivity validation (5-second timeout)
2. **Query Test** - Execute `SELECT 1` to verify functionality
3. **Pool Statistics** - Collect connection pool metrics
4. **Response Timing** - Measure total health check duration

#### `(*Pool) IsHealthy(ctx context.Context) bool`

Simple health check that returns boolean status.

```go
if pool.IsHealthy(ctx) {
    log.Println("Database is healthy")
} else {
    log.Println("Database is unhealthy")
}
```

### Query Execution

The `Pool` embeds `pgxpool.Pool`, providing access to all standard pgx methods:

#### Basic Queries

```go
// Query multiple rows
rows, err := pool.Query(ctx, "SELECT id, name FROM users WHERE active = $1", true)
if err != nil {
    return fmt.Errorf("query failed: %w", err)
}
defer rows.Close()

for rows.Next() {
    var id int
    var name string
    if err := rows.Scan(&id, &name); err != nil {
        return fmt.Errorf("scan failed: %w", err)
    }
    fmt.Printf("User: %d - %s\n", id, name)
}
```

#### Single Row Queries

```go
// Query single row
var count int
err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
if err != nil {
    return fmt.Errorf("count query failed: %w", err)
}
```

#### Command Execution

```go
// Execute commands (INSERT, UPDATE, DELETE)
tag, err := pool.Exec(ctx, 
    "INSERT INTO users (name, email) VALUES ($1, $2)", 
    "John Doe", "john@example.com")
if err != nil {
    return fmt.Errorf("insert failed: %w", err)
}

fmt.Printf("Inserted %d rows", tag.RowsAffected())
```

#### Transactions

```go
// Begin transaction
tx, err := pool.Begin(ctx)
if err != nil {
    return fmt.Errorf("begin transaction failed: %w", err)
}
defer tx.Rollback(ctx) // Always defer rollback

// Execute multiple operations
_, err = tx.Exec(ctx, "INSERT INTO users (name) VALUES ($1)", "User 1")
if err != nil {
    return fmt.Errorf("first insert failed: %w", err)
}

_, err = tx.Exec(ctx, "INSERT INTO users (name) VALUES ($1)", "User 2")
if err != nil {
    return fmt.Errorf("second insert failed: %w", err)
}

// Commit transaction
if err = tx.Commit(ctx); err != nil {
    return fmt.Errorf("commit failed: %w", err)
}
```

## Examples

### Complete Application Setup

```go
func SetupDatabase() (*database.Pool, error) {
    // Load configuration
    config, err := database.LoadConfigFromEnv()
    if err != nil {
        return nil, fmt.Errorf("config loading failed: %w", err)
    }
    
    // Override defaults for production
    config.MaxConns = 50
    config.MinConns = 10
    config.MaxConnLifetime = time.Hour
    config.MaxConnIdleTime = time.Minute * 15
    
    // Create pool
    ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
    defer cancel()
    
    pool, err := database.NewPool(ctx, config)
    if err != nil {
        return nil, fmt.Errorf("pool creation failed: %w", err)
    }
    
    // Verify health
    health := pool.HealthCheck(ctx)
    if !health.Healthy {
        pool.Close()
        return nil, fmt.Errorf("database unhealthy: %s", health.Error)
    }
    
    log.Printf("Database connected successfully in %v", health.ResponseTime)
    return pool, nil
}
```

### HTTP Health Check Handler

```go
func DatabaseHealthHandler(pool *database.Pool) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        health := pool.HealthCheck(r.Context())
        
        status := 200
        if !health.Healthy {
            status = 503
        }
        
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(status)
        
        response := map[string]interface{}{
            "status":        health.Healthy,
            "response_time": health.ResponseTime.String(),
            "error":         health.Error,
            "database": map[string]interface{}{
                "max_conns":    health.Stats.MaxConns,
                "total_conns":  health.Stats.TotalConns,
                "active_conns": health.Stats.AcquiredConns,
                "idle_conns":   health.Stats.IdleConns,
            },
        }
        
        json.NewEncoder(w).Encode(response)
    }
}
```

### Metrics Collection

```go
func CollectDatabaseMetrics(pool *database.Pool) {
    ticker := time.NewTicker(time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        health := pool.HealthCheck(context.Background())
        
        // Log metrics
        log.Printf("Database Health: %t", health.Healthy)
        log.Printf("Response Time: %v", health.ResponseTime)
        log.Printf("Active Connections: %d/%d", 
            health.Stats.AcquiredConns, health.Stats.MaxConns)
        
        // Export to monitoring system (Prometheus, etc.)
        if health.Healthy {
            recordMetric("database.healthy", 1)
        } else {
            recordMetric("database.healthy", 0)
        }
        
        recordMetric("database.response_time_ms", 
            float64(health.ResponseTime.Nanoseconds())/1000000)
        recordMetric("database.connections.active", 
            float64(health.Stats.AcquiredConns))
        recordMetric("database.connections.idle", 
            float64(health.Stats.IdleConns))
    }
}
```

### Repository Pattern Integration

```go
type UserRepository struct {
    pool *database.Pool
}

func NewUserRepository(pool *database.Pool) *UserRepository {
    return &UserRepository{pool: pool}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *User) error {
    query := `
        INSERT INTO users (name, email, created_at) 
        VALUES ($1, $2, $3) 
        RETURNING id`
    
    err := r.pool.QueryRow(ctx, query, user.Name, user.Email, time.Now()).
        Scan(&user.ID)
    if err != nil {
        return fmt.Errorf("failed to create user: %w", err)
    }
    
    return nil
}

func (r *UserRepository) GetUser(ctx context.Context, id int) (*User, error) {
    query := `SELECT id, name, email, created_at FROM users WHERE id = $1`
    
    var user User
    err := r.pool.QueryRow(ctx, query, id).
        Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt)
    if err != nil {
        if err == pgx.ErrNoRows {
            return nil, ErrUserNotFound
        }
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    
    return &user, nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, user *User) error {
    query := `
        UPDATE users 
        SET name = $2, email = $3, updated_at = $4 
        WHERE id = $1`
    
    tag, err := r.pool.Exec(ctx, query, 
        user.ID, user.Name, user.Email, time.Now())
    if err != nil {
        return fmt.Errorf("failed to update user: %w", err)
    }
    
    if tag.RowsAffected() == 0 {
        return ErrUserNotFound
    }
    
    return nil
}
```

### Graceful Shutdown

```go
func GracefulShutdown(pool *database.Pool, server *http.Server) {
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    
    <-sigChan
    log.Println("Shutting down gracefully...")
    
    // Create shutdown context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    // Shutdown HTTP server
    if err := server.Shutdown(ctx); err != nil {
        log.Printf("HTTP server shutdown error: %v", err)
    }
    
    // Close database pool
    pool.Close()
    log.Println("Database pool closed")
    
    log.Println("Shutdown complete")
}
```

## Best Practices

### ✅ Do's

1. **Always check configuration early**
   ```go
   func main() {
       config, err := database.LoadConfigFromEnv()
       if err != nil {
           log.Fatal("Database configuration failed:", err)
       }
       // Continue with application setup
   }
   ```

2. **Use context for timeouts and cancellation**
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
   defer cancel()
   
   rows, err := pool.Query(ctx, "SELECT * FROM large_table")
   ```

3. **Always defer resource cleanup**
   ```go
   pool, err := database.NewPool(ctx, config)
   if err != nil {
       return err
   }
   defer pool.Close() // Always defer pool closure
   
   rows, err := pool.Query(ctx, sql)
   if err != nil {
       return err
   }
   defer rows.Close() // Always defer rows closure
   ```

4. **Implement proper health checks**
   ```go
   func healthCheckEndpoint(pool *database.Pool) http.HandlerFunc {
       return func(w http.ResponseWriter, r *http.Request) {
           health := pool.HealthCheck(r.Context())
           if !health.Healthy {
               http.Error(w, health.Error, http.StatusServiceUnavailable)
               return
           }
           w.WriteHeader(http.StatusOK)
       }
   }
   ```

5. **Use transactions for related operations**
   ```go
   tx, err := pool.Begin(ctx)
   if err != nil {
       return err
   }
   defer tx.Rollback(ctx) // Always defer rollback
   
   // Multiple related operations
   _, err = tx.Exec(ctx, "INSERT INTO orders ...")
   if err != nil {
       return err
   }
   
   _, err = tx.Exec(ctx, "UPDATE inventory ...")
   if err != nil {
       return err
   }
   
   return tx.Commit(ctx)
   ```

### ❌ Don'ts

1. **Don't ignore health check results**
   ```go
   // ❌ DON'T: Ignore health status
   pool.HealthCheck(ctx) // Result ignored
   
   // ✅ DO: Check health status
   health := pool.HealthCheck(ctx)
   if !health.Healthy {
       return fmt.Errorf("database unhealthy: %s", health.Error)
   }
   ```

2. **Don't create multiple pools unnecessarily**
   ```go
   // ❌ DON'T: Create multiple pools for the same database
   pool1, _ := database.NewPool(ctx, config)
   pool2, _ := database.NewPool(ctx, config) // Wasteful
   
   // ✅ DO: Reuse a single pool across your application
   pool, _ := database.NewPool(ctx, config)
   userRepo := NewUserRepository(pool)
   orderRepo := NewOrderRepository(pool) // Reuse same pool
   ```

3. **Don't ignore connection errors**
   ```go
   // ❌ DON'T: Ignore connection errors
   pool, _ := database.NewPool(ctx, config) // Error ignored!
   
   // ✅ DO: Handle connection errors
   pool, err := database.NewPool(ctx, config)
   if err != nil {
       return fmt.Errorf("database connection failed: %w", err)
   }
   ```

4. **Don't use overly large connection pools**
   ```go
   // ❌ DON'T: Set unreasonably high connection limits
   config.MaxConns = 1000 // Too many connections
   
   // ✅ DO: Use reasonable connection limits
   config.MaxConns = 20   // Start with default
   config.MinConns = 5    // Adjust based on load testing
   ```

### Performance Considerations

- **Connection Pooling**: Use appropriate pool sizes based on your application load
- **Query Timeouts**: Set reasonable timeouts to prevent resource exhaustion
- **Health Check Frequency**: Balance monitoring needs with performance impact
- **Retry Logic**: Configure retry intervals based on your infrastructure
- **SSL Overhead**: Use SSL only when required for your environment

### Security Considerations

- **Environment Variables**: Store sensitive configuration in environment variables
- **SSL Modes**: Use appropriate SSL modes for your environment (prefer/require in production)
- **Connection Limits**: Prevent resource exhaustion with reasonable connection limits
- **Query Timeouts**: Protect against long-running queries
- **Error Handling**: Don't expose sensitive database information in error messages

## Testing

### Unit Testing

```go
func TestDatabaseConfig(t *testing.T) {
    // Test configuration loading
    os.Setenv("DATABASE_URL", "postgresql://user:pass@localhost:5432/testdb")
    defer os.Unsetenv("DATABASE_URL")
    
    config, err := database.LoadConfigFromEnv()
    assert.NoError(t, err)
    assert.Equal(t, "localhost", config.Host)
    assert.Equal(t, 5432, config.Port)
    assert.Equal(t, "testdb", config.Database)
}
```

### Integration Testing

```go
func TestDatabaseConnection(t *testing.T) {
    // Skip if no DATABASE_URL
    if os.Getenv("DATABASE_URL") == "" {
        t.Skip("DATABASE_URL not set, skipping integration test")
    }
    
    config, err := database.LoadConfigFromEnv()
    require.NoError(t, err)
    
    ctx := context.Background()
    pool, err := database.NewPool(ctx, config)
    require.NoError(t, err)
    defer pool.Close()
    
    // Test health check
    health := pool.HealthCheck(ctx)
    assert.True(t, health.Healthy)
    assert.NotEmpty(t, health.Stats)
    
    // Test query execution
    var result int
    err = pool.QueryRow(ctx, "SELECT 1").Scan(&result)
    assert.NoError(t, err)
    assert.Equal(t, 1, result)
}
```

### Test Utilities

```go
// Test helper for setting up test database
func SetupTestDB(t *testing.T) *database.Pool {
    config, err := database.LoadConfigFromEnv()
    require.NoError(t, err)
    
    // Override for testing
    config.MaxConns = 5
    config.MinConns = 1
    
    ctx := context.Background()
    pool, err := database.NewPool(ctx, config)
    require.NoError(t, err)
    
    // Verify health
    health := pool.HealthCheck(ctx)
    require.True(t, health.Healthy, "Database must be healthy for tests")
    
    return pool
}

// Test helper for transaction testing
func WithTestTransaction(t *testing.T, pool *database.Pool, 
    fn func(*testing.T, pgx.Tx)) {
    
    ctx := context.Background()
    tx, err := pool.Begin(ctx)
    require.NoError(t, err)
    defer tx.Rollback(ctx) // Always rollback test transactions
    
    fn(t, tx)
}
```

## Error Handling

The package provides structured error handling with predefined error variables:

### Configuration Errors
```go
var (
    ErrDatabaseURLRequired = fmt.Errorf("DATABASE_URL environment variable is required")
    ErrInvalidDatabaseURL  = fmt.Errorf("failed to parse DATABASE_URL")
    ErrInvalidPort         = fmt.Errorf("invalid port number")
    ErrDatabaseNameRequired = fmt.Errorf("database name is required in URL")
    ErrPasswordRequired    = fmt.Errorf("password is required in DATABASE_URL")
)
```

### Pool Errors
```go
var (
    ErrNilConfig          = fmt.Errorf("config cannot be nil")
    ErrPoolConfigParse    = fmt.Errorf("unable to parse pool config")
    ErrConnectionFailed   = fmt.Errorf("failed to create connection pool")
)
```

### Error Handling Patterns

```go
import (
    "errors"
    "crm-platform/pkg/database"
)

// Check for specific error types using errors.Is
config, err := database.LoadConfigFromEnv()
if err != nil {
    if errors.Is(err, database.ErrDatabaseURLRequired) {
        // Handle missing DATABASE_URL
        return setupDefaultConfig()
    }
    return fmt.Errorf("config loading failed: %w", err)
}

// Handle pool creation errors
pool, err := database.NewPool(ctx, config)
if err != nil {
    if errors.Is(err, database.ErrConnectionFailed) {
        // Handle connection failure - maybe retry or fallback
        return handleConnectionFailure(err)
    }
    return fmt.Errorf("pool creation failed: %w", err)
}

// Handle health check failures
health := pool.HealthCheck(ctx)
if !health.Healthy {
    log.Printf("Database unhealthy: %s", health.Error)
    // Implement fallback logic or alerting
    return handleUnhealthyDatabase(health)
}
```

## Migration Guide

### From Direct pgx Usage

**Before:**
```go
conn, err := pgx.Connect(ctx, "postgresql://user:pass@host/db")
```

**After:**
```go
config, err := database.LoadConfigFromEnv()
if err != nil {
    return err
}
pool, err := database.NewPool(ctx, config)
```

### From Other Connection Pooling Libraries

**Before:**
```go
db, err := sql.Open("postgres", connectionString)
```

**After:**
```go
config, err := database.LoadConfigFromEnv()
if err != nil {
    return err
}
pool, err := database.NewPool(ctx, config)
// pool implements all pgxpool.Pool methods
```

## Related Documentation

- [Tenant Package](../tenant/README.md) - Multi-tenant database isolation
- [Service Architecture](../../docs/architecture/services.md) - How services use the database package
- [Development Setup](../../docs/development/setup.md) - Local development with database