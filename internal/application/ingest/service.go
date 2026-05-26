package ingest

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jwgal/geopulse/internal/domain/event"
)

type Service struct {
	client     EventFetcher
	repository event.Repository
	logger     *slog.Logger
}

type Result struct {
	Fetched int
	Saved   int
	Failed  int
}

func NewService(client EventFetcher, repo event.Repository, logger *slog.Logger) *Service {
	return &Service{
		client:     client,
		repository: repo,
		logger:     logger,
	}
}

func (s *Service) IngestEvents(ctx context.Context) (*Result, error) {
	s.logger.Info("starting event ingestion")

	events, err := s.client.FetchEvents(ctx)
	if err != nil {
		s.logger.Error("failed to fetch events", "error", err)
		return nil, fmt.Errorf("fetch failed: %w", err)
	}

	result := &Result{Fetched: len(events)}

	for _, evt := range events {
		if err := s.repository.Save(ctx, evt); err != nil {
			s.logger.Error("faild to save event", "id", evt.ID(), "error", err)
		}
		result.Saved++
	}

	s.logger.Info("ingestion complete",
		"fetched", result.Fetched,
		"saved", result.Saved,
		"failed", result.Failed,
	)

	return result, nil
}
