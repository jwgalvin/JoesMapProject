# Step 07: HTTP Handlers and Routing

## Objective
Implement REST API handlers, routing, and middleware for the HTTP interface.

## Tasks

### 1. Create DTOs
Create `internal/interfaces/api/dto.go`:
```go
package api

import (
	"time"

	"github.com/yourusername/geopulse/internal/domain/event"
)

// EventDTO represents an event in API responses
type EventDTO struct {
	ID         string                 `json:"id"`
	EventType  string                 `json:"eventType"`
	Source     string                 `json:"source"`
	OccurredAt time.Time              `json:"occurredAt"`
	Location   LocationDTO            `json:"location"`
	Magnitude  float64                `json:"magnitude"`
	Depth      *float64               `json:"depth,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt  time.Time              `json:"createdAt"`
	UpdatedAt  time.Time              `json:"updatedAt"`
}

// LocationDTO represents a geographic location
type LocationDTO struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// EventsResponse wraps event list with pagination
type EventsResponse struct {
	Data       []EventDTO       `json:"data"`
	Pagination PaginationDTO    `json:"pagination"`
}

// PaginationDTO contains pagination metadata
type PaginationDTO struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

// ErrorResponse represents an API error
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error information
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

// GeoJSONFeatureCollection represents a GeoJSON feature collection
type GeoJSONFeatureCollection struct {
	Type     string           `json:"type"`
	Features []GeoJSONFeature `json:"features"`
}

