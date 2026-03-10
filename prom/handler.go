package prom

import (
	"fmt"
	"gitea-exporter/gitea"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Handler handle prometheus queries
type Handler struct {
	Targets    Targets `yaml:"targets"`
	logger     *slog.Logger
	collectors map[string]*GiteaCollector
	client     *gitea.Client
}

type collectorDescription struct {
	name string
	help string
}

// NewHandler creeates a new prometheus handler
func NewHandler(configFile string, logger *slog.Logger) *Handler {
	t := readConfig(configFile)
	h := Handler{
		Targets: t,
		logger:  logger,
	}
	h.init()
	return &h
}

func (h *Handler) init() {
	if len(h.Targets) > 0 {
		h.collectors = make(map[string]*GiteaCollector)
	}
	for k, v := range h.Targets {
		collector := NewGiteaCollector(v, k, h.logger)
		h.collectors[k] = collector
		collector.startBackgroundScrape(10 * time.Minute)
	}
}

// ProbeHandler handles the incoming prometheus queries
func (h *Handler) ProbeHandler(w http.ResponseWriter, r *http.Request) {
	targetParam := r.URL.Query().Get("target")
	if targetParam == "" {
		http.Error(w, "Target parameter missing", http.StatusBadRequest)
		return
	}
	_, ok := h.Targets[targetParam]
	if !ok {
		errorMsg := fmt.Sprintf("Invalid target: %s", targetParam)
		http.Error(w, errorMsg, http.StatusBadRequest)
		h.logger.Error(errorMsg)
		return
	}

	registry := h.collectors[targetParam].GetRegistry()

	httpHandler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	httpHandler.ServeHTTP(w, r)
}
