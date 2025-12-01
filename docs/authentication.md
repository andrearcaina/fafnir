# Authentication Guide

Authentication is handled through a combination of JWT (JSON Web Tokens) and OAuth2. This guide provides an overview of how authentication (and authorization) works in this project.

## Overview
The auth-service is responsible for managing user authentication and authorization. It issues JWT tokens that are used to authenticate requests to other services, and it also supports OAuth2 for third-party authentication.

For authorization, the auth-service checks user roles and permissions to ensure that users have access to the requested resources.

## Authentication Flow
1. **User Registration**: Users can register by providing their email and password. The auth-service will create a new user in the database and return a success response.
2. **User Login**: Users can log in by providing their email and password. The auth-service will validate the credentials and, if valid, issue a JWT token that is sent back to the client.
3. **Token Storage**: The JWT token is stored in an HttpOnly cookie to prevent XSS attacks. The cookie is sent with every request to the API Gateway.
4. **Token Validation**: The API Gateway validates the JWT token on each request. If the token is valid, the request is forwarded to the appropriate service. If the token is invalid or expired, the API Gateway returns a 401 Unauthorized response.
5. **Logout**: Users can log out, which will invalidate the JWT token and remove it from the HttpOnly cookie.
6. **Account Deletion**: Users can delete their accounts, which will remove their data from all associated databases and invalidate any active JWT tokens.

## Potential Future Enhancements
- **OAuth2 Support**: Implement OAuth2 authentication with providers like Google, Facebook, etc.
- **Password Reset**: Implement a password reset mechanism where users can request a password reset link.

## Endpoints

- `POST /auth/register`: Register a new user on auth db, and send an event `user.registered` to NATS. See [architecture](architecture.md) for more details.
- `POST /auth/login`: Log in a user and receive two HttpOnly Cookies: One with the JWT token and one with CSRF token.
- `DELETE /auth/logout`: Log out a user by invalidating the JWT token.
- `DELETE /auth/delete`: Delete the authenticated user's account.
- `GET /auth/me`: Get the authenticated user's information based on the JWT token.
