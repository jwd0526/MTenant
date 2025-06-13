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