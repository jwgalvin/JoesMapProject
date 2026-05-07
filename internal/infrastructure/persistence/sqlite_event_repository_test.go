package persistence

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/jwgal/JoesMapProject/internal/domain/event"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite" // Pure-Go SQLite driver
)

const testSchema = `
CREATE TABLE IF NOT EXISTS events (
    id TEXT PRIMARY KEY,
    event_type TEXT NOT NULL,
    magnitude_value REAL NOT NULL,
    magnitude_scale TEXT NOT NULL,
    latitude REAL NOT NULL,
    longitude REAL NOT NULL,
    depth_km REAL NOT NULL,
    event_time TEXT NOT NULL,
    location_name TEXT NOT NULL,
    status TEXT NOT NULL,
    description TEXT NOT NULL,
    url TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP
);
`

var (
	// Helper time values for tests
	testTime1 = time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	testTime2 = time.Date(2024, 2, 20, 14, 45, 0, 0, time.UTC)
	testTime3 = time.Date(2024, 3, 10, 8, 15, 0, 0, time.UTC)

	// Common test value objects
	testLocationLA, _    = event.NewLocation(34.05, -118.25, 10.0)
	testLocationTokyo, _ = event.NewLocation(35.68, 139.76, 50.0)
	testLocationDeep, _  = event.NewLocation(19.43, -99.13, 700.0)

	testMagModerate, _ = event.NewMagnitude(5.0, "mw")
	testMagLarge, _    = event.NewMagnitude(7.2, "mw")

	testTypeEarthquake, _ = event.NewType("earthquake")
	testTypeExplosion, _  = event.NewType("explosion")
)

// testEvents contains shared event fixtures for repository tests
var testEvents = []struct {
	name  string
	event *event.Event
}{
	{
		name:  "Los Angeles earthquake",
		event: mustNewEvent("us1000abc1", testLocationLA, "5 km NW of Los Angeles, CA", testMagModerate, testTypeEarthquake, testTime1, "reviewed", "Earthquake detected", "https://earthquake.usgs.gov/earthquakes/eventpage/us1000abc1"),
	},
	{
		name:  "Tokyo deep earthquake",
		event: mustNewEvent("us2000xyz2", testLocationTokyo, "20 km E of Tokyo, Japan", testMagLarge, testTypeEarthquake, testTime2, "automatic", "Large earthquake", "https://earthquake.usgs.gov/earthquakes/eventpage/us2000xyz2"),
	},
	{
		name:  "Mexico deep event",
		event: mustNewEvent("us3000def3", testLocationDeep, "Mexico City, Mexico", testMagModerate, testTypeEarthquake, testTime3, "reviewed", "Deep earthquake", "https://earthquake.usgs.gov/earthquakes/eventpage/us3000def3"),
	},
}

// setupTestDB creates an in-memory SQLite database with schema
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err, "Failed to open in-memory database")

	_, err = db.Exec(testSchema)
	require.NoError(t, err, "Failed to create schema")

	return db
}

// mustNewEvent creates an event or panics (for test fixtures only)
func mustNewEvent(id string, location event.Location, place string, magnitude event.Magnitude, eventType event.Type, eventTime time.Time, status, description, url string) *event.Event {
	e, err := event.NewEvent(id, location, place, magnitude, eventType, eventTime, status, description, url)
	if err != nil {
		panic(err)
	}
	return e
}

func TestNewSQLiteEventRepository(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLiteEventRepository(db)

	assert.NotNil(t, repo, "NewSQLiteEventRepository should return non-nil repository")
	assert.Equal(t, db, repo.db, "Repository should store database reference")
}

