# Step 05: USGS Client Implementation

## Objective
Implement the HTTP client to fetch and decode earthquake data from the USGS API.

## Tasks

### 1. Create USGS Data Structures
Create `internal/infrastructure/usgs/types.go`:
```go
package usgs

import "time"

// GeoJSONResponse represents the USGS earthquake feed response
type GeoJSONResponse struct {
	Type     string    `json:"type"`
	Metadata Metadata  `json:"metadata"`
	Features []Feature `json:"features"`
}

// Metadata contains feed metadata
type Metadata struct {
	Generated int64  `json:"generated"`
	URL       string `json:"url"`
	Title     string `json:"title"`
	Status    int    `json:"status"`
	API       string `json:"api"`
	Count     int    `json:"count"`
}

// Feature represents a single earthquake event
type Feature struct {
	Type       string     `json:"type"`
	Properties Properties `json:"properties"`
	Geometry   Geometry   `json:"geometry"`
	ID         string     `json:"id"`
}

// Properties contains event details
type Properties struct {
	Mag     *float64 `json:"mag"`
	Place   string   `json:"place"`
	Time    int64    `json:"time"` // Unix timestamp in milliseconds
	Updated int64    `json:"updated"`
	Tz      *int     `json:"tz"`
	URL     string   `json:"url"`
	Detail  string   `json:"detail"`
	Felt    *int     `json:"felt"`
	CDI     *float64 `json:"cdi"`
	MMI     *float64 `json:"mmi"`
	Alert   *string  `json:"alert"`
	Status  string   `json:"status"`
	Tsunami int      `json:"tsunami"`
	Sig     int      `json:"sig"`
	Net     string   `json:"net"`
	Code    string   `json:"code"`
	IDs     string   `json:"ids"`
	Sources string   `json:"sources"`
	Types   string   `json:"types"`
	NST     *int     `json:"nst"`
	Dmin    *float64 `json:"dmin"`
	RMS     *float64 `json:"rms"`
	Gap     *float64 `json:"gap"`
	MagType string   `json:"magType"`
	Type    string   `json:"type"`
	Title   string   `json:"title"`
}

// Geometry contains spatial coordinates
type Geometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"` // [longitude, latitude, depth]
}
```

### 2. Create USGS Client
Create `internal/infrastructure/usgs/client.go`:
```go
package usgs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/yourusername/geopulse/internal/domain/event"
)

// Client fetches earthquake data from USGS
type Client struct {
	httpClient *http.Client
	endpoint   string
}

// NewClient creates a new USGS client
func NewClient(endpoint string, timeout time.Duration) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		endpoint: endpoint,
	}
}

// FetchEvents retrieves earthquake events from USGS
func (c *Client) FetchEvents(ctx context.Context) ([]*event.Event, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "GeoPulse/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var geoJSON GeoJSONResponse
	if err := json.NewDecoder(resp.Body).Decode(&geoJSON); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return c.convertToEvents(geoJSON)
}

// convertToEvents converts USGS features to domain events
func (c *Client) convertToEvents(geoJSON GeoJSONResponse) ([]*event.Event, error) {
	events := make([]*event.Event, 0, len(geoJSON.Features))

	for _, feature := range geoJSON.Features {
		evt, err := c.convertFeature(feature)
		if err != nil {
			// Log error but continue processing other events
			continue
		}
		events = append(events, evt)
	}

	return events, nil
}

