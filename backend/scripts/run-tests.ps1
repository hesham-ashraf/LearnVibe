# PowerShell test runner script for LearnVibe project
# This script runs all types of tests and generates coverage reports

# Setup variables
$PROJECT_ROOT = (git rev-parse --show-toplevel)
$BACKEND_DIR = Join-Path -Path $PROJECT_ROOT -ChildPath "backend"
$COVERAGE_DIR = Join-Path -Path $BACKEND_DIR -ChildPath "coverage"
$SERVICES = @("cms", "gateway", "content-delivery")

# Create coverage directory if it doesn't exist
if (-not (Test-Path -Path $COVERAGE_DIR)) {
    New-Item -Path $COVERAGE_DIR -ItemType Directory -Force | Out-Null
}

Write-Host "=== LearnVibe Testing Framework ===" -ForegroundColor Yellow

# Function to run unit tests for a service
function Run-UnitTests {
    param (
        [string]$service
    )
    
    Write-Host "`nRunning unit tests for $service service..." -ForegroundColor Yellow
    
    # Go to the service directory
    Push-Location -Path (Join-Path -Path $BACKEND_DIR -ChildPath $service)
    
    # Run tests with coverage
    go test -v -coverprofile="$COVERAGE_DIR\$service.out" ./... -run "Test[^Integration]"
    $testResult = $LASTEXITCODE -eq 0
    
    # Generate HTML coverage report if tests passed
    if ($testResult) {
        go tool cover -html="$COVERAGE_DIR\$service.out" -o "$COVERAGE_DIR\$service.html"
        
        # Calculate coverage percentage
        $coverageOutput = go tool cover -func="$COVERAGE_DIR\$service.out" | Select-String "total:"
        if ($coverageOutput) {
            $coverage = ($coverageOutput -Split "\s+")[-1]
            Write-Host "Unit test coverage for $service`: $coverage" -ForegroundColor Green
            
            # Check if coverage is below threshold
            $coverageValue = [decimal]($coverage -replace "%", "")
            if ($coverageValue -lt 80) {
                Write-Host "Warning: Coverage below 80% threshold!" -ForegroundColor Red
            }
        }
    } else {
        Write-Host "Unit tests failed for $service" -ForegroundColor Red
    }
    
    Pop-Location
    return $testResult
}

# Function to run integration tests for a service
function Run-IntegrationTests {
    param (
        [string]$service
    )
    
    Write-Host "`nRunning integration tests for $service service..." -ForegroundColor Yellow
    
    # Check if integration tests exist
    $integrationPath = Join-Path -Path $BACKEND_DIR -ChildPath "$service\tests\integration"
    if (Test-Path -Path $integrationPath) {
        Push-Location -Path (Join-Path -Path $BACKEND_DIR -ChildPath $service)
        go test -v ./tests/integration
        $result = $LASTEXITCODE -eq 0
        Pop-Location
        return $result
    } else {
        Write-Host "No integration tests found for $service" -ForegroundColor Yellow
        return $true
    }
}

# Function to run contract tests
function Run-ContractTests {
    Write-Host "`nRunning contract tests..." -ForegroundColor Yellow
    
    # Check if contract tests exist
    $contractPath = Join-Path -Path $BACKEND_DIR -ChildPath "cms\tests\contract"
    if (Test-Path -Path $contractPath) {
        Push-Location -Path (Join-Path -Path $BACKEND_DIR -ChildPath "cms")
        go test -v ./tests/contract
        $result = $LASTEXITCODE -eq 0
        Pop-Location
        
        # Run provider verification if we have a content service
        $contentContractPath = Join-Path -Path $BACKEND_DIR -ChildPath "content-delivery\tests\contract"
        if (Test-Path -Path $contentContractPath) {
            Push-Location -Path (Join-Path -Path $BACKEND_DIR -ChildPath "content-delivery")
            $env:PACT_PROVIDER_VERIFICATION = "true"
            go test -v ./tests/contract
            $result = $result -and ($LASTEXITCODE -eq 0)
            Remove-Item Env:\PACT_PROVIDER_VERIFICATION
            Pop-Location
        }
        
        return $result
    } else {
        Write-Host "No contract tests found" -ForegroundColor Yellow
        return $true
    }
}

# Function to run k6 load tests
function Run-LoadTests {
    Write-Host "`nRunning load tests..." -ForegroundColor Yellow
    
    # Check if k6 is installed
    $k6Exists = Get-Command -Name k6 -ErrorAction SilentlyContinue
    if (-not $k6Exists) {
        Write-Host "k6 is not installed. Skipping load tests." -ForegroundColor Red
        return $true
    }
    
    # Check if load tests exist
    $loadTestPath = Join-Path -Path $BACKEND_DIR -ChildPath "tests\k6\load-test.js"
    if (Test-Path -Path $loadTestPath) {
        # Run with minimal VUs for CI environment
        $K6_OPTIONS = "--vus 5 --duration 10s"
        
        Write-Host "Running basic load test..." -ForegroundColor Yellow
        k6 run $K6_OPTIONS $loadTestPath
        return $LASTEXITCODE -eq 0
    } else {
        Write-Host "No k6 load tests found" -ForegroundColor Yellow
        return $true
    }
}

# Main execution
Write-Host "`nStarting database for tests..." -ForegroundColor Yellow
$dockerComposeTestPath = Join-Path -Path $BACKEND_DIR -ChildPath "docker-compose.test.yml"
if (Test-Path -Path $dockerComposeTestPath) {
    docker-compose -f $dockerComposeTestPath up -d postgres
} else {
    Write-Host "No docker-compose.test.yml file found" -ForegroundColor Red
}

# Run tests for each service
$errors = 0
foreach ($service in $SERVICES) {
    $servicePath = Join-Path -Path $BACKEND_DIR -ChildPath $service
    if (Test-Path -Path $servicePath) {
        # Run unit tests
        if (-not (Run-UnitTests -service $service)) {
            $errors++
        }
        
        # Run integration tests
        if (-not (Run-IntegrationTests -service $service)) {
            $errors++
        }
    }
}

# Run contract tests
if (-not (Run-ContractTests)) {
    $errors++
}

# Run load tests
if (-not (Run-LoadTests)) {
    $errors++
}

Write-Host "`nCleaning up..." -ForegroundColor Yellow
if (Test-Path -Path $dockerComposeTestPath) {
    docker-compose -f $dockerComposeTestPath down
}

# Final report
Write-Host "`n=== Test Results ===" -ForegroundColor Yellow
if ($errors -eq 0) {
    Write-Host "All tests passed successfully!" -ForegroundColor Green
    Write-Host "Coverage reports are available in: $COVERAGE_DIR"
    exit 0
} else {
    Write-Host "$errors test suites had errors." -ForegroundColor Red
    Write-Host "Coverage reports are available in: $COVERAGE_DIR"
    exit 1
} 