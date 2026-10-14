package domain

import "context"

// MetricsPublisher hands a completed Snapshot to the outside world.
// Implementations decide what "publish" means: expose to Prometheus,
// push to OTLP, or log it. The caller doesn't care which.
type MetricsPublisher interface {
	Publish(ctx context.Context, s Snapshot) error
}