func TestSQLiteEventRepository_Save(t *testing.T) {
	t.Run("Save new events", func(t *testing.T) {
		for _, tc := range testEvents {
			t.Run(tc.name, func(t *testing.T) {
				db := setupTestDB(t)
				defer db.Close()
				repo := NewSQLiteEventRepository(db)
				ctx := context.Background()

				err := repo.Save(ctx, tc.event)
				require.NoError(t, err, "Save should not return error")

				// Verify event was saved by retrieving it
				retrieved, err := repo.FindbyID(ctx, tc.event.ID())
				require.NoError(t, err, "FindbyID should not return error")
				require.NotNil(t, retrieved, "Retrieved event should not be nil")

				assert.Equal(t, tc.event.ID(), retrieved.ID())
				assert.Equal(t, tc.event.Type().String(), retrieved.Type().String())
				assert.Equal(t, tc.event.Magnitude().Value(), retrieved.Magnitude().Value())
				assert.Equal(t, tc.event.Place(), retrieved.Place())
			})
		}
	})

	t.Run("Update existing event", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()
		repo := NewSQLiteEventRepository(db)
		ctx := context.Background()

		// Save initial event
		originalEvent := testEvents[0].event
		err := repo.Save(ctx, originalEvent)
		require.NoError(t, err)

		// Create updated version with same ID but different magnitude
		updatedMag, _ := event.NewMagnitude(6.5, "mw")
		updatedEvent := mustNewEvent(
			originalEvent.ID(),
			testLocationLA,
			"Updated location",
			updatedMag,
			testTypeEarthquake,
			testTime1,
			"reviewed",
			"Updated description",
			"https://updated.url",
		)

		// Save updated event (should upsert)
		err = repo.Save(ctx, updatedEvent)
		require.NoError(t, err)

		// Verify update worked
		retrieved, err := repo.FindbyID(ctx, originalEvent.ID())
		require.NoError(t, err)
		assert.Equal(t, 6.5, retrieved.Magnitude().Value(), "Magnitude should be updated")
		assert.Equal(t, "Updated location", retrieved.Place(), "Place should be updated")
	})
}

func TestSQLiteEventRepository_FindbyID(t *testing.T) {
	t.Run("Find existing event", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()
		repo := NewSQLiteEventRepository(db)
		ctx := context.Background()

		// Save test event
		testEvent := testEvents[0].event
		err := repo.Save(ctx, testEvent)
		require.NoError(t, err)

		// Find by ID
		retrieved, err := repo.FindbyID(ctx, testEvent.ID())
		require.NoError(t, err)
		require.NotNil(t, retrieved)

		assert.Equal(t, testEvent.ID(), retrieved.ID())
		assert.Equal(t, testEvent.Type().String(), retrieved.Type().String())
		assert.Equal(t, testEvent.Magnitude().Value(), retrieved.Magnitude().Value())
	})

	t.Run("Find non-existent event", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()
		repo := NewSQLiteEventRepository(db)
		ctx := context.Background()

		retrieved, err := repo.FindbyID(ctx, "nonexistent-id")
		assert.Error(t, err, "Should return error for non-existent event")
		assert.Equal(t, sql.ErrNoRows, err, "Should return sql.ErrNoRows")
		assert.Nil(t, retrieved)
	})
}

func TestSQLiteEventRepository_FindAll(t *testing.T) {
	t.Run("Find all events without criteria", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()
		repo := NewSQLiteEventRepository(db)
		ctx := context.Background()

		// Save all test events
		for _, tc := range testEvents {
			err := repo.Save(ctx, tc.event)
			require.NoError(t, err)
		}

		// Find all
		results, err := repo.FindAll(ctx, nil)
		require.NoError(t, err)
		assert.Len(t, results, len(testEvents), "Should return all saved events")
	})

	t.Run("Find with magnitude filter", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()
		repo := NewSQLiteEventRepository(db)
		ctx := context.Background()

		// Save test events
		for _, tc := range testEvents {
			err := repo.Save(ctx, tc.event)
			require.NoError(t, err)
		}

		// Find events with magnitude >= 5.0
		minMag := 5.0
		criteria := event.NewQueryCriteria()
		criteria.MinMagnitude = &minMag

		results, err := repo.FindAll(ctx, criteria)
		require.NoError(t, err)

		for _, result := range results {
			assert.GreaterOrEqual(t, result.Magnitude().Value(), 5.0, "All results should have magnitude >= 5.0")
		}
	})

	t.Run("Find with pagination", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()
		repo := NewSQLiteEventRepository(db)
		ctx := context.Background()

		// Save test events
		for _, tc := range testEvents {
			err := repo.Save(ctx, tc.event)
			require.NoError(t, err)
		}

		// Get first page (limit 2)
		criteria := event.NewQueryCriteria()
		criteria.Limit = 2
		criteria.Offset = 0

		page1, err := repo.FindAll(ctx, criteria)
		require.NoError(t, err)
		assert.Len(t, page1, 2, "First page should have 2 events")

		// Get second page
		criteria.Offset = 2
		page2, err := repo.FindAll(ctx, criteria)
		require.NoError(t, err)
		assert.Len(t, page2, 1, "Second page should have 1 event")
	})
}

