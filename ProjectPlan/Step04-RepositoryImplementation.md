# Step 04: Repository Implementation

## Objective
Implement the SQLite repository that satisfies the domain repository interface.

## Tasks

### 1. Create SQLite Repository
Create `internal/infrastructure/persistence/sqlite_repository.go`:
```go
package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/yourusername/geopulse/internal/domain/event"
	"github.com/google/uuid"
)

// SQLiteRepository implements event.Repository using SQLite
type SQLiteRepository struct {
	db *Database
}

// NewSQLiteRepository creates a new SQLite event repository
func NewSQLiteRepository(db *Database) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

// Save inserts a new event or updates existing one
func (r *SQLiteRepository) Save(ctx context.Context, evt *event.Event) error {
	// Generate ID if not set
	if evt.ID == "" {
		evt.ID = uuid.New().String()
	}

	// Marshal metadata to JSON
	metadataJSON, err := json.Marshal(evt.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO events (
			id, event_type, source, source_event_id, occurred_at,
			latitude, longitude, magnitude, depth, metadata,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(source, source_event_id) DO UPDATE SET
			event_type = excluded.event_type,
			occurred_at = excluded.occurred_at,
			latitude = excluded.latitude,
			longitude = excluded.longitude,
			magnitude = excluded.magnitude,
			depth = excluded.depth,
			metadata = excluded.metadata,
			updated_at = excluded.updated_at
	`

	_, err = r.db.ExecContext(ctx, query,
		evt.ID,
		evt.EventType.String(),
		evt.Source,
		evt.SourceEventID,
		evt.OccurredAt,
		evt.Location.Latitude,
		evt.Location.Longitude,
		evt.Magnitude.Float64(),
		evt.Depth,
		string(metadataJSON),
		evt.CreatedAt,
		evt.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save event: %w", err)
	}

	return nil
}

// FindByID retrieves an event by its ID
func (r *SQLiteRepository) FindByID(ctx context.Context, id string) (*event.Event, error) {
	query := `
		SELECT id, event_type, source, source_event_id, occurred_at,
		       latitude, longitude, magnitude, depth, metadata,
		       created_at, updated_at
		FROM events
		WHERE id = ?
	`

	var evt event.Event
	var eventTypeStr string
	var metadataJSON string
	var depth sql.NullFloat64

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&evt.ID,
		&eventTypeStr,
		&evt.Source,
		&evt.SourceEventID,
		&evt.OccurredAt,
		&evt.Location.Latitude,
		&evt.Location.Longitude,
		&evt.Magnitude.Value,
		&depth,
		&metadataJSON,
		&evt.CreatedAt,
		&evt.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("event not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query event: %w", err)
	}

	// Parse event type
	evt.EventType, _ = event.ParseEventType(eventTypeStr)

	// Parse depth
	if depth.Valid {
		evt.Depth = &depth.Float64
	}

	// Parse metadata
	if err := json.Unmarshal([]byte(metadataJSON), &evt.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &evt, nil
}

// FindBySourceEventID retrieves an event by source and source event ID
func (r *SQLiteRepository) FindBySourceEventID(ctx context.Context, source, sourceEventID string) (*event.Event, error) {
	query := `
		SELECT id, event_type, source, source_event_id, occurred_at,
		       latitude, longitude, magnitude, depth, metadata,
		       created_at, updated_at
		FROM events
		WHERE source = ? AND source_event_id = ?
	`

	var evt event.Event
	var eventTypeStr string
	var metadataJSON string
	var depth sql.NullFloat64

	err := r.db.QueryRowContext(ctx, query, source, sourceEventID).Scan(
		&evt.ID,
		&eventTypeStr,
		&evt.Source,
		&evt.SourceEventID,
		&evt.OccurredAt,
		&evt.Location.Latitude,
		&evt.Location.Longitude,
		&evt.Magnitude.Value,
		&depth,
		&metadataJSON,
		&evt.CreatedAt,
		&evt.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Not found is not an error in this case
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query event: %w", err)
	}

	evt.EventType, _ = event.ParseEventType(eventTypeStr)
	
	if depth.Valid {
		evt.Depth = &depth.Float64
	}

	if err := json.Unmarshal([]byte(metadataJSON), &evt.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &evt, nil
}

// Query retrieves events matching the given criteria
func (r *SQLiteRepository) Query(ctx context.Context, criteria event.QueryCriteria) ([]*event.Event, error) {
	query, args := r.buildQuery(criteria, false)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	return r.scanEvents(rows)
}

// Count returns the total number of events matching the criteria
func (r *SQLiteRepository) Count(ctx context.Context, criteria event.QueryCriteria) (int, error) {
	query, args := r.buildQuery(criteria, true)

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count events: %w", err)
	}

	return count, nil
}

