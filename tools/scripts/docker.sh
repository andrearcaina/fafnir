#!/bin/bash

set -e

BASE_COMPOSE="-f deployments/compose/base.yml"
MONITORING_COMPOSE="-f deployments/compose/monitoring.yml"
ENV_FILE="--env-file infra/env/.env.dev"
PROJECT="-p fafnir"
FORMAT='table {{.Name}}\t{{.Status}}\t{{.Ports}}'

case "$1" in
  auth-service|user-service|security-service|api-gateway|web-app)
    docker compose $PROJECT $ENV_FILE $COMPOSE_FILES up -d "$1"
    ;;
  build)
    docker compose $PROJECT $ENV_FILE $BASE_COMPOSE build --pull --no-cache
    ;;
  run)
    COMPOSE_FILES="$BASE_COMPOSE"
    if [[ "$2" == "true" ]]; then
      COMPOSE_FILES+=" $MONITORING_COMPOSE"
    fi
    docker compose $PROJECT $ENV_FILE $COMPOSE_FILES up -d
    ;;
  start)
    docker compose $PROJECT $ENV_FILE $BASE_COMPOSE start
    ;;
  pause)
    docker compose $PROJECT $ENV_FILE $BASE_COMPOSE pause
    ;;
  stop)
    docker compose $PROJECT $ENV_FILE $BASE_COMPOSE down --volumes --remove-orphans
    ;;
  status)
    docker compose $PROJECT $ENV_FILE $BASE_COMPOSE ps --format "$FORMAT"
    ;;
  logs)
    if [[ -z "$2" ]]; then
      docker compose $PROJECT $ENV_FILE $BASE_COMPOSE logs --tail=100 -f
    else
      docker compose $PROJECT $ENV_FILE $BASE_COMPOSE logs --tail=100 -f "$2"
    fi
    ;;
  rm-volumes)
    docker volume rm fafnir_pgdata fafnir_prom_data || true
    ;;
  prune)
    docker images --format "{{.Repository}}" | grep -E '^(fafnir-|prom/|grafana/)' | xargs -r docker rmi || true
    docker builder prune -a -f
    ;;
  *)
    echo "Usage: $0 {auth-service|user-service|security-service|api-gateway|web-app|build|run [monitoring]|start|pause|stop|status|logs [service]|rm-volumes|prune}"
    exit 1
    ;;
esac