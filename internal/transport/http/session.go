package http

import (
	"CourseJob/internal/service"
	"CourseJob/internal/transport/http/dto"
	"encoding/json"
	nethttp "net/http"
)

func (h *Handler) CreateAttendanceSession(w nethttp.ResponseWriter, r *nethttp.Request) {
	if r.Method != "POST" {
		writeJSON(w, nethttp.StatusMethodNotAllowed, jsonResponse{
			"status": "error",
			"error":  "Method not allowed",
		})
		return
	}

	r.Body = nethttp.MaxBytesReader(w, r.Body, 1<<20) //~ 1 MB
	defer r.Body.Close()

	var req dto.AttendanceSessionRequest
	dec := json.NewDecoder(r.Body)

	if err := dec.Decode(&req); err != nil {
		writeJSON(w, nethttp.StatusBadRequest, jsonResponse{
			"status": "error",
			"error":  "invalid request body",
		})
		return
	}

	NormalizeSessionRequest(&req)

	if err := ValidatorSession(&req); err != nil {
		writeJSON(w, nethttp.StatusBadRequest, jsonResponse{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}
	input := service.AttendanceInput{
		Room:       req.Room,
		Source:     req.Source,
		StartedAt:  req.StartedAt,
		FinishedAt: req.FinishedAt,
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
		writeJSON(w, nethttp.StatusInternalServerError, jsonResponse{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}
	resp := jsonResponse{
		"status": "created",
		"data": dto.AttendanceResponse{
			SessionID:     result.SessionID,
			SavedEvents:   result.SavedEvents,
			NotFoundCards: result.NotFoundCards,
		},
	}

	writeJSON(w, nethttp.StatusCreated, resp)

}
