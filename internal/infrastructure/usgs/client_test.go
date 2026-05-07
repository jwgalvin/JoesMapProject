package usgs

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ptrFloat64(v float64) *float64 { return &v }

// validFeature returns a Feature with all fields populated and valid coordinates.
func validFeature() Feature {
	return Feature{
		ID:   "nc12345678",
		Type: "Feature",
		Properties: Properties{
			Mag:     ptrFloat64(3.5),
			Place:   "5km N of Somewhere, CA",
			Time:    1700000000000,
			URL:     "https://earthquake.usgs.gov/earthquakes/eventpage/nc12345678",
			MagType: "mw",
			Type:    "earthquake",
			Title:   "M 3.5 - 5km N of Somewhere, CA",
			Status:  "reviewed",
		},
		Geometry: Geometry{
			Type:        "Point",
			Coordinates: []float64{-122.5, 37.8, 10.0},
		},
	}
}

// feedJSON encodes a GeoJSONResponse to JSON for use in httptest handlers.
func feedJSON(t *testing.T, feed *GeoJSONResponse) []byte {
	t.Helper()
	b, err := json.Marshal(feed)
	require.NoError(t, err)
	return b
}

// newTestClient creates a Client pointed at the given test server URL.
func newTestClient(url string) *Client {
	return NewClient(url, 5*time.Second)
}

