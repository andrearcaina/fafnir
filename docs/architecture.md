# Project Architecture

### Latest Architecture Designs
- [System High Level Design](designs/images/latest/system_hld.png)
- [Kubernetes Infrastructure Design](designs/images/latest/k8s_infra.png)
- [NATS JetStream Diagram](designs/images/latest/nats_js.png)

### Key Architectural Principles
- **Microservices**: Each service is independently deployable, scalable and follows the single responsibility principle (each service does one thing well)
- **Backend-for-Frontend (BFF)**: Designed for tailored client experiences (this means the API Gateway is optimized for a singular frontend)
- **API Gateway Pattern**: Single entry point for all client requests, routing to appropriate services
- **Database per Service**: Each microservice manages its own database schema and data
- **Event-Driven Architecture**: Asynchronous communication between services using NATS JetStream
- **Service Isolation**: Internal services not exposed to public network (only API Gateway is public + Reverse Proxy to auth)
- **Centralized Tooling**: Shared tests, tools, CLIs, and scripts for build and deployment automation

### Project Structure

```
fafnir/
├── build/                   # Build configurations
│   └── docker/              # Centralized Dockerfiles
├── deployments/             # Deployment configurations
│   ├── archive/             # Archive folder containing old Kubernetes manifests
│   │   └── k8s/             # Kubernetes Manifests
│   ├── docker/              # Docker Compose files
│   └── helm/                # Helm charts
├── docs/                    # Documentation
│   └── designs/             # Excalidraw designs and images
├── infra/                   # Infrastructure configurations
│   ├── db/                  # Database configurations (initialization scripts)
│   ├── env/                 # Environment variables
│   └── monitoring/          # Prometheus, Loki, and Grafana configurations
├── proto/                   # Protocol buffer definitions
├── src/                     # Source code for microservices
│   ├── api-gateway/         # GraphQL API Gateway
│   ├── auth-service/        # Authentication service
│   ├── security-service/    # Authorization service
│   ├── user-service/        # User management service
│   ├── stock-service/       # Stock service
│   ├── order-service/       # Order service
│   ├── portfolio-service/   # Portfolio service
│   └── shared/              # Shared libraries and utilities
├── tests/                   # Testing suites
│   ├── e2e/                 # End-to-end tests 
│   └── locust/              # Load testing with Locust
├── tools/                   # Development tools
│   ├── cli/                 # some dev CLIs
│   │   └── seedctl/         # Database seeder
│   └── scripts/             # Build and deployment scripts
├── .gitattributes
├── .gitignore
├── LICENSE
├── Makefile                 # Build automation
└── README.md                # Project overview and documentation
```

### Core Services

| Service              | Description                                                      | Tech Stack                   | Ports           | Database        |
|----------------------|------------------------------------------------------------------|------------------------------|-----------------|-----------------|
| **api-gateway**      | GraphQL API Gateway - Single entry point for all client requests | Go, gqlgen, go-chi, promhttp | 8080 (public)   | -               |
| **auth-service**     | Authentication & JWT token management                            | Go, sqlc, go-chi, promhttp   | 8081 (internal) | auth_db         |
| **security-service** | Role-based access control and authorization                      | Go, sqlc, gRPC, promhttp     | 8082 (internal) | security_db     |
| **user-service**     | User profile management and CRUD operations                      | Go, sqlc, gRPC, promhttp     | 8083 (internal) | user_db         |
| **stock-service**    | Stock quote and metadata information                             | Go, sqlc, gRPC, promhttp     | 8084 (internal) | stock_db, redis |

### Infrastructure Services

| Service           | Description                                       | Ports           | Purpose                    |
|-------------------|---------------------------------------------------|-----------------|----------------------------|
| **postgres**      | Postgres database with per-service databases      | 5432 (internal) | Data persistence           |
| **redis**         | Redis caching for quick look up                   | 6379 (internal) | Caching                    |
| **prometheus**    | Metrics collection and monitoring                 | 9090 (dev only) | Observability              |
| **loki**          | Unified logging storage                           | 3100 (dev only) | Observability              |
| **grafana**       | Dashboard and Observability UI                    | 3000 (dev only) | Observability              |
| **nats jetstream**| Persistent event streaming message broker         | 4222 (internal) | Event Streaming            |
| **locust**        | Load testing tool for simulating concurrent users | 8089 (dev only) | Load testing               |

### Communication Patterns
This architecture employs a combination of synchronous and asynchronous communication patterns:
- **Synchronous Communication**: The API Gateway handles client requests and routes them to the appropriate microservices using REST or gRPC.
- **Asynchronous Communication**: Microservices communicate with each other using NATS JetStream for event-driven interactions, allowing for decoupled and scalable service interactions.

It is both event driven and request-response based, depending on the use case and service requirements.

Feel free to take a look at the [designs](designs/images) folder for visual representations of architecture, network, and data flow designs.

### Helpful Resources and Readings
- [Microservices](https://martinfowler.com/articles/microservices.html) by [Martin Fowler](https://martinfowler.com/)
- [What is Microservices Architecture?](https://webandcrafts.com/blog/what-is-microservices-architecture) by [Anjaly Chandran](https://webandcrafts.com/author/anjaly-chandran)
- [A pattern language for microservices](https://microservices.io/patterns/) by [Chris Richardson](https://microservices.io/about.html)
- [19 Microservices Patterns for System Design Interviews](https://dev.to/somadevtoo/19-microservices-patterns-for-system-design-interviews-3o39) by [Soma](https://dev.to/somadevtoo)
- [A Crash Course on Microservices Design Patterns](https://blog.bytebytego.com/p/a-crash-course-on-microservices-design) by [ByteByteGo](https://blog.bytebytego.com/about)
- [NATS Documentation](https://docs.nats.io/) by [NATS](https://nats.io/about/)
- [Kubernetes Documentation](https://kubernetes.io/docs/concepts/architecture/) by [Kubernetes](https://kubernetes.io/)
- [Helm Documentation](https://helm.sh/docs/) by [Helm](https://helm.sh/)
- [Prometheus Documentation](https://prometheus.io/docs/introduction/overview/) by [Prometheus](https://prometheus.io/)
- [Loki Documentation](https://grafana.com/oss/loki/) by [Grafana](https://grafana.com/)
- [Grafana Documentation](https://grafana.com/docs/) by [Grafana](https://grafana.com/)