# Project Architecture

## Overview

Fafnir is a modern, scalable microservices platform built with Go, featuring centralized tooling, multi-stage Docker builds, and comprehensive observability. The architecture follows microservices best practices with proper service isolation and security.

### Key Architectural Principles
- **API Gateway Pattern**: Single entry point for all external requests
- **Service Isolation**: Internal services not exposed to public network
- **Centralized Tooling**: Shared build tools, seeder, and scripts
- **Multi-Language Ready**: Structured to support Go, Java, C#, and Python services
- **Container-First**: Docker-native development and deployment
- **Observability**: Built-in monitoring with Prometheus and Grafana

## Technology Stack
- **Backend**: Go (as of right now) microservices with gRPC/REST communication
- **API Gateway**: GraphQL unified endpoint using gqlgen
- **Database**: PostgreSQL with per-service databases
- **Frontend**: Next.js with TypeScript and ShadCN/UI
- **Containerization**: Docker with multi-stage builds
- **Monitoring**: Prometheus, Grafana
- **Development**: Hot reload, centralized scripts, Make-based workflow

## Project Structure

```
fafnir/
├── build/                   # Build configurations
│   └── docker/              # Centralized Dockerfiles
│   └── ci/                  # CI configurations
├── deployments/             # Deployment configurations
│   └── compose/             # Docker Compose files
├── docs/                    # Documentation
├── frontend/                # Next.js web application
├── infra/                   # Infrastructure configurations
│   ├── env/                 # Environment files
│   ├── monitoring/          # Prometheus & Grafana configs
│   └── postgres/            # Database initialization
├── services/                # Microservices
│   ├── api-gateway/         # GraphQL API Gateway
│   ├── auth-service/        # Authentication service
│   ├── security-service/    # Authorization service
│   ├── user-service/        # User management service
│   └── shared/              # Shared libraries and utilities
└── tools/                   # Development tools
    ├── scripts/             # Build and deployment scripts
    └── seeder/              # Centralized database seeder
```

## Service Architecture

### Core Services

| Service              | Description                                                      | Tech Stack          | Ports           | Database    |
|----------------------|------------------------------------------------------------------|---------------------|-----------------|-------------|
| **api-gateway**      | GraphQL API Gateway - Single entry point for all client requests | Go, gqlgen, go-chi  | 8080 (public)   | -           |
| **auth-service**     | Authentication & JWT token management with OAuth 2.0 support     | Go, sqlc, go-chi    | 8081 (internal) | auth_db     |
| **user-service**     | User profile management and CRUD operations                      | Go, sqlc, go-chi    | 8083 (internal) | user_db     |
| **security-service** | Role-based access control and authorization                      | Go, sqlc, go-chi    | 8082 (internal) | security_db |
| **frontend**         | Next.js web application with TypeScript and ShadCN/UI            | Next.js, TypeScript | 3001 (public)   | -           |

### Infrastructure Services

| Service           | Description                                    | Ports           | Purpose              |
|-------------------|------------------------------------------------|-----------------|----------------------|
| **postgres**      | PostgreSQL database with per-service databases | 5432 (internal) | Data persistence     |
| **prometheus**    | Metrics collection and monitoring              | 9090 (dev only) | Observability        |
| **grafana**       | Metrics visualization and dashboards           | 3000 (dev only) | Monitoring UI        |

## Network Architecture

### Security Model
- **Public Access**: Only API Gateway (8080), Auth Service (8081), and Frontend (3001) are exposed
- **Internal Communication**: All microservices communicate via Docker internal network
- **Database Access**: Services connect to individual databases via internal network
- **Monitoring**: Prometheus scrapes metrics from internal service endpoints

## Development Workflow

### Centralized Tooling
- **Seeder**: Single tool in `tools/seeder/` seeds all service databases
- **Scripts**: Centralized in `tools/scripts/` for consistency
- **Dockerfiles**: Multi-stage templates in `build/docker/` for all services
- **Compose**: Modular files in `deployments/compose/` for different environments

### Build Process
1. **Development**: Hot reload with volume mounts
2. **Testing**: Isolated test stage in multi-stage Dockerfiles
3. **Production**: Optimized, minimal images with non-root users

### Database Management
- **Migrations**: Per-service migrations using Goose
- **Seeding**: Centralized seeder supports all services via CLI flags
- **Initialization**: Database creation handled by PostgreSQL init scripts

## Future Extensibility

The current structure is designed to easily accommodate:
- **Multi-Language Services**: Ready for Java, C#, Python services
- **Additional Databases**: Redis, MongoDB can be added easily
- **Cloud Deployment**: Kubernetes manifests can be added
- **CI/CD Pipelines**: GitHub Actions workflows ready to implement

## Data Flow
Below is the ideal data flow for the application. A concept drawing will be added later. For authentication data flow, check the [Authentication Guide](./authentication.md).
1. Frontend → asks API Gateway for data
2. API Gateway → routes request to appropriate service
3. Services → interacts with their own PostgreSQL database
4. Services → processes data and may call other services if needed
5. Services → returns data to API Gateway
6. API Gateway → aggregates data from multiple services if necessary
7. API Gateway → sends data back to Frontend
8. Frontend → displays data to the user
