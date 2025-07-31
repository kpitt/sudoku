# Sudoku Solver with Dancing Links - Makefile

.PHONY: all build test bench clean help

# Default target
all: test build

# Build all binaries
build:
	@echo "Building binaries..."
	go build -o bin/sudoku ./cmd/sudoku
	@echo "Built: bin/sudoku"

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
	rm -f *.prof
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
	@echo "  memory-analysis- Analyze memory usage"
	@echo "  help           - Show this help message"
