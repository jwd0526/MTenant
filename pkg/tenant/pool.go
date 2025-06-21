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
    wrappedSQL, err := tp.wrapWithTenantSchema(ctx, sql)
    if err != nil {
        return nil, err
    }
    
    return tp.Pool.Query(ctx, wrappedSQL, args...)
}

// QueryRow executes a query that returns a single row with tenant context
func (tp *TenantPool) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
    wrappedSQL, err := tp.wrapWithTenantSchema(ctx, sql)
    if err != nil {
        return &errorRow{err: err}
    }
    
    return tp.Pool.QueryRow(ctx, wrappedSQL, args...)
}

// Exec executes a command with tenant context isolation
func (tp *TenantPool) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
    wrappedSQL, err := tp.wrapWithTenantSchema(ctx, sql)
    if err != nil {
        return pgconn.CommandTag{}, err
    }
    
    return tp.Pool.Exec(ctx, wrappedSQL, args...)
}

// Begin starts a tenant-aware transaction
func (tp *TenantPool) Begin(ctx context.Context) (*TenantTx, error) {
    schemaName, err := ExtractTenantSchema(ctx)
    if err != nil {
        return nil, err
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

// wrapWithTenantSchema extracts tenant schema and wraps SQL with search path
func (tp *TenantPool) wrapWithTenantSchema(ctx context.Context, sql string) (string, error) {
    schemaName, err := ExtractTenantSchema(ctx)
    if err != nil {
        return "", fmt.Errorf("%w: %v", ErrSchemaWrapFailure, err)
    }
    
    return fmt.Sprintf(`SET search_path TO "%s", public; %s`, schemaName, sql), nil
}

// errorRow implements pgx.Row for error cases
type errorRow struct {
    err error
}

// Scan returns the stored error
func (r *errorRow) Scan(dest ...interface{}) error {
    return r.err
}