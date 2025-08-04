# Project Architecture

## Overview
- Utilizes Go (and possibly Java, C#, and Python) microservices to handle different responsibilities
- Links services via gRPC and/or REST
- Unified API Gateway using GraphQL
- PostgreSQL for data storage and management (later on can add Redis for caching)
- Next.js and TypeScript frontend for user interface and interaction

## Structure
- `docs/`: Documentation (architecture, development, database)
- `scripts/`: Scripts for development and deployment (mainly used to run docker and migration commands)
- `services/`: Go microservices source code (potentially might add more languages like Java, C#, Python)
- `shared/`: Shared Go code and protobufs (might remove Go code, and keep only protobufs)
- `infra/`: Infrastructure (docker, environment variables, postgres DB, k8s, monitoring)
- `frontend/`: Frontend web app (planning to use ShadCN/UI with Next.js)

## Microservice Responsibilities
| Service            | Description                                                                                | Tech Stack Used           | Reason of Choice                                                                                                                              |
|--------------------|--------------------------------------------------------------------------------------------|---------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------|
| `api-gateway`      | GraphQL API Entry point (routes client requests)                                           | Go, gqlgen, go-chi        | I wanted to try out GraphQL, and thought Go would be the best choice for it, especially since the other services I have implemented are in Go |
| `auth-service`     | The authorization and authentication server. Sends JWT HttpOnly cookies and uses OAuth 2.0 | Go, sqlc, go-chi          | Same reason as above, but just a simple RESTful API                                                                                           |
| `security-service` | The security service is to check if what roles and permissions the user has                | Go, sqlc, gRPC, protobufs | I wanted to use gRPC to learn about more about server intercommunication and why protobufs are the "best" at it                               |
| `user-service`     | The profile service where users can check their account information and details            | Go, sqlc, gRPC, protobufs | Same as security-service                                                                                                                      |

TODO:
- Implement more services like `notification-service`, `payment-service`, ...

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

## File Tree Structure
```plaintext
fafnir/
├── docs/
├── frontend/
├── infra/              
│    ├── env/              
│    ├── k8s/              
│    ├── monitoring/       
│    ├── postgres/
├── scripts/
│     ├── codegen.sh
│     ├── docker.sh
│     ├── help.sh
│     ├── migration.sh 
├── services/
│    ├── api-gateway/      
│    ├── auth-service/
│    ├── security-service/     
│    ├── user-service/     
├── shared/
├── .gitignore
├── Makefile          
├── README.md             
```