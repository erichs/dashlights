#!/bin/bash
#
# stress-dashlights.sh - Run dashlights repeatedly under load and collect timing statistics
# Usage: ./stress-dashlights.sh [light|medium|heavy] [samples]

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
LOAD_LEVEL="${1:-medium}"
SAMPLES="${2:-50}"
LOAD_PID=""
DASHLIGHTS_BIN="./dashlights"
TIMEOUT_THRESHOLD_MS=10  # Match the production watchdog timeout

# Statistics
TOTAL_RUNS=0
SUCCESSFUL_RUNS=0
TIMEOUTS=0
declare -a EXECUTION_TIMES=()
declare -a TIMEOUT_TIMES=()

# Cleanup function
cleanup() {
    echo -e "\n${YELLOW}Stopping test...${NC}"

    # Kill load generator if running
    if [ -n "$LOAD_PID" ]; then
        kill "$LOAD_PID" 2>/dev/null || true
        wait "$LOAD_PID" 2>/dev/null || true
    fi

    # Print summary
    print_summary

    exit 0
}

# Trap signals for cleanup
trap cleanup SIGINT SIGTERM EXIT

# Print summary statistics
print_summary() {
    if [ "$TOTAL_RUNS" -eq 0 ]; then
        echo -e "${RED}No test runs completed${NC}"
        return
    fi

    echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}STRESS TEST SUMMARY${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}\n"

    echo -e "${YELLOW}Load Level:${NC} $LOAD_LEVEL"
    echo -e "${YELLOW}Total Runs:${NC} $TOTAL_RUNS"

    local success_no_timeout=$((SUCCESSFUL_RUNS - TIMEOUTS))
    local success_pct=$(awk "BEGIN {printf \"%.1f\", ($success_no_timeout/$TOTAL_RUNS)*100}")
    local timeout_pct=$(awk "BEGIN {printf \"%.1f\", ($TIMEOUTS/$TOTAL_RUNS)*100}")

    echo -e "${GREEN}Under ${TIMEOUT_THRESHOLD_MS}ms:${NC} $success_no_timeout ($success_pct%)"
    echo -e "${RED}Over ${TIMEOUT_THRESHOLD_MS}ms:${NC}  $TIMEOUTS ($timeout_pct%) ${RED}← would timeout in production${NC}"

    if [ "${#EXECUTION_TIMES[@]}" -gt 0 ]; then
        # Calculate min, max, avg
        local min_time=999999
        local max_time=0
        local sum=0

        for time in "${EXECUTION_TIMES[@]}"; do
            # Parse time and convert to microseconds for comparison
            # Go's time.Duration can output µs, ms, or s depending on magnitude
            local time_value=$(echo "$time" | grep -Eo '^[0-9.]+')
            local time_unit=$(echo "$time" | grep -Eo '[µa-zA-Z]+$')
            local time_us
            case "$time_unit" in
                ms)
                    time_us=$(awk "BEGIN {print $time_value * 1000}")
                    ;;
                µs|us)
                    time_us="$time_value"
                    ;;
                s)
                    time_us=$(awk "BEGIN {print $time_value * 1000000}")
                    ;;
                *)
                    # Assume milliseconds if unknown
                    time_us=$(awk "BEGIN {print $time_value * 1000}")
                    ;;
            esac

            if [ "${time_us%.*}" -lt "${min_time%.*}" ]; then
                min_time=$time_us
            fi

            if [ "${time_us%.*}" -gt "${max_time%.*}" ]; then
                max_time=$time_us
            fi

            sum=$(awk "BEGIN {print $sum + $time_us}")
        done

        local avg=$(awk "BEGIN {printf \"%.2f\", $sum / ${#EXECUTION_TIMES[@]} / 1000}")
        min_time=$(awk "BEGIN {printf \"%.2f\", $min_time / 1000}")
        max_time=$(awk "BEGIN {printf \"%.2f\", $max_time / 1000}")

        echo -e "\n${YELLOW}Execution Time:${NC}"
        echo -e "  Min: ${min_time}ms"
        echo -e "  Avg: ${avg}ms"
        echo -e "  Max: ${max_time}ms"
    fi

    # Show timeout details if any
    if [ "${#TIMEOUT_TIMES[@]}" -gt 0 ]; then
        echo -e "\n${RED}Timeout Details (exceeded ${TIMEOUT_THRESHOLD_MS}ms):${NC}"
        for time in "${TIMEOUT_TIMES[@]}"; do
            echo -e "  ${RED}•${NC} $time"
        done
    fi

    echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

