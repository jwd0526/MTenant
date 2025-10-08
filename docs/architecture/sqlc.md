# SQLC Implementation Guide

**Last Updated:** 2025-10-08\
*Type-safe database access patterns*

Comprehensive documentation for type-safe database access using SQLC across all microservices.

## Overview

SQLC generates type-safe Go code from SQL queries, eliminating runtime errors and providing compile-time validation. Each service uses SQLC to interact with PostgreSQL in a tenant-aware manner.

## Configuration

### Standard sqlc.yaml Structure

All services follow this standardized configuration:

```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "./db/queries"
    schema: "./db/schema"
    gen:
      go:
        package: "db"
        out: "./internal/db"
        sql_package: "pgx/v5"
        emit_json_tags: true
        emit_prepared_queries: false
        emit_interface: true
        emit_exact_table_names: false
        emit_empty_slices: true
        emit_exported_queries: false
        emit_result_struct_pointers: false
        emit_params_struct_pointers: false
        emit_methods_with_db_argument: false
        emit_pointers_for_null_types: true
        emit_enum_valid_method: false
        emit_all_enum_values: false
```

### Service-Specific Overrides

#### Auth Service (users, password_reset_tokens)

```yaml
overrides:
  - column: "*.created_at"
    go_type: "time.Time"
  - column: "*.updated_at"  
    go_type: "time.Time"
  - column: "*.deleted_at"
    go_type: "database/sql.NullTime"
  - column: "*.last_login"
    go_type: "database/sql.NullTime"
  - column: "*.expires_at"
    go_type: "time.Time"
  - column: "*.used_at"
    go_type: "database/sql.NullTime"
  - column: "users.permissions"
    go_type: "encoding/json.RawMessage"
```

#### Tenant Service (tenants, invitations)

```yaml
overrides:
  - column: "*.created_at"
    go_type: "time.Time"
  - column: "*.updated_at"
    go_type: "time.Time"
  - column: "*.expires_at"
    go_type: "time.Time"
  - column: "*.accepted_at"
    go_type: "database/sql.NullTime"
  - column: "invitations.metadata"
    go_type: "encoding/json.RawMessage"
```

#### Contact Service (contacts, companies)

```yaml
overrides:
  - column: "*.created_at"
    go_type: "time.Time"
  - column: "*.updated_at"
    go_type: "time.Time"
  - column: "*.deleted_at"
    go_type: "database/sql.NullTime"
  - db_type: "jsonb"
    go_type: "encoding/json.RawMessage"
  - db_type: "numeric"
    go_type: "github.com/shopspring/decimal.Decimal"
```

#### Deal Service (deals, deal_contacts)

```yaml
overrides:
  - column: "*.created_at"
    go_type: "time.Time"
  - column: "*.updated_at"
    go_type: "time.Time"
  - column: "*.expected_close_date"
    go_type: "database/sql.NullTime"
  - column: "*.actual_close_date"
    go_type: "database/sql.NullTime"
  - db_type: "numeric"
    go_type: "github.com/shopspring/decimal.Decimal"
```

## Generated Code Structure

### File Organization

Each service generates these files in `internal/db/`:

```
internal/db/
├── db.go              # Database connection interface
├── models.go          # Struct definitions for tables
├── querier.go         # Interface for all queries
├── {table}.sql.go     # Generated query methods per table
```

### Example: Auth Service Generated Models

```go
// internal/db/models.go
type User struct {
    ID            int32           `json:"id"`
    Email         string          `json:"email"`
    PasswordHash  string          `json:"password_hash"`
    FirstName     string          `json:"first_name"`
    LastName      string          `json:"last_name"`
    Role          string          `json:"role"`
    Permissions   json.RawMessage `json:"permissions"`
    Active        *bool           `json:"active"`
    EmailVerified *bool           `json:"email_verified"`
    LastLogin     sql.NullTime    `json:"last_login"`
    CreatedBy     *int32          `json:"created_by"`
    UpdatedBy     *int32          `json:"updated_by"`
    CreatedAt     time.Time       `json:"created_at"`
    UpdatedAt     time.Time       `json:"updated_at"`
    DeletedAt     sql.NullTime    `json:"deleted_at"`
}

type PasswordResetToken struct {
    ID        int32        `json:"id"`
    UserID    *int32       `json:"user_id"`
    Token     string       `json:"token"`
    ExpiresAt time.Time    `json:"expires_at"`
    UsedAt    sql.NullTime `json:"used_at"`
    CreatedAt time.Time    `json:"created_at"`
}
```

### Example: Generated Query Interface

