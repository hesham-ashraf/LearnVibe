# LearnVibe Backend

This is the backend for the LearnVibe learning platform, following clean architecture principles and modern web development practices.

## Architecture

The backend is built as a set of microservices:

1. **API Gateway**: Routes requests to appropriate services, handles authentication and rate limiting
2. **CMS Service**: Manages courses, users, and enrollment
3. **Content Delivery Service**: Manages and delivers educational content

## Technologies

- **Programming Language**: Go
- **Web Framework**: Gin
- **Databases**:
  - PostgreSQL (primary data store)
  - Redis (caching)
- **Storage**: MinIO (S3-compatible object storage)
- **Message Broker**: RabbitMQ
- **Centralized Logging**: OpenSearch
- **Containerization**: Docker + Docker Compose
- **Authentication**: JWT + OAuth2 (Google)
- **API Security**: Rate limiting, HTTPS, JWT tokens

## Requirements Implemented

- ✅ Clean Architecture
- ✅ Containerization (Docker + Docker Compose)
- ✅ API Gateway
- ✅ CRUD functionality
- ✅ Caching using Redis
- ✅ Message Broker (RabbitMQ)
- ✅ Centralized Logging (OpenSearch)
- ✅ OAuth2 Authentication (Google)
- ✅ Secure API (rate limiting, JWT tokens)
- ✅ Integration Tests
- ✅ Unit Tests
- ✅ Load & Stress Testing (k6)
- ✅ Contract Testing

## Getting Started

### Prerequisites

- Docker and Docker Compose
- Go 1.20 or higher (for local development)
- PostgreSQL (for local development)
- Google OAuth credentials (for authentication)

### Running with Docker Compose

```bash
# Create .env file with required variables
cp .env.example .env

# Edit .env file to set your specific values
# especially the Google OAuth credentials

# Start all services
docker-compose up -d
```

The services will be available at:
- API Gateway: http://localhost:8000
- CMS Service: http://localhost:8080
- Content Delivery Service: http://localhost:8082
- OpenSearch Dashboards: http://localhost:5601
- RabbitMQ Management: http://localhost:15672
- MinIO Console: http://localhost:9001

### Local Development Setup

For local development without Docker:

```bash
# Install dependencies
go mod download

# Run PostgreSQL and Redis (recommended to use Docker for these)
docker-compose up -d postgres-cms postgres-content redis

# Run the CMS service
cd cms
go run main.go

# In another terminal, run the Content Delivery service
cd content-delivery
go run main.go

# In another terminal, run the API Gateway
cd gateway
go run main.go
```

## Testing

The project includes comprehensive testing, including unit tests, integration tests, contract tests, and load tests.

### Running Tests

```bash
# Run all tests
./run-tests.bat  # Windows
./scripts/run-tests.sh  # Linux/macOS

# Run unit tests only
cd backend
go test -v ./... -run "Test[^Integration]"

# Run integration tests
cd backend/cms
go test -v ./tests/integration

# Run contract tests
cd backend/cms
go test -v ./tests/contract

# Run load tests with k6
k6 run backend/tests/k6/load-test.js
```

## API Documentation

API documentation is available at each service's Swagger endpoint:

- API Gateway: http://localhost:8000/swagger/index.html
- CMS Service: http://localhost:8080/swagger/index.html
- Content Delivery Service: http://localhost:8082/swagger/index.html

## Clean Architecture Implementation

The project follows clean architecture principles:

- **Models**: Core business entities
- **Services**: Business logic and rules
- **Controllers**: Handle HTTP requests and responses
- **Repositories**: Data access layer

Each service is structured to maintain separation of concerns and dependency inversion.

## Security

Security features implemented:
- OAuth2 authentication with Google
- JWT token-based authentication
- Rate limiting at the API Gateway
- HTTPS support (configure in production)
- Input validation and sanitization 