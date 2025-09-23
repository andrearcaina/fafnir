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
- [ ] Build and design basic microservices for app utilizing other languages (Java, Python, C#, Ruby)
- [ ] Add Pub/Sub for async events/messaging and decoupled microservices (implement saga pattern)
- [ ] Sketch a basic design of data flow as well as other services (orchestrate events that happen in the app)
- [ ] Add orchestrating/simulation engine or key app events (saga pattern)
- [ ] Figure out how kubernetes fits in (test locally with minikube first maybe)
- [ ] Add unit and integration tests for each microservice (might be a lot of work)
- [ ] Add CI/CD pipeline for automated testing and deployment
- [ ] Somehow deploy to the cloud (GCP, Render, VPS?)
