package event

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// Helper time values for tests
	testTime1 = time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	testTime2 = time.Date(2024, 2, 20, 14, 45, 0, 0, time.UTC)
	testTime3 = time.Date(2024, 3, 10, 8, 15, 0, 0, time.UTC)

	// Common test value objects (must be valid)
	testLocationLA, _    = NewLocation(34.05, -118.25, 10.0)
	testLocationTokyo, _ = NewLocation(35.68, 139.76, 50.0)
	testLocationParis, _ = NewLocation(48.85, 2.35, 5.0)
	testLocationDeep, _  = NewLocation(19.43, -99.13, 700.0)

	testMagSmall, _    = NewMagnitude(2.5, "ml")
	testMagModerate, _ = NewMagnitude(5.0, "mw")
	testMagLarge, _    = NewMagnitude(7.2, "mw")
	testMagNegative, _ = NewMagnitude(-0.5, "ml")

	testTypeEarthquake, _ = NewType("earthquake")
	testTypeExplosion, _  = NewType("explosion")
	testTypeOther, _      = NewType("unknown")
)

// validEvents contains test cases for valid event configurations (shared fixture)
var validEvents = []struct {
	name      string
	id        string
	location  Location
	place     string
	magnitude Magnitude
	eventType Type
	eventTime time.Time
	status    string
}{
	{
		name:      "Los Angeles earthquake",
		id:        "us1000abc1",
		location:  testLocationLA,
		place:     "5 km NW of Los Angeles, CA",
		magnitude: testMagModerate,
		eventType: testTypeEarthquake,
		eventTime: testTime1,
		status:    "reviewed",
	},
	{
		name:      "Tokyo deep earthquake",
		id:        "us2000xyz2",
		location:  testLocationTokyo,
		place:     "20 km E of Tokyo, Japan",
		magnitude: testMagLarge,
		eventType: testTypeEarthquake,
		eventTime: testTime2,
		status:    "automatic",
	},
	{
		name:      "Paris explosion",
		id:        "us3000def3",
		location:  testLocationParis,
		place:     "Paris, France",
		magnitude: testMagSmall,
		eventType: testTypeExplosion,
		eventTime: testTime3,
		status:    "reviewed",
	},
	{
		name:      "Mexico deep event",
		id:        "us4000ghi4",
		location:  testLocationDeep,
		place:     "Mexico City, Mexico",
		magnitude: testMagModerate,
		eventType: testTypeEarthquake,
		eventTime: testTime1,
		status:    "automatic",
	},
	{
		name:      "Small magnitude event",
		id:        "us5000jkl5",
		location:  testLocationLA,
		place:     "Southern California",
		magnitude: testMagSmall,
		eventType: testTypeEarthquake,
		eventTime: testTime2,
		status:    "reviewed",
	},
	{
		name:      "Negative magnitude (precursor)",
		id:        "us6000mno6",
		location:  testLocationParis,
		place:     "Central France",
		magnitude: testMagNegative,
		eventType: testTypeOther,
		eventTime: testTime3,
		status:    "automatic",
	},
}

// invalidEvents contains test cases for invalid event configurations
var invalidEvents = []struct {
	name      string
	id        string
	location  Location
	place     string
	magnitude Magnitude
	eventType Type
	eventTime time.Time
	status    string
	wantErr   string
}{
	{
		name:      "Empty ID",
		id:        "",
		location:  testLocationLA,
		place:     "Los Angeles, CA",
		magnitude: testMagModerate,
		eventType: testTypeEarthquake,
		eventTime: testTime1,
		status:    "reviewed",
		wantErr:   "event ID cannot be empty",
	},
	{
		name:      "Zero time",
		id:        "us7000pqr7",
		location:  testLocationLA,
		place:     "Los Angeles, CA",
		magnitude: testMagModerate,
		eventType: testTypeEarthquake,
		eventTime: time.Time{}, // Zero value
		status:    "reviewed",
		wantErr:   "event time cannot be zero",
	},
	{
		name:      "Empty status",
		id:        "us8000stu8",
		location:  testLocationLA,
		place:     "Los Angeles, CA",
		magnitude: testMagModerate,
		eventType: testTypeEarthquake,
		eventTime: testTime1,
		status:    "",
		wantErr:   "event status cannot be empty",
	},
}

// we should follow the DRY principle and use shared fixtures for valid event configurations in our tests. This allows us to easily add more test cases without duplicating the setup code for creating valid events.

func TestNewEvent(t *testing.T) {
	t.Run("Valid Events", func(t *testing.T) {
		for _, tc := range validEvents {
			t.Run(tc.name, func(t *testing.T) {
				got, err := NewEvent(
					tc.id,
					tc.location,
					tc.place,
					tc.magnitude,
					tc.eventType,
					tc.eventTime,
					tc.status,
				)

				require.NoError(t, err, "NewEvent should not return error for valid input")

				assert.Equal(t, tc.id, got.ID())
				assert.Equal(t, tc.location, got.Location())
				assert.Equal(t, tc.place, got.Place())
				assert.Equal(t, tc.magnitude, got.Magnitude())
				assert.Equal(t, tc.eventType, got.Type())
				assert.Equal(t, tc.eventTime, got.Time())
				assert.Equal(t, tc.status, got.Status())
				assert.False(t, got.Time().IsZero(), "Event time should not be zero")
			})
		}
	})

	t.Run("Invalid Events", func(t *testing.T) {
		for _, tc := range invalidEvents {
			t.Run(tc.name, func(t *testing.T) {
				got, err := NewEvent(
					tc.id,
					tc.location,
					tc.place,
					tc.magnitude,
					tc.eventType,
					tc.eventTime,
					tc.status,
				)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
				assert.Nil(t, got, "Should return nil on error")
			})
		}
	})
}

