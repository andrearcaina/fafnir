# Database & Migrations

## Overview
- Uses Postgres (see `build/docker/postgres.Dockerfile`)
- Migrations managed with [Goose](https://github.com/pressly/goose)
- Install Goose if you haven't already:
```bash
> go install github.com/pressly/goose/v3/cmd/goose@latest
```

## Running Migrations
- Each service has its own database and migration files located in `services/<service-name>/migrations/`
- To run migrations for a specific service, use:
```bash
> make migrate-create db=<service_name> name=<migration_name>
> make migrate-up # do this after creation of migration files and updating sql 
```