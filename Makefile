# Name of the Redis Docker container
REDIS_CONTAINER_NAME = redis-container

# Interval in seconds to check if Redis is ready
REDIS_READY_CHECK_INTERVAL = 1

# Maximum number of retries to check if Redis is ready
REDIS_READY_MAX_RETRIES = 30

# Start Redis in a Docker container
redis-up:
	docker-compose up -d

# Wait for Redis to be ready
redis-ready:
	@echo "Waiting for Redis to be ready..."
	@RETRIES=0; \
	while [ $$RETRIES -lt $(REDIS_READY_MAX_RETRIES) ]; do \
		nc -z localhost 6379 && break; \
		echo "Retrying in $(REDIS_READY_CHECK_INTERVAL) seconds..."; \
		sleep $(REDIS_READY_CHECK_INTERVAL); \
		RETRIES=$$((RETRIES+1)); \
	done; \
	if [ $$RETRIES -eq $(REDIS_READY_MAX_RETRIES) ]; then \
		echo "Timed out waiting for Redis to be ready"; \
		exit 1; \
	fi

# Run redis
local-redis: redis-up redis-ready

# Run Go tests
integration-test: local-redis
	go test ./integration_tests/...

# Run unit tests
unit-test: 
	go test ./internal/...

# Run all local tests
local-all: unit-test integration-test

# Run integration tests for GitHub Actions
integration-test-gha: 
	go test ./integration_tests/...

# Run all tests for GitHub Actions
all-gha: unit-test integration-test-gha

# Stop Redis container
local-clean:
	docker-compose down

# Run linting using golangci-lint
lint:
	golangci-lint run ./...