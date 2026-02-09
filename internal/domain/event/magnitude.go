package event

import (
	"fmt"
	"strings"
)

const (
	MagnitudeScaleMw      = "mw"  // Moment magnitude (most common)
	MagnitudeScaleMl      = "ml"  // Local magnitude (Richter)
	MagnitudeScaleMb      = "mb"  // Body wave magnitude
	MagnitudeScaleMs      = "ms"  // Surface wave magnitude
	MagnitudeScaleMd      = "md"  // Duration magnitude
	MagnitudeScaleMww     = "mww" // W-phase moment magnitude
	MagnitudeScaleMwc     = "mwc" // Centroid moment magnitude
	MagnitudeScaleMwr     = "mwr" // Regional moment magnitude
	MagnitudeScaleUnknown = "unknown"
)

var validMagnitudeScales = map[string]string{
	MagnitudeScaleMw:  MagnitudeScaleMw,
	MagnitudeScaleMl:  MagnitudeScaleMl,
	MagnitudeScaleMb:  MagnitudeScaleMb,
	MagnitudeScaleMs:  MagnitudeScaleMs,
	MagnitudeScaleMd:  MagnitudeScaleMd,
	MagnitudeScaleMww: MagnitudeScaleMww,
	MagnitudeScaleMwc: MagnitudeScaleMwc,
	MagnitudeScaleMwr: MagnitudeScaleMwr,
}

type Magnitude struct {
	value float64
	scale string
}

// Constructor for magnitude
func NewMagnitude(value float64, scale string) (Magnitude, error) {
	// Validate magnitude range -1 to 10 (typical range for earthquakes)
	if value < -1.0 || value > 10.0 {
		return Magnitude{}, fmt.Errorf("magnitude must be between -1.0 and 10.0, got %f", value)
	}

	// normalize scale input
	normalizeScale := strings.ToLower(strings.TrimSpace(scale))

	// Validate scale
	knownScale, exists := validMagnitudeScales[normalizeScale]
	if !exists {
		knownScale = MagnitudeScaleUnknown // Default to "unknown" if not recognized
	}

	return Magnitude{value: value, scale: knownScale}, nil
}

func (m Magnitude) String() string {
	return fmt.Sprintf("%.1f %s", m.value, m.scale)
}

func (m Magnitude) Value() float64 {
	return m.value
}

func (m Magnitude) Scale() string {
	return m.scale
}

func (m Magnitude) IsKnown() bool {
	return m.scale != MagnitudeScaleUnknown
}
