# Database Design & Multi-Tenant Architecture

**Last Updated:** 2025-10-08\
*Multi-tenant schema architecture and data modeling*

Comprehensive documentation of the multi-tenant database architecture, schema isolation, and data modeling.

## Multi-Tenant Strategy

### Schema-Based Isolation

The platform uses PostgreSQL schemas to achieve complete tenant isolation:

```sql
-- Global registry (public schema)
public.tenants              -- Organization registry
public.invitations          -- Cross-tenant invitations

-- Tenant-specific schemas  
tenant_123.users           -- TenantCo users
tenant_123.contacts        -- TenantCo contacts
tenant_123.companies       -- TenantCo companies
tenant_123.deals           -- TenantCo deals
tenant_123.activities      -- TenantCo activities

tenant_456.users           -- AnotherCorp users  
tenant_456.contacts        -- AnotherCorp contacts
-- etc.
```

### Benefits of Schema Isolation

**Complete Data Separation:**
- No risk of cross-tenant data leakage
- Clear compliance boundaries for data protection
- Simplified backup and recovery per tenant

**Performance Isolation:**
- Tenant-specific indexing strategies
- Query optimization per tenant workload
- Resource usage isolation

**Scalability:**
- Distribute schemas across database servers
- Tenant-specific tuning and optimization
- Independent schema versioning

## Global Tables (Public Schema)

### Tenants Registry

Located in `public.tenants` - the core registry mapping organizations to schemas:

```sql
CREATE TABLE tenants (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    subdomain VARCHAR(63) UNIQUE NOT NULL,
    schema_name VARCHAR(63) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Essential indexes for tenant lookup
CREATE INDEX idx_tenants_created_at ON tenants (created_at);
```

**Key Operations:**
```sql
-- Primary lookup (happens on every request)
SELECT schema_name FROM tenants WHERE subdomain = 'example';

-- Tenant creation
INSERT INTO tenants (name, subdomain, schema_name) 
VALUES ('Example Corp', 'example', 'tenant_example');
```

### Cross-Tenant Invitations

Located in `public.invitations` - manages user invitations across tenant boundaries:

```sql
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

-- Prevent duplicate active invitations
CREATE UNIQUE INDEX idx_invitations_tenant_email_active 
ON invitations (tenant_id, email) 
WHERE accepted_at IS NULL AND expires_at > CURRENT_TIMESTAMP;
```

## Tenant Schema Templates

Each tenant schema contains identical table structures created from templates.

### User Management

**Users Table:**
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(254) UNIQUE NOT NULL,
    password_hash VARCHAR(60) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    role VARCHAR(20) CHECK (role IN ('admin', 'manager', 'sales_rep', 'viewer')) NOT NULL,
    permissions JSONB DEFAULT '{}',
    active BOOLEAN DEFAULT true,
    email_verified BOOLEAN DEFAULT false,
    last_login TIMESTAMPTZ,
    created_by INTEGER REFERENCES users(id),
    updated_by INTEGER REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

