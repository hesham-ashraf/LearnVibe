# LearnVibe - Learning Management System

LearnVibe is a microservices-based learning management system designed for scalability and modern educational needs.

## System Architecture

The system consists of three main services:

1. **CMS Service** - Manages courses, users, and enrollments
2. **Content Delivery Service** - Handles content storage and delivery
3. **API Gateway** - Routes requests to appropriate services and handles authentication

## Prerequisites

- Go 1.23 or later
- PostgreSQL 12 or later
- Redis for caching (Content Delivery Service)
- MinIO for object storage (Content Delivery Service)

## Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/hesham-ashraf/LearnVibe.git
cd LearnVibe
```

### 2. Database Setup

Create two PostgreSQL databases:
- `learnvibe` - For the CMS service
- `learnvibe_content` - For the Content Delivery service

```bash
# Example using psql
createdb learnvibe
createdb learnvibe_content
```

### 3. Environment Configuration

Each service requires its own environment configuration. Sample `.env` files are provided in each service directory.

#### CMS Service

Create/update `backend/cms/.env`:

```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=learnvibe
JWT_SECRET=your_jwt_secret
PORT=8080
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
GOOGLE_CALLBACK_URL=http://localhost:8080/auth/google/callback
```

#### Content Delivery Service

Create/update `backend/content-delivery/.env`:

```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=learnvibe_content
JWT_SECRET=your_jwt_secret
PORT=8090
STORAGE_PATH=./storage
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET=learnvibe-content
MINIO_USE_SSL=false
```

#### API Gateway Service

Create/update `backend/gateway/.env`:

```
PORT=8000
JWT_SECRET=your_jwt_secret
RATE_LIMIT=100
CIRCUIT_TIMEOUT=30s
CIRCUIT_MAX_REQUESTS=10
CIRCUIT_INTERVAL=60s
CMS_SERVICE_URL=http://localhost:8080
CONTENT_SERVICE_URL=http://localhost:8090
REQUEST_TIMEOUT=30s
```

### 4. Starting the Services

You need to start all three services for the complete system to function.

#### Terminal 1: Start CMS Service

```bash
cd backend/cms
go run main.go
```

#### Terminal 2: Start Content Delivery Service

```bash
cd backend/content-delivery
go run main.go
```

#### Terminal 3: Start API Gateway

```bash
cd backend/gateway
go run main.go
```

## API Testing with Postman

1. Import the Postman collection: `LearnVibe API.postman_collection.json`
2. Set up environment variables in Postman:
   - `gateway_url`: `http://localhost:8000`
   - `cms_url`: `http://localhost:8080`
   - `content_url`: `http://localhost:8090`
   - `jwt_token`: (This will be populated after authentication)

3. Run the Authentication API first to get a JWT token:
   - Send a request to `GET {{gateway_url}}/auth/google`
   - Complete the OAuth flow in the browser
   - Copy the JWT token and set it in the Postman environment

4. Test other APIs as needed:
   - Courses API
   - Content API
   - Enrollments API

## Troubleshooting

### Database Connection Issues

If you encounter database connection issues:
- Verify PostgreSQL is running
- Check database credentials in the `.env` files
- Ensure databases are created

### Redis Connection Issues

For Redis-related errors:
- Ensure Redis server is running
- If using a password-protected Redis instance, update the `REDIS_PASSWORD` environment variable

### MinIO/Storage Issues

For content storage issues:
- Verify MinIO is running
- Check the MinIO credentials in the `.env` file
- Ensure the bucket exists

## Development Guidelines

- Follow Go best practices
- Use proper error handling
- Write unit tests for new functionality
- Document APIs and code changes

## Built With

- [Go](https://golang.org/) - The programming language
- [Gin](https://github.com/gin-gonic/gin) - Web framework
- [GORM](https://gorm.io/) - ORM library
- [JWT](https://github.com/golang-jwt/jwt) - Authentication mechanism
- [PostgreSQL](https://www.postgresql.org/) - Database
- [Redis](https://redis.io/) - Caching
- [MinIO](https://min.io/) - Object storage

## License

This project is licensed under the MIT License.
