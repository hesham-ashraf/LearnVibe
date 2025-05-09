# Contract Tests for CMS Service

This directory contains contract tests for the CMS service using the Pact framework.

## Purpose

Contract tests verify that the interactions between the CMS service and other services adhere to defined contracts. These tests ensure compatibility between services as they evolve independently.

## Consumer vs Provider Tests

- **Consumer Tests**: Define expectations for how the CMS service will use other services
- **Provider Tests**: Verify that the CMS service meets the expectations of its consumers

## Running Tests

### Prerequisites

1. Install Pact CLI tools:
   ```bash
   # Linux/macOS
   curl -LO https://github.com/pact-foundation/pact-ruby-standalone/releases/download/v1.88.40/pact-1.88.40-linux-x86_64.tar.gz
   tar xzf pact-1.88.40-linux-x86_64.tar.gz
   export PATH=$PATH:$PWD/pact/bin
   
   # Windows
   # Download from https://github.com/pact-foundation/pact-ruby-standalone/releases
   # Extract and add to PATH
   ```

### Execute Tests

```bash
# Running consumer tests
go test -v ./tests/contract

# Running provider verification
cd ../content-delivery
PACT_PROVIDER_VERIFICATION=true go test -v ./tests/contract

# Windows PowerShell (provider verification)
cd ..\content-delivery
$env:PACT_PROVIDER_VERIFICATION="true"
go test -v ./tests/contract
```

## Contract File Management

Contract files (pacts) are stored in:
- Local development: `./pacts`
- CI environment: Uploaded to Pact Broker

## Adding New Contracts

To add a new contract test:

1. Create a new test file with the naming pattern `*_test.go`
2. Define the interactions using the Pact DSL
3. Run the tests to generate contract files
4. Ensure the provider verifies the contract 