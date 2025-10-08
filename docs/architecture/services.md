# Multi-Tenant CRM - Service Architecture

**Last Updated:** 2025-10-08\
*Define service architecture*

**Purpose:** This document defines the API contracts, data ownership, dependencies, and implementation patterns for all microservices in the platform. It ensures architectural soundness by mapping what each service exposes (DTOs), what it stores (Internal), and what it consumes (External).

---

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

---

## 1. tenant-service

### General
Manages tenant lifecycle including registration, schema provisioning, subdomain routing, and cross-tenant invitation system. This is a foundational service that other services depend on for tenant validation.

**Technology:** Go + PostgreSQL (global tables, not in tenant schemas)

### DTO (Public API Contracts)

#### TenantBasicInfo
```go
type TenantBasicInfo struct {
    ID         string `json:"id"`          // ULID format (26 chars)
    Name       string `json:"name"`        // Organization name
    Subdomain  string `json:"subdomain"`   // URL subdomain (e.g., "acme" -> acme.crm.com)
    SchemaName string `json:"schema_name"` // PostgreSQL schema name
}
```
**Used by:** auth-service (tenant validation), all services (tenant context validation)

#### TenantDetails
```go
type TenantDetails struct {
    ID         string    `json:"id"`
    Name       string    `json:"name"`
    Subdomain  string    `json:"subdomain"`
    SchemaName string    `json:"schema_name"`
    Status     string    `json:"status"`     // active, suspended, pending
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
}
```
**Used by:** Frontend, admin dashboard

#### InvitationInfo
```go
type InvitationInfo struct {
    ID        string    `json:"id"`          // ULID format
    TenantID  string    `json:"tenant_id"`   // ULID reference
    Email     string    `json:"email"`
    Role      string    `json:"role"`        // admin, manager, sales_rep, viewer
    Token     string    `json:"token"`
    ExpiresAt time.Time `json:"expires_at"`
    Status    string    `json:"status"`      // pending, accepted, expired
}
```
**Used by:** auth-service (user registration via invitation)

### Internal (Database Tables)

**Global Tables (NOT in tenant schemas):**

```sql
-- Tenant registry
CREATE TABLE tenants (
    id TEXT PRIMARY KEY,              -- ULID format (26 chars)
    name VARCHAR(255) NOT NULL,
    subdomain VARCHAR(63) UNIQUE NOT NULL,
    schema_name VARCHAR(63) UNIQUE NOT NULL,
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Cross-tenant invitations
CREATE TABLE invitations (
    id TEXT PRIMARY KEY,              -- ULID format
    tenant_id TEXT REFERENCES tenants(id),
    email VARCHAR(254) NOT NULL,
    role VARCHAR(20) NOT NULL,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    accepted_at TIMESTAMPTZ,
    invited_by INTEGER,               -- User ID from tenant schema
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
```

**Owned Data:**
- Tenant registration and metadata
- Subdomain to schema_name mapping
- Cross-tenant invitation tokens
- Tenant status (active/suspended)

### External (Dependencies)

**None** - This is a foundational service. Other services depend on it, but it doesn't consume data from other services.

---

## 2. auth-service

### General
User authentication, JWT token management, session handling, and password security. Provides authentication middleware for all other services.

**Technology:** Go + PostgreSQL (tenant-specific tables)

### DTO (Public API Contracts)

#### UserBasicInfo
```go
type UserBasicInfo struct {
    ID        int32  `json:"id"`
    Email     string `json:"email"`
    FirstName string `json:"first_name"`
    LastName  string `json:"last_name"`
    Role      string `json:"role"`      // admin, manager, sales_rep, viewer
}
```
**Used by:** ALL services for enriching created_by, owner_id, assigned_to references
**Critical:** This is the MOST USED DTO across the platform

#### UserDetails
```go
type UserDetails struct {
    ID            int32     `json:"id"`
    Email         string    `json:"email"`
    FirstName     string    `json:"first_name"`
    LastName      string    `json:"last_name"`
    Role          string    `json:"role"`
    Status        string    `json:"status"`        // active, inactive, pending
    EmailVerified bool      `json:"email_verified"`
    LastLoginAt   time.Time `json:"last_login_at"`
    CreatedAt     time.Time `json:"created_at"`
}
```
**Used by:** Frontend user management UI

#### JWTToken
```go
type JWTToken struct {
    AccessToken  string    `json:"access_token"`
    RefreshToken string    `json:"refresh_token"`
    ExpiresAt    time.Time `json:"expires_at"`
    TokenType    string    `json:"token_type"`    // "Bearer"
}
```
**Used by:** Frontend, all services for authentication

#### AuthContext
```go
type AuthContext struct {
    UserID   int32  `json:"user_id"`
    TenantID string `json:"tenant_id"`  // ULID from JWT claims
    Email    string `json:"email"`
    Role     string `json:"role"`
}
```
**Used by:** Middleware in all services (extracted from JWT)

### Internal (Database Tables)

**Tenant-specific tables:**

