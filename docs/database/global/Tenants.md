# Tenants Table Documentation

## Purpose

Core registry table for multi-tenant CRM architecture. Maps organization subdomains to isolated database schemas.

## Table Structure

```sql
-- Core tenant registry table
CREATE TABLE tenants (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    subdomain VARCHAR(63) UNIQUE NOT NULL,
    schema_name VARCHAR(63) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- User invitations (global table for cross-tenant invites)
CREATE TABLE invitations (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER REFERENCES tenants(id) ON DELETE CASCADE,
    email VARCHAR(254) NOT NULL,
    role VARCHAR(20) CHECK (role IN ('admin', 'manager', 'sales_rep', 'viewer')) NOT NULL,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    accepted_at TIMESTAMPTZ,
    invited_by INTEGER, -- References users table in tenant schema
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Performance indexes (beyond automatic UNIQUE indexes)
CREATE INDEX idx_tenants_created_at ON tenants (created_at);
CREATE INDEX idx_invitations_token ON invitations (token);
CREATE INDEX idx_invitations_tenant_id ON invitations (tenant_id);
CREATE INDEX idx_invitations_email ON invitations (email);
CREATE INDEX idx_invitations_expires_at ON invitations (expires_at);

-- Unique constraint for active invitations
CREATE UNIQUE INDEX idx_invitations_tenant_email_active 
ON invitations (tenant_id, email) 
WHERE accepted_at IS NULL AND expires_at > CURRENT_TIMESTAMP;
```

## Column Design Decisions
- **subdomain**: 63-char limit for DNS compliance, enables `example.yourcrm.com` URLs
- **schema_name**: Uses `tenant_` prefix pattern (e.g., `tenant_example`) to avoid PostgreSQL conflicts
- **TIMESTAMPTZ**: Timezone-aware timestamps for global usage

## Indexes
- Automatic indexes created by UNIQUE constraints on `subdomain` and `schema_name`
- Subdomain index critical for fast tenant lookup on every request

## Key Operations
```sql
-- Primary lookup (happens on every request)
SELECT schema_name FROM tenants WHERE subdomain = 'example';

-- Insert new tenant
INSERT INTO tenants (name, subdomain, schema_name) 
VALUES ('Example Corp', 'example', 'tenant_example');

-- Performance check
EXPLAIN SELECT * FROM tenants WHERE subdomain = 'test';
```

## Application Flow
1. Extract subdomain from URL (`example.yourcrm.com` → `example`)
2. Lookup schema: `SELECT schema_name FROM tenants WHERE subdomain = ?`
3. Set search path: `SET search_path = tenant_example`
4. All queries now isolated to that tenant's data

## Constraints Tested
- NOT NULL validation
- UNIQUE subdomain enforcement
- UNIQUE schema_name enforcement
- Performance via index usage verification