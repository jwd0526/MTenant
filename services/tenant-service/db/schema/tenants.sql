-- Core tenant registry table
CREATE TABLE tenants (
    id TEXT PRIMARY KEY, -- ULID format (26 chars)
    name VARCHAR(255) NOT NULL,
    subdomain VARCHAR(63) UNIQUE NOT NULL,
    schema_name VARCHAR(63) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- User invitations (global table for cross-tenant invites)
CREATE TABLE invitations (
    id TEXT PRIMARY KEY, -- ULID format (26 chars)
    tenant_id TEXT REFERENCES tenants(id) ON DELETE CASCADE,
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