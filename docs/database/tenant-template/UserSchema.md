# Users Table Documentation

**Last Updated:** 2025-10-08\
*Database schema reference for tenant template*

## Purpose
Template user table for tenant-specific schemas. Manages user authentication, authorization, and profile data within each isolated tenant environment.

## Table Structure
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

## Column Design Decisions
- **email**: 254-char RFC-compliant, unique within tenant scope
- **password_hash**: VARCHAR(60) for bcrypt output
- **role**: CHECK constraint for predefined roles
- **permissions**: JSONB for granular permission structures
- **created_by**: Self-referencing FK for audit trail

## Schema Context
- Lives within tenant schemas for complete data isolation
- Template copied from `tenant_template` to each new tenant schema
- Referenced by all other tables for audit tracking

## Role Hierarchy
- **admin**: Full tenant management, user creation, settings
- **manager**: Team oversight, reporting, deal management
- **sales_rep**: Contact/deal CRUD, activity logging
- **viewer**: Read-only access to assigned records

## Key Operations
```sql
-- User authentication
SELECT id, password_hash, role, active FROM users 
WHERE email = 'admin@techflow.com' AND active = true AND deleted_at IS NULL;

-- Create new team member
INSERT INTO users (email, password_hash, first_name, last_name, role, created_by)
VALUES ('sarah.johnson@techflow.com', '$2b$12$...', 'Sarah', 'Johnson', 'sales_rep', 1);

-- Update user role
UPDATE users SET role = 'manager', updated_by = 1 WHERE id = 3;

-- Soft delete user
UPDATE users SET deleted_at = CURRENT_TIMESTAMP, updated_by = 1 WHERE id = 5;
```

## Permissions Usage
```sql
-- Store granular permissions
UPDATE users SET permissions = '{
    "deals": {"view_all": true, "edit_all": false},
    "contacts": {"export": true, "import": true}
}' WHERE id = 2;
```

## Application Flow
1. User accesses tenant subdomain (`techflow.yourcrm.com`)
2. System sets search path: `SET search_path = tenant_techflow`
3. Authenticates against tenant-specific users table
4. Loads role and permissions for authorization