// buildQuery constructs SQL query based on criteria
func (r *SQLiteRepository) buildQuery(criteria event.QueryCriteria, isCount bool) (string, []interface{}) {
	var sb strings.Builder
	var args []interface{}

	if isCount {
		sb.WriteString("SELECT COUNT(*) FROM events WHERE 1=1")
	} else {
		sb.WriteString(`
			SELECT id, event_type, source, source_event_id, occurred_at,
			       latitude, longitude, magnitude, depth, metadata,
			       created_at, updated_at
			FROM events WHERE 1=1
		`)
	}

	// Magnitude filters
	if criteria.MinMagnitude != nil {
		sb.WriteString(" AND magnitude >= ?")
		args = append(args, *criteria.MinMagnitude)
	}
	if criteria.MaxMagnitude != nil {
		sb.WriteString(" AND magnitude <= ?")
		args = append(args, *criteria.MaxMagnitude)
	}

	// Event type filter
	if len(criteria.EventTypes) > 0 {
		placeholders := make([]string, len(criteria.EventTypes))
		for i, et := range criteria.EventTypes {
			placeholders[i] = "?"
			args = append(args, et.String())
		}
		sb.WriteString(fmt.Sprintf(" AND event_type IN (%s)", strings.Join(placeholders, ",")))
	}

	// Time range filters
	if criteria.FromTime != nil {
		sb.WriteString(" AND occurred_at >= ?")
		args = append(args, *criteria.FromTime)
	}
	if criteria.ToTime != nil {
		sb.WriteString(" AND occurred_at <= ?")
		args = append(args, *criteria.ToTime)
	}

	// Spatial filter (simple bounding box for SQLite)
	if criteria.Location != nil && criteria.RadiusKm != nil {
		// Calculate rough bounding box (1 degree ≈ 111 km)
		degreeOffset := *criteria.RadiusKm / 111.0
		
		sb.WriteString(" AND latitude BETWEEN ? AND ?")
		args = append(args, 
			criteria.Location.Latitude-degreeOffset,
			criteria.Location.Latitude+degreeOffset,
		)
		
		sb.WriteString(" AND longitude BETWEEN ? AND ?")
		args = append(args, 
			criteria.Location.Longitude-degreeOffset,
			criteria.Location.Longitude+degreeOffset,
		)
	}

	if !isCount {
		sb.WriteString(" ORDER BY occurred_at DESC")
		sb.WriteString(" LIMIT ? OFFSET ?")
		args = append(args, criteria.Limit, criteria.Offset)
	}

	return sb.String(), args
}

