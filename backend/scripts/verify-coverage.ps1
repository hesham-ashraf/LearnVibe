# PowerShell script to verify test coverage meets the 80% requirement

# Set minimum required coverage percentage
$MIN_COVERAGE = 80

# Array of services to check
$SERVICES = @("cms", "gateway", "content-delivery")

# Create coverage directory if it doesn't exist
$PROJECT_ROOT = (git rev-parse --show-toplevel)
$BACKEND_DIR = Join-Path -Path $PROJECT_ROOT -ChildPath "backend"
$COVERAGE_DIR = Join-Path -Path $BACKEND_DIR -ChildPath "coverage"

if (-not (Test-Path -Path $COVERAGE_DIR)) {
    New-Item -Path $COVERAGE_DIR -ItemType Directory -Force | Out-Null
}

# Function to run tests and check coverage
function Check-Coverage {
    param (
        [string]$service
    )
    
    Write-Host "Checking coverage for $service service..." -ForegroundColor Yellow
    
    # Change to service directory
    Push-Location -Path (Join-Path -Path $BACKEND_DIR -ChildPath $service)
    
    # Run tests with coverage
    $coverageFile = Join-Path -Path $COVERAGE_DIR -ChildPath "$service.out"
    go test -coverprofile="$coverageFile" ./... -run "Test[^Integration]"
    
    # Check if coverage file was created
    if (-not (Test-Path -Path $coverageFile)) {
        Write-Host "Error: Coverage file for $service was not created" -ForegroundColor Red
        Pop-Location
        return $false
    }
    
    # Generate HTML report for visual inspection
    $htmlFile = Join-Path -Path $COVERAGE_DIR -ChildPath "$service.html"
    go tool cover -html="$coverageFile" -o "$htmlFile"
    
    # Get coverage percentage
    $coverageOutput = go tool cover -func="$coverageFile" | Select-String "total:"
    
    if ($coverageOutput) {
        $coverageParts = ($coverageOutput -Split "\s+")
        $coverageString = $coverageParts[-1]
        $coverage = [decimal]($coverageString -replace "%", "")
        
        Write-Host "Coverage for $service: $coverage%" -ForegroundColor Cyan
        
        # Compare with minimum required coverage
        if ($coverage -lt $MIN_COVERAGE) {
            Write-Host "WARNING: Coverage for $service is below the required $MIN_COVERAGE%" -ForegroundColor Red
            Pop-Location
            return $false
        } else {
            Write-Host "SUCCESS: Coverage for $service meets or exceeds the required $MIN_COVERAGE%" -ForegroundColor Green
            Pop-Location
            return $true
        }
    } else {
        Write-Host "Error: Could not determine coverage for $service" -ForegroundColor Red
        Pop-Location
        return $false
    }
}

# Main execution
Write-Host "=== Verifying Test Coverage ===" -ForegroundColor Yellow
Write-Host "Minimum required coverage: $MIN_COVERAGE%" -ForegroundColor Yellow

$failedServices = @()

# Check coverage for each service
foreach ($service in $SERVICES) {
    if (-not (Check-Coverage -service $service)) {
        $failedServices += $service
    }
    Write-Host ""
}

# Print final results
Write-Host "=== Coverage Verification Results ===" -ForegroundColor Yellow
if ($failedServices.Count -eq 0) {
    Write-Host "All services meet the minimum coverage requirement of $MIN_COVERAGE%" -ForegroundColor Green
    exit 0
} else {
    Write-Host "The following services do not meet the minimum coverage requirement:" -ForegroundColor Red
    foreach ($service in $failedServices) {
        Write-Host "- $service" -ForegroundColor Red
    }
    Write-Host "Please add more tests to improve coverage" -ForegroundColor Yellow
    exit 1
} 