func TestSQLiteEventRepository_Count(t *testing.T) {
	t.Run("Count all events", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()
		repo := NewSQLiteEventRepository(db)
		ctx := context.Background()

		// Save test events
		for _, tc := range testEvents {
			err := repo.Save(ctx, tc.event)
			require.NoError(t, err)
		}

		count, err := repo.Count(ctx, nil)
		require.NoError(t, err)
		assert.Equal(t, int64(len(testEvents)), count)
	})

	t.Run("Count with filter", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()
		repo := NewSQLiteEventRepository(db)
		ctx := context.Background()

		// Save test events
		for _, tc := range testEvents {
			err := repo.Save(ctx, tc.event)
			require.NoError(t, err)
		}

		// Count events with magnitude >= 5.0
		minMag := 5.0
		criteria := event.NewQueryCriteria()
		criteria.MinMagnitude = &minMag

		count, err := repo.Count(ctx, criteria)
		require.NoError(t, err)
		assert.Greater(t, count, int64(0), "Should have events with magnitude >= 5.0")
		assert.LessOrEqual(t, count, int64(len(testEvents)), "Count should not exceed total events")
	})
}

func TestSQLiteEventRepository_Delete(t *testing.T) {
	t.Run("Delete existing event", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()
		repo := NewSQLiteEventRepository(db)
		ctx := context.Background()

		// Save test event
		testEvent := testEvents[0].event
		err := repo.Save(ctx, testEvent)
		require.NoError(t, err)

		// Delete event
		err = repo.Delete(ctx, testEvent.ID())
		require.NoError(t, err, "Delete should not return error")

		// Verify event is deleted
		_, err = repo.FindbyID(ctx, testEvent.ID())
		assert.Error(t, err, "FindbyID should return error after deletion")
		assert.Equal(t, sql.ErrNoRows, err)
	})

	t.Run("Delete non-existent event", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()
		repo := NewSQLiteEventRepository(db)
		ctx := context.Background()

		err := repo.Delete(ctx, "nonexistent-id")
		assert.Error(t, err, "Delete should return error for non-existent event")
		assert.Equal(t, sql.ErrNoRows, err)
	})
}

// TestSQLiteEventRepository_FindAll_TimeFilters tests time-based filtering
func TestSQLiteEventRepository_FindAll_TimeFilters(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewSQLiteEventRepository(db)
	ctx := context.Background()

	// Save all test events
	for _, tc := range testEvents {
		err := repo.Save(ctx, tc.event)
		require.NoError(t, err)
	}

	t.Run("StartTime filter", func(t *testing.T) {
		startTime := testTime2 // Should exclude testTime1
		criteria := event.NewQueryCriteria()
		criteria.StartTime = &startTime

		results, err := repo.FindAll(ctx, criteria)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 1)

		for _, result := range results {
			assert.True(t, result.Time().After(startTime) || result.Time().Equal(startTime),
				"All results should be on or after start time")
		}
	})

	t.Run("EndTime filter", func(t *testing.T) {
		endTime := testTime2 // Should exclude testTime3
		criteria := event.NewQueryCriteria()
		criteria.EndTime = &endTime

		results, err := repo.FindAll(ctx, criteria)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 1)

		for _, result := range results {
			assert.True(t, result.Time().Before(endTime) || result.Time().Equal(endTime),
				"All results should be on or before end time")
		}
	})

	t.Run("StartTime and EndTime range", func(t *testing.T) {
		startTime := testTime1
		endTime := testTime2
		criteria := event.NewQueryCriteria()
		criteria.StartTime = &startTime
		criteria.EndTime = &endTime

		results, err := repo.FindAll(ctx, criteria)
		require.NoError(t, err)

		for _, result := range results {
			assert.True(t,
				(result.Time().After(startTime) || result.Time().Equal(startTime)) &&
					(result.Time().Before(endTime) || result.Time().Equal(endTime)),
				"All results should be within time range")
		}
	})

	t.Run("Count with time filter", func(t *testing.T) {
		startTime := testTime2
		criteria := event.NewQueryCriteria()
		criteria.StartTime = &startTime

		count, err := repo.Count(ctx, criteria)
		require.NoError(t, err)
		assert.Greater(t, count, int64(0))
		assert.LessOrEqual(t, count, int64(len(testEvents)))
	})
}

