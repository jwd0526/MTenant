# CRM Platform - Deal Service

**Last Updated:** 2025-10-08\
*Production-ready implementation*

Deal managing service for the MTenant CRM.

## 🏗️ Architecture

This service implements a **schema-per-tenant** multi-tenancy pattern, providing complete data isolation while maintaining high performance and scalability.

### Key Features

- **Multi-Tenant Architecture**: Complete tenant isolation using PostgreSQL schemas
- **Type-Safe Database Operations**: SQLC-generated queries with compile-time safety
- **JWT Authentication**: Secure API access with role-based permissions
- **Comprehensive Testing**: Unit, integration, and E2E test coverage
- **Production Ready**: Health checks, metrics, graceful shutdown, Docker support

### Technology Stack

- **Go 1.24+**: Modern Go with generics and latest features
- **PostgreSQL**: Advanced database with JSONB, full-text search
- **pgx v5**: High-performance PostgreSQL driver with connection pooling
- **SQLC**: Type-safe SQL query generation
- **Gin**: High-performance HTTP web framework
- **Testify**: Comprehensive testing framework

## 🚀 Quick Start

### Prerequisites

- Go 1.24 or higher
- PostgreSQL 13 or higher
- Docker (optional)

### Local Development Setup

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd deal-service
   ```

2. **Set up PostgreSQL**
   ```bash
   # Using Docker
   docker run --name crm-postgres -e POSTGRES_PASSWORD=<credential> -e POSTGRES_USER=<credential> -e POSTGRES_DB=crm-platform -p 5433:5432 -d postgres:15
   
   # Or install PostgreSQL locally and create the database
   createdb crm-platform
   ```

3. **Configure environment**
   ```bash
   cp .env.local .env
   # Edit .env with your database credentials
   ```

4. **Install dependencies**
   ```bash
   go mod download
   ```

5. **Run database migrations**
   ```bash
   # Create tenant template schema (run SQL files in db/schema/)
   psql -d crm-platform -f db/schema/users.sql
   psql -d crm-platform -f db/schema/companies.sql
   psql -d crm-platform -f db/schema/contacts.sql
   psql -d crm-platform -f db/schema/deals.sql
   psql -d crm-platform -f db/schema/deal_contacts.sql
   ```

6. **Run the service**
   ```bash
   go run cmd/server/main.go
   ```

The service will start on port 8080 (configurable via `PORT` environment variable).

## 📁 Project Structure

```
deal-service/
├── cmd/server/           # Application entry point
├── database/            # Database connection and pooling
├── db/                  # SQL schemas and queries
│   ├── queries/         # SQLC query definitions
│   └── schema/          # PostgreSQL table definitions
├── internal/            # Private application code
│   ├── config/          # Configuration management
│   ├── db/              # Generated SQLC code
│   ├── errors/          # Error definitions
│   ├── handlers/        # HTTP request handlers
│   ├── middleware/      # HTTP middleware
│   └── models/          # Request/response models
├── tenant/              # Multi-tenancy implementation
├── tests/               # Test suites
│   ├── e2e/            # End-to-end tests
│   ├── integration/    # Integration tests
│   ├── unit/           # Unit tests
│   ├── fixtures/       # Test data fixtures
│   └── helpers/        # Test utilities
└── Dockerfile           # Container definition
```

## 🔧 Configuration

The service is configured via environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | Required |
| `PORT` | HTTP server port | `8080` |
| `ENVIRONMENT` | Environment (dev/prod) | `dev` |
| `SHARED_JWT_SECRET` | JWT signing secret | Required |

### Example Configuration

```bash
export DATABASE_URL="postgres://<credential>:<credential>@localhost:5433/crm-platform?sslmode=disable"
export PORT=8080
export ENVIRONMENT=dev
export SHARED_JWT_SECRET="your-secret-key"
```

## 🌐 API Endpoints

### Deals Management

```
POST   /api/v1/deals           # Create a new deal
GET    /api/v1/deals           # List deals with pagination
GET    /api/v1/deals/:id       # Get deal by ID
PUT    /api/v1/deals/:id       # Update deal
DELETE /api/v1/deals/:id       # Delete deal
POST   /api/v1/deals/:id/close # Close a deal
GET    /api/v1/deals/pipeline  # Pipeline analytics
GET    /api/v1/deals/owner/:id # Get deals by owner
```

### Authentication

All endpoints require JWT authentication via `Authorization: Bearer <token>` header.

### Multi-Tenant Access

Tenant isolation is automatic based on JWT claims. Each request is automatically scoped to the authenticated user's tenant.

## 🧪 Testing

The project includes comprehensive test coverage:

```bash
# Run all tests
go test ./... -v

# Run specific test suites
go test ./tests/unit/... -v          # Unit tests
go test ./tests/integration/... -v -p 1   # Integration tests (single-threaded)
go test ./tests/e2e/... -v -p 1          # E2E tests (single-threaded)

# Run with coverage
go test ./... -cover
```

**Note**: Integration and E2E tests must run with `-p 1` (single-threaded) due to shared tenant IDs.

### Test Database Setup

Tests require a separate test database:

```bash
export DATABASE_URL="postgres://<credential>:<credential>@localhost:5433/crm-platform?sslmode=disable"
```

## 🐳 Docker Support

Build and run with Docker:

```bash
# Build image
docker build -t deal-service .

# Run container
docker run -p 8080:8080 \
  -e DATABASE_URL="postgres://<credential>:<credential>@host.docker.internal:5433/crm-platform?sslmode=disable" \
  -e SHARED_JWT_SECRET="your-secret" \
  deal-service
```

## 📊 Monitoring & Health Checks

### Health Check Endpoint

```bash
GET /health
```

Returns database connection status and pool statistics.

### Metrics

The service exposes internal metrics via the database package:
- Connection pool statistics
- Query performance metrics
- Health check status

## 🔒 Security

### Authentication & Authorization
- JWT-based authentication
- Role-based access control
- Tenant isolation at database level

### Multi-Tenant Security
- Complete schema isolation per tenant
- Automatic tenant context extraction
- Protection against cross-tenant data access

## 🚀 Deployment

### Production Considerations

1. **Database**: Use managed PostgreSQL service (AWS RDS, Google Cloud SQL)
2. **Connection Pooling**: Configure appropriate pool sizes based on load
3. **SSL**: Enable SSL for database connections (`sslmode=require`)
4. **Secrets**: Use secure secret management for JWT keys
5. **Monitoring**: Set up logging and monitoring infrastructure

### Environment Variables for Production

```bash
export ENVIRONMENT=production
export DATABASE_URL="postgres://user:pass@prod-db:5432/crm?sslmode=require"
export SHARED_JWT_SECRET="production-secret-key"
export PORT=8080
```

## 📈 Performance

### Database Optimization
- Connection pooling with pgx v5
- Prepared statement caching
- Efficient indexing strategy
- Query performance monitoring

### Scalability
- Stateless service design
- Horizontal scaling ready
- Efficient tenant isolation
- Optimized database queries

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes with tests
4. Run full test suite
5. Submit pull request

### Development Guidelines

- Write tests for new features
- Follow Go best practices
- Update documentation
- Ensure type safety with SQLC
- Maintain tenant isolation

## 📚 Documentation

- [Database Package](./database/README.md) - Connection pooling and health checks
- [Tenant Package](./tenant/README.md) - Multi-tenancy implementation
- [API Documentation](./API.md) - API documentation
