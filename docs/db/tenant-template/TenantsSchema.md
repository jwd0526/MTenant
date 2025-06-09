# Users Table Documentation

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
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(id),
    deleted_at TIMESTAMPTZ
);
```

## Column Design Decisions
- **email**: 254-char RFC-compliant max, unique within tenant scope only
- **password_hash**: 60-char for bcrypt output, fixed length for security
- **role**: CHECK constraint for predefined roles, simpler than ENUMs
- **permissions**: JSONB for flexible, granular permission structures
- **active/email_verified**: Boolean flags for account status management
- **created_by**: Self-referencing FK for audit trail of user creation

## Schema Context
- Lives within tenant schemas (`tenant_[subdomain]`)
- Complete data isolation between tenants
- Template copied to each new tenant schema during organization setup

## Indexes
```sql
-- Automatic indexes
CREATE INDEX idx_users_email ON users (email);           -- Login lookups
CREATE INDEX idx_users_active ON users (active);         -- Active user queries
CREATE INDEX idx_users_role ON users (role);             -- Role-based queries
CREATE INDEX idx_users_created_at ON users (created_at); -- Audit queries
```

## Key Operations
```sql
-- User authentication
SELECT id, password_hash, role, active FROM users 
WHERE email = 'john@techflow.com' AND active = true;

-- Role-based user listing
SELECT id, first_name, last_name, email, role FROM users 
WHERE active = true AND deleted_at IS NULL;

-- Create new user
INSERT INTO users (email, password_hash, first_name, last_name, role, created_by)
VALUES ('sarah@techflow.com', '$2b$12$...', 'Sarah', 'Johnson', 'sales_rep', 1);

-- Soft delete user
UPDATE users SET deleted_at = CURRENT_TIMESTAMP WHERE id = 5;
```

## Role Hierarchy
- **admin**: Full tenant management, user creation, settings
- **manager**: Team oversight, reporting, deal management
- **sales_rep**: Contact/deal CRUD, activity logging
- **viewer**: Read-only access to assigned records

## Application Flow
1. User logs in via tenant subdomain (`techflow.yourcrm.com`)
2. Extract tenant schema from subdomain lookup
3. Set search path: `SET search_path = tenant_techflow`
4. Authenticate against tenant-specific users table
5. Load role and permissions for authorization

## Constraints Tested
- Email uniqueness within tenant scope
- Role validation via CHECK constraint
- NOT NULL enforcement on required fields
- Self-referencing FK for created_by audit trail
- Soft delete preservation of referential integrity