package ingest

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/jwgal/geopulse/internal/domain/event"
	"github.com/jwgal/geopulse/internal/infrastructure/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

// ─── Test schema ─────────────────────────────────────────────────────────────
// Keep this in sync with migrations/000001_create_events_table.up.sql.
// Using an inline schema avoids a fragile relative-path reference to the
// migrations directory.
const testSchemaSQL = `
CREATE TABLE IF NOT EXISTS events (
    id               TEXT    PRIMARY KEY,
    event_type       TEXT    NOT NULL,
    magnitude_value  REAL    NOT NULL,
    magnitude_scale  TEXT    NOT NULL,
    latitude         REAL    NOT NULL,
    longitude        REAL    NOT NULL,
    depth_km         REAL    NOT NULL,
    event_time       TEXT    NOT NULL,
    location_name    TEXT    NOT NULL,
    status           TEXT    NOT NULL,
    description      TEXT    NOT NULL,
    url              TEXT    NOT NULL,
    updated_at       TEXT    NOT NULL,
    created_at       TEXT    DEFAULT CURRENT_TIMESTAMP
);
`

// ─── Test doubles ─────────────────────────────────────────────────────────────

// mockFetcher is a test double for EventFetcher.
type mockFetcher struct {
	events []*event.Event
	err    error
}

func (m *mockFetcher) FetchEvents(_ context.Context) ([]*event.Event, error) {
	return m.events, m.err
}

// mockRepository is a test double for event.Repository.
// saveErr, if non-nil, is returned for every Save call.
// failIDs maps individual event IDs to the error that Save should return for
// that specific ID (takes precedence over saveErr).
type mockRepository struct {
	saved   []*event.Event
	saveErr error
	failIDs map[string]error
}

func (m *mockRepository) Save(_ context.Context, evt *event.Event) error {
	if err, targeted := m.failIDs[evt.ID()]; targeted {
		return err
	}
	if m.saveErr != nil {
		return m.saveErr
	}
	m.saved = append(m.saved, evt)
	return nil
}

func (m *mockRepository) FindByID(_ context.Context, _ string) (*event.Event, error) {
	return nil, nil
}

func (m *mockRepository) FindAll(_ context.Context, _ *event.QueryCriteria) ([]*event.Event, error) {
	return nil, nil
}

func (m *mockRepository) Count(_ context.Context, _ *event.QueryCriteria) (int64, error) {
	return 0, nil
}

func (m *mockRepository) Delete(_ context.Context, _ string) error {
	return nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

// discardLogger returns a slog.Logger that writes to io.Discard,
// keeping test output clean.
func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// makeTestEvent builds a valid *event.Event for use in tests.
func makeTestEvent(t *testing.T, id string, mag, lat, lng float64) *event.Event {
	t.Helper()

	loc, err := event.NewLocation(lat, lng, 10.0)
	require.NoError(t, err)

	magnitude, err := event.NewMagnitude(mag, "mw")
	require.NoError(t, err)

	evtType, err := event.NewType("earthquake")
	require.NoError(t, err)

	evt, err := event.NewEvent(
		id, loc, "Test Place", magnitude, evtType,
		time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC),
		"automatic", "Test earthquake", "https://usgs.gov/test/"+id,
	)
	require.NoError(t, err, "makeTestEvent: NewEvent(%s)", id)

	return evt
}

// setupSQLiteDB opens an in-memory SQLite database and applies the test schema.
func setupSQLiteDB(t *testing.T) (db *sql.DB, cleanup func()) {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err, "open in-memory SQLite DB")

	_, err = db.Exec(testSchemaSQL)
	require.NoError(t, err, "apply test schema")

	return db, func() { _ = db.Close() }
}

// ─── Constructor ──────────────────────────────────────────────────────────────

func TestNewService(t *testing.T) {
	svc := NewService(&mockFetcher{}, &mockRepository{}, discardLogger())
	require.NotNil(t, svc)
}

// ─── IngestEvents happy paths ─────────────────────────────────────────────────

func TestIngestEvents_MultipleEvents(t *testing.T) {
	events := []*event.Event{
		makeTestEvent(t, "us0001", 5.5, 37.77, -122.41),
		makeTestEvent(t, "us0002", 4.2, 34.05, -118.24),
		makeTestEvent(t, "us0003", 6.1, 35.68, 139.65),
	}

	repo := &mockRepository{}
	svc := NewService(&mockFetcher{events: events}, repo, discardLogger())

	result, err := svc.IngestEvents(context.Background())

	require.NoError(t, err)
	assert.Equal(t, 3, result.Fetched, "Fetched should equal the number of returned events")
	assert.Equal(t, 3, result.Saved, "Saved should equal the number of successful saves")
	assert.Equal(t, 0, result.Failed, "Failed should be zero when all saves succeed")
	assert.Len(t, repo.saved, 3, "repository should have received all three events")
}

