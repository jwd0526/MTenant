```markdown
# Deals Table Documentation

## Purpose
Sales opportunity management within tenant schemas. Stores deal information, pipeline progression, and sales forecasting data for tracking revenue opportunities through the sales process.

## Table Structure
```sql
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
   deal_source VARCHAR(100), -- 'website', 'referral', 'cold_call', 'trade_show'
   description TEXT,
   notes TEXT,
   created_at TIMESTAMP DEFAULT NOW(),
   updated_at TIMESTAMP DEFAULT NOW(),
   created_by INTEGER REFERENCES users(id),
   updated_by INTEGER REFERENCES users(id)
);

-- Additional contacts involved in the deal
CREATE TABLE deal_contacts (
   deal_id INTEGER REFERENCES deals(id) ON DELETE CASCADE,
   contact_id INTEGER REFERENCES contacts(id) ON DELETE CASCADE,
   role VARCHAR(100), -- 'decision_maker', 'influencer', 'champion'
   created_at TIMESTAMP DEFAULT NOW(),
   PRIMARY KEY(deal_id, contact_id)
);

-- Indexes for performance
CREATE INDEX idx_deals_primary_contact ON deals(primary_contact_id);
CREATE INDEX idx_deals_company ON deals(company_id);
CREATE INDEX idx_deals_owner ON deals(owner_id);
CREATE INDEX idx_deals_stage ON deals(stage);
CREATE INDEX idx_deals_expected_close ON deals(expected_close_date);
CREATE INDEX idx_deals_created_at ON deals(created_at);
CREATE INDEX idx_deal_contacts_deal ON deal_contacts(deal_id);
CREATE INDEX idx_deal_contacts_contact ON deal_contacts(contact_id);
```

## Column Design Decisions
- **title**: Deal name/description for easy identification in pipeline views
- **value**: DECIMAL(12,2) supports deals up to $999,999,999.99 with precise calculations
- **probability**: Percentage (0-100) for weighted forecasting and pipeline analysis
- **stage**: Current position in sales pipeline (Lead, Qualified, Proposal, etc.)
- **primary_contact_id**: Main contact person driving the deal
- **deal_source**: Lead origin tracking for ROI analysis and marketing attribution
- **expected_close_date**: Critical for forecasting and pipeline velocity metrics
- **actual_close_date**: Records when deal was won/lost for performance analysis

## Schema Context
- Lives within tenant schemas for complete data isolation
- Template copied from `tenant_template` to each new tenant schema
- References contacts and companies tables for relationship tracking
- Links to users table for ownership and audit trail

## Key Operations
```sql
-- Create new deal
INSERT INTO deals (title, value, probability, stage, primary_contact_id, company_id, owner_id, expected_close_date, deal_source, created_by)
VALUES ('Enterprise CRM Implementation', 150000.00, 25.00, 'Qualified', 1, 1, 1, '2024-12-31', 'referral', 1);

-- Pipeline view by stage
SELECT stage, COUNT(*) as deal_count, SUM(value) as total_value, SUM(value * probability / 100) as weighted_value
FROM deals 
WHERE actual_close_date IS NULL
GROUP BY stage 
ORDER BY 
    CASE stage
        WHEN 'Lead' THEN 1
        WHEN 'Qualified' THEN 2
        WHEN 'Proposal' THEN 3
        WHEN 'Negotiation' THEN 4
        WHEN 'Closed Won' THEN 5
        WHEN 'Closed Lost' THEN 6
    END;

-- Sales rep performance
SELECT u.first_name, u.last_name, COUNT(*) as deals_closed, SUM(d.value) as total_revenue
FROM deals d
JOIN users u ON d.owner_id = u.id
WHERE d.actual_close_date >= '2024-01-01' AND d.stage = 'Closed Won'
GROUP BY u.id, u.first_name, u.last_name
ORDER BY total_revenue DESC;
```

## Deal Contacts Usage
```sql
-- Add decision maker to deal
INSERT INTO deal_contacts (deal_id, contact_id, role)
VALUES (1, 2, 'decision_maker');

-- Find all stakeholders in a deal
SELECT c.first_name, c.last_name, c.email, dc.role
FROM contacts c
JOIN deal_contacts dc ON c.id = dc.contact_id
WHERE dc.deal_id = 1;

-- Deals where specific contact is involved
SELECT d.title, d.value, d.stage, dc.role
FROM deals d
JOIN deal_contacts dc ON d.id = dc.deal_id
WHERE dc.contact_id = 1;
```

## Forecasting Queries
```sql
-- Monthly revenue forecast
SELECT 
    DATE_TRUNC('month', expected_close_date) as forecast_month,
    SUM(value * probability / 100) as expected_revenue,
    COUNT(*) as deal_count
FROM deals 
WHERE actual_close_date IS NULL AND expected_close_date IS NOT NULL
GROUP BY DATE_TRUNC('month', expected_close_date)
ORDER BY forecast_month;

-- Pipeline velocity (average days in each stage)
SELECT stage, AVG(EXTRACT(days FROM (updated_at - created_at))) as avg_days_in_stage
FROM deals 
WHERE actual_close_date IS NOT NULL
GROUP BY stage;
```

## Application Flow
1. Sales rep creates deal and associates with primary contact and company
2. Deal progresses through pipeline stages with probability updates
3. Additional stakeholders added via deal_contacts junction table
4. Activities and communications automatically linked to deal context
5. Analytics track conversion rates, velocity, and forecasting accuracy
6. All queries filtered by tenant schema for data isolation
```