package http

import (
	"CourseJob/internal/service/attendance"
	authService "CourseJob/internal/service/auth"
	"CourseJob/internal/service/schedule"
	"encoding/json"
	"github.com/jackc/pgx/v5/pgxpool"
	nethttp "net/http"
	"sync/atomic"
)

type Handler struct {
	db                *pgxpool.Pool
	authService       *authService.Service
	attendanceService *attendance.AttendanceService
	scheduleService   *schedule.Service
	ready             *atomic.Bool
	scheduleEvents    *scheduleSSEBroker
}

func NewHandler(
	db *pgxpool.Pool,
	authService *authService.Service,
	attendanceService *attendance.AttendanceService,
	scheduleService *schedule.Service,
	ready *atomic.Bool,
) *Handler {
	return &Handler{
		db:                db,
		authService:       authService,
		attendanceService: attendanceService,
		scheduleService:   scheduleService,
		ready:             ready,
		scheduleEvents:    newScheduleSSEBroker(),
	}
}

type jsonResponse map[string]any

func writeJSON(w nethttp.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
