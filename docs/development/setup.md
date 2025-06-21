# Development Setup

Complete guide for setting up the MTenant CRM development environment.

## Prerequisites

- **Go 1.24.3+** - Backend services
- **Docker & Docker Compose** - Infrastructure services
- **Git** - Version control
- **Make** - Build automation

### Optional Tools
- **psql** - PostgreSQL command line client
- **redis-cli** - Redis command line client
- **NATS CLI** - Message queue testing

## Quick Start

1. **Clone and setup workspace:**
   ```bash
   git clone <repository-url>
   cd MTenant
   ```

2. **Start infrastructure:**
   ```bash
   make dev-up
   ```

3. **Verify services:**
   ```bash
   make dev-status
   ```

4. **Build all services:**
   ```bash
   make build
   ```

5. **Run tests:**
   ```bash
   make test
   ```

## Infrastructure Services

### PostgreSQL Database
- **Host:** localhost:5433
- **Database:** crm-platform
- **Username:** admin
- **Password:** admin
- **Connection String:** `postgresql://admin:admin@localhost:5433/crm-platform`

### NATS Message Queue
- **Client Port:** localhost:4222
- **Monitoring:** localhost:8222
- **Health Check:** `curl http://localhost:8222/healthz`

### Redis Cache
- **Host:** localhost:6379
- **No authentication required**
- **Memory Policy:** allkeys-lru (256MB limit)

## Go Workspace

The project uses Go workspaces to manage multiple modules including shared packages:

```go
// go.work
go 1.24.3

use (
    ./pkg                       // Shared packages
    ./services/auth-service
    ./services/communication-service
    ./services/contact-service
    ./services/deal-service
    ./services/tenant-service
)
```

### Module Dependencies

**Shared Package (`pkg/`):**
- `github.com/jackc/pgx/v5` - PostgreSQL driver and connection pooling
- Database configuration, health monitoring, and metrics collection

**Service Dependencies:**
- `crm-platform/pkg/database` - Shared database package (via workspace)
- Generated SQLC code for type-safe queries
- Service-specific business logic and handlers

## Directory Structure

```
MTenant/
├── services/                 # Microservices
│   ├── auth-service/
│   ├── tenant-service/
│   ├── contact-service/
│   ├── deal-service/
│   └── communication-service/
├── pkg/                      # Shared packages
│   ├── database/            # Database connection and pooling
│   │   ├── config.go        # Environment configuration
│   │   ├── pool.go          # Connection pool management
│   │   ├── health.go        # Health checks and monitoring
│   │   ├── metrics.go       # Database metrics collection
│   │   └── ex_test.go       # Integration tests
│   ├── middleware/          # HTTP middleware (planned)
│   └── utils/               # Common utilities (planned)
├── migrations/               # Database migrations
├── k8s/                     # Kubernetes manifests
├── docs/                    # Documentation
├── Makefile                 # Build automation
├── compose.yaml             # Development infrastructure
└── go.work                  # Go workspace configuration
```

## Development Workflow

### Starting Development

1. **Start infrastructure:**
   ```bash
   make dev-up
   ```

2. **Check service health:**
   ```bash
   make dev-status
   ```

3. **View logs:**
   ```bash
   make dev-logs              # All services
   make dev-logs-db           # PostgreSQL only
   make dev-logs-nats         # NATS only
   make dev-logs-redis        # Redis only
   ```

### Building Services

```bash
# Build all services
make build

# Build specific service
make build-auth-service
make build-tenant-service
make build-contact-service
make build-deal-service
make build-communication-service
```

### Running Tests

```bash
# Run all tests
make test

# Run tests for specific service
cd services/auth-service && go test -v ./...

# Run database integration tests (requires DATABASE_URL)
export DATABASE_URL="postgresql://admin:admin@localhost:5433/crm-platform?sslmode=disable"
cd pkg && go test -v ./database
```

### Database Operations

```bash
# Reset database (destroys all data)
make reset-db

# Stop infrastructure
make dev-down

# Restart infrastructure
make dev-restart
```

### SQLC Code Generation

Each service uses SQLC for type-safe database access:

```bash
# Generate code for all services (after schema changes)
cd services/auth-service && sqlc generate
cd services/tenant-service && sqlc generate
cd services/contact-service && sqlc generate
cd services/deal-service && sqlc generate
```

## Environment Testing

Use the comprehensive testing guide in [testing.md](./testing.md) to verify your environment.

## Common Issues

### Port Conflicts
If ports 5433, 4222, 6379, or 8222 are in use:
```bash
# Check what's using ports
lsof -i :5433
lsof -i :4222

# Stop conflicting services or modify compose.yaml
```

### Database Connection Issues
```bash
# Test database connectivity
docker exec crm-platform psql -U admin -d crm-platform -c "SELECT 1;"

# Check container health
docker-compose ps
```

### Go Module Issues
```bash
# Clean module cache
go clean -modcache

# Sync workspace
go work sync
```

### Build Issues
```bash
# Clean build artifacts
make clean

# Rebuild everything
make build
```

## Next Steps

1. Review [Makefile Reference](./makefile.md) for all available commands
2. Understand [Service Architecture](../architecture/services.md)
3. Learn [Database Design](../architecture/database.md)
4. Follow [Testing Guide](./testing.md) to validate your setup