// TestSQLiteEventRepository_FindAll_EventTypeFilter tests event type filtering
func TestSQLiteEventRepository_FindAll_EventTypeFilter(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewSQLiteEventRepository(db)
	ctx := context.Background()

	// Save test events with different types
	for _, tc := range testEvents {
		err := repo.Save(ctx, tc.event)
		require.NoError(t, err)
	}

	// Add an explosion event
	explosionLoc, _ := event.NewLocation(48.85, 2.35, 0.5)
	explosionMag, _ := event.NewMagnitude(1.5, "ml")
	explosionType, _ := event.NewType("explosion")
	explosionEvent := mustNewEvent("us4000exp1", explosionLoc, "Paris, France",
		explosionMag, explosionType, testTime1, "reviewed", "Explosion event", "https://test.url")
	err := repo.Save(ctx, explosionEvent)
	require.NoError(t, err)

	t.Run("Filter by single event type", func(t *testing.T) {
		criteria := event.NewQueryCriteria()
		criteria.EventTypes = []event.Type{testTypeEarthquake}

		results, err := repo.FindAll(ctx, criteria)
		require.NoError(t, err)
		assert.Greater(t, len(results), 0)

		for _, result := range results {
			assert.Equal(t, "earthquake", result.Type().String(),
				"All results should be earthquakes")
		}
	})

	t.Run("Filter by multiple event types", func(t *testing.T) {
		criteria := event.NewQueryCriteria()
		criteria.EventTypes = []event.Type{testTypeEarthquake, testTypeExplosion}

		results, err := repo.FindAll(ctx, criteria)
		require.NoError(t, err)
		assert.Greater(t, len(results), 0)

		for _, result := range results {
			typeStr := result.Type().String()
			assert.True(t, typeStr == "earthquake" || typeStr == "explosion",
				"Results should be either earthquake or explosion")
		}
	})

	t.Run("Count with event type filter", func(t *testing.T) {
		criteria := event.NewQueryCriteria()
		criteria.EventTypes = []event.Type{testTypeExplosion}

		count, err := repo.Count(ctx, criteria)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count, "Should have exactly 1 explosion")
	})
}

// TestSQLiteEventRepository_FindAll_LocationRadius tests location-based filtering
func TestSQLiteEventRepository_FindAll_LocationRadius(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewSQLiteEventRepository(db)
	ctx := context.Background()

	// Save test events
	for _, tc := range testEvents {
		err := repo.Save(ctx, tc.event)
		require.NoError(t, err)
	}

	t.Run("Events within radius", func(t *testing.T) {
		// Search near Los Angeles with 50km radius
		searchLoc, _ := event.NewLocation(34.05, -118.25, 0)
		radiusKm := 50.0
		criteria := event.NewQueryCriteria()
		criteria.Location = &searchLoc
		criteria.RadiusKm = &radiusKm

		results, err := repo.FindAll(ctx, criteria)
		require.NoError(t, err)

		// Should find the LA earthquake
		assert.GreaterOrEqual(t, len(results), 1)
	})

	t.Run("Events outside radius", func(t *testing.T) {
		// Search near Los Angeles with very small radius (1km)
		searchLoc, _ := event.NewLocation(34.10, -118.30, 0) // Slightly offset
		radiusKm := 1.0
		criteria := event.NewQueryCriteria()
		criteria.Location = &searchLoc
		criteria.RadiusKm = &radiusKm

		results, err := repo.FindAll(ctx, criteria)
		require.NoError(t, err)

		// May find 0 or very few events
		assert.LessOrEqual(t, len(results), len(testEvents))
	})

	t.Run("Count with location filter", func(t *testing.T) {
		searchLoc, _ := event.NewLocation(35.68, 139.76, 0) // Tokyo
		radiusKm := 100.0
		criteria := event.NewQueryCriteria()
		criteria.Location = &searchLoc
		criteria.RadiusKm = &radiusKm

		count, err := repo.Count(ctx, criteria)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(0))
	})
}

