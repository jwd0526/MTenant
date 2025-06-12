# Multi-Tenant SaaS CRM Platform: Technical Architecture Overview

## Project Overview

This is a multi-tenant CRM platform built with Go microservices and a Vue.js frontend. The system allows multiple organizations to use the same application while maintaining complete data isolation between tenants. Each tenant gets their own database schema, ensuring security and compliance while sharing the underlying infrastructure.

## Architecture Pattern

The system uses a microservices architecture with five core services:

- **Auth Service**: User authentication and JWT token management
- **Tenant Service**: Organization setup and user management
- **Contact Service**: Customer and company data management
- **Deal Service**: Sales pipeline and opportunity tracking
- **Communication Service**: Activity logging and email functionality

Supporting infrastructure includes PostgreSQL for data persistence, NATS for inter-service messaging, Redis for caching, and Kubernetes for orchestration.

## Technology Stack

**Backend Services**:
- Go 1.21+ with Gin HTTP framework
- SQLC for type-safe database queries
- PostgreSQL 15+ with schema-based tenant isolation
- NATS for asynchronous messaging
- Redis for session storage and caching

**Frontend**:
- Vue 3 with TypeScript
- Pinia for state management
- Tailwind CSS for styling
- Axios for HTTP client

**Infrastructure**:
- Kubernetes for container orchestration
- Docker for containerization
- Prometheus/Grafana for monitoring
- GitHub Actions for CI/CD

## Multi-Tenant Data Isolation

### Schema-Based Isolation

The platform uses PostgreSQL schemas to isolate tenant data. When TenantCo registers, the system creates a dedicated schema `tenantco_123` containing all necessary tables. This approach provides:

- Complete data separation between tenants
- Efficient queries within tenant boundaries
- Simplified backup and recovery per tenant
- Clear compliance boundaries for data protection

### Tenant Context Flow

1. User authenticates through Auth Service
2. JWT token includes tenant ID in claims
3. Middleware extracts tenant context from token
4. Database queries execute within correct tenant schema
5. All operations remain isolated to tenant's data

Example: When Sarah from TenantCo creates a contact, the query becomes:
```sql
INSERT INTO tenantco_123.contacts (name, email, company_id) VALUES (...)
```

## Service Breakdown

### Authentication Service

Handles user authentication, token generation, and validation.

**Core Functions**:
- User registration with password hashing (bcrypt)
- JWT token creation with tenant context
- Token validation middleware for other services
- Password reset workflows

**Database Schema**:
- User accounts stored in tenant-specific schemas
- Password hashes and security metadata
- Session management for token refresh

**API Endpoints**:
- `POST /api/auth/register` - User registration
- `POST /api/auth/login` - Authentication
- `POST /api/auth/refresh` - Token refresh
- `GET /api/auth/profile` - User profile data

### Tenant Management Service

Manages organization setup, user invitations, and tenant-level configuration.

**Core Functions**:
- Organization creation with subdomain validation
- Dynamic schema creation for new tenants
- User invitation and role management
- Tenant settings and configuration

**Schema Creation Process**:
1. Validate organization details and subdomain availability
2. Create tenant record in global registry
3. Execute `CREATE SCHEMA tenant_xxx`
4. Copy table structure from template schema
5. Create initial admin user in tenant schema

**User Management**:
- Role-based access control (Admin, Manager, Sales Rep, Viewer)
- Invitation workflow with email tokens
- User permission management within tenant context

### Contact Management Service

Handles customer and company data with full CRUD operations.

**Data Model**:
- Contacts with standard fields plus JSONB custom fields
- Companies with hierarchy support (parent-child relationships)
- Contact-company associations
- Full-text search across contact data

**Key Features**:
- Custom field definitions per tenant
- Advanced search and filtering
- CSV import/export functionality
- Soft delete with data preservation
- Activity associations for interaction history

**Performance Optimizations**:
- Database indexes on commonly queried fields
- Pagination for large contact lists
- Efficient full-text search using PostgreSQL capabilities

### Deal Management Service

Tracks sales opportunities through configurable pipeline stages.

**Pipeline Configuration**:
- Default stages: Lead → Qualified → Proposal → Negotiation → Closed Won/Lost
- Tenant-customizable stage definitions
- Probability assignments per stage
- Stage transition validation

**Analytics Capabilities**:
- Revenue forecasting based on deal probability
- Conversion rate analysis by stage
- Sales velocity tracking
- Performance metrics by sales rep
- Pipeline bottleneck identification

**Event Publishing**:
Deal progression triggers events that other services consume:
- Deal created → Communication service logs activity
- Stage changed → Update forecasting calculations
- Deal closed → Trigger automated follow-up workflows

