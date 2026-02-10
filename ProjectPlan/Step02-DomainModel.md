# Step 02: Domain Model Implementation

## Objective
Implement the core domain entities and value objects following Domain-Driven Design principles.

## Tasks

### 1. Create Event Type Value Object
Create `internal/domain/event/event_type.go`:
```go
package event

import "fmt"

// EventType represents the category of geospatial event
type EventType string

const (
	EventTypeEarthquake EventType = "earthquake"
	EventTypeStorm      EventType = "storm"
	EventTypeFlood      EventType = "flood"
	EventTypeUnknown    EventType = "unknown"
)

// Valid checks if the event type is recognized
func (et EventType) Valid() bool {
	switch et {
	case EventTypeEarthquake, EventTypeStorm, EventTypeFlood, EventTypeUnknown:
		return true
	}
	return false
}

// String returns the string representation
func (et EventType) String() string {
	return string(et)
}

// ParseEventType converts a string to EventType
func ParseEventType(s string) (EventType, error) {
	et := EventType(s)
	if !et.Valid() {
		return EventTypeUnknown, fmt.Errorf("invalid event type: %s", s)
	}
	return et, nil
}
```

### 2. Create Location Value Object
Create `internal/domain/event/location.go`:
```go
package event

import (
	"fmt"
	"math"
)

// Location represents a geographic coordinate
type Location struct {
	Latitude  float64
	Longitude float64
}

// NewLocation creates a validated Location
func NewLocation(lat, lng float64) (Location, error) {
	if lat < -90 || lat > 90 {
		return Location{}, fmt.Errorf("latitude must be between -90 and 90, got %f", lat)
	}
	if lng < -180 || lng > 180 {
		return Location{}, fmt.Errorf("longitude must be between -180 and 180, got %f", lng)
	}
	return Location{
		Latitude:  lat,
		Longitude: lng,
	}, nil
}

// DistanceKm calculates the distance to another location in kilometers
// Uses the Haversine formula
func (l Location) DistanceKm(other Location) float64 {
	const earthRadiusKm = 6371.0

	lat1 := l.Latitude * math.Pi / 180
	lat2 := other.Latitude * math.Pi / 180
	deltaLat := (other.Latitude - l.Latitude) * math.Pi / 180
	deltaLng := (other.Longitude - l.Longitude) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(deltaLng/2)*math.Sin(deltaLng/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}

// IsWithinRadius checks if the location is within a given radius of another location
func (l Location) IsWithinRadius(other Location, radiusKm float64) bool {
	return l.DistanceKm(other) <= radiusKm
}
```

### 3. Create Magnitude Value Object
Create `internal/domain/event/magnitude.go`:
```go
package event

import "fmt"

// Magnitude represents the intensity of an event
type Magnitude struct {
	Value float64
}

// NewMagnitude creates a validated Magnitude
func NewMagnitude(value float64) (Magnitude, error) {
	if value < 0 || value > 10 {
		return Magnitude{}, fmt.Errorf("magnitude must be between 0 and 10, got %f", value)
	}
	return Magnitude{Value: value}, nil
}

// Float64 returns the magnitude as a float64
func (m Magnitude) Float64() float64 {
	return m.Value
}

// IsSignificant returns true if magnitude is >= 5.0
func (m Magnitude) IsSignificant() bool {
	return m.Value >= 5.0
}
```

### 4. Create Event Entity
Create `internal/domain/event/event.go`:
```go
package event

import (
	"fmt"
	"time"
)

// Event represents a geospatial event (immutable design)
type Event struct {
	id        string
	location  Location
	magnitude Magnitude
	eventType Type
	time      time.Time
	place     string
	status    string
	updated   time.Time
}

// NewEvent creates a new Event with validation
func NewEvent(
	id string,
	location Location,
	place string,
	magnitude Magnitude,
	eventType Type,
	eventTime time.Time,
	status string,
) (*Event, error) {
	if id == "" {
		return nil, fmt.Errorf("event ID cannot be empty")
	}
	if eventTime.IsZero() {
		return nil, fmt.Errorf("event time cannot be zero")
	}
	if status == "" {
		return nil, fmt.Errorf("event status cannot be empty")
	}

	return &Event{
		id:        id,
		location:  location,
		place:     place,
		magnitude: magnitude,
		eventType: eventType,
		time:      eventTime,
		updated:   time.Now(),
		status:    status,
	}, nil
}

// Getters (read-only access to enforce immutability)
func (e *Event) ID() string           { return e.id }
func (e *Event) Location() Location   { return e.location }
func (e *Event) Magnitude() Magnitude { return e.magnitude }
func (e *Event) Type() Type            { return e.eventType }
func (e *Event) Time() time.Time       { return e.time }
func (e *Event) Place() string         { return e.place }
func (e *Event) Status() string        { return e.status }
func (e *Event) Updated() time.Time    { return e.updated }

// String returns a formatted string representation
func (e *Event) String() string {
	return fmt.Sprintf("Event ID: %s, Type: %s, Magnitude: %s, Location: %s, Place: %s, Time: %s",
		e.id, e.eventType.String(), e.magnitude.String(), e.location.String(), e.Place(), e.time.Format(time.RFC3339))
}

// IsSignificant returns true if the event magnitude meets the threshold
func (e *Event) IsSignificant(threshold float64) bool {
	return e.magnitude.Value() >= threshold
}

// UpdateStatus returns a new Event with updated status (immutable pattern)
func (e *Event) UpdateStatus(newStatus string, updatedTime time.Time) *Event {
	return &Event{
		id:        e.id,
		location:  e.location,
		place:     e.place,
		magnitude: e.magnitude,
		eventType: e.eventType,
		time:      e.time,
		status:    newStatus,
		updated:   updatedTime,
	}
}
```

