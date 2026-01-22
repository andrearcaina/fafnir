#!/bin/bash

set -e

COMPOSE_CMD="docker compose -p fafnir --env-file infra/env/.env.dev"
PROD_CMD="docker compose -p fafnir --env-file infra/env/.env.prod"
BASE_FILES="-f deployments/docker/compose.dev.yml"
PROD_FILES="-f deployments/docker/compose.prod.yml"
MONITORING_FILES="$BASE_FILES -f deployments/docker/monitoring.yml"

case "$1" in
  auth-service|user-service|security-service|stock-service|api-gateway)
    $COMPOSE_CMD $BASE_FILES up -d "$1" ;;
  build)
    $COMPOSE_CMD $BASE_FILES build --pull --no-cache ;;
  run)
    FILES=${2:+$MONITORING_FILES}
    $COMPOSE_CMD ${FILES:-$BASE_FILES} up -d ;;
  build-prod)
    $PROD_CMD $PROD_FILES build --pull --no-cache ;;
  prod)
    $PROD_CMD $PROD_FILES up -d ;;
  stats)
    docker stats $(docker ps --format '{{.Names}}' | grep '^fafnir-') ;;
  start|pause)
    $COMPOSE_CMD $BASE_FILES $1 ;;
  stop)
    $COMPOSE_CMD $BASE_FILES down --volumes --remove-orphans ;;
  status)
    $COMPOSE_CMD $BASE_FILES ps --format 'table {{.Name}}\t{{.Status}}\t{{.Ports}}' ;;
  logs)
    $COMPOSE_CMD $BASE_FILES logs --tail=100 -f ${2:+"$2"} ;;
  nats)
    docker run --network fafnir_fafnir-network --rm -it natsio/nats-box ;;
  rm-volumes)
    docker volume rm fafnir_pgdata fafnir_prom_data 2>/dev/null || true ;;
  prune)
    docker images --format "{{.Repository}}" | grep -E '^(fafnir-|prom/|grafana/)' | xargs -r docker rmi 2>/dev/null || true
    docker builder prune -a -f ;;
  *)
    echo "Usage: $0 {service|build|prod|run [monitoring]|start|pause|stop|status|logs [service]|nats|rm-volumes|prune}"
esac
