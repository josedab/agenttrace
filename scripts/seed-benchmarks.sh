#!/bin/bash
# Seed AgentTrace with curated benchmark datasets
# Run: ./scripts/seed-benchmarks.sh [API_URL] [API_KEY]

set -euo pipefail

API_URL="${1:-${AGENTTRACE_API_URL:-http://localhost:8080}}"
API_KEY="${2:-${AGENTTRACE_API_KEY:-}}"

if [ -z "$API_KEY" ]; then
    echo "Error: API key required. Pass as argument or set AGENTTRACE_API_KEY"
    echo "Usage: $0 [API_URL] [API_KEY]"
    exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
BENCHMARKS_FILE="$SCRIPT_DIR/../examples/benchmarks/benchmarks.json"

if [ ! -f "$BENCHMARKS_FILE" ]; then
    echo "Error: Benchmarks file not found at $BENCHMARKS_FILE"
    exit 1
fi

AUTH_HEADER="Authorization: Bearer $API_KEY"

echo "🏋️ Seeding benchmarks into AgentTrace..."
echo "   API: $API_URL"
echo ""

BENCHMARK_COUNT=$(jq '.benchmarks | length' "$BENCHMARKS_FILE")

for i in $(seq 0 $((BENCHMARK_COUNT - 1))); do
    BENCHMARK=$(jq ".benchmarks[$i]" "$BENCHMARKS_FILE")
    NAME=$(echo "$BENCHMARK" | jq -r '.name')
    DESCRIPTION=$(echo "$BENCHMARK" | jq -r '.description')
    CATEGORY=$(echo "$BENCHMARK" | jq -r '.category')

    echo "📦 Creating benchmark: $NAME"

    # Create a dataset for this benchmark
    DATASET_BODY=$(jq -n \
        --arg name "benchmark-$CATEGORY" \
        --arg desc "$DESCRIPTION" \
        '{"name": $name, "description": $desc}')

    DATASET_RESP=$(curl -s -X POST "$API_URL/api/public/datasets" \
        -H "$AUTH_HEADER" \
        -H "Content-Type: application/json" \
        -d "$DATASET_BODY")

    DATASET_ID=$(echo "$DATASET_RESP" | jq -r '.id // empty')
    if [ -z "$DATASET_ID" ]; then
        echo "   ⚠ Dataset may already exist, trying to fetch..."
        DATASET_RESP=$(curl -s "$API_URL/api/public/datasets?name=benchmark-$CATEGORY" \
            -H "$AUTH_HEADER")
        DATASET_ID=$(echo "$DATASET_RESP" | jq -r '.[0].id // .id // empty')
        if [ -z "$DATASET_ID" ]; then
            echo "   ❌ Failed to create or find dataset for $NAME, skipping"
            continue
        fi
    fi

    echo "   Dataset ID: $DATASET_ID"

    # Add tasks as dataset items
    TASK_COUNT=$(echo "$BENCHMARK" | jq '.tasks | length')
    echo "   Adding $TASK_COUNT tasks..."

    for j in $(seq 0 $((TASK_COUNT - 1))); do
        TASK=$(echo "$BENCHMARK" | jq ".tasks[$j]")
        INPUT=$(echo "$TASK" | jq '.input')
        EXPECTED=$(echo "$TASK" | jq '.expectedOutput')
        META=$(echo "$TASK" | jq '.metadata')

        ITEM_BODY=$(jq -n \
            --argjson input "$INPUT" \
            --argjson expected "$EXPECTED" \
            --argjson metadata "$META" \
            '{"input": $input, "expectedOutput": $expected, "metadata": $metadata}')

        ITEM_RESP=$(curl -s -X POST "$API_URL/api/public/datasets/$DATASET_ID/items" \
            -H "$AUTH_HEADER" \
            -H "Content-Type: application/json" \
            -d "$ITEM_BODY")

        ITEM_ID=$(echo "$ITEM_RESP" | jq -r '.id // empty')
        if [ -n "$ITEM_ID" ]; then
            echo "   ✅ Task $((j + 1))/$TASK_COUNT added"
        else
            echo "   ⚠ Task $((j + 1))/$TASK_COUNT may already exist"
        fi
    done

    echo ""
done

echo "✅ Benchmark seeding complete!"
echo "   $BENCHMARK_COUNT benchmarks processed"
echo ""
echo "View benchmarks at: $API_URL/api/public/benchmarks"
