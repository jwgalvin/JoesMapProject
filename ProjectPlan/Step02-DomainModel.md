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

// Event represents a geospatial event
type Event struct {
	ID            string
	EventType     EventType
	Source        string
	SourceEventID string
	OccurredAt    time.Time
	Location      Location
	Magnitude     Magnitude
	Depth         *float64 // Optional, in kilometers
	Metadata      map[string]interface{}
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// NewEvent creates a new Event with validation
func NewEvent(
	source string,
	sourceEventID string,
	eventType EventType,
	occurredAt time.Time,
	location Location,
	magnitude Magnitude,
) (*Event, error) {
	if source == "" {
		return nil, fmt.Errorf("source cannot be empty")
	}
	if sourceEventID == "" {
		return nil, fmt.Errorf("sourceEventID cannot be empty")
	}
	if !eventType.Valid() {
		return nil, fmt.Errorf("invalid event type: %s", eventType)
	}

	now := time.Now()
	
	return &Event{
		Source:        source,
		SourceEventID: sourceEventID,
		EventType:     eventType,
		OccurredAt:    occurredAt,
		Location:      location,
		Magnitude:     magnitude,
		Metadata:      make(map[string]interface{}),
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

// SetDepth sets the depth of the event
func (e *Event) SetDepth(depth float64) error {
	if depth < 0 {
		return fmt.Errorf("depth cannot be negative")
	}
	e.Depth = &depth
	return nil
}

// AddMetadata adds a metadata field
func (e *Event) AddMetadata(key string, value interface{}) {
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	e.Metadata[key] = value
}

// Update updates the event's mutable fields
func (e *Event) Update(magnitude Magnitude, occurredAt time.Time) {
	e.Magnitude = magnitude
	e.OccurredAt = occurredAt
	e.UpdatedAt = time.Now()
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
type Repository interface {
	// Save inserts a new event or updates existing one
	Save(ctx context.Context, event *Event) error
	
	// FindByID retrieves an event by its ID
	FindByID(ctx context.Context, id string) (*Event, error)
	
	// FindBySourceEventID retrieves an event by source and source event ID
	FindBySourceEventID(ctx context.Context, source, sourceEventID string) (*Event, error)
	
	// Query retrieves events matching the given criteria
	Query(ctx context.Context, criteria QueryCriteria) ([]*Event, error)
	
	// Count returns the total number of events matching the criteria
	Count(ctx context.Context, criteria QueryCriteria) (int, error)
}

// QueryCriteria defines event query parameters
type QueryCriteria struct {
	MinMagnitude *float64
	MaxMagnitude *float64
	EventTypes   []EventType
	Location     *Location
	RadiusKm     *float64
	FromTime     *time.Time
	ToTime       *time.Time
	Limit        int
	Offset       int
}

// NewQueryCriteria creates a QueryCriteria with default values
func NewQueryCriteria() QueryCriteria {
	return QueryCriteria{
		Limit:  100,
		Offset: 0,
	}
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
