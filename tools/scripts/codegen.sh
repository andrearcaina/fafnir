#!/bin/bash

# make generate codegen=<graphql|sqlc> [service=<service_name>]

case "$codegen" in
  graphql)
    cd "services/api-gateway" && go generate ./...
    ;;
  sqlc)
    [[ -z "$service" ]] && { echo "Service name required for SQLC. Use: auth, security, user"; exit 1; }
    case "$service" in
      auth|security|user) cd "services/$service-service" && sqlc generate ;;
      *) echo "Invalid service. Use: auth, security, user"; exit 1 ;;
    esac
    ;;
  *) echo "Usage: make generate codegen=<graphql|sqlc> [service=<service_name>]"; exit 1 ;;
esac