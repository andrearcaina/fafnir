#!/bin/bash

set -e

compose_files="-f infra/compose.yml -f infra/monitoring/compose.yml"
env_file="--env-file infra/env/.env.dev"
project="-p fafnir"
format='table {{.Name}}\t{{.Status}}\t{{.Ports}}'

case "$1" in
  auth-service)
    docker compose $project $env_file $compose_files up -d auth-service
    ;;
  user-service)
    docker compose $project $env_file $compose_files up -d user-service
    ;;
  api-gateway)
    docker compose $project $env_file $compose_files up -d api-gateway
    ;;
  web-app)
    docker compose $project $env_file $compose_files up -d web-app
    ;;
  build)
    docker compose $project $env_file $compose_files build --pull --no-cache
    ;;
  run)
    docker compose $project $env_file $compose_files up -d
    ;;
  status)
    docker compose $project $env_file $compose_files ps --format="$format"
    ;;
  stop)
    docker compose $project $env_file $compose_files down -v --remove-orphans
    ;;
  rm-volumes)
    docker volume rm fafnir_pgdata fafnir_prom_data || true
    ;;
  prune)
    docker images --format "{{.Repository}}" | grep -E '^(fafnir-|prom/|grafana/)' | xargs -r docker rmi || true
    docker builder prune -a -f
    ;;
  *)
    echo "Usage: $0 {auth-service|user-service|api-gateway|web-app|build|run|status|stop|rm-volumes|prune}"
    exit 1
    ;;
esac