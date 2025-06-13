## Container Status

```bash
# See all containers and their health status
docker-compose ps

# View real-time logs from all services
docker-compose logs -f

# Check if all containers are running
docker ps
```

## Test PostgreSQL

```bash
docker exec crm-platform psql -U admin -d crm-platform -c "SELECT version();"# Test basic connection
docker exec crm-platform psql -U admin -d crm-platform -c "SELECT version();"

# List all databases
docker exec crm-platform psql -U admin -d crm-platform -c "\l"

# Test creating a simple table (verify write permissions)
docker exec crm-platform psql -U admin -d crm-platform -c "CREATE TABLE test_table (id SERIAL PRIMARY KEY, name VARCHAR(50));"

# Insert test data
docker exec crm-platform psql -U admin -d crm-platform -c "INSERT INTO test_table (name) VALUES ('test');"

# Query test data
docker exec crm-platform psql -U admin -d crm-platform -c "SELECT * FROM test_table;"

# Clean up test table
docker exec crm-platform psql -U admin -d crm-platform -c "DROP TABLE test_table;"
```

## Test Redis

```bash
# Test basic connection
docker exec redis redis-cli ping

# Test setting and getting a value
docker exec redis redis-cli set test_key "hello world"
docker exec redis redis-cli get test_key

# Check Redis info and memory usage
docker exec redis redis-cli info memory

# Test the eviction policy is working
docker exec redis redis-cli config get maxmemory-policy

# Clean up test data
docker exec redis redis-cli del test_key
```

## Test NATS

```bash
# Check NATS server status via HTTP monitoring
curl http://localhost:8222/varz

# Check connections info
curl http://localhost:8222/connz

# Test publishing a message (requires NATS CLI)
# First install NATS CLI: https://github.com/nats-io/natscli
nats pub test.subject "Hello NATS"

# Subscribe to messages (run in separate terminal)
nats sub test.subject
```

## Test S2S Communication

```bash
# Test if services can reach each other by name within the network
docker exec crm-platform ping -c 3 nats-server
docker exec crm-platform ping -c 3 redis-cache

# Test if Redis is accessible from PostgreSQL container
docker exec crm-platform nc -zv redis-cache 6379

# Test if NATS is accessible from PostgreSQL container
docker exec crm-platform nc -zv nats-server 4222
```

## Test External Connectivity

```bash
# Test PostgreSQL from your host machine (requires psql installed)
psql -h localhost -p 5432 -U admin -d crm-platform -c "SELECT 1;"

# Test Redis from host (requires redis-cli installed)
redis-cli -h localhost -p 6379 ping

# Test NATS monitoring from host
curl http://localhost:8222/healthz
```

## Resource and Performance Checks

```bash
# Check container resource usage
docker stats

# Check disk usage for volumes
docker system df

# Check specific volume usage
docker exec crm-platform du -sh /var/lib/postgresql/data
```

## Network Connectivity Test

```bash
# Inspect the Docker network
docker network inspect $(docker-compose ps -q | head -1 | xargs docker inspect --format='{{range .NetworkSettings.Networks}}{{.NetworkID}}{{end}}')

# List all networks
docker network ls
```