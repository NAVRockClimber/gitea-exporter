package prom

import "net/http"

// HealthCheckHandler handles the healthchecks
func (h *Handler) HealthCheckHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
