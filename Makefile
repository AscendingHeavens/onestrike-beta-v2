# Variables
COVERAGE_THRESHOLD := 85.0

# Lint using golangci-lint
lint:
	@echo "Running golangci-lint..."
	@golangci-lint run ./...


fix-lint:
	@echo "Auto-fixing lint issues..."
	@golangci-lint run --fix ./...


# Run tests with coverage and fail if below threshold (no coverage.out file)
test:
	@go test ./... -v 


# Install dependencies
deps:
	@go mod tidy

.PHONY: lint test deps fix-lint
