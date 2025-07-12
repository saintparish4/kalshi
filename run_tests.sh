#!/bin/bash

# Performance Test Script for API Gateway
# This script runs comprehensive performance tests and generates reports

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
DEFAULT_BASE_URL="http://localhost:8080"
DEFAULT_TEST_DURATION="2m"
DEFAULT_MAX_RPS="10000"
DEFAULT_CONCURRENCY="1000"
RESULTS_DIR="performance_results"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "\n${BLUE}================================"
    echo -e "$1"
    echo -e "================================${NC}\n"
}

# Function to check if API Gateway is running
check_api_gateway() {
    local base_url=${1:-$DEFAULT_BASE_URL}
    print_info "Checking if API Gateway is running at $base_url..."
    
    if curl -s -o /dev/null -w "%{http_code}" "$base_url/health" | grep -q "200"; then
        print_success "API Gateway is running and responding"
        return 0
    else
        print_error "API Gateway is not responding at $base_url"
        print_info "Make sure your API Gateway is running with: go run cmd/server/main.go"
        return 1
    fi
}

# Function to setup test environment
setup_test_env() {
    print_info "Setting up test environment..."
    
    # Create results directory
    mkdir -p "$RESULTS_DIR"
    
    # Check if performance test binary exists
    if [ ! -f "cmd/performance/performance_test" ]; then
        print_info "Building performance test binary..."
        cd cmd/performance && go build -o performance_test . && cd ../..
    fi
    
    # Check if authenticated performance test binary exists
    if [ ! -f "cmd/auth-performance/auth-performance-test" ]; then
        print_info "Building authenticated performance test binary..."
        cd cmd/auth-performance && go build -o auth-performance-test . && cd ../..
    fi
    
    # Create test configuration
    cat > test_config.json << EOF
{
    "base_url": "${TEST_BASE_URL:-$DEFAULT_BASE_URL}",
    "duration": "${TEST_DURATION:-$DEFAULT_TEST_DURATION}",
    "initial_rps": 100,
    "max_rps": ${TEST_MAX_RPS:-$DEFAULT_MAX_RPS},
    "ramp_up_duration": "30s",
    "warmup_duration": "10s",
    "max_concurrency": ${TEST_CONCURRENCY:-$DEFAULT_CONCURRENCY},
    "test_endpoints": [
        "/health",
        "/metrics",
        "/api/v1/health",
        "/api/v1/metrics"
    ]
}
EOF
    
    print_success "Test environment ready"
}

# Function to run basic performance test
run_basic_test() {
    print_header "RUNNING BASIC PERFORMANCE TEST"
    
    export TEST_BASE_URL=${TEST_BASE_URL:-$DEFAULT_BASE_URL}
    export TEST_MAX_RPS=500
    export TEST_DURATION=1m
    export TEST_WORKER_POOL=50
    
    print_info "Test Configuration:"
    print_info "  Base URL: $TEST_BASE_URL"
    print_info "  Max RPS: $TEST_MAX_RPS"
    print_info "  Duration: $TEST_DURATION"
    print_info "  Worker Pool: $TEST_WORKER_POOL"
    print_info "  Concurrency: 100"
    
    ./cmd/performance/performance_test > "$RESULTS_DIR/basic_test_$TIMESTAMP.log" 2>&1
    
    if [ $? -eq 0 ]; then
        print_success "Basic test completed successfully"
        tail -30 "$RESULTS_DIR/basic_test_$TIMESTAMP.log"
    else
        print_error "Basic test failed"
        tail -10 "$RESULTS_DIR/basic_test_$TIMESTAMP.log"
    fi
}

# Function to run stress test
run_stress_test() {
    print_header "RUNNING STRESS TEST"
    
    export TEST_BASE_URL=${TEST_BASE_URL:-$DEFAULT_BASE_URL}
    export TEST_MAX_RPS=25000
    export TEST_DURATION=3m
    
    print_info "Test Configuration:"
    print_info "  Base URL: $TEST_BASE_URL"
    print_info "  Max RPS: $TEST_MAX_RPS"
    print_info "  Duration: $TEST_DURATION"
    print_info "  Concurrency: 2000"
    
    ./cmd/performance/performance_test > "$RESULTS_DIR/stress_test_$TIMESTAMP.log" 2>&1
    
    if [ $? -eq 0 ]; then
        print_success "Stress test completed successfully"
        tail -30 "$RESULTS_DIR/stress_test_$TIMESTAMP.log"
    else
        print_error "Stress test failed"
        tail -10 "$RESULTS_DIR/stress_test_$TIMESTAMP.log"
    fi
}

