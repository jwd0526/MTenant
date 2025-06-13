# Service Architecture

Detailed technical documentation of the microservices architecture and inter-service communication patterns.

## Architecture Overview

The MTenant CRM uses a microservices architecture with five core services, each responsible for a specific business domain:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   API Gateway   │    │   Load Balancer │
│   (Vue.js)      │◄──►│   (Future)      │◄──►│   (Ingress)     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                ┌───────────────┼───────────────┐
                │               │               │
        ┌───────▼───────┐ ┌─────▼─────┐ ┌─────▼─────┐
        │  Auth Service │ │   Tenant  │ │  Contact  │
        │     :8080     │ │  Service  │ │  Service  │
        └───────────────┘ │   :8081   │ │   :8082   │
                          └───────────┘ └───────────┘
                                │               │
                        ┌───────▼───────┐ ┌─────▼─────┐
                        │     Deal      │ │   Comm    │
                        │   Service     │ │  Service  │
                        │    :8083      │ │   :8084   │
                        └───────────────┘ └───────────┘
                                │
                    ┌───────────┼───────────┐
                    │           │           │
            ┌───────▼───┐ ┌─────▼─────┐ ┌──▼────┐
            │PostgreSQL │ │   NATS    │ │ Redis │
            │   :5433   │ │   :4222   │ │ :6379 │
            └───────────┘ └───────────┘ └───────┘
```

## Service Responsibilities

### Auth Service (`services/auth-service/`)

**Primary Purpose:** User authentication, authorization, and token management

**Core Functions:**
- User registration and password management
- JWT token generation and validation
- Session management and refresh tokens
- Password reset workflows
- Role-based access control

**Database Schema:**
- `users` - User accounts with roles and permissions
- `password_reset_tokens` - Temporary tokens for password recovery

**Key Endpoints:**
```
POST   /api/auth/register      # User registration
POST   /api/auth/login         # Authentication
POST   /api/auth/refresh       # Token refresh
GET    /api/auth/profile       # User profile
POST   /api/auth/logout        # Session termination
POST   /api/auth/forgot        # Password reset request
POST   /api/auth/reset         # Password reset confirmation
```

### Tenant Service (`services/tenant-service/`)

**Primary Purpose:** Multi-tenant organization management and schema provisioning

**Core Functions:**
- Organization registration and setup
- Dynamic schema creation for new tenants
- User invitation and role management
- Tenant configuration and settings
- Subdomain validation and management

**Database Schema:**
- `tenants` - Organization registry (global table)
- `invitations` - Cross-tenant invitation system

**Key Endpoints:**
```
POST   /api/tenants/register   # Organization registration
GET    /api/tenants/current    # Current tenant info
POST   /api/tenants/invite     # Send user invitation
GET    /api/tenants/users      # List tenant users
PUT    /api/tenants/settings   # Update tenant configuration
```

**Schema Creation Process:**
1. Validate organization details and subdomain uniqueness
2. Create tenant record in global registry
3. Execute `CREATE SCHEMA tenant_{id}` 
4. Copy table structure from tenant template
5. Create initial admin user in tenant schema
6. Return tenant context for authentication

### Contact Service (`services/contact-service/`)

**Primary Purpose:** Customer and company data management

**Core Functions:**
- Contact CRUD operations with custom fields
- Company management with hierarchical relationships
- Advanced search and filtering capabilities
- Data import/export functionality
- Contact-company association management

**Database Schema:**
- `contacts` - Individual contact records
- `companies` - Business entities with parent-child support

**Key Endpoints:**
```
POST   /api/contacts           # Create contact
GET    /api/contacts/:id       # Get contact details
PUT    /api/contacts/:id       # Update contact
DELETE /api/contacts/:id       # Soft delete contact
GET    /api/contacts           # List/search contacts
POST   /api/contacts/import    # Bulk import
GET    /api/contacts/export    # Data export