# Check if dashlights binary exists
if [ ! -f "$DASHLIGHTS_BIN" ]; then
    echo -e "${RED}Error: dashlights binary not found at $DASHLIGHTS_BIN${NC}"
    echo "Run 'make build' first"
    exit 1
fi

echo -e "${GREEN}Starting stress test...${NC}"
echo "Load level: $LOAD_LEVEL"
echo "Samples: $SAMPLES"
echo "Timeout threshold: ${TIMEOUT_THRESHOLD_MS}ms (production watchdog limit)"
echo -e "${YELLOW}Press Ctrl+C to stop early${NC}\n"

# Start load generator in background
bash scripts/load-test.sh "$LOAD_LEVEL" >/dev/null 2>&1 &
LOAD_PID=$!

# Give load generator time to start
sleep 2

echo -e "${BLUE}Running dashlights under load...${NC}\n"

# Run dashlights repeatedly
for i in $(seq 1 "$SAMPLES"); do
    printf "Run %3d/%d: " "$i" "$SAMPLES"

    # Run dashlights with debug mode and capture output
    output=$($DASHLIGHTS_BIN --debug 2>&1 || true)

    TOTAL_RUNS=$((TOTAL_RUNS + 1))

    # Check if it produced output (successful run)
    if echo "$output" | grep -q "Total execution:"; then
        # Extract execution time
        exec_time=$(echo "$output" | grep "Total execution:" | awk '{print $3}')
        EXECUTION_TIMES+=("$exec_time")
        SUCCESSFUL_RUNS=$((SUCCESSFUL_RUNS + 1))

        # Check if this would have timed out in production (>10ms)
        # Parse execution time and convert to milliseconds
        # Go's time.Duration can output µs, ms, or s depending on magnitude
        exec_time_value=$(echo "$exec_time" | grep -Eo '^[0-9.]+')
        exec_time_unit=$(echo "$exec_time" | grep -Eo '[µa-zA-Z]+$')
        case "$exec_time_unit" in
            ms)
                exec_time_ms="$exec_time_value"
                ;;
            µs|us)
                exec_time_ms=$(awk "BEGIN {printf \"%.4f\", $exec_time_value / 1000}")
                ;;
            s)
                exec_time_ms=$(awk "BEGIN {printf \"%.4f\", $exec_time_value * 1000}")
                ;;
            *)
                echo -e "${YELLOW}Warning:${NC} Unknown time unit '$exec_time_unit'. Assuming ms."
                exec_time_ms="$exec_time_value"
                ;;
        esac
        would_timeout=$(awk "BEGIN {print ($exec_time_ms > $TIMEOUT_THRESHOLD_MS) ? 1 : 0}")

        if [ "$would_timeout" -eq 1 ]; then
            TIMEOUTS=$((TIMEOUTS + 1))
            TIMEOUT_TIMES+=("$exec_time")
            echo -e "${RED}✗ TIMEOUT${NC} $exec_time (exceeded ${TIMEOUT_THRESHOLD_MS}ms)"
        else
            echo -e "${GREEN}✓${NC} $exec_time"
        fi
    else
        # Actually crashed or failed to produce output
        TIMEOUTS=$((TIMEOUTS + 1))
        echo -e "${RED}✗ CRASHED${NC}"
    fi

    # Small delay between runs
    sleep 0.1
done

# Cleanup will happen via trap
