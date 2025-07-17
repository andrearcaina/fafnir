.PHONY: run-auth-service run-user-service run-web-app build run status stop rm-volumes prune clean reset migrate-up migrate-down migrate-status migrate-create

format := "table {{.Name}}\t{{.Status}}\t{{.Ports}}"

run-auth-service:
	docker compose -p fafnir --env-file infra/env/.env -f infra/docker-compose.yml up -d auth-service

run-user-service:
	docker compose -p fafnir --env-file infra/env/.env -f infra/docker-compose.yml up -d user-service

run-web-app:
	docker compose -p fafnir --env-file infra/env/.env -f infra/docker-compose.yml up -d web-app

build:
	docker compose -p fafnir --env-file infra/env/.env -f infra/docker-compose.yml build --pull --no-cache

run:
	docker compose -p fafnir --env-file infra/env/.env -f infra/docker-compose.yml up -d

status:
	docker compose -p fafnir --env-file infra/env/.env -f infra/docker-compose.yml ps --format=${format}

stop:
	docker compose -p fafnir --env-file infra/env/.env -f infra/docker-compose.yml down --remove-orphans

rm-volumes:
	docker volume rm fafnir_pgdata || true

prune:
	docker images --format "{{.Repository}}" | grep "^fafnir-" | xargs -r docker rmi || true
	docker builder prune -a -f
	
clean: stop prune rm-volumes

reset: clean build run

migrate-up:
	export $$(cat infra/env/.env | xargs) && goose -dir infra/db/migrations up

migrate-down:
	export $$(cat infra/env/.env | xargs) && goose -dir infra/db/migrations down

migrate-status:
	export $$(cat infra/env/.env | xargs) && goose -dir infra/db/migrations status

# make migrate-create name=seed_db
migrate-create:
	goose -dir infra/db/migrations create $(name) sql
