# LearnVibe API Gateway

The API Gateway serves as the single entry point for the LearnVibe platform, routing requests to the appropriate microservices. It sits between the client applications and the backend services, providing a unified interface.

## Features

- **Request Routing**: Routes requests to the appropriate microservice based on the path.
- **Authentication**: Validates JWT tokens before forwarding requests to services.
- **Rate Limiting**: Prevents abuse by limiting the number of requests from a single IP.
- **Circuit Breaking**: Prevents cascading failures using the circuit breaker pattern.
- **Health Checks**: Monitors the health of backend services.
- **Load Balancing**: (Future enhancement) Distributes traffic across service instances.

## Architecture

The API Gateway connects to the following services:

1. **CMS Service** (Port 8080): Handles course management, user authentication, and enrollment.
2. **Content Delivery Service** (Port 8082): Manages content storage and retrieval.

## Getting Started

### Prerequisites

- Go 1.23 or later
- Running instances of the CMS Service and Content Delivery Service

### Installation

1. Clone the repository
2. Navigate to the API Gateway directory:
   ```
   cd backend/gateway
   ```
3. Install dependencies:
   ```
   go mod download
   ```

### Configuration

Create a `.env` file in the `backend/gateway` directory with the following settings:

```
PORT=8000
JWT_SECRET=your-secret-key
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_DURATION=60
CIRCUIT_BREAKER_MAX_REQUESTS=5
CIRCUIT_BREAKER_INTERVAL=30
CIRCUIT_BREAKER_TIMEOUT=10
CMS_SERVICE_URL=http://localhost:8080
CONTENT_SERVICE_URL=http://localhost:8082
CMS_HEALTH_ENDPOINT=/health
CONTENT_HEALTH_ENDPOINT=/health
REQUEST_TIMEOUT=30
```

### Running the Gateway

```
go run main.go
```

The API Gateway will start on port 8000 (or the port specified in your configuration).

## Routing

The gateway routes requests to microservices based on the path:

- `/auth/*`: CMS Service (Authentication endpoints)
- `/api/courses/*`: CMS Service (Course management)
- `/api/enrollments/*`: CMS Service (Enrollment management)
- `/api/admin/*`: CMS Service (Admin functionality)
- `/api/content/*`: Content Delivery Service (Content management)
- `/public/content/*`: Content Delivery Service (Public content access)

## Health Check

The gateway exposes a health check endpoint at `/health` that returns the status of all services.

## Error Handling

If a service is unavailable, the gateway will respond with appropriate error messages:

- 502 Bad Gateway: When the target service is not responding
- 503 Service Unavailable: When the service is temporarily unavailable (circuit open)

## Architecture Diagram

```
┌─────────────┐          ┌──────────────┐         ┌───────────────┐
│  Client App │ -------> │ API Gateway  │ ------> │  CMS Service  │
└─────────────┘          │  (Port 8000) │         │  (Port 8080)  │
                         │              │         └───────────────┘
                         │              │
                         │              │         ┌───────────────┐
                         │              │ ------> │Content Service│
                         └──────────────┘         │  (Port 8082)  │
                                                  └───────────────┘
``` 