# Function to run spike test
run_spike_test() {
    print_header "RUNNING SPIKE TEST"
    
    export TEST_BASE_URL=${TEST_BASE_URL:-$DEFAULT_BASE_URL}
    export TEST_MAX_RPS=50000
    export TEST_DURATION=30s
    
    print_info "Test Configuration:"
    print_info "  Base URL: $TEST_BASE_URL"
    print_info "  Max RPS: $TEST_MAX_RPS"
    print_info "  Duration: $TEST_DURATION"
    print_info "  Concurrency: 5000"
    
    ./cmd/performance/performance_test > "$RESULTS_DIR/spike_test_$TIMESTAMP.log" 2>&1
    
    if [ $? -eq 0 ]; then
        print_success "Spike test completed successfully"
        tail -30 "$RESULTS_DIR/spike_test_$TIMESTAMP.log"
    else
        print_error "Spike test failed"
        tail -10 "$RESULTS_DIR/spike_test_$TIMESTAMP.log"
    fi
}

# Function to run endurance test
run_endurance_test() {
    print_header "RUNNING ENDURANCE TEST"
    
    export TEST_BASE_URL=${TEST_BASE_URL:-$DEFAULT_BASE_URL}
    export TEST_MAX_RPS=10000
    export TEST_DURATION=10m
    
    print_info "Test Configuration:"
    print_info "  Base URL: $TEST_BASE_URL"
    print_info "  Max RPS: $TEST_MAX_RPS"
    print_info "  Duration: $TEST_DURATION"
    print_info "  Concurrency: 1000"
    
    print_warning "This test will run for 10 minutes..."
    
    ./cmd/performance/performance_test > "$RESULTS_DIR/endurance_test_$TIMESTAMP.log" 2>&1
    
    if [ $? -eq 0 ]; then
        print_success "Endurance test completed successfully"
        tail -30 "$RESULTS_DIR/endurance_test_$TIMESTAMP.log"
    else
        print_error "Endurance test failed"
        tail -10 "$RESULTS_DIR/endurance_test_$TIMESTAMP.log"
    fi
}

# Function to run authenticated basic test
run_auth_basic_test() {
    print_header "RUNNING AUTHENTICATED BASIC TEST"
    
    export TEST_BASE_URL=${TEST_BASE_URL:-$DEFAULT_BASE_URL}
    export TEST_MAX_RPS=500
    export TEST_DURATION=1m
    export TEST_WORKER_POOL=50
    
    print_info "Test Configuration:"
    print_info "  Base URL: $TEST_BASE_URL"
    print_info "  Max RPS: $TEST_MAX_RPS"
    print_info "  Duration: $TEST_DURATION"
    print_info "  Worker Pool: $TEST_WORKER_POOL"
    print_info "  Admin User: ${TEST_ADMIN_USER:-admin}"
    print_info "  Authentication: JWT + API Keys"
    
    ./cmd/auth-performance/auth-performance-test \
        -base-url "$TEST_BASE_URL" \
        -admin-user "${TEST_ADMIN_USER:-admin}" \
        -admin-pass "${TEST_ADMIN_PASS:-password}" \
        -duration "$TEST_DURATION" \
        -max-rps "$TEST_MAX_RPS" \
        -workers "$TEST_WORKER_POOL" \
        -output "$RESULTS_DIR/auth_basic_results_$TIMESTAMP.json" \
        > "$RESULTS_DIR/auth_basic_test_$TIMESTAMP.log" 2>&1
    
    if [ $? -eq 0 ]; then
        print_success "Authenticated basic test completed successfully"
        tail -30 "$RESULTS_DIR/auth_basic_test_$TIMESTAMP.log"
    else
        print_error "Authenticated basic test failed"
        tail -10 "$RESULTS_DIR/auth_basic_test_$TIMESTAMP.log"
    fi
}

