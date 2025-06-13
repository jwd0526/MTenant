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