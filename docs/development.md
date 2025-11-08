# Development Guide

## Prerequisites
- **Go 1.21+** - For microservices development
- **Docker & Docker Compose** - For containerized development
- **Make** - For build automation (use WSL2 on Windows)

## Quick Setup

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd fafnir
   ```

2. **Configure environment**
   ```bash
   cp infra/env/.env.dev.example infra/env/.env.dev
   # Edit .env.dev with your configuration
   ```

3. **Start development environment**
   ```bash
   make run              # Start all services
   make migrate-up       # Run database migrations
   make seed serviceDB=all      # Seed databases with test data
   ```

## Service Access Points

| Service          | URL                                            | Description          |
|------------------|------------------------------------------------|----------------------|
| **API Gateway**  | [http://localhost:8080](http://localhost:8080) | GraphQL endpoint     |
| **Auth Service** | [http://localhost:8081](http://localhost:8081) | Authentication API   |
| **Grafana**      | [http://localhost:3000](http://localhost:3000) | Monitoring dashboard |
| **Prometheus**   | [http://localhost:9090](http://localhost:9090) | Metrics collection   |

## Development Workflow

### Starting Services
```bash
# Start all core services
make run

# Start with monitoring stack
make run-monitoring

# Start individual services
make run-auth-service
make run-user-service
make run-security-service
make run-stock-service
make run-api-gateway
```

### Database Operations
```bash
# Run migrations
make migrate-up

# Check migration status
make migrate-status
```

### Seed Operations
```bash
# Populate databases
make seed db=all        # All databases
make seed db=auth       # Auth database only
make seed db=user       # User database only
make seed db=security   # Security database only
make seed db=stock      # Stock database only
```


### Code Generation
```bash
# Generate GraphQL resolvers
make generate codegen=graphql

# Generate SQLC code for specific service
make generate codegen=sqlc service=auth
```

## Make Commands

These commands help you manage the development environment using Docker:

| Command                  | Description                                                                                   |
|--------------------------|-----------------------------------------------------------------------------------------------|
| `make docker-build`      | Build all docker containers                                                                   |
| `make docker-start`      | Start all existing docker containers                                                          |
| `make docker-pause`      | Stops running all existing docker containers                                                  |
| `make docker-run`        | Creates and run docker containers (`make run monitoring=true` to run with grafana/prometheus) |
| `make docker-stop`       | Stops and deletes containers and volumes                                                      |
| `make docker-status`     | Check status of currently running docker containers                                           |
| `make docker-rm-volumes` | Remove all volumes of PostgreSQL DB                                                           |
| `make docker-prune`      | Prune all images and cached builds                                                            |
| `make docker-clean`      | Runs commands `docker-stop`, `docker-prune`, `docker-rm-volumes`                              |
| `make docker-reset`      | Runs commands `docker-clean`, `docker-build`, `docker-start`                                  |

You can also use the following commands to migrate the database:

| Command                                                          | Description                   |
|------------------------------------------------------------------|-------------------------------|
| `make migrate-up`                                                | Run DB migrations             |
| `make migrate-down`                                              | Remove all DB migrations      |
| `make migrate-status`                                            | Check status of DB migrations |
| `make migrate-create serviceDB=<db_name> name=<migration_name> ` | Create a migration sql file   |

You can run the following commands to generate the GraphQL resolvers based on the schema:

| Command         | Description                                                                                    |
|-----------------|------------------------------------------------------------------------------------------------|
| `make generate` | Generate GraphQL, sqlc, or proto boilerplate dependent on the .graphqls, .sql, or .proto files |

You can run the following commands to seed the database with initial data after migrations:

| Command     | Description                                           |
|-------------|-------------------------------------------------------|
| `make seed` | Seed database with initial data with `db=<target_db>` |

You can also run certain microservices individually:

| Command                     | Description                   |
|-----------------------------|-------------------------------|
| `make run-auth-service`     | Start the auth service        |
| `make run-user-service`     | Start the user service        |
| `make run-security-service` | Start the security service    |
| `make stock-service`        | Start the stock service       |
| `make run-api-gateway`      | Start the GraphQL API Gateway |

For more information on the commands, check out the scripts folder.

| Bash Script                  | Description                        |
|------------------------------|------------------------------------|
| `./tools/scripts/docker.sh`  | All the docker command logic       |
| `./tools/scripts/gqlgen.sh`  | All the gqlgen command logic       |
| `./tools/scripts/help.sh`    | The help command logic             |
| `./tools/scripts/migrate.sh` | The goose migrations command logic |
| `./tools/scripts/seed.sh`    | Seed command logic                 |


## Useful Links
- [Architecture Overview](architecture.md)
- [Database Guide](database.md)
- [GraphQL Guide](graphql.md)
- [Schema Design](schema.md)