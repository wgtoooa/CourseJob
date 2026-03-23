package http

import (
	"CourseJob/internal/service"
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5/pgxpool"
	nethttp "net/http"
	"time"
)

type Handler struct {
	DB                *pgxpool.Pool
	attendanceService *service.AttendanceService
}

func NewHandler(db *pgxpool.Pool, attendanceService *service.AttendanceService) *Handler {
	return &Handler{db, attendanceService}
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
	return
}

func (h *Handler) PingDB(w nethttp.ResponseWriter, r *nethttp.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	if err := h.DB.Ping(ctx); err != nil {
		writejSON(w, 500, response{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}
	writejSON(w, nethttp.StatusOK, response{
		"status": "ok",
		"db":     "connected",
	})

}
