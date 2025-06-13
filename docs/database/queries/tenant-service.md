# Tenant Service SQL Queries

Complete documentation of SQL queries used by the Tenant Service for organization management and user invitations.

## Overview

The Tenant Service manages multi-tenant organization setup, schema provisioning, and cross-tenant user invitations. It operates on both global tables (public schema) and tenant-specific schemas.

## Tenant Management Queries

### Tenant Creation

**Query: `CreateTenant`**
```sql
-- name: CreateTenant :one
INSERT INTO tenants (name, subdomain, schema_name)
VALUES ($1, $2, $3)
RETURNING id, name, subdomain, schema_name, created_at;
```

**Purpose:** Create a new tenant organization in the global registry.

**Parameters:**
- `$1` - Organization name (VARCHAR 255)
- `$2` - Subdomain (VARCHAR 63, must be unique)
- `$3` - Schema name (VARCHAR 63, format: `tenant_{id}`)

**Generated Go Method:**
```go
type CreateTenantParams struct {
    Name       string `json:"name"`
    Subdomain  string `json:"subdomain"`
    SchemaName string `json:"schema_name"`
}

func (q *Queries) CreateTenant(ctx context.Context, arg CreateTenantParams) (CreateTenantRow, error)
```

**Usage Example:**
```go
// Validate subdomain format and availability first
if !isValidSubdomain(subdomain) {
    return errors.New("invalid subdomain format")
}

exists, err := queries.CheckSubdomainExists(ctx, subdomain)
if exists {
    return errors.New("subdomain already taken")
}

// Create tenant record
schemaName := fmt.Sprintf("tenant_%s", generateSchemaName(subdomain))
tenant, err := queries.CreateTenant(ctx, db.CreateTenantParams{
    Name:       "Example Corp",
    Subdomain:  "example",
    SchemaName: schemaName,
})
if err != nil {
    return err
}

// Create schema and initialize tables
err = createTenantSchema(ctx, schemaName)
```

### Tenant Lookup

**Query: `GetTenantBySubdomain`**
```sql
-- name: GetTenantBySubdomain :one
SELECT id, name, subdomain, schema_name, created_at, updated_at
FROM tenants
WHERE subdomain = $1;
```

**Purpose:** Primary tenant lookup for request routing (happens on every request).

**Performance Notes:**
- Uses automatic UNIQUE index on subdomain
- Critical for request routing performance
- Should execute in <1ms

**Query: `GetTenantByID`**
```sql
-- name: GetTenantByID :one
SELECT id, name, subdomain, schema_name, created_at, updated_at
FROM tenants
WHERE id = $1;
```

**Query: `GetTenantBySchemaName`**
```sql
-- name: GetTenantBySchemaName :one
SELECT id, name, subdomain, schema_name, created_at, updated_at
FROM tenants
WHERE schema_name = $1;
```

**Purpose:** Administrative queries and schema management.

**Usage Example:**
```go
// Primary request routing (middleware)
func TenantMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Extract subdomain from request
        subdomain := extractSubdomain(c.Request.Host)
        
        // Lookup tenant
        tenant, err := queries.GetTenantBySubdomain(ctx, subdomain)
        if err != nil {
            c.JSON(404, gin.H{"error": "Tenant not found"})
            c.Abort()
            return
        }
        
        // Set tenant context
        c.Set("tenant_id", tenant.ID)
        c.Set("schema_name", tenant.SchemaName)
        c.Next()
    }
}
```

### Schema Name Utilities

**Query: `GetSchemaNameBySubdomain`**
```sql
-- name: GetSchemaNameBySubdomain :one
SELECT schema_name FROM tenants WHERE subdomain = $1;
```

**Purpose:** Lightweight lookup for database connection setup.

**Usage Example:**
```go
// Optimized tenant context setup
schemaName, err := queries.GetSchemaNameBySubdomain(ctx, subdomain)
if err != nil {
    return err
}

// Set search path for all subsequent queries
_, err = conn.Exec(ctx, "SET search_path = " + schemaName)
```

### Tenant Updates

**Query: `UpdateTenantName`**
```sql
-- name: UpdateTenantName :exec
UPDATE tenants
SET name = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;
```

