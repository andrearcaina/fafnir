# fáfnir

> The name "Fáfnir" is inspired from Norse mythology, and refers to a dwarf who transformed into a mythical Germanic dragon, so he can guard his treasure hoard of gold and such.

This is a purely educational project that serves as a hands-on demonstration of building a modern codebase
that explores microservices architecture, asynchronous event-driven design,
and best practices for creating a scalable, distributed application.
I do not intend to use this project for other purposes and is mostly
a playground for my learning, experimentation, and exploration.

Now for the application:
It is designed to function as the backend for a stock trading simulator/platform, including services for user management, authentication, stock data retrieval, buy/sell operations, and security/permissions.

## Documentation

For more detailed information, please refer to the documentation in the `docs/` directory, or visit the following links:

| Guide                                | Description                         |
| ------------------------------------ | ----------------------------------- |
| [Architecture](docs/architecture.md) | Project structure, service overview |
| [Development](docs/development.md)   | Setup, local dev, make commands     |
| [Database](docs/database.md)         | Migrations, Goose, DB details       |
| [GraphQL](docs/graphql.md)           | API schema, queries, mutations      |
| [Designs](docs/designs)              | Excalidraw designs and images       |

## TODO

In no particular order:

- [x] Design and implement additional microservices ([issue #15](https://github.com/andrearcaina/fafnir/issues/15)).
- [x] Integrate NATS for asynchronous events and messaging ([issue #11](https://github.com/andrearcaina/fafnir/issues/11)).
    - [x] Swapped from Pub/Sub to Worker Queues for message persistence with NATS JetStream ([issue #20](https://github.com/andrearcaina/fafnir/issues/20)).
    - [x] Implement more events across services than just user creation.
- [x] Create system architecture, network and data diagrams (upkeep as much as possible).
- [x] Build a simulation engine for orchestrating trading events ([PR #32](https://github.com/andrearcaina/fafnir/pull/32)).
- [ ] Add unit, integration, end-to-end and load/stress tests.
    - [x] Perform load testing using Locust to simulate concurrent users ([issue #8](https://github.com/andrearcaina/fafnir/issues/8)).
- [x] Explore Kubernetes local implementation ([issue #5](https://github.com/andrearcaina/fafnir/issues/5)).
    - [x] Use Helm for Kubernetes package management and manifests ([issue #29](https://github.com/andrearcaina/fafnir/issues/29)).
- [x] Explore centralizing logging with Elasticsearch ([issue #6](https://github.com/andrearcaina/fafnir/issues/6)).
    - [x] Migrate to utilizing [Loki](https://grafana.com/oss/loki/) for both Minikube and docker-compose ([issue #29](https://github.com/andrearcaina/fafnir/issues/29)).
- [ ] Implement a CI/CD pipeline for automated testing and Docker builds.
- [ ] Production deployment via DigitalOcean and Traefik ([issue #22](https://github.com/andrearcaina/fafnir/issues/22)).
