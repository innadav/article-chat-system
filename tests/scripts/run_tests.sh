#!/bin/bash

# A modern test runner for the Article Assistant API

set -e

# --- Configuration ---
API_BASE="http://localhost:8080"
TEST_DEFINITIONS="../tests.json"
RESULTS_DIR="../../test_results"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
RESULTS_FILE="$RESULTS_DIR/results_$TIMESTAMP.json"

# --- Colors ---
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

# --- Setup ---
mkdir -p "$RESULTS_DIR"
TEST_RESULTS_JSON="[]" # Start with an empty JSON array string

# --- Helper Functions ---
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_pass() { echo -e "${GREEN}[PASS]${NC} $1"; }
log_fail() { echo -e "${RED}[FAIL]${NC} $1"; }

# --- Main Test Execution Logic ---
run_api_test() {
    local test_name="$1"
    local query="$2"
    local should_contain="$3"
    
    local status="FAILED"
    local start_time=$(date +%s.%N)

    log_info "Running test: $test_name"

    local response=$(curl -s -X POST "$API_BASE/chat" \
        -H "Content-Type: application/json" \
        -d "$(jq -n --arg q "$query" '{"query":$q}')")
    
    local end_time=$(date +%s.%N)
    local duration=$(echo "$end_time - $start_time" | bc)
    
    # Validation
    if [[ -n "$response" ]] && echo "$response" | jq -e '.answer' > /dev/null 2>&1; then
        if echo "$response" | grep -qi "$should_contain"; then
            status="PASSED"
            log_pass "✅ $test_name"
        else
            log_fail "❌ $test_name (Response did not contain '$should_contain')"
        fi
    else
        log_fail "❌ $test_name (Invalid JSON response or missing .answer field)"
    fi
    
    log_info "⏱️  Response time: ${duration}s"
    
    TEST_RESULTS_JSON=$(echo "$TEST_RESULTS_JSON" | jq \
        --arg name "$test_name" --arg query "$query" --arg status "$status" \
        --arg duration "$duration" --arg should_contain "$should_contain" \
        --argjson response "$response" \
        '. += [{ "name": $name, "query": $query, "status": $status, "duration": ($duration | tonumber), "should_contain": $should_contain, "response": $response }]')
}

# --- Main Function ---
main() {
    log_info "Starting Article Assistant Test Runner..."
    if [[ ! -f "$TEST_DEFINITIONS" ]]; then
        log_fail "Test definitions file '$TEST_DEFINITIONS' not found."
        exit 1
    fi

    # Read test cases (bash 3.2 compatible)
    while IFS= read -r line; do
        test_cases+=("$line")
    done < <(jq -c '.[]' "$TEST_DEFINITIONS")
    log_info "Found ${#test_cases[@]} test cases."
    echo

    for test_case in "${test_cases[@]}"; do
        name=$(echo "$test_case" | jq -r '.name')
        query=$(echo "$test_case" | jq -r '.query')
        should_contain=$(echo "$test_case" | jq -r '.should_contain')
        
        run_api_test "$name" "$query" "$should_contain"
        echo
    done

    # Final reporting logic (simplified)
    local failed_tests=$(echo "$TEST_RESULTS_JSON" | jq '[.[] | select(.status == "FAILED")] | length')
    echo "$TEST_RESULTS_JSON" | jq '.' > "$RESULTS_FILE"
    log_info "Test run complete. Report saved to $RESULTS_FILE"
    
    if [[ $failed_tests -gt 0 ]]; then
        log_fail "Finished with $failed_tests failure(s)."
        exit 1
    else
        log_pass "All tests passed! 🎉"
    fi
}

main "$@"