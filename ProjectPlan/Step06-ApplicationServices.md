# Step 06: Application Services

## Objective
Implement the ingest and query application services that orchestrate domain logic and infrastructure.

## Tasks

### 1. Create Ingest Service
Create `internal/application/ingest/service.go`:
```go
package ingest

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/yourusername/geopulse/internal/domain/event"
)

// Service handles event ingestion from external sources
type Service struct {
	client     EventClient
	repository event.Repository
	logger     *slog.Logger
}

// NewService creates a new ingest service
func NewService(client EventClient, repo event.Repository, logger *slog.Logger) *Service {
	return &Service{
		client:     client,
		repository: repo,
		logger:     logger,
	}
}

// IngestEvents fetches events from external source and saves them
func (s *Service) IngestEvents(ctx context.Context) (*IngestResult, error) {
	s.logger.Info("starting event ingestion")

	// Fetch events from external API
	events, err := s.client.FetchEvents(ctx)
	if err != nil {
		s.logger.Error("failed to fetch events", "error", err)
		return nil, fmt.Errorf("fetch failed: %w", err)
	}

	result := &IngestResult{
		Fetched: len(events),
	}

	// Save each event
	for _, evt := range events {
		// Check if event already exists
		existing, err := s.repository.FindBySourceEventID(ctx, evt.Source, evt.SourceEventID)
		if err != nil {
			s.logger.Warn("error checking existing event", "error", err)
			result.Failed++
			continue
		}

		if existing != nil {
			// Update existing event
			existing.Update(evt.Magnitude, evt.OccurredAt)
			if err := s.repository.Save(ctx, existing); err != nil {
				s.logger.Error("failed to update event",
					"id", existing.ID,
					"error", err,
				)
				result.Failed++
				continue
			}
			result.Updated++
		} else {
			// Insert new event
			if err := s.repository.Save(ctx, evt); err != nil {
				s.logger.Error("failed to save event",
					"source_id", evt.SourceEventID,
					"error", err,
				)
				result.Failed++
				continue
			}
			result.Created++
		}
	}

	s.logger.Info("ingestion complete",
		"fetched", result.Fetched,
		"created", result.Created,
		"updated", result.Updated,
		"failed", result.Failed,
	)

	return result, nil
}

// IngestResult contains metrics about an ingestion run
type IngestResult struct {
	Fetched int
	Created int
	Updated int
	Failed  int
}
```

### 2. Create Scheduler
Create `internal/application/ingest/scheduler.go`:
```go
package ingest

import (
	"context"
	"log/slog"
	"time"
)

// Scheduler runs periodic event ingestion
type Scheduler struct {
	service  *Service
	interval time.Duration
	logger   *slog.Logger
	stopCh   chan struct{}
}

// NewScheduler creates a new ingestion scheduler
func NewScheduler(service *Service, interval time.Duration, logger *slog.Logger) *Scheduler {
	return &Scheduler{
		service:  service,
		interval: interval,
		logger:   logger,
		stopCh:   make(chan struct{}),
	}
}

// Start begins the periodic ingestion process
func (s *Scheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	s.logger.Info("scheduler started", "interval", s.interval)

	// Run immediately on start
	if err := s.runIngestion(ctx); err != nil {
		s.logger.Error("initial ingestion failed", "error", err)
	}

	for {
		select {
		case <-ticker.C:
			if err := s.runIngestion(ctx); err != nil {
				s.logger.Error("scheduled ingestion failed", "error", err)
			}
		case <-s.stopCh:
			s.logger.Info("scheduler stopped")
			return
		case <-ctx.Done():
			s.logger.Info("scheduler context cancelled")
			return
		}
	}
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	close(s.stopCh)
}

// runIngestion performs a single ingestion run
func (s *Scheduler) runIngestion(ctx context.Context) error {
	start := time.Now()
	
	result, err := s.service.IngestEvents(ctx)
	
	duration := time.Since(start)
	
	if err != nil {
		s.logger.Error("ingestion failed",
			"duration", duration,
			"error", err,
		)
		return err
	}

	s.logger.Info("ingestion succeeded",
		"duration", duration,
		"created", result.Created,
		"updated", result.Updated,
	)

	return nil
}
```

