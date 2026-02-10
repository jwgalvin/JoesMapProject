package event

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Shared fixtures at package level
var (
	validMagnitudes = []struct {
		name   string
		minMag float64
		maxMag float64
	}{
		{name: "Small earthquakes", minMag: 0.0, maxMag: 3.0},
		{name: "Moderate earthquakes", minMag: 4.0, maxMag: 6.0},
		{name: "Large earthquakes", minMag: 7.0, maxMag: 10.0},
	}

	validTimeRanges = []struct {
		name  string
		start time.Time
		end   time.Time
	}{
		{name: "Last 24 hours", start: time.Now().Add(-24 * time.Hour), end: time.Now()},
		{name: "Last week", start: time.Now().Add(-7 * 24 * time.Hour), end: time.Now()},
	}

	validPagination = []struct {
		name   string
		limit  int
		offset int
	}{
		{name: "First page", limit: 20, offset: 0},
		{name: "Second page", limit: 20, offset: 20},
		{name: "Large page", limit: 100, offset: 0},
	}

	validProximityLocations = []struct {
		name      string
		latitude  float64
		longitude float64
		depth     float64
	}{
		{name: "Los Angeles", latitude: 34.05, longitude: -118.25, depth: 10.0},
	}

	validTypeFilters = []struct {
		name  string
		types []Type
	}{
		{name: "Earthquake only", types: []Type{mustType("earthquake")}},
		{name: "Earthquake and explosion", types: []Type{mustType("earthquake"), mustType("explosion")}},
	}

	validStatuses = []struct {
		name     string
		statuses []string
	}{
		{name: "Reviewed", statuses: []string{"reviewed"}},
		{name: "Automatic and reviewed", statuses: []string{"automatic", "reviewed"}},
	}

	validSorts = []struct {
		name      string
		orderBy   string
		ascending bool
	}{
		{name: "Sort by time desc", orderBy: "time", ascending: false},
		{name: "Sort by magnitude asc", orderBy: "magnitude", ascending: true},
		{name: "Sort by depth asc", orderBy: "depth", ascending: true},
		{name: "Sort by place desc", orderBy: "place", ascending: false},
	}
)

func TestNewQueryCriteria(t *testing.T) {
	criteria := NewQueryCriteria()

	assert.Equal(t, 100, criteria.Limit)
	assert.Equal(t, 0, criteria.Offset)
	assert.Equal(t, "time", criteria.OrderBy)
	assert.False(t, criteria.Ascending)
}

func TestQueryCriteria_WithMagnitudeRange(t *testing.T) {
	t.Run("Valid ranges", func(t *testing.T) {
		for _, tc := range validMagnitudes {
			t.Run(tc.name, func(t *testing.T) {
				criteria := NewQueryCriteria()
				err := criteria.WithMagnitudeRange(tc.minMag, tc.maxMag)

				require.NoError(t, err)
				assert.Equal(t, tc.minMag, *criteria.MinMagnitude)
				assert.Equal(t, tc.maxMag, *criteria.MaxMagnitude)
			})
		}
	})

	t.Run("Invalid ranges", func(t *testing.T) {
		invalidCases := []struct {
			name   string
			minMag float64
			maxMag float64
		}{
			{name: "Min too low", minMag: -2.0, maxMag: 5.0},
			{name: "Max too high", minMag: 5.0, maxMag: 11.0},
			{name: "Max less than min", minMag: 6.0, maxMag: 4.0},
		}

		for _, tc := range invalidCases {
			t.Run(tc.name, func(t *testing.T) {
				criteria := NewQueryCriteria()
				err := criteria.WithMagnitudeRange(tc.minMag, tc.maxMag)

				assert.Error(t, err)
			})
		}
	})
}

func TestQueryCriteria_WithTimeRange(t *testing.T) {
	t.Run("Valid ranges", func(t *testing.T) {
		for _, tc := range validTimeRanges {
			t.Run(tc.name, func(t *testing.T) {
				criteria := NewQueryCriteria()
				err := criteria.WithTimeRange(tc.start, tc.end)

				require.NoError(t, err)
				assert.Equal(t, tc.start, *criteria.StartTime)
				assert.Equal(t, tc.end, *criteria.EndTime)
			})
		}
	})

	t.Run("End before start", func(t *testing.T) {
		criteria := NewQueryCriteria()
		start := time.Now()
		end := start.Add(-1 * time.Hour)

		err := criteria.WithTimeRange(start, end)
		assert.Error(t, err)
	})
}

