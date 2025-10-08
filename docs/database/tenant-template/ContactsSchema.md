# Contacts Table Documentation

**Last Updated:** 2025-10-08\
*Database schema reference for tenant template*

## Purpose
Customer and prospect data management within tenant schemas. Stores contact information, company associations, and custom fields for CRM sales activities.

## Table Structure
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

-- Performance indexes
CREATE INDEX idx_contacts_email ON contacts (email);
CREATE INDEX idx_contacts_company_id ON contacts (company_id);
CREATE INDEX idx_contacts_search ON contacts 
USING gin(to_tsvector('english', first_name || ' ' || last_name || ' ' || COALESCE(email, '')));
```

## Column Design Decisions
- **email**: 254-char RFC-compliant, nullable for contacts without email
- **phone**: VARCHAR(20) supports E.164 international format (`+15551234567`)
- **company_id**: Optional foreign key, contacts can exist independently
- **custom_fields**: JSONB for tenant-specific data like lead source, priority
- **deleted_at**: Soft delete preserves audit trails and foreign key integrity

## Schema Context
- Lives within tenant schemas for complete data isolation
- Template copied from `tenant_template` to each new tenant schema
- References tenant-specific users and companies tables

## Key Operations
```sql
-- Contact lookup by email
SELECT id, first_name, last_name, company_id FROM contacts 
WHERE email = 'john.smith@techflow.com' AND deleted_at IS NULL;

-- Full-text search
SELECT * FROM contacts 
WHERE to_tsvector('english', first_name || ' ' || last_name || ' ' || COALESCE(email, '')) 
@@ to_tsquery('english', 'john & smith');

-- Create new contact
INSERT INTO contacts (first_name, last_name, email, phone, company_id, created_by)
VALUES ('Sarah', 'Johnson', 'sarah.johnson@techflow.com', '+15551234567', 1, 1);

-- Soft delete contact
UPDATE contacts SET deleted_at = CURRENT_TIMESTAMP, updated_by = 1 WHERE id = 5;
```

## Custom Fields Usage
```sql
-- Store tenant-specific data
UPDATE contacts SET custom_fields = '{
    "lead_source": "trade_show",
    "priority": "high",
    "decision_maker": true
}' WHERE id = 1;

-- Query by custom criteria
SELECT * FROM contacts WHERE custom_fields->>'priority' = 'high';
```

## Application Flow
1. Sales rep creates/imports contact data
2. System validates email format and associates with company by domain
3. Stores custom fields for tenant-specific requirements
4. All queries filtered by tenant schema and soft delete status
