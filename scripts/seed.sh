#!/bin/bash

set -e

seed_auth_db() {
  cd services/auth-service && go run cmd/seeder/main.go
}

seed_security_db() {
  cd services/security-service && go run cmd/seeder/main.go
}

seed_user_db() {
  cd services/user-service && go run cmd/seeder/main.go
}

seed_all() {
  seed_auth_db
  cd ../..
  seed_security_db
  cd ../..
  seed_user_db
}

case "$1" in
  auth)
    seed_auth_db
    ;;
  security)
    seed_security_db
    ;;
  user)
    seed_user_db
    ;;
  all)
    seed_all
    ;;
  *)
    echo "Usage: $0 {auth|security|user|all}"
    exit 1
    ;;
esac
