.PHONY: run-auth-service run-user-service run status stop migrate-up migrate-down migrate-status migrate-create

format := "table {{.Name}}\t{{.Status}}\t{{.Ports}}"

run-auth-service:
	docker compose -p den --env-file infra/env/.env -f infra/docker-compose.yml up -d auth-service

run-user-service:
	docker compose -p den --env-file infra/env/.env -f infra/docker-compose.yml up -d user-service

run:
	docker compose -p den --env-file infra/env/.env -f infra/docker-compose.yml up -d

status:
	docker compose -p den --env-file infra/env/.env -f infra/docker-compose.yml ps --format=${format}

stop:
	docker compose -p den --env-file infra/env/.env -f infra/docker-compose.yml down --remove-orphans

migrate-up:
	export $$(cat infra/env/.env | xargs) && goose -dir infra/db/migrations up

migrate-down:
	export $$(cat infra/env/.env | xargs) && goose -dir infra/db/migrations down

migrate-status:
	export $$(cat infra/env/.env | xargs) && goose -dir infra/db/migrations status

migrate-create:
	goose -dir infra/db/migrations create $(name) sql