```sql
-- User accounts (within each tenant schema)
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(254) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    role VARCHAR(20) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    email_verified BOOLEAN DEFAULT false,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(id)
);

-- Password reset tokens (within tenant schema)
CREATE TABLE password_reset_tokens (
    id TEXT PRIMARY KEY,              -- ULID format
    user_id INTEGER REFERENCES users(id),
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
```

**Owned Data:**
- User credentials (hashed passwords)
- User profile information
- Authentication tokens
- Password reset flows
- User roles and permissions

### External (Dependencies)

#### From tenant-service:
- **TenantBasicInfo** - Validate tenant_id from JWT claims during authentication

---

## 3. contact-service

### General
Company and contact relationship management with hierarchical company support, custom fields, and comprehensive search capabilities.

**Technology:** Go + PostgreSQL (tenant-specific tables)

### DTO (Public API Contracts)

#### CompanyBasicInfo
```go
type CompanyBasicInfo struct {
    ID   int32  `json:"id"`
    Name string `json:"name"`
}
```
**Used by:**
- deal-service (displays company name on deals)
- communication-service (displays company name on activities)

#### CompanyDetails
```go
type CompanyDetails struct {
    ID              int32     `json:"id"`
    Name            string    `json:"name"`
    Domain          string    `json:"domain"`
    Industry        string    `json:"industry"`
    SizeCategory    string    `json:"size_category"`
    ParentCompanyID *int32    `json:"parent_company_id"`
    Address         Address   `json:"address"`
    Phone           string    `json:"phone"`
    Website         string    `json:"website"`
    CustomFields    JSONMap   `json:"custom_fields"`
    CreatedAt       time.Time `json:"created_at"`
    CreatedBy       int32     `json:"created_by"`
}

type Address struct {
    Street     string `json:"street"`
    City       string `json:"city"`
    State      string `json:"state"`
    Country    string `json:"country"`
    PostalCode string `json:"postal_code"`
}
```
**Used by:** Frontend company management UI

#### ContactBasicInfo
```go
type ContactBasicInfo struct {
    ID        int32  `json:"id"`
    FirstName string `json:"first_name"`
    LastName  string `json:"last_name"`
    Email     string `json:"email"`
}
```
**Used by:**
- deal-service (displays contact name and email on deals)
- communication-service (displays contact info on activities)

#### ContactDetails
```go
type ContactDetails struct {
    ID           int32     `json:"id"`
    FirstName    string    `json:"first_name"`
    LastName     string    `json:"last_name"`
    Email        string    `json:"email"`
    Phone        string    `json:"phone"`
    JobTitle     string    `json:"job_title"`
    CompanyID    *int32    `json:"company_id"`
    CompanyName  string    `json:"company_name"`    // Enriched from companies table
    OwnerID      *int32    `json:"owner_id"`
    OwnerName    string    `json:"owner_name"`      // Enriched from users
    Status       string    `json:"status"`
    Source       string    `json:"source"`
    Address      Address   `json:"address"`
    CustomFields JSONMap   `json:"custom_fields"`
    Notes        string    `json:"notes"`
    CreatedAt    time.Time `json:"created_at"`
}
```
**Used by:** Frontend contact management UI

### Internal (Database Tables)

```sql
-- Companies table
CREATE TABLE companies (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    domain VARCHAR(253),
    industry VARCHAR(100),
    size_category VARCHAR(50),
    parent_company_id INTEGER REFERENCES companies(id),
    street_address VARCHAR(255),
    city VARCHAR(100),
    state VARCHAR(100),
    country VARCHAR(100),
    postal_code VARCHAR(20),
    phone VARCHAR(50),
    website VARCHAR(255),
    custom_fields JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(id)
);

-- Contacts table
CREATE TABLE contacts (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(254),
    phone VARCHAR(50),
    job_title VARCHAR(100),
    company_id INTEGER REFERENCES companies(id),
    owner_id INTEGER REFERENCES users(id),
    status VARCHAR(50) DEFAULT 'lead',
    source VARCHAR(100),
    street_address VARCHAR(255),
    city VARCHAR(100),
    state VARCHAR(100),
    country VARCHAR(100),
    postal_code VARCHAR(20),
    custom_fields JSONB DEFAULT '{}',
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(id)
);
```

**Owned Data:**
- Company master records
- Contact master records
- Company hierarchies (parent-child relationships)
- Contact-to-company associations
- Custom fields for companies and contacts

### External (Dependencies)

#### From auth-service:
- **UserBasicInfo** - Enrich owner_id, created_by fields with user names for display

---

## 4. deal-service ✅ (IMPLEMENTED)

### General
Sales pipeline management, opportunity tracking, deal-contact associations, revenue forecasting, and pipeline analytics.

**Technology:** Go + PostgreSQL (tenant-specific tables)

### DTO (Public API Contracts)

#### DealBasicInfo
```go
type DealBasicInfo struct {
    ID    int32   `json:"id"`
    Title string  `json:"title"`
    Stage string  `json:"stage"`
    Value float64 `json:"value"`
}
```
**Used by:**
- communication-service (links activities to deals)
- Future reporting services

