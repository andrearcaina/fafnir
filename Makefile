.PHONY: default help run-auth-service run-user-service run-security-service run-api-gateway run build start pause stop status logs rm-volumes prune clean reset migrate-up migrate-down migrate-status migrate-create generate seed

### default makefile target command (runs help)

default: help

help:
	@echo "Usage: make <target> [options]"

### docker commands for running services

run-auth-service:
	./tools/scripts/docker.sh auth-service

run-user-service:
	./tools/scripts/docker.sh user-service

run-security-service:
	./tools/scripts/docker.sh security-service

run-stock-service:
	./tools/scripts/docker.sh stock-service

run-api-gateway:
	./tools/scripts/docker.sh api-gateway

run:
	./tools/scripts/docker.sh run $(monitoring)

build:
	./tools/scripts/docker.sh build

start:
	./tools/scripts/docker.sh start

pause:
	./tools/scripts/docker.sh pause

stop:
	./tools/scripts/docker.sh stop

status:
	./tools/scripts/docker.sh status

logs:
	./tools/scripts/docker.sh logs

rm-volumes:
	./tools/scripts/docker.sh rm-volumes

prune:
	./tools/scripts/docker.sh prune

clean: stop prune rm-volumes

reset: clean build run

### migration commands for database

migrate-up:
	./tools/scripts/migrate.sh up

migrate-down:
	./tools/scripts/migrate.sh down

migrate-status:
	./tools/scripts/migrate.sh status

# make migrate-create db=auth name=seed_some_data
migrate-create:
	./tools/scripts/migrate.sh create $(db) $(name)

#### codegen commands for generating GraphQL and SQLc go code from .graphqls and .sql files

# make generate codegen=<graphql|sqlc>
# if codegen is sqlc, then service is required (service=auth)
generate:
	./tools/scripts/codegen.sh generate $(codegen) $(service)

### seed commands for seeding certain databases

# make seed db=<db_name> (db_name=auth)
# or seed all databases (db=all)
seed:
	cd tools/seeder && go run main.go --db $(db)