func TestConvertFeature(t *testing.T) {
	t.Run("Valid features", func(t *testing.T) {
		tests := []struct {
			name          string
			feature       Feature
			wantID        string
			wantLat       float64
			wantLon       float64
			wantDepth     float64
			wantMag       float64
			wantMagScale  string
			wantEventType string
			wantPlace     string
			wantStatus    string
			wantTitle     string
			wantURL       string
			wantTime      time.Time
		}{
			{
				name:          "All fields populated",
				feature:       validFeature(),
				wantID:        "nc12345678",
				wantLat:       37.8,
				wantLon:       -122.5,
				wantDepth:     10.0,
				wantMag:       3.5,
				wantMagScale:  "mw",
				wantEventType: "earthquake",
				wantPlace:     "5km N of Somewhere, CA",
				wantStatus:    "reviewed",
				wantTitle:     "M 3.5 - 5km N of Somewhere, CA",
				wantURL:       "https://earthquake.usgs.gov/earthquakes/eventpage/nc12345678",
				wantTime:      time.UnixMilli(1700000000000),
			},
			{
				name: "Two coordinates — depth defaults to 0",
				feature: func() Feature {
					f := validFeature()
					f.Geometry.Coordinates = []float64{-122.5, 37.8}
					return f
				}(),
				wantID:        "nc12345678",
				wantLat:       37.8,
				wantLon:       -122.5,
				wantDepth:     0.0,
				wantMag:       3.5,
				wantMagScale:  "mw",
				wantEventType: "earthquake",
				wantStatus:    "reviewed",
				wantTime:      time.UnixMilli(1700000000000),
			},
			{
				name: "Nil magnitude pointer — defaults to 0.0",
				feature: func() Feature {
					f := validFeature()
					f.Properties.Mag = nil
					return f
				}(),
				wantID:       "nc12345678",
				wantLat:      37.8,
				wantLon:      -122.5,
				wantDepth:    10.0,
				wantMag:      0.0,
				wantMagScale: "mw",
				wantStatus:   "reviewed",
				wantTime:     time.UnixMilli(1700000000000),
			},
			{
				name: "Empty status — defaults to automatic",
				feature: func() Feature {
					f := validFeature()
					f.Properties.Status = ""
					return f
				}(),
				wantID:     "nc12345678",
				wantLat:    37.8,
				wantLon:    -122.5,
				wantDepth:  10.0,
				wantMag:    3.5,
				wantStatus: "automatic",
				wantTime:   time.UnixMilli(1700000000000),
			},
			{
				name: "USGS quarry alias maps to quarry blast",
				feature: func() Feature {
					f := validFeature()
					f.Properties.Type = "quarry"
					return f
				}(),
				wantID:        "nc12345678",
				wantStatus:    "reviewed",
				wantEventType: "quarry blast",
				wantTime:      time.UnixMilli(1700000000000),
			},
			{
				name: "Chemical explosion maps to explosion",
				feature: func() Feature {
					f := validFeature()
					f.Properties.Type = "chemical explosion"
					return f
				}(),
				wantID:        "nc12345678",
				wantStatus:    "reviewed",
				wantEventType: "explosion",
				wantTime:      time.UnixMilli(1700000000000),
			},
			{
				name: "Unknown event type maps to other",
				feature: func() Feature {
					f := validFeature()
					f.Properties.Type = "mystery"
					return f
				}(),
				wantID:        "nc12345678",
				wantStatus:    "reviewed",
				wantEventType: "other",
				wantTime:      time.UnixMilli(1700000000000),
			},
			{
				name: "Unrecognized mag scale maps to unknown",
				feature: func() Feature {
					f := validFeature()
					f.Properties.MagType = "xyz"
					return f
				}(),
				wantID:       "nc12345678",
				wantStatus:   "reviewed",
				wantMag:      3.5,
				wantMagScale: "unknown",
				wantTime:     time.UnixMilli(1700000000000),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := convertFeature(&tt.feature)
				require.NoError(t, err)
				require.NotNil(t, got)

				assert.Equal(t, tt.wantID, got.ID())
				assert.Equal(t, tt.wantTime, got.Time())
				if tt.wantStatus != "" {
					assert.Equal(t, tt.wantStatus, got.Status())
				}
				if tt.wantLat != 0 || tt.wantLon != 0 {
					assert.Equal(t, tt.wantLat, got.Location().Latitude)
					assert.Equal(t, tt.wantLon, got.Location().Longitude)
					assert.Equal(t, tt.wantDepth, got.Location().Depth)
				}
				if tt.wantMagScale != "" {
					assert.Equal(t, tt.wantMag, got.Magnitude().Value())
					assert.Equal(t, tt.wantMagScale, got.Magnitude().Scale())
				}
				if tt.wantEventType != "" {
					assert.Equal(t, tt.wantEventType, got.Type().String())
				}
				if tt.wantPlace != "" {
					assert.Equal(t, tt.wantPlace, got.Place())
				}
				if tt.wantTitle != "" {
					assert.Equal(t, tt.wantTitle, got.Description())
				}
				if tt.wantURL != "" {
					assert.Equal(t, tt.wantURL, got.URL())
				}
			})
		}
	})

	t.Run("Invalid features", func(t *testing.T) {
		tests := []struct {
			name    string
			feature Feature
			wantErr string
		}{
			{
				name: "Fewer than 2 coordinates",
				feature: func() Feature {
					f := validFeature()
					f.Geometry.Coordinates = []float64{-122.5}
					return f
				}(),
				wantErr: "insufficient coordinates",
			},
			{
				name: "Empty coordinates slice",
				feature: func() Feature {
					f := validFeature()
					f.Geometry.Coordinates = []float64{}
					return f
				}(),
				wantErr: "insufficient coordinates",
			},
			{
				name: "Latitude out of range",
				feature: func() Feature {
					f := validFeature()
					f.Geometry.Coordinates = []float64{-122.5, 91.0, 10.0}
					return f
				}(),
				wantErr: "invalid location",
			},
			{
				name: "Longitude out of range",
				feature: func() Feature {
					f := validFeature()
					f.Geometry.Coordinates = []float64{200.0, 37.8, 10.0}
					return f
				}(),
				wantErr: "invalid location",
			},
			{
				name: "Depth out of range",
				feature: func() Feature {
					f := validFeature()
					f.Geometry.Coordinates = []float64{-122.5, 37.8, 2000.0}
					return f
				}(),
				wantErr: "invalid location",
			},
			{
				name: "Magnitude value too high",
				feature: func() Feature {
					f := validFeature()
					f.Properties.Mag = ptrFloat64(11.0)
					return f
				}(),
				wantErr: "invalid magnitude",
			},
			{
				name: "Magnitude value too low",
				feature: func() Feature {
					f := validFeature()
					f.Properties.Mag = ptrFloat64(-2.0)
					return f
				}(),
				wantErr: "invalid magnitude",
			},
			{
				name: "Empty event type",
				feature: func() Feature {
					f := validFeature()
					f.Properties.Type = ""
					return f
				}(),
				wantErr: "invalid event type",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := convertFeature(&tt.feature)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				assert.Nil(t, got)
			})
		}
	})
}

