// Package prom collects metrics in the background and provides the metrics to the scraper
package prom

import (
	"log/slog"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// GiteaCollector collects metrics async from Gitea and stores them in memory
type GiteaCollector struct {
	sync.RWMutex
	querier    *querier
	logger     *slog.Logger
	targetName string
	registry   *prometheus.Registry
}

// NewGiteaCollector creates a new GiteaCollector
func NewGiteaCollector(server Server, targetName string, logger *slog.Logger) *GiteaCollector {
	g := GiteaCollector{
		querier:    newQuerier(server, targetName, logger),
		targetName: targetName,
		logger:     logger,
	}
	return &g
}

func (g *GiteaCollector) startBackgroundScrape(interval time.Duration) {
	go func() {
		g.query()
	}()

	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			g.query()
		}
	}()
}

func (g *GiteaCollector) query() {
	g.Lock()
	g.querier.fetchAll()
	g.registry = g.querier.registry
	g.Unlock()
}

// GetRegistry returns the registry with the collected values
func (g *GiteaCollector) GetRegistry() *prometheus.Registry {
	return g.registry
}
