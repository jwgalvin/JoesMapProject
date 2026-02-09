package event

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEventType(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectedErr bool
	}{
		{
			name:        "Valid Earthquake",
			input:       "earthquake",
			expected:    EventTypeEarthQuake,
			expectedErr: false,
		},
		{
			name:        "Valid Quarry Blast",
			input:       "quarry blast",
			expected:    EventTypeQuarryBlast,
			expectedErr: false,
		},
		{
			name:        "Valid Explosion",
			input:       "explosion",
			expected:    EventTypeExplosion,
			expectedErr: false,
		},
		{
			name:        "Valid Other",
			input:       "other",
			expected:    EventTypeOther,
			expectedErr: false,
		},
		{
			name:        "Valid Landslide",
			input:       "landslide",
			expected:    EventTypeLandslide,
			expectedErr: false,
		},
		{
			name:        "Valid Volcanic Eruption",
			input:       "volcanic eruption",
			expected:    EventTypeVolcanicEruption,
			expectedErr: false,
		},
		{
			name:        "Valid Ice Quake",
			input:       "ice quake",
			expected:    EventTypeIceQuake,
			expectedErr: false,
		},
		{
			name:        "Unknown Type",
			input:       "meteor strike",
			expected:    EventTypeOther,
			expectedErr: false,
		},
		{
			name:        "Empty String",
			input:       "",
			expected:    "",
			expectedErr: true,
		},
		{
			name:        "Whitespace String",
			input:       "   ",
			expected:    EventTypeOther,
			expectedErr: false,
		},
		{
			name:        "Case Insensitivity",
			input:       "EARTHQUAKE",
			expected:    EventTypeEarthQuake,
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewType(tt.input)

			if tt.expectedErr {
				require.Error(t, err, "expected error but got none")
			} else {
				require.NoError(t, err, "unexpected error: %v", err)
				assert.Equal(t, tt.expected, result.String(), "event type mismatch")
			}
		})
	}
}

func TestType_String(t *testing.T) {
	tests := []struct {
		name     string
		input    Type
		expected string
	}{
		{
			name:     "Earthquake",
			input:    Type{value: EventTypeEarthQuake},
			expected: EventTypeEarthQuake,
		},
		{
			name:     "Quarry Blast",
			input:    Type{value: EventTypeQuarryBlast},
			expected: EventTypeQuarryBlast,
		},
		{
			name:     "Explosion",
			input:    Type{value: EventTypeExplosion},
			expected: EventTypeExplosion,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			et, err := NewType(tt.input.value)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, et.String())
		})
	}
}

func TestType_IsKnown(t *testing.T) {
	tests := []struct {
		name     string
		input    Type
		expected bool
	}{
		{
			name:     "Known Type",
			input:    Type{value: EventTypeEarthQuake},
			expected: true,
		},
		{
			name:     "Unknown Type",
			input:    Type{value: EventTypeOther},
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			et, err := NewType(tt.input.value)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, et.IsKnown())
		})
	}
}
