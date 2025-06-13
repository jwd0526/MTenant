-- Remove indexes first
DROP INDEX IF EXISTS idx_tenants_subdomain;
DROP INDEX IF EXISTS idx_tenants_schema_name;
DROP INDEX IF EXISTS idx_tenants_status;

-- Remove tenants table and constraints
DROP TABLE IF EXISTS tenants;