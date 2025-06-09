# Contacts Table Documentation

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

-- Full-text search index
CREATE INDEX idx_contacts_search ON contacts 
USING gin(to_tsvector('english', first_name || ' ' || last_name || ' ' || COALESCE(email, '')));

-- Performance indexes
CREATE INDEX idx_contacts_company_id ON contacts (company_id);
CREATE INDEX idx_contacts_email ON contacts (email);
CREATE INDEX idx_contacts_created_by ON contacts (created_by);
```

## Column Design Decisions
- **phone**: VARCHAR(20) for E.164 international format (`+15551234567`)
- **email**: 254-char RFC-compliant, nullable for contacts without email
- **company_id**: Optional association, contacts can exist without companies
- **custom_fields**: JSONB for tenant-specific fields like industry, lead source, priority
- **deleted_at**: Soft delete preserves audit trails and foreign key integrity

## Schema Context
- Lives within tenant schemas for complete data isolation
- References tenant-specific users and companies tables
- Template copied to each new tenant schema during setup

## Key Operations
```sql
-- Contact lookup by email
SELECT id, first_name, last_name, company_id FROM contacts 
WHERE email = 'john@techflow.com' AND deleted_at IS NULL;

-- Full-text search across names and email
SELECT * FROM contacts 
WHERE to_tsvector('english', first_name || ' ' || last_name || ' ' || COALESCE(email, '')) 
@@ to_tsquery('english', 'john & smith');

-- Create new contact
INSERT INTO contacts (first_name, last_name, email, phone, company_id, created_by)
VALUES ('Sarah', 'Johnson', 'sarah@client.com', '+15551234567', 3, 1);

-- Soft delete contact
UPDATE contacts SET deleted_at = CURRENT_TIMESTAMP, updated_by = 1 WHERE id = 5;
```

## Custom Fields Usage
```sql
-- Store custom tenant-specific data
UPDATE contacts SET custom_fields = '{
    "lead_source": "website",
    "industry": "technology", 
    "priority": "high",
    "last_contacted": "2024-01-15"
}' WHERE id = 1;

-- Query by custom fields
SELECT * FROM contacts WHERE custom_fields->>'industry' = 'technology';
```

## Application Flow
1. Sales rep creates/imports contact data
2. System validates email format and phone number
3. Associates contact with company if specified
4. Stores custom fields for tenant-specific requirements
5. All queries filtered by tenant schema and soft delete status

## Constraints Tested
- Foreign key integrity for company_id and user references
- JSONB validation for custom_fields structure
- Soft delete preservation of related data
- Full-text search index performance on name/email searches
- NULL handling for optional fields (email, phone, company_id)