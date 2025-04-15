# Make ENV file for DEV environment
env.dev:
	@echo "Make ENV file for DEV environment..."
	[[ -f ./envs/dev/.env ]] && echo "ENV file already exists" || cp ./envs/dev/.env.default ./envs/dev/.env

# Check if ENV file for DEV environment exists
.env.dev.exists:
	@echo "Check if ENV file for DEV environment exists..."
	[[ -f ./envs/dev/.env ]] || (echo "ENV file does not exist" && exit 1)

# Make ENV file for STABLE environment
env.stable:
	@echo "Make ENV file for STABLE environment..."
	[[ -f ./envs/stable/.env ]] && echo "ENV file already exists" || cp ./envs/stable/.env.default ./envs/stable/.env

# Check if ENV file for STABLE environment exists
.env.stable.exists:
	@echo "Check if ENV file for STABLE environment exists..."
	[[ -f ./envs/stable/.env ]] || (echo "ENV file does not exist" && exit 1)

# Deploy ENV environment
REMOVE_VOLUMES=false
up.dev: .env.dev.exists
	@echo "Deploying DEV environment..."
	if [ $(REMOVE_VOLUMES) == true ]; then \
		echo "Removing volumes..." && \
		docker compose             \
		-f docker-compose.dev.yaml \
		-f ./bricks/mongodb.yaml   \
		-f ./bricks/rabbitmq.yaml  \
		 rm --volumes --force;     \
	 fi
	@docker compose \
		-f docker-compose.dev.yaml \
		-f ./bricks/mongodb.yaml   \
		-f ./bricks/rabbitmq.yaml  \
		--env-file ./envs/dev/.env \
		up               \
		--detach         \
		--build          \
		--force-recreate \
		--remove-orphans

# Deploy STABLE environment
up.stable: .env.stable.exists
	@echo "Deploying STABLE environment..."
	@docker compose \
		-f docker-compose.stable.yaml  \
		-f ./bricks/mongodb.yaml       \
		-f ./bricks/rabbitmq.yaml      \
		-f ./bricks/manager.yaml       \
		-f ./bricks/worker.yaml        \
		-f ./bricks/reverse-proxy.yaml \
		--env-file ./envs/stable/.env  \
		up               \
		--detach         \
		--force-recreate \
		--remove-orphans \
		--scale worker=3

# Print available commands
help:
	@echo "Available commands:"
	@echo "  env.dev    - Make ENV file for DEV environment"
	@echo "  env.stable - Make ENV file for STABLE environment"
	@echo "  up.dev     - Deploy DEV environment (REMOVE_VOLUMES=false)"
	@echo "  up.stable  - Deploy STABLE environment"

.DEFAULT_GOAL := help
.PHONY: help env.dev env.stable up.dev up.stable