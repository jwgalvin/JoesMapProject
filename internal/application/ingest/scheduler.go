package ingest

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type Scheduler struct {
	service  *Service
	interval time.Duration
	logger   *slog.Logger
	stopCh   chan struct{}
	stopOnce sync.Once
}

func NewScheduler(service *Service, interval time.Duration, logger *slog.Logger) *Scheduler {
	return &Scheduler{
		service:  service,
		interval: interval,
		logger:   logger,
		stopCh:   make(chan struct{}),
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	s.logger.Info("scheduler started", "interval", s.interval)

	if err := s.runIngestion(ctx); err != nil {
		s.logger.Error("initial ingestion failed", "error", err)
	}

	for {
		select {
		case <-ticker.C:
			if err := s.runIngestion(ctx); err != nil {
				s.logger.Error("scheduled ingestion failed", "error", err)
			}
		case <-s.stopCh:
			s.logger.Info("scheduler stopped")
			return
		case <-ctx.Done():
			s.logger.Info("scheduler context cancelled")
			return
		}
	}
}

func (s *Scheduler) Stop() {
	s.stopOnce.Do(func() { close(s.stopCh) })
}

func (s *Scheduler) runIngestion(ctx context.Context) error {
	start := time.Now()
	result, err := s.service.IngestEvents(ctx)
	duration := time.Since(start)

	if err != nil {
		s.logger.Error("ingestion failed", "duration", duration, "error", err)
		return err
	}

	s.logger.Info("ingestion succeeded",
		"duration", duration,
		"saved", result.Saved,
		"failed", result.Failed,
	)
	return nil
}
