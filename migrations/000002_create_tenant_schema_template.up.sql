-- Create template schema for tenant-specific tables
CREATE SCHEMA tenant_template;

-- Set search path to work in the template schema
SET search_path TO tenant_template;

-- Users table (tenant-specific)
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(254) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    role VARCHAR(20) CHECK (role IN ('admin', 'manager', 'sales_rep', 'viewer')) NOT NULL DEFAULT 'sales_rep',
    status VARCHAR(20) CHECK (status IN ('active', 'inactive', 'pending')) NOT NULL DEFAULT 'pending',
    email_verified BOOLEAN DEFAULT FALSE,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(id)
);

-- Companies table
CREATE TABLE companies (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    domain VARCHAR(253),
    industry VARCHAR(100),
    size_category VARCHAR(50) CHECK (size_category IN ('startup', 'small', 'medium', 'large', 'enterprise')),
    parent_company_id INTEGER REFERENCES companies(id),
    street_address VARCHAR(255),
    city VARCHAR(100),
    state VARCHAR(100),
    country VARCHAR(100),
    postal_code VARCHAR(20),
    phone VARCHAR(50),
    website VARCHAR(255),
    custom_fields JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(id)
);

-- Contacts table
CREATE TABLE contacts (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(254),
    phone VARCHAR(50),
    job_title VARCHAR(100),
    company_id INTEGER REFERENCES companies(id),
    owner_id INTEGER REFERENCES users(id),
    status VARCHAR(50) CHECK (status IN ('lead', 'prospect', 'customer', 'inactive')) DEFAULT 'lead',
    source VARCHAR(100),
    street_address VARCHAR(255),
    city VARCHAR(100),
    state VARCHAR(100),
    country VARCHAR(100),
    postal_code VARCHAR(20),
    custom_fields JSONB DEFAULT '{}',
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(id)
);

-- Deals table
CREATE TABLE deals (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    value DECIMAL(15,2),
    currency VARCHAR(3) DEFAULT 'USD',
    stage VARCHAR(100) NOT NULL,
    probability INTEGER CHECK (probability >= 0 AND probability <= 100) DEFAULT 0,
    expected_close_date DATE,
    actual_close_date DATE,
    owner_id INTEGER REFERENCES users(id),
    company_id INTEGER REFERENCES companies(id),
    primary_contact_id INTEGER REFERENCES contacts(id),
    source VARCHAR(100),
    close_reason VARCHAR(255),
    custom_fields JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(id)
);

-- Activities table (for logging all interactions)
CREATE TABLE activities (
    id SERIAL PRIMARY KEY,
    type VARCHAR(50) CHECK (type IN ('email', 'call', 'meeting', 'note', 'task', 'proposal')) NOT NULL,
    subject VARCHAR(255) NOT NULL,
    description TEXT,
    due_date TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    duration_minutes INTEGER,
    contact_id INTEGER REFERENCES contacts(id),
    company_id INTEGER REFERENCES companies(id),
    deal_id INTEGER REFERENCES deals(id),
    owner_id INTEGER REFERENCES users(id),
    custom_fields JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(id)
);

-- Create indexes for performance
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_status ON users(status);

CREATE INDEX idx_companies_name ON companies(name);
CREATE INDEX idx_companies_domain ON companies(domain);
CREATE INDEX idx_companies_parent_company_id ON companies(parent_company_id);

CREATE INDEX idx_contacts_email ON contacts(email);
CREATE INDEX idx_contacts_company_id ON contacts(company_id);
CREATE INDEX idx_contacts_owner_id ON contacts(owner_id);
CREATE INDEX idx_contacts_status ON contacts(status);
CREATE INDEX idx_contacts_name ON contacts(first_name, last_name);

CREATE INDEX idx_deals_stage ON deals(stage);
CREATE INDEX idx_deals_owner_id ON deals(owner_id);
CREATE INDEX idx_deals_company_id ON deals(company_id);
CREATE INDEX idx_deals_primary_contact_id ON deals(primary_contact_id);
CREATE INDEX idx_deals_expected_close_date ON deals(expected_close_date);
CREATE INDEX idx_deals_value ON deals(value);

CREATE INDEX idx_activities_type ON activities(type);
CREATE INDEX idx_activities_contact_id ON activities(contact_id);
CREATE INDEX idx_activities_company_id ON activities(company_id);
CREATE INDEX idx_activities_deal_id ON activities(deal_id);
CREATE INDEX idx_activities_owner_id ON activities(owner_id);
CREATE INDEX idx_activities_due_date ON activities(due_date);
CREATE INDEX idx_activities_created_at ON activities(created_at);

-- Reset search path
RESET search_path;