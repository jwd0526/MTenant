#!/bin/bash

# USAGE:
#
# ./test.sh <psql username> <psql password>
#
# If prompted to "Run './deal-service'...", build if not already:
#    go build -o deal-service ./cmd/server && ./deal-service
# Then press any key to setup tenants and run tests.

USERNAME=$1
PASSWORD=$2

export ENVIRONMENT=dev
export DATABASE_URL="postgres://$USERNAME:$PASSWORD@localhost:5433/crm-platform?sslmode=disable" ## Must be psql uri

go build -o deal-service ./cmd/server

while ! lsof -i :8080 >/dev/null 2>&1; do
    read -p "Run './deal-service', then press any key..."
done

echo "Service detected on port 8080"

echo "Setting up test-tenants"

# Restart crusty ass old ass tenants and birth new ones

go run ../../scripts/setup_test_tenants.go reset

echo "Verifying setup"

# make sure they made it ;)

go run ../../scripts/setup_test_tenants.go verify

echo "Running tests..." 

go test ./tests/api -v

# I love go!

go clean -testcache