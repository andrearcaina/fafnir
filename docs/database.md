# Database & Migrations

## Overview
- Uses Postgres (see `infra/db/Dockerfile`)
- Migrations managed with [Goose](https://github.com/pressly/goose)
- Install Goose if you haven't already:
```bash
> go install github.com/pressly/goose/v3/cmd/goose@latest
```

## Running Migrations
- Apply migrations: `make migrate-up`
- Migration files: `infra/db/migrations/`

## Creating a Migration
1. Generate a new migration:
```bash
> make migrate-create name=your_migration_name
```

2. Edit the generated file in `infra/db/migrations/`:
   - Add SQL commands in the `Up` function.
   - Add rollback commands in the `Down` function.

```sql
-- Example migration file
-- +goose Up
CREATE TABLE example_table (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL
);
    
-- +goose Down
DROP TABLE example_table;
```
   
3. Run the migration:
```bash
> make migrate-up
```