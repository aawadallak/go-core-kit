.PHONY: build test lint fmt tidy redis-local redis-stop

## Development

build:
	@go build ./...

test:
	@go test ./core/... ./plugin/txm/... ./pkg/... -race -count=1

test-all:
	@go test ./... -race -count=1

lint:
	@golangci-lint run

fmt:
	@gofmt -w -s .

tidy:
	@go mod tidy

check: fmt tidy lint test
	@echo "All checks passed."

## Redis (for local development / integration tests)

REDIS_PORT ?= 6379
REDIS_CONTAINER_NAME ?= redis-local

redis-local:
	@echo "Starting Redis container..."
	@docker run --name $(REDIS_CONTAINER_NAME) \
		-p $(REDIS_PORT):6379 \
		-d redis:alpine \
		redis-server --appendonly yes
	@echo "Redis running on port $(REDIS_PORT)"

redis-stop:
	@echo "Stopping Redis container..."
	@docker stop $(REDIS_CONTAINER_NAME)
	@docker rm $(REDIS_CONTAINER_NAME)
	@echo "Redis container stopped and removed"
