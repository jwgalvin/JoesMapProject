package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/jwgal/JoesMapProject/internal/domain/event"
)

const insertEventQuery = `
    INSERT INTO events (
        id,
        event_type,
        magnitude_value,
        magnitude_scale,
        latitude,
        longitude,
        depth_km,
        event_time,
        location_name,
        status,
        description,
        url,
        updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
    ON CONFLICT(id) DO UPDATE SET
        event_type = excluded.event_type,
        magnitude_value = excluded.magnitude_value,
        magnitude_scale = excluded.magnitude_scale,
        latitude = excluded.latitude,
        longitude = excluded.longitude,
        depth_km = excluded.depth_km,
        event_time = excluded.event_time,
        location_name = excluded.location_name,
        status = excluded.status,
        description = excluded.description,
        url = excluded.url,
        updated_at = CURRENT_TIMESTAMP
`

const selectEventByIDQuery = `
    SELECT
        id,
        event_type,
        magnitude_value,
        magnitude_scale,
        latitude,
        longitude,
        depth_km,
        event_time,
        location_name,
        status,
        description,
        url,
        updated_at
    FROM events
    WHERE id = ?
`

const selectAllEventsQuery = `
    SELECT
        id,
        event_type,
        magnitude_value,
        magnitude_scale,
        latitude,
        longitude,
        depth_km,
        event_time,
        location_name,
        status,
        description,
        url,
        updated_at
    FROM events
    WHERE 1=1
`

const countEventsQuery = `
    SELECT COUNT(*)
    FROM events
    WHERE 1=1
`

const deleteEventByIDQuery = `
	DELETE FROM events
	WHERE id = ?
`

const orderByTime = "time"

type SQLiteEventRepository struct {
	db *sql.DB
}

func NewSQLiteEventRepository(db *sql.DB) *SQLiteEventRepository {
	return &SQLiteEventRepository{db: db}
}

func (r *SQLiteEventRepository) Save(ctx context.Context, event *event.Event) error {
	_, err := r.db.ExecContext(
		ctx,
		insertEventQuery,
		event.ID(),
		event.Type().String(),
		event.Magnitude().Value(),
		event.Magnitude().Scale(),
		event.Location().LatitudeValue(),
		event.Location().LongitudeValue(),
		event.Location().DepthValue(),
		event.Time().Format(time.RFC3339),
		event.Place(),
		event.Status(),
		event.Description(),
		event.URL(),
	)
	return err
}

func (r *SQLiteEventRepository) FindbyID(ctx context.Context, id string) (*event.Event, error) {
	row := r.db.QueryRowContext(ctx, selectEventByIDQuery, id)
	return reconstructEventFromRow(row)
}

