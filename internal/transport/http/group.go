package http

import (
	"fmt"
	nethttp "net/http"
	"strconv"
	"strings"
)

// GetGroups returns groups for a selected course.
// @Summary Get groups by course
// @Description Returns unique groups for the selected course from schedule data.
// @Tags Catalog
// @Produce json
// @Param course query int true "Course number (1-4)"
// @Success 200 {object} GroupsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/groups [get]
func (h *Handler) GetGroups(w nethttp.ResponseWriter, r *nethttp.Request) {
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

	groups, err := h.scheduleService.GetGroupsByCourse(r.Context(), course)
	if err != nil {
		writeJSON(w, nethttp.StatusInternalServerError, map[string]interface{}{
			"status": "error",
			"error":  fmt.Sprintf("error getting groups: %v", err),
		})
		return
	}

	writeJSON(w, nethttp.StatusOK, map[string]interface{}{
		"status": "success",
		"course": course,
		"groups": groups,
	})
}