### 3. Create Query Service
Create `internal/application/query/service.go`:
```go
package query

import (
	"context"
	"fmt"
	"time"

	"github.com/yourusername/geopulse/internal/domain/event"
)

// Service handles event queries
type Service struct {
	repository event.Repository
}

// NewService creates a new query service
func NewService(repo event.Repository) *Service {
	return &Service{
		repository: repo,
	}
}

// QueryEvents retrieves events based on filter parameters
func (s *Service) QueryEvents(ctx context.Context, params QueryParams) (*QueryResult, error) {
	// Validate and sanitize parameters
	if err := params.Validate(); err != nil {
		return nil, fmt.Errorf("invalid query parameters: %w", err)
	}

	// Build criteria from parameters
	criteria := params.ToCriteria()

	// Get total count
	total, err := s.repository.Count(ctx, criteria)
	if err != nil {
		return nil, fmt.Errorf("failed to count events: %w", err)
	}

	// Get events
	events, err := s.repository.Query(ctx, criteria)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}

	return &QueryResult{
		Events: events,
		Pagination: Pagination{
			Limit:  criteria.Limit,
			Offset: criteria.Offset,
			Total:  total,
		},
	}, nil
}

// QueryParams defines API query parameters
type QueryParams struct {
	MinMagnitude *float64
	MaxMagnitude *float64
	EventType    *string
	Latitude     *float64
	Longitude    *float64
	RadiusKm     *float64
	FromTime     *time.Time
	ToTime       *time.Time
	Limit        int
	Offset       int
}

// Validate checks query parameters
func (p *QueryParams) Validate() error {
	if p.MinMagnitude != nil && (*p.MinMagnitude < 0 || *p.MinMagnitude > 10) {
		return fmt.Errorf("minMagnitude must be between 0 and 10")
	}
	if p.MaxMagnitude != nil && (*p.MaxMagnitude < 0 || *p.MaxMagnitude > 10) {
		return fmt.Errorf("maxMagnitude must be between 0 and 10")
	}
	if p.Limit < 0 || p.Limit > 1000 {
		return fmt.Errorf("limit must be between 0 and 1000")
	}
	if p.Offset < 0 {
		return fmt.Errorf("offset must be non-negative")
	}
	if p.RadiusKm != nil && (*p.RadiusKm < 1 || *p.RadiusKm > 20000) {
		return fmt.Errorf("radiusKm must be between 1 and 20000")
	}

	// Spatial query requires both lat/lng and radius
	hasLocation := p.Latitude != nil && p.Longitude != nil
	hasRadius := p.RadiusKm != nil
	if hasLocation != hasRadius {
		return fmt.Errorf("spatial query requires lat, lng, and radiusKm")
	}

	if p.Latitude != nil && (*p.Latitude < -90 || *p.Latitude > 90) {
		return fmt.Errorf("latitude must be between -90 and 90")
	}
	if p.Longitude != nil && (*p.Longitude < -180 || *p.Longitude > 180) {
		return fmt.Errorf("longitude must be between -180 and 180")
	}

	return nil
}

// ToCriteria converts query params to domain criteria
func (p *QueryParams) ToCriteria() event.QueryCriteria {
	criteria := event.NewQueryCriteria()
	criteria.MinMagnitude = p.MinMagnitude
	criteria.MaxMagnitude = p.MaxMagnitude
	criteria.FromTime = p.FromTime
	criteria.ToTime = p.ToTime
	criteria.Limit = p.Limit
	criteria.Offset = p.Offset

	// Handle event type
	if p.EventType != nil {
		if et, err := event.ParseEventType(*p.EventType); err == nil {
			criteria.EventTypes = []event.EventType{et}
		}
	}

	// Handle spatial query
	if p.Latitude != nil && p.Longitude != nil && p.RadiusKm != nil {
		if loc, err := event.NewLocation(*p.Latitude, *p.Longitude); err == nil {
			criteria.Location = &loc
			criteria.RadiusKm = p.RadiusKm
		}
	}

	return criteria
}

// QueryResult contains query results and pagination info
type QueryResult struct {
	Events     []*event.Event
	Pagination Pagination
}

// Pagination contains pagination metadata
type Pagination struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}
```

