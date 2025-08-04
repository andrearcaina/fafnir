# GraphQL API Guide

GraphQL is the primary way to interact with the microservices in this project. It provides a flexible and efficient way to query and manipulate data across the various services.

If you want to learn how authentication is handled in this project, please refer to the [Authentication Guide](./authentication.md).

Some key points about the GraphQL API:
- **Single Endpoint**: All requests go to `/graphql`.
- **Schema-Driven**: The API is defined by a GraphQL schema, which describes the types, queries, and mutations available.
- **Queries and Mutations**: 
  - **Queries** are used to fetch data.
  - **Mutations** are used to modify data.

## Example Queries

Health Check Query

```graphql
{
    health
}
```

## Example Mutations

None yet


## Example Responses

### Successful Health Check Response
```json
{
  "data": {
    "health": "API Gateway is running"
  }
}
```