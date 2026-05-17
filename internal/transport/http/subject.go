package http

import (
	"fmt"
	nethttp "net/http"
)

func (h *Handler) GetSubjects(w nethttp.ResponseWriter, r *nethttp.Request) {
	if r.Method != nethttp.MethodGet {
		writeJSON(w, nethttp.StatusMethodNotAllowed, map[string]interface{}{
			"status": "error",
			"error":  "method not allowed",
		})
		return
	}

	subjects, err := h.attendanceService.GetSubjects(r.Context())
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