POST   /api/companies          # Create company
GET    /api/companies/:id      # Get company details
GET    /api/companies          # List companies
```

### Deal Service (`services/deal-service/`)

**Primary Purpose:** Sales pipeline and opportunity management

**Core Functions:**
- Deal lifecycle management through configurable stages
- Revenue forecasting and probability tracking
- Sales analytics and performance metrics
- Deal-contact associations
- Pipeline reporting and insights

**Database Schema:**
- `deals` - Sales opportunities with stage tracking
- `deal_contacts` - Many-to-many contact associations

**Key Endpoints:**
```
POST   /api/deals              # Create opportunity
GET    /api/deals/:id          # Get deal details
PUT    /api/deals/:id          # Update deal/change stage
GET    /api/deals              # Pipeline view
GET    /api/deals/analytics    # Revenue analytics
POST   /api/deals/:id/contacts # Associate contacts
```

### Communication Service (`services/communication-service/`)

**Primary Purpose:** Customer interaction tracking and communication workflows

**Core Functions:**
- Activity logging (emails, calls, meetings, notes)
- Email sending with template support
- Interaction timeline and history
- Communication analytics and tracking
- Task management and reminders

**Database Schema:**
- `activities` - All customer interactions
- `email_templates` - Communication templates
- `email_tracking` - Delivery and engagement metrics

**Key Endpoints:**
```
POST   /api/activities         # Log activity
GET    /api/activities         # Activity timeline
POST   /api/emails/send        # Send email
GET    /api/emails/templates   # Email templates
POST   /api/tasks              # Create task/reminder
```

## Inter-Service Communication

### Synchronous Communication (HTTP)

**Direct API Calls:**
- Frontend to services for user interactions
- Service-to-service for immediate data requirements
- Real-time validation and data retrieval

**Example Flow - Contact Creation:**
```
Frontend → Contact Service: POST /api/contacts
Contact Service → Auth Service: GET /api/auth/validate-token
Contact Service → Database: INSERT INTO contacts
Contact Service → Frontend: 201 Created
```

### Asynchronous Communication (NATS)

**Event-Driven Messaging:**
Services publish events for decoupled operations and cross-service notifications.

**Event Patterns:**
```go
// Event structure
type Event struct {
    Type      string                 `json:"type"`
    TenantID  string                 `json:"tenant_id"`
    Timestamp time.Time              `json:"timestamp"`
    Data      map[string]interface{} `json:"data"`
    Actor     string                 `json:"actor"` // User who triggered event
}
```

**Key Events:**

**Auth Service Events:**
- `user.created` - New user registration
- `user.login` - Successful authentication
- `user.password_reset` - Password reset completed

**Tenant Service Events:**
- `tenant.created` - New organization registered
- `tenant.user_invited` - User invitation sent
- `tenant.user_joined` - Invitation accepted

**Contact Service Events:**
- `contact.created` - New contact added
- `contact.updated` - Contact information changed
- `company.created` - New company added

**Deal Service Events:**
- `deal.created` - New opportunity opened
- `deal.stage_changed` - Pipeline progression
- `deal.closed_won` - Successful deal closure
- `deal.closed_lost` - Lost opportunity

**Communication Service Events:**
- `email.sent` - Outbound email
- `email.opened` - Email engagement tracking
- `activity.logged` - Customer interaction recorded

### Event Flow Examples

**User Registration Flow:**
```
1. Frontend → Auth Service: POST /api/auth/register
2. Auth Service → Database: Create user record
3. Auth Service → NATS: Publish user.created event
4. Communication Service ← NATS: Receive user.created
5. Communication Service: Send welcome email
6. Tenant Service ← NATS: Receive user.created  
7. Tenant Service: Setup default user preferences
```

**Deal Closure Flow:**
```
1. Frontend → Deal Service: PUT /api/deals/123 {stage: "closed_won"}
2. Deal Service → Database: Update deal record
3. Deal Service → NATS: Publish deal.closed_won event
4. Communication Service ← NATS: Receive deal.closed_won
5. Communication Service: Log "Deal Closed Won" activity
6. Communication Service: Trigger follow-up email sequence
```

## Service Implementation

### Standard Service Structure

Each service follows this directory pattern:

```
services/{service-name}/
├── cmd/server/main.go          # Application entry point
├── internal/
│   ├── db/                     # SQLC generated code
│   ├── handlers/               # HTTP request handlers
│   ├── middleware/             # Authentication, logging
│   ├── services/               # Business logic
│   └── models/                 # Domain models
├── db/
│   ├── queries/                # SQL query definitions  
│   └── schema/                 # Database schema
├── Dockerfile                  # Container definition
├── go.mod                      # Go module dependencies (includes pkg/database)
└── sqlc.yaml                   # SQLC configuration
```

### Shared Database Package Integration

All services use the shared `pkg/database` package for standardized database connectivity:

```go
// Service main.go pattern
package main