#### DealDetails (Current Implementation)
```go
type DealResponse struct {
    ID                int32      `json:"id"`
    Title             string     `json:"title"`
    Value             *float64   `json:"value"`
    Probability       *float64   `json:"probability"`
    Stage             string     `json:"stage"`
    PrimaryContactID  *int32     `json:"primary_contact_id"`
    CompanyID         *int32     `json:"company_id"`
    OwnerID           *int32     `json:"owner_id"`
    ExpectedCloseDate *time.Time `json:"expected_close_date"`
    ActualCloseDate   *time.Time `json:"actual_close_date"`
    DealSource        *string    `json:"deal_source"`
    Description       *string    `json:"description"`
    CreatedAt         time.Time  `json:"created_at"`
    UpdatedAt         time.Time  `json:"updated_at"`
    CreatedBy         *int32     `json:"created_by"`

    // Enriched fields (from external services)
    PrimaryContactName *string    `json:"primary_contact_name"`
    CompanyName        *string    `json:"company_name"`
    OwnerName          *string    `json:"owner_name"`

    // Calculated fields
    DealAge            int        `json:"deal_age_days"`
    DaysUntilClose     *int       `json:"days_until_close"`
    WeightedValue      *float64   `json:"weighted_value"`
}
```
**Used by:** Frontend deal management UI

#### PipelineStats
```go
type PipelineViewResponse struct {
    Stages []PipelineStage `json:"stages"`
    Totals PipelineTotals  `json:"totals"`
}

type PipelineStage struct {
    Stage         string         `json:"stage"`
    DealCount     int            `json:"deal_count"`
    TotalValue    float64        `json:"total_value"`
    WeightedValue float64        `json:"weighted_value"`
    Deals         []DealResponse `json:"deals"`
}

type PipelineTotals struct {
    TotalDeals         int     `json:"total_deals"`
    TotalValue         float64 `json:"total_value"`
    TotalWeightedValue float64 `json:"total_weighted_value"`
}
```
**Used by:** Frontend pipeline dashboard, reporting

### Internal (Database Tables)

```sql
-- Deals table
CREATE TABLE deals (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    value NUMERIC(15,2),
    currency VARCHAR(3) DEFAULT 'USD',
    stage VARCHAR(100) NOT NULL,
    probability INTEGER DEFAULT 0,
    expected_close_date DATE,
    actual_close_date DATE,
    owner_id INTEGER,                  -- References users (auth-service)
    company_id INTEGER,                -- References companies (contact-service)
    primary_contact_id INTEGER,        -- References contacts (contact-service)
    source VARCHAR(100),
    close_reason VARCHAR(255),
    custom_fields JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER                 -- References users (auth-service)
);

-- Deal-to-contact associations (many-to-many)
CREATE TABLE deal_contacts (
    deal_id INTEGER REFERENCES deals(id),
    contact_id INTEGER,                -- References contacts (contact-service)
    role VARCHAR(100),                 -- decision_maker, influencer, user, etc.
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (deal_id, contact_id)
);
```

**Owned Data:**
- Deal records and pipeline stages
- Deal-specific metadata (probability, close dates)
- Deal-to-contact associations with roles
- Revenue forecasting calculations

### External (Dependencies)

#### From contact-service:
- **CompanyBasicInfo** (id, name) - Display company name on deals
- **ContactBasicInfo** (id, first_name, last_name, email) - Display contact names on deals

#### From auth-service:
- **UserBasicInfo** (id, first_name, last_name) - Display owner names and created_by names

**Current Implementation Note:** deal-service currently stores company_id, contact_id, owner_id as INTEGER references. It calls external service APIs to enrich these with names for display purposes.

---

## 5. communication-service

### General
Activity tracking (emails, calls, meetings, notes), email template management, email delivery tracking, and task/reminder system.

**Technology:** Go + PostgreSQL (tenant-specific tables)

### DTO (Public API Contracts)

#### ActivityInfo
```go
type ActivityInfo struct {
    ID          int32     `json:"id"`
    Type        string    `json:"type"`        // email, call, meeting, note, task
    Subject     string    `json:"subject"`
    Description string    `json:"description"`
    ContactID   *int32    `json:"contact_id"`
    CompanyID   *int32    `json:"company_id"`
    DealID      *int32    `json:"deal_id"`
    OwnerID     *int32    `json:"owner_id"`
    CompletedAt time.Time `json:"completed_at"`
    CreatedAt   time.Time `json:"created_at"`
}
```
**Used by:** Frontend activity timeline, reporting

#### ActivityDetails
```go
type ActivityDetails struct {
    ID            int32     `json:"id"`
    Type          string    `json:"type"`
    Subject       string    `json:"subject"`
    Description   string    `json:"description"`
    Status        string    `json:"status"`
    Direction     string    `json:"direction"`     // inbound, outbound
    DurationMins  int       `json:"duration_minutes"`
    ScheduledAt   time.Time `json:"scheduled_at"`
    CompletedAt   time.Time `json:"completed_at"`

    // Foreign key IDs
    ContactID     *int32    `json:"contact_id"`
    CompanyID     *int32    `json:"company_id"`
    DealID        *int32    `json:"deal_id"`
    OwnerID       *int32    `json:"owner_id"`

    // Enriched data (from external services)
    ContactName   string    `json:"contact_name"`
    CompanyName   string    `json:"company_name"`
    DealTitle     string    `json:"deal_title"`
    OwnerName     string    `json:"owner_name"`

    CustomFields  JSONMap   `json:"custom_fields"`
    CreatedAt     time.Time `json:"created_at"`
}
```
**Used by:** Frontend activity detail view

