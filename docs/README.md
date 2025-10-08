# MTenant CRM Documentation

**Last Updated:** 2025-10-08\
*Update implementation status*

This directory contains comprehensive documentation for the multi-tenant CRM platform.

## Documentation Structure

### Setup & Development
- [Development Setup](./development/setup.md) - Getting started with local development
- [Makefile Reference](./development/makefile.md) - Build system and development commands
- [Testing Guide](./development/testing.md) - Environment testing and validation

### Architecture
- [Service Architecture](./architecture/services.md) - Microservices design and communication
- [Shared Packages](./architecture/shared-packages.md) - Common functionality and database management
- [Database Design](./architecture/database.md) - Multi-tenant data isolation and schemas
- [SQLC Implementation](./architecture/sqlc.md) - Type-safe database access

### Database Schemas
- [Global Schemas](./database/global/) - Tenant registry and system tables
- [Tenant Templates](./database/tenant-template/) - Schema templates for tenant isolation
- [Database Migrations](./database/migrations/) - Migration strategy and golang-migrate guide

### Services
- [Auth Service](./services/auth-service.md) - Authentication and user management
- [Tenant Service](./services/tenant-service.md) - Organization and invitation management
- [Contact Service](./services/contact-service.md) - Contact and company management
- [Deal Service](./services/deal-service.md) - Sales pipeline management
- [Communication Service](./services/communication-service.md) - Activity tracking

## Quick Start

1. [Development Setup](./development/setup.md) - Set up your local environment
2. [Makefile Reference](./development/makefile.md) - Learn the build commands
3. [Service Architecture](./architecture/services.md) - Understand the system design
4. [Database Design](./architecture/database.md) - Learn the multi-tenant data model

## Implementation Status

**Completed:**
- ✅ Project structure and Go modules
- ✅ Docker infrastructure setup
- ✅ Database connection pool package (`pkg/database`)
- ✅ Tenant-aware database package (`pkg/tenant`)
- ✅ Database migrations (tenant registry and schema template)
- ✅ **Deal Service** - Fully implemented with handlers, business logic, and tests
- ✅ SQLC configurations for all services

**In Progress:**
- 🔨 Auth Service - SQLC setup complete, handlers needed
- 🔨 Tenant Service - SQLC queries and structure exist, implementation needed
- 🔨 Contact Service - SQLC setup complete, handlers needed

**Planned:**
- 📋 Communication Service - Complete implementation needed