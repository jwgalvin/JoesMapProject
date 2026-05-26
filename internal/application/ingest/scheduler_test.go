package ingest

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jwgal/geopulse/internal/domain/event"
	"github.com/stretchr/testify/assert"
)

// ─── Test double ──────────────────────────────────────────────────────────────

// countingFetcher records every call to FetchEvents and closes notify once
// the call count reaches target.  It is safe to use from a single goroutine
// (the scheduler never calls FetchEvents concurrently).
type countingFetcher struct {
	calls  atomic.Int32
	notify chan struct{} // closed exactly once when calls == target
	target int32
	err    error // returned on every call when non-nil
}

func newCountingFetcher(target int, err error) *countingFetcher {
	return &countingFetcher{
		notify: make(chan struct{}),
		target: int32(target),
		err:    err,
	}
}

func (f *countingFetcher) FetchEvents(_ context.Context) ([]*event.Event, error) {
	if n := f.calls.Add(1); n == f.target {
		close(f.notify)
	}
	return nil, f.err
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

// startScheduler launches sched.Start in a goroutine and returns a done channel
// that is closed when Start returns.
func startScheduler(ctx context.Context, sched *Scheduler) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		sched.Start(ctx)
	}()
	return done
}

// awaitOrTimeout fails the test if ch is not closed within deadline.
func awaitOrTimeout(t *testing.T, ch <-chan struct{}, deadline time.Duration, msg string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(deadline):
		t.Fatal(msg)
	}
}

// ─── Tests ────────────────────────────────────────────────────────────────────

// TestScheduler_New verifies the constructor returns a non-nil Scheduler.
func TestScheduler_New(t *testing.T) {
	svc := NewService(&mockFetcher{}, &mockRepository{}, discardLogger())
	sched := NewScheduler(svc, time.Hour, discardLogger())
	if sched == nil {
		t.Fatal("NewScheduler returned nil")
	}
}

// TestScheduler_RunsImmediatelyOnStart confirms that runIngestion is called
// before the first tick fires.  A 1-hour interval ensures the ticker never
// fires during the test; only the pre-loop call can reach the target of 1.
func TestScheduler_RunsImmediatelyOnStart(t *testing.T) {
	fetcher := newCountingFetcher(1, nil)
	svc := NewService(fetcher, &mockRepository{}, discardLogger())
	sched := NewScheduler(svc, time.Hour, discardLogger())

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := startScheduler(ctx, sched)

	awaitOrTimeout(t, fetcher.notify, 2*time.Second,
		"timed out: initial ingestion did not run before the first tick")

	sched.Stop()
	awaitOrTimeout(t, done, time.Second, "scheduler goroutine did not exit after Stop()")

	assert.EqualValues(t, 1, fetcher.calls.Load(),
		"only the pre-loop call should have fired (interval was 1 hour)")
}

// TestScheduler_TicksCallIngestion verifies that the ticker branch also fires,
// i.e. ingestion is called again after the initial run.
func TestScheduler_TicksCallIngestion(t *testing.T) {
	// Target 2: initial run + at least one tick.
	fetcher := newCountingFetcher(2, nil)
	svc := NewService(fetcher, &mockRepository{}, discardLogger())
	sched := NewScheduler(svc, time.Millisecond, discardLogger())

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := startScheduler(ctx, sched)

	awaitOrTimeout(t, fetcher.notify, 2*time.Second,
		"timed out: expected initial run + at least one tick")

	sched.Stop()
	awaitOrTimeout(t, done, time.Second, "scheduler goroutine did not exit after Stop()")

	assert.GreaterOrEqual(t, int(fetcher.calls.Load()), 2,
		"expected ≥2 ingestion calls (initial + ≥1 tick)")
}

// TestScheduler_StopsOnStop verifies that calling Stop() causes the Start
// goroutine to return cleanly without a context cancellation.
func TestScheduler_StopsOnStop(t *testing.T) {
	fetcher := newCountingFetcher(1, nil)
	svc := NewService(fetcher, &mockRepository{}, discardLogger())
	sched := NewScheduler(svc, time.Millisecond, discardLogger())

	done := startScheduler(context.Background(), sched)

	// Wait for the scheduler to be running before stopping it.
	awaitOrTimeout(t, fetcher.notify, 2*time.Second,
		"scheduler did not run initial ingestion")

	sched.Stop()

	awaitOrTimeout(t, done, time.Second,
		"scheduler goroutine did not exit after Stop()")
}

// TestScheduler_StopsOnContextCancel verifies that cancelling the context
// causes the Start goroutine to return cleanly.
func TestScheduler_StopsOnContextCancel(t *testing.T) {
	fetcher := newCountingFetcher(1, nil)
	svc := NewService(fetcher, &mockRepository{}, discardLogger())
	sched := NewScheduler(svc, time.Millisecond, discardLogger())

	ctx, cancel := context.WithCancel(context.Background())

	done := startScheduler(ctx, sched)

	awaitOrTimeout(t, fetcher.notify, 2*time.Second,
		"scheduler did not run initial ingestion")

	cancel()

	awaitOrTimeout(t, done, time.Second,
		"scheduler goroutine did not exit after context cancellation")
}

// TestScheduler_ContinuesAfterIngestionError verifies that an ingestion error
// (e.g. a failed fetch) is logged and swallowed: the scheduler keeps running
// rather than stopping at the first failure.
func TestScheduler_ContinuesAfterIngestionError(t *testing.T) {
	fetchErr := errors.New("upstream unavailable")
	// Target 3: initial run + 2 ticks, each returning an error.
	fetcher := newCountingFetcher(3, fetchErr)
	svc := NewService(fetcher, &mockRepository{}, discardLogger())
	sched := NewScheduler(svc, time.Millisecond, discardLogger())

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := startScheduler(ctx, sched)

	select {
	case <-fetcher.notify:
		// 3 error-returning calls completed → scheduler kept running after errors.
	case <-ctx.Done():
		t.Fatalf("timed out: only %d call(s) completed; scheduler may have stopped after the first error",
			fetcher.calls.Load())
	}

	sched.Stop()
	awaitOrTimeout(t, done, time.Second, "scheduler goroutine did not exit after Stop()")

	assert.GreaterOrEqual(t, int(fetcher.calls.Load()), 3,
		"expected ≥3 calls despite repeated errors")
}

// TestScheduler_StopIsIdempotent verifies that calling Stop() twice does not
// panic (i.e. does not close an already-closed channel).
func TestScheduler_StopIsIdempotent(t *testing.T) {
	fetcher := newCountingFetcher(1, nil)
	svc := NewService(fetcher, &mockRepository{}, discardLogger())
	sched := NewScheduler(svc, time.Millisecond, discardLogger())

	done := startScheduler(context.Background(), sched)

	awaitOrTimeout(t, fetcher.notify, 2*time.Second,
		"scheduler did not run initial ingestion")

	sched.Stop()
	awaitOrTimeout(t, done, time.Second, "scheduler did not exit after first Stop()")

	// Second Stop() must not panic.
	assert.NotPanics(t, sched.Stop)
}
