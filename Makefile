SERVICES := $(shell find services -maxdepth 1 -mindepth 1 -type d -exec basename {} \;)
PWD := $(shell pwd)

BUILD_TARGETS := $(addprefix build-, $(SERVICES))

$(BUILD_TARGETS): build-%:
	@echo "Building $*..."
	@mkdir -p services/$*/bin
	@cd services/$* && go build -o bin/$* ./cmd/server
	@echo "✓ $* built successfully"

build: $(BUILD_TARGETS)

.PHONY: build $(BUILD_TARGETS)

test:
	for service in $(SERVICES); do \
		cd $(PWD); \
		echo "Testing $$service..."; \
		cd services/$$service && go test -v -race -coverprofile=coverage.out ./...; \
		cd ../..; \
	done

.PHONY: test

DOCKER_TARGETS := $(addprefix docker-build-, $(SERVICES))

$(DOCKER_TARGETS): docker-build-%:
	@cd services/$* && docker build -t $* .

docker-build: $(DOCKER_TARGETS)

.PHONY: docker-build $(DOCKER_TARGETS)

dev-up:
	@echo "Starting development infrastructure..."
	docker-compose up -d
	@echo "✓ Infrastructure started"
	@echo "Services available:"
	@echo "  - PostgreSQL: localhost:5433 (user: admin, password: admin, db: crm-platform)"
	@echo "  - NATS: localhost:4222 (monitoring: localhost:8222)"
	@echo "  - Redis: localhost:6379"

dev-down:
	@echo "Stopping development infrastructure..."
	docker-compose down
	@echo "✓ Infrastructure stopped and removed"

dev-logs:
	@echo "Showing logs from all containers..."
	docker-compose logs -f

dev-logs-db:
	@echo "Showing PostgreSQL logs..."
	docker-compose logs -f db

dev-logs-nats:
	@echo "Showing NATS logs..."
	docker-compose logs -f nats-server

dev-logs-redis:
	@echo "Showing Redis logs..."
	docker-compose logs -f redis-cache

clean:
	@echo "Cleaning build artifacts and temporary files..."
	@for service in $(SERVICES); do \
		echo "Cleaning $$service..."; \
		rm -rf services/$$service/bin; \
		rm -f services/$$service/coverage.out; \
		rm -f services/$$service/*.log; \
	done
	@rm -f *.log
	@rm -f coverage.out
	@rm -rf tmp/
	@echo "✓ Cleanup complete"

reset-db:
	@echo "Resetting database with fresh schema..."
	@echo "Stopping database container..."
	docker-compose stop db
	@echo "Removing database container and volume..."
	docker-compose rm -f db
	docker volume rm $$(docker volume ls -q | grep crm) 2>/dev/null || true
	@echo "Starting fresh database..."
	docker-compose up -d db
	@echo "Waiting for database to be ready..."
	@echo "Checking database health..."
	@timeout=60; while [ $$timeout -gt 0 ]; do \
		if docker-compose exec -T db pg_isready -U admin -d crm-platform >/dev/null 2>&1; then \
			echo "✓ Database is ready"; \
			break; \
		fi; \
		echo "Waiting for database... ($$timeout seconds remaining)"; \
		sleep 2; \
		timeout=$$((timeout-2)); \
	done
	@echo "✓ Database reset complete"
	@echo "Database available at: postgresql://admin:admin@localhost:5433/crm-platform"

dev-status:
	@echo "Infrastructure status:"
	@docker-compose ps

dev-restart:
	@echo "Restarting development infrastructure..."
	docker-compose restart
	@echo "✓ Infrastructure restarted"

.PHONY: dev-up dev-down dev-logs dev-logs-db dev-logs-nats dev-logs-redis clean reset-db dev-status dev-restart