COMPOSE := $(or $(shell command -v docker-compose >/dev/null 2>&1 && echo docker-compose), $(shell command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1 && echo "docker compose"))

ifeq ($(COMPOSE),)
$(error docker-compose or Docker Compose plugin not found — install Docker Compose or enable the Compose plugin)
endif

# Download Go module dependencies for local development
deps:
	@echo "Downloading Go module dependencies..."
	go mod download

.PHONY: deps up down up-api db-drop flyway-migrate flyway-info flyway-history flyway-clean rebuild

 


up:
	$(COMPOSE) up -d --build

rebuild:
	$(COMPOSE) build --no-cache


down:
	$(COMPOSE) down -v

# Start only the API service (build and run detached)
up-api:
	$(COMPOSE) up -d --build api

# Drop all database objects (destructive). Run manually when needed.
db-drop: flyway-clean
	@echo "Dropped database objects via flyway-clean"

# Apply migrations using Flyway image (creates flyway_schema_history similar to Flyway/Spring Boot)
flyway-migrate:
	docker run --rm -v $(PWD)/internal/infrastructure/database/migrations:/flyway/sql --network host flyway/flyway \
		-url=jdbc:postgresql://localhost:5432/appdb -user=appuser -password=apppass -baselineOnMigrate=true migrate

flyway-info:
	docker run --rm -v $(PWD)/internal/infrastructure/database/migrations:/flyway/sql --network host flyway/flyway \
		-url=jdbc:postgresql://localhost:5432/appdb -user=appuser -password=apppass info

.PHONY: all

flyway-history:
	docker exec go-rest-api-pro-postgres-1 psql -U appuser -d appdb -c "SELECT * FROM flyway_schema_history ORDER BY installed_rank;"

flyway-clean:
	docker run --rm -v $(PWD)/internal/infrastructure/database/migrations:/flyway/sql --network host flyway/flyway \
		-url=jdbc:postgresql://localhost:5432/appdb -user=appuser -password=apppass -cleanDisabled=false clean

.PHONY: rebuild


