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
5. **OAuth2 Support**: The auth-service also supports OAuth2 for third-party authentication. Users can log in using their Google or other OAuth2 providers. The auth-service will handle the OAuth2 flow and issue a JWT token upon successful authentication.
6. **Logout**: Users can log out, which will invalidate the JWT token and remove it from the HttpOnly cookie.
7. **Password Reset**: Users can request a password reset, which will send a reset link to their email. The auth-service will handle the password reset flow and update the user's password in the database.

A concept drawing of the authentication flow will be added later.