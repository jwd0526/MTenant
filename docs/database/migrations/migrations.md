# Database Migrations Guide

Complete guide to database migrations for the multi-tenant CRM platform using golang-migrate.

## Overview

Database migrations provide version-controlled, systematic database schema evolution. In the MTenant CRM, migrations manage both global tables (tenant registry) and tenant-specific schema templates.

## Migration Architecture

### Multi-Tenant Migration Strategy

The MTenant CRM uses a two-tier migration approach:

```
Global Migrations (public schema):
├── tenant registry
├── cross-tenant invitations
└── system metadata

Tenant Schema Migrations:
├── template schema updates
├── applied to all existing tenants
└── automatic for new tenants
```

### Migration Tool: golang-migrate

**Installation:**
```bash
# Install golang-migrate tool
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Verify installation
migrate -version
```

**Database Connection:**
```bash
export DATABASE_URL="postgresql://admin:admin@localhost:5433/crm-platform?sslmode=disable"
```

## Project Migration Structure

### Directory Organization

```
MTenant/
├── migrations/
│   ├── global/                    # Global schema migrations
│   │   ├── 000001_create_tenants.up.sql
│   │   └── 000001_create_tenants.down.sql
│   │   
│   └── tenant-template/           # Tenant schema template migrations
│
└── scripts/

```

## Global Schema Migrations

### Creating Global Migrations

**Create Migration Files:**
```bash
# Navigate to global migrations
cd migrations/global

# Create new global migration
migrate create -ext sql -dir . -seq create_tenants_table
```

**Example: Tenant Registry Migration**

**File: `000001_create_tenants.up.sql`**
```sql
-- Create tenant registry table
-- Create tenants table for multi-tenant registry
CREATE TABLE tenants (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    subdomain VARCHAR(63) UNIQUE NOT NULL,
    schema_name VARCHAR(63) UNIQUE NOT NULL,
    status VARCHAR(50) DEFAULT 'active' NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints for data validation
    CONSTRAINT tenants_subdomain_format CHECK (subdomain ~ '^[a-z0-9-]+$'),
    CONSTRAINT tenants_schema_format CHECK (schema_name ~ '^tenant_[a-z0-9_]+$'),
    CONSTRAINT tenants_status_valid CHECK (status IN ('active', 'suspended', 'pending'))
);

-- Create indexes for performance
CREATE INDEX idx_tenants_subdomain ON tenants(subdomain);
CREATE INDEX idx_tenants_schema_name ON tenants(schema_name);
CREATE INDEX idx_tenants_status ON tenants(status);

-- Insert default system tenant for testing
INSERT INTO tenants (name, subdomain, schema_name) 
VALUES ('System Admin', 'admin', 'tenant_admin');
```

**File: `000001_create_tenants.down.sql`**
```sql
-- Remove indexes first
DROP INDEX IF EXISTS idx_tenants_subdomain;
DROP INDEX IF EXISTS idx_tenants_schema_name;
DROP INDEX IF EXISTS idx_tenants_status;

-- Remove tenants table and constraints
DROP TABLE IF EXISTS tenants;
```

## Migration Best Practices

### Multi-Tenant Considerations

**Schema Isolation:**
```sql
-- Always specify schema context in tenant migrations
SET search_path = template;

-- Use IF NOT EXISTS for safety
CREATE TABLE IF NOT EXISTS users (...);

-- Consider existing data
ALTER TABLE users ADD COLUMN phone VARCHAR(20) DEFAULT NULL;
```

**Performance at Scale:**
```sql
-- Create indexes concurrently for large tables
CREATE INDEX CONCURRENTLY idx_contacts_email ON contacts (email);

-- Use batched updates for data migrations
UPDATE contacts SET normalized_email = LOWER(email)
WHERE id BETWEEN 1 AND 10000;
```

### Testing Strategies

