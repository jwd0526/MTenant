package helpers

import (
	"context"
	"sync"
	"testing"
	"time"

	"crm-platform/deal-service/database"
	"crm-platform/deal-service/tenant"

	"github.com/stretchr/testify/require"
)

// TestDatabase manages multi-tenant database operations for testing
type TestDatabase struct {
	Pool          *database.Pool
	TenantPool    *tenant.TenantPool
	Config        *database.Config
	activeTenants map[string]context.Context
	mu            sync.RWMutex
	t             *testing.T
}

// TenantTx represents a tenant-specific transaction for test isolation
type TenantTx struct {
	*tenant.TenantTx
	TenantID string
}

// SetupTestDatabase creates a shared database instance for all tests
func SetupTestDatabase(t *testing.T) *TestDatabase {
	config, err := database.LoadConfigFromEnv()
	require.NoError(t, err, "Failed to load database config for tests")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := database.NewPool(ctx, config)
	require.NoError(t, err, "Failed to create test database pool")

	health := pool.HealthCheck(ctx)
	require.True(t, health.Healthy, "Test database is not healthy: %s", health.Error)

	tenantPool := tenant.NewTenantPool(pool)

	return &TestDatabase{
		Pool:          pool,
		TenantPool:    tenantPool,
		Config:        config,
		activeTenants: make(map[string]context.Context),
		t:             t,
	}
}

// Close cleans up the test database
func (td *TestDatabase) Close() {
	if td.Pool != nil {
		td.Pool.Close()
	}
}

// UsePredefinedTenant sets up a tenant context for a pre-existing test tenant
// Test tenants should be created by the setup_test_tenants.go script
func (td *TestDatabase) UsePredefinedTenant(tenantID string) {
	td.mu.Lock()
	defer td.mu.Unlock()

	ctx := context.Background()
	tenantCtx, err := tenant.NewContext(ctx, tenantID)
	require.NoError(td.t, err, "Failed to create tenant context for %s", tenantID)

	// Verify the tenant schema exists
	schemaName := tenant.GenerateSchemaName(tenantID)
	exists, err := tenant.SchemaExists(ctx, td.Pool, schemaName)
	require.NoError(td.t, err, "Failed to check if tenant schema exists: %s", schemaName)
	require.True(td.t, exists, "Test tenant schema %s does not exist. Run: go run tests/setup_test_tenants.go setup", schemaName)

	// Store tenant context for reuse
	td.activeTenants[tenantID] = tenantCtx
	
	td.t.Logf("Using predefined test tenant: %s", schemaName)
}

// GetTenantContext returns the context for a tenant (must be created first)
func (td *TestDatabase) GetTenantContext(tenantID string) context.Context {
	td.mu.RLock()
	defer td.mu.RUnlock()
	
	ctx, exists := td.activeTenants[tenantID]
	require.True(td.t, exists, "Tenant %s not found - call CreateTenantSchema first", tenantID)
	return ctx
}

// BeginTenantTx starts a transaction for test isolation within a tenant
// This is called for each test to provide perfect isolation via rollback
func (td *TestDatabase) BeginTenantTx(tenantID string) *TenantTx {
	tenantCtx := td.GetTenantContext(tenantID)
	
	tx, err := td.TenantPool.Begin(tenantCtx)
	require.NoError(td.t, err, "Failed to begin transaction for tenant %s", tenantID)

	return &TenantTx{
		TenantTx: tx,
		TenantID: tenantID,
	}
}

// Rollback rolls back the test transaction (automatic test cleanup)
func (tx *TenantTx) Rollback() {
	if tx.TenantTx != nil {
		_ = tx.TenantTx.Rollback(context.Background())
	}
}

// CleanTenantData removes all deals from a tenant (for test isolation)
func (td *TestDatabase) CleanTenantData(tenantID string) error {
	tenantCtx := td.GetTenantContext(tenantID)
	_, err := td.TenantPool.Exec(tenantCtx, "DELETE FROM deals")
	if err != nil {
		return err
	}
	// Also clean related tables if they exist (ignore errors for missing tables)
	_, _ = td.TenantPool.Exec(tenantCtx, "DELETE FROM deal_contacts")
	return nil
}

// IsHealthy checks if the database connection is healthy
func (td *TestDatabase) IsHealthy() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return td.Pool.IsHealthy(ctx)
}