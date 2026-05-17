package http

import (
	"CourseJob/internal/service"
	"CourseJob/internal/transport/http/dto"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	nethttp "net/http"
	"strconv"
	"strings"
)

func (h *Handler) GetPlanByCourse(w nethttp.ResponseWriter, r *nethttp.Request) {
	if r.Method != nethttp.MethodGet {
		writeJSON(w, nethttp.StatusMethodNotAllowed, map[string]interface{}{
			"status": "error",
			"error":  "method not allowed",
		})
		return
	}

	courseRaw := strings.TrimSpace(r.URL.Query().Get("course"))
	if courseRaw == "" {
		writeJSON(w, nethttp.StatusBadRequest, map[string]interface{}{
			"status": "error",
			"error":  "course query parameter is required",
		})
		return
	}

	course, err := strconv.Atoi(courseRaw)
	if err != nil || course < 1 || course > 4 {
		writeJSON(w, nethttp.StatusBadRequest, map[string]interface{}{
			"status": "error",
			"error":  "course must be an integer from 1 to 4",
		})
		return
	}

	items, err := h.attendanceService.GetPlanByCourse(r.Context(), course)
	if err != nil {
		log.Printf("get plan failed: course=%d err=%v", course, err)
		writeJSON(w, nethttp.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"error":   "plan_fetch_failed",
			"message": fmt.Sprintf("failed to get plan: %v", err),
		})
		return
	}

	response := make([]dto.PlanItemResponse, 0, len(items))
	for _, item := range items {
		response = append(response, dto.PlanItemResponse{
			Course:       item.Course,
			Subject:      item.Subject,
			PlannedPairs: item.PlannedPairs,
		})
	}

	writeJSON(w, nethttp.StatusOK, response)
}

func (h *Handler) UpsertPlanItem(w nethttp.ResponseWriter, r *nethttp.Request) {
	if r.Method != nethttp.MethodPut {
		writeJSON(w, nethttp.StatusMethodNotAllowed, map[string]interface{}{
			"status": "error",
			"error":  "method not allowed",
		})
		return
	}

	r.Body = nethttp.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()

	var req dto.PlanUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, nethttp.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"error":   "invalid_payload",
			"message": err.Error(),
		})
		return
	}

	err := h.attendanceService.UpsertPlanItem(r.Context(), service.PlanUpsertInput{
		Course:       req.Course,
		Subject:      req.Subject,
		PlannedPairs: req.PlannedPairs,
	})
	if err != nil {
		status := nethttp.StatusInternalServerError
		errorCode := "plan_upsert_failed"
		if isPlanValidationError(err) {
			status = nethttp.StatusBadRequest
			errorCode = "invalid_payload"
		}
		log.Printf("upsert plan failed: course=%d subject=%q planned_pairs=%d err=%v", req.Course, req.Subject, req.PlannedPairs, err)
		writeJSON(w, status, map[string]interface{}{
			"status":  "error",
			"error":   errorCode,
			"message": err.Error(),
		})
		return
	}

	writeJSON(w, nethttp.StatusOK, map[string]interface{}{
		"status": "success",
	})
}

func isPlanValidationError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(strings.TrimSpace(err.Error()))
	known := []string{
		"course must be an integer from 1 to 4",
		"planned_pairs must be non-negative",
		"subject is required",
		"plan item is nil",
	}
	for _, part := range known {
		if strings.Contains(msg, part) {
			return true
		}
	}

	var target *json.SyntaxError
	return errors.As(err, &target)
}
