# Tenant Package

A comprehensive multi-tenancy implementation for PostgreSQL using schema-per-tenant isolation. This package provides complete tenant isolation while maintaining high performance and security.

## 🏗️ Architecture Overview

The tenant package implements **schema-per-tenant** multi-tenancy, where each tenant gets their own PostgreSQL schema (e.g., `tenant_12345678-1234-1234-1234-123456789012`). This approach provides:

- **Complete Data Isolation**: No shared tables, complete tenant separation
- **Security**: Impossible to accidentally access other tenant's data
- **Scalability**: Each tenant can have custom schema modifications if needed
- **Performance**: No additional WHERE clauses needed for tenant filtering

## 📁 Package Structure

```
tenant/
├── context.go      # Tenant context management and validation
├── pool.go         # Tenant-aware database connection pooling
├── schema.go       # Schema creation, copying, and management
└── README.md       # This documentation
```

## 🔧 Core Components

### 1. Tenant Context (`context.go`)

Manages tenant information in Go contexts with proper validation.

```go
// Create tenant-aware context
ctx, err := tenant.NewContext(ctx, "12345678-1234-1234-1234-123456789012")
if err != nil {
    return err
}

// Extract tenant ID from context
tenantID, err := tenant.FromContext(ctx)
if err != nil {
    return err
}

// Check if context has tenant
if !tenant.HasTenant(ctx) {
    return errors.New("no tenant in context")
}
```

**Features:**
- UUID validation for tenant IDs
- Type-safe context key management
- Panic-safe tenant extraction with `MustFromContext`

### 2. Tenant Pool (`pool.go`)

Provides tenant-aware database operations with automatic schema isolation.

```go
// Create tenant pool
tenantPool := tenant.NewTenantPool(databasePool)

// All queries automatically use tenant's schema
rows, err := tenantPool.Query(ctx, "SELECT * FROM deals")
if err != nil {
    return err
}
```

**Key Features:**
- **Automatic Search Path**: Sets `search_path` to tenant schema automatically
- **Transaction Support**: Tenant-aware transactions with proper cleanup
- **Error Handling**: Graceful handling of schema-related errors
- **Health Checks**: Tenant-aware health monitoring

**Transaction Example:**
```go
tx, err := tenantPool.Begin(ctx)
if err != nil {
    return err
}
defer tx.Rollback(ctx)

// All operations within transaction use tenant schema
_, err = tx.Exec(ctx, "INSERT INTO deals (...) VALUES (...)")
if err != nil {
    return err
}

return tx.Commit(ctx)
```

### 3. Schema Management (`schema.go`)

Handles creation, copying, and management of tenant schemas.

```go
// Create new tenant schema
schemaName := tenant.GenerateSchemaName(tenantID)
err := tenant.CreateSchema(ctx, pool, schemaName)
if err != nil {
    return err
}

// Copy template schema to new tenant
err = tenant.CopyTemplateSchema(ctx, pool, "tenant_template", schemaName)
if err != nil {
    return err
}
```

**Schema Operations:**
- **Template Copying**: Copy table structures from template schema
- **Seed Data**: Copy initial data for lookup tables
- **Validation**: PostgreSQL identifier validation
- **Idempotent Operations**: Safe to run multiple times

## 🚀 Usage Examples

### Basic Setup

```go
package main

import (
    "context"
    "crm-platform/deal-service/database"
    "crm-platform/deal-service/tenant"
)

func main() {
    // 1. Create database pool
    config, err := database.LoadConfigFromEnv()
    if err != nil {
        log.Fatal(err)
    }
    
    pool, err := database.NewPool(context.Background(), config)
    if err != nil {
        log.Fatal(err)
    }
    defer pool.Close()
    
    // 2. Create tenant pool
    tenantPool := tenant.NewTenantPool(pool)
    
    // 3. Create tenant context
    ctx, err := tenant.NewContext(context.Background(), "12345678-1234-1234-1234-123456789012")
    if err != nil {
        log.Fatal(err)
    }
    
    // 4. All database operations are now tenant-scoped
    rows, err := tenantPool.Query(ctx, "SELECT * FROM deals")
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()
}
```

### HTTP Middleware Integration

```go
func TenantMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Extract tenant ID from JWT or header
        tenantID := extractTenantFromJWT(c)
        
        // Create tenant context
        tenantCtx, err := tenant.NewContext(c.Request.Context(), tenantID)
        if err != nil {
            c.JSON(400, gin.H{"error": "Invalid tenant"})
            c.Abort()
            return
        }
        
        // Replace request context
        c.Request = c.Request.WithContext(tenantCtx)
        c.Next()
    }
}
```

