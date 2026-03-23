package http

import (
	"CourseJob/internal/service"
	"CourseJob/internal/transport/http/dto"
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

func (h *Handler) CreateAttendanceSession(w nethttp.ResponseWriter, r *nethttp.Request) {
	if r.Method != "POST" {
		writejSON(w, 500, response{
			"status": "error",
			"error":  "Method not allowed",
		})
		return
	}
	var req dto.AttendanceSessionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writejSON(w, 500, response{
			"status": "error",
			"error":  "invalid request body",
		})
	}
	input := service.ProcessAttendanceInput{
		Room:       req.Room,
		Source:     req.Source,
		StartedAt:  req.StartedAt,
		FinishedAt: &req.FinishedAt,
		Scans:      make([]service.ProcessAttendanceScanInput, 0, len(req.Scans)),
	}

	for _, scan := range req.Scans {
		input.Scans = append(input.Scans, service.ProcessAttendanceScanInput{
			CardUID:   scan.CardUID,
			ScannedAt: scan.ScannedAt,
		})
	}

	result, err := h.attendanceService.ProcessAttendance(r.Context(), input)
	if err != nil {
		writejSON(w, 500, response{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}
	resp := dto.AttendanceResponse{
		SessionID:    result.SessionID,
		SavedEvents:  result.SavedEvents,
		NotFoundCard: result.NotFoundCards,
	}
	writejSON(w, nethttp.StatusCreated, response{
		"status": "created",
	})

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		writejSON(w, 500, response{
			"status": "error",
			"error":  "failed to encode response",
		})
		return
	}

}