### 4. Create Service Tests
Create `internal/application/ingest/service_test.go`:
```go
package ingest

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/yourusername/geopulse/internal/domain/event"
	"github.com/yourusername/geopulse/internal/infrastructure/persistence"
	"github.com/yourusername/geopulse/internal/infrastructure/usgs"
	"github.com/yourusername/geopulse/tests/testutil"
)

func TestIngestService_IngestEvents(t *testing.T) {
	// Setup test database
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	repo := persistence.NewSQLiteRepository(&persistence.Database{DB: db})
	
	// Create mock client with test data
	mockClient := usgs.NewMockClient()
	mockClient.AddMockEvent(5.5, 37.7749, -122.4194)
	mockClient.AddMockEvent(4.2, 34.0522, -118.2437)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(mockClient, repo, logger)

	// Run ingestion
	ctx := context.Background()
	result, err := service.IngestEvents(ctx)
	if err != nil {
		t.Fatalf("IngestEvents() failed: %v", err)
	}

	// Verify results
	if result.Fetched != 2 {
		t.Errorf("Expected 2 fetched, got %d", result.Fetched)
	}
	if result.Created != 2 {
		t.Errorf("Expected 2 created, got %d", result.Created)
	}

	// Run again to test update logic
	result, err = service.IngestEvents(ctx)
	if err != nil {
		t.Fatalf("Second IngestEvents() failed: %v", err)
	}

	if result.Updated != 2 {
		t.Errorf("Expected 2 updated, got %d", result.Updated)
	}
	if result.Created != 0 {
		t.Errorf("Expected 0 created on second run, got %d", result.Created)
	}
}
```

Create `internal/application/query/service_test.go`:
```go
package query

import (
	"context"
	"testing"
	"time"

	"github.com/yourusername/geopulse/internal/domain/event"
	"github.com/yourusername/geopulse/internal/infrastructure/persistence"
	"github.com/yourusername/geopulse/tests/testutil"
)

func TestQueryService_QueryEvents(t *testing.T) {
	// Setup test database
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	repo := persistence.NewSQLiteRepository(&persistence.Database{DB: db})

	// Insert test data
	insertTestEvents(t, repo)

	service := NewService(repo)
	ctx := context.Background()

	t.Run("basic query", func(t *testing.T) {
		params := QueryParams{
			Limit:  10,
			Offset: 0,
		}

		result, err := service.QueryEvents(ctx, params)
		if err != nil {
			t.Fatalf("QueryEvents() failed: %v", err)
		}

		if len(result.Events) == 0 {
			t.Error("Expected events to be returned")
		}
	})

	t.Run("magnitude filter", func(t *testing.T) {
		minMag := 5.0
		params := QueryParams{
			MinMagnitude: &minMag,
			Limit:        10,
			Offset:       0,
		}

		result, err := service.QueryEvents(ctx, params)
		if err != nil {
			t.Fatalf("QueryEvents() failed: %v", err)
		}

		for _, evt := range result.Events {
			if evt.Magnitude.Value < 5.0 {
				t.Errorf("Expected magnitude >= 5.0, got %f", evt.Magnitude.Value)
			}
		}
	})

	t.Run("invalid parameters", func(t *testing.T) {
		minMag := 15.0 // Invalid
		params := QueryParams{
			MinMagnitude: &minMag,
			Limit:        10,
		}

		_, err := service.QueryEvents(ctx, params)
		if err == nil {
			t.Error("Expected validation error")
		}
	})
}

func insertTestEvents(t *testing.T, repo *persistence.SQLiteRepository) {
	t.Helper()
	ctx := context.Background()

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
		
		repo.Save(ctx, evt)
	}
}
```

### 5. Run Service Tests
```powershell
go test ./internal/application/... -v
```

## Success Criteria
- ✓ Ingest service implemented with duplicate detection
- ✓ Scheduler created for periodic ingestion
- ✓ Query service with parameter validation
- ✓ Proper logging throughout
- ✓ All application tests pass

## Next Step
Proceed to **Step07-HTTPHandlers.md** to implement REST API handlers.
