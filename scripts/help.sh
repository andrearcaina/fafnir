#!/bin/bash

GREEN='\033[0;32m'
NC='\033[0m'

echo "Makefile commands:"
echo ""
echo "Services:"
echo -e "${GREEN}run-auth-service${NC}     - Run the auth service"
echo -e "${GREEN}run-user-service${NC}     - Run the user service"
echo -e "${GREEN}run-api-gateway${NC}      - Run the API gateway"
echo -e "${GREEN}run-web-app${NC}          - Run the web application"
echo ""
echo "Build & Run:"
echo -e "${GREEN}build${NC}                - Build all containers"
echo -e "${GREEN}run${NC}                  - Start all containers"
echo ""
echo "Manage Services:"
echo -e "${GREEN}status${NC}               - Show status of all containers"
echo -e "${GREEN}stop${NC}                 - Stop all containers"
echo ""
echo "Docker Cleanup:"
echo -e "${GREEN}rm-volumes${NC}           - Remove Docker volumes used by the project (fafnir_pgdata, fafnir_prom_data)"
echo -e "${GREEN}prune${NC}                - Prune unused Docker images and builders"
echo -e "${GREEN}clean${NC}                - Runs stop, prune, and rm-volumes"
echo -e "${GREEN}reset${NC}                - Reset environment (runs clean, build, and run)"
echo ""
echo "Database Migrations:"
echo -e "${GREEN}migrate-up${NC}           - Apply database migrations up"
echo -e "${GREEN}migrate-down${NC}         - Rollback database migrations down"
echo -e "${GREEN}migrate-status${NC}       - Show migration status"
echo -e "${GREEN}migrate-create db=<db_name> name=<migration_name>${NC} - Create a new migration with the specified db and migration name"