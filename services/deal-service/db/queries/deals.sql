-- name: CreateDeal :one
INSERT INTO deals (
    title, value, probability, stage, primary_contact_id, company_id, 
    owner_id, expected_close_date, deal_source, description, notes, created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
) RETURNING *;

-- name: GetDealByID :one
SELECT d.*, 
       c.first_name || ' ' || c.last_name as primary_contact_name,
       comp.name as company_name,
       u.first_name || ' ' || u.last_name as owner_name
FROM deals d
LEFT JOIN contacts c ON d.primary_contact_id = c.id AND c.deleted_at IS NULL
LEFT JOIN companies comp ON d.company_id = comp.id AND comp.deleted_at IS NULL
LEFT JOIN users u ON d.owner_id = u.id AND u.deleted_at IS NULL
WHERE d.id = $1;

-- name: ListDeals :many
SELECT d.*, 
       c.first_name || ' ' || c.last_name as primary_contact_name,
       comp.name as company_name,
       u.first_name || ' ' || u.last_name as owner_name
FROM deals d
LEFT JOIN contacts c ON d.primary_contact_id = c.id AND c.deleted_at IS NULL
LEFT JOIN companies comp ON d.company_id = comp.id AND comp.deleted_at IS NULL
LEFT JOIN users u ON d.owner_id = u.id AND u.deleted_at IS NULL
ORDER BY d.created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateDeal :one
UPDATE deals 
SET title = $2, value = $3, probability = $4, stage = $5,
    primary_contact_id = $6, company_id = $7, owner_id = $8,
    expected_close_date = $9, deal_source = $10, description = $11,
    notes = $12, updated_by = $13, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetDealsByStage :many
SELECT stage, COUNT(*) as deal_count, 
       COALESCE(SUM(value), 0) as total_value,
       COALESCE(SUM(value * probability / 100), 0) as weighted_value
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
        ELSE 99
    END;

-- name: GetDealsByOwner :many
SELECT * FROM deals 
WHERE owner_id = $1 AND actual_close_date IS NULL
ORDER BY expected_close_date ASC;

-- name: GetMonthlyForecast :many
SELECT 
    DATE_TRUNC('month', expected_close_date) as forecast_month,
    SUM(value * probability / 100) as expected_revenue,
    COUNT(*) as deal_count
FROM deals 
WHERE actual_close_date IS NULL AND expected_close_date IS NOT NULL
GROUP BY DATE_TRUNC('month', expected_close_date)
ORDER BY forecast_month;

-- name: CloseDeal :one
UPDATE deals 
SET stage = $2, actual_close_date = $3, updated_by = $4, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetSalesRepPerformance :many
SELECT u.first_name, u.last_name, COUNT(*) as deals_closed, SUM(d.value) as total_revenue
FROM deals d
JOIN users u ON d.owner_id = u.id
WHERE d.actual_close_date >= $1 AND d.stage = 'Closed Won'
GROUP BY u.id, u.first_name, u.last_name
ORDER BY total_revenue DESC;