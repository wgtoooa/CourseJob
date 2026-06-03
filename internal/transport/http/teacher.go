package http

import (
	"fmt"
	"net/http"
)

// GetTeachers returns teachers catalog.
// @Summary Get teachers
// @Description Returns all teachers from the catalog.
// @Tags Catalog
// @Produce json
// @Success 200 {object} TeachersResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/teachers [get]
func (h *Handler) GetTeachers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
			"status": "error",
			"error":  "method not allowed",
		})
		return
	}

	teachers, err := h.scheduleService.GetTeachers(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"status": "error",
			"error":  fmt.Sprintf("error getting teachers: %v", err),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":   "success",
		"teachers": teachers,
	})
}