#### TaskInfo
```go
type TaskInfo struct {
    ID          int32     `json:"id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Type        string    `json:"type"`        // task, reminder, follow_up
    Status      string    `json:"status"`      // pending, in_progress, completed
    Priority    string    `json:"priority"`    // low, normal, high, urgent
    DueDate     time.Time `json:"due_date"`
    AssignedTo  int32     `json:"assigned_to"`
    ContactID   *int32    `json:"contact_id"`
    DealID      *int32    `json:"deal_id"`
    CreatedAt   time.Time `json:"created_at"`
}
```
**Used by:** Frontend task management, notifications

#### EmailTemplateInfo
```go
type EmailTemplateInfo struct {
    ID       int32   `json:"id"`
    Name     string  `json:"name"`
    Subject  string  `json:"subject"`
    Category string  `json:"category"`    // welcome, follow_up, proposal
    IsActive bool    `json:"is_active"`
}
```
**Used by:** Frontend email composition, automation workflows

### Internal (Database Tables)

```sql
-- Activities table
CREATE TABLE activities (
    id SERIAL PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    subject VARCHAR(255) NOT NULL,
    description TEXT,
    due_date TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    duration_minutes INTEGER,
    contact_id INTEGER,              -- References contacts (contact-service)
    company_id INTEGER,              -- References companies (contact-service)
    deal_id INTEGER,                 -- References deals (deal-service)
    owner_id INTEGER,                -- References users (auth-service)
    custom_fields JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER               -- References users (auth-service)
);

-- Email templates
CREATE TABLE email_templates (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    subject VARCHAR(255) NOT NULL,
    body_html TEXT NOT NULL,
    body_text TEXT,
    category VARCHAR(100),
    variables JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER
);

-- Email tracking
CREATE TABLE email_tracking (
    id SERIAL PRIMARY KEY,
    activity_id INTEGER REFERENCES activities(id),
    email_address VARCHAR(255) NOT NULL,
    template_id INTEGER REFERENCES email_templates(id),
    message_id VARCHAR(255) UNIQUE,
    status VARCHAR(50) DEFAULT 'sent',
    sent_at TIMESTAMPTZ,
    delivered_at TIMESTAMPTZ,
    opened_at TIMESTAMPTZ,
    clicked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Tasks
CREATE TABLE tasks (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) DEFAULT 'task',
    status VARCHAR(50) DEFAULT 'pending',
    priority VARCHAR(20) DEFAULT 'normal',
    due_date TIMESTAMPTZ,
    reminder_at TIMESTAMPTZ,
    contact_id INTEGER,              -- References contacts (contact-service)
    deal_id INTEGER,                 -- References deals (deal-service)
    assigned_to INTEGER,             -- References users (auth-service)
    completed_at TIMESTAMPTZ,
    custom_fields JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER
);
```

**Owned Data:**
- Activity records (emails, calls, meetings, notes)
- Email templates and delivery tracking
- Tasks and reminders
- Communication timelines
- Email engagement metrics (opens, clicks)

### External (Dependencies)

#### From contact-service:
- **CompanyBasicInfo** (id, name) - Display company name on activities
- **ContactBasicInfo** (id, first_name, last_name, email) - Display contact name on activities

#### From deal-service:
- **DealBasicInfo** (id, title, stage, value) - Display deal context on activities

#### From auth-service:
- **UserBasicInfo** (id, first_name, last_name) - Display owner/assignee names

---

## Cross-Service Dependency Matrix

| Service → Depends On ↓ | tenant-service | auth-service | contact-service | deal-service | communication-service |
|------------------------|----------------|--------------|-----------------|--------------|----------------------|
| **tenant-service**     | -              | ❌           | ❌              | ❌           | ❌                   |
| **auth-service**       | ✅ TenantBasicInfo | -        | ❌              | ❌           | ❌                   |
| **contact-service**    | ❌             | ✅ UserBasicInfo | -          | ❌           | ❌                   |
| **deal-service**       | ❌             | ✅ UserBasicInfo | ✅ CompanyBasicInfo, ContactBasicInfo | - | ❌ |
| **communication-service** | ❌          | ✅ UserBasicInfo | ✅ CompanyBasicInfo, ContactBasicInfo | ✅ DealBasicInfo | - |

**Dependency Order (Implementation Priority):**
1. tenant-service (foundational - no dependencies)
2. auth-service (depends only on tenant-service)
3. contact-service (depends on auth-service)
4. deal-service ✅ (already implemented - depends on auth + contact)
5. communication-service (depends on all others)

---

## API Client Implementation Pattern

All service API clients are centralized in the global `pkg/clients/` directory for easy discovery and maintenance.

### Directory Structure

```
pkg/
└── clients/
    ├── tenant/
    │   ├── client.go      # TenantClient - HTTP client for tenant-service
    │   ├── dto.go         # TenantBasicInfo, InvitationInfo DTOs
    │   └── mock.go        # Mock client for testing
    ├── auth/
    │   ├── client.go      # AuthClient - HTTP client for auth-service
    │   ├── dto.go         # UserBasicInfo, JWTToken DTOs
    │   └── mock.go        # Mock client for testing
    ├── contact/
    │   ├── client.go      # ContactClient - HTTP client for contact-service
    │   ├── dto.go         # CompanyBasicInfo, ContactBasicInfo DTOs
    │   └── mock.go        # Mock client for testing
    ├── deal/
    │   ├── client.go      # DealClient - HTTP client for deal-service
    │   ├── dto.go         # DealBasicInfo DTOs
    │   └── mock.go        # Mock client for testing
    └── communication/
        ├── client.go      # CommunicationClient - HTTP client
        ├── dto.go         # ActivityInfo, TaskInfo DTOs
        └── mock.go        # Mock client for testing
