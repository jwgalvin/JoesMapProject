# Step 01: Project Setup and Initialization

## Objective
Initialize the Go project structure, set up dependency management, and create the base directory structure.

## Prerequisites
- Go 1.22+ installed
- Git installed
- Code editor (VS Code recommended)

## Tasks

### 1. Initialize Git Repository
```powershell
cd c:\Users\jwgal\JoesMapProject
git init
```

### 2. Create .gitignore
Create `.gitignore` file:
```
# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib
/api

# Test coverage
*.out
coverage.html

# Database files
*.db
*.sqlite
*.sqlite3
/data/

# Environment files
.env
.env.local

# IDE
.vscode/
.idea/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Build artifacts
/bin/
/dist/
```

### 3. Initialize Go Module
```powershell
go mod init github.com/yourusername/geopulse
```

*Replace `yourusername` with your actual GitHub username*

### 4. Create Directory Structure
```powershell
# Create all directories at once
New-Item -ItemType Directory -Path @(
    "cmd\api",
    "internal\domain\event",
    "internal\application\ingest",
    "internal\application\query",
    "internal\infrastructure\usgs",
    "internal\infrastructure\persistence",
    "internal\infrastructure\http",
    "internal\interfaces\api",
    "configs",
    "migrations",
    "scripts",
    "docs",
    "tests\unit",
    "tests\integration"
)
```

### 5. Create Configuration Files

#### Create `configs/.env.example`
```env
# Server Configuration
PORT=8080

# Database Configuration
DATABASE_PATH=./data/geopulse.db

# External APIs
USGS_POLL_INTERVAL=5

# Logging
LOG_LEVEL=info

# CORS
ENABLE_CORS=true
ALLOWED_ORIGINS=http://localhost:3000
```

#### Create `configs/config.yaml`
```yaml
api:
  rate_limit:
    requests_per_minute: 100
  query_limits:
    max_radius_km: 20000
    max_results: 1000
    default_limit: 100

external_apis:
  usgs:
    endpoint: "https://earthquake.usgs.gov/earthquakes/feed/v1.0/summary/all_day.geojson"
    timeout_seconds: 10
```

### 6. Install Core Dependencies
```powershell
# HTTP router
go get -u github.com/gorilla/mux

# Environment variable loader
go get -u github.com/joho/godotenv

# Database driver
go get -u github.com/mattn/go-sqlite3

# YAML configuration
go get -u gopkg.in/yaml.v3

# Testing
go get -u github.com/stretchr/testify
```

### 7. Create Initial README.md
Create `README.md`:
```markdown
# GeoPulse - Go Geospatial Events API

Real-time geospatial event tracking API built with Go.

## Prerequisites
- Go 1.22+
- SQLite3

## Quick Start

1. Clone repository
2. Copy `.env.example` to `.env`
3. Run the application:
   ```
   go run ./cmd/api
   ```

## Development

Run tests:
```
go test ./...
```

## API Documentation

See [docs/api.md](docs/api.md) for API documentation.

## License

MIT
```

### 8. Create Initial main.go Stub
Create `cmd/api/main.go`:
```go
package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	fmt.Println("GeoPulse API starting...")
	
	// Load environment variables
	// Initialize configuration
	// Set up database
	// Start HTTP server
	
	log.Println("Server started on port", os.Getenv("PORT"))
}
```

### 9. Verify Setup
```powershell
# Test that the project compiles
go build ./cmd/api

# Verify module dependencies
go mod tidy

# Run the stub (should print startup message)
go run ./cmd/api
```

## Success Criteria
- ✓ Go module initialized
- ✓ Directory structure created
- ✓ Dependencies installed
- ✓ Configuration files in place
- ✓ Initial main.go compiles and runs
- ✓ Git repository initialized

## Next Step
Proceed to **Step02-DomainModel.md** to implement the core domain entities.