import (
    "context"
    "log"
    
    "crm-platform/pkg/database"
    "crm-platform/auth-service/internal/db"
)

func main() {
    ctx := context.Background()
    
    // Load database configuration from environment
    dbConfig, err := database.LoadConfigFromEnv()
    if err != nil {
        log.Fatal("Failed to load database config:", err)
    }
    
    // Create connection pool with retry logic and monitoring
    dbPool, err := database.NewPool(ctx, dbConfig)
    if err != nil {
        log.Fatal("Failed to create database pool:", err)
    }
    defer dbPool.Close()
    
    // Verify database health on startup
    health := dbPool.HealthCheck(ctx)
    if !health.Healthy {
        log.Fatal("Database health check failed:", health.Error)
    }
    log.Printf("Database connected in %v", health.ResponseTime)
    
    // Create SQLC queries instance with shared pool
    queries := db.New(dbPool)
    
    // Initialize service handlers
    handler := NewHandler(queries, dbPool)
    
    // Start HTTP server
    log.Println("Service starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", handler))
}
```

### Authentication Middleware

All services (except Auth) use token validation middleware:

```go
func AuthMiddleware(authServiceURL string) gin.HandlerFunc {
    return func(c *gin.Context) {
        token := extractTokenFromHeader(c.GetHeader("Authorization"))
        
        // Validate token with Auth Service
        resp, err := http.Get(authServiceURL + "/validate?token=" + token)
        if err != nil || resp.StatusCode != 200 {
            c.JSON(401, gin.H{"error": "Unauthorized"})
            c.Abort()
            return
        }
        
        // Extract tenant context from token
        claims, err := parseTokenClaims(token)
        if err != nil {
            c.JSON(401, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }
        
        // Set tenant context for database queries
        c.Set("tenant_id", claims.TenantID)
        c.Set("user_id", claims.UserID)
        c.Next()
    }
}
```

### Tenant Context Injection

Database queries automatically use tenant context with the shared database pool:

```go
func (h *ContactHandler) GetContact(c *gin.Context) {
    tenantID := c.GetString("tenant_id")
    contactID := c.Param("id")
    
    // Set tenant schema using shared database pool
    schemaName := fmt.Sprintf("tenant_%s", tenantID)
    _, err := h.dbPool.Exec(c.Request.Context(), "SET search_path = $1", schemaName)
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to set tenant context"})
        return
    }
    
    // Query within tenant schema using SQLC queries
    contact, err := h.queries.GetContactByID(c.Request.Context(), contactID)
    if err != nil {
        c.JSON(404, gin.H{"error": "Contact not found"})
        return
    }
    
    c.JSON(200, contact)
}

// Helper function for tenant context management
func (h *Handler) setTenantContext(ctx context.Context, tenantID string) error {
    schemaName := fmt.Sprintf("tenant_%s", tenantID)
    _, err := h.dbPool.Exec(ctx, "SET search_path = $1", schemaName)
    return err
}
```

## Service Configuration

### Environment Variables

Standard configuration across all services:

```bash
# Database
DATABASE_URL=postgresql://admin:admin@localhost:5433/crm-platform

# NATS messaging
NATS_URL=nats://localhost:4222

# Redis caching  
REDIS_URL=redis://localhost:6379

# Service-specific
AUTH_SERVICE_URL=http://auth-service:8080
TENANT_SERVICE_URL=http://tenant-service:8081

# Application
PORT=8080
LOG_LEVEL=info
ENVIRONMENT=development
```

### Service Discovery

In Kubernetes environments, services discover each other via DNS:

```yaml
# Example: Contact Service calling Auth Service
auth_service_url: "http://auth-service.default.svc.cluster.local:8080"
```

## Error Handling

### Standardized Error Responses

```go
type ErrorResponse struct {
    Error   string `json:"error"`
    Code    string `json:"code,omitempty"`
    Details string `json:"details,omitempty"`
}

