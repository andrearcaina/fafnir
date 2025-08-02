.PHONY: default help run-auth-service run-user-service run-api-gateway run-web-app build run status stop rm-volumes prune clean reset migrate-up migrate-down migrate-status migrate-create

# default makefile target command

default: help

help:
	./scripts/help.sh

# docker commands for running services

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

run:
	./scripts/docker.sh run

status:
	./scripts/docker.sh status

stop:
	./scripts/docker.sh stop

prune:
	./scripts/docker.sh prune

rm-volumes:
	./scripts/docker.sh rm-volumes

clean: stop prune rm-volumes

reset: clean build run

# migration commands for database

migrate-up:
	./scripts/migrate.sh up

migrate-down:
	./scripts/migrate.sh down

migrate-status:
	./scripts/migrate.sh status

# make migrate-create postgres=auth_db name=seed_db
migrate-create:
	./scripts/migrate.sh create $(db) $(name)

# gqlgen commands for generating GraphQL schema and resolvers
generate:
	./scripts/gqlgen.sh generate