**Local Testing:**
```bash
# Test complete migration cycle
make reset-db

# Test rollback
migrate -database "$DATABASE_URL" -path "migrations" down 1
migrate -database "$DATABASE_URL" -path "migrations" up

# Check tables
docker exec -it crm-platform psql -U admin -d crm-platform -c "\dt"

# Check tenant data
docker exec -it crm-platform psql -U admin -d crm-platform -c "SELECT * FROM tenants;"

```

**Integration Testing:**
```bash
# Test with sample data
psql "$DATABASE_URL" -c "
INSERT INTO tenants (name, subdomain, schema_name) 
VALUES ('Test Corp', 'test', 'tenant_test');
"

# Create test tenant schema
psql "$DATABASE_URL" -c "CREATE SCHEMA tenant_test;"

# Apply tenant migrations
TENANT_URL="$DATABASE_URL&search_path=tenant_test"
migrate -database "$TENANT_URL" -path "migrations/tenant-template" up

# Verify structure
psql "$DATABASE_URL" -c "\dt tenant_test.*"
```

### Migration Safety Rules

**Development Phase:**
- Test both UP and DOWN migrations
- Use transactions for complex changes
- Keep migrations atomic and focused
- Document data impact in comments

**Production Phase:**
- Never edit existing migration files
- Test migrations on production copy first
- Plan downtime for breaking changes
- Have rollback plan ready

## Common Migration Patterns

### Adding Columns Safely

```sql
-- UP: Add nullable column with default
ALTER TABLE users ADD COLUMN timezone VARCHAR(50) DEFAULT 'UTC';

-- Later migration: Make required after data population
-- ALTER TABLE users ALTER COLUMN timezone SET NOT NULL;

-- DOWN: Remove column
ALTER TABLE users DROP COLUMN IF EXISTS timezone;
```

### Index Management

```sql
-- UP: Create index (use CONCURRENTLY in production)
CREATE INDEX CONCURRENTLY idx_deals_expected_close_date 
ON deals (expected_close_date) 
WHERE expected_close_date IS NOT NULL;

-- DOWN: Remove index
DROP INDEX IF EXISTS idx_deals_expected_close_date;
```

### Data Transformations

```sql
-- UP: Transform existing data
UPDATE contacts 
SET custom_fields = custom_fields || '{"migrated": true}'::jsonb
WHERE custom_fields IS NOT NULL;

-- DOWN: Reverse transformation (if possible)
UPDATE contacts 
SET custom_fields = custom_fields - 'migrated'
WHERE custom_fields ? 'migrated';
```

## Troubleshooting

### Common Issues

**Dirty Migration State:**
```bash
# Check status
migrate -database "$DATABASE_URL" -path "migrations/global" version

# Force clean state (use carefully)
migrate -database "$DATABASE_URL" -path "migrations/global" force 1
```

**Schema Not Found:**
```bash
# Verify schema exists
psql "$DATABASE_URL" -c "\dn"

# Create missing schema
psql "$DATABASE_URL" -c "CREATE SCHEMA tenant_example;"
```

**Permission Errors:**
```bash
# Check database permissions
psql "$DATABASE_URL" -c "
SELECT grantee, privilege_type 
FROM information_schema.role_table_grants 
WHERE table_name='schema_migrations';
"
```

### Recovery Procedures

**Failed Migration Recovery:**
```bash
# 1. Check what failed
migrate -database "$DATABASE_URL" -path "migrations/global" version

# 2. Fix manually if needed
psql "$DATABASE_URL" -c "-- Fix broken state"

# 3. Force clean and retry
migrate -database "$DATABASE_URL" -path "migrations/global" force <version>
migrate -database "$DATABASE_URL" -path "migrations/global" up
```

## Related Documentation

- [Database Design](./database.md) - Multi-tenant architecture overview
- [SQLC Configuration](../architecture/sqlc.md) - Code generation after migrations
- [Development Setup](../development/setup.md) - Local environment with migrations
- [Makefile Reference](../development/makefile.md) - Build system integration