// Usage
c.JSON(400, ErrorResponse{
    Error: "Validation failed",
    Code: "INVALID_EMAIL",
    Details: "Email format is invalid",
})
```

### Service Error Patterns

**Auth Service Errors:**
- `401 Unauthorized` - Invalid credentials
- `403 Forbidden` - Insufficient permissions
- `409 Conflict` - Email already exists

**Tenant Service Errors:**
- `409 Conflict` - Subdomain already taken
- `400 Bad Request` - Invalid subdomain format
- `402 Payment Required` - Subscription limits exceeded

**Contact/Deal Service Errors:**
- `404 Not Found` - Resource doesn't exist in tenant
- `422 Unprocessable Entity` - Business rule violations
- `400 Bad Request` - Invalid data format

## Health Checks

Each service implements standardized health check endpoints using the shared database package:

```go
func (h *Handler) HealthCheck(c *gin.Context) {
    // Comprehensive database health check with metrics
    dbHealth := h.dbPool.HealthCheck(c.Request.Context())
    
    status := gin.H{
        "status":        "healthy",
        "timestamp":     time.Now(),
        "version":       "1.0.0",
        "database": gin.H{
            "healthy":       dbHealth.Healthy,
            "response_time": dbHealth.ResponseTime.String(),
            "stats":         dbHealth.Stats,
        },
    }
    
    httpStatus := 200
    
    // Check database health
    if !dbHealth.Healthy {
        status["status"] = "unhealthy"
        status["database"].(gin.H)["error"] = dbHealth.Error
        httpStatus = 503
    }
    
    // Check external dependencies (NATS, Redis, etc.)
    if err := h.checkExternalDependencies(); err != nil {
        status["status"] = "degraded"
        status["external_services"] = gin.H{
            "error": err.Error(),
        }
        if httpStatus == 200 {
            httpStatus = 503
        }
    }
    
    c.JSON(httpStatus, status)
}

// Example health check response
{
    "status": "healthy",
    "timestamp": "2024-01-15T10:30:00Z",
    "version": "1.0.0",
    "database": {
        "healthy": true,
        "response_time": "2.5ms",
        "stats": {
            "max_conns": 20,
            "total_conns": 8,
            "idle_conns": 3,
            "acquired_conns": 5
        }
    }
}
```

## Performance Considerations

### Connection Pooling

Connection pooling is handled by the shared `pkg/database` package:

```go
// Database configuration automatically loaded from environment
config, err := database.LoadConfigFromEnv()
if err != nil {
    log.Fatal("Database config error:", err)
}

// Connection pool settings (from pkg/database defaults):
// MaxConns: 20
// MinConns: 5  
// MaxConnLifetime: 60 minutes
// MaxConnIdleTime: 5 minutes
// ConnectTimeout: 30 seconds
// QueryTimeout: 30 seconds

pool, err := database.NewPool(ctx, config)
if err != nil {
    log.Fatal("Database pool error:", err)
}
```

### Caching Strategy

- **User sessions** - Redis with 1-hour TTL
- **Tenant configurations** - Redis with 24-hour TTL  
- **Contact lookups** - Application-level caching with 15-minute TTL
- **API responses** - HTTP caching headers for static data

### Database Performance

- **Connection pooling** prevents connection exhaustion
- **Query timeouts** protect against long-running queries
- **Prepared statements** via SQLC reduce parsing overhead
- **Index usage** verified for all common query patterns

## Testing Strategy

### Unit Tests

Each service includes comprehensive tests:

```bash
# Run service tests
cd services/auth-service && go test -v ./...
cd services/tenant-service && go test -v ./...
# etc.
```

### Integration Tests

Test inter-service communication:

```go
func TestContactCreationFlow(t *testing.T) {
    // Setup test tenant
    tenant := createTestTenant()
    
    // Authenticate user
    token := authenticateTestUser(tenant.ID)
    
    // Create contact
    contact := createTestContact(token, tenant.Schema)
    
    // Verify event published
    verifyEventPublished("contact.created", contact.ID)
}
```

## Deployment Architecture

### Container Configuration

Each service runs in its own container:

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /service ./cmd/server

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=builder /service /service
EXPOSE 8080
CMD ["/service"]
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: auth-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: auth-service
  template:
    metadata:
      labels:
        app: auth-service
    spec:
      containers:
      - name: auth-service
        image: auth-service:latest
        ports:
        - containerPort: 8080
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: database-secret
              key: url
```

## Related Documentation

- [SQLC Implementation](./sqlc.md) - Database access patterns
- [Database Design](./database.md) - Multi-tenant data architecture  
- [Development Setup](../development/setup.md) - Local development environment