func TestQueryCriteria_WithProximity(t *testing.T) {
	t.Run("Valid Proximity", func(t *testing.T) {
		for _, tc := range validProximityLocations {
			t.Run(tc.name, func(t *testing.T) {
				loc, err := NewLocation(tc.latitude, tc.longitude, tc.depth)
				require.NoError(t, err)

				criteria := NewQueryCriteria()
				err = criteria.WithProximity(loc, 100.0)
				require.NoError(t, err)
				assert.Equal(t, &loc, criteria.Location)
				assert.Equal(t, 100.0, *criteria.RadiusKm)
			})
		}
	})

	t.Run("Negative Radius", func(t *testing.T) {
		loc, err := NewLocation(validProximityLocations[0].latitude, validProximityLocations[0].longitude, validProximityLocations[0].depth)
		require.NoError(t, err)

		criteria := NewQueryCriteria()
		err = criteria.WithProximity(loc, -50.0)
		assert.Error(t, err)
	})

	t.Run("Radius Exceeds Maximum", func(t *testing.T) {
		loc, err := NewLocation(validProximityLocations[0].latitude, validProximityLocations[0].longitude, validProximityLocations[0].depth)
		require.NoError(t, err)

		criteria := NewQueryCriteria()
		err = criteria.WithProximity(loc, 25000.0)
		assert.Error(t, err)
	})
}

func TestQueryCriteria_WithEventTypes(t *testing.T) {
	t.Run("Valid Event Types", func(t *testing.T) {
		for _, tc := range validTypeFilters {
			t.Run(tc.name, func(t *testing.T) {
				criteria := NewQueryCriteria()
				err := criteria.WithEventTypes(tc.types...)
				require.NoError(t, err)
				assert.Equal(t, tc.types, criteria.EventTypes)
			})
		}
	})

	t.Run("No Event Types", func(t *testing.T) {
		criteria := NewQueryCriteria()
		err := criteria.WithEventTypes()
		assert.Error(t, err)
	})
}

func TestQueryCriteria_WithStatuses(t *testing.T) {
	t.Run("Valid statuses", func(t *testing.T) {
		for _, tc := range validStatuses {
			t.Run(tc.name, func(t *testing.T) {
				criteria := NewQueryCriteria()
				err := criteria.WithStatuses(tc.statuses...)
				require.NoError(t, err)
				assert.Equal(t, tc.statuses, criteria.Statuses)
			})
		}
	})

	t.Run("No statuses", func(t *testing.T) {
		criteria := NewQueryCriteria()
		err := criteria.WithStatuses()
		assert.Error(t, err)
	})
}

func TestQueryCriteria_WithPagination(t *testing.T) {
	t.Run("Valid pagination", func(t *testing.T) {
		for _, tc := range validPagination {
			t.Run(tc.name, func(t *testing.T) {
				criteria := NewQueryCriteria()
				err := criteria.WithPagination(tc.limit, tc.offset)
				require.NoError(t, err)
				assert.Equal(t, tc.limit, criteria.Limit)
				assert.Equal(t, tc.offset, criteria.Offset)
			})
		}
	})

	t.Run("Invalid pagination", func(t *testing.T) {
		invalidCases := []struct {
			name   string
			limit  int
			offset int
		}{
			{name: "Negative limit", limit: -1, offset: 0},
			{name: "Limit too large", limit: 1001, offset: 0},
			{name: "Negative offset", limit: 10, offset: -5},
		}

		for _, tc := range invalidCases {
			t.Run(tc.name, func(t *testing.T) {
				criteria := NewQueryCriteria()
				err := criteria.WithPagination(tc.limit, tc.offset)
				assert.Error(t, err)
			})
		}
	})
}

func TestQueryCriteria_WithSort(t *testing.T) {
	t.Run("Valid sort fields", func(t *testing.T) {
		for _, tc := range validSorts {
			t.Run(tc.name, func(t *testing.T) {
				criteria := NewQueryCriteria()
				err := criteria.WithSort(tc.orderBy, tc.ascending)
				require.NoError(t, err)
				assert.Equal(t, tc.orderBy, criteria.OrderBy)
				assert.Equal(t, tc.ascending, criteria.Ascending)
			})
		}
	})

	t.Run("Invalid sort field", func(t *testing.T) {
		criteria := NewQueryCriteria()
		err := criteria.WithSort("invalid_field", true)
		assert.Error(t, err)
	})
}

func mustType(value string) Type {
	result, err := NewType(value)
	if err != nil {
		panic(err)
	}
	return result
}
