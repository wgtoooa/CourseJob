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
	db                *pgxpool.Pool
	attendanceService *service.AttendanceService
}

func NewHandler(db *pgxpool.Pool, attendanceService *service.AttendanceService) *Handler {
	return &Handler{db, attendanceService}
}

type jsonResponse map[string]any

func writeJSON(w nethttp.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func (h *Handler) Ping(w nethttp.ResponseWriter, r *nethttp.Request) {
	writeJSON(w, nethttp.StatusOK, jsonResponse{
		"status": "ok",
	})
	return
}

func (h *Handler) PingDB(w nethttp.ResponseWriter, r *nethttp.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	if err := h.db.Ping(ctx); err != nil {
		writeJSON(w, 500, jsonResponse{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}
	writeJSON(w, nethttp.StatusOK, jsonResponse{
		"status": "ok",
		"db":     "connected",
	})

}
