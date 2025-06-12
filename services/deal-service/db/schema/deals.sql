CREATE TABLE deals (
   id SERIAL PRIMARY KEY,
   title VARCHAR(200) NOT NULL,
   value DECIMAL(12,2),
   probability DECIMAL(5,2) CHECK (probability >= 0 AND probability <= 100),
   stage VARCHAR(50) NOT NULL,
   primary_contact_id INTEGER REFERENCES contacts(id) ON DELETE SET NULL,
   company_id INTEGER REFERENCES companies(id) ON DELETE SET NULL,
   owner_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
   expected_close_date DATE,
   actual_close_date DATE,
   deal_source VARCHAR(100),
   description TEXT,
   notes TEXT,
   created_at TIMESTAMP DEFAULT NOW(),
   updated_at TIMESTAMP DEFAULT NOW(),
   created_by INTEGER REFERENCES users(id),
   updated_by INTEGER REFERENCES users(id)
);

-- Indexes for performance
CREATE INDEX idx_deals_primary_contact ON deals(primary_contact_id);
CREATE INDEX idx_deals_company ON deals(company_id);
CREATE INDEX idx_deals_owner ON deals(owner_id);
CREATE INDEX idx_deals_stage ON deals(stage);
CREATE INDEX idx_deals_expected_close ON deals(expected_close_date);
CREATE INDEX idx_deals_created_at ON deals(created_at);