```

### Client Ownership

While clients are located in `pkg/clients/`, **each service team owns their client implementation**:

| Client Package | Owned By | Maintains |
|----------------|----------|-----------|
| `pkg/clients/tenant/` | tenant-service team | DTOs, client logic, mocks |
| `pkg/clients/auth/` | auth-service team | DTOs, client logic, mocks |
| `pkg/clients/contact/` | contact-service team | DTOs, client logic, mocks |
| `pkg/clients/deal/` | deal-service team | DTOs, client logic, mocks |
| `pkg/clients/communication/` | communication-service team | DTOs, client logic, mocks |

**Benefit:** In a monorepo, having all clients in one location makes cross-service contracts highly visible while still maintaining clear ownership.

### Example Client Implementation

**pkg/clients/contact/dto.go:**
```go
package contact

// Lightweight DTO - only what other services need
type CompanyBasicInfo struct {
    ID   int32  `json:"id"`
    Name string `json:"name"`
}

type ContactBasicInfo struct {
    ID        int32  `json:"id"`
    FirstName string `json:"first_name"`
    LastName  string `json:"last_name"`
    Email     string `json:"email"`
}
```

**pkg/clients/contact/client.go:**
```go
package contact

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
)

type Client struct {
    baseURL    string
    httpClient *http.Client
}

func NewClient(baseURL string) *Client {
    return &Client{
        baseURL:    baseURL,
        httpClient: &http.Client{Timeout: 10 * time.Second},
    }
}

func (c *Client) GetCompanyBasicInfo(ctx context.Context, tenantID string, companyID int32) (*CompanyBasicInfo, error) {
    url := fmt.Sprintf("%s/api/v1/companies/%d/basic", c.baseURL, companyID)

    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    req.Header.Set("X-Tenant-ID", tenantID)

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var company CompanyBasicInfo
    json.NewDecoder(resp.Body).Decode(&company)
    return &company, nil
}

func (c *Client) GetContactBasicInfo(ctx context.Context, tenantID string, contactID int32) (*ContactBasicInfo, error) {
    url := fmt.Sprintf("%s/api/v1/contacts/%d/basic", c.baseURL, contactID)

    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    req.Header.Set("X-Tenant-ID", tenantID)

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var contact ContactBasicInfo
    json.NewDecoder(resp.Body).Decode(&contact)
    return &contact, nil
}
```

### Example Client Usage in deal-service

**services/deal-service/internal/service/deal_service.go:**

```go
package service

import (
    "crm-platform/pkg/clients/contact"
    "crm-platform/pkg/clients/auth"
)

type DealService struct {
    contactClient *contact.Client
    authClient    *auth.Client
    tenantID      string
}

func NewDealService(contactURL, authURL, tenantID string) *DealService {
    return &DealService{
        contactClient: contact.NewClient(contactURL),
        authClient:    auth.NewClient(authURL),
        tenantID:      tenantID,
    }
}

func (s *DealService) EnrichDeal(ctx context.Context, deal *Deal) error {
    // Enrich with company name from contact-service
    if deal.CompanyID != nil {
        company, err := s.contactClient.GetCompanyBasicInfo(ctx, s.tenantID, *deal.CompanyID)
        if err == nil {
            deal.CompanyName = &company.Name
        }
    }

    // Enrich with contact name from contact-service
    if deal.PrimaryContactID != nil {
        contact, err := s.contactClient.GetContactBasicInfo(ctx, s.tenantID, *deal.PrimaryContactID)
        if err == nil {
            fullName := contact.FirstName + " " + contact.LastName
            deal.PrimaryContactName = &fullName
        }
    }

    // Enrich with owner name from auth-service
    if deal.OwnerID != nil {
        user, err := s.authClient.GetUserBasicInfo(ctx, s.tenantID, *deal.OwnerID)
        if err == nil {
            ownerName := user.FirstName + " " + user.LastName
            deal.OwnerName = &ownerName
        }
    }

    return nil
}
```

### Testing with Mock Clients

**pkg/clients/contact/mock.go:**

```go
package contact

