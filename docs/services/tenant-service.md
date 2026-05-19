# Tenant Service ✅ (IMPLEMENTED)

**Last Updated:** 2026-05-19\
*Full implementation with handlers, services, and schema provisioning*

Multi-tenant organization management and schema provisioning service.

## Overview

The Tenant Service manages organization registration, dynamic schema creation for new tenants, and tenant configuration. It's the foundational service that enables multi-tenancy across the platform by provisioning isolated PostgreSQL schemas for each organization.

## Implementation Status

**Status**: ✅ Complete and ready for deployment

- ✅ SQLC configuration and generated code
- ✅ Database schema and queries
- ✅ HTTP handlers (health.go, tenants.go)
- ✅ Business logic layer (tenant_service.go)
- ✅ Request/response models
- ✅ Error handling
- ✅ Server setup with routes
- ✅ Schema provisioning logic
- ✅ Health check endpoints
- ⏳ Invitation system (deferred)
- ⏳ Integration tests (planned)

## Database Schema

### Global Tables

**`tenants`** - Organization registry (global table)
```sql
CREATE TABLE tenants (
    id TEXT PRIMARY KEY, -- ULID format (26 chars)
    name VARCHAR(255) NOT NULL,
    subdomain VARCHAR(100) UNIQUE NOT NULL,
    schema_name VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**`invitations`** - Cross-tenant invitation system (defined but not yet used)
```sql
CREATE TABLE invitations (
    id TEXT PRIMARY KEY, -- ULID format
    tenant_id TEXT REFERENCES tenants(id),
    email VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    accepted_at TIMESTAMP,
    invited_by INTEGER,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## API Endpoints

**Port**: 8081 (default)

### Health Check
```
GET /health                         # Database health status
```

### Tenant Management
```
POST   /internal/tenants            # Create new tenant with schema
GET    /internal/tenants            # List all tenants
GET    /internal/tenants/:id        # Get tenant by ID
GET    /internal/tenants/subdomain/:subdomain  # Get tenant by subdomain
PUT    /internal/tenants/:id        # Update tenant details
GET    /internal/tenants/:id/health # Check tenant health (schema exists)
```

### Test/Development Endpoints
```
POST   /internal/test-tenants       # Bulk create test tenants
DELETE /internal/test-tenants?confirm=true  # Delete all test tenants
```

## Implemented Features

### Organization Registration

**CreateTenant** - Full tenant provisioning flow:
1. Validates subdomain uniqueness
2. Generates ULID for tenant ID
3. Creates tenant record in global registry
4. Generates schema name: `tenant_{ULID}`
5. Creates PostgreSQL schema
6. Copies template schema structure
7. Returns complete tenant details

**Example Request:**
```json
POST /internal/tenants
{
  "name": "Acme Corporation",
  "subdomain": "acme"
}
```

**Example Response:**
```json
{
  "id": "01HK153X003BMPJNJB6JHKXK8T",
  "name": "Acme Corporation",
  "subdomain": "acme",
  "schema_name": "tenant_01HK153X003BMPJNJB6JHKXK8T",
  "created_at": "2026-05-18T21:00:00Z",
  "updated_at": "2026-05-18T21:00:00Z"
}
```

### Dynamic Schema Creation

The service automatically provisions tenant schemas using the template approach:

1. **Schema Creation**: `CREATE SCHEMA IF NOT EXISTS "tenant_{id}"`
2. **Template Copy**: Copies all table structures from `template` schema
3. **Seed Data**: Copies initial data (roles, settings, pipeline stages)
4. **Verification**: Validates schema exists and is accessible

### Tenant Discovery

Other services use these endpoints to find tenant information:

- **By Subdomain**: Maps `acme.yourcrm.com` → tenant ID and schema
- **By ID**: Retrieves full tenant details for a given ULID
- **Health Check**: Verifies tenant's schema exists in PostgreSQL

### Tenant Health Monitoring

**GetTenantHealth** endpoint provides:
- Tenant record existence
- Schema existence verification
- Health status and diagnostic messages

**Example Response:**
```json
{
  "tenant_id": "01HK153X003BMPJNJB6JHKXK8T",
  "healthy": true,
  "schema_name": "tenant_01HK153X003BMPJNJB6JHKXK8T",
  "message": "tenant is healthy"
}
```

## Service Architecture

### Directory Structure
```
services/tenant-service/
├── cmd/server/
│   └── main.go                 # Server setup and routing
├── internal/
│   ├── handlers/
│   │   ├── health.go           # Health check handler
│   │   └── tenants.go          # Tenant CRUD handlers
│   ├── services/
│   │   └── tenant_service.go   # Business logic layer
│   ├── models/
│   │   ├── requests.go         # API request DTOs
│   │   └── responses.go        # API response DTOs
│   ├── errors/
│   │   └── errors.go           # Error definitions
│   ├── db/                     # SQLC generated code
│   └── config/
│       └── config.go           # Configuration helpers
├── db/
│   ├── queries/                # SQL query files
│   └── schema/                 # Database schema
└── tests/                      # Test files
```

### Implementation Layers

**Handlers** (`internal/handlers/`):
- Parse HTTP requests and bind JSON
- Call service layer for business logic
- Return appropriate HTTP status codes
- Error handling and response formatting

**Services** (`internal/services/`):
- Tenant CRUD operations
- Schema provisioning orchestration
- Business rule validation
- External package integration (pkg/tenant)

**Models** (`internal/models/`):
- Request validation structs
- Response formatting structs
- Error response structures

## Schema Provisioning Process

### Template-Based Provisioning

1. **Validation Phase**
   - Check subdomain uniqueness (via SQLC query)
   - Validate organization data
   - Generate ULID for tenant ID

2. **Creation Phase**
   - Create tenant record in global `tenants` table
   - Generate schema name: `tenant_{ULID}`
   - Execute `CREATE SCHEMA` command

3. **Population Phase**
   - Copy all table structures from `template` schema
   - Use `pkg/tenant.CopyTemplateSchema()` function
   - Copy seed data for default roles and settings

4. **Verification Phase**
   - Retrieve created tenant record
   - Return complete tenant response

### Template Schema

The service uses `pkg/tenant/schema.go` for:
- `CreateSchema()` - Create new PostgreSQL schema
- `CopyTemplateSchema()` - Copy table structures and seed data
- `SchemaExists()` - Verify schema health
- `GenerateSchemaName()` - Generate schema names

## Inter-Service Communication

### Outbound Calls
None - This is a foundational service with no dependencies

### Inbound Calls
- **Auth Service**: Tenant validation during login
- **All Services**: Subdomain-to-schema mapping
- **API Gateway**: Tenant routing decisions
- **Admin Tools**: Organization provisioning

## Service Configuration

### Environment Variables
```bash
# Database
DATABASE_URL=postgresql://admin:admin@localhost:5433/crm-platform

# Application
PORT=8081
LOG_LEVEL=info
```

## Multi-Tenant Considerations

### Schema Isolation
- Each tenant gets a dedicated PostgreSQL schema
- Complete data isolation between organizations
- Schema names follow pattern: `tenant_{ULID}`

### Tenant Context Management
- Services call tenant-service to map subdomain → schema
- `pkg/tenant` package provides schema utilities
- Middleware sets `search_path` based on tenant context

### Performance Optimization
- Connection pooling via `pkg/database`
- Subdomain lookup caching (planned)
- Schema template optimization

## Security Considerations

### Tenant Isolation
- Schema-level data isolation enforced by PostgreSQL
- Subdomain uniqueness validation
- ULID-based tenant IDs prevent enumeration

### Input Validation
- Subdomain format validation (alphanumeric, 3-63 chars)
- Name length validation
- SQL injection prevention via SQLC parameterized queries

### Access Control
- Internal-only endpoints (not exposed externally)
- Services authenticate via service-to-service tokens
- Admin operations require elevated permissions

## Testing Strategy

### Current Tests
- Build verification complete
- No compiler errors or warnings

### Planned Tests
- Schema provisioning integration tests
- Tenant isolation validation
- Subdomain uniqueness enforcement
- Health check endpoint tests
- Bulk operations tests

## Error Handling

The service uses structured error responses:

- `409 Conflict` - Subdomain already exists
- `404 Not Found` - Tenant doesn't exist
- `400 Bad Request` - Invalid input format
- `500 Internal Server Error` - Schema creation failed
- `503 Service Unavailable` - Unhealthy tenant (schema missing)
- `501 Not Implemented` - Feature not yet implemented

## Development Status

### Completed
- ✅ Full CRUD operations for tenants
- ✅ Schema provisioning with template copy
- ✅ Health monitoring endpoints
- ✅ Bulk tenant creation for testing
- ✅ Error handling and validation
- ✅ Server setup and routing
- ✅ Database integration via SQLC

### Deferred
- ⏳ Invitation system (table exists, handlers not implemented)
- ⏳ Tenant settings management
- ⏳ Migration support for tenant schemas
- ⏳ Comprehensive integration tests

## Next Steps

1. **Integration Testing**: Test with PostgreSQL database
2. **Service Client**: Create `pkg/clients/tenant/` for other services
3. **Invitation System**: Implement email-based invitations
4. **Caching Layer**: Add Redis caching for subdomain lookups
5. **Monitoring**: Add metrics and observability
6. **Documentation**: Create API documentation

## Related Documentation

- [Service Architecture](../architecture/services.md) - Overall microservices design
- [Database Design](../architecture/database.md) - Multi-tenant data architecture
- [SQLC Implementation](../architecture/sqlc.md) - Database query patterns
- [Tenant Schema Utilities](../architecture/shared-packages.md) - pkg/tenant package
- [Global Tables](../database/global/Tenants.md) - Tenants table schema
