package event

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// validLocations contains test cases for valid location values (shared fixture)
var validLocations = []struct {
	name      string
	latitude  float64
	longitude float64
	depth     float64
}{
	{
		name:      "Los Angeles",
		latitude:  34.05,
		longitude: -118.25,
		depth:     10.0,
	},
	{
		name:      "Tokyo",
		latitude:  35.68,
		longitude: 139.76,
		depth:     50.0,
	},
	{
		name:      "Paris",
		latitude:  48.85,
		longitude: 2.35,
		depth:     0.0,
	},
	{
		name:      "Mexico City",
		latitude:  19.4326,
		longitude: -99.1332,
		depth:     700.0,
	},
	{
		name:      "Sydney",
		latitude:  -33.8688,
		longitude: -151.2093,
		depth:     20.0,
	},
	{
		name:      "Equator and Prime Meridian",
		latitude:  0.0,
		longitude: 0.0,
		depth:     0.0,
	},
	{
		name:      "High Precision Location",
		latitude:  48.8566,
		longitude: 2.3522,
		depth:     15.5,
	},
}

// invalidCases contains test cases for invalid location values
var invalidCases = []struct {
	name      string
	latitude  float64
	longitude float64
	depth     float64
	wantErr   string
}{
	{
		name:      "Latitude too low",
		latitude:  -91.0,
		longitude: 0.0,
		depth:     10.0,
		wantErr:   "latitude must be between -90 and 90",
	},
	{
		name:      "Latitude too high",
		latitude:  91.0,
		longitude: 0.0,
		depth:     10.0,
		wantErr:   "latitude must be between -90 and 90",
	},
	{
		name:      "Longitude too low",
		latitude:  0.0,
		longitude: -181.0,
		depth:     10.0,
		wantErr:   "longitude must be between -180 and 180",
	},
	{
		name:      "Longitude too high",
		latitude:  0.0,
		longitude: 181.0,
		depth:     10.0,
		wantErr:   "longitude must be between -180 and 180",
	},
	{
		name:      "Depth too shallow",
		latitude:  0.0,
		longitude: 0.0,
		depth:     -20.0,
		wantErr:   "depth must be between -10 and 1000",
	},
	{
		name:      "Depth too deep",
		latitude:  0.0,
		longitude: 0.0,
		depth:     1500.0,
		wantErr:   "depth must be between -10 and 1000",
	},
}

func TestNewLocation(t *testing.T) {
	t.Run("Valid locations", func(t *testing.T) {
		for _, tc := range validLocations {
			t.Run(tc.name, func(t *testing.T) {
				loc, err := NewLocation(tc.latitude, tc.longitude, tc.depth)
				assert.NoError(t, err, "NewLocation() should not return error for valid location")
				assert.NotNil(t, loc, "NewLocation() should not return nil for valid location")

				// Verify all fields are set correctly
				assert.Equal(t, tc.latitude, loc.Latitude, "Latitude mismatch")
				assert.Equal(t, tc.longitude, loc.Longitude, "Longitude mismatch")
				assert.Equal(t, tc.depth, loc.Depth, "Depth mismatch")
			})
		}
	})

	t.Run("Edge cases", func(t *testing.T) {
		edgeCases := []struct {
			name      string
			latitude  float64
			longitude float64
			depth     float64
		}{
			{
				name:      "Latitude at minimum boundary",
				latitude:  -90.0,
				longitude: 0.0,
				depth:     10.0,
			},
			{
				name:      "Latitude at maximum boundary",
				latitude:  90.0,
				longitude: 0.0,
				depth:     10.0,
			},
			{
				name:      "Longitude at minimum boundary",
				latitude:  0.0,
				longitude: -180.0,
				depth:     10.0,
			},
			{
				name:      "Longitude at maximum boundary",
				latitude:  0.0,
				longitude: 180.0,
				depth:     10.0,
			},
			{
				name:      "Depth at minimum boundary",
				latitude:  0.0,
				longitude: 0.0,
				depth:     -10.0,
			},
			{
				name:      "Depth at maximum boundary",
				latitude:  0.0,
				longitude: 0.0,
				depth:     1000.0,
			},
		}

		for _, tc := range edgeCases {
			t.Run(tc.name, func(t *testing.T) {
				loc, err := NewLocation(tc.latitude, tc.longitude, tc.depth)
				assert.NoError(t, err, "NewLocation() should accept edge case values")
				assert.Equal(t, tc.latitude, loc.Latitude, "Latitude mismatch")
				assert.Equal(t, tc.longitude, loc.Longitude, "Longitude mismatch")
				assert.Equal(t, tc.depth, loc.Depth, "Depth mismatch")
			})
		}
	})

	t.Run("Invalid cases", func(t *testing.T) {
		for _, tc := range invalidCases {
			t.Run(tc.name, func(t *testing.T) {
				loc, err := NewLocation(tc.latitude, tc.longitude, tc.depth)
				assert.Error(t, err, "NewLocation() should return error for invalid location")
				assert.Contains(t, err.Error(), tc.wantErr, "Error message mismatch")
				assert.Equal(t, Location{}, loc, "NewLocation() should return zero value on error")
			})
		}
	})
}

