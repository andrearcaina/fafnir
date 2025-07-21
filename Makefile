.PHONY: run-auth-service run-user-service run-api-gateway run-web-app build run status stop rm-volumes prune clean reset migrate-up migrate-down migrate-status migrate-create

format := "table {{.Name}}\t{{.Status}}\t{{.Ports}}"

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