### 5. Create Repository Interface
Create `internal/domain/event/repository.go`:
```go
package event

import (
	"context"
	"time"
)

// Repository defines the interface for event persistence
// This lives in the domain layer and will be implemented in infrastructure layer
type Repository interface {
	// Save inserts a new event or updates existing one
	Save(ctx context.Context, event *Event) error
	
	// FindByID retrieves an event by its unique ID
	FindByID(ctx context.Context, id string) (*Event, error)
	
	// FindAll retrieves events matching the given criteria
	FindAll(ctx context.Context, criteria QueryCriteria) ([]*Event, error)
	
	// Count returns the total number of events matching the criteria
	Count(ctx context.Context, criteria QueryCriteria) (int64, error)
	
	// Delete removes an event by ID
	Delete(ctx context.Context, id string) error
}

// QueryCriteria defines event query parameters for filtering and pagination
type QueryCriteria struct {
	// Magnitude filters
	MinMagnitude *float64
	MaxMagnitude *float64
	
	// Type filter
	EventTypes []Type
	
	// Location-based filters
	Location *Location
	RadiusKm *float64 // Used with Location for proximity search
	
	// Time range filters
	FromTime *time.Time
	ToTime   *time.Time
	
	// Status filter
	Statuses []string
	
	// Pagination
	Limit  int
	Offset int
	
	// Sorting
	OrderBy   string // e.g., "time", "magnitude", "updated"
	Ascending bool
}

// NewQueryCriteria creates a QueryCriteria with sensible defaults
func NewQueryCriteria() QueryCriteria {
	return QueryCriteria{
		Limit:     100,
		Offset:    0,
		OrderBy:   "time",
		Ascending: false, // Most recent first by default
	}
}

// WithMagnitudeRange sets magnitude filters (fluent API)
func (qc QueryCriteria) WithMagnitudeRange(min, max float64) QueryCriteria {
	qc.MinMagnitude = &min
	qc.MaxMagnitude = &max
	return qc
}

// WithTimeRange sets time range filters (fluent API)
func (qc QueryCriteria) WithTimeRange(from, to time.Time) QueryCriteria {
	qc.FromTime = &from
	qc.ToTime = &to
	return qc
}

// WithPagination sets pagination parameters (fluent API)
func (qc QueryCriteria) WithPagination(limit, offset int) QueryCriteria {
	qc.Limit = limit
	qc.Offset = offset
	return qc
}
```

### 6. Create Domain Tests
Create `internal/domain/event/location_test.go`:
```go
package event

import (
	"testing"
)

func TestNewLocation(t *testing.T) {
	tests := []struct {
		name      string
		lat       float64
		lng       float64
		wantError bool
	}{
		{"valid location", 37.7749, -122.4194, false},
		{"latitude too high", 91.0, 0, true},
		{"latitude too low", -91.0, 0, true},
		{"longitude too high", 0, 181.0, true},
		{"longitude too low", 0, -181.0, true},
		{"edge case max", 90.0, 180.0, false},
		{"edge case min", -90.0, -180.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewLocation(tt.lat, tt.lng)
			if (err != nil) != tt.wantError {
				t.Errorf("NewLocation() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestLocationDistanceKm(t *testing.T) {
	// San Francisco
	sf, _ := NewLocation(37.7749, -122.4194)
	// Los Angeles
	la, _ := NewLocation(34.0522, -118.2437)

	distance := sf.DistanceKm(la)
	
	// Approximate distance is ~559 km
	if distance < 550 || distance > 570 {
		t.Errorf("Distance between SF and LA should be ~559km, got %f", distance)
	}
}
```

Create `internal/domain/event/event_test.go`:
```go
package event

import (
	"testing"
	"time"
)

func TestNewEvent(t *testing.T) {
	loc, _ := NewLocation(37.7749, -122.4194)
	mag, _ := NewMagnitude(5.5)
	
	event, err := NewEvent(
		"USGS",
		"us1234",
		EventTypeEarthquake,
		time.Now(),
		loc,
		mag,
	)
	
	if err != nil {
		t.Fatalf("NewEvent() failed: %v", err)
	}
	
	if event.Source != "USGS" {
		t.Errorf("Expected source USGS, got %s", event.Source)
	}
	
	if event.Magnitude.Value != 5.5 {
		t.Errorf("Expected magnitude 5.5, got %f", event.Magnitude.Value)
	}
}

func TestEventSetDepth(t *testing.T) {
	loc, _ := NewLocation(0, 0)
	mag, _ := NewMagnitude(5.0)
	event, _ := NewEvent("USGS", "test", EventTypeEarthquake, time.Now(), loc, mag)
	
	err := event.SetDepth(10.5)
	if err != nil {
		t.Fatalf("SetDepth() failed: %v", err)
	}
	
	if event.Depth == nil || *event.Depth != 10.5 {
		t.Errorf("Expected depth 10.5, got %v", event.Depth)
	}
	
	// Test negative depth
	err = event.SetDepth(-1)
	if err == nil {
		t.Error("Expected error for negative depth")
	}
}
```

### 7. Run Domain Tests
```powershell
go test ./internal/domain/event -v
```

## Success Criteria
- ✓ All value objects created with validation
- ✓ Event entity implemented with business logic
- ✓ Repository interface defined
- ✓ All domain tests pass
- ✓ No external dependencies in domain layer

## Next Step
Proceed to **Step03-DatabaseSetup.md** to set up SQLite and migrations.
