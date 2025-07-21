# GraphQL API Guide

GraphQL is the primary way to interact with the API Gateway in this project.

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

Login Mutation

### Login Mutation
```graphql
mutation {
  login(input: { user: "username", password: "password" }) {
    code
    message
    error
  }
}
```

If you only want the code, you can modify your mutation request to:

```graphql
mutation {
  login(input: { user: "username", password: "password" }) {
    code
  }
}
```

You can even use variables to make your mutation more dynamic:

```graphql
mutation Login($input: LoginInput!) {
  login(input: $input) {
    code
    message
    error
  }
}
```

Then, using variables, you can send the request with this in the body:
```json
{
  "input": {
    "user": "username",
    "password": "password"
  }
}
``` 


## Example Responses

### Successful Health Check Response
```json
{
  "data": {
    "health": "API Gateway is running"
  }
}
```

### Successful Login Response
```json
{
  "data": {
    "login": {
      "code": 200,
      "message": "Login successful",
      "error": ""
    }
  }
}
```

### Error Response
```json
{
  "data": {
    "login": {
      "code": 401,
      "message": "",
      "error": "Invalid credentials"
    }
  }
}
```