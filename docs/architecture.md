# Project Architecture

### Architecture Designs
- [Initial Design #1](designs/images/dev_design_1.png)
- [Revised Design of #1](designs/images/revised_dev_design_1.png)
- [Kubernetes Node & Network Design](designs/images/k8s_node_network_design_1.png)

### Key Architectural Principles
- **Microservices**: Each service has its own database and is independently deployable
- **Service Isolation**: Internal services not exposed to public network (only API Gateway is public + Reverse Proxy to auth)
- **Centralized Tooling**: Shared build tools, seeder, and scripts
- **Scalable Services**: Structured to add more services if needed (in any language too)
- **Container-First**: Docker-native development and deployment
- **Observability**: Built-in monitoring with Prometheus and Grafana

### Design Patterns Utilized
- **Backend-for-Frontend (BFF)**: Designed for tailored client experiences (this means the API Gateway is optimized for a singular frontend)
- **API Gateway Pattern**: Single entry point for all client requests, routing to appropriate services
- **Database per Service**: Each microservice manages its own database schema and data

### Technology Stack
- **Backend**: Go microservices with gRPC/REST communication
- **API Gateway**: GraphQL unified endpoint using gqlgen
- **Message Broker**: NATS for event based communication
- **Database**: PostgreSQL with per-service databases
- **Cache**: Redis Cache used for core services in need of fast response times
- **Containerization**: Docker with multi-stage builds
- **Monitoring**: Prometheus, Grafana
- **Development**: Hot reload (with air), centralized scripts, Make-based workflow

### Project Structure

```
fafnir/
├── build/                   # Build configurations
│   └── docker/              # Centralized Dockerfiles
│   └── ci/                  # CI configurations (if needed)
├── deployments/             # Deployment configurations
│   ├── compose/             # Docker Compose files
│   └── k8s/                 # Kubernetes Manifests
├── docs/                    # Documentation
│   └── designs/             # Excalidraw designs and images
├── infra/                   # Infrastructure configurations
│   ├── env/                 # Environment files
│   ├── monitoring/          # Prometheus & Grafana configs
│   └── postgres/            # Database initialization
├── services/                # Microservices
│   ├── api-gateway/         # GraphQL API Gateway
│   ├── auth-service/        # Authentication service
│   ├── security-service/    # Authorization service
│   ├── user-service/        # User management service
│   ├── stock-service/       # Stock service
│   └── shared/              # Shared libraries and utilities
└── tools/                   # Development tools
    ├── logctl/              # Centralized logging control tool
    ├── scripts/             # Build and deployment scripts
    └── seedctl/             # Centralized database seeder
    
```

### Core Services

| Service              | Description                                                      | Tech Stack         | Ports           | Database        |
|----------------------|------------------------------------------------------------------|--------------------|-----------------|-----------------|
| **api-gateway**      | GraphQL API Gateway - Single entry point for all client requests | Go, gqlgen, go-chi | 8080 (public)   | -               |
| **auth-service**     | Authentication & JWT token management                            | Go, sqlc, go-chi   | 8081 (internal) | auth_db         |
| **security-service** | Role-based access control and authorization                      | Go, sqlc, gRPC     | 8082 (internal) | security_db     |
| **user-service**     | User profile management and CRUD operations                      | Go, sqlc, gRPC     | 8083 (internal) | user_db         |
| **stock-service**    | Stock quote and metadata information                             | Go, sqlc, go-chi   | 8084 (internal) | stock_db, redis |

### Infrastructure Services

| Service        | Description                                  | Ports           | Purpose          |
|----------------|----------------------------------------------|-----------------|------------------|
| **postgres**   | Postgres database with per-service databases | 5432 (internal) | Data persistence |
| **redis**      | Redis caching for quick look up              | 6379 (internal) | Caching          |
| **prometheus** | Metrics collection and monitoring            | 9090 (dev only) | Observability    |
| **grafana**    | Metrics visualization and dashboards         | 3000 (dev only) | Monitoring UI    |

### Data Flow
Below is the ideal data flow for the application. It will be updated when NATS is implemented.
1. Client → asks API Gateway for data
2. API Gateway → routes request to appropriate service
3. Services → interacts with their own Postgres database
4. Services → processes data and may call other services if needed
5. Services → returns data to API Gateway
6. API Gateway → aggregates data from multiple services if necessary
7. API Gateway → sends data back to Client

### Helpful Resources and Readings
- [Microservices](https://martinfowler.com/articles/microservices.html) by [Martin Fowler](https://martinfowler.com/)
- [What is Microservices Architecture?](https://webandcrafts.com/blog/what-is-microservices-architecture) by [Anjaly Chandran](https://webandcrafts.com/author/anjaly-chandran)
- [A pattern language for microservices](https://microservices.io/patterns/) by [Chris Richardson](https://microservices.io/about.html)
- [19 Microservices Patterns for System Design Interviews](https://dev.to/somadevtoo/19-microservices-patterns-for-system-design-interviews-3o39) by [Soma](https://dev.to/somadevtoo)
- [A Crash Course on Microservices Design Patterns](https://blog.bytebytego.com/p/a-crash-course-on-microservices-design) by [ByteByteGo](https://blog.bytebytego.com/about)
- [NATS Documentation](https://docs.nats.io/) by [NATS](https://nats.io/about/)
- [Kubernetes Documentation](https://kubernetes.io/docs/concepts/architecture/) by [Kubernetes](https://kubernetes.io/)