```go
// internal/db/querier.go
type Querier interface {
    // User management
    CreateUser(ctx context.Context, arg CreateUserParams) (CreateUserRow, error)
    GetUserByEmail(ctx context.Context, email string) (GetUserByEmailRow, error)
    GetUserByID(ctx context.Context, id int32) (GetUserByIDRow, error)
    UpdateUserLastLogin(ctx context.Context, id int32) error
    VerifyUserEmail(ctx context.Context, arg VerifyUserEmailParams) error
    
    // Password reset tokens
    CreatePasswordResetToken(ctx context.Context, arg CreatePasswordResetTokenParams) (PasswordResetToken, error)
    GetPasswordResetToken(ctx context.Context, token string) (GetPasswordResetTokenRow, error)
    MarkPasswordResetTokenUsed(ctx context.Context, token string) error
}
```

## Query Patterns

### Naming Convention

All queries follow the pattern: `-- name: {Operation}{Entity} :{return_type}`

**Operation Types:**
- `Create` - INSERT operations
- `Get` - SELECT single row 
- `List` - SELECT multiple rows
- `Update` - UPDATE operations  
- `Delete` - DELETE operations
- `Check` - EXISTS queries
- `Count` - COUNT queries

**Return Types:**
- `:one` - Single row expected
- `:many` - Multiple rows expected
- `:exec` - No return value (INSERT/UPDATE/DELETE)

### Example Query Patterns

#### Auth Service Queries

```sql
-- name: CreateUser :one
INSERT INTO users (email, password_hash, first_name, last_name, role, created_by)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, email, first_name, last_name, role, active, email_verified, created_at;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, first_name, last_name, role, permissions, 
       active, email_verified, last_login, created_at, updated_at
FROM users
WHERE email = $1 AND active = true AND deleted_at IS NULL;

-- name: UpdateUserLastLogin :exec
UPDATE users 
SET last_login = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: CheckEmailExists :one
SELECT EXISTS(
    SELECT 1 FROM users 
    WHERE email = $1 AND deleted_at IS NULL
);
```

#### Contact Service Queries

```sql
-- name: CreateContact :one
INSERT INTO contacts (
    first_name, last_name, email, phone, company_id, custom_fields, created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: SearchContactsFullText :many
SELECT c.*, comp.name as company_name 
FROM contacts c
LEFT JOIN companies comp ON c.company_id = comp.id AND comp.deleted_at IS NULL
WHERE c.deleted_at IS NULL 
  AND to_tsvector('english', c.first_name || ' ' || c.last_name || ' ' || COALESCE(c.email, '')) 
      @@ to_tsquery('english', $1)
ORDER BY c.last_name, c.first_name
LIMIT $2 OFFSET $3;

-- name: FilterContacts :many
SELECT c.*, comp.name as company_name 
FROM contacts c
LEFT JOIN companies comp ON c.company_id = comp.id AND comp.deleted_at IS NULL
WHERE c.deleted_at IS NULL
  AND ($1::int IS NULL OR c.company_id = $1)
  AND ($2::text IS NULL OR c.custom_fields->>'status' = $2)
ORDER BY c.last_name, c.first_name
LIMIT $3 OFFSET $4;
```

## Code Generation Process

### Running SQLC

```bash
# Generate code for specific service
cd services/auth-service && sqlc generate
cd services/tenant-service && sqlc generate  
cd services/contact-service && sqlc generate
cd services/deal-service && sqlc generate

# Generate for all services (from project root)
find services -name "sqlc.yaml" -execdir sqlc generate \;
```

### Generated Method Examples

#### Parameters Struct

```go
type CreateUserParams struct {
    Email        string `json:"email"`
    PasswordHash string `json:"password_hash"`
    FirstName    string `json:"first_name"`
    LastName     string `json:"last_name"`
    Role         string `json:"role"`
    CreatedBy    *int32 `json:"created_by"`
}
```

#### Query Implementation

```go
func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (CreateUserRow, error) {
    row := q.db.QueryRow(ctx, createUser,
        arg.Email,
        arg.PasswordHash,
        arg.FirstName,
        arg.LastName,
        arg.Role,
        arg.CreatedBy,
    )
    var i CreateUserRow
    err := row.Scan(
        &i.ID,
        &i.Email,
        &i.FirstName,
        &i.LastName,
        &i.Role,
        &i.Active,
        &i.EmailVerified,
        &i.CreatedAt,
    )
    return i, err
}
```

## Usage in Services

### Database Connection Setup

Services now use the shared `pkg/database` package for standardized connection management:

