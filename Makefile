.PHONY: all help build build-all test test-integration fmt fmt-check check-ctx clean install hooks coverage coverage-html coverage-signals gosec vet revive install-fabric-pattern release load-light load-medium load-heavy load stress-test docker-install-test

# Detect Go bin directory portably
GOBIN := $(shell go env GOBIN)
ifeq ($(GOBIN),)
    GOBIN := $(shell go env GOPATH)/bin
endif

# Default target - format, build, test, and security scan
all: fmt build build-all test test-integration test-race check-ctx vet revive gosec
	@echo "✅ All checks passed!"

help:
	@echo "Available targets:"
	@echo "  make                   - Format, build, test, and security scan (default)"
	@echo "  make all               - Same as default"
	@echo "  make build             - Build the dashlights binary"
	@echo "  make build-all         - Build binaries for all release platforms to dist/"
	@echo "  make test              - Run all tests"
	@echo "  make test-integration  - Run integration tests (including performance)"
	@echo "  make test-race         - Run tests with race detector"
	@echo "  make check-ctx         - Check that all signals respect context cancellation"
	@echo "  make gosec             - Run security scanner (gosec) with audit mode"
	@echo "  make coverage          - Run tests with coverage report"
	@echo "  make coverage-html     - Generate HTML coverage report"
	@echo "  make coverage-signals  - Show detailed coverage for signals package"
	@echo "  make fmt               - Format all Go files"
	@echo "  make fmt-check         - Check if files need formatting (CI-friendly)"
	@echo "  make clean             - Remove built binaries"
	@echo "  make install           - Install dashlights to GOPATH/bin"
	@echo "  make hooks             - Install Git hooks from scripts/hooks/"
	@echo "  make load-light        - Generate light system load (Ctrl+C to stop)"
	@echo "  make load-medium       - Generate medium system load (Ctrl+C to stop)"
	@echo "  make load-heavy        - Generate heavy system load (Ctrl+C to stop)"
	@echo "  make stress-test       - Run dashlights repeatedly under load with stats"
	@echo "  make docker-install-test - Run dockerized install test plan"
	@echo "  make install-fabric-pattern - Install Fabric pattern for changelog generation"
	@echo "  make release           - Create a new release with AI-generated changelog"

# Build the binary
build:
	@echo "Generating repository URL..."
	@cd src && go generate main.go
	@echo "Building dashlights..."
	@go build -o dashlights ./src

# Build binaries for all release platforms
build-all:
	@echo "Generating repository URL..."
	@cd src && go generate main.go
	@echo "Building binaries for all release platforms..."
	@mkdir -p dist
	@echo "  Building linux/amd64..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/dashlights-linux-amd64 ./src
	@echo "  Building linux/arm64..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o dist/dashlights-linux-arm64 ./src
	@echo "  Building linux/arm (v6)..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=6 go build -o dist/dashlights-linux-armv6 ./src
	@echo "  Building linux/arm (v7)..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -o dist/dashlights-linux-armv7 ./src
	@echo "  Building darwin/amd64..."
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o dist/dashlights-darwin-amd64 ./src
	@echo "  Building darwin/arm64..."
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o dist/dashlights-darwin-arm64 ./src
	@echo "✅ Built 7 binaries in dist/"
	@ls -lh dist/

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

# Run go vet on the codebase
vet:
		@echo "Running go vet..."
		@go vet ./...
		@echo "✅ go vet passed"

# Run revive linter
revive:
		@echo "Running revive linter..."
		@if [ ! -f $(GOBIN)/revive ]; then \
				echo "revive not found. Installing..."; \
				go install github.com/mgechev/revive@latest; \
			fi
		@$(GOBIN)/revive ./...
		@echo "✅ Revive passed - no issues found!"


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
	@echo "  Run 'make coverage-signals' for detailed signals-only coverage."

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

# Check that all signals respect context cancellation
check-ctx:
	@echo "Checking signals respect context cancellation..."
	@missing=0; \
	checked=0; \
	for file in src/signals/*.go; do \
		if echo "$$file" | grep -q "_test.go"; then continue; fi; \
		if echo "$$file" | grep -q "signal.go"; then continue; fi; \
		if echo "$$file" | grep -q "registry.go"; then continue; fi; \
		if grep -q "func.*Check(ctx context.Context)" "$$file"; then \
			checked=$$((checked + 1)); \
			if ! grep -q "ctx.Done()" "$$file"; then \
				if [ $$missing -eq 0 ]; then \
					echo "⚠️  Signals without ctx.Done() checks:"; \
				fi; \
				basename="$$(basename $$file)"; \
				echo "   - $$basename"; \
				missing=$$((missing + 1)); \
			fi; \
		fi; \
	done; \
	if [ $$missing -eq 0 ]; then \
		echo "✅ All $$checked signals check ctx.Done()"; \
	else \
		with_ctx=$$((checked - missing)); \
		echo ""; \
		echo "Summary: $$with_ctx/$$checked signals check ctx.Done()"; \
		echo ""; \
		echo "Note: Simple signals (env var checks, single file reads) may not need ctx.Done()."; \
		echo "Signals with loops, directory traversal, or file scanning MUST check ctx.Done()."; \
		echo "See CONTRIBUTING.md 'Performance Requirements' for details."; \
	fi

# Clean built binaries
clean:
	@echo "Cleaning..."
	@rm -f dashlights coverage.out coverage.html
	@rm -rf dist/
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

# Install Fabric pattern for changelog generation
install-fabric-pattern:
	@echo "Installing Fabric pattern for changelog generation..."
	@if ! command -v fabric >/dev/null 2>&1; then \
		echo "❌ Error: fabric CLI not found"; \
		echo "Install it from: https://github.com/danielmiessler/fabric"; \
		exit 1; \
	fi
	@mkdir -p ~/.config/fabric/patterns/create_git_changelog
	@cp scripts/fabric-patterns/create_git_changelog/system.md ~/.config/fabric/patterns/create_git_changelog/
	@echo "✅ Fabric pattern installed to ~/.config/fabric/patterns/create_git_changelog/"

# Create a new release with AI-generated changelog
release: docker-install-test
	@bash scripts/release.sh

# Load testing targets
load-light:
	@echo "Starting LIGHT system load (Ctrl+C to stop)..."
	@bash scripts/load-test.sh light

load-medium:
	@echo "Starting MEDIUM system load (Ctrl+C to stop)..."
	@bash scripts/load-test.sh medium

load-heavy:
	@echo "Starting HEAVY system load (Ctrl+C to stop)..."
	@bash scripts/load-test.sh heavy

load: load-light  # Default to light load

# Stress test - run dashlights repeatedly under load
stress-test:
	@if [ ! -f dashlights ]; then \
		echo "Building dashlights first..."; \
		$(MAKE) build; \
	fi
	@echo "Stress testing dashlights..."
	@bash scripts/stress-dashlights.sh

docker-install-test:
	@bash scripts/dockerized-install-test.sh
