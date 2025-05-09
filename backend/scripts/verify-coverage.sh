#!/bin/bash
# Script to verify test coverage meets the 80% requirement

# Set minimum required coverage percentage
MIN_COVERAGE=80

# Array of services to check
SERVICES=("cms" "gateway" "content-delivery")

# Create coverage directory if it doesn't exist
COVERAGE_DIR="$(pwd)/coverage"
mkdir -p $COVERAGE_DIR

# Function to run tests and check coverage
check_coverage() {
    local service=$1
    echo "Checking coverage for $service service..."
    
    # Change to service directory
    cd "../$service" || { echo "Service directory $service not found"; return 1; }
    
    # Run tests with coverage
    go test -coverprofile="$COVERAGE_DIR/$service.out" ./... -run "Test[^Integration]"
    
    # Check if coverage file was created
    if [ ! -f "$COVERAGE_DIR/$service.out" ]; then
        echo "Error: Coverage file for $service was not created"
        return 1
    fi
    
    # Generate HTML report for visual inspection
    go tool cover -html="$COVERAGE_DIR/$service.out" -o "$COVERAGE_DIR/$service.html"
    
    # Get coverage percentage
    local coverage_output=$(go tool cover -func="$COVERAGE_DIR/$service.out" | grep "total:" | awk '{print $3}')
    local coverage=${coverage_output%\%}
    
    echo "Coverage for $service: $coverage%"
    
    # Compare with minimum required coverage
    if (( $(echo "$coverage < $MIN_COVERAGE" | bc -l) )); then
        echo "WARNING: Coverage for $service is below the required $MIN_COVERAGE%"
        return 1
    else
        echo "SUCCESS: Coverage for $service meets or exceeds the required $MIN_COVERAGE%"
        return 0
    fi
}

# Main execution
echo "=== Verifying Test Coverage ==="
echo "Minimum required coverage: $MIN_COVERAGE%"

# Start from the scripts directory
cd "$(dirname "$0")"

failed_services=()

# Check coverage for each service
for service in "${SERVICES[@]}"; do
    if ! check_coverage "$service"; then
        failed_services+=("$service")
    fi
    # Return to scripts directory
    cd "$(dirname "$0")"
    echo ""
done

# Print final results
echo "=== Coverage Verification Results ==="
if [ ${#failed_services[@]} -eq 0 ]; then
    echo "All services meet the minimum coverage requirement of $MIN_COVERAGE%"
    exit 0
else
    echo "The following services do not meet the minimum coverage requirement:"
    for service in "${failed_services[@]}"; do
        echo "- $service"
    done
    echo "Please add more tests to improve coverage"
    exit 1
fi 