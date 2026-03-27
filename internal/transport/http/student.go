package http

import (
	"CourseJob/internal/service"
	"CourseJob/internal/transport/http/dto"
	"CourseJob/internal/transport/http/validator"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	nethttp "net/http"
)

func (h *Handler) AddStudent(w nethttp.ResponseWriter, r *nethttp.Request) {
	if r.Method != nethttp.MethodPost {
		writeJSON(w, nethttp.StatusMethodNotAllowed, jsonResponse{
			"status": "error",
			"error":  "Method not allowed",
		})
		return
	}
	r.Body = nethttp.MaxBytesReader(w, r.Body, 1<<20) //~ 1 MB
	defer r.Body.Close()

	var req dto.StudentRequest
	dec := json.NewDecoder(r.Body)

	if err := dec.Decode(&req); err != nil {
		writeJSON(w, nethttp.StatusBadRequest, jsonResponse{
			"status": "error",
			"error":  "invalid request body",
		})
		return
	}
	validator.NormalizeStudentRequest(&req)

	if err := validator.ValidatorStudent(&req); err != nil {
		writeJSON(w, nethttp.StatusBadRequest, jsonResponse{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}
	input := service.StudentInput{
		FullName:  req.FullName,
		Course:    req.Course,
		GroupName: req.GroupName,
		CardUID:   req.CardUID,
		CreatedAt: req.CreatedAt,
	}
	if err := h.attendanceService.CreateStudent(r.Context(), &input); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			writeJSON(w, nethttp.StatusConflict, jsonResponse{
				"status": "error",
				"error":  "student with this card_uid already exists",
			})
			return
		}

		writeJSON(w, nethttp.StatusInternalServerError, jsonResponse{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}
	writeJSON(w, nethttp.StatusCreated, jsonResponse{
		"status": "created",
	})
}
