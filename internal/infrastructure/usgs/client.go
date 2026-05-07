package usgs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/jwgal/geopulse/internal/domain/event"
)

const userAgent = "GeoPulse/1.0"

// Client fetches earthquake event data from the USGS GeoJSON feed.
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new USGS Client with the given base URL and HTTP timeout.
func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: timeout},
		baseURL:    baseURL,
	}
}

// FetchEvents retrieves earthquake events from the USGS feed and converts
// them into domain Event objects. Features that fail conversion are skipped.
func (c *Client) FetchEvents(ctx context.Context) ([]*event.Event, error) {
	slog.Info("usgs: fetching events", "url", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch USGS data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var feed GeoJSONResponse
	if err := json.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return nil, fmt.Errorf("failed to decode USGS response: %w", err)
	}

	events := make([]*event.Event, 0, len(feed.Features))
	skipped := 0
	for i := range feed.Features {
		evt, err := convertFeature(&feed.Features[i])
		if err != nil {
			slog.Warn("usgs: skipping feature", "id", feed.Features[i].ID, "err", err)
			skipped++
			continue
		}
		events = append(events, evt)
	}
	slog.Info("usgs: fetch complete", "total", len(feed.Features), "converted", len(events), "skipped", skipped)
	return events, nil
}

// convertFeature maps a single USGS Feature to a domain Event.
func convertFeature(f *Feature) (*event.Event, error) {
	if len(f.Geometry.Coordinates) < 2 {
		return nil, fmt.Errorf("feature %s has insufficient coordinates", f.ID)
	}

	lon := f.Geometry.Coordinates[0]
	lat := f.Geometry.Coordinates[1]
	depth := 0.0
	if len(f.Geometry.Coordinates) >= 3 {
		depth = f.Geometry.Coordinates[2]
	}

	loc, err := event.NewLocation(lat, lon, depth)
	if err != nil {
		return nil, fmt.Errorf("invalid location for feature %s: %w", f.ID, err)
	}

	magVal := 0.0
	if f.Properties.Mag != nil {
		magVal = *f.Properties.Mag
	}
	mag, err := event.NewMagnitude(magVal, f.Properties.MagType)
	if err != nil {
		return nil, fmt.Errorf("invalid magnitude for feature %s: %w", f.ID, err)
	}

	evtType, err := event.NewType(f.Properties.Type)
	if err != nil {
		return nil, fmt.Errorf("invalid event type for feature %s: %w", f.ID, err)
	}

	status := f.Properties.Status
	if status == "" {
		status = "automatic"
	}

	return event.NewEvent(
		f.ID,
		loc,
		f.Properties.Place,
		mag,
		evtType,
		time.UnixMilli(f.Properties.Time),
		time.UnixMilli(f.Properties.Updated),
		status,
		f.Properties.Title,
		f.Properties.URL,
	)
}
