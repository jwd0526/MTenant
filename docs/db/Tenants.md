# Tenants Table Documentation

## Purpose

Core registry table for multi-tenant CRM architecture. Maps organization subdomains to isolated database schemas.

## Table Structure

```sql
CREATE TABLE tenants (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    subdomain VARCHAR(63) UNIQUE NOT NULL,
    schema_name VARCHAR(63) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
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