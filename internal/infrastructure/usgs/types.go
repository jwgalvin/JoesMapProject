package usgs

// GeoJSONResponse represents the top-level USGS earthquake GeoJSON feed response.
type GeoJSONResponse struct {
	Type     string    `json:"type"`
	Metadata Metadata  `json:"metadata"`
	Features []Feature `json:"features"`
}

// Metadata contains feed-level metadata returned by the USGS API.
type Metadata struct {
	Generated int64  `json:"generated"`
	URL       string `json:"url"`
	Title     string `json:"title"`
	Status    int    `json:"status"`
	API       string `json:"api"`
	Count     int    `json:"count"`
}

// Feature represents a single earthquake event entry in the GeoJSON feed.
type Feature struct {
	Type       string     `json:"type"`
	Properties Properties `json:"properties"`
	Geometry   Geometry   `json:"geometry"`
	ID         string     `json:"id"`
}

// Properties contains the event details for a USGS Feature.
type Properties struct {
	Mag     *float64 `json:"mag"`
	Place   string   `json:"place"`
	Time    int64    `json:"time"`    // Unix timestamp in milliseconds
	Updated int64    `json:"updated"` // Unix timestamp in milliseconds
	URL     string   `json:"url"`
	Alert   *string  `json:"alert"`
	Status  string   `json:"status"`
	MagType string   `json:"magType"`
	Type    string   `json:"type"`
	Title   string   `json:"title"`
}

// Geometry contains the spatial coordinates for a USGS Feature.
// Coordinates are ordered [longitude, latitude, depth_km]; depth may be absent.
type Geometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}