### Creating New Tenant

```go
func CreateTenant(ctx context.Context, pool *database.Pool, tenantID string) error {
    // 1. Generate schema name
    schemaName := tenant.GenerateSchemaName(tenantID)
    
    // 2. Create schema
    if err := tenant.CreateSchema(ctx, pool, schemaName); err != nil {
        return fmt.Errorf("failed to create schema: %w", err)
    }
    
    // 3. Copy template (if exists)
    templateExists, err := tenant.SchemaExists(ctx, pool, "tenant_template")
    if err != nil {
        return fmt.Errorf("failed to check template: %w", err)
    }
    
    if templateExists {
        err = tenant.CopyTemplateSchema(ctx, pool, "tenant_template", schemaName)
        if err != nil {
            return fmt.Errorf("failed to copy template: %w", err)
        }
    }
    
    return nil
}
```

## 🔒 Security Features

### 1. Tenant ID Validation

```go
// Only valid UUIDs are accepted as tenant IDs
var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func validateID(id string) error {
    if id == "" {
        return ErrBlankTenantID
    }
    
    if !uuidPattern.MatchString(id) {
        return ErrInvalidTenantID
    }
    
    return nil
}
```

### 2. SQL Injection Protection

```go
// All schema names are properly quoted and validated
sql := fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS "%s"`, schemaName)

// Search path uses quoted identifiers
sql := fmt.Sprintf(`SET search_path TO "%s", public`, schemaName)
```

### 3. Automatic Isolation

- **No Cross-Tenant Access**: Impossible to query other tenant's data
- **Context Validation**: All operations require valid tenant context
- **Schema Validation**: PostgreSQL identifier validation prevents injection

## 📊 Performance Considerations

### 1. Connection Pooling

```go
// Tenant pool reuses underlying connection pool
type TenantPool struct {
    *database.Pool  // Embeds database pool for efficiency
}

// Each query sets search path once per connection
func (tp *TenantPool) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
    if err := tp.ensureTenantSearchPath(ctx); err != nil {
        return nil, err
    }
    return tp.Pool.Query(ctx, sql, args...)
}
```

### 2. Search Path Optimization

- Search path is set once per connection
- Subsequent queries in same connection reuse the path
- Minimal overhead for tenant isolation

### 3. Schema Template Caching

```go
// Template existence is checked once
templateExists, err := tenant.SchemaExists(ctx, pool, "tenant_template")
if err == nil && templateExists {
    // Copy template efficiently
    err = tenant.CopyTemplateSchema(ctx, pool, "tenant_template", schemaName)
}
```

## 🧪 Testing Support

### Test Tenant Creation

```go
// Helper for test tenant creation
func CreateTestTenant(t *testing.T, pool *database.Pool, tenantID string) context.Context {
    ctx := context.Background()
    
    tenantCtx, err := tenant.NewContext(ctx, tenantID)
    require.NoError(t, err)
    
    schemaName := tenant.GenerateSchemaName(tenantID)
    err = tenant.CreateSchema(ctx, pool, schemaName)
    require.NoError(t, err)
    
    return tenantCtx
}
```

### Test Cleanup

```go
// Clean up test tenant
func CleanupTestTenant(t *testing.T, pool *database.Pool, tenantID string) {
    ctx := context.Background()
    schemaName := tenant.GenerateSchemaName(tenantID)
    
    _, err := pool.Exec(ctx, fmt.Sprintf(`DROP SCHEMA IF EXISTS "%s" CASCADE`, schemaName))
    if err != nil {
        t.Logf("Warning: Failed to cleanup schema %s: %v", schemaName, err)
    }
}
```

## 🚨 Error Handling

### Common Errors

```go
var (
    ErrTenantNotFound      = fmt.Errorf("context does not contain tenant key")
    ErrBlankTenantID       = fmt.Errorf("tenant ID cannot be blank")
    ErrInvalidTenantID     = fmt.Errorf("tenant ID must be a valid UUID")
    ErrSchemaNotFound      = fmt.Errorf("schema does not exist")
    ErrInvalidSchema       = fmt.Errorf("schema name invalid")
    ErrFailedTransaction   = fmt.Errorf("failed to begin transaction")
    ErrSetSearchPathFailure = fmt.Errorf("failed to set search path")
)
```

### Error Handling Patterns

```go
// Graceful error handling with context
tenantID, err := tenant.FromContext(ctx)
switch {
case errors.Is(err, tenant.ErrTenantNotFound):
    return errors.New("authentication required")
case errors.Is(err, tenant.ErrInvalidTenantID):
    return errors.New("invalid tenant format")
default:
    return err
}
```

## 🔧 Configuration

### Environment Setup

The tenant package uses the same database configuration as the main application:

```bash
# Required for tenant operations
export DATABASE_URL="postgres://user:pass@localhost:5432/dbname?sslmode=disable"

