# Integration Tests for CMS Service

This directory contains integration tests for the CMS service of LearnVibe.

## Purpose

Integration tests verify the interaction between different components of the CMS service, including:
- API endpoints
- Database interactions
- Authentication flows
- Service integrations

## Running Tests

### Prerequisites

1. PostgreSQL test database running
2. Environment variables set correctly:
   ```bash
   # Linux/macOS
   export DB_HOST=localhost
   export DB_PORT=5433
   export DB_USER=postgres
   export DB_PASSWORD=postgres
   export DB_NAME=learnvibe_test
   
   # Windows PowerShell
   $env:DB_HOST="localhost"
   $env:DB_PORT="5433"
   $env:DB_USER="postgres"
   $env:DB_PASSWORD="postgres"
   $env:DB_NAME="learnvibe_test"
   ```

### Execute Tests

```bash
# From the cms directory
go test -v ./tests/integration
```

## Test Categories

- `auth_flow_test.go`: Tests the complete authentication flow from registration to profile access
- Other test files will be added as new features are implemented

## Troubleshooting

If tests fail due to database connection issues:
1. Ensure the PostgreSQL instance is running
2. Verify the environment variables are set correctly
3. Check if the database has been initialized with the test schema 