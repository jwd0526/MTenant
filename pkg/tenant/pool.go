package tenant

import (
    "context"
    "fmt"

    "crm-platform/pkg/database"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgconn"
)

// Pool error definitions
var (
    ErrFailedTransaction = fmt.Errorf("failed to begin transaction")
    ErrSchemaWrapFailure = fmt.Errorf("failed to wrap SQL with tenant schema")
)

// TenantPool wraps database.Pool with tenant-aware functionality
type TenantPool struct {
    *database.Pool
}

// TenantTx wraps pgx.Tx with tenant context already set
type TenantTx struct {
    pgx.Tx
    schemaName string
}

// NewTenantPool creates a new tenant-aware database pool
func NewTenantPool(pool *database.Pool) *TenantPool {
    return &TenantPool{
        Pool: pool,
    }
}

// Query executes a query with tenant context isolation
func (tp *TenantPool) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
    // SECURITY: Explicit tenant validation - fail fast if no tenant context
    if !HasTenant(ctx) {
        return nil, fmt.Errorf("SECURITY VIOLATION: database operation attempted without tenant context")
    }
    
    // Always set search path before each query to handle connection reuse
    if err := tp.ensureTenantSearchPath(ctx); err != nil {
        return nil, fmt.Errorf("SECURITY: tenant isolation failed: %w", err)
    }
    
    return tp.Pool.Query(ctx, sql, args...)
}

// QueryRow executes a query that returns a single row with tenant context
func (tp *TenantPool) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
    // SECURITY: Explicit tenant validation - fail fast if no tenant context
    if !HasTenant(ctx) {
        return &errorRow{err: fmt.Errorf("SECURITY VIOLATION: database operation attempted without tenant context")}
    }
    
    // Always set search path before each query to handle connection reuse
    if err := tp.ensureTenantSearchPath(ctx); err != nil {
        return &errorRow{err: fmt.Errorf("SECURITY: tenant isolation failed: %w", err)}
    }
    
    return tp.Pool.QueryRow(ctx, sql, args...)
}

// Exec executes a command with tenant context isolation
func (tp *TenantPool) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
    // SECURITY: Explicit tenant validation - fail fast if no tenant context
    if !HasTenant(ctx) {
        return pgconn.CommandTag{}, fmt.Errorf("SECURITY VIOLATION: database operation attempted without tenant context")
    }
    
    // Set search path first
    if err := tp.ensureTenantSearchPath(ctx); err != nil {
        return pgconn.CommandTag{}, fmt.Errorf("SECURITY: tenant isolation failed: %w", err)
    }
    
    return tp.Pool.Exec(ctx, sql, args...)
}

// Begin starts a tenant-aware transaction
func (tp *TenantPool) Begin(ctx context.Context) (*TenantTx, error) {
    // SECURITY: Explicit tenant validation - fail fast if no tenant context
    if !HasTenant(ctx) {
        return nil, fmt.Errorf("SECURITY VIOLATION: database operation attempted without tenant context")
    }
    
    schemaName, err := ExtractTenantSchema(ctx)
    if err != nil {
        return nil, fmt.Errorf("SECURITY: tenant isolation failed: %w", err)
    }
    
    tx, err := tp.Pool.Begin(ctx)
    if err != nil {
        return nil, fmt.Errorf("%w: %v", ErrFailedTransaction, err)
    }
    
    // Set search path for the transaction
    err = SetSearchPath(ctx, tx, schemaName)
    if err != nil {
        tx.Rollback(ctx)
        return nil, fmt.Errorf("failed to set search path in transaction: %w", err)
    }
    
    return &TenantTx{
        Tx:         tx,
        schemaName: schemaName,
    }, nil
}

// Query executes a query within the tenant transaction
func (tt *TenantTx) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
    return tt.Tx.Query(ctx, sql, args...)
}

// QueryRow executes a query that returns a single row within the tenant transaction
func (tt *TenantTx) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
    return tt.Tx.QueryRow(ctx, sql, args...)
}

// Exec executes a command within the tenant transaction
func (tt *TenantTx) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
    return tt.Tx.Exec(ctx, sql, args...)
}

// Close closes the underlying database pool
func (tp *TenantPool) Close() {
    tp.Pool.Close()
}

// IsHealthy checks if the underlying pool is healthy
func (tp *TenantPool) IsHealthy(ctx context.Context) bool {
    return tp.Pool.IsHealthy(ctx)
}

// HealthCheck performs a health check on the underlying pool
func (tp *TenantPool) HealthCheck(ctx context.Context) *database.HealthStatus {
    return tp.Pool.HealthCheck(ctx)
}

// ensureTenantSearchPath sets the search path for the current connection
func (tp *TenantPool) ensureTenantSearchPath(ctx context.Context) error {
    schemaName, err := ExtractTenantSchema(ctx)
    if err != nil {
        return fmt.Errorf("%w: %v", ErrSchemaWrapFailure, err)
    }
    
    return SetSearchPath(ctx, tp.Pool, schemaName)
}

// errorRow implements pgx.Row for error cases
type errorRow struct {
    err error
}

// Scan returns the stored error
func (r *errorRow) Scan(dest ...interface{}) error {
    return r.err
}