# Function to run authenticated stress test
run_auth_stress_test() {
    print_header "RUNNING AUTHENTICATED STRESS TEST"
    
    export TEST_BASE_URL=${TEST_BASE_URL:-$DEFAULT_BASE_URL}
    export TEST_MAX_RPS=25000
    export TEST_DURATION=3m
    
    print_info "Test Configuration:"
    print_info "  Base URL: $TEST_BASE_URL"
    print_info "  Max RPS: $TEST_MAX_RPS"
    print_info "  Duration: $TEST_DURATION"
    print_info "  Admin User: ${TEST_ADMIN_USER:-admin}"
    print_info "  Authentication: JWT + API Keys"
    
    ./cmd/auth-performance/auth-performance-test \
        -base-url "$TEST_BASE_URL" \
        -admin-user "${TEST_ADMIN_USER:-admin}" \
        -admin-pass "${TEST_ADMIN_PASS:-password}" \
        -duration "$TEST_DURATION" \
        -max-rps "$TEST_MAX_RPS" \
        -workers 100 \
        -output "$RESULTS_DIR/auth_stress_results_$TIMESTAMP.json" \
        > "$RESULTS_DIR/auth_stress_test_$TIMESTAMP.log" 2>&1
    
    if [ $? -eq 0 ]; then
        print_success "Authenticated stress test completed successfully"
        tail -30 "$RESULTS_DIR/auth_stress_test_$TIMESTAMP.log"
    else
        print_error "Authenticated stress test failed"
        tail -10 "$RESULTS_DIR/auth_stress_test_$TIMESTAMP.log"
    fi
}

# Function to run authenticated spike test
run_auth_spike_test() {
    print_header "RUNNING AUTHENTICATED SPIKE TEST"
    
    export TEST_BASE_URL=${TEST_BASE_URL:-$DEFAULT_BASE_URL}
    export TEST_MAX_RPS=50000
    export TEST_DURATION=30s
    
    print_info "Test Configuration:"
    print_info "  Base URL: $TEST_BASE_URL"
    print_info "  Max RPS: $TEST_MAX_RPS"
    print_info "  Duration: $TEST_DURATION"
    print_info "  Admin User: ${TEST_ADMIN_USER:-admin}"
    print_info "  Authentication: JWT + API Keys"
    
    ./cmd/auth-performance/auth-performance-test \
        -base-url "$TEST_BASE_URL" \
        -admin-user "${TEST_ADMIN_USER:-admin}" \
        -admin-pass "${TEST_ADMIN_PASS:-password}" \
        -duration "$TEST_DURATION" \
        -max-rps "$TEST_MAX_RPS" \
        -workers 200 \
        -output "$RESULTS_DIR/auth_spike_results_$TIMESTAMP.json" \
        > "$RESULTS_DIR/auth_spike_test_$TIMESTAMP.log" 2>&1
    
    if [ $? -eq 0 ]; then
        print_success "Authenticated spike test completed successfully"
        tail -30 "$RESULTS_DIR/auth_spike_test_$TIMESTAMP.log"
    else
        print_error "Authenticated spike test failed"
        tail -10 "$RESULTS_DIR/auth_spike_test_$TIMESTAMP.log"
    fi
}

# Function to run authenticated endurance test
run_auth_endurance_test() {
    print_header "RUNNING AUTHENTICATED ENDURANCE TEST"
    
    export TEST_BASE_URL=${TEST_BASE_URL:-$DEFAULT_BASE_URL}
    export TEST_MAX_RPS=10000
    export TEST_DURATION=10m
    
    print_info "Test Configuration:"
    print_info "  Base URL: $TEST_BASE_URL"
    print_info "  Max RPS: $TEST_MAX_RPS"
    print_info "  Duration: $TEST_DURATION"
    print_info "  Admin User: ${TEST_ADMIN_USER:-admin}"
    print_info "  Authentication: JWT + API Keys"
    
    print_warning "This test will run for 10 minutes..."
    
    ./cmd/auth-performance/auth-performance-test \
        -base-url "$TEST_BASE_URL" \
        -admin-user "${TEST_ADMIN_USER:-admin}" \
        -admin-pass "${TEST_ADMIN_PASS:-password}" \
        -duration "$TEST_DURATION" \
        -max-rps "$TEST_MAX_RPS" \
        -workers 100 \
        -output "$RESULTS_DIR/auth_endurance_results_$TIMESTAMP.json" \
        > "$RESULTS_DIR/auth_endurance_test_$TIMESTAMP.log" 2>&1
    
    if [ $? -eq 0 ]; then
        print_success "Authenticated endurance test completed successfully"
        tail -30 "$RESULTS_DIR/auth_endurance_test_$TIMESTAMP.log"
    else
        print_error "Authenticated endurance test failed"
        tail -10 "$RESULTS_DIR/auth_endurance_test_$TIMESTAMP.log"
    fi
}

