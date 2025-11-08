# fafnir


This repository was created as an educational project to learn more about modern microservice architecture, system design patterns, and best practices for building a scalable distributed application. It is my best attempt at implementing a well-structured microservice codebase.


## Documentation

| Guide                                | Description                         |
|--------------------------------------|-------------------------------------|
| [Development](docs/development.md)   | Setup, local dev, make commands     |
| [Database](docs/database.md)         | Migrations, Goose, DB details       |
| [Architecture](docs/architecture.md) | Project structure, service overview |
| [GraphQL](docs/graphql.md)           | API schema, queries, mutations      |

## TODO
In no particular order:
- [ ] Build and design the rest of the microservices needed
- [ ] Add NATS and Pub/Sub pattern for async events/messaging (implement saga pattern)
- [ ] Sketch a basic design of data flow as well as other services (orchestrate events that happen in the app)
- [ ] Add orchestrating/simulation engine or key app events (saga pattern)
- [ ] Simulate load testing with multiple users and services interacting using Locust (simulate real world usage and concurrent users)
- [X] Figure out how kubernetes fits in (check issue [kubernetes local implementation](https://github.com/andrearcaina/fafnir/issues/5))
- [ ] Add unit and integration tests for each microservice (going to be a lot of work)
- [ ] Add CI/CD pipeline for automated testing and docker image builds
