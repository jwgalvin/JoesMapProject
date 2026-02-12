package event

import (
	"context"
	"fmt"
	"time"
)

type Repository interface {
	Save(ctx context.Context, event *Event) error
	FindByID(ctx context.Context, id string) (*Event, error)
	FindAll(ctx context.Context, QueryCriteria *QueryCriteria) ([]*Event, error)
	Count(ctx context.Context, QueryCriteria *QueryCriteria) (int64, error)
	Delete(ctx context.Context, id string) error
}

type QueryCriteria struct {
	MinMagnitude *float64
	MaxMagnitude *float64
	StartTime    *time.Time
	EndTime      *time.Time
	Location     *Location
	RadiusKm     *float64
	EventTypes   []Type
	Statuses     []string
	OrderBy      string
	Limit        int
	Offset       int
	Ascending    bool
}

func NewQueryCriteria() *QueryCriteria {
	return &QueryCriteria{
		Limit:     100,
		Offset:    0,
		OrderBy:   "time",
		Ascending: false,
	}
}

func (c *QueryCriteria) WithMagnitudeRange(minMag, maxMag float64) error {
	if minMag < -1.0 || maxMag > 10.0 {
		return fmt.Errorf("invalid magnitude range: minMag must be >= -1.0 and maxMag must be <= 10.0")
	}
	if maxMag < -1.0 || maxMag > 10.0 {
		return fmt.Errorf("invalid magnitude range: maxMag must be >= -1.0 and <= 10.0")
	}
	if maxMag < minMag {
		return fmt.Errorf("invalid magnitude range: maxMag cannot be less than minMag")
	}
	c.MinMagnitude = &minMag
	c.MaxMagnitude = &maxMag
	return nil
}

func (c *QueryCriteria) WithTimeRange(start, end time.Time) error {
	if end.Before(start) {
		return fmt.Errorf("invalid time range: end time cannot be before start time")
	}

	c.StartTime = &start
	c.EndTime = &end
	return nil
}

func (c *QueryCriteria) WithProximity(location Location, radiusKm float64) error {
	if radiusKm < 0 {
		return fmt.Errorf("radiusKm must be non-negative, got %f", radiusKm)
	}
	if radiusKm > 20000 {
		return fmt.Errorf("radiusKm must not exceed 20000 km, got %f", radiusKm)
	}
	c.Location = &location
	c.RadiusKm = &radiusKm
	return nil
}

func (c *QueryCriteria) WithEventTypes(types ...Type) error {
	if len(types) == 0 {
		return fmt.Errorf("at least one event type must be specified")
	}
	c.EventTypes = types
	return nil
}

func (c *QueryCriteria) WithStatuses(statuses ...string) error {
	if len(statuses) == 0 {
		return fmt.Errorf("at least one status must be specified")
	}
	c.Statuses = statuses
	return nil
}

func (c *QueryCriteria) WithPagination(limit, offset int) error {
	if limit < 0 {
		return fmt.Errorf("limit must be non-negative, got %d", limit)
	}
	if limit > 1000 {
		return fmt.Errorf("limit must not exceed 1000, got %d", limit)
	}
	if offset < 0 {
		return fmt.Errorf("offset must be non-negative, got %d", offset)
	}
	c.Limit = limit
	c.Offset = offset
	return nil
}

func (c *QueryCriteria) WithSort(orderBy string, ascending bool) error {
	validFields := map[string]bool{
		"time":      true,
		"magnitude": true,
		"depth":     true,
		"place":     true,
	}
	if !validFields[orderBy] {
		return fmt.Errorf("invalid orderBy field: %q", orderBy)
	}
	c.OrderBy = orderBy
	c.Ascending = ascending
	return nil
}
