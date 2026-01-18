-- name: CreateTenant :one
INSERT INTO tenants (id, name, subdomain, schema_name)
VALUES ($1, $2, $3, $4)
RETURNING id, name, subdomain, schema_name, created_at;

-- name: GetTenantBySubdomain :one
SELECT id, name, subdomain, schema_name, created_at, updated_at
FROM tenants
WHERE subdomain = $1;

-- name: GetTenantByID :one
SELECT id, name, subdomain, schema_name, created_at, updated_at
FROM tenants
WHERE id = $1;

-- name: GetTenantBySchemaName :one
SELECT id, name, subdomain, schema_name, created_at, updated_at
FROM tenants
WHERE schema_name = $1;

-- name: GetSchemaNameBySubdomain :one
SELECT schema_name FROM tenants WHERE subdomain = $1;

-- name: UpdateTenantName :exec
UPDATE tenants
SET name = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: CheckSubdomainExists :one
SELECT EXISTS(
    SELECT 1 FROM tenants 
    WHERE subdomain = $1
);

-- name: CheckSchemaNameExists :one
SELECT EXISTS(
    SELECT 1 FROM tenants 
    WHERE schema_name = $1
);

-- name: ListAllTenants :many
SELECT id, name, subdomain, schema_name, created_at
FROM tenants
ORDER BY name;

-- name: CountTenants :one
SELECT COUNT(*) FROM tenants;

-- name: GetRecentTenants :many
SELECT id, name, subdomain, schema_name, created_at
FROM tenants
ORDER BY created_at DESC
LIMIT $1;