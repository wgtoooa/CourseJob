package http

import (
	"fmt"
	nethttp "net/http"
)

// GetSubjects returns subjects catalog.
// @Summary Get subjects
// @Description Returns all subjects from the catalog.
// @Tags Catalog
// @Produce json
// @Success 200 {object} SubjectsResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/subjects [get]
func (h *Handler) GetSubjects(w nethttp.ResponseWriter, r *nethttp.Request) {
	if r.Method != nethttp.MethodGet {
		writeJSON(w, nethttp.StatusMethodNotAllowed, map[string]interface{}{
			"status": "error",
			"error":  "method not allowed",
		})
		return
	}

	subjects, err := h.scheduleService.GetSubjects(r.Context())
	if err != nil {
		writeJSON(w, nethttp.StatusInternalServerError, map[string]interface{}{
			"status": "error",
			"error":  fmt.Sprintf("error getting subjects: %v", err),
		})
		return
	}

	writeJSON(w, nethttp.StatusOK, map[string]interface{}{
		"status":   "success",
		"subjects": subjects,
	})
}