func TestLocation_String(t *testing.T) {
	for _, tc := range validLocations {
		t.Run(tc.name, func(t *testing.T) {
			loc, err := NewLocation(tc.latitude, tc.longitude, tc.depth)
			assert.NoError(t, err, "NewLocation() should not return error")

			got := loc.String()
			// Verify format includes all components
			assert.Contains(t, got, "Lat:", "String should contain latitude label")
			assert.Contains(t, got, "Lon:", "String should contain longitude label")
			assert.Contains(t, got, "Depth:", "String should contain depth label")
			assert.Contains(t, got, "km", "String should contain km unit")
		})
	}

	// Test specific formatting examples
	t.Run("Format examples", func(t *testing.T) {
		tests := []struct {
			name      string
			latitude  float64
			longitude float64
			depth     float64
			want      string
		}{
			{
				name:      "Los Angeles format",
				latitude:  34.05,
				longitude: -118.25,
				depth:     10.0,
				want:      "Lat: 34.0500, Lon: -118.2500, Depth: 10.00 km",
			},
			{
				name:      "Zero values format",
				latitude:  0.0,
				longitude: 0.0,
				depth:     0.0,
				want:      "Lat: 0.0000, Lon: 0.0000, Depth: 0.00 km",
			},
			{
				name:      "Negative depth format",
				latitude:  35.68,
				longitude: 139.76,
				depth:     -5.0,
				want:      "Lat: 35.6800, Lon: 139.7600, Depth: -5.00 km",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				loc, err := NewLocation(tt.latitude, tt.longitude, tt.depth)
				assert.NoError(t, err)
				assert.Equal(t, tt.want, loc.String(), "String() format mismatch")
			})
		}
	})
}

func TestLocation_LatitudeValue(t *testing.T) {
	for _, tc := range validLocations {
		t.Run(tc.name, func(t *testing.T) {
			loc, err := NewLocation(tc.latitude, tc.longitude, tc.depth)
			assert.NoError(t, err, "NewLocation() should not return error")
			assert.Equal(t, tc.latitude, loc.LatitudeValue(), "LatitudeValue() mismatch")
		})
	}

	// Test edge cases
	t.Run("Boundary values", func(t *testing.T) {
		tests := []struct {
			name     string
			latitude float64
		}{
			{"Minimum latitude", -90.0},
			{"Maximum latitude", 90.0},
			{"Zero latitude", 0.0},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				loc, err := NewLocation(tt.latitude, 0.0, 10.0)
				assert.NoError(t, err)
				assert.Equal(t, tt.latitude, loc.LatitudeValue())
			})
		}
	})
}

func TestLocation_LongitudeValue(t *testing.T) {
	for _, tc := range validLocations {
		t.Run(tc.name, func(t *testing.T) {
			loc, err := NewLocation(tc.latitude, tc.longitude, tc.depth)
			assert.NoError(t, err, "NewLocation() should not return error")
			assert.Equal(t, tc.longitude, loc.LongitudeValue(), "LongitudeValue() mismatch")
		})
	}

	// Test edge cases
	t.Run("Boundary values", func(t *testing.T) {
		tests := []struct {
			name      string
			longitude float64
		}{
			{"Minimum longitude", -180.0},
			{"Maximum longitude", 180.0},
			{"Zero longitude", 0.0},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				loc, err := NewLocation(0.0, tt.longitude, 10.0)
				assert.NoError(t, err)
				assert.Equal(t, tt.longitude, loc.LongitudeValue())
			})
		}
	})
}

func TestLocation_DepthValue(t *testing.T) {
	for _, tc := range validLocations {
		t.Run(tc.name, func(t *testing.T) {
			loc, err := NewLocation(tc.latitude, tc.longitude, tc.depth)
			assert.NoError(t, err, "NewLocation() should not return error")
			assert.Equal(t, tc.depth, loc.DepthValue(), "DepthValue() mismatch")
		})
	}

	t.Run("Boundary values", func(t *testing.T) {
		tests := []struct {
			name  string
			depth float64
		}{
			{"Minimum depth", -10.0},
			{"Maximum depth", 1000.0},
			{"Zero depth", 0.0},
			{"Negative depth (above sea level)", -5.0},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				loc, err := NewLocation(0.0, 0.0, tt.depth)
				assert.NoError(t, err)
				assert.Equal(t, tt.depth, loc.DepthValue())
			})
		}
	})
}

func TestLocation_IsShallow(t *testing.T) {
	tests := []struct {
		name  string
		depth float64
		want  bool
	}{
		{"Very shallow (10 km)", 10.0, true},
		{"Shallow (50 km)", 50.0, true},
		{"Just below threshold (69 km)", 69.0, true},
		{"At threshold (70 km)", 70.0, false},
		{"Just above threshold (71 km)", 71.0, false},
		{"Deep (300 km)", 300.0, false},
		{"Very deep (700 km)", 700.0, false},
		{"Negative depth (above sea level)", -5.0, true},
		{"Zero depth (surface)", 0.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc, err := NewLocation(0.0, 0.0, tt.depth)
			assert.NoError(t, err, "NewLocation() should not return error")
			assert.Equal(t, tt.want, loc.IsShallow(), "IsShallow() mismatch for depth %.1f", tt.depth)
		})
	}
}

func TestLocation_IsDeep(t *testing.T) {
	tests := []struct {
		name  string
		depth float64
		want  bool
	}{
		{"Shallow (10 km)", 10.0, false},
		{"Moderate (50 km)", 50.0, false},
		{"Intermediate (100 km)", 100.0, false},
		{"Just below threshold (299 km)", 299.0, false},
		{"At threshold (300 km)", 300.0, true},
		{"Just above threshold (301 km)", 301.0, true},
		{"Deep (500 km)", 500.0, true},
		{"Very deep (700 km)", 700.0, true},
		{"Maximum depth (1000 km)", 1000.0, true},
		{"Negative depth (above sea level)", -5.0, false},
		{"Zero depth (surface)", 0.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc, err := NewLocation(0.0, 0.0, tt.depth)
			assert.NoError(t, err, "NewLocation() should not return error")
			assert.Equal(t, tt.want, loc.IsDeep(), "IsDeep() mismatch for depth %.1f", tt.depth)
		})
	}
}
