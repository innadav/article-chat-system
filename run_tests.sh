#!/bin/bash

# A modern test runner for the Article Assistant API

set -e

# --- Configuration ---
API_BASE="http://localhost:8080"
TEST_DEFINITIONS="tests.json"
RESULTS_DIR="test_results"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
RESULTS_FILE="$RESULTS_DIR/results_$TIMESTAMP.json"

# --- Colors ---
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

# --- Setup ---
mkdir -p "$RESULTS_DIR"
TEST_RESULTS="[]" # Start with an empty JSON array

# --- Helper Functions ---
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_pass() { echo -e "${GREEN}[PASS]${NC} $1"; }
log_fail() { echo -e "${RED}[FAIL]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }

# --- Server Health Check ---
check_server() {
    log_info "Checking if server is running..."
    if curl -s "$API_BASE/chat" > /dev/null 2>&1; then
        log_pass "Server is running"
        return 0
    else
        log_fail "Server is not responding at $API_BASE"
        return 1
    fi
}

# --- Main Test Execution Logic ---
run_api_test() {
    local test_name="$1"
    local query="$2"
    local should_contain="$3"
    local temp_file="$4"
    
    local status="FAILED"
    local start_time=$(date +%s.%N)

    log_info "Running test: $test_name"

    # Make the API call and capture the response
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
            log_warn "⚠️  Missing expected content: $should_contain"
            log_fail "❌ $test_name"
        fi
    else
        log_fail "❌ Invalid JSON response or missing answer field"
    fi
    
    log_info "⏱️  Response time: ${duration}s"
    
    # Append the result to the temporary file using jq
    local current_results=$(cat "$temp_file")
    echo "$current_results" | jq \
        --arg name "$test_name" \
        --arg query "$query" \
        --arg status "$status" \
        --arg duration "$duration" \
        --arg should_contain "$should_contain" \
        --argjson response "$response" \
        '. += [{
            "name": $name,
            "query": $query,
            "status": $status,
            "duration": ($duration | tonumber),
            "should_contain": $should_contain,
            "response": $response
        }]' > "$temp_file"
}

# --- Main Function ---
main() {
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}  Article Assistant Test Runner  ${NC}"
    echo -e "${BLUE}================================${NC}"
    echo

    # Check if test definitions file exists
    if [[ ! -f "$TEST_DEFINITIONS" ]]; then
        log_fail "Test definitions file '$TEST_DEFINITIONS' not found"
        exit 1
    fi

    # Check server health
    if ! check_server; then
        exit 1
    fi

    log_info "Starting comprehensive tests..."
    echo

    # Read tests from the JSON file and execute them
    local test_count=$(jq 'length' "$TEST_DEFINITIONS")
    log_info "Found $test_count test cases"
    echo

    # Create a temporary file to store results
    local temp_results="/tmp/test_results_$$.json"
    echo "[]" > "$temp_results"

    # Process each test case
    while IFS= read -r test_case; do
        name=$(echo "$test_case" | jq -r '.name')
        query=$(echo "$test_case" | jq -r '.query')
        should_contain=$(echo "$test_case" | jq -r '.should_contain')
        
        run_api_test "$name" "$query" "$should_contain" "$temp_results"
        echo # Newline for readability
    done < <(jq -c '.[]' "$TEST_DEFINITIONS")

    # Read the final results
    TEST_RESULTS=$(cat "$temp_results")
    rm -f "$temp_results"

    # Calculate final summary
    local total_tests=$(echo "$TEST_RESULTS" | jq 'length')
    local passed_tests=$(echo "$TEST_RESULTS" | jq '[.[] | select(.status == "PASSED")] | length')
    local failed_tests=$((total_tests - passed_tests))
    local success_rate=0
    if [[ $total_tests -gt 0 ]]; then
        success_rate=$((passed_tests * 100 / total_tests))
    fi
    local total_time=$(echo "$TEST_RESULTS" | jq '[.[] | .duration] | add')

    echo
    log_info "Comprehensive tests completed: $passed_tests passed, $failed_tests failed"
    echo

    # Construct the final JSON report
    local final_report=$(jq -n \
        --arg timestamp "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
        --argjson summary "$(jq -n \
            --arg total "$total_tests" \
            --arg passed "$passed_tests" \
            --arg failed "$failed_tests" \
            --arg rate "$success_rate" \
            --arg time "$total_time" \
            '{total_tests: ($total|tonumber), passed_tests: ($passed|tonumber), failed_tests: ($failed|tonumber), success_rate: ($rate|tonumber), total_time: ($time|tonumber)}')" \
        --argjson tests "$TEST_RESULTS" \
        '{
            "timestamp": $timestamp,
            "test_summary": $summary,
            "tests": $tests
        }')

    # Save the report
    echo "$final_report" | jq '.' > "$RESULTS_FILE"
    log_info "Test run complete. Report saved to $RESULTS_FILE"
    
    # Generate a simple text report
    local report_file="$RESULTS_DIR/report_$TIMESTAMP.txt"
    {
        echo "Article Assistant Test Report"
        echo "============================"
        echo "Timestamp: $(date)"
        echo "Total Tests: $total_tests"
        echo "Passed: $passed_tests"
        echo "Failed: $failed_tests"
        echo "Success Rate: $success_rate%"
        echo "Total Time: ${total_time}s"
        echo
        echo "Test Results:"
        echo "-------------"
        echo "$TEST_RESULTS" | jq -r '.[] | "\(.name): \(.status) (\(.duration)s)"'
    } > "$report_file"
    
    log_info "Text report generated: $report_file"
    
    if [[ $failed_tests -gt 0 ]]; then
        log_warn "Some tests failed. Check the report for details."
        echo -e "${YELLOW}Success Rate: $success_rate%${NC}"
        exit 1
    else
        log_pass "All tests passed! 🎉"
        echo -e "${GREEN}Success Rate: $success_rate%${NC}"
    fi
}

main "$@"