### Communication Service

Manages all customer interaction tracking and email functionality.

**Activity Types**:
- Emails (sent, received, tracking data)
- Phone calls with duration and notes
- Meetings with attendees and outcomes
- General notes and observations
- Tasks with due dates and assignments

**Email System**:
- SMTP integration for sending emails
- Template system with variable substitution
- Open and click tracking via pixels and redirect URLs
- Thread management for conversation history
- Attachment handling and storage

**Tracking Implementation**:
- Invisible 1x1 pixel images for open tracking
- URL rewriting for click tracking
- Bounce and delivery status monitoring
- Privacy-compliant tracking with opt-out options

## Inter-Service Communication

### Synchronous Communication (HTTP)

Direct API calls for immediate data needs:
- Frontend to any service for user interactions
- Service-to-service calls for related data
- Real-time validation and data retrieval

### Asynchronous Communication (NATS)

Event-driven messaging for decoupled operations:

**Event Examples**:
- `user.created` → Trigger welcome email, setup default data
- `deal.stage_changed` → Update analytics, log activity
- `email.opened` → Update engagement metrics
- `tenant.created` → Initialize default configurations

**Message Structure**:
```json
{
  "event_type": "deal.closed_won",
  "tenant_id": "tenantco_123",
  "timestamp": "2025-06-11T10:30:00Z",
  "data": {
    "deal_id": "deal_456",
    "value": 50000,
    "contact_id": "contact_789"
  }
}
```

## Frontend Architecture

### Component Structure

**Layout Components**:
- Main navigation with route-based highlighting
- Header with user controls and organization context
- Dashboard with metrics and activity feeds

**Feature Components**:
- Contact list with search, filtering, and pagination
- Deal pipeline with drag-and-drop stage management
- Activity timeline with chronological interaction history
- Email composer with template selection

### State Management (Pinia)

**Store Organization**:
- `authStore` - User authentication and tenant context
- `contactStore` - Contact data and search state
- `dealStore` - Pipeline data and filtering
- `activityStore` - Communication history and tasks

**Data Flow**:
1. User action triggers store method
2. Store makes API call to appropriate service
3. Response updates store state
4. Components reactively update based on state changes

### API Integration

**HTTP Client Configuration**:
- Axios interceptors for automatic JWT token inclusion
- Tenant context headers on all requests
- Error handling and retry logic
- Response caching for frequently accessed data

## Database Design

### Global Tables

Located in the default `public` schema:
- `tenants` - Organization registry with subdomain mappings
- `migrations` - Schema version tracking across tenants

### Tenant-Specific Schema

Each tenant gets a complete schema with:
- `users` - Tenant user accounts and roles
- `contacts` - Customer and prospect information
- `companies` - Business entities with hierarchy support
- `deals` - Sales opportunities and pipeline data
- `activities` - All customer interaction history
- `custom_fields` - Tenant-specific field definitions

### Query Patterns

**Tenant-Aware Queries**:
All service queries include tenant context through schema selection:
```sql
-- Contact lookup with tenant isolation
SELECT * FROM tenantco_123.contacts WHERE email = $1;

-- Deal pipeline aggregation
SELECT stage, COUNT(*), SUM(value) 
FROM tenantco_123.deals 
WHERE status = 'open' 
GROUP BY stage;
```

## Security Implementation

### Authentication Flow

1. User submits credentials to Auth Service
2. Service validates against tenant-specific user table
3. bcrypt comparison for password verification
4. JWT generation with user and tenant claims
5. Token returned with refresh token for session management

### Authorization Middleware

**Token Validation**:
- JWT signature verification using shared secret
- Expiration time validation
- Tenant context extraction and validation
- User role and permission verification

**Request Context**:
Every authenticated request carries:
- User ID and email
- Tenant ID for schema selection
- Role and permission claims
- Request correlation ID for tracing

### Data Protection

**Encryption**:
- TLS 1.3 for all HTTP communications
- Database encryption at rest
- JWT signing with HMAC-SHA256 or RSA

**Access Control**:
- Role-based permissions (Admin, Manager, Sales Rep, Viewer)
- Resource-level access validation
- Audit logging for all data modifications

## Deployment Architecture

### Kubernetes Configuration

**Service Deployment**:
Each microservice runs as a separate Kubernetes deployment with:
- Multiple replicas for high availability
- Resource limits and requests
- Health checks and readiness probes
- Rolling update strategy for zero-downtime deployments

**Infrastructure Services**:
- PostgreSQL StatefulSet with persistent volumes
- NATS deployment with clustering support
- Redis deployment for caching and sessions
- Ingress controller for external traffic routing

### CI/CD Pipeline