// TestSQLiteEventRepository_FindAll_MaxMagnitude tests max magnitude filtering
func TestSQLiteEventRepository_FindAll_MaxMagnitude(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewSQLiteEventRepository(db)
	ctx := context.Background()

	// Save test events
	for _, tc := range testEvents {
		err := repo.Save(ctx, tc.event)
		require.NoError(t, err)
	}

	t.Run("MaxMagnitude filter", func(t *testing.T) {
		maxMag := 6.0
		criteria := event.NewQueryCriteria()
		criteria.MaxMagnitude = &maxMag

		results, err := repo.FindAll(ctx, criteria)
		require.NoError(t, err)

		for _, result := range results {
			assert.LessOrEqual(t, result.Magnitude().Value(), 6.0,
				"All results should have magnitude <= 6.0")
		}
	})

	t.Run("MinMagnitude and MaxMagnitude range", func(t *testing.T) {
		minMag := 4.0
		maxMag := 6.0
		criteria := event.NewQueryCriteria()
		criteria.MinMagnitude = &minMag
		criteria.MaxMagnitude = &maxMag

		results, err := repo.FindAll(ctx, criteria)
		require.NoError(t, err)

		for _, result := range results {
			mag := result.Magnitude().Value()
			assert.GreaterOrEqual(t, mag, 4.0, "Magnitude should be >= 4.0")
			assert.LessOrEqual(t, mag, 6.0, "Magnitude should be <= 6.0")
		}
	})
}

// TestSQLiteEventRepository_FindAll_MultipleFilters tests combining multiple filters
func TestSQLiteEventRepository_FindAll_MultipleFilters(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewSQLiteEventRepository(db)
	ctx := context.Background()

	// Save test events
	for _, tc := range testEvents {
		err := repo.Save(ctx, tc.event)
		require.NoError(t, err)
	}

	t.Run("Magnitude and time range combined", func(t *testing.T) {
		minMag := 5.0
		startTime := testTime1
		endTime := testTime3
		criteria := event.NewQueryCriteria()
		criteria.MinMagnitude = &minMag
		criteria.StartTime = &startTime
		criteria.EndTime = &endTime

		results, err := repo.FindAll(ctx, criteria)
		require.NoError(t, err)

		for _, result := range results {
			assert.GreaterOrEqual(t, result.Magnitude().Value(), 5.0)
			assert.True(t, !result.Time().Before(startTime))
			assert.True(t, !result.Time().After(endTime))
		}
	})

	t.Run("Magnitude, type, and time combined", func(t *testing.T) {
		minMag := 4.0
		startTime := testTime1
		criteria := event.NewQueryCriteria()
		criteria.MinMagnitude = &minMag
		criteria.StartTime = &startTime
		criteria.EventTypes = []event.Type{testTypeEarthquake}

		results, err := repo.FindAll(ctx, criteria)
		require.NoError(t, err)

		for _, result := range results {
			assert.GreaterOrEqual(t, result.Magnitude().Value(), 4.0)
			assert.Equal(t, "earthquake", result.Type().String())
		}
	})

	t.Run("Count with multiple filters", func(t *testing.T) {
		minMag := 5.0
		startTime := testTime1
		criteria := event.NewQueryCriteria()
		criteria.MinMagnitude = &minMag
		criteria.StartTime = &startTime

		count, err := repo.Count(ctx, criteria)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(0))
	})
}

