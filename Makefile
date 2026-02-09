.PHONY: help build run test test-coverage lint fmt vet clean install-tools migrate-up migrate-down docker-build docker-run deps tidy pre-commit-install pre-commit-run

# Variables
BINARY_NAME=geopulse
MAIN_PATH=./cmd/api
GO=go
GOFLAGS=-v
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html
GOLANGCI_LINT_VERSION=v1.61.0

# Default target
help:
	@echo "GeoPulse Makefile Commands:"
	@echo "  make build          - Build the application"
	@echo "  make run            - Run the application"
	@echo "  make test           - Run all tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make lint           - Run golangci-lint"
	@echo "  make fmt            - Format code with gofmt"
	@echo "  make vet            - Run go vet"
	@echo "  make clean          - Remove build artifacts"
	@echo "  make install-tools  - Install development tools (golang-migrate)"
	@echo "  make migrate-up     - Apply database migrations"
	@echo "  make migrate-down   - Rollback database migrations"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make docker-run     - Run Docker container"
	@echo "  make deps           - Download dependencies"
	@echo "  make tidy           - Tidy and verify dependencies"
	@echo "  make pre-commit-install - Install pre-commit hooks (requires Python)"
	@echo "  make pre-commit-run     - Run pre-commit checks (requires Python)"

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) $(MAIN_PATH)

# Run the application
run:
	@echo "Running $(BINARY_NAME)..."
	$(GO) run $(MAIN_PATH)

# Run all tests
test:
	@echo "Running tests..."
	$(GO) test ./... -v

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test ./... -coverprofile=$(COVERAGE_FILE) -covermode=atomic
	$(GO) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report generated: $(COVERAGE_HTML)"
	$(GO) tool cover -func=$(COVERAGE_FILE) | grep total

# Run linter
lint:
	@echo "Running golangci-lint..."
	golangci-lint run --timeout=5m

# Format code
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	$(GO) vet ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe
	rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	rm -f *.db *.db-shm *.db-wal
	rm -rf data/

# Install development tools
install-tools:
	@echo "Installing development tools..."
	@echo "Installing golang-migrate..."
	$(GO) install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "Tools installed successfully!"
	@echo "Note: golangci-lint must be installed separately - see https://golangci-lint.run/welcome/install/"

# Install pre-commit hooks (requires Python and pip)
pre-commit-install:
	@echo "Installing pre-commit..."
	pip install pre-commit || python -m pip install pre-commit
	@echo "Installing pre-commit hooks..."
	pre-commit install
	@echo "Pre-commit hooks installed!"

# Run pre-commit on all files (requires pre-commit-install first)
pre-commit-run:
	@echo "Running pre-commit checks on all files..."
	pre-commit run --all-files
	$(GO) install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "Tools installed successfully!"

# Apply database migrations
migrate-up:
	@echo "Applying database migrations..."
	@if not exist "data" mkdir data
	migrate -path migrations -database "sqlite3://data/geopulse.db" up

# Rollback database migrations
migrate-down:
	@echo "Rolling back database migrations..."
	migrate -path migrations -database "sqlite3://data/geopulse.db" down 1

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):latest .

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker-compose up

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GO) mod download

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	$(GO) mod tidy
	$(GO) mod verify

# Run all checks (format, vet, lint, test)
check: fmt vet lint test
	@echo "All checks passed!"

# CI pipeline simulation
ci: deps tidy check test-coverage
	@echo "CI pipeline completed!"