// scanEvents scans database rows into Event objects
func (r *SQLiteRepository) scanEvents(rows *sql.Rows) ([]*event.Event, error) {
	var events []*event.Event

	for rows.Next() {
		var evt event.Event
		var eventTypeStr string
		var metadataJSON string
		var depth sql.NullFloat64

		err := rows.Scan(
			&evt.ID,
			&eventTypeStr,
			&evt.Source,
			&evt.SourceEventID,
			&evt.OccurredAt,
			&evt.Location.Latitude,
			&evt.Location.Longitude,
			&evt.Magnitude.Value,
			&depth,
			&metadataJSON,
			&evt.CreatedAt,
			&evt.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}

		evt.EventType, _ = event.ParseEventType(eventTypeStr)
		
		if depth.Valid {
			evt.Depth = &depth.Float64
		}

		if err := json.Unmarshal([]byte(metadataJSON), &evt.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		events = append(events, &evt)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return events, nil
}
```

*Note: Update the import path to match your Go module name*

### 2. Install UUID Package
```powershell
go get -u github.com/google/uuid
```

### 3. Create Repository Tests
Create `internal/infrastructure/persistence/sqlite_repository_test.go`:
```go
package persistence

import (
	"context"
	"testing"
	"time"

	"github.com/yourusername/geopulse/internal/domain/event"
	"github.com/yourusername/geopulse/tests/testutil"
)

func TestSQLiteRepository_Save(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	repo := NewSQLiteRepository(&Database{db})
	ctx := context.Background()

	// Create test event
	loc, _ := event.NewLocation(37.7749, -122.4194)
	mag, _ := event.NewMagnitude(5.5)
	evt, _ := event.NewEvent(
		"USGS",
		"us1234",
		event.EventTypeEarthquake,
		time.Now(),
		loc,
		mag,
	)

	// Save event
	err := repo.Save(ctx, evt)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Verify ID was generated
	if evt.ID == "" {
		t.Error("Event ID should be generated")
	}

	// Retrieve and verify
	retrieved, err := repo.FindByID(ctx, evt.ID)
	if err != nil {
		t.Fatalf("FindByID() failed: %v", err)
	}

	if retrieved.Source != "USGS" {
		t.Errorf("Expected source USGS, got %s", retrieved.Source)
	}
}

func TestSQLiteRepository_Upsert(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	repo := NewSQLiteRepository(&Database{db})
	ctx := context.Background()

	loc, _ := event.NewLocation(37.7749, -122.4194)
	mag1, _ := event.NewMagnitude(5.5)
	
	evt, _ := event.NewEvent(
		"USGS",
		"us1234",
		event.EventTypeEarthquake,
		time.Now(),
		loc,
		mag1,
	)

	// First save
	repo.Save(ctx, evt)
	firstID := evt.ID

	// Update with same source event ID
	mag2, _ := event.NewMagnitude(6.0)
	evt2, _ := event.NewEvent(
		"USGS",
		"us1234",
		event.EventTypeEarthquake,
		time.Now(),
		loc,
		mag2,
	)

	// Second save should update
	repo.Save(ctx, evt2)

	// Should still find by original ID
	retrieved, err := repo.FindByID(ctx, firstID)
	if err != nil {
		t.Fatalf("FindByID() failed: %v", err)
	}

	// Magnitude should be updated
	if retrieved.Magnitude.Value != 6.0 {
		t.Errorf("Expected updated magnitude 6.0, got %f", retrieved.Magnitude.Value)
	}
}

func TestSQLiteRepository_Query(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	repo := NewSQLiteRepository(&Database{db})
	ctx := context.Background()

	// Insert test events
	events := createTestEvents(t)
	for _, evt := range events {
		repo.Save(ctx, evt)
	}

	// Test magnitude filter
	t.Run("filter by magnitude", func(t *testing.T) {
		minMag := 5.0
		criteria := event.NewQueryCriteria()
		criteria.MinMagnitude = &minMag

		results, err := repo.Query(ctx, criteria)
		if err != nil {
			t.Fatalf("Query() failed: %v", err)
		}

		for _, evt := range results {
			if evt.Magnitude.Value < 5.0 {
				t.Errorf("Expected magnitude >= 5.0, got %f", evt.Magnitude.Value)
			}
		}
	})

	// Test count
	t.Run("count events", func(t *testing.T) {
		criteria := event.NewQueryCriteria()
		count, err := repo.Count(ctx, criteria)
		if err != nil {
			t.Fatalf("Count() failed: %v", err)
		}

		if count != len(events) {
			t.Errorf("Expected count %d, got %d", len(events), count)
		}
	})
}

func createTestEvents(t *testing.T) []*event.Event {
	t.Helper()

	events := make([]*event.Event, 0, 5)
	
	for i := 0; i < 5; i++ {
		loc, _ := event.NewLocation(float64(30+i), float64(-120+i))
		mag, _ := event.NewMagnitude(float64(4 + i))
		
		evt, _ := event.NewEvent(
			"USGS",
			fmt.Sprintf("test%d", i),
			event.EventTypeEarthquake,
			time.Now().Add(-time.Duration(i)*time.Hour),
			loc,
			mag,
		)
		
		events = append(events, evt)
	}
	
	return events
}
```

*Note: Update import paths and add `fmt` import*

### 4. Run Repository Tests
```powershell
go test ./internal/infrastructure/persistence -v
```

## Success Criteria
- ✓ Repository implements domain interface
- ✓ Save/upsert logic works correctly
- ✓ Query filtering implemented
- ✓ Spatial queries work (bounding box)
- ✓ All repository tests pass
- ✓ Proper error handling

## Next Step
Proceed to **Step05-USGSClient.md** to implement the USGS data fetcher.
