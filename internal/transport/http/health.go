package http

import (
	"context"
	nethttp "net/http"
	"time"
)

func (h *Handler) Leave(w nethttp.ResponseWriter, r *nethttp.Request) {
	writeJSON(w, nethttp.StatusOK, jsonResponse{
		"status": "ok",
	})
	return
}

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