import "context"

type MockClient struct {
    GetCompanyBasicInfoFunc func(ctx context.Context, tenantID string, id int32) (*CompanyBasicInfo, error)
    GetContactBasicInfoFunc func(ctx context.Context, tenantID string, id int32) (*ContactBasicInfo, error)
}

func (m *MockClient) GetCompanyBasicInfo(ctx context.Context, tenantID string, id int32) (*CompanyBasicInfo, error) {
    if m.GetCompanyBasicInfoFunc != nil {
        return m.GetCompanyBasicInfoFunc(ctx, tenantID, id)
    }
    return &CompanyBasicInfo{ID: id, Name: "Mock Company"}, nil
}

func (m *MockClient) GetContactBasicInfo(ctx context.Context, tenantID string, id int32) (*ContactBasicInfo, error) {
    if m.GetContactBasicInfoFunc != nil {
        return m.GetContactBasicInfoFunc(ctx, tenantID, id)
    }
    return &ContactBasicInfo{ID: id, FirstName: "Mock", LastName: "Contact"}, nil
}
```

**Usage in tests:**

```go
func TestDealService_EnrichDeal(t *testing.T) {
    mockContact := &contact.MockClient{
        GetCompanyBasicInfoFunc: func(ctx context.Context, tenantID string, id int32) (*contact.CompanyBasicInfo, error) {
            return &contact.CompanyBasicInfo{ID: 456, Name: "Acme Corp"}, nil
        },
    }

    service := &DealService{
        contactClient: mockContact,
        tenantID:      "test-tenant",
    }

    deal := &Deal{CompanyID: int32Ptr(456)}
    service.EnrichDeal(context.Background(), deal)

    assert.Equal(t, "Acme Corp", *deal.CompanyName)
}
```

---

## Ideal Service Structure

Every service in the platform should follow this standardized directory structure for consistency and maintainability.

### Directory Layout

```
services/
└── [service-name]/
    ├── cmd/
    │   └── server/
    │       └── main.go           # Application entry point
    ├── internal/                 # Private implementation (not importable)
    │   ├── handlers/             # HTTP request handlers
    │   │   ├── [resource].go     # CRUD handlers for main resource
    │   │   └── system.go         # Health check, metrics handlers
    │   ├── service/              # Business logic layer
    │   │   └── [resource]_service.go  # Domain logic, orchestration
    │   ├── repository/           # Data access layer (optional)
    │   │   └── [resource]_repo.go     # Database abstraction
    │   ├── models/               # Request/response DTOs
    │   │   ├── requests.go       # API request models
    │   │   └── responses.go      # API response models
    │   ├── db/                   # SQLC generated code
    │   │   ├── models.go         # Database models (SQLC)
    │   │   ├── querier.go        # Query interface (SQLC)
    │   │   └── *.sql.go          # Query implementations (SQLC)
    │   └── config/               # Service-specific configuration
    │       └── config.go         # Config loading and validation
    ├── db/
    │   ├── queries/              # SQL queries for SQLC
    │   │   └── *.sql             # Query definitions
    │   └── schema/               # Database schema definitions
    │       └── *.sql             # Table definitions
    ├── tests/
    │   ├── api/                  # API integration tests
    │   ├── unit/                 # Unit tests
    │   ├── fixtures/             # Test data fixtures
    │   └── helpers/              # Test utilities
    ├── .env.example              # Environment variable template
    ├── Dockerfile                # Container definition
    ├── go.mod                    # Go module definition
    ├── go.sum                    # Go module checksums
    ├── sqlc.yaml                 # SQLC configuration
    └── README.md                 # Service documentation
