.PHONY: default help run-auth-service run-user-service run-api-gateway run-web-app build run status stop rm-volumes prune clean reset migrate-up migrate-down migrate-status migrate-create

format := "table {{.Name}}\t{{.Status}}\t{{.Ports}}"

RED=\033[0;31m
GREEN=\033[0;32m
NC=\033[0m

default: help

help:
	@echo "Makefile commands:"
	@echo ""
	@echo "Services:"
	@echo "$(GREEN)run-auth-service$(NC)     - Run the auth service"
	@echo "$(GREEN)run-user-service$(NC)     - Run the user service"
	@echo "$(GREEN)run-api-gateway$(NC)      - Run the API gateway"
	@echo "$(GREEN)run-web-app$(NC)          - Run the web application"
	@echo ""
	@echo "Build & Run:"
	@echo "$(GREEN)build$(NC)                - Build all containers"
	@echo "$(GREEN)run$(NC)                  - Start all containers"
	@echo ""
	@echo "Manage Services:"
	@echo "$(GREEN)status$(NC)               - Show status of all containers"
	@echo "$(GREEN)stop$(NC)                 - Stop all containers"
	@echo ""
	@echo "Docker Cleanup:"
	@echo "$(GREEN)rm-volumes$(NC)           - Remove Docker volumes used by the project (fafnir_pgdata, fafnir_prom_data)"
	@echo "$(GREEN)prune$(NC)                - Prune unused Docker images and builders"
	@echo "$(GREEN)clean$(NC)                - Runs stop, prune, and rm-volumes"
	@echo "$(GREEN)reset$(NC)                - Reset environment (runs clean, build, and run)"
	@echo ""
	@echo "Database Migrations:"
	@echo "$(GREEN)migrate-up$(NC)           		     - Apply database migrations up"
	@echo "$(GREEN)migrate-down$(NC) 	  		     - Rollback database migrations down"
	@echo "$(GREEN)migrate-status$(NC)                       - Show migration status"
	@echo "$(GREEN)migrate-create name=<migration_name>$(NC) - Create a new migration with the specified name"

run-auth-service:
	docker compose -p fafnir --env-file infra/env/.env.dev -f infra/compose.yml -f infra/monitoring/compose.yml up -d auth-service

run-user-service:
	docker compose -p fafnir --env-file infra/env/.env.dev -f infra/compose.yml -f infra/monitoring/compose.yml up -d user-service

run-api-gateway:
	docker compose -p fafnir --env-file infra/env/.env.dev -f infra/compose.yml -f infra/monitoring/compose.yml up -d api-gateway

run-web-app:
	docker compose -p fafnir --env-file infra/env/.env.dev -f infra/compose.yml -f infra/monitoring/compose.yml up -d web-app

build:
	docker compose -p fafnir --env-file infra/env/.env.dev -f infra/compose.yml -f infra/monitoring/compose.yml build --pull --no-cache

run:
	docker compose -p fafnir --env-file infra/env/.env.dev -f infra/compose.yml -f infra/monitoring/compose.yml up -d

status:
	docker compose -p fafnir --env-file infra/env/.env.dev -f infra/compose.yml -f infra/monitoring/compose.yml ps --format=${format}

stop:
	docker compose -p fafnir --env-file infra/env/.env.dev -f infra/compose.yml -f infra/monitoring/compose.yml down -v --remove-orphans

rm-volumes:
	docker volume rm fafnir_pgdata fafnir_prom_data || true

prune:
	docker images --format "{{.Repository}}" | grep -E '^(fafnir-|prom/|grafana/)' | xargs -r docker rmi || true
	docker builder prune -a -f

clean: stop prune rm-volumes

reset: clean build run

migrate-up:
	export $$(cat infra/env/.env.dev | xargs) && goose -dir infra/db/migrations up

migrate-down:
	export $$(cat infra/env/.env.dev | xargs) && goose -dir infra/db/migrations down

migrate-status:
	export $$(cat infra/env/.env.dev | xargs) && goose -dir infra/db/migrations status

# make migrate-create name=seed_db
migrate-create:
	goose -dir infra/db/migrations create $(name) sql
