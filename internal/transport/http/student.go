package http

import (
	"CourseJob/internal/service/attendance"
	"CourseJob/internal/transport/http/dto"
	"CourseJob/internal/transport/http/validator"
	"bytes"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	"io"
	nethttp "net/http"
)

// AddStudent creates one or many students.
// @Summary Create students
// @Description Creates one student from an object payload or many students from an array payload.
// @Tags Students
// @Accept json
// @Produce json
// @Param request body dto.StudentRequest true "Student payload (single object or array of objects)"
// @Success 201 {object} StudentCreateResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/student [post]
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

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, nethttp.StatusBadRequest, jsonResponse{
			"status": "error",
			"error":  "invalid request body",
		})
		return
	}

	reqs, err := parseStudentRequests(body)
	if err != nil {
		writeJSON(w, nethttp.StatusBadRequest, jsonResponse{
			"status": "error",
			"error":  "invalid request body",
		})
		return
	}

	inputs := make([]attendance.StudentInput, 0, len(reqs))
	for i := range reqs {
		validator.NormalizeStudentRequest(&reqs[i])
		if err := validator.ValidatorStudent(&reqs[i]); err != nil {
			writeJSON(w, nethttp.StatusBadRequest, jsonResponse{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		inputs = append(inputs, attendance.StudentInput{
			FullName:  reqs[i].FullName,
			Course:    reqs[i].Course,
			GroupName: reqs[i].GroupName,
			Email:     reqs[i].Email,
			CardUID:   reqs[i].CardUID,
			CreatedAt: reqs[i].CreatedAt,
		})
	}

	err = h.attendanceService.CreateStudents(r.Context(), inputs)
	if err != nil {
		if errors.Is(err, attendance.ErrStudentExists) {
			writeJSON(w, nethttp.StatusConflict, jsonResponse{
				"status": "error",
				"error":  "student with this card_uid already exists",
			})
			return
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			writeJSON(w, nethttp.StatusConflict, jsonResponse{
				"status": "error",
				"error":  "student with this card_uid or email already exists",
			})
			return
		}

		writeJSON(w, nethttp.StatusInternalServerError, jsonResponse{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	if len(inputs) == 1 {
		writeJSON(w, nethttp.StatusCreated, jsonResponse{
			"status": "created",
		})
		return
	}

	writeJSON(w, nethttp.StatusCreated, jsonResponse{
		"status":        "created",
		"created_count": len(inputs),
	})
}

func parseStudentRequests(body []byte) ([]dto.StudentRequest, error) {
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 {
		return nil, errors.New("empty body")
	}

	switch trimmed[0] {
	case '{':
		var req dto.StudentRequest
		if err := json.Unmarshal(trimmed, &req); err != nil {
			return nil, err
		}
		return []dto.StudentRequest{req}, nil
	case '[':
		var reqs []dto.StudentRequest
		if err := json.Unmarshal(trimmed, &reqs); err != nil {
			return nil, err
		}
		if len(reqs) == 0 {
			return nil, errors.New("empty students array")
		}
		return reqs, nil
	default:
		return nil, errors.New("invalid json")
	}
}
