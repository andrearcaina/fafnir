#!/bin/bash

set -e
source "infra/env/.env.dev"

DB_BASE="$GOOSE_DRIVER://$POSTGRES_USER:$POSTGRES_PASSWORD@$DB_HOST_LOCAL:$DB_PORT"
SERVICES=(auth security user)

goose_cmd() {
  local service=$1 action=$2
  local db_var="${service^^}_DB"
  GOOSE_DRIVER=$GOOSE_DRIVER GOOSE_DBSTRING="$DB_BASE/${!db_var}?sslmode=disable" \
    goose -dir "services/$service-service/internal/db/migrations" $action
}

case "$1" in
  up|down) for s in "${SERVICES[@]}"; do goose_cmd $s $1; done ;;
  status) for s in "${SERVICES[@]}"; do echo -e "\n$s service:"; goose_cmd $s status; done ;;
  create)
    [[ $# -lt 3 ]] && { echo "Usage: $0 create <db> <name>"; exit 1; }
    case "$2" in
      auth|security|user) goose -dir "services/$2-service/internal/db/migrations" create "$3" sql ;;
      infra) goose -dir "infra/postgres/migrations" create "$3" sql ;;
      *) echo "Invalid db. Use: auth, security, user, infra"; exit 1 ;;
    esac ;;
  *) echo "Usage: $0 {up|down|status|create <db> <name>}"; exit 1 ;;
esac