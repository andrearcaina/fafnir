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

Check Permission Query

```graphql
query HasPermission($request: HasPermissionRequest!) {
  checkPermission(request: $request) {
    hasPermission
    permissionCode
  }
}
```

Get Stock Quote Batch Query

```graphql
query {
  getStockQuoteBatch(symbols: ["AAPL", "MSFT", "TSLA", "AMZN", "GOOGL", "META"]) {
    code
    data {
      symbol
      price
      open
      previousClose
    }
  }
}

```


## Example Mutations

None yet

## Example 200 Responses

Successful Health Check Response

```json
{
  "data": {
    "health": "API Gateway is running"
  }
}


```

Successful Check Permission Response

```json
{
  "data": {
    "checkPermission": {
      "hasPermission": false,
      "permissionCode": "PERMISSION_DENIED"
    }
  }
}
```

Successful Get Stock Quote Batch Response

```json
{
  "data": {
    "getStockQuoteBatch": {
      "code": 200,
      "data": [
        {
          "symbol": "AAPL",
          "price": 268.47,
          "open": 269.795,
          "previousClose": 269.77
        },
        {
          "symbol": "MSFT",
          "price": 496.82,
          "open": 496.68,
          "previousClose": 497.1
        },
        {
          "symbol": "TSLA",
          "price": 429.52,
          "open": 437.89,
          "previousClose": 445.91
        },
        {
          "symbol": "AMZN",
          "price": 244.41,
          "open": 241.15,
          "previousClose": 243.04
        },
        {
          "symbol": "GOOGL",
          "price": 278.83,
          "open": 283.205,
          "previousClose": 284.75
        },
        {
          "symbol": "META",
          "price": 621.71,
          "open": 616.485,
          "previousClose": 618.94
        }
      ]
    }
  }
}
```

Example Error Response

```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "details": "Authentication token not found in cookies",
    "message": "Authentication required"
  }
}
```