# Function to generate summary report
generate_report() {
    print_header "GENERATING PERFORMANCE REPORT"
    
    local report_file="$RESULTS_DIR/performance_summary_$TIMESTAMP.md"
    
    cat > "$report_file" << EOF
# API Gateway Performance Test Report

**Test Date:** $(date)
**Test Environment:** ${TEST_BASE_URL:-$DEFAULT_BASE_URL}

## Test Results Summary

EOF
    
    # Extract results from each test
    for test_type in basic stress spike endurance; do
        local log_file="$RESULTS_DIR/${test_type}_test_$TIMESTAMP.log"
        if [ -f "$log_file" ]; then
            echo "### ${test_type^} Test Results" >> "$report_file"
            echo "" >> "$report_file"
            
            # Extract key metrics
            local rps=$(grep "Requests Per Second:" "$log_file" | awk '{print $4}' | head -1)
            local p95=$(grep "P95:" "$log_file" | awk '{print $2}' | head -1)
            local p99=$(grep "P99:" "$log_file" | awk '{print $2}' | head -1)
            local memory=$(grep "Heap Allocated:" "$log_file" | awk '{print $3}' | head -1)
            local tier=$(grep "Performance Tier:" "$log_file" | cut -d: -f2- | sed 's/^[[:space:]]*//' | head -1)
            
            if [ -n "$rps" ]; then
                echo "- **RPS:** $rps" >> "$report_file"
                echo "- **P95 Latency:** $p95" >> "$report_file"
                echo "- **P99 Latency:** $p99" >> "$report_file"
                echo "- **Memory Usage:** $memory MB" >> "$report_file"
                echo "- **Performance Tier:** $tier" >> "$report_file"
            else
                echo "- **Status:** Failed or incomplete" >> "$report_file"
            fi
            
            echo "" >> "$report_file"
        fi
    done
    
    # Extract results from authenticated tests
    for test_type in auth_basic auth_stress auth_spike auth_endurance; do
        local log_file="$RESULTS_DIR/${test_type}_test_$TIMESTAMP.log"
        local json_file="$RESULTS_DIR/${test_type}_results_$TIMESTAMP.json"
        
        if [ -f "$log_file" ]; then
            # Convert test_type to display name
            local display_name=$(echo "$test_type" | sed 's/auth_/Authenticated /' | sed 's/\b\w/\U&/g')
            echo "### $display_name Test Results" >> "$report_file"
            echo "" >> "$report_file"
            
            # Extract key metrics from log
            local rps=$(grep "Actual RPS:" "$log_file" | awk '{print $3}' | head -1)
            local avg_latency=$(grep "Average Latency:" "$log_file" | awk '{print $3}' | head -1)
            local error_rate=$(grep "Error Rate:" "$log_file" | awk '{print $3}' | head -1)
            local tier=$(grep "Performance Tier:" "$log_file" | cut -d: -f2- | sed 's/^[[:space:]]*//' | head -1)
            
            # Extract authentication breakdown
            local jwt_endpoints=$(grep "JWT Endpoints:" "$log_file" | awk '{print $3}' | head -1)
            local apikey_endpoints=$(grep "API Key Endpoints:" "$log_file" | awk '{print $4}' | head -1)
            local public_endpoints=$(grep "Public Endpoints:" "$log_file" | awk '{print $3}' | head -1)
            local total_keys=$(grep "API Keys Created:" "$log_file" | awk '{print $4}' | head -1)
            
            if [ -n "$rps" ]; then
                echo "- **RPS:** $rps" >> "$report_file"
                echo "- **Average Latency:** $avg_latency" >> "$report_file"
                echo "- **Error Rate:** $error_rate%" >> "$report_file"
                echo "- **Performance Tier:** $tier" >> "$report_file"
                echo "- **Authentication Breakdown:**" >> "$report_file"
                echo "  - JWT Endpoints: $jwt_endpoints" >> "$report_file"
                echo "  - API Key Endpoints: $apikey_endpoints" >> "$report_file"
                echo "  - Public Endpoints: $public_endpoints" >> "$report_file"
                echo "  - API Keys Created: $total_keys" >> "$report_file"
            else
                echo "- **Status:** Failed or incomplete" >> "$report_file"
            fi
            
            echo "" >> "$report_file"
        fi
    done
    
    cat >> "$report_file" << EOF

## Performance Tier Classification

- **God Tier:** 200,000+ RPS, P95 < 10ms, P99 < 25ms
- **Advanced Tier:** 50,000+ RPS, P95 < 25ms, P99 < 50ms
- **Mid-Tier:** 10,000+ RPS, P95 < 50ms, P99 < 100ms
- **Basic Tier:** 1,000+ RPS, P95 < 100ms, P99 < 200ms

## Files Generated

EOF
    
    # List all generated files
    ls -la "$RESULTS_DIR"/*_$TIMESTAMP.* | while read line; do
        echo "- $line" >> "$report_file"
    done
    
    print_success "Performance report generated: $report_file"
}

# Function to monitor system resources
monitor_resources() {
    print_info "Starting resource monitoring..."
    
    # Ensure results directory exists
    mkdir -p "$RESULTS_DIR"
    
    local monitor_file="$RESULTS_DIR/resource_monitor_$TIMESTAMP.log"
    
    {
        echo "Timestamp,CPU%,Memory%,LoadAvg,Connections"
        
        while true; do
            local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
            local cpu="0"
            local memory="0"
            local load="0"
            local connections="0"
            
            # Get CPU usage (Windows compatible)
            if command -v top &> /dev/null; then
                cpu=$(top -bn1 | grep "Cpu(s)" | sed "s/.*, *\([0-9.]*\)%* id.*/\1/" | awk '{print 100 - $1}')
            elif command -v wmic &> /dev/null; then
                cpu=$(wmic cpu get loadpercentage | awk 'NR==2{print $1}')
            fi
            
            # Get memory usage (Windows compatible)
            if command -v free &> /dev/null; then
                memory=$(free | grep Mem | awk '{printf "%.1f", $3/$2 * 100.0}')
            elif command -v wmic &> /dev/null; then
                memory=$(wmic OS get FreePhysicalMemory,TotalVisibleMemorySize | awk 'NR==2{printf "%.1f", (1-$1/$2)*100}')
            fi
            
            # Get load average
            if command -v uptime &> /dev/null; then
                load=$(uptime | awk -F'load average:' '{ print $2 }' | awk '{ print $1 }' | sed 's/,//')
            fi
            
            # Get connection count
            if command -v ss &> /dev/null; then
                connections=$(ss -tuln | wc -l)
            elif command -v netstat &> /dev/null; then
                connections=$(netstat -an | wc -l)
            fi
            
            echo "$timestamp,$cpu,$memory,$load,$connections"
            sleep 5
        done
    } > "$monitor_file" &
    
    local monitor_pid=$!
    echo $monitor_pid > "$RESULTS_DIR/monitor.pid"
    
    print_info "Resource monitoring started (PID: $monitor_pid)"
}