-- Performance indexes
CREATE INDEX idx_users_email ON users (email);
CREATE INDEX idx_users_active ON users (active);
CREATE INDEX idx_users_role ON users (role);
```

**Password Reset Tokens:**
```sql
CREATE TABLE password_reset_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_password_reset_tokens_token ON password_reset_tokens (token);
CREATE INDEX idx_password_reset_tokens_expires ON password_reset_tokens (expires_at);
```

### Contact Management

**Companies Table:**
```sql
CREATE TABLE companies (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    domain VARCHAR(255),
    industry VARCHAR(100),
    street VARCHAR(255),
    city VARCHAR(100),
    state VARCHAR(100),
    country VARCHAR(100),
    postal_code VARCHAR(20),
    parent_company_id INTEGER REFERENCES companies(id),
    custom_fields JSONB DEFAULT '{}',
    company_size VARCHAR(20),
    employee_count INTEGER,
    annual_revenue DECIMAL(15,2),
    revenue_currency VARCHAR(3) DEFAULT 'USD',
    created_by INTEGER REFERENCES users(id) NOT NULL,
    updated_by INTEGER REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

-- Performance indexes
CREATE INDEX idx_companies_domain ON companies (domain);
CREATE INDEX idx_companies_parent_id ON companies (parent_company_id);
CREATE INDEX idx_companies_industry ON companies (industry);
```

**Contacts Table:**
```sql
CREATE TABLE contacts (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(254),
    phone VARCHAR(20),
    company_id INTEGER REFERENCES companies(id),
    custom_fields JSONB DEFAULT '{}',
    deleted_at TIMESTAMPTZ,
    created_by INTEGER REFERENCES users(id) NOT NULL,
    updated_by INTEGER REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Performance and search indexes
CREATE INDEX idx_contacts_email ON contacts (email);
CREATE INDEX idx_contacts_company_id ON contacts (company_id);
CREATE INDEX idx_contacts_search ON contacts 
USING gin(to_tsvector('english', first_name || ' ' || last_name || ' ' || COALESCE(email, '')));
```

### Sales Pipeline

**Deals Table:**
```sql
CREATE TABLE deals (
   id SERIAL PRIMARY KEY,
   title VARCHAR(200) NOT NULL,
   value DECIMAL(12,2),
   probability DECIMAL(5,2) CHECK (probability >= 0 AND probability <= 100),
   stage VARCHAR(50) NOT NULL,
   primary_contact_id INTEGER REFERENCES contacts(id) ON DELETE SET NULL,
   company_id INTEGER REFERENCES companies(id) ON DELETE SET NULL,
   owner_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
   expected_close_date DATE,
   actual_close_date DATE,
   deal_source VARCHAR(100),
   description TEXT,
   notes TEXT,
   created_at TIMESTAMP DEFAULT NOW(),
   updated_at TIMESTAMP DEFAULT NOW(),
   created_by INTEGER REFERENCES users(id),
   updated_by INTEGER REFERENCES users(id)
);

-- Performance indexes
CREATE INDEX idx_deals_primary_contact ON deals(primary_contact_id);
CREATE INDEX idx_deals_company ON deals(company_id);
CREATE INDEX idx_deals_owner ON deals(owner_id);
CREATE INDEX idx_deals_stage ON deals(stage);
CREATE INDEX idx_deals_expected_close ON deals(expected_close_date);
CREATE INDEX idx_deals_created_at ON deals(created_at);
```

**Deal Contacts Association:**
```sql
CREATE TABLE deal_contacts (
    id SERIAL PRIMARY KEY,
    deal_id INTEGER REFERENCES deals(id) ON DELETE CASCADE,
    contact_id INTEGER REFERENCES contacts(id) ON DELETE CASCADE,
    role VARCHAR(50), -- e.g., 'decision_maker', 'influencer', 'user'
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(id)
);

-- Prevent duplicate associations
CREATE UNIQUE INDEX idx_deal_contacts_unique ON deal_contacts (deal_id, contact_id);
CREATE INDEX idx_deal_contacts_deal_id ON deal_contacts (deal_id);
CREATE INDEX idx_deal_contacts_contact_id ON deal_contacts (contact_id);
```

## Schema Creation Process

### New Tenant Workflow

1. **Validate Organization Data:**
   ```sql
   -- Check subdomain availability
   SELECT EXISTS(SELECT 1 FROM tenants WHERE subdomain = 'newcorp');
   ```

2. **Create Tenant Record:**
   ```sql
   INSERT INTO tenants (name, subdomain, schema_name)
   VALUES ('New Corp', 'newcorp', 'tenant_newcorp')
   RETURNING id, schema_name;
   ```

3. **Create Tenant Schema:**
   ```sql
   CREATE SCHEMA tenant_newcorp;
   SET search_path = tenant_newcorp;
   ```

4. **Copy Table Structure:**
   ```sql
   -- Copy all tables from template schema
   CREATE TABLE users (LIKE template.users INCLUDING ALL);
   CREATE TABLE companies (LIKE template.companies INCLUDING ALL);
   CREATE TABLE contacts (LIKE template.contacts INCLUDING ALL);
   CREATE TABLE deals (LIKE template.deals INCLUDING ALL);
   -- etc.
   ```

5. **Create Initial Admin User:**
   ```sql
   INSERT INTO tenant_newcorp.users (email, password_hash, first_name, last_name, role)
   VALUES ('admin@newcorp.com', '$2a$10$...', 'Admin', 'User', 'admin');
   ```

### Template Schema Maintenance

The `template` schema contains the canonical table definitions:

```sql
-- Update template schema when adding new features
CREATE SCHEMA IF NOT EXISTS template;
SET search_path = template;

-- Define canonical table structure
CREATE TABLE users (...);
CREATE TABLE companies (...);
-- etc.
```

## Query Patterns

### Tenant Context Setting

Every service request sets the tenant context:

```sql
-- Extract tenant from subdomain: example.yourcrm.com → example
SELECT schema_name FROM tenants WHERE subdomain = 'example';

-- Set search path for all subsequent queries
SET search_path = tenant_example;

-- All queries now operate within tenant scope
SELECT * FROM contacts WHERE email = 'john@example.com';
```

### Cross-Schema Operations

Global operations access multiple schemas:

```sql
-- Tenant management queries
SELECT t.name, t.subdomain, t.created_at,
       (SELECT COUNT(*) FROM tenant_users(t.schema_name)) as user_count
FROM tenants t
ORDER BY t.created_at DESC;

-- System-wide statistics
WITH tenant_stats AS (
    SELECT schema_name,
           get_table_count(schema_name, 'contacts') as contact_count,
           get_table_count(schema_name, 'deals') as deal_count
    FROM tenants
)
SELECT SUM(contact_count) as total_contacts,
       SUM(deal_count) as total_deals,
       COUNT(*) as tenant_count
FROM tenant_stats;
```

## Data Types and Conventions

### Standard Column Types

**Identifiers:**
- `id SERIAL PRIMARY KEY` - Auto-incrementing primary keys
- `{entity}_id INTEGER REFERENCES` - Foreign key references

**Text Fields:**
- `VARCHAR(254)` - Email addresses (RFC 5321 compliant)
- `VARCHAR(100)` - Names and short text
- `VARCHAR(255)` - Longer text fields
- `TEXT` - Unlimited text (notes, descriptions)

**Temporal Data:**
- `TIMESTAMPTZ` - Timezone-aware timestamps
- `DATE` - Date-only fields (close dates, etc.)
- All tables include `created_at`, `updated_at`

**Flexible Data:**
- `JSONB` - Custom fields, metadata, configurations
- `DECIMAL(12,2)` - Monetary values (deals, revenue)
- `DECIMAL(5,2)` - Percentages (probability, discounts)

### Audit Fields

Standard audit fields across all entities:

```sql
created_by INTEGER REFERENCES users(id) NOT NULL,
updated_by INTEGER REFERENCES users(id),
created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
deleted_at TIMESTAMPTZ  -- Soft delete pattern
```

### Custom Fields Pattern

JSONB columns for tenant-specific extensions:

```sql
-- Contacts custom fields
custom_fields JSONB DEFAULT '{}'

-- Example usage
UPDATE contacts SET custom_fields = '{
    "lead_source": "trade_show",
    "priority": "high", 
    "decision_maker": true,
    "custom_tags": ["enterprise", "hot_lead"]
}' WHERE id = 123;

-- Query by custom fields
SELECT * FROM contacts 
WHERE custom_fields->>'priority' = 'high'
  AND custom_fields->>'decision_maker' = 'true';
```

## Performance Optimization

### Indexing Strategy

**Primary Indexes:**
- All primary keys automatically indexed
- Foreign keys indexed for JOIN performance
- Unique constraints create automatic indexes

**Search Indexes:**
- Full-text search using PostgreSQL GIN indexes
- Composite indexes for common filter patterns
- Partial indexes for soft-deleted records

**Example Optimized Queries:**
```sql
-- Contact search with proper index usage
EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM contacts 
WHERE to_tsvector('english', first_name || ' ' || last_name) 
      @@ to_tsquery('english', 'john & smith');

-- Deal pipeline query with index support
EXPLAIN (ANALYZE, BUFFERS)
SELECT stage, COUNT(*), SUM(value)
FROM deals 
WHERE owner_id = 123 
  AND expected_close_date >= CURRENT_DATE
GROUP BY stage;
```

### Connection Pooling

```sql
-- Optimal connection pool settings per tenant load
-- Small tenants: 2-5 connections
-- Medium tenants: 5-10 connections  
-- Large tenants: 10-20 connections

-- Monitor connection usage
SELECT schema_name, active_connections, total_connections
FROM pg_stat_activity_by_schema;
```

## Data Migration Strategy

### Schema Versioning

Each tenant schema maintains version tracking:

```sql
CREATE TABLE schema_migrations (
    version VARCHAR(20) PRIMARY KEY,
    applied_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Track applied migrations per tenant
INSERT INTO schema_migrations (version) VALUES ('1.2.11');
```

### Migration Patterns

**Add Column (Safe):**
```sql
-- Add to template first
ALTER TABLE template.contacts ADD COLUMN linkedin_url VARCHAR(255);

-- Apply to all tenant schemas
DO $$
DECLARE
    tenant_record RECORD;
BEGIN
    FOR tenant_record IN SELECT schema_name FROM tenants LOOP
        EXECUTE format('ALTER TABLE %I.contacts ADD COLUMN linkedin_url VARCHAR(255)', 
                      tenant_record.schema_name);
    END LOOP;
END $$;
```

**Modify Column (Requires Downtime):**
```sql
-- Test migration on copy first
-- Apply during maintenance window
-- Validate data integrity post-migration
```

## Security Considerations

### Row Level Security

Future enhancement for additional security:

```sql
-- Enable RLS on sensitive tables
ALTER TABLE contacts ENABLE ROW LEVEL SECURITY;

-- Policy for tenant isolation
CREATE POLICY tenant_isolation ON contacts
FOR ALL TO application_role
USING (get_current_tenant_id() = extract_tenant_from_schema());
```

### Data Encryption

**At Rest:**
- PostgreSQL transparent data encryption
- Encrypted database volumes
- Backup encryption

**In Transit:**
- TLS 1.3 for all connections
- Certificate-based authentication
- VPN tunneling in production

### Access Control

**Database Users:**
- `application` - Read/write access to tenant schemas
- `readonly` - Analytics and reporting queries
- `admin` - Schema management and migrations
- `backup` - Backup operations only

## Monitoring and Maintenance

### Performance Monitoring

```sql
-- Query performance by tenant
SELECT schema_name, 
       avg(query_time) as avg_query_time,
       max(query_time) as max_query_time,
       count(*) as query_count
FROM query_log 
WHERE timestamp >= NOW() - INTERVAL '1 hour'
GROUP BY schema_name
ORDER BY avg_query_time DESC;

-- Index usage statistics
SELECT schemaname, tablename, indexname, 
       idx_scan, idx_tup_read, idx_tup_fetch
FROM pg_stat_user_indexes 
WHERE schemaname LIKE 'tenant_%'
ORDER BY idx_scan DESC;
```

### Maintenance Tasks

**Daily:**
- VACUUM ANALYZE on active tenant schemas
- Monitor connection pool usage
- Check for long-running queries

**Weekly:**
- REINDEX on heavily updated tables
- Analyze table statistics
- Clean up expired invitation tokens

**Monthly:**
- Full database statistics update
- Review and optimize slow queries
- Plan capacity for growing tenants

## Related Documentation

- [Global Schema Tables](./global/Tenants.md) - Tenant registry details
- [Tenant Template Schema](./tenant-template/) - Individual table documentation
- [SQLC Implementation](../architecture/sqlc.md) - Database access patterns
- [Query Documentation](./queries/) - SQL query reference by service