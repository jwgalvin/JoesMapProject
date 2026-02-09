# GeoPulse - Go Geospatial Events API

Real-time geospatial event tracking API built with Go, providing access to geospatial event data with filtering, queries, and RESTful endpoints.

## Features
- USGS earthquake data ingestion
- RESTful API with filtering and queries
- SQLite database with migrations
- Comprehensive test coverage
- Docker support

## Prerequisites
- **Go 1.22+** (tested with Go 1.25)
- **make** (GNU Make)
- **SQLite3** (embedded, no separate install needed)
- **Python 3.x** (optional, for pre-commit hooks)

### Windows Setup
```powershell
# Install make if not already installed
winget install GnuWin32.Make

# Add to PATH for current session
$env:Path += ";C:\Program Files (x86)\GnuWin32\bin"
```

## Quick Start

### 1. Environment Setup
```powershell
# Copy environment template
cp configs/.env.example configs/.env

# Edit configs/.env with your actual values
```

### 2. Install Dependencies
```powershell
make deps
make install-tools
```

### 3. Run Database Migrations
```powershell
make migrate-up
```

### 4. Run the Application
```powershell
make run
```

The API will be available at `http://localhost:8080`

## Development

### Build the Application
```powershell
make build
```

### Run Tests
```powershell
# Run all tests
make test

# Run tests with coverage report (HTML)
make test-coverage
```

### Code Quality
```powershell
# Format code
make fmt

# Run go vet
make vet

# Run linters (requires golangci-lint)
make lint

# Run all checks (fmt + vet + lint)
make check
```

### Database Migrations
```powershell
# Apply migrations
make migrate-up

# Rollback migrations
make migrate-down
```

### Docker
```powershell
# Build Docker image
make docker-build

# Run container
make docker-run
```

### Available Make Commands
Run `make help` to see all available commands.

## Project Structure
```
.
├── cmd/api/              # Application entrypoint
├── internal/
│   ├── domain/           # Domain models and business logic
│   ├── application/      # Use cases (ingest, query)
│   ├── infrastructure/   # External dependencies (DB, USGS)
│   └── interfaces/       # API handlers
├── configs/              # Configuration files
├── migrations/           # Database migrations
├── tests/                # Integration tests
└── docs/                 # Documentation
```

## Configuration

Configuration is managed through:
- **configs/.env** - Environment variables (local development, secrets)
- **configs/config.yaml** - Application configuration (rate limits, endpoints)

See [configs/.env.example](configs/.env.example) for available environment variables.

## API Documentation

See [docs/api.md](docs/api.md) for full API documentation.

## License

MIT
