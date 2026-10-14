package application

import (
	"context"
	"gitea-exporter/internal/domain"
	"log/slog"
	"time"
)

// Target bundles everything needed to collect from one Gitea server and
// hold its latest snapshot. The runner writes to it; readers (the metrics
// collector) read Name and Snapshot.
type Target struct {
	name    string
	source  domain.GiteaRepository
	service *Service
	store   *SnapshotStore
}

// NewTarget builds a Target for one server. source is the HTTP client;
// service applies this server's exclusion policy.
func NewTarget(name string, source domain.GiteaRepository, service *Service) *Target {
	return &Target{
		name:    name,
		source:  source,
		service: service,
		store:   NewSnapshotStore(),
	}
}

// Name returns the server name, used as the Prometheus target label.
func (target *Target) Name() string { return target.name }

// Snapshot returns the most recently collected snapshot, or nil if none yet.
func (target *Target) Snapshot() *domain.Snapshot { return target.store.Get() }

// Runner collects a Snapshot from every target on a fixed interval.
type Runner struct {
	targets  []*Target
	interval time.Duration
}

// NewRunner builds a Runner that collects every interval.
func NewRunner(interval time.Duration, targets ...*Target) *Runner {
	return &Runner{
		targets:  targets,
		interval: interval,
	}
}

// Run collects immediately, then every interval, until ctx is cancelled.
// It blocks, so call it as the last statement of main.
func (runner *Runner) Run(ctx context.Context) {
	runner.Collect(ctx)

	ticker := time.NewTicker(runner.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			runner.Collect(ctx)
		}
	}
}

// Collect runs one collection pass over every target. A target whose
// collection fails is logged and skipped; its store keeps the previous
// snapshot.
func (runner *Runner) Collect(ctx context.Context) {
	for _, target := range runner.targets {
		snapshot, err := target.service.Collect(ctx, target.source)
		if err != nil {
			slog.Error("collect failed", "target", target.name, "error", err)
			continue
		}
		target.store.Set(snapshot)
		slog.Info("collect finished", "target", target.name, "organizations", len(snapshot.Orgs))
	}
}
