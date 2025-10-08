-- Create tenants table for multi-tenant registry
CREATE TABLE tenants (
    id TEXT PRIMARY KEY, -- ULID format (26 chars)
    name VARCHAR(255) NOT NULL,
    subdomain VARCHAR(63) UNIQUE NOT NULL,
    schema_name VARCHAR(63) UNIQUE NOT NULL,
    status VARCHAR(50) DEFAULT 'active' NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints for data validation
    CONSTRAINT tenants_subdomain_format CHECK (subdomain ~ '^[a-z0-9-]+$'),
    CONSTRAINT tenants_schema_format CHECK (schema_name ~ '^tenant_[a-z0-9_]+$'),
    CONSTRAINT tenants_status_valid CHECK (status IN ('active', 'suspended', 'pending'))
);

-- Create indexes for performance
CREATE INDEX idx_tenants_subdomain ON tenants(subdomain);
CREATE INDEX idx_tenants_schema_name ON tenants(schema_name);
CREATE INDEX idx_tenants_status ON tenants(status);

-- Insert default system tenant for testing
INSERT INTO tenants (id, name, subdomain, schema_name) 
VALUES ('01HK153X003BMPJNJB6JHKXK8T', 'System Admin', 'admin', 'tenant_admin');