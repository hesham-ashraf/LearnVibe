# LearnVibe - Modern Learning Platform

LearnVibe is a comprehensive learning platform built with a focus on clean architecture, security, and scalability. The platform enables course creation, student enrollment, and content delivery through a microservices-based backend.

## Project Structure

```
learnvibe/
├── backend/             # Backend services (Go)
│   ├── gateway/         # API Gateway service
│   ├── cms/             # Course Management service
│   ├── content-delivery/ # Content Delivery service
│   └── tests/           # Cross-service tests
└── Documents/           # Project documentation
```

## Key Features

- **Learning Management System**: Course creation, student enrollment, progress tracking
- **Content Delivery**: Secure delivery of educational materials (videos, documents, etc.)
- **User Authentication**: JWT and OAuth2 (Google) authentication
- **API Security**: Rate limiting, HTTPS, JWT tokens
- **Scalable Architecture**: Microservices with containerization

## Technical Implementation

The backend is built using modern web development practices:

- **Clean Architecture**: Separation of concerns with distinct layers
- **Containerization**: Docker + Docker Compose for consistent environments
- **API Gateway**: Service orchestration and unified entry point
- **Caching**: Redis for high-performance data access
- **Message Brokers**: RabbitMQ for asynchronous communication
- **Centralized Logging**: OpenSearch for log aggregation and analysis
- **Storage**: MinIO for S3-compatible object storage
- **Testing**: Unit, integration, contract, and load testing

## Getting Started

### Prerequisites

- Docker and Docker Compose
- Go 1.20+ (for local development)
- PostgreSQL (for local development)
- Google OAuth credentials (for authentication)

### Quick Start

1. Clone the repository:
```bash
git clone https://github.com/yourusername/LearnVibe.git
cd LearnVibe
```

2. Configure environment variables:
```bash
cd backend
cp .env.example .env
# Edit .env file with your settings
```

3. Start the services:
```bash
docker-compose up -d
```

4. Access the services:
- API Gateway: http://localhost:8000
- CMS Service: http://localhost:8080
- Content Delivery Service: http://localhost:8082
- OpenSearch Dashboards: http://localhost:5601
- RabbitMQ Management: http://localhost:15672
- MinIO Console: http://localhost:9001

## Development

See the backend [README](backend/README.md) for detailed development instructions.

## Testing

The project includes comprehensive testing:

- Unit tests
- Integration tests
- Contract tests for service compatibility
- Load & stress testing with k6

To run tests:
```bash
cd backend
./run-tests.bat  # Windows
./scripts/run-tests.sh  # Linux/macOS
```

## Security Features

- OAuth2 authentication with Google
- JWT token-based authentication
- Role-based access control
- Rate limiting
- Input validation
- HTTPS support

## Requirements Implemented

- ✅ Clean Architecture
- ✅ Docker and Compose containerization
- ✅ CRUD functionality with API Gateway
- ✅ Redis caching
- ✅ RabbitMQ message broker
- ✅ OpenSearch centralized logging
- ✅ OAuth2 authentication
- ✅ Secure API implementation
- ✅ Comprehensive testing suite
- ✅ 80%+ test coverage
