#!/bin/bash

# Article Assistant Comprehensive Test Runner
# Tests all implemented features of the article chat system

set -e

# Configuration
API_BASE="http://localhost:8080"
RESULTS_DIR="test_results"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
RESULTS_FILE="$RESULTS_DIR/results_$TIMESTAMP.json"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Create results directory
mkdir -p "$RESULTS_DIR"

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
START_TIME=$(date +%s)

# Results storage
TEST_RESULTS=()

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
}

log_error() {
    echo -e "${RED}[FAIL]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_test() {
    echo -e "${PURPLE}[TEST]${NC} $1"
}

# Test function that captures responses
test_api_query() {
    local test_name="$1"
    local query="$2"
    local expected_intent="$3"
    local should_contain="$4"
    
    log_test "Test: $test_name"
    
    local response
    local start_time
    local end_time
    
    start_time=$(date +%s.%N)
    
    response=$(curl -s -X POST "$API_BASE/chat" \
        -H "Content-Type: application/json" \
        -d "{\"query\": \"$query\"}")
    
    end_time=$(date +%s.%N)
    local duration=$(echo "$end_time - $start_time" | bc)
    
    # Check if response contains expected content
    local content_check=""
    if [[ -n "$should_contain" ]]; then
        if echo "$response" | grep -qi "$should_contain"; then
            log_success "✅ Contains expected content: $should_contain"
            content_check="PASSED"
        else
            log_warning "⚠️  Missing expected content: $should_contain"
            content_check="FAILED"
        fi
    fi
    
    # Check if response is valid JSON with answer field
    local json_check=""
    if echo "$response" | jq -e '.answer' > /dev/null 2>&1; then
        log_success "✅ Valid JSON response with answer field"
        json_check="PASSED"
    else
        log_error "❌ Invalid JSON response or missing answer field"
        json_check="FAILED"
        ((FAILED_TESTS++))
    fi
    
    # Log timing
    log_info "⏱️  Response time: ${duration}s"
    
    # Store test result
    local test_result=$(jq -n --arg name "$test_name" --arg query "$query" --arg expected_intent "$expected_intent" --arg status "$json_check" --arg duration "$duration" --arg should_contain "$should_contain" --arg content_check "$content_check" --argjson response "$response" '
        {
            name: $name,
            query: $query,
            expected_intent: $expected_intent,
            status: $status,
            duration: ($duration | tonumber),
            should_contain: $should_contain,
            content_check: $content_check,
            response: $response
        }
    ')
    
    TEST_RESULTS+=("$test_result")
    
    if [[ "$json_check" == "PASSED" ]]; then
        ((PASSED_TESTS++))
    fi
    ((TOTAL_TESTS++))
    
    echo
}

# Test article injection
test_article_injection() {
    local test_name="$1"
    local url="$2"
    
    log_test "Test: $test_name"
    
    local response
    local start_time
    local end_time
    
    start_time=$(date +%s.%N)
    
    response=$(curl -s -X POST "$API_BASE/articles" \
        -H "Content-Type: application/json" \
        -d "{\"url\": \"$url\"}")
    
    end_time=$(date +%s.%N)
    local duration=$(echo "$end_time - $start_time" | bc)
    
    # Check if response is valid JSON
    local json_check=""
    if echo "$response" | jq -e '.url' > /dev/null 2>&1; then
        log_success "✅ Article successfully injected"
        json_check="PASSED"
    else
        log_error "❌ Failed to inject article"
        json_check="FAILED"
        ((FAILED_TESTS++))
    fi
    
    # Log timing
    log_info "⏱️  Response time: ${duration}s"
    
    # Store test result
    local test_result=$(jq -n --arg name "$test_name" --arg url "$url" --arg status "$json_check" --arg duration "$duration" --argjson response "$response" '
        {
            name: $name,
            url: $url,
            status: $status,
            duration: ($duration | tonumber),
            response: $response
        }
    ')
    
    TEST_RESULTS+=("$test_result")
    
    if [[ "$json_check" == "PASSED" ]]; then
        ((PASSED_TESTS++))
    fi
    ((TOTAL_TESTS++))
    
    echo
}

# Check if server is running
check_server() {
    log_info "Checking if server is running..."
    if ! curl -s "$API_BASE/chat" -X POST -H "Content-Type: application/json" -d '{"query":"test"}' > /dev/null 2>&1; then
        log_error "Server is not running at $API_BASE"
        log_info "Please start the server with: LLM_PROVIDER=openai ./server"
        exit 1
    fi
    log_success "Server is running"
}

# Inject test articles
inject_test_articles() {
    log_info "Injecting test articles..."
    
    # Test articles for comprehensive testing
    local test_articles=(
        "https://techcrunch.com/2025/07/25/sam-altman-warns-theres-no-legal-confidentiality-when-using-chatgpt-as-a-therapist/"
        "https://techcrunch.com/2025/07/25/meta-names-shengjia-zhao-as-chief-scientist-of-ai-superintelligence-unit/"
        "https://techcrunch.com/2025/07/27/wizard-of-oz-blown-up-by-ai-for-giant-sphere-screen/"
        "https://edition.cnn.com/2025/07/24/tech/intel-layoffs-15-percent-q2-earnings"
    )
    
    local injected=0
    local failed=0
    
    for url in "${test_articles[@]}"; do
        log_info "Injecting: $url"
        
        if curl -s -X POST "$API_BASE/articles" \
            -H "Content-Type: application/json" \
            -d "{\"url\": \"$url\"}" | jq -e '.url' > /dev/null 2>&1; then
            ((injected++))
            log_success "✅ Injected: $url"
        else
            ((failed++))
            log_warning "⚠️  Failed or already exists: $url"
        fi
        
        # Small delay to avoid overwhelming the server
        sleep 2
    done
    
    log_info "Injection complete: $injected success, $failed failed/skipped"
}

# Run comprehensive tests
run_comprehensive_tests() {
    log_info "=== Comprehensive Feature Tests ==="
    echo
    
    # Test 1: Article Summarization
    test_api_query \
        "Article Summarization" \
        "Summarize the article https://techcrunch.com/2025/07/25/sam-altman-warns-theres-no-legal-confidentiality-when-using-chatgpt-as-a-therapist/" \
        "SUMMARIZE" \
        "Sam Altman"
    
    # Test 2: Keywords/Topics Extraction
    test_api_query \
        "Keywords Extraction" \
        "Extract keywords from the article https://techcrunch.com/2025/07/25/sam-altman-warns-theres-no-legal-confidentiality-when-using-chatgpt-as-a-therapist/" \
        "KEYWORDS" \
        "KEYWORDS"
    
    # Test 3: Sentiment Analysis
    test_api_query \
        "Sentiment Analysis" \
        "What is the sentiment of the article https://techcrunch.com/2025/07/27/wizard-of-oz-blown-up-by-ai-for-giant-sphere-screen/?" \
        "SENTIMENT" \
        "sentiment"
    
    # Test 4: Topic-based Search
    test_api_query \
        "Topic Search" \
        "What articles discuss AI regulation?" \
        "FIND_BY_TOPIC" \
        "articles"
    
    # Test 5: Entity Extraction
    test_api_query \
        "Entity Extraction" \
        "What are the most commonly discussed entities across all articles?" \
        "FIND_COMMON_ENTITIES" \
        "entities"
    
    # Test 6: Article Comparison - Tone
    test_api_query \
        "Tone Comparison" \
        "What are the key differences in tone between https://techcrunch.com/2025/07/25/sam-altman-warns-theres-no-legal-confidentiality-when-using-chatgpt-as-a-therapist/ and https://techcrunch.com/2025/07/25/meta-names-shengjia-zhao-as-chief-scientist-of-ai-superintelligence-unit/?" \
        "COMPARE_TONE" \
        "tone"
    
    # Test 7: Article Comparison - Positivity
    test_api_query \
        "Positivity Comparison" \
        "Which article is more positive about AI: https://techcrunch.com/2025/07/25/sam-altman-warns-theres-no-legal-confidentiality-when-using-chatgpt-as-a-therapist/ or https://techcrunch.com/2025/07/25/meta-names-shengjia-zhao-as-chief-scientist-of-ai-superintelligence-unit/?" \
        "COMPARE_POSITIVITY" \
        "positive"
    
    # Test 8: Economic Trends Search
    test_api_query \
        "Economic Trends Search" \
        "What articles discuss economic trends?" \
        "FIND_BY_TOPIC" \
        "economic"
    
    # Test 9: General Article Search
    test_api_query \
        "General Article Search" \
        "Find articles about technology companies" \
        "FIND_BY_TOPIC" \
        "technology"
    
    # Test 10: Multiple Article Analysis
    test_api_query \
        "Multiple Article Analysis" \
        "Compare the sentiment of all articles about AI" \
        "SENTIMENT" \
        "sentiment"
    
    log_info "Comprehensive tests completed: $PASSED_TESTS passed, $FAILED_TESTS failed"
    echo
}

# Save results to JSON file
save_results() {
    log_info "Saving test results..."
    
    local end_time=$(date +%s)
    local total_time=$((end_time - START_TIME))
    
    # Create a temporary file for JSON construction
    local temp_json="/tmp/test_results_$$.json"
    
    # Start JSON structure
    cat > "$temp_json" << EOF
{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "test_summary": {
    "total_tests": $TOTAL_TESTS,
    "passed_tests": $PASSED_TESTS,
    "failed_tests": $FAILED_TESTS,
    "success_rate": $(( PASSED_TESTS * 100 / TOTAL_TESTS )),
    "total_time": $total_time
  },
  "api_info": {
    "base_url": "http://localhost:8080",
    "endpoints_tested": ["/chat", "/articles"]
  },
  "tests": [
EOF

    # Add test results
    for i in "${!TEST_RESULTS[@]}"; do
        if [[ $i -gt 0 ]]; then
            echo "," >> "$temp_json"
        fi
        echo "${TEST_RESULTS[$i]}" >> "$temp_json"
    done
    
    # Close JSON structure
    cat >> "$temp_json" << EOF
  ]
}
EOF
    
    # Copy to final location
    cp "$temp_json" "$RESULTS_FILE"
    rm "$temp_json"
    
    log_success "Results saved to: $RESULTS_FILE"
}

# Generate final report
generate_report() {
    log_info "Generating test report..."
    
    local report_file="$RESULTS_DIR/report_$TIMESTAMP.txt"
    local success_rate=$(( PASSED_TESTS * 100 / TOTAL_TESTS ))
    
    cat > "$report_file" << EOF
Article Assistant Comprehensive Test Report
==========================================
Timestamp: $(date)
Total Tests: $TOTAL_TESTS
Passed: $PASSED_TESTS
Failed: $FAILED_TESTS
Success Rate: $success_rate%

Features Tested:
✅ Article Summarization
✅ Keywords/Topics Extraction  
✅ Sentiment Analysis
✅ Topic-based Search
✅ Entity Extraction
✅ Article Comparison (Tone)
✅ Article Comparison (Positivity)
✅ Economic Trends Search
✅ General Article Search
✅ Multiple Article Analysis

API Endpoints:
- POST /chat - Main chat interface
- POST /articles - Article injection

Data Source: Test articles injected during test run
API Base: $API_BASE
Results Directory: $RESULTS_DIR
Results File: $RESULTS_FILE
EOF

    log_success "Report generated: $report_file"
    
    if [[ $FAILED_TESTS -gt 0 ]]; then
        log_warning "Some tests failed. Check the report for details."
        echo -e "${YELLOW}Success Rate: $success_rate%${NC}"
    else
        log_success "All tests passed! 🎉"
        echo -e "${GREEN}Success Rate: 100%${NC}"
    fi
}

# Main execution
main() {
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}  Article Assistant Test Runner  ${NC}"
    echo -e "${BLUE}================================${NC}"
    echo
    
    # Pre-flight checks
    check_server
    
    # Inject test articles
    inject_test_articles
    
    echo
    log_info "Starting comprehensive tests..."
    echo
    
    # Run test suites
    run_comprehensive_tests
    
    echo
    log_info "Tests completed!"
    
    # Save results to JSON file
    save_results
    
    # Generate final report
    generate_report
}

# Run main function
main "$@"
