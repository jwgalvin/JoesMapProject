package event

import (
	"fmt"
	"strings"
)

// value based on geopulse v1 api documentation
const (
	EventTypeEarthQuake       = "earthquake"
	EventTypeQuarryBlast      = "quarry blast"
	EventTypeExplosion        = "explosion"
	EventTypeOther            = "other"
	EventTypeLandslide        = "landslide"
	EventTypeVolcanicEruption = "volcanic eruption"
	EventTypeIceQuake         = "ice quake"
)

// Mapvalid value types from USGS to Geopulse types
var validEventTypes = map[string]string{
	"earthquake":         EventTypeEarthQuake,
	"quarry blast":       EventTypeQuarryBlast,
	"explosion":          EventTypeExplosion,
	"other":              EventTypeOther,
	"quarry":             EventTypeQuarryBlast, // USGS occasionally abbreviates "quarry blast" as "quarry"
	"landslide":          EventTypeLandslide,
	"rockburst":          EventTypeOther,
	"chemical explosion": EventTypeExplosion,
	"nuclear explosion":  EventTypeExplosion,
	"ice quake":          EventTypeIceQuake,
	"volcanic eruption":  EventTypeVolcanicEruption,
}

type EventType struct {
	value string
}

// Constructor for EventType
func NewEventType(value string) (EventType, error) {
	if value == "" {
		return EventType{}, fmt.Errorf("value type cannot be empty")
	}

	normalizeValue := strings.ToLower(strings.TrimSpace(value))

	knownValue, exists := validEventTypes[normalizeValue]
	if !exists {
		knownValue = EventTypeOther // Default to "other" if not recognized
	}

	return EventType{
		value: knownValue,
	}, nil
}

// Helper methods below
// string representation of EventType
func (et EventType) String() string {
	return et.value
}

func (et EventType) IsKnown() bool {
	return et.value != EventTypeOther
}
