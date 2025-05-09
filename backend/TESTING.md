# LearnVibe Testing Guide

This document outlines the testing strategy for the LearnVibe project, covering unit tests, integration tests, load tests, and contract tests.

## Testing Requirements

- Unit tests for core business logic (80% coverage target)
- Integration tests for API endpoints, service interactions, and database queries
- Load & stress testing to evaluate system performance under high concurrency
- Contract testing to ensure service compatibility

## Setting Up the Test Environment

### Prerequisites

- Go 1.20 or higher
- PostgreSQL database
- Docker (optional, for containerized testing)
- k6 (for load/stress testing)
- Pact (for contract testing)

### Environment Setup

1. Create a test database:

```bash
# Linux/macOS
createdb learnvibe_test

# Windows (using PostgreSQL CLI)
createdb -U postgres learnvibe_test
```

2. Configure environment variables for testing:

```bash
# Linux/macOS
export TEST_DATABASE_URL=postgres://postgres:postgres@localhost:5432/learnvibe_test
export JWT_SECRET=test-secret

# Windows (PowerShell)
$env:TEST_DATABASE_URL="postgres://postgres:postgres@localhost:5432/learnvibe_test"
$env:JWT_SECRET="test-secret"
```

3. Install test dependencies:

```bash
go get github.com/stretchr/testify/assert
go get github.com/stretchr/testify/mock
go get github.com/pact-foundation/pact-go
```

## Running Tests

### Unit Tests

Unit tests focus on testing individual components in isolation:

```bash
# Linux/macOS
cd backend/cms
go test -v ./... -run "Test[^Integration]"

# Windows
cd backend\cms
go test -v ./... -run "Test[^Integration]"

# Windows (using provided script)
.\scripts\run-unit-tests.bat
```

To generate coverage reports:

```bash
# Linux/macOS
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Windows
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Integration Tests

Integration tests verify the interaction between different components:

```bash
# Start test database - Linux/macOS
docker-compose -f docker-compose.test.yml up -d postgres

# Start test database - Windows
docker-compose -f docker-compose.test.yml up -d postgres

# Run integration tests - Linux/macOS
cd backend/cms
go test -v ./tests/integration

# Run integration tests - Windows
cd backend\cms
go test -v ./tests/integration
```

### Contract Tests

Contract tests verify that the interactions between services adhere to their contracts:

1. Install Pact:

```bash
# Install Pact standalone - Linux/macOS
curl -LO https://github.com/pact-foundation/pact-ruby-standalone/releases/download/v1.88.40/pact-1.88.40-linux-x86_64.tar.gz
tar xzf pact-1.88.40-linux-x86_64.tar.gz
export PATH=$PATH:$PWD/pact/bin

# Install Pact standalone - Windows
# Download from https://github.com/pact-foundation/pact-ruby-standalone/releases
# Extract to a location and add to PATH
```

2. Run contract tests:

```bash
# Consumer tests
cd backend/cms
go test -v ./tests/contract

# Provider verification
cd backend/content-delivery
PACT_PROVIDER_VERIFICATION=true go test -v ./tests/contract

# Windows (PowerShell)
cd backend\content-delivery
$env:PACT_PROVIDER_VERIFICATION="true"
go test -v ./tests/contract
```

### Load Tests

Load tests evaluate system performance under expected load:

1. Install k6:

```bash
# Install k6 - Linux/macOS (Ubuntu)
sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6

# Install k6 - Windows
# Download the installer from https://k6.io/docs/getting-started/installation/
```

2. Run load tests:

```bash
# Run basic load test
k6 run backend/tests/k6/load-test.js

# Run with specific scenario
k6 run --tag scenario=load backend/tests/k6/load-test.js
```

### Stress Tests

Stress tests evaluate system behavior under extreme conditions:

```bash
# Run stress test scenario
k6 run --tag scenario=stress backend/tests/k6/load-test.js
```

### Spike Tests

Spike tests evaluate system behavior under sudden traffic spikes:

```bash
# Run spike test scenario
k6 run --tag scenario=spike backend/tests/k6/load-test.js
```

## Continuous Integration

Tests are automatically run in CI using GitHub Actions:

- Unit and integration tests run on every push and pull request
- Contract tests run when changes affect service interfaces
- Load tests run on release branches

## Testing Best Practices

1. **Isolation**: Ensure tests don't interfere with each other
2. **Predictability**: Tests should be deterministic and repeatable
3. **Speed**: Keep tests fast to encourage frequent testing
4. **Independence**: Avoid dependencies between tests
5. **Coverage**: Aim for at least 80% code coverage with meaningful assertions

## Mocking External Services

For tests that rely on external services:

1. Use the testify mock package for unit tests
2. Use Docker containers for integration tests
3. Use Pact for contract testing

## Test Data Management

1. Use fixtures for common test data
2. Reset the database state between integration test runs
3. Use unique identifiers for test data to prevent collisions

## Windows-Specific Testing Information

### Running Tests on Windows

For running tests on Windows, we provide dedicated scripts:

1. For basic unit tests:
```
.\scripts\run-unit-tests.bat
```

2. For all tests:
```
.\scripts\run-tests.bat
```

### Known Windows Issues

1. **Controller Tests**: Some controller tests may fail due to GORM mock implementation issues. These have been temporarily disabled.

2. **Database Authentication**: Integration tests require proper database credentials. Set environment variables:
   ```powershell
   $env:DB_HOST = "localhost"
   $env:DB_PORT = "5433"
   $env:DB_USER = "postgres"
   $env:DB_PASSWORD = "postgres"
   $env:DB_NAME = "learnvibe_test"
   ```

3. **PowerShell Command Chaining**: Use semicolons `;` instead of `&&` for command chaining, or run commands separately.

### Recent Test Fixes

1. **PowerShell Script Compatibility**
   - Created `run-tests.ps1` with proper PowerShell syntax
   - Fixed path handling to use Windows-style paths
   - Added proper error handling

2. **Database Interface Issues**
   - Fixed nil pointer dereference in auth controller
   - Implemented two-step database query approach
   - Improved error handling in database interactions

3. **Test Mocking Issues**
   - Implemented proper test DB interface for simulating database
   - Improved test implementation for user password hashing
   - Fixed role verification tests

## Troubleshooting

Common issues:

1. **Database connection issues**: Ensure PostgreSQL is running and credentials are correct
2. **Authentication failures**: Verify JWT_SECRET is set correctly
3. **Timeout errors**: Check for slow database queries or hanging HTTP connections
4. **Docker issues on Windows**: Ensure Docker Desktop is running with WSL 2 backend
5. **Path issues**: Ensure you're using the correct path format for your OS

## Additional Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Testify Documentation](https://github.com/stretchr/testify)
- [k6 Documentation](https://k6.io/docs/)
- [Pact Documentation](https://docs.pact.io/) 