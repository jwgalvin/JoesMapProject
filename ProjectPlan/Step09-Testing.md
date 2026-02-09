# Step 09: Comprehensive Testing

## Objective
Add integration tests, improve test coverage, and set up testing infrastructure.

## Tasks

### 1. Create Integration Test Package
Create `tests/integration/api_test.go`:
```go
package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/yourusername/geopulse/internal/application/ingest"
	"github.com/yourusername/geopulse/internal/application/query"
	"github.com/yourusername/geopulse/internal/domain/event"
	"github.com/yourusername/geopulse/internal/infrastructure/persistence"
	httpinfra "github.com/yourusername/geopulse/internal/infrastructure/http"
	"github.com/yourusername/geopulse/internal/infrastructure/usgs"
	"github.com/yourusername/geopulse/internal/interfaces/api"
	"log/slog"
)

// TestFullIngestionAndQuery tests the complete flow
func TestFullIngestionAndQuery(t *testing.T) {
	// Setup
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	repo := persistence.NewSQLiteRepository(&persistence.Database{DB: db})
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create mock USGS client with test data
	mockClient := usgs.NewMockClient()
	
	// Add diverse test events
	addTestEvent(t, mockClient, 5.5, 37.7749, -122.4194) // San Francisco
	addTestEvent(t, mockClient, 4.2, 34.0522, -118.2437) // Los Angeles
	addTestEvent(t, mockClient, 6.0, 35.6762, 139.6503)  // Tokyo
	addTestEvent(t, mockClient, 3.5, 40.7128, -74.0060)  // New York

	// Run ingestion
	ingestService := ingest.NewService(mockClient, repo, logger)
	ctx := context.Background()
	
	result, err := ingestService.IngestEvents(ctx)
	if err != nil {
		t.Fatalf("Ingestion failed: %v", err)
	}

	if result.Created != 4 {
		t.Errorf("Expected 4 events created, got %d", result.Created)
	}

	// Test querying
	queryService := query.NewService(repo)
	handler := httpinfra.NewHandler(queryService)
	router := httpinfra.NewRouter(handler, logger, nil)

	// Test 1: Get all events
	t.Run("list all events", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/events?limit=10", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected 200, got %d", w.Code)
		}

		var response api.EventsResponse
		json.NewDecoder(w.Body).Decode(&response)

		if len(response.Data) != 4 {
			t.Errorf("Expected 4 events, got %d", len(response.Data))
		}

		if response.Pagination.Total != 4 {
			t.Errorf("Expected total 4, got %d", response.Pagination.Total)
		}
	})

	// Test 2: Filter by magnitude
	t.Run("filter by magnitude", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/events?minMagnitude=5.0&limit=10", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var response api.EventsResponse
		json.NewDecoder(w.Body).Decode(&response)

		if len(response.Data) != 2 {
			t.Errorf("Expected 2 events with mag >= 5.0, got %d", len(response.Data))
		}

		for _, evt := range response.Data {
			if evt.Magnitude < 5.0 {
				t.Errorf("Event magnitude %f is below minimum", evt.Magnitude)
			}
		}
	})

	// Test 3: GeoJSON output
	t.Run("geojson output", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/events/geojson?limit=10", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var geoJSON api.GeoJSONFeatureCollection
		json.NewDecoder(w.Body).Decode(&geoJSON)

		if geoJSON.Type != "FeatureCollection" {
			t.Errorf("Expected FeatureCollection, got %s", geoJSON.Type)
		}

		if len(geoJSON.Features) != 4 {
			t.Errorf("Expected 4 features, got %d", len(geoJSON.Features))
		}

		// Verify first feature structure
		feature := geoJSON.Features[0]
		if feature.Type != "Feature" {
			t.Errorf("Expected Feature type, got %s", feature.Type)
		}
		if feature.Geometry.Type != "Point" {
			t.Errorf("Expected Point geometry, got %s", feature.Geometry.Type)
		}
		if len(feature.Geometry.Coordinates) < 2 {
			t.Error("Expected at least lat/lng coordinates")
		}
	})

	// Test 4: Pagination
	t.Run("pagination", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/events?limit=2&offset=0", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var response api.EventsResponse
		json.NewDecoder(w.Body).Decode(&response)

		if len(response.Data) != 2 {
			t.Errorf("Expected 2 events (page 1), got %d", len(response.Data))
		}

		// Second page
		req = httptest.NewRequest("GET", "/v1/events?limit=2&offset=2", nil)
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)

		json.NewDecoder(w.Body).Decode(&response)

		if len(response.Data) != 2 {
			t.Errorf("Expected 2 events (page 2), got %d", len(response.Data))
		}
	})

	// Test 5: Invalid parameters
	t.Run("invalid parameters", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/events?minMagnitude=15", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected 400 Bad Request, got %d", w.Code)
		}

		var errResp api.ErrorResponse
		json.NewDecoder(w.Body).Decode(&errResp)

		if errResp.Error.Code != "INVALID_PARAMETER" {
			t.Errorf("Expected INVALID_PARAMETER error code, got %s", errResp.Error.Code)
		}
	})

	// Test 6: Health check
	t.Run("health check", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200 OK, got %d", w.Code)
		}

		var health map[string]interface{}
		json.NewDecoder(w.Body).Decode(&health)

		if health["status"] != "ok" {
			t.Errorf("Expected status ok, got %v", health["status"])
		}
	})
}

// TestUpsertBehavior tests that duplicate events are updated
func TestUpsertBehavior(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	repo := persistence.NewSQLiteRepository(&persistence.Database{DB: db})
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	mockClient := usgs.NewMockClient()
	
	// Add initial event
	loc, _ := event.NewLocation(37.7749, -122.4194)
	mag1, _ := event.NewMagnitude(5.0)
	evt1, _ := event.NewEvent("USGS", "duplicate-test", event.EventTypeEarthquake, time.Now(), loc, mag1)
	mockClient.Events = []*event.Event{evt1}

	// First ingestion
	ingestService := ingest.NewService(mockClient, repo, logger)
	ctx := context.Background()
	result1, _ := ingestService.IngestEvents(ctx)

	if result1.Created != 1 {
		t.Fatalf("Expected 1 created, got %d", result1.Created)
	}

	// Update event with new magnitude
	mag2, _ := event.NewMagnitude(6.0)
	evt2, _ := event.NewEvent("USGS", "duplicate-test", event.EventTypeEarthquake, time.Now(), loc, mag2)
	mockClient.Events = []*event.Event{evt2}

	// Second ingestion should update
	result2, _ := ingestService.IngestEvents(ctx)

	if result2.Updated != 1 {
		t.Errorf("Expected 1 updated, got %d", result2.Updated)
	}
	if result2.Created != 0 {
		t.Errorf("Expected 0 created on update, got %d", result2.Created)
	}

	// Verify magnitude was updated
	retrieved, _ := repo.FindBySourceEventID(ctx, "USGS", "duplicate-test")
	if retrieved.Magnitude.Value != 6.0 {
		t.Errorf("Expected updated magnitude 6.0, got %f", retrieved.Magnitude.Value)
	}
}

// Helper functions

func setupTestDatabase(t *testing.T) (*persistence.Database, func()) {
	t.Helper()

	dbPath := fmt.Sprintf("./test_%s.db", t.Name())
	
	db, err := persistence.NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Run migrations
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
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.Remove(dbPath)
		os.Remove(dbPath + "-shm")
		os.Remove(dbPath + "-wal")
	}

	return db, cleanup
}

func addTestEvent(t *testing.T, client *usgs.MockClient, magnitude, lat, lng float64) {
	t.Helper()
	client.AddMockEvent(magnitude, lat, lng)
}
```

