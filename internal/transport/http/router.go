package http

import (
	"github.com/go-chi/chi/v5"
	nethttp "net/http"
)

func NewRouter(h *Handler) nethttp.Handler {
	r := chi.NewRouter()

	r.Get("/health/live", h.Leave)
	r.Get("/health/ready", h.Ready)
	r.Post("/api/v1/attendance/sessions", h.CreateAttendanceSession)
	return r
}
