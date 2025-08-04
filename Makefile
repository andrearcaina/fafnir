.PHONY: default help run-auth-service run-user-service run-api-gateway run-web-app build start pause run stop status rm-volumes prune clean reset migrate-up migrate-down migrate-status migrate-create generate seed

### default makefile target command (runs help)

default: help

help:
	./scripts/help.sh

### docker commands for running services

run-auth-service:
	./scripts/docker.sh run-auth-service

run-user-service:
	./scripts/docker.sh run-user-service

run-api-gateway:
	./scripts/docker.sh run-api-gateway

run-web-app:
	./scripts/docker.sh run-web-app

build:
	./scripts/docker.sh build

start:
	./scripts/docker.sh start

pause:
	./scripts/docker.sh pause

run:
	./scripts/docker.sh run

stop:
	./scripts/docker.sh stop

status:
	./scripts/docker.sh status

prune:
	./scripts/docker.sh prune

rm-volumes:
	./scripts/docker.sh rm-volumes

clean: stop prune rm-volumes

reset: clean build run

### migration commands for database

migrate-up:
	./scripts/migrate.sh up

migrate-down:
	./scripts/migrate.sh down

migrate-status:
	./scripts/migrate.sh status

# make migrate-create connections=auth name=seed_some_data
migrate-create:
	./scripts/migrate.sh create $(db) $(name)

#### codegen commands for generating GraphQL and SQLc go code from .graphqls and .sql files

# make generate codegen=<graphql|sqlc>
# if codegen is sqlc, then service is required (service=auth)
generate:
	./scripts/codegen.sh generate $(codegen) $(service)

### seed commands for seeding certain databases

# make seed connections=<db_name> (db_name=auth)
# or seed all databases (connections=all)
seed:
	./scripts/seed.sh $(db)