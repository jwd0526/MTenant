#!/bin/bash

# USAGE: ./test.sh <postgres-username>

# CRM Platform Health Check Script
set -e  # Exit on any error

echo "=== Container Status ==="
echo

echo "Docker Compose Services:"
docker-compose ps

echo
echo "Running Containers:"
docker ps

echo
echo "Real-time logs (last 20 lines):"
docker-compose logs --tail=20

echo
echo "=== Test PostgreSQL ==="

echo "Testing PostgreSQL connection..."
docker exec crm-platform psql -U $1 -d crm-platform -c "SELECT version();"

echo
echo "Listing databases:"
docker exec crm-platform psql -U $1 -d crm-platform -c "\l"

echo
echo "Testing write permissions..."
docker exec crm-platform psql -U $1 -d crm-platform -c "CREATE TABLE IF NOT EXISTS test_table (id SERIAL PRIMARY KEY, name VARCHAR(50));"
docker exec crm-platform psql -U $1 -d crm-platform -c "INSERT INTO test_table (name) VALUES ('test');"

echo "Querying test data:"
docker exec crm-platform psql -U $1 -d crm-platform -c "SELECT * FROM test_table;"

docker exec crm-platform psql -U $1 -d crm-platform -c "DROP TABLE IF EXISTS test_table;"
echo "PostgreSQL test completed successfully!"

echo
echo "=== Test Redis ==="

echo "Testing Redis connection..."
docker exec redis-cache redis-cli ping

echo
echo "Testing Redis operations..."
docker exec redis-cache redis-cli set test_key "hello world"
docker exec redis-cache redis-cli get test_key

echo
echo "Redis memory info:"
docker exec redis-cache redis-cli info memory | grep used_memory_human

echo
echo "Redis eviction policy:"
docker exec redis-cache redis-cli config get maxmemory-policy

docker exec redis-cache redis-cli del test_key
echo "Redis test completed successfully!"

echo
echo "=== Test NATS ==="

echo "Checking NATS server status..."
if curl -f http://localhost:8222/varz > /dev/null 2>&1; then
    echo "✓ NATS monitoring endpoint is accessible"
else
    echo "✗ NATS monitoring endpoint not accessible"
fi

echo
echo "NATS health check:"
if curl -f http://localhost:8222/healthz > /dev/null 2>&1; then
    echo "✓ NATS health check passed"
else
    echo "✗ NATS health check failed"
fi

echo
echo "=== Test Service-to-Service Communication ==="

echo "Testing hostname resolution:"
if docker exec crm-platform getent hosts nats-server > /dev/null 2>&1; then
    echo "✓ PostgreSQL can resolve nats-server hostname"
    NATS_IP=$(docker exec crm-platform getent hosts nats-server | awk '{print $1}')
    echo "  nats-server IP: $NATS_IP"
else
    echo "✗ PostgreSQL cannot resolve nats-server hostname"
fi

if docker exec crm-platform getent hosts redis-cache > /dev/null 2>&1; then
    echo "✓ PostgreSQL can resolve redis-cache hostname"
    REDIS_IP=$(docker exec crm-platform getent hosts redis-cache | awk '{print $1}')
    echo "  redis-cache IP: $REDIS_IP"
else
    echo "✗ PostgreSQL cannot resolve redis-cache hostname"
fi

echo
echo "Testing port connectivity:"
if docker exec crm-platform sh -c 'timeout 5 bash -c "</dev/tcp/redis-cache/6379"' 2>/dev/null; then
    echo "✓ Redis port 6379 is accessible from PostgreSQL container"
else
    echo "✗ Redis port 6379 is not accessible from PostgreSQL container"
fi

if docker exec crm-platform sh -c 'timeout 5 bash -c "</dev/tcp/nats-server/4222"' 2>/dev/null; then
    echo "✓ NATS port 4222 is accessible from PostgreSQL container"
else
    echo "✗ NATS port 4222 is not accessible from PostgreSQL container"
fi

echo
echo "=== Test External Connectivity ==="

echo "Testing external access to services..."

# Test PostgreSQL from host using port 5433
echo "Testing PostgreSQL from host (port 5433)..."
export PGPASSWORD=$1

if psql -h localhost -p 5433 -U $1 -d crm-platform -c "SELECT 1;" > /dev/null 2>&1; then
    echo "✓ PostgreSQL is accessible from host on port 5433"
elif psql -h 127.0.0.1 -p 5433 -U $1 -d crm-platform -c "SELECT 1;" > /dev/null 2>&1; then
    echo "✓ PostgreSQL is accessible from host on 127.0.0.1:5433"
else
    echo "✗ PostgreSQL is not accessible from host"
    echo "  Try: export PGPASSWORD=$1 && psql -h localhost -p 5433 -U $1 -d crm-platform"
fi

# Test Redis from host
if command -v redis-cli >/dev/null 2>&1; then
    echo "Testing Redis from host..."
    if redis-cli -h localhost -p 6379 ping > /dev/null 2>&1; then
        echo "✓ Redis is accessible from host"
    else
        echo "✗ Redis is not accessible from host"
    fi
else
    echo "Note: redis-cli not installed - you can install it with:"
    echo "  macOS: brew install redis"
    echo "  Ubuntu: sudo apt-get install redis-tools"
fi

# Test NATS monitoring from host
echo "Testing NATS monitoring from host..."
if curl -f http://localhost:8222/healthz > /dev/null 2>&1; then
    echo "✓ NATS monitoring is accessible from host"
else
    echo "✗ NATS monitoring is not accessible from host"
fi

echo
echo "=== Network Diagnostics ==="

echo "Container network information:"
echo "PostgreSQL container IP: $(docker inspect crm-platform -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}')"
echo "Redis container IP: $(docker inspect redis-cache -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}')"
echo "NATS container IP: $(docker inspect nats-server -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}')"

echo
echo "Port mappings:"
echo "PostgreSQL: $(docker port crm-platform)"
echo "Redis: $(docker port redis-cache)"
echo "NATS: $(docker port nats-server)"

echo
echo "=== Resource and Performance Checks ==="

echo "Container resource usage:"
docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}\t{{.BlockIO}}"

echo
echo "Docker system disk usage:"
docker system df

echo
echo "PostgreSQL data directory usage:"
docker exec crm-platform du -sh /var/lib/postgresql/data 2>/dev/null || echo "Could not check PostgreSQL data directory"

echo
echo "=== Health Check Complete ==="
echo
echo "Connection Information:"
echo "- PostgreSQL: localhost:5433 (user: $1, password: $1, database: crm-platform)"
echo "- Redis: localhost:6379"
echo "- NATS: localhost:4222 (monitoring: http://localhost:8222)"
echo
echo "To connect to PostgreSQL from your host:"
echo "export PGPASSWORD=$1 && psql -h localhost -p 5433 -U $1 -d crm-platform"