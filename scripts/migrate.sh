#!/bin/bash

set -e

source "infra/env/.env.dev"

infra_db_string="$GOOSE_DRIVER://$POSTGRES_USER:$POSTGRES_PASSWORD@$DB_HOST_LOCAL:$DB_PORT/$GOOSE_DRIVER?sslmode=disable"
auth_db_string="$GOOSE_DRIVER://$POSTGRES_USER:$POSTGRES_PASSWORD@$DB_HOST_LOCAL:$DB_PORT/$AUTH_DB?sslmode=disable"
security_db_string="$GOOSE_DRIVER://$POSTGRES_USER:$POSTGRES_PASSWORD@$DB_HOST_LOCAL:$DB_PORT/$SECURITY_DB?sslmode=disable"
user_db_string="$GOOSE_DRIVER://$POSTGRES_USER:$POSTGRES_PASSWORD@$DB_HOST_LOCAL:$DB_PORT/$USER_DB?sslmode=disable"

migrate_up() {
  GOOSE_DRIVER=$GOOSE_DRIVER GOOSE_DBSTRING="$infra_db_string" goose -dir infra/postgres/migrations up && \
  GOOSE_DRIVER=$GOOSE_DRIVER GOOSE_DBSTRING="$auth_db_string" goose -dir services/auth-service/internal/db/migrations up && \
  GOOSE_DRIVER=$GOOSE_DRIVER GOOSE_DBSTRING="$security_db_string" goose -dir services/security-service/internal/db/migrations up && \
  GOOSE_DRIVER=$GOOSE_DRIVER GOOSE_DBSTRING="$user_db_string" goose -dir services/user-service/internal/db/migrations up
}

migrate_down() {
  GOOSE_DRIVER=$GOOSE_DRIVER GOOSE_DBSTRING="$infra_db_string" goose -dir infra/postgres/migrations down && \
  GOOSE_DRIVER=$GOOSE_DRIVER GOOSE_DBSTRING="$auth_db_string" goose -dir services/auth-service/internal/db/migrations down && \
  GOOSE_DRIVER=$GOOSE_DRIVER GOOSE_DBSTRING="$security_db_string" goose -dir services/security-service/internal/db/migrations down && \
  GOOSE_DRIVER=$GOOSE_DRIVER GOOSE_DBSTRING="$user_db_string" goose -dir services/user-service/internal/db/migrations down
}

migrate_status() {
  echo "Migration status for all databases:"

  echo "
The 'database instance' migration" && \
  GOOSE_DRIVER=$GOOSE_DRIVER GOOSE_DBSTRING="$infra_db_string" goose -dir infra/postgres/migrations status && \
  echo "
The 'auth service' migration" && \
  GOOSE_DRIVER=$GOOSE_DRIVER GOOSE_DBSTRING="$auth_db_string" goose -dir services/auth-service/internal/db/migrations status && \
  echo "
The 'security service' migration" && \
  GOOSE_DRIVER=$GOOSE_DRIVER GOOSE_DBSTRING="$security_db_string" goose -dir services/security-service/internal/db/migrations status && \
  echo "
The 'user service' migration" && \
  GOOSE_DRIVER=$GOOSE_DRIVER GOOSE_DBSTRING="$user_db_string" goose -dir services/user-service/internal/db/migrations status
}

migrate_create() {
  db="$1"
  name="$2"
  if [ -z "$db" ]; then
    echo "Error: db argument is required. Use db=auth, db=user, db=security, or db=infra."
    exit 1
  fi
  if [ -z "$name" ]; then
    echo "Error: name argument is required. Use name=<migration_name>."
    exit 1
  fi
  case "$db" in
    auth)
      goose -dir services/auth-service/internal/db/migrations create "$name" sql
      ;;
    security)
      goose -dir services/security-service/internal/db/migrations create "$name" sql
      ;;
    user)
      goose -dir services/user-service/internal/db/migrations create "$name" sql
      ;;
    infra)
      goose -dir infra/postgres/migrations create "$name" sql
      ;;
    *)
      echo "Error: Invalid db argument. Use db=auth, db=user, db=security, or db=infra."
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