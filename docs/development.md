# Development Guide

## Prerequisites before Starting
- **Go 1.24.5+** - For microservices development
- For local orchestration, there's two main ways to do it:
    - **Docker & Docker Compose** - For containerized development
    - **Minikube, kubectl & helm** - For local cluster and container orchestration (requires docker though)
- **GNU Make** - For DevOps automation scripts

## Setup Development Environment

1. **Clone the repository**
    ```bash
    git clone git@github.com:andrearcaina/fafnir.git
    cd fafnir
    ```

2. **Configure environment**
    ```bash
    cp infra/env/.env.dev.example infra/env/.env.dev
    # Edit .env.dev with your configuration
    ```

3. **Start development environment**
    - Using Docker Compose
      ```bash
      make docker-run       # Start all services
      make migrate-up       # Run database migrations
      make seed db=all      # Seed databases with test data
      ```

    - Using Kubernetes
      ```bash
      make docker-build              # Build docker images
      make kube-start                # Start Minikube cluster
      make kube-deploy               # Deploy services to Minikube
      make kube-forward pod=ps       # Port forward Postgres service
      
      # Open a new terminal
      make migrate-up                # Run database migrations
      make kube-tunnel               # Create tunnel to access services via load balancer
      ```
4. **Access Services**

   Grafana and Prometheus will be available if you started the monitoring stack (`make docker-run monitoring=true`).

   | Service         | URL                                            | Description                        |
   |-----------------|------------------------------------------------|------------------------------------|
   | **API Gateway** | [http://localhost:8080](http://localhost:8080) | Main entrypoint (GraphQL and REST) |
   | **Grafana**     | [http://localhost:3000](http://localhost:3000) | Monitoring dashboard               |
   | **Prometheus**  | [http://localhost:9090/metrics](http://localhost:9090/metrics) | Metrics collection                 |

## Development Automation Guide

A lot of these `make` commands are wrappers around bash scripts located in the `tools/scripts/` directory. 
I created a `Taskfile` for cross-platform compatibility, but `Makefile` was the first choice I used.
Most of these commands can be run just by using `make <command>`, or `task <command>` if you are using Task.
The difference is that with Task the command names are colon-separated instead of hyphen-separated.

These commands help you manage the development environment using Docker:

| Command                  | Description                                                                                   |
|--------------------------|-----------------------------------------------------------------------------------------------|
| `make docker-build`      | Build all docker containers                                                                   |
| `make docker-build-prod`  | Build all docker containers for  production                                                    |
| `make docker-start`      | Start all existing docker containers                                                          |
| `make docker-prod`       | Creates and run docker containers for production                                              |
| `make docker-pause`      | Stops running all existing docker containers                                                  |
| `make docker-run`        | Creates and run docker containers (`make run monitoring=true` to run with grafana/prometheus/loki) |
| `make docker-stop`       | Stops and deletes containers and volumes                                                      |
| `make docker-status`     | Check status of currently running docker containers                                           |
| `make docker-logs`       | Get logs of Docker services DB                                                                |
| `make docker-nats`       | Go into NATS container with natsio/natsbox  DB                                                |
| `make docker-rm-volumes` | Remove all volumes of Postgres DB                                                             |
| `make docker-prune`      | Prune all images and cached builds                                                            |
| `make docker-clean`      | Runs commands `docker-stop`, `docker-prune`, `docker-rm-volumes`                              |
| `make docker-reset`      | Runs commands `docker-clean`, `docker-build`, `docker-start`                                  |

These commands help you manage the development environment with Kubernetes (Minikube):

| Command                 | Description                                     |
|-------------------------|-------------------------------------------------|
| `make kube-start`       | Start Minikube cluster with configurations      |
| `make kube-stop`        | Stop Minikube cluster                           |
| `make kube-delete`      | Delete Minikube cluster                         |
| `make kube-uninstall`   | Uninstall Minikube cluster                      |
| `make kube-secrets`     | Create/Update secrets for Minikube cluster      |
| `make kube-docker`      | Load all docker images into Minikube cluster    |
| `make kube-deploy`      | Install services to Minikube cluster            |
| `make kube-upgrade`     | Upgrade all services in Minikube cluster        |
| `make kube-delete-pod pod=<pod_name>` | Delete a specific pod in Minikube cluster |
| `make kube-reset`       | Reset services in Minikube cluster              |
| `make kube-status`      | Check status of deployed services in Minikube   |
| `make kube-nodes`       | List all nodes in Minikube cluster              |
| `make kube-pods`        | List all pods in Minikube cluster               |
| `make kube-svc`         | List all services in Minikube cluster           |
| `make kube-deployments` | List all deployments in Minikube cluster        |
| `make kube-logs pod=<pod_name>` | View logs of a specific pod in Minikube cluster |
| `make kube-dashboard`   | Open Minikube dashboard in browser              |
| `make kube-ssh`         | SSH into Minikube cluster                       |
| `make kube-forward pod=<pod_name>` | Port forward a service from Minikube cluster    |
| `make kube-tunnel`      | Create a tunnel to access Minikube services     |

You can also use the following commands to migrate the database:

| Command                                                   | Description                   |
|-----------------------------------------------------------|-------------------------------|
| `make migrate-up`                                         | Run DB migrations             |
| `make migrate-down`                                       | Remove all DB migrations      |
| `make migrate-status`                                     | Check status of DB migrations |
| `make migrate-create db=<db_name> name=<migration_name> ` | Create a migration sql file   |

You can run the following commands to generate the GraphQL resolvers based on the schema:

| Command         | Description                                                                                          |
|-----------------|------------------------------------------------------------------------------------------------------|
| `make generate` | Generate GraphQL, SQLc, or proto boilerplate dependent on the `.graphqls`, `.sql`, or `.proto` files |

You can run the following commands to seed the database with initial data after migrations:

| Command     | Description                                           |
|-------------|-------------------------------------------------------|
| `make seed` | Seed database with initial data with `db=<target_db>` |

You can also run certain microservices individually:

| Command                        | Description                   |
|--------------------------------|-------------------------------|
| `make docker-auth-service`     | Start the auth service        |
| `make docker-user-service`     | Start the user service        |
| `make docker-security-service` | Start the security service    |
| `make docker-stock-service`    | Start the stock service       |
| `make docker-api-gateway`      | Start the GraphQL API Gateway |

You can also run locust for testing concurrent user load:

| Command                                                                                           | Description                                   |
|---------------------------------------------------------------------------------------------------|-----------------------------------------------|
| `make locust users=<total_users> spawn_rate=<spawn_rate> run_time=<run_time> headless=<headless>` | Run the locust CLI with custom configurations |

Commands for linting and vet:

| Command | Description |
|---------|-------------|
| `make lint` | Run linter on all services |
| `make vet` | Run vet on all services |

For more information on the commands, check out the `scripts/` folder.

| Bash Script                  | Description                     |
|------------------------------|---------------------------------|
| `./tools/scripts/codegen.sh` | All the codegen command logic   |
| `./tools/scripts/docker.sh`  | All the docker command logic    |
| `./tools/scripts/k8s.sh`     | All the kube command logic      |
| `./tools/scripts/migrate.sh` | All the migration command logic |
| `./tools/scripts/test.sh`    | All the test command logic      |

## Useful Links
- [Architecture Overview](architecture.md)
- [Database Guide](database.md)
- [GraphQL Guide](graphql.md)
- [Schema Design](schema.md)
