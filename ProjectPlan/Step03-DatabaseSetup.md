# Step 03: Database Setup and Migrations

## Objective
Set up SQLite database schema with proper migrations and indexing.

## Tasks

### 1. Install Migration Tool
```powershell
go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Verify installation:
```powershell
migrate -version
```

### 2. Create Initial Migration
Create `migrations/001_create_events_table.up.sql`:
```sql
CREATE TABLE IF NOT EXISTS events (
    id TEXT PRIMARY KEY,
    event_type TEXT NOT NULL,
    source TEXT NOT NULL,
    source_event_id TEXT NOT NULL,
    occurred_at TIMESTAMP NOT NULL,
    latitude REAL NOT NULL,
    longitude REAL NOT NULL,
    magnitude REAL NOT NULL,
    depth REAL,
    metadata TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(source, source_event_id)
);
```

Create `migrations/001_create_events_table.down.sql`:
```sql
DROP TABLE IF EXISTS events;
```

### 3. Create Index Migration
Create `migrations/002_add_indexes.up.sql`:
```sql
-- Index for time-range queries
CREATE INDEX IF NOT EXISTS idx_events_occurred_at ON events(occurred_at);

-- Index for magnitude filtering
CREATE INDEX IF NOT EXISTS idx_events_magnitude ON events(magnitude);

-- Composite index for spatial queries
CREATE INDEX IF NOT EXISTS idx_events_location ON events(latitude, longitude);

-- Index for event type filtering
CREATE INDEX IF NOT EXISTS idx_events_type ON events(event_type);

-- Index for source queries
CREATE INDEX IF NOT EXISTS idx_events_source ON events(source);
```

Create `migrations/002_add_indexes.down.sql`:
```sql
DROP INDEX IF EXISTS idx_events_occurred_at;
DROP INDEX IF EXISTS idx_events_magnitude;
DROP INDEX IF EXISTS idx_events_location;
DROP INDEX IF EXISTS idx_events_type;
DROP INDEX IF EXISTS idx_events_source;
```

### 4. Create Migration Helper Script
Create `scripts/migrate.sh`:
```bash
#!/bin/bash

DB_PATH=${DATABASE_PATH:-./data/geopulse.db}
MIGRATIONS_PATH="file://migrations"

case "$1" in
    up)
        migrate -path $MIGRATIONS_PATH -database "sqlite3://${DB_PATH}" up
        ;;
    down)
        migrate -path $MIGRATIONS_PATH -database "sqlite3://${DB_PATH}" down
        ;;
    force)
        migrate -path $MIGRATIONS_PATH -database "sqlite3://${DB_PATH}" force $2
        ;;
    version)
        migrate -path $MIGRATIONS_PATH -database "sqlite3://${DB_PATH}" version
        ;;
    *)
        echo "Usage: $0 {up|down|force VERSION|version}"
        exit 1
        ;;
esac
```

Create PowerShell version `scripts/migrate.ps1`:
```powershell
param(
    [Parameter(Mandatory=$true)]
    [ValidateSet("up", "down", "force", "version")]
    [string]$Action,
    
    [string]$Version
)

$DB_PATH = if ($env:DATABASE_PATH) { $env:DATABASE_PATH } else { "./data/geopulse.db" }
$MIGRATIONS_PATH = "file://migrations"

switch ($Action) {
    "up" {
        migrate -path $MIGRATIONS_PATH -database "sqlite3://${DB_PATH}" up
    }
    "down" {
        migrate -path $MIGRATIONS_PATH -database "sqlite3://${DB_PATH}" down
    }
    "force" {
        if (-not $Version) {
            Write-Error "Version required for force action"
            exit 1
        }
        migrate -path $MIGRATIONS_PATH -database "sqlite3://${DB_PATH}" force $Version
    }
    "version" {
        migrate -path $MIGRATIONS_PATH -database "sqlite3://${DB_PATH}" version
    }
}
```

### 5. Create Database Package
Create `internal/infrastructure/persistence/database.go`:
```go
package persistence

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// Database wraps the SQL database connection
type Database struct {
	*sql.DB
}

// NewDatabase creates a new database connection
func NewDatabase(dbPath string) (*Database, error) {
	// Ensure data directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(1) // SQLite works best with single connection
	db.SetMaxIdleConns(1)

	// Enable foreign keys and WAL mode for better concurrency
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}
	
	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		return nil, fmt.Errorf("failed to set WAL mode: %w", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Database{db}, nil
}

// Close closes the database connection
func (db *Database) Close() error {
	return db.DB.Close()
}
```

### 6. Create Database Tests
Create `internal/infrastructure/persistence/database_test.go`:
```go
package persistence

import (
	"os"
	"testing"
)

func TestNewDatabase(t *testing.T) {
	// Use temporary database
	dbPath := "./test_geopulse.db"
	defer os.Remove(dbPath)
	defer os.Remove(dbPath + "-shm")
	defer os.Remove(dbPath + "-wal")

	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("NewDatabase() failed: %v", err)
	}
	defer db.Close()

	// Verify connection
	if err := db.Ping(); err != nil {
		t.Errorf("Database ping failed: %v", err)
	}

	// Verify WAL mode is enabled
	var journalMode string
	err = db.QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	if err != nil {
		t.Fatalf("Failed to query journal mode: %v", err)
	}
	if journalMode != "wal" {
		t.Errorf("Expected WAL mode, got %s", journalMode)
	}
}
```

### 7. Run Initial Migration
```powershell
# Create data directory
New-Item -ItemType Directory -Force -Path data

# Run migrations
.\scripts\migrate.ps1 -Action up

# Verify migration
.\scripts\migrate.ps1 -Action version
```

### 8. Create Test Helper for Migrations
Create `tests/testutil/database.go`:
```go
package testutil

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// SetupTestDB creates a temporary test database with schema
func SetupTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	// Create temporary database file
	dbPath := fmt.Sprintf("./test_%s.db", t.Name())
	
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}

	// Create schema (copy from migration)
	schema := `
		CREATE TABLE events (
			id TEXT PRIMARY KEY,
			event_type TEXT NOT NULL,
			source TEXT NOT NULL,
			source_event_id TEXT NOT NULL,
			occurred_at TIMESTAMP NOT NULL,
			latitude REAL NOT NULL,
			longitude REAL NOT NULL,
			magnitude REAL NOT NULL,
			depth REAL,
			metadata TEXT,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(source, source_event_id)
		);
		
		CREATE INDEX idx_events_occurred_at ON events(occurred_at);
		CREATE INDEX idx_events_magnitude ON events(magnitude);
		CREATE INDEX idx_events_location ON events(latitude, longitude);
		CREATE INDEX idx_events_type ON events(event_type);
		CREATE INDEX idx_events_source ON events(source);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create test schema: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		db.Close()
		os.Remove(dbPath)
		os.Remove(dbPath + "-shm")
		os.Remove(dbPath + "-wal")
	}

	return db, cleanup
}
```

### 9. Test Database Setup
```powershell
go test ./internal/infrastructure/persistence -v
```

## Success Criteria
- ✓ Migration files created
- ✓ Migration scripts work (up/down)
- ✓ Database connection package implemented
- ✓ WAL mode enabled for concurrency
- ✓ Indexes created for query performance
- ✓ Test utilities created
- ✓ All tests pass

## Next Step
Proceed to **Step04-RepositoryImplementation.md** to implement the SQLite repository.
