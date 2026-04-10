COMPOSE := $(or $(shell command -v docker-compose >/dev/null 2>&1 && echo docker-compose), $(shell command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1 && echo "docker compose"))

# Short aliases for commonly used docker compose invocations
DC := $(COMPOSE)
DC_UP := $(DC) up -d
DC_UP_BUILD := $(DC) up -d --build
DC_DOWN := $(DC) down -v
DC_RUN := $(DC) run --rm
DC_BUILD := $(DC) build --no-cache
DC_PS := $(DC) ps
DC_LOGS := $(DC) logs -f --tail=200
DC_EXEC := $(DC) exec

ifeq ($(COMPOSE),)
$(error docker-compose or Docker Compose plugin not found — install Docker Compose or enable the Compose plugin)
endif

# Default environment variables (can be overridden by exporting or via .env)
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_NAME ?= appdb
DB_USER ?= appuser
DB_PASSWORD ?= apppass

# Compose helper
# Download Go module dependencies for local development
deps:
	@echo "Downloading Go module dependencies..."
	go mod download

.PHONY: deps up clean start rebuild down up-api db-drop flyway-migrate flyway-info flyway-history flyway-clean rebuild run migrate ps logs

 


up:
	$(DC_UP_BUILD)

.PHONY: clean start
# Remove containers, networks, and volumes created by compose
clean:
	$(DC_DOWN) --remove-orphans

# Start core services required for the project
start:
	$(DC_UP) postgres redis zookeeper kafka migrate

rebuild:
	$(DC_BUILD)


down:
	$(DC_DOWN)

# Start only the API service (build and run detached)
up-api:
	$(DC_UP_BUILD) api

# Drop all database objects (destructive). Run manually when needed.
db-drop: flyway-clean
	@echo "Dropped database objects via flyway-clean"

# Apply migrations using Flyway image (creates flyway_schema_history similar to Flyway/Spring Boot)
flyway-migrate:
	$(DC_RUN) migrate

flyway-info:
	$(DC_RUN) migrate info

.PHONY: all

flyway-history:
	$(DC_EXEC) postgres psql -U $(DB_USER) -d $(DB_NAME) -c "SELECT * FROM flyway_schema_history ORDER BY installed_rank;"

flyway-clean:
	$(DC_RUN) migrate -cleanDisabled=false clean

.PHONY: rebuild

# Convenience targets
run: deps
	go run cmd/api/main.go

migrate: flyway-migrate

ps:
	$(DC_PS)

logs:
	$(DC_LOGS)


