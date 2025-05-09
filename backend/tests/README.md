# LearnVibe Testing Suite

This directory contains the global testing infrastructure for the LearnVibe platform, focusing on end-to-end testing, load testing, and other cross-service tests.

## Directory Structure

- `k6/` - Load and stress testing scripts using k6
- Other tests specific to individual services are located in their respective directories

## Testing Strategy

LearnVibe follows a comprehensive testing strategy including:

1. **Unit Tests**: Testing individual components in isolation
2. **Integration Tests**: Testing interaction between components
3. **Contract Tests**: Ensuring compatibility between services
4. **End-to-End Tests**: Testing complete workflows across services
5. **Load Tests**: Evaluating performance under expected load
6. **Stress Tests**: Evaluating system behavior under extreme conditions
7. **Spike Tests**: Evaluating system behavior under sudden traffic spikes

## Load Testing with k6

[k6](https://k6.io/) is used for load, stress, and spike testing. The `k6/` directory contains test scripts for different scenarios.

### Running Load Tests

```bash
# Run basic load test
k6 run k6/load-test.js

# Run stress test
k6 run --tag scenario=stress k6/load-test.js

# Run spike test
k6 run --tag scenario=spike k6/load-test.js

# Run with custom VUs and duration
k6 run --vus 50 --duration 30s k6/load-test.js
```

### Load Test Thresholds

The load tests include the following thresholds:

- 95% of requests must complete in less than 500ms
- 99% of requests must complete in less than 1s
- Less than 1% error rate

## Contract Testing

Contract testing ensures that the interactions between services adhere to their defined contracts.

Contract tests are located in each service's `tests/contract` directory:
- `cms/tests/contract/`
- `content-delivery/tests/contract/`

## Continuous Integration

All tests are run as part of the CI pipeline in GitHub Actions. See `.github/workflows/test.yml` for details.

## Test Coverage

We aim for at least 80% test coverage across all services. Coverage reports are generated during test runs and available in the `coverage/` directory.

To verify coverage meets requirements:

```bash
# Linux/macOS
./scripts/verify-coverage.sh

# Windows
.\scripts\verify-coverage.ps1
```

## Adding New Tests

When adding new tests, follow these guidelines:

1. **Unit Tests**: Place in the same package as the code being tested
2. **Integration Tests**: Place in `<service>/tests/integration/`
3. **Contract Tests**: Place in `<service>/tests/contract/`
4. **Load Tests**: Place in `tests/k6/`

Each test file should have a clear naming convention:
- Unit tests: `<package>_test.go`
- Integration tests: `<feature>_integration_test.go`
- Contract tests: `<consumer>_<provider>_contract_test.go`
- Load tests: `<feature>_load_test.js` 