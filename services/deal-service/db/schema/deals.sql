CREATE TABLE deals (
   id SERIAL PRIMARY KEY,
   title VARCHAR(255) NOT NULL,
   description TEXT,
   value NUMERIC(15,2),
   currency VARCHAR(3) DEFAULT 'USD',
   stage VARCHAR(100) NOT NULL,
   probability INTEGER DEFAULT 0 CHECK (probability >= 0 AND probability <= 100),
   expected_close_date DATE,
   actual_close_date DATE,
   owner_id INTEGER,
   company_id INTEGER,
   primary_contact_id INTEGER,
   source VARCHAR(100),
   close_reason VARCHAR(255),
   custom_fields JSONB DEFAULT '{}',
   created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
   updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
   created_by INTEGER
);

-- Indexes for performance
CREATE INDEX idx_deals_primary_contact ON deals(primary_contact_id);
CREATE INDEX idx_deals_company ON deals(company_id);
CREATE INDEX idx_deals_owner ON deals(owner_id);
CREATE INDEX idx_deals_stage ON deals(stage);
CREATE INDEX idx_deals_expected_close ON deals(expected_close_date);
CREATE INDEX idx_deals_created_at ON deals(created_at);