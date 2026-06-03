package http

import (
	"fmt"
	nethttp "net/http"
)

// GetRooms returns rooms catalog.
// @Summary Get rooms
// @Description Returns all rooms from the catalog.
// @Tags Catalog
// @Produce json
// @Success 200 {object} RoomsResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/rooms [get]
func (h *Handler) GetRooms(w nethttp.ResponseWriter, r *nethttp.Request) {
	if r.Method != nethttp.MethodGet {
		writeJSON(w, nethttp.StatusMethodNotAllowed, map[string]interface{}{
			"status": "error",
			"error":  "method not allowed",
		})
		return
	}

	rooms, err := h.scheduleService.GetRooms(r.Context())
	if err != nil {
		writeJSON(w, nethttp.StatusInternalServerError, map[string]interface{}{
			"status": "error",
			"error":  fmt.Sprintf("error getting rooms: %v", err),
		})
		return
	}

	writeJSON(w, nethttp.StatusOK, map[string]interface{}{
		"status": "success",
		"rooms":  rooms,
	})
}
