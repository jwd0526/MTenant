# Makefile Reference

**Last Updated:** 2025-10-08\
*Build system and development commands*

Comprehensive guide to the build system and development commands.

## Overview

The Makefile provides a standardized interface for building, testing, and managing the development environment across all microservices.

## Build Commands

### Service Building

```bash
# Build all services
make build

# Build specific services
make build-auth-service
make build-tenant-service  
make build-contact-service
make build-deal-service
make build-communication-service
```

**What it does:**
- Creates `services/{service}/bin/` directory
- Compiles `./cmd/server` to `bin/{service}` executable
- Uses Go's native build with no external dependencies

**Output:**
```
Building auth-service...
✓ auth-service built successfully
```

### Docker Building

```bash
# Build Docker images for all services
make docker-build

# Build specific service image
make docker-build-auth-service
```

**What it does:**
- Builds Docker images using each service's Dockerfile
- Tags images with service name
- Uses multi-stage builds for optimization

## Testing Commands

### Comprehensive Testing

```bash
# Run tests for all services
make test
```

**What it does:**
- Iterates through all services in the `services/` directory
- Runs `go test -v -race -coverprofile=coverage.out ./...` for each
- Enables race detection and coverage reporting
- Fails fast on first test failure

**Output per service:**
```
Testing auth-service...
=== RUN   TestMain
=== RUN   TestBenchmark
--- PASS: TestBenchmark (0.00s)
PASS
coverage: 0.0% of statements
ok      crm-platform/auth-service       0.123s  coverage: 0.0% of statements
```

## Development Environment

### Infrastructure Management

```bash
# Start all development infrastructure
make dev-up

# Stop and remove all containers
make dev-down

# Restart all services
make dev-restart

# Check status of all containers
make dev-status
```

#### `make dev-up` Details

**Services Started:**
- PostgreSQL on port 5433
- NATS on ports 4222 (client) and 8222 (monitoring) 
- Redis on port 6379

**Features:**
- Health checks for database readiness
- Persistent volumes for data retention
- Custom network for service communication

**Output:**
```
Starting development infrastructure...
✓ Infrastructure started
Services available:
  - PostgreSQL: localhost:5433 (user: admin, password: admin, db: crm-platform)
  - NATS: localhost:4222 (monitoring: localhost:8222)  
  - Redis: localhost:6379
```

### Log Management

```bash
# View logs from all containers
make dev-logs

# View specific service logs
make dev-logs-db      # PostgreSQL logs
make dev-logs-nats    # NATS server logs  
make dev-logs-redis   # Redis logs
```

**Features:**
- Real-time log streaming with `-f` flag
- Container-specific filtering
- Timestamps and log levels preserved

### Database Operations

```bash
# Reset database with fresh schema
make reset-db
```

**What it does:**
1. Stops PostgreSQL container
2. Removes container and associated volume
3. Recreates fresh database container
4. Waits for database readiness (60-second timeout)
5. Validates connection health

**Warning:** This destroys all data. Use only for development reset.

**Output:**
```
Resetting database with fresh schema...
Stopping database container...
Removing database container and volume...
Starting fresh database...
Waiting for database to be ready...
✓ Database is ready
✓ Database reset complete
Database available at: postgresql://admin:admin@localhost:5433/crm-platform
```

## Maintenance Commands

### Cleanup

```bash
# Clean all build artifacts and temporary files
make clean
```

**What it removes:**
- `services/*/bin/` directories
- `*.log` files (root and service-level)
- `coverage.out` files  
- `tmp/` directory
- Any generated temporary files

**Output:**
```
Cleaning build artifacts and temporary files...
Cleaning auth-service...
Cleaning tenant-service...
Cleaning contact-service...
Cleaning deal-service...
Cleaning communication-service...
✓ Cleanup complete
```

## Internal Mechanics

### Service Discovery

The Makefile automatically discovers services:

```makefile
SERVICES := $(shell find services -maxdepth 1 -mindepth 1 -type d -exec basename {} \;)
```

This finds all directories in `services/` and uses them to generate build targets.

### Dynamic Target Generation

Build targets are generated dynamically:

```makefile
BUILD_TARGETS := $(addprefix build-, $(SERVICES))

$(BUILD_TARGETS): build-%:
    @echo "Building $*..."
    @mkdir -p services/$*/bin
    @cd services/$* && go build -o bin/$* ./cmd/server
    @echo "✓ $* built successfully"
```

The `%` pattern allows `make build-auth-service` to call the rule with `$* = auth-service`.

### Docker Integration

Docker commands use the same service discovery:

```makefile
DOCKER_TARGETS := $(addprefix docker-build-, $(SERVICES))

$(DOCKER_TARGETS): docker-build-%:
    @cd services/$* && docker build -t $* .
```

### Database Health Checking

The reset-db command includes sophisticated health checking:

```makefile
@timeout=60; while [ $$timeout -gt 0 ]; do \
    if docker-compose exec -T db pg_isready -U admin -d crm-platform >/dev/null 2>&1; then \
        echo "✓ Database is ready"; \
        break; \
    fi; \
    echo "Waiting for database... ($$timeout seconds remaining)"; \
    sleep 2; \
    timeout=$$((timeout-2)); \
done
```

## Adding New Services

To add a new service:

1. Create directory: `services/new-service/`
2. Add `cmd/server/main.go` entry point
3. Create `Dockerfile` for containerization
4. The Makefile will automatically discover and include it

No Makefile modifications required!

## Best Practices

### Development Workflow

1. **Start clean:**
   ```bash
   make clean
   make dev-up
   ```

2. **Develop and test:**
   ```bash
   make build
   make test
   ```

3. **Reset when needed:**
   ```bash
   make reset-db  # Only when database schema changes
   ```

### CI/CD Integration

The Makefile targets are designed for CI/CD:

```bash
# Typical CI pipeline
make clean
make build
make test
make docker-build
```

### Debugging

```bash
# Check what's running
make dev-status

# Follow logs for debugging
make dev-logs

# Test specific service
cd services/auth-service && go test -v ./...
```

## Performance Notes

- **Parallel builds:** Go's build system handles concurrency automatically
- **Incremental builds:** Only changed files are recompiled
- **Docker layer caching:** Multi-stage builds optimize image size
- **Test caching:** Go's test cache speeds up repeated test runs

## Troubleshooting

### Build Failures

```bash
# Clean and rebuild
make clean
make build

# Check Go workspace
go work sync
```

### Infrastructure Issues

```bash
# Check container status
make dev-status

# Restart everything
make dev-down
make dev-up
```

### Port Conflicts

Edit `compose.yaml` to use different ports if 5433, 4222, 6379, or 8222 are occupied.

## Related Documentation

- [Development Setup](./setup.md) - Environment setup guide
- [Testing Guide](./testing.md) - Comprehensive testing procedures
- [Service Architecture](../architecture/services.md) - Understanding the services built by these commands