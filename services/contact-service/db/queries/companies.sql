-- name: CreateCompany :one
INSERT INTO companies (
    name, domain, industry, street, city, state, country, postal_code,
    parent_company_id, custom_fields, company_size, employee_count, 
    annual_revenue, revenue_currency, created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
) RETURNING *;

-- name: GetCompanyByID :one
SELECT * FROM companies 
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetCompanyByDomain :one
SELECT id, name, industry FROM companies 
WHERE domain = $1 AND deleted_at IS NULL;

-- name: ListCompanies :many
SELECT * FROM companies 
WHERE deleted_at IS NULL 
ORDER BY name 
LIMIT $1 OFFSET $2;

-- name: UpdateCompany :one
UPDATE companies 
SET name = $2, domain = $3, industry = $4, street = $5, city = $6, 
    state = $7, country = $8, postal_code = $9, parent_company_id = $10,
    custom_fields = $11, company_size = $12, employee_count = $13,
    annual_revenue = $14, revenue_currency = $15, updated_by = $16, 
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteCompany :exec
UPDATE companies 
SET deleted_at = CURRENT_TIMESTAMP, updated_by = $2 
WHERE id = $1;

-- name: SearchCompaniesByName :many
SELECT * FROM companies 
WHERE name ILIKE '%' || $1 || '%' AND deleted_at IS NULL
ORDER BY name
LIMIT $2 OFFSET $3;

-- name: GetSubsidiaries :many
SELECT * FROM companies 
WHERE parent_company_id = $1 AND deleted_at IS NULL
ORDER BY name;

-- name: GetCompanyHierarchy :many
WITH RECURSIVE company_tree AS (
    SELECT c.id, c.name, c.parent_company_id, 0 as level 
    FROM companies c WHERE c.id = $1 AND c.deleted_at IS NULL
    UNION ALL
    SELECT c.id, c.name, c.parent_company_id, ct.level + 1
    FROM companies c JOIN company_tree ct ON c.parent_company_id = ct.id
    WHERE c.deleted_at IS NULL
)
SELECT ct.id, ct.name, ct.parent_company_id, ct.level FROM company_tree ct ORDER BY ct.level, ct.name;

-- name: GetCompaniesByRevenue :many
SELECT name, annual_revenue, company_size FROM companies 
WHERE annual_revenue > $1 AND deleted_at IS NULL 
ORDER BY annual_revenue DESC
LIMIT $2 OFFSET $3;

-- name: GetCompaniesByIndustry :many
SELECT * FROM companies 
WHERE industry = $1 AND deleted_at IS NULL
ORDER BY name;

-- name: UpdateCompanyCustomFields :one
UPDATE companies 
SET custom_fields = $2, updated_by = $3, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SearchCompaniesByCustomField :many
SELECT * FROM companies 
WHERE custom_fields->>$1 = $2 AND deleted_at IS NULL
ORDER BY name;