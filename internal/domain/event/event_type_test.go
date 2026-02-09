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
            name: "Valid Earthquake",
            input: "earthquake",
            expected: EventTypeEarthQuake,
            expectedErr: false,
        },
		{
            name: "Valid Quarry Blast",
            input: "quarry blast",
            expected: EventTypeQuarryBlast,
            expectedErr: false,
        },
		{
            name: "Valid Explosion",
            input: "explosion",
            expected: EventTypeExplosion,
            expectedErr: false,
        },
		{
            name: "Valid Other",
            input: "other",
            expected: EventTypeOther,
            expectedErr: false,
        },
		{
            name: "Valid Landslide",
            input: "landslide",
            expected: EventTypeLandslide,
            expectedErr: false,
        },
		{
            name: "Valid Volcanic Eruption",
            input: "volcanic eruption",
            expected: EventTypeVolcanicEruption,
            expectedErr: false,
        },
		{
            name: "Valid Ice Quake",
            input: "ice quake",
            expected: EventTypeIceQuake,
            expectedErr: false,
        },
		{
            name: "Unknown Type",
            input: "meteor strike",
            expected: EventTypeOther,
            expectedErr: false,
        },
		{
            name: "Empty String",
            input: "",
            expected: "",
            expectedErr: true,
        },
		{
            name: "Whitespace String",
            input: "   ",
            expected: EventTypeOther,
            expectedErr: false,
        },
		{
            name: "Case Insensitivity",
            input: "EARTHQUAKE",
            expected: EventTypeEarthQuake,
            expectedErr: false,
        },
	}

	for _, tt := range tests {
		 t.Run(tt.name, func(t *testing.T) {
            result, err := NewEventType(tt.input)

            if tt.expectedErr {
                require.Error(t, err, "expected error but got none")
            } else {
                require.NoError(t, err, "unexpected error: %v", err)
                assert.Equal(t, tt.expected, result.String(), "event type mismatch")
            }
        })
	}
}

func TestEventType_String(t *testing.T) {
	tests := []struct {
        name     string
        input    EventType
        expected string
    }{
        {
            name: "Earthquake",
            input: EventType{value: EventTypeEarthQuake},
            expected: EventTypeEarthQuake,
        },
        {
            name: "Quarry Blast",
            input: EventType{value: EventTypeQuarryBlast},
            expected: EventTypeQuarryBlast,
        },
        {
            name: "Explosion",
            input: EventType{value: EventTypeExplosion},
            expected: EventTypeExplosion,
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            et, err := NewEventType(tt.input.value)
            require.NoError(t, err)
            assert.Equal(t, tt.expected, et.String())
        })
    }
}

func TestEventType_IsKnown(t *testing.T) {
	tests := []struct {
        name     string
        input    EventType
        expected bool
    }{
        {
            name: "Known Type",
            input: EventType{value: EventTypeEarthQuake},
            expected: true,
        },
        {
            name: "Unknown Type",
            input: EventType{value: EventTypeOther},
            expected: false,
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            et, err := NewEventType(tt.input.value)
            require.NoError(t, err)
            assert.Equal(t, tt.expected, et.IsKnown())
        })
    }
}
