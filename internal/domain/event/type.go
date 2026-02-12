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

type Type struct {
	value string
}

// Constructor for Type
func NewType(value string) (Type, error) {
	if value == "" {
		return Type{}, fmt.Errorf("value type cannot be empty")
	}

	normalizeValue := strings.ToLower(strings.TrimSpace(value))

	knownValue, exists := validEventTypes[normalizeValue]
	if !exists {
		knownValue = EventTypeOther // Default to "other" if not recognized
	}

	return Type{
		value: knownValue,
	}, nil
}

// Helper methods below
// string representation of Type
func (et Type) String() string {
	return et.value
}

func (et Type) IsKnown() bool {
	return et.value != EventTypeOther
}