### 2. Create Test Coverage Script
Create `scripts/test.ps1`:
```powershell
#!/usr/bin/env pwsh

Write-Host "Running Go tests with coverage..." -ForegroundColor Green

# Run all tests with coverage
go test ./... -coverprofile=coverage.out -covermode=atomic

if ($LASTEXITCODE -eq 0) {
    Write-Host "`nTests passed!" -ForegroundColor Green
    
    # Generate HTML coverage report
    go tool cover -html=coverage.out -o coverage.html
    
    Write-Host "`nCoverage report generated: coverage.html" -ForegroundColor Cyan
    
    # Display coverage summary
    go tool cover -func=coverage.out | Select-Object -Last 1
} else {
    Write-Host "`nTests failed!" -ForegroundColor Red
    exit 1
}
```

### 3. Run All Tests
```powershell
# Run all tests
go test ./... -v

# Run with coverage
.\scripts\test.ps1

# Run only integration tests
go test ./tests/integration -v

# Run tests for specific package
go test ./internal/domain/event -v
```

### 4. Create Test Documentation
Create `docs/testing.md`:
```markdown
# Testing Guide

## Running Tests

### All Tests
```bash
go test ./...
```

### With Coverage
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Specific Package
```bash
go test ./internal/domain/event -v
```

### Integration Tests Only
```bash
go test ./tests/integration -v
```

## Test Structure

### Unit Tests
- Located alongside source files (*_test.go)
- Test individual components in isolation
- Mock external dependencies

### Integration Tests
- Located in `tests/integration/`
- Test full request-to-database flow
- Use real SQLite database (in-memory or file)

## Test Coverage Goals

- Domain layer: > 90%
- Application layer: > 80%
- Infrastructure layer: > 70%

## Writing Tests

### Domain Tests
```go
func TestEventValidation(t *testing.T) {
    // Test business rules and invariants
}
```

### Application Tests
```go
func TestIngestService(t *testing.T) {
    // Mock external dependencies
    mockClient := usgs.NewMockClient()
    // Test service logic
}
```

### Integration Tests
```go
func TestFullAPIFlow(t *testing.T) {
    // Setup real database
    // Test complete request flow
    // Verify persistence
}
```

## Continuous Integration

Tests run automatically on every push via GitHub Actions.
```

## Success Criteria
- ✓ Integration tests implemented
- ✓ Test coverage > 70% overall
- ✓ All tests pass
- ✓ Coverage report generated
- ✓ Test documentation created

## Next Step
Proceed to **Step10-CICDSetup.md** to configure GitHub Actions.
