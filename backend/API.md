# DayZ Server Management API

This document describes the REST API endpoints for managing DayZ servers.

## Base URL
```
http://localhost:8080/api
```

## Authentication
All protected endpoints require a JWT token in the Authorization header:
```
Authorization: Bearer <token>
```

## Public Endpoints

### Health Check
- **GET** `/health`
- **Response**: `{"status": "ok"}`

### Setup Admin User
- **POST** `/auth/setup`
- **Request Body**:
```json
{
  "username": "admin",
  "password": "password"
}
```
- **Response**:
```json
{
  "token": "jwt_token_here",
  "message": "Admin user created successfully"
}
```

### Login
- **POST** `/auth/login`
- **Request Body**:
```json
{
  "username": "admin",
  "password": "password"
}
```
- **Response**:
```json
{
  "token": "jwt_token_here",
  "message": "Login successful"
}
```

## Protected Endpoints (Require JWT)

### User Management

#### Get Current User
- **GET** `/auth/me`
- **Headers**: `Authorization: Bearer <token>`
- **Response**:
```json
{
  "id": 1,
  "username": "admin",
  "created_at": "2024-01-01T00:00:00Z"
}
```

### Server CRUD Operations

#### Create Server
- **POST** `/servers`
- **Headers**: `Authorization: Bearer <token>`
- **Request Body**:
```json
{
  "name": "MyServer",
  "port": 2302
}
```
- **Response**:
```json
{
  "status": "success",
  "data": {
    "id": 1,
    "name": "MyServer",
    "port": 2302,
    "status": "stopped"
  },
  "message": "Server created successfully"
}
```

#### List All Servers
- **GET** `/servers`
- **Headers**: `Authorization: Bearer <token>`
- **Response**:
```json
{
  "status": "success",
  "data": [
    {
      "id": 1,
      "name": "MyServer",
      "port": 2302,
      "status": "stopped",
      "config_path": null,
      "created_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

#### Get Server Details
- **GET** `/servers/{id}`
- **Headers**: `Authorization: Bearer <token>`
- **Response**:
```json
{
  "status": "success",
  "data": {
    "id": 1,
    "name": "MyServer",
    "port": 2302,
    "status": "stopped",
    "config_path": null,
    "created_at": "2024-01-01T00:00:00Z",
    "ftp_user": {
      "id": 1,
      "username": "dayz_MyServer",
      "home_dir": "/srv/dayz-servers/MyServer/files"
    }
  }
}
```

#### Delete Server
- **DELETE** `/servers/{id}`
- **Headers**: `Authorization: Bearer <token>`
- **Response**:
```json
{
  "status": "success",
  "message": "Server deleted successfully"
}
```

### Server Process Management

#### Start Server
- **POST** `/servers/{id}/start`
- **Headers**: `Authorization: Bearer <token>`
- **Response**:
```json
{
  "status": "success",
  "data": {
    "status": "running",
    "pid": 12345,
    "message": "Server started successfully"
  }
}
```

#### Stop Server
- **POST** `/servers/{id}/stop`
- **Headers**: `Authorization: Bearer <token>`
- **Response**:
```json
{
  "status": "success",
  "message": "Server stopped successfully"
}
```

#### Restart Server
- **POST** `/servers/{id}/restart`
- **Headers**: `Authorization: Bearer <token>`
- **Response**:
```json
{
  "status": "success",
  "message": "Server restarted successfully"
}
```

#### Get Server Status
- **GET** `/servers/{id}/status`
- **Headers**: `Authorization: Bearer <token>`
- **Response**:
```json
{
  "status": "success",
  "data": {
    "status": "running",
    "pid": 12345,
    "uptime": 3600,
    "message": ""
  }
}
```

### Server Logging

#### Get Server Logs
- **GET** `/servers/{id}/logs?lines=100&offset=0`
- **Headers**: `Authorization: Bearer <token>`
- **Query Parameters**:
  - `lines` (optional): Number of lines to return (default: 100, max: 1000)
  - `offset` (optional): Number of lines to skip (default: 0)
- **Response**:
```json
{
  "status": "success",
  "data": {
    "logs": "Log content here...",
    "message": "Retrieved 100 lines"
  }
}
```

#### Get Admin Logs
- **GET** `/servers/{id}/logs/admin`
- **Headers**: `Authorization: Bearer <token>`
- **Response**:
```json
{
  "status": "success",
  "data": {
    "logs": "Admin log content here...",
    "message": "Admin logs retrieved"
  }
}
```

### Server Configuration

#### Get Server Config
- **GET** `/servers/{id}/config`
- **Headers**: `Authorization: Bearer <token>`
- **Response**:
```json
{
  "status": "success",
  "data": {
    "config": "Server configuration content..."
  }
}
```

#### Update Server Config
- **PUT** `/servers/{id}/config`
- **Headers**: `Authorization: Bearer <token>`
- **Request Body**:
```json
{
  "config": "Server configuration content..."
}
```
- **Response**:
```json
{
  "status": "success",
  "message": "Configuration updated successfully"
}
```

### FTP User Management

#### Create FTP User
- **POST** `/servers/{id}/ftp/create`
- **Headers**: `Authorization: Bearer <token>`
- **Request Body**:
```json
{
  "username": "custom_user"
}
```
- **Response**:
```json
{
  "status": "success",
  "data": {
    "id": 1,
    "username": "custom_user",
    "password": "generated_password",
    "home_dir": "/srv/dayz-servers/MyServer/files",
    "message": "FTP user created successfully"
  }
}
```

#### Get FTP Credentials
- **GET** `/servers/{id}/ftp/credentials`
- **Headers**: `Authorization: Bearer <token>`
- **Response**:
```json
{
  "status": "success",
  "data": {
    "username": "dayz_MyServer",
    "host": "localhost",
    "port": 21,
    "home_dir": "/srv/dayz-servers/MyServer/files"
  }
}
```

#### Regenerate FTP Password
- **POST** `/servers/{id}/ftp/regenerate-password`
- **Headers**: `Authorization: Bearer <token>`
- **Response**:
```json
{
  "status": "success",
  "data": {
    "password": "new_generated_password",
    "message": "Password regenerated successfully"
  }
}
```

#### Delete FTP User
- **DELETE** `/servers/{id}/ftp/user`
- **Headers**: `Authorization: Bearer <token>`
- **Response**:
```json
{
  "status": "success",
  "message": "FTP user deleted successfully"
}
```

## Standard Response Format

All endpoints follow a standardized response format:

### Success Response
```json
{
  "status": "success",
  "data": {}, // Response data (optional)
  "message": "Operation completed successfully"
}
```

### Error Response
```json
{
  "status": "error",
  "error": "Error description"
}
```

## HTTP Status Codes

- `200 OK` - Request successful
- `201 Created` - Resource created successfully
- `400 Bad Request` - Invalid request data
- `401 Unauthorized` - Missing or invalid authentication
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource already exists
- `500 Internal Server Error` - Server error

## Error Handling

The API provides detailed error messages for common scenarios:

- **Validation Errors**: Invalid input data
- **Authentication Errors**: Missing or invalid JWT token
- **Database Errors**: Database operation failures
- **File System Errors**: Missing server files or configuration
- **Process Errors**: Server startup/shutdown failures

## Logging

All API requests are logged with:
- HTTP method and path
- User ID (for authenticated requests)
- Response status code
- Request duration

Errors are logged with full context for debugging.