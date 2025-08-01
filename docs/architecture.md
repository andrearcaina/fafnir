# Project Architecture

## Overview
- Utilizes Go microservices to handle different responsibilities
- Links services via gRPC and REST
- Unified API Gateway using GraphQL
- PostgreSQL for data storage and management (later on can add Redis for caching)
- Next.js frontend for user interface

## Structure
- `docs/`: Documentation (architecture, development, database)
- `scripts/`: Scripts for development and deployment (used by Makefile)
- `services/`: Go microservices (API Gateway, Auth, User)
- `shared/`: Shared Go code and protobufs
- `infra/`: Infrastructure (Docker, DB, k8s, monitoring)
- `frontend/`: Frontend web app

## Service Responsibilities
- **api-gateway**: Entry point, GraphQL API, routes requests
- **auth-service**: Auth logic, token management
- **user-service**: User CRUD

## Data Flow
1. Client → API Gateway (GraphQL)
2. API Gateway → Services via gRPC or REST
3. Services → interacts with Postgres DB
4. Frontend fetches data from API Gateway

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
│     ├── docker.sh
      ├── help.sh
      ├── migration.sh 
├── services/
│    ├── api-gateway/      
│    ├── auth-service/
│    ├── user-service/     
├── shared/
├── .gitignore
├── Makefile          
├── README.md             
```