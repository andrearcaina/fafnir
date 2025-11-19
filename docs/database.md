# Database & Migrations

## Architecture Design
- Each microservice has its own dedicated Postgres database
- This ensures data isolation and allows each service to manage its own schema and migrations independently
- Each service connects to its respective database using environment variables defined in the Docker Compose and Kubernetes configuration files

## Overview
- Uses Postgres (see `build/docker/postgres.Dockerfile`)
- Migrations managed with [Goose](https://github.com/pressly/goose)
- Install Goose if you haven't already:
```bash
> go install github.com/pressly/goose/v3/cmd/goose@latest
```

## Running Migrations
- Each service has its own database and migration files located in `src/<service-name>/internal/db/migrations/`
- To run migrations for a specific service, use:
```bash
> make migrate-create db=<service_name> name=<migration_name>
> make migrate-up # do this after creation of migration files and updating sql 
```