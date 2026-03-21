package http

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	nethttp "net/http"
	"time"
)

type Handler struct {
	DB *pgxpool.Pool
}

type response map[string]any

func writejSON(w nethttp.ResponseWriter, status int, data response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func (h *Handler) Ping(w nethttp.ResponseWriter, r *nethttp.Request) {
	writejSON(w, nethttp.StatusOK, response{
		"status": "ok",
	})
}

func (h *Handler) PingDB(w nethttp.ResponseWriter, r *nethttp.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	log.Println("1")
	if err := h.DB.Ping(ctx); err != nil {
		log.Println("2")
		writejSON(w, 500, response{
			"status": "error",
			"error":  err.Error(),
		})
		log.Println("3")
		writejSON(w, nethttp.StatusOK, response{
			"status": "ok",
			"db":     "connected",
		})
	}
}