# Function to stop resource monitoring
stop_monitoring() {
    if [ -f "$RESULTS_DIR/monitor.pid" ]; then
        local monitor_pid=$(cat "$RESULTS_DIR/monitor.pid")
        if kill -0 $monitor_pid 2>/dev/null; then
            kill $monitor_pid
            print_info "Resource monitoring stopped"
        fi
        rm -f "$RESULTS_DIR/monitor.pid"
    fi
}

# Function to run all tests
run_all_tests() {
    print_header "RUNNING COMPLETE PERFORMANCE TEST SUITE"
    
    # Start resource monitoring
    monitor_resources
    
    # Run tests in sequence
    run_basic_test
    sleep 30  # Cool down between tests
    
    run_stress_test
    sleep 30
    
    run_spike_test
    sleep 30
    
    # Ask user if they want to run endurance test
    echo -e "\n${YELLOW}Do you want to run the 10-minute endurance test? (y/N):${NC}"
    read -r response
    if [[ "$response" =~ ^[Yy]$ ]]; then
        run_endurance_test
    else
        print_info "Skipping endurance test"
    fi
    
    # Stop monitoring and generate report
    stop_monitoring
    generate_report
    
    print_success "Complete test suite finished!"
    print_info "Check the $RESULTS_DIR directory for detailed results"
}

