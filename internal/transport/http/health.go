package http

import (
	"context"
	nethttp "net/http"
	"time"
)

// Leave handles liveness checks.
// @Summary Liveness probe
// @Description Returns service liveness status.
// @Tags Health
// @Produce json
// @Success 200 {object} HealthLiveResponse
// @Router /health/live [get]
func (h *Handler) Leave(w nethttp.ResponseWriter, r *nethttp.Request) {
	writeJSON(w, nethttp.StatusOK, jsonResponse{
		"status": "ok",
	})
	return
}

// Ready handles readiness checks.
// @Summary Readiness probe
// @Description Verifies shutdown flag and database availability.
// @Tags Health
// @Produce json
// @Success 200 {object} HealthReadyResponse
// @Failure 503 {object} ErrorResponse
// @Router /health/ready [get]
func (h *Handler) Ready(w nethttp.ResponseWriter, r *nethttp.Request) {
	if h.ready != nil && !h.ready.Load() {
		writeJSON(w, nethttp.StatusServiceUnavailable, jsonResponse{
			"status": "error",
			"error":  "service is shutting down",
		})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		writeJSON(w, nethttp.StatusServiceUnavailable, jsonResponse{
			"status": "error",
			"error":  "database is unavailable",
		})
		return
	}

	writeJSON(w, nethttp.StatusOK, jsonResponse{
		"status": "ok",
		"db":     "connected",
	})
}
