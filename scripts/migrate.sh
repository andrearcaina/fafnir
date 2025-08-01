#!/bin/bash

set -e

source "infra/env/.env.dev"

migrate_up() {
  GOOSE_DRIVER=$GOOSE_DRIVER GOOSE_DBSTRING=$INFRA_DB_STRING goose -dir infra/postgres/migrations up && \
  GOOSE_DRIVER=$GOOSE_DRIVER GOOSE_DBSTRING=$AUTH_DB_STRING goose -dir services/auth-service/internal/db/migrations up && \
  GOOSE_DRIVER=$GOOSE_DRIVER GOOSE_DBSTRING=$USER_DB_STRING goose -dir services/user-service/internal/db/migrations up
}

migrate_down() {
  GOOSE_DRIVER=$GOOSE_DRIVER GOOSE_DBSTRING=$INFRA_DB_STRING goose -dir infra/postgres/migrations down && \
  GOOSE_DRIVER=$GOOSE_DRIVER GOOSE_DBSTRING=$AUTH_DB_STRING goose -dir services/auth-service/internal/db/migrations down && \
  GOOSE_DRIVER=$GOOSE_DRIVER GOOSE_DBSTRING=$USER_DB_STRING goose -dir services/user-service/internal/db/migrations down
}

migrate_status() {
  echo "Migration status for all databases:"

  echo "
The 'database instance' migration" && \
  GOOSE_DRIVER=$GOOSE_DRIVER GOOSE_DBSTRING=$INFRA_DB_STRING goose -dir infra/postgres/migrations status && \
  echo "
The 'auth service' migration" && \
  GOOSE_DRIVER=$GOOSE_DRIVER GOOSE_DBSTRING=$AUTH_DB_STRING goose -dir services/auth-service/internal/db/migrations status && \
  echo "
The 'user service' migration" && \
  GOOSE_DRIVER=$GOOSE_DRIVER GOOSE_DBSTRING=$USER_DB_STRING goose -dir services/user-service/internal/db/migrations status
}

migrate_create() {
  db="$1"
  name="$2"
  if [ -z "$db" ]; then
    echo "Error: db argument is required. Use db=auth_db, db=user_db, or db=infra_db."
    exit 1
  fi
  if [ -z "$name" ]; then
    echo "Error: name argument is required. Use name=<migration_name>."
    exit 1
  fi
  case "$db" in
    auth_db)
      goose -dir services/auth-service/internal/db/migrations create "$name" sql
      ;;
    user_db)
      goose -dir services/user-service/internal/db/migrations create "$name" sql
      ;;
    infra_db)
      goose -dir infra/postgres/migrations create "$name" sql
      ;;
    *)
      echo "Error: Invalid db argument. Use db=auth_db, db=user_db, or db=infra_db."
      exit 1
      ;;
  esac
}

case "$1" in
  up) migrate_up ;;
  down) migrate_down ;;
  status) migrate_status ;;
  create) migrate_create "$2" "$3" ;;
  *) echo "Usage: $0 {up|down|status|create <db> <name>}"; exit 1 ;;
esac