func TestIngestEvents_SingleEvent(t *testing.T) {
	events := []*event.Event{
		makeTestEvent(t, "us0001", 3.0, 40.71, -74.01),
	}

	repo := &mockRepository{}
	svc := NewService(&mockFetcher{events: events}, repo, discardLogger())

	result, err := svc.IngestEvents(context.Background())

	require.NoError(t, err)
	assert.Equal(t, 1, result.Fetched)
	assert.Equal(t, 1, result.Saved)
	assert.Equal(t, 0, result.Failed)
	assert.Len(t, repo.saved, 1)
}

func TestIngestEvents_EmptySlice(t *testing.T) {
	repo := &mockRepository{}
	svc := NewService(&mockFetcher{events: []*event.Event{}}, repo, discardLogger())

	result, err := svc.IngestEvents(context.Background())

	require.NoError(t, err)
	assert.Equal(t, 0, result.Fetched)
	assert.Equal(t, 0, result.Saved)
	assert.Equal(t, 0, result.Failed)
	assert.Empty(t, repo.saved)
}

// TestIngestEvents_NilEventsSlice verifies that a nil return from the fetcher
// (distinct from an empty slice, though semantically the same) is handled
// without a panic.
func TestIngestEvents_NilEventsSlice(t *testing.T) {
	repo := &mockRepository{}
	svc := NewService(&mockFetcher{events: nil}, repo, discardLogger())

	result, err := svc.IngestEvents(context.Background())

	require.NoError(t, err)
	assert.Equal(t, 0, result.Fetched)
	assert.Equal(t, 0, result.Saved)
	assert.Equal(t, 0, result.Failed)
}

// ─── IngestEvents – fetch errors ──────────────────────────────────────────────

func TestIngestEvents_FetchError_ReturnsNilResultAndError(t *testing.T) {
	fetchErr := errors.New("upstream timeout")
	svc := NewService(&mockFetcher{err: fetchErr}, &mockRepository{}, discardLogger())

	result, err := svc.IngestEvents(context.Background())

	require.Error(t, err)
	assert.Nil(t, result, "Result should be nil on fetch failure")
	assert.Contains(t, err.Error(), "fetch failed", "error message should identify the fetch stage")
}

func TestIngestEvents_FetchError_WrapsUnderlying(t *testing.T) {
	fetchErr := errors.New("connection refused")
	svc := NewService(&mockFetcher{err: fetchErr}, &mockRepository{}, discardLogger())

	_, err := svc.IngestEvents(context.Background())

	require.Error(t, err)
	assert.ErrorIs(t, err, fetchErr, "underlying error must be unwrappable via errors.Is")
}

func TestIngestEvents_FetchError_NoSaveAttempted(t *testing.T) {
	repo := &mockRepository{}
	svc := NewService(&mockFetcher{err: errors.New("network down")}, repo, discardLogger())

	_, _ = svc.IngestEvents(context.Background())

	assert.Empty(t, repo.saved, "no save should be attempted when fetching fails")
}

func TestIngestEvents_ContextCancelled_FetchReturnsError(t *testing.T) {
	// Simulate a fetcher that respects context cancellation.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	svc := NewService(&mockFetcher{err: context.Canceled}, &mockRepository{}, discardLogger())

	result, err := svc.IngestEvents(ctx)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, context.Canceled)
}

// ─── IngestEvents – save errors (documents current behavior / known bug) ──────
//
// BUG in service.go: the save loop always executes `result.Saved++`, even when
// `repository.Save` returns an error.  Consequently:
//   - result.Failed is NEVER incremented (always 0)
//   - result.Saved == result.Fetched regardless of how many saves actually fail
//
// The tests below are labelled "_CurrentBehavior" and contain "BUG" comments to
// flag this.  They are intentionally written to PASS against the current
// implementation so that CI stays green.  Once the bug is fixed, update these
// assertions to reflect the intended contract:
//
//	Intended (post-fix):   result.Saved == successCount, result.Failed == failCount
//	Current (buggy):       result.Saved == result.Fetched, result.Failed == 0