```

### Directory Descriptions

#### `cmd/server/`
**Purpose:** Application entry point and server initialization.

**Contents:**
- `main.go` - Bootstraps the service, sets up dependencies, starts HTTP server
- Initializes database connections
- Configures middleware stack
- Registers routes
- Handles graceful shutdown

**Responsibilities:**
- Keep minimal - only wiring/configuration
- No business logic
- No request handling

---

#### `internal/handlers/`
**Purpose:** HTTP request/response handling - thin layer that coordinates requests.

**Contents:**
- One file per resource (e.g., `deals.go`, `companies.go`)
- `system.go` for health checks and metrics

**Responsibilities:**
- Parse HTTP requests (JSON binding, URL params)
- Call service layer for business logic
- Return HTTP responses (JSON serialization, status codes)
- Error handling and HTTP error formatting

**Anti-patterns:**
- ❌ No database queries directly in handlers
- ❌ No business logic calculations
- ❌ No external API calls

---

#### `internal/service/`
**Purpose:** Business logic and orchestration layer.

**Contents:**
- One file per domain concept (e.g., `deal_service.go`)
- Orchestrates multiple operations
- Calls external services via API clients

**Responsibilities:**
- Business logic and validations
- Coordinating multiple data sources
- Calling external services (via `pkg/clients/`)
- Transaction management
- Complex calculations
- Data enrichment

**Example:**
```go
func (s *DealService) EnrichDeal(ctx context.Context, deal *Deal) error {
    // Business logic: enrich deal with external data
    if deal.CompanyID != nil {
        company, _ := s.contactClient.GetCompanyBasicInfo(ctx, *deal.CompanyID)
        deal.CompanyName = company.Name
    }
    return nil
}
```

---

#### `internal/repository/` (Optional)
**Purpose:** Data access abstraction layer - wraps database operations.

**Contents:**
- Repository interfaces and implementations
- Wraps SQLC queries with domain logic

**Responsibilities:**
- Abstract database operations
- Provide clean interface for service layer
- Handle transaction management
- Database-specific error handling

**When to use:**
- Complex database operations
- Need for multiple queries in one operation
- Want to mock database for testing
- Abstractions that simplify service layer

**When to skip:**
- Simple CRUD operations (use SQLC directly in service)
- Service is very small

---

#### `internal/models/`
**Purpose:** API contracts - request and response structures.

**Contents:**
- `requests.go` - Structures for incoming HTTP requests
- `responses.go` - Structures for outgoing HTTP responses

**Responsibilities:**
- Define API shape (what clients send/receive)
- Validation tags for input
- JSON serialization tags
- API versioning

**Distinction from `internal/db/models.go`:**
- API models (this directory) - External contract
- DB models (SQLC generated) - Internal database representation
- These are often different!

---

#### `internal/db/`
**Purpose:** Generated SQLC code - database layer.

**Contents:**
- `models.go` - Database table models (SQLC generated)
- `querier.go` - Query interface (SQLC generated)
- `*.sql.go` - Query implementations (SQLC generated)

**Responsibilities:**
- Type-safe database operations
- SQL query execution
- Data mapping (DB ↔ Go structs)

**Important:**
- ⚠️ **Never edit these files manually** - regenerated by SQLC
- Import these in service layer, not in handlers

---

#### `internal/config/`
**Purpose:** Service-specific configuration management.

**Contents:**
- `config.go` - Configuration structures and loading logic

**Responsibilities:**
- Environment variable parsing
- Configuration validation
- Default values
- Service-specific settings

**Note:** Use `pkg/config/` for platform-wide configuration utilities.

---

#### `db/queries/`
**Purpose:** SQL query definitions for SQLC.

**Contents:**
- `*.sql` files with named queries
- SQLC annotations (e.g., `-- name: GetDealByID :one`)

**Responsibilities:**
- Define all SQL queries
- Use prepared statement syntax
- Leverage PostgreSQL features

---

#### `db/schema/`
**Purpose:** Database table definitions.

**Contents:**
- `*.sql` files with CREATE TABLE statements
- Index definitions
- Constraint definitions

**Responsibilities:**
- Define tenant-specific table schemas
- Maintain schema evolution
- Document database structure

**Note:** These are templates - actual tables created in each tenant schema.

---

#### `tests/`
**Purpose:** All test code.

**Subdirectories:**
- `api/` - Integration tests hitting HTTP endpoints
- `unit/` - Unit tests for business logic
- `fixtures/` - Reusable test data
- `helpers/` - Test utilities and setup code

**Responsibilities:**
- Comprehensive test coverage
- Test tenant isolation
- Mock external dependencies
- Verify API contracts

---

### Layer Communication Pattern

```
HTTP Request
    ↓
Handler (parse request)
    ↓
Service (business logic)
    ↓
Repository (optional - data access)
    ↓
SQLC Queries (database)
    ↓
PostgreSQL (tenant schema)
```

**External Dependencies:**
```
Service Layer
    ↓
pkg/clients/[service]
    ↓
HTTP Call to other service
```

---

### Key Principles

1. **Handlers are thin** - Only HTTP concerns
2. **Services contain logic** - Business rules and orchestration
3. **SQLC handles data** - Type-safe database operations
4. **Models define contracts** - API requests/responses
5. **Tests verify behavior** - Comprehensive coverage

---

## Data Ownership Principles

1. **Each service owns its master data exclusively**
   - contact-service owns companies and contacts
   - deal-service owns deals and deal_contacts
   - auth-service owns users

2. **Foreign keys are stored as INTEGER IDs only**
   - Never duplicate master data
   - Store only the reference ID
   - Enrich via API calls when needed for display

3. **Junction tables belong to the service that owns the relationship**
   - `deal_contacts` belongs to deal-service (deal-specific context)
   - Even though it references contacts, the RELATIONSHIP is owned by deals

4. **DTOs are lightweight and focused**
   - Expose only what dependent services actually need
   - Don't expose entire database models
   - Keep DTOs stable to avoid breaking changes

---

## Validation Checklist

✅ **deal-service needs from contact-service:**
- CompanyBasicInfo: id ✅, name ✅
- ContactBasicInfo: id ✅, first_name ✅, last_name ✅, email ✅

✅ **deal-service needs from auth-service:**
- UserBasicInfo: id ✅, first_name ✅, last_name ✅

✅ **communication-service needs from deal-service:**
- DealBasicInfo: id ✅, title ✅, stage ✅, value ✅

✅ **All services have clear data ownership boundaries**

✅ **No circular dependencies exist**

✅ **DTOs expose all required fields for dependent services**

---

## Inter-Service Communication Patterns

### Synchronous Communication (HTTP)

Services communicate via RESTful HTTP APIs for immediate data requirements:

**Direct API Calls:**
- Frontend to services for user interactions
- Service-to-service for data enrichment (via `pkg/clients/`)
- Real-time validation and lookups

**Example Flow - Deal Enrichment:**
```
1. Frontend → Deal Service: GET /api/deals/123
2. Deal Service → Database: Query deal record
3. Deal Service → Contact Service: GET /api/companies/{id}/basic (via pkg/clients/contact)
4. Deal Service → Auth Service: GET /api/users/{id}/basic (via pkg/clients/auth)
5. Deal Service → Frontend: Enriched deal response with names
```

### Asynchronous Communication (NATS) - Planned

Event-driven messaging for decoupled operations and cross-service notifications.

**Event Structure:**
```go
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

