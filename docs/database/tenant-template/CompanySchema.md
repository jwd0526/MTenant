# Companies Table Documentation

**Last Updated:** 2025-10-08\
*Database schema reference for tenant template*

## Purpose
Business entity management within tenant schemas. Stores company information, organizational hierarchy, and business metrics for contact association and account-based selling.

## Table Structure
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

## Column Design Decisions
- **domain**: Clean domain format (`techflow.com`) for email-to-company matching
- **parent_company_id**: Self-referencing FK enables corporate hierarchy tracking
- **custom_fields**: JSONB for tenant-specific data like contract type, lead source
- **company_size**: Categorical sizing (`startup`, `small`, `medium`, `large`, `enterprise`)
- **annual_revenue**: DECIMAL(15,2) supports precise financial calculations

## Schema Context
- Lives within tenant schemas for complete data isolation
- Template copied from `tenant_template` to each new tenant schema
- Referenced by contacts table for associations

## Key Operations
```sql
-- Company lookup by domain
SELECT id, name, industry FROM companies 
WHERE domain = 'techflow.com' AND deleted_at IS NULL;

-- Create parent company
INSERT INTO companies (name, domain, industry, city, country, company_size, annual_revenue, created_by)
VALUES ('TechFlow Solutions', 'techflow.com', 'Technology', 'San Francisco', 'United States', 'medium', 25000000.00, 1);

-- Create subsidiary
INSERT INTO companies (name, domain, parent_company_id, industry, created_by)
VALUES ('TechFlow Europe', 'europe.techflow.com', 1, 'Technology', 1);

-- Account prioritization by revenue
SELECT name, annual_revenue, company_size FROM companies 
WHERE annual_revenue > 10000000 AND deleted_at IS NULL 
ORDER BY annual_revenue DESC;
```

## Hierarchy Support
```sql
-- Find all subsidiaries
SELECT * FROM companies WHERE parent_company_id = 1 AND deleted_at IS NULL;

-- Recursive hierarchy query
WITH RECURSIVE company_tree AS (
    SELECT id, name, parent_company_id, 0 as level 
    FROM companies WHERE id = 1 AND deleted_at IS NULL
    UNION ALL
    SELECT c.id, c.name, c.parent_company_id, ct.level + 1
    FROM companies c JOIN company_tree ct ON c.parent_company_id = ct.id
    WHERE c.deleted_at IS NULL
)
SELECT * FROM company_tree ORDER BY level, name;
```

## Custom Fields Usage
```sql
-- Store tenant-specific business data
UPDATE companies SET custom_fields = '{
    "lead_source": "trade_show",
    "contract_type": "enterprise",
    "renewal_date": "2024-12-31"
}' WHERE id = 1;

-- Query by custom criteria
SELECT * FROM companies WHERE custom_fields->>'contract_type' = 'enterprise';
```

## Application Flow
1. Sales rep creates company during prospect research
2. System validates domain format and checks for duplicates
3. Associates contacts automatically based on email domain matching
4. Tracks corporate hierarchy for enterprise account management
5. All queries filtered by tenant schema and soft delete status