# Sudoku Solver with Dancing Links - Makefile

.PHONY: all build test bench clean run-demo run-basic help

# Default target
all: test build

# Build all binaries
build:
	@echo "Building binaries..."
	go build -o bin/dancing_links_demo ./cmd/dancing_links_demo
	@echo "Built: bin/dancing_links_demo"

# Run all tests
test:
	@echo "Running all tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run Dancing Links specific tests
test-dl:
	@echo "Running Dancing Links tests..."
	go test -v ./internal/solver -run TestDancingLinks

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./internal/solver

# Run Dancing Links benchmarks only
bench-dl:
	@echo "Running Dancing Links benchmarks..."
	go test -bench=BenchmarkDancingLinks -benchmem ./internal/solver

# Run the interactive demo
run-demo:
	@echo "Running Dancing Links demonstration..."
	go run ./cmd/dancing_links_demo

# Run the basic example
run-basic:
	@echo "Running basic Dancing Links example..."
	go run ./examples/dancing_links_basic.go

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	@echo "Running linter..."
	golangci-lint run

# Vet code
vet:
	@echo "Running go vet..."
	go vet ./...

# Check for security issues (requires gosec)
security:
	@echo "Checking for security issues..."
	gosec ./...

# Run all quality checks
quality: fmt vet lint security
	@echo "All quality checks completed"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean ./...

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Update dependencies
update-deps:
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

# Create directory structure for builds
init:
	@echo "Creating directory structure..."
	mkdir -p bin/

# Profile Dancing Links performance
profile:
	@echo "Profiling Dancing Links performance..."
	go test -cpuprofile=cpu.prof -memprofile=mem.prof -bench=BenchmarkDancingLinks ./internal/solver
	@echo "CPU profile: cpu.prof"
	@echo "Memory profile: mem.prof"
	@echo "View with: go tool pprof cpu.prof"

# Generate documentation
docs:
	@echo "Generating documentation..."
	godoc -http=:6060 &
	@echo "Documentation server started at http://localhost:6060"

# Quick development cycle
dev: fmt vet test
	@echo "Development cycle completed"

# Release build (optimized)
release: clean init
	@echo "Building release binaries..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o bin/dancing_links_demo-linux-amd64 ./cmd/dancing_links_demo
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-w -s" -o bin/dancing_links_demo-darwin-amd64 ./cmd/dancing_links_demo
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-w -s" -o bin/dancing_links_demo-darwin-arm64 ./cmd/dancing_links_demo
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o bin/dancing_links_demo-windows-amd64.exe ./cmd/dancing_links_demo
	@echo "Release binaries built in bin/"

# Performance comparison
perf-compare:
	@echo "Running performance comparison..."
	@echo "Creating test puzzle..."
	@echo "Testing traditional solver vs Dancing Links..."
	go run ./examples/dancing_links_basic.go

# Memory usage analysis
memory-analysis:
	@echo "Analyzing memory usage..."
	go test -memprofile=mem.prof -bench=BenchmarkDancingLinksCreation ./internal/solver
	go tool pprof -http=:8080 mem.prof &
	@echo "Memory analysis available at http://localhost:8080"

# Help target
help:
	@echo "Available targets:"
	@echo "  all            - Run tests and build (default)"
	@echo "  build          - Build all binaries"
	@echo "  test           - Run all tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  test-dl        - Run Dancing Links tests only"
	@echo "  bench          - Run all benchmarks"
	@echo "  bench-dl       - Run Dancing Links benchmarks only"
	@echo "  run-demo       - Run interactive demonstration"
	@echo "  run-basic      - Run basic example"
	@echo "  fmt            - Format code"
	@echo "  lint           - Run linter (requires golangci-lint)"
	@echo "  vet            - Run go vet"
	@echo "  security       - Check for security issues (requires gosec)"
	@echo "  quality        - Run all quality checks"
	@echo "  clean          - Clean build artifacts"
	@echo "  deps           - Install dependencies"
	@echo "  update-deps    - Update dependencies"
	@echo "  profile        - Profile performance"
	@echo "  docs           - Start documentation server"
	@echo "  dev            - Quick development cycle (fmt, vet, test)"
	@echo "  release        - Build optimized release binaries"
	@echo "  perf-compare   - Compare solver performance"
	@echo "  memory-analysis- Analyze memory usage"
	@echo "  help           - Show this help message"
