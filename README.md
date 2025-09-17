# fafnir

This was built to demonstrate an understanding of microservices, GraphQL, GCP, Pub/Sub, gRPC, Docker, Kubernetes, Grafana, and Prometheus.

## Documentation

| Guide                                | Description                         |
|--------------------------------------|-------------------------------------|
| [Development](docs/development.md)   | Setup, local dev, make commands     |
| [Database](docs/database.md)         | Migrations, Goose, DB details       |
| [Architecture](docs/architecture.md) | Project structure, service overview |
| [GraphQL](docs/graphql.md)           | API schema, queries, mutations      |

## TODO
In no particular order:
- [ ] Build and design basic microservices for app utilizing other languages
- [ ] Add messaging queue with Pub/Sub for async events and decoupled microservices
- [ ] Sketch a basic design of data flow as well as other services (orchestrate events that happen in the app)
- [ ] Add orchestrating/simulation engine or key app events
- [ ] Figure out how kubernetes fits in (test locally with minikube first)
- [ ] Somehow deploy to the cloud