**Purpose:** Update organization name (subdomain cannot be changed after creation).

**Usage Example:**
```go
err := queries.UpdateTenantName(ctx, db.UpdateTenantNameParams{
    ID:   tenantID,
    Name: "Example Corp - Updated",
})
```

## Validation Queries

### Uniqueness Checks

**Query: `CheckSubdomainExists`**
```sql
-- name: CheckSubdomainExists :one
SELECT EXISTS(
    SELECT 1 FROM tenants 
    WHERE subdomain = $1
);
```

**Query: `CheckSchemaNameExists`**
```sql
-- name: CheckSchemaNameExists :one
SELECT EXISTS(
    SELECT 1 FROM tenants 
    WHERE schema_name = $1
);
```

**Purpose:** Validate uniqueness before tenant creation.

**Usage Example:**
```go
// Pre-creation validation
func validateTenantCreation(subdomain, schemaName string) error {
    // Check subdomain
    exists, err := queries.CheckSubdomainExists(ctx, subdomain)
    if err != nil {
        return err
    }
    if exists {
        return errors.New("subdomain already taken")
    }
    
    // Check schema name
    exists, err = queries.CheckSchemaNameExists(ctx, schemaName)
    if err != nil {
        return err
    }
    if exists {
        return errors.New("schema name conflict")
    }
    
    return nil
}
```

## Tenant Listing and Analytics

### Administrative Queries

**Query: `ListAllTenants`**
```sql
-- name: ListAllTenants :many
SELECT id, name, subdomain, schema_name, created_at
FROM tenants
ORDER BY name;
```

**Query: `CountTenants`**
```sql
-- name: CountTenants :one
SELECT COUNT(*) FROM tenants;
```

**Query: `GetRecentTenants`**
```sql
-- name: GetRecentTenants :many
SELECT id, name, subdomain, schema_name, created_at
FROM tenants
ORDER BY created_at DESC
LIMIT $1;
```

**Purpose:** Administrative dashboard and system monitoring.

**Usage Example:**
```go
// Admin dashboard data
totalTenants, err := queries.CountTenants(ctx)
recentTenants, err := queries.GetRecentTenants(ctx, 10)

dashboard := AdminDashboard{
    TotalTenants:  totalTenants,
    RecentTenants: recentTenants,
}
```

## Invitation Management Queries

### Invitation Creation

**Query: `CreateInvitation`**
```sql
-- name: CreateInvitation :one
INSERT INTO invitations (tenant_id, email, role, token, expires_at, invited_by, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, tenant_id, email, role, token, expires_at, created_at;
```

**Purpose:** Send user invitation to join a tenant organization.

**Parameters:**
- `$1` - Tenant ID (INTEGER, FK to tenants)
- `$2` - Email address (VARCHAR 254)
- `$3` - Role ('admin', 'manager', 'sales_rep', 'viewer')
- `$4` - Secure token (VARCHAR 255, unique)
- `$5` - Expiration timestamp (TIMESTAMPTZ)
- `$6` - Inviting user ID (INTEGER, from tenant schema)
- `$7` - Metadata (JSONB, optional invitation details)

**Generated Go Method:**
```go
type CreateInvitationParams struct {
    TenantID  *int32          `json:"tenant_id"`
    Email     string          `json:"email"`
    Role      string          `json:"role"`
    Token     string          `json:"token"`
    ExpiresAt time.Time       `json:"expires_at"`
    InvitedBy *int32          `json:"invited_by"`
    Metadata  json.RawMessage `json:"metadata"`
}
```

**Usage Example:**
```go
// Generate secure invitation token
token := generateSecureToken(32)
expiresAt := time.Now().Add(time.Hour * 72) // 72-hour expiry

// Optional metadata
metadata := map[string]interface{}{
    "invitation_message": "Welcome to our team!",
    "role_description": "Sales representative with full CRM access",
}
metadataJSON, _ := json.Marshal(metadata)

invitation, err := queries.CreateInvitation(ctx, db.CreateInvitationParams{
    TenantID:  &tenantID,
    Email:     "newuser@example.com",
    Role:      "sales_rep",
    Token:     token,
    ExpiresAt: expiresAt,
    InvitedBy: &invitingUserID,
    Metadata:  metadataJSON,
})

// Send invitation email
err = sendInvitationEmail(invitation.Email, invitation.Token)
```

