package event

import (
	"fmt"
	"time"
)

type Event struct {
	id        string
	location  Location
	magnitude Magnitude
	eventType Type
	time      time.Time
	place     string
	status    string
	updated   time.Time
}

func NewEvent(id string, location Location, place string, magnitude Magnitude, eventType Type, eventTime time.Time, status string) (*Event, error) {
	if id == "" {
		return nil, fmt.Errorf("event ID cannot be empty")
	}
	if eventTime.IsZero() {
		return nil, fmt.Errorf("event time cannot be zero")
	}
	if status == "" {
		return nil, fmt.Errorf("event status cannot be empty")
	}

	return &Event{
		id:        id,
		location:  location,
		place:     place,
		magnitude: magnitude,
		eventType: eventType,
		time:      eventTime,
		updated:   time.Now(),
		status:    status,
	}, nil
}

// normally want to avoid getters/setters in Go, but for this domain model we want to enforce immutability and encapsulation, so we provide read-only accessors
// also, the event was a Heavy struct, (160 bytes) so we want to avoid copying it around, hence the pointer receiver for methods that return values
func (e *Event) String() string {
	return fmt.Sprintf("Event ID: %s, Type: %s, Magnitude: %s, Location: %s, Place: %s, Time: %s",
		e.id, e.eventType.String(), e.magnitude.String(), e.location.String(), e.Place(), e.time.Format(time.RFC3339))
}

func (e *Event) IsSignificant(threshold float64) bool {
	return e.magnitude.Value() >= threshold
}

func (e *Event) Status() string {
	return e.status
}

func (e *Event) Type() Type {
	return e.eventType
}

func (e *Event) Magnitude() Magnitude {
	return e.magnitude
}

func (e *Event) Location() Location {
	return e.location
}

func (e *Event) Time() time.Time {
	return e.time
}

func (e *Event) ID() string {
	return e.id
}

func (e *Event) Updated() time.Time {
	return e.updated
}

func (e *Event) Place() string {
	return e.place
}

func (e *Event) UpdateStatus(newStatus string, updatedTime time.Time) *Event {
	return &Event{
		id:        e.id,
		location:  e.location,
		place:     e.place,
		magnitude: e.magnitude,
		eventType: e.eventType,
		time:      e.time,
		status:    newStatus,
		updated:   updatedTime,
	}
}
