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
	@echo "Running tests with coverage..."
	@output=$$(go test ./... -cover | tee /dev/stderr); \
	total=$$(echo "$$output" | grep "coverage:" | awk '{print substr($$2, 1, length($$2)-1)}'); \
	echo "Total coverage is $$total%"; \
	if [ $$(echo "$$total < $(COVERAGE_THRESHOLD)" | bc -l) -eq 1 ]; then \
		echo "Test coverage ($$total%) is below $(COVERAGE_THRESHOLD)% ❌"; \
		exit 1; \
	else \
		echo "Test coverage is OK ✅"; \
	fi

# Install dependencies
deps:
	@go mod tidy

.PHONY: lint test deps fix-lint
