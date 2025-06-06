SERVICES := $(shell find services -maxdepth 1 -mindepth 1 -type d -exec basename {} \;)
PWD := $(shell pwd)

show-services:
	echo "Found services: $(SERVICES)"

build-all:
	set -e; \
	for service in $(SERVICES); do \
		cd $(PWD) \
		echo "Building $$service..."; \
		mkdir -p services/$$service/bin; \
		cd services/$$service; \
		go mod tidy; \
		go build -o bin/$$service ./cmd/server; \
	done

clean:
	set -e; \
	for service in $(SERVICES); do \
		rm -rf services/$$service/bin; \
		rm services/$$service/coverage.out; \
	done

test:
	for service in $(SERVICES); do \
		cd $(PWD); \
		echo "Testing $$service..."; \
		cd services/$$service && go test -v -race -coverprofile=coverage.out ./...; \
		cd ../..; \
	done