**Event Flow Example - Deal Closure:**
```
1. Frontend → Deal Service: PUT /api/deals/123 {stage: "closed_won"}
2. Deal Service → Database: Update deal record
3. Deal Service → NATS: Publish deal.closed_won event
4. Communication Service ← NATS: Receive deal.closed_won
5. Communication Service: Log "Deal Closed Won" activity
6. Communication Service: Trigger follow-up email sequence
```

---

## Shared Packages

### Database Package (`pkg/database/`)

All services use the shared database package for standardized connectivity:

```go
// Service main.go pattern
package main

import (
    "context"
    "log"

    "crm-platform/pkg/database"
    "crm-platform/[service]/internal/db"
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

**Connection Pool Settings:**
- MaxConns: 20
- MinConns: 5
- MaxConnLifetime: 60 minutes
- MaxConnIdleTime: 5 minutes
- ConnectTimeout: 30 seconds
- QueryTimeout: 30 seconds

### Tenant Context Package (`pkg/tenant/`)

Standardized tenant context management across all services:

```go
// Middleware for tenant context injection
func TenantMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        tenantID := c.GetHeader("X-Tenant-ID")

        // Validate tenant exists
        // Set search_path for PostgreSQL

        c.Set("tenant_id", tenantID)
        c.Next()
    }
}

// Handler usage
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
```

### Middleware Package (`pkg/middleware/`)

Common middleware for all services:

**Authentication Middleware:**
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

        // Set tenant and user context
        c.Set("tenant_id", claims.TenantID)
        c.Set("user_id", claims.UserID)
        c.Next()
    }
}
```

---

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

# Service-specific URLs
AUTH_SERVICE_URL=http://auth-service:8080
TENANT_SERVICE_URL=http://tenant-service:8081
CONTACT_SERVICE_URL=http://contact-service:8082
DEAL_SERVICE_URL=http://deal-service:8083
COMMUNICATION_SERVICE_URL=http://communication-service:8084

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

---

## Error Handling

### Standardized Error Responses

All services use consistent error response format:

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

---

## Health Checks

Each service implements standardized health check endpoints:

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
```

**Example Response:**
```json
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

---

## Performance Considerations

### Connection Pooling

Connection pooling is handled by the shared `pkg/database` package with optimized settings for multi-tenant workloads.

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
- **Tenant schema isolation** ensures data separation

---

## Testing Strategy

### Unit Tests

Each service includes comprehensive unit tests:

```bash
# Run service tests
cd services/[service-name] && go test -v ./...
```

### Integration Tests

Test inter-service communication and dependencies:

```go
func TestDealEnrichment(t *testing.T) {
    // Setup test tenant
    tenant := createTestTenant()

    // Create test contact
    contact := createTestContact(tenant.ID)

    // Create deal with contact reference
    deal := createTestDeal(tenant.ID, contact.ID)

    // Verify enrichment via API client
    enrichedDeal := getDeal(deal.ID)
    assert.Equal(t, contact.FirstName, enrichedDeal.ContactName)
}
```

---

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
  name: deal-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: deal-service
  template:
    metadata:
      labels:
        app: deal-service
    spec:
      containers:
      - name: deal-service
        image: deal-service:latest
        ports:
        - containerPort: 8083
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: database-secret
              key: url
        - name: CONTACT_SERVICE_URL
          value: "http://contact-service:8082"
        - name: AUTH_SERVICE_URL
          value: "http://auth-service:8080"
```

---

## Related Documentation

- [Database Design](./database.md) - Multi-tenant data architecture
- [SQLC Implementation](./sqlc.md) - Database access patterns
- [Shared Packages](./shared-packages.md) - Platform-wide utilities

---

## Next Steps

1. Implement **tenant-service** (no dependencies - start here)
2. Complete **auth-service** (depends only on tenant-service)
3. Implement **contact-service** (depends on auth-service)
4. Verify **deal-service** integration (update to use real API clients)
5. Implement **communication-service** (depends on all others)

Each service implementation should:
- Follow the DTO contracts defined here
- Expose `pkg/client/` for other services
- Store only owned data internally
- Use API clients to fetch external data
