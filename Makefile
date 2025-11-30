.PHONY: all help build test test-integration fmt fmt-check clean install hooks coverage coverage-html coverage-signals gosec

# Detect Go bin directory portably
GOBIN := $(shell go env GOBIN)
ifeq ($(GOBIN),)
    GOBIN := $(shell go env GOPATH)/bin
endif

# Default target - format, build, test, and security scan
all: fmt build test test-integration test-race gosec
	@echo "✅ All checks passed!"

help:
	@echo "Available targets:"
	@echo "  make                   - Format, build, test, and security scan (default)"
	@echo "  make all               - Same as default"
	@echo "  make build             - Build the dashlights binary"
	@echo "  make test              - Run all tests"
	@echo "  make test-integration  - Run integration tests (including performance)"
	@echo "  make test-race         - Run tests with race detector"
	@echo "  make gosec             - Run security scanner (gosec) with audit mode"
	@echo "  make coverage          - Run tests with coverage report"
	@echo "  make coverage-html     - Generate HTML coverage report"
	@echo "  make coverage-signals  - Show detailed coverage for signals package"
	@echo "  make fmt               - Format all Go files"
	@echo "  make fmt-check         - Check if files need formatting (CI-friendly)"
	@echo "  make clean             - Remove built binaries"
	@echo "  make install           - Install dashlights to GOPATH/bin"
	@echo "  make hooks             - Install Git hooks from scripts/hooks/"

# Build the binary
build:
	@echo "Generating repository URL..."
	@cd src && go generate
	@echo "Building dashlights..."
	@go build -o dashlights ./src

# Run all tests
test:
	@echo "Running tests..."
	@go test ./...

# Run integration tests (including performance)
test-integration:
	@echo "Running integration tests..."
	@go test -tags=integration -v -run TestPerformanceThreshold ./src

test-race:
	@echo "Running concurrency race tests..."
	@go test -race ./...

# Run security scanner with audit mode
gosec:
	@echo "Running security scanner (gosec)..."
	@if [ ! -f $(GOBIN)/gosec ]; then \
		echo "gosec not found. Installing..."; \
		go install github.com/securego/gosec/v2/cmd/gosec@latest; \
	fi
	@$(GOBIN)/gosec -conf gosec-config.json ./...
	@echo "✅ Security scan passed - no issues found!"

# Generate coverage report
coverage:
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out
	@echo ""
	@echo "Coverage summary:"
	@go tool cover -func=coverage.out | grep total | awk '{print "Total: " $$3}'
	@echo ""
	@echo "Signals package coverage:"
	@go tool cover -func=coverage.out | grep src/signals/ | tail -1 | awk '{print $$1 ": " $$3}'

# Generate HTML coverage report
coverage-html: coverage
	@echo "Generating HTML coverage report..."
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"
	@echo "Open coverage.html in your browser to view detailed coverage"

# Show detailed coverage for signals package
coverage-signals:
	@echo "Running tests with coverage for signals package..."
	@go test -coverprofile=coverage.out ./src/signals/
	@echo ""
	@echo "Signals package coverage:"
	@go tool cover -func=coverage.out | grep "src/signals/" | awk '{printf "%-60s %s\n", $$1 ":" $$2, $$3}'
	@echo ""
	@echo "Summary:"
	@go tool cover -func=coverage.out | grep total | awk '{print "Total coverage: " $$3}'

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
	@rm -f dashlights coverage.out coverage.html
	@echo "✅ Clean complete"

# Install to GOPATH/bin
install:
	@echo "Installing dashlights..."
	@go install
	@echo "✅ Installed to $$(go env GOPATH)/bin/dashlights"

# Install Git hooks
hooks:
	@echo "Installing Git hooks..."
	@cp scripts/hooks/pre-commit .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "  ✓ Installed pre-commit hook"
	@cp scripts/hooks/pre-push .git/hooks/pre-push
	@chmod +x .git/hooks/pre-push
	@echo "  ✓ Installed pre-push hook"
	@echo "✅ Git hooks installed successfully"

