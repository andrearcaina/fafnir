# Development Guide

## Prerequisites
- Go
- Make (If you're on Windows, use WSL)
- Node.js & npm
- Docker & Docker Compose

## Setup
1. Clone the repo.
2. Copy `.env.example` to `.env` and update values.
3. Install dependencies:
    - Backend: `go mod tidy`
    - Frontend: `npm install` in `web/`

## Make Commands

These commands help you manage the development environment using Docker:

| Command               | Description                                         |
|-----------------------|-----------------------------------------------------|
| `make build`          | Build all docker containers                         |
| `make run`            | Start all docker containers                         |
| `make status`         | Check status of currently running docker containers |
| `make stop`           | Stop every docker containers                        |
| `make rm-volumes`     | Remove all volumes of PostgreSQL DB                 |
| `make prune`          | Prune all images and cached builds                  |
| `make clean`          | Runs commands `stop`, `prune` `rm-volumes`          |
| `make reset`          | Runs commands `clean`, `build` `run`                |

You can also use the following commands to migrate the database:

| Command                                                   | Description                   |
|-----------------------------------------------------------|-------------------------------|
| `make migrate-up`                                         | Run DB migrations             |
| `make migrate-down`                                       | Remove all DB migrations      |
| `make migrate-status`                                     | Check status of DB migrations |
| `make migrate-create db=<db_name> name=<migration_name> ` | Create a migration sql file   |


You can also run certain microservices individually:

| Command                 | Description                   |
|-------------------------|-------------------------------|
| `make run-auth-service` | Start the backend service     |
| `make run-user-service` | Start the frontend service    |
| `make run-api-gateway`  | Start the GraphQL API Gateway |
| `make run-web-app`      | Start the web app             |


## Running Locally
- Start docker containers: `make run`
- Run migrations: `make migrate-up`
- Access Next.js App: [http://localhost:5000](http://localhost:5000)
- Access Grafana App: [http://localhost:3000](http://localhost:3000)
- Access GraphQL API: [http://localhost:8080](http://localhost:8080)

## Useful Links
- [Architecture Overview](architecture.md)
- [Database Guide](database.md)
- [GraphQL Guide](graphql.md)
- [Schema Design](schema.md)