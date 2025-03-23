# Redis configuration
REDIS_PORT ?= 6379
REDIS_CONTAINER_NAME ?= redis-local

# Run Redis locally with Docker
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
