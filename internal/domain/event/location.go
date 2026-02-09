package event

import "fmt"

type Location struct {
    Latitude  float64
    Longitude float64
    Depth     float64
}

func NewLocation(latitude, longitude, depth float64) (Location, error) {
    if latitude < -90.0 || latitude > 90.0 {
        return Location{}, fmt.Errorf("latitude must be between -90 and 90 degrees")
    }

    if longitude < -180.0 || longitude > 180.0 {
        return Location{}, fmt.Errorf("longitude must be between -180 and 180 degrees")
    }

    return Location{
        Latitude:  latitude,
        Longitude: longitude,
        Depth:     depth,
    }, nil
}

func (loc Location) String() string {
    return fmt.Sprintf("Lat: %.4f, Lon: %.4f, Depth: %.2f km", loc.Latitude, loc.Longitude, loc.Depth)
}

func (loc Location) LatitudeValue() float64 {
    return loc.Latitude
}

func (loc Location) LongitudeValue() float64 {
    return loc.Longitude
}

func (loc Location) DepthValue() float64 {
    return loc.Depth
}

func (loc Location) IsShallow() bool {
    return loc.Depth < 70.0
}

func (loc Location) IsDeep() bool {
    return loc.Depth >= 300.0
}
