package http

import (
	"CourseJob/internal/service"
	"encoding/json"
	"github.com/jackc/pgx/v5/pgxpool"
	nethttp "net/http"
	"sync/atomic"
)

type Handler struct {
	db                *pgxpool.Pool
	attendanceService *service.AttendanceService
	ready             *atomic.Bool
}

func NewHandler(db *pgxpool.Pool, attendanceService *service.AttendanceService, ready *atomic.Bool) *Handler {
	return &Handler{
		db:                db,
		attendanceService: attendanceService,
		ready:             ready,
	}
}

type jsonResponse map[string]any

func writeJSON(w nethttp.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