// GeoJSONFeature represents a GeoJSON feature
type GeoJSONFeature struct {
	Type       string                 `json:"type"`
	Geometry   GeoJSONGeometry        `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
}

// GeoJSONGeometry represents GeoJSON geometry
type GeoJSONGeometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

// ToEventDTO converts domain event to DTO
func ToEventDTO(evt *event.Event) EventDTO {
	return EventDTO{
		ID:        evt.ID,
		EventType: evt.EventType.String(),
		Source:    evt.Source,
		OccurredAt: evt.OccurredAt,
		Location: LocationDTO{
			Latitude:  evt.Location.Latitude,
			Longitude: evt.Location.Longitude,
		},
		Magnitude: evt.Magnitude.Float64(),
		Depth:     evt.Depth,
		Metadata:  evt.Metadata,
		CreatedAt: evt.CreatedAt,
		UpdatedAt: evt.UpdatedAt,
	}
}

// ToGeoJSON converts events to GeoJSON FeatureCollection
func ToGeoJSON(events []*event.Event) GeoJSONFeatureCollection {
	features := make([]GeoJSONFeature, len(events))
	
	for i, evt := range events {
		coords := []float64{evt.Location.Longitude, evt.Location.Latitude}
		if evt.Depth != nil {
			coords = append(coords, *evt.Depth)
		}

		properties := map[string]interface{}{
			"id":         evt.ID,
			"eventType":  evt.EventType.String(),
			"source":     evt.Source,
			"occurredAt": evt.OccurredAt,
			"magnitude":  evt.Magnitude.Float64(),
		}

		// Add metadata
		for k, v := range evt.Metadata {
			properties[k] = v
		}

		features[i] = GeoJSONFeature{
			Type: "Feature",
			Geometry: GeoJSONGeometry{
				Type:        "Point",
				Coordinates: coords,
			},
			Properties: properties,
		}
	}

	return GeoJSONFeatureCollection{
		Type:     "FeatureCollection",
		Features: features,
	}
}
```

### 2. Create Middleware
Create `internal/infrastructure/http/middleware.go`:
```go
package http

import (
	"log/slog"
	"net/http"
	"time"
)

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapped, r)

			logger.Info("http request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.statusCode,
				"duration", time.Since(start),
				"remote_addr", r.RemoteAddr,
			)
		})
	}
}

// CORSMiddleware adds CORS headers
func CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			
			// Check if origin is allowed
			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			}

			// Handle preflight
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
```

### 3. Create HTTP Handlers
Create `internal/infrastructure/http/handlers.go`:
```go
package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/yourusername/geopulse/internal/application/query"
	"github.com/yourusername/geopulse/internal/interfaces/api"
)

// Handler contains HTTP handlers
type Handler struct {
	queryService *query.Service
}

// NewHandler creates a new HTTP handler
func NewHandler(queryService *query.Service) *Handler {
	return &Handler{
		queryService: queryService,
	}
}

// Health handles health check requests
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status": "ok",
		"timestamp": time.Now(),
	}
	writeJSON(w, http.StatusOK, response)
}

// ListEvents handles GET /v1/events
func (h *Handler) ListEvents(w http.ResponseWriter, r *http.Request) {
	params, err := parseQueryParams(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_PARAMETER", err.Error(), "")
		return
	}

	result, err := h.queryService.QueryEvents(r.Context(), params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "QUERY_FAILED", err.Error(), "")
		return
	}

	// Convert to DTOs
	dtos := make([]api.EventDTO, len(result.Events))
	for i, evt := range result.Events {
		dtos[i] = api.ToEventDTO(evt)
	}

	response := api.EventsResponse{
		Data: dtos,
		Pagination: api.PaginationDTO{
			Limit:  result.Pagination.Limit,
			Offset: result.Pagination.Offset,
			Total:  result.Pagination.Total,
		},
	}

	writeJSON(w, http.StatusOK, response)
}

// GetEventsGeoJSON handles GET /v1/events/geojson
func (h *Handler) GetEventsGeoJSON(w http.ResponseWriter, r *http.Request) {
	params, err := parseQueryParams(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_PARAMETER", err.Error(), "")
		return
	}

	result, err := h.queryService.QueryEvents(r.Context(), params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "QUERY_FAILED", err.Error(), "")
		return
	}

	geoJSON := api.ToGeoJSON(result.Events)
	writeJSON(w, http.StatusOK, geoJSON)
}

// parseQueryParams extracts query parameters from request
func parseQueryParams(r *http.Request) (query.QueryParams, error) {
	q := r.URL.Query()
	params := query.QueryParams{
		Limit:  100, // default
		Offset: 0,
	}

	// Parse magnitude filters
	if minMag := q.Get("minMagnitude"); minMag != "" {
		val, err := strconv.ParseFloat(minMag, 64)
		if err != nil {
			return params, err
		}
		params.MinMagnitude = &val
	}

	if maxMag := q.Get("maxMagnitude"); maxMag != "" {
		val, err := strconv.ParseFloat(maxMag, 64)
		if err != nil {
			return params, err
		}
		params.MaxMagnitude = &val
	}

	// Parse event type
	if eventType := q.Get("eventType"); eventType != "" {
		params.EventType = &eventType
	}

	// Parse spatial parameters
	if lat := q.Get("lat"); lat != "" {
		val, err := strconv.ParseFloat(lat, 64)
		if err != nil {
			return params, err
		}
		params.Latitude = &val
	}

	if lng := q.Get("lng"); lng != "" {
		val, err := strconv.ParseFloat(lng, 64)
		if err != nil {
			return params, err
		}
		params.Longitude = &val
	}

	if radius := q.Get("radiusKm"); radius != "" {
		val, err := strconv.ParseFloat(radius, 64)
		if err != nil {
			return params, err
		}
		params.RadiusKm = &val
	}

	// Parse time range
	if fromTime := q.Get("fromTime"); fromTime != "" {
		val, err := time.Parse(time.RFC3339, fromTime)
		if err != nil {
			return params, err
		}
		params.FromTime = &val
	}

	if toTime := q.Get("toTime"); toTime != "" {
		val, err := time.Parse(time.RFC3339, toTime)
		if err != nil {
			return params, err
		}
		params.ToTime = &val
	}

	// Parse pagination
	if limit := q.Get("limit"); limit != "" {
		val, err := strconv.Atoi(limit)
		if err != nil {
			return params, err
		}
		params.Limit = val
	}

	if offset := q.Get("offset"); offset != "" {
		val, err := strconv.Atoi(offset)
		if err != nil {
			return params, err
		}
		params.Offset = val
	}

	return params, nil
}

// writeJSON writes JSON response
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes error response
func writeError(w http.ResponseWriter, status int, code, message, field string) {
	response := api.ErrorResponse{
		Error: api.ErrorDetail{
			Code:    code,
			Message: message,
			Field:   field,
		},
	}
	writeJSON(w, status, response)
}
```

### 4. Create Router
Create `internal/infrastructure/http/router.go`:
```go
package http

import (
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
)

// NewRouter creates a configured HTTP router
func NewRouter(handler *Handler, logger *slog.Logger, corsOrigins []string) http.Handler {
	r := mux.NewRouter()

	// API v1 routes
	v1 := r.PathPrefix("/v1").Subrouter()
	v1.HandleFunc("/health", handler.Health).Methods("GET")
	v1.HandleFunc("/events", handler.ListEvents).Methods("GET")
	v1.HandleFunc("/events/geojson", handler.GetEventsGeoJSON).Methods("GET")

	// Apply middleware
	var h http.Handler = r
	h = LoggingMiddleware(logger)(h)
	if len(corsOrigins) > 0 {
		h = CORSMiddleware(corsOrigins)(h)
	}

	return h
}
```

### 5. Create Handler Tests
Create `internal/infrastructure/http/handlers_test.go`:
```go
package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourusername/geopulse/internal/application/query"
	"github.com/yourusername/geopulse/internal/domain/event"
	"github.com/yourusername/geopulse/internal/infrastructure/persistence"
	"github.com/yourusername/geopulse/internal/interfaces/api"
	"github.com/yourusername/geopulse/tests/testutil"
)

func TestHealthHandler(t *testing.T) {
	handler := &Handler{}
	
	req := httptest.NewRequest("GET", "/v1/health", nil)
	w := httptest.NewRecorder()

	handler.Health(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["status"] != "ok" {
		t.Errorf("Expected status ok, got %v", response["status"])
	}
}

func TestListEventsHandler(t *testing.T) {
	// Setup
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	repo := persistence.NewSQLiteRepository(&persistence.Database{DB: db})
	
	// Insert test data
	ctx := context.Background()
	loc, _ := event.NewLocation(37.7749, -122.4194)
	mag, _ := event.NewMagnitude(5.5)
	evt, _ := event.NewEvent("USGS", "test1", event.EventTypeEarthquake, time.Now(), loc, mag)
	repo.Save(ctx, evt)

	queryService := query.NewService(repo)
	handler := NewHandler(queryService)

	// Test request
	req := httptest.NewRequest("GET", "/v1/events?limit=10", nil)
	w := httptest.NewRecorder()

	handler.ListEvents(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var response api.EventsResponse
	json.NewDecoder(w.Body).Decode(&response)

	if len(response.Data) == 0 {
		t.Error("Expected events in response")
	}

	if response.Pagination.Total == 0 {
		t.Error("Expected non-zero total count")
	}
}

func TestGeoJSONHandler(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	repo := persistence.NewSQLiteRepository(&persistence.Database{DB: db})
	
	ctx := context.Background()
	loc, _ := event.NewLocation(37.7749, -122.4194)
	mag, _ := event.NewMagnitude(5.5)
	evt, _ := event.NewEvent("USGS", "test1", event.EventTypeEarthquake, time.Now(), loc, mag)
	repo.Save(ctx, evt)

	queryService := query.NewService(repo)
	handler := NewHandler(queryService)

	req := httptest.NewRequest("GET", "/v1/events/geojson", nil)
	w := httptest.NewRecorder()

	handler.GetEventsGeoJSON(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var geoJSON api.GeoJSONFeatureCollection
	json.NewDecoder(w.Body).Decode(&geoJSON)

	if geoJSON.Type != "FeatureCollection" {
		t.Errorf("Expected FeatureCollection, got %s", geoJSON.Type)
	}

	if len(geoJSON.Features) == 0 {
		t.Error("Expected features in GeoJSON")
	}
}
```

### 6. Run HTTP Tests
```powershell
go test ./internal/infrastructure/http -v
```

## Success Criteria
- ✓ DTOs created for API responses
- ✓ Middleware for logging and CORS
- ✓ Health check endpoint
- ✓ Events list endpoint with pagination
- ✓ GeoJSON endpoint
- ✓ Error handling implemented
- ✓ All HTTP tests pass

## Next Step
Proceed to **Step08-ConfigurationAndMain.md** to wire everything together in main.go.
