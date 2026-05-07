package ingest

import (
	"context"

	"github.com/jwgal/JoesMapProject/internal/domain/event"
)

// EventFetcher defines the interface for fetching events from an external source.
type EventFetcher interface {
	FetchEvents(ctx context.Context) ([]*event.Event, error)
}