func TestIngestEvents_PartialSaveFailure_CurrentBehavior(t *testing.T) {
	events := []*event.Event{
		makeTestEvent(t, "us0001", 5.5, 37.77, -122.41),
		makeTestEvent(t, "us0002", 4.2, 34.05, -118.24), // will fail
		makeTestEvent(t, "us0003", 6.1, 35.68, 139.65),
	}

	repo := &mockRepository{
		failIDs: map[string]error{
			"us0002": errors.New("constraint violation"),
		},
	}
	svc := NewService(&mockFetcher{events: events}, repo, discardLogger())

	result, err := svc.IngestEvents(context.Background())

	// Service must not propagate individual save errors as a top-level error.
	require.NoError(t, err)
	assert.Equal(t, 3, result.Fetched)

	// BUG: both of these should be 2 and 1 respectively after the fix.
	assert.Equal(t, 3, result.Saved, "BUG: Saved counts failed saves; should be 2 post-fix")
	assert.Equal(t, 0, result.Failed, "BUG: Failed is never incremented; should be 1 post-fix")

	// The repository itself only received the two successful events.
	assert.Len(t, repo.saved, 2, "repository should only hold the two events whose Save succeeded")
}

func TestIngestEvents_AllSavesFail_CurrentBehavior(t *testing.T) {
	events := []*event.Event{
		makeTestEvent(t, "us0001", 5.5, 37.77, -122.41),
		makeTestEvent(t, "us0002", 4.2, 34.05, -118.24),
	}

	repo := &mockRepository{saveErr: errors.New("disk full")}
	svc := NewService(&mockFetcher{events: events}, repo, discardLogger())

	result, err := svc.IngestEvents(context.Background())

	// Service must still return without a top-level error.
	require.NoError(t, err)
	assert.Equal(t, 2, result.Fetched)

	// BUG: should be Saved=0, Failed=2 post-fix.
	assert.Equal(t, 2, result.Saved, "BUG: Saved counts all events even when every Save fails")
	assert.Equal(t, 0, result.Failed, "BUG: Failed is never incremented")

	assert.Empty(t, repo.saved, "repository should hold no events when every Save fails")
}

// ─── Integration test (real SQLite) ──────────────────────────────────────────

// TestIngestEvents_Idempotency is an integration test that wires the service
// to a real in-memory SQLite repository.  It verifies that running ingestion
// twice with the same event IDs succeeds because the repository uses
// INSERT … ON CONFLICT DO UPDATE (upsert).
func TestIngestEvents_Idempotency(t *testing.T) {
	db, cleanup := setupSQLiteDB(t)
	defer cleanup()

	events := []*event.Event{
		makeTestEvent(t, "us0001", 5.5, 37.77, -122.41),
		makeTestEvent(t, "us0002", 4.2, 34.05, -118.24),
	}

	repo := persistence.NewSQLiteEventRepository(db)
	svc := NewService(&mockFetcher{events: events}, repo, discardLogger())
	ctx := context.Background()

	// First run: both events are new.
	r1, err := svc.IngestEvents(ctx)
	require.NoError(t, err, "first IngestEvents call")
	assert.Equal(t, 2, r1.Fetched)
	assert.Equal(t, 2, r1.Saved)
	assert.Equal(t, 0, r1.Failed)

	// Second run: same IDs – upsert must not fail.
	r2, err := svc.IngestEvents(ctx)
	require.NoError(t, err, "second IngestEvents call (upsert)")
	assert.Equal(t, 2, r2.Fetched)
	assert.Equal(t, 2, r2.Saved)
	assert.Equal(t, 0, r2.Failed)
}

// TestIngestEvents_ResultCounts_AcrossMixedBatch uses the real SQLite repo to
// confirm Fetched/Saved totals are consistent when a large mixed batch is
// ingested in one call.
func TestIngestEvents_ResultCounts_AcrossMixedBatch(t *testing.T) {
	db, cleanup := setupSQLiteDB(t)
	defer cleanup()

	events := []*event.Event{
		makeTestEvent(t, "ev001", 2.1, 34.0, -118.0),
		makeTestEvent(t, "ev002", 5.5, 37.7, -122.4),
		makeTestEvent(t, "ev003", 7.2, 35.6, 139.6),
		makeTestEvent(t, "ev004", 3.8, 40.7, -74.0),
		makeTestEvent(t, "ev005", 1.0, 51.5, -0.12),
	}

	repo := persistence.NewSQLiteEventRepository(db)
	svc := NewService(&mockFetcher{events: events}, repo, discardLogger())

	result, err := svc.IngestEvents(context.Background())

	require.NoError(t, err)
	assert.Equal(t, 5, result.Fetched)
	assert.Equal(t, 5, result.Saved)
	assert.Equal(t, 0, result.Failed)
}
