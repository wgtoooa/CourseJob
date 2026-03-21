package http

import (
	"github.com/go-chi/chi/v5"
	nethttp "net/http"
)

func NewRouter(h *Handler) nethttp.Handler {
	r := chi.NewRouter()

	r.Get("/ping", h.Ping)
	r.Get("/db/ping", h.PingDB)
	return r
}
