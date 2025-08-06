#!/bin/bash

# make generate codegen=<graphql|sqlc>
# if codegen is sqlc, then service is required (service=auth)

gql_codegen() {
  cd "services/api-gateway" && go generate ./...
}

sqlc_codegen() {
  case "$service" in
    auth)
      cd "services/auth-service" && sqlc generate
      ;;
    security)
      cd "services/security-service" && sqlc generate
      ;;
    user)
      cd "services/user-service" && sqlc generate
      ;;
    *)
      echo "Invalid service name. Use 'auth' or 'user'."
      exit 1
      ;;
  esac
}

case "$codegen" in
  graphql)
    gql_codegen
    ;;
  sqlc)
    if [[ -z "$service" ]]; then
      echo "Service name is required for SQLC code generation."
      exit 1
    fi
    sqlc_codegen "$service"
    ;;
  *)
    echo "Usage: make generate codegen=<graphql|sqlc> [service=<service_name>]"
    exit 1
    ;;
esac