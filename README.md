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
|--------------------------------------|-------------------------------------|
| [Architecture](docs/architecture.md) | Project structure, service overview |
| [Development](docs/development.md)   | Setup, local dev, make commands     |
| [Database](docs/database.md)         | Migrations, Goose, DB details       |
| [GraphQL](docs/graphql.md)           | API schema, queries, mutations      |
| [Designs](docs/designs)              | Excalidraw designs and images       |

## TODO
In no particular order:
- [ ] Design and implement additional microservices.
- [X] Integrate NATS (Pub/Sub pattern) for asynchronous events and messaging ([issue #11](https://github.com/andrearcaina/fafnir/issues/11)).
  - [ ] Implement more events across services than just user creation.
- [X] Create system architecture, network and data diagrams (upkeep as much as possible).
- [ ] Build a simulation/orchestration engine for app events.
- [ ] Perform load testing using Locust to simulate concurrent users ([issue #8](https://github.com/andrearcaina/fafnir/issues/8)).
- [X] Explore Kubernetes local implementation ([issue #5](https://github.com/andrearcaina/fafnir/issues/5)).
- [X] Explore centralized logging CLI with Elasticsearch for Minikube ([issue #6](https://github.com/andrearcaina/fafnir/issues/6)).
- [ ] Add unit and integration tests for each microservice.
- [ ] Implement a CI/CD pipeline for automated testing and Docker builds.
