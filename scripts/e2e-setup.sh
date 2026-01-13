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
    # Fallback: check if setup endpoint exists, if not try to get existing key
    echo "Setup endpoint not available, checking for existing test configuration..."

    # In E2E mode, the API should have a default test key
    # Use the default test key for E2E mode
    DEFAULT_KEY="sk-at-e2e-test-key-do-not-use-in-production"
    echo "Using default E2E test API key"
    echo "AGENTTRACE_API_KEY=$DEFAULT_KEY" >> "$GITHUB_ENV" 2>/dev/null || true
    echo "$DEFAULT_KEY"
fi
