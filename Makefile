REDIS_CONTAINER_NAME = redis-container
REDIS_READY_CHECK_INTERVAL = 1
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

unit-test: 
	go test ./internal/...

local-all: unit-test integration-test

integration-test-gha: 
	go test ./integration_tests/...

all-gha: unit-test integration-test-gha

# Stop Redis container
local-clean:
	docker-compose down