**GitHub Actions Workflow**:
1. Code commit triggers automated testing
2. Docker image building with multi-stage optimization
3. Security scanning for dependencies and containers
4. Kubernetes deployment with environment-specific configurations
5. Health check validation and rollback capability

**Environment Progression**:
- Development: Local Docker Compose setup
- Staging: Kubernetes cluster with production-like configuration
- Production: High-availability cluster with monitoring and alerting

## Monitoring and Observability

### Metrics Collection (Prometheus)

**System Metrics**:
- Resource utilization (CPU, memory, disk)
- Network traffic and latency
- Database connection pool status

**Application Metrics**:
- HTTP request duration and status codes
- Database query performance
- Business metrics (users, deals, revenue)

**Custom Metrics**:
- Tenant-specific usage patterns
- Feature adoption rates
- Performance by tenant size

### Logging Strategy

**Structured Logging**:
- JSON format for machine parsing
- Correlation IDs for request tracing
- Tenant context in all log entries
- Centralized log aggregation

**Log Levels**:
- DEBUG: Detailed execution flow
- INFO: Normal operations and state changes
- WARN: Recoverable errors and degraded performance
- ERROR: Failures requiring attention

### Distributed Tracing (Jaeger)

Request flow tracking across services:
- Frontend request → Auth Service → Target Service → Database
- Performance bottleneck identification
- Error propagation analysis
- Service dependency mapping

## Performance Considerations

### Database Optimization

**Indexing Strategy**:
- Primary keys on all ID columns
- Composite indexes for common query patterns
- Full-text search indexes for contact data
- Tenant ID indexes for isolation queries

**Connection Management**:
- Connection pooling with configurable limits
- Connection health monitoring
- Automatic retry with exponential backoff

### Caching Strategy

**Redis Implementation**:
- User session data for fast authentication
- Frequently accessed tenant configurations
- Contact and company lookup caching
- API response caching with TTL

### Frontend Performance

**Code Splitting**:
- Route-based code splitting for faster initial loads
- Component-level splitting for large features
- Dynamic imports for optional functionality

**Data Loading**:
- Pagination for large datasets
- Infinite scrolling for activity feeds
- Optimistic updates for better perceived performance

## Example: TenantCo User Journey

### Initial Setup

TenantCo registers through the frontend:
1. Frontend calls `POST /api/tenants/register`
2. Tenant Service validates subdomain "tenantco"
3. Creates schema `tenantco_123` with full table structure
4. Creates admin user Sarah in tenant schema
5. Returns authentication credentials

### Daily Operations

Sarah manages her sales pipeline:

**Adding a Contact**:
1. Frontend calls `POST /api/contacts` with contact data
2. Contact Service validates and creates record in `tenantco_123.contacts`
3. Returns contact ID and updates frontend state

**Creating a Deal**:
1. Frontend calls `POST /api/deals` with opportunity data
2. Deal Service creates record and publishes `deal.created` event
3. Communication Service receives event and logs "Deal Created" activity
4. Frontend updates pipeline view

**Sending Email**:
1. Frontend calls `POST /api/communications/emails/send`
2. Communication Service processes template and sends via SMTP
3. Creates activity record with tracking information
4. Returns tracking ID for status monitoring

### Data Flow Example

When Sarah moves a deal to "Closed Won":
1. Frontend calls `PUT /api/deals/123` with new stage
2. Deal Service updates record in `tenantco_123.deals`
3. Publishes `deal.closed_won` event to NATS
4. Communication Service creates "Deal Closed Won" activity
5. Analytics system updates revenue metrics
6. Frontend receives success response and updates UI

## Scalability Architecture

### Horizontal Scaling

**Service Level**:
- Kubernetes automatically scales pods based on CPU/memory usage
- Load balancing across service instances
- Database connection pooling prevents bottlenecks

**Database Level**:
- Read replicas for analytics and reporting queries
- Schema-based tenant distribution across database servers
- Automated backup and point-in-time recovery

### Performance Targets

The architecture supports:
- 1000+ concurrent users across all tenants
- Sub-200ms API response times for typical operations
- 10,000+ contacts per tenant without degradation
- 99.9% uptime with proper infrastructure configuration

### Growth Patterns

**Tenant Growth**:
- New tenant creation takes ~2 seconds including schema setup
- No performance impact on existing tenants
- Tenant-specific optimizations possible through schema tuning

**Data Growth**:
- Automatic archival of old activities and communications
- Configurable retention policies per tenant
- Efficient indexing maintains query performance

This architecture provides a solid foundation for a multi-tenant SaaS CRM that can scale from hundreds to hundreds of thousands of users while maintaining security, performance, and compliance requirements.