### Invitation Validation

**Query: `GetInvitationByToken`**
```sql
-- name: GetInvitationByToken :one
SELECT inv.id, inv.tenant_id, inv.email, inv.role, inv.token, 
       inv.expires_at, inv.accepted_at, inv.metadata,
       t.name as tenant_name, t.subdomain
FROM invitations inv
JOIN tenants t ON inv.tenant_id = t.id
WHERE inv.token = $1 AND inv.expires_at > CURRENT_TIMESTAMP;
```

**Purpose:** Validate invitation token and get tenant context.

**Features:**
- Joins with tenants table for complete context
- Checks expiration at database level
- Returns null if token expired or invalid

**Usage Example:**
```go
invitation, err := queries.GetInvitationByToken(ctx, token)
if err != nil {
    return c.JSON(400, gin.H{"error": "Invalid or expired invitation"})
}

if invitation.AcceptedAt.Valid {
    return c.JSON(400, gin.H{"error": "Invitation already accepted"})
}

// Process invitation acceptance
response := InvitationDetails{
    TenantName: invitation.TenantName,
    Subdomain:  invitation.Subdomain,
    Role:       invitation.Role,
    Email:      invitation.Email,
}
```

### Invitation Acceptance

**Query: `AcceptInvitation`**
```sql
-- name: AcceptInvitation :exec
UPDATE invitations 
SET accepted_at = CURRENT_TIMESTAMP
WHERE token = $1 AND accepted_at IS NULL;
```

**Purpose:** Mark invitation as accepted after user registration.

**Query: `GetPendingInvitations`**
```sql
-- name: GetPendingInvitations :many
SELECT id, email, role, expires_at, created_at
FROM invitations
WHERE tenant_id = $1 
  AND accepted_at IS NULL 
  AND expires_at > CURRENT_TIMESTAMP
ORDER BY created_at DESC;
```

**Purpose:** List pending invitations for tenant administrators.

**Usage Example:**
```go
// User accepts invitation
func acceptInvitation(token string, userDetails UserRegistration) error {
    // Validate invitation
    invitation, err := queries.GetInvitationByToken(ctx, token)
    if err != nil {
        return err
    }
    
    // Set tenant context and create user
    _, err = conn.Exec(ctx, "SET search_path = " + invitation.SchemaName)
    if err != nil {
        return err
    }
    
    // Create user in tenant schema
    user, err := queries.CreateUser(ctx, db.CreateUserParams{
        Email:     invitation.Email,
        FirstName: userDetails.FirstName,
        LastName:  userDetails.LastName,
        Role:      invitation.Role,
        // ... other fields
    })
    if err != nil {
        return err
    }
    
    // Mark invitation as accepted
    err = queries.AcceptInvitation(ctx, token)
    return err
}

// Admin view of pending invitations
pendingInvitations, err := queries.GetPendingInvitations(ctx, tenantID)
```

## Invitation Cleanup

### Expired Invitations

**Query: `CleanupExpiredInvitations`**
```sql
-- name: CleanupExpiredInvitations :exec
DELETE FROM invitations 
WHERE expires_at < CURRENT_TIMESTAMP - INTERVAL '30 days';
```

**Purpose:** Cleanup old expired invitations (run as scheduled job).

**Usage Example:**
```go
// Scheduled cleanup job (run daily)
func cleanupExpiredInvitations() {
    err := queries.CleanupExpiredInvitations(ctx)
    if err != nil {
        log.Printf("Failed to cleanup expired invitations: %v", err)
    }
}
```

## Performance Considerations

### Index Usage

**Critical Indexes:**
- `UNIQUE (subdomain)` - Primary tenant lookup
- `UNIQUE (schema_name)` - Schema validation
- `idx_invitations_token` - Token validation
- `idx_invitations_tenant_id` - Tenant invitation lists
- `idx_invitations_email` - User invitation lookup
- `idx_invitations_expires_at` - Cleanup queries