# Function to run all authenticated tests
run_all_auth_tests() {
    print_header "RUNNING COMPLETE AUTHENTICATED PERFORMANCE TEST SUITE"
    
    # Start resource monitoring
    monitor_resources
    
    # Run authenticated tests in sequence
    run_auth_basic_test
    sleep 30  # Cool down between tests
    
    run_auth_stress_test
    sleep 30
    
    run_auth_spike_test
    sleep 30
    
    # Ask user if they want to run endurance test
    echo -e "\n${YELLOW}Do you want to run the 10-minute authenticated endurance test? (y/N):${NC}"
    read -r response
    if [[ "$response" =~ ^[Yy]$ ]]; then
        run_auth_endurance_test
    else
        print_info "Skipping authenticated endurance test"
    fi
    
    # Stop monitoring and generate report
    stop_monitoring
    generate_report
    
    print_success "Complete authenticated test suite finished!"
    print_info "Check the $RESULTS_DIR directory for detailed results"
}

# Function to cleanup
cleanup() {
    print_info "Cleaning up..."
    stop_monitoring
    
    # Remove temporary files
    rm -f test_config.json
    
    print_success "Cleanup completed"
}

# Function to show help
show_help() {
    cat << EOF
API Gateway Performance Test Script

Usage: $0 [OPTIONS] [COMMAND]

Commands:
    all         Run all performance tests (default)
    auth        Run all authenticated performance tests
    basic       Run basic performance test only
    auth-basic  Run authenticated basic test only
    stress      Run stress test only
    auth-stress Run authenticated stress test only
    spike       Run spike test only
    auth-spike  Run authenticated spike test only
    endurance   Run endurance test only
    auth-endurance Run authenticated endurance test only
    monitor     Start resource monitoring only
    report      Generate report from existing results
    clean       Clean up test files and stop monitoring

Options:
    -u URL      Set base URL (default: http://localhost:8080)
    -r RPS      Set max RPS for tests (default: 10000)
    -d DURATION Set test duration (default: 2m)
    -c CONC     Set max concurrency (default: 1000)
    -h          Show this help message

Environment Variables:
    TEST_BASE_URL      Base URL for the API Gateway
    TEST_MAX_RPS       Maximum RPS to test
    TEST_DURATION      Test duration
    TEST_CONCURRENCY   Maximum concurrent connections
    TEST_ADMIN_USER    Admin username for authenticated tests (default: admin)
    TEST_ADMIN_PASS    Admin password for authenticated tests (default: password)

Examples:
    $0                                    # Run all tests with defaults
    $0 auth                               # Run all authenticated tests
    $0 -u http://localhost:9090 basic     # Run basic test on different port
    $0 -r 50000 stress                    # Run stress test with 50k RPS
    $0 -d 5m endurance                    # Run 5-minute endurance test
    $0 auth-basic                         # Run authenticated basic test
    $0 auth-stress                        # Run authenticated stress test

Authentication Notes:
    - Authenticated tests require admin credentials
    - Tests will automatically login and create API keys
    - Tests both JWT and API key authentication methods
    - Tests public, authenticated, and admin endpoints

EOF
}

# Function to validate prerequisites
check_prerequisites() {
    print_info "Checking prerequisites..."
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go to continue."
        exit 1
    fi
    
    # Check if curl is available
    if ! command -v curl &> /dev/null; then
        print_error "curl is not installed. Please install curl to continue."
        exit 1
    fi
    
    # Check if ss is available (for connection monitoring)
    if ! command -v ss &> /dev/null; then
        print_warning "ss command not found. Connection monitoring may not work properly."
    fi
    
    print_success "Prerequisites check passed"
}

# Function to estimate system capacity
estimate_capacity() {
    print_header "SYSTEM CAPACITY ESTIMATION"
    
    local cpu_cores=$(nproc 2>/dev/null || echo 4)
    local memory_gb=""
    local max_files=$(ulimit -n 2>/dev/null || echo 1024)
    
    # Try to get memory info - handle different OS
    if command -v free &> /dev/null; then
        memory_gb=$(free -g | awk '/^Mem:/{print $2}')
    elif command -v wmic &> /dev/null; then
        memory_gb=$(wmic computersystem get TotalPhysicalMemory | awk 'NR==2{print int($1/1024/1024/1024)}')
    else
        memory_gb="8"  # Default fallback
    fi
    
    print_info "System Resources:"
    print_info "  CPU Cores: $cpu_cores"
    print_info "  Memory: ${memory_gb}GB"
    print_info "  Max Open Files: $max_files"
    
    # Provide capacity estimates
    local estimated_rps=$((cpu_cores * 2000))
    local estimated_conns=$((memory_gb * 1000))
    
    print_info "\nEstimated Capacity:"
    print_info "  Theoretical Max RPS: ~$estimated_rps"
    print_info "  Recommended Max Connections: ~$estimated_conns"
    
    if [ "$max_files" -lt 65536 ] 2>/dev/null; then
        print_warning "Consider increasing ulimit -n to 65536 for high-load testing"
        print_info "Run: ulimit -n 65536"
    fi
    
    if [ "$memory_gb" -lt 4 ] 2>/dev/null; then
        print_warning "Low memory detected. Consider testing with lower concurrency levels."
    fi
}

# Parse command line arguments
while getopts "u:r:d:c:h" opt; do
    case $opt in
        u) TEST_BASE_URL="$OPTARG" ;;
        r) TEST_MAX_RPS="$OPTARG" ;;
        d) TEST_DURATION="$OPTARG" ;;
        c) TEST_CONCURRENCY="$OPTARG" ;;
        h) show_help; exit 0 ;;
        \?) print_error "Invalid option: -$OPTARG"; show_help; exit 1 ;;
    esac