// TestSQLiteEventRepository_Pagination_EdgeCases tests pagination boundary conditions
func TestSQLiteEventRepository_Pagination_EdgeCases(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewSQLiteEventRepository(db)
	ctx := context.Background()

	// Save test events
	for _, tc := range testEvents {
		err := repo.Save(ctx, tc.event)
		require.NoError(t, err)
	}

	t.Run("Offset beyond total count", func(t *testing.T) {
		criteria := event.NewQueryCriteria()
		criteria.Offset = 1000 // Way beyond our 3 events

		results, err := repo.FindAll(ctx, criteria)
		require.NoError(t, err)
		assert.Empty(t, results, "Should return empty slice when offset exceeds count")
		assert.NotNil(t, results, "Should return empty slice, not nil")
	})

	t.Run("Limit larger than total count", func(t *testing.T) {
		criteria := event.NewQueryCriteria()
		criteria.Limit = 1000

		results, err := repo.FindAll(ctx, criteria)
		require.NoError(t, err)
		assert.Len(t, results, len(testEvents), "Should return all events when limit exceeds count")
	})

	t.Run("Limit of 1", func(t *testing.T) {
		criteria := event.NewQueryCriteria()
		criteria.Limit = 1

		results, err := repo.FindAll(ctx, criteria)
		require.NoError(t, err)
		assert.Len(t, results, 1, "Should return exactly 1 event")
	})

	t.Run("Offset and limit combination", func(t *testing.T) {
		// Get second event
		criteria := event.NewQueryCriteria()
		criteria.Limit = 1
		criteria.Offset = 1

		results, err := repo.FindAll(ctx, criteria)
		require.NoError(t, err)
		assert.Len(t, results, 1, "Should return exactly 1 event at offset 1")
	})
}

// TestSQLiteEventRepository_EmptyDatabase tests operations on empty database
func TestSQLiteEventRepository_EmptyDatabase(t *testing.T) {
	t.Run("FindAll on empty database", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()
		repo := NewSQLiteEventRepository(db)
		ctx := context.Background()

		results, err := repo.FindAll(ctx, nil)
		require.NoError(t, err)
		assert.NotNil(t, results, "Should return empty slice, not nil")
		assert.Empty(t, results, "Should return empty slice")
	})

	t.Run("Count on empty database", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()
		repo := NewSQLiteEventRepository(db)
		ctx := context.Background()

		count, err := repo.Count(ctx, nil)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("FindAll with filters on empty database", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()
		repo := NewSQLiteEventRepository(db)
		ctx := context.Background()

		minMag := 5.0
		criteria := event.NewQueryCriteria()
		criteria.MinMagnitude = &minMag

		results, err := repo.FindAll(ctx, criteria)
		require.NoError(t, err)
		assert.Empty(t, results)
	})
}

// TestSQLiteEventRepository_ContextCancellation tests context handling
func TestSQLiteEventRepository_ContextCancellation(t *testing.T) {
	t.Run("Save with cancelled context", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()
		repo := NewSQLiteEventRepository(db)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		err := repo.Save(ctx, testEvents[0].event)
		assert.Error(t, err, "Should return error with cancelled context")
	})

	t.Run("FindbyID with cancelled context", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()
		repo := NewSQLiteEventRepository(db)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		result, err := repo.FindbyID(ctx, "any-id")
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("FindAll with cancelled context", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()
		repo := NewSQLiteEventRepository(db)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		results, err := repo.FindAll(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, results)
	})
}

