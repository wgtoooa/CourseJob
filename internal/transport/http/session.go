package http

import (
	"CourseJob/internal/service/attendance"
	"CourseJob/internal/transport/http/dto"
	"CourseJob/internal/transport/http/validator"
	"encoding/json"
	nethttp "net/http"
)

// CreateAttendanceSession saves attendance scans inside one session.
// @Summary Create attendance session
// @Description Accepts a session with scans and stores attendance events.
// @Tags Attendance
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer access token"
// @Param request body dto.AttendanceSessionRequest true "Attendance session payload"
// @Success 201 {object} AttendanceCreateResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/attendance/sessions [post]
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

	validator.NormalizeSessionRequest(&req)

	if err := validator.ValidatorSession(&req); err != nil {
		writeJSON(w, nethttp.StatusBadRequest, jsonResponse{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}
	input := attendance.AttendanceInput{
		Room:       req.Room,
		Source:     req.Source,
		StartedAt:  req.StartedAt,
		FinishedAt: req.FinishedAt,
		Data:       req.Data,
		Scans:      make([]attendance.ProcessAttendanceScanInput, 0, len(req.Scans)),
	}

	for _, scan := range req.Scans {
		input.Scans = append(input.Scans, attendance.ProcessAttendanceScanInput{
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