# Template schema name (optional, defaults to "tenant_template")
export TENANT_TEMPLATE_SCHEMA="tenant_template"
```

### Pool Configuration

```go
// Configure connection pool for multi-tenant usage
config := &database.Config{
    MaxConns:        20,    // Higher for multi-tenant load
    MinConns:        5,     // Maintain minimum connections
    MaxConnLifetime: time.Hour,
    MaxConnIdleTime: time.Minute * 5,
}
```

## 🚀 Advanced Usage

### Custom Schema Operations

```go
// Add custom tables to specific tenant
func AddCustomTable(ctx context.Context, pool *tenant.TenantPool, tableDDL string) error {
    _, err := pool.Exec(ctx, tableDDL)
    return err
}

// Migrate specific tenant schema
func MigrateTenant(ctx context.Context, pool *tenant.TenantPool, migrationSQL string) error {
    tx, err := pool.Begin(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)
    
    _, err = tx.Exec(ctx, migrationSQL)
    if err != nil {
        return err
    }
    
    return tx.Commit(ctx)
}
```

### Tenant Statistics

```go
// Get tenant schema statistics
func GetTenantStats(ctx context.Context, pool *database.Pool, tenantID string) (*TenantStats, error) {
    schemaName := tenant.GenerateSchemaName(tenantID)
    
    query := `
        SELECT 
            schemaname,
            COUNT(*) as table_count,
            pg_size_pretty(SUM(pg_total_relation_size(schemaname||'.'||tablename))) as total_size
        FROM pg_tables 
        WHERE schemaname = $1
        GROUP BY schemaname
    `
    
    var stats TenantStats
    err := pool.QueryRow(ctx, query, schemaName).Scan(
        &stats.SchemaName,
        &stats.TableCount,
        &stats.TotalSize,
    )
    
    return &stats, err
}
```

## 📚 Best Practices

### 1. **Always Use Context**
```go
// ✅ Good
ctx, err := tenant.NewContext(parentCtx, tenantID)
rows, err := tenantPool.Query(ctx, sql, args...)

// ❌ Bad
rows, err := pool.Query(context.Background(), sql, args...)
```

### 2. **Validate Tenant IDs Early**
```go
// ✅ Good - validate at API boundary
func (h *Handler) CreateDeal(c *gin.Context) {
    tenantID := extractTenantFromJWT(c)
    ctx, err := tenant.NewContext(c.Request.Context(), tenantID)
    if err != nil {
        c.JSON(400, gin.H{"error": "Invalid tenant"})
        return
    }
    // ... rest of handler
}
```

### 3. **Use Transactions for Related Operations**
```go
// ✅ Good - group related operations
tx, err := tenantPool.Begin(ctx)
if err != nil {
    return err
}
defer tx.Rollback(ctx)

// Multiple related operations
_, err = tx.Exec(ctx, "INSERT INTO deals ...")
_, err = tx.Exec(ctx, "INSERT INTO deal_contacts ...")

return tx.Commit(ctx)
```

### 4. **Handle Schema Creation Idempotently**
```go
// ✅ Good - safe to run multiple times
err := tenant.CreateSchema(ctx, pool, schemaName)
if err != nil && !isSchemaExistsError(err) {
    return err
}
```

## 🔍 Debugging

### Enable Query Logging

```go
// Log all queries for debugging
config.LogLevel = pgxpool.LogLevelDebug
```

### Check Current Search Path

```sql
-- Verify search path is set correctly
SHOW search_path;

-- Should return: "tenant_<uuid>", public
```

### Verify Schema Exists

```sql
-- Check if tenant schema exists
SELECT schema_name 
FROM information_schema.schemata 
WHERE schema_name = 'tenant_12345678-1234-1234-1234-123456789012';
```

---

This tenant package provides a robust foundation for multi-tenant applications with PostgreSQL, ensuring complete data isolation while maintaining performance and security.