func TestEvent_String(t *testing.T) {
	t.Run("String format validation", func(t *testing.T) {
		for _, tc := range validEvents {
			t.Run(tc.name, func(t *testing.T) {
				got, err := NewEvent(
					tc.id,
					tc.location,
					tc.place,
					tc.magnitude,
					tc.eventType,
					tc.eventTime,
					tc.status,
				)
				require.NoError(t, err)

				result := got.String()

				assert.Contains(t, result, tc.id, "String should contain event ID")
				assert.Contains(t, result, tc.place, "String should contain place")
				assert.Contains(t, result, tc.magnitude.String(), "String should contain magnitude")
				assert.Contains(t, result, tc.eventType.String(), "String should contain event type")
			})
		}
	})

	t.Run("Formate examples", func(t *testing.T) {
		tests := []struct {
			name         string
			fixture      int
			wantContains []string
		}{
			{
				name:    "Los Angeles earthquake format",
				fixture: 0, // index of the validEvents fixture
				wantContains: []string{
					"us1000abc1",
					"5 km NW of Los Angeles, CA",
					"5.0 mw",
					"earthquake",
				},
			},
			{
				name:    "Tokyo deep earthquake format",
				fixture: 1,
				wantContains: []string{
					"us2000xyz2",
					"20 km E of Tokyo, Japan",
					"7.2 mw",
					"earthquake",
				},
			},
			{
				name:    "Paris explosion format",
				fixture: 2,
				wantContains: []string{
					"us3000def3",
					"Paris, France",
					"2.5 ml",
					"explosion",
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tc := validEvents[tt.fixture]
				event, err := NewEvent(
					tc.id,
					tc.location,
					tc.place,
					tc.magnitude,
					tc.eventType,
					tc.eventTime,
					tc.status,
				)
				require.NoError(t, err)

				result := event.String()
				for _, want := range tt.wantContains {
					assert.Contains(t, result, want, "String should contain expected substring: %s", want)
				}
			})
		}
	})
}

func TestEvent_Getters(t *testing.T) {
	t.Run("All getters return correct values", func(t *testing.T) {
		for _, tc := range validEvents {
			t.Run(tc.name, func(t *testing.T) {
				event, err := NewEvent(
					tc.id,
					tc.location,
					tc.place,
					tc.magnitude,
					tc.eventType,
					tc.eventTime,
					tc.status,
				)
				require.NoError(t, err)

				assert.Equal(t, tc.id, event.ID())
				assert.Equal(t, tc.location, event.Location())
				assert.Equal(t, tc.place, event.Place())
				assert.Equal(t, tc.magnitude, event.Magnitude())
				assert.Equal(t, tc.eventType, event.Type())
				assert.Equal(t, tc.eventTime, event.Time())
				assert.Equal(t, tc.status, event.Status())
				assert.False(t, event.Updated().IsZero(), "Updated time should be set to current time")
			})
		}
	})
}

func TestEvent_IsSignificant(t *testing.T) {
	tests := []struct {
		name      string
		magnitude Magnitude
		want      bool
	}{
		{
			name:      "Magnitude below threshold",
			magnitude: testMagSmall, // 2.5 ml
			want:      false,
		},
		{
			name:      "Magnitude equal to threshold",
			magnitude: testMagModerate, // 5.0 mw
			want:      true,
		},
		{
			name:      "Magnitude above threshold",
			magnitude: testMagLarge, // 7.2 mw
			want:      true,
		},
		{
			name:      "Negative magnitude",
			magnitude: testMagNegative, // -0.5 ml
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := NewEvent(
				"test123",
				testLocationLA,
				"Test Location",
				tt.magnitude,
				testTypeEarthquake,
				testTime1,
				"reviewed",
			)
			require.NoError(t, err)

			assert.Equal(t, tt.want, event.IsSignificant(5.0))
		})
	}
}

func TestEvent_UpdateStatus(t *testing.T) {
	t.Run("Status updates correctly", func(t *testing.T) {
		original, err := NewEvent(
			"us1000abc",
			testLocationLA,
			"Los Angeles, CA",
			testMagModerate,
			testTypeEarthquake,
			testTime1,
			"reviewed",
		)

		require.NoError(t, err)

		newTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		updated := original.UpdateStatus("automatic", newTime)

		assert.Equal(t, "automatic", updated.Status())
		assert.Equal(t, newTime, updated.Updated())
		assert.Equal(t, original.ID(), updated.ID(), "ID should remain unchanged")
		assert.Equal(t, original.Location(), updated.Location(), "Location should remain unchanged")
		assert.Equal(t, original.Place(), updated.Place(), "Place should remain unchanged")
		assert.Equal(t, original.Magnitude(), updated.Magnitude(), "Magnitude should remain unchanged")
		assert.Equal(t, original.Type(), updated.Type(), "Type should remain unchanged")
		assert.Equal(t, original.Time(), updated.Time(), "Time should remain unchanged")
	})
}
