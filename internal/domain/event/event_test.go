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

type eventFixture struct {
	id        string
	location  Location
	place     string
	magnitude Magnitude
	eventType Type
	eventTime time.Time
	status    string
}

type invalidEventFixture struct {
	eventFixture
	wantErr string
}

// validEventFixtures returns fresh test cases for valid event configurations.
func validEventFixtures() map[string]eventFixture {
	return map[string]eventFixture{
		"Los Angeles earthquake": {
			id:        "us1000abc1",
			location:  testLocationLA,
			place:     "5 km NW of Los Angeles, CA",
			magnitude: testMagModerate,
			eventType: testTypeEarthquake,
			eventTime: testTime1,
			status:    "reviewed",
		},
		"Tokyo deep earthquake": {
			id:        "us2000xyz2",
			location:  testLocationTokyo,
			place:     "20 km E of Tokyo, Japan",
			magnitude: testMagLarge,
			eventType: testTypeEarthquake,
			eventTime: testTime2,
			status:    "automatic",
		},
		"Paris explosion": {
			id:        "us3000def3",
			location:  testLocationParis,
			place:     "Paris, France",
			magnitude: testMagSmall,
			eventType: testTypeExplosion,
			eventTime: testTime3,
			status:    "reviewed",
		},
		"Mexico deep event": {
			id:        "us4000ghi4",
			location:  testLocationDeep,
			place:     "Mexico City, Mexico",
			magnitude: testMagModerate,
			eventType: testTypeEarthquake,
			eventTime: testTime1,
			status:    "automatic",
		},
		"Small magnitude event": {
			id:        "us5000jkl5",
			location:  testLocationLA,
			place:     "Southern California",
			magnitude: testMagSmall,
			eventType: testTypeEarthquake,
			eventTime: testTime2,
			status:    "reviewed",
		},
		"Negative magnitude (precursor)": {
			id:        "us6000mno6",
			location:  testLocationParis,
			place:     "Central France",
			magnitude: testMagNegative,
			eventType: testTypeOther,
			eventTime: testTime3,
			status:    "automatic",
		},
	}
}

// invalidEventFixtures returns fresh test cases for invalid event configurations.
func invalidEventFixtures() map[string]invalidEventFixture {
	return map[string]invalidEventFixture{
		"Empty ID": {
			eventFixture: eventFixture{
				id:        "",
				location:  testLocationLA,
				place:     "Los Angeles, CA",
				magnitude: testMagModerate,
				eventType: testTypeEarthquake,
				eventTime: testTime1,
				status:    "reviewed",
			},
			wantErr: "event ID cannot be empty",
		},
		"Zero time": {
			eventFixture: eventFixture{
				id:        "us7000pqr7",
				location:  testLocationLA,
				place:     "Los Angeles, CA",
				magnitude: testMagModerate,
				eventType: testTypeEarthquake,
				eventTime: time.Time{},
				status:    "reviewed",
			},
			wantErr: "event time cannot be zero",
		},
		"Empty status": {
			eventFixture: eventFixture{
				id:        "us8000stu8",
				location:  testLocationLA,
				place:     "Los Angeles, CA",
				magnitude: testMagModerate,
				eventType: testTypeEarthquake,
				eventTime: testTime1,
				status:    "",
			},
			wantErr: "event status cannot be empty",
		},
	}
}

func TestNewEvent(t *testing.T) {
	t.Run("Valid Events", func(t *testing.T) {
		for name, tc := range validEventFixtures() {
			t.Run(name, func(t *testing.T) {
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
				assert.Equal(t, tc.eventTime, got.Time())
				assert.Equal(t, tc.status, got.Status())
				assert.False(t, got.Time().IsZero(), "Event time should not be zero")
			})
		}
	})

	t.Run("Invalid Events", func(t *testing.T) {
		for name, tc := range invalidEventFixtures() {
			t.Run(name, func(t *testing.T) {
				_, err := NewEvent(
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
			})
		}
	})
}

func TestEvent_String(t *testing.T) {
	t.Run("String format validation", func(t *testing.T) {
		for name, tc := range validEventFixtures() {
			t.Run(name, func(t *testing.T) {
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
				assert.Contains(t, result, tc.eventType.String(), "String should contain type")
			})
		}
	})

	t.Run("Format examples", func(t *testing.T) {
		tests := map[string]struct {
			fixture      eventFixture
			wantContains []string
		}{
			"Los Angeles earthquake format": {
				fixture: eventFixture{
					id:        "us1000abc1",
					location:  testLocationLA,
					place:     "5 km NW of Los Angeles, CA",
					magnitude: testMagModerate,
					eventType: testTypeEarthquake,
					eventTime: testTime1,
					status:    "reviewed",
				},
				wantContains: []string{
					"us1000abc1",
					"5 km NW of Los Angeles, CA",
					"5.0 mw",
					"earthquake",
				},
			},
			"Tokyo deep earthquake format": {
				fixture: eventFixture{
					id:        "us2000xyz2",
					location:  testLocationTokyo,
					place:     "20 km E of Tokyo, Japan",
					magnitude: testMagLarge,
					eventType: testTypeEarthquake,
					eventTime: testTime2,
					status:    "automatic",
				},
				wantContains: []string{
					"us2000xyz2",
					"20 km E of Tokyo, Japan",
					"7.2 mw",
					"earthquake",
				},
			},
			"Paris explosion format": {
				fixture: eventFixture{
					id:        "us3000def3",
					location:  testLocationParis,
					place:     "Paris, France",
					magnitude: testMagSmall,
					eventType: testTypeExplosion,
					eventTime: testTime3,
					status:    "reviewed",
				},
				wantContains: []string{
					"us3000def3",
					"Paris, France",
					"2.5 ml",
					"explosion",
				},
			},
		}

		for name, tt := range tests {
			t.Run(name, func(t *testing.T) {
				event, err := NewEvent(
					tt.fixture.id,
					tt.fixture.location,
					tt.fixture.place,
					tt.fixture.magnitude,
					tt.fixture.eventType,
					tt.fixture.eventTime,
					tt.fixture.status,
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
		for name, tc := range validEventFixtures() {
			t.Run(name, func(t *testing.T) {
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
			magnitude: testMagSmall,
			want:      false,
		},
		{
			name:      "Magnitude equal to threshold",
			magnitude: testMagModerate,
			want:      true,
		},
		{
			name:      "Magnitude above threshold",
			magnitude: testMagLarge,
			want:      true,
		},
		{
			name:      "Negative magnitude",
			magnitude: testMagNegative,
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

	t.Run("Original event remains unchanged", func(t *testing.T) {
		original, err := NewEvent(
			"us9000uvw9",
			testLocationParis,
			"Paris, France",
			testMagSmall,
			testTypeExplosion,
			testTime2,
			"reviewed",
		)
		require.NoError(t, err)

		originalUpdated := original.Updated()
		updated := original.UpdateStatus("automatic", testTime3)

		assert.Equal(t, "reviewed", original.Status())
		assert.Equal(t, originalUpdated, original.Updated())
		assert.NotEqual(t, original.Status(), updated.Status())
	})
}