func (r *SQLiteEventRepository) FindAll(ctx context.Context, criteria *event.QueryCriteria) ([]*event.Event, error) {
	if criteria == nil {
		criteria = event.NewQueryCriteria()
	}

	var queryBuilder strings.Builder
	queryBuilder.WriteString(selectAllEventsQuery)
	args := make([]any, 0)
	appendEventFilters(criteria, &queryBuilder, &args)

	orderByColumn := "event_time"
	switch criteria.OrderBy {
	case "magnitude":
		orderByColumn = "magnitude_value"
	case "depth":
		orderByColumn = "depth_km"
	case "place":
		orderByColumn = "location_name"
	case orderByTime:
		orderByColumn = "event_time"
	}

	queryBuilder.WriteString(" ORDER BY ")
	queryBuilder.WriteString(orderByColumn)
	if criteria.Ascending {
		queryBuilder.WriteString(" ASC")
	} else {
		queryBuilder.WriteString(" DESC")
	}

	queryBuilder.WriteString(" LIMIT ? OFFSET ?")
	args = append(args, criteria.Limit, criteria.Offset)

	rows, err := r.db.QueryContext(ctx, queryBuilder.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]*event.Event, 0)
	for rows.Next() {
		eventObj, err := reconstructEventFromScanner(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, eventObj)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *SQLiteEventRepository) Count(ctx context.Context, criteria *event.QueryCriteria) (int64, error) {
	if criteria == nil {
		criteria = event.NewQueryCriteria()
	}

	var queryBuilder strings.Builder
	queryBuilder.WriteString(countEventsQuery)
	args := make([]any, 0)
	appendEventFilters(criteria, &queryBuilder, &args)

	row := r.db.QueryRowContext(ctx, queryBuilder.String(), args...)
	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func (r *SQLiteEventRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, deleteEventByIDQuery, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func appendEventFilters(criteria *event.QueryCriteria, queryBuilder *strings.Builder, args *[]any) {
	if criteria.MinMagnitude != nil {
		queryBuilder.WriteString(" AND magnitude_value >= ?")
		*args = append(*args, *criteria.MinMagnitude)
	}
	if criteria.MaxMagnitude != nil {
		queryBuilder.WriteString(" AND magnitude_value <= ?")
		*args = append(*args, *criteria.MaxMagnitude)
	}
	if criteria.StartTime != nil {
		queryBuilder.WriteString(" AND event_time >= ?")
		*args = append(*args, *criteria.StartTime)
	}
	if criteria.EndTime != nil {
		queryBuilder.WriteString(" AND event_time <= ?")
		*args = append(*args, *criteria.EndTime)
	}
	if len(criteria.EventTypes) > 0 {
		queryBuilder.WriteString(" AND event_type IN (")
		for i, eventType := range criteria.EventTypes {
			if i > 0 {
				queryBuilder.WriteString(", ")
			}
			queryBuilder.WriteString("?")
			*args = append(*args, eventType.String())
		}
		queryBuilder.WriteString(")")
	}
	if len(criteria.Statuses) > 0 {
		queryBuilder.WriteString(" AND description IN (")
		for i, status := range criteria.Statuses {
			if i > 0 {
				queryBuilder.WriteString(", ")
			}
			queryBuilder.WriteString("?")
			*args = append(*args, status)
		}
		queryBuilder.WriteString(")")
	}
	if criteria.Location != nil && criteria.RadiusKm != nil {
		lat := criteria.Location.LatitudeValue()
		lon := criteria.Location.LongitudeValue()
		radiusKm := *criteria.RadiusKm

		deltaLat := radiusKm / 111.0
		cosLat := math.Cos(lat * math.Pi / 180.0)
		deltaLon := 180.0
		if math.Abs(cosLat) > 1e-6 {
			deltaLon = radiusKm / (111.0 * cosLat)
		}

		queryBuilder.WriteString(" AND latitude BETWEEN ? AND ?")
		*args = append(*args, lat-deltaLat, lat+deltaLat)
		queryBuilder.WriteString(" AND longitude BETWEEN ? AND ?")
		*args = append(*args, lon-deltaLon, lon+deltaLon)
	}
}

type rowScanner interface {
	Scan(dest ...any) error
}

// parseTimeFlexible attempts to parse time strings in multiple formats
func parseTimeFlexible(timeStr string) (time.Time, error) {
	// Try RFC3339 first (our standard format)
	if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
		return t, nil
	}

	// Try RFC3339Nano (with nanoseconds)
	if t, err := time.Parse(time.RFC3339Nano, timeStr); err == nil {
		return t, nil
	}

	// Try SQLite datetime format (no timezone)
	if t, err := time.Parse("2006-01-02 15:04:05", timeStr); err == nil {
		return t, nil
	}

	// Try SQLite datetime format with timezone
	if t, err := time.Parse("2006-01-02 15:04:05-07:00", timeStr); err == nil {
		return t, nil
	}

	// Try Go's default time.String() format
	if t, err := time.Parse("2006-01-02 15:04:05 -0700 MST", timeStr); err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s", timeStr)
}

func reconstructEventFromRow(eventRow *sql.Row) (*event.Event, error) {
	return reconstructEventFromScanner(eventRow)
}

func reconstructEventFromScanner(scanner rowScanner) (*event.Event, error) {
	var (
		eventID        string
		eventType      string
		magnitudeValue float64
		magnitudeScale string
		latitude       float64
		longitude      float64
		depthKm        float64
		eventTime      string
		locationName   string
		status         string
		description    string
		url            string
		updatedAt      string
	)

	err := scanner.Scan(
		&eventID,
		&eventType,
		&magnitudeValue,
		&magnitudeScale,
		&latitude,
		&longitude,
		&depthKm,
		&eventTime,
		&locationName,
		&status,
		&description,
		&url,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}

	location, err := event.NewLocation(latitude, longitude, depthKm)
	if err != nil {
		return nil, err
	}

	magnitude, err := event.NewMagnitude(magnitudeValue, magnitudeScale)
	if err != nil {
		return nil, err
	}

	eventTimeParsed, err := parseTimeFlexible(eventTime)
	if err != nil {
		return nil, err
	}

	eventTypeObj, err := event.NewType(eventType)
	if err != nil {
		return nil, err
	}

	eventObj, err := event.NewEvent(
		eventID,
		location,
		locationName,
		magnitude,
		eventTypeObj,
		eventTimeParsed,
		status,
		description,
		url,
	)
	if err != nil {
		return nil, err
	}

	return eventObj, nil
}