func TestFetchEvents(t *testing.T) {
	t.Run("Successful responses", func(t *testing.T) {
		tests := []struct {
			name      string
			feed      GeoJSONResponse
			wantCount int
			wantIDs   []string
		}{
			{
				name: "Single valid feature",
				feed: GeoJSONResponse{
					Type:     "FeatureCollection",
					Features: []Feature{validFeature()},
				},
				wantCount: 1,
				wantIDs:   []string{"nc12345678"},
			},
			{
				name: "Multiple valid features",
				feed: GeoJSONResponse{
					Type: "FeatureCollection",
					Features: []Feature{
						validFeature(),
						func() Feature {
							f := validFeature()
							f.ID = "us9876543"
							f.Geometry.Coordinates = []float64{139.76, 35.68, 50.0}
							return f
						}(),
					},
				},
				wantCount: 2,
				wantIDs:   []string{"nc12345678", "us9876543"},
			},
			{
				name: "Empty features list",
				feed: GeoJSONResponse{
					Type:     "FeatureCollection",
					Features: []Feature{},
				},
				wantCount: 0,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				body := feedJSON(t, &tt.feed)
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write(body)
				}))
				defer ts.Close()

				events, err := newTestClient(ts.URL).FetchEvents(context.Background())
				require.NoError(t, err)
				assert.Len(t, events, tt.wantCount)

				for i, wantID := range tt.wantIDs {
					assert.Equal(t, wantID, events[i].ID())
				}
			})
		}
	})

	t.Run("Invalid features are skipped", func(t *testing.T) {
		badFeature := validFeature()
		badFeature.ID = "bad1"
		badFeature.Geometry.Coordinates = []float64{-122.5} // insufficient

		goodFeature := validFeature()
		goodFeature.ID = "good1"

		feed := GeoJSONResponse{
			Type:     "FeatureCollection",
			Features: []Feature{badFeature, goodFeature},
		}
		body := feedJSON(t, &feed)

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(body)
		}))
		defer ts.Close()

		events, err := newTestClient(ts.URL).FetchEvents(context.Background())
		require.NoError(t, err)
		require.Len(t, events, 1)
		assert.Equal(t, "good1", events[0].ID())
	})

	t.Run("All features invalid returns empty slice", func(t *testing.T) {
		bad := validFeature()
		bad.Geometry.Coordinates = []float64{}

		feed := GeoJSONResponse{
			Type:     "FeatureCollection",
			Features: []Feature{bad},
		}
		body := feedJSON(t, &feed)

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(body)
		}))
		defer ts.Close()

		events, err := newTestClient(ts.URL).FetchEvents(context.Background())
		require.NoError(t, err)
		assert.Empty(t, events)
	})

	t.Run("HTTP error responses", func(t *testing.T) {
		tests := []struct {
			name       string
			statusCode int
		}{
			{"Internal server error", http.StatusInternalServerError},
			{"Not found", http.StatusNotFound},
			{"Service unavailable", http.StatusServiceUnavailable},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(tt.statusCode)
					_, _ = w.Write([]byte("error body"))
				}))
				defer ts.Close()

				events, err := newTestClient(ts.URL).FetchEvents(context.Background())
				require.Error(t, err)
				assert.Contains(t, err.Error(), "unexpected status")
				assert.Nil(t, events)
			})
		}
	})

	t.Run("Malformed JSON response", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"type": "FeatureCollection", "features": [INVALID`))
		}))
		defer ts.Close()

		events, err := newTestClient(ts.URL).FetchEvents(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode")
		assert.Nil(t, events)
	})

	t.Run("Cancelled context", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(feedJSON(t, &GeoJSONResponse{Features: []Feature{validFeature()}}))
		}))
		defer ts.Close()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		events, err := newTestClient(ts.URL).FetchEvents(ctx)
		require.Error(t, err)
		assert.Nil(t, events)
	})

	t.Run("Request headers are set correctly", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "application/json", r.Header.Get("Accept"))
			assert.Equal(t, "GeoPulse/1.0", r.Header.Get("User-Agent"))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(feedJSON(t, &GeoJSONResponse{}))
		}))
		defer ts.Close()

		_, err := newTestClient(ts.URL).FetchEvents(context.Background())
		require.NoError(t, err)
	})
}
