#!/bin/bash

set -e

# make generate codegen=<graphql|sqlc|proto> service=<service_name>

case "$codegen" in
  graphql)
    cd "src/api-gateway" && go generate ./...
    ;;
  sqlc)
    [[ -z "$service" ]] && { echo "Service name required for SQLC. Use: auth, security, user, stock, order, portfolio"; exit 1; }
    case "$service" in
      auth|security|user|stock|order|portfolio) cd "src/$service-service" && sqlc generate ;;
      *) echo "Invalid service. Use: auth, security, user, stock, order, portfolio"; exit 1 ;;
    esac
    ;;
  proto)
    [[ -z "$service" ]] && { echo "Service name required for Protobuf. Use: base, security, user, stock, order, portfolio"; exit 1; }
    case "$service" in
      base|security|user|stock|order|portfolio)
        cd "src/shared"
        protoc -I=../../proto \
          --go_out=pb/$service --go_opt=paths=source_relative \
          --go-grpc_out=pb/$service --go-grpc_opt=paths=source_relative \
          ../../proto/$service.proto
      ;;
      *) echo "Invalid service. Use: base, security, user, stock, order, portfolio"; exit 1 ;;
    esac
    ;;
  *) echo "Usage: make generate codegen=<graphql|sqlc|proto> service=<service_name>"; exit 1 ;;
esac