```go
package main

import (
    "context"
    "log"
    
    "crm-platform/pkg/database"
    "crm-platform/auth-service/internal/db"
)

func main() {
    ctx := context.Background()
    
    // Load database configuration from environment
    dbConfig, err := database.LoadConfigFromEnv()
    if err != nil {
        log.Fatal("Failed to load database config:", err)
    }
    
    // Create connection pool with retry logic and health monitoring
    dbPool, err := database.NewPool(ctx, dbConfig)
    if err != nil {
        log.Fatal("Failed to create database pool:", err)
    }
    defer dbPool.Close()
    
    // Verify database health on startup
    health := dbPool.HealthCheck(ctx)
    if !health.Healthy {
        log.Fatal("Database health check failed:", health.Error)
    }
    log.Printf("Database connected in %v", health.ResponseTime)
    
    // Create SQLC queries instance with shared connection pool
    queries := db.New(dbPool)
    
    // Use generated methods with enhanced connection management
    user, err := queries.CreateUser(ctx, db.CreateUserParams{
        Email:        "user@example.com",
        PasswordHash: "$2a$10$...",
        FirstName:    "John",
        LastName:     "Doe", 
        Role:         "sales_rep",
        CreatedBy:    nil,
    })
    if err != nil {
        log.Fatal("Failed to create user:", err)
    }
    
    log.Printf("Created user: %+v", user)
}
```

### Tenant-Aware Queries

```go
// Set tenant schema context using shared database pool
schemaName := "tenant_example"
_, err := dbPool.Exec(ctx, "SET search_path = $1", schemaName)
if err != nil {
    return err
}

// All subsequent queries operate within tenant schema
contacts, err := queries.ListContacts(ctx, db.ListContactsParams{
    Limit:  20,
    Offset: 0,
})

// Helper function for tenant context management
func setTenantContext(ctx context.Context, dbPool *database.Pool, tenantID string) error {
    schemaName := fmt.Sprintf("tenant_%s", tenantID)
    _, err := dbPool.Exec(ctx, "SET search_path = $1", schemaName)
    return err
}
```

## Best Practices

### Schema Changes

1. **Update schema files** in `db/schema/`
2. **Update queries** in `db/queries/` if needed
3. **Regenerate code:** `sqlc generate`
4. **Update Go imports** if new types added
5. **Test changes** thoroughly

### Query Design

**Efficient Queries:**
- Use indexes for WHERE clauses
- Limit result sets with LIMIT/OFFSET
- Use prepared statements (SQLC default)

**Tenant Isolation:**
- Always set search_path before queries
- Never hardcode schema names in SQL
- Validate tenant context in middleware

**Error Handling:**
- Check for `pgx.ErrNoRows` for missing records
- Handle constraint violations gracefully
- Log query errors with context

### Type Safety

**Leverage SQLC Types:**
```go
// Compile-time safety with shared database connection
var user db.User
user.Email = "test@example.com"  // string - OK
user.ID = "invalid"              // Compile error!

// Null handling
if user.LastLogin.Valid {
    fmt.Printf("Last login: %v", user.LastLogin.Time)
}

// Health monitoring integration
health := dbPool.HealthCheck(ctx)
if !health.Healthy {
    log.Printf("Database issue: %s", health.Error)
}
```

**Custom Field Handling:**
```go
// JSON fields
var permissions map[string]bool
err := json.Unmarshal(user.Permissions, &permissions)

// Custom fields in contacts
var customData map[string]interface{}
err := json.Unmarshal(contact.CustomFields, &customData)
```

## Troubleshooting

### Common Issues

**Schema/Query Mismatch:**
```bash
# Error: column doesn't exist
# Solution: Ensure schema files match actual database
sqlc generate  # Regenerate after schema changes
```

**Type Conversion Errors:**
```bash
# Error: cannot convert string to int32
# Solution: Check column overrides in sqlc.yaml
```

**Missing Query Methods:**
```bash
# Error: method not found
# Solution: Check query name and regenerate
sqlc generate
```

### Validation

```bash
# Validate SQLC configuration
sqlc compile

# Check generated code
go build ./internal/db
```

## Migration Integration

SQLC works seamlessly with migration tools:

1. **Create migration** with schema changes
2. **Update schema files** to match migration
3. **Update queries** if needed
4. **Regenerate SQLC code**
5. **Test integration**

## Performance Considerations

- **Connection Pooling:** Handled automatically by `pkg/database` with configurable limits
- **Query Caching:** SQLC generates prepared statements for optimal performance
- **Health Monitoring:** Built-in database health checks and metrics collection
- **Batch Operations:** Consider custom batch queries for bulk operations
- **Index Usage:** Ensure queries use appropriate indexes
- **Retry Logic:** Automatic connection retry with exponential backoff
- **Resource Monitoring:** Real-time connection pool statistics and query metrics

## Related Documentation

- [Shared Packages](./shared-packages.md) - Database package architecture and usage
- [Database Design](./database.md) - Multi-tenant schema architecture
- [Service Architecture](./services.md) - How services use SQLC with shared packages
- [Development Setup](../development/setup.md) - Local development with SQLC and shared packages