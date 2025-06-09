# Companies Table Documentation

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
    deleted_at TIMESTAMPTZ,
    created_by INTEGER REFERENCES users(id) NOT NULL,
    updated_by INTEGER REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Performance indexes
CREATE INDEX idx_companies_domain ON companies (domain);
CREATE INDEX idx_companies_parent_id ON companies (parent_company_id);
CREATE INDEX idx_companies_industry ON companies (industry);
CREATE INDEX idx_companies_revenue ON companies (annual_revenue);
CREATE INDEX idx_companies_created_by ON companies (created_by);
```

## Column Design Decisions
- **domain**: Clean domain format (`techflow.com`) for email-to-company matching and lead enrichment
- **parent_company_id**: Self-referencing FK enables corporate hierarchy tracking (subsidiaries, divisions)
- **custom_fields**: JSONB for tenant-specific data like lead source, contract type, or industry-specific metrics
- **company_size**: Categorical sizing for sales segmentation (`startup`, `small`, `medium`, `large`, `enterprise`)
- **annual_revenue**: DECIMAL(15,2) supports precise financial calculations up to $999 trillion
- **revenue_currency**: ISO currency codes for global CRM usage

## Schema Context
- Lives within tenant schemas for complete data isolation
- Referenced by contacts table for contact-company associations
- Template copied to each new tenant schema during setup

## Hierarchy Support
```sql
-- Create parent company
INSERT INTO companies (name, domain, industry, created_by)
VALUES ('Global Corp', 'globalcorp.com', 'Technology', 1);

-- Create subsidiary
INSERT INTO companies (name, domain, parent_company_id, created_by)
VALUES ('Global Tech Division', 'tech.globalcorp.com', 1, 1);

-- Find all subsidiaries
SELECT * FROM companies WHERE parent_company_id = 1;

-- Recursive hierarchy query
WITH RECURSIVE company_tree AS (
    SELECT id, name, parent_company_id, 0 as level FROM companies WHERE id = 1
    UNION ALL
    SELECT c.id, c.name, c.parent_company_id, ct.level + 1
    FROM companies c JOIN company_tree ct ON c.parent_company_id = ct.id
)
SELECT * FROM company_tree ORDER BY level;
```

## Key Operations
```sql
-- Company lookup by domain for email association
SELECT id, name FROM companies WHERE domain = 'techflow.com' AND deleted_at IS NULL;

-- Account prioritization by revenue
SELECT name, annual_revenue, company_size FROM companies 
WHERE annual_revenue > 10000000 ORDER BY annual_revenue DESC;

-- Create new company
INSERT INTO companies (name, domain, industry, city, country, company_size, created_by)
VALUES ('TechFlow Solutions', 'techflow.com', 'Technology', 'San Francisco', 'United States', 'startup', 1);

-- Market segmentation analysis
SELECT company_size, COUNT(*), AVG(annual_revenue) FROM companies 
WHERE deleted_at IS NULL GROUP BY company_size;
```

## Custom Fields Usage
```sql
-- Store tenant-specific business data
UPDATE companies SET custom_fields = '{
    "lead_source": "trade_show",
    "contract_type": "enterprise",
    "decision_maker": "CTO",
    "renewal_date": "2024-12-31"
}' WHERE id = 1;

-- Query by custom criteria
SELECT * FROM companies WHERE custom_fields->>'contract_type' = 'enterprise';
```

## Application Flow
1. Sales rep creates company during prospect research or contact import
2. System validates domain format and checks for duplicates
3. Associates contacts automatically based on email domain matching
4. Tracks corporate hierarchy for enterprise account management
5. All queries filtered by tenant schema and soft delete status

## Constraints Tested
- Self-referencing foreign key integrity for company hierarchy
- User foreign key constraints for audit trail
- JSONB validation for custom fields structure
- Domain uniqueness within tenant scope
- Soft delete preservation of contact associations
- Revenue and currency field validation