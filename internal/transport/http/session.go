package http

import (
	"CourseJob/internal/service"
	"CourseJob/internal/transport/http/dto"
	"encoding/json"
	nethttp "net/http"
)

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
		return
	}
	if err := ValidatorSession(&req); err != nil {
		writejSON(w, 500, response{
			"status": "error",
			"error":  err.Error(),
		})
		return
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