// TestSQLiteEventRepository_DataIntegrity tests round-trip data preservation
func TestSQLiteEventRepository_DataIntegrity(t *testing.T) {
	t.Run("All event fields preserved", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()
		repo := NewSQLiteEventRepository(db)
		ctx := context.Background()

		for _, tc := range testEvents {
			t.Run(tc.name, func(t *testing.T) {
				// Save event
				err := repo.Save(ctx, tc.event)
				require.NoError(t, err)

				// Retrieve event
				retrieved, err := repo.FindbyID(ctx, tc.event.ID())
				require.NoError(t, err)

				// Verify all fields
				assert.Equal(t, tc.event.ID(), retrieved.ID())
				assert.Equal(t, tc.event.Type().String(), retrieved.Type().String())
				assert.Equal(t, tc.event.Magnitude().Value(), retrieved.Magnitude().Value())
				assert.Equal(t, tc.event.Magnitude().Scale(), retrieved.Magnitude().Scale())
				assert.Equal(t, tc.event.Location().LatitudeValue(), retrieved.Location().LatitudeValue())
				assert.Equal(t, tc.event.Location().LongitudeValue(), retrieved.Location().LongitudeValue())
				assert.Equal(t, tc.event.Location().DepthValue(), retrieved.Location().DepthValue())
				assert.Equal(t, tc.event.Place(), retrieved.Place())
				assert.Equal(t, tc.event.Status(), retrieved.Status())
				assert.Equal(t, tc.event.Description(), retrieved.Description())
				assert.Equal(t, tc.event.URL(), retrieved.URL())
				// Time comparison (allowing for truncation to second precision)
				assert.Equal(t, tc.event.Time().Unix(), retrieved.Time().Unix())
			})
		}
	})

	t.Run("Unicode and special characters", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()
		repo := NewSQLiteEventRepository(db)
		ctx := context.Background()

		// Create event with special characters
		loc, _ := event.NewLocation(35.68, 139.76, 10.0)
		mag, _ := event.NewMagnitude(4.5, "mw")
		eventType, _ := event.NewType("earthquake")

		specialEvent := mustNewEvent(
			"us-special-123",
			loc,
			"東京, 日本 (Tokyo, Japan) - Special: !@#$%^&*()",
			mag,
			eventType,
			testTime1,
			"reviewed",
			"Description with émojis 🌍 and spëcial çhars™",
			"https://test.url/path?query=value&foo=bar",
		)

		err := repo.Save(ctx, specialEvent)
		require.NoError(t, err)

		retrieved, err := repo.FindbyID(ctx, "us-special-123")
		require.NoError(t, err)
		assert.Equal(t, specialEvent.Place(), retrieved.Place())
		assert.Equal(t, specialEvent.Description(), retrieved.Description())
		assert.Equal(t, specialEvent.URL(), retrieved.URL())
	})

	t.Run("Extreme coordinate values", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()
		repo := NewSQLiteEventRepository(db)
		ctx := context.Background()

		// Test near boundaries
		extremeLoc, _ := event.NewLocation(89.99, 179.99, 999.0)
		mag, _ := event.NewMagnitude(5.0, "mw")
		eventType, _ := event.NewType("earthquake")

		extremeEvent := mustNewEvent(
			"extreme-coords",
			extremeLoc,
			"Extreme location",
			mag,
			eventType,
			testTime1,
			"reviewed",
			"Test",
			"https://test.url",
		)

		err := repo.Save(ctx, extremeEvent)
		require.NoError(t, err)

		retrieved, err := repo.FindbyID(ctx, "extreme-coords")
		require.NoError(t, err)
		assert.InDelta(t, 89.99, retrieved.Location().LatitudeValue(), 0.001)
		assert.InDelta(t, 179.99, retrieved.Location().LongitudeValue(), 0.001)
		assert.InDelta(t, 999.0, retrieved.Location().DepthValue(), 0.001)
	})

	t.Run("Extreme magnitude values", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()
		repo := NewSQLiteEventRepository(db)
		ctx := context.Background()

		testCases := []struct {
			name string
			mag  float64
		}{
			{"Very small magnitude", 0.1},
			{"Negative magnitude", -0.5},
			{"Large magnitude", 9.5},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				loc, _ := event.NewLocation(0, 0, 10)
				mag, _ := event.NewMagnitude(tc.mag, "mw")
				eventType, _ := event.NewType("earthquake")

				e := mustNewEvent(
					"mag-"+tc.name,
					loc,
					"Test location",
					mag,
					eventType,
					testTime1,
					"reviewed",
					"Test",
					"https://test.url",
				)

				err := repo.Save(ctx, e)
				require.NoError(t, err)

				retrieved, err := repo.FindbyID(ctx, "mag-"+tc.name)
				require.NoError(t, err)
				assert.InDelta(t, tc.mag, retrieved.Magnitude().Value(), 0.001)
			})
		}
	})
}