// convertFeature converts a single USGS feature to a domain event
func (c *Client) convertFeature(feature Feature) (*event.Event, error) {
	// Validate coordinates
	if len(feature.Geometry.Coordinates) < 2 {
		return nil, fmt.Errorf("invalid coordinates for event %s", feature.ID)
	}

	// Extract coordinates (USGS format: [longitude, latitude, depth])
	longitude := feature.Geometry.Coordinates[0]
	latitude := feature.Geometry.Coordinates[1]

	location, err := event.NewLocation(latitude, longitude)
	if err != nil {
		return nil, fmt.Errorf("invalid location: %w", err)
	}

	// Handle nil magnitude
	magValue := 0.0
	if feature.Properties.Mag != nil {
		magValue = *feature.Properties.Mag
	}

	magnitude, err := event.NewMagnitude(magValue)
	if err != nil {
		return nil, fmt.Errorf("invalid magnitude: %w", err)
	}

	// Convert timestamp (USGS uses milliseconds)
	occurredAt := time.UnixMilli(feature.Properties.Time)

	// Create event
	evt, err := event.NewEvent(
		"USGS",
		feature.ID,
		event.EventTypeEarthquake,
		occurredAt,
		location,
		magnitude,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	// Set depth if available
	if len(feature.Geometry.Coordinates) >= 3 {
		depth := feature.Geometry.Coordinates[2]
		evt.SetDepth(depth)
	}

	// Add metadata
	evt.AddMetadata("place", feature.Properties.Place)
	evt.AddMetadata("url", feature.Properties.URL)
	evt.AddMetadata("type", feature.Properties.Type)
	evt.AddMetadata("mag_type", feature.Properties.MagType)
	
	if feature.Properties.Alert != nil {
		evt.AddMetadata("alert", *feature.Properties.Alert)
	}

	return evt, nil
}
```

*Note: Update import path*

### 3. Create Mock Client for Testing
Create `internal/infrastructure/usgs/mock_client.go`:
```go
package usgs

import (
	"context"
	"time"

	"github.com/yourusername/geopulse/internal/domain/event"
)

// MockClient is a mock USGS client for testing
type MockClient struct {
	Events []*event.Event
	Err    error
}

// NewMockClient creates a new mock client
func NewMockClient() *MockClient {
	return &MockClient{
		Events: make([]*event.Event, 0),
	}
}

// FetchEvents returns mock events
func (m *MockClient) FetchEvents(ctx context.Context) ([]*event.Event, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Events, nil
}

// AddMockEvent adds a mock event to the client
func (m *MockClient) AddMockEvent(magnitude float64, lat, lng float64) {
	loc, _ := event.NewLocation(lat, lng)
	mag, _ := event.NewMagnitude(magnitude)
	
	evt, _ := event.NewEvent(
		"USGS",
		"mock-event",
		event.EventTypeEarthquake,
		time.Now(),
		loc,
		mag,
	)
	
	m.Events = append(m.Events, evt)
}
```

### 4. Create Client Tests
Create `internal/infrastructure/usgs/client_test.go`:
```go
package usgs

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_FetchEvents(t *testing.T) {
	// Create mock USGS server
	mockResponse := GeoJSONResponse{
		Type: "FeatureCollection",
		Metadata: Metadata{
			Generated: time.Now().Unix(),
			Count:     2,
		},
		Features: []Feature{
			{
				Type: "Feature",
				ID:   "us1234",
				Geometry: Geometry{
					Type:        "Point",
					Coordinates: []float64{-122.4194, 37.7749, 10.5},
				},
				Properties: Properties{
					Mag:     floatPtr(5.5),
					Place:   "San Francisco Bay Area",
					Time:    time.Now().UnixMilli(),
					Type:    "earthquake",
					MagType: "ml",
					URL:     "https://earthquake.usgs.gov/earthquakes/eventpage/us1234",
				},
			},
			{
				Type: "Feature",
				ID:   "us5678",
				Geometry: Geometry{
					Type:        "Point",
					Coordinates: []float64{-118.2437, 34.0522, 8.2},
				},
				Properties: Properties{
					Mag:     floatPtr(4.2),
					Place:   "Los Angeles",
					Time:    time.Now().UnixMilli(),
					Type:    "earthquake",
					MagType: "ml",
					URL:     "https://earthquake.usgs.gov/earthquakes/eventpage/us5678",
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Create client pointing to mock server
	client := NewClient(server.URL, 10*time.Second)

	// Fetch events
	ctx := context.Background()
	events, err := client.FetchEvents(ctx)
	if err != nil {
		t.Fatalf("FetchEvents() failed: %v", err)
	}

	// Verify results
	if len(events) != 2 {
		t.Errorf("Expected 2 events, got %d", len(events))
	}

	// Verify first event
	evt := events[0]
	if evt.Source != "USGS" {
		t.Errorf("Expected source USGS, got %s", evt.Source)
	}
	if evt.Magnitude.Value != 5.5 {
		t.Errorf("Expected magnitude 5.5, got %f", evt.Magnitude.Value)
	}
	if evt.Depth == nil || *evt.Depth != 10.5 {
		t.Errorf("Expected depth 10.5, got %v", evt.Depth)
	}
}

func TestClient_HandleError(t *testing.T) {
	// Create server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := NewClient(server.URL, 10*time.Second)

	ctx := context.Background()
	_, err := client.FetchEvents(ctx)
	if err == nil {
		t.Error("Expected error for 500 response")
	}
}

func floatPtr(f float64) *float64 {
	return &f
}
```

### 5. Create Client Interface
Create `internal/application/ingest/client.go`:
```go
package ingest

import (
	"context"

	"github.com/yourusername/geopulse/internal/domain/event"
)

// EventClient defines the interface for fetching events from external sources
type EventClient interface {
	FetchEvents(ctx context.Context) ([]*event.Event, error)
}
```

### 6. Run Client Tests
```powershell
go test ./internal/infrastructure/usgs -v
```

## Success Criteria
- ✓ USGS client implemented
- ✓ GeoJSON decoding works
- ✓ Domain event conversion correct
- ✓ Mock client created for testing
- ✓ Error handling implemented
- ✓ All tests pass

## Next Step
Proceed to **Step06-ApplicationServices.md** to implement ingest and query services.