**Query Performance:**
```sql
-- Verify tenant lookup performance (critical path)
EXPLAIN (ANALYZE, BUFFERS) 
SELECT schema_name FROM tenants WHERE subdomain = 'example';

-- Should show: Index Scan using tenants_subdomain_key
-- Execution time: < 1ms

-- Verify invitation token lookup
EXPLAIN (ANALYZE, BUFFERS)
SELECT * FROM invitations WHERE token = 'abc123...';

-- Should show: Index Scan using idx_invitations_token
```

### Subdomain Validation

**Validation Rules:**
```go
func isValidSubdomain(subdomain string) bool {
    // Length: 1-63 characters (DNS compliance)
    if len(subdomain) < 1 || len(subdomain) > 63 {
        return false
    }
    
    // Characters: letters, numbers, hyphens only
    matched, _ := regexp.MatchString(`^[a-z0-9-]+$`, subdomain)
    if !matched {
        return false
    }
    
    // Cannot start or end with hyphen
    if strings.HasPrefix(subdomain, "-") || strings.HasSuffix(subdomain, "-") {
        return false
    }
    
    // Reserved subdomains
    reserved := []string{"www", "api", "admin", "app", "mail", "ftp"}
    for _, r := range reserved {
        if subdomain == r {
            return false
        }
    }
    
    return true
}
```

## Error Handling

### Common Error Patterns

**Subdomain Conflicts:**
```go
_, err := queries.CreateTenant(ctx, params)
if err != nil {
    if strings.Contains(err.Error(), "duplicate key") &&
       strings.Contains(err.Error(), "subdomain") {
        return c.JSON(409, gin.H{
            "error": "Subdomain already taken",
            "code": "SUBDOMAIN_EXISTS",
        })
    }
}
```

**Invalid Invitations:**
```go
invitation, err := queries.GetInvitationByToken(ctx, token)
if errors.Is(err, pgx.ErrNoRows) {
    return c.JSON(400, gin.H{
        "error": "Invalid or expired invitation",
        "code": "INVALID_INVITATION",
    })
}
```

**Duplicate Invitations:**
```go
_, err := queries.CreateInvitation(ctx, params)
if err != nil {
    if strings.Contains(err.Error(), "idx_invitations_tenant_email_active") {
        return c.JSON(409, gin.H{
            "error": "User already has a pending invitation",
            "code": "INVITATION_EXISTS",
        })
    }
}
```

## Schema Provisioning Integration

### Complete Tenant Setup Flow

```go
func provisionNewTenant(orgData OrganizationData) (*Tenant, error) {
    // 1. Validate data
    if !isValidSubdomain(orgData.Subdomain) {
        return nil, errors.New("invalid subdomain")
    }
    
    // 2. Check availability
    exists, err := queries.CheckSubdomainExists(ctx, orgData.Subdomain)
    if exists {
        return nil, errors.New("subdomain taken")
    }
    
    // 3. Create tenant record
    schemaName := fmt.Sprintf("tenant_%d", generateUniqueID())
    tenant, err := queries.CreateTenant(ctx, db.CreateTenantParams{
        Name:       orgData.Name,
        Subdomain:  orgData.Subdomain,
        SchemaName: schemaName,
    })
    if err != nil {
        return nil, err
    }
    
    // 4. Create database schema
    err = createTenantSchema(ctx, schemaName)
    if err != nil {
        // Rollback tenant record
        queries.DeleteTenant(ctx, tenant.ID)
        return nil, err
    }
    
    // 5. Create initial admin user
    _, err = conn.Exec(ctx, "SET search_path = " + schemaName)
    if err != nil {
        return nil, err
    }
    
    adminUser, err := queries.CreateUser(ctx, db.CreateUserParams{
        Email:        orgData.AdminEmail,
        PasswordHash: hashPassword(orgData.AdminPassword),
        FirstName:    orgData.AdminFirstName,
        LastName:     orgData.AdminLastName,
        Role:         "admin",
    })
    if err != nil {
        return nil, err
    }
    
    return &tenant, nil
}
```

## Related Documentation

- [Global Tenant Schema](../global/Tenants.md) - Complete table definitions
- [SQLC Configuration](../../architecture/sqlc.md) - Code generation setup
- [Database Design](../../architecture/database.md) - Multi-tenant architecture
- [Tenant Service Architecture](../../services/tenant-service.md) - Service implementation