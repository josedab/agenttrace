#!/usr/bin/env bash
set -euo pipefail

API_URL="${API_URL:-http://localhost:8080}"
WEB_URL="${WEB_URL:-http://localhost:3000}"

check_url() {
  local name="$1"
  local url="$2"
  local expected="$3"
  local response

  response="$(curl --fail --silent --show-error "$url")"
  if [[ "$response" != *"$expected"* ]]; then
    echo "$name returned an unexpected response: $response" >&2
    exit 1
  fi
  echo "ok - $name"
}

check_url "API health" "$API_URL/health" '"status":"healthy"'
check_url "API readiness" "$API_URL/readyz" '"status":"ready"'
check_url "API version" "$API_URL/version" '"version"'
check_url "Prometheus metrics" "$API_URL/metrics" "agenttrace_http"
check_url "Web health" "$WEB_URL/api/health" '"status":"ok"'
