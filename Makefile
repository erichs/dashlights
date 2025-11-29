.PHONY: all help build test test-integration fmt fmt-check clean install hooks

# Default target - format, build, and test
all: fmt build test test-integration test-race
	@echo "✅ All checks passed!"

help:
	@echo "Available targets:"
	@echo "  make                   - Format, build, and test (default)"
	@echo "  make all               - Same as default"
	@echo "  make build             - Build the dashlights binary"
	@echo "  make test              - Run all tests"
	@echo "  make test-integration  - Run integration tests (including performance)"
	@echo "  make fmt               - Format all Go files"
	@echo "  make fmt-check         - Check if files need formatting (CI-friendly)"
	@echo "  make clean             - Remove built binaries"
	@echo "  make install           - Install dashlights to GOPATH/bin"

# Build the binary
build:
	@echo "Building dashlights..."
	@go build -o dashlights

# Run all tests
test:
	@echo "Running tests..."
	@go test ./...

# Run integration tests (including performance)
test-integration:
	@echo "Running integration tests..."
	@go test -tags=integration -v -run TestPerformanceThreshold

test-race:
	@echo "Running concurrency race tests..."
	@go test -race ./...

# Format all Go files
fmt:
	@echo "Formatting Go files..."
	@go fmt ./...
	@echo "✅ All files formatted"

# Check if files need formatting (for CI)
fmt-check:
	@echo "Checking Go file formatting..."
	@test -z "$$(gofmt -l .)" || (echo "❌ Files need formatting. Run 'make fmt'" && gofmt -l . && exit 1)
	@echo "✅ All files properly formatted"

# Clean built binaries
clean:
	@echo "Cleaning..."
	@rm -f dashlights
	@echo "✅ Clean complete"

# Install to GOPATH/bin
install:
	@echo "Installing dashlights..."
	@go install
	@echo "✅ Installed to $$(go env GOPATH)/bin/dashlights"