done

shift $((OPTIND-1))

# Set command (default to 'all')
COMMAND=${1:-all}

# Main execution
main() {
    print_header "API GATEWAY PERFORMANCE TESTING SUITE"
    
    # Trap cleanup on exit
    trap cleanup EXIT
    
    # Check prerequisites
    check_prerequisites
    
    # Show system capacity estimation
    estimate_capacity
    
    # Setup test environment
    setup_test_env
    
    # Check if API Gateway is running
    if ! check_api_gateway "${TEST_BASE_URL:-$DEFAULT_BASE_URL}"; then
        exit 1
    fi
    
    # Execute command
    case $COMMAND in
        all)
            run_all_tests
            ;;
        auth)
            run_all_auth_tests
            ;;
        basic)
            monitor_resources
            run_basic_test
            stop_monitoring
            generate_report
            ;;
        auth-basic)
            monitor_resources
            run_auth_basic_test
            stop_monitoring
            generate_report
            ;;
        stress)
            monitor_resources
            run_stress_test
            stop_monitoring
            generate_report
            ;;
        auth-stress)
            monitor_resources
            run_auth_stress_test
            stop_monitoring
            generate_report
            ;;
        spike)
            monitor_resources
            run_spike_test
            stop_monitoring
            generate_report
            ;;
        auth-spike)
            monitor_resources
            run_auth_spike_test
            stop_monitoring
            generate_report
            ;;
        endurance)
            monitor_resources
            run_endurance_test
            stop_monitoring
            generate_report
            ;;
        auth-endurance)
            monitor_resources
            run_auth_endurance_test
            stop_monitoring
            generate_report
            ;;
        monitor)
            monitor_resources
            print_info "Monitoring started. Press Ctrl+C to stop."
            read -r
            ;;
        report)
            generate_report
            ;;
        clean)
            cleanup
            rm -rf "$RESULTS_DIR"
            print_success "All test files cleaned up"
            ;;
        *)
            print_error "Unknown command: $COMMAND"
            show_help
            exit 1
            ;;
    esac
    
    print_success "Performance testing completed!"
}

# Run main function
main "$@"