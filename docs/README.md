# MTenant CRM Documentation

This directory contains comprehensive documentation for the multi-tenant CRM platform.

## Documentation Structure

### Setup & Development
- [Development Setup](./development/setup.md) - Getting started with local development
- [Makefile Reference](./development/makefile.md) - Build system and development commands
- [Testing Guide](./development/testing.md) - Environment testing and validation

### Architecture
- [Service Architecture](./architecture/services.md) - Microservices design and communication
- [Database Design](./architecture/database.md) - Multi-tenant data isolation and schemas
- [SQLC Implementation](./architecture/sqlc.md) - Type-safe database access

### Database Schemas
- [Global Schemas](./database/global/) - Tenant registry and system tables
- [Tenant Templates](./database/tenant-template/) - Schema templates for tenant isolation
- [Database Migrations](./database/migrations.md) - Migration strategy and golang-migrate guide
- [Query Documentation](./database/queries/) - SQL query documentation by service

### Services
- [Auth Service](./services/auth-service.md) - Authentication and user management
- [Tenant Service](./services/tenant-service.md) - Organization and invitation management
- [Contact Service](./services/contact-service.md) - Contact and company management
- [Deal Service](./services/deal-service.md) - Sales pipeline management
- [Communication Service](./services/communication-service.md) - Activity tracking

### Operations
- [Deployment](./operations/deployment.md) - Kubernetes and container deployment
- [Monitoring](./operations/monitoring.md) - Observability and alerting
- [Security](./operations/security.md) - Security considerations and compliance

## Quick Start

1. [Development Setup](./development/setup.md) - Set up your local environment
2. [Makefile Reference](./development/makefile.md) - Learn the build commands
3. [Service Architecture](./architecture/services.md) - Understand the system design
4. [Database Design](./architecture/database.md) - Learn the multi-tenant data model

## Implementation Status

✅ **Completed through Ticket 1.2.11:**
- Project structure and Go modules
- Docker infrastructure setup
- Database schemas for all services
- SQLC configurations and code generation
- Service skeletons and basic structure

🔄 **Next Steps:**
- Install golang-migrate tool (Ticket 1.2.11)
- Create initial database migrations  
- Implement service handlers and business logic