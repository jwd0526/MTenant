# Tenant Service

Multi-tenant organization management and schema provisioning service.

## Overview

The Tenant Service manages organization registration, dynamic schema creation for new tenants, user invitation systems, and tenant configuration. It's the core service that enables multi-tenancy across the platform.

## Current Implementation Status

**Status**: Basic SQLC setup completed, placeholder main.go implementation
- ✅ SQLC configuration (`sqlc.yaml`) 
- ✅ Database schema (`db/schema/`)
- ✅ SQL queries (`db/queries/`)
- ✅ Generated code (`internal/db/`)
- ❌ HTTP handlers and business logic (planned)
- ❌ Schema provisioning logic (planned)
- ❌ Invitation system (planned)

## Database Schema

### Global Tables

**`tenants`** - Organization registry (global table)
```sql
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    subdomain VARCHAR(100) UNIQUE NOT NULL,
    status VARCHAR(50) DEFAULT 'active',
    schema_name VARCHAR(100) UNIQUE NOT NULL,
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**`invitations`** - Cross-tenant invitation system
```sql
CREATE TABLE invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'user',
    invited_by UUID REFERENCES users(id),
    token VARCHAR(255) UNIQUE NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    accepted_at TIMESTAMP
);
```

## SQLC Queries

The service includes comprehensive SQLC queries for tenant management:

- **Tenant Management**: `CreateTenant`, `GetTenantByID`, `GetTenantBySubdomain`, `UpdateTenant`, `ListTenants`
- **Schema Management**: `GetTenantBySchemaName`, `UpdateTenantSettings`
- **Invitations**: `CreateInvitation`, `GetInvitationByToken`, `AcceptInvitation`, `ListTenantInvitations`

## Planned API Endpoints

**Current Status**: Endpoints not implemented - service has placeholder main.go

### Tenant Management
```
POST   /api/tenants/register       # Create new organization
GET    /api/tenants/current        # Get current tenant info
PUT    /api/tenants/settings       # Update tenant configuration
GET    /api/tenants/:id            # Get tenant details
DELETE /api/tenants/:id            # Deactivate tenant
```

### User Management
```
GET    /api/tenants/users          # List tenant users
POST   /api/tenants/invite         # Send user invitation
GET    /api/tenants/invitations    # List pending invitations
DELETE /api/tenants/invitations/:id # Cancel invitation
```

### Schema Management
```
POST   /api/tenants/provision      # Provision tenant schema
GET    /api/tenants/schema/status  # Check schema status
PUT    /api/tenants/schema/migrate # Run tenant migrations
```

## Planned Features

### Organization Registration
- Subdomain validation and uniqueness checking
- Automatic schema provisioning
- Initial admin user creation
- Default configuration setup

### Dynamic Schema Creation
1. Validate organization details and subdomain uniqueness
2. Create tenant record in global registry
3. Execute `CREATE SCHEMA tenant_{id}`
4. Copy table structure from tenant template
5. Create initial admin user in tenant schema
6. Return tenant context for authentication

### User Invitation System
- Email-based invitations with secure tokens
- Role-based invitation (admin, user, read-only)
- Invitation expiry and cleanup
- Multi-tenant user management

### Tenant Configuration
- Custom tenant settings (JSONB)
- Feature flags per tenant
- Branding and customization options
- Usage limits and quotas

## Service Configuration

### Environment Variables
```bash
# Database
DATABASE_URL=postgresql://admin:admin@localhost:5433/crm-platform

# Email Service (for invitations)
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=noreply@example.com
SMTP_PASSWORD=smtp-password

# Application
PORT=8081
LOG_LEVEL=info

# Tenant Configuration
DEFAULT_SCHEMA_TEMPLATE=tenant_template
MAX_USERS_PER_TENANT=100
```

## Schema Provisioning Process

### Template-Based Provisioning
1. **Validation Phase**
   - Check subdomain availability
   - Validate organization data
   - Verify admin user details

2. **Creation Phase**
   - Create tenant record in global registry
   - Generate unique schema name
   - Execute schema creation SQL

3. **Population Phase**
   - Copy table structure from template
   - Insert default data
   - Create admin user record

4. **Verification Phase**
   - Test schema accessibility
   - Validate table creation
   - Return tenant context

### Schema Template Structure
```sql
-- Executed for each new tenant
CREATE SCHEMA tenant_{tenant_id};
SET search_path = tenant_{tenant_id};

-- Copy all tables from tenant_template
-- contacts, companies, deals, activities, etc.
```

## Inter-Service Communication

### Outbound Calls (Planned)
- **Auth Service**: Create initial admin user
- **Email Service**: Send invitation emails
- **Migration Service**: Apply tenant-specific migrations

### Inbound Calls (Planned)
- **All Services**: Tenant validation and context
- **Auth Service**: Tenant information during user registration
- **Frontend**: Organization management and user invitations

## Multi-Tenant Considerations

### Schema Isolation
- Each tenant gets a dedicated PostgreSQL schema
- Complete data isolation between tenants
- Independent table structures and data

### Tenant Context Management
- Automatic schema switching based on tenant ID
- Middleware for tenant context injection
- Secure tenant validation for all operations

### Performance Optimization
- Connection pooling per tenant
- Schema caching strategies
- Efficient tenant lookup mechanisms

## Security Considerations

### Tenant Isolation
- Complete schema-level data isolation
- Tenant context validation on all operations
- Secure subdomain validation

### Invitation Security
- Cryptographically secure invitation tokens
- Time-limited invitation expiry
- Email verification requirements

### Access Control
- Tenant admin role management
- Cross-tenant access prevention
- Audit logging for tenant operations

## Testing Strategy

### Current Tests
- Basic SQLC generated code tests
- Database connection tests
- Utility function tests

### Planned Tests
- Schema provisioning integration tests
- Tenant isolation validation tests
- Invitation flow end-to-end tests
- Performance tests for multi-tenant operations

## Development Status

**Current Directory Structure:**
```
services/tenant-service/
├── cmd/server/
│   ├── main.go                 # Placeholder implementation
│   └── main_test.go           # Basic tests
├── internal/
│   ├── db/                    # Generated SQLC code
│   ├── benchmark_test.go      # Performance tests
│   └── utils_test.go          # Utility tests
├── db/
│   ├── queries/               # SQL query files
│   └── schema/               # Database schema
├── Dockerfile                 # Container definition
├── go.mod                    # Go dependencies
└── sqlc.yaml                 # SQLC configuration
```

## Next Implementation Steps

1. **Schema Provisioning**: Implement dynamic tenant schema creation
2. **Invitation System**: Build email-based invitation flow
3. **HTTP Handlers**: Create REST API endpoints
4. **Tenant Middleware**: Develop tenant context middleware
5. **Configuration Management**: Add tenant settings system
6. **Migration Support**: Add tenant-specific migration handling
7. **Integration Testing**: Test with auth and other services

## Related Documentation

- [Service Architecture](../architecture/services.md) - Overall microservices design
- [Database Design](../architecture/database.md) - Multi-tenant data architecture
- [SQLC Implementation](../architecture/sqlc.md) - Database query patterns
- [Tenant Service Queries](../database/queries/tenant-service.md) - SQL query documentation
- [Database Migrations](../database/migrations.md) - Migration strategy for multi-tenancy