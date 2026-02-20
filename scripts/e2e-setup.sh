#!/bin/bash
# E2E Test Environment Setup Script
# This script sets up the test environment for E2E tests

set -e

API_URL="${API_URL:-http://localhost:8080}"
MAX_RETRIES=60
RETRY_INTERVAL=2

echo "Waiting for API to be ready at $API_URL..."

for i in $(seq 1 $MAX_RETRIES); do
    if curl -s "$API_URL/health" | grep -q "ok"; then
        echo "API is ready!"
        break
    fi
    if [ $i -eq $MAX_RETRIES ]; then
        echo "ERROR: API failed to become ready after $MAX_RETRIES attempts"
        exit 1
    fi
    echo "Attempt $i/$MAX_RETRIES - waiting for API..."
    sleep $RETRY_INTERVAL
done

# Create test organization and project using internal setup endpoint
# This endpoint is only available in E2E_TEST_MODE
echo "Creating test organization and project..."
SETUP_RESPONSE=$(curl -s -X POST "$API_URL/api/internal/e2e-setup" \
    -H "Content-Type: application/json" \
    -d '{"orgName": "e2e-test-org", "projectName": "e2e-test-project"}')

if echo "$SETUP_RESPONSE" | grep -q "apiKey"; then
    API_KEY=$(echo "$SETUP_RESPONSE" | grep -o '"apiKey":"[^"]*"' | cut -d'"' -f4)
    echo "Test API key created successfully"
    echo "AGENTTRACE_API_KEY=$API_KEY" >> "$GITHUB_ENV" 2>/dev/null || true
    echo "$API_KEY"
else
    # Fallback: require AGENTTRACE_API_KEY to be set via environment
    echo "Setup endpoint not available, checking for existing test configuration..."

    if [ -n "${AGENTTRACE_API_KEY:-}" ]; then
        echo "Using AGENTTRACE_API_KEY from environment"
        echo "$AGENTTRACE_API_KEY"
    else
        echo "ERROR: E2E setup endpoint failed and AGENTTRACE_API_KEY is not set."
        echo "Please set AGENTTRACE_API_KEY in your CI environment variables or ensure the API is running in E2E_TEST_MODE."
        exit 1
    fi
fi
