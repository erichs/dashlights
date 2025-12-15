#!/bin/bash
#
# load-test.sh - Generate predictable system load (CPU + filesystem)
# Usage: ./load-test.sh [light|medium|heavy]

set -eo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Track worker PIDs for cleanup
CPU_PIDS=()
IO_PIDS=()
TEMP_DIR=""

# Cleanup function
cleanup() {
    echo -e "\n${YELLOW}Cleaning up...${NC}"

    # Kill CPU workers
    if [ ${#CPU_PIDS[@]} -gt 0 ]; then
        for pid in "${CPU_PIDS[@]}"; do
            kill "$pid" 2>/dev/null || true
        done
    fi

    # Kill I/O workers
    if [ ${#IO_PIDS[@]} -gt 0 ]; then
        for pid in "${IO_PIDS[@]}"; do
            kill "$pid" 2>/dev/null || true
        done
    fi

    # Clean up temp directory
    if [ -n "$TEMP_DIR" ] && [ -d "$TEMP_DIR" ]; then
        rm -rf "$TEMP_DIR"
    fi

    echo -e "${GREEN}Cleanup complete${NC}"
    exit 0
}

# Trap signals for cleanup
trap cleanup SIGINT SIGTERM EXIT

# Parse load level
LOAD_LEVEL="${1:-light}"

case "$LOAD_LEVEL" in
    light)
        CPU_WORKERS=1
        IO_WORKERS=1
        IO_OPS_PER_SEC=10
        ;;
    medium)
        CPU_WORKERS=2
        IO_WORKERS=2
        IO_OPS_PER_SEC=50
        ;;
    heavy)
        CPU_WORKERS=4
        IO_WORKERS=4
        IO_OPS_PER_SEC=200
        ;;
    *)
        echo -e "${RED}Error: Invalid load level: $LOAD_LEVEL${NC}"
        echo "Usage: $0 [light|medium|heavy]"
        exit 1
        ;;
esac

# Create temp directory for I/O operations
TEMP_DIR="/tmp/dashlights-load-test-$$"
mkdir -p "$TEMP_DIR"

echo -e "${GREEN}Starting ${LOAD_LEVEL} system load...${NC}"
echo "CPU workers: $CPU_WORKERS"
echo "I/O workers: $IO_WORKERS"
echo "I/O ops/sec: ~$IO_OPS_PER_SEC per worker"
echo "Temp dir: $TEMP_DIR"
echo -e "${YELLOW}Press Ctrl+C to stop${NC}\n"

# Spawn CPU workers
for i in $(seq 1 "$CPU_WORKERS"); do
    yes > /dev/null &
    pid=$!
    CPU_PIDS+=($pid)
    echo "Started CPU worker $i (PID: $pid)"
done

# I/O worker function
io_worker() {
    local worker_id=$1
    local ops_per_sec=$2
    local sleep_time=$(awk "BEGIN {print 1.0/$ops_per_sec}")

    while true; do
        # Create a small file
        echo "test data $$" > "$TEMP_DIR/file-$worker_id-$RANDOM.txt" 2>/dev/null || true

        # Stat some files
        stat "$TEMP_DIR" >/dev/null 2>&1 || true

        # List directory
        ls "$TEMP_DIR" >/dev/null 2>&1 || true

        # Delete old files (keep last 100)
        file_count=$(ls -1 "$TEMP_DIR" 2>/dev/null | wc -l)
        if [ "$file_count" -gt 100 ]; then
            ls -1t "$TEMP_DIR" | tail -n +101 | xargs -I {} rm -f "$TEMP_DIR/{}" 2>/dev/null || true
        fi

        sleep "$sleep_time"
    done
}

# Spawn I/O workers
for i in $(seq 1 "$IO_WORKERS"); do
    io_worker "$i" "$IO_OPS_PER_SEC" &
    pid=$!
    IO_PIDS+=($pid)
    echo "Started I/O worker $i (PID: $pid)"
done

echo -e "\n${GREEN}Load generation active${NC}"

# Wait indefinitely (cleanup will happen on signal)
wait