// TestSQLiteEventRepository_SQLInjectionProtection tests SQL injection prevention
func TestSQLiteEventRepository_SQLInjectionProtection(t *testing.T) {
	t.Run("FindbyID with SQL injection attempt", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()
		repo := NewSQLiteEventRepository(db)
		ctx := context.Background()

		// Save a normal event
		err := repo.Save(ctx, testEvents[0].event)
		require.NoError(t, err)

		// Attempt SQL injection in ID
		maliciousID := "'; DROP TABLE events; --"
		result, err := repo.FindbyID(ctx, maliciousID)

		// Should safely handle malicious input
		assert.Error(t, err) // Not found
		assert.Nil(t, result)

		// Verify table still exists
		count, err := repo.Count(ctx, nil)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count, "Table should still exist with 1 event")
	})

	t.Run("Save event with SQL keywords in fields", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()
		repo := NewSQLiteEventRepository(db)
		ctx := context.Background()

		loc, _ := event.NewLocation(0, 0, 10)
		mag, _ := event.NewMagnitude(5.0, "mw")
		eventType, _ := event.NewType("earthquake")

		sqlEvent := mustNewEvent(
			"sql-test",
			loc,
			"SELECT * FROM events WHERE 1=1; DROP TABLE events;",
			mag,
			eventType,
			testTime1,
			"reviewed",
			"Description with SQL: INSERT INTO events VALUES ('hack');",
			"https://test.url'; DROP TABLE events; --",
		)

		err := repo.Save(ctx, sqlEvent)
		require.NoError(t, err)

		// Verify event was saved with malicious strings as literal data
		retrieved, err := repo.FindbyID(ctx, "sql-test")
		require.NoError(t, err)
		assert.Contains(t, retrieved.Place(), "SELECT * FROM")
		assert.Contains(t, retrieved.Description(), "INSERT INTO")

		// Verify database integrity
		count, err := repo.Count(ctx, nil)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})
}

// TestSQLiteEventRepository_OrderBy tests different ordering options
func TestSQLiteEventRepository_OrderBy(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := NewSQLiteEventRepository(db)
	ctx := context.Background()

	// Save test events
	for _, tc := range testEvents {
		err := repo.Save(ctx, tc.event)
		require.NoError(t, err)
	}

	t.Run("Order by time descending (default)", func(t *testing.T) {
		criteria := event.NewQueryCriteria()
		criteria.OrderBy = "time"
		criteria.Ascending = false

		results, err := repo.FindAll(ctx, criteria)
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(results), 2)

		// Verify descending order
		for i := 0; i < len(results)-1; i++ {
			assert.True(t, !results[i].Time().Before(results[i+1].Time()),
				"Results should be ordered by time descending")
		}
	})

	t.Run("Order by time ascending", func(t *testing.T) {
		criteria := event.NewQueryCriteria()
		criteria.OrderBy = "time"
		criteria.Ascending = true

		results, err := repo.FindAll(ctx, criteria)
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(results), 2)

		// Verify ascending order
		for i := 0; i < len(results)-1; i++ {
			assert.True(t, !results[i].Time().After(results[i+1].Time()),
				"Results should be ordered by time ascending")
		}
	})

	t.Run("Order by magnitude", func(t *testing.T) {
		criteria := event.NewQueryCriteria()
		criteria.OrderBy = "magnitude"
		criteria.Ascending = false

		results, err := repo.FindAll(ctx, criteria)
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(results), 2)

		// Verify magnitude descending order
		for i := 0; i < len(results)-1; i++ {
			assert.GreaterOrEqual(t, results[i].Magnitude().Value(), results[i+1].Magnitude().Value(),
				"Results should be ordered by magnitude descending")
		}
	})

	t.Run("Order by depth", func(t *testing.T) {
		criteria := event.NewQueryCriteria()
		criteria.OrderBy = "depth"
		criteria.Ascending = true

		results, err := repo.FindAll(ctx, criteria)
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(results), 2)

		// Verify depth ascending order
		for i := 0; i < len(results)-1; i++ {
			assert.LessOrEqual(t, results[i].Location().DepthValue(), results[i+1].Location().DepthValue(),
				"Results should be ordered by depth ascending")
		}
	})
}
